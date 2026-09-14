# A.7 — Anonymity-Set Strategy

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T0.7 (anonymity-set strategy — bootstrap own crowd + Railgun import bridge; POI binding)

---

## A.7.1 Goal

Make the shielded value that A settles actually private in practice, not merely private in principle. The base construction (A.0.1) hides sender, recipient, and amount in-circuit; what it cannot manufacture is the **crowd** those hidden transactions blend into. A fresh pool starts with an anonymity set near zero, which [ADR-0006](../09-architecture-decisions.md#adr-0006) records as *the biggest non-engineering risk* to A. T0.7 is therefore a **strategy item** (per [§5](../05-building-block-view.md), status "strategy", not net-new crypto). It specifies the base shielded-payments primitive A delivers, the cold-start bootstrap plan, the Railgun onboarding bridge that imports day-one liquidity, the POI binding that connects the pool to its alongside proof-of-innocence stack, and the concrete deltas and acceptance those imply.

Liquidity import is not anonymity-set inheritance (A.7.4): accepting Railgun's value is a public boundary hop, never native ingestion of their crowd.

## A.7.2 Boundary

- **In scope.** The shielded-payments primitive as a *capability statement* (it is inherited, not built, per A.7.3); bootstrap incentive and flywheel policy; the Railgun-to-ours import bridge script and flow (client-side orchestration of two existing entrypoints); POI-node integration and the settlement-side gated-entry policy; the ICD surface (POI allow-list root and policy) consumed by [A.5](./A.5-deposit-payout-contract.md); and privacy metrics.
- **Out of scope.** The pool contract itself and the `changeFee`/vkey levers; see [A.2](./A.2-pool-deployment.md). The circuit and ceremony that a cross-pool proof would need; see [A.3](./A.3-trusted-setup.md) and [A.9](./A.9-native-commitment.md). The deposit/payout contract internals that *enforce* gated-entry; see [A.5](./A.5-deposit-payout-contract.md). Native amount-privacy-in-play; see [A.9](./A.9-native-commitment.md).
- **Phase-0 assumption.** T0.7's bootstrap and bridge presume a live pool (T0.0, [A.2](./A.2-pool-deployment.md)) and ceremony (T0.1, [A.3](./A.3-trusted-setup.md)). The residual dependency is that those builds ship before the crowd or import path carries real value. Until they do, A.7 is a design and fixturenet exercise only.

## A.7.3 Reuse inventory (cite A.1 + pinned commits)

Everything T0.7 leans on already exists; T0.7 adds no cryptography. Cite, don't re-document.

| Piece | Source (see A.1) | What it gives T0.7 |
|---|---|---|
| **Shielded-payments primitive** = `transact()` with `unshield=NONE` | [A.1.7](./A.1-reuse-inventory.md) (primitive); `transact` in [A.1.1](./A.1-reuse-inventory.md), `RailgunSmartWallet.sol` L102–224 | The base capability A delivers **for free**: inputs nullified, outputs created, `sumIn==sumOut` in-circuit; **sender/recipient/amount hidden** — only `merkleRoot`, `boundParamsHash`, `nullifiers[]`, output-commitment hashes are public. This is a *native Railgun transfer*, no deposit/payout contract, no channel. |
| JoinSplit circuit (defines the primitive) | [A.1.2](./A.1-reuse-inventory.md), `src/library/joinsplit.circom` L11–118 | Fixes note/commitment/nullifier format; proves EdDSA ownership + Merkle membership + range + value conservation. The privacy set = the commitments in the accumulator this circuit proves membership of. |
| Merkle accumulator + nullifier set (`TREE_DEPTH=16`, `rootHistory`) | [A.1.1](./A.1-reuse-inventory.md), `Commitments.sol` L28–55, L108–252 | **The anonymity set, concretely**: every shielded commitment ever inserted. Its cardinality (and recent activity) is what T0.7 metrics measure. |
| `shield(ShieldRequest[])` / `transact()` unshield | [A.1.1](./A.1-reuse-inventory.md), `RailgunSmartWallet.sol` L23–97 / L102–224 | The **two hops** the import bridge chains: unshield-out of Railgun's pool, then shield-in to ours. Both are pre-existing public entrypoints on each respective pool. |
| **POI** — Private Proofs of Innocence | [A.1.7](./A.1-reuse-inventory.md); docs.railgun.org/…/private-proofs-of-innocence | An **alongside partner system, not in the pool contracts** ([A.1.1](./A.1-reuse-inventory.md) L26). On shield, an **Unshield-Only Standby Period** (default 1h) precedes a blinded non-inclusion ZK proof vs list-provider datasets; proofs carry forward through transfers. |
| Cross-pool proofs — *not present* | [A.1.9](./A.1-reuse-inventory.md) delta; [ADR-0006](../09-architecture-decisions.md#adr-0006) | No shipped mechanism spends/nullifies a foreign-pool note in ours. Would need a circuit that trusts Railgun's root ⇒ a **T0.1 re-run** (A.3/A.9). Deferred/optional. |

**Pins.** Railgun repos (pool, `circuits-v2`, engine) are **UNPINNED** — read at `master`/`main` HEAD 2026-09-05; **pin a commit before build** ([A.1](./A.1-reuse-inventory.md), A.0.6). The bridge and POI integration reference Railgun's *own live* pool + list-provider stack, which are external and independently versioned.

## A.7.4 Net-new delta (what A actually builds)

T0.7 is mostly policy, orchestration, and one external-dependency integration. Four deltas:

1. **Bootstrap incentives and LP flywheel.** Per [ADR-0006](../09-architecture-decisions.md#adr-0006), seed our own crowd via incentives, our own market-making capital, and an LP flywheel, so that shield and transact volume compounds the effective anonymity set. k-anonymity compounds with volume ([ADR-0006](../09-architecture-decisions.md#adr-0006) consequences). Early transactions have a small, weak set; the strategy's job is to grow it fast enough that effective-k crosses a usable threshold before real value depends on it. This is program and economic policy, not contract code, and A.7 owns the *specification*, not the treasury.
2. **Railgun onboarding import-bridge script.** A client-side flow that chains unshield-from-Railgun to shield-into-ours in one public boundary hop. It composes two pre-existing entrypoints (A.7.3) and mints fresh commitments in our accumulator; it deploys no new contract and touches neither pool's internals. This gives day-one dual liquidity ([ADR-0006](../09-architecture-decisions.md#adr-0006)).
3. **POI-node integration.** Stand up or connect the separate POI node stack, a dependency not redeployable from the four repos scoped here ([A.1.7](./A.1-reuse-inventory.md), [A.1.10](./A.1-reuse-inventory.md), [A.0.5](./A.0-overview.md) gate 2). This includes choosing list providers and the **Unshield-Only Standby Period** at the client and POI node (the T0.0 "POI config", a client or node setting, not a pool setting), and running the list-provider dataset feed that the standby proof checks against.
4. **Gated-entry policy in the settlement path.** T0.3 enforces that value entering a channel carries a valid POI (A.7.5 ICD). This is settlement-side policy, not an on-chain pool flag. The pool has no POI hook ([A.1.1](./A.1-reuse-inventory.md) L26), so the check lives in the deposit path A.5 implements against the allow-list root A.7 supplies.

**Liquidity import is not anonymity-set inheritance.** A Railgun note cannot be spent or nullified in our contract. Our JoinSplit circuit proves membership in *our* accumulator against *our* nullifier set, not Railgun's. So "accept Railgun's liquidity" means the unshield-then-shield hop (delta 2), which lands *fresh* commitments in *our* set. It does not import their crowd: the imported value contributes to our k only from the moment it is shielded into us, exactly like any other deposit. Inheriting their set would require cross-pool membership proofs. Those are deferred and optional ([ADR-0006](../09-architecture-decisions.md#adr-0006)), because they would need a T0.1 re-run and would trust Railgun's root (A.7.7; see A.3 and A.9).

## A.7.5 Interfaces (ICD)

**A.7 to A.5 (consumed by the deposit/payout contract).**
- **POI allow-list root and policy.** A.7 owns the choice of list providers and standby period, and publishes the resulting allow-list root and the policy predicate it stands for. A.5's deposit path consults this root to enforce gated-entry: a `transact()` unshield-in whose value lacks a valid POI is rejected before `MultiAssetHolder.deposit`. The root and policy are the ICD; the check is A.5's ([A.1.4](./A.1-reuse-inventory.md) deposit-in).
- This is the same **POI allow-list root** A exposes at the work-package boundary ([`00-work-packages.md`](../00-work-packages.md) "A exposes"). Consumers treat it as opaque; only its freshness and validity semantics are contractual.

**A.7 with the POI node stack (external dependency).**
- **In:** shield events and the commitment stream from the pool ([A.1.1](./A.1-reuse-inventory.md) events), plus list-provider datasets. **Out:** blinded non-inclusion proofs and the allow-list root A.7 republishes. The standby period (default 1h) is a latency term that downstream flows (bridge, deposit) must budget for.

**Import bridge (client flow, no on-chain ICD).**
- Two calls, two pools: `Railgun.transact()` unshield-out, then `Ours.shield([ShieldRequest…])` in. There is no contract-to-contract interface; the only shared artifact is the ERC20 that transits the public boundary hop between them.

**Non-interfaces (explicit).** T0.7 exposes no pool-side POI setting (there is none, [A.1.1](./A.1-reuse-inventory.md) L26) and no cross-pool proof surface in v1.

## A.7.6 Acceptance / verification

1. **Shielded-payments primitive (inherited-capability check).** On the fixturenet, a `transact()` with `unshield=NONE` moves value between two of our notes; assert the public transcript exposes only `merkleRoot`/`boundParamsHash`/`nullifiers[]`/output-commitment hashes, and no sender, recipient, or amount ([A.1.7](./A.1-reuse-inventory.md)). This confirms A delivers shielded payments without any A-net-new code.
2. **Import-path smoke.** End-to-end on a fixturenet with a stand-in "foreign" pool: run the bridge script `unshield-from-Railgun → shield-into-ours`; assert (a) exactly one public boundary hop (one unshield, one shield, value conserved across it), (b) a fresh commitment appears in *our* accumulator, and (c) the foreign note's nullifier is spent in the foreign pool only, never referenced by ours (proving import is not inheritance, A.7.4).
3. **Gated-entry enforcement.** A deposit-in carrying a valid POI proceeds to `MultiAssetHolder.deposit`; one lacking a valid POI, or checked against a stale allow-list root, is rejected in the settlement path before escrow (with A.5). This confirms the A.7-to-A.5 ICD gate is enforced.
4. **POI standby honored.** A shield followed by a bridge or deposit before the **Unshield-Only Standby Period** elapses is treated per policy (no valid POI yet); after the period and non-inclusion proof, it passes. This confirms the POI-node integration and its latency term.
5. **Privacy metrics (the strategy's real KPI).** Instrument and report:
   - **Anonymity-set size** — cardinality of live commitments in our accumulator ([A.1.1](./A.1-reuse-inventory.md) `Commitments.sol`).
   - **Effective-k** — the realistic same-denomination/recent-activity crowd a given transaction blends into (not raw set size; volume-weighted per [ADR-0006](../09-architecture-decisions.md#adr-0006) "compounds with volume").
   - **Boundary-exposure** — count and volume of public boundary hops (deposit/payout per [ADR-0005](../09-architecture-decisions.md#adr-0005), plus each import hop), since these are where amounts are transparent and are what the metric bootstrap trades against.

   Acceptance means these three are computed and trend-visible on the fixturenet; the bootstrap target is effective-k rising with volume, not a fixed one-shot number.

## A.7.7 Risks / open

- **Cold-start weakness window.** Until the bootstrap flywheel lifts effective-k, early users have a small crowd, and large or oddly-denominated transactions are self-deanonymizing. The mitigation is policy: incentivize volume and encourage denomination bucketing. There is no engineering fix in v1.
- **Cross-pool proofs deferred (optional/future).** Inheriting Railgun's set is the strong-privacy prize, but it needs a T0.1 re-run and trusts Railgun's root ([ADR-0006](../09-architecture-decisions.md#adr-0006)); it is scoped to A.3/A.9, not v1. If greenlit, it changes circuits, which means a Phase-2 re-run and a fresh audit (compare the [ADR-0003](../09-architecture-decisions.md#adr-0003) and [ADR-0005](../09-architecture-decisions.md#adr-0005) cost shape).
- **POI is an external stack.** The POI node and list-provider datasets are a separate dependency ([A.1.10](./A.1-reuse-inventory.md)); the list-provider choice is a trust and liveness surface, and their downtime stalls gated-entry. Own the failure mode (degrade-closed versus open) as a policy decision with A.5.
- **Boundary-exposure versus liquidity tension.** Each import hop and each deposit/payout is a transparent-amount event ([ADR-0005](../09-architecture-decisions.md#adr-0005)); more liquidity import can *raise* boundary-exposure even as it raises set size. The metric suite (A.7.6.5) exists to keep this trade visible.

See also: feeds [A.5](./A.5-deposit-payout-contract.md) (gated-entry, POI root ICD); relates to [A.8](./A.8-interfaces-acceptance.md) (POI allow-list root at the WP boundary) and to [A.3](./A.3-trusted-setup.md)/[A.9](./A.9-native-commitment.md) (cross-pool proofs if greenlit). Grounded in [A.1.7](./A.1-reuse-inventory.md) and [ADR-0006](../09-architecture-decisions.md#adr-0006).
