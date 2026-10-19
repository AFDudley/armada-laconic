## REQ-27 — Worked example, verifier check 5: finalized outcome pays npk2

> "5. The finalized outcome of C pays `npk2`, and `npk2 = Poseidon(M, random2)`."
> — engineering/16-selective-disclosure.md, "What the verifier checks"

**Verdict:** implemented · confidence: high

**Where enforcement lives:** hop `H5`'s check (REQ-21): `finalizedOutcomesIn` + `SHIELD`
allocation match + `checkOpening` for `npk2`, all in `checker.ts`'s `H5` case.

**How the verdict was reached:** identical mechanism to REQ-21.
