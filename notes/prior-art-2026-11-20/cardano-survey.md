# Cardano prior art survey: a planned Lean 4 DSL for proof-carrying machines and contracts

Date: 2026-11-20. This report answers one question. Does the Cardano ecosystem
already have the DSL that Armada plans to build in Lean 4? A claim without a
citation carries the mark `[INFERENCE]`. Sub-agents performed the web research
for each candidate and read each primary source. This writer verified the
most load-bearing finding (Lean-blaster) against the live repository.

## The planned DSL, for reference

Armada plans three Lean 4 `elab` declarations (`labbook/2026-11-20-temper-and-the-dsl.md`
§3–4):
- `machine`: a proof-carrying state machine. It generates the dispatch, one
  preservation obligation, and one can-happen statement for each call.
- `contract`: an abstract model of on-chain code. It generates a syntax tree,
  an evaluation equality, and bytecode proof statements per compiler profile.
- `implements … via`: a refinement obligation for each call, through a
  gluing invariant.

It must also support environment events, owner-tagged moves (honest,
adversary, environment), one machine embedded in another, trusted primitives
with laws, and a second target (client code) from the same `machine`.

## The table

Columns: (1) invariant in types or proved per transition; (2) proof
obligations generated from a declaration; (3) abstract on-chain
representation with a refinement proof to the deployed artifact; (4) one
model, two targets (on-chain and client); (5) Nitro features present; (6)
proof assistant / last commit / license; (7) reuse for EVM and Solidity.

