# Unified client stack and verified note sync

Lab book entry. Session 2026-10-19 22:53 UTC to 23:36 UTC.
Session id `01a14aa3-1d15-74fd-9f8d-df18ee172022`.
Repos read: `armada-relayer`, `armada-client`, `watcher-ts`, `ponder` (cerc-io
fork), `mobymask-v2-watcher-ts`, `ipld-eth-server`, `ipld-eth-db`,
`ipld-eth-statedb`, `ipld-eth-state-snapshot`, `plugeth-statediff`,
`eth-statediff-service`, `ipfs-ethdb`, and upstream `status-im/nimbus-eth1`.

A sibling entry, `2026-10-19-docs-versus-code.md`, records the docs audit that
led to these questions.

## 1. The question

The docs audit found that the client tier and the read path in the docs assume a
Laconic substrate that the code does not use. The session asked three things.
What does the code use to read the pool today? What stack should the client use
on every platform? Can a client in a browser verify every note it reads, with
peers serving the data?

## 2. Evidence

### What the session read or measured

- **Today's read path.** `armada-relayer` `relayer-v2/watcher/` is a Ponder
  indexer (`ponder` `0.16.8`) with Postgres 16 under Docker Compose. It serves
  `/v2/quick-sync`, which `armada-sdk` and `armada-interface` use. Ponder has one
  operator and no proofs (`local://inv-InvClientOps.md` §2, §5).
- **The cerc-io Ponder fork.** Branch `laconic-esm` of `cerc-io/ponder`
  (`41acffe`, 2023-10-19) adds indexer, watcher and standalone modes
  (`Ponder.ts:175-434`) and pays for queries with Nitro vouchers over
  `@cerc-io/peer` (`a246447`, `dc5093f`, `99ae6c3`, `34f839e`). It adds no
  proofs: its diff of `packages/core/src` has no line that matches proof,
  `getStorageAt`, merkle or IPLD. It is based on Ponder `0.0.91`.
- **The earlier unification decision.** Session `01a06ee5`, 2026-09-24 19:52 UTC
  to 2026-09-25 03:16 UTC, chose one TypeScript codebase for wallet, relayer and
  watchtower on Linux, iOS and Android. It chose nodejs-mobile (option B) and
  `ts-nitro` over Waku. `armada-client` `docs/SPEC.md` (`origin/main` `78a8fdc`)
  records D1 (Node 18.20.4 everywhere), D6 (`ts-nitro` with a Waku transport),
  D7 (a Waku light node) and D8 (LevelDB through `abstract-level`). It has no
  browser target and does not mention indexing. `armada-client` `main` has no
  `src/`.
