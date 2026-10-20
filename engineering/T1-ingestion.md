# T1 · Ingestion

**Status:** proven by hand on the fixturenet, unintegrated · tier doc · 2026-10-20
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T1
**Depends on / Blocks:** consumes T0.0 (pool event ABI, storage layout) and T0.2 (adjudicator events, the merged deposit/payout boundary). Blocks T2 (the proxy is what a client trusts; T2 only caches).

T1 is the proof path: how a client, on any target, verifies Ethereum state against a header a beacon light client checks. `status-im/nimbus-eth1` ships a verified proxy that does this job upstream of the client (ADR-0009, `labbook/2026-10-19-nimbus-verified-proxy-proofs.md`). The patched proxy has proved one pool log and one storage slot on the armada-nitro fixturenet, by hand (`labbook/2026-10-20-nimbus-fixturenet-and-proxy-proof.md`). No client wires it in yet.

```
  beacon light client (trusted block root)
        │
        ▼
  ┌──────────────────────────────┐
  │ nimbus verified proxy         │  native / WASM / libverifproxy.a
  │  verifies headers, storage    │
  │  (EIP-1186), and receipts     │
  │  (eth_getBlockReceipts +      │
  │  receipt-trie rebuild)        │
  └──────────────────────────────┘
        │  verified {value, proof} / verified receipts
        ▼
     client (T2 caches, re-verified through the proxy)
```

## T1.0 Proof source — the nimbus verified proxy

**What it is.** `status-im/nimbus-eth1` ships `libverifproxy`, a C static library with Go bindings, and `nimbus_verified_proxy_wasm`, a WASM build (`Makefile:346-398`). It resolves a block tag to a header the light client has verified and recomputes the block hash (`engine/blocks.nim:105-128,148-173`), exposing `stateRoot` and `receiptsRoot`. It verifies an account proof against `stateRoot` and a storage proof against `storageRoot` (`engine/accounts.nim:26-125,197-244`). The README marks it "Stability: experimental."

**What is unbuilt.** No Armada repo runs the proxy today. There is no server operating it against any Armada chain, and no WASM bundle ships in `armada-interface`. `libverifproxy.a` builds for Linux, iOS, and Android (Android needs `-d:PREFER_HASHTREE_SHA256=false`); no client repo links it in.

## T1.1 Inclusion and completeness

**What it is.** `engine/receipts.nim:53-57` fetches a block's receipts with `eth_getBlockReceipts` and returns an error unless the rebuilt receipt trie's root matches `header.receiptsRoot`. This proves the inclusion of any one log in that receipt set, and it proves completeness of the block's logs at the same time: an omitted log would change the rebuilt root. `engine/rpc_frontend.nim:383-391` returns `eth_getBlockReceipts` only after this check.

**Why not `eth_getLogs`.** A range query over `eth_getLogs` verifies each log it returns, but it cannot detect a log it silently omits (`engine/receipts.nim:132-173`). Note sync therefore fetches the receipts of each pool-touching block and filters locally, not a log range.

## T1.2 Fixturenet patch

**What it is needed for.** The first proof on `armada-nitro`'s fixturenet needs the proxy to trust the fixturenet's own consensus genesis, not mainnet or a public testnet. `engine/engine.nim:140-142` loads `getMetadataForNetwork`/`genesisParamsForNetwork` by name; `engine/genesis_params.nim:48-70` hard-coded `genesisTime`, `genesisValidatorsRoot`, `genesisBlockRoot`, and the EIP-2935 fork time only for a short, named list of public networks; any other name fell back to mainnet values. The execution chain id was never the obstacle — a fixturenet can use any chain id, and `library/verifproxy.nim:46-52` only maps a name to a chain id. The obstacle was the consensus genesis.

**What is built.** A custom-network patch loads genesis time, genesis validators root, genesis block root, and the EIP-2935 fork time from the fixturenet's own testnet directory, which lighthouse already serves. The patch lives on `AFDudley/nimbus-eth1` `armada/verified-proxy-fixes` at `b0205d1d` (upstream base `master@2c8eabf`), alongside a two-line `convHeader` fix (`827f95ff`) that copies `blockAccessListHash` (EIP-7928) and `slotNumber` (EIP-7843) so the proxy also verifies post-Glamsterdam headers. Neither commit is filed upstream.

## Where this leaves the earlier plan

The nimbus state-diff emitter, an IPLD mapper, and the cerc-io `ipld-eth-server` pipeline are not on the client's proof path. That pipeline is geth-only (`plugeth` `v1.13.14-cerc-2`) and dormant since August 2024. `ipld-eth-server`'s current schema (statediff v5) stores logs and receipts as flat IPLD objects with no receipt trie, so no proof connects a single log to `receiptsRoot` in that schema; an earlier schema (statediff v4) did build and publish a receipt trie, but v5 removed the fields that made it walkable (`ipld-eth-db` `9f060ff`, `f67f034`), with no recorded reason. The nimbus proxy does the same verification upstream and needs none of this. A narrower version of the pipeline returns as a watcher-only archive layer, not a client proof path: see "Archive layer for watchers" below.

