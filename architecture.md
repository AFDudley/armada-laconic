# Architecture — the tier stack

<p class="lede">Armada settles private execution on a Nitro rail. The motion is the same everywhere: notes in → normal Nitro → notes out. Value leaves the Railgun shielded pool, moves through ordinary Nitro state channels off-chain, and returns as shielded notes. That rail carries a mobile-first, non-custodial, front-running-resistant venue, delivered through bonded Laconic <em>watcher parties</em> beneath Armada's Adapters tier. Seven tiers, each mapped to existing code where it exists.</p>

<p style="margin-top:14px">
  <span class="tag">non-custodial</span>
  <span class="tag">Ethereum L1-anchored</span>
  <span class="tag">Nitro settlement</span>
  <span class="tag">pool immutable / untouched</span>
</p>

<div class="note">
Tiers are named in the software-architecture sense (an n-tier stack), not the "Layer&nbsp;1 / Layer&nbsp;2" of blockchains. Every claim maps to code on the <a href="build-plan.html">Build plan</a>, with pinned commits and status. This page is the structure; the build plan is the status and backlog. Code links use GitHub where a mirror exists, otherwise the canonical Laconic Gitea (<code>git.vdb.to</code>).
</div>

## Contents

<div class="toc">
  <a href="#shape">The shape of the network (mobile-first)</a>
  <a href="#settlement">The settlement rail — notes in, normal Nitro, notes out</a>
  <a href="#t0">T0 · Ethereum anchor — settlement &amp; shielded custody</a>
  <a href="#t1">T1 · Ingestion (nimbus-eth1 → watchers)</a>
  <a href="#t2">T2 · Watcher-party substrate</a>
  <a href="#t3">T3 · Ordering &amp; fault model</a>
  <a href="#t4">T4 · Execution / settlement platform (ex_net)</a>
  <a href="#t5">T5 · Adapters (Armada's tier)</a>
  <a href="#t6">T6 · Client / apps (mobile-first)</a>
  <a href="#identity">Identity — optional match filter</a>
  <a href="#privacy">Cross-cutting · Privacy boundary (Design A)</a>
  <a href="#refs">References</a>
</div>

## The shape of the network — mobile-first {#shape}

Every service is delivered through watcher parties, so most of the network is browser wallets and mobile apps talking peer-to-peer. The only fixed infrastructure is (a) a handful of **STUN/TURN** servers for NAT traversal and (b) a small number of **nimbus-eth1** state-diff emitters that feed L1 state into the parties. Everything else (sync, metering, matching, settlement, threshold attestation) is run by the bonded service-provider federations and the clients.

<pre class="diagram">
        browser wallet / mobile app                 browser wallet / mobile app
    ( Laconic wallet (Android/iOS/web)  +  in-browser Nitro node  +  libp2p relay )     T6
                 \                                              /
                  \___________  p2p (WebRTC/WS) _____________ /
                        |   NAT traversal via STUN/TURN  |         &lt;- only fixed infra #1
                        v                                v
            ┌─────────────────────────────────────────────────────┐
            │            WATCHER PARTY  (bonded SP federation)      │  T2–T4
            │  shared cache · Nitro clearing · matching/venue       │
            │  commit-reveal sequencing · threshold Schnorr (DSS)   │
            └─────────────────────────────────────────────────────┘
                        ^                                ^
                        |  proof-carrying L1 state       |               T1
            ┌───────────────────────────┐                |
            │  nimbus-eth1 emitters     │  &lt;- only fixed infra #2
            │  (state diffs → IPLD)     │
            └───────────────────────────┘
                        ^
                        |
            ┌─────────────────────────────────────────────────────┐
            │   ETHEREUM L1:  Railgun shielded pool  +  Nitro       │  T0
            │   NitroAdjudicator / ForceMove / MultiAssetHolder     │
            └─────────────────────────────────────────────────────┘
</pre>

<p class="small">
Clients hold their own keys and funds (channels are L1-collateralized); a watcher party never takes custody and can be exited unilaterally. The two boxes of fixed infra stay small: STUN/TURN carry no trust, and the nimbus-eth1 emitters only publish verifiable state.
</p>

## The settlement rail — notes in, normal Nitro, notes out {#settlement}

Every adapter resolves to one motion. Value leaves the Railgun shielded pool as notes, moves through ordinary go-nitro state channels off-chain, and returns to shielded notes at the end. Nitro does the settling; the shielded pool holds custody at the boundary, and the pool contract is never touched.

A cross-chain swap is that shape stretched across two chains. A shielded A→B swap prices and matches inside a Nitro channel, and the two legs settle against mirrored channels through the `Bridge` construction<sup><a href="#r-bridge">[10]</a></sup>, so neither side leaves the trader's custody until the co-signed outcome is final. The construction is documented end to end in the [nitro-on-railgun overview](engineering/A-nitro-on-railgun/A.0-overview.html) and the [shielded cross-chain-swap design](engineering/shielded-nitro-bridge-design.html).

## T0 · Ethereum anchor — settlement & shielded custody {#t0}

Settlement is non-custodial Nitro state channels on Ethereum L1. Funds are locked in the `NitroAdjudicator` / `MultiAssetHolder` and allocated off-chain by co-signed state; a channel closes cooperatively at internet speed, or unilaterally via the `ForceMove` dispute game (`challenge`, `checkpoint`, `conclude`)<sup><a href="#r-forcemove">[5]</a></sup>. This is the escape hatch that makes the venue safe: if a party misbehaves, the user force-closes to their last co-signed state, and no settlement receipt means the force-close returns the pre-fill deposit. The adjudicator only ever honors user-signed states, so a byzantine federation can stall you but cannot fabricate balances or move unsigned funds.

<pre>contract HashLockedSwap is IForceMoveApp {
    struct AppData { bytes32 h; bytes preImage; }   // reveal preimage of h to unlock
    ...
}</pre>

<p class="src">source: <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/examples/HashLockedSwap.sol#L11-L33">go-nitro · HashLockedSwap.sol#L11–L33</a></p>

Cross-chain and L1↔L2 movement reuses the `Bridge` mirrored-channel construction<sup><a href="#r-bridge">[10]</a></sup>. **Built here:** the Railgun shielded pool (Armada's, immutable) and the full Nitro adjudicator / ForceMove / MultiAssetHolder<sup><a href="#r-forcemove">[5]</a></sup>, live on Ethereum. **Net-new here:** a venue/party registry and bond contract, an on-chain sequencing-cert and fraud-proof verifier, and the Nitro↔Railgun boundary adapter (see [Privacy](#privacy)). Full status on the [Build plan](build-plan.html#t0).

## T1 · Ingestion (nimbus-eth1 → watchers) {#t1}

L1 state reaches the watcher parties through a `nimbus-eth1`<sup><a href="#r-nimbus">[16]</a></sup> state-diff emitter, whose stateless/witness and Aristo primitives<sup><a href="#r-nimbusstateless">[17]</a></sup> emit full, proof-carrying Ethereum state diffs (intermediate and leaf MPT trie nodes) from a bounded working set, indexed and served by the IPLD services, `ipld-eth-server`<sup><a href="#r-ipldserver">[18]</a></sup> and `ipld-eth-state-snapshot`<sup><a href="#r-snapshot">[19]</a></sup>.

This is what makes private sync possible. A watcher party ingests the Railgun shielded-pool commitments and nullifiers and the Nitro adjudicator events through the nimbus-eth1 emitter, then serves proof-carrying, identical-for-everyone note streams metered by Nitro. Clients scan those streams **locally**, so no RPC provider ever fingerprints a user against the pool, the single property Armada's anonymity set depends on. A few emitters serve the whole network; the phones and browsers do the rest. The generic `watcher-ts` framework<sup><a href="#r-watcherts">[2]</a></sup> exists; the emitter module and the Railgun/adjudicator-specific watcher config are the build work.

## T2 · Watcher-party substrate {#t2}

A **watcher party** is a bonded federation of service providers that collectively serves one data need. Here it is also the execution venue: it delivers Armada's private support services and clears the associated payments over Nitro.

<table>
<tr><th>Service</th><th>Delivered by the watcher party as…</th></tr>
<tr><td>Private shielded-pool sync</td><td>a shared cache serving proof-carrying, everyone-gets-the-same-bytes note streams (clients scan locally &amp; privately)</td></tr>
<tr><td>Metering / clearing</td><td>per-request Nitro vouchers<sup><a href="#r-vouchers">[8]</a></sup> netted into co-signed channel allocations over virtual channels<sup><a href="#r-vfund">[9]</a></sup>, settled to L1 in USDC</td></tr>
<tr><td>Matching / venue</td><td>commit-reveal sequencing + a matcher over revealed orders (T3–T4)</td></tr>
<tr><td>Attestation</td><td>threshold-Schnorr (DSS) inclusion receipts, sequencing certs, liquidity-proof snapshots</td></tr>
<tr><td>Membership / bonding</td><td>an on-L1 registry + bond; disputes and slashing settle on L1 (T0)</td></tr>
</table>

<div class="note">
Why "party," not "chain": the federation only needs to (a) agree on <em>which</em> commitments are in an epoch and (b) fix an order, a sequencing-and-attestation task rather than replicated global execution. That is a far weaker (and cheaper) primitive than BFT consensus, and it is what lets the network live on phones.
</div>

### The federation signature — threshold Schnorr (DSS)

The party signs everything it attests to with a **threshold Schnorr signature**. The `chain-signatures` library provides Ethereum-compatible Schnorr (`ethschnorr.Sign` / `Verify`)<sup><a href="#r-ethschnorr">[12]</a></sup> and the Distributed Schnorr Signature protocol (Stinson &amp; Strobl <em>(t, n)</em>) in `ethdss`<sup><a href="#r-ethdss">[13]</a></sup>, built on a `kyber` DKG. Two properties matter:

- **On-chain verifiable.** Because it is Ethereum-flavoured Schnorr, an L1 contract can verify a party's aggregate signature, so a sequencing cert or a censored-commit receipt becomes a slashable fraud proof against the bond (T0).
- **Threshold direction is safety-first.** *t*-of-*n* tolerates *t−1* malicious for safety and *n−t* offline for liveness. We pick *t* high (e.g. 4-of-7) so forging an attestation needs a large coalition; a liveness failure degrades to halt, not loss, because settlement is non-custodial. A BFT-style small quorum like 3-of-11 would be wrong, since it would let any 3 forge.

<div class="note">
The threshold key is used for <strong>signing only</strong>. We do not run a threshold-encrypted mempool: with the sequencer and key-holder being the same federation, encryption gives no real fairness (a colluding threshold can decrypt-then-order). Fair ordering is achieved with commit-reveal instead (T3).
</div>

## T3 · Ordering & fault model {#t3}

### Fair ordering = commit-reveal

Orders are hidden from the venue during ordering with **commit-reveal** rather than client-side ZK: a phone generating a SNARK per order costs seconds, a reveal round-trip is tens of milliseconds, and the user is already interactive through settlement. The flow: `commit = H(order ‖ salt)` → the party seals the epoch and emits threshold-signed inclusion receipts (so censorship is slashable) → a post-seal randomness beacon fixes positions → the user reveals → the matcher runs over revealed orders in the fixed order. The party cannot content-front-run (it only saw hashes) or position-manipulate (the beacon is post-seal). ZK order submission stays an opt-in for offline submitters or server-side provers.

### Latency & receipts

A cooperative fill is **3 RTT + Δ** (commit, reveal, settle), with Δ ≈ the party's co-located sealing window (single-digit ms). Settlement is one round trip whose two legs are both mandatory: the user signs first, and the venue returns a bond-enforced settlement receipt. There is no fire-and-forget; every fill terminates in either a co-signed receipt or a slashable proof of its absence.

### Griefing is a non-issue here

Because the venue is the one providing asks, a user who withholds a reveal harms no third party: the ask is a standing quote available to all, so a non-revealed commit blocks no other user's liquidity. The only theoretical residual is a sub-second free option on a stale quote, worth roughly nothing as long as the reveal timeout ≤ the quote-repricing interval (a knob the venue controls). Anti-griefing collapses to a nominal anti-spam commit fee.

### Failure = halt, not loss

If the party stalls or drops below its liveness threshold, users force-close to their last signed state and are made whole; a **watchtower** guarantees the latest state even for an offline user. Non-custody means a dead party freezes trading, never funds.

<p class="small">In Ethereum terms this tier is a decentralized sequencer + DA + fraud-proof ordering over state-channel settlement, not a rollup: there is no shared L2 state root, and balances are enforced by the channel adjudicator. See the <a href="build-plan.html#t3">Build plan</a>; this is the highest-design-risk, mostly net-new tier.</p>

## T4 · Execution / settlement platform (ex_net) {#t4}

This is the platform that Armada's adapters run on. The matcher is a frequent-batch-auction, curve-order, filler-filtered CLOB/RFQ hybrid that prices, matches, and settles privately into Nitro channels. Its own page, [Execution platform](execution-platform.html), covers it in depth: the ex_net lineage, the 2019–2020 `matcher` PoC, capital efficiency vs. AMMs, and how Swaps and yield become applications on top. Settlement reuses go-nitro's multi-asset swap protocol<sup><a href="#r-swap">[6]</a></sup> and `SwapChannel`<sup><a href="#r-swapchan">[7]</a></sup>; the net-new pieces are the LP/market-maker vault, receipt-or-slash wiring, and multi-asset outcomes.

<div class="note"><strong>v1 vs v2.</strong> This matcher is a <strong>v2</strong> feature (price discovery and market-making). Armada's <strong>v1</strong> execution uses a posted price: a single <code>priceSetter</code> (one L1 wallet → governance) publishes bid/ask on both sides, take-it-or-leave-it, cleared over Nitro. Nothing is left to front-run, so v1 needs no matcher and no fair-ordering (T3). See <a href="yield-clearing.html">Yield &amp; clearing</a>.</div>

## T5 · Adapters (Armada's tier) {#t5}

Armada's Swaps and Aave-v4 yield adapters are applications on the venue, not peers to it. In **v1**, Swaps clear at the posted price, and the value-moving adapter is a Railgun RelayAdapt recipe (atomic unshield → call → reshield), with no custodial vault and no DSS. For yield, ETH is a shielded wstETH note (intrinsic and non-rebasing, so no Aave or LP is needed) and USDC is an LP-buffered Aave rail; see [Yield & clearing](yield-clearing.html). CCTP (cross-chain USDC) is Armada's, already built. Nothing here touches the immutable pool.

<div class="note">
<strong>Audit boundary.</strong> Positioning ex_net under multiple adapters makes it a shared dependency with its own audit surface, a larger ask than one independently-deployed adapter, and a deviation from Armada's "each adapter independently deployed &amp; audited" discipline.
</div>

### The CCTP adapter — cross-chain USDC on/off-ramp

"Privacy on Ethereum" describes where the shielded pool lives, not where the money is. Circle mints native USDC on around a dozen chains, so most users' USDC is not already on mainnet. The **CCTP adapter** is the cross-chain rail that fills and drains the pool, an adoption and liquidity adapter rather than a privacy primitive. CCTP is Circle's burn-and-mint protocol: burn on the source chain, mint native USDC on the destination against a Circle attestation, non-custodial, 1:1, with no slippage and no wrapped-asset or bridge-hack risk. It matters because a shielded pool's privacy scales with its crowd size, so deposits must be sourceable from wherever USDC actually sits, and private balances must be able to exit to other chains.

As an adapter it chains two operations into one user action, without touching the immutable pool. Inbound (mint-and-shield): burn on the source chain → attestation → mint on Ethereum, with a CCTP v2 hook depositing the minted USDC straight into the Railgun pool at a fresh, unlinkable destination. Outbound (unshield-and-burn): unshield → burn on Ethereum → mint on the destination chain.

<div class="note">
<strong>CCTP is transport, not privacy.</strong> The burn shows sender and amount on the source chain and the mint shows recipient and amount on the destination; if the same public addresses appear on both ends the cross-chain link is trivial. Privacy comes entirely from binding the mint <em>into the shield</em> (or sourcing the burn from it). Residual leaks: amount and timing correlation across the hop (mitigate by atomic batched aggregation and staying shielded; a large lone crossing stands out on its own), and the mint transaction's origin. That last point is why the destination mint should be submitted by the broadcaster/keeper (§ Layer 2 of <a href="mobile-privacy.html">Mobile privacy</a>) over the Nym/Waku private transport, not from the user's own address. Native USDC is freezable by Circle and CCTP depends on Circle's attestation service, a centralization and censorship consideration orthogonal to the adapter mechanics.
</div>

## T6 · Client / apps (mobile-first) {#t6}

The client is the **Laconic wallet**, a React Native app on Android and iOS<sup><a href="#r-lacwallet">[21]</a></sup> with a browser build<sup><a href="#r-lacwalletweb">[22]</a></sup>. It custodies keys and signs both Cosmos and EIP-155 requests over WalletConnect, and pairs with an in-browser or mobile Nitro node<sup><a href="#r-tsnitro">[4]</a></sup> for payments and a libp2p relay. The wallet is bare-bones today, and `MobyMask`<sup><a href="#r-mobymask">[3]</a></sup> and the swap demos are demo fragments rather than user-facing products; they demonstrate the mobile-first shape (app-specific signing, in-browser Nitro, and a p2p relay to a keeper) rather than a finished app. T6 is a real build: grow the wallet into the Armada-branded front-end, wire the Armada SDK, and add wallet updates (return pubkey on `cosmos_signAmino`, Nitro and commit-reveal integration, Railgun boundary UX).

## Identity — an optional match filter {#identity}

Identity is an optional, well-defined attribute a party can attach to itself and optionally require of counterparties, evaluated by the same matcher as any other filter (pair, size, price, allow/block lists). It is not a separate layer. The default is identity-blind, and a party opts in only when a counterparty or a jurisdiction requires it. What follows is how such an attribute is sourced and proved; all of it is optional.

### ex_net identity proofs

ex_net<sup><a href="#r-exnet">[1]</a></sup> defines the primitive: a user links an external key to their trading identity by signing a message with the external private key that embeds their venue public key. In its forward-looking form the user can prove a match (e.g. passport ↔ liveness) and record only the proof, never the underlying data. Identity is an optional, user-initiated, selectively-disclosed proof beside the shielded default; the venue stays identity-blind unless a counterparty or jurisdiction requires a proof. The shipped precedent is laconicd's onboarding module, whose on-chain `Participant` record carries `role` and `kyc_id`<sup><a href="#r-onboard">[14]</a></sup>.

### The Laconic Matchmaker

The Laconic Matchmaking Services proposal<sup><a href="#r-match">[15]</a></sup> supplies the compliance-gating half: a privacy-preserving matchmaking service that pairs technically-qualified operators with capital backers "in a way that's compliant with international broker-dealer laws and respects their users' privacy," implemented as encrypted computation that matches number ranges and finds intersections of number intervals. In exchange terms, the matcher can verify that two parties satisfy each other's constraints (jurisdiction, accreditation, size band, KYC status) without either party, or the venue, learning the other's identity or exact values. The underlying FHE or encrypted-computation is an implementation detail; the design depends only on the two abstractions above.

<p class="small"><strong>Off-the-shelf path, Self (zk-passport), optional future work (not built).</strong> The selective-disclosure identity above has a credible off-the-shelf option in <a href="https://github.com/selfxyz/self">Self</a> (zk-passport, formerly OpenPassport): the user proves attributes (nationality, age, OFAC-clear, personhood) on-device from a passport or EU-ID chip via the standalone Self app, and Armada verifies the returned attestation off-chain, anchored to an Ethereum address (a nullifier gives uniqueness; the shielded Railgun identity stays unlinked). Proof-of-funds stays native, a shielded note-balance proof. It is optional and deferred: none of it is required for the shielded default, and none of it is implemented.</p>

## Cross-cutting · Privacy boundary (Design A) {#privacy}

Privacy is delivered by funding and settling channels through a Railgun-style shielded pool via an adapter<sup><a href="#r-railgun">[20]</a></sup>, the same adapter pattern Armada already uses, with a Nitro adapter added. The venue's visibility begins at the shield boundary and never reaches behind it: it sees the shielded-side order terms it must match and an ephemeral channel identity, but not who you are, where the funds came from, your on-chain history, or whether two trades were the same person. The venue is a blind, non-custodial matcher.

<p class="small">
Caveats: identity privacy is only as strong as the Railgun anonymity set and is subject to amount and timing correlation (mitigate by atomic batched aggregation and a fresh per-channel identity; amounts themselves are hidden inside the shield, so the leak is only at the transparent boundary). The one thing still visible to the venue is the trade's amounts and terms. Hiding those too is <strong>Design B</strong> (ZK matching plus a "shielded ForceMove" dispute over hidden state), out of scope for v1.
</p>

<div class="note">
<strong>Namada is not Design B.</strong> Namada's MASP is a strong multi-asset shielded-note primitive, but Namada is an L1 with global consensus rather than a state channel, so it has no off-chain, non-custodial, unilateral exit and runs at chain speed. It attains privacy the way ex_net attained fairness: with a chain. MASP-the-circuit could serve as a note primitive for a future Design B; Namada-the-chain does not provide the channel-speed exit this design requires.
</div>

<hr/>

## References {#refs}

<ol class="refs">
<li id="r-exnet"><strong>ex_net whitepaper</strong>, Vulcanize (© 2017). <a href="./ex_net_whitepaper.pdf">ex_net_whitepaper.pdf</a> (in this repo).</li>
<li id="r-watcherts"><strong>watcher-ts</strong>. <a href="https://github.com/cerc-io/watcher-ts/blob/18ca4e1a08328770af8e10f15f2128a459c9c704/README.md">github.com/cerc-io/watcher-ts</a> @ <code>18ca4e1</code>.</li>
<li id="r-mobymask"><strong>mobymask</strong> — reference p2p app. <a href="https://github.com/cerc-io/mobymask/blob/23291987766f863e366a344973670c57cd82d678/README.md">github.com/cerc-io/mobymask</a> @ <code>2329198</code>.</li>
<li id="r-tsnitro"><strong>ts-nitro</strong> — in-browser Nitro node. <a href="https://github.com/cerc-io/ts-nitro/tree/884d616a9a026742ef1ce57233801a13d681375a">github.com/cerc-io/ts-nitro</a> @ <code>884d616</code>.</li>
<li id="r-forcemove"><strong>ForceMove.sol</strong> (<code>challenge</code> L39, <code>checkpoint</code> L90, <code>conclude</code> L127). <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/ForceMove.sol#L39-L127">go-nitro · ForceMove.sol</a>; <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/NitroAdjudicator.sol">NitroAdjudicator.sol</a>, <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/MultiAssetHolder.sol">MultiAssetHolder.sol</a>.</li>
<li id="r-swap"><strong>Nitro swap protocol</strong> (Deepstack). <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/protocols/swap/swap.go#L44-L146">go-nitro · protocols/swap/swap.go</a>; demo <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/demo/swaps.md">demo/swaps.md</a>.</li>
<li id="r-swapchan"><strong>SwapChannel</strong>. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/channel/swap.go#L10-L14">go-nitro · channel/swap.go#L10–L14</a>.</li>
<li id="r-vouchers"><strong>Payment vouchers</strong>. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/payments/vouchers.go">go-nitro · payments/vouchers.go</a>.</li>
<li id="r-vfund"><strong>Virtual channels</strong>. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/protocols/virtualfund/virtualfund.go">go-nitro · protocols/virtualfund/virtualfund.go</a>.</li>
<li id="r-bridge"><strong>Bridge.sol</strong> — mirrored L1↔L2 channels. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/Bridge.sol#L8-L23">go-nitro · Bridge.sol#L8–L23</a>.</li>
<li id="r-ethschnorr"><strong>ethschnorr</strong> — Ethereum Schnorr. <a href="https://git.vdb.to/cerc-io/chain-signatures/src/commit/9016a7c4a7f046cac8f57b3335a1afcde346e75f/ethschnorr/ethschnorr.go#L85-L133">git.vdb.to · chain-signatures/ethschnorr</a>.</li>
<li id="r-ethdss"><strong>ethdss</strong> — Distributed Schnorr (Stinson–Strobl (t,n)). <a href="https://git.vdb.to/cerc-io/chain-signatures/src/commit/9016a7c4a7f046cac8f57b3335a1afcde346e75f/ethdss/ethdss.go#L55-L287">git.vdb.to · chain-signatures/ethdss</a>.</li>
<li id="r-onboard"><strong>laconicd onboarding</strong> — <code>Participant</code> (<code>role</code>, <code>kyc_id</code>). <a href="https://git.vdb.to/cerc-io/laconicd/src/commit/d130608f672983cb65acfbddc89976747a417aac/proto/cerc/onboarding/v1/onboarding.proto#L18-L41">git.vdb.to · onboarding.proto</a>.</li>
<li id="r-match"><strong>Laconic Matchmaking Services</strong> proposal, Nov 2022. Internal Google Doc — <a href="https://docs.google.com/document/d/1AkUfIT_fBDgErBUy2YFCCUsPwHhL0WQngVRgKSqjqi0/">Proposal</a>. Privacy-preserving, broker-dealer-compliant matching via encrypted set-intersection / range matching.</li>
<li id="r-nimbus"><strong>nimbus-eth1</strong> — Nim Ethereum L1 execution client (stateless/witness execution + Aristo state DB). <a href="https://github.com/status-im/nimbus-eth1">github.com/status-im/nimbus-eth1</a>.</li>
<li id="r-nimbusstateless"><strong>nimbus-eth1 stateless &amp; Aristo</strong> — witness generation/verification and canonical MPT-node proofs. <a href="https://github.com/status-im/nimbus-eth1/tree/master/execution_chain/stateless">execution_chain/stateless</a>, <a href="https://github.com/status-im/nimbus-eth1/tree/master/execution_chain/db/aristo">db/aristo</a>.</li>
<li id="r-ipldserver"><strong>ipld-eth-server</strong>. <a href="https://github.com/cerc-io/ipld-eth-server/blob/330bc3d7e2dd6acf4f7a1e744d39ea462ecb2323/README.md">github.com/cerc-io/ipld-eth-server</a> @ <code>330bc3d</code>.</li>
<li id="r-snapshot"><strong>ipld-eth-state-snapshot</strong>. <a href="https://github.com/cerc-io/ipld-eth-state-snapshot/tree/9e483fc9f7dfa993e3d1e98a311cb4e45917cd51">github.com/cerc-io/ipld-eth-state-snapshot</a> @ <code>9e483fc</code>.</li>
<li id="r-railgun"><strong>Railgun</strong> — shielded-pool contracts &amp; circuits. <a href="https://github.com/Railgun-Privacy/contract">contract</a>, <a href="https://github.com/Railgun-Privacy/circuits-v2">circuits-v2</a>.</li>
<li id="r-lacwallet"><strong>laconic-wallet</strong> — React Native (Android/iOS). <a href="https://git.vdb.to/cerc-io/laconic-wallet/src/commit/bb5223afdab0d16822bb3c119b6ba2350dab921e/android">git.vdb.to · cerc-io/laconic-wallet</a> @ <code>bb5223a</code>.</li>
<li id="r-lacwalletweb"><strong>laconic-wallet-web</strong> — browser build. <a href="https://git.vdb.to/cerc-io/laconic-wallet-web/src/commit/2a4a478e32d0a677711ac41f93ff40f364851c3a">git.vdb.to · cerc-io/laconic-wallet-web</a> @ <code>2a4a478</code>.</li>
</ol>

<p class="small" style="margin-top:26px">Code links pinned to commits: go-nitro <code>435eb2b</code>, chain-signatures <code>9016a7c</code>, laconicd <code>d130608</code> (<code>roysc/nitro-integration</code>). Internal Google Docs require Laconic/Vulcanize access.</p>
