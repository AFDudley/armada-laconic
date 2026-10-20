# 8. Cross-cutting Concepts

arc42 §8 · 2026-10-20

Some concerns recur **across** tiers, tracing "lines" through the §5 registry. Each such concept is described **once** here. Its realizing building blocks are named by their `T#.#` id — the detail lives in the tier docs and is not repeated. Where tiers group the registry vertically, these concepts group it horizontally. Status in the table below is the concept's weakest linked item: a concept with any not-built item is marked accordingly, even where another of its items is built.

| # | Concept | Realized by | Status |
|---|---|---|---|
| 8.1 | Settlement model — nitro-on-railgun | T0.0, T0.2 | built |
| 8.2 | Privacy model | T0.0, T0.2, T0.6, T0.7 | partly built (plan for T0.6, T1 integration) |
| 8.3 | Non-custody & exit | T0.2, T6.3 | partly built (T0.2 built; T6.3 not built) |
| 8.4 | Transport | T2.3, T6.5 | not built |
| 8.5 | Metering | T2.1, T4.3 | partly built (different mechanism) |
| 8.6 | Trusted setup | T0.1 | in progress (ceremony open) |
| 8.7 | Identity (optional) | T6.8 | not built |

## 8.1 Settlement model — nitro-on-railgun

The core architectural idea is notes in → normal Nitro → notes out (ADR-0004). A shielded note is unshielded into `ArmadaNitroAdjudicator`'s `depositFromNote`, which funds a Nitro channel on the same contract. Arbitrary off-chain ForceMove games then play out, and on finalize — whether by cooperative close or by post-challenge force-close — the outcome is shielded back into fresh notes through a `SHIELD`-type allocation. This gives Nitro's optimistic exits and expressive settlement games over shielded value with no new cryptography. There is no separate deposit/payout contract; the code merges that boundary into the adjudicator itself (ruling 1).

The decisive property is **isolation**. The Nitro game analysis and the shielded-pool privacy analysis stay fully separate, meeting only inside `depositFromNote` and the SHIELD payout. Neither side has to reason about the other's internals. Armada's adapters do not use Railgun's RelayAdapt/Cookbook recipe pattern anywhere, in settlement or elsewhere: the Aave-v4 yield adapter calls the pool's own `adaptContract`/`adaptParams` binding instead (T5.0b).

The optional amount-privacy upgrade (T0.6, §8.2) would preserve this isolation exactly, changing only what the boundary opens — exposing commitments rather than cleartext balances — while the Nitro game still runs its normal challenge, checkpoint, and conclude flow unchanged. It is not built.

→ detail in §5 T0.0, T0.2; runtime sequences in §6.

## 8.2 Privacy model

The privacy model states what Armada hides, from whom it hides it, and where value becomes visible. Every other part of the system exists to defend it.

### Inside the pool, amounts are hidden

Value lives in Armada's own pool as commitments. The pool follows the Railgun design. A commitment is a hash that hides the note's owner and its amount. A zero-knowledge proof checks every spend. The proof shows that the spend conserves value and that each amount is in range, but it does not reveal the amounts. A large note therefore looks the same as a small one. The pool needs no fixed denominations, unlike Tornado Cash. This is built: `PrivacyPool` (T0.0) deploys the commitment tree, shield, and transact paths.

### Amounts become public only at the boundary

Hidden value becomes visible where it crosses from the pool to the public chain. In the design, this happens at the same boundary every time: inside `ArmadaNitroAdjudicator`, when a user shields or unshields, when a channel takes a deposit or makes a payout, or when a channel is force-closed.

A design could also hide the amounts inside a channel. The smaller one keeps the dispute game as it is. A finalized channel then lists its payouts as commitments instead of amounts, and the payout opens them into fresh notes. It changes the circuits, so it needs a new trusted-setup ceremony and a new audit, and it makes proving on phones slower. This is item T0.6, not built, and not scheduled.

A larger design moves the dispute game itself into a proof, so that every state change proves its own validity against hidden balances. It is out of scope.[^proven-dispute]

