# 11. Risks & Technical Debt

arc42 §11 · 2026-09-04

This chapter catalogs the material risks to the v1 spine and its Phase-3 hardening,
together with the shortcuts we ship on purpose. Every row owns a `T#.#` item (§5
registry) and, where a decision has already scoped it, an `ADR-####` (§9). All ratings
are v1-relative: **impact** measures the consequence if a risk bites, and **likelihood**
measures the odds it bites before we mitigate.

## Risk register

| # | Risk | Impact | Likelihood | Mitigation | Owner |
|---|---|---|---|---|---|
| R1 | **Anonymity-set cold-start** — a fresh pool's k-anonymity starts near zero, and privacy only compounds with volume. This is the biggest **non-engineering** risk: no code fixes it, it needs a crowd. | High (the product's core promise degrades to k≈1) | High at launch, decaying with volume | Bootstrap own crowd (incentives, own market-making capital, LP flywheel) **+** Railgun onboarding bridge for day-one dual liquidity; aggregation-window / dribble discipline at the boundary; instrument effective-k. Import ≠ set inheritance. | T0.7 · ADR-0006 |
| R2 | **go-nitro maturity gaps** — outcomes are effectively **single-asset** (swaps need ETH-in/USDC-out); **dispute wiring** exists in the adjudicator but is not driven end-to-end in the node; and **watchtower response** plus **persistent connections/liveness** are unproven for our flow. go-nitro is documented pre-production. | High (settlement is spend-authorizing; a gap = stuck or lost funds) | Medium — primitives exist, our wiring does not | Treat all four as **pre-mainnet blockers, not spine-demo blockers**; multi-asset outcome encoding first (the swap-demo pull); drive challenge/response E2E on fixturenet; watchtower response is T6.3 gated on T2 freshness; liveness ties to transport. | T0.2 / T0.3 / T6.3 · ADR-0004 |
| R3 | **T1 ingestion long pole** — the nimbus-eth1 **stateless/witness path is upstream-in-progress**, and the state-diff **emitter module** wired to our watched set is server-side net-new. A lagging or stalled emitter is a **safety** problem (a missed nullifier or missed adjudicator `Challenge`), not just latency. | High (stale feed breaks watchtower safety, T6.3) | Medium–High — the one genuinely unbuilt v1-spine tier | Scope to just-enough local state (watched-address subtrie via Aristo), reuse cerc-io ipld-eth codecs/backfill; T1.2 exposes a **head cursor**; T6.3 **gates on feed freshness** and treats staleness as a liveness alarm with redundant-party fallback. | T1.0 · ADR-0009 |
| R4 | **Mobile transport crux** — native **libp2p/waku on React Native is unbuilt**, since `@waku/sdk` / js-libp2p is not RN-compatible (no WebCrypto/WebRTC/Node built-ins in Hermes). The interim **WebRTC-over-tunnel** path is the least-proven part of the combination and may not traverse a network underlay cleanly. | High (no first-class mobile peer → foreground-only, fragile connectivity) | Medium — the proven `status-go` gomobile pattern de-risks it | Sequenced, not competing: ship interim **WebView** stack (`@cerc-io/peer` + `ts-nitro`) as the realistic v1 client, then port the p2p/channel/messaging layer to **one go-nitro + go-waku gomobile module** behind the same transport interface T6.2/T6.3 bind to. Native E2E is a Phase-3 gate. | T6.5 · ADR-0008 |
| R5 | **Audit surface** — anything **spend-authorizing** is audit-critical before mainnet: the T0.3 deposit/payout contract (the notes-in/notes-out boundary), the T0.0 pool config deltas (fee=0, POI gating, settlement-hook authorization), and the T6.1 WebView↔RN key bridge / T6.0 enclave custody. | Critical (a leak here = drained funds or leaked keys) | Low if audited, catastrophic if skipped | Freeze spend-authorizing surface early; pin Railgun OSS commits and re-audit config deltas on any bump; narrow, audited bridge with raw key material never crossing into the WebView; **external audit gates mainnet**, not the demo. | T0.0 / T0.3 / T0.6 / T6.0 / T6.1 · ADR-0002 |
| R6 | **Trusted-setup ceremony logistics** — Phase-2 MPC security is **1-of-N honest**, so the assumption is only credible if contributors are genuinely independent. Coordinating ≥5 organizationally-distinct participants, publishing transcript + beacon, and regenerating/redeploying verifiers is a process risk, not a code risk. | Medium (a botched ceremony undermines soundness of every proof) | Low — snarkjs path is standard, Phase-1 is inherited | Reuse vetted **Perpetual PoT** Phase-1 (never re-run); target **≥5 independent contributors**; publish `.zkey`/`.vkey` hashes + verification transcript in-repo; CI fails on `zkey` hash mismatch. Re-run **only** on a circuit change (T0.6 / cross-pool). | T0.1 · ADR-0003 |
| R7 | **Venue market-making — hedge MEV / liquidity fragmentation (v2)** — the automated solver's residual hedge is a **public** tx: front-runnable / MEV-able and dependent on public-venue liquidity; per-venue shielded inventory also **splinters liquidity**. **This is a v2 risk** — v1 fills from a static inventory and neither hedges nor rebalances (ADR-0011). | Medium (worse fills, thin depth — never lost custody) | N/A in v1 (deferred with T4.2/T4.5) | When built: internalize deep inventory, **batch + decorrelate** the residual, route the RelayAdapt hedge through a broadcaster; an **opt-in shared hub / pooled-maker vault** deepens depth without forcing consolidation. | T4.2 · ADR-0011 |
| R8 | **Price-setter trust until governance** — v1 posts prices from a **single privileged `priceSetter` wallet**, so a wrong or malicious price yields bad fills. | Low–Medium (bad fills, **not** lost custody — fills are take-it-or-leave-it and force-closable) | Medium while single-key | Fills are take-it-or-leave-it and force-closable via T0.3; **migrate `priceSetter` → governance** (same admin-key→governance path as T0.0 POI authority and T0.3 app allow-list); trust is **removed entirely** only by the T4.6 CLOB upgrade. | T4.0 · ADR-0007 |

