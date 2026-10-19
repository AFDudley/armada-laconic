## REQ-15 — Verification status: whole circuit vs. the statement, "18 tests"

> "The whole circuit against the statement above | 18 tests: parity with the armada-sdk engine
> vectors, the 9 captured spends, acceptance, rejection and a Groth16 round trip | Tests only"
> — engineering/16-selective-disclosure.md, table "Verification status"

**Verdict:** implemented · confidence: high

**Verification performed:** counted `test(` blocks in `armada-circuits/tests/disclosure.test.js`:
3 reference-parity tests, 1 ownership/real-spends test, 3 acceptance tests, 10 rejection tests,
1 Groth16 round-trip test = 18. Exact match.

The "9 captured spends" claim is independently asserted inside the test itself:
`assert.ok(notes >= 9, 'expected the 9 captured spent notes, got ${notes}')`
(`tests/disclosure.test.js`, test `ownership: reproduces nullifier and note public key of every
captured spend`), reading from `fs.readdirSync(FIXTURES).filter(/^(transfer|unshield)-.*\.json$/)`.

**How the verdict was reached:** direct count of `test()` blocks matches 18; the 9-spend sub-claim
is enforced by the test's own assertion, not just descriptive text.
