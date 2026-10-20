# A.5 — The deposit/payout boundary

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** the boundary between the shielded pool and a `go-nitro` channel, built inside `ArmadaNitroAdjudicator`

---

## A.5.1 The boundary lives inside the adjudicator

`armada-nitro` built the deposit/payout boundary inside `ArmadaNitroAdjudicator`: `depositFromNote` and the `SHIELD`-type branch of `_executeSingleAssetExit`. This page describes that boundary; the adjudicator itself is [A.4](./A.4-adjudicator-integration.md).

## A.5.2 Deposit-in

```solidity
function depositFromNote(
    bytes32 channelId,
    uint256 expectedHeld,
    Transaction[] calldata transactions
) external;
```

Checks, in order: not an external-destination channel; exactly one transaction; `boundParams.unshield == UnshieldType.NORMAL`; `boundParams.adaptContract == address(this)`; `boundParams.adaptParams == depositBinding(channelId, expectedHeld)`; the note's `npk` equals this contract's address; the token is ERC-20 and allow-listed; `held == expectedHeld`; no reentrancy in progress. It then locks the channel, measures the ERC-20 balance delta across `pool.transact(transactions)`, requires the received amount equals the note's value, clears the lock, and credits `holdings[asset][channelId]`.

`depositBinding(channelId, expectedHeld) = keccak256(abi.encode(channelId, expectedHeld))` is the binding an unshield's `adaptParams` must equal, so the proof itself ties a specific unshield to a specific channel and holding — a stranger cannot redirect someone else's unshield into their own channel.

## A.5.3 Payout-out

`_executeSingleAssetExit` builds a `ShieldRequest` for every allocation with `allocationType == SHIELD` and `amount > 0` (zero-amount `SHIELD` allocations are skipped, since the pool rejects zero-value notes), then calls `pool.shield(requests, address(0))` once per asset exit. Non-`SHIELD` allocations keep normal `go-nitro` payout behavior. The pool charges its shield fee inside `pool.shield`; the adjudicator does not compute it.

## A.5.4 Reused go-nitro surface

`MultiAssetHolder`'s holdings accounting, `ForceMove`'s `challenge`/`checkpoint`/`conclude`, and the exit-format `Outcome` structs are reused, patched per [A.4](./A.4-adjudicator-integration.md). `RelayAdapt`/Cookbook, Railgun's own value-moving recipe, is not used anywhere in this codebase; there is no venue hedge leg (T4.2) to use it in, since no venue exists.

## A.5.5 Testing

`ArmadaNitroAdjudicator.t.sol` covers: crediting a channel from a shielded note (`test_depositFromNote_creditsChannelFromShieldedNote`), rejecting an unshield bound to another channel, rejecting a stale `expectedHeld`, rejecting a bound unshield submitted around the adjudicator, cooperative close with a shield payout to the custodian, force-close letting the user exit to a note when the hub stops responding, and a stale hub challenge overridden by the user's latest state.

## A.5.6 Risks / open

- **Spend-authorizing.** This boundary moves user value across two trust domains in one flow. It is the top audit target of the on-chain anchor.
- **No native ETH.** The pool is ERC-20/ERC-721 only; `_executeSingleAssetExit` requires `asset != address(0)` before shielding ("Cannot shield ETH"). An ETH-denominated channel needs a WETH wrap.

---

Cross-refs: the adjudicator and its six patches → [A.4](./A.4-adjudicator-integration.md); the pool's `shield`/`transact` entry points → [A.2](./A.2-pool-deployment.md); the boundary's design → `labbook/2026-10-19-docs-versus-code.md`.
