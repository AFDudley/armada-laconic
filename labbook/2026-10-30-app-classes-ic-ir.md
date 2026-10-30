# 2026-10-30 — Compliant apps, desirable apps, and what the Liveness theorems may assume

Session: armada-laconic coordinator, discussion with the operator, 2026-10-30. This entry records terms, evidence and
findings. It does not change a theorem type. Lean changes go to the operator for approval separately.

## 1. Trigger

- `finalizable_or_two_enabled` was stated for every app. Opus found a counterexample (`appsAll`, an app that supports
  every state; `docs/counterexamples/finalizable_or_two_enabled.lean`, merged in `48d483e`).
- The operator ruled to constrain the evaluated app, not the channel. The coordinator proposed
  `Rules.NoUnilateralSupport`, and the operator approved it (`labbook/2026-10-29-machine-types-for-approval.md`,
  "App property"; merged `48d483e`).
- Opus then found a second counterexample that satisfies `NoUnilateralSupport` (`appsU`;
  `docs/counterexamples/finalizable_or_two_enabled_happ.lean` on `worker/aln-6ca`, `dbcafcc`). The coordinator
  proposed a stronger property (`noStall`, a turn bound). The operator stopped this and asked what makes an app
  desirable at all.

## 2. Terms (operator, 2026-10-30)

- **valid** keeps ForceMove's meaning only: a single move that passes the framework's checks (ForceMove §3.1.4,
  ForceMove:735-754). It is not used for apps.
