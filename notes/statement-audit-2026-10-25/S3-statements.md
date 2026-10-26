# S3 — SATP (2022) statement extraction

Source: `~/code/git_puller/repos/armada-laconic/.scratch/rewrite-B/satp-2022.txt`
(pdftotext -layout of `armada-nitro/docs/references/nitro/sources/satp-2022.pdf`).
Read start to end. Last line read: **789** (line 790 is blank).
Source tag: **S3**. Line numbers below are the `.txt` line numbers.

Paper: Stewart, Kennedy, Kerzhner, Knee, Geihs, Stammler, "SATP: A simple and
scalable protocol for virtual state channel networks", 2022-09-28.

## Scope note (important)

SATP specifies only TWO on-chain adjudicator operations with code listings:
`challenge` (Listing 1, lines 300-315) and `reclaim` (Listing 2, lines 344-364).
`conclude` is named in prose only (§2.3, lines 286-289); there is no listing.
There is **no** `deposit`, `withdraw`, `transfer`, `checkpoint`, or `claim`
listing in SATP. Appendix B.1 (lines 764-768) states this directly: "This work
glosses over depositing into and withdrawing from a channel... We will write a
full specification of SATP including deposits and withdrawals." So the model's
`deposit`, `topup`, `transfer`, and `checkpoint` are NOT grounded in S3; they
come from Nitro 2019 / ForceMove / go-nitro. The task brief named those listings;
the source does not contain them. I report that gap, not an absence of effort.

SATP replaces Nitro 2019's `claim` with `reclaim` (§2.4, footnote 11, line 381:
"This decoupling of ledger channel updates is the main contribution of Nitro
protocol"). The docs site says SATP matches the current implementation, so
`reclaim` is the newer design. See the "SATP vs Nitro 2019" section.

---

## Definitions the model must match (class d)

### S3-01 — ChannelConsts (constants)
Quote (lines 198-203):
```
interface ChannelConsts {
  peers : PublicKey []
  appDef : Address
  nonce : uint
  challengeDuration : uint
}
```
Property: a channel's fixed part has peers (public keys), an application
definition address, a nonce, and a challenge duration. Caller cannot change
these after channel creation.
Coverage: COVERED by `FixedPart` (`Operations.lean`/`Channel.lean`): the model's
`fp.challengeDuration` is used in `.challenge` (`Operations.lean:204`), and
`channelId fp` folds peers/appDef/nonce.

### S3-02 — ChannelVars (variables)
Quote (lines 205-211):
```
interface ChannelVars {
  version : uint
  isFinal : boolean
  outcome : Outcome
  appData : bytes
  signatures : Signature []
}
```
Property: a channel's variable part has a monotone version, an isFinal flag, an
outcome, app data, and signatures.
Coverage: PARTIAL. The model's `VariablePart` uses `turnNum` for `version`,
`isFinal`, `outcome`. SATP calls the counter `version`; the model calls it
`turnNum`. Same role. `appData`/`signatures` folded into `AuthorizedBy`/`appRel`.

### S3-03 — ChannelState
Quote (lines 218-221): `interface ChannelState { consts, vars }`.
Property: a full state is constants plus variables.
Coverage: COVERED (FixedPart + VariablePart).

### S3-04 — AdjudicationState / statusOf
Quote (lines 234-241):
```
interface AdjudicationState { version : uint; outcome : Outcome; finalizationTime : uint }
The adjudicator stores adjudication states in a mapping statusOf, using channel ids as keys.
```
Property: the adjudicator stores, per channel id, a version, an outcome, and a
finalization time.
Coverage: COVERED by `ChannelStatus` (`Channel.lean`): `turnNumRecord`,
`outcomeHash`, `finalizesAt`. Note: SATP stores `outcome`; the model stores
`outcomeHash` (hash of the outcome) — an encoding choice, not a difference in
meaning.
INTERNAL INCONSISTENCY in SATP: the interface field is `finalizationTime`
(line 237), but every code listing uses `finalizesAt` (lines 297, 313, 348-349).
The model uses `finalizesAt`, which matches the listings.

### S3-05 — channel id
Quote (line 213): "The channel's id is computed from its constants as
hash(peers, appDef, nonce)."
Property: channelId = hash(peers, appDef, nonce). The nonce lets a fixed peer
set make many distinct channels (lines 214-215).
Coverage: COVERED by `channelId fp`. NOTE: SATP's hash input is
(peers, appDef, nonce) with no chainId. Nitro 2019 / go-nitro include chainId.
See "SATP vs Nitro 2019".

