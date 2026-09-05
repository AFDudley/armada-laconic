# A.2 — Pool Deployment

work package A · reuse-oriented spec · 2026-09-05
**Parent:** [`A.0`](A.0-overview.md)
**Owns:** T0.0 — clean-room reimplement the shielded pool (spec-compatible with Railgun) and deploy it under our control

---

## A.2.1 Goal

Stand up **our own** shielded pool — the settlement substrate every other A item binds to — **clean-room reimplemented, spec-compatible with Railgun's design** ([ADR-0002](../09-architecture-decisions.md#adr-0002), [ADR-0014](../09-architecture-decisions.md#adr-0014)) and owned by us, not rented from Railgun's live deployment. T0.0 delivers a running pool on the laconic fixturenet configured with **fee = 0**, a token allow-list that admits USDC, a preserved SNARK-safety configuration, and a Groth16 verifier registry populated with **our own** verification keys from the T0.1 ceremony (→ [A.3](A.3-trusted-setup.md)). The end-state proof is a **shield → transact-unshield roundtrip that verifies against our registered vkeys** (A.2.6).

The pool is **clean-room reimplemented** ([ADR-0014](../09-architecture-decisions.md#adr-0014)), **not** reused as-deployed OSS: T0.0 authors **spec-compatible protocol contracts** (matching A.1.1, using no Railgun-licensed source) *and* the deploy/governance around them. This is an **audit-critical** crypto-engineering artifact, not config-only. The net-new artifacts are the pool contract stack itself, the governance transactions (fee, vkeys), and ownership of the upgradeable-proxy admin (A.2.4). What the pool exposes downstream — addresses, event ABI, shield/unshield entrypoints, the reserved deposit/payout hook — is the ICD in A.2.5.

## A.2.2 Boundary

**In scope (T0.0):**
- **Clean-room author + deploy** the upgradeable proxy → `RailgunSmartWallet` (is `RailgunLogic` is `{Commitments, TokenBlocklist, Verifier}`) — spec-compatible with the stack cited in A.1.1.
- Own the proxy admin / owner keys (A.2.4).
- Governance config: `changeFee(0,0,0)`; leave USDC **off** `TokenBlocklist`; preserve `snarkSafetyVector` / `checkSafetyVectors` magic constants.
- Register one verification key per chosen `(nIn,nOut)` circuit combo via `Verifier.setVerificationKey`, consuming T0.1 ceremony output (→ A.3).
- **Pin a Railgun commit** before any of the above (A.1.10 / A.2.7).

**Out of scope (elsewhere):**
- The ceremony that *produces* the vkeys and the 91-vs-54 combo-subset reconciliation → [A.3](A.3-trusted-setup.md) (T0.1). T0.0 only *registers* what A.3 hands over.
- The deposit/payout contract itself → [A.5](A.5-deposit-payout-contract.md) (T0.3). T0.0 exposes the shield/unshield entrypoints it binds to; it does not build T0.3.
- **POI is not a pool setting.** POI is an alongside partner stack; its "config" is a client / POI-node-layer choice of list providers + standby, and gated entry is enforced at the settlement layer (→ [A.7](A.7-anonymity-set.md)). T0.0 sets **no** on-chain POI switch (A.1.7).
- Index/scan consumers of pool events → B (T1.2) and T6.1; T0.0 only publishes the event ABI (A.2.5).

**Phase-0 (licensing gate — RESOLVED via clean-room, [ADR-0014](../09-architecture-decisions.md#adr-0014)).** Per A.1.10 / [A.0 gate 1](A.0-overview.md#a05-open-gates-must-resolve-before--during-a): Railgun's on-chain contracts are SPDX `UNLICENSED` and `circuits-v2` carries an explicit "*No License is provided… under any circumstances*" file, so redeploying them was never executable. **Resolution:** T0.0 **clean-room reimplements** the pool spec-compatible with A.1.1 (and T0.1 the circuits, A.3) — no grant/relicense needed. (`Railgun-Community/engine` and `cookbook` are MIT reference, but they are not the on-chain pieces T0.0 implements.)

## A.2.3 Reuse inventory (cite A.1 + pinned commit)

Everything below is the **reference spec** T0.0 clean-room reimplements — spec-compatible with `Railgun-Privacy/contract`, **not** reused-as-deployed OSS ([ADR-0014](../09-architecture-decisions.md#adr-0014)); cite [A.1.1](A.1-reuse-inventory.md#a11-shielded-pool-t00--clean-room-reference-spec) for pinned file/line citations. **The Railgun reference is UNPINNED** (read at `master`/`main` HEAD 2026-09-05); T0.0 MUST pin the reference commit before build (A.1.10, A.0.6).

| Spec-referenced piece | A.1.1 citation | Role in T0.0 |
|---|---|---|
| Proxy → `RailgunSmartWallet` (`RailgunLogic` ⊇ `{Commitments, TokenBlocklist, Verifier}`) | `RailgunSmartWallet.sol` L23–224 | The full stack we deploy and own. |
| `shield(ShieldRequest[])` | `RailgunSmartWallet.sol` L23–97 | Payout-out entrypoint (→ A.5 / T0.3); emits `Shield`. |
| `transact(Transaction[])` | `RailgunSmartWallet.sol` L102–224 | Unshield-in entrypoint; `boundParams.unshield=NORMAL`; emits `Transact`, `Nullified`. |
| `changeFee(shield,unshield,nft)` | `RailgunLogic.sol` L146–161 | **fee=0 lever**: `changeFee(0,0,0)` (A.2.4). |
| `snarkSafetyVector` / `checkSafetyVectors` | `RailgunLogic.sol` L111–127 | Magic-constant guard; MUST be preserved (A.2.4) or `transact` reverts. |
| `Verifier.verificationKeys[nIn][nOut]`, `setVerificationKey` | `Verifier.sol` L27, L36–45; `Snark.sol` L157–188 | Registry T0.0 populates with T0.1 vkeys (→ A.3). |
| Events `Action/Transact/Shield/Unshield/Nullified` | `RailgunLogic.sol` L57–77 | Commitment-insert + nullifier-spend feed (A.2.5). |
| Merkle accumulator + nullifier set (`TREE_DEPTH=16`, `rootHistory`) | `Commitments.sol` L28–55, L108–252 | The tree T6.1 rebuilds; T0.3 proofs reference a historical root of it. |
| `TokenBlocklist` | `TokenBlocklist.sol` L33–73 | Leave USDC unblocked; unshield always allowed (A.2.4). |
| ABI structs (`ShieldRequest`, `CommitmentPreimage`, `TokenData`, `Transaction`, `BoundParams`, `SnarkProof`) | `Globals.sol` L21–122 | The encode surface exported to T0.3 + T6.2 (A.2.5). |

**RelayAdapt/Cookbook** (Railgun's own recipe layer) is **not** part of T0.0 — Armada settles via nitro-on-railgun, not RelayAdapt ([ADR-0004](../09-architecture-decisions.md#adr-0004)). It is named here only to disambiguate: the "deposit/payout hook" T0.0 reserves is the T0.3 contract binding (A.5), never a RelayAdapt recipe.

## A.2.4 Net-new delta (what T0.0 actually authors)

Per [ADR-0014](../09-architecture-decisions.md#adr-0014) (and [A.1.9 item 4](A.1-reuse-inventory.md#a19-net-new-deltas-what-a-actually-builds)), T0.0's net-new surface is now the **clean-room pool contracts** *and* their deployment + governance — an audit-critical build, not config-only:

1. **Clean-room author + own the pool contract stack.** Independently implement the proxy → `RailgunSmartWallet` / `RailgunLogic` / `{Commitments, TokenBlocklist, Verifier}` stack **spec-compatible with A.1.1** (no Railgun-licensed source, [ADR-0014](../09-architecture-decisions.md#adr-0014)), deploy it under keys we control, and hold upgrade + owner authority (this is the whole point of ADR-0002 — own upgrades, audit, fee, and POI policy). This contract set is the **audit-critical** net-new artifact of T0.0; downstream items assume the proxy address is stable and admin-owned.
2. **`changeFee(0,0,0)` governance action.** Call `changeFee(shieldFee, unshieldFee, nftFee)` = `(0,0,0)` (`RailgunLogic.sol` L146–161) so shield/unshield/NFT fees are zero — Armada earns from venue spread + watcher metering, not a pool skim (ADR-0002). This is an owner-gated transaction, part of the deploy runbook.
3. **Vkey registration.** For each chosen `(nIn,nOut)` combo (the subset reconciled in A.3 — **91 generated, not 54**; A.1.8), call `Verifier.setVerificationKey(nIn, nOut, vkey)` with the Phase-2 output of **our** ceremony (→ A.3). Until this is done, `transact`/`shield` proofs have no key to verify against.
4. **Token allow-list posture.** Deploy `TokenBlocklist` empty of USDC (and of any asset A/C settles) so shield admits it; unshield is always allowed regardless of blocklist (`TokenBlocklist.sol` L33–73).
5. **Reproduce SNARK-safety config.** Carry `snarkSafetyVector` and the `checkSafetyVectors` magic constants (matching the reference `RailgunLogic.sol` L111–127) **verbatim into our clean-room implementation**; a mismatch makes every `transact` revert. This is a *must-match* delta — explicitly noted because a clean-room build that regenerates or drops these constants breaks the pool.

**What T0.0 does NOT author:** no fee-taking logic, and **no on-chain POI switch** — POI is client/POI-node-layer policy (A.1.7 → A.7), and gated entry lives at the settlement layer (T0.3, A.5). Asserting an on-chain POI config here would be wrong.

## A.2.5 Interfaces (ICD)

What T0.0 exposes to sibling items and downstream work packages. Struct shapes are Railgun's `Globals.sol` L21–122 (A.1.1) — cited, not restated.

**(a) Deployment addresses** — published to the registry ([`../05-building-block-view.md`](../05-building-block-view.md)) once deployed:
- **Pool proxy** address (the `RailgunSmartWallet` behind it) — consumed by T0.3 (A.5), T6.1/T6.2 (A.6), and B/T1.2 indexers.
- **`Verifier`** address — for audit / re-registration; owner-gated.

**(b) Commitment-insert / nullifier-spend event ABI** → **B (T1.2)** indexes, **T6.1** scans (A.1.1, `RailgunLogic.sol` L57–77):
- `Shield` — new commitment leaves minted (payout / shielded receive).
- `Transact` — commitment outputs of a spend.
- `Nullified` — nullifiers consumed (spend / unshield-in).
- `Unshield` — ERC20 leaving the pool (the deposit-in leg observed publicly).
- `Action` — the umbrella accumulator event.
The Merkle root each event advances lives in `Commitments` (`rootHistory`, `TREE_DEPTH=16`); T6.1 rebuilds the tree, and T0.3 proofs must reference a historical root in that window (A.1.1).

**(c) Shield / unshield entrypoints + reserved deposit/payout hook** → **A.5 (T0.3)**:
- `shield(ShieldRequest[])` — the **payout-out** entrypoint T0.3 calls to mint fresh notes to beneficiaries (A.1.4, A.5).
- `transact(Transaction[])` with `boundParams.unshield=NORMAL` and `unshieldPreimage.npk = bytes32(uint160(T0.3 address))` — the **deposit-in** entrypoint: `transferTokenOut` (`RailgunLogic.sol` L318–364) lands the ERC20 in T0.3, which then escrows via `MultiAssetHolder.deposit` (A.1.4). The "reserved deposit/payout hook" is precisely this address-as-recipient convention — **not** a code hook inside the pool; T0.0 reserves nothing beyond a stable, allow-listed asset path and the pool's normal unshield-recipient semantics.

**(d) POI allow-list policy** → **A.7**: T0.0 sets **no** on-chain POI value. The exposed "policy" is the decision surface (choose list providers + standby period) resolved at the client / POI-node layer and enforced at the settlement layer (T0.3 gated entry). Documented here as an explicit *non-interface* so downstream items don't look for a pool setter that doesn't exist (A.1.7).

## A.2.6 Acceptance / verification

**Primary acceptance (the roundtrip):** on the laconic fixturenet, drive a full **shield → transact-unshield roundtrip** and confirm each leg **proves against our own registered vkeys** — i.e., `Verifier` verifies the Groth16 proof using keys set by *our* T0.1 ceremony, not Railgun's:

1. **Deploy + govern.** Proxy deployed under our admin; `changeFee(0,0,0)` mined and read back as zero; `snarkSafetyVector`/`checkSafetyVectors` match the pinned source; USDC absent from `TokenBlocklist`; one vkey per chosen combo registered in `Verifier` (A.2.4).
2. **Shield.** `shield([ShieldRequest…])` for USDC → `Shield` event emitted; a new commitment leaf appears; `rootHistory` advances (`Commitments.sol` L108–252).
3. **Transact-unshield.** `transact([...])` spending that note with `boundParams.unshield=NORMAL` → `Verifier` accepts the proof **against our registered vkey for that `(nIn,nOut)`**; `Transact` + `Nullified` emitted; the ERC20 exits to the recipient. A tampered/wrong-key proof MUST revert — this is the discriminating check that we register and enforce *our* keys.
4. **Event/ABI conformance.** The emitted `Shield/Transact/Nullified/Unshield` decode against the A.2.5(b) ABI — the contract exported to B/T1.2 and T6.1.

**Fixturenet integration hook (walking skeleton, [A.0.4](A.0-overview.md#a04-walking-skeleton-as-first-integration-target)):** this roundtrip is the pool half of `shield → unshield-in → deposit → settle → conclude → shield`. T0.0 acceptance is the shield/unshield legs proving against our vkeys; the deposit/settle/conclude legs are A.4/A.5. Full end-to-end acceptance is owned by [A.8](A.8-interfaces-acceptance.md).

**Explicitly not acceptance for T0.0:** ceremony soundness (A.3), deposit escrow / channel settle (A.4/A.5), POI enforcement (A.7). Registering vkeys is T0.0; *producing* them is T0.1.

## A.2.7 Risks / open

- **✅ Licensing (Phase-0) — RESOLVED via clean-room ([ADR-0014](../09-architecture-decisions.md#adr-0014)).** T0.0 clean-room reimplements the pool (and T0.1 the circuits) spec-compatible with A.1.1/A.1.2 — no Railgun grant/relicense required, so this no longer blocks build. The residual work is that clean-room implementation being audit-critical (A.2.4).
- **Unpinned Railgun commit.** Contracts + circuits were read at HEAD, no SHA pinned. **Pin a Railgun commit before build** (A.0.6); the pinned SHA drives which `snarkSafetyVector` constants and struct layouts T0.0 carries. (go-nitro @435eb2b, ts-nitro @884d616, mobymask @2329198 are already pinned; Railgun is the outstanding one.)
- **Combo-subset dependency on A.3.** T0.0 can only register the vkeys A.3 reconciles (91 generated vs the cited ~54 registered subset; A.1.8). Registering the wrong subset means some `(nIn,nOut)` `transact`/`shield` shapes have no key and revert. Sequence T0.0 vkey registration after A.3 fixes the subset.
- **Safety-vector regression.** A clean-room build that regenerates rather than **reproduces** `snarkSafetyVector`/`checkSafetyVectors` (matching the reference `RailgunLogic.sol` L111–127) silently breaks `transact`. Treat these as pinned constants, and cover them in A.2.6 step 1.
- **Proxy-admin key custody.** Owning the upgrade admin (A.2.4) concentrates upgrade + fee + vkey authority; key-management/governance for the admin is an operational risk to flag to the settlement client / watcher operators (A.6), out of T0.0 code scope.
- **POI mis-modeling.** Repeated because it's a common error: there is no on-chain POI setter to reimplement; POI is a separate partner stack and a settlement-layer policy (A.1.7 → A.7). Don't add a pool config for it.

→ Ceremony that feeds A.2.4 vkey registration: [A.3](A.3-trusted-setup.md). Contract that binds A.2.5(c): [A.5](A.5-deposit-payout-contract.md). Deep-dive citations: [A.1.1](A.1-reuse-inventory.md#a11-shielded-pool-t00--clean-room-reference-spec), [A.1.10](A.1-reuse-inventory.md#a110-licensing--risks).
