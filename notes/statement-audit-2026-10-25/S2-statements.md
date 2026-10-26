# S2 — Nitro Protocol (Close, 2019, ePrint 2019/219)

Source text: `.scratch/rewrite-B/nitro-2019-219.txt` (1505 lines; last content line
1504; line 1505 is blank). I read every line in order, lines 1–1504.
Line numbers below are text line numbers.

The model maps paper terms to Armada terms. A "balance of an address" is the
value of notes paid to that note key. A payout creates a pool note at the
recipient's note key. I keep each paper property even when Armada changes its
form.

The 24 model theorems are in `ArmadaLean/Adjudicator/Safety.lean`. I name them
in the coverage lines.

---

## §3.1 System of balances (lines 139–231)

### S2-01 — Adjudicator stores a balance per address; absent = 0
- Class: (d) definition.
- Quote (161–163): "We model the adjudicator as having a simple mapping that
  stores the quantity of coins for an address. If an address does not appear in
  the table we take the balance to be zero."
- Property: the adjudicator state maps an address to a value. A missing address
  reads as value 0. Armada form: `heldValue` reads a note's value, and no note
  reads as 0.
- Coverage: COVERED. `heldValue_zero_iff_no_collateral` states that a channel
  with no recorded collateral has `heldValue = 0`. `value_exists_only_as_notes`
  states there is no raw-balance field.

### S2-02 — Address kinds; channel id = hash(participants, nonce)
- Class: (d) definition (channel id format).
- Quote (152–160): "store the balance of coins held for a given address.
  Addresses can correspond either to participants in the network or to state
  channels. A participant address is a regular blockchain address... A channel
  address is formed by taking the hash of the participant addresses along with a
  nonce, k".
- Property: a channel address is the hash of the fixed participant set and a
  nonce. The caller cannot choose a channel id that collides with another.
- Coverage: COVERED. `channelId fp` computes the id from `FixedPart`. Many
  theorems key channels by `channelId fp`.

### S2-03 — Unforgeability assumption
- Class: (f) cryptography assumption.
- Quote (158–160): "we assume that the properties of the signature scheme and
  hashing algorithm make it impossible to find a private key for A or to
  construct a channel whose address is A, if they are not previously known."
- Property: no caller can forge a signature for an address, and no caller can
  build a second channel with a given id.
- Coverage: COVERED. `SignatureAssumptions` and `HashAssumptions` are explicit
  arguments to the theorems. `AuthorizedBy` reads recorded signatures.

### S2-04 — Deposit raises a balance; anyone may deposit
- Class: (c) can-happen.
- Quote (165–168): "The deposit operation, DA (x), is an on-chain operation used
  to increase A's balance by x coins. There are no restrictions on who can
  deposit coins for an address".
- Property: any caller can call deposit. The deposit raises the target's value
  by x.
- Coverage: COVERED. `deposit_can_happen` builds an execution where a deposit
  succeeds. `Step` takes any `Actor`.

### S2-05 — A deposit must carry x coins into the adjudicator
- Class: (a) on-chain safety.
- Quote (166–168): "the transaction must always include a transfer of x coins
  into the adjudicator." Also (230–231): "The total coins held by the smart
  contract should always equal the sum of the balances."
- Property: a deposit of x adds a note of value x; the caller cannot raise a
  balance without supplying the matching value.
- Coverage: PARTIAL. `deposit_can_happen` records a `.signNote` that supplies a
  spendable note before the deposit; `notedValue_conserved` forbids minting.
  There is no theorem that states a deposit's recorded amount must equal the
  note value supplied in the same call.

### S2-06 — Withdrawal only from participant addresses, only with the private key
- Class: (a) on-chain safety.
- Quote (169–199): "The withdrawal operation, WA (x), can be used to withdraw
  coins held at participant address, A, by any party with the knowledge of the
  corresponding private key... withdraw(fromAddr, toAddr, amount, signature),
  where signature is A's signature of the other parameters". Also (227–229):
  "funds can only be withdrawn from participant addresses - and only by a party
  who knows the private key."
- Property: a withdrawal needs the holder's signature. Armada form: a payout
  creates a pool note at the recipient note key; the recipient spends that note
  from the pool with the recipient's own key.