### S3-06 — SupportProof
Quote (lines 247-250): `interface SupportProof { consts, vars : ChannelVars[] }`.
Also line 244: "A support proof is an array of states, each state containing one
or more signatures from the channel's peer list."
Property: a support proof is one constants object plus an array of variable
parts, each with at least one peer signature.
Coverage: COVERED. The model passes `proof : List VariablePart` plus a
`candidate` to `.challenge`/`.checkpoint` (`Operations.lean`), gated by `appRel`.

### S3-07 — latestSupportedState signature + meaning
Quote (lines 254-259):
```
function latestSupportedState(proof : SupportProof) : ChannelState
```
"The variable part returned is understood by the adjudicator to be the most
recent version of the channel's state to be supported by all peers in the
channel, even if not explicitly signed by every peer."
Property: the application maps a support proof to the single latest supported
state. "Supported" is defined per-app and may hold even without every signature.
Coverage: COVERED as the model's abstraction: `appRel` is a universal parameter
(`Safety.lean` sections A-C quantify over any app), standing in for
`latestSupportedState`. The can-happen theorems instantiate it with the real
ConsensusApp (`Safety.lean` "concrete app" section).

### S3-08 — Consensus app rule
Quote (lines 262-266):
```
function latestSupportedState({const, vars}) {
  require(vars.length == 1)
  require(signedByEveryone(vars[0], consts.peers))
  return {consts, vars: vars[0]}
}
```
(Appendix A restates it as `requireConsensus`, lines 639-642:
`require(vars.length == 1); require(signedByAll(vars[0], consts.peers))`.)
Property: the consensus app accepts exactly one state, signed by every peer.
Coverage: COVERED by `ConsensusApp` (`Safety.lean` "concrete app"): empty proof
plus `SignedByAll`.

### S3-09 — Allocation
Quote (lines 320-321): "An allocation is a data structure A(a) representing an
amount a to be paid to a destination."
Coverage: COVERED by the model's allocation records (`Outcome.lean`).

### S3-10 — Guarantee
Quote (lines 322-323): "A guarantee is a data structure G(x, [A, B]) representing
an amount x and an ordered list of destinations A and B."
Property: a guarantee carries an amount and an ORDERED destination list.
Coverage: COVERED by `decodeGuarantee`/`reallocateGuarantee` (`Outcome.lean`);
`reclaim` uses the two ordered destinations.

### S3-11 — item.amount notation
Quote (lines 324-325): "item.amount = x when either item = A(x) or item =
G(x, [A, B])."
Class: d (notation). No separate model obligation.

### S3-12 — Outcome = ordered dictionary; order is priority
Quote (lines 334-341): "An outcome is an ordered dictionary, mapping a
destination to either an allocation or a guarantee... Keys are ordered from left
to right... o[0] is o[A], o[1] is o[X], o[2] is o[B]."
Footnote 9 (line 377): "the order may also dictate which destinations receive
priority in case the channel is insufficiently funded."
Property: an outcome is an ordered list of (destination -> allocation | guarantee)
entries; index order encodes payout priority under short funding.
Coverage: COVERED as `Outcome` is a `List` with index-based access; `afford`
(`Outcome.lean`) consumes held value in index order — this IS the priority rule.

### S3-13 — isOpen
Quote (lines 294-298):
```
function isOpen(status) { return status.finalizesAt == 0 || status.finalizesAt > now }
```
Property: a channel is open when it has no recorded finalization time, or its
finalization time is still in the future.
Coverage: COVERED by `ChannelStatus.mode` (`Channel.lean`): open when
`finalizesAt == 0` or `finalizesAt > now`; finalized when `0 < finalizesAt <= now`.

---

## On-chain adjudicator operations (class a safety + class c can-happen)

### S3-14 — challenge (Listing 1)
Quote (lines 300-315):
```
function challenge({consts, vars}: SupportProof) {
  let id = channelId(consts)
  let adjudicatorStatus = statusOf(id)
  require(isOpen(adjudicatorStatus))
  let supportedStateVars = Application.at(state.appDef).latestSupportedState(consts, vars).vars
  require(supportedStateVars.version >= adjudicatorStatus.version)
  statusOf(id) = {
    outcome: supportedStateVars.outcome,
    version: supportedStateVars.version,
    finalizesAt: now + consts.challengeDuration,
  }
}
```
Caller: any peer. The caller chooses the support proof. The caller cannot change
the result unless the proof passes `latestSupportedState` and carries a version
at least the recorded one.
Properties:
- (a) challenge succeeds only when the channel is open (S3-13).
- (a) challenge succeeds only when the supported version is >= the recorded
  version (no caller can lower the recorded version).
