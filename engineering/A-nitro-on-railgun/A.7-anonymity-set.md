# A.7 — Anonymity-Set Strategy

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T0.7 — strategy, not code

---

## A.7.1 Goal

A shielded pool's privacy is k-anonymity against the set of unspent, indistinguishable notes. The pool already hides sender, recipient, and amount in-circuit (`TransactModule`, `PrivacyPool.transact`); what it cannot manufacture is the crowd those hidden transactions blend into. A fresh pool starts near zero, which ADR-0006 records as the biggest non-engineering risk. T0.7 is a strategy item, not a contract, same as the earlier plan's framing, but the pool it strategizes around is built, not planned.

## A.7.2 The crowd grows only from the pool's own deposits

Every note a spend can hide among is a note in this pool's tree. No cross-pool proof and no Railgun onboarding bridge exist or are planned: a spend hides only among the notes it could consume, so a Railgun note that this pool accepted would let its owner spend the same value twice, and the pool would pay out tokens it never received (`labbook/2026-10-19-docs-versus-code.md` decision 2). The only lever is bringing in users and volume to this pool directly.

## A.7.3 The shielded-payments primitive — built

`PrivacyPool.transact` with `boundParams.unshield == NONE` is a native shielded transfer: inputs nullified, outputs created, value conservation checked in-circuit, with only `merkleRoot`, `boundParamsHash`, `nullifiers[]`, and output-commitment hashes public. No sender, recipient, or amount is exposed. The privacy set is the set of commitments in `PrivacyPool`'s accumulator (16-deep tree per `treeNumber`, `rootHistory`).

## A.7.4 Bootstrap — strategy, not built

Per ADR-0006, the crowd grows from the pool's own deposits, through incentivized early volume and, over time, a market-making or LP flywheel. This is program and economic policy, not contract code, and no bootstrap mechanism is built or specified in detail here.

**Moving from Railgun is a wallet feature (T6.7), not built yet.** It is a Railgun unshield followed by a shield into this pool. No contract of ours reads or touches Railgun's pool. Moved value counts towards this crowd only from the moment it is shielded here, like any other deposit. `armada-interface` has no Railgun-unshield code today; see [T6-client-apps.md](../T6-client-apps.md).

## A.7.5 Metrics — not instrumented

The earlier plan named three metrics: anonymity-set size (live commitment count), effective-k (the realistic crowd a spend hides in after correlation), and boundary-exposure rate (value transiting the shield/unshield and channel deposit/payout hops versus staying shielded). None of the three is instrumented in any repo checked. They remain a design target, computable from the pool's own events (`Shield`, `Transact`, `Nullified`, `rootHistory`) once a client reads them.

## A.7.6 Risks / open

- **Cold-start weakness window.** Until the pool's own volume grows, early users have a small crowd, and large or oddly denominated transactions are self-deanonymizing. There is no engineering fix; it is a bootstrap policy problem.
- **Boundary exposure versus liquidity.** Each shield, unshield, and channel deposit/payout is a transparent-amount event. More deposits raise set size and boundary-exposure together. No instrumented trade-off view exists yet.

See also: [T0-ethereum-anchor.md](../T0-ethereum-anchor.md), [T6-client-apps.md](../T6-client-apps.md) (T6.7, the move from Railgun), [ADR-0006](../09-architecture-decisions.md#adr-0006), `labbook/2026-10-19-docs-versus-code.md`.
