# A.1 — Reuse Inventory (the deep dive)

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md) · **Feeds:** A.2–A.9

This is the grounded inventory of what already exists for work package A, with pinned file/line citations, what each piece gives us, which A-item it serves, and the net-new deltas. Per [ADR-0012](../09-architecture-decisions.md#adr-0012), specs cite these rather than re-documenting them. **Provenance:** go-nitro **@435eb2b**, ts-nitro **@884d616**, mobymask **@2329198** are pinned; **Railgun** repos were read at `master`/`main` HEAD (2026-09-05) and **must be pinned to a commit before build** (A.2/A.3/A.5).

**Clean-room reframe ([ADR-0014](../09-architecture-decisions.md#adr-0014)).** Railgun's on-chain pool and `circuits-v2` are unlicensed, so T0.0/T0.1 are **not** reuse — our engineers **clean-room reimplement** them, **spec-compatible** with the Railgun design pinned here. Accordingly, **A.1.1 (pool) and A.1.2 (circuits) are the *reference spec* the clean-room implements, not "reuse as-deployed OSS"**; this whole inventory doubles as that reference spec. `go-nitro`/`ts-nitro` (A.1.3/A.1.5) and Railgun's MIT `engine`/`cookbook` remain genuine OSS reuse, unaffected.

---

## A.1.1 Shielded pool (T0.0) — clean-room reference spec

This is the **reference spec** T0.0 clean-room reimplements — a **spec-compatible** shielded pool, independently authored under no Railgun license ([ADR-0014](../09-architecture-decisions.md#adr-0014)); the citations below pin the exact behavior/format our contracts must match, they are **not** reused-as-deployed OSS. Pool = upgradeable proxy → `RailgunSmartWallet` is `RailgunLogic` is `{Commitments, TokenBlocklist, Verifier}`. **Only two public mutating entrypoints**; there is no separate deposit/withdraw — deposit/withdraw *are* shield and unshield-via-transact.

| Piece | Citation (`Railgun-Privacy/contract` @HEAD) | Gives us / A-item |
|---|---|---|
| `shield(ShieldRequest[])` | `contracts/logic/RailgunSmartWallet.sol` L23–97 | **T0.3 payout** entrypoint: mints fresh commitments; emits `Shield`. |
| `transact(Transaction[])` | `RailgunSmartWallet.sol` L102–224 | **T0.3 deposit**: a tx with `boundParams.unshield=NORMAL` unshields ERC20 out; emits `Transact`,`Nullified`. |
| `transferTokenOut` (unshield recipient = `address(npk)`) | `RailgunLogic.sol` L318–364 | Confirms **deposit target**: set `unshieldPreimage.npk = T0.3 address` to escrow into the channel. |
| `changeFee(shield,unshield,nft)` | `RailgunLogic.sol` L146–161 | **T0.0 fee=0 lever**: call `changeFee(0,0,0)`. |
| Events `Action/Transact/Shield/Unshield/Nullified` | `RailgunLogic.sol` L57–77 | Commitment-insert + nullifier-spend feed **B (T1.2)** indexes and **T6.1** scans. |
| Merkle accumulator + nullifier set (`TREE_DEPTH=16`, `rootHistory`) | `Commitments.sol` L28–55, L108–252 | The tree **T6.1** rebuilds and **T0.3** proofs reference a historical root of. |
| Groth16 verifier registry `verificationKeys[nIn][nOut]`, `setVerificationKey` | `Verifier.sol` L27, L36–45; `Snark.sol` L157–188 | **T0.0/T0.1** register one vkey per circuit combo after the ceremony. |
| ABI structs (`ShieldRequest`,`CommitmentPreimage`,`TokenData`,`Transaction`,`BoundParams`,`SnarkProof`) | `Globals.sol` L21–122 | Exact encode surface for **T0.3** + **T6.2**. |
| `TokenBlocklist` | `TokenBlocklist.sol` L33–73 | T0.0 leaves USDC unblocked; unshield always allowed. |

**POI** is **not** in these contracts — it is an alongside partner system (A.1.7). The pool-side levers are `changeFee(0,0,0)` and vkey registration; everything else is client/settlement-layer policy.

## A.1.2 JoinSplit circuits & trusted setup (T0.1) — clean-room reference spec

This is the **reference spec** T0.1's circuits are **authored clean-room** against — independently written, **spec-compatible** with the JoinSplit behavior below, under no `circuits-v2` license ([ADR-0014](../09-architecture-decisions.md#adr-0014)). The citations pin the note/commitment/nullifier format and public-signal arity our own circuits must reproduce; they are the design we match, not source we reuse.

| Piece | Citation (`circuits-v2` @HEAD) | Gives us |
|---|---|---|
| JoinSplit circuit (the payments primitive) | `src/library/joinsplit.circom` L11–118 | Public signals `merkleRoot,boundParamsHash,nullifiers[],commitmentsOut[]`; proves EdDSA ownership, Merkle membership, nullifier, range `<2^120`, `sumIn===sumOut`. Defines note/commitment/nullifier format **T6.1/T0.3** must match. |
| Nullifier derivation | `src/library/nullifier-check.circom` L11–13 | `nullifier = Poseidon(nullifyingKey, leafIndex)`. |
| Circuit-combo generator | `lib/circuitConfigs.js` | Generates **91** `(nIn,nOut)` combos (`nIn+nOut ≤ 14`). **⚠️ reconcile vs the cited "~54" registered subset** (A.3). |
| Ceremony driver | `scripts/prepare_ceremony` L23–79 | Phase-1 = reuse `powersOfTau28_hez_final_20.ptau` (2²⁰); Phase-2 = per-circuit zkey from r1cs+ptau. Confirms ADR-0003 shape. |
| Trusted-setup docs | docs.railgun.org/…/trusted-setup-ceremony | Groth16 + Perpetual PoT Phase-1 + per-circuit Phase-2; new ceremony on any circuit change. |

## A.1.3 go-nitro adjudicator (T0.2) — reuse as-is

`NitroAdjudicator = ForceMove + MultiAssetHolder + exit-format Outcome`.

| Piece | Citation (`cerc-io/go-nitro` @435eb2b) | Gives us / A-item |
|---|---|---|
| `deposit(asset, channelId, expectedHeld, amount)` | `packages/nitro-protocol/contracts/MultiAssetHolder.sol` L36–70 | **The escrow entrypoint T0.3 calls** after unshield; `holdings[asset][channelId]` L24–29; `expectedHeld` reorg guard. |
| `concludeAndTransferAllAssets(FixedPart, candidate)` | `NitroAdjudicator.sol` L33–42; `transferAllAssets` L123–190 | **Finalize + payout**; loops all assets → payouts to destinations. |
| `_executeSingleAssetExit` / `_isExternalDestination` | `MultiAssetHolder.sol` L411–427 | External destinations get funds transferred out → **T0.3 payout** is an external destination that re-shields. |
| `challenge` / `checkpoint` / `conclude` | `ForceMove.sol` L39–80 / L88–119 / L126–172 | Dispute state machine; **`checkpoint` (higher-turn, no finalize) is the exact primitive T6.3 uses** to defeat a stale challenge. |
| `recoverVariablePart` (signing domain) | `ForceMove.sol` L236–268 | State sigs = `NitroUtils.hashState(fixedPart, variablePart)` — **T0.3/T6.2 must match**. |
| Outcome encoding (`SingleAssetExit`,`Allocation`,`AllocationType{simple,withdrawHelper,guarantee}`) | exit-format `ExitFormat.sol` L26–73 | The wire format **T0.3 ABI-encodes** for payout. |
| `HashLockedSwap.sol` reference app | `examples/HashLockedSwap.sol` L33–83 | Pattern for a ForceMove app — **but single-asset/2-party only** (A.1.9 delta). |
| multi-asset swap negotiation | `protocols/swap/swap.go` L105–408 | Confirms **multi-asset ETH-in/USDC-out IS supported off-chain** (corrects our "single-asset gap"). |
| vouchers | `payments/vouchers.go` L23–52 | Incremental in-channel micropayment primitive (reused by C metering + settlement). |

## A.1.4 The deposit/payout binding (T0.3) — how the two systems wire

T0.3 is a public-side ERC20 counterparty; it touches pool + adjudicator **only** through the ABIs above:

- **Deposit-in:** `RailgunSmartWallet.transact([...])` with `boundParams.unshield=NORMAL` and `unshieldPreimage.npk = bytes32(uint160(T0.3))` → `transferTokenOut` sends ERC20 to T0.3 → T0.3 `IERC20.approve` + `MultiAssetHolder.deposit(asset, channelId, expectedHeld, amount)`. `channelId = NitroUtils.getChannelId(fixedPart)`.
- **Payout-out:** channel finalizes via `concludeAndTransferAllAssets`; outcome names T0.3 as an external destination (or uses an `AllocationType.withdrawHelper`); T0.3 receives the ERC20 and `RailgunSmartWallet.shield([ShieldRequest...])` mints fresh notes (built like engine `ShieldNote.serialize`, `Railgun-Community/engine` `src/note/shield-note.ts` L46–121).
- **App:** the channel's `appDefinition` (`NitroAdjudicator.sol` L193–207 → `IForceMoveApp.stateIsSupported`) is a plain settlement app for the walking skeleton; C's real quote/settle app registers here.

## A.1.5 Settlement client (T6.2)

| Piece | Citation | Gives us |
|---|---|---|
| In-browser Nitro node | `cerc-io/ts-nitro` @884d616 `packages/nitro-node/src/node/node.ts` L69–190 | The client engine (fund/defund/pay). **Dispute API thinner than go-nitro** — no challenge/checkpoint surfaced (A.1.9). |
| Browser P2P over `@cerc-io/peer` | ts-nitro `…/p2p-message-service/service.ts` L16–120, 453–527 | `/nitro/msg/1.0.0` + `/nitro/peerinfo/1.0.0`; relay/WebRTC dial (no inbound listen). |
| Reference libp2p host (server side) | go-nitro `node/engine/messageservice/p2p-message-service/service.go` L36–46, 105–140 | noise/tcp/ws + DHT `scaddr→peerID`. |
| **Submit-on-behalf keeper** (origin-private submission) | `cerc-io/mobymask` @2329198 `packages/server/index.ts` L48–58; `Delegatable.sol` `execute`/`_msgSender` | The **broadcaster analog**: a funded relayer submits on behalf of a signer who stays origin-private — reusable *shape* for T6.2 submission (write-side privacy). |
| Invocation build/sign + OpenRPC client (iframe/postMessage) | mobymask `packages/react-app/src/reportPhishers.js`; `packages/client/typescript/src/index.ts` | Templates for a wallet-hosted (iframe) submission client. |

## A.1.6 Self-watchtower (T6.3) — partly reusable, core is net-new

| Reusable | Citation (go-nitro @435eb2b) |
|---|---|
| Event source (`ChallengeRegistered`/`ChallengeCleared`) | `node/engine/chainservice/eth_chainservice.go` L494–539; `chainservice.go` L86–160 (`ChallengeRegisteredEvent{candidate,sigs,FinalizesAt,IsInitiatedByMe}`) |
| Tx submit (`Challenge`/`Checkpoint`) | `eth_chainservice.go` L376–386; payloads `protocols/interfaces.go` L59–114 |
| Manual trigger | `engine.go` L882–915 (`handleCounterChallengeRequest`) |

**Net-new (confirmed):** there is **no automatic watchtower**. On an adversarial `ChallengeRegistered`, a non-initiator engine auto-creates a `directdefund` (conclude+withdraw), **not** a higher-turn `checkpoint` (`engine.go` L540–612). Only a *manual* `CounterChallengeRequest` reaches checkpoint/challenge. T6.3 = the missing **watch → compare turnNum → auto-submit higher-turn checkpoint before `FinalizesAt`** loop, **plus** always-on liveness (browser node has no background watch loop / guaranteed relay connection). See A.6.

## A.1.7 Anonymity set (T0.7) & the payments primitive

- **Shielded payments** = a `transact()` with `unshield=NONE`: inputs nullified, outputs created, `sumIn==sumOut` in-circuit; sender/recipient/amount hidden (only root, boundParamsHash, nullifiers, output-commitment hashes are public). This is the base capability A inherits for free.
- **POI** (docs.railgun.org/…/private-proofs-of-innocence): an **alongside** partner system, not in the pool. On shield, an Unshield-Only Standby Period (default 1h) precedes a blinded non-inclusion ZK proof vs list-provider datasets; proofs carry forward through transfers. **T0.0 POI "config" = choose list providers + standby at the client/POI-node layer; T0.3 gated-entry enforces a valid POI at the settlement layer.** The POI node stack is a **separate dependency** — not redeployable from the four repos scoped here.

## A.1.8 Corrections to earlier docs (the deep dive caught these)

1. **mobymask `@2329198` is NOT a browser-peer/watcher.** It is classic delegatable MobyMask (hardhat + react-app + OpenRPC server) — no libp2p, no watcher loop. It gives the **submit-on-behalf keeper** only. The browser-peer/watcher lineage lives in `mobymask-v2` / `mobymask-v2-watcher-ts` / `@cerc-io/peer`. → Fix T2/T6/§8 citations that attribute the P2P substrate or "browser-peer app" to `@2329198`.
2. **"go-nitro outcomes are effectively single-asset" is overstated.** `MultiAssetHolder`/`transferAllAssets`/`swap.go` support multi-asset; the *gap* is that no shipped ForceMove **app** does multi-asset atomic swap (HashLockedSwap is single-asset). Restate the §11/T0.2 maturity gap accordingly.
3. **Circuit count is 91, not 54.** Reconcile the registered subset (A.3); don't assert 54.
4. **Railgun citations are unpinned.** Pin a commit before build.

## A.1.9 Net-new deltas (what A actually builds)

1. **T0.3 deposit/payout contract** — no Railgun/Nitro equivalent (RelayAdapt is atomic-multicall, not channel-escrow). Deposit-in, `MultiAssetHolder.deposit`, outcome→external-destination, re-shield. (A.5)
2. **Multi-asset ForceMove settlement app** — for ETH-in/USDC-out atomicity; HashLockedSwap is single-asset only. (A.4/A.5; C authors the real quote/settle variant.)
3. **T6.3 auto-watchtower loop + liveness** — watch `ChallengeRegistered` → submit higher-turn `checkpoint` before `FinalizesAt`; always-on node. (A.6)
4. **`changeFee(0,0,0)` governance action + vkey registration** for the chosen combo subset. (A.2/A.3)
5. **POI policy wiring** at the settlement layer + integrating the separate POI node stack. (A.7)
6. **In-browser dispute API** — ts-nitro exposes only fund/defund/pay; add the challenge/checkpoint surface. (A.6)
7. **(opt) native-ETH / native-commitment** — pool is ERC20/721 only; needs a WETH wrap or T0.6. (A.9)

## A.1.10 Licensing & risks

- **✅ LICENSING — RESOLVED via clean-room ([ADR-0014](../09-architecture-decisions.md#adr-0014)).** `circuits-v2` `License.md`: *"No License is provided for any party under any circumstances"*; the pool contracts are SPDX `UNLICENSED`. Redeploying them was never executable — so **T0.0 (pool) and T0.1 (circuits) are clean-room reimplemented** by our engineers, spec-compatible with the design pinned in A.1.1/A.1.2, using no Railgun-licensed source. This turns the single largest risk from an **open blocker into a resolved gate** and reclassifies T0.0/T0.1 as **audit-critical net-new**. `Railgun-Community/engine` and `cookbook` remain **MIT** reference; `go-nitro`/`ts-nitro` are unaffected OSS reuse.
- **Unpinned Railgun commit** — pin before build.
- **91-vs-54 circuit ambiguity** — decide the registered subset (drives ceremony count).
- **`snarkSafetyVector`/`checkSafetyVectors`** magic constants must be **reproduced verbatim in our clean-room implementation** (matching the reference `RailgunLogic.sol` L111–127) or `transact` reverts.
- **POI is a separate partner stack**, not redeployable from these repos.
- **ERC1155 unsupported** by the pool (bounds T0.6).

---

→ Item specs build on this: [A.2](./A.2-pool-deployment.md) · [A.3](./A.3-trusted-setup.md) · [A.4](./A.4-adjudicator-integration.md) · [A.5](./A.5-deposit-payout-contract.md) · [A.6](./A.6-settlement-client-watchtower.md) · [A.7](./A.7-anonymity-set.md) · [A.8](./A.8-interfaces-acceptance.md) · [A.9](./A.9-native-commitment.md).
