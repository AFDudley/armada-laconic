# S1 — ForceMove (Close, Stewart, 2018) — property extraction

Source: `~/code/git_puller/repos/armada-laconic/.scratch/rewrite-B/forcemove-2018.txt`
(pdftotext -layout of `armada-nitro/docs/references/nitro/sources/forcemove-2018.pdf`).
I read every line, 1 to 1901. Line 1902 is a page-number footer. Last line read: 1902.

Model compared: `~/code/git_puller/repos/armada-lean-adjudicator-v4` at `5f5116a`,
`ArmadaLean/Adjudicator/Safety.lean` (24 theorems), `Channel.lean` (modes).

## How ForceMove relates to the model

The model is a Nitro/ForceMove adjudicator. ForceMove is the ancestor paper.
Nitro replaced several ForceMove operations. SATP and the statechannels docs
say ForceMove is partly out of date. I extract every property, and I mark the
items that the newer sources replace when the text shows the difference.

Three ForceMove facts drive the mode mapping:
1. ForceMove has five modes (PreFundSetup, Collaborative, Challenge, Concluded,
   Terminated; Figure 5, lines 1237-1247).
2. The model has three modes: `open`, `challenge`, `finalized`
   (`Channel.lean:182-222`).
3. ForceMove says a concluded game "behaves exactly the same as one that was
   terminated" (lines 1504-1507). So the model collapses PreFundSetup +
   Collaborative into `open`, and Concluded + Terminated into `finalized`.

Armada encoding note: ForceMove "resolution" (an outcome array) is held by the
adjudicator and paid out. In Armada a payout is a pool note at the recipient's
note key, not an ERC20 balance. I keep each property and report its Armada form.

---

## Class (d) — definitions the model must match

### S1-D01 — Channel and channelId
Quote (600-607):
```
struct Channel contains
   channelType: Address
   participants: Address[]
   channelNonce: uint
Channel::id c 7→ h(c.gameType, c.participants, c.channelNonce)
Channel::numberOfParticipants c 7→ c.participants.length
```
Property: a channel is `(channelType, participants, channelNonce)`. The channel
id is the hash of `(gameType, participants, channelNonce)`. `numberOfParticipants`
is the length of `participants`.
Caller: none; this is a format.
Coverage: PARTIAL. Model `channelId fp = channelIdOf fp.participants
fp.channelNonce fp.appDefinition fp.challengeDuration` (`Channel.lean:66-67`).
Difference: Nitro renames `channelType`/`gameType` to `appDefinition` and ADDS
`challengeDuration` to the hash inputs. ForceMove has no `challengeDuration` in
the channel.

### S1-D02 — channelId uniqueness and replay risk
Quote (610-632): "The resulting channelId is designed to uniquely identify the
channel"; "A failure to [pick a unique nonce] can lead to funds being lost due
to replay attacks from the previous channel. If the channel id is not unique,
players should refuse to join the channel."
Property: the channel id must be unique. A non-unique id enables a replay
attack and fund loss.
Caller: participants pick the nonce.
Coverage: MISSING as a theorem. The model takes `channelId` as a hash value and
does not state a collision-freedom or replay property. This is partly an
assumption (hash injectivity) — the model uses `outcomeHashOf_inj` and a
`HashAssumptions` argument, so hash injectivity exists in the model, but no
theorem states the replay/uniqueness consequence.

### S1-D03 — Outcome format
Quote (644-665): "an outcome consists of a list of addresses and the assets
that the game has allocated to them"; "a ForceMove game, g, is closed if, in
all possible outcomes, it only allocates funds to the participants"; "outcome,
of length γ.numberOfParticipants, where outcome[i] represents the amount of
coin to be distributed to participants[i]"; `type Outcome := Uint[]`.
Property: an outcome maps each participant index to an amount. A closed game
only pays its own participants.
Coverage: PARTIAL. Model `Outcome` is a richer list of asset exits with
allocations and guarantee metadata (`Outcome.lean`, used throughout Safety).
Nitro generalized the flat `Uint[]` into per-asset allocations with guarantees.
The "closed game" distinction is not a model theorem.

### S1-D04 — Move
Quote (674-683): `struct Move contains state: State signature: Signature`;
`Move::channelId: m 7→ m.state.channelId`; `Move::signer: m 7→
m.signature.signer`; "From the signature you can deduce the signer".
Property: a move is a state plus a signature. The signer is read from the
signature.
Coverage: COVERED in spirit. Model uses `AuthorizedBy`/`SignedByAll` over
`s.signedBy`, which records which participant signed which digest
(`Safety.lean:36-46`). The model tracks signing as recorded digests, not as a
`Move` struct carrying one signature.

