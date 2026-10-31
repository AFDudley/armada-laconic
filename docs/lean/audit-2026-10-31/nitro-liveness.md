# Audit: ArmadaLean.Nitro.Liveness

Slice `nitro-liveness`. Module: `ArmadaLean/Nitro/Liveness.lean`, 8,676 lines, 188 declarations.
Facts source: `.scratch/audit/facts.json` (after the fix below), cross-checked by direct read of
`ArmadaLean/Nitro/Liveness.lean` on the detached worktree `armada-lean/.scratch/audit-wt`
(`integrated` at `522e16d`). `draft/refined-types` consulted read-only at
`armada-lean/.scratch/types-wt`, reference only.

**Tooling defect found and fixed during this audit.** `.scratch/audit/facts.lean:32,53` called
`ConstantInfo.value?` with the default `allowOpaque := false`. Lean core
(`Declaration.lean:483-486`) returns `none` for `.thmInfo` unless `allowOpaque := true`, so every
`theorem`'s proof term was invisible to the extractor: `users`, `nitro_base_ops`,
`pool_base_ops` and `sorry` were wrong for all 937+281 theorems repo-wide, not only
`Liveness.lean` (`def`/`inductive` fields were unaffected). Main regenerated
`.scratch/audit/facts.json` with the fix during this audit. Every finding below in sections 3-6
uses the regenerated file, cross-checked by direct read where stated.

**Counts.** Layer map: 10 line-range sections, 9 confirmed cross-module public declarations.
Missing layer API: 1 duplicated case-split pattern across at least 6 declarations (905+ lines).
Refined-type candidates: 5 clusters (2 already covered by `draft/refined-types`, 1 by the
not-yet-merged spec-module `ReachableNetwork`, 1 new `Route` candidate, 1 new `Threat` candidate).
Layer violations: 5 confirmed, 2 field false positives explained. Size: 34 declarations over 60
lines. Dead code: 0 (the only `users: 0` facts outside spec-level theorems are the two
compiler-generated `OtherSteps.casesOn`/`recOn`, not authored declarations). Statement issues: 3
(the tooling defect, a stale module header, and a 19-instance unstated premise). Plan: 9 units in
layer order.

## 1. Layer map

`Liveness.lean` contains no base operations (those are `ArmadaLean/Nitro/Adjudicator.lean`,
`Channel.lean`) and sits mostly in the top three §4a layers. Sections below are by line range,
not by Lean `section`/`namespace` (the file has none except `namespace OtherSteps`,
Liveness.lean:3370-3405).

