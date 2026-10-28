# Verification chain: trusted primitives, the pool, and liveness (2026-10-28)

Operator rulings and findings on the chain in `engineering/build-plan.md` "Verification chain".

## Rulings

1. **The zero-knowledge proof verifier stays trusted.** Step 4 does not refine the Groth16 verifier to the
   specification's `VerifyAccepts`, `VerifyOwnedAccepts` and `EdDSAVerify` primitives.
2. **Poseidon stays trusted** ("yes assume poseidon is trusted").
3. **The new pool comes from the existing armada-poc pool,** probably by removing features.

## Poseidon finding (agent `PoseidonCheck`, checked by the coordinator)

- The pool calls Poseidon at `MerkleModule.sol:45` (`PoseidonT3`), `ShieldModule.sol:377` and `TransactModule.sol:442`
  (`PoseidonT4`), and `ShieldedRagequit.sol:204`.
- `armada-poc/contracts/privacy-pool/types/PoseidonLibs.sol` is a stub. Each body is
  `revert("…library not linked")`. The deployed code is `armada-poc/lib/poseidon_bytecode.json`, which `circomlibjs`
  generated (`_legacy/scripts/generate_poseidon.cjs`). No Solidity source and no Lean or Stele definition of the
  Poseidon permutation exist in the ship-armada repositories, EquiVM or EVMLean.
- So step 4 treats the Poseidon library as a trusted primitive at the call boundary (ruling 2).

## Liveness and "can happen" theorems

The full report is `labbook/2026-10-28-liveness-refinement-design.md` (agent `LivenessRefinement`, Opus).

- An EquiVM certificate (`EquiVM/Stele/Refine.lean:163-183`) is a one-way inclusion. Every bytecode run runs out of
  gas or matches some Stele run. It preserves safety theorems. It does not preserve "can happen" or liveness theorems.
- Recommended design: prove each liveness theorem once, in Lean. One generic game-transfer theorem carries the results
  to bytecode. It uses one two-way lemma per operation, composed from a translator lemma (step 1) and a gas-indexed
  EquiVM certificate (step 4).
- Finding: in EVMLean, `blobSNARKV` and `blobECDSARECOVER` run Python through `totallySafePerformIO`
  (`Ethereum/SNARKV.lean:6-12`, `Ethereum/EllipticCurves.lean:10-16`). The report states that the Lean kernel reduces
  both to `""`. Then every Groth16 check fails in the model, and the pool's safety certificate says nothing about a
  successful `transact`. The coordinator read the definitions. The coordinator did not re-run the `rfl` check.
- The report lists 7 open questions for the operator (§(e)): bounds on input sizes, stored outcomes versus go-nitro
  fingerprints, built-in versus external application contracts, token liveness, the target chain's gas limits, a report
  of the precompile defect to the EVMLean and EquiVM maintainers, and whether adversary-enabledness theorems must reach
  bytecode.

## Operator answers to the open questions (2026-10-28)

- **O1, input-size bounds.** They need EVM measurements. They are decided when the gas of each operation is measured,
  not before.
- **O4, token liveness.** The proofs trust the token contract. This is residual trust: a token with a blocklist or a
  pause (for example USDC) can block a payout, and no proof removes it.
- **O5, gas limits.** Use current mainnet. Measured on 2026-10-28: block 26291702 has `gasLimit` 60,000,000
  (`eth_getBlockByNumber` on `ethereum-rpc.publicnode.com`). The per-transaction cap is 2^24 = 16,777,216 gas
  (EIP-7825, active since Fusaka). Every operation's gas threshold must fit the transaction cap. The cap is the
  binding limit. Whether a later fork changed the cap was not checked `[INFERENCE]`.
- **O6, the EVMLean precompile defect.** Do not notify the maintainers. Check that nobody has filed it. If we fix it,
  we submit the fix.
- **O7, adversary-enabledness theorems.** They stay at Lean level. The simulation and end-to-end tests check them
  again on the deployed code.
- **O2 and O3** wait for a fuller explanation to the operator.