[^proven-dispute]: **Why a proven dispute game is out of scope.** ADR-0005 names the smaller design *fork-lite* (item T0.6) and the larger one *fork-full*, and calls the larger one research-grade. The reasons came out in a design session on 2026-10-18 (session `01a150cb`, 23:03–23:26 UTC):

    - Every rule of the dispute game that reads a balance must become a circuit constraint. The audited Nitro contracts would turn into a verifier plus a new circuit, with a new ceremony and a new audit.
    - Every party, the hub included, must prove every state change. Today a state change costs one signature.
    - Each participant must keep the hidden balances, and keep them available, or lose the ability to challenge a stale state.
    - A related variant attaches a complete payout proof to every signed state. The operator rejected it as "the worst of both worlds" (23:16 UTC). It pays for a proof on every update, and each state is still valid only by its signatures.

    Where hidden state is the point, as in a game with hidden information, the proof belongs in that game's rules. Those rules live in the game's `IForceMoveApp`, not in the dispute game. The dispute game then stays as it is, and only the game's moves carry proofs (23:18 UTC).

### Privacy grows with the crowd

The pool hides each spend among all the unspent notes that look alike. This crowd is the **anonymity set**. A spend is only as private as the crowd is large, so a new pool starts with almost no privacy and gains it with use.

The crowd grows only from deposits into Armada's own pool (ruling 5). A note in another pool cannot hide a spend here, so the crowd grows by bringing in new users; there is no cross-pool proof and no onboarding bridge.

The wallet also makes it easy for a Railgun user to move funds in. The move is a Railgun unshield followed by a shield into Armada's pool, so it is an ordinary deposit (T0.7, T6.7). It costs the user some privacy. The token, the amount and the time are public on both pools, so an observer can match the two sides. The moved value also leaves Railgun's large crowd for the smaller one. The wallet is meant to show these costs before the user confirms; this flow has no shipped code yet in `armada-interface` (T6.7).

### Selective disclosure shows where value came from

No check stands between a user and the pool. Deposits, transfers, withdrawals and channel deposits proceed without any proof of origin, and no Proofs of Innocence exist or are planned (ruling 4).

A user who must show where value came from, for example to an exchange, uses **selective disclosure**. This is a voluntary, off-chain proof of a chain of custody for chosen value. The chain runs from a shielding address, through notes and channels, to a payout or a withdrawal. The proof reveals nothing else about the wallet, and it never becomes a condition of pool use. The disclosure circuit exists only on unmerged branches: armada-circuits `disclosure-circuit` (`7175b46`, `src/operations/NullifierDisclosure.circom`) and armada-sdk `selective-disclosure` (`0aa2ba1`).

### Reading and writing leave no trace, as a goal

A user touches the pool in two ways. The user reads the pool to find their notes, and writes to it to submit a transaction. The design calls for different components to protect each side of the same crowd; today only part of this is built.

On the read side, the goal is that no RPC provider learns which commitments or nullifiers a user looks for. Today `armada-interface` reads over plain RPC and trusts the answer. The nimbus verified proxy (T1) proved a read against a beacon-anchored header by hand on the armada-nitro fixturenet, but no client integrates it yet.

On the write side, the goal is that a keeper or broadcaster submits transactions for the user, so no address of the user's own appears as the sender. Today `armada-interface` submits transactions directly from the user's own key; no broadcaster or address-rotation code exists (T6.5).

One limit remains even once built. Verifying a read does not hide a user's IP address from whichever server they ask. Section 8.4 covers transport, which is also not built.

*Items:* T0.0, T0.2, T0.6, T0.7, T1, T6.1, T6.5, T6.7. *Decisions:* ADR-0002, ADR-0005, ADR-0006, ADR-0009. *Design:* [16 selective disclosure](./16-selective-disclosure.md). *Quality scenarios:* §10. *Cold-start risk:* §11.

## 8.3 Non-custody & exit

No tier ever takes custody of user value, and correctness of settlement means a user can **always exit unilaterally**. Value is either a shielded note (T0.0) or an allocation in a Nitro channel the user co-signs (T0.2) — never held by a venue, watcher, or relay. A liveness failure therefore degrades to **halt, not loss**, at the contract level.

