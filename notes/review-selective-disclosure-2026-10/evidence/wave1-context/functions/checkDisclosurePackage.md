## `checkDisclosurePackage` in /home/rix/code/git_puller/repos/armada-sdk/src/disclosure/checker.ts (L76-341)

**Purpose:** The verifier's main entry point. Validates every aspect of a disclosure package: package format, EIP-712 signatures, hop references, note openings, and proof verification for each hop. Returns on the first failure, naming the check and the hop index. On success, returns the facts the verifier learned (masterPublicKey, endpoints, transaction IDs, channel IDs).

**Inputs & Assumptions:**

- `params.package` (DisclosurePackage): the complete package to check. Trust: untrusted; comes from a third party.
- `params.source` (DisclosureChainSource): a read interface to pool events (Shield, Transact, Nullified, Unshield) and hub records (Deposited, finalized outcomes). Trust: semi-trusted; depends on the source implementation and how it was populated.
- `params.prover` (ProverAdapter): Groth16 verifier, bound to the disclosure circuit's proving key. Trust: trusted; the verifier holds the correct vkey.
- `params.vkey` (object): the disclosure circuit's verifying key. Trust: trusted; must match the prover's key.
- Preconditions:
  - The vkey is the correct disclosure circuit verifying key (no validation of vkey integrity in this function; established by caller or artifact loader).
  - The ProverAdapter.verify method is correctly implemented (assumed; not checked here).

**Outputs & Effects:**

- Returns `DisclosureCheckResult`:
  - On failure: `{ ok: false, failure: { check: DisclosureCheckName, detail: string, hop?: number } }`, stopping at the first failing check.
  - On success: `{ ok: true, learned: { masterPublicKey, endpoints, transactionIds, channelIds } }`.
- No side effects; pure function (the verifyProofs caching with verifiedProofs set is local and function-scoped).
- External calls: prover.verify (for each disclosure proof), source methods (shieldsIn, transactsIn, nullifiersIn, unshieldsIn, depositsIn, finalizedOutcomesIn, senderOf).

**Block-by-Block:**

```typescript
// L87-89: Version check
if (pkg.version !== DISCLOSURE_PACKAGE_VERSION) {
  return failure('package-version', `unsupported package version ${String(pkg.version)}`);
}
```
- **What:** Reject packages with unsupported schema versions.
- **Why here:** The package format is versioned; any version mismatch is a hard error before processing any hops.
- **Assumes:** `DISCLOSURE_PACKAGE_VERSION` is the current valid version (established in types.ts L18).
- **Establishes:** If the function continues, the package version is correct.
- **Depended on by:** All downstream logic that relies on the DisclosurePackage schema.

```typescript
// L91-93: Hop hash binding
if (hashHops(pkg.hops) !== pkg.statement.hopsHash) {
  return failure('statement-hops-hash', 'hopsHash does not match the hop list');
}
```
- **What:** Verify that the signed statement's `hopsHash` matches a fresh hash of the hop list.
- **Why here:** The statement is signed and names a hash of the hops. If the hop list changed after signing, the hashes won't match. This check ensures the hops have not been tampered with.
- **Assumes:** `hashHops` correctly computes `keccak256` of canonical JSON of the hop list (statement.ts L51-63).
- **Establishes:** If the function continues, the hop list has not been substituted or modified since signing.
- **Depended on by:** All hop verification downstream; a mismatch means the hops are not what was signed.

```typescript
// L95: Statement hash derivation (not checked, used later)
const statementHash = statementHashField(pkg.domain, pkg.statement).toString();
```
- **What:** Convert the full EIP-712 digest to a BN254 field element (the value the disclosure circuit consumes).
- **Why here:** The disclosure proofs sign `Poseidon(S, M, N, T, L)` where S is this field-reduced value. The statementHash must be recomputed here to match against every proof's public signal.
- **Assumes:** `statementHashField` correctly reduces the EIP-712 digest modulo SNARK_SCALAR_FIELD (statement.ts L73-75).
- **Establishes:** `statementHash` is the canonical reduced value the proofs must commit to.
- **Depended on by:** verifyProof closure (L126-128) and every hop that verifies a proof.

