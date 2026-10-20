# Full proofs for Armada with the nimbus verified proxy

Lab book entry. Session 2026-10-19, after 23:36 UTC.
Session id `01a14aa3-1d15-74fd-9f8d-df18ee172022`.
Repos read: `status-im/nimbus-eth1` (clone at
`armada-laconic/.scratch/nimbus-eth1`, `4df52ef`, 2026-10-19), `armada-nitro`
(`origin/integrated`), and the audit files listed below.

This entry follows `2026-10-19-unified-stack-and-verified-note-sync.md`. It
answers that entry's open item "The proof source".

## 1. The question

How difficult is it to give Armada the full proofs it needs, with nimbus-eth1 as
the Ethereum data source? A client on any target must verify, against a header
that a beacon light client checks:

1. the execution header, which gives `stateRoot` and `receiptsRoot`;
2. storage values: `rootHistory[tree][root]`, `nullifiers[tree][n]` and the
   adjudicator's slots;
3. the inclusion of one pool or armada-nitro log;
4. the completeness of a block's pool logs;
5. a snapshot that the watcher network serves as a cache.

## 2. Evidence

### What the session read or measured

- An opus agent (`NimbusProofs`) wrote the assessment
  `local://assess-nimbus-proofs.md`. It read the clone and built nothing.
- The lead checked the central claim by hand.
  `nimbus_verified_proxy/engine/receipts.nim:53-57` fetches the block's receipts
  with `eth_getBlockReceipts` and returns the error "downloaded receipts do not
  evaluate to the receipts root of the block" unless
  `orderedTrieRoot(...) == header.receiptsRoot`.
- Headers: `engine/blocks.nim:105-128,148-173` resolves a block tag to a header
  that the light client verified, and recomputes the block hash.
  `convHeader` exposes `stateRoot` and `receiptsRoot` (`blocks.nim:32-58`).
- Storage: `engine/accounts.nim:26-125,197-244` verifies the account proof
  against `stateRoot` and the storage proof against `storageRoot`.
- Completeness: `engine/rpc_frontend.nim:383-391` returns `eth_getBlockReceipts`
  only after the receipt-trie check.
- A limit: `eth_getLogs` over a range verifies each log it returns, but cannot
  detect a log it omits (`engine/receipts.nim:132-173`).
- Targets: a native binary; `libverifproxy.a` with C and Go bindings; a WASM
  build (`Makefile:346-398`). The `Makefile` has no iOS or Android target
  (`Makefile:66-73`). The library takes the transports from the host app
  (`library/verifproxy.nim:203-327`).
- Portal: the history network validates headers, bodies and receipts
  (`portal/network/history/history_validation.nim:17-110`). The tree has no
  state network, so Portal cannot serve storage proofs.
- The README marks the proxy "Stability: experimental".
- The armada-nitro fixturenet runs a consensus chain: "real geth + lighthouse"
  (`armada-nitro` `docs/ENVIRONMENT.md:7`), with `sigp/lighthouse:v8.2.2`
  (`ENVIRONMENT.md:29`).

### Operator rulings, verbatim

- 2026-10-19 — "you should have all the inforation you need to determine how
  difficult it would be to the full proofs armada needs with nimbus-eth. have an
  opus subagent do this work."
- 2026-09-25T03:45Z (session `01a06ee5`) — "we need to prove everything via
  fixturenet in laconic so before we worry about putting it on sepolia."

## 3. Findings and decisions

1. **The verified proxy covers needs 1 to 4.** It anchors headers to the beacon
   light client. It verifies storage with EIP-1186 proofs. It proves a log by
   rebuilding the block's whole receipt trie, which also proves completeness for
   that block.
2. **The T1 pipeline in the docs is not needed for proofs.** The docs plan a
   nimbus state-diff emitter, an IPLD mapper and a `watcher-ts` proof feed
   (`engineering/T1-ingestion.md:10-48`). The proxy replaces that path. The
   cerc-io pipeline (`plugeth-statediff`, `ipld-eth-db`, `ipld-eth-server`) and
   the statediff v4 receipt-trie code do not port.
