# 8. Cross-cutting Concepts

arc42 §8 · 2026-09-04

Some concerns recur **across** tiers, tracing "lines" through the §5 registry. Each such concept is described **once** here. Its realizing building blocks are named by their `T#.#` id — the detail lives in the tier docs and is not repeated — and its governing decision by `ADR-####`. Where tiers group the registry vertically, these concepts group it horizontally.

| # | Concept | Realized by | Governing ADR |
|---|---|---|---|
| 8.1 | Settlement model — nitro-on-railgun | T0.2, T0.3, T0.6 | ADR-0004 |
| 8.2 | Privacy model | T0.0, T0.3, T0.6, T0.7 | ADR-0005, ADR-0006 |
| 8.3 | Non-custody & exit | T0.2, T0.3, T6.3 | ADR-0004 |
| 8.4 | Transport | T2.3, T6.5 | ADR-0008 |
| 8.5 | Metering | T2.1, T4.3 | — |
| 8.6 | Trusted setup | T0.1 | ADR-0003 |
| 8.7 | Identity (optional) | T6.8 | — |

## 8.1 Settlement model — nitro-on-railgun

The core architectural idea is notes in → normal Nitro → notes out (ADR-0004). A shielded note is unshielded into the **deposit/payout contract** (T0.3), which funds a vanilla `go-nitro` channel (T0.2). Arbitrary off-chain ForceMove games and vouchers then play out, and on finalize — whether by cooperative close or by post-challenge force-close — the outcome is shielded back into fresh notes. This gives us Nitro's optimistic exits and expressive settlement games over Railgun-style value with no new cryptography and without governance-gated RelayAdapt modules.

The decisive property is **isolation**. The Nitro game analysis and the Railgun privacy analysis stay fully separate, meeting only at the **allocation→note boundary** (T0.3). Neither side has to reason about the other's internals; the boundary is the single place where the two models are stitched together. RelayAdapt, Railgun's own recipe, appears only inside the venue solver's atomic hedge leg (T4.2) — **never** in settlement.

The optional amount-privacy upgrade (T0.6, §8.2/ADR-0005) preserves this isolation exactly. It changes only what the allocation→note boundary opens, exposing commitments rather than cleartext balances; the Nitro game still runs its normal challenge, checkpoint, and conclude flow unchanged.

→ detail in §5 T0.2, T0.3; runtime sequences in §6; deep dive in [`A.0 nitro-on-railgun overview`](./A-nitro-on-railgun/A.0-overview.md) and [`shielded-nitro-bridge-design.md`](./shielded-nitro-bridge-design.md).

## 8.2 Privacy model

The privacy model states what Armada hides, from whom it hides it, and where value becomes visible. Every other part of the system exists to defend it.

### Inside the pool, amounts are hidden

Value lives in our own pool as commitments. The pool follows the Railgun design. A commitment is a hash that hides the note's owner and its amount. A zero-knowledge proof checks every spend. The proof shows that the spend conserves value and that each amount is in range, but it does not reveal the amounts. A large note therefore looks the same as a small one. The pool needs no fixed denominations, unlike Tornado Cash.

### Amounts become public only at the boundary

Hidden value becomes visible where it crosses from the pool to the public chain. In the design we ship, this happens at three points:

- when a user shields or unshields;
- when a Nitro channel takes a deposit or makes a payout;
- when a channel is force-closed.

Two designs could also hide the amounts inside a channel. The smaller one keeps the dispute game as it is. A finalized channel then lists its payouts as commitments instead of amounts, and the payout opens them into fresh notes. It changes our circuits, so it needs a new trusted-setup ceremony and a new audit, and it makes proving on phones slower. It is an optional later upgrade, not part of version 1.

The larger design moves the dispute game itself into a proof, so that every state change proves its own validity against hidden balances. It is out of scope.[^proven-dispute]

