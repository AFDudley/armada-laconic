## `loadDisclosureArtifacts` in /home/rix/code/git_puller/repos/armada-sdk/src/node/disclosure-artifacts.ts (L13-L24)

**Purpose:** Reads the compiled disclosure circuit artifacts (WebAssembly, zkey, and verification key) from the local filesystem and assembles them into an `ArtifactSet` for the prover. Node-only; kept off browser entries to avoid bundling large binary files.

**Inputs & Assumptions:**
- `baseDir` (string): A filesystem path pointing to the directory `armada-circuits/build/disclosure/`. Trust: semi-trusted (supplied by the host application; if malicious, can read arbitrary files from the filesystem).
  - Precondition: The directory must exist and contain three files:
    - `main_disclosure_js/main_disclosure.wasm`
    - `final.zkey`
    - `vkey.json`
  - The function does not verify these files exist or are readable before attempting to read them; it will reject if any file is missing or unreadable (see "Outputs & Effects").

**Outputs & Effects:**
- Returns a `Promise<ArtifactSet>` (prover/index.ts:17-22) with:
  - `wasm`: `Uint8Array` of the WebAssembly module binary.
  - `zkey`: `Uint8Array` of the Groth16 proving key.
  - `vkey`: `object` (parsed from JSON, the Groth16 verification key).
  - `vkeyRaw`: `Uint8Array` of the raw vkey.json bytes (optional, but provided here).
- Rejects with an error if:
  - Any file does not exist or cannot be read (Node.js `ENOENT` or `EACCES` errors propagate as rejected promises).
  - The vkey.json file is not valid UTF-8 or valid JSON (the `JSON.parse` call at line 20 will throw).
  - The file paths are invalid (e.g., contain null bytes).
- No state writes; no external effects (only filesystem reads).

**Block-by-Block:**

```typescript
const [wasm, zkey, vkeyBytes] = await Promise.all([
  readFile(join(baseDir, 'main_disclosure_js', 'main_disclosure.wasm')),
  readFile(join(baseDir, 'final.zkey')),
  readFile(join(baseDir, 'vkey.json')),
]);
```
- **What:** Reads three files in parallel using `Promise.all`, waiting for all three to complete before continuing.
- **Why here:** Parallelizes I/O to reduce total latency. Must complete before vkey parsing and the return statement.
- **Assumes:** `readFile` from `node:fs/promises` (Node.js built-in) reads a file and returns a `Promise<Buffer>`. If any read fails, the promise rejects, and `Promise.all` rejects immediately. The three file paths are constructed by `join` from `node:path`, which concatenates path segments with the OS separator. The function trusts that `baseDir` is correctly provided and that the expected file structure exists at that location.
  - Precondition: `baseDir` must be readable and must point to a directory with the specified subdirectories and files.
- **Establishes:** Three `Buffer` objects containing the file contents, or a rejection.
- **Depended on by:** The next two blocks (conversion to `Uint8Array` and JSON parsing).

```typescript
const vkeyRaw = new Uint8Array(vkeyBytes);
```
- **What:** Converts the vkey.json file bytes (a `Buffer`) to a `Uint8Array`.
- **Why here:** Preserves the raw bytes for integrity checks later (the `ArtifactSet.vkeyRaw` field, prover/index.ts:23).
- **Assumes:** `Buffer` is iterable and can be used to construct a `Uint8Array` via the buffer constructor (which copies the bytes). The bytes are assumed to be UTF-8 encoded JSON.
- **Establishes:** `vkeyRaw` is a byte array matching the original file contents.
- **Depended on by:** The return statement (included in the returned object).

```typescript
return {
  wasm: new Uint8Array(wasm),
  zkey: new Uint8Array(zkey),
  vkey: JSON.parse(new TextDecoder().decode(vkeyRaw)) as object,
  vkeyRaw,
};
```
- **What:** Constructs and returns an `ArtifactSet` object with:
  - `wasm` and `zkey` as byte arrays (converted from `Buffer`).
  - `vkey` as a parsed JSON object (decoded from UTF-8 bytes and parsed).
  - `vkeyRaw` as the original byte array (optional, provided for integrity checks).
