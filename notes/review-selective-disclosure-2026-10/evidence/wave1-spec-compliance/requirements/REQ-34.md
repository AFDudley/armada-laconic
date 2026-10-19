## REQ-34 — D20.1: voluntary, off-chain chain-of-custody protocol, reveals nothing else

> "**D20.1** Armada provides a voluntary, off-chain selective disclosure protocol. With it, a
> user proves a chain of custody for chosen value: from a shielding address, through notes,
> spends and armada-nitro channels, to a payout note or a withdrawal address. The user reveals
> nothing else about the wallet."
> — engineering/09-architecture-decisions.md, ADR-0020

**Verdict:** implemented · confidence: high

**What this demands:** the protocol must be opt-in (no on-chain or pool-level trigger), must
support exactly the chain shape named (shield → note → spend → channel → payout/withdrawal), and
must not leak wallet data beyond the chosen chain.

**Where enforcement lives:** `buildDisclosurePackage`/`checkDisclosurePackage` are pure
library functions in `armada-sdk/src/disclosure/`, invoked only by application code that chooses
to call them — no pool, chain-sync, or wallet-core code in the SDK calls into `disclosure/`
(confirmed: `src/disclosure/` is self-contained; its only cross-module imports are `../core/index`
for hashing/commitment helpers and `../prover/index` for the Groth16 adapter, both one-directional).
The six `Hop` variants (`types.ts`) cover exactly shield (H1) → note (via H1/H3/H5) → spend (H2)
→ channel (H4/H5) → payout note (H5) / withdrawal (H6) — matching the named chain shape. "Reveals
nothing else": the only wallet data ever placed in a package is `masterPublicKey` (public by
design — see "Key facts"), the chosen `random`s (one per disclosed note, REQ-33), and the
disclosure proofs' public signals (`statementHash, M, N, T, L`, REQ-02-REQ-06) — `nullifyingKey`
and the spending private key never appear in any package field (REQ-35).

**How the verdict was reached:** confirmed the module's isolation (no inbound calls from
non-disclosure code) and the hop-type coverage against the named chain shape; cross-checked
against REQ-35 for the "reveals nothing else" clause specifically re: the two secret keys.
