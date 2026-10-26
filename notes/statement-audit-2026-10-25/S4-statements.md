# S4 — docs.statechannels.org protocol tutorial — property extraction

Source tag: **S4**. Source: the saved HTML pages under
`~/code/git_puller/repos/armada-nitro/docs/references/nitro/sources/docs.statechannels.org/`.
These docs are generated from `statechannels/go-nitro` `docs/`. They name
the functions the contracts expose.

Model compared against: `~/code/git_puller/repos/armada-lean-adjudicator-v4`
at commit `5f5116a`, the 24 theorems in `ArmadaLean/Adjudicator/Safety.lean`.

## How I read the source

The `read` tool shows the local HTML raw, not rendered. I wrote a
stdlib HTML-to-text extractor (`.scratch/extract.py` in armada-laconic)
that keeps only the `<article>` element and drops `script`/`style`/`nav`/
`svg`/`form`. I read every page start to end. Line numbers below are
lines in that extracted text, not the raw HTML. The last extracted
content line per file:

| page | last extracted line |
|---|---|
| index.html | 19 |
| 0010-states-channels | 263 |
| 0020-execution-rules | 65 |
| 0030-outcomes | 112 |
| 0040-lifecycle-of-a-channel | 207 |
| 0060-funding-a-channel | 328 |
| 0070-finalizing-a-channel | 80 (raw article ends at HTML line 951) |
| 0080-defunding-a-channel | 175 |
| research-papers | 33 |

I confirmed the extractor dropped no section: a raw heading grep of the
finalizing page showed the same h1/h2/h3 set my text has, and the only
`respond` mention is prose with no function signature.

---

## A. Definitions the model must match (class d)

### S4-0010-01 — State = FixedPart + VariablePart
Quote (0010:17–35): "a state is broken up into fixed and variable
parts: `struct FixedPart { address[] participants; uint48 channelNonce;
address appDefinition; uint48 challengeDuration; }` ... `struct
VariablePart { Outcome.SingleAssetExit[] outcome; bytes appData; uint48
turnNum; bool isFinal; }`".
Property: the model's `FixedPart` has exactly `participants`,
`channelNonce`, `appDefinition`, `challengeDuration`; `VariablePart`
has exactly `outcome`, `appData`, `turnNum`, `isFinal`.
Class: d. Coverage: **COVERED** — `Channel.lean:49–61` defines both
structures with these fields and cites this quote.

