# armada-laconic

Design and collaboration documents for delivering Armada's Ethereum-privacy services
on Laconic infrastructure. Two additive offers for Armada's shielded pool — **more
yield** and the **mobile-first private support services** that surround it — so Armada's
team can stay focused on the crypto core. Non-custodial, and we never touch the immutable pool.

**▶ Live site — [afdudley.github.io/armada-laconic](https://afdudley.github.io/armada-laconic/)**

## What's on the site

- **[index.html](./index.html)** — *Overview.* The two offers, how they plug into
  Armada's three-tier model, and the tier map.
- **[architecture.html](./architecture.html)** — *The tier stack.* Seven tiers
  (T0 Ethereum anchor → T6 client): watcher parties, Nitro, Schnorr DSS, ordering
  & fault model, ex_net execution platform, adapters, identity, and the privacy
  boundary — with line-anchored citations to existing implementations.
- **[build-plan.html](./build-plan.html)** — *Status & backlog.* What's built vs.
  partial vs. net-new per tier, with a per-component confidence mark and a
  sequenced net-new backlog.
- **[execution-platform.html](./execution-platform.html)** — *ex_net.* The
  frequent-batch-auction, curve-order matcher (T4) that powers the swaps and yield
  adapters: lineage, the 2019–2020 matcher PoC, CLOB+RFQ in one order type, and
  capital efficiency vs. AMMs.
- **[laconic_ethereum_privacy_via_armada.html](./laconic_ethereum_privacy_via_armada.html)** —
  *The thesis.* Laconic as the sync / metering / client substrate beneath Armada's
  Railgun-circuit shielded pool.
- **[ex_net_whitepaper.pdf](./ex_net_whitepaper.pdf)** — Vulcanize's 2017 design ancestor.

## Code

Gitea-only Laconic repos (`git.vdb.to`) are vendored under [`code/`](./code) as
shallow, history-free snapshots at the pinned commits, so citations stay reviewable
if Gitea is unavailable.
