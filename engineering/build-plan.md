# Build Plan — one linear sequence

status companion · 2026-10-20 · reconciled to §5 registry, §9 ADRs, §11 risks

The [Building Block View (§5)](./05-building-block-view.md) is the structure and the [ADRs (§9)](./09-architecture-decisions.md) are the decisions; this page is the **one ordered sequence of stages** that gets the system from what is built today to what remains. Every stage states its goal, the items it delivers, what it depends on, its status, and its acceptance check. Done stages come first.

Points and hours use [`estimation-method.md`](./estimation-method.md): one point is 10,000 tokens of peak context; a chunk is one worker's one-session unit of work; one 30-minute session completes about 10 points, or two small (≤5-point) chunks. A stage's hours are its session count × 0.5. Built items have no chunks.

**Live state.** [`notes/handoff-2026-10-20.md`](../notes/handoff-2026-10-20.md), section 0, records the status of every stage and chunk, updated whenever one changes. This page keeps stage status at the level done / in progress / next / later; the handoff note carries the day-to-day detail.

**Review.** The family's husband reviews every chunk before it merges, with extra care on the audit-critical chunks P1, P2a, P2b, N1, N2 and G3. There is no separate reviewer step (operator ruling, CHUNK-RULES.md rule 2).

**Acceptance.** Every chunk's acceptance reuses the armada-nitro environment — the laconic-so fixturenet and reconciler, and the gates under `armada-nitro/env/gates/`: `nitro-e2e`, `pool-e2e`, `watchtower`, `hub-custodian-frost`. A chunk that changes the pool, the circuits, governance or the SDK extends one of these gates (a new step, or a new gate directory beside them) rather than building a separate harness (operator ruling, CHUNK-RULES.md rule 3). The gate chunks G7a/G7b, N7, N8 and C5 are extensions of these gates. Everything is proven on the fixturenet before mainnet (CHUNK-RULES.md rule 4).

**Routing.** A chunk's route is `5090` at up to 9 points; above 9 points it splits or starts on Sonnet (`estimation-method.md` §2, §4). P2 and G7 are split below because they exceeded 9 points. Every pebble is filed starting on the 5090 regardless of route; a worker that outgrows the window escalates to Sonnet through `contextPromotion` (CHUNK-RULES.md rule 1). Circuit chunks carry extra routing risk because circom on the 27B model is unmeasured; P1 and N5 are flagged below.

**Decisions.** Lab book entries, not ADRs, record decisions for now (operator ruling, 2026-10-19/20). This plan does not schedule new ADR writing.

## Stage list

| # | Goal | Status |
|---|---|---|
| S1 | Foundation: pool, circuits (dev ceremony), adjudicator, CCTP + Aave-v4 adapters, the v1 custom governor and revenue contracts, the FROST hub custodian core, the web wallet | done |
| S2 | Contract-owned notes: a non-upgradeable pool deploy with owner-authorized spends | in progress |
| S3 | Hidden-balance ZK channel apps on armada-nitro: the generic `IForceMoveApp` | in progress |
| S4 | Baal treasury governance: a Safe-owned treasury replaces the v1 custom governor | next |
| S5 | armada-nitro side of contract-owned collateral: commitments as channel collateral | next |
| S6 | armada-client Phase 1: daemon with wallet and relayer, e2e on the fixturenet | next |
| S7 | Proof path: patch the fixturenet genesis, operate the nimbus verified proxy, run the first proof | done, by hand; not yet a standing gate |
| S8 | Client verified reads: the Armada verification library, the browser WASM proxy, the mobile `libverifproxy.a` cross-compiles | next |
| S9 | armada-client Phase 2: Waku transport, ts-nitro channels, watchtower, hub role | next |
| S10 | Client support for contract-owned notes and ZK channel apps | next |
| S11 | armada-client Phase 3: mobile (React Native, mobile prover, Android compliance, UI) | later |
| S12 | Watcher federation and the optional `watcher-ts` cache | later |
| S13 | The production Phase-2 ceremony over the final shape set | later |
| S14 | Venue v1: posted-price RFQ, quote/settle app, fee-split | later |
| S15 | v2 / deferred backlog: ordering, venue solver, LP-buffered USDC yield, ex_net matcher, Swaps, routing interface, registry + bond, fraud-proof verifier, native commitment, address-rotation audit, optional identity | later |
| S16 | External audit, mainnet cutover | later |

