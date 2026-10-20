# 4. Solution Strategy

arc42 §4 · **2026-10-20**

This section records the fundamental decisions that shape the system and the order in
which they ship. Every claim grounds in the §5 registry (for item ids) and the §9 ADRs
(for rationale); the supporting detail lives in the `T#.#` tier docs, and the risk
detail lives in §11.

## 4.1 Core strategy — *nitro-on-railgun*

The core idea is one sentence (ADR-0004):

> **notes in → normal Nitro → notes out.**

A shielded note is unshielded into `ArmadaNitroAdjudicator`'s `depositFromNote`, which
funds a Nitro channel. Parties then play arbitrary off-chain **ForceMove** games, and a
`SHIELD`-type allocation pays the channel outcome back out into fresh notes. The code
merges this deposit/payout boundary into the adjudicator itself; there is no separate
boundary contract and no `fundChannel` (ruling 1, `labbook/2026-10-19-docs-versus-code.md`;
T0.2 in §5). The Nitro game analysis and the Railgun privacy analysis stay fully
isolated; they meet only inside `depositFromNote` and the SHIELD payout path. This buys
optimistic exits and arbitrary settlement games over shielded value **with no new
cryptography**. Armada's adapters do not use Railgun's RelayAdapt/Cookbook recipe
pattern anywhere; the Aave-v4 yield adapter instead calls the pool's own
`adaptContract`/`adaptParams` binding (T5.0b, `ArmadaYieldAdapter.sol:27-40`).

