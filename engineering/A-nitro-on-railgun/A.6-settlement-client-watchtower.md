# A.6 — Settlement client & self-watchtower

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T6.2 settlement client · T6.3 self-watchtower

---

## A.6.1 Goal

Deliver the **client edge of A**: the software that a party runs to (T6.2) drive a settlement channel through its whole life — fund it from a deposit, exchange off-chain ForceMove states and vouchers, and defund it back to a payout — and (T6.3) keep that channel **safe while it is open**, by automatically defeating a counterparty who force-closes on a stale state. Both are client-side; neither touches pool or adjudicator internals, only their ABIs (→ A.5).

Per [ADR-0012](../09-architecture-decisions.md#adr-0012) this is integration-plus-glue: the channel engine, the P2P substrate, the voucher primitive, and the chain event/tx machinery are all reused at pinned commits (A.1.5, A.1.6, A.1.3). **T6.3's auto challenge-response logic is itself trivial to write** — the primitives already exist (`ForceMove.checkpoint`, the `ChallengeRegistered` feed, and the Checkpoint tx build; A.1.6, A.1.3). **The real net-new work is the *tooling and liveness* that make that logic runnable in a browser:** a **thin in-browser dispute API** on the ts-nitro node (which stock ts-nitro omits) and an **always-on watch/relay** so a challenge is *seen* inside its window (A.1.6, A.1.9). This client is hosted by **D** (T6.0/T6.5) and consumes **B**'s feeds (T2) for state freshness.

## A.6.2 Boundary

**In scope (T6.2):**
- Run the ts-nitro in-browser node as the channel engine (fund/defund/pay) over `@cerc-io/peer` (A.1.5).
- Drive the A.5 deposit/payout lifecycle from the client: request deposit-in, build/sign ForceMove `VariablePart`s, exchange vouchers, request payout-out.
- **Submit-on-behalf** path for write-side origin privacy: on-chain transactions (deposit escrow, conclude, challenge, checkpoint) are relayed by a funded submitter so the signing party stays origin-private (A.1.5 keeper shape).
- The **net-new dispute API** surfaced on the node: `challenge`, `checkpoint`, `getSupportedState` — absent from stock ts-nitro (A.1.9 delta 6).

**In scope (T6.3):**
- The automatic watchtower loop: subscribe `ChallengeRegistered` → compare turnNums → `checkpoint` before `FinalizesAt`, else fall back to `directdefund` (A.1.6).
- Always-on liveness: a persistent relay connection and background watch so the challenge is *seen* inside its window.

**Out of scope / consumed:** the deposit/payout contract, outcome encoding and the app registry are **A.5** (T0.3); the adjudicator (`ForceMove`/`MultiAssetHolder`) and its dispute primitives are reused **A.4** (T0.2); note-scanning of the pool is **B** (T6.1); the freshness signal the watchtower gates on is **B**'s per-contract head-cursor (T2, → A.6.5); the wallet host, transport, and mobile proving are **D** (T6.0/T6.5/T6.6). Terminology per [ADR-0010](../09-architecture-decisions.md#adr-0010): this is the **deposit/payout** lifecycle, a **boundary** with B, a client of tier **T6** — never adapter/seam/layer.

## A.6.3 Reuse inventory (cite A.1 + pinned commits)

Pinned: ts-nitro **@884d616**, go-nitro **@435eb2b**, mobymask **@2329198**. Railgun is **unpinned — pin before build** (A.1.10). Cites resolve to A.1; do not re-document.

| Reused piece | Citation (via A.1) | What it gives T6.2/T6.3 |
|---|---|---|
| In-browser Nitro node (fund/defund/pay) | A.1.5 · ts-nitro @884d616 `packages/nitro-node/src/node/node.ts` L69–190 | The channel engine T6.2 wraps. **Dispute surface is thin — no `challenge`/`checkpoint`** (A.1.9 → net-new, A.6.4). |
| Browser P2P over `@cerc-io/peer` | A.1.5 · ts-nitro `…/p2p-message-service/service.ts` L16–120, 453–527 | `/nitro/msg/1.0.0` + `/nitro/peerinfo/1.0.0`; relay/WebRTC dial (no inbound listen) — the transport the always-on liveness (A.6.4) keeps warm. |
| Submit-on-behalf keeper | A.1.5 · mobymask @2329198 `packages/server/index.ts` L48–58; `Delegatable.sol` `execute`/`_msgSender` | The **broadcaster analog**: a funded relayer submits on behalf of a signer who stays origin-private — reusable *shape* for T6.2 write-side privacy. |
| Vouchers | A.1.3 · go-nitro @435eb2b `payments/vouchers.go` L23–52 | Incremental in-channel micropayment primitive — the in-channel payment T6.2 sends/redeems (shared with C metering & settlement). |
| Event source `ChallengeRegistered`/`ChallengeCleared` | A.1.6 · go-nitro `node/engine/chainservice/eth_chainservice.go` L494–539; `chainservice.go` L86–160 (`ChallengeRegisteredEvent{candidate,sigs,FinalizesAt,IsInitiatedByMe}`) | The feed T6.3 subscribes to; `FinalizesAt` is the deadline, `candidate` carries the challenger's turnNum. |
| Tx submit `Challenge`/`Checkpoint` | A.1.6 · `eth_chainservice.go` L376–386; payloads `protocols/interfaces.go` L59–114 | The tx builders T6.3 invokes to push a higher-turn state on-chain. |
| `checkpoint` (higher-turn, no finalize) / `challenge` / `conclude` | A.1.3 · go-nitro `ForceMove.sol` L88–119 / L39–80 / L126–172 | **`checkpoint` is the exact primitive T6.3 uses** to reset `FinalizesAt` with a later supported state and defeat a stale close. |
| `recoverVariablePart` signing domain | A.1.3 · `ForceMove.sol` L236–268 | State sigs = `NitroUtils.hashState(fixedPart, variablePart)` — the dispute API (A.6.4) must reproduce this exactly, same as T0.3/A.5. |
| Manual counter-challenge (pattern, not the loop) | A.1.6 · go-nitro `engine.go` L882–915 (`handleCounterChallengeRequest`) | Shows how a checkpoint payload is assembled once triggered — T6.3 automates the *trigger*. |

**A.1.8 corrections honored.** mobymask **@2329198** gives the **submit-on-behalf keeper only** — it is classic delegatable MobyMask (hardhat + react-app + OpenRPC server), **NOT** a browser-peer or watcher. The P2P substrate here is ts-nitro / `@cerc-io/peer`; the browser-peer/watcher lineage lives in `mobymask-v2` / `mobymask-v2-watcher-ts` and is not this commit. Multi-asset settlement *is* supported by the adjudicator (A.1.8 §2 / A.4), so the client's lifecycle is multi-asset-capable even though the ETH-in/USDC-out ForceMove **app** is net-new (A.5).

## A.6.4 Net-new delta (what A actually builds)

Two deltas, both from A.1.9 (items 3 and 6):

**D1 — In-browser dispute API (T6.2).** Stock ts-nitro `node.ts` (L69–190) exposes only `fund`/`defund`/`pay`; the dispute path is not surfaced (A.1.5 / A.1.9 delta 6). Add, on the node:
- `getSupportedState(channelId) → {variablePart, sigs, turnNum}` — the highest-turn state the local store holds with a complete support proof.
- `challenge(channelId)` and `checkpoint(channelId, supportedState)` — build the payload (`protocols/interfaces.go` shape, A.1.6), sign in the `recoverVariablePart` domain (A.1.3), and hand the tx to the submit-on-behalf relayer (A.6.3).
This is a surface over reused primitives, not new crypto — signing/hashing reuse `NitroUtils.hashState`; it exists so the browser client can dispute, which today only the go-nitro server engine can.

**D2 — Automatic watchtower loop + always-on liveness (T6.3).** go-nitro has **no automatic watchtower** (A.1.6): a non-initiator engine auto-creates a `directdefund` (conclude+withdraw) on an adversarial `ChallengeRegistered`, **not** a higher-turn `checkpoint`; only a *manual* `CounterChallengeRequest` reaches checkpoint (`engine.go` L540–612, L882–915). The loop itself is **small and trivial once the tooling (D1) and liveness are correct** — the checkpoint primitive, the `ChallengeRegistered` feed, and the Checkpoint tx build all already exist (A.1.6/A.1.3); what is missing is only the *automatic trigger* wired over them:

```
on ChallengeRegistered(channelId, candidate, FinalizesAt, IsInitiatedByMe):
  if IsInitiatedByMe: return                       # our own close, nothing to defend
  local  = getSupportedState(channelId)            # D1
  onchain = candidate.turnNum
  if local.turnNum > onchain:                      # we hold a strictly later supported state
      checkpoint(channelId, local)                 # D1 → resets FinalizesAt, no finalize (ForceMove L88–119)
  else:                                            # counterparty's state is current/ahead: no fraud
      directdefund(channelId)                      # let the exit proceed correctly (reused engine path)
  # all of the above MUST land before FinalizesAt
```

- **Freshness gate.** "highest locally-held supported state" is only trustworthy if the client is synced. T6.3 **gates the loop on B's per-contract head-cursor freshness signal** (T2, → A.6.5): if the feed is stale, the client cannot assert `local.turnNum` is authoritative and MUST prefer the safe `directdefund` exit rather than a checkpoint that could itself be stale.
- **Always-on liveness (net-new).** The browser node has no background watch loop and no guaranteed relay connection (A.1.6). T6.3 adds a persistent `@cerc-io/peer` relay connection and a background subscriber so `ChallengeRegistered` is *observed within its window*. A missed challenge = a lost channel; this is the **A.0.5 gate 5** liveness requirement. Where a phone cannot stay connected for the full challenge window, the loop MUST be delegable to an always-on submit-on-behalf relayer (A.6.3 keeper shape) that watches and checkpoints for the offline party — this is the primary reason the keeper is reused write-side.

The `checkpoint`-not-`directdefund` choice is the crux: `directdefund` still *exits*, but at the challenger's (possibly stale, adversarial) outcome; `checkpoint` re-establishes the true latest state so the channel finalizes correctly or stays open. The response logic above is **trivial given the tooling** — a turnNum compare plus a `checkpoint` call over primitives that already exist — so the genuinely net-new, load-bearing work is the **tooling (D1's in-browser dispute API) and the always-on liveness** that let it run at all, **not** the primitive and **not** the branch logic.

