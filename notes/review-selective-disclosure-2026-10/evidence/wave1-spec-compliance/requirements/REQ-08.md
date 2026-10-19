## REQ-08 — Why the signed message includes every public input (incl. treeNumber)

> "In Groth16, a public input that no constraint uses is not bound by the proof. The message hash
> uses all five public inputs, so the proof binds each of them. It also stops a holder of the
> viewing key from reusing an old signature with a different nullifier. T is in the message
> because the same `nullifyingKey` and leaf index give the same nullifier in every tree."
> — engineering/16-selective-disclosure.md, "The disclosure circuit"

**Verdict:** implemented · confidence: high

**What this demands:** the EdDSA-signed message MUST be a function of all five public inputs
(`statementHash`, `masterPublicKey`, `nullifier`, `treeNumber`, `leafIndex`), including
`treeNumber` specifically (since `nullifier = Poseidon(nullifyingKey, leafIndex)` does not itself
depend on `treeNumber`, a cross-tree replay is possible unless `treeNumber` is also signed).

**Where enforcement lives:**

```circom
// NullifierDisclosure.circom:L86-L91
component msg = Poseidon(5);
msg.inputs[0] <== statementHash;
msg.inputs[1] <== masterPublicKey;
msg.inputs[2] <== nullifier;
msg.inputs[3] <== treeNumber;
msg.inputs[4] <== leafIndex;
```

All five public inputs of `main_disclosure.circom`'s public-input list
(`component main {public [statementHash, masterPublicKey, nullifier, treeNumber, leafIndex]}`)
appear in `msg.inputs`. `treeNumber` is `msg.inputs[3]`, confirming the specific claim about T.

Confirmed by `rejects: each public input changed after signing`
(`tests/disclosure.test.js:230-238`), which re-tests `treeNumber`, `statementHash`, and
`leafIndex` individually, each failing at `FAIL.signature`.

**How the verdict was reached:** direct 1:1 match between the five `Poseidon(5)` inputs and the
five public signals; the specific `treeNumber` sub-claim is independently confirmed by its own
test case.
