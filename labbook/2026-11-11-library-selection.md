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
| Mappings | 45 nested (3 in pool core) | — | nested mappings cost most; flatten | single-level mapping; a two-key map is one mapping over a packed key. **Open decision D1** |
| Fixed arrays | `bytes32[TREE_DEPTH]` 2 | — | fixed bounds avoid loop invariants | include, length fixed at compile time |
| Dynamic arrays in storage | — | bounded only | need loop machinery | exclude from the first cut |
| Revert | 397 strings, 15 custom errors | custom errors | a string literal moves the free pointer for the rest of the function | custom errors only |
| Loops | 43 (tree levels: constant; batch: argument) | bounded | compile-time bounds avoid invariants; no surveyed project handles runtime bounds generically | compile-time bound; a batch loop takes a checked maximum. **Open decision D2** |
| Token calls | SafeERC20 everywhere, plus a balance-before/after check | SafeERC20 (USDT returns no bool) | one callee spec per return shape | SafeERC20 semantics: success = call ok and return data empty or `true`; named assumption on the balance effect. Replaces the current exact `TokenAssumption` (return data = 1) |
| Other external calls | SNARK verifier, fee module, CCTP | checks-effects-interactions | named assumptions compose (Isabelle/Solidity, our P8) | one external-call construct, after all effects, with a named assumption per callee |
| Events | 113 | one per state change | emit from one program point | include |
| Reentrancy guard | 8 files | full-word sentinel | — | include, as a library construct |
| Access control | owner, 3 pause gates, `feeAdmin` | first-class, hard to skip | — | include (owner and role checks as preconditions) |
| Hashing | keccak 29; Poseidon pervasive | — | keccak is native; Poseidon is not | keccak native; Poseidon as a named primitive with its Solidity implementation proved or assumed. **Open decision D3** |
| `delegatecall`, proxies | 1 site (router) | — | about doubles proof size (USDT) | exclude |
| `assembly` | 4 | — | — | exclude |
| `unchecked` | 0 | non-default | — | exclude from the first cut |
| Dispatcher | solc's | — | generate from the library's function list | generated |
| Compiler build | — | — | every bytecode lemma is tied to one build | one pinned solc build; a version change re-proves the helper specs |

## Architecture from the survey
- **Verity** (`github.com/lfglabs-dev/verity`, MIT, active; paper `veritylabs.dev/papers/verity.pdf`) is the closest existing project: a Lean 4 embedded language in which one macro produces both an executable shallow model and a deep-embedded IR from one definition, with a generic correctness theorem for its backend over a stated fragment. Limits: it emits **Yul**, not Solidity; no gas; loops, events and external calls are outside its generic theorem.
- **Its dual-embedding pattern is the template**: one Lean definition gives the specification and the `SolAst` term, so the per-program source proof (our P4) becomes generic or disappears.
- **The three open problems no surveyed project solves**: exact gas against real solc bytecode, generically (our Blanc walks do it per contract); loops with runtime bounds, generically; storage related to a richer model (the Merkle tree) and named verifier assumptions, as reusable constructs.

## Open decisions for the operator
- **D0 (settled, not open): our own library, using Verity's dual-embedding pattern. Do not build on Verity.** Verity fails three of this project's requirements: (1) it emits Yul, and the operator requires Solidity (2026-11-08); (2) it has no gas model, and our bytecode proofs carry exact gas; (3) its generic theorem excludes loops, events and external calls, and the pool needs all three, so for the pool most of its guarantee would be per-contract bridge facts anyway. What it does have that we lack is one idea: a single Lean definition yields both the executable model and the compiler's input term. That removes our separate translator and the per-program source proof (P4). We take that pattern, not the project.
- **D1. Two-key maps as one packed-key mapping.** It changes the storage layout from Armada's. That is free for a generated contract, but it breaks layout compatibility with the existing deployment.
- **D2. Batch loops**: a fixed maximum batch size (proof stays loop-free per batch element), or a runtime bound with a loop invariant.
- **D3. Poseidon**: prove the Solidity Poseidon implementation against a Lean specification, or state it as a named assumption.
