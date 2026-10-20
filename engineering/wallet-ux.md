# Wallet User Experience — web today; mobile a plan

UX description · 2026-10-20 · maps to the **T6 client tier** (§5)

This describes what a person does and sees in `armada-interface`, the live Armada web app, the one shipped client today. Mobile is a spec only; no mobile client exists.

## 1. Platforms — one web app today

| | Browser (built) | Mobile (not built) |
|---|---|---|
| Shell | `armada-interface`, a React 19 + Vite web app | `armada-client` has no application code; its `docs/SPEC.md` names Linux server, browser, iOS, and Android as the plan |
| Custody | Keys in browser memory (`keyManager.ts`) | A device secure enclave, not built |
| Note-scanner | In-browser (`sync.ts`) | Not built |
| Proving | `snarkjs` Web Worker (`sdk-prover.ts`), worker-first | A native mobile prover, not built |
| Transport | Whatever RPC `armada-interface` is configured against; no verified proof path yet | Not built |

## 2. Onboarding & custody (built)

- **Create / import.** The wallet generates and holds keys in browser memory. This is a web custody model, not BIP-39/HD/secure-enclave device custody.
- **No accounts, no server login.** The wallet talks to the pool and to the relayer's current indexer directly; there is no Armada account system.
- **What the user sees:** a shielded balance per asset, a receive flow, and actions: shield, transfer (shielded send), unshield, and the cross-chain shield/unshield path over CCTP.

## 3. Core surfaces, as built

### 3.1 Shielded balance (read) — built, not yet verified

The balance is computed in-browser by `sync.ts` from note scanning and trial decryption. Today this reads through whatever RPC or the relayer's Ponder-indexed `/v2/quick-sync` endpoint the client is configured against, not through a cryptographically verified proof path. The nimbus verified proxy (T1, ADR-0009) is the intended trust boundary and has proved reads by hand on the armada-nitro fixturenet; `armada-interface` does not use it yet, so a user's balance read today trusts the RPC or the relayer, not a client-side proof.

### 3.2 Add funds / shield-in — built

`armada-interface/src/features/shield/handler.ts` builds a shield transaction that mints a fresh shielded note (`PrivacyPool.shield`). This is the one moment an amount is public. No compliance check gates the deposit. Cross-chain inbound over CCTP arrives the same way, as a fresh shielded note (`PrivacyPool`'s CCTP message handlers).

### 3.3 Send / private transfer — built

`armada-interface/src/features/transfer-shielded/handler.ts` builds a `PrivacyPool.transact` call with `unshield == NONE`. Sender, recipient, and amount are hidden in-circuit; only nullifiers and output-commitment hashes are public. The wallet proves in-browser with the `snarkjs` Web Worker and submits. There is no submit-on-behalf relay for write-side origin privacy; the transaction is submitted from the user's own address.

### 3.4 Unshield / exit — built

`armada-interface/src/features/unshield/handler.ts` unshields to a public address. There is no channel to force-close, because no settlement-client channel lifecycle exists yet (T6.2, [T6-client-apps.md](./T6-client-apps.md)); the only exit path today is a direct pool unshield.

## 4. Surfaces that are not built

- **Swap and earn over a venue.** No venue (T4) exists. There is no posted-price swap or in-pool yield routing UI beyond the Aave-v4 yield adapter's own shield/redeem flow, where one exists.
- **Cross-chain swap through fronting hubs.** No code implements the hub-fronted instant-swap design (ADR-0015/0016 era). `armada-nitro`'s adjudicator and `HubCustodian` exist, but no client drives a channel through them.
- **Self-watchtower and the "never lose funds while offline" guarantee.** No client watchtower exists ([A.6](./A-nitro-on-railgun/A.6-settlement-client-watchtower.md)). A user cannot today rely on an automatic dispute response, because there is no channel lifecycle to dispute.
- **Address rotation and per-venue loyalty keys.** Not checked against the current client code; treat as unconfirmed rather than absent.
- **Mobile custody, mobile proving, and native transport.** None built; see §1.
- **A Nym or other IP-privacy mixnet mode.** No such transport exists in any client, relayer, or interface repo checked.
- **Move from Railgun.** The design in ADR-0006 and T6.7 is unchanged: a Railgun unshield followed by a shield here, with security costs the wallet would show before confirming. No such flow exists in `armada-interface` yet.

## 5. What stays true regardless of what ships

- The wallet never custodies funds with a third party, in the surfaces that exist today (direct pool shield, transfer, unshield are all user-signed).
- The one leak is the shield/unshield boundary amount. The UI should surface this; whether it does today was not checked in this rewrite.

---

Cross-refs: client-tier status → [§5](./05-building-block-view.md) T6, T6-client-apps.md; the proof path this client does not yet use → [T1-ingestion.md](./T1-ingestion.md); decisions → [§9](./09-architecture-decisions.md) ADR-0006, ADR-0009.
