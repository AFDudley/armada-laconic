# Operating rules for the build programme

This document holds the operator rules and the procedures that are not written in another document. It does not record
state. The handoff (`notes/handoff.md`) records only the work in flight, the next steps and the open questions.

## Rules written elsewhere

Read these documents. This file does not repeat them.

| Subject | Document |
|---|---|
| The 5090 first, escalation to Sonnet, at most 4 Sonnet sessions, husband review, audit-critical chunks, acceptance on the armada-nitro gates | `engineering/chunk-rules.md` |
| Points, the 9-point chunk limit, the 12-point limit on any route | `engineering/estimation-method.md` |
| Lab book entries, not ADRs, record decisions; Sepolia is out of scope | `engineering/build-plan.md` |
| Nothing goes upstream; the upstream defects | `labbook/2026-10-20-nimbus-fixturenet-and-proxy-proof.md` §9 |
| Lean source order, source names by title, citation of every theorem, axioms only for primitives, statement changes only with operator approval | `docs/lean/writing-lean.md` §10–§11, `docs/process.md` 2.4 and 8.3 |
| Non-vacuity witness, proof size limits | `docs/lean/verification.md` §5a, §6a |
| E17 test rule (a mutation for a met clause, red first for a broken clause) | armada-nitro `docs/tasks/E17-nitro-spec-conformance.md` |
| Fixturenet start, recovery, ports, `sources.lock.json` | armada-nitro `stack/README.md`, `docs/ENVIRONMENT.md` |
| The gate Dockerfile in each repo | exophial `docs/REPO_INIT.md` "The gate Dockerfile" |
| Filing with `exophial intake`, the duplicate check, `exophial integrate --branch` | skill `exophial-defect-forking`, exophial `docs/AUTHORING.md` |
| Reducer liveness and stop | skill `exophial-reducer-liveness` |
| Timeouts are a symptom, not a setting | `~/.claude/CLAUDE.md` "Technical Discipline" |

## 1. Lean specification work (operator rulings, 2026-10-26/27)

- Use the terms of the papers. Do not invent type names. A different construction is a different type. Example: the
  channel kinds are `ledger` and `virtual` (SATP §1).
- The symbolic layer contains no contract layout and no encoding.
- Use Mathlib wherever it applies. The "no Mathlib" pins in `docs/lean/README.md` and `docs/lean/writing-lean.md`
  describe the old `adjudicator-model-v4` branch.
- The operator decides theorem types. If the papers determine the fix for a defect, the coordinator fixes it and does
  not ask.
- Do not suggest nitrolite.
- Do not propose pool contract changes until the Lean work is complete.
- eprint.iacr.org blocks `curl` (Cloudflare). The papers are in armada-nitro `docs/references/nitro/sources/`. Do not
  download them again. The Railgun whitepaper and audits are in armada-laconic `.scratch/rewrite-B/sources/`.
- All armada-lean branches use toolchain `leanprover/lean4:v4.33.1` and one Mathlib rev. Run `lake exe cache get`
  before `~/.elan/bin/lake build`.

## 2. Models

- Subagents: Sonnet by default. Use Opus only when Sonnet fails at the task. Do not use Fable. Keep the `fable` agent
  definition file. Evidence: `labbook/2026-10-27-nitro-liveness-game-defects.md`.
- Circuit work (armada-circuits) runs on Sonnet.
- Husbands run on Sonnet, as the `task` agent.
- At most 5 workers run at once. When an Anthropic account passes 60% of its 5-hour window, or its 7-day use grows
  faster than 0.6%/h, at most 2 Sonnet workers run.
- The 5090 serves one session (`local_agents/router/pools.json`, `maxSessions: 1`). `exophial4` shares it through
  `pool/27b-class`. vLLM's fp8 KV cache holds 168,779 tokens, so two large contexts evict each other.
- `pool/5090` and `pool/5090-local` both run `vllm/qwen3.8-27b-ablit` (window 146,816 tokens). `pool/5090` promotes to
  `anthropic/claude-sonnet-5` when the window overflows. `pool/5090-local` does not promote. A worker on
  `pool/5090-local` that outgrows the window fails, and the husband merges that chunk by hand.

### What Qwen on the 5090 did (measured 2026-10-20)

- It finished small, plain changes: `arp-133` (config and CI, peak 59k tokens).
- It escalated when the work outgrew the window: G1 at about 228k tokens, G2 at 373k.
- It looped on its own tooling and did not detect the loop: circom (about 120 retries on a JS harness) and a Solidity
  re-pin (`arn-4d9`, four runs).
