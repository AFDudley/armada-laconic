# Statement candidates — merged from S1–S4

Inputs: `local://S1-statements.md` (ForceMove 2018), `local://S2-statements.md`
(Nitro 2019), `local://S3-statements.md` (SATP 2022), `local://S4-statements.md`
(docs.statechannels.org). Model: `armada-lean-adjudicator-v4@5f5116a`,
`ArmadaLean/Adjudicator/Safety.lean` (24 theorems). Contracts:
`armada-nitro/src/nitro/` and `armada-nitro/src/ArmadaNitroAdjudicator.sol`.

Line-number keys:
- `FM:n` = `.scratch/rewrite-B/forcemove-2018.txt` line n.
- `N19:n` = `.scratch/rewrite-B/nitro-2019-219.txt` line n.
- `SATP:n` = `.scratch/rewrite-B/satp-2022.txt` line n.
- `DOC/<page>:n` = docs page, line n of the `<article>` text that
  `.scratch/extract.py` prints (n is the raw HTML line). Pages: `0010`
  states-channels, `0020` execution-rules, `0030` outcomes, `0040` lifecycle,
  `0060` funding, `0070` finalizing, `0080` defunding, `research`.
- `FM.sol:n` = `src/nitro/ForceMove.sol`; `MAH.sol:n` = `src/nitro/MultiAssetHolder.sol`;
  `NA.sol:n` = `src/nitro/NitroAdjudicator.sol`; `ANA.sol:n` = `src/ArmadaNitroAdjudicator.sol`;
  `Ops:n` = model `Operations.lean`.

Precedence: DOC and SATP define current behavior (`DOC/research:33`: "Terminology
and protocol descriptions from this paper are up to date, and should match the
current implementation reasonably well."; `DOC/research:11,23`: ForceMove and
Nitro 2019 terminology "is out of date with respect to current implementation").

## Verification notes (step 2)

I opened every source line that a MISSING, PARTIAL or CONTRADICTED item cites.
All quotes below are re-read from the source. Corrections to the four reports:

1. S1-D06, S1-A01 (CONTRADICTED "mover == signer vs SignedByAll") — wrong
   premise. The model `.challenge` step needs `appRel` plus one authorized
   challenger (`Ops:193-201`), not `SignedByAll`. The contract also delegates
   the support rule to the app (`FM.sol:206-217`). Moved to D (superseded:
   the mover rule is now app logic).
2. S3-14 says strict `>` in challenge mode is "a ForceMove-2019 refinement".
   Wrong. ForceMove 2018 forbids any force-move during a challenge (`FM:1345`).
   Nitro 2019 has no such rule. The strict `>` comes from the contract
   (`FM.sol:50-51`) and DOC (`DOC/0070:76`). See F1.
3. S3-05 and S3 disagreement list say "Nitro 2019 / go-nitro include chainId".
   Wrong. `N19:155-157` hashes participants and a nonce only. `DOC/0010:155-163`
   and `NitroUtils.sol:177-185` hash `participants, channelNonce, appDefinition,
   challengeDuration`, with no chainId. See F6.
4. S4-0080-04 cites `DOC/index:9` "no per-payment fee" for `payout_takes_no_fee`.
   That sentence is about off-chain payments, not the on-chain payout. I
   dropped it as support. `N19:765-770` supports the theorem instead.
5. S2-43 (CONTRADICTED, claim vs reclaim) is not a model defect. Nitro 2019
   `claim` is superseded by SATP/DOC `reclaim`. Moved to D.
6. S2-29 (PARTIAL, "unaffordable transfer fails") — the contract does not
   fail. It pays `min(amount, surplus)` (`MAH.sol:51,167`). Moved to D/F4.
7. S2-16 marks `unbeatable_pays` as COVERED for "directly funded pays all".
   The theorem has NO funding hypothesis, so it claims more than the source.
   Moved to B1.
8. S4-0070-05/06 mark `unbeatable_pays*` COVERED "exactly". The hypotheses
   have gaps (B1, B2). Kept the source mapping; the change is in B.
9. S1-A16 and S2-26/27 say `afford` covers priority payout. `afford`
   (`Outcome.lean:185-189`) is a definition only. No theorem states the exact
   payout amount. Listed as A-01.
10. S4-0040-03 quote line: "turNumRecord is the highest turn number ..." is
    `DOC/0040:150`, and the conclude reset is `DOC/0040:151` (report said
    149-152; content correct).
11. S2-22 (two participants cannot hold different finalizable outcomes,
    `N19:612-614`) is a consequence of the paper's definition of
    finalizability, not an adjudicator rule. Merged into A-04 (on-chain
    single-outcome rule, `N19:301-302`).

All other cited quotes say what the reports claim.

---

## A. Candidate statements NOT in the model (MISSING, on-chain, in scope)

### A-01 — transfer pays the exact priority-order amount
- Statement. Hypotheses: a `transfer` call on channel `cid`, asset `i`, with
  `indices`, succeeds. Let `allocs` be the asset's allocations and `held` the
  channel's held value before the call. Conclusion: each selected entry `k` is
  paid exactly `min(allocs[k].amount, held − Σ_{j<k} min(allocs[j].amount, …))`
  (that is, `afford allocs held` at `k`). Every unselected entry keeps its
  amount. The sum paid is at most `held`. Caller: anyone. The caller chooses
  only `indices`; the caller cannot change any amount.
