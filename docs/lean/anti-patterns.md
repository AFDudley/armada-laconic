# Anti-patterns, each one measured

Each entry gives the evidence (file, declaration, line count at `68f49f6`), why
it is wrong, and the rule in `writing-lean.md` or `worker-briefs.md` that
prevents it. Line counts come from a declaration index built on 2026-10-26
(the script in `verification.md`, section 6). Sources for the measurements:
`labbook/2026-10-26-lean-adjudicator-proof-size.md` and `docs/process.md`.

## Code anti-patterns

### A. The adjudicator copies the pool's step

**Evidence.** `DepositStep` (`ArmadaLean/Adjudicator/Notes.lean:169-195`, 27
lines of body) and `PoolOwnedSpendStep` (`Notes.lean:278-339`) restate the
body of `Step.transact` (`Pool.lean:1414-1442`) and `Step.transactOwned`
(`Pool.lean:1492-1514`). `Lemmas.lean` and `Safety.lean` name the pool's
internal operators (`removeAll`, `markSpent`, `insertResult`, `applyRecords`,
`applyOwnedRecords`, `withTreePositionsFrom`, `transactSpentPositions`) 265
times (146 in `Lemmas.lean`, 119 in `Safety.lean`). The re-proved facts total
2,598 lines. Examples: `PoolOwnedSpend_update_position_live_core`
(`Lemmas.lean:2379`, 219 lines) re-proves `spend_live_exact`
(`PoolSafety.lean:376`); `PoolOwnedSpendStep_notedValue_conserved`
(`Lemmas.lean:2758`, 65 lines) re-proves `transactOwned_balance`
(`PoolSafety.lean:62`); `adjudicator_pool_step` (`Lemmas.lean:1074`, 95 lines)
proves only that the copy agrees with the pool step.

**Why wrong.** Nothing in the types says a deposit is a pool step. Each pool
theorem must be re-derived on the copy. A change in `pool-model` must be
copied by hand into `Notes.lean` (the tree rollover merge at `a6cf16c` cost
one full re-adaptation of `Lemmas.lean` and `Safety.lean`,
`.scratch/rewrite-B/ADJUDICATOR-proofs-successor.md`, job 1 and 2).

**Rule.** `writing-lean.md` rule 1. `worker-briefs.md` design check D2 (no
pool-internal operator names in an adjudicator file).

### B. Every invariant walks all 11 calls

**Evidence.** `Adjudicator.Step` (`Operations.lean:133-580`) lists every
unchanged field as an equation in each of its 11 cases
(`Operations.lean:219-223, 268-270, 296-298, 330-332, 502-503, 514-515,
532-534, 558-560, 577-580`). Six lemmas prove only that one field is unchanged:
`step_configuredAssets` (`Lemmas.lean:1184`, 40 lines), `step_self` (1229),
`step_now_frame` (1277), `step_signedBy_mono` (1346), `step_signedBy_frame`
(1394), `poolSignedBy_frame` (1478); 293 lines together. The large invariants
repeat the per-call frame reasoning: `unbeatablePaysInv_step`
(`Safety.lean:1446`, 826 lines), `TimeoutInv_appliesAll` (3372, 362),
`finalized_outcome_was_concluded_or_challenged` (132, 284),
`challenge_mode_status_matches` (3175, 172), `holdingWellFormed_step`
(`Lemmas.lean:4741`, 169), `finalization_permanent` (773, 150),
`turn_number_record_monotone` (631, 142). 2,815 lines in total. `obtain ⟨…⟩`
appears 515 times in the two files (267 + 248).

**Why wrong.** The frame problem is solved by rule, not by a conjunct per
field (Borgida, Mylopoulos and Reiter 1995). With `{ s with … }` the unchanged
field is `rfl`. With a footprint predicate, an invariant about `channels` has
real cases only for the calls that write `channels`.

**Measured on `68f49f6`.** All six frame lemmas are proved by
`cases call <;> simp only [Adjudicator.Step] at hstep <;> grind` (one line
each, 1.5 s for all six). The footprint lemma `step_pool_frame` in
`writing-lean.md` rule 6 is proved the same way. `grind` is used zero times in
the model.

