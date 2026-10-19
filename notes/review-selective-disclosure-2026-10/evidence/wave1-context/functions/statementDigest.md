## `statementDigest` in src/disclosure/statement.ts (L98-L107)

**Purpose:** Compute the EIP-712 digest of the statement — the 32-byte hash that EOAs sign. The digest is the full-width (256-bit) binding that commits the EOA to the entire statement: the wallet's masterPublicKey, the verifier, the context, the issue time, the endpoints and the hop list hash. The EOA signature covers this digest.

**Inputs & Assumptions:**
- `domain` (DisclosureDomain): chainId and verifyingContract, passed to `eip712Domain()` to create the ethers domain object.
- `statement` (DisclosureStatement): the statement to hash, passed to `statementMessage()` to create the message object.
- Trust: trusted (supplied by the caller after construction or validation).
- Implicit: Ethers' `TypedDataEncoder.hash()` is called with the domain, types schema, and message. The function returns a 0x-prefixed 32-byte hex string.
- Preconditions: No validation. Assumes the domain and statement are well-formed (validated upstream if they come from external sources).

**Outputs & Effects:**
- Returns a 0x-prefixed 32-byte hex string (the EIP-712 digest).
- No state writes, no external calls (except to ethers).
- The digest is the payload the EOA signs. Changing any field in the statement (masterPublicKey, verifier, context, issuedAt, endpoints, hopsHash) changes the digest.

**Block-by-Block:**

```typescript
// L98-L107
export function statementDigest(domain: DisclosureDomain, statement: DisclosureStatement): string {
  return TypedDataEncoder.hash(eip712Domain(domain), DISCLOSURE_EIP712_TYPES, statementMessage(statement));
}
```

- **What:** Call ethers' EIP-712 hash function with the domain, type schema, and message.
- **Why here:** EIP-712 requires hashing according to the standard (the domain separator is mixed with the struct hash). Ethers' `TypedDataEncoder.hash()` implements this. The function is a direct call to ethers; no intermediate steps.
- **Assumes:** `eip712Domain(domain)` returns a valid `TypedDataDomain` object (per eip712Domain analysis). `DISCLOSURE_EIP712_TYPES` is the correct EIP-712 type schema (defined L17-L27). `statementMessage(statement)` returns a message object whose structure matches the schema (per statementMessage analysis). Ethers' `TypedDataEncoder.hash()` succeeds and returns a 0x-prefixed 32-byte hex string.
- **Establishes:** The digest is deterministic and reproducible by any verifier who knows the domain and statement. The digest is bound to the domain (chainId and verifyingContract), so it cannot be reused on a different chain or pool.
- **Depended on by:** `statementHashField()` (L110-L111) and `verifyEndpointSignature()` (L118-L119). Both extract the digest and process it further.

**Cross-Function Dependencies:**
- Callee `eip712Domain` (internal): constructs the ethers domain object. Fully internal, deterministic.
- Callee `statementMessage` (internal): constructs the message object. Fully internal, deterministic.
- Callee `TypedDataEncoder.hash` (ethers external-source-available): implements EIP-712 hashing per the standard. Takes a domain, types schema, and message; returns a 0x-prefixed 32-byte hex string (the domain separator mixed with the struct hash).
- Constant: `DISCLOSURE_EIP712_TYPES` (L17-L27) — the EIP-712 schema. Must match the structure of `statementMessage()`'s return value.
- Callers: `statementHashField()` (L111) and `verifyEndpointSignature()` (L119). Both depend on this digest for signature operations.

**Open Questions:**
- None. The function delegates to ethers' standard EIP-712 implementation with no custom logic.
