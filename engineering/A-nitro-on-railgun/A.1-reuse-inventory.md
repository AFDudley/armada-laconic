# A.1 — Reuse Inventory

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md)

This inventory states what each component of the settlement rail is: reused as-is, reused then patched, or built new, each with a code citation.

## A.1.1 Shielded pool (T0.0) — built, own implementation

`PrivacyPool` (`armada-poc@102f17b8:contracts/privacy-pool/PrivacyPool.sol`) routes by `delegatecall` to `ShieldModule`, `TransactModule`, `MerkleModule`, `VerifierModule`. It is not a proxy. It matches Railgun's note/commitment/nullifier shape (spec-compatible, ADR-0014), not Railgun's deployed contracts. See [T0-ethereum-anchor.md](../T0-ethereum-anchor.md) for the full citation set.

## A.1.2 The `Transact` circuit (T0.1) — built, Phase-2 open

`armada-circuits@ef2682f:src/operations/Transact.circom` (167 lines) plus `merkle.circom` (55 lines), instantiated as 19 `(nIn,nOut)` shapes. Built on `circomlib ^2.0.5`. Groth16; Phase-1 reuses the community Perpetual Powers of Tau; Phase-2 has not run (`v1.0.0-ceremony` manifest: every shape `"status": "open"`).

## A.1.3 `go-nitro` adjudicator (T0.2) — reused, then patched

`armada-nitro@55939ef:src/nitro/` vendors `go-nitro` at the import commit `8edacc1` ("unmodified"). Six later commits patch it:

| Commit | Change |
|---|---|
| `5ff5eee` | Accept EIP-1271 participants; remove the owner-controlled L2↔L1 bridge stub. |
| `25f40ba` | Fix a guarantee-type check in `MultiAssetHolder.transfer` (wrong allocation index). |
| `592a7f3` | Credit deposits by the measured balance delta; revert on a short transfer. |
| `04b8431` | Add a hub-configured asset allow-list to the adjudicator. |
| `32ad5e5` | Add a per-channel reentrancy lock across six holdings-mutating entry points. |
| `8570389` | Pay out ERC-20 via `SafeERC20`. |

This is a patched adjudicator, not a vanilla deploy. `ArmadaNitroAdjudicator` (`src/ArmadaNitroAdjudicator.sol`) extends it and adds `depositFromNote` and a `SHIELD`-type payout branch. See [A.4](./A.4-adjudicator-integration.md).

## A.1.4 The deposit/payout boundary — built

`depositFromNote` and `_executeSingleAssetExit`'s `SHIELD` branch, both on `ArmadaNitroAdjudicator`, are the boundary. See [A.5](./A.5-deposit-payout-contract.md).

## A.1.5 `HubCustodian` — built, FROST

`armada-nitro@55939ef:src/HubCustodian.sol` verifies a group signature with `FROST.verifySignature` (`frost-secp256k1-evm`, pinned `65d03d0`). It authorizes a hot delegate by EIP-1271 with a timed revocation, and rotates its group key with `FROST.isValidPublicKey`.

## A.1.6 `armada-sdk` — built client-side primitives, no nitro code

`armada-sdk@db5757f` wraps the vendored Railgun engine for note/shield/transact primitives, and ships a Groth16 prover (`src/prover/`, `snarkjs` + WASM, worker-first). It has no code addressing the nitro adjudicator or `HubCustodian` (`git grep -ln "nitro\|Nitro\|HubCustodian"` returns nothing in `src`). No `exportDisclosure` API exists yet (selective disclosure is still in progress, A.0.5).

## A.1.7 Governance and revenue contracts — built

`armada-poc@102f17b8:contracts/governance/`: `ArmadaGovernor` (UUPS, five proposal types, a Security Council veto), `ArmadaTreasuryGov`, `TreasurySteward`, `ArmadaToken` (ERC20Votes, 12M fixed supply), `ArmadaWindDown`, `ArmadaRedemption`, `AdapterRegistry`, `RevenueCounter`, `RevenueLock`, `RevenueReserveDistributor`; `contracts/crowdfund/ArmadaCrowdfund`.

## A.1.9 Risks

- **The pool and the circuit are audit-critical.** Both are built and tested, but the Phase-2 ceremony has not run, so no deployed verifying key yet traces to a published, multi-party transcript.
- **The patched adjudicator is the audit surface, not a vanilla deploy.** The six `src/nitro` commits, not only `ArmadaNitroAdjudicator`'s own code, need audit coverage.

See also: [T0-ethereum-anchor.md](../T0-ethereum-anchor.md), `labbook/2026-10-19-docs-versus-code.md`, and `.scratch/rewrite-evidence/inv-InvPool.md`, `inv-InvNitroSdk.md` (in `armada-laconic`).
