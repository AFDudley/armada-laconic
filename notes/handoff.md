# Handoff — Armada build programme

This note holds only open questions, what is to be done, and what is pending (operator ruling, 2026-11-07). A
completed item is removed from it. To see what is done, read the issue bus of each repo.

## Documents

All paths are absolute. Read the goal document first. Open the others when a task needs them.

| What | Document |
|---|---|
| **Read first.** The process: four layers, the strata, the one rule, the measured anti-patterns | `~/code/git_puller/repos/armada-laconic/docs/process.md` |
| The goal, the scope and the order of the proof work: the pool first, then armada-nitro; governance out | `~/code/git_puller/repos/armada-laconic/labbook/2026-11-06-proof-scope-and-order.md` |
| The Lean contract library: evidence, selection and decisions E1, E2, D0-D3 | `~/code/git_puller/repos/armada-laconic/labbook/2026-11-11-library-selection.md` |
| Symbolic execution and solc in the chain (option A; option B ruled out) | `~/code/git_puller/repos/armada-laconic/labbook/2026-11-11-symbolic-execution-and-solc.md` |
| The concrete specification of the pool (strata 1-5, stratum 6 contents) | `~/code/git_puller/repos/armada-laconic/labbook/2026-11-11-concrete-specification-design.md` |
| Programmatic compiler swap (other solc versions, Solar): compiler profiles, operator ruling 2026-11-11 | `~/code/git_puller/repos/armada-laconic/labbook/2026-11-11-programmatic-compiler-swap.md` |
| The build plan (S-items and their status) | `~/code/git_puller/repos/armada-laconic/engineering/build-plan.md` |
| The operating rules (models, limits, routing) | `~/code/git_puller/repos/armada-laconic/engineering/operating-rules.md` |
| All decisions | `~/code/git_puller/repos/armada-laconic/labbook/` |
| The proof toolchain plan (items A-E) | `~/code/git_puller/repos/evm-proof-toolchain/docs/plan.md` |
| The Lean model to Stele translator (superseded: Stele dropped 2026-11-08) | `~/code/git_puller/repos/evm-proof-toolchain/docs/translator.md` |
| The Lean model of the pool and of Nitro | `~/code/git_puller/repos/armada-lean/` (namespace `ArmadaLean`) |

Owners: pane w3:p5 armada-lean, and the evm-proof-toolchain items of w3:p4 (that session ended 2026-11-07; the operator
has not named a new owner); pane w4:p1 lean-mend; pane w6:p1 exophial, and stage ML of exophial `docs/merge_plan.md`.

## Open questions to the operator

1. AxiomFix (w4:p1) has three edits, E1-E3, that give back the v4.29 axiom sets of all 43 axiom findings of lean-mend
   run 10. The exact text is in evm-proof-toolchain `.scratch/axiom-fix/edits.md`. Apply them, run the gate, merge,
   and start lean-mend run 11?
2. Does the E11 pause of 2026-10-26 still hold? The E11 Aave yield gate merged on armada-nitro (bus `arn-0a2`).
3. F3 governance is out of scope since 2026-11-06. Close or keep: `arp-6cc` and `arp-445.1` (closed, but armada-poc has
   no merge commit for them); `arp-acf` and E7 (they wait for rulings); the replacement of the live armada-1
   deployment and the move of ARM holders.
4. Give the S7 proof gate job to a husband?
5. mtm production: who owns the failing reconcile, and when is the cron line enabled again?
6. Prune dangling Docker images? Remove marks from `pool/27b-class` and `pool/long`?
7. Branch cleanup: armada-poc's fork has 153 branches with no local copy, armada-sdk's fork 19. Unclassified branches
   and worktrees with uncommitted changes in armada-nitro, armada-poc and armada-circuits: list them with
   `git branch -a` and `git worktree list` in each repo.
8. `armada-laconic/.scratch/supervise/supervisor.py` runs (pid 1882621, started 2026-10-31, last sample
   2026-11-07 03:17). It is a second scheduler that reads exophial internals. w6:p1: leave it running, queue no work
   into it; stopping it is the operator's decision.
9. evm-proof-toolchain `integrated` is 353 commits ahead of the pushed `main` (`e148eb42`), because
   `branch.integrated.remote` is not set. Operator, 2026-11-08: do not push it now. When to push it?

## To be done

1. After question 1 (owner w3:p5): apply E1-E3 on a branch, run `scripts/gate.sh`, merge, then lean-mend run 11. Then
   update the status of plan item A2 in evm-proof-toolchain `docs/plan.md`.
2. Owner w3:p5: add `armada-stele/` to the evm-proof-toolchain worker image and gate, on its own env-change branch.
   armada-stele is pinned to armada-lean `9a685f58` (evm-proof-toolchain `3e671aeb`) and builds with no change.
