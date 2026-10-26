# Adjudicator model: audit for contract layout in the symbolic layer

Date: 2026-10-26. Read-only audit. No model file changed.

## 1. Scope and method

- Model: `AFDudley/armada-lean`, branch `adjudicator-model-v4`, commit `23e01b5`.
  The audit read a detached worktree at
  `armada-lean-adjudicator-v4/.scratch/wt/audit-adj`. Every model path below is
  relative to that worktree. Line numbers are for `23e01b5`.
- Files: `ArmadaLean/Adjudicator/{State,Channel,Outcome,Notes,Operations,Assumptions,Reachability,Safety}.lean`.
- Statements: the 40 theorems in `README.md`, section "Adjudicator" (lines 108-149).
  A script extracted the text of each statement, from `theorem` to `:= by`.
  A regular expression per finding then listed the statements whose text names the element.
- Sources:
  - S1 = `forcemove-2018.pdf`, S2 = `nitro-2019-219.pdf`, S3 = `satp-2022.pdf`.
    The audit converted each one with `pdftotext -layout`.
    A source line number below is a line of that text output.
  - S4 = the `docs.statechannels.org/protocol-tutorial_00x0-*.html` pages.
    The audit removed the HTML tags. "S4 0080:55" means page 0080, line 55 of the tag-free text.
  - All source files are in `armada-nitro/docs/references/nitro/sources/`.
    The text copies are in `armada-laconic/.scratch/audit-adj/` (gitignored).
  - Armada rules: `labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md`, section 3.
- The confirmed error is out of scope as a finding. That error is `AdjudicatorState.holdings`
  (`State.lean:62`), which the operator ruled on 2026-10-26.
  Section 4 lists the elements that exist only because of it.
- Classes: **paper**, **Armada rule**, **contract layout**, **contract-only rule**, **encoding**.
  An element whose only source is S4 is "contract vocabulary". Each finding states
  whether S1–S3 support it.

## 2. Findings

The findings are in the order of the table in section 3. That order is the
number of the 40 statements whose text names the element.

### A1. `configuredAssets`: the hub's asset allow-list

- **Declaration.** `AdjudicatorState.configuredAssets : Nat → Prop` (`State.lean:63`).
  `isAsset` (`State.lean:145`). The `assets` parameter of `genesis` (`State.lean:191`, `:195`).
  The `.deposit` check `s.configuredAssets asset` (`Operations.lean:197`).
- **What it represents.** The deploy-time asset list of the armada-nitro hub contract.
- **Class.** Contract-only rule (allow-list).
- **Evidence.**
  - Model: `State.lean:138-140`: "Whether `asset` is on the hub's configured list
    (`ArmadaNitroAdjudicator.isAsset`)". The only cited source is the contract.
  - Model: `README.md:146`: `configuredAssets_invariant`, source "armada-nitro code (R1 mutant)".
  - Source: S2 §3.1 (line 143-148) restricts the paper to one asset "for simplicity only".
    It states that "all the work here can be applied to state channel networks that
    manage an arbitrary number of asset types". S1, S3 and S4 have no asset list.
    The labbook (section 3, decisions 1–24) has no ruling on an asset list.
  - The audit searched S1–S4 and the labbook for "configured", "asset list" and "allow". The search found no source.
- **Statements whose text names it (23).** `configuredAssets_invariant` names the field.
  The other 22 name it through the parameter `genesis assets selfInputs pool0`:
  `finalized_outcome_was_concluded_or_challenged`, `turn_number_record_monotone`,
  `deposit_effect`, `value_leaves_only_via_transfer_of_finalized_channel`,
  `holdings_point_to_live_notes`, `value_exists_only_as_notes`, `unbeatable_pays`,
  `unbeatable_pays_after_timeout`, `notedValue_conserved_reachable`,
  `checkpoint_can_happen`, `deposit_can_happen`, `guarantee_reallocation_order_dependent`,
  `topup_can_happen`, `challenge_can_happen`, `challenge_timeout_can_happen`,
  `conclude_can_happen`, `payout_can_happen`, `reclaim_can_happen`,
  `bulk_conclude_and_transfer_can_happen`, `finalized_affordable_pays`,
  `holdings_never_zero_valued`, `heldValue_zero_iff_no_collateral`.
- **Replacement.** Delete the field, `isAsset`, the `genesis` parameter and the
  `.deposit` check. An asset is any token that the pool holds.
  Delete `configuredAssets_invariant`. If the operator wants an allow-list,
  the concrete layer (the Stele spec) states it.

### A2. `selfInputs`: the EVM `CREATE`/`CREATE2` address derivation

- **Declaration.** `AdjudicatorState.selfInputs : List Term` (`State.lean:64`).
  `AdjudicatorState.self := Term.keccak s.selfInputs` (`State.lean:92`).
  The `selfInputs` parameter of `genesis` (`State.lean:191`, `:196`).
- **What it represents.** The adjudicator's own address, as Keccak of its creation inputs.
- **Class.** Encoding (Ethereum address derivation).
- **Evidence.**
  - Model: `State.lean:80-84`: "`CREATE`'s address is `keccak256(rlp(creator, nonce))`'s
    low 160 bits ... (Ethereum Yellow Paper §7, 'Contract Creation')".
  - Source: S1–S4 do not derive the adjudicator's address. The Armada rule needs
    only an owner identity for the adjudicator's notes (labbook decision 9).
  - [INFERENCE] The shape exists to satisfy the pool lemma `user_npk_ne_contract_npk`,
    which takes the owner as `Term.keccak addrInputs` (`Lemmas.lean:3101`).
    The pool model owns that encoding. The pool audit covers it.
- **Statements whose text names it (23).** The same 23 statements as A1, through
  `genesis assets selfInputs pool0`.
- **Replacement.** An opaque `self : Term` (or a typed contract-owner identity) and
  one stated fact: no user note key equals a contract note key. [INFERENCE] This
  depends on the interface that the pool model exports.

### A3. Stored hashes (`stateHash`, `outcomeHash`) and the calldata that rehydrates them

