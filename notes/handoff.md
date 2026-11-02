# Handoff — Armada build programme

This note holds the work in flight, the next steps, what waits for the operator, and the open questions. It does not
copy state that a command, the pebble bus or another document shows. Rules and procedures are in
`engineering/operating-rules.md` and the documents that it lists. Remove an item when it is resolved; git history keeps
the record. Commit every change to `main`. A claim that the writer did not observe has the mark [INFERENCE].

The plan is `engineering/build-plan.md`. Decisions are in `labbook/`.

## 1. Lean specification (the main track)

The old adjudicator model `adjudicator-model-v4` is reference only. Do not extend it
(`labbook/2026-10-26-adjudicator-layout-audit.md`).

`github.com/AFDudley/armada-lean` has one trunk, in the clone `~/code/git_puller/repos/armada-lean`. It is an exophial
instance (prefix `aln`): exophial owns `integrated`, and the fork's `main` follows it. Approval tags:
`pool-symbolic-approved-2026-10-28` (`labbook/2026-10-28-pool-symbolic-types-approval.md`),
`nitro-proofs-approved-2026-10-28` (`labbook/2026-10-28-nitro-repaired-types-approval.md`),
`nitro-theorems-approved-2026-10-27`. `armada-nitro-spec` (worktree `armada-lean-armada-nitro-spec`) is the Armada
layer on the approved, not the repaired, Nitro spec. It waits for a rebuild on `main`.

**In flight (2026-10-31, about 08:00Z).** Phase 2 (Nitro Liveness proofs). The trunk `integrated` is at `5696729`.
The supervisor runs (`pgrep -af 'python3 supervisor.py'`), but it has had 0 workers since 03:34Z: every queued
pebble is parked or blocked. Restarting needs the operator's orders (below, item 1). The watcher wakes on
`supervisor_down` (operating rules §6, linger).

- Merged on 2026-10-31: funding-point type change (`4dad847`), `aln-223` `hog` (`1771430`), D7 voucher rule
  (`ae9b109`, with `Nitro.network_received_voucher_signers`), `turbo_order_independent` (`aln-dbf`).
- `draft/out-of-scope` (`5b8946e`, worktree `armada-lean/.scratch/oos-wt`) marks the three Nitro §5.3-§5.4
  sub-channel theorems out of scope (operator ruling). Not merged. `apply` failed twice (CAS lost; then
  "re-run acceptance gate REJECTED" with no step named) and `integrate --branch` failed with `returncode=-9`. The
  configured gate passes on that tree. `podman events` shows seven `exophial-gate-*` containers dying 03:42-03:52Z.
  Report not yet sent to the exophial coordinator (waits for the operator). `aln-e4e`, `aln-bd2`, `aln-d55` are
  closed as out of scope.
- `aln-a73` (`satp_theorem_3_2`): its voucher-stall claim is wrong after D7. Opus's counter-move file
  `armada-lean/.scratch/a73-verify/tree/docs/counterexamples/a73_counter_move.lean` (`counter_after_adv_checkpoint`).
- `aln-eb2` (`turbo_value_achievable`) is false: a Turbo value routed through a virtual channel cannot be
  transferred. Checked counterexample on `worker/aln-eb2` (`986c203`). Needs the Turbo-scope ruling.
- `draft/spec-module-2` (worktree `armada-lean/.scratch/spec2-wt`, Opus): the spec module, verified (build, check,
  statement-change test). Cutover waits for the operator's citation choice. Citation research:
  `armada-lean/.scratch/pool-citations.md`.
- Type design approved: `docs/lean/typed-library.md`. Premise analysis: `armada-lean/.scratch/hyp-analysis.md`
  (89 theorems, 363 premises, 338 with no stated source).

**Waits for the operator** (one ruling at a time, in this order).
1. Orders that restart Nitro proving: unpark `aln-50b` (D7 merged), `aln-6f0` (stall was the podman outage; kept
   work `2e5a05d` passes `check.sh`), `aln-a73` (claim refuted); close and refile `aln-cc8` and `aln-223`.
2. `Constrained.distinct` (`Transact.lean:146-148`) credits the circuit with a distinctness check that
   `Transact.circom` lacks (`SECURITY_REVIEW.md:51`). The contract enforces it (`TransactModule.sol:384-385`).
   Repair in the model: approve?
