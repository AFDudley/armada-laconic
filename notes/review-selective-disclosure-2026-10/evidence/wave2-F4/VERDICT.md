# fp-check verdict — Finding F4 (test suites validate the implementation against itself)

## Step 0 — Claim restated

F4 claims: both the circuit test suite (`armada-circuits/tests/disclosure.test.js`) and the SDK test
suite validate the selective-disclosure code against fixtures produced by the project's own codebase,
with no independent third-party reference vectors. F4 names the disclosure-specific pieces (the signed
message, `statementHash`, `hopsHash`, package checks) as having no external anchor at all. F4 also
claims this gap "is the mechanism that let F1 go undetected." F4 self-labels as "a scope/assurance
limit, not a specific exploit," not a vulnerability claim.

This is a test-design/coverage finding, not a bug with an attacker-reachable path. The fp-check gates
for reachability, impact, PoC, and math bounds apply to vulnerability claims; the equivalent question
here is whether the factual premise is accurate and whether the finding's own characterization of
severity and causation holds up.

## Provenance table (built by reading every vector/fixture header and every assertion)

| File | Primitive | Source | Disclosure-specific? |
|---|---|---|---|
| `armada-sdk/test/vectors/keyset-vectors.json` | key derivation | `engineVersion: 9.6.0` (stock Railgun engine); one vector has `testnetCrosscheck: byte-identical` against a live testnet wallet | No — pre-existing Railgun key derivation |
| `nullifier-vectors.json` | `nullifier = Poseidon(nullifyingKey, leafIndex)` | engine 9.6.0 | No |
| `poseidon-vectors.json` | Poseidon-BN254 | engine 9.6.0 | No |
| `merkle-vectors.json` | depth-16 Poseidon merkle | engine 9.6.0 | No |
| `npk-vectors.json` | note public key | engine 9.6.0 | No |
| `commitment-vectors.json` | commitment hash | engine 9.6.0 | No |
| `boundparams-hash-vectors.json` | `boundParamsHash` | engine 9.6.0, stock on-chain keccak formula | No |
| `transaction-struct-vectors.json` | `TransactionStructV2` | engine 9.6.0, real Groth16 proof | No |
| `note-ciphertext-vectors.json` | ECIES note ciphertext | engine 9.6.0 | No |
| `eddsa-spend-auth-vectors.json` | EdDSA spend-authorization signing (pre-existing Transact message) | engine 9.6.0 | No — this is the Transact signing scheme, reused as a convenient already-signed vector by the circuit test, not disclosure logic |
| `armada-circuits/tests/fixtures/generated/*.json` | shield/transfer/unshield witnesses | captured `chainId: 31337` (local Hardhat/Anvil chain), `timestamp: 2026-06-30`, own `txHash` — a local devnet run of the project's own stack, not an independently-operated deployment | No |
| — (no vectors file exists) | `statementHash` / `statementDigest` | none | **Yes** — `statement.test.ts` only checks `field < SNARK_SCALAR_FIELD` and `field == digest % r`, both tautologies by definition, no external value compared |
| — (no vectors file exists) | `hopsHash` | none | **Yes** — only order-sensitivity is checked (`hashHops(hops) !== hashHops(reversed)`), no reference value |
| EIP-712 signing/verification mechanics | signature round trip | `ethers.js` `signTypedData`/independent verifier — a third-party library implementing the public EIP-712 standard | Partial — the *signing mechanics* have a real external anchor (ethers); the *statement schema* (which fields, in what order) does not |
| — (no vectors file exists) | the circuit's `msg = Poseidon(5)` over `statementHash, masterPublicKey, nullifier, treeNumber, leafIndex` | none | **Yes** — `disclosure.test.js`'s "accepts"/"rejects" tests build this message in JS and feed it to the circuit; passing proves circuit and JS agree with each other, not that either matches an external spec |
| — (no vectors file exists) | `checker.ts` package-level checks (`depositBinding` keccak match, chain-event cross-checks) | none (verified only by matching the Solidity formula by manual read, per `AuditAdversarial/FINDINGS.md` "Reviewed, nothing found" §SHIELD_ALLOCATION_TYPE) | **Yes** |

Two corrections to F4's own wording:

1. F4 says the test suites' breadth claim "rests **entirely** on fixtures produced by the same
   codebase." That overstates it. Nine of the ten vector files (everything except the disclosure-specific
   `msg`/`statementHash`/`hopsHash`/package checks) trace to `@railgun-community`'s stock engine 9.6.0,
   a maintained third-party library, and the circuit test cross-checks them using plain `circomlibjs`
   0.1.7 (armada-circuits) against values the SDK generated via the separately-maintained
   `@railgun-community/circomlibjs` 0.0.8 fork (armada-sdk/vendor) — two independently-versioned
   implementations of the same EdDSA-Poseidon/BabyJubJub math, a real (if family-related) differential
   check. One keyset vector additionally carries a live-testnet byte-identical crosscheck.
