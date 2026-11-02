# Errors in the Baal governance code

This document lists the errors found in the Baal governance migration (family F3, build plan
stage S4). The code is `ship-armada/armada-poc`, branch `integrated`, commit `afbb81b5`. The
design reference is `labbook/2026-10-18-treasury-governance-baal-safe-shielded.md` (the
"lab book"). The decision numbers below refer to its §3.

**Status of each error.** Each error comes from reading the code. No test or fork run has
reproduced an error yet. Each entry names the test that shows the error.

## Summary

| ID | Error | Effect | Plan rule it breaks |
|---|---|---|---|
| E1 | `ragequit` reverts for every holder who is not on the allowlist while transfers are restricted | No ordinary holder can exit until governance enables transfers | Decisions 1 and 20: "any holder can exit at any time" |
| E2 | The `ragequit` payout divides by the supply after the burn | A holder gets more than a pro-rata share. A holder who burns half the supply gets the full balance. | Decision 20, G3 acceptance |
| E3 | The `ragequit` payout uses the public ERC-20 balance of the Safe, but pays from a pool note that the Safe owns | The payout does not follow the shielded treasury | Decisions 12, 14, 20 |
| E4 | The ARM token still trusts a `timelock` address that the new design does not have | Four token powers have an owner that the plan does not specify | Decisions 1, 9, 10 |
| E5 | `addBurner` lets that `timelock` give any address the power to burn any holder's ARM | Holders can lose ARM without their signature | Decision 3, `ARM_TOKEN.md` §12 "No clawback" |
| E6 | The token keeps code and comments for the deleted wind-down, and a whitelist that can still grow | The token does not match the plan | Decisions 1 and 5 |
| E7 | The treasury ARM stays in the payout divisor | All other holders together can exit with only part of each treasury asset | Operator ruling 2026-10-18 21:21:32 |
| E8 | G6 did not rewrite `specs/GOVERNANCE.md` or `specs/ARM_TOKEN.md` | The branch has two governance specs that contradict each other | G6 brief, S4 and S5 |

## E1. `ragequit` reverts while transfers are restricted

**Code.**
- `ShieldedRagequit.sol:80` calls `armToken.burn(msg.sender, armAmount)`.
- `ArmadaToken.burn` (`ArmadaToken.sol:190`) calls `_burn`.
- In OpenZeppelin 4.9.6 (the installed version), `_burn` calls
  `_beforeTokenTransfer(account, address(0), amount)` (`ERC20.sol:280`).
- `ArmadaToken._beforeTokenTransfer` (`ArmadaToken.sol:282–287`) contains this check:
  `require(from == address(0) || transferWhitelist[from])` while `transferable` is false.

**Effect.** A burn has `from = account`, which is not `address(0)`. A holder who is not on
the allowlist fails the check, so `ragequit` reverts. While transfers are restricted, only
allowlisted addresses can exit. Transfers stay restricted until a governance proposal calls
`setTransferable(true)`. A captured majority can refuse to make that call. So the majority
can block the exit, which is the event the plan exists to prevent.

**Test that shows it.** Deploy with `transferable = false`. Give ARM to an address that is not
on the allowlist. Call `ragequit`. Expect a revert with "ArmadaToken: transfers restricted".

**Fix direction.** Let `_beforeTokenTransfer` accept `to == address(0)` when the caller is an
authorized burner, or do the burn check before the transfer gate.

## E2. The payout divides by the supply after the burn

**Code.** `ShieldedRagequit.sol:80` burns first. `ShieldedRagequit.sol:96` then computes
`payout = armAmount * safeBal / armToken.totalSupply()`. The burn has already lowered
`totalSupply()`.

**Effect.** With supply `S` and burn `a`, the payout is `a × B / (S − a)`, not `a × B / S`.
The extra payout is a factor of `S / (S − a)`.

| Burn as a share of supply | Payout as a share of the Safe balance | Correct share |
|---|---|---|
| 1% | 1.01% | 1% |
| 10% | 11.1% | 10% |
| 50% | 100% | 50% |

Holders who exit first take value from holders who exit later. A holder with half the
supply empties the treasury of each listed token.

**Test that shows it.** Supply 100, Safe balance 100 USDC. One holder burns 50. Expect a
payout of 50. The code pays 100.

**Fix direction.** Read `totalSupply()` once before the burn, and use that value for every
token in the loop.

## E3. The payout base is the public balance of the Safe

**Code.**
- `ShieldedRagequit.sol:94` reads `token.balanceOf(address(safe))`, which is the Safe's
  public ERC-20 balance.
- `_buildOwnedTxn` spends a pool note that the Safe owns, through `pool.transactOwned`.

**Effect.** Decision 12 puts the whole treasury in notes that the Safe owns. Decision 17 says
that a raw ERC-20 transfer to the Safe credits no note. [INFERENCE] The public balance of the
Safe then holds only stuck raw transfers, so:
- the payout is about zero when the treasury is fully shielded;
- the payout follows tainted dust that an outside party can send, not the treasury value.