3. **Note sync uses `eth_getBlockReceipts`, not `eth_getLogs`.** The client
   fetches the receipts of each pool-touching block and filters locally. A range
   query cannot show that no log is missing.
4. **`watcher-ts` becomes an optional, unverified cache on servers.** Its event
   "proof" (`{blockHash, receiptCID, log}`) has no link to `receiptsRoot` and is
   not used. A client re-verifies anything it gets from the cache through the
   proxy.
5. **Need 5 is not in the proxy.** The commitment part of a snapshot verifies
   through need 2 (`rootHistory`). Ciphertexts and nullifiers need signatures
   from independent watchers (the T2.4 federation role).
6. **What runs where.**
   - Server: the native proxy, with a self-hosted execution node. Historical
     `eth_getProof` needs an archive node.
   - Browser: the WASM proxy, over untrusted beacon REST and execution RPC URLs.
   - Mobile: `libverifproxy.a`, cross-compiled for iOS and Android. That is new
     work, because no upstream target exists.
   - Portal can replace the execution node for needs 1, 3 and 4, not for need 2.
7. **Difficulty**, in the docs' linear points:

   | Item | Work | Replaces | Points |
   |---|---|---|---|
   | WI-A | Operate the native proxy on servers | T1.0, T1.1 | 2 |
   | WI-B | Browser WASM build in the JS client | client tier | 3 |
   | WI-C | iOS and Android build of `libverifproxy.a` | near T6.5 | 5 |
   | WI-D | Armada verification library: slot keys, receipt scan | T1.2, T6.1 | 3 |
   | WI-E | Execution data node behind the proxy (archive) | T1.0 | 3 |
   | WI-F | Beacon light-client source for each chain | T1.0 | 4 |
   | WI-G | Snapshot cache and watcher attestation | T2.4 | 5 |
   | WI-H | `watcher-ts` as an optional server cache | T1.2 | 2 |

   The ingestion tier falls from 15 points (T1.0 8, T1.1 4, T1.2 3) to about 12
   (WI-A, WI-D, WI-E, WI-F). It is no longer a long pole. The cost moves to the
   client tier: WI-B 3 and WI-C 5.
8. **Correction to the assessment: the first proof runs on the fixturenet.** The
   assessment says a local fixturenet has no consensus chain and proposes
   Sepolia first. The armada-nitro fixturenet runs lighthouse v8.2.2. The
   operator's 2026-09-25 ruling puts the fixturenet before Sepolia. The first
   proof:
   1. On the fixturenet, shield into the pool and record the `Shield` log's
      block.
   2. Point the patched proxy (decision 9) at lighthouse's beacon REST API and at
      geth's RPC, with a trusted block root from lighthouse.
   3. Call `eth_getBlockReceipts(block)` through the proxy and find the `Shield`
      log. This proves inclusion and completeness.
   4. Compute the `rootHistory[tree][root]` slot key for the new root. Call
      `eth_getStorageAt` through the proxy and expect a non-zero value.
9. **The proxy needs a custom-network patch for the fixturenet.** The execution
   chain id is not the obstacle: a fixturenet can use any chain id (operator,
   2026-10-19), and `library/verifproxy.nim:46-52` only maps a name to a chain
   id. The obstacle is the consensus genesis. `engine/engine.nim:140-142` loads
   `getMetadataForNetwork(networkName)` and `genesisParamsForNetwork(networkName)`.
   `engine/genesis_params.nim:48-70` hard-codes `genesisTime`,
   `genesisValidatorsRoot`, `genesisBlockRoot` and the EIP-2935 fork time for
   `sepolia` and `hoodi`; any other name gets the mainnet values. `--network`
   lists only `mainnet`, `hoodi`, `sepolia` and OP Stack networks
   (`nimbus_verified_proxy_conf.nim:44-48`). Lighthouse on the fixturenet makes
   its own beacon genesis, and the light client checks sync-committee signatures
   against it. So the proxy needs a patch that reads the genesis parameters and
   the runtime config (fork versions and epochs) from the fixturenet's testnet
   directory. [INFERENCE] 1 to 2 points: the patch touches `genesis_params.nim`
   and the metadata load, and nimbus-eth2 already loads custom network
   directories for its own testnets.

## 4. The check for each decision

