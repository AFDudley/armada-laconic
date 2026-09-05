# 2. Constraints

arc42 §2 · 2026-09-04

These are the fixed rules the architecture must obey, not decisions to revisit here. Most
are anchored by an ADR (→ §9) and realized by a `T#.#` item (→ §5); anything without a
governing ADR is a hard external or organizational fact. Each entry is stated so that it
can be checked, following the pattern of constraint, consequence, and governing ADR or
item.

## 2.1 Technical constraints

**TC-1 — Ethereum L1-anchored.** All value custody, adjudication, and the settlement
boundary live in on-chain contracts (T0); custody never depends on the trust of an L2 or
sidechain.
- *Consequence:* Gas cost and L1 confirmation latency bound every deposit/payout and
  force-close, and watchers read L1 state directly rather than a trusted RPC of record.
- *Governs:* §5 T0.

**TC-2 — Railgun's deployed protocol is consumed unmodified.** We redeploy the Railgun
OSS contracts and circuits under our own control (T0.0) and never edit or touch Railgun's
live pool.
- *Consequence:* Shielded-pool cryptography is inherited rather than re-invented. Upgrades
  and audit become our responsibility, but note format and nullifier semantics stay
  Railgun-compatible, and value from Railgun enters only via the unshield→shield bridge
  hop (T0.7).
- *Governs:* ADR-0002 · §5 T0.0, T0.7.

**TC-3 — Groth16 over BN254 proving.** v1 keeps Groth16: Phase-1 reuses the community
Perpetual Powers of Tau, and Phase-2 is our own 1-of-N MPC (T0.1).
- *Consequence:* Each circuit requires a per-circuit trusted setup, so any circuit change
  such as fork-lite (T0.6) re-runs Phase-2 and a fresh audit. In return, on-chain
  verification and mobile prover performance stay Groth16-cheap.
- *Governs:* ADR-0003 · §5 T0.1, T0.6.

**TC-4 — go-nitro is pre-production.** Settlement rides `go-nitro` ForceMove and
MultiAssetHolder (T0.2) through the deposit/payout contract (T0.3).
- *Consequence:* Multi-asset outcomes, dispute wiring, and watchtower response are the
  long poles (→ §11). We build no new state-channel cryptography, only the
  note↔allocation boundary.
- *Governs:* ADR-0004 · §5 T0.2, T0.3.

**TC-5 — Mobile = React Native + device secure enclave.** Wallet custody rests on
BIP-39/HD keys held in the platform secure enclave (T6.0), and js-libp2p is not
RN-compatible.
- *Consequence:* A native gomobile transport (go-waku, go-libp2p-noise) must be built,
  with an interim WebView browser-stack until it lands (T6.5). Keys never leave the
  enclave, and the WASM note-scanner talks to it over a key bridge (T6.1).
- *Governs:* ADR-0008 · §5 T6.0, T6.1, T6.5.

**TC-6 — EVM gas / verifier limits.** Every on-chain verifier and settlement path must
fit within L1 gas and pairing-check budgets.
- *Consequence:* This favors Groth16's constant-size proof and cheap verify, reinforcing
  TC-3, and rules out on-chain designs whose per-op cost scales with participant count.
- *Governs:* ADR-0003 · §5 T0.

**TC-7 — On-device proving memory wall.** Groth16 mobile proving (snarkjs-WASM /
rapidsnark, T6.6) runs inside phone RAM limits, using artifacts from T0.0/T0.1.
- *Consequence:* Circuit size is capped by what a phone can prove, so a heavier circuit
  such as fork-lite raises prover cost and is deferred, which ties back to TC-3.
- *Governs:* ADR-0005 · §5 T6.6.

## 2.2 Organizational & scope constraints

**OC-1 — Non-custodial at every tier.** No tier ever holds user funds in a way that lets
it move them unilaterally; custody is contract-enforced (T0).
- *Consequence:* Watchers, the venue solver (T4.2), and Adapters (T5) operate only on
  user-signed channel state or transparent-boundary hops, never on discretionary custody.
- *Governs:* §5 T0 · quality scenarios in →§10.

**OC-2 — Railgun's deployed protocol is untouched; we redeploy the OSS.** We stand up our own pool by redeploying the Railgun OSS under our control (T0.0, fee=0, our POI) and never fork, patch, or govern Railgun's *deployed* protocol or live pool ([ADR-0002](./09-architecture-decisions.md#adr-0002)).
- *Consequence:* Blast radius splits cleanly — pool and audit changes never route through the product work above T0.0. This reinforces TC-2.
- *Governs:* ADR-0002 · §5 T0.0.

**OC-3 — One team; a reuse boundary, not an org split.** A single engineering team builds the whole Armada product; the boundary is **reuse vs net-new**, not Armada-vs-Laconic ([ADR-0013](./09-architecture-decisions.md#adr-0013)). Reuse Railgun OSS + Laconic prior art (nitro, watchers, wallet, chain-signatures, ingestion); build only the deltas. The adapters (T5.0 — CCTP built, Aave-v4 yield, Swaps) are in scope, owned by work package C.
- *Consequence:* Interfaces between work packages — the T0.3 deposit/payout ICD and the T5.1 routing interface — are specified before parallel work begins (→ §0, §3, §4). "Adapters" (capital-A) means the T5 tier (ADR-0010).
- *Governs:* ADR-0013 · §0 · §5 T5.

**OC-4 — Internal source docs require Laconic/Vulcanize access.** The deep design Google
Docs (architecture, build-plan, mobile-privacy, execution-platform, yield-clearing,
glossary) are internal to Laconic/Vulcanize.
- *Consequence:* The engineering docs in this folder must be self-sufficient for anyone
  without that access. Cite the pinned public URLs already in the T-docs, and never gate a
  claim behind an internal-only doc.
- *Governs:* — (organizational fact).

## 2.3 Convention constraints

**CC-1 — arc42 is the one documentation template.** There is one item model, the §5
registry; tiers and "lines" are views and tags, not competing hierarchies.
- *Consequence:* All engineering docs conform to the 12 arc42 sections. Building-block
  detail stays in the T-docs, decisions are ADRs, and cross-cutting concerns live in §8.
- *Governs:* ADR-0001.

**CC-2 — Terminology is fixed.** Use **deposit/payout contract** (not "adapter") for
T0.3; reserve **RelayAdapt / Adapt / Cookbook** for Railgun's own recipe and **Adapters**
(capital-A) for Armada's T5; use **boundary** not "seam"; **tier** (T0–T6) not "layer";
and **venue / clearing** not "marketplace."
- *Consequence:* Docs stay grep-clean for the banned terms, and collisions with
  Railgun/Ethereum/Feathers vocabulary are avoided.
- *Governs:* ADR-0010.

**CC-3 — Reference by id.** Cite building blocks by `T#.#` and decisions by `ADR-####`,
and cross-reference sibling sections by number. Never renumber registry ids or edit a
decided ADR; supersede it instead.
- *Consequence:* The §5 registry stays the single source of truth for ids, status, and
  release, and the ADR log stays append-only.
- *Governs:* ADR-0001, ADR-0010.

---

*Constraints trace into strategy in →§4 and quality scenarios in →§10, and the maturity
risks behind TC-4/TC-5 are tracked in →§11. See §5 for every `T#.#` item and §9 for the
governing ADRs.*
