# Pool model layout audit (2026-10-26)

## Scope

- Model: `armada-lean-pool-model`, branch `pool-model`, HEAD
  `7f2d1ec7ac499efce9d4f494b98036a1d64b2751`. The audit read a detached worktree at
  `armada-lean-pool-model/.scratch/wt/audit-pool`.
- Contract source: `armada-poc` at `ec8bdb97`. Line numbers below are from that commit.
  The model cites older line numbers in some places. The difference is a few lines.
- Armada rules: `labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md`
  (cited below as "CON").
- Adjudicator use sites: `armada-lean-adjudicator-v4` at `23e01b5` (remote
  `AFDudley/armada-lean`).
- Rule for the pool: the contract and the circuits are the source. A finding is an
  element that is storage or encoding detail, not protocol meaning. A finding is also an
  element that an Armada rule contradicts.
- "Statements" means the 31 theorems in `README.md` sections "Contract-owned notes" and
  "Pool contract". A statement "mentions" an element when its signature text names the
  element. Indirect use through `Step` or `Reachable` is not counted.

## Findings

### P-1. `token : Nat` — the token as a zero-extended address number

- Lean: `CommitmentPreimage.token : Nat` (`PoolOwned.lean:77`). `PoolState.live : Nat → …`
  and `PoolState.held : Nat → Nat` (`Pool.lean:1135`, `:1140`) take the same `Nat` key.
  `TransactWitness.token`, `TransactOwnedWitness.token` (`Transact.lean:72`,
  `TransactOwned.lean:83`).
- What it represents: the token that a note holds and that the pool keeps in custody.
- Class: encoding (fixed-width address cast).
- Evidence, model: `PoolOwned.lean:71-73`: "`token` abstracts `_getTokenID(token)`".
  `Transact.lean:61-62`: "`token`/`value` are raw numbers (a token ID field element …)".
- Evidence, source: `TransactModule.sol:449-454`, `_getTokenID`:
  `return bytes32(uint256(uint160(_tokenData.tokenAddress)));`. The comment at
  `ShieldModule.sol:381` says "ERC20 tokens are identified by their address,
  zero-extended". So the `Nat` is the output of an address cast.
- Inconsistency: under the operator's address ruling the model types every address as a
  Keccak `Term` (`PoolOwned.lean:34-61`). A token is an address too, but the model keeps
  it as the cast number. The `Nat` type also lets the model write the sentinel
  `pre.token = 0` (P-4).
- Statements whose text mentions it (10): `spend_live_exact` (`{token : Nat}`),
  `shield_creates_exactly`, `shield_can_happen`, `transact_can_happen`,
  `unshield_can_happen`, `owned_transact_can_happen`, `owned_spend_witness_exists`,
  `owned_spend_can_happen`, `live_note_unspent` (`t : Nat`), `live_leaf_matches`
  (`t : Nat`). `core_invariant` uses it through `HeldInvariant` (`∀ t`).
- Severity: low. [INFERENCE] No statement is false because of the `Nat` type. The type
  permits arithmetic and the `token = 0` sentinel, which have no protocol meaning.
- Replacement: an abstract `Token` type with decidable equality, or an address `Term`.
  The circuit layer applies `tokenId : Token → Term` inside `commitment`. Delete
  `token = 0` from the sentinel (P-4).

### P-2. `boundParamsHash : Nat` — a Keccak hash as a free number

- Lean: `TransactPublic.boundParamsHash : Nat` (`Transact.lean:103`),
  `TransactOwnedPublic.boundParamsHash : Nat` (`TransactOwned.lean:101`). `msgHashOf`
  hashes it as `Term.val pub.boundParamsHash` (`Transact.lean:113`).
- What it represents: the hash of the transaction's bound parameters. These are
  `treeNumber`, `unshield`, `chainID`, `adaptContract`, `adaptParams`,
  `commitmentCiphertext`, `minGasPrice`.
- Class: encoding (a hash stored instead of the value, and the hash as a number).
- Evidence, model: `Transact.lean:96-100`: "`boundParamsHash` is a plain field element
  (`keccak256(...) % SNARK_SCALAR_FIELD` …) — a different hash family this project has
  never modeled via Poseidon, so it stays a raw `Nat`". This reason is stale.
  `Poseidon.lean:120` now defines `Term.keccak`.
- Evidence, source: `PrivacyPoolStorage.sol:46-56`:
  `return uint256(keccak256(abi.encode(core))) % SNARK_SCALAR_FIELD;`.
