# Independent review of selective disclosure, 2026-10-16 to 2026-10-17

This directory records the review of the selective-disclosure work and its five verdicts (F1–F5). Session
`01a140be-05cd-7095-b3fb-32a573ee56ce` ran the review.

## What the review covered

| Target | Version reviewed |
|---|---|
| Design page 16 and ADR-0020 (armada-laconic) | The text before `f8a4739` (not committed at review time) |
| Disclosure circuit, ship-armada/armada-circuits#3 | Branch `disclosure-circuit`, head `7175b46` |
| SDK builder and checker, ship-armada/armada-sdk#142 | Branch `selective-disclosure`, head `7094351` |

## Method

`METHOD.md` gives the method, the threat model and the rules. It follows Trail of Bits' published practice: the
`audit-context-building`, `spec-to-code-compliance` and `fp-check` skills, and two ZKDocs pages (Fiat-Shamir, random
sampling). The skills' workflows could not run here, so the agents applied their rules by hand.

- **Wave 1:** three agents that did not write the code. One built context, one checked page 16 against the code
  requirement by requirement, and one hunted for breaks under the threat model.
- **Wave 2:** for each candidate finding, an agent that did not produce it tried to disprove it, and ran a reproducer.

## Verdicts

| ID | Finding | Verdict | Evidence | Resolution |
|---|---|---|---|---|
| F1 | The checker's H4 hop did not check that the funding nullifier is in the deposit transaction's `Nullified` event. A prover could claim any channel deposit on chain as its own. | True, High | A real proof over an unrelated note, paired with a stranger's public `Deposited` event and no `Nullified` events, returned `ok: true`. | Fixed in armada-sdk `8307604`. The regression test fails before the fix and passes after it. |
| F2 | The Picus uniqueness result for the circuit's wiring proves less than page 16 claimed, because Picus holds every input fixed, private ones included. | False | With the real public inputs declared, the compiled constraint file changed only in 2 header bytes, and Picus gave the same result. Picus defines uniqueness this way for every circuit. | Page 16 states how Picus treats inputs. |
| F3 | The checker did not tie events to the chain and pool named in the statement. | True, Medium | Two honest packages, one for chain 1 and one for chain 999, both passed against the same event record. | Fixed in armada-sdk `a863d2c`. Page 16 "Chain scope" and ADR-0020 D20.6. |
| F4 | The tests check the code only against the project's own fixtures. | False | 9 of 10 SDK vector files come from the stock Railgun engine 9.6.0. The parts new to disclosure cannot have external vectors, because no other party implements the scheme. | None. |
| F5 | The H2 and H6 hops do not check that the spent note's opening belongs to the statement's wallet. | False | The gap in the code is real but has no effect. The bad opening never reaches the checker's result, and the proof's public `masterPublicKey` binds the spend to the wallet. | None. |

The review also found two documentation errors in page 16, now corrected: a wrong count of EdDSA rejection tests, and
verifier check 7 marked "optional". Check 7 became hop H7, attribution as a required claim (ADR-0020 D20.5).

## Evidence

| Directory | Content |
|---|---|
| `evidence/wave1-context/` | Context dossier, one file per circuit template and SDK function |
| `evidence/wave1-spec-compliance/` | 37 numbered requirements from page 16 and ADR-0020, each with a verdict |
| `evidence/wave1-adversarial/` | The candidate findings F1–F4 |
| `evidence/wave2-F1/`, `wave2-F2/`, `wave2-F3-F5/`, `wave2-F4/` | One `VERDICT.md` per refutation, with the commands run and their output |

The evidence files are the agents' output, unchanged. They name test keys from the armada-sdk public test vectors
only.
