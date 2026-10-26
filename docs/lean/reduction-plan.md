# Proof-size reduction plan

This file is the execution layer. It breaks the reduction described in
`labbook/2026-10-26-lean-adjudicator-proof-size.md` into steps sized for the
5090 worker: one file or one small group of declarations, at most 9 points,
exact declarations, exact commands, script-checkable acceptance. The rules
behind each step are in `writing-lean.md`; the checks are in
`verification.md`; the brief format is in `worker-briefs.md`.

Baseline: `adjudicator-model-v4` at `68f49f6`; `pool-model` at `7f2d1ec`.
Lines at baseline: `Lemmas.lean` 4,941; `Safety.lean` 6,699; `Notes.lean`
402; `Operations.lean` 582. The full adjudicator build takes 11 s on this
host (measured 2026-10-26), so every step can rebuild from scratch.

## Execution constraints (this file only)

- Worker: Qwen ~27B on vLLM (`sonic` role), tools `bash`, `grep`, `glob`,
  3600 s per task, one task at a time; promoted to Sonnet on context
  overflow. Every brief must therefore be executable by `bash` alone.
- Points follow `engineering/estimation-method.md`: one point is 10,000 tokens of
  peak context one worker holds at once, and a chunk for the 5090 is at most 9
  points. The figures in this file are estimates from a proxy (about 1 point per
  declaration under 100 lines rewritten or added, 3 per declaration of 100 lines
  or more, 1 per merge, 1 per measurement report). Re-estimate a step by the
  method before dispatch if it reads more than its own file and one dependency.
- Every step ends with `.scratch/check.sh <baseline> [allowlist]` printing
  `EXIT 0` (`verification.md` section 1), a commit, and a push. The baseline
  for every step is `68f49f6` unless the step says otherwise. The commit
  subject starts with the step id.
- "Parallel" below means two steps may run at the same time on two workers
  (one 5090 worker plus Sonnet workers). Two steps that edit the same file
  may run in parallel only when they edit disjoint declarations; the second
  to finish merges the first (`git merge`, never rebase).
- A step marked **measurement** has a report as its product. Its line-count
  target is "report the numbers". The step after it uses the numbers.
- A target marked `[INFERENCE]` is an estimate. A step that misses an
  estimate still passes if `check.sh` passes; the coordinator re-plans the
  next step from the measured number.

## Verified facts the plan rests on (2026-10-26, `68f49f6`)

| Fact | How checked |
|---|---|
| The six frame lemmas are each proved by `cases call <;> simp only [Adjudicator.Step] at hstep <;> grind` (`step_self` with `grind [AdjudicatorState.self]`) | `lake env lean` trial, then a full build on a scratch copy: `Lemmas.lean` 4,941 → 4,718, axioms unchanged |
| `Call.touchesPool` and `step_pool_frame` (`writing-lean.md` rule 6) elaborate | `lake env lean` trial |
| `update_same`, `update_other`, `ChannelStatus.mode_eq_open_iff`, `mode_eq_finalized_iff`, `mode_eq_challenge_iff`, `challengePost`, `step_challenge_iff` (texts in step B1) elaborate | `lake env lean` trial |
| One-line `grind` on `turn_number_record_monotone` closes the 4 calls outside the `channels` footprint and fails on the 7 inside it, with or without the `mode` API | `lake env lean` trial |
| `simp only` and `rfl` read through an `abbrev` post-state in a constructor result (pattern for A0) | `lake env lean` trial on a small model, `writing-lean.md` examples |
| `#print axioms` approved set; signature script; `check.sh` | `verification.md` |

Everything else below is design, marked where it is an estimate.

## Step table

