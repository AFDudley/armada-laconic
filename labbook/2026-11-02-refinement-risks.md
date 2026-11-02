# Four refinement risks: facts, existing plans, the operator's ruling, and the decisions (2026-11-02)

Source of the four risks: a claude.ai review of armada-lean `integrated`, transcript
`.scratch/claude-share-46520279.clean.txt` (lines 757-908). The review proposed a fix for each risk. This entry checks
each claim against the code and against the existing plans. It records what the plans already contain, what is new,
and what needs an operator ruling. It replaces the text of commit `24b9524` (process.md rules 9.5-9.8). That commit
stated the review's proposals as rules, with no check against the code or the plans.

Pins used for every citation: armada-lean `integrated` `7ee616ac`; EquiVM `lean-v4.33.1` `3defd676`; armada-nitro
`integrated` `295b6c2e`; armada-poc `integrated` `fbcf2b1e`; armada-circuits `integrated` `9699ba9`. A claim with no
citation that I did not observe has the mark `[INFERENCE]`.

Plans read in full for this entry: `docs/process.md`, `engineering/build-plan.md`, `docs/lean/type-doctrine.md`,
`docs/lean/verification.md`, `docs/lean/proof-dispatch-plan.md`, `docs/lean/typed-library.md`,
`docs/lean/audit-2026-10-31.md`, `notes/handoff.md`, `engineering/11-risks-and-technical-debt.md`,
`labbook/2026-10-28-liveness-refinement-design.md`, `labbook/2026-10-28-verification-chain-rulings.md`,
`labbook/2026-11-01-l2bridgeapp-verifier.md`; armada-lean `AGENTS.md`, `docs/verification-chain.md`,
`docs/how-to-verify.md`, `docs/spec-module-design.md`. Sections read: `docs/lean/writing-lean.md` §4, §10, §11;
`docs/lean/worker-briefs.md` "Executability".

## Defects of `24b9524`

1. It stated four proposals as rules. Risk 1 needs an operator ruling (§1 below). Risks 3 and 4 change the scope of the
   verification chain, which is the operator's decision (§3, §4).
2. Its rule 9.6 cited "open questions O1 and O5 of the design" as open. The operator answered both on 2026-10-28.
   O1: the input-size bounds wait for gas measurements. O5: current mainnet, and the binding limit is the
   per-transaction cap of 2^24 gas (EIP-7825), not the block gas limit (`2026-10-28-verification-chain-rulings.md`
   "Operator answers").
3. Its rule 9.6 required that each honest move "does not revert spuriously". That contradicts the correction of
   2026-10-28: with go-nitro fingerprints, a stale honest call reverts, and the game-transfer argument must allow a
   bounded number of failed honest calls (`2026-10-28-verification-chain-rulings.md` "Correction: O2").
4. Its closing paragraph said that none of the four risks is in the specification layer. That is false. Risk 1 includes
   a model departure that the model itself records (`ArmadaLean/TransactOwned.lean:60-70`). The design's step L1 adds
   size bounds to approved statements (`2026-10-28-liveness-refinement-design.md` §(d) L1). Risk 3 adds
   specification-level definitions.
5. Its rule 9.7 said to build the strategy by filtering candidates through `Decidable`. The predicates are not
   decidable as written (§3, fact 3).
6. Its rule 9.8 said to check the verifying key against "the constants in the deployed verifier's bytecode". The pool
   stores its keys in storage that the owner sets (§4, fact 4). They are not bytecode constants.

## 1. Risk 1: concretization of raw values

The review's claim: in the model `val n ≠ hash f xs` by construction. On chain, a word can equal a hash output. So the
step-1 hypothesis must cover a raw value that equals a hash output, and the map from words back to terms must depend on
the hashes that the run computed.

### Facts

1. `Term` has two constructors, `val : Nat → Term` and `hash : HashFn → List Term → Term`. `HashFn` has two
   constructors, `poseidon` and `keccak` (`ArmadaLean/Poseidon.lean:58-70`). A `val` term is never a `hash` term, and a
   Poseidon term is never a Keccak term, by construction (`Term.poseidon_ne_keccak`, `Poseidon.lean:163`).