| Candidate | 1. Invariant | 2. Generated obligations | 3. On-chain refinement | 4. One model, two targets | 5. Nitro features | 6. Assistant / last commit / license | 7. EVM/Solidity reuse |
|---|---|---|---|---|---|---|---|
| **Marlowe** (Isabelle proofs) | Proved once, per semantic function, over a fixed interpreter. Not type-carried. (`fc20.ifca.ai/wtsc/WTSC2020/WTSC20_paper_18.pdf` §5.2) | No. Six theorems (validity, money, quiescence, timeout, termination, transaction bound) are hand-proved once for the fixed Marlowe semantics, not generated per user contract. (same paper §5.1–5.4; `isabelle/Core/*.thy`) | No. One fixed interpreter runs both on-chain and off-chain; no separate abstract model and no refinement proof to a script. (paper §3.3) | Yes. One Haskell semantics compiles to the validator, runs in the emulator, and ports to PureScript for the Playground. (paper §3.3, §6) | Environment: a timeout fires without a deposit (paper §2). Owner tags: partial, `Party` only, no honest/adversary/environment split. Embedding: none found `[INFERENCE]`. Trusted primitives with laws: none found `[INFERENCE]` (`isabelle/Util/ByteString.thy` has no crypto laws). | Isabelle/HOL. Proof repo `input-output-global/marlowe` archived, last push 2024-09-02. Apache-2.0. | Design only. The theories model Marlowe's own five constructs and EUTxO; no EVM/Solidity artifact. `[INFERENCE]` |
| **Plutus `StateMachine`** (plutus-apps) | Neither. `StateMachine s i` carries an unchecked transition function; nothing forces an invariant. (`plutus-contract/src/Plutus/Contract/StateMachine/OnChain.hs`, `StateMachine` definition) | No. `mkValidator` compiles the transition function to a validator; it emits no theorem. (same file, `mkValidator`) | No. The Haskell transition function is itself the on-chain code; no separate model, no refinement proof. (same file, module header) | Yes. The same `smTransition` drives `mkValidator` on-chain and `StateMachineClient` off-chain. (`StateMachine.hs`, `$statemachine` doc comment) | Environment: none, every step needs a submitted transaction. Owner tags: none `[INFERENCE]`. Embedding: none `[INFERENCE]`. Trusted primitives with laws: a `ThreadToken` NFT exists as a mechanism, with no stated law. | No proof assistant. Repo `IntersectMBO/plutus-apps` archived in maintenance mode, last push 2024-09-27. Apache-2.0. | Design only. All types are Cardano/Plutus-specific. `[INFERENCE]` |
| **Plutarch** | No state-machine construct exists. Only per-value type tagging (`Term s a`). (plutarch-plutus.org/Introduction.html, "Overview") | None. The toolchain reports UPLC and cost benchmarks, not proofs. (GitHub README, "Why Plutarch?") | No stated proof. `Term`/Plutarch-Type is the abstract layer; `compile` maps it to UPLC with no proof arrow in the guide's own diagram. (Introduction.html, bridge diagram) | No. On-chain validators only. `[INFERENCE]` | None of the four found. `[INFERENCE]` | Plain Haskell/GHC, no proof assistant. Repo pushed 2026-11-20. MIT. | Design only (typed eDSL over a low-level stack-machine target). Code is UPLC-specific down to its builtin set. |
| **Aiken** | No state-machine or invariant construct; validators are hand-written functions. (GitHub README; docs.cardano.org/.../aiken) | None. `aiken check` runs unit and fuzz tests, not proofs. (aiken-lang.org/faq Q17) | No. `gen_uplc.rs` compiles the surface language straight to UPLC with no stated proof. (`crates/aiken-lang/src/gen_uplc.rs`) | No. On-chain validators and blueprints only. `[INFERENCE]` | None of the four found. `[INFERENCE]` | No proof assistant, Rust type checker only. Repo pushed 2026-11-17. Apache-2.0. | Design only: the built-in property-based fuzz-testing toolchain (`aiken/fuzz`) is the one transferable idea, and it is a testing idea, not a proof idea. |
| **Plinth** (formerly PlutusTx) | Not a state-machine DSL. No type-carried or per-transition invariant construct. (plutus.cardano.intersectmbo.org/docs/essential-concepts/plinth-and-plutus-core) | Compiler-internal only: `--certify` proves each optimization pass preserves UPLC meaning, not a per-contract obligation. (plutus.cardano.intersectmbo.org/metatheory/latest, "Certified Compilation") | Yes, at the compiler level. Plinth compiles to UPLC; UPLC and the CEK machine have a full Agda formalization (progress, preservation); `--certify` checks each pass against it. Plinth itself has no fully verified semantics, only the target language does. (IOG paper, "Plinth: A Plugin-Powered Language Built on Haskell", `iog.io/api/research/pdf/HRZ4UJBX`) | No. Off-chain code is ordinary Haskell outside the compiled subset. `[INFERENCE]` | None of the four as DSL constructs. `[INFERENCE]` | Agda for the metatheory only, not for user contracts. `IntersectMBO/plutus` monorepo, last push 2026-11-19. Apache-2.0. | Design only: a certifying compiler checking each optimization pass against a mechanized target semantics is close to Armada's per-compiler-profile bytecode proof. Not the code itself (UPLC-specific). |
| **Agda `formal-ledger-specifications`** | Proved per transition. The library's `LedgerInvariant STS P` is a reusable predicate, proved once per STS instance by hand. (`Interface/STS.lagda.md`, "Invariants") | Partially. `STS-total` (totality/can-happen) and `LedgerInvariant` (preservation) are reusable obligation *shapes*; about 25 concrete systems each instantiate and prove them by hand, not from one macro. (same source, "Totality", "Invariants") | No bytecode target; it models the ledger layer. It extracts Haskell via agda2hs and proves `bisimilarityProof` between a conformance model and the Conway spec, used to conformance-test cardano-ledger, the closest analogue but not a bytecode refinement. (`Ledger.Conway.Conformance.html`) | Not found. One extracted Haskell ledger artifact only. `[INFERENCE]` | Environment: yes, `Γ` is explicitly "read from... the outside world"; epoch/time events are named. Owner tags: not found `[INFERENCE]`. Embedding: rules call sub-rules (CHAIN→...→UTXO) but no stated "changes only through the parent" guarantee `[INFERENCE]`. Trusted primitives with laws: yes, `PKKScheme` postulates sign/verify and proves `isSigned-correct`. (`Ledger.Core.Specification.Crypto.html`) | Agda. `IntersectMBO/formal-ledger-specifications`, last push 2026-11-19. Apache-2.0. | Design only: the generic `Interface.STS` relation/trace machinery and the totality/invariant obligation pattern are language-agnostic. `[INFERENCE]` |
| **plutus-metatheory** | Carried in the types. Terms are intrinsically typed (`Algorithmic`); preservation holds by construction of the reduction relation's type. (README "Features"; metatheory/latest) | Hand-written per aspect, not from one declaration. `progress` is one theorem over the whole language; soundness/completeness are separate modules. (metatheory/latest, module list) | Yes, the closest ledger-adjacent match. The intrinsically typed term is the abstract model; `VerifiedCompilation`/`--certify` proves, per optimization pass, that output matches a syntactic specification relation. CK/CEK correspondence is also proved. (metatheory/latest, "Certified Compilation") | No. Targets the on-chain Plutus Core language and a CLI tool (`plc-agda`) only. `[INFERENCE]` | None of the four; the formalism is term reduction, not a stateful ledger with roles. `[INFERENCE]` | Yes for trusted primitives: builtins specified via `Builtin.Signature`, some postulated; integer-division builtins carry proved laws (`Builtin.Integer.Properties`). | Agda. Deprecated standalone repo moved into `IntersectMBO/plutus`; directory last commit 2026-11-19. Apache-2.0. | Design only: intrinsic typing, the reduction→CC→CK→CEK correspondence chain, and the certified-compiler-pass pattern are language-agnostic. No EVM/Solidity code. |
| **Hydra Head formal spec** (`cardano-scaling/hydra/spec`) | Hand-written transition relations and per-state `*Valid` bundles; invariants (e.g. global solvency) proved as closed theorems over the model, not type-carried. (`docs/dev/agda.md`, "Module map", "The trust story" tier 1) | Yes, hand-authored, not macro-generated. §7 proves consistency, soundness, completeness, reachability, no-settlement-without-unanimity, non-stuckness, value conservation, bounded contest window, solvency, each parameterised over 4 named `Assumptions`. `ReferenceBridge.agda` proves spec-validity implies checker-acceptance. (agda.md, "The trust story" tiers 1 and 3) | Partial. The spec extracts to a Haskell reference checker (MAlonzo) with a proved bridge to it. But the link from that checker to the deployed Plutus validator and compiled `deposit.ak` UPLC is a differential **test** (`HeadValidatorAgreement`), not a machine-checked proof; the gap (6 mocks, 7 encoding postulates) is itself enumerated. (agda.md, "The extraction pipeline", "The agreement tests") | Yes. `OnChain.lagda.typ` models the validator; `OffChain.lagda.typ` §6 models the `hydra-node` handler, bound by `Hydra.OffChainAgreementSpec`. (agda.md, "Module map", "The agreement tests") | Trusted primitives with laws: yes, hashing, multisig, and a KZG-style accumulator are postulated with stated laws (agda.md, tier 2). Owner tags / environment events: an honest-behaviour premise exists in the trust base but no explicit environment-event or move-tag primitive was found, `[INFERENCE: absent]`. Embedding: not found, `[INFERENCE: absent]`. | Agda + Typst, Haskell MAlonzo extraction. Apache-2.0. Spec dir last commit 2026-11-04, actively maintained. | Design only: the literate spec, extraction-plus-differential-testing pipeline, and especially the trust-ledger gate (`check-trust-ledger.sh`) that names every postulate/mock and fails the build on drift are transferable. No EVM/Solidity code. |
| **Lean-blaster + PlutusCoreBlaster + CardanoLedgerAPIBlaster** (IOG, Lean 4) | Carried as a `Prop` field on a typeclass, not in the state type. `StateMachine α β` has `init`, `next`, `assumptions`, `invariants : α → β → Prop`. (`Blaster/StateMachine/StateMachine.lean`; README "State-Machine Formalization", verified live 2026-11-20) | Yes, from one declaration. `#kind` performs k-induction (base case + SMT-checked step case) over the whole transition relation; `#bmc` performs bounded counterexample search. This covers the whole relation in one command, unlike Armada's one-obligation-per-call design. (`Blaster/StateMachine/KInduction.lean`; README "K-Induction", "Bounded Model Checking") | Yes, the strongest match found. `PlutusCoreBlaster` formalizes UPLC terms and two CEK machines (step-limited, cost-model-integrated) over the real AST. `#import_uplc`/`#prep_uplc` load a *compiled* `.flat`/CBOR validator from Plinth, Aiken, or Plutarch and produce matched `prop`/`exec` pairs for the `blaster` tactic. The CEK termination proof for the budgeted evaluator is listed "pending"; some lemma proofs are "in progress". (`reservoir.lean-lang.org/@input-output-hk/PlutusCore`, "Libraries and modules", "Current scope"; `CardanoLedgerApiBlaster` README) | No. The pipeline is one-directional: source language → compiled UPLC → Lean import → proof. No client code is emitted. (`iog.io/news/automated-formal-verification-for-cardano-smart-contracts`) | Trusted primitives: the `Cryptograph` library implements and *conformance-tests* hash functions and EC signatures against NIST vectors; this is execution-checked, not postulated-with-proved-law, the opposite choice from Hydra. (blog post, "What are the new Lean 4 formalizations?") Environment events, owner tags, embedding: none found in the three READMEs read in full. `[INFERENCE: absent, not confirmed by an explicit negative statement]` | Lean 4, pinned v4.24.0 across all three repos (verified live). Apache-2.0 on all three. Actively developed: 441 commits on Lean-blaster (verified: 441, 62 stars), PlutusCoreBlaster/CardanoLedgerApiBlaster updated within a day of this survey per Reservoir. | Yes, in part, by design. Lean-blaster's README states it is "an application-agnostic automated theorem proving backend for Lean 4... with the Cardano application in mind." The `StateMachine`/`#bmc`/`#kind` machinery has no Cardano-specific type in its signature and is reusable for an EVM-targeted `machine` unmodified. Only `PlutusCoreBlaster` (UPLC/CEK) and `CardanoLedgerAPIBlaster` would need EVM-bytecode/Solidity-storage replacements. |
| **Djed** (stablecoin, Isabelle proof) | Proved once, over an inductive `transition` relation on a `bank-state` record. Not type-carried. (eprint 2021/1069, Appendix E) | No. Eight stability theorems (peg maintenance, no insolvency, no bank runs, etc.) are hand-proved once for the fixed protocol, not generated per declaration. (eprint 2021/1069 §3, Appendix E) | No. The paper models the bank algebraically; the Scala prototype and independent Solidity/ErgoScript/Plutus ports carry no refinement proof back to the theory. (eprint 2021/1069 §8; `github.com/input-output-hk/djed-stablecoin-prototype` README) | No. The theory, the Scala simulator, and each chain's port are unrelated artifacts. (eprint 2021/1069 §8) | None of the four. The only external input is an oracle exchange rate with no trust model attached. (eprint 2021/1069, Appendix E, `locale stablecoin`) | Isabelle/HOL, exists only as a PDF appendix, no public proof repo found. Scala prototype inactive since ~2021. No license found for the theory itself. | Design only: the bank-state/action/transition pattern and the stability-theorem list. The separate `DjedAlliance/Djed-Solidity` port is unverified and unconnected to the theory. |
| **agda2plinth** (Ferariu, Wadler, Melkonian, FMBC 2025) | Proved once, over an inductive relation `_⊢_~[_]~>_` with one constructor per contract action; a `ValidS` invariant is preserved by `validStateTransition`. (`MultiSig/Proofs.agda`) | No. The relation, invariant, and preservation theorem are hand-written per contract (`AccountSim`, `DEx`, `MultiSig` each separate); no single macro elaborates a declaration into them. (repository file listing) | Yes. `Model.agda` defines the real validator logic `agdaValidator`, compiled by agda2hs to Haskell/Plinth and onward to UPLC; the paper states it is "separately proven... functionally equivalent" to this validator, cross-checked with the Cardano Node Emulator. (FMBC 2025 paper, abstract; `MultiSig/Model.agda`) | No. Only the on-chain validator is generated. (README, file listing) | None of the four demonstrated. Transition rules check signatures but do not classify callers as honest/adversary/environment. (`MultiSig/Proofs.agda`) | Agda (agda-stdlib v2.2) + agda2hs v1.3. MIT. Last commit 2025-04-01; inactive academic prototype after that date. | Design only: model + hand-proved transition invariant + agda2hs-style compiler + emulator cross-check. agda2hs targets Haskell, not EVM bytecode. |