**Test that shows it.** Shield the treasury into a note that the Safe owns, with a public
Safe balance of zero. Call `ragequit`. Expect a payout of zero. Then send dust to the Safe
and compare the payouts.

**Fix direction.** The payout base must be the value of the Safe-owned notes. Decide the
source of that value before G3 is final, for example a counter that the pool keeps.

## E4. The token trusts a `timelock` that the new design does not have

**Code.** `ArmadaToken.sol:18` holds `address public immutable timelock`. These functions
accept calls only from that address:
- `addToWhitelist` (`:142`)
- `addAuthorizedDelegator` (`:154`)
- `addBurner` (`:178`)
- `setTransferable` (`:236`, which also accepts `windDownContract`)

**Effect.** G6 deleted the governor that controlled the old `TimelockController`. In the plan,
Baal acts through the Safe. The plan has no timelock. The constructor still needs a non-zero
`_timelock`. The handoff (`notes/handoff.md`) lists the deploy-script migration
as open, so no script sets this address for the new design. Two results are possible, and
both are wrong:
- If the address is the Safe, then Baal holds these powers, but no document records that.
- If the address is an old timelock, then whoever holds that timelock's roles holds these
  powers, outside Baal.

**Fix direction.** Name the Safe as the token authority in the constructor and in
`ARM_TOKEN.md`, or remove each function that the plan does not need.

## E5. `addBurner` makes a clawback path

**Code.** `burn(address account, uint256 amount)` (`ArmadaToken.sol:190`) needs only
`authorizedBurner[msg.sender]`. It needs no allowance from `account`. `addBurner`
(`:178`) lets the `timelock` add burners. There is no limit on the count, and burners
cannot be removed.

**Effect.** One proposal can authorize a contract that burns any holder's ARM. That breaks
the `ARM_TOKEN.md` §12 invariant "No clawback". A captured majority could also burn ARM to
change the payout divisor (see E2 and E7). Decision 3 allows burns "used only by ragequit".

**Fix direction.** Fix the burner set at deploy to `ShieldedRagequit` only, and delete
`addBurner`.

## E6. Old wind-down code and an allowlist that can grow

**Code.**
- `ArmadaToken.sol:3`, the ABOUTME line, still says "Transfer-restricted until wind-down".
- `setWindDownContract` (`:103`) and the `windDownContract` caller of `setTransferable`
  (`:237`) remain.
- `addToWhitelist` (`:142`) adds allowlist entries after deploy.

**Effect.**
- Decision 1 drops wind-down, but the token still has a wind-down path that a deployer can
  set one time.
- Decision 5 says that the allowlist is "fixed at construction". The code lets governance add
  entries.

**Fix direction.** Delete the wind-down code. Set the allowlist in the constructor and delete
`addToWhitelist`. Alternatively, record in the lab book why the allowlist must grow.

## E7. Treasury ARM stays in the payout divisor

**Code.** The divisor is `armToken.totalSupply()` (`ShieldedRagequit.sol:96`). The ARM in the
Safe and in the Hedgey lockup counts in that total.

**Effect.** The operator ruling of 2026-10-18 21:21:32 says: "the ARM held in the treasury
should simply be excluded." Decision 6 says instead that this ARM "needs no exclusion list",
because those addresses never call ragequit. That reasoning answers who exits. It does not
answer the divisor question. [INFERENCE] With 7.8M of 12M ARM in the Safe and the lockup, all
other holders together can take at most about 35% of each treasury asset.

**Decision needed.** The operator must say whether "excluded" means "left out of the
divisor". If so, the divisor is `totalSupply − balanceOf(Safe) − balanceOf(lockup)`, read
before the burn.

## E8. Two governance specs on the same branch

**Code.** G6 (`fffbbf31`) wrote new files `GOVERNANCE.md` and `ARM_TOKEN.md` at the repo root.
G6 did not change `specs/GOVERNANCE.md` (935 lines) or `specs/ARM_TOKEN.md`. Those files still
describe the deleted governor, the steward, the veto, wind-down and redemption.

**Effect.** The G6 brief, S4, says: "If `GOVERNANCE.md` exists, rewrite it". The file existed
in `specs/`. The branch now has two specs that contradict each other. The new root
`ARM_TOKEN.md` also repeats E4 and E6: it describes `addToWhitelist` and `setTransferable` as
"timelock-only".

**Fix direction.** Rewrite `specs/GOVERNANCE.md` and `specs/ARM_TOKEN.md` in place, and delete
the copies at the root.

## Related open item

The `ARM_TOKEN.md` §12 invariant "Treasury ARM does not vote" depends on `initNoDelegation`
(`ArmadaToken.sol:115`), which the deployer calls one time. The new treasury is the Safe. No
deploy script for the new design exists yet, so nothing confirms that the Safe is in the
`noDelegation` set. If the Safe is not in the set, a Baal proposal can delegate the Safe's
ARM and vote with the treasury.
