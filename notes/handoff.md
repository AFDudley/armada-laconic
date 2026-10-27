# Handoff — Armada build programme

This note holds the open state of the work: rules, what runs, what waits, and what is undecided. Remove an item when
it is resolved; git history keeps the record. Commit every change to `main`. A claim that the writer did not observe
has the mark [INFERENCE].

The plan is `engineering/build-plan.md`. The estimation method is `engineering/estimation-method.md` (1 point = 10,000
tokens of peak worker context; at most 9 points per 5090 chunk, 12 on any route). Decisions are in `labbook/`.

## 0. Programme state

### 0.0 Lean specification state, 2026-10-27 (read this first)

**The old adjudicator model is reference only.** `adjudicator-model-v4` (`23e01b5`) specifies one channel's dispute mechanics and payouts. It copies contract layout (the holdings slot), and it has no ledger or virtual channels. The 40 statements in its `README.md` were checked against the model, not derived from the papers. Do not extend it. Audit: `labbook/2026-10-26-adjudicator-layout-audit.md`.

**Branches in `github.com/AFDudley/armada-lean`:**

| Branch | Head | Contents | Proofs |
|---|---|---|---|
| `nitro-proofs` | `94436c3` | the Nitro specification, repaired; channel kinds `ledger`/`virtual` | 63 Safety proved; 4 of 34 Liveness proved, 30 `sorry`; 88 non-vacuity witnesses |
| `nitro-theorems` | `69b9ea2`, tag `nitro-theorems-approved-2026-10-27` | the 93 Nitro theorem types the operator approved, before the repair | all `sorry` |
| `pool-symbolic` | `3128085` | the pool model without contract layout (operator approved its changed types) | 15 `sorry` |
| `armada-nitro-spec` | `7085053` | the Armada layer (44 theorem types, `ArmadaLean/ArmadaNitro/`) on the APPROVED, not the repaired, Nitro spec; `docs/verification-chain.md` | all `sorry` |

Worktrees: `~/code/git_puller/repos/armada-lean-{nitro-proofs,nitro-theorems,pool-symbolic,armada-nitro-spec}`. Verify only in a separate tree, for example `armada-lean-nitro-theorems/.scratch/v` (detached; check out the branch to test, then `~/.elan/bin/lake build`). All four branches use toolchain `leanprover/lean4:v4.33.1` and the same Mathlib rev `0df444a3` (`lake exe cache get` first).

**In flight.** Nothing. The rename to `ledger` and the removal of the uncited theorems are at `94436c3`, checked in a separate tree (build green, 30 `sorry` in Liveness only, 88 witnesses with no `sorry`, the grep for "asset holding" empty).

**Waiting for the operator.** Approval of the repaired Nitro theorem types against the tag: about 90 changed (most mechanically, from rulings already given), 3 removed, new ones mostly lemmas. Evidence: `nitro-proofs` `docs/nitro-repair-2026-10-27.md` §4 and §6. Ask whether to review (a) by group or (b) only the substantive changes (Liveness hypotheses G1–G16, the 2ε clock bound, Turbo restricted to transfers).

**Next steps, in order.**
1. Operator approves the repaired Nitro types; tag the approved commit.
2. Review the executable operations of the Nitro and Armada layers for composition, typed state and footprints. They are the input of the programmatic Lean → Stele translation (`engineering/build-plan.md` "Verification chain"). Make the changes that this review finds before any more proofs.
3. Rebuild `armada-nitro-spec` on the repaired Nitro spec: the payee note key becomes Nitro's external-destination parameter `E`; channels follow the `ledger`/`virtual` kinds; apply the pending payee-key rulings (`payout_spendable`, `payment_note_only_payee`, no ciphertexts). Its 44 types then come to the operator.
4. Proofs: Nitro Liveness, the 15 pool `sorry`, then the Armada layer. Every proof task meets `docs/lean/verification.md` §5a (non-vacuity witness) and §6a (size and complexity).
5. File the armada-circuits need: owned circuits with more than one input (merges inside a channel).

**Rulings of 2026-10-26/27 that the next session must apply** (details in the labbook notes and `docs/lean/`):
- Source order: the latest paper first (SATP 2022, then Nitro 2019, then ForceMove 2018), then the statechannels team's go-nitro at `64c4ee9a1`. armada-nitro's vendored copy is cerc-io's fork `435eb2b`; do not use it as a source. Name sources by title, not S1–S4.
- Every theorem cites a paper passage, a contract line, or an operator ruling. No uncited theorems; arguments go in the labbook (`writing-lean.md` rule 11).
- Use the papers' terms. Do not invent type names. A different construction is a different type (channel kinds are `ledger` and `virtual`, SATP §1).
- No contract layout or encoding in the symbolic layer. Axioms may only declare primitives (`docs/process.md` 2.4).
- Use Mathlib wherever it applies.
- Model choice for subagents: Sonnet by default, Opus only when Sonnet fails at the task, no Fable. Keep the `fable` agent definition file. Evidence: the Fable agent that wrote `nitro-theorems` took the destination definition from the oldest paper (Nitro 2019), not from SATP and go-nitro. Its liveness game made 29 of 35 liveness theorems false or vacuous (`labbook/2026-10-27-nitro-liveness-game-defects.md`). Acceptance now requires a non-vacuity witness for every approved theorem (`docs/lean/verification.md` §5a).
- The operator decides theorem types. A defect whose fix the papers determine is the coordinator's to fix, not a question.
- Do not suggest following nitrolite. Do not propose pool contract changes until the Lean work is complete.

