# T4 · Execution / venue

**Status:** draft · tier doc · 2026-09-04
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T4
**Release:** v1 (T4.0, T4.1, T4.3, T4.4) · v2 (T4.2, T4.5, T4.6)
**Depends on / Blocks:** consumes T0.3 (deposit/payout + ForceMove app registry + channel lifecycle + vouchers, incl. multi-asset outcomes), T0.0 (pool + asset allow-list must include the pair), T2 (proof-carrying feeds + Nitro metering); routes to Armada yield via T5.1; T4.6 blocked by T3 (fair ordering). Publishes to T6 (quote RPC, posted price, app id, loyalty-key scheme, capacity hint).

T4 is where the spine becomes a product. Idle shielded value earns yield and swaps A→B **privately, at a posted price, in the user's custody throughout**. The v1 venue **posts a price and fills from a simple pre-funded shielded inventory** — an Armada-operated account holding both sides of a pair, topped up out-of-band — and settles every fill non-custodially over the T0.3 rail. This is **clearing, not a marketplace** (ADR-0007): the running, netted accounting of who owes what, settled over Nitro, never a custodial vault and never a service-provider directory. **Automated market-making** — the aggregating solver that internalizes flow, hedges the residual, and rebalances Aave (T4.2), together with the LP-buffered USDC-yield rail that rides it (T4.5) — is **out of v1 and deferred to v2** (ADR-0011).

v1 is an **RFQ posted-price dealer market** (ADR-0007): Armada posts both sides take-it-or-leave-it and corrects a wrong price by reposting. A provably-fair CLOB with real price discovery is deferred to **T4.6 + T3**. The end-to-end v1 deliverable is a private **shielded ETH-yield note (wstETH) → shielded USDC** swap on a fixturenet: the user deposits into a channel (T0.3), quotes and settles at the posted price against the venue's **pre-funded USDC inventory**, and the payout endpoint mints fresh USDC notes — with no linkage back to the deposit note beyond the boundary amount (Design A, ADR-0005). No hedging or rebalancing runs in v1; when inventory runs low the venue simply withdraws quotes until it is refilled.

The value-moving construction, step by step (the yield → USDC route):

1. **Note → channel** (deposit, T0.3): the user unshields a shielded note (wstETH or USDC) into the deposit endpoint, opening and funding a Nitro channel. This is the one public boundary amount (Design A).
2. **Quote** (T4.1): the wallet requests a quote, and the venue returns the current signed posted bid/ask (T4.0). Take-it-or-leave-it.
3. **Settle** (T4.1): acceptance drives the quote/settle ForceMove app to a co-signed fill state — user leg out, venue leg in, at the posted price.
4. **Fill from inventory** (T4.0/T4.1, v1): the venue fills the leg from its pre-funded static inventory at the posted price. Automated netting/hedging/rebalancing (T4.2) and the LP-buffered USDC rail (T4.5) are v2 (ADR-0011).
5. **Payout → fresh notes** (T0.3): channel finalize routes the outcome through the payout endpoint, shielding fresh USDC notes to the user — a fresh-shield exit that preserves POI (T0.0).

## T4.0 Posted-price contract

**Status:** net-new · v1.

The on-chain price surface. A single privileged `priceSetter` holds the current bid/ask per pair — one ETH L1 wallet to start, **migratable to a governance address later** via the same admin-key→governance path used for the T0.0 POI update authority and the T0.3 game allow-list.

- **What it is:** the venue holds the current posted price, executes fills at it, and clears and settles over Nitro. **No order book, no auction, no queue.** A posted, take-it-or-leave-it price has **nothing to front-run**, which is precisely what removes the entire fair-ordering problem (commit-reveal, sequencing, epoch set-agreement) from v1 and defers it to T4.6.
- **Reuse vs build:** build — a small contract plus poster (net-new, small).
- **Interface exposed:** `priceSetter`-gated bid/ask storage per pair, with a read path the T4.1 settle app consults. The posted-price contract address plus the `priceSetter`/governance authority are **published to T6** so the wallet can display and verify the price it is filling against.
- **Key tasks:** bid/ask storage per pair; governance-migratable authority; the read path for the settle app.
- **Risk:** a wrong price means bad fills, **not lost custody** — fills are take-it-or-leave-it and force-closable via T0.3. Single-`priceSetter` trust is reduced by the governance migration and removed entirely only by the T4.6 CLOB upgrade, at the cost of months of work.

