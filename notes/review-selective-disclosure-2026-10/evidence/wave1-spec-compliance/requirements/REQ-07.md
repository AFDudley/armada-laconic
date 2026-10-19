## REQ-07 — Why the circuit checks the spending key

> "A holder of the viewing key alone, such as an auditor or a compromised device, could prove N
> without the signature check. With the signature check, only the holder of the spending key can
> make a disclosure."
> — engineering/16-selective-disclosure.md, "The disclosure circuit"

**Verdict:** implemented · confidence: high

**What this demands:** knowledge of `nullifyingKey` alone (derivable from the viewing key) MUST
be insufficient to produce a valid witness; the EdDSA signature by the spending key
`(Ax, Ay)` MUST be required.

**Where enforcement lives:** the EdDSA check at `NullifierDisclosure.circom:L99-L105` (see
REQ-05) is unconditional (`eddsa.enabled <== 1`) and binds `(Ax, Ay)` into `masterPublicKey`
(REQ-01), so a prover who only knows `nullifyingKey` cannot supply a signature that verifies
without also controlling the private key behind `(Ax, Ay)`.

Confirmed by `rejects: a signature by another spending key` (`tests/disclosure.test.js:219-222`):
the same `nullifyingKey`/`masterPublicKey`/`nullifier` with a *different* spending key's signature
is rejected at the `EdDSAPoseidonVerifier` (`FAIL.signature`).

**How the verdict was reached:** the claim is a direct restatement of the EdDSA constraint's
effect, confirmed by the test designed to isolate exactly this property.