Order (operator ruling, 2026-10-20): the next work dispatched is, in this order of waves — S2 and S3 in wave 1; S4 and S5 in wave 2, both after P2b. Briefs for all four families (P, N-zk-app, G, N-collateral) are already written and workers start on them first, so S2–S5 sit directly after S1, ahead of the client and proof-path stages.

## S1 — Foundation (done)

**Delivers:** T0.0, T0.1 (circuits compiled; Phase-2 open), T0.2, T5.0a, T5.0b, T6.0, T6.1, T6.2 (partly); the governance/revenue row of §5 T0 (`ArmadaGovernor`, `ArmadaTreasuryGov`, `TreasurySteward`, `ArmadaToken`, `ArmadaCrowdfund`, `ArmadaWindDown`, `ArmadaRedemption`, `AdapterRegistry`, `RevenueCounter`, `RevenueLock`, `RevenueReserveDistributor`); `HubCustodian.sol` — built but not yet a registered `T#.#` id.

**Depends on:** nothing.

**Acceptance (already true):** `armada-poc`'s and `armada-circuits`' Hardhat/TS and Foundry (fuzz, invariant, Halmos) suites pass; `armada-nitro`'s `pool-e2e`, `nitro-e2e` and `hub-custodian-frost` gates are green against the armada-nitro environment's fixturenet.

## S2 — Contract-owned notes (in progress)

**Goal:** the pool deploys non-upgradeable, with owner-authorized contract-owned spends, so a Safe treasury can hold notes directly.

**Delivers:** a hardened T0.0 deploy; new T0.1 circuit shapes.

**Depends on:** S1.

Chunk ids and points are copied from the 209-point list (`labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md` §7); P2 is split into P2a and P2b, each at or under the 9-point cap.

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| P1 | Contract-owned `Transact` variant, differential tests, manifest shapes | — | 9 | 5090 (circom; unmeasured, flagged for escalation) |
| P2a | Owner auth path, preimage registry | P1 | 6 | 5090 |
| P2b | ~~Fee-tagged outputs,~~ pool-level `cumulativeFees`, `ShieldModule` fee note, tests (fee-tagged outputs withdrawn 2026-10-22: there is no swap fee) | P2a | 5 | 5090 |
| P4 | Phase-2 run for the new shapes | P1 | 4 | 5090 |
| P3 | Non-upgradeable deployment: constructor keys, no proxy, deploy script | P2b, P4 | 6 | 5090 |

**Stage total:** 30 points, 4 sessions, 2.0 h.

**Acceptance:** P1 — a differential test of the variant against captured vectors, extends `armada-circuits`' own suite; P2a/P2b — `forge test` for the owner path and the fee registry, extends `pool-e2e`'s coverage; P3 — the non-upgradeable deploy script, checked by the reconciler's `armada` module `verify` step; P4 — `snarkjs zkey verify` over the new shapes.

## S3 — Hidden-balance ZK channel apps (in progress)

**Goal:** a generic ZK `IForceMoveApp` lets a channel hide its balances, proven once at finalization.

**Delivers:** a net-new capability on T0.2 (no `T#.#` id yet; the lab book records the decision).

**Depends on:** S1 (independent of S2).

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| N4 | Generic ZK `IForceMoveApp`, verifier wiring, tests with a trivial circuit | — | 7 | 5090 |
| N5 | Hidden-balance ledger circuit, vectors, ceremony, integration with N4 | N4 | 9 | 5090 (circom; unmeasured, flagged for escalation) |
| N6 | Game template circuit and one example game | N4 | 7 | 5090 |

**Stage total:** 23 points, 3 sessions, 1.5 h.

**Acceptance:** N4 — `forge test` with a trivial circuit through the ZK app, extends `nitro-e2e`; N5 — ledger-circuit vectors and the Phase-2 shapes they add; N6 — the game template exercised by one example game, extends `nitro-e2e`.

