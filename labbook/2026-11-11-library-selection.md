# The Lean contract library: evidence and selection

Date: 2026-11-11. Session `01a195d3`. Inputs (all in `.scratch/`):
- `library-inputs/armada-inventory.md`: the constructs the Armada Solidity uses, with counts.
- `library-inputs/solidity-practice.md`: Solidity best practice, with sources.
- `library-inputs/proof-evidence.md`: the proof cost of each construct, from Blanc's proofs of Dai, USDC, USDT and Curve and from our own `gen/TIMES.md`.
- `survey/2026-11-11-lean-to-evm-survey.md`: 17 existing projects (Lean, Coq, Isabelle, K, our own).
Related: `labbook/2026-11-11-symbolic-execution-and-solc.md` (option A: print Solidity, run solc, prove the bytecode; option B ruled out).

## Goal
A Lean library. A program written against it compiles to Solidity, and its compiled bytecode is proved to refine the program. The library's constructs are the input language.

## Selection table

| Construct | Armada | Practice | Proof evidence | Selection |
|---|---|---|---|---|
| Word type | `uint256` 947+ | `uint256` default | — | `UInt256`, checked arithmetic by default; overflow is an explicit revert condition (DeepSEA's representation relation: a `Nat` in the spec, a bounded word in the code) |
| Packed small uints | `uint120` 32 | only for packing | packing adds slot arithmetic | explicit packed-field construct, not the default |
| `address`, `bytes32`, enums | 856 / 361 / 2 | `address` vs `address payable` is a compile-time type | — | include; payable as a Lean-side type only |
| Structs | 34 | — | — | include, as storage records |
| Mappings | 45 nested (3 in pool core) | — | nested mappings cost most in per-contract proofs | single-level and nested mappings, each one library construct (decision D1) |
| Fixed arrays | `bytes32[TREE_DEPTH]` 2 | — | fixed bounds avoid loop invariants | include, length fixed at compile time |
| Dynamic arrays in storage | — | bounded only | need loop machinery | exclude from the first cut |
| Revert | 397 strings, 15 custom errors | custom errors | a string literal moves the free pointer for the rest of the function | custom errors only |
| Loops | 43 (tree levels: constant; batch: argument) | bounded | compile-time bounds avoid invariants; no surveyed project handles runtime bounds generically | one loop construct: a runtime count with a checked constant maximum, proved once with a fold invariant (decision D2) |
| Token calls | SafeERC20 everywhere, plus a balance-before/after check | SafeERC20 (USDT returns no bool) | one callee spec per return shape | SafeERC20 semantics: success = call ok and return data empty or `true`; named assumption on the balance effect. Replaces the current exact `TokenAssumption` (return data = 1) |
| Other external calls | SNARK verifier, fee module, CCTP | checks-effects-interactions | named assumptions compose (Isabelle/Solidity, our P8) | one external-call construct, after all effects, with a named assumption per callee |
| Events | 113 | one per state change | emit from one program point | include |
| Reentrancy guard | 8 files | full-word sentinel | — | include, as a library construct |
| Access control | owner, 3 pause gates, `feeAdmin` | first-class, hard to skip | — | include (owner and role checks as preconditions) |
| Hashing | keccak 29; Poseidon pervasive | — | keccak is native; Poseidon is not | keccak native; Poseidon a trusted named primitive (decision D3) |
| `delegatecall`, proxies | 1 site (router) | — | about doubles proof size (USDT) | exclude |
| `assembly` | 4 | — | — | exclude |
| `unchecked` | 0 | non-default | — | emitted only where a proved bound makes the check redundant; never written by a program |
| Dispatcher | solc's | — | generate from the library's function list | generated |
| Compiler build | — | — | every bytecode lemma is tied to one build | one pinned solc build; a version change re-proves the helper specs |

## Architecture from the survey
- **Verity** (`github.com/lfglabs-dev/verity`, MIT, active; paper `veritylabs.dev/papers/verity.pdf`) is the closest existing project: a Lean 4 embedded language in which one macro produces both an executable shallow model and a deep-embedded IR from one definition, with a generic correctness theorem for its backend over a stated fragment. Limits: it emits **Yul**, not Solidity; no gas; loops, events and external calls are outside its generic theorem.
- **Its dual-embedding pattern is the template**: one Lean definition gives the specification and the `SolAst` term, so the per-program source proof (our P4) becomes generic or disappears.
- **The three open problems no surveyed project solves**: exact gas against real solc bytecode, generically (our Blanc walks do it per contract); loops with runtime bounds, generically; storage related to a richer model (the Merkle tree) and named verifier assumptions, as reusable constructs.

## Decisions (2026-11-11)
- **E1 (operator): shallow first.** A concrete specification is ordinary Lean in a contract monad. The emitter reflects it into `SolAst`. The program is its own meaning, so the source theorem is composed from construct lemmas.
- **D0: our own library, with Verity's dual-embedding pattern; not built on Verity.** Verity emits Yul, and no tool turns Yul into Solidity at all: solc goes only from Solidity to Yul, decompilers produce non-compiling pseudocode from bytecode, and Verity has only a Solidity importer (`.scratch/survey/2026-11-11-yul-to-solidity.md`). The operator requires Solidity.
- **E2: the library program replaces `Pool.apply`, by the repair pass** (`docs/process.md` R.2-R.5). Build the library program next to `Pool.apply`; prove the two equal through the gluing invariant (9.1); carry PE4, PE5 and PE7 to the library program through that equality; then delete `Pool.apply`. `Pool.apply` was made for the Stele translator, and that reason is gone.
- **D1: nested mappings stay, as one library construct.** They are the idiomatic Solidity form and Armada's layout. The proof cost that `proof-evidence.md` reports is per contract in Blanc's proofs; in the library it is paid once, in the construct's lemma. Keccak non-collision for the slots is a named hypothesis, as for single mappings.
- **D2: one loop construct over a runtime count with a checked constant maximum** (`require(n <= MAX)` at the entry), proved once with a fold invariant. Compile-time bounds are the case `n = MAX`. The value of each `MAX` is a size bound on an approved statement, so it goes to the operator with its gas measurement (`docs/process.md` 9.5).
- **D3: Poseidon is a trusted named primitive** (ruling 2026-10-28, `armada-lean/docs/verification-chain.md`). Its Solidity implementation is tied to it by a named hypothesis, not proved.

Note (2026-11-11, later): the "Compiler build" row of the selection table ("one pinned solc build; a version change re-proves the helper specs") is superseded. The operator ruled that the compiler must be swappable by one command, to other solc versions and to Solar. Each compiler profile has its own generated pattern catalogue and helper specifications. See `labbook/2026-11-11-programmatic-compiler-swap.md`.
