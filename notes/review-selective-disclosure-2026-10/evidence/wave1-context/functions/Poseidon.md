## `Poseidon` (and `PoseidonEx`) in armada-circuits/node_modules/circomlib/circuits/poseidon.circom (L198-L208, L72-L196)

Vendored circomlib 2.0.5 template, not modified by this PR. Used throughout
`NullifierDisclosure`/`NullifierOwnership` as `Poseidon(2)`, `Poseidon(3)`, `Poseidon(5)`, and
inside `EdDSAPoseidonVerifier` as `Poseidon(5)` for the challenge hash.

**Purpose:** `Poseidon(nInputs)` is a thin wrapper (L198-207) around `PoseidonEx(nInputs, 1)`
(the full sponge permutation, L72-196) that fixes `initialState = 0` and exposes a single output.
It is the hash function every binding in the disclosure statement relies on
(`masterPublicKey`, `nullifier`, and the signed message are all Poseidon outputs).

**Inputs & Assumptions:**
- `inputs[nInputs]` (field elements): untrusted witness or wired signals, no range restriction
  beyond being field elements -- Poseidon is defined over the full field, so no separate range
  check is needed before hashing (unlike, say, feeding a value into a bit decomposition sized for
  a narrower range).
- `nInputs` is a compile-time parameter; the round-count tables (`N_ROUNDS_P`, L85) are indexed by
  `t = nInputs + 1` and only populated for `t` in [2, 17] (`N_ROUNDS_P[16]` array, L85) i.e.
  `nInputs` in [1, 16]. All uses in this review's scope (2, 3, 5) are within that range.

**Outputs & Effects:**
- `out` (single field element, via `PoseidonEx`'s `out[1]`): the hash. No side effects.

**Block-by-Block:**

```circom
// L198-207
component pEx = PoseidonEx(nInputs, 1);
pEx.initialState <== 0;
for (var i=0; i<nInputs; i++) { pEx.inputs[i] <== inputs[i]; }
out <== pEx.out[0];
```
- **What:** Delegates to `PoseidonEx` with a fixed zero initial state and one output.
- **Assumes:** `PoseidonEx`'s round structure (Ark/Sigma/Mix/MixS/MixLast, L18-196) correctly
  implements the Poseidon permutation with the circomlib-generated round constants
  (`POSEIDON_C`/`POSEIDON_S`/`POSEIDON_M`/`POSEIDON_P`, imported from `poseidon_constants.circom`,
  not independently re-read in this pass -- their correctness is an external, well-known-library
  assumption, not re-derived here).
- **Establishes:** `out` is the Poseidon hash of `inputs`, under the standard assumption that
  Poseidon (with these constants) is collision-resistant and preimage-resistant. This analysis
  does not re-derive that cryptographic property; it is taken as the premise the whole disclosure
  statement rests on (every binding in `NullifierOwnership`/`NullifierDisclosure` is a Poseidon
  equality).

**Cross-Function Dependencies:**
- Callees: `Sigma`, `Ark`, `Mix`, `MixS`, `MixLast` (all internal, same file) -- the S-box and
  linear-layer sub-templates of the permutation; not individually re-verified in this pass.
- Callers in scope: `NullifierOwnership.circom:47,53` (Poseidon(3), Poseidon(2));
  `NullifierDisclosure.circom:88` (Poseidon(5)); `eddsaposeidon.circom:55` (Poseidon(5), inside
  `EdDSAPoseidonVerifier`).
- Design-doc-level verification: the "Verification status" table in
  `engineering/16-selective-disclosure.md` records that `Poseidon(2)`, `Poseidon(3)`, `Poseidon(5)`
  passed Picus with cvc5 at *weak* safety ("properly constrained"), via the standalone files
  `tests/picus/poseidon2.circom`, `poseidon3.circom`, `poseidon5.circom`
  (`scripts/picus.sh: check poseidon2 8`, etc.). Weak safety (per the Picus README, not re-read in
  this pass) is a weaker guarantee than the strong safety claimed for the `wiring` piece -- this
  analysis did not independently re-verify what weak-vs-strong safety means for the specific case
  of fixed-arity Poseidon instances; noted as an open question below for completeness, though the
  design doc already states this gap plainly.
- Shared state: none.

**Open Questions:**
- The round constants (`POSEIDON_C`, `POSEIDON_S`, `POSEIDON_M`, `POSEIDON_P` from
  `poseidon_constants.circom`) were not read in this pass; their provenance (matching the published
  Poseidon paper's recommended parameters, cited in this file's own comment at L80-81) is assumed,
  not independently confirmed.
- The precise distinction between Picus "weak safety" and "strong safety" for the Poseidon pieces,
  and what residual gap weak safety leaves open, was not independently investigated here -- the
  design doc states the result but not the definition; carried forward as a gap in this analysis
  rather than resolved.
