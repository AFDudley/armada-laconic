# The concrete specification of the pool: strata 1-5

Date: 2026-11-11. Session `01a195d3`. Process: `docs/process.md` (strata 1-5 for the concrete-specification layer; the
one rule: the whole machine shallow and running before any part is deep). Decisions it builds on:
`labbook/2026-11-11-library-selection.md` (E1, E2, D0-D3). Vocabulary: `glossary.md`, "Verification terms".

Sources read for this entry: `armada-lean/ArmadaLean/Pool.lean` (`Pool.ContractState` :1082, `Pool.Call` :1117,
`Pool.apply` :1549), `armada-lean/docs/pool-executable.md` (PE1-PE7), `armada-lean/docs/verification-chain.md`,
`armada-poc/contracts/privacy-pool/storage/PrivacyPoolStorage.sol` (storage layout).

## 1. The contract monad (the core of the Lean contract library)

A concrete specification is ordinary Lean (E1). Its functions run in one monad:

```lean
structure Env where
  caller : Address
  value  : UInt256
  self   : Address
  ext    : World        -- the external world the call can reach (token ledgers, the verifier)

inductive Result (σ ε α : Type) where
  | ok     (a : α) (s : σ) (logs : List Event)
  | revert (e : ε)

def Contract (σ ε α : Type) := Env → σ → Result σ ε α
```

