# Handoff to the forked coordinator (pane w3:p4): everything

Date: 2026-10-31. From the coordinator in pane w3:p1 (session 01a1888d). Operator order: all work moves to
w3:p4, except the one worker still running in w3:p1.

## Division of work

- **w3:p4 owns everything:** the type doctrine and the typed library, the spec module, the premise analysis,
  Lean husbandry of parked pebbles, the open operator questions, every merge and every exophial command
  (`apply`, `integrate`, filing; read `operating-rules.md` §5-§6 and the `exophial-before-use` rule first),
  the exophial coordinator liaison (pane w6:p1), and the host process limits (operating rules §6).
- **w3:p1 keeps only `ProveSatp32`** (the `satp_theorem_3_2` proof) until it finishes. w3:p1 forwards each
  branch it reports (`draft/participant-noland`, then `draft/prove-satp-theorem-3-2b`) to w3:p4 for
  verification and merge. Its state: `docs/wip/satp_theorem_3_2/HANDOFF.md` on its branch.

## Merge queue at handoff (armada-lean, one `apply` at a time)

- In flight from w3:p1 (job started before the handoff; its results go to w3:p4 by message):
  `draft/w1-step-frame` (9b70d05), then `draft/w1-peel-validmove` (8d8074c). Logs:
  `armada-lean/.scratch/apply-<branch>.log`.
- Next: `draft/participant-noland` (`ProveSatp32`, 7edd4f8 before rebase; model repair: `Nitro.participant`
  third disjunct excludes `.land`, operator ruling 2026-10-31; invariant `participant_received_unforged`;
  ported helpers). Rebase onto the trunk after the two above, verify, `apply`.
- Other repos: `draft/cgroup-parent` (one commit each, sets `dispatcher.cgroup_parent: exophial-gate.slice`) in
  armada-poc, armada-circuits, armada-nitro: `integrate --branch` each. armada-poc `sync/upstream-2026-10-30`
  (8813d21a): last gate hit the repo's 300 s gate timeout during the first image build. armada-circuits
  `draft/inputs-distinct` (a505b9f): verified by its agent; changes all 51 shapes' keys.
- How w3:p1 verified a branch before merging: based on `integrated`; `/home/rix/bin/lean-run scripts/check.sh
  --json` all `_ok` true; `git diff integrated -- scripts/sigs-baseline.tsv` (empty, or only approved rows);
  `#print axioms` on new proofs. `apply` for `ArmadaLean/`-only changes, `integrate --branch` when `scripts/`,
  `tests/`, `.exophial/` or `contracts/specs/` change. Send any exophial refusal's log path to w6:p1.

## Host limits (set up by w3:p1 today; details in operating rules §6)

- `~/.config/systemd/user/lean.slice` (MemoryHigh 48G, MemoryMax 64G), used by `/home/rix/bin/lean-run`
  (max 4 concurrent). `~/.config/systemd/user/exophial-gate.slice` (MemoryHigh 40G, MemoryMax 56G), used by
  exophial through `dispatcher.cgroup_parent`. `init.scope` of `user@1001.service` has
  `ManagedOOMPreference=omit`; `/etc/systemd/system/user@1001.service.d/10-restart-on-oomd-kill.conf`.
- The supervisor (`armada-laconic/.scratch/supervise/supervisor.py`) runs; its watcher (`watch.py`) was
  restarted by w3:p1 each hour. Restart it from w3:p4: `WATCH_IDLE_S=86400 python3 watch.py` in that directory.

## Typing work

- Design, approved: `docs/lean/typed-library.md`. Rulings recorded there: `TimedChannel ε := 2·ε ≤
  challengeDuration`; `LedgerChannel` is `kind = .ledger` (the app is a separate premise, `RunsConsensus`);
  `TurboOutcomeNetwork`. Your doctrine branch `draft/type-doctrine` supersedes or extends it: reconcile the two.
- `draft/refined-types` (armada-lean, tip `2bdf598`, your commits and RefinedTypes'): merging now through
  `integrate --branch`; w3:p1 reports the result to you.
- Spec module: armada-lean `draft/spec-module-2` (worktree `.scratch/spec2-wt`, Opus). Verified on its base
  `626ae18`: build, check, statement-change test. Not merged. Cutover per its `docs/spec-module-design.md`:
  rerun `scripts/convert_spec.py` on the current trunk, verify, send to w3:p1. It must absorb the refined types.
- Premise analysis: armada-lean `.scratch/hyp-analysis.md` and `.json` (89 theorems, 363 premises; provenance
  from doc comments; 338 with no stated source).
- Measured: armada-lean has 0 `Subtype` arguments in 937 theorems on trunk; 6 of 22 structures carry proofs.

## Open questions waiting for the operator (bring one at a time, in prose)

1. Refile orders for `aln-cc8` (funding point) and `aln-223` (`satp_end_user_recovers`): statements changed.
2. `deposit_bounds_value` (`aln-c5d`): three questions (Turbo game, SATP game, current mismatch).
3. IR/IC design for `finalizable_or_two_enabled`, `single_enabled_is_finalizable` (`aln-6ca`).
4. `addTo` versus go-nitro (keeps `allocationType`, `MultiAssetHolder.sol:340-366`).
5. Your doctrine §4 decisions (rule-2 exception, switch-on, briefs).

## Husbands: parked and blocked pebbles in the supervisor queue

- Queue: `armada-laconic/.scratch/supervise/queue.jsonl`; supervisor log beside it. Supervisor has 0 workers.
- `aln-eb2` (`turbo_value_achievable`): its counterexample no longer applies after the refined types merge
  (proved `chain5_not_turboOutcomeNetwork`); candidate for unpark once merged (needs operator order).
- `aln-a73` (`satp_theorem_3_2`): covered by `ProveSatp32` on `draft/prove-satp-theorem-3-2b`; close as
  delivered when it merges (w3:p1).
- `aln-6ca`, `aln-78b`, `aln-5ab`, `aln-9b7`, `aln-c5d`, `aln-cc8`, `aln-223`: see `notes/handoff.md` §1 and
  the questions above.
- Closed today: `aln-e4e`, `aln-bd2`, `aln-d55` (out of scope), `aln-6f0`, `aln-50b` (delivered).

## Pool side (lower priority; operator: Nitro first)

- Vault remodel `draft/vault-locked-shares` (`26a1238`): needs a vault state machine for the locked-shares
  invariant (review note in its history). Paused.
- Circuits `draft/inputs-distinct` (armada-circuits, `a505b9f`): verified by its agent; invalidates all 51
  shapes' keys (list in its report). Waits to merge.
- armada-poc `sync/upstream-2026-10-30`: last gate hit the repo's 300 s gate timeout during the first image
  build; retry pending (w3:p1).

## Host rules (operating rules §6)

Every Lean command through `/home/rix/bin/lean-run` (max 4, `lean.slice`). Do not copy `.lake/build` between
worktrees of different branches. `systemd-oomd` omits the user manager's `init.scope`.