**This is integration, not *new* cryptography, with one caveat.** The thesis holds for
the **settlement rail** (T0.2's merged adjudicator + T6.x), which is reuse plus a
patched `go-nitro`: six commits change the vendored `src/nitro` contracts (EIP-1271
participants, a guarantee-check fix, measured-balance deposit crediting, an asset
allow-list, a reentrancy lock, `SafeERC20` payout), so this is a patched adjudicator,
not a vanilla deploy (`armada-nitro@55939ef:src/ArmadaNitroAdjudicator.sol`). It does
not hold for the shielded **pool (T0.0) and its circuits (T0.1)**: those are Armada's
own pool and circuits, its own implementation of the Railgun design, spec-compatible,
and therefore audit-critical crypto-engineering (of a well-understood, de-risked
design, not novel cryptography). The hard cryptographic assets break down as:

- **T0.0 / T0.1** — the shielded pool and its circuits, Armada's own implementation of
  the Railgun design (ADR-0014), with the pool already deployed and the circuits'
  Phase-2 ceremony still open (`v1.0.0-ceremony` manifest, every shape `"status":
  "open"`, no beacon). Amounts are hidden in-circuit; the SNARK enforces
  value-conservation and range, so *$50 and $5M are equally opaque* — no Tornado-style
  denomination buckets.
- **T0.2** — `ArmadaNitroAdjudicator`, a patched `go-nitro` `NitroAdjudicator` /
  `ForceMove` / `MultiAssetHolder`, deployed and carrying the deposit/payout role too
  (see above).

Against that, the **remaining v1 surface** is enumerable:

| Item | Status | id |
|---|---|---|
| Shielded pool + circuits (Armada's own implementation of the Railgun design; audit-critical) | built / Phase-2 ceremony open | **T0.0, T0.1** |
| Nitro adjudicator with the merged deposit/payout boundary | built | **T0.2** |
| Proof path — the nimbus verified proxy over execution and beacon data | in progress (upstream, unintegrated) | **T1.0, T1.1** |
| Fixturenet patch for the proxy's genesis loader | not built | **T1.2** |
| Posted-price venue + quote/settle app + fee-split | not built | **T4.0, T4.1, T4.3** |
| Wallet bits (note-scanner, settlement client, watchtower, transport, proving) | partly built | **T6.1–T6.7** |
| Anonymity-set strategy (process, not code) | strategy | **T0.7** |

The only *novel* cryptography anywhere in the plan is **T0.6** (native channelized
commitment, "fork-lite"). It is not built and, if greenlit, re-runs the Phase-2
ceremony over Armada's own circuits (ADR-0003). v1 needs **no novel cryptography and no
new consensus/ordering** (ADR-0007) — the pool and circuits (T0.0/T0.1) reimplement a
known design rather than invent one — and because a posted price has nothing to
front-run, the fair-ordering stack (T3) stays a plan.

## 4.2 How the top quality goals are met

Each §10 quality goal maps to a concrete approach realized by named items:

| Quality goal (§10) | Approach | Realized by |
|---|---|---|
| **Non-custody** — funds never held by a third party; failure = halt, not loss | Value lives in shielded notes or in `ArmadaNitroAdjudicator` escrow allocated only by co-signed state; every position is unilaterally **force-closable** to its last co-signed state via the ForceMove dispute game (`challenge`→`checkpoint`→`conclude`) | T0.2; watchtower response T6.3 (not built; `armada-nitro`'s e2e gate harness has a test-only watchtower) |
| **Read-time privacy** — sync/read without leaking to a public endpoint | The **nimbus verified proxy** verifies `eth_getBlockReceipts` and storage proofs client-side against a beacon-anchored header, so a client never trusts a public RPC's answer unchecked; `watcher-ts` is not the proof path — it is an optional, unverified Postgres cache | proxy T1.0/T1.1 (in progress, unintegrated); anonymity set T0.7 |
| **Mobile-first** — a phone is a full non-custodial peer | One TypeScript client stack on four targets — Linux server, browser, iOS, Android (`armada-client` `docs/SPEC.md`) — with native `libverifproxy.a` on server and mobile and a WASM build in the browser; mobile has no upstream iOS/Android build yet | client stack T6.5 (not built); proxy build T1.0 |
| **Clearing at a fair, un-front-runnable price** | **RFQ posted-price** dealer market: Armada posts both sides take-it-or-leave-it and reposts if wrong — a posted price has nothing to front-run; this is a plan, not built | posted-price venue T4.0/T4.1/T4.3 (not built); solver T4.2 (not built) |

Amounts are public only at the deposit/payout boundary inside `ArmadaNitroAdjudicator`
today. Hiding amounts in-play is the not-built T0.6 upgrade.

## 4.3 Delivery status

The registry in §5 states what is built, in progress, or a plan. Read as groups:

- **Built or partly built today.** The pool, circuits (pending the Phase-2 ceremony),
  and the merged adjudicator (T0.0–T0.2); CCTP and Aave-v4 yield adapters (T5.0a,
  T5.0b); governance and revenue contracts (`ArmadaGovernor`, `ArmadaTreasuryGov`,
  `TreasurySteward`, `ArmadaToken`, `ArmadaCrowdfund`, `ArmadaWindDown`,
  `ArmadaRedemption`); the `armada-interface` web client, note-scanner, and browser
  prover (T6.0–T6.2, T6.6, T6.7).
- **In progress, unintegrated.** The nimbus verified proxy (T1.0, T1.1) exists
  upstream in `nimbus-eth1` but is not wired into any Armada repo; the fixturenet patch
  for its genesis loader (T1.2) is not built; the selective-disclosure circuit
  (armada-circuits branch `disclosure-circuit`, armada-sdk branch
  `selective-disclosure`) is on unmerged branches.
- **Plan, no code.** The posted-price venue and quote/settle app (T4), the ordering
  stack (T3), Swaps (T5.0c), the watchtower client (T6.3), mobile transport (T6.5),
  native mobile proving (T6.6's native half), and the economic-security contracts for a
  later federated venue (T0.4, T0.5, T2.4).

### Walking skeleton — the first integration target

Before hardening any single tier, stand up **one thin end-to-end slice** on a
fixturenet that exercises every boundary at least once:

```
shield → depositFromNote → trivial ForceMove settle → SHIELD payout → scan
  T0.0        T0.2                T0.2 / T4.1 stub         T0.2        T6.1
```

Concretely, the slice shields a known value into the pool (T0.0), then unshields-in
through `depositFromNote` to fund a Nitro channel (T0.2). It drives a **trivial**
ForceMove app — a single co-signed fill, standing in for the real T4.1 quote/settle —
to a finalized outcome, pays that outcome back out into fresh notes via the SHIELD
payout (T0.2), and scans and confirms balances on-device against the nimbus verified
proxy (T6.1 over a T1.0/T1.1 proof, once integrated). This exercises the
*notes-in-Nitro-notes-out* thesis end-to-end with the smallest possible game.

### Critical path through the dependency DAG

```
        T0.0 ─┐
              ├─► T0.2 ──► T1.0/T1.1 ──► T1.2 ─┐
        T0.1 ─┘                               ├─► T4 (venue)  ─► T5
                                               └─► T6 (wallet) ─► walking skeleton
```

- **Already done:** T0.0, T0.1 (pool and circuits, Phase-2 ceremony pending) and T0.2
  (adjudicator with the merged deposit/payout boundary) are built. Nothing above T0
  needed to wait for them; they already exist.
- **Long pole:** integrating the nimbus verified proxy (T1.0, T1.1) into an Armada
  client and writing the fixturenet genesis patch (T1.2). Every read-side consumer
  (client reads, future T4 venue, T6 wallet) assumes a verified view of chain state,
  and today no client code calls the proxy.
- **Also unbuilt:** the T4 venue and T6's watchtower, mobile transport, and native
  proving. These sit behind the T1 proof path for any read they need, but do not
  block on each other.

---

Cross-refs: quality goals → §10; runtime flows for these slices → §6; deployment of the
fixturenet + peers → §7; recurring concepts (privacy, settlement, transport, metering)
→ §8; long-pole/maturity risks → §11; item ids/status/code → §5; decisions → §9.
