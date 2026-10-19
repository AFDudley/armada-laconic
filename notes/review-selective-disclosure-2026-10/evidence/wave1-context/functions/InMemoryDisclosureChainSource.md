## `InMemoryDisclosureChainSource` in /home/rix/code/git_puller/repos/armada-sdk/src/disclosure/chain-source.ts (L57-103)

**Purpose:** An in-memory implementation of the DisclosureChainSource interface. Indexes pre-decoded pool events (Shield, Transact, Nullified, Unshield) and hub records (Deposited, finalized outcomes) by transaction ID, enabling the disclosure checker to query for events without network I/O. Used in tests and by callers that already have events in memory.

**Inputs & Assumptions:**

- `data` (DisclosureChainData): a struct containing optional arrays of decoded events and an optional map of txid → sender. Trust: untrusted; comes from the caller (likely provided from untrusted sources like raw logs, an indexer, or a cache).
- Preconditions:
  - Each event in the input arrays has a valid `txid` field (string). No validation occurs; if txid is undefined, the indexing will fail.
  - The arrays are in any order; the implementation does not assume sorted order.
  - Sender addresses (if provided) are already checksummed or normalized (the code stores them as-is).

**Outputs & Effects:**

- Constructor: builds internal index maps (shields, transacts, nullifiers, unshields, deposits, finalizedOutcomes) and stores the senders map. No external calls.
- Per-query methods: return an empty array (not undefined) if the txid is not in the index, or the events that match the txid. No side effects; read-only access.

**Block-by-Block:**

```typescript
// L50-55: Interface definition (DisclosureChainSource)
export interface DisclosureChainSource {
  shieldsIn(txid: string): readonly DecodedShieldCommitment[];
  transactsIn(txid: string): readonly DecodedTransactCommitment[];
  nullifiersIn(txid: string): readonly DecodedNullifier[];
  unshieldsIn(txid: string): readonly DecodedUnshield[];
  depositsIn(txid: string): readonly HubDepositRecord[];
  finalizedOutcomesIn(txid: string): readonly HubFinalizedOutcomeRecord[];
  senderOf(txid: string): string | undefined;
}
```
- **What:** The read interface the checker calls. Every method takes a txid and returns events/data from that transaction.
- **Why here:** The checker is decoupled from the event source; different sources (chain RPC, indexer, test fixture) can implement this interface.
- **Establishes:** A contract for what the source must provide. The return types all specify `readonly`, so the checker cannot modify the underlying data.
- **Depended on by:** checkDisclosurePackage (checker.ts), which calls every method except senderOf in the hop loop, and senderOf optionally in H1.

```typescript
// L46-49: Input shape (DisclosureChainData)
export interface DisclosureChainData {
  readonly shields?: readonly DecodedShieldCommitment[];
  readonly transacts?: readonly DecodedTransactCommitment[];
  readonly nullifiers?: readonly DecodedNullifier[];
  readonly unshields?: readonly DecodedUnshield[];
  readonly deposits?: readonly HubDepositRecord[];
  readonly finalizedOutcomes?: readonly HubFinalizedOutcomeRecord[];
  readonly senders?: Readonly<Record<string, string>>;
}
```
- **What:** The shape the caller passes to the constructor. Every array is optional (may be undefined or omitted).
- **Why here:** Flexibility for tests and callers that may not have all event types available (e.g., a test with only shield commitments).
- **Establishes:** The constructor contract; missing fields are treated as empty arrays.
- **Depended on by:** The constructor (L64-70).

```typescript
// L38-45: Indexing helper
function indexByTxid<T extends { readonly txid: string }>(rows: readonly T[] | undefined): Map<string, T[]> {
  const index = new Map<string, T[]>();
  for (const row of rows ?? []) {
    const bucket = index.get(row.txid);
    if (bucket === undefined) {
      index.set(row.txid, [row]);
    } else {
      bucket.push(row);
    }
  }
  return index;
}
```
- **What:** Build a map from txid → array of events for a given event type.
- **Why here:** The indexing happens once in the constructor; subsequent queries are O(1) by txid.
- **Assumes:**
  - Every row has a `txid` field (the generic constraint `T extends { readonly txid: string }`).
  - The rows array is finite and does not change during indexing.
  - Multiple rows can have the same txid (a single transaction can contain multiple Shield events, multiple Nullified events, etc.).
