# Qwen on the 5090 proves armada-lean theorems one step at a time through the Lean REPL (2026-11-18/19)

This entry continues `2026-10-30-qwen-lean-phase3.md`. The question: does the local model
(`vllm/qwen3.8-27b-ablit` on the RTX 5090) prove armada-lean theorems when a harness drives it one
step at a time through the Lean REPL, where whole-proof sampling did not? The operator's
hypothesis: the model fails because it writes whole proofs blind, and the REPL is the missing
piece. No exophial verb ran for this work.

## The test set

The 23 validated sorry-to-proof pairs of `exophial/.scratch/apollo/armada-lean-pairs-validated.json`
(22 distinct theorems; `turbo_value_is_value` appears twice). For each pair, the task commit holds
the theorem with `sorry`, and a later commit holds the proof. Opus (agent `RepairR2Ops`,
2026-11-01) proved the six `OpEquations` theorems; a Sonnet worker (`aln-f9c`, 2026-10-29)
proved `spend_bound_to_adapt`. The others came from Sonnet or Opus workers.

The task commits pin Lean v4.33.1. `.scratch/repl-spike/prepare_trees.sh` makes one worktree per
task commit and applies armada-lean's own move to v4.35.0 (`2de4240`: `lean-toolchain`,
`lakefile.lean`, `lake-manifest.json`). All 17 trees built with that change alone; none needed
lean-mend.

## The harness

`.scratch/repl-spike/spike.py`. Per pair, the REPL imports the built module, replays the scope
commands before the theorem, and restates it as `<name>_spike := by sorry`. Two controls run
first: the human proof must pass the check, and `sorry` must fail it. The model then gets the
statement, the file context, the current goals, the steps applied so far, and the steps Lean
rejected at this state, and returns one step: a `#check`/`#print` lookup, or a tactic. Lean runs
it. A depth-first search keeps every verified state and backs up after 4 rejections at a state.
The final check elaborates the restated theorem and reads `#print axioms` against armada-lean's
allowlist. Every model and REPL exchange is logged.

Two measures make the model usable at all:
- **Budget forcing** (Muennighoff et al., s1, 2025). The model's thinking did not end within
  4,096 or 16,384 tokens on these prompts. The harness caps it at 2,048 tokens, closes it, and lets
  the model continue into its answer, which takes about 1 s. With thinking off, the model
  degenerates into repeated `obtain` lines.
- **Lookups.** Without them, the thinking spends its budget guessing lemma signatures that one
  `#check` returns.

## Results

| Run | Pairs | Proved | Tactics sent | Accepted | Parse errors | Hours |
|---|---|---|---|---|---|---|
| Whole-proof sampling, 2026-11-17 (`baseline-k8.log`, k=8) | 22 | 2 theorems | — | — | — | 5.1 |
| Step search, first run | 23 | **6** | 373 | 83 | 179 | 2.32 |
| Rerun of the 17 unproved, after two harness fixes | 17 | **3** | 468 | 115 | 216 | 2.97 |
| Splitter (a multi-line reply sent as single tactics) | 13 | 0 | 469 | 138 | 147 | 2.50 |
| Item 1 (JSON reply, multi-line sent as one `(…)` block) | 13 | 0 | 323 | 51 | 176 | 2.19 |
| Items 2+3 (list of tactics, one goal at a time, best-first, 4 candidates per state) | 13 | **2** | 1,019 | 375 | 273 | 1.53 |

**11 of 23 pairs, 10 of 22 distinct theorems, proved by step search**, against 2 by whole-proof
sampling with the same model:

| Theorem | Steps | Seconds |
|---|---|---|
| `live_values_spendable` | 1 | 12 |
| `apply_conclude_equation` | 2 | 48 |
| `transactOwned_not_blocked` | 7 | 96 |
| `apply_challenge_equation` | 6 | 103 |
| `spend_bound_to_adapt` | 12 | 221 |
| `apply_checkpoint_equation` | 15 | 255 |
| `transactOwned_records_all` | 26 | 424 |
| `owned_transact_can_happen` | 20 | 433 |
| `turbo_value_is_value` (one of its two pairs) | 38 | 652 |
| `transact_unshield_exact` (items 2+3) | 20 calls | 367 |
| `owned_spend_witness_exists` (items 2+3) | 24 calls | 394 |

## Comparison with the frontier models

When the spike succeeds, it takes 12 s to 11 min (median about 4 min). On the one direct Sonnet
case, `spend_bound_to_adapt`, the Sonnet worker took 4.5 min (dispatch 03:04:30, commit 03:08:57)
and the spike 3.7 min. Opus proved all six `OpEquations` theorems within about 3 min of proof
writing; the spike proved three of them and failed the other three after 11 min each, in every
run. The frontier models proved all 22 theorems; the spike proved 8. The spike uses no plan
tokens.

## Harness defects found, each of which falsified a result until fixed

1. A buffered `readline` hid the REPL's replies from `select`, so every load timed out.
2. The scope pattern matched prose lines in doc comments, so 4 Liveness pairs did not load.
3. A forced answer often stops before its closing code fence; 25 valid steps were discarded.
4. The REPL's tactic mode accepts one tactic per request; 284 of 757 attempts were multi-line.
5. The REPL reports a failed tactic in the top-level `message` field, not in `messages`. Two
   analyses (the coordinator's and a research subagent's) first counted those failures as accepted.