- Effect: a Dolev-Yao attacker can choose any `Nat`. No field of the bound parameters
  exists in the model. `Step.transact` takes `treeNumber` as a free argument
  (`Pool.lean:1351`). Nothing ties it to the signed message, but the contract binds it
  through this hash. The unshield kind and the `adaptParams` (where CON decision 11
  puts the owned-note preimage) are absent for the same reason.
- Statements whose text mentions it (5): `owned_spend_can_happen`,
  `transact_can_happen`, `unshield_can_happen`, `owned_transact_can_happen`,
  `owned_spend_witness_exists`. `transact_spends_only_owned` uses it through
  `msgHashOf pub`.
- Replacement: a `BoundParams` record with the fields the protocol reads (`treeNumber`,
  `unshield : Option recipient`, `adaptParams`). The public input is
  `Term.keccak [encode boundParams]`, or `msgHashOf` hashes the record term directly.
  `Step.transact` reads `treeNumber` and the unshield from the record. The `% p`
  reduction belongs to the refinement layer.

### P-3. `shieldsPaused` and `Step.admin` — an owner switch that an Armada rule removes

- Lean: `PoolState.shieldsPaused : Bool` (`Pool.lean:1143`). `Step.shield`'s
  `hNotPaused` (`Pool.lean:1255`). `Step.admin` (`Pool.lean:1462-1463`).
- What it represents: an owner-installed pause controller
  (`PrivacyPool.setShieldPauseContract`).
- Class: contract-only rule. An Armada rule contradicts it.
- Evidence, model: `Pool.lean:138-141`: "only `shieldsPaused` remains … a shield pause,
  if the design keeps one, blocks only new value entering the pool". The model does not
  cite an Armada rule that keeps the pause.
- Evidence, source: `PrivacyPool.sol:466` (`setShieldPauseContract`, owner-gated),
  `ShieldModule.sol:394-399` (`_requireShieldsNotPaused`).
- Evidence, Armada rule: CON:173-176, decision 7: "The pool must not be upgradeable; it
  deploys with no proxy and no admin, and it sets verifying keys and fees once at
  deployment". With no admin, `Step.admin` has no caller. [INFERENCE] A controller that
  the deployer fixes at deployment can still pause by its own logic. No labbook entry
  that I found specifies such a controller.
- Statements whose text mentions it (4): `shield_creates_exactly` (`hNotPaused`),
  `shield_can_happen` (`hNotPaused`), `transact_cannot_be_blocked` and
  `transactOwned_cannot_be_blocked` (`{ s with shieldsPaused := b }`).
- Side effect: both "cannot be blocked" statements prove only that the admin step
  cannot block a spend. They say nothing about the steps of other users.
- Replacement: delete the field, `hNotPaused` and `Step.admin`. Restate the two
  "cannot be blocked" statements over every `Step` by any caller: a step that spends
  none of the same notes keeps `hRoot`, `hRemove` and `hNullify` of a pending valid spend
  true. [INFERENCE] `rootsAt'` only grows and `spent` grows only by the nullifiers of the
  notes that a step spends, so the proof is short.

### P-4. `isZeroPreimage` — the all-zero preimage as "no disclosure"

- Lean: `isZeroPreimage` (`PoolOwned.lean:95-96`):
  `pre.npk = Term.val 0 ∧ pre.value = 0 ∧ pre.token = 0`. `recordStep` skips it
  (`PoolOwned.lean:240-242`). `Step.transactOwned.hNonZero` (`Pool.lean:1436-1437`).
- What it represents: the calldata encoding of an absent `publicOutputs` entry.
- Class: encoding (sentinel for "absent").
- Evidence, model: `PoolOwned.lean:87-94`: "An output that stays ciphertext-only
  carries an all-zero preimage … The real contract compares the raw stored `npk` field
  to the literal `bytes32(0)`".
- Evidence, source: `TransactModule.sol:582-585`:
  `return _preimage.npk == bytes32(0) && _preimage.value == 0 && _preimage.token.tokenAddress == address(0) …`.
  `TransactModule.sol:569`: `return (false, "Missing public output");`.
- Effect: the owned path must disclose every output (`requireAll = true`), so the
  sentinel only adds a hypothesis. The hypothesis forbids an owned output with
  `npk = Term.val 0`, value 0 and token 0. That restriction exists only because of the
  encoding.
