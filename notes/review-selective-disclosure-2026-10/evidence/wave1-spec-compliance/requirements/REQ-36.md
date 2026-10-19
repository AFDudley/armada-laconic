## REQ-36 — D20.3: spending key authorizes the statement; binding holds in both directions

> "**D20.3** The spending key authorizes each disclosure statement inside the disclosure proof.
> Each public address in the chain signs the same statement, so the binding between the
> addresses and the shielded wallet holds in both directions."
> — engineering/09-architecture-decisions.md, ADR-0020

**Verdict:** implemented · confidence: high

**What this demands:** (a) every disclosure proof in a package signs the *same* statement hash
S; (b) every endpoint EOA signs that same statement (not a per-hop or per-address variant).

**Where enforcement lives:** `buildDisclosurePackage` (`builder.ts`) computes `statementHash`
once (`statementHashField(params.domain, statement)`) before the proof loop, and passes that same
value into `formatDisclosureInput` for every entry in `params.spends` (`builder.ts`, the `for
(const spend of params.spends)` loop) — one S for every proof. The same `statement` object
(unchanged across the loop) is also what every `signEndpoint` call signs (`builder.ts`, the
`for (const endpoint of params.endpoints)` loop). On the checking side, `checkDisclosurePackage`
computes `statementHash` once (`checker.ts:L90`) and `verifyProof` rejects any proof whose own
`entry.statementHash` differs (`checker.ts:L121-L123`, check `proof-statement-hash`) — so a
package cannot smuggle in a proof signed over a different statement.

**How the verdict was reached:** both the builder (construction) and the checker (independent
verification) enforce "one statement, every signer," matching the requirement; the "both
directions" half is REQ-16.