- **Establishes:** A map ready for O(1) lookups. If a txid has no rows, the map does not have a key for it; a missing key will later return undefined from Map.get, which the query methods convert to an empty array.
- **Depended on by:** The constructor (L72-77), once for each event type.

```typescript
// L57-70: Constructor
export class InMemoryDisclosureChainSource implements DisclosureChainSource {
  private readonly shields: Map<string, DecodedShieldCommitment[]>;
  private readonly transacts: Map<string, DecodedTransactCommitment[]>;
  private readonly nullifiers: Map<string, DecodedNullifier[]>;
  private readonly unshields: Map<string, DecodedUnshield[]>;
  private readonly deposits: Map<string, HubDepositRecord[]>;
  private readonly finalizedOutcomes: Map<string, HubFinalizedOutcomeRecord[]>;
  private readonly senders: Readonly<Record<string, string>>;

  constructor(data: DisclosureChainData) {
    this.shields = indexByTxid(data.shields);
    this.transacts = indexByTxid(data.transacts);
    this.nullifiers = indexByTxid(data.nullifiers);
    this.unshields = indexByTxid(data.unshields);
    this.deposits = indexByTxid(data.deposits);
    this.finalizedOutcomes = indexByTxid(data.finalizedOutcomes);
    this.senders = data.senders ?? {};
  }
```
- **What:** Initialize the source with indexed event maps and the sender map.
- **Why here:** Once-per-constructor setup; subsequent queries do not re-index.
- **Assumes:** The indexByTxid helper correctly builds the maps; events are not modified after construction.
- **Establishes:** The object is ready for queries (shieldsIn, transactsIn, etc.).
- **Postcondition:** The source is immutable; the maps are private and readonly.

```typescript
// L72-78: Query methods (one example; all others follow the same pattern)
shieldsIn(txid: string): readonly DecodedShieldCommitment[] {
  return this.shields.get(txid) ?? [];
}

transactsIn(txid: string): readonly DecodedTransactCommitment[] {
  return this.transacts.get(txid) ?? [];
}

nullifiersIn(txid: string): readonly DecodedNullifier[] {
  return this.nullifiers.get(txid) ?? [];
}

unshieldsIn(txid: string): readonly DecodedUnshield[] {
  return this.unshields.get(txid) ?? [];
}

depositsIn(txid: string): readonly HubDepositRecord[] {
  return this.deposits.get(txid) ?? [];
}

finalizedOutcomesIn(txid: string): readonly HubFinalizedOutcomeRecord[] {
  return this.finalizedOutcomes.get(txid) ?? [];
}

senderOf(txid: string): string | undefined {
  return this.senders[txid];
}
```
- **What:** Return events from a transaction, or an empty array if not found.
- **Why here:** The interface contract. Each method queries the pre-indexed map and returns the result or an empty array.
- **Assumes:**
  - The txid is a valid string (no normalization; txid is case-sensitive and must match exactly).
  - The caller does not modify the returned array (it is marked readonly).
- **Establishes:**
  - If the txid is found, the array contains all events of that type for that txid.
  - If the txid is not found, the method returns an empty array (not undefined, not null). This is critical: the checker relies on this (e.g., checker.ts L170, L319 use `.some()` and `.find()` on the returned arrays, which work on empty arrays).
  - `senderOf` returns undefined if the txid is not in the senders map (note: this is different from the array-returning methods, which return `[]`; see Open Questions).
- **Depended on by:** checkDisclosurePackage (checker.ts), which calls these methods for every hop and expects empty arrays when no events are found.

**Cross-Function Dependencies:**

- **DisclosureChainSource interface** (defined above, L50-55): The source implements this interface. The checker depends on the contract that all methods return consistent types.

- **DecodedShieldCommitment, DecodedTransactCommitment, DecodedNullifier, DecodedUnshield** (event-decoder.ts): The types of the pool events. The source stores and returns these types; the checker filters and compares them (e.g., by tree, position, nullifier, hash).

- **HubDepositRecord, HubFinalizedOutcomeRecord** (types in chain-source.ts, L13-32): The types for hub records. The source returns these; the checker compares them against the hop's claimed values.