### S4-0010-02 — channelId formula
Quote (0010:155–163): "`bytes32 channelId = keccak256(abi.encode(
fixedPart.participants, fixedPart.channelNonce, fixedPart.appDefinition,
fixedPart.challengeDuration));`".
Property: `channelId fp` is a hash over exactly those four fixed fields.
Class: d. Coverage: **COVERED** — `Channel.lean:66` `channelId` via
`channelIdOf` over the same four fields; `channelIdOf_inj` records that
no caller can collide two different fixed parts.

### S4-0010-03 — stateHash formula
Quote (0010:169–175): "`bytes32 stateHash = keccak256(abi.encode(
channelId, variablePart.appData, variablePart.outcome,
variablePart.turnNum, variablePart.isFinal));`".
Property: `stateHash fp vp` hashes channelId, appData, outcome, turnNum,
isFinal.
Class: d. Coverage: **COVERED** — `Channel.lean:71` `stateHash` via
`stateHashOf` over the same five inputs.

### S4-0010-04 — signedBy bitmask from signature recovery
Quote (0010:237–262): "The adjudicator smart contract will recover the
signer from each signature on a SignedVariablePart, and convert the
resulting list into a signedBy bitmask ... For example ... all signed
... signedBy = 0b111 = 7. If only participant with index 0 had signed,
... signedBy = 0b001 = 1."
Property: the adjudicator derives, per state, a per-participant signed
set by recovering each signature's signer.
Class: d (and f, an assumption on ecrecover). Coverage: **PARTIAL** —
the model tracks `s.signedBy p d` and `SignedByAll sig`/`AuthorizedBy
sig` (Safety.lean passim), and `SignatureAssumptions` is the explicit
recovery assumption, but the model does not model a numeric bitmask
`signedBy : Uint256`; it uses a predicate. Difference: form only
(predicate vs bitfield), same meaning.

### S4-0030-01 — Outcome = array of SingleAssetExit
Quote (0030:13–21): "An Outcome is an array of SingleAssetExits, each
specifying: an asset, optional assetMetadata, an ordered list of
allocations".
Property: `Outcome` is a list of `SingleAssetExit { asset,
(assetMetadata), allocations }`.
Class: d. Coverage: **COVERED** — `Outcome.lean:120` `SingleAssetExit`,
`:152` `abbrev Outcome := List SingleAssetExit`.

### S4-0030-02 — Allocation = destination, amount, allocationType, metadata
Quote (0030:24–34): "An allocation is a destination and an amount, an
allocationType identifier, optional metadata. The allocationType
identifier is usually set to 0, meaning 'simple'."
Property: `Allocation { destination, amount, allocationType, metadata }`;
`allocationType = simple` is the normal case.
Class: d. Coverage: **COVERED** — `Outcome.lean:66` `Allocation` with
those four fields; `AllocationType` inductive (`:38`) with `simple`.

### S4-0030-03 — Simple allocation is paid via `transfer`
Quote (0030:36–38): "Simple allocations do not have any metadata, and
allow for funds to be moved on-chain using the transfer method."
Property: a simple allocation's payout path is the `transfer` call.
Class: d. Coverage: **COVERED** — `Operations.lean` `.transfer` case
pays simple allocations; `transfer_pays_only_the_finalized_outcome`.

### S4-0030-04 — Guarantee allocation; moved via `reclaim`; metadata left/right
Quote (0030:40–51): "When allocationType is set to guarantee, funds
cannot be transferred in the usual way. Instead, they may be moved to
another channel on chain using the reclaim method ... `struct Guarantee
{ bytes32 left; bytes32 right; }` where left and right correspond to
channel participants."
Property: a guarantee allocation cannot be paid by `transfer`; it is
moved by `reclaim`; its metadata decodes to `(left, right)`.
Class: d (+ mechanics). Coverage: **PARTIAL** — `Outcome.lean:266`
`reallocateGuarantee` and the `.reclaim` Step case model the reclaim
path; `decodeGuarantee` decodes `(left, right)`. Difference: no theorem
states "a guarantee allocation is rejected by `transfer`"; only the
reclaim path is exercised (`reclaim_can_happen`). See S4-0080-05.

### S4-0030-05 — Destination is a channel id or a padded 20-byte address
Quote (0030:52–55): "A destination is 32 byte identifier which may
either denote a channel ID or a so-called 'external destination' (a 20
byte Ethereum address left-padded with zeros)."
Property: a destination is either a channel id or an external
destination = address in the low 20 bytes.
Class: d. Coverage: **COVERED** — `Outcome.lean:174` `isExternal
destination := ∃ a, a < 2^160 ∧ destination = Term.val a`.

### S4-0040-01 — Three channel modes
Quote (0040:96–99): "the adjudication status ... summarized in finite
states Open, Challenge and Finalized -- these are called channel modes."
Property: a channel is in exactly one of `Open`, `Challenge`,
`Finalized`.
Class: d. Coverage: **COVERED** — `Channel.lean:182` `ChannelMode` with
`open`/`challenge`/`finalized`.

### S4-0040-02 — `_mode` derivation from finalizesAt vs now
Quote (0040:129–147): "`if (finalizesAt == 0) return ChannelMode.Open;
else if (finalizesAt <= block.timestamp) return ChannelMode.Finalized;
else return ChannelMode.Challenge;` ... When finalizesAt is in the past,
... Finalized; ... in the future, ... Challenge. When it is ... 0, ...
Open."
Property: `mode(status, now)` = Open iff finalizesAt=0, Finalized iff
finalizesAt≤now, Challenge otherwise.
Class: d. Coverage: **COVERED** — `Channel.lean:219` `ChannelStatus.mode`
is this exact three-way test.

### S4-0040-03 — turnNumRecord meaning and conclude-resets-to-0
Quote (0040:149–152): "turNumRecord is the highest turn number that is
known to the chain to be supported by a full set of signatures. The
exception to this rule is that it is set to 0 when the channel is
concluded via a conclude call."
Property: `turnNumRecord` is the highest fully-supported turn number;
`conclude` sets it to 0.
Class: d (+ a, a reset rule). Coverage: **PARTIAL** — the conclude Step
case sets `turnNumRecord := 0` (`Operations.lean:267`), so the reset
is modeled, but no theorem STATES either "turnNumRecord is the highest
supported turn" or "conclude zeroes it". MISSING as a stated property.

### S4-0040-04 — Channel status storage (turnNumRecord, finalizesAt, fingerprint)
Quote (0040:157–188): "The fingerprint ... `uint160(uint256(keccak256(
abi.encode(stateHash, outcomeHash))))` ... Status data is stored ...
`mapping(bytes32 => bytes32) public statusOf;` ... most significant 48
bits to the turnNumRecord ... next 48 bits to finalizesAt ... next 160
bits to the fingerprint."
Property: per-channel on-chain status packs turnNumRecord, finalizesAt,
and a fingerprint of (stateHash, outcomeHash).
Class: d. Coverage: **PARTIAL (intentional divergence)** —
`Channel.lean:199` `ChannelStatus` stores `turnNumRecord`,
`finalizesAt`, `stateHash`, `outcomeHash` UNPACKED, deliberately
dropping the 160-bit fingerprint/slot packing. The model's own doc
(`Channel.lean:188–198`) argues the packing has no semantic content and
no caller influences it. Difference: representation only.

### S4-0040-05 — Timeout is a passive transition
Quote (0040:116–117): "The remaining transition timeout is a passive
transition which happens when a block is produced with a sufficiently
large timestamp."
Property: no caller action finalizes a challenged channel at timeout;
advancing `now` past `finalizesAt` alone does it.
Class: b/c. Coverage: **COVERED** — `challenge_timeout_can_happen`
(Safety.lean:1434): a `.tick` to `finalizesAt` reaches `.finalized`
mode with no further call.

### S4-0020-01 — ConsensusApp = unanimous countersignature
Quote (0020:58–60): "the consensus app ... in order for a state to be
supported, it must be unanimously countersigned -- that is, signed by
all of the channel participants."
Property: under the ConsensusApp rule, a state is supported iff every
participant signed it.
Class: d. Coverage: **COVERED** — the can-happen theorems use
`consensusRelation sig` for `appRel`; `SignedByAll` is unanimous signing.

### S4-0010-05 — isFinal is an instant-checkout escape hatch
Quote (0010:145–150): "This is a boolean flag which allows the channel
execution rules to be bypassed and for the channel to be finalized
'instantly' without waiting for the challenge-response window to lapse."
Also 0020:23: "an escape hatch for an 'instant checkout' ... which
bypasses the application rules altogether".
Property: `conclude` on a unanimously-signed `isFinal` state finalizes
immediately, bypassing `appRel`.
Class: b/c/d. Coverage: **COVERED** — `conclude_can_happen`
(Safety.lean:1498) finalizes with NO `appRel`/ConsensusApp machinery;
the `.conclude` Step case checks only `SignedByAll` and
`candidate.isFinal = true`.

---

## B. On-chain adjudicator safety (class a)

### S4-0060-01 — deposit must NOT target an external destination
Quote (0060:14–18): "There are some rules to obey when calling deposit.
Firstly, destination must NOT be an external destination. ... The first
rule prevents funds being escrowed against something other than a
channelId: funds may only be unlocked from channels, so you shouldn't
deposit into anything else."
Property: caller = depositor. A `deposit` whose `destination` is an
external destination must revert (no successful `Step`).
Class: a (revert condition). Coverage: **MISSING as a theorem** — the
`.deposit` Step case requires `¬ isExternal channelId`
(`Operations.lean:142`), so the model ENFORCES it, but no theorem states
"a deposit to an external destination has no successful Step". Only
`deposit_can_happen` (a positive witness) exists.

### S4-0060-02 — deposit requires on-chain holdings == expectedHeld
Quote (0060:14–18): "Secondly, the on-chain holdings for destination
must be equal to expectedHeld. ... The second rule prevents loss of
funds: since holdings are paid out in preferential order, depositing
before a counterparty has deposited implies that they can withdraw your
funds. The check is performed in the same transaction as the deposit,
making this safe even in the event of a chain re-org."
Property: caller = depositor. A `deposit` succeeds only if the channel's
currently held value equals the supplied `expectedHeld`; otherwise it
reverts.
Class: a (revert condition). Coverage: **MISSING as a theorem** — the
`.deposit` Step case requires `heldValue s asset channelId = expectedHeld`
(`Operations.lean:141`), so it is enforced, but no theorem states the
revert. `topup_can_happen` only witnesses the additive (merge) behaviour
of two correct deposits.

### S4-0060-03 — deposit adds to, never replaces, the prior holding
Quote (0060:18): "since holdings are paid out in preferential order,
depositing before a counterparty has deposited implies that they can
withdraw your funds." (Mechanism that the expectedHeld check protects;
the additive effect is the model's `depositBinding` reading.)
Property: a second valid deposit increases held value by exactly the
deposited amount, rather than overwriting.
Class: a/c. Coverage: **COVERED** — `topup_can_happen`
(Safety.lean:968): two deposits give `heldValue s' = heldValue mid + amt2`,
the fresh note MERGED with the prior holding.

