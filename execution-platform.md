# Execution platform (ex_net)

<p class="lede">A <strong>v2</strong> feature for price <em>discovery</em> and market-making: a frequent-batch-auction, curve-order matcher that unifies CLOB and RFQ, capital-efficient and non-custodial. It is not in Armada's v1 scope. v1 executes at a <a href="yield-clearing.html">posted price</a>, and this matcher is the later upgrade that replaces the posted price with real price discovery.</p>

<p style="margin-top:14px">
  <span class="tag">scope: v2 (post-v1 upgrade)</span>
  <span class="tag">validated in depth</span>
  <span class="tag">non-custodial</span>
  <span class="tag">powers swaps + yield</span>
</p>

<div class="note">
This is the price-discovery upgrade, kept out of Armada's v1 (v1 = posted price + clearing; see <a href="yield-clearing.html">Yield &amp; clearing</a>). It sits here because the code is the best-grounded part of the scope: the 2019–2020 <code>matcher</code> PoC was read in depth this cycle. See <a href="architecture.html#t4">Architecture · T4</a> and <a href="build-plan.html#versions">Build plan · versions</a>.
</div>

## Contents

<div class="toc">
  <a href="#why">Why a platform, not an adapter</a>
  <a href="#lineage">Lineage — ex_net (2017) → matcher (2019–20)</a>
  <a href="#matcher">The matcher — what it actually is</a>
  <a href="#fba">Frequent batch auction (fair ordering)</a>
  <a href="#clobrfq">CLOB + RFQ in one order type</a>
  <a href="#capital">Capital efficiency vs. AMMs</a>
  <a href="#settle">Settlement into Nitro</a>
  <a href="#apps">Swaps &amp; yield as applications</a>
  <a href="#refs">References</a>
</div>

## Why a platform, not an adapter {#why}

A shielded A→B swap needs a venue that prices, matches, and settles privately and non-custodially. Moving idle balances into yield needs the same primitives plus metering and accounting. These are not two separate adapters; they are two applications of one execution and settlement substrate. ex_net generalizes what adapters need: move assets, price and swap, provide liquidity and yield, settle privately. It sits beneath Armada's Adapters tier, and Swaps and yield become thin integrations on top rather than peers.

<div class="note">
<strong>Trade-off.</strong> A shared substrate under multiple adapters is a bigger surface than one independently-deployed adapter, and it needs its own audit, a deviation from Armada's per-adapter audit discipline. The pool stays immutable and untouched; ex_net operates at the adapter and execution tier only.
</div>

## Lineage — ex_net (2017) → matcher (2019–20) {#lineage}

Vulcanize's 2017 **ex_net** whitepaper<sup><a href="#r-exnet">[1]</a></sup> specified a bonded, federated venue for the privacy-preserving, cross-chain exchange of crypto-assets, where assets never leave the trader's custody (they sit in parent-chain payment channels) and the federation only matches, prices, and authorizes settlement. It gave the vocabulary still in use: traders, liquidity providers, validators; bonding; per-block auditable snapshots; liquidity proofs; identity proofs; and the observation that a "decentralized exchange" as a noun cannot prevent operator front-running.

The `matcher`<sup><a href="#r-matcher">[24]</a></sup> is the concrete engine: a private Vulcanize repo (commits 2019–2020) with three branches, `master` (spec and skeletons), `ian_branch` (a fuller `matcher.py`), and the most-developed `xsleonard_branch` (`algorithm.md`, `matcher2.py`, a runnable `matcher-sim.py`). Together they are a working proof-of-concept of the ex_net matching layer.

## The matcher — what it actually is {#matcher}

A per-block batch-auction matcher over order *curves*, across multiple trading pairs, on a payment-channel hub, with per-order counterparty filtering. From the design docstring: "reducing free options and cryptographically binding both parties based on their stated intentions." In detail:

- **Order = a curve, not a point.** An order specifies price-varying liquidity over a range (`Curve` / `cRange`), flattened into fillable `Chunk` units, a depth view over one order.
- **Filler-filtered.** Each order carries an allowlist (`Order_Curve.fillers`): empty means open to anyone, non-empty means only listed pubkeys may fill. This is the RFQ and private-swap boundary inside one order type.
- **Multi-pair rings.** A DFS path/ring clearing (`find_path` / `execute_path`) fills an order across multiple pairs (`LOL-ABC → LOL-XYZ → XYZ-WTF`), giving price improvement transparently.
- **Receipts / vouchers / plasma.** Validators issue three receipt types (asset-for-voucher, order-curve-accepted, voucher-transfer), the ex_net settlement lineage.

<p class="src">source: <code>vulcanize/matcher</code> (private) — <code>xsleonard_branch</code>: <code>algorithm.md</code>, <code>matcher2.py</code>, <code>matcher-sim.py</code><sup><a href="#r-matcher">[24]</a></sup></p>

## Frequent batch auction (fair ordering) {#fba}