### S1-D05 — State types and shared attributes; mover
Quote (699-715):
```
type State := PreFundsetupState | PostFundsetupState | GameState | ConcludeState
struct *State contains channel, turnNum, resolution
State::mover: s 7→ s.channel.participants[s.turnNum%s.nParticipants]
```
Property: every state carries channel, turnNum, resolution. The mover is
`participants[turnNum mod nParticipants]`.
Coverage: PARTIAL. Model splits state into `FixedPart` and `VariablePart`
(Nitro form). It has `turnNum` on `VariablePart`. The `mover =
participants[turnNum mod n]` rule is NOT a model theorem; the model requires
`SignedByAll` (all participants sign the candidate), not a single mover.

### S1-D06 — Mover determined by turnNum; cyclical turns
Quote (724-727): "the mover is fully determined by the turn number ... players
must take turns in a cyclical order."
Property: turn order is fixed and cyclical; turnNum selects the mover.
Coverage: CONTRADICTED by the model's signing choice. Model requires
`SignedByAll` for `conclude` and authorized `challenge`
(`Safety.lean:38-46,117-121`), i.e. all participants sign, not one mover per
turn. Nitro replaced per-turn single-signer with all-sign candidates. Record as
a ForceMove-only rule that Nitro supersedes.

### S1-D07 — GameState
Quote (932-938): `struct GameState contains channel, turnNum, resolution,
gameAttributes: Byte[]`; `stateType 7→ Game`.
Coverage: PARTIAL. Model `VariablePart` carries `appData` (the Nitro name for
gameAttributes). COVERED in form.

### S1-D08 — ConcludeState
Quote (1006-1011): conclude states carry only the universal properties;
`stateType 7→ Conclude`.
Coverage: PARTIAL. Model uses `isFinal`-style final states / `conclude`, not a
distinct `ConcludeState` type. Nitro replaced the conclude state-type with an
`isFinal` flag.

### S1-D09 — PreFundsetupState
Quote (1071-1078): `struct PreFundsetupState contains channel, turnNum,
position, stateCount, resolution`; `stateType 7→ PREFUNDSETUP`.
Coverage: MISSING. The model has no PreFundSetup state type and no `stateCount`.
Nitro removed the n-round prefund handshake; funding is a `deposit` call keyed
by channel id. Record as superseded.

### S1-D10 — PostFundsetupState
Quote (1138-1145): same attributes and transition rules as PreFundSetup.
(The struct is mis-labelled `PreFundsetupState` in the text at line 1138; the
section heading 3.2.5 and prose confirm it is PostFundSetup.)
Coverage: MISSING. Same as S1-D09; superseded by Nitro `deposit`.

### S1-D11 — Modes of play
Quote (1230-1270, Figure 5): PreFundSetup, Collaborative, Challenge, Concluded,
Terminated; transitions force-move, respond, expire, conclude/conclusion proof.
"the game enters the Challenge mode when a player triggers the force-move";
"If no response is received within the timeout period, the game transitions
into the Terminated mode."
Coverage: PARTIAL. Model modes are `open`/`challenge`/`finalized`
(`Channel.lean:182-222`). Mapping: PreFundSetup+Collaborative→open;
Challenge→challenge; Concluded+Terminated→finalized. Model `mode` is a function
of `finalizesAt` and `now`: `open` if `finalizesAt=0`; `finalized` if
`finalizesAt ≤ now`; else `challenge`.

### S1-D12 — Challenge struct and predicates
Quote (1292-1297):
```
struct Challenge contains endState: State endTime: Uint
Challenge::inProgress: c 7→ c.endTime > now()
Challenge::terminated: c 7→ c.endTime > 0 && !c.inProgress
```
Property: a challenge stores the end state and an end time. It is in progress
while `endTime > now`. It is terminated when `endTime > 0` and not in progress.
Coverage: COVERED. Model `ChannelStatus.mode` with `finalizesAt` encodes the
same two predicates: `finalizesAt=0` means no challenge; `finalizesAt ≤ now`
means terminated/finalized; else in progress (`Channel.lean:219-222`).

### S1-D13 — Noncollaborative interface (the adjudicator method set)
Quote (1285-1307, Interface 3): `forceMove(moves)`, `refute(m)`,
`respondWithMove(m)`, `alternativeRespondWithMove(moves)`,
`withdraw(channelId, player)`; plus implementation-dependent `setChallenge`,
`getChallenge`, `getCurrentOutcome`, `cancelChallenge`, `challengeInProgress`,
`isTerminated`.
Coverage: PARTIAL / superseded. Model operations are `deposit`, `challenge`,
`checkpoint`, `conclude`, `transfer`, `reclaim` (`Operations.lean`, from the
Step cases in `Safety.lean`). Mapping and disagreements are in the disagreement
list at the end.

---