2. The Nitro model does not use `Term`. A destination is `Address.external (e : E)` or `Address.channel (c :
   FixedPart)`, with `E` a parameter (`ArmadaLean/Nitro/Address.lean`). The two are disjoint by construction. A
   participant key is an opaque `Nat`.
3. The step-1 hypothesis in the plans is one sentence: "no Keccak or Poseidon collision among the values a run uses"
   (`engineering/build-plan.md` S5, item 2; armada-lean `docs/verification-chain.md:30-31`). The plans call the
   representation map "already known" (`build-plan.md` S5, "How each step is done").
4. These contract rules accept words that the model does not represent, or map one word to two model values:
   - **Destination kind.** `_isExternalDestination` returns true when the top 12 bytes of a `bytes32` destination are
     zero (armada-nitro `src/nitro/MultiAssetHolder.sol:474-476`). A transfer then pays the low 20 bytes as an address
     (`:445-446`). A channel id is a 256-bit Keccak output. A channel id whose top 96 bits are zero is paid out as an
     external address. The model cannot represent this case (fact 2). Grinding a channel nonce to reach it costs about
     2^96 Keccak evaluations (expected trials for a 96-bit zero prefix).
   - **Owner field.** The owned path authorizes with `msg.sender != address(uint160(uint256(ownerField)))` (armada-poc
     `contracts/privacy-pool/modules/TransactModule.sol:637`). The contract accepts any `ownerField` whose low 160 bits
     equal the caller. The model represents only the canonical value, as a Keccak term, and records this as a
     departure (`ArmadaLean/TransactOwned.lean:60-70`). Here the contract allows more than the model. Process rule 9.2
     forbids that.
   - **Address width.** An address is the low 160 bits of a Keccak output. The model treats it as a whole Keccak term
     (`Poseidon.lean` module comment, "DEPARTURE"). By the birthday bound, two creation inputs with one address cost
     about 2^80 Keccak evaluations. Whether that is feasible for an attacker is `[INFERENCE]`, not measured here.
   - **Reduction mod p.** The bound-parameters hash is `keccak256(abi.encode(core)) % SNARK_SCALAR_FIELD` (armada-poc
     `contracts/privacy-pool/storage/PrivacyPoolStorage.sol:56`). A non-ERC-20 token id is the same form
     (`TransactModule.sol:453`). The model keeps these symbolic and says that the reduction "belongs to a refinement
     layer" (`ArmadaLean/BoundParams.lean:6-10`).
   - **Field range of public inputs.** `Term.val` is an unbounded `Nat`. The verifier rejects every public input at or
     above the field prime (`contracts/privacy-pool/types/SnarkTypes.sol:35`), and shield rejects such an `npk`
     (`ShieldModule.sol:246,260`). So for circuit inputs the contract already enforces one canonical word. The decoder
     of the refinement must carry this range. It is not an aliasing case.
   - **Third hash function.** The `L2BridgeApp` design binds public values through SHA-256 digests
     (`labbook/2026-11-01-l2bridgeapp-verifier.md` §8, item 5). `HashFn` has no SHA-256 constructor.
5. A copied hash output (an attacker resubmits a commitment or a root that it saw) is not an aliasing case if the map
   from words to terms turns a word that the run computed as a hash into that hash term. This needs every model
   operation to accept an arbitrary `Term` in each argument position that calldata fills `[INFERENCE: I did not check
   every operation's argument types]`.
6. `Poseidon.lean`'s module comment gives the soundness argument of the symbolic model: an attacker "cannot conjure a
   `hash` term without supplying every one of its arguments". A map that turns a copied word into a hash term gives
   the attacker that term with no arguments. Whether an approved theorem depends on the argument is not checked here
   `[INFERENCE]`.

### What the review adds, and what is wrong in it

- Correct and new: the hypothesis must name more than hash-hash collisions (fact 4). The map must be relative to the
  run (fact 5).
- Wrong as stated: "BN254 field reduction needs the same treatment for Poseidon values". For circuit inputs the
  contract already rejects non-canonical words (fact 4, field range). The reduction that needs a hypothesis is Keccak
  mod p.
- The review's example (a Keccak channel id with 12 leading zero bytes) is correct for the current armada-nitro code.

### Questions (closed 2026-11-02: §7 records the ruling or the decision)

