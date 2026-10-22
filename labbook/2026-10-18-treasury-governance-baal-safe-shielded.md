# Baal governance and a shielded treasury for Armada

Lab book entry, 2026-10-18 to 2026-10-19. Session `01a150cb-a1b2-727e-87fc-e762323ef56d`
(20:55 to 12:47). Repos affected: `armada-poc` (governance contracts, the ARM token,
`ShieldedRagequit`, RevenueLock), `armada-circuits` (the contract-owned note circuit),
`armada-nitro` (the settlement rail, one shared gate), `armada-laconic` (the ADR).

Revision, 2026-10-22 (session `01a156a6`): §3a makes the treasury a public pool balance,
removes every crowdfund code path, and fixes the ragequit divisor. It replaces decisions 6,
12, 13, 14, 15, 16 and 20, and makes 17 and 18 mechanical. The replaced decisions stay in
place, marked.

## 1. The question

The operator asked the session to study MolochDAO and its successors. The operator then
asked which parts fit Armada. The real question grew from there. The current Armada
governance in `armada-poc` locks holders in until a 2027 wind-down. One large proposal can
drain the treasury first. The question became: can a Moloch v3 (Baal) treasury replace the
custom governor, keep a fixed ARM supply, keep the treasury private, and still let any
holder exit at any time?

## 2. Evidence

### What the session read or measured

- **MolochDAO and Baal.** Baal (Moloch v3) uses a Safe as the treasury. Baal is a module on
  the Safe. The deployed v3 is the `feat/baalZodiac` branch. That branch has no "process
  your Yes votes first" rule. `ragequit(to, shares, loot, tokens[])` works at any time.
  `minRetentionPercent` fails a proposal if too many members exit before processing. Quorum
  counts Yes votes only. Audits: Hacken 2022-08-02 and Quantstamp 2022-10-18. The last push
  to the repo is 2024-06-11, which is after both audits (session reply, 21:14).
- **Baal's payout formula.** Baal's `_ragequit` pays `burned × treasuryBalance /
  sharesSupply` for each listed token (session replies, 21:24 and 21:40).
- **The governance-capture finding.** `armada-launch-audit/findings/governance-capture.md`
  states that one Extended proposal drains the treasury in 47–56 days. It states that
  holders cannot exit before the 2027-12-31 wind-down deadline (session replies, 21:40).
- **The ARM token.** `ArmadaToken.sol` `_beforeTokenTransfer` permits only whitelisted
  *senders* while `transferable` is false. ARM is fixed supply and has no burn. ARM uses OZ
  4.9 `ERC20Votes`, which counts votes by block number. Baal calls `getPastVotes` with a
  timestamp (session reply, 21:24).
- **RevenueCounter and RevenueLock.** `RevenueCounter.sol:121` and `:145`: the owner can
  call `addRevenue` and `attestRevenue` for free. `RevenueCounter` is UUPS and `Ownable`.
  `RevenueLock` is about 434 lines. About half of that defends a free revenue source with a
  `MAX_REVENUE_INCREASE_PER_DAY` ratchet and a wind-down freeze (session reply, 22:00).
- **The pool circuit.** `Transact.circom:156–164`: a spend needs the note in a committed
  tree, a Merkle path proof, and an EdDSA signature by the owner's spending key.
  `Transact.circom:15`: nullifiers are public inputs. A contract holds no secret key, so a
  contract cannot sign such a spend (session replies, 21:44 and 22:19).
- **Selective disclosure.** `16-selective-disclosure.md` line 37: each spend carries an
  EdDSA signature by the spending key (session reply, 21:44).
- **armada-nitro.** `ConsensusApp.sol:18–26` requires every participant's signature.
  `ForceMove.sol:206–212` lets the app decide what counts as a supported state.
  `ArmadaNitroAdjudicator.sol:71–79`: `deposit` is public.
  `ArmadaNitroAdjudicator.sol:131–160`: a `SHIELD` allocation pays a note on finalization.
  Page 15 (T4.3): protocol fees on swaps are an allocation on the channel state (session
  replies, 21:51 and 03:16).
