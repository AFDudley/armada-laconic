# Leveraging Laconic for Ethereum Privacy (via Armada)

Forward-looking companion to the backward-looking audit in
[`laconic_asset_inventory.md`](./laconic_asset_inventory.md). Where the audit
says what Laconic *is*, this says what it can *become*: the infrastructure layer
beneath Ethereum privacy protocols, with Armada as the first aligned counterparty.

---

## The thesis, in one paragraph

Laconic's durable value is not its Cosmos chain or token — it is the software the
team built to serve verifiable Ethereum data cheaply and privately: **watchers**
(proof-carrying, metered state indexing), **Nitro state-channel micropayments**,
and a working **in-browser client stack** (MobyMask: MetaMask snap + payment node
+ p2p relay + keeper). Ethereum privacy protocols need exactly this substrate and
do not want to build it themselves. Armada is the first such protocol whose
architecture leaves a Laconic-shaped hole.

## What Armada is (and isn't)

Armada builds the **crypto core**: a Railgun-circuit shielded pool for USDC, with
governance-added **adapters** (cross-chain USDC via CCTP, Aave-v4 shielded yield,
swaps) and an **SDK** whose boundary is "build apps, never touch keys or circuits."
Armada deliberately does **not** build the indexing, payment-metering, or client
substrate around the pool. That gap is what Laconic fills.

## How Laconic plugs in — three points

1. **Anonymity-preserving sync (watchers).** A shielded pool's privacy dies at
   read-time: any app that scans the pool over a normal RPC endpoint fingerprints
   its users to that provider. A Laconic watcher over the pool serves
   proof-carrying, everyone-gets-the-same-bytes note-streams, metered by Nitro, so
   clients scan **locally and privately**. This directly protects the pool's shared
   anonymity set — the single property the whole design rests on.

2. **Metered payments + clearing (Nitro).** Armada's model is metered flows with
   integrator fee-sharing. Laconic already meters per request (Nitro vouchers, proven
   end-to-end in MobyMask) and **clears** them over state channels: vouchers net into
   co-signed channel allocations that split fees (user / protocol / integrator) and
   settle to L1 in **USDC**. No marketplace, no fee vault — just clearing.

3. **Reference client (MobyMask stack).** Armada's SDK boundary is exactly the
   MobyMask client shape: a snap for app-specific signing, an in-browser Nitro node
   for payments, and a p2p relay to a keeper that submits transactions. It is a
   head start on the Armada front-end and integrator app templates, including the
   mobile proving-cost path.

## This is the substrate Laconic already built

This is not a bespoke deal — it reuses what Laconic already built: verifiable Ethereum
data served privately, per-request metering, non-custodial clearing, and an in-browser
client stack. Armada is the first real customer for that substrate.

| Laconic capability | Armada need it fills |
|---|---|
| Watchers — proof-carrying, metered private sync | Protects the shielded pool's anonymity set at read-time |
| Nitro — per-request vouchers + channel **clearing** | Cheap metered fees + integrator revenue-share, settled in USDC |
| MobyMask / wallet client stack + p2p relay + keeper | The front-end / SDK shape and origin-private submission |

(Maturity: the pieces ship as libraries/demos — `watcher-ts`, `go-nitro` vouchers/channels,
MobyMask, the Laconic wallet; the pool-specific integration — Railgun watcher config, the
Nitro↔Railgun boundary, the clearing splits — is the new work. Nothing here requires a
service-provider marketplace, only clearing.)

## Commercial shape (simple)

- **One unified USDC fee** the user/integrator sees — not two stacked tolls.
- **Laconic runs and maintains** the shared watcher/relay/keeper fleet and takes a
  **fee-share that scales with served volume**.
- Laconic can charge on its own rail, so it is **not dependent on Armada's fee
  switch** — the two fee surfaces are merged into one coordinated schedule.
- **All code is open source** (or will be). Nothing to license. The collaboration
  is operation + fee-sharing, not IP transfer.

## Boundary

Laconic contributes **nothing** to the cryptography — the circuits, the pool, and
any threshold/MPC are Armada's, and that is the audit-gating path. Laconic makes
the layer *around* the crypto work: sync, payments, client, deployment, operation.

## Why it matters for Laconic's next chapter

It converts Laconic's durable IP into **recurring, volume-aligned USDC revenue**
against a funded, aligned counterparty — and does so by finally running the
network's original designed model against real demand, rather than the
never-delivered token-unlock path.
