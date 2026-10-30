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

| Theorem | Result | `theorem_in_sorry` | Other checks | Assistant turns | Tool calls | Peak context (tokens) | Output tokens | Wall time | Qwen's proof length | Failure mode |
|---|---|---|---|---|---|---|---|---|---|---|
| `step_feeRatio` | PASS | `false` | all `true` | 13 | 13 (bash 7, read 4, grep 1, edit 1) | 18,695 | 2,857 | 250.8s | 1 line, identical to Sonnet's (`cases h <;> rfl`) | none |
| `live_leaf_matches` | PASS | `false` | all `true` | 39 | 45 (bash 23, read 11, edit 11) | 45,277 | 26,234 | 447.5s | 23 lines (declsizes.py: 32 total vs. Sonnet's 13) | none |
| `owned_transact_can_happen` | PASS | `false` | all `true` | 45 | 52 (bash 24, read 14, grep 3, edit 11) | 62,058 | 31,462 | 482.5s | 75 lines (declsizes.py: 79 total vs. Sonnet's 61), a different witness-construction path from Sonnet's | none |

3 of 3 pass. Every run's `scripts/check.sh --json <theorem>` printed `theorem_in_sorry: false`
with `build_ok`, `axioms_ok`, `signatures_ok`, `nonvacuity_ok` and `reach_ok` all `true`, the
acceptance stated in the brief and in this plan's Phase 3 instructions. No run touched a file
other than the one the brief named; `signatures_ok: true` confirms no statement text changed.
Commit shas (on the scratch branches, not `integrated`): `step_feeRatio` at `3e3375a`,
`live_leaf_matches` at `0494761`, `owned_transact_can_happen` at `7ed59b4`.

`owned_transact_can_happen`'s proof is not a copy of Sonnet's: Qwen derived `hRemove` from the
brief's `hLive` hypothesis and rebuilt the position-insertion step independently, where Sonnet's
version used `hLive` directly. Both proofs are closed, sorry-free, and pass every check; this is
a different valid derivation, not a verbatim match.

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

Qwen (`vllm/qwen3.8-27b-ablit`) proves every one of the three pool theorems tested, at every
size tried: a one-line frame lemma, a 13-line mid-size lemma, and a 61-line can-happen theorem.
The earlier "Qwen cannot write Lean" record (Phase 3 step 1: no Qwen Lean run exists on this
host; the one Lean brief written for Qwen in `reduction-plan.md` gave it only `bash`, `grep` and
`glob`, no `edit`) traces to a setup gap, not a model limit: with `read`/`edit`/`bash` and a
brief that names the exact file, line range and lemma, Qwen closes real pool-safety goals it
cannot have memorized (the proof texts above are not literal quotes of Sonnet's; Qwen reproduced
`step_feeRatio`'s tactic independently and wrote structurally different proofs for the other two).

A pool pebble should route to Qwen when three conditions hold together: the theorem is small to
mid-size (the two passing sizes here, judged by Sonnet's own proof length, not a line-count
guess), the brief names every lemma and line range the proof needs (no open-ended Mathlib
search), and a human or Sonnet review follows before the proof reaches a merge — this experiment
never ran through the gate container or a reviewer, and the can-happen proof's different
derivation path is exactly the kind of change a design-check reviewer should read, even though
the acceptance script passed. The one can-happen theorem tested here also passed, so can-happen
size alone is not yet grounds to withhold a pebble; a larger sample across more can-happen
theorems would sharpen that line. Peak context stayed well under the model's 146,816-token
window on all three (46% of window at worst), so context was not the limiting resource at this
theorem size.

None of the three Qwen proofs is merged into `armada-lean`. They sit on scratch branches
(`qwen-phase3-frame-step-feeratio`, `qwen-phase3-mid-live-leaf-matches`,
`qwen-phase3-canhappen-owned-transact`) in `armada-lean/.scratch/qwen-phase3/`, cut from
`integrated` at `03db4a5` and never pushed.
