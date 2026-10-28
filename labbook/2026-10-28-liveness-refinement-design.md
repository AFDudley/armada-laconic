# Liveness and "can happen" theorems through the verification chain

Inputs read: EquiVM at `3defd676` (branch `lean-v4.33.1`), EVMLean (local checkout), `armada-lean-nitro-proofs` at `94436c3` (tag `nitro-proofs-approved-2026-10-28`), `armada-lean-pool-symbolic` at `3128085`, `armada-nitro` (branch `integrated`), `armada-poc`. No file was changed.

## (a) What an EquiVM certificate proves today

A1. **Shape.** `runtimeRefinementFor` has two constructors (`Stele/Refine.lean:163-183`). `execution`: the EVM result `Ξ σ σ₀ g A I` and some `steleExec` result are related by `execResultsEquiv`. `outOfGas`: the EVM result is `.error .OutOfGass`, and the Stele side is unconstrained (`:180-183`).

A2. **Direction.** The Stele result in `execution` is an implicit (existential) argument (`:172`). So the certificate is an inclusion in one direction: every bytecode behaviour is out of gas or is a Stele behaviour. `Stele/Behaviors.lean:52-67` states it as `runtimeRefinement_behaviors_included`. No theorem or definition states the converse. The certificate is not a bisimulation.

A3. **Quantification.** `runtimeRefinement` quantifies over every `σ σ₀ g A I` with `I.code = bytecode` and `I.calldata.size < 2^256` (`Refine.lean:221-233`). `contractRefinement` adds the constructor (`:332-353`).

A4. **Reverts are covered, payloads are not.** An EVM revert matches only a Stele `.reverted`, `.noMatch` or `.decodingFailed` (`Refine.lean:91-104`). `INVALID` matches `.reverted` (`:107-110`). A static violation matches `.staticViolation` (`:115-118`). The revert payload is unconstrained (`:95-104`). `runtimeRefinement_no_crash` excludes every other exception (`Behaviors.lean:118-135`).

A5. **Not compared.** Gas use, logs, substate and revert payloads are not compared (`README.md:245-247`). Stele `emit` does not model the log (`Stele/Semantics/Exec.lean:567-570`).

A6. **Gas.** Stele tracks no gas (`Stele/Syntax/Basic.lean:263`). `GasBound` narrows the covered starting gas (`Refine.lean:190-216`, example `g.toNat < 2^64`). It is a restriction on coverage, not a guarantee of success. The source marks non-termination as a TODO: "non-terminating EVM programs are currently equivalent to any spec" (`Refine.lean:159-160`, `:181`). EVMLean proves that execution with fuel derived from gas cannot fail with `OutOfFuel` (`EVMLean/Ethereum/Theory/README.md:17-19`).

A7. **Stele is nondeterministic in three places** (`Stele/Interp.lean:22-25`): the gas and input substate of an external call (`callViaEVM.callMade`, `Stele/Semantics/Calls.lean:46-79`, comment `:49-53`), the same for `new` (`Calls.lean:153-178`), and the word of `letGas` (`Exec.lean:190-194`). An external call that returns `false` makes the statement `.reverted` (`externalCallFailure`, `Exec.lean:362-368`).

A8. **No determinism theorem exists.** A search for `det|unique|functional|inj` theorems in `Stele/`, `Reasoning/`, `Proofs/` found only word and selector injectivity lemmas and `dispatch_unique` (`Reasoning/Dispatch.lean:185`).

A9. **Derived corollary (valid today).** EVMLean's `Ξ` is a function, so each input has one EVM result. Take an input where every `steleExec` run ends in `.returned _ st _`. Then, by A1 and A4, the EVM result is out of gas, or success with account map `st.accountMap`. No EquiVM change is necessary for this corollary. [INFERENCE: I derived it from `Refine.lean:82-118`; no file states it.]