| Id | Phase | File | Points | Depends on | Parallel with |
|---|---|---|---:|---|---|
| 0.1 | setup | `.scratch/` | 1 | none | all |
| 0.2 | cleanup | `Lemmas.lean` | 3 | 0.1 | B0, A0 |
| B0 | B measurement | `Lemmas.lean` | 6 | 0.1 | 0.2, A0 |
| A0 | A (pool side) | `Pool.lean` on `pool-model` | 3 | 0.1 | 0.2, B0 |
| A0m | A merge | adjudicator worktree | 1 | A0 | none |
| A1 | A measurement | `Notes.lean` | 7 | A0m | B1 |
| B1 | B API | `State.lean`, `Channel.lean`, `Operations.lean` | 7 | B0 | A1 |
| B2 | B measurement | `Safety.lean` | 4 | B1 | A2 |
| A2 | A | `Lemmas.lean`, `Safety.lean` | 8 | A1 | B2 |
| A3 | A | `Notes.lean` | 7 | A1 | B3a |
| B3a | B | `Operations.lean` | 6 | B1 | A3 |
| A4 | A | `Lemmas.lean` | 9 | A3 | A5, A6 |
| A5 | A | `Lemmas.lean` | 8 | A3 | A4, A6 |
| A6 | A | `Lemmas.lean` | 8 | A3 | A4, A5 |
| A7 | A | `Notes.lean`, `Lemmas.lean` | 9 | A3 | B3b |
| A8 | A | `Lemmas.lean` | 6 | A1, A3, A7 | B4.x |
| B3b | B | `Operations.lean` | 9 | A1, A3, A7, B3a | A8 |
| B4.1 … B4.7 | B | `Safety.lean`, `Lemmas.lean` | 3 to 9 each | B3a, B3b (B4.6, B4.7 also A8) | each other (disjoint declarations) |
| B5 | B cutover | `Operations.lean` | 9 | all B4.x, C1, C2 | none |
| C1 | C | `Safety.lean` | 6 | B3a | A4 to A8 |
| C2 | C | `Safety.lean` | 9 | A8, B3b | B4.x |
| C3 | C | `Safety.lean` | 6 | C1 | C4 |
| C4 | C | `Safety.lean` | 9 | C2 | C3 |
| C5 | C | `Safety.lean` | 9 | C2, C4 | C6 |
| C6 | C | `Safety.lean` | 9 | C2, C4 | C5 |
| C7 | C | `Safety.lean` | 6 | C2 | C5, C6 |
| H0 | H approval | `docs/` only | 1 | A2 | all |
| H1 … | H | `State.lean` … | per approval | H0 approved, A8, B5 | none |

## Phase 0: setup and cleanup

### 0.1 Install the scripts and record the baseline

- File: `.scratch/sigs.py`, `.scratch/declsizes.py`, `.scratch/check.sh`
  from `verification.md` section 1.
- Commands: the three heredocs, then `.scratch/check.sh 68f49f6`.
- Acceptance: output equal to the expected block in `verification.md`
  section 1. Nothing is committed (`.scratch/` is ignored; confirm with
  `git check-ignore .scratch/`).
- Points: 1.

### 0.2 Delete the unused effect lemma

- File: `ArmadaLean/Adjudicator/Lemmas.lean`.
- Declarations: delete `PoolOwnedSpendStep_effects` (line 3638, 142 lines,
  with its doc comment). Edit the comment at line 4738 that names it: delete
  the parenthesis `(built on `PoolOwnedSpendStep_effects`)`.
- Commands: `python3 - <<'PY'` with an exact-match deletion from the doc
  comment start (`sed -n '3600,3640p'` first to find the `/--` line) to the
  line before the next `/--`; `lake build`; `.scratch/check.sh 68f49f6`.
- Acceptance: `check.sh` EXIT 0; `declsizes.py Lemmas.lean
  PoolOwnedSpendStep_effects` prints `absent`; `grep -c
  PoolOwnedSpendStep_effects ArmadaLean/Adjudicator/*.lean` prints 0 for every
  file; `wc -l < Lemmas.lean` ≤ 4,800.
- Points: 3.

## Phase A: compose the adjudicator's steps over the pool's step

### A0 Name the pool's post-states (`pool-model` branch)

This is a pool change. It goes to `pool-model` in
`~/code/git_puller/repos/armada-lean-pool-model`, by the pool builder, then
merges into the adjudicator branch (A0m). Rule: a pool fact belongs to the
pool model.

- File: `ArmadaLean/Pool.lean`.
- Declarations: add, immediately before `inductive Step` (line 1290):

  ```lean
  /-- The state after `Step.transact`, as a structure update. -/
  abbrev Step.transactPost (s : PoolState) {N M : Nat} (w : TransactWitness N M)
      (treeNumber : Nat) (newNotes unshieldTail : List CommitmentPreimage)
      (rest : List ((Nat × Nat) × CommitmentPreimage)) (spent' : NullifierSet)
      (records : List (Term × CommitmentPreimage))
      (ownedRecords : List (Term × OwnedDisclosure)) : PoolState :=
    <the exact `let … { s with … }` expression at Pool.lean:1430-1442>

  /-- The state after `Step.transactOwned`, as a structure update. -/
  abbrev Step.transactOwnedPost (s : PoolState) {N M : Nat} (w : TransactOwnedWitness N M)
      (pub : TransactOwnedPublic N M) (treeNumber : Nat)
      (rest : List ((Nat × Nat) × CommitmentPreimage)) (spent' : NullifierSet)
      (ownedOutputs : List (Fin M × OwnedDisclosure)) : PoolState :=
    <the exact expression at Pool.lean:1504-1514>
  ```

  Then replace the result expression of the `transact` constructor by
  `Step s (Step.transactPost s w treeNumber newNotes unshieldTail rest spent'
  records ownedRecords)` and of `transactOwned` by `Step s
  (Step.transactOwnedPost s w pub treeNumber rest spent' ownedOutputs)`.
  Checked 2026-10-26: an `abbrev Foo.post` declared before `inductive Foo`
  elaborates, and a constructor may return `Foo n (Foo.post n)`. The
  parameter list above is read off the constructor at Pool.lean:1414-1442;
  the worker copies the body with `sed -n`, never by retyping.
