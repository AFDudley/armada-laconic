## `verifyEndpointSignature` in src/disclosure/statement.ts (L117-L131)

**Purpose:** Verify that a 65-byte EIP-712 signature over the statement was made by a specific EOA account. The function recovers the signer's address from the signature and compares it to the expected account. Both must match for the endpoint to be authenticated. Used to verify that both the shield (H1) and unshield (H6) EOAs have approved the disclosure statement.

**Inputs & Assumptions:**
- `domain` (DisclosureDomain): the pool deployment (chainId and verifyingContract). Trust: trusted (same as the statement).
- `statement` (DisclosureStatement): the statement signed. Trust: trusted.
- `account` (string): the expected EOA address (any case, the function normalizes via `getAddress()`). Trust: semi-trusted (external input, but the signature recovery checks it).
- `signature` (string): 0x-prefixed 65-byte hex string. Trust: untrusted (provided by the user or an external party).
- Implicit: Ethers' `verifyTypedData()` recovers the signer; `getAddress()` normalizes addresses to checksummed form.
- Preconditions: No validation. Assumes the signature is a valid 65-byte EIP-712 signature (it may not be, and recovery will fail). Assumes the account is a valid Ethereum address (may not be, and `getAddress()` will throw).

**Outputs & Effects:**
- Returns a boolean: true if the recovered address (normalized) matches the account (normalized), false otherwise.
- No state writes, no external calls (except to ethers).
- On success (returns true), the caller can trust that the EOA at `account` approved this statement.
- On failure (returns false), either the signature is invalid, the signature does not recover to `account`, or the address normalization failed.

**Block-by-Block:**

```typescript
// L117-L131
export function verifyEndpointSignature(
  domain: DisclosureDomain,
  statement: DisclosureStatement,
  account: string,
  signature: string,
): boolean {
  let recovered: string;
  try {
    recovered = verifyTypedData(
      eip712Domain(domain),
      DISCLOSURE_EIP712_TYPES,
      statementMessage(statement),
      signature,
    );
  } catch {
    return false;
  }
  try {
    return getAddress(recovered) === getAddress(account);
  } catch {
    return false;
  }
}
```

- **What:** Recover the signer from the EIP-712 signature and compare addresses.
- **Why here:** Two try-catch blocks separate recovery (L123-L128) from address normalization (L129-L132). Ethers' `verifyTypedData()` can throw if the signature is malformed or the recovery fails; `getAddress()` can throw if the address is invalid.
- **Assumes on entry:** The signature is a 65-byte EIP-712 signature (may be malformed). The account is a string (may not be a valid address).
- **Establishes on success path (L130):** The signature was created by signing the exact statement (same domain, types, message) with the private key of the address `account`. The signature is valid and recovers to that address.
- **Establishes on recovery failure (L127, first catch):** The signature is invalid, malformed, or does not recover to any valid address. The function returns false (no information is leaked).
- **Establishes on address comparison failure (L132, second catch):** One of the addresses (`recovered` or `account`) is not a valid Ethereum address, or the recovered address is not a valid address (though ethers guarantees `verifyTypedData()` returns an address string, this catch is defensive). The function returns false.
- **Depended on by:** builder.ts `buildDisclosurePackage()` (L107-L109) — the function is called once per endpoint EOA to verify the EOA has signed the statement. The signature is obtained via `signEndpoint()` callback (L109), so the caller is responsible for obtaining the signature from the endpoint.

**Path Analysis:**

1. **Happy path (signature is valid and recovers to account):**
   - `verifyTypedData()` succeeds (L123), recovers the signer address, assigns to `recovered`.
   - `getAddress(recovered)` normalizes the recovered address (L130).
   - `getAddress(account)` normalizes the expected account (L130).
   - Comparison returns true (L130).
   - Function returns true.

2. **Invalid signature path:**
   - `verifyTypedData()` throws (malformed signature, wrong signature, or recovery fails) (L123).
   - Catch block (L127) returns false immediately.
   - Second try-catch is not executed.

3. **Signature recovers to wrong address path:**
   - `verifyTypedData()` succeeds and recovers a different address (L123).
   - First try-catch succeeds (no exception).
   - `getAddress()` calls normalize both addresses (L130).
   - Comparison fails (L130), function returns false.

4. **Invalid address path (unlikely):**
   - `verifyTypedData()` succeeds (L123).
   - `getAddress()` throws when normalizing the recovered or expected address (L130).
   - Catch block (L132) returns false.

**Cross-Function Dependencies:**
- Callee `eip712Domain` (internal): constructs the domain object for signature recovery. Fully internal, deterministic.
- Callee `statementMessage` (internal): constructs the message object for signature recovery. Fully internal, deterministic.
- Callee `verifyTypedData` (ethers external-source-available): recovers the signer address from an EIP-712 signature. Takes domain, types, message, and signature; returns the signer's address as a string (checksummed or lowercase, ethers' behavior). Throws if the signature is invalid, malformed, or recovery fails. Per ethers docs, recovery is deterministic.
- Callee `getAddress` (ethers external-source-available): normalizes an Ethereum address to checksummed form. Takes any case, returns checksummed or throws if invalid. Per ethers docs, it is deterministic and does not require network access.
- Constant: `DISCLOSURE_EIP712_TYPES` (L17-L27) — the schema, used by `verifyTypedData()` to reconstruct the message hash.
- Callers: builder.ts `buildDisclosurePackage()` (L107-L109) — the function collects endpoint signatures and calls the `signEndpoint()` callback, which returns signatures. The SDK does not itself verify them; verification happens off-chain when the package is checked.

**Open Questions:**
- Ethers' `verifyTypedData()` behavior on recovery failure: does it throw an exception or return a special value? The SDK treats it as throwing (catches the exception). Per ethers v5 and v6 docs, `verifyTypedData()` (or `_TypedDataEncoder.recover()` in earlier versions) throws on recovery failure.
- The second try-catch (L129-L132) is defensive against invalid addresses, but ethers guarantees that `verifyTypedData()` returns a valid address string (never null or malformed). If the catch is truly unreachable, it is dead code; if `getAddress()` is called with a non-string or `recovered` is somehow invalid, this catch handles it. Confirm whether `recovered` can ever be invalid.
- The SDK's `buildDisclosurePackage()` does not call `verifyEndpointSignature()` itself; it trusts the caller to provide signatures from the correct endpoints via the `signEndpoint` callback (L109). Verification is deferred to off-chain verifiers. This is correct per the design (the SDK builds packages, verifiers check them), but note that the SDK does not self-verify endpoint signatures.
