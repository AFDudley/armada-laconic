# Programmatic compiler swap: other solc versions and Solar

Date: 2026-11-11. Session `01a195d3`, from the brief of session `01a1cbb5` ("yul lifting",
`.scratch/briefs/2026-11-11-programmatic-compiler-swap.md`). Process: `docs/process.md`.

## Operator ruling (2026-11-11)

The pipeline must support a **programmatic** swap of the compiler. The swap targets are other solc versions and Solar
(`paradigmxyz/solar`). "Programmatic" means: one command changes the compiler tuple and runs the pipeline again. The
output is the proof result, or a list of the constructs and helpers that failed, each with its failing goal. No person
or agent edits a file for the swap itself. An agent works only on the reported failures.

The operator's words: "we should be able to do a ... replacement of solc with different versions of solc and solar",
then "I meant more programatic swap not blind."

## Finding: the documents before this entry did not support it

The trust argument already allows any compiler. `labbook/2026-11-11-symbolic-execution-and-solc.md:17-19`: the theorem
is about the exact bytecode, so a defective compiler makes the proof fail; it cannot make a false proof.

The proof automation was tied to one build:

| Plan element | Source | Effect of a swap |
|---|---|---|
| A construct has "its expected bytecode block pattern for the pinned solc build" | `glossary.md:130` (before this entry) | every pattern becomes invalid |
| The predicted certificate must match the lifted certificate | `glossary.md:132` (before this entry) | a different compiler fails this gate before any walk |
| Helper callee specifications are "proved once per solc version" | `symbolic-execution-and-solc.md:23`, `library-selection.md:36` | each version needs a new library, by hand |
| P5 pattern catalogue "under the pinned settings, located through solc's source maps" | `2026-11-10-lean-to-solidity-design.md:67` | nobody has checked that Solar emits compatible source maps |
| Solar "would need its own automation package" | `2026-11-02-equivm-toolchain.md:68-70` | written for EquiVM; the same fact applies to Blanc |
| The "vintage" tuple and a regeneration script | `2026-11-02-equivm-toolchain.md:95-98`, `:123` | a proposal only, not decided |

What already runs for any compiler: the emitter, the compile command, `lift.py` with `Cert.check` (M0 passed six
compiles of two fixtures, solc 0.8.31, viaIR on and off: `2026-11-10-lean-to-solidity-design.md:44-45`), and P7 when
P5 and P6 exist for the build.

## Checked on 2026-11-11 (this session)

- Every line the brief cites holds, with one correction: the library excludes `assembly` and `delegatecall`
  (`library-selection.md:32-33`) and dynamic storage arrays (`library-selection.md:23`, not `:32-34`).
- solc 0.8.31 (`~/.svm/0.8.31/solc-0.8.31 --help`) has `--ir`, `--ir-optimized` and `--yul-optimizations <steps>`.
- Standard JSON with `settings.optimizer.details.yulDetails.optimizerSteps` set compiles a test contract under viaIR,
  `osaka`, 200 runs, with no error.
- `--ir` output of a contract with `x += a` names the helpers `checked_add_t_uint256`,
  `abi_decode_tuple_t_uint256` and `cleanup_t_uint256`.
- Not checked: Solar's options, its source maps, and whether `--ir-optimized` with source maps connects each named
  helper to its bytecode block.

## Proposed design (all `[INFERENCE]`, untested)

1. **Compiler profile.** One record: the command tuple (compiler, version, viaIR flag, optimizer runs, EVM version,
   and the Yul optimizer step sequence where the compiler has one), the pattern catalogue, and the helper
   specifications. Each construct lemma takes the profile as a parameter. The predicted certificate is composed from
   the patterns of that profile, so the match stays a valid gate for each profile. It names no single solc build.
2. **P5 by script.** For each construct, compile one fixture and read its block from the source map (M0b's method,
   `2026-11-10-lean-to-solidity-design.md:46-48`, as a program).
3. **P6 statements generated, `xrun` proves them.** The statement comes from the construct semantics and the extracted
   block. A construct without a loop is straight-line code, and symbolic execution computes its states
   (`docs/process.md:352`). V7's hand time went to solc idioms and statement forms
   (`2026-11-10-lean-to-solidity-design.md:42-43`); a generated statement removes the second cost.
4. **Helper specifications by the same method.** Locate each helper entry in the lifted certificate and generate the
   walk statement.
5. **Failure report.** For each construct and helper: "proved", or the failing goal.
6. **Acceptance.** One command builds the same library program with solc 0.8.31, one other solc version, and Solar.
   It gives the proof result or the failure list, with zero hand edits.

Facts in favour: the library subset excludes `assembly`, `delegatecall` and dynamic storage arrays, so Solar is more
likely to accept it `[INFERENCE]`. M0b found that the inlined and called forms of a construct differ only in the call
wrapper, so one lemma can cover both. Open risk: loops (D2). The fold invariant must be attached at the lifted loop
head, and a compiler can change the loop form.

## The compiler source: black box for trust, grey box for generation

- **Trust.** The compiler stays a black box. An argument from its source would make it trusted, which is option B,
  and the operator ruled option B out (`symbolic-execution-and-solc.md:47`).
- **Generation.** The generator may use any compiler information. Wrong information makes a proof fail, not a false
  theorem.
- **Use of the source.** Under viaIR, solc builds helpers from fixed templates with fixed names
  (`checked_add_t_uint256`, `abi_decode_tuple_…`, `cleanup_t_…`; three of them observed in `--ir` output above). The
  source gives the complete template inventory once for each compiler family `[INFERENCE]`. A template diff between
  versions names the helper specifications to prove again. For Solar, the source is the only reference for its code
  shapes.
- **Limit of the source.** The optimizer output is a combination of rewrite rules. The source gives no finite pattern
  list for it. So the walk over the real bytecode stays the general method.
- **Better input than the source.** `--ir` gives unoptimized Yul with named helpers (observed). `--ir-optimized` and
  the source maps may connect them to the output `[INFERENCE]`. Standard JSON accepts
  `optimizer.details.yulDetails.optimizerSteps` (observed), so the profile can pin the step sequence. A diff of the
  generated helper Yul between two builds is more precise than a diff of the compiler source `[INFERENCE]`.
