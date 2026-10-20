# 1. Introduction & Goals

arc42 §1 · **2026-10-20**

## 1.1 Purpose

These documents specify Armada, a Railgun-based privacy product for USDC, as built by one team working from the ship-armada repos: `armada-poc`, `armada-circuits`, `armada-deployments`, `armada-nitro`, `armada-sdk`, `armada-client`, `armada-relayer`, `armada-interface` ([ADR-0013](./09-architecture-decisions.md#adr-0013)). Armada concentrates USDC into a single, immutable shielded pool (one anonymity set), with adapters (CCTP, Aave-v4 yield) and a web app on top. The shielded pool and its circuits are Armada's own implementation of the Railgun design, spec-compatible, built and tested (T0.0/T0.1, [ADR-0014](./09-architecture-decisions.md#adr-0014)).

The core engineering construction is the deposit/payout boundary inside `ArmadaNitroAdjudicator`: fund a `go-nitro` channel from a shielded note and settle its outcome back into fresh notes. This *notes in → normal Nitro → notes out* flow, built, lets optimistic exits and arbitrary off-chain settlement operate over shielded value with no new cryptography ([ADR-0004](./09-architecture-decisions.md#adr-0004)). The privacy analysis (the pool) and the game analysis (Nitro) meet only at that boundary (→ §8).

## 1.2 What is built, and what remains

- **Shielded payments.** Built. A user shields USDC into the pool and sends and receives it as shielded notes without revealing sender, recipient, or amount, through `armada-interface`'s web UI (**T0.0**, **T6.1**, **T6.2**).
- **Yield.** Built. CCTP and Aave-v4 yield adapters route through the pool's own `adaptContract` binding (**T5.0**), not Railgun's RelayAdapt/Cookbook recipe, which this codebase does not use.
- **Private exchange and clearing.** Not built. No venue, no posted price, and no matcher exist in `armada-poc` (**T3**, **T4**).
- **Verified, mobile-first client support.** In progress. The proof path is the nimbus verified proxy (**T1**, [ADR-0009](./09-architecture-decisions.md#adr-0009)), proven by hand on the armada-nitro fixturenet but wired into no client. `armada-interface` is web-only; `armada-client`'s four-target plan (Linux server, browser, iOS, Android) has native builds for Linux, iOS, and Android, and a WASM build that stalls in a browser.
- **Settlement-client channel lifecycle and a client watchtower.** Not built. The adjudicator side is built and tested; no client drives it yet (**T6.2**, **T6.3**).

## 1.3 Top quality goals

| # | Goal | Statement |
|---|---|---|
| 1 | **Non-custody** | No component ever holds user funds unilaterally; the pool and the adjudicator enforce this on-chain, and this holds in the code as built (→ §2, §8). |
| 2 | **Transfer and read-time privacy** | Shielded transfer is built and in-circuit. Read-time privacy depends on the proof path (**T1**), which is not yet integrated into any client; today a client reads chain state over RPC or an unverified cache. |
| 3 | **Mobile-first operability** | Not yet true. The live client is web-only; mobile has no code. |
| 4 | **Private non-custodial clearing** | Not yet true. No venue exists to clear a trade over. |
| 5 | **Force-closable liveness** | The adjudicator supports it and `armada-nitro`'s tests exercise it; no client-side watchtower exists to automate it (**T0.2**, **T6.3**). |

## 1.4 Stakeholders

| Stakeholder | Concern |
|---|---|
| **Armada dev team** | A proof path that lets any client verify state without trusting an RPC or a cache; a settlement client and watchtower for the adjudicator that already exists. |
| **End users (web today, mobile planned)** | Private payments and non-custody, already available in `armada-interface`; a usable mobile app, not yet available. |
| **Governance participants** | The built governance and revenue contracts (`ArmadaGovernor`, `ArmadaTreasuryGov`, `ArmadaToken`, `ArmadaCrowdfund`, `ArmadaWindDown`, `ArmadaRedemption`) and their operation. |
| **Watcher operators (future)** | A viable snapshot-attestation role (T2.4), still a design, not code. |

## 1.5 Requirements traceability, by tier

| Tier | Coverage |
|---|---|
| **T0** Ethereum anchor | Pool, circuits, adjudicator, governance built; T0.4–T0.6 not built. |
| **T1/T2** Proof path and cache | Proven by hand on the fixturenet; unintegrated. |
| **T3** Ordering | Entirely not built; a plan. |
| **T4** Execution / venue | Entirely not built; a plan, except T4.4's yield case, which needs no new contract. |
| **T5** Adapters | CCTP and Aave-v4 yield built; Swaps not built. |
| **T6** Client / apps | The web app is built; mobile, the settlement-client lifecycle, and the watchtower are not. |

→ Full scope boundary in §3; solution strategy in §4; risks in §11.
