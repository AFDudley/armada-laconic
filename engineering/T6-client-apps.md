# T6 · Client / apps

**Status:** partly built · tier doc · 2026-10-20
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T6
**Depends on / Blocks:** depends on T0.0 (pool, circuit artifacts), T0.2 (adjudicator, the deposit/payout boundary), T1 (the proof path a verifying client would use).

`armada-client`, the repo meant to host a React Native wallet, has no application code: 45 files, all docs and config, no `src/`. The live client is `armada-interface`, a React 19 + Vite web app wired to `armada-sdk`, independently written, not a fork of any Laconic repo. `armada-client` `docs/SPEC.md` (`origin/main`) records the plan for a unified TypeScript stack on four targets: Linux server, browser, iOS, Android (D1: Node 18.20.4 everywhere; D6: `ts-nitro` with a Waku transport; D7: a Waku light node; D8: LevelDB through `abstract-level`). None of that plan has shipped code yet.

## T6.0 Wallet base — built, web only

`armada-interface` holds keys in browser memory (`src/lib/shielded/keyManager.ts`), not a BIP-39/HD custody chain in a device secure enclave. There is no React Native code, no `react-native-keychain`, and no mobile build. This is a real, working custody path for a web wallet.

## T6.1 Note-scanner + prover — built, in the browser

`armada-interface/src/lib/shielded/sync.ts` scans and trial-decrypts notes in the browser. `armada-interface/src/lib/shielded/sdk-prover.ts` and `armada-sdk@db5757f:src/prover/index.ts` run a `snarkjs` Web Worker Groth16 prover, worker-first with a same-thread fallback for tests, against `armada-circuits`' build artifacts (SHA-256 manifest-checked). No native mobile prover exists.

**What is not yet true.** The scanner reads chain state over whatever RPC `armada-interface` is configured with, not yet through the T1 nimbus verified proxy. Verified note sync (ADR-0009) is still a plan for this client, not a shipped behavior.

## T6.2 Settlement client — partly built

`armada-interface/src/features/{shield,transfer-shielded,unshield}/handler.ts` drive `PrivacyPool` shield and transact calls. No code drives an `armada-nitro` channel lifecycle: no open/fund-from-note, propose/accept state, cooperative close, or force-close exists in `armada-interface` or `armada-sdk`. `armada-sdk`'s own grep for `nitro`/`Nitro`/`HubCustodian` across its `src/` returns nothing.

## T6.3 Self-hosted watchtower — not built (client); built (test harness only)

`armada-nitro`'s e2e gate harness has a watchtower (`env/gates/watchtower/watch.ts`, run by the `armada-watchtower-e2e` and `armada-watchtower-discrimination-e2e` jobs). No client-side watchtower exists anywhere in `armada-interface` or `armada-sdk`.

## T6.4 Address rotation + loyalty keys — not checked

No inventory read covered this item in either `armada-interface` or `armada-sdk`. State it as not checked, not as absent, until a targeted read confirms either way.

## T6.5 Mobile transport — not built

`armada-client` `docs/SPEC.md` is the plan: one TypeScript codebase for wallet, relayer, and watchtower, on Linux server, browser, iOS, and Android, using `ts-nitro` over a Waku transport. No mobile code exists, interim or native. `libverifproxy.a`, the proof path's mobile library (T1), has no upstream iOS or Android build either, so even a mobile client that existed today could not yet verify.

## T6.6 Groth16 mobile proving — partly built

A browser Web Worker prover exists (T6.1). No native mobile (rapidsnark/mopro) prover exists in any repo checked across the client, relayer, interface, SDK, and nitro repos.

## T6.7 Armada-branded app + SDK wiring — built (app); move from Railgun not built

`armada-interface` is the live, Armada-branded web app, wired to `@armada/sdk` (`armada-sdk`). It is a real, shipping product surface, not a plan.

**The wallet move from Railgun does not exist yet.** The design is unchanged from ADR-0006: a Railgun unshield followed by a shield into Armada's pool, a wallet feature, no contract reads or touches Railgun's pool. The security costs the wallet must show before the user confirms: the token, the amount and the time are public on both pools; the unshield recipient is public; the value leaves Railgun's larger crowd for Armada's smaller one; a later exit can link back; Railgun charges its unshield fee. None of this flow has shipped code in `armada-interface` yet.

## T6.8 Identity — not built

No optional zk-passport or similar attestation exists in the client.

## Sources

- `armada-client` `origin/main@78a8fdc` — `docs/SPEC.md` (the four-target plan; zero lines of `src/`).
- `armada-interface` `origin/main@6968cd7` — `src/lib/shielded/{sync.ts,keyManager.ts,sdk-prover.ts}`, `src/features/{shield,transfer-shielded,unshield}/handler.ts`.
- `armada-sdk` `origin/main@db5757f` — `src/prover/index.ts` and the rest of `src/`.
- `armada-nitro` `origin/integrated@55939ef` — `env/gates/watchtower/watch.ts`.
- `labbook/2026-10-19-docs-versus-code.md`, `labbook/2026-10-19-unified-stack-and-verified-note-sync.md`.
- Site docs: [`05-building-block-view.html`](./05-building-block-view.html), [`T1-ingestion.html`](./T1-ingestion.html), [`09-architecture-decisions.html`](./09-architecture-decisions.html#adr-0006).
