# Hub yield and swaps as channel state

design · 2026-10-11 · accepted ([ADR-0017](./09-architecture-decisions.md#adr-0017))

## What this decides

How a user earns yield and swaps assets with a hub over the armada-nitro rail, and which properties the contracts must enforce so that these flows are correct by construction rather than by careful operation. The rail is the armada-nitro adjudicator and hub custodian of [ADR-0016](./09-architecture-decisions.md#adr-0016), sequenced v1.5, so this design is v1.5 as well.

The answer: **yield-bearing assets are held in the channel as non-rebasing share tokens, and a swap is a co-signed reallocation of one channel state.** No new ForceMove application is needed for the first hub games. The on-chain work is a small, hardening delta to the adjudicator.

This is the design the E11 task in armada-nitro (`docs/tasks/E11-aave-lido-channels.md`) calls "option 2". It refines the first games in [§14](./14-shielded-games-over-nitro.md).

## The two options

**Option 1: one ForceMove app per game** (§14 as first written). A two-asset settlement app enforces a quoted WETH↔wstETH swap, and a lending app pays principal plus interest at close. Lending needs an on-chain rate reference (Aave's reserve index) for disputes, and both apps are net-new contracts to write and audit ([§11 R2](./11-risks-and-technical-debt.md)).

**Option 2: assets as channel state** (chosen for the first hub games).

- The channel holds several assets at once. go-nitro's outcome format already carries one `SingleAssetExit` per asset, and `MultiAssetHolder` keeps `holdings[asset][channelId]` per asset.
- **Yield is a property of the asset, not of a game.** A user who wants USDC yield holds `waUSDC` (Aave's ERC-4626 StataToken over the USDC reserve); a user who wants ETH yield holds `wstETH`. Both are non-rebasing: the token count is fixed and the value per token grows. Holding the token in a channel, or as a note, earns the yield with no contract tracking interest.
- **A swap is one co-signed state** that moves value between assets: for example, the user's `waUSDC` allocation goes to the hub and an equal-value USDC allocation goes to the user. It runs under `ConsensusApp` (every participant signs), like any other cooperative update.

Option 2 needs no rate oracle, no new app, and no new dispute logic: a dispute replays the latest co-signed state exactly as for a single-asset channel.

## Invariants

Each invariant names what makes it hold. "By construction" means a violating state cannot be represented or cannot be reached through the contracts, not that an operator is expected to avoid it.

### I1 · Solvency per asset

For every asset `a`: the sum of `holdings[a][c]` over all channels `c` never exceeds `IERC20(a).balanceOf(adjudicator)`.

This holds only for tokens whose balances change exclusively by transfers, with the transferred amount equal to the requested amount. It fails for:

- **rebasing tokens** (`stETH`, `aUSDC`): balances drift while `holdings` stay fixed; a negative rebase (Lido slashing) leaves the last channel to exit short;
- **fee-on-transfer tokens**: the adjudicator receives less than it credits;
- **tokens with transfer hooks** (ERC-777 style): the hook re-enters during a deposit (see I6).

**Enforcement:**

1. **An asset allow-list in the adjudicator.** Only allow-listed assets may be deposited, by either path. Initial list: USDC, WETH, `wstETH`, `waUSDC`. `stETH` and `aUSDC` are never listed; their non-rebasing wrappers are.
2. **Balance-delta accounting on every deposit path.** `depositFromNote` already credits the measured balance change and requires it to equal the note's value. The inherited go-nitro `deposit`, which the hub custodian uses to fund its side, credits the requested `amount` without measuring. Override it to measure the balance change and require equality, so a token outside the assumptions is refused at the deposit rather than discovered at the exit.

The pool's own token allow-list ([T0](./T0-ethereum-anchor.md)) governs which notes can exist. It does not cover `deposit`, which takes tokens straight from an address, so the adjudicator needs its own.

### I2 · Swaps are atomic

A swap changes two assets' allocations in one state. ForceMove accepts or rejects whole states, so a half-executed swap (one leg moved, the other not) cannot be represented on chain. Nothing in the contracts needs to coordinate the legs.

### I3 · No free option in a swap

Under `ConsensusApp` a state is valid only when every participant has signed it, so whoever signs last holds an option: they can countersign if the price is still good and withhold if it moved.

**Signing order fixes who holds it: the user signs the swap state first, the hub countersigns.** The hub quotes, so the hub bears the option it priced. If the hub withholds, the previous state stands and the user has lost nothing; this is denial of service, which [ADR-0015](./09-architecture-decisions.md#adr-0015) already accepts from a hub. The user can never hold a signed state the hub must honour later at a stale price.

Quote expiry is enforced by the hub refusing to countersign after it; states carry no timestamps, and none are needed.

### I4 · Amounts are token units; rates never reach the chain

Every channel amount is denominated in its token's own units: `waUSDC` shares, not USDC-equivalent. The exchange rate between USDC and `waUSDC` (`convertToAssets`) is read **off-chain, at quote time**, by both parties, and agreed by signing.

So no contract on the settlement path calls `convertToAssets`, `previewRedeem`, or any price source. There is no oracle to manipulate and no rounding mode to get wrong in a dispute. Yield accrues to whoever holds the shares when they are redeemed, which is the definition of holding a yield-bearing token.

### I5 · Each asset exits independently

A channel's exit must not let one asset's failure block another's. Causes that make one asset's payout revert:

- the pool refuses to shield that token (removed from its allow-list, or paused);
- the token itself pauses or blocks transfers (Circle's USDC blocklist; an Aave reserve pause freezing the StataToken).

go-nitro already provides per-asset exits: `transfer(assetIndex, ...)` pays one asset of a finalized channel. `transferAllAssets` and `concludeAndTransferAllAssets` pay every asset in one call and revert as a whole. **A client's or watchtower's payout of one asset never depends on another asset's payout**, so one frozen asset never strands the others. `ArmadaNitroAdjudicator._executeSingleAssetExit` is already per asset; this is a client-behaviour requirement plus a test, not a contract change.

### I6 · Deposits cannot be re-entered

`depositFromNote` reads `holdings[asset][channelId]`, calls `pool.transact`, then writes `held + received`. A re-entrant deposit to the same channel and asset during that call would be overwritten (a lost update). Today this is closed only by assumption: the pool is our own contract and the tokens have no hooks. (This is the slither reentrancy finding on `depositFromNote`.)

Close it by construction with two measures:

- the allow-list (I1) excludes tokens with hooks;
- a reentrancy guard on every function that moves holdings (`deposit`, `depositFromNote`, `transfer`, `transferAllAssets`, `concludeAndTransferAllAssets`, `reclaim`).

### I7 · The hub's Aave position stays off the channel

The hub turns USDC into `waUSDC` (ERC-4626 `deposit`) and back (`redeem`) as its own inventory management, through `HubCustodian.execute` under its FROST group key. Users never depend on these calls:

- a user who holds `waUSDC` exits with a `waUSDC` note, a transferable token, whether or not Aave has liquidity to redeem it;
- a user who swaps to USDC exits with the hub's USDC inventory.

Aave's redemption liquidity (`maxRedeem` falls while the reserve is borrowed out) is therefore the hub's risk, never a condition on a user's exit.

The hub's calls follow ERC-4626 practice: an exact `forceApprove` of the deposited amount, and a minimum-output check on `deposit` (shares) and `redeem` (assets) against the preview taken when the operation was signed, since a FROST-signed operation can execute later than it was signed.

## Implementation

This page states what must hold. How the contracts enforce it lives with the contracts: armada-nitro's task **E16 · Adjudicator hardening for multi-asset hub channels** (`docs/tasks/E16-adjudicator-hardening.md`) turns I1, I4, I5 and I6 into concrete contract requirements and tests that fail before and pass after. I2 and I3 need no contract change; I7 concerns the hub's own operations.

One design choice belongs here rather than there: the asset allow-list is **immutable**, so adding an asset is a new adjudicator deployment. With four assets and channels that are short-lived relative to deployments, that is acceptable and keeps governance off the settlement path. A governed list is a later decision.

## What this does not cover

- **Borrowing.** A borrow needs liquidation rules and an on-chain price, which makes it a real ForceMove app (option 1). It stays v2.
- **Unilateral swap enforcement.** Option 2 swaps need both signatures. A swap routed across hubs through virtual channels needs the hash-locked swap pattern of [ADR-0015](./09-architecture-decisions.md#adr-0015).
- **Protocol fees on swaps.** The fee split (T4.3) is an allocation on the same state and needs no further mechanism.

## How it is tested

armada-nitro's E11 exercises I2, I4 and I7 on the fixturenet (Aave yield is real; `waUSDC` as a channel asset with an in-channel swap; `wstETH` as a channel asset; the hub's redeem). E16 tests I1, I5 and I6 at the contract level.

## Where this fits

- First games and dispute structure: [14](./14-shielded-games-over-nitro.md).
- Settlement rail: [A.0 overview](./A-nitro-on-railgun/A.0-overview.md), [A.5 deposit/payout contract](./A-nitro-on-railgun/A.5-deposit-payout-contract.md).
- Proving that hub yield came from the user's own shielded value: [16](./16-selective-disclosure.md).
- Decisions: [ADR-0011](./09-architecture-decisions.md#adr-0011) (USDC yield, see D17.6), [ADR-0013](./09-architecture-decisions.md#adr-0013) (brings the Aave yield adapter into scope), [ADR-0016](./09-architecture-decisions.md#adr-0016) (adjudicator and custodian), [ADR-0017](./09-architecture-decisions.md#adr-0017) (this design), [ADR-0019](./09-architecture-decisions.md#adr-0019) (contract requirements live with the contracts).
