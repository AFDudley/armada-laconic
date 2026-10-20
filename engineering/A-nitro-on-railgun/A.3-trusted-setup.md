# A.3 — Trusted setup (Phase-1 reuse + own Phase-2)

work package A · **2026-10-20**
**Parent:** [`A.0`](./A.0-overview.md) · **Owns:** T0.1 — the built circuit; the open ceremony

---

## A.3.1 What is built

`armada-circuits@ef2682f:src/operations/Transact.circom` is Armada's own circuit, built on `circomlib ^2.0.5`, instantiated as 19 `(nIn,nOut)` shapes (`main_NxM.circom`), from `(1,1)` through `(8,4)`. This is a single `Transact` template, not a separate circuit per combination, and it is not named or shaped as "JoinSplit" in the circuit code, though it fills the same role. Proving is Groth16, through `snarkjs` in `armada-sdk`'s worker-first prover.

## A.3.2 Phase-1 — reused, unchanged

The community Perpetual Powers of Tau is reused as the Phase-1 ceremony input. Armada runs no universal ceremony of its own.

## A.3.3 Phase-2 — open, not finished

The `v1.0.0-ceremony` tag's `ceremony/manifest.json` lists every shape `"status": "open"`, with no beacon. `main` carries only `scripts/setup-dev.sh`, which the repository marks unsafe for anything but local testing. Today's registered verifying keys, wherever the pool has any registered, trace to this dev script, not to a published, multi-party transcript. Security rests on a 1-of-N honest assumption once the real ceremony runs (ADR-0003).

## A.3.4 The 19 shapes

`armada-poc@102f17b8:lib/sdk/armada-artifacts.ts:28-47` lists the 19 `ARMADA_SHAPES` the circuits build produces, and `lib/artifacts.ts` lists the subset the deploy/registration tooling actually registers: `(1,1)` lend/redeem, `(1,2)` shield, `(2,2)` simple transfer, `(2,3)` transfer+change, `(8,4)` medium consolidation, plus smaller consolidation, transfer-with-change, and fan-out shapes. Keys are loaded from `armada-circuits/build/<N>x<M>/`, fetched by `scripts/fetch-circuits.sh`.

## A.3.5 Selective disclosure — a second circuit, in progress

`NullifierDisclosure.circom` (112 lines) exists on the `disclosure-circuit` branch (`7175b46`), with matching client code on `armada-sdk`'s `selective-disclosure` branch (`0aa2ba1`). It adds shapes to this same Phase-2 ceremony, not a new ceremony kind (ADR-0003 D3.4).

## A.3.6 Re-run trigger

Phase-2 is re-run only when circuits change. The selective-disclosure circuit is one trigger; a future T0.6 (native channelized commitment, [A.9](./A.9-native-commitment.md)) would be another, if it is ever built.

## A.3.7 Testing

`armada-circuits` and `armada-sdk` ship test vectors (`boundparams-hash-vectors.json`, `commitment-vectors.json`, `eddsa-spend-auth-vectors.json`, `merkle-vectors.json`, `nullifier-vectors.json`, `poseidon-vectors.json`, `transaction-struct-vectors.json`) and differential tests against them. No CI gate for a production `zkey verify` exists yet, because no production ceremony has run.

## A.3.8 Risks / open

- **No production ceremony.** The dominant open risk for T0.1 is the ceremony itself, not the circuit code. Mainnet cannot rely on a verifying key traced to `setup-dev.sh`.
- **Any claim of a finished ceremony in older docs or decks is wrong.** Check the `v1.0.0-ceremony` tag's manifest before repeating such a claim.

---

Cross-refs: circuit/pool citations → [A.1](./A.1-reuse-inventory.md); vkey registration → [A.2](./A.2-pool-deployment.md); scheme decision → [ADR-0003](../09-architecture-decisions.md#adr-0003); the ceremony status ruling → `labbook/2026-10-19-docs-versus-code.md`.