- Commands: `python3 - <<'PY'` exact replacement; `lake build`.
- Acceptance: `lake build` EXIT 0 on `pool-model`; `git diff --stat` touches
  only `ArmadaLean/Pool.lean`; `grep -c "Step.transactPost\|Step.transactOwnedPost"
  ArmadaLean/Pool.lean` prints ≥ 4; no proof text in `PoolSafety.lean`,
  `PoolOwned.lean`, `TransactOwned.lean` changes. If a pool proof breaks
  because `abbrev` is not transparent enough for it, the worker reports the
  first error and stops; the coordinator decides between `abbrev` and
  `@[reducible] def`.
- Points: 3.

### A0m Merge `pool-model` into `adjudicator-model-v4`

- Commands: in the adjudicator worktree, `git fetch AFDudley.github.com &&
  git merge AFDudley.github.com/pool-model`; `.scratch/check.sh 68f49f6`.
- Acceptance: `check.sh` EXIT 0; `git log --oneline -1` shows a merge commit.
- Points: 1. Record the merge sha as `<A0m>`; later A steps use it as the
  baseline.

### A1 Compose `DepositStep` over `Step.transactPost` (measurement)

- File: `ArmadaLean/Adjudicator/Notes.lean`.
- Declarations: `DepositStep` (line 169). New body: keep the first five
  conjuncts (lines 172-176) and the existential binders through
  `records` (lines 177-184). Replace the `after = (let res := … { before with
  … })` conjunct (lines 185-195) by

  ```lean
      after = Step.transactPost before w spendTreeNumber (transactAllOutputs w) [] rest spent'
        records [(pub.commitmentsOut idx, (⟨self, r⟩ : OwnedDisclosure))]
  ```

  Add, after `DepositStep`:

  ```lean
  /-- A `DepositStep` is a pool `Step.transact`. -/
  theorem DepositStep_step {N M : Nat} {before after : PoolState} {self : Term}
      {token amount : Nat} {w : TransactWitness N M} {pub : TransactPublic N M} {idx : Fin M}
      {spendTreeNumber treeNumber position : Nat}
      (h : DepositStep before self token amount w pub idx spendTreeNumber treeNumber position
        after) : _root_.Step before after
  ```

  proved by `obtain` of the witnesses, `subst`, then `exact Step.transact w
  pub spendTreeNumber hN hM hC (transactAllOutputs w) [] rest spent' records _
  hRoot (List.append_nil _).symm (by simp) hRemove hNullify hRecords ?_`
  where the last goal (`hOwnedRecords`) is `⟨_, hidxrec, hown.symm ▸ rfl⟩`-shaped.
  `[INFERENCE: this proof text is the shape of the existing deposit case of
  `adjudicator_pool_step` (Lemmas.lean:1087-1100), which already builds this
  term; the worker copies that case. Not type-checked here.]`

  Note the `held` field: `Step.transactPost` writes `held := fun t => if t =
  w.token then s.held t - listSum ([].map …) else s.held t`. The old
  `DepositStep` left `held` unchanged. The one adjudicator use of `.held`
  (`Safety.lean`, one hit) needs `simp` with `List.map_nil`, `listSum`'s nil
  case and `Nat.sub_zero`, `ite_self`. Add `@[simp] theorem
  transactPost_held_nil` in `Notes.lean` stating `(Step.transactPost s w tn
  newNotes [] rest spent' records owned).held = s.held` `[INFERENCE: statement
  shape; proof by `funext t; simp [Step.transactPost, listSum]`]`.
- Then rewrite the deposit case of `adjudicator_pool_step`
  (`Lemmas.lean:1080-1100`) to `exact DepositStep_step hdstep`-shaped use, and
  fix every proof that `unfold DepositStep` or destructures its `after`
  (`grep -n "DepositStep" Lemmas.lean Safety.lean` lists them: at `68f49f6`,
  `Lemmas.lean` deposit lemmas and `Safety.lean:3916-3919, 4374, 4768, 4819,
  5219, 5747`). A can-happen proof that builds a `DepositStep` witness now
  supplies the same witnesses and `rfl` for the post-state.
- Commands: `python3 - <<'PY'` replacement; `lake build`;
  `.scratch/check.sh <A0m>`.
- Acceptance: `check.sh <A0m>` EXIT 0; D2: `sed -n '/^def DepositStep/,/^theorem
  DepositStep_step/p' Notes.lean | grep -cE "removeAll|markSpent|insertResult|applyRecords|applyOwnedRecords|withTreePositionsFrom"`
  prints ≤ 2 (the `removeAll`/`markSpent` hypotheses stay; the post-state
  copy is gone); D3: the same range contains `Step.transactPost`; D1:
  `grep -c "DepositStep_step" Lemmas.lean Safety.lean` ≥ 2 in total;
  **measurement**: report `declsizes.py` for every declaration whose name
  starts with `DepositStep` or `deposit_` in `Lemmas.lean` and `Safety.lean`,
  before and after.