- (a) challenge records the supported outcome and version, and sets
  `finalizesAt = now + challengeDuration`.
- (c) a challenge must be possible for a validly supported state.
Coverage:
- Safety / monotonicity: COVERED by the `.challenge` `Step` case
  (`Operations.lean:194-206`): open mode requires `candidate.turnNum >=
  turnNumRecord`, challenge mode requires strict `>`, finalized mode is `False`;
  it sets `finalizesAt := now + challengeDuration`.
  PARTIAL difference: SATP uses a single `version >= recorded` check (via
  `isOpen`, which stays true during an unexpired challenge), so SATP allows a
  re-challenge at an EQUAL version while a challenge is active. The model
  requires strict `>` during an active challenge. The model is stricter. This is
  a ForceMove-2019 refinement, not an SATP rule.
- Can-happen: COVERED by `challenge_can_happen` (`Safety.lean:1383`) and
  `challenge_timeout_can_happen` (`Safety.lean:1434`).

### S3-15 — reclaim (Listing 2)
Quote (lines 344-364):
```
function reclaim(ledgerChannelId, targetChannelId) {
  let ledgerStatus = statusOf(ledgerChannelId)
  let targetStatus = statusOf(targetChannelId)
  require(0 < ledgerStatus.finalizesAt <= now)
  require(0 < targetStatus.finalizesAt <= now)
  let guarantee = ledgerStatus.outcome[targetChannelId]
  require(isGuarantee(guarantee))
  let [left, right] = guarantee.peers
  let o_ledger = ledgerStatus.outcome
  let o_target = targetStatus.outcome
  o_ledger[left] += o_target[0]
  o_ledger[right] += o_target[1]
  delete o_ledger[targetChannelId]
  statusOf(ledgerChannelId).outcome = o_ledger
}
```
Caller: any peer (no signature check in the listing). The caller chooses which
ledger channel and which target channel.
Properties:
- (a) reclaim succeeds only when BOTH the ledger channel and the target channel
  are finalized (`0 < finalizesAt <= now` for each).
- (a) reclaim succeeds only when the ledger's entry for the target is a guarantee.
- (a) reclaim credits the guarantee's first destination (`left`) by the target's
  first allocation (`o_target[0]`) and the second destination (`right`) by the
  target's second allocation (`o_target[1]`), then deletes the guarantee entry.
- (a) conservation: the total moved into the ledger equals `o_target[0] +
  o_target[1]`; no value is minted or burned.
- (c) a reclaim must be possible once both channels are finalized.
Coverage:
- Safety: COVERED by the `.reclaim` `Step` case (`Operations.lean:360-...`):
  requires source finalized (line 361), target finalized (line 369), reads the
  guarantee, and applies `reallocateGuarantee` with the two ordered destinations
  and `a0`, `a1` (`Outcome.lean`). Holdings/pool unchanged (line ~355 note).
- Can-happen: COVERED by `reclaim_can_happen` (`Safety.lean:1876`).

### S3-16 — assets released only on explicit consent
Quote (lines 284-289): "assets deposited into an adjudicator contract are
released only when given explicit consent on the final outcome by the channel's
participants. As timeouts significantly worsen user experience, adjudicators
typically allow peers to collaboratively conclude a channel without delay, by
supporting a state with isFinal = true and submitting the support proof to a
Conclude operation."
Properties:
- (a) no caller can release a channel's assets without either a finalized outcome
  (via challenge timeout) or a concluded (isFinal, all-signed) outcome.
- (c) a collaborative conclude via `isFinal = true` must be possible with no delay.
Coverage:
- (a) COVERED by `finalized_outcome_was_concluded_or_challenged`
  (`Safety.lean:125`): a finalized outcome traces back to a conclude (SignedByAll)
  or a challenge step.
- (c) COVERED by `conclude_can_happen` (`Safety.lean:1498`).
NOTE: SATP gives no `conclude` listing; only this prose. The model's `.conclude`
case (`Operations.lean:262-270`) requires `isFinal = true` and `SignedByAll`,
which matches the prose.