3. Timeout: define a sensible timeout as `2·ε ≤ challengeDuration` (type `TimedChannel ε`)?
4. Spec-module citations: (a) add verified citations, then cut over; or (b) cut over now.
5. Send the gate-kill evidence to the exophial coordinator?
6. Turbo scope (`aln-eb2`): Turbo theorems take a `TurboChain` type with no virtual channels?
7. Vault: `Vault.lean` models the pre-fix contract; upstream fixed #365 with locked dead shares (`04753139`);
   `mitigated_*` model a virtual-offset fix that no contract has. Remodel after the armada-poc upstream merge?
8. Phase 3: rerun `live_leaf_matches` with a bare brief from a fresh agent
   (`labbook/2026-10-30-qwen-lean-phase3.md`, Control 1).
9. armada-poc `sync/upstream-2026-10-30` (`8813d21a`, fork only; report `.scratch/poc-upstream-sync.md`): merge it,
   and accept or change two points: the yield vault takes upstream `LOCKED_SHARES`, and `verifyOwned` becomes
   virtual.
10. armada-circuits `integrated` `9699ba9` is not pushed.

**Next steps** (`docs/lean/proof-dispatch-plan.md`; operator rulings 2026-10-28).
1. Phase 2 continues, one Liveness pebble at a time. A counterexample stops the work on that statement
   (`docs/process.md` 8.3). Before a type question goes to the operator, read every source passage it cites in full.
   The paper decides honest behaviour. go-nitro client code does not (`docs/lean/writing-lean.md` §10).
2. Review the merged Opus proofs against `docs/lean/verification.md` §6a (size per declaration, duplicated blocks,
   library first): `finalizable_unique`, `satp_intermediary_before_round2`, `ledger_partial_withdrawal`.
3. The executable-operations review is done. Reports: `.scratch/ops-review-nitro.md`, `.scratch/ops-review-pool.md`.
   The cleanup merged (`cfae60e`). Open: an API per operation, with exact-effect and frame lemmas (`writing-lean.md`
   §3 and §6, `verification.md` §6a). After that, remove the `simp only [Network.applyAdv, Network.land]` unfold that
   the cleanup added in `Network.applyAdv_chain_reachable`. `Machine.interleave` stays out until the first
   cross-chain user (`labbook/2026-10-29-nitro-reachability-ruling.md`).
4. Rebuild `armada-nitro-spec` on the repaired Nitro spec. The payee note key becomes Nitro's external-destination
   parameter `E`. Channels follow the `ledger` and `virtual` kinds. Apply the pending payee-key rulings
   (`payout_spendable`, `payment_note_only_payee`, no ciphertexts). Then give its theorem types to the operator.
5. Open exophial defects: `exo-a72c`, `exo-c7d6`, `exo-c310`, and `exo-7ea` (a pebble keeps the base of its first
   claim). Track each with `pb show <id>` in `~/code/git_puller/repos/exophial`. The installed build has the fix that
   stops a collided pebble from parking as STALLED.

**Findings of 2026-10-26/27** (lab book): `2026-10-27-gonitro-reclaim-dos.md`, `2026-10-27-payout-griefing.md`,
`2026-10-27-nitro-liveness-game-defects.md`, `2026-10-26-utxo-channel-formalizations.md`, `2026-10-27-bend2-to-evm.md`.

## 2. Other tracks

- **E11 pause (operator ruling, 2026-10-26)**: all E11 work waits until the Lean work is complete. Conflict: the E11
  Aave yield gate merged on armada-nitro (`fddf4337`, bus `arn-0a2`; `arn-273` closed). No record of a lifted pause
  was found in `labbook/` or this note. Ask the operator whether the pause still holds before an E11 pebble is queued.
- **F4 collateral** is replaced by the Lean work.
- **F3 governance** (husband `HusbandF3j`):
  - `arp-6cc` (E2) and `arp-445.1` (E5) are closed, but armada-poc `integrated` has no merge commit for them. The
    husband must find out why.
  - `arp-acf` (E3 follow-up) and E7 (treasury ARM in the payout divisor) wait for operator rulings.
  - G7 (`arn-5c3`, `arn-d62`): held for `exo-7ea` (section 1, "In flight").
  - Open: the replacement of the live armada-1 deployment and the move of ARM holders.
- **F2** (husband `HusbandF2c`, also E11): N9 waits for N3. **F5 client:** waits for N3.
- **Dispatch.** One supervisor runs (`pgrep -af supervisor.py`), queue at
  `.scratch/supervise/queue.jsonl`, log at `.scratch/supervise/supervisor.log`.
- **Gates.** armada-circuits now has `.exophial/gate/Dockerfile` on `integrated`. The armada-poc gate still omits
  `test:config`.

## 3. Open items with no other record

