# Handoff — Armada build programme

This note holds the work in flight, the next steps, what waits for the operator, and the open questions. It does not
copy state that a command, the pebble bus or another document shows. Rules and procedures are in
`engineering/operating-rules.md` and the documents that it lists. Remove an item when it is resolved; git history keeps
the record. Commit every change to `main`. A claim that the writer did not observe has the mark [INFERENCE].

The plan is `engineering/build-plan.md`. Decisions are in `labbook/`.

## 1. Lean specification (the main track)

The old adjudicator model `adjudicator-model-v4` is reference only. Do not extend it
(`labbook/2026-10-26-adjudicator-layout-audit.md`).

`github.com/AFDudley/armada-lean` has one trunk, in the clone `~/code/git_puller/repos/armada-lean` (bus prefix `aln`).
The dispatcher owns the trunk branch; the fork's `main` follows it. Approval tags:
`pool-symbolic-approved-2026-10-28`, `nitro-proofs-approved-2026-10-28`, `nitro-theorems-approved-2026-10-27`.
`armada-nitro-spec` (worktree `armada-lean-armada-nitro-spec`) is the Armada layer on the approved, not the repaired,
Nitro spec. It waits for a rebuild on `main`.

**Owner.** The coordinator in pane w3:p4 (session `01a192b4`) owns all Lean work, by operator order of 2026-10-31:
merges, husbandry, the liaison with the dispatcher's coordinator (pane w6:p1) and the host limits
(`engineering/operating-rules.md` §6). The earlier note `notes/handoff-typing-and-husbands.md` is removed; this
section replaces it.

**State at 2026-11-02, about 12:35Z** (from the w3:p4 session and `git log`; `main` at `89e9510d`).

- **Proofs.** Every Nitro and pool spec theorem is proved except the two Nitro-24 claims,
  `finalizable_or_two_enabled` and `single_enabled_is_finalizable` [INFERENCE: from the coordinator's 12:26Z report;
  I did not run `check.sh`]. Proved since 2026-10-31: `satp_theorem_3_2` (`0e1c470f`), `turbo_value_achievable`
  (`5b3e181e`), `turbo_value_is_value`, `satp_end_user_recovers` (`ab061228`), `funding_point_universally_finalizable`
  (`893d1175`), `deposit_bounds_value` (`5830ade1`). The three Nitro 2019 §5.3-§5.4 sub-channel theorems are block
  comments (`18559321`, operator ruling 2026-11-01).
- **On the trunk since 2026-10-31.** The spec module (statements protected by type, computed check), the refined types
  (`TimedChannel`, `LedgerChannel`, `TurboOutcomeNetwork`), the type-doctrine linters and `decisionScope` (0 findings
  since `f5374cb5`), the committed gate baseline with a freshness hook (`8b035d82`), recorded searches (`says`), the
  round combinator, `HonestSend`, `SupportWindow`, the repair pass R.2 (one equation per operation) and R.4 (six
  adjudicator operations), the one safety invariant `Adjudicator.Inv` (`3172bf8b`), and the verification-chain
  pointers to `labbook/2026-11-02-refinement-risks.md` §7.
- **Nitro-24** (branch `draft/nitro24-appir`, tip `10e82339`). Four counterexamples, each a checked failure test in
  `docs/counterexamples/`. Three honest-side rules are decided (one state per turn; after a final state, only final
  states; `CanMoveTo` checks validity). They change the meaning of 29 Spec Liveness statements through their types
  (the branch's repair doc §7.1). The coordinator's diagnosis (12:29Z): the claims swap a quantifier against the paper
  (Nitro:640-652). `Enabled` is "no single strategy prevents Ω", while `Finalizable` needs one strategy that prevents
  every other outcome at once. Patching does not converge. No other theorem uses the two claims.
- **Codemods.** The local-model pilot merged 1 of 3 (`shield_can_happen`, `034b8af1`, net 0 lines). The other two
  looped for 3 hours on thin briefs; the operator ruled that looping means a bad spec. `LayerCodemods` (Sonnet) found
  that `layerDiscipline` counts constants that enter only through a case split's type or a bridge lemma's statement:
  7 of 15 hits are false. `CodemodOpus` (Opus, branch `draft/cm-opus`, dispatched 12:33Z, no commit yet) fixes that
  measure, checks the premise lints, replaces the real hits with shared lemmas, and removes the R.4 duplicates
  (`reclaim_decompose` beside `reclaim_some`). Acceptance: every `check.sh` field true, `sorry` stays 2, `verify_says`
  passes, net line change of `ArmadaLean/` below 0.
- **Pebble bus.** The 24 codemod pebbles are out of the supervisor queue. `aln-290`, `aln-36e`, `aln-8ab` are parked.
  The repair-pass pebbles `aln-185`, `aln-6a9`, `aln-ba0`, `aln-904`, `aln-299` are closed with their merge shas. No
  reducer runs [INFERENCE: last reported at 09:57Z].
- **Not scheduled** (`labbook/2026-11-02-refinement-risks.md` §7): the EquiVM and EVMLean changes V1-V3, the O1 gas
  measurement, the conformance vectors, and the R1CS importer. Also not started: the general liveness theorem (R.6).

**Waits for the operator.**
1. Nitro-24 direction (asked 12:29Z): (1) measure against p's honest strategy, `Enabled_σ` (the coordinator's
   recommendation); (2) a composition premise; (3) remove both claims and record the counterexamples. The operator
   asked at 12:34Z whether Nitro-24 adds anything to the model; that discussion is open.
