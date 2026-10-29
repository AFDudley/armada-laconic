# 2026-10-29 — Nitro theorems over unreachable states: ruling

## Facts (armada-lean `integrated`, 2026-10-29)

- `NetworkState` is `{chain, received, sent}` (`ArmadaLean/Nitro/Strategy.lean:34`). No Liveness theorem except
  `satp_lemma_3_1` limits it to reachable states. `Assumptions` (`Strategy.lean:162`) constrains only the adversary.
  `Views.Consistent` (`Strategy.lean:270`) constrains only agreement between views.
- `Reachable apps s` and `ReachableSigned` (`Adjudicator.lean:308, 331`) exist for the chain only.
- Three workers reported theorems false as stated: `directly_funded_pays` (aln-852), `finalizable_funded_pays`
  (aln-ceb), `ledger_partial_withdrawal` (aln-d1a). Each counterexample has coins in a virtual-kind channel. No
  reachable chain has that: `deposit` and `transfer` refuse a virtual channel (`Adjudicator.lean:100-152`).
  `[INFERENCE]` Under chain reachability each of the three is true as stated.
- Safety: 8 of the 63 theorems in `Safety.lean` take a reachability hypothesis.

## Operator rulings

1. The Nitro Liveness theorems assume a reachable network state: views reachable from a genesis network (the genesis
   chain, empty sent and received lists) by honest moves and adversary rounds. This constrains the chain and each
   participant's sent and received lists. It is a new definition. A kind hypothesis (`c.kind = .ledger`) is not the
   repair.
2. Audit `Safety.lean` the same way: list each theorem that is false or vacuous over unreachable states.

## Consequences

- The definition and the restated theorem types need operator approval of their Lean text before any proof pebble.
- The 6 Nitro proofs already merged hold for every state, so they still hold with the added hypothesis.
- `aln-852`, `aln-ceb` and `aln-d1a` wait for the restated types.
