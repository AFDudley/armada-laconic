# A.0 — The on-chain settlement rail · Overview

work package A · **2026-10-20**
**Parent:** [`../00-work-packages.md`](../00-work-packages.md) · [`../05-building-block-view.md`](../05-building-block-view.md)
**Owns:** T0.0 pool · T0.1 trusted setup · T0.2 adjudicator (with the deposit/payout boundary merged in) · T0.4–T0.6 (plan) · T0.7 anonymity-set strategy · governance and revenue contracts

---

## A.0.1 What is built

The settlement rail — *notes in, normal Nitro, notes out* — is built, not planned. `armada-poc` built the shielded pool and its governance contracts. `armada-circuits` built the `Transact` circuit. `armada-nitro` built `ArmadaNitroAdjudicator`, a patched `go-nitro` adjudicator that also carries the deposit/payout boundary: `depositFromNote` funds a channel from a note, and a `SHIELD`-type allocation pays a note back out.

The hub principal, `HubCustodian`, signs with FROST over secp256k1 (`frost-secp256k1-evm`).

## A.0.2 The construction, as built

```
 shielded note ──depositFromNote──►  ArmadaNitroAdjudicator.deposit (holdings[asset][channelId])
                                                  │ off-chain ForceMove states / vouchers
                                                  ▼
 fresh notes ◄──pool.shield()── _executeSingleAssetExit (SHIELD allocation) ◄── conclude / force-close
```

`depositFromNote(channelId, expectedHeld, transactions)` requires a single `UnshieldType.NORMAL` transaction whose `adaptContract`/`adaptParams` are bound to the channel and holding, and whose note pays the adjudicator (`npk == address(this)`). It measures the ERC-20 balance delta and credits `holdings[asset][channelId]`. On payout, `_executeSingleAssetExit` builds a `ShieldRequest` for every `SHIELD`-type allocation and calls `pool.shield(requests, address(0))` once per asset exit. `IArmadaPool` is the two-function interface the adjudicator calls against the pool: `shield` and `transact`.

## A.0.3 Document map

| Doc | Item(s) | Contents |
|---|---|---|
| [A.1](./A.1-reuse-inventory.md) | all | What is reused (`go-nitro`, patched) versus built (the pool, the circuit, the adjudicator, the custodian), with citations. |
| [A.2](./A.2-pool-deployment.md) | T0.0 | The built pool: `PrivacyPool` and its four modules. |
| [A.3](./A.3-trusted-setup.md) | T0.1 | The built circuit and the open Phase-2 ceremony. |
| [A.4](./A.4-adjudicator-integration.md) | T0.2 | The patched adjudicator and its six post-vendor commits. |
| [A.5](./A.5-deposit-payout-contract.md) | T0.2 | The deposit/payout boundary, merged into the adjudicator. |
| [A.6](./A.6-settlement-client-watchtower.md) | T6.2, T6.3 | The web client's settlement calls; the watchtower that exists only in the test harness. |
| [A.7](./A.7-anonymity-set.md) | T0.7 | The crowd-growth strategy. |
| [A.8](./A.8-interfaces-acceptance.md) | ICD | What the rail exposes; the test suites that already run against it. |
| [A.9](./A.9-native-commitment.md) | T0.6 *(not built)* | Amount-privacy-in-play, still a deferred circuit change. |

## A.0.4 Testing

`armada-poc` ships Hardhat/TS and Foundry suites (fuzz, invariant, Halmos) for the pool, crowdfund, governance and yield paths. `armada-nitro` ships Forge suites for the adjudicator, the custodian, the asset allow-list, the reentrancy guard, and fixturenet e2e jobs that shield, deposit from a note, and run against a real `geth` + `lighthouse` chain. There is no separate "walking skeleton" to build; the equivalent flow already runs in `armada-nitro`'s `env/gates/nitro-e2e`.

## A.0.5 What remains

1. **T0.1's Phase-2 ceremony.** The shapes are open, with no beacon and no published transcript.
2. **The selective-disclosure circuit.** In progress on unmerged branches (`armada-circuits` `disclosure-circuit`, `armada-sdk` `selective-disclosure`).
3. **T0.4, T0.5, T0.6.** No code exists for the registry/bond contract, the fraud-proof verifier, or native channelized commitment.
4. **A client watchtower.** `armada-nitro`'s e2e gate harness has one; no client-side watchtower exists.

## A.0.6 Provenance

`armada-poc@102f17b8` · `armada-circuits@ef2682f` · `armada-deployments@bac3639` · `armada-nitro@55939ef` (`origin/integrated`) · `armada-sdk@db5757f`. Full citations: [A.1](./A.1-reuse-inventory.md).
