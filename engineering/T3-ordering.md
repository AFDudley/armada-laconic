# T3 · Ordering

**Status:** draft · tier doc · 2026-09-04
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T3
**Release:** v2 (T3.0, T3.1, T3.2)
**Depends on / Blocks:** consumes T2.4 (threshold DKG signing), T0.4 (venue bond), T0.5 (sequencing-cert + fraud-proof verifier); depends on T3.2 for the receipts T3.0 emits and T3.1 attests; **blocks T4.6** (ex_net matcher + LP vault) — this whole tier exists only to feed it.

T3 is the **fair-ordering stack**, and it is **v2 in its entirety**. It becomes necessary *only* when a venue upgrades to the ex_net matcher (**T4.6**), where price *discovery* runs across multiple LPs with competitive quotes and a book therefore exists to front-run. The v1 posted-price venue (T4.0) has **nothing to front-run**: Armada posts bid/ask on both sides, take-it-or-leave-it, cleared over Nitro, so v1 needs neither commit-reveal ordering nor epoch set-agreement (ADR-0007; architecture T4 note). These are the two highest-design-risk, research-adjacent items in the whole system, and confining them here is precisely what lets v1 collapse to integration plus a small posted-price contract plus the Design-A boundary.

The tier's job is narrow: **fix which orders are in an epoch and fix their order**, then prove it. It does **not** run global execution, because balances are enforced by the Nitro channel adjudicator (T0.2) rather than by a shared state root. In Ethereum terms, this is a **decentralized sequencer + DA + fraud-proof ordering** *over state-channel settlement* — **not a rollup** (architecture T3). A stalled or sub-threshold party therefore means **halt, not loss**: users force-close to their last signed state, and a watchtower (T6.3) guarantees the latest state even for an offline user.

Fairness rests on **commit-reveal** (T3.0) rather than a per-order client SNARK, because a phone SNARK per order costs seconds while a reveal round-trip costs tens of milliseconds; it also avoids a threshold-encrypted mempool, since a colluding threshold that is also the sequencer gains no real fairness from encryption (architecture T2 note). The three items compose as a pipeline: **T3.0** hides and orders, **T3.1** agrees the set and publishes it, and **T3.2** signs each step so a violation is slashable on L1 (the T0.5 verifier against the T0.4 bond). Everything below is **months of work per venue, opt-in**, and never a v1 prerequisite.

## T3.0 Commit-reveal sequencer + randomness beacon

**What it is.** This is the core fair-ordering primitive: orders stay hidden from the venue *during ordering* through **commit-reveal**, not client-side ZK. A phone generating a SNARK per order costs seconds, whereas a commit → reveal round-trip costs tens of milliseconds, and the user is already interactive through settlement anyway. Commit-reveal defeats **both** latency games and **operator front-running** without imposing per-order proving cost on the client.

**Flow (one epoch).**
1. **Commit.** Each order enters as `commit = H(order ‖ salt)` — a hiding, binding hash. The venue sees only hashes, so it cannot content-front-run.
2. **Seal.** The party seals the epoch, admitting no further commits. On seal it emits **threshold-signed inclusion receipts** (T3.2) for every commit it accepted, so *censorship of a commit becomes slashable* (T0.4/T0.5).
3. **Beacon.** A **post-seal randomness beacon** fixes intra-epoch positions. Because the beacon is drawn *after* the seal, the party cannot position-manipulate: it committed to the set before knowing the order.
4. **Reveal.** Users reveal `order ‖ salt`; the matcher (T4.6) runs over the revealed orders **in the beacon-fixed order** and clears the epoch as a **batch auction at a uniform price** (Budish–Cramton–Shim frequent-batch-auction), which removes the intra-epoch latency race entirely (architecture T3).

The party therefore **cannot content-front-run**, because it only saw hashes, and **cannot position-manipulate**, because the beacon is post-seal. ZK order submission remains an **opt-in** for offline submitters who cannot stay online for the reveal leg; it is a fallback, not the default path.

**Latency & receipts.** A cooperative fill takes **3 RTT + Δ** — commit, reveal, settle — with Δ ≈ the party's co-located sealing window (single-digit ms). Settlement is one round trip whose two legs are **both mandatory**: the user signs first, and the venue returns a **bond-enforced settlement receipt** (T3.2). Nothing is fire-and-forget; every fill terminates in either a co-signed receipt **or** a slashable proof of its absence.

**Griefing is a non-issue.** Because the venue is the one *providing asks*, a user who withholds a reveal harms no third party: the ask is a standing quote available to all, so a non-revealed commit blocks no other user's liquidity. The only theoretical residual is a sub-second free option on a stale quote, worth ≈ nothing as long as the reveal timeout ≤ the quote-repricing interval — a knob the venue controls. Anti-griefing thus collapses to a nominal anti-spam commit fee.

