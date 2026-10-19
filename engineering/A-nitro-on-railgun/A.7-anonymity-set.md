# A.7 — Anonymity-Set Strategy

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T0.7 (anonymity-set strategy — bootstrap the crowd with new users)

---

## A.7.1 Goal

Make the shielded value that A settles actually private in practice, not merely private in principle. The base construction (A.0.1) hides sender, recipient, and amount in-circuit; what it cannot manufacture is the **crowd** those hidden transactions blend into. A fresh pool starts with an anonymity set near zero, which [ADR-0006](../09-architecture-decisions.md#adr-0006) records as *the biggest non-engineering risk* to A. T0.7 is therefore a **strategy item** (per [§5](../05-building-block-view.md), status "strategy", not net-new crypto). It specifies the base shielded-payments primitive A delivers, the cold-start bootstrap plan, and the privacy metrics that show whether the crowd grows.

The crowd grows only from deposits into our own pool. Every note a spend can hide among is a note in our tree, so the only lever is to bring in users and volume.

## A.7.2 Boundary

- **In scope.** The shielded-payments primitive as a *capability statement* (it is inherited, not built, per A.7.3); bootstrap incentive and flywheel policy; and privacy metrics.
- **Out of scope.** The pool contract itself and the `changeFee`/vkey levers; see [A.2](./A.2-pool-deployment.md). The wallet move from Railgun, which is a wallet feature (T6.7, [`T6-client-apps.md`](../T6-client-apps.md)). Native amount-privacy-in-play; see [A.9](./A.9-native-commitment.md). Selective disclosure, which is off-chain and gates nothing ([ADR-0020](../09-architecture-decisions.md#adr-0020), [16](../16-selective-disclosure.md)).
- **Phase-0 assumption.** T0.7's bootstrap presumes a live pool (T0.0, [A.2](./A.2-pool-deployment.md)) and ceremony (T0.1, [A.3](./A.3-trusted-setup.md)). The residual dependency is that those builds ship before the crowd carries real value. Until they do, A.7 is a design and fixturenet exercise only.

## A.7.3 Reuse inventory (cite A.1 + pinned commits)

Everything T0.7 leans on already exists; T0.7 adds no cryptography. Cite, don't re-document.

| Piece | Source (see A.1) | What it gives T0.7 |
|---|---|---|
| **Shielded-payments primitive** = `transact()` with `unshield=NONE` | [A.1.7](./A.1-reuse-inventory.md) (primitive); `transact` in [A.1.1](./A.1-reuse-inventory.md), `RailgunSmartWallet.sol` L102–224 | The base capability A delivers **for free**: inputs nullified, outputs created, `sumIn==sumOut` in-circuit; **sender/recipient/amount hidden** — only `merkleRoot`, `boundParamsHash`, `nullifiers[]`, output-commitment hashes are public. This is a native transfer inside our pool, with no deposit/payout contract and no channel. |
| JoinSplit circuit (defines the primitive) | [A.1.2](./A.1-reuse-inventory.md), `src/library/joinsplit.circom` L11–118 | Fixes note/commitment/nullifier format; proves EdDSA ownership + Merkle membership + range + value conservation. The privacy set = the commitments in the accumulator this circuit proves membership of. |
| Merkle accumulator + nullifier set (`TREE_DEPTH=16`, `rootHistory`) | [A.1.1](./A.1-reuse-inventory.md), `Commitments.sol` L28–55, L108–252 | **The anonymity set, concretely**: every shielded commitment ever inserted. Its cardinality (and recent activity) is what T0.7 metrics measure. |

**Pins.** Railgun repos (pool, `circuits-v2`, engine) are **UNPINNED** — read at `master`/`main` HEAD 2026-09-05; **pin a commit before build** ([A.1](./A.1-reuse-inventory.md), A.0.6).

## A.7.4 Net-new delta (what A actually builds)

T0.7 is policy, not contract code. One delta:

1. **Bootstrap incentives and LP flywheel.** Per [ADR-0006](../09-architecture-decisions.md#adr-0006), seed our own crowd with new users, through incentives, our own market-making capital, and an LP flywheel, so that shield and transact volume compounds the effective anonymity set. Early transactions have a small, weak set; the strategy's job is to grow it fast enough that effective-k crosses a usable threshold before real value depends on it. This is program and economic policy, not contract code, and A.7 owns the *specification*, not the treasury.

Value that a user moves in from Railgun through the wallet (T6.7) is an ordinary deposit. It counts towards the crowd only from the moment it is shielded here, like any other deposit.

## A.7.5 Interfaces (ICD)

T0.7 exposes no on-chain interface. It consumes the pool's commitment and nullifier events ([A.1.1](./A.1-reuse-inventory.md) events, indexed by T1.2) to compute its metrics.

## A.7.6 Acceptance / verification

1. **Shielded-payments primitive (inherited-capability check).** On the fixturenet, a `transact()` with `unshield=NONE` moves value between two of our notes; assert the public transcript exposes only `merkleRoot`/`boundParamsHash`/`nullifiers[]`/output-commitment hashes, and no sender, recipient, or amount ([A.1.7](./A.1-reuse-inventory.md)). This confirms A delivers shielded payments without any A-net-new code.
2. **Privacy metrics (the strategy's real KPI).** Instrument and report:
   - **Anonymity-set size** — cardinality of live commitments in our accumulator ([A.1.1](./A.1-reuse-inventory.md) `Commitments.sol`).
   - **Effective-k** — the realistic same-denomination/recent-activity crowd a given transaction blends into (not raw set size; volume-weighted per [ADR-0006](../09-architecture-decisions.md#adr-0006) "grows only with our own volume").
   - **Boundary-exposure** — count and volume of public boundary steps (shield, unshield, and deposit/payout per [ADR-0005](../09-architecture-decisions.md#adr-0005)), since these are where amounts are transparent and are what the bootstrap trades against.

   Acceptance means these three are computed and trend-visible on the fixturenet; the bootstrap target is effective-k rising with volume, not a fixed one-shot number.

## A.7.7 Risks / open

- **Cold-start weakness window.** Until the bootstrap flywheel lifts effective-k, early users have a small crowd, and large or oddly-denominated transactions are self-deanonymizing. The mitigation is policy: incentivize volume and encourage common amounts. There is no engineering fix in v1.
- **Boundary-exposure versus liquidity tension.** Each shield, unshield, and deposit/payout is a transparent-amount event ([ADR-0005](../09-architecture-decisions.md#adr-0005)); more deposits raise boundary-exposure even as they raise set size. The metric suite (A.7.6.2) exists to keep this trade visible.

See also: relates to [A.8](./A.8-interfaces-acceptance.md) (roll-up acceptance) and to T6.7 (the wallet move from Railgun). Grounded in [A.1.7](./A.1-reuse-inventory.md) and [ADR-0006](../09-architecture-decisions.md#adr-0006).
