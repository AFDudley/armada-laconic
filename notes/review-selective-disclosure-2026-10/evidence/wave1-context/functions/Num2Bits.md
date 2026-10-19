## `Num2Bits` in armada-circuits/node_modules/circomlib/circuits/bitify.circom (L25-L39)

Vendored circomlib 2.0.5 template, not modified by this PR. `NullifierDisclosure` instantiates
`Num2Bits(16)` twice (for `leafIndex` and `treeNumber`) and `EdDSAPoseidonVerifier` instantiates
`Num2Bits(253)` once (for S).

**Purpose:** Range-checks `in` to n bits by decomposing it into bits and asserting the bits
recompose to `in` exactly.

**Inputs & Assumptions:**
- `in` (field element), `n` (compile-time parameter, the bit width): `n` is fixed by the caller at
  compile time, not a signal.

**Outputs & Effects:**
- `out[n]`: the bit decomposition (little-endian). In `NullifierDisclosure`'s two uses
  (`NullifierDisclosure.circom:73-76`), `out` is never read afterward -- the components
  `leafBits`/`treeBits` exist purely for the side effect of the range constraint, a standard
  circomlib usage pattern.

**Block-by-Block:**

```circom
// L31-L38
for (var i = 0; i<n; i++) {
    out[i] <-- (in >> i) & 1;
    out[i] * (out[i] -1 ) === 0;
    lc1 += out[i] * e2;
    e2 = e2+e2;
}
lc1 === in;
```
- **What:** For each bit position, assigns the witness bit (`<--`, unconstrained assignment),
  asserts it is boolean (`out[i]*(out[i]-1) === 0`), and accumulates the weighted sum `lc1`. After
  the loop, asserts `lc1 === in`.
- **Why here:** The boolean assertion per bit must hold for the final sum to be meaningful as a
  bit decomposition; the final `lc1 === in` is what actually enforces the range, since `lc1` can
  only take values in [0, 2^n - 1] given each `out[i]` is boolean.
- **Assumes:** nothing beyond the field being large enough that 2^n - 1 does not wrap around p
  (true here: n ≤ 253 and p is the ~254-bit BN254/BabyJubJub scalar field).
- **Establishes:** `in` ∈ [0, 2^n). This is the line the disclosure circuit's test suite names
  directly: `FAIL.range16 = /Num2Bits_\d+ line: 38/` (`tests/disclosure.test.js:51`), i.e. line 38
  of this file is exactly the `lc1 === in` constraint confirmed above.

**Cross-Function Dependencies:**
- Callers (in the files under this review's scope): `NullifierDisclosure.circom:73-76` (n=16, for
  `leafIndex` and `treeNumber`); `eddsaposeidon.circom:42-43` (n=253, for S, inside
  `EdDSAPoseidonVerifier`).
- No callees of its own (pure arithmetic template).
- Shared state: none.

**Open Questions:** none -- this template is small and its single constraint is fully traced
above.