---

## Virtual-channel application rules (Appendix A) — class d/e

These are off-chain application logic that the adjudicator runs through
`latestSupportedState`. The model abstracts the app as `appRel`, so it does not
implement this specific app. I list the rules; they are off-chain app behavior,
not core adjudicator logic. The one on-chain-enforced safety fact they produce is
the value invariant in S3-19.

### S3-17 — latestSupportedState dispatch by version
Quote (lines 644-667): version <= 1 -> requireConsensus (prefund v0, postfund v1);
version == 4 -> requireConsensus and require isFinal (concluding);
else -> validateOutcomeChange. Returns the last state received.
Class: d/e (app logic). Coverage: not modeled (app is `appRel`); acceptable.

### S3-18 — validateOutcomeChange transitions
Quote (lines 669-741): state-type lattice None -> A or B -> AB. `vars.length in
[1,2,3]`. start.type == None, signed by all; challenge.version == 2, signed by
the mover; response.version == 3, type == AB, signed by the other; vouchers
co-signed; `voucher2.version > voucher1.version`; outcomes equal the voucher
amounts.
Class: d/e. Coverage: not modeled; acceptable.

### S3-19 — intermediary-funds invariant (voucher sum)
Quote (lines 705-717, 734-738):
```
let x = start.outcome[0].amount + start.outcome[1].amount
// The following two requirements guarantee that intermediary's funds are invariant
require(voucher1.alice + voucher1.bob == x)
require(challenge.outcome == { alice: voucher1.alice, bob: voucher1.bob })
...
require(voucher2.alice + voucher2.bob == x)
```
Also Lemma A.1 (lines 746-753): "intermediaries can ensure that o[A] + o[B] are
invariant by only committing to a state with isFinal = true when it doesn't
change the value of o[A] + o[B]."
Property (class a, enforced by the app on-chain): any outcome change keeps the
total allocated amount `o[A] + o[B]` equal to the funded total `x = a0 + b0`.
No caller can make the virtual channel finalize with a different total.
Coverage: MISSING in the model as a stated theorem. The model's `appRel` is
universal and does not carry this invariant. The closest core-adjudicator fact
is conservation of pool value (`notedValue_conserved`), which is a DIFFERENT
guarantee (pool-wide, not per-channel outcome total). SATP's sum-preservation is
an app-level property the model does not state. Mark MISSING (off-chain app, but
the operator asked to list even changed-form properties).

---

## Security / liveness claims (class b/e/f)

### S3-20 — Theorem 3.2: intermediaries' balances are invariant
Quote (lines 503-507): "Let d(V, Pi) be the amount debited to V from Pi's
balances... Let c(V, Pi) be the amount credited to Pi's balance from guarantees
targeting V... For all V, for all i = 1,...,n, if Pi follows Protocols 3.1-3.3,
then c(Pi) = d(Pi)."
Also lines 498-500: "intermediaries' balances are invariant and are not locked up
indefinitely in the virtual channel."
Class: e (off-chain protocol guarantee; depends on Protocols 3.1-3.3 and the app
rules). Caller: an honest intermediary Pi. The guarantee is: an honest
intermediary who follows the protocol ends with its balance unchanged.
Coverage: MISSING as a direct statement. The model's `reclaim` + conservation
underpin it, but the full "credited == debited for every honest intermediary"
statement is not in the model. Off-chain; list it, do not drop it.

### S3-21 — Lemma 3.1 / Lemma A.1: finalize total is preserved
Quote (lines 511-513): "Suppose V finalizes on-chain. Either V is finalized with
version = 0 and its original outcome {A : A(a0), B : A(b0)} or it is finalized
with an outcome {A : A(a), B : A(b)}, where a0 + b0 = a + b."
Lemma A.1 (lines 746-750): same, restated with turns.
Class: a (on-chain-enforced via the app) + b (peers "can take unilateral actions
to ensure V finalized with an outcome {A : a, B : b}" — an unbeatable-strategy /
liveness claim, line 748-749).
Coverage: MISSING. This is the same per-channel total-preservation as S3-19, plus
the liveness half ("peers can ensure finalization"). The model's `unbeatable_pays`
/ `unbeatable_pays_after_timeout` (`Safety.lean:444`, `:500`, both carry `sorry`)
state an analogous liveness guarantee for the Nitro layer, not for the SATP
virtual-channel app. Partial overlap in spirit; the SATP sum-preservation is not
stated.

