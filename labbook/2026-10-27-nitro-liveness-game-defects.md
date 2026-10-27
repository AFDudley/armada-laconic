# Nitro liveness: defects in the game model and in the statements

Date: 2026-10-27. Branch `nitro-proofs` at `aa52443` (`AFDudley/armada-lean`, worktree
`~/code/git_puller/repos/armada-lean-nitro-proofs`). No file in the build changed. No theorem type changed.

The approved game form is `docs/nitro-theorems.md:64` ("The game (Nitro §4)", approved 2026-10-27): one adversary
controls every other participant; one round of `Unbeatable` is a finite feasible adversary prefix, then one honest
move; `Assumptions.unimpeded` forbids the adversary to move the clock to a pending deadline of the honest
participant; only the honest `wait` reaches a deadline; `Prevents` is a safety property over `Run`s.

Part 1 lists defects of the game definitions (D1-D8). A repair of D1-D8 changes definitions. Part 2 lists gaps in the
statements of single theorems (G1-G13). A repair of G1-G13 changes theorem types, so it needs operator approval.
Part 3 is the table of all 35 theorems.

## Lean checks

All checks are outside the build, in the worktree's `.scratch/`. Command:
`~/.elan/bin/lake env lean .scratch/<file>` in the worktree.

- `.scratch/GameDefects.lean`: one check for each of D1-D8 (11 declarations). `.scratch/GameDefectsAx.lean` prints
  their axioms: `propext`, `Classical.choice`, `Quot.sound` only. No `sorryAx`.