- Coverage: MISSING. The model `Call` has no `withdraw` constructor
  (`Operations.lean:71–86`). The withdraw step is replaced by a pool note. No
  model theorem states that only the note-key owner can spend the paid note.
  `holdings_point_to_live_notes` states the note is live, not who can spend it.

### S2-07 — No withdrawal from a channel address
- Class: (a) on-chain safety.
- Quote (227–229): "Anyone can deposit into an address of either type but funds
  can only be withdrawn from participant addresses".
- Property: a caller cannot withdraw directly from a channel address; a channel's
  funds leave only through redistribution (transfer/claim) to a participant.
- Coverage: MISSING. No withdraw op exists in the model, so the restriction is
  implicit, not stated as a theorem.

### S2-08 — Total held = sum of balances (conservation)
- Class: (a) on-chain safety.
- Quote (230–231): "The total coins held by the smart contract should always
  equal the sum of the balances."
- Property: no call creates or destroys pool value; the pool total is invariant.
- Coverage: COVERED. `notedValue_conserved` (one step) and
  `notedValue_conserved_reachable` (whole reachable set) state per-token
  conservation. `value_exists_only_as_notes` reinforces it.

---

## §3.2 State channel outcomes (lines 233–304)

### S2-09 — Allocation outcome format
- Class: (d) definition (outcome format).
- Quote (236–240): "An example of a type of outcome is the allocation... which
  consists of a list of recipient addresses and totals that specify how the
  channel's balance should be distributed."
- Property: an allocation is a list of (destination, amount) pairs.
- Coverage: COVERED. `Outcome`/`Allocation` in `Outcome.lean` has this shape.

### S2-10 — Two steps: finalization then redistribution
- Class: (d) definition.
- Quote (261–266): "split the updating of the balances into two steps: 1.
  Finalization... 2. Redistribution: updating the balances in the adjudicator
  according to that outcome."
- Property: a payout happens only after a channel is finalized; redistribution
  reads the finalized outcome.
- Coverage: COVERED (by design). `transfer`/`reclaim` require the channel mode to
  be `.finalized`; `transfer_pays_only_the_finalized_outcome` ties the payout to
  `ChannelFinalizedWith`.

### S2-11 — Exactly two ways to finalize an outcome
- Class: (a) on-chain safety.
- Quote (288–299): "There are two ways for an outcome to become finalized: (i)
  through a challenge that times out before anyone responds, and (ii) through the
  registration of a conclusion proof."
- Property: a finalized channel traces back to a timed-out challenge or a
  conclusion proof; nothing else finalizes a channel.
- Coverage: COVERED. `finalized_outcome_was_concluded_or_challenged` states these
  two paths are the only ones over the reachable set.

### S2-12 — Non-collaborative close by challenge timeout
- Class: (c) can-happen.
- Quote (292–296): "one participant launches an on-chain challenge, starting a
  timeout period; if no other participant responds during the timeout period,
  then the challenge times out and the outcome corresponding to the challenge
  state is finalized."
- Property: a participant can challenge; after the timeout with no response the
  channel finalizes.
- Coverage: COVERED. `challenge_can_happen` and `challenge_timeout_can_happen`.

### S2-13 — Collaborative close by conclusion proof; no wait
- Class: (c) can-happen.
- Quote (296–300): "all the participants sign a special conclude state, called a
  conclusion proof; any participant can then register this outcome on-chain to
  create a finalized outcome. By closing the channel collaboratively the
  participants avoid having to wait for the challenge period."
- Property: once all participants sign, any one caller can conclude and finalize
  the channel at once.
- Coverage: COVERED. `conclude_can_happen` finalizes instantly.

### S2-14 — Only one outcome is finalized per channel
- Class: (a) on-chain safety.
- Quote (301–302): "it is crucial that the rules of operation only allow one
  outcome to be finalized for each channel."
- Property: a channel has at most one finalized outcome at once; no caller can
  finalize a second, conflicting outcome.
- Coverage: PARTIAL. The model stores a single `stateHash`/`outcomeHash` per
  channel status, so a finalized channel is single-valued by construction
  (`ChannelFinalizedWith` reads fixed fields). No theorem states that a finalized
  channel cannot be re-finalized to a different outcome.

### S2-15 — Anyone can trigger redistribution operations
- Class: (c) can-happen.
- Quote (312–314): "The operations change the state of the adjudicator...
  There is no restriction on who can trigger the operations."
