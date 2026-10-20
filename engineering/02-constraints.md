# 2. Constraints

arc42 §2 · 2026-10-20

These are the fixed rules the architecture must obey, not decisions to revisit here. Most are anchored by an ADR (→ §9) and realized by a `T#.#` item (→ §5). Each entry states a constraint, its consequence, and the governing ADR or item.

## 2.1 Technical constraints

**TC-1 — Ethereum L1-anchored.** All value custody, adjudication, and the settlement boundary live in on-chain contracts (T0); custody never depends on the trust of an L2 or sidechain.
- *Consequence:* Gas cost and L1 confirmation latency bound every deposit/payout and force-close. A client reads L1 state through the nimbus verified proxy, never a trusted RPC of record.
- *Governs:* §5 T0, T1.

**TC-2 — The shielded pool is Armada's own implementation of the Railgun design.** `armada-poc` built the pool and its four modules (T0.0); `armada-circuits` built the `Transact` circuit (T0.1). Note and nullifier format are spec-compatible with Railgun; Armada never edits or touches Railgun's live pool (ADR-0002, ADR-0014).
- *Consequence:* The deployed shield fee is 50 bps, not 0 (ADR-0002). Value from Railgun enters as an ordinary deposit, through a wallet move (T6.7, not yet built).
- *Governs:* ADR-0002, ADR-0014 · §5 T0.0, T0.1.

**TC-3 — Groth16 over BN254 proving.** The pool keeps Groth16: Phase-1 reuses the community Perpetual Powers of Tau, and Phase-2 is Armada's own MPC over its own circuits (T0.1). The production Phase-2 ceremony has not run.
- *Consequence:* Each circuit change, such as the selective-disclosure circuit (ADR-0020) or a future T0.6, re-runs Phase-2 and needs a fresh audit. On-chain verification and browser prover performance stay Groth16-cheap.
- *Governs:* ADR-0003 · §5 T0.1, T0.6.

**TC-4 — `go-nitro` is patched, not vanilla.** `armada-nitro` vendored `go-nitro` and then applied six commits: EIP-1271 participants, a guarantee-check fix, measured-balance deposit crediting, an asset allow-list, a per-channel reentrancy lock, and `SafeERC20` payout. The deposit/payout boundary lives inside `ArmadaNitroAdjudicator`.
- *Consequence:* Audit of the settlement rail covers the patches, not only the deposit/payout glue.
- *Governs:* ADR-0004 · §5 T0.2.

**TC-5 — The client stack targets servers, browsers, iOS and Android; no mobile code exists yet.** `armada-client` `docs/SPEC.md` names the four targets. `armada-client` has no application code; the live client, `armada-interface`, is a web app.
- *Consequence:* Key custody today is browser memory (`keyManager.ts`), not a device secure enclave. Mobile custody, a mobile prover, and `libverifproxy.a` for iOS/Android are unbuilt work (T6.5, T6.6).
- *Governs:* §5 T6.0, T6.1, T6.5.

**TC-6 — EVM gas / verifier limits.** Every on-chain verifier and settlement path must fit within L1 gas and pairing-check budgets.
- *Consequence:* This favors Groth16's constant-size proof and cheap verify, reinforcing TC-3.
- *Governs:* ADR-0003 · §5 T0.

**TC-7 — The proof path is the nimbus verified proxy.** A client verifies headers against a beacon light client, storage with EIP-1186 proofs, and log inclusion and completeness by rebuilding a block's receipt trie (`eth_getBlockReceipts`). `watcher-ts` is an optional, unverified server cache.
- *Consequence:* `eth_getLogs` ranges are not used for note sync, because an omitted log is undetectable in a range. Mobile needs a `libverifproxy.a` cross-compile that does not exist upstream yet.
- *Governs:* ADR-0009 · §5 T1, T2.

## 2.2 Organizational & scope constraints

**OC-1 — Non-custodial at every tier.** No tier ever holds user funds in a way that lets it move them unilaterally; custody is contract-enforced (T0).
- *Consequence:* Watchers and any future venue (T4) operate only on user-signed channel state or transparent-boundary hops, never on discretionary custody.
- *Governs:* §5 T0 · quality scenarios in →§10.

**OC-2 — Railgun's deployed protocol is untouched; Armada runs its own pool.** Armada's own pool and circuits (T0.0/T0.1) are built for its own fee policy and settlement boundary, and never fork, patch, or govern Railgun's deployed protocol or live pool (ADR-0002, ADR-0014).
- *Consequence:* Blast radius splits cleanly — pool and audit changes never route through Railgun's deployment. The crowd grows only from Armada's own deposits (T0.7).
- *Governs:* ADR-0002, ADR-0014 · §5 T0.0, T0.1.

**OC-3 — One team; a reuse boundary, not an org split.** A single engineering team builds the whole Armada product, working from the ship-armada repos (ADR-0013). Only `go-nitro`, vendored and then patched, remains Laconic prior art in the build.
- *Consequence:* The governance and revenue contracts (`ArmadaGovernor`, `ArmadaTreasuryGov`, `ArmadaToken`, `ArmadaCrowdfund`, and the rest) are in scope and built, alongside the pool and the adjudicator.
- *Governs:* ADR-0013 · §0 · §5.

**OC-4 — No licensing discussion in these docs.** Licensing is neither open nor a risk this product needs to focus on or discuss here (operator ruling, `labbook/2026-10-19-docs-versus-code.md`).
- *Consequence:* No licensing section or risk item appears anywhere in this doc set.
- *Governs:* — (operator ruling).

## 2.3 Convention constraints

**CC-1 — arc42 is the one documentation template.** There is one item model, the §5 registry; tiers are views and tags, not competing hierarchies.
- *Consequence:* All engineering docs conform to the 12 arc42 sections. Building-block detail stays in the T-docs, decisions are ADRs, and cross-cutting concerns live in §8.
- *Governs:* ADR-0001.

**CC-2 — Terminology is fixed.** Use **the deposit/payout boundary** (not "adapter") for the merged role inside `ArmadaNitroAdjudicator`; reserve **RelayAdapt / Adapt / Cookbook** for Railgun's own recipe, which this codebase does not use, and **Adapters** (capital-A) for Armada's T5; use **boundary** not "seam"; **tier** (T0–T6) not "layer"; and **venue / clearing** not "marketplace."
- *Consequence:* Docs stay grep-clean for the banned terms, and collisions with Railgun/Ethereum/Feathers vocabulary are avoided.
- *Governs:* ADR-0010.

**CC-3 — Reference by id.** Cite building blocks by `T#.#` and decisions by `ADR-####`, and cross-reference sibling sections by number. Never renumber registry ids. A decided ADR is edited in place only on an explicit operator ruling to fix text the code now contradicts (ruling, `labbook/2026-10-19-docs-versus-code.md`); otherwise supersede it.
- *Consequence:* The §5 registry stays the single source of truth for ids and status, and every in-place edit cites the ruling that permitted it.
- *Governs:* ADR-0001, ADR-0010, ADR-0018.

---

*Constraints trace into strategy in →§4 and quality scenarios in →§10, and the maturity risks are tracked in →§11. See §5 for every `T#.#` item and §9 for the governing ADRs.*
