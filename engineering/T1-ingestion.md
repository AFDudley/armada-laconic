# T1 · Ingestion

**Status:** draft · tier doc · 2026-09-04
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T1
**Release:** v1 (T1.0, T1.1, T1.2)
**Depends on / Blocks:** consumes T0.0 (pool commitment/nullifier ABI + storage layout), T0.2 (adjudicator events), T0.3 (deposit/payout events) · internal chain T1.0 → T1.1 → T1.2 · blocks T2.0 (serves feeds over this ingest), T2.1 (metering rides the ingested stream); head-cursor consumed by T6.3.

T1 is the mechanism that gets L1 state into the watcher parties, and it is the **v1 long pole** — the single tier in the v1 spine that is genuinely unbuilt rather than integration work. Everything above it (T2 serve/meter, T4 venue, T6 wallet) assumes a fresh, proof-carrying view of the Railgun pool (T0.0) and the Nitro settlement rail (T0.2/T0.3), and this tier produces that view. A **nimbus-eth1 state-diff source** (T1.0) emits proof-carrying Ethereum state diffs; those diffs are **encoded as IPLD** (T1.1); and a **watcher-ts ingest config** (T1.2) points the resulting pipeline at *our* contracts. The tier is ingest-only: the proof-carrying GraphQL feeds, voucher metering, and P2P substrate that *serve* this state to clients are T2 (see T2.0). This doc stops at "state is indexed and queryable" — it does not publish or meter anything. A leak or a lag here collapses the read-side anonymity guarantee the whole system rests on (overview §7 state-feed staleness), which is why the ingestion source is called out as the real risk.

```
  L1 canonical head (Railgun pool T0.0 · deposit/payout T0.3 · adjudicator T0.2)
        │
        ▼
  ┌──────────────────────────────┐
  │ T1.0  nimbus-eth1 emitter     │  stateless/witness + Aristo
  │  changed MPT nodes for the    │  → {nodes, block, stateRoot}
  │  watched-address set only     │
  └──────────────────────────────┘
        │  proof-carrying state diffs
        ▼
  ┌──────────────────────────────┐
  │ T1.1  IPLD encode             │  ipld-eth-server / -state-snapshot
  │  content-address every node   │  leaf → … → state root walkable
  └──────────────────────────────┘
        │  authenticated node paths
        ▼
  ┌──────────────────────────────┐
  │ T1.2  watcher-ts ingest cfg   │  our ABIs + storage layout
  │  index commitments/nullifiers │  head cursor per contract
  │  + Deposit/Payout + adj events│
  └──────────────────────────────┘
        │  indexed, proof-complete state  (ingest boundary)
        ▼
     T2.0 serve · T2.1 meter   ·   staleness cursor → T6.3
```

**What we build vs. reuse.**

| Piece | Item | Status | Source |
|---|---|---|---|
| nimbus-eth1 stateless/witness + Aristo primitives | T1.0 | **reuse** | `status-im/nimbus-eth1` |
| State-diff **emitter module** (watched-set scoping, per-block publish) | T1.0 | **build** — net-new, server-side | this tier |
| Ethereum trie-node IPLD codecs + index/serve skeleton | T1.1 | **reuse** | `ipld-eth-server` / `ipld-eth-state-snapshot` |
| nimbus-eth1 → IPLD **mapper** + subtrie backfill | T1.1 | **build** | this tier |
| `watcher-ts` framework (indexer, IPLD state store, block loop) | T1.2 | **reuse** | `cerc-io/watcher-ts` |
| Pool/adjudicator **ingest config** (ABIs, storage layout, cursor) | T1.2 | **build** — net-new | this tier |

The reused halves are production or upstream-real; the built halves — the emitter, the mapper, the config — are what make T1 the v1 long pole. Only T1.2's config is *fully* net-new; T1.0 and T1.1 are `partial`, because the primitives exist but the wiring to our working set does not.

## T1.0 State-diff source (nimbus-eth1)

**What it is.** T1.0 is the upstream that turns canonical L1 blocks into a stream of **proof-carrying Ethereum state diffs** — the intermediate and leaf Merkle-Patricia trie nodes that changed in each block — for exactly the contracts we watch. Its registry status is **partial**: the primitives exist upstream, but the emitter module wired to our working set does not.

