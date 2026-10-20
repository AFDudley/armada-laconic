# The nimbus pair in the fixturenet, and the first proof through the proxy

Lab book entry. Session 2026-10-20, about 04:20 to 08:05 UTC.
Session id `01a14aa3-1d15-74fd-9f8d-df18ee172022`.
Repos changed: `armada-nitro` (branch `nimbus-fixturenet`, not merged) and
`AFDudley/nimbus-eth1` (branch `armada/verified-proxy-fixes`, a new fork of
`status-im/nimbus-eth1`). The lead husbanded three Sonnet agents (`NimbusPair`,
`NimbusPair2`, `ProxyPatches`). The running log is
`armada-laconic/.scratch/plan/husbandry-nimbus-proxy.md`. The agents' reports are
under `.scratch/{NimbusPair,NimbusPair2,ProxyPatches}/REPORT.md`.

This entry closes the first open question of
`2026-10-20-proxy-and-archive-open-questions.md` §6.

## 1. The question

Can a non-validating nimbus pair (nimbus-eth1 and nimbus-eth2) follow the
armada-nitro fixturenet? And can the nimbus verified proxy prove one Armada pool
event there, against its own light client?

## 2. Pinned sources

| Source | Commit, tag or digest |
|---|---|
| armada-nitro base | `integrated` `f234fad3` |
| armada-nitro result | `nimbus-fixturenet` `fd25bbe9` (pushed to origin) |
| nimbus-eth1 upstream base | `master` `2c8eabf13ad9896e870ca5e990e185b920f91125` |
| nimbus-eth1 result | `AFDudley/nimbus-eth1` `armada/verified-proxy-fixes` `b0205d1db9acb7a04f0aec0e53ad0115b605224e` |
| nimbus-eth1 image | `statusim/nimbus-eth1:linux-amd64-v0.4.3@sha256:2689fe3746e19b017b5402ed590c17a4a857a456f4a2c2e5f51ae013a32dd19d` |
| nimbus-eth2 image | `statusim/nimbus-eth2:amd64-v26.10.1@sha256:c052f011b8f005c18826046e8f3413482284ebd4bf9568b61d4dc548a6141442` |
| lighthouse | `sigp/lighthouse:v8.2.2`, commit `e423a66763bb1bd780492d635123f208d80c3538` |
| geth | `ethereum/client-go:v1.17.6` (from `docs/ENVIRONMENT.md`) |

## 3. Operator rulings, verbatim

- 2026-10-20 — "we should obvioulsy have nimbus run in the fixturenet as well?
  that's already documented?"
- 2026-10-20 — "make the code fix." Then, asked which fix: "Both".
- 2026-10-20 — "husband these agents similar to but not exaclty like
  ../exophial husband and make sure the work continues to completion even if it
  times out due to the fixturenet slowness."
- 2026-10-20 — on `/tmp/dbg.json` filling `/tmp`: "Delete it".
- 2026-10-20 — "yes. write the lab book"

## 4. Evidence and findings

### 4.1 The nimbus pair: four commits on `nimbus-fixturenet`

| Commit | Change |
|---|---|
| `df6dc1d9` | Adds nimbus-eth2 (no validators) and nimbus-eth1 to the fixturenet compose file; runs both as root because their volume directories are root-owned; adds the new volumes to the RAM-disk setup; uses `/ip4` multiaddrs because nimbus-eth2 does not resolve `/dns4`; sets `--sync-light-client=false` |
| `0cc9fb35` | Starts nimbus-eth2 with `trustedNodeSync` from lighthouse's REST API |
| `f22338e0` | Escapes `$TNS_LOG` and `$CKPT_SLOT` as `$$` in the compose command |
| `fd25bbe9` | Peers nimbus-eth1 with geth over devp2p with a fixed node key and enode |

The causes:
- At genesis, nimbus-eth2 sends a Status message with `finalized_epoch=1` where 0
  is expected. Lighthouse answers "Different finalized chain" and bans the peer
  once its own finalized epoch passes 1. This is what the first agent saw as
  intermittent "Stream Reset". The lead's first hypothesis, `--target-peers 0`,
  was not the cause.