- **Q1.1 Contract rules that accept more than the model.** For the destination-kind rule and the owner-field
  truncation (fact 4), which side changes?
  - (a) The contract: the rewritten adjudicator encodes the destination kind explicitly, and `TransactModule` rejects
    `ownerField ≥ 2^160`. The model then stays as approved.
  - (b) The model: it represents the concrete rule, and the affected statements change.
  - (c) A named hypothesis on each affected theorem: no channel id that the run uses has 96 leading zero bits.
  Recommendation: (a). Process rule 1.2 makes an illegal state unrepresentable. Rule 9.2 forbids a contract that allows
  more than the model. Option (c) keeps a contract behaviour that no theorem covers.
- **Q1.2 The scope of the named cryptographic hypothesis.** A new named assumption needs the operator's approval
  (`docs/lean/verification.md` §3; `writing-lean.md` §11). Which of these does the step-1 theorem take: Keccak-256
  collision freedom; collision freedom of its 160-bit truncation (addresses); collision freedom of Keccak mod p;
  Poseidon over BN254; SHA-256 (for `L2BridgeApp`)?
- **Q1.3 The form of the step-1 refinement.** Approve the run-relative map of fact 5 as the form: each ABI position
  has a typed decoder; a word that the run computed as a hash decodes to that hash term; any other word decodes to
  `val` in its range? A yes also decides that `HashFn` gets a SHA-256 constructor when `L2BridgeApp` is modelled.

## 2. Risk 2: liveness under a one-way certificate

The review's claim: an EquiVM certificate preserves safety only. Its out-of-gas exit is where channel attacks live. The
review proposed a gas bound and a "no spurious revert" lemma for each honest move, and one generic transfer theorem.

### Facts

1. The certificate has two constructors, `execution` and `outOfGas` (EquiVM `Stele/Refine.lean:163-183`). In
   `outOfGas` the Stele side is unconstrained (`:180-183`). An EVM revert matches only a Stele `.reverted`, `.noMatch`
   or `.decodingFailed` (`:91-104`). So EquiVM already tells a revert apart from an out-of-gas result. The review
   could not confirm this. It is confirmed.
2. The existing design already contains the review's proposal, in more detail:
   `labbook/2026-10-28-liveness-refinement-design.md`. P-EN and P-GAS are the review's "no spurious revert" and "gas
   bound". The game-transfer theorem L3 is the review's `liveness_transfers`. The theorem classes A-E are the review's
   classification. The design adds what the review lacks: P-ENC (calldata fixed at choice time), P-CALLEE (finding
   A10: a bytecode that forwards too little gas to a callee still refines the spec), P-NOEXTRA, and the EVMLean
   precompile defect (A12, V1).
3. The operator answered O1, O4, O5, O6 and O7 on 2026-10-28. O2 and O3 were resolved by correction and from the record
   (`2026-10-28-verification-chain-rulings.md`).
4. The O2 correction changes one requirement of the design and of the review: the game transfer must allow a bounded
   number of failed honest calls (at most the entries plus the guarantees of the channel).
5. No record shows that the operator approved the design itself (D4 with D5b). Before this entry, `build-plan.md` S5
   said "Its open questions wait for the operator". This entry updates that line.
6. No plan names an owner or a time for the EquiVM and EVMLean prerequisites V1-V3 (opaque precompiles, `CALL` gas
   rules, the gas-indexed certificate). No plan schedules the gas measurement that O1 waits for.

### What the review adds

- One ordering point: the review ranks the design's open items above the remaining Liveness proofs. The
  proof-dispatch plan already accepted the risk that the design changes how the Liveness theorems are stated
  (`docs/lean/proof-dispatch-plan.md` "Risks that the operator accepted").
- A computed theorem classification (one class per spec theorem, from its statement). This fits the spec module's rule
  that the check stores no list (armada-lean `docs/spec-module-design.md` §1, §7). The review's form, a waiver per
  theorem diffed through `gate-baseline.json`, is a stored list. The operator rejected a ratchet of that kind
  (`docs/lean/type-doctrine.md` §4, decision 3).

### Questions (closed 2026-11-02: §7 records the ruling or the decision)

- **Q2.1** Do you approve the design of `2026-10-28-liveness-refinement-design.md` §(d) (D4 with a D5b certificate),
  with the O2 correction (a bounded number of failed honest calls), as the route for class B, C and E theorems?
