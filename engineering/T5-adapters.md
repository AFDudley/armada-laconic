# T5 · Adapters

**Status:** partly built · tier doc · 2026-10-20
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T5

CCTP and Aave-v4 yield are built, in `armada-poc`. Swaps has no contract. None of the three uses Railgun's RelayAdapt/Cookbook recipe pattern, because no RelayAdapt contract exists anywhere in this codebase; each instead uses the pool's own `adaptContract`/`adaptParams` binding mechanism.

## T5.0a CCTP — built

`PrivacyPool`/`PrivacyPoolClient`/`CCTPHookRouter` implement cross-chain shield and unshield over Circle's CCTP v2. `atomicCrossChainUnshield` binds the destination tuple into the proof via `CCTPBindingLib`, so a relayer cannot redirect or resubmit a cross-chain unshield. Cross-chain is USDC-only. A CCTP mint arrives on the destination chain as a fresh shielded USDC note, handled by `PrivacyPool.handleReceiveFinalizedMessage`/`handleReceiveUnfinalizedMessage`.

## T5.0b Aave-v4 yield — built

`ArmadaYieldAdapter`/`ArmadaYieldVault` (`contracts/yield/`) wrap an Aave V4 Spoke. `ArmadaYieldAdapter.lendAndShield`/`redeemAndShield` are trustless lend/redeem through the adapt-contract pattern: `YieldAdaptParams` binds `npk`, an encrypted bundle, and a shield key (or a seven-argument form with a fee destination for redeem), and the adapter checks `boundParams.adaptContract == address(this)` and re-derives `adaptParams` before acting. The vault is a plain ERC-20 (not a proxy) wrapper, shares priced net of a pending protocol fee, with a high-water mark so a loss recovery is not double-charged. This is the mechanism for USDC yield today: a direct adapter recipe, with the amount public only at the transparent supply boundary. The LP-buffered rail (T4.5) is not built.

## T5.0c Swaps — not built

No swap contract exists anywhere in `armada-poc`'s `contracts/` tree.

## T5.1 Routing interface — not built

There is no T4 venue to route to or from, so no routing interface exists. CCTP and Aave-v4 yield are reached directly through the pool's `transact`/`shield` calls with the relevant `adaptContract` bound in, not through a venue intermediary.

## Audit surface

`ArmadaYieldAdapter`/`ArmadaYieldVault` and `CCTPHookRouter` are real, deployed contracts moving user value through an external protocol (Aave, Circle's CCTP). They are gated by `IAdapterRegistry.authorizedAdapters`/`withdrawOnlyAdapters`, managed by the non-upgradeable, timelock-owned `AdapterRegistry`. This is the real audit surface for this tier today, not a future shared-dependency concern.

## Sources

- `armada-poc@102f17b8:contracts/cctp/CCTPHookRouter.sol`, `contracts/privacy-pool/PrivacyPool.sol:166,196,224`, `contracts/privacy-pool/CCTPBindingLib.sol`.
- `armada-poc@102f17b8:contracts/yield/ArmadaYieldAdapter.sol:27-40,146-158,189-199,286-298`, `contracts/yield/ArmadaYieldVault.sol:48-57,88-99,258-266,408-422`, `contracts/yield/YieldAdaptParams.sol:30-38,56-64`.
- `armada-poc@102f17b8:contracts/governance/AdapterRegistry.sol:16-78`.
- `.scratch/rewrite-evidence/cmp-CmpArc42.md`, `inv-InvPool.md` (in `armada-laconic`).
- Registry: [`05-building-block-view.md`](./05-building-block-view.md); the venue this tier does not yet route through: [`T4-execution-venue.md`](./T4-execution-venue.md).