- After a checkpoint sync, nimbus-eth1 has no peer to fetch the blocks it lacks.
  The engine API does not carry them.

Checked by the lead on the fresh deployment `NimbusPair2` (cluster
`laconic-38bb3cec1850f7ae`, built from the committed branch):
- nimbus-eth1 and geth both report block 1864 with hash `0xd6a96f7c4a9819df…`.
- nimbus-eth2 and lighthouse both report finalized root `0x0ddbd3b3036dd193…` at
  slot 1792.
- The agent reports that `armada-env-verify` (pool-e2e: shield, transfer and
  unshield with real proofs) passed with RC=0. The lead did not rerun it.

### 4.2 The proxy: two commits on `armada/verified-proxy-fixes`

| Commit | Change |
|---|---|
| `827f95ff` | `convHeader` copies `blockAccessListHash` (EIP-7928) and `slotNumber` (EIP-7843): 2 lines in `nimbus_verified_proxy/engine/blocks.nim` |
| `b0205d1d` | Loads genesis parameters and chain id from a custom network directory |

No third commit was needed in nimbus.

### 4.3 Why the proxy rejected every update from lighthouse

- Lighthouse v8.2.2 serves, as the best update for sync-committee period 0, the
  first signed update: attested slot 11, finalized slot 0. It keeps serving it
  after finality, after a restart, and at epochs 4, 12 and 29.
- `consensus/types/src/light_client/light_client_update.rs:402-463`, at
  `e423a66`, ends `is_better_light_client_update` with "Tiebreaker 2: Prefer
  older data". Checked by the lead. Once that update is the best, no later
  update replaces it within the period.
- The proxy, and stock nimbus-eth2 v26.10.1 in its own light-client sync, reject
  that update at the finality-branch check (`light_client_sync.nim:109-120`).
- nimbus-eth2 serves a different best update for period 0: attested slot 141,
  finalized slot 64, which is real finality. Checked by the lead with
  `/eth/v1/beacon/light_client/updates?start_period=0&count=1`.
- Not settled: which client follows the spec for an update whose finalized
  header is still at genesis. The agent calls it a lighthouse defect; that is
  its inference.

### 4.4 The first proof through the proxy

Run by hand by the lead on `NimbusPair2`, 2026-10-20 08:00 UTC:
- Proxy `b0205d1d`, `--network=<testnet dir copied from the nimbus-cl
  container>`, `--trusted-block-root` from nimbus-eth2's finalized header,
  beacon API nimbus-eth2, execution API nimbus-eth1.
- The light-client bootstrap verified. `eth_blockNumber` answered `0x75b`.
- `eth_getBlockReceipts(886)`, which rebuilds the block's receipt trie against
  the anchored `receiptsRoot`, returned the pool's `Shield` log in 0.65 s.
  Pool `0x4458acb1185ad869f982d51b5b0b87e23767a3a9`, tx
  `0x4e154bc4b66b815f76a96335204d34691633b69dcae4361d71b74ecf54ebfd10`.
- `eth_getStorageAt(pool, 0x10003363518423059521f839b17dc2cd424ffca6581ce68d336e98a9c710fc12, latest)`
  returned `0x…01`. That slot is `rootHistory[0][0x2d3853e8fd27ce53c14c81ce73a3ad894789d2bb4fd37b6744da71d9bd6db1d4]`
  (base slot 50; the root is `merkleRoot()` read from nimbus-eth1).

Lighthouse and geth produce the chain, and no answer above is trusted from
either of them.

### 4.5 Host incidents

- `/tmp` (24 GB tmpfs) filled twice from `/tmp/dbg.json`, a forge debug trace of
  armada-nitro's `ArmadaNitroAdjudicatorTest`. It was 24 GB at 04:xx UTC (the
  `NimbusPair` agent deleted it) and 15.6 GB at 06:06 UTC (the lead deleted it,
  operator-approved). No process held it open. Its writer is not known. A full
  `/tmp` broke every container healthcheck on the host.
- Three other fixturenet deployments ran on the host throughout. No agent
  touched them.

## 5. Decisions

