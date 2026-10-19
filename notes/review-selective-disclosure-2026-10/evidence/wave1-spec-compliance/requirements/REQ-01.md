## REQ-01 — Circuit table: masterPublicKey constraint

> "`masterPublicKey` M | public | `M == Poseidon(Ax, Ay, nullifyingKey)`"
> — engineering/16-selective-disclosure.md, table "The disclosure circuit"

**Verdict:** implemented · confidence: high

**What this demands of an implementation:** the circuit MUST constrain the public signal
`masterPublicKey` to equal `Poseidon(Ax, Ay, nullifyingKey)` for the private witness `Ax, Ay,
nullifyingKey`, on every path (no bypass).

**Where enforcement lives:**

```circom
// NullifierDisclosure.circom:L75-L81
component own = NullifierOwnership();
own.spendingPublicKey[0] <== spendingPublicKey[0];
own.spendingPublicKey[1] <== spendingPublicKey[1];
own.nullifyingKey <== nullifyingKey;
own.leafIndex <== leafIndex;
masterPublicKey === own.masterPublicKey;
```
```circom
// NullifierDisclosure.circom:L42-L47 (NullifierOwnership)
component mpk = Poseidon(3);
mpk.inputs[0] <== spendingPublicKey[0];
mpk.inputs[1] <== spendingPublicKey[1];
mpk.inputs[2] <== nullifyingKey;
masterPublicKey <== mpk.out;
```

This is an unconditional `===` equality constraint (not a tagged/switchable branch), so it holds
on every witness. Confirmed functionally by the rejection test `rejects: a masterPublicKey of
another wallet` (`tests/disclosure.test.js:213-217`), which fails at `NullifierDisclosure_N line:
84` (the `===` above).

**Paths walked:** single path; `NullifierDisclosure` has no branches around this constraint.

**How the verdict was reached:** the constraint is literally the formula in the requirement,
unconditional, and a negative test confirms it rejects a mismatch.
