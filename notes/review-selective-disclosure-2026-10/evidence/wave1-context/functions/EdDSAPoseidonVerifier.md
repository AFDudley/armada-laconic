## `EdDSAPoseidonVerifier` in armada-circuits/node_modules/circomlib/circuits/eddsaposeidon.circom (L27-L112)

Vendored circomlib 2.0.5 template, not modified by this PR. Analyzed because
`NullifierDisclosure` depends on it for the only non-hash cryptographic check in the disclosure
statement, and because the design doc explicitly flags it as unverified by Picus
(`engineering/16-selective-disclosure.md`, "Verification status").

**Purpose:** Verifies an EdDSA signature (R8, S) by public key A over message M, using the
Poseidon-based variant of EdDSA on BabyJubJub (RFC 8032-style, with Poseidon replacing the usual
hash). `enabled` toggles whether the check is enforced.

**Inputs & Assumptions:**
- `enabled` (field element, meant to be 0/1): the caller controls this. `NullifierDisclosure` wires
  a constant `1` (`NullifierDisclosure.circom:105`), so for this PR's use the check is always
  enforced. Nothing in this template itself constrains `enabled` to {0,1} -- if a caller passed an
  arbitrary non-boolean value, the two `*enabled === 0` constraints (L51, L83) would behave
  differently (e.g. `enabled=0` always satisfies them regardless of S/A, effectively disabling
  those two sub-checks, while the final equality checks are separately gated by
  `ForceEqualIfEnabled` at L104/L108, which `nothing found` in this file formally documents a
  stricter treatment of non-boolean `enabled` -- out of scope here since the caller fixes it to the
  constant 1).
- `Ax`, `Ay` (public key, meant to be on-curve): **not checked on-curve by this template** --
  confirmed by reading the whole body, the only point operations are `BabyDbl`/`BabyAdd`
  (algebraic formulas, no curve-membership assertion) and `EscalarMulAny`/`EscalarMulFix` (not
  read in full here, but the design doc and the circuit's own comment
  (`NullifierDisclosure.circom:95`) state this gap explicitly, and the caller compensates with its
  own `BabyCheck` on A). Trust: the caller must have already validated this, or validate it
  itself.
- `S`, `R8x`, `R8y`: the signature. `R8` is likewise not on-curve-checked here (same gap, same
  caller compensation via `BabyCheck` on R8).
- `M`: the message, an arbitrary field element.

**Outputs & Effects:** No output signals; only constraints.

**Block-by-Block:**

```circom
// L42-L51
component snum2bits = Num2Bits(253);
snum2bits.in <== S;
component compConstant = CompConstant(2736030358979909402780800718157159386076813972158567259200215660948447373040);
for (i=0; i<253; i++) { snum2bits.out[i] ==> compConstant.in[i]; }
compConstant.in[253] <== 0;
compConstant.out*enabled === 0;
```
- **What:** Decomposes S into 253 bits and checks S < the BabyJubJub subgroup order
  (2736030358979909402780800718157159386076813972158567259200215660948447373040), when `enabled`.
- **Why here:** Must happen before S is used in `EscalarMulFix` below, to stop a malleable
  signature where S' = S + subgroup_order still satisfies the verification equation (since the
  scalar multiplication is only defined correctly mod the subgroup order).
- **Assumes:** `Num2Bits(253)` and `CompConstant` are correct (not independently re-analyzed here;
  `Num2Bits` is analyzed separately for the n=16 use in `NullifierDisclosure`, and its general
  correctness is assumed identical for n=253).
- **Establishes:** S is in [0, subgroup_order) whenever `enabled=1`. Confirmed exercised by the
  disclosure circuit's test `rejects: a malleable signature (S + subgroup order)`
  (`tests/disclosure.test.js:247-250`), matching `FAIL.sRange` = `EdDSAPoseidonVerifier_\d+ line: 51`.

```circom
// L55-L64
component hash = Poseidon(5);
hash.inputs[0..4] <== R8x, R8y, Ax, Ay, M;
component h2bits = Num2Bits_strict();
h2bits.in <== hash.out;
```
- **What:** Computes the Fiat-Shamir-style challenge h = Poseidon(R8x, R8y, Ax, Ay, M) and
  decomposes it strictly into 254 bits (with alias check, i.e. the canonical representative, not
  just any value congruent mod the field).
- **Assumes:** `Poseidon(5)`'s constraints are correct.
- **Establishes:** `h2bits.out` is the canonical bit decomposition of h, used as the scalar for the
  escalar multiplication below.

