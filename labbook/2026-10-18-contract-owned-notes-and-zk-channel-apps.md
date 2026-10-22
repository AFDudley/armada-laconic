# Contract-owned notes and ZK channel apps

Lab book entry. Session 2026-10-18 20:55 UTC to 2026-10-19 12:47 UTC.
Session id `01a150cb-a1b2-727e-87fc-e762323ef56d`.
Repos affected: `armada-laconic` (docs), `armada-circuits` (pool circuits),
`armada-poc` (pool contracts), `armada-nitro` (adjudicator), `armada-sdk`,
`armada-client` and `laconic-wallet` (client).

This entry records the pool and armada-nitro decisions. A sibling entry records
the governance decisions. The two share the pool as a dependency.

Revision, 2026-10-22 (session `01a156a6`): §3a replaces decisions 1, 2, 3, 9 and
11 for contract owners. A contract owner holds a public per-token balance inside
the pool, not notes. The replaced decisions stay in place, marked, so the history
stays readable.

## 1. The question

The session asked one design question in two halves. First: can a Safe own
shielded notes, so that the treasury spends inside the pool and holds no public
ERC-20 balance? Second: does the same contract-owned note help armada-nitro, and
can a channel hide its balances with a ZK ForceMove app? The session then
measured the code that exists, sized the work in points, and wrote an execution
plan.

## 2. Evidence

### What the session read or measured

- `Transact.circom` already has the hook for a contract-owned variant. Line 158
  is `eddsa.enabled <== 1`. A variant sets `enabled` from a public flag and
  derives the note keys from a public `owner` address (session reply 22:26 UTC).
- Nullifiers are public already. Every spend publishes its nullifiers on-chain as
  public inputs (`Transact.circom:15`). The private value is the `nullifyingKey`
  (session reply 22:32 UTC).
- ADR-0003 governs the ceremony. D3.2 reuses the community Perpetual Powers of
  Tau for Phase-1 and runs an own Phase-2 MPC. D3.4: "Phase-2 is re-run **only**
  when circuits change". A.3.3 row 4: "new ceremony required on any circuit
  change". A.3.4.1: about 54 to 55 `(nIn, nOut)` combos. The cost is "a few days
  of MPC" (session reply 22:38 UTC).
- A.2.7 defers the admin key: "Owning the upgrade admin concentrates upgrade,
  fee, and vkey authority; key management and governance for the admin is an
  operational risk … out of T0.0 code scope." A.3.2 calls key registration "a
  T0.0 governance action" (session reply 22:45 UTC).
- A.2 is an `armada-laconic` document. The author is A. F. Dudley, September
  2026. A.2.2 lines 18 to 19 say "Author and deploy the upgradeable proxy to
  `RailgunSmartWallet`" and "Own the proxy admin / owner keys" (session reply
  22:48 UTC).
- `IForceMoveApp.stateIsSupported(fixedPart, proof, candidate)` is the only place
  game logic lives in nitro. ForceMove keeps turn numbers, signatures, challenge
  windows and finalization in Solidity (session reply 23:18 UTC).
- `armada-nitro/src` holds 2,390 lines at `f720847` (2026-10-18). `IForceMoveApp`
  is 22 lines. `ConsensusApp` is 26 lines. `MultiAssetHolder` is 546 lines.
  `StrictTurnTaking` already encodes the mover-signs-own-turn rule (session reply
  23:32 UTC).
- Page 14 line 16 read "First games: Aave USDC lending and borrowing, and
  swapping shielded WETH for wstETH." Page 15 line 101 read "Borrowing. A borrow
  needs liquidation rules and an on-chain price, which makes it a real ForceMove
  app (option 1). It stays v2." (session reply 23:45 UTC).
