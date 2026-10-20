# T3 · Ordering

**Status:** not built · tier doc · 2026-10-20
**Parent:** [`05-building-block-view.md`](./05-building-block-view.md) · **Tier:** T3

T3 is a fair-ordering stack for a future matcher. No code exists for any item in this tier. It becomes necessary only if a venue upgrades to a matcher where price discovery runs across multiple parties with competitive quotes, so that a book exists to front-run. No venue of any kind exists yet (T4), so T3 has no near-term dependency to serve.

No threshold-signing library has been chosen for this tier. `armada-nitro`'s `HubCustodian` uses FROST over secp256k1 for a different purpose (channel-participant signing, ADR-0016); whether a future T3 attestation layer would reuse FROST or choose something else is an open question, not a decision.

## T3.0 Commit-reveal sequencer + randomness beacon — not built

The design idea: orders stay hidden from the venue during ordering through commit-reveal, not client-side proving. A commit-reveal round trip costs tens of milliseconds against seconds for a per-order proof. One epoch: commit (`H(order ‖ salt)`), seal (with threshold-signed inclusion receipts so a censored commit becomes a slashable fraud proof), a post-seal randomness beacon fixes intra-epoch order, then reveal and clear as a batch auction at a uniform price. No code implements any part of this.

## T3.1 Epoch set-agreement / DA — not built

The design idea: a federation agrees which commitments are in an epoch and publishes the revealed epoch, a sequencing-and-attestation task, not full BFT consensus or shared global state. A member whose accepted commit is excluded from the attested set holds a fraud proof. No code implements any part of this.

## T3.2 Attestation — not built

The design idea: the federation signs sequencing certificates, inclusion receipts, and liquidity-proof snapshots with a threshold signature an L1 contract can verify, turning a censored commit into a slashable fraud proof. Which threshold scheme this tier would use is undecided; see the correction above.

## Dependencies this tier would need

A registry and bond contract (T0.4) and a fraud-proof verifier (T0.5), neither of which exist. A federation of watcher parties willing to bond and sign (T2.4), which also does not exist.

## Sources

- The design reasoning above traces to the pre-code architecture docs; it has not been re-verified against any code, because no code implements it.
- `HubCustodian` (`armada-nitro@55939ef:src/HubCustodian.sol`) is the one FROST implementation in the codebase, and it serves channel-participant signing, not this tier.
- Registry / bond and fraud-proof verifier referenced by id — T0.4, T0.5; federation — T2.4; matcher consumer — T4.6. See [`05-building-block-view.html`](./05-building-block-view.html).
