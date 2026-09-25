# Nitro Bridge + DSS — Code Audit

audit · 2026-09-10 · read-only source review

**Repos & pins audited**

| Component | Repo | Ref (SHA) | Notes |
|---|---|---|---|
| Nitro bridge | `cerc-io/go-nitro` (fork of `statechannels/go-nitro`) | `435eb2b0…90c665` (`main` HEAD) | This **is** our pinned `@435eb2b`; the bridge exists at the pin. |
| DSS crypto library | `cerc-io/chain-signatures` | `main`/`v0.1.0` `9016a7c4…`, `dev` `09e5ac97…` | GitHub 404; read via `git.vdb.to` mirror. |
| DSS service | `cerc-io/laconicd` | `roysc/nitro-integration` `d130608f…` | Local mirror `code/laconicd`; imports `chain-signatures v0.1.0`. |

---

## Summary

**Verdict.** The Nitro bridge is a **working L1-anchored demo, not a trust-minimised bridge.** Its happy path (mirror an L1 ledger channel onto an L2, pay off-chain, exit back to L1) is implemented and passes an anvil-backed end-to-end test. But the two properties that would make a cross-chain bridge "extremely secure" — a real L2 chain with its own adjudication (so a user can unilaterally exit on the L2 side) and a decentralised signer in place of a single operator key — are **not built**. Today one operator, holding one private key for both endpoints, custodies all real collateral on L1 and is the user's counterparty on both hops. The DSS that is meant to decentralise that key exists as separate, unaudited, not-yet-wired code.

**How much is written — the bridge.** The core is `bridge/bridge.go` (~503 LOC) plus three objectives: `bridgedfund` (~423), `bridgeddefund` (~263) and `mirrorbridgeddefund` (~463). Mirroring (clone the L1 state, swap allocations, recreate the ledger on L2 with **no on-chain deposit**), cooperative exit, and a ForceMove **challenge/checkpoint** exit are all implemented — **but only on L1**, against the real `NitroAdjudicator` and a generated `Bridge` contract binding. The L2 chain service `node/engine/chainservice/laconicd_chainservice.go` (~74 LOC) is a **complete stub**: `SendTransaction` is a no-op (`return nil, nil`), every feed/getter returns nil/zero, and `internal/node/bridge.go` still reads `// TODO: Implement laconicd chain service`. So L2 has **no** deposit, adjudicator, challenge, or withdrawal — L2 balances live only as co-signed off-chain states plus the bridge's single-node buntdb map. Rough completeness: mirroring ~85%, L1 exit ~60–70%, **L2 on-chain enforcement ~0–5%**. Tests: one unit test + a ~970-LOC happy-path e2e; **no** adversarial, offline-operator, or L2-on-chain coverage. `go.mod` has **no DSS/threshold dependency** — the bridge signs with a single ECDSA key everywhere.

**How much is written — the DSS.** Split across two repos. `chain-signatures` is a **pure crypto library** (~90% done, unit-tested, but every file banners *"XXX: Do not use in production until this code has been audited."*): `ethschnorr` (single-signer, Ethereum-verifiable Schnorr on secp256k1/keccak256) and `ethdss` (a `(t,n)` distributed Schnorr, Stinson–Strobl, from DEDIS kyber). The **service** — `laconicd/server/distsig` — drives a kyber **Rabin/Pedersen DKG** to produce the shared key and per-signature nonce, then combines `≥t` partial signatures into one group Schnorr signature (~70% at the logic level, unit-tested with a 4-of-7 group). Its **production wiring is ~10%**: the CometBFT ABCI++ vote-extension transport is stubbed (`ExtendVoteWithLaconic` returns an empty extension; `VerifyVoteExtensionWithLaconic` "just accept"s), the `Manager` is **never instantiated in production**, the Nitro consumer is commented out (`UseDistsig`, `distsigManager`), threshold config isn't in genesis, and the qualified-set filter is commented out. **End-to-end, the DSS does not run.**

