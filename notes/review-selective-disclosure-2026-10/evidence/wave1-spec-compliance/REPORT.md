# Spec-to-code compliance review — selective disclosure (ADR-0020)

Specification: `engineering/16-selective-disclosure.md` and ADR-0020 in
`engineering/09-architecture-decisions.md`.
Code: `armada-circuits` (`NullifierDisclosure.circom`, `main_disclosure.circom`,
`tests/disclosure.test.js`, `tests/picus/`, `scripts/picus.sh`) and `armada-sdk`
(`src/disclosure/`, `src/node/disclosure-artifacts.ts`).
Method: `spec-to-code-compliance` (manual application — see `local://audit-method.md`). 37
requirements checked, one file per requirement under `requirements/`.

## Verdict tally

| Verdict | Count | REQs |
|---|---|---|
| implemented | 31 | 01,02,04,05,06,07,08,09,10,11,12,13,15,16,17,18,19,21,22,23,24,26,27,28,31,32,33,34,35,36,37 |
| partial | 3 | 03, 20, 25 |
| absent | 1 | 29 |
| undecidable | 2 | 14 (count mismatch, Low), 30 (descriptive, no implementation demand) |

## Headline finding: hop H4 does not verify its nullifier against the deposit transaction (REQ-20, REQ-03, REQ-25)

Three readings of the same spec text — the circuit table's off-circuit rule ("the verifier
checks off-circuit ... that N is in the pool's Nullified event for tree T"), the H4 hop-table
row, and the worked example's verifier check 3 ("n1 is in the Nullified event **of the deposit
transaction**") — all require that, before a disclosure proof can support an "H4: spend funds a
channel" claim, the proof's nullifier must be shown to appear in a `Nullified` event inside the
same transaction that carries the `Deposited` event.

`armada-sdk/src/disclosure/checker.ts`'s `H2` case (`checkSpendLeaf` + `checkNullifierEvent`
against `hop.nullifiedTxid`) and `H6` case (the same two helpers against `hop.unshieldTxid`) both
do this. The `H4` case (`checker.ts:L222-L241`) does not: it verifies the Groth16 proof
(`verifyProof`) and the `Deposited`/`depositBinding` facts, but never calls
`checkNullifierEvent`/`checkSpendLeaf` against `hop.depositTxid`.

`ArmadaNitroAdjudicator.sol`'s `depositFromNote` (`L93-L128`) calls `pool.transact(transactions)`
*inside* the same call, before emitting `Deposited` — confirming that in the real system the
funding spend's `Nullified` event and the `Deposited` event are the same transaction, exactly as
the page's own worked example narrates it ("the transaction publishes the nullifier n1 and the
binding to channel C" — singular transaction). The missing check is not a stylistic gap; it is
the one fact that ties "a spend happened" to "that spend is what funded this channel."

