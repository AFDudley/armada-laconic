# 10. Quality Requirements

arc42 §10 · 2026-09-04

This section refines the quality goals named in §1 into a **quality tree** that decomposes each
attribute into sub-attributes, then makes those sub-attributes testable as concrete
**stimulus → response** scenarios. Every scenario names the `T#.#` item(s) that realize the
guarantee and, where a decision fixes its shape, the governing `ADR-####`. The quality tree is a
view over the §5 registry; it invents no new items.

## Quality tree

```
Armada × Laconic quality
├─ Security / Privacy         "leak nothing that collapses the anonymity set"
│  ├─ Custody               non-custodial always — parties serve/meter, never hold value   (T0.2/T0.3, ADR-0004)
│  ├─ Read-time anonymity   no observer learns which notes a user scans                     (T2.0/T6.1, ADR-0009)
│  ├─ Write-time origin     on-chain msg.sender is never the user's EOA                     (T6.5/T2.2, ADR-0008)
│  └─ Amount privacy        amounts opaque in-pool; public only at the transparent boundary (T0.0/T0.7; opt T0.6, ADR-0005)
├─ Safety / Liveness         "a dead or lying counterparty costs time, never funds"
│  ├─ Force-closability      an honest party can always exit unilaterally                    (T0.2/T0.3, ADR-0004)
│  ├─ Watchtower correctness a stale-state force-close is always answered in-window          (T6.3/T0.2)
│  └─ Feed freshness         staleness is detected (head cursor), never silently missed      (T2.0/T6.3; v2 T2.4)
├─ Operability               "phone-first, near-zero fixed infra"
│  ├─ Mobile-first           full shield/trade/settle/scan on-device, keys in the enclave    (T6.0/T6.1/T6.5)
│  └─ Low fixed infra        only state-diff emitters + STUN/TURN are fixed; rest is P2P     (T1/T2.2/T2.3)
├─ Performance               "clearing latency + on-device proving are usable"
│  ├─ RFQ latency            posted-price take-it-or-leave-it over a tens-of-ms hot loop      (T4.0/T4.1/T2.3, ADR-0007)
│  └─ On-device proving      Groth16 proof within acceptable time/memory                      (T6.6, ADR-0003)
└─ Auditability              "correctness is checkable, not trusted"
   ├─ Local verifiability    every feed response carries a storage proof the client checks    (T2.0)
   └─ On-chain enforceability v2 attestations are slashable fraud proofs on L1                (T2.4/T0.4/T0.5)
```

The priorities are clear. **Read-time anonymity** is the do-first, load-bearing property: if
reads fingerprint the user, the anonymity set collapses to one and every other protection is
moot (T2.0/T6.1 ★). **Watchtower correctness** is the highest-priority *safety* surface (T6.3).
Both dominate the test budget.

## Quality scenarios

Each scenario states a stimulus (its source and trigger), the required response, and the
observable measure that makes it testable. Measures run on fixturenet unless noted otherwise.

**QS-1 · Read-time anonymity — feed authenticity.**
*Stimulus:* a watcher party serves a **tampered `value`** for a pool storage slice.
*Response:* the client verifies the `{value, proof}` storage/Merkle proof against the L1 state
root **locally** and rejects any mismatch, placing no trust in the watcher.
*Measure:* after shielding or depositing a known value, a doctored `value` MUST fail client-side
proof verification while the honest slice verifies. → **T2.0** (see §5 T2.0, §8 privacy).

**QS-2 · Read-time anonymity — anti-fingerprint.**
*Stimulus:* a data or RPC provider tries to learn **which commitments/nullifiers** a user scans.
*Response:* every subscriber to a slice receives **byte-identical** feed responses and scans
locally, so there is no per-user query to fingerprint.
*Measure:* a headless server peer and an in-browser peer subscribed to the same slice receive
identical bytes, and a device scan issues **zero** `eth_getLogs` / `eth_getStorageAt` calls to a
public endpoint. → **T2.0 / T6.1** (mobile-privacy §4 ★, ADR-0009).

**QS-3 · Watchtower correctness — stale-state force-close.**
*Stimulus:* a counterparty **force-closes an old (superseded) state** while the user is offline.
*Response:* the phone-resident watchtower submits the user's **higher-turn co-signed state**
through T0.3 `respond`/`checkpoint` **within the challenge window**, and the correct outcome
finalizes.
*Measure:* an old state is force-closed, the offline user's watchtower responds, and funds settle
to the correct fresh notes rather than the stale allocation. → **T6.3 / T0.2 / T0.3** (ADR-0004).