- **Declaration.**
  - `ChannelStatus.stateHash : Term`, `ChannelStatus.outcomeHash : Term` (`Channel.lean:418-419`).
  - `outcomeHashOf` (`Assumptions.lean:123`), `Outcome.toTerm`, `SingleAssetExit.toTerm`,
    `Allocation.toTerm` (`Outcome.lean:82`, `:129`, `:157`).
  - `Call.transfer` arguments `claimedStateHash`, `outcome` (`Operations.lean:76`).
  - `Call.bulkTransferAllAssets` arguments `outcome`, `claimedStateHash` (`Operations.lean:81`).
  - `Call.reclaim` arguments `sourceOutcome`, `targetOutcome`, `claimedTargetStateHash` (`Operations.lean:78-80`).
  - The checks `(s.channels cid).stateHash = claimedStateHash` and
    `(s.channels cid).outcomeHash = outcomeHashOf outcome` (`Operations.lean:370-371`, `:438`, `:446-447`, `:497-498`).
  - `ChannelFinalizedWith` (`Safety.lean:85-87`).
- **What it represents.** The packed `statusOf` slot. The contract stores the hash of the
  outcome, not the outcome. So every payout call must resend the outcome and the
  state hash, and the contract checks them against the stored hash.
- **Class.** Encoding (gas optimisation: a hash stored in place of the value).
  The call arguments are contract layout (calldata).
- **Evidence.**
  - Model: `Channel.lean:404-414` states the fingerprint is "a gas-storage encoding with
    no semantic content". The model then keeps the two hashes that feed the fingerprint.
  - Model: `Operations.lean:370-371` checks the caller's hash against the stored hash.
  - Source S4 0070:101: "To save gas, information is only stored on chain in a hashed format."
  - Source S4 0080:55: "Because of the information is stored on chain, it is necesary
    to supply both the stateHash and encoded outcome of the channel when calling transfer."
  - Source S4 0040:192: "When the adjudicator needs to verify the exact state or outcome,
    the data is provided in the function arguments, as part of the calldata."
  - Source S3 §2.1 (lines 234-241): `AdjudicationState { version; outcome; finalizationTime }`.
    S3 stores the outcome itself and no state hash.
  - Source S1 §4.2 (lines 1292-1294): `struct Challenge { endState: State; endTime: Uint }`.
    S1 stores the whole state.
  - Source S2 Fig. 3 (lines 243-247) and Fig. 5: the adjudicator table has an "Outcome" column with the outcome.
  - Source S2 §5.1 (line 766) and S3 Listing 2 (line 344): the operations `T(A, B, x)` and
    `reclaim(ledgerChannelId, targetChannelId)` take no outcome and no hash.
- **Statements whose text names it (18).** `finalized_outcome_was_concluded_or_challenged`,
  `transfer_pays_only_the_finalized_outcome`, `transfer_order_independent_for_simple_allocations`,
  `finalization_permanent`, `transfer_never_pays_guarantee`, `reclaim_value_neutral`,
  `value_leaves_only_via_transfer_of_finalized_channel`, `unbeatable_pays_after_timeout`,
  `guarantee_reallocation_order_dependent`, `payout_can_happen`, `reclaim_can_happen`,
  `payout_takes_no_fee`, `bulk_transfer_all_assets_effect`,
  `bulk_conclude_and_transfer_all_assets_effect`, `finalized_affordable_pays`,
  `zero_payout_mints_nothing`, `full_payout_leaves_no_remainder`, `nullify_iff_commitment_was_some`.
- **Replacement.** The channel status stores the outcome (S3) and the turn number.
  Delete the `stateHash`/`outcomeHash` fields and `ChannelFinalizedWith`'s hash arguments.
  `transfer` reads the stored outcome. `reclaim` takes the two channel ids (S3 Listing 2).
  Delete `claimedStateHash`, `claimedTargetStateHash`, and the outcome arguments.
  `finalized_outcome_was_concluded_or_challenged` and `finalization_permanent` then
  state facts about the stored outcome with `OutcomeDerivedFrom`.
  The function `stateHash fp vp` stays, because it is the digest that participants sign (S1 §3.1.3).

### A4. The outcome as an array, addressed by position (`assetIndex`), and `hassetunique`

- **Declaration.**
  - `abbrev Outcome := List SingleAssetExit` (`Outcome.lean:152`).
  - The `assetIndex` argument of `Call.transfer` (`Operations.lean:76`).
  - `sourceAssetIndex`, `indexOfTargetInSource`, `targetAssetIndex` of `Call.reclaim` (`Operations.lean:78-80`).
  - The hypothesis `hassetunique` (`Safety.lean:3151`, `:3784`, `:3938`).
- **What it represents.** The ABI array `SingleAssetExit[]`. A caller names an exit by its
  array position. An array can hold two exits for one asset. So three statements need
  a hypothesis that the asset occurs once.
- **Class.** Contract layout (ABI array encoding).
- **Evidence.**
  - Model: `Safety.lean:3151-3152`: `hassetunique : ∀ i (hi : i < candidate.outcome.length), i ≠ assetIndex → (candidate.outcome.get ⟨i, hi⟩).asset ≠ ...`.
  - Model: the proofs use `hassetunique` only to feed `bulkTransferFold_index_payout` (`Safety.lean:1904`, `:2201`, `:3690`).
    [INFERENCE] The bulk call (A9) pays every array entry, so a duplicate asset entry drains the one holding first.
  - Source S4 0080:55: "an asset index is required for slicing the outcome into a single asset outcome".
  - Source S3 §2.4 (line 334): "An outcome is an ordered dictionary, mapping a destination to either
    an allocation or a guarantee". S3 Listing 2 addresses entries by key (`o_ledger[left]`,
    `delete o_ledger[targetChannelId]`). S2 §3.2 (lines 236-240) and §5.1 (lines 742-745) define an
    allocation as a list of address/total pairs in which each address occurs once.
    No source addresses an asset by position.
- **Statements whose text names it (17).** `transfer_pays_only_the_finalized_outcome`,
  `transfer_order_independent_for_simple_allocations`, `transfer_never_pays_guarantee`,
  `reclaim_value_neutral`, `value_leaves_only_via_transfer_of_finalized_channel`,
  `unbeatable_pays`, `unbeatable_pays_after_timeout`, `guarantee_reallocation_order_dependent`,
  `payout_can_happen`, `reclaim_can_happen`, `payout_takes_no_fee`,
  `bulk_conclude_and_transfer_can_happen`, `finalized_affordable_pays`,
  `zero_payout_mints_nothing`, `full_payout_leaves_no_remainder`,
  `nullify_iff_commitment_was_some`, `cumulative_payout_never_exceeds_allocation` (`o.getD i`).
  `hassetunique` is in 3 of them: `unbeatable_pays`, `unbeatable_pays_after_timeout`, `finalized_affordable_pays`.
