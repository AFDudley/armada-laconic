## REQ-35 — D20.2: disclosure proof links spend to note, keeps nullifyingKey and viewing key secret

> "**D20.2** A disclosure proof links a spend to the note it spent. The proof keeps the
> `nullifyingKey` and the viewing key secret. The protocol never asks the user to disclose
> either key."
> — engineering/09-architecture-decisions.md, ADR-0020

**Verdict:** implemented · confidence: high

**Where enforcement lives:** `nullifyingKey` is a private circuit witness only
(`NullifierDisclosure.circom:L63`, REQ-06); `main_disclosure.circom`'s public-input list
(`[statementHash, masterPublicKey, nullifier, treeNumber, leafIndex]`) excludes it. On the SDK
side, `DisclosureSigningKeys.nullifyingKey` (`circuit.ts`) is consumed only inside
`formatDisclosureInput` to build the witness; it is not copied into `DisclosureProofEntry` or any
other `DisclosurePackage` field (`types.ts`). The viewing key itself never appears anywhere in
`src/disclosure/` — only its already-one-way-derived `nullifyingKey` does, and that stays private
as above.

**Searched:** `nullifyingKey` in `types.ts` (the package's own field definitions) → 0 hits.
`viewingPrivateKey` / `viewingKey` anywhere in `src/disclosure/` → 0 hits.

**How the verdict was reached:** both named secrets (`nullifyingKey`, viewing key) are absent
from every serializable structure; the proof-links-spend-to-note clause is REQ-01/REQ-02 (the
circuit's own ownership constraints).
