# A.2 — Pool Deployment

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](A.0-overview.md)
**Owns:** T0.0 — author our own shielded pool (spec-compatible with Railgun) and deploy it under our control

---

## A.2.1 Goal

Stand up our own shielded pool, the settlement substrate every other A item binds to. It is our own implementation, spec-compatible with Railgun's design ([ADR-0002](../09-architecture-decisions.md#adr-0002), [ADR-0014](../09-architecture-decisions.md#adr-0014)), and owned by us rather than rented from Railgun's live deployment. T0.0 delivers a running pool on the laconic fixturenet configured with fee = 0, a token allow-list that admits USDC, a preserved SNARK-safety configuration, and a Groth16 verifier registry populated with our own verification keys from the T0.1 ceremony (see [A.3](A.3-trusted-setup.md)). The end-state proof is a shield then transact-unshield roundtrip that verifies against our registered vkeys (A.2.6).

The pool is our own implementation of the Railgun design ([ADR-0014](../09-architecture-decisions.md#adr-0014)), not reused as-deployed OSS. T0.0 authors spec-compatible protocol contracts, matching A.1.1, along with the deploy and governance around them. This is an **audit-critical** crypto-engineering artifact, not config-only. The net-new artifacts are the pool contract stack itself, the governance transactions for fee and vkeys, and ownership of the upgradeable-proxy admin (A.2.4). The ICD in A.2.5 describes what the pool exposes downstream: addresses, event ABI, shield/unshield entrypoints, and the reserved deposit/payout hook.

## A.2.2 Boundary

**In scope (T0.0):**
- Author and deploy the upgradeable proxy to `RailgunSmartWallet` (is `RailgunLogic` is `{Commitments, TokenBlocklist, Verifier}`), spec-compatible with the stack cited in A.1.1.
- Own the proxy admin / owner keys (A.2.4).
- Governance config: `changeFee(0,0,0)`; leave USDC off `TokenBlocklist`; preserve `snarkSafetyVector` / `checkSafetyVectors` magic constants.
- Register one verification key per chosen `(nIn,nOut)` circuit combo via `Verifier.setVerificationKey`, consuming T0.1 ceremony output (see A.3).
- Pin a Railgun commit before any of the above (A.1.10 / A.2.7).

**Out of scope (elsewhere):**
- The ceremony that *produces* the vkeys, and the 91-vs-54 combo-subset reconciliation, belong to [A.3](A.3-trusted-setup.md) (T0.1). T0.0 only *registers* what A.3 hands over.
- The deposit/payout contract itself belongs to [A.5](A.5-deposit-payout-contract.md) (T0.3). T0.0 exposes the shield/unshield entrypoints it binds to; it does not build T0.3.
- POI is not a pool setting. POI is an alongside partner stack; its "config" is a client- or POI-node-side choice of list providers and standby, and gated entry is enforced in the settlement path (see [A.7](A.7-anonymity-set.md)). T0.0 sets no on-chain POI switch (A.1.7).
- Consumers that index and scan pool events belong to B (T1.2) and T6.1; T0.0 only publishes the event ABI (A.2.5).

## A.2.3 Reuse inventory (cite A.1 + pinned commit)

Everything below is the **reference spec** T0.0 implements, spec-compatible with `Railgun-Privacy/contract`, not reused-as-deployed OSS ([ADR-0014](../09-architecture-decisions.md#adr-0014)); cite [A.1.1](A.1-reuse-inventory.md#a11-shielded-pool-t00-reference-spec) for pinned file/line citations. The Railgun reference is UNPINNED, read at `master`/`main` HEAD 2026-09-05, so T0.0 MUST pin the reference commit before build (A.1.10, A.0.6).

| Spec-referenced piece | A.1.1 citation | Role in T0.0 |
|---|---|---|
| Proxy → `RailgunSmartWallet` (`RailgunLogic` ⊇ `{Commitments, TokenBlocklist, Verifier}`) | `RailgunSmartWallet.sol` L23–224 | The full stack we deploy and own. |
| `shield(ShieldRequest[])` | `RailgunSmartWallet.sol` L23–97 | Payout-out entrypoint (→ A.5 / T0.3); emits `Shield`. |
| `transact(Transaction[])` | `RailgunSmartWallet.sol` L102–224 | Unshield-in entrypoint; `boundParams.unshield=NORMAL`; emits `Transact`, `Nullified`. |
| `changeFee(shield,unshield,nft)` | `RailgunLogic.sol` L146–161 | fee=0 lever: `changeFee(0,0,0)` (A.2.4). |
| `snarkSafetyVector` / `checkSafetyVectors` | `RailgunLogic.sol` L111–127 | Magic-constant guard; MUST be preserved (A.2.4) or `transact` reverts. |
| `Verifier.verificationKeys[nIn][nOut]`, `setVerificationKey` | `Verifier.sol` L27, L36–45; `Snark.sol` L157–188 | Registry T0.0 populates with T0.1 vkeys (→ A.3). |
| Events `Action/Transact/Shield/Unshield/Nullified` | `RailgunLogic.sol` L57–77 | Commitment-insert + nullifier-spend feed (A.2.5). |
| Merkle accumulator + nullifier set (`TREE_DEPTH=16`, `rootHistory`) | `Commitments.sol` L28–55, L108–252 | The tree T6.1 rebuilds; T0.3 proofs reference a historical root of it. |
| `TokenBlocklist` | `TokenBlocklist.sol` L33–73 | Leave USDC unblocked; unshield always allowed (A.2.4). |
| ABI structs (`ShieldRequest`, `CommitmentPreimage`, `TokenData`, `Transaction`, `BoundParams`, `SnarkProof`) | `Globals.sol` L21–122 | The encode surface exported to T0.3 + T6.2 (A.2.5). |

RelayAdapt and Cookbook (Railgun's own recipe layer) are not part of T0.0; Armada settles via nitro-on-railgun, not RelayAdapt ([ADR-0004](../09-architecture-decisions.md#adr-0004)). They are named here only to disambiguate: the "deposit/payout hook" T0.0 reserves is the T0.3 contract binding (A.5), never a RelayAdapt recipe.

## A.2.4 Net-new delta (what T0.0 actually authors)

Per [ADR-0014](../09-architecture-decisions.md#adr-0014) (and [A.1.9 item 4](A.1-reuse-inventory.md#a19-net-new-deltas-what-a-actually-builds)), T0.0's net-new surface is the **pool contracts** together with their deployment and governance, an audit-critical build, not config-only:

1. **Author and own the pool contract stack.** Implement the proxy to `RailgunSmartWallet` / `RailgunLogic` / `{Commitments, TokenBlocklist, Verifier}` stack, spec-compatible with A.1.1 ([ADR-0014](../09-architecture-decisions.md#adr-0014)); deploy it under keys we control; and hold upgrade and owner authority. This is the whole point of ADR-0002: own upgrades, audit, fee, and POI policy. This contract set is the **audit-critical** net-new artifact of T0.0, and downstream items assume the proxy address is stable and admin-owned.
2. **`changeFee(0,0,0)` governance action.** Call `changeFee(shieldFee, unshieldFee, nftFee)` with `(0,0,0)` (`RailgunLogic.sol` L146–161) so shield, unshield, and NFT fees are zero; Armada earns from venue spread and watcher metering, not a pool skim (ADR-0002). This is an owner-gated transaction, part of the deploy runbook.
3. **Vkey registration.** For each chosen `(nIn,nOut)` combo (the subset reconciled in A.3, with 91 generated, not 54; A.1.8), call `Verifier.setVerificationKey(nIn, nOut, vkey)` with the Phase-2 output of our ceremony (see A.3). Until this is done, `transact`/`shield` proofs have no key to verify against.
4. **Token allow-list posture.** Deploy `TokenBlocklist` empty of USDC (and of any asset A/C settles) so shield admits it; unshield is always allowed regardless of blocklist (`TokenBlocklist.sol` L33–73).
5. **Reproduce SNARK-safety config.** Carry `snarkSafetyVector` and the `checkSafetyVectors` magic constants (matching the reference `RailgunLogic.sol` L111–127) verbatim into our implementation; a mismatch makes every `transact` revert. This is a must-match delta, noted because a build that regenerates or drops these constants breaks the pool.

**What T0.0 does not author:** no fee-taking logic, and no on-chain POI switch. POI is client- or POI-node-side policy (A.1.7; see A.7), and gated entry lives in the settlement path (T0.3, A.5). Asserting an on-chain POI config here would be wrong.

## A.2.5 Interfaces (ICD)

What T0.0 exposes to sibling items and downstream work packages. Struct shapes are Railgun's `Globals.sol` L21–122 (A.1.1); they are cited, not restated.

**(a) Deployment addresses**, published to the registry ([`../05-building-block-view.md`](../05-building-block-view.md)) once deployed:
- Pool proxy address (the `RailgunSmartWallet` behind it), consumed by T0.3 (A.5), T6.1/T6.2 (A.6), and B/T1.2 indexers.
- `Verifier` address, for audit and re-registration; owner-gated.

**(b) Commitment-insert / nullifier-spend event ABI**, which B (T1.2) indexes and T6.1 scans (A.1.1, `RailgunLogic.sol` L57–77):
- `Shield` — new commitment leaves minted (payout / shielded receive).
- `Transact` — commitment outputs of a spend.
- `Nullified` — nullifiers consumed (spend / unshield-in).
- `Unshield` — ERC20 leaving the pool (the deposit-in leg observed publicly).
- `Action` — the umbrella accumulator event.
The Merkle root each event advances lives in `Commitments` (`rootHistory`, `TREE_DEPTH=16`); T6.1 rebuilds the tree, and T0.3 proofs must reference a historical root in that window (A.1.1).

**(c) Shield / unshield entrypoints plus reserved deposit/payout hook**, for A.5 (T0.3):
- `shield(ShieldRequest[])` is the payout-out entrypoint T0.3 calls to mint fresh notes to beneficiaries (A.1.4, A.5).
- `transact(Transaction[])` with `boundParams.unshield=NORMAL` and `unshieldPreimage.npk = bytes32(uint160(T0.3 address))` is the deposit-in entrypoint. `transferTokenOut` (`RailgunLogic.sol` L318–364) lands the ERC20 in T0.3, which then escrows via `MultiAssetHolder.deposit` (A.1.4). The "reserved deposit/payout hook" is precisely this address-as-recipient convention, not a code hook inside the pool; T0.0 reserves nothing beyond a stable, allow-listed asset path and the pool's normal unshield-recipient semantics.

**(d) POI allow-list policy**, for A.7: T0.0 sets no on-chain POI value. The exposed "policy" is the decision surface — choose list providers and a standby period — resolved in the client or POI-node software and enforced in the settlement path (T0.3 gated entry). It is documented here as an explicit non-interface so downstream items do not look for a pool setter that does not exist (A.1.7).

## A.2.6 Acceptance / verification

**Primary acceptance (the roundtrip):** on the laconic fixturenet, drive a full shield then transact-unshield roundtrip and confirm each leg proves against our own registered vkeys. That is, `Verifier` verifies the Groth16 proof using keys set by our T0.1 ceremony, not Railgun's:

1. **Deploy and govern.** Proxy deployed under our admin; `changeFee(0,0,0)` mined and read back as zero; `snarkSafetyVector`/`checkSafetyVectors` match the pinned source; USDC absent from `TokenBlocklist`; one vkey per chosen combo registered in `Verifier` (A.2.4).
2. **Shield.** `shield([ShieldRequest…])` for USDC emits the `Shield` event; a new commitment leaf appears; `rootHistory` advances (`Commitments.sol` L108–252).
3. **Transact-unshield.** `transact([...])` spending that note with `boundParams.unshield=NORMAL` produces a proof that `Verifier` accepts against our registered vkey for that `(nIn,nOut)`; `Transact` and `Nullified` are emitted; the ERC20 exits to the recipient. A tampered or wrong-key proof MUST revert; this is the discriminating check that we register and enforce our keys.
4. **Event/ABI conformance.** The emitted `Shield/Transact/Nullified/Unshield` decode against the A.2.5(b) ABI, the contract exported to B/T1.2 and T6.1.

**Fixturenet integration hook (walking skeleton, [A.0.4](A.0-overview.md#a04-walking-skeleton-as-first-integration-target)):** this roundtrip is the pool half of `shield → unshield-in → deposit → settle → conclude → shield`. T0.0 acceptance is the shield/unshield legs proving against our vkeys; the deposit/settle/conclude legs are A.4/A.5. Full end-to-end acceptance is owned by [A.8](A.8-interfaces-acceptance.md).

**Explicitly not acceptance for T0.0:** ceremony soundness (A.3), deposit escrow / channel settle (A.4/A.5), POI enforcement (A.7). Registering vkeys is T0.0; *producing* them is T0.1.

## A.2.7 Risks / open

- **Own pool build is audit-critical.** T0.0 authors the pool ([ADR-0014](../09-architecture-decisions.md#adr-0014)), spec-compatible with A.1.1; the residual work is that this build is audit-critical (A.2.4), not that any grant is outstanding.
- **Unpinned Railgun commit.** Contracts and circuits were read at HEAD, with no SHA pinned. Pin a Railgun commit before build (A.0.6); the pinned SHA drives which `snarkSafetyVector` constants and struct layouts T0.0 carries. (go-nitro @435eb2b, ts-nitro @884d616, mobymask @2329198 are already pinned; Railgun is the outstanding one.)
- **Combo-subset dependency on A.3.** T0.0 can only register the vkeys A.3 reconciles (91 generated vs the cited ~54 registered subset; A.1.8). Registering the wrong subset means some `(nIn,nOut)` `transact`/`shield` shapes have no key and revert. Sequence T0.0 vkey registration after A.3 fixes the subset.
- **Safety-vector regression.** A build that regenerates rather than reproduces `snarkSafetyVector`/`checkSafetyVectors` (matching the reference `RailgunLogic.sol` L111–127) silently breaks `transact`. Treat these as pinned constants, and cover them in A.2.6 step 1.
- **Proxy-admin key custody.** Owning the upgrade admin (A.2.4) concentrates upgrade, fee, and vkey authority; key management and governance for the admin is an operational risk to flag to the settlement client and watcher operators (A.6), out of T0.0 code scope.
- **POI mis-modeling.** Repeated because it is a common error: there is no on-chain POI setter to reimplement; POI is a separate partner stack and a settlement-side policy (A.1.7; see A.7). Do not add a pool config for it.

See also: Ceremony that feeds A.2.4 vkey registration: [A.3](A.3-trusted-setup.md). Contract that binds A.2.5(c): [A.5](A.5-deposit-payout-contract.md). Deep-dive citations: [A.1.1](A.1-reuse-inventory.md#a11-shielded-pool-t00-reference-spec), [A.1.10](A.1-reuse-inventory.md#a110-risks).
