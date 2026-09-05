# T0 · Ethereum anchor

**Status:** draft · tier doc · 2026-09-04
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T0
**Release:** v1 (T0.0, T0.1, T0.2, T0.3, T0.7) · v2 (T0.4, T0.5) · opt (T0.6)
**Depends on / Blocks:** root tier — depends on nothing above it. Blocks T1.2 (indexes T0.0 commitments/nullifiers + T0.2/T0.3 events), T2.1 (voucher metering over T0.2), T4.1 (quote/settle app in the T0.3 app registry), T6.2 (settlement client drives T0.3), T6.3 (watchtower response with T0.2), T6.6 (mobile proving from T0.0/T0.1 artifacts). T0.4 underpins T2.4 bonding; T0.5 underpins T3 slashing.

The on-chain anchor is where value actually lives and where disputes settle. It is almost entirely **reuse**: an unmodified Railgun shielded pool (T0.0) redeployed under our control, and vanilla `go-nitro` adjudicator/ForceMove/MultiAssetHolder (T0.2). The only genuinely new v1 code is the **deposit/payout contract** (T0.3) — the Nitro↔Railgun boundary that realizes *notes in → normal Nitro → notes out* — plus our own Phase-2 ceremony (T0.1, process not code) and the anonymity-set strategy (T0.7). Two v2 economic-security contracts (T0.4 registry+bond, T0.5 sequencing-cert + fraud-proof verifier) and one optional cryptographic upgrade (T0.6 native channelized commitment) round out the tier. Amounts are public only at the T0.3 boundary — **Design A**, the default…

## T0.0 Shielded pool — redeploy Railgun OSS; POI policy; fee=0

**What it is.** This is our own shielded pool, a redeployment of the Railgun OSS contracts and circuits under our control, and it forms the root of the spine. Every downstream item references the pool address, its commitment/nullifier event ABI, and its `shield`/`unshield` entrypoints.

**Reuse vs. build.** The work here is pure reuse plus configuration; we do **not** fork the protocol.