A10. **The corollary fails for any body with an external call.** By A7, a Stele run exists in which the callee gets a small `callGas` and fails. Then `.reverted` is a Stele behaviour, so a bytecode that forwards too little gas to its callee still refines the spec. Example: `Ξ_ECREC` with less than 3000 gas returns the empty account map, which is a failed call (`EVMLean/Ethereum/PrecompiledContracts.lean:31-34`). [INFERENCE for general callee code: an account with code fails with zero call gas.]

A11. **Gas-exact reasoning exists only on another branch, only for pure code.** Commit `bbb6c42b` ("Add Precompile-focused, exact-gas RDx proof framework") is on `evmification`, not on the checked-out `lean-v4.33.1` (`git branch -a --contains bbb6c42b`). Its `ExactGasPost` states "gas below the cost gives failure; gas at or above the cost gives this exact result" (`git show bbb6c42b:Reasoning/Bytecode.lean`, lines 173-186). Its scope is code with no storage effects and no external calls (`Misc/precompile_prompt.md:31-41`). `ReachExact.lean` "deliberately provides no rule for `GAS` or `CALL`" (`git show bbb6c42b:Reasoning/ReachExact.lean`, lines 5-16). The current library's `RD` keeps an out-of-gas disjunct at every step (`Reasoning/Reach.lean:8-16`). So `RD` proves partial correctness only.

A12. **EVMLean's ecrecover and pairing precompiles are constants in the Lean logic.** `totallySafePerformIO` is defined as `Inhabited.default`; `@[implemented_by unsafePerformIO]` changes only compiled code (`EVMLean/Ethereum/PerformIO.lean:6-7`). `blobECDSARECOVER` and `blobSNARKV` call it (`EllipticCurves.lean:10-16`, `SNARKV.lean:6-12`). The kernel accepted these checks (`lake env lean --stdin` in EVMLean, exit 0):
- `example (d : String) : blobSNARKV d = "" := rfl`
- `example (e v r s : String) : blobECDSARECOVER e v r s = "" := rfl`
- `example (d : ByteArray) : (SNARKV d).toOption = none := by unfold SNARKV; rw [show blobSNARKV (toHex d) = "" from rfl]; decide`
- `example (e v r s : ByteArray) : ECDSARECOVER e v r s = ByteArray.ofBlob (padLeft 128 "") := rfl`

Consequences. In the model, every BN254 pairing call fails (`Ξ_SNARKV` error branch, `PrecompiledContracts.lean:310-328`). Every in-range ecrecover returns one fixed address. So no bytecode path after a Groth16 verification is reachable in the model. A safety certificate for the pool's `transact` therefore says nothing about the real success path. A completeness hypothesis ("an honest proof verifies", "an honest signature recovers its signer") contradicts these definitions. A theorem that assumes it is vacuous. EquiVM uses these precompiles unchanged (`Reasoning/ExternalCall.lean:770`).

**Summary of (a).** An EquiVM certificate proves: on every input, the bytecode runs out of gas, or its result is the result of some Stele run, with equal account maps on success and matched revert kinds. It is a one-way inclusion. It does not cover gas, logs or revert payloads. It does not constrain the gas that the bytecode forwards to callees. It preserves safety. It does not preserve enabledness.

## (b) What each class of theorem needs from the concrete machine

Facts about the abstract game:
- Every Nitro operation is `Option`-valued; `none` is a revert (`Nitro/Adjudicator.lean:9-11`, `:289-299`). `Step` is a successful `apply` (`:301-304`).
- The honest participant chooses its move at the start of a round; the adversary moves; then the honest move applies (`Nitro/Strategy.lean:12-18`, `UnbeatableIn` `:172-178`). The move's arguments are fixed at choice time, and the operation reads the state at application time.
- A failed honest operation changes nothing: `(apply …).getD net.chain` (`Strategy.lean:84`).
- The adversary can play any operation by another caller that carries no forged signature (`Capable`, `Strategy.lean:150-158`), and it can move the clock within the round bound `ε` (`Threat.Round`, `:125-127`).
- The model has no gas or duration measure (`docs/nitro-theorems.md:263`). It stores outcome values, not fingerprints (`docs/nitro-theorems.md:292`, docs-0080-02).
- The pool `Step` is an inductive relation with witness hypotheses (`hC : Constrained w pub`) and ghost fields (`live`, `published`, `payouts`) (`armada-lean-pool-symbolic/ArmadaLean/Pool.lean:874-883`, `:970-1075`).

