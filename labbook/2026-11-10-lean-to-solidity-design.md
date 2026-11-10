# From the Lean pool model to proven Solidity (2026-11-10)

This entry replaces the uncommitted draft `2026-11-08-lean-to-solidity-on-jaune-blanc.md` (session `01a1bd34`, ended).
The draft assumed a Stele layer, which the operator dropped on 2026-11-09. Its operator statements O1-O8 and its
checked facts are kept here. Owner: session `01a195d3` (armada-laconic coordinator).

Each rule cites an operator statement (§2) or a measurement (§3). A rule that does neither starts with **Proposal:**.
A claim that was not observed has the mark `[INFERENCE]`.

## 1. Goal

The deliverable is Solidity source for the pool, readable and efficient, together with a machine-checked proof that
its compiled bytecode refines the Lean model `Pool.apply`. Agents do not write per-function proofs. A program
produces them from a fixed lemma library. An agent writes a lemma only when the program meets a construct that the
library does not cover.

## 2. Operator statements

From session `01a1bd34` (2026-11-08):
- **O1.** The deliverable is Solidity, compiled by a pinned compiler.
- **O2.** Jaune and Blanc replace EVMLean and EquiVM where they apply.
- **O3.** The step to Solidity is reversible by construction.
- **O5.** The Solidity is efficient and readable for its solc vintage.
- **O6.** The Lean model follows best practices for readability and modularity, because its shape becomes the
  Solidity's shape.
- **O7.** The pool's first source is the Railgun whitepaper, "which isn't great".
- **O8.** viaIR "should be fine". The pool statement audit "seems pretty straight forward".

From this session (2026-11-09 and 2026-11-10):
- **O9.** "we don't need stele. but we need solidity along with its bytecode proven."
- **O10.** "the whole point is that agents shouldn't be reasoning about any of this in the first place."
- **O11.** viaIR is the default.
- **O12.** Target the current Ethereum mainnet. armada-poc moves to solc 0.8.31, `evmVersion` `osaka` (Fusaka).

O4 of the draft (`SolSem v := Stele ∘ parse`) is withdrawn by O9.

## 3. Measured facts

- **Blanc proves bytecode against `Pool.apply` directly** (`.scratch/jaune-eval-v7-RESULTS.md`): `setFeeRatio` (V5)
  and the accounting part of `shield` (V7), with existence of the successful run, exact gas for the contract's own
  cost, and Lean's three axioms only.
- **Hand proofs are slow.** V7's Blanc track took 11 h 39 min of elapsed time for one function. Most of it was solc
  idioms, gas bookkeeping and statement forms found by trial, none of it specific to the function.
- **Blanc certifies viaIR output** (M0, `.scratch/m0/`): all six compiles of the two fixtures, solc 0.8.31, `osaka`,
  optimizer 200 runs, viaIR on and off, pass `lift.py` and the kernel's `cert_check` and `jumps_ok`. No computed jump.
- **viaIR costs shapes, not correctness** (M0b, `.scratch/m0/m0b/`). Under viaIR a construct is inlined when used
  once and shared when used more often. The inlined and called forms are the same instructions apart from the call
  wrapper (stack moves only) for checked `+` and `*`, `m[k] += v`, `require`, and the external call's sub-pieces.
  Two forms differ in content, and the source predicts both: `a - k` with a literal `k` compiles to `add` of
  `not(k-1)`; `a / k` with a literal `k` has no zero check. `require` shares its revert tail only between sites with
  the same message string.
- **No tool emits Solidity from a model.** EquiVM has only `sol2stele.py` and `vy2stele.py` (source to Stele).
- **Pinned toolchain:** solc 0.8.31, `evmVersion` `osaka`, optimizer 200 runs, viaIR (O11, O12). armada-poc builds
  and tests the same on 0.8.31 as on 0.8.17 (`e6d49868`: forge 515/515).

## 4. Design

The chain has one bridge, a small Solidity syntax tree in Lean with its own executable semantics. Every part of the
chain is keyed to the constructs of that tree.

| # | Part | Artifact |
|---|---|---|
| P1 | **Syntax** | `SolAst`: the Solidity subset the printer emits. Its constructs are exactly the inventory of `Pool.apply`'s needs (§5). |
| P2 | **Semantics** | `SolAst.eval`: a Lean function from a call and a contract state to `Option` state, written per construct, total, with no EVM detail. |
| P3 | **Printer and parser** | `print : SolAst → String`, `parse : String → Option SolAst`, and one theorem `parse_print`. This is O3 without Stele. |
| P4 | **Model link** | For each pool function, a `SolAst` term and a theorem that `eval` of it equals `Pool.apply` on the decoded state (a relation for the tree, §5). |
| P5 | **Pattern catalogue** | For each construct, its bytecode forms under the pinned settings, located through solc's source maps (M0b gives the forms). |
| P6 | **Lemma library** | For each construct form, one Blanc lemma: the bytecode segment refines `eval` of the construct. Forward with exact gas (`RunExact`), `RunEx` for external calls, inversion for reverts. Plus one generic call-and-return lemma. |
| P7 | **Generator** | From a `SolAst` term, solc's source map and Blanc's certificate: the composed theorem that the bytecode refines `eval` of the term. Composed with P4: the bytecode refines `Pool.apply`. |
| P8 | **Assumptions** | The token and the trusted verifier, each stated once as a named structure, as `TransactProofSoundness` is. Keccak slot non-collision as named hypotheses. |
| P9 | **Gate** | Per contract: the theorems build, 0 `sorry`, and `#print axioms` shows Lean's three axioms plus the declared hypotheses only. |