- Class: safety.
- Quotes:
  - `N19:747-751` "The allocation is in priority order, so that if the channel
    does not hold enough funds to pay all the coins that are due, then the
    addresses at the beginning of the allocation will receive funds first. We
    say that 'A can afford x for B', if B would receive at least x coins, were
    the coins currently held by A to be paid out in priority order."
  - `N19:765-770` T(A,B,x) "Reduces the funds held in channel A by x.
    Increases the funds held by B by x. Reduces the amount owed to B in the
    outcome of A by x."
  - `FM:1845-1849` `pending := sum(r[i] - w[i]; i < k)`; `amount := min(owed
    - w[k], Adjudicator.balance - pending)`.
  - `DOC/0060:54` "If a channel were to finalize and be liquidated when
    underfunded ... the funds would be paid out in priority order."
  - `SATP:377-379` (fn 9) "the order may also dictate which destinations
    receive priority in case the channel is insufficiently funded."
- Contract: `MAH.sol:157-190` (`affordsForDestination = min(amount, surplus)`
  at :167, `surplus -= affordsForDestination` at :189).
- Gap: `transfer_pays_only_the_finalized_outcome` states only `paid ≤ amount`.

### A-02 — transfer order does not change a simple-allocation payout
- Statement. Hypotheses: a finalized channel whose asset outcome has only
  simple allocations; no deposit and no reclaim on it. Conclusion: for each
  entry, the total that entry receives from any sequence of `transfer` calls
  that pays all entries is the same (equal to `afford allocs held` at that
  entry). A transfer of a later entry never reduces what an earlier entry can
  receive. Caller: anyone; influences only the call order.
- Class: safety.
- Quotes:
  - `N19:842-843` "It is also easy to see that calling transfers out of order
    does not affect the ultimate result."
  - `FM:1814-1816` "The resolution at this point would attempt to give
    aResolution to Alice and bResolution to Bob, paying out Alice first,
    allowing her to recover her funds even if Bob does not deposit."
- Contract: `MAH.sol:161-190` (surplus is consumed for every index, selected
  or not).

### A-03 — the turn-number record never decreases on a live channel
- Statement. Hypotheses: a step from `s` to `s'`; channel `cid` is not
  finalized at `s`. Conclusion: if the call is `challenge` or `checkpoint`,
  then `s'.turnNumRecord ≥ s.turnNumRecord`; strictly `>` for `checkpoint`
  and for `challenge` in challenge mode. Only `conclude` lowers it (to 0), and
  `conclude` finalizes the channel. No other call changes it. Caller: anyone
  holding a supported state; the caller chooses the state, not the rule.
- Class: safety.
- Quotes:
  - `DOC/0020:11` "Each state has a version number, with greater version
    numbers superceding lesser ones."
  - `DOC/0020:19-21` "The only things enforced by the core protocol are: the
    rule that higher turn numbers take precedence over lower ones".
  - `DOC/0010:138` "The turn number is the mechanism by which newer states
    take precedence over older ones."
  - `DOC/0040:150-151` "turNumRecord is the highest turn number that is known
    to the chain to be supported by a full set of signatures. The exception to
    this rule is that it is set to 0 when the channel is concluded via a
    conclude call."
  - `DOC/0070:66` checkpoint establishes "a new and higher turnNumRecord".
  - `SATP:308` `require(supportedStateVars.version >= adjudicatorStatus.version)`.
  - `FM:1377-1381` "each of these responses lead to an increase in the turnNum
    ... this prevents the game from stalling indefinitely."
  - `FM:1543-1545` "store the turnNum of the response state in the
    adjudicator. Only allow new force-moves if their response will increase
    this stored turnNum."
- Contract: `FM.sol:48-55`, `FM.sol:100`, `FM.sol:277-294`, `FM.sol:160-167`.

### A-04 — finalization is permanent and fixes the state
- Statement. Hypotheses: channel `cid` is finalized at reachable `s`; any
  later reachable `s'`. Conclusion: `cid` is finalized at `s'`; its
  `stateHash` and `finalizesAt` are unchanged; no `challenge`, `checkpoint` or
  `conclude` on `cid` succeeds; only `transfer` and `reclaim` change its
  `outcomeHash`, and only by an `OutcomeDerivedFrom` rewrite. Caller: anyone.
- Class: safety.
- Quotes:
  - `N19:301-302` "it is crucial that the rules of operation only allow one
    outcome to be finalized for each channel."
  - `N19:612-614` "The definition of finalizability excludes the case where
    two different finalizable outcomes are held by different participants".
  - `FM:1262-1265` "the game transitions into the Terminated mode. In this
    mode, no further play is possible".
  - `SATP:294-303` `isOpen` and `require(isOpen(adjudicatorStatus))` in
    `challenge`.
  - `DOC/0040:104-110` the mode diagram has no edge out of `Finalized`.
- Contract: `FM.sol:54`, `FM.sol:99`, `FM.sol:147`, `FM.sol:301-303`;
  `MAH.sol:205,409-413` (fingerprint rewrite only).