Concrete properties, named for use below:
- **P-SIM** (forward simulation): every concrete transaction that changes the contracts' storage is an abstract `Step` under the representation map R, or it changes nothing abstract. Every contract that can write the storage is in scope.
- **P-EN** (enabledness with equal effect): if the abstract operation returns `some s'`, the concrete call succeeds and its state relates to `s'` by R.
- **P-GAS**: P-EN holds for every starting gas at or above a bound G(op), and G(op) fits the block and transaction gas limits. G(op) depends only on the calldata, not on state that the adversary can grow. Access lists give no help, so G covers cold accesses.
- **P-ENC** (state-independent encoding): the honest participant can build the calldata at choice time from what it holds. The same calldata stays valid after any adversary moves of the round wherever the abstract operation still succeeds.
- **P-CALLEE**: each callee on the path (pool, application contract, token, precompile) succeeds with its specified effect when it gets its own G.
- **P-CLOCK**: `now` maps to `block.timestamp`. A transaction sent at time t is included before t + ε (Nitro:537-539, the hypothesis that the game already carries).
- **P-NOEXTRA**: the concrete contracts have no entry point outside the abstract operation set (admin, pause, blocklist, key setters, upgrades).

| Class | Examples | Needs |
|---|---|---|
| A. Safety (∀ step ⇒ property) | most of `Safety.lean`, `PoolSafety.lean` | P-SIM. EquiVM plus a forward step-1 proof give it. |
| B. One step is possible | `deposit_can_happen` (`Safety.lean:89-92`), `transfer_can_happen` (`:386-389`), `reclaim_can_happen` (`:1195`), `transact_can_happen` (`PoolSafety.lean:647`), `owned_merge_can_happen` (`:1864-1880`) | P-EN and P-GAS for that operation, P-CALLEE, and the completeness hypotheses (`TransactProofCompleteness`, `PoolSafety.lean:552-554`; `HonestSignature`, `:562-564`). |
| C. A strategy exists | `directly_funded_pays`, `conclusion_proof_finalizable`, `latest_state_finalizable`, `Finalizable`, `CanObtain`, `FundedFor`, the lower half of `Value` (`Liveness.lean:45-131`, `Strategy.lean:217-256`) | P-EN, P-GAS, P-ENC for each honest move; P-SIM and P-NOEXTRA for each adversary move; P-CLOCK. |
| D. No strategy exists, or an outcome is "enabled" | `turbo_value_upper_bound` (`Liveness.lean:289-294`), `consensus_enabled_outcomes` (`:221-233`), `finalizable_unique` (`:137-140`), the upper half of `Value`, `deposit_bounds_value` | Class C in the other direction: P-SIM for each honest move, and P-EN and P-GAS for each adversary move, because the abstract counter-strategy of the adversary must be playable concretely. |
| E. Non-interference | `spend_not_blocked` (`PoolSafety.lean:1797-1804`) | P-SIM, so that every concrete step is a `Step`; P-NOEXTRA; P-EN and P-GAS for the pending spend at the later state. |