```typescript
// L97-113: Endpoint signature verification (H1, H6 endpoints; also verifies signing keys)
for (const endpoint of pkg.statement.endpoints) {
  const sig = pkg.endpointSignatures.find((s) => s.role === endpoint.role && s.account === endpoint.account);
  if (sig === undefined) {
    return failure('endpoint-signature', `no signature for endpoint role "${endpoint.role}"`);
  }
  if (!verifyEndpointSignature(pkg.domain, pkg.statement, endpoint.account, sig.signature)) {
    return failure('endpoint-signature', `signature for endpoint role "${endpoint.role}" does not recover to ${endpoint.account}`);
  }
}
```
- **What:** Verify that every endpoint named in the statement has a valid EIP-712 signature over the statement.
- **Why here:** The endpoints (public addresses) are part of the signed statement. Both EIP-712 signatures (e.g., 0xA and 0xB in the Aave example) must be present and recover to their claimed accounts. This binds the EOAs to the disclosure.
- **Assumes:**
  - Every endpoint in `pkg.statement.endpoints` is meant to have a signature (none are optional).
  - `verifyEndpointSignature` correctly recovers the signer from an EIP-712 signature and compares it to the claimed account (statement.ts L98-110).
  - The domain (chainId, verifyingContract) pins the pool, preventing cross-pool signature reuse (spec §"Key binding").
- **Establishes:** Every endpoint EOA has authorized the statement, including the hops, the masterPublicKey, and the verifier/context.
- **Depended on by:** H1 and H6, which use the endpoints to bind shielding/unshielding addresses. An endpoint without a valid signature is a binding failure.

```typescript
// L115-118: Map construction for O(1) lookups
const openingByLabel = new Map(pkg.openings.map((o) => [o.label, o]));
const proofById = new Map(pkg.proofs.map((p) => [p.id, p]));
const endpointByRole = new Map(pkg.statement.endpoints.map((e) => [e.role, e]));
const verifiedProofs = new Set<string>();
```
- **What:** Pre-index openings, proofs, and endpoints for O(1) reference resolution in the hop loop.
- **Why here:** Each hop references openings and proofs by label/id and endpoints by role. Pre-indexing avoids repeated linear searches.
- **Assumes:** Labels, ids, and roles in the hop list are exact string matches (no case normalization, no alias resolution).
- **Establishes:** Maps ready for use; no validation occurs here (see requireOpening for reference checking).
- **Depended on by:** verifyProof closure, requireOpening, checkOpening, checkHop (every hop).