- Points: 7 (DepositStep 1, DepositStep_step 1, transactPost_held_nil 1,
  adjudicator_pool_step deposit case 1, up to 3 broken proofs at 1 each).
  If more than three proofs break, the worker fixes the first three, leaves
  `sorry` in the others, commits `WIP:` and reports; the coordinator issues
  A1b.

### A2 Deposit lemmas through pool theorems

- Files: `Lemmas.lean`, `Safety.lean`.
- Declarations (sizes at `68f49f6`): `DepositStep_notedValue_conserved`
  (`Safety.lean:3912`, 30) via `transact_balance` (`PoolSafety.lean:45`);
  `DepositStep_update_wellFormed` (`Lemmas.lean:4285`, 170) and
  `DepositStep_frame_wellFormed` (4193, 92) via `spend_live_exact`
  (`PoolSafety.lean:376`), `transactOutputRecords_valid` (`Pool.lean:623`),
  `applyRecords_returns` (`Pool.lean:559`), `insertResult_leavesAt_length_ge`
  (`Pool.lean:1103`); `deposit_sets_heldValue` (3780, 43);
  `deposit_spend_requires_signer` (1570, 33) via `transact_spends_only_owned`.
- Acceptance: `check.sh <A0m>` EXIT 0; the sum of `declsizes.py` LINES for
  the five names ≤ 200 (`68f49f6`: 368) `[INFERENCE]`; `grep -cE
  "removeAll|markSpent|insertResult|applyRecords|withTreePositionsFrom"` over
  the five declarations' ranges ≤ 5.
- Points: 8 (3 + 1 + 1 + 1 + 1 + 1 measurement).

### A3 Compose `PoolOwnedSpendStep` over `Step.transactOwnedPost`

- File: `Notes.lean`. Declaration: `PoolOwnedSpendStep` (line 278). In both
  `match` arms replace the `after = { before with … }` conjunct (lines
  301-311 and 329-339) by `after = Step.transactOwnedPost before w pub
  <treeNumber|spendTreeNumber> rest spent' <ownedOutputs>` where for the
  `none` arm `rest` is `before.live token` and `ownedOutputs` is
  `(List.finRange ownedPreimages.length).map fun j => (Fin.natAdd
  payoutPreimages.length j, (⟨self, ownedRandom j⟩ : OwnedDisclosure))`.
  Checked 2026-10-26: `example (l : List ((Nat × Nat) × CommitmentPreimage)) :
  removeAll l [] = some l := rfl` elaborates, so the `none` arm supplies
  `rest := before.live token` with `hRemove := rfl`. Add `PoolOwnedSpendStep_step : … →
  _root_.Step before after`, copied from the `transfer` case of
  `adjudicator_pool_step` and `PoolOwnedSpendStep_reachable` (`Lemmas.lean:948`).
- Acceptance: `check.sh <A0m>` EXIT 0; D2 over the `PoolOwnedSpendStep` range
  ≤ 4; D3 `Step.transactOwnedPost` present twice in the range; D1
  `PoolOwnedSpendStep_step` used by `PoolOwnedSpendStep_reachable` (its body
  becomes `Reachable.step hreach (PoolOwnedSpendStep_step hmatch)`, ≤ 8
  lines).
- Points: 7.

### A4 Owned-spend liveness through `spend_live_exact`

- File: `Lemmas.lean`. Declarations: `PoolOwnedSpend_update_position_live_core`
  (2379, 219), `PoolOwnedSpend_update_position_live` (2598, 160). Pool
  theorems: `spend_live_exact`, `removeAll_perm` (`Pool.lean:195`),
  `removeAll_subset` (216), `withTreePositionsFrom_pos_lt` (387),
  `insertResult_leavesAt_length_ge` (1103).
- Acceptance: `check.sh <A0m>` EXIT 0; sum of the two LINES ≤ 150
  (`68f49f6`: 379) `[INFERENCE]`.
- Points: 9 (3 + 3 + 3 for the callers that change).

### A5 Owned-spend value and ownership through pool theorems

- File: `Lemmas.lean`. Declarations: `PoolOwnedSpendStep_notedValue_conserved`
  (2758, 65) via `transactOwned_balance`; `PoolOwnedSpendStep_preserves_user_note`
  (2823, 64) via `transactOwned_spends_only_owned`;
  `PoolOwnedSpendStep_payout_live` (3248, 55) and
  `PoolOwnedSpendStep_fresh_position_bounds` (2350, 29) via `spend_live_exact`
  and `withTreePositionsFrom_pos_lt`; `PoolOwnedSpendStep_sets_heldValue`
  (2124, 59) via `transactOwned_records_all` (`PoolSafety.lean:522`).
- Acceptance: `check.sh <A0m>` EXIT 0; sum of the five LINES ≤ 150
  (`68f49f6`: 272) `[INFERENCE]`.
- Points: 8.