## Class (a) — on-chain adjudicator safety

### S1-A01 — validMove
Quote (744-751):
```
function validMove(fromMove, toMove)
   Require(validTransition(fromMove.state, toMove.state))
   Require(toMove.state.mover == toMove.signer)
```
Property: a move is valid only if it is a valid transition AND the signer equals
the mover that the turnNum selects.
Caller: anyone who submits moves on-chain; can influence only which signed moves
they present, not the mover rule.
Coverage: PARTIAL. Model gates by `AuthorizedBy`/`SignedByAll`, not by
`mover == signer`. The transition check is abstracted into `appRel`. The
mover==signer half is CONTRADICTED by the all-sign model (see S1-D06).

### S1-A02 — validTransition universal rules
Quote (761-778, 905-906): `Require(toState.channelId == fromState.channelId)`;
`Require(toState.turnNum == fromState.turnNum + 1)`; "no details of the channel
can change ... channelType, channelId, or participants must all remain the
same"; "the turn number must increment".
Property: a valid transition keeps the channel fixed and increments turnNum by
exactly 1.
Coverage: PARTIAL. Model keeps channel fixed by keying on `channelId fp`. The
"turnNum increments by exactly 1" rule is ABSTRACTED into `appRel` (a universal
parameter, `Safety.lean:27-29`); no theorem states `turnNum == +1`. Nitro uses
strictly-increasing / app-defined progression, not a fixed +1.

### S1-A03 — validTransition dispatch
Quote (903-917): dispatch on `s1.stateType` to
`validTransitionFromPrefundsetup`/`...Postfundsetup`/`...Game`/`...Conclude`.
Coverage: MISSING / superseded. The model abstracts app logic into `appRel` and
has no per-stateType dispatch. Nitro folds the phases.

### S1-A04 — Game transitions
Quote (971-980):
```
if s2.stateType == GAME then Require(GameLibrary.validTransition(s1, s2))
else Require(s2.stateType == CONCLUDE); Require(s2.resolution == s1.resolution)
```
Property: GAME→GAME needs the app's `validTransition`. GAME→CONCLUDE needs the
resolution unchanged.
Coverage: PARTIAL. GAME→GAME maps to `appRel`. GAME→CONCLUDE "outcome unchanged"
maps to `conclude_moves_no_value` (`Safety.lean:2056`), which states conclude
moves no value. COVERED for the conclude half.

### S1-A05 — Conclude transitions
Quote (1017-1022): `validTransitionFromConclude`: `Require(s2.stateType ==
CONCLUDE)`; `Require(s2.resolution == s1.resolution)`.
Property: from CONCLUDE only CONCLUDE, with resolution unchanged.
Coverage: COVERED in part by `conclude_moves_no_value` (conclude changes no
value). The "only CONCLUDE after CONCLUDE" sequencing is not separately stated.

### S1-A06 — PreFundSetup transitions
Quote (1092-1125): resolution must match; CONCLUDE allowed; at
`stateCount == nParticipants-1` must move to PostFundSetup with `stateCount==0`
and gameAttributes preserved; else next PreFundSetup with `stateCount+1`.
Coverage: MISSING / superseded (see S1-D09). The model has no prefund round.

### S1-A07 — PostFundSetup transitions
Quote (1150-1184): CONCLUDE (resolution match); at `stateCount==nParticipants-1`
move to GAME with `GameLibrary.validTransition`; else next PostFundSetup with
`stateCount+1`, resolution and gameAttributes matched.
Coverage: MISSING / superseded (see S1-D10).

### S1-A08 — Backing out (PRE/POSTFUNDSETUP → CONCLUDE)
Quote (1187-1208): "they always have the option to transition to a conclude
state before the GAME phase begins"; "If a player transitions from either a
PREFUNDSETUP or POSTFUNDSETUP state, s, to a CONCLUDE state, s0, we require the
resolution to match."
Property: a player can conclude before GAME; the resolution must match on that
transition.
Coverage: MISSING as stated. The resolution-match half maps to
`conclude_moves_no_value`. The pre-GAME backing-out option is superseded by
Nitro.

### S1-A09 — validConclusionProof
Quote (1030-1058):
```
Require(firstMove.state.stateType == CONCLUDE)
n := firstMove.state.nParticipants
Require(n == moves.length)
for k in 0...n-2 do Require(validMove(moves[k], moves[k+1]))
```
"A (valid) conclusion proof is a sequence of n valid, signed CONCLUDE states."
Property: a conclusion proof is n consecutive valid signed CONCLUDE moves, one
per participant, starting from a CONCLUDE state.
Coverage: PARTIAL. Model `conclude` requires `SignedByAll` of the candidate
(`Safety.lean:217-227`), i.e. all participants signed one final candidate. Nitro
replaced the n-state sequence with one candidate signed by all. Same intent
(every participant agrees the game is over), different shape.

