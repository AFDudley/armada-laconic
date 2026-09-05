# T2 · Watcher-party substrate

**Status:** draft · tier doc · 2026-09-04
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T2
**Release:** v1 (T2.0, T2.1, T2.2, T2.3) · v2 (T2.4)
**Depends on / Blocks:** consumes T1.2 (ingested pool/rail state) and the event ABIs of T0.0 / T0.2 / T0.3; T2.4 bonds against T0.4 and routes its fraud path through T0.5 / T3. Blocks T4 (venue solver consumes these feeds + metering) and T6 — the feed schema and head cursor gate the T6.3 watchtower, and the metering interface drives the T6.2 settlement client.

A **watcher party** is the Service tier that turns ingested L1 state into **proof-carrying, everyone-gets-the-same-bytes** feeds of our shielded pool (T0.0) and the settlement rail (T0.2 / T0.3), **metered privately over Nitro vouchers**. It is a **bonded federation of peers, not a client-server RPC**: the overwhelming majority of the network is browser wallets and mobile apps talking peer-to-peer, with only a handful of state-diff emitters (T1) and STUN/TURN as fixed infra ([architecture T2](../architecture.html)). The party never takes custody and holds no game state; it **serves and meters data**. Its correctness rests on two properties — *feed authenticity*, enforced by the storage proof the client checks, and *voucher accounting*, enforced by go-nitro's channel machinery — not on consensus. Everything the anonymity-set guarantee depends on flows from this: clients scan and decrypt notes **locally**, so no RPC provider ever fingerprints a user against the pool ([mobile-privacy §4 ★](../mobile-privacy.html)). Ingestion itself is **not owned here** — T1.2 delivers the proof-carrying state that this tier consumes as a feed.

```
   L1 (pool T0.0 + rail T0.2/T0.3 events)
        │  proof-carrying state diffs (via T1.2 ingest)
        ▼
  ┌─────────────────────────────────────────────┐
  │  WATCHER PARTY  (bonded peer federation)     │  Service tier
  │  watcher-ts config: our pool + adjudicator   │
  │  ├─ T2.0 GraphQL feed: getStorageAt→{value,proof}
  │  ├─ T2.1 payments.ts: Nitro voucher metering │
  │  └─ T2.4 (v2) threshold-Schnorr attestation  │
  └─────────────────────────────────────────────┘
        ▲  Waku pub/sub (discovery/gossip)  │  libp2p-noise (1:1 hot loop)   ← T2.3
        │  ── T2.2 peer substrate ──         ▼
   browser / mobile / server peers  ── scan + verify LOCALLY ──►  venue (T4), wallet (T6)
```

### Watcher vs. watchtower vs. keeper vs. relay *(kept distinct — glossary)*

These names collide across the source repos. This tier owns exactly the first and must not be conflated with the others:

| Term | Role | Owner |
|---|---|---|
| **Watcher** (Laconic) | proof-carrying, metered **state indexer** — serve verifiable query results / note-streams. **Read-and-serve.** | **this tier (T2)** |
| **Watchtower** (state channel) | watches L1 for a stale-state challenge and **responds with your latest signed state** (`checkpoint`/`respond`). = a watcher (read) **+ a bonded responder (write)**. | T6.3 (composes T2 feed + T0.2 responder) |
| **Keeper / broadcaster** | submits transactions on the user's behalf (mobymask lineage); write-side origin privacy. | T6 (T6.5 transport / wallet) |
| **Relay** (libp2p) | pure network transport / NAT traversal (circuit-relay, STUN/TURN). No trust, no on-chain role. | T2.3 |

**Duality:** the watcher protects the **read side** of pool privacy (no RPC fingerprints as you scan the pool), while the keeper/broadcaster protects the **write side** (no EOA links as you submit) — same anonymity set, opposite ends. A watchtower is this watcher's read feed plus T0.2's bonded challenge responder; T6.3 composes the two. This tier ships only the **read-and-serve** half.

---

## T2.0 Proof-carrying feeds

