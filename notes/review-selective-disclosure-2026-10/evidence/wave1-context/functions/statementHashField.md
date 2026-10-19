## `statementHashField` in src/disclosure/statement.ts (L110-L115)

**Purpose:** Reduce the EIP-712 digest into the BN254 scalar field, producing the 256-bit statement hash S that the disclosure circuit's spending-key signature commits to. The EOA signs the full 256-bit digest; the spending key signs this field element. Both cover the same statement; the reduction keeps ~128-bit binding strength in the circuit's field.

**Inputs & Assumptions:**
- `domain` (DisclosureDomain): passed to `statementDigest()` to compute the full EIP-712 digest.
- `statement` (DisclosureStatement): passed to `statementDigest()`.
- Trust: trusted.
- Implicit: `statementDigest()` returns a 0x-prefixed 32-byte hex string representing a 256-bit integer. The string is converted to BigInt via `BigInt(...)`. Modular reduction is performed by the `%` operator. `SNARK_SCALAR_FIELD` is the BN254 scalar field prime (types.ts L8).
- Preconditions: No validation. Assumes the digest is a valid 32-byte hex string. Assumes `SNARK_SCALAR_FIELD` is the correct BN254 scalar field prime (= 21888242871839275222246405745257275088548364400416034343698204186575808495617n per types.ts L8).

**Outputs & Effects:**
- Returns a bigint in the range [0, SNARK_SCALAR_FIELD).
- No state writes, no external calls (except to `statementDigest()`).
- The field element is deterministic: the same domain and statement always produce the same field element. This value is the input `statementHash` to the disclosure circuit (circuit.ts `formatDisclosureInput()` L43).

**Block-by-Block:**

```typescript
// L110-L115
export function statementHashField(domain: DisclosureDomain, statement: DisclosureStatement): bigint {
  return BigInt(statementDigest(domain, statement)) % SNARK_SCALAR_FIELD;
}
```

- **What:** Get the EIP-712 digest and reduce it modulo the BN254 scalar field.
- **Why here:** The circuit operates in the BN254 field (witness values are field elements). The statement hash S is a public input to the circuit (spec section "The disclosure circuit", public inputs table). The circuit's EdDSA signature verification uses the field-reduced hash (circuit.ts L44, the `message` passed to `signEDDSA()`). The reduction ensures the hash fits in the field without wrapping.
- **Assumes:** `statementDigest()` returns a 0x-prefixed hex string. JavaScript's `BigInt(...)` can parse it. The modulo operation `% SNARK_SCALAR_FIELD` reduces any 256-bit value into the field. `SNARK_SCALAR_FIELD` is a constant bigint (not mutable).
- **Establishes:** A field element that is identical to the message hash the circuit's EdDSA signature verification checks (circuit.ts `formatDisclosureInput()` L44: the message is `Poseidon([statementHash, masterPublicKey, nullifier, treeNumber, leafIndex])`). The reduction is deterministic and idempotent (reducing twice gives the same result).
- **Depended on by:** builder.ts `buildDisclosurePackage()` (L96) — the field element is computed once and passed to `formatDisclosureInput()` for every proof (L99), ensuring all proofs in the package commit to the same statement.

**Cross-Function Dependencies:**
- Callee `statementDigest` (internal): returns the 256-bit EIP-712 digest. Fully internal and deterministic.
- Constant: `SNARK_SCALAR_FIELD` (types.ts L8) — the BN254 scalar field prime, 21888242871839275222246405745257275088548364400416034343698204186575808495617n.
- Callers: builder.ts `buildDisclosurePackage()` (L96) — the statement hash is computed once and threaded into every disclosure proof via `formatDisclosureInput()`.

**Open Questions:**
- The spec (engineering/16-selective-disclosure.md, section "The disclosure circuit", public inputs table) states the circuit takes "statement hash S" as a public input, with no visibility constraint listed (neither "public" nor "private"). The SDK implementation treats it as public (it is a public signal 0, returned by `disclosurePublicSignals()` in circuit.ts L51). Confirm that S is indeed a public signal in NullifierDisclosure.circom.
- The reduction `digest % SNARK_SCALAR_FIELD` keeps ~128 bits of binding strength because the field modulus is large. No explicit validation confirms that the digest is less than 2^256 (it always is, by construction from keccak256). The reduction is correct but assumes the field modulus is less than 2^256, which is true for BN254.
