# How to write a brief for a Lean step

A brief is a unit of Lean work handed to someone else: a person, a large
model, or a small model. The rules below hold for every reader. A brief
must be executable with no judgment call. It names the file, the
declaration, the command and the acceptance script. A brief that needs the
reader's judgment to fill a gap gets a different result from each reader.

## Dispatch

Lean work runs through exophial, and the workers run on Sonnet (operator, 2026-10-28;
`engineering/operating-rules.md` §1). The worker backend is `omp`, because the
search rule below needs omp's `task` and `web_search` tools. The pool proofs run
one at a time. The Nitro proofs run at the same time as the pool proofs. No
Lean work goes to Qwen on the 5090 until the experiments in
`engineering/operating-rules.md` §1 show that Qwen can write Lean.

Each worker works in its own worktree of the branch that the brief names. It
merges and never rebases (`anti-patterns.md` P4).

## Search before proving, and keep the context clean

Use Mathlib where it makes sense (operator ruling, 2026-10-27). Before a
worker writes a general definition or lemma (one whose statement names no
project type), it searches for an existing one:

1. In Lean, inside the worktree: `exact?`, `apply?`, `rw?` and `simp?` in a
   trial file (`lake env lean .scratch/trial.lean`).
2. In the Mathlib source (`.lake/packages/mathlib/`), and on Loogle
   (loogle.lean-lang.org) and LeanSearch (leansearch.net) for a lemma by
   type or by description.
3. On the internet, for an existing Lean proof of the same fact (a Merkle
   tree, list permutation and removal, a hash injectivity argument).

The worker does not read search results in its own context. It spawns a
`scout` subagent (omp `task` tool) for a search of code or of the Mathlib
source, or it uses `web_search` (Gemini) for the internet. It takes back
only the answer: the declaration name, its file, its full signature, and the
source URL.

The report lists each search, what it found, and what the worker used. Each
general lemma written by hand states why no library lemma fits
(`verification.md` §6a).

## The acceptance tests the design

"`lake build` passes with zero `sorry`" is necessary. It is not sufficient.
Sources A, B, C and D in `anti-patterns.md` all passed that gate. Every
brief carries at least one design check from the list below, plus a
line-count target. Every design check is a command with an expected output.

### Design checks a script can run

**D1. A new declaration is used by a named statement.** The brief names the
statements. The check:

```bash
# Counter-example at 68f49f6: PoolOwnedSpendStep_effects is named once outside
# its declaration, in a comment (Lemmas.lean:4738). It would fail this check.
grep -n "\bPoolOwnedSpendStep_effects\b" ArmadaLean/Adjudicator/Lemmas.lean ArmadaLean/Adjudicator/Safety.lean
# Check for a new lemma (here DepositStep_step, added by step A1):
grep -c "\bDepositStep_step\b" ArmadaLean/Adjudicator/Lemmas.lean ArmadaLean/Adjudicator/Safety.lean
```

Expected: a count ≥ 2 in the file that holds the named statement (the
declaration line plus one use), or the brief lists the exact line numbers of
the uses. A hit inside a comment does not count; the brief says which lines
are uses.

**D2. No pool-internal operator name in an adjudicator file.** After phase A
the adjudicator layer speaks only through the pool's public step and its
theorems.

```bash
grep -cE "removeAll|markSpent|insertResult|applyRecords|applyOwnedRecords|withTreePositionsFrom|transactSpentPositions|ownedSpentPositions" ArmadaLean/Adjudicator/Notes.lean
```

Expected: the target number the brief states (`68f49f6`: 28; after A1 the
`DepositStep` body contributes 0).

**D3. A definition references a named pool constructor or post-state.**

```bash
grep -n "Step.transactPost\|_root_.Step.transact\b" ArmadaLean/Adjudicator/Notes.lean
```