| Decision | Check | Result |
|---|---|---|
| 1, 3 | `engine/receipts.nim:53-57` | confirmed by the lead |
| 1 | `engine/accounts.nim`, `engine/blocks.nim` | read by the agent; not run |
| 6 | `Makefile:66-73` | no mobile target; confirmed by the agent |
| 8 | `armada-nitro` `docs/ENVIRONMENT.md:7,29` | lighthouse present; proxy not yet run against it |
| 9 | `engine/engine.nim:140-142`, `engine/genesis_params.nim:48-70`, `nimbus_verified_proxy_conf.nim:44-48` | read by the lead; patch not written |

## 5. Rejected options

- **The nimbus state-diff emitter with IPLD and `watcher-ts` proofs (T1.0, T1.1
  as written).** Rejected for the proof path. The proxy does the same
  verification upstream.
- **Porting the statediff v4 receipt-trie code.** Rejected. nimbus already
  rebuilds and checks the receipt trie.
- **`eth_getLogs` ranges for note sync.** Rejected. Omission is not detected.
- **Sepolia for the first proof.** Rejected under the 2026-09-25 ruling. The
  fixturenet has a consensus chain.

## 6. Open questions that only a build or a run can answer

- Does the metadata loader in this build (`getMetadataForNetwork`) accept a
  network directory path, or must the decision-9 patch add one? Does lighthouse
  v8.2.2 serve the light-client API by default, or only with a flag?
- Does `libverifproxy.a` cross-compile and link for iOS and Android?
- What are the WASM bundle size and the per-call latency in a browser?
- From a fresh trusted block root, how far back can the proxy prove? The block
  walk stops at `maxBlockWalk` (1000) and `maxWindowJumps` (500 jumps of 8191
  blocks) (`library/verifproxy.nim:149-150`).
- How large is an archive node for the target chain?
- How many receipt bytes does a pool-touching block carry on `demo3`?

## 7. Addendum, 2026-10-20: the open questions, resolved

The text above stays as written. Builds ran on kelce (Linux x64) and marks
(macOS arm64, Xcode 26.4). Work dirs are under `armada-laconic/.scratch/`:
`qbuild-native`, `QWasm`, `QAndroid`, `qanswers`; on marks,
`~/code/nimbus-eth1-ios`. All builds use nimbus-eth1 `4df52ef`.

### Operator rulings, verbatim

- 2026-10-20 — "resolve the open questions."
- 2026-10-20 — "it probably does support other chain ids but fixturenets mean
  you can use any chainid you need to."
- 2026-10-20 — "remember portal network from the EF as well."

### Pinned sources

| Source | Commit or version |
|---|---|
| nimbus-eth1 (all builds) | `4df52efd2ffda8df24e5d7360d464e2aa2a67808` (2026-10-19) |
| nimbus-eth1 `master` checked for the `convHeader` fix | `2c8eabf13ad9896e870ca5e990e185b920f91125` (2026-10-20): still unfixed |
| nimbus-eth2 (vendored) | `963756c91c19b7824cea57e7aab5857ce2c776c7` (v26.10.1) |
| lighthouse | tag `v8.2.2`, commit `e423a66763bb1bd780492d635123f208d80c3538` |
| reth (`rpc_server.rs`, `constants.rs`) | `main` at `f88813adb966324ecfc530a2f78df38ed0df5750` (read 2026-10-20) |
| emsdk | 3.1.73; emsdk repo at `35659964bc2762fffec0d257997fb61475f765b7` |
| Android NDK | r27c |
| armada-nitro | `origin/integrated` `55939ef87d9c` |
| armada-deployments | `ship-armada.github.com/main` `bac36395e78e` (demo3) |
| armada-poc | `origin/main` `102f17b8f8d4` |

The two-line `convHeader` patch exists only in the scratch tree
`.scratch/qbuild-native/src` (`git diff`: `nimbus_verified_proxy/engine/blocks.nim`,
2 insertions).

### Answers

