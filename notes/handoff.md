# Handoff — Armada build programme

This note holds the work in flight, the next steps, what waits for the operator, and the open questions. It does not
copy state that a command, the pebble bus or another document shows. Rules and procedures are in
`engineering/operating-rules.md` and the documents that it lists. Remove an item when it is resolved; git history keeps
the record. Commit every change to `main`. A claim that the writer did not observe has the mark [INFERENCE].

The plan is `engineering/build-plan.md`. Decisions are in `labbook/`.

## 1. Lean specification (the main track)

The old adjudicator model `adjudicator-model-v4` is reference only. Do not extend it
(`labbook/2026-10-26-adjudicator-layout-audit.md`).

Branches in `github.com/AFDudley/armada-lean`. Each has a worktree at `~/code/git_puller/repos/armada-lean-<branch>`.
Verify only in a separate tree, for example `armada-lean-nitro-theorems/.scratch/v`.

| Branch | Contents |
|---|---|
| `nitro-proofs` | the repaired Nitro specification; channel kinds `ledger` and `virtual`; tag `nitro-proofs-approved-2026-10-28` (operator approval, `labbook/2026-10-28-nitro-repaired-types-approval.md`) |
| `nitro-theorems` | the Nitro theorem types that the operator approved before the repair; tag `nitro-theorems-approved-2026-10-27` |
| `pool-symbolic` | the pool model without contract layout; every theorem type approved (`labbook/2026-10-28-pool-symbolic-types-approval.md`) |
| `armada-nitro-spec` | the Armada layer on the approved, not the repaired, Nitro spec; `docs/verification-chain.md` |

**In flight.** Nothing.

**Waits for the operator.** Nothing on this track.

**Next steps, in order** (operator, 2026-10-28: the pool proofs come before the Nitro proofs).
1. Prove the `sorry` theorems of `pool-symbolic` (`grep -rn sorry ArmadaLean/`). Their types are approved. A
   counterexample stops the work on that statement (`docs/process.md` 8.3).
2. Review the executable operations of the Nitro and Armada layers for composition, typed state and footprints. They
   are the input of the Lean → Stele translation (`engineering/build-plan.md` "Verification chain"). Make the changes
   that the review finds before any Nitro or Armada proofs. A change to an approved theorem type goes back to the
   operator.
3. Rebuild `armada-nitro-spec` on the repaired Nitro spec. The payee note key becomes Nitro's external-destination
   parameter `E`. Channels follow the `ledger` and `virtual` kinds. Apply the pending payee-key rulings
   (`payout_spendable`, `payment_note_only_payee`, no ciphertexts). Then give its theorem types to the operator.
4. Proofs: Nitro Liveness, then the Armada layer. Some approved Nitro types had no review of their Lean text. A
   counterexample stops the work on that statement.
5. File the armada-circuits need: owned circuits with more than one input (merges inside a channel). This needs an
   operator order (`operating-rules.md` §5). `owned_merge_can_happen` holds for the deployed pool only with them.

**Findings of 2026-10-26/27** (lab book): `2026-10-27-gonitro-reclaim-dos.md`, `2026-10-27-payout-griefing.md`,
`2026-10-27-nitro-liveness-game-defects.md`, `2026-10-26-utxo-channel-formalizations.md`, `2026-10-27-bend2-to-evm.md`.

## 2. Other tracks

- **Paused (operator ruling, 2026-10-26).** All E11 work waits until the Lean work is complete. Do not queue an E11
  pebble.
- **F4 collateral** is replaced by the Lean work.
- **F3 governance** (husband `HusbandF3j`):
  - `arp-6cc` (E2) and `arp-445.1` (E5) are closed, but armada-poc `integrated` has no merge commit for them. The
    husband must find out why.
  - `arp-acf` (E3 follow-up) and E7 (treasury ARM in the payout divisor) wait for operator rulings.
  - G7b-spend `arn-d62` failed its gate on 2026-10-21. Its worktree is kept.
  - Open: the replacement of the live armada-1 deployment and the move of ARM holders.
- **F2** (husband `HusbandF2c`, also E11): N9 waits for N3. **F5 client:** waits for N3.
- **Dispatch.** Before you queue work, start the supervisor and the watcher (`operating-rules.md` §4). On 2026-10-27
  neither ran.
- **Gates.** armada-circuits has no `.exophial/gate/Dockerfile` on `integrated`. It needs one before its next gated
  merge. The armada-sdk Dockerfile (`c2a04b1`) is the closest pattern. The armada-poc gate omits `test:config`.

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
  - `docs/lean/README.md` and `docs/lean/writing-lean.md` pin `adjudicator-model-v4` and "no Mathlib".
- **Cleanup in armada-nitro.** Old worker worktrees under `.worktrees/`, `wip/arn-*` and `hand/*` branches, and the
  `backup/main-before-reset-2026-10-1*` branches.

## 4. Questions to the operator

- `supervisor.py` replaces scheduling that exophial lacks (host-wide limits, quota, local versus cloud routing). Move
  it into exophial as a feature?
- Give the S7 proof gate job to a husband?
- mtm production: who owns the failing reconcile, and when is the cron line enabled again?
- Prune dangling Docker images? Remove marks from `pool/27b-class` and `pool/long`?
