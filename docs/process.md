# The process

How to structure work on the Armada contracts: the practice, the order of the work, and the gates and judgments of each
stratum. A **gate** is one check that runs, with an id, a kind (static, mutation or execution), the stratum it grades
and the inputs it reads. A **judgment** is a decision that a person or an agent makes and that no machine checks again
(`glossary.md`, "Gate" and "Judgment"). A judgment is not a gate. Adapted from
exophial4's `docs/process.md` (itself ported from exophial3 after five research slices over the practical literature,
2026-08-17). Every step cites a source. The armada-specific examples come from the adjudicator rewrite of
2026-10-23 to 2026-10-26, where this document's absence cost the most.

`glossary.md` and `engineering/12-glossary.md` are the vocabulary authority. `engineering/build-plan.md` S5 states the
verification chain this process serves.

## The one rule every source shares

**BUILD THE WHOLE MACHINE, SHALLOW AND RUNNING, BEFORE ANY PART OF IT IS DEEP.**

Five independent literatures give this same first move. Lamport writes a `TypeInvariant` over a stub `Next` and runs the
checker on day one [Lamport 2002]. Jackson writes minimal signatures and an empty `run`, and confirms an instance before
he adds a fact [Jackson, *Software Abstractions*]. Hughes runs the model alone before he trusts any result about the
code [Hughes 2016]. Cockburn builds the core against a substitute for every dependency [Cockburn 2005]. Abrial writes
the abstract machine and its one invariant before any transition [Abrial 2010].

For Armada this means, in order:
1. every specification statement written, with `sorry` proofs, and the whole Lean model building;
2. the fixturenet standing up with every contract deployed by the production deploy;
3. one scenario driven through every contract call on that fixturenet;
4. only then, deep work on any one proof or contract.

## The layers

The adjudicator, the pool and every later contract pass through the same four layers. Each layer refines the one above
it [Abrial 2010]. Revised 2026-11-11: Stele and EquiVM are dropped (operator rulings 2026-11-08 and 2026-11-11); the
concrete layer is a program written with the Lean contract library, and the implementation is the Solidity that the
library emits (`labbook/2026-11-11-library-selection.md`, `labbook/2026-11-11-symbolic-execution-and-solc.md`).

| Layer | Artifact | What it decides | Gates of the layer (one check each) |
|---|---|---|---|
| Specification | Lean model in `AFDudley/armada-lean`, symbolic (Dolev–Yao) | what must hold, against an attacker | `lake build`, zero `sorry`, `#print axioms` |
| Concrete specification | the same machine as a program written with the Lean contract library (concrete words, storage, ABI) | the same contract with concrete values | refinement proof to the Lean model, through the gluing invariant of 9.1 |
| Implementation | the Solidity the library emits, and its bytecode from the compiler of one compiler profile | what runs | a Blanc proof for each build that the bytecode refines the concrete specification; the lifted certificate matches the one the library predicts for that profile |
| System | the fixturenet (laconic-so stack) | that the deployed system does it | scenario run, recorded verdict lines |

Forge unit tests are an inner loop only. They are never the acceptance of anything.

The Lean model is the one specification for every component that must agree with another: contracts, circuits and the
shared logic of clients. Each target gets the route that can be checked: a proof where a tool exists (Blanc for
bytecode, an R1CS import for circuits), conformance vectors from the model where none exists
(`labbook/2026-11-02-lean-as-the-unified-model.md`).

solc is not verified and is not trusted. The bytecode theorem is about the exact bytecode of one build. The compiler is
swappable by one command (operator ruling 2026-11-11): other solc versions and Solar are compiler profiles, each with
its own generated pattern catalogue and helper specifications, and a swap reports the proof result or the failing
goals (`labbook/2026-11-11-programmatic-compiler-swap.md`). The generated Solidity is idiomatic and
gas-efficient: it may differ from the concrete specification wherever a construct lemma or a proof obligation shows
the two equal (packed storage, `unchecked` arithmetic under a proved bound, custom errors, SafeERC20, events).


## The practice, in the order it applies