**Findings recorded this session:** go-nitro's `reclaim` lock, reproduced with `forge` (`labbook/2026-10-27-gonitro-reclaim-dos.md`); payout griefing and the payout rule (`labbook/2026-10-27-payout-griefing.md`); the liveness game defects (`labbook/2026-10-27-nitro-liveness-game-defects.md`); UTXO channel literature (`labbook/2026-10-26-utxo-channel-formalizations.md`); Bend2 to EVM (`labbook/2026-10-27-bend2-to-evm.md`).


### 0.1 Rules (operator)

- Lab book entries, not ADRs, record decisions for now.
- Every implementation pebble starts on the 5090. A worker that outgrows the window escalates to Sonnet, and the 5090
  takes the next pebble. At most 4 Sonnet workers and 5 workers in all run at once. When an Anthropic account passes
  60% of its 5-hour window, or its 7-day use grows faster than 0.6%/h, at most 2 Sonnet workers run.
- Circuit work (armada-circuits) runs on Sonnet.
- A husband agent owns each family. It reviews every merged diff, not only the gate result, and it owns timeouts,
  loops and parks on its pebbles. P1, P2, N1, N2 and G3 are audit-critical. There is no separate reviewer agent.
- Acceptance reuses the armada-nitro fixturenet, reconciler and gates (`env/gates/nitro-e2e`, `pool-e2e`,
  `watchtower`, `hub-custodian-frost`). Fixturenet first, then mainnet. Sepolia is out of scope.
- Nothing goes upstream. Upstream defects are recorded in `labbook/2026-10-20-nimbus-fixturenet-and-proxy-proof.md`
  §9.
- F3 is a complete rewrite of governance. It replaces the live armada-1 governance and the ARM token with the Baal
  design (`labbook/2026-10-18-treasury-governance-baal-safe-shielded.md`).
- Rules for brief writers: `engineering/chunk-rules.md`, `BRIEF-WRITER.md`, `ESTIMATION.md`
  (gitignored; copy into the repo if they must survive).

### 0.2 Documentation

- `engineering/build-plan.md`: S3 omits N9 (9 points); S5 uses the old N1–N8 sizes (43) instead of the split briefs
  (50).
- The site link to this note is broken: `notes/` is not rendered.

### 0.3 Exophial instances and dispatch

| Repo | Prefix | Backend, worker model, `max_workers` | Gate (`gate.command`) |
|---|---|---|---|
| armada-nitro | `arn` | `omp`, `pool/5090`, 5 | `pre-commit run --all-files` (forge build and test) |
| armada-poc | `arp` | `omp`, `pool/5090`, 5 | `npm ci --legacy-peer-deps --offline`, pre-commit, anvil-address check, hardhat compile, contract sizes, `tsc`, `forge test --offline`; timeout 300 s. `test:config` is still out of the gate, although `arp-d0f` merged. |
| armada-circuits | `arx` | `claude_code`, `claude-sonnet-5`, 1 | `npm ci --offline`, pre-commit (generated `contracts/specs/*.spec.json` excluded from prettier), compile of all shapes, circomspect; timeout 240 s. circom2 cannot read a symlinked `node_modules`. |
| armada-sdk | `ars` | `omp`, `pool/5090-local`, 1 | `npm ci --offline`, pre-commit, lint, typecheck, vendor typecheck, build, vitest, POI guard, vendor-divergence check; timeout 150 s. |

**Dispatch.** One host-wide supervisor, `.scratch/supervise/supervisor.py` (detached; log `supervisor.log`), is the
only caller of `exophial dispatch`. Husbands append reviewed, unparked pebbles to `.scratch/supervise/queue.jsonl` as
`{"repo": ..., "id": ...}`. Each tick it counts held claim locks, our leases on the 5090 in the omp router
(`local_agents/router`), and the quota from `exophial/.scratch/oauth_token.py --explain`. It reads each repo's route
from that repo's `worker_model` default:
- `pool/5090-local`: the 5090 with no cloud promotion. Starts whenever none of our workers holds the 5090, whatever
  the Sonnet count. A worker that outgrows the window compacts or fails, and the husband merges that chunk by hand.
- `pool/5090`: starts when none of our workers holds the 5090 and Sonnet is below its cap.
- any other model: starts when Sonnet is below its cap.

It sets `TMPDIR`, a per-repo `TMUX_TMPDIR` (to avoid `exo-fb55`), the token and the oracle variables. It logs every
start and every hold.

