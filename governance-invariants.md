# Governance invariants extracted from the Armada specs

This document lists the invariants that the original Armada governance specification
states or implies. It also lists the places where an invariant cannot be written,
because the specification is silent, ambiguous, or contradicts itself.

**Sources.** `ship-armada/armada-poc`, branch `main`, commit `7bdefadd`:

| Key | File |
|---|---|
| `G` | `specs/GOVERNANCE.md` (960 lines) |
| `A` | `specs/ARM_TOKEN.md` (468 lines) |
| `R` | `specs/REVENUE_LOCK.md` (307 lines) |
| `M` | `specs/PARAMETER_MANIFEST.md` (264 lines) |

A citation such as `G:86` means `specs/GOVERNANCE.md` line 86 at `7bdefadd`.

**Out of scope.** This pass did not extract invariants from `FEE_STRUCTURE.md`,
`CROWDFUND.md` or `MONITORING.md`. It reads them only where `GOVERNANCE.md` depends on them.

## How to read the tables

Each invariant is a property that an implementation must satisfy. The wording avoids
naming a contract where the specification allows it, so the same list can be checked
against a different design, for example the Baal redesign.

| Column | Meaning |
|---|---|
| Kind | **S** = the specification states it as a rule. **D** = derived from a description of a mechanism; the intent is an inference. **C** = holds only if a deploy step is done correctly. |
| Check | How a machine can decide it. **prop** = a fuzz, invariant or symbolic test on the contracts. **deploy** = a check of the deployed state. **static** = a check of the code or ABI shape. |

---

## 1. ARM token (`T`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| T1 | `totalSupply()` equals 12,000,000 × 10^18 at all times. | S | A:23, A:401 | prop |
| T2 | No code path creates ARM after construction. | S | A:402 | static, prop |
| T3 | No code path destroys ARM. | S | A:403 | static, prop |
| T4 | A holder's balance decreases only through a transfer that the holder signed or approved. | S | A:104, A:405 | prop |
| T5 | While transfers are restricted, a transfer succeeds if and only if the sender is on the allowlist. No other rule blocks a transfer. | S | A:150, A:404 | prop |
| T6 | `transferable` changes only from false to true, and never back. | S | A:407 | prop |
| T7 | Only the governance executor and the wind-down contract can set `transferable`. | S | A:173 | static, prop |
| T8 | An address on the allowlist stays on it, with one exception: the deployer can remove its own entry once. | S | A:46–47, A:148, A:410 | prop |
| T9 | ARM held by the treasury has zero votes, because `delegate()` from the treasury reverts. | C | A:414, A:280 | deploy, prop |
| T10 | ARM that is not delegated has zero votes. | S | A:413 | prop |
| T11 | A revenue-lock contract never delegates its own balance. | S | A:412, R:208 | static |
| T12 | Two ARM units from different allocations have the same rights after any restriction on them ends. | S | A:415 | static |
| T13 | After deploy: RevenueLock holds 2,400,000 ARM, the crowdfund holds 1,800,000 ARM, the treasury holds 7,800,000 ARM, and the bootstrap holder holds 0 ARM with no allowances. | S, C | A:41–47, M:164 | deploy |
| T14 | The token has no proxy and no upgrade path. | S | A:332–342, G:841 | static, deploy |

## 2. Voting power (`V`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| V1 | The votes counted for a proposal are the delegated balances at block `creation − 1`. No later transfer or delegation changes them. | S | G:36 | prop |
| V2 | Delegation has one level. A delegate's received votes are not forwarded. | S | G:58 | prop |
| V3 | During voting, a voter can change a choice but cannot withdraw a cast vote. Total participation on a proposal never decreases while voting is open. | S | G:56 | prop |
| V4 | ARM that leaves the crowdfund by `claim` or leaves RevenueLock by `release` is delegated in the same transaction. | S | G:62, R:209 | prop |

