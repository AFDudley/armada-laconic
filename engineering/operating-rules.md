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
- Use Mathlib where it makes sense (operator ruling, 2026-10-27). All specification branches require Mathlib
  `v4.33.1`. Before a worker writes a general definition or lemma, it searches Mathlib (`docs/lean/worker-briefs.md`).
- Lean work runs through exophial (operator, 2026-10-28). Workers run on Sonnet. The pool proofs run one at a time,
  and the Nitro proofs run at the same time as the pool proofs. Qwen on the 5090 had difficulty with Lean, partly
  because of a bad system prompt. Experiments decide whether Qwen can write Lean before any Lean work goes to it.
- The operator decides theorem types. If the papers determine the fix for a defect, the coordinator fixes it and does
  not ask.
- A theorem type sent for approval carries its state domain: "reachable" (and which predicate) or "every state,
  because …". The operator rules on the domain with the type. Reachability is never added after proving starts
  (operator ruling 2026-10-29; `docs/lean/writing-lean.md` §11; gate `scripts/reach.py` in armada-lean).
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
- Local-model subagents use the `build-pool` agent (`pool/27b-class`; operator, 2026-10-30). The `sonic` agent has
  no model override in `~/.omp/agent/config.yml`, so it runs on the session default (Opus). 2026-10-30: four
  "Qwen" cleanup steps ran on Opus as `sonic`; their session files show `claude-opus-5-5`, and vLLM logged 0
  requests. Confirm local use from the vLLM journal (`journalctl -u vllm@qwen3.8-27b-ablit | grep 'POST
  /v1/chat/completions'`), not from one `nvidia-smi` sample: GPU use reads 0% while the agent waits on a build.

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
- `exophial self-update` restarts the repo's reducer with the caller's environment. A plain shell has no
  `EXOPHIAL_ORACLE_*` variables and no `~/.elan/bin` on `PATH`, so the oracle cannot find `lake`. On 2026-10-30 such
  a restart parked `aln-2ed`, whose proof was complete ("cannot resolve these declared binaries via shutil.which").
  Run the update with the repo's `ORACLE_ENV` and `PATH_PREFIX` from `supervisor.py`, or stop the reducer after the
  update and let the supervisor start the next one.
- A reducer that grades against a known-faulty oracle blocks the other pebbles of its repo (example: `exo-b9b`). Stop
  it with `exophial stop`. Check its worktree for a left-over mutant with `git status`. Queue the other pebbles again.

## 5. Filing and merging

- File only on an explicit operator order. A question is not an order.
- State the search scope of every claim that something does not exist.
- A spec is graded from the base commit. A comment added after derivation does not change the spec. Close the pebble
  and file it again.
- A filing body states observations only. It names no predicate, no field to compare and no oracle shape: the spec
  derivation designs the check. `--cites` quotes a committed rule, not a declaration line. Lean example: cite
  `docs/how-to-verify.md` ("A proof can be faked with `sorry` …").
- An adversarial `reviewer` agent approves the exact intake argv before it is filed (section 10).
- When a derivation escalates or derives a wrong check, correct the filing form, close the pebble and file it again.
  Do not add notes that name fields or spec shapes: a note reaches the derivation as a remedy. Do not change the repo
  to fit one derivation. 2026-10-28: the Lean filings stated a predicate, and five derivations escalated or
  invented fields (`docs/lean/proof-dispatch-plan.md`, Phase 0 result, correction).
- Merge paths:
  - a worker pebble: intake, estimate, spec review, queue;
  - critical-severity Solidity: the spec derives with `forge_mutation_gated_test` (exophial `docs/ORACLES.md`). If
    intake still escalates, the husband writes the change with red-first forge tests and merges it with
    `exophial integrate --branch <branch>`;
  - an operator-known change outside `scripts/`, `tests/` and `contracts/specs/`: `exophial apply <branch>`.
- Fixturenet grading (operator ruling, 2026-10-25), interim: a recorded host run grades a fixturenet pebble, and
  `exophial apply` merges it. The commit quotes each fixturenet clause with the host command, the tested sha and the
  verdict lines. The coordinator closes the original pebble as superseded.
- A pebble keeps the base recorded at its first claim, and grading reads the spec and the repo config at that base
  (exophial coordinator, 2026-10-30; `exo-7ea`). A rebase of the branch does not move it. If the base predates a
  needed change (a refreshed spec, an `environment_command`), the recovery is the one the coordinator names:
  refresh the spec, rebase in the kept worktree, then unpark, so the unpark records the new merge-base. When no
  route exists (`arn-5c3`, `arn-d62`: base before `environment_command`), hold the pebble until the exophial fix
  merges. Never hand-merge, never hand-write a base-sha, and never rename a branch to mint a new identity.
- A spec derived before an approved type change still checks the old statement text. Refresh it
  (`exophial specify --refresh`) before that pebble is graded (`aln-c03`, 2026-10-30).

## 6. Environment for exophial commands

- Token: `CLAUDE_CODE_OAUTH_TOKEN="$(python3 ~/code/git_puller/repos/exophial/.scratch/oauth_token.py)" exophial ...`.
  The helper picks the Anthropic account with the most headroom. Do not print the token.
