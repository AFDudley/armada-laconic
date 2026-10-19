## REQ-17 — Hop H1: Address shields note

> "H1 · Address shields note | EOA → note | The `Shield` event shows `npk`. The user discloses
> `masterPublicKey` and `random`, and the verifier checks `npk = Poseidon(masterPublicKey,
> random)`. The EOA signs the statement."
> — engineering/16-selective-disclosure.md, table "The hops"

**Verdict:** implemented · confidence: high (stronger-than-spec in one sub-check)

**Where enforcement lives — armada-sdk/src/disclosure/checker.ts, case `'H1'` (`L164-L178`):**

```ts
const openErr = checkOpening(opening, i);          // npk == Poseidon(M, random)
const shieldErr = checkShieldLocation(opening, hop.shield, i);  // Shield event at (tree, position)
const sender = params.source.senderOf(hop.shield.txid);
if (sender !== undefined && !addressesEqual(sender, endpoint.account)) {
  return failure('shield-sender', ...);
}
```
`checkOpening` (`checker.ts`) computes `ShieldNote.getNotePublicKey(M, random)` and compares it
with the opening's `npk` — matches the requirement's formula exactly. `checkShieldLocation`
compares the recomputed commitment hash with the `Shield` record at the stated `(tree,
position)` in `hop.shield.txid`. The EOA's statement signature is checked package-wide
(`checker.ts:L92-L99`, REQ-16) before any hop runs.

**Stronger than spec:** the `shield-sender` check (comparing the `Shield` transaction's sender
to the endpoint account) is not named anywhere in the H1 row or in the "Key facts"/"Key binding"
sections. It is conditionally applied (`if sender !== undefined`), so it silently does nothing
when the caller's `DisclosureChainSource` does not supply sender data (`InMemoryDisclosureChainSource`
defaults `senders` to `{}` when omitted, `chain-source.ts`). Not a spec divergence (the spec names
no requirement here to diverge from); noted for completeness.

**How the verdict was reached:** every clause of the table row maps 1:1 to a cited check; the one
extra check present in code is optional and additive, not a contradiction.