- **Linear unlock contracts.** Safe has no built-in linear unlock. Hedgey `TokenLockupPlans`
  is a lockup with no revocation, audited by Consensys Diligence 2023-06. Sablier Lockup
  Linear needs the deployer to set `cancelable = false`. The Zodiac Exit module is audited
  2021, 2022, and 2022-11 (session replies, 21:24 and 21:29).
- **Prior session.** Session `01a140be` (2026-10-15) discussed a treasury spending key held
  as a t-of-n threshold EdDSA key on BabyJubJub (session reply, 21:48).

### Operator rulings, verbatim, with timestamps

- 2026-10-18 21:12:18 — "nouns lido and majeur add nothing that we'd need for armada. I
  don't think we'd need shamans either.  otherwise, it looks like v3 is a better treasury
  system than the one we currently have."
- 2026-10-18 21:19:51 — "Update: we want to keep some features, like fixed supply of ARM
  tokens, but have the treasury which holdes the crowdfunded ETH and the remaining ARM
  governed by Baal, with a minimal set of shamen."
- 2026-10-18 21:21:32 — "the ARM held in the treasury should simply be excluded. and it
  should be able to have a linear unlock of tokens, but I assume Safe already supports this.
  and yes, it's USDC not ETH, sorry for my confusion there."
- 2026-10-18 21:27:17 — "a linear unlock is a timelock, vesting is a conditional unlock,
  those are the definitions of the words, wtf do you think they mean?"
- 2026-10-18 21:39:26 — "Let's go with your recommendation that ARM is a Baal share, is that
  also an ERC20?"
- 2026-10-18 21:39:49 — "We'd want to disable upgrades."
- 2026-10-18 21:44:04 — "authorizeadapter and setfeecollector would need to be voted on by
  ARM(baal share holders) via the Safe. The safe, which is the tresury, will also need to
  take shielded notes, to fix the attack where permissionless deposits taint the tresury."
- 2026-10-18 21:47:09 — "the safe doesn't need to hold the note per se, but the treasury
  needs to be shielded and governed by an ERC20."
- 2026-10-18 21:58:36 — "2 is the only option that makes sense. If a court can order enough
  ARM holders to send funds to some address, so be it."
- 2026-10-18 21:59:24 — "the only option is to have quarantined funds left where they are,
  or moved into the treasury channel. KISS."
- 2026-10-18 22:16:46 — "okay here's the problem you seem confused about. all note spending
  needs to be resolved. there can't be some pool of pending notes not in the governed
  channel. Repeat this requirement to cinfirm you understand."
- 2026-10-18 22:18:27 — "no. they must also be shielded. so the idea is that the time
  between a shielded note and channel value is minimized. repeat to confirm understanding"
- 2026-10-18 22:30:13 — "\"Pool-level blacklist risk is shared\" and \"No scanning\" are
  irrelevant. we are using selective disclosure circuits in place of POI. Update your
  respones based on the above."
- 2026-10-18 22:48:00 — "3. oh. you need to call that shieldedragequit you screwball."
- 2026-10-21 14:03:26 — "this owner of the safe is the baal contract. confirm."
- 2026-10-21 14:12:41 — "let's remove 2 as a separate codepath for the crowdfund, baal won't
  be participating in the crowdfund."
- 2026-10-21 14:16:39 — "1. this would cause a lot of problems and isn't in the existing
  specs? 2. through step 3. as I just mentioned. 3. probably. if it actually works. The
  crowdfund is already deployed baal doesn't need to have hard coded support it for it.
  that's an improper separation of concerns."
- 2026-10-21 14:20:35 — "yes. make those changes and the pool needs to reject naked erc20
  deposits, is that possible"
- 2026-10-21 14:22:30 — "the mark and sweep is the stuck funds discussion, it's the same
  thing caused by the same issue. confirm."
- 2026-10-22 07:24:49 — "2. Yes. write the entries. 3. Yes. this needs to be well
  documented." Item 3 answered: "does 'excluded' mean the payout divisor is `totalSupply −
  Safe − Hedgey lockup`, read before the burn?"

## 3. Findings and decisions

Each decision is one sentence. Each traces to the evidence or a ruling above.

