# Armada

<p class="lede">Armada is a shielded pool for private USDC, built by one engineering team working from the ship-armada repos. The pool, its circuits, the settlement rail on Nitro state channels, and the web client are built. The proof path, a mobile client, and any exchange venue remain a plan.</p>

<p style="margin-top:14px">
  <span class="tag">non-custodial</span>
  <span class="tag">Ethereum L1-anchored</span>
  <span class="tag">pool built, 50 bps shield fee</span>
</p>

<div class="note">
Armada is asset-privacy infrastructure for USDC: a shielded pool, spec-compatible with the Railgun design and independently authored; governance-added adapters for cross-chain transfer (CCTP, built) and yield (Aave-v4, built); and a web client (`armada-interface`), wired to the pool through `armada-sdk`. See the <a href="glossary.html">glossary</a> for terms across the pool, Nitro, and the proof path.
</div>

## Settlement over Nitro — built

`ArmadaNitroAdjudicator` carries the deposit/payout boundary inside itself: a shielded note funds a Nitro channel (`depositFromNote`), the channel runs off-chain, and a `SHIELD`-type allocation shields the outcome back into fresh notes on payout. There is no separate boundary contract and no `fundChannel`. Because the movement happens off-chain, there is no per-transaction event on the pool to link against.

- **Clears the exchange.** Not built. No posted-price venue or matcher exists yet; swaps and yield that clear through a venue remain a plan.
- **Moves value across chains.** CCTP is built: `PrivacyPool`/`PrivacyPoolClient`/`CCTPHookRouter` shield and unshield across chains, USDC-only. The hub-fronted cross-chain-swap design over Nitro channels (see `shielded-nitro-bridge-design.html`) is not built; `HubCustodian` exists and signs with FROST, but no client drives a channel through it yet.
- **Brings existing balances in.** Not built. The wallet move from Railgun (unshield there, shield here) is designed (ADR-0006) but has no code in `armada-interface` yet.

<p class="small">The <a href="engineering/A-nitro-on-railgun/A.0-overview.html">on-chain anchor overview</a> covers what is built; the <a href="engineering/shielded-nitro-bridge-design.html">cross-chain-swap construction</a> is still a design document.</p>

## What is built

<div class="cards">
  <div class="card">
    <h3>The on-chain anchor</h3>
    <p class="big">The pool, the circuits, the adjudicator, and governance.</p>
    <ul>
      <li><strong>Pool.</strong> `PrivacyPool`, routed by delegatecall to four modules, no proxy, a 50 bps deployed shield fee.</li>
      <li><strong>Circuits.</strong> `Transact(N,M)`, 19 shapes. The production Phase-2 ceremony has not run.</li>
      <li><strong>Adjudicator.</strong> A patched `go-nitro`, with the deposit/payout boundary merged in.</li>
      <li><strong>Governance.</strong> `ArmadaGovernor`, `ArmadaTreasuryGov`, `ArmadaToken`, `ArmadaCrowdfund`, `ArmadaWindDown`, `ArmadaRedemption`, and the revenue contracts.</li>
    </ul>
  </div>
  <div class="card">
    <h3>Yield</h3>
    <p class="big">Two mechanisms, both built.</p>
    <ul>
      <li><strong>ETH.</strong> A shielded wstETH note carries intrinsic staking yield; no new contract is needed.</li>
      <li><strong>USDC.</strong> `ArmadaYieldAdapter`/`ArmadaYieldVault`, a direct recipe through the pool's own `adaptContract` binding, not Railgun's RelayAdapt/Cookbook, which this codebase does not use.</li>
    </ul>
  </div>
</div>

## What is still a plan

- **The proof path.** The nimbus verified proxy is the intended trust boundary (see `engineering/09-architecture-decisions.html#adr-0009`); it runs nowhere in Armada's stack today. A client reads chain state over RPC or an unverified cache.
- **A mobile client.** `armada-client` has no application code. `armada-interface`, the live web app, is the only shipped client.
- **A settlement-client channel lifecycle and a watchtower.** The adjudicator side is built and tested; no client drives it.
- **An exchange venue.** No posted-price contract, quote/settle app, or matcher exists.

## How it fits together, as built

<pre class="diagram">
   CLIENT                    armada-interface (web): shield · transfer · unshield           built
   ─────────────────────────────────────────────────────────────────────────────────────
   ADAPTERS                  [ CCTP ]      [ Aave-v4 yield ]      [ Swaps: not built ]
   ──────────────────────────────\ ───────────┼─────────── / ────────────────────────────
   ON-CHAIN ANCHOR            PrivacyPool · ArmadaNitroAdjudicator (deposit/payout merged)
                              governance and revenue contracts                              built
   ─────────────────────────────────────────────────────────────────────────────────────
   PROOF PATH                 nimbus verified proxy — unintegrated · watcher-ts optional cache — not built
</pre>

<p class="small">See <a href="architecture.html">Architecture</a> for the full tier stack, and <a href="engineering/05-building-block-view.html">the item registry</a> for per-item status with code citations.</p>

## Engineering documentation

The full architecture is documented under <a href="engineering/README.html">Engineering</a>, organized as <a href="https://arc42.org">arc42</a>. The <a href="engineering/05-building-block-view.html">building-block view</a> holds the tier stack and the item registry, with a detail page for each tier (<a href="engineering/T0-ethereum-anchor.html">T0</a> through <a href="engineering/T6-client-apps.html">T6</a>). The <a href="engineering/09-architecture-decisions.html">architecture decisions</a> record each choice, and the <a href="engineering/build-plan.html">build plan</a> tracks status and the remaining work.