## S4 — Baal treasury governance (next)

**Goal:** a Baal (Moloch v3) treasury on a Safe replaces the v1 custom governor; any ARM holder can exit at any time.

**Delivers:** supersedes the governance/revenue row of §5 T0 built in S1.

**Depends on:** S2 (needs P2b).

Chunk ids and points are copied from the 209-point list (`labbook/2026-10-18-treasury-governance-baal-safe-shielded.md` §7); G7 is split into G7a and G7b, each at or under the 9-point cap.

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| G1 | ARM shares token, non-upgradeable, transfer allowlist, with tests | — | 6 | 5090 |
| G2 | Baal + Safe summon script, module wiring, fixturenet deploy | G1 | 8 | 5090 |
| G3 | `ShieldedRagequit` contract with tests, including the exit proof built with the SDK | G2, P2b (S2) | 9 | 5090 (audit-critical) |
| G4 | RevenueLock rewrite on `cumulativeFees`, with tests | G1, P2b (S2) | 6 | 5090 |
| G5 | Hedgey lockup, crowdfund delivers ARM shares, crowdfund proceeds shielded to the Safe | G1, P2b (S2) | 8 | 5090 |
| G6 | Remove old stack, deploy scripts, `GOVERNANCE.md` and `ARM_TOKEN.md` rewrite | G2, G5 | 8 | 5090 |
| G7a | Gate: crowdfund → claim → vote, shield fees only | G3–G6 | 6 | 5090 |
| G7b | Gate: spend → ragequit, shield fees only | G7a | 4 | 5090 |

**Stage total:** 55 points, 8 sessions, 4.0 h.

**Acceptance:** G7a/G7b extend the armada-nitro environment's `pool-e2e`/`nitro-e2e` gates with the crowdfund → claim → vote → spend → ragequit scenario, shield fees only, on the fixturenet; a run where a captured majority blocks an exit, or where a holder cannot exit at any time, must fail. No new harness.

## S5 — armada-nitro side of contract-owned collateral (next)

**Goal:** `ArmadaNitroAdjudicator` holds its channel collateral as contract-owned notes, so collateral never leaves the pool. N8 is withdrawn (there is no protocol swap fee).

**Specification (operator ruling, 2026-10-24):** the adjudicator is rewritten from a Lean model, and the commented Lean model is its specification. The model is in `AFDudley/armada-lean`: branch `owned-notes` (contract-owned notes, a pool feature, proved) and branch `adjudicator-model-v4` (the adjudicator as an ordinary pool user, 23 statements). Rewrite briefs are derived from the model. The N1a–N3, N7, N8 and R1 briefs below are superseded and deleted from armada-nitro; they remain in history at armada-nitro `5519526c`. The merged N1–N3 code on `integrated` has confirmed defects (deposit cannot be funded by any client, reclaim creates value, payouts send a zero proof) and is replaced by the rewrite.

**Verification chain (operator rulings, 2026-10-25):** three layers, two refinement proofs.