| Component | Source |
|---|---|
| Pool / accumulator (Merkle tree, nullifier set, Groth16 verifier) | [`Railgun-Privacy/contract`](https://github.com/Railgun-Privacy/contract) |
| ZK circuits (Circom) — 54 JoinSplit circuits by UTXO in→out count | [`Railgun-Privacy/circuits-v2`](https://github.com/Railgun-Privacy/circuits-v2) |

The proof system is **Groth16** via Circom and snarkjs ([`iden3/snarkjs`](https://github.com/iden3/snarkjs)); browser and node prove through snarkjs-WASM, while mobile proves via T6.6. We deploy the **full 54-circuit set** for wallet compatibility. Trimming to a common subset is a later gas and setup optimization: each circuit dropped is one fewer Phase-2 key from T0.1, but it also costs a wallet capability.

**Config deltas from stock Railgun** (deploy/config, not protocol change):
1. **Fee = 0.** Zero the shield/unshield fee params. Revenue is venue spread plus watcher metering, not a pool skim, and this is the whole reason to own the deployment.
2. **POI policy.** Keep Railgun's Proof-of-Innocence construction but set **our** policy — gated entry plus fresh-shield exit (T0.7, §5 overview). The allow-list root and update authority are our L1 governance wallet initially, moving to on-chain governance later.
3. **Settlement-contract allow-list.** Reserve and authorize the T0.3 deposit/payout contract address in the pool config. Here we only leave the config hook; the contract itself is T0.3.
4. **Token allow-list.** Start with USDC, ETH/wstETH, and expand later.

**Interface exposed.** The pool publishes pool and verifier addresses per network; the commitment-insert and nullifier-spend event ABI (consumed by T1.2, decrypted by T6.1); `shield`/`unshield` plus the reserved deposit/payout hook that T0.3 binds to; circuit `wasm` and `zkey` artifacts and hashes for T6.6; and the POI allow-list root plus update authority.

**Testing.** Use a Laconic fixturenet for deterministic E2E (shield → commitment event → unshield → nullifier, proofs against our keys). Reuse Railgun's Hardhat suite as-is, and add tests only for our config deltas — fee=0 enforced, POI gating, and deposit/payout hook authorization.

**Risks.** Railgun OSS version drift is the main concern: pin commits and re-audit the config deltas on any bump. Anything touching pool spend authorization or POI config is audit-critical before mainnet.

## T0.1 Trusted setup — reuse Perpetual PoT Phase-1; our own Phase-2 MPC (1-of-N)

**What it is.** This is the Groth16 setup for the T0.0 circuit set. It has two phases, and only Phase 2 is ours. It is a **process** deliverable, not code.

**Reuse vs. build.**
- **Phase 1 — reuse.** We consume the community **Perpetual Powers of Tau** BN254 artifact ([`privacy-ethereum/perpetualpowersoftau`](https://github.com/privacy-ethereum/perpetualpowersoftau)). It is circuit-agnostic, large, and vetted — the expensive, risky part, and a good one already exists. We **never** run our own universal ceremony; Railgun sits in exactly this position.
- **Phase 2 — ours.** This is a circuit-specific MPC over our deployed set. Security rests on **1-of-N honest**: a single participant destroying their toxic waste suffices. We publish the full transcript plus a final public randomness beacon.

**Tooling.** The `snarkjs zkey new → contribute → beacon → verify` sequence is the primary path and remains the standard ([`iden3/snarkjs`](https://github.com/iden3/snarkjs)). Target **≥5 independent contributors** (team plus external), organizationally distinct, so the 1-honest assumption is credible. Publish the `.zkey`/`.vkey` hashes and the verification transcript in-repo, then generate the Solidity verifier(s) from the final `vkey`s and wire them into T0.0's verifier contract.

**Re-run trigger.** Phase 2 is repeated **only** when circuits change; the canonical trigger is T0.6 (native channelized commitment) or the optional T0.7 cross-pool path. Phase 1 is never re-run.

> **Reuse-Railgun-keys alternative, rejected.** Because we deploy their circuits unmodified, we *could* ship Railgun's published Phase-2 `.zkey`s and run no ceremony at all. A few days of our own MPC removes any dependence on trusting Railgun's specific ceremony instance at negligible cost, so running our own Phase-2 is the default.

**Testing.** Run `snarkjs zkey verify` in CI against the published transcript, and fail the build on any `zkey` hash mismatch.

## T0.2 Nitro adjudicator — `go-nitro` NitroAdjudicator / ForceMove / MultiAssetHolder

**What it is.** This is the on-chain settlement adjudicator and the off-chain channel machinery that move value once T0.3 has funded a channel from a note. It is vanilla `go-nitro`, and the adjudicator is unchanged.

**Reuse vs. build.** We fully reuse [`cerc-io/go-nitro`](https://github.com/cerc-io/go-nitro) (a fork of `statechannels/go-nitro`): `NitroAdjudicator.sol`, `ForceMove.sol`, `MultiAssetHolder.sol`, virtual channels ([`protocols/virtualfund/virtualfund.go`](https://github.com/cerc-io/go-nitro)), vouchers ([`payments/vouchers.go`](https://github.com/cerc-io/go-nitro)), and `examples/HashLockedSwap.sol` as a reference ForceMove app. Nothing here is protocol-new; it is `go-nitro` deploy and configure.

**Interface.** Off-chain play is plain **signed ForceMove states plus vouchers**, with virtual channels via a hub — pure state machines, no circuits. The `NitroAdjudicator` address is consumed by T0.3 (which escrows into it) and by T2.1 (whose voucher metering reuses `payments/vouchers.go`). The `ForceMove` app-registry adjudication interface is where T4.1's quote/settle app registers.

**Dispute / challenge machinery.** This is the standard ForceMove `forceMove` / `respond` / `checkpoint` on `NitroAdjudicator`, where a challenge window finalizes if left unanswered. This challenge/dispute *machinery* is owned here (T0.2). The **watchtower response** — submitting a higher-turn state within the window so a user never loses funds to a stale-state force-close — is **not** owned here; it is **T6.3**, built against this adjudicator and gated on T2 feed freshness. Reference T6.3 rather than duplicating it.

**Maturity gaps (real long poles → Phase-3 hardening; document, don't assume done):**
1. **Single-asset outcomes.** go-nitro outcomes are effectively single-asset, but **multi-asset outcomes** (ETH-in / USDC-out) are needed for swaps, which means outcome encoding plus payout handling for ≥2 assets.
2. **Dispute wiring.** Challenge/response paths exist in the adjudicator but are not fully wired in the node for our flow; they must be driven end-to-end.
3. **Persistent connections / liveness.** Reliable state and voucher delivery is required, which ties to T2.3 transport.

go-nitro is documented **pre-production**, so treat these as pre-mainnet blockers; none of them block the v1 spine demo.

## T0.3 Deposit/payout contract — the Nitro↔Railgun boundary (notes in / notes out)

**What it is.** This is the one genuinely new v1 contract: modest Solidity that binds the T0.0 pool to T0.2 channels and realizes the thesis —

> **Notes in → normal Nitro → notes out.**

```
  shielded note ──unshield──►  DEPOSIT ──fund──►  Nitro channel (NitroAdjudicator escrow)
                                                          │
                                            off-chain ForceMove states / vouchers (T0.2)
                                                          │
  fresh shielded notes ◄──shield── PAYOUT ◄──outcome── channel finalize / force-close
```

- **Deposit (unshield-in).** The user unshields from the T0.0 pool into the deposit endpoint, which opens and funds a Nitro channel via `NitroAdjudicator`. **This is the one boundary where the amount is public** (Design A). Deposit records the consumed nullifier and the funded `channelId`.
- **Off-chain play.** This is vanilla T0.2: signed states, vouchers, and virtual channels via a hub. The application logic lives in **ForceMove apps** — pure state machines such as oracle-as-signer markets, TLSNotary-fed outcomes, and HashLockedSwap.
- **Payout (shield-out).** On finalize (cooperative close *or* post-challenge), the payout endpoint reads the channel outcome and **shields it into fresh notes** — a fresh-shield exit that preserves POI. Settlement returns to **notes, not public keys**.

The Nitro game analysis is fully isolated from the Railgun privacy analysis; the two meet only at the allocation→note boundary. This adds Nitro's optimistic exits and arbitrary games to Railgun-style notes **without new cryptography** and **without governance-gated RelayAdapt** modules. RelayAdapt is Railgun's own recipe, used only inside T4.2's solver and never here.

**POI & allow-lists.** Deposit checks the T0.0 POI allow-list before funding (gated entry). Payout mints new commitments, so recipients get clean notes with no linkage to the deposit note beyond the public deposit amount (fresh-shield exit). ForceMove app bytecode and adjudication logic is allow-listed, so both the **games** and the **participants** are gated; the update authority is our L1 governance wallet, moving to on-chain governance later.

**Interface exposed (consumers bind to this, MUST NOT re-derive):**
- Contract address and ABI per network, plus the `NitroAdjudicator` (T0.2) address it uses.
- Events: `Deposit(channelId, asset, amount, consumedNullifier)`, `Payout(channelId, newCommitments[])`, plus adjudicator `Challenge`/`Checkpoint`/`Finalized` — indexed by T1.2, reacted to by T6.3.
- Channel lifecycle API: open/fund-from-note, propose/accept state, cooperative close, force-close, respond — driven by T6.2 (settlement client) and T4.2 (venue solver).
- Voucher format — reused by T2.1 for Nitro metering.
- **ForceMove app registry:** allow-listed app IDs plus the adjudication interface — where **T4.1**'s quote/settle game registers.

**Maturity note.** The go-nitro gaps in T0.2 (single-asset outcomes, dispute wiring) surface here first, since deposit/payout must encode outcomes; multi-asset is the swap-demo pull.

**Testing.** Run fixturenet E2E: note → deposit → off-chain updates → (a) cooperative close and (b) force-close plus T6.3 response → payout → fresh notes scanned. Adversarial cases that MUST pass are a stale-state force-close defeated by the watchtower (T6.3), an unresponsive counterparty where the honest party still exits correctly, and a double-fund or replayed nullifier rejected at deposit. Reuse go-nitro's Go tests, and add tests only for the deposit/payout contract.

**Risks.** Deposit/payout is spend-authorizing and therefore audit-critical before mainnet. The boundary amount-leak is inherent to Design A; hiding it in-play is T0.6.

## T0.4 Registry + bond contract — venue/party bonding *(v2)*

**What it is.** This is a **v2** economic-security contract: the federation's on-chain entry point, where watcher-party and venue membership is registered and where **disputes and slashing settle on L1**. It is net-new and design-stage.

**Role.** It holds the bond that T2.4 posts (federation membership plus threshold-DKG signing bond) and is the settlement point for slashing driven by T0.5's verifier and, in v2, by T3 ordering faults. Membership and bonding coordination proper lives in T2.4; T0.4 is only the on-L1 registry and bond escrow that those parties reference.

It is kept concise deliberately, as a v2 contract underpinning T2.4 bonding rather than v1 detail. It carries no v1 dependency, and the v1 spine ships without it.

## T0.5 Sequencing-cert + fraud-proof verifier *(v2)*

**What it is.** This is a **v2** L1 verification contract. It verifies a watcher party's **aggregate Ethereum-flavoured Schnorr (DSS)** signature ([`chain-signatures`](https://git.vdb.to/cerc-io/chain-signatures) `ethschnorr` / `ethdss`), so a bad ordering or a censored-commit receipt becomes a **slashable fraud proof against the bond** (T0.4). It is net-new and design-stage.

**Role.** Because the federation's inclusion receipts and sequencing certs are on-chain-verifiable Schnorr, T3's ordering guarantees (T3.2 DSS attestation) become enforceable: a forged or absent attestation is provable to this contract, which triggers slashing on T0.4. It underpins **T3 slashing**. The ordering *protocol* that produces the certs is T3; this is only the L1 verifier and slashing point.

It is kept concise as v2 economic-security and verification, not v1 detail.

## T0.6 Native channelized commitment — amount-privacy-in-play *(opt)*

**What it is.** This is the one optional cryptographic upgrade in the plan — **fork-lite**. It hides trade **amounts even during play and on dispute** by making channel allocations reference **hidden-amount commitments** instead of plaintext outcome balances, so amounts stay concealed on-chain even when a dispute forces state to L1. It is a **decision/design note**, Phase-4, not on the spine, and blocks nothing.

**Where the line is.** Design A (default, T0.3) leaks amounts only at the two transparent boundaries. **Fork-full** (Design B — shielded ForceMove, ZK matching plus dispute over hidden state) is research-grade and **OUT** unless separately greenlit. T0.6 is the middle point only.

**Why "lite."** It stays inside the primitives T0.0/T0.1 already ship:
- **No new proof system** — keep Groth16 via Circom and snarkjs; Halo2 is a separate, rejected decision.
- **No shielded dispute game** — `NitroAdjudicator` (T0.2) still runs its normal challenge/checkpoint/conclude; only the on-chain outcome it settles points at **commitments** rather than cleartext balances, opened by the T0.3 payout endpoint into fresh notes as today.
- **Scope of the circuit delta** — extend the payout/settlement side of the JoinSplit note format so a channel's final allocation is amount-hiding commitments (Pedersen/Poseidon values already native to the pool) with a range/consistency proof that allocations sum to the deposited value. This is a **format extension**, not a new construction; the Nitro game stays isolated, meeting T0.3 only at the allocation→note boundary.

**Trusted-setup consequence (the load-bearing coupling).** Fork-lite **changes circuits**, which means it **re-runs T0.1's Phase-2 MPC** over the amended set (Phase-1 / Perpetual PoT is never re-run): `snarkjs zkey new → contribute → beacon → verify`, ≥5 contributors, published transcript plus hashes, and regenerate plus redeploy the Solidity verifier. New `wasm` and `zkey` artifacts ship to T6.6. T0.1 already names T0.6 as the canonical "circuits changed" trigger, so nothing new is invented — the pipeline is simply re-executed.

**Buys vs. costs.** It buys amounts hidden on-chain even during disputes, and a venue that is also amount-blind for the settled allocation. It costs circuit engineering plus a fresh spend-authorization-adjacent **audit** (the dominant cost), a new Phase-2 ceremony on each format change, and heavier **mobile prover** time — the exact axis flagged by mobile-proving research ([`mobile-proving-research.md`](../mobile-proving-research.md); [zkmopro perf](https://zkmopro.org/docs/performance)) and constrained in T6.6.

**Decision: defer to Phase-4; ship Design A first.** Design A delivers the full feature set with zero new crypto or ceremony. Greenlight T0.6 only on a concrete requirement (venue, counterparty, or regulatory posture) that cannot tolerate amounts on-chain during a force-close, **and** only with budget for a circuit audit plus ceremony re-run. Fork-full stays out regardless.

**Testing (lighter — PoC gate).** On a fixturenet, force-close a committed-amount channel and show that the challenge/finalize state reveals **no cleartext amount** while payout still opens to the correct fresh notes; that a proof where allocations don't sum to the deposit is rejected (no value creation); and that `snarkjs zkey verify` against the new transcript passes in CI.

## T0.7 Anonymity-set strategy — bootstrap own crowd + Railgun import bridge

**What it is.** This is a **strategy/decision** owning exactly one lever — **crowd size** (and the timing discipline that keeps a large deposit from re-narrowing the set on exit). A shielded pool's privacy is k-anonymity against the set of unspent, indistinguishable notes; a fresh pool starts near zero and **compounds with volume**, which makes crowd size the single biggest non-engineering risk. The mechanics it rides — `shield`/`unshield`, POI, the import path — all land in T0.0.

**What the pool already gives for free** (narrowing what T0.7 must solve): amounts are hidden inside the shield, since the SNARK enforces value-conservation plus range — *$50 and $5,000,000 are equally opaque* — so there are **no Tornado-style denomination buckets**; the amount leaks only at the transparent boundary (`shield`/`unshield` and the T0.3 deposit hop). The read side (watcher, T2.0) and write side (broadcaster) are covered by the substrate, so T0.7 is purely about **populating the set**.

**Default strategy (settled).** No new crypto:
1. **Bootstrap our own crowd** — incentivized early volume, our own market-making capital, and the yield-clearing flywheel where saturating depositors convert to LPs (T4.5).
2. **Railgun onboarding bridge** — a one-click **import** in which the user **unshields from Railgun → shields into our pool**. This is **one public boundary hop** (amount and token visible only at that instant, standard for any deposit from a public balance); it costs Railgun's unshield fee once, after which the user is on our fee-0 rail with day-one dual liquidity. It is deploy/config — a scripted flow scoped in T0.0 — and T0.7 sets the policy and growth plan.

**Liquidity import ≠ anonymity-set inheritance (load-bearing).** The bridge imports **value**, not Railgun's crowd: after the hop the value sits in **our** tree, indistinguishable among **our** notes. **"Accept Railgun's" = the unshield→shield hop, NOT native note ingestion** — a Railgun note lives in *their* Merkle tree with *their* nullifier accounting and **cannot be spent/nullified in our contract** (double-spend safety). Importing a large balance does not teleport a user into a large crowd; it moves capital that must then *build* a crowd here.

**POI interaction.** The import shield is a normal entry, checked against **our POI allow-list root** (T0.0's root plus update authority); a note's Railgun history does not exempt it. The T0.3 deposit endpoint enforces the same gate before funding a channel, and payout is a fresh-shield exit, so settlement never re-links to the imported deposit beyond the one public amount. Importing does **not** import POI status.

**Bootstrap levers (discipline, not crypto).** Incentivize early depositors and LPs, and convert saturating whales to LPs to deepen the pool and grow the crowd together. Mitigate the boundary risk for large actors with **aggregation, not buckets**: trade an **aggregation window** (bigger means a larger crowd but more latency) against **LP build-up** depth — two clocks — or **dribble** value in and out incrementally. Quantization matters only at the transparent boundary hops (round or common amounts), a light convention.

**Optional advanced path — cross-pool membership proofs.** *Optional / future, NOT a prerequisite.* Our withdrawal circuits reference **Railgun's published Merkle root** so a spend proves membership in *either* tree, hiding in Railgun's larger crowd. This proves **provenance/innocence, NOT spend authority** — we still never let a Railgun note be spent or nullified in our contract; the foreign root is a membership witness for privacy only. It **trusts Railgun's root** (a new assumption) and requires **circuit work ⇒ a new T0.1 Phase-2 ceremony**. Defer it, and adopt only if organic plus bridged growth stalls and the trust plus ceremony cost is judged worthwhile.

**Metrics (instrument via T1.2/T2.0 feeds).** Track anonymity-set size (unspent indistinguishable commitments over time); **effective k** (the realistic crowd a spend hides in after correlation — track the worst-case lone large crossing separately); boundary-exposure rate (value transiting the transparent hop vs. staying shielded); bridge conversion; and LP-conversion rate.

## Sources

- go-nitro (cerc-io @435eb2b) — https://github.com/cerc-io/go-nitro — `NitroAdjudicator.sol`, `ForceMove.sol`, `MultiAssetHolder.sol`, `protocols/virtualfund/virtualfund.go`, `payments/vouchers.go`, `examples/HashLockedSwap.sol` · upstream https://github.com/statechannels/go-nitro
- Railgun OSS pool + POI construction — https://github.com/Railgun-Privacy/contract · circuits (note/commitment format) https://github.com/Railgun-Privacy/circuits-v2
- Railgun RelayAdapt / Cookbook recipe (T4.2 only, not this tier) — https://github.com/Railgun-Community/cookbook
- Railgun trusted-setup / proving system — https://docs.railgun.org/wiki/learn/privacy-system/trusted-setup-ceremony
- Perpetual Powers of Tau (Phase-1 reuse) — https://github.com/privacy-ethereum/perpetualpowersoftau
- snarkjs (Phase-2 CLI) — https://github.com/iden3/snarkjs
- chain-signatures (@9016a7c) — `ethschnorr` / `ethdss` (T0.5 verifier) — https://git.vdb.to/cerc-io/chain-signatures
- Mobile proving cost (T0.6 / T6.6 constraint) — [`mobile-proving-research.md`](../mobile-proving-research.md) · https://zkmopro.org/docs/performance
- Site docs — [`architecture.html`](../architecture.html) (Design A boundary; net-new T0 contracts) · [`build-plan.html`](../build-plan.html) (T0 build order; registry+bond, cert verifier) · [`yield-clearing.html`](../yield-clearing.html) (amount-hiding vs. buckets, aggregation window / dribble / LP flywheel) · [`glossary.html`](../glossary.html) (one anonymity set; watcher/keeper duality)