**Consequence:** the disclosure circuit places no constraint tying a proof's `(leafIndex,
treeNumber)` to a real pool commitment (that binding is exclusively the off-circuit job this
check performs). A prover can mint a fresh, self-controlled key pair, choose any 16-bit
`(leafIndex, treeNumber)`, and produce a valid Groth16 disclosure proof for it without ever
having used the pool. A `DisclosurePackage` containing only such a proof inside an `H4` hop,
naming the `channelId`/`expectedHeld`/`asset` of *any* real `Deposited` event on chain (funded by
anyone), passes `checkDisclosurePackage` with `ok: true` — a false "my note funded this channel"
claim from an actor with no real history in the pool, matching the stated threat model's "claims
another wallet's ... channel deposit" directly. See REQ-20 for the full trace, the exact missing
call, and the suggested one-line fix (`checkNullifierEvent(proof, hop.depositTxid, i)`, mirroring
`H2`/`H6`).

No test in either repo isolates `H4` to exercise this path; the only `H4` test instance always
pairs it with an `H2` hop pointing at the identical transaction id, which masks the gap (the `H2`
check fails first on any mutation that would have exposed it). See REQ-20's "Searched" section.

## Low-severity documentation drift (REQ-14)

The "Verification status" table's claim of "11 rejection tests" for circomlib's
`EdDSAPoseidonVerifier` does not match the actual count under any counting method tried: 10 named
`test('rejects: ...')` blocks exist in `tests/disclosure.test.js`, of which 8 `rejects()` call
sites specifically target the EdDSA verifier; circomlib's own vendored test suite has 3 `it(...)`
blocks. The two example cases named in the row (forged small-order key, `S + subgroup order`) are
both real and correctly described. No behavioral consequence; a page-16 wording fix.

## Absent, explicitly-optional feature (REQ-29)

Verifier check 7 in the worked example ("Optional: the co-signed states ... EIP-1271") has no
representation anywhere in `armada-sdk/src/disclosure/` — no package field, no checker logic.
Since the spec itself marks this check optional and it is *additional* attribution evidence (not
part of the custody proof itself), its absence does not weaken any accepted package's custody
claim; it is a feature gap against the worked example, not a security finding.

## Reverse direction — code behavior no target document mentions

Page 16 explicitly lists "the package format and its encoding (EIP-712 domain, statement schema,
serialization)" under **Open**, i.e. undecided. `armada-sdk/src/disclosure/` resolves this with
concrete choices that neither `engineering/16-selective-disclosure.md` nor ADR-0020 describes
(they are documented only in the SDK's own `docs/guide/selective-disclosure.md`, a third
document outside this review's named specification):

- **EIP-712 domain**: `name = "ArmadaSelectiveDisclosure"`, `version = "1"`, scoped by
  `chainId`/`verifyingContract` (`statement.ts`, `DISCLOSURE_EIP712_DOMAIN_NAME/_VERSION`,
  `eip712Domain`).
- **`Statement`/`Endpoint` EIP-712 type schema**: `masterPublicKey, verifier, context, issuedAt,
  endpoints[], hopsHash` (`statement.ts`, `DISCLOSURE_EIP712_TYPES`).
- **`hopsHash`**: `keccak256` of a canonical (key-sorted at every depth) JSON encoding of the hop
  list (`statement.ts`, `canonicalJson`/`hashHops`) — binds the full hop list into the fixed-size
  EIP-712 schema without naming it a required disclosure-protocol mechanism anywhere in page 16.
- **`statementHash = digest mod r`**: the EIP-712 digest (a 256-bit keccak output) reduced into
  the BN254 scalar field for the circuit's `statementHash` public input (`statement.ts`,
  `statementHashField`). Page 16 only says "S is the hash of the full disclosure statement";
  the specific reduction mechanism, and the SDK's own ~128-bit-binding security note about it, are
  undocumented in the named specification.
- **`issuedAt`**: a Unix-seconds field on `DisclosureStatement`, signed as part of S
  (`types.ts`, `statement.ts`), with no replay-window or expiry semantics attached anywhere in
  the code or in page 16/ADR-0020 (neither document mentions a time-bound disclosure at all).
- **`DISCLOSURE_PACKAGE_VERSION = 1`** and the `package-version` check (`types.ts`, `checker.ts`)
  — a schema-versioning mechanism, undocumented in the target spec.
- **18 named `DisclosureCheckName` values** (`checker.ts`) returned on first failure — the
  specific enumeration and "first failing check" contract (matching the SDK's own D6, but not
  named in page 16/ADR-0020).
- **The optional `shield-sender` check in H1** (`checker.ts`): compares the `Shield` transaction's
  sender against the claimed endpoint account when the chain source supplies it. Not named in the
  H1 table row or the "Key facts"/"Key binding" sections; silently skipped when
  `DisclosureChainSource.senderOf` returns `undefined` (e.g. `InMemoryDisclosureChainSource` with
  no `senders` map supplied). Not a spec violation (spec states no requirement to violate), but
  worth the SDK team's attention since its coverage is conditional on caller-supplied data.

None of the above contradicts page 16/ADR-0020; page 16 explicitly defers the package-format
decision, so these are the SDK authors' resolution of an acknowledged open question, not a
divergence from a stated requirement.

## Independently re-verified claims

- Circuit size (REQ-11): compiled `main_disclosure.circom` directly with the repo's pinned
  `circom2` at `--O1`; got exactly 8,248 non-linear / 1,835 linear constraints, matching both
  page 16 and `armada-circuits/README.md`.
- Picus results (REQ-12, REQ-13): read the committed `build/picus/*.log` files rather than
  trusting the prose; `wiring` (strong safety) and `poseidon2/3/5` (weak safety) all report
  "properly constrained" / exit code 8; the `control_minimal` negative control reports
  "underconstrained" / exit code 9 with a concrete counterexample, confirming the harness
  discriminates rather than always reporting safe.
- Test counts (REQ-15): counted 18 `test()` blocks in `tests/disclosure.test.js`, matching the
  page's claim exactly; the "9 captured spends" sub-claim is enforced by the test's own
  `assert.ok(notes >= 9, ...)`, not just asserted in prose.
- `SHIELD_ALLOCATION_TYPE` cross-check (REQ-21): `armada-sdk/src/disclosure/types.ts`'s
  `SHIELD_ALLOCATION_TYPE = 3` matches `ArmadaNitroAdjudicator.sol`'s
  `uint8 constant SHIELD_ALLOCATION_TYPE = 3` exactly.
- `depositBinding` formula cross-check (REQ-26): the checker's
  `keccak256(AbiCoder.defaultAbiCoder().encode(['bytes32','uint256'], [channelId, expectedHeld]))`
  matches the Solidity `keccak256(abi.encode(channelId, expectedHeld))` exactly, same argument
  order and ABI types.

## Non-goals observed (not verdicted)

No instance was found where both page 16/ADR-0020 and the code are silent about the same
behavior — every finding above is anchored to specific spec text.

## Scope note

Per the method's rules, this review is read-only; no repository file was modified. Compilation
(`circom2`) was run only to independently re-verify a stated constraint-count claim (REQ-11), into
a throwaway directory now removed.