### A6 Owned-spend registry facts through pool theorems

- File: `Lemmas.lean`. Declarations: `PoolOwnedSpend_update_preimage_nonzero`
  (1818, 81) via `npk_preimage_not_zero` (`PoolOwned.lean:358`);
  `PoolOwnedSpendStep_preimageOf_eq` (1736, 82) and
  `PoolOwnedSpendStep_preimageOf_known` (3608, 30) via
  `pool_preimageOf_consistent` (`PoolSafety.lean:495`) and
  `applyRecords_returns`; `PoolOwnedSpendStep_ownedRandom_tail` (1899, 121)
  via `pool_ownedRandomOf_consistent` (`PoolSafety.lean:551`) and
  `applyOwnedRecords_reads` (`Pool.lean:710`).
- Acceptance: `check.sh <A0m>` EXIT 0; sum of the four LINES ≤ 140
  (`68f49f6`: 314) `[INFERENCE]`.
- Points: 8.

### A7 Compose `DepositMergeStep`; re-prove its well-formedness

- Files: `Notes.lean` (`DepositMergeStep`, line 368: replace lines 391-400 by
  `after = Step.transactOwnedPost before w pub spendTreeNumber rest spent'
  [(0, (⟨pub.ownerField, randomOut⟩ : OwnedDisclosure))]`; add
  `DepositMergeStep_step`), `Lemmas.lean`
  (`DepositMergeStep_update_wellFormed`, 4455, 286).
- Acceptance: `check.sh <A0m>` EXIT 0; D2 over the `DepositMergeStep` range
  ≤ 2; `DepositMergeStep_update_wellFormed` LINES ≤ 120 `[INFERENCE]`.
- Points: 9.

### A8 Collapse `adjudicator_pool_step`

- File: `Lemmas.lean`. Declarations: `adjudicator_pool_step` (1074, 95),
  `bulkTransferFold_pool_reachable` (1026, 48). Each case becomes one use of
  `DepositStep_step`, `DepositMergeStep_step`, `PoolOwnedSpendStep_step`, or
  `grind` for a call outside the pool footprint (`step_pool_frame` from
  B1).
- Acceptance: `check.sh <A0m>` EXIT 0; `adjudicator_pool_step` LINES ≤ 40;
  `bulkTransferFold_pool_reachable` LINES ≤ 25; no `obtain ⟨` with more than
  six components in either declaration (`grep -c "obtain ⟨[^⟩]*,[^⟩]*,[^⟩]*,[^⟩]*,[^⟩]*,[^⟩]*,"`
  over their ranges prints 0).
- Points: 6.

## Phase B: footprints and frames

### B0 Six frame lemmas by `grind` (measurement, verified)

The brief is written in full in `worker-briefs.md`, "A filled example".
Observed on a scratch copy: `Lemmas.lean` 4,941 → 4,718; build green;
axioms approved. Points: 6. Parallel with 0.2 and A0.

### B1 Reading-lemma API and footprint predicates

- Files and declarations (texts type-checked 2026-10-26 on `68f49f6`):
  - `State.lean`, after `update` (line 22):

    ```lean
    @[simp] theorem update_same {α : Type} [DecidableEq α] {β : Type} (f : α → β) (a : α) (v : β) :
        update f a v a = v := by simp [update]

    @[simp] theorem update_other {α : Type} [DecidableEq α] {β : Type} (f : α → β) (a : α) (v : β)
        (x : α) (h : x ≠ a) : update f a v x = f x := by simp [update, h]
    ```

  - `Channel.lean`, after `ChannelStatus.mode` (line 237):

    ```lean
    theorem ChannelStatus.mode_eq_open_iff (c : ChannelStatus) (now : Nat) :
        c.mode now = .open ↔ c.finalizesAt = 0 := by
      unfold ChannelStatus.mode; split <;> simp_all <;> split <;> simp_all

    theorem ChannelStatus.mode_eq_finalized_iff (c : ChannelStatus) (now : Nat) :
        c.mode now = .finalized ↔ c.finalizesAt ≠ 0 ∧ c.finalizesAt ≤ now := by
      unfold ChannelStatus.mode; split <;> simp_all <;> split <;> simp_all

    theorem ChannelStatus.mode_eq_challenge_iff (c : ChannelStatus) (now : Nat) :
        c.mode now = .challenge ↔ c.finalizesAt ≠ 0 ∧ now < c.finalizesAt := by
      unfold ChannelStatus.mode; split <;> simp_all <;> split <;> simp_all <;> omega
    ```

    (Do not tag these `@[simp]` yet: a `simp` call elsewhere that rewrites a
    `mode` hypothesis would change its shape; measure first in B2.)
  - `Operations.lean`, after `Step` (line 580): `Call.touchesPool` and
    `step_pool_frame` from `writing-lean.md` rule 6 (verified), plus
    `Call.touchesChannels : Call → Bool` (true for `challenge`, `checkpoint`,
    `conclude`, `transfer`, `reclaim`, `bulkTransferAllAssets`,
    `bulkConcludeAndTransferAllAssets`) and `step_channels_frame` with the
    conclusion `s'.channels = s.channels`, proved by the same one line
    `[INFERENCE: the channels variant was not run; it has the same shape as
    the verified pool variant]`.
