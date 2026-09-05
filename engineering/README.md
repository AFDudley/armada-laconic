# Armada × Laconic — Engineering Documentation

This folder is the engineering architecture documentation for the Laconic work that surrounds Armada's shielded pool. The system it describes is a non-custodial, privately-settled exchange and settlement substrate built almost entirely by composing existing, battle-tested parts: an unmodified Railgun shielded pool, vanilla `go-nitro` state channels, the cerc-io watcher stack, and a mobile wallet. Its load-bearing construction is **nitro-on-railgun** — *notes in, normal Nitro, notes out* — and the recurring theme throughout these documents is that the hard problem here is integration, not new cryptography.

The documentation is organized according to **[arc42](https://arc42.org)** (Starke and Hruschka), a fixed twelve-section template for architecture description. We adopted it deliberately, as recorded in [ADR-0001](./09-architecture-decisions.md#adr-0001), after an earlier single planning document collapsed under its own weight: it was trying to be a system decomposition, a work breakdown, a decision log, and a schedule all at once, and it had to be re-cut every time one of those roles pulled against another. arc42 separates those concerns into sections that each do one job, which is what keeps this set stable as the design evolves.

Three ideas make the structure hold together. **One model, many views:** the item registry in [§5 Building Block View](./05-building-block-view.md) is the single source of truth — every buildable item has one `T#.#` id carrying its tier, status, and release facets, so tiers and delivery groupings are views over one registry rather than competing hierarchies. **Immutable decisions:** every significant choice is an append-only ADR in [§9](./09-architecture-decisions.md); a decided ADR is superseded, never edited, so the reasoning stays legible. **A work breakdown over a fixed baseline:** because the requirements are given and most components already exist, delivery is organized as a WBS of interface-bounded work packages ([§0 Work Packages](./00-work-packages.md)) built incrementally against the frozen ADR baseline ([ADR-0012](./09-architecture-decisions.md#adr-0012)).

If you are new to the system, read it in this order: [§1 Introduction & Goals](./01-introduction-and-goals.md) for what is being built and why; [§0 Work Packages](./00-work-packages.md) for how the scope is split for delivery; [§4 Solution Strategy](./04-solution-strategy.md) for the approach and increments; then the [§5 registry](./05-building-block-view.md) for the whole item map. From there, follow any `T#.#` id into its tier document. When a document states a choice rather than a mechanism, it cites the governing `ADR-####`; follow that link for the context and the alternatives weighed.

## Sections

| § | Doc | Contents |
|---|---|---|
| 0 | [Work Packages (WBS)](./00-work-packages.md) | delivery breakdown into A/B/C/D scopes; operating model [ADR-0012](./09-architecture-decisions.md#adr-0012) |
| 1 | [Introduction & Goals](./01-introduction-and-goals.md) | thesis, v1 capabilities, quality goals, v1 requirements traceability |
| 2 | [Constraints](./02-constraints.md) | non-custodial, pool immutable, Ethereum-anchored, mobile-first |
| 3 | [Context & Scope](./03-context-and-scope.md) | Armada-builds vs Laconic-builds boundary; external systems |
| 4 | [Solution Strategy](./04-solution-strategy.md) | nitro-on-railgun; integration-not-new-crypto; delivery increments |
| 5 | [Building Block View](./05-building-block-view.md) | the tier stack + **item registry** (T0–T6 detail docs) |
| 6 | [Runtime View](./06-runtime-view.md) | swap, deposit→Nitro→payout, challenge/watchtower flows |
| 7 | [Deployment View](./07-deployment-view.md) | watcher parties (browser/mobile/server peers), relays, fixturenet |
| 8 | [Cross-cutting Concepts](./08-crosscutting-concepts.md) | privacy, settlement, transport, metering, identity, trusted setup |
| 9 | [Architecture Decisions](./09-architecture-decisions.md) | ADRs 0001–0012 |
| 10 | [Quality Requirements](./10-quality-requirements.md) | non-custody, read-time privacy, latency, force-closability |
| 11 | [Risks & Technical Debt](./11-risks-and-technical-debt.md) | cold-start, go-nitro maturity, ingestion long pole, mobile transport |
| 12 | [Glossary](../glossary.md) | canonical `glossary.md` (rendered to `../glossary.html`); terms & collision resolutions |

The seven building-block detail documents referenced by §5 are the tier specs: [`T0-ethereum-anchor`](./T0-ethereum-anchor.md), [`T1-ingestion`](./T1-ingestion.md), [`T2-watcher-substrate`](./T2-watcher-substrate.md), [`T3-ordering`](./T3-ordering.md), [`T4-execution-venue`](./T4-execution-venue.md), [`T5-adapters`](./T5-adapters.md), and [`T6-client-apps`](./T6-client-apps.md). A public-facing overview of the same system lives at [`../index.html`](../index.html).

## Conventions

A few rules keep the set consistent. Refer to items by their `T#.#` id and to decisions by their `ADR-####` id, so that cross-references survive edits and reorganization. Treat the ADR log as append-only: when a decision changes, supersede the old ADR rather than editing it. Finally, the project's terminology is fixed by [ADR-0010](./09-architecture-decisions.md#adr-0010), which resolves the word collisions that repeatedly caused confusion — use *deposit/payout contract* rather than "adapter," *boundary* rather than "seam," *tier* rather than "layer," and *venue* or *clearing* rather than "marketplace."
