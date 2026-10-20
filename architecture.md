# Architecture — the tier stack

<p class="lede">Armada settles private execution on a Nitro rail. The motion is the same everywhere: notes in → normal Nitro → notes out. Value leaves Armada's shielded pool, moves through ordinary Nitro state channels, and returns as shielded notes. Seven tiers, each mapped to existing code where it exists, and marked as a plan where it does not.</p>

<p style="margin-top:14px">
  <span class="tag">non-custodial</span>
  <span class="tag">Ethereum L1-anchored</span>
  <span class="tag">Nitro settlement</span>
  <span class="tag">pool immutable / untouched</span>
</p>

<div class="note">
Tiers are named in the software-architecture sense (an n-tier stack), not the "Layer&nbsp;1 / Layer&nbsp;2" of blockchains. Every claim maps to code on the <a href="engineering/05-building-block-view.html">item registry</a> and the <a href="engineering/build-plan.html">Build plan</a>, with pinned commits and status. This page is the structure; the registry is the status and code location.
</div>

## Contents

<div class="toc">
  <a href="#shape">The shape of the network (today vs. the plan)</a>
  <a href="#settlement">The settlement rail — notes in, normal Nitro, notes out</a>
  <a href="#t0">T0 · Ethereum anchor — settlement &amp; shielded custody</a>
  <a href="#t1">T1 · Ingestion (the proof path)</a>
  <a href="#t2">T2 · Watcher substrate</a>
  <a href="#t3">T3 · Ordering &amp; fault model (plan)</a>
  <a href="#t4">T4 · Execution / venue (plan)</a>
  <a href="#t5">T5 · Adapters</a>
  <a href="#t6">T6 · Client / apps</a>
  <a href="#identity">Identity — optional match filter</a>
  <a href="#privacy">Cross-cutting · Privacy boundary</a>
  <a href="#refs">References</a>
</div>

## The shape of the network — today vs. the plan {#shape}

The deployed shape is a conventional server-client topology, not a peer-to-peer federation. `armada-interface`, a React web app, talks plain RPC/HTTP to Ethereum L1 and to `armada-relayer`, a conventional Express/Postgres server that indexes pool events with Ponder and charges a flat relay fee. No watcher-party federation, no bonded service-provider network, and no threshold-signed attestation substrate exist anywhere in the client, relayer, interface, or SDK repos.

<pre class="diagram">
                browser (today)
      armada-interface: keys in memory, note-scan,
      browser Groth16 prover, shield/transact calls
                        |
                        v   plain RPC / HTTP
      ┌─────────────────────────────────────────┐
      │  armada-relayer (server, today)           │
      │  Ponder/Postgres indexer · flat-fee actor │
      └─────────────────────────────────────────┘
                        |
                        v
      ┌─────────────────────────────────────────┐
      │   ETHEREUM L1: shielded pool + adjudicator │  T0
      │   ArmadaNitroAdjudicator (merged boundary) │
      └─────────────────────────────────────────┘

      ─────────── plan, no code today ───────────
      nimbus verified proxy integration (T1, proven on the fixturenet) — a client
      would verify reads instead of trusting the relayer
      watcher federation / threshold attestation (T2.4) —
      design only
      execution venue (T4) — posted-price contract + app
      mobile client (T6.5) — armada-client's docs/SPEC.md plan
</pre>