- **compliant app**: an app that meets the `IForceMoveApp` interface, and nothing more. In the model, any `Rules`
  value (`ArmadaLean/Nitro/Channel.lean:135-136`): a pure `Bool` function of the fixed part, the support proof and the
  candidate. Source of the word: go-nitro `TrivialApp.sol` and `VirtualPaymentApp.sol` ("complies with the
  ForceMoveApp interface"), and `IForceMoveApp.sol` ("ForceMove-compliant state channel").
- **desirable app**: a compliant app that is incentive compatible (IC) and individually rational (IR) at every state,
  for every role.
  - IR at every state: from any state, an online honest participant can enforce at least what its holdings give it.
    This is ForceMove's "the capability to reclaim the fair amount of funds according to the current state of the
    game" (ForceMove:410-416) and its definition of security (ForceMove:125-130).
  - IC at every state: at no state does a participant gain by deviating from the rules or by aborting. This is
    ForceMove's "game-theoretically sound" (ForceMove:339-347).
  - IR and IC are standard mechanism-design terms. Their textbook definitions were stated from memory in the
    discussion, not from a source read in the session.
- **role**: the position of a participant in an app, for example payer and payee. A guarantee belongs to a role.
- **holdings**: the signed states and vouchers a participant holds. IR is measured against them.

Being compliant says nothing about being IC or IR. A Liveness claim that holds only for desirable apps must say so.

## 3. Evidence: what the sources say makes an app acceptable

| Source | What it says | Where |
|---|---|---|
| ForceMove 2018 | An app is a game library `validTransition`, which "must be pure – it can only depend on the properties of states passed in". The framework adds the support shape: `n` consecutive valid moves, each one turn higher and signed by its mover. | ForceMove:957-966, 744-778, 1340-1354 |
| ForceMove 2018 | A player may sign several moves at one turn ("alternative moves"); the next player chooses. | ForceMove:782-794 |
| ForceMove 2018 | "No rational player would fund the game until everyone has agreed to the proposed initial" state. | ForceMove, footnote 4 |
| Nitro 2019 | §4.4 states its results "for a ForceMove channel". "The consensus game is a ForceMove application." | Nitro:633, 662-664 |
| SATP 2022 | "An application is free to define support in an arbitrary manner", including unilateral support. "Care must be taken to ensure application rules encode the fair distribution of assets." | SATP §2.2, footnote 7 |
| go-nitro | `stateIsSupported` "Must revert or return false when invalid support proof and a candidate are supplied". "Invalid" is not defined. `view` is allowed; `pure` is recommended "to make security analysis easier". | `interfaces/IForceMoveApp.sol` |
| go-nitro | `TrivialApp` "complies with the ForceMoveApp interface and allows all transitions, regardless of the data. Used for testing purposes." | `TrivialApp.sol:6-19` |
| docs.statechannels.org | "Nitro is a very open protocol … Each application gets full control over when a state can be considered supported." | "Execution Rules" page |

Finding: no source defines when an app is acceptable. The newest paper allows arbitrary support. Each source result
holds only for the apps it names: Nitro §4.4 for ForceMove apps, the SATP theorems for the consensus app and the
virtual-channel app.

## 4. The apps in go-nitro, by support rule

| App | Support rule |
|---|---|
| `ConsensusApp`, `InterestBearingApp` | `Consensus.sol`: an empty proof; the candidate is signed by every participant |
| `CountingApp`, `examples/HashLockedSwap`, `examples/SingleAssetPayments` | `StrictTurnTaking.sol`: each participant signs the one state it moves, at consecutive turns |
| `VirtualPaymentApp` | its own rules: unanimous prefund (turn 0) and postfund (turn 1); a redemption (turn 2) signed by Bob alone, with a voucher Alice signed for an amount `x`, and outcome Alice − `x`, Bob = `x` |
| `TrivialApp` | supports everything; testing only |

Unanimous support is the strictest case of turn-taking: all `n` signatures on one state instead of one signature on
each of `n` consecutive states.

## 5. Counterexamples: compliant apps that are not desirable

| Case | Status | What fails |
|---|---|---|
| `appsAll` (supports every state) | Lean counterexample, merged `48d483e` | IR: either side checkpoints at a higher turn in every round; nothing finalizes |
| `appsU` (supports any candidate once each participant signed some state in the proof) | Lean counterexample, `worker/aln-6ca` `dbcafcc` | IR: satisfies `NoUnilateralSupport`, and still lets either side stall |
| Replayable ReceiptApp (Alice 0, Bob 5; a certified receipt pays Alice 1; the receipt does not name its turn) | discussed, not formalized | IR for Bob: one receipt replayed five times pays 5 |
| Unbound griefing (the unanimous rule without its per-turn signature check: any candidate whose outcome `p` signed, at any turn) | discussed, not formalized | IR for termination: no outcome `p` did not sign can finalize, but the others re-post an old outcome at ever higher turns, so nothing finalizes |
| Rock-Paper-Scissors without the default-winner rule | ForceMove:339-347 | IC: Alice gains by stopping once she knows she lost |

The replayable ReceiptApp must stay constructible. Games are Turing complete, so compliance cannot exclude cheating
apps (operator, 2026-10-30). The model must therefore separate compliant from desirable, and state which one each
claim is about.

## 6. Roles: the one-way payment app

In `VirtualPaymentApp`, Alice is the payer and Bob the payee.

- Bob signed the postfund, in which Bob gets 0. A definition of safety as "only an outcome `p` signed can finalize"
  counts a finalization that pays Bob 0 as safe for Bob, although Bob holds a voucher worth `x`. That definition is
  wrong for the payee.
- Bob's real guarantee is a value: an online Bob can always answer with a redemption and obtain at least his best
  voucher.
- So a guarantee depends on the role and on the holdings. Nitro's own notions express this: what `p` can force
  (`Finalizable`, and value through `CanObtain`) and what `p` cannot prevent (`Enabled`).

## 7. Nitro requires an online participant

The papers never use the word "online". The requirement is in the definitions: every guarantee is an unbeatable
strategy that `p` carries out (Nitro §4.3, Nitro:525-542), and "it is possible for any party to apply any operation,
O, on-chain before time t + ε" (Nitro:535-537). SATP:165-166: "Any peer can successfully record transactions on the
blockchain in O(1) time." The model matches: every Liveness theorem is an `UnbeatableIn` strategy of `p` with the
time bound `T.ε`.

## 8. Where the model holds player behaviour

| What | Where | What it constrains |
|---|---|---|
| `p`'s possible moves | `HonestMove`, `HonestMove.Feasible` (`ArmadaLean/Nitro/Strategy.lean:92-132`) | only the calls (`Op.HonestFeasible` refuses `.deposit` and `.tick`); `Reaction.Feasible` places no condition on `sends` |
| `p`'s strategy | `UnbeatableIn apps T allowed G` (`Strategy.lean:471`) | the move tree; `allowed` limits the move class |
| The other participants | `Threat`, `Threat.RoundSeq`, `Threat.Round`, `Bounds`, `Capable`, `Assumptions` (`Strategy.lean:179-226`, `:437-461`) | no forged signature of `p`; any unforged call or delivery; ticks within ε |
| Past behaviour | the `Nitro.participant` and `Nitro.network` steps (`ArmadaLean/Nitro/Reachable.lean:164`, `:486`) | the same honest moves and adversary moves; an honest send is gated only by `Network.VoucherSendable` |

`funding_point_universally_finalizable` is false as stated (Lean counterexample on `worker/aln-cc8`, `3235226`)
because a reachable history may hold two states that `p` sent at one turn: ForceMove's alternative moves. This is a
statement gap (`docs/process.md`, "Model defect or statement gap?"), not an app property. The candidate fix is a
hypothesis on `p`'s sent history, not a change to the step rules.

## 9. What the coordinator did wrong in this discussion

- It used go-nitro client code (`AddSignedState`, `Merge`) as evidence for honest behaviour, before it read
  ForceMove §3.1.5. `docs/lean/writing-lean.md` §10 puts the paper first.
- It invented app properties (`NoUnilateralSupport`, `noStall`, the turn bound) with no source, as stand-ins for IR.
  Each was weaker than what it stood for, and Opus found a counterexample to the first one.
- It defined "safe for `p`" before it asked which participant and which holdings. The one-way payment app shows the
  definition was wrong.
- It built classification tables from its own axes and patched them after each new case, instead of starting from
  the sources' definitions.

## 10. Open, for the operator

1. A definition of IR and IC at every state, in Lean, relative to each role's entitlement from its holdings. Nitro's
   `Value`, `CanObtain` and `Finalizable` are the building blocks. This needs a design.
2. Which theorems take that property. At least `finalizable_or_two_enabled` and `single_enabled_is_finalizable`.
   `Rules.NoUnilateralSupport` and its two lemmas are then deleted.
3. The statement of `funding_point_universally_finalizable`: a hypothesis that `p` sent no two states on `c` at one
   turn below `c.n`.