**Why a bespoke state source (reuse-vs-build).** We do not need a full execution-layer archive node behind the watchers. The watcher only ever answers reads about a bounded set of slots — the Railgun pool's commitment/nullifier tree (T0.0) and the deposit/payout plus adjudicator storage (T0.2/T0.3) — so the requirement is narrow: keep just enough local state to compute the slices we consume, and discard the rest. A general JSON-RPC provider is both too much and too little. It is too much because it is a full world-state archive, and too little because a bare `eth_getProof` against a provider re-introduces the exact RPC-fingerprinting we are trying to kill and gives no self-contained diff stream. The historical cerc-io ingestion lineage here is [`ipld-eth-server`](https://github.com/cerc-io/ipld-eth-server) plus [`ipld-eth-state-snapshot`](https://github.com/cerc-io/ipld-eth-state-snapshot), which indexed and served IPLD state from a `plugeth`-instrumented geth. That plugeth-style state feed is wildly out of date and is not the direction; it is cited only as prior art for the IPLD indexing layer T1.1 inherits.

**nimbus-eth1 as the target.** The direction is [`nimbus-eth1`](https://github.com/status-im/nimbus-eth1), the Nim L1 execution client, because its **stateless/witness execution and Aristo state DB** primitives already produce canonical, proof-carrying MPT trie-node diffs from a bounded working set rather than a full state trie. Two properties matter for us:

- **Witness / stateless execution** (`execution_chain/stateless`) generates and verifies the witness — the exact set of trie nodes touched by a block — so a diff is self-authenticating against the block state root without holding the whole trie.
- **Aristo state DB** (`execution_chain/db/aristo`) is the canonical MPT-node store that lets us persist and re-serve those nodes, and lets a scoped instance cache only the watched-contract subtrie while dropping everything else. This is what makes a mobile-capable emitter conceivable later — a phone could carry the working set for just our pool — though the mobile-scoped emitter is a stretch goal, not a v1 dependency.

**Effort honesty.** This is the long pole and it is partly unbuilt. The upstream primitives are real, but a **state-diff emitter module** wired to nimbus-eth1 — subscribing to canonical head, extracting the changed trie nodes for our watched addresses, and publishing them downstream — does not exist, and it is a server-side net-new build. The registry marks it `partial` for that reason: the direction and the primitives are settled, but the emitter that turns them into a T1.1 diff stream is the work. The dominant risk is **staleness**. A lagging or stalled emitter means the watcher's view falls behind, and a missed nullifier or a missed adjudicator `Challenge` is a safety problem, not just a latency one — which is why T1.2 exposes a head cursor and T6.3 gates on it.

**Interface exposed.** T1.0's output is a per-block set of `{changed MPT nodes, block number, state root}` for the watched address set, consumed by T1.1 for IPLD encoding. Its input is the watched-address set, supplied by the T1.2 config.

**Key tasks.**

- Pin `nimbus-eth1` at a build commit; stand up a synced instance against the target L1.
- Build the **state-diff emitter module**: subscribe to canonical head, pull the block witness, and extract the changed intermediate+leaf trie nodes scoped to the watched-address set.
- Emit `{nodes, block, stateRoot}` downstream to T1.1 on each block; expose the watched-address set as config input from T1.2.
- Prove the bounded working set: assert Aristo holds only the watched subtrie and reorgs re-emit the corrected nodes.
- (Stretch, not v1-blocking) a scoped **mobile emitter** caching only our pool's subtrie.

## T1.1 IPLD proof-carrying state diffs

**What it is.** T1.1 is the encoding layer. Each state diff from T1.0 is content-addressed and stored as **IPLD**: every trie node — the intermediate branch and extension nodes, and the leaf that holds the actual slot value — becomes an addressable block, so the path from a contract's storage slot up to the block **state root** is a walkable, self-verifying chain of IPLD links. Its registry status is **partial**: `ipld-eth-server` and `ipld-eth-state-snapshot` already implement the IPLD codecs and the index/serve shape; what is partial is feeding them from nimbus-eth1 diffs rather than the retired plugeth path.

**Reuse-vs-build.** The IPLD codecs for Ethereum trie nodes and the indexing/serving skeleton are reused from the cerc-io ipld-eth lineage — [`ipld-eth-server`](https://github.com/cerc-io/ipld-eth-server) on the query/serve side and [`ipld-eth-state-snapshot`](https://github.com/cerc-io/ipld-eth-state-snapshot) for bulk state-snapshot ingest during cold-start and backfill of the watched subtrie. The build is the **mapper from nimbus-eth1's Aristo/witness node output into those IPLD codecs**, plus the backfill path: snapshot the watched-contract subtrie once, then apply per-block diffs.

**Why proof-carrying matters.** Because every slice is stored as the actual trie nodes down to the leaf, a downstream reader can reconstruct a **Merkle/storage proof** for any slot against the block state root. This is precisely the property that lets T2.0 answer a read as `getStorageAt → {value, proof}`, where the `proof` inherits on-chain verifiability: a phone verifies the value against the L1 state root and never has to trust the watcher. Without the intermediate nodes there is no proof to carry, so T1.1's job is to make sure the diff stream retains the full node path, not just leaf values.

**What T2.0 needs from it (the interface).** T1.1 exposes, per watched slot and block, the set of IPLD blocks constituting the proof path — the leaf plus its ancestor nodes to the state root — along with the resolved `value`. T2.0's proof-carrying feed binds to this: it selects the current-head node set for a queried slot and packages `{value, proof}` from these IPLD blocks. T1.1 does not decide feed schema, metering, or transport; those are T2. It only guarantees that for any watched slot at any ingested block, the full authenticated node path is retrievable. It consumes T1.0's raw diff set and produces this addressable, proof-complete store.

**Key tasks.**

- Vendor/pin `ipld-eth-server` + `ipld-eth-state-snapshot`; reuse their Ethereum trie-node IPLD codecs and index/serve skeleton.
- Write the **nimbus-eth1 → IPLD mapper**: map Aristo/witness node output onto the IPLD codecs, content-addressing every node.
- Backfill path: snapshot the watched-contract subtrie once (state-snapshot), then apply T1.0 per-block diffs.
- Guarantee node-path completeness: for any watched slot at any ingested block, leaf→state-root ancestors are all retrievable.

## T1.2 Watcher ingest config

**What it is.** T1.2 is the net-new piece of this tier and the only fully net-new item in T1: the [`watcher-ts`](https://github.com/cerc-io/watcher-ts) ingest configuration that points the generic framework at **our** contracts. The `watcher-ts` framework — the indexer, the IPLD-backed state store, and the block-processing loop — is production code, reused wholesale; the build is the pool/adjudicator-specific config: the ABIs, the storage layout, and the choice of which events and slots to track. The complementary serve-side config — the GraphQL `{value, proof}` feed schema — lives in T2.0, so this section covers only what `watcher-ts` *ingests and indexes*.

**Reuse-vs-build.** We reuse `watcher-ts` (its `indexer.ts` block loop and `util/src/types.ts` state types) consuming the T1.1 IPLD store; we build the contract descriptors below.

**What it indexes.** The config tracks exactly the state the settlement spine reads:

- **T0.0 — Railgun shielded pool.** Commitment-insert and nullifier-spend events, plus the commitment Merkle-tree and nullifier-set storage slots, from the redeployed pool's ABI and storage layout. This is the note stream clients scan locally: a missed commitment means a missed note, and a missed nullifier means a double-spend blind spot.
- **T0.3 — deposit/payout contract.** The `Deposit(channelId, asset, amount, consumedNullifier)` and `Payout(channelId, newCommitments[])` events and their storage — the Nitro↔Railgun boundary where value enters and leaves channels.
- **T0.2 — Nitro adjudicator.** The `Challenge`, `Checkpoint`, and `Finalized` events and the relevant `NitroAdjudicator` storage, so a watchtower (T6.3) can see a stale-state challenge land.

The config supplies T1.0 its **watched-address set** — these three contracts — and consumes the T1.1 IPLD store to resolve their slots block-by-block, closing the tier loop.

**Head-cursor / staleness detection.** The config exposes, per watched contract, the **current chain-head cursor**: the block number, and for the pool the commitment-tree index, that the watcher has ingested up to. This is the tier's liveness signal. Because the T1.0 emitter is the long pole, a consumer must be able to tell that the cursor has stopped advancing and refuse to act on a stale view. **T6.3's self-hosted watchtower gates its challenge-response on this cursor being fresh** — it will not sleep-walk into missing a `Challenge` because the emitter lagged. T2.0 also republishes this cursor on its feeds, but the cursor itself is produced here at ingest.

**Testing / risks (inline).** Correctness is validated on the laconic fixturenet: shield, deposit, or challenge a **known** value, then assert that the indexed slot resolves to the expected `value` and that the T1.1 proof path verifies against the block state root — a tampered node must fail verification. Staleness is tested directly: stop or lag the T1.0 emitter, assert the head cursor stops advancing and a consumer detects it, then resume and assert gap-fill catches up. This is the guard on the tier's headline risk. The **identical-bytes invariant** is an ingest-adjacent concern surfaced at serve time (T2.0), but it originates here: the config must index deterministically so that every subscriber resolves byte-identical node sets for a given slot at a given block. Reuse `watcher-ts`'s framework tests as-is, and add tests only for our config, the contract descriptors, and cursor/staleness behavior.

**Key tasks.**

- Author the `watcher-ts` contract descriptors: pool commitment/nullifier events + slots (T0.0), deposit/payout events + storage (T0.3), adjudicator `Challenge`/`Checkpoint`/`Finalized` (T0.2).
- Point the indexer at the T1.1 IPLD store; resolve tracked slots block-by-block over the ingested node paths.
- Emit the per-contract **head cursor** (block number + pool commitment-tree index) as the tier's liveness signal for T6.3.
- Add fixturenet tests for feed correctness, staleness detection, and the deterministic/identical-bytes ingest invariant.

## Open questions (genuine unknowns)

- **Emitter maturity.** nimbus-eth1's stateless/witness path is under active upstream development. Whether its witness output is stable and complete enough to drive a production emitter today — versus needing upstream contribution — is unresolved, and is the core reason T1.0 is `partial` rather than `built`.
- **Reorg semantics through IPLD.** How deep a reorg the emitter and IPLD store must tolerate before re-emitting corrected node paths, and how T1.2 rolls the head cursor back on reorg, needs to be pinned against the target L1's finality.
- **Backfill cost.** The sizing and time of the one-time watched-subtrie snapshot (via `ipld-eth-state-snapshot`) for the Railgun pool's commitment tree at genesis-of-deployment is unmeasured.
- **Mobile-scoped emitter.** Whether an Aristo working set for just our pool is small enough to run on-device is the open question behind the stretch mobile emitter; v1 does not depend on it.

## Sources

- **nimbus-eth1** — Nim L1 execution client (stateless/witness execution + Aristo state DB); the state-diff source direction. https://github.com/status-im/nimbus-eth1 — `execution_chain/stateless` (witness generation/verification), `execution_chain/db/aristo` (canonical MPT-node store).
- **ipld-eth-server** — prior cerc-io IPLD ingestion/serve lineage; IPLD codecs + index/serve skeleton reused by T1.1. https://github.com/cerc-io/ipld-eth-server (@ `330bc3d`).
- **ipld-eth-state-snapshot** — bulk watched-subtrie snapshot for cold-start/backfill under T1.1. https://github.com/cerc-io/ipld-eth-state-snapshot (@ `9e483fc`).
- **watcher-ts** — generic proof-carrying indexer framework; T1.2 is its pool/adjudicator ingest config. https://github.com/cerc-io/watcher-ts (@ `18ca4e1`) — `indexer.ts`, `util/src/types.ts`.
- Companion site docs: [`architecture.html`](../architecture.html) (T1 · Ingestion — nimbus-eth1 → IPLD → watcher), [`build-plan.html`](../build-plan.html) (T1 target + net-new emitter module). Contract ids/status/release per the [`05-building-block-view.md`](./05-building-block-view.md) registry.