## A.6.5 Interfaces (ICD)

**Consumed from A.5 (T0.3, sibling in this package):**
- Channel-lifecycle API: `requestDeposit(channelId, asset, amount)` → drives `RailgunSmartWallet.transact` unshield-in then `MultiAssetHolder.deposit`; `requestPayout(channelId, outcome)` → `concludeAndTransferAllAssets` to the payout external destination that re-shields. (A.5 · A.1.4)
- `channelId = NitroUtils.getChannelId(fixedPart)`; the ForceMove **app registry** (`appDefinition`) that T6.2 selects the settlement app from (A.5 · A.1.4). Walking skeleton uses the trivial app (→ A.0.4/A.8); C's quote/settle app registers later.
- Outcome / exit-format encoding (`SingleAssetExit`, `Allocation`) the client assembles for defund (A.5 · A.1.3).

**Consumed from B (feeds, cross-package boundary):**
- The **proof-carrying feed** `getStorageAt → {value, proof}` (T2.0) to read on-chain channel/holdings state with a proof rather than trusting a bare RPC.
- The **per-contract head-cursor freshness signal** (T2, "which T6.3 gates on" per `00-work-packages.md`) — the input to the A.6.4 freshness gate. **T6.3 depends on B (T2 feeds) for freshness.**
- The Nitro **voucher metering interface** (T2.1) — the same voucher primitive (A.6.3) funds B's watcher metering; T6.2 issues/redeems these in-channel.