**Rule.** `writing-lean.md` rules 2, 6 and 8.

### C. Can-happen statements are hand-built executions

**Evidence.** `transfer_after_transfer_can_happen` (`Safety.lean:5921`, 494
lines), `topup_can_happen` (4621, 342), `payout_can_happen` (5119, 290),
`bulk_conclude_and_transfer_can_happen` (5648, 273),
`guarantee_reallocation_order_dependent` (4435, 186), `reclaim_can_happen`
(5409, 146), `deposit_can_happen_witness` (4273, 126), `checkpoint_can_happen`
(4157, 116), `conclude_can_happen` (5058, 61), `challenge_can_happen` (4963,
52), `challenge_timeout_can_happen` (5015, 43). 2,255 lines. Each builds
literal states, witnesses, Merkle paths and range-check bits from `genesis`.
The general lemma `transfer_step_exists` (`Safety.lean:2656`, 389 lines) gives
only a lower bound on the payout, so no can-happen proof can use it.

**Why wrong.** A general lemma with an exact effect is written once and
instantiated 11 times. A literal execution is written 11 times and each copy
breaks when a definition changes.

**Rule.** `writing-lean.md` rule 7.

### D. A property over free data where a type could carry it

**Evidence.** `AdjudicatorState.holdings : Nat → Term → Option ((Nat × Nat) ×
Term)` (`State.lean:62`). `HoldingWellFormed` (`Lemmas.lean:3997-4004`)
restates five facts about each holding. Its closure is 26 declarations and
1,866 lines: `DepositMergeStep_update_wellFormed` (`Lemmas.lean:4455`, 286),
`DepositStep_update_wellFormed` (4285, 170), `holdingWellFormed_step` (4741,
169), `PoolOwnedSpend_update_wellFormed` (4025, 110),
`DepositStep_frame_wellFormed` (4193, 92), `bulkTransferFold_wellFormed`
(4135, 58), and the source-A lemmas above.

**Why wrong.** `docs/process.md` Stratum 1: an illegal state must be
unrepresentable, not detected. exophial4's `drs.bend` went from 4,349 lines
and 153 laws to 665 lines when scope became a type
(`exophial4/docs/labbook/a-scoped-drs-type.md`).

**Limit.** The facts "live in the pool" and "preimage recorded" depend on the
pool state. They come from pool theorems after A, not from the holding's
type. The holding's type carries the adjudicator's own facts only (asset,
random, positive value). This change alters the shape of `holdings` in 15
statement signatures (`verification.md`, section 4 lists them). It needs the
operator's approval before any code changes.

**Rule.** `writing-lean.md` rule 4. `reduction-plan.md` phase H.

### E. A combined effect lemma that no statement uses

**Evidence.** `PoolOwnedSpendStep_effects` (`Lemmas.lean:3638`, 142 lines).
It is named once outside its own declaration, in a comment
(`Lemmas.lean:4738`). No statement depends on it. 182 lines of the two files
are used by no statement.

**Why wrong.** Every line costs maintenance. A lemma with no caller is a
caller waiting to be written by hand a second time.

**Rule.** `worker-briefs.md` design check D1 (a new declaration must be used
by a named statement). `reduction-plan.md` step 0.1 deletes it.

### F. `grind` unused; `@[simp]` unused

**Evidence.** `grep -rn grind ArmadaLean` is empty. `grep -rn "@\[simp\]"
ArmadaLean` is empty. Every proof destructures or unfolds.

**Why wrong.** The core automation that closes the frame cases in one line is
present in `v4.33.1` and verified (B above). Without a `@[simp]` API every
proof repeats the unfolding.

**Rule.** `writing-lean.md` rules 3 and 8.

### G. Numeric addresses

**Evidence.** Before the symbolic redesign, addresses were `Nat` with
`% 2^160` arithmetic (`docs/process.md`, "Arithmetic on a hash"). Fixed at
`State.lean:92`: `self = Term.keccak selfInputs`.

