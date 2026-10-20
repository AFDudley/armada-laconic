# 7. Deployment View

arc42 §7 · 2026-10-20

This view describes where the §5 building blocks run today, and on what the plan would
add. Today's deployment is a conventional server-client topology, not a peer-to-peer
federation: `armada-interface` is a web app talking plain RPC/wagmi to Ethereum L1 and
to `armada-relayer`, a conventional Express/Postgres service. The earlier plan's P2P
watcher federation (bonded peers, Waku pub/sub, libp2p-noise) has no code anywhere in
the client, relayer, interface, or SDK repos (ruling 7, brief §4).

## Infrastructure nodes — what runs where

### Ethereum L1 (on-chain)

The settlement substrate. The T0 contracts are deployed here: the shielded pool
(T0.0, built), the merged Nitro adjudicator with the deposit/payout boundary
(T0.2, built), and the governance/revenue contracts (built). No registry or bond
contract exists yet (T0.4, not built). Fixturenet, testnet, and mainnet differ only in
*which* L1 this set is deployed to (see Environments). → §5 T0.

### Relayer node — `armada-relayer` (server, today)

`armada-relayer` runs a Ponder-based indexer over Postgres (`relayer-v2/watcher`),
reading `MessageTransmitter`/`PrivacyPool`/`PrivacyPoolClient` events by plain RPC, and
an actor that submits transactions and charges a flat USDC fee from a gas estimate
(`fee-calculator.ts`, T2.1). This is a conventional server, not a bonded P2P
federation, and it is not the proof path: a client that reads from it is trusting the
server, not verifying a proof (T2.2).

### Proof path — the nimbus verified proxy (upstream, unintegrated)

`status-im/nimbus-eth1` ships a verified proxy — native (`libverifproxy.a`), WASM, and
a CLI — that anchors a header to a beacon light client and verifies storage and
receipts client-side (T1.0, T1.1). No Armada repo runs or embeds it today: no server
operates it against any Armada chain, no WASM bundle ships in `armada-interface`, and
`libverifproxy.a` has no iOS or Android cross-compile target yet. Once integrated, this
runs inside the client process, not as a separate server node (`client (T2 caches,
re-verified through the proxy)` in [`T1-ingestion.md`](./T1-ingestion.md)). → §5 T1.

### Venue node — plan, not built

The earlier plan calls for an off-chain venue node holding a `priceSetter` authority
and filling quotes from a pre-funded static inventory (T4.0, T4.1), with a later
automated market-making keeper (T4.2, v2 plan). No venue node, no price-setter
contract, and no quote/settle app exist in `armada-poc` or `armada-nitro`. → §5 T4.

### Client — `armada-interface` (end-user device, today)

The live client is `armada-interface`, a React 19 + Vite web app that runs in a
browser, not a phone-resident P2P peer. It holds keys in browser memory
(`keyManager.ts`, T6.0), scans and trial-decrypts notes locally (`sync.ts`, T6.1), and
drives the pool's shield/transact calls (`handler.ts`, T6.2). It talks to Ethereum L1
and to `armada-relayer` over plain HTTP/RPC, not libp2p or Waku. No channel lifecycle
code, no client-side watchtower (T6.3), and no mobile build exist. `armada-client`, the
repo meant to host a unified TypeScript stack across Linux server, browser, iOS, and
Android (`docs/SPEC.md`), has no application code (T6.5).

## Deployment diagram (blocks → nodes)

```
                     ┌──────────────────────────────────────────┐
                     │  ETHEREUM L1  (fixturenet | testnet |     │
                     │  mainnet)                                 │
                     │  T0.0 pool · T0.2 adjudicator (merged     │
                     │  deposit/payout) · governance/revenue     │
                     └──────────────────────────────────────────┘
                          ▲ shield / deposit / settle / challenge
                          │
   ┌──────────────────────┴───────────────────────────────────┐
   │  armada-relayer (server, today)                           │
   │  Ponder/Postgres indexer · flat-fee actor (T2.1, T2.2)    │
   │  not a proof path — a client trusts this server            │
   └─────────────────────────────────────────────────────────┘
                          │ plain RPC / HTTP
                          ▼
   ┌─────────────────────────────────────────────────────────┐
   │  armada-interface (browser, today)                        │
   │  keys in memory (T6.0) · note-scan + browser prover (T6.1)│
   │  shield/transact handlers (T6.2) · no channel lifecycle    │
   └─────────────────────────────────────────────────────────┘

   ─────────────── plan, no code today ───────────────────────
   nimbus verified proxy integration (T1) — would run inside the
   client process, verifying reads instead of trusting the relayer
   venue node (T4) — posted-price contract + quote/settle app
   client watchtower (T6.3) — only a test-harness watchtower exists
   mobile client (T6.5) — armada-client's docs/SPEC.md plan
```

## Environments

| Environment | Purpose | L1 target |
|---|---|---|
| **fixturenet** | `armada-nitro`'s e2e gate harness runs deterministic tests here, including the test-only watchtower (`env/gates/watchtower/watch.ts`). | orchestrated local L1 (real `geth` + `lighthouse`, per `docs/ENVIRONMENT.md`) |
| **Sepolia** | A real deploy exists: `armada-poc`'s `scripts/deploy_privacy_pool.ts` deploys `PrivacyPool` with a 50 bps shield fee. | Sepolia |
| **mainnet `armada-1`** | Crowdfund, governance, and CCTP are live. The pool ships in a later launch; the Phase-2 ceremony is not finished (`v1.0.0-ceremony` tag, shapes "open", no beacon). | Ethereum L1 |

→ §6 for the runtime scenarios these nodes execute · §5 T0/T1/T2/T4/T6 for
building-block detail · §11 for the proof-path-integration and mobile-client gaps.
