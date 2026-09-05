# 3. Context & Scope

arc42 §3 · **2026-09-05**

This section draws the system boundary and catalogues the external systems the build touches. Per [ADR-0013](./09-architecture-decisions.md#adr-0013), this is **one engineering team building the whole Armada product** — a Railgun-based privacy layer for USDC — reusing Laconic's prior work as the substrate. The boundary is therefore **reuse vs net-new**, not an Armada-vs-Laconic org split. The static decomposition of the in-scope side is the [§5 registry](./05-building-block-view.md); here we fix only the perimeter. Terminology follows [ADR-0010](./09-architecture-decisions.md#adr-0010).

## The product

Armada is pluggable asset-privacy infrastructure for USDC: routine flows (payments, payroll, treasury, yield) route through a single **immutable Railgun shielded pool** (one anonymity set), and privacy compounds as more flows join (`index.html`; Armada thesis, blog.armada.blue). The team builds this whole product — the Railgun-based settlement core (nitro-on-railgun), the watcher-party sync substrate, the yield & exchange venue and its adapters, and the mobile client — **reusing Laconic prior art** (nitro integration, watchers, wallet, chain-signatures, ingestion) throughout. There is no internal org boundary; the whole product is in scope.

## In scope — the whole product (by work package)

The four [work packages (§0)](./00-work-packages.md) are the entire scope; each owns a set of `T#.#` items.

| WP | §5 items | Note |
|---|---|---|
| **A · nitro-on-railgun** | T0.0–T0.3, T0.6 *(opt)*, T0.7; T6.2, T6.3 | Clean-room pool + circuits + own Phase-2 (ADR-0014); the deposit/payout boundary is the net-new integration contract (ADR-0002/0003/0004). |
| **B · watcher parties** | T1.0–T1.2, T2.0–T2.4; T6.1 | Proof-carrying ingest + metered private feeds; reuse the cerc-io watcher / ipld / nimbus lineage (ADR-0008/0009). |
| **C · yield & exchange** | T4.0–T4.6, T3.\* *(v2)*; **T5.0 adapters (CCTP built · Aave-v4 yield · Swaps), T5.1 routing** | Posted-price venue v1; adapters build-or-reuse Railgun's Cookbook recipe (license gate). **USDC yield is the priority** (ADR-0007/0011/0013). |
| **D · client** | T6.0, T6.4–T6.8 | Wallet/app, mobile transport, on-device proving (ADR-0008). |

## Reused, not built (prior art & reference)

Cited at pinned commits; never re-documented (ADR-0012 reuse-spec discipline).

| Component | Note |
|---|---|
| **Railgun design** (`contract` / `circuits-v2` spec; MIT `engine` / `cookbook` as reference) | **Reference spec only** — the shielded pool + JoinSplit circuits are **clean-room reimplemented** (T0.0/T0.1, spec-compatible, no Railgun-licensed source), not reused OSS (ADR-0014). |
| `go-nitro`, `watcher-ts`, `ts-nitro`, `@cerc-io/peer`, `mobymask`, `laconic-wallet`, `chain-signatures`, `ipld-eth-*` | Laconic / cerc-io prior art, reused at pinned commits. |
| **Perpetual Powers of Tau** Phase-1 | Inherited universal ceremony; we run only our own Phase-2 (ADR-0003). |
| Railgun **RelayAdapt / Cookbook** recipe | Railgun's own value-moving recipe; the C adapters may reuse it (license permitting), distinct from our T0.3 deposit/payout contract (ADR-0010). |

## External systems & interfaces (genuine third parties)

Direction is stated relative to our system.

| External system | Interface & direction | Touched by | Decision |
|---|---|---|---|
| **Ethereum L1** | Settlement root: our clean-room pool, `go-nitro` adjudicator, and deposit/payout contract live here. Bidirectional — write channel state, read commitment/nullifier/adjudicator events. | T0.0, T0.2, T0.3 (read at T1.2) | ADR-0002/0004/0014 |
| **Railgun deployed pool** (live) | Onboarding bridge: user **unshields from Railgun → shields into our pool** — one public boundary hop; a Railgun note is never spent/nullified in our contract. Inbound. | T0.7 | ADR-0006 |
| **Aave v4** | The yield protocol the USDC-yield rail reaches through, via our Aave-v4 adapter's recipe (`unshield USDC → supply → reshield aUSDC`). Amounts public only at the transparent supply boundary (Design A). | T5.0 (C) | ADR-0005 |
| **Circle CCTP** | Cross-chain USDC mint wrapped into the shield by our CCTP adapter — arrives as a fresh shielded USDC note. Inbound; adapter already built. | T5.0 (C), enters via T0.3 | — |
| **Public Waku network** | Transport underlay: pub/sub for feed/quote gossip + discovery; libp2p-noise direct streams for the 1:1 hot loop. Bidirectional P2P. Noise gives peer-auth, **not** IP privacy. | T2.3 (T6.5 mobile) | ADR-0008 |
| **nimbus-eth1** (upstream) | State-diff source: stateless/witness + Aristo, mapped into IPLD for proof-carrying diffs. Inbound; we build only enough local cache to serve, not a full archive EL. Upstream path in-progress (§11). | T1.0, T1.1 | ADR-0009 |

## Not in v1 / excluded

- Amount privacy in-play beyond Design A: **T0.6 fork-lite** is optional (Phase-4); **fork-full** is excluded unless separately greenlit (ADR-0005).
- **Market-making** (venue solver T4.2) and the **LP-buffered private USDC-yield rail** (T4.5) are **v2** (ADR-0011); v1 fills from a static inventory and does basic USDC yield via the direct adapter recipe.

---

→ Building-block detail and every `T#.#` id: [§5](./05-building-block-view.md). Scope decision: [ADR-0013](./09-architecture-decisions.md#adr-0013) (whole-product scope; reuse boundary). Adapter detail: [`T5-adapters.md`](./T5-adapters.md).
