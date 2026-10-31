# 2026-10-31: What the transact circuit can check about nullifiers

## Question

The Pool model's `Constrained.distinct` (`armada-lean/ArmadaLean/Transact.lean`) says that the circuit makes
the input notes of one proof pairwise distinct. `transact_no_double_spend` was proved from it. Is the model
correct, and what does the theorem mean?

## Findings

- The input-distinctness block exists only in commit `e79919d` ("circuit: constrain input notes pairwise
  distinct"), on branch `feat/public-ceremony` (ship-armada/armada-circuits PR #1, open). Upstream `main`
  (`ef2682f`) and the fork's `integrated` (`9699ba9`) have no `IsEqual` in `Transact.circom`.
  `TransactOwned.circom` has no such block on any branch.
- The block compares `leavesIndices[i]` with `leavesIndices[j]` for the inputs of one proof
  (`Transact.circom:145-161` at `e79919d`). The circuit has no input that holds the set of spent nullifiers,
  so it cannot see an earlier spend.
- The contract rejects every double-spend: `_accumulateAndNullify` checks and marks each nullifier in turn
  (`armada-poc` `TransactModule.sol:379-386`). Railgun has the same loop (`RailgunLogic.sol:493-510`,
  Railgun-Privacy/contract `36bcf5ed`). The loop also rejects a nullifier that one proof lists twice.
- A note is its slot: `nullifier = Poseidon(nullifyingKey, leafIndex)` (`Transact.circom:96-100`), and the
  contract keys its map by tree number.
- The model already proves the contract property: `pool_no_double_spend`, `owned_no_double_spend`.

## Rulings (operator)

1. The model follows the newest and best version of the Armada system. The circuit model keeps
   `Constrained.distinct`, pinned to `e79919d`.
2. `transact_no_double_spend` is renamed `transact_inputs_distinct_nullifiers`. Its statement is unchanged.
   The old name claimed more than the statement proves.
3. Document that the circuit cannot stop the same nullifier twice across transactions, and that on versions
   without `e79919d` (and in `TransactOwned.circom`) nothing in the circuit stops it within one proof either.
4. The circuit block is not defence in depth: on a correct contract it adds no independent layer.

## Done

`armada-lean` branch `draft/rename-inputs-distinct` (`d292661`): the rename, one `sigs-baseline.tsv` row, and
the documentation in `Transact.lean`, `Nullifier.lean`, `TransactOwned.lean`, `README.md`,
`docs/what-the-proofs-mean.md`, `docs/how-to-verify.md`, `docs/axioms-circuit.txt`. `check.sh` passes.

## Open

- The newest Armada system needs one circuit branch: the fork's `integrated` (36 commits on `main`, including
  the owned circuits) does not have `e79919d`.
