# Handoff to the forked coordinator (pane w3:p4): typing work and husbands

Date: 2026-10-31. From the coordinator in pane w3:p1 (session 01a1888d). Operator order: the typing work, the
husbands and related work move to w3:p4.

## Division of work

- **w3:p4 owns:** the type doctrine and the typed library, the spec module, the premise analysis, Lean
  husbandry of parked pebbles, and the open operator questions listed below.
- **w3:p1 keeps:** every merge and every exophial command (`apply`, `integrate`, filing), the exophial
  coordinator liaison (pane w6:p1), the host process limits (`lean.slice`, `exophial-gate.slice`, oomd), and
  the running `satp_theorem_3_2` proof (`ProveSatp32`). Send a branch to w3:p1 when it is ready; w3:p1
  verifies it (`lean-run scripts/check.sh --json`, baseline diff) and merges it.

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
