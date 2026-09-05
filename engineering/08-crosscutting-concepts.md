# 8. Cross-cutting Concepts

arc42 §8 · 2026-09-04

Some concerns recur **across** tiers, tracing "lines" through the §5 registry. Each such concept is described **once** here. Its realizing building blocks are named by their `T#.#` id — the detail lives in the tier docs and is not repeated — and its governing decision by `ADR-####`. Where tiers group the registry vertically, these concepts group it horizontally.

| # | Concept | Realized by | Governing ADR |
|---|---|---|---|
| 8.1 | Settlement model — nitro-on-railgun | T0.2, T0.3, T0.6 | ADR-0004 |
| 8.2 | Privacy model | T0.0, T0.3, T0.6, T0.7 | ADR-0005, ADR-0006 |
| 8.3 | Non-custody & exit | T0.2, T0.3, T6.3 | ADR-0004 |
| 8.4 | Transport | T2.3, T6.5 | ADR-0008 |
| 8.5 | Metering | T2.1, T4.3 | — |
| 8.6 | Trusted setup | T0.1 | ADR-0003 |
| 8.7 | Identity (optional) | T6.8 | — |

## 8.1 Settlement model — nitro-on-railgun

The load-bearing architectural idea is **notes in → normal Nitro → notes out** (ADR-0004). A shielded note is unshielded into the **deposit/payout contract** (T0.3), which funds a vanilla `go-nitro` channel (T0.2). Arbitrary off-chain ForceMove games and vouchers then play out, and on finalize — whether by cooperative close or by post-challenge force-close — the outcome is shielded back into **fresh** notes. This gives us Nitro's optimistic exits and expressive settlement games over Railgun-style value **with no new cryptography** and **without governance-gated RelayAdapt** modules.

The decisive property is **isolation**. The Nitro game analysis and the Railgun privacy analysis stay fully separate, meeting only at the **allocation→note boundary** (T0.3). Neither side has to reason about the other's internals; the boundary is the single place where the two models are stitched together. RelayAdapt, Railgun's own recipe, appears only inside the venue solver's atomic hedge leg (T4.2) — **never** in settlement.

The optional amount-privacy upgrade (T0.6, §8.2/ADR-0005) preserves this isolation exactly. It changes only what the allocation→note boundary opens, exposing commitments rather than cleartext balances; the Nitro game still runs its normal challenge, checkpoint, and conclude flow unchanged.

→ detail in §5 T0.2, T0.3; runtime sequences in §6.

## 8.2 Privacy model

This concept fixes what is hidden, from whom, and where value is exposed — the model every other tier is built to defend.

- **Shielded value (T0.0).** Value lives as commitments in our own Railgun pool (ADR-0002). The SNARK enforces value-conservation and range, so **amounts inside the shield are opaque regardless of size**; no Tornado-style denomination buckets are needed.
- **Design-A boundary leak (T0.3, ADR-0005).** Amounts become public **only** at the two transparent boundaries — the `shield`/`unshield` hop and the T0.3 deposit/payout — and on force-close. Hiding amounts *in-play and on dispute* requires the **native channelized commitment** upgrade (T0.6, "fork-lite"), an optional Phase-4 item that is not on the v1 spine; fork-full (shielded ForceMove) is excluded unless separately greenlit.
- **Anonymity set + bootstrap/import (T0.7, ADR-0006).** Pool privacy is k-anonymity over the set of unspent indistinguishable notes: a fresh pool starts near zero and compounds with volume. There are two levers — bootstrapping our own crowd, plus a Railgun onboarding bridge (unshield-from-Railgun → shield-into-ours, one public boundary hop). Crucially, **liquidity import is not anonymity-set inheritance**: a Railgun note cannot be spent or nullified in our contract, so the bridge only moves capital that must then build a crowd here.
- **POI: gated entry / fresh-shield exit (T0.0, T0.3).** We keep Railgun's Proof-of-Innocence construction under **our** allow-list policy. The deposit endpoint checks the POI root before funding (gated entry), and payout mints fresh notes that preserve POI (fresh-shield exit) with no linkage back to the deposit note beyond the public deposit amount. Importing value does **not** import POI status.
- **Read-side vs write-side unlinkability (duality).** The two ends of the same anonymity set are protected by different components. The **read side** is covered by the T2 watcher's proof-carrying, identical-bytes feeds and local scan (T2.0, consumed by the T6.1 scanner), so no RPC fingerprint reveals *which* commitments or nullifiers you read. The **write side** is covered by the keeper/broadcaster submit-on-behalf path (T6.5), so no EOA links you to *submitting*. Same set, opposite ends. Per-channel address rotation (T6.4) decorrelates the Design-A boundary across interactions. Note (ADR-0008): transport noise gives IP-level encryption and peer-id auth **only, not IP privacy** — see §8.4.

→ detail in §5 T0.0, T0.3, T0.7 and T2.0, T6.1, T6.4; quality scenarios in §10; cold-start risk in §11.

## 8.3 Non-custody & exit

No tier ever takes custody of user value, and correctness of settlement means a user can **always exit unilaterally**. Value is either a shielded note (T0.0) or an allocation in a Nitro channel the user co-signs (T0.2) — never held by a venue, watcher, or relay. A liveness failure therefore degrades to **halt, not loss**.

