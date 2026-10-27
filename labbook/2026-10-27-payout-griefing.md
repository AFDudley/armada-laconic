# Payout griefing and the payout amount rule (2026-10-27)

Decisions on how much an Armada adjudicator payout pays, and how far another caller can delay a pending payout. Operator rulings are marked. Gas figures in the "Cost asymmetry" section are measured with `forge test --gas-report` (see that section for commits, tests, and commands).

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

**Minimum credit: proposed, then withdrawn (operator ruling 2026-10-27).** A minimum for a credit into a finalized channel was ruled first, so that each staling credit would cost the attacker value. It was dropped the same day:

- The measured gas asymmetry (next section, about 70–90 to 1) already bounds the attack economically.
- In the approved game model the honest participant chooses each move after the adversary's steps in that round, so the honest move is never stale in the model. Staleness is a latency and cost problem outside the theorems, as Nitro §4.2 leaves on-chain costs out of scope.
- The minimum made the model much larger: an outcome rule at `challenge`, `checkpoint` and `conclude`; three cases that lock value (an underfunded entry, a permanent remainder after a partial payout, a reclaim credit); an exception for credits that complete an entry; and a check inside `reclaim`. A ledger channel could also no longer send a small amount to a sub-channel.

**The attack stays documented, as not practical.** After `c` is finalized, an attacker can make a pending payout on `c` stale by crediting `c` from a channel of its own. There is no limit on the number of such credits. Each one costs the attacker two pool spends and a `conclude` (about 2.2–2.65M gas), plus the credited value, which goes to `c`'s entries. Each one costs the victim one failed call (about 30,000 gas, if the amount check comes first) and a new proof. The victim can usually avoid even the failed call by simulating before it sends.

## Cost asymmetry, and the check-order rule

Measured gas, `forge test --gas-report`, repo commits as shown. Each run used a fresh detached worktree (`git worktree add --detach .scratch/gas HEAD`); no fixturenet ran.

| Step | Function | Test | Min | Avg | Max | Verification | Repo @ commit | Command |
|---|---|---|---|---|---|---|---|---|
| pool spend (private transfer) | `PrivacyPool.transact` | `FeeTaggedOutputsTest.test_transact_feeOutput_countsOnlyFeeRecipient` | 1,022,034 | 1,022,034 | 1,022,034 | mocked (`testingMode`) | armada-poc @ `ec8bdb978749b06d4c28a22042123df370b9dcff` | `forge test --match-contract FeeTaggedOutputsTest --match-test test_transact_feeOutput_countsOnlyFeeRecipient --gas-report` |
| pool spend (shield, a deposit into the pool) | `PrivacyPool.shield` | `FeeTaggedOutputsTest.test_shield_toFeeRecipient_incrementsCumulativeFees` | 818,794 | 860,576 | 902,359 | n/a (shield carries no proof) | armada-poc @ `ec8bdb978749b06d4c28a22042123df370b9dcff` | `forge test --match-contract FeeTaggedOutputsTest --match-test test_shield_toFeeRecipient_incrementsCumulativeFees --gas-report` |
| pool spend (unshield) | `PrivacyPool.transact` (`SolvencyHandler.unshieldLocal`) | `PrivacyPoolSolvencyInvariantTest.test_handlerActionsAreLive` | 991,488 | 991,488 | 991,488 | mocked (`testingMode`) | armada-poc @ `ec8bdb978749b06d4c28a22042123df370b9dcff` | `forge test --match-contract PrivacyPoolSolvencyInvariantTest --match-test test_handlerActionsAreLive --gas-report` |
| pool spend (owned transact, contract-owned output) | `PrivacyPool.transactOwned` | `TransactOwnedTest.test_transactOwned_fromOwner_succeeds_recordsPreimage` | 866,247 | 866,247 | 866,247 | mocked (`testingMode`) | armada-poc @ `ec8bdb978749b06d4c28a22042123df370b9dcff` | `forge test --match-contract TransactOwnedTest --match-test test_transactOwned_fromOwner_succeeds_recordsPreimage --gas-report` |
| one Groth16 verification, in isolation (no pool call) | `Snark.verify` | `SnarkVerifierTest.test_realProofVerifies` | 234,901 | 234,901 | 234,901 | real (deployed Armada circuits fixture) | armada-poc @ `ec8bdb978749b06d4c28a22042123df370b9dcff` | `forge test --match-contract SnarkVerifierTest --match-test test_realProofVerifies --gas-report` |
| `conclude` (cooperative close, `conclude` + `transferAllAssets` in one call; no test calls bare `conclude` alone) | `NitroAdjudicator.concludeAndTransferAllAssets` | `ArmadaNitroAdjudicatorTest.test_cooperativeClose_shieldsUserNoteAndPaysCustodian` | 1,239,078 | 1,239,078 | 1,239,078 | real ECDSA signature checks (`ForceMove`); pool leg mocked (`testingMode`, see `test/helpers/ArmadaNotes.sol`) | armada-nitro @ `30790220585feffee922adcd76ba0958184eeb42` | `forge test --match-contract ArmadaNitroAdjudicatorTest --match-test test_cooperativeClose_shieldsUserNoteAndPaysCustodian --gas-report` |
| `challenge` | `ForceMove.challenge` | `ArmadaNitroAdjudicatorTest.test_staleHubChallenge_isOverriddenByUsersLatestState` | 115,848 | 127,389 | 138,931 | real ECDSA signature checks | armada-nitro @ `30790220585feffee922adcd76ba0958184eeb42` | `forge test --match-contract ArmadaNitroAdjudicatorTest --match-test test_staleHubChallenge_isOverriddenByUsersLatestState --gas-report` |
| `transfer` (force-close payout after challenge expiry) | `MultiAssetHolder.transfer` | `ArmadaNitroAdjudicatorTest.test_transfer_partial_pays_destination_fresh_commitment` | 1,183,606 | 1,183,606 | 1,183,606 | real; pool leg mocked (`testingMode`) | armada-nitro @ `30790220585feffee922adcd76ba0958184eeb42` | `forge test --match-contract ArmadaNitroAdjudicatorTest --match-test test_transfer_partial_pays_destination_fresh_commitment --gas-report` |
| `depositFromNote` | `ArmadaNitroAdjudicator.depositFromNote` | `ArmadaNitroAdjudicatorTest.test_cooperativeClose_shieldsUserNoteAndPaysCustodian` | 1,024,200 | 1,038,778 | 1,053,357 | real preimage/owner checks; pool leg mocked (`testingMode`) | armada-nitro @ `30790220585feffee922adcd76ba0958184eeb42` | `forge test --match-contract ArmadaNitroAdjudicatorTest --match-test test_cooperativeClose_shieldsUserNoteAndPaysCustodian --gas-report` |
| a stale payout that fails at the adjudicator's amount check | — | no existing test exercises this exact revert | — | — | — | [not measured; no matching test] | — | — |