- Statements whose text mentions it (4): `transactOwned_cannot_be_blocked`
  (`hNonZero`), `transactOwned_records_all` (`hNonZero`), `owned_transact_can_happen`
  (`hOutputsNonZero : ∀ j, npkOut j ≠ Term.val 0`), `owned_spend_can_happen`
  (`hOutputsNonZero`, and the conclusion `¬ isZeroPreimage`).
- Replacement: a disclosure is `Option CommitmentPreimage`. The plain path takes
  `Fin M → Option CommitmentPreimage`. The owned path takes `Fin M → CommitmentPreimage`
  (all disclosed). `recordStep` has no zero test. `hNonZero`, `hOutputsNonZero` and
  `npk_preimage_not_zero` go away.

### P-5. Position statements added for the adjudicator's `holdings` slot

- Lean: `live_note_unspent` (`PoolSafety.lean:1723`), `live_leaf_matches` (`:1938`),
  `live_values_lt_2pow120` (`:2016`), `leaf_stable` (`:1776`).
- What they represent: facts about one live entry, stated by its raw
  `(treeNumber, position)` key, so that the adjudicator can reason about a stored
  `(Nat × Nat)` position.
- Class: contract layout, inherited from the confirmed `holdings` error.
- Evidence, model: `README.md:78-79` gives the source of `live_note_unspent` as
  "adjudicator payout freshness" and of `live_leaf_matches` as "adjudicator payout
  witness". `PoolSafety.lean:1935-1937`: "Ties an adjudicator's holding — a recorded
  `(treeNumber, position)` and preimage — to the actual on-chain leaf". Commit `d406cea`
  says of `live_note_unspent`: "Requested by the adjudicator proof". Commit `b86cffc`
  says of `live_leaf_matches`: "Ties an adjudicator's holding (a recorded (treeNumber,
  position) and preimage) to the real on-chain leaf".
- Evidence, use: `armada-lean-adjudicator-v4` `Adjudicator/Safety.lean:2459`
  (`live_values_lt_2pow120` bounds `heldValue`), `:2483`
  (`holdings_position_leaf_value_placeholder` uses `live_leaf_matches`), `:2504`
  (`live_note_nullifier_unspent_placeholder` uses `live_note_unspent`). All three read
  `mid.holdings asset cid = some (pos, c)`.
- Note: each fact is true of the pool. The defect is the statement shape. An ordinary
  pool user needs one fact: "I hold a live note, so I can spend it". The four
  statements break that fact into position-level parts that a stored position needs.
- Statements whose text mentions it (4): the four above.
- Replacement: one statement, `live_note_spendable`. For a reachable `s` and a live
  owned note whose opening is public, there is `s'` with `Step.transactOwned s s'` that
  spends it. A plain-path twin covers a key holder. The four facts become lemmas of its
  proof. The adjudicator then holds a note, not a position.

### P-6. `uint120` width in the pool's state and behaviour

- Lean: `Step.shield.hRange : RangeChecked120 (shieldBase feeBP grossAmount)`
  (`Pool.lean:1257`). `LiveValuesBounded` (`PoolSafety.lean:1946-1947`):
  `p.2.value < 2 ^ 120`.
- What it represents: the Solidity type width of `CommitmentPreimage.value`.
- Class: encoding (fixed-width cast in the symbolic layer).
- Evidence, model: `Pool.lean:1246-1248`: "`hRange` is the note value's own field width:
  `CommitmentPreimage.value` is `uint120`". The shield has no circuit, so the
  `Num2Bits(120)` predicate `RangeChecked120` here encodes an ABI width, not a circuit
  constraint.
- Evidence, source: `PoolTypes.sol:82`: `uint120 value;`. The ABI bound applies to the
  gross request value. `_getFee` then casts: `base = uint120(_amount - …)`
  (`ShieldModule.sol:364`). [INFERENCE] The model's bound on `shieldBase` is weaker
  than the contract's bound on the gross value. It does not match the contract either.
- Statements whose text mentions it (3): `shield_creates_exactly`, `shield_can_happen`,
  `live_values_lt_2pow120`.
- Replacement: the protocol meaning is "a shield creates a spendable note". Define
  `Spendable v` once in the circuit layer as the circuit's input predicate (`rangeIn`).
  `Step.shield` requires `Spendable value`. The pool statement becomes "every live note
  is spendable". The number 120 stays in the circuit model.

### P-7. The unshield as "the last output, at most one, recipient in `npk`"

