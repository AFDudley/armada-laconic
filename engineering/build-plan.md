# Build Plan — status, effort & sequencing

status companion · 2026-09-10 · reconciled to §5 registry, §9 ADRs, §11 risks

The [Building Block View (§5)](./05-building-block-view.md) is the structure and the [ADRs (§9)](./09-architecture-decisions.md) are the decisions; this page is the **status, effort, and sequencing** view over the same `T#.#` items. Every status/release facet matches the §5 registry, and every long pole matches the §11 risk register. This supersedes the earlier root `build-plan.html`, which predated the own-pool decision ([ADR-0014](./09-architecture-decisions.md#adr-0014)) and the matcher/ordering→v2 re-scoping ([ADR-0007](./09-architecture-decisions.md#adr-0007)/[0011](./09-architecture-decisions.md#adr-0011)).

## Version scope

| Version | Scope | Tiers / items |
|---|---|---|
| **v1** — the private rail | Private, mobile-first USDC rails: shielded pool + settlement + posted-price venue + yield, on a phone | T0.0–0.3, T0.7 · T1 · T2.0–2.3 · T4.0/4.1/4.3/4.4 · T5.0/5.1 · T6.0–6.7 |
| **v1.5** — private cross-chain swap | Private cross-chain swap over nitro-railgun channels + fronting hubs; a ~10–12-point increment on the finished v1 rail, ahead of the v2 research work | ADR-0015/0016: nitro-railgun adjudicator + DSS custodian |
| **v2** — matcher, fair-ordering & bonded federation | Price *discovery* + market-making + provably-fair ordering + economic security | T3.\* · T4.2/4.5/4.6 · T0.4/0.5 · T2.4 |
| **optional** | Amount-privacy-in-play; thin identity | T0.6 fork-lite · T6.8 |

**Why the split de-risks v1** (ADR-0007/0011): a posted-price, take-it-or-leave-it venue has nothing to front-run, so v1 needs neither commit-reveal ordering nor epoch set-agreement. Those two highest-risk, research-grade items move to v2 with the matcher. v1 fills exchange from a static pre-funded inventory, with no market-making. The one genuinely novel-crypto item on the v1 path is the **own pool + circuits** (ADR-0014); everything else is integration plus small contracts. (The earlier doc's "v3 — Design B", hiding amounts from the venue via shielded ForceMove, is the excluded fork-full of ADR-0005; the lighter T0.6 fork-lite is the *optional* row above.)

## Status by tier

Facets: **status** (built · reuse · net-new · partial) · **confidence** (`validated` = code read/grounded this cycle · `design` = specified). The **A deep dive** ([A.1](./A-nitro-on-railgun/A.1-reuse-inventory.md)) and the **nitro-bridge/DSS audit** ([audit](./nitro-bridge-audit.md)) code-read the settlement rail (go-nitro/ts-nitro) and the DSS crypto; the own pool/circuits are `design`-level against a pinned reference spec.

### T0 · Ethereum anchor
- **T0.0 shielded pool — net-new** (ADR-0014): our own implementation of the Railgun design, matching its note/commitment/nullifier format and `snarkSafetyVector`. The single largest, audit-critical build. `design`
- **T0.1 circuits + trusted setup — net-new** (ADR-0014): our own JoinSplit circuits + a Phase-2 MPC over the community Phase-1 (never re-run). Audit-critical; ceremony is a process risk (§11 R6). `design`
- **T0.2 Nitro adjudicator — reuse** go-nitro `@435eb2b` (ForceMove / MultiAssetHolder), live on Ethereum. `validated`
- **T0.3 deposit/payout contract — net-new** — the notes-in / notes-out boundary; **never an "adapter"** (ADR-0010). Spend-authorizing (§11 R5). `design`
- **T0.7 anonymity-set strategy — process** — bootstrap the crowd + the Railgun onboarding import bridge (§11 R1). `design`

### T1 · Ingestion — **the v1 long pole (§11 R3)**
- **T1.0 nimbus-eth1 state-diff emitter → IPLD — net-new**; the stateless/witness + Aristo path is upstream-in-progress. The one genuinely unbuilt v1-spine tier; a stale feed is a **safety** problem for the watchtower, not just latency. `design`
- **T1.1 IPLD proof-carrying diffs — partial** — reuse cerc-io `ipld-eth-*` codecs/backfill. `design`
- **T1.2 watcher ingest config — net-new** — index T0.0 commitments/nullifiers + T0.2/T0.3 events; exposes a head cursor T6.3 gates on. `design`

### T2 · Watcher substrate — mostly reuse (cheapest tier)
- **T2.0 proof-carrying feeds — reuse+config** (watcher-ts `getStorageAt → {value, proof}`). **T2.1 metering — reuse+config** (go-nitro vouchers via `payments.ts`). **T2.2 peer substrate — reuse** (ts-nitro / `@cerc-io/peer`). **T2.3 transport — reuse** (Waku pub/sub + libp2p-noise; **optional Nym underlay**, ADR-0008). `T2.1/T2.2 validated; feeds/transport design`
- **T2.4 federation + bond + threshold DKG — v2**: chain-signatures Schnorr lib built + unit-tested; DKG ceremony/resharing, bonding, and signing-path wiring are net-new. v1 runs an **unbonded/trusted** federation.

### T4 · Execution / venue
- **T4.0 posted-price contract — net-new (small)** — single `priceSetter` → governance (§11 R8, ADR-0007). **T4.1 quote/settle ForceMove app — net-new** — HashLockedSwap-grade; the **multi-asset ETH-in/USDC-out app** is the go-nitro maturity gap (§11 R2). **T4.3 fee-split — net-new (small)**. **T4.4 ETH/wstETH yield — config** (held as a non-rebasing note). `go-nitro channels validated`
- **v2:** venue solver / market-making (T4.2), LP-buffered USDC-yield rail (T4.5), ex_net matcher + LP vault (T4.6) — ADR-0011.

### T5 · Adapters
- **T5.0 Adapters — CCTP (built) · Aave-v4 yield · Swaps** — Aave/Swaps build or adapt a Cookbook/RelayAdapt-style recipe. **USDC yield is in v1** via the Aave-v4 adapter; the v1 *mechanism* is open (direct adapter recipe over the public Design-A boundary vs the v2 LP-buffered rail T4.5). **T5.1 routing interface — interface** (how the venue routes to/from the Adapters). `design`

### T6 · Client / apps — **mobile transport is the crux (§11 R4)**
- **Reality check:** the wallet is bare-bones (laconic-wallet fork) and the swap/mobymask demos are fragments, not products — a starting point, not a front-end.
- **T6.0 wallet base — partial/reuse** (BIP-39/HD, secure enclave; §11 R5) · **T6.1 WASM note-scanner + key bridge — net-new** (audit surface) · **T6.2 settlement client — net-new** · **T6.3 self-hosted watchtower — net-new** · **T6.4 address rotation — net-new (small)** · **T6.5 native mobile transport — net-new, the crux** (go-nitro + go-waku gomobile; RN interim WebView) · **T6.6 Groth16 mobile proving — reuse (constraint)** · **T6.7 Armada app + SDK — product**. `wallet/ts-nitro validated; rest design`

## Effort — linear difficulty points

A **linear 1–10** relative-difficulty estimate (1 = trivial config/reuse; 10 = massive, audit-critical, novel; a 10 is ~10× a 1). Weighted by net-new-ness + audit + novelty, **not** calendar time. The nitro-railgun adjudicator (v1.5) is included as an extra line.

| Item | Status | Pts |
|---|---|---:|
| T0.0 shielded pool | net-new | **10** |
| T0.1 circuits + trusted setup | net-new | **9** |
| T0.2 Nitro adjudicator | reuse | 2 |
| T0.3 deposit/payout contract | net-new | 5 |
| T0.7 anon-set strategy + import bridge | process | 3 |
| T1.0 nimbus state-diff emitter | net-new (long pole) | **8** |
| T1.1 IPLD proof-carrying diffs | partial | 4 |
| T1.2 watcher ingest config | net-new | 3 |
| T2.0 proof-carrying feeds | reuse+config | 2 |
| T2.1 voucher metering | reuse+config | 2 |
| T2.2 P2P peer substrate | reuse | 2 |
| T2.3 transport (+ optional Nym) | reuse | 3 |
| T4.0 posted-price contract | net-new | 3 |
| T4.1 quote/settle app | net-new | 4 |
| T4.3 fee-split | net-new | 3 |
| T4.4 ETH/wstETH yield | config | 2 |
| T5.0 Adapters (CCTP · Aave · Swaps) | reuse/net-new | 5 |
| T5.1 routing interface | interface | 2 |
| T6.0 wallet base | partial/reuse | 4 |
| T6.1 WASM note-scanner | net-new | 5 |
| T6.2 settlement client | net-new | 5 |
| T6.3 self-hosted watchtower | net-new | 5 |
| T6.4 address rotation | net-new | 2 |
| T6.5 native mobile transport | net-new (crux) | **8** |
| T6.6 Groth16 mobile proving | reuse (constraint) | 5 |
| T6.7 Armada app + SDK | product | 4 |
| **v1 subtotal** | | **110** |
| nitro-railgun adjudicator *(v1.5)* | net-new, audit-critical | **6** |
| **Total** | | **116** |

**~35% concentrates in five items** — pool (10) + circuits (9) + mobile transport (8) + nimbus ingestion (8) + adjudicator (6) = 41. The client tier (T6) is the biggest bucket (38, many moderate net-new pieces gated by the crux); the watcher substrate (T2) is cheapest (9, almost all reuse). The points weight *audit* heavily, and the two partials (T1.0 nimbus, T6.0 wallet) carry upstream-maturity schedule risk the raw points do not.

**The v1.5 cross-chain-swap increment.** Given Armada already deploys a pool per chain, the *marginal* engineering to add private cross-chain swaps on top of a finished v1 is small: the nitro-railgun adjudicator (6) + the hash-locked swap app (~2) + DSS hub wiring (~3–4), ≈ **10–12 points**, versus v1's 110 and the research-grade v2 lift. It reuses the whole v1 rail and has no hard v2 dependency (no ordering, no matcher, no market-making, no bonded federation; hubs front from static inventory and can run a single-entity DSS). Three caveats bound it: (1) it is all spend-authorizing, so the real cost is audit rather than code; (2) each additional chain re-incurs the anonymity-set cold-start (§11 R1), gating the privacy payoff per chain; (3) the DSS crypto is currently unaudited with stubbed production wiring — wiring and auditing it is real, but a subset of v2's federation work (key-security without bonding/slashing).

## Long poles (ranked, §11)

1. **Own pool + circuits** — T0.0/T0.1 (R9): the largest, audit-critical crypto-engineering build; a spec deviation is a soundness/fund-safety bug.
2. **Mobile transport** — T6.5 (R4): RN-native gomobile; WebView interim, native is a Phase-3 gate.
3. **T1 nimbus ingestion** — T1.0 (R3): upstream-dependent; staleness breaks watchtower safety.
4. **Audit surface** — T0.3 / T0.0 / T0.1 / T6.1 / T6.0 (R5): external audit gates mainnet.
5. **go-nitro maturity** — the multi-asset ForceMove app + dispute wiring driven end-to-end (R2).
6. **Ceremony logistics** — T0.1 (R6): ≥5 independent contributors, published transcript + beacon.

## Sequenced backlog

1. **Walking skeleton** (§4): shield → deposit → trivial ForceMove settle → payout → scan on a laconic fixturenet — retires integration risk before any tier deepens.
2. **Spike the long poles**: nimbus emitter (T1.0), the native mobile module (T6.5), the multi-asset settlement app (go-nitro, R2).
3. **Own-pool build**: pool (T0.0) + circuits/ceremony (T0.1) against the A.1 reference spec; freeze the spend-authorizing surface (T0.3) early.
4. **Venue + Adapters**: posted-price contract + quote/settle app + fee-split (T4.0/4.1/4.3) over the rail; ETH yield (T4.4) + USDC via the Aave adapter (T5.0/5.1).
5. **Wallet** (T6.\*): WebView MVP, then native; note-scanner, settlement client, self-watchtower.
6. **v1.5 — cross-chain shielded swap**: the nitro-railgun adjudicator + DSS custodian, deployed per chain (ADR-0015/0016; [construction](./shielded-nitro-bridge-design.md), [audit](./nitro-bridge-audit.md)) — a ~10–12-point increment ahead of the v2 research work.
7. **v2**: bonded federation + DKG (T2.4/T0.4/T0.5), ordering (T3), matcher + market-making (T4.2/4.5/4.6).

## Pinned commits & reference code

go-nitro `435eb2b` · ts-nitro `884d616` · mobymask `2329198` · chain-signatures `9016a7c` · laconicd `d130608` (`roysc/nitro-integration`) · watcher-ts `18ca4e1` · ipld-eth-server `330bc3d` · laconic-wallet `bb5223a` · laconic-wallet-web `2a4a478` · nimbus-eth1 (`status-im`, upstream). Railgun `Railgun-Privacy/contract` + `circuits-v2` are the **design reference** (pin a commit before build, ADR-0014). Full file/line citations: [A.1 reuse inventory](./A-nitro-on-railgun/A.1-reuse-inventory.md) and the [nitro-bridge + DSS audit](./nitro-bridge-audit.md).

---

Cross-refs: registry/status → [§5](./05-building-block-view.md); decisions → [§9](./09-architecture-decisions.md); risks/long-poles → [§11](./11-risks-and-technical-debt.md); solution strategy → [§4](./04-solution-strategy.md); cross-chain swap → [the construction](./shielded-nitro-bridge-design.md).
