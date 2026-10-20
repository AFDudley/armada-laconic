# A.9 — Native channelized commitment (amount-privacy-in-play)

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T0.6 — not built, deferred

---

## A.9.1 Status

T0.6 remains exactly what ADR-0005 decided: deferred, not on the spine, not greenlit. No code exists for it. This page records the plan so it can be picked up if greenlit; it is not a build in progress.

## A.9.2 What it would do

Today, amounts cross the shield/unshield boundary and the adjudicator's deposit/payout boundary in the clear (Design A, [T0-ethereum-anchor.md](../T0-ethereum-anchor.md)). A force-close also leaves cleartext allocation amounts on-chain, because `MultiAssetHolder.holdings` and `ExitFormat.Allocation.amount` are plain `uint256`. Native channelized commitment would make a channel allocation reference a hidden-amount commitment instead, so a force-close reveals no amount. The shield/unshield boundary itself would stay public; only the force-close leak would close.

## A.9.3 What it would need

1. **A circuit change.** Extend the `Transact` circuit's note/commitment format so an allocation carries a hidden-amount commitment, with an in-circuit proof that committed allocation amounts sum to the committed channel deposit.
2. **A Phase-2 ceremony re-run** (ADR-0003 D3.4), on top of whatever shapes the production ceremony has by then covered.
3. **A change to the adjudicator's payout branch** (`_executeSingleAssetExit`'s `SHIELD` path, [A.5](./A.5-deposit-payout-contract.md)) to decode the new allocation encoding and re-shield against hidden-amount commitments.
4. **A fresh audit**, because any circuit change invalidates prior audit coverage.
5. **Heavier browser or mobile proving**, since a wider circuit costs more witness and proving time.

## A.9.4 Why it is still deferred

ADR-0005's reasoning holds unchanged: Design A ships the full v1 feature set with no new crypto or ceremony. Greenlighting T0.6 needs a concrete requirement — a venue, a counterparty, or a regulatory posture — that cannot tolerate amounts on-chain during a force-close, plus budget for a circuit audit and a ceremony re-run. No such requirement has been named. Fork-full (a fully shielded ForceMove dispute game) stays excluded regardless; `labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md` records the operator's rejection of a complete per-state-change proof as "the worst of both worlds."

## A.9.5 Risks / open

- **Cost versus benefit is unchanged from ADR-0005.** No new evidence has shifted this calculus.
- **Native ETH is a separate, smaller item.** The pool is ERC-20/ERC-721 only; a WETH-wrap path needs no circuit change and is not T0.6.

---

See also: [A.5](./A.5-deposit-payout-contract.md) (the payout branch this would change), [A.3](./A.3-trusted-setup.md) (the ceremony this would re-trigger), [ADR-0005](../09-architecture-decisions.md#adr-0005), [ADR-0003](../09-architecture-decisions.md#adr-0003), `labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md`.