## 3. Quorum (`Q`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| Q1 | `quorum(p) = max(eligible(p) × bps(type) / 10,000, 200,000 ARM)`. `eligible(p)` and the percentage are fixed when `p` is created. | S | G:86–89, G:302 | prop |
| Q2 | `eligible(p)` excludes ARM held by the treasury, RevenueLock, the reserve distributor and the crowdfund. It includes undelegated ARM in ordinary wallets. | S | G:92–103, G:130 | prop |
| Q3 | The 200,000 ARM floor applies to every vote type, including the defeat of a steward spend and veto ratification. | S | G:302 | prop |

See gap **N6**: the specification uses two different terms for the quorum base.

## 4. Proposals and classification (`P`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| P1 | An executable proposal (other than a steward proposal) executes only if quorum was met, FOR > AGAINST, and the proposal delay, voting period and execution delay have all elapsed. | S | G:24, G:136–148 | prop |
| P2 | A proposal that grants authority, loosens a constraint or increases risk exposure is Extended. | S | G:178–198 | prop |
| P3 | If any action in a batch is Extended, the whole proposal is Extended. A mixed-direction change is Extended. | S | G:225, G:262 | prop |
| P4 | A classification that depends on the treasury balance uses a reference value fixed at creation or queue time. | S | G:255–258 | prop |
| P5 | A signaling proposal has no queue, no execution and no veto. | S | G:271–285 | prop |
| P6 | No proposal of any kind, including a steward proposal, is accepted within 10 days after crowdfund finalization. | S | G:306, M:135 | prop |
| P7 | A proposer must hold at least 5,000 delegated ARM at the snapshot. | S | G:172 | prop |
| P8 | A queued proposal does not expire. Anyone can retry its execution. | S | G:807 | prop |

## 5. Treasury outflow (`TR`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| TR1 | For each asset, total outflow in any trailing window ≤ `max(bps × treasury balance, absolute limit)`, using the active parameters. At launch: USDC max($100,000, 10%), ARM max(250,000, 3%), ETH max(25 ETH, 10%), 30-day windows. | S | G:776, G:784, G:794, M:156–158 | prop |
| TR2 | TR1 counts every outflow channel together: governance, steward and any authorized module. | S | G:799 | prop |
| TR3 | A change that loosens an outflow parameter takes effect at least 24 days after it executes. A change that tightens one takes effect at once. | S | G:813–814, M:145 | prop |
| TR4 | The longest Extended governance cycle is shorter than the 24-day activation delay. | S | G:816 | prop |
| TR5 | A new loosening change replaces a pending one and restarts the 24-day timer. | S | G:831 | prop |
| TR6 | A proposal is queued only if its total spend per asset ≤ the effective limit at queue time. | S | G:805 | prop |
| TR7 | The outflow floor is 0 and no function changes it. | S | G:803, M:156–158 | static |
| TR8 | The treasury has no upgrade path at launch. | S | G:842 | static |

## 6. Treasury steward (`S`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| S1 | The steward spends only ERC-20 tokens in the budget table, and never ETH. | S | G:383, G:404 | prop |
| S2 | For each token, steward spend in any trailing window ≤ that token's budget. Unused budget does not carry over. At launch: USDC $60,000 per 30 days. | S | G:423, M:163 | prop |
| S3 | Steward spend counts toward TR1. | S | G:425–426 | prop |
| S4 | A steward proposal is defeated if and only if quorum is met and AGAINST > FOR. Otherwise it passes. | S | G:416 | prop |
| S5 | A steward proposal never pays the steward's registered address. | S | G:447 | prop |
| S6 | Removing the steward voids every queued steward proposal at execution. | S | G:443 | prop |
| S7 | A steward term is 180 days. Election is Extended. Removal is Standard and takes effect at once. | S | G:436–440, M:162 | prop |

