# 10. Quality Requirements

arc42 §10 · 2026-10-20

This section refines the quality goals named in §1 into a **quality tree** that
decomposes each attribute into sub-attributes, then makes those sub-attributes
testable as concrete **stimulus → response** scenarios. Every scenario names the
`T#.#` item(s) that realize the guarantee and states whether that item is built,
in progress, or a plan. The quality tree is a view over the §5 registry; it invents no
new items.

## Quality tree

```
Armada × Laconic quality
├─ Security / Privacy         "leak nothing that collapses the anonymity set"
│  ├─ Custody               non-custodial always — the pool/channel hold value, no vault   (T0.0/T0.2, built)
│  ├─ Read-time anonymity   no observer learns which notes a user scans                     (T1, in progress)
│  ├─ Write-time origin     on-chain msg.sender is never the user's EOA                     (T6.5, not built)
│  └─ Amount privacy        amounts opaque in-pool; public only at the boundary             (T0.0/T0.7; opt T0.6, not built)
├─ Safety / Liveness         "a dead or lying counterparty costs time, never funds"
│  ├─ Force-closability      an honest party can always exit unilaterally                    (T0.2, built)
│  ├─ Watchtower correctness a stale-state force-close is always answered in-window          (T6.3, not built — test harness only)
│  └─ Feed freshness         staleness is detected, never silently missed                    (T1/T2.4, not built)
├─ Operability               "phone-first where possible, conventional server where not"
│  ├─ Mobile-first           shield/trade/settle/scan on a phone                             (T6.0/T6.1/T6.5, not built)
│  └─ Low fixed infra        armada-relayer is a conventional server today, not P2P          (T2.2, built — different mechanism)
├─ Performance               "clearing latency + on-device proving are usable"
│  ├─ RFQ latency            posted-price take-it-or-leave-it venue                          (T4.0/T4.1, not built)
│  └─ On-device proving      Groth16 proof within acceptable time/memory                      (T6.6, partly built — browser only)
└─ Auditability              "correctness is checkable, not trusted"
   ├─ Local verifiability    a client verifies reads against a beacon-anchored header         (T1, in progress)
   └─ On-chain enforceability a federation's attestation is a slashable fraud proof on L1      (T2.4/T0.4/T0.5, not built)
```

Read-time anonymity is the do-first property: if reads fingerprint the user, the
anonymity set collapses to one and every other protection is moot. Today no client
verifies a read (T1 is upstream and unintegrated), so this property is not yet met.
Watchtower correctness is the highest-priority *safety* surface and is also not met
today outside a test harness (T6.3).

## Quality scenarios

Each scenario states a stimulus, the required response, and the observable measure
that makes it testable. A scenario marked **plan** has no code to test yet; its
measure describes what a future build must satisfy.

**QS-1 · Read-time anonymity — proof verification (plan).**
*Stimulus:* an RPC provider or indexer serves a tampered value for a pool storage slot
or an incomplete receipt set.
*Response:* the client verifies the value against a beacon-anchored header through the
nimbus verified proxy — a storage proof (EIP-1186) for a slot, or a rebuilt receipt
trie root for a block's receipts — and rejects any mismatch, placing no trust in the
server.
*Measure:* after shielding or depositing a known value, a doctored response MUST fail
client-side verification while the honest one verifies. No client calls the proxy
today. → **T1** (not integrated).

**QS-2 · Read-time anonymity — no RPC fingerprint (not met today).**
*Stimulus:* an RPC or indexing provider tries to learn which commitments or nullifiers
a user scans.
*Response:* the design calls for every subscriber to a feed slice to receive
identical bytes and scan locally, so there is no per-user query to fingerprint.
*Measure:* `armada-interface` today issues ordinary RPC calls configured per
deployment; no identical-bytes feed or local-only scan guarantee exists. → **T1/T6.1**
(the scanner is built; the privacy guarantee depends on the unbuilt proxy integration).

