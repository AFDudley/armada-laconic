## REQ-33 — Limit: disclosed data is permanent

> "**Disclosed data is permanent.** After disclosure, M links the package to the user's 0zk
> address, and `random1` and `random2` open N1 and N2 for anyone who holds the package."
> — engineering/16-selective-disclosure.md, "Limits"

**Verdict:** implemented · confidence: high

**What this demands:** `NoteOpening.random` must be carried in the package as recoverable
plaintext (not hashed/encrypted away), so anyone holding the package can open the note.

**Where enforcement lives:** `NoteOpening.random` (`types.ts`) is typed as `0x-prefixed hex`, and
`checkDisclosurePackage`'s `checkOpening` reads it directly as a cleartext witness to recompute
`npk` (`checker.ts`). `serializeDisclosurePackage` (`serialize.ts`) writes the whole package,
`random` included, with no redaction.

**How the verdict was reached:** the data is plainly present and unobscured in the serialized
form; nothing in the code attempts (or fails at attempting) to make this reversible/ephemeral,
matching the stated permanence.
