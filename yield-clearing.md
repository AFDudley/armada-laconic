# Yield & clearing

<p class="lede">The money side of the v1 venue — how value is priced, exchanged, cleared, and earns yield — kept non-custodial and private end-to-end, using Railgun for value privacy and Nitro for settlement. No service-provider marketplace; no custodial vault.</p>

<p class="legend" style="margin-top:14px">
  Maturity:
  <span class="chip ok">built</span> reusable prod code ·
  <span class="chip part">partial</span> exists, needs integration ·
  <span class="chip new">net-new</span> must build
</p>

<div class="note">
<strong>Scope.</strong> Everything here is v1, kept simple. Price <em>discovery</em> (an order book / matcher) is a separate v2 feature (see <a href="execution-platform.html">Execution platform</a>); v1 uses a posted price. Design B (hiding trade amounts from the venue) is v3. See the <a href="build-plan.html">Build plan</a> for the full version split.
</div>

## Contents {#contents}

<div class="toc">
  <a href="#pricing">1. Pricing — posted price (v1)</a>
  <a href="#clearing">2. Clearing over Nitro</a>
  <a href="#recipe">3. The "adapter" is a Railgun recipe</a>
  <a href="#amounts">4. Amount privacy</a>
  <a href="#yield">5. Yield — ETH vs USDC</a>
  <a href="#maturity">6. What's built vs net-new</a>
  <a href="#refs">References</a>
</div>

## 1. Pricing — posted price (v1) {#pricing}

v1 has no price discovery. Armada posts a price on both sides of a pair (bid and ask) and participants take it or leave it. If the price is wrong, with no fills or lopsided fills, Armada posts a new one. Simple and manual.

- **Who sets it:** a single privileged `priceSetter` — one ETH L1 wallet to start, migratable to a governance address later (the same admin-key→governance path Armada already uses for its steward).
- **What the venue does:** hold the current posted price, execute fills at it, and clear/settle over Nitro. No order book, no auction, no queue.
- **Why it's enough:** a posted, take-it-or-leave-it price has nothing to front-run, which removes the entire fair-ordering problem (commit-reveal, sequencing, the epoch set-agreement) from v1. Those live in [v2](execution-platform.html) with the matcher.

<div class="note"><strong>v2 upgrade path:</strong> when you want competitive quotes / multiple LPs setting prices instead of one admin quote, the ex_net matcher (frequent-batch-auction, curve orders, CLOB/RFQ) replaces the posted price with real price discovery. Not needed for Armada now.</div>

## 2. Clearing over Nitro {#clearing}

The money side is clearing, not a marketplace: the running, multi-party accounting of who owes what, netted and settled non-custodially. It is entirely Nitro:

- **Meter** — per-use vouchers<sup><a href="#r-vouchers">[2]</a></sup> (signed IOUs).
- **Clear** — each metered payment updates a co-signed channel allocation that splits value among user / protocol / integrator (the integrator revenue-share is just a split in the allocation — no fee vault, no negotiation).
- **Settle** — a channel checkpoint/close nets those allocations to L1 in USDC, non-custodially (force-closable).

Both trades and fees clear the same way. Integrator revenue-sharing, one of Armada's stated goals, is delivered as a clearing allocation, using shipped go-nitro code (`MultiAssetHolder` / `virtualfund` / `payments`).

## 3. The "adapter" is a Railgun recipe {#recipe}

"Adapter" is overloaded (see [glossary](glossary.html)). The value-moving adapter here is Railgun's RelayAdapt / Cookbook "recipe"<sup><a href="#r-cookbook">[5]</a></sup> — an atomic `unshield → external call → reshield` in one transaction. It is:

- **Not DSS.** The watcher-party threshold signature (`chain-signatures`) only attests (receipts/certs); it never holds or moves funds.
- **Not a custodial vault.** The recipe reverts entirely if any step fails; the user holds a shielded note before and after. Custody never leaves the user.
- **Origin-private.** A broadcaster submits it, so the on-chain origin is not the user's `0zk` address.

