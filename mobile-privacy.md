# Mobile end-to-end privacy

<p class="lede">A shielded pool only makes a user private if <em>every</em> surface around it is private too — the network path, how the client reads the chain, how transactions reach the chain, and where keys live. This document follows a single transaction through each of those surfaces, names what is hidden and from whom, cites the code that implements it, and states what remains a plan. There is no mobile client yet; "mobile" below names a plan, not a shipped surface.</p>

<p style="margin-top:14px">
  <span class="tag">audience: Armada engineering</span>
  <span class="tag">non-custodial</span>
  <span class="tag">web client today</span>
</p>

<div class="panel">
<strong>Status.</strong> The live client, `armada-interface`, is a browser web app. No mobile build, no Waku transport, no libp2p, and no Nym mixnet integration exist in any ship-armada repo. Most of this page describes a plan (ruling 7, brief §4). Where a property is built, it is labeled <strong>built</strong>; everything else is a plan with no code.
</div>

## Contents {#contents}

<div class="toc">
  <a href="#threat">1. Threat model — who observes what</a>
  <a href="#flow">2. The end-to-end flow</a>
  <a href="#l1">3. Network &amp; transport (plan)</a>
  <a href="#l2">4. Read &amp; submit</a>
  <a href="#l3">5. Settlement, identity &amp; custody</a>
  <a href="#leaks">6. What still leaks</a>
  <a href="#difficulty">7. What remains to build</a>
  <a href="#refs">References</a>
</div>

## 1. Threat model — who observes what {#threat}

"Private" is meaningless without naming the observer. A user transacting against a shielded pool is potentially exposed to several distinct parties, each of which sees a different slice of the interaction.

<table>
<tr><th>Observer</th><th>What they could learn without mitigation</th><th>Blinded by</th><th>Status</th></tr>
<tr><td>Network observer / ISP / gateway</td><td>The device's IP, and traffic timing/volume that correlates sender ↔ recipient</td><td>an optional mixnet underlay</td><td>plan</td></tr>
<tr><td>RPC / data provider</td><td>Exactly which commitments, nullifiers, and addresses the wallet queries — enough to fingerprint the user against the pool</td><td>verified reads through the nimbus proxy (§4) ★</td><td>in progress, unintegrated</td></tr>
<tr><td>A future execution venue</td><td>Who you are, your funding source, your history, and that two trades are the same person</td><td>the shield boundary inside `ArmadaNitroAdjudicator` (§5)</td><td>boundary built; no venue exists</td></tr>
<tr><td>On-chain observers / mempool</td><td>Your EOA as the transaction origin, and the amounts you move in the clear</td><td>a keeper/broadcaster submitting on your behalf (§4) + the shielded pool (§5)</td><td>pool built; keeper/broadcaster plan</td></tr>
<tr><td>The counterparty</td><td>Your live IP; identity attributes beyond what a trade requires</td><td>a mixnet underlay (§3) + selective-disclosure identity (§5)</td><td>plan</td></tr>
</table>
<p class="small">★ marks the single property Armada's anonymity-set guarantee rests on: if the read path leaks, every downstream protection is moot because the observer already knows which notes are yours. Today no client verifies a read; `armada-interface` trusts whatever RPC it is configured with.</p>

## 2. The end-to-end flow {#flow}

The stages below trace one action: a user shielding, holding, and unshielding value. The section numbers in brackets map each stage to its detailed treatment. Stages marked **(plan)** have no code today.

<pre class="diagram">
  DEVICE            browser-memory keys today (armada-interface)             [§5]
    │  no BIP-39/HD custody, no secure enclave, no mobile build
    ▼
  NETWORK / IP      (plan) a mixnet underlay would hide IP + traffic metadata [§3]
    │  today: plain HTTPS to whatever RPC/relayer endpoint is configured
    ▼
  READ  (sync)      today: plain RPC, trusted; (plan) verify via nimbus proxy [§4] ★
    │  phone/browser scans + decrypts notes LOCALLY once synced
    ▼
  SUBMIT (write)    today: the user's own key submits; (plan) a broadcaster   [§4]
    │  would submit on the user's behalf so msg.sender ≠ the user
    ▼
  SETTLE (value)    shielded pool (notes/nullifiers) + Nitro channel          [§5]
    │  non-custodial · force-close escape hatch · boundary inside the adjudicator
    ▼
  IDENTITY         (plan) optional, selectively-disclosed proof               [§5]
       no identity mechanism exists in the client today