- Lean: `Step.transact`'s `unshieldTail : List CommitmentPreimage`,
  `hSplit : transactAllOutputs w = newNotes ++ unshieldTail` and
  `hUnshieldLen : unshieldTail.length ≤ 1` (`Pool.lean:1353`, `:1358-1359`).
- What it represents: an optional withdrawal of an amount to a recipient.
- Class: encoding (list of length ≤ 1 for `Option`; recipient address packed into
  `npk`; position "last commitment" as a flag).
- Evidence, source: `PoolTypes.sol:78`: "For unshield outputs, `npk` encodes the
  recipient address." `TransactModule.sol:328-331`: "Hash must match last commitment".
  `TransactModule.sol:413`: `address recipient = address(uint160(uint256(_note.npk)));`.
- Effect: the model has no recipient. `held` falls by the amount, but no statement can
  say who receives it. `unshield_can_happen` names the recipient only as
  `npkOut : Term`.
- Statements whose text mentions it (3): `transact_cannot_be_blocked` (`hSplit`,
  `hUnshieldLen`), `transact_unshield_exact` (`unshieldTail : List CommitmentPreimage`),
  `unshield_can_happen` (`npkOut`, `valueOut` as the payout).
- Replacement: `unshield : Option (Term × Nat)` (recipient address term, amount),
  carried in `BoundParams` (P-2). The circuit-output equality that ties it to the last
  commitment is the refinement layer's job. Add a ghost payout record so that
  `transact_unshield_exact` can say "recipient `r` receives exactly `v`".

### P-8. `Term.val 0` as the "absent" default of a list lookup

- Lean: `getD pos (Term.val 0)` in `live_leaf_matches` (`PoolSafety.lean:1940`),
  `leaf_has_merkle_witness` (`:1038`) and `owned_spend_witness_exists` (`hLeafValue`,
  `:1083`). `PoolState.rootsAt`/`leavesAt` default to `⟨[], []⟩` for a tree that does
  not exist (`Pool.lean:710`, `:716`).
- What it represents: "the leaf at this position", with a default value for "no leaf".
- Class: encoding (sentinel for "absent").
- Evidence, model: the same term `Term.val 0` has three meanings. It is the empty-leaf
  value `ZERO_VALUE` (`Merkle.lean:54-57`, `:89`). It is the lookup default above. It is
  the `npk` of the "no disclosure" sentinel (P-4). The real `ZERO_VALUE` is a different
  constant (`PrivacyPoolStorage.sol:150`).
- Effect: [INFERENCE] the statements are sound, because a commitment is a hash term and
  never equals `Term.val 0`. But a reader must know that fact to see that the
  equation implies the position exists.
- Statements whose text mentions it (3): `live_leaf_matches`,
  `leaf_has_merkle_witness`, `owned_spend_witness_exists`.
- Replacement: `(s.leavesAt' tn)[pos]? = some (hashCommitment pre)`. A named constant
  `emptyLeaf : Term` fills unused tree slots. `rootsAt'` returns `Option` or the model
  keeps all trees in one list (P-11).

### P-9. Basis-point fee arithmetic in the shield step

- Lean: `Step.shield (token grossAmount feeBP : Nat)` with `hFeeBP : feeBP ≤ BASIS`,
  and the note value `shieldBase feeBP grossAmount` (`Pool.lean:1254-1259`).
  `BASIS := 10000`, `shieldBase` (`Fee.lean:20`, `:28`).
- What it represents: the value that enters the pool after the fee.
- Class: encoding (rate as integer basis points, floor division) and contract-only rule.
- Evidence, model: `Pool.lean:96-101` says the fee goes from the caller's wallet to the
  treasury and never passes through the pool. So the pool's books see only the net
  value. The core invariant does not depend on the formula.
- Evidence, source: `ShieldModule.sol:353-371` (`_getFee`). The contract has two other
  fee branches (privileged caller, fee module, `ShieldModule.sol:285-324`). The model
  covers only the flat branch.
- Evidence, Armada rule: CON decision 7 sets fees once at deployment. The model lets
  each shield choose `feeBP`.
- Statements whose text mentions it (2): `shield_creates_exactly`, `shield_can_happen`.
- Replacement: `Step.shield (token) (value : Nat) (npk)`, where `value` is the amount
  that enters the pool. The fee rule stays in `Fee.lean` as a separate specification,
  over a deployment constant.

### P-10. `ownedRandomOf` — a new storage mapping for the owned-note random

