# Selective-disclosure audit — findings (candidates)

Scope: `armada-circuits` branch `disclosure-circuit` (`src/operations/NullifierDisclosure.circom`,
`src/main/main_disclosure.circom`, `tests/disclosure.test.js`, `tests/picus/`, `scripts/picus.sh`) and
`armada-sdk` branch `selective-disclosure` (`src/disclosure/*`, `src/node/disclosure-artifacts.ts`,
`docs/guide/selective-disclosure.md`), against `engineering/16-selective-disclosure.md` and ADR-0020.

Each finding is a **candidate**: it has not yet been refuted by an independent reviewer. Severity is
not assigned (context-building/hunting rule); impact is described under the stated threat model.

Tool note: this review had read/grep/glob/write-to-scratch tools only, no code execution. Every
"reproducer" below is source code to run with `node_modules/.bin/vitest`, not an executed result. The
static evidence (file/line citations, direct quotation of the methodology's own source) is the primary
proof for each finding; it does not depend on running anything.

---

## F1 — Checker's H4 hop never ties the disclosed nullifier to the deposit it claims to fund

**Claim.** `checkDisclosurePackage` accepts an `H4` ("spend funds a channel") hop whose disclosed
nullifier has no relationship at all to the `Deposited` event / spend it is paired with. A prover can
claim "my wallet funded channel C" using (a) a valid disclosure proof for *any* nullifier that really
belongs to their own wallet, however unrelated, plus (b) the public `channelId`/`expectedHeld`/`asset`
of a `Deposited` event that a **different wallet** actually produced.

**Root cause.** `armada-sdk/src/disclosure/checker.ts`, the `H4` case of `checkHop`, lines 237–258:

```
237  case 'H4': {
238    const proof = proofById.get(hop.proof);
239    if (proof === undefined) return failure('hop-reference', ...);
240    const proofErr = await verifyProof(proof, i);
241    if (proofErr !== null) return proofErr;
242    const deposit = params.source
243      .depositsIn(hop.depositTxid)
244      .find((d) => addressesEqual(d.channelId, hop.channelId));
245    if (deposit === undefined) return failure('deposited-event', ...);
248    const binding = keccak256(AbiCoder.defaultAbiCoder().encode(['bytes32','uint256'], [hop.channelId, BigInt(hop.expectedHeld)]));
249    if (BigInt(deposit.adaptParams) !== BigInt(binding)) return failure('deposit-binding', ...);
252    if (!addressesEqual(deposit.asset, hop.asset)) return failure('deposited-event', ...);
255    transactionIds.add(hop.depositTxid);
256    channelIds.add(hop.channelId);
257    return null;
258  }
```

`verifyProof` (lines ~113–131) only checks that `proof.statementHash`/`proof.masterPublicKey` match the
package's own statement, and that the Groth16 proof verifies for `proof`'s own five public signals. It
never looks at `hop.depositTxid`. Nothing in the H4 case calls `params.source.nullifiersIn(hop.depositTxid)`
or compares `proof.nullifier`/`proof.treeNumber` to anything the deposit transaction published.

Compare with the sibling hops that *do* carry a disclosure proof — `H2` (line 212) and `H6` (line 291) —
which both call:

```
312  // The proof's nullifier must appear in the named transaction's Nullified event, in the proof's tree.
313  function checkNullifierEvent(proof, txid, hopIndex) {
314    const found = params.source.nullifiersIn(txid)
315      .some((n) => n.tree === proof.treeNumber && n.nullifier === BigInt(proof.nullifier));
316    if (!found) return failure('nullifier-event', ...);
317    return null;
318  }
```

H4 has no analogous call. This is the only hop that consumes a disclosure proof and skips it.

**Spec says the check belongs here.** `engineering/16-selective-disclosure.md`, "What the verifier
checks" (the Aave-hub worked example), step 3: *"The disclosure proof for `n1`, and that `n1` is in the
`Nullified` event of the deposit transaction."* Step 4 is the `adaptParams`/`Deposited` check that H4
*does* implement. The spec requires both; the code implements only the second. (`armada-nitro`'s
`depositFromNote`, `src/ArmadaNitroAdjudicator.sol` lines 93–125, confirms step 3 is checkable: it calls
`pool.transact(transactions)` for the funding spend in the *same* transaction that emits `Deposited`, so
the spent note's nullifier is in that transaction's `Nullified` event precisely when a real funding spend
happened there.)

The SDK's own guide (`armada-sdk/docs/guide/selective-disclosure.md`, "D5") describes the *implemented*
(not the spec-required) behavior: *"`H2` and `H6` confirm that a proof's `(treeNumber, leafIndex)` is the
spent note's leaf, and that the nullifier is in the `Nullified` event."* H4 is not named. Code and this
one doc paragraph agree with each other and disagree with the authoritative spec — consistent with an
oversight made while implementing D5, rather than a deliberate relaxation.

**Trigger.** Build (or hand-craft, per `types.ts`'s plain-JSON `DisclosurePackage` shape) a package with:
- a valid disclosure proof `p` for the prover's own wallet `M`, for *any* nullifier they can produce
  (e.g. a trivial, already-disclosed, low-value spend of their own);
- an `H4` hop `{ proof: p.id, channelId: C, expectedHeld: E, asset: A, depositTxid: T }` where `T` is the
  txid of a **real** `Deposited(C, A, held)` event emitted by `depositFromNote` for a completely
  different wallet's spend. `channelId`/`expectedHeld`/`asset` are all public (readable from the chain,
  or from `adaptParams = depositBinding(C, E)` which is itself public), so the prover needs no
  privileged information to build a hop that satisfies lines 244–254 exactly.

`checkDisclosurePackage` returns `ok: true` for this hop and adds `C` to `learned.channelIds`, i.e. the
verifier is told "this wallet funded channel C," which is false.

**Why no existing test caught it.** `armada-sdk/src/disclosure/disclosure.test.ts`'s only package that
exercises `H4` is the shared "Aave hub" fixture (`beforeAll`, ~lines 103–110):

```
{ kind: 'H2', spends: 'N1', proof: 'n1', nullifiedTxid: TX.deposit },
{ kind: 'H4', proof: 'n1', channelId: CHANNEL, expectedHeld: V1.toString(), asset: USDC, depositTxid: TX.deposit },
```

The `H2` hop for the *same* proof `n1` and the *same* `TX.deposit` always precedes the `H4` hop in
`pkg.hops`, and hops are checked in array order (the `for` loop just above `checkHop`'s definition). So
the existing negative test "rejects a nullifier that is not in the Nullified event" (lines 261–264, which
filters `nullifiers` to drop the entry for `TX.deposit`) is caught by **H2's** `checkNullifierEvent` call,
before `checkHop` ever reaches the `H4` entry. No test builds a package where an `H4` hop's nullifier is
*only* referenced by that `H4` hop (no co-located `H2`), which is exactly the condition that exposes the
missing check. This is itself evidence for Finding F4 below (fixture/test dependence masking a real gap).

**Impact under the threat model.** Directly matches "A dishonest prover who wants a verifier to accept a
chain of custody that is false: claims another wallet's ... channel deposit ... Substitutes a hop." A
verifier relying on `checkDisclosurePackage`'s `learned.channelIds`/`ok` result for a compliance decision
(e.g. "this withdrawal's principal passed through channel C, which the hub attests paid X% yield") would
accept a forged funding claim.

**Reproducer.** Static evidence above (file/line citations) is self-contained and requires no execution.
A runnable test, written but **not executed** in this session (no execution tool; also, this checkout's
`armada-circuits/build/disclosure/` has no `main_disclosure.wasm`/`final.zkey` — confirmed via glob — so
`compile.sh`+`setup-dev.sh` would need to run first):

```ts
// Save as armada-sdk/src/disclosure/__repro_f1.test.ts and run with armada-sdk's vitest once the
// disclosure circuit artifacts are built (scripts/compile.sh && scripts/setup-dev.sh in armada-circuits).
// Reuses the exact fixture plumbing already in disclosure.test.ts.
import { describe, it, expect, beforeAll } from 'vitest';
import { keccak256, AbiCoder } from 'ethers';
import { buildDisclosurePackage } from './builder';
import { checkDisclosurePackage } from './checker';
import { InMemoryDisclosureChainSource } from './chain-source';
// ... reuse `prover`, `artifacts`, `keys`, `usdc`, `domain` from disclosure.test.ts's beforeAll, or
// re-derive them the same way.

it('F1: H4 accepts a deposit nobody disclosed the funding nullifier for', async () => {
  const STRANGER_CHANNEL = `0x${'cd'.repeat(32)}`;
  const STRANGER_TX = `0x${'ee'.repeat(32)}`;
  const EXPECTED_HELD = 999_000000n; // public, read off-chain
  const ASSET = '0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48';

  // Build a package with ONE valid disclosure proof for the prover's own, unrelated note (leaf 42),
  // and an H4 hop pointing at a stranger's deposit. No H2 hop for the same proof/tx.
  const pkg = await buildDisclosurePackage({
    keyset: keys, domain, verifier: 'v', context: 'f1-repro', issuedAt: 1_700_000_000,
    endpoints: [],
    openings: [{ label: 'N', random: `0x${'33'.repeat(16)}`, token: usdc, value: 1n, tree: 0, position: 42 }],
    spends: [{ id: 'n', opening: 'N' }],
    hops: [{ kind: 'H4', proof: 'n', channelId: STRANGER_CHANNEL, expectedHeld: EXPECTED_HELD.toString(), asset: ASSET, depositTxid: STRANGER_TX }],
    prover, artifacts, signEndpoint: async () => { throw new Error('no endpoints'); },
  });

  // The "stranger's" real on-chain Deposited event. Its adaptParams is PUBLIC data; the prover did not
  // need to produce it, only read it.
  const adaptParams = keccak256(AbiCoder.defaultAbiCoder().encode(['bytes32', 'uint256'], [STRANGER_CHANNEL, EXPECTED_HELD]));
  const source = new InMemoryDisclosureChainSource({
    deposits: [{ txid: STRANGER_TX, channelId: STRANGER_CHANNEL, asset: ASSET, held: EXPECTED_HELD.toString(), adaptParams }],
    nullifiers: [], // <-- deliberately NO Nullified event in STRANGER_TX for the disclosed nullifier,
                    //     or for anything at all: proves H4 never looks here.
  });

  const result = await checkDisclosurePackage({ package: pkg, source, prover, vkey: artifacts.vkey });
  expect(result.ok).toBe(true);                       // BUG: should be rejected.
  expect(result.learned?.channelIds).toContain(STRANGER_CHANNEL); // the verifier is told a false fact.
});
```

**Fix shape (not applied — read-only review):** add, in the `H4` case, the same call H2/H6 make:
`checkNullifierEvent(proof, hop.depositTxid, i)` before accepting the hop.

---

## F2 — The cited Picus "proof of uniqueness" for the wiring does not test the circuit's real public/private split

**Claim.** The design doc's verification-status table claims a machine-checked "Proof of uniqueness" for
"Our wiring: the range checks, the equality checks, the on-curve checks and the message routing," via
Picus strong safety on `tests/picus/wiring.circom`. The specific Picus invocation described does not
establish that property, because the abstraction it runs against does not preserve the real circuit's
public/private signal split.

**Root cause.** `armada-circuits/tests/picus/wiring.circom` ends with:

```
component main = Wiring();
```

— no `{public [...]}` list. Circom's own rule (confirmed from `docs.circom.io`, "Signals & Variables"):
*"the input signals of the main component are private if not stated otherwise using the keyword public."*
So every one of `Wiring()`'s 11 declared `signal input`s — `statementHash, masterPublicKey, nullifier,
treeNumber, leafIndex, spendingPublicKey[2], nullifyingKey, signature[3]`, plus the free hash stand-ins
`mpkHash, nfHash, msgHash` — compiles to a **private** R1CS input. Only the output `verifierM` is public.

This matters because Picus's own uniqueness check, read from the Picus source at the pinned commit
`138b151` (`scripts/picus.sh` pins this commit), does **not** distinguish circom's `public`/`private`
declaration at all for the purpose of deciding what is "known." `picus/r1cs/r1cs-grammar.rkt`, lines
406–411:

```rkt
; compute the list of input variables (aka `knowns` in Ecne) and output variables (aka `outputs` in Ecne)
(define istart (+ 2 (header-section-npubout hs0)))
(define iend (+ 1 (header-section-npubout hs0) (header-section-npubin hs0) (header-section-nprvin hs0)))
(define input-list-ecne (cons 1 (for/list ([i (range istart (+ 1 iend))]) i)))
```

`iend` spans `npubout + npubin + nprvin` — Picus's "knowns" set is **every declared input wire, public
and private together** (`picus.rkt` then uses this directly as the `input-set` held fixed across the two
compared models in its counter-example search). Whether a signal is circom-`public` or circom-`private`
only changes `npubin` vs. `nprvin`'s split inside that same span; it does not remove a private input from
the "known" set Picus holds fixed. (Confirmed this is not R1CS-source-specific: `picus/circom/circom-
parser.rkt`'s `parse-component` does read the circom-source `public [...]` list, but that list only feeds
back into the same R1CS header fields `circom` itself produces — it is not consulted independently by the
uniqueness algorithm.)

Consequence for `wiring.circom`: because `spendingPublicKey`, `nullifyingKey`, `signature`, `mpkHash`,
`nfHash` and `msgHash` are *themselves* part of what Picus holds fixed in both compared models, the
"strong safety: safe" result does not — and cannot — show that an adversary who controls only the
private witness (while the 5 *real* public inputs `statementHash, masterPublicKey, nullifier,
treeNumber, leafIndex` are fixed) is forced to a unique witness. That is exactly the property "proof of
uniqueness" would need to mean for this to back the stated security claim. What the check actually shows
is much weaker and nearly automatic: given a complete, fully-specified valuation of *all* 11 inputs, the
few remaining signals (the bit-wires inside the two `Num2Bits(16)` calls, and `x2`/`y2` inside the two
`BabyCheck()` calls) compute deterministically — which they trivially do, since every remaining
constraint is a single forward `<==`/`===` with no extra degree of freedom.

Independently of the Picus claim, I manually re-derived the real circuit's wiring
(`src/operations/NullifierDisclosure.circom`) and it does appear sound: all five public inputs are bound
into the EdDSA-signed message (`msg.inputs[0..4]`), and `masterPublicKey`/`nullifier` are additionally
bound through the `NullifierOwnership` equality checks. So this finding is about the **verification
claim**, not a confirmed break in the circuit itself — but the claim, as written, substitutes a
near-tautological result for the soundness-relevant one, leaving the reader with false confidence that a
machine proof backs the wiring when none that tests the public/private boundary actually exists. (Picus
cannot run the real check here at all: the design doc itself notes "Picus did not finish on
`EdDSAPoseidonVerifier`," which is why the piecewise abstraction exists in the first place — and the one
file that *does* carry the real `{public [...]}` declaration, `main_disclosure.circom`, is never fed to
Picus.)

**Reproducer.** None executed (no execution tool; Picus also needs Docker + cvc5, per `scripts/picus.sh`).
The evidence is the quoted Picus source (`picus/r1cs/r1cs-grammar.rkt` lines 406–411, commit `138b151`)
plus the absence of a `public` annotation in `wiring.circom`, both independently re-readable.
A confirming experiment for whoever has Docker: add `{public [statementHash, masterPublicKey, nullifier,
treeNumber, leafIndex]}` to `wiring.circom`'s `component main` line and re-run
`scripts/picus.sh`'s `wiring` check with `--strong`; if the result changes from `safe`, that is direct
confirmation the current (unannotated) run was not testing the public/private boundary.

---

## F3 — No cross-check that hop events belong to the statement's declared chain/pool

**Claim.** `checkDisclosurePackage` never verifies that the chain events satisfying `H1`–`H6` (via
`params.source`) were emitted by the pool/chain named in `pkg.domain`. The binding exists only for the two
EIP-712 EOA signatures (`H1`'s shielder, `H6`'s withdrawer); none of `H2`/`H3`/`H4`/`H5` (and, per F1,
even the nullifier-matching parts of `H2`/`H6`) are checked against `pkg.domain.chainId` /
`verifyingContract` at all.

**Root cause.** `DisclosureChainSource` (`armada-sdk/src/disclosure/chain-source.ts`) and its backing
types (`DecodedShieldCommitment`, `DecodedTransactCommitment`, `DecodedNullifier`, `DecodedUnshield` in
`src/sync/event-decoder.ts`; `HubDepositRecord`, `HubFinalizedOutcomeRecord` in `chain-source.ts`) carry
no chain id or contract-address field anywhere — only `tree`, `position`, `blockNumber`, `txid`. Every
lookup in `checker.ts` (`shieldsIn`, `transactsIn`, `nullifiersIn`, `unshieldsIn`, `depositsIn`,
`finalizedOutcomesIn`) is keyed purely by `txid`. `pkg.domain` is read in exactly two places in
`checker.ts`: inside `verifyEndpointSignature(pkg.domain, ...)` and `statementHashField(pkg.domain,
pkg.statement)` — never compared against anything `params.source` returns.

This matters specifically for this feature because `masterPublicKey` and the disclosure circuit's
`nullifier` carry **no** chain/pool domain separator by design (same Poseidon formulas as
`Transact.circom`, so the same 0zk address works unmodified across multiple Armada pool deployments —
this is an intentional, pre-existing Railgun-style property, not a new bug). A wallet active on more than
one deployment of the pool contract (different chain, or a second deployment on the same chain) would
produce an identical `masterPublicKey`, and could produce a colliding `nullifier` for a coincident leaf
index, on each deployment. Nothing in the checker's code path stops a `source` implementation that
aggregates data from more than one deployment (e.g. a shared multi-chain indexer behind one verifier
service) from satisfying `H2`/`H4`/`H6` with an event from the wrong deployment.

**Why this is a candidate and not asserted higher confidence.** The SDK's own docs put this burden
explicitly on the integrator: *"A verifier adapts its own chain access (RPC logs, an indexer) to these
per-transaction lookups"* (`docs/guide/selective-disclosure.md`). A verifier that constructs one
`DisclosureChainSource` scoped to exactly `pkg.domain.chainId`/`verifyingContract` per check is not
exposed. The gap is that the SDK provides **no enforcement, no parameter, and no test** that would catch
a verifier that gets this wrong — `pkg.domain.chainId`/`verifyingContract` is accepted into
`CheckDisclosureParams` but silently unused for anything except the two EOA signatures, with nothing
flagging that `source` must itself be pre-scoped. No test in `disclosure.test.ts` exercises two different
`chainId`s or two pool deployments at all.

**Trigger (conditional on verifier implementation).** A verifier service backed by a shared, multi-chain
`DisclosureChainSource` (or any `source` implementation that does not itself filter by
chain/contract) would accept a package whose `H2`/`H4`/`H6` hops reference transactions on a different
deployment than `pkg.domain` names, for a wallet active on both.

**Impact under the threat model.** "events from another pool or chain" (explicitly named in the
assignment's threat model). Severity depends entirely on verifier deployment choices this review cannot
observe from the SDK alone.

**Reproducer.** None written — exploitability is a property of a *caller's* `DisclosureChainSource`
implementation, which does not exist in this repo to instantiate against two deployments. The gap itself
(absence of any chain/contract field on the event types, and absence of any domain comparison in
`checker.ts`) is directly verifiable by reading the cited files.

---

## F4 — Circuit and SDK test suites validate the implementation against itself

**Claim.** Both test suites' claim to breadth ("18 tests," "9 captured spends," "reference parity")
rests entirely on fixtures produced by the same codebase (or its own vendored dependency), not on an
independent third-party reference. This is a scope/assurance limit, not a specific exploit, and it is
also the mechanism that let F1 go undetected.

**Evidence.**
- `armada-circuits/tests/disclosure.test.js` reads "reference parity" vectors from
  `armada-sdk/test/vectors` via `SDK_VECTORS` (default
  `path.join(ROOT, '..', 'armada-sdk', 'test', 'vectors')`) — i.e. the circuit is checked against the
  *sibling project's own* generated vectors, not an externally published EdDSA-Poseidon/BabyJubJub test
  vector set. (Partially mitigated: `armada-sdk` vendors `railgun-engine` for key derivation per the
  design doc, so the key-derivation piece traces to a third-party library; the *disclosure*-specific
  pieces — the nullifier/message/signature construction added by this PR — do not have an equivalent
  external anchor.)
- `armada-circuits/tests/fixtures/generated/*.json` (`shield.json`, `transfer-1x1.json`, …,
  `unshield-2x2.json`) are named "generated," consistent with being produced by the project's own capture
  tooling rather than captured from an independently-operated deployment.
- `armada-sdk/src/disclosure/disclosure.test.ts` loads its keyset from
  `armada-sdk/test/vectors/keyset-vectors.json`, "vector 0 (all-ones)" — a single, synthetic, author-
  chosen key, reused across every test in the file.
- F1 is a direct demonstration of the risk: the one fixture that exercises `H4` always pairs it with an
  `H2` hop on the same proof and transaction, so the missing `H4` check was never isolated by any test,
  despite 15+ named rejection tests covering every *other* check in the same file.

**Impact.** A systematic error shared between the circuit and the SDK (e.g. a consistent wrong ordering
of Poseidon inputs, or a wrong field constant) would satisfy every test in both suites while still being
wrong relative to the deployed verifier contract or any independent implementation. This is a coverage
statement, not a confirmed defect.

---

## Reviewed, nothing found

Documented per the context-building convention — these were checked and did not yield a finding:

- **EdDSA malleability / small-order keys** (`NullifierDisclosure.circom`): `BabyCheck` is applied to both
  `spendingPublicKey` and `signature[0..1]` (the `A` and `R8` points), and circomlib's
  `EdDSAPoseidonVerifier` internally enforces `S < subgroup order` (via `CompConstant`) and `8·A ≠ 0` (via
  the triple-`BabyDbl` + `IsZero` sequence) — the standard mitigation for small-order-key and
  S-malleability attacks, matching what `Transact.circom` already relies on. No path found by which an
  attacker lacking the real spending key could produce an accepted signature over a fresh,
  statement-specific message.
- **Hop reordering**: `hashHops` (`statement.ts`) computes `canonicalJson` over the hop *array*, which
  preserves element order (`[${value.map(canonicalJson).join(',')}]`) — only object keys are sorted.
  Reordering hops changes `hopsHash`, hence the signed `statement`, hence both EOA signatures and every
  disclosure proof's `statementHash`. Reordering without the real keys is not possible.
- **`statementHash = digest mod r` collision/bias** (ZKDocs random-sampling): the reduction from a
  256-bit keccak digest into the ~254-bit BN254 scalar field biases some residues by a factor of
  `(⌊2^256/r⌋+1)/⌊2^256/r⌋ ≈ 5/4`, but `statementHash` is a deterministic commitment of fully public
  statement content, not a value an adversary benefits from steering toward a biased/weak residue (unlike
  a Fiat–Shamir challenge the ZKDocs page is primarily concerned with). The relevant property is
  collision resistance between two *different* statements, which is the standard keccak-into-a-smaller-
  field birthday bound (~127 bits here), matching the design doc's own "~128-bit binding" claim. No
  sharper weakness found.
- **Builder key leakage**: `DisclosurePackage`/`DisclosureStatement`/`NoteOpening`/`DisclosureProofEntry`
  (`types.ts`) carry no `nullifyingKey`, viewing key or `spendingPrivateKey` field; `buildDisclosurePackage`
  (`builder.ts`) only derives `npk`/`masterPublicKey` (already public via the 0zk address) and serializes
  the five circuit public signals plus the Groth16 proof blob. (Caveat, not flagged as a bug: the builder
  serializes *every* opening in `params.openings` without filtering to the ones actually referenced by
  `params.hops` — that is caller-controlled input, not an SDK-internal leak.)
- **`SHIELD_ALLOCATION_TYPE` / `depositBinding` constant-and-formula cross-check**: `types.ts`'s
  `SHIELD_ALLOCATION_TYPE = 3` matches `armada-nitro/src/ArmadaNitroAdjudicator.sol`'s
  `SHIELD_ALLOCATION_TYPE = 3`; `checker.ts`'s `keccak256(AbiCoder.defaultAbiCoder().encode(['bytes32',
  'uint256'], [channelId, expectedHeld]))` matches the contract's `depositBinding(bytes32, uint256) =
  keccak256(abi.encode(channelId, expectedHeld))` exactly.
- **Every public input bound in the circuit**: `statementHash, masterPublicKey, nullifier, treeNumber,
  leafIndex` are each referenced inside the EdDSA-signed `Poseidon(5)` message
  (`NullifierDisclosure.circom`, the `msg` component); `masterPublicKey` and `nullifier` are additionally
  bound through the `NullifierOwnership` equality constraints; `leafIndex`/`treeNumber` are range-checked
  to 16 bits via `Num2Bits(16)` (preventing field-aliasing/wraparound). No under-constrained public signal
  found by manual constraint tracing (independent of the Picus-methodology concern in F2).

---

## Files read (coverage)

- `armada-laconic/engineering/16-selective-disclosure.md`, `engineering/09-architecture-decisions.md`
  (ADR-0020 only).
- `armada-circuits/src/operations/NullifierDisclosure.circom`, `src/main/main_disclosure.circom`,
  `src/operations/Transact.circom`, `tests/disclosure.test.js` (structure + fixture loading),
  `scripts/picus.sh`, `tests/picus/wiring.circom`, `tests/picus/control_minimal.circom`,
  `tests/picus/poseidon2.circom`, `node_modules/circomlib/circuits/eddsaposeidon.circom`,
  `node_modules/circomlib/circuits/babyjub.circom`.
- `armada-sdk/src/disclosure/{checker,builder,statement,circuit,serialize,types,chain-source}.ts`,
  `disclosure.test.ts`, `statement.test.ts`, `docs/guide/selective-disclosure.md`,
  `src/node/disclosure-artifacts.ts`, `src/prover/index.ts` (interface only),
  `src/sync/event-decoder.ts` (decoded-event shapes only).
- `armada-nitro/src/ArmadaNitroAdjudicator.sol` (`depositBinding`, `depositFromNote`,
  `_executeSingleAssetExit` region, `SHIELD_ALLOCATION_TYPE`).
- Veridise Picus source at pinned commit `138b151` (`README.md`, `NOTES.md`, `picus.rkt`,
  `picus/reader.rkt`, `picus/r1cs/r1cs-grammar.rkt`, `picus/circom/circom-parser.rkt`) and
  `docs.circom.io`'s "Signals & Variables" page — read to verify the exact semantics behind F2's claim,
  not part of the audited repos themselves.
- Not read in depth (out of the stated PR scope / pre-existing): `armada-circuits/tests/diff-test.js`,
  `armada-poc/contracts/privacy-pool/*` (relied on only via the spec's description of `Shield`/`Transact`/
  `Nullified`/`Unshield` event shapes, which the SDK's pre-existing `event-decoder.ts` already encodes).