### S1-A10 — forceMove
Quote (1340-1354):
```
Require(n == firstMove.state.nParticipants)
Require(!Adjudicator.challengeInProgress(channelId))
Require(!Adjudicator.isTerminated(channelId))
for k in 0...n-2 do Require(validMove(moves[k], moves[k+1]))
c := moves[n-1]
Adjudicator.setChallenge(c, now() + defaultExpirationTime)
```
Property: a force-move needs a valid run of n moves, no challenge in progress,
and a non-terminated channel. It sets a challenge on the last state with
`endTime = now + defaultExpirationTime`.
Caller: the challenger (any participant). Influences the move sequence and
timing.
Coverage: PARTIAL. Maps to the model `.challenge` step and
`challenge_can_happen` (`Safety.lean:1383`). The preconditions
"not already in challenge" and "not terminated" are not stated as a standalone
safety theorem; they live inside `Step`. Nitro keeps a challenge-like operation
but validates a candidate signed by all, not an n-move run.

### S1-A11 — refute  [SUPERSEDED by Nitro `checkpoint`]
Quote (1399-1408):
```
Require(Adjudicator.challengeInProgress(channelId))
Require(refutation.stateNonce > challengeState.stateNonce)
Require(refutation.signedBy(challengeState.mover))
Adjudicator.cancelChallenge(channelId)
```
Property: a later state signed by the challenge's mover cancels the challenge.
Coverage: MISSING / superseded. The model has `checkpoint`, not `refute`. Nitro
`checkpoint` records a later agreed state and clears the challenge. The model's
`checkpoint` case appears in `Safety.lean:206-216`.

### S1-A12 — respondWithMove  [SUPERSEDED by Nitro `respond`]
Quote (1431-1440):
```
Require(Adjudicator.challengeInProgress(channelId))
Require(response.signedBy(response.mover))
Require(validTransition(challengeState, response.state))
Adjudicator.cancelChallenge(channelId)
```
Property: a valid signed transition from the challenge state cancels the
challenge.
Coverage: MISSING / superseded. The model's Step set (deposit, challenge,
checkpoint, conclude, transfer, reclaim) has no explicit `respond`; a later
`challenge`/`checkpoint` plays the equivalent role under Nitro.

### S1-A13 — alternativeRespondWithMove  [SUPERSEDED]
Quote (1464-1478): n moves, `moves[n-2].turnNum == challengeState.turnNum`,
all valid, cancel challenge. "an unavoidable consequence of the rule that a
player has the freedom to choose their preferred move" (1481-1483).
Coverage: MISSING / superseded. Nitro removed alternative moves (single
signed candidate model).

### S1-A14 — conclude
Quote (1495-1507):
```
endState := proof.moves[0].state
Require(!Adjudicator.isTerminated(endState.channelId))
Require(validConclusionProof(proof))
Adjudicator.setChallenge(endState, now())
```
"a game concluded with a conclusion proof behaves exactly the same as one that
was terminated by an expired challenge."
Property: a valid conclusion proof on a non-terminated channel finalizes it
immediately (`endTime = now`). Concluded then equals terminated.
Caller: any participant.
Coverage: COVERED. Model `conclude` fires with `SignedByAll`
(`conclude_can_happen`, `Safety.lean:1498`); it sets `finalizesAt = now` and
clears `stateHash` (`Safety.lean:121-124`). The "concluded == finalized" mapping
is the model mode design.

### S1-A15 — withdraw requires termination (protocol)
Quote (1716-1728): "Once the game has ended ... a game's adjudicator should
release funds for withdrawal"; `Require(Adjudicator.isTerminated(channelId))`.
Property: funds can be released only after the channel is terminated/finalized.
Caller: any party withdrawing.
Coverage: COVERED. Model `transfer` requires the channel in `finalized` mode;
`transfer_pays_only_the_finalized_outcome` (`Safety.lean:362`) and the transfer
proof's `hmode` precondition enforce this.