## The verdict

Cardano does not have the whole planned DSL. No candidate generates all three
declarations (`machine`, `contract`, `implements … via`) from a closed
surface, and no candidate supports the five Nitro features together.

Cardano has three separate parts of it, in three different projects, each
further ahead than the others on one axis:

1. **The `machine` axis.** Lean-blaster's `StateMachine α β` typeclass, with
   the `#kind` command, is the closest existing precedent for one declaration
   that states `init`, `next`, and `invariants` and generates a preservation
   obligation by automated proof. It differs from Armada's design in grain:
   `#kind` proves the whole relation at once, by k-induction and SMT, where
   Armada plans one dependent-`if` `Pre`/`Effect` obligation per call, proved
   by a human-written Lean term. (`Blaster/StateMachine/StateMachine.lean`,
   `Blaster/StateMachine/KInduction.lean`)
2. **The `contract`-to-bytecode axis.** `PlutusCoreBlaster` is the closest
   match to Armada's plan to prove the deployed UPLC or bytecode, not just a
   whole-language metatheory. It imports the actual compiled artifact and
   proves properties against it directly, which is Armada's own target for
   its bytecode theorem. `plutus-metatheory`'s `VerifiedCompilation`/
   `--certify` is the second-closest match: a per-compiler-pass proof against
   a mechanized target semantics, the same shape as Armada's per-
   compiler-profile bytecode theorem, but a proof about compiler
   transformations, not about one program's refinement from a model.