- **Why here:** Final return; assembles the loaded data into the interface the prover expects.
- **Assumes:**
  - `new Uint8Array(wasm)` and `new Uint8Array(zkey)` copy the `Buffer` contents into byte arrays. The wasm file is expected to be a valid WebAssembly module (not checked here; the prover will reject if it is not).
  - `new TextDecoder().decode(vkeyRaw)` decodes the bytes as UTF-8. If the bytes are not valid UTF-8, this throws a `TypeError` (by default; the `fatal` option is not set, so invalid UTF-8 characters are replaced with U+FFFD).
  - `JSON.parse()` parses the decoded string. If the JSON is invalid, this throws a `SyntaxError`.
  - The resulting object is cast to `object` for type satisfaction, but no schema is checked (the verification key may be malformed, and the prover will detect this when it tries to use it).
- **Establishes:** An `ArtifactSet` ready to pass to the prover.
- **Depended on by:** Callers (e.g., the prover initialization) that use the artifacts to generate proofs or verify them.

**Cross-Function Dependencies:**
- `readFile` (node:fs/promises): Reads files from the filesystem; depends on it to read the files and reject if they do not exist or are unreadable.
- `join` (node:path): Constructs paths by concatenating segments; depends on it to produce the correct path on the current OS.
- `TextDecoder` (global, available in Node.js and browsers): Decodes UTF-8 bytes to a string; depends on it to parse the vkey.json bytes.
- `JSON.parse` (built-in): Parses JSON; depends on it to deserialize the verification key object.
- `Uint8Array` (built-in): Constructs byte arrays; depends on it to convert `Buffer` objects to the expected interface type.
- Callers: The prover or circuit initialization code that needs the compiled artifacts. E.g., the checker module (in src/disclosure/* or a separate checker crate) would call this to load artifacts before generating proofs.

**Open Questions:**
- **File validation:** The function reads the files but does not validate their contents or sizes. A corrupted wasm file (e.g., truncated, or with wrong magic bytes) will not be detected until the prover tries to load it as a WASM module. The same applies to the zkey file (the Groth16 proving key); it should match the wasm's constraints, but this is not checked here. Should the function validate the WASM magic bytes (`0x00 0x61 0x73 0x6d`) and the zkey header?
- **Hash integrity:** The ArtifactSet.vkeyRaw field is provided (line 23), which is used for integrity checks (prover/index.ts:23, "Integrity hashes the raw bytes"). But the function does not check the integrity itself; it assumes a manifest is verified separately (e.g., by `verifyArtifactIntegrity` in prover/manifest.ts, exported from prover/index.ts:70). Should this function check the integrity, or is it the caller's responsibility?
- **Path traversal:** The `baseDir` parameter is concatenated with relative paths like `'main_disclosure_js/main_disclosure.wasm'` using `join`. If a caller passes a malicious `baseDir` or if the path segments could be influenced by an attacker, path traversal is possible (e.g., `baseDir = '/tmp/'; realPath = join(baseDir, '../../etc/passwd')`). The function does not validate `baseDir` (e.g., check if it is within a safe directory). Is this a concern, or is `baseDir` always controlled by the SDK?
- **WASM/zkey/vkey version matching:** The three files are assumed to be from the same circuit build (same version of NullifierDisclosure.circom and the same Phase-2 ceremony). If a caller mixes artifacts from different builds, the prover will fail. Should the function validate that the artifacts are compatible (e.g., by checking a version header in each file)?
- **Large file handling:** The function reads entire files into memory without checking their size. A malicious or corrupted `final.zkey` (which is typically hundreds of MB) could consume all available memory. Should the function impose a size limit?
- **TextDecoder error handling:** If the vkey.json bytes are not valid UTF-8, `TextDecoder.decode()` by default replaces invalid sequences with U+FFFD (the replacement character), rather than throwing. The resulting string may be invalid JSON or parse to an incorrect object. Should the function use the `fatal: true` option to throw on invalid UTF-8, or is silent replacement acceptable?