### A-05 — only an authorizing participant can register a challenge
- Statement. Hypotheses: a `challenge fp proof candidate challenger` step
  succeeds. Conclusion: `challenger ∈ fp.participants` and `challenger`
  authorized `stateHash fp candidate`. Equivalent negative form: a
  non-participant, or a participant who did not sign, cannot register a
  challenge. Caller: anyone may submit; the caller cannot forge the
  challenger's authorization.
- Class: safety.
- Quotes:
  - `DOC/0070:52-56` "The challenger needs to sign this data:
    keccak256(abi.encode(challengeStateHash, 'forceMove')) ... This mechanism
    allows the challenge to be authorized only by a channel participant."
  - `FM:1546-1548` "Add a certificate argument to the force-move operation, to
    identify the challenger ... Require that the challenger is one of the
    participants."
- Contract: `FM.sol:63`, `FM.sol:184-197`. Model enforces it at `Ops:200-201`.

### A-06 — a deposit to an external destination fails
- Statement. Hypotheses: `deposit channelId expectedHeld asset amount` where
  `channelId` is an external destination. Conclusion: no step succeeds.
  Caller: the depositor.
- Class: safety.
- Quotes: `DOC/0060:16` "Firstly, destination must NOT be an external
  destination." `DOC/0060:18` "The first rule prevents funds being escrowed
  against something other than a channelId".
- Contract: `ANA.sol:96`. Model: `Ops:142`.

### A-07 — a deposit with a wrong expectedHeld fails
- Statement. Hypotheses: `deposit channelId expectedHeld asset amount` and
  `heldValue s asset channelId ≠ expectedHeld`. Conclusion: no step succeeds.
  Caller: the depositor. Another caller can front-run and change `held`; the
  rule makes the deposit fail rather than lose funds.
- Class: safety.
- Quotes: `DOC/0060:16` "Secondly, the on-chain holdings for destination must
  be equal to expectedHeld." `DOC/0060:18` "The second rule prevents loss of
  funds: since holdings are paid out in preferential order, depositing before
  a counterparty has deposited implies that they can withdraw your funds."
- Contract: `ANA.sol:102,112`. Model: `Ops:141`.

### A-08 — a deposit adds exactly its amount to one channel
- Statement. Hypotheses: a `deposit channelId … asset amount` step from `s` to
  `s'`. Conclusion: `heldValue s' asset channelId = heldValue s asset channelId
  + amount`; every other `(asset', channel')` held value is unchanged; no
  channel status changes. Caller: anyone; the caller cannot raise a holding
  without supplying a note of that value.
- Class: safety.
- Quotes:
  - `N19:165-168` "The deposit operation, DA(x), is an on-chain operation used
    to increase A's balance by x coins. There are no restrictions on who can
    deposit coins for an address, but the transaction must always include a
    transfer of x coins into the adjudicator."
  - `N19:785-787` "a deposit into the system cannot reduce the value of any
    address and cannot increase the value of any address by more than the
    value of the deposit."
- Contract: `ANA.sol:86,111-112`. Gap: `topup_can_happen` shows one witness;
  no universal statement.

### A-09 — checkpoint can happen and clears a challenge to Open
- Statement. Existential: there is a reachable `s` with channel `cid` in
  challenge mode and a `checkpoint` step to `s'` with a higher supported turn;
  at `s'` the mode is `open`, `turnNumRecord` is the new turn, and
  `finalizesAt = 0`. Caller: anyone with a supported higher-turn state.
- Class: can-happen.
- Quotes:
  - `DOC/0070:66` "The checkpoint method allows anyone with a supported
    off-chain state to establish a new and higher turnNumRecord on chain, and
    leave the resulting channel in the 'Open' mode. It can be used to clear a
    challenge."
  - `DOC/0040:105,108` `Open --> Open: checkpoint`; `Challenge--> Open: checkpoint`.
  - `DOC/0070:35` "This delay allows the challenge to be cleared by a timely
    and well-formed respond or checkpoint transaction."
  - `FM:1266-1269` "If a valid response is received, the game returns to the
    Collaborative mode".
- Contract: `FM.sol:90-119`. Model: `Ops:227-237`. Only lifecycle edge with
  no can-happen theorem.

### A-10 — transfer never pays a guarantee allocation
- Statement. Hypotheses: a `transfer` step succeeds. Conclusion: no selected
  entry has `allocationType = guarantee`. Negative form: a `transfer` whose
  `indices` select a guarantee (or `indices = []` over an outcome with a
  guarantee) has no step. Caller: anyone.
- Class: safety.
- Quotes: `DOC/0030:41` "When allocationType is set to guarantee, funds cannot
  be transferred in the usual way. Instead, they may be moved to another
  channel on chain using the reclaim method."
- Contract: `MAH.sol:169-170` `revert('cannot transfer a guarantee')`. Model
  enforces it at `Ops:314`.

### A-11 — reclaim only reallocates a finalized guarantee, value-neutral
- Statement. Hypotheses: a `reclaim` step on source `L` and the guarantee entry
  at `indexOfTargetInSource`. Conclusion: `L` and the target `V` are both
  finalized; the entry is a guarantee; the target asset equals the source
  asset; `a0 + a1 = guarantee.amount`; the new source outcome credits `left`
  by `a0` and `right` by `a1` and removes the guarantee entry; the source
  outcome total is unchanged; `holdings` and `pool` are unchanged. Caller:
  anyone; the caller chooses which guarantee, not the amounts.
