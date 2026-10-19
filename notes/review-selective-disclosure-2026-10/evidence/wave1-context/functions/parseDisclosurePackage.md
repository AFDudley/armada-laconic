## `parseDisclosurePackage` in /home/rix/code/git_puller/repos/armada-sdk/src/disclosure/serialize.ts (L24-L51)

**Purpose:** Parses a JSON string and validates its top-level structure to ensure it is a disclosure package with the right version and all required sections. Callers then populate detailed field-by-field checks (e.g., numeric bounds, signature verification, circuit checks) in the checker module.

**Inputs & Assumptions:**
- `json` (string): Untrusted JSON input from a user, network, or storage. Trust: untrusted.
  - Precondition: none explicitly stated; any string is accepted.

**Outputs & Effects:**
- Returns a `DisclosurePackage` if the JSON parses and the shape is valid.
- Throws `InvalidRequestError` (from errors.ts:163) on any of these failures:
  - JSON syntax error (line 32).
  - Not an object (line 36).
  - Missing or wrong `version` field (line 39).
  - Missing `domain` or `statement`, or they are not objects (lines 41-44).
  - Missing `endpointSignatures`, `openings`, `proofs`, or `hops`, or they are not arrays (lines 45-49).
- No state writes; no external effects.

**Block-by-Block:**

```typescript
let parsed: unknown;
try {
  parsed = JSON.parse(json);
} catch (err) {
  throw new InvalidRequestError('parseDisclosurePackage: invalid JSON', { cause: err });
}
```
- **What:** Parses the JSON string; catches syntax errors and re-throws as `InvalidRequestError` with the original error as the cause.
- **Why here:** Must happen first; subsequent checks depend on `parsed` being a value.
- **Assumes:** `JSON.parse` throws a `SyntaxError` (or similar) on invalid JSON, caught here. The error is wrapped in `InvalidRequestError` for a consistent error interface.
- **Establishes:** `parsed` is a value (any type); no structure is guaranteed yet.
- **Depended on by:** All subsequent checks.

```typescript
if (!isObject(parsed)) {
  throw new InvalidRequestError('parseDisclosurePackage: not an object');
}
```
- **What:** Checks that the top-level value is an object (not an array, string, number, null, etc.).
- **Why here:** The subsequent checks access object properties; a non-object would error when accessing `parsed.version`.
- **Assumes:** `isObject` (defined at lines 11-13) returns `true` iff the value is `typeof 'object'` and not `null`. This filters out `null` (which is `typeof 'object'` in JavaScript), arrays (which are objects but fail the shape checks later if they reach them), and primitives.
- **Establishes:** `parsed` is a non-null object, so property access is safe.
- **Depended on by:** Property access on `parsed` in lines 39 and 41-44.

```typescript
if (parsed.version !== DISCLOSURE_PACKAGE_VERSION) {
  throw new InvalidRequestError(`parseDisclosurePackage: unsupported version ${String(parsed.version)}`);
}
```
- **What:** Checks that the version field equals the expected version constant (DISCLOSURE_PACKAGE_VERSION, types.ts:25, which is `1`).
- **Why here:** Version mismatch is a top-level rejection; if the version is wrong, the rest of the structure may not match.
- **Assumes:** `parsed.version` is accessed and compared for equality. If the field is missing, `parsed.version` is `undefined`, which does not equal `1`, so the error is thrown. If the field is present but a different type (e.g., a string `"1"`), the strict equality comparison fails. The error message calls `String(parsed.version)` to include the actual value in the error.
- **Establishes:** The version is correct (or an error was thrown).
- **Depended on by:** Subsequent validation (version mismatch makes deeper checks pointless).

```typescript
for (const key of ['domain', 'statement'] as const) {
  if (!isObject(parsed[key])) {
    throw new InvalidRequestError(`parseDisclosurePackage: missing "${key}"`);
  }
}
```
- **What:** Checks that `domain` and `statement` are both present and are objects.
- **Why here:** These are the two required object sections; they must be present and shaped correctly for the checker to extract the statement hash, verifier, and other critical data.
- **Assumes:** `isObject` is applied to each field. If the field is missing, `parsed[key]` is `undefined`, which fails the check. If the field is present but not an object (e.g., a string or array), it fails. The error message says "missing" even if the field is present but not an object; this could be clearer, but the intent is that the field is not usable.
- **Establishes:** Both `parsed.domain` and `parsed.statement` are non-null objects.
- **Depended on by:** Callers that extract fields from these objects.

