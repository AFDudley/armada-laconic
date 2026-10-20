# Build Plan — status and remaining work

status companion · 2026-10-20 · reconciled to §5 registry, §9 ADRs, §11 risks

The [Building Block View (§5)](./05-building-block-view.md) is the structure and the [ADRs (§9)](./09-architecture-decisions.md) are the decisions; this page is the **status and sequencing** view over the same `T#.#` items, read from the ship-armada code (`armada-poc`, `armada-circuits`, `armada-nitro`, `armada-sdk`, `armada-client`, `armada-relayer`, `armada-interface`). Governance contracts, added to §5 as built, are not repeated here. Full audit: `labbook/2026-10-19-docs-versus-code.md`.

## What exists today

The pool, the circuits, the adjudicator, and the client SDK are built. The plan that remains is the proof path (T1/T2), the client's mobile and verification build (T6), and the venue (T3/T4), which has no code.

### T0 · Ethereum anchor — built, mostly

- **T0.0 shielded pool — built.** `PrivacyPool` routes to four modules by `delegatecall`, with no proxy. 50 bps deployed shield fee. `armada-poc@102f17b8`.
- **T0.1 circuits — partly built.** `Transact(N,M)`, 19 shapes, compiled. The production Phase-2 ceremony has not run; only the unsafe dev script exists. `armada-circuits@ef2682f`; `v1.0.0-ceremony` tag.
- **T0.2/T0.3 adjudicator and deposit/payout — partly built, merged.** `ArmadaNitroAdjudicator` extends a patched `go-nitro`, and carries the deposit/payout boundary inside itself (`depositFromNote`, `SHIELD` payout). No separate boundary contract. `armada-nitro@55939ef`.
- **T0.4, T0.5 — not built.** Registry/bond and fraud-proof verifier have no code.
- **T0.6 — not built.** Native channelized commitment is an optional, deferred circuit change.
- **T0.7 — strategy, not code.** The crowd grows from the pool's own deposits. The wallet's move from Railgun (T6.7) is not built.
- **Governance and revenue — built.** `ArmadaGovernor`, `ArmadaTreasuryGov`, `TreasurySteward`, `ArmadaToken`, `ArmadaCrowdfund`, `ArmadaWindDown`, `ArmadaRedemption`, `AdapterRegistry`, `RevenueCounter`, `RevenueLock`, `RevenueReserveDistributor`.

### T1/T2 · Proof path — in progress, replanned

`labbook/2026-10-19-nimbus-verified-proxy-proofs.md` replans this tier. The nimbus verified proxy, not a state-diff emitter, is the proof path (ADR-0009). The table below is that entry's work-item list, in the docs' points.

| Item | Work | Replaces | Points |
|---|---|---|---:|
| WI-A | Operate the native proxy on servers | T1.0, T1.1 | 2 |
| WI-B | Browser WASM build in the JS client | client tier | 3 |
| WI-C | iOS and Android build of `libverifproxy.a` | near T6.5 | 5 |
| WI-D | Armada verification library: slot keys, receipt scan | T1.2, T6.1 | 3 |
| WI-E | Execution data node behind the proxy (archive) | T1.0 | 3 |
| WI-F | Beacon light-client source for each chain | T1.0 | 4 |
| WI-G | Snapshot cache and watcher attestation | T2.4 | 5 |
| WI-H | `watcher-ts` as an optional server cache | T1.2 | 2 |

The ingestion tier falls from 15 points (the earlier T1.0/T1.1/T1.2 estimate) to about 12 (WI-A, WI-D, WI-E, WI-F), and is no longer the v1 long pole. The cost moves to the client tier: WI-B (3) and WI-C (5). The fixturenet needs one more patch: the proxy's genesis loader hard-codes `sepolia`/`hoodi` and needs a custom-network path for the armada-nitro fixturenet's lighthouse genesis ([INFERENCE] 1 to 2 points).

### T3/T4 · Ordering and venue — not built

No posted-price contract, no quote/settle app, and no venue solver exist in `armada-poc`. These stay a plan. T4.4 (ETH yield as a shielded note) needs no separate contract, since any non-blocklisted ERC-20 can already be shielded.

### T5 · Adapters — partly built

CCTP and the Aave-v4 yield adapter are built, the Aave path through the pool's own `adaptContract`/`adaptParams` binding, not a Railgun RelayAdapt/Cookbook recipe (no RelayAdapt contract exists in this codebase). Swaps has no contract.

### T6 · Client / apps — the web app is built; mobile is not

`armada-client`, meant to host a React Native wallet, has no application code. The live client is `armada-interface`, a React 19 + Vite web app, wired to `armada-sdk`. It drives shield/transact calls; it does not drive an armada-nitro channel lifecycle. Note scanning and Groth16 proving run in-browser (`sdk-prover.ts`, a `snarkjs` Web Worker). No native mobile prover and no mobile transport exist; `armada-client` `docs/SPEC.md` names the four targets (Linux server, browser, iOS, Android) as the plan.

## Long poles (ranked, §11)

1. **Phase-2 ceremony** — T0.1: the shapes are open, with no beacon and no published transcript.
2. **Mobile build of `libverifproxy.a`** — WI-C: no upstream iOS or Android target exists yet.
3. **Fixturenet patch for the proxy** — the custom-genesis loader needs a patch before the first proof can run on the fixturenet.
4. **Mobile client** — T6.5: `armada-client` has no code; the four-target stack is still a spec.
5. **Audit surface** — T0.0, T0.1, T0.2/T0.3: external audit gates mainnet, as it did before.
6. **The venue** — T3, T4: no code exists; this is still a plan, not a near-term build.

## Sequenced backlog

1. **Proof path.** Operate the native proxy on a server against the armada-nitro fixturenet (WI-A, WI-E, WI-F), after the fixturenet genesis patch. Run the first proof: shield on the fixturenet, prove the `Shield` log's inclusion and the `rootHistory` slot through the patched proxy (`labbook/2026-10-19-nimbus-verified-proxy-proofs.md` §3.8).
2. **Phase-2 ceremony.** Run the production MPC over the 19 `Transact` shapes and publish the transcript and beacon.
3. **Client verification.** WASM proxy in the browser (WI-B), the Armada verification library (WI-D), and a `watcher-ts` server cache (WI-H) if an unverified cache is wanted.
4. **Mobile.** Cross-compile `libverifproxy.a` for iOS and Android (WI-C); build `armada-client`'s four-target stack against it.
5. **Snapshot attestation.** The watcher-signed snapshot cache (WI-G, T2.4), once several independent watchers are running.
6. **Venue.** T3/T4, once the rail above is live. No code exists today; this item has no shortcuts from existing work.

## Pinned refs

`armada-poc` `origin/main@102f17b8` · `armada-circuits` `ship-armada.github.com/main@ef2682f` · `armada-deployments` `ship-armada.github.com/main@bac3639` · `armada-sdk` `origin/main@db5757f` · `armada-client` `origin/main@78a8fdc` · `armada-relayer` `origin/main@bbf5e62` · `armada-interface` `origin/main@6968cd7` · `armada-nitro` `origin/integrated@55939ef` · `status-im/nimbus-eth1` `master` (clone `4df52ef`, 2026-10-19).

---

Cross-refs: registry/status → [§5](./05-building-block-view.md); decisions → [§9](./09-architecture-decisions.md); risks/long-poles → [§11](./11-risks-and-technical-debt.md); solution strategy → [§4](./04-solution-strategy.md).