- **indexByTxid** (internal): Called during construction to build the index maps. The correctness of the maps depends on the correctness of indexByTxid (which is straightforward: it groups rows by txid).

**Implementations in Other Modules:**

The checker.ts file also refers to a DisclosureChainSource, which is an interface. There may be other implementations:
- **Test fixture sources:** built in-memory for unit tests, using this class or a hand-written version.
- **RPC-based sources:** fetch events from an Ethereum RPC, decode them, and return them.
- **Indexer-based sources:** query an indexer (e.g., The Graph, Ponder, Subsquid) and return events.

The checker does not care which implementation is used; it only cares that the interface is met.

**Open Questions:**

1. **Empty array vs. undefined.** The array-returning methods (shieldsIn, transactsIn, etc.) return `[]` when a txid is not found, but `senderOf` returns `undefined`. The checker uses `.some()` and `.find()` on the array methods (which return `[]` safely), but checks `if (sender !== undefined)` for senderOf. This asymmetry is intentional but could be confusing for a new implementer. Should the interface document this? The spec does not mention how an absent sender should be handled.

2. **No validation of event data.** The constructor accepts any array of events and indexes them by txid. There is no check that the events are well-formed (e.g., that txid is a valid transaction hash, that tree and position are non-negative, that hashes are valid hex). A malicious or corrupted data source could cause the checker to accept false disclosures. For example, a source that returns fake commitment hashes could make the checker pass a false opening. Is there a way to verify the integrity of the indexed events (e.g., a Merkle root check against a trusted reference)?

3. **txid matching is case-sensitive and exact.** The checker passes a txid from the package to the source; if the txid has a different case or padding (e.g., "0x123abc" vs. "0x123ABC"), the lookup will fail, and the source will return an empty array. The checker will then fail with "no Shield commitment at (txid)..." without indicating that the txid format was wrong. Is there a canonical format enforced anywhere? The event-decoder.ts does not show how txids are produced; they are passed in with LogMeta.

4. **Multiple events per txid are grouped but not ordered.** If a transaction has multiple Shield commitments, they are stored in an array in the order they were processed by indexByTxid. The checker then searches this array for a specific (tree, position) pair. If the array is not sorted or there is no other guarantee of order, a later (tree, position) lookup might miss an event if the array is incomplete or out of order. How is the order of events in the source determined?

5. **No binding between events and blocks.** The DecodedShieldCommitment includes a blockNumber (event-decoder.ts L73), but the InMemoryDisclosureChainSource indexes only by txid, not by blockNumber. The checker never uses blockNumber; it just checks that events exist at the right (tree, position) in the right txid. If the same txid appears in two different blocks (unlikely but possible in a reorg), the source cannot distinguish them. Is this a problem? The threat model is unclear.

6. **Sender data is unvalidated.** The senders map (L69) is stored as-is; if the caller provides malicious sender data, the checker will trust it. The sender check in H1 (checker.ts L217-219) is optional and comes after signature verification, so a false sender does not immediately break the disclosure. But an implementer who relies on the sender check might be surprised if the source provides incorrect data. Should the source validate that sender addresses are well-formed?

7. **No mechanism to report incomplete data.** If a caller creates an InMemoryDisclosureChainSource with only a subset of events (e.g., shields and transacts, but no nullifiers), the checker will silently treat missing event types as empty arrays. A verifier might accept a disclosure that references a nullifier that is not in the source. How should a caller know if the source is complete? Is there a way to audit the coverage?

8. **The readonly constraint is not enforced at runtime.** The methods return `readonly` arrays, but this is a TypeScript-only constraint. At runtime, a caller could cast the array to mutable, push events into the source's internal maps, and corrupt the index. Should the source return copies of the arrays to prevent mutation?

9. **Constructor does not validate the input data shape.** If the caller passes a DisclosureChainData object with malformed fields (e.g., an event without a txid field), the indexByTxid function will attempt to access `row.txid` on undefined and fail. The error will be a runtime exception, not a clear validation error. Should the constructor check the shape of the input?

10. **No cache key to distinguish different source instances.** If a verifier uses two different InMemoryDisclosureChainSource instances for the same disclosure (e.g., one with events and one without), the checker will use the first one and return different results depending on which source was passed. There is no way to detect this or to canonicalize the source. Is this a problem? The threat model is unclear.
