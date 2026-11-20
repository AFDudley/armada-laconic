# lean-temper and the DSL: two steps toward one model (2026-11-20)

Session `01a1f69f` (pane `w3:pG`). Operator rulings of this date are marked **Ruling**. A claim
that was not observed has the mark `[INFERENCE]`. Related: `2026-10-27-bend2-to-evm.md`,
`2026-11-02-lean-as-the-unified-model.md`, `2026-11-10-lean-to-solidity-design.md`,
`2026-11-11-library-selection.md`, `2026-11-11-concrete-specification-design.md`,
`2026-11-19-qwen-repl-step-search.md`; lean-temper `docs/design.md` §9.1 and §10 (commit
`65c49f7`).

## 1. The goal

**Ruling.** The Lean model, with a well-structured set of proofs, is the source for Armada. The
project generates the Solidity and the shared application code for armada-nitro from it.
lean-temper and a DSL are two steps toward this goal. They are not alternatives.

The structure of the model matters for the Solidity, because the model's shape becomes the
Solidity's shape (O6, `2026-11-10-lean-to-solidity-design.md` §2).

## 2. Other provers and Bend2 do not change the choice of Lean

- Isabelle/HOL, Rocq and HOL4 have no equivalent of the four layers this project uses: EVMLean,
  the Lean contract library, Blanc and the armada-lean model. Isabelle would remove most of the
  version churn. Lean's churn is controlled by pinned toolchains and lean-mend (plan item E1).
- Bend2 is not built on Lean. Its default checker is unverified, its translation into the
  `bendtt.lean` kernel has no proof, and its compiler is "99% AI-written and unaudited"
  (bend2.dev reference, 2026-11-09). It has no tactics and no EVM target.
  `2026-10-27-bend2-to-evm.md` already rejects it as a route to bytecode.

## 3. What "DSL" means here

The Lean community uses the term: the community metaprogramming book has a chapter "Embedding
DSLs By Elaboration", and Lean 4's macro system comes from the hygienic macros of Scheme and
Racket (Ullrich and de Moura, "Beyond Notations", 2020). Lake's lakefile language is a DSL in
this sense.

The design has two parts, and only the second is a language:

1. **A construct library.** Each construct carries its semantics, its Solidity form, its
   evaluation lemma, its bytecode lemma, and its interface lemmas and `simp` set. This is plain
   bottom-up abstraction, in the sense of Graham's *On Lisp*.
2. **A closed surface.** A command (`contract`, `machine`) accepts only the library's
   constructs. It generates the standard statements: invariant preservation, can-happen, the
   equation lemmas, and the bytecode theorem for each compiler profile. A library alone cannot
   close the construct set. The open miss list of `armada-sol/ArmadaSol/Translate/Scan.lean`
   shows the cost of an open set.

The idiomatic order in Lean: definitions and lemmas first, `notation` or `macro` next, and a
custom `elab` only to generate declarations or to close the construct set.

**Ruling.** The closed surface is a Lean `elab` command, not a library with linters. The command
expands to plain Lean, so the host keeps every solved problem. A linter cannot close the
construct set or generate the statements; the command does both.

**The strata as clauses.** The strata of exophial4 `docs/process.md`, which armada-laconic
`docs/process.md` adapts, become clauses of the command. A violated stratum is an elaboration
error, not a red gate:

| Stratum | Clause |
|---|---|
| 1. The state space is a type | `state`: an inductive type or a structure with `Prop` fields |
| 2. The invariant before any transition | `invariant`, required before any `call`; proof-carrying fields are subtypes |
| 3. Parse once, at the edge | `init` and calldata decoding are the only constructors of refined values |
| 4. A flat machine; every state reachable | a flat `machine`; one generated can-happen statement for each `call` over `Machine.Reachable` |
| 5. Totality by the compiler | the generated dispatch `match` over `Call` |
| 6. Run shallow, end to end | the shallow embedding runs; statements carry `sorry` until proved |
| The tour | computed from the declared calls |
| The next chart | `refines … via Glue` |

Stratum 0 (vocabulary) and stratum 2.3 (the coarsest representation) stay decisions for a person.
The command states the proof obligations; it does not discharge them.

**Ruling.** The DSL has no knowledge of the pool. It defines a system as the combination of two
general parts, and the pool is one instance:

