# Builder codes, attribution & liquidity — TACEO vs ex_net

<p class="lede">Two ways to do private <em>payment for order flow</em> / builder-code attribution — TACEO's cryptographic construction and ex_net's peer-to-peer construction — plus the liquidity question underneath both. This is a discussion draft, not a recommendation; which model we want is still open.</p>

<p>
  <span class="tag">discussion draft</span>
  <span class="tag">no recommendation</span>
  <span class="tag">standalone</span>
</p>

<div class="note">
Standalone by design: it doesn't assume the rest of the site's direction. It summarizes each proposal on its own terms, then compares. Sources: TACEO, <a href="https://core.taceo.io/articles/privacy-preserving-builder-codes/">"Privacy-Preserving Builder Codes for Revenue Sharing"</a> (Mar 2026); the <a href="ex_net_whitepaper.pdf">ex_net whitepaper</a> (2017). "Armada" here = the Railgun shielded pool + Nitro clearing + ex_net-style venue described elsewhere on this site.
</div>

## 1. The TACEO proposal {#taceo}

**The problem it solves.** Wallets and apps (MetaMask, Phantom, aggregators) route enormous volume to AMMs and want a verifiable revenue share for it — builder codes / payment for order flow, the on-chain analog of Robinhood→Citadel. Proving attribution by tagging transactions publicly (e.g. ERC-8021) exposes exactly what competitors would pay to see: which users trade where, how much, and when. Attribution destroys privacy.

**The construction** — three steps between a wallet and an AMM:

1. **Tag.** The wallet holds a secret key `sk`. Per transaction it samples a nonce `x` and computes a blind referral tag `r = H(sk, x)`, attached to the tx. The AMM sees the tag but learns nothing about the wallet's identity or its other trades.
2. **Commit.** The AMM determines the fee `b` generated and inserts `[r, b]` into a public Referral Tree (a Merkle tree), publishing the root on-chain. Anyone can audit; understating fees or omitting a referral is provable via Merkle proof.
3. **Redeem.** To claim, the wallet obtains a nullifier `n = OPRF(r, sk_oprf)` from the TACEO:OPRF service, whose key `sk_oprf` is held by a network of nodes via threshold MPC (no single party holds it). It then submits a zero-knowledge proof: "I hold valid OPRF nullifiers on tags that sit in the Referral Tree summing to `$X` in fees; I know the `sk`/nonces behind them; my payout rate is `p`." The contract checks the proof and that the nullifiers are unused, then pays out privately and marks them spent.

The OPRF does the central work: it obliviously (without learning the input) and via a no-single-party key transforms a public tag into a deterministic, unlinkable nullifier, severing the link between the tag in the tree and the redemption. An optional Fee Tree (public-key → private revenue-share %) keeps differentiated pricing hidden until redemption.

**What it guarantees:** verifiable attribution with zero linkability. The wallet proves "I referred this volume" and gets paid; the user stays unlinkable; and even the fee-committing AMM cannot reverse-engineer which trades, which users, or what a competitor earns.

<p class="small"><strong>What it assumes:</strong> deep <em>public</em> AMM liquidity (Uniswap/Jupiter) — the trade is public and the taker already has full depth; only attribution needs hiding. Cost: an MPC OPRF network + ZK circuits.</p>

## 2. The ex_net proposal {#exnet}

ex_net (2017) is a federated, bonded, non-custodial cross-chain exchange venue. Validators run matching engines and take a fee percentage on every filled order; orders are backed by assets escrowed in payment channels or vouchers. It addresses the same attribution-and-privacy problem structurally, through its order and routing model, rather than with heavy cryptography.

**Order model (three types):**

