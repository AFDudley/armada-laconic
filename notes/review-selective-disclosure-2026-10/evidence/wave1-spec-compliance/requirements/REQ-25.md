## REQ-25 — Worked example, verifier check 3: disclosure proof for n1 + n1 in Nullified event of the deposit transaction

> "3. The disclosure proof for `n1`, and that `n1` is in the `Nullified` event of **the deposit
> transaction**."
> — engineering/16-selective-disclosure.md, "What the verifier checks" (emphasis added)

**Verdict:** partial · confidence: high

**What this demands:** two things, about the *same* transaction: (a) the disclosure proof for
`n1` verifies, and (b) `n1` is in the `Nullified` event **of the deposit transaction** — i.e. the
nullifier-event check is explicitly anchored to the deposit transaction named in check 4
(REQ-26), not to an independently-supplied transaction id.

**Where enforcement lives:** clause (a) is `verifyProof` (REQ-18, shared with `H2`). Clause (b),
as written here, would require the nullifier-event check to run against the *deposit*
transaction. In the implementation, the only place `checkNullifierEvent` runs is inside the `H2`
case, against `hop.nullifiedTxid` — a field the package author supplies independently of `H4`'s
`hop.depositTxid` (REQ-20). Nothing in `checker.ts` requires `hop.nullifiedTxid === hop.depositTxid`
for a shared proof id, and the `H4` case itself never calls `checkNullifierEvent` at all.

**Paths walked:** see REQ-20 and REQ-03 for the full trace; this requirement is the same gap,
cited from the worked example's own prose rather than from the circuit table.

**Searched:** same as REQ-20.

**How the verdict was reached:** `partial`, for the same reason as REQ-20 and REQ-03 — this is
one textual statement of the same underlying requirement, restated three times across the page
(circuit table, H4 row, and here), and the code satisfies it in the `H2`/`H6` cases but not when
`H4` is used, including when `H4` is used without a paired `H2` naming the same transaction.

**Open questions:** none beyond those in REQ-20.