| Declaration | What it is | What it generates |
|---|---|---|
| `machine` | A proof-carrying state machine: a state type, an invariant, and calls in PE3 form (`Pre`, `Effect`). A call takes and returns states that carry the invariant, so a machine value cannot exist without its preservation proofs. | the `Machine` instance, the dispatch `match`, one preservation obligation and one can-happen statement for each call |
| `contract` | An abstract representation of Solidity: storage of library types, functions as programs in the `Contract` monad, built only from library constructs | the `SolAst` term, the theorem `eval term = program`, and the bytecode statements for each compiler profile |
| `implements … via` | The binding of a contract to a machine through a representation relation (the gluing invariant) | one refinement obligation for each call |

The pool is `machine Pool`, `contract PoolSol`, and `PoolSol implements Pool via Glue`. The Nitro
adjudicator is a second instance of the same three declarations. The machine does not depend on
Solidity, so the application code for armada-nitro is a second target of the same `machine`.
No pool-specific name appears in the DSL's own library.

### 3.1 Prior art (three surveys, 2026-11-20)

Reports: `notes/prior-art-2026-11-20/{cardano-survey,k-rv-survey,actor-survey}.md`. No surveyed
system provides the whole DSL. No system has `machine` with invariants in types and one generated
obligation for each call, together with `embeds` and owner-tagged moves.

| Part of the DSL | Closest precedent | Reuse |
|---|---|---|
| `machine`, obligations from one declaration | Lean-blaster `StateMachine` with `#kind` (IOG, Lean 4); Move Prover `spec` | Design only. Lean-blaster closes a goal with `admit` (its README: "The tool does not perform proof reconstruction right now"), so its proofs fail the zero-`sorry` gate. It pins Lean v4.24.0. Its `#bmc` can search for counterexamples before a proof (stratum 8.2). |
| `contract` and `implements … via`: a bytecode leg and a proof-assistant leg from one declaration | `act` (`argotorg/act`): hevm equivalence and a Rocq export from each `behaviour` | Design only. It confirms the shape. |
| Proofs about one build, not the compiler | KEVM, Kontrol, PlutusCoreBlaster | Already the rule of `docs/process.md` |
| `embeds` | Disel send-hooks (POPL 2018 §1.2, §3.1); TLA+ shared-state composition (*Specifying Systems* §10.4.1) | Compare with `Network.chain` (`ArmadaLean/Nitro/Reachable.lean:510`) before the syntax is fixed |
| Owner-tagged moves | P language ghost and real machines with erasure (PLDI 2013 §3.3); TLA+ `E ⊢→ M` (§10.7); I/O automata input-enabledness | The input-enabled rule matches R.2: the model keeps failing calls |
| Environment events | Timed Rebeca time progress (arXiv:2309.07302 §3); Verdi system transformers (PLDI 2015 §2.4) | Verdi: add time as a transformer of a `machine`, not as a hand-written case |
| A second target (client code) | Hydra Agda spec `OffChainAgreementSpec`; Obsidian (OOPSLA 2020) | Design only |
| A list of every trusted assumption | Hydra `spec/check-trust-ledger.sh` | Fits the primitives and named hypotheses |

The actor model does not have `embeds`, by design: "the internal state of a component must only
be accessible through operations provided by the interface" (Agha et al. 1997 §1). `embeds` is
that rule applied to a field: the embedded state changes only through its own machine's calls.

## 4. Temper and the DSL

- Temper reads existing Lean, finds its repeated patterns, and rewrites it into a normal form.
  The DSL declares that normal form as syntax.
- **Ruling.** Temper approaches the DSL and stops before the syntax. Its passes are a naive,
  exhaustive extraction of the vocabulary. The formal concept analysis on lean-temper branch
  `concepts/fca` (`2f2afb9`, `1b9e31a`) gives candidate types. The boundary analysis (step 4)
  gives the `init` and `call` forms.
- **Ruling.** A port by agents is not cheaper than temper. If the DSL covers the model, a
  program with the same coverage exists. A port runs that program once per instance and keeps no
  record of it. The same agents wrote the 36 aliases and the 414 refine findings.
- Extraction by level:

| Level | Can temper extract it? |
|---|---|
| A vocabulary for each domain (pool, Nitro) | Yes, by formal concept analysis |
| "Contract": `if h : Pre then some (Effect ⟨_, h⟩) else none` in `Pool.apply` (`Pool.lean:1552`) and `Adjudicator.apply` (`Nitro/Adjudicator.lean:290`) | Probably, with an analysis of definition bodies (anti-unification) that does not exist `[INFERENCE]` |
| "Machine" (`ArmadaLean/Machine.lean`) | No. It came from Event-B by ruling on 2026-10-29. `Machine.interleave` had the wrong shape for Nitro, and the five instances share only `init` and `step`. |

The `Pre`/`Effect` pattern is the interface between the specification and the contract
library: it states what each library `call` must refine.

## 5. Proofs are the hard part

An earlier draft of this analysis made three claims that were wrong:

1. "A DSL that elaborates to definitionally equal terms keeps old proofs." Tactics are not
   invariant under definitional equality. `simp` matches lemmas through a discrimination tree at
   reducible transparency, and `rw` abstracts by syntactic key ("motive is not type correct",
   Lean v4.15.0 #6168). If the DSL elaborates to the same `Expr`, it is only notation. If it moves
   a fact into a type, the terms differ.
2. "Transport whole theorems as corollaries, then delete the old form." A corollary through an
   equivalence still refers to the old form, so the old form cannot be deleted. That is an
   adapter, which lean-temper `docs/design.md` §9 forbids. PUMPKIN Pi transforms the proof
   term so that it no longer refers to the old type (Ringer et al., PLDI 2021). **The same defect
   is in E2** (`2026-11-11-library-selection.md`): carrying PE4, PE5 and PE7 through the equality
   and then deleting `Pool.apply` works only if those proofs are rebuilt without `Pool.apply`.
3. "Statement rewrites are local." The alias inventory has 14 Class D findings that need binder
   reordering, 6 Class C losses and 15 implicit parameters, and a change set holds every caller
   (lean-temper `docs/design.md` §2).

What remains: proof migration is a transformation of proof terms across an equivalence, done by
a program and checked by the kernel. A DSL supplies the equivalence for each construct; it does
not remove the transformation.

## 6. Specification statements and helpers

armada-lean computes the split (`docs/spec-module-design.md`): a theorem is a specification
theorem when its doc comment cites a source; it has a record `N.Spec.x` under
`ArmadaLean/Spec/`; `#check_spec_bindings` binds the two. A text count at `integrated` gives about
197 records and about 1,285 theorems.

- A specification statement changes only with a kernel-checked equivalence proof and operator
  approval (rule 8.3).
- A helper may change to any statement that builds.
- The definitions that a specification statement reaches (`Pool.apply`, `ContractState`, `Step`,
  `Reachable`) carry its meaning. `#check_spec_bindings` does not protect them.
- `transact_value_conserved`, `transact_no_double_spend` and `fee_conservation` are helpers by
  the citation rule. Promoting one is an operator decision.

## 7. Roadmap

**Ruling.** Three phases, in order (lean-temper `docs/design.md` §9.1 and §10):

1. **The weak form of done.** No file under `ArmadaLean/Spec/` changes; only `theorem`
   declarations change; the whole model builds and the gate runs on all of it; the step's linter
   and the proof-level rule report nothing. It covers steps 1, 2 and 6.
2. **The strong form of done.** One run on the whole model gives one verdict. Specification
   statements and the protected definition set are unchanged or proved equivalent; helpers build;
   no finding and no adapter remain. It needs the protected set, a configuration mode that
   accepts an equivalence proof, and the whole-model run. Steps 3, 4 and 5 run under it.
3. **`FlashProve/lean`.** The analysis passes run with a new configuration and no code change.
   Obstacles: the linters are in `ArmadaLean/Lint/`; FlashProve has no `Spec/` convention; it pins
   Lean v4.29.1 and lean-temper pins v4.35.0.

The present acceptance set (36 alias findings) is not the whole specification, and it is not the
helpers.

## 8. Open

1. The kind of all 27 alias declarations. A text search resolved 26, all `theorem`.
2. The share of the 217 + 197 refine findings that are invariant or reachability premises on a
   state. A Lean script over the linter's work list can measure it.
3. The contract library (E1) is not built: `armada-sol` has no `Contract` monad.
4. The route for the armada-nitro application code (`2026-11-02-lean-as-the-unified-model.md` §4,
   question 1).
