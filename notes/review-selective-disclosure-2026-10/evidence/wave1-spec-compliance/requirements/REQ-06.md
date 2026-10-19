## REQ-06 — Circuit table: nullifyingKey private, no table-listed constraint of its own

> "`nullifyingKey` | private | —"
> — engineering/16-selective-disclosure.md, table "The disclosure circuit"

**Verdict:** implemented · confidence: high

**What this demands:** `nullifyingKey` is a private witness signal, used by the other
constraints (REQ-01, REQ-02) but not independently constrained.

**Where enforcement lives:**

```circom
// NullifierDisclosure.circom:L63
signal input nullifyingKey;
```

It never appears in the public-input list of `main_disclosure.circom` and carries no range check
of its own; it is consumed only inside `NullifierOwnership` (REQ-01, REQ-02). Never serialized
into a `DisclosurePackage` field (`armada-sdk/src/disclosure/types.ts`) or any proof's public
signals (`DISCLOSURE_PUBLIC_SIGNAL_NAMES`, `circuit.ts:L7-L13`) — matching D20.2's "the protocol
never asks the user to disclose either key."

**How the verdict was reached:** matches the table row exactly; cross-checked against D20.2
(REQ-35) for the "stays secret" half of the claim.