```typescript
// L120-141: verifyProof closure
const verifyProof = async (entry: DisclosureProofEntry, hopIndex: number): Promise<DisclosureCheckResult | null> => {
  if (verifiedProofs.has(entry.id)) {
    return null;
  }
  if (entry.statementHash !== statementHash) {
    return failure('proof-statement-hash', `proof "${entry.id}" statementHash is not the statement's`, hopIndex);
  }
  if (entry.masterPublicKey !== pkg.statement.masterPublicKey) {
    return failure('proof-master-public-key', `proof "${entry.id}" masterPublicKey is not the statement's`, hopIndex);
  }
  const signals = [
    BigInt(entry.statementHash),
    BigInt(entry.masterPublicKey),
    BigInt(entry.nullifier),
    BigInt(entry.treeNumber),
    BigInt(entry.leafIndex),
  ];
  const ok = await params.prover.verify(entry.proof, signals, params.vkey);
  if (!ok) {
    return failure('proof-verification', `proof "${entry.id}" does not verify`, hopIndex);
  }
  verifiedProofs.add(entry.id);
  return null;
};
```
- **What:** Verify a disclosure proof's validity (Groth16 proof check) and public signal consistency.
- **Why here:** (This is a closure defined once, called multiple times.) Proofs are expensive to verify; caching avoids re-verification if the same proof appears in multiple hops (e.g., H2 and H4 on the same nullifier).
- **Assumes:**
  - `entry.statementHash`, `entry.masterPublicKey`, `entry.nullifier`, `entry.treeNumber`, `entry.leafIndex` are the five public signals of the disclosure proof (spec §"The disclosure circuit").
  - `params.prover.verify` correctly verifies the Groth16 proof against the vkey and signals; returns true only if the proof is valid.
  - The vkey matches the disclosure circuit's outputs.
- **Checks:**
  - `entry.statementHash` must equal the statement's field-reduced hash (L126-128). Prevents proof reuse for a different statement (spec: "if a holder of the viewing key forges a proof for a different statement, the verifier rejects it").
  - `entry.masterPublicKey` must equal the statement's masterPublicKey (L129-131). Prevents a prover from disclosing a nullifier under a false masterPublicKey.
  - The Groth16 proof verifies (L132-139). Confirms the prover holds the spending key and nullifyingKey, and knows the exact (treeNumber, leafIndex) of the spent note. The circuit also checks that the signature inside the proof (in private inputs) is valid (spec: "the spending key of M authorizes the statement hash S for this nullifier").
- **Establishes:** The proof is valid; its public signals (nullifier, treeNumber, leafIndex) are authentic and were signed by the spending key.
- **Depended on by:** H2, H4, H6 (only these hops use disclosure proofs). The proof verification is cached; if the same proof id appears in multiple hops, it verifies only once.

```typescript
// L144-148: Reference resolution helper
const requireOpening = (label: string, hopIndex: number): NoteOpening | DisclosureCheckResult => {
  const opening = openingByLabel.get(label);
  return opening ?? failure('hop-reference', `hop references unknown opening "${label}"`, hopIndex);
};
```
- **What:** Resolve an opening by label; fail if not found.
- **Why here:** Hops reference openings by label. This helper centralizes the error reporting for missing references.
- **Assumes:** Every referenced opening label is in the openings array (no validation of this; it's a package format invariant).
- **Establishes:** If the return is not a DisclosureCheckResult (i.e., the opening exists), it's guaranteed to be a NoteOpening.
- **Depended on by:** checkHop for H1, H2, H3, H5, H6.

```typescript
// L151-161: Opening validation helper
const checkOpening = (opening: NoteOpening, hopIndex: number): DisclosureCheckResult | null => {
  if (opening.masterPublicKey !== pkg.statement.masterPublicKey) {
    return failure('opening-master-public-key', `opening "${opening.label}" is not for the statement's masterPublicKey`, hopIndex);
  }
  const npk = ShieldNote.getNotePublicKey(BigInt(opening.masterPublicKey), opening.random);
  if (npk !== BigInt(opening.npk)) {
    return failure('opening-npk', `opening "${opening.label}" npk != Poseidon(M, random)`, hopIndex);
  }
  return null;
};
```
- **What:** Verify that an opened note belongs to the wallet's masterPublicKey and that npk = Poseidon(M, random).
- **Why here:** Every opened note must claim to belong to the disclosed masterPublicKey M. If the opening's npk does not match Poseidon(M, random), the opening is false or the random is wrong.
- **Assumes:** `ShieldNote.getNotePublicKey` correctly computes Poseidon(M, random) (re-exported from railgun-engine, external black-box; see Cross-Function Dependencies).
- **Establishes:** The opened note is consistent with the disclosed masterPublicKey and random.
- **Depended on by:** H1, H3, H5 (all hops that open notes). H2 and H6 open the spent note but do not call checkOpening (only checkSpendLeaf).

```typescript
// L164-181: Shield commitment location and hash check
const checkShieldLocation = (opening: NoteOpening, loc: CommitmentLocation, hopIndex: number): DisclosureCheckResult | null => {
  if (opening.tree !== loc.tree || opening.position !== loc.position) {
    return failure('shield-commitment', `opening "${opening.label}" position does not match hop shield location`, hopIndex);
  }
  const record = params.source
    .shieldsIn(loc.txid)
    .find((s) => s.tree === loc.tree && s.position === loc.position);
  if (record === undefined) {
    return failure('shield-commitment', `no Shield commitment at (${loc.txid}, tree ${loc.tree}, position ${loc.position})`, hopIndex);
  }
  if (BigInt(`0x${record.hash}`) !== commitmentHash(opening)) {
    return failure('shield-commitment', `Shield commitment at (${loc.txid}) does not match the opened note`, hopIndex);
  }
  return null;
};
```
- **What:** Verify that a Shield commitment exists at the claimed (tree, position) in the claimed transaction and matches the opened note's hash.
- **Why here:** H1 (shielding) and H5 (receiving from a channel) both land on Shield commitments. The opening's location must match the hop's shield location, and the on-chain commitment hash must match the computed hash of the opened note.
- **Assumes:**
  - `opening.tree` and `opening.position` are the leaf coordinates in the pool's Merkle tree.
  - `params.source.shieldsIn(txid)` returns all Shield commitments in that transaction (or an empty array if none). See Cross-Function Dependencies.
  - `commitmentHash(opening)` = Poseidon(npk, tokenHash, value) matches the on-chain commitment hash (L62-64).
- **Checks:**
  - (L166-168) The opening's location matches the hop's stated location. If they differ, the opening describes a different leaf than the hop claims.
  - (L169-173) A Shield commitment exists at that (txid, tree, position). If not found, the hop references a non-existent event.
  - (L174-176) The on-chain hash matches the computed hash of the opened note. If it doesn't, the opening is false (wrong npk, random, token, or value).
- **Establishes:** The note was shielded at the claimed location with the claimed value, owner, and token.
- **Depended on by:** H1 and H5.

**H1 (shielding): verifies that an EOA shielded a note.**
```typescript
    case 'H1': {
      const opening = requireOpening(hop.note, i);
      if ('ok' in opening) return opening;
      const endpoint = endpointByRole.get(hop.endpoint);
      if (endpoint === undefined) return failure('hop-reference', `H1 references unknown endpoint "${hop.endpoint}"`, i);
      const openErr = checkOpening(opening, i);
      if (openErr !== null) return openErr;
      const shieldErr = checkShieldLocation(opening, hop.shield, i);
      if (shieldErr !== null) return shieldErr;
      const sender = params.source.senderOf(hop.shield.txid);
      if (sender !== undefined && !addressesEqual(sender, endpoint.account)) {
        return failure('shield-sender', `Shield tx sender ${sender} is not endpoint ${endpoint.account}`, i);
      }
      transactionIds.add(hop.shield.txid);
      return null;
    }