1. The Lean model (`AFDudley/armada-lean`) is a symbolic (Dolev–Yao) specification. Hashes are symbolic terms with distinct function symbols (Poseidon, Keccak). Addresses are Keccak-derived terms, not numbers. It is the only layer that reasons about an attacker.
2. A Stele specification (EquiVM's specification language) gives the same contract with concrete values (`address` is a 20-byte word). A new refinement proof shows that the Stele specification implements the Lean model. This proof maps each symbolic term to its concrete value, and it carries the one cryptographic assumption as a named hypothesis: no Keccak or Poseidon collision among the values a run uses. No existing tool provides this proof.
3. The bytecode runs on EVMLean (`AccountAddress = Fin 2^160`, transparent Keccak). EquiVM proves that the bytecode refines the Stele specification.

So the rewrite needs both proofs: Lean model → Stele (new) and Stele → bytecode (EquiVM).

**How each step is done (operator rulings, 2026-10-27).**

- **Lean model → Stele is a programmatic translation, not hand work.** The specification's operations are executable functions (`Option`-valued, `Bool` app rules). A translator emits the Stele functions and their refinement proof from them. Its one fixed input is the representation map, and that map is already known: Keccak over `abi.encode` for channel ids and state hashes (go-nitro `NitroUtils.sol`), Poseidon for note commitments, keys and nullifiers (the pool's circuits and library), 20-byte addresses, `uint256` amounts and `uint120` note values, the Solidity ABI and storage-layout rules. The specification omits these because their concrete form is already known. Where the specification departs from go-nitro (the outcome recorded at finalization, the channel kind, a channel's own notes, the payee note key), each needs one new field, chosen by the same conventions.
- **Stele → Solidity is transliteration** (Stele is "syntactically close to" Solidity, EquiVM `README.md`). solc compiles it.
- **Bytecode → Stele is proved by agents with EquiVM** (Lazaropoulos and Paraskevopoulou, arXiv 2607.26306: 35 production contracts proved end to end). EquiVM and EVMLean use the same toolchain as `armada-lean`, `leanprover/lean4:v4.33.1`.
- **The specification's shape is the translator's input,** so the operations must be small, composed and typed, and the proofs must meet the size and complexity criteria in `docs/lean/verification.md` §6a.

**Alternatives considered and rejected (2026-10-27).**

- **Verity** (LFG Labs, a Lean embedded contract language with a verified compiler to Yul; solc's Yul-to-bytecode step is not verified): a step backwards. It adds a young compiler to the trusted base and gives nothing EquiVM's bytecode-level check does not.
- **Writing the specification in Bend2:** no route is shorter than this chain (`labbook/2026-10-27-bend2-to-evm.md`). A Bend2 EVM backend exists (`ind-igo/bend-evm`), but it proves only contract ≡ IR ≡ Yul model and tests the rest, solc included. Exporting Bend2 to Lean is a research project (session `01a162aa`, 2026-10-25). Bend2 could later serve as an implementation language: EquiVM checks bytecode from any source, because Stele is source-language agnostic. The Lean model → Stele proof is needed in every route.

**Delivers:** a hardening delta on T0.2.

**Depends on:** S2 (needs P2b).

Chunk ids and points are copied from the 209-point list (`labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md` §7). N8 is withdrawn: there is no protocol swap fee (treasury lab book, decision 23 struck and §3a T9).

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| N1 | `MultiAssetHolder` on commitments | P2b (S2) | 11 | Sonnet (audit-critical) |
| N2 | `depositFromNote` as in-pool spend, payout as spend ~~including fee-tagged outputs to the Safe~~, tests | N1 | 10 | Sonnet (audit-critical) |
| N3 | Commitment-outcome finalization branch, tests | N2 | 8 | 5090 |
| N7 | E11 and E16 gates updated for in-pool collateral | N2, N3 | 8 | 5090 |
| ~~N8~~ | ~~Gate: in-channel swap fee → Safe-owned note → `cumulativeFees` → RevenueLock release~~ (withdrawn 2026-10-22) | — | 0 | — |

**Stage total:** 37 points, 4 sessions, 2.0 h.

**Acceptance:** N7 extends `nitro-e2e` with in-pool-collateral coverage of E11/E16. No new harness.

## S6 — armada-client Phase 1 (next)

**Goal:** one daemon runs the wallet and the relayer against the armada-nitro environment, end to end.

**Delivers:** the client's `config`, `contracts`, `wallet`, `relay` packages and its daemon/CLI (armada-client `docs/SPEC.md` §7 Phase 1, tasks P1-01 to P1-10).

**Depends on:** S1 (the armada-nitro environment's deployment manifest).

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| CL1-1 | Toolchain, workspace, `config` package, vendor `armada-sdk`, `contracts` package (P1-01–04) | — | 7 | 5090 |
| CL1-2 | `wallet` + `relay` packages wired to the armada-nitro environment's deployment manifest (P1-05–07) | CL1-1 | 9 | 5090 |
| CL1-3 | Daemon/CLI/HTTP server, the E2E relay gate, Linux distributable (P1-08–10) | CL1-2 | 9 | 5090 |

**Stage total:** 25 points, 3 sessions, 1.5 h.

**Acceptance:** armada-client's own E2E relay gate (P1-09) against the armada-nitro environment's fixturenet; this reuses the reconciler's deployment manifest (P1-05), not a new chain setup.

## S7 — Proof path (done, by hand; not yet a standing gate)

**Goal:** operate the nimbus verified proxy against the armada-nitro fixturenet and run the first proof.

**Delivers:** T1.0, T1.1, T1.2.

**Depends on:** S1.

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| T1-2 | Fixturenet genesis patch for the proxy's custom-network loader (`genesis_params.nim`, the metadata loader) | — | 5 | 5090 |
| T1-1 | Operate the native proxy on a server; run the first proof (`Shield` log inclusion + `rootHistory` slot) | T1-2 | 9 | 5090 |

**Stage total:** 14 points, 2 sessions, 1.0 h.

**Status.** Both chunks are done, by hand, not through the chunk/pebble process. The genesis patch is `AFDudley/nimbus-eth1` `armada/verified-proxy-fixes` at `b0205d1d` (base `master@2c8eabf`), alongside a `convHeader` fix at `827f95ff`. The fixturenet runs a non-validating `nimbus-eth1`/`nimbus-eth2` pair, merged to `armada-nitro` `integrated` at `990c9fa9` (pebble `arn-b58`). The lead ran the first proof by hand on 2026-10-20: `eth_getBlockReceipts` found the pool's `Shield` log; `eth_getStorageAt` on the `rootHistory[tree][root]` slot returned non-zero. The light-client source is nimbus-eth2, not lighthouse, because lighthouse v8.2.2's period-0 best update stays a pre-finality update on this fixturenet (`labbook/2026-10-20-nimbus-fixturenet-and-proxy-proof.md`).

**Acceptance, not yet met.** The same proof script, run as a step inside `armada-nitro`'s existing gates (alongside `nitro-e2e`, `pool-e2e`, `watchtower`), not run by hand. No chunk currently schedules this integration; it is the open item carried into S8.

## S8 — Client verified reads (next)

**Goal:** a client reads chain state through the proxy on every target, not a trusted RPC.

**Delivers:** T1.0 (client targets), T6.1 (verified reads).

**Depends on:** S7.

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| T1-3 | Armada verification library: slot keys for `rootHistory`/`nullifiers`, receipt scan for pool/armada-nitro logs | T1-1 | 7 | 5090 |
| T1-4 | Fix the browser WASM build's stall at its first `sync()` call, then wire it into `armada-interface`'s note-scanner | T1-1 | 7 | 5090 |
| T1-5a | iOS cross-compile of `libverifproxy.a` — done | T1-1 | 9 | 5090 |
| T1-5b | Android cross-compile of `libverifproxy.a` (`-d:PREFER_HASHTREE_SHA256=false`) — done | T1-1 | 9 | 5090 |

**Stage total:** 32 points, 4 sessions, 2.0 h.

**Status.** T1-5a and T1-5b are done: `libverifproxy.a` builds for iOS (device and simulator) and Android, outside the chunk/pebble process (`labbook/2026-10-19-nimbus-verified-proxy-proofs.md` §7). No client repo links either build in yet, so T1-3 and T1-4 remain open.

**Acceptance:** `armada-interface`'s note-scanner verifies a read through the WASM proxy in a browser test; `libverifproxy.a` links into a minimal iOS/Android test harness. No armada-nitro gate covers client-library builds; this is new verification alongside the existing gates.

## S9 — armada-client Phase 2 (next)

**Goal:** the client opens and settles armada-nitro channels over Waku, with a watchtower and a hub role.

**Delivers:** T6.2 (channel lifecycle), T6.3 (client watchtower), T2.3 (transport).

**Depends on:** S6 (the daemon).

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| CL2-1 | Waku transport package; relay-over-Waku signed fee announcements (P2-01, 02) | — | 7 | 5090 |
| CL2-2 | Vendor `ts-nitro` (Node 18 build); adapt it to `ArmadaNitroAdjudicator` (custodian-aware signature check via `isActiveDelegate`) (P2-03, 04) | CL2-1 | 9 | 5090 |
| CL2-3 | `WakuMessageService`; watchtower; hub role (P2-05–07) | CL2-2 | 9 | 5090 |
| CL2-4 | Acceptance e2e for the contract-layer paths (P2-08) | CL2-3 | 7 | 5090 |

**Stage total:** 32 points, 4 sessions, 2.0 h.

**Acceptance:** CL2-4 runs against the armada-nitro environment; it extends the `watchtower` gate for the watchtower path and `nitro-e2e` for the channel-lifecycle path.

## S10 — Client support for contract-owned notes and ZK channel apps (next)

**Goal:** the wallet proves, attaches and watches the new contract-owned and hidden-balance artifacts.

**Delivers:** T6.1 (prover), T6.2 (state updates), T6.3 (watchtower proofs).

**Depends on:** S2 (P4), S3 (N5), S5 (N3), S9 (ts-nitro vendored).

Chunk ids and points are copied from the 209-point list (`labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md` §7); C6 also depends on S4's G3.

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| C1 | Prover integration for the new artifacts: contract-owned variant, ledger, game | P4 (S2), N5 (S3) | 7 | 5090 |
| C2 | State update with proof: generate, attach, verify counterparty, persist state + preimage + proof | C1 | 10 | Sonnet |
| C3 | Watchtower stores and submits proofs | C2 | 5 | 5090 |
| C4 | ts-nitro: ZK app and commitment-outcome finalization | N3 (S5), N4 (S3) | 9 | 5090 |
| C5 | Gate: hidden-balance `waUSDC` swap through the mobile client on the fixturenet | C2–C4, N7 (S5) | 10 | Sonnet |
| C6 | Wallet: `ShieldedRagequit` flow with on-device proof | G3 (S4), C1 | 6 | 5090 |

**Stage total:** 47 points, 6 sessions, 3.0 h.

**Acceptance:** C5 extends `nitro-e2e` with the hidden-balance `waUSDC` swap through the mobile client on the fixturenet.

## S11 — armada-client Phase 3: mobile (later)

**Goal:** the wallet runs on iOS and Android, inside the daemon's Node runtime.

**Delivers:** T6.5 (mobile transport), T6.6 (native mobile proving).

**Depends on:** S9.

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| CL3-1 | React Native shell with the nodejs-mobile bridge (relay package shipped, disabled by default) (P3-01) | — | 7 | 5090 |
| CL3-2 | `leveldown-nodejs-mobile` behind `abstract-level` (P3-02) | CL3-1 | 4 | 5090 |
| CL3-3 | Mobile prover (P3-03) — [INFERENCE] whether Railway's `native-prover` proves Armada's circuit set is unverified (SPEC.md §8) | CL3-1 | 9 | 5090 |
| CL3-4 | Android 16 KB page-size compliance of the nodejs-mobile build (P3-04) — blocked on an upstream nodejs-mobile release | CL3-1 | 4 | 5090 |
| CL3-5 | Mobile UI (P3-05) | CL3-1 | 9 | 5090 |

**Stage total:** 33 points, 4 sessions, 2.0 h.

**Acceptance:** [INFERENCE] armada-client's own mobile acceptance path against the armada-nitro environment; no named gate covers the mobile build in the sources this plan cites.

## S12 — Watcher federation and the optional cache (later)

**Goal:** a client accepts a snapshot a quorum of watchers signs, as an optional cache, never a trust boundary.

**Delivers:** T2.4, T2.2 (optional).

**Depends on:** S7.

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| T2-1 | Watcher-signed snapshot cache; federation attestation (T2.4) | T1-3 (S8) | 9 | 5090 |
| T2-2 | `watcher-ts` as an optional unverified server cache (T2.2); optional, deferred | T1-1 (S7) | 4 | 5090 |

**Stage total:** 13 points, 2 sessions, 1.0 h.

**Acceptance:** [INFERENCE] no existing gate covers federation; a client accepts a snapshot CID signed by a quorum of watchers, then checks the commitment part against `rootHistory` through the T1 proxy.

## S13 — The production Phase-2 ceremony (later)

**Goal:** every deployed verifying key traces to a published, multi-party transcript and a final randomness beacon.

**Delivers:** completes T0.1.

**Depends on:** S2 (P4), S3 (N5) — the ceremony runs once over the final shape set, including the new shapes these stages add, so it is not re-run per shape (ADR-0003 D3.4).

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| CER-1 | Coordinate and run the multi-party Phase-2 ceremony over the final shape set; publish the transcript and the beacon | P4 (S2), N5 (S3) | 10 | Sonnet — [INFERENCE] sizes the tooling/verification half only; the ceremony's real duration is external-participant-bound and is not covered by these worker-hour rules (`estimation-method.md` §5) |

**Stage total:** 10 points, 1 session, 0.5 h (engineering half only).

**Acceptance:** a published, multi-party transcript and a final randomness beacon (ADR-0003 D3.3); `snarkjs zkey verify` over every shape.

## S14 — Venue v1 (later)

**Goal:** a posted-price dealer market clears trades over the existing settlement rail; a posted price has nothing to front-run.

**Delivers:** T4.0, T4.1, T4.3.

**Depends on:** S8 (verified reads), S10 (the ForceMove-app client pattern).

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| T4-1 | Posted-price contract (T4.0) | — | 9 | 5090 |
| T4-2 | Quote/settle ForceMove app (T4.1) | T4-1 | 7 | 5090 |
| T4-3 | Fee-split allocation wiring (T4.3) | T4-2 | 4 | 5090 |
| T4-4 | Gate: quote → settle → fee-split on the fixturenet | T4-1–T4-3 | 9 | 5090 |

**Stage total:** 29 points, 4 sessions, 2.0 h.

**Acceptance:** T4-4 extends `nitro-e2e` with the venue scenario; no new harness.

## S15 — v2 / deferred backlog (later)

**Goal:** the items ADR-0007/0011 defer to v2, plus items the registry marks not built with no near-term dependency.

**Delivers:** T3.0–T3.2, T4.2, T4.5, T4.6, T5.0c, T5.1, T0.4, T0.5, T0.6, T6.4 (audit), T6.8 (optional).

**Depends on:** S14 for the venue-adjacent items (T4.2, T4.5, T4.6, T3); S1 for the rest.

Every chunk below is self-sized from `estimation-method.md` §2's anchors; none comes from the 209-point list. [INFERENCE] throughout — these are not scheduled, and a real brief would re-estimate from the code at hand.

| Chunk | Deliverable | Depends on | Pts | Route |
|---|---|---|---:|---|
| T3-1 | Commit-reveal sequencer + randomness beacon (T3.0) | — | 9 | 5090 |
| T3-2 | Epoch set-agreement / DA (T3.1) | T3-1 | 10 | Sonnet |
| T3-3 | Sequencing-cert / inclusion-receipt attestation (T3.2) | T3-1 | 7 | 5090 |
| T4-5 | Venue solver: cross-venue internalization, flash-loan hedge, Aave rebalance (T4.2) | S14 | 10 | Sonnet |
| T4-6 | LP-buffered USDC yield rail (T4.5) | S14 | 9 | 5090 |
| T4-7 | ex_net matcher + LP vault (T4.6) | T3-1–T3-3, S14 | 10 | Sonnet |
| T5-1 | Swaps adapter contract (T5.0c) | — | 9 | 5090 |
| T5-2 | Routing interface for a future T4 venue (T5.1) | S14 | 4 | 5090 |
| T0-4 | Registry + bond contract (T0.4) | — | 9 | 5090 |
| T0-5 | Sequencing-cert + fraud-proof verifier (T0.5) | T0-4 | 10 | Sonnet |
| T0-6 | Native channelized commitment / fork-lite circuit (T0.6) | — | 9 | 5090 (circom) |
| T6-4 | Audit `armada-interface` for address rotation / loyalty-key support (T6.4) | — | 4 | 5090 |
| T6-8 | Optional Self zk-passport identity attestation (T6.8) | — | 7 | 5090 |

**Stage total:** 107 points, 12 sessions, 6.0 h.

**Acceptance:** each item's own eventual scenario on the armada-nitro environment, per its tier doc; none is scheduled, so no gate extension is named yet.

## S16 — External audit, mainnet cutover (later)

**Goal:** the finished v1 rail passes external audit and replaces the existing armada-1 mainnet deployment.

**Delivers:** the fixturenet-before-mainnet rule, for every tier this plan touches.

**Depends on:** S7 (fixturenet proof proven), S13 (ceremony, for audit scope); [INFERENCE] realistically also S4, S5, S10, since mainnet should carry the new governance and collateral design, not the superseded one.

No engineering chunk remains for this stage. Sepolia is currently out of scope (operator, 2026-10-20), so the plan has no Sepolia re-run of the proof.

**Stage total:** 0 points. External audit and the mainnet deploy are coordination/operational work, not engineering chunks; not sized (`estimation-method.md` §5).

**Acceptance:** proven on the fixturenet, then mainnet (CHUNK-RULES.md rule 4); external audit of `ArmadaNitroAdjudicator`, `HubCustodian` and the pool; the mainnet deploy supersedes the existing armada-1 crowdfund + governance + CCTP-only deployment.

## Plan totals

| | Points | Sessions | 5090 hours |
|---|---:|---:|---:|
| S2 | 30 | 4 | 2.0 |
| S3 | 23 | 3 | 1.5 |
| S4 | 55 | 8 | 4.0 |
| S5 | 37 | 4 | 2.0 |
| S6 | 25 | 3 | 1.5 |
| S7 | 14 | 2 | 1.0 |
| S8 | 32 | 4 | 2.0 |
| S9 | 32 | 4 | 2.0 |
| S10 | 47 | 6 | 3.0 |
| S11 | 33 | 4 | 2.0 |
| S12 | 13 | 2 | 1.0 |
| S13 | 10 | 1 | 0.5 |
| S14 | 29 | 4 | 2.0 |
| S15 | 107 | 12 | 6.0 |
| S16 | 0 | 0 | 0.0 |
| **Total** | **487** | **61** | **30.5** |

**Critical path, 209-point programme chunks (`labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md` §8, `labbook/2026-10-18-treasury-governance-baal-safe-shielded.md` §7, as split in S2 and S4):** P1 → P2a → P2b → G3 → G7a → G7b — 6 chunks, 6 sessions, 3.0 h, before rework.

**Critical path, this plan [INFERENCE, the longest stage chain]:** S2 (contract-owned notes) → S5 (armada-nitro collateral) → S10 (client support) — 4 + 5 + 6 = 15 sessions, 7.5 h, before the venue (S14) and the v2 backlog (S15). S4 (governance) runs off the same S2 fork and finishes sooner (4 + 8 = 12 sessions), so it does not extend the critical path.

## Pinned refs

`armada-poc` `origin/main@102f17b8` · `armada-circuits` `ship-armada.github.com/main@ef2682f` · `armada-deployments` `ship-armada.github.com/main@bac3639` · `armada-sdk` `origin/main@db5757f` · `armada-client` `origin/main@78a8fdc` · `armada-relayer` `origin/main@bbf5e62` · `armada-interface` `origin/main@6968cd7` · `armada-nitro` `origin/integrated@990c9fa9` · `status-im/nimbus-eth1` `master` (clone `4df52ef`, 2026-10-19) · `AFDudley/nimbus-eth1` `armada/verified-proxy-fixes@b0205d1d` (base `master@2c8eabf`).

---

Cross-refs: registry/status → [§5](./05-building-block-view.md); decisions → [§9](./09-architecture-decisions.md); risks/long-poles → [§11](./11-risks-and-technical-debt.md); solution strategy and the sequence at a high level → [§4](./04-solution-strategy.md); estimation method → [`estimation-method.md`](./estimation-method.md); live state → [`notes/handoff-2026-10-20.md`](../notes/handoff-2026-10-20.md).