## 7. Security Council (`SC`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| SC1 | The Security Council cannot move treasury funds, upgrade a contract, change a fee or governance parameter, execute an arbitrary transaction, pause unshields or pause the governor. | S | G:471–478 | static, prop |
| SC2 | The Security Council can veto only a queued proposal, during its execution delay. | S | G:466, G:485 | prop |
| SC3 | A veto starts a 14-day ratification vote at once. If FOR wins, or quorum is not met, the proposal is cancelled permanently. If AGAINST wins with quorum, the proposal is restored with at least a 2-day delay, and the Security Council address becomes zero. | S | G:486–500, M:140 | prop |
| SC4 | A restored proposal cannot be vetoed again. | S | G:504 | prop |
| SC5 | A Security Council shield pause has no effect, before or after wind-down. | S | G:465 | prop |
| SC6 | The Security Council can cancel the crowdfund only before finalization. | S | G:467 | prop |

## 8. Revenue and the revenue lock (`R`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| R1 | `recognizedRevenueUsd` never decreases. | S | G:587 | prop |
| R2 | USDC fee revenue enters the counter only from a monotonic fee-collector total, never from a treasury balance. | S | G:582 | prop |
| R3 | The lock computes entitlement only from `maxObservedRevenue`, never from the counter directly. | S | R:214 | static |
| R4 | `maxObservedRevenue` never decreases and rises by at most $10,000 × elapsed days per update. | S | A:202–204, R:211–212, M:159 | prop |
| R5 | The milestone table cannot change. No unlock is based on time. | S | G:560, R:206 | static |
| R6 | `RL.balance + Σ released = totalAllocation`, and each beneficiary's released amount ≤ its allocation. | S | R:204, R:210 | prop |
| R7 | ARM leaves the lock only by `release` to the entitled beneficiary. | S | R:196 | static, prop |
| R8 | The beneficiary list cannot change, and its total is exactly 2,400,000 ARM. | S | R:77–79, R:207 | static, deploy |

## 9. Wind-down and redemption (`W`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| W1 | The permissionless trigger succeeds if and only if `now > deadline` and cumulative revenue < threshold. | S | G:678 | prop |
| W2 | A governance trigger is Extended. | S | G:680 | prop |
| W3 | After the trigger, `transferable` is true. | S | G:690, A:166–171 | prop |
| W4 | After the trigger, no proposal can be created, queued or executed, including steward proposals. | S | G:693, G:758 | prop |
| W5 | After the trigger, unshield works forever. Shield and shielded transfer are disabled. | S | G:692 | prop |
| W6 | After the trigger, treasury ARM never moves, and `sweepToken(ARM)` reverts. | S | G:694–695 | prop |
| W7 | A redeemer of `d` ARM receives `d / circulating × balance` of each non-ARM asset held for redemption, where `circulating` follows the formula at G:710. | S | G:700–719 | prop |
| W8 | A `claim` or `release` does not change `circulating`. | S | G:737 | prop |
| W9 | Total redemption payouts never exceed the assets swept for redemption. The last redeemer can always be paid. | S | G:738 | prop |
| W10 | Redemption opens 14 days after the trigger, and then stays open with no deadline. | S | G:735 | prop |
| W11 | At go-live: `lockedAtWindDown + unsold crowdfund ARM ≤ circulatingSupplyOf([treasury, redemption])`. | S, C | G:721–727 | deploy |
| W12 | Early-network ARM that is still locked cannot be redeemed. | S | G:754 | prop |
| W13 | After the trigger, the revenue counter and the lock ratchet are frozen. | S | G:691 | prop |

See gap **N13** about W13 and gap **N11** about W7.

## 10. Upgrades and immutability (`U`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| U1 | These contracts have no upgrade path: ARM token, crowdfund, redemption, revenue lock, wind-down, shielded pool, and (at launch) the treasury. | S | G:841–850 | static, deploy |
| U2 | Only an Extended governance proposal can upgrade the governor, the fee module or the revenue counter. | S | G:847–849 | prop |
| U3 | Authorizing an adapter is Extended. Removing one is Standard. | S | G:915–917 | prop |
| U4 | The revenue lock reads a fixed counter address through a fixed interface. | S | G:592–593 | static |

