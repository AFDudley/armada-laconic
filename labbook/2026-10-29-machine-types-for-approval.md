# 2026-10-29 — `Machine`, the network machine and the restated types: for operator approval

Source: armada-lean branch `draft/reachable-network`, tip `9d4216a`, worktree `armada-lean/.scratch/reach-draft`.
`integrated` (`ed82fb3`) is not changed. `lake build` passes, and `scripts/check.sh` passes every check. Rulings:
`labbook/2026-10-29-nitro-reachability-ruling.md`.

## 1. Definitions

`ArmadaLean/Machine.lean` (imports `Mathlib.Logic.Relation` only):

```lean
structure Machine (S : Type) where
  init : S → Prop
  step : S → S → Prop

def Machine.Reachable {S : Type} (M : Machine S) (s : S) : Prop :=
  ∃ s₀, M.init s₀ ∧ Relation.ReflTransGen M.step s₀ s

theorem Machine.invariant {S : Type} (M : Machine S) {P : S → Prop}
    (h0 : ∀ s, M.init s → P s) (hs : ∀ s s', P s → M.step s s' → P s') :
    ∀ s, M.Reachable s → P s            -- proved
```

The machines (`ArmadaLean/Nitro/Reachable.lean`, `ArmadaLean/PoolMachine.lean`):

| Machine | State | `init` | `step` | Line |
|---|---|---|---|---|
| `Pool.machine` | `PoolState` | `∃ r, s = genesis r` | `Step` | `PoolMachine.lean:16` |
| `Nitro.chain apps` | `Adjudicator E` | `s = Adjudicator.genesis` | `Step apps` | `Reachable.lean:155` |
| `Nitro.participant apps T` | `NetworkState E p` | genesis chain, empty lists | an honest move `p` can make, or a move `T` allows | `Reachable.lean:146` |
| `Nitro.network apps Ts` | `Network E` | genesis chain, all lists empty | see below | `Reachable.lean:312` |

`Network` has ONE chain and per-participant lists (`Reachable.lean:199`). Each participant's view is a projection,
`Network.view n p = ⟨n.chain, n.received p, n.sent p⟩`, so chains cannot diverge (Stratum 1.2).

```lean
def Nitro.network (apps : Apps E) (Ts : Threats E apps) : Machine (Network E) where
  init := fun n => n = ⟨Adjudicator.genesis, fun _ => [], fun _ => []⟩
  step := fun n n' => ∃ p : Participant,
    (∃ m : HonestMove E, m.Feasible (n.view p) ∧
      (∀ op, m = .chain op → ∀ ss ∈ op.proofStates, n.Sendable ss) ∧
      n' = Network.applyHonest apps n p m) ∨
    (∃ m : AdvMove E, (Ts.T p).feasible (n.view p) m ∧
      (∀ ss, m = .deliver ss → n.Sendable ss) ∧
      (∀ caller op, m = .chain caller op → ∀ ss ∈ op.proofStates, n.Sendable ss) ∧
      n' = Network.applyAdv apps n p m)
```

- `Network.Sendable n ss`: every signer named in `ss` (and in its voucher) sent that state (`Reachable.lean:227`).
  This is unforgeability, SATP:168, in the step relation (ruling on conjunct 3).
- An on-chain call is public: its proof states reach every other participant's `received` (`Reachable.lean:240-268`).
  This extends the D9 repair (`Strategy.lean:88-96`) from one observer to the whole network.

## 2. Proved properties (no `sorry`)

| Lemma | Says |
|---|---|
| `Nitro.chain_reachable_iff`, `Pool.machine_reachable_iff` | the machines' `Reachable` equals the existing `Reachable` predicates |
| `Nitro.participant_chain_reachable`, `Nitro.network_chain_reachable` | a reachable view or network has a reachable chain |
| `Nitro.network_chain_reachableSigned` | the chain of a reachable network is `ReachableSigned` for the states its participants sent |
| `Reachable.balances_virtual_zero` | a reachable chain holds no coins in a virtual-kind channel |
| `Network.views_chain_eq` | conjunct 1 of `Views.Consistent`, by construction |
| `Nitro.network_views_signers` | conjunct 3 of `Views.Consistent` for every reachable network |
| `Nitro.network_views_consistent_iff` | for a reachable network, `Views.Consistent` ⇔ conjunct 2 (synchrony) |

