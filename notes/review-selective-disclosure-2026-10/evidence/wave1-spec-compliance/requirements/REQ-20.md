## REQ-20 — Hop H4: Spend funds a channel

> "H4 · Spend funds a channel | spend → armada-nitro channel | `depositFromNote` requires
> `adaptParams == depositBinding(channelId, expectedHeld)`
> (`armada-nitro/src/ArmadaNitroAdjudicator.sol`). The spend proof commits to `boundParamsHash`,
> so the spending key authorized this channel. The adjudicator emits `Deposited(channelId, asset,
> held)`. | No"
> — engineering/16-selective-disclosure.md, table "The hops"

**Verdict:** partial · confidence: high

**What this demands of an implementation:** the "spend → channel" link requires that the *same*
spend (the one whose nullifier a disclosure proof proves) is the spend that funded this
channel — not merely that some deposit event with matching parameters exists somewhere on chain.
`depositFromNote` (`armada-nitro/src/ArmadaNitroAdjudicator.sol:L93-L128`) calls `pool.transact(transactions)`
*inside the same call* (`L121`) before emitting `Deposited` (`L128`), so the `Nullified` event for
the funding spend and the `Deposited` event are published in the **same transaction**. The
worked example on the same page confirms this reading directly: "Alice spends N1 through
`depositFromNote(C, expectedHeld, [tx])` ... **The transaction publishes the nullifier n1 and the
binding to channel C**" (engineering/16-selective-disclosure.md, "Example" step 2 — one
transaction, both facts).

**Where enforcement lives — armada-sdk/src/disclosure/checker.ts, case `'H4'` (`L222-L241`):**

```ts
case 'H4': {
  const proof = proofById.get(hop.proof);
  if (proof === undefined) return failure('hop-reference', ...);
  const proofErr = await verifyProof(proof, i);
  if (proofErr !== null) return proofErr;
  const deposit = params.source.depositsIn(hop.depositTxid).find((d) => addressesEqual(d.channelId, hop.channelId));
  if (deposit === undefined) return failure('deposited-event', ...);
  const binding = keccak256(AbiCoder.defaultAbiCoder().encode(['bytes32', 'uint256'], [hop.channelId, BigInt(hop.expectedHeld)]));
  if (BigInt(deposit.adaptParams) !== BigInt(binding)) return failure('deposit-binding', ...);
  if (!addressesEqual(deposit.asset, hop.asset)) return failure('deposited-event', ...);
  transactionIds.add(hop.depositTxid);
  channelIds.add(hop.channelId);
  return null;
}
```

`verifyProof` checks only that the Groth16 proof is valid for the claimed `(statementHash, M, N,
T, L)` — it never checks that `N` was ever nullified on chain at all (see REQ-03). Nothing in the
`H4` case calls `checkSpendLeaf` or `checkNullifierEvent` against `hop.depositTxid` (contrast with
`H2`, REQ-18, and `H6`, REQ-22, which both call these helpers against their own transaction id).
`hop.depositTxid` is accepted purely because some `Deposited` event exists there with matching
`channelId`/`expectedHeld`/`asset` — a fact that is public and independent of who actually funded
it.

**Consequence (why this is not merely academic):** the disclosure circuit places no constraint
tying `(nullifyingKey, leafIndex)` to a real commitment (REQ-03); a prover can generate a fresh
`(Ax, Ay, nullifyingKey)` they alone control, pick any `leafIndex`/`treeNumber` within 16 bits,
derive `N = Poseidon(nullifyingKey, leafIndex)`, and produce a valid Groth16 proof for `(S, M, N,
T, L)` without ever having interacted with the pool. A package containing only this proof in an
`H4` hop, naming the `channelId`/`expectedHeld`/`asset` of **any** real `Deposited` event
(deposited by anyone, at any time), passes `checkDisclosurePackage` with `ok: true` and
`learned.channelIds` including that channel — i.e. the checker accepts a false "my note funded
this channel" claim from an actor with no real history in the pool at all. This matches the
stated threat model directly: "A dishonest prover who wants a verifier to accept a chain of
custody that is false: claims another wallet's ... channel deposit" (local://audit-method.md,
"Threat model to apply").

Even for an honest-looking package that *does* pair the `H4` hop with a same-id `H2` hop
(the pattern every example and test in this repo uses, see "Searched" below), nothing requires
`H2`'s `nullifiedTxid` to equal `H4`'s `depositTxid`. A prover who genuinely owns a note and has
genuinely spent it (so `H2` passes against the real spend transaction) can still name an
*unrelated* `depositTxid` in the paired `H4` hop — any transaction with a `Deposited` event for
some `channelId`/`expectedHeld` they did not fund — and the package still passes.

**Paths walked:**
- `checkHop('H4', …)` → `verifyProof` → `depositsIn` → `deposit-binding` → `deposited-event` →
  return `null` (pass). No path inside this case ever consults `nullifiersIn`. ✗ — the gap.

**Searched:**
- `checkNullifierEvent` in `checker.ts` → 2 call sites (`H2` L206, `H6` L296); 0 inside the `H4`
  case body (`L222-L241`).
- `checkSpendLeaf` in `checker.ts` → 2 call sites (`H2`, `H6`); 0 in `H4`.
- any cross-hop equality check between an `H2`/`H6` hop's transaction id and an `H4` hop's
  `depositTxid`, keyed by shared `proof` id → 0 hits; `proofById` only caches that a given proof id
  verified once (`verifiedProofs` set, `checker.ts:L117-L135`), it does not record or compare
  *which transaction* that verification happened against.
- `armada-sdk/src/disclosure/disclosure.test.ts` for a dedicated `H4` unit test (parallel to the
  "H2 spend-leaf linkage" and "H3 output-note opening" `describe` blocks) → 0 hits; `H4` appears
  only once, inside the combined Aave-hub example (`L119`), always with
  `depositTxid === nullifiedTxid === TX.deposit` for the paired `H2` hop (`L118`). The one test
  that removes a `Nullified` record for `TX.deposit` (`rejects a nullifier that is not in the
  Nullified event`, `L226-L228`) fails at the `H2` hop (array index before `H4`), so it never
  reaches — and therefore never exercises — `H4`'s own lack of this check.

**How the verdict was reached:** `partial`, not `contradicted`, because the deposit-binding and
`Deposited`-event checks that *are* present are correct and faithfully implement the first two
clauses of the requirement. It is not `absent` either, since part of the requirement (deposit
binding) is enforced. What is missing is the clause that ties the spend to *this* channel-funding
transaction specifically, which both the circuit table's off-circuit rule (REQ-03) and the page's
own worked example (REQ-25) require and which `H2`/`H6` both implement for their own transactions.

**Suggested fix:** in the `H4` case, after `verifyProof` succeeds, call `checkNullifierEvent(proof,
hop.depositTxid, i)` (the same helper `H2`/`H6` already use), so the proof's nullifier must appear
in a `Nullified` event inside `hop.depositTxid` itself — the transaction that also carries the
`Deposited` event.
