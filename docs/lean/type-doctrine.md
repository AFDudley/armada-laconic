# Type doctrine and linters for armada-lean statements

Status: proposal for operator review, 2026-10-31. Branch `draft/type-doctrine`. Nothing here is in force. The
design it enforces is `docs/lean/typed-library.md` (approved 2026-10-31).

## 1. What failed

`writing-lean.md` §4 said "A property goes in a type when the type can carry it" before the Nitro model was
written. The models that wrote the model did not follow it. A prose rule that no tool checks did not hold.

Measured on armada-lean `integrated` at `afda79b` with a Lean metaprogram that walks every `ArmadaLean`
declaration (`forallMetaTelescope` over each theorem type; structure fields by `getStructureFieldsFlattened`):

| Measure | Count |
|---|---|
| Theorems | 937 |
| Hypotheses (`Prop` binders) | 1,485 |
| Data binders | 4,952 |
| Data binders of a `Subtype` type (after `whnfR`) | 0 |
| Structures | 22 |
| Structures with a proof field | 6 (13 proof fields, 53 data fields) |
| Type classes | 0 |

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
| `looseHypothesis` | A spec theorem has a hypothesis whose only free variable is one data binder, and the theorem's source citation does not state it. | 2 | Single-input checks. |
| `aliasHypothesis` | A hypothesis is `x = f y` where `x` is a binder of the same theorem. | 3 | `hvn`-style duplicates. |
| reachability | Already designed in the spec module (reachable-state types). | 3 | Raw state plus `Reachable` hypothesis. |

`repeatedPremise` is the strongest signal. The premise analysis already computes its clusters
(`armada-lean/.scratch/hyp_analysis.py`, Lean-elaborated binder types, conjunction normalization). The linter
moves that computation into the build.

`premiseVariant` reports only. The operator rules on each variant (as for `TimedChannel`, operator ruling
2026-10-31: `2·ε`).

## 4. Open decisions (operator)

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

   Proposal: the criterion is repetition, not the declaration's role. A precondition that one declaration alone
   needs may be a proof argument ("evidence as argument"). A condition that two or more declarations need is a
   type, in helpers and spec theorems alike (`repeatedPremise`, all declarations). `looseHypothesis` then
   applies to every declaration, with the same single-use exception.
2. **The one exception to rule 2.** The design says "no allow-list". Rule 2 needs one narrow exception: a
   condition the source states for that theorem alone. Proposal: the exception is the theorem's own source
   citation, which the spec module already parses (`spec_lib.py`); the citation sentence must name the
   hypothesis. No list.
3. **Switch-on.** 1,485 existing hypotheses would fail at once. Options:
   - convert first, then turn the linters on (`typed-library.md` §5 step 4). Not a ratchet.
   - a baseline of known violations. A ratchet, which the operator rejected for `check.sh`.

   Proposal: convert first; the linters run in report-only mode until the conversion is complete.
4. **Briefs.** Every statement-authoring brief names this document and `typed-library.md` as required reading,
   and the reviewer checks rules 1-5. Proposal: yes, until the linters are on; after that the build is the check.

## 4a. Layers: build up, do not stay at the bottom (operator, 2026-10-31)

The proofs stayed at the bottom layer. Measured on armada-lean at `afda79b` plus `draft/prove-satp-peers-can-finalize`
(Lean metaprogram: does a theorem's proof term name one of the base operations `Adjudicator.apply`,
`NetworkState.applyAdv`, `Adjudicator.transfer`, `.challenge`, `.checkpoint`, `.conclude`, `.reclaim`):

| Measure | Count |
|---|---|
| Theorems scanned | 937 |
| Proof term names a base operation | 222 |
| Theorems whose doc comment cites a paper passage | 91 |
| Of those, proof term names a base operation | 34 |

So 34 of the 91 paper-level theorems are proved by reasoning about single contract calls, not from a layer of
facts about channels, finalization or redistribution. Other symptoms: `consensus_round_safe_core` (536 lines),
48 declarations over 60 lines, duplicated 81-99 line blocks, `satp_peers_can_finalize` (249 lines),
`turbo_value_upper_bound` (1,812 added lines on its branch). [INFERENCE] The count is an approximation: a proof
that uses only lemmas about an operation does not name it; a `simp` that unfolds it does.

[INFERENCE] Why it happened: the work went top-down from approved statements, and each proof was dispatched as
one independent pebble on its own branch. No unit of work owned a middle layer, so each worker went down to the
base operations, and parallel branches could not share what they built (the `consensus_round_safe_core` twins).
The paper itself is layered: per-channel finalization (Nitro §4.4), then redistribution (§4.3, Turbo §5.2),
then constructions (§4.6, "outcomes first").

Rules:

1. **Each layer has a typed API, and a proof uses only the layer directly below.** Layers for Nitro, in order:
   base operations (`Adjudicator`, `Op`) → per-call exact-effect and frame lemmas (`writing-lean.md` §6) → per-channel
   finalization (finalizable outcomes as typed values) → redistribution (Turbo, guarantees) → constructions
   (ledger, virtual channels) → paper theorems.
2. **A paper theorem's proof term does not name a base operation.** Linter `layerDiscipline` (report-only like
   §3): it lists each spec theorem whose proof term names a base operation. Today: 34.
3. **Dispatch the layer before the theorems that need it.** A worker proves a layer's API as its own unit of work;
   theorem workers then compose. A worker that needs a fact the layer below lacks reports it; it does not prove
   it locally on its branch.

## 5. Order of work

1. Merge the spec module (`draft/spec-module-2`): statement protection and reachable-state types.
2. Merge the refined channel and outcome types (`draft/refined-types`, in progress in the coordinator session).
3. Write the linters in report-only mode; report the counts per linter on the trunk.
4. Convert the remaining clusters, one operator ruling per variant.
5. Turn the linters on as failures.

## 6. Metric

The measurement in §1 is the metric. Record it on each merge: subtype and structure-typed binders up, `Prop`
binders per spec theorem down, `repeatedPremise` reports to 0.
