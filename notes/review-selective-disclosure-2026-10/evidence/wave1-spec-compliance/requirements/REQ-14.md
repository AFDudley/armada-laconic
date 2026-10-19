## REQ-14 — Verification status: circomlib EdDSAPoseidonVerifier, "11 rejection tests"

> "circomlib `EdDSAPoseidonVerifier` | 11 rejection tests, including a forged signature for a
> small-order key and `S` plus the subgroup order | Tests only"
> — engineering/16-selective-disclosure.md, table "Verification status"

**Verdict:** partial · confidence: medium

**What this demands:** the test suite covering `EdDSAPoseidonVerifier`'s rejection behavior
contains 11 test cases, including (at least) a forged signature for a small-order key and an
`S + subgroup order` malleability case.

**Searched:**
- `test\('rejects:` in `armada-circuits/tests/disclosure.test.js` → 10 hits (not 11):
  `a nullifyingKey of another wallet`, `a nullifier of a different leaf`, `a masterPublicKey of
  another wallet`, `a signature by another spending key`, `a signature over another statement`,
  `each public input changed after signing`, `a malleable signature (S + subgroup order)`,
  `leafIndex or treeNumber outside 16 bits, even when signed`, `points off the curve`, `a forged
  signature for a small-order spending key`.
- Of those 10, only a subset target `EdDSAPoseidonVerifier` specifically (the others target the
  circuit's own `===` wiring or `Num2Bits`): `a signature by another spending key`, `a signature
  over another statement`, `each public input changed after signing` (3 sub-cases), `a malleable
  signature`, `a forged signature for a small-order spending key` (2 sub-cases) — 8 EdDSA-specific
  `rejects()` calls across 5 `test()` blocks.
- `node_modules/circomlib/test/eddsaposeidon.js` (the vendored circomlib's own unit tests for this
  template) → 3 `it(...)` blocks total (`Sign a single number`, `Detect Invalid signature`, `Test
  a dissabled circuit with a bad signature`), not 11.

Both the forged-small-order-key case and the `S + subgroup order` case named in the requirement
are present (`tests/disclosure.test.js:263-277` and `:241-245`), so the *content* of the claim is
correct.

**How the verdict was reached:** `partial` rather than `contradicted`: the two named example cases
are real and correctly described, but the count "11" does not match either a plausible reading
(10 named `test()` blocks) or a stricter reading (8 EdDSA-specific assertions, or 3 in circomlib's
own suite). This is a documentation-count drift with no behavioral consequence — the tests
themselves are real and do what they claim — so it is Low severity per the divergence rubric, not
a security finding.

**Open questions:** which exact counting method produced "11" is not derivable from the text
alone; flagged as a doc-accuracy note for the page-16 author.
