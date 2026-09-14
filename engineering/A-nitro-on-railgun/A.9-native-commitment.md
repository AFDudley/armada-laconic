# A.9 — Native channelized commitment (amount-privacy-in-play)

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T0.6 *(optional; Phase-4)*

---

## A.9.1 Goal

Armada needs **amount-privacy-in-play**: the *value* of an in-flight
settlement should stay hidden not only off-chain but even when a channel is
force-closed on-chain. In the v1 spine (Design A,
[ADR-0005](../09-architecture-decisions.md#adr-0005)) amounts are public at the
transparent deposit/payout boundary (see [A.5](./A.5-deposit-payout-contract.md))
and become fully cleartext on dispute, because `MultiAssetHolder.holdings` and
the `ExitFormat.Allocation.amount` fields that finalize a channel are plain
`uint256` on the public chain. Native channelized commitment, the "fork-lite"
upgrade (T0.6), closes that last leak. It makes channel allocations reference
*hidden-amount commitments* rather than cleartext `uint256`, so a force-close
reveals no amount.

This is a circuit change, and it sits off the v1 critical path.
The decision of record (ADR-0005) is to ship Design A first, defer fork-lite to
Phase-4, and exclude fork-full (shielded ForceMove, research-grade) unless
separately greenlit. This document specifies T0.6 as a scoped, optional upgrade
so it can be greenlit cleanly later; it does not authorise build now.

## A.9.2 Boundary

Three concentric privacy postures; T0.6 moves the amount-leak inward by exactly
one boundary.

| Posture | Amounts hidden off-chain? | Amounts hidden at deposit/payout? | Amounts hidden on force-close? | Status |
|---|---|---|---|---|
| **Design A** (v1 default) | yes (in-channel) | **no** — public at the T0.3 boundary (→ A.5) | **no** — cleartext in the on-chain outcome | shipping |
| **Fork-lite** (T0.6, this doc) | yes | still public at the boundary hop | **yes** — allocation→note boundary hides amounts across dispute | **deferred, Phase-4** |
| **Fork-full** | yes | yes (shielded ForceMove) | yes | **excluded** (ADR-0005) |

- **In scope (T0.6):** hide amounts across the allocation-to-note boundary, the
  encoding by which a channel outcome maps to Railgun commitments, so that
  `challenge`/`conclude` expose no cleartext value. The transparent boundary hop
  itself, the moment value crosses in or out of the shielded pool at T0.3, is
  unchanged, and Design A's public boundary is retained.
- **Out of scope (T0.6):** shielding the ForceMove state machine, hiding
  *participants* or *channel topology*, and native-ETH support (a related but
  distinct net-new item, A.1.9 item 7; see A.9.4). Fork-full is excluded.
- **Asset boundary:** the pool is ERC20/721 only, and ERC1155 is unsupported
  (A.1.10). T0.6 inherits that bound and must not assume multi-standard notes.

## A.9.3 Reuse inventory (cite A.1 + pinned commits)

T0.6 reuses the *same* two systems as the spine and changes only the
note/commitment cryptography that binds them — see A.1 for pinned citations.

| Reused piece | Source (A.1 ref) | Role in T0.6 |
|---|---|---|
| JoinSplit circuit — note/commitment/nullifier format, `sumIn===sumOut`, range `<2^120` | A.1.2 (`circuits-v2` `src/library/joinsplit.circom` L11–118, **reference — pin before build**) | Our own JoinSplit format (spec-compatible with A.1.2) is what T0.6 extends; the sum-check is the invariant the hidden-amount variant must preserve. |
| Nullifier derivation `Poseidon(nullifyingKey, leafIndex)` | A.1.2 (`nullifier-check.circom` L11–13) | Unchanged; hidden-amount notes still nullify identically. |
| Circuit-combo generator (**91** combos) + ceremony driver | A.1.2 (`lib/circuitConfigs.js`; `scripts/prepare_ceremony` L23–79) | The **Phase-2 machinery re-run** for the changed circuit set (→ A.3). |
| Groth16 verifier registry `verificationKeys[nIn][nOut]`, `setVerificationKey` | A.1.1 (`Verifier.sol` L27, L36–45) | New/updated vkeys registered post-ceremony (→ A.2/A.3). |
| Outcome encoding `SingleAssetExit`/`Allocation`/`AllocationType` | A.1.3 (exit-format `ExitFormat.sol` L26–73) | The wire format whose `amount`/`metadata` fields T0.6 repurposes to carry a **commitment** instead of cleartext value. |
| `MultiAssetHolder.deposit` / `concludeAndTransferAllAssets` | A.1.3 (`MultiAssetHolder.sol` L36–70; `NitroAdjudicator.sol` L33–42) | Escrow + finalize; T0.6 changes *what the allocation means*, not these entrypoints. |
| T0.3 deposit/payout contract | A.1.4 / [A.5](./A.5-deposit-payout-contract.md) | The public-side counterparty that must learn the new allocation↔note encoding on payout. |

**Pinning:** go-nitro **@435eb2b**, ts-nitro **@884d616**, mobymask **@2329198**
are pinned; **Railgun (`circuits-v2`, `contract`) is UNPINNED — pin a commit
before any T0.6 build** (A.1.10).

## A.9.4 Net-new delta (what T0.6 actually builds)

This is A.1.9 item 7 (opt), narrowed to fork-lite.

1. **Hidden-amount commitment format.** Extend the JoinSplit note/commitment
   format so a channel *allocation* references a hidden-amount commitment (a
   Pedersen or Poseidon-style value commitment) rather than a cleartext
   `uint256`. The public outcome carries the commitment; the amount is a private
   witness.
2. **Allocations-sum circuit constraint.** Add an in-circuit proof that the
   committed allocation amounts sum to the committed channel deposit, the
   fork-lite analogue of JoinSplit's `sumIn===sumOut` (A.1.2). This ensures a
   hidden-amount force-close cannot create value, and it is the core new
   constraint.
3. **Circuit change requires a ceremony re-run.** Any change to the circuit set
   forces a fresh Phase-2 MPC over the changed circuits
   ([ADR-0003](../09-architecture-decisions.md#adr-0003), ADR-0005). Re-run T0.1
   Phase-2 and register new or updated vkeys (see [A.3](./A.3-trusted-setup.md),
   A.2). Phase-1 (`powersOfTau28_hez_final_20.ptau`, 2²⁰) is still inherited.
4. **T0.3 payout adaptation.** T0.3 must decode the new allocation-to-note
   encoding and re-shield against hidden-amount commitments rather than cleartext
   exit amounts (see [A.5](./A.5-deposit-payout-contract.md)).
5. **Fresh audit and heavier mobile proving.** New circuits invalidate the v1
   audit and enlarge the proving witness, so they require a fresh security audit
   and heavier mobile proving. A wider or deeper circuit costs the D-owned mobile
   prover (T6.6). Both are ADR-0005 consequences.

**Related net-new (adjacent, not T0.6):** the pool is ERC20/721 only, with no
native ETH (A.1.10). Native-ETH settlement needs either a WETH-wrap path at the
T0.3 boundary or a native path, which is A.1.9 item 7's other half. It is
separable from fork-lite, since Design A can gain WETH-wrap without any circuit
change, and should be scoped independently; T0.6 does not deliver it.

## A.9.5 Interfaces (ICD)

T0.6 is inward-facing: it changes cryptography and the T0.3 payout encoding,
not the A-to-B/C/D surface shape.

- **Circuit set to T0.1/T0.0.** The changed `(nInputs,nOutputs)` circuit set (a
  superset or variant of the reconciled combo subset, A.3) plus its Phase-2
  transcript and per-circuit vkeys, registered via `setVerificationKey` (A.1.1).
  Consumers: A.2 (registration), A.3 (ceremony).
- **Allocation encoding to T0.3.** A revised mapping between allocations and
  notes: outcome `Allocation` fields carry a hidden-amount commitment, plus any
  `withdrawHelper` or `metadata` payload, instead of cleartext `uint256`.
  Consumer: [A.5](./A.5-deposit-payout-contract.md) (payout decode and
  re-shield). The `MultiAssetHolder`/`NitroAdjudicator` entrypoints (A.1.3) are
  unchanged.
- **Note format to T6.1 scanner / B.** Hidden-amount output commitments must
  still be *scannable* (same nullifier derivation, A.1.2) so B's note-scanner
  (T6.1) and indexers keep working; only the value field's opacity changes.
- No new on-chain governance surface is introduced beyond the vkey
  re-registration Design A already performs (A.2).

## A.9.6 Acceptance / verification

Because T0.6 is optional and deferred, its gate is a lighter PoC rather than a
full-spine acceptance run. Greenlight is a Phase-4 decision, and this PoC
de-risks that decision.

1. **Force-close reveals no cleartext amount.** Drive a hidden-amount channel to
   an adversarial `challenge`/`conclude` on a fixturenet and confirm the on-chain
   outcome (`ExitFormat` allocations, emitted events, `holdings`) exposes no
   cleartext value, only commitments. *Fail: any recoverable plaintext amount.*
2. **Allocations-sum proof rejects value creation.** Feed the circuit an outcome
   whose committed allocations do not sum to the committed deposit, and confirm
   that proof generation and verification reject it (the value-conservation
   invariant, A.9.4 item 2). *Fail: a value-creating outcome verifies.*
3. **Payout round-trips.** A hidden-amount channel finalizes, T0.3 decodes the
   new allocation encoding, then re-shields fresh notes whose scanned balances
   match the private intended split (see A.5). *Fail: amount mismatch or
   unscannable note.*
4. **Ceremony/audit prerequisites named, not run.** The PoC MAY use a throwaway
   Phase-2 over the changed circuits. Production greenlight is contingent on a
   real T0.1 Phase-2 re-run and a fresh audit (A.9.4 items 3, 5), recorded as the
   Phase-4 entry criteria and not exercised at PoC.

## A.9.7 Risks / open

- **Cost vs benefit (ADR-0005).** Fork-lite adds ceremony, audit, and prover
  cost for a privacy gain the v1 feature set does not require, hence its
  deferral. Revisit only if a concrete requirement emerges, such as large-value
  settlement de-anonymisation on dispute.
- **Mobile proving budget (T6.6).** Heavier circuits may exceed the D-owned
  mobile prover's practical bound; quantify witness and proving-time growth in
  the PoC before greenlight.
- **Gate is cost and audit.** T0.6 edits our own circuits and re-runs Phase-2 over them (T0.0/T0.1 are our own implementation, [A.2](./A.2-pool-deployment.md), [A.3](./A.3-trusted-setup.md)). Fork-lite's gate is cost and audit (A.9.1, A.9.7).
- **Unpinned Railgun commit.** Pin `circuits-v2`/`contract` before build (A.1.10).
- **Native-ETH is separable.** Do not let the WETH-wrap/native-ETH item ride on
  the T0.6 greenlight; scope it independently (A.9.4).
- **Fork-full** stays excluded unless separately greenlit (ADR-0005).

---

See also: [A.5](./A.5-deposit-payout-contract.md) (payout adaptation), [A.3](./A.3-trusted-setup.md) (ceremony re-run), [A.1.9](./A.1-reuse-inventory.md)/[A.1.10](./A.1-reuse-inventory.md) (deltas & risks), [ADR-0005](../09-architecture-decisions.md#adr-0005), [ADR-0003](../09-architecture-decisions.md#adr-0003).