3. Owner w3:p5: translator M3 (`transact`) as exophial pebbles: M3a tree batch insertion and root query, M3b the owned
   `transact`, M3c its generator, M3d `transact` with unshield. Each acceptance includes a non-vacuity theorem.
4. Translator open items (`docs/translator.md` §9 and §10): the trusted call hypotheses have no concrete witness yet
   (step-2 work); the Solidity rewrite of the pool must compute the fee with full precision (`mulDiv`) and must not
   record the empty tree's root.
5. The pool's executable form is complete in armada-lean (PE1-PE5 and the fee-admin setter proved, `sorry_count` 0,
   tip `c3e423e2`). The next step of the order is the Solidity rewrite of the pool from `Pool.apply`.
6. Jaune/Blanc evaluation (`.scratch/jaune-eval/`, briefs there). V1-V4 pass for Jaune/Blanc: 4968/4968 and 3692/3692
   execution-spec fixture files; all 7 Blanc exhibits rest on Lean's three axioms; gas is an exact equality; armada-lean
   on v4.35.0 fails one module. EquiVM's ERC20 proof has 2007 `native_decide` axioms. EVMLean's Cancun count (341
   failures) is not clean: the host lacked its Python crypto modules. Next: an evaluation image built on the
   evm-proof-toolchain worker image plus those modules and solc; the EVMLean rerun in it; then V5 (`setFeeRatio`
   bytecode against `Pool.apply` in Blanc) and V6 (the same task in EquiVM and Stele).
7. Proof path (lab book `2026-10-20-nimbus-fixturenet-and-proxy-proof.md`): analyse whether the plan needs per-log
   proofs (an unchecked script and data are in `.scratch/PerLogAnalysis/`); add the proof as a job in the
   armada-nitro gates for the S7 acceptance.
8. armada-nitro E16: one frozen asset blocks every payout of a close (`script/CustodianFrostScenario.s.sol`,
    `concludeAndTransferAllAssets`). The fix is part of E11. The E16 brief's scope statement about E11 contradicts
    itself.
9. armada-nitro E17, not scheduled: `_executeSingleAssetExit` ignores a `withdrawHelper` allocation; `reclaim`
    credits only target allocations 0 and 1. No test confirms either.
10. Selective disclosure (page 16, ADR-0020): no hop is checked against events of real transactions; the hub's
    EIP-1271 check ran only against a stand-in. The production proof system and ceremony are open.
11. Hub tracking of cross-chain swaps: no reference hub daemon or design page exists
    (`engineering/shielded-nitro-bridge-design.md` §4-§6, ADR-0015).
12. Documentation: `engineering/build-plan.md` S3 omits N9; S5 uses the old N1-N8 sizes. The site link to this note
    is broken because `notes/` is not rendered.
13. The armada-poc gate omits `test:config`. armada-circuits `main` is behind `integrated`
    (`git log --oneline main..integrated` in armada-circuits).
14. Nitro, when its work starts: rename `Nitro.Address E` to `Nitro.Destination E` (it is a tagged destination, not an
    address), and use `ArmadaLean.Address` for its external case and for `Participant.key` (go-nitro
    `INitroTypes.sol:14` `address[] participants`). The first Armada client that builds channels runs go-nitro's
    `FixedPart.Validate` check (risk R10).

## Pending (waits for someone else)

1. w4:p1 (lean-mend): a rule trigger for axiom findings, so a later run makes edits like E1-E3 itself (plan item E1).
2. w6:p1, stage ML: `exo-748.1` (the Lean package) is parked; its fix is `exo-503` (law L2). `exo-748.2`, `exo-748.3`,
   `exo-c57` and `exo-313` wait behind it. armada-lean has no open spec theorem now, so the next statement commit
   with an open theorem gives w6:p1 its replay subject.
3. Pebble `exo-041` (parked: the gate container drops `.claude`; law L3, epic `exo-4d4d`): skills come from one source,
   the installed tool. Until it merges, a worker in armada-lean or evm-proof-toolchain cannot read
   `exophial-worker-orient`.
4. F2 (husband `HusbandF2c`): N9 waits for N3. F5 client waits for N3.
5. The WASM proxy stall issue in armada-client: the operator approves the issue body (draft in session `01a14aa3`).
6. w6:p1, pebble `exo-6dd.2`: the gate and worker containers read one host Lake artifact store. Law L3: exophial's
   image-store owner is its only writer; the gate mounts it read-only; a worker never writes a store a gate reads.
   It starts after exophial `integrated` is green again. w6:p1 sends the merge sha.