- Prior-art search returned no single source with all four properties. The
  nearest single match is Celer cChannel (2018): `IntendSettle` accepts a state
  proof "if the input is newer", a recency-only window, with no SNARK. Perun
  (2018) and Nitro have the generic adjudicator with no proof. zkChannels/Bolt,
  zk-PCN, Intmax2 and Dark Forest carry a proof per state but are not a generic
  channel app. The earlier mention of "zkHoldem" is unverified (subagent
  `ZkChannelPriorArt`, 41 queries, reply 23:39 UTC).
- Repository state (12:30 to 12:33 UTC). `armada-circuits` has a fork under
  AFDudley, branch `disclosure-circuit`. `armada-sdk` has a fork, branch
  `selective-disclosure`, `origin` is ship-armada. `armada-nitro` is the AFDudley
  origin, branch `integrated`, with `.pebbles` prefix `arn`. `armada-poc` had no
  fork remote; the session added a `fork` remote to the existing
  `AFDudley/armada-poc`. `fork/main` is at `102f17b8`; `origin/main` is at
  `e78b1f73`; the fork is ahead of upstream. Trackers still need init on
  `armada-poc`, `armada-circuits` and `armada-sdk`.
- Mobile state (12:31 UTC). `armada-client` has zero lines of application code.
  P1-01 merged a Node 18.20.4 toolchain and an empty npm-workspaces skeleton.
  P1-02 and P1-03 worktrees exist from 2026-09-25 with no package code.
  `laconic-wallet` holds 5,322 lines. `armada-interface` holds 37k lines (web,
  not mobile). `armada-sdk` holds 13k lines.
- All 23 circuit files in `armada-circuits` are circom (12:42 UTC). The repo pins
  `circomlib ^2.0.5`. The own circuit code is 334 lines: `Transact.circom` 167,
  `NullifierDisclosure.circom` 112, `merkle.circom` 55, plus 19 `main_NxM`
  wrappers. Railgun builds on `circomlib` the same way (12:47 UTC).
- Context scale (03:20 to 03:28 UTC). One point is about 10,000 tokens of peak
  context (`exophial/src/exophial/estimate.py`). The 27B model on the 5090 holds
  163,200 tokens, so the dispatcher refuses any pebble above 16 points. A cloud
  worker starts at about 86,000 tokens before its first tool call. A pilot pebble
  on marks took 923 s over 62 turns. The 209-point list is about 27 sessions of
  work.

### Operator rulings, verbatim

- 2026-10-18T22:30:13Z — "Pool-level blacklist risk is shared" and "No scanning"
  are irrelevant. we are using selective disclosure circuits in place of POI.
  Update your respones based on the above.
- 2026-10-18T22:48:00Z — 1. The requirement is that armada be largely NOT
  upgradable. whose doc is a.2.7 ours (a.f.dudley/laconic) or armada's? 2. I'm
  having a difficult time reasoning about what happens if smart wallets and just
  programs can use notes. 3. oh. you need to call that shieldedragequit you
  screwball.
- 2026-10-18T23:16:16Z — You're confused. a complete proof for each state change
  is the worst of both worlds, it's a horrible design. Do you see why?
- 2026-10-18T23:44:40Z — I don't get how you get so confused about this. what
  fucking product decision wasn't made? I see this is a case where some agent
  made up some bullshit and I didn't argue with it. we need hubs to take aave
  positions, providing liquidity and then swapping the waUSDC or whatever in
  private. that's the only plan this other thing is some weird shit you made up
  that just doesn't make sense. confirm.
- 2026-10-18T23:46:41Z — yes make those doc fixes.
- 2026-10-18T23:47:39Z — the whole point of armada-nitro is that we don't need in
  privacy pool lending/borrowing we can just use hubs and existing aave. confirm.
- 2026-10-18T23:55:41Z — I think we can remove 14? any contract can hold notes?
  17 looks like it can also be removed since it's a negative assertion. I don't
  understand how 18, 19, 20 became optional. 21. should probably be removed based
  on the above.
