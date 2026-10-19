## REQ-28 — Worked example, verifier check 6: disclosure proof for n2 + Unshield event to 0xB

> "6. The disclosure proof for `n2`, and the `Unshield` event to `0xB` in the same transaction."
> — engineering/16-selective-disclosure.md, "What the verifier checks"

**Verdict:** implemented · confidence: high

**Where enforcement lives:** hop `H6`'s check (REQ-22): `verifyProof` + `checkSpendLeaf` +
`checkNullifierEvent` + `unshieldsIn(...).some((u) => addressesEqual(u.to, endpoint.account))`, all
against `hop.unshieldTxid` — the single transaction id covers both the nullifier-event check and
the `Unshield` event check, satisfying "in the same transaction" directly (unlike `H4`/check 3-4,
REQ-20/REQ-25, where the nullifier-event check and the deposit check use two independently
supplied transaction ids).

**How the verdict was reached:** identical mechanism to REQ-22; the "same transaction" clause is
satisfied here specifically because `H6`, unlike `H4`, uses one shared transaction id for both
sub-checks.
