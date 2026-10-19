## REQ-12 — Verification status: our wiring, Picus strong safety

> "Our wiring: the range checks, the equality checks, the on-curve checks and the message
> routing | Picus with cvc5, strong safety, with the hashes and the verifier replaced by free
> inputs: properly constrained | Proof of uniqueness"
> — engineering/16-selective-disclosure.md, table "Verification status"

**Verdict:** implemented · confidence: high

**What this demands:** a Picus run in strong-safety mode over a reduced circuit — the real
wiring with each Poseidon output and the EdDSA verifier's input message replaced by free
inputs/outputs — MUST report "properly constrained."

**Where enforcement lives:** `armada-circuits/tests/picus/wiring.circom` mirrors
`NullifierDisclosure.circom`'s range checks (`Num2Bits(16)` on `leafIndex`/`treeNumber`),
equality checks (`masterPublicKey === mpkHash`, `nullifier === nfHash`), on-curve checks
(`BabyCheck()` on `spendingPublicKey` and `signature[0..1]`), with `mpkHash`, `nfHash`, `msgHash`
as free inputs and `verifierM` as an output standing for the EdDSA message. `scripts/picus.sh`
invokes it with `--strong`, expecting exit code 8 (properly constrained).

**Verification performed:** read the committed run log,
`armada-circuits/build/picus/wiring.log`:
```
The circuit is properly constrained
Exiting Picus with the code 8
```
Matches the expected exit code in `scripts/picus.sh` (`check wiring 8 --strong`).

**How the verdict was reached:** the Picus input file is a faithful reduction of the real
circuit (checked side by side), and the logged result is the claimed one.