**Bridge trust model (as coded).** Single trusted operator. `bridge.go` builds both node stores from the one `configOpts.StateChannelPK` (comment: *"the same private key is being used for both nodes"*); `cmd/start-bridge` exposes exactly one state-channel key flag. Real custody is the L1 ledger channel's ERC20 in the `NitroAdjudicator`; L2 has no custody contract. A user **cannot** unilaterally exit with their true L2 balance: there is no L2 adjudicator, and the L1→L2 reflection (`MirrorBridgedDefund`) is **only ever initiated by the bridge's own L1 node** — the user's L2 node never holds the L1 `ConsensusChannel` needed to push an L2 state onto L1. The user can challenge the *L1* ledger, but that enforces the L1 state, which does not reflect L2 payments unless the (single-key) bridge acts. Liveness and L2 data-availability therefore rest on **one honest, online operator**, enforced by nothing in code (dropped-tx retry is bounded to 1). This is the grounded correction to the "extremely secure / unilateral exit" framing raised earlier in the design discussion: **as written, it is an operator-trust bridge.**

**How the DSS works.** A kyber Rabin DKG (rounds: deal → certify → commit → finish) jointly creates a long-term distributed secret and a one-time nonce; each signer emits a partial `γᵢ = βᵢ − αᵢ·h(m)` signed with its own `ethschnorr` key; a collector verifies each partial against the public commitment polynomials (identifiable abort) and Lagrange-interpolates once `≥t` valid partials arrive, yielding one on-chain-verifiable Schnorr signature. Threshold `t = ceil(n·ratio)`, default **4/7**. Trust: `≥t` colluding signers can forge any group signature (standard); `<t` cannot; but a minority of size `n−t+1` can **stall liveness** (no timeout/fallback), and there is **no bonding or slashing** of signers in either repo — so no economic security on the set. The secp256k1 suite is explicitly **not constant-time**.

**Bottom line for Armada.** This maps to our v2/net-new items **T2.4** (federation + threshold DKG signing = the DSS, bonded via **T0.4**) and **T0.4/T0.5** (registry/bond, fraud-proof verifier). Our registry already marks these v2 and unbuilt — consistent with the audit. An armada-nitro cross-chain bridge is viable **in principle**, but reaching its security story requires four unbuilt pieces: a real L2 chain service (unilateral L2 exit + self-watchtower), replacement of the single operator key with the DSS signer set, bonding/slashing for economic security, and the security audit both codebases' own headers demand.

---

## 1. Scope & provenance

Read-only review of the source at the refs in the table above. Every claim below is anchored to a file (and, where quoted, a line) at those SHAs. `chain-signatures` and `laconicd` are also present in this repo under `code/` (the local mirror `code/laconicd` pins `roysc/nitro-integration @ d130608`).

Confirmed: the go-nitro bridge package exists at our pin `@435eb2b` (it is `main` HEAD, commit *"Remove sending swap updates to client (#110)"*, 2024-10-30), so this is not a pre-bridge commit. The DSS is **not** part of go-nitro at any ref — it is `chain-signatures` (library) consumed by `laconicd` (service).

## 2. What the bridge is (architecture)

One operator process runs **two** Nitro nodes that share one state-channel key: an **L1 node** on a real EVM chain (`NitroAdjudicator` + a generated `Bridge` contract) and an **L2 node** on a stubbed "laconicd" chain.

Flow (`bridge/bridge.go`, `node/node.go`):

