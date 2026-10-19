## `NullifierDisclosure` in armada-circuits/src/operations/NullifierDisclosure.circom (L59-L112)

**Purpose:** The statement circuit for selective disclosure (ADR-0020). Proves that a published
`nullifier` N belongs to a leaf (`treeNumber` T, `leafIndex` L) of the wallet with
`masterPublicKey` M, and that the spending key of M authorized the disclosure of N for a given
`statementHash` S. Wired to the Groth16 entry point in `src/main/main_disclosure.circom:5` with
`[statementHash, masterPublicKey, nullifier, treeNumber, leafIndex]` public.

**Inputs & Assumptions:**
- `statementHash`, `masterPublicKey`, `nullifier`, `treeNumber`, `leafIndex` (public signals,
  `main_disclosure.circom:5`): untrusted in the sense that a malicious prover chooses them, but
  bound to the proof by Groth16 once something inside the circuit uses them in a constraint.
  Comment at L20-21 states the design rationale: "A Groth16 public input that no constraint uses
  is not bound by the proof" -- see Block-by-Block L87-93 for where each one is used.
- `spendingPublicKey[2]`, `nullifyingKey`, `signature[3]` (private witness): untrusted prover
  input; constrained by the blocks below.
- Implicit: none. No Merkle path, no tree root -- the spec (`engineering/16-selective-disclosure.md`,
  "The disclosure circuit") states the leaf position is public in the `Shield`/`Transact` events, so
  this circuit does not need one.
- Precondition relied on by callers (per spec): the verifier, off-circuit, checks that the
  disclosed commitment is actually at (T, L) and that N is in the pool's `Nullified` event for T
  (comment L24-25). `nothing found` inside this circuit that performs or requires that check --
  it is entirely the SDK checker's responsibility (see `checkDisclosurePackage` analysis).

**Outputs & Effects:**
- No output signals. The template's only effect is to add constraints; satisfiability of those
  constraints against a public-input vector is what a Groth16 verifier checks.

**Block-by-Block:**

```circom
// L72-L76
component leafBits = Num2Bits(16);
leafBits.in <== leafIndex;
component treeBits = Num2Bits(16);
treeBits.in <== treeNumber;
```
- **What:** Range-checks `leafIndex` and `treeNumber` each to 16 bits.
- **Why here:** Comment L72: "The pool's trees have depth 16, and treeNumber is a uint16 in
  events." Must happen before these values are folded into `msg` (L87-93) and before the
  off-circuit lookup, so a prover cannot supply an out-of-range value that happens to collide
  with an in-range one only by fluke of field arithmetic (it cannot, since `Num2Bits(n)`'s
  `lc1 === in` constraint (`bitify.circom:38`) is an exact equality over the field, not a
  truncation -- a value >= 2^16 has no satisfying bit decomposition and the witness generator
  errors, per `tests/disclosure.test.js` `FAIL.range16` regex matching `Num2Bits_\d+ line: 38`
  and the test `rejects: leafIndex or treeNumber outside 16 bits, even when signed`
  (`tests/disclosure.test.js:252-255`)).
- **Assumes:** `Num2Bits`'s own constraints are correct (see `Num2Bits.md`).
- **Establishes:** `leafIndex`, `treeNumber` ∈ [0, 2^16). Depended on by the off-circuit verifier,
  which per spec interprets T and L as a real tree/leaf position.

```circom
// L78-L85
component own = NullifierOwnership();
own.spendingPublicKey[0] <== spendingPublicKey[0];
own.spendingPublicKey[1] <== spendingPublicKey[1];
own.nullifyingKey <== nullifyingKey;
own.leafIndex <== leafIndex;
masterPublicKey === own.masterPublicKey;
nullifier === own.nullifier;
```
- **What:** Derives M and N from the private keys and asserts they equal the public `masterPublicKey`
  and `nullifier`.
- **Why here:** Must happen before the message hash (L87-93) uses `masterPublicKey`/`nullifier`,
  so the message commits to the *checked* values, not to unchecked public inputs.
