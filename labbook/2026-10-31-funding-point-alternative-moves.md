# 2026-10-31 — Funding point: the guarantee is per participant; q0's double-sign is the canonical own-violation

## Finding

- `funding_point_universally_finalizable` (armada-lean `ArmadaLean/Nitro/Liveness.lean`, Nitro:631-640) concluded
  `UniversallyFinalizable apps Ts v c Ω`: every participant can force the agreed outcome `Ω`.
- Lean counterexample (Opus, `docs/counterexamples/funding_point_universally_finalizable.lean` on `worker/aln-cc8`,
  `3235226`): participant `q0` signs two setup states at its own turn 1 (below `c.n = 2`), one with `Ωg`, one with
  `Ωb`. The other side puts the `Ωb` state on-chain, and `q0` cannot stop it, because `q0` signed it. So `q0` cannot
  force `Ωg`.
- Nothing is wrong with the protocol or the model. ForceMove §3.1.5 (ForceMove:782-794) permits "alternative moves":
  "the player gives the next player the right to choose the move they want to progress from". Whatever finalizes is an
  outcome `q0` signed. The theorem over-claimed: it promised `Ωg` to a participant that handed the choice away.
- Only the equivocator loses. In the two-participant case, `p0` moves next at turn 2 and can put a turn-2 `Ωg` state
  above anything the other side posts [INFERENCE: Opus's argument, not checked in Lean]. `q0`, the last mover below
  `n`, has no later setup turn to answer from.

## Ruling (operator, 2026-10-31)

- The guarantee is per participant, conditioned on that participant's own sends; not a hypothesis on all
  participants. Forbidding alternative moves for honest participants in the model would contradict ForceMove §3.1.5.
- New definition and statement (`docs/lean/writing-lean.md` §11; state domain unchanged: every reachable network):

```lean
def MadeNoAlternativeSetupMoves {p : Participant} (net : NetworkState E p) (c : FixedPart) : Prop :=
  ∀ s1 ∈ net.sent, ∀ s2 ∈ net.sent, s1.state.channel = c → s2.state.channel = c →
    s1.state.turnNum = s2.state.turnNum → s1.state.turnNum < c.n → s1.state = s2.state

theorem funding_point_universally_finalizable … (hypotheses unchanged) :
    ∀ p ∈ c.participants, MadeNoAlternativeSetupMoves (v p) c →
      Finalizable apps (Ts.T p) c Ω (v p)
```

## The canonical own-violation example

`q0` signing two setup states at one turn is the project's canonical example of a participant violating its own
guarantee. It sits beside the other examples of `labbook/2026-10-30-app-classes-ic-ir.md` §5: unbound griefing, the
replayable ReceiptApp, `appsAll`/`appsU`, and Rock-Paper-Scissors without the default winner. Those are app-level; this
one is participant-level: a compliant app, a correct model, and a participant whose own signatures remove its claim.
The counterexample file is kept in armada-lean `docs/counterexamples/` as that example.

## Next

The statement change merges with `exophial integrate --branch` (it changes `scripts/sigs-baseline.tsv`). Pebble
`aln-cc8` was derived for the old conclusion and its history carries the counterexample, so it is closed and refiled
(operator order and reviewer-approved argv), as `aln-147` was.
