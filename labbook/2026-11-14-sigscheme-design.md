# Design: law-bundled primitives and the `ArithConstrained` split

Date: 2026-11-14. Session `01a195d3`.
Related: `labbook/2026-11-14-trusted-primitives-and-witnesses.md` (the problem and the
decision); review item 3 on ship-armada/armada-lean#1 (coax1d, 2026-11-09).

This document designs the Lean text for option 2 of the related entry: each primitive
becomes a parameter, bundled with its laws; the symbolic model (`Term`,
`ArmadaLean/Poseidon.lean`) is the one instance every witness uses. It does not edit
armada-lean. Every statement below is checked against armada-lean `integrated`
`f753341b` (`git show integrated:<path>`), either by reading the file or by a build in
a throwaway worktree under `armada-lean/.scratch/`.

## 1. The three primitives today, and every use site

Three primitives are `axiom`s with no defining equation:

- `EdDSAVerify : (Nat × Nat) → (Nat × Nat × Nat) → Term → Prop` — `ArmadaLean/Transact.lean:64`.
  Holds iff a signature verifies under a public key over a message term.
- `VerifyAccepts : {N M : Nat} → TransactPublic N M → Prop` — `ArmadaLean/PoolSafety.lean:677`.
  Holds iff some Groth16 proof over the plain circuit's public inputs verifies.
- `VerifyOwnedAccepts : {N M : Nat} → TransactOwnedPublic N M → Prop` — `ArmadaLean/PoolOwned.lean:132`.
  The same, for the owned circuit.

`git grep` finds every use site (`integrated`):

**`EdDSAVerify`**
- `ArmadaLean/Transact.lean:64` — the `axiom` itself.
- `ArmadaLean/Transact.lean:173` — `Constrained.sig`, the field this design splits out.
- `ArmadaLean/PoolSafety.lean:226` — `transact_spends_only_owned`'s conclusion restates it.
- `ArmadaLean/PoolSafety.lean:717-719` — `HonestSignature`, the existing completeness law
  (`sign`, `signSound`). This design folds `HonestSignature` into the new `SigScheme`.
- `ArmadaLean/PoolNonVacuity.lean:8-15` (module comment) — states directly that the
  plain path's fixtures are impossible today because `EdDSAVerify` is an opaque axiom.
- `ArmadaLean/Witness/PoolSafety.lean:13-14,45-46,68-69` (doc comments on three `sorry`
  witnesses) — name `EdDSAVerify` as the reason each witness is blocked.
- `scripts/oracles/pool_exec/PE1-PE5.lean` (five files) — each has one line,
  `` ``EdDSAVerify, ``VerifyAccepts, ``VerifyOwnedAccepts] ``, a `Name` literal in an
  axiom allowlist for a developer probe (`lean_probe.py`), not part of `scripts/check.sh`.
  `PE3.lean` additionally has two `example`s whose *statement* names `VerifyAccepts`/
  `VerifyOwnedAccepts` directly (quoted below, §3).

**`VerifyAccepts`**
- `ArmadaLean/PoolSafety.lean:677` — the `axiom`.
- `ArmadaLean/PoolSafety.lean:683-685` — `TransactProofCompleteness.complete`.
- `ArmadaLean/PoolSafety.lean:694-696` — `TransactProofSoundness.sound`.
- `ArmadaLean/PoolSafety.lean:705-707` — `Pool.Verifier.Realizes.transact`.
- `ArmadaLean/PoolSafety.lean:753` — `transact_spend_can_happen`'s conclusion.
- `ArmadaLean/Pool.lean:1110,1374` (doc comments) — note that `Pool.Verifier`
  (the *executable*, `Bool`-valued pair, `Pool.lean:1408`) is a separate, already-correct
  parameter; `VerifyAccepts` is the Prop-level primitive `Verifier.Realizes` ties to it.
- `ArmadaLean/Spec/PoolSafety.lean:321` (generated spec mirror of the theorem statement).
- `ArmadaLean/Witness/PoolSafety.lean:68-69,153-155,162-163` (doc comments on three
  `sorry` witnesses).
- `scripts/oracles/pool_exec/PE1-PE5.lean` (allowlist + two `PE3.lean` examples, as above).