**Why wrong.** A hash output has no arithmetic. Truncation lemmas and address
bounds appeared that no source supports.

**Rule.** `writing-lean.md` rule 11. `docs/process.md` Stratum 1.2.

### H. A definition replaced by `True`

**Evidence.** During the proof work a builder replaced a definition with
`True` (coordinator's record, this brief; the diff is not at `68f49f6`). The
statements that read it became vacuous and still built with no `sorry`.

**Why wrong.** `lake build` and a `sorry` count cannot detect a vacuous
statement. Only a design check can.

**Rule.** `verification.md` section 5 (no `True`/`False` body in a `def`;
signature diff). `worker-briefs.md` D4.

## Process anti-patterns

### P1. Acceptance that tests "no sorry" instead of the design

**Evidence.** Briefs in the 2026-10-23 to 2026-10-26 work accepted a chunk
when `lake build` passed with zero `sorry`. Sources A, B, C and D above all
passed that gate.

**Why wrong.** "No sorry" is necessary. It says nothing about reuse, size or
shape. A worker optimizes to the gate it is given.

**Rule.** `worker-briefs.md`: every brief has at least one design check that
a script runs, plus a line-count target.

### P2. Parallel workstreams split against their dependency

**Evidence.** The general step lemmas (`transfer_step_exists`) and the
can-happen proofs were built by different workers at the same time. The
general lemmas gave a lower bound; the can-happen proofs built literal
executions; neither could use the other.

**Why wrong.** The can-happen proofs depend on the exact form of the general
lemmas. Work that depends on an interface cannot start before the interface
is fixed.

**Rule.** `reduction-plan.md` states for each step what it depends on and
whether it may run in parallel. `worker-briefs.md` section "Interfaces".

### P3. A builder that stops to write a plan

**Evidence.** `docs/process.md`, "A plan instead of the work" (Bowen and
Hinchey 1995).

**Rule.** `worker-briefs.md` template: the plan is in the brief. The reader
executes the listed declarations in the listed order. A reader who finds the
plan wrong stops and reports the exact declaration and error; the reader does
not write a new plan. This holds for a person and for a model alike: a plan
written mid-task is a second specification, and two specifications diverge.

### P4. Rebase and force-push of published history

**Evidence.** A published branch was rebased and force-pushed during the proof
work (`.scratch/rewrite-B/ADJUDICATOR-proofs-successor.md`, rule "Never
rebase, never force-push").

**Rule.** `verification.md` section 7. `git merge` only. Push after each step
with `git push` without `--force`.

### P5. Coordinator claims about Lean made without checking

**Evidence.** Claims recorded in the coordinator's log: Mathlib "build time"
as a blocker (the project has no Mathlib; `lakefile.lean` lists no
dependency); "no Dolev–Yao library exists" (DyLean, github.com/BobDyLean/dylean,
core Lean, and LeanDY, github.com/SecPriv/leandy, exist; neither was
evaluated `[INFERENCE: repositories named in the coordinator's research list,
not opened for this file]`); "Lean uses Z3" (`grind` is a core tactic with its
own engines, *Lean Language Reference*, chapter "The `grind` tactic"; Lean
core has no Z3 dependency).

**Rule.** Every Lean claim in a brief or a plan is checked by a command on
`v4.33.1` or marked `[INFERENCE]`. `worker-briefs.md` section "Claims".

### P6. Transcribed constants and silent stops

**Evidence.** A worker reproduced hex constants by typing them and produced
truncated hex and repeated substrings. Another run exited with no output
(coordinator's record, 2026-10-26).

**Why wrong.** A long constant has no redundancy; one wrong character is a
different value and the type checker may still accept it (a `Term.val n`
with the wrong `n` builds). A run with no output gives no evidence and no
error to act on.

**Rule.** A constant longer than 16 characters is produced by a command
(`git rev-parse`, `sed -n 'N p'`, `grep`), never transcribed, by anyone.
Every brief ends with the acceptance script; a run with no script output is
a failed run. `worker-briefs.md` section "Executability".