- 2026-10-18T23:57:07Z — but funding comes from existing notes. confirm.
- 2026-10-19T12:33:15Z — fork armada-poc
- 2026-10-21T13:52:42Z — how can publicOutputs make any sense as optional?
- 2026-10-21T13:59:25Z — your two invariants are correct. explain how the
  conclusion from those two is this registry that was never built. what is that
  registry for?
- 2026-10-21T14:12:41Z — let's remove 2 as a separate codepath for the
  crowdfund, baal won't be participating in the crowdfund.
- 2026-10-21T14:16:39Z — 1. this would cause a lot of problems and isn't in the
  existing specs? 2. through step 3. as I just mentioned. 3. probably. if it
  actually works. The crowdfund is already deployed baal doesn't need to have
  hard coded support it for it. that's an improper separation of concerns.
- 2026-10-21T14:20:35Z — yes. make those changes and the pool needs to reject
  naked erc20 deposits, is that possible
- 2026-10-21T14:22:30Z — the mark and sweep is the stuck funds discussion, it's
  the same thing caused by the same issue. confirm.
- 2026-10-22T07:24:49Z — 2. Yes. write the entries. 3. Yes. this needs to be
  well documented.

## 3. Findings and decisions

Each finding is one sentence. Each one traces to the evidence or a ruling above.

1. A Safe can own shielded notes through a `Transact` variant that derives the
   note keys from a public `owner` address and disables the EdDSA check, because
   the hook already exists at `Transact.circom:158` (reply 22:26 UTC).
   **Replaced for contract owners by §3a R1 (2026-10-22).**
2. The pool authorizes a contract-owned spend by `msg.sender == owner`, and it
   keeps a public `balance[owner]` registry, because the contract holds no
   spending key (reply 22:22 UTC). **Replaced by §3a R1–R3: the balance is the
   holding, not a registry beside notes (2026-10-22).**
3. A contract-owned note has a public preimage and a public `nullifyingKey`, and
   the creator publishes the preimage in the creating transaction, because the
   contract cannot decrypt a ciphertext (replies 22:32, 23:00 UTC). **The
   invariant stays; §3a R2 enforces it: a contract owner receives value only as
   a public credit (2026-10-22).**
4. A public nullifier is not a show-stopper, because double-spend prevention
   needs one nullifier per leaf, not a secret key, and the treasury notes are
   public by construction (ruling 2026-10-18T22:30:13Z; reply 22:32 UTC).
5. Selective disclosure circuits replace POI, so no gate blocks a deposit, a
   transfer, or a withdrawal (ruling 2026-10-18T22:30:13Z).
6. A new Groth16 ceremony is triggered by any circuit change, and the
   contract-owned variant adds shapes to the Phase-2 ceremony that Armada already
   runs, so it adds no new ceremony kind (reply 22:38 UTC; ADR-0003 D3.4).
7. The pool must not be upgradeable; it deploys with no proxy and no admin, and
   it sets verifying keys and fees once at deployment, because A.2 today keeps
   Railgun's upgradeable proxy and that contradicts the requirement (ruling
   2026-10-18T22:48:00Z; reply 22:48 UTC).
8. The vkey setter must not be an authority the Baal DAO can reach, because a
   swapped verifying key drains every pool user's funds and ragequit protects
   only ARM holders (reply 22:34 UTC).
9. The adjudicator owns its channel collateral as notes, so the collateral never
   leaves the pool and there is no public `deposit` path, no raw-transfer
   surplus, and no `sweep` (reply 22:54 UTC). **Probably replaced by §3a R6
   (2026-10-22).**
10. The change to armada-nitro's privacy claims is practically zero, because
    amounts and timing stay public and the sender was already hidden; the one
    gain is that a payout no longer names the recipient's `npk` in a `Shield`
    event (reply 23:01 UTC).
11. The adjudicator-owned note's preimage is published at creation, in the same
    transaction, inside `adaptParams`, and the adjudicator verifies it against
    the proof's output commitment before it credits anything (reply 23:00 UTC).
    **Probably replaced by §3a R6 (2026-10-22).**