Defects of these kinds in the code that the rewrite replaces:
- go-nitro `transfer` takes `outcomeBytes` and checks them against the stored fingerprint (`armada-nitro/src/nitro/MultiAssetHolder.sol:59-62`, `:128`). A transfer by the adversary changes the stored outcome, so the honest call with the earlier bytes reverts. The abstract `transfer` reads the outcome at application time (`Adjudicator.lean:140-152`). This violates P-ENC.
- `PoolSpend.mintOwnedCommitments` sends a zero proof and reads `pool.merkleRoot()` at call time (`armada-nitro/src/armada/libraries/PoolSpend.sol:87-92`). The first makes every payout revert under a real verifier (also recorded at `build-plan.md:119`). The second violates P-ENC. The EVMLean model (A12) rejects every proof, so neither a safety certificate nor the model shows the defect.
- The armada-poc pool has `addToBlocklist`, `setVerificationKey`, `setTestingMode`, `setShieldPauseContract` and other admin entry points (`armada-poc/contracts/privacy-pool/PrivacyPool.sol:304-476`). Each violates P-NOEXTRA for class C, D and E theorems.

## (c) Designs evaluated

### D1. Two-way simulation (bisimulation) at step 1 and step 4

- Translator emits: a two-way per-operation relation between Lean and Stele.
- EquiVM must prove: every Stele run is matched by a bytecode run. This statement is false as written. Stele chooses the callee gas freely (A7), and the bytecode has one run per input. Stele has no gas, so a bytecode out-of-gas result has no Stele counterpart. A correct version must index the Stele side by gas and remove the free gas choice. That version is the step-4 obligation of D2/D4.
- Expressible today: no. It replaces `Refine.lean`, which is in the trusted base (`README.md:245-247`).
- Cost: a new core relation, and the existing proof library is built for the one-way relation.
- Residual: the same gas and callee residual as D2.
- At step 1 alone, a two-way relation is cheap: the translator generates the Stele body from the Lean function, so both directions come from one construction. D4 keeps this part.

### D2. Forward simulation plus an enabledness obligation at each step

- Translator emits: the forward refinement and, per operation, "Lean returns `some s'` on the domain ⇒ every Stele run with adequate callee results returns a state related to `s'`".
- EquiVM must prove: per entry point, "Stele returns `st` and starting gas ≥ G ⇒ the bytecode succeeds with `st.accountMap`, with each callee given at least its own G".
- Expressible today: step 1 yes, step 4 no (A9, A10, A11).
- Cost: two lemmas per operation (6 Nitro entry points, 3 pool entry points). Step 4 needs new gas reasoning.
- Residual: chain inclusion, gas limits, callee liveness, completeness of the cryptographic primitives.
- Gap: D2 gives per-operation facts only. Each game theorem still needs its own transfer argument.

### D3. Translate the liveness game to Stele level

- Translator emits: the game (network states, strategies, rounds) over Stele states with concrete encodings, and a proof of each theorem there.
- EquiVM must prove: the same step-4 obligation as D2. Stele has only a one-message judgment (`steleExec`, `Exec.lean:719-791`), so the game needs a new multi-transaction layer.
- Expressible today: no.
- Cost: highest. It proves each Liveness theorem a second time over bytes, Keccak and ECDSA values instead of symbolic terms. The theorem types are approved at Lean level, so a Stele-level copy needs a second approval.
- Residual: the same as D2.

### D4. Liveness proved only at Lean level, with one per-operation lemma from Lean to bytecode

- Translator emits: per operation, a two-way step-1 lemma: on the domain Dom, the Stele run returns a state related to `s'` exactly when the Lean operation returns `some s'`.
- EquiVM must prove: the D2 step-4 obligation. The composition of the two lemmas is "enabled in spec ⇒ succeeds in bytecode".
- Generic part, proved once at Lean level: a game-transfer theorem. The abstract and concrete games relate by an alternating simulation: honest moves map in one direction, adversary moves in the other. The two-way per-operation lemma serves both directions, so classes C and D both transfer. [INFERENCE: the general result is the alternating refinement of Alur, Henzinger, Kupferman and Vardi (1998). I did not read it.]
- Expressible today: step 1 yes. Step 4 no. The game-transfer theorem is plain Lean over the existing `Strategy.lean` definitions.
- Cost: two lemmas per operation and one generic theorem. No Liveness theorem is proved twice.
- Residual: as D2.

### D5 (a choice inside D2 and D4). The form of the step-4 obligation

