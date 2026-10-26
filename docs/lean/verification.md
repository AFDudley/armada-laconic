# Verification: the checks every change must pass

Every check below is a command. A worker runs the command. The command's exit
code and output decide. No check is satisfied by a report.

Worktree: `~/code/git_puller/repos/armada-lean-adjudicator-v4` (branch
`adjudicator-model-v4`). Toolchain: `~/.elan/bin/lake`, Lean `v4.33.1`.
Baseline commit for the reduction: `68f49f6`. Scratch files go in `.scratch/`
inside the worktree. Never `/tmp`.

All scripts in this file were run on `68f49f6` on 2026-10-26. The full
check (section 1) takes 3 s when the build is current.

## 1. The acceptance script

Install once per worktree. The heredocs write three files. Copy each block
exactly.

```bash
cd ~/code/git_puller/repos/armada-lean-adjudicator-v4 && mkdir -p .scratch && cat > .scratch/sigs.py <<'EOF'
#!/usr/bin/env python3
"""Print the statement of every theorem in a Lean file, one line each.

Usage: sigs.py FILE.lean
Output: NAME<TAB>SIGNATURE, sorted by NAME. A signature is the text from
the `theorem` keyword to the first `:=` at nesting depth zero, with all
whitespace collapsed. Compare two revisions with `diff`.
"""
import re
import sys

text = open(sys.argv[1], encoding="utf-8").read()
head = re.compile(r"^theorem\s+([\w.']+)", re.M)
out = []
for m in head.finditer(text):
    i = m.start()
    depth = 0
    j = i
    while j < len(text):
        c = text[j]
        if c in "([{⟨":
            depth += 1
        elif c in ")]}⟩":
            depth -= 1
        elif depth == 0 and text.startswith(":=", j):
            break
        j += 1
    sig = " ".join(text[i:j].split())
    out.append((m.group(1), sig))
for name, sig in sorted(out):
    print(f"{name}\t{sig}")
EOF
```

```bash
cd ~/code/git_puller/repos/armada-lean-adjudicator-v4 && cat > .scratch/declsizes.py <<'EOF'
#!/usr/bin/env python3
"""Print the size of every declaration in a Lean file.

Usage: declsizes.py FILE.lean [NAME ...]
Output: NAME<TAB>START_LINE<TAB>LINES. With NAME arguments, only those
declarations; a missing NAME prints `NAME<TAB>absent<TAB>0`.
A declaration runs from its keyword line to the line before the next
declaration keyword, so its doc comment counts toward the NEXT declaration.
"""
import re
import sys

decl = re.compile(r"^(?:@\[[^\]]*\]\s*)?(?:private |protected |noncomputable )?"
                  r"(theorem|lemma|def|abbrev|structure|inductive|instance)\s+([\w.']+)")
lines = open(sys.argv[1], encoding="utf-8").read().split("\n")
starts = [(i, m.group(2)) for i, l in enumerate(lines, 1) if (m := decl.match(l))]
sizes = {}
for k, (i, n) in enumerate(starts):
    end = starts[k + 1][0] - 1 if k + 1 < len(starts) else len(lines)
    sizes[n] = (i, end - i + 1)
names = sys.argv[2:] or [n for _, n in starts]
for n in names:
    if n in sizes:
        print(f"{n}\t{sizes[n][0]}\t{sizes[n][1]}")
    else:
        print(f"{n}\tabsent\t0")
EOF
```

