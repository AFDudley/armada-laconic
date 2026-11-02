# Type doctrine and linters for armada-lean statements

Status: in force from 2026-10-31 (operator: the coordinator in pane w3:p4 integrates the coding best practices,
linting and checks into all armada-lean work). The design it enforces is `docs/lean/typed-library.md`. The linters
of §3 are on the trunk in report-only mode; §4 decision 3 says when they become failures.

## 1. What failed

`writing-lean.md` §4 said "A property goes in a type when the type can carry it" before the Nitro model was
written. The models that wrote the model did not follow it. A prose rule that no tool checks did not hold.

Measured on armada-lean `integrated` at `afda79b` with a Lean metaprogram (`forallMetaTelescope` over each
theorem type; structure fields by `getStructureFieldsFlattened`). The first column counts declarations in the
`ArmadaLean` namespace, which holds the Nitro theorems. The Pool modules (`ArmadaLean/*.lean` outside
`ArmadaLean/Nitro/`) declare their theorems at the root namespace; the second column selects them by module.

| Measure | Nitro (namespace `ArmadaLean`) | Pool (modules outside `Nitro/`) |
|---|---|---|
| Theorems | 937 | 281 |
| Hypotheses (`Prop` binders) | 1,485 | 367 |
| Data binders | 4,952 | 1,181 |
| Data binders of a `Subtype` type (after `whnfR`) | 0 | 0 |
| Structures | 22 (all modules) | |
| Structures with a proof field | 6 (13 proof fields, 53 data fields) | |
| Type classes | 0 | 0 |

Linters must select declarations by module, not by namespace: a namespace filter misses the Pool and
declarations such as `Nitro.network_received_voucher_signers`.

The premise analysis (`armada-lean/.scratch/hyp-analysis.md`, 89 in-scope Nitro theorems) found 363 hypotheses.
Only 25 have a doc-comment sentence that names them together with a source. The same premise was decided in
separate repairs, one theorem at a time (`2 * T.ε ≤ challengeDuration` in 10 theorems; `hχopen`/`hχsent`;
`hχk`; `MadeNoAlternativeSetupMoves`; `hog`; `hR1`).

[INFERENCE] Why the models default to hypotheses: Mathlib states general lemmas with separate hypotheses (for
example `(hn : 0 < n)`), and that style dominates Lean training data. For a mathematical side condition that
style is correct. For a domain invariant ("this channel is a ledger channel", "this timeout is sensible") it is
wrong. The linters below must tell the two cases apart; a plain ban on hypotheses would also flag correct
Mathlib-style lemmas.

Conclusion: the doctrine needs mechanical enforcement in the build, not more prose.

## 2. Doctrine

Each rule is testable by a linter in §3.

1. **An invariant that two or more statements need is a type.** It is defined once, in `ArmadaLean/Types/`, as a
   `structure` (composed with `extends`) or, for a local unnamed restriction, a `Subtype`.
2. **A spec theorem's hypotheses are relational or source-stated.** A hypothesis that constrains one input alone
   is a refined type of that input, unless the theorem's own source states that condition for that theorem.
3. **One source of truth per input.** A statement never takes a value together with an equation that defines it
   from another argument (`v` with `hvn : v = n.views`).
4. **A step returns the refined type.** The invariant is proved once, where the value is made, not again in each
   theorem that uses it.
5. **Every refined type has its API:** a coercion to its carrier, a decidable smart constructor where the
   predicate is decidable, and `@[simp]` projection lemmas (`writing-lean.md` §3).

## 3. Linters

Batteries environment linters (`@[env_linter]`,
`armada-lean/.lake/packages/batteries/Batteries/Tactic/Lint/Frontend.lean`), run by the spec check
(`draft/spec-module-2`, `scripts/spec_check.py`), which fails on any report.

