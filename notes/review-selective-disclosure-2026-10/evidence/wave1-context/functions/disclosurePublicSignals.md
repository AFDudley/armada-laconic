## `disclosurePublicSignals` in /home/rix/code/git_puller/repos/armada-sdk/src/disclosure/circuit.ts (L60-L62)

**Purpose:** Extracts the five public signals from a circuit input in the exact order the Groth16 proof exposes them. Used by the prover to label the proof's outputs and by the verifier to re-derive what the proof should have committed to.

**Inputs & Assumptions:**
- `input` (DisclosureCircuitInput): A complete circuit witness with five public-signal fields and four private witness fields. Trust: trusted (must be produced by `formatDisclosureInput` or an equivalent source).
  - Precondition: `input.statementHash`, `input.masterPublicKey`, `input.nullifier`, `input.treeNumber`, `input.leafIndex` must all be present and be bigints.

**Outputs & Effects:**
- Returns a `bigint[]` with exactly five elements in this order: `[statementHash, masterPublicKey, nullifier, treeNumber, leafIndex]`.
- No state writes; no external effects.

**Block-by-Block:**

```typescript
export function disclosurePublicSignals(input: DisclosureCircuitInput): bigint[] {
  return [input.statementHash, input.masterPublicKey, input.nullifier, input.treeNumber, input.leafIndex];
}
```
- **What:** Constructs an array of the five public signal values from the circuit input.
- **Why here:** This is the entire function; it extracts and orders the public signals.
- **Assumes:** The field names on `DisclosureCircuitInput` match the public signal names declared in `DISCLOSURE_PUBLIC_SIGNAL_NAMES` (circuit.ts:3-8) and in main_disclosure.circom:1 (`public [statementHash, masterPublicKey, nullifier, treeNumber, leafIndex]`). The order is critical: Groth16 proofs expose signals in declaration order, and a mismatch will cause verification to fail.
- **Establishes:** An array of public signals in the order the circuit declares them, suitable for passing to a verifier or for the prover to label the proof.
- **Depended on by:** The prover (to extract signals from the witness before passing them to snarkjs), and the verifier (to reconstruct the signals from the proof for comparison against on-chain data).

**Cross-Function Dependencies:**
- None. The function is a pure data extraction.
- Callers: The prover (in src/prover/*) and the disclosure checker (in src/disclosure/* or the checker module).

**Open Questions:**
- Is there a test that verifies the order of these signals matches the circom declaration exactly? If the order were wrong, the prover would succeed but the verifier would reject all proofs from this version.
- The array returned is a mutable `bigint[]`. If a caller modifies it, does that affect the proof or the verifier? (Unlikely, but the immutability of the signal array might be enforced elsewhere.)

