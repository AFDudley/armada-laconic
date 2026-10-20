# The proxy and archive open questions, second round

Lab book entry. Session 2026-10-20, after commit `ed40f28`.
Session id `01a14aa3-1d15-74fd-9f8d-df18ee172022`.
Four Sonnet agents (`SWasmBrowser`, `SAristoEmitter`, `SArchiveFacts`,
`SFixturenetProof`) worked the open questions of
`2026-10-19-nimbus-verified-proxy-proofs.md` §7 and
`2026-10-20-archive-state-for-nimbus.md` §6. Their reports are under
`armada-laconic/.scratch/{SWasmBrowser,AristoEmitter,SArchiveFacts,SFixturenetProof}/REPORT.md`.
The lead checked the claims marked "checked".

## 1. The questions

1. Why does the WASM proxy stall in a browser, and what is its latency?
2. Can nimbus' Aristo database emit standard trie nodes for an archive emitter?
3. Does Erigon serve historical `eth_getProof`? How large are Sepolia
   archives? Are the EF Portal clients maintained? Which event is topic
   `0x3a5b9dc2…`?
4. Can the proxy prove one pool event on the armada-nitro fixturenet?

## 2. Pinned sources

| Source | Commit or version |
|---|---|
| nimbus-eth1 | `4df52efd2ffda8df24e5d7360d464e2aa2a67808`, plus the 2-line `convHeader` patch in `.scratch/qbuild-native/src` |
| nimbus-eth2 (vendored) | `963756c91c19b7824cea57e7aab5857ce2c776c7` (v26.10.1) |
| lighthouse | `sigp/lighthouse:v8.2.2`, `v8.2.2-e423a66` |
| armada-nitro | `55939ef87d9ce8ecc1e5a8912b495031aa71531e` (`origin/integrated`) |
| armada-poc | `102f17b8f8d46bcec85ef03dfc36bd0396407687` (`origin/main`) |
| plugeth-statediff | `9a18f199e13ec41cfaeea3d9e4bf0f7968d13e88` |
| ipld-eth-server | `330bc3d7e2dd6acf4f7a1e744d39ea462ecb2323` |
| ipld-eth-statedb | `f256f13ca075045f5eb1fe27c3300070ef4b8777` |
| ipfs-ethdb | `3e6321e71090bc71c6a8cbef0b819712f45d9d19` |
| go-ethereum `v1.10.26-statediff-4.2.2-alpha` | `fda1723d02799ef541d1289cc55175ad5e61a116` |
| erigon | v3.7.1, `8c1e389` (2026-10-02) |
| trin | `master` `30aeef83a32ed30e747141137f0b89e1f7da8cbe` (2025-09-23); release v0.3.3 `4add18d` (2025-05-28) |
| ultralight | last commit `76d287f` (2025-05-21) |

## 3. Operator ruling, verbatim

- 2026-10-20 — "write the lab book entries for all this research and then
  commit it. be sure to record git commit hashes where appropriate. dispatch
  sonnet subagents to resolve the open questions after what you've figured out
  so far has been committed and pushed to the repo."

## 4. Findings