```
- Resolves the note opening and the shielding endpoint (both by reference).
- Checks that the opening is valid (belongs to M, npk = Poseidon(M, random)).
- Checks that the Shield commitment exists and matches the opened note.
- Optionally checks that the transaction sender is the shielding EOA (if sender data is available). The sender check is optional ("if sender !== undefined") — if the chain source does not provide senders, this check is skipped.
- Adds the shield transaction to the learned set.
- **Spec mapping:** H1 in the spec (engineering/16-selective-disclosure.md, "The hops" table): "EOA → note · The `Shield` event shows `npk`. The user discloses `masterPublicKey` and `random`, and the verifier checks `npk = Poseidon(masterPublicKey, random)`. The EOA signs the statement." This code checks the npk derivation and confirms an endpoint signed; it does not explicitly verify the EOA sent the transaction (only optionally checks the sender if available).

**H2 (note is spent): verifies that a note was spent via a nullifier disclosure proof.**
```typescript
    case 'H2': {
      const opening = requireOpening(hop.spends, i);
      if ('ok' in opening) return opening;
      const proof = proofById.get(hop.proof);
      if (proof === undefined) return failure('hop-reference', `H2 references unknown proof "${hop.proof}"`, i);
      const proofErr = await verifyProof(proof, i);
      if (proofErr !== null) return proofErr;
      const posErr = checkSpendLeaf(proof, opening, i);
      if (posErr !== null) return posErr;
      const nullErr = checkNullifierEvent(proof, hop.nullifiedTxid, i);
      if (nullErr !== null) return nullErr;
      transactionIds.add(hop.nullifiedTxid);
      return null;
    }