Expected: at least one hit inside the body of the named definition (the
brief gives the line range after the edit, or says "between the lines
`def DepositStep` and the next `def`").

**D4. No `True`/`False` body and no signature change.** `scripts/check.sh`
(armada-lean): its signature and non-vacuity checks (`verification.md` sections 2 and 4).

**D5. A file stays under N lines; a declaration stays under N lines.**

```bash
wc -l < ArmadaLean/Adjudicator/Notes.lean
python3 .scratch/declsizes.py ArmadaLean/Adjudicator/Lemmas.lean step_self
```

Expected: the brief's numbers.

**D6. No literal `PoolState` construction in a can-happen proof.** After phase
C, a can-happen proof instantiates the exact-effect lemmas; it builds no pool
state by hand.

```bash
sed -n '/^theorem payout_can_happen/,/^theorem /p' ArmadaLean/Adjudicator/Safety.lean | grep -c "PoolState :=\|: PoolState :=\|{ .* with live :="
```

Expected: 0.

**D7. A tactic is present or absent.**

```bash
sed -n '/^theorem step_self/,/^theorem /p' ArmadaLean/Adjudicator/Lemmas.lean | grep -c "obtain ⟨"
```

Expected: 0 after step B0.

**D8. The frame holds by `rfl`.** For a step rewritten with `{ s with … }`,
the brief adds an `example` to a trial file:

```lean
example (s : AdjudicatorState) (fp : FixedPart) (candidate : VariablePart) :
    (challengePost s fp candidate).now = s.now := rfl
```

and checks it with `lake env lean .scratch/trial.lean` (`verification.md`
section 8). `[INFERENCE: challengePost does not exist yet; the example shows
the shape the B2 brief will state once the name is fixed.]`

### The line-count target

Every brief states a target as an inequality on numbers the scripts print:

- `declsizes.py FILE NAME` ≤ N for each named declaration, or `absent`;
- `wc -l < FILE` ≤ N;
- `grep -c` of a pattern ≤ N.

The target is the measured goal of the step, not a guess. When a step is a
measurement, the target is "report the numbers"; the next brief uses them.

## Interfaces

A brief that depends on a declaration another step produces names it with its
exact signature. If the signature is not fixed yet, the brief cannot be
issued. Two steps that share an interface never run in parallel unless the
interface is already committed (anti-pattern P2).

## Claims

Every statement about Lean in a brief is one of:

- checked: the brief gives the command and the observed output, or
- `[INFERENCE]`: marked as such.

A brief that says "`grind` will close this" without a run on `v4.33.1` is
not issuable.

## Executability

These rules hold for every reader of a brief. Their justification: a brief
is a specification of an edit, and `docs/process.md` practice 3 applies to
it ("write the specification, then derive the program from it").

- An edit is an exact replacement that fails when its anchor text is absent
  (`python3 - <<'PY'` with an `assert`, or `sed -n` line ranges plus a
  heredoc). An edit never asks the reader to retype an existing block.
  Retyping introduces errors at every skill level.
- A constant longer than 16 characters (a sha, a Merkle root, a hash input)
  is produced by a command (`git rev-parse`, `sed -n 'N p'`), never
  transcribed. Transcribed hex has been corrupted in this project
  (`anti-patterns.md` P6).
- A brief edits one file, or one small group of declarations that one build
  checks. `reduction-plan.md` gives the size limits used for execution.
- The reader runs `scripts/check.sh` (armada-lean, `verification.md` "Current check") before reporting. A report
  with no script output is a failed run, whoever wrote it.
- The reader does not write a plan, change a statement text, add an axiom, or
  add an allowlist name. These are the operator's and the coordinator's
  decisions (`docs/process.md` 8.3; `verification.md` sections 3 and 4).
- When a declaration cannot be completed, the reader leaves `sorry` in that
  one declaration, runs the script, commits with `WIP:` in the subject,
  pushes, and reports the exact error text. A partial result with evidence
  is useful. A silent stop is not.

## Template

```
# Brief <STEP-ID>: <one-line title>

## Worktree and branch
The exophial worker worktree of armada-lean, cut from `integrated`.
Run `git branch --show-current` first. Expected: the worker branch exophial names.

## Depends on
<step ids, each with its commit sha>, or "none".

## Read first
docs/lean/writing-lean.md rules <n, m>. docs/lean/verification.md section 1.
<file:line-range> for each declaration below.

## Search first
For each general lemma the proof needs: the searches of "Search before proving"
in docs/lean/worker-briefs.md, through a scout or web_search. Report each one.

## Change
1. In <FILE>, declaration <NAME> (<start line> at <SHA>): <exact edit>.
   New text:
   ```lean
   <full new declaration>
   ```
2. ...

## Commands
cd <worktree path>
<edit commands>
~/.elan/bin/lake build
scripts/check.sh
scripts/check.sh --json <NAME>

## Acceptance (all must hold)
- `scripts/check.sh` exits 0.
- `scripts/check.sh --json <NAME>`: `<NAME>` is not in `sorry`.
- D<k>: <command> prints <expected>.
- `python3 scripts/declsizes.py <FILE> <NAME>` prints LINES ≤ <N>.
- <other targets>

## Report
Commit sha. Full output of `scripts/check.sh`. The numbers from each
acceptance command. Each search, what it found, and what was used. If a
declaration still carries `sorry`: its name and the first error line of
`lake build`.

## Do not
Change any theorem text. Add an axiom. Rebase. Force-push. Edit a file not
named above. Write a plan. Read search results in your own context.
```

## A filled example (step B0 of `reduction-plan.md`, on the old `adjudicator-model-v4` branch)

```
# Brief B0: six frame lemmas by grind

## Worktree and branch
~/code/git_puller/repos/armada-lean-adjudicator-v4, branch adjudicator-model-v4.
Baseline: 68f49f6.

## Depends on
none.

## Read first
docs/lean/writing-lean.md rules 6 and 8. docs/lean/verification.md section 1.
ArmadaLean/Adjudicator/Lemmas.lean:1184-1530.

## Change
Replace the proof body (everything after `:= by`) of each of these six
theorems in ArmadaLean/Adjudicator/Lemmas.lean with one line. Keep each
statement text unchanged.
- step_configuredAssets (line 1184), step_now_frame (1277),
  step_signedBy_mono (1346), step_signedBy_frame (1394),
  poolSignedBy_frame (1478):
    cases call <;> simp only [Adjudicator.Step] at hstep <;> grind
- step_self (1229):
    cases call <;> simp only [Adjudicator.Step] at hstep <;> grind [AdjudicatorState.self]
(Both lines checked on 68f49f6, 2026-10-26, 1.5 s for all six.)

## Commands
cd ~/code/git_puller/repos/armada-lean-adjudicator-v4
python3 - <<'PY'
import re, pathlib
p = pathlib.Path("ArmadaLean/Adjudicator/Lemmas.lean")
t = p.read_text()
names = ["step_configuredAssets", "step_self", "step_now_frame",
         "step_signedBy_mono", "step_signedBy_frame", "poolSignedBy_frame"]
for n in names:
    m = re.search(r"^theorem " + n + r"\b.*?:= by\n(.*?)(?=^\S)", t, re.S | re.M)
    assert m, n
    body = ("  cases call <;> simp only [Adjudicator.Step] at hstep <;> grind"
            + (" [AdjudicatorState.self]" if n == "step_self" else "") + "\n\n")
    t = t[:m.start(1)] + body + t[m.end(1):]
p.write_text(t)
PY
~/.elan/bin/lake build
.scratch/check.sh 68f49f6

## Acceptance
- `.scratch/check.sh 68f49f6` prints EXIT 0.
- D7: for each name, `sed -n '/^theorem NAME/,/^theorem /p' ArmadaLean/Adjudicator/Lemmas.lean | grep -c "obtain ⟨"` prints 0.
- `python3 .scratch/declsizes.py ArmadaLean/Adjudicator/Lemmas.lean step_configuredAssets step_self step_now_frame step_signedBy_mono step_signedBy_frame poolSignedBy_frame`: every LINES ≤ 16.
- `wc -l < ArmadaLean/Adjudicator/Lemmas.lean` ≤ 4720 (68f49f6: 4941).

## Report
Commit sha, check.sh output, the six LINES values, wc -l.

## Do not
Change any theorem text. Touch any other declaration. Rebase. Force-push.
```

This brief was run on 2026-10-26 on a scratch copy of `68f49f6`
(`.scratch/trial-b0`, deleted after the run). Observed: the Python
replacement matched all six theorems; `lake build` completed in 11 s;
`Lemmas.lean` went from 4,941 to 4,718 lines; the six LINES values were 10,
14, 11, 14, 10, 11; `#print axioms` on the 59 `Safety.lean` theorems listed
only approved names (one theorem gained `Classical.choice`, which is
approved).
