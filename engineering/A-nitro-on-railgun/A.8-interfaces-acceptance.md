# A.8 — Interfaces (ICD) & Acceptance

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** the **A-level ICD** (what A exposes to B/C/D, what A consumes) and the **standing acceptance harness** (walking skeleton + adversarial suite). Rolls up the per-item acceptance of T0.0/T0.1/T0.2/**T0.3**/T0.7/T6.2/T6.3.

---

## A.8.1 Goal

Be the **integration/verification hub** for work package A: one place that (a) consolidates every interface A publishes to its siblings B, C, and D and every interface A consumes from them, and (b) defines the executable proof that *notes-in → normal Nitro → notes-out* holds end-to-end and survives an adversary. The construction A verifies here is the one A.0.2 draws and A.1.4 pins; this doc names the wire it exposes and the harness that keeps it honest, and defers every internal to the owning A.x spec.

Per [ADR-0012](../09-architecture-decisions.md#adr-0012), A binds to B/C/D **only** through the interfaces in A.8.5 — the interface *is* the boundary. Per rule 5, the A.8.6 walking skeleton is the programme's **first integration target** and thereafter the standing regression harness against which every A item is deepened.

## A.8.2 Boundary

- **In scope:** the consolidated contract/event/API surface A publishes; the freshness contract A depends on from B; the fixturenet acceptance slice and adversarial cases; the roll-up acceptance matrix.
- **Out of scope (owned elsewhere):** the deposit/payout ABI internals (→ A.5), pool config (→ A.2), ceremony artifacts (→ A.3), adjudicator/app choice (→ A.4), client + watchtower loop (→ A.6), POI/anon-set (→ A.7). This doc *cites* them; it never re-specifies them.
- **Phase-0 assumption.** The **licensing gate** ([A.1.10](./A.1-reuse-inventory.md), [A.0.5](./A.0-overview.md) gate 1) is **resolved via clean-room** ([ADR-0014](../09-architecture-decisions.md#adr-0014)): the pool (T0.0, → A.2) and circuits (T0.1, → A.3) are **clean-room reimplemented**, so the skeleton runs against our clean-room pool — no Railgun grant/relicense needed. The ICD shapes below are stable regardless.

## A.8.3 Reuse inventory (cite A.1)

The exposed surface is mostly *reused* ABI, published under A's control:

| Exposed piece | Reused from (A.1 §, pinned commit) | Net-new? |
|---|---|---|
| Pool `shield`/`transact`/`Shield`/`Transact`/`Nullified` events | [A.1.1](./A.1-reuse-inventory.md) `RailgunLogic.sol` L57–77 (Railgun ref @HEAD — **pin before build**) | **net-new (clean-room, spec-compatible; A.2)** |
| Circuit `wasm`/`zkey` artifacts (91 combos) | [A.1.2](./A.1-reuse-inventory.md) `circuits-v2` reference (→ A.3) | **net-new (clean-room circuits; A.3)** |
| Adjudicator `challenge`/`checkpoint`/`conclude`, `MultiAssetHolder.deposit`, `concludeAndTransferAllAssets`, exit-format `Outcome` | [A.1.3](./A.1-reuse-inventory.md) go-nitro **@435eb2b** | reused as-is |
| `ChallengeRegistered`/`Checkpointed`/`ChallengeCleared` events | go-nitro `ForceMove.sol` L39–119 (NitroScout) | reused |
| Voucher `{ChannelId, Amount, Signature}` + `Hash()` | [A.1.3](./A.1-reuse-inventory.md) `payments/vouchers.go` L23–52 | reused |
| Channel-lifecycle node API (fund/defund/pay) | [A.1.5](./A.1-reuse-inventory.md) ts-nitro **@884d616** `node.ts` L69–190 | reused (+delta, A.1.9 §6) |
| `Deposit`/`Payout` events; T0.3 fund/payout entrypoints; app registry; POI-gate | [A.1.4](./A.1-reuse-inventory.md), [A.1.9](./A.1-reuse-inventory.md) §1 (→ A.5) | **net-new** |

## A.8.4 Net-new delta (what A.8 itself adds)

A.8 authors no runtime code except the **test-only** pieces of the harness:

1. The **consolidated ICD document** (A.8.5) — the single addressable contract every sibling WP binds to, so B/C/D never read A's internals.
2. The **trivial single-asset ForceMove app** (HashLockedSwap-grade stand-in for C's T4.1 quote/settle app) used solely by the skeleton — cited from [A.1.3](./A.1-reuse-inventory.md) `HashLockedSwap.sol` L33–83; **not** shipped as product (A.1.8 §2: no multi-asset ForceMove app ships).
3. The **fixturenet acceptance harness + adversarial suite** (A.8.6) — the standing regression that gates every A merge.

Everything the harness *drives* is owned by an A.x item; A.8 only orchestrates and asserts.

## A.8.5 Interfaces (ICD)

### A.8.5.1 What A exposes (to B, C, D)

All addresses/roots are published in a per-fixturenet/per-deployment manifest; ABIs are frozen at the pinned commits in A.8.3. Authoritative field-by-field ABI lives in the owning A.x (mostly A.5); this is the consolidated index.

**(a) Deposit/payout contract — address + ABI (→ A.5).**
The net-new T0.3 contract is A's public-side ERC20 counterparty of unshield/shield and of channel deposit/conclude ([A.1.4](./A.1-reuse-inventory.md)). Consumers get: its deployed **address** (the value set as `unshieldPreimage.npk = bytes32(uint160(T0.3))` for deposit-in, and named as the external destination for payout-out), and its ABI. Full ABI authoritative in **A.5**; the boundary methods are the fund path (`IERC20.approve` + `MultiAssetHolder.deposit(asset, channelId, expectedHeld, amount)`) and the payout path (`shield([ShieldRequest…])` on receipt of the outcome transfer). `channelId = NitroUtils.getChannelId(fixedPart)`.

**(b) Boundary events — `Deposit` / `Payout`.**
Emitted by T0.3 so B (T1.2) can index the Railgun↔Nitro crossing without decoding pool internals:
- `Deposit(bytes32 indexed channelId, address asset, uint256 amount, bytes32 consumedNullifier)` — one per unshield-in that escrows into a channel; `consumedNullifier` ties the crossing to the pool `Nullified` record (replay-defence anchor, A.8.6.3).
- `Payout(bytes32 indexed channelId, bytes32[] newCommitments)` — one per outcome re-shield; `newCommitments` are the fresh commitment hashes B/T6.1 scan for.

These sit alongside the reused adjudicator events A also surfaces: `ChallengeRegistered(channelId, finalizesAt, proof, candidate)`, `Checkpointed`, `ChallengeCleared` (`ForceMove.sol` L39–119), and the pool `Shield`/`Transact`/`Nullified` (A.8.5.1(f)). Field ABI: A.5 (`Deposit`/`Payout`), A.4 (adjudicator), A.2 (pool).

**(c) Channel-lifecycle API (→ A.6).**
The settlement-client (T6.2) surface siblings drive to move value through a channel. Reuses ts-nitro `node.ts` fund/defund/pay ([A.1.5](./A.1-reuse-inventory.md)) plus the net-new in-browser dispute surface ([A.1.9](./A.1-reuse-inventory.md) §6):
- **open / fund-from-note** — unshield a note into T0.3 and `deposit` into a freshly-derived `channelId` (directfund/virtualfund shape).
- **propose / accept** — exchange signed `VariablePart`s over the app; signing domain = `NitroUtils.hashState(fixedPart, variablePart)` (`ForceMove.sol` L236–268) — C's T4.1 app and D's client MUST match it.
- **cooperative close** — co-sign a final state; `concludeAndTransferAllAssets` routes the outcome to T0.3 for re-shield, no on-chain dispute.
- **force-close** — `challenge(FixedPart, proof, candidate, challengerSig)` starting the `finalizesAt` window (A.8.6.3 exit).
- **respond** — `checkpoint(FixedPart, proof, candidate)` overwrites `turnNumRecord` with a strictly-higher supported turn *without* finalizing (`ForceMove.sol` L88–119) — the primitive T6.3 uses. Detail: A.6.

**(d) ForceMove app registry (→ A.5 / A.4).**
The set of `appDefinition` addresses T0.3 accepts for a funded channel (`NitroAdjudicator.sol` L193–207 → `IForceMoveApp.stateIsSupported`, [A.1.4](./A.1-reuse-inventory.md)). **C registers its T4.1 quote/settle app here**; the skeleton registers only the trivial stand-in (A.8.4 §2). The registry is the extension point by which C's clearing rides A's rail without A knowing the game — A validates only that the channel's `appDefinition` is registered and its outcome is exit-format-valid.

**(e) Voucher format (→ A.6, consumed by B·T2.1 and C).**
`Voucher{ChannelId, Amount, Signature}` with `Hash() = keccak256(abi(Destination, Uint256))` and `Sign/RecoverSigner` ([A.1.3](./A.1-reuse-inventory.md) `vouchers.go` L23–52). The incremental in-channel micropayment primitive: B meters watcher service against it (T2.1), C meters settlement against it. Redeemed by concluding the channel at the largest signed amount.

**(f) Pool commitment-insert / nullifier-spend event ABI + circuit artifacts (→ A.2, A.3; consumed by B·T1.2, T6.1, T6.6).**
- Events `Shield`, `Transact`, `Nullified` (and `Action`/`Unshield`) at `RailgunLogic.sol` L57–77 ([A.1.1](./A.1-reuse-inventory.md)) — the commitment-insert + nullifier-spend feed B indexes (T1.2) and the note-scanner rebuilds the tree from (T6.1). Tree params: `TREE_DEPTH=16`, `rootHistory` (`Commitments.sol` L28–55).
- Circuit **`wasm` + `zkey` artifacts**, one pair per registered `(nIn,nOut)` combo out of the **91** generated (A.1.8 §3 — 91, not 54; the registered subset is reconciled in A.3). Published from T0.1's ceremony (→ A.3) and consumed by D's mobile prover (T6.6) and any prover B runs. The matching on-chain `verificationKeys[nIn][nOut]` registration is A.2's job.

**(g) POI allow-list policy / root (→ A.7).**
POI is **not** a pool setting — it is an alongside partner system ([A.1.7](./A.1-reuse-inventory.md), [A.0.5](./A.0-overview.md) gate 3). A exposes the **allow-list root** (chosen list-provider datasets + standby period) and the **gate policy**: T0.3 gated-entry enforces a valid POI proof at the *settlement layer* before a deposit is accepted. The POI node stack is a separate dependency, not redeployable from the four scoped repos. Consumers treat the root as opaque policy input. Detail: A.7.

### A.8.5.2 What A consumes (from B; hosted by D)

A depends on exactly one sibling interface at runtime, plus D as host:

- **B's proof-carrying feed** — `getStorageAt → {value, proof}` (B's T2.0 surface). A's settlement client (T6.2) and watchtower (T6.3) read adjudicator `statusOf[channelId]` / holdings and pool roots through it rather than trusting a bare RPC, so a light client verifies chain state under the anonymity-set constraints.
- **B's freshness signal — the T6.3 gate.** The per-contract **head-cursor freshness signal** B publishes is a **hard gate on T6.3**: the watchtower MUST NOT decide it holds a higher-turn state (and MUST NOT sleep through a `finalizesAt` window) on stale data. Registry: T6.3 is *"gated on T2 feed freshness."* If the feed is stale beyond the safety margin relative to `finalizesAt`, T6.3 escalates/alerts rather than assuming safety. Contract detail owned by B; A's obligation is to **block** on it. See A.6.
- **D hosts A** — the settlement client (T6.2), watchtower loop (T6.3), and note-scanner bridge run inside D's wallet/app shell (T6.0/T6.5). A publishes the client edges; D provides transport + lifecycle.

C consumes A (settlement) but does so *through* A.8.5.1(a)/(c)/(d)/(e) — from A's side that is an exposed surface, not a dependency.

## A.8.6 Acceptance / verification

### A.8.6.1 The walking skeleton (standing acceptance harness — expands A.0.4)

A thin end-to-end slice on a laconic fixturenet, exercising **every A boundary once**, using the trivial ForceMove app (A.8.4 §2). It is A's first integration target and thereafter the regression harness every A item is deepened against.

```
shield → transact()/unshield-in → T0.3.deposit → MultiAssetHolder.deposit
      → trivial ForceMove settle (off-chain) → concludeAndTransferAllAssets
      → T0.3 receives → shield()/payout → scan
```

**Preconditions (Phase-0 gate, A.8.2):** clean-room pool deployed with `changeFee(0,0,0)` and vkeys registered for the chosen combo subset (A.2/A.3); adjudicator deployed (A.4); T0.3 deployed and its app registry seeded with the trivial app (A.5); two funded client identities (A.6).

| # | Step | Emits / asserts | Exercises (A.x) |
|---|---|---|---|
| 1 | **shield** a note into the pool | `Shield`; commitment in tree | A.2, A.7 (payments primitive A.1.7) |
| 2 | **transact() unshield-in** with `unshieldPreimage.npk = T0.3` | `Transact`, `Nullified`; ERC20 lands in T0.3 via `transferTokenOut` | A.5 (deposit-in), A.2 |
| 3 | T0.3 **`approve` + `MultiAssetHolder.deposit(asset,channelId,expectedHeld,amount)`** | `Deposit(channelId,asset,amount,consumedNullifier)`; `holdings[asset][channelId]` credited | A.5, A.4 |
| 4 | **trivial ForceMove settle** off-chain: parties co-sign states to a final outcome | supported final `VariablePart` (signing domain `hashState`) | A.4 (app), A.6 (client) |
| 5 | **`concludeAndTransferAllAssets(FixedPart, candidate)`** → `transferAllAssets` routes outcome to T0.3 (external destination) | channel finalized; ERC20 transferred to T0.3 | A.4, A.5 |
| 6 | T0.3 **`shield([ShieldRequest…])`** mints fresh notes to beneficiaries | `Shield`, `Payout(channelId,newCommitments[])` | A.5 |
| 7 | **scan**: note-scanner rebuilds tree, finds `newCommitments`; client confirms balance | new commitments discovered; nullifier of step 2 marked spent | A.6 (client), B·T6.1 (consumer) |

Pass = a note's value made a full round-trip through a Nitro channel and re-emerged as fresh, scannable notes, with `sumIn == sumOut` across the boundary and no orphaned holdings. This proves *notes-in-Nitro-notes-out* before any item is deepened.

### A.8.6.2 A-level acceptance (per-item roll-up)

Each A item passes its own acceptance in its A.x; the harness asserts they compose:

| Item | Acceptance criterion (verified by) | Spec |
|---|---|---|
| **T0.0** pool | Clean-room pool with `changeFee(0,0,0)`; a shield+unshield round-trips; `snarkSafetyVector` reproduced so `transact` does not revert | A.2 |
| **T0.1** setup | Phase-2 zkey per registered combo; a JoinSplit proof verifies against the registered vkey; artifacts published (A.8.5.1(f)) | A.3 |
| **T0.2** adjudicator | go-nitro @435eb2b deployed unmodified; deposit→conclude→transfer path lands funds at an external destination | A.4 |
| **T0.3** deposit/payout | Deposit-in escrows the exact unshielded amount into `channelId`; payout-out re-shields the full outcome; `Deposit`/`Payout` emitted; rejects unregistered `appDefinition` and invalid POI | A.5 |
| **T0.7** anon-set | Import-bridge hop (unshield-Railgun → shield-ours) works; POI gate honoured at deposit | A.7 |
| **T6.2** client | Drives the full A.8.6.1 lifecycle; dispute surface (challenge/checkpoint) present, not just fund/defund/pay | A.6 |
| **T6.3** watchtower | Auto watch→compare-turn→checkpoint loop fires before `finalizesAt`; **blocks on B freshness** (A.8.5.2); always-on liveness | A.6 |

### A.8.6.3 Adversarial suite (must pass alongside the skeleton)

| Case | Scenario | Expected result | Exercises |
|---|---|---|---|
| **Stale-state force-close defeated** | Adversary `challenge`s with an old (lower-turn) state; `ChallengeRegistered(finalizesAt)` fires | T6.3 detects the event via B's fresh feed, finds it holds a strictly-higher supported state, submits `checkpoint` **before `finalizesAt`** → `ChallengeCleared`, channel not finalized on the stale state | A.6 (loop), A.4 (`ForceMove.sol` L88–119), A.8.5.2 (freshness gate) |
| **Replayed nullifier rejected** | Attempt a second deposit-in reusing an already-consumed nullifier (`consumedNullifier` seen in a prior `Deposit`) | Pool `Nullified` set rejects the re-spend in-circuit; the crossing never produces a second `Deposit`; T0.3 credits nothing | A.2 (nullifier set, `Commitments.sol`), A.5 |
| **Unresponsive-counterparty exit** | Counterparty goes silent mid-channel | Party `challenge`s with the latest supported state, waits out `finalizesAt` with no valid counter, then `concludeAndTransferAllAssets` → `Payout` re-shields its rightful outcome; funds are never stranded | A.4 (challenge/conclude), A.6, A.5 |

These are the go-nitro maturity long-poles (ADR-0004, [A.0.5](./A.0-overview.md) gates 4–5) made executable: dispute wiring and watchtower response are net-new (A.1.9 §3, §6), so the suite is the proof they actually work, not just the happy path.

## A.8.7 Risks / open

- **Watchtower liveness vs freshness (gate 5).** The stale-force-close defence is only as good as B's freshness signal and T6.3's always-on node; a stale feed or a dead node during a `finalizesAt` window is a fund-loss path. A.8.5.2 makes the gate mandatory; residual risk owned in A.6.
- **App-registry trust surface.** A validates `appDefinition` membership + exit-format validity, not game semantics; a buggy registered app (C's T4.1) can still produce a valid-but-wrong outcome. Boundary stays clean (A.4/A.5); correctness of the game is C's acceptance, not A's.
- **No multi-asset ForceMove app in the skeleton.** The harness settles single-asset (trivial app). ETH-in/USDC-out atomicity (net-new, A.1.9 §2) needs its own acceptance once C's app lands; the skeleton does **not** prove multi-asset settlement (A.1.8 §2).
- **Licensing (Phase-0) — RESOLVED via clean-room ([ADR-0014](../09-architecture-decisions.md#adr-0014)).** The harness runs against our **clean-room** pool/circuits ([A.1.10](./A.1-reuse-inventory.md)/[A.0.5](./A.0-overview.md) gate 1, A.2/A.3) — no Railgun grant needed. Residual: those clean-room builds must ship before steps 1–2 and 6 run.
- **Railgun commit unpinned.** All pool/circuit citations resolve at `master`/`main` HEAD; the manifest MUST pin a Railgun SHA before the skeleton is treated as reproducible (A.0.6, A.1.10).

→ Owning specs: [A.2](./A.2-pool-deployment.md) · [A.3](./A.3-trusted-setup.md) · [A.4](./A.4-adjudicator-integration.md) · [A.5](./A.5-deposit-payout-contract.md) · [A.6](./A.6-settlement-client-watchtower.md) · [A.7](./A.7-anonymity-set.md) · [A.9](./A.9-native-commitment.md). Reuse citations: [A.1](./A.1-reuse-inventory.md).