3. **The on-chain-plus-client axis.** Hydra's Agda spec is the only candidate
   found that proves, not merely asserts, a correspondence between one
   formal model and both the on-chain validator and the off-chain
   (`hydra-node`) handler, through `Hydra.OffChainAgreementSpec`. This is the
   closest existing precedent for Armada's "second target from the same
   `machine`" requirement, though Hydra's on-chain-to-bytecode link is a
   differential test, not a proof, and the spec has no embedding or
   owner-tagged-move primitive.

No candidate combines all three axes. None was found with `embeds` (one
machine's state as a field of another, changing only through the first
machine's calls) or with owner-tagged moves distinguishing honest, adversary,
and environment actors. These two Nitro-specific features appear in no
Cardano source read by any research agent.

## Design ideas worth copying, with their exact source

1. **The `StateMachine` typeclass plus `#kind` k-induction.** One
   declaration (`init`, `next`, `assumptions`, `invariants`) yields a
   base-case and step-case obligation, discharged by SMT, with
   counterexample generation on failure. Source:
   `github.com/input-output-hk/Lean-blaster`,
   `Blaster/StateMachine/KInduction.lean`; README "State-Machine
   Formalization" and "K-Induction".
2. **Proving against the compiled artifact, not a model of it.**
   `PlutusCoreBlaster`'s `#prep_uplc` imports the actual compiled `.flat`/
   CBOR UPLC program and produces matched proof-facing and execution-facing
   terms for the same program. Source: `CardanoLedgerApiBlaster` README,
   "Preparing the validator for verification";
   `reservoir.lean-lang.org/@input-output-hk/PlutusCore`.
3. **A certifying compiler checked per optimization pass against a
   mechanized target semantics.** `plutus-metatheory`'s
   `VerifiedCompilation`/`--certify` is the closest shape to Armada's
   per-compiler-profile bytecode theorem, applied to a compiler's
   transformations rather than to one refinement. Source:
   `plutus.cardano.intersectmbo.org/metatheory/latest`, "Certified
   Compilation".
4. **A trust-ledger gate that names every unproved assumption.** Hydra's
   `spec/check-trust-ledger.sh` enumerates every postulate and every mock by
   name and fails the build when the list drifts, forcing an explicit,
   reviewed inventory of what is assumed instead of proved. This directly
   suits the gap between Armada's `implements … via` proof and its bytecode
   theorem. Source: `cardano-scaling/hydra`,
   `docs/dev/agda.md`, "The trust story" tier 3.
5. **Deriving ledger validation rules from the node's actual
   implementation, not from a separate specification.**
   `CardanoLedgerAPIBlaster` re-derives validation predicates from the code
   the node runs, stated reason: "some design choices made there enable
   script optimizations" that a separate spec would miss. Source: IOG blog
   post, `iog.io/news/automated-formal-verification-for-cardano-smart-contracts`
   (2026-05-12).

## What this changes for Armada

Lean-blaster is reusable now, not only as a design to copy. Its README
states it is application-agnostic, and its `StateMachine`/`#bmc`/`#kind`
commands carry no Cardano-specific type in their signature. Armada's
`machine` declaration could depend on Lean-blaster's typeclass and
k-induction command for its invariant-preservation obligation, instead of
building that automation from a blank page. The `contract`-to-bytecode step
has no reusable code (PlutusCoreBlaster and plutus-metatheory are both
UPLC-specific), so Armada's plan to pair its `contract` declaration with
Blanc for the EVM target stays the right route; only the general shape
(proving against the compiled artifact; a certifying-compiler pattern per
profile) transfers.
</content>
