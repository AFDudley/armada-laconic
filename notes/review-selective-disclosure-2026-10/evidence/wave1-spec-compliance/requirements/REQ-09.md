## REQ-09 — Why S is public (binds verifier, context, hops, EOA addresses)

> "S is the hash of the full disclosure statement: the verifier, a context string, the list of
> hops and the EOA addresses. A proof for one statement cannot be reused in another statement."
> — engineering/16-selective-disclosure.md, "The disclosure circuit"

**Verdict:** implemented · confidence: high

**What this demands:** `statementHash` (S) MUST derive from a hash that covers the verifier
string, the context string, the hop list, and the EOA addresses — so a disclosure proof over S
cannot be replayed against a different statement.

**Where enforcement lives — armada-sdk/src/disclosure/statement.ts:**

```ts
// statement.ts
export function statementHashField(domain, statement) {
  return BigInt(statementDigest(domain, statement)) % SNARK_SCALAR_FIELD;
}
export function statementDigest(domain, statement) {
  return TypedDataEncoder.hash(eip712Domain(domain), DISCLOSURE_EIP712_TYPES, statementMessage(statement));
}
export function statementMessage(statement) {
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

`statementMessage` includes `verifier`, `context`, and `endpoints` (the EOA addresses) directly;
the hop list is bound through `hopsHash = keccak256(canonicalJson(hops))`
(`statement.ts`, `hashHops`). `checkDisclosurePackage` recomputes `hashHops(pkg.hops)` and
compares it with `pkg.statement.hopsHash` (`checker.ts:L89-L91`) before trusting any hop, so a
swapped hop list is caught before any proof is checked against it.

**How the verdict was reached:** every named component of S (verifier, context, hops, EOA
addresses) is traced to a field that `statementDigest`/`statementHashField` actually hashes over,
either directly or via `hopsHash`.
