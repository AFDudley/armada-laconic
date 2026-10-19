## `serializeDisclosurePackage` in /home/rix/code/git_puller/repos/armada-sdk/src/disclosure/serialize.ts (L6-L8)

**Purpose:** Converts a typed `DisclosurePackage` object to its JSON wire format. Used to produce the audit record that a user forwards to a verifier.

**Inputs & Assumptions:**
- `pkg` (DisclosurePackage): A complete disclosure package with all required sections (domain, statement, signatures, openings, proofs, hops). Trust: trusted (the package is assumed to be already validated and correctly populated by SDK code).

**Outputs & Effects:**
- Returns a `string` containing the JSON-serialized package, pretty-printed with two-space indentation (for readability in audit records).
- No state writes; no external effects.

**Block-by-Block:**

```typescript
return JSON.stringify(pkg, null, 2);
```
- **What:** Serializes the object to JSON using the built-in `JSON.stringify` with `null` as the replacer (no filtering) and `2` as the indentation width.
- **Why here:** The entire function is this one call.
- **Assumes:** All fields on `pkg` are JSON-serializable (strings, numbers, booleans, arrays, objects). Looking at DisclosurePackage (types.ts:163-168), all fields are primitives or arrays/objects containing primitives:
  - `version` is a number literal (DISCLOSURE_PACKAGE_VERSION, types.ts:25).
  - `domain` and `statement` are objects with string and number fields.
  - `endpointSignatures`, `openings`, `proofs`, `hops` are arrays of serializable objects.
  - `proof` fields within DisclosureProofEntry are Groth16Proof objects (defined in prover/index.ts:48-52) with string fields `a`, `b`, `c` containing coordinate tuples.
  - `token` within NoteOpening is a TokenData object (re-exported from railgun-engine, core/index.ts:41), with string/number fields.
  - `random` and `npk` in NoteOpening are hex strings (0x-prefixed).
  - `hopsHash` and `signature` in statement/signatures are hex strings.
- **Establishes:** A JSON string suitable for transmission, storage, or display.
- **Depended on by:** Code that needs to forward the package (e.g., email, blockchain storage, HTTP endpoints).

**Cross-Function Dependencies:**
- `JSON.stringify` (built-in): Serializes the object. Depends on it to produce valid JSON for all field types present.
- Callers: Any code that needs to send a package to a verifier or store it durably. The reverse operation is `parseDisclosurePackage`.

**Open Questions:**
- Are there any cycles or non-serializable fields (e.g., `Date` objects, functions, `undefined` values) that could cause `JSON.stringify` to silently omit fields or error? The types.ts definitions do not show any, but a runtime check or test would confirm.
- The pretty-printing (indentation=2) is always applied. Is there a use case (e.g., compact transmission) that needs a minified version?

