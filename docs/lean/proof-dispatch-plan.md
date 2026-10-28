# Lean proof dispatch plan

Operator rulings, 2026-10-28:
- Lean work runs through exophial.
- The pool proofs run one at a time, on Sonnet.
- The Nitro proofs run on Sonnet, at the same time as the pool proofs.
- Mathlib is used where it makes sense.
- Experiments decide whether Qwen on the 5090 can write Lean. A bad system prompt was part of the earlier difficulty.

The brief format and the search rule are in `worker-briefs.md`. The checks are in `verification.md`.

## Phase 0. Set up exophial for Lean

No exophial instance exists for `armada-lean` today.

Decisions:
1. **One trunk, one instance.** The pool and Nitro specifications are merged into one trunk (merge `62197f4`). In the
   clone `~/code/git_puller/repos/armada-lean`, exophial owns `integrated`, and the fork's `main` follows it (operator,
   2026-10-28: "clean up all the branches … make sure that you're working in a consistent system"). The Nitro files
   import only Nitro files and Mathlib, so the merge does not couple them. The pool pebbles and the Nitro pebbles run
   in this one instance. Dependency edges keep the pool pebbles in order. `armada-nitro-spec` stays a separate branch
   until it is rebuilt on `integrated`.
2. **The gate is the Lean check.** `gate.command` runs `lake build`, then `scripts/check.sh`, which replaces the
   `.scratch/check.sh` of `verification.md` section 1. The script is in the tracked tree, because the gate container
   sees only tracked files. It checks: no new `sorry`, no new axiom, no change of an approved statement text, and the
   non-vacuity witnesses (`verification.md` §5a).
3. **The gate container.** `.exophial/gate/Dockerfile` installs elan and `leanprover/lean4:v4.33.1`, and runs
   `lake exe cache get` at build time. The gate then runs offline. The armada-sdk Dockerfile (`c2a04b1`) is the
   pattern for the base image and the log handling.
4. **The worker route.** `worker_model` is `anthropic/claude-sonnet-5` on the `omp` backend. The omp backend gives the
   worker the `task` tool (for `scout` subagents) and `web_search`, which the search rule needs.
5. **One pebble per theorem.** The pebble states the expectation: the named theorem is proved with no `sorry`, and the
   gate passes. Under the intake contract, the pebble states no remedy. The brief's search rule and design checks go
   into the spec's acceptance.

Acceptance of Phase 0: the gate passes in its container on `integrated`, and the intake dry-run of the first pool
pebble is accepted.

Result (2026-10-28, agent `LeanExophialSetup`):
- `armada-lean` is an exophial instance, prefix `aln`. The checkout is on `integrated`. The fork's `main` follows
  `integrated` through `.github/workflows/ff-main-from-integrated.yml`. Push to ship-armada is disabled.
- Config: `dispatcher.backend: omp`, `max_workers: 2`, default worker `claude-sonnet-5`,
  `gate.command: lake build && scripts/check.sh`, gate timeout 1800 s. The cold gate image builds in about 7 min.
  The gate itself runs in about 13 s.
- `scripts/check.sh` compares theorem statements with the committed `scripts/sigs-baseline.tsv`, because the gate
  container has no `.git`. A check against `git show` passed every change in the container, so it was replaced.
- A pebble's observation is `scripts/check.sh --json <theorem>`. It prints one JSON object with `sorry` (a list of
  names) and `theorem_under_test`. The predicate is
  `{"observe": "stdout_json", "op": "not_member", "elem": "theorem_under_test", "collection": "sorry"}`.
  The intake dry-run for `leaf_has_merkle_witness` with this predicate was accepted.
- A change under `scripts/` merges only by `exophial integrate --branch`. `exophial apply` refuses it as acceptance
  machinery.
- The supervisor's `ORACLE_ENV` has an `armada-lean` entry (`lake:lean:elan`, `.elan`). Its route is Sonnet.
- The Lean worker rules reach each worker through `.exophial/doctrine/worker_base.md` in the armada-lean checkout.
  exophial puts that file into every worker prompt (`dispatcher.py:947-950`, `doctrine/__init__.py:56-87`). The file
  is a copy of the wheel's doctrine plus a section "Lean proof work". `.gitignore` excludes it, so it is not in git:
  a new clone loses it, and a new exophial wheel's doctrine does not reach it.

