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

## Generic state: `Machine` and `System` (operator, 2026-10-29)

- Reachability is defined once, generically. A `Machine` is one reactive component (motif #13; Abrial 2010, Event-B
  `MACHINE`): states, initial states, one step relation. `Machine.Reachable` and one invariant principle
  (`Machine.invariant`) serve every machine (`docs/process.md` Stratum 2, and 4.4 on parallel inductions).
- Hierarchy (operator): a system contains plants, and a plant contains machines. A `System` is a composition of
  machines. Its reachable states are those of the composed machine. The network game is a system: the chain, one view
  per participant and the adversary step together. Network reachability is the composed machine's `Reachable`, not a
  separate predicate.
- Not "statechart": a statechart adds hierarchy to a flat machine (`docs/process.md` Stratum 4 draws flat first).
- Each type sent for approval gives its state domain (`engineering/operating-rules.md` §1). The four pool can-happen
  theorems (`PoolSafety.lean:639-794`) start from an arbitrary state; `docs/process.md:130` requires "built by real
  steps from genesis". Their restated form goes to the operator with the `Machine` text.
- `Machine` and `System` get glossary entries with this source when the Lean text is approved.
- Cross-chain activities get their own proofs later (operator, 2026-10-29). Within one chain the participants share
  the chain state, so the network machine has one chain (Stratum 1.2). Between chains nothing is shared or atomic:
  each chain is its own machine with its own clock and finality, and the system composes them by interleaving
  (`Machine.interleave` in draft `757d843`, deleted while unused, restored with its first cross-chain user). The
  cross-chain theorems carry their own finality and relay assumptions over the composed system's reachable states.
