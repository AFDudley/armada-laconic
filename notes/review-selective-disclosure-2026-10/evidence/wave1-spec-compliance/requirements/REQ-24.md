## REQ-24 — Worked example, verifier check 2: npk1 = Poseidon(M, random1) against Shield event

> "2. `npk1 = Poseidon(M, random1)` against the `Shield` event of step 1."
> — engineering/16-selective-disclosure.md, "What the verifier checks"

**Verdict:** implemented · confidence: high

**Where enforcement lives:** this is hop `H1`'s check (REQ-17): `checkOpening` +
`checkShieldLocation` in `checker.ts`. The worked example's `hops` array (reconstructed from
`armada-sdk/src/disclosure/disclosure.test.ts:L116-L122`) places this as the first hop.

**How the verdict was reached:** identical mechanism to REQ-17; no separate code path exists for
"the worked example" versus the general H1 case, which is correct (the example is an instance of
the general hop semantics, not a special case).