### S4-0070-01 — conclude needs a full finalization proof
Quote (0070:11): "Once a full set of n such signatures exists (this set
is known as a finalization proof) the channel is said to be 'closed' or
'finalized'." (0070:19–23): "anyone in possession of a finalization
proof may use it to finalize the outcome on-chain ... conclude ...
without having to wait for a challenge to expire."
Property: caller = anyone holding n isFinal signatures. `conclude`
succeeds only if all participants signed the final candidate; it then
sets the channel `Finalized`.
Class: a/c. Coverage: **COVERED** (safety side) — the `.conclude` Step
case requires `SignedByAll` and `candidate.isFinal = true`
(`Operations.lean:261–268`); `conclude_can_happen` witnesses success and
`.finalized` mode. Also fed into
`finalized_outcome_was_concluded_or_challenged` (S4-0070-04).

### S4-0070-02 — a challenge may be registered only by a participant
Quote (0070:52–56): "The challenger needs to sign this data:
`keccak256(abi.encode(challengeStateHash, 'forceMove'))` in order to
form challengerSig. This signals their intent to challenge this channel
with this particular state. This mechanism allows the challenge to be
authorized only by a channel participant."
Property: caller = challenger. A `challenge` succeeds only if the
challenger is a channel participant and authorized the challenge
message.
Class: a (authorization). Coverage: **MISSING as a theorem** — the
`.challenge` Step case requires `challenger ∈ fp.participants` and
`appRel`/`AuthorizedBy` (`Operations.lean:199–201`), so it is enforced,
but no theorem states "a non-participant cannot register a challenge".
`challenge_can_happen` is only a positive witness.