**Consumed from A.4 (T0.2, sibling):** the adjudicator address + `ForceMove`/`MultiAssetHolder` ABIs — the dispute API (A.6.4 D1) and watch loop (D2) bind to `challenge`/`checkpoint`/`conclude` and the `ChallengeRegistered`/`ChallengeCleared` events there.

**Exposed to D (host) and up through A's ICD (A.8):** the settlement-client channel-lifecycle API (fund/pay/defund + the net-new challenge/checkpoint/getSupportedState); the watchtower as a background service D hosts (or delegates to a keeper). A.8 aggregates what A exposes to B/C/D.

**Exposed to C:** the same lifecycle + voucher surface is what C's yield/exchange clearing drives over A's rail (A.8).

## A.6.6 Acceptance / verification

Verified on a **laconic fixturenet** (→ A.8), reusing the walking-skeleton channel (A.0.4). Two adversarial scenarios plus the happy path:

1. **Watchtower defeats a stale force-close (T6.3, the headline test).**
   - Open a channel, advance it to a supported state at `turnNum = N` (both parties signed).
   - Counterparty issues `challenge` with an earlier state `turnNum = M < N` (a stale/adversarial close).
   - **Expected:** the watchtower observes `ChallengeRegistered` over its persistent relay connection, computes `local.turnNum = N > M`, submits a `checkpoint` with the `N` state **before `FinalizesAt`**, and the challenge is cleared (`ChallengeCleared`); the stale outcome never finalizes. Funds settle at the true latest state.