| Lines | Layer (§4a) | Content |
|---|---|---|
| 41-545 | per-call effect, feeding per-channel finalization | `directly_funded_step`/`_round` (41, 405): per-call slack-at-`d` bound. `UnbeatableIn.mono`, `RoundSeq.now_le/sent_eq/balances_external_mono/finalizedWith_mono` (260-404): frame lemmas over one adversary round. |
| 545-2910 | per-channel finalization, with embedded per-call case-split machinery (§2) | `round_keeps_status`(545) → `conclude_round_finalized`(750) → `conclusion_proof_finalizable`(784, **cites**) → `latest_round_keeps_status_nil`(814, **cites**)/`_nil_prelanding`(1002)/`_nil_D`(1295) → `land_challenge_other_statusOf`(1474)/`land_own_challenge_transitions_status`(1496) → `latest_round_keeps_status_two_ordered`(1611)/`_two`(1733) → `latest_round_keeps_status`(1787, **cites**) → `latest_wait_keeps_status`(2205)/`wait_resolves_challenge`(2397) → `latest_state_finalizable`(2439, **cites**)/`_two`(2520) → `late_round_keeps_status_nil`(2661)/`late_round_keeps_status`(2910). All carry the three-way "finalized / open / challenged-with-deadline" invariant for one consensus-app channel across one round of play. |
| 3310-3408 | per-channel finalization (two-chain agreement) | `namespace OtherSteps` (3370-3405); `NetworkState.applyAdv_received_subset`(3341), `applyHonest_finalizedWith`(3408). |
| 3408-4195 | per-channel finalization | `ChanAgree`/`ChanAgreeOrFin`(3469-3677, "`c` has one status on two chains"), `SigOK`/`VoucherOK`/`JointGenuine`(3705-3767, **cites** `VoucherOK`), `mirror_land`(3797), `finalizable_joint`(3859), `finalizable_unique`(3996, **cites**), `channel_finalizability_cases`(4015, **cites**), `finalizable_or_two_enabled`(4048, **cites**, sorry), `single_enabled_is_finalizable`(4066, **cites**, sorry), `funding_point_universally_finalizable`(4082, **cites**, sorry), `concluded_universally_finalizable`(4101, **cites**, sorry). Nitro §4.4. |
| 4195-5438 | per-call effect → paper theorems (consensus liveness) | `funded_entry_step`/`_round`/`_honest_step`/`_unbeatable`(4195-4630): the "funded for `d`" invariant across one call/round. `finalizable_funded_pays`(4631, **cites**). `consensus_round_safe_core`(4666, 536 lines, shared case split, §5) → `consensus_round_safe_nil`(5203)/`consensus_round_safe`(5299). |
| 5438-5756 | paper theorems (consensus game) | `ConsensusSafe0`(5438), `consensus_checkpoint_resolves_nil`(5470)/`_resolves`(5530), `consensus_enabled_outcomes`(5624, **cites**, Nitro-49). |
| 5756-6194, 6544-6717 | redistribution / Turbo (§4.3, §5.2) and constructions (§4.6) | `RoundSeq.transfers`(5756, **cites**), `Ahead.applyAll_sublist`(5782, **cites**), `turbo_replay_unbeatable`(5815), `turbo_redistribution_unbeatable`(5867, **cites**), `deposit_bounds_value`(5884, **cites**, sorry), `payout_to_other_affects_only_owner`(5895, **cites**, sorry), `consensus_channel_stays_open`(5908, **cites**), `participant_can_prevent_cycles`(6025, **cites**), `turbo_value_upper_bound`/`_achievable`/`_is_value`(6062-6094, **cites**, all sorry). Constructions: `ledger_offload_funds_subchannel`(6544, **cites**, sorry), `subchannel_open_value_preserving`(6565, **cites**, sorry), `subchannel_close_value_preserving`(6594, **cites**, sorry), `ledger_topup_reorder_safe`/`_deposit_funds`(6619-6640, **cites**), `balances_mono_while_open_step`(6669), `open_step`(6717). |
| 6096-6488 | per-call effect (Turbo `transfer` base op) | `Allocation.reduceOwed*`, `allocationOf_dest_applyAll`, `Edge_applyAll_iff`, `IsDAG_applyAll_iff`, `transfer_exhausted_fixed`, `Allocation.owed/dedup*`, `transfer_pointwise_absorbed`, `applyAll_*_absorbed`, `balances_eq_of_mutual_ahead`, `Ahead_right_congr`, `turbo_order_independent`(6443, **cites**), `turbo_value_exists_unique`(6488, **cites**). Exact-effect/frame lemmas for one `.transfer` call, the layer Turbo's paper theorems should sit on. |
| 6856-7572 | per-call effect/frame, general | `Op.targets`(6856)/`statusOf_apply_of_not_targets`(6868, doc comment: "a local copy of `NonVacuity.lean`'s `opOn`" — cross-module duplication, see §2). `channel_stays_open_round`(6898), `finalOutcome_isSome_apply`(7133), `directly_funded_step_open`(7156), `applyAdv_steps`/`RoundSeq.steps`(7375-7406), `finalOutcome_*`(7417-7460), `funded_step_addr`(7460), `Adjudicator.PaysAtLeast`/`PayPhase`(7476-7560). |
| 7572-7735 | paper theorems (payment guarantees) | `transfer_round_pays`(7572), `conclude_transfer_pays`(7606, **cites**), `ledger_partial_withdrawal`(7651, **cites**, sorry). |
| 7735-8676 | SATP layer, plus its own per-call/invariant machinery | `satp_outcome_change_keeps_total`(7735, **cites**) through `satp_theorem_3_2`(8069, **cites**, sorry); `ledgerCredit`(8314)/`ledger_credit_step`(8323)/`LedgerPairInv`(8374)/`.step`(8388); `RoundSeq.chainInv`(8436), `UnbeatableIn.and_chainInv`(8450); `satp_intermediary_before_round2`(8470, **cites**), `satp_end_user_recovers`(8585, **cites**, sorry), `satp_reclaim_credits_intermediary`(8614, **cites**). SATP:284-553. |

