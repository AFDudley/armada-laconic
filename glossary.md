# Glossary

Terms span three worlds — **Vulcanize / Laconic / cerc-io** (the substrate), **Nitro** (state channels), and **Railgun / Armada** (the shielded pool) — plus this project's own design vocabulary. Several words collide with one another or with existing Ethereum and Railgun terminology, so the collisions are resolved first and enforced by [ADR-0010](./engineering/09-architecture-decisions.md#adr-0010); item ids (`T#.#`) point at the [§5 building-block registry](./engineering/05-building-block-view.md).

This file is the single canonical source of record. It is rendered to [`glossary.html`](./glossary.html) by `tools/render.sh`.

## 1. Word-collision resolutions (ADR-0010)

The choices below are fixed, and the documentation is kept grep-clean of the banned terms.

| Concept | Use this | Never this | Why / who owns the other name |
|---|---|---|---|
| T0.3 boundary contract funding/settling a Nitro channel from shielded notes | **deposit/payout contract** | "adapter" | "adapter" is overloaded (see the three rows below) |
| Railgun's own value-moving recipe (`unshield → call → reshield`) | **RelayAdapt / Adapt modules / Cookbook** | — | Railgun-only; never used to name our T0.3 or T5 |
| Armada's governance-added modules (CCTP, Aave-yield, Swaps) | **Adapters** (capital-A, tier **T5**) | "deposit/payout contract", "RelayAdapt" | Armada's app tier only (→ §5 T5) |
| Off-chain state indexer that ingests chain state and serves verifiable results | **watcher** (Laconic; **T2**) | watchtower, keeper | protects the **read** side of pool privacy |
| Bonded responder to a stale-state ForceMove challenge | **watchtower** (**T6.3**) | watcher, keeper | = watcher (read) **+** bonded responder (write) |
| Node that submits a user's tx on-chain and pays gas | **keeper** / **broadcaster** | watcher, relay | Laconic **keeper** ≡ Railgun **Broadcaster**; protects the **write** side |
| libp2p transport relay (NAT traversal, message routing) | **relay** (unqualified) | keeper, broadcaster | pure network plumbing, with no trust or on-chain role (**T2.3**) |
| This project's stack rungs (T0–T6), in the n-tier sense | **tier** | "layer" | "layer" collides with Ethereum L1/L2 and modular-DA vocabulary |
| Where trades clear (posted-price / matcher) | **venue** / **clearing** | "marketplace" | Armada needs clearing, not a service-provider marketplace |
| Boundary/interface where two analyses meet | **boundary** (or interface) | "seam" | "seam" collides with Feathers' testability term |
| Amount-privacy design | **Design A** / **fork-lite** / **fork-full** | — | see the design-term rows below |

**Read/write duality:** the watcher guards the read side, so scanning the pool leaves no RPC fingerprint, while the keeper/broadcaster guards the write side, so submitting a transaction leaves no EOA link. Both defend the same anonymity set from opposite ends. → §8, §5 T2/T6.

### Further term distinctions

These words also carry more than one meaning; the ADR-0010 table above governs, and these rows keep the remaining senses straight.

- **"Relay" — two unrelated things.** *Relay* (Laconic / libp2p) is a **libp2p transport relay**: routes p2p messages and does NAT traversal between browsers/phones and keepers — pure network plumbing, no trust, no on-chain role (**T2.3**). *Relay* (Railgun / Armada) is a **Broadcaster** (formerly "Relayer"): submits a user's encrypted ZK transaction on-chain and **pays gas**, so the public origin is the broadcaster's address, not the user's; talks to users over Waku. The Railgun relay's Laconic analog is the **keeper**, *not* the Laconic relay; unqualified "relay" here = the libp2p transport relay.
- **"Swap" — three senses.** *swap* (Nitro `protocols/swap`) — go-nitro's multi-asset off-chain swap protocol / `SwapChannel` settlement primitive. *Swaps* (Armada adapter) — the user-facing "swap shielded A→B" adapter, an application on the execution platform (**T5**). *matcher swap* — a fill from the ex_net matcher, settled via the Nitro swap primitive.
- **"Adapter" — three senses.** *Adapter* (Armada) — a governance-added module in Armada's Adapters tier (CCTP, Aave-v4 yield, Swaps; **T5**). *deposit/payout contract* — the net-new boundary contract funding/settling Nitro channels from Railgun notes (Design A privacy boundary; **T0.3**), never called "adapter". *`IForceMoveApp`* — a Nitro app-logic interface (e.g. `HashLockedSwap`); "adapter" in the Nitro sense.
- **"Validator" — two senses.** *Validator* (ex_net) — a bonded federation member that matches/prices/authorizes settlement; ≈ today's **watcher-party member**. *Validator* (Cosmos / laconicd) — a PoA consensus validator on the laconicd chain, unrelated to the exchange venue.
- **"Proof" — distinct artifacts.** *Liquidity proof* — signed attestation that assets are available to trade for a bounded time (ex_net). *Inclusion receipt* — threshold-signed proof an order-commit was included in an epoch (censorship becomes slashable). *Sequencing certificate* — threshold-signed proof of an epoch's fixed order. *Fraud proof* — on-chain evidence of misbehavior, slashable against a bond. *ZK proof / zk-SNARK* — Railgun's Groth16/BN254 proof of a valid shielded transaction.
- **"Pool" — two senses.** *Shielded pool* — Railgun's privacy pool (one shared anonymity set), Armada's core. *LP pool / vault* — market-making capital in the execution platform, unrelated to the shielded pool.
- **"Layer" vs "Tier".** *Tier (T0–T6)* — *this project's* stack rungs, in the software-architecture (n-tier) sense (see §5). *Layer 1 / Layer 2 (Ethereum)* — blockchains; we write "Ethereum L1" only where we mean the chain. *Execution / settlement / DA layers* — modular-blockchain vocabulary, deliberately *not* used to name our tiers.

