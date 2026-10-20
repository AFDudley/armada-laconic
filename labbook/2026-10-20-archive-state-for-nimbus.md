# Archive state for nimbus: an ipld-eth-server-like layer

Lab book entry. Session 2026-10-20, after 01:00 UTC.
Session id `01a14aa3-1d15-74fd-9f8d-df18ee172022`.
Repos read: nimbus-eth1 (`.scratch/nimbus-eth1`,
`4df52efd2ffda8df24e5d7360d464e2aa2a67808`), and the cerc-io audit in
`local://audit-ipld-eth-server.md` (`ipld-eth-server` `330bc3d7`,
`ipld-eth-statedb` `f256f13`, `ipfs-ethdb` `3e6321e`, `plugeth-statediff`
`9a18f19`).

This entry follows `2026-10-19-nimbus-verified-proxy-proofs.md` §7, which found
that the proxy's deep header jumps need storage proofs at old blocks.

## 1. The question

Does Armada need something like `ipld-eth-server` to support nimbus in archival
work?

## 2. Evidence

### What the session read or measured

- The nimbus-eth1 execution client serves no state below its base.
  `execution_chain/rpc/server_api.nim:111-112`: `if header.number <
  api.chain.baseNumber: return err("Historical data not available")`.
  `eth_getProof` uses this path (`server_api.nim:666-681`).
- The execution client has no archive-state option. `execution_chain/conf.nim`
  offers era1, era and ere directories for block history (lines 94-104),
  `history-expiry` for Portal history (line 493) and `prune` (line 535).
- `ipld-eth-server`'s `eth_getProof` (`pkg/eth/api.go:891`, `trie.Prove` at
  lines 968 and 981) reads each trie node by its hash-derived CID from
  `ipld.blocks` through `ipld-eth-statedb` `trie_by_cid` and `ipfs-ethdb`
  (`SELECT data FROM ipld.blocks WHERE key = $1`).
- `plugeth-statediff` writes every new state and storage trie node per block,
  intermediate nodes included, into `ipld.blocks` (its README: "including
  intermediate state and storage trie nodes"; `builder.go:296-327`).
- The cerc-io feed is geth-only (`plugeth` `v1.13.14-cerc-2`) and has had no
  feature work since August 2024.
- geth v1.17 serves historical `eth_getProof` with `--history.trienode=N`, at
  about 6.5 TB for mainnet (geth.ethereum.org/docs/fundamentals/archive, edited
  2026-02-24).

### Operator rulings, verbatim

- 2026-10-20 — "we need to use something like ipld-eth-server to support nimbus
  in archival work. confirm."
- 2026-10-20 — "write the lab book entries for all this research and then
  commit it. be sure to record git commit hashes where appropriate. dispatch
  sonnet subagents to resolve the open questions after what you've figured out
  so far has been committed and pushed to the repo."

## 3. Findings and decisions

1. **Confirmed for historical state.** nimbus-eth1 cannot serve state proofs at
   old blocks, so archival state needs a separate layer. Something like
   `ipld-eth-server` fits.
2. **Not needed for block history.** Headers, bodies and receipts come from era
   files and the Portal network. Receipts verify against `receiptsRoot` with no
   archive layer.
3. **The reusable part is the server, not the feed.** `ipld-eth-server`'s proof
   path reads trie nodes by hash and does not depend on geth. [INFERENCE] A
   nimbus emitter that writes the same per-block trie nodes into `ipld.blocks`
   would let that path serve historical `eth_getProof`, perhaps unchanged.
4. **The archive layer serves watchers, not clients.** Watchers need it to build
   snapshots, verify deep headers and prove old state. Clients verify recent
   state through the proxy and old receipts through Portal or era files.
5. **T1.0 and T1.1 return in a narrower role:** an archival backend for
   watchers, not the client proof path.
6. **Fallback:** geth v1.17 with `--history.trienode` already serves historical
   `eth_getProof`.

## 4. The check for each decision

| Decision | Check | Result |
|---|---|---|
| 1 | `server_api.nim:111-112`; `conf.nim` options | read by the lead |
| 3 | audit Q5 and Q7; `api.go:891,968,981` | read; emitter not built |
| 6 | geth archive docs | read; not run |

## 5. Rejected options

- **Reviving the geth statediff feed as is.** Rejected. It is geth-only and
  dormant, and ADR-0009 moves the source to nimbus-eth1.

## 6. Open questions

- Can nimbus' Aristo database emit standard RLP Merkle-Patricia trie nodes for
  each block's changes, so an emitter can fill `ipld.blocks`? This is the main
  technical question.
- Should the emitter also store receipt-trie nodes, as statediff v4 did? Per-log
  proofs need them; per-block receipt sets do not.
- Does Erigon serve historical `eth_getProof`, and how large are Sepolia
  archives for geth v1.17 and reth?