```circom
// L70-L83
component dbl1 = BabyDbl(); dbl1.x <== Ax; dbl1.y <== Ay;
component dbl2 = BabyDbl(); dbl2.x <== dbl1.xout; dbl2.y <== dbl1.yout;
component dbl3 = BabyDbl(); dbl3.x <== dbl2.xout; dbl3.y <== dbl2.yout;
component isZero = IsZero(); isZero.in <== dbl3.x;
isZero.out*enabled === 0;
```
- **What:** Computes 8*A by three doublings (comment L68-69: "This also ensure[s] that the result
  is in the subgroup"), then checks 8*A is not the point at infinity (x=0), when `enabled`.
- **Why here:** A small-order A (order 1, 2, 4, or 8) makes 8*A the identity, which would make the
  final verification equation `S*B8 == R8 + h*8*A` degrade to `S*B8 == R8` -- a check independent
  of A entirely, letting any key "sign" by choosing R8 = S*B8 for an S of the attacker's choosing.
  This block closes that gap, **but only for small-order A** -- it does not itself establish A is
  on-curve at all (see Inputs & Assumptions above); an off-curve A with a non-identity `dbl3.x` would
  pass this `isZero` check despite being invalid, which is exactly why `NullifierDisclosure`'s own
  `BabyCheck` on A (`NullifierDisclosure.circom:96-98`) is necessary and not redundant.
- **Establishes:** 8*A != identity whenever `enabled=1`, assuming A is already known to be
  on-curve (established by the caller, not here). Confirmed exercised by
  `rejects: a forged signature for a small-order spending key`
  (`tests/disclosure.test.js:264-281`), matching `FAIL.smallOrderA` =
  `EdDSAPoseidonVerifier_\d+ line: 83`.

```circom
// L85-L111
component mulAny = EscalarMulAny(254); ... mulAny.p[0..1] <== dbl3.xout, dbl3.yout;
component addRight = BabyAdd(); addRight.x1,y1 <== R8x,R8y; addRight.x2,y2 <== mulAny.out[0..1];
component mulFix = EscalarMulFix(253, BASE8); ... (scalar = S's bits)
component eqCheckX = ForceEqualIfEnabled(); eqCheckX.enabled <== enabled; eqCheckX.in[0,1] <== mulFix.out[0], addRight.xout;
component eqCheckY = ForceEqualIfEnabled(); (same pattern for y)
```
- **What:** Computes right = R8 + h*(8*A) and left = S*B8 (B8 = the standard base point, fixed
  generator of the prime-order subgroup), and asserts left == right coordinate-wise, when
  `enabled`.
- **Assumes:** `EscalarMulAny`, `BabyAdd`, `EscalarMulFix`, `ForceEqualIfEnabled` are correct
  (not independently re-read in this analysis pass; flagged as an open question below since they
  are the templates Picus times out on per the design doc).
- **Establishes:** the full EdDSA verification equation holds, i.e. a valid signature by the
  holder of the private key behind A over message M, *given* A and R8 are on-curve (caller's job)
  and *given* the escalar-multiplication templates are themselves correctly constrained.

**Cross-Function Dependencies:**
- Callee `Num2Bits(253)`, `Num2Bits_strict`, `CompConstant`, `Poseidon(5)`, `BabyDbl`, `IsZero`,
  `EscalarMulAny`, `BabyAdd`, `EscalarMulFix`, `ForceEqualIfEnabled` (all internal, circomlib):
  depended on for every sub-check above. `EscalarMulAny`/`EscalarMulFix`/`ForceEqualIfEnabled` are
  not read in this pass (out of the explicitly named target list); their correctness is an
  unverified assumption here, consistent with the design doc's own statement that Picus cannot
  finish on the scalar multiplications and "No published work proves this template either"
  (`engineering/16-selective-disclosure.md`, "Verification status").
- Caller `NullifierDisclosure` (`NullifierDisclosure.circom:104-111`): supplies `enabled=1`
  (constant), and compensates for this template's missing on-curve checks on A and R8 with its own
  `BabyCheck` calls (L96-101). The caller's correctness therefore depends on this template
  correctly implementing the equation *conditional on* on-curve inputs, which is exactly the
  boundary this template documents in its own comments (L68-69, L80) but does not enforce for A
  (on-curveness) anywhere in its own body.
- Caller `Transact.circom:132-139`: the same template, same `enabled=1` pattern, used for ordinary
  spend authorization -- this is the "reference formula" the design doc cites as the precedent for
  treating this template as trusted (`engineering/16-selective-disclosure.md`, "Verification
  status": "the verifier stays a trusted circomlib component, as it is for `Transact.circom`").
- Shared state: none.

**Open Questions:**
- `EscalarMulAny`, `EscalarMulFix`, and `ForceEqualIfEnabled` were not read as part of this
  analysis pass (outside the stated target list of EdDSAPoseidonVerifier, BabyCheck, Num2Bits,
  Poseidon). Their internal correctness is assumed, matching the design doc's acknowledged gap
  (Picus timeout, no published proof) rather than this analysis independently confirming it.
- Whether a non-boolean `enabled` value (anything other than literal 0 or 1) could partially
  disable one of the two `*enabled === 0` checks while leaving the final `ForceEqualIfEnabled`
  checks active in some inconsistent combination is `nothing found` addressed in this template;
  moot for the disclosure circuit specifically because `enabled` is wired to the constant `1`
  (`NullifierDisclosure.circom:105`), not a witness-controlled signal.
