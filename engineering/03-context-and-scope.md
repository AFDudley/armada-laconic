# 3. Context & Scope

arc42 §3 · **2026-10-20**

This section draws the system boundary and catalogues the external systems the build touches. Per [ADR-0013](./09-architecture-decisions.md#adr-0013), this is **one engineering team building the whole Armada product** — a Railgun-based privacy layer for USDC. The team works from the ship-armada repos: `armada-poc`, `armada-circuits`, `armada-deployments`, `armada-nitro`, `armada-sdk`, `armada-client`, `armada-relayer`, `armada-interface`. The static decomposition of the in-scope side is the [§5 registry](./05-building-block-view.md); here we fix only the perimeter. Terminology follows [ADR-0010](./09-architecture-decisions.md#adr-0010).

## The product

Armada is pluggable asset-privacy infrastructure for USDC: routine flows (payments, payroll, treasury, yield) route through a single immutable shielded pool (one anonymity set), and privacy compounds as more flows join. The pool, its governance contracts, the circuits, the adjudicator, and the web client are built. The proof path, the mobile client, and any future venue remain a plan.

## In scope — the whole product (by work package)

The four [work packages (§0)](./00-work-packages.md) are the entire scope; each owns a set of `T#.#` items, per the real status in [§5](./05-building-block-view.md).

| WP | §5 items | Note |
|---|---|---|
| **A · on-chain anchor** | T0.0–T0.7; governance and revenue contracts | The pool, circuits, adjudicator and governance are built. T0.4–T0.6 remain a plan. |
| **B · proof path** | T1.0–T1.2, T2.0–T2.4; T6.1 | The nimbus verified proxy, not a state-diff emitter (ADR-0009); `watcher-ts` is an optional server cache. |
| **C · venue and adapters** | T3.\*, T4.\* (plan); T5.0, T5.1 | CCTP and Aave-v4 yield are built; Swaps and the venue have no code. |
| **D · client** | T6.0, T6.2–T6.8 | `armada-interface` (web) is built; the mobile stack is a spec (`armada-client` `docs/SPEC.md`). |

## Built, not reused (own implementation)

Cited at pinned commits.

| Component | Note |
|---|---|
| Shielded pool (`armada-poc`) and `Transact` circuit (`armada-circuits`) | Own implementation, spec-compatible with the Railgun design (ADR-0014). Railgun's `contract`/`circuits-v2`/`engine`/`cookbook` served only as a behavioral reference during the build. |
| Governance and revenue contracts (`armada-poc`) | `ArmadaGovernor`, `ArmadaTreasuryGov`, `TreasurySteward`, `ArmadaToken`, `ArmadaCrowdfund`, `ArmadaWindDown`, `ArmadaRedemption`, `AdapterRegistry`, `RevenueCounter`, `RevenueLock`, `RevenueReserveDistributor`. |
| `ArmadaNitroAdjudicator`, `HubCustodian` (`armada-nitro`) | A patched `go-nitro` (ADR-0004) plus a FROST-gated hub custodian (ADR-0016), built, not the vendored original and not `chain-signatures`. |
| `armada-interface`, `armada-sdk` | The live web client and its SDK: shield/transact calls, note scanning, and an in-browser Groth16 prover. |

## Reused, not built (prior art & reference)

| Component | Note |
|---|---|
| `go-nitro` | Vendored, then patched by six commits (ADR-0004). The only Laconic prior art still in the build. |
| Railgun design (`contract` / `circuits-v2` spec; `engine` / `cookbook` as reference) | Reference spec only — the pool and circuits are Armada's own implementation (ADR-0014). |
| Perpetual Powers of Tau Phase-1 | Inherited universal ceremony; Armada's own Phase-2 has not run yet (ADR-0003). |
| Circle CCTP | The cross-chain USDC rail the pool's CCTP path wraps. |
| Aave v4 (mock in the current build) | The yield protocol `ArmadaYieldAdapter` reaches, through the pool's own `adaptContract`/`adaptParams` binding, not Railgun's RelayAdapt/Cookbook. |

## External systems & interfaces (genuine third parties)

Direction is stated relative to Armada's system.

| External system | Interface & direction | Touched by | Decision |
|---|---|---|---|
| **Ethereum L1** | Settlement root: the pool, the patched adjudicator, and governance contracts live here. Bidirectional — write channel and pool state, read commitment/nullifier/adjudicator events. | T0.0, T0.2 (read at T1) | ADR-0002, ADR-0004, ADR-0014 |
| **Railgun deployed pool** (live) | Wallet move (not yet built): a user would unshield from Railgun, then shield into Armada's pool — one public step that shows token, amount and time. No contract of Armada's reads or touches Railgun's pool. Inbound. | T6.7 | ADR-0006 |
| **Aave v4** | The yield protocol the Aave-v4 adapter reaches (`unshield USDC → supply → reshield aUSDC`), through the pool's own adapt-contract binding. Amounts public only at the transparent supply boundary. | T5.0 (C) | — |
| **Circle CCTP** | Cross-chain USDC mint wrapped into a shield by `CCTPHookRouter` — arrives as a fresh shielded USDC note. Bidirectional; built. | T5.0 (C), via T0.0 | — |
| **Beacon chain light client** | The proof anchor the nimbus verified proxy checks every header against. Inbound. | T1 | ADR-0009 |
| **The armada-nitro fixturenet** | Real `geth` + `lighthouse` consensus, used for testing the adjudicator and, once patched, the first nimbus-proxy proof. | T0.2, T1 | ADR-0009, `labbook/2026-10-19-nimbus-verified-proxy-proofs.md` |

## Not built / excluded

- Amount privacy in-play beyond the shield/unshield boundary (T0.6, optional, not built).
- A venue, a matcher, or market-making (T3, T4): no code exists in `armada-poc`.
- A Railgun onboarding bridge or any cross-pool membership proof: a spend hides only among the notes it could consume, so no code builds one (`labbook/2026-10-19-docs-versus-code.md` decision 2).
- No Proofs of Innocence as an entry gate; selective disclosure (ADR-0020) is voluntary and off-chain.

---

→ Building-block detail and every `T#.#` id: [§5](./05-building-block-view.md). Scope decision: [ADR-0013](./09-architecture-decisions.md#adr-0013). Adapter detail: [`T5-adapters.md`](./T5-adapters.md).