1. Acquire the context before you touch anything. Read the code, the design record and **every source in full**. A
   keyword search over a paper is not a reading of it.
2. Settle the design in conversation, so the reasoning is visible before the diff exists.
3. Write the specification, then derive the program from it [Dijkstra; Wirth].
4. Make the state space a type first. An illegal state must be unrepresentable.
5. Parse once, at the edge. The boundary converts untrusted data to a precise type, and nothing downstream re-checks.
6. State the invariant before the transition relation [Abrial; Lamport].
7. Discharge totality by machine. An exhaustive case split is the prover's job, not a reviewer's eye.
8. Classify last. A partition over an open type is not checkable.
9. Keep the pure computation apart from the effects.
10. Let a failure propagate. Add no fallback, no retry, no blanket catch.
11. Verify by running the artifact you ship. Execution is evidence; reading is opinion.
12. Own every consequence of a change. Find every caller, and delete the orphan.
13. Keep the code and the record as one graph. A change to one makes the other suspect.

## Stratum 0. Fix the vocabulary

One word per concept, before any design work [Abelson and Sussman 1996, stratified design]. Use the glossary's term.
Write in ASD-STE100. An agent reads a brief under the same constraints a person does; one claim per sentence is a
defect control, not a courtesy.

**Judgment J0 (stratum 0; inputs: each brief and each specification comment).** Every new term in a specification
comment or a brief is in a glossary, with a citation.

## Stratum 1. Make the state space a type

**1.1 Write every state as its own type, and one constructor per legal configuration** [Minsky 2011]. A record with
optional fields lets a caller build a combination that never occurs.

**1.2 Make the illegal state unrepresentable, not detected.** Armada examples:
- A contract address is a hash output. Model it as a symbolic Keccak term, never as a number. A number invites
  arithmetic (`% 2^160`, bounds) that has no meaning for a hash. Derive `self` from its creation inputs
  (`self = Term.keccak selfInputs`) so "self is an address" holds by construction, with no hypothesis.
- A note is identified by its leaf position, `(treeNumber, index)`, never by its content. Two notes with the same
  commitment are two notes.
- Keccak and Poseidon are distinct function symbols, so their outputs never collide by construction.

**1.3 Name every absence.** An absent holding is `none`, and a zero-value note is never recorded.

**Judgment J1 (stratum 1; input: each hypothesis).** Each state fact that a hypothesis would otherwise supply is
checked: can the type carry it instead?

## Stratum 2. State the invariant before any transition