## 11. Pool access and exit (`X`)

| ID | Invariant | Kind | Source | Check |
|---|---|---|---|---|
| X1 | No role can pause unshields, before or after wind-down. | S | G:465, G:476, G:692 | prop |
| X2 | No proposal makes pool use depend on disclosure, attestation, an allowlist, a blocklist, identity or proof of provenance, unless an Extended proposal first amends the Immutable table. | S | G:355, G:363–371 | not decidable on-chain; see N3 |

---

## 12. Gaps: places where no invariant can be written

Each gap names what is missing, and the decision that would close it.

### Missing bound or threat model

**N1. No bound on treasury loss under capture.** The specification calls the outflow limits "the primary defense against governance capture" (G:769). It also says that the system degrades "predictably (bounded leakage, slow degradation)" (G:933). But one Extended proposal can do each of the following:
- loosen the limits (G:814);
- upgrade the governor (G:847);
- reach any target that the timelock controls (see N2).

"Approximately 47 days" (G:823) is a timeline estimate, not a bound. So no statement of the form "the treasury loses at most X per window, whatever governance does" follows from the text. **Missing:** which voting majority the design trusts, and the loss bound that must hold against a majority it does not trust.

**N2. The set of governance powers is open.** The Governable table (G:314–345) is not stated to be complete. G:215 makes every unlisted action Standard. No section lists the contracts that the timelock owns or the roles it holds. So "governance cannot do X" can be written only where a contract has no setter at all (T1–T14, R3–R8, U1). W6 ("treasury ARM never moves after the trigger") also depends on this list, because a timelock role holder other than the governor is not excluded. **Missing:** a closed list of the timelock's powers, with the proposal type for each.

**N3. "Immutable" mixes two meanings.** Some rows hold because the code cannot change (the token, the lock). The pool-access row is a promise that an Extended proposal can amend (G:355, G:371). X2 is therefore a rule about governance behaviour, not an on-chain property. **Missing:** for each Immutable row, whether code enforces it or governance promises it.

**N4. Governor invariants across an upgrade.** The governor is upgradeable by one Extended proposal (G:847). P1–P8, Q1–Q3, SC2–SC4 and W4 are all governor behaviour. The specification does not say which of them a governor upgrade must keep. The two-stage upgrade process that would protect them is not implemented (G:853). **Missing:** the list of governor invariants that an upgrade must preserve.

**N5. No lower bounds on timing or quorum.** The specification gives launch values and the direction rule (G:229–240), but no minimum voting period, execution delay or quorum percentage. So "the Security Council always has at least N days to veto" cannot be written. M:165 says that the governor rejects a timelock `updateDelay` above 2 days. That is an upper cap, so the timelock delay can be lowered. **Missing:** the minimum value of each timing and quorum parameter.

**N6. Two terms for the quorum base.** G:86–103 defines `eligibleSupply`, which includes undelegated ARM (G:130). The Parameters table (G:296), the steward defeat rule (G:416), the circuit-breaker participation rule (G:434) and the ratification rule (G:500) say "circulating voting power", which reads as delegated ARM only. **Missing:** one defined term.

**N7. The base of the percentage limit is undefined.** "10% of USDC in treasury" (G:776) does not say when the balance is read: at execution, at queue time, or at the start of the window. TR1 therefore has two readings, and they give different results while the balance falls during a window. **Missing:** the time at which the balance is read.

### The specification contradicts itself

**N8. Security Council size.** "2-of-3 multisig" (G:459) against "Core team (2), external security (2), community (1)" (G:515), which is five members. **Missing:** the signer count and the threshold.

**N9. Authority over `delegateOnBehalf`.** A:254 and A:262 say that the callers are "set immutably" in the constructor and that auditors must verify that the set cannot change. A:409 says that governance can add callers. The audit finding B-03 depends on this choice. **Missing:** a ruling: fixed set or add-only set.