1. User `A` opens a `directfund` ledger channel with the bridge on **L1** — a real ERC20 deposit into the adjudicator (`node_test/bridge_test.go`).
2. On L1 fund completion the bridge clones the L1 supported state, **swaps `Allocations[0]<->[1]`**, and calls `nodeL2.CreateBridgeChannel(...)` → `bridgedfund` recreates an equivalent ledger on **L2 with no on-chain deposit** (`node.go`: *"No chain interactions are involved while creating this channel"*).
3. On L2 the mirrored ledger funds virtual/payment channels; payments move balances **off-chain** only.
4. Exit: `APrime` runs `CloseBridgeChannel` (`bridgeddefund` → a final `IsFinal` L2 state); the bridge sees it complete and runs `nodeL1.MirrorBridgedDefund`, which in `mirrorbridgeddefund.CreateL1StateBasedOnL2` rebuilds the L1 outcome from the finalized L2 state (re-swap allocations, `TurnNum++`), finalizes the L1 channel **cooperatively** (`crank` → `NewMirrorWithdrawAllTransaction`) or **via ForceMove** (`crankWithChallenge` → `SignChallengeMessage`/`NewChallengeTransaction`/`NewCheckpointTransaction` → `NewMirrorTransferAllTransaction`), then withdraws on L1.

Outcome equivalence between the two chains is maintained purely by the allocation-swap + `TurnNum` bump. **All real custody and settlement is on L1; L2 is off-chain-only because its chain service is a stub.**

Key types: `Bridge{ nodeL1/storeL1/chainServiceL1; nodeL2/storeL2/chainServiceL2/messageServiceL2; mirrorChannelMap }` and `MirrorChannelDetails{ L1ChannelId, IsCreated }` (`bridge/durablestore.go`, single-node buntdb, not replicated).

## 3. Completeness — file inventory & status (`@435eb2b`)

| File | LOC | Status |
|---|---|---|
| `bridge/bridge.go` | ~503 | Orchestrator; implemented (happy path). |
| `bridge/durablestore.go` | ~115 | buntdb L1↔L2 map; implemented. |
| `bridge/policy_maker.go` | ~35 | L1/L2 permissive policies; implemented. |
| `bridge/durablestore_test.go` | ~50 | **Only** bridge unit test (store round-trip). |
| `protocols/bridgedfund/…` | ~423 | L2 fund (off-chain only); implemented. |
| `protocols/bridgeddefund/…` | ~263 | L2 defund (off-chain only); implemented. |
| `protocols/mirrorbridgeddefund/…` | ~463 | L1 exit (coop + challenge); implemented — **L1 only**. |
| `node/engine/chainservice/laconicd_chainservice.go` | ~74 | **STUB** — all no-ops. |
| `node/engine/chainservice/bridge/Bridge.go` | ~83 KB | Generated L1 `Bridge` contract binding (real, L1). |
| `internal/node/bridge.go` | ~37 | `InitializeL2Node`; carries the L2 TODO. |
| `cmd/start-bridge/main.go` | ~230 | Daemon; single `StateChannelPK` for both nodes. |
| `node_test/bridge_test.go` | ~970 | Anvil e2e, **happy path** (create→mirror→pay→exit). |
| `rpc/bridge-server.go`, `node_test/bridge_rpc_test.go` | — | RPC surface + test. |

**Quoted incompleteness:**

- `laconicd_chainservice.go`: `SendTransaction(...) { return nil, nil }`; `EventFeed() { return nil }`; `GetChainId() { return nil, nil }`; `GetLastConfirmedBlockNum() { return 0 }`; `GetBlockByNumber` returns an empty block. → L2 on-chain deposit/adjudication/challenge/withdraw **0%**.
- `internal/node/bridge.go`: `// TODO: Implement laconicd chain service`.
- `node/node.go` `CreateSwapChannel`: `// Since no contract present for swap channels yet`, `// TODO: Handle sad path for swap channels`.
- `node/node.go` `handleError`: `// TODO instead of a panic, errors should be returned to the caller` then `panic(err)`.

Rough %-complete: mirroring ~85%; funding/collateral (L1 real, L2 none) ~50%; exit/dispute (L1 real, L2 none) ~60%; L2 chain service ~5%; daemon ~90% (but drives an unfinished L2).

## 4. Bridge trust model (as coded)