## What the failures are

The 13 unproved pairs fall into two groups: equations that unfold a definition into cases
(`apply_deposit/transfer/reclaim_equation`) and long can-happen or witness constructions. In the
item-1 run, 163 of 176 parse errors were the model's own Lean errors inside multi-line tactics:
bullets indented under the tactic before them (80), invented tactics such as `first_goal` (18),
and unbalanced brackets (26). Two replies degenerated into `?_ ?_ ?_ …` until the token cap.

Of the 179 accepted tactics in the first two runs, 68 (38%) were a single choice that Lean can
enumerate (`unfold X`, `simp [..]`, `split`, `by_cases` on a condition in the goal); the rest
introduce new terms (`have` 42, `obtain` 15, `case` 13, `let` 10, `refine` 9).

## Training data built (no GPU training while vLLM serves)

- `.scratch/repl-spike/data/train-alltactics.jsonl`: 18,912 (goal, tactic) pairs from the
  REPL's `allTactics` over all 90 model files at armada-lean `ab5b476`, with the 22 test theorems
  excluded. `data/heldout-alltactics.jsonl`: the 751 pairs inside the test theorems.
- `data/search-triples.jsonl`: 1,220 (goal, tactic, verdict) records from the search logs.

## Item 4: a scorer for enumerable steps (`.scratch/repl-spike/scorer.py`)

At each theorem's root state, a `run_tac` lists the goal's `ArmadaLean` constants and the local
hypotheses. The candidates are 16 generic tactics plus `unfold`, `simp only [..]` and `rw [..]`
for each constant and `cases`, `subst` and `exact` for each hypothesis: 25 to 133 per state. Lean
runs every candidate; the model only scores each one, with `prompt_logprobs`, in the spike's
reply form. A candidate that closes the goal counts as progress only if the restated theorem then
passes the spike's axiom check. Measured on all 22 root states (`runs/scorer3.jsonl`, GPU not
shared):

| Ranking | Top-1 candidate makes real progress | Random-pick expectation |
|---|---|---|
| Mean log-probability per token | **17 of 22** | 5.99 |
| Summed log-probability | **16 of 22** | 5.99 |

Both rankings are about three times better than a random pick. Scoring took 6 to 65 s of model
time per state, for 25 to 133 candidates. The summed ranking loses 5 states to `exact?`, which
it ranks first and which fails the check. The per-token ranking misses on long constructions,
for example `simp only [Step]`, which leaves the goal of `transact_can_happen` unchanged.
"Progress" means that Lean accepts the tactic and the goal changes, which is a weak measure:
unfolding the head definition nearly always changes the goal.

Two defects in earlier versions of this measure. First, `exact?` closes every root goal by citing
the original theorem, which is still in the environment with `sorry`; the spike's axiom check
rejects such a proof, but the first scorer counted it as progress. Second, the base environment
had `open Lean Elab Tactic Meta`, whose names collide with project names such as `Term`, so 8 of
22 states did not load; the introspection now uses qualified names.

## Items 2 and 3: list-of-tactics replies, one goal at a time, best-first search

The reply is now `{"kind", "steps": [...]}`: the model's list decides where each tactic ends,
and the harness sends each element as one tactic. The grammar caps a tactic at 600 characters and
a step at 8 tactics. A first best-first run (`runs/bfs-defective`, 8 pairs) gave 11% parse errors
and 44% accepted tactics, against 54% and 16% for item 1. But its closing rule was wrong: a state
closed after 4 rejections, and one expansion of 4 candidates closed the root, so every pair
ended `exhausted` after 4 to 30 of its 40 calls and proved nothing. A state now closes after 4
expansions that produce no new state. The corrected run is `runs/bfs`.

The corrected run proved 2 of the 13 pairs that every earlier version failed:
`transact_unshield_exact` and `owned_spend_witness_exists`. Both are long constructions. The
first builds the whole post-step state record and closes the arithmetic with
`transact_balance`; the second builds the witness and the public input with `Classical.choose`
and closes the Merkle path facts with `treePath_correct_of_getElem?`. 7 pairs still ended
`exhausted` before their budget, after 6 to 28 calls, because every open state became barren.
The 4 answers sampled from one thinking phase are often identical, so a state gets fewer
distinct candidates than `n=4` suggests.

## Conclusions

1. The operator's hypothesis holds on this test set. With the same model, a harness that shows
   the goal after every step and keeps every verified state proves 10 of 22 theorems, where
   whole-proof sampling proved 2.
2. Most of the gains came from the harness, and most of the losses were harness defects. Five
   separate defects each falsified a result before they were found. A model that writes Lean
   needs a harness that hands every question about Lean text to Lean, and every question about
   the reply's shape to a grammar.
3. Where it succeeds, the step search is about as fast as a Sonnet worker per proof (median about
   4 min). It proves about half of the theorems that the frontier models proved, and it uses no
   plan tokens.
4. The model ranks enumerable candidate actions well (17 of 22 at the root against 6 by chance),
   so a scorer can make those steps with no thinking and no free text. About 40% of the accepted
   tactics in the search logs were such steps.
5. The remaining failures are long case splits and long constructions that the search abandons.
   The next measurements: the scorer inside the search for enumerable steps, more distinct
   candidates per state, and a LoRA fine-tune on the 18,912 training pairs.
