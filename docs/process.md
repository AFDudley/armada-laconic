# The process

How to structure work on the Armada contracts: the practice, the order of the work, and the gates. Adapted from
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
it [Abrial 2010].

| Layer | Artifact | What it decides | Gate |
|---|---|---|---|
| Specification | Lean model in `AFDudley/armada-lean` | what must hold, against an attacker | `lake build`, zero `sorry`, `#print axioms` |
| Concrete specification | Stele (EquiVM) | the same contract with concrete values | refinement proof to the Lean model |
| Implementation | Solidity, compiled bytecode on EVMLean | what runs | EquiVM refinement proof to Stele |
| System | the fixturenet (laconic-so stack) | that the deployed system does it | scenario run, recorded verdict lines |

Forge unit tests are an inner loop only. They are never the acceptance of anything.

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

**GATE.** Every new term in a specification comment or a brief is in a glossary, with a citation.

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

**GATE.** Each state fact that a hypothesis would otherwise supply is checked: can the type carry it instead?

## Stratum 2. State the invariant before any transition

**2.1 State the one invariant the machine must never violate, before writing the steps** [Abrial 2010]. For the
adjudicator: every holding is well-formed (its note is live at its recorded position, its preimage is recorded and
nonzero, its key is `npk self r` with `r` published, its token is the holding's asset). Write it as ONE invariant
with ONE step lemma.

**2.2 State the precondition and the postcondition of each operation before its body** [Dijkstra 1976]. In Lean this
is the `Step` case. In Solidity it is the function's `require`s and its documented effect.

**2.3 Choose the coarsest representation that can still state the invariant** [Wirth 1971]. Symbolic hashes (Dolev–Yao)
state everything the specification needs. Concrete values belong to the lower layers.

**2.4 Every assumption is a named hypothesis, never a global axiom.** It carries a citation to the source that justifies
it (a paper, a contract line, a chain property such as positive `block.timestamp`).

**GATE.** The specification states its invariant before its step relation. `#print axioms` on every statement lists
only the model's named primitives.

## Stratum 3. Parse once, at the edge

**3.1 At every point raw input enters, return the precise type** [King 2019]. For a contract: decode and check the
calldata once at the entry point. For the model: the `Step` precondition is the edge.

**3.2 Parse all input before acting on any of it** [King 2019, citing Bratus et al. 2016]. A function that pays out and
then finds a later input invalid cannot undo the payment.

**3.3 Never re-check downstream.**

**GATE.** Each contract entry point checks its input once, before any effect.

## Stratum 4. Draw the machine flat before any state is composite

**4.1 Draw the states flat first** [Samek 2008]. The channel modes are Open, Challenge and Finalized.

**4.2 Draw every transition, then read the diagram for a dead end.** Every call kind has a "can happen" theorem.

**4.3 Make a state composite only when siblings share a transition, an action or a sub-process** [Samek 2008].

**4.4 Refactor to the shared form when a repeated pattern appears, not later** [Samek 2008]. The adjudicator proofs
grew four parallel invariants, each with its own helper copies, because each new fact was added as a new induction.
The repeated pattern appeared at the second one. That is when to refactor.

**GATE.** Every transition has a "can happen" theorem, built by real steps from genesis.

## Stratum 5. Let the prover prove totality

**5.1 End every case split over the call type exhaustively.** Lean's `cases` over `Call` fails when a case is missing.

**5.2 Add a new call kind to the type first, then build.** Every invariant reports the missing case.

**GATE.** Adding a `Call` constructor breaks every invariant proof that does not handle it.

## Stratum 6. Run the machine, shallow, end to end

This stratum decides whether the design becomes a system. No other stratum matters more.

**6.1 Bind every dependency to a real or substitute implementation first** [Cockburn 2005]. In the model: every
statement with a `sorry` proof, building. On the fixturenet: the production deploy, never a test cheat. A library
placed with `vm.etch` exists only in forge's simulation; the real chain never has it.

**6.2 Drive one path through every call, and read the result.** One scenario per call kind, on the fixturenet, with
recorded verdict lines (PASS, FAIL or ERROR, and the stage where it stopped).

**GATE.** The whole model builds with every statement present. The fixturenet stands up from the production deploy.

## The tour. Let coverage order the remaining work

**7.1 Take all-transitions coverage as the baseline for "tested"** [Chow 1978]. Each call kind, in each mode where it
can fire, is exercised by a can-happen theorem and by a fixturenet scenario.

**7.2 Do not target all-paths coverage.** Every source calls it prohibitive [Utting and Legeard 2007].

**GATE.** Every call kind appears in at least one can-happen theorem and one scenario.

## Stratum 8. Check with properties

**8.1 Check the core with properties, not examples** [Hughes 2020]: invariants, postconditions, metamorphic relations
(for example, transfer order does not change payouts), inductive properties, model-based properties.

**8.2 When a check fails, suspect the model first** [Hughes 2016]. In the adjudicator work, every failed proof was a
model defect or a statement gap: reclaim credits, remainder keys, unchecked pool signatures, numeric addresses, the
one-tree pool, payment guarantees without a funding condition.

**8.3 A counterexample stops the work on that statement.** Report it before changing any text. A statement changes
only with the operator's approval.

**GATE.** Every statement is proved with no `sorry`, or it carries a reported counterexample.

## The next chart. Refine one level at a time

**9.1 To add detail, write a machine that refines the previous one.** Write the gluing invariant first [Abrial 2010].
Lean model to Stele: each symbolic term maps to its concrete value, with one named hypothesis (no Keccak or Poseidon
collision among the values a run uses). Stele to bytecode: EquiVM.

**9.2 Strengthen each new guard so the refined machine does no more than the abstract one allowed** [Abrial 2010]. A
contract may add input checks a paper assumes (go-nitro's reclaim checks that SATP leaves to honest signers); it may
never allow more.

**9.3 Finish a level when nothing is left to prove, never when it runs.**

**9.4 Expect to back up** [Wirth 1971].

**GATE.** Zero open obligations for the level.

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

## Sources

Abelson and Sussman 1996; Abrial 2010; Bowen and Hinchey 1995; Bratus et al. 2016; Chow 1978; Cockburn 2005; Dijkstra
1976; Hughes 2016, 2020; Jackson, *Software Abstractions*; King 2019; Lamport 2002; Minsky 2011; Parnas and Clements
1986; Samek 2008; Utting and Legeard 2007; Wirth 1971. Full citations are in exophial4 `docs/process.md`.
