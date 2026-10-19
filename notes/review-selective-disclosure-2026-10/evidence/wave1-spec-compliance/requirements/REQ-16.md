## REQ-16 — Key binding between EOAs and the shielded wallet, both directions

> "So the statement binds each EOA to M in both directions: the EOA signs S (EIP-712), and S
> names M; the spending key of M signs S inside the disclosure circuit, and S names the EOA. A
> key holder on only one side cannot make the binding alone."
> — engineering/16-selective-disclosure.md, "Key binding between EOAs and the shielded wallet"
(restated as D20.3 — see REQ-36)

**Verdict:** implemented · confidence: high

**What this demands:** (a) each EOA's EIP-712 signature covers a statement that names M; (b) the
spending key of M signs (inside the disclosure proof) a message that names the EOA, via S.

**Where enforcement lives:**

(a) `DISCLOSURE_EIP712_TYPES.Statement` includes `masterPublicKey` directly
(`statement.ts:L25-L32`); `verifyEndpointSignature` (`statement.ts`) recovers the signer from
exactly this typed message and compares it with the claimed endpoint account
(`checker.ts:L95-L99`, the `endpoint-signature` check, run for every endpoint before any hop is
checked).

(b) `DISCLOSURE_EIP712_TYPES.Statement` also includes `endpoints` (the EOA addresses) directly
(`statement.ts:L25-L32`); `statementHashField` reduces the EIP-712 digest of that same statement
into S (`statement.ts`); the disclosure circuit's spending-key signature covers
`Poseidon(S, M, N, T, L)` (REQ-05/REQ-08). So S — which the spending key signs inside the proof —
is a hash that already names the EOA (via the `endpoints` field of the statement S derives from).

**Paths walked:** both directions resolve to the same `statement.ts` functions used by both the
builder (`builder.ts`, producing the endpoint signatures and the proofs over the same
`statementHash`) and the checker (`checker.ts`, verifying both independently before trusting any
hop).

**How the verdict was reached:** each half of the "both directions" claim traces to a distinct,
cited mechanism, and both mechanisms share the same underlying `DisclosureStatement` object, so
neither can be satisfied by a key holder on only one side.
