## REQ-37 — D20.4: no change to pool/circuits/nitro contracts; never a pool-use condition

> "**D20.4** The protocol needs no change to the pool, its circuits or the armada-nitro
> contracts. It must never become a condition to deposit into, transfer in, or withdraw from the
> pool."
> — engineering/09-architecture-decisions.md, ADR-0020
(restated in the "Policy" section of engineering/16-selective-disclosure.md)

**Verdict:** implemented · confidence: high

**What this demands:** the patch under review MUST NOT modify any pool contract, pool circuit, or
armada-nitro contract, and no pool/nitro code path may call into the disclosure module.

**Where enforcement lives:** the circuit PR adds only new files (`NullifierDisclosure.circom`,
`main_disclosure.circom`, their tests and Picus pieces) under `armada-circuits`; `Transact.circom`
(read for reference, REQ-01/REQ-02/REQ-08's formulas) is unmodified by this change. The SDK PR
adds only new files under `src/disclosure/` and `src/node/disclosure-artifacts.ts`; nothing in
`armada-nitro/src/ArmadaNitroAdjudicator.sol` (read for REQ-20/REQ-26's formulas) is part of this
patch. No file under `src/disclosure/` is imported by the SDK's core shield/transact/sync code
(confirmed via REQ-34's isolation check), so disclosure cannot be wired into a deposit/transfer/
withdrawal path by this patch.

**How the verdict was reached:** scope check across both target repos — the only new or changed
files are under `armada-circuits/{src/operations/NullifierDisclosure.circom, src/main/main_disclosure.circom,
tests/disclosure.test.js, tests/picus/*, scripts/picus.sh}` and
`armada-sdk/src/disclosure/*` + `src/node/disclosure-artifacts.ts` + `docs/guide/selective-disclosure.md`;
no pool/circuit/nitro-contract file is touched.