- `σ` is the contract's storage, a Lean structure whose fields have library storage types (§2).
- `ε` is the contract's custom-error type, one constructor per error. A revert carries no string (selection table).
- External calls read and change `Env.ext` only through library constructs whose behaviour is a named hypothesis
  (`TokenAssumption`, the verifier's soundness and completeness). The external world is modelled around the chain step,
  as Nitro models signers and the network around `Adjudicator.apply` (`docs/process.md` 9.1).
- Combinators: `get`/`set` of a field, mapping read and write, `require (c) (e)`, `emit`, the token-call construct,
  the external-call construct, `for i < n` with `require (n ≤ MAX)` (D2). Each is a construct (glossary).

## 2. Stratum 1: the concrete state as a type

One storage field per fact the contract stores, typed so that an illegal configuration cannot be built. The table maps
each `Pool.ContractState` field (PE1) to its concrete form and to its part of the gluing invariant.

| Specification field | Concrete storage | Gluing invariant (decoder) |
|---|---|---|
| `trees : List TreeState` | `treeNumber : UInt256`; `nextLeafIndex : UInt256`; `filledSubTrees : FixedArray TREE_DEPTH Bytes32`; `merkleRoot : Bytes32`; `rootHistory : Mapping UInt256 (Mapping Bytes32 Bool)` (Armada's layout, `PrivacyPoolStorage.sol:153-171`) | a **relation**, not a decode: the frontier and root are those of the current tree's leaf list, and `rootHistory[t]` holds exactly the roots of tree `t` (the tree relation, open since P4) |
| `spent : Nat → NullifierSet` | `nullifiers : Mapping UInt256 (Mapping Bytes32 Bool)` (`:174`) | `nullifiers[t][n] = true ↔ n ∈ spent t`, through the hash decoder |
| `held : Token → Nat` | **not stored**: custody is each token's balance of the pool | `held tok = balanceOf tok self` in `Env.ext`, under `TokenAssumption` |
| `preimageOf : PreimageRegistry` | `preimageOf : Mapping Bytes32 CommitmentPreimage` (`:248`) | entrywise, through the hash decoder |
| `treasuryFees : Token → Nat` | `treasuryFees : Mapping Address UInt256` (Armada: `cumulativeFees`, `:269`) | entrywise |
| `feeRatio : FeeRatio` | `feeRatio : Packed (UInt128 × UInt128)`, one slot | `num`, `den`, with `FeeRatio`'s three proof fields from the invariant (§3) |
| `feeAdmin : Term` | `immutable feeAdmin : Address` | the address decoder (160-bit truncation collision assumption) |
| `treasury : Term` | `immutable treasury : Address` | the address decoder |

Reasons for the three choices that differ from a one-to-one copy:
- **`held` is not stored.** Armada stores no custody mapping; the custody is the token's own ledger. Storing it would
  cost two extra `SSTORE`s per shield and per payout, and it could disagree with the token's balance. The concrete
  `shield` keeps Armada's balance-before/after check (`ShieldModule.sol:291-294`), which also rejects fee-on-transfer
  tokens.
- **`feeRatio` is one packed slot.** It is read on every shield; one `SLOAD` instead of two. `num ≤ den < 2^128` is
  part of the invariant.
- **`feeAdmin` and `treasury` are immutable.** The deployment fixes them (ruling 2026-11-06, PE notes); an immutable
  costs no `SLOAD`.

Nested mappings stay nested (D1). Every keccak-derived slot carries the named non-collision hypothesis.

## 3. Stratum 2: the invariant before any transition

One invariant of the concrete machine, stated before the steps, with one step lemma per call:

`ConcreteInv s` :=
- the fee ratio is a valid `FeeRatio`: `0 < den`, `num ≤ den`, `num * 100 ≤ den * maxFeePercent`;
- the tree fields are consistent: `merkleRoot` is the root of the frontier `filledSubTrees` at `nextLeafIndex`, and
  `rootHistory[treeNumber][merkleRoot] = true`;
- `nextLeafIndex ≤ 2 ^ TREE_DEPTH`.

`ConcreteInv` replaces the separate hypotheses `hFee`, `FeeInv` and `StorageShaped` of the P4 theorems (Stratum 1's
gate: can the type carry it?). Where it can, the type carries it instead (the fee ratio as a subtype of the packed
pair).

The gluing invariant `Glue s_spec s_conc ext` (9.1) is the table of §2 plus the three collision assumptions. It is
stated before the refinement proof.

## 4. Stratum 3: parse once, at the edge

- Each external function decodes its calldata once, through the library's `AbiType` instances, into typed arguments.
  Nothing downstream decodes again.
- **All input is checked before any effect** (3.2). For `transact`: every element of the batch is decoded and
  validated (the proof, the root in history, the nullifiers fresh, the bounds `N ≤ 8`, `M ≤ 4`, the bound-parameters
  hash) before the first nullifier is marked and before any payout. Armada's spec section 8.1 already pins this order
  (`TransactModule.sol:36-37`).
- The batch is a loop construct over the batch length with `require (n ≤ MAX_BATCH)` at the entry (D2).

## 5. Stratum 4: the machine, flat

The pool has one mode. Its transitions are the four calls of `Pool.Call`: `shield`, `transact`, `transactOwned`,
`setFeeRatio`. Each gets a "can happen" theorem from genesis on the concrete machine, built by real steps (gate 4).

## 6. Stratum 5: totality by machine

The concrete specification's dispatch is a `match` over one inductive `Call` with the four constructors. The step
lemma of `ConcreteInv` and every refinement proof end in that exhaustive `match`, so a new call breaks every proof that
does not handle it (gate 5).

## 7. What the shallow machine (stratum 6) must contain

Before any part goes deep:
1. the library skeleton: the monad, the storage and ABI types, and every construct of §1, each with its Solidity
   form; every construct lemma stated with a `sorry` proof;
2. the concrete specification of all four calls, building;
3. the emitter producing Solidity for all four, compiling with the pinned solc build;
4. the statements, each with a `sorry` proof and building: `ConcreteInv` preserved; `Glue` preserved by each call
   (the refinement to the specification); the equality with `Pool.apply` (E2); each function's two bytecode
   theorems (success; no success on revert);
5. the emitted contract deployed on the fixturenet by the production deploy, and one scenario per call with recorded
   verdict lines.

Acceptance of stratum 6: all of 1-5 hold. Only then does any proof go deep.

## 8. Items that need a value or a ruling

- `MAX_BATCH`: a size bound on an approved statement; it goes to the operator with its gas measurement (9.5).
- `held` not stored changes where custody lives in the gluing invariant, not what the specification says. Operator
  review of §2.