```
- Resolves the spent note opening and the disclosure proof.
- Verifies the disclosure proof (Groth16 + public signal checks).
- Checks that the proof's leaf position (treeNumber, leafIndex) matches the opening's location (checkSpendLeaf).
- Checks that the proof's nullifier appears in a Nullified event for the claimed tree in the claimed transaction (checkNullifierEvent).
- Adds the nullifying transaction to the learned set.
- **Spec mapping:** H2: "note → spend transaction · A disclosure proof that a published nullifier belongs to the note at a given leaf, for the disclosed `masterPublicKey`." This code verifies the proof, the leaf position, and the nullifier event on-chain. The spending key's authorization is confirmed by the proof verification (inside verifyProof, the circuit checks the EdDSA signature).

**H3 (spend creates note): verifies that a Transact commitment was created by a spend.**
```typescript
    case 'H3': {
      const opening = requireOpening(hop.note, i);
      if ('ok' in opening) return opening;
      const openErr = checkOpening(opening, i);
      if (openErr !== null) return openErr;
      if (opening.tree !== hop.transact.tree || opening.position !== hop.transact.position) {
        return failure('transact-commitment', `opening "${opening.label}" position does not match hop transact location`, i);
      }
      const record = params.source
        .transactsIn(hop.transact.txid)
        .find((t) => t.tree === hop.transact.tree && t.position === hop.transact.position);
      if (record === undefined) {
        return failure('transact-commitment', `no Transact commitment at (${hop.transact.txid}, tree ${hop.transact.tree}, position ${hop.transact.position})`, i);
      }
      if (BigInt(`0x${record.hash}`) !== commitmentHash(opening)) {
        return failure('transact-commitment', `Transact commitment at (${hop.transact.txid}) does not match the opened note`, i);
      }
      transactionIds.add(hop.transact.txid);
      return null;
    }
```
- Resolves and validates the output note opening.
- Checks that the opening's location matches the hop's Transact location.
- Queries the chain source for a Transact commitment at that location.
- Verifies that the on-chain commitment hash matches the computed hash of the opened note.
- Adds the transact transaction to the learned set.
- **Spec mapping:** H3: "spend → output note · The user opens the output commitment: `npk`, token, value and `random`. The verifier recomputes the hash and compares it with the `Transact` event." This code does exactly that.

**H4 (spend funds a channel): verifies that a spend funded an armada-nitro hub channel.**
```typescript
    case 'H4': {
      const proof = proofById.get(hop.proof);
      if (proof === undefined) return failure('hop-reference', `H4 references unknown proof "${hop.proof}"`, i);
      const proofErr = await verifyProof(proof, i);
      if (proofErr !== null) return proofErr;
      const deposit = params.source
        .depositsIn(hop.depositTxid)
        .find((d) => addressesEqual(d.channelId, hop.channelId));
      if (deposit === undefined) {
        return failure('deposited-event', `no Deposited event for channel ${hop.channelId} in ${hop.depositTxid}`, i);
      }
      const binding = keccak256(AbiCoder.defaultAbiCoder().encode(['bytes32', 'uint256'], [hop.channelId, BigInt(hop.expectedHeld)]));
      if (BigInt(deposit.adaptParams) !== BigInt(binding)) {
        return failure('deposit-binding', `deposit adaptParams != depositBinding(channelId, expectedHeld) in ${hop.depositTxid}`, i);
      }
      if (!addressesEqual(deposit.asset, hop.asset)) {
        return failure('deposited-event', `Deposited asset ${deposit.asset} != ${hop.asset} in ${hop.depositTxid}`, i);
      }
      transactionIds.add(hop.depositTxid);
      channelIds.add(hop.channelId);
      return null;
    }