**Ephemeral reducers.** When a repo's ephemeral reducer still runs, the supervisor adds the pebble to that reducer's
scope with exophial's `bus.merge_scope`. The reducer reads `.exophial/ephemeral_scope.json` again on every tick.
`exophial dispatch` would only expedite the id onto the scoped reducer, and that reducer never claims it. The reducer
selects nothing while it gates a branch.

**Parked pebbles.** A parked pebble stays in the queue, held with reason `parked`. Its husband resolves it per the
doctrine. If the husband cannot, the coordinator finds the root cause, fixes it and gives the pebble back. The watcher
wakes the coordinator on each new parked hold.

**Faulty oracle.** A reducer that grades a pebble against a known-faulty oracle blocks the other pebbles of its repo
for hours. An example is the forge mutation scope fault `exo-b9b`. The coordinator stops that reducer with
`exophial stop`, checks the worktree for a left-over mutant (`git status`), and queues the other pebbles again.

**One 5090 session.** The 5090 serves one session (`local_agents/router/pools.json`, `maxSessions: 1`). `exophial4`
shares it through `pool/27b-class`. A second session does not fit: vLLM's fp8 KV cache holds 168,779 tokens
(`kv_cache_max_concurrency` 1.03 at full context), so two large contexts would evict each other.

**Watcher.** `.scratch/supervise/watch.py` (background job of the coordinator) wakes the coordinator when GPU 0
averages under 15% for 10 minutes, when an account passes the quota limits, and every hour. Samples:
`.scratch/supervise/samples.csv`. The 7-day rate is coarse: omp reports whole percents.

**Husbands** (doctrine `exophial/src/exophial/doctrine/repo/.claude/doctrine/husband.md`; logs
`<repo>/.scratch/plan/husbandry-<family>.md`): `HusbandF2c` owns F2 and E11. It replaced `HusbandF2b`, which replaced
`HusbandF2`, on 2026-10-22. `HusbandF3j` owns F3, armada-sdk and small armada-poc work. F1 and F4 have no open work.
Operator rule: husbands run on Sonnet, as the `task` agent.

**Husbands do not compact (operator rule, 2026-10-22).** A compacted husband loses its reasoning and keeps acting on a
summary. To replace one: tell it to stop new work and write a "SUCCESSOR HANDOFF" section at the end of its log (open
items with next actions, worktrees and jobs, learnings, facts later proved wrong); verify the section; then start a new
husband from a brief that points at that section. The first brief of this kind is
`armada-laconic/.scratch/husband-briefs/HusbandF2b.md`. Each brief tells the husband to keep its log current and, if
its own context compacts, to write the handoff section and yield at once.

**Merge paths.**
- A worker pebble: intake, `exophial pebble estimate <id> --points N` (`exo-df06`), spec review, queue.
- Critical-severity Solidity derives a spec with `forge_mutation_gated_test` (`exophial/docs/ORACLES.md` §515). If
  intake still escalates, the husband writes the change with red-first forge tests and merges it with
  `exophial integrate --branch <branch>`.
- An operator-known change outside `scripts/`, `tests/` and `contracts/specs/`: `exophial apply <branch>`.

**No protocol swap fee (2026-10-22).** Treasury lab book decision 23 (the N8 swap-fee edge) was an agent error with no ruling, and it is struck. T9 states that the only protocol fee is the shield-in fee. N8 is withdrawn, and the fee-tagged outputs of P2 and N2 go. armada-poc `b1679891` marks the swap fee superseded in its briefs, and armada-nitro `ec06acc0` strips it from N2b. The merged P2b defect (a plain transact output counts as a fee) is filed as `arp-fed`, open.