**QS-3 · Watchtower correctness — stale-state force-close (plan for a client; built as
a test harness).**
*Stimulus:* a counterparty force-closes an old (superseded) state while the user is
offline.
*Response:* a watchtower would submit the user's higher-turn co-signed state through
`respond`/`checkpoint` within the challenge window, and the correct outcome finalizes.
*Measure:* `armada-nitro`'s e2e gate harness (`env/gates/watchtower/watch.ts`)
exercises exactly this and passes in CI. No deployed client watchtower exists, so no
user is protected automatically today. → **T6.3** (not built); dispute machinery
**T0.2** (built).

**QS-4 · Force-closability / non-custody — dead counterparty.**
*Stimulus:* a counterparty goes unresponsive mid-channel.
*Response:* the honest party unilaterally force-closes and exits through
`ArmadaNitroAdjudicator`'s ForceMove path; because no party ever holds custody, the
worst case is delay, not loss.
*Measure:* with the counterparty silent, the honest party drives the force-close to
finalize, and the SHIELD payout mints correct fresh notes. The contract-level guarantee
holds; a client must drive this manually today, since no code in `armada-interface` or
`armada-sdk` drives a channel lifecycle. → **T0.2** (built); client driver **T6.2**
(partly built).

**QS-5 · Write-time origin privacy (plan).**
*Stimulus:* a user submits an on-chain settlement/shield transaction from a wallet.
*Response:* a keeper/broadcaster would submit on behalf of the user, so `msg.sender` is
the relay rather than the user's EOA.
*Measure:* today `armada-interface` submits transactions directly from the user's own
key; no relayed/metered broadcast path exists for settlement. → **T6.5** (not built).

**QS-6 · Amount privacy at the boundary — large exit (process, not code).**
*Stimulus:* a large deposit or withdrawal would re-narrow the anonymity set when it
crosses the boundary inside `ArmadaNitroAdjudicator`.
*Response:* the anonymity set grows only with deposits into Armada's own pool
(ruling 5); no aggregation/dribble mechanism exists in the
contracts today. In-pool amounts are already SNARK-opaque; only the boundary amount is
public.
*Measure:* an exit's boundary amount is visible on-chain; no mitigation beyond pool
growth exists today. → **T0.7** (strategy, not code).

**QS-7 · On-device proving within budget.**
*Stimulus:* a shield or transact proof is generated on the user's device.
*Response:* `armada-sdk`'s `snarkjs` Web Worker prover produces a standard Groth16
proof in the browser, worker-first with a same-thread fallback, against
`armada-circuits`' SHA-256-manifest-checked build artifacts.
*Measure:* the browser prover completes without blocking the UI thread and the
verifier accepts its proof. No native mobile (rapidsnark/mopro) prover exists, so this
scenario is untested on a phone. → **T6.1/T6.6** (browser: built; mobile: not built).

**QS-8 · RFQ latency / nothing-to-front-run (plan).**
*Stimulus:* two takers hit a venue's posted quote near-simultaneously.
*Response:* a posted-price dealer market would have a fixed bid/ask with nothing to
front-run.
*Measure:* no venue, price-setter contract, or quote/settle app exists in `armada-poc`
or `armada-nitro`, so there is nothing to measure yet. → **T4.0/T4.1** (not built).

**QS-9 · Metering integrity (different mechanism than planned).**
*Stimulus:* a peer tries to under-pay or double-count a metered read or relay.
*Response:* `armada-relayer`'s actor charges a flat USDC fee computed from a gas
estimate per transaction (`fee-calculator.ts`), not a go-nitro payment-voucher stream.
There is no voucher-replay surface to test, because there are no vouchers.
*Measure:* a relayed transaction is charged the computed fee; a transaction without
payment is not relayed. → **T2.1** (built, a different mechanism from the plan).

**QS-10 · Feed-freshness liveness (plan).**
*Stimulus:* an indexer withholds or lags behind the chain head.
*Response:* the design calls for a published head cursor a consumer checks, falling
back to a direct read if the cursor lags; a later bonded federation would make
withholding slashable.
*Measure:* today `armada-relayer`'s Ponder indexer exposes no head-cursor freshness
check that a client consults, and no federation exists. → **T2.4/T0.4/T0.5** (not
built).

---

**Traceability.** These scenarios refine the §1 quality goals and exercise the §6
runtime flows; the concepts they defend are described once in §8 (privacy, settlement,
transport, metering, trusted setup). The gap between each plan scenario and today's
code is tracked in §11.
