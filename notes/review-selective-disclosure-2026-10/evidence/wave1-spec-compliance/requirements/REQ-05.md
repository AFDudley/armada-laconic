## REQ-05 — Circuit table: (Ax, Ay) and signature private, on-curve + EdDSA verify

> "(Ax, Ay), signature (R8, S_sig) | private | Both points are on BabyJubJub. The EdDSA signature
> by (Ax, Ay) over `Poseidon(S, M, N, T, L)` verifies."
> — engineering/16-selective-disclosure.md, table "The disclosure circuit"

**Verdict:** implemented · confidence: high

**Where enforcement lives:**

```circom
// NullifierDisclosure.circom:L93-L105
component aOnCurve = BabyCheck();
aOnCurve.x <== spendingPublicKey[0];
aOnCurve.y <== spendingPublicKey[1];
component rOnCurve = BabyCheck();
rOnCurve.x <== signature[0];
rOnCurve.y <== signature[1];

component eddsa = EdDSAPoseidonVerifier();
eddsa.enabled <== 1;
eddsa.Ax <== spendingPublicKey[0];
eddsa.Ay <== spendingPublicKey[1];
eddsa.R8x <== signature[0];
eddsa.R8y <== signature[1];
eddsa.S <== signature[2];
eddsa.M <== msg.out;
```

`eddsa.enabled <== 1` is a literal, not a signal that any witness can set to 0, so the check
cannot be disabled. `msg.out` is `Poseidon(statementHash, masterPublicKey, nullifier, treeNumber,
leafIndex)` (REQ-08), matching `Poseidon(S, M, N, T, L)`.

Confirmed by rejection tests: `rejects: points off the curve` (`tests/disclosure.test.js:253-260`,
`BabyCheck_N line: 82`), `rejects: a signature by another spending key`
(`tests/disclosure.test.js:219-222`), `rejects: a malleable signature (S + subgroup order)`
(`tests/disclosure.test.js:241-245`, `S < subgroup order` inside `EdDSAPoseidonVerifier`),
`rejects: a forged signature for a small-order spending key` (`tests/disclosure.test.js:263-277`,
the `8*A != identity` check inside `EdDSAPoseidonVerifier`).

**How the verdict was reached:** every sub-clause (on-curve for both points, EdDSA verify over
the exact message, unconditionally enabled) has a direct constraint and a direct negative test.