- D5a, a separate liveness theorem next to the existing certificate. It must restate the Stele run because A2 gives only "some" run. The agent symbolically executes each path twice.
- D5b, a gas-indexed certificate that replaces the existing one. Each transition declares a threshold Γ_f, and each call site declares a floor γ. When the starting gas is at least Γ_f, the strengthened judgment requires `callGas ≥ γ` and forbids the out-of-gas case. The existing `runtimeRefinement` follows from it as a corollary, because the strengthened judgment is a sub-relation and the out-of-gas case stays allowed below Γ_f. Stele already has the starting gas: `steleExec` sets `machineState.gasAvailable := .ofUInt256 g` (`Exec.lean:754`). One proof per contract gives both safety and liveness. [INFERENCE: feasibility rests on `RD` already counting the used gas `C` (`Reach.lean:8-16`) and on RDx (A11). Neither has a `CALL` rule.]

## (d) Recommended design

**D4 with the D2 decomposition and a D5b step-4 certificate.** Prove every liveness theorem once, at Lean level. Transfer the theorems to bytecode with one generic game-transfer theorem. Feed that theorem one two-way lemma per operation. Compose each lemma from a translator-emitted step-1 lemma and a gas-indexed EquiVM certificate.

### Proof obligations per chain step

**Lean level (specification repositories).**
- L1. Add the domain predicate Dom (O1 lists what it must contain) as a hypothesis to each `*_can_happen` theorem and to each honest-move enabledness that a liveness proof uses. Dom contains `uint256` and `uint120` ranges, the participant count, the allocation length and the support-proof length. Each bound is a new guard (`docs/process.md:182-184`), so each needs operator approval.
- L2. Give the pool an executable operation form (`PoolState → Call → Option PoolState`), with `VerifyAccepts` as a primitive. The translator takes executable functions (`build-plan.md:131`). The current pool `Step` is relational (`Pool.lean:970-1075`).
- L3. Prove the game-transfer theorem once. Input: a concrete game over EVM account maps, R, the per-operation two-way lemma, and P-CLOCK. Output: `UnbeatableIn` transfers from abstract to concrete, and from concrete to abstract for class D.

**Step 1 (translator).**
- T1. Forward: a returning Stele run gives `apply = some s'` with R. A reverting run gives `none`.
- T2. Backward: on Dom, `apply = some s'` gives "every Stele run, with each callee result fixed to its specified success, returns `st` with R s' st". The translator generates the body, so T1 and T2 come from one structural proof. The Stele-level "every run" style already exists (`Proofs/ERC20/API.lean:30-34`).
- T3. Named hypotheses: no Keccak or Poseidon collision (existing), ECDSA completeness (an honest signature recovers its signer), and Groth16 completeness (`TransactProofCompleteness` and `OwnedProofCompleteness`; this is the trusted primitive of the 2026-10-28 ruling).
- T4. The calldata encoding is a relation `Enc(op, calldata)` that the honest participant can satisfy from its `NetworkState` (proofs, signatures). It must not include state that another caller can change (P-ENC). The translator rejects a Lean operation whose Stele form needs such data.

**Step 2 (transliteration) and step 3 (solc).** No proof, because step 4 checks the result. The Solidity form must keep the properties that step 4 must prove:
- S1. External calls forward all gas except fixed reserves. `SnarkTypes.sol:78` and `:93` already use `staticcall(sub(gas(), 2000), …)`.
- S2. No branch depends on `gasleft()` on a liveness path, except a declared floor such as the EIP-1271 cap (`NitroUtils.sol:93-94`, `:119`).
- S3. Every loop bound comes from Dom.
- S4. No entry point exists outside the Lean operation set (P-NOEXTRA). For the pool, the feature-removal list must include every `PrivacyPool.sol:304-476` setter that the Lean `Step` does not model.
- S5. The off-chain client reads proof states from calldata, not from events, because EquiVM does not compare logs (A5). The repaired game reads them from calldata (`Strategy.lean:88-96`).