### S4-0070-03 — higher turn number supersedes lower (core protocol rule)
Quote (0020:11): "Each state has a version number, with greater version
numbers superceding lesser ones." (0020:21): "the rule that higher turn
numbers take precedence over lower ones." (0070:66): checkpoint
"establish a new and higher turnNumRecord"; (0070:76): "a challenge may
be 'cleared' by another more recent challenge."
Property: challenge/checkpoint/respond can only advance the channel to a
state with a turn number at least as high (strictly higher while
challenged / for checkpoint) than `turnNumRecord`.
Class: a (core protocol safety). Coverage: **MISSING as a theorem** —
the Step cases enforce it: `.challenge` requires
`candidate.turnNum ≥ turnNumRecord` (open) or `> turnNumRecord`
(challenge) (`Operations.lean:196–197`); `.checkpoint` requires
`candidate.turnNum > turnNumRecord` (`:230`). No theorem STATES that
`turnNumRecord` is monotone, or that a lower-turn state cannot supersede
a higher one.

### S4-0070-04 — a finalized outcome came from conclude or a challenge
Quote (0070:11,19–35): finalization is reached either by a finalization
proof (`conclude`) or by a `challenge` whose outcome is registered and
then times out.
Property: for any reachable state, if a channel is finalized with a
given (stateHash, outcomeHash), then either all participants signed a
candidate deriving that outcome, or a `challenge` step registered that
exact (stateHash, outcomeHash).
Class: a. Coverage: **COVERED** —
`finalized_outcome_was_concluded_or_challenged` (Safety.lean:125) is
exactly this disjunction.

### S4-0080-01 — defunding can only happen after finalization
Quote (0080:5): "Defunding can only happen after a channel finalizes."
(0080:9): "If the channel is funded on-chain, it can be defunded using
the transfer method. It must first be finalized on chain, via either a
happy or sad path." (0080:88): "the channel must be finalized on chain
and then defunded using the on-chain transfer or reclaim methods."
Property: caller = anyone. `transfer` and `reclaim` succeed only on a
channel whose mode is `Finalized`.
Class: a. Coverage: **COVERED** — `.transfer` requires
`(s.channels cid).mode s.now = .finalized` (`Operations.lean:308`),
restated by `transfer_pays_only_the_finalized_outcome`
(`ChannelFinalizedWith`); `.reclaim` requires source mode `.finalized`
(`Operations.lean:361`).

### S4-0080-02 — transfer must match the stored stateHash and outcome
Quote (0080:9–10): "Because of the information is stored on chain, it is
necesary to supply both the stateHash and encoded outcome of the channel
when calling transfer." (0040:188): "The chain will then check that the
hydrated data matches the status that has been stored."
Property: caller = anyone. `transfer` succeeds only if the supplied
`(claimedStateHash, outcome)` hash to the channel's stored
`(stateHash, outcomeHash)`.
Class: a. Coverage: **COVERED** —
`transfer_pays_only_the_finalized_outcome` (Safety.lean:362): the
conclusion includes `ChannelFinalizedWith s cid claimedStateHash
(outcomeHashOf outcome)`, i.e. the supplied data must equal the stored
fingerprint (cites `MultiAssetHolder._requireMatchingFingerprint`).

