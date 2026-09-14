# A.5 — Deposit/Payout Contract (T0.3)

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** **T0.3 — deposit/payout contract (net-new)**; the multi-asset outcome-construction helper it carries (shared with [A.4](./A.4-adjudicator-integration.md))

---

## A.5.1 Goal

T0.3 is the one genuinely net-new on-chain component of work package A: the **deposit/payout contract**. It bridges a shielded Railgun note into a `go-nitro` state channel, and bridges the channel's finalized outcome back into fresh notes. This is the concrete realization of *notes in, normal Nitro, notes out* ([ADR-0004](../09-architecture-decisions.md#adr-0004), [A.0.2](./A.0-overview.md)). It is the **boundary** ([ADR-0010](../09-architecture-decisions.md#adr-0010)) at which the Railgun privacy analysis and the Nitro game analysis meet, and nowhere else. T0.3 holds no shielded state, runs no ZK verification, and arbitrates no channel; it is a public-side ERC20 counterparty of *unshield/shield* at the pool and of *deposit/conclude* at the adjudicator.

Every downstream scope clears over this contract: payments (A.7), yield and exchange (C, via its quote/settle app), and watcher metering (B, via events). It is **spend-authorizing**: it moves user value across two trust domains in one flow. That makes it the audit-critical artifact of the whole programme (A.5.7).

## A.5.2 Boundary

**In scope.** The T0.3 Solidity contract: its deposit-in path, payout-out path, outcome-construction helper, and event surface, along with the ForceMove app-registry slot that C's quote/settle app registers into. Its ICD to B/C/D falls in scope too (A.5.5).