- **Replacement.** `Outcome := asset → allocation list` (a finite map). `transfer` and
  `reclaim` take an asset, not an index. Then a duplicate asset cannot occur, and
  `hassetunique` disappears.

### A5. `indices`: the selection list of `transfer`, the `[] = all` sentinel, and `StrictlyIncreasing`

- **Declaration.** The `indices` argument of `Call.transfer` (`Operations.lean:76-77`).
  `transferEffects`'s `selected i := indices.isEmpty || i ∈ indices` (`Outcome.lean:225`).
  `StrictlyIncreasing` (`Outcome.lean:210-213`) and its check (`Operations.lean:372`).
  `TransferOnlyDerivedFrom`'s `indices = [] ∨ k ∈ indices` (`Safety.lean:1290`).
- **What it represents.** `MultiAssetHolder.transfer`'s calldata. A caller picks a set of
  allocation positions. The empty list means "every position".
- **Class.** Encoding (the empty list is a sentinel for "all"). `StrictlyIncreasing` is a gas optimisation.
- **Evidence.**
  - Model: `Outcome.lean:207-209` quotes the contract: "Checks that the supplied indices are
    strictly increasing. This allows us... to write a more efficient claim function."
  - Model: `Outcome.lean:203-204`: "`indices == []` means 'pay out to all' (`:149`, the exact words of the code comment)".
  - Source S4 0080:84: "This magic value (a zero length array) implies we want to pay out all of the allocationItems".
  - Source S2 §5.1 (lines 765-772): `T(A, B, x)` names one destination `B` and one amount `x`.
    "If A cannot afford x for B, then T(A, B, x) fails". S1 §6.3 `withdraw(channelId, k)` names one position.
    S3 has no transfer pseudocode.
  - The priority computation inside `transferEffects` (`afford`) is S2 §5.1/§5.2. Only the selection layer is contract layout.
- **Statements whose text names it (11).** `transfer_pays_only_the_finalized_outcome`,
  `transfer_order_independent_for_simple_allocations`, `transfer_never_pays_guarantee`,
  `value_leaves_only_via_transfer_of_finalized_channel`, `cumulative_payout_never_exceeds_allocation`
  (through `TransferOnlyDerivedFrom`), `guarantee_reallocation_order_dependent`,
  `payout_can_happen`, `payout_takes_no_fee`, `zero_payout_mints_nothing`,
  `full_payout_leaves_no_remainder`, `nullify_iff_commitment_was_some`.
- **Replacement.** S2's `T(A, B, x)`: `transfer cid asset destination`, which pays the
  affordable amount for that destination. A batch of destinations is a sequence of calls.
  Delete `StrictlyIncreasing` and the empty-list sentinel. If the operator keeps a
  selection, use a finite set of destinations with no sentinel.

### A6. `finalizesAt = 0` means "no deadline", and the `hnowpos` hypotheses it forces

- **Declaration.** `ChannelStatus.finalizesAt : Nat` (`Channel.lean:417`).
  `ChannelStatus.mode`'s `if c.finalizesAt = 0 then .open` (`Channel.lean:436`).
  `.checkpoint` writes `finalizesAt := 0` (`Operations.lean:294`).
  `.conclude` writes `finalizesAt := s.now` (`Operations.lean:328`).
  `genesis` sets `now := 0` (`State.lean:197`).
  The hypothesis `hnowpos : 0 < s.now` (`Safety.lean:637`, `:3148`, `:3782`).
- **What it represents.** The 48-bit `finalizesAt` field of the packed slot, with `0` as the
  "Open" sentinel. A `conclude` at time `0` writes `0`, and the model reads that as Open.
  So three statements must assume `0 < s.now`.
- **Class.** Encoding (sentinel).
- **Evidence.**
  - Model: `Safety.lean:605-609`: "Needs `0 < s.now`: `conclude` writes `finalizesAt := s.now` ...
    at which `finalizesAt = 0` reads as `Open`, not `Finalized`".
  - Model: `Safety.lean:3077-3080`: "treats `0` as the SAME sentinel `ChannelStatus.mode` reads as 'no deadline set'".
  - Source S4 0040:165: "When it is (the magic and initial value of) 0, the channel is Open."
  - Source S3 §2.3 (line 297): `return status.finalizesAt == 0 || status.finalizesAt > now`.
  - Source S1 §4.2 (line 1297): `Challenge::terminated: c ↦ c.endTime > 0 && !c.inProgress`. S1 `conclude`
    calls `setChallenge(endState, now())` (line 1500). S1 `cancelChallenge` sets `c.endTime := 0` (line 1768).
  - So S1 and S3 pseudocode use the same sentinel. The sentinel has the same defect there.
    The paper-level meaning is the mode diagram: S1 Fig. 5 (Collaborative, Challenge, Concluded,
    Terminated) and S2 Fig. 4 (Challenge, Finalized). No source needs the clock to be positive.
- **Statements whose text names it (7).** `turn_number_record_monotone` (`hnowpos`),
  `unbeatable_pays` (`hnowpos`), `unbeatable_pays_after_timeout` (`hnowpos`),
  `finalization_permanent` (`finalizesAt`), `checkpoint_can_happen` (`finalizesAt = 0`),
  `challenge_timeout_can_happen` (`finalizesAt > 0`), `bulk_conclude_and_transfer_all_assets_effect` (`finalizesAt := s.now`).
- **Replacement.** A sum type for the channel phase: `open | challenge (deadline : Nat) | finalized`
  (or `finalizesAt : Option Nat`). `conclude` sets `finalized` directly. `mode` reads
  `challenge d` as finalized when `d ≤ now`. Delete the three `hnowpos` hypotheses.
  `checkpoint_can_happen` and `challenge_timeout_can_happen` then state the phase, not the number.

### A7. Statements whose only source is the armada-nitro code or its R1 mutation tests

- **Declaration.** Six statements: `heldValue_zero_iff_no_collateral` (`Safety.lean:6010`),
  `configuredAssets_invariant` (`:6047`), `zero_payout_mints_nothing` (`:6077`),
  `full_payout_leaves_no_remainder` (`:6157`), `nullify_iff_commitment_was_some` (`:6219`),
  `topup_can_happen` (`:4687`).