## 3. Restated types, with state domains

| Theorem | Added hypothesis | State domain |
|---|---|---|
| `directly_funded_pays` | `(hr : (Nitro.participant apps T).Reachable net)` | reachable view |
| `finalizable_funded_pays` | `(hr : (Nitro.participant apps T).Reachable net)` | reachable view |
| `ledger_partial_withdrawal` | `(n : Network E) (hr : (Nitro.network apps Ts).Reachable n) (hvn : v = n.views)` | reachable network; `hv` = synchrony |
| `shield_can_happen`, `transact_can_happen`, `unshield_can_happen`, `owned_transact_can_happen` | `(hr : Pool.machine.Reachable s)` | reachable pool state (`process.md:130`) |

`[INFERENCE]` The three Nitro theorems are true under these hypotheses: `Reachable.balances_virtual_zero` forces a
virtual channel's balance to 0, so the owed amount is 0 and the goal holds with no move. The existing pool proofs close
the four can-happen theorems unchanged. In the draft they are new theorems (`*_reachable`, `PoolMachine.lean:41-137`)
beside the originals. On approval they replace the originals.

The other Liveness theorems gain one line each (the draft lists them; their bodies are `sorry` or proved for every
state):
- single view, `(hr : (Nitro.participant apps T).Reachable net)`: `conclusion_proof_finalizable`,
  `latest_state_finalizable`, `late_observer_can_checkpoint`, `finalizable_or_two_enabled`,
  `single_enabled_is_finalizable`, `consensus_enabled_outcomes`, `turbo_redistribution_unbeatable`,
  `deposit_bounds_value`, `payout_to_other_affects_only_owner`, `participant_can_prevent_cycles`,
  `turbo_value_upper_bound`, `turbo_value_achievable`, `turbo_value_is_value`, `satp_peers_can_finalize`,
  `satp_theorem_3_2`, `satp_intermediary_before_round2`, `satp_end_user_recovers`;
- network, `(n : Network E) (hr : (Nitro.network apps Ts).Reachable n) (hvn : v = n.views)`: `finalizable_unique`,
  `channel_finalizability_cases`, `funding_point_universally_finalizable`, `concluded_universally_finalizable`,
  `ledger_offload_funds_subchannel`; two views (`v`, `v'`) get one each: `subchannel_open_value_preserving`,
  `subchannel_close_value_preserving`;
- no change: `satp_lemma_3_1` (has `ReachableSigned`); and, holding for every state, `turbo_order_independent`,
  `turbo_value_exists_unique`, `ledger_topup_reorder_safe`, `ledger_topup_deposit_funds`,
  `satp_outcome_change_keeps_total`, `satp_reclaim_credits_intermediary`.

Six Nitro proofs are merged. Each holds for every state, so each still holds with its added hypothesis.

## 4. Open finding (no theorem depends on it now)

A view of a reachable network is not always reachable as a single `Nitro.participant` machine. In the network, `p`'s
adversary can submit an on-chain call with another participant `q` as the caller (Nitro:312-313: anyone can trigger
an operation). `q`'s own threat excludes a call with caller `q` (`Bounds.notHonest`). So no single-view theorem may
be applied to a view of a network until the network decides who may call as a participant. The theorems above never
combine the two forms.

## 5. Decisions asked

1. Approve the `Machine` definitions and the four machines in §1.
2. Approve the restated types and their state domains in §3, with the four pool can-happen theorems replacing the
   originals.
3. §4: who may submit an on-chain call as a participant `q` in the network. The coordinator recommends: only `q`'s
   own honest moves (a call is a signed transaction from `q`'s key), so an adversary call has a caller outside the
   participants of the channel it touches. Defer, or rule now.
