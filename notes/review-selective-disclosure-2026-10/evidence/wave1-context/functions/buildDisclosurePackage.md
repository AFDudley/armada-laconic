## `buildDisclosurePackage` in src/disclosure/builder.ts (L79-L147)

**Purpose:** Build a complete disclosure package from wallet keys, note openings, nullifier proofs, and statement data. The function orchestrates the entire disclosure-proof generation: constructing note openings (computing note public keys), deriving the EIP-712 statement, reducing it to the scalar field for the circuit, generating Groth16 proofs for each spend, and collecting EOA signatures. The result is a forwardable package that proves a chain of custody without revealing the viewing key or nullifying key.

**Inputs & Assumptions:**
- `params` (BuildDisclosureParams): contains keyset (signing keys), domain (chainId, verifyingContract), verifier (audit recipient address), context (purpose string), issuedAt (Unix seconds, defaults to now), endpoints (EOA roles and addresses), openings (note data), spends (proof IDs and opening labels), hops (the chain of custody), prover (Groth16 adapter), artifacts (compiled circuit files), and signEndpoint callback.
- Trust: Mixed. The keyset, domain, and openings are trusted (provided by the wallet). The verifier, context, and hops are semi-trusted (provided by the user; they determine what is claimed). The signature callback result (signature strings) is untrusted.
- Implicit: `ShieldNote.getNotePublicKey()` computes npk from masterPublicKey and random. `poseidon()` and `signEDDSA()` are from the core library. Ethers' EIP-712 functions are used internally. The prover's `prove()` method generates a Groth16 proof.
- Preconditions: No validation. Assumes each opening's random is a 16-byte 0x-prefixed hex string; the keyset contains valid BN254 scalars; the hops array matches the declared spends; the `signEndpoint` callback returns 0x-prefixed 65-byte signature strings.

**Outputs & Effects:**
- Returns a DisclosurePackage: version, domain, statement, endpointSignatures (signed by EOAs), openings (note data), proofs (Groth16 proofs with public signals), and hops.
- No persistent state writes. Internal: constructs maps and arrays.
- The package is ready to forward to a verifier; the verifier can check the EOA signatures, the Groth16 proofs, and the consistency of the hop chain against on-chain events.

**Block-by-Block:**

```typescript
// L81-L84
const masterPublicKeyDec = params.keyset.masterPublicKey.toString();

const openings: NoteOpening[] = params.openings.map((o) => {
  const npk = ShieldNote.getNotePublicKey(params.keyset.masterPublicKey, o.random);
```

- **What:** Convert the master public key to decimal string; compute the note public key for each opening.
- **Why here:** The npk is derived from the masterPublicKey and the note random; both must be in the opening so a verifier can recompute and check it. The masterPublicKey is stored as a bigint in the keyset; it is converted to a decimal string for JSON serialization.
- **Assumes:** `params.keyset.masterPublicKey` is a valid bigint (BN254 scalar). `params.openings` is an array of BuilderOpening objects. `ShieldNote.getNotePublicKey()` takes a bigint and a 0x hex string, returns a bigint (the npk).
- **Establishes:** The `openings` array contains one NoteOpening per BuilderOpening, with npk computed correctly. The npk can be checked: `npk == Poseidon(masterPublicKey, random)` per the spec.
- **Depended on by:** L92 (openingByLabel map), L99 (looking up opening by spend.opening label).

```typescript
// L85-L93
  return {
    label: o.label,
    masterPublicKey: masterPublicKeyDec,
    random: o.random,
    npk: `0x${npk.toString(16)}`,
    token: o.token,
    value: o.value.toString(),
    tree: o.tree,
    position: o.position,
  };
});
const openingByLabel = new Map(openings.map((o) => [o.label, o]));
```

- **What:** Structure each opening for the package and build a lookup map by label.
- **Why here:** The openings are included in the package for the verifier; they must be complete and indexed by label for the spends to reference them. The npk is formatted as 0x-prefixed hex (L88); the value is formatted as decimal string (L90).
- **Assumes:** Each opening has a unique label. The BuilderOpening fields are correctly typed (random is 0x hex, token is TokenData, value is bigint, tree and position are integers).
- **Establishes:** A map from label → NoteOpening for fast lookup (L92). Each opening's npk is now in the package, formatted as 0x hex. The token and value are in the package, in the formats required by the DisclosurePackage type.
- **Depended on by:** L99 (lookup by spend.opening label to get the tree and position).

```typescript
// L94-L103
const issuedAt = params.issuedAt ?? Math.floor(Date.now() / 1000);
const statement: DisclosureStatement = {
  masterPublicKey: masterPublicKeyDec,
  verifier: params.verifier,
  context: params.context,
  issuedAt,
  endpoints: params.endpoints.map((e) => ({ role: e.role, account: e.account })),
  hopsHash: hashHops(params.hops),
};
```

