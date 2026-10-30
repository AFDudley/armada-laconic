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