## T4.1 Quote/settle ForceMove app

**Status:** net-new · v1.

A **pure state machine** — states `propose-quote → accept → fill → settle/abort` — registered in the **T0.3 ForceMove app registry**, where its app ID occupies an allow-listed slot.

- **What it is:** it drives a channel to a co-signed fill state at the posted price — user leg out, venue leg in. Acceptance of a signed quote (T4.0) advances the app, and a **missing settlement receipt forces a close to the pre-fill state**, so a stalled settle never costs funds.
- **Reuse vs build:** build the app (pure state machine). The adjudication and dispute machinery is reused from T0.2 (`ForceMove.sol` / `NitroAdjudicator.sol`), and `protocols/swap/swap.go` in go-nitro is the reference for the ETH-in/USDC-out multi-asset fill.
- **Interface consumed:** the T0.3 channel lifecycle API (open/fund-from-note, propose/accept state, cooperative close, force-close, respond), the app-registry slot, and the T4.0 posted-price read path.
- **Interface published:** the quote/settle app ID plus the adjudication interface go to T6; the wallet drives the fill state machine directly.
- **Key tasks:** define the states; register the app ID in the T0.3 allow-list; wire the read of the T4.0 price into the fill transition.
- **Testing:** accept-at-posted-price fills; a stale or withdrawn quote is rejected; a missing settlement receipt forces a close to the pre-fill state with no fund loss.

## T4.2 Venue solver

**Status:** net-new · **v2** (ADR-0011).

In **v1 the venue fills at the posted price from a simple pre-funded static inventory — no solver runs.** T4.2 is the **v2** automated market-maker: an **off-chain keeper**, the maker brain. It maintains **shielded inventory** (for example, shielded aUSDC / wstETH notes backed by public positions), nets most fills internally, computes the residual, hedges only that residual on public liquidity, and then re-shields. Quote-based market-making is structurally more capital-efficient than an AMM, with no dead curve and no forced LVR.

- **Internalizes flow:** most fills net against inventory; only the **residual** touches a public venue.
- **Hedges the residual:** a venue holding shielded assets can quote across venues by borrowing public liquidity to close the residual **atomically** and then reshielding — private quotes across venues using internal balancing, without exposing per-user flow. The value-moving construction for the atomic hedge leg is **Railgun's own RelayAdapt / Cookbook recipe** (`unshield → external call → reshield`, atomic, non-custodial, origin-private via a broadcaster). It is **distinct from the T0.3 deposit/payout contract**, which alone gates pool spend authority.
- **Rebalances:** the solver rebalances to and from Aave (via T5.1 → T5.0) on a **decorrelated schedule**, one batched public tx tied to no single user.
- **Sees only a shielded note address**, never a user identity; cross-fill profiling is defeated in the wallet by T6.4 address rotation. A user who *wants* continuity presents an **optional, venue-only linkable loyalty key** from T6.4 — opt-in, venue-scoped, never global.
- **Non-custody invariant:** every fill is atomic against the venue's *own* inventory over a channel; the venue is a counterparty, not a vault. Custody stays with the user (a shielded note) before and after. The venue's aggregate hedge/Aave position is public by the nature of market-making, but decorrelated from any single user's amount or timing (aggregation, not buckets).

- **Reuse vs build:** build the keeper (net-new); reuse the RelayAdapt/Cookbook recipe for the atomic hedge and `waku-broadcaster-client` for origin-private submission and quote gossip.
- **Interface consumed:** T2 proof-carrying feeds (pool commitments/nullifiers, adjudicator events) that the solver watches; the T5.1 routing interface to the Armada Swaps/Aave/CCTP adapters (T5.0) for the hedge and rebalance legs.
- **Key tasks:** maintain shielded inventory; net fills; compute the residual; execute the RelayAdapt hedge; rebalance to and from Aave on a decorrelated schedule.
- **Risk:** the residual hedge is a *public* tx that can be front-run or MEV'd, and it depends on the public venue's liquidity. Mitigate by internalizing as much flow as possible (deep inventory), batching and decorrelating the residual, and routing the RelayAdapt leg through a broadcaster so its origin is not the venue's operational address. Per-venue inventory also splinters liquidity; an **opt-in shared hub / pooled-maker vault** (passive LPs first-class) deepens depth without forcing consolidation.