1. Armada adopts Moloch v3 (Baal) on a Safe, and it drops the custom governor, wind-down,
   redemption, treasury-governor, steward, and veto, because exit replaces rule-based
   defense (ruling 21:12:18; governance-capture finding).
2. Armada keeps no Nouns fork, no Lido dual governance, and no Majeur feature (ruling
   21:12:18).
3. ARM is the Baal shares token itself, an ERC20, with 12M minted once, `mint` that reverts,
   and `burn` used only by ragequit, so there is no wrapper and no second token (rulings
   21:39:26 and 21:19:51; the overpay analysis at 21:35 and 21:37).
4. ARM deploys as a plain contract, not a proxy, and it drops `UUPSUpgradeable` and
   `Initializable`, so the token is not upgradeable (ruling 21:39:49).
5. ARM needs a transfer allowlist fixed at construction for RevenueLock, the Hedgey lockup,
   and the crowdfund, because shares stay paused until a proposal unpauses them for listing
   (session replies, 21:24 and 22:00).
6. Treasury ARM sits in the Safe and in the Hedgey lockup, and neither address ever
   ragequits, so that ARM carries no claim and needs no exclusion list (ruling 21:21:32;
   session reply, 21:40). **Replaced by §3a T7: treasury ARM is excluded from the payout
   divisor (2026-10-22).**
7. A linear unlock releases the reserved ARM, and the session uses Hedgey `TokenLockupPlans`
   because Safe has no built-in unlock and Hedgey cannot be configured wrong (ruling
   21:27:17; session reply, 21:29).
8. Baal handles the token-weighted votes (session reply, 22:37; "Yes").
9. The Safe is the treasury, its only owner is the Baal contract, and its threshold is 1, so
   no human signer makes a second spending path (session replies, 22:36 and 22:37).
10. `authorizeAdapter` and `setFeeCollector` become ordinary Baal proposals, so ARM holders
    vote and the Safe executes after the grace period (ruling 21:44:04).
11. Pool users are not ARM holders, so their only protection against an adapter change is the
    delay `votingPeriod + gracePeriod`, which must be long enough for them to react (session
    reply, 21:44).
12. The treasury is shielded, and treasury value enters as a shielded note and becomes
    governed value in the same transaction, so no treasury note stays outstanding (rulings
    21:44:04, 21:47:09, 22:16:46, and 22:18:27). **Stays; §3a T1 and T2 give its form: the
    value is credited to `balance[Safe][token]` in the same transaction (2026-10-22).**
13. A contract-owned `Transact` variant makes this possible: the `npk` derives from the
    owning contract's address, the nullifying key is public, and the EdDSA check is replaced
    by a binding to `msg.sender` (session reply, 22:19). **Replaced by §3a T1 (2026-10-22).**
14. The Safe owns the notes directly, so the design deletes the treasury channel, the
    custodian, `RagequitApp`, and both threshold DKGs, which makes the design the most
    parsimonious shape (session replies, 22:22 and 22:24). **The Safe holds no notes; §3a
    T1 (2026-10-22). The deleted channel, custodian, `RagequitApp` and DKGs stay deleted.**
15. Fees are shielded into Safe-owned notes, and `ShieldModule` creates the fee note to the
    Safe's `npk` and increments `cumulativeFees` in the same permissionless call (ruling
    21:44:04; session replies, 22:19 and 22:22). **Replaced by §3a T2(a) (2026-10-22).**
16. Deposits stay permissionless: anyone spends their own note into a Safe-owned note, and
    the spend proof carries selective disclosure, which is the taint control (ruling
    22:30:13; session replies, 22:19 and 21:55). **Replaced by §3a T2(b) (2026-10-22).**
17. A raw ERC-20 `transfer` to a contract credits no note and enters no outcome, so tainted
    dust is stuck, not mixed (session reply, 21:55). **§3a T4 gives the mechanism
    (2026-10-22).**
18. Quarantined funds have two states only: a Baal proposal shields them into the treasury,
    or it leaves them where they are, with no refund path and no burn path (ruling 21:59:24;
    session reply, 21:59). **§3a T4 gives the mechanism (2026-10-22).**
19. A court order is handled by ARM holders voting to send funds to an address, which the
    design accepts as the only sensible path (ruling 21:58:36).
