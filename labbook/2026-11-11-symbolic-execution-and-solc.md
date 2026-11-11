# Symbolic execution and the Solidity compiler: what each does in the pipeline

Date: 2026-11-11. Session `01a195d3`. Related: `labbook/2026-11-10-lean-to-solidity-design.md`.

## The pipeline, and what each step is trusted for

| Step | What runs | What it produces | Trusted? |
|---|---|---|---|
| 1. Lean program | the library-based Lean definition | the specification | yes: it is the specification |
| 2. Translator | a Lean program (`ArmadaSol/Translate/`) | a `SolAst` term | no: checked by step 3 |
| 3. Source proof (P4) | Lean's kernel | `eval (term) = Lean program` | yes: the kernel |
| 4. Printer | `ArmadaSol.print` | Solidity text | proved: `parse (print c) = c` |
| 5. **Solidity compiler** | `solc 0.8.31 --via-ir --optimize --optimize-runs 200 --evm-version osaka` | runtime bytecode | **no**: checked by steps 6-7 |
| 6. Lift | Blanc's `lift.py` | a control-flow certificate | no: checked by `Cert.check`, decided by the kernel |
| 7. **Symbolic execution** | `xrun`/`xcallM` walks in Lean | a theorem: the bytecode refines the Lean program, for all inputs | yes: the kernel |

## Running solc is not a proof step

solc is an ordinary, unverified program. We run it, as any project does, to obtain the bytecode that is deployed. Nothing in the result depends on solc being correct. If solc had a bug, the symbolic execution of its output would fail to prove the theorem, or would prove a theorem about the bytecode that solc actually produced. The theorem is about **that exact bytecode**, not about solc in general.

Consequences:
- The proof is per build. A change of solc version, optimizer setting or EVM version gives different bytecode, so the bytecode proof must run again.
- The bytecode shape is solc's choice (viaIR interleaves source statements, shares tails, inserts overflow guards and ABI helper routines). Symbolic execution must handle whatever solc emits. That is why we need a library of **callee specifications for solc's own helper routines** (ABI decode, checked arithmetic, panic, memory allocation), proved once per solc version and reused by every contract.

## Symbolic execution is the proof step

Concrete execution (anvil, `cast call`) runs the bytecode on one input. It is a test. We ran 37 such cases; they are evidence, not proof.

Symbolic execution runs the bytecode on **symbolic** inputs (variables, not numbers). Each instruction produces an expression; each conditional jump splits the run into two paths, each with a condition; the gas is a symbolic sum. At the end, each successful path's final state is compared with the Lean program's result. In Lean, Blanc's `xs` tactic does one instruction by applying that instruction's step lemma; `xrun` repeats it; `xcallM` crosses an internal call with the callee's specification. The kernel checks each step, so the result is a theorem for all inputs that satisfy the hypotheses.

What symbolic execution needs from us:
- **Branch facts**: the condition of each branch, as hypotheses derived from the Lean program's preconditions (the bridge lemmas: `require`, checked arithmetic, selector).
- **Callee specifications** at real internal calls, each generic in its context (any stack below its arguments, any memory that satisfies its invariant). Intermediate states inside a straight-line path are never written by hand; `xrun` computes them.
- **Named assumptions** at external calls (`TokenAssumption`): the callee is not our code, so its behaviour is a hypothesis of the theorem.
- **Loop invariants** for loops (not yet needed; Blanc's Vyper/Curve walks have loop heads).

## Options for the compiler step, and the choice

| Option | Deliverable | Bytecode proof | Cost |
|---|---|---|---|
| A. Print Solidity, run solc, prove the bytecode (current) | readable, auditable Solidity, deployable with standard tools | per build, by symbolic execution | a callee-spec library per solc version |
| B. Compile the library directly to EVM bytecode with our own compiler | no Solidity | the compiler could be proved correct once | a full verified compiler; no Solidity to audit |
| C. Compile to Yul, run solc's Yul optimizer | Yul, not Solidity | per build | similar to A, less readable output |

The operator's requirement is Solidity together with proven bytecode (2026-11-08: "we need solidity along with its bytecode proven"). So option A stays. Its cost, the callee-spec library for solc's helper routines, is also reusable across contracts, which suits a library design.

Operator ruling (2026-11-11): option B is off the table for this project. A direct-to-bytecode compiler would be a fork of a different project, not part of this work. This project's deliverable stays Solidity, compiled by solc, with its bytecode proved by symbolic execution.

Note (2026-11-11, later): the consequences above name one solc build and helper specifications "proved once per solc version". The operator ruled that the compiler must be swappable by one command, to other solc versions and to Solar. The pipeline is now organized by compiler profile: each profile has its own pattern catalogue and helper specifications, generated for that build. See `labbook/2026-11-11-programmatic-compiler-swap.md`. The trust argument in this entry is unchanged.
