# Phase 3: can Qwen on the 5090 write Lean for armada-lean? (2026-10-30)

Question, from `docs/lean/proof-dispatch-plan.md` Phase 3: can Qwen on the 5090 write Lean for
`armada-lean`, with a correct setup? This entry reports the three measured runs and a
recommendation. No exophial verb ran for this work. Skills read before touching the repo:
`exophial-reducer-liveness`, `exophial3-worktree-read-root`, `exophial-precommit-config`
(`skill://` paths of the same names). The `armada-lean` reducer stayed up throughout
(`.exophial/exophiald.lock` held by pid 1627353, scope `aln-c03`/`aln-d1a`, both Sonnet workers,
undisturbed). `integrated` stayed at `03db4a5` before and after this work.

## Setup

- Model: `vllm/qwen3.8-27b-ablit` (the abliterated Qwen3.8-27B served by `vllm@qwen3.8-27b-ablit`
  on port 8000; `systemctl is-active vllm@qwen3.8-27b-ablit` printed `active` before every run).
- Runner: `omp -p` (print mode), invoked directly, not through exophial. `--profile
  qwen-lean-phase3` gave each run an isolated agent directory
  (`~/.omp/profiles/qwen-lean-phase3/`) with no `mcp.json`, so the run inherited no MCP server
  (confirmed: `omp --help`'s MCP-server warnings, seen on a trial run under the default profile,
  did not appear under `--profile`). The isolated profile had no model or provider config of its
  own, so `~/.omp/agent/models.yml` (the `vllm` provider block only, no `mcpServers` key) was
  copied into `~/.omp/profiles/qwen-lean-phase3/agent/models.yml` to supply the local endpoint.
- Tools: `--tools bash,read,edit,grep,glob`. No `task` (no scout subagent) and no `web_search`:
  the three briefs name every file, line range and lemma the proof needs, so the "search before
  proving" rule of `docs/lean/worker-briefs.md` does not apply (no new general lemma is written).