Phase 1 filing (2026-10-28, agent `PoolPebbles`): the 15 pebbles `aln-9b1` … `aln-1d9` are filed in the order below.
Each pebble depends on the one before it. The dispatcher skips a pebble with an open dependency (`bus.ready_pebbles`
wraps `pb ready`, `bus.py:690-727`). All 15 are in the supervisor queue. Review each derived spec
(`contracts/specs/derived-<id>.spec.json`) before its predecessor merges.

## Phase 1. Pool proofs (serial, Sonnet)

Trunk `integrated`, pool files (from the old branch `pool-symbolic`, tag `pool-symbolic-approved-2026-10-28`). Every type is approved
(`labbook/2026-10-28-pool-symbolic-types-approval.md`). The order puts
each lemma before the theorems that use it. The uses come from the doc comments in `ArmadaLean/PoolSafety.lean`
(lines 777, 1527, 1679, 1758, 1858-1859). The order inside each group is `[INFERENCE]`.

| # | Theorem | Used by |
|---|---|---|
| 1 | `leaf_has_merkle_witness` | `owned_spend_witness_exists`, `owned_merge_can_happen` |
| 2 | `live_leaf_matches` | `owned_merge_can_happen` |
| 3 | `live_values_spendable` | `owned_merge_can_happen` |
| 4 | `owned_spend_witness_exists` | `owned_merge_can_happen` |
| 5 | `transact_unshield_exact` | the can-happen theorems |
| 6 | `transactOwned_records_all` | the can-happen theorems |
| 7 | `transact_can_happen` | — |
| 8 | `unshield_can_happen` | — |
| 9 | `owned_transact_can_happen` | — |
| 10 | `owned_spend_can_happen` (`PoolOwned.lean`) | — |
| 11 | `spend_not_blocked` | 12, 13 |
| 12 | `transact_not_blocked` | — |
| 13 | `transactOwned_not_blocked` | — |
| 14 | `spend_bound_to_adapt` | — |
| 15 | `owned_merge_can_happen` | — |

A pebble starts only after the one before it merges. A counterexample stops that theorem and goes to the operator
(`docs/process.md` 8.3). The remaining pebbles wait.

## Phase 2. Nitro proofs (Sonnet, at the same time as Phase 1)

Trunk `integrated`, Nitro files (from the old branch `nitro-proofs`, tag `nitro-proofs-approved-2026-10-28`). The open proofs are the Liveness theorems
in `ArmadaLean/Nitro/Liveness.lean`. The re-audit in `docs/nitro-repair-2026-10-27.md` §3 classifies each one as sound
or unknown. Start with the theorems marked sound. The order of the rest comes from the dependency of each proof.

Risks that the operator accepted by running Phase 2 now:
- The executable-operations review (handoff, next steps) can change the Nitro definitions that these proofs read.
- The liveness-refinement design (Opus agent, 2026-10-28) can change how the Liveness theorems are stated.
- 11 Nitro types were approved without a review of their Lean text.

In each case a proof that breaks is redone, and a counterexample goes to the operator.

## Phase 3. Qwen experiments

Question: can Qwen on the 5090 write Lean for this project, with a correct setup?

1. Read the system prompt and the tools that the earlier Qwen Lean runs received. Record what was wrong with them.
   Result of the search (2026-10-28): no written record exists in armada-laconic (`labbook`, `docs`, `engineering`,
   `notes`). The ingest-sessions index has no Lean session whose model is Qwen: 13 queries, and every armada-lean
   session found ran `claude-sonnet-5` or `claude-opus-*` (agent `QwenLeanHistory`). The one Lean brief written for
   Qwen, `reduction-plan.md`, gave it only `bash`, `grep` and `glob`: no scout and no web search. The earlier Qwen
   runs were probably outside the index, or were not Lean runs `[INFERENCE]`. The operator's account is the evidence.
2. Fix the setup: the omp backend with the Lean tools, the brief format of `worker-briefs.md`, and a scout for search.
3. Run Qwen on pebbles that Sonnet already solved, through the same exophial gate. Use 3 sizes: a one-line frame lemma,
   a mid-size lemma, and a can-happen theorem.
4. Record for each run: pass or fail at the gate, peak context, turns, wall time, and the failure mode.

Acceptance: a lab book entry with the measured results, and a ruling on whether Qwen takes Lean pebbles.
