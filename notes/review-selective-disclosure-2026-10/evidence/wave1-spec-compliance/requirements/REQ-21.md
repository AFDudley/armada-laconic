## REQ-21 — Hop H5: Channel pays a note

> "H5 · Channel pays a note | channel → payout note | The finalized outcome has a `SHIELD`
> allocation whose `destination` is the payout `npk`. The pool's `Shield` event in the same
> transaction shows the note. The user opens the `npk` as in H1. | No"
> — engineering/16-selective-disclosure.md, table "The hops"

**Verdict:** implemented · confidence: high

**Where enforcement lives — checker.ts, case `'H5'` (`L243-L261`):**

```ts
const outcome = params.source.finalizedOutcomesIn(hop.concludeTxid).find((o) => addressesEqual(o.channelId, hop.channelId));
const paysNote = outcome?.allocations.some(
  (a) => a.allocationType === SHIELD_ALLOCATION_TYPE && BigInt(a.destination) === BigInt(opening.npk),
);
...
const shieldErr = checkShieldLocation(opening, hop.shield, i);  // Shield event matches the opened note
```
`SHIELD_ALLOCATION_TYPE = 3` (`types.ts`) matches `ArmadaNitroAdjudicator.sol`'s
`SHIELD_ALLOCATION_TYPE = 3` / `uint8 public constant SHIELD = SHIELD_ALLOCATION_TYPE`
(`armada-nitro/src/ArmadaNitroAdjudicator.sol:L23,37`) — cross-checked, exact match. `checkOpening`
(same helper as H1) opens the `npk` "as in H1," per the requirement's last clause.

**How the verdict was reached:** each clause (SHIELD allocation to the payout npk; Shield event
in the stated transaction; opening reused from H1) maps to a cited line; the shared-constant
cross-check against the adjudicator contract confirms the allocation-type comparison is not an
off-by-value error.