## Technical debt / known shortcuts

Each of the following is accepted deliberately for v1: a scoped, reversible interim
rather than an unbounded liability. We list them here so they stay tracked rather than
forgotten.

- **Interim WebView transport (T6.5).** v1 mobile hosts the browser peer stack
  (`@cerc-io/peer` + `ts-nitro`) inside a WebView, giving foreground-only,
  WebRTC-over-tunnel connectivity. This reuses production browser code verbatim and
  validates the whole flow, but the native **go-nitro + go-waku gomobile** module remains
  the production target — a Phase-3 hardening gate before mainnet. The debt retires once
  the native module lands behind the same transport interface. → R4, ADR-0008.
- **Single `priceSetter` (T4.0).** One privileged wallet holds the posted bid/ask in v1.
  Interim trust is reduced by the **governance migration** and removed only by the T4.6
  provably-fair CLOB (T4.6 + T3), which is months of work deferred to v2. → R8, ADR-0007.
- **v1 Design-A boundary leak accepted (T0.3).** Amounts are **public at the deposit/payout
  boundary** and on force-close, which is inherent to Design A; we ship it as the default
  because it needs **zero new crypto**. Hiding amounts in-play is the **optional** fork-lite
  upgrade (T0.6), which changes circuits and therefore re-runs the T0.1 Phase-2 ceremony,
  requires a fresh audit, and imposes heavier mobile proving; it is greenlit only on a
  concrete requirement. In the meantime the leak is mitigated by aggregation-window /
  dribble / quantization at the boundary (T0.7), not eliminated. → R1, ADR-0005.
- **Static venue inventory, no market-making (T4.0/T4.1).** v1 fills exchange from a
  pre-funded inventory topped up out-of-band; the automated market-making solver (T4.2)
  and the LP-buffered USDC-yield rail (T4.5) — and USDC yield itself — are deferred to v2.
  Capacity is bounded and refilled manually; running dry withdraws quotes, never a custody
  risk. → ADR-0011.

Cross-refs: risk mitigations are realized in §6 (runtime) and §8 (cross-cutting concepts);
quality targets these risks defend are in §10; ownership and status per §5 registry.
