# lean_def_to_structure

One Lean program that converts a predicate `def P (binders) : Prop := body`
of an armada-lean checkout to the one-field structure
`structure P (binders) : Prop where holds : body`, adds the lemma
`P.iff : P args ↔ body`, and repairs the proofs that break. The same input
gives the same output. The tool lives here, not in armada-lean, because
armada-lean goes upstream.

Why: the `repeatedPremise` linter (`ArmadaLean/Lint/TypeDoctrine.lean`)
excludes a hypothesis once its head is an `ArmadaLean`-declared structure.
Workers made this conversion by hand many times. The brief is
`.scratch/briefs/2026-11-14-codemod-def-to-structure.md`.

## Build and run

```
cd tools/lean_def_to_structure && lake build
.lake/build/bin/def_to_structure --checkout /path/to/armada-lean-worktree \
  --slot 2 --names RegistryConsistent,Adjudicator.IsDAG --report report.json
```

`lean-toolchain` here must equal the target's. The checkout is a throwaway
worktree, never a live branch. Every `lake` command runs as
`flock ~/.cache/lean-run/locks/slot-N.lock nice <command>` with
`~/.elan/bin` on `PATH` and `CI` unset. Logs, `state.json` and the LSP log
go to `<checkout>/.scratch/codemod/`.

Tests: `lake env <checkout> .lake/build/bin/tests <checkout>` (the snippets
parse with `Init` and `Mathlib.Tactic.Says` from the checkout).

## Files

- `Main.lean`: the driver. Steps: the base build and `scripts/type_lint.lean`
  (the "before" list); `importModules` of the target (discovery, and the
  parser for the loop); discovery and the guards; the conversion; the
  dependents in import order through `lake serve`; one full check at the
  end.
- `DefToStructure/Lsp.lean`: the `lake serve` client. JSON-RPC with
  `Lean.JsonRpc`; `didOpen`/`didChange` with the whole text; `waitIdle`
  returns the diagnostics once `$/lean/fileProgress` reports nothing left for
  the current version. Queries: `textDocument/definition`, and
  `$/lean/rpc/call` with `Lean.Widget.getInteractiveTermGoal`,
  `Lean.Widget.getInteractiveGoals`, `Lean.Widget.getInteractiveDiagnostics`
  and `Lean.Widget.getGoToLocation`.
- `DefToStructure/Parse.lean`: the target file as command `Syntax`. The
  parser runs with the imports' environment; only the commands that change
  later parsing elaborate (`open`, `namespace`, `section`, `end`,
  `set_option`, `universe`, `macro`, `syntax`, `notation`, …). Declarations
  are the server's job.
- `DefToStructure/Patterns.lean`: the decision. One diagnostic in, one repair
  or none out.
- `DefToStructure/Discovery.lean`: `findDeclarationRanges?`, the `def`
  pieces from the parsed declaration, the conversion text, the import graph
  from file headers.
- `Tests.lean`: 54 cases of the decision (the 43 recorded cases of the first
  tool, the silent projection, the parenthesized-receiver goal, the two
  applications under `unfold`, the negation cases and the `Decidable` instances).
- `scope.lean`: scope measurement, `lake env lean scope.lean` in a built
  checkout.

## Elaboration data

A predicate is recognized by **location**: the position of its name in the
`structure` declaration the tool wrote, or the position of its `holds`
field. Every fact below is a declaration location that the server returns.

| fact | source |
|------|--------|
| the syntax at the diagnostic range | the parsed file (`Parse.stackAt`) |
| the expected type's head of a term | `getInteractiveTermGoal` at the term's first token, then `getGoToLocation` on the head tag of its `type` (the tag with the longest all-`appFn` `subexprPos`) |
| the goal's head before a tactic | `getInteractiveGoals` at the tactic's start, the same way |
| what an identifier or field index resolves to | `textDocument/definition` |
| the type's head of the term a diagnostic is about | `getInteractiveDiagnostics`, then `getGoToLocation` on the head tag of the **second expression embedded in the diagnostic** |
| the expected type's head of a projection term `h.1` | the third embedded expression, the same way |
| the head under `¬` | the same queries, the head tag under the first argument (`/1/0/…`) |

