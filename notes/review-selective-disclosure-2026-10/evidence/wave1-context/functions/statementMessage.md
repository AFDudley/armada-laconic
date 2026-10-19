## `statementMessage` in src/disclosure/statement.ts (L53-L61)

**Purpose:** Extract the EIP-712 message object from a DisclosureStatement. The message is the payload that gets hashed and signed; it commits the EOA to the exact statement: the wallet's masterPublicKey, the verifier (audit recipient), the context (purpose), the issue time, the endpoints and the hash of the hop list.

**Inputs & Assumptions:**
- `statement` (DisclosureStatement): contains masterPublicKey (decimal string), verifier, context, issuedAt, endpoints array, and hopsHash. Trust: trusted (comes from the caller after validation, or is constructed fresh in buildDisclosurePackage).
- Implicit: The returned object is consumed by ethers.js `TypedDataEncoder.hash()` and `verifyTypedData()`, which expect field names and types to match `DISCLOSURE_EIP712_TYPES` (L17-L27).
- Preconditions: No validation. Assumes `statement.masterPublicKey` is a valid decimal string; `statement.endpoints` is an array of objects with `role` and `account` strings; `statement.hopsHash` is a 0x-prefixed 32-byte hex string.

**Outputs & Effects:**
- Returns a plain object with keys: `masterPublicKey`, `verifier`, `context`, `issuedAt`, `endpoints`, `hopsHash`.
- No state writes, no external calls.
- The returned object is hashable by ethers' EIP-712 encoder; the structure binds the EOA signature to the exact claim.

**Block-by-Block:**

```typescript
// L53-L61
export function statementMessage(statement: DisclosureStatement): Record<string, unknown> {
  return {
    masterPublicKey: statement.masterPublicKey,
    verifier: statement.verifier,
    context: statement.context,
    issuedAt: statement.issuedAt,
    endpoints: statement.endpoints.map((e) => ({ role: e.role, account: e.account })),
    hopsHash: statement.hopsHash,
  };
}
```

- **What:** Reshape the statement into the EIP-712 message structure.
- **Why here:** The message object must exactly match the shape defined in `DISCLOSURE_EIP712_TYPES` (L17-L27), field-by-field. Extracting it ensures the caller cannot accidentally omit a field or add extra ones.
- **Assumes:** Each field in `statement` has the type declared in `DISCLOSURE_EIP712_TYPES`. The `endpoints` array is correctly shaped (each endpoint has `role` and `account`). The remapping on L57 creates new objects, preserving the values.
- **Establishes:** An object whose structure and field names are guaranteed to match the EIP-712 schema. Ethers' `TypedDataEncoder.hash()` expects this exact shape.
- **Depended on by:** `statementDigest()` (L104) and `verifyEndpointSignature()` (L119). Also in builder.ts `buildDisclosurePackage()` (L105) — the function extracts the message once and reuses it for all endpoint signings.

**Cross-Function Dependencies:**
- Constants: `DISCLOSURE_EIP712_TYPES` (L17-L27) — defines the field names and types that the message must match. The return value conforms to the `Statement` type in that constant.
- Callers: `statementDigest()`, `verifyEndpointSignature()`, and builder.ts `buildDisclosurePackage()`.

**Open Questions:**
- None. The function is a pure projection with no hidden dependencies.
