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

## Own-violations are IR violations (operator, 2026-10-31)

A participant that signs what the protocol tells it not to sign violates its own individual rationality
(`labbook/2026-10-30-app-classes-ic-ir.md` §2). The papers assume protocol-following participants, so they never state
these premises; more instances are expected. Each is handled the same way: a premise on what that participant itself
signed, never on the others, with the example recorded as an own-violation.

Second instance, `satp_end_user_recovers` (pebble `aln-223`; Opus counterexample
`docs/counterexamples/satp_end_user_recovers.lean` on `worker/aln-223`, `96f61c2`, checked by the coordinator): Alice's
latest `L0` allocation `o` holds a second guarantee `Vx` that names the virtual channel `r.V` as a peer. A reclaim of
`Vx` raises `r.V`'s entry, and `reclaim L0 r.V` never succeeds again. SATP Protocol 3.1 Round 2 (SATP:411-418) has
each participant sign only guarantees `G(a0 + b0, [P_i, P_{i+1}])`, whose peers are participants. Approved premise:
every guarantee in `o` names external destinations as its peers,
`∀ e ∈ o, ∀ x ps, e.item = .guarantee x ps → ∀ d ∈ ps, d.isExternal = true`. State domain unchanged.

Separate, open: the model's `Allocation.addTo` turns the item into `.alloc`; go-nitro `compute_reclaim_effects`
(`MultiAssetHolder.sol:340-352`) keeps `allocationType` and adds only to `amount`. Changing `addTo` changes an
approved definition and does not repair the theorem.

## D7 voucher rule repaired: a model defect, not an own-violation (operator, 2026-10-31)

- `satp_peers_can_finalize` (pebble `aln-50b`; Opus refutation `docs/counterexamples/satp_peers_can_finalize_voucher_stall.lean`
  on `worker/aln-50b`, `d5ee56a`, checked by the coordinator) is false in the model with SATP-honest behaviour only.
- Cause: repair D7 (armada-lean `docs/nitro-repair-2026-10-27.md` §2) makes `NetworkState.Holds` require, for a
  voucher signature of `q`, a received state that carries the voucher AND lists `q` among that state's signers. A
  peer that sees an on-chain challenge carrying a voucher signed by Alice and Bob, where only Alice signed the state,
  cannot resubmit it. The other side raises the turn with that proof and `V` stays open for ever.
- Sources: SATP Listing 3 (SATP:630-637): a voucher carries its own `signatures`; go-nitro calls are public calldata.
  A real peer that has seen the state can resubmit it.
- Ruling: drop `∧ q ∈ r.signers` from the voucher clause of `NetworkState.Holds`; a received state that carries the
  voucher is enough. This changes an approved definition. Every proof that reads `Holds` is rechecked by the build.

## Next

The statement change merges with `exophial integrate --branch` (it changes `scripts/sigs-baseline.tsv`). Pebble
`aln-cc8` was derived for the old conclusion and its history carries the counterexample, so it is closed and refiled
(operator order and reviewer-approved argv), as `aln-147` was.
