# Nitro-on-Railgun — Engineering Plan (Overview)

**Status:** draft · top-level doc · **2026-09-04**
**Scope:** the whole system; this file is the index. Each workstream below expands into its own numbered step-doc (`01-*.md` … `07-*.md`), written separately.

---

## 1. Thesis

A **non-custodial, privately-settled exchange/settlement substrate** built by composing existing, battle-tested parts:

> **Notes in → normal Nitro → notes out.** A shielded pool holds value; Nitro state channels move it off-chain via signed messages; settlement mints shielded notes back. Watcher parties serve private sync + metering; shielded RFQ venues provide liquidity; a mobile wallet ties it together.

The load-bearing engineering fact: **this is almost entirely integration work.** The core construction (Design A) needs **no new cryptography** — it runs on an unmodified Railgun-style pool plus vanilla `go-nitro`. The one optional crypto upgrade (native amount-privacy-during-play) is **fork-lite: moderate, no research-grade circuits**.

---

## 2. What we're building

Our **own deployment** of the stack (not a modification of Railgun's protocol, not a rent of Railgun's pool):

- Our own **shielded pool** (redeployed Railgun OSS), so we control fees (zero them), POI policy, and features.
- The **nitro-on-railgun settlement rail** (external adjudicator: notes in, plaintext Nitro, notes out).
- **Watcher parties** serving proof-carrying, Nitro-metered sync of pool + adjudicator events.
- **Shielded RFQ venues** — private aggregating solvers; optionally DSS-federated + provably-fair (ex_net-grade).
- Our own **wallet** (+ mobile light client + self-hosted watchtower).

We monetize via **venue spread + watcher metering**, not a pool skim — which is precisely why owning the deployment beats paying Railgun's ~0.25% fee.

---

## 3. Core engineering claim — mostly integration, minimal new crypto

| Concern | Why it's not new crypto |
|---|---|
| Settlement/custody | Vanilla `go-nitro` `NitroAdjudicator` + ForceMove, unchanged |
| Shielded value | Redeployed Railgun OSS pool + circuits, unchanged |
| Adapter (notes ↔ channel) | Custom **deposit** (unshield-in) + **payout** (shield-out) endpoints only — modest Solidity |
| Games/settlement logic | Off-chain ForceMove apps — pure state machines, no circuits |
| Private sync + metering | `watcher-ts` config + `ts-nitro` vouchers, existing frameworks |
| Mobile | `@cerc-io/peer` + WebView host — existing browser stack |
| Fairness (optional) | `chain-signatures` (DSS) + commit-reveal sequencer — existing primitives |

**Only genuinely new code** = the deposit/payout adapter, the watcher config for our pool, the WASM scanner wiring, and (optional) the native channelized commitment. Everything else is deploy-and-configure.

---

## 4. Architecture (layered stack)

```
┌─────────────────────────────────────────────────────────────┐
│  CLIENT      wallet · mobile light client · self-watchtower  │  ← our fork of laconic-wallet
├─────────────────────────────────────────────────────────────┤
│  VENUE       shielded RFQ venue (aggregating solver)         │  ← internalize + flash-loan hedge
│              └ optional: DSS-federated, provably-fair (ex_net)│     chain-signatures + sequencer
├─────────────────────────────────────────────────────────────┤
│  SERVICE     watcher parties: proof-carrying feeds, metering │  ← watcher-ts + ts-nitro
├─────────────────────────────────────────────────────────────┤
│  SETTLEMENT  nitro-on-railgun adjudicator (notes in/out)     │  ← go-nitro + custom endpoints
├─────────────────────────────────────────────────────────────┤
│  POOL        our shielded-pool deployment (Railgun OSS)      │  ← redeploy; our fee=0, our POI
└─────────────────────────────────────────────────────────────┘
```

Each layer is independent above the pool: the game/venue logic is decoupled from the privacy layer, joined only at the **allocation → note settlement** seam.

---

## 5. Feature → component map

| Feature | Layer / component |
|---|---|
| Notes in / normal Nitro / notes out | Settlement (adapter + go-nitro) |
| Identity/source/destination/linkage privacy | Pool |
| Amounts leak only at boundary *(Design A)* / **hidden always** *(fork-lite)* | Pool / optional native commitment |
| POI preserved, gated entry + fresh-shield exit | Pool + adapter |
| Arbitrary off-chain games, oracle-as-signer, TLSNotary markets | Settlement (ForceMove apps) |
| Watchtower = note-scanner loop, one feed | Service + Client |
| Mobile light client over relays | Client |
| Metering private + funded at note creation | Service + Pool |
| Shielded RFQ venue, aggregating solver, flash-loan hedge | Venue |
| Address rotation, per-venue loyalty keys | Client |
| Provably-fair ordering (ex_net venue) | Venue (optional DSS + sequencer) |

---

## 6. Reuse inventory (exists — deploy/configure)

- **Railgun OSS** — pool, Groth16 verifier, Poseidon, adapters. *(license gate — see §10)*
- **`go-nitro`** (`cerc-io/go-nitro`) — `NitroAdjudicator`, ForceMove, HashLockedSwap, vouchers.
- **`watcher-ts`** (`cerc-io/watcher-ts`) — proof-carrying GraphQL framework.
- **`@cerc-io/peer`** — browser libp2p over relay; **`ts-nitro`** — voucher metering.
- **`chain-signatures`** (`cerc-io/chain-signatures`) — threshold Schnorr / DSS.
- **`laconic-wallet-web`** — wallet base (Cosmos + EIP155 accounts).
- **`mobymask-ui`** — reference browser-peer app.