2. Vault: `Vault.lean` models the pre-fix contract; upstream fixed #365 with locked dead shares (`04753139`);
   `mitigated_*` model a virtual-offset fix that no contract has. `draft/vault-locked-shares` (`26a1238`) is paused.
   Remodel after the armada-poc upstream merge?
3. armada-poc `sync/upstream-2026-10-30` (`8813d21a`, fork only; report `.scratch/poc-upstream-sync.md`): not merged
   (checked: not an ancestor of the trunk tip `fbcf2b1e`). Accept or change two points: the yield vault takes upstream
   `LOCKED_SHARES`, and `verifyOwned` becomes virtual.
4. armada-circuits: the local trunk `9699ba9` is ahead of the fork's (`f790107`, `git ls-remote`). The branch
   `draft/inputs-distinct` (`a505b9f`) adds the input-distinctness block that `Constrained.distinct` models; it is not
   merged and changes the keys of all 51 shapes. Under the 2026-11-02 rulings the Lean model is the specification, so
   this is a merge to schedule, not a model question.
5. Phase 3 (local model for Lean): the codemod pilot is new evidence (1 of 3, after two corrections). The earlier
   question, a rerun of `live_leaf_matches` with a bare brief (`labbook/2026-10-30-qwen-lean-phase3.md`), is not run.

Resolved since the last version of this note: Nitro proving restarted and finished except Nitro-24; `TimedChannel`
(2·ε, 2026-10-31); spec-module cutover; the sub-channel theorems; the Turbo scope (`TurboOutcomeNetwork`); the
refinement-risk questions (`labbook/2026-11-02-refinement-risks.md` §7).

**Next steps.**
1. Merge `draft/cm-opus` when `CodemodOpus` reports, after a check of its diff against its acceptance.
2. Act on the Nitro-24 direction when the operator gives it. Until then no agent proves the two claims.
3. Rebuild `armada-nitro-spec` on the repaired Nitro spec. The payee note key becomes Nitro's external-destination
   parameter `E`. Channels follow the `ledger` and `virtual` kinds. Apply the pending payee-key rulings
   (`payout_spendable`, `payment_note_only_payee`, no ciphertexts). Then give its theorem types to the operator.
4. Specs for any work item name the base operation the lint flags, the lemma that replaces it or the lemma to add,
   and how to run Lean in the worktree (operator, 2026-11-02: looping is a spec defect).

**Findings of 2026-10-26/27** (lab book): `2026-10-27-gonitro-reclaim-dos.md`, `2026-10-27-payout-griefing.md`,
`2026-10-27-nitro-liveness-game-defects.md`, `2026-10-26-utxo-channel-formalizations.md`, `2026-10-27-bend2-to-evm.md`.
Later decisions: `2026-11-01-l2bridgeapp-verifier.md`, the desirable-game rulings (`2026-10-30-app-classes-ic-ir.md`),
`2026-11-02-refinement-risks.md`, and `docs/process.md` "The repair pass".

## 2. Other tracks

- **E11 pause (operator ruling, 2026-10-26)**: all E11 work waits until the Lean work is complete. Conflict: the E11
  Aave yield gate merged on armada-nitro (`fddf4337`, bus `arn-0a2`; `arn-273` closed). No record of a lifted pause
  was found in `labbook/` or this note. Ask the operator whether the pause still holds before an E11 pebble is queued.
- **F4 collateral** is replaced by the Lean work.
- **F3 governance** (husband `HusbandF3j`):
  - `arp-6cc` (E2) and `arp-445.1` (E5) are closed, but armada-poc `integrated` has no merge commit for them. The
    husband must find out why.
  - `arp-acf` (E3 follow-up) and E7 (treasury ARM in the payout divisor) wait for operator rulings.
  - G7 (`arn-5c3`, `arn-d62`): held for `exo-7ea` (a pebble keeps the base of its first claim).
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