- Lean: `PoolState.ownedRandomOf : OwnedRandomRegistry` (`Pool.lean:1142`),
  `OwnedRandomRegistry := Term → Option OwnedDisclosure` (`PoolOwned.lean:336`),
  `recordOwnedStep` (`:348-350`).
- What it represents: the public opening `(ownerField, random)` of a contract-owned
  note.
- Class: contract layout (a proposed slot, `mapping(commitment => (owner, random))`).
  It is in neither the contract nor the papers.
- Evidence, model: `PoolOwned.lean:303-313`: "the merged contract's `CommitmentPreimage`
  struct … has exactly three fields … The approved design changes this … records
  `random` … in a second, public registry". Commit `5d8a5f7` calls the change
  "operator-approved". I found no labbook entry that records this approval.
- Evidence, Armada rule: CON:187-189, decision 11: "The adjudicator-owned note's
  preimage is published at creation, in the same transaction, inside `adaptParams`, and
  the adjudicator verifies it against the proof's output commitment". The rule puts the
  opening in call data that the adjudicator checks, not in a pool slot. CON decision 3
  (`:161-164`) says the creator publishes the preimage. It names no storage.
- Statements whose text mentions it (2): `pool_ownedRandomOf_consistent`,
  `owned_spend_witness_exists` (`ownedRandomOf` argument, `hOwned`).
- Replacement: the opening is part of the step's public input (inside `adaptParams`,
  P-2). A ghost relation `Published s c opening` records every opening that a step
  carried. The statement is "every published opening hashes to its commitment". The
  pool needs no slot for it.

### P-11. The tree forest as "closed trees + current leaves + current roots"

- Lean: `PoolState.closedTrees`, `.leaves`, `.roots` (`Pool.lean:1137-1139`).
  `insertTreeNumber`, `insertStartPos`, `insertResult` (`Pool.lean:722-742`), with the
  test `curLeaves.length + count > 2 ^ TreeDepth`.
- What it represents: all Merkle trees and their root histories.
- Class: contract layout.
- Evidence, source: the contract keeps a current-tree pointer and per-tree maps:
  `treeNumber` (`PrivacyPoolStorage.sol:162`), `rootHistory` (`:171`), and
  `nextLeafIndex`/`filledSubTrees`. `MerkleModule.sol:66` holds the rollover test. The
  model's split into "closed" and "current" copies the pointer.
- Statements whose text mentions it (1): `shield_creates_exactly` states the new
  position as `(insertTreeNumber s.closedTrees s.leaves 1, insertStartPos s.leaves 1)`.
- Replacement: `trees : List TreeState`, where the last tree is open. The shield
  statement says the new position is fresh: no earlier leaf has it, and it holds the new
  commitment. The rollover formula stays inside the step definition.

## Table (ordered by the number of statements whose text mentions the element)

| # | Element | File:line | Class | Statements |
|---|---|---|---|---|
| P-1 | `token : Nat` | `PoolOwned.lean:77`, `Pool.lean:1135,1140` | encoding | 10 |
| P-2 | `boundParamsHash : Nat` | `Transact.lean:103`, `TransactOwned.lean:101` | encoding | 5 |
| P-3 | `shieldsPaused`, `Step.admin` | `Pool.lean:1143,1255,1462` | contract-only rule, against CON decision 7 | 4 |
| P-4 | `isZeroPreimage`, `hNonZero` | `PoolOwned.lean:95`, `Pool.lean:1436` | encoding | 4 |
| P-5 | position statements for `holdings` | `PoolSafety.lean:1723,1938,2016,1776` | contract layout (inherited) | 4 |
| P-6 | `RangeChecked120` / `2 ^ 120` in pool | `Pool.lean:1257`, `PoolSafety.lean:1946` | encoding | 3 |
| P-7 | `unshieldTail`, `hSplit`, `hUnshieldLen` | `Pool.lean:1353,1358-1359` | encoding | 3 |
| P-8 | `getD … (Term.val 0)` | `PoolSafety.lean:1038,1083,1940` | encoding | 3 |
| P-9 | `feeBP`, `BASIS`, `shieldBase` in `Step.shield` | `Pool.lean:1254-1259` | encoding, contract-only rule | 2 |
| P-10 | `ownedRandomOf` | `Pool.lean:1142`, `PoolOwned.lean:336` | contract layout, against CON decision 11 | 2 |
| P-11 | `closedTrees` / `leaves` / `roots` split | `Pool.lean:1137-1139,722-742` | contract layout | 1 |

