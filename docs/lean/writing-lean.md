# How Lean is written in this project

This file applies `docs/process.md` to Lean 4. It does not repeat that file.
Read `docs/process.md` first.

Mathlib: the specification trunk (armada-lean `integrated`, mirrored to the fork's `main`) and the branch
`armada-nitro-spec` require Mathlib `v4.33.1` on toolchain `leanprover/lean4:v4.33.1`. Use Mathlib where it makes
sense (operator ruling, 2026-10-27).
Before you write a general definition or lemma (one whose statement names no project type), search Mathlib for it.
`verification.md` §6a states the check.

Examples: `AFDudley/armada-lean`, branch `adjudicator-model-v4` at `68f49f6` (now reference only) and branch
`pool-model` at `7f2d1ec`. Those branches had no Mathlib. Every Lean example below was type-checked on `v4.33.1` on
2026-10-26 unless the example is marked `[INFERENCE]`. The check command was `~/.elan/bin/lake env lean <file>`
inside the adjudicator worktree.

Each rule has four parts: the rule, a Lean example, the project declaration
that shows the rule or breaks it, and the source that justifies it.

## 1. Compose existing definitions. Never copy a body.

**Rule.** A new definition calls an existing definition by name. A new
definition never restates the body of an existing definition.

```lean
structure S where
  a : Nat
  b : Nat

def bump (s : S) : S := { s with a := s.a + 1 }

/-- Correct: `twoBumps` is written through `bump`. -/
def twoBumps (s : S) : S := bump (bump s)
```

**In the project.** `DepositStep` (`ArmadaLean/Adjudicator/Notes.lean:169-195`)
breaks this rule. It restates the result of the pool's `Step.transact`
(`ArmadaLean/Pool.lean:1414-1442`) with `removeAll`, `markSpent`, `insertResult`
and `applyRecords`. The pool theorems `spend_live_exact`
(`PoolSafety.lean:376`), `transact_spends_only_owned` (`PoolSafety.lean:152`)
and `pool_preimageOf_consistent` (`PoolSafety.lean:495`) speak about
`Step.transact`. They do not apply to the copy. The adjudicator re-proves them
(2,598 lines, `labbook/2026-10-26-lean-adjudicator-proof-size.md`, source A).

**Source.** Abadi and Lamport, "Composing Specifications", ACM TOPLAS 15(1),
1993: a composed system reuses the proved properties of its components.
`docs/process.md`, practice 12: "Own every consequence of a change."

## 2. State is a structure. A step updates its footprint with `{ s with … }`.

**Rule.** A state is a `structure`. A step writes the new state as
`{ s with field := … }`. It names only the fields it changes. A field it does
not name is then equal to the old field by `rfl`.

```lean
def bump (s : S) : S := { s with a := s.a + 1 }

/-- The unchanged field needs no lemma. -/
theorem bump_b (s : S) : (bump s).b = s.b := rfl
```

**In the project.** `PoolState` (`Pool.lean:1187`) and `Step.transact`
(`Pool.lean:1433-1442`) follow the rule: the constructor's result is
`{ s with live := …, spent := …, … }`. `Adjudicator.Step`
(`ArmadaLean/Adjudicator/Operations.lean:133-580`) breaks it. Each of the 11
cases lists every unchanged field as an equation, for example
`s'.configuredAssets = s.configuredAssets ∧ s'.selfInputs = s.selfInputs ∧
s'.now = s.now` (`Operations.lean:220-223`). Every invariant then
destructures every case with `obtain ⟨_, _, _, _, _, _, h, _, _, _, _⟩`
(515 such destructurings in `Lemmas.lean` and `Safety.lean`).

**One caution, checked.** A function-valued field written as
`fun t => if t = k then f t else f t` is not `rfl`-equal to `f`. The rewrite
`if_pos`/`ite_self` is needed. `simp` closes it:

```lean
example (f : Nat → Nat) (k t : Nat) : (if t = k then f t else f t) = f t := by simp
```

So write a step's unchanged function field as the old field, never as a
trivial `if`.

**Source.** Borgida, Mylopoulos and Reiter, "On the Frame Problem in Procedure
Specifications", IEEE TSE 21(10), 1995: state what an operation changes; the
rest is unchanged by rule. *Functional Programming in Lean*, "Structures":
structure update syntax.

## 3. Each definition gets a small `@[simp]` API. Proofs use the API.

**Rule.** After each definition, state its reading lemmas and tag them
`@[simp]`. A proof about the definition calls `simp` and the API. A proof does
not `unfold` the definition and does not destructure its body.

```lean
@[simp] theorem bump_a (s : S) : (bump s).a = s.a + 1 := rfl
@[simp] theorem bump_b' (s : S) : (bump s).b = s.b := rfl

theorem twoBumps_a (s : S) : (twoBumps s).a = s.a + 2 := by
  simp [twoBumps]
```

For a function-valued field updated at one key, the API has two lemmas:

```lean
def upd (f : Nat → Nat) (k v : Nat) : Nat → Nat := fun x => if x = k then v else f x
@[simp] theorem upd_same (f : Nat → Nat) (k v : Nat) : upd f k v k = v := by simp [upd]
@[simp] theorem upd_other (f : Nat → Nat) (k v x : Nat) (h : x ≠ k) : upd f k v x = f x := by
  simp [upd, h]
```

**In the project.** `@[simp]` appears zero times in `ArmadaLean/` at
`68f49f6` (`grep -rn "@\[simp\]" ArmadaLean` is empty). `update`
(`State.lean:21`) has no reading lemmas at all; every proof that reads
`update s.holdings asset …` unfolds it. `Pool.lean` has the API shape, without
the attribute, for `withTreePositionsFrom` (`withTreePositionsFrom_map`,
`_treeNumber`, `_pos_ge`, `_unique`, `_pos_lt`, `_nodup`, `_snd_mem`,
`Pool.lean:293-442`). That is the shape to copy, with `@[simp]` added where the
right-hand side is simpler than the left.

**Source.** Lean core `doc/std/naming.md` ("Naming algorithm for theorems"):
reading lemmas are named after the definition and the projection.
Mathlib's convention of a simp-normal form with small lemmas around each
definition `[INFERENCE: stated from memory of the Mathlib style guide, not
re-read for this file]`.

## 4. A property goes in a type when the type can carry it.

**Rule.** When a fact holds for every value the program builds, put the fact in
the type. Do not state it as a predicate over free data and prove it preserved
by every step.

```lean
structure PosNat where
  val : Nat
  pos : 0 < val

def PosNat.add (x y : PosNat) : PosNat := ⟨x.val + y.val, by have := x.pos; omega⟩

/-- No lemma `add_pos` is needed: `(x.add y).pos` is the proof. -/
example (x y : PosNat) : 0 < (x.add y).val := (x.add y).pos
```

**In the project.** `AdjudicatorState.self` (`State.lean:92`) follows the rule:
`self = Term.keccak selfInputs` by construction, so "self is an address" needs
no hypothesis. `HoldingWellFormed` (`Lemmas.lean:3997-4004`) breaks the rule. A
holding is free data `((treeNumber, position), commitment)`
(`State.lean:62`). The predicate restates five facts about it and
`holdingWellFormed_step` (`Lemmas.lean:4741`, 169 lines) re-proves them for
every call. Its closure is 26 declarations and 1,866 lines. exophial4's
`drs.bend` went from 4,349 lines and 153 laws to 665 lines when scope became a
type (`~/code/git_puller/repos/exophial4/docs/labbook/a-scoped-drs-type.md`;
665 lines verified on disk; "1 law" after the change is the coordinator's figure
`[INFERENCE]`).

**Limit.** A fact that depends on another state (for example "this note is
live in the pool") cannot live in the holding's own type. It comes from a pool
theorem (rule 1). The holding's type carries only the facts the adjudicator
decides: its asset, its random `r`, and its positive value.

**Statement inputs too.** The rule also applies to a theorem's inputs. When
more than one theorem needs the same fact about an input, the fact is a refined
type (`{c : FixedPart // 2 * ε ≤ c.challengeDuration}`, or a `structure`), not
a `Prop` argument that each theorem repeats. A theorem that takes the raw type
then states, by its type, that it holds for every value. Nitro broke this rule.
Separate repairs added `hχk`, `hχopen`, `hχsent`, the `2·ε` timeout premises
and own-signing premises one theorem at a time, so the same fact was decided
differently in different theorems, and a theorem that omitted a premise still
type-checked (operator, 2026-10-31). The Limit above still applies: "χ is still
open" depends on the chain state, so it is a reachable-state fact, not a fact
about the channel's own type. A prose rule did not hold: the trunk had 1,485
hypotheses and 0 refined-type inputs on 2026-10-31. `docs/lean/type-doctrine.md`
proposes the linters that make the build check this rule.

**Source.** `docs/process.md` Stratum 1 (Minsky 2011, "make illegal states
unrepresentable"). Bird and Paterson 1999 for the scoped case.

## 5. One invariant per fact. State it before the step lemma.

**Rule.** Each fact gets one `def` named `…Inv` or a plain predicate. Its
one-step lemma is `…_step`. Its reachability lemma is `…_reachable`. State the
invariant before any proof about the steps.

```lean
inductive Step : S → S → Prop where
  | incA (s : S) (h : s.a < 10) : Step s { s with a := s.a + 1 }
  | incB (s : S) : Step s { s with b := s.b + 1 }

def SmallA (s : S) : Prop := s.a ≤ 10

theorem inv_step {s s' : S} (hinv : SmallA s) (h : Step s s') : SmallA s' := by
  cases h with
  | incA hlt => exact hlt
  | incB => exact hinv
```

**In the project.** `core_invariant` (`PoolSafety.lean:132`) follows the rule:
`HeldInvariant` is stated once and proved by one induction. `Safety.lean`
breaks it. `unbeatablePaysInv_step` (`Safety.lean:1446`, 826 lines),
`TimeoutInv_appliesAll` (`Safety.lean:3372`, 362 lines),
`finalization_permanent` (`Safety.lean:773`, 150 lines) and
`turn_number_record_monotone` (`Safety.lean:631`, 142 lines) each walk all 11
calls with their own copies of the frame reasoning.

**Source.** `docs/process.md` Stratum 2 (Abrial 2010; Lamport 2002).

## 6. One frame lemma per call. One footprint predicate per field group.

**Rule.** For each call, state which fields it may write. Prove one lemma that
every other field is unchanged. Prove an invariant about a field only for the
calls in that field's footprint.

```lean
inductive Op where
  | incA
  | incB
  deriving DecidableEq

def applyOp : Op → S → S
  | .incA, s => { s with a := s.a + 1 }
  | .incB, s => { s with b := s.b + 1 }

/-- Frame lemma: a call outside the footprint of `b` leaves `b` unchanged. -/
theorem applyOp_b_frame (op : Op) (s : S) (h : op ≠ .incB) : (applyOp op s).b = s.b := by
  cases op with
  | incA => rfl
  | incB => exact absurd rfl h
```

**In the project, checked on the real model.** This footprint predicate and
frame lemma type-check against `68f49f6` in 1.5 s:

```lean
/-- Footprint: which calls may write `pool` or `holdings`. -/
def ArmadaLean.Adjudicator.Call.touchesPool : Call → Bool
  | .deposit .. | .transfer .. | .reclaim .. | .bulkTransferAllAssets ..
  | .bulkConcludeAndTransferAllAssets .. => true
  | _ => false

theorem step_pool_frame (sig : SignatureAssumptions)
    (appRel : AdjudicatorState → FixedPart → List VariablePart → VariablePart → Prop)
    (actor : Actor) (s s' : AdjudicatorState) (call : Call)
    (hstep : Step sig appRel actor s call s') (hcall : call.touchesPool = false) :
    s'.pool = s.pool ∧ s'.holdings = s.holdings := by
  cases call <;> simp only [Call.touchesPool, Bool.true_eq_false] at hcall <;>
    simp only [Adjudicator.Step] at hstep <;> grind
```

The six existing one-field frame lemmas `step_self`, `step_configuredAssets`,
`step_now_frame`, `step_signedBy_frame`, `step_signedBy_mono`,
`poolSignedBy_frame` (`Lemmas.lean:1184-1530`, 293 lines) each take 11
`obtain` cases. Each is proved by this one line on `68f49f6` (checked):

```lean
  cases call <;> simp only [Adjudicator.Step] at hstep <;> grind
```

(`step_self` needs `grind [AdjudicatorState.self]`.)

**Source.** Kassios, "Dynamic Frames", FM 2006. Borgida et al. 1995.

## 7. A general existence lemma states the exact effect.

**Rule.** A lemma that says "the step exists" also says what the result is,
with an equation, not a bound. A can-happen statement then instantiates the
lemma. It does not build its own execution.

```lean
theorem incA_exists (s : S) :
    ∃ s', applyOp .incA s = s' ∧ s'.a = s.a + 1 ∧ s'.b = s.b :=
  ⟨_, rfl, rfl, rfl⟩
```

**In the project.** `transfer_step_exists` (`Safety.lean:2656`, 389 lines)
breaks the rule: it gives only a lower bound on the payout. So
`transfer_after_transfer_can_happen` (`Safety.lean:5921`, 494 lines),
`topup_can_happen` (342), `payout_can_happen` (290) and
`bulk_conclude_and_transfer_can_happen` (273) each build a literal execution
from `genesis` with hand-written witnesses. `sign_step` (`Safety.lean:4067`,
11 lines) and `signNote_step` (13 lines) follow the rule: they state the exact
result and the can-happen proofs reuse them. `finalized_affordable_pays`
(`23e01b5`) also follows it. It replaced `transfer_after_transfer_can_happen`
with an 11-line proof that reuses the invariant `FinalizedPaysInv` and the
exact payout lemma `finalizedPaysInv_pays`.

**Source.** Abelson and Sussman 1996, §2.2.4 (stratified design).
`docs/process.md` Stratum 4, gate.

## 8. Tactics: which one, when.

Each line below was checked on `v4.33.1`.

| Goal shape | Tactic | Checked example |
|---|---|---|
| Linear arithmetic on `Nat`/`Int` | `omega` | `example (a b : Nat) (h : a ≤ b) : a - b = 0 := by omega` |
| Closed decidable proposition | `decide` | `example : (3 : Nat) < 5 := by decide` |
| Rewrite with the `@[simp]` API | `simp` / `simp [defs]` | `example (l : List Nat) : (l ++ []).length = l.length := by simp` |
| Equalities, propositional logic, case splits, linear arithmetic, mixed | `grind` | `example (x y : Nat) (h : x < y) : x + 1 ≤ y := by grind` |
| Split an inductive hypothesis or value | `cases h with` / `rcases` | `rcases o with _ \| n` |
| Pull witnesses out of `∃`/`∧` | `obtain ⟨x, hx, hy⟩ := h` | `obtain ⟨n, _, hpos⟩ := h` |

Order of attempt for a goal after `cases call`: `grind` first. If `grind`
fails, `simp only [the API]` then `grind`. If that fails, `obtain` the exact
conjunct and `exact` it. Never start with `obtain`.

**Measured limit, `68f49f6`.** `cases call <;> simp only [Adjudicator.Step]
at hstep <;> grind [update, ChannelStatus.mode]` on the statement of
`turn_number_record_monotone` (`Safety.lean:631`, 142 lines) closes 4 of the
11 call cases and fails on 7 (the cases whose footprint includes
`channels`; the goal reads `ChannelStatus.mode`, an `if` chain over
`finalizesAt`, through `update`). `grind` closes a frame case by itself. A
footprint case needs the reading lemmas of `update` and `ChannelStatus.mode`
first (rule 3). This is why `reduction-plan.md` puts the `@[simp]` API
before the invariant rewrites.

`grind` is a core tactic in `v4.33.1`. It is not an external SMT solver. It
produces an ordinary Lean proof term. Source: *Lean Language Reference*,
chapter "The `grind` tactic" (lean-lang.org/doc/reference/latest, read
2026-10-26). `grind` is used zero times in the model at `68f49f6`
(`grep -rn grind ArmadaLean` is empty).

`decide` evaluates the proposition. It fails on an open term and on a large
computation. `omega` handles `+`, `-`, `*` by a constant, `≤`, `<`, `=`, `%`
and `/` by a constant. It does not handle a function application it cannot
read.

## 9. Naming and file layout.

**Names.** Types and predicates: `UpperCamelCase` (`PoolState`,
`HoldingWellFormed`). Data: `lowerCamelCase` (`heldValue`, `insertStartPos`).
Theorems: `snake_case`, built from the conclusion (`spend_live_exact`,
`applyRecords_preserves_key`). Reading lemma of definition `f` on projection
`p`: `f_p` (`bump_a`). Frame lemma: `<step>_<field>_frame`. One-step
invariant lemma: `<Inv>_step`. Reachability form: `<fact>_reachable`. Never
disambiguate with a trailing `'`. Source: Lean core `doc/std/naming.md`
(read 2026-10-26).

**Layout.** One file per concept, in dependency order. The current order is
`Channel.lean` → `Outcome.lean` → `Assumptions.lean` → `Notes.lean` →
`State.lean` → `Operations.lean` → `Reachability.lean` → `Lemmas.lean` →
`Safety.lean`. A definition's `@[simp]` API goes in the file that defines it.
A statement goes in `Safety.lean`. A proof helper goes in `Lemmas.lean` or in
the file of the definition it reads. Multi-line declaration headers indent by
four spaces; bodies by two. Source: Lean core `doc/std/style.md` (read
2026-10-26).

**Namespace.** Adjudicator declarations live in `ArmadaLean.Adjudicator`.
Pool declarations live in `ArmadaLean`. `Step` is ambiguous inside the
adjudicator namespace. Write `_root_.Step` for the pool step and
`Adjudicator.Step` when `simp only [...]` needs the adjudicator step (checked:
`simp only [Step]` fails with "Ambiguous term").

## 10. Doc comments cite a source with file and line.

**Rule.** Every `def`, `structure`, `inductive` and statement carries a doc
comment. The comment names what the declaration models and cites the source
with a file path and a line range, or a paper section. The comment states the
fact. It does not argue.

```lean
/-- `TransactModule.transactOwned` (TransactModule.sol:488-519): the same
    bookkeeping shape as `transact`, with no unshield. -/
```

(`Pool.lean:1443-1446`, trimmed.)

A doc comment that exceeds 25 lines is a design record, not a doc comment.
Move it to `docs/` and cite it by path. The module comment of `Notes.lean`
(`Notes.lean:1-110`) and the doc comment of `PoolOwnedSpendStep`
(`Notes.lean:229-277`) exceed this limit.

**Source order (operator ruling, 2026-10-26).** Start with the latest paper.
Then check the code. For Nitro, the latest paper is SATP (2022). The code is
the statechannels team's go-nitro at `64c4ee9a1` (2023-11-14; last contract
change `9d3a669e5`, 2023-10-13). Do not use armada-nitro's vendored copy as
the code source. It is cerc-io's fork at `435eb2b` (2024-10-30), with changes
to `MultiAssetHolder`, `NitroAdjudicator` and `StatusManager`. Where the code
corrects the paper, follow the code and cite both. An example is the challenge
rule: SATP Listing 1 allows a re-challenge at an equal version, and go-nitro
requires a strictly higher turn number. Name sources by title (ForceMove,
Nitro, SATP, go-nitro), not by labels such as S1–S4.

**Source.** `docs/process.md` practice 13 ("Keep the code and the record as
one graph") and the Literate Code Principle in `~/.claude/CLAUDE.md`.

## 11. Hypotheses, axioms and statement texts.

- An axiom may declare an uninterpreted primitive: a verifier, a hash or a
  parser, by its type only (`EdDSAVerify`, `VerifyAccepts`, `decodeGuarantee`).
  It never asserts a property. A property comes from the construction (the
  Dolev–Yao `Term` constructors give collision resistance) or from a named
  hypothesis or structure field (`SignatureAssumptions`,
  `PoolSignatureLiveness`, `Notes.lean:211`). `#print axioms` on every
  statement lists only the declared primitives and `propext`,
  `Classical.choice`, `Quot.sound`. Operator ruling, 2026-10-27;
  `docs/process.md` 2.4.
- Every top-level `theorem` cites its source in its doc comment: a paper
  passage (title and lines), a go-nitro or pool contract line, or an operator
  ruling with its date. A theorem with no such source is not added, whoever
  proposes it, the coordinator included. An argument the coordinator makes for a
  design choice goes in the labbook, not into the theorem list. A fact that a
  proof needs is a lemma in `Lemmas.lean`, not a theorem. Operator ruling,
  2026-10-27.
- A statement text in `Safety.lean` changes only with the operator's approval.
- Reachability is decided when a theorem is typed, never when it is proved (operator ruling 2026-10-29,
  `labbook/2026-10-29-nitro-reachability-ruling.md`). Reachability is a type, not a hypothesis (operator ruling
  2026-10-31, rule 4): a theorem over a system state (a state type of a declared `Machine`;
  `ArmadaLean/Machines.lean` lists them) takes the reachable-state subtype of its machine. A theorem that holds
  for every state takes the raw state type, and its doc comment gives the reason, for example "holds for every
  state; the generality makes it stronger". Until the spec-module check replaces `scripts/check.sh`, the old
  form (a `Reachable` hypothesis, or a row in `scripts/reach-exempt.tsv`) stays valid. A non-vacuity witness
  proves that its state is reachable. The type-approval record gives the state domain of each type, and the
  operator rules on it with the type.
- A reviewer of a type or of a game definition asks three questions:
  - Can the hypotheses all hold in a state that no history produces?
  - Can the honest player make a move that no adversary round can follow? If yes, `UnbeatableIn.move` holds
    vacuously and every goal is provable. `Threat.Round.exists` (`Strategy.lean`) must stay proved (defect of
    2026-10-29, `labbook/2026-10-29-machine-types-for-approval.md`).
  - Does a witness for a theorem about funds use a positive amount? A witness with zero coins shows only that the
    hypotheses hold together, not that the theorem says anything about money (D10, same record).
- A definition is never replaced by `True`, `False` or a constant. That makes a
  statement vacuous. `verification.md` has the check.
- Numeric arithmetic on a hash or an address is forbidden. `self`, commitments
  and roots are `Term`s. `docs/process.md` Stratum 1.2.