- **Q2.2** Who makes the EquiVM and EVMLean changes V1-V3, and before or after the Liveness proofs close? O6 says that a
  fix goes upstream as a pull request.

Next step that needs no ruling: measure the gas of each adjudicator operation on the armada-nitro fixturenet, with
inputs at the hard limits. That is the measurement O1 waits for. The rewrite will change the numbers `[INFERENCE]`,
but the measurement shows which operations approach the 2^24 cap.

## 3. Risk 3: the honest strategy is a relation, not a program

The review's claim: a client (go-nitro, mobile, web) needs an executable strategy, proved to satisfy the strategy
theorems, and checked against each client with conformance vectors.

### Facts

1. The type exists: `Strategy E p := NetworkState E p → HonestMove E` (`ArmadaLean/Nitro/Strategy.lean:795`), with its
   plays `Run` (`:800-805`).
2. The theorems state `UnbeatableIn`, an inductive `Prop` whose `move` case picks the honest move inside the proof
   (`Strategy.lean:769-776`). A proof of a `Prop` gives no program. So an executable strategy is a new definition with
   a new proof.
3. The feasibility predicates are not decidable as written. `NetworkState.HoldsPrefund` quantifies over every signed
   state (`Strategy.lean:98-99`). `HonestSend` takes the predicate `prefundHeld : FixedPart → Prop`
   (`ArmadaLean/Types/Honest.lean:105-117`). Only `GuaranteesExternal` has a `Decidable` instance
   (`Types/Honest.lean:69`). A decision procedure needs these restated as searches over the finite `sent` and
   `received` lists.
4. No plan ties a client to the model. `process.md` "The layers" covers contracts only. The client stages vendor
   ts-nitro (`build-plan.md` S9 CL2-2, S10 C4) and accept on e2e gates. O7 says that simulation and e2e tests check
   the class D theorems on the deployed code.
5. No test checks the representation map ("already known", `build-plan.md` S5) against the contracts.

### Question (closed 2026-11-02: §7 records the decision)

- **Q3.1** Do clients enter the verification chain?
  - (a) No. Clients are accepted on the e2e gates only (the current plan).
  - (b) Conformance vectors generated from the Lean model (ABI encodings, state hashes, valid and invalid transitions),
    run against the contracts with `eth_call` and against each client core.
  - (c) (b), plus an executable strategy in Lean, proved to realize the approved strategy theorems, and ported to the
    clients.
  Recommendation: (b) now. It is a test, not a proof, but it checks the representation map of risk 1 against the real
  contracts, and encoding mismatches are cheap to find this way. (c) after the Liveness proofs close, because its proof
  reuses them `[INFERENCE: cost not measured]`.

## 4. Risk 4: hand-modelled circuits and deployed keys

The review's claim: `Transact.lean` models `Transact.circom` by hand. The fix is an R1CS importer with one lemma per
gadget template, and a check that the deployed key matches the circuit.

### Facts

1. `Transact.lean` cites `Transact.circom` line by line. `docs/how-to-verify.md` §5 asks a person to compare them.
2. The model follows `Transact.circom` at commit `e79919d` (branch `feat/public-ceremony`). That commit is not on
   armada-circuits `main` or `integrated` (`ArmadaLean/Transact.lean:15-19`). One model-to-circuit difference is
   already an open operator item (`notes/handoff.md` "Waits for the operator", item 2, `Constrained.distinct`).
3. `Ledger.circom` and `Battleship.circom` (stages S3 N5 and N6) have no Lean model. A search of `ArmadaLean/` for
   either name finds nothing.
4. The pool stores one Groth16 key per circuit shape. The owner sets it, "with no structural validation"
   (armada-poc `contracts/privacy-pool/modules/VerifierModule.sol:32-46`). Chunk P3 plans constructor keys
   (`build-plan.md` S2).
5. Existing ties: `scripts/compile.sh` emits `.r1cs` and `.sym` for every shape (armada-circuits `:62`). P4 and S13
   accept on `snarkjs zkey verify`, which ties a zkey to its R1CS and the ptau. No plan checks that the key in the
   deployed pool equals the key exported from that zkey.
