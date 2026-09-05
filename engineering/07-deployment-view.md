# 7. Deployment View

arc42 §7 · 2026-09-04

This view describes where the §5 building blocks physically run, and on what
hardware. The headline topology fact is that the system is **not
client-server**: a watcher party is a **bonded peer federation** (T2.2) whose
overwhelming majority is browser tabs and phones talking peer-to-peer. Only a
handful of server-side emitters (T1) and a **circuit-relay + STUN/TURN**
bootstrap (T2.3) exist as fixed transport infrastructure. The wallet, including
its watchtower (T6.3), is **phone-deployable by construction**. Transport runs
on Waku pub/sub + libp2p-noise per ADR-0008, and the state source is
nimbus-eth1 per ADR-0009.

## Infrastructure nodes — what runs where

### Ethereum L1 (on-chain)
The settlement substrate. The **T0** contracts are deployed here, and this is the
one node no participant hosts: the clean-room, Railgun-spec-compatible shielded **pool** (T0.0),
the **Nitro adjudicator** (T0.2), the **deposit/payout contract** (T0.3), and, at
v2, the **registry + bond** (T0.4). Everything else is a client or observer of
these contracts. Fixturenet, testnet, and mainnet differ only in *which* L1 this
set is deployed to (see Environments). → §5 T0.

### State-diff emitter node — nimbus-eth1 → IPLD (T1, server-side)
This is the only *heavy* server role on the v1 spine and its long pole
(ADR-0009). A node runs the **nimbus-eth1 state-diff emitter** (T1.0) scoped to
the watched address set (pool + rail contracts), maps its Aristo/witness output
into **IPLD** proof-carrying diffs (T1.1), and indexes them with a **watcher-ts
ingest config** (T1.2). It holds only the watched-contract subtrie rather than a
full archive EL, and it publishes a per-contract **head cursor** (a liveness
signal) that T6.3 gates on. A phone-scoped emitter remains a stretch goal, **not**
a v1 node. Only a handful of these run; they are the servers in an otherwise P2P
network. → §5 T1.

### Watcher-party peers — P2P federation (T2, browser / mobile / server)
A **watcher party is a bonded federation of peers, not an RPC server** (T2.2).
Browser, mobile, and headless-server peers all run the same libp2p base
(`ts-nitro` over `@cerc-io/peer`) and are **first-class peers**: a phone is not
a second-class client of a server. Each peer serves proof-carrying
`getStorageAt → {value, proof}` feeds (T2.0) metered by Nitro vouchers (T2.1)
over that ingested state. Because responses are proof-carrying and
identical-bytes, a consumer trusts the math rather than the peer.

**Fixed transport infra is deliberately minimal (ADR-0008, T2.3):**
- **Waku pub/sub** — feed discovery, gossip, async note-streams; the transport
  Railgun's broadcaster network already speaks. No privileged Waku node is ours
  to run; peers subscribe by content-topic.
- **libp2p-noise direct streams** — the 1:1 voucher/quote hot loop (tens-of-ms).
- **circuit-relay + STUN/TURN bootstrap** — the *only* fixed transport infra, so
  browser/mobile peers behind NAT can dial each other. Pure transport: no trust,
  no on-chain role. → §5 T2.3.

At v2 the federation bonds against T0.4 and signs what it serves with a
threshold Schnorr key (T2.4). These are the same peers with an added
economic-security role.

### Venue node — posted price v1; solver keeper v2 (off-chain)
In **v1** the venue node is minimal: it holds the on-chain posted-price authority through a single `priceSetter` wallet (T4.0, migratable to governance) and fills quotes at the posted price from a **pre-funded static inventory**. It is a service node, never custody — every fill is atomic against its own inventory over a channel. The full **maker keeper** — watching T2 feeds, internalizing flow, hedging the residual via Railgun's RelayAdapt recipe through a broadcaster, and rebalancing Aave — is the **v2** automated market-maker (T4.2, ADR-0011). → §5 T4.2.

### Mobile wallet + self-watchtower (T6, end-user device)
The wallet is everything a real user runs on their own phone or browser: shield,
trade, settle, and **scan their own notes locally** (T6.1), with keys never
leaving the OS secure enclave (T6.0). The wallet is a **watcher-party (P2P)
client**, not a client of a server; it talks peer-to-peer to bonded parties for
feeds and to counterparties for the settlement hot loop.

