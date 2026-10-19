## `formatDisclosureInput` in /home/rix/code/git_puller/repos/armada-sdk/src/disclosure/circuit.ts (L38-L58)

**Purpose:** Constructs a complete witness for the NullifierDisclosure.circom circuit. Builds the two private scalars (nullifyingKey), the nullifier from the leaf index, the message that the spending key must sign, and the signature itself. This function prepares all inputs the circuit will receive.

**Inputs & Assumptions:**
- `keys` (DisclosureSigningKeys): An object with `spendingPublicKey` (tuple of bigints), `spendingPrivateKey` (Uint8Array), `nullifyingKey` (bigint), and `masterPublicKey` (bigint). Trust: semi-trusted (derived from the wallet's internal key material; the wallet must have produced these keys correctly).
  - Precondition: `spendingPrivateKey` must be 32 bytes; this is assumed from wallet-node derivation (vendor/railgun-engine/src/key-derivation/wallet-node.ts:90, `getSpendingKeyPair()` derives via `keys-utils.ts:10`, `getPublicSpendingKey(privateKey)` which checks `privateKey.length !== 32` at keys-utils.ts:10).
  - Precondition: `spendingPublicKey` must be on the BabyJubJub curve; this is assumed because it comes from `getPublicSpendingKey`, which delegates to `circomlibjs` eddsa.prv2pub (keys-utils.ts:10). The circuit itself enforces this again on-circuit (NullifierDisclosure.circom:122-129).
  - Precondition: `nullifyingKey` must be a valid scalar in the BabyJubJub field; it derives from `Poseidon(viewingPrivateKey)` in wallet-node.ts:103.
  - Precondition: `masterPublicKey` must equal `Poseidon(spendingPublicKey[0], spendingPublicKey[1], nullifyingKey)`; this is pre-computed and passed in, but the circuit re-derives and checks it (NullifierDisclosure.circom:113-120).
- `statementHash` (bigint): A hash of the full disclosure statement (verifier, context, hops, endpoints). Trust: untrusted (caller-supplied; the circuit binds it into the signed message).
- `treeNumber` (number): The tree identifier in the pool. Trust: semi-trusted (from the disclosed commitment location). Precondition: must fit in 16 bits (the circuit enforces this, NullifierDisclosure.circom:101-103).
- `leafIndex` (number): The position of the note in the tree. Trust: semi-trusted (from the disclosed commitment location). Precondition: must fit in 16 bits (the circuit enforces this, NullifierDisclosure.circom:101-103).

**Outputs & Effects:**
- Returns a `DisclosureCircuitInput` object with fields:
  - `statementHash`, `masterPublicKey`, `nullifier`, `treeNumber`, `leafIndex`: public signal values in bigint form.
  - `spendingPublicKey`: the public point, as a tuple of two bigints.
  - `nullifyingKey`: the private scalar (revealed to the circuit, not the verifier).
  - `signature`: the EdDSA signature as `[R8x, R8y, S]` — the result of `signEDDSA`.
- No state writes; no external effects.

**Block-by-Block:**

```typescript
const nullifier = poseidon([keys.nullifyingKey, BigInt(leafIndex)]);
```
- **What:** Derives the nullifier from the nullifying key and the leaf index, matching the pool's on-chain formula (Transact.circom step 2d, engineering/16-selective-disclosure.md key facts).
- **Why here:** Must be computed before the message hash because the message includes the nullifier.
- **Assumes:** `poseidon` from vendor/railgun-engine/src/utils/poseidon.ts produces a deterministic hash of exactly two bigints. The implementation tries a WASM version first (poseidon.ts:34-46), falls back to circomlib's Javascript poseidon (circom.poseidon, from @railgun-community/circomlibjs). Both produce `bigint` (poseidon.ts:34, 39).
- **Establishes:** The nullifier value that the circuit will check against the disclosed nullifier in the public input.
- **Depended on by:** The message hash (next block) and the return value's `nullifier` field.

```typescript
const message = poseidon([
  statementHash,
  keys.masterPublicKey,
  nullifier,
  BigInt(treeNumber),
  BigInt(leafIndex),
]);
```
- **What:** Constructs the message that the spending key will sign, including all five public signals so the proof binds each of them (engineering/16-selective-disclosure.md, "Why the signed message includes every public input").
- **Why here:** Must be computed after nullifier, and before the signature, so the signature is over the complete message.
- **Assumes:** `poseidon` accepts exactly five bigints and produces a bigint. The order matches NullifierDisclosure.circom:107-111 (the `msg` component's input order).
- **Establishes:** The message that `signEDDSA` will receive.
- **Depended on by:** The `signEDDSA` call.

```typescript
const sig = signEDDSA(keys.spendingPrivateKey, message);
```
- **What:** Signs the message with the spending private key, using EdDSA over BabyJubJub with Poseidon hashing.
- **Why here:** Must come after the message is computed; the signature is the last field of the witness.
- **Assumes:** `signEDDSA` from vendor/railgun-engine/src/utils/keys-utils.ts:22 calls `eddsa.signPoseidon(Buffer.from(privateKey), message)` from @railgun-community/circomlibjs. It returns a `Signature` object with `.R8` (a point, `[bigint, bigint]`) and `.S` (a scalar, `bigint`). The private key length check happens inside `keys-utils.ts:10` (`getPublicSpendingKey`), which the wallet has already called to derive the public key; here it is not re-checked, so if `keys.spendingPrivateKey` is not 32 bytes, `signEDDSA` will fail inside the eddsa library.
- **Establishes:** The EdDSA signature that the circuit will verify against the disclosed public key.
- **Depended on by:** The return value's `signature` field.

```typescript
return {
  statementHash,
  masterPublicKey: keys.masterPublicKey,
  nullifier,
  treeNumber: BigInt(treeNumber),
  leafIndex: BigInt(leafIndex),
  spendingPublicKey: [keys.spendingPublicKey[0], keys.spendingPublicKey[1]],
  nullifyingKey: keys.nullifyingKey,
  signature: [sig.R8[0], sig.R8[1], sig.S],
};
```
- **What:** Packages the public signals, private inputs, and signature into the typed `DisclosureCircuitInput` object.
- **Why here:** Final return; delivers the witness in the shape the circuit prover expects.
- **Assumes:** The field names and order match the circuit's `signal` declarations in NullifierDisclosure.circom:62-68. The `signature` array is extracted from `sig.R8` and `sig.S` and reordered as `[R8x, R8y, S]` per the circuit's `signal input signature[3]` comment (NullifierDisclosure.circom:68, "EdDSA [R8x, R8y, S]").
- **Establishes:** A complete circuit witness, ready to pass to the prover.
- **Depended on by:** Callers that invoke `disclosurePublicSignals` on this object or pass it to the circuit prover.

**Cross-Function Dependencies:**
- `poseidon` (vendor/railgun-engine/src/utils/poseidon.ts): Hash function. Depends on it to hash exactly N bigints deterministically and reduce the result modulo the BN254 scalar field (SNARK_SCALAR_FIELD, types.ts:5). Poseidon uses a WASM module if available (poseidon.ts:34-46), or falls back to circomlib's JavaScript poseidon (poseidon.ts:40-46). Both paths must produce the same output for disclosure proofs to match on-circuit re-derivations.
- `signEDDSA` (vendor/railgun-engine/src/utils/keys-utils.ts:22): EdDSA signature. Depends on it to sign the message with the private key, producing a signature whose public component matches the disclosed `spendingPublicKey` and whose hash preimage (the message) is exactly as constructed here. The signature is verified in the circuit (NullifierDisclosure.circom:134-139) and will fail if the private key does not match the public key or the message is corrupted.
- Callers: Any code that calls `formatDisclosureInput` must hold the wallet's `spendingPrivateKey` and `nullifyingKey`. The function reveals `nullifyingKey` and the spending key's signature to the circuit, but keeps the private key itself secret (it is used only to sign, never returned).

**Open Questions:**
- Does the wallet-derivation path guarantee that `keys.spendingPrivateKey` is always 32 bytes? The check happens in `keys-utils.ts:10` inside `getPublicSpendingKey`, which `wallet-node.ts:90` calls. If `formatDisclosureInput` is called with a key that was not derived this way, the `signEDDSA` call may fail in the eddsa library with a less clear error.
- The assumption that `masterPublicKey` equals `Poseidon(spendingPublicKey[0], spendingPublicKey[1], nullifyingKey)` is passed in pre-computed. What ensures the caller computed it correctly? The circuit checks it (NullifierDisclosure.circom:113-120), but if the circuit's check fails, the prover will reject the witness. Should the function re-derive and compare it here, or is the circuit's check sufficient?
- The `treeNumber` and `leafIndex` are converted to `BigInt` here, but the circuit enforces they fit in 16 bits (NullifierDisclosure.circom:101-103). If a caller passes a number outside that range, does TypeScript's `BigInt()` coercion catch it, or does the circuit silently wrap?

