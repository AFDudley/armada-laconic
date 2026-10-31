# Typed library for armada-lean statements

Status: approved by the operator, 2026-10-31. Refinements compose by `structure … extends …`, not by nested
subtypes (operator question, 2026-10-31). No Lean change yet.

## 1. The defect

Nitro statements check invariants as loose `Prop` arguments on raw types. A theorem that omits an argument
still type-checks. Separate repairs added the same kind of argument to one theorem at a time, so the same
fact is decided differently in different theorems (operator, 2026-10-31). Measured on `integrated`
(text counts in `Liveness.lean` and `Safety.lean`; they include proof lines):

| Pattern | Liveness | Safety | Defect |
|---|---|---|---|
| `(hr : (Nitro.network apps Ts).Reachable n)` with `(hvn : v = n.views)` | 8 | 0 | Two inputs for one value, joined by an equation |
| `IsLedgerChannel apps …`, `.kind = .ledger` | 7 | 7 | Two encodings of "ledger channel" that can disagree |
| `challengeDuration`, `DeadlineAfter` | 98 | 6 | Timeout bounds restated per theorem, with different factors |
| `finalOutcome … = none`, `isFinal = false` | 62 | 2 | "Still open" restated per theorem |
| `MadeNoAlternativeSetupMoves`, `isExternal = true` | 2 | 1 | Own-signing and outcome-shape rules as one-off premises |

`writing-lean.md` §4 already says "a property goes in a type". The guide did not apply that rule to theorem
inputs, and no brief told an author to read the sources below.

## 2. Sources and the rules taken from them

- Alexis King, *Parse, don't validate* (2019), <https://lexi-lambda.github.io/blog/2019/11/05/parse-don-t-validate/>.
  - "Strengthen the argument type" instead of checking the argument in each function.
  - A check that returns no evidence "is extremely easy to forget … the code that needs it would still
    typecheck". A loose `Prop` argument is that kind of check at the statement level.
  - "Use a data structure that makes illegal states unrepresentable."
  - "Push the burden of proof upward as far as possible": do the check once, where the value is made.
  - "Avoid denormalized representations … Strive for a single source of truth." The `v` and `hvn : v =
    n.views` pair breaks this rule.
  - "Use abstract datatypes to make validators look like parsers": give a smart constructor when the fact
    is not structural.
- David Thrane Christiansen, *Functional Programming in Lean*, §5.2 "Subtypes" and §8.7 "Special Types",
  <https://lean-lang.org/functional_programming_in_lean/>.
  - `{x : α // p x}` holds a value and evidence that it satisfies `p`.
  - The checker is a decidable test that returns `Option` of the subtype (`Nat.asFastPos?`).
  - Subtypes "transfer the burden of maintaining these rules to the users". So use them inside a library,
    and give users an API that discharges the invariants.
  - A subtype with one data field has the same run-time form as its carrier, so it costs nothing.
  - `CheckedInput` bundles several refined fields in one `structure`, indexed by the context
    (`thisYear`) it was checked against.
- *Mathematics in Lean*, chapter 7 "Structures",
  <https://leanprover-community.github.io/mathematics_in_lean/C07_Structures.html>.
  - A structure bundles data and its `Prop` constraints.
  - A structure is preferred to a bare subtype for a named concept, because it gets a name, projections,
    dot notation and `@[ext]` lemmas.

## 3. Rules for this project

1. **One decided invariant, one type.** An invariant that more than one statement needs is a refined type.
   - Use a `structure` when the paper or a contract names the concept (`LedgerChannel`).
   - Use a `Subtype` only for a local, unnamed restriction.
   - The predicate is defined once. Each gap ruling of the operator becomes exactly one definition.
2. **Every refined type has an API** (`writing-lean.md` §3):
   - a coercion to its carrier, so existing operations still apply;
   - a decidable smart constructor `ofX? : X → Option T` when the predicate is decidable;
   - `@[simp]` lemmas for the projections.
   Proofs use the API, not `.val`/`.property` by hand.
3. **One source of truth per input.** A statement takes `n : ReachableNetwork apps Ts` and uses `n.views`.
   It never takes `v` and `hvn : v = n.views` together.
4. **A construction carries its evidence, indexed by the state it lives in** (as `CheckedInput thisYear`):
   - `LedgerWithSub.In n` bundles the parts `w` and the proof of `w.Holds` in `n.views`;
   - this replaces the separate `w` and `hw` arguments.
5. **State-dependent facts live in state types.** Some facts depend on the chain or on a history, not on
   the value alone (`writing-lean.md` §4, Limit). Examples: "χ is open", a pending deadline, "p signed no
   alternative setup state". Each such fact is a named predicate on the reachable state, or a field of a
   refined state type. It is never a fresh `∀ p, …` premise in one theorem.
6. **The build enforces rule 1.** A Batteries environment linter (`@[env_linter]`,
   `.lake/packages/batteries/Batteries/Tactic/Lint/Frontend.lean`) named `looseHypothesis`:
   - It reports a spec theorem with a `Prop` binder whose free variables are exactly one data binder.
     That is a one-input check, so the input needs a refined type.
   - The spec check (`draft/spec-module-2`) runs the linter and fails on a report.
   - The linter has no allow-list. The exception route is a refined type.
   - Limit: relational premises over two or more inputs (`a' ≤ a`, `hχopen` over `v` and `w`) pass the
     linter. Rule 5 and review cover them.

## 4. Library layout

`ArmadaLean/Types/`, imported by the statement modules:

| File | Types |
|---|---|
| `Reachable.lean` | `ReachableState M`, `ReachableNetwork apps Ts`, `ReachableViews` (the spec-module work defines the first two; they move here) |
| `Channel.lean` | `LedgerChannel apps` (one encoding of "ledger": `kind` and `appDef` must agree), `TimedChannel ε` (the timeout bound, after the operator rules on the factor) |
| `Outcome.lean` | outcome refinements, for example guarantees with external peers (`hog`) |
| `Construction.lean` | `LedgerWithSub.In n` and the other paper constructions |
| `Honest.lean` | own-signing records: a participant's view together with the fact that it followed the protocol (`MadeNoAlternativeSetupMoves`) |

Each file has a doc comment that cites this document and the source of each predicate (`writing-lean.md` §10).

## 5. Migration

1. **Build the library and the linter.** The reachable types come from the spec-module work, which already
   converts reachability to a type.
2. **Collect the gaps.** The programmatic analysis (`armada-lean/.scratch/hyp-analysis.md`) groups
   same-shaped premises.
   - Each group is a candidate type.
   - Where theorems disagree (for example `2·ε` against `ε`), the operator rules on the variant, one gap at
     a time.
   - A group whose premise the paper or a contract states needs no ruling.
3. **Convert by script, headers only,** as the spec-module conversion does:
   - binder plus premise becomes a binder of the refined type;
   - a proof that used the premise uses the API lemma.
   The statement meaning is unchanged. The `Spec/` diff shows each change.
4. **Turn on the linter** when the conversion is complete. From then on, a loose one-input premise fails
   the build.

## 6. Risks

- Subtypes move proof work to users (*Functional Programming in Lean* §5.2). Mitigation: rule 2's API, with
  `@[simp]` lemmas, so proofs do not unpack the subtype by hand.
- A coercion can hide a definition from `simp`. Mitigation: a `@[simp]` lemma for each coercion.
- The linter does not see relational premises (rule 6, Limit). Rule 5 and review cover them.
- In-flight proof branches rebase onto changed headers. The rolling-cutover script converts a branch in the
  same way (spec-module runbook).