```typescript
for (const key of ['endpointSignatures', 'openings', 'proofs', 'hops'] as const) {
  if (!Array.isArray(parsed[key])) {
    throw new InvalidRequestError(`parseDisclosurePackage: "${key}" must be an array`);
  }
}
```
- **What:** Checks that `endpointSignatures`, `openings`, `proofs`, and `hops` are all present and are arrays.
- **Why here:** These are the four array sections; they must be arrays (not objects or primitives) for iteration in the checker.
- **Assumes:** `Array.isArray` correctly identifies arrays (and rejects objects, strings, null, etc.). If a field is missing, `parsed[key]` is `undefined`, and `Array.isArray(undefined)` returns `false`, triggering the error.
- **Establishes:** All four fields are arrays (possibly empty).
- **Depended on by:** Callers that iterate over these arrays.

```typescript
return parsed as unknown as DisclosurePackage;
```
- **What:** Casts the parsed object to `DisclosurePackage`, asserting the shape is correct.
- **Why here:** Final return; the function has validated the presence and type of the top-level fields, so the shape matches the interface (at least at the top level).
- **Assumes:** The cast is safe because the previous checks have verified:
  - Top-level is an object.
  - `version` field matches.
  - `domain` and `statement` are objects.
  - `endpointSignatures`, `openings`, `proofs`, `hops` are arrays.
  - The DisclosurePackage interface requires no other top-level fields, so extra fields are allowed (they will not cause a type error).
- **Establishes:** The return value is typed as `DisclosurePackage`, ready for use by the checker.
- **Depended on by:** Callers that receive the parsed package and perform field-level validation.

**Cross-Function Dependencies:**
- `JSON.parse` (built-in): Parses the JSON string; depends on it to accept valid JSON and throw on syntax errors.
- `InvalidRequestError` (src/errors.ts:163): Custom error class; used to wrap parse failures and validation errors.
- `isObject` (defined at lines 11-13): Local helper; filters non-objects.
- `DISCLOSURE_PACKAGE_VERSION` (types.ts:25): Version constant; compared against the parsed version.
- Callers: Any code that receives a disclosure package from an untrusted source (e.g., a user, an HTTP request) must call this function before consuming the package.

**Open Questions:**
- **Numeric field coercion:** The DisclosurePackage type includes numeric fields like `tree` (NoteOpening, types.ts:138), `position` (types.ts:139), `treeNumber` (DisclosureProofEntry, types.ts:150), `leafIndex` (types.ts:151), `chainId` (DisclosureDomain, types.ts:36), and `issuedAt` (DisclosureStatement, types.ts:56). JSON.parse preserves JSON number types (which are 64-bit floats), but JavaScript bigints in the package (like `masterPublicKey` string representations, types.ts:45, 51, types.ts:120) are JSON strings. The function does not validate that numeric fields fit in their intended ranges (e.g., `tree` and `position` should fit in 32 bits; `chainId` should be positive; `issuedAt` should be a Unix timestamp). If a JSON value is `"not a number"` or a float like `1.5` or `9007199254740992` (larger than JavaScript's safe integer limit), the parser accepts it, and the type cast succeeds. Does the checker validate these bounds?
- **String field size and encoding:** Fields like `masterPublicKey` and `nullifier` are decimal strings representing BN254 scalars (types.ts:44, 51, 121, types.ts:127). JSON.parse accepts arbitrarily long strings. Are there checks that these strings represent valid scalars (< SNARK_SCALAR_FIELD, types.ts:5)? Hex strings like `random` (types.ts:121) and `signature` (types.ts:78) should start with `0x` and contain exactly the right number of hex digits (16 bytes for random = 32 hex chars + 0x; 65 bytes for signature = 130 hex chars + 0x). Nothing here validates the format.
- **Array element count:** The arrays are checked to be arrays, but not for minimum/maximum size. Is there a limit on the number of proofs, hops, or signatures? If a package has 1 million hops, does the checker iterate all of them, or is there a bound?
- **Deep field types:** The objects within arrays (e.g., elements of `proofs`, which are `DisclosureProofEntry`) are not shape-checked. A `proofs` array could contain non-objects, or objects missing required fields. The checker must validate these individually.
- **Circular references:** JSON.parse rejects circular references in the input JSON (they are not valid JSON), so this is not a concern here. However, if the input is hand-crafted, a checker should be aware of the possibility.