1. The fixturenet runs a nimbus pair beside geth and lighthouse.
2. On the fixturenet, the proxy's light-client source is nimbus-eth2, not
   lighthouse, while the period-0 update question is open.
3. The proxy fixes live on `AFDudley/nimbus-eth1` `armada/verified-proxy-fixes`.
   They are not filed upstream.

## 6. The check for each decision

| Item | Check | Result |
|---|---|---|
| Nimbus pair follows the chain | block hash and finalized root comparison, by the lead | equal |
| pool-e2e on the new stack | `armada-env-verify` | agent reports RC=0; not rerun |
| Proof through the proxy | §4.4, by the lead | passed |
| Lighthouse tiebreak | `light_client_update.rs:402-463` at `e423a66` | read |

## 7. Open items

- `nimbus-fixturenet` (`fd25bbe9`) is not merged. It needs armada-nitro's own
  integration path; a husband does not merge onto `integrated` by hand.
- The genesis-finality update: which client departs from the spec? File the
  result with that client.
- File the two proxy defects upstream (`convHeader` fields; no custom network),
  after the operator approves the issue bodies.
- The writer of `/tmp/dbg.json`.
- The deployment `~/deployments/NimbusPair2` still runs.

## 8. Addendum, 2026-10-20 (after commit `6c18941`)

The text above stays as written.

### Operator rulings, verbatim

- 2026-10-20 — "we never need to use or worry about sepolia."
- 2026-10-20 — "have a subagent figure out which client is off spec. don't file
  anything upstream until we're done here."
- 2026-10-20 — on merging `nimbus-fixturenet`: "1. yes, of course."
- 2026-10-20 — "3. of course update the docs."
- 2026-10-20 — "you sould tear it down." (the `NimbusPair2` deployment)

### Lighthouse v8.2.2 departs from the spec; nimbus-eth2 conforms

A Sonnet agent (`SpecCheck`) settled §4.3's open question. Its report is
`.scratch/SpecCheck/REPORT.md`; its saved data is in the same directory. The
lead checked the three central points:
- The spec. consensus-specs tag `v1.7.0-beta.4`
  (`aa16bb4c156184e9548a997d53efaf2a228a6304`),
  `specs/altair/light-client/full-node.md` `create_light_client_update`: the
  server sets `update.finalized_header` only `if finalized_block.message.slot
  != GENESIS_SLOT`. Otherwise it asserts
  `attested_state.finalized_checkpoint.root == Bytes32()` and leaves the header
  empty. `sync-protocol.md` `validate_light_client_update` mirrors this on the
  verifier side (agent: lines 437-453).
- Lighthouse. `consensus/types/src/light_client/light_client_update.rs:286-298`
  at `e423a66763bb1bd780492d635123f208d80c3538` builds the Fulu
  `finalized_header` from any finalized block that has the same fork. It has no
  genesis-slot check. The agent reports the same pattern for every fork back to
  Altair.
- The data. At slot 11, the attested state's `finalized_checkpoint` is epoch 0
  with a zero root (`nb_state_slot11.json`). Lighthouse's period-0 update
  carries a real header: slot 0, state root `0x52cb65c1…`
  (`lh_update_p0.json`).

The agent also reports that nimbus-eth2's own update (finalized slot 64)
matches the slot-141 state's finalized root (`0x3defce2a75…`), and that
nimbus-eth2 (`light_client_sync.nim:96-120` at `963756c`) rejects lighthouse's
update with a check equivalent to the spec's.

Consequence: the defect is in lighthouse, at genesis-finalized updates, on any
chain that runs light-client-era forks from genesis. Together with lighthouse's
"prefer older data" tiebreak, it pins a bad best update for the whole of
period 0. Nothing is filed upstream, by order.

### The merge

`exophial apply nimbus-fixturenet` merged the branch into armada-nitro
`integrated` as `990c9fa9d3526a459bfb374e10cfb7cf0c5b48b0`, pebble `arn-b58`,
2026-10-20 08:12 UTC. Before the apply, the lead rebased the branch onto
`integrated` `d064a536`, with no conflicts; the four commits became `412ed6f9`,
`87fe5880`, `f479b2ea` and `990c9fa9`. The lead checked the tree:
`origin/integrated` contains `990c9fa9`, the compose file has the nimbus
services, and `docs/ENVIRONMENT.md:30` names the pair. `origin/main` was still
`d064a536` at that check; the repo's fast-forward workflow moves it.