- `.scratch/LivenessCex.lean`: the negation of two Liveness types, through `type_of%`:
  `participant_can_prevent_cycles_false` and `consensus_enabled_outcomes_false`. Axioms: `propext`, `Quot.sound`.
  The file defines the canonical adversary `T` (every move of another participant that forges no signature of `p`,
  and ticks short of `p`'s pending deadlines) and proves `Assumptions apps p T`.

A claim marked `[INFERENCE]` has a written argument and no Lean proof.

## Part 1. Defects of the game definitions

### D1. `Threat.TransfersOnly` contradicts `Assumptions.capable`

- Definitions: `Threat.TransfersOnly` (`Liveness.lean:15-17`); `Assumptions.capable` and `capableDeliver`
  (`Strategy.lean:120-125`).
- Fault: `capable` makes every operation of another participant feasible when it forges no signature of `p`. A
  deposit has no proof states, so it is feasible. `TransfersOnly` says every feasible move is a transfer or a tick.
  The two together are `False`.
- Against the approved form: the form makes the adversary capable of every move of the other participants. Nitro:779-781
  restricts both players to transfers for the Turbo analysis ("strategies and counter-strategies involving transfer
  operations only"). The model restricts only the adversary, and that restriction cannot hold.
- Effect: `turbo_redistribution_unbeatable`, `turbo_value_upper_bound`, `turbo_value_achievable`,
  `turbo_value_is_value` are vacuous. Each has `A` and `hT`.
- Lean: `GameDefects.d1_transfersOnly_false`.
- Repair: split `Assumptions` into the bounds (`unforgeable`, `unforgeableDeliver`, `notHonest`, `unimpeded`) and the
  capabilities (`capable`, `capableDeliver`, `capableTick`). The Turbo theorems take the bounds, and a capability
  for transfers and ticks only. The honest moves in the Turbo game are transfers and `wait` only (Nitro:779-781 names
  both players). This changes the type of the four theorems.

### D2. `hsetup` and `hfinal` quantify over every `SignedState` value

- Hypotheses: `satp_lemma_3_1` (`Liveness.lean:423-426`); `satp_theorem_3_2` (`Liveness.lean:476-477`).
- Fault: a signed state is data with a signer list (`Channel.lean:77-80`). `∀ ss : SignedState, ss.ByAll → …` ranges
  over all data, including states that no peer signed. The state with the empty allocation and signers
  `r.V.participants` breaks `hsetup`. The final state with the empty allocation breaks `hfinal`.
- Against the approved form: `docs/nitro-theorems.md:62` makes forgery a game rule (`Assumptions.unforgeable`), not a
  property of the data. The hypotheses read as a claim about honest signers but speak about all data.
- Effect: `satp_lemma_3_1` and `satp_theorem_3_2` are vacuous.
- Lean: `GameDefects.d2_hsetup_false`, `GameDefects.d2_hfinal_false`.
- Repair: add a list `signed : List SignedState` (every state the peers signed) and a hypothesis that each step of the
  history submits proof states from `signed` only. `hsetup` and `hfinal` then range over `signed`. This changes the type
  of both theorems.

### D3. An honest deposit is free

- Definitions: `HonestMove.Feasible` (`Strategy.lean:57-60`), `NetworkState.applyHonest` (`Strategy.lean:64-69`),
  `Adjudicator.deposit` (`Adjudicator.lean:87-95`), `CanObtain` and `Value` (`Strategy.lean:190-200`).
- Fault: the honest participant may play `.chain (.deposit a e x)` for any `x`. No wallet is debited. `p` can raise
  any channel balance without bound.
- Consequence `[INFERENCE]`: `p` can obtain any balance `Y` at any participant address, with no `wait` move. The
  strategy: take a fresh channel `c2 = ⟨[p], N, ⟨0⟩, 0⟩` (fresh `N`); deposit into `c2` until it holds `Y` (the
  adversary cannot finalize `c2`: only `p` may challenge, and a conclude needs `p`'s signature on a `c2` state); send the
  final state `st` with outcome `[q : Y]`; conclude `c2`; transfer `c2` to `q`. So `CanObtain` holds for every `x`,
  and `Value apps T net x` is false for every `x`.
- Against the approved form: the form gives the honest participant "calling on-chain operations". Nitro:482-484
  defines value as what `p` can obtain; Nitro:779-781 ignores deposits. With free deposits there is no maximum.
- Effect: `deposit_bounds_value`, `payout_to_other_affects_only_owner`, `subchannel_open_value_preserving`,
  `subchannel_close_value_preserving` are vacuous (`Value` is false on both sides). `directly_funded_pays`,
  `finalizable_funded_pays`, `ledger_partial_withdrawal` (first conjunct) hold by the self-channel strategy.
  `ledger_offload_funds_subchannel` and the `FundedFor` conjuncts of `ledger_partial_withdrawal` hold by deposits into
  the target channel (the outflow of a finalized channel is bounded by its outcome total).
- Lean: `GameDefects.d3_deposit_free` (the deposit is feasible and adds `x`). The consequence is not formalized.
- Repair: make the honest deposit infeasible in the game: `HonestMove.Feasible net (.chain (.deposit ..)) := False`.
  `deposit_bounds_value` already applies its deposit outside the game. A wallet field in `NetworkState` is the other
  option; it changes the meaning of `Value` (wallet coins become obtainable). The first option is the correct one: it
  is the Nitro:779-781 restriction, and it changes one definition.

### D4. The adversary moves the clock with the chain operation `tick`

- Definitions: `Op.tick` (`Adjudicator.lean:220`, applied at line 246), `NetworkState.applyAdv` (`Strategy.lean:75`),
  `Assumptions.unimpeded` (`Strategy.lean:115-116`), `Assumptions.capable` (`Strategy.lean:120-122`).
- Fault: `unimpeded` bounds only `AdvMove.tick`. `capable` makes `.chain q (.tick dt)` feasible for every `dt`, because
  `Op.tick` has no proof states. The adversary moves the clock past every pending deadline.
- Against the approved form: "only the honest `wait` reaches a deadline" (`docs/nitro-theorems.md:64`).
- Effect: with D4 the adversary finalizes any challenge it registers before `p` can answer. It touches
  `late_observer_can_checkpoint`, `funding_point_universally_finalizable`, `latest_state_finalizable`,
  `conclusion_proof_finalizable`, `concluded_universally_finalizable`, `satp_peers_can_finalize`,
  `satp_intermediary_before_round2`, `satp_end_user_recovers`.
- Lean: `GameDefects.d4_chain_tick`.
- Repair: `capable` excludes `Op.tick`, and `unimpeded` also bounds `.chain _ (.tick dt)`. `Op.tick` stays for
  `Reachable` (Safety needs time steps). go-nitro has no tick operation; time is `block.timestamp`.

### D5. `Run` never forces `p` to move, so `Prevents` ignores the strategy

- Definitions: `Run` (`Strategy.lean:149-155`), `Prevents` (`Strategy.lean:159-161`), `Enabled` (`Strategy.lean:177-179`).
- Fault: every adversary-only run is a `Run` of every strategy `σ`. So `Prevents G net` holds exactly when no
  adversary-only run reaches `G`; the strategy `fun _ => none` is the best one. A second fault follows from the time
  rule: in an adversary-only run the clock never reaches a deadline of `p` (D4 aside), so `p` "prevents" every timeout
  outcome by never playing `wait`.
- Against the approved form: Nitro:641-644 "p has a strategy to prevent". In the model a strategy never helps.
- Effect: `Enabled` means "the adversary alone can finalize it". `finalizable_is_enabled` is false: on a fresh channel
  `[p]`, `p` finalizes by `send` then `conclude`, and the adversary alone cannot. It touches `finalizable_or_two_enabled`,
  `single_enabled_is_finalizable`, `consensus_enabled_outcomes`, `participant_can_prevent_cycles`.
- Lean: `GameDefects.d5_adv_run`, `GameDefects.d5_prevents_iff`.
- Repair: a strategy is total (`NetworkState p → HonestMove`), and a run is a sequence of rounds as in `Unbeatable`
  (an adversary prefix, then `σ`'s move). For the time fault, the correct rule is: after `p`'s move, the adversary may
  move the clock past the deadline of any challenge that was pending before that move (`p` had one move to answer it,
  Nitro:537-539). This changes "only the honest `wait` reaches a deadline" for `Prevents` runs, so it needs a ruling.

### D6. `Finalized` reads the stored outcome, and `transfer` rewrites it

- Definitions: `NetworkState.Finalized` (`Strategy.lean:164-166`); `Adjudicator.transfer` (`Adjudicator.lean:121`) and
  `Adjudicator.reclaim` (`Adjudicator.lean:146`) write `withOutcome` of the reduced allocation.
- Fault: `Finalized c Ω` is true only until the next transfer from `c`. A strategy can "finalize" `Ω` and another
  strategy can then "finalize" the reduced outcome.
- Against the approved form: Nitro:602-603 `⟦χ ↦ Ω⟧` is the finalized outcome. A payout does not change which outcome
  finalized.
- Effect: `finalizable_unique` is false (`c` finalized with `[p : 1]` and holding 1: `[p : 1]` is finalizable for `p` at
  once, and `[p : 0]` for `q` by one transfer). `conclusion_proof_finalizable` and `concluded_universally_finalizable`
  are false (the adversary concludes first and transfers `[p : 1]` out, so `[p : 1]` never reads again).
  `consensus_enabled_outcomes` is false (formalized, see Part 2, G4).
- Lean: `GameDefects.d6_finalized_mutates`.
- Repair: `ChannelStatus.finalized` keeps the outcome at finalization in its own field. `transfer` and `reclaim`
  rewrite a second field, the remaining allocation. `Finalized` reads the first field; payouts read the second. The
  weaker repair `Finalized c Ω := ∃ Ω', finalOutcome = some Ω' ∧ Redistributed Ω Ω'` does not save
  `finalizable_unique` (the same example still gives two outcomes). The Safety reading lemmas of `ChannelStatus`
  change with the first repair.

### D7. Proof signatures that no game rule checks

- Definitions: `validMove`, `validRun`, `forceMoveRules` (`Channel.lean:142-164`); `validChallenge` and
  `validResponse` (`SATP.lean:43-66`, voucher signers at lines 50 and 63).
- Fault 1: `validRun` checks `byMover` of the second state of each pair only. The first proof state needs no signature.
  With `vt` false, a participant alone supports a final state at its own turn, with any outcome.
- Fault 2: `voucher.signers` is data inside `appData`. `Assumptions.unforgeable` covers `SignedState.signers` only. Any
  participant writes a voucher "signed" by the other end user.
- Against the approved form: ForceMove:1340-1354 needs each of the `n` states signed by its mover (go-nitro
  StrictTurnTaking has the same rule). SATP:703 `voucherSignedByAll` checks signatures.
- Effect: `funding_point_universally_finalizable` is false (with `n = 2`, `vt` false and D4, the adversary registers
  `[⟨s0, []⟩, ⟨f1, [q]⟩]` with outcome `[]` and lets it time out) `[INFERENCE]`. It touches
  `satp_peers_can_finalize` and the SATP lemmas.
- Lean: `GameDefects.d7_unsigned_head`, `GameDefects.d7_voucher_unchecked`.
- Repair: `forceMoveRules` also checks `byMover` of `proof.head?`. `unforgeable` covers every signature a proof carries,
  with the voucher signers included; `p` signs a voucher only through a `send`.

### D8. `Views.Consistent` accepts forged received signatures

- Definition: `Views.Consistent` (`Strategy.lean:208-210`).
- Fault: the definition requires that what `p` sent `q` received. It does not require that a signature of `q` that `p`
  received is on a state `q` sent.
- Effect `[INFERENCE]`: `channel_finalizability_cases` is false. `p` received `q`'s signature on a final state `F` with
  outcome `[]` that `q` never sent; `q` received `p`'s signature on a final state `F'` with outcome `[p : 0]` that `p`
  never sent. Each concludes its own state against its own adversary, so two participants have different
  finalizable outcomes and the channel is open: all four cases fail. The same views break `finalizable_unique`.
- Lean: `GameDefects.d8_consistent_forged`.
- Repair: add `∀ p q, p ≠ q → ∀ r ∈ (v p).received, q ∈ r.signers → ∃ s ∈ (v q).sent, s.state = r.state`.

## Part 2. Gaps in single statements

Each gap changes a theorem type. Each counterexample uses the canonical adversary of `.scratch/LivenessCex.lean`.

- G1. No history for the chain. `net.chain` is any `Adjudicator`, so a channel can be challenged at the turn of `p`'s
  latest state with a state nobody signed. `latest_state_finalizable` (`Liveness.lean:86`): `c` challenged at
  `ss.turnNum` with outcome `[]`; `p` can neither raise the record nor conclude. `satp_peers_can_finalize`
  (`Liveness.lean:439`): `V` challenged at turn 7; every candidate `p` can support has turn at most 3.
  `satp_intermediary_before_round2` (`Liveness.lean:487`): both ledgers challenged at their latest turns with outcome
  `[]`. `[INFERENCE]` for all three. Repair: a hypothesis on `statusOf c` (no record at or above `ss.turnNum` other
  than `ss.state`), or a history hypothesis that ties each registered state to the signatures of `p`.
- G2. `≤` in the latest-state hypotheses. `hlatest` (`Liveness.lean:90`) and `HoldsLatestLedger` (`SATP.lean:155`)
  allow a second state of `p` at the same turn. The adversary challenges with it; `p` cannot override at an equal turn
  (ForceMove.sol:48-50). Repair: `t.state.turnNum < ss.state.turnNum ∨ t.state = ss.state`.
- G3. No `0 < c.challengeDuration` in `conclusion_proof_finalizable`, `late_observer_can_checkpoint`,
  `concluded_universally_finalizable`, `consensus_enabled_outcomes`, `single_enabled_is_finalizable`. A challenge
  with duration 0 finalizes at once (`Adjudicator.lean:168`, `Channel.lean:228`).
- G4. `hsent` constrains final states only (`conclusion_proof_finalizable`, `concluded_universally_finalizable`) or
  states at or below one turn (`consensus_enabled_outcomes`). With G3 or D4 the adversary registers any other state
  `p` signed. Formalized: `LivenessCex.consensus_enabled_outcomes_false`. `p` holds the all-signed non-final `st`
  with `[p : 1]`, the duration is 0, the channel holds 1; the adversary challenges with `st` and transfers, so
  `[p : 0]` is enabled and is neither `Ω1` nor `Ω2`. This example also uses D6.
- G5. `late_observer_can_checkpoint` (`Liveness.lean:106`) allows `p` to have signed a state above `ss`. The adversary
  challenges with it at the higher turn; `p`'s checkpoint with `ss` then reverts, and the channel never reopens
  `[INFERENCE]`. Repair: the `hlatest` hypothesis of `latest_state_finalizable`.
- G6. `satp_end_user_recovers` (`Liveness.lean:506`). The suspicion holds: `HoldsLatestLedger` lets Alice have signed an
  earlier FINAL state `F` of `L0`. Bob concludes `L0` with `F` (outcome `[alice : 0, bob : 0]`, no guarantee for `V`)
  before Alice moves; `L0` stays finalized with an outcome that owes Alice 0, and the goal needs 1 (`a0 = 1`, `b0 = 0`,
  `a = 1`, `b = 0`) `[INFERENCE]`. A second gap: `hfin` lacks `a + b = r.a0 + r.b0`, without which `reclaim` reverts
  (MultiAssetHolder.sol:365) and the guarantee entry for `V` never leaves `L0`. Repair: no final sent state of `L0`
  other than the latest, and the hypothesis `a + b = r.a0 + r.b0`.
- G7. `satp_theorem_3_2` (`Liveness.lean:462`). The suspicion holds in substance: an intermediary that signed an earlier
  FINAL state of `Lm` or `Lp` loses to a conclude with it. Today the theorem is vacuous by D2, so no counterexample
  exists. After the D2 repair it needs the G6 hypothesis and the G1, G2 repairs `[INFERENCE]`. The same holds for
  `satp_intermediary_before_round2`.
- G8. `participant_can_prevent_cycles` (`Liveness.lean:253`): the goal covers channels that `p` is not in. Formalized:
  `LivenessCex.participant_can_prevent_cycles_false` (two channels of `q` alone, concluded with `[Y : 0]` and
  `[X : 0]`). Repair: the goal speaks of cycles through channels of `p` (Nitro:792-794 "any participant in the
  channel").
- G9. `funding_point_universally_finalizable` (`Liveness.lean:172`): `SetupPhaseKeepsOutcome` says a valid setup
  transition keeps the outcome. It does not say `vt` accepts the setup transitions, so no run of setup states need be
  supported. Repair: a hypothesis that `vt` accepts each setup transition (ForceMove:1092-1125).
- G10. `satp_peers_can_finalize`: `hsent` constrains `fundedTotal` only (the first two amounts), so a sent state may
  carry a third entry; `validStart` (`SATP.lean:37-40`) checks no turn, so an all-signed type-None state at any turn is
  a candidate. Repair: `hsent` requires `PairOutcome … ∧ a + b = r.a0 + r.b0`, or `turnNum ≤ 1` with the original
  outcome.
- G11. `directly_funded_pays`, `finalizable_funded_pays`: the first entry of `p` may be a guarantee, which `transfer`
  refuses (MultiAssetHolder.sol:171-172). Today both hold by D3. After the D3 repair both are false for that case
  `[INFERENCE]`. Repair: `o.find p.address = some ⟨p.address, .alloc _⟩`, or a goal in terms of `paidTo`.
- G12. `finalizable_or_two_enabled` (`Liveness.lean:145`): on a fresh channel `[p, q]` where nobody signed, `p` has no
  finalizable outcome and no outcome is enabled `[INFERENCE]`. Nitro:653-656 speaks of a channel past its setup.
  Repair: a hypothesis that `p` holds a supported state of `c` (for example the funding point).
- G13. `single_enabled_is_finalizable` (`Liveness.lean:153`): `p` signed `st` (outcome `[]`) and holds no signature of
  `q`. With duration 0 the adversary finalizes `st` alone, so `[]` is the one enabled outcome; `p` cannot finalize it
  `[INFERENCE]`. Repair: a hypothesis that `p` holds the support of every state it signed.

The two suspicions about weak goals: `directly_funded_pays` and `ledger_partial_withdrawal` are satisfiable trivially
(D3). The coins do not come from `p`'s address (a deposit to a participant address reverts, `Adjudicator.lean:95`).
They come from a free deposit into a fresh channel of `p`, which then pays `p`. The strategy uses no `wait`, so it
also meets `UnbeatableNoWait`.

## Part 3. Status of the 35 Liveness theorems

Status: **sound** (true and the proof exercises the claim; none is proved yet, so each is `[INFERENCE]` by a hand
check); **vacuous** (true for a reason that does not exercise the claim: a false hypothesis, or a defect that lets any
strategy reach the goal); **false** (a counterexample; "Lean" when formalized); **unknown**.

| # | Theorem | Line | Status | Defects and gaps |
|---|---|---|---|---|
| 1 | `directly_funded_pays` | 45 | vacuous (trivial goal) | D3, G11 |
| 2 | `finalizable_funded_pays` | 55 | vacuous (trivial goal) | D3, D6, G11 |
| 3 | `conclusion_proof_finalizable` | 67 | false | D6, D4, G3, G4 |
| 4 | `latest_state_finalizable` | 86 | false | G1, G2, D4, D6 |
| 5 | `late_observer_can_checkpoint` | 106 | false | D4, G3, G5 |
| 6 | `finalizable_unique` | 121 | false | D6, D8 |
| 7 | `channel_finalizability_cases` | 131 | false | D8, D6 |
| 8 | `finalizable_or_two_enabled` | 145 | false | G12, D5 |
| 9 | `single_enabled_is_finalizable` | 153 | false | G13, G3, D4, D5 |
| 10 | `finalizable_is_enabled` | 162 | false | D5 |
| 11 | `funding_point_universally_finalizable` | 172 | false | D7, D4, G9 |
| 12 | `concluded_universally_finalizable` | 183 | false | D6, D4, G3, G4 |
| 13 | `consensus_enabled_outcomes` | 201 | false (Lean) | G4, G3, D6, D5 |
| 14 | `turbo_redistribution_unbeatable` | 218 | vacuous (Lean: D1) | D1; after repair, a channel target can be drained [INFERENCE] |
| 15 | `deposit_bounds_value` | 229 | vacuous | D3 |
| 16 | `payout_to_other_affects_only_owner` | 242 | vacuous | D3 |
| 17 | `participant_can_prevent_cycles` | 253 | false (Lean) | G8, D5 |
| 18 | `turbo_value_upper_bound` | 263 | vacuous (Lean: D1) | D1, D3 |
| 19 | `turbo_value_achievable` | 272 | vacuous (Lean: D1) | D1 |
| 20 | `turbo_value_is_value` | 281 | vacuous (Lean: D1) | D1, D3 |
| 21 | `turbo_order_independent` | 290 | unknown (believed true) | none found |
| 22 | `turbo_value_exists_unique` | 301 | unknown (believed true) | none found |
| 23 | `ledger_offload_funds_subchannel` | 310 | vacuous (trivial goal) | D3 |
| 24 | `subchannel_open_value_preserving` | 320 | vacuous | D3 |
| 25 | `subchannel_close_value_preserving` | 340 | vacuous | D3 |
| 26 | `ledger_topup_reorder_safe` | 357 | sound | none |
| 27 | `ledger_topup_deposit_funds` | 368 | sound | none |
| 28 | `ledger_partial_withdrawal` | 382 | vacuous (trivial goal) | D3, D6, D4 |
| 29 | `satp_outcome_change_keeps_total` | 404 | sound | none |
| 30 | `satp_lemma_3_1` | 420 | vacuous (Lean: D2) | D2; after repair G10 (`validStart` has no turn check) [INFERENCE] |
| 31 | `satp_peers_can_finalize` | 439 | false | G1, G10, D7, D4 |
| 32 | `satp_theorem_3_2` | 462 | vacuous (Lean: D2) | D2; after repair G6, G7, G1, G2 |
| 33 | `satp_intermediary_before_round2` | 487 | false | G1, G2, G7, D4, D6 |
| 34 | `satp_end_user_recovers` | 506 | false | G6, G1, D4, D6 |
| 35 | `satp_reclaim_credits_intermediary` | 527 | sound | none |

Counts: sound 4, vacuous 14, false 15, unknown 2.

## Order of repair

1. D4, D3, D7, D8, D1-split: definition changes inside the approved form; no theorem type changes except the four
   Turbo theorems (D1).
2. D6 and D5: model changes that need a ruling (a second outcome field; a total strategy and a clock rule for
   `Prevents`).
3. D2 and G1-G13: theorem type changes, for operator approval.
4. Then re-run this audit before any proof. Theorems 21, 22, 26, 27, 29, 35 touch none of the game definitions and can be
   proved now.