**QS-4 · Force-closability / non-custody — dead counterparty.**
*Stimulus:* a counterparty (or watcher party) goes **unresponsive** mid-channel.
*Response:* the honest party **unilaterally force-closes** and exits; because no party ever held
custody, the worst case is delay, not loss.
*Measure:* with the counterparty silent, the honest party drives the force-close to finalize and
payout mints correct fresh notes, and no configuration lets a party move user value.
→ **T0.2 / T0.3** (ADR-0004).

**QS-5 · Write-time origin privacy.**
*Stimulus:* a user submits an on-chain settlement/shield tx from a phone.
*Response:* a **keeper/broadcaster submits on-behalf**, so `msg.sender` is the relay rather than
the user's EOA, and the relayed request is metered by the same Nitro voucher stream.
*Measure:* the finalized tx's origin is the broadcaster address, unlinkable to the user's key,
and the request nets a metering voucher. → **T6.5 / T2.2** (ADR-0008).

**QS-6 · Amount privacy at the boundary — large exit.**
*Stimulus:* a **large deposit** would re-narrow the anonymity set when it crosses the transparent
Design-A boundary on exit.
*Response:* value is **aggregated over a window or dribbled** in and out incrementally, balancing
two clocks — window size against LP build-up — so that no single boundary crossing stands out.
This is deliberately **not** Tornado-style denomination buckets, since in-pool amounts are
already SNARK-opaque.
*Measure:* a whale exit split by the aggregation/dribble policy shows no lone large boundary
crossing above the configured threshold, and effective-k stays above target. → **T0.7 / T4.5**
(ADR-0005, ADR-0006).

**QS-7 · On-device proving within budget.**
*Stimulus:* a Railgun-scale JoinSplit proof is generated **on a phone**.
*Response:* the native Groth16 prover (rapidsnark/mopro) produces a standard proof within
acceptable time and **stays under the ~3 GB app-memory wall**; low-memory devices fall back to
the coSNARK server-assist, which still emits a standard Groth16 proof and leaves the verifier
unchanged.
*Measure:* benchmarking the real Railgun circuit on target devices shows proof-gen completing
without OOM, and the assist path yields a verifier-accepted proof. → **T6.6** (ADR-0003).

**QS-8 · RFQ latency / nothing-to-front-run.**
*Stimulus:* two takers hit the venue's posted quote near-simultaneously.
*Response:* the v1 venue is a **posted-price dealer market**, so a fixed bid/ask has nothing to
front-run; the quote/settle ForceMove exchange runs over the tens-of-ms libp2p-noise hot loop and
reposts if the price is wrong.
*Measure:* both fills clear at the posted price with no ordering advantage, and the 1:1 exchange
completes over direct-noise rather than Waku store-and-forward. → **T4.0 / T4.1 / T2.3**
(ADR-0007).

**QS-9 · Metering integrity — voucher replay.**
*Stimulus:* a peer **replays or reorders** a paid-feed voucher to under-pay or double-count.
*Response:* voucher accounting rejects the replayed or out-of-order voucher on go-nitro's
nonce/redemption logic, and an unpaid or underpaid request is refused; because the design is
non-custodial, the worst case is denied service.
*Measure:* N metered requests net to exactly the expected co-signed allocation, and a replayed
voucher is rejected without changing the channel balance. → **T2.1** (see §5 T2.1).

**QS-10 · Feed-freshness liveness (withholding watcher).**
*Stimulus:* a watcher **withholds** or lags the feed so the storage proof, which only catches a
*lying* value, cannot help.
*Response:* the consumer detects staleness via the published **head cursor** and treats it as a
liveness alarm, falling back to a redundant party or direct submit; v2 adds a **bonded
federation** whose attestation makes a withholding party slashable.
*Measure:* an artificially lagged emitter trips the freshness gate before the challenge window
elapses and the watchtower fails over, never silently missing the window. → **T2.0 / T6.3**;
v2 **T2.4 / T0.4 / T0.5**.

---

**Traceability.** These scenarios refine the §1 quality goals and exercise the §6 runtime flows;
the concepts they defend are described once in §8 (privacy, settlement, transport, metering,
trusted setup). The open exposures behind the deferred and v2 rows — cold-start anonymity set,
go-nitro maturity, ingestion freshness, and mobile transport — are tracked in §11.