**2.1 State the one invariant the machine must never violate, before writing the steps** [Abrial 2010]. For the
adjudicator: every holding is well-formed (its note is live at its recorded position, its preimage is recorded and
nonzero, its key is `npk self r` with `r` published, its token is the holding's asset). Write it as ONE invariant
with ONE step lemma.

**2.2 State the precondition and the postcondition of each operation before its body** [Dijkstra 1976]. In Lean this
is the `Step` case. In Solidity it is the function's `require`s and its documented effect.

**2.3 Choose the coarsest representation that can still state the invariant** [Wirth 1971]. Symbolic hashes (Dolev–Yao)
state everything the specification needs. Concrete values belong to the lower layers.

**2.4 An axiom may declare a primitive. It never asserts a property.** An uninterpreted primitive (a verifier, a hash, a
parser) may be an `axiom` that only declares its type, such as `EdDSAVerify` or `decodeGuarantee`. Such an axiom states
nothing, so it cannot make the logic inconsistent. A property (unforgeability, collision resistance, timely inclusion,
positive `block.timestamp`) comes from the construction, as the Dolev–Yao terms give collision resistance, or from a named
hypothesis on each theorem that needs it. It is never an `axiom`, because a false axiom lets Lean prove anything. Each
assumption cites its source (a paper, a contract line, a chain property). Operator ruling, 2026-10-27.

**Judgment J2 (stratum 2; input: the specification).** The specification states its invariant before its step
relation.

**Gate G2 (stratum 2; execution of `#print axioms`; input: every statement).** The output lists only the model's named
primitives.

## Stratum 3. Parse once, at the edge

**3.1 At every point raw input enters, return the precise type** [King 2019]. For a contract: decode and check the
calldata once at the entry point. For the model: the `Step` precondition is the edge.

**3.2 Parse all input before acting on any of it** [King 2019, citing Bratus et al. 2016]. A function that pays out and
then finds a later input invalid cannot undo the payment.

**3.3 Never re-check downstream.**

**Judgment J3 (stratum 3; input: each contract entry point).** Each contract entry point checks its input once, before
any effect.

## Stratum 4. Draw the machine flat before any state is composite

**4.1 Draw the states flat first** [Samek 2008]. The channel modes are Open, Challenge and Finalized.

**4.2 Draw every transition, then read the diagram for a dead end.** Every call kind has a "can happen" theorem.

**4.3 Make a state composite only when siblings share a transition, an action or a sub-process** [Samek 2008].

**4.4 Refactor to the shared form when a repeated pattern appears, not later** [Samek 2008]. The adjudicator proofs
grew four parallel invariants, each with its own helper copies, because each new fact was added as a new induction.
The repeated pattern appeared at the second one. That is when to refactor.

**Judgment J4 (stratum 4; inputs: the transitions and the theorems).** Every transition has a "can happen"
theorem.

**Gate G4 (stratum 4; execution of `lake build`; input: the can-happen theorems).** Each one is built by real steps
from genesis, and Lean accepts it.

## Stratum 5. Let the prover prove totality

**5.1 End every case split over the call type exhaustively.** Lean's `cases` over `Call` fails when a case is missing.

**5.2 Add a new call kind to the type first, then build.** Every invariant reports the missing case.

**Gate G5 (stratum 5; mutation: a new `Call` constructor, then `lake build`; input: the invariant proofs).** Adding a `Call` constructor breaks
every invariant proof that does not handle it.

## Stratum 6. Run the machine, shallow, end to end

This stratum decides whether the design becomes a system. No other stratum matters more.

**6.1 Bind every dependency to a real or substitute implementation first** [Cockburn 2005]. In the model: every
statement with a `sorry` proof, building. On the fixturenet: the production deploy, never a test cheat. A library
placed with `vm.etch` exists only in forge's simulation; the real chain never has it.

**6.2 Drive one path through every call, and read the result.** One scenario per call kind, on the fixturenet, with
recorded verdict lines (PASS, FAIL or ERROR, and the stage where it stopped).

**Gate G6a (stratum 6; execution of `lake build`; input: the whole model).** The whole model builds with every
statement present.

**Gate G6b (stratum 6; execution of the fixturenet stand-up; input: the production deploy).** The fixturenet stands up
from the production deploy.

## The tour. Let coverage order the remaining work

**7.1 Take all-transitions coverage as the baseline for "tested"** [Chow 1978]. Each call kind, in each mode where it
can fire, is exercised by a can-happen theorem and by a fixturenet scenario.

**7.2 Do not target all-paths coverage.** Every source calls it prohibitive [Utting and Legeard 2007].

**Judgment J7 (the tour; inputs: the can-happen theorems and the scenarios).** Every call kind appears in at least
one can-happen theorem and one scenario.

## Stratum 8. Check with properties

**8.1 Check the core with properties, not examples** [Hughes 2020]: invariants, postconditions, metamorphic relations
(for example, transfer order does not change payouts), inductive properties, model-based properties.

**8.2 When a check fails, suspect the model first** [Hughes 2016]. In the adjudicator work, every failed proof was a
model defect or a statement gap: reclaim credits, remainder keys, unchecked pool signatures, numeric addresses, the
one-tree pool, payment guarantees without a funding condition.

**8.3 A counterexample stops the work on that statement.** Report it before changing any text. A statement changes
only with the operator's approval.

**Gate G8 (stratum 8; execution of `lake build`; input: every statement).** Lean accepts every proof, and it reports
each `sorry`.

**Judgment J8 (stratum 8; inputs: each statement with a `sorry` and the counterexample reports).** Each such statement
carries a reported counterexample.

## The next chart. Refine one level at a time

**9.1 To add detail, write a machine that refines the previous one.** Write the gluing invariant first [Abrial 2010].
Lean model to the concrete specification (the library program): a typed decoder for each ABI position maps each
concrete word to its term; a word that the run computed as a hash maps to that hash term, and any other word to `val`
in its range. The encodings need no assumption beyond what the EVM forces, so the theorem names three items: collision
freedom of Keccak-256, collision freedom of its 160-bit truncation, and no Poseidon output equal to a Keccak output
among the run's values (operator ruling 2026-11-02; `labbook/2026-11-02-refinement-risks.md` §7). The library program
is the chain's executable step, as `Adjudicator.apply` is for Nitro; the off-chain world (signers, notes, witnesses,
the network) is modelled around it. Concrete specification to bytecode: Blanc, by symbolic execution of the build's
bytecode.