**Public API (declarations used from outside `Liveness.lean`, verified by qualified-name grep,
not the raw `users` count — see footnote).** 9 of 188 declarations, consumed almost entirely by
`ArmadaLean/Nitro/NonVacuity.lean`:
`NetworkState.applyAdv_sent`(271), `NetworkState.applyAdv_now_le`(285) — frame lemmas, used by
NonVacuity's own round reasoning; `SetupPhaseKeepsOutcome`(28), `SetupPhaseAccepted`(34),
`conclusion_proof_finalizable`(784), `finalizable_unique`(3996), `latest_state_finalizable`(2439),
`satp_outcome_change_keeps_total`(7735) — cited by NonVacuity.lean to build non-vacuity
witnesses for these paper theorems; `directly_funded_pays`(495) — used by both NonVacuity.lean
and Reachable.lean. No other ArmadaLean module calls into `Liveness.lean`. This is consistent
with Liveness sitting at the top of the stack: the 937-theorem Nitro namespace's `nitro_base_ops`
measurement in `type-doctrine.md` includes Liveness's own internal layering problem (below), but
nothing downstream of Liveness composes on top of it except NonVacuity's witnesses.

*Footnote on `users`.* The `users` field in facts.json counts ArmadaLean-wide references,
including same-file ones; most of the 188 declarations have `users > 0` purely from other
`Liveness.lean` declarations (confirmed: `directly_funded_step`, users caught by
`directly_funded_round`:480 and `funded_step_addr`:7473, both in-file). "Public API" above was
computed separately by grepping the qualified two-component suffix of each name across every
other `.lean` file and discarding generic-name collisions (`ChanAgree.symm`, `SigOK.swap`,
`Adjudicator.PayPhase.pays`, etc. — all zero real hits once qualified; short-name grep falsely
matched unrelated `.symm`/`.swap`/`.step` declarations in `PoolOwned.lean`/`PoolSafety.lean`).

## 2. Missing layer API

**One pattern, proved independently at least six times.** `Liveness.lean` repeats the same
per-`Op`-constructor case split — "does `op` target channel `c`; if not, `c`'s
(`statusOf`, `finalizedWith`) record is unchanged; if so, use the call's own named success lemma
(`Adjudicator.conclude_some`, `.challenge_some`, `.checkpoint_some`, `.transfer_some`,
`.reclaim_some`, `.deposit_some`)" — as an inline `have step : … := by cases op with …` block,
instead of calling one shared per-call frame/effect lemma:

- `latest_round_keeps_status_nil`:848-1000 (Liveness.lean:875-990, the `other` case: full
  `cases op with | deposit | tick | transfer | reclaim | challenge | checkpoint | conclude`).
