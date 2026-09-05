# T6 · Client / apps

**Status:** draft · tier doc · 2026-09-04
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T6
**Release:** v1 (T6.0–T6.7) · opt (T6.8)
**Depends on / Blocks:** depends on T0.0 (circuit `wasm`/`zkey`, pool commitment/nullifier ABI, POI root), T0.3 (deposit/payout + channel lifecycle + app registry + vouchers), T2 (proof-carrying feeds + metering; T6.3 gates on T2 feed freshness). Blocks end-user delivery — nothing downstream.

The Client tier is the wallet and everything a real user runs on their own device. It shields, trades, settles, and **scans their own notes** end-to-end, and it does so without keys ever leaving the OS secure enclave. It is a fork of `laconic-wallet` / `laconic-wallet-web`, a React-Native app with a browser build that already ships production-grade custody (T6.0), extended with a WASM note-scanner (T6.1), a settlement client that drives T0.3 (T6.2), a phone-resident watchtower (T6.3), key-derivation policy for unlinkability (T6.4), the mobile transport crux (T6.5), Groth16 proving (T6.6), and the Armada-branded product build-up plus SDK wiring (T6.7). The wallet is a **watcher-party (P2P) client**, not a client of a server: it talks peer-to-peer to bonded watcher parties (T2) for feeds and metering and to counterparties/hubs for channel play, and it never surrenders custody — a watcher party is exited unilaterally by force-close. The load-bearing constraint is that **mobile is unbuilt**, so this doc states one concrete interim→production path, not a menu.

**What we build vs. reuse** (only three pieces are genuinely new logic; the transport is a *port*, not new protocol):

| Piece | Registry item | Status | Source |
|---|---|---|---|
| BIP-39 + HD + secure-enclave key custody | T6.0 | **reuse** | `laconic-wallet` @ `bb5223a` (`src/utils/accounts.ts`) · `laconic-wallet-web` @ `2a4a478` |
| In-browser Nitro node + libp2p relay | T6.5 (interim) | **reuse** | `ts-nitro` @ `884d616` (`@chainsafe/libp2p-noise` via `@cerc-io/peer`) |
| Submit-on-behalf keeper / broadcaster relay | T6.5 | **reuse (ref.)** | `mobymask` @ `2329198` · `waku-broadcaster-client` |
| Groth16 prover (Railgun circuits) | T6.6 | **reuse (constraint)** | T0.0 `wasm`+`zkey`; snarkjs-WASM (browser) / rapidsnark (mobile) |
| WASM note-scanner + WebView↔RN key bridge | T6.1 | **build** | net-new glue |
| Self-hosted watchtower loop | T6.3 | **build** | ties to T0.2/T0.3 dispute path |
| Address rotation + per-venue loyalty keys | T6.4 | **build** | HD-derivation policy |
| Native gomobile transport (production) | T6.5 | **build (port)** | `go-waku` `library/mobile` · go-libp2p-noise |

**What the wallet must blind, and with what** — each surface is a sibling component the wallet *drives*, not something it re-solves:

| Observer | Would learn without mitigation | Wallet-side mitigation |
|---|---|---|
| RPC / data provider | Exactly which commitments/nullifiers you scan — fingerprints you against the pool ★ | Local scan+decrypt off the T2 proof-carrying feed (T6.1) — no public RPC query |
| On-chain / mempool | Your EOA as tx origin; amounts in the clear | Keeper/broadcaster submit-on-behalf (T6.5) + T0.0 shielded pool |
| The venue / matcher | Who you are, funding source, that two trades are you | Fresh per-channel identity via rotation (T6.4); Design-A boundary (T0.3) |
| A misbehaving party | Could stall you | Self-hosted watchtower + unilateral force-close (T6.3) — never custody |

The ★ read-path surface is the one the whole anonymity-set guarantee rests on: if reads leak, every other protection is moot. It is therefore the wallet's do-first item (T6.1).

## T6.0 Wallet base

