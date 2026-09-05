# 9. Architecture Decisions

arc42 §9 · **2026-09-04**

Decisions are recorded as **ADRs** in the Michael Nygard format: append-only, numbered, and status-tracked. A decided ADR is **never edited**. Instead, supersede it with a new one and mark the old record `Superseded by ADR-NNNN`. Status is one of {proposed · accepted · superseded · deprecated}.

## Index

| ADR | Title | Status |
|---|---|---|
| [0001](#adr-0001) | Adopt arc42 for engineering documentation | accepted |
| [0002](#adr-0002) | Deploy our own Railgun pool (don't rent Railgun's) | accepted |
| [0003](#adr-0003) | Keep Groth16; reuse Perpetual PoT Phase-1 + own Phase-2 | accepted |
| [0004](#adr-0004) | Settle via nitro-on-railgun, not RelayAdapt | accepted |
| [0005](#adr-0005) | Amount privacy = Design A default; fork-lite deferred | accepted |
| [0006](#adr-0006) | Bootstrap anonymity set + Railgun onboarding bridge | accepted |
| [0007](#adr-0007) | Venue = RFQ posted-price v1; provably-fair CLOB is v2 | accepted · amended by 0011 |
| [0008](#adr-0008) | Transport = Waku pub/sub + libp2p-noise 1:1 | accepted |
| [0009](#adr-0009) | State ingestion targets nimbus-eth1; retire plugeth | accepted |
| [0010](#adr-0010) | Terminology conventions | accepted |
| [0011](#adr-0011) | Market-making & LP-buffered USDC yield are out of v1 | accepted |
| [0012](#adr-0012) | Delivery model: reuse-oriented incremental delivery; scopes as WBS work packages | accepted |
| [0013](#adr-0013) | Single team; scope spans the whole Armada product; Laconic is prior art | accepted |
| [0014](#adr-0014) | Pool + circuits are a clean-room reimplementation (license-clean) | accepted |

---

## ADR-0001
**Adopt arc42 for engineering documentation** · accepted · 2026-09-04

**Context.** A single mutable plan document was carrying four responsibilities at once — system decomposition, work breakdown, decision log, and schedule — and had been re-cut repeatedly, from workstreams to tiers to a tier/line matrix. The churn stemmed from the absence of separate, single-purpose artifacts and a stable "one model, many views" discipline.

**Decision.** Adopt **arc42** (Starke/Hruschka) as the single documentation template. The system uses one item model (the §5 registry), and tiers and lines become *views/tags* rather than competing hierarchies. Decisions live here as ADRs, cross-cutting concerns in §8, and risks in §11.

**Consequences.** All engineering docs conform to the 12 arc42 sections. Building-block detail stays in the `T0..T6` docs (§5), and "lines" are expressed through solution strategy (§4) and cross-cutting concepts (§8) rather than a parallel file tree.

**Alternatives.** C4/Simon Brown is diagram-centric and light on decisions and cross-cutting concerns; Views-and-Beyond/SEI is rigorous but heavyweight; and a bespoke structure is the status quo that failed.

## ADR-0002
**Deploy our own Railgun pool (don't rent Railgun's)** · accepted · 2026-09-04 · amended by ADR-0014

**Context.** We need a shielded pool with our own fee policy and POI policy, plus the freedom to attach a settlement contract. Railgun's OSS is redeployable.

**Decision.** Redeploy the Railgun OSS contracts and circuits under our own control (**T0.0**), configured with **fee = 0**, our POI allow-list, our token allow-list, and a reserved deposit/payout hook.

**Consequences.** We own upgrades and audit, and revenue comes from venue spread plus watcher metering rather than a pool skim. We do **not** modify Railgun's deployed protocol or touch their live pool; instead, a Railgun onboarding bridge (ADR-0006) imports value from their pool.

**Alternatives.** Renting Railgun's live pool — which carries their ~0.25% fee, no fee or POI control, and no settlement hook — was rejected.

## ADR-0003
**Keep Groth16; reuse Perpetual PoT Phase-1 + run our own Phase-2** · accepted · 2026-09-04 · amended by ADR-0014

**Context.** Groth16 requires a per-circuit trusted setup. Two questions remained open: whether to run our own trusted setup, and whether to switch to Halo2 to avoid setup altogether.

**Decision.** Keep **Groth16**. Reuse the community **Perpetual Powers of Tau** for Phase-1, and run **our own Phase-2 MPC** over our circuit set. Security rests on a **1-of-N honest** assumption, backed by a published transcript and a final randomness beacon (**T0.1**). Phase-2 is re-run **only** when circuits change — for example, the ADR-0005 fork-lite upgrade.

**Consequences.** There is no universal ceremony, since the expensive and risky part is inherited and vetted. Each circuit-set change costs a few days of MPC, and we take on no dependence on Railgun's specific ceremony instance. On-chain verification and mobile prover performance remain Groth16-cheap.

**Alternatives.** Reusing Railgun's published Phase-2 zkeys was rejected because it trusts their instance. Halo2/transparent setups were rejected because they carry worse mobile prover performance and on-chain gas, and would require a fresh audit, all to remove a soundness-only 1-of-N assumption.

## ADR-0004
**Settle via nitro-on-railgun, not RelayAdapt** · accepted · 2026-09-04

**Context.** We need optimistic exits and arbitrary off-chain settlement games over shielded value. Railgun's native path is RelayAdapt recipes, which are governance-gated on-chain call composition.

**Decision.** Build a **deposit/payout contract** (**T0.3**) that funds a `go-nitro` channel from a shielded note and settles the channel outcome back into fresh notes: **notes in → normal Nitro → notes out**. The Nitro game analysis stays isolated from the Railgun privacy analysis; the two meet only at the allocation→note **boundary**.

**Consequences.** We gain optimistic exits and arbitrary ForceMove games with **no new cryptography** and without governance-gated adapters. RelayAdapt is used only for the venue's atomic hedge leg (T4.2), never for settlement. The go-nitro maturity gaps — multi-asset outcomes, dispute wiring, and watchtower response — become our long poles (§11).

**Alternatives.** RelayAdapt-only settlement was rejected: it is governance-gated, less expressive, and offers no channel exits.

## ADR-0005
**Amount privacy = Design A default; fork-lite deferred** · accepted · 2026-09-04

**Context.** In the base construction, amounts are public at the deposit/payout boundary and on force-close. Hiding amounts in-play requires a circuit change.

**Decision.** Ship **Design A**, in which amounts are public only at the transparent boundary, as the default. **Native channelized commitment** ("fork-lite", **T0.6**), which provides amount-privacy-during-play, is an **optional Phase-4 upgrade** rather than part of the v1 spine. **Fork-full** — shielded ForceMove, research-grade — is **excluded** unless separately greenlit.

**Consequences.** v1 needs zero new crypto. Fork-lite, if greenlit, changes circuits and therefore re-runs Phase-2 (ADR-0003), requires a fresh audit, and imposes heavier mobile proving.

**Alternatives.** Fork-lite in v1 was rejected: it is unneeded for the feature set and adds ceremony, audit, and prover cost. Fork-full is out of scope.

## ADR-0006
**Bootstrap anonymity set + Railgun onboarding bridge** · accepted · 2026-09-04

**Context.** A fresh pool's anonymity set starts near zero, which is the biggest non-engineering risk. The request was to have our own liquidity while also accepting Railgun's.

**Decision.** **Bootstrap our own crowd** through incentives, our own market-making capital, and an LP flywheel, **plus a Railgun onboarding bridge**: a user unshields from Railgun, then shields into our pool in one public boundary hop. **Cross-pool membership proofs**, which would inherit Railgun's crowd, are deferred and optional, since they need a Phase-2 re-run and trust Railgun's root.

**Consequences.** This gives day-one dual liquidity. **Liquidity import ≠ anonymity-set inheritance**: a Railgun note cannot be spent or nullified in our contract, so "accept Railgun's" means the unshield→shield hop, not native ingestion. Privacy compounds with volume (**T0.7**).

**Alternatives.** Cross-pool proofs in v1 are deferred. Renting Railgun's pool for its set was rejected in ADR-0002.

## ADR-0007
**Venue = RFQ posted-price v1; provably-fair CLOB is v2** · accepted · 2026-09-04 · amended by ADR-0011

**Context.** Armada needs clearing, not a service-provider marketplace. Fair-ordering machinery is expensive.

**Decision.** The v1 venue is an **RFQ posted-price dealer market** (**T4**): Armada posts bid/ask on both sides on a take-it-or-leave-it basis and reposts if wrong, driven initially by a single `priceSetter` wallet that gives way to governance later. A **provably-fair CLOB** — a commit-reveal sequencer plus DSS, ex_net-grade — is a **v2** per-venue opt-in (**T3 + T4.6**).

**Consequences.** A posted price has nothing to front-run, so v1 removes the entire ordering problem. ex_net becomes a special case — a DSS-federated venue that also proves fair ordering — layered on the same spine rather than a rewrite.

**Alternatives.** A CLOB/matcher in v1 was rejected: it is months of work requiring a DSS and sequencer, and it is not needed to clear.

## ADR-0008
**Transport = Waku pub/sub + libp2p-noise 1:1** · accepted · 2026-09-04

**Context.** The system needs both broadcast, for feed and quote discovery, and low-latency point-to-point exchange, for vouchers and quotes, running on servers, browsers, and phones.

**Decision.** Use **Waku pub/sub** for gossip and discovery, and **libp2p-noise direct streams** for the 1:1 hot loop (**T2.3**). Mobile production uses **native gomobile** modules — go-waku plus go-libp2p — with an interim WebView browser-stack (**T6.5**).

**Consequences.** This reuses the transport Railgun's broadcaster network already speaks. Noise provides encryption and peer-id authentication only — **not IP privacy**, since a mixnet underlay is a separate concern. Native mobile libp2p is unbuilt (§11).

**Alternatives.** gossipsub-only is worse for 1:1 latency. js-libp2p on React Native is not RN-compatible, which is why the native port exists.

## ADR-0009
**State ingestion targets nimbus-eth1; retire plugeth** · accepted · 2026-09-04

**Context.** Watchers need fresh, proof-carrying Ethereum state. Our historical plugeth state-diff fork is badly out of date, and EIP-1186 alone is insufficient for what we emit.

**Decision.** Target **nimbus-eth1** — stateless/witness execution plus the Aristo state DB — as the state-diff source, mapped into IPLD for proof-carrying diffs (**T1**). We build only enough local cache to compute what we consume, not a full archive EL.

**Consequences.** T1 is the v1 long pole, since the nimbus stateless path is upstream-in-progress. It reuses the cerc-io ipld-eth lineage for codecs and backfill.

**Alternatives.** Reviving plugeth was rejected as stale. Bare JSON-RPC offers no proofs. Third-party state-diff services and ExEx were evaluated but not chosen for v1.

## ADR-0010
**Terminology conventions** · accepted · 2026-09-04

**Context.** Several terms collided with established meanings or with Railgun/Ethereum vocabulary, causing repeated confusion.

**Decision.** Enforce:
- **deposit/payout contract** for T0.3 — **never "adapter."**
- **RelayAdapt / Adapt modules / Cookbook** only for **Railgun's own** value-moving recipe; **Adapters** (capital-A tier) only for **Armada's T5**.
- **boundary** (or interface), **never "seam"** (collides with Feathers' testability term).
- **tier** (T0–T6), **never "layer"** (collides with Ethereum L1/L2).
- **venue / clearing**, **never "marketplace."**

**Consequences.** Docs are grep-clean for the banned terms, and cross-references use `T#.#` ids.

**Alternatives.** Ad-hoc wording is the status quo that caused the confusion.

## ADR-0011
**Market-making and LP-buffered USDC yield are out of v1** · accepted · 2026-09-05 · refines ADR-0007

**Context.** ADR-0007 fixed the v1 venue as a posted-price dealer market but left the dealer's *operation* implicit. An automated market-making solver — cross-venue internalization, flash-loan hedging, and Aave rebalancing (**T4.2**) — and the LP-buffered USDC-yield rail that rides it (**T4.5**) carry the venue's real operational complexity and risk: hedge MEV, LP cold-start, and the two timing clocks. None of it is needed to demonstrate private exchange and clearing.

**Decision.** v1 exchange is filled from a **simple pre-funded venue inventory** — an Armada-operated account holding both sides of a pair, filling at the posted price and topped up out-of-band. **No automated market-making** runs in v1: no internalization, no flash-loan hedge, no Aave rebalance loop. The **venue solver (T4.2)** and the **LP-buffered USDC-yield rail (T4.5)** move to **v2**. Consequently v1 yield is **ETH/wstETH only** (held as a note, T4.4), and **USDC yield defers to v2**.

**Consequences.** The net-new v1 venue surface shrinks to the posted-price contract (T4.0), the quote/settle app (T4.1), and the fee-split wiring (T4.3), all settling over reused go-nitro clearing. The yield→USDC swap still ships, filled from static inventory rather than a hedged maker. The hedge-MEV and LP-cold-start risks become v2 concerns (§11). v1 capacity is bounded by the static inventory and refilled manually; running dry means quotes are withdrawn, never a custody risk.

**Alternatives.** Ship the automated solver + LP rail in v1 (rejected — that is the market-making operation, the venue's hardest and riskiest part, and it is not required to clear). Drop exchange from v1 entirely (rejected — the yield→USDC swap is a v1 deliverable and works against static inventory).

## ADR-0012
**Delivery model: reuse-oriented incremental delivery; scopes as WBS work packages** · accepted · 2026-09-05

**Context.** Requirements are given (the Armada docs), most components already exist (Railgun, go-nitro, watcher-ts, the wallet), and the key architecture choices are already made and frozen as ADRs 0002–0011. The remaining job is to specify and integrate **incrementally** without re-opening settled decisions or drifting across adjacent scopes. This is the classic shape of a **reuse-oriented, requirements-given programme** — not greenfield design, and not agile requirements-discovery.

**Decision.** Adopt an operating model with five rules, each drawn from established practice:
1. **Reuse-oriented specs** (Sommerville reuse-oriented SE; CBSE). Each spec states **interfaces + the delta (config + net-new glue) + acceptance**, and *cites* reused components at pinned commits rather than re-documenting their internals.
2. **WBS of work packages** (PMI, 100% rule; MECE). Decompose the programme into deliverable-oriented work packages — **A** nitro-on-railgun, **B** watcher parties, **C** yield, **D** client — each a system with a context boundary (arc42 §3) and an interface-control contract (ICD) to its siblings (INCOSE; ISO/IEC 15288). See [`00-work-packages.md`](./00-work-packages.md).
3. **Fixed architecture baseline.** ADRs 0002–0011 are the baseline; specs build on them and never relitigate. Change means *superseding* an ADR, never editing a decided one.
4. **Traceability** (ISO/IEC/IEEE 29148). Every requirement traces up to an Armada requirement and down to a work package, and is written to the 29148 bar: unambiguous, verifiable, complete, traceable.
5. **Risk-first increments** (Boehm spiral / Boehm–Lane ICSM; McConnell staged delivery). Deliver against the baseline in increments sequenced by risk; one **walking skeleton** retires integration risk first, then each work package deepens; the long poles (T1 ingestion, mobile transport, go-nitro maturity) get spikes first.

**Consequences.** Tiers (T0–T6) remain the shared **building-block vocabulary** (the product breakdown); work packages **own** subsets of tier-items and **consume** the rest across ICDs. The arc42 docs become per-work-package architecture descriptions plus a program context. Scope drift is structurally prevented: a work package may touch only its own boundary and its own deltas — everything else is a cited baseline decision or an external interface.

**Alternatives.** One monolithic plan across all tiers (rejected — it caused repeated re-cutting and scope drift). Pure agile backlog discovery (rejected — requirements are given, not to be discovered). Greenfield design docs (rejected — most components exist; the work is integration).

## ADR-0013
**Single engineering team; scope spans the whole Armada product; Laconic is prior art** · accepted · 2026-09-05 · amends ADR-0012 and the §3 build-ownership boundary

**Context.** Earlier docs drew an *Armada-builds-vs-Laconic-builds* ownership boundary (§3; T5 marked "out of Laconic build scope"; the ADR-0012 / §0 WBS treated Armada's Adapters as external). That boundary no longer holds. This is **one engineering team** building **Armada — a Railgun-based privacy product for USDC** (confirmed by Armada's own thesis and docs: Railgun ZK circuits (BN254/Groth16), a shielded USDC pool, Private Proofs of Innocence, non-custodial). **Laconic's existing work — nitro integration, watchers, wallet, chain-signatures, ingestion — is prior art the team reuses**, not a separate vendor's deliverable.

**Decision.** Scope **both** the Armada-side and Laconic-side tasks as one programme. The WBS covers the whole product, including what was "T5 Armada Adapters" (CCTP, Aave-v4 yield, Swaps) — now in-scope work, not external systems. The governing framing is **"build Armada on Railgun, reusing Laconic prior art,"** not "Laconic supplies a substrate to Armada."

**Consequences.** §1 purpose reframes from *Laconic work to satisfy Armada requirements* to *the team's work to build Armada, a Railgun-based product, on Laconic prior art*. §3's build-ownership boundary becomes a **reuse boundary** (what already exists vs what is net-new), not an org boundary. T5's items enter the WBS; "external systems" shrinks to genuine third parties (Ethereum L1, Circle/CCTP, Aave, the Railgun OSS we redeploy, nimbus-eth1 upstream, the Waku network). The architecture baseline (ADRs 0002–0011) is unchanged — those technical choices stand.

**Alternatives.** Keep the two-org boundary (rejected — it's one team; the boundary created artificial "out of scope" gaps like T5 and mis-framed the yield/adapter work).

## ADR-0014
**Pool + circuits are a clean-room reimplementation (license-clean)** · accepted · 2026-09-05 · amends ADR-0002, ADR-0003

**Context.** The A deep dive ([A.1.10](../A-nitro-on-railgun/A.1-reuse-inventory.md)) found Railgun's `circuits-v2` carries an explicit *"No License is provided for any party under any circumstances"* file and the pool contracts are SPDX `UNLICENSED`. So ADR-0002 (redeploy the OSS pool) and ADR-0003 (own Phase-2 over Railgun's circuits) are **not executable as written**. `go-nitro`/`ts-nitro` are separately-licensed OSS and unaffected; Railgun's `engine`/`cookbook` are MIT (usable as reference), but they are not the on-chain/circuit pieces.

**Decision.** Reimplement the **shielded-pool contracts** (T0.0) and the **JoinSplit circuits** (T0.1) **clean-room** — independently authored, spec-compatible with the Railgun design the deep dive documents (A.1.1/A.1.2), using no Railgun-licensed source. Our engineers own this build. Everything else stands: reuse `go-nitro` (T0.2), `ts-nitro` (T6.2), and `engine`/`cookbook` as MIT reference; reuse the community Perpetual PoT Phase-1 and run our own Phase-2 **over our own circuits** (ADR-0003 otherwise unchanged); keep Groth16.

**Consequences.** T0.0 and T0.1 change status from **reuse/redeploy → net-new (clean-room)** — a real, audit-critical crypto-engineering workstream, though of a well-understood design that the deep dive de-risks by pinning the exact behavior/format to match. The "integration, not new cryptography" thesis (ADR-0012, §4) now holds for the **settlement rail** (T0.2 adjudicator + T0.3 deposit/payout + T6.x) but **not** for the pool/circuits. **A.1 (the reuse inventory) doubles as the reference spec** the clean-room implements. T0.6 (native commitment) is a change to our own circuits. Note format may match Railgun (to ease the T0.7 onboarding bridge) or diverge — our choice; our anonymity set is separate regardless.

**Alternatives.** Obtain a Railgun grant/relicense (unavailable — we go clean-room). Use Railgun's live deployed pool directly (rejected — no fee=0/own POI, no settlement hook; contradicts ADR-0002's rationale).
