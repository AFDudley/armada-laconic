# fp-check verdict — F3 and F5 (armada-sdk `disclosure/checker.ts`, `chain-source.ts`)

Reviewer: `RefuteF3F5Checker`. Did not produce either candidate. Standard verification path
(single component, clear claim, no concurrency) for both.

Files read: `engineering/16-selective-disclosure.md`, `armada-sdk/docs/guide/selective-disclosure.md`,
`armada-sdk/src/disclosure/{checker,builder,types,statement,chain-source,disclosure.test}.ts`.
Two reproducer tests ran against the real dev circuit artifacts
(`armada-circuits/build/disclosure`, real Groth16 proofs, not mocked), then removed. `git status
--porcelain` is clean in both `armada-sdk` and `armada-laconic` after cleanup.

---

## F3 — No cross-check that hop events belong to the statement's declared chain/pool

**Claim restated.** `checkDisclosurePackage` reads `pkg.domain` only to scope the two EIP-712
EOA signatures and the `statementHash` field. It never compares `pkg.domain.chainId` /
`verifyingContract` against the chain events `params.source` returns for H1–H6. A dishonest
prover can declare one pool deployment in `pkg.domain` and satisfy the hops with real events from
a different deployment, if the verifier's `DisclosureChainSource` serves more than one deployment.

### Step 0 — bug class and threat model
Logic bug: a missing cross-check between two data sources (the self-declared statement domain,
and the externally-supplied chain-event source) that the protocol intends to describe one single
deployment. Threat model: "An attacker who manipulates the chain-event source the checker reads"
and "events from another pool or chain," both named in `local://audit-method.md`.

### Data Flow
```
Source: pkg.domain (prover-chosen, part of the package) — Trust: self-asserted by the discloser
Source: params.source (verifier-chosen DisclosureChainSource) — Trust: verifier-supplied dependency
Path: pkg.domain -> statementHashField() -> entry.statementHash check (H2/H4/H6) -> PASS/FAIL
Path: hop.*Txid -> params.source.xxxIn(txid) -> event match -> PASS/FAIL
These two paths never intersect. grep confirms `pkg.domain` appears exactly twice in checker.ts:
  verifyEndpointSignature(pkg.domain, ...)            [checker.ts:101]
  statementHashField(pkg.domain, pkg.statement)       [checker.ts:94]
Neither call reads anything from params.source. No decoded event type (DecodedShieldCommitment,
DecodedTransactCommitment, DecodedNullifier, DecodedUnshield, HubDepositRecord,
HubFinalizedOutcomeRecord) carries a chainId or contract-address field.
```

### Attacker Control
The attacker is the disclosure's own prover. They freely choose `pkg.domain` at build time, and
every proof/signature they build is self-consistently bound to that chosen domain (the circuit and
EIP-712 digest have no independent ground truth to check it against). The only external fact not
under their control is: does the verifier's `source` resolve `hop.*Txid` to a real, matching event.

### PoC (executed)
Built two full disclosure packages (`buildDisclosurePackage`, real Groth16 proof each) for the
same wallet and the same spent note, under two different, mutually exclusive domains:

```
domainReputablePool = { chainId: 1,   verifyingContract: 0xaaaa...aa }
domainSketchyPool   = { chainId: 999, verifyingContract: 0xbbbb...bb }
```

Both packages were checked against one `InMemoryDisclosureChainSource`, built from a single
`Nullified` event record with no chain tag (as the type requires). Result:

```
domain=reputable-pool (chainId 1) result: {"ok":true, "learned": {...}}
domain=sketchy-pool   (chainId 999) result: {"ok":true, "learned": {...}}
```

Both pass. The identical chain-agnostic event record satisfies the hop for two packages that
claim incompatible pool deployments. Nothing in `checkDisclosurePackage` distinguished them.

### Spec and docs check
`engineering/16-selective-disclosure.md`, section "Key binding between EOAs and the shielded
wallet" and the guide's D1, describe `pkg.domain` purely as replay protection for the EIP-712
signature and the proof's `statementHash`: "A signature made for one pool does not verify against
another." The spec never states that a verifier must also scope its chain-event source to
`pkg.domain`. `chain-source.ts`'s own docstring ("A verifier adapts its own chain access... to
these per-transaction lookups") does not say this either. So this is not a spec contradiction —
the spec is silent on the chain-source implementer's scoping duty. Under the compliance rubric in
`local://audit-method.md`, this is `absent`, not `contradicted`: the requirement that `source` be
single-deployment-scoped exists nowhere in writing, in the spec or in the SDK's own API docs.