## 2. Laconic / Vulcanize / cerc-io projects & components

Hosting: **GitHub** `cerc-io` where a mirror exists, otherwise the Laconic **Gitea** `git.vdb.to`. Orgs: **Vulcanize** (company, 2017–), **Laconic** (network/company), **cerc-io** (code org).

| Name | What it is |
|---|---|
| **ex_net** | Vulcanize's 2017 whitepaper: a bonded, federated, privacy-preserving cross-chain exchange venue — the design ancestor of the v2 matcher ([PDF](./ex_net_whitepaper.pdf)). Realized by **T4.6 + T3** (v2). |
| **matcher** | private Vulcanize repo (2019–20): curve-order, frequent-batch-auction matching-engine PoC — the engine behind ex_net. (Branches `master` / `ian_branch` / `xsleonard_branch`.) |
| **laconicd** | the Laconic Cosmos-SDK chain daemon; embeds a go-nitro server + onboarding module (`role` / `kyc_id`). Gitea. |
| **laconic-wallet** | React Native (Android/iOS) wallet: keys + Cosmos/EIP-155 signing over WalletConnect. Bare-bones today. Gitea. |
| **laconic-wallet-web** | browser build of the wallet. Gitea. |
| **watcher-ts** | the generic **watcher** framework (proof-carrying, metered state indexing). GitHub. |
| **mobymask** / v2 / v3 | reference p2p **demo** app (snap + in-browser Nitro + libp2p relay + keeper). A demo, not a product; proves the mobile-first shape. |
| **ts-nitro** | TypeScript port of Nitro running a state-channel node **in-browser / on mobile**. GitHub. |
| **go-nitro** | cerc-io fork of the (statechannels) **Nitro** Go impl: adjudicator, ForceMove, vouchers, virtual channels, swap protocol, Bridge. Live on Ethereum. |
| **chain-signatures** | library: `ethschnorr` (Ethereum-compatible Schnorr) + `ethdss` (Distributed Schnorr, Stinson–Strobl *(t,n)*) on a `kyber` DKG. Gitea. |
| **nimbus-eth1** | Nim Ethereum L1 execution client with stateless/witness execution + the Aristo state DB — emits canonical, proof-carrying MPT-node state diffs. GitHub. |
| **ipld-eth-server** / ipld-eth-state-snapshot | index and serve Ethereum state as IPLD. GitHub. |
| **laconic-so** | "stack orchestrator": deploys stacks (ops tool, not a library). |
| **Clearing (Nitro)** | Netting + non-custodial settlement of trades and fees over Nitro channels: per-request vouchers meter, co-signed channel allocations split fees (user / protocol / integrator), an L1 close settles in USDC. What the watcher party provides on the money side — *not* a service-provider marketplace or fee vault. |
| **Laconic Matchmaker** | "Laconic Matchmaking Services" (2022): privacy-preserving, broker-dealer-compliant matching via encrypted set-intersection / range matching (FHE). |
| **DSS** | Distributed Schnorr Signature (see chain-signatures); the federation's threshold signature. Realized by **T2.4** (attestation **T3.2**). |
| **Deepstack** | the `deep-stack` project; authored go-nitro's swap protocol. Distinct from the older statechannels HashLockedSwap. |

## 3. Nitro / state-channel terms