### S4-0080-03 — transfer pays only destinations and amounts in the outcome
Quote (0080:9–10,45): "a list of indices to 'target' ... This magic
value (a zero length array) implies we want to pay out all of the
allocationItems." Priority-order payout: 0060:54 "the funds would be
paid out in priority order."
Property: caller = anyone. Every amount `transfer` pays now goes to a
destination already in the outcome and is no more than that
allocation's recorded amount; the selection and amounts follow the
priority-order `afford`.
Class: a. Coverage: **COVERED** —
`transfer_pays_only_the_finalized_outcome` conclusion: each `paid`
shares a destination with some `original` and `paid.amount ≤
original.amount`; `afford`/`transferEffects` encode priority order
(`Outcome.lean:185–231`).

---

## C. On-chain liveness / unbeatable strategy (class b)

### S4-0070-05 — conclude path: a signed finalization cannot be blocked (happy path liveness)
Quote (0070:9,19–22): "In the so-called 'happy' case, all participants
cooperate ... anyone in possession of a finalization proof may use it to
finalize the outcome on-chain ... without having to wait for a challenge
to expire. ... This enables assets to be released."
Property: caller = honest participant `h` holding a unanimously-signed
final state. No other caller's interleaved actions (any number, any
order, excluding `h`'s own identities and a `tick`) can stop `h` from
running conclude-then-transfer and obtaining a spendable note worth
their allocation.
Class: b. Coverage: **COVERED** — `unbeatable_pays` (Safety.lean:444)
states exactly this; its doc cites S4 lifecycle and Nitro-2019 §4.4.
(Proof body is `sorry`; the STATEMENT is what this stage grades.)