**Step 4 (EquiVM, agent-written).**
- E1. The gas-indexed certificate (D5b) for each contract: adjudicator, pool router and modules (the router uses `delegatecall`, `PrivacyPool.sol:20-25`), and each application contract that a liveness theorem names (`apps c.appDef = consensusRules`).
- E2. For each transition f: Γ_f over Dom, with all accesses cold, and a proof that Γ_f plus the intrinsic calldata gas fits the target chain's limits.
- E3. Callee lemmas at the `Θ` boundary: the pool for the adjudicator, the application contract, the token, and the precompiles 0x01 and 0x06-0x08.

**EquiVM and EVMLean prerequisites (owners: EquiVM and EVMLean maintainers).**
- V1. Make `ECDSARECOVER` and `SNARKV` opaque, so that no proof can depend on a fixed value (A12). State ECDSA and pairing correctness as named hypotheses over the opaque functions. Without V1, T3 is inconsistent with the model, and the pool's safety certificate is empty on the `transact` success path.
- V2. Port RDx to `lean-v4.33.1`, and add `CALL`, `STATICCALL`, `DELEGATECALL` and `SSTORE` gas rules with the 63/64 forwarding rule.
- V3. Add the gas-indexed judgment and certificate (D5b) in a new file, and prove the existing `runtimeRefinement` as its corollary.
- V4. The EVM locality property that P-SIM needs is work in progress in EVMLean (`EVMLean/Ethereum/Theory/README.md:32-37`).

### Residual trust after the design

- Chain inclusion within ε, and a monotone `block.timestamp` (P-CLOCK). This is already a game hypothesis. No proof can remove it.
- The honest participant holds enough ETH to pay gas, and its transaction fits the gas price market. The model has no wallet (`Strategy.lean:69-72`).
- The faithfulness of EVMLean to the deployed EVM (`README.md:241-244`), and of the opaque precompiles after V1.
- ECDSA and Groth16 completeness (T3), and Keccak and Poseidon collision freedom.
- The token contract's liveness, unless the token's bytecode also gets E1 and E3.

### Why the other designs are worse

- D1: as written, its statement is false (A7, A6). Its correct form is the D4 step-4 obligation plus a change to the trusted relation.
- D2 alone: it gives the same lemmas as D4 but no game transfer, so each of about 30 Liveness theorems needs its own transfer argument.
- D3: it proves each game theorem twice, at a harder level, and the Stele-level types need a new approval. It still needs the D4 step-4 obligation.
- D5a instead of D5b: the agent proves each contract twice. The two proofs can diverge, because A2 lets the safety proof pick a different Stele run.

## (e) Open questions for the operator

O1. Dom bounds. Which maximum participant count, allocation length, support-proof length and note-input count do you approve as new guards in the Lean operations? Without bounds, gas grows with input size, and the `*_can_happen` theorems are false at bytecode level for large inputs.

O2. Stored fingerprints. Does the rewrite store outcome values, as the model does (docs-0080-02), or go-nitro fingerprints? With fingerprints, P-ENC fails for `transfer` and `reclaim` under front-running. Then the Lean operations must take the expected outcome as an argument, and the Liveness theorem types change.

O3. Application contracts. The Liveness theorems assume `apps c.appDef = consensusRules`. Should the rewrite build the consensus rules into the adjudicator, or call an external application contract (`ForceMove.sol:212`)? An external application needs its own E1 and E3.

O4. Token liveness. Do the payout theorems trust the token contract, or do you restrict the protocol to tokens with their own certificates? A blocklist or pause in the token blocks payouts.

O5. The target chain. Which block gas limit and transaction gas limit must Γ_f fit?

O6. The EVMLean precompile defect (A12, V1). Do you approve a report to the EVMLean and EquiVM maintainers? Until it is fixed, the pool's safety certificate is vacuous on the `transact` success path.

O7. Class D theorems. They need enabledness of adversary moves (row D of the table). Do you keep them in the set that must reach bytecode, or keep them as Lean-level statements only?