- **State channel** — off-chain, co-signed balance updates collateralized on L1; closed cooperatively at internet speed or unilaterally via dispute.
- **NitroAdjudicator** — the L1 contract that holds funds and adjudicates channel outcomes (**T0.2**).
- **ForceMove** — Nitro's dispute game: `challenge` → `checkpoint` → `conclude`. Only ever honors **user-signed** states (go-nitro **T0.2**).
- **MultiAssetHolder** — holds multiple assets per channel.
- **Virtual channel** (`virtualfund`) — a channel routed through a hub without a new on-chain deposit (go-nitro **T0.2**).
- **Voucher** — a signed pay-per-request micropayment over a (virtual) channel; the metering primitive (metering **T2**, fee-split **T4.3**).
- **Force-close** — unilateral exit to your last co-signed state (safety escape hatch: failure = halt, not loss).
- **HashLockedSwap** — an `IForceMoveApp` example: reveal a preimage to unlock (HTLC-style).
- **SwapChannel** — go-nitro's channel type for the swap protocol.
- **Bridge.sol** — mirrored L1↔L2 channel construction (cross-chain movement).
- **Watchtower** — see §1 (**T6.3**): a watcher (read) + a bonded responder (write) to a stale-state ForceMove challenge.

## 4. Railgun / Armada terms

- **Railgun** — Ethereum privacy protocol: a shielded pool using zk-SNARK circuits (BN254 / Groth16). Immutable core (**T0.0**).
- **Armada** — pluggable **asset-privacy infrastructure for USDC** on Railgun; three tiers (shielded pool / adapters / apps+SDK), crowdfunded, ARM-token DAO + steward. `@ship_armada`, `armada.wtf`. In testnet.
- **Shielded pool** — the shared privacy pool; privacy compounds with volume (one anonymity set). Railgun pool **T0.0**.
- **Shielded note / Note** — a shielded UTXO-like value commitment inside the pool, and the unit of shielded value. Railgun pool **T0.0**.
- **Anonymity set** — the crowd of shielded notes among which a spend is indistinguishable; privacy compounds with volume. Bootstrap own crowd + Railgun import bridge (**T0.7**, ADR-0006). Import ≠ set inheritance.
- **Commitment** — a note's on-chain hash (its value commitment), added to the pool's Merkle tree. Railgun pool **T0.0**; indexed by **T1.2**.
- **Nullifier** — a note's spend-marker; prevents double-spending without revealing which note was spent. Railgun pool **T0.0**; indexed by **T1.2**.
- **0zk address** — a Railgun private receiving address.
- **Broadcaster / Relayer** — see §1 ("Relay", Railgun); the **keeper**'s Railgun analog.
- **Waku** — decentralized libp2p-based p2p messaging (filter / lightpush / store), carrying broadcaster↔user comms and gossip/discovery. Transport **T2.3** (ADR-0008).
- **CCTP** — Circle's Cross-Chain Transfer Protocol (burn-and-mint native USDC); Armada's cross-chain Adapter, used as *mint-and-shield*. **T5** (Armada Adapters).
- **Aave-yield adapter** — USDC shielded yield via an **LP-buffered**, batched Aave rail (no per-user public deposit); ETH yield instead uses a shielded wstETH note. Armada's adapter tier (**T5**).
- **MASP** — Namada's Multi-Asset Shielded Pool note primitive — a note primitive, **not** a state channel (contrast Design B). Reference only.
- **RelayAdapt / recipe** — Railgun's cross-contract module: an atomic `unshield → external call → reshield` (Cookbook "recipes", e.g. an Aave deposit). The value-moving "adapter" — *not* a custodial vault, *not* DSS.
- **wstETH / weETH** — non-rebasing LSTs (value accrues via a rising exchange rate). Held as a shielded note they earn ETH staking yield intrinsically, with no Aave/LP rail needed. ETH yield **T4.4**.
- **POI** — Proofs of Innocence: an allow-list membership policy proving a note is not from a sanctioned or illicit source, without deanonymizing it. Our POI policy on the redeployed pool **T0.0** (ADR-0002).

## 5. This project's design terms

