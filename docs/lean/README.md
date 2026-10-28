# Lean in this project

These files say how Lean is written for `AFDudley/armada-lean`, the
specification of the Armada contract rewrite. They apply `docs/process.md`
to Lean 4. They do not repeat it. The rules hold for every writer of Lean
here: a person, a large model, or a small model. Only the execution files
are written for one executor.

Two layers:

1. **General rules.** `writing-lean.md`, `anti-patterns.md`,
   `verification.md`, `worker-briefs.md`. Each rule is a Lean or
   proof-engineering rule with its source. The project's code is the example
   and the evidence, not the scope.
2. **Execution.** `proof-dispatch-plan.md`: the current proof work (pool and
   Nitro, dispatched with exophial on Sonnet). `reduction-plan.md`: the
   proof-size reduction of the old `adjudicator-model-v4` branch.

Mathlib: the specification branches (`nitro-proofs`, `pool-symbolic`, `armada-nitro-spec`) require Mathlib
`v4.33.1` on toolchain `leanprover/lean4:v4.33.1`. Use Mathlib where it makes sense (operator ruling, 2026-10-27).

Examples: most Lean examples in these files come from branch `adjudicator-model-v4` at `68f49f6` and branch
`pool-model` at `7f2d1ec`, which had no Mathlib. They were type-checked on 2026-10-26 unless marked `[INFERENCE]`.
`adjudicator-model-v4` is now reference only. The rules hold for the current branches.

## Reading order

1. `docs/process.md` (the practice these files apply).
2. `writing-lean.md` (the rules, with examples).
3. `anti-patterns.md` (what the rules prevent, measured).
4. `verification.md` (the checks, as commands).
5. `worker-briefs.md` (how to hand work to someone else).
6. `proof-dispatch-plan.md` (the current proof work).
7. `reduction-plan.md` (the old branch's steps).

## Which file for which task

| Task | Read |
|---|---|
| Write or change a definition, a structure, a step relation | `writing-lean.md` rules 1, 2, 4, 9, 10 |
| Write a proof | `writing-lean.md` rules 3, 5, 6, 7, 8 |
| Decide whether a fact is a hypothesis, a type, or an invariant | `writing-lean.md` rule 4 and 5; `docs/process.md` "Hypothesis or construction?" |
| Check a change before commit or report | `verification.md` section 1 (the script), then sections 2 to 7 |
| Type-check one snippet outside the build | `verification.md` section 8 |
| Write a brief for another person or model | `worker-briefs.md` (template at the end) |
| Execute a step of the reduction | `reduction-plan.md` (the step), then `verification.md` section 1 |
| Review a step | `verification.md`; the step's acceptance in `reduction-plan.md`; `anti-patterns.md` for what to look for |
| Find why a proof is long | `anti-patterns.md` A to F |
| Request a statement change | `docs/process.md` 8.3; `verification.md` section 4 |

## Vocabulary

- **Statement**: a theorem in `ArmadaLean/Adjudicator/Safety.lean` that is
  part of the contract. Its text changes only with the operator's approval.
- **Helper**: any other declaration. It may change freely if the statements
  still build.
- **Footprint**: the set of state fields a call may write.
- **Frame lemma**: a lemma that a call leaves a field unchanged.
- **Exact-effect lemma**: a lemma that a step exists and states its result
  as an equation.
- **Baseline**: the commit a check compares against (`68f49f6` for the
  reduction).
- **Measurement step**: a step whose product is a report of numbers, which
  the next step uses.