```
- Resolves the disclosure proof (the same proof as H2, linking the spent note to its nullifier).
- Verifies the proof (same as H2).
- Queries the chain source for a Deposited event in the claimed transaction for the claimed channel.
- Computes depositBinding(channelId, expectedHeld) locally using keccak256 and AbiCoder.
- Verifies that the on-chain adaptParams (from the Deposited event) matches the computed binding. **This is the critical check:** the binding ties the spending key's authorization (inside the proof) to the specific channel and amount via the spend's boundParamsHash. See spec §"The hops" table for H4.
- Verifies that the asset matches the hop's stated asset.
- Adds the deposit transaction and channel ID to the learned sets.
- **Spec mapping:** H4: "spend → armada-nitro channel · `depositFromNote` requires `adaptParams == depositBinding(channelId, expectedHeld)` (`armada-nitro/src/ArmadaNitroAdjudicator.sol`). The spend proof commits to `boundParamsHash`, so the spending key authorized this channel. The adjudicator emits `Deposited(channelId, asset, held)`." This code verifies the binding (line 242-244) and checks the Deposited event.

**H5 (channel pays a note): verifies that a channel payout note was shielded.**
```typescript
    case 'H5': {
      const opening = requireOpening(hop.note, i);
      if ('ok' in opening) return opening;
      const openErr = checkOpening(opening, i);
      if (openErr !== null) return openErr;
      const outcome = params.source
        .finalizedOutcomesIn(hop.concludeTxid)
        .find((o) => addressesEqual(o.channelId, hop.channelId));
      const paysNote = outcome?.allocations.some(
        (a) => a.allocationType === SHIELD_ALLOCATION_TYPE && BigInt(a.destination) === BigInt(opening.npk),
      );
      if (outcome === undefined || paysNote !== true) {
        return failure('finalized-shield-allocation', `channel ${hop.channelId} has no SHIELD allocation to note "${opening.label}" in ${hop.concludeTxid}`, i);
      }
      const shieldErr = checkShieldLocation(opening, hop.shield, i);
      if (shieldErr !== null) return shieldErr;
      transactionIds.add(hop.concludeTxid);
      transactionIds.add(hop.shield.txid);
      channelIds.add(hop.channelId);
      return null;
    }
```
- Resolves and validates the payout note opening.
- Queries the chain source for a finalized outcome for the claimed channel in the claimed conclude transaction.
- Checks that the outcome has a SHIELD allocation whose destination is the payout note's npk.
- Verifies that the Shield commitment exists and matches the opened note (same as H1).
- Adds the conclude and shield transactions and the channel ID to the learned sets.
- **Spec mapping:** H5: "channel → payout note · The finalized outcome has a `SHIELD` allocation whose `destination` is the payout `npk`. The pool's `Shield` event in the same transaction shows the note. The user opens the `npk` as in H1." This code checks the finalized outcome (with SHIELD allocation) and verifies the Shield event.

**H6 (note leaves the pool): verifies that a note was unshielded to an EOA.**
```typescript
    case 'H6': {
      const opening = requireOpening(hop.spends, i);
      if ('ok' in opening) return opening;
      const proof = proofById.get(hop.proof);
      if (proof === undefined) return failure('hop-reference', `H6 references unknown proof "${hop.proof}"`, i);
      const endpoint = endpointByRole.get(hop.endpoint);
      if (endpoint === undefined) return failure('hop-reference', `H6 references unknown endpoint "${hop.endpoint}"`, i);
      const proofErr = await verifyProof(proof, i);
      if (proofErr !== null) return proofErr;
      const posErr = checkSpendLeaf(proof, opening, i);
      if (posErr !== null) return posErr;
      const nullErr = checkNullifierEvent(proof, hop.unshieldTxid, i);
      if (nullErr !== null) return nullErr;
      const paidOut = params.source.unshieldsIn(hop.unshieldTxid).some((u) => addressesEqual(u.to, endpoint.account));
      if (!paidOut) {
        return failure('unshield-event', `no Unshield to ${endpoint.account} in ${hop.unshieldTxid}`, i);
      }
      transactionIds.add(hop.unshieldTxid);
      return null;
    }
