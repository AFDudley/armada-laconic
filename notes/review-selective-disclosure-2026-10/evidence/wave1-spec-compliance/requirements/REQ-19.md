## REQ-19 — Hop H3: Spend creates note

> "H3 · Spend creates note | spend → output note | The user opens the output commitment: `npk`,
> token, value and `random`. The verifier recomputes the hash and compares it with the `Transact`
> event. | No"
> — engineering/16-selective-disclosure.md, table "The hops"

**Verdict:** implemented · confidence: high

**Where enforcement lives — checker.ts, case `'H3'` (`L180-L196`):**

```ts
const openErr = checkOpening(opening, i);   // npk == Poseidon(M, random)
if (opening.tree !== hop.transact.tree || opening.position !== hop.transact.position) { ... }
const record = params.source.transactsIn(hop.transact.txid).find((t) => t.tree === ... && t.position === ...);
if (BigInt(`0x${record.hash}`) !== commitmentHash(opening)) { ... }
```
`commitmentHash` (`checker.ts`) is `ShieldNote.getShieldNoteHash(npk, tokenHash, value)` —
recomputes the full commitment (`npk`, token, value) and compares it against the `Transact`
event's recorded hash at the stated leaf. Matches the requirement exactly, including "No new
component" (no disclosure proof is used for H3).

**How the verdict was reached:** the requirement's two clauses (open the commitment; compare
with the `Transact` event) map 1:1 to the two comparisons shown; confirmed by the SDK's "H3
output-note opening" test (`armada-sdk/src/disclosure/disclosure.test.ts:L315-L336`), which
exercises both the acceptance and the mismatch path.
