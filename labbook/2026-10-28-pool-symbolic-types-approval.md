# Approval of the pool-symbolic theorem types (2026-10-28)

The operator approved every changed and every new theorem type of the pool model. The model is `armada-lean` branch
`pool-symbolic` at `3128085`. The changes come from the pool layout audit (`labbook/2026-10-26-pool-layout-audit.md`,
findings P-1 to P-11) and are listed in `pool-symbolic` `docs/pool-symbolic-changes.md`.

## The 17 changed types (approved 2026-10-27)

On 2026-10-26 at 23:28 the coordinator asked the operator to approve 17 changed types, with option A: "approve all
17, and model the shield fee inside the shield step". The coordinator showed the fee theorems as a typical change
and `shield_can_happen` as the one change of meaning. The commit `88571c0` (2026-10-27 01:26, "Model the shield fee in
the shield step (operator ruling on P-9)") applied option A. Until this entry, that commit message was the only
record of the approval.

- Fees: `fee_bounded`, `fee_conservation`, `fee_monotone`, `fee_zero_below_threshold`.
- Can happen: `shield_can_happen`, `transact_can_happen`, `unshield_can_happen`, `owned_transact_can_happen`,
  `owned_spend_can_happen`.
- Exact effects: `shield_creates_exactly`, `transact_unshield_exact`, `transactOwned_records_all`.
- Spend witnesses and live notes: `leaf_has_merkle_witness`, `owned_spend_witness_exists`, `live_leaf_matches`,
  `live_values_spendable`, `leaf_stable`.

## The 5 new types (approved 2026-10-28)

The commit `479a4f0` added the channel-design theorems 1 to 3. The coordinator showed the Lean text of each, with
its plain meaning and its contract source.

| Theorem | Plain meaning | Source |
|---|---|---|
| `spend_not_blocked` | a step that spends none of a pending spend's notes leaves that spend enabled | `TransactModule.sol:373-399,538` |
| `transact_not_blocked` | the same, for a pending `transact` | as above |
| `transactOwned_not_blocked` | the same, for a pending `transactOwned` | as above |
| `spend_bound_to_adapt` | a spend bound to `(adaptContract, adaptParams)` runs only by that contract, with those bound parameters | `TransactModule.sol:530-532`, `PrivacyPoolStorage.sol:46-57` |
| `owned_merge_can_happen` | an owner can merge up to 8 of its live notes into one note of their exact sum, with custody unchanged | `TransactModule.sol:488-519` |

`owned_merge_can_happen` holds for the deployed pool only if owned circuits with more than one input exist. The
deployed owned circuits have one input (`main_{1x1,1x2,1x3}_owned.circom`).

## Result

All 15 `sorry` theorems of `pool-symbolic` have approved types. The pool proofs are next, before the Nitro proofs
(operator, 2026-10-28).