```
- Resolves the spent note opening, the disclosure proof, and the withdrawal endpoint.
- Verifies the disclosure proof (same as H2).
- Checks that the proof's leaf position matches the opening.
- Checks that the proof's nullifier appears in a Nullified event.
- Checks that an Unshield event to the endpoint's account exists in the claimed transaction.
- Adds the unshield transaction to the learned set.
- **Spec mapping:** H6: "note → EOA · H2 for the spent note, plus the `Unshield` event to the recipient EOA. The recipient EOA signs the statement." This code verifies the H2 checks (proof, nullifier, leaf position) and the Unshield event; the endpoint's signature was already verified at the top of checkDisclosurePackage.

```typescript
// L317-325: checkSpendLeaf helper
function checkSpendLeaf(proof: DisclosureProofEntry, opening: NoteOpening, hopIndex: number): DisclosureCheckResult | null {
  if (proof.treeNumber !== opening.tree || proof.leafIndex !== opening.position) {
    return failure('spend-leaf-position', `proof "${proof.id}" leaf (${proof.treeNumber}, ${proof.leafIndex}) != opening "${opening.label}" (${opening.tree}, ${opening.position})`, hopIndex);
  }
  return null;
}
```
- **What:** Verify that the leaf position in the disclosure proof matches the spent note's position.
- **Why here:** A disclosure proof commits to a specific (treeNumber, leafIndex). The opening must be at exactly that position; if not, the proof and opening describe different notes.
- **Assumes:** The proof's treeNumber and leafIndex are the public signals from the proof (correct by verifyProof).
- **Establishes:** The proof's nullifier is for the note at the claimed position.
- **Depended on by:** H2 and H6.

```typescript
// L327-335: checkNullifierEvent helper
function checkNullifierEvent(proof: DisclosureProofEntry, txid: string, hopIndex: number): DisclosureCheckResult | null {
  const found = params.source
    .nullifiersIn(txid)
    .some((n) => n.tree === proof.treeNumber && n.nullifier === BigInt(proof.nullifier));
  if (!found) {
    return failure('nullifier-event', `nullifier of proof "${proof.id}" not in Nullified event of ${txid}`, hopIndex);
  }
  return null;
}
```
- **What:** Verify that the proof's nullifier appears in a Nullified event for the correct tree in the claimed transaction.
- **Why here:** The proof's nullifier is a public signal; the verifier must see it in the on-chain Nullified event to confirm the note was spent (not just that the prover claims it was).
- **Assumes:** `params.source.nullifiersIn(txid)` returns all Nullified events in that transaction (or an empty array).
- **Establishes:** The nullifier was published in an on-chain event.
- **Depended on by:** H2 and H6.

```typescript
// L337-343: Success return
return {
  ok: true,
  learned: {
    masterPublicKey: pkg.statement.masterPublicKey,
    endpoints: pkg.statement.endpoints.map((e) => ({ role: e.role, account: e.account })),
    transactionIds: [...transactionIds],
    channelIds: [...channelIds],
  },
};
```
- **What:** Return the verified facts.
- **Why here:** All checks have passed; the verifier can trust the package's claims.
- **Establishes:** The masterPublicKey, endpoints (EOAs), transaction hashes, and channel IDs are authentic and have been verified against on-chain events and cryptographic proofs.
- **Postcondition:** The caller learns what the package proved and can trust those facts.

```typescript
// L346-353: addressesEqual helper
function addressesEqual(a: string, b: string): boolean {
  try {
    return getAddress(a) === getAddress(b);
  } catch {
    return BigInt(a) === BigInt(b);
  }
}
```
- **What:** Compare two hex strings (addresses or bytes32) by value, tolerant of case differences and padding.
- **Why here:** Addresses and channel IDs may be stored in different cases (checksummed or lower-case). This normalizes before comparison.
- **Assumes:** `getAddress` (from ethers) checksums and normalizes EVM addresses; if it throws, fall back to bigint comparison.
- **Depended on by:** H1, H4, H5, H6 (any hop comparing addresses).

**Cross-Function Dependencies:**

- **verifyEndpointSignature** (statement.ts, L98-110): Recovers the signer of an EIP-712 signature and compares to the claimed account. Called at L110. Returns true only if the signature is valid and recovers to the expected account.

- **hashHops** (statement.ts, L51-63): Computes keccak256 of the canonical JSON of the hop list. Called at L92. The statement commits to this hash; any modification of the hop list will break the match.

- **statementHashField** (statement.ts, L73-75): Reduces the EIP-712 digest modulo SNARK_SCALAR_FIELD. Called at L95. This is the value the disclosure circuit's public input expects.

- **ShieldNote.getShieldNoteHash and ShieldNote.getNotePublicKey** (external, re-exported from railgun-engine): Used by commitmentHash (L62-64) and checkOpening (L157). These compute the Poseidon hashes that bind notes to their openings.

- **params.source.shieldsIn, transactsIn, nullifiersIn, unshieldsIn** (DisclosureChainSource interface): Return pool events. Called by checkShieldLocation (L170), H3 (L269), checkNullifierEvent (L319), and H6 (L286). The source must return all events in a transaction that match the query; if none match, it returns an empty array.

- **params.source.depositsIn, finalizedOutcomesIn** (DisclosureChainSource interface): Return hub records. Called by H4 (L240) and H5 (L261). These are armada-nitro events, not pool events.

- **params.source.senderOf** (DisclosureChainSource interface): Returns the transaction sender (from address), if known. Called by H1 (L216). If not implemented, returns undefined, and the sender check is skipped (L217-219).

- **params.prover.verify** (ProverAdapter interface): Verifies a Groth16 proof. Called by verifyProof (L138). Must return true only if the proof verifies against the vkey and signals.

- **commitmentHash** (internal, L62-64): Computes the commitment leaf hash. Calls ShieldNote.getShieldNoteHash. Used by checkShieldLocation (L176) and H3 (L274).

- **getTokenDataHash** (external, re-exported from railgun-engine): Computes the token data hash from a TokenData struct. Used by commitmentHash.

- **keccak256, AbiCoder** (ethers): Used to compute and verify the depositBinding in H4. External dependencies from ethers.js.

**Open Questions:**

1. **Sender validation in H1 is optional.** Line 217 checks `if (sender !== undefined)`, meaning if the chain source does not provide sender information, the check is silently skipped. The spec says "The EOA signs the statement" but does not explicitly require verification that the EOA sent the transaction. Is this intentional? If a chain source provides malicious or missing sender data, a verifier might accept a disclosure where the shielding EOA did not actually send the transaction.

2. **H4 does not call checkSpendLeaf or checkNullifierEvent.** H4 verifies the disclosure proof but does not explicitly check that the proof's (treeNumber, leafIndex) matches an opening or that the proof's nullifier appears in a Nullified event. **Nothing found:** The code does not enforce that H4 uses the same proof as H2; a package could have an H4 without an H2. In that case, the proof's nullifier and leaf are never checked against on-chain events or the opening.

3. **Spent note openings in H2/H6 are not validated with checkOpening.** H2 and H6 open a spent note but do not call checkOpening (lines 200-202, 276-277), only checkSpendLeaf. If the opening's masterPublicKey does not match the statement's, or the npk is not Poseidon(M, random), the error is not caught. H3 and H5 do call checkOpening, but H2 and H6 do not. What is the logic?

4. **The hash field of shield/transact commitments is stored as a no-0x hex string.** The comparison at L176 and L274 is `BigInt(`0x${record.hash}`) !== commitmentHash(opening)`. The coupling to the underlying representation is fragile.

5. **addressesEqual falls back to bigint comparison.** If getAddress throws (e.g., if the input is not a valid EVM address), the function tries `BigInt(a) === BigInt(b)`. This error-silencing pattern may mask the error. For channel IDs, why not use BigInt comparison directly?

6. **No binding between transaction hash and block/tree.** The chain source is indexed by txid only. There is no check that the txid came from a specific block or that events in one transaction belong to the same tree. A chain source implementation that substitutes events from a different transaction with the same hash would not be detected.

7. **The Nullified event search in checkNullifierEvent does not restrict by transaction sender.** Line 320 searches for any nullifier in the transaction that matches the tree and value. It does not verify that the nullifier came from the same spend. A malicious chain source or transaction could include multiple Nullified events; the verifier does not bind each nullifier to its corresponding spend.

8. **DisclosureChainSource is an interface; the implementation details are hidden.** The checker calls methods on params.source, but does not validate that the source is correct or complete. A malicious or incomplete source could cause the verifier to accept a false disclosure.

9. **The circuit's EdDSA signature check is not visible here.** The verifyProof function calls params.prover.verify, which returns a boolean. This check is inside the circuit; if the prover is dishonest, it could return true without actually checking the signature.

10. **Groth16 public signal binding is complex.** The verifyProof function constructs the signals in a specific order (L132-137) and passes them to the prover. If the prover implementation uses a different order, or if the vkey was compiled for a different order, the verification will pass for invalid proofs.
