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