- **What:** Build the EIP-712 statement with the issuedAt timestamp (default to now), the verifier, the context, the endpoints, and the hop list hash.
- **Why here:** The statement must exist before any proof is generated, because every proof's signed message commits to the statement hash (spec: "the spending key of M signs S inside the disclosure circuit").
- **Assumes:** `params.issuedAt` is a Unix seconds timestamp or undefined (then defaults to now). `params.verifier` and `params.context` are strings. `params.endpoints` is an array of role/account pairs. `hashHops()` returns a 0x-prefixed 32-byte hash.
- **Establishes:** A statement with the exact hopsHash; the hash binds the statement to this hop list. All proofs will commit to this statement (via statementHash), so changing any hop would invalidate all proofs.
- **Depended on by:** L105 (statementHashField, which reduces the statement to the field).

```typescript
// L105
const statementHash = statementHashField(params.domain, statement);
```

- **What:** Compute the statement hash as a BN254 field element.
- **Why here:** The statementHash is the input to every disclosure proof (circuit.ts `formatDisclosureInput()` L43). It must be computed once and reused for all proofs to ensure consistency.
- **Assumes:** `statementHashField()` returns a bigint in [0, SNARK_SCALAR_FIELD). The reduction is deterministic.
- **Establishes:** The field element statementHash, which is the public signal 0 (per circuit.ts L51) and is committed to by the spending-key signature inside the circuit.
- **Depended on by:** L99 (passed to formatDisclosureInput for every proof).

```typescript
// L107-L115
const proofs: DisclosureProofEntry[] = [];
for (const spend of params.spends) {
  const opening = openingByLabel.get(spend.opening);
  if (opening === undefined) {
    throw new InvalidRequestError(`buildDisclosurePackage: spend "${spend.id}" references unknown opening "${spend.opening}"`);
  }
```

- **What:** Validate that each spend references a known opening.
- **Why here:** Spends reference openings by label; a missing opening is a caller error (malformed request). Fail fast with a clear error message.
- **Assumes:** `spend.opening` is a string (the label). `openingByLabel` contains one entry per opening.
- **Establishes:** If the function returns (no exception), every spend references a valid opening with known tree and position.
- **Depended on by:** L118 (the opening is used to populate circuit input).

```typescript
// L116-L121
  const input = formatDisclosureInput(params.keyset, statementHash, opening.tree, opening.position);
  const signals = disclosurePublicSignals(input);
  const proof = await params.prover.prove(input, params.artifacts);
  proofs.push({
    id: spend.id,
    statementHash: signals[0]!.toString(),
```

- **What:** Format the circuit witness, compute public signals, generate the Groth16 proof, and record the result.
- **Why here:** For each nullifier the user wants to prove, the witness is formatted (with the keyset, the statement hash, and the leaf location), the public signals are extracted (for the package), and the proof is generated (off-thread via the prover adapter).
- **Assumes:** `formatDisclosureInput()` takes the keyset and the statement hash and produces the full witness (public inputs + private data + signature). `disclosurePublicSignals()` extracts the five public signals (statementHash, masterPublicKey, nullifier, treeNumber, leafIndex). `params.prover.prove()` is async and returns a Groth16Proof (object with a, b, c arrays).
- **Establishes:** The proof is generated with the statement hash committed inside (via the signed message). Changing the statement hash would invalidate the proof.
- **Depended on by:** L121-L125 (the signals are recorded in the DisclosureProofEntry).

```typescript
// L121-L126
    statementHash: signals[0]!.toString(),
    masterPublicKey: signals[1]!.toString(),
    nullifier: signals[2]!.toString(),
    treeNumber: Number(signals[3]!),
    leafIndex: Number(signals[4]!),
    proof,
  });
}
```

- **What:** Extract the public signals from the proof and format them as strings/numbers for the package.
- **Why here:** The DisclosureProofEntry type requires the signals as decimal strings (or numbers for treeNumber and leafIndex). The signals come back as bigints from the circuit.
- **Assumes:** `signals` is an array of five bigints, in order (per circuit.ts `DISCLOSURE_PUBLIC_SIGNAL_NAMES`). The non-null assertion (`!`) is safe because `disclosurePublicSignals()` always returns exactly five signals.
- **Establishes:** A DisclosureProofEntry for this spend, with the public signals and the proof. The statementHash signal is the field-reduced digest; the nullifier is derived from the nullifying key and leaf index. Both are deterministic.
- **Depended on by:** L144 (proofs array is returned in the package).

```typescript
// L128-L135
const fullDomain = eip712Domain(params.domain);
const message = statementMessage(statement);
const endpointSignatures: EndpointSignature[] = [];
for (const endpoint of params.endpoints) {
  const signature = await params.signEndpoint({
    account: endpoint.account,
    domain: fullDomain,
    types: DISCLOSURE_EIP712_TYPES,
    message,
  });
```