This is what turns a shielded USDC note into a shielded aUSDC ("aave") note, or wraps a CCTP mint into the shield, all atomically, in the user's custody.

## 4. Amount privacy {#amounts}

Arbitrary-amount privacy is already solved inside the shield: Railgun's note commitments hide the amount and the zk-SNARK enforces value-conservation + range, so $50 and $5,000,000 are equally opaque for any shielded-to-shielded activity. No denomination buckets (that was Tornado's crutch for lacking amount-hiding commitments).

The amount is only exposed at the transparent boundary (`shield`/`unshield`, the CCTP mint, an Aave deposit) because the counterparty must receive it in the clear. Mitigations there are aggregation, not buckets:

- **Stay shielded** — keep value in the pool so crossings are rare.
- **Atomic batched aggregation** — a keeper batches many users' boundary ops into one aggregate tx (CoinJoin-style), non-custodially (atomic), so individual amounts vanish into the aggregate.
- **Decorrelate** entry from exit (amount + timing).

<div class="note"><strong>The irreducible exception:</strong> a large lone public crossing (a $5M shield or a $5M Aave supply) stands out — no scheme hides one $5M public event in a crowd of $50s. But that is a one-time entry data point: once inside, exits, yield, and transfers are private, provided exits are decorrelated and use common/aggregated sizes. Entry can be clear; everything after is private.</div>

## 5. Yield — ETH vs USDC {#yield}

The two assets have very different yield shapes, because ETH carries yield intrinsically and USDC does not.

### ETH — a shielded wstETH note

wstETH (and weETH) are non-rebasing LSTs: fixed balance, value accrues via a rising exchange rate vs ETH. Two consequences:

- **Non-rebasing ⇒ a clean shielded note** — the note's value tracks by price, not by a changing balance (the rebasing gotcha that afflicts aUSDC/aWETH disappears).
- **Yield with no Aave, no LP, no boundary tx** — a shielded wstETH note earns staking yield just by being held. The entire LP/aggregation/saturation machinery below is not needed for base ETH yield.

<p class="small">Optional higher yield: Aave E-Mode looping (supply wstETH, borrow WETH, re-supply) for leveraged staking yield — opt-in, with liquidation + LST risk (accepted as standard DeFi risk).</p>

### USDC — needs an LP: buffered, batched Aave

USDC has no intrinsic yield, so USDC yield requires an external protocol (Aave), and doing that privately needs an LP as a liquidity buffer, not a per-user public Aave deposit:

- **Taker (private):** the user swaps a shielded USDC note ↔ the LP's shielded aUSDC note, atomically, at the posted price. The user never touches Aave, so their amount/timing never appear on-chain. Capacity-bounded by LP depth.
- **LP / maker (compensated):** the LP holds inventory (shielded notes backed by a public Aave position) and rebalances to/from Aave in aggregate, on its own decorrelated schedule — one batched public tx, tied to no user. The LP earns the spread; its aggregate position is public by the nature of market-making.
- **Non-custodial both sides** — swaps are atomic; the LP trades its own inventory and never holds a taker's funds.

#### Saturation → never a forced privacy loss

If a deposit exceeds LP inventory, the taker does not get pushed to a public Aave tx. Instead they keep their shielded USDC and either:

- **Dribble in** — swap incrementally as LP capacity replenishes (just get the yield note later; nothing leaks), or
- **Become an LP** — provide the liquidity themselves, earning yield + spread. Their large entry was already public, so being a maker adds no new exposure, and it deepens the pool, which fixes the LP cold-start (saturation is the recruitment signal). This is the compounding flywheel Armada already describes.

#### Two timing clocks

- **Aggregation window** (privacy knob) — batching needs an epoch to gather enough swaps and decorrelate the LP's public rebalance; bigger window = larger crowd = more latency.
- **LP build-up** (cold-start) — privacy needs LP depth; depth takes time + capital and must be accumulated without the LP's own positions being trivially linkable. Thin early LP = low capacity + weak privacy; improves as it grows (and as saturating whales convert to LPs).