12. A Safe's balance cannot be made private by any circuit, because the Safe must
    know its own balance to act and holds no secret (reply 23:03 UTC).
13. Hidden channel balances need a new nitro circuit, because the adjudicator's
    balance logic moves inside a proof and channel participants hold keys to prove
    with (reply 23:03 UTC).
14. A complete proof for each state change is rejected; the channel agrees on
    output commitments, shares preimages at signing, and proves once at
    finalization (ruling 2026-10-18T23:16:16Z; reply 23:16 UTC).
15. A generic ZK `IForceMoveApp` holds the game rules in circom and wraps a
    Groth16 verifier, because `stateIsSupported` is the only place game logic
    lives (reply 23:18 UTC).
16. A ZK game circuit is one per game type, not one per instance, because the
    Phase-2 ceremony binds to the circuit and every instance uses the same
    verifying key (reply 23:20 UTC).
17. The mover attaches a proof to each signed state, and the counterparty
    verifies it off-chain before countersigning, because proving per move is
    intrinsic to hidden state (reply 23:21 UTC).
18. The construction is a hybrid validium with an optimistic recency window: the
    proofs give correctness, and the window settles which valid state is newest,
    not whether a state is valid (replies 23:22, 23:23 UTC).
19. Hubs stay unchanged and do not gain private balances, because a hub is a
    contract and a contract-owned note is as public as the ERC-20 balance it
    replaces (reply 23:48 UTC).
20. The plan is that hubs take Aave positions, provide liquidity, and swap the
    `waUSDC` inside a channel in private, so no in-pool lending or borrowing is
    built (rulings 2026-10-18T23:44:40Z and 2026-10-18T23:47:39Z).
21. Channel funding comes from a spend of the user's existing note, so an
    observer sees "a channel was funded with X" and learns nothing about who
    funded it (ruling 2026-10-18T23:57:07Z; reply 23:57 UTC).
22. The session added a `fork` remote to `AFDudley/armada-poc`, which already
    existed on GitHub (ruling 2026-10-19T12:33:15Z; reply 12:33 UTC).
23. All circuits are circom on `circomlib`, and the contract-owned variant is a
    small delta because it rewires existing `circomlib` components and adds none
    (replies 12:42, 12:47 UTC).
24. The mobile surface is a skeleton, so the client chunks C1 to C6 cannot run
    until `armada-client` Phase 1 and Phase 2 merge (reply 12:31 UTC).

## 3a. Revision of 2026-10-22: contract balances in the pool

Session `01a156a6` (2026-10-21 13:30 UTC to 2026-10-22 07:25 UTC) traced the
merged armada-poc code from a deposit to `ShieldedRagequit`. It found three
defects that decisions 1 to 3 did not prevent:

- A plain `transact` makes `publicOutputs` optional
  (`TransactModule.sol:543-605`), so a sender can create a contract-owned note
  with no recorded preimage. No one can spend that note.
- The P1 circuit derives `npk = Poseidon(ownerField, random)`
  (`armada-circuits` `src/operations/TransactOwned.circom:66-70`). The fresh
  `random` hides the owner, so the pool cannot tell a contract-owned output from
  a private one, and cannot enforce decision 3.
- The `balance[owner]` registry of decision 2 was never built. So
  `ShieldedRagequit` reads one note that the caller names, not the treasury.

Two invariants stay. A contract-owned value is public at creation (decision 3).
Incoming treasury value becomes governed value in the same transaction (sibling
entry, decision 12). A merge of incoming value into an owner's note needs the
owner as caller, and a payer is not the owner. So the holding itself must be a
pool balance (rulings 2026-10-21T13:52:42Z, 13:59:25Z and 14:16:39Z).

R1. A contract owner holds a public per-token balance in the pool,
    `balance[owner][token]`, and no notes.