20. Ragequit is a contract named `ShieldedRagequit`, which burns the holder's ARM and pays
    `burned / supply × treasury` as a note to the holder's `npk`, both a Baal manager shaman
    and a Safe module (ruling 22:48:00; session reply, 22:22). **Replaced by §3a T6 and T8
    (2026-10-22).**
21. `ShieldedRagequit` is the one shaman, which satisfies the "minimal set of shamans"
    constraint (ruling 21:19:51; session reply, 22:22).
22. RevenueLock keeps its purpose but reads `cumulativeFees` from the pool, so its ratchet
    and wind-down freeze go, and every unlock corresponds to real USDC that ARM holders own
    (session reply, 22:00).
23. Governance does not depend on armada-nitro, and the one shared edge is N8, which proves
    that a swap fee reaches RevenueLock (session replies, 03:16 and 03:17).

## 3a. Revision of 2026-10-22: the treasury is a pool balance

Session `01a156a6` (2026-10-21 to 2026-10-22). The sibling entry's §3a (R1 to R5) makes a
contract owner's holding a public per-token balance inside the pool. For the treasury, that
gives the decisions below. The merged code that these decisions correct is listed after
them.

T1. The treasury is `balance[Safe][token]` inside the pool. The Safe holds no notes, and
    Baal is the Safe's only owner (decision 9; ruling 2026-10-21 14:03:26). Ruling 21:44:04
    ("the safe ... will also need to take shielded notes") and ruling 21:47:09 ("the safe
    doesn't need to hold the note per se, but the treasury needs to be shielded") hold in
    this form: treasury value never sits as an ERC-20 balance at the Safe's address, and every
    payment into it comes from a shielded payer, through a spend that carries selective
    disclosure.
T2. Value reaches the treasury in two ways only.
    (a) Shield fees. Each shield's fee credits `balance[Safe][token]` and
        `cumulativeFees[token]` in the same transaction.
    (b) Any payment. A payer's private `transact` ends in an unshield whose destination is
        `balance[Safe][token]` (sibling entry R2). Sale proceeds, the USDC leg of a swap with
        the Safe, and a gift all take this path.
    Only (a) counts toward `cumulativeFees`, so only real fees release ARM from RevenueLock
    (decision 22).
T3. The crowdfund has no code path into the pool or into governance. It is already deployed,
    and Baal does not take part in it. Its proceeds reach the treasury only as a payment
    under T2(b) (rulings 2026-10-21 14:12:41 and 14:16:39: hard-coded crowdfund support is an
    improper separation of concerns).
T4. Raw ERC-20 transfers and airdrops to the pool cannot be refused. They form the pool's
    surplus (sibling entry R4). A Baal proposal sweeps the surplus into `balance[Safe][token]`,
    or leaves it where it is ("mark and sweep"). This is decisions 17 and 18 in mechanical form
    (ruling 2026-10-21 14:22:30).
T5. A governance spend is a Baal proposal that the Safe executes. The pool sees
    `msg.sender == Safe`, debits `balance[Safe][token]`, and creates the recipient's note. No
    fee applies. This extends decision 10 to every spend.
T6. Ragequit. A holder calls `ShieldedRagequit` with an ARM amount and an `npk`. The contract:
    (a) reads the ARM supply and the divisor (T7) before the burn;
    (b) as Baal's shaman, burns the holder's ARM;
    (c) for each token, computes `payout = burned × balance[Safe][token] / divisor`;
    (d) as a Safe module, makes the Safe call the pool through `execTransactionFromModule`,
        so that the pool sees the Safe as the caller, debits the payout from
        `balance[Safe][token]`, and creates a note for it at the holder's `npk`.
    No fee applies. The exit needs no treasury signature and no vote, so any holder can exit
    at any time (decision 1).
T7. Treasury ARM is excluded from the divisor: `divisor = supply − ARM held by the Safe − ARM
    held by the Hedgey lockup`, all read before the burn. That ARM never ragequits and carries
    no claim, so counting it would let the other holders together claim only part of each
    treasury asset (ruling 21:21:32, "the ARM held in the treasury should simply be
    excluded"; confirmed as this divisor on 2026-10-22 07:24:49).
T8. The Safe has exactly two modules, Baal and `ShieldedRagequit`. `ShieldedRagequit` may
    make the Safe call only the pool payout of T6(d): no other function and no other target.

Merged code that these decisions correct (armada-poc `integrated`, 2026-10-21):

- `ArmadaCrowdfund.finalize` shields the proceeds to `pool.feeRecipient()` (lines 566-583).
  That is a crowdfund code path (against T3), and `_recognizeFee` counts the whole sale in
  `cumulativeFees` (against T2).
- No production deploy script calls `setFeeRecipient`; only
  `test-foundry/helpers/CrowdfundPoolDeploy.sol` does. That is a second, test-only deploy path.
- `ShieldedRagequit` reads one caller-named note through `preimageOf`, requires the raw
  address as `npk`, which the P1 circuit never produces, and calls `transactOwned` itself, so
  the pool sees the shaman, not the Safe, as the caller (against T1, T6 and T8).
- `PrivacyPool.verify` and `verifyOwned` return `true` in `testingMode`, and 12 test files
  turn it on (against sibling entry R8).
- The crowdfund window change (`arp-dda`, `arp-29f`) and the G7a lifecycle (`arn-5c3`) exist
  only to run the crowdfund inside the governance gate (against T3).

## 4. The check that would show each decision false, or enforce it

These checks come from the gate chunks of the 209-point list (03:17). The entry names them.
It does not build them.

- **G1 tests** show ARM is non-upgradeable and has the transfer allowlist. A test that mints
  past 12M, or that upgrades the token, must fail. This enforces decisions 3, 4, and 5.
- **G2 fixturenet deploy** summons Baal and the Safe and wires the module. A deploy where the
  Safe has a human owner, or where the token is a proxy, must fail. This enforces decisions
  1, 4, and 9.
- **G3 tests**, on the production deploy with real proofs. A test where the exit pays anything
  other than `burned × balance[Safe][token] / divisor` read before the burn (T6, T7), where the
  pool sees a caller other than the Safe, where the module makes the Safe call anything but the
  payout (T8), or where the exit needs the treasury's signature, must fail. This enforces
  decisions 1, 12, T6, T7 and T8.
- **G4 tests** rewrite RevenueLock on `cumulativeFees`. A test that inflates the counter for
  free, or that releases ARM with no matching fee, must fail. This enforces decision 22.
- **G5 tests** show the Hedgey lockup. A test where crowdfund proceeds raise `cumulativeFees`,
  or where any pool or governance code references the crowdfund, must fail. This enforces T2
  and T3.
- **G7 gate** runs payment → vote → spend → ragequit on the fixturenet: ARM holders, a payment
  to the Safe under T2(b), a Baal vote, a governance spend (T5), and a ragequit (T6). It does
  not use the crowdfund. A run where a captured majority blocks an exit, or where a holder
  cannot exit at any time, must fail. This enforces decisions 1, T2, T5 and T6.
- **Pool invariant tests** (sibling entry R5) show that `balance[Safe][token]` equals its public
  credits minus its debits, and that a raw transfer changes only the surplus. This enforces T1
  and T4.
- **N8 gate** proves in-channel swap fee → Safe-owned note → `cumulativeFees` → RevenueLock
  release. A run where a swap fee never reaches RevenueLock must fail. This enforces decisions
  22 and 23.

## 5. Rejected options

- **The current `armada-poc` governance** (governor, treasury-governor, steward, wind-down,
  redemption, the veto). Rejected because exit replaces rule-based defense, and because the
  governance-capture finding shows the old defenses fail (ruling 21:12:18; session reply,
  21:40).
- **A Nouns fork, Lido dual governance, Majeur features, and extra shamans.** Rejected
  because they add nothing Armada needs (ruling 21:12:18).
- **An ARM wrapper plus a separate Baal shares token plus a burn gate.** Rejected because the
  operator chose ARM as the Baal share directly, which removes the overpay bug and the
  custom code (ruling 21:39:26; the overpay worked example, 21:35).
- **Mandatory wrap with stock Baal ragequit** (idx option 1, 21:35). Rejected because it
  makes the fixed supply only nominal, and the direct ARM-as-share design is simpler (ruling
  21:39:26; session reply, 21:37).
- **The Zodiac Exit module and Sablier.** Superseded by the in-pool `ShieldedRagequit` and by
  Hedgey, which cannot be configured wrong (session replies, 21:29 and 22:22).
- **A treasury channel with a federation custodian, a BabyJubJub threshold signer, and a
  `RagequitApp`.** Rejected because the Safe owns the contract-owned notes directly, which
  removes the channel, the custodian, and both DKGs (ruling 22:30:13; session replies, 22:19
  and 22:22).
- **A gated `deposit`, a `depositFee` entry, a `totalHoldings` counter, a `sweep`, and a
  separate quarantine Safe.** Deleted with the Safe-owned-notes design; the treasury Safe is
  the quarantine, and a proposal shields raw transfers or leaves them (ruling 21:59:24;
  session reply, 22:22).
- **Stuck-funds options 1, 3, 4, 5, and 6** (reintroduce taint, a quarantine channel, return
  to sender, burn, shield into the pool). Rejected on 21:56; option 2 then simplified to
  "left alone or moved into the treasury" (ruling 21:59:24).
- **A refund path and a burn path for quarantined funds.** Removed by the KISS ruling; only
  "left alone" or "moved into the treasury" remain (ruling 21:59:24; session reply, 21:59).
- **N-of-M human Safe owners.** Rejected as default, because human owners are a second
  spending path; the Safe owner is Baal only, threshold 1 (session replies, 22:36 and 22:37).
- **Safe-owned notes for the treasury** (decisions 13 and 14). Replaced on 2026-10-22 by a
  public pool balance (§3a T1). A payer cannot merge value into a note the Safe owns, so notes
  left the treasury split across many outputs, some with no recorded preimage.
- **A crowdfund code path into the pool or governance** (the merged G5 `finalize`). Rejected
  as an improper separation of concerns (ruling 2026-10-21 14:16:39).
- **Treasury ARM in the payout divisor** (decision 6). Rejected; §3a T7.
- **An inequality pool invariant** ("tokens ≥ unspent notes + balances"). Rejected; it hides a
  leak. The invariants are exact (sibling entry R5).

## 6. Open questions

- Is it acceptable to place the whole pool's solvency behind one more Groth16 ceremony and
  one more verifying-key authority, so the treasury can stay in-pool? The session named this
  as the item that decides whether the design is acceptable. The operator did not rule on it
  (session reply, 22:34).
- If the team wants N-of-M human Safe owners for some purpose, how is their power scoped, for
  example through a Zodiac Roles module, so they cannot move treasury notes (session reply,
  22:37)?
- Should a fee be charged only on the selective-disclosure-gated side, so taint control holds
  for fees as well as for deposits (session reply, 22:15)?
- Do the pool contracts live in `armada-poc/contracts`, or must a third pebble tracker and
  repo init cover them? [INFERENCE] The plan lists this as a prerequisite to confirm (plan,
  03:31).

## 7. The chunks this entry covers

Copied from the 209-point list (2026-10-19 03:17).

| # | Chunk | Depends on | Pts |
|---|---|---|---:|
| G1 | ARM shares token, non-upgradeable, transfer allowlist, with tests | — | 6 |
| G2 | Baal + Safe summon script, module wiring, fixturenet deploy | G1 | 8 |
| G3 | `ShieldedRagequit` contract with tests, including the exit proof built with the SDK | G2, P2 | 9 |
| G4 | RevenueLock rewrite on `cumulativeFees`, with tests | G1, P2 | 6 |
| G5 | Hedgey lockup, crowdfund delivers ARM shares, crowdfund proceeds shielded to the Safe | G1, P2 | 8 |
| G6 | Remove old stack, deploy scripts, `GOVERNANCE.md` and `ARM_TOKEN.md` rewrite | G2, G5 | 8 |
| G7 | Gate: crowdfund → claim → vote → spend → ragequit on the fixturenet, shield fees only | G3–G6 | 10 |
| D1 | Governance ADR, contract-owned-notes ADR | — | 5 |

The Governance subtotal is 55 points. D1 is 5 points inside the Documents subtotal of 11.
Governance-only work stays independent of nitro. The one shared edge is N8.