1. **Custom network: the loader already accepts a directory.**
   `getMetadataForNetwork` (nimbus-eth2 `963756c`,
   `beacon_chain/networking/network_metadata.nim:355-372`) loads `config.yaml`
   from a directory for any name that is not a preset. The fixturenet starts
   lighthouse with `--testnet-dir /data/metadata` (`armada-nitro`
   `stack/compose/docker-compose-armada-fixturenet.yml:111`). Decision 9's patch
   therefore shrinks to two hard-coded tables: `genesisParamsForNetwork`
   (`engine/genesis_params.nim:48-70`) and the chain-id map
   (`library/verifproxy.nim:46-52`). For a directory, both can be read from the
   loaded metadata.
2. **Lighthouse serves the light-client API by default.** v8.2.2 has only a
   `--disable-light-client-server` flag (`beacon_node/src/config.rs:207-209,
   1456`). The fixturenet does not pass it and serves HTTP on port 5052. No
   fixturenet change is needed.
3. **The proxy does not verify Sepolia today. This is a defect in nimbus
   upstream.** Sepolia headers now carry `blockAccessListHash` (EIP-7928) and
   `slotNumber` (EIP-7843). `convHeader` (`engine/blocks.nim:32-58`) does not
   copy them, so the recomputed block hash never matches. Two independent
   providers both failed with "the downloaded anchor header doesn't match the
   anchor hash". `nim-eth` `headers.nim:42-43` and `nim-web3`
   `eth_api_types.nim:98-99` already define both fields. A two-line patch that
   copies them fixed it: `eth_blockNumber` and `eth_getBlockReceipts` then
   verified on Sepolia. The fixturenet runs every fork through Osaka, not
   Glamsterdam, so it does not hit this defect [INFERENCE: from
   `armada-nitro` `docs/ENVIRONMENT.md:7`].
4. **Native build: yes.** `make -j16 update` took 746 s;
   `make -j16 nimbus_verified_proxy` took 205 s. The binary is 215 MB with
   debug info and 51 MB stripped. `libverifproxy.a` (x86-64) is 334 MB.
5. **Measured on Sepolia** (patched native proxy; beacon
   `lodestar-sepolia.chainsafe.io`; execution `sepolia.gateway.tenderly.co` and
   `ethereum-sepolia-rpc.publicnode.com`):
   - The light-client bootstrap verified about 0.7 s after start.
   - `eth_blockNumber`: 0.23 s.
   - `eth_getBlockReceipts`: 0.29 s for a 462 KB verified receipt set.
   - `eth_getStorageAt` verified at the head: slot 0 of the demo3 pool returned
     `0x…29347919508ef4458aa1caa47e96dc16e4f5e8d6`, which is the demo3
     `shieldModule`.
   - Most storage reads failed: "distance to target block exceeds maximum proof
     window". The free providers serve `eth_getProof` only at their own tip. The
     proxy anchors a few blocks behind it. So the proxy needs its own execution
     node with a proof window.
