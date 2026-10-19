## REQ-31 — Limit: a received note ends the chain

> "**A received note ends the chain.** If a note came from another user, the user can open it
> (H3) but cannot prove the sender's history. The sender must make their own disclosure."
> — engineering/16-selective-disclosure.md, "Limits"

**Verdict:** implemented · confidence: high

**What this demands:** hop `H3` (opening a received note) MUST NOT be able to assert anything
about who sent that note or that note's own prior history.

**Where enforcement lives:** `checker.ts`'s `H3` case (REQ-19) only compares a recomputed
commitment hash against a `Transact` event; it has no access to, and makes no claim about, the
spend that produced that output note (no nullifier, no spender identity). `checkOpening`
(shared helper) only requires `opening.masterPublicKey === pkg.statement.masterPublicKey` — i.e.
it proves the *current* holder of the note has `M`, never anything about the note's origin.

**Searched:** any reference to a sender/origin field on `Hop3` (`types.ts`) → none; `Hop3` carries
only `{ kind, note, transact }`.

**How the verdict was reached:** the limitation holds by the absence of any code path that
attempts the thing the limit says is impossible.
