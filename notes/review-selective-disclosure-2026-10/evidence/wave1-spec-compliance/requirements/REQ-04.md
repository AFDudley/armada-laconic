## REQ-04 — Circuit table: statement hash S public, no additional constraint

> "statement hash S | public | —"
> — engineering/16-selective-disclosure.md, table "The disclosure circuit"

**Verdict:** implemented · confidence: high

**What this demands of an implementation:** `statementHash` is a public signal with no
constraint of its own beyond entering the signed message (REQ-08).

**Where enforcement lives:**

```circom
// NullifierDisclosure.circom:L60, L86-L91 and src/main/main_disclosure.circom
signal input statementHash;
...
component msg = Poseidon(5);
msg.inputs[0] <== statementHash;
```
```circom
// src/main/main_disclosure.circom
component main {public [statementHash, masterPublicKey, nullifier, treeNumber, leafIndex]} = NullifierDisclosure();
```

`statementHash` carries no range check and no equality constraint of its own; it is used only as
`msg.inputs[0]`. This matches the table's "—" exactly.

**How the verdict was reached:** absence of a dedicated constraint is correct here, since the
requirement itself specifies none.