- **`watcher-ts` platforms** (`18ca4e1a`, 2025-08-13).
  `packages/util/src/database.ts:98` asserts `config.type === 'postgres'`. The
  indexer packages depend on `typeorm` 0.2.37, `pg`, `express` and
  `apollo-server-express`. History: `47ce52c1` "Sqlite based index for ERC20
  data (#31)", then `3aaa9dd7` "Use postgres instead of sqlite. (#38)". Only
  `packages/peer` targets browsers. No React Native or mobile code exists.
- **nimbus-eth1 upstream** (`master`, read 2026-10-19). The `Makefile` has a
  `libverifproxy` C static library (with Go bindings) and a
  `nimbus_verified_proxy_wasm` target built with `-d:emscripten`. It sets linker
  flags for darwin, windows and linux, with no Android or iOS target. The README
  marks the verified proxy "Stability: experimental". The proxy checks RPC
  answers with `eth_getProof` against a beacon light client.
- **MobyMask lineage.** `../mobymask` does not exist on this machine.
  `mobymask-v2-watcher-ts` (`f1db07c`) runs a libp2p peer, a Nitro node and a
  Mokka consensus engine (`src/server.ts:46-58`). It pays its upstream per
  `eth_getStorageAt` with Nitro vouchers (`server.ts:66-85`). `watcher-ts`
  attaches to each event `proof: {blockHash, receiptCID, log: {cid, ipldBlock}}`
  (`util/src/indexer.ts:706-715`). No code in `watcher-ts` verifies a proof. The
  only check is the test `assertProofData`, which checks that the fields are
  present. `verifyCheckpointData` compares a checkpoint with the watcher's own
  database.
- **`ipld-eth-server` audit** (`local://audit-ipld-eth-server.md`; server
  `330bc3d7`):
  - `getStorageAt` returns only the storage leaf node. `eth.storage_cids` has no
    path column.
  - `eth_getProof` is a real EIP-1186 proof. `pkg/eth/api.go:891` defines
    `GetProof`, and lines 968 and 981 call `trie.Prove`. The trie reader gets any
    node by CID from `ipld.blocks`. `plugeth-statediff` writes the intermediate
    state and storage trie nodes there (its README: "including intermediate state
    and storage trie nodes").
  - Logs and receipts are flat IPLD objects. No receipt trie is built or stored
    (`plugeth-statediff` `indexer/ipld/eth_parser.go:70-93`). No proof connects a
    single log to `receiptsRoot`.
  - The server can assemble every receipt of a block in transaction order as raw
    RLP (`pkg/eth/sql.go:71-109`). No single documented call exposes that list.
  - The raw header RLP is available through GraphQL `allEthHeaderCids {
    blockByCid { data } }`.
  - The pipeline is geth-only (`plugeth` `v1.13.14-cerc-2`). No repo of the nine
    mentions nimbus.
  - `ipld-eth-state-snapshot` writes a complete state and storage trie at one
    block height.
  - Feature work stopped in August 2024. `ipld-eth-server`'s last commits
    (2025-11-05) are CI pins.
  - `watcher-ts` `ipld-eth-client/src/eth-queries.ts:94,111` queries
    `blockByMhKey`. The server renamed it to `blockByCid` in `6d748715`
    (2023-04-14), so the two are out of step.

### Operator rulings, verbatim

- 2026-10-19T23:00:53Z — we should just ignore the existing ponder stuff and port
  everything to a unified stack that runs on servers, browsers, and mobile.
- 2026-10-19T23:04:19Z — wasn't this unification already discussed?
- 2026-10-19T23:10:25Z — ts-watcher is what we'll use for indexing and it should
  already support all 4 target platforms.
- 2026-10-19T23:15:47Z — part of the issue is nimbus on mobile, but obviously not
  the browser.
- 2026-10-19T23:18:48Z — review how ../mobymask works and suggest how we could
  use the go-watcher ts-watcher stuff to do browers-to-browser full verification
  of notes and everything.
- 2026-10-19T23:27:19Z — you can snapshot to get the size fixed. that can be
  supplimental data provided via the watcher network that isn't on chain because
  it's a cache.
- 2026-10-19T23:36:40Z — make the as lab book entries for the moment.

## 3. Findings and decisions

1. **One stack on four targets.** The client stack runs on servers, browsers,
   iOS and Android (ruling 23:00:53Z). This extends the 2026-09-25 decision in
   `armada-client` `docs/SPEC.md` with a browser target. It does not replace
   that decision.
2. **No Ponder.** Neither the relayer's Ponder indexer nor the cerc-io Ponder
   fork is part of the plan (ruling 23:00:53Z).
3. **`watcher-ts` indexes.** It is the indexer (ruling 23:10:25Z). The operator
   expected it to run on all four targets. The code at `18ca4e1a` runs the
   indexer only on a Node server with Postgres. So the indexer runs on servers,
   and the other targets verify and serve, but do not index, until `watcher-ts`
   gets a storage backend other than Postgres.
4. **Anchors.** A client trusts only headers that a beacon light client checks.
   Servers run nimbus. Mobile uses the nimbus verified proxy through
   `libverifproxy`, which nobody has yet built for Android or iOS. Browsers use
   its WASM build (ruling 23:15:47Z, refined by the upstream `Makefile`).
5. **What a client can verify from `ipld-eth-server` data.**
   - Storage values: yes, with `eth_getProof`, not `getStorageAt`.
   - A single log: no. No receipt-trie proof exists.
   - Completeness of a block's logs: yes, by rebuilding the receipts trie from
     the block's full receipt list and comparing with `receiptsRoot`. This also
     proves that each log is included.
6. **Verified note sync design** (proposal, ruling 23:18:48Z):
   - Commitments: the client rebuilds the commitment tree and checks the root
     against `rootHistory[tree][root]` with an `eth_getProof` storage proof. A
     missing, extra or changed leaf changes the root.
   - Events and nullifiers: the client checks each block that touches the pool by
     rebuilding its receipts trie.
   - Privacy: every client fetches the same full stream and trial-decrypts
     locally. No request names a user's note or nullifier.
   - Distribution: verified data is content-addressed. Any peer, a browser
     included, can serve it, and the receiver verifies it again. Servers are paid
     with Nitro vouchers (the MobyMask `X-Payment` pattern).
7. **Snapshots bound the size** (ruling 23:27:19Z). The watcher network serves a
   snapshot at block N as a cache, not chain data. A new client verifies the
   snapshot, then syncs from N to head with full proofs. The commitment part of a
   snapshot is fully verifiable against `rootHistory`. Ciphertexts and the
   nullifier set are not: the tree root does not cover ciphertexts, and the
   nullifier set has no on-chain root. A bad ciphertext or a missing nullifier
   causes delay or a hidden note, not theft. Independent watchers sign the
   snapshot CID, and a client accepts a CID that several of them sign.
   `watcher-ts` state checkpoints (IPLD, by CID) are the mechanism to reuse.

## 4. The check for each decision

| Decision | Check | Result |
|---|---|---|
| 3 | `watcher-ts` `util/src/database.ts:98`; package dependencies | Postgres-only indexer, confirmed |
| 4 | nimbus-eth1 `Makefile` targets `libverifproxy`, `nimbus_verified_proxy_wasm` | present upstream; not built or run here |
| 5 | Audit `local://audit-ipld-eth-server.md`; lead spot check of `api.go:891,968,981` and `eth-queries.ts:94,111` | confirmed |
| 6, 7 | None yet | design only; no code |

## 5. Rejected options

- **The relayer's Ponder indexer as the read path.** Rejected: one operator, no
  proofs (ruling 23:00:53Z).
- **Porting the cerc-io Ponder fork to Ponder 0.16.** Rejected (ruling
  23:00:53Z). It spread reads across paid peers but added no proofs.
- **Per-log proofs from `ipld-eth-server`.** Not possible: no receipt trie is
  stored (decision 5).
- **Single-slot nullifier lookups by the client.** Rejected: a lookup names the
  user's nullifier to the server.

## 6. Open items

- **The proof source.** The pipeline (`plugeth-statediff` to `ipld-eth-db` to
  `ipld-eth-server`) is geth-only and dormant since August 2024. ADR-0009 retires
  plugeth for nimbus-eth1. A nimbus-fed source is T1.0, the long pole.
- **API gaps.** The server needs a documented call for a block's full receipt
  list. `watcher-ts` must move from `blockByMhKey` to `blockByCid`.
- **A verifier library.** None exists. It must run in browsers, Node and mobile.
- **The signer set** for snapshot attestation. This is the T2.4 federation role.
- **Size.** Measure on `demo3` the receipt bytes per pool-touching block and the
  snapshot size.
- **`watcher-ts` beyond servers.** Decide whether to add a non-Postgres backend,
  or to keep the indexer server-only.