**Status:** reuse+config · **Release:** v1.

**What it is.** The `watcher-ts` framework pointed at **our** contracts, serving each read as `getStorageAt → {value, proof}` (`packages/util/src/types.ts`). The watcher returns the relevant state slice **with a storage/Merkle proof**, so a phone verifies authenticity without trusting the watcher. Every subscriber receives **identical bytes** for a given slice — the anti-fingerprint property that stops a data provider from learning which commitments/nullifiers a user scans (mobile-privacy §4 ★).

**Reuse vs. build.** The framework — proof-carrying GraphQL, `indexer.ts`, and storage-proof serving — is production code and is **reused** as-is at `watcher-ts` @ `18ca4e1`. The net-new work is the **pool/adjudicator-specific config**: the ABIs, storage layout, and GraphQL query schema for our pool's commitment-insert / nullifier-spend events (T0.0) and for the deposit/payout contract's `Deposit`/`Payout` plus the adjudicator's `Challenge`/`Checkpoint`/`Finalized` events (T0.2 / T0.3).

**Interface consumed.** State comes from **T1.2**, not a bare JSON-RPC node. T1.2 delivers a proof-carrying state feed (intermediate + leaf MPT nodes), so the watcher's `{value, proof}` responses inherit on-chain verifiability. This tier consumes that feed; it does not build ingestion.

**Interface exposed (to T4 / T6).** Consumers MUST bind to this, not re-derive it:
- **Feed schema (GraphQL).** Proof-carrying queries over pool and rail events: commitment-insert / nullifier-spend (T0.0); `Deposit(channelId, asset, amount, consumedNullifier)`, `Payout(channelId, newCommitments[])`, and the adjudicator's `Challenge`/`Checkpoint`/`Finalized` (T0.2 / T0.3). Every response carries a `{value, proof}` the client verifies locally.
- **Head cursor / liveness.** Each feed exposes its current chain-head cursor (the block / tree-index it has ingested up to) so a consumer can detect staleness and gap-fill. This cursor is what **gates the T6.3 watchtower**: its challenge-response fires only while the feed is fresh.

**Key tasks.** Author the watcher config for our pool (ingest commitment-insert / nullifier-spend; expose the `getStorageAt → {value, proof}` GraphQL feed); extend it to ingest the rail events from T0.2 / T0.3 and add their feed queries; publish the schema and head cursor for T4 / T6.

**Observers this tier blinds.** The feeds exist to deny one specific observer (mobile-privacy §1, §4 ★):

| Observer | Would learn without this tier | Blinded by |
|---|---|---|
| RPC / data provider | Exactly which commitments/nullifiers/addresses the wallet queries — enough to fingerprint the user against the pool | proof-carrying, identical-bytes feed + **local** note-scan |
| The watcher itself | Nothing forgeable — a lying `value` fails the storage proof the client checks | `getStorageAt → {value, proof}` verification |
| Metering observer | Who is paying for which reads | private per-request Nitro vouchers, funded at note creation (T2.1) |

A leak here collapses the whole anonymity set, which makes this the **do-first** item on the spine's critical path for T4 / T6.

**Testing / risks (inline).** Shield, deposit, or challenge a known value on the fixturenet, then assert that the feed's `{value, proof}` matches the on-chain slice and that the storage proof **verifies** against the block state root; a tampered `value` MUST fail proof verification client-side. **Identical-bytes invariant:** the anonymity guarantee assumes every subscriber to a slice receives byte-identical responses, so a config that lets query parameters vary the served bytes per requester silently re-introduces fingerprinting — assert this in tests, not just at review. State-feed staleness is the known long pole, but it lives in T1 (see T1.0): a lying watcher is caught by the proof, whereas a *withholding* watcher is a liveness problem mitigated by the T2.4 federation + bond.

---

## T2.1 Nitro voucher metering

**Status:** reuse+config · **Release:** v1.