```bash
cd ~/code/git_puller/repos/armada-lean-adjudicator-v4 && cat > .scratch/check.sh <<'EOF'
#!/usr/bin/env bash
# Acceptance check for one change to the adjudicator model.
# Usage: .scratch/check.sh BASELINE_SHA [ALLOWLIST_FILE]
# Run from the worktree root. Exit 0 means every check passed.
set -u
BASE="$1"
ALLOW="${2:-/dev/null}"
LAKE="$HOME/.elan/bin/lake"
OUT=.scratch/check
mkdir -p "$OUT"
fail=0
say() { printf '%s\n' "$*"; }

# 1. Build, exit 0, no sorry warning.
"$LAKE" build > "$OUT/build.log" 2>&1
rc=$?
if [ "$rc" -ne 0 ]; then say "FAIL build exit $rc (see $OUT/build.log)"; fail=1; else say "PASS build"; fi
if grep -q "declaration uses 'sorry'" "$OUT/build.log"; then say "FAIL sorry warning in build"; fail=1; else say "PASS no sorry warning"; fi

# 2. No vacuous Prop body.
if grep -rnE ":=\s*(True|False)\s*$" ArmadaLean > "$OUT/vacuous.txt"; then say "FAIL vacuous body:"; cat "$OUT/vacuous.txt"; fail=1; else say "PASS no vacuous body"; fi

# 3. Axioms of every theorem in Safety.lean.
python3 .scratch/sigs.py ArmadaLean/Adjudicator/Safety.lean | cut -f1 \
  | sed 's/^/#print axioms ArmadaLean.Adjudicator./' > "$OUT/ax-body.txt"
{ echo "import ArmadaLean"; cat "$OUT/ax-body.txt"; } > "$OUT/axioms.lean"
"$LAKE" env lean "$OUT/axioms.lean" > "$OUT/axioms.out" 2>&1
python3 - "$OUT/axioms.out" <<'PY' || fail=1
import re, sys
ok = {"propext", "Classical.choice", "Quot.sound", "EdDSAVerify",
      "ArmadaLean.Adjudicator.decodeGuarantee"}
text = open(sys.argv[1]).read().replace("\n", " ")
bad = 0
for m in re.finditer(r"'([\w.]+)' depends on axioms: \[([^\]]*)\]", text):
    names = {n.strip() for n in m.group(2).split(",") if n.strip()}
    extra = names - ok
    if extra:
        print("FAIL axioms", m.group(1), sorted(extra)); bad = 1
if "error" in text.lower():
    print("FAIL axioms.lean did not elaborate"); bad = 1
print("PASS axioms" if not bad else "FAIL axioms")
sys.exit(bad)
PY

# 4. Statement signatures unchanged against the baseline.
git show "$BASE:ArmadaLean/Adjudicator/Safety.lean" > "$OUT/Safety-base.lean"
python3 .scratch/sigs.py "$OUT/Safety-base.lean" > "$OUT/sigs-base.tsv"
python3 .scratch/sigs.py ArmadaLean/Adjudicator/Safety.lean > "$OUT/sigs-now.tsv"
python3 - "$OUT/sigs-base.tsv" "$OUT/sigs-now.tsv" "$ALLOW" <<'PY' || fail=1
import sys
def load(p):
    d = {}
    for line in open(p, encoding="utf-8"):
        if "\t" in line:
            k, v = line.rstrip("\n").split("\t", 1); d[k] = v
    return d
base, now = load(sys.argv[1]), load(sys.argv[2])
allow = {l.strip() for l in open(sys.argv[3]) if l.strip() and not l.startswith("#")}
bad = 0
for k, v in base.items():
    if k in allow: continue
    if k not in now: print("FAIL statement removed:", k); bad = 1
    elif now[k] != v: print("FAIL statement changed:", k); bad = 1
print("PASS signatures" if not bad else "FAIL signatures")
sys.exit(bad)
PY

# 5. Line counts, now and at the baseline.
for f in ArmadaLean/Adjudicator/Lemmas.lean ArmadaLean/Adjudicator/Safety.lean ArmadaLean/Adjudicator/Notes.lean ArmadaLean/Adjudicator/Operations.lean; do
  b=$(git show "$BASE:$f" | wc -l); n=$(wc -l < "$f"); say "LINES $f base=$b now=$n"
done

exit $fail
EOF
chmod +x .scratch/check.sh
```

Run:

```bash
cd ~/code/git_puller/repos/armada-lean-adjudicator-v4 && .scratch/check.sh 68f49f6; echo EXIT $?
```

Expected output on `68f49f6` (observed 2026-10-26):

```
PASS build
PASS no sorry warning
PASS no vacuous body
PASS axioms
PASS signatures
LINES ArmadaLean/Adjudicator/Lemmas.lean base=4941 now=4941
LINES ArmadaLean/Adjudicator/Safety.lean base=6699 now=6699
LINES ArmadaLean/Adjudicator/Notes.lean base=402 now=402
LINES ArmadaLean/Adjudicator/Operations.lean base=582 now=582
EXIT 0
```

A step passes when `EXIT 0` prints and the step's own line-count target
(section 6) is met.

## 2. Build

`~/.elan/bin/lake build` must exit 0. The build log must not contain
`declaration uses 'sorry'`. A `sorry` inside a comment does not trigger the
warning; the `grep -rn sorry` in `docs/how-to-verify.md` reports one comment
hit at `Safety.lean:3` and is not used here.

## 3. Axioms

`#print axioms` on every `theorem` in `Safety.lean` (59 at `68f49f6`) must
list only these names:

```
propext
Classical.choice
Quot.sound
EdDSAVerify
ArmadaLean.Adjudicator.decodeGuarantee
```

Observed on `68f49f6`: 33 theorems list all five; 20 list four (no
`Classical.choice`); the rest list subsets. `sorryAx` must never appear. The
script fails on any other name. A new named assumption needs the operator's
approval and an edit to the `ok` set in `check.sh`, in the same commit.

## 4. Statement signatures