- **Proof path (nimbus verified proxy).** Lab book: `2026-10-20-nimbus-fixturenet-and-proxy-proof.md` and
  `2026-10-20-proxy-and-archive-open-questions.md`. Open operator orders:
  1. File a local issue for the WASM proxy stall. armada-client is the tracker. `exophial intake` there needs a
     committed armada-client document that states the rule (ADR-0009 D9.2, `build-plan.md` T1-4). armada-client
     doctrine requires operator approval of the issue body. A draft is in session `01a14aa3`.
  2. Analyse whether the plan needs per-log proofs. A stopped agent left an unchecked script and data in
     `.scratch/PerLogAnalysis/`.
  3. The S7 acceptance needs the proof as a job in the armada-nitro gates. No chunk schedules it.
- **E16 (armada-nitro).** The hub's close script pays with `concludeAndTransferAllAssets`
  (`script/CustodianFrostScenario.s.sol`). With a second asset, one frozen asset blocks every payout of that close.
  The fix is part of E11. The E16 brief's scope statement about E11 contradicts itself.
- **E17 (armada-nitro).** Not scheduled. Candidate defects that no test confirms: `_executeSingleAssetExit` ignores a
  `withdrawHelper` allocation (clause 31, Sil4); `reclaim` credits only target allocations 0 and 1 (Sil6).
- **Selective disclosure.** Design: page 16 and ADR-0020. Circuit: ship-armada/armada-circuits#3. SDK: draft
  ship-armada/armada-sdk#142. No hop is checked against events from real transactions. The hub's EIP-1271 check ran
  only against a stand-in. `arn-12e` covers both and is parked under the E11 pause. The production proof system and
  ceremony are open.
- **Hub tracking of cross-chain swaps.** `engineering/shielded-nitro-bridge-design.md` §4–§6 and ADR-0015 put hub
  inventory, rebalancing and pricing out of protocol. No reference hub daemon or design page exists.
- **exophial.** `exophial/acceptance-e2e-outer-gate.md` is untracked and not filed.
- **mtm production.** The cron line is disabled (backup `.scratch/crontab.before-2026-10-16`). The reconcile failed at
  "Converge the gateway surface to the promoted ref". Before the cron line is enabled again, its checkout must follow
  `origin/main`.
- **Documentation.**
  - `engineering/build-plan.md`: S3 omits N9 (9 points). S5 uses the old N1–N8 sizes (43), not the split briefs (50).
  - The site link to this note is broken because `notes/` is not rendered.
- **AFDudley forks push over SSH.** Every repo's `AFDudley.github.com` (or `fork`) remote push URL is `git@github.com:AFDudley/<repo>.git`. `armada-client` is the one exception: its only remote, `origin`, still pushes over HTTPS.
- **armada-circuits `integrated` tip.** `9699ba9` adds `ff-main-from-integrated`. `main` is 36 commits behind it
  (`git log --oneline main..9699ba9` in `~/code/git_puller/repos/armada-circuits`); the workflow has not caught
  `main` up yet.
- **Branch cleanup, 2026-10-28.** Deleted branches survive as `archive/<name>` tags on each fork. Left for a decision:
  - armada-poc's fork has 153 branches with no local copy (`brent/*`, `iskay/*`, `feature/*`, …). armada-sdk's fork
    has 19 such branches.
  - Unclassified branches: armada-nitro `arn765-abl`, `backup/main-before-reset-2026-10-13`, `docs/r1-single-chunk`,
    `epic/arn-46d`, `hand/e16-2026-10-14`, `migrate/briefguards2`, `task/arn-46d1-rebase`; armada-poc
    `design/interactive-baseline`, `design/rendered-baseline`, `docs/design-system-policy`, `exo/drop-license-check`;
    armada-circuits `ceremony/1x2-0001-iskay`, `disclosure-circuit`, `docs/n6-battleship-brief`, `feat/public-ceremony`.
  - Worktrees with uncommitted changes: armada-nitro `worker-arn-781`, `worker-arn-bd4`; armada-poc `hand/arp-caf-e3`,
    `worker-arp-dec`; armada-circuits `.scratch/cfg-sonnet`, `.scratch/n5fix`, `.scratch/probe-circom`.

## 4. Questions to the operator

- `supervisor.py` replaces scheduling that exophial lacks (host-wide limits, quota, local versus cloud routing). Move
  it into exophial as a feature?
- Give the S7 proof gate job to a husband?
- mtm production: who owns the failing reconcile, and when is the cron line enabled again?
- Prune dangling Docker images? Remove marks from `pool/27b-class` and `pool/long`?