- Property: any caller can call transfer/claim on a finalized channel.
- Coverage: COVERED. `Step` takes any `Actor`; `payout_can_happen` and
  `reclaim_can_happen` build such executions.

---

## §4 Reasoning about state channels (lines 359–709)

### S2-16 — Directly funded channel pays the full allocation
- Class: (b) unbeatable strategy.
- Quote (395–398): "where the channel L itself has enough coins in the
  adjudicator to cover the entire allocation. In this case, we say the channel is
  directly funded... A will receive all x coins allocated to them in ω."
- Property: if a channel holds enough to cover an allocation, the allocated
  participant can obtain the full allocated amount.
- Coverage: COVERED. `unbeatable_pays` states an allocated participant can reach
  a pool note of the owed value. `payout_can_happen` pays the whole deposit.

### S2-17 — Definition of "funded for A with x coins"
- Class: (d) definition.
- Quote (480–481): "A channel χ is funded for participant A with x coins, if A
  has an unbeatable strategy for obtaining a state where χ is directly funded with
  x coins."
- Property: funding is an unbeatable-strategy statement, not a one-step fact.
- Coverage: COVERED (as the frame of `unbeatable_pays`), which quantifies over all
  attacker interleavings and still reaches the payout.

### S2-18 — Definition of "value of a network state"
- Class: (d) definition.
- Quote (482–484): "The value for participant A of a network state Σ is the
  maximum x for which A has an unbeatable strategy for obtaining a state where the
  balance of A's address in the adjudicator is x coins."
- Property: value = max amount A can unbeatably bring to A's own note key.
- Coverage: PARTIAL. `unbeatable_pays` gives a lower bound (A obtains the owed
  amount). No theorem states the maximality (no strategy exceeds the value).

### S2-19 — Blockchain assumption: transactions are unimpeded
- Class: (f) chain assumption.
- Quote (537–540): "Transactions are unimpeded: given that the current time is t
  and  > 0, then it is possible for any party to apply any operation, O, on-chain
  before time t + ."
- Property: the honest caller can always get a call on-chain before any deadline.
- Coverage: PARTIAL. `unbeatable_pays` lets the honest party append `payCalls`
  after any attacker interleaving, which encodes "can still act". The time bound
  t+ε is not modeled; the theorem forbids attacker `tick` calls instead.

### S2-20 — Blockchain assumption: transactions can be front-run
- Class: (f) chain assumption.
- Quote (540–542): "Transactions can be front-run: given two parties, p1 and p2 ,
  and two operations, O1 and O2 , there is no way for p1 to ensure that they can
  apply O1 to the chain before p2 applies O2 ."
- Property: a strategy must not depend on winning an ordering race.
- Coverage: PARTIAL. `unbeatable_pays` quantifies over all attacker interleavings
  `mid`, so the honest payout cannot rely on going first. The front-run freedom is
  modeled as an arbitrary attacker prefix, not as a per-pair race.

### S2-21 — Finalizable outcome; strategy drives it on-chain
- Class: (b) unbeatable strategy.
- Quote (590–595): "an outcome, Ω, is finalizable for participant A, if A has an
  unbeatable strategy for finalizing this outcome in the adjudicator... [χ ↦ Ω]A
  → Jχ ↦ ΩK".
- Property: a finalizable outcome can be forced on-chain by its holder against any
  opponent.
- Coverage: COVERED. `unbeatable_pays` takes `SignedByAll` (the concluded,
  finalizable case) and drives the channel to a payout.

### S2-22 — No two conflicting finalizable outcomes held by different participants
- Class: (a) on-chain safety.
- Quote (597–614): "exactly one of the following statements is true about a
  channel χ... The definition of finalizability excludes the case where two
  different finalizable outcomes are held by different participants, as in this
  case at least one participant's strategy would be beatable by the other
  participant's strategy."
- Property: a reachable state cannot let two participants each force a different
  outcome of the same channel.
- Coverage: MISSING. No model theorem states this mutual exclusion over the
  reachable set.

### S2-23 — Universally finalizable at funding point and concluded state
- Class: (d) definition.
- Quote (632–640): "the outcome of a channel is finalizable by all its
  participants, we say that the outcome is universally finalizable. For a
  ForceMove channel, this happens at... 1. ...the funding point. 2. ...the
  concluded state... all channels have one universally finalizable state at the
  beginning of their lifecycle and one at the end".
- Property: every channel has a universally finalizable state at open and at
  conclude.
