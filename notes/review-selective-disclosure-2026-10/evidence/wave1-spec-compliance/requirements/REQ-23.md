## REQ-23 — Worked example, verifier check 1: both EIP-712 signatures over S

> "1. Both EIP-712 signatures over S."
> — engineering/16-selective-disclosure.md, "What the verifier checks"

**Verdict:** implemented · confidence: high

**Where enforcement lives — checker.ts:L92-L99:**
```ts
for (const endpoint of pkg.statement.endpoints) {
  const sig = pkg.endpointSignatures.find((s) => s.role === endpoint.role && s.account === endpoint.account);
  if (sig === undefined) return failure('endpoint-signature', ...);
  if (!verifyEndpointSignature(pkg.domain, pkg.statement, endpoint.account, sig.signature)) {
    return failure('endpoint-signature', ...);
  }
}
```
Runs before any hop is checked, for every endpoint named in the statement (in the worked
example, exactly two: `0xA` and `0xB`).

**How the verdict was reached:** matches the requirement directly; shares its mechanism with
REQ-16.
