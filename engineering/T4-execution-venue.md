# T4 · Execution / venue

**Status:** not built · tier doc · 2026-10-20
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T4

No contract in `armada-poc` implements a venue. There is no posted-price contract, no quote/settle app, and no matcher. T4.4, the one item marked built in the registry, needs no venue at all: any non-blocklisted ERC-20, including wstETH, can already be shielded as a note, so ETH-denominated yield already works without a swap path.

The design reasoning below is unchanged from the earlier docs, because no evidence contradicts it; it is simply unimplemented. It states the plan, not a build in progress.

## T4.0 Posted-price contract — not built

The idea: a single privileged `priceSetter` posts bid/ask per pair, take-it-or-leave-it, migratable to governance later. No such contract exists.

## T4.1 Quote/settle ForceMove app — not built

The idea: a pure state machine (`propose-quote → accept → fill → settle/abort`) registered as a `go-nitro` ForceMove app. `armada-nitro`'s test suite (`ArmadaNitroAdjudicator.t.sol` and siblings) has no quote/settle app test. No such app exists.

## T4.2 Venue solver — not built

The idea: an off-chain keeper that internalizes flow, hedges the residual on public liquidity, and rebalances Aave on a decorrelated schedule. The earlier docs routed the hedge leg through Railgun's RelayAdapt/Cookbook recipe; no RelayAdapt contract exists anywhere in `armada-poc`, so any future hedge leg would need a different mechanism, likely the pool's own `adaptContract`/`adaptParams` binding, the same pattern `ArmadaYieldAdapter` already uses.

## T4.3 Fee-split allocation — not built

The idea: a co-signed channel allocation split across user, protocol, and integrator, cleared over `go-nitro`. No code implements this.

## T4.4 ETH yield — built, as a consequence of the pool's design

A shielded wstETH note earns staking yield just by being held, because the pool shields any non-blocklisted ERC-20. No separate contract is needed and none exists; this item needed no venue to begin with.

## T4.5 USDC yield (LP-buffered, private) — not built

The earlier design: a private taker swaps shielded USDC for the venue's shielded aUSDC at a posted price; the venue rebalances to and from Aave in aggregate on its own schedule. Today's built USDC yield path is different and simpler: `ArmadaYieldAdapter`/`ArmadaYieldVault` let a user shield directly into Aave-v4 yield through the pool's own `adaptContract` binding, with no LP-buffered privacy layer. The private LP-buffered design remains unbuilt.

## T4.6 ex_net matcher + LP vault — not built

The idea: a DSS-federated venue proving fair ordering, layered on T3. Neither T3 nor any part of this tier is built, so this item has no foundation yet to build on.

## Sources

- `armada-poc@102f17b8:contracts/` — subdirectories are `aave-mock, cctp, crowdfund, fees, governance, periphery, privacy-pool, test, yield`; no venue, matcher, or Swaps directory exists.
- `.scratch/rewrite-evidence/cmp-CmpArc42.md` (in `armada-laconic`) — T4 item status table.
- Registry: [`05-building-block-view.md`](./05-building-block-view.md); the built yield path: [`T5-adapters.md`](./T5-adapters.md), [`T0-ethereum-anchor.md`](./T0-ethereum-anchor.md).
