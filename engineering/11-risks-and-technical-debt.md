# 11. Risks & Technical Debt

arc42 §11 · 2026-10-20

This chapter catalogs the material risks to the code as built, plus the shortcuts accepted on purpose. Every row owns a `T#.#` item (§5 registry) and, where a decision has already scoped it, an `ADR-####` (§9).

## Risk register

| # | Risk | Impact | Owner |
|---|---|---|---|
| R1 | **Anonymity-set cold-start** — the pool's k-anonymity starts near zero and compounds only with volume. No code fixes this; it needs a crowd, and no cross-pool proof exists or is planned to shortcut it. | High (the product's core promise degrades to k≈1) | T0.7 · ADR-0006 |
| R2 | **No production Phase-2 ceremony.** `v1.0.0-ceremony`'s manifest lists every shape `"status": "open"`, with no beacon. Only the explicitly unsafe `scripts/setup-dev.sh` exists. Any deployed verifying key today traces to it, not a published transcript. | Critical before mainnet (an unaudited, uncontributed ceremony undermines soundness of every proof) | T0.1 · ADR-0003 |
| R3 | **Proof path unintegrated.** The nimbus verified proxy exists upstream but runs nowhere in Armada's stack. A client today reads chain state over RPC or an unverified cache (`armada-relayer`'s Ponder indexer), not a verified proof. | High (today's reads are trust-the-RPC, not trust-math) | T1, T2 · ADR-0009 |
| R4 | **No mobile client.** `armada-client` has no application code; `libverifproxy.a` has no upstream iOS/Android target. | High (no mobile product exists) | T6.5 |
| R5 | **No settlement-client channel lifecycle or watchtower.** The adjudicator side is built and tested; no client drives `depositFromNote`, channel state, or dispute response. A user cannot today use a Nitro channel from `armada-interface`. | High (settlement over channels is unreachable from any shipped client) | T6.2, T6.3 |
| R6 | **The six `go-nitro` patches are the audit surface, not only `ArmadaNitroAdjudicator`'s own code.** EIP-1271 participants, a guarantee-check fix, measured-balance crediting, an asset allow-list, a reentrancy lock, and `SafeERC20` payout all sit on top of a vendored base. | High (settlement is spend-authorizing; a patch bug is a fund-safety bug) | T0.2 · ADR-0004 |
| R7 | **No venue exists.** T3 and T4 are entirely unbuilt. This is a real gap against the product's exchange ambitions, not a sequencing choice to defer. | Medium (no exchange/clearing capability today) | T3, T4 |
| R8 | **`HubCustodian`'s FROST scheme has no Armada-specific audit cited in these docs.** `frost-secp256k1-evm` is a third-party library, pinned by commit. | Medium–High before any hub goes live | ADR-0016 |
| R9 | **Own pool and circuits are real, audit-critical crypto-engineering, now built and tested, but not yet externally audited** in anything cited here. A subtle spec deviation (note/commitment format, nullifier semantics, verifier wiring) is a soundness or fund-safety bug. | High (a soundness/format bug = unprovable spend or value creation) | T0.0, T0.1 · ADR-0014 |

## Technical debt / known shortcuts

- **The deployed shield fee is not zero.** `scripts/deploy_privacy_pool.ts` sets 50 bps against a 0-bps contract default. Any downstream integration that assumed fee = 0 is wrong. → ADR-0002.
- **The live indexer is `armada-relayer`'s Ponder/Postgres watcher, not a verified feed.** It serves `/v2/quick-sync` with no proof. It stays live until the proof path (T1) is integrated and a client stops depending on it as a trust boundary.
- **The web client is the only client.** `armada-interface` covers shield, transfer, and unshield; it does not cover channel settlement, watchtower response, or mobile. This is a real capability gap, not a deferred nicety.
- **Selective disclosure is unmerged.** `NullifierDisclosure.circom` and its SDK counterpart exist only on `disclosure-circuit`/`selective-disclosure` branches. → ADR-0020.

Cross-refs: ownership and status per the [§5 registry](./05-building-block-view.md); decisions in [§9](./09-architecture-decisions.md); the audit that produced this register is `labbook/2026-10-19-docs-versus-code.md`.
