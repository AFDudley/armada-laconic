# A.2 — Pool Deployment

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T0.0 — the built shielded pool

---

## A.2.1 What is built

`PrivacyPool` (`armada-poc@102f17b8:contracts/privacy-pool/PrivacyPool.sol`) is deployed and tested, not a plan. It is not a proxy: it holds all state and routes by `delegatecall` to four modules (`ShieldModule`, `TransactModule`, `MerkleModule`, `VerifierModule`), each set once at `initialize()` with no setter. This matches Railgun's note/commitment/nullifier shape, under Armada's own control (ADR-0002, ADR-0014).

## A.2.2 What was configured, as deployed

- **Fee.** `shieldFee` defaults to 0 bps, owner-settable, capped at `MAX_SHIELD_FEE_BPS = 1000`. The Sepolia deploy script sets it to 50 bps (`scripts/deploy_privacy_pool.ts:264-265`).
- **Token policy.** A blocklist, not an allow-list: any ERC-20 not blocklisted can be shielded, and USDC can never be blocked.
- **No admin proxy.** `PrivacyPool` has no upgrade admin to own; `owner` is a plain address wired to a timelock by the deploy scripts, and the four module addresses are immutable after `initialize()`.
- **Verifier registry.** `VerifierModule` holds `verificationKeys[nullifierCount][commitmentCount]`, owner-set via `PrivacyPool.setVerificationKey`, with no deletion path. 19 shapes from `armada-circuits` are registered, `(1,1)` through `(8,4)`.
- **No entry gate.** Reading every file under `contracts/privacy-pool/` finds no POI or disclosure-proof check. Selective disclosure (ADR-0020) is voluntary, off-chain, and needs no pool setting.

## A.2.3 Cross-chain

`PrivacyPool`/`PrivacyPoolClient`/`CCTPHookRouter` shield and unshield across chains over CCTP v2. `atomicCrossChainUnshield` is USDC-only; `handleReceiveFinalizedMessage`/`handleReceiveUnfinalizedMessage` route incoming messages by minimum finality.

## A.2.4 Interfaces (what the pool exposes)

**Addresses.** The pool and `VerifierModule` addresses per network, consumed by `ArmadaNitroAdjudicator` (`IArmadaPool`), by `armada-relayer`'s indexer, and by `armada-interface`/`armada-sdk`.

**Events.** `Shield`, `Transact`, the per-module events `ShieldModule`/`TransactModule` emit for commitment inserts and nullifier spends, and `rootHistory[treeNumber][root]` for Merkle-root membership.

**Entry points.** `shield(ShieldRequest[], address integrator)`, `transact(Transaction[])`, `atomicCrossChainUnshield(...)`. `ArmadaNitroAdjudicator.depositFromNote` calls `pool.transact` with an `adaptContract`/`adaptParams` binding to a specific channel and holding; `_executeSingleAssetExit`'s `SHIELD` branch calls `pool.shield` for payout.

## A.2.5 Testing

`armada-poc`'s Hardhat/TS suites (`privacy_pool_integration.ts`, `privacy_pool_adversarial.ts`, `privacy_pool_gas.ts`, `privacy_pool_hardening.ts`) and Foundry suites (`PrivacyPoolFuzz.t.sol`, `PrivacyPoolInvariant.t.sol`, `PrivacyPoolFullInvariant.t.sol`, `PrivacyPoolSolvencyInvariant.t.sol`, `SnarkVerifier.t.sol`, Halmos symbolic tests) already run against the pool.

## A.2.6 Risks / open

- **Phase-2 ceremony not finished (A.3).** The pool can only register vkeys the ceremony has produced; today's registered keys trace to the unsafe dev script, not a production ceremony.
- **The deployed fee is not zero.** Any doc or integration that assumes fee = 0 is wrong; it is 50 bps on the Sepolia deployment, and the owner can raise it to 1000 bps.
- **`setVerificationKey` has no structural validation and silently overwrites.** The authority that can call it (today the deploying owner, later governance) is a real key-management risk, same as the earlier plan flagged for a proxy admin.

See also: [A.3](./A.3-trusted-setup.md) (the ceremony that feeds the verifier registry), [A.5](./A.5-deposit-payout-contract.md) (the adjudicator that calls `shield`/`transact`), [T0-ethereum-anchor.md](../T0-ethereum-anchor.md).
