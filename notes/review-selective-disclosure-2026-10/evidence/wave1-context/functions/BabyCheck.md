## `BabyCheck` in armada-circuits/node_modules/circomlib/circuits/babyjub.circom (L69-L83)

Vendored circomlib 2.0.5 template, not modified by this PR.

**Purpose:** Asserts that (x, y) lies on the BabyJubJub twisted-Edwards curve
`a*x^2 + y^2 = 1 + d*x^2*y^2` with the circomlib-standard parameters a=168700, d=168696.
`NullifierDisclosure` instantiates it twice, for the spending public key A and for the signature's
R8, to close the on-curve gap left by `EdDSAPoseidonVerifier` (see that file's analysis).

**Inputs & Assumptions:**
- `x`, `y` (field elements): fully untrusted; that is the point of the check.

**Outputs & Effects:** No output signals; one equality constraint.

**Block-by-Block:**

```circom
// L73-L82
signal x2; signal y2;
var a = 168700; var d = 168696;
x2 <== x*x;
y2 <== y*y;
a*x2 + y2 === 1 + d*x2*y2;
```
- **What:** Computes x^2, y^2 and asserts the curve equation.
- **Why here:** A single multiplicative constraint; no ordering dependency within the template.
- **Assumes:** `a`, `d` are the correct BabyJubJub parameters (hardcoded, same values used
  throughout circomlib's `babyjub.circom`, e.g. `BabyAdd` at L31-32 of the same file).
- **Establishes:** (x, y) is a point on the curve (including the point at infinity-equivalent
  representations the twisted-Edwards form admits, e.g. (0,1) and (0,-1), which are low-order
  points still "on-curve" by this equation alone -- see Open Questions).

**Cross-Function Dependencies:**
- Callers: `NullifierDisclosure.circom:96-98` (on `spendingPublicKey`, i.e. A) and
  `NullifierDisclosure.circom:99-101` (on `signature[0..1]`, i.e. R8). Both callers rely on this
  template to reject any (x,y) not satisfying the curve equation, closing the gap identified in
  `EdDSAPoseidonVerifier.md` (that template's `BabyAdd`/`BabyDbl` calls do not themselves check
  curve membership).
- No callees.
- Shared state: none.

**Open Questions:**
- `BabyCheck` on its own accepts low-order curve points, e.g. (0, 1) and (0, -1 mod p), which
  satisfy `a*0 + 1 = 1 + d*0*1`. For A specifically, `EdDSAPoseidonVerifier`'s separate small-order
  check (`eddsaposeidon.circom:80-83`, analyzed in `EdDSAPoseidonVerifier.md`) covers the identity
  case (8*A = identity) but was not traced in this pass for every low-order point individually
  (orders 1, 2, 4, 8) -- whether every non-identity low-order point on the curve is also excluded
  by that `isZero` check, or only the identity itself, is `nothing found` resolving in this
  analysis pass; the disclosure circuit's own test only covers A=0-order via
  `rejects: a forged signature for a small-order spending key`
  (`tests/disclosure.test.js:264-281`, using A=(0,1) and A=(0,p-1), both of which the test
  confirms are rejected).