- **Custody:** real ERC20 collateral sits in the L1 `NitroAdjudicator` for the `A↔B` ledger. **L2 has no custody contract** (stub chain). The operator is participant `B` on L1 **and** `BPrime` on L2 under **one key** (both stores built from `configOpts.StateChannelPK`; `GetNodeInfo` comment *"the same private key is being used for both nodes"*).
- **Signing:** standard Nitro 2-of-2 co-signing per ledger; the L1 challenge is signed by the bridge's single L1 key (`NitroAdjudicator.SignChallengeMessage(..., *secretKey)`).
- **Malicious/offline operator:** if the L2 peer is unreachable at mirror time the bridge simply defunds the L1 channel. The L2→L1 reflection is initiated **only** by the bridge (`processCompletedObjectivesFromL2 → nodeL1.MirrorBridgedDefund`). A `CounterChallenge` API exists for responding to a challenge.
- **User unilateral exit?** **Not established / not implemented** for reflecting L2 balances. No L2 adjudicator exists; `mirrorbridgeddefund.NewObjective` needs the L1 `ConsensusChannel` that the user's L2 node does not hold. The user can challenge the *L1* ledger, but that enforces the L1 state, which omits L2-side payments unless the bridge acts. **Exit correctness depends on the single-key bridge being live and honest.**
- **Liveness / DA:** the only durable L2 record is off-chain co-signed states + the bridge's single-node buntdb map; no L2 chain persists state; dropped-tx retry is bounded to 1. The design **assumes one honest, non-crashing operator** on the L2 side — enforced by nothing in code.

**DSS linkage:** none. `go.mod@435eb2b` has no `chain-signatures`/threshold/MPC module; no bridge file imports a distributed signer. "consensus" in this repo means only Nitro's own on-chain `ConsensusApp` (ForceMove) and the 2-party `consensus_channel` — neither is a DSS.

## 5. The DSS — mechanism and (non-)integration