## 7. Net-new inventory (build)

1. **Deposit/payout adapter** — unshield-in funding + shield-out settlement around `NitroAdjudicator`.
2. **Watcher config** — `watcher-ts` ingesting our pool commitments/nullifiers + adjudicator challenge events.
3. **WASM note-scanner + WebView↔RN key bridge** — mobile local decryption + signing.
4. **Watchtower response** — submit higher-turn state on challenge (unwired in go-nitro).
5. **Venue solver** — RFQ quoting, internalization, flash-loan hedging; optional DSS + commit-reveal sequencer.
6. **(Optional) native channelized commitment** — amount-privacy-during-play (fork-lite).

---

## 8. Workstream decomposition (→ step-docs)

| # | Workstream | Step-doc | New crypto? |
|---|---|---|---|
| 01 | **Pool deployment** — redeploy OSS, trusted setup, POI policy, fee=0 | `01-pool-deployment.md` | No |
| 02 | **Settlement rail** — adapter endpoints, go-nitro integration, watchtower | `02-settlement-rail.md` | No |
| 03 | **Watcher party** — watcher-ts config, proof feeds, Nitro metering | `03-watcher-party.md` | No |
| 04 | **Wallet & mobile** — laconic-wallet fork, WASM scanner, bridge, rotation, loyalty | `04-wallet-mobile.md` | No |
| 05 | **Venue / RFQ** — solver, hedging; optional DSS + fair-ordering | `05-venue-rfq.md` | No (opt: sequencer) |
| 06 | **Native channelized commitment** *(optional)* — amount-privacy-during-play | `06-native-commitment.md` | Fork-lite (moderate) |
| 07 | **Anonymity-set strategy** — bootstrap vs cross-pool membership proofs | `07-anonymity-set.md` | Only if cross-pool proofs |

Workstreams 01–05 are the **integration spine** and can largely proceed in parallel once 01 lands the pool address + 02 fixes the adapter interface. 06–07 are **decisions**, not prerequisites.

---

## 9. Sequencing / phases

- **Phase 0 — Gates.** Resolve prover/circuit licensing (§10) and anonymity-set strategy (07). Blocks pool deployment specifics.
- **Phase 1 — Spine (testnet).** 01 pool → 02 rail → 03 watcher → 04 wallet. Deliverable: a shielded value can enter a channel, trade off-chain, settle to a note, scanned by our wallet.
- **Phase 2 — Venues.** 05 RFQ solver on the spine. Deliverable: private yield→USDC swap end-to-end.
- **Phase 3 — Hardening / production.** go-nitro maturity gaps (multi-asset outcomes, dispute wiring, watchtower), audits of the adapter, mainnet.
- **Phase 4 — Optional upgrades.** 06 native commitment (amount-privacy-during-play); ex_net-grade provably-fair venues (05 DSS + sequencer).

---

## 10. Key decisions & open gates

1. **Prover/circuit license** *(hard gate, resolve first)* — decides whether §Pool reuses Railgun's public params or runs our own ceremony.
2. **Anonymity-set strategy** — bootstrap our own crowd (weak privacy early) vs. build cross-pool membership proofs to inherit Railgun's set (advanced; trusts their tree).
3. **Design A vs fork-lite** — accept boundary amount-leak (no new crypto) vs. native amount-privacy-during-play (moderate crypto). Not required for the feature list.
4. **RFQ vs provably-fair CLOB** — dealer market (fair-ordering moot) vs. DSS + sequencer (months, opt-in per venue).
5. **Transport** — Waku vs `@cerc-io/libp2p` gossipsub; relay-operator dependency.

---

## 11. Non-goals

- **No modification of Railgun's deployed protocol.** We redeploy the OSS; we don't fork their governance or touch their live pool.
- **No new research-grade cryptography** in the base plan. Fork-full shielded-ForceMove (amounts hidden even on dispute) is explicitly out unless separately greenlit.
- **No global consensus chain.** Fair ordering, if wanted, is a per-venue capability, not a shared L1.
- **No custody.** Every layer is non-custodial or force-closable by construction.

---

## 12. Risks

- **Anonymity set** — a fresh pool starts near zero; privacy compounds with volume. The single biggest non-engineering risk. *(→ 07)*
- **go-nitro maturity** — single-asset outcomes, dispute wiring, watchtower response, persistent-connection are the real long poles, not the Railgun bridge. *(→ 02)*
- **Mobile is unbuilt** — WebView-hosting-browser-stack is the interim; native go-libp2p/go-waku is the production path. *(→ 04)*
- **Licensing** — prover/circuit terms may force our own ceremony. *(→ 01, §10)*
- **Audit surface** — anything touching pool spend authorization or the adapter endpoints is audit-critical before mainnet. *(→ Phase 3)*

---

*Next: expand `01-pool-deployment.md` … `07-anonymity-set.md` as individual step-docs. This overview is the canonical index; update it when workstream scope or sequencing changes.*
