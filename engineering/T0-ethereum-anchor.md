# T0 · Ethereum anchor

**Status:** built, mostly · tier doc · 2026-10-20
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T0
**Depends on / Blocks:** root tier — depends on nothing above it. Blocks T1 (indexes T0.0 commitments/nullifiers and T0.2 events), T6.2 (settlement client drives the T0.2 deposit/payout boundary), T6.3 (watchtower response against T0.2).

The on-chain anchor is where value lives and where disputes settle. `armada-poc` built the shielded pool (T0.0) and its governance contracts. `armada-circuits` built the `Transact` circuit (T0.1). `armada-nitro` built the adjudicator, and merged the deposit/payout boundary into it (T0.2, ruling 1 of `labbook/2026-10-19-docs-versus-code.md`). T0.4, T0.5 and T0.6 remain a plan.

## T0.0 Shielded pool — built

`PrivacyPool` (`armada-poc@102f17b8:contracts/privacy-pool/PrivacyPool.sol`) is the router. It holds all state and routes to four modules by `delegatecall`: `ShieldModule`, `TransactModule`, `MerkleModule`, `VerifierModule`, each set once at `initialize()` with no setter. `PrivacyPool` is not a proxy.

**What is in the Railgun design, and what Armada changed.** The pool reimplements Railgun's note/commitment/nullifier shape: a 16-deep commitment tree per `treeNumber` (65,536 leaves), `rootHistory[treeNumber][root] -> isValid`, and `nullifiers[treeNumber][nullifier] -> isSpent`, set in `TransactModule._accumulateAndNullify` before payout. A batch that would overflow the current tree rolls to `(treeNumber+1, 0)`. Token policy is a blocklist, not an allow-list: any ERC-20 not blocklisted can be shielded, and USDC can never be blocked.

**Fee.** `shieldFee` defaults to 0 bps and is owner-settable, capped at `MAX_SHIELD_FEE_BPS = 1000` (10%). The Sepolia deploy script sets it to 50 bps (`scripts/deploy_privacy_pool.ts:264-265`). Armada's own revenue model (`ArmadaFeeModule`, a tiered take on shield and yield fees) is built alongside the pool (`contracts/fees/ArmadaFeeModule.sol`).

**No entry gate.** No proof-of-innocence or disclosure-proof check exists anywhere under `contracts/privacy-pool/`. Selective disclosure (ADR-0020, [16](./16-selective-disclosure.md)) is voluntary and off-chain and gates nothing.

**Cross-chain.** `PrivacyPool`/`PrivacyPoolClient`/`CCTPHookRouter` shield and unshield across chains over CCTP v2 (`atomicCrossChainUnshield`, `handleReceiveFinalizedMessage`). Cross-chain unshield is USDC-only.

**Verifier.** `VerifierModule` holds `verificationKeys[nullifierCount][commitmentCount]`, owner-set via `PrivacyPool.setVerificationKey`, with no deletion path. 19 `(nIn, nOut)` shapes are registered from `armada-circuits`' build, from `(1,1)` through `(8,4)`.

**Testing.** `armada-poc` ships Hardhat/TS suites for the pool, crowdfund, governance, revenue and yield paths, and Foundry suites including fuzz, invariant and Halmos symbolic-execution tests for the pool and the adjudicator boundary.

## T0.1 Circuits + trusted setup — partly built

`armada-circuits` (`ef2682f`) is one circom template, `Transact(N,M)` (`src/operations/Transact.circom`, 167 lines, plus `merkle.circom`, 55 lines), instantiated as 19 `main_NxM.circom` entry points. It builds on `circomlib ^2.0.5`, the way Railgun's own circuits do. Groth16 proving; Phase-1 reuses the community Perpetual Powers of Tau.

**The Phase-2 ceremony has not run.** The `v1.0.0-ceremony` tag's `ceremony/manifest.json` lists every shape `"status": "open"` with no beacon. `main` carries only `scripts/setup-dev.sh`, labeled unsafe for anything but local testing. Until the ceremony finishes, every deployed verifying key traces to the dev script, not a published, multi-party transcript.

**Selective disclosure is a second, unmerged circuit.** `NullifierDisclosure.circom` (112 lines) exists on the `disclosure-circuit` branch (`7175b46`), with the matching client code on `armada-sdk`'s `selective-disclosure` branch (`0aa2ba1`). It is in progress, not merged. It will add shapes to this same Phase-2 ceremony, not start a new one (ADR-0003 D3.4).

**Re-run trigger.** Phase 2 is repeated only when circuits change. The selective-disclosure circuit and any future T0.6 (native channelized commitment) change are both triggers.

## T0.2 Nitro adjudicator and deposit/payout boundary — partly built, merged

`ArmadaNitroAdjudicator` (`armada-nitro@55939ef:src/ArmadaNitroAdjudicator.sol`) extends `go-nitro`'s `NitroAdjudicator`. The vendor import (`8edacc1`, "unmodified") is followed by six commits that patch the vendored contracts:

1. Accept EIP-1271 participants (removes the owner-controlled L2↔L1 bridge stub).
2. Fix a guarantee-type check in `MultiAssetHolder.transfer` (wrong index).
3. Credit deposits by the measured balance delta, reverting on a short transfer.
4. Add a hub-configured asset allow-list, gating `deposit` and `depositFromNote`.
5. Add a per-channel reentrancy lock across the six holdings-mutating entry points.
6. Pay out ERC-20 via `SafeERC20`.

This is a patched adjudicator, not a vanilla deploy. It is also the deposit/payout boundary, not a second contract:

- **Deposit.** `depositFromNote(channelId, expectedHeld, transactions)` requires a `UnshieldType.NORMAL` transaction whose `adaptContract`/`adaptParams` bind to this channel and holding, requires the note pays the adjudicator (`npk == address(this)`), and credits `holdings[asset][channelId]` by the measured received amount.
- **Payout.** `_executeSingleAssetExit` pays every allocation of type `SHIELD` as a fresh note, calling `pool.shield(requests, address(0))`; non-`SHIELD` allocations keep normal `go-nitro` payout behavior.

`IArmadaPool` (`src/armada/IArmadaPool.sol`) is the two-function interface the adjudicator calls: `shield` and `transact`.

**HubCustodian.** `HubCustodian.sol` is a FROST-gated principal for a hub. It verifies a group signature (`FROST.verifySignature`, via `frost-secp256k1-evm`, pinned `65d03d0`) before executing a call, and authorizes a hot delegate by EIP-1271 with a timed revocation.

**Testing.** `armada-nitro`'s Forge suites cover the adjudicator (`ArmadaNitroAdjudicator.t.sol`), the custodian (`HubCustodian.t.sol`), the asset allow-list, the reentrancy guard, and the measured-deposit fix, plus fixturenet e2e jobs that shield, deposit from a note, and run the adjudicator against a real `geth` + `lighthouse` chain.

## T0.4 Registry + bond contract — not built

No federation registry or bond contract exists. This stays a plan for a future bonded-federation phase (T2.4).

## T0.5 Sequencing-cert + fraud-proof verifier — not built

No L1 verifier for sequencing certs exists. This stays a plan for a future ordering tier (T3).

## T0.6 Native channelized commitment — not built

Hiding amounts during play, not only at the shield/unshield boundary, would need a circuit change and a re-run of the Phase-2 ceremony (ADR-0003 D3.4, ADR-0005). No such circuit exists. This stays an optional, deferred upgrade.

## T0.7 Anonymity-set strategy — strategy, not code

A shielded pool's privacy is k-anonymity against the set of unspent, indistinguishable notes. The pool already hides amounts inside a shield; the SNARK enforces value conservation and range, so a note's size carries no information. The amount leaks only at the shield/unshield and channel deposit/payout boundaries.

**The crowd grows only from the pool's own deposits.** A spend hides only among the notes it could consume, so a Railgun note that this pool accepted would let its owner spend the same value twice, and the pool would pay out tokens it never received (`labbook/2026-10-19-docs-versus-code.md` decision 2).

**Moving from Railgun is a wallet feature (T6.7), not a protocol component.** It is a Railgun unshield followed by a shield into this pool. No contract of ours reads or touches Railgun's pool. The wallet shows the security costs before the user confirms: the token, the amount and the time are public on both pools; the unshield recipient is public; the value leaves Railgun's larger crowd for this smaller one; a later exit can link back; Railgun charges its unshield fee. This wallet feature is not built; `armada-interface` has no Railgun-unshield code today.

## Sources

- `armada-poc` `origin/main@102f17b8` — `contracts/privacy-pool/`, `contracts/fees/`, `contracts/governance/`.
- `armada-circuits` `ship-armada.github.com/main@ef2682f` — `src/operations/Transact.circom`; `disclosure-circuit@7175b46` — `src/operations/NullifierDisclosure.circom`.
- `armada-deployments` `ship-armada.github.com/main@bac3639`; tag `v1.0.0-ceremony:ceremony/manifest.json`.
- `armada-nitro` `origin/integrated@55939ef` — `src/ArmadaNitroAdjudicator.sol`, `src/HubCustodian.sol`, `src/armada/IArmadaPool.sol`.
- Evidence: `labbook/2026-10-19-docs-versus-code.md`; `.scratch/rewrite-evidence/inv-InvPool.md`, `inv-InvNitroSdk.md`, `cmp-CmpArc42.md` (in `armada-laconic`).
- Site docs — [`05-building-block-view.html`](./05-building-block-view.html) (registry) · [`09-architecture-decisions.html`](./09-architecture-decisions.html) (ADR-0002, 0003, 0004, 0014, 0016) · [`build-plan.html`](./build-plan.html) (status and sequencing).