| Linter | Fails when | Rule | Catches |
|---|---|---|---|
| `repeatedPremise` | The same hypothesis shape, after binder names are normalized, occurs in two or more theorems. | 1 | The drift itself: `2·ε` restated 10 times, `hχopen`. |
| `premiseVariant` | One field of one binder type is constrained differently across theorems (`ε` in one, `2·ε` in another). | 1 | Conflicting decisions on one gap. |
| `looseHypothesis` | A spec theorem has a hypothesis whose only input is one data binder, and the theorem's source citation does not state it. A type parameter (`E`) or an instance (`DecidableEq E`) is not an input. | 2 | Single-input checks. |
| `aliasHypothesis` | A hypothesis is `x = f y` where `x` is a binder of the same theorem and `f y` does not contain `x` (`o = o.dedup` is a property, not an alias). | 3 | `hvn`-style duplicates. |
| reachability | Already designed in the spec module (reachable-state types). | 3 | Raw state plus `Reachable` hypothesis. |
| `decisionScope` | An `omega` call whose goal context holds more than N arithmetic hypotheses (`≤`, `<`, `=`, `≠` and their `∧`/`∨`/`¬` over `Nat`/`Int`), or a `decide` call whose goal has a free variable or exceeds a term-size bound, or any `decide +kernel` / `native_decide`. N and the bound come from measuring the trunk. | 4a.5 | Blow-up risk of decision procedures. |
| `linter.flexible` (Mathlib, `Mathlib/Tactic/Linter/FlexibleLinter.lean`) | A rigid tactic (`rw`, `exact`, …) acts on the output of a flexible one (`simp`, `split`, `norm_cast`, …). | 4a.5 | Non-terminal searches. Mathlib's own linter; enabled, not rewritten. |

