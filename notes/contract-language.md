# Contract language: Solidity, not Vyper — internal note (not published)

Status: operator decision, 2026-10-18. armada-nitro stays in Solidity. This note records why, and when to
reconsider. It is not an ADR. If the decision needs a clause, promote it to `engineering/09-architecture-decisions.md`.

A claim that was not checked in the 2026-10-18 session has the mark [INFERENCE].

## Context: most new code will be written by machines

Agents (exophial workers) will write most new armada-nitro code. So the question is not "which language does a
person prefer". It is "which language, and which checks, make a wrong program fail before it merges".

Three layers decide that:

| Layer | Choice | What it guards |
|---|---|---|
| Source language | Solidity | How hard it is to write the program wrong |
| Proofs | hevm symbolic tests, at the bytecode level | Properties of what the chain runs, with no trust in the compiler |
| End to end | Fixturenet jobs, with results checked against the chain | The deployed system works as a whole |

The second and third layers check the bytecode. They do not depend on the source language. For machine-authored
code they are the main guard, because they are independent of the author.

## Why Solidity

**The decisive reason is the existing code, not the language.** In armada-nitro (`src/`, 2026-10-18):

- `src/nitro/` is imported Nitro, about 1,700 lines, and E16 edited it in place (allow-list, reentrancy guard and
  `safeTransfer` in `MultiAssetHolder.sol`).
- `src/ArmadaNitroAdjudicator.sol` (180 lines) is `contract ArmadaNitroAdjudicator is NitroAdjudicator`.
- The tree has 8 uses of inline `assembly`.

Vyper has no inheritance and cannot extend a Solidity contract (Vyper 0.4 modules compose Vyper code only). A move to
Vyper means a port of Nitro first. That turns reviewed upstream code into new, unreviewed code, ends the ability to
diff against upstream Nitro, and must reproduce the inline assembly, which Vyper does not have. hevm and the
fixturenet already check the bytecode, so the port buys nothing they do not.

Smaller points for Solidity:
- User-defined value types (`type Amount is uint256;`) give units as types. Vyper has no equivalent.
- The verification tools (Foundry, SMTChecker, hevm, Kontrol, Certora) are built around Solidity first.
- [INFERENCE] Models have seen far more Solidity than Vyper, so machine-written Solidity is likely to be more idiomatic.
  This is not measured.

## Where Vyper is equal or better

- **Analysis.** No modifiers, no inheritance, no recursion, loops always bounded. The code is easier to audit, and
  easier for hevm to explore completely. For machine authors, fewer constructs means fewer ways to be wrong.
- **Construction.** Neither language has sum types with data or an exhaustive `match`. Neither gives compiler-proven
  totality (the construction in `lobaria/docs/process.md`, Strata 1 and 5).
- **Tests.** Foundry tests and scripts are Solidity in both cases. Foundry can compile and test Vyper contracts.
- **Compiler risk.** Both compilers have had serious bugs. Vyper's reentrancy-lock bug (0.2.15–0.3.0) caused the 2023
  Curve exploits. Bytecode-level verification covers this risk for either language.

## Other EVM languages

- **Fe** (fe-lang.org, release 26.4.1 on 2026-10-18) is the only EVM language with enums that carry data and a
  compiler-checked exhaustive `match`. It also has explicit effects and immutability by default. Its maturity is the
  risk: a new compiler backend (Sonatina), and little evidence of audited production use [INFERENCE: not researched].
- **Yul** and **Huff** have no types beyond the 256-bit word. Raw bytecode is the same.

## When to reconsider

- A large new contract that does not extend Nitro. Vyper's restrictions then help machine authors, and nothing has to
  be ported. Weigh that against a second compiler and toolchain in the repository.
- Fe shows an audit and production record. Compiler-proven totality is the strongest construction-level guard for
  machine-authored code on the EVM.
- armada-nitro stops tracking upstream Nitro, or replaces it.

## Related

- `lobaria/docs/process.md`: the construction and verification strata this comparison uses.
- hevm (`argotorg/hevm`) is the chosen symbolic tool: maintained, release 0.58.0 on 2026-06-29. Halmos
  (`a16z/halmos`) had no commit after 2025-08-06.
