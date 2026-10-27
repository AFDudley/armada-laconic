# Payout griefing and the payout amount rule (2026-10-27)

Decisions on how much an Armada adjudicator payout pays, and how far another caller can delay a pending payout. Operator rulings are marked. Gas figures are typical magnitudes, not measured; the measurement is a follow-up (end of this note).

## The problem

An Armada channel holds its assets as pool notes. A payout is a pool spend, and its zero-knowledge proof fixes the notes it spends and the amounts it creates. If another call changes what the payout may pay before it runs, the proof no longer matches and the payout fails ("stale"). The payee must build a new proof and try again. This is griefing: it costs the payee time and gas, not value.

## What the sources say

- Nitro (2019) §5.1: `T(A, B, x)` pays any `x` that A can afford for B, and fails otherwise. Any caller may run it.
- Front-running griefing exists in Nitro's account model too: an attacker pays B 1 unit first, and B's transfer of `x` fails. Nitro's answer: B loses no value (every unit goes to B), unbeatable strategies are adaptive (§4.3), and "Modelling the cost of the on-chain operations is beyond the scope of this paper" (§4.2).
- ForceMove (2018) §4.5: griefing "cannot be completely eliminated"; its defences restrict and certify the harmful call (challenges).
- go-nitro (`64c4ee9a1`): `transfer(indices)` has no amount; it pays the full affordable amount. Every transfer rewrites the outcome hash, so a transfer makes other pending transfers on the channel fail (`_requireMatchingFingerprint`). Because each entry is paid in full, each entry can cause this at most once.

## Options considered

| Option | Rule | Griefing |
|---|---|---|
| A | pay any `x ≤` affordable (Nitro paper) | unbounded: an attacker pays the payee 1 unit at a time, for gas only, and each payout stales the pending one |
| B | pay exactly the affordable amount (go-nitro) | bounded, see below |
| C | third parties pay the full amount; only the payee may pay less | moves the problem; needs payee authentication that no source has |

**Operator ruling: option B.**

## Bounds under B

After channel `c` is finalized, a pending payout on `c` can go stale only through:

| Source | Bound | Attacker's cost per staling |
|---|---|---|
| a payout of one of `c`'s entries, in full | number of entries in `c`'s outcome | gas; the payout pays a real entry in full |
| a reclaim against `c` | number of guarantees in `c`'s outcome | gas |
| a credit into `c` from another channel's payout | none by count: the attacker can open any number of channels that pay `c` | the credited amount plus gas; the amount goes to `c`'s entries |

Credits into a finalized channel cannot be forbidden: a sub-channel can finalize before its ledger channel pays it.

**Operator ruling: a credit into an already-finalized channel must be at least a minimum fixed at deployment.** A credit smaller than the cost of collecting it is dust: it gives `c`'s entries nothing of value and serves only to grief. With the minimum, each staling by credit costs the attacker at least the minimum, paid to the victim channel's entries. The value is open; it should be near the cost of collecting one payout.

## Cost asymmetry, and the check-order rule

Typical magnitudes [not measured]:

| Step | Cost |
|---|---|
| pool spend (proof check + Poseidon Merkle insertions) | hundreds of thousands of gas, up to about 1M |
| `conclude` (signature checks, one status write) | tens of thousands of gas |
| building a proof | seconds of client CPU, no gas |
| a stale payout that fails at the adjudicator's amount check | tens of thousands of gas |

Per staling credit, the attacker pays two pool spends (a deposit into its own channel, the transfer into `c`), one `conclude`, and the credited amount. The victim pays one cheap failed call and one new proof, and can usually avoid the failed call by simulating first. So the attacker pays tens of times more than the victim.

That ratio holds only if a payout checks its amount before it touches the pool. If the check came after the proof check and the pool spend, each stale payout would cost the victim a full pool spend.

**Specification rule (new theorem):** a payout whose amount differs from the channel's affordable amount fails on the adjudicator's state alone, for every pool witness and proof. The refinement to the contract must place this check before the pool call.

## Theorems to add (armada-nitro-spec, types first)

- The payout rule: a transfer's amount equals the affordable amount (B).
- The minimum credit into a finalized channel, as one named deploy-time parameter.
- The check-order theorem above.
- Non-interference as a bound: the steps that can make a pending payout on `c` stale are payouts of `c`'s entries, reclaims against `c`, and credits into `c`, each credit at least the minimum.

## Follow-up

Measure the gas of a pool transact, a shield, an owned transact and an adjudicator `conclude` with `forge test --gas-report` on the existing armada-poc and armada-nitro tests, and replace the magnitudes above with the measured figures.
