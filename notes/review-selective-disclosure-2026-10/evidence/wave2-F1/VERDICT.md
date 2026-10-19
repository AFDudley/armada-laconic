# fp-check verdict — Candidate F1 (H4 nullifier-event gap)

Reviewer: `RefuteF1H4` (did not produce this finding).
Target: `armada-sdk/src/disclosure/checker.ts`, branch `selective-disclosure`, H4 case of `checkHop`.
Route: Standard verification (single component, clear claim, no concurrency, straightforward data flow).

## Step 0 — Claim restated

The `H4` case of `checkDisclosurePackage` proves "a spend funds channel C" with two checks only:
(1) a `Deposited` event for `C` exists in the named transaction with `adaptParams ==
depositBinding(channelId, expectedHeld)`, and (2) the Groth16 disclosure proof verifies for the
prover's own wallet. It never checks that the proof's nullifier is in that transaction's `Nullified`
event. `channelId`, `expectedHeld`, and `asset` are all public chain data, so a prover can pair a valid
proof for ANY of their own nullifiers with a DIFFERENT wallet's real `Deposited` event and the check
passes. The verifier is told a false fact: that the prover's wallet funded channel C.

Threat model match: "a dishonest prover who wants a verifier to accept a chain of custody that is
false: claims another wallet's ... channel deposit" (audit-method.md).

## Step 1 — Data flow

```
Source: hop.channelId / hop.expectedHeld / hop.asset / hop.depositTxid — attacker-chosen, Trust: untrusted
Path: checkHop('H4') [checker.ts:237]
  → verifyProof(proof, i) [checker.ts:113-136]       — checks proof.statementHash, proof.masterPublicKey
                                                         match the package's OWN statement, and the Groth16
                                                         proof verifies for the proof's OWN 5 public signals
                                                         (statementHash, masterPublicKey, nullifier, treeNumber,
                                                         leafIndex). Never reads hop.depositTxid.
  → params.source.depositsIn(hop.depositTxid).find(d => channelId match) [checker.ts:242-247]
  → binding = keccak256(abi.encode(channelId, expectedHeld)); compare to deposit.adaptParams [248-250]
  → asset match [252-254]
  → accept: transactionIds.add, channelIds.add, return null [255-257]
```

Validation points: `proof-statement-hash`, `proof-master-public-key`, `proof-verification`,
`deposited-event`, `deposit-binding`. None of these reads `params.source.nullifiersIn(...)` or
compares `proof.nullifier`/`proof.treeNumber` against anything the deposit transaction published.

Sibling hops H2 (line 212) and H6 (line 291), which also consume a disclosure proof, both call
`checkNullifierEvent(proof, txid, hopIndex)` (checker.ts:313-321), which does read
`params.source.nullifiersIn(txid)` and requires the proof's `(treeNumber, nullifier)` to appear there.
H4 has no analogous call. Confirmed by a full read of `checker.ts` (341 lines, single file, no other
H4-related logic elsewhere in the file).

No 3+ trust boundaries, no callbacks/async ambiguity. Standard route confirmed, no escalation.

## Step 2 — Exploitability