## T4.3 Fee-split allocation

**Status:** net-new · v1.

Both the trade and the venue's fee **clear the same way**: each metered fill updates a co-signed **channel allocation** that splits value across **user / protocol / integrator**.

- **What it is:** the integrator revenue-share is *just a split in the allocation* — **no fee vault, no negotiation** — netted and settled to L1 in USDC over go-nitro `MultiAssetHolder` / `virtualfund` / `payments`. Per-quote and per-settle RPC is metered with T2 Nitro vouchers, so the venue is **paid for quoting without ever taking custody**. This is why the money side is **clearing** — running netted accounting — rather than a service-provider directory.
- **Reuse vs build:** reuse go-nitro clearing (`payments/vouchers.go`, `protocols/virtualfund/virtualfund.go`, `MultiAssetHolder.sol`); build only the allocation-split wiring on channel state.
- **Interface consumed:** T0.3 channel state plus voucher format; T2.1 vouchers for paid quote/settle RPC.
- **Key tasks:** wire the user / protocol / integrator split into the co-signed channel state so it nets on close.
- **Testing:** the three shares net to the intended amounts across many fills and survive **both** a cooperative close and a force-close; an unpaid quote/settle request is refused without exposing the requester.

## T4.4 ETH yield

**Status:** config · v1.

The trivial yield shape. **wstETH** (and weETH) are **non-rebasing** LSTs: the balance is fixed and value accrues via a rising exchange rate against ETH. Two consequences follow.

- **Non-rebasing ⇒ a clean shielded note** — the note's value tracks by price, not by a changing balance, so the rebasing gotcha that afflicts aUSDC/aWETH disappears.
- **Yield with no Aave, no LP, no boundary tx** — a shielded wstETH note **earns staking yield just by being held**. The T4.2/T4.5 LP/aggregation/saturation machinery is **not needed** for base ETH yield; the venue is needed only to **swap that yield to USDC** privately (the T4.1 quote/settle path).

- **Reuse vs build:** config — the T0.0 Railgun pool already holds multi-asset notes (wstETH), so this is note-config integration with no new contract.
- **Interface consumed:** the T0.0 asset allow-list must include the pair (e.g. wstETH/USDC); the swap-to-USDC route runs through T4.1, filled from the venue's **static inventory** in v1 (the T4.2 solver is v2).
- **Optional higher yield:** Aave **E-Mode looping** (supply wstETH, borrow WETH, re-supply) via T5.1 → T5.0 — opt-in, with liquidation and LST risk accepted as standard DeFi risk.

## T4.5 USDC yield

**Status:** net-new · **v2** (ADR-0011). *(USDC yield defers with the market-making rail it rides; v1 yield is ETH/wstETH only.)*

USDC has no intrinsic yield, so USDC yield requires an external protocol (Aave) — and doing that privately requires the venue to act as an **LP / liquidity buffer**, not a per-user public Aave deposit. This is the **LP-buffered, batched Aave rail**:

- **Taker (private):** the user swaps a shielded **USDC note ↔ the venue's shielded aUSDC note**, atomically, at the T4.0 posted price. The user never touches Aave, so their amount and timing never appear on-chain. Capacity is bounded by LP depth.
- **LP / maker (compensated):** the venue holds inventory (shielded notes backed by a public Aave position) and **rebalances to and from Aave in aggregate, on its own decorrelated schedule** — one batched public tx, tied to no user (T4.2). The LP earns the spread; its aggregate position is public by the nature of market-making, decorrelated from individual flow (aggregation, not buckets).
- **Non-custodial both sides** — swaps are atomic; the venue trades *its own* inventory and never holds a taker's funds.

**Saturation → never a forced privacy loss.** If a deposit exceeds LP inventory, the taker is **not** pushed to a public Aave tx. Instead they keep their **shielded USDC** and either:

- **Dribble in** — swap incrementally as LP capacity replenishes; nothing leaks, and they simply get the yield note later; or
- **Become an LP** — provide the liquidity themselves, earning yield plus spread. Their large entry was already public, so being a maker adds no new exposure — and it **deepens the pool**, which fixes the LP cold-start (saturation is the recruitment signal — the compounding flywheel).