**9.2 Strengthen each new guard so the refined machine does no more than the abstract one allowed** [Abrial 2010]. A
contract may add input checks a paper assumes (go-nitro's reclaim checks that SATP leaves to honest signers); it may
never allow more. The Lean model is the specification, and the contracts are rewritten from it. The current Solidity
is not a source to reconcile with the model (operator ruling 2026-11-02). Where it allows more, it is a defect that the
rewrite does not copy: for example go-nitro's 12-byte destination rule and the owned path's 160-bit `ownerField`
truncation.

**9.3 Finish a level when nothing is left to prove, never when it runs.**

**9.4 Expect to back up** [Wirth 1971].

**9.5 A safety certificate carries only safety.** A one-way inclusion (an EquiVM certificate) does not carry a
"can happen" or liveness theorem to bytecode (`labbook/2026-10-28-liveness-refinement-design.md` §(a)). Blanc proves
the existence of a successful run with its exact gas, so the bytecode theorem for each function has two parts: the
run exists when the concrete specification succeeds, and no successful run exists when it reverts. Each Liveness
theorem is proved once in Lean and carried to bytecode through those two parts (that entry's §(d)). The transfer
allows a bounded number of failed honest calls (`labbook/2026-10-28-verification-chain-rulings.md`, O2 correction). A
size bound that this route adds to an approved statement goes to the operator with its gas measurement.

**9.6 Tie each artifact that the model does not produce to the model.** A circuit: import its R1CS into Lean, bind
its Poseidon and EdDSA templates to the trusted primitives, and prove the rest equivalent to the model's constraint
predicate. A deployed key: check it against the key exported from the ceremony zkey. A client: run conformance vectors
that the model generates. `labbook/2026-11-02-refinement-risks.md` §7 gives the reasons and the order.

**Judgment J9 (the next chart; input: the level's proof obligations).** Zero open obligations for the level.

## The repair pass. Bring existing code to the strata

The strata above describe building forward. Rule 4.4 assumes each repeated pattern is refactored when it appears.
When a model is already proved but was not built that way, the repair pass brings it to the strata. Operator ruling,
2026-11-01. The Nitro model is the first case: its 115 spec theorems state about seven ideas (one equation per
operation, one safety invariant, one liveness principle, the Turbo value algorithm, the ledger constructions, the
virtual-channel protocol, and proof helpers).

A tool does the repair pass, not a person or an agent. The tool is lean-temper (`AFDudley/lean-temper`, design record
`docs/design.md`), a compiler from a Lean project that builds to the same project in the form of the strata. Operator
ruling, 2026-11-20. lean-temper replaced the codemods that agents ran from 2026-11-01 to 2026-11-19.

**R.1 Measure the gap before changing anything.** The audit (`docs/lean/audit-2026-10-31.md`) and the type-doctrine
linters (`docs/lean/type-doctrine.md` §3) give the work lists: each `repeatedPremise` group is one missing refined
type, and each `layerDiscipline` hit is one proof that unfolds a base operation. Before a pass is built, a Lean script
makes its inventory from the linter's work list (operator ruling, 2026-11-19).

**R.2 Rewrite to the strata's forms.** The targets are one equation per operation (2.2), one invariant with one step
lemma (2.1), and refined types instead of hypotheses (1.2, 3.1). An operation equation takes the dependent-`if` form,
so the effect receives the precondition as a value: `apply (.transfer a b) s = if h : s.Finalized a then some
(transferEffect ⟨a, h⟩ b s) else none`. The model keeps failing calls, because the adversary may submit any call;
go-nitro reverts them. Each lean-temper pass replaces a form with an equivalent form, and Lean checks each change
before the tool keeps it. lean-temper has passes for aliases, refined types and premise variants (`design.md` §5). It
has no pass for the layer rule (`layerDiscipline`) yet.

**R.3 Prove each changed specification statement equivalent to the statement it replaces.** Lean proves `New ↔ Old`.
A new statement proved equivalent to an approved statement changes shape, not meaning. The same applies to each
definition that a specification statement reaches. A helper theorem needs no equivalence proof.

**R.4 Rewrite each proof in place.** The tool changes the proof of the theorem that callers already use. It adds no
helper that restates an existing theorem's premises, and no adapter: no function, `subst` or transport that restores
an old context (`design.md` §5.2).

**R.5 Delete the replaced statements and their helpers.** No alias and no deprecated copy stays.

**Done.** A run is done in the weak form or in the strong form of lean-temper `design.md` §9.1. In both forms, the
whole model builds, each pass's linter reports no finding in its scope, and no adapter remains. The weak form
changes only theorems, and no specification statement or definition. The strong form is one run on the whole model
with one verdict, and it changes specification statements and definitions only with the proofs of R.3.

The net line change of the model code (`ArmadaLean/` outside `ArmadaLean/Lint/`) is a measurement, not a condition.
armada-lean `scripts/check.sh --json --base REF` reports it as `model_net_lines`. A run that adds lines is examined
for a missing deletion (R.5) or an adapter (R.4).

Lessons from the manual passes on the Nitro adjudicator (armada-lean `7057602b`, `cb84006a`, `draft/r4-rest`,
2026-11-01, and the pass of 2026-11-02). They are the reasons for the rules of the tool:

- **A new helper that restates an existing theorem's premises raises the counts.** It raised `repeatedPremise`,
  `looseHypothesis` and `aliasHypothesis` above the baseline, and armada-lean's lint check (`scripts/check.sh`,
  `lint_ok`) refused it. Hence R.4.
- **Do not write `s.apply …` in a `show` or `have` when the theorem's own statement does not name `apply`.** The
  ascription puts `Adjudicator.apply` into the proof term, and `layerDiscipline` flags it. State the derived equation in
  the operation's own form (`have heq : s.challenge … = if h : … then … else none := apply_challenge_equation …`) and
  rely on definitional equality with `s.apply apps caller (.challenge …)`.
- **A base operation that the statement itself names is not a layer violation.** An equation about `apply` puts
  `apply` into every proof term, whatever the technique. The linter subtracts the statement's constants from the proof's
  (armada-lean `88356264`). Before this fix, 26 of 35 Nitro `layerDiscipline` hits were this case.
- **Check that an acceptance clause can be met before it goes to a worker.** The first R.4 pebble required "no
  `layerDiscipline` finding" while the linter had the defect above. The worker spent about 290k tokens before it proved
  the clause impossible.
- **A work item whose acceptance is a lint count gets gamed.** The manual merges made `Safety.lean` larger
  (`reclaim_decompose` beside `reclaim_some`, which kept 68 uses), and a codemod worker copied a step's post-state
  into a new `def` so that the base operations left the proof. Hence a tool with no person in the loop, R.4 and R.5.
- **Measure the linter before the code.** 10 of 15 `layerDiscipline` hits counted names that entered a proof only
  through a type (a constructor field type, a `casesOn`, a rewrite motive), and `looseHypothesis` counted type
  parameters as inputs, which hid 106 findings (armada-lean `draft/cm-opus`, failure tests in
  `ArmadaLean/Lint/TypeDoctrineTest.lean`). A commit that corrects a measure may raise a count; it is its own commit
  and states the defect and the counts.
- **A spec theorem is finished when its proof and its witness are closed.** `sorry_count` counts proofs only. A
  witness proved by `sorry` stays in `witness_open`, and `scripts/check.sh --json NAME` reports both
  `theorem_in_sorry` and `theorem_witness_open` for one theorem.

**Gate GRa (the repair pass; execution of the lean-temper run; input: armada-lean at a pinned commit).** The run is done
in the weak or the strong form.

**Gate GRb (the repair pass; execution of `lake build`; input: the run's result).** armada-lean builds.

**Gate GRc (the repair pass; execution of `scripts/check.sh`; input: the run's result).** No type-doctrine count rises
above `scripts/gate-baseline.json` (`lint_ok`).

## The decision rules to keep getting right

**Hypothesis or construction?** If a fact holds in every real execution, make the type carry it. Add a hypothesis only
for a fact about the world the model cannot construct (an honest participant, a chain property), and cite its source.

**Model defect or statement gap?** A counterexample against the model is a model defect when the real contract or
the source rules it out. It is a statement gap when the source itself needs the extra condition.

**Unbuildable or merely detected?** The test is whether the result type is stronger than the input type [King 2019].

## Anti-patterns, each one measured

**Keyword reading of a source.** The first adjudicator statements came from searches over the papers. The payment
guarantee used Nitro §4.4 (finalizing) and missed §5.1-5.2 (what a transfer pays). A full reading added 14 statements
(`notes/statement-audit-2026-10-25/`).

**One invariant written four times.** Facts added one at a time as new inductions, each with helper copies; about 11,800
lines of proof, most of it duplication.

**Arithmetic on a hash.** Addresses as numbers led to truncation lemmas, address bounds and a deposit rule that no
source supports.

**A test that does not use the production deploy.** Poseidon placed by a forge cheat; every test passed; the real
fixturenet had no library.

**Evidence that never reached the claim.** A deposit run that stopped at the funding check was reported as proof that
the Poseidon link worked; the Poseidon call was never reached.

**A plan instead of the work.** Builders who stop to write a plan for work whose plan is already decided move the goal
by zero [Bowen and Hinchey 1995].

**A specification that stops tracking the code.** Written once and never re-run, it has not been verified; it has
merely not been falsified [Parnas and Clements 1986].

Measured in the bytecode work of 2026-11-10 and 2026-11-11 (armada-laconic session `01a195d3`, evm-proof-toolchain
`armada-sol/gen/TIMES.md`):

**Deep before running.** One function's bytecode walks went deep for two days while most of the pool had no concrete
form and nothing was deployed. This breaks the one rule at the top of this document.

**Rebuilding a tool's own mechanism.** Blanc already had forward walks (`xrun`, 228 uses on Dai, USDC and USDT),
memory-growth walks (`WalkFpMove.lean`), walks to an outcome that a callee decides (`RunEx.lean`, 187 files) and a
shape for an external-call boundary (`Usdc/TransferLiveProxy.lean:95-103`). Agents rebuilt each one, about 20 active
agent hours, because nobody read how the tool proves its own examples (practice 1). A report that a tool "does not
work" is accepted only with the failing goal and a place where the same tool works.

**Cutting one function at arbitrary points.** A split of one straight-line path into parallel pieces forced each
piece to state an intermediate state by hand, and three of those states were wrong. Split work only at boundaries
that have a specification (a callee entry, a library lemma, a component with a defined input and output). Symbolic
execution computes the states inside a path; nobody writes them.

**A narrowed trial instead of the work.** A "cheap first test" answers a local question and leaves the design
undecided. State the requirement, the interfaces and the acceptance, and build that.

**Transliteration instead of refinement.** A rule that the emitted Solidity copies the model one to one removed a
gas optimization (`if (fee > 0)` around a token call) and produced unidiomatic code. The emitted code may differ from
the concrete specification wherever a lemma shows the two equal.

## Sources

Abelson and Sussman 1996; Abrial 2010; Bowen and Hinchey 1995; Bratus et al. 2016; Chow 1978; Cockburn 2005; Dijkstra
1976; Hughes 2016, 2020; Jackson, *Software Abstractions*; King 2019; Lamport 2002; Minsky 2011; Parnas and Clements
1986; Samek 2008; Utting and Legeard 2007; Wirth 1971. Full citations are in exophial4 `docs/process.md`.