- **What it represents.** Facts that the R1 mutants of `ArmadaNitroAdjudicator.sol` test.
- **Class.** Contract-only rule.
- **Evidence.**
  - Model: `README.md:132`, `:145-149` cite "armada-nitro code (R1 mutant)" or "armada-nitro code (`ArmadaNitroAdjudicator.sol:60-72`)".
  - Each statement's text is about the confirmed `holdings` encoding or pool storage:
    - `heldValue_zero_iff_no_collateral`: `s.holdings asset cid = none ∨ ∃ pos c, ... ∧ s.pool.preimageOf c = none`.
    - `full_payout_leaves_no_remainder`: `s'.holdings ... = none`.
    - `nullify_iff_commitment_was_some`: `s.holdings ... = none ↔ s'.pool.spent = s.pool.spent`.
    - `zero_payout_mints_nothing`: `(∃ pos, (pos, pre) ∈ s'.pool.live pre.token)`.
    - `topup_can_happen`: a second deposit that runs `DepositMergeStep`.
    - `configuredAssets_invariant`: finding A1.
  - Source: S2 §3.1 (lines 161-163) defines only a balance. S2 §5.1 (lines 766-772) defines the
    effect of a transfer on balances. No source defines a "nullify", a "remainder note" or a zero-value note.
    S2 §5.4.3 ("Topping Up a Ledger Channel") defines a top-up as a plain deposit.
- **Statements whose text names it (6).** The six statements themselves.
- **Replacement.**
  - `heldValue_zero_iff_no_collateral`: delete. With a balance it is a definition.
  - `configuredAssets_invariant`: delete (A1).
  - `full_payout_leaves_no_remainder`: restate as `heldValue s' asset cid = 0`.
  - `nullify_iff_commitment_was_some`: delete. It states pool internals.
  - `zero_payout_mints_nothing`: restate as "a transfer that pays `0` changes no balance and no note owner's value".
    [INFERENCE] The operator must rule if a zero-value note is an Armada-level concern.
  - `topup_can_happen`: restate as `deposit_effect` at a funded channel (S2 §3.1, §5.4.3).

### A8. `expectedHeld` on `deposit`

- **Declaration.** The `expectedHeld` argument of `Call.deposit` (`Operations.lean:71`).
  The check `heldValue s asset channelId = expectedHeld` (`Operations.lean:198`).
- **What it represents.** The re-org and front-run guard of `MultiAssetHolder.deposit`.
- **Class.** Contract vocabulary (S4 only). S1 handles the same hazard off-chain. S2 and S3 do not have it.
- **Evidence.**
  - Source S4 0060:53-56: `deposit(address asset, bytes32 destination, uint256 expectedHeld, uint256 amount)`.
    "the on-chain holdings for destination must be equal to expectedHeld ... prevents loss of funds".
  - Source S2 §3.1 (lines 165-168): `D_A(x)` has no expected value. "There are no restrictions on who can deposit coins".
  - Source S1 §6.2 (lines 1797-1812): Bob "waits until he can verify ... that the adjudicator contains aResolution".
    S1 puts the check in the participants' protocol, not in the adjudicator.
  - Model: `Operations.lean:161-163` ties `expectedHeld` to `depositBinding` (`ArmadaNitroAdjudicator.sol:60-72`), a contract source.
- **Statements whose text names it (5).** `deposit_wrong_expected_held_fails`, `deposit_external_destination_fails`,
  `deposit_effect`, `deposit_can_happen`, `topup_can_happen` (all through `.deposit cid expectedHeld asset amount`).
- **Replacement.** `deposit channelId asset amount` (S2 `D_A(x)`). Delete
  `deposit_wrong_expected_held_fails` (A-07), or keep it by an explicit operator ruling.
  The operator approved A-07 on 2026-10-25. This finding asks for a new ruling.

### A9. The batch calls `bulkTransferAllAssets` and `bulkConcludeAndTransferAllAssets`

- **Declaration.** `Call.bulkTransferAllAssets`, `Call.bulkConcludeAndTransferAllAssets` (`Operations.lean:81-82`).
  `bulkTransferFold` (`Operations.lean:96-125`). Their `Step` cases (`Operations.lean:495-515`).
- **What it represents.** Contract utility functions that combine `conclude` and `transfer` in one transaction.
- **Class.** Gas optimisation (batching). Call set modeled because the contract has the function.
- **Evidence.**
  - Source S2 §3.3 (lines 313-317): "we present all operations separately and assume they are called separately.
    In practice, we expect that implementations would provide some utility methods that combine
    common sequences of operations, to improve gas efficiency."
  - Source S4 0080:110: "a convenience method concludeAndTransferAllAssets ... batching them to save gas."
  - Source S3 footnote 10 (line 380): "A gas-optimized Reclaim(X) might reclaim all reclaimable guarantees".
  - No source names `transferAllAssets`. Model `Operations.lean:474-475` names it "the OTHER half the name implies".
  - The bulk path forces `hassetunique` (A4) and adds the third disjunct of `value_leaves_only_via_transfer_of_finalized_channel`.
- **Statements whose text names it (4).** `bulk_transfer_all_assets_effect`, `bulk_conclude_and_transfer_all_assets_effect`,
  `bulk_conclude_and_transfer_can_happen`, `value_leaves_only_via_transfer_of_finalized_channel`.
- **Replacement.** Delete both calls and `bulkTransferFold`. A batch is a sequence of `.conclude` and
  `.transfer` steps, and `Reachable` already covers every sequence. Delete the three B4 statements
  and the bulk disjuncts of A-13. The operator ruled for B4 on 2026-10-25 (`README.md:102-103`).
  This finding asks for a new ruling, with S2 §3.3 as the evidence.

### A10. The adjudicator restates the pool's storage update field by field

- **Declaration.** `DepositStep` (`Notes.lean:169-195`) and `PoolOwnedSpendStep` (`Notes.lean:278-339`).
  They write `live`, `spent`, `closedTrees`, `leaves`, `roots`, `preimageOf`, `ownedRandomOf`.
  They compute positions with `insertTreeNumber`/`insertStartPos` (`Notes.lean:175-176`, `:283-284`).
  They carry the circuit-shape widths `N ≤ 8`, `M ≤ 4` (`Notes.lean:177`, `:291`, `:315`).
  `.transfer` computes a leaf position as `insertStartPos ... + outputs.length - 1` (`Operations.lean:386`).