- On `pool/5090-local` it degraded near its window: G6 `arp-93a` ran 174 turns to 116k tokens, then repeated one
  planning line. Chunks that read many files fit the local pool poorly.

## 3. Husbands

- A husband owns each family. It owns the timeouts, loops and parks of its pebbles. The doctrine is exophial
  `src/exophial/doctrine/repo/.claude/doctrine/husband.md`. The log is `<repo>/.scratch/plan/husbandry-<family>.md`.
- A husband does not compact (operator rule, 2026-10-22). A compacted husband loses its reasoning and acts on a summary.
- To replace a husband:
  1. Tell it to stop new work.
  2. Tell it to write a "SUCCESSOR HANDOFF" section at the end of its log: open items with next actions, worktrees and
     jobs, learnings, and facts later proved wrong.
  3. Verify the section.
  4. Start a new husband from a brief that points at that section. Example brief:
     `.scratch/husband-briefs/HusbandF2b.md`.
- Each brief tells the husband to keep its log current. It also tells the husband to write the handoff section and
  yield at once if its own context compacts.
- A parked pebble goes to its husband, never to an unchanged unpark. The husband changes the precondition, then queues
  the pebble. If the husband cannot resolve it, the coordinator finds the root cause, fixes it and gives the pebble back.

## 4. Dispatch

- The supervisor `.scratch/supervise/supervisor.py` is the only caller of `exophial dispatch`. Its docstring states
  the routes and the start rules. Its log is `.scratch/supervise/supervisor.log`.
- Husbands append reviewed, unparked pebbles to `.scratch/supervise/queue.jsonl`, one object per line:
  `{"repo": "armada-nitro", "id": "arn-4d9"}`.
- Do not run `exophial dispatch` by hand. While a repo's ephemeral reducer runs, `exophial dispatch <new id>` sends
  the id to that reducer. The reducer serves only its own scope, so it never claims the new id (lost: `arp-c7a`,
  `arp-bb3`, `arp-d0f`).
- The watcher `.scratch/supervise/watch.py` wakes the coordinator when GPU 0 averages under 15% for 10 minutes, when an
  account passes the quota limits, and every hour. It is a background job of the coordinator session, so it stops
  when that session ends.