### S1-A16 — simple withdraw: priority-order payout
Quote (1837-1853):
```
Require(Adjudicator.isTerminated(channelId))
Require( 0 <= k && k < n )
r := s.resolution
pending := sum(r[i] - w[i]; i := 0; i < k; i++)
owed := r[k] - w[k]
Require(Adjudicator.balance > pending)
amount := min(owed - w[k], Adjudicator.balance - pending)
Adjudicator.withdrawnAmounts[k] += amount
Adjudicator.transfer(participant, amount)
```
Property: participant `k` is paid only from the balance left after the full
amounts owed to all earlier participants `i < k` are reserved (`pending`). The
payout is `min(what k is owed, balance after pending)`. This is the
priority/afford rule: earlier participants in the outcome list are paid first;
an underfunded channel pays the high-priority participants in full and the
low-priority ones partially or not at all. `withdrawnAmounts` tracks partial
withdrawals so no one is double-paid.
Caller: any party; influences only `k`.
Coverage: PARTIAL — this is a key property and is only partly captured. Model
`transferEffects`/`afford` (`Outcome.lean`) is the Armada form of this priority
reserve. `transfer_pays_only_the_finalized_outcome` states a payout pays no more
than a recorded allocation (`Safety.lean:362`). `full_payout_leaves_no_remainder`
(`Safety.lean:2340`) and `zero_payout_mints_nothing` (`Safety.lean:2262`) cover
edge cases. But NO theorem states the priority-order/afford rule itself: that
participant k receives `min(owed, held − sum of earlier owed)`. This is the same
gap the operator flagged from Nitro §5.1-5.2 ("A can afford x for B"). MISSING as
an explicit priority/afford statement.

### S1-A17 — simple adjudicator channelId gating  [SUPERSEDED: multi-channel]
Quote (1636-1638, 1754-1790): "Operations fail for all states whose channelId
differs from that in the adjudicator"; every simple-adjudicator function has
`Require(channelId == Adjudicator.channelId)`.
Property: the simple adjudicator serves exactly one hard-coded channel; calls
for other channel ids fail.
Coverage: N/A / superseded. The model is multi-channel: `s.channels` is a map
`cid → ChannelStatus`. Nitro generalized to one contract for many channels.

### S1-A18 — cancelChallenge
Quote (1765-1770): `cancelChallenge` sets `c.endTime := 0`.
Property: cancelling a challenge resets the end time to 0 (back to no-challenge).
Coverage: COVERED in form. Model cancellation is a later `challenge`/`checkpoint`
rewriting `finalizesAt`. `finalizesAt = 0` is the model's "no challenge" state
(`Channel.lean:220`).

### S1-A19 — GameLibrary.validTransition is pure
Quote (963-966): "we enforce that the GameLibrary.validTransition function must
be pure – it can only depend on the properties of states passed in, and cannot
read from or write to the blockchain."
Property: app transition logic is a pure function of the two states.
Coverage: COVERED by construction. Model `appRel : AdjudicatorState → FixedPart
→ List VariablePart → VariablePart → Prop` is a pure relation parameter
(`Safety.lean:27-29`). Note: the model's `appRel` can read `AdjudicatorState`,
which is broader than "only the two states"; this is a weakening the model makes.

### S1-A20 — invalid move equals no move
Quote (685-686, 150-152): "making an invalid move is equivalent to not making
any move at all"; "sending invalid state updates ... are viewed by the protocol
as equivalent to not sending any update."
Property: an invalid update has no effect on the on-chain game state.
Coverage: COVERED by construction. Model Step only fires when its preconditions
hold; an invalid call produces no step, so state is unchanged.

### S1-A21 — force-move extensions (default protocol)
Quote (1543-1548): "At the time of response, store the turnNum of the response
state in the adjudicator. Only allow new force-moves if their response will
increase this stored turnNum."; "Add a certificate argument to the force-move
operation, to identify the challenger and prove that they intended to play that
force-move. Require that the challenger is one of the participants."; "These
extensions should probably be part of the default protocol".
Property: (a) a new challenge must raise a stored response turnNum (monotone
progress, anti-replay). (b) the challenger must be a named participant, proven
by a certificate.
Caller: the challenger.
Coverage: PARTIAL. (a) The model records `stateHash`/turn progress and the
monotone `signedBy`; the Nitro `turnNumRecord`/`checkpoint` carries this. No
model theorem states "a new challenge must raise the stored turnNum." (b) The
model's `challenge` carries a `challenger` and uses `AuthorizedBy`
(`Safety.lean:134-137`); "challenger must be a participant" is partly captured by
authorization. MISSING as an explicit monotone-turnNum theorem.

---

## Class (b) — liveness / unbeatable strategy