**Reuse vs build.** Net-new. The commit-reveal mempool and the post-seal beacon have no upstream to reuse; the beacon can be sourced from the sealed-set hash plus a delay/VDF-style draw, but the sequencing protocol itself is new work.

**Key tasks.**
- Define the commit encoding `H(order ‖ salt)`, salt entropy, and the nominal anti-spam commit fee.
- Implement the epoch clock + **seal barrier** (co-located sealing window Δ, single-digit ms).
- Source the **post-seal randomness beacon** (sealed-set hash + delay/VDF-style draw); show it is unbiasable before seal.
- Wire the reveal collector and hand the **beacon-fixed order** to the matcher (T4.6).
- Implement the two-leg settlement exchange (user-signs-first → bond-enforced receipt) and the force-close fallback (T0.2 / T6.3).
- Tune the reveal timeout ≤ quote-repricing interval to bound the stale-quote free option.

**Interface.** Consumes the epoch set membership and DA guarantee from **T3.1**, so reveals are checked against the sealed set, and the inclusion/settlement receipts from **T3.2**. Emits the **beacon-fixed revealed order** to the matcher (**→ T4.6**); a missing settlement leg force-closes to the pre-fill channel state on T0.2, made whole by the watchtower (T6.3).

**Testing / risk.** The adversarial cases are a party that seals *after* peeking (defeated by pre-seal receipts + post-seal beacon), a withheld reveal (bounded free option, above), and a censored commit (a receipt-backed fraud proof, T3.2). This is the highest-design-risk area of the system after T3.1.

## T3.1 Epoch set-agreement / DA

**What it is.** This is the agreement layer under the sequencer: the federation must (a) agree on **which** commitments are in an epoch and (b) publish that revealed epoch so anyone can verify the clear. It is the **rollup-flavored set-agreement** — leader-proposes + threshold-attest + receipt-backed censorship fraud proofs — but deliberately **not BFT consensus**. The party only needs to fix a *set* and an *order*, a sequencing-and-attestation task rather than replicated global execution; that is a far weaker and cheaper primitive than BFT, and it is what lets the network live on phones (architecture T3 note). It is the **highest-risk, research-adjacent** item in the plan (build-plan T3).

**Set-agreement.** A leader proposes the sealed commit set for the epoch, and the federation **threshold-attests** it (T3.2). A member that observes its own accepted commit **excluded** from the attested set holds an inclusion receipt (T3.2) that contradicts the sealed set — a **censorship fraud proof** slashable on L1 (the T0.5 verifier against the T0.4 bond). Safety is favored over liveness: a stalled or sub-threshold party **halts**, and users force-close to their last signed state (non-custody, architecture T3).

**Data availability.** Set-agreement is worthless if the revealed epoch is not **available**, since a censored reveal set would let the party equivocate about what was cleared. The revealed epoch is therefore **published (DA)** after reveal: the ordered, revealed order set backing each uniform-price clear is made retrievable, so any watcher can independently re-run the batch auction and check the clear against the beacon-fixed order from T3.0. Without this DA step there is no external check on the matcher's output and no ground truth for a fraud proof. DA publish is **net-new** (build-plan T3); its substrate is the same watcher-party feeds (T2) rather than a separate DA chain.

**Reuse vs build.** Net-new; both the leader/attest/fraud-proof protocol and the DA publish path are new. It reuses the threshold-signing primitive (T3.2 / T2.4) for the attestations, but the sequencing protocol around it is new.

**Key tasks.**
- Leader election / rotation for epoch proposal (deliberately non-BFT).
- The sealed-set proposal + **threshold-attest** exchange (signed via T3.2).
- Exclusion detection: match member inclusion receipts against the attested set; assemble the **censorship fraud proof** for T0.5.
- **DA publish** path for the revealed epoch over T2 feeds; retrievability + an independent re-clear check.
- Halt semantics: sub-threshold ⇒ stop and surface the force-close signal to clients (no partial epoch).

**Interface.** Consumes commits + reveals from **T3.0** and threshold attestations from **T3.2**. Produces the **sealed epoch set** (to T3.0, gating reveals) and the **published revealed epoch** (to any watcher, and to T4.6 as the auditable input to each clear). A disagreement is resolved on L1 via T0.5 against T0.4.

