# Proof scope and order: the pool first, governance out (2026-11-06)

Operator rulings of 2026-11-06, recorded with the facts that were checked for them. Pins: armada-lean `integrated`
`e6c605e1`; armada-poc `integrated` `fbcf2b1e`; armada-nitro `integrated` `295b6c2e`. A claim with no citation that I
did not observe has the mark `[INFERENCE]`.

## 1. The goal

Prove Armada end to end, and match the proved artifacts to the fixturenet deployment. For each contract the chain is
the one of `engineering/build-plan.md` S5 and `docs/process.md` §9: the Lean model (its executable `apply` steps) →
Stele specification (refinement proof) → Solidity written from the model → pinned `solc` → bytecode (EquiVM proof) →
the fixturenet deployment (the deployed bytecode equals the proved bytecode; the deployed keys equal the ceremony keys).

## 2. Rulings

1. **Bottom up.** The shielded pool goes through the chain first. armada-nitro follows.
2. **Governance and permissioned features are out of scope.** On the governance side the pool keeps two things only:
   a fee rate that one named address can change, and a treasury address. Out of scope: the governance and revenue
   contracts that S1 lists (`ArmadaGovernor`, `ArmadaTreasuryGov`, `TreasurySteward`, `ArmadaToken`,
   `ArmadaCrowdfund`, `ArmadaWindDown`, `ArmadaRedemption`, `AdapterRegistry`, `RevenueCounter`, `RevenueLock`,
   `RevenueReserveDistributor`), `HubCustodian`, and all of S4 (Baal treasury governance).
3. **One named address can change the fee rate.** The deployment fixes that address. No other caller and no other
   step changes the rate.
4. **S2 is satisfied in the model.** A contract (for example a Safe) can hold notes and spend them.

## 3. Facts checked

**The model already has the scope of ruling 2, except the fee setter.** No admin and no pause (`ArmadaLean/Pool.lean`
lines 82-90, the ruling on finding P-3). The fee rate is a deployment parameter, `PoolState.feeRatio`, and no step
changes it (`step_feeRatio`). Shield fees go to the treasury (`PoolState.treasuryFees`; the shield theorems of
`ArmadaLean/PoolSafety.lean`). Ruling 3 needs a new step, `setFeeRatio`, and its theorems. The request went to the
armada-lean pool owner (herdr `w3:p5`) on 2026-11-06.

**S2 in the model.** `Step.transactOwned` (`Pool.lean:1180`); `owned_owner_binding` (`PoolOwned.lean:138`);
`owned_no_double_spend` (`PoolOwned.lean:175`); `preimageOf_consistent` (`PoolOwned.lean:282`);
`owned_spend_can_happen` (`PoolOwned.lean:331`); `transactOwned_spends_only_owned` (`PoolSafety.lean:229`);
`owned_transact_can_happen` (`PoolSafety.lean:843`). `Pool.lean`, `PoolOwned.lean`, `PoolSafety.lean` and
`TransactOwned.lean` have no `sorry` and no `admit`.

**S2 in the code, written by hand and not derived from the model.** armada-poc P2a `b3bcf88b` (`transactOwned`,
`_validateOwnedTransaction` at `TransactModule.sol:609`, the public `preimageOf` registry at
`PrivacyPoolStorage.sol:248`) and P2b `4081579b`. A Safe holds a pool note in `4e4fcb89`. The tests are Foundry unit
tests: `test-foundry/TransactOwned.t.sol`, and `ShieldedRagequit.t.sol` and `CrowdfundShieldedProceeds.t.sol`, which
are governance tests and now out of scope. No fixturenet gate uses the owned path: armada-nitro
`env/gates/nitro-e2e/prepare.ts:234` uses the plain `transact` only. I found no commit for P3 (the non-upgradeable
deploy).

**The pool Solidity is not yet the proved artifact.** It differs from the model in the recorded departures: the
pause controller, `testingMode`, the `ownerField` truncation (`TransactModule.sol:637`), and the `feeRecipient`
recognition that feeds `cumulativeFees` for the revenue contracts. Under the ruling of 2026-11-02 ("the sol isn't
authoritative") it is a starting point to cut down, not a source that the model must match.

## 4. Order for the pool

1. armada-lean: the `setFeeRatio` step and its theorems; the executable pool form (`docs/pool-executable.md`).
2. armada-poc: rewrite `PrivacyPool` from the model. Keep shield, transact, owned transact, the preimage registry,
   the fee rate with its one setter, and the treasury address. Remove the pause controller, `testingMode`, the
   `feeRecipient` and `cumulativeFees` hooks, and the `ownerField` truncation. S2's P3 and P4 move into this step.
3. evm-proof-toolchain: the Lean → Stele translator, the Stele specification of the pool, and the EquiVM proof of its
   bytecode, after D1 (no opaque precompiles) and C1 (a pinned `solc`).
4. The fixturenet: the deployed pool bytecode equals the proved bytecode, and the deployed keys equal the ceremony
   keys.

## 5. Where this entry is cited

`engineering/build-plan.md` S1, S2, S4 and S5.
