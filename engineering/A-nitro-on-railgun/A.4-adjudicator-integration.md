# A.4 — Adjudicator Integration

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** **T0.2** — Nitro adjudicator (`go-nitro` NitroAdjudicator / ForceMove / MultiAssetHolder + exit-format Outcome)

---

## A.4.1 Goal

Stand up the on-chain adjudicator that anchors Armada settlement: a redeployed `go-nitro` `NitroAdjudicator`, comprising its `ForceMove` dispute machine, its `MultiAssetHolder` escrow, and the exit-format `Outcome` payout path. It runs on the laconic fixturenet so channels funded from a shielded note can settle off-chain via ForceMove and finalize back to the deposit/payout boundary (T0.3; see [A.5](./A.5-deposit-payout-contract.md)). Per [ADR-0004](../09-architecture-decisions.md#adr-0004) and [ADR-0012](../09-architecture-decisions.md#adr-0012), this is integration, not new cryptography. The adjudicator is reused as-is at its pinned commit. The only genuinely net-new code T0.2 introduces is a multi-asset ForceMove settlement app (ETH-in/USDC-out) and the end-to-end wiring of the dispute paths (challenge/checkpoint/conclude) through the node.

The walking-skeleton target (A.0.4) is to fund a channel via `MultiAssetHolder.deposit`, run a trivial single-asset settlement app off-chain, then call `concludeAndTransferAllAssets` to the T0.3 external destination. This proves *notes-in-Nitro-notes-out* before the real multi-asset app or the auto-watchtower (T6.3; see [A.6](./A.6-settlement-client-watchtower.md)) is deepened.

## A.4.2 Boundary

T0.2 owns the adjudicator contract set and its dispute wiring, and nothing on either private side of it:

- **In scope:** deploy `NitroAdjudicator` (its `ForceMove`, `MultiAssetHolder`, and exit-format pieces) at commit @435eb2b; author the multi-asset settlement app that binds to it; drive challenge, checkpoint, and conclude end-to-end from the node; and expose the adjudicator address, the `appDefinition` and app-registry surface, and the signing domain (see A.4.5) that T0.3 (A.5) and C's quote/settle app (T4.1) consume.
- **Out of scope (cited baseline / sibling):** the shielded pool (clean-room per A.2) and its shield/unshield ABIs ([A.1.1](./A.1-reuse-inventory.md#a11-shielded-pool-t00--clean-room-reference-spec), A.2); the net-new deposit/payout contract itself (T0.3, A.5), for which T0.2 only defines the ABIs that T0.3 calls into and finalizes out of; the settlement client and auto-watchtower (T6.2/T6.3, A.6); and C's real quote/settle app variant (T4.1), for which T0.2 delivers the reference multi-asset app and registry mechanism it plugs into.
- The Railgun/Nitro boundary ([ADR-0010](../09-architecture-decisions.md#adr-0010)) is realized in T0.3. Here `RelayAdapt` and `Cookbook` name Railgun's own recipe layer, not this boundary. T0.2 is the public-side channel counterparty that T0.3 deposits into and concludes from.
- **Phase-0 note (licensing gate — resolved via clean-room, [ADR-0014](../09-architecture-decisions.md#adr-0014)):** T0.2 was never licence-blocked; go-nitro is Apache/MIT-lineage OSS and freely redeployable. The pool (T0.0) and circuits (T0.1) that share its fixturenet were `UNLICENSED`, but are now clean-room reimplemented (A.2/A.3), so that gate is resolved. The adjudicator still comes up independently for its own acceptance (A.4.6); the joined skeleton (deposit-from-note) now waits only on those clean-room builds, not on a licensing grant. This dependency is recorded in the A.4.6 acceptance, where A.4 couples to A.2 and A.3.

## A.4.3 Reuse inventory (cite A.1 + pinned commit)

Everything here is reused **as-is** from `cerc-io/go-nitro` **@435eb2b** per [A.1.3](./A.1-reuse-inventory.md#a13-go-nitro-adjudicator-t02--reuse-as-is). Citations are pinned there; this spec cites, it does not re-document (ADR-0012 rule 1).

| Reused piece | Pinned citation (@435eb2b) | Role in T0.2 |
|---|---|---|
| `MultiAssetHolder.deposit(asset, channelId, expectedHeld, amount)` | `packages/nitro-protocol/contracts/MultiAssetHolder.sol` **L36–70**; `holdings[asset][channelId]` **L24–29** | The **escrow entrypoint** T0.3 calls post-unshield; `expectedHeld` is the reorg/race guard. `channelId = NitroUtils.getChannelId(fixedPart)`. |
| `concludeAndTransferAllAssets(FixedPart, candidate)` + `transferAllAssets` | `NitroAdjudicator.sol` **L33–42**; `transferAllAssets` **L123–190** | **Finalize + payout**: loops every asset in the outcome → transfers to named destinations. The payout leg of the skeleton. |
| `_executeSingleAssetExit` / `_isExternalDestination` | `MultiAssetHolder.sol` **L411–427** | Confirms **T0.3 payout is an external destination** that receives ERC20 and re-shields (A.5). |
| `challenge` / `checkpoint` / `conclude` | `ForceMove.sol` **L39–80** / **L88–119** / **L126–172** | The dispute state machine. `checkpoint` (higher-turn, **no finalize**) is the exact primitive T6.3 uses to defeat a stale challenge (A.6). |
| `recoverVariablePart` — **signing domain** | `ForceMove.sol` **L236–268** | State sigs = `NitroUtils.hashState(fixedPart, variablePart)`. **T0.3/T6.2 and every app must sign this exact domain** (→ A.4.5). |
| exit-format `Outcome` structs (`SingleAssetExit`, `Allocation`, `AllocationType{simple, withdrawHelper, guarantee}`) | exit-format `ExitFormat.sol` **L26–73** | The wire format T0.3 ABI-encodes for payout; `withdrawHelper` is the alternate re-shield route. |
| `IForceMoveApp.stateIsSupported` app hook | `NitroAdjudicator.sol` **L193–207** | The `appDefinition` dispatch point — where the settlement app (skeleton or C's) is bound. |
| `HashLockedSwap.sol` reference app | `examples/HashLockedSwap.sol` **L33–83** | Pattern for a ForceMove app — **single-asset / 2-party only** (basis of the A.4.4 delta). |
| multi-asset swap negotiation | `protocols/swap/swap.go` **L105–408** | Proof that **multi-asset ETH-in/USDC-out is supported off-chain** by the protocol/holder — correcting the "single-asset gap" (A.1.8 §2). |
| vouchers | `payments/vouchers.go` **L23–52** | In-channel micropayment primitive reused by C metering + settlement (not net-new here). |

For the walking skeleton, the `appDefinition` is a **plain settlement/consensus app** (HashLockedSwap-grade single-asset stand-in, per A.0.4) — reused directly as the trivial-settle proof. The **multi-asset app is the only app-side net-new item** (A.4.4).

## A.4.4 Net-new delta

Two deltas, both flagged in [A.1.9](./A.1-reuse-inventory.md#a19-net-new-deltas-what-a-actually-builds) §2 and A.0.5 gate 4. Everything else in A.4.3 is reuse.

1. **Multi-asset ForceMove settlement app (ETH-in / USDC-out).** The escrow (`MultiAssetHolder`), the payout loop (`transferAllAssets`), and the off-chain negotiation (`swap.go`) all already support multiple assets. The A.1.8 §2 correction is load-bearing here: the gap is not the holder, it is the app. No shipped ForceMove app produces a multi-asset atomic outcome; `HashLockedSwap.sol` (`examples/…` L33–83) is single-asset and 2-party. T0.2 authors an `IForceMoveApp` whose `stateIsSupported` validates a state that debits ETH and credits USDC in one outcome, so the exchange leg (C, T4.1) can settle atomically over reused clearing. C authors the real quote/settle variant against this same `IForceMoveApp` interface and registry; T0.2 ships the reference app and the registration mechanism.

2. **Dispute paths driven end-to-end through the node.** The contract `challenge`, `checkpoint`, and `conclude` methods exist and are reused, but go-nitro's default engine does not wire them adversarially the way Armada needs (A.1.6). On an adversarial `ChallengeRegistered`, a non-initiator engine auto-creates a `directdefund` (conclude and withdraw), not a higher-turn `checkpoint` (`engine.go` L540–612). T0.2's delta is to exercise challenge, checkpoint, and conclude end-to-end from the node (open, fund, dispute, finalize) so the paths are proven and callable. The automatic watch-and-checkpoint control loop and its liveness are T6.3 (A.6), not T0.2.

The dispute lifecycle T0.2 must drive end-to-end (contract methods reused, wiring net-new):

```
  open ──deposit──► FUNDED ──(cooperative: final state signed)──► conclude ──► payout (T0.3)
                       │
           (party goes silent / stale state)
                       ▼
                   challenge(FixedPart, candidate)     ForceMove.sol L39–80
                       │  emits ChallengeRegistered{FinalizesAt}
          ┌────────────┴─────────────────────────────┐
          ▼ higher-turn state exists                  ▼ FinalizesAt reached
   checkpoint(...)  L88–119  emits ChallengeCleared    conclude(...)  L126–172
   (defeats stale challenge, NO finalize)              → concludeAndTransferAllAssets → payout
          │                                            → transferAllAssets loops every asset
          └──► channel returns to FUNDED (off-chain continues)
```

The `checkpoint` branch is the T6.3 primitive. T0.2 proves it is callable and correct; T6.3 (A.6) automates when it fires.

### Maturity gaps (restated accurately per A.1.8)

- **Multi-asset app is net-new.** The holder and protocol are multi-asset; the app is the missing piece (delta 1). The accurate framing is that the app is net-new, not that go-nitro outcomes are single-asset (A.1.8 §2 correction).
- **Dispute wiring end-to-end.** The paths are reachable but not the engine default; challenge and checkpoint must be driven explicitly (delta 2).
- **No auto-watchtower.** There is no automatic higher-turn checkpoint response; only a manual `CounterChallengeRequest` reaches checkpoint or challenge (`engine.go` L882–915). The automatic loop that watches, compares `turnNum`, and submits before `FinalizesAt` is net-new and owned by T6.3 (see [A.6](./A.6-settlement-client-watchtower.md)). It is out of T0.2 scope, named here so the boundary is unambiguous.
- **Persistent-connection liveness.** Defending a challenge requires an always-on node for the full challenge window; the browser node (ts-nitro) has no background watch loop or guaranteed relay connection (A.1.6). This is also a T6.3 concern, noted here so T0.2 acceptance does not over-claim liveness.

## A.4.5 Interfaces (ICD)

What T0.2 exposes to its consumers. These are the surfaces that T0.3 (A.5), C's app (T4.1), and the client/watchtower (A.6) bind to. A.0's WP-level "A exposes" list (`00-work-packages.md` §22) is the roll-up.

- **Adjudicator address.** The deployed `NitroAdjudicator` address on the fixturenet — the single on-chain target for `deposit`, `challenge`, `checkpoint`, `conclude`, `concludeAndTransferAllAssets`. Published to T0.3 (A.5) and to the settlement client (A.6).
- **Escrow entrypoint (deposit).** `MultiAssetHolder.deposit(asset, channelId, expectedHeld, amount)` — T0.3 calls this after `IERC20.approve`; `channelId = NitroUtils.getChannelId(fixedPart)`; `expectedHeld` must equal current `holdings[asset][channelId]` (reorg guard). ABI per A.4.3 / A.1.3.
- **Finalize entrypoint (payout).** `concludeAndTransferAllAssets(FixedPart, candidate)`: the outcome names T0.3 as an external destination (or an `AllocationType.withdrawHelper`), and T0.3 receives the ERC20 and re-shields (A.5 owns the encoding).
- **`appDefinition` / app-registry surface.** The `IForceMoveApp` binding point (`NitroAdjudicator.sol` L193–207, `stateIsSupported`). T0.2 exposes three things: (a) the trivial single-asset app address (skeleton); (b) the reference multi-asset app address; and (c) the registry convention, meaning how an `appDefinition` address is chosen per channel and validated. C's quote/settle app (T4.1) registers here, and the T0.3 app registry (A.5) records the mapping. **This is the shared contract between A.4, A.5, and C:** the `appDefinition` is set in `FixedPart` at channel open and is immutable for the channel's life.
- **Signing domain.** `recoverVariablePart` (`ForceMove.sol` L236–268): every state signature is over `NitroUtils.hashState(fixedPart, variablePart)`. **T0.3 (A.5), T6.2 (A.6), and both apps MUST sign this exact domain**, because a mismatch fails `stateIsSupported` or `recoverVariablePart`. This is the single most binding ICD element; publish the exact `hashState` derivation (the fields and encoding of `FixedPart` and `VariablePart`) so consumers cannot drift.
- **Outcome wire format.** exit-format `Outcome` (`ExitFormat.sol` L26–73): `SingleAssetExit[]` each with `Allocation[]` (`AllocationType{simple, withdrawHelper, guarantee}`). T0.2 fixes which `AllocationType` the skeleton uses (`simple` external destination); A.5 owns the T0.3-side encode/decode.
- **Dispute events.** `ChallengeRegistered` / `ChallengeCleared` (surfaced via chainservice, cf. A.1.6 / `eth_chainservice.go` L494–539) — the freshness signal T6.3 watches and B (T1.2) indexes. T0.2 guarantees these fire on the redeployed adjudicator; T6.3 owns the response loop.

## A.4.6 Acceptance / verification

Verified on a laconic fixturenet (per A.0.4 walking skeleton). Two lifecycles must pass:

1. **Cooperative path — channel open / fund / cooperative-close.**
   - Deploy `NitroAdjudicator` @435eb2b; confirm the adjudicator address and `appDefinition` (trivial single-asset app) resolve.
   - `MultiAssetHolder.deposit` funds `holdings[asset][channelId]`; assert `holdings` reflects the deposit and `expectedHeld` guard holds on a second deposit.
   - Run the settlement app off-chain to a finalizable state, then call `concludeAndTransferAllAssets`; assert every allocation is transferred to its destination and the T0.3 external destination receives the ERC20 (payout leg reachable). No on-chain dispute is triggered on the happy path.
2. **Adversarial path — force-close + conclude.**
   - From a funded channel, submit `challenge` (`ForceMove.sol` L39–80), then assert `ChallengeRegistered` fires with a `FinalizesAt`.
   - Submit a higher-turn `checkpoint` (L88–119), then assert `ChallengeCleared` and that the challenge is defeated without finalize (this proves the T6.3 primitive is callable; the auto loop is A.6).
   - Let a challenge run to `FinalizesAt` and `conclude` (L126–172), then assert the channel finalizes and `concludeAndTransferAllAssets` pays out the challenged outcome.
   - **Signing-domain check:** a state signed against the wrong `hashState` domain (see A.4.5) is rejected by `recoverVariablePart`, which proves consumers cannot silently drift.
3. **Multi-asset app (delta 1):** the reference `IForceMoveApp` accepts an ETH-in/USDC-out atomic outcome via `stateIsSupported`, and `transferAllAssets` pays both assets to their respective destinations in one `conclude`. The skeleton may run single-asset; this gate proves the net-new app before C builds T4.1 on it.

**Explicitly reported non-claims:** T0.2 does not demonstrate an automatic watchtower response or full-challenge-window liveness; those are T6.3 and A.6. The joined deposit-from-shielded-note skeleton depends on the clean-room pool and circuits (A.4.2, A.1.10; licensing resolved via [ADR-0014](../09-architecture-decisions.md#adr-0014)) for the pool side. The adjudicator lifecycles above run independently of that dependency and are the deliverable proof for T0.2.

## A.4.7 Risks / open

- **Licensing (Phase-0 — resolved via clean-room, [ADR-0014](../09-architecture-decisions.md#adr-0014)).** The adjudicator is freely redeployable; the coupled note-to-deposit path waits on the clean-room pool and circuit builds (A.2/A.3), no longer on a Railgun licensing grant. Track the build dependency, but do not block T0.2's own acceptance.
- **`appDefinition` immutability.** The app is set at channel open in `FixedPart`; a channel cannot switch apps mid-life. C (T4.1) and A.5's registry must choose the app before open, so this constraint is surfaced in the registry ICD (A.4.5, A.5).
- **Multi-asset app is unaudited net-new.** It is the only novel contract in T0.2; its `stateIsSupported` predicate is security-critical (atomicity of the swap outcome). Author narrowly, mirror the reused `HashLockedSwap` shape, and hand C a reviewed reference to fork (T4.1).
- **Dispute-path default divergence.** go-nitro's engine defaults to `directdefund`, not `checkpoint`, on adversarial challenge (A.1.6). T0.2 must drive checkpoint explicitly for its acceptance; do not assume the stock engine covers it — that assumption is exactly what A.6/T6.3 exists to fix.
- **Pin before build.** go-nitro is pinned (@435eb2b); the Railgun side is not (A.1.10). Any integration test spanning A.4 with A.2 or A.5 must pin a Railgun commit first (A.0.6).