**What it is.** Private, per-request metering of the paid feed methods. Paid RPC methods (`constants.ts`) are metered by **go-nitro payment vouchers** (`payments/vouchers.go`) via watcher-ts `payments.ts`: a peer pays per-request vouchers that net into a co-signed **virtual-channel** allocation (`protocols/virtualfund/virtualfund.go`), settled to L1 in USDC. Metering funds are provisioned **at note creation** and stay **private** — the voucher stream is the same mechanism T0.2 / T0.3 publishes, reused here, so watcher payment leaks nothing about the reader.

**Reuse vs. build.** go-nitro vouchers plus virtual channels and watcher-ts `payments.ts` are **reused**. The build is the **metering wiring**: register the paid feed methods over `constants.ts`/`payments.ts`, fund a virtual channel, and meter per-request vouchers against the T2.0 feed methods.

**Interface exposed (to T4 / T6).** The **voucher-metering interface** — the set of paid RPC/feed methods, the go-nitro voucher format (reused from T0.2), and the virtual-channel handshake that funds metering. T6.2's settlement client drives this, and T4's venue solver consumes the same feeds under the same metering. This is the same **Nitro clearing** the T4 venue uses for fees; T2 is where it first appears on the read path, which is why watcher metering is one of the two revenue pillars (venue spread + watcher metering).

**Key tasks.** Register paid methods; fund the virtual channel; meter per-request vouchers via `payments.ts`; run the fixturenet E2E: a peer subscribes → receives `{value, proof}` → verifies locally → pays a voucher → the voucher nets into the channel allocation → cooperative close settles to the correct USDC balance.

**Testing / risks (inline).** For N metered requests, the voucher total must equal the expected amount and net into a co-signed allocation, while an unpaid or underpaid request is refused. **Voucher replay / double-count:** accounting MUST reject a replayed or out-of-order voucher; correctness rides on go-nitro's voucher nonce/redemption logic (`payments/vouchers.go`), so treat any deviation in our wiring as safety-critical for the metering channel balance. **Metering griefing:** unpaid reads are cheaply refused, and anti-spam is a nominal request fee rather than a heavy protocol — because metering is non-custodial, the worst case is refused service.

---

## T2.2 P2P peer substrate

**Status:** reuse · **Release:** v1.