R2. Any payer credits a contract owner with a private `transact` that ends in an
    unshield whose destination is `balance[owner][token]` instead of an ERC-20
    transfer. The unshield output's recipient, token and value are already
    public and proven by the existing `Transact` circuit
    (`TransactModule.sol:314-329`; `Transact.circom:103-112`), so the change is
    in the contracts only: no new circuit and no new ceremony. The pool credits
    the balance in that transaction.
R3. The owner spends only as the caller. The pool debits `balance[owner][token]`
    when `msg.sender == owner` and creates the recipient's note the way `shield`
    creates one, from the internal balance and with no proof. No shield fee
    applies, because shield fees apply only to shields
    (`ShieldModule.sol:130,178`).
R4. The pool keeps `accounted[token]`. Every shield and credit raises it, and
    every unshield and debit lowers it. A standard ERC-20 transfer runs no code
    in the recipient, so the pool cannot refuse a raw transfer or an airdrop.
    Such tokens form the surplus `balanceOf(pool) − accounted[token]`, which no
    note and no balance can claim (ruling 2026-10-21T14:20:35Z).
R5. The pool invariants are exact. Per token, `accounted[token]` equals the
    value behind unspent notes plus the sum of `balance[owner][token]` over all
    owners. Per owner, `balance[owner][token]` equals that owner's public
    credits minus its debits. An inequality is rejected, because it hides a
    leak (ruling 2026-10-21T14:20:35Z).
R6. The adjudicator (decisions 9 and 11) probably moves to
    `balance[adjudicator][token]` too. A separate check must show that the
    mechanism works for channel collateral before F4 continues (ruling
    2026-10-21T14:16:39Z).
R7. The treasury does not need the owned circuit, the owned verifying keys,
    `transactOwned`, `verifyOwned`, `ownerField` or the optional
    `publicOutputs`. Their removal waits on R6, because N1a and
    `ShieldedRagequit` use them today.
R8. Tests run the production deploy and real proofs. The pool keeps no proof
    bypass: `testingMode` and `VERIFICATION_BYPASS` go. The upstream audit fix
    `534e6306` ("Fix C-1/C-2: remove setTestingMode() and VERIFICATION_BYPASS")
    already did this, but it was never merged into `integrated` (ruling
    2026-10-21T13:30:34Z).

## 4. The check for each decision

Each named check would show the decision false or would enforce it. The gate
chunks come from the 209-point list and the plan. This entry names the checks; it
does not build them.

| Decision | Check |
|---|---|
| 1, 2, 3 (contract-owned variant, owner auth, preimage registry) | P1 differential test of the variant against captured vectors; P2 `forge test` for the owner path and the registry. |
| 6 (new shapes add to the ceremony, not a new kind) | P4 Phase-2 run for the new shapes; `snarkjs zkey verify` over the new shapes in CI. |
| 7 (pool not upgradeable) | P3 non-upgradeable deployment: constructor keys, no proxy, deploy script. |
| 9, 11 (adjudicator-owned notes, preimage at creation) | N2 `depositFromNote` as an in-pool spend and payout as a spend; N7 re-runs E11 and E16 on in-pool collateral. |
| 13, 15, 16 (hidden-balance circuit, generic ZK app, per game type) | N4 `forge test` with a trivial circuit through the ZK app; N5 ledger circuit vectors; N6 game template; F2 gate resolves one hidden-state move on the fixturenet. |
| 20 (hubs + Aave, swap fee to treasury) | N8 gate: in-channel swap fee → Safe-owned note → `cumulativeFees` → RevenueLock release. |
| 24 (client reaches the mobile surface) | C5 gate: hidden-balance `waUSDC` swap through the mobile client on the fixturenet. |
| R1, R2, R3 (balances, credit by unshield, owner-only debit) | `forge test` on the production deploy: an unshield to a balance raises `balance[owner]`; a debit by any caller other than the owner reverts; a payout pays no shield fee. |
| R4, R5 (accounting, exact invariants) | Invariant tests on the production deploy: the per-token equality and the per-owner ledger hold; a raw transfer changes only the surplus. |
| R8 (no proof bypass) | A test that `PrivacyPool` has no `testingMode` and no `VERIFICATION_BYPASS`; every pool test proves with the real circuit artifacts. |