**Attacker control.** The attacker controls every field of the H4 hop: `proof` (any proof over their
own wallet's own note, including a trivial/dust-value note they already hold), `channelId`,
`expectedHeld`, `asset`, `depositTxid`. `channelId`/`expectedHeld`/`asset` are public on-chain values
(or derivable from the public `Deposited` event itself). The attacker needs zero privileged
information about the victim's deposit — they read the event, they don't produce it.

**Reachability.** Demonstrated by execution (below): a crafted package reaches `checkHop('H4')`,
passes every check present in the code, and returns `ok: true`.

## Step 3 — Impact

Real security impact: the checker's output (`result.ok`, `result.learned.channelIds`) is the trust
decision a downstream verifier acts on (per the spec's worked example, "this withdrawal's principal
passed through channel C"). Forging `learned.channelIds` to include a channel the prover never funded
is an information-disclosure / false-attestation bug — the verifier accepts a false chain-of-custody
claim. Not an operational/robustness issue.

## Step 4 — PoC — executed, not sketched

Reproducer adapted from `AuditAdversarial/FINDINGS.md`'s F1 reproducer to the real fixture plumbing
in `armada-sdk/src/disclosure/disclosure.test.ts` (same artifact loading, same `loadKeys()`,
`createSnarkjsProver`, `loadDisclosureArtifacts`). Saved temporarily as
`armada-sdk/src/disclosure/__repro_f1.test.ts`, run with the repo's own `vitest` against the real dev
circuit artifacts (`armada-circuits/build/disclosure/{main_disclosure.wasm,final.zkey,vkey.json}`),
then deleted. `git status --porcelain` in both `armada-sdk` and `armada-circuits` is clean after
removal.

Key construction (differs from the FINDINGS.md sketch only in matching the real fixture's import
paths and `DisclosureSigningKeys` shape):

- One opening/spend (`N1`, the prover's own note, position 5), proof id `n1`.
- `hops: [{ kind: 'H4', proof: 'n1', channelId: STRANGER_CHANNEL, expectedHeld: V1, asset: USDC,
  depositTxid: STRANGER_TX }]` — **no H2 hop at all**, so the disclosed nullifier is referenced only
  by this H4 hop.
- `source.nullifiers = []` — no `Nullified` event anywhere in the chain-source fixture, for this
  nullifier or any other. This isolates exactly the condition the finding describes (no co-located H2
  masking the gap).
- `source.deposits = [{ txid: STRANGER_TX, channelId: STRANGER_CHANNEL, asset: USDC,
  held: EXPECTED_HELD, adaptParams: keccak256(abi.encode(channelId, expectedHeld)) }]` — a
  self-consistent "stranger's" real `Deposited` event, built only from public fields.

Command and real output:

```
$ cd armada-sdk && npx vitest run src/disclosure/__repro_f1.test.ts
 RUN  v1.6.1 /home/rix/code/git_puller/repos/armada-sdk

stdout | src/disclosure/__repro_f1.test.ts > F1 repro: H4 accepts a deposit nobody disclosed the funding nullifier for
F1 repro result: {"ok":true,"learned":{"masterPublicKey":"16488779361287078159140615926408243965175068214317808385887533182539086787177","endpoints":[],"transactionIds":["0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"],"channelIds":["0xcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcdcd"]}}

 ✓ src/disclosure/__repro_f1.test.ts  (1 test) 841ms

 Test Files  1 passed (1)
      Tests  1 passed (1)
```

The test asserts `result.ok === true` and `result.learned.channelIds` contains `STRANGER_CHANNEL`
(`0xcdcd...cd`, the deposit transaction `0xeeee...ee`). Both assertions passed, against a package
built with a **real** Groth16 proof (the dev-circuit prover actually ran and the real `vkey.json`
actually verified it), not a stub.

## Step 5 — Checked for anything that closes the gap elsewhere

- `builder.ts` (`buildDisclosurePackage`, full file read): builds one proof per `spends` entry and
  places hops verbatim as given by the caller. No cross-hop validation that a proof id used by an H4
  hop is also used by an H2 hop for the same `depositTxid`. Nothing stops `hops` from containing a
  bare H4 with no paired H2.
- `types.ts` `Hop4`: a doc comment on `proof` says "the same nullifier an H2 proves" — this is
  documentation of *intended usage*, not an enforced structural constraint. The type system does not
  require a corresponding H2 hop to exist.
- `checkDisclosurePackage` has no post-loop pass that cross-checks proof-to-hop-kind pairing; it
  returns `ok: true` directly after the per-hop loop (checker.ts:323).
- Spec (`engineering/16-selective-disclosure.md:128-136`, "What the verifier checks"): step 3 ("The
  disclosure proof for n1, and that n1 is in the Nullified event of the deposit transaction") and
  step 4 (the `adaptParams`/`Deposited` check) are two distinct, independently numbered requirements.
  The code implements only step 4. Step 3 is not framed as optional or as defense-in-depth for step 4
  — they prove different facts (step 3: this nullifier was spent in this transaction; step 4: this
  transaction's deposit binds to this channel/amount). A forger who controls only step 4's public
  inputs cannot satisfy step 3 without actually spending into that deposit.
- `armada-sdk/docs/guide/selective-disclosure.md` "D5" describes only H2/H6 as performing the
  nullifier-event check; H4 is not named — consistent with the code (not an independent control that
  closes the gap).

Nothing in the circuit, the builder, the statement hash, or the EIP-712 signatures ties the disclosed
`hop.proof`'s nullifier to `hop.depositTxid`. The statement hash commits to the hops array as declared
(preventing post-signature tampering) but does not verify the hops' truth — that is the checker's job,
and the checker skips it for H4.

## Devil's advocate (7 questions)

1. Pattern-matching bias? No — confirmed by full-file code read (no call anywhere) and by an executed
   proof-carrying reproducer, not a visual pattern.
2. Trust boundary confusion? No — `hop.channelId`/`expectedHeld`/`asset`/`depositTxid` are all
   attacker-supplied hop fields, confirmed attacker-controlled in `types.ts`'s `Hop4`. `params.source`
   itself (the verifier's chain reader) is trusted and reflects a genuine, independently-produced
   on-chain event; the attack works precisely because the attacker needs no control over that event —
   only its public contents.
3. Proof rigor? Demonstrated by execution: `result.ok === true` with zero `Nullified` events present
   anywhere in the chain source.
4. Defense-in-depth confusion? No — step 3 and step 4 of the spec prove different facts; step 3 is not
   redundant with step 4.
5. LLM hallucination? No — the reproducer used the real dev circuit (`main_disclosure.wasm`,
   `final.zkey`) and the real `vkey.json`, producing an actually-valid Groth16 proof that actually
   verified; this is empirical evidence, not inference.
6. Dismissing as too complex? No — the exploit needs only a prover's own trivial note (which every
   legitimate user of H1-H6 already has) and reading a public `Deposited` event. No privileged
   information, no race, no chain reorg.
7. Invented mitigations? Checked `builder.ts`, `types.ts`, the full `checker.ts`, and the spec
   directly; found none. (See Step 5.)

No question produced unresolved uncertainty. No escalation to deep verification needed.

## Gate review

| Gate | Result |
|---|---|
| 1. Process | PASS — all phases executed with evidence, including a real proof-carrying execution. |
| 2. Reachability | PASS — crafted package reaches `checkHop('H4')` and every check present in the code; demonstrated live. |
| 3. Real Impact | PASS — forged `learned.channelIds` is a false attestation a downstream verifier (per the spec's own worked example) relies on for a compliance/provenance decision. |
| 4. PoC Validation | PASS — executed vitest test, real Groth16 proof, `result.ok === true`, `channelIds` contains the stranger's channel. |
| 5. Math Bounds | PASS (structural form) — code read confirms the H4 branch has no code path that calls `checkNullifierEvent`/`nullifiersIn`; the gap is unconditional, not bounded by any guard. |
| 6. Environment | PASS — no environmental protection blocks this; the Groth16 verifier works exactly as designed (it proves the prover knows the nullifier's secret), but nothing in the checker ties that proof to the specific deposit transaction being claimed. |

## Verdict

**F1 TRUE POSITIVE** — `checker.ts`'s H4 case accepts a channel-funding claim without checking that
the disclosed proof's nullifier is in the `Nullified` event of the named deposit transaction, letting
a prover claim credit for another wallet's real on-chain deposit using only their own unrelated note
and public chain data. Confirmed by live execution against the real circuit artifacts, not only by
static code reading.

**Severity (by consequence).** High. This breaks the core integrity property of the H4 hop: that the
disclosed spend actually funded the named channel deposit. The forgery needs no privileged
information — any wallet holder with one dust-value note can forge a funding claim for any public
`Deposited` event on the chain. A verifier using `checkDisclosurePackage` for a compliance or
chain-of-custody decision (the explicit use case in the spec's own worked example) accepts a false
attestation. This is a spoofing/information-integrity bug matching the audit's primary threat model
("a dishonest prover who wants a verifier to accept a chain of custody that is false").

**Minimal fix shape.** In the `H4` case of `checkHop` (`checker.ts:237-258`), add the same call the H2
and H6 cases already make: `const nullErr = checkNullifierEvent(proof, hop.depositTxid, i); if
(nullErr !== null) return nullErr;` before accepting the hop (after the proof verifies and before — or
after — the deposit-binding checks). This closes the gap with the function the file already has; no
new logic needed.

## Reproducer disposal

`armada-sdk/src/disclosure/__repro_f1.test.ts` was created to run this check and deleted after the
run. `git status --porcelain` is clean in both `armada-sdk` and `armada-circuits`.