- **What it represents.** The pool contract's Merkle tree layout, tree rollover, and registered circuit shapes.
- **Class.** Contract layout (pool storage) and concrete widths, inside the adjudicator layer.
- **Evidence.**
  - Model: `Notes.lean:26-30`: "it builds the exact pool state `Step.transact`/`Step.transactOwned` would produce for the SAME witness".
    The pool already defines `Step.transact` and `Step.transactOwned` (`Pool.lean:1414`, `:1492`).
  - Armada rule: `README.md:93`: "it is an ordinary pool user". Labbook decision 9.
    An ordinary user calls the pool's operation. It does not restate the pool's storage.
  - Model: `Notes.lean:81-87` gives the reason: `Pool.Step` alone does not say which constructor fired.
    [INFERENCE] A per-constructor relation exported by the pool removes that reason.
- **Statements whose text names it (4).** `zero_payout_mints_nothing` (`s'.pool.live`, positions),
  `nullify_iff_commitment_was_some` (`s'.pool.spent`), `bulk_transfer_all_assets_effect` and
  `bulk_conclude_and_transfer_all_assets_effect` (`bulkTransferFold`).
- **Replacement.** The pool exports one relation per operation (for example
  `TransactOwnedStep pool owner inputs outputs pool'`). The adjudicator calls it with its owner identity.
  No tree number, position, or circuit width appears in the adjudicator.
  [INFERENCE] Coordinate this with the pool audit.

### A11. Pool witness internals in the honest-recipient hypothesis `hhonestPool`

- **Declaration.** `hhonestPool` (`Safety.lean` statements at `:3136`, `:3768`, `:3926`).
  Its text names `TransactWitness N M`, `TransactPublic N M`, `msgHashOf`, `Constrained`,
  `masterPublicKey w.publicKey w.nullifyingKey`, `w.randomIn i`.
- **What it represents.** "The recipient's key signed no pool spend of a note with key `dest.destination`",
  written in the pool circuit's witness fields.
- **Class.** Encoding (circuit witness layout) in statement text. [INFERENCE] The hypothesis follows
  from the conclusion shape "a live note exists in the final state" together with the `.deposit` gap that
  `Operations.lean:174-194` records.
- **Evidence.**
  - Model: `Safety.lean:3128-3135` explains that `hhonestPool` closes a gap that `deposit_spend_requires_signer`
    and `poolSignedBy_frame` leave open.
  - Source: S2 §4.1 (line 483) defines the goal as "an unbeatable strategy for obtaining a state where the
    balance of A's ..." increases. No source names a circuit witness.
- **Statements whose text names it (3).** `unbeatable_pays`, `unbeatable_pays_after_timeout`, `finalized_affordable_pays`.
- **Replacement.** State the conclusion as an event: the strategy's payout step credits `dest` with
  at least `x` (a note of value at least `x` with key `dest.destination` enters the pool in that step).
  The survival of the note is then a pool fact (the owner alone can spend it) and needs no adjudicator hypothesis.
  [INFERENCE] The operator must rule whether "paid" means "credited" or "still unspent at the end".

### A12. `conclude` resets `turnNumRecord` to `0`

- **Declaration.** `.conclude` and `.bulkConcludeAndTransferAllAssets` write `turnNumRecord := 0` (`Operations.lean:328`, `:512`).
- **What it represents.** `ForceMove.conclude`'s `ChannelData(0, uint48(block.timestamp), bytes32(0), ...)`.
- **Class.** Contract vocabulary (S4 only). S1 keeps the turn number.
- **Evidence.**
  - Source S4 0040:168: "The exception to this rule is that it is set to 0 when the channel is concluded via a conclude call."
  - Source S1 §4.4.4 (line 1500): `conclude` calls `setChallenge(endState, now())`. The stored `endState` keeps its `turnNum`.
  - S3 does not give `conclude` pseudocode.
  - Model: the reset forces the third disjunct of A-03, `turnNumRecord = 0 ∧ mode = .finalized` (`Safety.lean:82`).