### S3-22 — Theorem 3.1: Protocol 3.3 cost and time
Quote (lines 491-493): "Protocol 3.2 [3.3] consumes O(1) gas and terminates in
time O(max(Li-1.challengeDuration, Li.challengeDuration) + V.challengeDuration)."
Class: e (off-chain complexity claim). Coverage: MISSING / out of scope (the Lean
model does not reason about gas or wall-clock complexity).

### S3-23 — Conjecture 3.1: two-round lower bound
Quote (lines 451-452): "A state channel protocol requires at least two rounds to
securely fund a virtual channel."
Class: e (off-chain conjecture). Coverage: MISSING / out of scope.

### S3-24 — Liveness assumption
Quote (lines 166-167): "(Liveness) Any peer can successfully record transactions
on the blockchain in O(1) time."
Class: f (assumption). Coverage: implicit in the can-happen theorems (a peer can
submit each call). Not stated as a named assumption.

### S3-25 — Safety assumption (unforgeable signatures)
Quote (line 168): "(Safety) Forging cryptographic signatures is infeasible."
Class: f (assumption). Coverage: COVERED by `SignatureAssumptions`
(`Assumptions.lean`), threaded through every call-sequence theorem. The model
spans two schemes (Nitro ECDSA, pool EdDSA); the assumption applies to both
(`Operations.lean:462-464`).

### S3-26 — deposits / withdrawals deferred
Quote (lines 764-768): "This work glosses over depositing into and withdrawing
from a channel... We will write a full specification of SATP including deposits
and withdrawals."
Class: f (scope statement). Consequence: the model's `deposit`, `topup`,
`transfer` (withdrawal/payout), and `checkpoint` are NOT grounded in S3. They
come from Nitro 2019 / ForceMove / go-nitro.

---

## Coverage of the 24 model theorems against S3

- `finalized_outcome_was_concluded_or_challenged` — COVERED by S3-16
  (concluded-or-challenged) + S3-14/S3-15.
- `transfer_pays_only_the_finalized_outcome` — NOT IN S3. SATP has no `transfer`
  (withdrawal) listing (S3-26). Grounded in Nitro 2019, not SATP.
