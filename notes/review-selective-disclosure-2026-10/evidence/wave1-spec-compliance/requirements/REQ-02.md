## REQ-02 — Circuit table: nullifier constraint

> "`nullifier` N | public | `N == Poseidon(nullifyingKey, L)`"
> — engineering/16-selective-disclosure.md, table "The disclosure circuit"

**Verdict:** implemented · confidence: high

**What this demands of an implementation:** the circuit MUST constrain the public signal
`nullifier` to equal `Poseidon(nullifyingKey, leafIndex)`.

**Where enforcement lives:**

```circom
// NullifierDisclosure.circom:L82 (and NullifierOwnership L50-L53)
nullifier === own.nullifier;
```
```circom
component nf = Poseidon(2);
nf.inputs[0] <== nullifyingKey;
nf.inputs[1] <== leafIndex;
nullifier <== nf.out;
```

Unconditional `===`. Confirmed by `rejects: a nullifier of a different leaf`
(`tests/disclosure.test.js:207-211`), failing at `NullifierDisclosure_N line: 85`.

**Paths walked:** single path, no branch.

**How the verdict was reached:** formula matches exactly; negative test confirms enforcement.
