# The pool for armada-nitro: the owned path, the tree, and the calls Nitro uses

Date: 2026-11-24. This entry continues `2026-11-11-concrete-specification-design.md`. That entry's stratum 6 holds:
the shallow pool `PoolSol` (armada-laconic-mono `armada-sol/ArmadaSol/PoolContract.lean`) is deployed on the
fixturenet, and each call has a can-happen theorem and a scenario. This entry adds what armada-nitro needs, in the
order of `docs/process.md`: vocabulary, types, invariant, edge, machine, totality, then the running system.

Sources, read in full for the parts cited:
- the model, armada-lean `3700113`: `Pool.lean` (`Step`, `genesis`, `insertLeaves`), `PoolOwned.lean`
  (`recordStep`, `markSpent`, `OwnedOpening`), `TransactOwned.lean` (`TransactOwnedPublic`, `OwnedConstrained`),
  `BoundParams.lean` (`BoundParams`, `AdaptAllows`), `Merkle.lean` (`treeRoot`);
- the owned circuit, ship-armada/armada-circuits PR #7 (open), for the order of the public signals;
- armada-nitro `src/armada/IArmadaPool.sol` and the calls that `ArmadaNitroAdjudicator.sol` and `PoolSpend.sol` make:
  `transact`, `transactOwned`, `preimageOf`, `merkleRoot`, `treeNumber`.

ship-armada/armada-poc PR #653 (open) is a check, not a source (process.md 9.2): where it differs from the model, the
model decides.

## 0. Vocabulary (stratum 0)

New terms, added to `glossary.md`: **frontier**, **zero value**, **owned public outputs**, **disclosure**, **tree
rollover**.

## 1. Defects found in the shallow pool

Each one is a place where the concrete machine allows more than the model, or cannot reach a state the model reaches.

| # | Defect | Model | Effect |
|---|---|---|---|
| D1 | The constructor sets no fee ratio. `ConcreteInv` is false at genesis, and `shield` reverts `invalidFeeRatio` until `setFeeRatio` runs. | `genesis r feeAdmin treasury` takes the ratio (`Pool.lean:1147`) | the invariant has no base case |
| D2 | The constructor writes no root. `rootHistory` is empty at genesis. | `genesisTrees` holds the empty-tree root (`Pool.lean:1141`) | no proof against the empty tree is accepted |
| D3 | `transact` checks every transaction of a batch against the state before the batch. Two transactions with one nullifier both pass. | each transaction is one `Step`; `markSpent` refuses a nullifier spent by the step before (`PoolOwned.lean:202`) | a double spend in one batch |
| D4 | A nullifier that occurs twice in one transaction passes. | `markSpent` refuses it within one list | a double spend in one transaction |
| D5 | `shield` writes no preimage. | the shield step records the note (`recordStep`, `Pool.lean`, the `shield` case) | `preimageOf` of a shielded note is absent |
| D6 | `verifyOwned` gives `N+M+3` public signals. The owned circuit gives `2N+M+4`: `npk[N]`, `transcriptHash`, then the inputs. | `TransactOwned.lean:150-156`: the two outputs are public signals that the witness determines | every real owned proof fails |
| D7 | `transactOwned` writes no preimage. | every output is disclosed (`ownedOutputRecords`, `requireAll = true`) | Nitro's `preimageOf` reads fail |
| D8 | `insertLeaf` sets the root to `hash(root, leaf)`. | `TreeState.extend` takes `treeRoot TreeDepth leaves` | no real proof names a root of our tree |

## 2. The state as types (stratum 1)

**2.1 The zero values are constants, not storage.** Armada stores `zeros[16]` and `newTreeRoot`, which its
`initializeMerkle` computes. Here they are 17 contract constants, so no step can change them and no invariant is
needed for them. The one fact about them, `zeros[i+1] = PoseidonT3(zeros[i], zeros[i])` and `newTreeRoot` the root of
the empty tree, is checked once, by the `armada_sol` reconciler verify against the deployed Poseidon, and enters the
proofs as the named hypothesis `ZerosAgree` on the world (process.md 2.4). The values came from Armada's fixturenet pool
`zeros(i)` on 2026-11-23.

**2.2 The owned public outputs are calldata.** `Transaction` gains `ownedNpk : List Bytes32` and
`ownedTranscriptHash : Bytes32`, the ABI of PR #653. The model has no field for them (`TransactOwned.lean:153`: they
are functions of the witness), so `Glue` relates them to `npk pub.ownerField (w.randomIn i)`.

**2.3 The parsed call is a typed value, not the ABI's tagged fields.** The ABI carries `unshield` (an enum) beside
`unshieldPreimage`, and `adaptContract` with zero meaning none. The model has `unshield : Option Unshield` and
`adaptContract : Option Term` (`BoundParams.lean:52-56`). The concrete specification parses each `Transaction` once into
that form (section 4). This needs an `Option` form in Charter: a decision for stratum 1 of Charter itself (D-b below).