## 5. Rejected options

- **Proof of Innocence (POI) as a gate.** Rejected. The operator ruled that
  selective disclosure circuits replace POI, so no gate blocks a pool action
  (ruling 2026-10-18T22:30:13Z).
- **An upgradeable pool proxy, as A.2 keeps it today.** Rejected. The operator
  ruled that armada must be largely not upgradable, which the proxy contradicts
  (ruling 2026-10-18T22:48:00Z).
- **A complete `transact` proof attached to every state change.** Rejected. It
  pays fork-lite's cost and gains none of its benefit, and it accumulates
  mutually exclusive valid spends of the same notes (ruling 2026-10-18T23:16:16Z;
  reply 23:16 UTC).
- **An in-channel borrow app with a lender hub, LTV and liquidation.** Rejected.
  The operator ruled that hubs use existing Aave and swap `waUSDC` in private,
  with no in-pool lending or borrowing (rulings 2026-10-18T23:44:40Z and
  2026-10-18T23:47:39Z). The "borrow app" came from page 14's word "borrowing"
  and from page 15's parked note, not from a plan (reply 23:46 UTC).
- **An "armada-aave" clean-address product as the borrow path.** Rejected with
  the borrow app, for the same ruling.
- **Private hub balances through a threshold BabyJubJub key.** Rejected for the
  hub. A hub is a contract and holds no spending key, so a hub-owned note is as
  public as an ERC-20 balance (reply 23:48 UTC).
- **A circuit that makes the Safe treasury balance private.** Rejected as
  impossible. The Safe must know its own balance to act (reply 23:03 UTC).
- **List items 14, 17 and 21 of the agreed-changes list.** Removed. Any contract
  can hold notes, item 17 was a negative assertion, and item 21 was superseded by
  commitment outcomes (ruling 2026-10-18T23:55:41Z; reply 23:55 UTC).

### A correction the session made to its own evidence

The session first called the new ceremony "a new way to drain the whole pool"
(reply 22:34 UTC). It then withdrew that framing after reading ADR-0003: the
variant adds shapes to the Phase-2 ceremony that Armada already runs, so it is
the same 1-of-N honest assumption, not a new ceremony kind (reply 22:38 UTC).
Both statements are in the transcript. The later statement is the finding.

## 6. Open questions

- Who holds the proxy admin and the `setVerificationKey` authority after launch,
  given that A.2.7 defers this and the answer must not be the Baal DAO?
- Does the pool restrict a contract-owned note's creation to the owner, so that a
  stranger cannot create an unrecorded commitment at the adjudicator's `npk`?
  Resolved by §3a R1 and R2 (2026-10-22): a contract owner holds no notes.
- Where do the pool contracts live? The plan marks `armada-poc/contracts` as an
  [INFERENCE] and asks the operator to confirm, or to init a third repo.
- Is a per-pebble Sonnet route override available in exophial, or must a husband
  re-file the pebble to reach the Sonnet route?
- Which four-character tracker prefix does each repo take? The session suggested
  `arpc`, `arcx`, `arsd` and `arcl`; the operator did not confirm them.

## 7. Chunks this entry covers

Copied from the 209-point list of 2026-10-19 03:17 UTC. This entry covers the
pool chunks, the armada-nitro chunks, the client chunks, and the D2 document
chunk.