- Coverage: PARTIAL. The concluded case is modeled by `SignedByAll` +
  `conclude_can_happen`. The funding-point universal finalizability is not stated.

### S2-24 — Enabled outcomes; the per-participant dichotomy
- Class: (d) definition.
- Quote (641–658): "The enabled outcomes for a participant, p, is defined as the
  set of outcomes that p has no strategy to prevent from being finalized... For
  any participant, p... exactly one... is true: 1. p has at least one finalizable
  outcome. 2. p has at least two enabled outcomes. ...if a participant has only
  enabled a single outcome, that outcome must be finalizable for them."
- Property: a participant either can force an outcome, or faces at least two
  possible outcomes.
- Coverage: MISSING. The model does not define "enabled outcome".

(Consensus game §4.5 lines 660–688 and "Outcomes First" §4.6 lines 694–728 are
off-chain protocol reasoning; see S2-49.)

---

## §5 Turbo protocol (lines 731–1047)

### S2-25 — Each address appears once in an allocation
- Class: (d) definition.
- Quote (742–746): "We assume that each address only appears once in the
  allocation and require that implementations enforce this by ignoring any
  additional entries for a given address after the first."
- Property: the adjudicator must treat only the first entry for a repeated
  destination in an allocation.
- Coverage: MISSING. No model theorem states address uniqueness or first-entry
  handling in `transferEffects`.

### S2-26 — Allocations pay in priority order
- Class: (d) definition.
- Quote (747–749): "The allocation is in priority order, so that if the channel
  does not hold enough funds to pay all the coins that are due, then the addresses
  at the beginning of the allocation will receive funds first."
- Property: an underfunded channel pays earlier allocation entries first.
- Coverage: COVERED. `afford`/`transferEffects` in `Outcome.lean` pay in list
  order.

### S2-27 — Definition: "A can afford x for B"
- Class: (d) definition.
- Quote (749–751): "We say that 'A can afford x for B', if B would receive at
  least x coins, were the coins currently held by A to be paid out in priority
  order."
- Property: affordability = B's priority-ordered share is at least x.
- Coverage: COVERED. `afford` in `Outcome.lean`.

### S2-28 — Transfer effects when affordable
- Class: (a) on-chain safety.
- Quote (765–770): "If A can afford x for B, then T(A, B, x): 1. Reduces the funds
  held in channel A by x. 2. Increases the funds held by B by x. 3. Reduces the
  amount owed to B in the outcome of A by x."
- Property: a successful transfer debits A by x, pays B x, and cuts B's allocation
  entry by x; no value is withheld.
- Coverage: COVERED. `transfer_pays_only_the_finalized_outcome` (paid ≤ allocated,
  same destination), `payout_takes_no_fee` (paid + remainder = held),
  `full_payout_leaves_no_remainder`, `zero_payout_mints_nothing`.

### S2-29 — Unaffordable transfer leaves the state unchanged
- Class: (a) on-chain safety.
- Quote (771–772): "If A cannot afford x for B, then T(A, B, x) fails, leaving the
  on-chain state unchanged."
- Property: a transfer that is not covered has no effect.
- Coverage: PARTIAL. In the model an unaffordable/ill-formed transfer has no
  successful `Step` (so no state change). `bulk_paths_never_succeed` states the
  bulk paths never step. No theorem states the per-arg failure case for a plain
  `transfer` directly.

### S2-30 — Turbo redistribution is unbeatable
- Class: (b) unbeatable strategy.
- Quote (774–777): "if you can find one strategy to move a certain amount into an
  address, then no-one else can prevent this from occurring."
- Property: a found transfer strategy cannot be blocked by any other caller.
- Coverage: COVERED (as a superset). `unbeatable_pays` states this for Nitro,
  which includes Turbo.

### S2-31 — Deposits/withdrawals cannot help or harm a strategy
- Class: (f) assumption / reasoning.
- Quote (779–789): "Deposits and withdrawals cannot be required as part of a
  strategy and cannot help as part of a counter-strategy... a deposit into the
  system cannot reduce the value of any address and cannot increase the value of
  any address by more than the value of the deposit. Withdrawals can only occur
  from participant addresses... so cannot affect values elsewhere."
- Property: a deposit cannot lower any address value and raises value by at most
  its own amount; a withdrawal moves only the owner's own funds.