- **What:** Build the EIP-712 domain and message objects, then request signatures from each endpoint EOA.
- **Why here:** The EOAs must sign the statement; the SDK does not hold their private keys. The `signEndpoint` callback receives the domain, types, and message, and returns a signature (the callback implements the EOA signing, e.g., via MetaMask).
- **Assumes:** `params.signEndpoint()` is an async function that returns a 0x-prefixed 65-byte signature string. The domain and message are constructed once and reused for all endpoints (consistent across all signatures).
- **Establishes:** The domain and message are identical for all endpoint signatures, so the verifier can check all signatures against the same statement.
- **Depended on by:** L137-L138 (the signature is collected in the endpointSignatures array).

```typescript
// L136-L139
  endpointSignatures.push({ role: endpoint.role, account: endpoint.account, signature });
}

return {
```

- **What:** Collect the endpoint signatures into an array and return the package.
- **Why here:** Each endpoint's signature is recorded with its role and account for the verifier.
- **Assumes:** Each signature is a valid 0x hex string. The roles and accounts match the statement's endpoints.
- **Establishes:** The endpointSignatures array, with one entry per endpoint.
- **Depended on by:** L142 (returned in the package).

```typescript
// L140-L147
  version: DISCLOSURE_PACKAGE_VERSION,
  domain: params.domain,
  statement,
  endpointSignatures,
  openings,
  proofs,
  hops: params.hops,
};
```

- **What:** Return the complete disclosure package.
- **Why here:** All components are now assembled: the version, domain, statement, endpoint signatures, note openings, Groth16 proofs with public signals, and the hop list.
- **Assumes:** All fields are populated and correctly typed per the DisclosurePackage interface.
- **Establishes:** The package is ready to forward to a verifier. The package is immutable (readonly fields everywhere).

**Cross-Function Dependencies:**
- Callee `ShieldNote.getNotePublicKey` (core external-source-available): computes the note public key from masterPublicKey and random. Deterministic, no side effects.
- Callee `hashHops` (statement internal): produces the hop list hash. Per hashHops analysis.
- Callee `statementHashField` (statement internal): reduces the statement digest to the scalar field. Per statementHashField analysis.
- Callee `formatDisclosureInput` (circuit internal): formats the circuit witness. Takes keyset, statement hash, tree, leaf index; returns the full input (public + private signals).
- Callee `disclosurePublicSignals` (circuit internal): extracts the five public signals from the input. Per circuit.ts L51-L52: returns [statementHash, masterPublicKey, nullifier, treeNumber, leafIndex].
- Callee `params.prover.prove` (ProverAdapter external interface): generates a Groth16 proof. Async, takes formatted input and artifacts; returns a Groth16Proof.
- Callee `eip712Domain` (statement internal): constructs the EIP-712 domain. Per eip712Domain analysis.
- Callee `statementMessage` (statement internal): constructs the EIP-712 message. Per statementMessage analysis.
- Callee `params.signEndpoint` (EndpointSigner callback, untrusted): requests a signature from an endpoint EOA. Async, takes account, domain, types, message; returns a signature string. The SDK does not verify the signature; it collects it for the package.
- Constants: `DISCLOSURE_EIP712_TYPES` (statement L17-L27), `DISCLOSURE_PACKAGE_VERSION` (types L12).
- Callers: Exported function, called by user code to build a disclosure package.

**Open Questions:**
- **Circuit input threading:** L116 calls `formatDisclosureInput()` with `statementHash` (the field-reduced digest from L105). Per circuit.ts L44, `formatDisclosureInput()` passes this hash to `signEDDSA()` as the message to sign. The circuit's EdDSA verification then checks that the signature over `Poseidon([statementHash, masterPublicKey, nullifier, treeNumber, leafIndex])` is valid (circuit.ts comment and spec). The statement hash in the proof's public signal (signals[0], returned by L121) is the same field-reduced value. Confirm that the SDK's statementHashField output matches exactly what the circuit's public signal 0 expects (both are the BN254 scalar field element that the EdDSA message is reduced to).
- **No self-verification of endpoint signatures:** The function collects endpoint signatures via the `signEndpoint` callback but does not verify them (no call to `verifyEndpointSignature()`). This is by design (the SDK is the prover, not the verifier), but the function assumes the callback returns valid signatures. If a callback returns an invalid signature, the package will include it, and the verifier will reject it. No error is raised by the SDK.
- **Timestamp timezone:** L94 uses `Math.floor(Date.now() / 1000)`, which is the Unix timestamp (UTC). The statement includes this timestamp; no conversion or validation of the caller-provided `issuedAt`. The verifier is responsible for checking the timestamp if a time window is required (e.g., preventing old disclosures from being reused).