- Acceptance: `check.sh 68f49f6` EXIT 0; D1: `step_pool_frame` is used by
  A8 later, so for this step the check is `lake env lean` of a trial file
  with `example … := step_pool_frame …` `[INFERENCE: trial text to be written
  by the brief author from the verified statement]`; `grep -c "@\[simp\]"
  ArmadaLean/Adjudicator/State.lean` prints 2.
- Points: 7.

### B2 One 11-call invariant through the API (measurement)

- File: `Safety.lean`. Declaration: `turn_number_record_monotone` (631, 142).
  Rewrite as `cases call` where the 4 calls outside the `channels` footprint
  are closed by `simp only [Adjudicator.Step] at hstep; grind` and the 7
  inside it keep hand-written cases that use `update_same`, `update_other`
  and the three `mode_eq_*_iff` lemmas instead of `unfold update; rw [if_pos
  rfl]` and `rcases hmode : … mode …`.
- Acceptance: `check.sh 68f49f6` EXIT 0; **measurement**: report LINES
  before and after and the count of `obtain ⟨` in the declaration. Target
  ≤ 90 `[INFERENCE]`.
- Points: 4.

### B3a Post-state lemmas for the six pure calls

- File: `Operations.lean`, after `Step`. For each of `tick`, `sign`,
  `signNote`, `challenge`, `checkpoint`, `conclude`: an `abbrev <call>Post`
  and a `theorem step_<call>_iff` with the shape of `challengePost` and
  `step_challenge_iff` (both verified; full text below). The `iff` is proved
  by `simp only [Adjudicator.Step]`, `constructor`, `rintro` of the conjuncts,
  `cases s'`, `simp only at …`, `subst …`, `rfl` in one direction and
  `rintro ⟨…, rfl⟩; exact ⟨…, rfl, …⟩` in the other.

  ```lean
  /-- Post-state of `.challenge`, as a structure update. -/
  abbrev challengePost (s : AdjudicatorState) (fp : FixedPart) (candidate : VariablePart) :
      AdjudicatorState :=
    { s with channels := (update s.channels (channelId fp)
        ({ turnNumRecord := candidate.turnNum, finalizesAt := s.now + fp.challengeDuration,
           stateHash := stateHash fp candidate, outcomeHash := outcomeHashOf candidate.outcome } :
          ChannelStatus)) }

  /-- The `.challenge` case of `Step`, read as a precondition and a post-state. -/
  theorem step_challenge_iff (sig : SignatureAssumptions)
      (appRel : AdjudicatorState → FixedPart → List VariablePart → VariablePart → Prop)
      (actor : Actor) (s s' : AdjudicatorState) (fp : FixedPart) (proof : List VariablePart)
      (candidate : VariablePart) (challenger : Nat) :
      Step sig appRel actor s (.challenge fp proof candidate challenger) s' ↔
        ((match (s.channels (channelId fp)).mode s.now with
          | .open => candidate.turnNum ≥ (s.channels (channelId fp)).turnNumRecord
          | .challenge => candidate.turnNum > (s.channels (channelId fp)).turnNumRecord
          | .finalized => False) ∧
         appRel s fp proof candidate ∧
         challenger ∈ fp.participants ∧
         AuthorizedBy sig s challenger (stateHash fp candidate)) ∧
        s' = challengePost s fp candidate := by
    simp only [Adjudicator.Step]
    constructor
    · rintro ⟨h1, h2, h3, h4, hch, hpool, hhold, hcfg, hself, hnow, hsb, hpsb⟩
      refine ⟨⟨h1, h2, h3, h4⟩, ?_⟩
      cases s'
      simp only at hch hpool hhold hcfg hself hnow hsb hpsb
      subst hch hpool hhold hcfg hself hnow hsb hpsb
      rfl
    · rintro ⟨⟨h1, h2, h3, h4⟩, rfl⟩
      exact ⟨h1, h2, h3, h4, rfl, rfl, rfl, rfl, rfl, rfl, rfl, rfl⟩
  ```

  The number of conjuncts differs per call; the worker reads each case's
  conjunct count from `Operations.lean` with `sed -n` and writes the
  `rintro` pattern with that many names. The `.sign` and `.signNote` cases
  also carry `actor = .nitro p` / `actor = .pool pk` in the precondition.
- Acceptance: `check.sh 68f49f6` EXIT 0; `grep -c "^abbrev .*Post\b"
  Operations.lean` prints 6; `grep -c "^theorem step_.*_iff"` prints 6; for
  each `<call>Post`, a trial `example` that the untouched fields are `rfl`
  (`worker-briefs.md` D8) elaborates.