- Coverage: MISSING. No model theorem bounds a deposit's effect on other
  addresses, nor isolates a withdrawal's effect.

### S2-32 — Only DAG outcome networks are in scope
- Class: (f) assumption.
- Quote (790–795): "We will only consider the case where the network of outcomes
  forms a directed acyclic graph (DAG)... it is also possible for any participant
  in the channel to prevent this from happening. We therefore consider non-DAG
  outcome networks to be outside the scope".
- Property: the value results assume an acyclic outcome graph; any participant can
  block a cycle.
- Coverage: MISSING. The model does not restrict the outcome graph to a DAG.

### S2-33 — Turbo value algorithm defines node value
- Class: (d) definition.
- Quote (818–835): "Turbo Value Algorithm 1. Choose a topological ordering...
  2. ...Values[n]... balance held for n... 3. For each n... remainingFunds...
  x = min(payout, remainingFunds)... 4. Then Values[n] gives the value of node n."
- Property: a value algorithm computes each node's extractable value by
  priority-ordered payout over a topological order.
- Coverage: MISSING. No value algorithm in the model.

### S2-34 — No strategy exceeds the computed value (upper bound)
- Class: (b) unbeatable strategy.
- Quote (836–837): "it is impossible to find a strategy that gives any node a
  value higher than allocated by this algorithm."
- Property: no caller can extract more than a node's allocation.
- Coverage: PARTIAL. `transfer_pays_only_the_finalized_outcome` bounds each paid
  amount by the allocated amount in the finalized outcome. The network-wide
  maximality is not stated.

### S2-35 — A strategy reaches the computed value (achievability)
- Class: (b) unbeatable strategy.
- Quote (837–839): "It is also not hard to construct a strategy for a node to
  obtain the value allocated by the algorithm, if necessary by actually
  implementing the algorithm up until that node."
- Property: the allocated value is reachable by a sequence of transfers.
- Coverage: COVERED. `unbeatable_pays` and `payout_can_happen` build such a
  sequence.

### S2-36 — Turbo transfers are order independent
- Class: (a) on-chain safety.
- Quote (842–843): "It is also easy to see that calling transfers out of order
  does not affect the ultimate result."
- Property: in Turbo (allocations only), the payout result is independent of
  transfer order.
- Coverage: MISSING. No model theorem states order independence.

### S2-37 — Offloading a ledger channel directly funds the sub-channel
- Class: (b) unbeatable strategy.
- Quote (859–865): "χ is funded with 6 coins by L for both A and B... put this
  outcome on-chain. Once it is on-chain, the transfer operation T(L, χ, 6) is all
  that is required to make χ directly funded. From the Turbo redistribution
  result, we know that this redistribution strategy is unbeatable."
- Property: a transfer from a ledger channel to a sub-channel makes the
  sub-channel directly funded, and it cannot be blocked.
- Coverage: PARTIAL. `payout_can_happen` moves a channel's funds to a destination.
  Transfer to a channel destination that re-funds a sub-channel is not stated as
  such.

(Ledger-channel example constructions §5.4 lines 885–1047: opening/closing a
sub-channel, top-up, partial withdrawal. These are off-chain protocol; see
S2-49.)

---

## §6 Nitro protocol (lines 1050–1455)

### S2-38 — Nitro outcomes and operations
- Class: (d) definition.
- Quote (1051–1053): "In Nitro protocol, the outcome of a channel can be either
  an allocation or a guarantee. There are two on-chain redistribution operations:
  the transfer and the claim."
- Property: an outcome is an allocation or a guarantee; redistribution is transfer
  or claim.
- Coverage: COVERED. `Outcome` carries an `allocationType`; `Call` has `transfer`
  and `reclaim`. See the claim-vs-reclaim disagreement note below.

### S2-39 — Guarantee targets an allocation with a new priority order
- Class: (d) definition.
- Quote (1070–1074): "The new type of channels are called guarantee channels...
  A guarantee targets an allocation, while specifying a different priority order
  to pay out to the destinations."
- Property: a guarantee names a target allocation and a reprioritization.
- Coverage: COVERED. `decodeGuarantee`/`reallocateGuarantee` in `Outcome.lean`.

### S2-40 — Guarantee funds move only after the target is finalized
- Class: (a) on-chain safety.
- Quote (1072–1074): "The funds held in a guarantee channel can only be paid out
  once the outcome of the target channel is finalized."
