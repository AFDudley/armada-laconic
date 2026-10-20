# A.8 — Interfaces (ICD) & Acceptance

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** the A-level interface list and the test suites that already verify it.

---

## A.8.1 What this page is

A rollup of what the built on-chain anchor exposes to B, C, and D, and the test suites that already verify it. There is no separate "walking skeleton" to design: `armada-nitro`'s fixturenet e2e jobs already run the full deposit-from-note to payout flow against a real `geth` + `lighthouse` chain.

## A.8.2 What A exposes

**Pool address and ABI** (`armada-poc`): `shield`, `transact`, `atomicCrossChainUnshield`; events `Shield`, `Transact`, commitment-insert and nullifier-spend events per module; `rootHistory[treeNumber][root]`. See [A.2](./A.2-pool-deployment.md).

**Adjudicator address, `IArmadaPool`, and the deposit/payout boundary** (`armada-nitro`): `depositFromNote`, the `SHIELD`-allocation payout branch, `isAsset`, go-nitro's `challenge`/`checkpoint`/`conclude`. See [A.4](./A.4-adjudicator-integration.md), [A.5](./A.5-deposit-payout-contract.md).

**`HubCustodian`**: `execute` (FROST-verified), `authorizeDelegate`/`revokeDelegate`, `isValidSignature` (EIP-1271), `rotateGroupKey`.

**Circuit artifacts**: `Transact.circom`'s 19 shapes, from `armada-circuits`' build, consumed by `armada-sdk`'s prover. See [A.3](./A.3-trusted-setup.md).

## A.8.3 What A consumes

Nothing today. The adjudicator talks to the pool directly through `IArmadaPool`; there is no watcher dependency inside the on-chain anchor. A client reading chain state is T1/T2's concern (the nimbus verified proxy, [ADR-0009](../09-architecture-decisions.md#adr-0009)), not an A-level dependency.

## A.8.4 Acceptance — what already runs

`armada-nitro`'s `env/gates/nitro-e2e/prepare.ts` shields USDC from a test account, builds a real-proof unshield of an ERC-20 note to the adjudicator for `CHANNEL_COUNT` channels, with each channel's note built from a fresh Railgun wallet/shield to avoid nullifier collisions. The fixturenet stack (`armada-nitro-fixturenet`) runs a single-node PoS devnet on upstream `geth` + `lighthouse`.

Forge tests (`ArmadaNitroAdjudicator.t.sol`, `HubCustodian.t.sol`, `ReentryGuard.t.sol`, `WatchtowerPerAssetPayoutCheck.t.sol`, and the rest listed in [A.4](./A.4-adjudicator-integration.md) and [A.5](./A.5-deposit-payout-contract.md)) cover the cooperative and adversarial cases: deposit from a note, stale-challenge override by a user's latest state, a revoked delegate's states settling during the revocation delay, and a hostile contract participant unable to exhaust a counterparty's challenge.

`armada-poc`'s Hardhat/TS and Foundry suites cover the pool side: shield/transact round-trips, fuzz and invariant tests, and Halmos symbolic-execution tests for allocation, checkpoint, and fee paths.

## A.8.5 What is not covered

- **No multi-asset ForceMove settlement app ships.** The adjudicator and its reused `go-nitro` holder support multi-asset outcomes; no app that uses that to settle an atomic multi-asset swap exists, because no venue (T4) exists to need one.
- **No client watchtower.** `armada-nitro`'s e2e harness has a watchtower (`env/gates/watchtower/watch.ts`); no client-side watchtower exists in `armada-interface`.
- **No production Phase-2 ceremony.** Every acceptance run above proves against dev-script verifying keys, not a production transcript.

## A.8.6 Risks / open

- **The deposit/payout boundary is spend-authorizing** and is the top audit target, same as the earlier plan's T0.3 would have been.
- **No multi-asset app means no acceptance evidence for an atomic multi-asset swap.** This stays a gap until a venue is built.
- **FROST is a third-party crypto library** (`frost-secp256k1-evm`), pinned by commit, with no Armada-specific audit cited here.

---

See also — owning specs: [A.2](./A.2-pool-deployment.md) · [A.3](./A.3-trusted-setup.md) · [A.4](./A.4-adjudicator-integration.md) · [A.5](./A.5-deposit-payout-contract.md) · [A.6](./A.6-settlement-client-watchtower.md) · [A.7](./A.7-anonymity-set.md) · [A.9](./A.9-native-commitment.md). Reuse citations: [A.1](./A.1-reuse-inventory.md).