- Class: safety.
- Quotes:
  - `SATP:344-364` Listing 2: `require(0 < ledgerStatus.finalizesAt <= now)`,
    `require(0 < targetStatus.finalizesAt <= now)`, `require(isGuarantee(
    guarantee))`, `o_ledger[left] += o_target[0]`, `o_ledger[right] +=
    o_target[1]`, `delete o_ledger[targetChannelId]`.
  - `N19:1072-1074` "The funds held in a guarantee channel can only be paid out
    once the outcome of the target channel is finalized."
  - `DOC/0080:143` "the parent and child channels must both be finalized on
    chain. If the child channel is funded with a guarantee, funds may be
    transferred from the child channel to the parent channel using reclaim."
- Contract: `MAH.sol:307,317-322,329-330,384-387`. Note: the deployed
  `ANA.sol` `_settleReclaimExit` (`ANA.sol:192-199`) mints value into the
  target. The model does not follow it (departure item 2, `Ops:351-355`). The
  "holdings unchanged" half is false against that code.
- Gap: only `reclaim_can_happen` exists; no safety bound for reclaim.

### A-12 — guarantee reallocation can change a later transfer's payout
- Statement. Existential (can-happen): there is a reachable `s`, a finalized
  channel `L` with a guarantee entry, and two call orders (`reclaim` then
  `transfer d`, versus `transfer d` then `reclaim`) that pay destination `d`
  different amounts. This bounds A-02 and B1: order independence holds only
  for simple allocations.