The engine clears as a **frequent batch auction**: discrete per-block clearing at a uniform price per pair, allocation by pro-rata plus a deterministic lottery (`_distribute_chunks` → shuffle), and explicitly not first-in-queue price-time priority. There is no `priority` / `cancel` / `market` speed queue in the engine. This is the Budish–Cramton–Shim design<sup><a href="#r-fba">[23]</a></sup> for defeating latency and HFT front-running, and combined with commit-reveal ([Architecture · T3](architecture.html#t3)) it also blocks operator front-running. Fills feel like a CLOB but settle like a batch auction.

## CLOB + RFQ in one order type {#clobrfq}

The atomic order is already a limit order (the `Curve` is the aggregated book; `OrderSet` builds a bid/ask depth view), so a familiar limit-order and depth-book UX drops out directly. The one difference that must be disclosed is that matching is a frequent batch auction (uniform price, pro-rata and lottery) rather than continuous price-time priority.

- **Open orders** (empty `fillers`) are the *lit* book, a public CLOB view.
- **Filler-filtered / blinded orders** stay off-book, the RFQ and private-swap side.

One construction spans a public CLOB and bilateral RFQ, with the privacy caveat that a lit book only ever shows opt-in open orders. Because privacy forces request → quote → fill bilaterally over hidden participants and hidden flow, private note swaps are an RFQ system by nature: a public CLOB over hidden participants is a contradiction. The lit book is the public branch, and private note swaps are the P2P branch taken to its conclusion.

## Capital efficiency vs. AMMs {#capital}

The venue is quote-based market-making (maker/keeper prices off-chain, any strategy), settled in channels, and structurally more capital-efficient than an AMM for two reasons:

- **No dead curve.** An AMM spreads LP capital across the whole price range (0→∞); the vast majority is never near the market. A quoting maker deploys capital at the touch, sized to actual demand.
- **No forced LVR or impermanent loss.** An AMM is a mechanical stale quote that arbitrageurs pick off every block; a quote-based maker reprices to live (bounded by the reveal/quote epoch), a direct capital-return advantage. Balance-sheet reuse across pairs from one hub channel raises utilization per dollar locked.

This is the empirically-known result: 0x RFQ, Hashflow, and professional FX/OTC exist because quote-based market-making delivers better prices with less locked capital than AMMs. Passive LPs are still first-class: they can fund hub capacity, post resting liquidity, or deposit into a pooled-maker vault where a keeper runs the market-making. The real difference is who runs the pricing (an active maker off-chain), not whether passive capital can participate.

## Settlement into Nitro {#settle}

Fills settle non-custodially into **Nitro state channels**: go-nitro's multi-asset swap protocol<sup><a href="#r-swap">[6]</a></sup> and `SwapChannel`<sup><a href="#r-swapchan">[7]</a></sup> over virtual channels<sup><a href="#r-vfund">[9]</a></sup>, with the `ForceMove` dispute game<sup><a href="#r-forcemove">[5]</a></sup> as the escape hatch. Every fill terminates in a co-signed settlement receipt or a slashable proof of its absence; no receipt means force-close to the pre-fill state. Net-new wiring: the LP/market-maker vault, receipt-or-slash, and multi-asset outcomes (see [Build plan · T4](build-plan.html#t4)).

## Swaps & yield as applications {#apps}

<table>
<tr><th>Armada adapter</th><th>How it uses the platform</th></tr>
<tr><td><strong>Swaps</strong></td><td>a thin integration exposing "swap shielded A→B"; the matcher does the pricing + matching + channel settlement underneath.</td></tr>
<tr><td><strong>Yield (Aave v4)</strong></td><td>ex_net is the <em>private rail</em> moving idle shielded balances into Aave and back, with metering/accounting/settlement.</td></tr>
<tr><td><strong>Yield (native)</strong></td><td>ex_net <em>originates</em> yield — LP/market-making spread + swap/routing fees — an additional source alongside Aave.</td></tr>
<tr><td><strong>Future adapters</strong></td><td>governance-added, on the same substrate; nothing touches the immutable pool.</td></tr>
</table>

<hr/>

## References {#refs}

<ol class="refs">
<li id="r-exnet"><strong>ex_net whitepaper</strong>, Vulcanize (© 2017). <a href="./ex_net_whitepaper.pdf">ex_net_whitepaper.pdf</a> (in this repo).</li>
<li id="r-forcemove"><strong>ForceMove.sol</strong> — dispute game. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/ForceMove.sol#L39-L127">go-nitro · ForceMove.sol</a>.</li>
<li id="r-swap"><strong>Nitro swap protocol</strong> (Deepstack). <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/protocols/swap/swap.go#L44-L146">go-nitro · protocols/swap/swap.go</a>; demo <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/demo/swaps.md">demo/swaps.md</a>.</li>
<li id="r-swapchan"><strong>SwapChannel</strong>. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/channel/swap.go#L10-L14">go-nitro · channel/swap.go#L10–L14</a>.</li>
<li id="r-vfund"><strong>Virtual channels</strong>. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/protocols/virtualfund/virtualfund.go">go-nitro · protocols/virtualfund/virtualfund.go</a>.</li>
<li id="r-railgun"><strong>Railgun</strong> — shielded-pool contracts &amp; circuits. <a href="https://github.com/Railgun-Privacy/contract">contract</a>, <a href="https://github.com/Railgun-Privacy/circuits-v2">circuits-v2</a>.</li>
<li id="r-fba"><strong>Frequent batch auctions</strong> — E. Budish, P. Cramton &amp; J. Shim, "The High-Frequency Trading Arms Race: Frequent Batch Auctions as a Market Design Response," <em>Quarterly Journal of Economics</em> 130(4), 2015, 1547–1621. <a href="https://doi.org/10.1093/qje/qjv027">doi:10.1093/qje/qjv027</a>.</li>
<li id="r-matcher"><strong>matcher</strong> — curve-based orderbook batch-auction matching-engine PoC (Vulcanize, private; commits 2019–2020; branches <code>master</code> / <code>ian_branch</code> / <code>xsleonard_branch</code>). <span class="small">Footnote: its range-order ("curve") liquidity model predates Uniswap v3 concentrated liquidity (2021).</span></li>
</ol>

<p class="small" style="margin-top:26px">Code links pinned to commits: go-nitro <code>435eb2b</code>. <code>matcher</code> is a private Vulcanize repo (access-gated). Internal Google Docs require Laconic/Vulcanize access.</p>
