# 4. Solution Strategy

arc42 §4 · **2026-09-04**

This section records the fundamental decisions that shape the system and the order in
which they ship. Every claim grounds in the §5 registry (for item ids) and the §9 ADRs
(for rationale); the supporting detail lives in the `T#.#` tier docs, and the risk
detail lives in §11.

## 4.1 Core strategy — *nitro-on-railgun*

The load-bearing idea is one sentence (ADR-0004):

> **notes in → normal Nitro → notes out.**

A shielded note is unshielded into a **deposit/payout contract** (T0.3) that funds a
plain `go-nitro` channel. Parties then play arbitrary off-chain **ForceMove** games, and
the channel outcome is shielded back into fresh notes. The Nitro game analysis and the
Railgun privacy analysis stay fully isolated; they meet only at the allocation→note
**boundary** (ADR-0010: *boundary*, not seam). This buys optimistic exits and arbitrary
settlement games over shielded value **with no new cryptography** and without
governance-gated RelayAdapt recipes. RelayAdapt is Railgun's own module, used only in
the T4.2 hedge leg and never for settlement.

**This is integration, not *new* cryptography — with one caveat (ADR-0014).** The thesis
holds fully for the **settlement rail** (T0.2 adjudicator + T0.3 deposit/payout + T6.x),
which is pure reuse and integration. It does **not** hold for the shielded **pool (T0.0)
and JoinSplit circuits (T0.1)**: because Railgun's contracts/circuits are unlicensed
(A.1.10), those are **clean-room net-new** — independently authored, spec-compatible with
the Railgun design, and therefore audit-critical crypto-engineering (of a well-understood,
de-risked design, not novel cryptography). The hard cryptographic assets break down as:

- **T0.0 / T0.1** — the shielded pool and its JoinSplit circuits, **clean-room
  reimplemented** (spec-compatible with Railgun, fee=0, our POI policy; ADR-0014), plus our
  own Phase-2 ceremony. Net-new and audit-critical, but a reimplementation of a known
  design, not novel cryptography. Amounts are hidden in-circuit; the SNARK enforces
  value-conservation + range, so *$50 and $5M are equally opaque* — no Tornado-style
  denomination buckets.
- **T0.2** — vanilla `go-nitro` `NitroAdjudicator` / `ForceMove` / `MultiAssetHolder`,
  deployed and configured, not forked.

Against that, the **net-new v1 surface** is enumerable — the clean-room pool/circuits plus
a thin integration layer:

| Net-new for v1 | id |
|---|---|
| Clean-room shielded pool + JoinSplit circuits (spec-compatible with Railgun; audit-critical) | **T0.0, T0.1** |
| Deposit/payout contract (the Nitro↔Railgun boundary) | **T0.3** |
| Watcher ingest config (index pool commitments/nullifiers + T0.3 events) | **T1.2** |
| Posted-price venue + quote/settle app + fee-split (v1 fills from static inventory) | **T4.0, T4.1, T4.3** |
| Wallet bits (note-scanner, settlement client, watchtower, transport, proving) | **T6.1–T6.7** |
| Anonymity-set strategy (process, not code) | **T0.7** |

The only *novel* cryptography optional anywhere in the plan is **T0.6** (native
channelized commitment, "fork-lite"). It is deferred off the spine (ADR-0005) and, if
greenlit, re-runs the Phase-2 ceremony over our own circuits (ADR-0003). v1 needs **no
novel cryptography and zero new consensus/ordering** (ADR-0007) — the clean-room
pool/circuits (T0.0/T0.1) reimplement a known design rather than invent one — and because
a posted price has nothing to front-run, the fair-ordering stack (T3) does not exist in v1.

## 4.2 How the top quality goals are met

Each §10 quality goal maps to a concrete approach realized by named items:

| Quality goal (§10) | Approach | Realized by |
|---|---|---|
| **Non-custody** — funds never held by a third party; failure = halt, not loss | Value lives in Railgun notes or in `NitroAdjudicator` escrow allocated only by co-signed state; every position is unilaterally **force-closable** to its last co-signed state via the ForceMove dispute game (`challenge`→`checkpoint`→`conclude`) | T0.2, T0.3; watchtower response T6.3 |
| **Read-time privacy** — sync/read without leaking to a public endpoint | **Proof-carrying feeds**: the watcher serves `getStorageAt → {value, proof}` verified client-side against the block state root, so a client never calls a public RPC; anonymity-set privacy compounds with the crowd | feeds T2.0 (ingested via T1); anonymity set T0.7 |
| **Mobile-first** — a phone is a full non-custodial peer | **WebView-first, then native**: interim in-browser ts-nitro stack, then a native gomobile transport behind the same interface; on-device Groth16 proving keeps the shield/unshield prover phone-cheap | WebView→native T6.5; Groth16 proving T6.6 (ADR-0008) |
| **Clearing at a fair, un-front-runnable price** | **RFQ posted-price** dealer market: Armada posts both sides take-it-or-leave-it and reposts if wrong — a posted price has nothing to front-run; v1 fills from a static inventory (no market-making) and settles non-custodially over the T0.3 rail | posted-price venue T4.0/T4.1/T4.3 (ADR-0007); solver T4.2 is v2 (ADR-0011) |

