## REQ-30 — Limit: the payout is a net amount

> "**The payout is a net amount.** If one channel carries many games, the payout is the final net
> allocation. The co-signed states are necessary to attribute the parts of the net amount."
> — engineering/16-selective-disclosure.md, "Limits"

**Verdict:** undecidable · confidence: high

**What this demands:** this is a descriptive caveat about *interpretation* of a package's
evidence (the payout figure is net, not gross), not an instruction to the implementation — it
names no check, field, or behavior the code must have. The code's `H5` check (REQ-21/REQ-27)
reads whatever `amount` a `SHIELD` allocation carries and does not claim it is anything other
than that allocation's value, so there is nothing for the code to get wrong here.

**How the verdict was reached:** no checkable implementation requirement is stated; flagged
`undecidable` rather than `implemented`, since "implemented" would imply a demand was met rather
than that no demand exists. Not a finding.
