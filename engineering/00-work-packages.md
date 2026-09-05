# 0. Work Packages (WBS)

delivery frame · **2026-09-05** · operating model: [ADR-0012](./09-architecture-decisions.md#adr-0012)

This is the **work breakdown**: the deliverable-oriented decomposition of the programme, sitting above the arc42 architecture description. It is MECE and obeys the 100% rule — the four work packages together are the whole programme's scope (the entire Armada product; [ADR-0013](./09-architecture-decisions.md#adr-0013)), and none overlaps another. Each work package is a **system with a boundary**: it *owns* a set of `T#.#` items from the [§5 registry](./05-building-block-view.md) (the shared building-block vocabulary — the tiers are **not** re-partitioned) and *consumes* everything else across a named interface (ICD). A work package may only touch its own boundary and its own deltas; everything else is a cited [baseline decision](./09-architecture-decisions.md) or an external interface. That constraint is what keeps each scope from drifting into the others.

## The work packages

| WP | Mission | Owns (`T#.#`) | Consumes (ICD from) |
|---|---|---|---|
| **A · nitro-on-railgun** | Non-custodial private **settlement of shielded value** over Nitro — *notes in → normal Nitro → notes out* — delivering **shielded payments** and the rail the others build on. | T0.0, T0.1, T0.2, **T0.3**, T0.6 *(opt)*, T0.7; settlement client T6.2, watchtower T6.3 | B (feeds), D (host) |
| **B · watcher parties** | Private, metered, proof-carrying **sync** so browser/mobile/server peers discover & verify pool/adjudicator state without leaking the anonymity set. | T1.0, T1.1, T1.2, T2.0, T2.1, T2.2, T2.3, T2.4 *(v2)*; note-scanner T6.1 | A (contract event ABIs to index) |
| **C · yield & exchange** | Let users **earn yield (USDC first) and swap privately**, cleared over A's rail; owns the venue **and the adapters**. | T4.0, T4.1, T4.3, T4.4; T4.2 *(v2)*, T4.5 *(v2)*, T3.\* *(v2)*, T4.6 *(v2)*; **T5.0 adapters (CCTP built · Aave-v4 yield · Swaps), T5.1 routing** | A (settlement), B (feeds); Aave · Circle-CCTP (external protocols) |
| **D · client** | The **wallet/app** that hosts A/B/C on a phone with almost no fixed infra. | T6.0, T6.4, T6.5, T6.6, T6.7, T6.8 *(opt)* | A, B, C (their client edges) |

Note the deliberate split of the T6 client tier: **A** owns the settlement client (T6.2) and watchtower (T6.3), **B** owns the note-scanner (T6.1), and **D** owns the host, transport, proving, and app shell — because each of those pieces is a client edge of a different scope, hosted by D.

## Interface contracts (ICDs)

Each work package binds to its siblings only through these; the interface *is* the boundary.

- **A exposes:** deposit/payout contract address + ABI + events (`Deposit`/`Payout`, adjudicator `Challenge`/`Checkpoint`/`Finalized`); the channel-lifecycle API; the ForceMove app registry; the voucher format; the pool commitment/nullifier event ABI; circuit `wasm`/`zkey` artifacts; the POI allow-list root.
- **B exposes:** the proof-carrying feed schema (`getStorageAt → {value, proof}`); the Nitro voucher metering interface; the per-contract head-cursor freshness signal (which T6.3 gates on).
- **C exposes:** the quote RPC; the posted price + `priceSetter` authority; the quote/settle app id; the per-venue loyalty-key scheme; the capacity hint.
- **D exposes:** the Armada-branded app + SDK surface over A/B/C.

## Sequencing (risk-first, one walking skeleton)

Per [ADR-0012](./09-architecture-decisions.md#adr-0012) rule 5:

1. **A is work package #1** — it lands the deposit/payout interface every other WP binds to; nothing above it can integrate until that interface exists.
2. **Walking skeleton** — a thin slice spanning A + B + D (shield → deposit → trivial ForceMove settle → payout → scan on a fixturenet) retires integration risk before any tier is deepened (see §4).
3. **B's T1 ingestion is the long pole** — the nimbus-eth1 emitter is the one genuinely-unbuilt v1 item; spike it first (§11 R3, ADR-0009).
4. **C builds on A + B** once their interfaces are real; **D** proceeds in parallel against the same interfaces.

## Open decisions, by owning work package

Kept with their owner so they no longer contaminate the core scope:

- **A:** native amount-privacy (T0.6, opt, ADR-0005); cross-pool membership proofs (T0.7, opt, ADR-0006).
- **B:** federation bonding + threshold DKG (T2.4, v2); the T1 nimbus long pole (ADR-0009).
- **C:** **USDC yield is the priority and in scope** via the Aave-v4 adapter (ADR-0013); the open decision is the **v1 mechanism** — direct adapter recipe (public Design-A boundary) vs the private **LP-buffered rail** (T4.5, which needs v2 market-making, ADR-0011) — and may amend ADR-0011. **Adapter source + license** (build vs reuse Railgun's Cookbook Aave/Swaps recipes) is an open gate. ETH/wstETH yield (T4.4) is the trivial path; market-making stays v2 (ADR-0011).
- **D:** mobile transport crux (T6.5, ADR-0008); optional identity (T6.8).

## Relationship to the arc42 docs

The arc42 set is the **architecture description**; this WBS is the **delivery frame** around it. The tier docs (`T0`–`T6`) are the building-block detail each work package points into — A → T0/T6.2/T6.3, B → T1/T2/T6.1, C → T3/T4/T5, D → T6.0/T6.4–T6.8. When a work package is picked up for delivery, its scope is written to the six-part reuse-spec bar (goal + boundary + reuse inventory + delta + ICD + acceptance) defined in ADR-0012, against the frozen baseline.
