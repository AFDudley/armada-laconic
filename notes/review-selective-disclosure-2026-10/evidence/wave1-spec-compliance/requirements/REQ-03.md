## REQ-03 — Circuit table: leafIndex/treeNumber range + off-circuit location/event check

> "`leafIndex` L, `treeNumber` T | public | Each fits in 16 bits. The verifier checks off-circuit
> that the disclosed commitment is at (T, L), and that N is in the pool's `Nullified` event for
> tree T."
> — engineering/16-selective-disclosure.md, table "The disclosure circuit"

**Verdict:** partial · confidence: high

**What this demands of an implementation:** two independent properties.
1. In-circuit: `leafIndex` and `treeNumber` each fit in 16 bits.
2. Off-circuit (the verifier/checker's job, for *every* use of a disclosure proof): confirm the
   disclosed commitment sits at `(T, L)`, and confirm `N` appears in the pool's `Nullified` event
   for tree `T`.

**Where enforcement lives (in-circuit, part 1):**

```circom
// NullifierDisclosure.circom:L68-L71
component leafBits = Num2Bits(16);
leafBits.in <== leafIndex;
component treeBits = Num2Bits(16);
treeBits.in <== treeNumber;
```
Unconditional, confirmed by `rejects: leafIndex or treeNumber outside 16 bits, even when signed`
(`tests/disclosure.test.js:247-251`). Part 1 is **implemented**.

**Where enforcement lives (off-circuit, part 2) — armada-sdk/src/disclosure/checker.ts:**

Every use of a disclosure proof funnels through `verifyProof` (`checker.ts:L116-L135`), which
verifies the Groth16 proof and the statement/M binding, but never checks that `N` is in a
`Nullified` event. That check is instead bolted onto specific hop cases:

- `H2` (`checker.ts:L197-L206`) calls `checkSpendLeaf` (commitment location) and
  `checkNullifierEvent(proof, hop.nullifiedTxid, i)` — **does** perform part 2, against
  `hop.nullifiedTxid`.
- `H6` (`checker.ts:L287-L300`) calls the same two helpers against `hop.unshieldTxid` — **does**
  perform part 2.
- `H4` (`checker.ts:L222-L241`) calls only `verifyProof`, then checks the `Deposited` event and
  `depositBinding`. It never calls `checkSpendLeaf` or `checkNullifierEvent`. **Does not** perform
  part 2 at all: nothing ties the disclosed `(T, L)` to a real commitment, and nothing ties `N` to
  any `Nullified` event, for an H4-referenced proof.

**Paths walked:**
- `checkHop('H2', …)` → `verifyProof` → `checkSpendLeaf` → `checkNullifierEvent`. ✓ both parts of
  the off-circuit check run.
- `checkHop('H6', …)` → same chain. ✓.
- `checkHop('H4', …)` → `verifyProof` only → `depositsIn`/`deposit-binding`. ✗ — the gap. A
  disclosure proof can be referenced by an `H4` hop alone (the package schema does not require a
  matching `H2`/`H6` hop for the same proof id), in which case the mandatory off-circuit check in
  this requirement never runs for that proof.

**Searched:**
- `checkNullifierEvent` → 3 call sites: `checker.ts:L206` (H2), `L296` (H6); no call inside the
  `H4` case (`L222-L241`).
- `checkSpendLeaf` → 2 call sites (H2, H6); none in H4.
- any requirement elsewhere in `checker.ts` that a package's `H4` hop share a proof id with an
  `H2`/`H6` hop → none; `proofById` is a flat map keyed only by `DisclosureProofEntry.id`, and each
  hop case resolves it independently.
- `src/disclosure/disclosure.test.ts` (SDK integration tests) → the only `H4` test instance
  (`L119`) always uses the same `TX.deposit` value that the paired `H2` hop uses as
  `nullifiedTxid` (`L118`), and hops run in array order with H2 before H4, so a missing-nullifier
  test (`rejects a nullifier that is not in the Nullified event`, `L226-L228`) fails at the `H2`
  check before `H4` is ever reached. No test isolates `H4` to show it independently enforces part 2.

**How the verdict was reached:** `partial`, not `absent`, because part 2 of the requirement is
genuinely enforced on two of the three code paths that consume a disclosure proof (H2, H6). It is
not `contradicted` either, because nothing actively asserts the wrong thing — it is simply missing
on the third path. See REQ-20 (H4 evidence) for the exploit this enables.

**Open questions:** none; the gap is directly readable from the `H4` case body.