- Property: a claim fails unless the target allocation channel is finalized.
- Coverage: PARTIAL. `reclaim_can_happen` concludes both the source and the target
  channel before the reclaim. No single theorem states that a claim with a
  non-finalized target has no effect.

### S2-41 — Guarantee format: listed addresses move to the top
- Class: (d) definition.
- Quote (1094–1101): "the guarantee provides a list of addresses, with the rule
  that these addresses will be moved to the top of the priority order if they
  appear in the outcome. For example... (L|A, B, C) and the outcome of L is
  (C : 1, D : 2, A : 4)... pay out as though the outcome was (A : 4, C : 1, D : 2)."
- Property: a guarantee reorders the target outcome by moving its listed
  destinations, in order, to the front.
- Coverage: COVERED. `reallocateGuarantee` implements the reprioritization.

### S2-42 — Definition: "G can afford x for B"
- Class: (d) definition.
- Quote (1109–1111): "a guarantee 'G can afford x for B', if B would receive at
  least x coins, were the coins currently held by G to be paid out according to
  G's prioritization of its target."
- Property: guarantee affordability uses the guarantee's reprioritized order.
- Coverage: COVERED. `afford` applied through `reallocateGuarantee`.

### S2-43 — Claim effects when affordable
- Class: (a) on-chain safety.
- Quote (1139–1145): "Nitro adds the claim operation, C(G, A, x)... If G acts as
  guarantee for L and can afford x for A, then C(G, A, x) has the following three
  effects: • Reduces the funds held in channel G by x. • Increases the funds held
  in channel A by x. • Reduces the amount owed to A in the outcome of L by x.
  Otherwise, the claim operation has no effect."
- Property: a successful claim debits G by x, pays destination A x, and cuts A's
  entry in L's outcome by x; an unaffordable claim has no effect.
- Coverage: CONTRADICTED / PARTIAL. The model op is `reclaim`, with a different
  signature: `(sourceChannelId, sourceOutcome, sourceAssetIndex,
  indexOfTargetInSource, targetOutcome, targetAssetIndex, claimedTargetStateHash)`
  (`Operations.lean:79–81`). The model `reclaim` reallocates the guarantee source
  from the target's finalized outcome via `reallocateGuarantee`; it does not take
  a named destination A and an amount x the way `C(G, A, x)` does. See the
  cross-source note: Nitro 2019 "claim C(G,A,x)" is not the current "reclaim".

### S2-44 — Nitro redistribution is order dependent
- Class: (a) on-chain safety.
- Quote (1147–1151): "it is possible to construct situations where the same
  outcome can lead to different values, depending on the order in which guarantees
  are claimed."
- Property: with multiple guarantees, the claim order can change who gets paid.
  (This contrasts S2-36, the Turbo order-independence fact.)
- Coverage: MISSING. The model states no order-dependence fact for Nitro claims.

### S2-45 — Lower-bound guarantee on distribution
- Class: (b) unbeatable strategy.
- Quote (1152–1177): "putting some lower bounds on how funds are distributed...
  participant I will receive their 5 coins, even though we cannot say anything
  about how the remaining 5 coins will be distributed... Nitro Lower-bound Value
  Algorithm... Minimum Payout Calculation".
- Property: each destination has a guaranteed minimum payout, computed by the
  minimum-payout calculation over all claim orders.
- Coverage: MISSING. No lower-bound value algorithm in the model.

### S2-46 — Overfunded and fully funded cases pay every destination
- Class: (b) unbeatable strategy.
- Quote (1216–1218): "If vi > aj, then we say the system is overfunded. In this
  case, we know that all destinations will receive their allocations, pi = ai
  regardless of the order of payout." Also (1251–1253): "In the fully funded case,
  where vi = ai... pi = ai, just as in the overfunded case."
- Property: when guarantees cover the allocation, every destination receives its
  full allocated amount.
- Coverage: PARTIAL. The directly-funded full-payout case is covered by
  `unbeatable_pays`/`payout_can_happen`. The guarantee overfunded/fully-funded
  statement is not stated.

### S2-47 — A or B can offload a virtual channel
- Class: (b) unbeatable strategy.
- Quote (1299–1316): "A then calls T(L, G, x) to move the funds from L to G...
  Finally A calls C(G, A, χ)... As G has χ as top priority, the operation is
  successful. By symmetry, the previous case also covers the case where B wants to
  offload."