[^proven-dispute]: **Why a proven dispute game is out of scope.** ADR-0005 names the smaller design *fork-lite* (item T0.6) and the larger one *fork-full*, and calls the larger one research-grade. The reasons came out in a design session on 2026-10-18 (session `01a150cb`, 23:03–23:26 UTC):

    - Every rule of the dispute game that reads a balance must become a circuit constraint. The audited Nitro contracts would turn into a verifier plus a new circuit, with a new ceremony and a new audit.
    - Every party, the hub included, must prove every state change. Today a state change costs one signature.
    - Each participant must keep the hidden balances, and keep them available, or lose the ability to challenge a stale state.
    - A related variant attaches a complete payout proof to every signed state. The operator rejected it as "the worst of both worlds" (23:16 UTC). It pays for a proof on every update, and each state is still valid only by its signatures.

    Where hidden state is the point, as in a game with hidden information, the proof belongs in that game's rules. Those rules live in the game's `IForceMoveApp`, not in the dispute game. The dispute game then stays as it is, and only the game's moves carry proofs (23:18 UTC).

### Privacy grows with the crowd

The pool hides each spend among all the unspent notes that look alike. This crowd is the **anonymity set**. A spend is only as private as the crowd is large, so a new pool starts with almost no privacy and gains it with use.

The crowd grows only from deposits into our own pool. A note in another pool cannot hide a spend here, so we grow the crowd by bringing in new users.

The wallet also makes it easy for a Railgun user to move funds in. The move is a Railgun unshield followed by a shield into our pool, so it is an ordinary deposit. It costs the user some privacy. The token, the amount and the time are public on both pools, so an observer can match the two sides. The moved value also leaves Railgun's large crowd for our smaller one. The wallet shows these costs before the user confirms.

### Selective disclosure shows where value came from

No check stands between a user and the pool. Deposits, transfers, withdrawals and channel deposits proceed without any proof of origin.

A user who must show where value came from, for example to an exchange, uses **selective disclosure**. This is a voluntary, off-chain proof of a chain of custody for chosen value. The chain runs from a shielding address, through notes and channels, to a payout or a withdrawal. The proof reveals nothing else about the wallet, and it never becomes a condition of pool use.

### Reading and writing leave no trace

A user touches the pool in two ways. The user reads the pool to find their notes, and writes to it to submit a transaction. Different components protect each side of the same crowd.

On the read side, a watcher serves every user the same bytes, each with a proof that the bytes are correct. The user's device scans them locally. So no RPC provider learns which commitments or nullifiers a user looks for.

On the write side, a keeper or broadcaster submits transactions for the user. So no address of the user's own appears as the sender. A fresh address for each channel also stops one boundary crossing from being linked to the next.

One limit remains. The transport encrypts traffic and authenticates peers, but it does not hide a user's IP address. Section 8.4 covers the transport.

*Items:* T0.0, T0.3, T0.6, T0.7, T2.0, T6.1, T6.4, T6.5, T6.7. *Decisions:* ADR-0002, ADR-0005, ADR-0006, ADR-0008, ADR-0020. *Design:* [16 selective disclosure](./16-selective-disclosure.md). *Quality scenarios:* §10. *Cold-start risk:* §11.

## 8.3 Non-custody & exit

No tier ever takes custody of user value, and correctness of settlement means a user can **always exit unilaterally**. Value is either a shielded note (T0.0) or an allocation in a Nitro channel the user co-signs (T0.2) — never held by a venue, watcher, or relay. A liveness failure therefore degrades to **halt, not loss**.

The exit guarantee rests on the ForceMove dispute machinery — `forceMove`, `respond`, and `checkpoint` on `NitroAdjudicator`, owned in T0.2 — plus its counterpart, the **watchtower**, which supplies the response half. This is the **watcher/watchtower duality**: a watchtower is a watcher's read feed (T2 freshness) **plus** a bonded challenge responder (T0.2). The phone-resident watchtower (T6.3) composes the two. It watches for a stale-state force-close over the T2 feed and submits the user's higher-turn co-signed state within the challenge window, so a dead counterparty freezes trading but never funds. T6.3 **gates on T2 feed freshness**: a stale feed is treated as a liveness alarm — fall back to a redundant party or direct submit — not a silently missed window.

→ detail in §5 T0.2, T6.3; challenge/watchtower flow in §6; "never miss a valid challenge" scenario in §10.

## 8.4 Transport