| # | Chunk | Depends on | Pts |
|---|---|---|---:|
| **Pool, armada-circuits + pool contracts** | | | |
| P1 | Contract-owned `Transact` variant, differential tests, manifest shapes | — | 9 |
| P2 | Owner auth path, preimage registry, fee-tagged outputs, pool-level `cumulativeFees`, `ShieldModule` fee note, tests | P1 | 11 |
| P3 | Non-upgradeable deployment: constructor keys, no proxy, deploy script | P2, P4 | 6 |
| P4 | Phase-2 run for the new shapes | P1 | 4 |
| **armada-nitro** | | | |
| N1 | `MultiAssetHolder` on commitments | P2 | 11 |
| N2 | `depositFromNote` as in-pool spend, payout as spend including fee-tagged outputs to the Safe, tests | N1 | 10 |
| N3 | Commitment-outcome finalization branch, tests | N2 | 8 |
| N4 | Generic ZK `IForceMoveApp`, verifier wiring, tests with a trivial circuit | — | 7 |
| N5 | Hidden-balance ledger circuit, vectors, ceremony, integration with N4 | N4 | 9 |
| N6 | Game template circuit and one example game | N4 | 7 |
| N7 | E11 and E16 gates updated for in-pool collateral | N2, N3 | 8 |
| N8 | Gate: in-channel swap fee → Safe-owned note → `cumulativeFees` → RevenueLock release | N2, N7, G4 | 6 |
| **Client, armada-sdk / ts-nitro / wallet** | | | |
| C1 | Prover integration for the new artifacts: contract-owned variant, ledger, game | P4, N5 | 7 |
| C2 | State update with proof: generate, attach, verify counterparty, persist state + preimage + proof | C1 | 10 |
| C3 | Watchtower stores and submits proofs | C2 | 5 |
| C4 | ts-nitro: ZK app and commitment-outcome finalization | N3, N4 | 9 |
| C5 | Gate: hidden-balance `waUSDC` swap through the mobile client on the fixturenet | C2–C4, N7 | 10 |
| C6 | Wallet: `ShieldedRagequit` flow with on-device proof | G3, C1 | 6 |
| **Documents** | | | |
| D2 | A.2, A.3, A.5, ADR-0005 note, page 16 note, ZK-app page with prior art | — | 6 |

## 8. The execution plan

The session drafted a plan to run this work with exophial and husbands on the
5090, with escalation to Sonnet (plan of 2026-10-19 03:31 UTC).

- **Routing.** Workers run on the local 27B model. The dispatcher refuses any
  pebble above 16 points on the 5090 (163,200 tokens). A worker that outgrows the
  window promotes to `claude-sonnet-5` without anyone acting.
- **Chunk cap for the local route is 9 points**, because the one measured worker
  held an 86,000-token startup floor. P2 and N1, both 11 points, must split or
  start on Sonnet.
- **Start on Sonnet, no local attempt: circuits P1 and N5.** circom on a 27B
  model is an unmeasured risk and both are on the critical path.
- **Families and husbands.** One husband owns one family. Two husbands run at a
  time. F1 pool owns P1 to P4 (armada-circuits, armada-poc). F2 zk-app owns N4,
  N5, N6 (armada-nitro, armada-circuits). F4 collateral owns N1, N2, N3, N7, N8
  (armada-nitro). F5 client owns C1 to C5 (armada-sdk, wallet). F6 docs owns D1
  and D2 as a single Sonnet subagent, prose only.
- **Waves.** Wave 1 runs F1 pool and F2 zk-app, which are independent. Wave 2
  runs F3 governance and F4 collateral, both waiting on P2. Wave 3 runs F5
  client, waiting on N3, N4 and N5. The critical path is P1 → P2 → G3 → G7.
- **Review.** One `reviewer` pass (Sonnet) runs on the audit-critical chunks P1,
  P2, N1, N2 and G3 before a family closes. Findings become child pebbles.
- **Size.** 209 points, about 27 chunks, two husbands in parallel, so about 14
  husband sessions of work plus rework.
- **Prerequisite.** `armada-poc`, `armada-circuits` and `armada-sdk` need tracker
  init before any pebble filing.
