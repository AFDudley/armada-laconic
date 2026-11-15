# lean_def_to_structure

A deterministic codemod for armada-lean. It converts a predicate
`def P (binders) : Prop := body` to the one-field structure
`structure P (binders) : Prop where holds : body`, adds the lemma
`P.iff : P args ↔ body`, and repairs the proofs that break. The same input
gives the same output. The tool lives here, not in armada-lean, because
armada-lean goes upstream.

Why: the `repeatedPremise` linter (`ArmadaLean/Lint/TypeDoctrine.lean`)
excludes a hypothesis once its head is an `ArmadaLean`-declared structure.
Workers made this conversion by hand many times. The brief is
`.scratch/briefs/2026-11-14-codemod-def-to-structure.md`.

## Files

- `discover.lean`: the Lean half. `lake env lean discover.lean` in the
  target checkout, with `CODEMOD_NAMES=A,B` in the environment. For each
  name it prints one JSON line: the full name, the module path, the
  declaration range (`findDeclarationRanges?`), and the byte offsets of the
  `def` pieces from the parsed syntax of the declaration (`def` keyword,
  `declId`, each binder with its kind and names, the type, `:=`, the body).
  It also reports the head constant of the body and whether that head
  unfolds to a structure.
- `probe/`: a small Lake package with one executable, `probe`. It links
  Lean and elaborates one file of the target checkout the way `lean` does
  (run under `lake env` of the target, so that `LEAN_PATH` is set). It
  prints every message with its range, like `lean --json`, plus one line
  per numeric projection `t.1` that resolves to a converted predicate's
  `holds` field (read from the info trees). The driver builds it once;
  `probe/lean-toolchain` must equal the target's `lean-toolchain`.
- `codemod.py`: the Python half. The driver. See its docstring for the
  steps.
- `repairs.py`: the fixed repair patterns. `match(message, source,
  predicates)` returns one `Repair` or `None`.
- `test_repairs.py`: the matcher against recorded Lean messages, both the
  matching and the refusing cases. `python3 test_repairs.py`.
- `scope.lean`: the scope measurement. It lists the `repeatedPremise`
  groups whose head is an `ArmadaLean` `def … : Prop` that is not a
  structure.

## Usage

```
python3 tools/lean_def_to_structure/codemod.py \
  --checkout /path/to/armada-lean-worktree --slot 2 \
  --names RegistryConsistent,Adjudicator.IsDAG \
  --report report.json
```

The checkout is a throwaway worktree, never a live branch. The tool runs
every Lean command as `flock ~/.cache/lean-run/locks/slot-N.lock nice
<command>` with `~/.elan/bin` on `PATH` and `CI` unset. Logs go to
`<checkout>/.scratch/codemod/`.

## Repair patterns

Each pattern is a fixed edit at the position of one Lean error. An error
that matches no pattern stops the tool. The tool then writes the report
with the file, the line and the message, and it changes nothing more.
No pattern inserts `sorry`, `admit`, `axiom` or `native_decide`.

| pattern        | Lean message                                                         | edit                                     |
|----------------|----------------------------------------------------------------------|------------------------------------------|
| `holds-apply`  | `Function expected at h … has type P args`                           | `h.holds x`                              |
| `holds-field`  | `Invalid field f: … does not contain P.f` / `Invalid projection …`   | `h.holds.f`, `h.holds.2.1`               |
| `holds-type`   | type mismatch, the term has type `P args`                            | `term.holds`                             |
| `holds-proj`   | the probe finds `h.1` that resolves to `P.holds` (`h.1` meant the first part of the body; Lean reports the mismatch far away, or never) | `h.holds.1` |
| `wrap`         | type mismatch, the expected type is `P args`                         | `⟨term⟩`                                 |
| `wrap-nested`  | argument mismatch inside `{ holds := … }`, the body is a structure   | `⟨term⟩`                                 |
| `constructor`  | `Tactic t failed … ⊢ P args`                                         | `constructor` before `t`                 |
| `unfold-iff`   | `Tactic unfold failed to unfold P`                                   | `rw [P.iff]`, then `unfold` of the rest  |
| `simp-iff`     | `simp made no progress`, the simp set lists `P`                      | `P.iff` in place of `P`                  |

## Loop

1. `lake build`, then `lake build ArmadaLean.Lint.TypeDoctrine
   ArmadaLean.Lint.TypeDoctrineTest` (the modules the gate builds).
2. For each failing module: the probe gives every message with its range
   and every `t.1` on a converted type. Each `t.1` gets `.holds`; the
   matcher maps each error to one repair. The repairs of one round apply
   together, last position first, without overlap. The file runs again
   until it has no error and no such projection.
3. Back to step 1 until the build passes.
4. `docs/counterexamples/*.lean` are outside the build. Each runs through
   `lake env lean --json` and the same repair loop.

A round that applies no repair stops the tool. A repair that would repeat
an earlier repair (same pattern, same detail, same line text) stops the
tool.

## Guards

- A predicate under `ArmadaLean/Spec/` is refused. The repair loop never
  edits a file under `ArmadaLean/Spec/`.
- `git grep "Spec\.<name>\b" -- ArmadaLean/Spec` must find nothing for
  each predicate.
- A predicate that is already a `holds` structure with `P.iff` is
  skipped; the repair loop still runs.
- `--resume` continues a stopped run after a hand edit of the file that
  stopped it: it skips the base build and the discovery, and reads the
  first run's `lint-before.log` and `discover.log` from
  `<checkout>/.scratch/codemod/`.

## Output

A commit-ready working tree in the checkout, with
`scripts/gate-baseline.json` file hashes refreshed
(`scripts/update_gate_baseline.py --hashes-only`), and a JSON report:
the predicates, the repairs (file, line, pattern, predicate, detail), the
linter summaries before and after (`scripts/type_lint.lean` after full
builds), the finding lines removed and added, the `bash scripts/check.sh
--json` `_ok` fields, the files changed, and `stopped` with the message
when the tool stopped.