6. Ruling 1 of 2026-10-28 trusts the Groth16 verifier: step 4 does not refine it to `VerifyAccepts`. The theorems take
   `Constrained w pub` as the soundness hypothesis (`ArmadaLean/PoolOwned.lean:93-99`; `PoolSafety.lean:578-581`). The
   ruling does not say whether the compiled circuit implements `Constrained`, or whether the deployed key belongs to
   that circuit.
7. The `L2BridgeApp` design already pins its verifier and program keys and reproduces them from source
   (`labbook/2026-11-01-l2bridgeapp-verifier.md` §2.1 item 4, §5).

### Questions (closed 2026-11-02: §7 records the ruling or the decision)

- **Q4.1** Does ruling 1 also trust that each compiled circuit implements its Lean model? If not, which check?
  - (a) An R1CS importer into Lean, with one proved lemma per gadget template and a proof for each circuit's own
    constraints (the review's proposal).
  - (b) A differential test: an executable form of `Constrained` against the compiled witness generator, on generated
    inputs. It finds disagreements. It proves nothing.
- **Q4.2** Add to the acceptance of P3 and S13 a check that the key in each deployed pool shape equals the key exported
  from the ceremony zkey of that shape? Recommendation: yes. It is cheap, and `L2BridgeApp` already does the same for
  RISC Zero.
- **Q4.3** Do the S3 circuits (`Ledger.circom`, `Battleship.circom`) need Lean models before the hidden-balance channel
  theorems use them?

## 5. Order

The review orders risk 1 first, because risks 2, 3 and 4 reuse its map. The design's step-1 lemmas T1 and T2 need the
representation map R. The vectors of Q3.1 (b) test R. The Poseidon abstraction of Q4.1 (a) uses the same terms.
So the order holds `[INFERENCE]`. The ruling and the decisions of §7 settle Q1.1-Q1.3 first.

## 6. Where this entry is integrated

- `docs/process.md` "The next chart": 9.1 (the named hypotheses), 9.2 (the current Solidity is not a source; the
  rewrite does not copy the two defects), 9.5 (a safety certificate carries only safety; the decided route), 9.6 (the
  decided work for the four gaps). The text of `24b9524` is removed.
- `engineering/build-plan.md` S5: item 2 of "Verification chain", the liveness bullet, "Refinement decisions", and the
  acceptance of P3 and S13.
- `engineering/11-risks-and-technical-debt.md`: R10.
- `notes/handoff.md`: item 11 is removed from "Waits for the operator", because no question is left for the operator.

armada-lean is not changed here. Two of its documents state what this entry corrects:
`docs/verification-chain.md:30-31` ("one named assumption") and `:60-66` ("seven open questions … wait for the
operator").

## 7. Operator ruling and decisions (2026-11-02)

### Operator ruling

"The sol isn't authoritative regarding anything at this point." The Lean model is the specification. The contracts are
rewritten from it. The current Solidity is not a fact to reconcile with the model. The operator's standing instruction
also applies: technical questions are decided by us, not asked. A question goes to the operator only if its decision
changes an approved theorem statement or adds an assumption beyond item 2 below. Defect 1 above ("risks 3 and 4 are
the operator's decision") is superseded by this instruction.

1. **Q1.1: closed, answer (a).** The 12-byte destination rule (armada-nitro `MultiAssetHolder.sol:474-476`) and the
   `ownerField` truncation (armada-poc `TransactModule.sol:637`) are defects of the current code. The rewrite does not
   copy them.
2. **Q1.2: closed.** The representation map of the rewrite chooses encodings that need no assumption beyond those that
   the EVM forces: an explicit destination kind, and no Keccak mod p encoding. The step-1 theorem names collision
   freedom of Keccak-256 and of its 160-bit truncation. Poseidon stays trusted (ruling of 2026-10-28). SHA-256 is added
   only if `L2BridgeApp` is modelled.

### Encodings that follow from item 2 (decided)

- **Destination.** A destination carries its kind as its own ABI field: an external 20-byte address, or a 32-byte
  channel id. The model's `Address.external` and `Address.channel` map to the two kinds one to one. Reason: the kinds
  then stay disjoint by construction, as in the model (`ArmadaLean/Nitro/Address.lean`).
- **Bound-parameters hash.** The rewrite passes the low 248 bits of `keccak256(abi.encode(core))` to the circuit as one
  public input. 2^248 is below the BN254 field prime (about 2^253.6), so no reduction happens. Reasons:
  - Two Keccak outputs that agree in their low 248 bits also agree in their low 160 bits. So collision freedom of the
    160-bit truncation (item 2) implies collision freedom of this encoding. No assumption is added.
  - The model keeps `hashBoundParams` as `Term.keccak` (`ArmadaLean/BoundParams.lean`) and one public input
    `boundParamsHash`. Approved statements name that one field (`ArmadaLean/Spec/PoolSafety.lean:192,302,500`,
    `Spec/PoolOwned.lean:89`) and `hashBoundParams_inj` (`Spec/BoundParams.lean:17`). So no statement changes.
  - Rejected: two 128-bit limbs (adds a public input, so the statements that name `boundParamsHash` change); a
    Poseidon hash of the fields (gas for many fields on chain, and a model definition change).
- **Token id.** The rewrite supports ERC-20 tokens only, with the zero-extended address as the token id
  (`TransactModule.sol:450-451`). The Keccak mod p token id of other token types (`:453`) goes. Reason: the model's
  `Token` is an address term only (`ArmadaLean/Token.lean`), and the build plan removes every entry point that the
  Lean model does not model (S5, "The new pool comes from the existing armada-poc pool").
- **Owner field.** The owned circuit and the contract take the 20-byte owner address as the field. A value at or above
  2^160 is rejected.

### Check of fact 6 before Q1.3

Question: does an approved theorem depend on "the attacker cannot conjure a hash term" for a term that it only
observed? Answer: no. Evidence:

- The model's operations already take a caller-chosen `Term` in each position where calldata carries a word: the
  shield `npk` (`ArmadaLean/Pool.lean:1070`), the transact `caller` (`:1101`), the output keys `npkOut` and the Merkle
  siblings of the witness (`ArmadaLean/Transact.lean`, `TransactWitness`), and the public inputs. So the model's
  attacker already names any hash term there, with no arguments. A map that decodes a copied word to a hash term gives
  the attacker nothing new in those positions.
- The positions where the circuit hashes private raw values (`pk`, `nk`, `random`, `value`, `token`) are `Nat` fields
  (`TransactWitness`). They decode to `val` and never to a hash term.
- `Term.val_ne_hash` (`ArmadaLean/Poseidon.lean:153`) has no use outside its own file. The `reduceCtorEq` steps in the
  pool modules separate `Option` constructors, not `Term` constructors (`Pool.lean:141-250`, `PoolOwned.lean:184,287`,
  `PoolSafety.lean:253-992`, `BoundParams.lean:81`).
- One approved spec theorem uses a constructor separation: `user_npk_ne_contract_npk`
  (`ArmadaLean/TransactOwned.lean:202-207`, statement `ArmadaLean/Spec/TransactOwned.lean`), through
  `Term.poseidon_ne_keccak`. Its concrete content: `Poseidon(mpk, r₁) = Poseidon(owner, r₂)` needs either a Poseidon
  collision, or a Poseidon output `mpk` equal to a 160-bit owner address. The first case is Poseidon's collision
  freedom. The second case is the separation of Poseidon's outputs from Keccak's, which the approved model encodes by
  the `HashFn` tag. Concretely it is a cross-function collision assumption: no Poseidon output that a run uses equals
  a Keccak output that the run uses. It is not part of "Poseidon stays trusted": that ruling (2026-10-28) says only
  that step 4 treats the deployed Poseidon bytecode as computing Poseidon correctly at the call boundary. It is not new
  either: the hypothesis of 2026-10-25 ("no Keccak or Poseidon collision among the values a run uses") covers it. So
  the step-1 theorem names it, and no operator ruling is needed (correction by the coordinator, 2026-11-02; my first
  reading of this point was wrong). The concrete work to break it: an output of Poseidon must equal one of k given
  addresses, about 2^254 / k evaluations if the outputs are uniform `[INFERENCE]`.
- Observed, with no theorem that depends on it: the plain nullifier is `Poseidon(nk, leafIndex)` with `nk` a raw value
  that the spender chooses (`ArmadaLean/Nullifier.lean:34`, `Transact.lean:100`). A spender can choose `nk` equal to an
  owner address. Then a plain nullifier and an owned nullifier can be equal at one leaf index. One leaf holds one
  commitment, and the spent set is keyed by tree, so this does not let one note be spent twice `[INFERENCE: argued, not
  proved]`. No approved theorem states that the two nullifier kinds differ.

### Decisions

- **Q1.3: decided, yes.** The step-1 refinement uses the run-relative map: each ABI position has a typed decoder; a word
  that the run computed as a hash decodes to that hash term; any other word decodes to `val` within its range. Reason:
  the fact-6 check above. With the encodings above, the step-1 theorem names three items: collision freedom of
  Keccak-256, collision freedom of its 160-bit truncation, and the separation of Poseidon outputs from Keccak outputs
  on the run's values. Separately, step 4 treats the deployed Poseidon as correct at the call boundary (2026-10-28).
- **Q2.1: decided, the design of `2026-10-28-liveness-refinement-design.md` §(d)** (D4 with a D5b certificate), with
  the O2 correction: the game transfer allows a bounded number of failed honest calls. Reasons, from that entry §(d):
  D1 is false as stated; D2 alone needs a transfer argument for each Liveness theorem; D3 proves each theorem twice and
  needs a second type approval; D5a lets the safety proof and the liveness proof pick different Stele runs. Limit: step
  L1 of the design adds size bounds as hypotheses to approved statements. That changes approved statements, so each
  bound goes to the operator with its gas measurement (O1). The bounds are the system's hard limits only (ruling of
  2026-10-28, "no limit on inputs or outputs below the system's hard limits").
- **Q2.2: decided.** Our agents make V1-V3 in the forks `AFDudley/EquiVM` and `AFDudley/EVMLean`, and offer each
  change upstream as a pull request (O6). Order: V1 (opaque precompiles) first, because without it the pool's safety
  certificate says nothing about a successful `transact` (design A12). V2 and V3 follow, before the first step-4
  certificate of a rewritten contract. None of them blocks the Lean-level Liveness proofs, because D4 proves those
  once, at Lean level.
- **Q3.1: decided, (b) now and (c) later.** Conformance vectors generated from the Lean model (ABI encodings, state
  hashes, valid and invalid transitions) run against the rewritten contracts with `eth_call` and against each client
  core. Reason: they test the representation map of Q1.3 against the real bytecode, at low cost. (c), an executable
  strategy proved to realize the strategy theorems, follows after the Liveness proofs close, because its proof uses
  them. It needs the feasibility predicates restated as searches over the held lists (§3, fact 3); that is a new
  definition with an equivalence proof, not a statement change.
- **Q4.1: decided.** Ruling 1 of 2026-10-28 covers the Groth16 verifier only, not the claim that a circuit implements
  its Lean model. The check is (a): import each circuit's R1CS (`compile.sh` already emits `.r1cs` and `.sym`) into
  Lean. The `.sym` template paths bind each Poseidon instance to the trusted Poseidon (ruling 2) and each EdDSA
  instance to the existing primitive `EdDSAVerify`. Every other constraint is proved equivalent to the model's
  `Constrained`. Reasons: (b) proves nothing; the hand model already disagrees with a circuit once (`notes/handoff.md`
  item 2, `Constrained.distinct`). Under the operator's ruling the Lean model is the specification, so a disagreement is
  a circuit defect, and process rule 8.3 applies to it. Order: after the Q1.3 map, because the importer decodes field
  elements through it. Until then `docs/how-to-verify.md` §5 (a person compares) is the only check.
- **Q4.2: decided, yes.** The acceptance of P3 and S13 checks that the key stored for each shape in the deployed pool
  equals the key that `snarkjs` exports from the ceremony zkey of that shape. Reason: P4 and S13 already tie the zkey to
  the R1CS (`snarkjs zkey verify`). This check closes the last link, from the zkey to the deployed key.
- **Q4.3: decided, yes.** `Ledger.circom` and `Battleship.circom` get Lean models, imported as in Q4.1, before any
  theorem about a hidden-balance channel relies on them. Reason: such a channel's app must prove
  `Rules.NoUnilateralSupport` before `finalizable_or_two_enabled` and `single_enabled_is_finalizable` apply to it
  (`2026-10-28-verification-chain-rulings.md`, "Later change"). The app's rules are the circuit's constraints.

No decision above changes an approved theorem statement. Q2.1's step L1 will, and goes to the operator with the
measured bounds. No question is left for the operator now.