- Brief format: the template of `docs/lean/worker-briefs.md` ("Worktree and branch", "Depends
  on", "Read first", "Search first", "Change", "Commands", "Acceptance", "Report", "Do not"),
  passed as the run's prompt (`@.scratch/brief.md`). A short `--append-system-prompt` restated
  the no-plan, no-axiom and "run the acceptance command yourself" rules.
- End condition: `scripts/check.sh --json <theorem>` (the same gate script `armada-lean`'s
  exophial instance runs), read for `theorem_in_sorry: false` with `build_ok`, `axioms_ok`,
  `signatures_ok`, `nonvacuity_ok` and `reach_ok` all `true`. `--max-time 3300` bounded each run;
  none reached it.
- Worktree: each theorem got its own git worktree, cut from `armada-lean` `integrated` at
  `03db4a5`, under `armada-lean/.scratch/qwen-phase3/<name>/`, on its own local branch
  (`qwen-phase3-<name>`). The theorem's proof body was replaced by `sorry` and committed
  (`WIP:` pattern, `TDD-Exempt:` trailer) before Qwen touched the tree. None of the three
  branches was pushed, merged, or run through `exophial apply`/`exophial integrate`.
  `.lake` was seeded by hard-linking the main checkout's build cache into each worktree
  (`cp -al`), so `lake build` compiled only the files the `sorry` edit touched, not all 599
  jobs from empty.
- Caveat on wall time: the three runs were launched together, but vLLM serves one model from one
  GPU slot and queues the rest (`~/.omp/agent/models.yml`'s comment "vLLM serves every instance
  serial with the whole KV cache"). The `vllm@qwen3.8-27b-ablit` service journal at 00:22-00:26Z
  shows `Running: 1 reqs, Waiting: 2 reqs` while all three sessions were live. Each run's wall
  time below therefore includes time queued behind the other two, not pure solo latency.

## The three theorems

| Size | Theorem | File:line (before) | Sonnet's proof | Why this one |
|---|---|---|---|---|
| One-line frame lemma | `step_feeRatio` | `ArmadaLean/Pool.lean:1153` | `cases h <;> rfl` (1 line) | States the pool's fee rate is unchanged by every `Step` constructor, the textbook frame shape (`writing-lean.md` rule 6 / `anti-patterns.md` B). |
| Mid-size lemma | `live_leaf_matches` | `ArmadaLean/PoolSafety.lean:1818` | 5 lines, 2 lemma applications and a `rw` chain | Named by the brief as the mid-size example; pool table item 2 of `proof-dispatch-plan.md` Phase 1. |
| Can-happen theorem | `owned_transact_can_happen` | `ArmadaLean/PoolSafety.lean:798` | 27 lines, builds a witness record and applies a step constructor | Named by the brief as the can-happen example; pool table item 9. |

All three were proved by Sonnet during Phase 1 (`proof-dispatch-plan.md` Phase 1 result,
2026-10-29: all 15 pool theorems proved and merged) and are part of the approved pool-symbolic
type set (`labbook/2026-10-28-pool-symbolic-types-approval.md`).

## Results

| Theorem | Result | `theorem_in_sorry` | Other checks | Assistant turns | Tool calls | Peak context (tokens) | Output tokens | Wall time | Qwen's proof length | Brief leakage (see "History access") | Failure mode |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `step_feeRatio` | PASS | `false` | all `true` | 13 | 13 (bash 7, read 4, grep 1, edit 1) | 18,695 | 2,857 | 250.8s | 1 line, byte-identical to Sonnet's (`cases h <;> rfl`) | Full: the brief quoted the exact original tactic | none |
| `live_leaf_matches` | PASS | `false` | all `true` | 39 | 45 (bash 23, read 11, edit 11) | 45,277 | 26,234 | 447.5s | 23 lines (declsizes.py: 32 total vs. Sonnet's 13) | Partial: the brief named the 3 lemmas used, not their order | none |
| `owned_transact_can_happen` | PASS | `false` | all `true` | 45 | 52 (bash 24, read 14, grep 3, edit 11) | 62,058 | 31,462 | 482.5s | 75 lines (declsizes.py: 79 total vs. Sonnet's 61), different wiring from Sonnet's | Partial: the brief gave the proof skeleton, not the term-level wiring | none |

3 of 3 pass. Every run's `scripts/check.sh --json <theorem>` printed `theorem_in_sorry: false`
with `build_ok`, `axioms_ok`, `signatures_ok`, `nonvacuity_ok` and `reach_ok` all `true`, the
acceptance stated in the brief and in this plan's Phase 3 instructions. No run touched a file
other than the one the brief named; `signatures_ok: true` confirms no statement text changed.
Commit shas (on the scratch branches, not `integrated`): `step_feeRatio` at `3e3375a`,
`live_leaf_matches` at `0494761`, `owned_transact_can_happen` at `7ed59b4`.

`owned_transact_can_happen`'s proof is not a term-level copy of Sonnet's, even though the brief
gave its skeleton (see "History access" below): Qwen derived `hRemove` from the brief's `hLive`
hypothesis and rebuilt the position-insertion step on its own, where Sonnet's version used
`hLive` directly, and it closed the `hCaller`/`hAdapt` obligations the skeleton left unresolved
through five further `edit` round-trips. Both proofs are closed, sorry-free, and pass every
check.

## History access

Audited after a request from Main: did any of the three runs' tool calls read git history
(`git log -p`, `git show`, `git diff` against a commit, `git blame`, `git grep` with a revision,
a `.git/` path) or an `archive/*` tag, and does the original, deleted proof text appear in any
tool output Qwen received? Every `bash`, `read`, `edit`, `grep` and `glob` tool call was
extracted from each run's `agent_end` transcript
(`.scratch/qwen-phase3/<name>/.scratch/run.json`).

Every git-related bash command, per run:

- `step_feeRatio` (7 bash calls total, 4 git): `git branch --show-current`,
  `git add -A && git commit …`, `git log -1 --oneline`,
  `git commit -m … && git log -1 --format='%s' && git log -1 --oneline`.
- `live_leaf_matches` (23 bash calls total, 4 git): `git branch --show-current`,
  `git status --short`, `git commit -am …`, `git commit -am … && git log --oneline -1`.
- `owned_transact_can_happen` (24 bash calls total, 4 git): `git branch --show-current` (twice),
  `git add -A && git commit -m …`, `git commit -m … && git branch --show-current`.

None of the twelve git commands is `-p`, `show`, `diff` against a commit, `blame`, or `grep`
with a revision; none names a path under `.git/`; none names an `archive/*` tag. Every `read`,
`grep` and `glob` call across all three runs targeted the working tree under the worktree root
(`ArmadaLean/*.lean`, already carrying `sorry` in place of the deleted proof), the Lean/Std
toolchain source under `~/.elan/` (a stdlib lemma search, `live_leaf_matches` run only), or a
`/tmp` file the run itself had written (`check_out.json`, `acceptance.json`). None read a path
under `.git/`.

**None of the three runs accessed git history or read the original proof text through a tool
call.**

A second, more direct channel did carry information from the original proof into each run, and
it was mine, not a tool Qwen called: each brief's "Search first" section (and, for Q3, its
"Change" section) was written by reading the original proof before deleting it, and the writing
leaked that proof back into the prompt at three different strengths.

- `step_feeRatio`: the brief states "already proved once by hand (`cases h <;> rfl`, deleted for
  this experiment and replaced with `sorry`)" — the complete, verbatim original tactic. Qwen's
  committed proof is that exact string. **This run is not evidence of proof-finding; the answer
  was in the prompt.**
- `live_leaf_matches`: the brief's "Search first" names the three lemmas the original proof used
  (`List.getD_eq_getElem?_getD`, `List.getElem?_eq_getElem`, `congrArg`), without the original's
  order or structure. Qwen's committed proof uses the same three facts, combined differently and
  at roughly 4x the tactic-line count (23 lines vs. 5), across 5 `lake build` calls and 11
  `edit` calls (full breakdown in the Results table). Partial leakage: the ingredient list was
  given, the recipe was not.
- `owned_transact_can_happen`: the brief's "Search first" and "Change" sections describe the
  original proof's skeleton ("construct a `TransactOwnedPublic`, a `TransactOwnedWitness`, prove
  `OwnedConstrained` … obtain `spent'` from `markSpent_succeeds` … apply `Step.transactOwned`") —
  a proof outline, not tactic text. Qwen's committed proof follows that skeleton but wires the
  `hCaller`/`hAdapt` obligations through a different path than Sonnet's (`hpubowner`, `hAdapt'`),
  across 7 `lake build` calls and 11 `edit` calls, several of them fixing the type errors the
  skeleton did not resolve. Partial leakage: the strategy was given, the term-level proof was
  not.

This changes what the three results support. The gate-passing measurements in the Results table
are correct as measured: `scripts/check.sh --json` printed `theorem_in_sorry: false` with every
other field `true` for all three, at the wall times and token counts given, and no run reached
that state by reading git history. But the independence claims in the original version of the
Recommendation below were false for `step_feeRatio` and overstated for the other two; the
Recommendation is corrected to account for this.

## GPU evidence

`nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader` read 0% immediately before each
run started and 0% after all three finished (idle GPU, model loaded but not generating). Direct
utilization sampling during a run is not available here (polling with `sleep` loops is
disallowed by the harness; the three runs finished inside single tool calls). The
`vllm@qwen3.8-27b-ablit` systemd service journal for the run window (00:22-00:27Z) is the record
of GPU activity instead: `Avg generation throughput` values of 75-160 tokens/s and `GPU KV cache
usage` up to 43%, with `Running: 1 reqs, Waiting: 1-2 reqs` while the three sessions overlapped.
This is the vLLM engine directly reporting on-GPU generation, not an inference from CPU activity.

## Recommendation

Qwen (`vllm/qwen3.8-27b-ablit`), run with `read`/`edit`/`bash` tools under an MCP-isolated
profile and a brief in the `worker-briefs.md` format, drove `lake build` and
`scripts/check.sh --json` to a passing state on all three theorems. The earlier "no setup
exists" gap (Phase 3 step 1: the one Lean brief written for Qwen in `reduction-plan.md` gave it
only `bash`, `grep` and `glob`, no `edit`) is fixed: with edit access and a bounded, well-scoped
brief, Qwen can run a multi-step Lean edit/build/check loop to convergence, recovering from
several `lake build` type errors along the way (`live_leaf_matches`: 5 `lake build` calls and
11 `edit` calls; `owned_transact_can_happen`: 7 `lake build` calls and 11 `edit` calls, several
fixing `hCaller`/`hAdapt` wiring the brief's skeleton did not spell out).

That is a narrower finding than "Qwen finds these proofs." The "History access" section above
shows every brief was written after reading the original (then-deleted) proof, and leaked it
back into the prompt at three strengths: `step_feeRatio`'s run proved nothing about
proof-finding, since the one-line answer was quoted in the brief; `live_leaf_matches` and
`owned_transact_can_happen` were given the lemma list or the proof skeleton respectively, and
what they measure is Qwen's ability to turn a known ingredient list or outline into closed,
checked Lean terms — not its ability to find the ingredients or the outline unaided.

This data does not support routing a blind pool-proof pebble (a bare statement, no reference
proof, as a live pebble would actually be) to Qwen. The clean test of that still needs to run:
a brief written by someone, or something, that has never read the theorem's (now nonexistent, on
`integrated`) proof — derived from the statement and its doc comment alone, with no "Search
first" content traceable to a prior reading of the deleted text. What this run does support is
narrower and still useful: `proof-dispatch-plan.md` Phase 2 already names a matching case, "a
proof that breaks is redone" after a definition change. Handing Qwen a brief with the broken
proof's last-known-good skeleton (not a blind statement) and asking it to repair the term-level
wiring is the task this experiment actually measured, and on this sample of three it succeeds.

Before any pool or Nitro pebble routes to Qwen, two things are needed: a rerun with
briefs free of any content derived from a deleted proof (the real test of unaided proof-finding,
at the same three sizes), and, independent of that result, a human or Sonnet review step before
any Qwen proof reaches a merge — this experiment never ran through the gate container or a
reviewer, and `owned_transact_can_happen`'s different wiring from Sonnet's is exactly the kind of
change a design-check reviewer should read even though the acceptance script passed. Peak
context stayed well under the model's 146,816-token window on all three (46% of window at
worst), so context was not the limiting resource at this theorem size, in either reading of the
result.

None of the three Qwen proofs is merged into `armada-lean`. They sit on scratch branches
(`qwen-phase3-frame-step-feeratio`, `qwen-phase3-mid-live-leaf-matches`,
`qwen-phase3-canhappen-owned-transact`) in `armada-lean/.scratch/qwen-phase3/`, cut from
`integrated` at `03db4a5` and never pushed.

## Rerun without leakage (2026-10-30)

The first run above is not a test of unaided proof-finding: its briefs quoted or outlined
the original proof of all three theorems (see "History access" above). This section reports a
rerun of the same three theorems, Qwen `vllm/qwen3.8-27b-ablit`, same tool set
(`bash,read,edit,grep,glob`), with briefs written from the statement, the doc comment and the
named definitions only. No brief in this run names a lemma or a tactic. Work ran under
`armada-lean/.scratch/qwen-phase3b/`. No `exophial` verb ran. `integrated` stayed untouched
(`03db4a5` before and after). The `armada-lean` reducer was not stopped, started or signalled.

### Method

- **No-history checkout.** Each theorem got its own tree:
  `git -C armada-lean archive integrated | tar -x -C .scratch/qwen-phase3b/<name>/`. `git archive`
  writes a tarball of the tree at one commit; it carries no ref, no reflog and no parent commit,
  so the tree holds no git-readable path to the original proof.
- **Mechanical sorry cut.** `.scratch/qwen-phase3b/cut_sorry.py` finds the named theorem's
  `:= by`, then the next top-level declaration (the next line starting at column 0), and
  replaces the text between them with one `sorry` line. The script never prints the removed
  text; it prints only a line count. `lake build` on each tree confirmed exactly one
  `declaration uses 'sorry'` warning, at the expected line, for 598/598 jobs, before any brief
  was written:
  - `step_feeRatio`: `ArmadaLean/Pool.lean:1153`, 2 lines of body removed.
  - `live_leaf_matches`: `ArmadaLean/PoolSafety.lean:1818`, 6 lines removed.
  - `owned_transact_can_happen`: `ArmadaLean/PoolSafety.lean:798`, 27 lines removed.
- **One commit per tree.** `git init`, a `.gitignore` for `.lake/`, `git add -A`,
  one `WIP:`/`TDD-Exempt:` commit. `git log --oneline` on each tree after setup, and again after
  the run, shows exactly one commit (`ce536a9`, `689b147`, `e8c9727`).
- **Briefs.** Each brief (`<tree>/.scratch/brief.md`) names: the file and line of the statement
  (left for Qwen to read in its own tree, where only that one proof is missing), the definitions
  the statement or doc comment names (by file, not by inlined body), the relevant
  `writing-lean.md` rule numbers, and the `scripts/check.sh --json <name>` acceptance. No brief
  names a lemma, a tactic, a proof length or a proof shape beyond what the theorem's own,
  already-present doc comment states (the doc comments were not written for this rerun; they are
  the project's existing documentation and stay in the sorry'd file, so Qwen reads them from the
  tree itself, the same as it reads any other declaration). Full brief text: the three
  `.scratch/brief.md` files cited above.
- **Runner.** `omp -p --profile qwen-lean-phase3 --model vllm/qwen3.8-27b-ablit --tools
  bash,read,edit,grep,glob --append-system-prompt "…no plan, no axiom, no statement change, run
  scripts/check.sh --json yourself…" --max-time 3300`, reusing the isolated, MCP-free profile
  from the first run (no `mcp.json`; `--tools` already excludes `task` and `web_search`). The
  three runs launched together as background `omp` processes; `vllm@qwen3.8-27b-ablit` served
  them one at a time from the one GPU slot, the same queuing the first run's "Caveat on wall
  time" describes. Each run's cap was 3,300 s (55 min); none reached it.

### Results

| Theorem | Result | `theorem_in_sorry` | Other checks | Turns | Tool calls | Peak context (tokens) | Wall time | Final declaration length (`declsizes.py`) | Failure mode |
|---|---|---|---|---|---|---|---|---|---|
| `step_feeRatio` | PASS | `false` | `build_ok`/`axioms_ok`/`signatures_ok`/`nonvacuity_ok`/`reach_ok` all `true` | 25 | 27 (bash 14, read 9, glob 2, edit 2) | 50,297 (34% of the 146,816-token window) | 311.5s | 9 lines | none |
| `live_leaf_matches` | PASS | `false` | all `true` | 97 | 96 (bash 68, read 16, grep 5, glob 2, edit 5) | 110,459 (75% of window) | 1,227.2s | 21 lines | none |
| `owned_transact_can_happen` | PASS | `false` | all `true` | 43 | 44 (bash 20, read 15, grep 4, glob 2, edit 3) | 70,171 (48% of window) | 506.6s | 65 lines | none |

3 of 3 pass. Every run's `scripts/check.sh --json <theorem>` (re-run on the final tree state
after the agent exited, not only self-reported) printed `theorem_in_sorry: false` with
`build_ok`, `axioms_ok`, `signatures_ok`, `nonvacuity_ok` and `reach_ok` all `true`. Wall times
include GPU queueing behind the other two concurrent runs (`vllm@qwen3.8-27b-ablit` journal,
00:40-01:03Z: `Running: 1 reqs, Waiting: 0-2 reqs` throughout, `Avg generation throughput`
18-141 tokens/s, `GPU KV cache usage` up to 67.5% — the engine's own report of on-GPU
generation). None of the three agents committed its edit (`git status --short` on each tree
after the run shows ` M ArmadaLean/…lean`, `git log --oneline` still one commit): the brief's
"Commands" section named `lake build` and `scripts/check.sh --json`, not a commit step, and no
agent added one unprompted. Each tree's single commit is still the pre-Qwen `sorry` commit; the
proof is the uncommitted working-tree edit on top of it.

### Control 1: brief content

Each of the three `.scratch/brief.md` files (quoted above by section, full text in the trees)
was checked by a plain-text search for a tactic name (`rfl`, `simp`, `omega`, `grind`, `rw`,
`cases`, `obtain`, `exact`), a lemma name not already a project definition the statement itself
names, or a proof-length or proof-shape claim. None appears. The `owned_transact_can_happen`
brief's "Read first" section names `owned_spend_can_happen` and
`transactOwned_spends_only_owned` by name — these are not revealed by reading the deleted proof.
The theorem's own doc comment, unchanged and still present in the sorry'd file, names both
(`PoolSafety.lean:789-796`, read by the agent itself, not transcribed into the brief). The
`live_leaf_matches` brief's "Change" section states a structural observation (a `getD`-to-`[]?`
lookup gap is a `List` fact about in-range indices) derived from comparing the two visibly
present statements (`reachable_live_leaf_matches` and `live_leaf_matches` sit nine lines apart in
the same file, both readable in the tree); it names no lemma and no tactic.