- Property: either end participant can finalize their channels, transfer to the
  guarantee, and claim the virtual channel into direct funding.
- Coverage: PARTIAL. `payout_can_happen`/`reclaim_can_happen` show the single ops
  succeed; the full offload sequence is not stated as one theorem.

### S2-48 — The intermediary can reclaim its collateral
- Class: (b) unbeatable strategy.
- Quote (1317–1343): "the case where I wants to offload the channel and reclaim
  their funds. This is important to ensure that A and B cannot lock I's funds
  indefinitely... I has to claim on both guarantees, offloading χ before being
  able to reclaim their funds. The virtual channel became a direct channel and the
  intermediary was able to recover their collateral."
- Property: the intermediary always has a strategy to recover its locked
  collateral; other parties cannot lock it forever.
- Coverage: MISSING. No model theorem states intermediary liveness.

---

## §4.5–4.6, §5.3–5.4, §6.3–6.5 off-chain constructions

### S2-49 — Off-chain protocol constructions and rules
- Class: (e) off-chain protocol property. Listed, not dropped.
- Quotes: consensus game (660–688); Outcomes First (694–709); value-preserving
  single-channel updates and the Simple Transition Rule (500–524); ledger
  sub-channel open/close, top-up, partial withdrawal (891–1047); virtual channel
  construction, opening, closing, three-party (1257–1455).
- Property: these specify participant behavior and off-chain state sequences, not
  on-chain adjudicator rules. The adjudicator must only support the on-chain ops
  these constructions call (deposit, transfer, claim, conclude, challenge).
- Coverage: N/A for on-chain theorems. The supporting on-chain ops are covered by
  the can-happen theorems (S2-04, S2-12, S2-13, S2-15, S2-37, S2-47).

---

## Cross-source disagreements noticed

- Operation name and shape: Nitro 2019 (this source) names the guarantee
  redistribution op "claim", `C(G, A, x)` (lines 1052, 1139–1145). It pays a named
  destination A an amount x from guarantee G. The model's op is `reclaim` with a
  source-channel/target-channel signature (`Operations.lean:79–81`) that
  reallocates the guarantee from the target's finalized outcome. The task brief
  flags "Nitro 2019 `claim` versus SATP `reclaim`". The model follows the later
  `reclaim` shape, not the 2019 `claim` shape. This is S2-43's CONTRADICTED mark.
- Order dependence: Nitro 2019 states Turbo transfers are order independent
  (S2-36, line 842) but Nitro claims are order dependent (S2-44, lines 1147–1151).
  A model that treats both the same would contradict one of these two facts.

---

## Class counts (49 properties)

- (a) on-chain adjudicator safety: 13 — S2-05, S2-06, S2-07, S2-08, S2-11, S2-14,
  S2-22, S2-28, S2-29, S2-36, S2-40, S2-43, S2-44.
- (b) on-chain liveness / unbeatable strategy: 10 — S2-16, S2-21, S2-30, S2-34,
  S2-35, S2-37, S2-45, S2-46, S2-47, S2-48.
- (c) can-happen: 4 — S2-04, S2-12, S2-13, S2-15.
- (d) definition the model must match: 16 — S2-01, S2-02, S2-09, S2-10, S2-17,
  S2-18, S2-23, S2-24, S2-25, S2-26, S2-27, S2-33, S2-38, S2-39, S2-41, S2-42.
- (e) off-chain protocol property: 1 — S2-49.
- (f) assumption the source makes: 5 — S2-03, S2-19, S2-20, S2-31, S2-32.

## Coverage summary

- COVERED: S2-01, S2-02, S2-03, S2-04, S2-08, S2-09, S2-10, S2-11, S2-12, S2-13,
  S2-15, S2-16, S2-17, S2-21, S2-26, S2-27, S2-28, S2-30, S2-35, S2-38, S2-39,
  S2-41, S2-42.
- PARTIAL: S2-05, S2-14, S2-18, S2-19, S2-20, S2-23, S2-29, S2-34, S2-37, S2-40,
  S2-46, S2-47.
- MISSING: S2-06, S2-07, S2-22, S2-24, S2-25, S2-31, S2-32, S2-33, S2-36, S2-44,
  S2-45, S2-48.
- CONTRADICTED: S2-43.
- N/A (off-chain): S2-49.
