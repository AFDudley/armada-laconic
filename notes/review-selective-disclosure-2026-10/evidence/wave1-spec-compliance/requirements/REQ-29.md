## REQ-29 — Worked example, verifier check 7 (optional): co-signed swap states

> "7. Optional: the co-signed states. Alice's signature recovers to her channel participant key.
> The hub's signature passes `HubCustodian.isValidSignature` (EIP-1271) at a block where the
> hub's delegate was active."
> — engineering/16-selective-disclosure.md, "What the verifier checks"

**Verdict:** absent · confidence: high

**What this demands:** the package format and the checker SHOULD be able to carry and verify the
optional co-signed swap-state attribution evidence described in "The source of the yield."

**Searched (in `armada-sdk/src/disclosure/`):**
- `isValidSignature` → 0 hits.
- `EIP-1271` / `1271` → 0 hits.
- `co-signed` / `swap state` → 0 hits.
- `HubCustodian` → 0 hits.
- `DisclosurePackage` (`types.ts`) fields → `version, domain, statement, endpointSignatures,
  openings, proofs, hops` — no field of any kind to carry a co-signed state, a channel
  participant key, or a delegate-activity window.
- `DisclosureCheckName` (`checker.ts`) → 18 named checks, none referencing signatures other than
  `endpoint-signature` (EOA statement signatures) and `proof-verification` (the Groth16
  disclosure proof).

**How the verdict was reached:** the requirement is explicitly marked "Optional" by the spec
itself, and nothing in the current schema or checker can represent it — not a partial
implementation under a different name, a genuine absence. Per the divergence rubric this is Low
severity: it is an explicitly optional piece of *additional* attribution evidence ("The source of
the yield"), not a required proof of custody; its absence does not let a false custody claim pass
`checkDisclosurePackage`.