### Checker defect vs. documented responsibility
Not a documented responsibility today — it is an unstated assumption. `CheckDisclosureParams.source`
is a caller-supplied dependency, so the checker cannot itself police what that dependency returns.
But two things push this past "caller's problem, not ours": (1) the decoded event/record types the
SDK itself defines (`event-decoder.ts`, `chain-source.ts`) carry no field a checker could even
compare against `pkg.domain`, so correct-by-construction scoping is impossible without a type
change; (2) the spec's own worked example names the disclosure recipient "Acme Exchange
compliance" — a third-party compliance/audit verifier is exactly the deployment shape (one service,
many client wallets, plausibly many deployments) most likely to run one shared indexer.

### Devil's Advocate
1. Pattern-matching bias? No — the PoC demonstrates the gap directly, not a "looks dangerous"
   pattern.
2. Trust-boundary confusion? The attacker does not control `params.source`; they control `pkg.domain`
   and the hop txids, and the PoC shows their honest, self-consistent choice of domain is enough —
   no tampering with the source itself was needed.
3. Proof rigor: the grep-confirmed absence of any domain read beyond the two cited call sites, plus
   the executed PoC, are both direct evidence, not inference.
4. Defense-in-depth confusion? A correctly-configured, single-deployment verifier (exactly what the
   spec's domain design assumes) is not exposed. This is a primary-control gap for the
   multi-deployment verifier case the spec never rules out, not a failed secondary control.
5. Real or hallucinated? Confirmed by execution against real circuit artifacts, not speculation.
6. Dismissing because the exploit needs a specific verifier shape? No — a shared/multi-deployment
   compliance verifier is the primary use case the spec's own example describes, so the precondition
   is realistic, not contrived.
7. Invented mitigations? None found on re-read: `statementHash` binds the statement to the domain,
   but never binds any specific chain EVENT to the domain, because the events the hops cite are
   looked up with no domain parameter in the interface at all.

### Gate Review
| Gate | Verdict |
|---|---|
| 1. Process | PASS — all phases documented with evidence |
| 2. Reachability | PASS — attacker (self-disclosing prover) fully controls `pkg.domain` and hop txids; PoC confirms the checker never rejects a mismatch |
| 3. Real Impact | PASS, conditional — a false chain-of-custody claim is accepted for a verifier whose source spans more than one deployment; this is the threat model's "events from another pool or chain" |
| 4. PoC Validation | PASS — executed, with real proofs, against real checker code |
| 5. Math Bounds | N/A (not a bounds bug); structural absence of any domain comparison is directly shown by code reading and by the PoC |
| 6. Environment | PARTIAL — a single-deployment verifier (the spec's implicit default) is not exposed; a shared multi-deployment verifier is fully exposed, and nothing in the SDK stops or warns against that shape |

**BUG F3 TRUE POSITIVE — the checker never ties the chain events it accepts for H1–H6 to the
pool/chain named in `pkg.domain`, and the SDK's own event and chain-source types carry no field
that would let it. This is exploitable, not merely theoretical, for any verifier that serves more
than one pool deployment from one chain-event source — the exact shape a third-party compliance
verifier (the spec's own worked example) is likely to run.**

**Severity.** Medium. Consequence: a verifier accepts a disclosure package's chain-of-custody claim
(amounts, channel, counterparties) for the wrong pool deployment, believing it describes the
deployment named in `pkg.domain`. This does not grant fund theft or let an attacker spend someone
else's note (ownership is still bound by the Groth16 proof's own `masterPublicKey` signal); the
harm is a false compliance/audit claim accepted as true for the named deployment. Severity is
capped at Medium because it requires a non-default verifier architecture (a shared, unscoped,
multi-deployment chain-event source) that the SDK does not ship and does not demonstrate — wave 1's
own reproducer note ("exploitability is a property of a caller's `DisclosureChainSource`
implementation, which does not exist in this repo") still holds; this review's PoC closes that gap
by constructing the minimal case, but the vulnerable source implementation remains hypothetical.

**Minimal fix shape.** Add a chain/contract field to the decoded event and hub-record types that
`chain-source.ts` consumes (`DecodedShieldCommitment`, `DecodedTransactCommitment`,
`DecodedNullifier`, `DecodedUnshield`, `HubDepositRecord`, `HubFinalizedOutcomeRecord`), populated
by the caller's decoder/sync layer from the log's actual chain id and emitting contract. In
`checker.ts`, compare that field against `pkg.domain.chainId` / `pkg.domain.verifyingContract` at
every `source.xxxIn(txid)` lookup, and fail with a new named check (for example
`event-domain-mismatch`) on a mismatch. Until that type change lands, document the missing
invariant explicitly in `chain-source.ts`'s docstring and in `docs/guide/selective-disclosure.md`:
one `DisclosureChainSource` instance must be scoped to exactly one `(chainId, verifyingContract)`
pair, and a verifier serving several deployments must construct and use a separate, scoped source
per check.

---

## F5 — H2/H6 open the spent note with `requireOpening` but never call `checkOpening`

**Claim restated.** H2 and H6 call `requireOpening(hop.spends, i)` to resolve the spent note's
opening, then only run `checkSpendLeaf` (tree/position) and `checkNullifierEvent` on it. They never
call `checkOpening`, which is what checks `opening.masterPublicKey === statement.masterPublicKey`
and `npk == Poseidon(masterPublicKey, random)`. H3 and H5 do call `checkOpening` on the notes they
open. A package with a wrong `masterPublicKey` or a forged `npk` on the spent-note opening in an H2
or H6 hop is not rejected for that reason.

### Step 0 — bug class and threat model
Logic bug: an asymmetric missing check between sibling code paths (H3/H5 vs. H2/H6) in the same
function. Threat model: "A dishonest prover who wants a verifier to accept a chain of custody that
is false: claims another wallet's note."

### Data Flow
```
Source: pkg.openings[label] (prover-chosen, part of the package) — Trust: self-asserted
Path (H2/H6): requireOpening(hop.spends,i) -> checkSpendLeaf(proof, opening, i) [tree/position
  only] -> checkNullifierEvent(proof, txid, i) -> hop accepted
Path (H1/H3/H5): requireOpening(..) -> checkOpening(opening,i) [masterPublicKey + npk check] ->
  checkShieldLocation/transact match against the REAL on-chain commitment hash -> hop accepted
```
`checkOpening` is skipped for H2/H6; confirmed by reading every line of the `H2`/`H6` case blocks
in `checker.ts` (lines 203–216, 280–299) against the `H1`/`H3`/`H5` case blocks (187–202, 217–236,
259–279), which all call `checkOpening` before anything else.

### PoC (executed)
Built a minimal, standalone H2-only package (no H1/H3/H5 creation hop for the spent note — the
shape the SDK's own `disclosure.test.ts` "H2 spend-leaf linkage" test already uses and accepts),
with a real Groth16 proof. Checked once unmodified, then again with the spent note's opening
corrupted to an unrelated `masterPublicKey` ("999") and an `npk` that is not `Poseidon(M, random)`
for any key ("0xdeadbeef"):

```
UNMODIFIED minimal H2-only package: {"ok":true, "learned": {"masterPublicKey":"1648877...787177", ...}}
CORRUPTED opening (wrong masterPublicKey + npk): {"ok":true, "learned": {"masterPublicKey":"1648877...787177", ...}}
```

Confirms the claim exactly: `ok: true` both times. H1/H3/H5's equivalent corruption is rejected —
`disclosure.test.ts` already has "rejects an opening for a different masterPublicKey" and "rejects
an opening whose npk is not Poseidon(M, random)" tests for an H1-opened note, both of which fail
with `opening-master-public-key` / `opening-npk`.

### What the attacker actually gains — traced to the end
`learned.masterPublicKey` in both runs above is the real wallet key
(`16488779...86787177`), taken from `pkg.statement.masterPublicKey`, not from the corrupted
opening. `DisclosureLearned` (`checker.ts:48-53`) exposes exactly `masterPublicKey`, `endpoints`,
`transactionIds`, `channelIds` — none of these are derived from the H2/H6 "spends" opening's
`npk`/`masterPublicKey`/`value`/`token`. Grepped every consumer of `pkg.openings` in the SDK
(`builder.ts`, `checker.ts`, test files only): nothing outside `checker.ts` reads it, and inside
`checker.ts` the H2/H6 branch reads only `opening.tree`/`opening.position` for the tautological
`checkSpendLeaf` comparison (both values are already public signals of the proof itself — the
opening adds no fact the proof did not already state).

The spend's actual identity binding to the wallet does not come from the opening at all. It comes
from `verifyProof`, run unconditionally for every H2/H4/H6 hop (`checker.ts:117-122`):
`entry.masterPublicKey !== pkg.statement.masterPublicKey` fails the check, and the Groth16
verification itself enforces `M == Poseidon(Ax, Ay, nullifyingKey)` and the EdDSA signature over
`Poseidon(S, M, N, T, L)` inside the circuit (`engineering/16-selective-disclosure.md`, "The
disclosure circuit"). A prover cannot produce a passing proof claiming a wallet `M` they do not
hold the spending key for, regardless of what they write in an unrelated opening object. So the
missing `checkOpening` call cannot be used to misattribute a spend to a different wallet than the
one the Groth16 proof — independently and unconditionally — already proves.

### Spec and docs check
The spec's hop table lists H2's evidence as only "A disclosure proof that a published nullifier
belongs to the note at a given leaf, for the disclosed masterPublicKey." No opening check is named
for H2, unlike H1 ("the verifier checks npk = Poseidon(masterPublicKey, random)"), H3 ("the
verifier recomputes the hash and compares it"), and H5 ("opens the npk as in H1"). The worked
Aave-hub example's seven-step "What the verifier checks" list has no opening check at step 3 (H2)
or step 6 (H6) either — only "The disclosure proof for n1, and that n1 is in the Nullified event."

The SDK guide is explicit and goes further, in D5: *"H2 and H6 confirm that a proof's (treeNumber,
leafIndex) is the spent note's leaf, and that the nullifier is in the Nullified event. This
separates the proof of ownership (the opening) from the proof of the spend (the disclosure
proof)."* This states, by name, that ownership for a spent note is proven once, by the opening on
its creation hop (H1/H3/H5) — not re-proven on every later hop that references the same note. In
the honest, well-formed chain (every example in the spec and the default test fixture), the H2/H6
"spends" label is always the SAME label as an earlier H1/H3/H5 "note"/creation hop in the same
package, and `openingByLabel` (`checker.ts:106`) is a single `Map` built once from `pkg.openings` —
so the SAME JS object is reused, and `checkOpening` already ran on it via the creation hop.

**Resolving the spec-compliance disagreement.** `implemented` is the correct rating for H2/H6. The
spec explicitly scopes H2/H6's verifier-side evidence to the proof and the nullifier event, not to
an opening check, and the SDK guide's D5 independently documents the same design choice with its
own rationale. The asymmetry is intentional, not an oversight: the opening check is unnecessary for
H2/H6 because (a) the spend's wallet binding is already fully proven in-circuit by the disclosure
proof, and (b) a note's existence/value binding to the chain is proven once, at its creation hop,
and reused by reference for every later hop on the same label.

### Devil's Advocate
1. Pattern-matching bias? The missing call is real and confirmed by direct code reading and
   execution — not a pattern-match. The uncertainty was always about consequence, not existence.
2. Trust-boundary confusion? No internal/external confusion; the opening is prover-supplied in both
   the checked and unchecked branches.
3. Proof rigor: traced `learned`'s fields to their actual source for both runs; the PoC output shows
   the corrupted fields never appear in the trusted output.
4. Defense-in-depth confusion? Close to the mark here: `checkOpening` on H2/H6, if added, would be a
   second, redundant proof of M-ownership layered on top of the circuit's own in-proof M-binding —
   exactly a defense-in-depth addition, not a primary control. Its absence does not remove the
   primary control (the Groth16 proof's masterPublicKey signal), which still fully holds.
5. Hallucinated? No — reproduced with real proofs, real checker code.
6. Dismissing a real vuln because the exploit seems complex? Checked the opposite direction too: is
   there ANY downstream consumer that trusts `pkg.openings` directly, bypassing `learned`? Grepped
   the whole SDK; none exists. The guide's own "Check a package" example only reads `result.failure`
   and (implicitly) `result.learned`.
7. Invented mitigations? Re-read `checker.ts` H2/H6 blocks line by line after reaching this
   conclusion; confirmed no hidden `checkOpening` call and no other comparison of `opening.npk`/
   `opening.masterPublicKey` anywhere in the file for these two hop kinds.

### Gate Review
| Gate | Verdict |
|---|---|
| 1. Process | PASS — all phases documented with evidence |
| 2. Reachability | PASS (the specific omission is reachable and demonstrated) |
| 3. Real Impact | FAIL — no security-relevant claim depends on the unchecked fields; the spend's wallet binding is independently, unconditionally proven by the Groth16 proof's own `masterPublicKey` public signal, and the corrupted fields never reach `learned` or any other consumer |
| 4. PoC Validation | PASS as a demonstration of the code gap, but the PoC shows no exploitable consequence |
| 5. Math Bounds | N/A |
| 6. Environment | PASS as a block — the design (proof-based ownership for the spend, opening-based ownership for the note's creation, linked only by label/leaf position) entirely prevents the corrupted fields from mattering, by construction, not by luck |

**BUG F5 FALSE POSITIVE — `checkOpening`'s absence in H2/H6 is confirmed by code reading and by an
executed PoC (a corrupted `masterPublicKey`/`npk` on an H2 spent-note opening is accepted), but the
corrupted fields are never used for anything: not for the spend's wallet-identity binding (proven
independently and unconditionally by the Groth16 proof's own `masterPublicKey` public signal, which
the circuit ties to the prover's own spending and nullifying keys), not for the checker's
`learned` output (which takes `masterPublicKey` from the statement, never from an opening), and not
by any other consumer of `pkg.openings` in the SDK (grepped; none exists). The spec's hop table and
the SDK guide's D5 both independently document this split (proof proves the spend's ownership,
opening proves the note's existence/value, checked once at creation) as the intended design, so
`implemented` is the correct spec-compliance rating for H2/H6, resolving the wave-1 disagreement.**