The exit guarantee rests on the ForceMove dispute machinery — `forceMove`, `respond`, and `checkpoint` on `ArmadaNitroAdjudicator`, owned in T0.2 and built — plus its counterpart, the **watchtower**, which supplies the automated response half. A watchtower is a read feed plus a bonded challenge responder. `armada-nitro`'s e2e gate harness has exactly this (`env/gates/watchtower/watch.ts`), but no client-resident watchtower exists (T6.3, not built). Today a user who notices a stale-state challenge must respond manually, within the window, by calling `respond`/`checkpoint` themselves; the contract-level guarantee holds regardless of whether an automated watchtower exists, but nothing watches on the user's behalf yet.

→ detail in §5 T0.2, T6.3; challenge flow in §6; this gap tracked in §11.

## 8.4 Transport

The design calls for two transports (ADR-0008): **Waku pub/sub** for feed discovery, gossip, and async delivery, and **libp2p-noise direct streams** for a 1:1 hot loop. Neither exists in any ship-armada repo: no Waku dependency, no libp2p dependency, and no code of either kind in `armada-client`, `armada-relayer`, or `armada-interface` (T2.3, T6.5, not built). `armada-interface` talks plain HTTP/RPC to Ethereum L1 and to `armada-relayer` today.

**IP privacy — the optional Nym underlay.** The design calls for an optional Nym mixnet underlay beneath whatever transport ships, uniform and opt-in, off by default for latency and on for maximum privacy. This remains a design note; no transport exists yet to underlay.

→ detail in §5 T2.3, T6.5; deployment of relays/peers in §7.

## 8.5 Metering

The design calls for reads and settlement to be paid the same way, with private per-request go-nitro payment vouchers that net into a co-signed virtual-channel allocation. This is not built. Today `armada-relayer`'s actor charges a flat USDC fee computed from a gas estimate (`fee-calculator.ts`), a conventional per-transaction charge, not a voucher stream (T2.1, built, a different mechanism). No go-nitro voucher code (`payments.ts`-equivalent) exists in `armada-poc`, `armada-relayer`, or `armada-interface`. The venue fee-split (T4.3) is not built, because no venue exists.

→ detail in §5 T2.1, T4.3; this gap tracked in §11.

## 8.6 Trusted setup

Groth16 needs a per-circuit setup, and the concept here is a **reuse-Phase-1 / own-Phase-2** discipline (ADR-0003), owned as a process — not code — in T0.1. Phase-1 reuses the community Perpetual Powers of Tau, the expensive, vetted, circuit-agnostic part; Armada never runs its own universal ceremony. Phase-2 is **Armada's own**: a circuit-specific MPC over its deployed set. The `v1.0.0-ceremony` tag's manifest lists every circuit shape as `"status": "open"`, with no beacon — the production ceremony has not finished. Only `scripts/setup-dev.sh` exists today, explicitly labeled unsafe for anything but local testing.

The key coupling is that any circuit change re-runs Phase-2 while Phase-1 is never re-run. The canonical trigger would be the T0.6 amount-privacy upgrade (§8.2), if it is ever greenlit.

→ detail in §5 T0.1; ceremony status in §11.

## 8.7 Identity (optional)

Identity is called out here explicitly to record that it is **not** cross-cutting-required. It is a single **optional** wallet parameter (T6.8, not built). A user may source a selective-disclosure attestation off-band — for example a Self zk-passport (nationality, age, OFAC-clear, personhood) — proven on-device and verified off-chain, anchored to an Ethereum address, with the shielded pool identity left unlinked. No such mechanism exists in the client today. It threads through no other tier, so it is listed as a concept only to fix its non-cross-cutting scope.

→ detail in §5 T6.8.

---

*Decisions behind these concepts: [§9 ADRs](./09-architecture-decisions.md). The building blocks they span: [§5 registry](./05-building-block-view.md). Quality scenarios that exercise them: [§10](./10-quality-requirements.md); open risks: [§11](./11-risks-and-technical-debt.md).*