### S4-0070-06 — challenge/timeout path: a registered challenge pays out at timeout (sad path liveness)
Quote (0070:28–37): "When cooperation breaks down ... challenge ...
register a challenge state on chain ... with a finalization time set at
some delay after the transaction is processed. ... If no such
transaction is forthcoming, the challenge will time out, allowing the
outcome registered to be finalized. A finalized outcome can then be used
to extract funds." (0040:110): `Challenge --timeout--> Finalized'`.
Property: caller = honest participant `h` named in a registered
challenge. If no higher-turn state is presented before `finalizesAt`,
`h`'s `tick`-then-`transfer` strategy succeeds after any interleaving
(excluding `h`'s own identities and a `tick`).
Class: b. Coverage: **COVERED** — `unbeatable_pays_after_timeout`
(Safety.lean:500); doc cites S4 lifecycle `Challenge --timeout-->
Finalized'`. (Proof body `sorry`; statement graded.)

---

## D. No value is created or destroyed (class a, value conservation)

These are the armada-nitro note-pool reading of S4's "funds locked up"
language (0010:11: "A state channel controls funds which are locked up").

### S4-C-01 — a single call neither mints nor burns pool value
Class: a. Coverage: **COVERED** — `notedValue_conserved`
(Safety.lean:555) one-step; `notedValue_conserved_reachable`
(Safety.lean:692) over the reachable set. Source basis is the
armada-nitro rule, not S4 text directly; S4 only states funds are
locked and released, never minted on-chain.

### S4-C-02 — recorded collateral always points to a live note
Class: a. Coverage: **COVERED** — `holdings_point_to_live_notes`
(Safety.lean:713); `value_exists_only_as_notes` (Safety.lean:2091).
Again an Armada-encoding statement, consistent with S4's "funds locked
up" but sourced from the Armada layer.

### S4-0070-07 — conclude moves no value, only finalizes
Quote (0070:25): "an expired challenge is stored against the channelId.
(This is an implementation detail -- the important point is that the
chain shows that the channel has been finalized.)" conclude releases
nothing itself; defunding (transfer) does.
Property: `conclude` changes only channel status, not holdings or pool.
Class: a. Coverage: **COVERED** — `conclude_moves_no_value`
(Safety.lean:2056): `s'.pool = s.pool ∧ s'.holdings = s.holdings`.

### S4-0080-04 — transfer takes no fee (full held value accounted)
Property: the amount paid out plus the remainder equals the held value;
no fee is skimmed by the payout.
Class: a. Coverage: **COVERED** — `payout_takes_no_fee`
(Safety.lean:2029). Not an explicit S4 claim; S4 (index:9) says
"no per-payment fee". Consistent, Armada-sourced.

---

## E. Can-happen (operations must be possible) (class c)

### S4-0060-04 — a deposit can happen
Quote (0060:9): "The deposit method allows ETH or ERC20 tokens to be
escrowed against a channel." Coverage: **COVERED** —
`deposit_can_happen` (Safety.lean:815).

### S4-0060-05 — a top-up can happen
Quote (0060:53): channel may be "underfunded, exactly funded or
overfunded at different points in time" (multiple deposits). Coverage:
**COVERED** — `topup_can_happen` (Safety.lean:968).

### S4-0070-08 — a challenge can happen
Quote (0070:31,43–48). Coverage: **COVERED** — `challenge_can_happen`
(Safety.lean:1383) reaches `.challenge` mode.

### S4-0040-06 — a challenge can time out
Quote (0040:110,116). Coverage: **COVERED** —
`challenge_timeout_can_happen` (Safety.lean:1434).

### S4-0070-09 — a conclude can happen
Quote (0070:19–23). Coverage: **COVERED** — `conclude_can_happen`
(Safety.lean:1498).

### S4-0080-06 — a payout (transfer) can happen
Quote (0080:9,33–39). Coverage: **COVERED** — `payout_can_happen`
(Safety.lean:1553).

### S4-0080-05 — a reclaim can happen
Quote (0080:100): "The operation should mirror the on-chain reclaim
method." (0080:142–144): "If the child channel is funded with a
guarantee, funds may be transferred from the child channel to the parent
channel using reclaim." Coverage: **COVERED** — `reclaim_can_happen`
(Safety.lean:1876). Note: no SAFETY bound on what reclaim pays (the
guarantee-reallocation analogue of `transfer_pays`) exists — see the
MISSING list.

### S4-0070-10 — a checkpoint can happen and clears a challenge to Open
Quote (0070:64–68): "The checkpoint method allows anyone with a
supported off-chain state to establish a new and higher turnNumRecord on
chain, and leave the resulting channel in the 'Open' mode. It can be
used to clear a challenge." (0040:105,108): `Open --checkpoint--> Open`,
`Challenge --checkpoint--> Open`.
Property: caller = anyone with a supported higher-turn state. A
`checkpoint` succeeds, raises `turnNumRecord`, and leaves the channel
`Open` (clearing finalizesAt).
Class: c (+ a for the mode-transition guarantee). Coverage: **MISSING** —
the `.checkpoint` Call and Step case exist (`Operations.lean:227–236`,
setting `finalizesAt := 0`), but there is NO `checkpoint_can_happen`
theorem and no theorem that checkpoint clears a challenge to `Open`.
This is the only lifecycle transition with a defined call but no
can-happen witness.

---

## F. Mutation-test / encoding statements present in the model (class a/d)

These come from an armada-nitro brief, not S4, but relate to S4's
encoding rules (0010 overflow warning; "zero balance held as no note").
Listed for completeness, no new S4 property:
- `bulk_paths_never_succeed` (2072) — `transferAllAssets`/
  `concludeAndTransferAllAssets` always revert. S4 mentions the
  convenience method `concludeAndTransferAllAssets` (0080:77) as a
  gas-batching helper; the model's statement REMOVES it. See
  CONTRADICTED list.
- `holdings_never_zero_valued` (2163), `heldValue_zero_iff_no_collateral`
  (2195), `configuredAssets_invariant` (2232),
  `zero_payout_mints_nothing` (2262), `full_payout_leaves_no_remainder`
  (2340), `nullify_iff_commitment_was_some` (2404).

---

## Off-chain protocol properties and assumptions (classes e, f) — listed, not dropped

### S4-0010-06 (e) — participants list length 2..255, nonzero, self-included
Quote (0010:100): "you ... should check that it has length at least 2,
but no more than 255, and include a public key (account) that you
control. Each entry should be a nonzero ethereum address."
Off-chain client check. Not modeled (adjudicator-external).

### S4-0010-07 (e) — never reuse a channel id
Quote (0010:108): "You should never join a channel which re-uses a
channel ID." Off-chain. ChannelNonce (0010:104) exists to avoid replay.

### S4-0010-08 (f) — challengeDuration bounds
Quote (0010:122–124): "at least 1 block time (~15 seconds) and less than
2^48-1 seconds ... closed long before 2^48 - 1 - challengeDuration."
Assumption on parameter range. The model's `unbeatable_pays` needs only
`0 < fp.challengeDuration` (Safety.lean:454), a weaker hypothesis.

### S4-0010-09 (f) — turnNum overflow bounds
Quote (0010:142): "The turn number must not exceed 281,474,976,710,655
because then it will overflow on chain. It should not exceed
4,294,967,295 because it may then overflow off-chain."
Assumption. Not modeled — the model uses unbounded `Nat` for turnNum, so
on-chain uint48 overflow is out of scope.

### S4-0010-10 (e/f) — isFinal, once unpreventable, means stop executing
Quote (0010:150): "As soon as an isFinal=true state is enabled ... it is
not safe to continue executing the state channel. It should be finalized
immediately." Off-chain participant behaviour.

### S4-0020-02 (d/f) — app rule contract: stateIsSupported MUST revert/false on bad proof
Quote (0020:34–46): "Must revert or return false when invalid support
proof and a candidate are supplied. ... function stateIsSupported(
FixedPart, RecoveredVariablePart[] proof, RecoveredVariablePart
candidate) external view returns (bool, string)."
Property: the app-rule interface. Modeled as the `appRel` parameter
(`Operations.lean:21–25`); the model quantifies over any `appRel`.
Partial: the "must revert or return false" obligation on the app
contract is an assumption the model carries as `appRel`'s truth value,
not a theorem.

### S4-0020-03 (e) — VirtualPaymentApp is a mini-adjudicator with vouchers
Quote (0020:64). Off-chain auxiliary application.

### S4-0040-07 (e) — off-chain channel lifecycle states
Quote (0040:16–57): Proposed → ReadyToFund (all prefund sigs) → Funded
(all postfund sigs, turnNum=1) → Running (turnNum>1) → Finalized (all
isFinal sigs). Off-chain state derived from held signatures.

### S4-0040-08 (e/f) — Funded requires sufficient backing
Quote (0040:40–47): a participant signs postfund only when "sufficient
deposits are accrued on chain" OR "sufficient money is directly
allocated ... from an existing funded channel" OR "sufficient money is
guaranteed ... from an existing funded channel". Off-chain safety rule
for depositors.

### S4-0040-09 (e) — fund when ReadyToFund, defund when Finalized
Quote (0040:83). Off-chain.

### S4-0040-10 (e) — wait for a supported prefund before depositing
Quote (0040:36). Off-chain safety rule.

### S4-0060-06 (e/f) — outcome priority order governs safe deposit timing
Quote (0060:54): "If a channel were to finalize and be liquidated when
underfunded ... funds would be paid out in priority order ... depositing
safely requires an understanding of that priority order ... participants
should not deposit until those with higher priority have had their
deposits confirmed." Off-chain depositor rule; the on-chain priority
order itself is S4-0080-03 (afford), COVERED.

### S4-0060-07 (e) — fund a child channel from a ledger channel
Quote (0060:56–120): support a modified parent outcome that decrements
participants and appends an allocation targeting the child. Off-chain
construction.

### S4-0060-08 (e) — virtual funding with guarantees; metadata left=P_i, right=P_{i+1}
Quote (0060:174–294): each ledger channel adds a guarantee for V;
"the left metadata field in L_i's guarantee is set to participant P_i's
destination and the right metadata field is set to participant
P_(i+1)'s destination." Off-chain construction; the guarantee STRUCT is
S4-0030-04.

### S4-0060-09 (e/f) — ETH vs ERC20 deposit mechanics
Quote (0060:20,50): ETH deposit sets asset = zero address and sends
value; ERC20 requires prior `approve`. Chain-mechanics assumption; the
Armada model holds value as pool notes, not raw ETH/ERC20 (departures
noted in `Operations.lean`).

### S4-0070-11 (e) — respond clears a challenge; cache events to respond
Quote (0070:35,78–80): "cleared by a timely and well-formed respond or
checkpoint transaction"; (0070:950-951 prose) "to respond, the challenge
state itself must be (re)submitted ... Clients should ... cache
information from Events ... in order to be able to respond." The docs
name `respond` but give NO function signature. The model has no
`respond` Call (only `checkpoint`). Off-chain for the caching part; the
on-chain `respond` entry point is under-specified in S4 and absent from
the model — noted, not a derivable statement from this source.

### S4-0070-12 (e/f) — clients must cache adjudicator events (hashed storage)
Quote (0070:78–80; 0040:188–193): only hashes are stored on chain; the
pre-image must be tracked off-chain from events. Assumption on client
infrastructure.

### S4-0080-07 (e) — cooperative off-chain defunding mirrors reclaim
Quote (0080:90–100): remove the child allocation, increment parent
participants' external destinations; for a guarantee, award per the
guarantee metadata mapping, "mirror the on-chain reclaim method."
Off-chain construction.

### S4-0080-08 (e) — transfer-in-then-out and reclaim-then-out defunding flows
Quote (0080:102–144): if cooperation fails, finalize both channels, then
`transfer` into the child (simple funding) or `reclaim` from the child
(guarantee funding), then transfer out. Off-chain orchestration of the
on-chain calls S4-0080-01..06.

### S4-0080-09 (e) — track on-chain storage across partial transfers
Quote (0080:80–84): after each `transfer`, the new outcome must be
recomputed via `computeTransferEffectsAndInteractions` and tracked from
`AllocationUpdated` events to make the next `transfer`. Off-chain.

---

## Cross-source disagreements I noticed

1. **`reclaim` (S4/SATP) vs `claim` (Nitro 2019).** S4 names the
   guarantee-move entry point `reclaim` (0030:40–42, 0080:142–144). The
   Nitro 2019 paper names it `claim`. S4 matches SATP 2022. The model
   uses `.reclaim`. (This is the task's cited example.)

2. **Guarantee as a richer outcome allocation (S4/SATP) vs guarantee
   channels (Nitro 2019).** S4 research-papers page (research:26–33):
   SATP 2022 "replacement of guarantee channels with a richer Outcome
   format." Nitro 2019 funds virtual channels via separate guarantee
   channels. S4 encodes the guarantee as an `allocationType` inside the
   outcome (0030:40–51), which the model follows (`AllocationType.guarantee`).

3. **transfer argument order.** S4 `transfer(assetIndex, channelId,
   outcomeBytes, stateHash, indices)` (0080:33–39). The model's
   `Call.transfer assetIndex channelId claimedStateHash outcome indices`
   (`Operations.lean:77`) orders `stateHash` before `outcome`. Same
   fields, different order — a definitional cosmetic difference, not a
   semantic conflict.

4. **S4 says the signature struct is `{v, r, s}` / `{R, S, V}`**
   (0010:181–193). The model carries signatures only through the
   `SignatureAssumptions` abstraction, not a concrete `(v,r,s)` struct.
   Form only.

---

## Counts per class (distinct S4 properties)

- (a) on-chain adjudicator safety: 10
  (S4-0060-01, -0060-02, -0060-03, -0070-01, -0070-03, -0070-04,
   -0080-01, -0080-02, -0080-03, -0040-03) plus value-conservation
   group S4-C-01, S4-C-02, S4-0070-07, S4-0080-04 (Armada-sourced, 4).
- (b) liveness / unbeatable strategy: 3
  (S4-0070-05, -0070-06, -0040-05).
- (c) can-happen: 8
  (S4-0060-04, -0060-05, -0070-08, -0040-06, -0070-09, -0080-06,
   -0080-05, -0070-10).
- (d) definitions the model must match: 16
  (S4-0010-01..05, -0030-01..05, -0040-01, -0040-02, -0040-04,
   -0020-01, -0010-05, -0020-02).
- (e) off-chain protocol properties: 15
  (S4-0010-06, -0010-07, -0010-10, -0020-03, -0040-07, -0040-08,
   -0040-09, -0040-10, -0060-06, -0060-07, -0060-08, -0070-11,
   -0080-07, -0080-08, -0080-09).
- (f) assumptions: 4
  (S4-0010-08, -0010-09, -0060-09, -0070-12). (Several items carry both
  e and f; counted once under their primary class above.)

## MISSING items (one line each)

- S4-0060-01: no theorem states a deposit to an external destination has no successful Step (enforced in Step only).
- S4-0060-02: no theorem states a deposit with holdings ≠ expectedHeld reverts (enforced in Step only).
- S4-0070-02: no theorem states a non-participant cannot register a challenge (enforced in Step only).
- S4-0070-03: no theorem states turnNumRecord is monotone / higher turn supersedes lower (enforced per-call only).
- S4-0040-03: no theorem states turnNumRecord is the highest supported turn, nor that conclude zeroes it.
- S4-0070-10: no `checkpoint_can_happen` theorem and no theorem that checkpoint clears a challenge to Open (Call+Step exist, untested).
- S4-0080-05 (safety side): no reclaim payout-bound theorem analogous to transfer_pays_only_the_finalized_outcome (only reclaim_can_happen exists).

## CONTRADICTED items (one line each)

- S4-0080 convenience method vs model `bulk_paths_never_succeed`: S4 (0080:77) presents `concludeAndTransferAllAssets` as a valid gas-batching helper ("There is a convenience method concludeAndTransferAllAssets which combines concluding with transferring for every asset"); the model's theorem `bulk_paths_never_succeed` (Safety.lean:2072) asserts that call ALWAYS reverts. The model intentionally departs (documented as `departures-found.md` item 5), so the S4 happy-path convenience and the model's statement conflict directly.