A note on "mocked (`testingMode`)": `PrivacyPool.transact` and `PrivacyPool.transactOwned` skip the Groth16 pairing check when `testingMode` is on (every pool test in both repos runs this way). The isolated `Snark.verify` call above shows the real cost a mocked `transact`/`transactOwned`/`depositFromNote` figure omits: adding 234,901 gas to a mocked pool-spend figure approximates its real-proof cost (for example, `transact` at 1,022,034 mocked becomes roughly 1,256,935 with a real proof).

Per staling credit, the attacker pays two pool spends and one `conclude`. The measured `concludeAndTransferAllAssets` (1,239,078) is not a bare `conclude`: it also runs a payout, so it overstates the step. A bare `conclude` does signature checks and one status write, like `challenge` (measured 115,848 to 138,931). With two real-proof pool spends (2 × 1,256,935) and a `conclude` near the `challenge` maximum (about 140,000), the attacker pays about 2,650,000 gas, before the credited amount. The victim's one failed call has no matching test, so it stays an estimate: tens of thousands of gas. At 30,000 gas, the attacker pays about 90 times the victim's gas. With mocked pool spends (2 × 1,022,034 + 140,000 ≈ 2,180,000), the ratio is about 70. Two figures here are not measured: the bare `conclude` and the victim's failed call.

That ratio holds only if a payout checks its amount before it touches the pool. If the check came after the proof check and the pool spend, each stale payout would cost the victim a full pool spend.

**Specification rule (new theorem):** a payout whose amount differs from the channel's affordable amount fails on the adjudicator's state alone, for every pool witness and proof. The refinement to the contract must place this check before the pool call.

## Theorems to add (armada-nitro-spec, types first)

- The payout rule: a transfer's amount equals the affordable amount (B).
- The check-order theorem above.
- Non-interference as a list of causes: a pending payout on `c` goes stale only through a payout of one of `c`'s entries, a reclaim against `c`, or a credit into `c`. Reclaims are at most the number of guarantees. Credits have no count bound in the theorems; their cost is the attack above.

## Follow-up

Done: the "Cost asymmetry" section now has the measured gas table (2026-10-27). Open item: the victim's failed-call row (a stale payout rejected at the adjudicator's amount check) still has no matching test in either repo; a future test that exercises exactly this revert would let that row move from estimate to measurement.
