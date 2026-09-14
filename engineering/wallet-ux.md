# Wallet User Experience — browser & mobile

UX description · 2026-09-10 · maps to the **T6 client tier** (§5) · settlement rail per §6

This describes what a person actually does and sees in the Armada wallet, on a browser and on a phone, and which building blocks make each surface work. It is the user-facing companion to the arc42 runtime view (§6) and the cross-chain swap construction (`shielded-nitro-bridge-design.md`). Guiding constraint: **a phone is a full non-custodial peer, not a thin client of a server** (ADR-0008).

## 1. Platforms — one wallet, two shells

The wallet logic is identical across platforms; only the shell and transport differ (T6.5, ADR-0008).

| | Browser | Mobile |
|---|---|---|
| Shell | PWA / WebView carrying the in-browser `ts-nitro` stack | Native app (React Native) over a **native gomobile** transport (go-waku + go-libp2p-noise), T6.5 |
| Custody | BIP-39/HD keys; WebCrypto/OS keystore | Keys in the **device secure enclave**; never leave it (T6.0, TC-5) |
| Note-scanner | WASM in a worker (T6.1) | WASM behind a WebView↔RN **key bridge** (T6.1) |
| Proving | Groth16 in WASM (snarkjs) | on-device Groth16 (snarkjs-WASM / rapidsnark), T6.6 |
| Transport | Waku pub/sub + libp2p-noise; interim WebView relay | native libp2p peer; same interface (T6.5) |

The interim WebView stack ships first (spine demo); the native module lands behind the *same* transport interface the settlement client (T6.2) and watchtower (T6.3) already bind to, so screens don't change when it arrives.

## 2. Onboarding & custody

- **Create / import.** Standard BIP-39 seed; keys derived and held in the platform secure enclave (T6.0). The seed is shown once for backup and never leaves the device; the app cannot exfiltrate it.
- **No accounts, no server login.** There is no Armada account. The wallet is peer software — it joins the transport (T2.3), scans, and transacts. Nothing to sign up for.
- **What the user sees:** a shielded balance per asset, a receive address (rotating — T6.4), and actions: **Add funds**, **Send**, **Swap / Earn**, **Withdraw**.

## 3. Core surfaces

### 3.1 Shielded balance (read)
The balance is computed **on-device** by the WASM note-scanner (T6.1) from a **proof-carrying feed** (T2.0): the wallet reads `getStorageAt → {value, proof}` slices over the transport, verifies each proof against the L1 state root, and trial-decrypts locally to find its notes. **No `eth_getStorageAt`/`eth_getLogs` ever leaves the device to a public RPC** — the act of checking your balance does not deanonymize you (ADR-0006). Every subscriber to a slice gets byte-identical bytes, so the feed provider can't learn which notes are yours.

### 3.2 Add funds / shield-in
Bring an ERC-20 (e.g. USDC) into the shield: the wallet builds a shield transaction that mints a fresh shielded note (T0.0 `shield`). **This is the one moment an amount is public** (Design A, ADR-0005). The deposit path checks the POI allow-list root before admitting value (gated entry, ADR-0006). Cross-chain inbound (e.g. CCTP) arrives the same way — as a fresh shielded note.

### 3.3 Send / private transfer
A shielded payment: a `transact` with `unshield=NONE` (T0.0). Sender, recipient, and amount are hidden in-circuit; only nullifiers and output-commitment hashes are public. The user picks a recipient shielded address and amount; the wallet proves on-device (T6.6) and submits — optionally via the **submit-on-behalf keeper** so no EOA of yours appears on-chain (write-side unlinkability, §8.4).

### 3.4 Swap & earn (same-chain)
Cleared over Nitro at a **posted price** with nothing to front-run (T4, ADR-0007). Two shapes the user won't distinguish visually:
- **Yield (mint/redeem):** USDC → aUSDC or ETH → wstETH is a protocol deposit, no counterparty needed; the wallet can self-serve the adapter recipe, or route through an LP for extra privacy (the ADR-0011 direct-vs-LP-buffered choice) — either way the user just taps **Earn**.
- **Swap (A↔B):** filled from venue inventory at the posted quote (T4.0/T4.1), settled non-custodially over the T0.3 rail.

### 3.5 Cross-chain swap (the fronted, instant experience)
This is the surface the `shielded-nitro-bridge-design.md` construction powers (ADR-0015/0016). The user picks **source chain + asset** and **destination chain + asset**, sees an instant quote, and taps **Swap**. From the user's side it feels like an instant swap; underneath:
- the wallet funds a channel to a hub from a shielded note (amortized — reused across swaps),
- the hub **fronts** the destination asset from standing inventory, bound by an HTLC so the advance is atomic,
- the user receives the destination asset **immediately** as a fresh shielded note on the destination chain.