**2.4 Absence is named.** A registry entry with every field zero is no entry (`Decoders.preimage`, already so). A note
of value zero is never recorded (process.md 1.3): `shield` and the records refuse `value = 0`.

## 3. The invariant, before any transition (stratum 2)

`ConcreteInv` becomes, over the world's hash functions `H` (Poseidon, from `World.library`):
1. the fee ratio is valid and below the cap (unchanged);
2. `nextLeafIndex ≤ 2^16`, and `merkleRoot ∈ rootHistory[treeNumber]` (unchanged);
3. **registry consistency**: every nonzero entry `preimageOf[k] = p` has `k = H.poseidonT4(p.npk, p.token, p.value)`
   and `p.value > 0`.

The tree's shape (the frontier and the root are those of the current tree's leaf list) is not a concrete fact: it
needs the leaf list, which only the model has. It is part of `Glue` (design record §2, "a relation, not a decode"):
`TreeRel` states that, for the current tree, the model's leaves have length `nextLeafIndex`, `merkleRoot` decodes to
`treeRoot TreeDepth leaves`, and `filledSubTrees[l]` decodes to the frontier node of level `l`.

New statements, each with a `sorry` proof at first: `init_establishes` (D1, D2), the four `*_preserves` (now with
conjunct 3), and `TreeRel` preserved by `insertLeaves`.

## 4. Parse once, at the edge (stratum 3)

Each entry point parses the whole batch into typed values and checks it before any effect (process.md 3.2):
- per transaction: the shape (`N ≤ 8`, `M ≤ 4`; owned `N = 1`, `M ≤ 3`), the owned outputs' length (`ownedNpk.length
  = N` for owned, `0` for plain), the root, the adapt contract (`AdaptAllows`), the gas price and the chain id
  (concrete checks the model leaves to the world; allowed by 9.2), the unshield (last output commitment equals the
  hash of `unshieldPreimage`; never for owned), the disclosed preimages (each hashes to its commitment; all of them for
  owned), the owner (owned);
- over the batch: every nullifier is unspent and occurs once in the whole batch (D3, D4).

Only then do the effects run: mark spent, insert leaves, record preimages, pay the unshield.

## 5. The machine (strata 4 and 5)

`Call` keeps its four constructors, as the model's `Pool.Call`. Each call keeps its can-happen theorem from genesis
(armada-laconic-mono `ArmadaSol/PoolCanHappen.lean`); the theorems change with the calls. G5 (a fifth constructor
breaks every proof that does not handle it) is checked once, on a scratch change.

## 6. Running (stratum 6)

Each item ends with the emitted `out/PoolSol.sol` deployed by the `armada_sol` reconciler module, `provenance.json`
written again, and the scenario of each changed call run. A `transactOwned` scenario needs owned keys: PR #7 has the
sources and no release (D-a below).

## 7. The items, in order

| # | Item | Fixes | Acceptance |
|---|---|---|---|
| P1 | genesis: constructor takes the fee ratio and writes the empty-tree root; `init_establishes` stated | D1, D2 | can-happen of `shield` with no `setFeeRatio` before it |
| P2 | the tree: 17 constants, the frontier walk with rollover; Charter forms for a range loop, an indexed storage array and a select; `TreeRel` stated | D8 | fixturenet `transact` PASS with a real proof against PoolSol's own tree |
| P3 | the edge: one parse of the batch; duplicate nullifiers refused | D3, D4 | can-happen theorems; a revert theorem for a duplicate |
| P4 | the owned public outputs | D6 | fixturenet `transactOwned` PASS (needs D-a) |
| P5 | the registry: `shield` records; disclosure for `transact`; all outputs for `transactOwned`; getters `preimageOf`, `merkleRoot`, `treeNumber` | D5, D7 | Nitro's calls resolve against PoolSol |
| P6 | the unshield payout and the adapt check | — | Nitro's `depositFromNote` scenario on PoolSol |
| P7 | Armada's events and `shield(ShieldRequest[], integrator)`, so the SDK and Nitro's gates run on PoolSol | — | armada-nitro's gates on PoolSol |

P1 to P5 make a pool that supports `transactOwned` as Nitro uses it. P6 and P7 make Nitro's gates run on it.

## 8. Decisions for the operator

- **D-a.** Owned keys: a local dev key setup from PR #7, for the fixturenet only, or wait for a release.
- **D-b.** `Option` in Charter, so the edge returns the model's typed form (2.3). The alternative keeps the ABI's
  tagged fields in the concrete specification, which AGENTS.md forbids ("no tagged-untyped payloads").
- **D-c.** The ABI of `Transaction` follows PR #653 before it merges (2.2). armada-nitro's `IArmadaPool.sol` then
  changes in the same step.
- **D-d.** `MAX_BATCH` (design record §8). The fixturenet uses 8.
