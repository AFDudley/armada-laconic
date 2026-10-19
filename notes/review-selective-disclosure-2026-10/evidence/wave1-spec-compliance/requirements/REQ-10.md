## REQ-10 — The circuit needs no Merkle path

> "The circuit needs no Merkle path, because the leaf position is public in the `Shield` and
> `Transact` events."
> — engineering/16-selective-disclosure.md, "The disclosure circuit"

**Verdict:** implemented · confidence: high

**What this demands:** `NullifierDisclosure.circom` MUST NOT include a Merkle-path component;
leaf location is established off-circuit.

**Where enforcement lives:**

`NullifierDisclosure.circom` includes only `poseidon.circom`, `bitify.circom`, `babyjub.circom`,
`eddsaposeidon.circom` (`L22-L25`); it does not include `../lib/merkle.circom`, unlike
`Transact.circom` (`include "../lib/merkle.circom"`, `Transact.circom:L18`, used at
`MerkleTreeProof(16)`, `Transact.circom:L74`). The off-circuit half of the claim is the
`checkSpendLeaf` comparison in `checker.ts` (REQ-18/REQ-22), present for `H2`/`H6`.

**How the verdict was reached:** absence of the include and the component is a direct,
grep-confirmed fact; it matches the claim exactly.