- **Tier (T0–T6)** — the stack: **T0** Ethereum anchor · **T1** ingestion · **T2** watcher-party substrate · **T3** ordering · **T4** execution platform (ex_net) · **T5** adapters · **T6** client/apps. Software-architecture tiers — not blockchain layers.
- **Watcher party** — a bonded federation of service providers serving one data need, and here also the execution venue. A sequencing-and-attestation role, **not** a BFT chain (**T2**, federation **T2.4**).
- **Proof-carrying feed** — a watcher feed whose query results ship with verifiable proofs (nimbus-eth1 state-diff → IPLD), so consumers verify rather than trust (**T1.1**, served over **T2**).
- **Commit-reveal** — fair ordering: `commit = H(order‖salt)` → seal epoch + inclusion receipts → post-seal beacon fixes positions → reveal → matcher runs. Blocks operator front-running without client-side ZK (ordering **T3** v2, ADR-0007).
- **Post-seal randomness beacon** — randomness revealed *after* the epoch is sealed, fixing order positions.
- **Epoch set-agreement** — the watcher party agreeing *which* commits belong to an epoch and in what order (leader-propose + threshold-attest + censorship fraud proofs). The core net-new, highest-risk piece (**T3.1**, v2).
- **Data availability (DA)** — publishing the revealed epoch so anyone can recompute/challenge.
- **Frequent batch auction (FBA)** — discrete per-block clearing at a uniform price, allocation pro-rata + lottery; defeats latency/HFT front-running (Budish–Cramton–Shim 2015). Matcher **T4.6** (v2).
- **Curve order** — an order specifying price-varying liquidity over a **range** (vs a point), flattened into fillable chunks. (Predates Uniswap v3 concentrated liquidity.)
- **Filler-filtered order** — an order carrying an allowlist of who may fill it; empty ⇒ open (lit/CLOB), non-empty ⇒ private/RFQ.
- **CLOB / RFQ** — Central Limit Order Book (public, lit) / Request-for-Quote (bilateral, private). The matcher unifies both in one order type.
- **LP vault** — pooled market-making capital run by a keeper (lets passive LPs participate).
- **Threshold Schnorr / DKG / t-of-n** — the federation signature (via chain-signatures); we pick **t high** (e.g. 4-of-7) so liveness failure = halt, not loss.
- **DSS** — Distributed Schnorr Signature: a Stinson–Strobl *(t,n)* threshold Schnorr over a `kyber` DKG (`chain-signatures`); the federation's threshold signature (**T2.4**, attestation **T3.2**).
- **Perpetual PoT (Phase-1)** — Community Perpetual Powers of Tau: the universal, inherited and vetted Groth16 trusted-setup phase. Reused as-is (**T0.1**, ADR-0003).
- **Phase-2** — per-circuit MPC over our circuit set; security is **1-of-N honest**, with a published transcript plus a final randomness beacon. Re-run only when circuits change (our own, **T0.1**, ADR-0003).
- **Design A / Design B** — **A** (this project, default v1): boundary privacy via a Railgun-style shielded pool + deposit/payout contract; the venue is a blind matcher but sees trade amounts/terms. Amounts are public only at the transparent boundary (ADR-0005). **B** (deferred, v3): ZK matching + "shielded ForceMove" to hide amounts/terms from the venue itself.
- **fork-lite** — native channelized commitment giving amount-privacy *during play*. A circuit change requires re-running Phase-2 (ADR-0003) plus a fresh audit. Optional Phase-4 upgrade **T0.6** (ADR-0005).
- **fork-full** — shielded ForceMove, research-grade; hides amounts in-play at the game level. Excluded unless separately greenlit (ADR-0005).
- **Posted-price venue (v1)** — a single `priceSetter` (one L1 wallet → governance) posts bid/ask on both sides, take-it-or-leave-it, cleared over Nitro. No order book, no price discovery — nothing to front-run; discovery is the v2 matcher (**T4.0**, ADR-0007).
- **Taker / maker (LP)** — **taker**: swaps notes privately at the posted price, capacity-bounded and fully private. **maker/LP**: provides inventory, holds a public aggregate position, and earns the spread — voluntary and compensated, never forced into a public deposit. Saturating takers are *offered* the maker role. Venue **T4**.
- **v1 / v2 / v3** — the scope split: **v1** = Armada support (posted price + clearing + private sync + mobile client + CCTP + boundary + base yield); **v2** = ex_net matcher (price discovery + fair ordering); **v3** = Design B. See [Build plan · versions](./build-plan.html#versions) (→ §4, build-plan).

## Orgs

- **Vulcanize** — the originating company (ex_net, matcher, vulcanizedb lineage; 2017–).
- **Laconic** — the network/company; **cerc-io** is its code org; **git.vdb.to** its Gitea.
- **Deepstack** (`deep-stack`) — authored go-nitro's swap protocol.

## Cross-references

- The collision policy is fixed by **[ADR-0010](./engineering/09-architecture-decisions.md#adr-0010)**; cross-cutting concepts that reuse these terms are described once in **→ §8** and point at their `T#.#` realizations.
- Item ids (`T#.#`) resolve in the **[§5 building-block registry](./engineering/05-building-block-view.md)**.
- Companion to the [Overview](./index.html), [Architecture](./architecture.html), [Build plan](./build-plan.html), and [Execution platform](./execution-platform.html). Internal Google Docs require Laconic/Vulcanize access.