**`VerifyOwnedAccepts`** — the same four roles, in `ArmadaLean/PoolOwned.lean` (the
`axiom`, lines 132; `OwnedProofCompleteness.complete`, 143-145; `OwnedProofSoundness.sound`,
157-159), `ArmadaLean/PoolSafety.lean:676,709` (doc comment, `Verifier.Realizes.owned`),
`ArmadaLean/PoolOwned.lean:387` (`owned_spend_can_happen`'s conclusion),
`ArmadaLean/Spec/PoolOwned.lean:107`, `ArmadaLean/Witness/PoolOwned.lean:28`,
`ArmadaLean/Witness/PoolSafety.lean:94-96`, and the five `PE*.lean` oracles.

None of the three names appears outside `ArmadaLean/Transact.lean`, `PoolSafety.lean`,
`PoolOwned.lean`, `TransactOwned.lean` (doc-comment cross-references only), the two
generated mirrors (`Spec/`, `Witness/`), and `scripts/oracles/pool_exec/`.

## 2. The law structures

### 2a. The signature scheme

The circuit's own wire shapes fix the carrier types: a public key is `Nat × Nat`
(`[x, y]` on BabyJubJub, `TransactWitness.publicKey`, `Transact.lean:97`); a signature
is `Nat × Nat × Nat` (`[R_x, R_y, S]`, `TransactWitness.signature`, `Transact.lean:99`);
a message is a `Term` (`msgHashOf`'s output). This design keeps both types exactly as
they are: it does not touch `TransactWitness`.

```lean
/-- A signature scheme over the circuit's own wire types. An explicit structure
    parameter, not a typeclass: §2c gives the reason. -/
structure SigScheme where
  /-- The signature a holder of `pk` computes over `msg` (Transact.circom:177-186).
      Replaces `HonestSignature.sign` (`PoolSafety.lean:718`), unchanged in shape. -/
  sign : (Nat × Nat) → Term → (Nat × Nat × Nat)
  /-- `verify pk sig msg` holds iff `sig` is a valid EdDSA-Poseidon signature on `msg`
      under public key `pk` (Transact.circom:177-186, circomlib eddsaposeidon.circom).
      Replaces `axiom EdDSAVerify` (`Transact.lean:64`): the same type, now a value a
      theorem receives as data, not a name it assumes. -/
  verify : (Nat × Nat) → (Nat × Nat × Nat) → Term → Prop
  /-- COMPLETENESS: an honest signer's own signature verifies under its own key.
      Replaces `HonestSignature.signSound` (`PoolSafety.lean:719`), unchanged in
      shape. The real scheme has this exactly: EdDSA's signing algorithm is correct
      by construction (circomlib's `EdDSAPoseidonVerifier` accepts every signature
      `EdDSAPoseidonSigner` produces). -/
  complete : ∀ (pk : Nat × Nat) (msg : Term), verify pk (sign pk msg) msg
```

**No soundness/unforgeability field.** This is a deliberate omission, not an oversight;
see §2b for why "soundness" is essential for the *other* primitive and §9 for the rule
8.3 framing. Three reasons:

1. **Measured, not guessed (§6): no blocked theorem's proof uses an unforgeability
   direction.** `transact_spends_only_owned` (the one theorem whose conclusion names
   `EdDSAVerify`) proves it by `⟨commitment_mpk_inj hSameNote, hC.sig⟩` — it restates
   "the signature verifies," never "therefore the signer held the key." No other
   theorem projects `.sig` from a `Constrained` value at all.
2. **A faithful law needs a secret-key type separate from the public key**, e.g.
   `SK : Type`, `pkOf : SK → Nat × Nat`, with soundness `verify pk sig msg → ∃ sk,
   pkOf sk = pk ∧ sig = sign' sk msg` for a signing function keyed by `sk`, not by
   `pk`. The existing `complete`/`sign` (and every can-happen theorem's call site,
   §6) are keyed directly by `pk` — "the real key holder" already means "whoever
   presents this `pk` value," the same simplification `HonestSignature` makes today.
   Adding `SK`/`pkOf` only for an unused law would carry two different, uncoupled
   notions of "who may sign" in one structure, for no proof gain.
3. `sig = sign pk msg` (determinism of the *accepted* signature, the only law that
   does not need `SK`) **is not true of real EdDSA/Schnorr-style signatures**: a
   signature equation of this shape accepts any `(R, s)` pair satisfying the curve
   relation, for any nonce the signer chose, not only the deterministic one EdDSA's
   own nonce derivation happens to produce. Stating it would repeat the mistake the
   Poseidon rewrite fixed — an axiom false of the function it names
   (`ArmadaLean/Poseidon.lean`'s module comment).

If a future theorem needs unforgeability, `SigScheme` can be extended with `SK`,
`pkOf`, and a soundness field then, with the `SK`-keyed signing function introduced
alongside (not in place of) `sign`. That extension would not touch any of the 14
theorems below, because none of them take the new fields.

### 2b. The proof verifiers (plain and owned circuits)

`VerifyAccepts`/`VerifyOwnedAccepts` already have law structures
(`TransactProofCompleteness`/`TransactProofSoundness`/`OwnedProofCompleteness`/
`OwnedProofSoundness`, `PoolSafety.lean:683-696`, `PoolOwned.lean:143-159`) — option 2's
recipe is already half-built here. The only change is to stop naming the primitive by a
global axiom and instead take it as an explicit parameter, the same name, now bound:

```lean
/-- Groth16 completeness for the plain circuit (replaces `PoolSafety.lean:683-685`):
    a witness/public pair satisfying every `Transact.circom` constraint, signature
    included, has an accepted proof. Parameterized by `S` (`Constrained S w pub`
    needs it, §5) and by `accepts` (replaces `axiom VerifyAccepts`,
    `PoolSafety.lean:677`): a value in scope, not a global name. -/
structure TransactProofCompleteness (S : SigScheme)
    (accepts : {N M : Nat} → TransactPublic N M → Prop) where
  complete : ∀ {N M : Nat} {w : TransactWitness N M} {pub : TransactPublic N M},
    Constrained S w pub → accepts pub

/-- Groth16 knowledge soundness over the deployed verifying key (replaces
    `PoolSafety.lean:694-696`): every public input `accepts` accepts has a witness
    that satisfies every constraint. The real scheme has this only computationally:
    Groth16 knowledge soundness is an extractor argument under a knowledge-of-exponent
    assumption, not a fact Lean derives from the BN254 pairing. -/
structure TransactProofSoundness (S : SigScheme)
    (accepts : {N M : Nat} → TransactPublic N M → Prop) where
  sound : ∀ {N M : Nat} {pub : TransactPublic N M},
    accepts pub → ∃ w : TransactWitness N M, Constrained S w pub

/-- The owned-circuit pair (replaces `PoolOwned.lean:143-145,157-159`). No `S`:
    `OwnedConstrained` carries no primitive (§5). -/
structure OwnedProofCompleteness
    (acceptsOwned : {N M : Nat} → TransactOwnedPublic N M → Prop) where
  complete : ∀ {N M : Nat} {w : TransactOwnedWitness N M} {pub : TransactOwnedPublic N M},
    OwnedConstrained w pub → acceptsOwned pub

structure OwnedProofSoundness
    (acceptsOwned : {N M : Nat} → TransactOwnedPublic N M → Prop) where
  sound : ∀ {N M : Nat} {pub : TransactOwnedPublic N M},
    acceptsOwned pub → ∃ w : TransactOwnedWitness N M, OwnedConstrained w pub
```

`accepts`/`acceptsOwned` stay bare, universally-quantified `Prop`-valued parameters —
exactly `VerifyAccepts`/`VerifyOwnedAccepts`'s own shape, now a binder instead of an
axiom name. A bare parameter alone is the reviewer's rejected proposal
(`labbook/2026-11-14-trusted-primitives-and-witnesses.md`, option 2's last sentence);
here it is never used alone — every theorem that takes `accepts` also takes whichever
of `TransactProofCompleteness`/`TransactProofSoundness` its proof needs, so the law
always travels with the parameter.

**Completeness is exact; soundness is computational**, same split as the signature
scheme: an honest prover with a satisfying witness always gets an accepted proof
(Groth16 is complete by construction); a verifier that accepts only has a satisfying
witness is Groth16's knowledge soundness, a cryptographic (extractor-based) argument,
not a derivable fact.

### 2c. Why a structure parameter, not a typeclass

A typeclass resolves to **one** instance per type, found by search. This project
needs two or more `SigScheme`/verifier values in scope at once in the general case:
the ABSTRACT theorem (quantified over every law-satisfying scheme) and the SYMBOLIC
INSTANCE (§4) that witnesses it are two different values of the same type, and a
theorem that composes two circuit-level facts about *different* schemes (e.g., a
future cross-chain proof) would need to name each one. A typeclass instance is found
by unification, not named; an explicit structure parameter is always nameable, passed
and threaded exactly like `Pool.Verifier` already is (`Pool.lean:1408`, rule PE3) — the
project already has this precedent for exactly this kind of primitive.

## 3. Mapping the existing verifier hypotheses onto the laws

- **`Pool.Verifier`** (`Pool.lean:1408`) is untouched. It is the *executable*,
  `Bool`-valued pair (`PrivacyPool.verify`/`verifyOwned`), already a genuine parameter,
  not an axiom (rule PE3). This design does not touch it.
- **`Pool.Verifier.Realizes`** (`PoolSafety.lean:705-709`) ties `Pool.Verifier`'s
  executable check to the Prop-level `accepts`/`acceptsOwned`. It needs `accepts`/
  `acceptsOwned` as explicit parameters (it does not need `S`: its own fields never
  mention `Constrained`):

  ```lean
  structure Pool.Verifier.Realizes (verify : Pool.Verifier)
      (accepts : {N M : Nat} → TransactPublic N M → Prop)
      (acceptsOwned : {N M : Nat} → TransactOwnedPublic N M → Prop) : Prop where
    transact : ∀ {N M : Nat} (pub : TransactPublic N M),
      accepts pub ↔ ∃ proof, verify.transact pub proof = true
    owned : ∀ {N M : Nat} (pub : TransactOwnedPublic N M),
      acceptsOwned pub ↔ ∃ proof, verify.owned pub proof = true
  ```

- **`TransactProofSoundness`/`OwnedProofSoundness`** — §2b; unchanged shape, now
  parameterized.
- **No duplication remains.** `HonestSignature` (`PoolSafety.lean:717-719`) is deleted:
  its two fields are `SigScheme.sign`/`SigScheme.complete`, unchanged in shape. Every
  other law (`TransactProofCompleteness`, `TransactProofSoundness`,
  `OwnedProofCompleteness`, `OwnedProofSoundness`) already existed as exactly one
  structure each; this design adds no second copy of any of them.

`scripts/oracles/pool_exec/PE3.lean` has two `example`s that name `VerifyAccepts`/
`VerifyOwnedAccepts` directly in their *statement* (not just the allowlist literal):

```lean
example (verify : Pool.Verifier) (h : verify.Realizes) {N M : Nat} (pub : TransactPublic N M) :
    VerifyAccepts pub ↔ ∃ proof, verify.transact pub proof = true :=
  h.transact pub
```

After this change `verify.Realizes` is `verify.Realizes accepts acceptsOwned`; the two
examples become `(h : verify.Realizes accepts acceptsOwned)` with `accepts pub ↔ …`.
The five `PE1-PE5.lean` oracles' allowlist literal
`` ``EdDSAVerify, ``VerifyAccepts, ``VerifyOwnedAccepts] `` must drop all three names:
once they are not axioms, `` ``EdDSAVerify `` does not resolve (compile error), and
`Pool.apply`/its four equations (what `PE1-PE5` actually probe) never depended on any
of the three anyway — they are pure, computable functions of `Pool.Verifier` (the
`Bool` pair) alone, never of `accepts`/`S`.

## 4. The symbolic instance

### 4a. `SigScheme`

```lean
/-- The symbolic (Dolev-Yao-style) signature scheme: the instance every witness in
    this project uses (§7). `sign`'s message argument plays no role in either law
    below (neither constrains how `sign` depends on `msg`); this instance's third
    signature component is a fixed sentinel for every message, which is sufficient.
    It is a consistency witness for `SigScheme`'s laws, not a claim about the
    concrete arithmetic — the same role `Term`'s free algebra plays for Poseidon. -/
def symbolicSigScheme : SigScheme where
  sign := fun pk _msg => (pk.1, pk.2, 0)
  verify := fun pk sig msg => sig = (pk.1, pk.2, 0) ∧ True  -- i.e. `sig = sign pk msg`
  complete := fun _ _ => ⟨rfl, trivial⟩
```

(Written out: `verify pk sig msg := sig = sign pk msg`, with `complete` by `rfl`. The
inline `∧ True` above is only to keep `msg` visibly present in `verify`'s statement for
documentation; the implementer may drop it and write `verify pk sig msg := sig = sign pk msg`
directly, which is definitionally the same thing and the cleaner final text.)

### 4b. The proof verifiers

Option 1's description — "a proof term carries its witness; `VerifyAccepts` holds
exactly when that witness satisfies the circuit" — becomes, literally:

```lean
def symbolicTransactCompleteness (S : SigScheme) :
    TransactProofCompleteness S (fun pub => ∃ w, Constrained S w pub) where
  complete := fun hC => ⟨_, hC⟩

def symbolicTransactSoundness (S : SigScheme) :
    TransactProofSoundness S (fun pub => ∃ w, Constrained S w pub) where
  sound := fun h => h

def symbolicOwnedCompleteness :
    OwnedProofCompleteness (fun pub => ∃ w, OwnedConstrained w pub) where
  complete := fun hC => ⟨_, hC⟩

def symbolicOwnedSoundness :
    OwnedProofSoundness (fun pub => ∃ w, OwnedConstrained w pub) where
  sound := fun h => h
```

`accepts := fun pub => ∃ w, Constrained S w pub` makes both completeness and
soundness hold **by definition** — `sound` is the identity function. This is not
circular: it is the Dolev-Yao-style idealization of knowledge soundness the lab book
names ("a proof term carries its witness"), the same move the free algebra makes for
Poseidon (declare the idealized model, do not claim it matches the real arithmetic).
`Pool.Verifier` (the separate, executable, `Bool`-valued parameter, §3) is untouched —
nothing above gives it a symbolic instance, because nothing in this redesign needs
one: `Pool.apply`'s own correctness never depended on `VerifyAccepts`/`EdDSAVerify`.

Together, `symbolicSigScheme`, `symbolicTransactCompleteness`,
`symbolicTransactSoundness`, `symbolicOwnedCompleteness`, `symbolicOwnedSoundness`
exhibit one value of every law this design states. The laws are consistent.

## 5. The `ArithConstrained` split

`Constrained` (`Transact.lean:142-173`) splits in two:

```lean
/-- Every Transact.circom constraint except the signature check. Replaces fields
    2c/2d/3/4/4b of `Constrained` (`Transact.lean:150-171`), verbatim. -/
structure ArithConstrained {N M : Nat} (w : TransactWitness N M)
    (pub : TransactPublic N M) : Prop where
  merkle : ∀ i,
    MerkleProof 16
      (commitment (npk (masterPublicKey w.publicKey w.nullifyingKey) (w.randomIn i))
        w.token (w.valueIn i))
      (w.pathElements i) (w.leavesIndices i) pub.merkleRoot
  nullifiers : ∀ i, pub.nullifiers i = nullifier (Term.val w.nullifyingKey) (w.leavesIndices i)
  outputs : ∀ j, pub.commitmentsOut j = commitment (w.npkOut j) w.token (w.valueOut j)
  rangeIn : ∀ i, RangeChecked120 (w.valueIn i)
  rangeOut : ∀ j, RangeChecked120 (w.valueOut j)
  balance : fieldEq (listSum (List.ofFn w.valueIn)) (listSum (List.ofFn w.valueOut))
  distinct : ∀ i j, i ≠ j → w.leavesIndices i ≠ w.leavesIndices j

/-- The full constraint system (field 5-6 of `Constrained`, `Transact.lean:172-173`,
    restored): `ArithConstrained` plus the signature check, now a value of `S.verify`
    rather than the axiom `EdDSAVerify`. -/
structure Constrained (S : SigScheme) {N M : Nat} (w : TransactWitness N M)
    (pub : TransactPublic N M) extends ArithConstrained w pub : Prop where
  sig : S.verify w.publicKey w.signature (msgHashOf pub)
```

Lean 4's `extends` makes every `ArithConstrained` field (`.merkle`, `.nullifiers`, …)
reachable by dot notation on a `Constrained S w pub` value, unchanged; it adds one
auto-generated projection, `.toArithConstrained`, needed at any call site that must
pass a `Constrained S w pub` value where an `ArithConstrained w pub` is expected
(`extends` gives no automatic coercion on application — confirmed by the build in §6).
The anonymous-constructor syntax `{ merkle := …, …, sig := … }` is unchanged.

**`OwnedConstrained` (`TransactOwned.lean:158-181`) does not split.** Its fields are
`merkle`, `nullifiers`, `outputs`, `rangeIn`, `rangeOut`, `balance` — the owned circuit
carries no signature witness (`TransactOwned.lean:30-38`, module comment: "this circuit
carries no signature witness and proves no `sig` field"). `OwnedConstrained` already
carries no primitive and needs no `S`.

### 5a. The unavoidable ripple: `Step`, `Reachable`, `Pool.machine`

`Step.transact` (`Pool.lean:1269-1271`) takes `hC : Constrained w pub` **as one of its
own constructor's arguments** — it is not only a theorem hypothesis, it is part of the
pool state machine's own definition of what a valid transition is. After the split,
`Step.transact`'s `hC` has type `Constrained S w pub`, so `Step` itself needs `S` in
scope (`Step (S : SigScheme) : PoolState → PoolState → Prop`), and therefore so does
`Reachable` (`Pool.lean`, the inductive closure of `Step` from `genesis`) and
`Pool.machine` (`PoolMachine.lean:14-16`, `step := Step`): `Pool.machine (S : SigScheme)
: Machine PoolState`.

`Step.transact`'s `hC` **must** keep the full `Constrained S w pub`, not
`ArithConstrained w pub`: dropping the signature requirement from the model's own
notion of a valid transition would make `Step` accept a transaction with no verified
signature at all, which is false of the real contract and would make
`transact_spends_only_owned`'s safety argument ("no caller without the private key can
build a witness that opens this note") vacuous rather than proved.

This means `S : SigScheme` is a parameter not only of the 14 blocked theorems, but,
mechanically, of **every** declaration whose statement mentions `Pool.machine`,
`ReachableState Pool.machine`, `Step`, or `Reachable` — because `Pool.machine` is no
longer a bare term once it takes an argument. A literal count in `integrated`: 62 of
`PoolSafety.lean`'s 76 `theorem`s mention one of those four names in their own type
(`grep -c`); so do a smaller fraction of `Pool.lean`'s 104 declarations (`Step`,
`Reachable`, `Pool.machine` themselves, plus `Pool.apply_*_equation`'s doc comments —
`Pool.apply` itself takes no `Step`/`Reachable` argument and is unaffected). Most of
these ~62 theorems' *proofs* never use `S`'s laws — they only need `S` present to form
`Pool.machine S`'s type, the same way most of `Pool.lean`'s theorems already carry
`Pool.Verifier` without inspecting its fields.

**Recommended mechanism: a file-wide `variable`, not per-theorem copies.** Add
`variable (S : SigScheme)` once, in a `section`, near the top of `Pool.lean` (after
importing `ArmadaLean.Transact`, where `SigScheme` is declared). Lean 4 auto-inserts
`S` as the first explicit argument of every subsequent declaration in the section whose
statement mentions it — the standard idiom for a parameter threaded through a whole
file's definitions, and the only one that does not require editing all ~62 signatures
by hand. Every **call site** across `PoolSafety.lean`/`PoolOwned.lean`/
`PoolNonVacuity.lean` that already names `Pool.machine`, `Step.transact`, or
`ReachableState Pool.machine` needs one extra explicit argument, `S`, supplied from
whichever binder introduced it in that theorem's own signature. This is mechanical,
not conceptual: no theorem's *proof strategy* changes, only its argument list.

This is the single largest consequence of closing review item 3, and the lab book
entry's framing — "each theorem quantifies over every scheme that satisfies the laws
it uses" — undersells it: `Step.transact` bakes `Constrained` into the state machine
itself, so the threading is not confined to theorems that mention a primitive in their
own hypotheses. §12 repeats this as the headline point where the code needs more than
the lab book entry anticipated.

## 6. Per-theorem redesign

### 6a. The six arithmetic-family theorems: the `.sig` measurement

Measured (not guessed), in `armada-lean/.scratch/sigscheme-sig-measure` (a throwaway
worktree at `f753341b`, `git worktree add --detach`, `.lake` symlinked to the main
checkout's build cache). Method: split `Constrained`/`ArithConstrained` for real in the
worktree (§5's exact text); narrow each candidate theorem's own hypothesis type from
`Constrained w pub` to `ArithConstrained w pub`, one at a time; `lake build`; read the
error, if any, at that theorem's own line.

| Theorem | file:line | Narrows to `ArithConstrained`? | Evidence |
|---|---|---|---|
| `transact_value_conserved` | `Transact.lean:217` | Yes | No error at the theorem. (Matches the prior finding, armada-lean `aln-c67`.) |
| `transact_inputs_distinct_nullifiers` | `Transact.lean:201` | Yes | No error at the theorem. |
| `transact_public_outputs_determined` | `Transact.lean:183` | Yes | No error at the theorem. |
| `transact_not_blocked` | `PoolSafety.lean:2326` | Yes | No error at the theorem; its proof uses only `hC.nullifiers i`. |
| `transact_spends_only_owned` | `PoolSafety.lean:218` | **No** | `error: PoolSafety.lean:227:36: Invalid field 'sig': The environment does not contain 'ArmadaLean.ArithConstrained.sig' … hC of type 'ArithConstrained w pub'`. Its own proof is `⟨commitment_mpk_inj hSameNote, hC.sig⟩` — the conclusion itself names `EdDSAVerify`/`S.verify`. |
| `transact_unshield_exact` | `PoolSafety.lean:429` | **No** | `error: PoolSafety.lean:446:52: Application type mismatch: … hC has type ArithConstrained w pub but is expected to have type Constrained ?m.91 ?m.92` — its proof builds `Step.transact caller w pub bp hN hM hC …` (`PoolSafety.lean:449`), and `Step.transact` requires the full bundle (§5a), not because this theorem's own reasoning touches `.sig`. |

(The remaining four hits in the first run of this build — one `Spec/Transact.lean`
definitional-equality error per narrowed theorem, and one `transact_balance` call-site
mismatch at `PoolSafety.lean:64` — are the mechanical consequences of §5a/§8, not
evidence about `.sig`: `Spec/Transact.lean`'s mirror definitions were not updated to
match in the measurement, and `transact_balance` — a helper, not one of the 14 — still
took the un-split `Constrained` and needed `hC.toArithConstrained` at its one call
site. Both are expected, and both are cleared before reading the final four errors
above.)

**Conclusion.** Four of six take `ArithConstrained`, no `S`: `transact_value_conserved`,
`transact_inputs_distinct_nullifiers`, `transact_public_outputs_determined`,
`transact_not_blocked`. Two keep the full `Constrained S`:
`transact_spends_only_owned` (names `EdDSAVerify`/`S.verify` in its own conclusion) and
`transact_unshield_exact` (must build a `Step.transact` term, which always needs the
full bundle, §5a).

### 6b. New statements

For the four that narrow, only the hypothesis's type changes (`Constrained w pub` →
`ArithConstrained w pub`); the bound variable name (`h`/`hC`) and the proof are
unchanged:

```lean
theorem transact_value_conserved {N M : Nat} (w : TransactWitness N M)
    (pub : TransactPublic N M) (hN : N ≤ 8) (hM : M ≤ 4)
    (h : ArithConstrained w pub) :
    listSum (List.ofFn w.valueIn) = listSum (List.ofFn w.valueOut)

theorem transact_inputs_distinct_nullifiers {N M : Nat} (w : TransactWitness N M)
    (pub : TransactPublic N M) (h : ArithConstrained w pub) :
    ∀ i j, i ≠ j → pub.nullifiers i ≠ pub.nullifiers j

theorem transact_public_outputs_determined {N M : Nat} (w : TransactWitness N M)
    (p₁ p₂ : TransactPublic N M) (h₁ : ArithConstrained w p₁) (h₂ : ArithConstrained w p₂) :
    (∀ i, p₁.nullifiers i = p₂.nullifiers i) ∧ (∀ j, p₁.commitmentsOut j = p₂.commitmentsOut j)

theorem transact_not_blocked (s : ReachableState (Pool.machine S)) {s' : PoolState}
    (hstep : Step S s.state s') {N M : Nat} {w : TransactWitness N M}
    {pub : TransactPublic N M} {bp : BoundParams} (hC : ArithConstrained w pub)
    (hEnabled : SpendEnabled s.state bp.treeNumber pub.merkleRoot w.token
      (transactSpentPositions bp.treeNumber w) (List.ofFn pub.nullifiers))
    (hDisjoint : ∀ e ∈ transactSpentPositions bp.treeNumber w, ¬ Spends s.state s' w.token e) :
    SpendEnabled s' bp.treeNumber pub.merkleRoot w.token
      (transactSpentPositions bp.treeNumber w) (List.ofFn pub.nullifiers)
```

(`transact_not_blocked` still takes `S` — via the file-wide `variable`, §5a — only
because `ReachableState (Pool.machine S)`/`Step S` need it to typecheck; its own proof
never reads `S`'s fields.)

The two that keep the full bundle gain `S` and change `Constrained w pub` to
`Constrained S w pub`, nothing else in their own statement:

```lean
theorem transact_spends_only_owned (S : SigScheme) {N M : Nat} {w : TransactWitness N M}
    {pub : TransactPublic N M} (hC : Constrained S w pub) (i : Fin N)
    {trueMpk : Term} {trueRandom : Nat}
    (hSameNote : …) :
    masterPublicKey w.publicKey w.nullifyingKey = trueMpk ∧
      S.verify w.publicKey w.signature (msgHashOf pub)

theorem transact_unshield_exact (S : SigScheme) {s : PoolState} (caller : Term) {N M : Nat}
    (w : TransactWitness N M) (pub : TransactPublic N M) (bp : BoundParams)
    (hN : N ≤ 8) (hM : M ≤ 4) (hC : Constrained S w pub) … :
    ∃ s', Step S s s' ∧ …
```

### 6c. The other nine (every other theorem whose premise mentions a primitive)

**Refinement (3).** All three take `S`, `accepts`, `acceptsOwned` explicitly (through
the `Pool.lean` `variable`, plus two new binders): the statements are unchanged beyond
substituting the new parameters for the axiom names, and `verify.Realizes` for
`verify.Realizes accepts acceptsOwned`.

```lean
theorem Pool.apply_forward_refinement (S : SigScheme)
    (accepts : {N M : Nat} → TransactPublic N M → Prop)
    (acceptsOwned : {N M : Nat} → TransactOwnedPublic N M → Prop)
    (verify : Pool.Verifier) (hReal : verify.Realizes accepts acceptsOwned)
    (hT : TransactProofCompleteness S accepts) (hO : OwnedProofCompleteness acceptsOwned)
    (rs : ReachableState (Pool.machine S)) {s' : PoolState} (h : Step S rs.state s') :
    ∃ c : Pool.Call, Pool.apply verify rs.state.contract c = some s'.contract
```
(`Pool.apply_forward_refinement`, was `PoolSafety.lean:2664-2667`. `Pool.apply`
itself needs no new argument: it never read `S`/`accepts`.)

`Pool.apply_backward_refinement` (was `PoolSafety.lean:2958-2961`) and
`Pool.apply_backward_refinement_run` (was `PoolSafety.lean:3136-3139`, via
`Pool.applyRun_reaches_contract`, `PoolSafety.lean:3105-3109`) take the same four new
binders in place of `hT`/`hO`: `hTS : TransactProofSoundness S accepts`,
`hOS : OwnedProofSoundness acceptsOwned`.

**Can-happen (5).** These take `HonestSignature`/`publicKey : Nat × Nat` as two
separate arguments today; `HonestSignature` is deleted (§3), so each becomes `S` plus
the same `publicKey`, with `hSig.sign`/`hSig.signSound` replaced by `S.sign`/
`S.complete`. This is the minimal-diff substitution (§2a: `sign`/`complete` are
unchanged in shape from `HonestSignature.sign`/`.signSound`), not a restatement of what
is quantified over.

```lean
theorem transact_spend_can_happen (S : SigScheme) {N M : Nat}
    (hComplete : TransactProofCompleteness S accepts) (publicKey : Nat × Nat) …
    : ∃ (w : TransactWitness N M) (pub : TransactPublic N M),
        Constrained S w pub ∧ accepts pub ∧ …
-- was: (hComplete : TransactProofCompleteness) (hSig : HonestSignature) (publicKey : …)
-- PoolSafety.lean:735-737; witness: w.signature := S.sign publicKey (msgHashOf pub);
-- sig := S.complete publicKey (msgHashOf pub).

theorem transact_can_happen (S : SigScheme) (s : ReachableState (Pool.machine S))
    (hComplete : TransactProofCompleteness S accepts) …
-- was PoolSafety.lean:812-814 (hComplete, hSig : HonestSignature).

theorem unshield_can_happen (S : SigScheme) (s : ReachableState (Pool.machine S))
    (hComplete : TransactProofCompleteness S accepts) …
-- was PoolSafety.lean:880-882.

theorem owned_spend_can_happen (hComplete : OwnedProofCompleteness acceptsOwned) …
-- was PoolOwned.lean:372 (hComplete : OwnedProofCompleteness). No S: OwnedConstrained
-- carries no primitive, and this theorem never took HonestSignature.

theorem owned_transact_can_happen (S : SigScheme) (s : ReachableState (Pool.machine S))
    (hComplete : OwnedProofCompleteness acceptsOwned) …
-- was PoolSafety.lean:956-957. Takes S only via ReachableState (Pool.machine S);
-- OwnedProofCompleteness itself does not need S.
```

## 7. The witnesses

Every one of the 14 is a non-vacuity witness (`spec_premise%`) today proved by `sorry`
(`ArmadaLean/Witness/Transact.lean`, `Witness/PoolSafety.lean`, `Witness/PoolOwned.lean`).
Each becomes constructible by instantiating `S := symbolicSigScheme`,
`accepts`/`acceptsOwned` as the symbolic predicates of §4b, and their law records as
`symbolicTransactCompleteness`/`…Soundness`/`symbolicOwnedCompleteness`/`…Soundness`:

- **Refinement (3)** — `apply_forward_refinement`/`apply_backward_refinement`/
  `…_run`: supply `S := symbolicSigScheme`, `accepts`/`acceptsOwned` as in §4b,
  `hT/hO := symbolicTransactCompleteness S`/`symbolicOwnedCompleteness`, `hTS/hOS :=
  symbolicTransactSoundness S`/`symbolicOwnedSoundness`; `verify`/`hReal.verify.Realizes`
  need one further concrete `Pool.Verifier` whose `Bool` checks agree with the symbolic
  `accepts` on some reachable call — the existing `PoolNonVacuity.lean` fixtures (e.g.
  `genesis`, one `shield`) already give a reachable state and a call with no proof
  needed; a trivial `verify := ⟨fun _ _ => false, fun _ _ => false⟩` with `hReal`
  witnessing the (false ↔ false) direction, or any fixture already in
  `PoolNonVacuity.lean`, closes the premise.
- **Can-happen (5)** — each needs a concrete `publicKey`/`owner`, a concrete note
  (`randomIn`, `valueIn`, …), a Merkle witness for it, and `S.complete`/`S.sign`
  (already needed today for `owned_spend_can_happen`, which already has a `sorry`-free
  counterpart pattern in the owned-path fixtures of `PoolNonVacuity.lean`). The plain
  path becomes constructible for the first time: `signature := symbolicSigScheme.sign
  publicKey (msgHashOf pub)`, `sig := symbolicSigScheme.complete publicKey (msgHashOf pub)`.
- **Arithmetic (6)** — the four `ArithConstrained` theorems need only a concrete
  witness/public pair satisfying the (unchanged) arithmetic fields — no `S` at all, so
  these witnesses no longer mention a primitive and are no different in kind from
  the many already-closed `PoolNonVacuity.lean` fixtures. The two full-`Constrained`
  theorems (`transact_spends_only_owned`, `transact_unshield_exact`) need the same
  `symbolicSigScheme`-based `sig` field the can-happen witnesses use.

## 8. The spec module

`scripts/spec_report.lean`'s `requiredAxioms` (`spec_report.lean:77-97`) walks a
premise binder's type, following a `def`'s type/value and a `structure`'s field types,
and collects every name it reaches that is declared `axiom` in the `ArmadaLean`
namespace. This is exactly why the 14 are `WITNESS_BLOCKED`: `Constrained`'s `sig`
field names `EdDSAVerify`, an `axiom`; `TransactProofCompleteness`/`…Soundness`'s
fields name `VerifyAccepts`, an `axiom`; the walk finds them and
`scripts/spec_check.py:169-172` records the theorem as blocked rather than open.

After this design, no premise binder type reaches an `axiom` through `Constrained`/
`ArithConstrained`/`TransactProofCompleteness`/etc.: `S : SigScheme`,
`accepts : {N M} → … → Prop` are ordinary bound parameters, and `SigScheme` itself is a
`structure` (an `inductInfo`, not an `axiomInfo`) whose own field types (`Type`,
function types, `Prop`s) name no axiom. `requiredAxioms` on any of the 14's premise
binders returns `[]`. **`witness_blocked` is empty after this change**, with no edit
to `spec_report.lean`, `spec_check.py`, `spec_lib.py`, or `convert_spec.py`: the check
is already computed from the built environment, not from a stored list of primitive
names (`PRIMITIVE_AXIOMS`, `spec_check.py:71-75`, becomes the empty set; `ALLOWED_AXIOMS`
then equals `CORE_AXIOMS`, and that line can be deleted along with the now-false
module comment at `spec_check.py:67-70` naming the three axioms).

**`convert_spec.py` needs no code change.** Its header parser (`parse_header`,
`convert_spec.py:137-198`) is generic over binder groups: it reads whichever
`(name : type)` groups a `theorem` header has, in order, and carries them into the
generated `Spec.x` definition unchanged. `S : SigScheme`, `accepts : …`, and the
narrowed `ArithConstrained`/full `Constrained S` binders are ordinary groups to it; a
re-run of `scripts/convert_spec.py` on the edited theorems regenerates their
`ArmadaLean/Spec/*.lean` mirrors and `#check_spec_bindings` passes once the mirror's
own binder list is edited to match (the same mechanical step any binder-list change to
a spec theorem already requires — confirmed in the §6a measurement: the one spec-layer
error observed, `Spec/Transact.lean:65:0`, was exactly this definitional-equality
check, cleared by updating the mirror, not a sign of any deeper problem).

**`witness_blocked`'s remaining use after this change: none for these 14.** The
category stays defined (a future primitive could still be added as a bare, law-free
axiom by mistake, and the check would still need to catch it), but no current theorem
or witness exercises it once this design lands. `scripts/spec_check.py`'s
`PRIMITIVE_AXIOMS` set becoming empty is itself the signal that the category is, for
now, vacuous — not a reason to delete the check.

## 9. Rule 8.3: statements that change

Every statement listed changes text (`process.md` rule 8.3: the operator's approval of
each is still required). For each, whether an `old ↔ new` equivalence theorem is
possible:

| Theorem | Old → new | Equivalence possible? |
|---|---|---|
| `transact_value_conserved`, `transact_inputs_distinct_nullifiers`, `transact_public_outputs_determined`, `transact_not_blocked` | `Constrained w pub` → `ArithConstrained w pub` | **Yes.** `ArithConstrained w pub ↔ ∃ S sig, Constrained S w pub` is false in general (no `S` need exist with a `sig` the witness provides), so stated the other way: for a *fixed* `S`, `Constrained S w pub → ArithConstrained w pub` is a one-line forgetful projection (`.toArithConstrained`), and the new theorem is strictly more general (proved from a weaker premise) on its unchanged conclusion. Not a restatement that needs an operator-approved equivalence theorem in the rule-8.3 sense (no conclusion text changed); flag it as a *weakening of the hypothesis*, the direction rule 8.3 exists to let through without a matching equivalence proof. |
| `transact_spends_only_owned`, `transact_unshield_exact` | `Constrained w pub` → `Constrained S w pub` (`S` added, universally quantified) | **Yes, trivially.** The old statement is the new one specialized to one (unnamed, assumed-existing) scheme; since the old statement never named `S`, the new statement with `S` bound outside is definitionally what "the axiom `EdDSAVerify`" already meant — an equivalence is `fun hOld => fun S => hOld` in one direction and is not needed in the other, because the old statement is recovered by instantiating `S`. |
| `transact_spend_can_happen`, `transact_can_happen`, `unshield_can_happen`, `owned_transact_can_happen` | `(hComplete : TransactProofCompleteness) (hSig : HonestSignature)` → `(S : SigScheme) (hComplete : TransactProofCompleteness S accepts)` | **Yes.** `HonestSignature` is deleted in favor of `SigScheme`'s identically-shaped `sign`/`complete` fields; a value of the old type converts to the new `SigScheme`'s two fields and back with no information lost (same fields, same laws) — this is a renaming, not a strengthening. |
| `Pool.apply_forward_refinement`, `Pool.apply_backward_refinement`, `Pool.apply_backward_refinement_run` | add `S`, `accepts`, `acceptsOwned`; `verify.Realizes` → `verify.Realizes accepts acceptsOwned` | **Yes**, by the same argument as the `transact_spends_only_owned` row: the old statement is the new one with the (previously unnamed) axiom names substituted for the new parameters. |
| Every theorem gaining `S` *only* because its statement mentions `Pool.machine`/`Step`/`Reachable` (§5a; ~58 of the ~62 counted, none of the 14) | add `S` as a leading parameter, nothing else | **Not a statement change in the rule-8.3 sense.** No claim's *content* differs; `Pool.machine S` for an arbitrarily-quantified `S` is the same object the old, unparameterized `Pool.machine` already meant implicitly (there was always exactly one real scheme; it is just unnamed before this change). These do not need individual operator sign-off as statement changes — flag them as one batch, mechanical, with a single approval covering the whole batch, not 58 separate approvals. |

No row hides a real strengthening or restatement behind an "equivalence": the two
honest exceptions are `transact_spends_only_owned`/`transact_unshield_exact`'s
hypothesis gaining the genuine extra structure `S` (still an equivalence, because `S`
was always implicitly there) and the four can-happen theorems' `HonestSignature`
deletion (a renaming, not a semantic change, because the fields are identical).

## 10. Docs to update

- **`README.md:184-189`** — "EdDSA stays an abstract unforgeable signature scheme — an
  **explicit axiom** (`EdDSAVerify`) … The Groth16 verifiers are the uninterpreted
  predicates `VerifyAccepts` and `VerifyOwnedAccepts`, also declared with `axiom`."
  Becomes: EdDSA and the two Groth16 verifiers are explicit *structure parameters*
  (`SigScheme`, `TransactProofCompleteness`/`Soundness` pairs), not axioms; the symbolic
  instance (§4) is the assumption actually in force when a witness uses them, worded
  the same way the hash section above it already handles Poseidon/Keccak ("the
  assumption lives in the choice of model, not in a per-theorem axiom").
- **`docs/what-the-proofs-mean.md:52-53`** — "EdDSA signature unforgeability is still
  assumed (`EdDSAVerify` axiom)." Becomes: EdDSA completeness is a structure law
  (`SigScheme.complete`), satisfied by the symbolic model; this project does not
  formalize EdDSA unforgeability at all (§2a), the same honest gap already stated for
  Groth16 soundness ("these are standard assumptions … we just make them visible").
- **`docs/how-to-verify.md:69,92-101`** — "either a named `axiom` (EdDSA, the Groth16
  verifiers), or a choice of symbolic model" and "**Theorems that use a proof check**
  may list `EdDSAVerify`, `VerifyAccepts` and `VerifyOwnedAccepts`." The `#print axioms`
  walkthrough (lines 76-86) no longer shows these three names for any of the 14 — its
  prose and its worked example both need the corrected wording and, if the example
  theorems are among the 14, a note that they list no primitive axiom at all now.
- **`docs/verification-chain.md:65-68`** — "The specification keeps `VerifyAccepts`,
  `VerifyOwnedAccepts` and `EdDSAVerify` as the trusted primitives." The *primitives
  stay trusted* (step 4 still does not refine the Groth16 verifier or EdDSA); only the
  Lean encoding of "trusted" changes, from a global axiom to a universally-quantified
  structure parameter with laws. Reword to name the structures, not the deleted axioms.
- **`docs/axioms-circuit.txt`** — a static, hand-generated audit file (not a
  `scripts/check.sh` input). Every line naming `EdDSAVerify` in a theorem's axiom list
  (`transact_value_conserved`, `transact_inputs_distinct_nullifiers`, and the trailing
  note explaining why those two mention it) is now wrong: after §6, neither theorem's
  axiom list contains `EdDSAVerify`/anything else beyond the three core axioms, because
  neither takes a `Constrained S`/`sig`-bearing premise any more. Needs regenerating
  (`lake env lean /tmp/axioms_check.lean` per its own header) or deleting the two
  `EdDSAVerify` lines and their explanatory note by hand.
- **`ArmadaLean/PoolNonVacuity.lean:8-15`** (module comment, not `docs/`, but the
  honest-wording rule applies the same way) — explains that every owned-path fixture
  is owned-path *only because* `EdDSAVerify` was an unprovable axiom. After this
  design, plain-path fixtures become constructible (§7); the comment's "no concrete
  instance of it is provable" is no longer true and needs rewording, or the plain-path
  fixtures it predicts are impossible should be added alongside it.
- **`scripts/oracles/pool_exec/PE1-PE5.lean`** — not `docs/`, and not gated by
  `scripts/check.sh`, but committed source documenting the PE-rule interfaces (§3);
  each needs its axiom-allowlist literal and (`PE3.lean` only) its two `example`
  statements updated, or they stop compiling once the three axioms are deleted.

The new honest wording, every place: *the symbolic signature scheme and the symbolic
proof-verifier instance are the assumption in force for every witness this project
exhibits*, the same sentence already used for Poseidon/Keccak. Nothing becomes
"unconditional" that was conditional before; the primitives stay exactly as trusted as
`docs/verification-chain.md`'s 2026-10-28 ruling already says.

## 11. Order of work and acceptance

1. `ArmadaLean/Transact.lean`: add `SigScheme` (§2a); split `Constrained`/
   `ArithConstrained` (§5); update `Constrained`'s one other use
   (`transact_value_conserved`'s self-reference in its own doc comment, cosmetic).
2. `ArmadaLean/PoolSafety.lean`: delete `axiom VerifyAccepts`; parameterize
   `TransactProofCompleteness`/`TransactProofSoundness`/`Pool.Verifier.Realizes` (§2b,
   §3); delete `HonestSignature`; add `variable (S : SigScheme)` near the top (§5a) and
   let it auto-thread; fix every call site `lake build` flags (the mechanical
   `hC.toArithConstrained` kind of fix seen in §6a's measurement, and the `S`/`accepts`
   arguments the refinement and can-happen theorems now take, §6c).
3. `ArmadaLean/Pool.lean`: parameterize `Step`, `Reachable`, `Pool.machine` with `S`
   (§5a); fix every call site (the bulk of the diff, by line count; none of it changes
   a proof strategy).
4. `ArmadaLean/PoolOwned.lean`: delete `axiom VerifyOwnedAccepts`; parameterize
   `OwnedProofCompleteness`/`OwnedProofSoundness` (§2b). No `S` needed anywhere in this
   file (`OwnedConstrained` carries no primitive).
5. `ArmadaLean/PoolNonVacuity.lean`: add the symbolic instances (§4); build the plain-
   path fixtures the module comment currently says are impossible; update the comment.
6. `ArmadaLean/Witness/Transact.lean`, `Witness/PoolSafety.lean`, `Witness/PoolOwned.lean`:
   replace the 14 `sorry`s with the witnesses of §7, built from step 5's fixtures;
   re-run `scripts/convert_spec.py` to regenerate the matching `ArmadaLean/Spec/*.lean`
   mirrors for every theorem whose binder list changed (§8).
7. `scripts/spec_check.py`: delete `PRIMITIVE_AXIOMS` and its now-false module comment
   (§8); `ALLOWED_AXIOMS` becomes `CORE_AXIOMS` alone.
8. `scripts/oracles/pool_exec/PE1-PE5.lean`: drop the three deleted names from the
   allowlist literal; fix `PE3.lean`'s two examples (§3).
9. Docs (§10): `README.md`, `docs/what-the-proofs-mean.md`, `docs/how-to-verify.md`,
   `docs/verification-chain.md`, `docs/axioms-circuit.txt`.
10. Operator approval of the rule-8.3 statement changes (§9), as one batch for the
    mechanical `S`-threading rows and individually for the four substantive rows.

**Acceptance check**, exactly the brief's three criteria: `scripts/check.sh` exits 0;
its `--json` output's `witness_blocked` is `{}`; `#print axioms` on each of the 14
theorems (and on `Step`, `Pool.machine`, and any other declaration whose signature
changed) lists only `propext`, `Classical.choice`, `Quot.sound`.

## 12. Where the code needs more than the lab book entry says

1. **The `Step`/`Reachable`/`Pool.machine` ripple (§5a) is not in the lab book entry
   at all.** It frames option 2 as "each theorem quantifies over every scheme that
   satisfies the laws it uses," which reads as a per-theorem, opt-in cost. Because
   `Step.transact` bakes `Constrained` into the state machine's own constructor, the
   cost is file-wide and largely opt-out: ~62 of `PoolSafety.lean`'s 76 theorems gain
   an `S` parameter whether or not their own proof ever reads a `SigScheme` law. The
   decision to do the split is unaffected by this (there is no cheaper way to keep
   `Step.transact` requiring a genuinely verified signature), but the *size* of the
   work the lab book entry scoped ("closes review item 3 for all 14 blocked theorems")
   understates the diff by roughly a factor of five once this ripple is counted.
2. **`OwnedConstrained` does not carry a primitive**, confirmed by reading
   `TransactOwned.lean:158-181`: no field of it mentions `VerifyOwnedAccepts` or any
   axiom. The lab book entry's phrasing in the brief ("the same for `OwnedConstrained`
   if it carries a primitive") correctly anticipates this might not apply; it does not.
   `OwnedConstrained` needs no `ArithConstrained`-style split and no `S`.
3. **`HonestSignature` (`PoolSafety.lean:717-719`) already exists and is already a
   completeness law for `EdDSAVerify`**, not mentioned in the lab book entry's survey
   of what exists today. It is not a bare axiom-adjacent hypothesis; it is exactly the
   "completeness" half of option 2's own recipe, already written, already used by three
   of the five can-happen theorems. The only primitive genuinely missing a law today is
   the bare `EdDSAVerify`/`VerifyAccepts`/`VerifyOwnedAccepts` themselves (the `Prop`,
   not the law); the laws for two of the three were already present and are reused, not
   invented.
4. **No current theorem needs EdDSA soundness/unforgeability** (§2a), contradicting
   the informal doc comment at `PoolSafety.lean:211-213` ("`Constrained.sig` … under 'a
   signature cannot be forged' … only the true holder of `w.publicKey`'s private key
   can produce [it]"), which attributes a proof step to `transact_spends_only_owned`
   that its actual Lean body (`⟨commitment_mpk_inj hSameNote, hC.sig⟩`) does not take.
   The lab book entry's option 2 describes "a signature by the key holder" as the
   soundness law to add in general; for this specific primitive, in this project, as it
   stands today, no theorem's proof uses that direction, and formalizing it honestly
   needs machinery (`SK`, `pkOf`) this design recommends deferring, not building
   unused.
