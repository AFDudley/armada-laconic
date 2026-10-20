# T1 · Ingestion

**Status:** in progress, upstream, unintegrated · tier doc · 2026-10-20
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T1
**Depends on / Blocks:** consumes T0.0 (pool event ABI, storage layout) and T0.2 (adjudicator events, the merged deposit/payout boundary). Blocks T2 (the proxy is what a client trusts; T2 only caches).

T1 is the proof path: how a client, on any target, verifies Ethereum state against a header a beacon light client checks. The earlier docs planned a nimbus-eth1 state-diff emitter mapped into IPLD, served by `watcher-ts`. `status-im/nimbus-eth1` already ships a verified proxy that does this job upstream of the client, so that path is replaced (ADR-0009, `labbook/2026-10-19-nimbus-verified-proxy-proofs.md`).

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

**What is unbuilt.** No Armada repo runs the proxy today. There is no server operating it against any Armada chain, no WASM bundle shipped in `armada-interface`, and no mobile cross-compile (`Makefile:66-73` has no iOS or Android target).

## T1.1 Inclusion and completeness

**What it is.** `engine/receipts.nim:53-57` fetches a block's receipts with `eth_getBlockReceipts` and returns an error unless the rebuilt receipt trie's root matches `header.receiptsRoot`. This proves the inclusion of any one log in that receipt set, and it proves completeness of the block's logs at the same time: an omitted log would change the rebuilt root. `engine/rpc_frontend.nim:383-391` returns `eth_getBlockReceipts` only after this check.

**Why not `eth_getLogs`.** A range query over `eth_getLogs` verifies each log it returns, but it cannot detect a log it silently omits (`engine/receipts.nim:132-173`). Note sync therefore fetches the receipts of each pool-touching block and filters locally, not a log range.

## T1.2 Fixturenet patch

**What it is needed for.** The first proof on `armada-nitro`'s fixturenet needs the proxy to trust the fixturenet's own consensus genesis, not mainnet, Sepolia, or Hoodi. `engine/engine.nim:140-142` loads `getMetadataForNetwork`/`genesisParamsForNetwork` by name; `engine/genesis_params.nim:48-70` hard-codes `genesisTime`, `genesisValidatorsRoot`, `genesisBlockRoot`, and the EIP-2935 fork time only for `sepolia` and `hoodi`; any other name falls back to mainnet values. `nimbus_verified_proxy_conf.nim:44-48` lists only `mainnet`, `hoodi`, `sepolia`, and the OP Stack networks under `--network`. `armada-nitro`'s fixturenet runs its own `lighthouse` beacon genesis (`docs/ENVIRONMENT.md:7,29`), so the proxy needs a patch that reads genesis and runtime-config parameters from the fixturenet's testnet directory. [INFERENCE] Estimated at 1 to 2 points: the patch touches `genesis_params.nim` and the metadata load, and nimbus-eth2 already loads custom network directories for its own testnets.

The execution chain id is not the obstacle — a fixturenet can use any chain id, and `library/verifproxy.nim:46-52` only maps a name to a chain id. The obstacle is the consensus genesis.

## Where this leaves the earlier plan

The nimbus state-diff emitter, an IPLD mapper, and the cerc-io `ipld-eth-server` pipeline are not on the proof path. That pipeline is geth-only (`plugeth` `v1.13.14-cerc-2`) and dormant since August 2024. `ipld-eth-server`'s current schema (statediff v5) stores logs and receipts as flat IPLD objects with no receipt trie, so no proof connects a single log to `receiptsRoot` in that schema; an earlier schema (statediff v4) did build and publish a receipt trie, but v5 removed the fields that made it walkable (`ipld-eth-db` `9f060ff`, `f67f034`), with no recorded reason. The nimbus proxy does the same verification upstream and needs none of this.

## The first proof (planned, not yet run)

1. On the fixturenet, shield into the pool and record the `Shield` log's block.
2. Point the patched proxy (T1.2) at `lighthouse`'s beacon REST API and at `geth`'s RPC, with a trusted block root from `lighthouse`.
3. Call `eth_getBlockReceipts(block)` through the proxy and find the `Shield` log. This proves inclusion and completeness.
4. Compute the `rootHistory[tree][root]` slot key for the new root. Call `eth_getStorageAt` through the proxy and expect a non-zero value.

## Open questions that only a build or a run can answer

- Does the metadata loader accept a network directory path, or must the T1.2 patch add one? Does `lighthouse` v8.2.2 serve the light-client API by default, or only with a flag?
- Does `libverifproxy.a` cross-compile and link for iOS and Android?
- What are the WASM bundle size and the per-call latency in a browser?
- From a fresh trusted block root, how far back can the proxy prove? The block walk stops at `maxBlockWalk` (1000) and `maxWindowJumps` (500 jumps of 8191 blocks) (`library/verifproxy.nim:149-150`).
- How large is an archive node for the target chain, and how many receipt bytes does a pool-touching block carry?

## Sources

- `status-im/nimbus-eth1` `master`, clone at `.scratch/nimbus-eth1` in `armada-laconic` (`4df52ef`, 2026-10-19) — `engine/blocks.nim`, `engine/accounts.nim`, `engine/receipts.nim`, `engine/rpc_frontend.nim`, `engine/engine.nim`, `engine/genesis_params.nim`, `nimbus_verified_proxy_conf.nim`, `library/verifproxy.nim`, `Makefile`.
- `armada-nitro` `origin/integrated` — `docs/ENVIRONMENT.md` (fixturenet consensus chain: real `geth` + `lighthouse`).
- `labbook/2026-10-19-nimbus-verified-proxy-proofs.md` (full assessment and work-item table); `labbook/2026-10-19-unified-stack-and-verified-note-sync.md` (the question this entry answers, and the addendum on receipt-trie history).
- Site docs: [`09-architecture-decisions.html`](./09-architecture-decisions.html#adr-0009), [`build-plan.html`](./build-plan.html) (WI-A..WI-H).