The base is the **Laconic wallet**, a React-Native app (Android/iOS) with a browser build that custodies keys and signs both Cosmos and EIP-155 requests. It already ships **BIP-39 mnemonic generation, HD derivation, and OS-secure-enclave storage** (iOS Keychain / Android Keystore) via `react-native-keychain` in `src/utils/accounts.ts` — the one piece that is production-grade today. This is **reuse**, not build: fork [`laconic-wallet`](https://git.vdb.to/cerc-io/laconic-wallet) @ `bb5223a` and the browser [`laconic-wallet-web`](https://git.vdb.to/cerc-io/laconic-wallet-web) @ `2a4a478` at pinned commits, then confirm that BIP-39 + HD + `react-native-keychain` custody builds on device/emulator and in the browser.

Everything else in T6 is additive signing paths and hosted stacks layered on this base. T6.0 exposes the **enclave-guarded key store** that T6.1's bridge, T6.2's channel signing, T6.3's watchtower submit, and T6.4's derivation policy all consume, and raw key material never leaves it. Custody is the trust root of the tier: an audited leak here is critical, so the enclave boundary — and the T6.1 bridge across it — is the do-first audit surface before mainnet.

**Key tasks.** Fork at pinned commits; confirm the custody path (BIP-39 generate, HD derive, `react-native-keychain` store/retrieve) builds and runs on device / emulator and in the browser build before anything else lands; establish the enclave as the single key origin every later component calls through.

*Risk:* the key store is the tier's trust root — any path that materializes a secret outside the enclave is an audit blocker.

## T6.1 WASM note-scanner + WebView↔RN key bridge

Two coupled net-new pieces make on-device privacy real.

1. **WASM note-scanner.** A local scan-and-decrypt loop pulls the proof-carrying commitment/nullifier slice from T2's watcher feed, verifies the storage/Merkle proof against L1 (so the phone trusts math, not the watcher), then **trial-decrypts commitments locally** to find the user's notes and compute balances. No `eth_getLogs` / `eth_getStorageAt` ever leaves the device to a public RPC. This is the ★ read-path property the whole anonymity-set guarantee rests on: if reads fingerprint the user, the set collapses to one and every other protection is moot, which is why it is the wallet's do-first item. The scanner consumes T0.0's commitment-insert / nullifier-spend event ABI and the T2 feed format.
2. **WebView↔RN key bridge.** Decryption and signing need the HD keys, but keys must never leave the enclave. The bridge exposes a **narrow, audited message channel**: the WebView (scanner / prover / Nitro node) requests a signature or a viewing-key decryption; the RN shell performs it against `react-native-keychain`-held secrets (T6.0) and returns only the result. Raw key material never crosses into the WebView. This is the net-new glue that lets the reused browser stack operate on enclave-guarded keys.

The bridge is transport-agnostic and survives the T6.5 interim→production swap unchanged, and the scanner is written once against the feed interface and reused verbatim by the T6.3 watchtower (the same loop on a different event).

**Key tasks.** Build the WebView↔RN bridge (narrow sign / viewing-key-decrypt request-response, keys stay in enclave) before the scanner, then build the scanner (feed pull → proof verify → local trial-decrypt → balances). Assert on network egress: no public-RPC query may ever be issued.

*Testing:* on the fixturenet, shield then scan on device — balances must match with zero `eth_getLogs` / `eth_getStorageAt` to a public endpoint. *Risk:* the bridge is a trust boundary; a leak of key material across it is critical and gated on audit before mainnet.

## T6.2 Settlement client

The settlement client is the wallet logic that **drives T0.3's channel lifecycle**: open / fund-from-note, propose / accept state, cooperative close, force-close, and `respond`. It binds to T0.3's deposit/payout contract address + ABI, its `Deposit` / `Payout` events, its voucher format, and its ForceMove **app registry** (the allow-listed app IDs the wallet may play — e.g. the T4 quote/settle app). A trade runs as follows: shield into the pool (T0.0), fund a channel from a note through T0.3, update channel state off-chain via signed messages with the counterparty/venue, then settle by minting fresh shielded notes back through T0.3's payout path.

This is the peer that ties the spine together: it consumes T6.1's bridge for state signatures, T6.5's transport for the message hot loop, and T6.6's prover for the shield/unshield proofs. Amounts are public only at the T0.3 deposit/payout boundary (Design A); hiding them in-play is T0.6 (out of this tier). The client reuses go-nitro's channel semantics and never re-derives them — it is a driver over T0.3, not a re-implementation of the adjudicator.

**Interface consumed.** From T0.3: deposit/payout address + ABI, `Deposit` / `Payout` events, the channel lifecycle API (open / fund-from-note, propose / accept, cooperative close, force-close, `respond`), voucher format, and the ForceMove app registry (allow-listed app IDs). It exposes the driven channel state to T6.3 (the co-signed higher-turn states the watchtower replays).

**Key tasks.** Wire the lifecycle calls over T6.5 transport; fund-from-note and cooperative-close happy path first, then force-close / `respond`. *Testing:* fixturenet happy path — shield → open+update a channel → settle to fresh notes.

## T6.3 Self-hosted watchtower

The watchtower is **the same T6.1 note-scanner loop watching a different event**: it subscribes to T0.2's adjudicator `ChallengeRegistered` feed (delivered over the T2 feed) and, on a challenge that would finalize a **stale** state, submits the user's **higher-turn co-signed state** via the adjudicator's `checkpoint` (higher-turn, no finalize) within the challenge window. Non-custody means a dead counterparty freezes trading, never funds — but only if *someone* answers the challenge for an offline user, which is exactly this loop's job.

**The auto challenge-response is trivial once the client/node tooling is correct.** Every hard primitive already exists: `ForceMove.checkpoint` on the adjudicator (T0.2), the `ChallengeRegistered` feed off the T2 watcher, and the Checkpoint tx build the T6.2 settlement client already drives. Detecting a stale-finalizing challenge and firing the higher-turn `checkpoint` before `FinalizesAt` is a thin loop over those primitives — **not new dispute logic**. The real work is the **tooling + liveness**: ts-nitro today surfaces only fund/defund/pay, so we add an **in-browser dispute API** that exposes the challenge/checkpoint surface to the wallet, plus an **always-on node** so the loop still answers while the phone is asleep.

**A phone can run the loop — but robust liveness needs an always-on node.** It needs only (a) the feed subscription the wallet already maintains and (b) the ability to sign + submit a single `checkpoint` transaction — the same Checkpoint tx build T6.2 already drives, relayed gaslessly via the keeper/broadcaster (T6.5). With no prover and no heavy state, it is a lightweight reactive loop inside the wallet process, phone-deployable by construction. But because a browser/mobile node has no guaranteed background watch loop, the **liveness prerequisite** is an always-on node: users wanting stronger liveness delegate the identical loop to an always-on watcher party (T2), while the default self-hosted loop runs on the device.

**Freshness gate.** The watchtower is only as safe as the feed is fresh: a lagging T2 emitter means a missed challenge, so the loop **gates on T2 feed freshness** — it treats a stale feed as a liveness alarm (fall back to a redundant party / direct submit) rather than silently missing the window. The safety-critical correctness (never miss a valid challenge, never submit a superseded state) is owned jointly with T0.2/T0.3; this doc owns the phone-resident *deployment* of that loop. It is the highest-priority test surface in the tier.

**Key tasks.** Build the **in-browser dispute API** (surface challenge/checkpoint over ts-nitro, which today exposes only fund/defund/pay); subscribe to `ChallengeRegistered` over the T2 feed; on a stale-finalizing challenge, drive the existing Checkpoint tx build to submit the higher-turn state via `checkpoint` before `FinalizesAt`, relayed gaslessly (T6.5); add the feed-freshness alarm + always-on/redundant-party fallback for liveness.

*Testing (safety-critical, must pass):* the phone-resident watchtower defeats a stale-state force-close — a counterparty force-closes an old state, the offline user's watchtower submits the higher-turn state, and the correct outcome finalizes. *Risk:* a missed or superseded response can cost a user funds — highest-priority test surface, owned with T0.2/T0.3.

## T6.4 Address rotation + per-venue loyalty keys

Both are Client-tier key-derivation policy over T6.0's existing HD tree — no new crypto, no contract change.

- **Address rotation.** Derive a **fresh per-channel identity** for each channel/deposit so the venue cannot link two trades to one person and so the transparent Design-A deposit boundary (T0.3) is decorrelated across interactions. Rotation is a derivation-path + note-recipient policy; the T6.1 scanner already sweeps the whole HD range, so rotated notes are found with no extra bookkeeping.
- **Per-venue loyalty keys.** For features that *want* continuity (rebates, reputation, allow-list membership at a chosen venue), derive a **stable per-venue key linkable ONLY to that venue** — a deterministic child key keyed by venue id. It gives the user a persistent handle at exactly one venue while staying unlinkable across venues and unlinkable to the rotating per-channel identities. The T4 venue opts into requiring it as a match filter; the default remains rotated / identity-blind.

**Key tasks.** Implement rotation as a per-channel derivation-path + note-recipient policy (no scanner change — T6.1 sweeps the whole HD range); implement loyalty keys as a deterministic venue-id-keyed child. *Testing:* rotated notes are still discovered by the scanner; the loyalty key is stable at its venue and unlinkable across venues.

## T6.5 Mobile transport

Transport is the tier crux, the load-bearing unbuilt piece. It carries two traffic classes, per ADR-0008 (both, per need — not one or the other):

- **Waku pub/sub** for gossip, discovery, async, and interop with Railgun's broadcaster network ([`waku-broadcaster-client`](https://github.com/Railgun-Community/waku-broadcaster-client)): recipient-unlinkable, and mandatory for the gasless origin-privacy submit path.
- **libp2p-noise direct streams** for the Nitro settlement hot loop — cooperative fills need tens-of-ms round trips that Waku's multi-hop store-and-forward cannot meet. This is the same `/nitro/msg/1.0.0` noise transport go-nitro's p2p message service (`node/engine/messageservice/p2p-message-service/service.go`) uses.

**One sequenced path, interim → production — not competing options:**

- **Interim (spine demo).** Host the existing browser peer stack **inside a WebView** in the RN wallet: `@cerc-io/peer` + [`ts-nitro`](https://github.com/cerc-io/ts-nitro) @ `884d616` run unmodified in the WebView JS runtime — circuit-relay + webrtc-star + gossipsub for connectivity, `@chainsafe/libp2p-noise` for direct streams. The RN shell owns keys + UI; the WebView owns the Nitro node + p2p. This reuses production browser code verbatim and is the realistic v1 mobile client (foreground-only) that validates the whole flow.
- **Production (Phase-3 hardening, before mainnet).** Move the hard layer native: the whole p2p / channel / messaging layer becomes **one gomobile module** — **go-nitro + [go-waku](https://github.com/waku-org/go-waku)** (`library/mobile`), both Go on mature go-libp2p + noise, compiled to `.aar` / `.xcframework` and bridged to the RN shell (the proven `status-go` pattern). The port is required, not optional polish, because `@waku/sdk` / js-libp2p is not React-Native-compatible (no WebCrypto / WebRTC / Node built-ins in Hermes), and the interim WebRTC-over-tunnel path is the least-proven part of the combination. Native modules escape the WebView's background-execution constraint and give the phone a first-class libp2p peer. RN stays the **shell only**; `ts-nitro` remains the browser build's node.

The boundary between the phases is the transport interface T6.2/T6.3 consume, and the scanner, prover, and channel-driving logic are written **once** against it and survive the swap. **Write-time origin privacy** uses the keeper/broadcaster submit-on-behalf pattern ([`mobymask`](https://github.com/cerc-io/mobymask) @ `2329198` keeper / Railgun broadcaster) so on-chain `msg.sender` is never the user's EOA — metered by the same T2 voucher mechanism.

**Key tasks.** Stand up the interim WebView stack and smoke a direct + virtual channel against a T0.3 fixturenet; define the transport interface T6.2/T6.3 bind to; then build the go-nitro + go-waku gomobile module behind that same interface and re-run E2E. *Risk:* WebRTC-over-tunnel in the interim may not traverse a network underlay cleanly — the least-proven part of the combination and a core reason the native port exists, not optional polish. Native gomobile E2E is a Phase-3 gate before mainnet, not before the demo.

## T6.6 Groth16 mobile proving

Client-side proving is the dominant mobile difficulty, and the decision is settled: **keep Railgun's Circom + Groth16 / BN254 construction** — the win is a better *prover runtime*, not a different pool. The prover loads T0.0's circuit `wasm` + `zkey` artifacts + hashes (produced by T0.0 / T0.1's trusted setup) and never regenerates them.

- **Browser build:** snarkjs-WASM ([snarkjs](https://github.com/iden3/snarkjs)) against T0.0's `wasm` + `zkey`.
- **Mobile:** a **native Groth16 prover** (mopro / rapidsnark). Measured on a 2024 phone, native proof-gen is ~0.15–3.4 s and **~8–20× faster than browser snarkjs** for small-to-medium circuits — a mostly-solved *engineering* problem, shipped at scale by World App.

The **binding limit is memory, not speed**: a Railgun-scale transaction circuit (large Merkle note-scan + shielded transfer) sits at the upper edge of on-device provers and can hit the ~3 GB app-memory wall that crashes heavier circuits (`mobile-proving-research.md`; [zkmopro](https://zkmopro.org/docs/performance)). The mitigations, in maturity order, are to **benchmark the actual Railgun circuit first** on target devices (mopro's own advice), then add a **coSNARK server-assist fallback** for weak / low-memory devices — its output is still a standard Groth16 proof, verifier unchanged — which may become the default floor on low-end hardware. Switching to Halo2 / UltraHonk is rejected: it costs mobile prover perf, on-chain gas, and a fresh audit, and Halo2-RSA already crashes on today's phones. Groth16 stays.

**Key tasks.** Wire snarkjs-WASM (browser) and native rapidsnark (mobile) against T0.0 artifacts; **benchmark the real Railgun circuit** on target devices before committing to on-device-only; implement the coSNARK server-assist fallback path for low-memory devices. *Risk:* a Railgun-scale circuit may exceed phone app-memory; the assist may have to be the default floor on low-end hardware.

## T6.7 Armada-branded app + SDK wiring

The shipped starting point is honest: the Laconic wallet is **bare-bones**, and MobyMask / the swap flows are **demo fragments, not products** — the right foundation, not a finished front-end. T6.7 is the product build-up on top of T6.0: grow the bare-bones wallet into the **Armada-branded app** and **wire the Armada SDK** so integrator apps, treasury / payment tools, and the front-end all sit on the same client stack.

Concretely this is net-new "build up," not new protocol. First brand + productize the RN shell and browser build (UI, WalletConnect, orchestration in Hermes), then expose the T6.1–T6.6 capabilities — local scan, T6.2 settlement, T6.5 transport, T6.6 proving — as a coherent **Armada SDK surface** that integrator apps call without re-solving privacy. Armada is pluggable asset-privacy infrastructure for USDC (shielded pool + governance-added adapters + apps/SDK tier); T6.7 is that apps/SDK tier, funded by integrator revenue-share. It routes to the T4 venue and, through T5, to Armada's adapters (Swaps, Aave-v4 yield, CCTP), so an app built on the SDK gets shield → trade → settle → scan for free.

**Key tasks.** Brand + productize the RN shell and browser build; define the Armada SDK surface over T6.1–T6.6; wire routing to the T4 venue and, via T5, to the Swaps / Aave-v4 / CCTP adapters so an SDK app inherits shield → trade → settle → scan without re-solving privacy.

## T6.8 Identity (optional)

Identity is **not** cross-cutting: it is a single optional parameter the wallet may attach, default off. A user can source a selective-disclosure attestation off-band — e.g. [Self](https://github.com/selfxyz/self) (zk-passport): prove nationality / age / OFAC-clear / personhood on-device from a passport or EU-ID chip in the standalone Self app, verified **off-chain and anchored to an Ethereum address**, with the shielded Railgun identity left unlinked. The wallet stores the resulting attestation as one more optional signing input, surfaced only when a T4 venue's match filter requires it. Deferred, not built.

## Sources

- laconic-wallet (custody base) — https://git.vdb.to/cerc-io/laconic-wallet @ `bb5223a` · browser build https://git.vdb.to/cerc-io/laconic-wallet-web @ `2a4a478`
- ts-nitro (in-browser / mobile Nitro node; `@chainsafe/libp2p-noise` via `@cerc-io/peer`) — https://github.com/cerc-io/ts-nitro @ `884d616`
- go-nitro p2p message service (libp2p-noise `/nitro/msg/1.0.0`) — https://github.com/cerc-io/go-nitro (`node/engine/messageservice/p2p-message-service/service.go`)
- go-waku (gomobile `library/mobile` bindings) — https://github.com/waku-org/go-waku
- Railgun Waku broadcaster client (submit-on-behalf) — https://github.com/Railgun-Community/waku-broadcaster-client
- mobymask keeper (submit-on-behalf reference) — https://github.com/cerc-io/mobymask @ `2329198` (`packages/server/index.ts`)
- watcher-ts (proof-carrying feed, voucher metering) — https://github.com/cerc-io/watcher-ts @ `18ca4e1` (`indexer.ts`, `payments.ts`)
- Railgun contracts / circuits — https://github.com/Railgun-Privacy/contract · https://github.com/Railgun-Privacy/circuits-v2
- snarkjs (browser Groth16 prover) — https://github.com/iden3/snarkjs
- Mobile proving — `mobile-proving-research.md` (in repo) · https://zkmopro.org/docs/performance
- Self (optional zk-passport identity) — https://github.com/selfxyz/self