2. **Unresponsive counterparty still exits correctly (T6.2/T6.3 fallback).**
   - Counterparty goes silent (no newer state exists; we legitimately want to close).
   - We `challenge` with our latest supported state; counterparty never responds.
   - **Expected:** no higher-turn state appears; after `FinalizesAt` the channel concludes at our state and `concludeAndTransferAllAssets` routes the payout to A.5's payout destination which re-shields. The exit completes without counterparty cooperation. (When *we* are the offline party, the delegated keeper performs this.)
3. **Freshness-gated safety.**
   - Force B's head-cursor stale (feed lag). A `ChallengeRegistered` arrives.
   - **Expected:** the loop refuses to assert authority on a possibly-stale `local`; it takes the safe `directdefund` path (exit, not checkpoint) rather than risk checkpointing an outdated state. Exit is correct-if-conservative.
4. **Lifecycle happy path (T6.2).** fund (deposit-in) → exchange states + at least one voucher → cooperative defund (payout-out) → fresh notes scan (B/T6.1). Proves the client drives the full A.5 boundary once.

**Verification tactics:** drive scenarios from a test harness against the fixturenet adjudicator; assert on emitted `ChallengeRegistered`/`ChallengeCleared` and final `holdings`/payout; assert the checkpoint tx lands strictly before `FinalizesAt` (timing is the failure mode). Reuse go-nitro's dispute test fixtures as oracles for turnNum comparison. No project-wide suite here — scoped scenario tests only (main integration lands in A.8).

## A.6.7 Risks / open

- **Liveness is the crux (A.0.5 gate 5).** A phone that sleeps through the full challenge window loses the channel. Mitigation is the persistent-relay + delegated-keeper path (A.6.4); the residual risk is trusting/funding that keeper. The keeper sees *that* a party disputes but stays origin-private on submission (A.6.3) — it does **not** learn channel contents beyond what it relays.
- **ts-nitro dispute-API maturity.** The net-new `challenge`/`checkpoint` surface (D1) must exactly reproduce go-nitro's signing/support-proof semantics (`ForceMove` L88–119, L236–268). Divergence = an invalid checkpoint that the adjudicator rejects, silently losing the defense. Cross-test the browser payload against the go-nitro server engine.
- **Freshness dependency on B.** T6.3 correctness is only as good as B's freshness signal (T2). If B under-reports staleness, the loop could checkpoint a stale state; the conservative `directdefund` fallback (A.6.4) bounds the blast radius to a correct-but-suboptimal exit.
- **Railgun unpinned** (A.1.10): the payout re-shield the lifecycle drives crosses into unpinned Railgun code — **pin before build** (A.5/A.2).
- **Multi-asset app gap (A.1.8 §2 / A.1.9 delta 2):** the client lifecycle is multi-asset-ready, but until C's multi-asset ForceMove app ships, T6.2 exercises only the single-asset trivial app (A.4/A.5). Not a client-side blocker.

→ Siblings: [`A.1`](./A.1-reuse-inventory.md) (§A.1.5, A.1.6, A.1.3, A.1.8, A.1.9) · [`A.4`](./A.4-adjudicator-integration.md) (adjudicator/dispute primitives) · [`A.5`](./A.5-deposit-payout-contract.md) (channel-lifecycle API, outcome encoding, app registry) · [`A.8`](./A.8-interfaces-acceptance.md) (aggregate ICD + fixturenet). Baseline: [ADR-0004](../09-architecture-decisions.md#adr-0004), [ADR-0010](../09-architecture-decisions.md#adr-0010), [ADR-0012](../09-architecture-decisions.md#adr-0012).
