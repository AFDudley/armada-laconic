## `hashHops` in src/disclosure/statement.ts (L74-L83)

**Purpose:** Produce the canonical keccak256 hash of the hop list, which binds the EIP-712 statement to the exact chain of custody the user claims. The hash is deterministic: the same hops in the same order always produce the same hash, so a verifier can reproduce it. Two different hop lists hash to different values, binding the statement to one claim only.

**Inputs & Assumptions:**
- `hops` (readonly Hop[]): an array of hop objects (each is one of Hop1-Hop6 union types). Trust: semi-trusted (comes from the user's disclosure package; the hash commits the statement to whatever is provided, but the statement itself must be verified).
- Implicit: The `canonicalJson()` function (L63-L72) produces a deterministic string representation. The `keccak256()` and `toUtf8Bytes()` are from ethers.js.
- Preconditions: No validation. Assumes the hop list is non-empty (an empty list is valid but unusual); each hop object is a valid hop shape (one of H1-H6).

**Outputs & Effects:**
- Returns a 0x-prefixed 32-byte hex string (the keccak256 hash).
- No state writes, no external calls.
- The hash is reproducible: given the same hop list, the same hash is returned. Used as the `hopsHash` field in the statement, binding the EOA signature to this exact hop list.

**Block-by-Block:**

```typescript
// L63-L72
function canonicalJson(value: unknown): string {
  if (value === null || typeof value !== 'object') {
    return JSON.stringify(value);
  }
  if (Array.isArray(value)) {
    return `[${value.map(canonicalJson).join(',')}]`;
  }
  const entries = Object.entries(value as Record<string, unknown>)
    .filter(([, v]) => v !== undefined)
    .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0));
  return `{${entries.map(([k, v]) => `${JSON.stringify(k)}:${canonicalJson(v)}`).join(',')}}`;
}
```

- **What:** Recursively serialize a value to canonical JSON with sorted object keys at every depth.
- **Why here:** EIP-712 hashes are not inherently canonical; the hop list object must serialize identically across implementations (SDK, verifier, off-chain tools). Sorting keys guarantees canonicality.
- **Assumes:** `Object.entries()` preserves the values correctly. The filter removes undefined fields, consistent with `statementMessage()`'s use of hop data. The sort is lexicographic (JavaScript string comparison).
- **Establishes:** A string representation of the hops where the same object always serializes to the same bytes (property of canonical JSON). No two different objects can serialize to the same string.
- **Depended on by:** `hashHops()` L77 — the canonical JSON is the input to keccak256.

```typescript
// L74-L76
export function hashHops(hops: readonly Hop[]): string {
  return keccak256(toUtf8Bytes(canonicalJson(hops)));
}
```

- **What:** Hash the canonical JSON bytes of the hop list.
- **Why here:** The statement's `hopsHash` field is a 32-byte digest; ethers' `keccak256()` takes UTF-8 bytes and returns a 0x-prefixed hex string (the digest).
- **Assumes:** `canonicalJson(hops)` returns a valid UTF-8 string. `toUtf8Bytes()` succeeds (it does not validate content, only encodes to UTF-8). `keccak256()` returns a 0x-prefixed 32-byte hash as a hex string.
- **Establishes:** The hop list is now bound to the statement via a cryptographic hash. The statement signer commits to this exact hop list. Changing any hop (adding, removing, reordering, or modifying a field) changes the hash.
- **Depended on by:** `buildDisclosurePackage()` builder.ts L95 — `hashHops()` is called once per package and stored in the statement.

**Cross-Function Dependencies:**
- Callee `keccak256` (ethers external-source-available): cryptographic hash function. Per ethers docs and EIP-712, `keccak256(bytes)` returns a 0x-prefixed 32-byte hex string.
- Callee `toUtf8Bytes` (ethers external-source-available): encodes a string to UTF-8 bytes. Takes any string, returns a Uint8Array.
- Callee `canonicalJson` (internal): produces the hop list serialization. Fully internal, no external state.
- Callers: `buildDisclosurePackage()` builder.ts (L95), called when constructing the statement to bind it to the hop list.

**Open Questions:**
- None. The function is deterministic and has no hidden dependencies.