- **Unblinded** — sent to the pool of bonded validators to be matched.
- **Blinded** — values encrypted to one or more specific parties; anyone can see an order exists, but only sender/receiver see the values (the receiver's wallet must be online to decrypt).
- **P2P** — matched and settled off-chain, directly between counterparties over payment channels.

**Attribution is native to routing.** Orders carry explicit filler allowlists and router allowlists; a router/filler creates a derived order backed by the original (a blinded order doubles as privacy-preserving proof-of-backing). The chain of derived orders is the attribution path — whoever routed or filled is recorded in the order structure and earns their cut from the fee schedule, settled in-channel. No public tag, no Merkle tree of fees, no nullifier, no OPRF.

**Privacy comes from P2P plus selective encryption**, not global anonymity plus an OPRF: sensitive flow is either off-chain (P2P) or per-recipient-encrypted (blinded), so "who trades where, how much, when" is never placed on a public venue. Supporting mechanisms: composable per-account orderbooks with custom matching rules, per-pair/per-account fee schedules, and counterparty blocklists; liquidity proofs (signed, selectively-blinded statements of available liquidity across chains); ephemeral liquidity (time-bounded, validator-enforced commitments that fix "ghost liquidity"); and validators opportunistically filling against public DEX activity.

<p class="small"><strong>Armada mapping:</strong> Railgun shielded pool (hides values), Nitro clearing (co-signed channel allocations already split fees <em>user / protocol / <strong>integrator</strong></em> — the integrator is the builder), and bonded <strong>watcher parties</strong> as the emergent P2P venues. Cost: public-key encryption + payment channels + access-control lists + bonded-validator honesty — no MPC/OPRF/ZK required for the base case.</p>

## 3. Comparison {#charts}

### 3a. Attribution & privacy

<table>
<tr><th>Guarantee</th><th>TACEO (crypto-heavy)</th><th>ex_net / Armada (crypto-light)</th></tr>
<tr><td>Attribution mechanism</td><td>blind tag committed by the AMM to a public Merkle tree</td><td>routing/filler allowlists + derived orders — the routing path <em>is</em> the attribution</td></tr>
<tr><td>Exactly-once (no double-claim)</td><td>OPRF nullifiers</td><td>fill→settle-once channel accounting (or Railgun nullifiers)</td></tr>
<tr><td>Verifiable fees / anti-fraud</td><td class="ok">trustless Merkle fraud proof</td><td>bonded validators + signed caches + liquidity proofs (economic + signature)</td></tr>
<tr><td>Hidden from public / competitors</td><td class="ok">yes (OPRF unlinks tag↔redeem)</td><td class="ok">yes (nothing on a public venue; P2P + encryption + shielded settle)</td></tr>
<tr><td>Hidden from the fee-committing counterparty</td><td class="ok">yes (threshold OPRF, no single de-anonymizer)</td><td class="part">no — parties in the routing/fill path see attribution (trust-scoped)</td></tr>
<tr><td>Differentiated private pricing</td><td>optional Fee Tree (ZK)</td><td>per-account fee schedules + ACLs (private because the order is P2P/encrypted)</td></tr>
<tr><td>Payout privacy</td><td>confidential payments (Merces)</td><td>off-chain channel settlement (shielded via the Railgun adapter)</td></tr>
<tr><td>Cost</td><td>MPC OPRF network + ZK circuits</td><td>PK-encryption + payment channels + allowlists + bonding</td></tr>
</table>
<p class="small">The one property ex_net/Armada does <em>not</em> match for free is <strong>counterparty-blind attribution</strong> (hiding it from the venue itself). That is what the OPRF buys. If Armada needs it, the fix is a threshold VOPRF hosted on the bonded federation (net-new crypto), or a commitment-style tag where the redemption nullifier is derived from the builder's own secret (pure Railgun-style, no threshold key) — an open design question.</p>

### 3b. Liquidity for shielded actors

The deeper axis TACEO takes for granted: once you shield the actor, what liquidity can they still reach without deshielding? Three tiers:

<table>
<tr><th>Tier</th><th>Counter-liquidity</th><th>Privacy</th><th>Depth</th></tr>
<tr><td><strong>1 · Internal cross</strong></td><td>other shielded actors (coincidence-of-wants)</td><td class="ok">highest; trade never leaves the pool</td><td class="part">thin — only opposite-side shielded flow</td></tr>
<tr><td><strong>2 · Private LP inventory</strong></td><td>a maker/LP who posts inventory and holds a <em>public aggregate</em> position (Armada's posted-price v1; LP-buffered Aave/wstETH)</td><td class="ok">taker private; LP absorbs the boundary exposure</td><td class="part">bounded by LP capital + risk appetite</td></tr>
<tr><td><strong>3 · Emergent watcher-party venues</strong></td><td><strong>small, opt-in groups coordinated over the private p2p network</strong> that provide liquidity to the shielded pool</td><td class="ok">group is the coordination/privacy boundary; identity-gating optional</td><td class="new">federated — scales by venues emerging, not one book</td></tr>
</table>

**Tier 3 is the key idea.** Liquidity provision is itself a watcher-party service: the same bonded p2p federation primitive that serves proof-carrying state to the pool can also provide liquidity to it. Such venues emerge spontaneously and permissionlessly — many coexist and compete rather than one central book (ex_net: "multiple exchanges will exist"; the marketplace's Asker/Responder model). Coordination happens within a bounded opt-in group over p2p, so flow never touches a public book. Each venue can leverage identity if it wants: an institutional/compliant pool that gates counterparties by KYC / accreditation / jurisdiction, or a fully open anonymous one. That is why [identity is an optional match filter, not a layer](architecture.html#identity): each emergent venue picks its own policy. Bridging aggregated net flow to public AMMs is one optional behavior of such a party, not the definition.

<p class="small"><strong>Consequence:</strong> depth for shielded actors is composed from many small private venues; the privacy/liquidity frontier becomes <em>federated</em> rather than a single trade-off. Limits: each venue's individual book is shallow (the bet is that many venues + composability aggregate to real depth); identity-gated venues trade openness for counterparty quality/compliance; and cross-venue discovery (finding who will privately serve your side) is itself a p2p problem, for which ex_net's liquidity proofs / advertising is the crypto-light answer.</p>

### 3c. Which problem each construction actually solves

<table>
<tr><th>Axis</th><th>TACEO</th><th>ex_net / Armada</th></tr>
<tr><td>Liquidity source</td><td>deep public AMMs (assumed given)</td><td>tiered: internal cross → private LP → emergent watcher-party venues</td></tr>
<tr><td>The hard problem</td><td>attribution privacy over public flow</td><td>liquidity depth for shielded actors</td></tr>
<tr><td>Who bears boundary exposure</td><td>n/a (trade is public)</td><td>the LP / market-maker (public aggregate), paid the spread</td></tr>
<tr><td>Venue shape</td><td>one deep public venue + crypto on top</td><td>many emergent, opt-in, identity-optional private venues</td></tr>
<tr><td>Best fit</td><td>an adversarial, transparent public AMM</td><td>a P2P, federated, non-custodial shielded venue</td></tr>
</table>

## 4. For discussion {#discussion}

Neither is strictly better; they solve different halves and assume different worlds. Questions to settle:

- **Do we need counterparty-blind attribution at all?** If our venue is a bonded, blind watcher party rather than an adversarial public AMM, the cheap ex_net/Nitro path may suffice. If we do need it, is it worth a threshold VOPRF on the federation, or can a commitment-style tag avoid it?
- **Is the liquidity model federated (many emergent watcher-party venues) or LP-anchored (a few large market-makers)?** This drives everything: depth, anonymity-set size, and how much we lean on public-AMM bridging.
- **How much identity gating?** Per-venue opt-in identity lets compliant/institutional pools and open anonymous pools coexist on the same shielded pool. Is that the desired shape?
- **Where does public depth enter?** Only via optional LP hedging / watcher-party bridging, or as a first-class routing target (which reintroduces TACEO's attribution problem)?
