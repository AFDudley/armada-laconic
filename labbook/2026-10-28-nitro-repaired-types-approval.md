# Approval of the repaired Nitro theorem types (2026-10-28)

The operator approved the repaired Nitro theorem types on 2026-10-27 and 2026-10-28. The approved commit is
`armada-lean` `nitro-proofs` `94436c3`, tag `nitro-proofs-approved-2026-10-28`. The baseline was tag
`nitro-theorems-approved-2026-10-27` (`69b9ea2`). The evidence was `nitro-proofs` `docs/nitro-repair-2026-10-27.md`
§4 and §6, and `labbook/2026-10-27-nitro-liveness-game-defects.md`.

## What the operator approved, and how

The coordinator showed the changes in groups. Each group below records what the operator saw before the approval.

| Group | Theorems | Shown |
|---|---|---|
| External destinations (D-E only) | `channel_address_inj`, `channel_address_ne_external`, `deposit_to_external_fails`, `held_falls_only_by_payout`, `paid_out_monotone`, `transfer_effects`, `deposit_bounds_value`, `payout_to_other_affects_only_owner`, `subchannel_open_value_preserving`, `subchannel_close_value_preserving`, `satp_outcome_change_keeps_total`, the D-E part of `satp_reclaim_credits_intermediary`; removed `deposit_to_participant_fails`, `channel_address_ne_participant` | the Lean text of the first two; the operator approved the class |
| D6: the finalized outcome is kept apart from the remainder | `statusOrigin_redistribute`, `conclude_can_happen` | a plain explanation |
| Channel kinds `ledger` and `virtual` | `deposit_can_happen`, `deposits_in_order_succeed`, `transfer_can_happen`, `reclaim_requires`, `reclaim_can_happen`, `reclaim_effects`, `reclaim_fails_without_entries`, the §6.3 part of `satp_reclaim_credits_intermediary` | a plain explanation and a table |
| Clock bound: `challengeDuration ≥ 2ε`, and at least ε left on a running challenge | the 12 theorems that carry `hdur : 2 * ε ≤ …` | a plain explanation |
| Turbo: transfers only for both players (D1); collection claims for external destinations only (G15) | `turbo_redistribution_unbeatable`, `turbo_value_upper_bound`, `turbo_value_achievable`, `turbo_value_is_value` | a plain explanation and a table |
| Newest state (G1, G2, G4, G5) | `latest_state_finalizable`, `consensus_enabled_outcomes`, `late_observer_can_checkpoint`, `satp_peers_can_finalize`, `satp_intermediary_before_round2`, `satp_theorem_3_2`, the definition `HoldsLatestLedger` | the approved and the repaired Lean text of each, with a table of every difference. The approval covers every row: G10, D2 (`hsetup` removed), G14b, G6/G7 and the `base` goal argument. |
| Covered by the groups above | `conclusion_proof_finalizable`, `concluded_universally_finalizable` | the coordinator compared each with the tag: every change is `E`, D6 or the clock bound |
| Approved without review of the Lean text | `directly_funded_pays` (G11), `finalizable_funded_pays` (G11, G14), `finalizable_or_two_enabled` (G12), `single_enabled_is_finalizable` (G13), `funding_point_universally_finalizable` (G9), `participant_can_prevent_cycles` (G8), `ledger_partial_withdrawal` (G16), `ledger_offload_funds_subchannel` (G16), `satp_lemma_3_1` (D2), `satp_end_user_recovers` (G6, new `hab`); removed `finalizable_is_enabled` | a one-line summary of each change |

The 55 Safety theorems that changed only by the parameter `E` are approved with the external-destination group.

Operator, on the last two groups: "I'm just going to approve them and we'll have to deal with it if they're wrong."

## Weakened goals and known risks

- `late_observer_can_checkpoint`: the goal changed from "the channel ends open" to "the channel ends not finalized".
  The adversary can register `ss` itself as a challenge (ForceMove:1511-1529).
- The theorems in the last group were approved without review of their Lean text. A proof attempt that finds a
  counterexample stops the work on that statement (`docs/process.md` 8.3).
- `satp_end_user_recovers` reads `V`'s stored outcome (`finalOutcome`) in `hfin`. Since §6.3, `reclaim` reads the
  finalized outcome (`finalizedWith`). For a virtual channel the two are equal after finalization, because a
  virtual channel accepts no deposit and no transfer `[INFERENCE]`. The proof must establish this.