- Points: 6 (the six `iff` lemmas; each under 30 lines).

### B3b Post-state lemmas for the five pool-touching calls

- File: `Operations.lean`. `deposit`, `transfer`, `reclaim`,
  `bulkTransferAllAssets`, `bulkConcludeAndTransferAllAssets`. Each
  `step_<call>_iff` has the shape `Step … ↔ <pre> ∧ ∃ <witnesses>, <witness
  facts> ∧ s' = { s with pool := …, holdings := …, channels := … }`. The
  witnesses are the ones already bound in the case (`Operations.lean:201-218`
  for deposit). The `DepositStep`/`DepositMergeStep`/`PoolOwnedSpendStep`
  predicates stay as they are after A1, A3, A7.
- Acceptance: `check.sh <A0m>` EXIT 0; `grep -c "^theorem step_.*_iff"`
  prints 11; D8 trial for each new `<call>Post`.
- Points: 9 (deposit 3, transfer 3, reclaim 1, bulk 1 + 1).

### B4.x Invariants by footprint

One step per invariant. Each rewrites the proof with the `step_<call>_iff`
lemmas: `rw [step_<call>_iff] at hstep` (or `obtain ⟨hpre, rfl⟩`), then the
frame cases close by `grind` and the footprint cases by `simp
[update_same, update_other, …]` and the `mode_eq_*_iff` lemmas.

| Step | Declaration (file:line, lines at `68f49f6`) | Points | Target LINES `[INFERENCE]` |
|---|---|---:|---:|
| B4.1 | `finalization_permanent` (`Safety.lean:773`, 150) | 3 | 60 |
| B4.2 | `challenge_mode_status_matches` (`Safety.lean:3175`, 172) | 3 | 70 |
| B4.3 | `finalized_outcome_was_concluded_or_challenged` (`Safety.lean:132`, 284) | 6 | 110 |
| B4.4 | `TimeoutInv_appliesAll` (`Safety.lean:3372`, 362) | 6 | 140 |
| B4.5 | `holdingWellFormed_step` (`Lemmas.lean:4741`, 169) | 6 | 80 |
| B4.6 | `unbeatablePaysInv_step` (`Safety.lean:1446`, 826), frame cases only | 9 | 500 |
| B4.7 | `unbeatablePaysInv_step`, footprint cases | 9 | 300 |

- Acceptance for each: `check.sh <baseline>` EXIT 0; LINES ≤ target or a
  report; D7: `grep -c "⟩ := hstep"` over the declaration prints 0.
- Dependencies: B3a for B4.1-B4.4; B3a, B3b and A8 for B4.5-B4.7. B4.1-B4.5
  may run in parallel with each other (disjoint declarations). B4.7 follows
  B4.6.

### B5 Cut over `Step` to the post-state form

- File: `Operations.lean`. Replace each case body by the right-hand side of
  its `step_<call>_iff` (precondition ∧ `s' = <call>Post …`). Delete the 11
  `iff` lemmas; every use `rw [step_<call>_iff] at hstep` becomes
  `obtain ⟨hpre, rfl⟩ := hstep`.
- Acceptance: `check.sh <baseline>` EXIT 0; `grep -c "s'\.configuredAssets =
  s\.configuredAssets" Operations.lean` prints 0; `grep -c "step_.*_iff"
  ArmadaLean/Adjudicator/*.lean` prints 0 in every file; `wc -l <
  Operations.lean` ≤ 450.
- Depends on all B4.x, C1 and C2 (so no proof still destructures the old
  shape). Points: 9.

## Phase C: exact-effect lemmas, then can-happen by instantiation

### C1 Exact-effect lemmas for the six pure calls

- File: `Safety.lean`, in the section that holds `sign_step` (4067) and
  `signNote_step` (4078). Add `tick_step`, `challenge_step`,
  `checkpoint_step`, `conclude_step`: from any `s` that meets the
  precondition of `step_<call>_iff`, `Step sig appRel actor s (.<call> …)
  (<call>Post s …)`. Each is `(step_<call>_iff …).mpr ⟨hpre, rfl⟩`.
  `conclude_step_finalizes` (2467, 36) becomes a corollary of
  `conclude_step` and `mode_eq_finalized_iff`.
- Acceptance: `check.sh 68f49f6 .scratch/allow-C1.txt` with allowlist
  `conclude_step_finalizes` EXIT 0 `[INFERENCE: whether its signature must
  change; if it need not, use no allowlist]`; D1: each new lemma is used by
  at least one can-happen proof after C3; for this step, `lake env lean` of a
  trial file that applies each lemma.
- Points: 6.

### C2 Exact-effect lemmas for `deposit` and `transfer`