The exit guarantee rests on the ForceMove dispute machinery — `forceMove`, `respond`, and `checkpoint` on `NitroAdjudicator`, owned in T0.2 — plus its counterpart, the **watchtower**, which supplies the response half. This is the **watcher/watchtower duality**: a watchtower is a watcher's read feed (T2 freshness) **plus** a bonded challenge responder (T0.2). The phone-resident watchtower (T6.3) composes the two. It watches for a stale-state force-close over the T2 feed and submits the user's higher-turn co-signed state within the challenge window, so a dead counterparty freezes trading but never funds. T6.3 **gates on T2 feed freshness**: a stale feed is treated as a liveness alarm — fall back to a redundant party or direct submit — not a silently missed window.

→ detail in §5 T0.2, T6.3; challenge/watchtower flow in §6; "never miss a valid challenge" scenario in §10.

## 8.4 Transport

We run two transports, each chosen for what it does best (ADR-0008). **Waku pub/sub** handles feed discovery, gossip, and async delivery; it is recipient-unlinkable within a content-topic's anonymity set, and it is the transport Railgun's broadcaster network already speaks. **libp2p-noise direct streams** handle the 1:1 hot loop — voucher and quote exchange, low-latency reads — that Waku's multi-hop store-and-forward cannot meet at tens-of-ms latency. A circuit-relay plus STUN/TURN bootstrap is the only fixed transport infrastructure. This substrate is owned by T2.3.

On mobile the same two transports are a **port, not a new protocol** (T6.5). An interim WebView browser stack carries the spine demo, followed by a production **native gomobile** module (go-waku plus go-libp2p-noise) sitting behind the same transport interface that the settlement client (T6.2) and watchtower (T6.3) bind to. The js-libp2p stack is not React-Native-compatible, which is why the native port is required rather than optional polish (§11 crux).

**Scope boundary (recurring caveat):** noise provides encryption and peer-id auth **only — not IP privacy**; endpoints see each other's IPs. A mixnet underlay that hides IP is a separate, out-of-scope concern, not part of T2.3 or T6.5.

→ detail in §5 T2.3, T6.5; deployment of relays/peers in §7.

## 8.5 Metering

Reads and settlement are **paid the same way**: private per-request **go-nitro payment vouchers** that net into a co-signed virtual-channel allocation, settled to L1 in USDC. On the read path (T2.1), paid feed methods are metered with vouchers **funded at note creation**, so paying for reads leaks nothing about the reader. On the write/settle path (T4.3), the venue fee is **just a fee-split in the channel allocation** — user, protocol, and integrator shares on the co-signed channel state, with no fee vault and no negotiation — netted on close. Both paths reuse the identical go-nitro voucher/virtual-channel mechanism (`payments/vouchers.go`, `protocols/virtualfund/virtualfund.go`), the same one T0.2/T0.3 already publish. Because metering is non-custodial, the worst case is refused service, never loss. These two flows are the project's two revenue pillars — **watcher metering + venue spread**.

→ detail in §5 T2.1, T4.3; runtime metering step in §6.

## 8.6 Trusted setup

Groth16 needs a per-circuit setup, and the concept here is a **reuse-Phase-1 / own-Phase-2** discipline (ADR-0003), owned as a process — not code — in T0.1. Phase-1 reuses the community Perpetual Powers of Tau, the expensive, vetted, circuit-agnostic part; we never run our own universal ceremony. Phase-2 is **ours**: a circuit-specific MPC over our deployed set, with **1-of-N honest** security (≥5 organizationally distinct contributors), a published transcript, and a final randomness beacon. The Solidity verifier is generated from the final vkey and wired into T0.0.

The load-bearing coupling is that **any circuit change re-runs Phase-2** while Phase-1 is never re-run. The canonical trigger is the T0.6 amount-privacy upgrade (§8.2), and the optional cross-pool membership path (T0.7) is the other. New `wasm` and `zkey` artifacts then ship to the mobile prover (T6.6). This is why fork-lite carries a ceremony, fresh-audit, and heavier-prover cost (ADR-0005): nothing new is invented, the same pipeline is simply re-executed.

→ detail in §5 T0.1; prover constraint in §5 T6.6; deferral rationale in §11.

## 8.7 Identity (optional)

Identity is called out here explicitly to record that it is **not** cross-cutting-required. It is a single **optional** wallet parameter (T6.8), default off. A user may source a selective-disclosure attestation off-band — for example a Self zk-passport (nationality, age, OFAC-clear, personhood) — proven on-device and verified **off-chain, anchored to an Ethereum address**, with the shielded pool identity left unlinked. The wallet stores it as one more optional signing input, surfaced only when a T4 venue's match filter requires it. It threads through no other tier and is deferred rather than built, so it is listed as a concept only to fix its non-cross-cutting scope.

→ detail in §5 T6.8; per-venue loyalty-key policy in §5 T6.4.

---

*Decisions behind these concepts: [§9 ADRs](./09-architecture-decisions.md). The building blocks they span: [§5 registry](./05-building-block-view.md). Quality scenarios that exercise them: [§10](./10-quality-requirements.md); open risks: [§11](./11-risks-and-technical-debt.md).*
