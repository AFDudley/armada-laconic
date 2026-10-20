# A.6 — Settlement client & self-watchtower

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md)
**Owns:** T6.2 settlement client (partly built) · T6.3 self-watchtower (not built)

---

## A.6.1 Status

**T6.2, the settlement client, is partly built.** `armada-interface/src/features/{shield,transfer-shielded,unshield}/handler.ts` drives `PrivacyPool` shield and transact calls. No code anywhere in `armada-interface` or `armada-sdk` drives an `armada-nitro` channel lifecycle: no fund-from-note, propose/accept state, cooperative close, or force-close. `armada-sdk`'s own search for `nitro`/`Nitro`/`HubCustodian` across its `src/` returns nothing.

**T6.3, the self-watchtower, is not built.** `armada-nitro`'s e2e gate harness has a watchtower (`env/gates/watchtower/watch.ts`), run by the `armada-watchtower-e2e` and `armada-watchtower-discrimination-e2e` jobs, but it is test-harness infrastructure, not a client the wallet ships. No client-side watchtower exists.

## A.6.2 What exists to build on

The adjudicator side of this boundary is built and tested ([A.4](./A.4-adjudicator-integration.md), [A.5](./A.5-deposit-payout-contract.md)): `depositFromNote`, the `SHIELD`-type payout branch, `go-nitro`'s `challenge`/`checkpoint`/`conclude`, and the per-channel reentrancy lock. `armada-nitro`'s own Forge tests already exercise the watchtower-relevant cases from the contract side: a stale hub challenge overridden by the user's latest state, a revoked delegate's states settling during the revocation delay, a hostile contract participant unable to exhaust a counterparty's challenge, and per-asset payout checks under combined-versus-per-asset settlement (`WatchtowerPerAssetPayoutCheck.t.sol`). What is missing is the client: a channel-lifecycle driver and a watch loop that calls these primitives from `armada-interface` or a future mobile client.

## A.6.3 What a client would need to build

**The lifecycle driver (T6.2).** A caller that funds a channel from a note (`depositFromNote`), exchanges off-chain `ForceMove` states, and drives cooperative close or force-close. None of `armada-interface`'s shield/transact handlers touch this; it is new client code, not a config change.

**The watchtower loop (T6.3).** Subscribe to the adjudicator's challenge events, compare the locally held state's turn number against the challenger's, and submit `checkpoint` with a higher-turn state before `finalizesAt`. The contract-side primitive is reused unchanged (`go-nitro`'s `ForceMove.checkpoint`); the loop and its always-on liveness are the net-new client work. A client that cannot stay connected for the full challenge window would need to delegate to an always-on operator running the same loop; no such delegation mechanism exists today.

**Reading chain state.** Neither piece has a verified read path yet. T1's nimbus verified proxy is the intended trust boundary (ADR-0009); `armada-interface` does not use it today.

## A.6.4 Risks / open

- **No client-side evidence that the dispute path works end-to-end from a browser or phone.** The contract-side tests prove the adjudicator behaves correctly when poked the right way; no client pokes it that way yet.
- **Liveness is unaddressed.** A design for always-on delegation, if the client cannot stay connected through a challenge window, has not been specified for this codebase.

See also — siblings: [A.1](./A.1-reuse-inventory.md) · [A.4](./A.4-adjudicator-integration.md) (adjudicator and dispute primitives) · [A.5](./A.5-deposit-payout-contract.md) (the boundary a client would drive) · [T6-client-apps.md](../T6-client-apps.md) (T6.2/T6.3 status in the full client tier).
