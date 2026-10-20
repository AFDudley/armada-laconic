# T2 · Watcher substrate

**Status:** in progress · tier doc · 2026-10-20
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T2
**Depends on / Blocks:** T1 (the proxy is what a client trusts); T2 only caches what T1 can re-verify.

The earlier docs planned a bonded federation of watcher parties serving proof-carrying feeds over `watcher-ts`, `ts-nitro`, `@cerc-io/peer`, and Waku. None of that exists in any of the ship-armada repos. With the nimbus verified proxy as the proof path (T1, ADR-0009), a server-side watcher becomes optional: a cache a client may read from, then re-verify through the proxy. It is never a trust boundary.

```
  nimbus verified proxy (T1, the trust boundary)
        ▲
        │  re-verify anything read from the cache
        │
  ┌─────────────────────────────┐
  │  optional watcher-ts cache   │  server-only, Postgres-backed
  │  (T2.2, not built)           │
  └─────────────────────────────┘
        ▲
        │  today's live cache is a different system:
  ┌─────────────────────────────┐
  │  armada-relayer's Ponder     │  built, out of the plan
  │  watcher (Postgres)          │
  └─────────────────────────────┘
```

## T2.0 Verified feeds — superseded by T1

**What it is.** The earlier plan's proof-carrying GraphQL feed (`watcher-ts` `getStorageAt → {value, proof}`) is not needed. The client verifies directly against the nimbus proxy, so no server-side proof-carrying feed sits between the client and the chain.

## T2.1 Relay fee — built, a different mechanism

**What it is.** `armada-relayer`'s actor charges a flat USDC fee from a gas estimate (`relayer-v2/.../fee-calculator.ts`, fee = gas estimate × gas price), not go-nitro payment vouchers. No voucher code (`payments.ts`-equivalent) exists in `armada-poc`, `armada-relayer`, or `armada-interface`.

## T2.2 Indexer — not built (watcher-ts); built, to be replaced (Ponder)

**What is live today.** `armada-relayer`'s `relayer-v2/watcher/` is a Ponder indexer (`ponder` `0.16.8`) over Postgres 16, under Docker Compose. It indexes `MessageTransmitter`/`PrivacyPool`/`PrivacyPoolClient` ABI events by direct RPC, and serves `/v2/quick-sync`, which `armada-sdk` and `armada-interface` use. It has one operator and no proofs. This is out of the plan (`labbook/2026-10-19-unified-stack-and-verified-note-sync.md` decision 2); neither the relayer's Ponder indexer nor the cerc-io Ponder fork is part of the plan.

**What `watcher-ts` would add.** `watcher-ts@18ca4e1a:packages/util/src/database.ts:98` asserts a Postgres backend; the indexer packages depend on `typeorm`, `pg`, `express`, `apollo-server-express`. It runs server-only today. No React Native or browser indexing code exists. Its event "proof" field (`{blockHash, receiptCID, log}`) has no link to `receiptsRoot` and would not be trusted even if used; it attaches structure, not a checked proof (`util/src/indexer.ts:706-715`; the only check in the codebase, `assertProofData`, checks that fields are present, not that they verify).

## T2.3 Transport — not built

No Waku, no libp2p dependency, and no code of either exists in the client, relayer, or interface repos checked.

## T2.4 Federation + bond — design only, not built

**The idea (not yet code).** A client's view of a note-ciphertext and nullifier stream benefits from a cache, because the receipts of every pool-touching block, fetched one at a time through the proxy, could be slow from a cold start. A watcher network could serve a **snapshot** at block N as that cache. A new client verifies the snapshot, then syncs from N to head with full proofs through the proxy.

- The commitment part of a snapshot is fully verifiable against `rootHistory`, independent of who served it.
- Ciphertexts and the nullifier set are not: the tree root does not cover ciphertexts, and the nullifier set has no on-chain root. A bad ciphertext or a missing nullifier from a snapshot causes delay or a hidden note, not theft.
- Independent watchers would sign the snapshot CID, and a client would accept a CID that several of them sign. `watcher-ts` state checkpoints (IPLD, by CID) are a plausible mechanism to reuse for this, if `watcher-ts` is ever adopted for this purpose.

No bonding, no DKG ceremony, and no signer set exist. This is a design note (`labbook/2026-10-19-unified-stack-and-verified-note-sync.md` decision 7), not a build in progress.

## What stays true across the tier

A client never trusts a watcher. It re-verifies anything it reads from a cache through the T1 proxy: commitment roots against `rootHistory`, and any claimed log against a receipt-trie rebuild. The watcher/watchtower/keeper/relay terminology distinction from the earlier docs still holds: a watcher serves and caches; a watchtower (T6.3) additionally responds to a stale-state challenge; this tier covers only the read-and-cache half, and no implementation of any of the three exists outside `armada-nitro`'s own test harness watchtower.

## Sources

- `armada-relayer` `origin/main` — `relayer-v2/watcher/ponder.config.ts`, `fee-calculator.ts`.
- `watcher-ts@18ca4e1a` — `packages/util/src/database.ts`, `util/src/indexer.ts`.
- `labbook/2026-10-19-unified-stack-and-verified-note-sync.md` (the unification and no-Ponder, no-React-Native-indexing rulings; the snapshot proposal).
- `labbook/2026-10-19-nimbus-verified-proxy-proofs.md` (why the proxy, not a server feed, is the trust boundary).
- Site docs: [`T1-ingestion.html`](./T1-ingestion.html), [`09-architecture-decisions.html`](./09-architecture-decisions.html#adr-0009).
