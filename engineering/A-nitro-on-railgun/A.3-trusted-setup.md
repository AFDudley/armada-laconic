# A.3 — Trusted setup (Phase-1 reuse + own Phase-2)

work package A · reuse-oriented spec · **2026-09-05**
**Parent:** [`A.0`](./A.0-overview.md) · **Owns:** T0.1 (trusted setup / ceremony)

---

## A.3.1 Goal

Produce the Groth16 proving and verifying artifacts our own pool ([A.2](./A.2-pool-deployment.md), T0.0) needs, under a ceremony we control. Per [ADR-0003](../09-architecture-decisions.md#adr-0003) the crypto scheme is kept as-is (BN254/Groth16). Per [ADR-0014](../09-architecture-decisions.md#adr-0014) the JoinSplit circuits are our own implementation, spec-compatible with A.1.2's JoinSplit behavior and format, rather than reused verbatim. T0.1 therefore delivers both our own `.circom` circuits and a fresh **Phase-2** contribution set over them:

- Reuse the community **Phase-1** (Perpetual Powers of Tau, `powersOfTau28_hez_final_20.ptau`, 2²⁰) unchanged. The expensive, universal, already-vetted part is inherited.
- Run our own **Phase-2 MPC** per circuit, over the circuit set we choose to register (A.3.4). Security rests on a **1-of-N honest** assumption, backed by a published transcript and a final randomness beacon.
- Emit, per registered circuit combo, the triple `wasm` (prover witness), `zkey` (prover key), and `vkey` (verifying key). The `vkey` is registered on-chain via `Verifier.setVerificationKey` (see [A.2](./A.2-pool-deployment.md)); the `wasm` and `zkey` become the prover artifacts that D's T6.6 prover and the T6.2 client ship (see A.3.5).

We take no dependence on Railgun's specific ceremony instance. Reusing their published Phase-2 zkeys would trust their instance ([ADR-0003](../09-architecture-decisions.md#adr-0003) alternatives), so we re-contribute Phase-2 ourselves.

## A.3.2 Boundary

- **T0.1 owns:** our own `.circom` circuits (spec-compatible with A.1.2, [ADR-0014](../09-architecture-decisions.md#adr-0014)), the Phase-2 MPC run, the snarkjs pipeline that turns `.circom` into `.r1cs`/`.wasm`, then per-circuit `.zkey`, then `.vkey`, the transcript and beacon publication, the artifact manifest with hashes, and the CI check that verifies every `zkey`.
- **T0.1 does *not* own:** on-chain vkey registration (that is a T0.0 governance action, [A.2](./A.2-pool-deployment.md)); the prover runtime that consumes `wasm`/`zkey` at proving time (that is D's T6.6 prover and the T6.2 client, see A.3.5). Terminology per [ADR-0010](../09-architecture-decisions.md#adr-0010): the ceremony is Phase-1 reuse plus Phase-2 MPC, not an "adapter" and not RelayAdapt/Cookbook, which is Railgun's own recipe contract, unrelated here.
- **Provenance.** Phase-1 ptau is a fixed, externally hosted file (2²⁰), genuinely reused as-is. The reference circuit design is Railgun `circuits-v2`, currently unpinned (read at `master` HEAD 2026-09-05). Pin the reference commit before authoring, so our own circuits are spec-compatible against a fixed target (A.1 provenance; [A.0.6](./A.0-overview.md)). The ceremony output is reproducible against our own circuit sources and the pinned reference commit.

## A.3.3 Reuse inventory (cite A.1.2 + pinned commits)

Everything T0.1 references — the design its circuits reproduce, plus the genuinely reused Phase-1 ptau — is cited in **A.1.2**; this spec does not re-document it (per [ADR-0012](../09-architecture-decisions.md#adr-0012)):

| Spec-referenced / reused piece | A.1.2 citation (`circuits-v2` @HEAD — **pin reference before build**) | What it gives T0.1 |
|---|---|---|
| JoinSplit circuit (payments primitive) | `src/library/joinsplit.circom` L11–118 | The **behavior our own circuit reproduces**, compiled per `(nIn,nOut)`. Public signals `merkleRoot, boundParamsHash, nullifiers[nIn], commitmentsOut[nOut]` fix the vkey's public-input arity and the note/commitment/nullifier format T6.1/T0.3 must match. Also enforces EdDSA ownership, Merkle membership, range `<2¹²⁰`, `sumIn===sumOut`. |
| Nullifier derivation | `src/library/nullifier-check.circom` L11–13 | `nullifier = Poseidon(nullifyingKey, leafIndex)` — the derivation the pool's double-spend set keys on; unchanged by the ceremony. |
| Circuit-combo generator | `lib/circuitConfigs.js` | Enumerates the `(nIn,nOut)` combos → **91** combos (`nIn+nOut ≤ 14`). Our own circuit set mirrors this *superset*; A.3.4 decides the registered subset. |
| Ceremony driver | `scripts/prepare_ceremony` L23–79 | Confirms the ADR-0003 shape: Phase-1 downloads `powersOfTau28_hez_final_20.ptau` (2²⁰), Phase-2 seeds a per-circuit `zkey` from `r1cs + ptau` (via `child_process/zkey_create.js`). We reuse its Phase-1 step and *replace* its Phase-2 with our own multi-party contributions + beacon. |
| Trusted-setup docs | docs.railgun.org/…/trusted-setup-ceremony | Confirms Groth16 + Perpetual-PoT Phase-1 + per-circuit Phase-2 MPC, transcript on IPFS, **new ceremony required on any circuit change**. |

Pinned toolchain (from the `circuits-v2` reference README, A.1.2): **circom 2.0.6, circomlib 2.0.5** — our own circuits MUST compile with these to keep the r1cs (and thus the vkey public-input layout) bit-identical to what our `Verifier.verify` expects.

## A.3.4 Net-new delta (what T0.1 actually builds)

Per [ADR-0014](../09-architecture-decisions.md#adr-0014) and A.1.9, the net-new work is (0) our own `.circom` circuits themselves — spec-compatible with A.1.2's JoinSplit behavior and format. It also includes our own Phase-2 over those circuits and the pipeline around both:

1. **The Phase-2 MPC** over our own circuit set: N independent contributors, each destroying its toxic waste, over a public coordination channel; then a verifiable-delay or hash beacon finalizes each `zkey`. The 1-of-N honest assumption means soundness holds if any one contributor was honest. Deliverables: a published transcript (contributor list, per-contribution hashes, beacon value and source) and the final per-circuit `zkey`.
2. **The snarkjs pipeline** (net-new automation reusing snarkjs stages): `circom` compile produces `.r1cs` and `.wasm`; `snarkjs groth16 setup <r1cs> <ptau> <zkey0>`; iterate `snarkjs zkey contribute` per contributor; `snarkjs zkey beacon` to finalize; `snarkjs zkey export verificationkey <zkey> <vkey.json>`. Run once per registered combo.
3. **CI `zkey`-verify** — a pipeline stage that runs `snarkjs zkey verify <r1cs> <ptau> <zkey>` for every registered combo and fails the build on any mismatch, plus a manifest-hash check (A.3.5). This is the machine-checkable proof that the artifacts came from this ptau and these r1cs (A.3.6).

### A.3.4.1 Circuit-set reconciliation — the 91-vs-~54 decision

**Reconciled (A.1.2, A.1.8 correction #3):** `lib/circuitConfigs.js` generates 91 `(nIn,nOut)` combos: nIn runs from 1 to 13 and nOut from 1 to 14 minus nIn, so summing 1 through 13 gives 91. The widely cited "~54" is not what the source emits. It is Railgun's registered and published artifact subset, a proper subset of the 91. Our own circuits mirror this 91-combo superset (spec-compatible, [ADR-0014](../09-architecture-decisions.md#adr-0014)), so the reconciliation below picks the registered subset over our own circuit set, not Railgun's deployed one.

**Decision:** we ceremony and register that subset rather than the full 91. Concretely, we adopt Railgun's published-artifact manifest at the pinned circuit commit as the baseline subset. That subset is characterised by a bound of at most 5 outputs, at most 13 inputs, and at most 14 in total, which yields about 54 to 55 combos. The authoritative list is the published manifest at the pinned commit, enumerated verbatim into `manifest.json` (A.3.5), not re-derived by rule at build time.

**Rationale.**
- The extra 37 combos are all high-`nOut` shapes with six or more outputs that no shipping Railgun wallet or our T6.2 client emits; a `transact()` with six or more output notes is not produced by any client path. Ceremonying them would multiply the Phase-2 runs and the T6.6 prover-artifact footprint (wasm and zkey per combo) by about 1.7 times, all for shapes nothing can request.
- Our own settlement shapes are low-arity: deposit-in is an unshield `transact()` (few inputs, typically an `nOut = 1` change or unshield leg); shielded transfers (A.1.7) are small `(nIn,nOut)`; payout-out is `shield()`, which requires no proof at all (A.1.1, where `shield()` is not a JoinSplit). So the registered subset comfortably covers every A/B/C/D transaction shape.
- Matching Railgun's registered subset preserves wallet/prover interop (D's T6.6, the broadcaster shapes) and keeps the vkey registry (A.2) in lockstep with the artifact set clients ship.

**Consequence (drives T6.6 and A.2):** the registered subset fixes (a) the number of Phase-2 runs T0.1 executes, (b) the number of `setVerificationKey` calls T0.0 makes (see A.2), and (c) the exact `wasm`/`zkey` set the T6.6 prover bundles. A shape outside the subset is unprovable until the subset is extended and re-ceremonied (A.3.7 re-run trigger).

## A.3.5 Interfaces (ICD)

**Exposed by T0.1** (this is the "circuit `wasm`/`zkey` artifacts" edge A exposes in the [§0 ICD](../00-work-packages.md)):

- **`manifest.json`** — the registered-subset index. Per combo: `{ nIn, nOut, circuitId: "${nIn}x${nOut}", wasmHash, zkeyHash, vkeyHash }` (SHA-256), plus top-level `{ ptauName: "powersOfTau28_hez_final_20.ptau", ptauHash, circuitsCommit, circomVersion: "2.0.6", circomlibVersion: "2.0.5", beacon, transcriptURI }`. The manifest is the single source of truth for *which* combos exist and their content hashes.
- **Per combo, three artifacts:**
  - `${circuitId}.vkey.json` → **consumed by T0.0 / [A.2](./A.2-pool-deployment.md)**: registered by `Verifier.setVerificationKey(nIn, nOut, VerifyingKey)` (A.1.1 — `verificationKeys[nIn][nOut]`, `Verifier.sol` L27, L36–45). One call per combo. Registration MUST use the manifest's `vkeyHash` to confirm the on-chain key matches the ceremony output.
  - `${circuitId}.wasm` + `${circuitId}.zkey` → **consumed by D's T6.6 prover and the T6.2 settlement client**: witness generation + Groth16 proving. Distributed with `wasmHash`/`zkeyHash` from the manifest so clients can verify integrity on download.
- **Public-input contract (must not drift):** proofs the prover produces carry public signals in the order `Verifier.verify` assembles them — `[merkleRoot, hashBoundParams(boundParams), ...nullifiers, ...commitments]` (A.1.1 `Verifier.sol` L98–113; A.1.2 `joinsplit.circom` L11–15). The vkey's public-input count for combo `(nIn,nOut)` is therefore `2 + nIn + nOut`. This ordering/arity is the hard interface between the ceremony vkey and the pool.

**Consumed by T0.1:** the pinned `circuits-v2` sources (A.3.3) and the external `powersOfTau28_hez_final_20.ptau` (2²⁰). No other A item feeds T0.1.

## A.3.6 Acceptance / verification

1. **CI `zkey verify` (machine-checkable, mandatory).** `snarkjs zkey verify <r1cs> <ptau> <zkey>` passes for every combo in `manifest.json`, and the CI stage fails on any single failure. This proves that each `zkey` derives from this ptau and this r1cs (that is, from the reused Phase-1 and the pinned circuits) and that the beacon was applied.
2. **Manifest-hash integrity.** The recomputed SHA-256 of each `wasm`/`zkey`/`vkey` matches `manifest.json`, and `ptauHash`/`circuitsCommit`/`circomVersion` match the pinned inputs. This guards against silent artifact swaps between ceremony and distribution.
3. **End-to-end proof on fixturenet.** On a laconic fixturenet running our own pool ([A.2](./A.2-pool-deployment.md)), generate a witness with `${circuitId}.wasm`, prove with `${circuitId}.zkey`, and submit a `transact()`. The proof then verifies against the registered `vkey`, meaning `Verifier.verify` returns true and the transaction does not revert, for the combos exercised by the walking skeleton ([A.0.4](./A.0-overview.md)). A deliberately corrupted proof MUST revert. This closes the loop from ceremony to registration ([A.2](./A.2-pool-deployment.md)) to proving (T6.2/T6.6) to on-chain verification.
4. **Transcript published.** The contributor list, per-contribution hashes, and the beacon value and source are published (transcript URI in the manifest), so the 1-of-N assumption is externally auditable ([ADR-0003](../09-architecture-decisions.md#adr-0003)).

The walking skeleton's `transact()` (unshield-in / deposit leg) is the first consumer; its combo(s) MUST be in the registered subset (A.3.4) and pass check #3 before A deepens.

## A.3.7 Risks / open

- **Own circuits are audit-critical net-new.** T0.1 authors the circuits ([ADR-0014](../09-architecture-decisions.md#adr-0014)), spec-compatible with A.1.2, and runs Phase-2 over them; T0.0 authors the matching pool (A.2). The residual is that the circuits are audit-critical, not that any grant is outstanding.
- **Unpinned Railgun reference commit.** Our own circuits are only spec-reproducible against a pinned reference; the r1cs, and hence the vkey public-input layout, depends on the design we match. Pin the `circuits-v2` reference before authoring (A.1 provenance) and record `circuitsCommit` in the manifest.
- **Re-run trigger — any circuit change requires a Phase-2 re-run (T0.1).** Phase-1 (ptau) is never re-run, but Phase-2 is re-run whenever a circuit changes (docs, [ADR-0003](../09-architecture-decisions.md#adr-0003)). The concrete v1 trigger is T0.6 native or channelized commitment ("fork-lite", [ADR-0005](../09-architecture-decisions.md#adr-0005); see [A.9](./A.9-native-commitment.md)): it changes the circuit set, so it re-runs the Phase-2 MPC, re-registers vkeys (A.2), reships prover artifacts (T6.6), and requires a fresh audit. Extending the registered subset (A.3.4) beyond the baseline is the same event.
- **Subset-vs-superset drift.** If a future client path needs a combo outside the registered subset (six or more outputs, or more than 14 inputs and outputs combined), that shape is unprovable until the subset is extended and re-ceremonied. The manifest is the guardrail: clients MUST reject shapes absent from it rather than fail opaquely at `Verifier.verify`.
- **Ceremony liveness/coordination.** The 1-of-N honest guarantee needs at least one genuinely independent, non-colluding contributor and a credibly random beacon; a compromised run is fine for soundness as long as one contributor stays honest, but a fully colluding run is not. Contributor diversity is an operational requirement, not a code one.

---

Cross-refs: circuit/pool reuse citations → [A.1.2](./A.1-reuse-inventory.md); risks → [A.1.10](./A.1-reuse-inventory.md); vkey registration → [A.2](./A.2-pool-deployment.md); prover artifacts → D's **T6.6**; re-run on circuit change → [A.9](./A.9-native-commitment.md); scheme/ceremony decision → [ADR-0003](../09-architecture-decisions.md#adr-0003); fork-lite → [ADR-0005](../09-architecture-decisions.md#adr-0005).