Design A (amounts public only at the transparent T0.3 boundary, ADR-0005) is the default
that makes all four goals reachable with no ceremony. Hiding amounts in-play is the
opt-in T0.6 upgrade, which sits outside the v1 spine.

## 4.3 Delivery increments

Increments are the **release** facet of the §5 registry, read as a shipping order:

- **v1 — Armada support (the spine).** Clean-room pool + circuits + own Phase-2 (T0.0/T0.1), Nitro
  adjudicator (T0.2), **deposit/payout boundary (T0.3)**, ingestion (T1.0–T1.2),
  watcher feeds + metering + transport (T2.0–T2.3), **RFQ posted-price venue filled from a
  static inventory (T4.0, T4.1, T4.3) + ETH/wstETH yield (T4.4)**, Armada adapters +
  routing (T5.0/T5.1), full wallet (T6.0–T6.7), anonymity-set bootstrap + Railgun import
  bridge (T0.7 / ADR-0006). This is the clean-room pool/circuits (audit-critical, of a
  known design; ADR-0014) plus integration + one small boundary contract + thin app code
  — no market-making, no new consensus, no novel ZK research (ADR-0011).
- **v2 — market-making + ex_net matcher (price discovery + fair ordering).** The venue
  solver / automated market-making (T4.2) and the LP-buffered USDC-yield rail (T4.5); the
  whole ordering stack (T3.0–T3.2); the ex_net matcher + LP vault (T4.6); and the
  economic-security contracts they need on-chain (registry+bond T0.4,
  sequencing-cert/fraud-proof verifier T0.5, federation+bond DKG T2.4). This layers on the
  *same* spine — a DSS-federated venue that also market-makes and proves fair ordering,
  not a rewrite (ADR-0007, ADR-0011).
- **opt — amount-privacy-in-play.** Fork-lite T0.6 (Phase-4, re-runs Phase-2 + fresh
  audit) and the optional Self zk-passport identity T6.8. This is greenlit only against
  a concrete requirement; fork-full (Design B) stays excluded (ADR-0005).

### Walking skeleton — the first integration target

Before hardening any single tier, stand up **one thin end-to-end slice** on a Laconic
fixturenet that exercises every boundary at least once:

```
shield → deposit (T0.3) → trivial ForceMove settle → payout → scan
  T0.0        T0.3            T0.2 / T4.1 stub          T0.3     T6.1
```

Concretely, the slice shields a known value into the pool (T0.0), then unshields-in
through the deposit/payout contract to fund a `go-nitro` channel (T0.3 over T0.2). It
drives a **trivial** ForceMove app — a single co-signed fill, HashLockedSwap-grade,
standing in for the real T4.1 quote/settle — to a finalized outcome, shields that
outcome into fresh notes via payout (T0.3), and scans and confirms balances on-device
with **zero** public `eth_getStorageAt`/`eth_getLogs` (T6.1 over a T1.2/T2.0 feed). This
proves the *notes-in-Nitro-notes-out* thesis end-to-end with the smallest possible game,
and it turns every subsequent tier task into "deepen a boundary that already works"
rather than a big-bang integration. The fixturenet E2E is the standing acceptance
harness each tier doc already references (T0.3, T1.0, T2.0/T2.1, T4, T6.2 all test
against it).

### Critical path through the dependency DAG

```
        T0.0 ─┐
              ├─► T0.3 ──► T1.2 ──► T2.0/T2.1 ─┐
        T0.2 ─┘    │                            ├─► T4 (venue)  ─► T5
        T0.1 ──────┘        T1.0/T1.1 ──────────┤
                                                └─► T6 (wallet) ─► walking skeleton
```

- **Do-first:** T0.0 + T0.1 (the clean-room pool and circuits + Phase-2 keys) and T0.2 feed **T0.3**, the
  boundary contract every other tier binds to. Nothing above T0 can integrate until
  T0.3's address, ABI, events, and lifecycle exist.
- **Then fan out in parallel** once T0.3 is real: the **T1 → T2** feed path (ingest
  pool + T0.3 events, serve proof-carrying reads), the **T4** venue, and the **T6**
  wallet can all progress concurrently against T0.3's published interface.
- **The long pole is T1 ingestion** (ADR-0009): the nimbus-eth1 state-diff emitter
  (T1.0) is the one v1 item that is genuinely unbuilt rather than integration, and
  every read-side consumer (T2 feeds → T4/T6) assumes its fresh, proof-carrying view.
  The proof-carrying feed leak-safety (T2.0) is the *do-first* item on the read path
  because a single leak collapses the whole anonymity set. go-nitro maturity gaps
  (the net-new multi-asset ForceMove app, dispute wiring) surface first at T0.3 and are Phase-3
  hardening, not spine blockers.

---

Cross-refs: quality goals → §10; runtime flows for these slices → §6; deployment of the
fixturenet + peers → §7; recurring concepts (privacy, settlement, transport, metering)
→ §8; the long-pole/maturity risks summarized here → §11; item ids/status/release →
§5; decisions → §9 (ADR-0002…0010).