**Testing / risk.** Simulate a leader that omits a commit (caught by the excluded member's receipt), a party that stalls below threshold (must halt, users made whole), and a withheld reveal set (the DA gap must be detectable, not silently absorbed). Because there is no shared state root, tests assert *ordering + availability*, not global-state equivalence.

## T3.2 DSS attestation

**What it is.** This is the signature layer that makes the tier's claims **on-chain verifiable and slashable**. The federation signs everything it attests to — **sequencing certificates**, per-commit **inclusion receipts**, and **liquidity-proof snapshots** — with a **threshold Schnorr** signature. Concretely, this means `chain-signatures` Ethereum-compatible Schnorr (`ethschnorr.Sign` / `Verify`) and the Stinson–Strobl `(t, n)` **Distributed Schnorr Signature** protocol in `ethdss`, built over a `kyber` DKG (architecture "federation signature"). This is the mechanism that turns a **censored commit** into a **slashable fraud proof**: the missing or contradicted receipt is verified by an L1 contract (T0.5) against the venue bond (T0.4).

**Why it works on-chain.** Because it is **Ethereum-flavoured Schnorr**, an L1 contract can verify the party's *aggregate* signature directly, so a sequencing cert or a censored-commit inclusion receipt becomes a slashable fraud proof against the bond with no re-execution needed (architecture DSS §). The threshold direction is **safety-first**: `t`-of-`n` tolerates `t−1` malicious for safety and `n−t` offline for liveness. We pick **t high** (e.g. **4-of-7**) so that forging an attestation needs a large coalition, and liveness failure degrades to **halt, not loss** because settlement is non-custodial. A BFT-style small quorum (e.g. 3-of-11) would be *wrong*, since it would let any 3 forge (architecture DSS §).

**Signing only — no encrypted mempool.** The threshold key is used for **signing only**. We deliberately do **not** run a threshold-encrypted mempool: with the sequencer and the key-holders being the same federation, encryption gives no real fairness, because a colluding threshold can decrypt-then-order. Fair ordering comes from commit-reveal (T3.0), not encryption (architecture T2 note).

**Reuse vs build.** The **signing primitive is built** — `ethschnorr` / `ethdss` in `chain-signatures` — and the **DKG / reshare / signing wiring is T2.4** (federation + bond + threshold DKG). What is net-new *here* is the **attestation protocol**: which objects get signed (sequencing cert, inclusion receipt, liquidity-proof snapshot), their wire encoding, and the receipt-emission points inside the T3.0/T3.1 flow. The **L1 verifier is T0.5** and the **bond it slashes is T0.4** — both referenced, not built here.

**Key tasks.**
- Define the signed-object schemas — sequencing certificate, inclusion receipt, liquidity-proof snapshot — and their wire encoding.
- Bind `ethschnorr` / `ethdss` signing (from the T2.4 DKG) to the seal/attest emission points in T3.0 / T3.1.
- Produce an aggregate-signature format the **T0.5** L1 verifier can check, gas-bounded.
- Index receipts so a client/watcher can assemble a fraud proof and submit it against the **T0.4** bond.
- Choose `(t, n)` parameters (t high, e.g. 4-of-7) jointly with T2.4.

**Interface.** Consumes the threshold key material and signing service from **T2.4**. Emits **inclusion receipts** at seal-time to **T3.0**, **sequencing certificates** over the attested set to **T3.1**, and both — plus liquidity-proof snapshots — to the **T0.5** verifier, so a bad or missing attestation slashes **T0.4**. The matcher (**T4.6**) consumes liquidity-proof snapshots as its auditable inventory basis.

**Testing / risk.** Verify that (1) a valid `(t,n)` aggregate verifies on the L1 verifier (T0.5) while an under-threshold or forged one does not; (2) an inclusion receipt contradicting a sealed set (T3.1) is accepted as a fraud proof and slashes the bond (T0.4); (3) a `t`-high quorum halts rather than forges under `≥ n−t` offline. The DSS/sequencer upgrade adds a **larger shared-substrate audit surface** than a single posted-price contract, so it is audit-critical before any mainnet ex_net venue (architecture audit note).

## Sources

- Architecture — T3 ordering (commit-reveal + post-seal beacon, latency/receipts, halt-not-loss, "sequencer + DA + fraud-proof, not a rollup"); the federation signature (threshold Schnorr / DSS, on-chain verifiable, safety-first `t`-high, signing-only / no encrypted mempool); T4 v1-vs-v2 posted-price note — `architecture.html`
- Build plan — T3 net-new breakdown (commit-reveal mempool + beacon, **epoch set-agreement** = highest risk, inclusion receipts + sequencing certificate, DA publish, watchtower); v2 = ex_net/matcher scope — `build-plan.html`
- Venue / RFQ — the optional provably-fair CLOB path (commit-reveal sequencer, DSS attestation, inclusion receipts, liquidity-proof snapshots, L1 slashing), months-per-venue / opt-in framing — see T4.6 and `execution-platform.html`
- chain-signatures (threshold Schnorr `ethschnorr`, Stinson–Strobl `(t,n)` DSS `ethdss`, `kyber` DKG) — https://git.vdb.to/cerc-io/chain-signatures
- Registry / bond and cert + fraud-proof verifier referenced by id — T0.4, T0.5; threshold DKG signing — T2.4; matcher consumer — T4.6
