## REQ-32 — Limit: the package can be forwarded

> "**The package can be forwarded.** The verifier can pass the package to others, and each one
> can check it. This is necessary for an audit record. S names the verifier and the context, so
> the package states for whom and for what the user made it."
> — engineering/16-selective-disclosure.md, "Limits"

**Verdict:** implemented · confidence: high

**What this demands:** (a) the package format must be self-contained and independently
re-checkable by any holder; (b) S must name the verifier and the context (descriptive fields,
not enforced against the actual checking party).

**Where enforcement lives:** `serializeDisclosurePackage`/`parseDisclosurePackage`
(`serialize.ts`) round-trip the full `DisclosurePackage` as plain JSON — a third party can call
`checkDisclosurePackage` on a parsed copy with no additional secret material (confirmed: no
field in `DisclosurePackage` or `CheckDisclosureParams` requires anything beyond the package, a
`DisclosureChainSource`, a prover, and the public `vkey`). `DisclosureStatement.verifier` and
`.context` are plain strings carried in the signed statement (`types.ts`, `statement.ts`); per
REQ-32's own text these are descriptive only ("the package states for whom and for what"), and
correctly, `checkDisclosurePackage` never compares them against the calling party's identity.

**How the verdict was reached:** forwardability is a structural property (plain, self-contained
JSON; no per-recipient secret), confirmed by reading `serialize.ts` and the `CheckDisclosureParams`
shape; the "S names..." clause is confirmed by `DISCLOSURE_EIP712_TYPES.Statement` carrying
`verifier`/`context` fields that are part of the signed digest (REQ-09).