- Set `TMPDIR=<repo>/.scratch` for `exophial dispatch`, `apply`, `integrate`, `intake` and `specify`.
- The grader runs each oracle in a sealed HOME and reads the oracle variables (`EXOPHIAL_ORACLE_TOOLCHAIN_BINARIES`,
  `EXOPHIAL_ORACLE_DATA_DIRS`) from the caller's environment. The supervisor sets them per repo in `ORACLE_ENV`
  (`.scratch/supervise/supervisor.py`). That dict is the record. A repo must also be in its `WATCHED` tuple, or the
  supervisor drops the repo's queue entries. A binary in `node_modules/.bin` is not on PATH. Do not list it.
- `exophial init` takes no `-e` flag. To set a prefix or languages, run the packaged play:
  `~/.local/share/uv/tools/exophial/bin/ansible-playbook <exophial>/ansible/init.yml -e exo_target_repo=<repo>
  -e exo_pebble_prefix=<p> -e '{"exo_languages_override": ["typescript"]}'` (JSON form for lists).
- `systemd-oomd` can kill the user manager of `rix` (`user@1001.service`, its `init.scope`) when memory pressure in
  that tree stays above 50% for 20 s (stock drop-in `10-oomd-user-service-defaults.conf`). Seen 2026-10-29 15:37Z,
  2026-10-31 01:39Z and 14:31Z. `/run/user/1001` then disappears, and podman, the supervisor, `ingest-*` and the
  vLLM embed/rerank services stop. Linger was never off: `/var/lib/systemd/linger/rix` exists since 2026-06-24.
  Fix (2026-10-31): `/etc/systemd/system/user@1001.service.d/10-restart-on-oomd-kill.conf` sets
  `Restart=on-failure`, `RestartSec=2s`; revert by removing it and `sudo systemctl daemon-reload`. Services without
  their own restart policy (`ingest-sessions`) still need a manual start. Commands that run podman need
  `XDG_RUNTIME_DIR=/run/user/1001`. The watcher wakes on `supervisor_down`.
- Process management (operator, 2026-10-31: the user manager must not be killed; heavy work is bounded):
  - `init.scope` of `user@1001.service` has `ManagedOOMPreference=omit` (`~/.config/systemd/user/init.scope.d/50-oomd-omit.conf`;
    cgroup xattr `user.oomd_omit=1`), so `systemd-oomd` never picks the user manager.
  - Every Lean command runs through `~/bin/lean-run <cmd>` (for example `lean-run lake build`): a scope in
    `lean.slice` (`~/.config/systemd/user/lean.slice`: `MemoryHigh=48G`, `MemoryMax=64G`,
    `ManagedOOMMemoryPressure=kill`; per scope 12G/16G), at most 4 at once. Agent briefs must say so.
  - Measured: `lake env lean ArmadaLean/Nitro/Liveness.lean` 1.78 GiB peak RSS; warm `lake build` 0.44 GiB.
  - Gate and worker containers still run outside any bounded slice: exophial's podman argv has no `--cgroup-parent`
    (sent to the exophial coordinator).

## 6a. Reading sessions

- Read omp and Claude Code sessions through ingest_sessions (operator, 2026-10-30), not by grepping raw JSONL. The
  database is Postgres on `127.0.0.1:5433` (container `ingest-pg`). Read-only access:
  `cd ~/code/git_puller/repos/ingest_sessions/ingest_sessions && uv run python -c 'import psycopg; from ingest_sessions
  import db; c = psycopg.connect(db._dsn(role="ro")); ...'`. Tables: `sessions`, `records`, `summaries`, `history`.
  The MCP server (`ingest-sessions-server`, port 8741) ingests new sessions and serves the `query` tool.
- After a host crash, check `systemctl --user is-active ingest-sessions` (with `XDG_RUNTIME_DIR=/run/user/$(id -u)`).
  On 2026-10-29 it failed its start with `dependency` because `ingest-pg` came up 8 s later, and nothing restarted it.
  `max(modified)` in `sessions` shows when ingestion stopped.

## 7. Git and repositories

All repositories are under `~/code/git_puller/repos/`.

- Never pass `--git-dir` or `--work-tree` that points at a main checkout. On 2026-10-21 such a command, run from a
  scratch worktree, changed the armada-nitro main checkout.
- Never run `git stash` in armada-nitro.
- armada-nitro: exophial commits `.pebbles/events.jsonl` in its own commits. Do not commit it by hand. After a merge,
  if a modified file equals the pre-merge base, restore it with `git checkout -- <file>` (`exo-cba`).
- armada-laconic: commit to `main`. Push with `git push AFDudley.github.com HEAD:main`. The remote URL holds a token. Do
  not print it.
- armada-lean (prefix `aln`): exophial owns `integrated`, and the fork's `main` follows it through
  `.github/workflows/ff-main-from-integrated.yml`. Push to the ship-armada remote is disabled. A change under
  `scripts/` merges only by `exophial integrate --branch`, because `exophial apply` refuses acceptance machinery.
  Old branches survive as `archive/<name>` tags.
- `atlas` is a branch that exophial reserves (`integrator.py:198-203`). Never delete it.
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