**Containerized gate (exophial `0.2.1+acc53c92`, ADR D49).** Each repo must track `.exophial/gate/Dockerfile`, with `.gitignore` exceptions `!.exophial/gate/` and `!.exophial/gate/Dockerfile`; without it every gate is `unevaluable`. armada-nitro has one (`5e2b8447`, merged at `cd8ceb50`: FROM exophial-e2e-base, node 24.18.0, rust 1.88.0, foundry v1.8.3, pre-commit 4.5.1, `fetch-submodules.sh` for offline lib/*). armada-poc has one on `integrated`. armada-sdk and armada-circuits have none on `integrated`; they need one before their next gated merge. Copy armada-nitro's pattern.

**Never pass `--git-dir` or `--work-tree` that points at a main checkout.** On 2026-10-21 a husband ran
`git --git-dir=../../.git checkout 6ada67cd --` from a scratch worktree, and the command changed the armada-nitro main
checkout. The lab book entries contract-owned-notes §3a and treasury §3a hold the revisions of 2026-10-22. F3 work
(G7, `arp-acf`, the crowdfund changes) must follow the revised entries.

**Fixturenet grading (operator rulings, 2026-10-25).** Interim rule: a recorded host run grades a fixturenet pebble, and `exophial apply` merges it. The commit quotes each fixturenet clause with the host command, tested sha and verdict lines. The original pebble is closed as superseded. Final design: the exophial integrator starts the pebble's laconic-so stack beside the gate container (D49 amendment in progress). R2 scenarios merged this way:
- `arn-c01` → `arn-8de` `2e8802a7`;
- `arn-ac6` → `arn-d8e` `a93213b4`;
- `arn-920` → `arn-40d` `33abb203`.
`arn-eb0` (tmpfs leak) merged at `5dcdbe93`. Open: the shared teardown for three R2 scripts, and `arn-5fb` (parked, see
"E11 paused").

**ingest_sessions.** `is-8c0` (Postgres supervision) is done and verified live. The ingest checkout is on `integrated`, and ingest pebbles run on the 5090 through the supervisor queue. `is-c52` waits for the exophial service-stack feature.

**E11 paused (operator ruling, 2026-10-26).** All E11 work waits until the Lean adjudicator work is complete. `arn-5fb` is parked (two failed 5090 attempts; `HubCustodian` is not reviewed); `arn-848` and `arn-12e` stay parked. Do not queue any E11 pebble.

**Timeouts are not tuned.** A timeout is a symptom; find the cause (operator rule). The coordinator raised armada-nitro `task_timeout` to 7200 s (`bf3e8b12`) and the operator corrected it; `488930d3` reverts it to 3600 s. The real cause was leftover fixturenets.

**Leftover fixturenets.** The gate scripts (`check_governance_stack.sh`, `check_pool_spend.sh`, the E11 and disclosure scripts) do not tear down their own fixturenet. On 21 Oct, 18 clusters were up (memory 156 GB, load 112), and fresh deploys stalled. The coordinator removed the 15 stale ones at 10:52 (`docker rm -f` by compose project, volumes, `sudo umount <dep>/data`). The husbands add an exit trap to each script. Before removing a cluster, map it to its deployment directory with the label `com.docker.compose.project.working_dir`, and keep clusters younger than the newest gate run.

**Withheld dispatch.** The dispatcher withholds a pebble whose title or description contains "blocked on", "do not dispatch" or a `Precondition:` line, unless it has a dependency edge (`exophial/src/exophial/precondition_hold.py`). The log line is "states an uncovered dispatch precondition". `arn-5c3` waited 40 min for this; `exophial pebble dep-add arn-5c3 arn-d36` released it at 12:44. Do not write "blocked on" in a pebble body unless a dependency edge exists. `exophial pebble grade-kept <id>` re-grades a kept worktree.

**Stale reducers.** A running reducer keeps the exophial code it loaded at start. After an exophial install, a gate that runs on a reducer started before the install uses the old oracle (armada-nitro `arn-e16` 22:09; armada-poc `arp-3e7` 00:4x). Compare `Starting reducer` in `.exophial/logs/dispatcher.log` with the install time before you trust a verdict. Do not stop a reducer that has a live worker; let it drain, then unpark.

**Push.** armada-poc's `fork` HTTPS remote uses a `gh` token without `workflow` scope, so a push that touches
`.github/workflows/` fails. Push over SSH: `git push git@github.com:AFDudley/armada-poc.git integrated:integrated`.

**Init.** `exophial init` takes no `-e` flag. Run the packaged play to set a prefix or languages:
`~/.local/share/uv/tools/exophial/bin/ansible-playbook <exophial>/ansible/init.yml -e exo_target_repo=<repo>
-e exo_pebble_prefix=<p> -e '{"exo_languages_override": ["typescript"]}'` (JSON form for lists).

### 0.4 Open chunks (209-point list)

| Family | In flight | Next |
|---|---|---|
| F1 pool (briefs armada-poc `docs/tasks/P*`) | — | — (P1, P2a, P2b, P3, P4 merged) |
| F2 zk-app (briefs armada-nitro `docs/tasks/N*`; N6 is Battleship; N9 reveals final balances at settlement) and E11 hub yield and swaps (`docs/tasks/E11-aave-lido-channels.md`; an additional hub option beside the ZK apps) | N4, N5 and N6 (circuits and apps) merged; the N6 app `arn-e16` merged by hand at armada-nitro `9addf467`. `arn-2fd`, `arn-ced` and `arn-a18` are closed. | N9 after N3. E11 is paused (see "E11 paused"). |
| F3 governance (briefs armada-poc `docs/tasks/G*`, `C6*`; husband `HusbandF3j`) | G1–G6, C6, the deploy-suite migration `arp-e44`, and the `baal-errors.md` items E1, E3 (part), E6 and E8 merged. armada-nitro pins `lib/armada-poc` at `db15af17`. | E2 `arp-6cc` and E5 `arp-445.1` are closed, but armada-poc `integrated` has no merge commit for them; the husband must find out why. `arp-acf` (E3 follow-up: the sum of all Safe notes the caller supplies) waits for an operator ruling. Holders cannot see the Safe's private notes, and concurrent exits spend the same notes. E7 (treasury ARM in the payout divisor) waits for an operator ruling and is not filed. The E4+E5 epic `arp-445` is open. G7: `arn-2be`, `arn-da4` and `arn-57b` are closed. G7b-spend `arn-d62` failed its gate on 2026-10-21; its worktree is kept. G7b-exit `arn-317` is open. Open: the replacement of the live armada-1 deployment and the move of ARM holders. |
| F4 collateral | — | Replaced by the Lean adjudicator rewrite (§0.0). armada-nitro `4c10a4de` deleted the N1a–N3, N7 and N8 briefs. `arn-451` and `arn-61e` are closed. |
| F5 client (C1–C5) | — | Needs N3. |
| Small | — | — |

Critical path: G7 (G3 merged).

### 0.4.1 What Qwen on the 5090 does (measured 2026-10-20)

- Finishes small, plain changes: `arp-133` (config and CI, peak 59k tokens).
- Escalates when the work outgrows the 146k window: G1 at about 228k, G2 at 373k.
- Loops on its own tooling: circom (about 120 retries on a JS harness) and a Solidity re-pin (`arn-4d9`, four runs;
  one wrote a 32.9 GB forge debug dump and timed out on it three times). It does not detect its own loop.
- On `pool/5090-local` (no promotion) it degrades near its window instead of escalating: G6 `arp-93a` ran 174
  turns to 116k tokens, then repeated one planning line and ended with an empty message. Chunks that read many files
  fit the local pool poorly.
- Not tried: Nim.

### 0.5 Proof path (nimbus verified proxy)

Lab book: `2026-10-19-nimbus-verified-proxy-proofs.md` (incl. §7), `2026-10-20-archive-state-for-nimbus.md`,
`2026-10-20-proxy-and-archive-open-questions.md`, `2026-10-20-nimbus-fixturenet-and-proxy-proof.md` (incl. §8).
- **The proof works by hand on the fixturenet:** through the proxy, with nimbus-eth2 as light-client source and
  nimbus-eth1 as data source, `eth_getBlockReceipts` proves a pool `Shield` log and `eth_getStorageAt` proves a
  `rootHistory` slot. It is not a gate job.
- **Proxy patches:** github.com/AFDudley/nimbus-eth1 branch `armada/verified-proxy-fixes` = `b0205d1d`
  (`827f95ff` convHeader `blockAccessListHash`/`slotNumber`; `b0205d1d` custom network directory), on upstream
  `2c8eabf`.
- **Lighthouse v8.2.2 is off spec** for genesis-finalized light-client updates (fills `finalized_header` where
  consensus-specs `v1.7.0-beta.4` `full-node.md` requires the empty header). With its "prefer older data" tiebreak it
  pins a bad period-0 update, so on the fixturenet the proxy's light-client source is nimbus-eth2. Upstream as
  sigp/lighthouse#9865 (open, not in v8.2.2).
- **Archive:** nimbus-eth1 serves no state below its base. A nimbus Aristo -> `ipld.blocks` emitter (4-5 points) would
  let `ipld-eth-server` serve historical `eth_getProof` for watchers. Fallbacks: geth v1.17 `--history.trienode`,
  Erigon v3.4+ `--prune.include-commitment-history`. Portal: Fluffy only; trin and Ultralight are idle.
- **Builds:** native, iOS and Android `libverifproxy.a` work (Android needs `-d:PREFER_HASHTREE_SHA256=false`). The
  WASM proxy stalls in a browser at the first `sync()`. Record and test page: `.scratch/SWasmBrowser/{REPORT.md,site2}`.
- **Open, handed off (operator orders):**
  1. *File a local issue for the WASM stall.* armada-client is the tracker with exophial, but `exophial intake`
     requires `--cites`, a document committed in armada-client that states the rule. No armada-client document names
     the verified proxy. The rule is in armada-laconic ADR-0009 D9.2 and `build-plan.md` T1-4. armada-client doctrine
     (`coordinator.md:73`) requires operator approval of the issue body. A dry-run draft is in session `01a14aa3`.
  2. *Per-log proofs: does the plan need them?* Operator: "probably not practically speaking. some sort of analysis
     should be done to inform the decision." Question: list every place the docs need an event proven (who proves, to
     whom, does that party have its own proxy), especially ADR-0020 disclosure packages; measure on mainnet the full
     block receipt set, a single-log Merkle proof, and the receipt-trie node storage per block; compare peer
     bandwidth. A stopped agent left a partial script and data at `.scratch/PerLogAnalysis/` (unchecked).
  3. *S7 gate.* `build-plan.md` S7 acceptance needs the proof as a job in armada-nitro's gates. No chunk schedules it.

## 1. Repositories

| Repo | Path | State |
|---|---|---|
| armada-nitro | `~/code/git_puller/repos/armada-nitro` | Checkout on `integrated`. A push to `origin integrated` runs `.github/workflows/ff-main-from-integrated.yml`, which fast-forwards `origin/main`. Exophial commits `.pebbles/events.jsonl` in its own `specify(…)` and `bus(…)` commits; do not commit it by hand. After a merge, if a modified file equals the pre-merge base, restore it with `git checkout -- <file>` (`exo-cba`). Never run `git stash`. |
| armada-laconic | `~/code/git_puller/repos/armada-laconic` | `main`. Push with `git push AFDudley.github.com HEAD:main`. The remote URL holds a `gho_` token; do not print it. |
| armada-poc | `~/code/git_puller/repos/armada-poc` | Checkout on `integrated`. Remotes `fork` (AFDudley) and `origin` (ship-armada). Push `integrated` over SSH (section 0.3). |
| armada-circuits | `~/code/git_puller/repos/armada-circuits` | Checkout on `integrated`. Branch `disclosure-circuit` = `7175b46` on remote `fork`, head of ship-armada/armada-circuits#3. |
| armada-sdk | `~/code/git_puller/repos/armada-sdk` | Checkout on `integrated`. Branch `selective-disclosure` = `0aa2ba1` on the fork, head of draft ship-armada/armada-sdk#142. |
| exophial | `~/code/git_puller/repos/exophial` | Checkout on `integrated`. Uncommitted edits that are not ours: `src/exophial/ops/CLAUDE.md`, `scripts/gates/invoker_reachability.json`. |
| mtm-deployment | `~/code/git_puller/repos/mtm-deployment` | Section 8. |

On ship-armada repositories we have read access only. Open pull requests from the forks. `gh pr edit` fails there with
a GraphQL "Projects (classic)" error; change a PR body with `gh api -X PATCH repos/ship-armada/<repo>/pulls/<n> -F
body=@<file>`.

## 2. Environment rules for every exophial command

- **Derivation transport.** `CLAUDE_CODE_OAUTH_TOKEN="$(python3 ~/code/git_puller/repos/exophial/.scratch/oauth_token.py)"
  exophial <verb> ...`. The helper picks the Anthropic account with the most headroom. Do not print the token.
- **No `/tmp`.** Set `TMPDIR=<repo>/.scratch` for `exophial dispatch`, `apply`, `integrate`, `intake` and `specify`.
  `/tmp` on kelce is a size-capped tmpfs.
- **Oracle variables.** The grader runs each oracle in a sealed HOME and reads these from the caller's environment.
  The supervisor sets them per repo (`supervisor.py` `ORACLE_ENV`): armada-nitro
  `EXOPHIAL_ORACLE_TOOLCHAIN_BINARIES=cargo:rustc:forge:cast:anvil:laconic-so`,
  `EXOPHIAL_ORACLE_DATA_DIRS=.cargo:.rustup:.foundry:.svm:deployments:src/stack-orchestrator`; armada-poc
  `forge:cast:anvil:node:npx:npm` and `.foundry:.svm`; armada-circuits `node:npx:npm:circomspect`. A binary in
  `node_modules/.bin` is not on PATH and must not be listed.
- **Reducer liveness.** `python3 ~/.claude/skills/exophial-reducer-liveness/scripts/reducer_liveness.py <repo>`.
- **Installed build.** `exophial self-update` installs the rolling `dev` release. Check with `exophial version`.
- **Worker models.** `pool/5090` and `pool/5090-local` both run `vllm/qwen3.8-27b-ablit` (one session, context window
  146,816 tokens). `pool/5090` promotes to `anthropic/claude-sonnet-5` on overflow; `pool/5090-local` does not.

## 3. Procedure for a pebble

1. Read the target tool's documentation before you file or run anything. For exophial, read `docs/AUTHORING.md` §1.1.
2. File with `exophial intake`: the expectation that must hold (a broken promise or a net-new capability), what holds
   now, how anyone can observe it with fields a check can assert on, and the constraints. State no cause, mechanism or
   remedy. Check with `--dry-run`, then file. Then `exophial pebble estimate <id> --points N`.
3. File only on an explicit operator order. A question is not an order.
4. Before you file, search the target repo's bus for a duplicate, and state the search scope of every absence claim.
5. Review the derived spec before any worker starts. A spec is graded from the base commit. A comment added after
   derivation does not change the spec; close and file again instead.
6. Merge by the paths in section 0.3.
7. A parked pebble goes to its family's husband, never an unchanged unpark. The husband changes the precondition, then
   queues it through the supervisor.

## 4. armada-nitro

### 4.1 Work outside the 209-point families

1. **E11, hub yield and swaps:** paused (section 0.3, "E11 paused"). Owner `HusbandF2c`.
2. **E17, Nitro specification conformance** (`docs/tasks/E17-nitro-spec-conformance.md`). Not scheduled. Section 4.5.
3. **An external audit** of `ArmadaNitroAdjudicator` and `HubCustodian`, last. ADR-0016 records that the risk of this
   construction is "audit, not code".

`arn-315` (the selective-disclosure checker) is closed. `arn-2fd` and `arn-12e` superseded it on 2026-10-20, and each
run now starts its own fixturenet with `stack/scripts/fixturenet-new`.

### 4.2 E16 open items

- **The hub's close script** pays with `concludeAndTransferAllAssets` (`script/CustodianFrostScenario.s.sol:136`). Its
  channel holds only USDC, so nothing fails today. Once E11 adds a second asset, one frozen asset (for example USDC
  under Circle's blocklist) blocks every payout of that close. The fix is part of E11.
- **E16's scope statement about E11 contradicts itself.** The header says that E16 prepares the adjudicator for E11's
  multi-asset channels, but "Out of scope" lists only "the hub's Aave deposit/redeem operations (E11)", and the
  acceptance section names only `armada-nitro-e2e`.
- Optional: a mutation check per E16 test, to show that each test can fail. Only the SafeERC20 test is proven able to
  fail.

### 4.3 Known facts

- `test/WatchtowerPerAssetPayoutCheck.t.sol:142-146` skips 4 failure-path tests under `forge test`. They run through
  `scripts/check-{respond,settlestale}-per-asset-payout.sh`.
- `scripts/run-armada-custodian-e2e-check.py:139-151`: the script's exit code covers seven fields, not
  `payout_matches_allocations` or `is_active_delegate`. The arn-77e spec judges both as runner predicates.
- `stack/compose-jobs/docker-compose-armada-custodian-e2e.yml:171` mounts `${FROST_SIGNER_BIN}` with no default. A bare
  `laconic-so deployment run-job armada-custodian-e2e` exits 1. Run it through the check script.
- `arn-7f8` (S6 gate script prints a constant) is parked. Operator decision: leave it.

### 4.4 Fixturenet

- **No shared fixturenet.** On 2026-10-27 `~/deployments/armada-fixturenet-e12` does not exist, and no `e12` container
  runs. Gates start their own fixturenet with `stack/scripts/fixturenet-new <name>`.
- When geth halts on a full 3G RAM disk (the design, `stack/README.md:84-85, 93-95`), with `D=~/deployments/<name>`:
  `laconic-so deployment --dir $D stop`, `sudo umount $D/data` once (the RAM disk; a hidden boot-time tmpfs below it
  stays), then `stack/scripts/fixturenet-recover <name>`. A stray process on 127.0.0.1:8545 (for example a worker's
  `anvil`) blocks geth's port.
- `fixturenet-new` maps ports with `localhost-fixed-random`; jobs reach geth at `http://armada-fixturenet-geth:8545`.
  Read a host port with `docker port <geth-ctr> 8545`.
- `run-job` needs `ARMADA_IMAGE_TAG=<full sha of the deployment's build>`, and reads the deployment's own copy of each
  job file.
- The env image takes `armada-poc` from `env/sources.lock.json`, not from the `lib/armada-poc` submodule. A pin change
  must update both.
- The hidden boot-time tmpfs mount on `$D/data` (from `arn-faa`) stays until the next reboot. Do not unmount it.
- Cleanup: worktrees under `.worktrees/` from earlier workers (`worker-arn-4d9/.scratch/dbg.json` is 32.9 GB);
  `backup/main-before-reset-2026-10-13`, `-13b`, `-13c`, `-14`; 26 `wip/arn-*` branches; `hand/e16-2026-10-14`.

### 4.5 E17, Nitro specification conformance

- Test rule (operator): a clause the code meets is proven by a mutation that breaks it; a clause the code does not
  meet gets a red commit first.
- **The papers are in armada-nitro `docs/references/nitro/sources/`** (committed, sha256 in that `README.md`):
  ForceMove 2018 `forcemove-2018.pdf`, Nitro 2019 `nitro-2019-219.pdf`, SATP 2022 `satp-2022.pdf`, and the
  statechannels docs `docs.statechannels.org/*.html`. Name them by title, not S1–S4. eprint.iacr.org blocks `curl`
  (Cloudflare), so do not download the Nitro 2019 paper again. The Railgun whitepaper and audits are in armada-laconic
  `.scratch/rewrite-B/sources/`.
  armada-nitro `docs/references/nitro/clauses.md` lists 31 clauses, 8 differences and 7 silences.
- No pebble. A defect is filed only when a test shows a clause broken. Candidates, not confirmed by a test: the
  adjudicator's `_executeSingleAssetExit` ignores a `withdrawHelper` allocation (clause 31, Sil4); `reclaim` credits
  only target allocations 0 and 1 (Sil6).
- `notes/contract-language.md` records the decision to stay with Solidity for machine-authored code. hevm is the chosen
  symbolic tool.

## 5. Selective disclosure (armada-laconic, armada-circuits, armada-sdk)

- **Design:** page 16 and ADR-0020 (clauses D20.1–D20.6) in armada-laconic.
- **Circuit:** ship-armada/armada-circuits#3, ready for review. `EdDSAPoseidonVerifier` is a trusted circomlib
  component; a Lean proof of the circuit against page 16 is planned at low priority.
- **SDK:** draft ship-armada/armada-sdk#142. It depends on #3.
- **Open gaps:** no hop has been checked against events from real transactions, and the hub's EIP-1271 check has run
  only against a stand-in. `arn-12e` covers both; it is parked (section 0.3, "E11 paused"). The co-signed channel
  states need a recorded source in that run.
- **The disclosure-circuit setup is a dev setup.** The proof system and ceremony for production are open (page 16).

## 6. Designs discussed, not written

### 6.1 How a hub tracks cross-chain swaps

- The design (`engineering/shielded-nitro-bridge-design.md` §4–§6, ADR-0015) declares hub inventory, rebalancing and
  pricing "out of protocol", at most an open-source "reference hub daemon". No such daemon or design page exists.
- A cross-chain swap uses one channel on each chain, each with one asset, linked by a hash lock (go-nitro
  `HashLockedSwap.sol`). A single-chain swap is one co-signed state of one channel with two assets (ADR-0017).
- What a hub would have to track [INFERENCE]: the latest co-signed state of each channel on each chain; each swap as a
  pair of conditional payments with its hash and whether the secret is revealed; its inventory and promised amounts
  per asset per chain; its net imbalance between chains; and every open challenge against its channels.

## 7. exophial

### 7.1 Open filings

| Pebble | State |
|---|---|
| `exo-9b6da` (integrate merges a pebble whose linked spec file is absent), children `.1` and `.2` | P1. `arn-b27` is one more instance. |
| `exo-dfa` (base replay records `red_on_base` for a metamorphic check whose script is missing on the base) | P2, spec derived. `arn-3d8`, `arn-521`, `arn-00d`, `arn-40f` are more instances. |
| `exo-36ed` (`docs/AUTHORING.md` §1.1.1 names a deleted function) | P3 |
| `exo-b31`, `exo-ac16`, `exo-cba`, `exo-0a7`, `exo-f8b`, `exo-8105` | From earlier sessions. |

### 7.2 Acceptance test, e2e test and outer gate

`exophial/acceptance-e2e-outer-gate.md` (untracked) is correct and complete: the documents use "acceptance" for three
things with different red rules, and base replay applies one rule to all. It is not committed and not filed.

### 7.3 Grading model

- The worker creates each probe script named in `declared_new_entrypoints`. The spec fixes the `then` predicates over
  the probe's output, and the runner owns the verdict (`exophial/docs/ORACLES.md` §205–232, `spec.schema.json:192`).
- Base replay overlays each probe and its local support files onto the base tree, then grades the base implementation
  with the branch's probe (`base_replay.py:19-39`). `_is_implementation_file` (`base_replay.py:954-961`) matches only
  `src/exophial/*`.

### 7.4 Observed, not filed

- After unpark, `exophial dispatch` reuses the kept worker branch; it does not cut a fresh branch from `integrated`.
  A note added on unpark does not change the result (seen on `arn-781` and `arx-799`, `HusbandF2`). To change a failed
  pebble's precondition, hand-fix its branch and integrate it, or close it as superseded and file it again.
- While a repo's ephemeral reducer runs, `exophial dispatch <new id>` expedites the id onto it; the reducer is scoped
  to its own ids, so the new id is never claimed (lost: `arp-c7a`, `arp-bb3`, `arp-d0f`, re-queued). The supervisor
  dispatches into a repo only when its reducer lock is free.

## 8. mtm-deployment and mtm production

- **The cron line is disabled.** Backup: `armada-laconic/.scratch/crontab.before-2026-10-16`.
- **mtm production is behind.** The reconcile failed 31958 times at "Converge the gateway surface to the promoted
  ref". Live gateway `7eeb2471`, live SPA `88959b89`, desired ref `bb2367bf`. Cron ran from a checkout 21 commits
  behind `origin/main`. Whether that caused the failures is not checked [INFERENCE].
- Before the cron line is enabled again, the checkout it runs from must follow `origin/main`.

## 9. Questions to the operator

- `supervisor.py` replaces scheduling that exophial lacks (host-wide limits, quota, local versus cloud routing). Move
  it into exophial as a feature? (`exo-8f0`, a related filing, merged.)
- Give the S7 proof gate job (section 0.5) to a husband?
- mtm production: who owns the failing reconcile, and when is the cron line enabled again?
- Prune dangling Docker images? Remove marks from `pool/27b-class` and `pool/long`? The drive swap on the `kelce-qvos`
  pool (mirror-0 `sdc` …01064Z with mirror-2 `sdf` …00043M).

## 10. Known failure pattern of the assistant

The assistant again and again:
- used a tool without reading its documentation first;
- stated a rule or a number that it had not measured (for example a 3-point limit for Qwen);
- filed solution-shaped reports instead of the broken promise;
- stated that something does not exist after a partial search;
- wrote inferences as observations, and used passive voice that hid who acts;
- acted beyond the approved scope;
- merged by hand outside exophial;
- ran long commands in the foreground and blocked the operator;
- presented options as equal when one was clearly correct.

Fixes that work do not depend on the assistant: an adversarial reviewer that must approve each filing, and the
operator asking "what did you read, what did you search, and where?".