- `unbeatable_pays` / `unbeatable_pays_after_timeout` — PARTIAL match to the
  liveness half of S3-21 (Lemma A.1 "peers can take unilateral actions to ensure
  V finalized"). SATP states it for the virtual-channel app total, not for the
  generic Nitro allocation payout. Both carry `sorry`.
- `notedValue_conserved` / `notedValue_conserved_reachable` — PARTIAL. SATP's
  conservation facts are per-channel outcome-total (S3-15 reclaim, S3-19/S3-21).
  Pool-wide note conservation is an Armada encoding property, not an SATP claim.
- `holdings_point_to_live_notes` — NOT IN S3 (Armada note-pool encoding).
- `deposit_can_happen` / `topup_can_happen` — NOT IN S3 (S3-26 defers deposits).
- `challenge_can_happen` — COVERED by S3-14 (c).
- `challenge_timeout_can_happen` — COVERED by S3-13 + S3-16 (timeout finalizes).
- `conclude_can_happen` — COVERED by S3-16 (c).
- `payout_can_happen` — NOT IN S3 (no withdrawal/transfer; S3-26).
- `reclaim_can_happen` — COVERED by S3-15 (c).
- `payout_takes_no_fee` — NOT IN S3 (no withdrawal listing). SATP never mentions
  a payout fee; the no-fee claim is from Nitro 2019 §5.1.
- `conclude_moves_no_value` — PARTIAL. SATP conclude is prose (S3-16); it says
  conclude finalizes, says nothing explicit about moving value. Consistent, not
  stated.
- `bulk_paths_never_succeed` — NOT IN S3 (go-nitro bulk entry points; not in SATP).
- `value_exists_only_as_notes` — NOT IN S3 (Armada encoding).
- `holdings_never_zero_valued` — NOT IN S3 (Armada/circuit encoding).
- `heldValue_zero_iff_no_collateral` — NOT IN S3 (Armada encoding).
- `configuredAssets_invariant` — NOT IN S3 (go-nitro asset configuration).
- `zero_payout_mints_nothing` — NOT IN S3 (transfer/payout; S3-26).
- `full_payout_leaves_no_remainder` — NOT IN S3 (transfer/payout; S3-26).
- `nullify_iff_commitment_was_some` — NOT IN S3 (Armada note nullifier encoding).

SATP grounds a NARROW slice of the model: channel/outcome definitions (S3-01..13),
`challenge` (S3-14), `reclaim` (S3-15), conclude-or-challenge finalization (S3-16),
and the signature-unforgeability assumption (S3-25). The payout/deposit/checkpoint
family and all Armada note-encoding theorems trace to Nitro 2019 / ForceMove /
go-nitro, not SATP.

---

## SATP vs Nitro 2019 (and other sources) — disagreements

- **reclaim vs claim.** SATP (§2.4, Listing 2, lines 342-364) names the guarantee
  payout operation `reclaim(ledgerChannelId, targetChannelId)`. Nitro 2019 names
  it `claim`. SATP footnote 11 (line 381) calls this guarantee-decoupling "the
  main contribution of Nitro protocol." SATP's `reclaim` credits the guarantee's
  two ordered destinations by the target's first two allocations and deletes the
  guarantee entry. The docs site says SATP matches the implementation, so
  `reclaim` is the newer name. The model uses `reclaim` — it follows SATP.
- **No transfer / withdrawal in SATP.** Nitro 2019 / ForceMove specify a
  `transfer` (pay allocations out) operation. SATP omits it and defers it to
  future work (S3-26, lines 764-768). The model's `transfer` follows Nitro 2019.
- **channelId inputs.** SATP (line 213) hashes (peers, appDef, nonce) — no
  chainId. Nitro 2019 / go-nitro include chainId. Difference to confirm against
  S2/S1 reports.
- **challenge re-challenge rule.** SATP `challenge` allows resubmission at an
  EQUAL version while a challenge is unexpired (`isOpen` stays true, check is
  `version >=`). ForceMove-2019 / the model require strict `>` during an active
  challenge. SATP is more permissive here.
- **conclude: listing vs prose.** SATP gives `conclude` as prose only (§2.3). The
  model's `.conclude` case follows the Nitro/ForceMove `isFinal + SignedByAll`
  rule.
- **Internal SATP inconsistency.** `AdjudicationState.finalizationTime` (interface,
  line 237) vs `finalizesAt` (all listings, lines 297/313/348). The listings win;
  the model uses `finalizesAt`.

---

## Class counts (26 extracted properties, S3-01..S3-26)

- (a) on-chain adjudicator safety: 4 — S3-14, S3-15, S3-16, S3-19 (part),
  S3-21 (part).
- (b) on-chain liveness / unbeatable strategy: 1 — S3-21 (the "peers can ensure
  finalization" half of Lemma A.1).
- (c) can-happen: 3 — S3-14, S3-15, S3-16 (each also carries a can-happen clause).
- (d) definitions the model must match: 13 — S3-01..S3-13 (plus S3-17/S3-18 app
  logic, listed as d/e).
- (e) off-chain protocol property: 5 — S3-17, S3-18, S3-20, S3-22, S3-23.
- (f) assumptions: 3 — S3-24, S3-25, S3-26.

(Several properties carry two classes; counted under each above. 26 distinct IDs.)

## MISSING and CONTRADICTED (one line each)

- MISSING S3-19 — per-channel outcome total invariant (voucher sum `alice+bob==x`,
  intermediary funds invariant); the model's universal `appRel` does not state it.
- MISSING S3-20 — Theorem 3.2 `c(Pi) == d(Pi)` honest-intermediary balance
  invariance; not stated in the model.
- MISSING S3-21 — Lemma 3.1 / A.1 finalize-total preservation (`a0+b0 == a+b`);
  model's `unbeatable_pays` is the Nitro-layer analogue only.
- MISSING S3-22 — Theorem 3.1 O(1) gas and time bound; out of the Lean model's scope.
- MISSING S3-23 — Conjecture 3.1 two-round lower bound; out of scope.
- CONTRADICTED (minor) S3-14 — SATP allows re-challenge at EQUAL version during an
  active challenge (`version >=` + `isOpen`); the model requires strict `>` during
  an active challenge (`Operations.lean:197`). The model follows ForceMove-2019,
  which is stricter than SATP's listing.
- No other contradiction; the model's `reclaim` (S3-15) agrees with SATP's
  listing.