`decisionScope` reads the elaborated tactic info tree (each `omega`/`decide` node's goal and local context), on
the pattern of Mathlib's flexible linter. It is static and deterministic: it bounds the INPUT of the procedure,
which is what drives its blow-up (`omega` reads every arithmetic hypothesis in scope; `decide` evaluates the whole
computation). The fix it asks for is a smaller scope (`clear`, or the call inside a small `have`) or a layer
lemma, not a hand-written arithmetic proof. Time is still measured, but outside the gate: a periodic profile
(`lake env lean -Dprofiler=true`) lists the most expensive calls, because wall time is too noisy to gate on and
kernel time is not counted in heartbeats.

Measured (TypeLinters, `ArmadaLean/Lint/DecisionScope.lean`, `draft/refined-types` `4188271`): `omega` consumes at
most 13 hypotheses on the trunk (p90 7.2, p50 5); `decide` goals reach size 293 (p90 151). Bounds set to the trunk's
own maxima: `maxOmegaHyps = 13`, `maxDecideSize = 300`; the trunk has 0 findings. Over 137 `omega` calls paired with
profiler time, hypothesis count and time correlate weakly (Pearson r ≈ 0.13): today's calls are all small, and the
slowest (`cover_pay`, 287 ms) consumes 11. The bound therefore stops growth; it does not rank today's cost. It
already found one regression: `NonVacuity.lean:2660` on `draft/refined-types`, `unfold Route.LedgerAfterRound2;
decide`, size 21,601 (74 times the trunk maximum). Mathlib's `linter.flexible`: 157 findings on the trunk.

`repeatedPremise` is the strongest signal. The premise analysis already computes its clusters
(`armada-lean/.scratch/hyp_analysis.py`, Lean-elaborated binder types, conjunction normalization). The linter
moves that computation into the build.

`repeatedPremise` excludes three kinds of premise from every count (`ArmadaLean/Lint/TypeDoctrine.lean`, trunk
`b78c293`):

1. **A structure projection.** The `structure` command generates one projection per field. The field is not a
   premise that an author repeats.
2. **A carrier difference.** The key strips `ReachableState.state e` to `e` before it numbers the bound
   variables. Otherwise the carrier's implicit arguments give one premise two keys.
3. **A premise that already is a type, or a parameter.** The linter skips a premise whose head is an
   `ArmadaLean` structure (tested by the declaring module, so the Pool's root-namespace structures count too), a
   `def` that unfolds to a structure or to `Subtype`, or a local variable. Rule 1 makes such types (`SatpEnds`), so
   a declaration that takes one is the intended result. Rule 1 cannot apply to an abstraction's own parameter
   (`Inv net`).
4. **The result of one base-operation call** (`s.apply apps caller op = some s'`, `markSpent spent ns = some
   spent'`). It is the input of a per-call lemma (§4a rule 1, `writing-lean.md` §6). An equation `x = op …` with
   `x` a binder stays an alias.
5. **A transition relation:** an `ArmadaLean` inductive or field whose type ends in `S → S → Prop` (`Step`,
   `Threat.RoundSeq`, `OtherSteps`, `ReachableFrom`). A step lemma takes the step (`writing-lean.md` §5).

The output is deterministic: the linter sorts the rows by a total order on the shape, then groups adjacent rows.
Measured on `0f208d6`: 257 groups before rules 1-3, 250 after.

Measure corrections, 2026-11-02 (armada-lean `draft/cm-opus`, failure tests in
`ArmadaLean/Lint/TypeDoctrineTest.lean`, which `scripts/gate_extra.py` builds before it runs the linter): exclusions 4
and 5 and the module test of exclusion 3 took `repeatedPremise` from 247 to 224 groups and `premiseVariant` from 10 to
9; the input rule of `looseHypothesis` took it from 122 to 228 findings (the code did not change; the old count hid
106 findings, largest groups `s.IsDAG` 14, `OnlyTransfers ops` 9, `b.KeysNodup` 9); the property rule of
`aliasHypothesis` took it from 57 to 54.

`premiseVariant` reports only. The operator rules on each variant (as for `TimedChannel`, operator ruling
2026-10-31: `2·ε`).

## 4. Decisions

1. **Scope: every declaration, helpers included.** An earlier draft exempted helper lemmas. *Functional
   Programming in Lean* does not support that exemption:
   - §5.2 "Subtypes": subtypes belong *inside* a library — "it's a good idea to use them internally to a
     library, providing an API to users that automatically ensures that all invariants are satisfied, with any
     necessary proofs being internal to the library." Helper lemmas are the library's internals.
   - §7.5 "Pitfalls … Pros and Cons": "In idiomatic Lean code, indexed datatypes are not used very often.
     Instead, subtypes and explicit propositions are typically used to enforce important invariants." The book's
     line is between indexed families and subtypes/propositions, not between helpers and theorems.
   - "Interlude: Evidence as Arguments": a proof argument is the idiomatic way to state one operation's own
     precondition (`def third (xs : List α) (ok : xs.length > 2)`).

   Decided (operator, 2026-10-31): the criterion is repetition, not the declaration's role. A precondition that
   one declaration alone needs may be a proof argument ("evidence as argument"). A condition that two or more
   declarations need is a type, in helpers and spec theorems alike (`repeatedPremise`, all declarations).
   `looseHypothesis` then applies to every declaration, with the same single-use exception.
2. **The one exception to rule 2.** Decided (coordinator, 2026-10-31; technical, under the operator's order
   above): a condition the source states for that theorem alone. The exception is the theorem's own source
   citation, which the spec module parses (`spec_lib.py`); the citation sentence must name the hypothesis. No list.
3. **Switch-on.** Decided (coordinator, 2026-10-31): convert first, then turn the linters into failures, one
   linter at a time, when its count on the trunk reaches 0. Superseded 2026-11-01: the gate container has no git
   history, so the gate compares every count with the committed `scripts/gate-baseline.json` (armada-lean
   `AGENTS.md`, "The gate baseline"). A merge must not raise a count. A branch that lowers a count commits the lower
   baseline. A commit that only corrects a linter's measure may raise a count; its message states the defect and
   the counts before and after (coordinator decision, 2026-11-02).
4. **Briefs.** Decided (coordinator, 2026-10-31): every brief that writes Lean names this document and
   `typed-library.md`, and carries the standing rules: no statement change without an operator ruling; every
   Lean command through `/home/rix/bin/lean-run`; no copied `.lake/build`; layer rule (§4a 1-3); search rule
   (§4a 5); `omega` on a small context. The reviewer checks rules 1-5 and the linter counts.

## 4a. Layers: build up, do not stay at the bottom (operator, 2026-10-31; applies to Nitro and the Pool)

The proofs stayed at the bottom layer. Measured on armada-lean at `afda79b` plus `draft/prove-satp-peers-can-finalize`
(Lean metaprogram: does a theorem's proof term name a base operation). Base operations:
- Nitro: `Adjudicator.apply`, `NetworkState.applyAdv`, `Adjudicator.transfer`, `.challenge`, `.checkpoint`,
  `.conclude`, `.reclaim`.
- Pool: `removeAll`, `markSpent`, `applyRecords`, `recordPreimages`, `insertLeaves`, `insertTree`, `insertStart`,
  `withTreePositionsFrom`, and case analysis on `Step` (`Step.casesOn`).

A spec-level theorem is one whose doc comment cites its source: a paper passage for Nitro; a contract or circuit
line, or an operator ruling, for the Pool.

| Measure | Nitro | Pool |
|---|---|---|
| Theorems scanned | 937 | 281 |
| Proof term names a base operation | 222 | 71 |
| Spec-level theorems | 91 | 44 |
| Of those, proof term names a base operation | 34 | 12 |

So 34 of 91 Nitro and 12 of 44 Pool spec-level theorems are proved by reasoning about single operations, not
from a layer of facts about channels, finalization, redistribution, notes or pool steps. Other symptoms:
- Nitro: `consensus_round_safe_core` (536 lines), 48 declarations over 60 lines, duplicated 81-99 line blocks,
  `satp_peers_can_finalize` (249 lines), `turbo_value_upper_bound` (1,812 added lines on its branch).
- Pool: the 12 descents trace mainly to `markSpent_succeeds` called inline instead of through a per-call wrapper
  (8 of 12; audit `armada-lean/.scratch/audit-wt/.scratch/audit/pool.md` §4). An earlier draft cited
  `HoldingWellFormed` here; that predicate belongs to the retired `adjudicator-model-v4` (commits `ce6f4ad`,
  `3feb0b0`, on no current branch), not to the trunk Pool model.

The counts are upper bounds (AuditNitroSafety, 2026-10-31). A proof that correctly uses a per-call lemma, for
example `obtain ⟨…⟩ := Adjudicator.transfer_some h`, still names `Adjudicator.transfer` in its proof term,
because the lemma's instantiated type mentions it. The audit built its violation lists by reading the proofs.
`layerDiscipline` must instead report a proof that UNFOLDS a base operation: its proof term names the
operation's equation or match lemmas (`<op>.eq_1`, `<op>.eq_def`, `<op>.match_<n>`) or its `casesOn`.

[INFERENCE] Why it happened: the work went top-down from approved statements, and each proof was dispatched as
one independent pebble on its own branch. No unit of work owned a middle layer, so each worker went down to the
base operations, and parallel branches could not share what they built (the `consensus_round_safe_core` twins).
The Nitro paper itself is layered: per-channel finalization (Nitro §4.4), then redistribution (§4.3, Turbo
§5.2), then constructions (§4.6, "outcomes first"). The Pool contracts are layered too: tree insertion and
nullifier marking are internal functions that every public operation calls (`TransactModule.sol`
`_accumulateAndNullify`, `MerkleModule` `insertLeaves`).

Rules:

1. **Each layer has a typed API, and a proof uses only the layer directly below.**
   - Nitro, in order: base operations (`Adjudicator`, `Op`) → per-call exact-effect and frame lemmas
     (`writing-lean.md` §6) → per-channel finalization (finalizable outcomes as typed values) → redistribution
     (Turbo, guarantees) → constructions (ledger, virtual channels) → paper theorems.
   - Pool, in order: base operations (tree insertion, nullifier marking, preimage records) → one exact-effect
     and one frame lemma per `Step` constructor → state invariants as types (a holding carries its
     well-formedness, `writing-lean.md` §4) → pool properties (value conservation, no double-spend,
     `core_invariant`) → can-happen and not-blocked theorems.
2. **A spec-level theorem's proof term does not name a base operation.** Linter `layerDiscipline`: it lists each
   spec-level theorem whose proof names a base operation in a value position (not a binder type, not an implicit
   argument, not a matcher motive, not an argument Lean copies from the goal for `Eq.refl` or `congrArg`), or
   uses one of its unfolding auxiliaries (`<op>.eq_<n>`, `<op>.eq_def`, `<op>.match_<n>`, `<op>.induct`). A name
   that the statement itself contains is exempt, and so are, one level deep, the constructor field types, `casesOn`
   and `rec` of an `ArmadaLean` inductive that the statement names: a statement about `Step s s'` is a step lemma.
   Counts: 34 (Nitro) and 12 (Pool) at `afda79b`; 35 and 12 when the gate started; 1 and 0 after the measure
   correction and the fixes of 2026-11-02 (armada-lean `draft/cm-opus`). The one Nitro hit,
   `satp_peers_can_finalize`, splits on `apply` with `cases`; the same split occurs at 87 helper sites, and one
   codemod of those sites is the fix.
3. **Dispatch the layer before the theorems that need it.** A worker proves a layer's API as its own unit of work;
   theorem workers then compose. A worker that needs a fact the layer below lacks reports it; it does not prove
   it locally on its branch.
4. **Seal a layer once its API exists.** Mathlib's style guide, "Transparency and API design": "Central to Lean
   being a practically performant proof assistant is avoiding checking of definitional equality for very large
   terms … irreducible definitions are never unfolded unless the user explicitly requests it (e.g using the
   unfold tactic, or by using the unseal command)"; the cost it names is the need to "recycle API lemmas"
   (<https://leanprover-community.github.io/contribute/style.html>). Here that cost is the point: once a
   layer's exact-effect and frame lemmas exist (Waves 1-2 of `audit-2026-10-31.md`), mark its base definitions
   `@[irreducible]` (or `seal` them after their API). A higher-layer proof that unfolds them then fails to
   elaborate, so Lean enforces rule 1 itself, and `layerDiscipline` becomes a check that nothing unseals them.
   A non-vacuity witness that needs computation (`decide`, `rfl`) uses an explicit, local `unseal` in its own
   section; the spec check lists every `unseal`.
5. **Record every searched proof in the source (operator, 2026-10-31: software practice, not mathematics
   practice).** A search tactic depends on the library state; its result is pinned in the source, as a lockfile
   pins dependencies.
   - Search tactics (`simp`, `simp [..]`, `grind`, `exact?`, `apply?`, `rw?`, `aesop`) are written with `says`
     (`grind?` prints a `grind only [...]` form; checked on the pinned toolchain)
     (`Mathlib/Tactic/Says.lean`): `simp? [h] says simp only [h, List.append_nil]`. The build runs only the
     recorded proof. With `set_option says.verify true` (set in CI and on a toolchain or Mathlib upgrade) Lean
     reruns the search and fails if the result changed.
   - Decision procedures (`omega`, `decide`) need no recording: their result depends only on the goal, not on
     the library's lemma set. Their COST is not small by default, and it is paid on every build:
     - `omega` reads every arithmetic hypothesis in the context, used or not. Measured with
       `#count_heartbeats in` (`Mathlib/Util/CountHeartbeats.lean`): 54 heartbeats for a 3-hypothesis goal, 161
       for the same goal with 12 hypotheses, 5 of them irrelevant. In long proofs, `clear` the unused facts or
       call `omega` in a small `have`. Its worst case is exponential.
     - `decide` evaluates the decision by reduction, so its cost is the cost of the computation. Measured:
       `(List.range 200).sum = 19900` by `decide` fails with "maximum recursion depth"; `decide +kernel`
       proves it. Kernel time is not counted in heartbeats, so `#count_heartbeats` under-reports it; time the
       build instead. `decide` also needs the definitions it computes through to be unsealed (rule 4).
     - `native_decide` is forbidden: it adds the axiom `Lean.ofReduceBool` and trusts the compiler.
     - A decision procedure call that costs more than a set budget is replaced by a short explicit proof. The
       budget is set from a measurement of the trunk's heaviest calls.
   - Term proofs from `show_term` are not recorded: an elaborated term depends on elaboration details and is
     not reviewable.
   Mathlib's guide does not freeze terminal `simp` calls, because a frozen list breaks when a lemma is renamed
   (style guide, "Squeezing simp calls"). Here the toolchain and Mathlib are pinned (`lake-manifest.json`), so
   renames arrive only with a deliberate upgrade, and `says.verify` shows each break at that point. Also: one
   normal form per concept with `@[simp]` lemmas that rewrite to it (style guide, "Normal forms"); measure
   heavy proofs before and after a change (style guide, "Profiling for performance").
   Trunk today: 0 `says`; 373 bare `simp`/`simp [..]` lines (about 31 non-terminal); 638 `simp only`; 83 `grind`
   (Nitro `Lemmas` 31, `Liveness` 34, `Safety` 18); 187 `omega` in 12 modules (`PoolSafety` 52, `Nitro/Lemmas` 39,
   `Liveness` 29, `Value` 18, `Pool` 17, others 32); 281 `decide`.
   Profile (`lake env lean -Dprofiler=true -Dprofiler.threshold=50`, audit worktree at `522e16d`): tactic time
   per file is `Liveness` 34.6 s, `Lemmas` 10.2 s, `PoolSafety` 0.96 s, `Value` 0.76 s, `Pool` 0.51 s. `omega`
   calls over 50 ms: 10 in `Lemmas` (1.70 s together), 1 in `PoolSafety`, none at top level in `Liveness`. The
   profiler charges a nested tactic to the enclosing one (`Liveness`: `refine` 12.0 s over 18 calls), so the
   cost inside those blocks is not attributed yet.
   Recorded (`draft/record-searches`, armada-lean `0c9b07e`): 1,192 of 1,269 search calls recorded with `says`,
   77 left as written (32 combinator sites with a different result per goal; 45 that did not reproduce);
   `scripts/verify_says.sh` passes on the whole tree and fails on a deliberately wrong recording. Effect, same
   machine, two runs each: `simp` time per file falls by about 40% (`Liveness` 2.25-2.31 s → 1.29-1.44 s,
   `Safety` 0.59-0.71 s → 0.30-0.37 s); total tactic time and a clean build (45.5-49.8 s → 41.4-47.7 s) change
   within run-to-run noise. Search was not the main cost; the long proofs are (§4a).
   Pitfall found: Mathlib's `says` re-runs every recorded search when the environment variable `CI` is set, and
   agent shells set `CI=true`. The first measurement was therefore slower than the trunk. Fixed twice:
   `leanOptions := #[⟨`weak.says.no_verify_in_CI, true⟩]` in `lakefile.lean` (builds), and `unset CI` in
   `/home/rix/bin/lean-run` (direct `lake env lean` runs).

## 5. Order of work

1. Merge the spec module (`draft/spec-module-2`): statement protection and reachable-state types.
2. Merge the refined channel and outcome types (`draft/refined-types`, in progress in the coordinator session).
3. Write the linters in report-only mode; report the counts per linter on the trunk.
4. Convert the remaining clusters, one operator ruling per variant.
5. Turn the linters on as failures.

## 6. Metric

The measurement in §1 is the metric. Record it on each merge: subtype and structure-typed binders up, `Prop`
binders per spec theorem down, `repeatedPremise` reports to 0.