### S1-B01 — force-move advances or terminates; fair recovery
Quote (183-188): "If an opponent is inactive, a participant can issue a
force-move operation on-chain, which will wait a predetermined length of time
for the opponent to respond and otherwise terminates the game so that the assets
can be withdrawn." (1314-1318): "the force-move operation will either result in
an advance to the state of the game or in the termination of the game. In the
latter case, each participant is allowed to reclaim the 'fair' proportion of
their funds."
Property: an inactive opponent cannot lock funds. The honest party can always
force the game to advance or terminate, then recover the outcome.
Caller: the honest participant; the attacker can only delay by the timeout.
Coverage: COVERED (as the model's `sorry` theorems). This is
`unbeatable_pays` and `unbeatable_pays_after_timeout` (`Safety.lean:444,500`),
the two unbeatable-strategy statements. They say: conclude-then-transfer, or
tick-to-timeout-then-transfer, succeeds after any interleaving by other actors.

### S1-B02 — each response raises turnNum; game always advances
Quote (1377-1381): "each of these responses lead to an increase in the turnNum
from the challenge state stored in the adjudicator. This ensures that the
force-move will always advance the game from the blockchain's perspective. In
combination with extensions discussed in Section 4.5, this prevents the game
from stalling indefinitely."
Property: every valid response strictly advances turnNum, so the game cannot
stall forever.
Coverage: MISSING. No model theorem states that a response/checkpoint strictly
increases the recorded turn, nor a no-infinite-stall property. Related to S1-A21.

### S1-B03 — timeout resolution decides the split
Quote (449-453): "a deadline is set for Alice to respond. If this deadline
expires before Alice responds, the resolution stored on the challenge state s
will decide how the funds should be split." (1260-1265): "If no response is
received within the timeout period, the game transitions into the Terminated
mode. In this mode, no further play is possible, but the players can reclaim
their funds according to the outcome of the game."
Property: on timeout, the outcome stored on the challenge state is the final
split; players reclaim per that outcome.
Caller: challenger sets the stored outcome; timeout is the chain's.
Coverage: COVERED. `challenge_timeout_can_happen` (`Safety.lean:1434`) shows a
challenge times out to finalized. `transfer_pays_only_the_finalized_outcome`
(`Safety.lean:362`) pays that outcome.
`finalized_outcome_was_concluded_or_challenged` (`Safety.lean:125`) proves a
finalized outcome traces to a challenge (or conclude), not to a bare assertion.

### S1-B04 — valid response before timeout returns to Collaborative
Quote (1266-1269): "In order to prevent the game from being terminated, some
party must respond to the force-move on-chain before the game timeout passes. If
a valid response is received, the game returns to the Collaborative mode".
Property: a valid, timely response cancels the challenge and reopens the channel.
Coverage: PARTIAL. Model `checkpoint`/later `challenge` can clear a running
challenge (`finalizesAt` back toward open). No standalone "a timely valid
response reopens the channel" theorem.

### S1-B05 — failing to notice a force-move cannot hurt you
Quote (1511-1529): "as long as Alice notices Bob's challenge within the timeout
period, she cannot hurt her position by playing on as though still in
Collaborative mode. ... When the force-move is cancelled, all the moves made in
Collaborative still stand."
Property: playing on off-chain during a challenge is safe if you respond in
time; no off-chain progress is lost.
Coverage: MISSING. No model theorem states this no-regret property for a late
observer. This is close to Nitro §4.3 "transactions can be front-run" that the
model's section B cites but does not isolate as its own statement.

### S1-B06 — PreFundSetup round arms each player to recover funds
Quote (862-871): "By the end of the PREFUNDSETUP phase, every player should hold
a set of n signed PREFUNDSETUP states ... This is exactly what they need to
launch a challenge on-chain to recover their funds if one of their opponents
stall".
Property: completing the prefund round gives each player the signed states
needed to force-move and recover deposits.
Coverage: MISSING / superseded. No prefund round in the model (see S1-D09). The
general "you can always recover" is S1-B01.

### S1-B07 — fund in order; earlier depositor recovers even if later defaults
Quote (1197-1208): "It is only safe to [fund] in order – this ensures that they
are paid out their deposits in case of an aborted game." (1812-1821): "The
resolution at this point would attempt to give aResolution to Alice and
bResolution to Bob, paying out Alice first, allowing her to recover her funds
even if Bob does not deposit."
Property: with priority-order payout, the earlier depositor is made whole before
the later one; so depositing in order is safe even if a later party never funds.
Caller: depositors.
Coverage: PARTIAL. This is the depositor-side consequence of the priority/afford
rule (S1-A16). The model's `transferEffects`/`afford` encode priority, but no
theorem states "the earlier-priority depositor is fully paid before the later
one when the pool is short." MISSING as a stated safety consequence.

---

## Class (c) — can-happen (operation must be possible)

### S1-C01 — deposit possible
Quote (874-876): "the funding step involves each player making an on-chain
transaction into an on-chain adjudicator contract, in a pre-defined order".
(1797-1808) lists the concrete deposit steps.
Coverage: COVERED. `deposit_can_happen` (`Safety.lean:815`),
`topup_can_happen` (`Safety.lean:968`).

### S1-C02 — forceMove / challenge possible
Quote (1340-1354): forceMove succeeds when no challenge is in progress and the
channel is not terminated, given a valid move run.
Coverage: COVERED. `challenge_can_happen` (`Safety.lean:1383`).

### S1-C03 — withdraw / payout possible after termination
Quote (1716-1728, 1837-1853): withdraw releases funds once terminated.
Coverage: COVERED. `payout_can_happen` (`Safety.lean:1553`),
`reclaim_can_happen` (`Safety.lean:1876`).

(ForceMove does not frame operations as explicit "can-happen" theorems. These
three are the operations the text says must be possible.)

---

## Class (e) — off-chain / app-level protocol properties (listed, not dropped)

### S1-E01 — payment-game rule: never take, can give; sum conserved
Quote (536-538): "you should never unilaterally be able to take funds from your
opponent, but you should be able to unilaterally give funds to your opponent."
(546-551, Specification 1):
```
Require(sum(s1.resolution) == sum(s2.resolution))
i := s1.indexOfMover
Require(s1.resolution[i] >= s2.resolution[i])
```
Property: a payment transition conserves the total and never increases the
mover's own allocation (so a mover can only give, not take).
Class: (e) app-library rule, not the generic adjudicator. Mark: app-level.
Coverage: CONTRADICTED-adjacent / out of scope. The model's generic adjudicator
does not fix a payment app; app logic is `appRel`. Value conservation of the
POOL (not of one app's resolution) is `notedValue_conserved`
(`Safety.lean:555`). Note: this is a per-app invariant; the model keeps it
abstract.

### S1-E02 — commit-reveal and default-winner incentive
Quote (339-346): the acceptor updates the resolution "as though he had won ...
Without the added incentive of Bob being the default winner ... it could be in
Alice's interest to end the game." (458-461): "the only fair resolution in the
RPS.Accept state is to award all the stake to the non-revealer".
Property: app design must make stalling unprofitable (default-winner rule).
Class: (e) app design / game-theoretic. Mark: off-chain app design.
Coverage: out of scope for the adjudicator model.

### S1-E03 — alternative moves give the next player the choice
Quote (781-797): "if there exist multiple valid moves ... the player gives the
next player the right to choose the move they want to progress from. A player
therefore theoretically does not gain anything by making multiple moves".
Property: signing several same-turnNum moves only hands choice to the opponent.
Class: (e) protocol behavior. Mark.
Coverage: MISSING / superseded. Nitro dropped alternative moves (see S1-A13).

### S1-E04 — rational backing out before GAME
Quote (1187-1196): a player who dislikes the initial conditions concludes rather
than proceeds; before funding they may simply ignore the game.
Class: (e) participant behavior. Mark.
Coverage: out of scope; prefund removed in Nitro.

### S1-E05 — griefing cannot be fully eliminated; deposit mitigates refute case
Quote (1566-1575): three situations are "indistinguishable from the perspective
of the blockchain", so some griefing is unavoidable. (1576-1586): a forfeitable
deposit at force-move time lets you penalize the one clear bad-faith case (a
refuted force-move).
Class: (e)/(f) economic limit plus an optional safety extension. Mark.
Coverage: MISSING. No model theorem about griefing bounds or forfeitable
deposits. Nitro handles anti-replay via `turnNumRecord`, not a deposit.

---

## Class (f) — assumptions the source makes

### S1-F01 — cryptographic hash
Quote (606-611): `h` "is a cryptographic hash function, such as the keccak256
hash used by ethereum". The channel id relies on it being collision-resistant.
Coverage: present. Model uses `HashAssumptions` and `outcomeHashOf_inj` (hash
injectivity) in several theorems.

### S1-F02 — signatures bind the signer
Quote (683): "From the signature you can deduce the signer".
Coverage: present. Model `SignatureAssumptions` is an explicit argument to every
call-sequence theorem (`Safety.lean:22-26`).

### S1-F03 — chain clock and timeout
Quote (1261, 1353): a game timeout is set on force-move; `setChallenge(c, now()
+ defaultExpirationTime)`; `inProgress` and `terminated` compare to `now()`.
Coverage: present. Model uses `s.now` and `finalizesAt` (`Channel.lean:219-222`),
and `challenge_timeout_can_happen` advances time.

### S1-F04 — adjudicator holds funds in escrow
Quote (1278-1281): "The purpose of the adjudicator is to both hold the players'
funds in escrow throughout the game and to manage the dispute process."
Coverage: present in Armada form. Model holds funds as pool notes owned by the
adjudicator (`Safety.lean:54-63`, `value_exists_only_as_notes`
`Safety.lean:2091`).

### S1-F05 — funding is decoupled from game logic
Quote (1609-1628): "Funding happens externally to the channel and there is no
way of telling from the properties of the channel how (or even whether) the game
is funded." Non-funded games still run (5.1).
Property: the adjudicator does not require a channel to be funded for the game
logic to be valid.
Coverage: present by construction. Model separates `deposit`/holdings from
challenge/conclude logic.

---

## Counts per class

- (a) on-chain adjudicator safety: 21 (S1-A01 … S1-A21)
- (b) liveness / unbeatable strategy: 7 (S1-B01 … S1-B07)
- (c) can-happen: 3 (S1-C01 … S1-C03)
- (d) definitions: 13 (S1-D01 … S1-D13)
- (e) off-chain / app-level: 5 (S1-E01 … S1-E05)
- (f) assumptions: 5 (S1-F01 … S1-F05)
- Total: 54 properties.

## Coverage summary

- COVERED: S1-A14, S1-A15, S1-A18, S1-A19, S1-A20, S1-B01, S1-B03, S1-C01,
  S1-C02, S1-C03, S1-D12, and (in Armada form) the assumptions F01-F05.
- PARTIAL: S1-D01, S1-D03, S1-D04, S1-D05, S1-D07, S1-A01, S1-A02, S1-A04,
  S1-A05, S1-A09, S1-A10, S1-A16, S1-A21, S1-B04, S1-B07, S1-D11.
- MISSING: S1-D02, S1-D08, S1-D09, S1-D10, S1-A03, S1-A06, S1-A07, S1-A08,
  S1-A11, S1-A12, S1-A13, S1-A17, S1-B02, S1-B05, S1-B06, S1-E03, S1-E04, S1-E05.
  Several MISSING items are superseded by Nitro (prefund rounds, refute,
  respondWithMove, alternativeRespondWithMove, single-channel gating); those are
  gaps only against ForceMove, not against the model's intent.
- CONTRADICTED: S1-D06 and S1-A01 (mover==signer / per-turn single signer) vs
  the model's `SignedByAll` all-sign requirement.

## Priority gaps the model should probably state (not superseded by Nitro)

1. S1-A16 / S1-B07 — the priority-order / afford payout rule: participant k is
   paid `min(owed, held − sum of earlier-priority owed)`. This is the ForceMove
   form of the same Nitro §5.1-5.2 payment property the operator flagged. The
   model has `transferEffects`/`afford` but no theorem stating the priority
   reserve or the "earlier depositor is made whole first" consequence.
2. S1-B02 / S1-A21 — monotone turn progress: a response/new challenge must
   strictly raise the recorded turn, so the game cannot stall forever and an old
   state cannot replay. The model tracks `stateHash`/`signedBy` but states no
   strict-progress theorem.
3. S1-B05 — no-regret for a late challenge observer: a timely valid response
   loses no off-chain progress.
4. S1-D02 — channel-id uniqueness / replay-safety consequence.

---

## Cross-source disagreements I noticed from this text

- Operation names and set. ForceMove (Interface 3, lines 1285-1290):
  `forceMove`, `refute`, `respondWithMove`, `alternativeRespondWithMove`,
  `withdraw`. Model (Nitro form): `challenge`, `checkpoint`, `conclude`,
  `transfer`, `reclaim`, `deposit`. Nitro merges ForceMove's `refute` +
  `respondWithMove` + `alternativeRespondWithMove` into `respond`/`checkpoint`,
  and replaces `withdraw`'s priority payout with `transfer`/`reclaim`. The
  operator's note names "Nitro 2019 `claim` versus SATP `reclaim`": ForceMove
  has neither; its payout verb is `withdraw` (line 1725). So the payout verb
  changed twice: ForceMove `withdraw` → Nitro-2019 `claim` → SATP/model
  `reclaim`/`transfer`.
- Signing model. ForceMove: one mover signs each move, chosen by
  `turnNum mod n` (lines 714, 748). Model/Nitro: the candidate must be signed by
  all participants (`SignedByAll`). This is a real semantic change, not a
  rename.
- Modes. ForceMove: five modes (line 1239, Figure 5). Model/Nitro: three
  (`open`/`challenge`/`finalized`, `Channel.lean:182-222`).
- Conclude representation. ForceMove: an n-state CONCLUDE sequence proof (lines
  1030-1058). Model/Nitro: one `isFinal` candidate signed by all. Same intent.
- channelId inputs. ForceMove hashes `(gameType, participants, channelNonce)`
  (line 606). Model/Nitro hashes `(participants, channelNonce, appDefinition,
  challengeDuration)` (`Channel.lean:66-67`). Nitro adds `challengeDuration`.
- Resolution/outcome. ForceMove stores `resolution` on every state (line 707).
  Model/Nitro separates the outcome and stores an `outcomeHash`, rewritten by
  `transfer`/`reclaim` (`Safety.lean:99-116`).

I did not read the other sources (SATP, Nitro 2019, docs). The disagreements
above are inferences from this text plus the model, marked where I state the
Nitro/model side.
