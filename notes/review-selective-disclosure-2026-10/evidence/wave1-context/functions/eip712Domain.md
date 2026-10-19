## `eip712Domain` in src/disclosure/statement.ts (L45-L51)

**Purpose:** Construct the EIP-712 domain object that scopes the disclosure statement signature to one pool deployment and one blockchain. The domain binds the signature to a specific chain and verifying contract, preventing replay across chains or across different pool deployments.

**Inputs & Assumptions:**
- `domain` (DisclosureDomain): contains chainId and verifyingContract. Trust: trusted (provided by the caller who knows the pool deployment).
- Implicit: The `TypedDataDomain` object is consumed by ethers.js `TypedDataEncoder.hash()` and `verifyTypedData()`.
- Preconditions: No explicit checks. Assumes chainId is a valid integer and verifyingContract is a valid Ethereum address (checksummed or lowercase); validation happens downstream in ethers' EIP-712 processing.

**Outputs & Effects:**
- Returns a `TypedDataDomain` object with name, version, chainId, and verifyingContract.
- No state writes, no external calls.
- The returned domain binds the EOA's signature to exactly one pool and one chain ID; a signature cannot be moved between pools or chains.

**Block-by-Block:**

```typescript
// L45-L51
export function eip712Domain(domain: DisclosureDomain): TypedDataDomain {
  return {
    name: DISCLOSURE_EIP712_DOMAIN_NAME,
    version: DISCLOSURE_EIP712_DOMAIN_VERSION,
    chainId: domain.chainId,
    verifyingContract: domain.verifyingContract,
  };
}
```

- **What:** Wraps the pool-scoped domain into an ethers EIP-712 domain object.
- **Why here:** Called before any signature verification or generation; the domain must be fixed before a statement is signed.
- **Assumes:** `DISCLOSURE_EIP712_DOMAIN_NAME` and `DISCLOSURE_EIP712_DOMAIN_VERSION` are correctly defined as constants (L13, L14, both hard-coded strings). The input `domain.chainId` is a valid integer; the input `domain.verifyingContract` is a valid Ethereum address.
- **Establishes:** A `TypedDataDomain` that correctly identifies the pool deployment and chain. All downstream uses of `eip712Domain()` return an identical object for the same input.
- **Depended on by:** `statementDigest()` (L102) and `verifyEndpointSignature()` (L118) — both call `eip712Domain()` and pass its output to ethers.js for signature operations.

**Cross-Function Dependencies:**
- Callee `TypedDataDomain` (ethers external-source-available): ethers' type alias that maps to the EIP-712 domain object shape. The structure is documented in EIP-712 and ethers' type definitions.
- Callers: `statementDigest()` (L102), `verifyEndpointSignature()` (L118), and in builder.ts `buildDisclosurePackage()` (L104, where it calls `eip712Domain()` and passes the result to the endpoint signing request).
- Constants: `DISCLOSURE_EIP712_DOMAIN_NAME` (L13, string "ArmadaSelectiveDisclosure"), `DISCLOSURE_EIP712_DOMAIN_VERSION` (L14, string "1").

**Open Questions:**
- None. The function is a simple wrapper with no hidden dependencies or conditional logic.