**N10. How the allowlist is set.** A:129 and A:410 say that the initial allowlist is set in the constructor. A:131 and A:67 say that a one-shot `initWhitelist` call sets it after deploy. A:148 and A:410 say that no entry is ever removed, but A:47 allows the deployer to remove its own entry. T8 states the exception. **Missing:** one statement that the code and the deploy script both follow.

**N11. Later revenue-lock cohorts.** G:730 says that the revenue lock "is a one-time launch construct… no future lock contracts". R:240–254 describes later cohorts and says that they are not in the redemption denominator (R:252). With any cohort, the per-ARM equality in W7 and the quorum base in Q2 change. **Missing:** a ruling: cohorts allowed or not. If allowed, the rule for how cohort ARM counts in W7 and Q2.

**N12. The bootstrap holder.** A:32 describes "a single EOA, preferably held on a Ledger". A:55 says that it "must be a deployment multisig. A single EOA is not acceptable." **Missing:** a ruling.

**N13. What wind-down freezes.** G:691 says that `freeze()` rejects `attestRevenue` and `syncStablecoinRevenue`. It does not name `addRevenue`. W13 ("the counter is frozen") cannot be written exactly. **Missing:** whether every write path to the counter stops at the trigger.

**N14. "Proposals fail by default."** G:24 states this as the model. Steward proposals pass by default (G:412–416). During the launch window, a steward proposal cannot be defeated at all (G:443). **Missing:** the scoped rule, for example: "every proposal except a steward proposal within budget fails by default".

### Intent not decided

**N15. Equal payout per ARM in redemption.** G:738 accepts that ARM sent to the treasury after the trigger raises the payout rate for later redeemers. It notes that a snapshot at the trigger would prevent this, but that the snapshot is not implemented. **Missing:** whether every redeemer must receive the same amount per ARM, or only W9 (no over-payment) must hold.

**N16. Revenue attestation has no bound.** `addRevenue` accepts any amount that a Standard proposal passes (G:583). "Attestations must reference verifiable on-chain receipts" (G:590) is a "should" rule that nothing checks on-chain. The only limit is the lock's rate cap (R4), which delays an inflated value but does not reject it. "Expand the qualifying revenue definition" (G:591) has no definition of what may be added. **Missing:** an upper bound on recognized revenue, or a verification rule, and a definition of qualifying revenue.

### Rules that are not on-chain

The following are obligations on people or operations. A check of the contracts cannot decide them. Each one is listed with the invariant that fails if the obligation is not met.

| Obligation | Source | What fails if it is not met |
|---|---|---|
| The Security Council publishes a written rationale for each veto. | G:485 | Nothing on-chain |
| The Security Council vetoes only when it expects the community to support it. | G:502 | Nothing on-chain |
| Monitoring calls `syncObservedRevenue()` at least daily. | A:206–208 | R4 still holds per call, but the effective rate limit loosens |
| The launch team submits no steward proposal before the quiet period ends. | G:443 | Nothing. The launch-window exposure stays open. |
| An adapter with user positions moves to withdraw-only instead of being removed. | G:917 | Nothing on-chain |
| The deployer seeds `initNoDelegation([treasury])`, `initAuthorizedDelegators` and `setWindDownContract` correctly. | A:67, A:280 | T7, T9 |
| Large treasury proposals are discussed in public before they are submitted. | G:666–670 | Nothing on-chain |
| Surplus deployment follows the phase table. | G:650–658 | Nothing. G:658 calls it "guidelines, not constraints". |

### Mechanisms described but not built

These sections describe future behaviour. They give no invariant for the code that exists now:
- the proposal bond (G:154–168);
- the steward circuit breaker (G:428–434);
- the two-stage upgrade process (G:853–909);
- TWAB voting and delegation expiry (G:939–940);
- snapshotting outflow parameters at queue time (G:825);
- directional classification of parameter setters (G:264–266, G:960).