- **Statements whose text names it (2).** `turn_number_record_monotone`, `bulk_conclude_and_transfer_all_assets_effect`.
- **Replacement.** A finalized channel has no turn record (A6's phase type). State A-03 over non-finalized
  phases only. `finalization_permanent` already covers the finalized phase. Delete the third disjunct.

### A13. `Term.val 0` as "no hash" in the channel status

- **Declaration.** `emptyStatus := ⟨0, 0, Term.val 0, Term.val 0⟩` (`Channel.lean:430`).
  `.checkpoint` writes `stateHash := Term.val 0, outcomeHash := Term.val 0` (`Operations.lean:294-295`).
  `.conclude` writes `stateHash := Term.val 0` (`Operations.lean:328`, `:512`).
- **What it represents.** `bytes32(0)` in an unwritten or cleared storage slot.
- **Class.** Encoding (sentinel).
- **Evidence.**
  - Model: `Channel.lean:421-429` quotes "`_unpackStatus(someRandomChannelId)` returns `(0,0,0)`".
  - Source S4 0080:83: after a happy-path conclude, the caller passes `constants.HashZero` as `stateHash`:
    "If the channel was concluded on the happy path, we can use this default value".
  - Source S2 §3.1 (lines 162-163): "If an address does not appear in the table we take the balance to be zero."
    The paper uses absence, not a zero value.
- **Statements whose text names it (1, and 1 indirect).** `bulk_conclude_and_transfer_all_assets_effect`.
  Indirect: `finalized_outcome_was_concluded_or_challenged` states nothing about `stateH` in its conclude
  disjunct, because conclude clears the hash (`Safety.lean:128-131`).
- **Replacement.** A3 and A6 remove it. An Open channel carries no outcome. A finalized channel carries its outcome.

### A14. `isExternal`: the 20-byte address test on an untyped 32-byte destination

- **Declaration.** `isExternal destination := ∃ a : Nat, a < 2 ^ 160 ∧ destination = Term.val a` (`Outcome.lean:174`).
  `Allocation.destination : Term` (`Outcome.lean:67`). The `.deposit` check `¬ isExternal channelId` (`Operations.lean:199`).
- **What it represents.** The `bytes32` destination that holds either a channel id or a left-padded address.
- **Class.** Encoding (padding, `2^160` width), and arithmetic on a symbolic object.
- **Evidence.**
  - Source S4 0030:77: "a 32 byte identifier which may either denote a channel ID or a so-called 'external destination'
    (a 20 byte Ethereum address left-padded with zeros)."
  - Source S4 0060:55-56: "destination must NOT be an external destination".
  - Source S2 §3.1 (lines 227-228): "Anyone can deposit into an address of either type".
    Footnote 3 (line 269): "But there is nothing to be gained from depositing into a participant address."
    S2 allows the deposit that A-06 forbids.
  - [INFERENCE] The untyped destination also hides a gap. `.transfer` mints a note with `npk = a.destination`
    for every payout (`Operations.lean:377-379`). A payout to a channel id (S2 §5.3, §6.4: `T(L, G, x)`)
    then becomes a note with a channel id as its key, not a credit to that channel's balance.
- **Statements whose text names it (1).** `deposit_external_destination_fails`.
- **Replacement.** `Destination := channel ChannelId | user NoteKey`. `deposit` takes a `ChannelId`.
  Delete `isExternal` and A-06. The type then shows the channel-to-channel case that `.transfer` must handle.

### A15. `Allocation.metadata` and the `decodeGuarantee` axiom

- **Declaration.** `Allocation.metadata : Nat` (`Outcome.lean:70`). `axiom decodeGuarantee : Nat → Term × Term` (`Outcome.lean:258`).
  Its uses in `.reclaim` (`Operations.lean:443`) and `OutcomeDerivedFrom.reclaimStep` (`Outcome.lean:494-502`).
- **What it represents.** The ABI-encoded bytes `struct Guarantee { bytes32 left; bytes32 right; }`.
- **Class.** Encoding (metadata bytes and a decoder).
- **Evidence.**
  - Source S4 0030:70-73: "The metadata is an encoding of the following struct: `struct Guarantee { bytes32 left; bytes32 right; }`".
  - Source S3 §2.4 (line 322): "A guarantee is a data structure G(x, [A, B]) representing an amount x and an ordered list of destinations A and B."
    S2 §6.1 (lines 1094-1101): a guarantee "provides a list of addresses". No source has a decoder.
  - The axiom adds one entry to the trust base for a pure encoding.
- **Statements whose text names it (1, and 2 through `.reclaim`).** `reclaim_value_neutral`.
  Through the `.reclaim` step: `reclaim_can_happen`, `guarantee_reallocation_order_dependent`.
- **Replacement.** `Allocation := simple (dest) (amount) | guarantee (target) (amount) (left right : Destination)`.
  Delete `metadata`, `decodeGuarantee`, and `AllocationType`'s use as a tag field.

### A16. Contract revert conditions in `reclaim`

- **Declaration.** In `.reclaim` (`Operations.lean:450-458`): `hfoundLeft`/`hfoundRight` (an allocation for each
  destination exists), `allocations.length ≥ 2`, and `a0 + a1 = guarantee.amount`. The same in
  `OutcomeDerivedFrom.reclaimStep` (`Outcome.lean:491-498`).
- **What it represents.** `require(foundLeft)`, `require(foundRight)` and the sum check of `compute_reclaim_effects` (`MultiAssetHolder.sol:384-387`).
- **Class.** Contract-only rule. [INFERENCE] The sum check carries protocol meaning. The "found" checks do not.
- **Evidence.**
  - Model: `Operations.lean:427-433`: S3 Listing 2 "ASSUMES, with no on-chain check, that the target's first two
    allocation amounts sum to the guarantee's own amount; go-nitro enforces both".
  - Source S3 Listing 2 (lines 344-364) has no such `require`. S3 Lemma 3.1 (lines 511-513) derives
    `a0 + b0 = a + b` from the virtual-channel app. S3 models the outcome as a dictionary, so `o_ledger[left]` always has an entry.
  - [INFERENCE] Without the sum check, `reclaim_value_neutral` is false for an adversarial target outcome.
    So the check is a real safety precondition that S3 moves into the app.
- **Statements whose text names it (1, and 2 through `.reclaim`).** `reclaim_value_neutral` (`a0 + a1 = guarantee.amount`).
  Through the `.reclaim` step: `reclaim_can_happen`, `guarantee_reallocation_order_dependent`.
- **Replacement.** Keep the sum condition. Cite it as the S3 Lemma 3.1 precondition, not as a contract revert.
  With a destination-keyed outcome (A4), crediting `left` and `right` needs no "found" check.

### A17. Numeric tags inside the hash terms, and the unused `shield` tag

- **Declaration.** `AllocationType.shield` (`Outcome.lean:41`). `AllocationType.toNat` (`Outcome.lean:48-51`).
  `Term.val (if isFinal then 1 else 0)` in `stateHashOf` (`Assumptions.lean:99`).
- **What it represents.** `SHIELD_ALLOCATION_TYPE` of `ArmadaNitroAdjudicator.sol:18`, and the ABI number encodings of an enum and a `bool`.
- **Class.** Encoding. `shield` is contract vocabulary with no source.
- **Evidence.**
  - Model: `Outcome.lean:32-36`: "No paper names this tag; it exists only because ...". A search of `ArmadaLean/`
    for `.shield` as an `AllocationType` found no use outside `toNat` (`Outcome.lean:51`).
  - `toNat` and the `bool` encoding exist only to build `Allocation.toTerm` and `stateHashOf`.
- **Statements whose text names it (0).**
- **Replacement.** Delete `shield`. After A3, only the signing digest `stateHashOf` stays. A symbolic term can carry the
  `Bool` and the allocation constructor directly, with no numeric tag.

## 3. Table of findings

Ordered by the number of the 40 statements whose text names the element.
"+n" counts statements that reach the element only through a `Step` case or a definition.

| # | Element | Class | Statements | Count |
|---|---|---|---|---|
| A1 | `configuredAssets`, `isAsset`, `genesis assets` | contract-only rule | 22 by parameter, 1 direct | 23 |
| A2 | `selfInputs`, `self := Term.keccak selfInputs` | encoding | 23 by parameter | 23 |
| A3 | `ChannelStatus.stateHash`/`outcomeHash`, `claimedStateHash`, outcome call arguments | encoding (gas) + calldata layout | 18 | 18 |
| A4 | `Outcome` as array, `assetIndex`, `hassetunique` | contract layout | 17 (3 with `hassetunique`) | 17 |
| A5 | `indices`, `[]` = all, `StrictlyIncreasing` | encoding + gas | 11 | 11 |
| A6 | `finalizesAt = 0` sentinel, `hnowpos` | encoding | 7 (3 with `hnowpos`) | 7 |
| A7 | statements sourced only by armada-nitro code / R1 | contract-only rule | the 6 statements | 6 |
| A8 | `expectedHeld` on `deposit` | contract vocabulary (S4 only) | 5 | 5 |
| A9 | `bulkTransferAllAssets`, `bulkConcludeAndTransferAllAssets` | gas (batching), call set | 4 | 4 |
| A10 | pool storage restated in `DepositStep`/`PoolOwnedSpendStep`, `N ≤ 8`, `M ≤ 4` | contract layout (pool) | 4 | 4 |
| A11 | pool witness fields in `hhonestPool` | encoding | 3 | 3 |
| A12 | `conclude` sets `turnNumRecord := 0` | contract vocabulary (S4 only) | 2 | 2 |
| A13 | `Term.val 0` as "no hash" | encoding | 1 (+1 indirect) | 1 |
| A14 | `isExternal`, `2 ^ 160`, untyped destination | encoding + width | 1 | 1 |
| A15 | `Allocation.metadata`, `decodeGuarantee` axiom | encoding | 1 (+2) | 1 |
| A16 | `reclaim` revert conditions | contract-only rule | 1 (+2) | 1 |
| A17 | `AllocationType.shield`, `toNat`, `bool` as 0/1 | encoding | 0 | 0 |

Dependencies: A13 disappears with A3 and A6. A12 disappears with A6. `hassetunique` (A4) also disappears with A9.
A7 shrinks to two statements after the `holdings` fix and A1.

Decisions for the operator: A8 and A9 remove statements that the operator approved on 2026-10-25 (A-07, B4).
A11 and A7 (`zero_payout_mints_nothing`) need a ruling on what "paid" means under the Armada rule.

## 4. Elements that depend on the confirmed `holdings` error (not new findings)

These elements exist because `holdings` stores a commitment reference. The 2026-10-26 ruling covers them.

- `heldValue`'s `match ... | none => 0 | some (_, c) => ... getD 0` (`State.lean:133-136`).
- `bulkTransferFold`'s `holdings` type and its `held` match (`Operations.lean:97-103`).
- `PoolOwnedSpendStep`'s `nullified : Option ((Nat × Nat) × Term)` (`Notes.lean:279`).
- `DepositMergeStep` (`Notes.lean:368-400`) and the `some` branch of `.deposit` (`Operations.lean:211-218`).
- `.deposit`'s `0 < amount` (`Operations.lean:200`). It protects only the zero-note rule of `holdings_never_zero_valued`.
- The remainder position `insertStartPos ... + outputs.length - 1` in `.transfer` and `bulkTransferFold` (`Operations.lean:119`, `:386`).
- `HoldingWellFormed` (`Lemmas.lean:4009`).
- Statement text: `holdings_point_to_live_notes` (`pos : Nat × Nat`), `value_exists_only_as_notes`,
  `holdings_never_zero_valued`, `heldValue_zero_iff_no_collateral`, `full_payout_leaves_no_remainder`,
  `nullify_iff_commitment_was_some`, `topup_can_happen`.

## 5. Coverage: elements checked and found legitimate

### `AdjudicatorState` fields

- `pool` — Armada rule: the adjudicator's money is pool notes (labbook decision 9).
- `channels : Term → ChannelStatus` — paper: S3 §2.1 "mapping statusOf, using channel ids as keys". Its contents: A3, A6, A12, A13.
- `holdings` — the confirmed error. Section 4.
- `configuredAssets` — finding A1.
- `selfInputs` — finding A2.
- `now` — paper: S2 §4.3 "the current time is t"; S1 §4.1 timeout.
- `signedBy` — model device for S2 §3.1 / S3 §1.4 unforgeability. Legitimate.
- `poolSignedBy` — model device, same role at the pool layer. Its key type `Nat × Nat` (a BabyJubJub point) comes from the pool model. The pool audit covers it.

### `ChannelStatus`, `ChannelMode`

- `turnNumRecord` — paper: S3 `AdjudicationState.version`; S1 §4.4 stored `turnNum`.
- `finalizesAt` — the deadline is paper (S3 `finalizationTime`, S1 `endTime`). The `0` sentinel is finding A6.
- `stateHash`, `outcomeHash` — finding A3.
- `ChannelMode` (`open`, `challenge`, `finalized`) — paper: S1 Fig. 5, S2 Fig. 4.
- `ChannelStatus.mode` — paper concept. The `= 0` branch is A6.
- `emptyStatus` — the zero turn and Open phase are legitimate. The hash sentinels are A13.

### `FixedPart`, `VariablePart`, identity

- `participants`, `channelNonce`, `appDefinition`, `challengeDuration` — paper: S3 §2.1 `ChannelConsts` (`peers`, `nonce`, `appDef`, `challengeDuration`).
- `outcome`, `appData`, `turnNum`, `isFinal` — paper: S3 §2.1 `ChannelVars` (`version` = `turnNum`).
- `channelId`/`channelIdOf` — paper: S2 §3.1 and S3 §2.1 hash of constants. S3 omits `challengeDuration` from the hash; S4 includes it. The difference is minor and is not a layout error.
- `stateHash fp vp` (the function) — legitimate as the signed digest (S1 §3.1.3). Only the stored field is A3.

### `Outcome`, `SingleAssetExit`, `Allocation`

- `Allocation.destination` — paper (S2 §5.1 address, S3 destination). Its untyped encoding is A14.
- `Allocation.amount` — paper (S2 §5.1 total, S3 `A(a)`).
- `Allocation.allocationType` with `simple`/`guarantee` — paper distinction (S2 §6.1, S3 §2.4). The `shield` tag is A17. The tag-field form is A15.
- `Allocation.metadata` — finding A15.
- `Allocation.default`, `⟨0, []⟩` defaults in `getD` — Lean totality devices, no contract content.
- `SingleAssetExit.asset`, `.allocations` — S4 shape; S2 §3.1 supports many assets. The array form is A4.
- `afford` — paper: S2 §5.1 "A can afford x for B", §5.2 step 3(b)(i).
- `transferEffects` priority computation — paper (S2 §5.2). Its selection layer is A5.
- `creditFirst` — paper: S2 §5.1 "ignoring any additional entries for a given address after the first".
- `reallocateGuarantee` — paper: S3 Listing 2.
- `OutcomeDerivedFrom` — model device over legitimate rewrites. It inherits `decodeGuarantee` (A15) and the reclaim conditions (A16).

### Assumptions, signatures, reachability

- `HashAssumptions` — empty and vacuous. Not a layout error.
- `SignatureAssumptions.contractAuthorizes`, `AuthorizedBy` — the EIP-1271 branch is named after a contract standard. [INFERENCE] It is legitimate as "a contract participant authorizes", because Armada hubs are contracts (labbook decision 19).
- `SignedByAll` — paper: S3 §2.2 `signedByEveryone`.
- `Actor`, `AppliesAll`, `Reachable`, `update` — model devices.
- `PoolIsLive` — Armada rule ("a live note"). Its internal position quantifier is pool layout but does not appear in statement text.
- `PoolSignatureLiveness` — a stated assumption (EdDSA completeness).
- `genesis` — `channels := emptyStatus` (A13), `holdings := none` (section 4), `configuredAssets` (A1), `selfInputs` (A2), `now := 0` (A6), `signedBy`/`poolSignedBy := False` (legitimate), `pool` parameter (Armada rule: the pool exists before the adjudicator).
- `heldValue` — the concept is paper (S2 "funds held in channel A"). The definition is section 4.

### `Call` constructors and their `Step` cases

- `deposit` — paper: S2 §3.1 `D_A(x)`; Armada rule: funding by a note spend (labbook decision 21). Arguments: A1, A8, A14; merge branch: section 4.
- `challenge` — paper: S1 §4.3 `forceMove`, S3 Listing 1. The turn rules (`≥` when Open, `>` in Challenge) are S3 line 308 and S1 §4.5.
- `challenge`'s `challenger` argument and its signature check (`Operations.lean:261-262`) — paper: S1 §4.5 (lines 1545-1550):
  "Add a certificate argument to the force-move operation, to identify the challenger and prove that they
  intended to play that force-move. Require that the challenger is one of the participants."
  So `challenge_requires_authorized_challenger` (A-05) has a paper source. The audit first listed it as
  S4-only and then withdrew that finding. [INFERENCE] One defect remains, of a different kind: the model
  folds the `'forceMove'` string into `stateHash fp candidate` (`Operations.lean:246-252`). So the certificate
  is the same digest as a state signature and does not prove the intent that S1 §4.5 requires.
- `checkpoint` — name is S4 only. S1 §4.4 supports it: each response raises the turn number and cancels the challenge. Legitimate. The cleared hashes: A13.
- `conclude` — paper: S1 §4.4.4, S2 Fig. 4, S3 §2.3. No payout in `conclude` is an Armada rule. The reset: A12.
- `transfer` — paper: S2 §5.1. Arguments: A3, A4, A5. No fee: Armada rule (`README.md:94`).
- `reclaim` — paper: S3 Listing 2. Arguments: A3, A4. Conditions: A15, A16.
- `bulkTransferAllAssets`, `bulkConcludeAndTransferAllAssets` — finding A9.
- `tick` — paper: S1 §4.1 "expire", S4 passive timeout. Model device for time.
- `sign`, `signNote` — model devices for S1 §3.1.3 and S3 §1.4.
- Absent operations: S2 `withdraw` (the Armada rule replaces it with a payout note); S2 `claim` (S3 `reclaim` replaces it); S1 `refute`/`respondWithMove`/`alternativeRespondWithMove` (`checkpoint` and `challenge` cover them). The absences are legitimate.

### Hypothesis groups of the 40 statements

- `sig`, `appRel` — legitimate parameters (S3 §2.2 application rule).
- `assets`, `selfInputs`, `pool0` — A1, A2; `pool0` is legitimate.
- `hreach`, `hReachPool` — legitimate (reachable from deployment).
- `hstep`, `hstep1`/`hstep2`/... , `happ`, `calls` — legitimate.
- `hnowpos` — A6.
- `hduration : 0 < fp.challengeDuration` — legitimate: a zero challenge period gives no time to respond (S1 §4.1, S2 §4.3).
- `hnotfinal`, `hfinal`, `hsigned`, `hp` — legitimate (S2 §4.4 concluded point, S3 §2.2).
- `hregistered`, `hfin`, `hoh`, `ChannelFinalizedWith` — the mode part is legitimate. The hash parts are A3.
- `hmode` — legitimate.
- `hdest`, `destIdx` — positional lookup; part of A4.
- `hassetunique` — A4.
- `hnoguarantee`, `hnog` — legitimate restriction to allocations. [INFERENCE] It keeps `reclaim` from rewriting the watched exit (S2 §6.2 ordering problem).
- `pk`, `nk`, `r`, `hh_dest` — Armada rule: a payout is a note to the recipient's key.
- `hxpos`, `haffordable` — paper: S2 §5.1.
- `hhonest` — paper: S2 §4.4 honest participant.
- `hhonestPool` — A11.
- Interleaving restrictions (`actor ≠ Actor.nitro h`, `actor ≠ Actor.pool pk`, no `.tick`) and `hnohigher` — model choices. [INFERENCE] The no-`.tick` restriction is stronger than S2 §4.3 but is not a layout error.
- `h`, `hij`, `h1`, `h2`, index bounds — Lean devices. The index form is A4.
- `hext` — A14. `hwrong` — A8. `hdec` — legitimate. `hder` (`TransferOnlyDerivedFrom`) — legitimate except its `indices` (A5).
- `hholds`, `hpre`, `hheld`, `hpos` — section 4 (holdings), except `hpos : 0 < held`, which is legitimate.
- `hfull`, `hzero`, `hdest ∈` — legitimate as payout facts. The statements around them are A7.
- `hashA`, `poolSigLive`, `ownedComplete` — stated assumptions. Legitimate.
- `challenger` (in `challenge_requires_authorized_challenger`, `challenge_can_happen`, `unbeatable_pays_after_timeout`, and the `.challenge` steps of `finalized_outcome_was_concluded_or_challenged` and `turn_number_record_monotone`) — paper: S1 §4.5 certificate. See the `challenge` lines above.