- **Assumes:** See `NullifierOwnership.md`.
- **Establishes:** the public `masterPublicKey` and `nullifier` are exactly the values produced by
  the private `spendingPublicKey`/`nullifyingKey`/`leafIndex` witness. Depended on by the
  off-circuit verifier (it trusts that if the proof verifies, M and N are consistent with *some*
  keys) and by the signature check below (binds M to the signer).

```circom
// L87-93
component msg = Poseidon(5);
msg.inputs[0] <== statementHash;
msg.inputs[1] <== masterPublicKey;
msg.inputs[2] <== nullifier;
msg.inputs[3] <== treeNumber;
msg.inputs[4] <== leafIndex;
```
- **What:** Hashes all five public inputs into one message.
- **Why here:** Must happen after the public inputs are wired in, and before the EdDSA check
  (L104-111) consumes `msg.out` as `eddsa.M`. This is the step that gives the rationale at L20-21
  teeth: every public signal is an argument to this hash, and the hash is what the signature
  covers, so Groth16 binds all five.
- **Assumes:** `Poseidon(5)`'s constraints are correct (see `Poseidon.md`).
- **Establishes:** `msg.out` is a single field element binding all five public inputs. Depended on
  by the EdDSA check (L111) and, transitively, by the "rejects: a signature over another
  statement" / "rejects: each public input changed after signing" tests
  (`tests/disclosure.test.js:224-245`), which show that changing *any* of `statementHash`,
  `treeNumber`, or `leafIndex` after signing makes `FAIL.signature` fire (the `ForceEqualIfEnabled`
  inside `EdDSAPoseidonVerifier`), because the signature was computed over the old `msg.out`.

```circom
// L96-101
component aOnCurve = BabyCheck();
aOnCurve.x <== spendingPublicKey[0];
aOnCurve.y <== spendingPublicKey[1];
component rOnCurve = BabyCheck();
rOnCurve.x <== signature[0];
rOnCurve.y <== signature[1];
```
- **What:** Asserts both `spendingPublicKey` (A) and `signature[0..1]` (R8) satisfy the BabyJubJub
  curve equation.
- **Why here:** Comment L95: "circomlib's verifier does not check this itself." Confirmed by
  reading `EdDSAPoseidonVerifier` (see `EdDSAPoseidonVerifier.md`): its `BabyAdd`/`BabyDbl` calls
  compute the twisted-Edwards addition formula algebraically without asserting either operand is
  on the curve first, so an off-curve R8 (or A) could make the final `ForceEqualIfEnabled` pass by
  coincidence of the addition/doubling formulas, were these two `BabyCheck` calls absent.
- **Assumes:** `BabyCheck`'s own constraint is correct (see `BabyCheck.md`).
- **Establishes:** A and R8 are curve points. Confirmed enforced by test
  `rejects: points off the curve` (`tests/disclosure.test.js:257-263`), which perturbs R8.y and
  A.y by +1 and expects `FAIL.onCurve` (`BabyCheck_\d+ line: 82`).

```circom
// L104-111
component eddsa = EdDSAPoseidonVerifier();
eddsa.enabled <== 1;
eddsa.Ax <== spendingPublicKey[0];
eddsa.Ay <== spendingPublicKey[1];
eddsa.R8x <== signature[0];
eddsa.R8y <== signature[1];
eddsa.S <== signature[2];
eddsa.M <== msg.out;
```
- **What:** Verifies the EdDSA (BabyJubJub, Poseidon) signature over `msg.out` by key A.
- **Why here:** Last block; depends on everything above (the message, the on-curve checks).
  `enabled <== 1` is a constant, not a signal -- the check is unconditional, never toggleable off
  by a witness value.