The last two rows are the one place where the diagnostic itself is a source.
They read the `Expr` objects Lean attached to the message (through their RPC
references), not the message text. No position query gives those two types:
the server's info lookup (`InfoTree.hoverableInfoAtM?`) prefers the
innermost info at a position, and every position of a projection node or an
application node belongs to one of its children (`h` and `1` in `h.1`); an
intermediate of a dotted identifier (`hord.isDAG` in `hord.isDAG.irrefl`)
has no info at all (`Lean/Elab/App.lean`, `elabAppLValsAux` elaborates it
with `expectedType? := none`). Lean puts `e` first and `eType` second in
`throwTypeMismatchError`, `throwFunctionExpected`, the invalid-field and the
invalid-projection errors, and the expected type third in the mismatch
errors.

## Repair patterns

Each pattern is a fixed edit. An error that matches no pattern stops the
tool; the report names the file, the line and the message, and nothing more
changes. No pattern inserts `sorry`, `admit`, `axiom` or `native_decide`.

| pattern | decision | edit |
|---------|----------|------|
| `holds-apply` | the term at the range is an application with no term info (it did not elaborate), and the diagnosed term's type is `P …` | `f.holds args` |
| `holds-field` | the range sits in a projection chain; the first field that resolves to nothing is the failing step; the diagnosed term's type is `P …` | `h.holds.f`, `h.holds.2.1` |
| `holds-type` | a term whose type is `P …` where the expected type is not | `t.holds`, `(t).holds` for an application |
| `wrap` | a term whose expected type is `P …` and whose type is not; not already `⟨…⟩` | `⟨t⟩` |
| `wrap-nested` | the argument of `⟨…⟩ : P` whose expected type is `P`'s body, itself a structure | `⟨t⟩` |
| `constructor` | a tactic from a fixed list failed and the goal before it is `P …` | `constructor` before the tactic |
| `unfold-iff` | the range is a name in `unfold` that resolves to `P` | `simp only [P.iff]` (every application, as `unfold` did; `rw` rewrites one instantiation only), then `unfold` of the rest |
| `simp-iff` | the range is a `simp` family call (or `says`) with a name that resolves to `P` | `P.iff` in place of `P` |
| `decidable-iff` | the stack holds an `instance` whose declared type is `Decidable (X)` with the head of `X` resolving to `P`, and whose value is an `inferInstanceAs` application | `decidable_of_iff _ (P.iff ..).symm` |
| `not-wrap` / `not-holds` | a term of type `¬ P args` where `¬ body` is expected, or the converse: the heads under `Not` on both sides (the first argument's head, tag `/1/0/…` of the explicit form) | `fun h => t ⟨h⟩` / `fun h => t h.holds`, parenthesized in an argument position |
| `holds-proj` | every `t.k` in the file whose index resolves to `P.holds` (the silent case: `t.k` was part `k` of the body, now it is the whole body) | `t.holds.k` |

## Loop

For each file that imports a converted module, in import order: `didOpen`;
until the file has no error and no silent projection: read the diagnostics,
decide one repair per error (last position first, no overlapping edits),
apply, write the file, `didChange`. A round with an error and no repair
stops the tool. A repair that would repeat (same pattern, same detail, same
line text) stops the tool. An edited module is rebuilt with
`lake build <Module>` before the next file opens, so that a dependent's
imports are the repaired `.olean` files. No full `lake build` runs inside the
loop.

At the end: `lake build`, the linter modules, `scripts/type_lint.lean` (the
"after" list), `scripts/update_gate_baseline.py --hashes-only`,
`bash scripts/check.sh --json` (it also elaborates
`docs/counterexamples/*.lean`), and the JSON report.

## Guards

- A predicate under `ArmadaLean/Spec/` is refused. The loop never writes a
  file under `ArmadaLean/Spec/`.
- `git grep "Spec\.<name>\b" -- ArmadaLean/Spec` must find nothing.
- `--resume` continues a stopped run after a hand edit of the file that
  stopped it: it skips the base build, the discovery and the conversion, and
  reads `state.json` and `lint-before.jsonl`.

## Report

The predicates, the repairs (file, line, pattern, predicate, detail), the
rounds, the linter summaries before and after, the finding lines removed and
added (the `example` field dropped: it is a nondeterministic
representative), the `check.sh --json` `_ok` fields, the files changed, the
timings of every `lake` command, `stopped` and `stops`.