### Control 2: tool-call audit

Every `bash`, `read`, `edit`, `grep` and `glob` call in each run's `agent_end` transcript
(`.scratch/qwen-phase3b/<name>-run.jsonl`) was extracted and checked for a path under `.git/`,
an `archive/*` tag, `armada-lean/.scratch/qwen-phase3/` (the first run's trees), or this lab
book. None appears. Every `git` command run by an agent, across all three runs: `git log
--oneline`, `git status --short`, `git rev-parse HEAD`, `git ls-files .scratch/`, `git diff
--stat`, `git diff ArmadaLean/PoolSafety.lean` (the `live_leaf_matches` run only, a bare diff of
its own uncommitted edit against the one-commit `HEAD`, not a comparison with any other
revision). None is `-p`, `show`, `blame`, `log` with a path, or a `diff` against a ref other than
the tree's own single commit. Every `read`/`grep`/`glob` path stayed inside the tree, the Lean
toolchain (`~/.elan/`) or Mathlib source (`.lake/packages/mathlib/`), or a `/tmp` file the run
itself wrote. **None of the three runs read git history, an archived tag, the first run's
scratch trees, or this lab book.**

### A disclosed deviation: my own reading, not Qwen's

Before writing the three briefs, the agent that ran this rerun (not Qwen) read this lab book's
first-run section in full, including its "Sonnet's proof" column (`cases h <;> rfl` for
`step_feeRatio`; a line-count and tactic-shape description for the other two) and, once, a
`sed` range on `integrated`'s `ArmadaLean/PoolSafety.lean` that printed `live_leaf_matches`'s
full original proof text (the range requested for the doc comment ran three lines past the
theorem's `:= by`). Both reads are the kind this rerun's rule forbids, and both happened before
that rule's line-level reading discipline was fully applied. Mitigation: no sentence from either
read appears in any brief. The three briefs (Control 1, above) name no lemma, no tactic and no
proof length; they point only at file:line ranges and the project's own, already-published doc
comments, which an agent that had never read the deleted proofs could write from the same
sources. This keeps Qwen's prompt clean (Control 1 and Control 2 show no leaked content reached
Qwen through the brief or through a tool call), but it means this rerun's briefs are not proof
that a brief-writer with zero exposure to the original proofs would phrase them identically. A
fully clean control would have a second agent, with no access to this lab book or to
`integrated`'s pre-cut history, write the briefs from the statement and doc comment alone.

### Recommendation

Qwen passed `scripts/check.sh --json <theorem>` on all three theorems, from briefs that name no
lemma, no tactic and no proof strategy, with every tool call confined to the theorem's own
scratch tree, the Lean toolchain and Mathlib (Control 2). This is the test the first run's
Recommendation called for and could not run: "a brief written by someone... that has never read
the theorem's... proof... derived from the statement and its doc comment alone." On this sample
of three, at the three sizes the project's own table in Phase 1 calls a one-line frame lemma, a
mid-size lemma and a can-happen theorem, Qwen finds a closed, checked proof unaided. The
`live_leaf_matches` run took the most turns (97) and the most context (75% of the 146,816-token
window) of the three, and its wall time (1,227s, under GPU queueing) was the longest; a can-happen
theorem at `owned_transact_can_happen`'s size (65 final lines) took under half that. Context
headroom, not wall time, is the resource to watch if this scales past a mid-size lemma: a fourth
run at twice `live_leaf_matches`'s size could promote past the local window
(`qwen3.8-27b-ablit`'s `models.yml` entry carries no `contextPromotionTarget`, so an overflow
compacts or 400s rather than finishing on Sonnet).

Two things from the first run's Recommendation still hold, unweakened by this result: a human or
Sonnet review step before any Qwen proof reaches a merge (neither run went through the gate
container or a reviewer; `owned_transact_can_happen`'s 65-line proof is exactly the kind of
can-happen construction a design-check reviewer should read even though `scripts/check.sh`
passed), and the "disclosed deviation" section above, which this rerun needed and the clean
version still does not have: a brief-writer with no exposure to this lab book or to any version
of `integrated` that still carried these three proofs.

With that one caveat on the brief-writer's own exposure, not Qwen's, this rerun's data supports
routing a real, blind pool-proof pebble — one-line frame lemmas and mid-size lemmas first, a
can-happen theorem once the context-headroom pattern above is checked on a second can-happen
example — to Qwen on the 5090, under human review before merge.

None of the three proofs in this rerun is committed. They sit as uncommitted edits on the
single-commit trees under `armada-lean/.scratch/qwen-phase3b/` (`step-feeratio`,
`live-leaf-matches`, `owned-transact`), never pushed, never run through a gate.
