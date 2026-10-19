## `NullifierOwnership` in armada-circuits/src/operations/NullifierDisclosure.circom (L39-L57)

**Purpose:** Derives `masterPublicKey` (M) and `nullifier` from the wallet's private keys and a
leaf index, using the same two Poseidon formulas as `Transact.circom` steps 1 and 2d
(NullifierDisclosure.circom:23). It exists as a separate template "so it can be tested against
real spends" (L36-L37), and `NullifierDisclosure` instantiates it once (L79).

**Inputs & Assumptions:**
- `spendingPublicKey[2]` (field elements, meant to be [Ax, Ay] on BabyJubJub): untrusted witness
  values. Nothing in this template checks they are on the curve -- that check is the caller's job
  (see `NullifierDisclosure` L96-L98, `aOnCurve = BabyCheck()`).
- `nullifyingKey` (field element): untrusted witness value. No range or non-zero check here or in
  the caller.
- `leafIndex` (field element): untrusted witness value. No range check inside this template; the
  caller range-checks it separately (`NullifierDisclosure.circom:73-74`, `Num2Bits(16)`) before
  wiring it in at L83.
- Implicit: none (pure arithmetic template, no external state, no caller identity).

**Outputs & Effects:**
- `masterPublicKey <== mpk.out` (L51): `Poseidon(spendingPublicKey[0], spendingPublicKey[1], nullifyingKey)`.
- `nullifier <== nf.out` (L56): `Poseidon(nullifyingKey, leafIndex)`.
- No side effects, no state writes. Pure signal computation.

**Block-by-Block:**

```circom
// L47-L51
component mpk = Poseidon(3);
mpk.inputs[0] <== spendingPublicKey[0];
mpk.inputs[1] <== spendingPublicKey[1];
mpk.inputs[2] <== nullifyingKey;
masterPublicKey <== mpk.out;
```
- **What:** Computes M = Poseidon(Ax, Ay, nullifyingKey).
- **Why here:** Matches `Transact.circom:60` (`mpkHash = Poseidon(3)`) exactly in input order, so a
  wallet's M is the same whether derived from a spend or a disclosure.
- **Assumes:** The three inputs are field elements; Poseidon's own constraints (see
  `Poseidon.md` / `PoseidonEx.md`) are the only thing enforcing the hash relation.
- **Establishes:** `masterPublicKey` is a deterministic function of the three witness values --
  nothing here ties it to any specific real wallet; that binding comes only from whatever the
  caller compares `masterPublicKey` against (see Cross-Function Dependencies).

```circom
// L53-L56
component nf = Poseidon(2);
nf.inputs[0] <== nullifyingKey;
nf.inputs[1] <== leafIndex;
nullifier <== nf.out;
```
- **What:** Computes N = Poseidon(nullifyingKey, leafIndex).
- **Why here:** Matches `Transact.circom:93-96` (step 2d) exactly.
- **Assumes:** Same as above -- Poseidon's constraints are the only enforcement.
- **Establishes:** `nullifier` is deterministic in (`nullifyingKey`, `leafIndex`); the template
  itself does not establish that this `nullifier` corresponds to a real published nullifier -- that
  is checked off-circuit per `NullifierDisclosure.circom:24-25`.

**Cross-Function Dependencies:**
- Callee `Poseidon(3)` (internal, circomlib): depended on to make `masterPublicKey` a correct,
  collision-resistant hash of its three inputs. See `Poseidon.md` / `PoseidonEx.md`.
- Callee `Poseidon(2)` (internal, circomlib): depended on to make `nullifier` a correct hash of
  its two inputs.
- Caller `NullifierDisclosure` (`NullifierDisclosure.circom:79-85`): instantiates this template and
  then asserts `masterPublicKey === own.masterPublicKey` (L84) and `nullifier === own.nullifier`
  (L85) against the circuit's *public* inputs of the same names. This is the step that turns "some
  function of some private keys" into "the function of private keys that the public values claim to
  be". Without that equality check, this template's outputs would be unconstrained relative to the
  public signals.
- Caller `tests/disclosure.test.js` also instantiates a standalone build of this template
  (`ownership_test.wasm`, loaded at `tests/disclosure.test.js:107`) to replay 9 captured real spends
  and check that the derived `masterPublicKey`/`nullifier` match the chain's recorded values
  (`tests/disclosure.test.js:108-124`). This is test-level evidence, not an in-circuit guarantee.
- Shared state: none (no persisted state).
- Invariant couplings: the formulas here are a verbatim copy of `Transact.circom`'s formulas
  (`Transact.circom:60-63` for M, `Transact.circom:93-96` for N). If `Transact.circom` ever changes
  these formulas, this template must change with it or disclosures will not match real nullifiers.
  `nothing found` enforces this coupling mechanically (no shared template, no imported constant
  linking the two files) -- it is a documentation-level invariant only
  (`NullifierDisclosure.circom:23`, `Transact.circom` comment block).

**Open Questions:**
- `nullifyingKey` has no explicit non-zero or range constraint in this template or in
  `NullifierDisclosure`. The spec (`engineering/16-selective-disclosure.md`, "Key facts") defines
  `nullifyingKey = Poseidon(viewingPrivateKey)`, which is itself a Poseidon output and so already a
  field element in range; whether a zero or otherwise degenerate `nullifyingKey` can occur in
  practice, and whether that would matter, is not addressed by this template -- `nothing found`
  that rules it out here.
