## REQ-18 — Hop H2: Note is spent

> "H2 · Note is spent | note → spend transaction | A disclosure proof that a published nullifier
> belongs to the note at a given leaf, for the disclosed `masterPublicKey`. | **Yes: the
> disclosure circuit**"
> — engineering/16-selective-disclosure.md, table "The hops"

**Verdict:** implemented · confidence: high

**Where enforcement lives — checker.ts, case `'H2'` (`L197-L206`):**

```ts
const proofErr = await verifyProof(proof, i);              // Groth16 proof verifies for (S, M, N, T, L)
const posErr = checkSpendLeaf(proof, opening, i);           // proof's (T, L) == opening's (tree, position)
const nullErr = checkNullifierEvent(proof, hop.nullifiedTxid, i); // N is in Nullified event of nullifiedTxid
```

`verifyProof` checks the Groth16 proof against `vkey` and that `statementHash`/`masterPublicKey`
match the package statement (`checker.ts:L116-L135`). `checkSpendLeaf` and
`checkNullifierEvent` (`checker.ts`) jointly realize "belongs to the note at a given leaf" and "a
published nullifier" — exactly the off-circuit half of REQ-03 for this hop. All three checks run
unconditionally for every `H2` hop.

**How the verdict was reached:** the three sub-clauses of the requirement (disclosure proof,
leaf binding, published-nullifier binding) each map to one of the three function calls, and all
three are exercised by `tests/disclosure.test.js`'s acceptance tests and by the SDK's "H2
spend-leaf linkage" test (`armada-sdk/src/disclosure/disclosure.test.ts:L294-L313`).
