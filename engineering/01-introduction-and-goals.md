# 1. Introduction & Goals

arc42 §1 · **2026-09-05**

## 1.1 Purpose

These documents specify the engineering work to build **Armada — a Railgun-based privacy product for USDC** — as **one team**, reusing **Laconic's prior art** (nitro, watchers, wallet, chain-signatures, ingestion) as the substrate ([ADR-0013](./09-architecture-decisions.md#adr-0013)). Armada concentrates USDC into a single immutable Railgun shielded pool (one anonymity set), with adapters (CCTP, Aave-v4 yield, Swaps) and an apps/SDK tier on top; the load-bearing fact is that the hard problem is **integration, not new cryptography** — with one caveat: because Railgun's contracts/circuits are unlicensed, the shielded pool and its circuits are **clean-room reimplemented** (spec-compatible with the Railgun design; T0.0/T0.1, [ADR-0014](./09-architecture-decisions.md#adr-0014)), so that piece is net-new, audit-critical crypto-engineering of a known design, while the **settlement rail stays pure integration**. Scope spans the whole product — there is no Armada-vs-Laconic build split. This plan traces to Armada's requirements; the mapping is the traceability matrix (§1.5).

The core engineering thesis is **nitro-on-railgun**: fund a `go-nitro` state-channel game from a shielded note and settle its outcome back into fresh notes. This *notes in → normal Nitro → notes out* flow lets optimistic exits and arbitrary off-chain settlement operate over shielded value **with no new cryptography** and without governance-gated adapters ([ADR-0004](./09-architecture-decisions.md#adr-0004); realized in **T0.3**). The privacy analysis (Railgun) and the game analysis (Nitro) meet only at the allocation→note **boundary** (→ §8). Every capability below is a composition of that one construction with the shielded pool, the watcher substrate, and the wallet.

## 1.2 What v1 delivers

v1 provides four user-facing capabilities, in dependency order — each builds on the one before it, and all four are non-custodial and mobile-first.

- **Shielded payments** — private, non-custodial transfer of value between users. A user shields USDC into the pool and can send and receive it as shielded notes without revealing sender, recipient, or (away from the transparent boundary) amount. This is the foundational capability for privacy-infrastructure-for-USDC and the base case every other capability composes with: it is Railgun-style shielded transfer over our **clean-room pool** (**T0.0**), made usable by on-device note scanning and signing (**T6.1**, **T6.2**) and private note discovery (**T2.0**).
- **Shielded yield** — value held in the pool earns yield privately, and **USDC yield is the priority** (Armada is privacy-for-USDC). USDC yield runs through the **Aave-v4 adapter**, in scope under work package C ([ADR-0013](./09-architecture-decisions.md#adr-0013)); **ETH yield** is a shielded **wstETH note**, intrinsic and non-rebasing, needing no Aave or LP (**T4.4**). Whether v1 USDC yield uses the direct adapter recipe (public Design-A boundary) or the private **LP-buffered rail** (**T4.5**, which needs the v2 market-making, ADR-0011) is an open C-scope decision (§0).
- **Private exchange & clearing** — users swap one shielded asset for another (for example, ETH-yield → USDC) at a **posted price**, cleared non-custodially over Nitro, with nothing to front-run in v1 (**T4.0**, **T4.1**, **T4.3**, settling via **T0.3**/**T0.2**; [ADR-0007](./09-architecture-decisions.md#adr-0007)). v1 keeps this deliberately simple: the venue fills from a **pre-funded static inventory** with **no automated market-making**. The aggregating solver (**T4.2**) and price discovery via the ex_net matcher (**T4.6** + **T3.\***) are v2 ([ADR-0011](./09-architecture-decisions.md#adr-0011)).
- **Mobile-first private support services** — the substrate that runs all of the above on phones and browsers with almost no fixed infrastructure: **private sync** (proof-carrying note streams that protect the anonymity set *at read-time*, closing the fragility where the set would otherwise die over normal RPC, **T2.0**), **metering** (pay-per-request Nitro vouchers, **T2.1**), the **client and relay** (keys, Cosmos/EIP-155 signing, an in-browser Nitro node, and libp2p transport, **T6.0**/**T2.2**/**T2.3**), and the **wallet/app** built up from the Laconic wallet and demo fragments (**T6.0**, **T6.7**).

The venue (work package **C**) clears exchange and yield and hosts the **adapters** — Swaps, Aave-v4 yield, CCTP — as applications on the settlement rail; the shielded pool stays immutable and untouched, and every capability is non-custodial.

## 1.3 Top quality goals

The goals are ranked, one sentence each; concrete scenarios and metrics follow in → §10.

| # | Goal | Statement |
|---|---|---|
| 1 | **Non-custody** | No component ever holds user funds unilaterally — value moves only through the shielded pool and channels the user co-signs (→ §2, §8). |
| 2 | **Transfer & read-time privacy** | Payments and note discovery run over the shielded pool and proof-carrying private-sync feeds — never normal RPC — so neither a transfer nor the act of scanning for it deanonymizes the anonymity set (**T2.0**, **T0.7**; [ADR-0006](./09-architecture-decisions.md#adr-0006)). |
| 3 | **Mobile-first operability** | Wallet, note-scanning, proving, transport, and watchtower all run on a phone with almost no fixed infrastructure (**T6.\***; [ADR-0008](./09-architecture-decisions.md#adr-0008)). |
| 4 | **Private non-custodial clearing** | The venue clears trades and yield over Nitro without taking custody or leaking order flow — a posted price has nothing to front-run in v1 ([ADR-0007](./09-architecture-decisions.md#adr-0007)). |
| 5 | **Force-closable liveness** | Any party can unilaterally force-close a channel and exit on-chain via the adjudicator, so the system never traps funds on operator failure (**T0.2**, **T6.3**; [ADR-0004](./09-architecture-decisions.md#adr-0004)). |

## 1.4 Stakeholders

| Stakeholder | Concern |
|---|---|
| **Armada dev team** | A substrate that meets v1 requirements and plugs into the existing pool/Adapter/SDK model without modifying the immutable pool; a clear build-scope split (→ §3). |
| **Laconic / Vulcanize** | Reuse of the cerc-io watcher/nitro/wallet stack; delivery of the net-new contracts and the mobile crux on schedule. |
| **End users (mobile)** | Private payments, non-custody, read-time privacy, and a usable phone app that syncs, proves, and force-closes on its own. |
| **Venue operators / LPs** | Posted-price spread + clearing economics; a static pre-funded inventory in v1, with internalization / hedge / LP-yield economics arriving in v2 (**T4.2**, **T4.5**). |
| **Watcher-party operators** | Metered revenue for serving proof-carrying feeds; a viable federation/bond path (**T2.1**, **T2.4**, **T0.4**). |

## 1.5 Requirements traceability matrix

Each Armada v1 requirement is mapped to the satisfying `T#.#` registry item(s) (§5) with a coverage note. Status facets (`net-new` / `net-new (clean-room)` · `reuse` · `partial`) are those of the §5 registry.

### Capability requirements

| v1 requirement | Satisfied by | Coverage note |
|---|---|---|
| **Shielded payments** (private non-custodial USDC transfer) | **T0.0** shielded transfer · **T6.1**/**T6.2** scan + client · **T2.0** private discovery | Transfer itself is `net-new (clean-room)` (our shielded pool, spec-compatible with Railgun; [ADR-0014](./09-architecture-decisions.md#adr-0014)); the rest of the work is the on-device scanner/bridge (**T6.1**, `net-new`) and private sync. |
| **Shielded yield — ETH** (wstETH note) | **T4.4** | `config` only — intrinsic, non-rebasing; strongest yield path. |
| **Shielded yield — USDC** (priority; via Aave-v4 adapter) | **T5.0** adapter → **T4.5** *(private rail, v2)* | **In scope** ([ADR-0013](./09-architecture-decisions.md#adr-0013)); v1 mechanism open — direct adapter recipe (public boundary) vs LP-buffered private rail (T4.5, v2 market-making, ADR-0011). |
| **Private exchange & clearing** (posted price) | **T4.0** price · **T4.1** quote/settle · **T4.3** fee-split · settles via **T0.3**/**T0.2** | `net-new` v1 spine; the venue fills from a **static inventory** — no market-making, solver **T4.2** is v2 ([ADR-0011](./09-architecture-decisions.md#adr-0011)). A posted price has nothing to front-run ([ADR-0007](./09-architecture-decisions.md#adr-0007)). |
| **Spread + clearing fees** | **T4.3** | Fee-split (user/protocol/integrator) encoded on channel state. |
| **Private sync** (read-time anonymity) | **T2.0** serving · **T0.7** set strategy · fed by **T1.\*** | Serving is `reuse+config` (watcher-ts); **depends on the T1 ingestion long pole — see gap below**. |
| **Metering** (Nitro vouchers) | **T2.1** | `reuse+config` of `payments.ts`, funded at note creation — strong. |
| **Client + relay** (keys, signing, in-browser Nitro node, libp2p) | **T6.0** custody/signing · **T2.2** peer substrate · **T2.3** transport | `reuse` across the board ([ADR-0008](./09-architecture-decisions.md#adr-0008)); Noise gives peer-auth, **not IP privacy** (§11). |
| **Front-end** (wallet → Armada app) | **T6.0** wallet base · **T6.7** app + SDK | `partial/reuse` fork + `product` wiring; build-up, not new invention. |
| **Deploy / ops** (bonded federations) | **T2.4** federation+bond · **T0.4** registry/bond | **`partial` and `v2`** — v1 runs an unbonded/trusted federation; bonding + threshold DKG land in v2. |
| **Price discovery** (ex_net matcher) | **T4.6** + **T3.\*** | **`v2` — out of v1 scope**; layered on the same spine, not a rewrite. |

### Tier-map coverage

| Tier | v1 items | Coverage note |
|---|---|---|
| **T0** Ethereum anchor | T0.0, T0.1, T0.2, T0.3, T0.7 | Pool + circuits `net-new (clean-room)` (spec-compatible, [ADR-0014](./09-architecture-decisions.md#adr-0014)); adjudicator `built/reuse`; boundary + Phase-2 setup `net-new` ([ADR-0002](./09-architecture-decisions.md#adr-0002)/[0003](./09-architecture-decisions.md#adr-0003)/[0004](./09-architecture-decisions.md#adr-0004)). T0.4–T0.6 are v2/opt. |
| **T1** Ingestion | T1.0, T1.1, T1.2 | **Weakest v1 coverage — the long pole; see flag below.** |
| **T2** Watcher substrate | T2.0, T2.1, T2.2, T2.3 | Mostly `reuse`/`reuse+config`; T2.4 (bond) is v2. |
| **T3** Ordering | — | Entirely **v2**; not on the v1 spine ([ADR-0007](./09-architecture-decisions.md#adr-0007)). |
| **T4** Execution / venue | T4.0, T4.1, T4.3, T4.4 | Small posted-price contract + static-inventory fill + Design A boundary; **solver T4.2, USDC rail T4.5, matcher T4.6 are v2** ([ADR-0011](./09-architecture-decisions.md#adr-0011)). |
| **T5** Adapters | T5.0, T5.1 | **In scope — work package C** ([ADR-0013](./09-architecture-decisions.md#adr-0013)); CCTP built, Aave-v4 yield + Swaps build-or-reuse Railgun Cookbook (license gate). |
| **T6** Client / apps | T6.0–T6.7 | `partial`; T6.5 native mobile transport is the "crux" (§11). T6.8 is opt. |

### Gaps explicitly flagged

- **Weakest v1 requirement — T1 ingestion (nimbus-eth1).** T1.0 and T1.1 are `partial` because the nimbus-eth1 stateless/witness path is still upstream-in-progress, which makes T1 the **v1 long pole** ([ADR-0009](./09-architecture-decisions.md#adr-0009)). Both private sync and watcher freshness for the watchtower (**T6.3**) depend on it, so it is the coverage risk to watch (→ §11).
- **Adapters are in scope (not external).** Per [ADR-0013](./09-architecture-decisions.md#adr-0013) this is one team building the whole Armada product; the adapters (**T5.0** — CCTP built, Aave-v4 yield, Swaps) are owned by work package **C**, built or reused from Railgun's Cookbook recipes (license gate). The Aave-v4 adapter is the **USDC-yield** mechanism — the priority capability.

→ Full build-scope boundary in §3; delivery increments in §4; risks in §11.
