## REQ-22 — Hop H6: Note leaves the pool

> "H6 · Note leaves the pool | note → EOA | H2 for the spent note, plus the `Unshield` event to
> the recipient EOA. The recipient EOA signs the statement. | No (uses H2)"
> — engineering/16-selective-disclosure.md, table "The hops"

**Verdict:** implemented · confidence: high

**Where enforcement lives — checker.ts, case `'H6'` (`L287-L300`):**

```ts
const proofErr = await verifyProof(proof, i);
const posErr = checkSpendLeaf(proof, opening, i);
const nullErr = checkNullifierEvent(proof, hop.unshieldTxid, i);  // "H2 for the spent note"
const paidOut = params.source.unshieldsIn(hop.unshieldTxid).some((u) => addressesEqual(u.to, endpoint.account));
if (!paidOut) { return failure('unshield-event', ...); }
```
Unlike `H4` (REQ-20), `H6` *does* re-run the full H2-equivalent check (`checkSpendLeaf` +
`checkNullifierEvent`) against its own transaction id (`hop.unshieldTxid`) rather than trusting a
separately-referenced `H2` hop — so the "H2 for the spent note" clause holds even if no separate
`H2` hop exists in the package. The recipient EOA's signature is checked package-wide
(REQ-16); `endpointByRole.get(hop.endpoint)` resolves which endpoint must have paid out.

**How the verdict was reached:** `H6` independently re-derives the H2 guarantee rather than
relying on hop-array composition, which is the correct and more robust pattern — see REQ-20 for
the contrasting case where `H4` does not do this.