Why P2 sits in the middle: P6's lemmas then talk about `eval` of one construct, not about `Pool.apply`. So the library
is reusable across contracts, and only P4 depends on the pool. P4 is a statement about two Lean functions, so it
needs no EVM reasoning.

**Measured (P4, 2026-11-10):** P2 copies the model's structure one to one, and P4 closes by unfolding plus arithmetic.
The two P4 theorems for `setFeeRatio` and the shield accounting carry three kinds of hypothesis: the token assumption
(P8), `uint256` overflow bounds (the contract is `uint256`, the model `Nat`), and storage shape plus `FeeRatio`
invariants, which a reachability proof over the contract's own calls should discharge (§7, item 4).

**Rule (2026-11-10):** every constant in a printed contract comes from a model term, never a typed-in number. The first
P4 proof passed with a 1% cap in the example contracts while the model's cap was 10% (`faf809c3`); the theorem hid it
behind a hypothesis. The cap is now `ArmadaLean.maxFeePercent` in the `SolAst` term.

**Status (2026-11-10):** P1-P3 merged in evm-proof-toolchain as `81ed0b04` (`parse_print`, Lean's three axioms; the
printed fixtures compiled byte-identical to M0's; `eval` agreed with anvil). P4 merged as `431a0f81` and `4d34cfd8`
(both theorems, Lean's three axioms; anvil 37/37).

**Proposal:** the printer emits an exact pragma `pragma solidity 0.8.31;`, not a caret range, so that the source names
its compiler.

**Proposal:** the printer writes NatSpec from the model's docstrings, and each function cites its Lean equation.

## 5. The Solidity subset

The inventory of `Pool.apply`'s constructs (session `01a195d3`, agent `PoolConstructInventory`, 2026-11-10) gives the
subset. Beyond the constructs of `setFeeRatio` and the shield accounting fixture, the four operations need:

- **storage:** nested mappings (the nullifier set and the root history per tree), a mapping from `bytes32` to a struct
  (the preimage registry), and the incremental Merkle tree state;
- **calldata:** structs and bounded dynamic arrays (public inputs with up to 8 nullifiers and 4 commitments, bound
  parameters, disclosed records, published openings);
- **control:** loops bounded by those sizes and by the tree depth;
- **hashing:** Poseidon (as a call to the deployed Poseidon contract) and `keccak256(abi.encode(...))`;
- **external calls:** `transferFrom` in, `transfer` out, and two `staticcall`s to the trusted verifier.

No events are needed by the model. The existing contracts emit events for wallet sync. Whether the new pool emits
them is a client question, not a model one, and is open.

**The tree is a relation, not a decode.** The model keeps each tree's leaf list (`TreeState.leaves`). A contract
keeps only the incremental state. A function from storage to the model's state cannot rebuild the leaves, so P4 for
the tree is a refinement relation: the incremental state is consistent with the model's leaves. Its core lemma is
that incremental insertion computes `treeRoot` of the appended leaves. The other fields keep a decode function.

The first cut of P1-P3 covers only `setFeeRatio` and the shield accounting fixture (agent `SolAstP1P3`), with types
designed to extend to the rest.

## 6. Order

1. **Inputs, in parallel:** the construct inventory (§5); armada-lean on Lean v4.35.0 and Mathlib v4.35.0, so that one
   Lake build imports the model and Blanc.
2. **P1 + P2 + P3** for the subset.
3. **P4** for `setFeeRatio`, then `shield`.
4. **P5 + P6** for the constructs those two functions use, starting from V5's and V7's proved segments and
   `BlancSolc` (1,299 lines).
5. **P7**, then acceptance: `setFeeRatio` and `shield` re-proved with zero hand-written proof lines.
6. `transact` and `transactOwned` through the same chain. Each agent hour they need is a lemma-library miss, recorded
   by the gate.

**Where the code lives.** **Proposal:** a new Lake project `armada-sol/` in evm-proof-toolchain, on Lean v4.35.0,
requiring armada-lean by git and Blanc by path. `armada-stele/` stays until P7's acceptance, then is removed.

## 7. Open

1. The construct inventory (§5).
2. ~~armada-lean on Lean v4.35.0~~: merged 2026-11-10 as `2de42407` (tip `d960899c`). The 3 `grind` failures were
   stale replay anchors (`grind only [#3a2f, …]`), recorded under v4.33.1; plain `grind only` closes both goals.
3. The pool statement audit (O8) is not started. It comes before `transact`.
4. V5's `hFee` and `hAdminShaped` need a reachability proof over the contract's own calls. P4 inherits them.

## 8. Where this entry is cited

Not yet cited. `engineering/build-plan.md` S5 and `docs/process.md` "The layers" cite it when they change.
