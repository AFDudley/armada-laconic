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

## The EVMLean precompile defect: confirmed, not filed, not fixed (agent `EvmleanPrecompile`)

- `totallySafePerformIO` is a plain `def` with body `Inhabited.default` and the attribute
  `@[implemented_by unsafePerformIO]` (EVMLean `Ethereum/PerformIO.lean:6-7`, local `bcca6b1`). The kernel sees only
  `Inhabited.default`.
- `theorem (d : String) : blobSNARKV d = "" := rfl` and the same for `blobECDSARECOVER` type-check. `#print axioms`
  lists only `propext`, `Classical.choice`, `Quot.sound`.
- Upstream: `lefterislazar/EVMLean` and `argotorg/EquiVM` have no issue or pull request about it (searched all states
  for totallySafePerformIO, SNARKV, ECDSARECOVER, ecrecover, precompile, performIO, rfl). The upstream EVMLean
  default branch (`a91ce83`) still wraps both calls in `totallySafePerformIO` with the same body.
- Per the operator's answer to O6, nobody is notified. A fix, if we make one, goes upstream as a pull request.

## Correction: O2 as first explained was wrong (coordinator, 2026-10-28)

The coordinator told the operator that with go-nitro fingerprints an adversary can make an honest payout stale in
every round through credits, so the payout Liveness theorems fail at bytecode level, and recommended that the
adjudicator store whole outcomes. That was wrong:

- A transfer rewrites only the paying channel's fingerprint (`MultiAssetHolder.sol:205`, `channelId` is
  `fromChannelId`). A credit into a channel only adds to `holdings[asset][destination]` (`:448`). So a credit does not
  change the receiving channel's fingerprint, and go-nitro computes the payout from the holdings at run time.
- So a fingerprint-checked transfer on channel `c` goes stale only through a payout from `c` or a reclaim against `c`.
  Each entry is paid in full once, so the count is bounded by `c`'s entries and guarantees. The payout-griefing entry
  already says this (`labbook/2026-10-27-payout-griefing.md:14`).
- After finalization no deadline runs. A bounded number of stale calls delays a payout and costs gas. It does not stop
  the payout. The new outcome is public in the calldata of the call that made the old one stale.
- The unbounded staleness through credits belongs to the Armada payout: its pool-spend proof fixes the amounts and
  reads the Merkle root at call time (`PoolSpend.sol:87-92`). Store-or-fingerprint does not change that.
- The design report's claim (`2026-10-28-liveness-refinement-design.md` §(e) O2: "P-ENC fails … the Liveness theorem
  types change") is wrong for the same reason. The model's stored outcome is an abstraction of the fingerprint: the
  representation map relates them under the existing no-collision hypothesis. The bytecode keeps go-nitro's
  fingerprint. The game-transfer argument must allow a bounded number of failed honest calls (at most the entries plus
  the guarantees of the channel), instead of requiring that every honest call succeeds.

## O3: which rules a fund-holding channel runs (answered from the record, not a question)

The coordinator asked the operator whether a hub-user channel runs the consensus rules or N5's hidden-balance app.
The decision record answers it, and the coordinator did not search it first:

- `labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md` decisions 13-15: hidden channel balances need a
  circuit, and a generic ZK `IForceMoveApp` holds the rules, because `stateIsSupported` is the only place app logic
  lives. Decision 20: hubs swap `waUSDC` inside a channel in private. `build-plan.md` S3: "a generic ZK
  `IForceMoveApp` lets a channel hide its balances"; C5 gates a hidden-balance `waUSDC` swap.
- So a channel that holds funds with hidden balances runs the ZK app (N5's hidden-balance ledger circuit through
  `ZkForceMoveApp`), not `consensusRules`.
- Consequence: the approved Nitro Liveness theorems assume `apps c.appDef = consensusRules` (10 theorems) or
  `satpVirtualRules`. They do not cover a hidden-balance channel. The Armada-layer rebuild (`armada-nitro-spec`) must
  state the liveness of hidden-balance channels, and at bytecode level that needs the ZK app's own certificate, with
  Groth16 trusted. This is a finding for the rebuild.

### Correction to the O3 consequence above

The consequence above overstated the gap. Classified by the rules each theorem assumes (armada-lean `integrated`,
`ArmadaLean/Nitro/Liveness.lean`): 19 theorems hold for any `apps` (payouts, Turbo, the funding and sub-channel
theorems, and the general dichotomy `finalizable_or_two_enabled` / `single_enabled_is_finalizable` of Nitro 2019
§4.4); 5 read no app; 6 assume `consensusRules`; 4 assume `satpVirtualRules`. The Armada layer
(`armada-nitro-spec`, `ArmadaNitro/Liveness.lean`) states its payout and funding theorems for any `apps` too, and fixes
`satpVirtualRules` only in the two SATP theorems. So a hidden-balance channel is covered by every app-generic theorem.

Not covered for a hidden-balance channel: the 6 consensus-specific finalization theorems (`latest_state_finalizable`,
`conclusion_proof_finalizable`, `consensus_enabled_outcomes`, `late_observer_can_checkpoint`,
`concluded_universally_finalizable`, `participant_can_prevent_cycles`). The papers state those for the consensus game
only. A ZK-app analogue would be a new theorem about the app's own rules. Whether one is wanted is open. The earlier
statement "10 theorems assume `consensusRules`" was wrong: 6 assume `consensusRules` and 4 assume `satpVirtualRules`.

Later change (operator, 2026-10-30): `finalizable_or_two_enabled` and `single_enabled_is_finalizable` no longer hold
for any `apps`. They take `(happ : (apps c.appDef).NoUnilateralSupport)`, because the dichotomy is false for an app
that supports states with no signature (`labbook/2026-10-29-machine-types-for-approval.md`, "App property"). An app
that a hidden-balance channel runs must prove `NoUnilateralSupport` before these two theorems apply to it.

## Ruling: no limit on inputs or outputs below the system's hard limits (operator, 2026-10-28)

"The hard limits of inputs and outputs are technical constraints of the system, we shouldn't needlessly constrain it
more than that."

- The pool's verifying-key registry maps `(nullifierCount, commitmentCount)` to a key and sets no bound of its own
  (armada-poc `VerifierModule.sol:14,37-46,55`). The largest standard shape that armada-circuits builds is 8 inputs and
  4 outputs (`scripts/compile.sh` `SHAPES`, entry `"8,4"`). The approved Lean model admits `N ≤ 8` for an owned spend.
- Owned circuits: one shape for every `N` from 1 to 8 inputs and every `M` from 1 to 4 outputs, the same bounds as the
  standard set. A shape that does not fit a hard limit (the ceremony's powers of tau, the transaction gas cap of
  2^24) is reported with the measurement, not dropped silently. `arx-bf1` asked only for `N`x1 and is replaced.
- Armada layer (design choice D1 of `armada-nitro-spec` `docs/armada-nitro-theorems.md`): no limit on the notes per
  channel below what the pool executes. The rebuild applies this.