- **Reuse vs build:** build — the LP inventory plus batched rebalance is net-new; the atomic USDC↔aUSDC conversion is a Railgun RelayAdapt recipe (T4.2), and the Aave leg routes via T5.1 → T5.0.
- **Interface consumed:** T4.0 posted price, T4.2 solver rebalancing, T5.1 → T5.0 for the Aave leg; the T0.0 allow-list must include USDC/aUSDC.
- **Risk — two timing clocks:** the **aggregation window** (the privacy knob — batching needs an epoch to gather enough swaps and decorrelate the LP's public rebalance; a bigger window means a larger crowd but more latency) and **LP build-up** (the cold-start — privacy needs LP depth, which takes time and capital; thin early LP means low capacity and weak privacy, improving as it grows and as saturating whales convert to LPs).

## T4.6 ex_net matcher + LP vault (v2)

**Status:** net-new · v2.

The provably-fair upgrade is **this T4 venue + T3 ordering**. An **ex_net venue is a special case** of the v1 venue: a DSS-federated venue that can additionally **prove fair ordering** through colocated servers, per-block ordering proofs, and liquidity-proof snapshots — the ex_net primitives.

**nitro-on-railgun already satisfies the ex_net primitives.** Assets never leave custody — they sit in Nitro channels funded from shielded notes (T0.3), and the federation only prices, matches, and authorizes settlement, never holding funds. So ex_net is a **capability upgrade layered on the same spine, not a different system**.

The primitives map onto parts the spine already ships:

- **Assets never leave the trader's custody** — they sit in T0.3-funded Nitro channels. ✅ already true.
- **Federation only matches / prices / authorizes settlement** — the venue posts prices and co-signs fills; it never holds funds. ✅ already true.
- **Per-block auditable snapshots + liquidity proofs** — DSS-signed inclusion receipts and liquidity-proof snapshots. ⛔ needs **T3** (commit-reveal sequencer + beacon, epoch set-agreement/DA, DSS attestation via `chain-signatures` `ethdss`/`ethschnorr`).
- **Bonding + slashing on L1** — an on-L1 registry + bond (T0.4) so a censored-commit or bad-ordering cert is a fraud proof (T0.5). ⛔ needs the v2 L1 slashing path.

So v1 is an ex_net venue **minus the fair-ordering proofs**; adding them is the T4.6 + T3 upgrade, not a rewrite. Because privacy forces **request → quote → fill** bilaterally over hidden participants and hidden flow, private note swaps are an **RFQ system by nature** — a public CLOB over hidden participants is a contradiction — so v1 RFQ is the correct shape for the private branch, not a compromise. The matcher (frequent-batch-auction, curve orders, filler-filtered CLOB/RFQ hybrid) plus the pooled **LP vault** is **v2 price discovery** on the *public* branch; it is months of work per venue and per-venue opt-in, never a v1 prerequisite.

## Sources

- Yield & clearing (posted price v1, clearing-not-marketplace, RelayAdapt recipe, amount privacy, ETH vs USDC yield, LP-buffered saturation flywheel, built-vs-net-new) — `yield-clearing.html`
- Execution platform (ex_net, matcher, capital efficiency, RFQ-by-nature, settlement into Nitro) — `execution-platform.html`
- Architecture (T2 DSS attestation, T3 commit-reveal ordering, T4 ex_net + v1-vs-v2 posted-price note) — `architecture.html`
- Build plan (v1/v2 split) — `build-plan.html`
- Registry (T4 ids/status/release) — [`05-building-block-view.md`](./05-building-block-view.md); RFQ-vs-CLOB decision — [ADR-0007](./09-architecture-decisions.md)
- go-nitro (cerc-io @435eb2b) — swap protocol, vouchers, virtualfund, multi-asset holder, adjudicator — https://github.com/cerc-io/go-nitro (`protocols/swap/swap.go`, `payments/vouchers.go`, `protocols/virtualfund/virtualfund.go`, `MultiAssetHolder.sol`, `ForceMove.sol`, `NitroAdjudicator.sol`)
- Railgun RelayAdapt / Cookbook (atomic `unshield → call → reshield` hedge recipe) — https://github.com/Railgun-Community/cookbook
- Waku broadcaster client (quote gossip + origin-private submission transport) — https://github.com/Railgun-Community/waku-broadcaster-client
- chain-signatures (cerc-io @9016a7c; v2 DSS: `ethschnorr`, `ethdss` Stinson–Strobl (t,n)) — https://git.vdb.to/cerc-io/chain-signatures