## 6. What's built vs net-new {#maturity}

<table>
<tr><th>Piece</th><th>Code</th><th>Maturity</th></tr>
<tr><td>Clearing (meter + settle)</td><td>go-nitro vouchers / virtualfund / MultiAssetHolder<sup><a href="#r-vouchers">[2]</a></sup></td><td class="ok">built</td></tr>
<tr><td>Fee-split allocation (user/protocol/integrator)</td><td>on go-nitro channel state</td><td class="new">net-new</td></tr>
<tr><td>Posted-price venue (priceSetter → governance)</td><td>small contract + poster</td><td class="new">net-new (small)</td></tr>
<tr><td>Adapter = RelayAdapt recipe</td><td>Railgun Cookbook<sup><a href="#r-cookbook">[5]</a></sup></td><td class="part">Railgun built; our recipes net-new</td></tr>
<tr><td>ETH yield = shielded wstETH note</td><td>Railgun multi-asset (holds wstETH<sup><a href="#r-wsteth">[6]</a></sup>)</td><td class="part">pool built; note config integration</td></tr>
<tr><td>USDC yield = LP-buffered Aave</td><td>Railgun recipe + Aave<sup><a href="#r-aave">[7]</a></sup> + LP inventory</td><td class="new">net-new (LP + batched rebalance)</td></tr>
<tr><td>Optional levered yield (looping)</td><td>Aave E-Mode<sup><a href="#r-aave">[7]</a></sup></td><td class="new">net-new + external (Aave v4 not live)</td></tr>
</table>
<p class="small">No new cryptography for any of this — it is Railgun (value privacy) + Nitro (clearing/settlement) + a small posted-price contract + LP operations. Price discovery, fair-ordering, and ZK matching are explicitly out of v1.</p>

<hr/>

## References {#refs}

<ol class="refs">
<li id="r-exnet"><strong>ex_net whitepaper</strong>, Vulcanize (© 2017). <a href="ex_net_whitepaper.pdf">ex_net_whitepaper.pdf</a>.</li>
<li id="r-vouchers"><strong>go-nitro</strong> — clearing/settlement: <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/payments/vouchers.go">payments/vouchers.go</a>, <code>protocols/virtualfund</code>, <code>MultiAssetHolder.sol</code>, @ <code>435eb2b</code>.</li>
<li id="r-forcemove"><strong>ForceMove.sol</strong> — non-custodial force-close. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/ForceMove.sol">go-nitro · ForceMove.sol</a>.</li>
<li id="r-railgun"><strong>Railgun</strong> — shielded pool (BN254; arbitrary amounts hidden in-circuit). <a href="https://github.com/Railgun-Privacy/contract">contract</a>, <a href="https://github.com/Railgun-Privacy/circuits-v2">circuits-v2</a>.</li>
<li id="r-cookbook"><strong>Railgun Cookbook / RelayAdapt</strong> — atomic <code>unshield → cross-contract call → reshield</code> recipes (Aave deposit, swaps, combos). <a href="https://github.com/Railgun-Community/cookbook">github.com/Railgun-Community/cookbook</a>.</li>
<li id="r-wsteth"><strong>wstETH / weETH</strong> — non-rebasing LSTs (value via rising exchange rate). <a href="https://docs.lido.fi/contracts/wsteth/">Lido wstETH</a>.</li>
<li id="r-aave"><strong>Aave v3</strong> — WETH lending (aWETH) + wstETH/weETH supply &amp; E-Mode looping for leveraged staking yield. <a href="https://aave.com/blog/lido-aave-case-study">Aave · Lido case study</a>. (v4 not yet live.)</li>
</ol>
<p class="small" style="margin-top:26px">Companion to <a href="architecture.html">Architecture</a>, the <a href="build-plan.html">Build plan</a> (v1/v2/v3), and <a href="mobile-privacy.html">Mobile privacy</a>.</p>
