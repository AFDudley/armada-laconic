# Shielded Nitro-Railgun Cross-Chain Swap — Construction Draft

design draft · 2026-09-10 · builds on `nitro-bridge-audit.md`

## 0. Purpose & posture

A general cross-chain swap for Armada, with an Armada shielded pool deployed on **every** chain. Value crosses as nitro-railgun notes on each side, cleared over Nitro state channels. Self-custodial; private; the hub is trusted for **liveness only, never custody**.

One framing to hold throughout: **the shielded pool is the privacy anchor, not the channel.** Value enters the bridge already shielded (sender/recipient/amount hidden in-pool), funded through fresh rotating keys, carried over the shared Armada transport (optionally under a Nym mixnet). The channel and adjudicator machinery therefore operate on already-anonymous inputs and do **not** have to reproduce Lightning-style routing privacy — no onion routing, no PTLC. (Lightning needs those because there the channel graph *is* the privacy layer; here it is not.)

## 1. Principals & roles

- **User** — holds shielded notes; wants to swap A@X for B@Y, or move an asset X→Y, privately and self-custodially.
- **Hub** — a bonded, always-on cross-chain routing party holding inventory on both chains; the only entity that touches the on-chain boundary, and only via **aggregate** rebalancing. Every hub is a **DSS construction** (threshold key). Few, fat, well-connected.
- **Maker / endpoint LP** — provides swap inventory and quotes at an endpoint, reached *through* a hub; needs one channel, no routing, no cross-chain capital, no bond. Many, permissionless. (**Hub ⊆ LP; LP ⊄ hub.**)
- **DSS committee** — the threshold-key holders behind a hub. Single legal entity (pure key hygiene) or multiple entities (Byzantine trust + bond/slash). Same crypto either way (`chain-signatures` threshold Schnorr).
- **Custodian contract** (per hub, per chain) — the DSS-controlled on-chain principal that gates the hub's collateral, verifies the group signature, and issues/rotates the hub's hot delegate key.
- **nitro-railgun adjudicator** (per chain) — the net-new core, §3.
- **Armada shielded pool** (per chain) — clean-room Railgun pool + circuits (T0.0/T0.1), independent per chain; each chain's anonymity set is its own.
- **Transport** — the **unified Armada transport** (ADR-0008: Waku pub/sub + libp2p-noise, T2.3), the same one every Armada service uses — watcher feeds, Nitro channels, DSS signing coordination, wallet. IP/metadata privacy is the **optional Nym mixnet underlay** (opt-in, uniform across services, latency cost); it is *not* a bridge-specific transport. *(Open: Nym-underlay latency vs interactive co-signing — validate for the opt-in path.)*

## 2. Privacy model — what protects what

Three independent surfaces, each with its own mechanism; do not conflate them:

| Surface | Mechanism | Established |
|---|---|---|
| Value / provenance | the shielded pool (in-circuit hiding) | before the channel exists |
| Network / metadata | unified transport (ADR-0008); **optional Nym underlay** for IP privacy | at transport |
| Counterparty graph | multi-hop virtual channels through hubs; fresh rotating endpoint keys | at routing |
| Cross-chain link | user's swap settles **off-chain**, fronted from hub inventory — the user never performs the on-chain crossing | the hub does, on its own account |

Consequence: per-swap activity is an **off-chain** channel-state update — there is **no per-user on-chain event at all**, so nothing exists to correlate. The user never performs the cross-chain crossing; the hub fronts the destination from standing inventory and squares its own books later. The hub's rebalancing is its own on-chain activity over already-anonymized inputs, unlinkable to any specific user's (off-chain) swap — it is **out of protocol** (§6) and not user-privacy-critical. The HTLC hashlock is off-chain, seen only by counterparties deliberately on both legs, so it is **not** a meaningful leak.

## 3. The nitro-railgun adjudicator (net-new core)

Distinct from **both** (i) the go-nitro L2 adjudicator — a stub, and merely a plain `NitroAdjudicator` on a second chain — and (ii) work-package A's **T0.3** glue, which rides the *vanilla ECDSA* adjudicator with cleartext outcomes. We deploy a purpose-built adjudicator on **each** Armada chain. It is ForceMove's dispute machine plus three changes the vanilla contract cannot express:

**(a) Note-native custody — the Railgun↔Nitro boundary.** Funded by an **unshield** from the local pool (a note's value escrowed into the adjudicator); settled by **shielding** fresh notes on conclude. "Deposit" = unshield-in; "payout / exit" = shield-out. This folds A's T0.3 role *into* the adjudicator; it is the spend-authorizing, audit-critical boundary.

**(b) Contract-signature participants (EIP-1271) — resolving the DSS/ECDSA clash.** The audit confirmed ForceMove verifies state sigs by hard `ecrecover` against `fixedPart.participants` (EOAs), with **no pluggable verifier**, while the DSS is threshold **Schnorr** — so a hub cannot be an EOA participant signed by the group key. Fix:
- The channel **participant is the custodian contract address** — stable, so it satisfies ForceMove's fixed-participant immutability.
- The adjudicator verifies a custodian participant's signature via **EIP-1271 `isValidSignature`**, not `ecrecover`. (On-chain verification only ever happens on challenge/checkpoint/conclude, so the extra call cost is dispute-only.)
- The custodian's `isValidSignature` accepts the hub's **current hot delegate EOA** signature and can **rotate that delegate without changing the channel's `FixedPart`**. Cold DSS/group key gates funds, rotates the delegate, and slashes; the hot delegate does fast per-state signing.

This is exactly the "group delegate credentials + bind custodian contract" the laconic nitro integration *declared but never built*. It also dissolves the key-rotation-vs-channel-immutability tension: the channel names the stable custodian, and rotation happens inside it.

**(c) Unilateral exit → re-shield, on every chain.** Any party can force-close at the latest supported state (challenge/checkpoint/conclude) and have its allocation **shielded back to its Railgun address**. Because the adjudicator exists on *both* chains, **each leg has real enforcement** — this is what makes the HTLC genuinely trustless-atomic and reduces hub trust to liveness. This is precisely the gap the L2 stub left open; we fill it by **building the adjudicator, not finishing the stub**.

**(d) Amount privacy on exit — phased.** v1: outcomes are cleartext on a forced conclusion (Design A; a *contested* swap leaks its size on-chain). v2: outcome allocations carry **hidden-amount commitments** (T0.6 "fork-lite") so even a forced exit reveals no amount — requires the circuit change + a fresh ceremony (A.3/A.9); deferred.

**Reused inside it:** ForceMove challenge/checkpoint/conclude logic, exit-format outcomes, MultiAssetHolder-style holdings + `expectedHeld` guard, the pool's shield/unshield entrypoints, and the `chain-signatures` Schnorr verifier (`SchnorrSECP256K1.sol`) *inside* the custodian.

## 4. Swap flow — happy path

**Fronting is the core mechanism, not an add-on.** A hub with inventory on both chains gives the user the destination asset *now* out of its standing Y-inventory; it cannot teleport the user's specific X-value across. The user's input simply joins the hub's X-inventory, and the hub nets the resulting cross-chain imbalance later, in aggregate. That instant advance is what makes the swap usable, is where the hub earns its spread, and — because the destination funds come from a common pool rather than the user's own value crossing — is what severs the per-user cross-chain link (§2).

1. **Fund (chain X).** User unshields a note into the nitro-railgun adjudicator on X, opening a ledger channel to a hub under a fresh key, over the shared Armada transport (Nym underlay optional). Amortized — one channel backs many swaps.
2. **Route + quote.** A virtual channel reaches the maker (or hub inventory) through the hub graph; a price is quoted (same-asset moves are ~1:1 minus the fronting spread).
3. **Fronted swap (off-chain, atomic).** The hub **fronts** the destination side to the user from its standing Y-inventory in one round-trip, bound by an HTLC so the advance is atomic: the hub is guaranteed the user's X-side the instant the user receives Y. No value crosses chains here — only the hub's inventory shifts. Settlement stays off-chain, no per-swap chain event.
4. **Receive (chain Y).** The user holds Y immediately as a channel allocation and can withdraw it as a fresh shielded note on Y whenever it wants (shield-out via the Y adjudicator).

## 5. Unhappy path — why the adjudicator must be on every chain

- **Timeout / abort.** Standard HTLC timeouts. If the Y-leg never completes, the user reclaims the X-leg at the latest supported state and re-shields — a guaranteed refund.
- **Hub offline / malicious.** User force-exits via challenge/checkpoint on the chain where its collateral sits, settling to a note. The hub can stall progress; it cannot take funds.
- **Enforcement location.** Because the nitro-railgun adjudicator is on *both* chains, each leg is enforceable on its own chain — the HTLC is not "atomic only where an adjudicator happens to exist" (the audited bridge's fatal gap, where L2 had none and exit depended on a live, honest operator).
- **Watchtower.** Each party runs — or delegates to a keeper — a self-hosted watchtower (T6.3) to checkpoint a stale close inside the challenge window. Self-run; no operator to trust.

## 6. Hub economics & rebalancing — out of protocol

Rebalancing, inventory sizing, and pricing are **operator business logic — not protocol, and never a wallet concern.** The protocol neither specifies nor requires them; it only has to *not prevent* a hub from managing its own inventory. A hub squares the net cross-chain imbalance that fronting creates however it likes — the same public primitives, CCTP, a CEX, OTC — on its own schedule and at its own risk. We may ship a **reference hub daemon** as open-source convenience, but it sits outside the protocol boundary.

Two facts worth stating even though they're out of protocol, because they bound viability:
- Fronting makes the hub's risk **capital / inventory / price + reconciliation latency**, not custody of user funds (HTLC atomicity + unilateral exit cover custody). Inventory is the hard liquidity ceiling: run dry and the hub widens or withdraws quotes (the ADR-0011 static-inventory / T4.5 "saturation → dribble / become-LP" behavior).
- Bonding/slashing (T0.4/T2.4) gives the multi-entity DSS economic teeth; single-entity hubs get key-security only.

## 7. Trust summary

| Party | Trusted for | NOT trusted for | Enforced by |
|---|---|---|---|
| Hub | liveness (routing / quotes) | custody of funds | unilateral exit + adjudicator on both chains |
| DSS committee | producing the group signature | unilateral theft — needs `t`-of-`n`; no single key | threshold Schnorr + custodian; bond/slash if multi-entity |
| Maker | honoring its quote in-channel | custody | HTLC atomicity + exit |
| Pool / circuits | correctness | — | clean-room audit |

## 8. Reuse vs net-new (grounded)

- **Reuse:** ForceMove dispute logic, exit-format, MultiAssetHolder patterns (go-nitro `@435eb2b`); Railgun pool/circuits design (clean-room, T0.0/T0.1); `chain-signatures` threshold-Schnorr crypto + `SchnorrSECP256K1.sol`; ts-nitro client; the unified Waku + libp2p transport (ADR-0008, optional Nym underlay).
- **Net-new (must build):** the nitro-railgun adjudicator (§3 a–c); the custodian contract + delegate issuance/rotation/slash; the DSS↔adjudicator wiring (unbuilt per audit); unilateral-exit-to-reshield; (v2) hidden-amount outcomes (T0.6) + ceremony. *(Hub rebalancing/inventory/pricing are out of protocol — §6 — at most a reference daemon, not a protocol deliverable.)*
- **DSS scope — EVM + Nitro only, no laconicd:** we reuse only `chain-signatures`' threshold-Schnorr crypto and the kyber DKG/signing *logic*; we do **not** use laconicd's CometBFT vote-extension transport — signing coordination rides the unified Armada transport (ADR-0008). Net-new is the EVM custodian's group-sig verification + delegate rotation and hosting the DKG/signing over that transport. Bonding/slashing is absent; both DSS codebases are self-declared **unaudited**.

## 9. Deferred / out of scope for the first cut

- **PTLC / adaptor sigs** — unnecessary; the pool is the privacy anchor. Revisit only if the channel graph ever carries identity.
- Nym-underlay latency validation (the opt-in IP-privacy path); denomination/anonymity-set policy; quote/RFQ privacy; MEV/ordering (T3); cross-chain POI gated-entry.

## 10. The one on-chain leak that survives v1

A *contested* swap forced to conclusion reveals its outcome amount on-chain (Design A). Happy-path swaps leak nothing (off-chain). Closing this fully is the T0.6 hidden-amount-outcome work in §3(d) — deferred, and the only reason to pull it forward is if adversarial force-closes of large swaps become a real deanonymization vector.

## 11. Comparison to existing bridges

Almost every popular bridge is **transparent by construction** — each leg is a public on-chain transaction, so source address, destination address, asset, amount, and timing are public on both chains, and the two legs are explicitly linkable (shared message/nonce, or a trivial amount+time match). Bridges are, in practice, a primary cross-chain *deanonymization* surface. This holds for lock-and-mint (canonical rollup bridges, Polygon PoS), fast liquidity bridges (Hop, Across, Stargate, Synapse, Celer), message bridges (Wormhole, LayerZero, Axelar, CCIP), and CCTP. Even privacy chains that cross chains (Namada MASP+IBC, Penumbra shielded IBC, Zcash-via-a-bridge) hide sender/receiver *within* their own shielded set but **expose the amount at the transparent boundary** and do **not decouple the two legs**. HTLC atomic swaps are non-custodial but fully public (addresses, amounts, shared hashlock).

### Privacy

| System | Sender | Receiver | Amount | Legs unlinkable | IP/metadata | Custody trust |
|---|---|---|---|---|---|---|
| Canonical / lock-mint | public | public | public | no (explicit) | no | chain / some custody |
| Fast liquidity (Hop/Across/Stargate) | public | public | public | no | no | bonder + msg layer |
| Message bridges (Wormhole/LZ/Axelar) | public | public | public | no | no | external validator set |
| CCTP | public | public | public | no | no | Circle (can freeze) |
| Privacy chains x-chain (Namada/Penumbra) | hidden in-set | hidden in-set | public at boundary | no | no | IBC/bridge |
| HTLC atomic swap | public | public | public | no (shared hash) | no | none |
| **This system** | **hidden** | **hidden** | **hidden** (except amortized channel-open) | **yes** (off-chain, fronted) | **yes** (optional Nym) | **none** (unilateral exit + threshold DSS) |

The categorical difference is the last two columns: because per-swap settlement is **off-chain**, there is *no per-user on-chain event to correlate* — the matched pair every other bridge exposes does not exist here — and fronting means the user never performs the on-chain crossing at all.

### Performance

Mechanically this **is** a fast-liquidity bridge (it fronts), so it sits in the Across/Hop/Stargate latency class, not the slow canonical class.

| System | User-perceived latency | Steady-state cost / swap | Non-custodial |
|---|---|---|---|
| Canonical optimistic (rollup native) | withdrawal ~7 days (deposit min) | gas both sides | trust chain |
| CCTP (raw) | ~13–19 min finality + mint | low | Circle |
| Message bridges | minutes | gas + fee | validator set |
| Fast liquidity (Across/Hop/Stargate) | secs–mins (bonder fronts) | gas both sides + ~0.05–0.3% LP fee | bonder fronting |
| **This system** | **sub-second–seconds steady-state** (off-chain HTLC after channel open) | **~0 on-chain gas per swap** + hub spread | **yes** |

Honest nuances: the first swap includes a channel open — an on-chain unshield + an **on-device Groth16 proof** (seconds, more on a phone, T6.6) + a confirmation — amortized over many off-chain swaps. The optional Nym underlay adds mixnet latency (off by default). Capital model matches every fast bridge (hub needs destination inventory; run-dry widens quotes), plus a Lightning-style channel-liquidity lock while a channel is open.

### Where this is narrower / not yet proven

- **Scope:** it moves value between chains that *both* run an Armada shielded pool — private multichain value movement inside the Armada deployment, not a universal token↔chain bridge.
- **Maturity:** a designed, post-v1 construction (ADR-0015/0016) with unbuilt, unaudited dependencies. The compared bridges are live — though for many, "battle-tested" means tested by the field's largest hacks (Ronin $624M, Wormhole $320M, BNB bridge $570M, Nomad $190M, Multichain collapse), almost all **custody / validator-set** failures this design structurally avoids.
- **Residual leak:** the amortized channel-open/close amount is public (Design A), and a *contested* force-close reveals a swap's size (closable with T0.6, §3d). Transparent bridges leak the amount on *every* transfer, always.

**Bottom line.** Privacy: a different class — popular bridges deanonymize; this keeps sender, receiver, amount, and the cross-chain link private. Security: stronger than the trusted-validator bridges responsible for the field's largest losses (non-custodial, unilateral exit, threshold key). Performance: on par with fast bonder bridges, cheaper per-swap at volume, traded against amortized channel-setup + on-device proving and optional Nym latency.