6. **WASM build: yes.** emsdk 3.1.73 (pinned in the repo's CI);
   `make nimbus_verified_proxy_wasm` took 176 s. `verifproxy_wasm.wasm` is
   14,358,847 bytes (3,893,999 gzipped); `verifproxy_wasm.js` 279,013 bytes;
   `verifproxy.js` 7,365 bytes.
7. **WASM in a browser: partly resolved.** The build uses pthreads, so the page
   must be cross-origin isolated (COOP and COEP headers). Without them,
   `SharedArrayBuffer` is undefined. With them, headless Chromium loaded the
   module, and it fetched the light-client bootstrap in 0.54 s. `init` then
   never resolved, and the page sent no further requests. The per-call latency
   in a browser is not measured. The cause is not known.
8. **iOS: yes**, device and simulator, built on marks. `libverifproxy.a` is
   64.8 MB for `arm64-apple-ios` and 65.6 MB for the simulator.
   `nimbus_verified_proxy/tests/test_api.c` links into executables with
   `LC_BUILD_VERSION` platform 2 (iOS) and platform 7 (iOS Simulator), `minos
   15.0`. No source patch was needed. The nim command needs
   `-d:disable_libbacktrace` and explicit `--path:` flags for the vendored
   packages.
9. **Android aarch64: yes, with one flag.** The first build failed in the
   hashtree SHA-256 assembly. Nim defines `linux` for `--os:android`, so
   `ssz_serialization/digest.nim:55-61` selects hashtree with clang. With
   `-d:PREFER_HASHTREE_SHA256=false`, `libverifproxy.a` builds (150,850,796
   bytes, LLVM bitcode members because of LTO). `test_api.c` links with
   `aarch64-linux-android24-clang++ -flto -fuse-ld=lld -static-libstdc++` into
   a 69 MB `ELF 64-bit … ARM aarch64 … interpreter /system/bin/linker64`. NDK
   r27c.
10. **Historical reach covers demo3.** The proxy jumps back through the
    EIP-2935 history contract: at most `maxWindowJumps` (500) windows of 8,191
    blocks, 4,095,500 blocks (`engine/blocks.nim:62-74`). The Sepolia head
    (11,951,609) is 363,501 blocks after the demo3 deployment (11,588,108), or 45
    windows. Each jump needs a storage proof at an older block, so it needs an
    archive provider.
11. **Archive providers.**
    - geth: v1.17 serves historical `eth_getProof` with
      `--history.trienode=N`; v1.16 does not. Path-based archive is about 2 TB
      for mainnet, about 6.5 TB with historical trie data
      (geth.ethereum.org/docs/fundamentals/archive, edited 2026-02-24).
    - reth: `--rpc.eth-proof-window` defaults to 0 and is capped at
      `MAX_ETH_PROOF_WINDOW` = 28·24·60·60/2 = 1,209,600 blocks
      (`crates/rpc/rpc-server-types/src/constants.rs:62,69`). Archive 2.31 TB
      at mainnet block 24,396,823 (reth.rs/run/storage).
    - nimbus-eth1 execution client: no. `eth_getProof` returns "Historical data
      not available" below its base (`execution_chain/rpc/server_api.nim:111`).
    - Erigon and Sepolia sizes: not checked.
12. **Receipt bytes on demo3** (all 343 pool logs, block 11,588,108 to
    11,952,017, via `sepolia.gateway.tenderly.co`; data in
    `.scratch/qanswers/receipts/`): 197 pool-touching blocks. Receipt JSON per
    block: min 186,791, median 358,250, p95 600,882, max 4,274,379 bytes. Total
    79,026,844 bytes for receipts against 618,530 bytes of pool logs, a ratio of
    about 128 to 1.
13. **Snapshot size on demo3.** About 375 commitments (242 in topic
    `0x3a5b9dc2…`, which by its counts is `Shield`, and 133 in `Transact`
    `0x56a618cd…`) and 120 nullifiers (`Nullified` `0x781745c5…`). Their event
    data, ciphertexts included, is 198,208 bytes. Method: decode the array length
    of each event's dynamic argument.
14. **The EF Portal network.** nimbus' Portal client (Fluffy) now carries only
    block bodies and receipts, keyed by block number
    (`portal/network/history/content/content_keys.nim:15-17`). The host supplies
    trusted headers through a `getHeader` callback
    (`history_network.nim:33-53`). The tree has no state network. So Portal can
    replace an archive node for old receipts (needs 3 and 4), with headers from
    the proxy. It cannot serve storage proofs (need 2). A third-party crate
    states that the EF's Rust client trin "is unmaintained as of 2026"
    (github.com/willow-network/eth-historian, read 2026-10-20). That claim is not
    checked.

### What the answers change

- **The snapshot is the default sync path, not an option.** A full verified
  scan of demo3 costs about 79 MB of receipts. The snapshot data is about
  200 KB. A client checks the snapshot's commitments against `rootHistory` at a
  recent block. `rootHistory` and `nullifiers` entries are only ever set, never
  cleared [INFERENCE: from the mapping semantics in `inv-InvPool.md` §2], so a
  proof at a recent block covers them. Only snapshot builders, the watchers,
  need old receipts and old state: an archive node or Portal.
- **The decision-9 patch is smaller than estimated.** It touches two tables.
- **Mobile is buildable.** WI-C (5 points) drops to the integration work. The
  libraries already build for iOS and Android.
- **Two upstream defects in nimbus-eth1** surfaced: the missing header fields
  in `convHeader`, and hashtree assembly selected for Android with clang.
  Neither is filed.