</pre>

## 3. Network & transport (plan) {#l1}

Before any application-level privacy matters, the network path itself must not betray the user. Transport encryption is not network privacy: even a design that authenticates and encrypts a connection still lets both endpoints see each other's IP address. Today `armada-interface` makes plain HTTPS requests to its configured RPC endpoint and to `armada-relayer`; no mixnet, no Waku, and no libp2p transport exist in any ship-armada repo.

An earlier design called for routing all traffic through a Nym mixnet underlay, with a fast 2-hop mode for latency-sensitive paths and a slower 5-hop mode for maximal metadata protection, plus two libp2p transport modes above it (Waku pub/sub for discovery/gossip, direct libp2p-noise streams for a low-latency hot loop). None of this is built. A `cerc-io/mtm-vpn-client` fork of Nym's client is referenced in the earlier design as available to reuse, but no Armada repo integrates it.

<table>
<tr><th>Component</th><th>What it would hide — and from whom</th><th>Status</th></tr>
<tr><td>Nym mixnet underlay</td><td>Device IP and traffic timing/volume, from ISP, gateways, and the counterparty</td><td>plan — not integrated by any Armada repo</td></tr>
<tr><td>Waku pub/sub</td><td>Which subscriber consumes a message (topic anonymity set)</td><td>plan — no dependency in any client/relayer/interface repo</td></tr>
<tr><td>libp2p-noise</td><td>Transport encryption + peer-id authentication; no IP privacy on its own</td><td>plan — same</td></tr>
</table>

## 4. Read & submit {#l2}

This layer contributes two privacy properties. Interest-privacy would ensure no one learns which notes a user reads; origin-privacy would ensure the on-chain transaction sender is not the user's address. Neither is complete today.

### Read-time private sync ★ {#read}

This is the property the whole design rests on. A shielded pool hides a user only inside a crowd, but that guarantee collapses if the way the user reads the pool is revealing. Over a conventional RPC endpoint, the provider observes every call the wallet issues and can fingerprint precisely which commitments and nullifiers the user scans for. Today `armada-interface`'s scanner (`src/lib/shielded/sync.ts`) reads over whatever RPC it is configured with and trusts the answer; it issues no verification step.

The fix in progress is the **nimbus verified proxy** (T1): a client that routes reads through it would verify storage against a beacon-anchored header (EIP-1186 proofs) and verify a block's receipts by rebuilding the receipt trie against `header.receiptsRoot`, so the client trusts the math, not the RPC provider. No Armada client integrates the proxy yet; no server runs it against any Armada chain; `libverifproxy.a` has no mobile cross-compile target. This replaces the earlier plan's watcher-based "proof-carrying, everyone-gets-the-same-bytes" feed, which depended on `watcher-ts` and a nimbus-eth1 state-diff emitter that do not exist either.

### Write-time origin privacy {#submit}

An earlier design called for a broadcaster or keeper to submit the user's transaction and pay the gas, so the on-chain `msg.sender` is not the user's address. Today `armada-interface` submits transactions directly from the user's own key; no broadcaster, keeper, or relay-on-behalf path exists for settlement.

<table>
<tr><th>Component</th><th>What it would hide — and from whom</th><th>Status</th></tr>
<tr><td>★ Verified read-time sync (nimbus proxy)</td><td>Which notes/addresses/commitments you care about, from the RPC provider</td><td>proxy exists upstream, unintegrated (T1)</td></tr>
<tr><td>Write-time origin privacy (broadcaster/keeper)</td><td>Your EOA as tx origin, from chain and mempool observers</td><td>plan — not built</td></tr>
</table>

## 5. Settlement, identity & custody {#l3}