The contract is the 45 statements of `Safety.lean` (32 safety, 11 can-happen,
2 payment guarantees, `labbook/2026-10-26-lean-adjudicator-proof-size.md`).
The coordinator has not published the list of 45 names. Until it does, the
script freezes every `theorem` in `Safety.lean` against the baseline. A step
that must change a helper signature in `Safety.lean` (for example
`transfer_step_exists`) lists the helper in an allowlist file and passes the
file as the second argument:

```bash
printf 'transfer_step_exists\n' > .scratch/allow-C1.txt
.scratch/check.sh 68f49f6 .scratch/allow-C1.txt
```

The allowlist is part of the brief. A worker never adds a name to it.

The signature is the text from `theorem` to the first top-level `:=`,
whitespace-collapsed. A reordering of hypotheses counts as a change. A doc
comment change does not.

Statements whose signature names the shape of `holdings` (relevant to phase H
of the reduction plan; found with
`python3 .scratch/sigs.py ArmadaLean/Adjudicator/Safety.lean | grep -P '\.holdings|some \(\(' | cut -f1`):
`bulk_conclude_and_transfer_all_assets_effect`,
`bulk_transfer_all_assets_effect`, `conclude_moves_no_value`,
`conclude_step_finalizes`, `deposit_can_happen_witness`,
`full_payout_leaves_no_remainder`, `heldValue_zero_iff_no_collateral`,
`holdings_never_zero_valued`, `holdings_point_to_live_notes`,
`holdings_position_leaf_value_placeholder`, `holdings_position_live_reachable`,
`live_note_nullifier_unspent_placeholder`, `nullify_iff_commitment_was_some`,
`reclaim_value_neutral`, `value_exists_only_as_notes` (15).

## 5. No vacuous definition

`grep -rnE ":=\s*(True|False)\s*$" ArmadaLean` must print nothing. This catches
a `def … : Prop := True` body. It does not catch `fun _ => True`. The
signature check (section 4) and the design checks in `worker-briefs.md` cover
the rest: a statement whose hypotheses lost a definition changes its
signature or its `#print axioms` set, or the design check D1 fails.

## 6. Line counts

Measure before and after. The baseline numbers at `68f49f6`:

| File | Lines |
|---|---:|
| `ArmadaLean/Adjudicator/Lemmas.lean` | 4,941 |
| `ArmadaLean/Adjudicator/Safety.lean` | 6,699 |
| `ArmadaLean/Adjudicator/Notes.lean` | 402 |
| `ArmadaLean/Adjudicator/Operations.lean` | 582 |

Per declaration:

```bash
python3 .scratch/declsizes.py ArmadaLean/Adjudicator/Lemmas.lean step_self step_configuredAssets
```

prints `NAME<TAB>START<TAB>LINES`, or `NAME<TAB>absent<TAB>0` when the
declaration is gone. A step's target is written as an inequality on these
numbers (`reduction-plan.md`). Other counts used as design metrics:

```bash
grep -c "obtain ⟨" ArmadaLean/Adjudicator/Lemmas.lean ArmadaLean/Adjudicator/Safety.lean
# 68f49f6: 267 and 248
grep -cE "removeAll|markSpent|insertResult|applyRecords|applyOwnedRecords|withTreePositionsFrom|transactSpentPositions" ArmadaLean/Adjudicator/Lemmas.lean ArmadaLean/Adjudicator/Safety.lean ArmadaLean/Adjudicator/Notes.lean
# 68f49f6: 146, 119, 28
grep -rc "grind" ArmadaLean/Adjudicator/*.lean
# 68f49f6: 0 in every file
```

## 7. Git rules

- Work on `adjudicator-model-v4` in the worktree above. Confirm with
  `git branch --show-current` before any git command.
- Integrate `pool-model` with `git merge AFDudley.github.com/pool-model`.
  Never `git rebase`. Never `git push --force` or `--force-with-lease`.
- After each step passes section 1: `git add -A ArmadaLean && git commit`
  with the step id in the subject line (`A1: compose DepositStep over
  Step.transactPost`), then `git push AFDudley.github.com adjudicator-model-v4`.
- Report the commit sha and the full output of `.scratch/check.sh` to the
  coordinator. The output is the evidence. A summary is not.
- A pool fact goes to `pool-model`, not to the adjudicator branch. The step
  says which branch it edits.

## 8. Type-checking one example file

To check a Lean snippet outside the build (an example in a brief, a trial of
`grind` on one lemma):

```bash
cd ~/code/git_puller/repos/armada-lean-adjudicator-v4 && cat > .scratch/trial.lean <<'EOF'
import ArmadaLean
open ArmadaLean ArmadaLean.Adjudicator
-- declarations here
EOF
~/.elan/bin/lake env lean .scratch/trial.lean; echo EXIT $?
```

`EXIT 0` with no output means every declaration in the file elaborated. The
build must be current first (`lake build`). Delete the trial file after use.
