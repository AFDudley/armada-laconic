# Why the Lean adjudicator proofs are 11,500 lines

A working note, 2026-10-26, modeled on exophial4's `docs/labbook/a-scoped-drs-type.md`. It records where the proof
lines of the adjudicator model go, the sources of that cost, the published work that names each source, and what a
change would remove. Nothing here changes code. Pins: `AFDudley/armada-lean` `adjudicator-model-v4` at `68f49f6`;
`pool-model` at `7f2d1ec`; Lean `v4.33.1`, no mathlib.

## What the model is for

The adjudicator model is the specification of the armada-nitro adjudicator rewrite (`engineering/build-plan.md` S5).
Its 45 statements (32 safety statements, 11 can-happen statements, 2 payment guarantees) are the contract. Everything
else in `Lemmas.lean` and `Safety.lean` exists to prove them. All 45 are proved; every one depends only on
`EdDSAVerify`, `decodeGuarantee` and Lean's standard axioms.

## Where the lines go

Method: a script parsed every declaration in `Lemmas.lean` (4,925 lines) and `Safety.lean` (6,615 lines), built the
use graph by name, and sorted each declaration by what it does. A declaration "re-proves pool bookkeeping" when its
proof manipulates the pool's internal operations (`removeAll`, `markSpent`, `insertResult`, `applyRecords`,
`applyOwnedRecords`, `withTreePositionsFrom`, `transactSpentPositions`). A declaration "walks all calls" when it case
splits over every `Call` constructor.

| Source | Lines | Share |
|---|---:|---:|
| A. Pool bookkeeping re-proved at the adjudicator layer | 2,598 | 23% |
| B. Invariants and frame facts that walk all 11 calls | 2,815 | 24% |
| C. Can-happen statements built as literal executions | 2,255 | 20% |
| D. Other proofs (statements, outcome arithmetic, folds) | 3,690 | 32% |
| Unused by any statement | 182 | 2% |
| Total | 11,540 | |

There are 515 `obtain ⟨…⟩` destructurings in the two files.

## Source A: the adjudicator copies the pool's step instead of calling it

`DepositStep` (`Notes.lean`, 42 lines) and `PoolOwnedSpendStep` (90 lines) restate the body of the pool model's
`Step.transact` and `Step.transactOwned` inline: `removeAll` on the live set, `markSpent` on the nullifiers,
`insertResult` on the tree, `applyRecords` on the preimage registry. Nothing in the types says that an adjudicator
deposit IS a pool step. So no pool theorem applies to it, and the adjudicator layer re-proves the pool's facts on its
own copy:

| Adjudicator lemma (lines) | Pool theorem that already states the fact |
|---|---|
| `PoolOwnedSpend_update_position_live_core` (219), `PoolOwnedSpend_update_position_live` (160), `PoolOwnedSpendStep_payout_live` (55), `PoolOwnedSpendStep_fresh_position_bounds` (29) | `spend_live_exact` |
| `PoolOwnedSpendStep_preserves_user_note` (64), `deposit_spend_requires_signer` (33) | `transact_spends_only_owned`, `transactOwned_spends_only_owned` |
| `PoolOwnedSpendStep_notedValue_conserved` (65), `DepositStep_notedValue_conserved` (30) | `transact_balance`, `transactOwned_balance`, `core_invariant` |
| `PoolOwnedSpend_update_preimage_nonzero` (81), `PoolOwnedSpendStep_preimageOf_eq` (82), `_preimageOf_known` (30), `_ownedRandom_tail` (121) | `pool_preimageOf_consistent`, `pool_ownedRandomOf_consistent`, `npk_preimage_not_zero` |
| `PoolOwnedSpendStep_reachable` (78), `adjudicator_pool_step` (95), `bulkTransferFold_pool_reachable` (48) | none needed: these prove that the copy agrees with the pool step |

The `DRS.bend` note names the same pattern: "nothing in the types relates" the value to the structure that governs it,
so laws must prove the relation after the fact.

**Published remedy.** Compose specifications; do not restate them. A system built from components reuses the
components' proved properties through a composition theorem [Abadi and Lamport, "Composing Specifications", ACM TOPLAS
15(1), 1993]. Event-B's shared-variable decomposition gives the same rule for refinement [Abrial 2010, ch. 13]. Here:
define `DepositStep` as "a pool `Step.transact` from `before` to `after`, and the adjudicator's own constraints on the
output it keeps", and the same for the owned spend. Then `adjudicator_pool_step` holds by definition and every pool
theorem applies directly.