- Do not write "blocked on", "do not dispatch" or a `Precondition:` line in a pebble unless a dependency edge exists.
  The dispatcher withholds such a pebble (`exophial/src/exophial/precondition_hold.py`; log line "states an uncovered
  dispatch precondition"). `exophial pebble dep-add <id> <dep>` adds the edge.
- After unpark, `exophial dispatch` reuses the kept worker branch. A note added on unpark does not change the result.
  To change the precondition of a failed pebble, fix its branch by hand and integrate it, or close it as superseded
  and file it again.
- A reducer keeps the exophial code that it loaded at start. Before you trust a verdict after an exophial install,
  compare `Starting reducer` in `.exophial/logs/dispatcher.log` with the install time. Do not stop a reducer that has
  a live worker. Let it drain, then unpark.
- A reducer that grades against a known-faulty oracle blocks the other pebbles of its repo (example: `exo-b9b`). Stop
  it with `exophial stop`. Check its worktree for a left-over mutant with `git status`. Queue the other pebbles again.

## 5. Filing and merging

- File only on an explicit operator order. A question is not an order.
- State the search scope of every claim that something does not exist.
- A spec is graded from the base commit. A comment added after derivation does not change the spec. Close the pebble
  and file it again.
- Merge paths:
  - a worker pebble: intake, estimate, spec review, queue;
  - critical-severity Solidity: the spec derives with `forge_mutation_gated_test` (exophial `docs/ORACLES.md`). If
    intake still escalates, the husband writes the change with red-first forge tests and merges it with
    `exophial integrate --branch <branch>`;
  - an operator-known change outside `scripts/`, `tests/` and `contracts/specs/`: `exophial apply <branch>`.
- Fixturenet grading (operator ruling, 2026-10-25), interim: a recorded host run grades a fixturenet pebble, and
  `exophial apply` merges it. The commit quotes each fixturenet clause with the host command, the tested sha and the
  verdict lines. The coordinator closes the original pebble as superseded.

## 6. Environment for exophial commands

- Token: `CLAUDE_CODE_OAUTH_TOKEN="$(python3 ~/code/git_puller/repos/exophial/.scratch/oauth_token.py)" exophial ...`.
  The helper picks the Anthropic account with the most headroom. Do not print the token.
- Set `TMPDIR=<repo>/.scratch` for `exophial dispatch`, `apply`, `integrate`, `intake` and `specify`.
- The grader runs each oracle in a sealed HOME and reads the oracle variables from the caller's environment. The
  supervisor sets them per repo (`ORACLE_ENV`):

  | Repo | `EXOPHIAL_ORACLE_TOOLCHAIN_BINARIES` | `EXOPHIAL_ORACLE_DATA_DIRS` |
  |---|---|---|
  | armada-nitro | `cargo:rustc:forge:cast:anvil:laconic-so` | `.cargo:.rustup:.foundry:.svm:deployments:src/stack-orchestrator` |
  | armada-poc | `forge:cast:anvil:node:npx:npm` | `.foundry:.svm` |
  | armada-circuits | `node:npx:npm:circomspect` | — |

  A binary in `node_modules/.bin` is not on PATH. Do not list it.
- `exophial init` takes no `-e` flag. To set a prefix or languages, run the packaged play:
  `~/.local/share/uv/tools/exophial/bin/ansible-playbook <exophial>/ansible/init.yml -e exo_target_repo=<repo>
  -e exo_pebble_prefix=<p> -e '{"exo_languages_override": ["typescript"]}'` (JSON form for lists).

## 7. Git and repositories

All repositories are under `~/code/git_puller/repos/`.

- Never pass `--git-dir` or `--work-tree` that points at a main checkout. On 2026-10-21 such a command, run from a
  scratch worktree, changed the armada-nitro main checkout.
- Never run `git stash` in armada-nitro.
- armada-nitro: exophial commits `.pebbles/events.jsonl` in its own commits. Do not commit it by hand. After a merge,
  if a modified file equals the pre-merge base, restore it with `git checkout -- <file>` (`exo-cba`).
- armada-laconic: commit to `main`. Push with `git push AFDudley.github.com HEAD:main`. The remote URL holds a token. Do
  not print it.
- armada-poc: the HTTPS remote `fork` uses a `gh` token without `workflow` scope. A push that touches
  `.github/workflows/` fails. Push over SSH: `git push git@github.com:AFDudley/armada-poc.git integrated:integrated`.
- ship-armada repositories give read access only. Open pull requests from the forks. `gh pr edit` fails there with a
  GraphQL "Projects (classic)" error. To change a PR body, use
  `gh api -X PATCH repos/ship-armada/<repo>/pulls/<n> -F body=@<file>`.

## 8. Fixturenets

- A gate script must tear down its own fixturenet with an exit trap. On 2026-10-21, 18 leftover clusters used 156 GB
  of memory, and fresh deploys stalled.
- To remove a leftover cluster:
  1. Map it to its deployment directory with the label `com.docker.compose.project.working_dir`.
  2. Keep clusters younger than the newest gate run.
  3. Remove it with `docker rm -f` by compose project, then remove its volumes and run `sudo umount <dep>/data`.
- Recovery after geth halts on a full RAM disk, with `D=~/deployments/<name>`:
  1. `laconic-so deployment --dir $D stop`.
  2. `sudo umount $D/data` once. A hidden boot-time tmpfs below it stays until the next reboot. Do not unmount it.
  3. `stack/scripts/fixturenet-recover <name>`.

  A stray process on 127.0.0.1:8545, for example a worker's `anvil`, blocks geth's port.
- `run-job` needs `ARMADA_IMAGE_TAG=<full sha of the deployment's build>`. It reads the deployment's own copy of each
  job file.
- The env image takes armada-poc from `env/sources.lock.json`, not from the `lib/armada-poc` submodule. A pin change
  updates both.
- A bare `laconic-so deployment run-job armada-custodian-e2e` exits 1, because the compose file mounts
  `${FROST_SIGNER_BIN}` with no default. Run it through `scripts/run-armada-custodian-e2e-check.py`.

## 9. Gates

- circom2 cannot read a symlinked `node_modules`. This applies to the armada-circuits gate.

## 10. Known failure pattern of the assistant

The assistant often did these things:

- used a tool without reading its documentation first;
- stated a rule or a number that it had not measured;
- filed solution-shaped reports instead of the broken promise;
- stated that something does not exist after a partial search;
- wrote inferences as observations, and used passive voice that hid who acts;
- acted beyond the approved scope;
- merged by hand outside exophial;
- ran long commands in the foreground and blocked the operator;
- presented options as equal when one was clearly correct.

Two controls work, and neither depends on the assistant: an adversarial reviewer that must approve each filing, and
the operator asking "what did you read, what did you search, and where?".