2. F4 says the vector/fixture gap "is the mechanism that let F1 go undetected." Verified false as stated.
   F1 (missing H4 ownership check) is a missing-constraint bug: no reference-vector set, internal or
   externally sourced, would catch it, because reference vectors only confirm "does computing X give
   value V," not "is X constrained in the circuit." `disclosure.test.js` already has 15+ dedicated
   adversarial rejection tests (`rejects: a nullifyingKey of another wallet`, etc., lines 200–281)
   targeting exactly this failure class, independent of vector provenance. F1's root cause, per
   `AuditAdversarial/FINDINGS.md` itself, is that the only fixture exercising H4 always co-occurs with
   H2 on the same proof/transaction — a fixture **composition/isolation** gap. An externally-captured
   fixture set with the same H4+H2 co-occurrence would have hidden the same bug. Provenance (internal
   vs external) and isolation (which hops are combined in one fixture) are different axes; F4 conflates
   them.

## Executed evidence

```
$ cd armada-sdk && npx vitest run src/disclosure/statement.test.ts
 ✓ src/disclosure/statement.test.ts  (4 tests) 31ms
 Test Files  1 passed (1)
      Tests  4 passed (4)
```
Confirms by execution (not just reading) that the four `statement.test.ts` assertions are exactly the
self-consistency/library-cross-check pattern read in source: no external `statementHash`/`hopsHash`
value is compared against anywhere in the file.

## Devil's advocate

1. Is the core factual premise (disclosure-specific constructs have no external reference) real, or
   pattern-matching? Real — verified directly: `grep`/read of all 10 `armada-sdk/test/vectors/*.json`
   headers shows none covers `statementHash`, `hopsHash`, the disclosure `msg` formula, or
   `checker.ts` checks; `statement.test.ts` and `disclosure.test.js` were read line by line and contain
   no comparison against a value not computed by the same PR's own code (except the EIP-712 mechanics
   check against `ethers`).
2. Is "no third-party vector for a brand-new construction" itself a defect? No single other system
   implements this project's own `statementHash`/`hopsHash`/disclosure-message scheme — by
   construction, no externally-published vector set for it can exist until a second independent
   implementation is built. Demanding one is not an actionable fix; the available alternatives are
   formal verification (the circuit side, Picus — a separate finding, F2) and adversarial/manual
   property tests (which the suite already has, e.g. the "rejects" block).
3. Does the finding claim a concrete attacker-reachable impact? No. F4 explicitly says "not a specific
   exploit" and "not a confirmed defect." It fails the fp-check Reachability, Impact, PoC and Math
   Bounds gates because it never attempted to pass them — it is self-scoped as a methodology note, not
   a vulnerability.
4. Am I dismissing a real assurance gap because the exploit is absent? No concrete consequence is
   demonstrated or claimed by F4 for this gap beyond the already-refuted F1 linkage. The coverage
   observation for the disclosure-specific pieces is real and worth recording as methodology, but it is
   not a vulnerability and its headline causal claim (tying it to F1) does not hold under inspection.

## Verdict

**F4 — FALSE POSITIVE as a vulnerability finding.**

Gate 2 (Reachability) FAIL: no attacker-controlled path; F4 identifies no trigger or attacker action,
only a property of the test suite's design.

Gate 3 (Real Impact) FAIL: F4 itself states "not a confirmed defect" and names no RCE/privesc/info-
disclosure consequence; the one impact scenario offered ("a systematic error shared between circuit and
SDK would satisfy every test") is hypothetical and unconnected to F1, the only concrete example cited.

Gate 4 (PoC Validation) FAIL: no PoC was or could be offered — there is no attack to demonstrate.

The underlying factual observation is correct and narrower than stated: the disclosure-specific
constructs (`statementHash`, `hopsHash`, the circuit's `msg` Poseidon(5), `checker.ts` package checks)
have no reference-vector coverage, only self-consistency tests and (for EIP-712 signing mechanics only)
a real third-party library cross-check. This is the expected, unavoidable state for a brand-new
construction, not a remediable test-design defect — there is no third party to source vectors from.
The 9 non-disclosure vector files are anchored to stock Railgun engine 9.6.0 (one with a live-testnet
crosscheck), contradicting F4's "rests entirely" framing. F4's causal claim that this gap "let F1 go
undetected" does not hold: F1 was a missing-constraint bug exposed only by fixture/hop-isolation
coverage, an axis independent of vector provenance.

**Recommendation for the report, if F4 is carried forward at all:** reframe as a non-severity
methodology note — "the disclosure-specific message/statement construction has no external reference
vectors (expected for a novel scheme); assurance for it rests on the formal circuit check (F2/Picus)
and the suite's adversarial rejection tests, not on vector provenance" — and drop the F1 causal claim.

## Cleanup

No repository file was edited. `armada-sdk` test was run read-only (`vitest run`, no writes).
`git status --porcelain` in `armada-sdk` is clean. `armada-circuits` shows untracked files
(`.scratch_picus_f2/`, `tests/picus/wiring_public.circom`) from a sibling agent's F2 work, not touched
by this task.