### Teardown

`laconic-so deployment --dir ~/deployments/NimbusPair2 stop --delete-volumes`
removed the deployment (cluster `laconic-38bb3cec1850f7ae`), after `SpecCheck`
had saved its data. The lead deleted the deployment directory.

### Sepolia

Sepolia is no longer a target or a test network. The Sepolia measurements in
`2026-10-19-nimbus-verified-proxy-proofs.md` §7 (timings, demo3 receipt and
snapshot sizes, historical reach) stay as recorded data. They do not drive
sizing. A fixturenet or mainnet measurement replaces them.

### Open items now

- The lighthouse defect, the two proxy defects and the Android hashtree
  selection: file upstream only when the operator says the work here is done.
- The docs update runs on branch `docs-proof-path` (agent `DocsProofPath`).
- The writer of `/tmp/dbg.json`.
- armada-nitro `docs/ENVIRONMENT.md` still orders "fixturenet, then Sepolia".

## 9. Addendum, 2026-10-20 (after commit `ae47f57`)

The text above stays as written.

### Operator rulings, verbatim

- "3. just say sapolia is currently out of scope."
- "4. don't upstream anything just leave a note in the lab books."
- "8 it was another worker, ignore it." (the writer of `/tmp/dbg.json`)

### Upstream defects: recorded here, not filed

By order, nothing goes upstream. This list is the record.

| Defect | Upstream | Evidence | Local state |
|---|---|---|---|
| `convHeader` drops `blockAccessListHash` and `slotNumber`, so header hashes never match on a chain that sets them | status-im/nimbus-eth1, `nimbus_verified_proxy/engine/blocks.nim:32-58` at `2c8eabf` | `2026-10-19-nimbus-verified-proxy-proofs.md` addendum | Fixed on github.com/AFDudley/nimbus-eth1 `armada/verified-proxy-fixes`, commit `827f95ff` |
| The proxy accepts no custom network (its genesis-parameter loader knows only built-in networks such as `sepolia` and `hoodi`) | status-im/nimbus-eth1, verified proxy | §3 of this entry | Fixed on the same branch, commit `b0205d1d` |
| `ssz_serialization/digest.nim:55-61` selects the hashtree assembly for `--os:android` with clang, and the Android link fails | status-im/nim-ssz-serialization | `2026-10-19-nimbus-verified-proxy-proofs.md` §7 | Worked around with `-d:PREFER_HASHTREE_SHA256=false`; no patch |
| Lighthouse fills `finalized_header` in genesis-finalized light-client updates (§8) | sigp/lighthouse | §8 of this entry | Already known upstream: sigp/lighthouse#9865, "Fix: use empty finalized header when finalized_block_root is zero", open since 2026-08-19, mergeable, waiting for maintainers to run CI (read 2026-10-20). Not in v8.2.2. No filing needed |

### Merges and cleanup

- armada-laconic `858da68`: the docs now describe the proof path as proven by hand on the fixturenet, and the build plan has no Sepolia stage work. The lead removed the docs agent's new clause D9.4 from ADR-0009, because decisions go into lab book entries.
- armada-nitro `6d9d46e3` (`exophial apply`, pebble `arn-d0a`): `docs/ENVIRONMENT.md` says Sepolia is currently out of scope.
- The lead removed the scratch builds (`QWasm`, `QAndroid`, `SWasmBrowser/src`, `SFixturenetProof`, the lighthouse source copies) and marks `~/code/nimbus-eth1-ios`. The agent reports stay. The lead stopped `~/deployments/SFixturenetProof` with `--delete-volumes`, unmounted the 3 GB tmpfs of it and of `NimbusPair2`, and deleted both directories. The armada-nitro branch `nimbus-fixturenet` is deleted locally and on `origin`.
- The WASM stall test page and report stay at `.scratch/SWasmBrowser/{REPORT.md,site2}` for the stall issue.