## The first proof

Run by hand on the armada-nitro fixturenet, 2026-10-20 (`labbook/2026-10-20-nimbus-fixturenet-and-proxy-proof.md` §4.4):

1. On the fixturenet, pool-e2e shielded 10 USDC into the pool, logging a `Shield` event in block 886.
2. The patched proxy (`b0205d1d`) ran with `--network=<the fixturenet's testnet directory>` and a `--trusted-block-root` from nimbus-eth2's finalized header. Its beacon source was nimbus-eth2; its execution source was nimbus-eth1.
3. `eth_getBlockReceipts(886)` through the proxy rebuilt the receipt trie against the anchored `receiptsRoot` and returned the `Shield` log in 0.65 s. This proved inclusion and completeness.
4. `eth_getStorageAt` through the proxy, on the `rootHistory[tree][root]` slot for the new root, returned `0x…01`. Both lighthouse and geth produce the chain; neither answer above is trusted from them.

**Why lighthouse is not the light-client source on this fixturenet.** Lighthouse v8.2.2 serves, as its best update for sync-committee period 0, the first signed update: attested slot 11, finalized slot 0 (not real finality). Its tiebreak rule prefers this older update over later ones within the same period (`consensus/types/src/light_client/light_client_update.rs:402-463` at `e423a66`), so the proxy's finality-branch check rejects every update it offers. nimbus-eth2 serves a different best update for the same period (attested slot 141, finalized slot 64), which the proxy accepts. Which client departs from the spec for an update whose finalized header is still at genesis is open; neither side has a verdict yet.

## Archive layer for watchers

nimbus-eth1 serves no state below its base: `execution_chain/rpc/server_api.nim:111-112` returns "Historical data not available" for `eth_getProof` below that point. A client never needs this: it verifies recent state through the proxy and reaches old receipts through Portal (nimbus Fluffy; the EF's own `trin` and `Ultralight` are idle) or era files. A watcher that builds snapshots or checks deep headers does need an archive layer. The feasible build is an emitter from nimbus-eth1's Aristo database into an `ipld-eth-server`-style `ipld.blocks` store, estimated at 4 to 5 points — Aristo already encodes standard RLP trie nodes (`aristo_serialise.nim:67-75`, `aristo_proof.nim:49-58`), and `ipld-eth-server`'s proof path keys nodes by a hash-derived CID that needs no server change. Fallbacks, if the emitter is not built: geth v1.17 with `--history.trienode=N`, or Erigon v3.4+ with `--prune.include-commitment-history=true` (`labbook/2026-10-20-archive-state-for-nimbus.md`, `labbook/2026-10-20-proxy-and-archive-open-questions.md`).

## Open questions that only a further build or run can answer

- Does the fixturenet proof still pass once the chain has finalized past genesis, separating a lighthouse defect from a genesis-era edge case?
- What sets the early exit inside `deliverBeaconTransport` in the WASM build, so the browser proxy's first `sync()` call stops stalling?
- What is the measured size of receipt-trie nodes per block, for watchers that want per-log proofs rather than per-block snapshots?

## Sources

- `status-im/nimbus-eth1` `master` (`4df52ef`, 2026-10-19) — `engine/blocks.nim`, `engine/accounts.nim`, `engine/receipts.nim`, `engine/rpc_frontend.nim`, `engine/engine.nim`, `engine/genesis_params.nim`, `nimbus_verified_proxy_conf.nim`, `library/verifproxy.nim`, `Makefile`, `aristo_serialise.nim`, `aristo_proof.nim`.
- `AFDudley/nimbus-eth1` `armada/verified-proxy-fixes` (`b0205d1d`, base `master@2c8eabf`) — the `convHeader` fix and the custom-network patch.
- `armada-nitro` `integrated` (`990c9fa9`, pebble `arn-b58`) — the nimbus pair in the fixturenet; `docs/ENVIRONMENT.md` (fixturenet consensus: `geth` + `lighthouse` + the nimbus pair).
- `labbook/2026-10-19-nimbus-verified-proxy-proofs.md` (full assessment and work-item table); `labbook/2026-10-20-nimbus-fixturenet-and-proxy-proof.md` (the nimbus pair and the first proof); `labbook/2026-10-20-archive-state-for-nimbus.md`, `labbook/2026-10-20-proxy-and-archive-open-questions.md` (the archive layer).
- Site docs: [`09-architecture-decisions.html`](./09-architecture-decisions.html#adr-0009), [`build-plan.html`](./build-plan.html) (S7, S8).