1. **WASM in a browser: still stalled; the stall is located.** `init` resolves
   in 150 to 520 ms. The first `sync()` stalls. JavaScript calls the exported
   `wasmDeliverBeaconTransport` to hand back the light-client bootstrap, the
   call returns with no exception, and the Nim `deliverBeaconTransport` body
   never runs: its first, unconditional statement never prints, although
   `strings` shows it in the binary. Three instrumented rebuilds ruled out
   timer throttling, a freed transport context (`GC_ref` added), missing
   `setupForeignThreadGc()`, microtask reentry, and worker threads.
   [INFERENCE, the agent's] A Nim error flag that is already set makes the
   generated C exit early. Latency past `init` is not measured. Bootstrap
   fetches took 398 to 573 ms.
2. **Aristo can feed an archive emitter.** Checked:
   `execution_chain/db/aristo/aristo_serialise.nim:67-75` encodes standard
   17-item branch nodes and extension nodes, and
   `aristo/aristo_proof.nim:49-58` converts any vertex to RLP nodes. The agent
   found the per-block change set in the transaction frame
   (`aristo_tx_frame.nim:165-250`) and a hook point after `processBlock`
   (`core/chain/forked_chain.nim:536-544`). The batch-import path
   (`persist_blocks.nim:124-259`) checkpoints once per batch, so it needs a
   forced per-block checkpoint. `ipld-eth-server`'s proof path keys nodes by
   CID from `(codec, keccak256(rlp))` with codecs 0x96 (state trie) and 0x98
   (storage trie) (`plugeth-statediff` `builder.go:323,444`, `shared.go:35,37`),
   so nodes written in that format need no server change. The agent estimates
   the emitter at 4 to 5 points, against 8 for T1.0 in `build-plan.md`.
3. **Receipt-trie nodes cost about as much as the receipts.** statediff v4
   rebuilt and wrote every receipt-trie node each block (`eth_parser.go`
   `processReceiptsAndLogs`; `eth_receipt_trie.go` `GetNodes()`), with no
   reuse between blocks. [INFERENCE] The storage is the same order as the
   receipt bytes: 79 MB for demo3's 197 pool-touching blocks.
4. **Erigon serves historical `eth_getProof`** with
   `--prune.include-commitment-history=true` (stable since v3.4). The default
   `--prune.commitment-history.distance` keeps all history. Mainnet archive:
   2.25 TB in use, 4 TB recommended (measured 2026-09-17).
5. **Sepolia archive sizes are not published** for geth v1.17 or reth. The
   agent checked geth.ethereum.org docs and release notes, and reth.rs.
6. **The EF Portal clients are idle.** Checked: trin's last commit is
   `30aeef8` on 2025-09-23. Its last release is v0.3.3 (2025-05-28). Ultralight's
   last commit is `76d287f` on 2025-05-21. Neither repo is archived. nimbus'
   Fluffy is the active Portal client.
7. **Topic `0x3a5b9dc2…` is `Shield`.** Checked: `cast keccak
   'Shield(uint256,uint256,(bytes32,(uint8,address,uint256),uint120)[],(bytes32[3],bytes32)[],uint256[])'`
   returns it. The `Transact` signature needs the compiled ABI, because
   `CommitmentCiphertext` has five fields there.
8. **The fixturenet proof did not complete.**
   - The agent wrote the custom-network patch: 6 files, 180 lines,
     `.scratch/SFixturenetProof/proxy-custom-network.patch`. It reads genesis
     time, genesis validators root, genesis block root and the EIP-2935 fork
     time from the testnet directory.
   - The patched proxy agreed with lighthouse on chain id, genesis and the Fulu
     fork digest. The light-client bootstrap verified.
   - Every light-client update failed check 7, the finality-branch Merkle proof
     (`light_client_sync.nim:109-120`). The proxy never learned the next sync
     committee, so every state read answered "light client doesn't know the
     current and next sync committees, sync first".
   - The pool event itself is real. pool-e2e shielded 10 USDC into pool
     `0x4458AcB1185aD869F982D51b5b0b87e23767A3A9`. The `Shield` log is in block
     772 (tx `0x0dda8b45…aab0f`). `rootHistory` (base slot 50) for the new root
     returned `0x…01`. Both were read directly from geth, not through the proxy.
9. **Two explanations for the fixturenet failure; neither is confirmed.**
   - The agent's: the vendored nimbus-eth2 light-client code for the
     Fulu-era forks is at draft spec versions (`v1.6.0-alpha.0` to
     `v1.7.0-alpha.12` in its comments), and its finalized-root gindex differs
     from lighthouse v8.2.2.
   - The lead's [INFERENCE]: the same proxy synced on Sepolia, where Fulu is
     active, which weakens a general gindex mismatch. The rejected updates
     attested slot 11 and were signed at slot 12, before the new chain had
     finalized. An update whose finalized header is still at genesis may be
     the trigger. A run after the fixturenet finalizes would separate the two.

## 5. What this changes

- The archive emitter for watchers is feasible on nimbus. It is smaller than
  the docs' T1.0 estimate.
- Erigon v3.4 and later, and geth v1.17, serve historical proofs. They are
  fallbacks for the archive role.
- Portal history depends on Fluffy. The EF's own clients are idle.
- The fixturenet first proof needs one more fix after the custom-network patch.
- The browser proxy needs a fix inside nimbus' WASM bindings.

## 6. Open questions

- Does the fixturenet proof work once the chain has finalized? If not, which
  side computes the finality branch differently?
- What sets the early exit in `deliverBeaconTransport` in the WASM build?
- Measured size of receipt-trie nodes per block, if per-log proofs are wanted.