The final layer moves value privately, keeps settlement non-custodial, and treats identity as an optional, undesigned-but-planned attribute rather than a default.

### Value & settlement {#value}

Value privacy comes from Armada's own shielded pool — a zk-SNARK construction of notes, commitments, and nullifiers that hides amount, sender, recipient, and cross-trade linkage, built and deployed (T0.0). Settlement runs over non-custodial Nitro state channels: funds are held by `ArmadaNitroAdjudicator` and moved only by user-signed state, so a misbehaving counterparty can stall a user but never seize funds; the user unilaterally force-closes to their last co-signed state (T0.2, built). The pool and the channel are joined inside the same adjudicator contract — `depositFromNote` funds a channel from a note, and a `SHIELD` payout re-shields a channel's outcome — with no separate boundary contract. No client drives a channel lifecycle yet (T6.2, partly built): `armada-interface` calls the pool's shield/transact entry points but has no open/fund-from-note, propose/accept, or close code.

### Identity {#identity}

Identity stays optional and has no shipped mechanism. An earlier design named a laconicd onboarding module and an FHE-based matchmaker as precedents; neither is part of any ship-armada repo or plan. The credible option on the table is a Self zk-passport attestation — proven on-device, verified off-chain, anchored to an Ethereum address, with the shielded pool identity left unlinked (T6.8). This is not built and not scheduled.

### Device key custody {#custody}

Today, `armada-interface` holds keys in browser memory (`src/lib/shielded/keyManager.ts`). There is no BIP-39/HD custody chain, no OS secure enclave, and no mobile code. An earlier design called for on-device BIP-39 mnemonic generation, HD key derivation, and secure-enclave storage via `react-native-keychain`; none of that exists, because no React Native client exists.

<table>
<tr><th>Component</th><th>What it hides — and from whom</th><th>Status</th></tr>
<tr><td>Shielded pool</td><td>Amounts, sender, recipient, cross-trade linkage — from a venue, RPC, and chain observers (∝ anonymity-set size)</td><td>built (T0.0)</td></tr>
<tr><td>Deposit/payout boundary (inside the adjudicator)</td><td>Who you are, fund origin, history — from a future venue</td><td>built (T0.2)</td></tr>
<tr><td>Non-custodial Nitro channels</td><td>Not privacy: guarantees no custody — force-close to last co-signed state</td><td>built (T0.2); no client drives it yet</td></tr>
<tr><td>Identity attestation</td><td>Would gate nothing by default; optional proof to a counterparty</td><td>plan — not built (T6.8)</td></tr>
<tr><td>Device key custody</td><td>Keys stay on-device</td><td>built, browser memory only — not HD, not secure-enclave</td></tr>
</table>

## 6. What still leaks {#leaks}

- **No read is verified today.** `armada-interface` trusts its configured RPC. Until the nimbus verified proxy is integrated, a malicious or compromised RPC provider could serve a tampered answer undetected, and any RPC provider can see every query the wallet makes.
- **The transaction sender is the user's own key.** Without a broadcaster or keeper, every on-chain call from `armada-interface` is attributable to the user's address.
- **Amounts are public only at the boundary.** Arbitrary amounts are hidden inside the pool; the only exposure is the transparent boundary (shield/unshield, deposit, payout, force-close). A large lone public crossing stands out; once inside, transfers are private.
- **No network-level IP privacy exists.** Every request the browser makes exposes the user's IP to whatever server receives it.

## 7. What remains to build {#difficulty}

The hard work ahead is engineering and integration, not new cryptography: the pool, circuits, and adjudicator already exist and need no redesign for this page's purposes. Ranked by what blocks what:

<table>
<tr><th>#</th><th>Item</th><th>Why it matters</th><th>Status</th></tr>
<tr><td>1</td><td>★ Integrate the nimbus verified proxy into a client</td><td>Anonymity-set-critical: until a client verifies reads, every other protection on this page is moot. The proxy exists upstream; no client wires it in, and the fixturenet needs a genesis patch first (T1.2)</td><td>not built</td></tr>
<tr><td>2</td><td>A channel-driving client</td><td>No code in `armada-interface` or `armada-sdk` opens, funds, or closes a Nitro channel; today's client only shields and transacts against the pool directly</td><td>not built (T6.2)</td></tr>
<tr><td>3</td><td>A keeper/broadcaster for origin privacy</td><td>Needed so the on-chain sender is not the user's own address</td><td>not built</td></tr>
<tr><td>4</td><td>A mobile client</td><td>`armada-client`'s `docs/SPEC.md` plans one TypeScript stack across Linux server, browser, iOS, and Android; zero lines of `src/` exist</td><td>not built (T6.5)</td></tr>
<tr><td>5</td><td>A native mobile prover</td><td>A browser Web Worker `snarkjs` prover exists (T6.1); no native mopro/rapidsnark prover exists for a phone</td><td>not built (T6.6)</td></tr>
<tr><td>6</td><td>A network-privacy underlay</td><td>No mixnet, Waku, or libp2p integration exists anywhere in the client stack</td><td>plan</td></tr>
</table>

### Built today

- **Shielded pool** — live and deployed (T0.0).
- **Nitro adjudicator, with the merged deposit/payout boundary** — live on Ethereum (T0.2).
- **Browser Groth16 prover** — a `snarkjs` Web Worker in `armada-interface`/`armada-sdk` (T6.1).
- **Device key custody (browser only)** — `armada-interface`'s in-memory key manager (T6.0).

<div class="note">
<strong>Bottom line.</strong> The do-first item is integrating the nimbus verified proxy (§4): until a client verifies reads, the pool's anonymity-set guarantee is only a design intention, not a property any user has today.
</div>

<hr/>

## References {#refs}

<ol class="refs">
<li id="r-poc"><strong>armada-poc</strong> — the shielded pool. <code>origin/main@102f17b8</code>.</li>
<li id="r-nitro"><strong>armada-nitro</strong> — `ArmadaNitroAdjudicator`, patched `go-nitro`. <code>origin/integrated@55939ef</code>.</li>
<li id="r-interface"><strong>armada-interface</strong> — the live web client. <code>src/lib/shielded/{sync.ts,keyManager.ts,sdk-prover.ts}</code>, <code>src/features/{shield,transfer-shielded,unshield}/handler.ts</code>. <code>origin/main@6968cd7</code>.</li>
<li id="r-sdk"><strong>armada-sdk</strong> — the browser Groth16 prover. <code>src/prover/index.ts</code>. <code>origin/main@db5757f</code>.</li>
<li id="r-client"><strong>armada-client</strong> — the unbuilt mobile-stack plan. <code>docs/SPEC.md</code>, <code>origin/main@78a8fdc</code>.</li>
<li id="r-nimbus"><strong>nimbus-eth1 verified proxy</strong> — the proof path (T1), unintegrated. <a href="https://github.com/status-im/nimbus-eth1">github.com/status-im/nimbus-eth1</a>, <code>master@4df52ef</code>.</li>
<li id="r-forcemove"><strong>ForceMove.sol</strong> (challenge/checkpoint/conclude), vendored into `armada-nitro`. <a href="https://github.com/cerc-io/go-nitro/blob/435eb2b02777d740483f5b5953ed5b88ef90c665/packages/nitro-protocol/contracts/ForceMove.sol">go-nitro · ForceMove.sol</a>.</li>
<li id="r-mopro"><strong>Mobile proving — benchmarks &amp; alternatives</strong> — on-device Groth16 timings and alternatives, relevant once a native mobile prover is scoped. <a href="mobile-proving-research.html">mobile-proving-research</a>.</li>
</ol>

<p class="small" style="margin-top:26px">Companion to <a href="architecture.html">Architecture</a> (tier stack) and the <a href="engineering/build-plan.html">Build plan</a> (status &amp; backlog). Earlier references to Nym, Waku, libp2p, <code>watcher-ts</code>, MobyMask, <code>laconic-wallet</code>, and a Laconic FHE matchmaker are removed: none of them are reused in any ship-armada repo (ruling 7, brief §4).</p>