**Layer 1 — `chain-signatures` (pure crypto, no I/O):**
- `ethschnorr` — single-signer Schnorr on secp256k1, `ChallengeHash = keccak256(pubkey‖msg‖rAddress)`, on-chain-verifiable via `SchnorrSECP256K1.sol` (contract not in repo).
- `ethdss` (336 LOC, pkg `clientdss`) — `(t,n)` distributed Schnorr (Stinson–Strobl, from kyber `sign/dss`): `PartialSig` / `ProcessPartialSig` (verifies the sender's `ethschnorr` sig **and** checks the partial against the public polynomials) / `EnoughPartialSig` (`len ≥ T`) / `Signature` (`share.RecoverSecret` Lagrange over `T` shares).
- `secp256k1` — kyber Group/Point/Scalar over btcec; header warns **"NOT CONSTANT TIME!"**.
- Every file: *"XXX: Do not use in production until this code has been audited."*

**Layer 2 — `laconicd/server/distsig` (the service):**
- `manager.go` — `Manager` per validator; `initDKG` builds a kyber **Rabin** DKG (`dkg.NewDistKeyGenerator`, `t = ceil(len(members)·ThresholdRatio)`); `StartDKG/StartSignature/ProcessMessages/CompletedSignatures`. **Never instantiated in production.**
- `keygen.go` — Rabin/Pedersen Joint-Feldman DKG state machine (deal→certify→commit→finish→done). Comments: prepareMessages *"during ExtendVote"*, processMessages *"during VerifyVoteExtension (TODO: verify)"*.
- `signature.go` — wraps `clientdss`; `ProcessPartialSig` then final `Signature()` once `EnoughPartialSig()`.
- `config.go` — `Enable=false` by default; `ThresholdRatio=4/7`; `// TODO: set this in genesis`.

**Threshold model:** `≥t` colluding signers **can** forge any group signature; `<t` cannot; a minority of size `n−t+1` **can stall** signing (no timeout/fallback in `ethdss`). **No bonding, no slashing** of the signer set in either repo (`laconicd`'s `x/bond` is a name-registry deposit; `x/slashing` is generic validator downtime) — so **no economic security** on the signers. Membership = the pubkey set passed to `StartDKG`; join/leave = re-running the DKG (no proactive resharing).

**Integration status:** consumed by `laconicd` (`go.mod` imports `chain-signatures v0.1.0`; `distsig` imports `clientdss`/`ethschnorr`/`secp256k1`), **not** by go-nitro. The consensus transport is stubbed (`app/abci.go`: `ExtendVoteWithLaconic` returns an empty vote extension with *"TODO: Implement the full distsig logic here"*; `VerifyVoteExtensionWithLaconic` *"For now, just accept"*). The Nitro consumer is **disabled/declared-only**: `server/nitro/config.go` `UseDistsig` commented out; `x/nitro/keeper/keeper.go` `// distsigManager *distsig.Manager`; `server/nitro/server.go` roadmap *"initialize group delegate credentials, with distsig over ABCI … bind custodian contract."* The intended end state — the laconicd validator set threshold-signing as the bridge's custodian instead of one operator key — is **not built**.

## 6. Gaps & risks (priority order)

1. **L2 has no on-chain enforcement** (stub chain service). Without an L2 adjudicator there is no user-side unilateral exit and no self-watchtower on L2 — the single largest gap.
2. **Single operator key for both hops.** Full operator trust today; a compromised/absent operator is a fund-safety and liveness risk. The DSS is the intended fix but is unwired.
3. **DSS not operational** end-to-end (ABCI transport stubbed, Manager not instantiated, Nitro consumer commented out) and **unaudited** by its own declaration; secp256k1 not constant-time; qualified-set filter commented out.
4. **No economic security** on the signer set (no bond/slashing) — signer misbehaviour has no on-chain penalty.
5. **Test coverage is happy-path only** — no adversarial, offline-operator, or L2-on-chain tests; bridge node errors `panic` rather than return.
6. **Liveness assumption:** with a 4/7 default, any 4 of 7 signers offline halts signing; and the bridge's L2 DA is a single non-replicated buntdb.

## 7. Relevance to Armada (this repo)

- The DSS/federation is exactly our **T2.4** ("federation + bond + threshold DKG signing — chain-signatures DSS; bond = T0.4"), and the bond/fraud-proof scaffolding is **T0.4/T0.5** — all marked **v2** and net-new in the §5 registry. The audit confirms they are unbuilt, so the registry is accurate.
- An **armada-nitro bridge to other chains** would compose this bridge with work-package A's settlement rail. As-built it is a **trusted single-operator, L1-anchored** construction with a stub L2 — not the trust-minimised bridge the earlier discussion imagined. It becomes trust-minimised only with: (a) a real L2 chain service (unilateral L2 exit + self-run watchtower), (b) the DSS signer set replacing the single `StateChannelPK`, (c) bonding/slashing (T0.4/T2.4) for economic security, and (d) the audits both codebases require. None of (a)–(d) is in scope of v1 A/B/C/D.
- Recommendation: treat a cross-chain bridge as a **separate v2 design axis**, not an extension of the v1 rail, and do not describe it as "secure" beyond "single-operator demo" until at least (a) and (b) land.

## 8. Citation index

- **Bridge:** `bridge/bridge.go`, `bridge/durablestore.go`, `protocols/{bridgedfund,bridgeddefund,mirrorbridgeddefund}/`, `node/engine/chainservice/laconicd_chainservice.go`, `node/engine/chainservice/bridge/Bridge.go`, `internal/node/bridge.go`, `node/node.go`, `cmd/start-bridge/main.go`, `node_test/bridge_test.go`, `go.mod` — all `@435eb2b0…90c665` (`cerc-io/go-nitro`, `main`).
- **DSS library:** `ethdss/ethdss.go`, `ethschnorr/ethschnorr.go`, `secp256k1/suite.go`, `go.mod` — `chain-signatures` `dev@09e5ac97…` / `main@9016a7c4…` (`v0.1.0`).
- **DSS service:** `server/distsig/{manager,keygen,signature,config,suite}.go`, `app/abci.go`, `server/nitro/{config,server}.go`, `x/nitro/keeper/keeper.go`, `go.mod` — `cerc-io/laconicd` `roysc/nitro-integration@d130608f…`.
- Full scout transcripts: `agent://BridgeScout`, `agent://DssScout`.