**Out of scope / cited baseline.** The pool contracts (our own implementation per [A.2](./A.2-pool-deployment.md), [A.1.1](./A.1-reuse-inventory.md#a11-shielded-pool-t00-reference-spec)); the adjudicator (reused as-is, [A.1.3](./A.1-reuse-inventory.md#a13-go-nitro-adjudicator-t02-reuse-as-is), integrated in [A.4](./A.4-adjudicator-integration.md)); the settlement client that drives T0.3 (T6.2, [A.6](./A.6-settlement-client-watchtower.md)); the multi-asset ForceMove app itself (net-new but co-owned with A.4, where C authors the real quote/settle variant, T4.1); and POI policy and node stack ([A.7](./A.7-anonymity-set.md)).

**Phase-0 assumption.** T0.3 links against our own pool (T0.0, [A.2](./A.2-pool-deployment.md)) and a completed ceremony over our own circuits (T0.1, [A.3](./A.3-trusted-setup.md)). T0.3's own source is net-new and can be developed against a fixturenet pool. The Railgun reference commit is still unpinned, so pin it before build ([A.1.10](./A.1-reuse-inventory.md#a110-risks)).

## A.5.3 Reuse inventory (cite A.1 + pinned commits)

T0.3 is pure glue: every surface it touches is reused through a public ABI. It writes **no** new cryptography ([ADR-0004](../09-architecture-decisions.md#adr-0004)).

| Reused surface | Cited in A.1 | Pin | Role for T0.3 |
|---|---|---|---|
| `RailgunSmartWallet.transact(Transaction[])` | [A.1.1](./A.1-reuse-inventory.md#a11-shielded-pool-t00-reference-spec) `RailgunSmartWallet.sol` L102–224 | Railgun ref *(unpinned)* | Deposit-in trigger; `boundParams.unshield=NORMAL` unshields ERC20 out |
| `transferTokenOut` (recipient = `address(npk)`) | [A.1.1](./A.1-reuse-inventory.md#a11-shielded-pool-t00-reference-spec) `RailgunLogic.sol` L318–364 | Railgun ref *(unpinned)* | Confirms the unshield lands ERC20 at the T0.3 address |
| `adaptContract` bind check | `RailgunLogic.sol` L432–436 (see [A.1.4](./A.1-reuse-inventory.md#a14-the-depositpayout-binding-t03-how-the-two-systems-wire)) | Railgun *(unpinned)* | Must be `0` **or** `msg.sender` — governs the atomic vs. two-step deposit design (A.5.4) |
| ABI structs (`ShieldRequest`, `Transaction`, `BoundParams`, `CommitmentPreimage`, `TokenData`) | [A.1.1](./A.1-reuse-inventory.md#a11-shielded-pool-t00-reference-spec) `Globals.sol` L21–122 | Railgun ref *(unpinned)* | Exact encode surface for deposit trigger + payout shield |
| `RailgunSmartWallet.shield(ShieldRequest[])` | [A.1.1](./A.1-reuse-inventory.md#a11-shielded-pool-t00-reference-spec) `RailgunSmartWallet.sol` L23–97 | Railgun ref *(unpinned)* | Payout-out: mints fresh commitments; emits `Shield` |
| `ShieldNote.serialize` (note build) | [A.1.4](./A.1-reuse-inventory.md#a14-the-depositpayout-binding-t03-how-the-two-systems-wire) engine `src/note/shield-note.ts` L46–121 | engine *(unpinned)* | Reference for constructing each payout `ShieldRequest` (client-side, T6.2) |
| `MultiAssetHolder.deposit(asset, channelId, expectedHeld, amount)` | [A.1.3](./A.1-reuse-inventory.md#a13-go-nitro-adjudicator-t02-reuse-as-is) `MultiAssetHolder.sol` L36–70; `holdings` L24–29 | go-nitro **@435eb2b** | The escrow entrypoint T0.3 calls after unshield; `expectedHeld` reorg guard |
| `NitroUtils.getChannelId(fixedPart)` | [A.1.3](./A.1-reuse-inventory.md#a13-go-nitro-adjudicator-t02-reuse-as-is), [A.1.4](./A.1-reuse-inventory.md#a14-the-depositpayout-binding-t03-how-the-two-systems-wire) | go-nitro **@435eb2b** | Derives `channelId` T0.3 funds |
| `concludeAndTransferAllAssets(FixedPart, candidate)` / `transferAllAssets` | [A.1.3](./A.1-reuse-inventory.md#a13-go-nitro-adjudicator-t02-reuse-as-is) `NitroAdjudicator.sol` L33–42; L123–190 | go-nitro **@435eb2b** | Finalize + payout; routes outcome to destinations |
| `_executeSingleAssetExit` / `_isExternalDestination` | [A.1.3](./A.1-reuse-inventory.md#a13-go-nitro-adjudicator-t02-reuse-as-is) `MultiAssetHolder.sol` L411–427 | go-nitro **@435eb2b** | T0.3 is registered as an **external destination**; funds transferred out to it |
| Outcome encoding (`SingleAssetExit`, `Allocation`, `AllocationType{simple,withdrawHelper,guarantee}`) | [A.1.3](./A.1-reuse-inventory.md#a13-go-nitro-adjudicator-t02-reuse-as-is) exit-format `ExitFormat.sol` L26–73 | go-nitro **@435eb2b** | The wire format T0.3's helper ABI-encodes for payout |
| `appDefinition` → `IForceMoveApp.stateIsSupported` | [A.1.4](./A.1-reuse-inventory.md#a14-the-depositpayout-binding-t03-how-the-two-systems-wire) `NitroAdjudicator.sol` L193–207 | go-nitro **@435eb2b** | The app-registry slot; C's quote/settle app (T4.1) registers here |
| vouchers | [A.1.3](./A.1-reuse-inventory.md#a13-go-nitro-adjudicator-t02-reuse-as-is) `payments/vouchers.go` L23–52 | go-nitro **@435eb2b** | In-channel micropayment format T0.3 exposes in its ICD (C metering / B) |

## A.5.4 Net-new delta

Two net-new artifacts ([A.1.9](./A.1-reuse-inventory.md#a19-net-new-deltas-what-a-actually-builds) items 1–2):

```
 depositor        pool (T0.0)         T0.3           adjudicator (T0.2)      beneficiary
   │  transact(unshield=NORMAL,          │                  │                    │
   │  npk=T0.3, adaptContract∈{0,T0.3}) ─►│ transferTokenOut │                    │
   │                    ERC20 ───────────►│                  │                    │
   │                                      │ approve+deposit ─►│ holdings[asset]     │
   │                                      │ (expectedHeld)   │  [channelId]+=amt   │
   │  ══════ off-chain ForceMove states / vouchers · settle or force-close ══════  │
   │                                      │ settleAndPayout ─►│ conclude+transfer   │
   │                                      │◄──── ERC20 ───────│ (T0.3 = external)   │
   │                                      │ shield([...]) ──► pool mints fresh note ─►│ scan
```

**(a) The deposit/payout contract itself.** No Railgun or Nitro component does this. The closest Railgun primitive is RelayAdapt/Cookbook, which is Railgun's own atomic-multicall recipe rather than channel escrow (A.5.7, [ADR-0004](../09-architecture-decisions.md#adr-0004)). T0.3 provides:

- **Deposit-in.** A Railgun `transact([...])` carries `boundParams.unshield = NORMAL` and an `unshieldPreimage.npk = bytes32(uint160(T0.3))`. The pool's `transferTokenOut` (`RailgunLogic.sol` L318–364) interprets the note-public-key field as `address(uint160(npk))` and sends the unshielded ERC20 to T0.3. T0.3 then calls `IERC20.approve(adjudicator, amount)` followed by `MultiAssetHolder.deposit(asset, channelId, expectedHeld, amount)` with `channelId = NitroUtils.getChannelId(fixedPart)`. The `expectedHeld` argument carries the caller's view of current holdings and acts as the **reorg and concurrent-fund guard**: `deposit` reverts if the actual `holdings[asset][channelId]` differs (A.5.6 double-fund case).
  - **Atomicity design point.** The `adaptContract` field on `transact` must be `address(0)` or `msg.sender` (`RailgunLogic.sol` L432–436). Two shapes are admissible:
    1. **Atomic (preferred).** T0.3 is both the `adaptContract` and the caller of `transact` (`msg.sender == adaptContract`). The unshield lands in T0.3 and the `MultiAssetHolder.deposit` fires in the same transaction through T0.3's post-unshield hook, so no window exists in which unescrowed ERC20 sits at T0.3.
    2. **Two-step.** With `adaptContract = 0`, any relayer calls `transact` so that funds land at T0.3, then a follow-up call invokes `T0.3.fundChannel(...)`. This is simpler, but it leaves a transient balance that a sweep function must reconcile against the intended `channelId`.
  - The choice between the two is left open here and resolved in build against gas and relayer-model constraints (A.5.7); the ICD (A.5.5) is identical under both.
- **Payout-out.** When the channel finalizes, `concludeAndTransferAllAssets(FixedPart, candidate)` (`NitroAdjudicator.sol` L33–42) runs `transferAllAssets`, which pays each `Allocation` for each asset. T0.3 is named as an **external destination** (`_isExternalDestination`, `MultiAssetHolder.sol` L411–427): a `bytes32` destination whose top 12 bytes are zero, that is, `bytes32(uint160(T0.3))`. The adjudicator therefore transfers that beneficiary's ERC20 out to T0.3. T0.3 then builds and submits `RailgunSmartWallet.shield([ShieldRequest...])`, minting a fresh note to the true beneficiary's Railgun address; the note is built per engine `ShieldNote.serialize` (`shield-note.ts` L46–121). Where the payee wants a helper-driven exit rather than a plain balance transfer, the allocation uses `AllocationType.withdrawHelper` (`ExitFormat.sol` L26–73) and T0.3 is the helper target.
- **POI gated-entry hook.** T0.3 exposes a deposit-time hook at which valid-POI enforcement binds, taking a supplied POI attestation or status check keyed to the unshielded note. This is a settlement-side policy rather than a pool setting ([A.1.7](./A.1-reuse-inventory.md#a17-anonymity-set-t07-the-payments-primitive), [A.0.5](./A.0-overview.md#a05-open-gates-must-resolve-before-during-a) gate 2). The concrete list-provider and standby policy, along with the separate POI node dependency, are specified in [A.7](./A.7-anonymity-set.md). T0.3 owns only the enforcement point, rejecting a deposit if the POI predicate fails; A.7 owns the predicate itself.

**(b) A multi-asset outcome-construction helper.** `MultiAssetHolder`, `transferAllAssets`, and `swap.go` already support multi-asset outcomes ([A.1.8](./A.1-reuse-inventory.md#a18-corrections-to-earlier-docs-the-deep-dive-caught-these) correction 2). The gap is that no shipped ForceMove app builds a multi-asset atomic ETH-in/USDC-out outcome, since HashLockedSwap is single-asset and two-party ([A.1.9](./A.1-reuse-inventory.md#a19-net-new-deltas-what-a-actually-builds) item 2). T0.3 carries the encoding helper that both the walking-skeleton app and C's real quote/settle app (T4.1) use to produce a well-formed outcome: it assembles `SingleAssetExit[]` with per-asset `Allocation[]` (`ExitFormat.sol` L26–73), mixing external destinations that re-shield with channel and guarantee destinations. The app logic is co-owned with [A.4](./A.4-adjudicator-integration.md); the outcome encoder lives with T0.3 because both deposit and payout depend on it.

Worked shape for an ETH-in/USDC-out atomic settle (both legs re-shielded to different beneficiaries):
```
Outcome = [
  SingleAssetExit{ asset: WETH, allocations: [
      Allocation{ destination: bytes32(uint160(T0.3)), amount: aE, allocationType: simple, metadata: ⟨note A⟩ } ]},
  SingleAssetExit{ asset: USDC, allocations: [
      Allocation{ destination: bytes32(uint160(T0.3)), amount: aU, allocationType: withdrawHelper, metadata: ⟨note B⟩ } ]}
]
```
Each external-destination allocation carries the target Railgun address in its `metadata`, which T0.3's payout path maps 1:1 to a `ShieldRequest` (A.5.6 conservation check). `guarantee` allocations are left for nested-channel / virtual-funding topologies (not exercised by the walking skeleton).

## A.5.5 Interfaces (ICD) — what T0.3 exposes to B/C/D

Consumers bind to T0.3 only through this contract; the interface is the boundary ([ADR-0012](../09-architecture-decisions.md#adr-0012)).

**Address and ABI.** The deployed T0.3 address and its ABI are published in the [§5 registry](../05-building-block-view.md); T0.3 links a pool address (T0.0) and an adjudicator address (T0.2) as immutables set at deploy.

**Deposit path (driven by D/T6.2, consumed by C).**
```solidity
// Two-step shape; atomic shape folds these behind the transact hook (A.5.4).
function fundChannel(
    TokenData    asset,          // Railgun/ERC20 token descriptor (Globals.sol)
    bytes32      channelId,      // = NitroUtils.getChannelId(fixedPart)
    uint256      expectedHeld,   // reorg / double-fund guard (MultiAssetHolder.deposit)
    uint256      amount,
    POIAttestation calldata poi  // gated-entry predicate input (A.7)
) external;                      // approves + calls MultiAssetHolder.deposit
event Deposit(bytes32 indexed channelId, address indexed asset, uint256 amount, uint256 held);
```
The caller supplies `channelId` and `fixedPart` off-chain, deriving `getChannelId` client-side. The unshield leg is a normal `RailgunSmartWallet.transact` with `unshield=NORMAL`, `npk=bytes32(uint160(T0.3))`, and `adaptContract` set to either `0` or T0.3 (A.5.4).

**Payout path (driven by D/T6.2).**
```solidity
function settleAndPayout(
    FixedPart calldata fixedPart,
    SignedVariablePart calldata candidate,   // finalizing state (conclude proof)
    ShieldRequest[] calldata shields          // fresh-note requests, one per external beneficiary
) external;                                   // concludeAndTransferAllAssets → receive ERC20 → shield
event Payout(bytes32 indexed channelId, address indexed asset, uint256 amount, uint256 shieldedCommitments);
```

**Channel-lifecycle API (consumed by D/T6.2, gated by T6.3).** These are thin pass-throughs and views over the reused adjudicator, so a client need not re-derive them: `getChannelId(fixedPart) → bytes32`, `holdings(asset, channelId) → uint256`, and the finalize entrypoints above. Dispute primitives (`challenge`/`checkpoint`/`conclude`, `ForceMove.sol` L39–172) are not re-exposed by T0.3. The watchtower calls them directly on the adjudicator (T6.3, [A.6](./A.6-settlement-client-watchtower.md)); T0.3 references them only to document the pre-fill fallback (A.5.6).

**ForceMove app registry.** The `channelId` binds an `appDefinition` (`NitroAdjudicator.sol` L193–207, checked through `IForceMoveApp.stateIsSupported`). T0.3 defines the **registry convention**: the walking skeleton uses a trivial single-asset app, and C's quote/settle app (T4.1) registers its `appDefinition` here. C exposes that app id across its ICD ([`00-work-packages.md`](../00-work-packages.md) "C exposes: … the quote/settle app id"). T0.3 places no constraint on app logic beyond `stateIsSupported`; it constrains only the outcome shape its payout path can settle (via the A.5.4b helper).

**Voucher format.** T0.3 surfaces the go-nitro voucher struct (`payments/vouchers.go` L23–52) as its ICD type for in-channel micropayments. C consumes it for metering and streaming settle, and B references it as the Nitro voucher metering interface B exposes. T0.3 does not mint or verify vouchers, since that is in-channel and off-chain; it only fixes the wire type so siblings agree.

**Events (consumed by B / T1.2).** The `Deposit` and `Payout` events above, plus the reused pool `Shield`/`Transact`/`Nullified` events and the adjudicator `Deposited`/`AllocationUpdated` events. B's watcher-ingest (T1.2) indexes T0.3's `Deposit`/`Payout` alongside T0.0 and T0.2 events ([§5](../05-building-block-view.md) T1.2). The event ABIs are the A-to-B ICD ([`00-work-packages.md`](../00-work-packages.md)).

## A.5.6 Acceptance / verification

**Walking-skeleton E2E ([A.0.4](./A.0-overview.md#a04-walking-skeleton-as-first-integration-target)).** The test runs on a laconic fixturenet with our own pool (fee=0, T0.0), the adjudicator (T0.2), and a trivial single-asset ForceMove app standing in for C's quote/settle app: `shield → transact()/unshield-in (npk=T0.3) → T0.3.fundChannel → MultiAssetHolder.deposit → off-chain settle → concludeAndTransferAllAssets → T0.3 receives ERC20 → shield() → scan finds the fresh note`. Passing means *notes-in-Nitro-notes-out* with correct amounts and a scannable output note. This is the acceptance gate before T0.3 deepens to multi-asset, POI, and the atomic hook. Detail in [A.8](./A.8-interfaces-acceptance.md).

**Adversarial cases (must all hold):**
1. **Replayed nullifier rejected.** Re-submitting a deposit `transact` whose input note was already spent reverts at the pool nullifier check (`Nullified` set, [A.1.1](./A.1-reuse-inventory.md#a11-shielded-pool-t00-reference-spec) `Commitments.sol`), so T0.3 never sees phantom funds. Assert the second `transact` reverts and that no `Deposit` fires.
2. **Double-fund rejected via `expectedHeld`.** Two concurrent `fundChannel` calls for the same `channelId` share the same `expectedHeld`, and the second reverts inside `MultiAssetHolder.deposit` because the actual `holdings` no longer matches `expectedHeld` (`MultiAssetHolder.sol` L36–70). Assert that exactly one `Deposit` fires and that the escrowed total is correct, with no over-escrow and no lost funds.
3. **Missing settlement forces a close to pre-fill.** If the counterparty never counter-signs a settling state, the depositor force-closes: it issues `challenge` with the latest supported state, waits out `FinalizesAt`, then calls `conclude` to the pre-fill outcome, so escrow returns to the depositor's external destination and re-shields. Assert that the depositor recovers the deposited amount, minus gas, as a fresh note. The watchtower's higher-turn `checkpoint` defense of an in-flight channel is T6.3's concern ([A.6](./A.6-settlement-client-watchtower.md)); here the test proves the safety fallback exists.

**Property checks.** Three conservation properties hold. For every `channelId`, `Σ Payout ≤ Σ Deposit` per asset (fee=0, [ADR-0002](../09-architecture-decisions.md#adr-0002)); every external-destination allocation re-shields to exactly one `ShieldRequest`; and the `adaptContract` bind is honored, so a `transact` with `adaptContract ∉ {0, T0.3}` cannot route funds to a T0.3-triggered deposit. These are verified with a throwaway fixturenet harness driving the flow, not a permanent suite. Permanent tests earn their place only where a plausible bug would fail them, and the three adversarial cases qualify and become regression tests.

## A.5.7 Risks / open

- **Spend-authorizing, therefore audit-critical.** T0.3 is the single component that moves user value across the Railgun-to-Nitro boundary and authorizes both the unshield destination and the shield mint. A bug is a direct fund-loss or fund-theft path. It is the top audit target of work package A; the atomicity design (A.5.4) and the POI hook are the highest-risk surfaces.
- **RelayAdapt contrast (do not conflate).** Railgun's own value-moving recipe is RelayAdapt/Cookbook, a governance-gated atomic-multicall over shielded balances ([A.1.9](./A.1-reuse-inventory.md#a19-net-new-deltas-what-a-actually-builds) item 1, [ADR-0004](../09-architecture-decisions.md#adr-0004), [ADR-0010](../09-architecture-decisions.md#adr-0010)). T0.3 is not an adapter and does not reuse RelayAdapt for settlement: RelayAdapt composes calls within one atomic transaction, whereas T0.3 bridges into a persistent state channel with optimistic exits. RelayAdapt is used only for the venue's atomic hedge leg (T4.2, v2), never for settlement ([ADR-0004](../09-architecture-decisions.md#adr-0004)).
- **Atomicity versus relayer model unresolved.** Whether the deposit is atomic (T0.3 as both `adaptContract` and caller) or two-step (A.5.4) depends on the relayer and broadcaster model (T6.2, [A.6](./A.6-settlement-client-watchtower.md)) and on gas. A transient T0.3 balance in the two-step shape needs a `channelId`-keyed sweep and reconcile. Decide in build.
- **Multi-asset app is net-new (gate 3).** T0.3 provides the outcome encoder; the atomic multi-asset ETH-in/USDC-out app is still net-new ([A.0.5](./A.0-overview.md#a05-open-gates-must-resolve-before-during-a) gate 3, [A.4](./A.4-adjudicator-integration.md)). The walking skeleton uses a single-asset stand-in until C authors T4.1.
- **Native ETH.** The pool is ERC20/721 only ([A.1.10](./A.1-reuse-inventory.md#a110-risks)); ETH-in requires a WETH wrap at the boundary or the optional native commitment ([A.9](./A.9-native-commitment.md)).
- **Unpinned Railgun reference commit** (A.5.2) gates a reproducible production build of the deposit/payout flow. Pin a Railgun commit before build.