- File: `Safety.lean`. Replace `transfer_step_exists` (2656, 389; lower bound)
  by `transfer_step` with the exact post-state: `Step … (.transfer …) {mid
  with pool := pool', holdings := …, channels := …}` where `pool'` is the
  `PoolOwnedSpendStep` witness built from `owned_spend_witness_exists`
  (`PoolSafety.lean:1074`) and `leaf_has_merkle_witness` (1036). Add
  `deposit_step` with the exact post-state from a `DepositStep` witness. The
  payment guarantees `unbeatable_pays` (3045) and `unbeatable_pays_after_timeout`
  (3734) keep their texts; their proofs use `transfer_step` and
  `transfer_step_pays_selected` (2345).
- Acceptance: `check.sh <baseline> .scratch/allow-C2.txt` with allowlist
  `transfer_step_exists` EXIT 0 (the name is deleted; `transfer_step` replaces
  it); `declsizes.py Safety.lean transfer_step_exists` prints `absent`;
  `transfer_step` LINES ≤ 200 `[INFERENCE]`.
- Points: 9.

### C3 to C7 Can-happen statements by instantiation

Each step rewrites the listed proofs so that every `Step` fact comes from a
`<call>_step` lemma and every pool fact from a pool theorem. No literal
`PoolState` or `AdjudicatorState` is built except `genesis`.

| Step | Declarations (lines at `68f49f6`) | Points | Target total LINES `[INFERENCE]` | Depends on |
|---|---|---:|---:|---|
| C3 | `checkpoint_can_happen` (116), `challenge_can_happen` (52), `challenge_timeout_can_happen` (43), `conclude_can_happen` (61) | 6 | 120 | C1 |
| C4 | `deposit_can_happen_witness` (126), `deposit_can_happen` (36), `topup_can_happen` (342) | 9 | 150 | C2 |
| C5 | `payout_can_happen` (290), `reclaim_can_happen` (146) | 9 | 150 | C2, C4 |
| C6 | `bulk_conclude_and_transfer_can_happen` (273), `transfer_after_transfer_can_happen` (494) | 9 | 200 | C2, C4 |
| C7 | `guarantee_reallocation_order_dependent` (186) | 6 | 80 | C2 |

- Acceptance for each: `check.sh <baseline>` EXIT 0; D6 over each
  declaration's range prints 0; LINES ≤ target or a report.
- C3 and C4 may run in parallel; C5, C6 and C7 may run in parallel with
  each other.

## Phase H: holdings as a type (needs the operator's approval)

### H0 Approval request (no code)

- Product: a note in `docs/lean/` or the labbook with: the 15 statements in
  `Safety.lean` whose signature names the shape of `holdings`
  (`verification.md` section 4); the proposed type

  ```lean
  /-- One channel holding: the note the adjudicator minted for `(asset, channelId)`.
      The commitment is derived, never stored. -/
  structure Holding where
    key : Nat × Nat        -- (treeNumber, position)
    random : Nat           -- `npk self random` is the note key
    value : Nat
    pos : 0 < value
  ```

  with `holdings : Nat → Term → Option Holding` and
  `Holding.commitment (self : Term) (asset : Nat) (h : Holding) : Term :=
  commitment (npk self h.random) asset h.value`; the facts the type then
  carries (`pre.npk = npk self r`, `pre.token = asset`, `pre.value ≠ 0`:
  three of the five conjuncts of `HoldingWellFormed`,
  `Lemmas.lean:3997-4001`); the facts that stay in the invariant (live at
  `key`, preimage recorded, distinct keys) and come from pool theorems after
  A; and the rewritten text of each of the 15 statements.
  `[INFERENCE: the type above is a proposal; the operator rules on it.]`
- Points: 1. Depends on A2 (so the request cites measured numbers).

### H1 and later

Issued only after H0 is approved. H1: the type in `State.lean` and
`heldValue` (133) through `Holding.value`. H2: `Step` cases that write
`holdings` (`Operations.lean`). H3: `HoldingWellFormed` reduced to the two
state-dependent conjuncts; `holdingWellFormed_step` re-proved. H4: the 15
statements re-stated per the approval, with the allowlist from H0. Each at
most 9 points; sizes set when H0 is approved.

## Order of execution on one 5090 worker

0.1, B0, 0.2, A0 (pool builder), A0m, B1, A1, B2, A2, A3, B3a, A4, A5, A6,
A7, A8, B3b, C1, B4.1, B4.2, B4.3, B4.4, B4.5, C2, B4.6, B4.7, C3, C4, C5,
C6, C7, B5, H0, then H1-H4 after approval.

With Sonnet workers in parallel: B0 ∥ A0; A1 ∥ B1; A2 ∥ B2; A3 ∥ B3a;
A4 ∥ A5 ∥ A6; A7 ∥ C1; B4.1-B4.5 ∥ each other; C3 ∥ C4; C5 ∥ C6 ∥ C7.

## Expected totals

From the labbook, `[INFERENCE]`: A 2,598 → 300-600; B 2,815 → 800-1,200;
C 2,255 → 300-600; D 3,690 → 2,500-3,200; total 11,540 → 4,000-5,600. B0
alone is measured: −223 lines. Every later estimate is replaced by the
measured number from the step's report.
