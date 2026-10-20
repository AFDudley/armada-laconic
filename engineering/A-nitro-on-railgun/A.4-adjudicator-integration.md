# A.4 — Adjudicator Integration

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T0.2 — the patched `go-nitro` adjudicator, with the deposit/payout boundary merged in

---

## A.4.1 What is built

`ArmadaNitroAdjudicator` (`armada-nitro@55939ef:src/ArmadaNitroAdjudicator.sol`) extends `go-nitro`'s `NitroAdjudicator`. It is deployed and tested on `armada-nitro`'s fixturenet (real `geth` + `lighthouse`), not a plan.

## A.4.2 Not a vanilla deploy

The vendor import (`8edacc1`, "unmodified") is followed by six commits that patch the vendored `go-nitro` contracts:

1. `5ff5eee` — accept EIP-1271 participants; remove the owner-controlled L2↔L1 bridge stub.
2. `25f40ba` — fix a guarantee-type check in `MultiAssetHolder.transfer` (wrong allocation index; traced to an upstream `statechannels/statechannels` bug carried into `cerc-io/go-nitro v0.1.7`).
3. `592a7f3` — credit deposits by the measured balance delta; revert on a short transfer.
4. `04b8431` — add a hub-configured asset allow-list (`configuredAssets`), gating `deposit` and `depositFromNote`.
5. `32ad5e5` — add a per-channel reentrancy lock (`holdingsLock[channelId]`) across six holdings-mutating entry points.
6. `8570389` — pay out ERC-20 via `SafeERC20`.

Audit of the settlement rail must cover these six patches, not only `ArmadaNitroAdjudicator`'s own added code.

## A.4.3 The deposit/payout boundary is inside this contract

`depositFromNote(channelId, expectedHeld, transactions)` requires a single `UnshieldType.NORMAL` transaction bound to this channel and holding via `adaptContract`/`adaptParams`, requires the note pays the adjudicator (`npk == address(this)`), measures the ERC-20 balance delta through `pool.transact`, and credits `holdings[asset][channelId]`.

`_executeSingleAssetExit` pays every allocation of type `SHIELD` (`SHIELD_ALLOCATION_TYPE = 3`) as a fresh note via `pool.shield(requests, address(0))`; non-`SHIELD` allocations keep normal `go-nitro` payout behavior (external destination transfer, or internal holdings credit).

There is no separate deposit/payout contract and no `fundChannel`; see [A.5](./A.5-deposit-payout-contract.md).

## A.4.4 `HubCustodian`

`armada-nitro@55939ef:src/HubCustodian.sol` is a FROST-gated principal for a hub, usable as a channel participant through EIP-1271. `execute` verifies a group signature with `FROST.verifySignature` (`frost-secp256k1-evm`, pinned `65d03d0`) before calling out; `authorizeDelegate`/`revokeDelegate` manage a hot ECDSA delegate with a timed revocation; `rotateGroupKey` rotates the FROST group key.

## A.4.5 Interfaces

**Adjudicator address and `IArmadaPool`.** `ArmadaNitroAdjudicator`'s constructor takes `(IArmadaPool pool_, address[] assets_)`. `IArmadaPool` is the two-function interface it calls: `shield(ShieldRequest[], address integrator)` and `transact(Transaction[])`.

**Asset allow-list.** `isAsset(address)` reads `configuredAssets`, set once at construction; `deposit` and `depositFromNote` are gated on it.

**Events.** `Deposited`, go-nitro's standard challenge/checkpoint/finalize events, and the pool's own `Shield`/`Transact`/`Nullified` events on payout.

## A.4.6 Testing

`armada-nitro`'s Forge suites cover: the deposit-from-note flow and its guards (`ArmadaNitroAdjudicator.t.sol`); the custodian (`HubCustodian.t.sol`, `FrostSigner.t.sol`); the asset allow-list (`AdjudicatorAssetList.t.sol`); the measured-deposit fix (`DepositAmount.t.sol`); the guarantee-transfer regression (`NitroGuaranteeTransfer.t.sol`); the reentrancy guard across all six entry points (`ReentryGuard.t.sol`, `ReentryDeposit.t.sol`); and watchtower-relevant per-asset payout checks (`WatchtowerPerAssetPayoutCheck.t.sol`). Fixturenet e2e jobs (`armada-nitro-e2e`, `armada-custodian-e2e`, `armada-watchtower-e2e`) run the same flows against a real consensus chain.

## A.4.7 Risks / open

- **The six patches are the audit surface**, not only the new `ArmadaNitroAdjudicator` code. A bug reintroduced by a future `go-nitro` upstream merge would need to re-apply or re-verify all six.
- **`HubCustodian`'s FROST scheme is unaudited by any Armada-specific audit cited in these docs.** `frost-secp256k1-evm` is a third-party library, pinned by commit, not by a security review recorded here.
- **The submodule pin discrepancy.** `armada-nitro`'s README cites `lib/armada-poc` at `532bc064`; the actual git submodule pin is `2db1d338`. This is unresolved; flag it rather than assume either is current.

---

Cross-refs: reuse inventory → [A.1](./A.1-reuse-inventory.md); the boundary detail → [A.5](./A.5-deposit-payout-contract.md); FROST decision → [ADR-0016](../09-architecture-decisions.md#adr-0016).