<p class="small">
Clients hold their own keys and funds today (`armada-interface`'s browser key manager); channels, once a client drives them, would be L1-collateralized and unilaterally exitable through the built adjudicator. The earlier plan's "only fixed infra is STUN/TURN plus a few emitters" claim does not describe today's deployment: `armada-relayer` is itself a conventional fixed server, not a thin bootstrap node.
</p>

## The settlement rail — notes in, normal Nitro, notes out {#settlement}

Every flow resolves to one motion. Value leaves Armada's shielded pool as a note, funds a Nitro channel, and returns to a shielded note at the end. `ArmadaNitroAdjudicator` does the settling and holds the deposit/payout boundary in the same contract as the ForceMove dispute logic — there is no separate boundary contract. This is built: `depositFromNote` unshields a note into a channel, and a `SHIELD`-type allocation in `_executeSingleAssetExit` re-shields a channel's outcome.

No cross-chain swap is built. A multi-chain construction, with a shielded pool deployed on each chain and legs settling through mirrored channels, remains a draft (<a href="engineering/shielded-nitro-bridge-design.html">cross-chain swap design</a>); it has no deployed custodian and no cross-chain adjudicator. Armada's one deployed hub custodian signs with FROST (see T2 below).

## T0 · Ethereum anchor — settlement & shielded custody {#t0}

Settlement is non-custodial Nitro state channels on Ethereum L1. Funds are locked in `ArmadaNitroAdjudicator` and allocated off-chain by co-signed state; a channel closes cooperatively at internet speed, or unilaterally via the `ForceMove` dispute game (`challenge`, `checkpoint`, `conclude`)<sup><a href="#r-forcemove">[5]</a></sup>. If a party misbehaves, the user force-closes to their last co-signed state. The adjudicator only ever honors user-signed states, so a stalled counterparty can delay a user but cannot fabricate balances or move unsigned funds.

<pre>contract HashLockedSwap is IForceMoveApp {
    struct AppData { bytes32 h; bytes preImage; }   // reveal preimage of h to unlock
    ...
}</pre>

<p class="src">source: <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/examples/HashLockedSwap.sol#L11-L33">go-nitro · HashLockedSwap.sol#L11–L33</a> (the upstream `go-nitro` this builds on, now vendored and patched into `armada-nitro`)</p>

**Built here:** the shielded pool (`PrivacyPool`, T0.0), `ArmadaNitroAdjudicator` — a patched `go-nitro` adjudicator that also carries the deposit/payout boundary (T0.2) — and the governance/revenue contracts (`ArmadaGovernor`, `ArmadaTreasuryGov`, `TreasurySteward`, `ArmadaToken`, `ArmadaCrowdfund`, `ArmadaWindDown`, `ArmadaRedemption`). The circuits exist, with the Phase-2 trusted-setup ceremony still open (T0.1). **Not built here:** a venue/party registry and bond contract (T0.4), an on-chain sequencing-cert and fraud-proof verifier (T0.5), and a native channelized commitment for amount-privacy-in-play (T0.6). Full status on the <a href="engineering/05-building-block-view.html">item registry</a>.

## T1 · Ingestion — the proof path {#t1}

The earlier plan called for a `nimbus-eth1` state-diff emitter mapping into IPLD, served by `watcher-ts`. That pipeline is geth-only (`plugeth`), dormant since August 2024, and its current IPLD schema stores no receipt trie, so no proof connects a single log to `receiptsRoot` in it (`labbook/2026-10-19-nimbus-verified-proxy-proofs.md`). `status-im/nimbus-eth1` instead ships a **verified proxy** — native, WASM, and a C/Go-bound library — that does the verification job upstream of the client: it anchors a header to a beacon light client, verifies storage with EIP-1186 proofs, and verifies a block's receipts by rebuilding the receipt trie against `header.receiptsRoot`. Rebuilding the trie proves both the inclusion of one log and the completeness of the block's logs at once, which an `eth_getLogs` range cannot do, since an omitted log inside a range is undetectable.

**What is built:** the proxy proved a pool log and a storage slot by hand on the armada-nitro fixturenet, through its own nimbus-eth1/nimbus-eth2 pair; its fixturenet patch lives on `AFDudley/nimbus-eth1` `armada/verified-proxy-fixes`, not upstream. **What is not:** no Armada repo runs or embeds it as a standing service; no server operates it against any Armada chain; `armada-interface` ships no WASM bundle (the WASM build compiles but stalls at its first `sync()` call in a browser); `libverifproxy.a` builds for iOS and Android, but no client links it in. See <a href="engineering/T1-ingestion.html">T1 ingestion</a> for the full assessment.

## T2 · Watcher substrate {#t2}

The earlier plan called for a bonded federation of watcher parties serving proof-carrying feeds over `watcher-ts`, `ts-nitro`, `@cerc-io/peer`, and a threshold-signed attestation. None of that exists in any ship-armada repo.

With the nimbus verified proxy as the trust boundary (T1), a server-side watcher becomes optional: a cache a client may read from, then re-verify through the proxy. It is never itself a trust boundary. Today's live cache is `armada-relayer`'s Ponder/Postgres indexer (`relayer-v2/watcher`), a conventional server out of the earlier plan, serving `/v2/quick-sync` over plain REST with no proof attached. `armada-relayer`'s actor charges a flat USDC fee computed from a gas estimate, not go-nitro payment vouchers.

<table>
<tr><th>Service</th><th>Earlier plan</th><th>Today</th></tr>
<tr><td>Private shielded-pool sync</td><td>proof-carrying feed, watcher-ts GraphQL</td><td>not built; a client trusts whatever RPC it is configured with</td></tr>
<tr><td>Metering</td><td>per-request Nitro vouchers</td><td>flat USDC fee per transaction (`fee-calculator.ts`)</td></tr>
<tr><td>Matching / venue</td><td>commit-reveal sequencing + a matcher</td><td>not built — no venue exists (T4)</td></tr>
<tr><td>Attestation</td><td>threshold-Schnorr (DSS) inclusion receipts</td><td>not built; the one deployed threshold scheme is FROST, inside `armada-nitro`'s `HubCustodian`, unrelated to a watcher federation</td></tr>
<tr><td>Membership / bonding</td><td>an on-L1 registry + bond</td><td>not built (T0.4)</td></tr>
</table>

### The one deployed threshold signature — FROST

`armada-nitro`'s `HubCustodian` signs with **FROST** over secp256k1 (KECCAK-256), via the pinned `frost-secp256k1-evm` library. FROST secures one contract's hub-delegate signing, not a watcher federation's attestation; no federation, bond, or attestation substrate exists to apply it to.

## T3 · Ordering & fault model (plan) {#t3}

The earlier design calls for commit-reveal ordering so a venue cannot front-run or position-manipulate hidden orders, with threshold-signed inclusion receipts making censorship slashable. No sequencer, randomness beacon, data-availability layer, or attestation code exists in `armada-poc` or `armada-nitro`. This tier has no code and depends on T4 (a venue to order for) and T0.4/T0.5 (the bond and fraud-proof verifier), neither of which is built either. See <a href="engineering/T3-ordering.html">T3 ordering</a>.

## T4 · Execution / venue (plan) {#t4}

The plan calls for a posted-price RFQ venue in v1 — a single `priceSetter` wallet publishes bid/ask, take-it-or-leave-it, cleared over Nitro — with an automated market-making matcher as a later upgrade. No price-setter contract, no quote/settle ForceMove app, and no venue of any kind exists in `armada-poc` or `armada-nitro`. See <a href="engineering/T4-execution-venue.html">T4 execution venue</a> and <a href="execution-platform.html">Execution platform</a> for the full plan.

## T5 · Adapters {#t5}

Armada's adapters are applications that move value in and out of the pool without touching the pool contract itself. **Built:** CCTP (cross-chain USDC, via `PrivacyPool`/`PrivacyPoolClient`/`CCTPHookRouter`) and Aave-v4 yield, through the pool's own `adaptContract`/`adaptParams` binding (`ArmadaYieldAdapter.lendAndShield`/`redeemAndShield`) — a bespoke Armada construction, not a Railgun RelayAdapt/Cookbook recipe; no RelayAdapt contract exists anywhere in `armada-poc`. For ETH yield, any non-blocklisted ERC-20 can already be shielded as a note, so a shielded wstETH note needs no new contract. **Not built:** a Swaps contract. See <a href="yield-clearing.html">Yield &amp; clearing</a>.

<div class="note">
<strong>Audit boundary.</strong> A future execution venue shared across multiple adapters would be a shared dependency with its own audit surface. This stays a plan consideration (T4), not a build decision, since no venue exists yet.
</div>

### The CCTP adapter — cross-chain USDC on/off-ramp

"Privacy on Ethereum" describes where the shielded pool lives, not where the money is. Circle mints native USDC on around a dozen chains, so most users' USDC is not already on mainnet. The **CCTP adapter** is the cross-chain rail that fills and drains the pool, an adoption and liquidity adapter rather than a privacy primitive. CCTP is Circle's burn-and-mint protocol: burn on the source chain, mint native USDC on the destination against a Circle attestation, non-custodial, 1:1, with no slippage and no wrapped-asset or bridge-hack risk. This is built: `PrivacyPool`/`PrivacyPoolClient`/`CCTPHookRouter` implement cross-chain shield/unshield over CCTP v2.

<div class="note">
<strong>CCTP is transport, not privacy.</strong> The burn shows sender and amount on the source chain and the mint shows recipient and amount on the destination; if the same public addresses appear on both ends the cross-chain link is trivial. Privacy comes entirely from binding the mint <em>into the shield</em>. Residual leaks: amount and timing correlation across the hop, and the mint transaction's origin, since no broadcaster/keeper submits it on the user's behalf today.
</div>

## T6 · Client / apps {#t6}

The earlier plan called for a fork of `laconic-wallet`, a React Native app, pairing with an in-browser or mobile Nitro node and a libp2p relay. `armada-client`, the repo meant to host that fork, has no application code: 45 files, all docs and config, no `src/`. The live client is `armada-interface`, a React 19 + Vite **web** app, independently written and wired to `armada-sdk`. It holds keys in browser memory (no BIP-39/HD custody, no secure enclave), scans and trial-decrypts notes locally, and runs a `snarkjs` Web Worker Groth16 prover in the browser. No native mobile prover exists. No code drives an `armada-nitro` channel lifecycle: no open/fund-from-note, propose/accept state, cooperative close, or force-close exists in `armada-interface` or `armada-sdk`. No client-side watchtower exists; only `armada-nitro`'s e2e gate harness has one, for tests. `armada-client` `docs/SPEC.md` records the plan for a unified TypeScript stack on four targets — Linux server, browser, iOS, Android — with no shipped code yet. See <a href="engineering/T6-client-apps.html">T6 client apps</a>.

## Identity — an optional match filter {#identity}

Identity stays optional and unbuilt. The design calls for a user to source a selective-disclosure attestation off-band — for example a Self zk-passport (nationality, age, OFAC-clear, personhood) — proven on-device and verified off-chain, anchored to an Ethereum address, with the shielded pool identity left unlinked. No such mechanism exists in the client today (T6.8). It would thread through no other tier if built, which is why it is called out separately rather than as a cross-cutting concept.

## Cross-cutting · Privacy boundary {#privacy}

Privacy is delivered by funding and settling channels through Armada's own shielded pool: value is a commitment inside the pool, and it becomes visible only where it crosses into `ArmadaNitroAdjudicator`, at a deposit, a payout, or a force-close. There is no separate boundary adapter; the boundary is inside the adjudicator contract itself (see <a href="#settlement">the settlement rail</a>, above).

<p class="small">
Caveats: identity privacy is only as strong as the pool's anonymity set, which grows only from deposits into Armada's own pool — there is no cross-pool proof and no onboarding bridge from another pool (ruling 5). Amounts are hidden inside the pool; the only leak is at the boundary amount, and a contested force-close reveals a channel's settled size on-chain. Hiding that too is the not-built T0.6 upgrade.
</p>

<div class="note">
<strong>No Proofs of Innocence.</strong> No check stands between a user and the pool. Deposits, transfers, withdrawals and channel deposits proceed without any proof of origin, and none is planned (ruling 4). A user who must show where value came from uses the voluntary, off-chain <a href="engineering/16-selective-disclosure.html">selective disclosure</a> protocol instead; it never gates pool use.
</div>

<hr/>

## References {#refs}

<ol class="refs">
<li id="r-poc"><strong>armada-poc</strong> — the shielded pool, governance, crowdfund, and adapter contracts. <code>origin/main@102f17b8</code>.</li>
<li id="r-circuits"><strong>armada-circuits</strong> — the `Transact(N,M)` circuit template and Phase-2 ceremony manifest. <code>ship-armada.github.com/main@ef2682f</code>; tag <code>v1.0.0-ceremony</code>.</li>
<li id="r-nitro"><strong>armada-nitro</strong> — `ArmadaNitroAdjudicator`, patched `go-nitro`, `HubCustodian`, and the nimbus pair in the fixturenet. <code>origin/integrated@990c9fa9</code>.</li>
<li id="r-sdk"><strong>armada-sdk</strong> — the browser Groth16 prover and client SDK. <code>origin/main@db5757f</code>.</li>
<li id="r-interface"><strong>armada-interface</strong> — the live web client. <code>origin/main@6968cd7</code>.</li>
<li id="r-relayer"><strong>armada-relayer</strong> — the Ponder/Postgres indexer and flat-fee actor. <code>origin/main</code>.</li>
<li id="r-client"><strong>armada-client</strong> — the unbuilt mobile-stack plan, `docs/SPEC.md`. <code>origin/main@78a8fdc</code>.</li>
<li id="r-gonitro"><strong>go-nitro</strong> (upstream, vendored into armada-nitro) — <code>cerc-io/go-nitro@435eb2b</code>.</li>
<li id="r-forcemove"><strong>ForceMove.sol</strong> (<code>challenge</code> L39, <code>checkpoint</code> L90, <code>conclude</code> L127). <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/ForceMove.sol#L39-L127">go-nitro · ForceMove.sol</a>.</li>
<li id="r-nimbus"><strong>nimbus-eth1 verified proxy</strong> — the proof path (T1), proven by hand on the fixturenet. <a href="https://github.com/status-im/nimbus-eth1">github.com/status-im/nimbus-eth1</a>, `master@4df52ef`; fixturenet fixes at <a href="https://github.com/AFDudley/nimbus-eth1">AFDudley/nimbus-eth1</a> `armada/verified-proxy-fixes@b0205d1d`.</li>
<li id="r-frost"><strong>frost-secp256k1-evm</strong> — the hub custodian's threshold scheme (T2). Pinned submodule <code>65d03d0</code> in `armada-nitro`.</li>
<li id="r-railgun"><strong>Railgun</strong> — the shielded-pool design Armada's own pool and circuits implement. <a href="https://github.com/Railgun-Privacy/contract">contract</a>, <a href="https://github.com/Railgun-Privacy/circuits-v2">circuits-v2</a>.</li>
</ol>

<p class="small" style="margin-top:26px">`watcher-ts`, `ts-nitro`, `@cerc-io/peer`, `mobymask`, `laconic-wallet`, and `ipld-eth-*` have no code anywhere in a ship-armada repo (ruling 3, 7, brief §4).</p>
