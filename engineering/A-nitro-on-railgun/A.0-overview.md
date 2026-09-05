# A.0 — nitro-on-railgun · Overview & Scope

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`../00-work-packages.md`](../00-work-packages.md) · [`../05-building-block-view.md`](../05-building-block-view.md)
**Owns:** T0.0 pool · T0.1 trusted setup · T0.2 adjudicator · **T0.3 deposit/payout contract (net-new)** · T0.6 native commitment *(opt)* · T0.7 anonymity-set strategy · T6.2 settlement client · T6.3 self-watchtower
**Baseline (frozen):** ADRs 0002–0011 · **operating model:** [ADR-0012](../09-architecture-decisions.md#adr-0012) · **scope:** [ADR-0013](../09-architecture-decisions.md#adr-0013)

---

## A.0.1 Goal

Deliver the **settlement substrate** of Armada — the Railgun-based rail on which payments, yield (C), and exchange (C) all clear: **notes in → normal Nitro → notes out**. A shielded Railgun note is unshielded into a net-new **deposit/payout contract** that escrows the value into a `go-nitro` state channel; parties settle off-chain via ForceMove; the channel outcome is re-shielded into fresh notes. A also delivers the base **shielded-payments** capability (native Railgun transfer) and the client that drives settlement + a self-watchtower.

Per [ADR-0012](../09-architecture-decisions.md#adr-0012) the **settlement rail** is **integration, not new cryptography**: the adjudicator is reused as-is (T0.2), and the net-new settlement code is the deposit/payout contract (T0.3), the T6.3 auto-watchtower loop, and a multi-asset ForceMove settlement app. **Caveat ([ADR-0014](../09-architecture-decisions.md#adr-0014)):** the pool (T0.0) and the JoinSplit circuits (T0.1) are **not** reuse — our engineers **clean-room reimplement** them, spec-compatible with the Railgun design the deep dive pins (A.1.1/A.1.2), using no Railgun-licensed source. That is a real, **audit-critical crypto-engineering** workstream (of a well-understood, de-risked design), so the "not new cryptography" thesis holds for the rail but **not** for the pool/circuits.

## A.0.2 The construction, concretely

The deep dive (A.1) pins the exact binding surface:

```
 shielded note ──transact() unshield-in──►  DEPOSIT (T0.3)
   RailgunSmartWallet.transact([...])           │ MultiAssetHolder.deposit(asset, channelId, expectedHeld, amount)
   unshieldPreimage.npk = T0.3 address           ▼
                                            go-nitro channel (holdings[asset][channelId])
                                                  │ off-chain ForceMove states / vouchers
                                                  ▼
 fresh notes ◄──shield() shield-out── PAYOUT (T0.3) ◄── concludeAndTransferAllAssets(...) → external destination
   RailgunSmartWallet.shield([ShieldRequest...])
```

Deposit-in = a Railgun `transact()` whose `unshieldPreimage.npk` is the T0.3 address, so `transferTokenOut` lands the ERC20 in T0.3, which calls `MultiAssetHolder.deposit`. Payout-out = the channel's `concludeAndTransferAllAssets` routes the outcome to T0.3 (an external destination), which `shield()`s fresh notes to each beneficiary. **T0.3 never touches pool or adjudicator internals** — it is a public-side ERC20 counterparty of unshield/shield and of channel deposit/conclude. Full citations: [A.1](./A.1-reuse-inventory.md).

## A.0.3 Document map

| Doc | Item(s) | Contents |
|---|---|---|
| [A.1](./A.1-reuse-inventory.md) | all | **The deep dive** — reuse inventory with pinned file/line citations (Railgun, go-nitro/ts-nitro, mobymask), corrections, net-new deltas, licensing |
| [A.2](./A.2-pool-deployment.md) | T0.0 | **Clean-room reimplement the pool** (spec-compatible with Railgun, [ADR-0014](../09-architecture-decisions.md#adr-0014)); fee=0; POI policy; token allow-list; vkey registration |
| [A.3](./A.3-trusted-setup.md) | T0.1 | **Clean-room JoinSplit circuits** ([ADR-0014](../09-architecture-decisions.md#adr-0014)); Phase-1 reuse + own Phase-2; circuit-set reconciliation (91 vs 54); artifacts |
| [A.4](./A.4-adjudicator-integration.md) | T0.2 | go-nitro adjudicator reuse; the ForceMove app choice; maturity gaps |
| [A.5](./A.5-deposit-payout-contract.md) | **T0.3** | The net-new contract: deposit-in, payout-out, outcome encoding, app registry |
| [A.6](./A.6-settlement-client-watchtower.md) | T6.2, T6.3 | ts-nitro client + submit-on-behalf keeper; the net-new auto-watchtower loop |
| [A.7](./A.7-anonymity-set.md) | T0.7 | Bootstrap + Railgun import bridge; POI binding; shielded-payments primitive |
| [A.8](./A.8-interfaces-acceptance.md) | ICD | What A exposes to B/C/D; the walking skeleton; acceptance/verification |
| [A.9](./A.9-native-commitment.md) | T0.6 *(opt)* | Amount-privacy-in-play; circuit delta ⇒ re-run T0.1 |

Each doc is written to the six-part reuse-spec bar ([ADR-0012](../09-architecture-decisions.md#adr-0012)): goal · boundary · reuse inventory · delta · ICD · acceptance.

## A.0.4 Walking skeleton (A's first integration target)

A thin end-to-end slice on a laconic fixturenet, exercising every A boundary once:

```
shield → transact()/unshield-in → MultiAssetHolder.deposit → trivial ForceMove settle
       → concludeAndTransferAllAssets → shield()/payout → scan
  T0.0        T0.3                         T0.2                T0.2/T0.3
```

Use a **trivial single-asset ForceMove app** (HashLockedSwap-grade) as a stand-in for C's real quote/settle app. Proves *notes-in-Nitro-notes-out* before deepening any item. Detail: [A.8](./A.8-interfaces-acceptance.md).

## A.0.5 Open gates (must resolve before / during A)

1. **✅ LICENSING — RESOLVED via clean-room ([ADR-0014](../09-architecture-decisions.md#adr-0014)).** Railgun's `circuits-v2` carries an explicit *"No License is provided for any party under any circumstances"* file and the pool contracts are SPDX `UNLICENSED` (A.1.10), so redeploying them was never executable. **Resolution:** our engineers **clean-room reimplement** the pool (T0.0) and the JoinSplit circuits (T0.1) — independently authored, spec-compatible with the Railgun design (A.1.1/A.1.2), using no Railgun-licensed source. This **unblocks A.2 and A.3** (they build our own contracts/circuits) and converts T0.0/T0.1 from reuse to audit-critical net-new. (`go-nitro`/`ts-nitro` are separately-licensed OSS and unaffected; `engine`/`cookbook` remain MIT reference.)
2. **Circuit-set count.** Source generates **91** `(nInputs,nOutputs)` combos; the widely-cited "~54" is a registered subset. T0.1/T0.0 must reconcile which subset to ceremony + register (A.3).
3. **POI is not on-chain.** POI is an alongside partner system, not a pool setting; T0.3 gated-entry is client/settlement-layer policy, and the POI node stack is a separate dependency (A.7).
4. **Multi-asset ForceMove app is net-new.** `MultiAssetHolder` supports multi-asset, but no shipped ForceMove *app* does ETH-in/USDC-out atomically (HashLockedSwap is single-asset/2-party). A.5/A.4.
5. **Watchtower liveness.** T6.3 auto challenge-response is net-new *and* needs an always-on node for the full challenge window (A.6).

## A.0.6 Provenance

Pinned commits for the deep dive: go-nitro **@435eb2b**, ts-nitro **@884d616**, mobymask **@2329198**. Railgun repos were read at `master`/`main` HEAD (2026-09-05) — **no SHA was pinned; A.2/A.3/A.5 must pin a Railgun commit before build.** Full citations: [A.1](./A.1-reuse-inventory.md).