**What it is.** The peer stack that lets browser, mobile, and server peers interoperate on one libp2p base — **NOT client-server**. Peers run the in-browser/mobile Nitro node ([`ts-nitro`](https://github.com/cerc-io/ts-nitro) @ `884d616`) over `@cerc-io/peer` — **circuit-relay + webrtc-star + gossipsub** — so a phone, a browser tab, and a headless server are all first-class peers. A **submit-on-behalf** relay follows the [`mobymask`](https://github.com/cerc-io/mobymask) @ `2329198` keeper lineage (`packages/server/index.ts`, `submitInvocations → registry.invoke`), reused here so a light peer can delegate a metered T2.0 request without its own public origin. The write-path keeper/broadcaster in T6 reuses the identical lineage — the same pattern at the opposite end of the anonymity set.

**Reuse vs. build.** This is **entirely reuse** — `ts-nitro` over `@cerc-io/peer` plus the mobymask relay lineage, with no protocol changes. The work is deploy and configure: bring up peers both **in a browser** and **on a server**.

**Interface exposed.** The **submit-on-behalf endpoint** — the keeper-relay method for delegating a metered request from a light peer (T6's broadcaster/keeper reuses the identical lineage).

**Key tasks.** Vendor/pin `ts-nitro` (`884d616`) and `mobymask` (`2329198`) at their commits; bring up the peer substrate over `@cerc-io/peer`; run peers both in a browser and on a server; add the submit-on-behalf relay for delegated metered requests.

**Testing / risks (inline).** The same feed is served and consumed by a headless server peer *and* an in-browser peer over `@cerc-io/peer`, and the T2.0 identical-bytes property holds across both. **Mobile transport maturity** is the real risk — real libp2p on React Native is the crux, tracked in T6.5; the peer config here is Med difficulty, while the transport under it (T2.3) is High.

---

## T2.3 Transport

**Status:** reuse · **Release:** v1. Transport gate settled in ADR-0008 — **both, per need**.

**What it is.** Two transports, each for what it is good at:
- **Waku pub/sub** for feed **discovery, gossip, and async** delivery. Subscribers pull by content-topic, giving recipient-unlinkability bounded by the topic's anonymity set, and it is the transport Railgun's broadcaster network already speaks ([go-waku](https://github.com/waku-org/go-waku); [waku-broadcaster-client](https://github.com/Railgun-Community/waku-broadcaster-client)). Mobile uses the **gomobile** native module (the `@waku/sdk` is not React-Native-compatible → T6.5).
- **libp2p-noise direct streams** for the **1:1 hot loop** (voucher exchange, low-latency reads), as in go-nitro's p2p message service (`node/engine/messageservice/p2p-message-service/service.go`, `/nitro/msg/1.0.0`) and `ts-nitro`. Noise gives encryption plus peer-id auth only — **not IP privacy** (endpoints see each other's IP); the mixnet underlay that hides IP is T6's concern, not this tier's.

A **circuit-relay + STUN/TURN** bootstrap lets browser and mobile peers behind NAT dial each other — the only fixed transport infra (architecture T2).

| Mode | Carries | Recipient-unlinkable? | Latency | Mobile path |
|---|---|---|---|---|
| **Waku pub/sub** | feed discovery, gossip, async note-streams, broadcaster interop | yes (topic anonymity set) | multi-hop store-and-forward | go-waku gomobile module |
| **libp2p-noise** | 1:1 voucher/metering hot loop, low-latency reads | no (endpoints see IPs) | tens-of-ms | ts-nitro over `@cerc-io/peer` |

**Interface exposed.** Stable Waku **content-topics** per feed for discovery and subscription (analogous to Railgun's fixed `/railgun/v2/...` topics), plus the direct-noise protocol id for the 1:1 hot loop:
```
# illustrative content-topic shape (mirrors Railgun's /railgun/v2/...):
/armada/v1/${chainId}-pool-commitments/json
/armada/v1/${chainId}-pool-nullifiers/json
/armada/v1/${chainId}-rail-adjudicator/json   # Challenge/Checkpoint/Finalized
```

**Key tasks.** Wire Waku pub/sub content-topics for feed discovery/gossip; wire direct libp2p-noise for the 1:1 hot loop; stand up the circuit-relay + STUN/TURN bootstrap.

**Reuse vs. build / flagged alternative.** Everything here is **reuse** (go-waku, Railgun's waku-broadcaster-client, go-nitro's message service). A single `@cerc-io/libp2p` gossipsub transport for everything is the **flagged alternative**, rejected as the default because Waku is mandatory for broadcaster interop and gives async recipient-unlinkability, while direct-noise is needed for the tens-of-ms Nitro hot loop that Waku's multi-hop store-and-forward cannot meet. Using **both per need** is the settled choice; a gossipsub-only stack remains a fallback if Waku operations prove too heavy on mobile.

---

## T2.4 Federation + bond + threshold DKG signing (v2)

**Status:** partial · **Release:** v2. Kept at v2 depth — never a v1 prerequisite.

**What it is.** The economic-security layer that upgrades the read-and-serve federation into a **bonded** party that can **attest** to what it serves and orders. A watcher party is a bonded federation of service providers that collectively serves one data need; at v2 it also signs everything it attests to with a **threshold Schnorr signature** produced by a distributed key.

**Threshold DKG signing.** The `chain-signatures` library provides Ethereum-compatible Schnorr (`ethschnorr.Sign` / `Verify`) and the **Distributed Schnorr Signature** protocol (Stinson–Strobl `(t, n)`) in `ethdss`, built on a `kyber` DKG. Two properties matter:
- **On-chain verifiable.** Because it is Ethereum-flavoured Schnorr, an L1 contract verifies the party's aggregate signature — so a sequencing cert or a censored-commit receipt becomes a **slashable fraud proof** against the bond.
- **Threshold direction is safety-first.** `t-of-n` tolerates `t−1` malicious for safety and `n−t` offline for liveness. We pick **t high** (e.g. **4-of-7**) so forging an attestation needs a large coalition, and a liveness failure degrades to **halt, not loss** because settlement is non-custodial. A BFT-style small quorum (e.g. 3-of-11) would be wrong — it would let any 3 forge.

The threshold key is used for **signing only**. We do **not** run a threshold-encrypted mempool: with the sequencer and key-holder being the same federation, encryption gives no real fairness, since a colluding threshold can decrypt-then-order. Fair ordering is achieved with commit-reveal in T3, not here.

**Interface consumed / relationship to siblings.**
- **Bond = T0.4.** The on-L1 registry + bond contract is owned by T0.4; this tier is a *member* that bonds against it — reference it, do not re-own it.
- **Fraud path = T0.5 / T3.** The DSS signs the sequencing certs, inclusion receipts, and liquidity-proof snapshots that T3 produces; the sequencing-cert + fraud-proof verifier is T0.5, and a censored-commit or bad-ordering receipt settles as slashing on L1 (disputes → T0.4 bond via T0.5). T2.4 provides the **signature primitive and the bonded membership**; the ordering that it attests to lives in T3.
- **What it attests over T2.0–T2.3.** Attestation makes a *withholding* watcher (the liveness gap flagged in T2.0) accountable: the bonded federation is what closes the availability hole the storage proof alone cannot.

**Reuse vs. build.** `chain-signatures` (`ethschnorr`, `ethdss`, kyber DKG) is **reused** as the primitive; the DKG setup, the party's signing wiring, and the L1 verify-and-slash integration are the v2 build — **months of work per venue**, layered on the same spine, not a rewrite (yield-clearing / architecture T2–T3).

**Testing / risks (inline, v2).** DSS and fair-ordering tests are v2, not v1. The DSS/sequencer upgrade adds a larger shared-substrate audit surface (architecture T4 note): anything that makes an aggregate signature verifiable-and-slashable on L1 is audit-critical before mainnet.

---

## Sources

- watcher-ts (proof-carrying feeds + voucher metering) — https://github.com/cerc-io/watcher-ts (`packages/util/src/types.ts` `getStorageAt→{value,proof}`, `indexer.ts`, `payments.ts`, `constants.ts`) @ `18ca4e1`
- ts-nitro (in-browser/mobile Nitro node; `@cerc-io/peer` circuit-relay + webrtc-star + gossipsub, `@chainsafe/libp2p-noise`) — https://github.com/cerc-io/ts-nitro @ `884d616`
- go-nitro (payment vouchers, virtual channels, p2p message service) — https://github.com/cerc-io/go-nitro (`payments/vouchers.go`, `protocols/virtualfund/virtualfund.go`, `node/engine/messageservice/p2p-message-service/service.go`) @ `435eb2b`
- mobymask (submit-on-behalf keeper relay) — https://github.com/cerc-io/mobymask (`packages/server/index.ts`) @ `2329198`
- go-waku (mobile Waku, gomobile bindings) — https://github.com/waku-org/go-waku
- Railgun Waku broadcaster client (fixed content-topics) — https://github.com/Railgun-Community/waku-broadcaster-client
- chain-signatures (threshold DSS: `ethschnorr`, `ethdss` Stinson–Strobl (t,n), kyber DKG) — https://git.vdb.to/cerc-io/chain-signatures @ `9016a7c`
- nimbus-eth1 (state-diff emitter / EL state source, consumed via T1.2) — https://github.com/status-im/nimbus-eth1
- Companion site docs: [`architecture.html`](../architecture.html) (T2 watcher-party substrate, threshold Schnorr DSS), [`mobile-privacy.html`](../mobile-privacy.html) (§4 read-time private sync ★), [`yield-clearing.html`](../yield-clearing.html)