The **self-hosted watchtower (T6.3) runs on the phone** as a lightweight reactive
loop. It watches T0.2 `Challenge` events over the T2 feed and submits the user's
higher-turn co-signed state within the challenge window, needing no prover and no
heavy state. It is phone-deployable by construction and gated on T2 feed
freshness. Users wanting stronger liveness may *additionally* delegate the
identical loop to an always-on watcher party, but the default runs on-device.

**Two mobile deployment forms (T6.5, ADR-0008):**
- **Interim (spine demo):** the browser peer stack runs **inside a WebView** in
  the RN shell — `@cerc-io/peer` + `ts-nitro` unmodified. RN owns keys + UI; the
  WebView owns the Nitro node + p2p. Foreground-only, reuses production browser
  code verbatim — the realistic v1 mobile client.
- **Production (Phase-3, before mainnet):** the p2p/channel/messaging stack
  becomes **one native gomobile module** (go-nitro + go-waku, `library/mobile`,
  `.aar` / `.xcframework`), escaping the WebView background-execution constraint
  and giving the phone a first-class native libp2p peer. The transport interface
  T6.2/T6.3 bind to is identical across both forms. → §5 T6.5.

## Deployment diagram (blocks → nodes)

```
                     ┌──────────────────────────────────────────┐
                     │  ETHEREUM L1  (fixturenet | testnet |     │
                     │  mainnet)                                 │
                     │  T0.0 pool · T0.2 adjudicator ·           │
                     │  T0.3 deposit/payout · T0.4 registry(v2)  │
                     └──────────────────────────────────────────┘
                          ▲ settle / challenge      │ canonical head
        RelayAdapt hedge  │                          ▼
                          │        ┌───────────────────────────────┐
   ┌──────────────────┐   │        │ EMITTER NODE (server, few)    │
   │ VENUE SOLVER     │───┘        │ T1.0 nimbus-eth1 → T1.1 IPLD  │
   │ KEEPER (T4.2)    │            │ → T1.2 watcher-ts ingest cfg  │
   │ off-chain maker  │            └───────────────────────────────┘
   │ priceSetter T4.0 │                     │ proof-carrying diffs
   └──────────────────┘                     ▼  + head cursor
                          ┌───────────────────────────────────────┐
     fixed infra only:    │   WATCHER-PARTY  P2P FEDERATION (T2)   │
   ┌───────────────────┐  │  proof-carrying feeds T2.0 · meter T2.1│
   │ circuit-relay +   │◀─┤  bonded threshold-Schnorr (T2.4, v2)   │
   │ STUN/TURN (T2.3)  │  └───────────────────────────────────────┘
   └───────────────────┘   ▲ Waku pub/sub (gossip) · libp2p-noise (1:1)
             │             │           T2.2 peers: browser/mobile/server
             ▼             ▼
   ┌───────────────────────────────────────────────────────────────┐
   │  END-USER DEVICES — peers, NOT clients of a server             │
   │  browser build   ·   MOBILE WALLET (T6): WebView(interim) /    │
   │                      native gomobile(prod, T6.5)               │
   │  local note-scan T6.1 · settlement client T6.2 ·               │
   │  PHONE-RESIDENT self-watchtower T6.3 · Groth16 prover T6.6     │
   └───────────────────────────────────────────────────────────────┘
```

## Environments

| Environment | Purpose | L1 target |
|---|---|---|
| **fixturenet** (primary) | Deterministic E2E under the **laconic stack orchestrator**: shield → deposit → Nitro → challenge → payout, proofs against our keys. Every tier's inline tests run here (T0–T6). | orchestrated local L1 |
| **public testnet** | Shared multi-party rehearsal — real network conditions, NAT traversal over the T2.3 bootstrap, broadcaster interop. | public test L1 |
| **mainnet** (later) | Production. Gated: pool-spend authorization (T0.0/T0.3/T0.6), the T2.4/T3 DSS surface, and the T6.5 native gomobile port are all **Phase-3, audit-critical before mainnet**. | Ethereum L1 |

The **fixturenet is the primary deterministic test environment** and the one on
which the whole v1 spine is demonstrated: the end-to-end v1 deliverable, a
private shielded wstETH → USDC swap, is a fixturenet run (T4). Testnet adds real
P2P conditions, and mainnet follows the audit gates rather than a schedule.

→ §6 for the runtime scenarios these nodes execute · §5 T1/T2/T4/T6 for
building-block detail · §11 for the ingestion and mobile-transport long poles ·
ADR-0008 (transport), ADR-0009 (ingestion).