We run two transports, each chosen for what it does best (ADR-0008). **Waku pub/sub** handles feed discovery, gossip, and async delivery; it is recipient-unlinkable within a content-topic's anonymity set, and it is the transport Railgun's broadcaster network already speaks. **libp2p-noise direct streams** handle the 1:1 hot loop — voucher and quote exchange, low-latency reads — that Waku's multi-hop store-and-forward cannot meet at tens-of-ms latency. A circuit-relay plus STUN/TURN bootstrap is the only fixed transport infrastructure. This substrate is owned by T2.3.

On mobile the same two transports are a **port, not a new protocol** (T6.5). An interim WebView browser stack carries the spine demo, followed by a production **native gomobile** module (go-waku plus go-libp2p-noise) sitting behind the same transport interface that the settlement client (T6.2) and watchtower (T6.3) bind to. The js-libp2p stack is not React-Native-compatible, which is why the native port is required rather than optional polish (§11 crux).

**IP privacy — the optional Nym underlay.** Noise provides encryption and peer-id auth **only — not IP privacy**; endpoints otherwise see each other's IPs. IP/metadata privacy is an **optional Nym mixnet underlay** that any Armada service or user can run beneath T2.3/T6.5 — the *same* underlay everywhere, uniform and opt-in, off by default for latency and on for maximum privacy. It is not a bespoke per-feature transport.

→ detail in §5 T2.3, T6.5; deployment of relays/peers in §7.

## 8.5 Metering

Reads and settlement are **paid the same way**: private per-request **go-nitro payment vouchers** that net into a co-signed virtual-channel allocation, settled to L1 in USDC. On the read path (T2.1), paid feed methods are metered with vouchers **funded at note creation**, so paying for reads leaks nothing about the reader. On the write/settle path (T4.3), the venue fee is **just a fee-split in the channel allocation** — user, protocol, and integrator shares on the co-signed channel state, with no fee vault and no negotiation — netted on close. Both paths reuse the identical go-nitro voucher/virtual-channel mechanism (`payments/vouchers.go`, `protocols/virtualfund/virtualfund.go`), the same one T0.2/T0.3 already publish. Because metering is non-custodial, the worst case is refused service, never loss. These two flows are the project's two revenue pillars — **watcher metering + venue spread**.

→ detail in §5 T2.1, T4.3; runtime metering step in §6.

## 8.6 Trusted setup

Groth16 needs a per-circuit setup, and the concept here is a **reuse-Phase-1 / own-Phase-2** discipline (ADR-0003), owned as a process — not code — in T0.1. Phase-1 reuses the community Perpetual Powers of Tau, the expensive, vetted, circuit-agnostic part; we never run our own universal ceremony. Phase-2 is **ours**: a circuit-specific MPC over our deployed set, with **1-of-N honest** security (≥5 organizationally distinct contributors), a published transcript, and a final randomness beacon. The Solidity verifier is generated from the final vkey and wired into T0.0.

The key coupling is that any circuit change re-runs Phase-2 while Phase-1 is never re-run. The canonical trigger is the T0.6 amount-privacy upgrade (§8.2). New `wasm` and `zkey` artifacts then ship to the mobile prover (T6.6). This is why fork-lite carries a ceremony, fresh-audit, and heavier-prover cost (ADR-0005): nothing new is invented, the same pipeline is simply re-executed.

→ detail in §5 T0.1; prover constraint in §5 T6.6; deferral rationale in §11.

## 8.7 Identity (optional)

Identity is called out here explicitly to record that it is **not** cross-cutting-required. It is a single **optional** wallet parameter (T6.8), default off. A user may source a selective-disclosure attestation off-band — for example a Self zk-passport (nationality, age, OFAC-clear, personhood) — proven on-device and verified **off-chain, anchored to an Ethereum address**, with the shielded pool identity left unlinked. The wallet stores it as one more optional signing input, surfaced only when a T4 venue's match filter requires it. It threads through no other tier and is deferred rather than built, so it is listed as a concept only to fix its non-cross-cutting scope.

→ detail in §5 T6.8; per-venue loyalty-key policy in §5 T6.4.

---

*Decisions behind these concepts: [§9 ADRs](./09-architecture-decisions.md). The building blocks they span: [§5 registry](./05-building-block-view.md). Quality scenarios that exercise them: [§10](./10-quality-requirements.md); open risks: [§11](./11-risks-and-technical-debt.md).*