The user never performs an on-chain cross-chain crossing, so nothing links the two sides. Latency is one round-trip, not a bridge-settlement wait. If the user opts into the **Nym privacy mode** (below), the same swap runs with IP/metadata hidden at a small latency cost.

### 3.6 Withdraw / exit
**Unshield-out** to a public address, or **force-exit** a channel. The happy path is a cooperative close that shields the outcome back into fresh notes (T0.3 payout). If a counterparty or hub misbehaves or goes silent, the wallet's **self-watchtower** force-closes at the latest signed state and re-shields — funds are never trapped (ADR-0004, T6.3). The user is guaranteed to be able to exit at the last state they signed.

## 4. Services the user relies on but never operates

- **Note scanning (T6.1)** runs continuously in the background over private feeds; new incoming notes appear without any query the user issues.
- **Self-watchtower (T6.3)** watches for adversarial force-closes and submits a higher-turn checkpoint inside the challenge window. On a phone that can sleep through that window, the user is offered a one-tap **delegate to a keeper**: an always-on relayer that watches and checkpoints on their behalf, learning *that* a dispute occurred but not who they are or the channel contents (§8.4 write-side privacy). This is the "never lose funds while offline" affordance.
- **Transport (T2.3)** — the unified Waku + libp2p-noise transport, shared by every service. Invisible.
- **Proving (T6.6)** — shield/unshield/transfer proofs are generated on-device; the user sees a brief "preparing" state. Mobile proving cost is a real budget (T6.6 constraint), so the wallet batches and pre-proves where it can.
- **Metering (T2.1)** — per-read payments to watcher parties ride Nitro vouchers funded **at note creation**; the user never tops up a separate balance and the payment leaks nothing about what they read.

## 5. Privacy controls the user actually holds

- **Nym privacy mode (opt-in).** A toggle for the **optional Nym mixnet underlay** (ADR-0008): off by default for lowest latency; on to hide IP/metadata for maximum privacy. Uniform across every service the wallet uses, not a per-feature setting.
- **Address rotation / per-venue keys (T6.4).** Receive addresses rotate, and venue interactions use per-venue loyalty keys, so activity doesn't cluster under one identity.
- **Submit-on-behalf.** On-chain submissions can be relayed by a keeper so no EOA links the user to *writing* (write-side unlinkability), complementing the read-side privacy of §3.1.
- **What stays public:** exactly the shield/unshield **boundary amounts** (Design A, ADR-0005) — surfaced in the UI as "this amount is visible on-chain," never hidden from the user.

## 6. Browser vs mobile — the differences

- **Background execution.** A browser tab and a WebView can be suspended, which is why the watchtower's keeper-delegation matters most on mobile, and why the **native** module (T6.5) exists — it keeps a first-class libp2p peer alive for the challenge window. Until native ships, the WebView build leans on delegated keepers for liveness.
- **Proving cost.** Phones are slower at Groth16; the mobile UX shows proof progress and pre-proves opportunistically (T6.6).
- **Key custody.** Mobile uses the hardware secure enclave; browser uses the best available keystore. Keys never leave either.

## 7. Failure & edge UX (never surprising, never fund-losing)

- **Offline during a dispute** → the delegated keeper handles the checkpoint; the user is notified after the fact.
- **Hub/LP runs dry** → quotes widen or withdraw (out-of-protocol operator behavior, ADR-0015/§6); the user simply sees "no quote right now," never a stuck swap.
- **Counterparty vanishes mid-flow** → HTLC timeout + unilateral exit refunds the user to a fresh note; shown as "swap reverted, funds returned."
- **Proof fails / cancelled** → nothing was submitted; balance unchanged.

## 8. What the wallet never does

- Never custodies funds with a third party; never lets a hub, keeper, or venue take value (liveness-only trust, unilateral exit).
- Never queries a public RPC for reads (feeds + local verify only).
- Never exposes the seed or lets the app move keys off the device.
- Never hides the one leak (boundary amounts) from the user.

---

Cross-refs: client-tier items → [§5](./05-building-block-view.md) T6.0–T6.7, T2.0/T2.1/T2.3; runtime flows → [§6](./06-runtime-view.md); transport & privacy → [§8](./08-crosscutting-concepts.md); decisions → [§9](./09-architecture-decisions.md) ADR-0004/0005/0006/0008/0011/0015/0016; cross-chain swap detail → [the construction](./shielded-nitro-bridge-design.md).