- `latest_round_keeps_status_nil_prelanding`:1014-1293 (doc comment, 1006-1009, names the
  duplication itself: "the same turn-bound argument … as `latest_round_keeps_status_nil`'s own
  `tick`/`other` cases, extracted standalone because those two cases never consult `pending`'s
  contents").
- `latest_round_keeps_status_nil_D`:1304-1473 (sibling of the above, same case split, a third
  conclusion shape).
- `late_round_keeps_status_nil`:2674-2908 (doc comment, 2666-2667, again names the duplication:
  "the turn-bound argument of `latest_round_keeps_status_nil` applies unchanged to `challenge`
  and `checkpoint`").
- `latest_round_keeps_status`:1787-2203 and `late_round_keeps_status`:2910-3266 each additionally
  redo the `land` case of this same split on top of their own `_nil` sub-lemma.
- `directly_funded_step`:63-258 and `directly_funded_step_open`:7168-7373 are the same pattern a
  second time, for a different invariant (funding slack, not status): the doc comment at
  7156-7165 states `directly_funded_step_open` is "`directly_funded_step` generalized", built by
  walking the same case tree again with adjusted sub-cases, not by calling the original.

A per-call frame half of the needed lemma already exists, later in the same file, but is not
reused by any of the above: `Op.targets`:6859 and `statusOf_apply_of_not_targets`:6868-6896. Its
own doc comment (6856-6858) says it is "a local copy of `NonVacuity.lean`'s `opOn`" — a third,
cross-module copy of the same classification.

**Proposed fix.** Promote the `have step : …` block common to the five `*_keeps_status*`
declarations to one named lemma, and route `directly_funded_step_open` through
`directly_funded_step` instead of re-deriving it. The frame half is already `Op.targets` /
`statusOf_apply_of_not_targets`; the missing half is the "targets" case, which each site already
calls the SAME five named success lemmas (`Adjudicator.challenge_some` etc.) for — only the
packaging as a per-call step differs:

```lean
/-- One on-chain call's effect on channel `c`'s latest-record invariant: if `op`
    does not target `c` (`Op.targets`), the record is unchanged
    (`statusOf_apply_of_not_targets`); if it does, the call's own success lemma
    gives the exact new record, bounded by `hbounds.unforgeable` when the
    signer is not `p` itself. Replaces the `have step := by cases op with …`
    block duplicated in `latest_round_keeps_status_nil`:850-991,
    `_nil_prelanding`:1026-1292, `_nil_D`, and `late_round_keeps_status_nil`. -/
theorem Op.latest_record_step {apps : Apps E} {p : Participant} {T : Threat E p}
    (hbounds : Bounds apps p T) {c : FixedPart} {ss : SignedState E}
    (hc : apps c.appDef = consensusRules) (hp : p ∈ c.participants)
    {caller : Participant} {op : Op E} {s s' : Adjudicator E}
    (hfeas : T.feasible ⟨caller, op⟩ /- or the round's own feasibility witness -/)
    (h : s.apply apps caller op = some s')
    (hU : ∀ ss' ∈ op.proofStates, p ∈ ss'.signers → ∃ x, x ∈ sent ∧ x.state = ss'.state) :
    (s.finalizedWith c.address = some ss.state.outcome ∨
      ∃ T0 ch, s.statusOf c = .challenged T0 ch ∧ ss.state.turnNum ≤ T0 ∧ ch.state = ss.state) →
    (s'.finalizedWith c.address = some ss.state.outcome ∨
      ∃ T0 ch, s'.statusOf c = .challenged T0 ch ∧ ss.state.turnNum ≤ T0 ∧ ch.state = ss.state)
```

`[INFERENCE]`: the exact binder list needs the author's pass, because the four sites' conclusions
differ in shape (two-way vs. three-way disjunction, with/without an `open` branch); a single
lemma may need to be the strongest (three-way) shape, with the weaker sites calling it and
dropping the unused disjunct, or the Op-level step may need to be generic over the "record" being
carried (a small typeclass/structure argument) so the five call sites instantiate it once per
invariant instead of once per invariant *and* per round-shape. Either design removes the
case-split duplication; which one is for the implementing worker to choose, not this audit.

## 3. Refined-type candidates

All instance counts, agreement, and premise shape below are read from `.scratch/hyp-analysis.md`
(the 89 in-scope Nitro theorems correspond almost exactly to Liveness's 45 `cites: true`
declarations) and cross-checked by direct read.

| Predicate | Users (Liveness) | Agree? | Status |
|---|---|---|---|
| `(Nitro.participant apps T).Reachable net` | 19: `directly_funded_pays`, `finalizable_funded_pays`, `conclusion_proof_finalizable`, `latest_state_finalizable`, `late_observer_can_checkpoint`, `finalizable_or_two_enabled`, `single_enabled_is_finalizable`, `consensus_enabled_outcomes`, `turbo_redistribution_unbeatable`, `deposit_bounds_value`, `payout_to_other_affects_only_owner`, `participant_can_prevent_cycles`, `turbo_value_upper_bound/_achievable/_is_value`, `satp_peers_can_finalize`, `satp_theorem_3_2`, `satp_intermediary_before_round2`, `satp_end_user_recovers` | Yes, all 19 identical | **New, not yet in a type.** `typed-library.md` §4 already names the destination, `Reachable.lean`'s `ReachableNetwork apps Ts`, from the not-yet-merged `draft/spec-module-2` (order-of-work item 1). Depends on that merge; do not define a second type here. |
| `2 * T.ε ≤ c.challengeDuration` | 9: `conclusion_proof_finalizable`, `latest_state_finalizable`, `late_observer_can_checkpoint`, `finalizable_or_two_enabled`, `single_enabled_is_finalizable`, `consensus_enabled_outcomes`, `satp_theorem_3_2`, `satp_intermediary_before_round2`, `satp_end_user_recovers` | Yes, all 9 identical (the sibling `Safety.lean` instance `0 < c.challengeDuration` is a different, weaker claim for a different theorem, not a disagreement) | **Already covered.** `draft/refined-types`' `ArmadaLean/Types/Channel.lean:27` defines `TimedChannel ε := {c : FixedPart // c.IsTimed ε}` with exactly this predicate. Convert these 9 call sites once that branch merges; do not repeat the type definition. |
| `apps c.appDef = consensusRules` | 8: `conclusion_proof_finalizable`, `latest_state_finalizable`, `late_observer_can_checkpoint`, `finalizable_or_two_enabled`, `single_enabled_is_finalizable`, `funding_point_universally_finalizable`, `concluded_universally_finalizable`, `consensus_enabled_outcomes` | Yes, all 8 identical | **Already covered as a Prop.** `draft/refined-types`' `ArmadaLean/Nitro/Constructions.lean:45` defines `FixedPart.RunsConsensus apps c := apps c.appDef = consensusRules`, named but not yet wrapping `c` in a type. Once merged, these 8 sites should at minimum spell the hypothesis as `c.RunsConsensus apps` (one name, `repeatedPremise`-safe); whether it also becomes a bundled type with `LedgerChannel` is an operator decision outside this audit's scope. |
| `Route.peers`/`Route.WellFormed`/`Route.ledgers`/`Route.V.appDef`/`Route.peers.getLast`/`Route.peers.head` (4-6 instances each) | All 5 `satp_*` end-to-end theorems: `satp_lemma_3_1`, `satp_peers_can_finalize`, `satp_theorem_3_2`, `satp_intermediary_before_round2`, `satp_end_user_recovers` take `r.WellFormed apps` and then separately re-derive or re-assume facts about `r.peers`/`r.ledgers`/`r.peers.getLast`/`r.peers.head` that `WellFormed` already implies | `[INFERENCE]`, not checked instance-by-instance | **New candidate**, not addressed by `draft/refined-types`. Propose a `Route.WellFormed`-indexed construction (`typed-library.md`'s `Construction.lean` pattern, like `LedgerWithSub.In n`): a structure carrying `r` and the proof of `r.WellFormed apps`, with the five `satp_*` theorems taking it instead of `r : Route E` plus a repeated `hr : r.WellFormed apps` plus ad hoc re-derivations of the peer/ledger facts. |
| `Threat.TransfersOnly`/`Threat.CapableTransfers` (4 instances each) | `turbo_redistribution_unbeatable`, `turbo_value_upper_bound`, `turbo_value_achievable`, `turbo_value_is_value` (plus `turbo_order_independent`/`turbo_replay_unbeatable` for `List.OnlyTransfers`) | Yes | **Partially covered.** `draft/refined-types`' `TurboOutcomeNetwork` (`Outcome.lean:42`) refines the *Adjudicator state* side (`Adjudicator E` with `IsTurboOutcomeNetwork`). It does not cover the *Threat* side (`Threat.TransfersOnly`/`Threat.CapableTransfers`, repeated as separate raw hypotheses on 4-5 of the same theorems). A `TurboThreat` companion type, or folding these two predicates into one `Threat`-indexed structure alongside `TurboOutcomeNetwork`, is a new candidate for the same operator ruling pass. |

## 4. Layer violations

Spec-level (`cites: true`) theorems whose proof names a base operation (`Op.apply`,
`NetworkState.applyAdv`, `Adjudicator.transfer`, `.challenge`, `.checkpoint`, `.conclude`,
`.reclaim`), verified by reading the proof body, not by the field count alone (two field hits
below are shown to be false positives from the conclusion's own type, not the proof's reasoning —
see the note after the table).

| Theorem | Line | Base op named, confirmed by read | Layer API that would replace it |
|---|---|---|---|
| `Threat.RoundSeq.transfers` | 5760 | `simp only [NetworkState.applyAdv]` (5779); `n.chain.apply apps caller (.transfer a b)` as a direct `cases` discriminant (5780) | A per-call exact-effect lemma for `.transfer` under `NetworkState.applyAdv`, from the §2/6096-6488 Turbo per-call layer (`Adjudicator.apply_transfer`, already named and used two lines later in the sibling `Adjudicator.Ahead.applyAll_sublist`, so it exists — this site should use it instead of unfolding `NetworkState.applyAdv` directly). |
| `Adjudicator.Ahead.applyAll_sublist` | 5789 | `cases ht : sg.transfer a b with` (5803), a direct dot-call on the base op, not through `Adjudicator.apply_transfer`'s own lemma | The frame lemma this theorem is building IS the layer API being skipped one level down; it should be stated in terms of `Adjudicator.apply_transfer`'s exact-effect form rather than re-pattern-matching the base call. |
| `latest_round_keeps_status_nil` | 827 | `unfold Adjudicator.challenge` would appear in the sibling `latest_round_keeps_status`; this one directly unfolds `simp only [NetworkState.applyAdv]` five times (905, 914, and via `.applyAdv_sent`/`.applyAdv_now_le` call sites at 860-873) instead of composing the already-existing `NetworkState.applyAdv_sent`/`_now_le` frame API | The proposed `Op.latest_record_step` lemma of §2; once the per-call step is a named lemma, the round-level induction no longer needs to unfold `NetworkState.applyAdv` itself. |
| `latest_round_keeps_status` | 1787 | `rw [Adjudicator.apply_challenge]; unfold Adjudicator.challenge` (1862, 1878, 1908) — a direct unfold of the base op's definition, not its success lemma `Adjudicator.challenge_some` | Same `Op.latest_record_step` (§2); the `land` case here should call the same per-call step the `_nil` siblings already (separately) reconstruct. |
| `consensus_channel_stays_open` | 5908 | `simp only [NetworkState.applyAdv]` (5938, 5950); `net1.chain.transfer`/`.reclaim` read via `Adjudicator.transfer_some`/`reclaim_some` (correct, layer-API style) for the payout cases, but the round-clock and sent-list bookkeeping still unfolds `NetworkState.applyAdv` directly | Same `NetworkState.applyAdv_sent`/`_now_le` frame API already defined in this file (271, 285) but not used here. |

**Field false positives, checked by full read.** `satp_intermediary_before_round2`:8508-8583 and
`satp_reclaim_credits_intermediary`:8620-8674 both show `nitro_base_ops: 1`, but their proofs
compose cleanly through layer lemmas (`latest_state_finalizable_two`, `LedgerPairInv.step`,
`UnbeatableIn.and_chainInv`/`.imp`, `reclaim_can_happen`); neither unfolds or dot-calls a base op.
Their CONCLUSIONS state an exact base-op effect (`s.reclaim Lm r.V = some s1`,
Liveness.lean:8634) — the correct "general existence lemma states the exact effect" style
(`writing-lean.md` §7) — and the elaborated proof term's implicit motive, built from that
conclusion's own type, is what the extractor's `getUsedConstants` sees. `type-doctrine.md`'s own
caveat applies here: "a proof that uses only lemmas about an operation does not name it" — the
converse also holds, a conclusion that names an operation's exact effect is not a proof that
descends to it.

## 5. Size

34 declarations over 60 lines (full list: `consensus_round_safe_core`:4666(536),
`latest_round_keeps_status`:1787(417), `late_round_keeps_status`:2910(357),
`latest_round_keeps_status_nil_prelanding`:1002(291), `late_round_keeps_status_nil`:2661(248),
`channel_stays_open_round`:6898(234), `directly_funded_step_open`:7156(218),
`directly_funded_step`:41(218), `funded_entry_step`:4195(206), `round_keeps_status`:545(204),
`latest_wait_keeps_status`:2205(191), `latest_round_keeps_status_nil`:814(187),
`latest_round_keeps_status_nil_D`:1295(179), `latest_state_finalizable_two`:2520(141),
`consensus_round_safe`:5299(139), `open_step`:6717(138), `finalizable_joint`:3859(136),
`consensus_enabled_outcomes`:5624(131), `ledger_claim_step`:8120(122),
`latest_round_keeps_status_two_ordered`:1611(121), `consensus_channel_stays_open`:5908(116),
`satp_intermediary_before_round2`:8470(114), `land_own_challenge_transitions_status`:1496(114),
`consensus_round_safe_nil`:5203(95), `funded_entry_round`:4449(94),
`consensus_checkpoint_resolves`:5530(93), `directly_funded_round`:405(89),
`satpOrigin_reachableSigned`:7914(88), `ledger_partial_withdrawal`:7651(83),
`latest_state_finalizable`:2439(79), `ChanAgree.apply`:3591(77),
`finalOutcome_of_fresh_finalizedWith`:8243(70), `satp_reclaim_credits_intermediary`:8614(61),
`mirror_land`:3797(61)).

Causes, by evidence:

- **Duplication (named twin), confirmed.** `latest_round_keeps_status_nil`(187),
  `_nil_prelanding`(291), `_nil_D`(179), `late_round_keeps_status_nil`(248),
  `latest_round_keeps_status`(417, embeds `_nil`'s split again for its `land` case),
  `late_round_keeps_status`(357, same): 905+ lines across six declarations are the same
  per-`Op`-constructor case split (§2), each site's doc comment naming the others as its source.
  `directly_funded_step`(218)/`directly_funded_step_open`(218): confirmed twin, §2.
- **Case split per constructor, not yet a duplicate but the doctrine's target pattern.**
  `consensus_round_safe_core`(536, doc comment 4666-4673 calls itself "the shared core of the
  block that both [callers] carried in full" — already a first deduplication pass, but still one
  536-line inline `cases m with` rather than one frame lemma per `Op` constructor, same gap as
  §2's proposed `Op.latest_record_step`), `channel_stays_open_round`(234), `funded_entry_step`
  (206), `round_keeps_status`(204), `open_step`(138), `ledger_claim_step`(122),
  `land_own_challenge_transitions_status`(114) — each has `nitro_base_ops ≥ 1` (confirmed field,
  consistent with a direct per-constructor split rather than composed per-call lemmas).
- **Inline re-derivation of a fact.** `latest_state_finalizable_two`(141) and
  `latest_round_keeps_status_two_ordered`(121)/`_two`(53) re-derive, per channel, the same
  "which channel's call lands first" ordering argument that `Threat.RoundSeq.land`'s own
  unconstrained-position property (cited at Liveness.lean:4490-4491 inside
  `satp_intermediary_before_round2`'s doc comment) already states once; `[INFERENCE]`, not traced
  line-by-line.
- **Other (genuine case-exhaustive proof, not obviously reducible).** `finalizable_joint`(136),
  `consensus_enabled_outcomes`(131, no `nitro_base_ops` hit — a four-way joint-strategy case
  split intrinsic to the consensus game, not an Op-constructor split), `mirror_land`(61),
  `ChanAgree.apply`(77), `satpOrigin_reachableSigned`(88) — `[INFERENCE]`, not individually
  read to confirm; flagged for the implementing worker to re-check before assuming they reduce.

## 6. Dead code

None. The mechanical scan (`cites: false` and `users: 0` in the regenerated facts.json) returns
exactly `ArmadaLean.Nitro.OtherSteps.casesOn` and `.recOn` (Liveness.lean:3361) — the
compiler-generated recursor/eliminator Lean emits automatically for every `inductive`
(`OtherSteps`), not an authored declaration, present regardless of use. Every other declaration
in the module either has a cross-reference inside `Liveness.lean` itself (confirmed for a sample
including `directly_funded_step`, whose `users: 0` in the *first*, buggy facts.json run was the
tooling defect, not dead code — see the top-of-report note) or is itself a `cites: true` spec
theorem (exempted by this report's instructions, §6).

## 7. Statement issues seen

1. **Tooling defect (not a statement defect), evidence given at the top of this report.**
   `.scratch/audit/facts.lean`'s `ci.value?` call defaulted `allowOpaque` to `false`, making
   `users`/`nitro_base_ops`/`pool_base_ops`/`sorry` unreliable for every `theorem` in the first
   facts.json. Fixed and regenerated by Main during this audit (confirmed:
   `directly_funded_step` went from `users: 0` to `users: 2`; `payout_to_other_affects_only_owner`
   now correctly shows `sorry: true`). This report uses only the regenerated file.
2. **Stale module header.** `Liveness.lean:1-7` still reads "Every proof is `sorry`, except the
   theorems that read no game definition." This is false on the current state: 31 of the 45
   `cites: true` theorems in this module are fully proved (§sorry count above, 14 of 45). A
   reader trusting the header would misjudge the module's completion state by more than 2×.
3. **Premise with no stated source, 19 instances.** `.scratch/hyp-analysis.md`'s
   `NetworkState.Reachable` cluster marks every one of its 19 Liveness instances
   (`directly_funded_pays` through `satp_end_user_recovers`, full list in §3) as Class
   **unstated [INFERENCE]**: no doc-comment sentence names the `hr`/`_hr`/`hnr` hypothesis
   together with a source. The reachability precondition is real (an unreachable network can
   violate several of these theorems, by the same class of counterexample `typed-library.md`:§1
   cites for the `hr` premise elsewhere in the repo), but no Liveness doc comment says so; it is
   present only because the type doctrine's §1 finding (same pattern, repo-wide) applies here
   too. Not a model defect — the premise is correct — but it is exactly the "same fact decided
   separately, nineteen times, with no stated reason" pattern decision 1 targets.

## 8. Plan

Layer order (base → paper theorems), each unit sized for one worker.

1. **Merge `draft/spec-module-2`'s `ReachableNetwork apps Ts`.** Prerequisite for unit 4. Not in
   this slice's scope (cross-cutting); name as a dependency, do not duplicate.
2. **Merge `draft/refined-types`.** Brings `TimedChannel`, `LedgerChannel`, `TurboOutcomeNetwork`,
   `FixedPart.RunsConsensus` into trunk. Prerequisite for units 5 and 6. Not in this slice's
   scope; name as a dependency, do not duplicate.
3. **Write `Op.latest_record_step`** (§2), the per-call frame/effect lemma for one channel's
   latest-record invariant under one `Op`. Input: the five duplicated `have step := …` blocks at
   Liveness.lean:850-991, 1026-1292 (`_nil_prelanding`), `_nil_D`'s equivalent, and
   2674-2908 (`late_round_keeps_status_nil`); reuse `Op.targets`/`statusOf_apply_of_not_targets`
   (6859-6896) for the non-targeting half. Acceptance: one new lemma, proved once, that each of
   those four theorems' `have step` block becomes a direct call to (no `cases op with`
   remaining in the four call sites). No change to any theorem's statement. Depends on nothing
   else in this plan.
4. **Rewrite `latest_round_keeps_status_nil`/`_nil_prelanding`/`_nil_D`/`late_round_keeps_status_nil`
   to use unit 3's lemma**, and fold `latest_round_keeps_status`/`late_round_keeps_status`'s
   embedded copy of the same split into a call on the `_nil` sibling plus the `land` case alone.
   Acceptance: `grep -c "cases op with"` in these six declarations drops to the minimum needed
   for the `land` case only; `wc -l` combined drops from ~1,575 to a worker-estimated ≤900;
   layer-violation findings 3 and 4 in §4 (direct `NetworkState.applyAdv`/`Adjudicator.challenge`
   unfolds) are gone. Depends on unit 3.
5. **Merge `directly_funded_step_open` into `directly_funded_step`**, or state the general form
   once and derive the external-destination case as a one-line corollary (per the doc comment at
   7156-7165, the two proofs differ only in how the "not this channel" sub-cases are closed).
   Acceptance: one of the two 218-line proofs becomes a corollary of the other, under 20 lines.
   Independent of units 3-4.
6. **Convert the `NetworkState.Reachable` cluster (19 sites, §3) to `ReachableNetwork`**, once
   unit 1 lands. Headers-only conversion per `typed-library.md` §5 step 3. Depends on unit 1.
7. **Convert the `2 * T.ε ≤ c.challengeDuration` (9 sites) and `apps c.appDef = consensusRules`
   (8 sites) clusters to `TimedChannel`/`FixedPart.RunsConsensus`**, once unit 2 lands.
   Headers-only. Depends on unit 2.
8. **Design and rule on the `Route.WellFormed`-indexed construction** (§3, the SATP `Route`
   candidate, no agreed shape yet) and the `TurboThreat` companion to `TurboOutcomeNetwork`
   (§3). These are new operator decisions, not yet ruled; this unit is "produce the candidate
   type's Lean statement for operator review", not "convert the call sites" (that is unit 9).
   Independent of units 1-7.
9. **Convert the 5 `satp_*` Route sites and the 4-5 Turbo `Threat` sites** to whatever unit 8
   rules. Depends on unit 8.

Units 3-4 are the highest-value, most load-bearing: they remove the single largest duplicated
block in the module (905+ lines across six declarations) and three of the five confirmed layer
violations in §4. Units 6-7 and 9 are mechanical once their type dependencies land. Fixing
`consensus_round_safe_core`'s 536-line inline split (§5) into per-constructor frame lemmas is the
same shape of work as unit 3 but for a different invariant (`ConsensusSafe` rather than the
three-way status record); not separately planned here, flagged for a follow-up slice since it is
not cited by any other Liveness declaration outside its own two callers and so is lower-leverage
than units 3-4.
