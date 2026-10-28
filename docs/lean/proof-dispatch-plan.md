# Lean proof dispatch plan

Operator rulings, 2026-10-28:
- Lean work runs through exophial.
- The pool proofs run one at a time, on Sonnet.
- The Nitro proofs run on Sonnet, at the same time as the pool proofs.
- Mathlib is used where it makes sense.
- Experiments decide whether Qwen on the 5090 can write Lean. A bad system prompt was part of the earlier difficulty.

The brief format and the search rule are in `worker-briefs.md`. The checks are in `verification.md`.

## Phase 0. Set up exophial for Lean

No exophial instance exists for `armada-lean` today. The checkouts `armada-lean-pool-symbolic` and
`armada-lean-nitro-proofs` have no `.exophial/` directory.

Decisions:
1. **One instance per line of work, in a new clone.** Make `armada-lean-exo-pool` and `armada-lean-exo-nitro`.
   Each new clone tracks one branch of the AFDudley fork (`pool-symbolic`, `nitro-proofs`). The existing checkouts
   stay untouched, so a mistake in the setup cannot damage them. Exophial owns `integrated` in each new clone.
   The clone's push configuration maps its trunk to the fork branch.
2. **The gate is the Lean check.** `gate.command` runs `lake build`, then `.scratch/check.sh` from `verification.md`
   section 1. The script moves from `.scratch/` into the tracked tree, because the gate container sees only tracked
   files. The script checks: no new `sorry`, no new axiom, no change of an approved statement text, and the
   non-vacuity witnesses (`verification.md` §5a).
3. **The gate container.** `.exophial/gate/Dockerfile` installs elan and `leanprover/lean4:v4.33.1`, and runs
   `lake exe cache get` at build time. The gate then runs offline. The armada-sdk Dockerfile (`c2a04b1`) is the
   pattern for the base image and the log handling.
4. **The worker route.** `worker_model` is `anthropic/claude-sonnet-5` on the `omp` backend. The omp backend gives the
   worker the `task` tool (for `scout` subagents) and `web_search`, which the search rule needs.
5. **One pebble per theorem.** The pebble states the expectation: the named theorem is proved with no `sorry`, and the
   gate passes. Under the intake contract, the pebble states no remedy. The brief's search rule and design checks go
   into the spec's acceptance.

Acceptance of Phase 0: in each new clone, one trivial pebble runs end to end and merges through the gate.

## Phase 1. Pool proofs (serial, Sonnet)

Branch `pool-symbolic`. Every type is approved (`labbook/2026-10-28-pool-symbolic-types-approval.md`). The order puts
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

Branch `nitro-proofs`, approved at tag `nitro-proofs-approved-2026-10-28`. The open proofs are the Liveness theorems
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
   No written record of those runs exists in armada-laconic (search of `labbook`, `docs`, `engineering`, `notes` on
   2026-10-28).
2. Fix the setup: the omp backend with the Lean tools, the brief format of `worker-briefs.md`, and a scout for search.
3. Run Qwen on pebbles that Sonnet already solved, through the same exophial gate. Use 3 sizes: a one-line frame lemma,
   a mid-size lemma, and a can-happen theorem.
4. Record for each run: pass or fail at the gate, peak context, turns, wall time, and the failure mode.

Acceptance: a lab book entry with the measured results, and a ruling on whether Qwen takes Lean pebbles.