## Source B: every invariant walks all 11 calls

`Step` (`Operations.lean`, 450 lines) states, for each of the 11 calls, the whole post-state, including every field the
call leaves alone (`s'.configuredAssets = s.configuredAssets ∧ s'.selfInputs = s.selfInputs ∧ s'.now = s.now ∧ …`).
Every invariant then destructures every case. Six lemmas do nothing else but show one field is unchanged:
`step_self`, `step_configuredAssets`, `step_now_frame`, `step_signedBy_frame`, `step_signedBy_mono`,
`poolSignedBy_frame` (293 lines). The big invariants repeat the same per-call frame reasoning: `unbeatablePaysInv_step`
(826), `TimeoutInv_appliesAll` (362), `finalized_outcome_was_concluded_or_challenged` (284), `challenge_mode_status_matches`
(172), `holdingWellFormed_step` (169), `finalization_permanent` (150), `turn_number_record_monotone` (142).

**Published remedy.** This is the frame problem in procedure specifications [Borgida, Mylopoulos and Reiter, "On the
Frame Problem in Procedure Specifications", IEEE TSE 21(10), 1995]: state what each operation may change, and let
everything else be unchanged by rule, not by a conjunct per field. Dynamic frames [Kassios, FM 2006] make the
footprint a first-class part of the specification. Here: define each call as an update of its own footprint
(`{ s with channels := …, holdings := … }`), so unchanged fields are equal by construction, and add one lemma per call
that names its footprint (the channel and the holdings it touches). An invariant about channel `cid` then has real
cases only for calls whose footprint includes `cid`.

## Source C: can-happen statements are hand-built executions

Each of the 11 can-happen statements builds its own execution from `genesis` with literal states, witnesses, Merkle
paths and range-check bits (2,255 lines; `transfer_after_transfer_can_happen` alone is 494). The general step lemmas the
payment guarantees use (`transfer_step_exists`, `conclude_step_finalizes`) give only a lower bound on the payout, so the
can-happen proofs cannot be built from them.

**Remedy.** Prove general existence lemmas with exact effects (for each call: from any reachable state that meets the
call's precondition, the step exists, and its result is exactly this), then derive each can-happen statement by
instantiating them. This is the stratified design rule: the parts of one layer are the primitives of the next [Abelson
and Sussman 1996, §2.2.4].

## Source of holding well-formedness

A holding stores a free `((treeNumber, position), commitment)`; `HoldingWellFormed` is a predicate proved preserved by
every step (its closure is 26 declarations, 1,866 lines, mostly counted in A and B). This is the same shape as the
DRS case: a property proved about free data, where the type could carry it [Minsky 2011, "make illegal states
unrepresentable"; Bird and Paterson 1999 for the scoped case]. After A, most of `HoldingWellFormed` (liveness, the
recorded preimage, the owner record) follows from pool theorems, and what is left is the adjudicator's own fact: the
holding names a note an adjudicator step created, keyed `npk self r`.

## What a change would remove

[INFERENCE] Not measured. Each estimate assumes the remedy above and the same statements.

| Source | Lines now | Estimated after | Reason |
|---|---:|---:|---|
| A | 2,598 | 300 to 600 | the pool facts come from pool theorems; what stays is the adjudicator's own constraint on its output |
| B | 2,815 | 800 to 1,200 | frame cases become definitional; invariants keep only the calls that touch their channel |
| C | 2,255 | 300 to 600 | each statement instantiates general exact step lemmas |
| D | 3,690 | 2,500 to 3,200 | some folds and frames shrink with A and B; the outcome arithmetic stays |
| Total | 11,540 | about 4,000 to 5,600 | |

The order follows the dependencies: A first (it changes `Notes.lean` and removes the most), then B (it changes `Step` in
`Operations.lean`), then C (it uses both). The first measurement should be A on `DepositStep` alone: redefine it over
`Step.transact`, re-prove the deposit cases, and count.

Statement texts do not change in any of the three: the statements speak through `Step`, `heldValue`, `PoolIsLive` and
the pool's public state, not through the inline bookkeeping.
