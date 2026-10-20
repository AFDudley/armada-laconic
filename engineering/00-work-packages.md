# 0. Work Packages (WBS)

delivery frame · **2026-10-20** · operating model: [ADR-0012](./09-architecture-decisions.md#adr-0012)

This is the work breakdown: the deliverable-oriented decomposition of the programme. The four work packages are MECE over the whole Armada product ([ADR-0013](./09-architecture-decisions.md#adr-0013)). Each owns a set of `T#.#` items from the [§5 registry](./05-building-block-view.md); most of work package A's items, and parts of C's, are already built.

## The work packages

| WP | Mission | Owns (`T#.#`) | Status |
|---|---|---|---|
| **A · on-chain anchor** | Non-custodial settlement of shielded value: the pool, the circuits, the adjudicator, and governance. | T0.0–T0.7; the deposit/payout boundary (merged into T0.2) | Mostly built |
| **B · proof path** | A client verifies Ethereum state against a header a beacon light client checks, through the nimbus verified proxy. | T1.0–T1.2, T2.0–T2.4; note-scanner T6.1 | In progress, upstream, unintegrated |
| **C · venue and adapters** | Let users earn yield and swap, over a future venue; the adapters are already live. | T3.\*, T4.\* (plan); T5.0, T5.1 | Adapters partly built; venue not built |
| **D · client** | The wallet/app. | T6.0, T6.2–T6.8 | The web app is built; mobile is a spec only |

## What is built, per work package

- **A.** `armada-poc` built the pool and its governance and revenue contracts. `armada-circuits` built the `Transact` circuit. `armada-nitro` built the patched adjudicator and the merged deposit/payout boundary. The Phase-2 ceremony has not run.
- **B.** The nimbus verified proxy exists upstream, unintegrated. No server runs it against any Armada chain; no WASM bundle ships; no mobile build exists.
- **C.** CCTP and the Aave-v4 yield adapter are built (`armada-poc`). No venue, no matcher, and no Swaps contract exist.
- **D.** `armada-interface` (web) is built and wired to `armada-sdk`. `armada-client`, meant to host a mobile client, has no application code; its `docs/SPEC.md` is a plan for four targets (Linux server, browser, iOS, Android).

## What remains

1. **The Phase-2 ceremony.** No shape has a published, multi-party transcript yet.
2. **The proof path.** Operate the nimbus proxy on a server; patch it for the fixturenet's genesis; ship a WASM build in the browser client; cross-compile for mobile.
3. **A settlement-client channel lifecycle and a client watchtower.** Neither exists in `armada-interface` or `armada-sdk`.
4. **Mobile.** `armada-client`'s four-target plan has no code yet.
5. **A venue.** No code exists for T3 or T4; this is the least mature part of the plan.

## Interface contracts, as built

- **The adjudicator exposes:** `IArmadaPool` (`shield`, `transact`), `depositFromNote`, the `SHIELD`-type payout path, go-nitro's standard challenge/checkpoint/conclude.
- **The pool exposes:** `shield`, `transact`, `atomicCrossChainUnshield`, the commitment-insert and nullifier-spend events, `rootHistory`.
- **The proof path, once integrated, exposes:** verified `{value, proof}` reads and verified receipts, through the nimbus proxy, to any client.
- **The client exposes:** `armada-interface`'s shield/transfer/unshield handlers, over `@armada/sdk`.

## Sequencing, revised

1. **The on-chain anchor is already built**, so it is no longer the first thing to land; it is the baseline everything else now targets.
2. **The proof path is the next concrete target**, not the long pole it used to be: operate the proxy, patch the fixturenet genesis, and run the first proof (`labbook/2026-10-19-nimbus-verified-proxy-proofs.md` §3.8).
3. **The client's settlement lifecycle and watchtower** are the next client-side gap, once a verified read path exists to build them against.
4. **A venue** has no code and no near-term dependency forcing it; it stays last.

## Relationship to the arc42 docs

The arc42 set is the architecture description; this WBS is the delivery frame around it. The tier docs (`T0`–`T6`) carry the building-block detail and the real status for each item.
