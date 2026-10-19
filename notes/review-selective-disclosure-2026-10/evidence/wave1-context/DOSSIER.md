# Selective disclosure — audit-context dossier

Scope: `armada-circuits` disclosure circuit (`NullifierDisclosure`, `NullifierOwnership`,
`EdDSAPoseidonVerifier`, `BabyCheck`, `Num2Bits`, `Poseidon`) and the `armada-sdk` disclosure
module (every exported function in `src/disclosure/` and `src/node/disclosure-artifacts.ts`).
Context-building only. No verdicts, no severity. Every per-template/per-function record is in
`functions/*.md` in this directory. This file records what spans several of them.

## The pieces

**Circuit side** (`armada-circuits`): one Groth16 circuit, `NullifierDisclosure`
(`src/operations/NullifierDisclosure.circom:59-112`), wired to five public signals
(`statementHash, masterPublicKey, nullifier, treeNumber, leafIndex`) by
`src/main/main_disclosure.circom:5`. It calls `NullifierOwnership` (same file, L39-57) and four
circomlib templates: `Num2Bits(16)` x2, `BabyCheck` x2, `Poseidon` (as 2/3/5-ary instances), and
`EdDSAPoseidonVerifier`.

**SDK side** (`armada-sdk/src/disclosure/`): `statement.ts` (EIP-712 domain/digest/signature
verification), `circuit.ts` (circuit witness adapter), `builder.ts` (`buildDisclosurePackage`,
the prover-side orchestrator), `checker.ts` (`checkDisclosurePackage`, the verifier-side
orchestrator — the largest and most security-critical file, 341 lines), `chain-source.ts`
(`InMemoryDisclosureChainSource`, the event-lookup interface the checker depends on),
`serialize.ts` (wire format), and `src/node/disclosure-artifacts.ts` (loads the wasm/zkey/vkey
from disk for the Node prover).

## Ways in / who's on the other side / what sticks around

- **Ways in:** a prover calls `buildDisclosurePackage` to produce a `DisclosurePackage`; a
  verifier calls `checkDisclosurePackage` on an untrusted package plus a `DisclosureChainSource`
  it supplies.
- **Who's on the other side:** per the threat model, a dishonest prover (wants a false chain of
  custody accepted), a viewing-key-only holder (wants to forge a disclosure without the spending
  key), a verifier/third party (wants to learn more than disclosed), and an attacker who controls
  or corrupts the `DisclosureChainSource` the checker reads from.
- **Data that sticks around:** none, server-side — both `buildDisclosurePackage` and
  `checkDisclosurePackage` are pure functions over their arguments
  (`checkDisclosurePackage.md`, "Outputs & Effects": "No side effects; pure function").
  `InMemoryDisclosureChainSource` builds an immutable in-memory index at construction
  (`chain-source.ts:57-70`) and never mutates it afterward.
- **Calls you can't see inside:** `params.prover.verify` / `params.prover.prove` (a `ProverAdapter`,
  not read in this pass — the snarkjs/Groth16 verifier is a black box from the checker's point of
  view, `checkDisclosurePackage.md` precondition list); ethers.js internals
  (`TypedDataEncoder`, `verifyTypedData`, `getAddress`, `keccak256`, `AbiCoder`); the vendored
  `railgun-engine` key-derivation and Poseidon/EdDSA wrappers (`poseidon.ts`, `keys-utils.ts`,
  `ShieldNote.getNotePublicKey`/`getShieldNoteHash`, `getTokenDataHash`) — all "external-source-
  available" per the subagent analyses but not independently re-verified line-by-line in this pass.

## What binds the statement, the proofs, the signatures and the chain events together

This is the chain of custody the circuit and the SDK jointly build. Each link is one file:line
citation; a break in any link is a break in the whole chain.

1. **The statement is canonicalized before anything signs it.** `hashHops` keccak256-hashes a
   canonical JSON serialization of the hop list (`statement.ts:51-63`, `hashHops.md`), and that
   hash (`hopsHash`) is one field of the `Statement` EIP-712 struct
   (`statement.ts:24-32`, `DISCLOSURE_EIP712_TYPES`). `checkDisclosurePackage` recomputes
   `hashHops(pkg.hops)` and compares it to `pkg.statement.hopsHash` before processing any hop
   (`checker.ts:91-93`, `checkDisclosurePackage.md` block "Hop hash binding"). This is what stops
   a hop substitution after signing.
2. **One statement hash feeds two different signature schemes.** `statementDigest` /
   `statementHashField` (`statement.ts:73-115`) compute, from the same `(domain, statement)` pair,
   both the full 256-bit EIP-712 digest the EOAs sign, and (by reducing it mod
   `SNARK_SCALAR_FIELD`) the field element `S` the circuit's spending-key signature commits to.
   `buildDisclosurePackage` computes `statementHashField` exactly once per package
   (`builder.ts:103`) and threads the identical value into every `formatDisclosureInput` call for
   every H2/H4/H6 spend (`builder.ts:111`). `checkDisclosurePackage` independently recomputes the
   same field value (`checker.ts:95`) and checks every disclosure proof's `statementHash` public
   signal against it (`checker.ts:126-128`, inside `verifyProof`). So one statement commits both
   endpoint EOAs and every spending key to the same claim.