By weight, not by count: P-3, P-5, P-2 and P-7 change what the statements say. P-1 and
P-8 change only types and notation.

## Coverage: elements checked and found legitimate

`PoolState` fields:
- `live` (ghost, `(treeNumber, position)` key): circuit meaning; the nullifier is `Poseidon(key, leafIndex)` and the proof names one tree.
- `spent : Nat → NullifierSet` (per tree): circuit meaning; the nullifier has no tree component (`PrivacyPoolStorage.sol:174`).
- `held`: Armada rule R4, `accounted[token]` (CON:248-252).
- `roots` as a full history: S-RG "the current or a previous Merkle root"; `rootHistory` keeps every root.
- `leaves`: public data (events carry every commitment).
- `preimageOf`: real contract storage (`PrivacyPoolStorage.sol:248`) and CON decision 3; the sentinel inside `recordStep` is P-4.
- `ownedRandomOf`, `shieldsPaused`, `closedTrees`: findings P-10, P-3, P-11.

`Step` constructors:
- `Step.shield`: one note per request is the contract's meaning; the `ShieldRequest[]` batch is correctly not modeled.
- `Step.transact`: one `Transaction` per step; the `Transaction[]` batch is correctly not modeled.
- `Step.transactOwned`: legitimate shape; `hCaller` is CON decision 2.
- `Step.admin`: finding P-3.

Hypotheses of the steps and statements:
- `hRoot`, `hRemove`, `hNullify`, `markSpent`, `removeAll`: legitimate.
- `hN : N ≤ 8`, `hM : M ≤ 4`: the registered circuit shapes; `no_wraparound_minting` needs them (circuit field arithmetic).
- `hRecords` (sublist, optional disclosure): contract meaning (`TransactModule.sol:556-561`).
- `hOwnedRecords`, `hOwnedOutputs`: legitimate check; they depend on P-10 for where the opening lives.
- `hValue : 0 < grossAmount`: contract revert rule (`ShieldModule.sol:244`); protocol behaviour, not encoding.
- `hPathLen : … = 16`: circuit depth; the literal duplicates `TreeDepth`.
- `hRangeIn`/`hRangeOut` in the can-happen statements: circuit witness constraints, legitimate there.
- `hCaller : caller = ownerField` with no `% 2 ^ 160`: operator address ruling; truncation recorded as a departure.

Circuit and crypto layer:
- `Term`, `HashFn`, `Term.keccak` for addresses: operator ruling.
- `npk`, `commitment`, `nullifier`, `masterPublicKey`, `msgHashOf`: circuit formulas.
- `RangeChecked120`, `p`, `fieldEq`, `no_wraparound_minting`, `two_pow_120_lt_p`: circuit arithmetic.
- `merkleRoot`, `treeRoot`, `treePath`: legitimate; the `filledSubTrees` gas algorithm is correctly omitted.
- `TreeDepth = 16` and tree rollover: the circuit fixes the path length, so a tree has a capacity.
- `EdDSAVerify`, `VerifyAccepts`, `VerifyOwnedAccepts`, the completeness records: legitimate abstractions.
- `testingMode`, `emergencyPaused`, `withdrawOnlyMode` not modeled: legitimate (CON R8 and the freeze removal).
- `npk < SNARK_SCALAR_FIELD` checks and the unshield-to-pool guard not modeled: encoding checks, correctly omitted.

Statements with no finding: `core_invariant`, `pool_no_double_spend`,
`transact_spends_only_owned` (except P-2 through `msgHashOf`),
`transactOwned_spends_only_owned`, `transact_root_was_produced`,
`pool_preimageOf_consistent`, `live_positions_distinct_nullifiers`, `owned_owner_binding`,
`owned_no_double_spend`, `preimageOf_consistent`, `transactOwned_value_conserved`,
`commitment_owner_inj`, `user_npk_ne_contract_npk`.

## Other observations (not of this class)

- `insertResult` with an empty batch pushes a root (`Pool.lean:741-742`). The contract
  returns early with no root write (`MerkleModule.sol:57-60`). `transact` inserts only
  when `commitmentsCount > 0` (`TransactModule.sol:83`). The model writes one duplicate
  root. No statement depends on it.
- `unshield_can_happen` takes `hValuePos : 0 < valueOut` and does not use it
  (`PoolSafety.lean:798`, `:835`). `Step.transact` has no such check, but the contract
  has one (`TransactModule.sol:301-303`).
