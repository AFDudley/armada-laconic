# Lean as the one model, with a tool per target (2026-11-02)

Operator order, 2026-11-02: document the suggestion that Lean is the unified modelling language, and that tools such
as EquiVM and Bend2 carry the proofs to executable code where they are needed. This entry records the suggestion,
where it came from, which parts the plans already decide, and which parts are open. It decides nothing new.

## 1. Source

The suggestion comes from a claude.ai chat that the operator shared (`.scratch/claude-share-46520279.clean.txt`,
gitignored):

- lines 658-686: "a single chain" papers → Lean → Bend2 → Solidity is rejected; "unify the spec layer and branch to
  the targets from there". The chat's diagram (lines 671-677):

  ```
  papers ──► Lean: specs + reference model (single source of truth)
                │
                ├─► EVM:      EquiVM against solc output
                ├─► ZK:       circuits proven against the same model
                └─► app logic: Bend2 (JS, Python, C) or a Lean-fragment → TypeScript emitter,
                               checked against the model by differential tests
  ```
  (The chat also listed Verity and ProofForge for EVM. Verity is rejected, `engineering/build-plan.md` S5.)
- lines 702-730: applied to Nitro's actors on several platforms. The state types, encodings, transition rules and
  per-actor state machines must agree byte for byte between the contract and every client; networking, storage,
  keys and UI are per-platform code outside the model.
- lines 768 and 857-875: the client side needs an executable honest strategy, written in a portable fragment of Lean,
  proved against the strategy theorems, ported, and checked with conformance vectors.

The chat's claims about external tools were not all checked here. Each claim below that this repository does not
already verify carries `[INFERENCE]`.

## 2. The rule

1. **One model.** The Lean model in `AFDudley/armada-lean` is the specification of every component that must agree
   with another: contracts, circuits, and the shared logic of clients. The current Solidity is not a source
   (operator ruling 2026-11-02, `labbook/2026-11-02-refinement-risks.md` §7).
2. **One route per target, chosen by what can be checked.**
   - A proof where a tool connects the target to the model: EquiVM for bytecode; an R1CS import for circuits.
   - Conformance vectors generated from the model where no such tool exists: TypeScript, Bend2 output, go-nitro.
3. **No unverified translator in the trusted path.** A port is trusted only through the check of its output, never
   through the translator that produced it. Proofs are not ported; they stay with the Lean model.

## 3. Targets

| Target | Route | Check | Status |
|---|---|---|---|
| Contracts (EVM) | Lean → Stele → Solidity → solc → bytecode | step-1 refinement proof; EquiVM proof of the bytecode | decided (build-plan S5, operator rulings 2026-10-25, 2026-10-27); toolchain in a new monorepo (`labbook/2026-11-02-equivm-toolchain.md` §7) |
| Circuits (circom) | hand-written circom, R1CS imported into Lean | proof that the R1CS equals the model's constraint predicate | decided (refinement-risks §7, Q4.1); not scheduled |
| Deployed keys | ceremony zkey → deployed pool key | equality check | decided (refinement-risks §7, Q4.2; P3 and S13 acceptance) |
| Client shared logic (wallet, relayer, watchtower, hub) | open (§4) | conformance vectors from the model | vectors decided (refinement-risks §7, Q3.1); route open |
| UI, networking, storage, keys | per-platform code | e2e gates | outside the model |

## 4. What fits and what is open

**Bend2.** `labbook/2026-10-27-bend2-to-evm.md` rejects Bend2 as a route to verified EVM bytecode. That finding
stands. This entry gives Bend2 a different role: a client-side target for shared logic. Facts from the Bend2 report:
its compiler emits C and JavaScript; a Python backend exists on the fork `AFDudley/bend`; it has no tactics, so every
proof is an explicit term. Not checked here: whether its numeric types (the chat names `Nat`, `U32`, `F32`) can carry
`uint256` amounts and field elements, and how a C build links into iOS and Android `[INFERENCE]`.

**The TypeScript client decision.** `labbook/2026-10-19-unified-stack-and-verified-note-sync.md` and armada-client
`docs/SPEC.md` choose one TypeScript codebase for wallet, relayer and watchtower, with `ts-nitro` over Waku. That
decision is about the runtime. It does not say where the shared protocol logic comes from. This entry does not change
it.

**Open questions (technical; decided by us under the operator's standing instruction when the client work starts):**

1. Where the client's shared logic comes from:
   - (a) written by hand in TypeScript (`ts-nitro`), checked by the conformance vectors only;
   - (b) emitted from a portable fragment of Lean to TypeScript;
   - (c) written in Bend2, with Bend2 laws generated from the Lean statements, compiled to JavaScript.
2. The portable fragment, if (b) or (c): which Lean constructs a ported definition may use, enforced by a linter in
   `ArmadaLean/Lint/` next to `TypeDoctrine` `[INFERENCE: not designed]`.
3. The order. The conformance vectors come first: they also test the representation map of the step-1 refinement
   against the real contracts. The executable honest strategy comes after the Liveness proofs close (refinement-risks
   §7, Q3.1).

## 5. Where this entry is cited

- `engineering/build-plan.md` S5, "Alternatives considered and rejected", the Bend2 bullet.
- `docs/process.md` "The layers".
- armada-lean `docs/verification-chain.md` states the contract chain only; a pointer to this entry is for the
  armada-lean coordinator (not changed here).