- Class: can-happen (negative fact; it limits the payment guarantee).
- Quotes:
  - `N19:1148-1151` "it is possible to construct situations where the same
    outcome can lead to different values, depending on the order in which
    guarantees are claimed."
  - `N19:1152-1154,1174-1177` lower bounds only ("participant I will receive
    their 5 coins, even though we cannot say anything about how the remaining
    5 coins will be distributed").
- The source states this for Nitro 2019 `claim`. [INFERENCE from
  `MAH.sol:359-387` and `:161-190`, not a source quote] The same effect
  exists with `reclaim`: outcome `[A:0, B:0, d:5, G:5(left=A,right=B)]`, held
  5, target outcome `[5,0]`. `transfer d` first pays `d` 5. `reclaim` first
  gives `[A:5, B:0, d:5]`, then `transfer d` pays 0.

### A-13 — channel value leaves only through transfer of a finalized channel
- Statement. Hypotheses: a step from `s` to `s'` with
  `heldValue s' a c < heldValue s a c`. Conclusion: the call is
  `transfer _ c …` and `c` is finalized at `s`. Caller: anyone; no caller can
  lower a channel's collateral any other way.
- Class: safety.
- Quotes:
  - `N19:227-229` "funds can only be withdrawn from participant addresses -
    and only by a party who knows the private key."
  - `DOC/0080:5` "Defunding can only happen after a channel finalizes."
  - `SATP:284-286` "assets deposited into an adjudicator contract are released
    only when given explicit consent on the final outcome by the channel's
    participants."
  - `FM:1718-1726` "Once the game has ended ... a game's adjudicator should
    release funds"; `Require(Adjudicator.isTerminated(channelId))`.
- Contract: holdings change only at `MAH.sol:228`, `ANA.sol:168`,
  `ANA.sol:199` (reclaim, departure) and `NA.sol:95` (bulk path, departure).
- Gap: `transfer_pays_only_the_finalized_outcome` constrains the transfer
  step only. No statement excludes a decrease by another call.

### A-14 — cumulative payout to an entry never exceeds its allocation
- Statement. Hypotheses: reachable `s0` where `c` is finalized with outcome
  `o`; any call sequence to `s1` with no `reclaim` on `c`. Conclusion: the sum
  of all amounts paid to entry `k` of asset `i` across all `transfer` calls is
  at most `o[i].allocations[k].amount`. Caller: anyone; repeated calls cannot
  pay more in total.
- Class: safety.
- Quotes:
  - `N19:836-837` "it is impossible to find a strategy that gives any node a
    value higher than allocated by this algorithm."
  - `N19:770` T "Reduces the amount owed to B in the outcome of A by x."
  - `FM:1843,1850` `w := Adjudicator.withdrawnAmounts`;
    `Adjudicator.withdrawnAmounts[k] += amount`.
- Contract: `MAH.sol:173,205` (amount reduced, fingerprint rewritten).
- Gap: the existing theorem bounds one call against the current outcome only.

---

## B. Existing statements that need a change

### B1 — `unbeatable_pays` (Safety.lean:444) — payment guarantee (known item i)
- Current text: h ∈ participants; all participants signed final `candidate`;
  channel not finalized; `0 < challengeDuration`; h signed no other final
  state; `dest ∈ candidate.outcome[i].allocations`; `dest.amount = owed`.
  After any interleaving by other actors with no tick, h's calls reach a live
  pool note at `dest.destination` worth `owed`.
- Source text:
  - `N19:749-751` "We say that 'A can afford x for B', if B would receive at
    least x coins, were the coins currently held by A to be paid out in
    priority order."
  - `N19:774-777` "if you can find one strategy to move a certain amount into
    an address, then no-one else can prevent this from occurring."
  - `N19:395-398` "where the channel L itself has enough coins in the
    adjudicator to cover the entire allocation ... A will receive all x coins
    allocated to them in ω."
  - `N19:480-481` "A channel χ is funded for participant A with x coins, if A
    has an unbeatable strategy for obtaining a state where χ is directly
    funded with x coins."
- Difference: the theorem pays `dest.amount` with no funding hypothesis. The
  source pays only what the channel "can afford" in priority order. The
  contract pays `min(amount, surplus)` (`MAH.sol:167`). For an underfunded
  channel, or a `dest` behind larger earlier entries, the current text is
  false. Second difference: guarantee entries make payouts order-dependent
  (A-12), so the source's "no-one else can prevent this" is the Turbo
  (simple-allocation) case only (`N19:775`, "reasoning about redistribution is
  easy in Turbo").
- Proposed restatement: keep all current hypotheses; replace `howed` with
  (a) `x ≤ (afford (candidate.outcome.get i).allocations (heldValue s asset
  (channelId fp))).get k`, where `k` is `dest`'s index, and (b) every
  allocation in `candidate.outcome.get i` is `simple`. Conclusion: h reaches a
  live note at `dest.destination`, token `asset`, value `≥ x` (or a set of
  notes with sum `≥ x`). Plain English: "If the channel can afford x for h in
  priority order, and the outcome holds no guarantee, then after any actions
  by others, h can obtain at least x." A deposit by others cannot lower this
  (`N19:785-787`) and transfers by others do not change it (A-02).
- Note: the case `x = dest.amount` with `held ≥ Σ amounts` is the
  "directly funded" corollary (`N19:395-398`).

### B2 — `unbeatable_pays_after_timeout` (Safety.lean:500)
- Current text: channel in challenge mode with `candidate`'s `stateHash` and
  `outcomeHash`; same honesty, no-tick and destination hypotheses as B1;
  plus "for every `highTurn > candidate.turnNum`, no
  `challenge fp proof ⟨candidate.outcome, candidate.appData, highTurn,
  candidate.isFinal⟩ challenger` step exists at `mid`". Conclusion as B1.
- Source text:
  - `DOC/0070:35` "This delay allows the challenge to be cleared by a timely
    and well-formed respond or checkpoint transaction. ... If no such
    transaction is forthcoming, the challenge will time out, allowing the
    outcome registered to be finalized."
  - `DOC/0070:76` "a challenge may be 'cleared' by another more recent
    challenge."
  - `N19:292-296` "if no other participant responds during the timeout
    period, then the challenge times out and the outcome corresponding to the
    challenge state is finalized."
  - Affordability quotes as B1.
- Differences:
  1. Same affordability gap as B1.
  2. The "no higher state" hypothesis covers only a `challenge` with the SAME
     `proof`, `challenger`, outcome, `appData` and `isFinal`. Another actor can
     still clear the challenge with `checkpoint` (`Ops:227-237`, actor
     unconstrained) or with a `challenge` that carries a different outcome or
     proof. Both are the source's "respond or checkpoint" and "more recent
     challenge". So the hypothesis does not exclude them, and the conclusion
     can fail.
- Proposed change: replace the `highTurn` hypothesis with "no supported state
  of `fp` with turn number `> candidate.turnNum` exists for any actor during
  `interleaved`": `∀ st ∈ states of interleaved, ∀ proof cand, cand.turnNum >
  candidate.turnNum → ¬ appRel st fp proof cand`. Plain English: "If no one
  holds a supported later state, then after the timeout h can obtain at least
  x." Add the B1 affordability and no-guarantee hypotheses.

### B3 — `finalized_outcome_was_concluded_or_challenged` (Safety.lean:125)
- Current text: the conclude disjunct is `∃ fp candidate, channelId fp = cid ∧
  SignedByAll sig s fp candidate ∧ …` with no `isFinal` term.
- Source text: `SATP:287-289` "by supporting a state with isFinal = true and
  submitting the support proof to a Conclude operation." `DOC/0070:11` "A
  participant wishing to end the state channel will sign a state with
  isFinal = true ... Once a full set of n such signatures exists (this set is
  known as a finalization proof)". Contract: `FM.sol:148`.
- Difference: the disjunct accepts any unanimously signed state. The source
  requires a final state.
- Proposed change: add `candidate.isFinal = true` to the conclude disjunct.
  The model `.conclude` step already requires it (`Ops:264`), so the stronger
  statement holds by the same invariant.

### B4 — `bulk_paths_never_succeed` (Safety.lean:2072)
- Current text: `transferAllAssets` and `concludeAndTransferAllAssets` have no
  successful step.
- Source text: `DOC/0080:77` "There is a convenience method
  concludeAndTransferAllAssets which combines concluding with transferring for
  every asset -- batching them to save gas."
- Difference: DOC presents the bulk path as valid. The model removes it on
  purpose (Armada departure item 5; `Ops:387-403`). The deployed code still
  runs it (`NA.sol:21-51`, raw arithmetic on a commitment at `NA.sol:95`).
- Proposed change: none to the text. Record it as an Armada-specific rule
  with no paper support (see C). The operator decides whether to keep it.

---

## C. Existing statements and their sources (all 24)

| # | Theorem | Sources | Status |
|---|---|---|---|
| 1 | `finalized_outcome_was_concluded_or_challenged` | S1 (FM:1262-1265, 1504-1507); S2-11 (N19:288-300); S3-16 (SATP:284-292); S4-0070-04 | COVERED; change B3 |
| 2 | `transfer_pays_only_the_finalized_outcome` | S1-A15 (FM:1718-1726); S2-10, S2-28 (N19:261-266, 765-770); S4-0080-01/02/03 (DOC/0080:5,9) | COVERED; stronger form A-01 |
| 3 | `unbeatable_pays` | S1-B01 (FM:183-188); S2-16/21/30 (N19:395-398, 590-595, 774-777); S4-0070-05 (DOC/0070:19-21) | PARTIAL; change B1 |
| 4 | `unbeatable_pays_after_timeout` | S1-B03 (FM:449-453, 1314-1318); S2-12 (N19:292-296); S4-0070-06 (DOC/0070:35) | PARTIAL; change B2 |
| 5 | `notedValue_conserved` | S2-08 (N19:230-231), Armada form | COVERED |
| 6 | `notedValue_conserved_reachable` | S2-08 (N19:230-231), Armada form | COVERED |
| 7 | `holdings_point_to_live_notes` | none (Armada note encoding) | NO SOURCE |
| 8 | `deposit_can_happen` | S1-C01 (FM:874-876); S2-04 (N19:165-168); S4-0060-04 (DOC/0060:9) | COVERED |
| 9 | `topup_can_happen` | S2-04 (N19:165-168); S4-0060-05 (DOC/0060:54) | COVERED |
| 10 | `challenge_can_happen` | S1-C02 (FM:1340-1354); S2-12; S3-14 (SATP:300-315); S4-0070-08 (DOC/0070:31-33) | COVERED |
| 11 | `challenge_timeout_can_happen` | S1-B03 (FM:1260-1265); S2-12 (N19:292-296); S3-13 (SATP:294-298); S4-0040-05/06 (DOC/0040:110,116) | COVERED |
| 12 | `conclude_can_happen` | S1-A14 (FM:1495-1507); S2-13 (N19:296-300); S3-16 (SATP:287-289); S4-0070-09 (DOC/0070:19-21) | COVERED |
| 13 | `payout_can_happen` | S1-C03 (FM:1718-1720); S2-15 (N19:312-314); S4-0080-06 (DOC/0080:9) | COVERED |
| 14 | `reclaim_can_happen` | S3-15 (SATP:344-364); S4-0080-05 (DOC/0080:143) | COVERED |
| 15 | `payout_takes_no_fee` | S2 (N19:765-770: debit x, credit x) | COVERED (S4 index:9 support dropped, note 4) |
| 16 | `conclude_moves_no_value` | S2-10 (N19:261-266, finalization separate from redistribution); S4-0070-07 (DOC/0070:25) | COVERED |
| 17 | `bulk_paths_never_succeed` | none; DOC/0080:77 contradicts | NO SOURCE; contradicted (B4) |
| 18 | `value_exists_only_as_notes` | none (Armada encoding) | NO SOURCE |
| 19 | `holdings_never_zero_valued` | none (Armada encoding) | NO SOURCE |
| 20 | `heldValue_zero_iff_no_collateral` | S2-01 (N19:161-163 "If an address does not appear in the table we take the balance to be zero") | COVERED |
| 21 | `configuredAssets_invariant` | none (go-nitro/Armada asset list) | NO SOURCE |
| 22 | `zero_payout_mints_nothing` | indirect only (N19:765-770) | NO DIRECT SOURCE |
| 23 | `full_payout_leaves_no_remainder` | indirect only (N19:765-770) | NO DIRECT SOURCE |
| 24 | `nullify_iff_commitment_was_some` | none (Armada nullifier encoding) | NO SOURCE |

No source supports: #7, #17, #18, #19, #21, #24. Only indirect support: #22,
#23. These are Armada-layer rules; the four sources do not state them.

---

## D. SUPERSEDED items (older source replaced by a newer one)

| Item (report IDs) | Old quote | Replacing quote |
|---|---|---|
| Single mover per turn; `mover == signer` (S1-D05/D06, S1-A01) | FM:748 `Require(toMove.state.mover == toMove.signer)` | DOC/0020:19 "Each application gets full control over when a state can be considered supported."; SATP:254-259 `latestSupportedState`; contract `FM.sol:206-217` |
| turnNum must increment by exactly 1 (S1-A02) | FM:762-778 `Require(toState.turnNum == fromState.turnNum + 1)` | DOC/0010:138 "The turn number usually increments as the channel progresses."; core rule only "higher turn numbers take precedence" (DOC/0020:21) |
| PreFundSetup / PostFundSetup state types and transitions; pre-GAME back-out; prefund round arms recovery (S1-D09/D10, S1-A03/A06/A07/A08, S1-B06) | FM:1071-1184 | DOC/0010:17-35 one `VariablePart` with `turnNum`/`isFinal`; SATP:644-667 prefund/postfund are app versions 0/1 |
| ConcludeState type and n-state conclusion proof (S1-D08, S1-A05, S1-A09) | FM:1030-1058 | DOC/0070:11 "sign a state with isFinal = true ... full set of n such signatures"; `FM.sol:148-157` |
| No force-move while a challenge is in progress (S1-A10) | FM:1345 `Require(!Adjudicator.challengeInProgress(channelId))` | DOC/0070:76 "a challenge may be 'cleared' by another more recent challenge"; `FM.sol:50-51` |
| refute, respondWithMove, alternativeRespondWithMove (S1-A11/A12/A13, S1-E03) | FM:1399-1478 | DOC/0070:66 checkpoint "can be used to clear a challenge"; DOC/0070:76; `FM.sol:39-119` (no respond method) |
| Single-channel simple adjudicator (S1-A17) | FM:1636-1638 | DOC/0040:174-180 `mapping(bytes32 => bytes32) public statusOf` |
| Five modes (S1-D11) | FM:1237-1247 | DOC/0040:97-99 "Open, Challenge and Finalized" |
| channelId = h(gameType, participants, nonce) (S1-D01) | FM:606 | DOC/0010:155-163 four fields incl. challengeDuration |
| Nitro 2019 `claim` C(G,A,x) (S2-43) | N19:1139-1145 | SATP:342-364 `reclaim(ledgerChannelId, targetChannelId)`; DOC/0030:41 |
| Guarantee channels (S2-39/41/42) | N19:1070-1101 | DOC/research:27 "the replacement of guarantee channels with a richer Outcome format"; DOC/0030:40-51 |
| Unaffordable T(A,B,x) fails (S2-29) | N19:771-772 | Contract `MAH.sol:51` "Transfers as many funds ... as can be afforded"; DOC/0080:45 indices select entries, no amount argument |
| Withdraw W_A(x) from a participant address with a signature (S2-06) | N19:169-199 | DOC/0030:54 external destination; transfer pays it directly (`MAH.sol:445-446`); Armada pays a pool note (`ANA.sol:144-147`) |
| Forfeitable challenger deposit (S1-E05 part) | FM:1576-1586 | not in DOC/SATP/contract; challenger signature only (DOC/0070:52-56) |

---

## E. Off-chain and out-of-scope items

- S1-D02 / S4-0010-07: participants must pick a unique nonce; never join a reused channel id (FM:629-632, DOC/0010:108). Model has `channelIdOf_inj` (Assumptions.lean:77).
- S1-B05: a late observer loses nothing by playing on (FM:1518-1525). References superseded ops.
- S1-B07 / S4-0060-06 / S4-0040-10: deposit in priority order; wait for higher-priority deposits (FM:1197-1199, DOC/0060:54). On-chain part is A-01/A-07.
- S1-E01: payment app conserves sum, mover only gives (FM:536-551). App logic.
- S1-E02: default-winner incentive (FM:339-346, 458-461). App design.
- S1-E04: rational back-out before GAME (FM:1187-1196).
- S1-E05: griefing cannot be eliminated (FM:1566-1575).
- S2-17/18/24, S2-33, S2-45: value, enabled-outcome, Turbo and lower-bound algorithms (N19:480-484, 641-658, 818-835, 1178-1180). Analysis definitions; A-01/A-12 carry the on-chain part.
- S2-19/20, S3-24: chain liveness and front-running assumptions (N19:537-542, SATP:166-167). Model encodes them as the `interleaved` quantifier.
- S2-23: universally finalizable at funding point (N19:632-640).
- S2-31 (withdraw half): withdrawals affect only the owner (N19:787-789). Pool layer.
- S2-32: DAG outcome networks only (N19:790-795).
- S2-37, S2-47, S2-48: ledger/virtual offload and intermediary recovery (N19:859-865, 1299-1343).
- S2-49: off-chain constructions (N19:500-728, 885-1047, 1257-1455).
- S3-17/18/19: virtual-payment app rules and voucher-sum invariant (SATP:639-741).
- S3-20: Theorem 3.2 c(Pi) = d(Pi) (SATP:503-507).
- S3-21: Lemma 3.1/A.1 total preserved in V (SATP:511-513, 746-753).
- S3-22: O(1) gas and time bound (SATP:491-493).
- S3-23: two-round lower bound conjecture (SATP:451-452).
- S3-26: deposits/withdrawals not specified in SATP (SATP:765-768).
- S4-0010-06: participants length 2..255, nonzero, self-included (DOC/0010:100).
- S4-0010-08/09: challengeDuration and turnNum range bounds (DOC/0010:122-124, 142). Model uses unbounded Nat.
- S4-0010-10: stop executing once isFinal is enabled (DOC/0010:150).
- S4-0020-02: app must revert or return false on a bad proof (DOC/0020:34-46). Model parameter `appRel`.
- S4-0020-03: VirtualPaymentApp (DOC/0020:64).
- S4-0040-07/08/09: off-chain lifecycle Proposed..Finalized and funding conditions (DOC/0040:16-83).
- S4-0060-07/08: ledger and virtual funding constructions (DOC/0060:56-294).
- S4-0060-09: ETH vs ERC20 deposit mechanics (DOC/0060:20,50). Armada uses notes.
- S4-0070-11/12: respond prose; clients cache events (DOC/0070:35,80).
- S4-0080-07/08/09: off-chain defunding and outcome tracking (DOC/0080:80-144).

---

## F. Source disagreements and the contract resolution

### F1 — re-challenge at an equal turn in challenge mode (known item ii)
- SATP: `SATP:294-298` `isOpen` is true while `finalizesAt > now`, so a
  running challenge is "open"; `SATP:308` `require(supportedStateVars.version
  >= adjudicatorStatus.version)`. So SATP accepts a re-challenge at an EQUAL
  version during a challenge.
- ForceMove 2018: `FM:1345` forbids any force-move during a challenge;
  `FM:1543-1545` proposes "Only allow new force-moves if their response will
  increase this stored turnNum."
- DOC: `DOC/0070:76` "a challenge may be 'cleared' by another more recent
  challenge."
- Contract: `FM.sol:48-55`. Open mode calls `_requireNonDecreasedTurnNumber`
  (`FM.sol:288-294`, `newTurnNumRecord >= turnNumRecord`). Challenge mode calls
  `_requireIncreasedTurnNumber` (`FM.sol:277-280`, `newTurnNumRecord >
  turnNumRecord`). Finalized mode reverts (`FM.sol:54`, `:301-303`).
- Resolution: the contract implements strict `>` in challenge mode and `≥` in
  open mode. The model (`Ops:195-198`) matches the contract. SATP's listing is
  more permissive than the deployed code. No model change.

### F2 — payout verb and shape
- `FM:1725` `withdraw`; `N19:1139-1145` `claim C(G,A,x)`; `SATP:344` and
  `DOC/0030:41` `reclaim`; `DOC/0030:36-38` simple allocations use `transfer`.
- Contract: `MAH.sol:59-109` `transfer`, `MAH.sol:237-277` `reclaim`. Model
  follows the contract.

### F3 — reclaim checks
- SATP Listing 2 (`SATP:344-364`) checks both channels finalized and the entry
  is a guarantee. It does not check stored hashes, asset match, or that the
  credited sum equals the guarantee amount.
- Contract adds fingerprint checks (`MAH.sol:308-312,331-335`), asset match
  (`MAH.sol:329`), left/right present (`MAH.sol:385-386`), and
  `totalReclaimed == guarantee.amount` (`MAH.sol:387`). Model follows the
  contract (`Ops:369-378`).

### F4 — unaffordable transfer
- `N19:771-772` "If A cannot afford x for B, then T(A, B, x) fails, leaving the
  on-chain state unchanged."
- Contract `MAH.sol:51,167` pays `min(amount, surplus)`; there is no `x`
  argument; a zero payout succeeds. Model follows the contract
  (`zero_payout_mints_nothing`).

### F5 — repeated destinations in an allocation
- `N19:744-746` "require that implementations enforce this by ignoring any
  additional entries for a given address after the first."
- Contract `MAH.sol:51-52` "Assumes no repeated entries." Nothing enforces it;
  each entry is paid in priority order. Model `afford` also pays each entry.
  No statement proposed: a statement of the N19 rule is false against the
  contract.

### F6 — channelId inputs
- `FM:606` `h(gameType, participants, channelNonce)`; `N19:155-157`
  participants and a nonce; `SATP:213` `hash(peers, appDef, nonce)`;
  `DOC/0010:155-163` participants, channelNonce, appDefinition,
  challengeDuration.
- Contract `NitroUtils.sol:177-185` uses the DOC four fields. Model
  `Channel.lean:66` matches.

### F7 — app rule purity
- `FM:963-966` the transition function "must be pure ... cannot read from or
  write to the blockchain."
- DOC/0020:34-46 `stateIsSupported(...) external view`. Contract
  `FM.sol:206-217` calls the app as `view`. Model `appRel` takes
  `AdjudicatorState`, which matches `view`.

### F8 — conclude during a challenge
- `FM:1497` conclude needs only `!isTerminated`. DOC/0040:109
  `Challenge--> Finalized: conclude`. SATP gives prose only (`SATP:286-289`).
- Contract `FM.sol:147` `_requireChannelNotFinalized`. No turn check on
  conclude, and it sets `turnNumRecord := 0` (`FM.sol:160-167`). Model
  `Ops:261-268` matches.

### F9 — bulk payout path
- DOC/0080:77 presents `concludeAndTransferAllAssets` as valid.
- Deployed base code `NA.sol:21-51` runs it; `NA.sol:95` subtracts from a
  commitment as a raw integer. Model says it never succeeds (Armada departure
  item 5). The contract does not resolve this; the deployed code disagrees
  with the model. See B4.

### F10 — reclaim value effect in Armada
- SATP:358-363 and DOC/0030:41 move value between outcome entries only.
- Deployed `ANA.sol:192-199` `_settleReclaimExit` mints a commitment worth
  `priorHeld + guaranteeAmount` for the target. Model holds `holdings` and
  `pool` unchanged (departure item 2, `Ops:351-355`). The deployed code
  disagrees with the sources and the model. See A-11.