- **Assumes:** `EdDSAPoseidonVerifier`'s constraints are correct, in particular that it rejects
  `S >= subgroup order` and a small-order `A` (see `EdDSAPoseidonVerifier.md`). The design doc
  (`engineering/16-selective-disclosure.md`, "Verification status") records that this template is
  *not* independently formally verified here -- Picus times out on its scalar multiplications, and
  no published proof covers it; it is treated as "a trusted circomlib component, as it is for
  `Transact.circom`."
- **Establishes:** only a holder of the private key behind A produced a valid signature over
  `msg.out`, which (by the two blocks above) commits to all five public inputs. Combined with
  L84-85, this establishes that the *spending* key of the wallet claiming M authorized disclosure
  of N for statement S at (T, L) -- not merely a holder of the viewing/nullifying key. This is the
  circuit's central security property per spec ("Why the circuit checks the spending key").

**Cross-Function Dependencies:**
- Callee `Num2Bits(16)` x2 (internal, circomlib): range-checks L and T. See `Num2Bits.md`.
- Callee `NullifierOwnership` (internal, this file): derives M, N from keys. See
  `NullifierOwnership.md`.
- Callee `Poseidon(5)` (internal, circomlib): computes the signed message. See `Poseidon.md`.
- Callee `BabyCheck` x2 (internal, circomlib): on-curve checks for A and R8. See `BabyCheck.md`.
- Callee `EdDSAPoseidonVerifier` (internal, circomlib): the signature check. See
  `EdDSAPoseidonVerifier.md`.
- Caller `main_disclosure.circom:5`: `component main {public [statementHash, masterPublicKey,
  nullifier, treeNumber, leafIndex]} = NullifierDisclosure();`. This is the only place that fixes
  which five signals are public (and in what order the Groth16 verifier contract will expect
  them) -- the order here must match whatever `armada-sdk`'s `disclosurePublicSignals` /
  `formatDisclosureInput` produce (see SDK circuit.ts analysis for that cross-check) and whatever
  order the off-chain verifier passes to a Groth16 `verify` call.
- Caller `tests/disclosure.test.js`: exercises acceptance (`tests/disclosure.test.js:155-180`),
  ten distinct rejection cases (`tests/disclosure.test.js:184-282`), and one Groth16 round trip
  (`tests/disclosure.test.js:285-294`) that re-verifies the proof after perturbing each public
  signal independently and expects failure every time -- this is the test-level confirmation of
  the L20-21 design rationale (all five public inputs are bound).
- Picus coverage: `tests/picus/wiring.circom` models this template's range checks, equality
  checks, and on-curve checks with the three Poseidon hashes and the EdDSA check replaced by free
  inputs/outputs, and reports "properly constrained" under `--strong` (`scripts/picus.sh:
  check wiring 8 --strong`). It does **not** cover the EdDSA check itself, and the file's own
  header comment (`tests/picus/wiring.circom:1-6`) warns that a change to
  `NullifierDisclosure.circom` not mirrored there "leaves the change unchecked by Picus" --
  i.e. the Picus coverage can silently go stale relative to the real circuit.
- Shared state: none.
- Invariant couplings: the public-signal order (`main_disclosure.circom:5`) is a cross-repo
  contract with the SDK's circuit adapter and with whatever Groth16 verifier contract or
  verification key consumes `publicSignals`. No in-repo mechanism enforces that order match; it is
  convention plus the shared comment block (L17).

**Open Questions:**
- Whether the off-circuit check that "the nullifier is in the pool's `Nullified` event for
  treeNumber" (L24-25) is actually performed by the SDK checker on every disclosure proof it
  processes, not just referenced in the design -- this circuit only has `nothing found` on that
  point by construction (it is explicitly off-circuit); the SDK-side checker analysis
  (`checkDisclosurePackage.md`) is where this must be confirmed or left open.
- Whether anything outside this circuit enforces that the public-signal order in
  `main_disclosure.circom:5` matches the order the SDK's `formatDisclosureInput`/
  `disclosurePublicSignals` produce, and the order a verifier contract or off-chain `groth16.verify`
  call expects -- `nothing found` inside the circuit repo itself (see SDK `circuit.ts` analysis).
