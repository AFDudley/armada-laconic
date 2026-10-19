## REQ-26 — Worked example, verifier check 4: deposit adaptParams + Deposited event

> "4. The deposit transaction's `adaptParams` equals `depositBinding(C, expectedHeld)`, and
> `Deposited(C, USDC, …)` is in the same transaction."
> — engineering/16-selective-disclosure.md, "What the verifier checks"

**Verdict:** implemented · confidence: high

**Where enforcement lives — checker.ts, `H4` case (`checker.ts:L228-L238`):** the
`deposit-binding` and `deposited-event` checks, as detailed in REQ-20. Cross-checked the binding
formula against the contract: `checker.ts` computes
`keccak256(AbiCoder.defaultAbiCoder().encode(['bytes32', 'uint256'], [channelId, expectedHeld]))`;
`ArmadaNitroAdjudicator.sol:L63-L65` defines `depositBinding(channelId, expectedHeld) = return
keccak256(abi.encode(channelId, expectedHeld))`. Identical ABI types and argument order — exact
match.

**How the verdict was reached:** this specific clause (binding formula + same-transaction
`Deposited` event) is correctly implemented; it is the *other* clause of the same `H4` hop (the
nullifier-to-deposit-transaction link, REQ-20/REQ-25) that is missing. Reported separately
because this clause stands on its own and is correct.