3. **The disclosure circuit binds all five public signals into one signature.** Inside the
   circuit, `msg = Poseidon(statementHash, masterPublicKey, nullifier, treeNumber, leafIndex)`
   (`NullifierDisclosure.circom:87-93`) and the EdDSA check is over `msg.out`
   (`NullifierDisclosure.circom:111`) — this is the mechanism that makes "a Groth16 public input
   that no constraint uses is not bound by the proof" (comment, `NullifierDisclosure.circom:20-21`)
   not apply to any of the five. The SDK's `formatDisclosureInput` constructs the identical
   five-input Poseidon message before signing (`circuit.ts`, `formatDisclosureInput.md`
   block "message = poseidon([...])"), and `disclosurePublicSignals` exposes the five values in
   the same order the circuit declares them public
   (`main_disclosure.circom:5` vs. `circuit.ts:60-62`, cross-checked in `disclosurePublicSignals.md`
   and `formatDisclosureInput.md`).
4. **Proof checking is per-hop, not per-package.** `checkDisclosurePackage`'s `verifyProof` closure
   is called once per hop that references a disclosure proof (H2, H4, H6) and caches by proof id
   (`checker.ts:120-141`, `checkDisclosurePackage.md`), re-checking `statementHash` and
   `masterPublicKey` against the statement on every call before the Groth16 verification. This
   means every H2/H4/H6 hop, not just the first, is bound to the same statement and the same
   wallet.
5. **On-chain events close the loop that the circuit leaves open by design.** The circuit does not
   check that a disclosed nullifier is actually published, or that a disclosed note opening
   matches a real commitment — the design doc states this is intentionally off-circuit
   (`NullifierDisclosure.circom:24-25`). `checkDisclosurePackage` is where this closes: H1/H5
   recompute the note's commitment hash and compare it against `DisclosureChainSource.shieldsIn`/
   a finalized-outcome allocation (`checker.ts:164-181`, `261-281` per `checkDisclosurePackage.md`);
   H2/H6 compare the proof's `(treeNumber, leafIndex)` against the opening
   (`checkSpendLeaf`, `checker.ts:317-325`) and the proof's `nullifier` against
   `DisclosureChainSource.nullifiersIn` (`checkNullifierEvent`, `checker.ts:327-335`); H4
   recomputes `depositBinding(channelId, expectedHeld)` locally with `keccak256`/`AbiCoder` and
   compares it to the on-chain `Deposited` event's `adaptParams` (`checker.ts:240-249`). Every one
   of these checks is only as strong as the `DisclosureChainSource` implementation the verifier
   supplies — see the open questions on `InMemoryDisclosureChainSource` below.
6. **The EOA layer and the wallet layer are bound in both directions, per spec.** The endpoint
   loop (`checker.ts:97-113`) requires every endpoint named in the statement to have a
   `verifyEndpointSignature`-valid EIP-712 signature recovering to the claimed account
   (`statement.ts`, `verifyEndpointSignature.md`), and the statement itself names
   `masterPublicKey`. This matches the spec's "Key binding between EOAs and the shielded wallet":
   the EOA signs S and S names M; the spending key of M signs S inside the circuit, and S names
   the EOA (via the `endpoints` field of the same signed statement).

## Where the chain can diverge from "every claim is checked" — recorded as open questions, not findings

Two gaps recur across the per-function records and are structural rather than local to one
function, so they are recorded here instead of in only one file:

- **H4 is weaker than H2/H6.** H4 calls `verifyProof` (so `statementHash`/`masterPublicKey` are
  checked, and the Groth16 proof is checked) but, unlike H2 and H6, does not call
  `checkSpendLeaf` or `checkNullifierEvent` (`checker.ts` H4 block, `checkDisclosurePackage.md`
  open question 2). A package can contain an H4 hop with no corresponding H2 hop for the same
  proof; in that case the proof's `(treeNumber, leafIndex, nullifier)` are never compared against
  any opening or any on-chain `Nullified` event. `nothing found` in `checker.ts` that forecloses
  this combination.
- **H2/H6 do not call `checkOpening` on the spent note.** H3 and H5 both call `checkOpening`
  (which checks `opening.masterPublicKey === statement.masterPublicKey` and
  `npk == Poseidon(M, random)`) on the notes they open; H2 and H6 open the *spent* note via
  `requireOpening` but only call `checkSpendLeaf`, never `checkOpening`
  (`checker.ts` H2/H6 blocks, `checkDisclosurePackage.md` open question 3). `nothing found` that
  explains this asymmetry from the code alone.

## Index of nothing-found assumptions and open questions, by file

See each `functions/*.md` file for the full citations; this is the roll-up the acceptance
criterion asks for, consolidated below in the final reply rather than duplicated here verbatim
to avoid drift between this dossier and the per-file records.
