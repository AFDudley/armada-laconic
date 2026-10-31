# Pool slice audit — armada-lean type and layer doctrine

Scope: `Pool.lean`, `PoolSafety.lean`, `PoolOwned.lean`, `PoolMachine.lean`, `Transact.lean`,
`TransactOwned.lean`, `Merkle.lean`, `Nullifier.lean`, `Poseidon.lean`, `Field.lean`, `RangeCheck.lean`,
`ValueConservation.lean`, `BoundParams.lean`, `Token.lean`, `Fee.lean`, `Vault.lean` — every
`ArmadaLean/*.lean` module outside `ArmadaLean/Nitro/`, on trunk `integrated` (`522e16d`) at
`/home/rix/code/git_puller/repos/armada-lean/.scratch/audit-wt`.

Counts: 16 modules, 445 declarations in `facts.json`, 345 after removing compiler-generated
`recOn`/`casesOn`/`noConfusion`/`.elim`/`inst*` records. 187 theorems; 44 of them spec-level
(doc comment cites a `.sol:`/`.circom:` line or an operator ruling). 12 of the 44 spec-level
theorems are layer violations (§4). 9 declarations exceed 60 lines (§5). 8 declarations are dead
code (§6). No `sorry` in scope (`grep -rn sorry` on all 16 files, empty; `facts.json`'s `sorry`
field agrees).

## 1. Layer map

| Module | Doctrine layer (§4a) | Public API (used by another module) |
|---|---|---|
| `Field.lean` | arithmetic primitive, below Pool base ops | `p`, `fieldEq`, `eq_of_fieldEq_of_lt`, `sum_bound_sub_one_lt_p` (used by `ValueConservation.lean`) |
| `RangeCheck.lean` | arithmetic primitive | `RangeChecked120`, `Spendable`, `rangeCheck120_sound` (used by `Pool.lean`'s `shield`, `Transact.lean`, `TransactOwned.lean`) |
| `Poseidon.lean` | cryptographic primitive (Dolev-Yao term algebra) | `Term`, `Term.hash2`/`hash3`/`keccak`, `Term.hash2_inj`/`hash3_inj`/`poseidon_ne_keccak` — used pervasively by every module below |
| `Merkle.lean` | cryptographic primitive, over `Poseidon` | `TreeDepth`, `emptyLeaf`, `treeRoot`, `treePath`, `merkleRoot_inj_leaf`, `treePath_correct_of_getElem?`, `treePath_length` — used by `Pool.lean` (`TreeState`), `Transact.lean`/`TransactOwned.lean` (`Constrained.merkle`) |
| `Nullifier.lean` | cryptographic primitive, over `Poseidon` | `nullifier`, `nullifier_pos_inj`, `distinct_leaves_distinct_nullifiers` — used by `Pool.lean`/`PoolSafety.lean`/`Transact.lean`/`TransactOwned.lean` |
| `Token.lean` | symbolic data primitive | `Token`, `tokenId`, `tokenId_inj` — used pervasively |
| `Fee.lean` | pool-policy primitive | `FeeRatio`, `shieldFee`/`shieldBase`, `fee_conservation` — used by `Pool.lean`'s `shield`, `PoolSafety.shield_creates_exactly` |
| `BoundParams.lean` | symbolic data primitive | `BoundParams`, `Unshield`, `hashBoundParams`, `hashBoundParams_inj`, `AdaptAllows`, `withdrawn` — used by `Pool.lean`'s `transact`/`transactOwned` |
| `ValueConservation.lean` | arithmetic layer, over `Field` | `listSum`, `listSum_append`/`_perm`, `no_wraparound_minting` — used by `Transact.lean`/`TransactOwned.lean` (circuit balance) and `Pool.lean`/`PoolSafety.lean` (held-value bookkeeping) |
| `Transact.lean` | circuit-constraint layer (plain path) | `EdDSAVerify`, `masterPublicKey`, `npk`, `commitment`, `TransactWitness`, `TransactPublic`, `Constrained`, `transact_value_conserved`, `commitment_mpk_inj` — used by `Pool.lean`/`PoolSafety.lean` |
| `TransactOwned.lean` | circuit-constraint layer (owned path) | `TransactOwnedWitness`, `TransactOwnedPublic`, `OwnedConstrained`, `transactOwned_value_conserved` — used by `Pool.lean`/`PoolSafety.lean`/`PoolOwned.lean` |
| `PoolOwned.lean` | pool base operations + owned-path safety | `CommitmentPreimage`, `hashCommitment`, `NullifierSet`, `markSpent`, `PreimageRegistry`, `recordStep`, `OwnedOpening`, `owned_owner_binding`, `owned_no_double_spend`, `markSpent_succeeds` — used by `Pool.lean`/`PoolSafety.lean` |
| `Pool.lean` | pool base operations (`removeAll`, `withTreePositionsFrom`, `applyRecords`, `insertLeaves`/`insertTree`/`insertStart`) + state and step relation | `PoolState`, `TreeState`, `Step`, `genesis`, `Reachable`, `transactSpentPositions`, `ownedSpentPositions` — used by `PoolSafety.lean`/`PoolMachine.lean` |
| `PoolMachine.lean` | adapter to the generic `Machine` interface | `Pool.machine` — used by `PoolSafety.lean`'s `can_happen` theorems (as an unused hypothesis, §7) |
| `PoolSafety.lean` | per-call frame lemmas → state invariants → pool properties → can-happen/not-blocked | `core_invariant`, `pool_no_double_spend`, `transact_spends_only_owned`, `leaf_has_merkle_witness`, `live_values_spendable`, `transact_not_blocked`, `shield_can_happen`/`transact_can_happen`/`unshield_can_happen`/`owned_transact_can_happen`/`owned_merge_can_happen` — these are the module's own deliverables, not consumed elsewhere in scope |
| `Vault.lean` | standalone — no import, no reference to `Step`/`PoolState`/`Reachable` | `c2s`/`c2a`/`c2sV`/`c2aV`, `inflation_attack_profitable`, `mitigated_attack_unprofitable` — not consumed by any other module in scope. **Being remodelled on `draft/vault-locked-shares`; this audits the trunk version only.** |

## 2. Missing layer API

Five places where the same fact is proved twice, each with a proposed single replacement.

**(a) `transact_value_conserved` / `transactOwned_value_conserved` — verbatim identical proof.**
`Transact.lean:215-229` and `TransactOwned.lean:157-171`: both bodies are
`apply no_wraparound_minting _ _ _ _ _ _ h.balance` followed by the same four tactics
(`rw [List.length_ofFn]; exact hN`/`hM`, then `List.mem_ofFn.mp` + `rangeCheck120_sound`), differing
only in the ambient hypothesis's name (`h : Constrained w pub` vs `h : OwnedConstrained w pub`).
Replace with one lemma generic over a structure that both `Constrained` and `OwnedConstrained`
could extend, or that both pass their shared fields to:
```lean
theorem value_conserved_of {N M : Nat} (hN : N ≤ 8) (hM : M ≤ 4)
    (valueIn : Fin N → Nat) (valueOut : Fin M → Nat)
    (hbalance : fieldEq (listSum (List.ofFn valueIn)) (listSum (List.ofFn valueOut)))
    (hrangeIn : ∀ i, RangeChecked120 (valueIn i)) (hrangeOut : ∀ j, RangeChecked120 (valueOut j)) :
    listSum (List.ofFn valueIn) = listSum (List.ofFn valueOut) := by
  apply no_wraparound_minting _ _ (by rw [List.length_ofFn]; exact hN)
    (by rw [List.length_ofFn]; exact hM)
    (fun v hv => rangeCheck120_sound v (by obtain ⟨i, hi⟩ := List.mem_ofFn.mp hv; rw [← hi]; exact hrangeIn i))
    (fun v hv => rangeCheck120_sound v (by obtain ⟨j, hj⟩ := List.mem_ofFn.mp hv; rw [← hj]; exact hrangeOut j))
    hbalance
```
`transact_value_conserved`/`transactOwned_value_conserved` become one-line callers.

**(b) `commitment_mpk_inj` / `commitment_owner_inj` — verbatim identical proof.**
`Transact.lean:245-249` and `TransactOwned.lean:182-187`: both are
`obtain ⟨hnpk, -, -⟩ := Term.hash3_inj h; exact (Term.hash2_inj hnpk).1`, same hypothesis/goal
shape (`commitment (npk X₁ r₁) token value = commitment (npk X₂ r₂) token value → X₁ = X₂`).
Replace with:
```lean
theorem commitment_key_inj {X₁ X₂ : Term} {r₁ r₂ : Nat} {token : Token} {value : Nat}
    (h : commitment (npk X₁ r₁) token value = commitment (npk X₂ r₂) token value) : X₁ = X₂ := by
  obtain ⟨hnpk, -, -⟩ := Term.hash3_inj h
  exact (Term.hash2_inj hnpk).1
```
in `Nullifier.lean` or a shared base module both `Transact.lean` and `TransactOwned.lean` import;
`commitment_mpk_inj`/`commitment_owner_inj` become aliases or disappear.

**(c) `step_preserves_live_spent_invariant`'s `transact`/`transactOwned` arms — line-for-line duplicate.**
`PoolSafety.lean:1203-1430` (`transact`, 228 lines) and `PoolSafety.lean:1431-1656` (`transactOwned`,
226 lines) run the identical `by_cases`/`rcases` tree and call the identical lemmas
(`withTreePositionsFrom_treeNumber`, `withTreePositionsFrom_pos_lt`/`_ge`, `removeAll_perm`,
`removeAll_disjoint`, `markSpent_superset`), differing only in which term supplies the spending key
(`nullifier (Term.val w.nullifyingKey) …` at `PoolSafety.lean:1309` vs `nullifier pub.ownerField …`
at `PoolSafety.lean:1537`) and which "spent positions" helper is named
(`transactSpentPositions` vs `ownedSpentPositions`). Matching blocks: `PoolSafety.lean:1300-1321`
vs `:1528-1549` (nullifier-superset skeleton); `:1215-1234` vs `:1443-1462` (membership-threading);
`:1415-1430` vs `:1641-1656` (`Nodup`-closing). Replace with one lemma parametrized by the key term
and the spent-positions list:
```lean
theorem live_spent_invariant_step_of {s : PoolState} {token : Token} {treeNumber : Nat}
    {spentPositions newPositions : List ((Nat × Nat) × CommitmentPreimage)}
    {rest : List ((Nat × Nat) × CommitmentPreimage)} {spent' : NullifierSet}
    (keyOf : (Nat × Nat) × CommitmentPreimage → Term)
    (hRemove : removeAll (s.live token) spentPositions = some rest)
    (hNullify : markSpent (s.spent treeNumber) (spentPositions.map (fun p => keyOf p)) = some spent')
    (hinv : LiveSpentInvariant s) : LiveSpentInvariant (/- the post-step state -/ s) := ...
```
(exact signature depends on `LiveSpentInvariant`'s shape, `PoolSafety.lean:1098` region — a worker
dispatched on this item reads the current five-part conjunction first). `transact`'s and
`transactOwned`'s arms both become one call.

**(d) `owned_merge_can_happen`'s inline re-derivation.**
`PoolSafety.lean:2255-2278` ("inputs no longer live") duplicates `PoolSafety.lean:1443-1462`/
`:1649-1654`'s reasoning by hand instead of calling the lemma proposed in (c).
`PoolSafety.lean:2279-2297` ("output is live") directly unfolds `withTreePositionsFrom`/
`withPositionsFrom` instead of calling a per-call "freshly inserted note is live" lemma — no such
lemma currently exists for `Step.transactOwned`'s output case (the closest, `withTreePositionsFrom
_leaf_matches`, used at `PoolSafety.lean:1783-1789`, proves leaf-content matching, not live-set
membership). Propose adding:
```lean
theorem transactOwned_new_note_live {s : PoolState} {N M : Nat} (w : TransactOwnedWitness N M)
    (pub : TransactOwnedPublic N M) (bp : BoundParams) (hN : N ≤ 8) (hM : M ≤ 4)
    (hC : OwnedConstrained w pub) (j : Fin M) (s' : PoolState) (hstep : Step s s')
    (hwitness : /- this step is this transactOwned call -/ True) :
    ((insertTree s.trees (ownedNewNotes w).length,
      insertStart s.trees (ownedNewNotes w).length + j),
      (⟨w.npkOut j, w.token, w.valueOut j⟩ : CommitmentPreimage)) ∈ s'.live w.token := ...
```
as a per-call frame lemma next to `step_live_leaf_matches`; `owned_merge_can_happen` then calls it
instead of unfolding `insertTree`/`insertStart`/`withPositionsFrom` at `PoolSafety.lean:2258-2296`.

**(e) `applyRecords_preserves_consistency` hand-rolled instead of calling `foldl_preserves`.**
`Pool.lean:502-514` re-proves "a property preserved by every step of a fold survives the whole
fold" by its own `induction records generalizing reg with …`, even though the identical generic
lemma `foldl_preserves` already exists (`PoolOwned.lean:245-248`, imported transitively via
`Pool.lean:92`) and is already used for exactly this purpose by
`PoolOwned.recordPreimages_preserves_consistency` (`PoolOwned.lean:263-270`). Replace
`applyRecords_preserves_consistency`'s body with a call to `foldl_preserves (fun r cp =>
recordStep r cp.1 cp.2) RegistryConsistent …`.

## 3. Refined-type candidates

**(a) The "well-formed spend witness" bundle — 4 agreeing instances.**
`transact_can_happen` (`PoolSafety.lean:660`), `unshield_can_happen` (`:728`),
`owned_transact_can_happen` (`:804`) and `owned_merge_can_happen` (`:2139`) each take, as separate
hypotheses, the same seven-part bundle: root membership (`hRoot : root ∈ s.rootsAt' bp.treeNumber`,
`:668`/`:737`/`:811`), path length (`hPathLen : ∀ i, (pathElements i).length = 16`,
`:669`/`:738`/`:812`), input/output range checks (`hRangeIn`/`hRangeOut : RangeChecked120`,
`:675-676`/`:744`/`:816-817`), distinct leaf indices (`hDistinctLeaves`, `:678`/`:746`/`:819`; renamed
`hDistinct` at `:2145`), nullifier freshness (`hFresh`, `:679`/`:747`/`:820`, same shape, only the key
term varies), Merkle membership (`hMerkle`, `:670-674`/`:739-743`/`:813-815`), and the live-removal
precondition (`hLive`/`hRemove`, `:681-686`/`:749-754`/`:823-827`). All four instances agree on the
same constraint — no operator ruling needed, only a type. Candidate:
```lean
structure SpendWitness (N M : Nat) where
  treeNumber : Nat
  keyOf : Fin N → Term
  randomIn valueIn : Fin N → Nat
  leavesIndices : Fin N → Nat
  pathElements : Fin N → List (Term × Bool)
  npkOut : Fin M → Term
  valueOut : Fin M → Nat
  pathLen : ∀ i, (pathElements i).length = TreeDepth
  rangeIn : ∀ i, RangeChecked120 (valueIn i)
  rangeOut : ∀ j, RangeChecked120 (valueOut j)
  distinctLeaves : (List.ofFn leavesIndices).Pairwise (· ≠ ·)
```
with the per-path specifics (`keyOf` instantiated to `masterPublicKey w.publicKey w.nullifyingKey`
on the plain path, to `fun _ => pub.ownerField` on the owned path; EdDSA vs. owner-equality) staying
as separate hypotheses, per Transact's/TransactOwned's own near-duplicate structures (§2a/b apply
to `Constrained`/`OwnedConstrained` themselves too).

**(b) `pathLen = TreeDepth` — repeated single-input check, 6 sites.**
`Constrained.pathLen` (`Transact.lean:144`), `OwnedConstrained.pathLen` (`TransactOwned.lean:126`),
and the four `hPathLen` hypotheses above all restate `∀ i, (pathElements i).length = 16` (or `= TreeDepth`
once named) over one input, `pathElements`. Doctrine rule 2 ("a hypothesis that constrains one input
alone is a refined type of that input"): `pathElements : Fin N → List (Term × Bool)` should be
`pathElements : Fin N → {p : List (Term × Bool) // p.length = Merkle.TreeDepth}`, removing `pathLen`
from every one of the 6 sites.

**(c) `RangeChecked120` input/output values — repeated single-input check, 8 sites.**
`Constrained.rangeIn`/`.rangeOut` (`Transact.lean:158-160`), `OwnedConstrained.rangeIn`/`.rangeOut`
(`TransactOwned.lean:139-141`), and the `hRangeIn`/`hRangeOut` hypotheses of the three `can_happen`
theorems (§3a) all restate `RangeChecked120 (valueIn i)`/`RangeChecked120 (valueOut j)` over one
input each. `RangeCheck.lean` already names the typed carrier (`Spendable`, `RangeCheck.lean:41`).
`valueIn`/`valueOut : Fin N → Nat` should be `Fin N → {v : Nat // RangeChecked120 v}`.

**(d) Distinct leaf indices — repeated relational premise, agreeing.**
`distinct_leaves_distinct_nullifiers`'s own hypothesis (`Nullifier.lean:46`,
`leaves.Pairwise (· ≠ ·)`) and the `hDistinctLeaves`/`hDistinct` hypotheses of §3a's four theorems
state the same fact over one input (`leavesIndices`/`pos`). Candidate:
`leavesIndices : Fin N ↪ Nat` (`Function.Embedding`, which already carries injectivity as `Fin N →
Nat` plus a proof) in place of `Fin N → Nat` with a separate `Pairwise` hypothesis — `List.ofFn
f.injective`-style lemmas replace `ofFn_pairwise_ne` (`PoolSafety.lean:540-551`).

## 4. Layer violations

12 of 44 spec-level theorems — matches the doctrine's §4a Pool count exactly, and matches
`TypeLinters`' independent `layerDiscipline` linter run (same 12 names).

| Theorem | Line | Base op named directly | What it should call instead |
|---|---|---|---|
| `PoolOwned.owned_no_double_spend` | `PoolOwned.lean:174` | `unfold markSpent at h; by_cases …` (`:177-192`) | None — this IS the primitive theorem about `markSpent`; not a true violation (see §7). |
| `PoolOwned.owned_spend_can_happen` | `PoolOwned.lean:314` | calls `markSpent_succeeds` inline (`:366`) rather than through a `Step`-level frame lemma | A `PoolOwned`-level "spend succeeds" frame lemma, mirrored by (e) below for the Pool level |
| `PoolSafety.spent_monotone` | `:259` | `cases hstep with` over all 3 constructors (`:266-280`), each calling `markSpent_subset` | Acceptable — it IS the per-call frame lemma the doctrine asks for (§4a item 2); flagged only because `pool_base_ops` counts the `cases hstep` itself |
| `PoolSafety.pool_no_double_spend` | `:298` | calls `markSpent` indirectly via `owned_no_double_spend`/`spent_monotone_from` | Already goes through lemmas; `pool_base_ops` counts the transitive `markSpent` reference in `owned_no_double_spend`'s own citation |
| `PoolSafety.shield_creates_exactly` | `:404` | calls `insertLeaves`/`insertStart`/`insertTree` directly in its witness (`:431-437`) | A `shield`-call frame lemma (`insertLeaves_leavesAt`/`insertStart_eq` already exist at `Pool.lean:762`/`:805`; `shield_creates_exactly` should state its result purely in terms of those, not re-derive from `insertTree s.trees 1` literally) |
| `PoolSafety.pool_preimageOf_consistent` | `:495` | calls `applyRecords_preserves_consistency`/`recordStep_preserves_consistency` — both are the base-op-adjacent fold lemmas, not per-call wrappers | §2e's `foldl_preserves`-based rewrite reduces this to one indirection, not two |
| `PoolSafety.transact_spend_can_happen` | `:576` | calls `markSpent_succeeds` inline (`:628`) | Same gap as `owned_spend_can_happen` |
| `PoolSafety.shield_can_happen` | `:639` | calls `Step.shield` directly with `insertTree`/`insertStart` already appearing in its witness via `shield_creates_exactly` | See `shield_creates_exactly` above |
| `PoolSafety.transact_can_happen` | `:650` | calls `markSpent_succeeds` inline (`:705`) | §2 item (d)'s "per-call is-live/is-spendable frame lemma", generalized to `transact` |
| `PoolSafety.unshield_can_happen` | `:728` | calls `markSpent_succeeds` inline (`:776`) | Same |
| `PoolSafety.owned_transact_can_happen` | `:795` | calls `markSpent_succeeds` inline (`:843`) | Same |
| `PoolSafety.owned_merge_can_happen` | `:2124` | calls `markSpent_succeeds` (`:2244`) and directly unfolds `insertTree`/`insertStart`/`withTreePositionsFrom`/`withPositionsFrom` (`:2258-2296`) | §2 item (d)'s proposed `transactOwned_new_note_live` |

Net: the real, actionable gap is one missing layer — "a per-call lemma: after a successful
`markSpent`/`insertLeaves` call inside `Step.transact`/`.transactOwned`/`.shield`, the new note is
live and the spent positions are gone" — which §2's items (c) and (d) name precisely. 8 of the 12
reduce to the SAME missing lemma (`markSpent_succeeds` called inline instead of through a
`Step`-level wrapper); fixing that one lemma clears 8 of the 12 violations in one unit of work.

## 5. Size

9 declarations over 60 lines (excludes `Step.recOn`/`Step.casesOn`, compiler-generated).

| Declaration | Line | Lines | Cause |
|---|---|---|---|
| `PoolSafety.step_preserves_live_spent_invariant` | `:1098` | 559 | Case split over 3 `Step` constructors; `transact` (`:1203-1430`) and `transactOwned` (`:1431-1656`) arms are near-verbatim duplicates — §2(c) |
| `PoolSafety.owned_merge_can_happen` | `:2124` | 174 | Inline re-derivation of facts `step_preserves_live_spent_invariant`'s `transactOwned` arm and `step_live_leaf_matches` already prove — §2(d) |
| `PoolSafety.step_live_leaf_matches` | `:1681`(doc)/`:1691`(body) | 124/114 | Case split over 3 `Step` constructors; `transact` (`:1728-1765`) and `transactOwned` (`:1766-1804`) arms near-identical, differ only by `newNotes` vs `ownedNewNotes w` |
| `Pool.Step` (the `inductive` itself) | `:1047` | 107 | Not a violation: the state machine's 3 real operations, each with a full per-field provenance doc comment (writing-lean.md's documentation convention). Acceptable size for a type definition, listed for completeness. |
| `PoolSafety.unshield_can_happen` | `:721`(doc)/`:728`(body) | 73/66 | Near-duplicate of `transact_can_happen` with `M` fixed to 1 and the unshield branch forced — same hypothesis-bundle duplication as §3(a), same `markSpent_succeeds`-inline cause as §4 |
| `PoolSafety.transact_can_happen` | `:650`(doc)/`:660`(body) | 70 | Same family — §3(a)/§4 |
| `PoolOwned.owned_spend_can_happen` | `:314` | 64/48 | Builds a full witness term explicitly, structurally identical to `PoolSafety.transact_spend_can_happen` (`:576`, its own doc comment: "exactly as `owned_spend_can_happen` does") — twin across files |
| `PoolSafety.transact_spend_can_happen` | `:576` | 62 | Twin of `owned_spend_can_happen` above |
| `PoolSafety.owned_transact_can_happen` | `:795` | 60 | Same family — §3(a)/§4 |

All 9 reduce to two causes: (1) the `transact`/`transactOwned` case-split duplication (§2c/d, 3
declarations), and (2) the circuit-level "X succeeds" witness-construction family
(`owned_spend_can_happen`/`transact_spend_can_happen`/`transact_can_happen`/`unshield_can_happen`/
`owned_transact_can_happen`, 5 declarations, all sharing §3(a)'s hypothesis bundle and §4's
`markSpent_succeeds`-inline gap). Fixing §2(c)/(d) and the `markSpent_succeeds` wrapper (§4) would
bring all 9 under 60 lines without losing any statement.

## 6. Dead code

Checked `users == 0` in the regenerated `facts.json`, excluding spec-level theorems (`cites = true`)
and Vault's exploit-demonstration theorems (the module's whole point, not meant to be reused), then
confirmed each by `git grep` across the full repository (`docs/`, `scripts/`, `.pebbles/`, not just
`ArmadaLean/`).

| Declaration | Line | Evidence |
|---|---|---|
| `Field.p_pos` | `Field.lean:15` | `git grep -n p_pos` outside `Field.lean`: only `scripts/sigs-baseline.tsv` (a generated catalog, not a use). |
| `Pool.withPositionsFrom_nodup` | `Pool.lean:412` | Same: only in `sigs-baseline.tsv`. Its tree-tagged "counterpart" `withTreePositionsFrom_nodup` (`Pool.lean:429-442`) runs its own separate induction instead of calling it — this is also a §2-style missing-API gap. |
| `PoolMachine.Pool.machine_reachable_iff` | `PoolMachine.lean:19` | `git grep -n machine_reachable_iff`: zero hits outside its own declaration line. The module's own doc comment (`PoolMachine.lean:1-7`) says `PoolSafety.lean`'s 4 can-happen theorems take `Pool.machine.Reachable s` as a hypothesis specifically so this file is imported first — but those 4 theorems never call `machine_reachable_iff` (§7). |
| `PoolOwned.preimageOf_consistent` | `PoolOwned.lean:281` | `git grep`: only doc-comment mentions in `Pool.lean:500` and `PoolSafety.lean:498`. `PoolSafety.pool_preimageOf_consistent` (`:502-516`) re-derives the fact via `applyRecords_preserves_consistency`/`recordStep_preserves_consistency` instead of calling it. |
| `Poseidon.Term.val_ne_hash` | `Poseidon.lean:148` | `git grep`: zero hits outside its own declaration. |
| `Vault.attackAttackerShares` | `Vault.lean:55` | `git grep -n "attackAttackerShares\|attackVictimShares\|attackAttackerOut"` outside `Vault.lean`: no output. The theorems documented as using them (`inflation_attack_victim_gets_nothing`, `:74-84`; `inflation_attack_profitable`, `:86-102`) restate the same formulas (`c2s 1 (1 + X) Y`, `c2a 1 (1 + X + Y) (c2s 0 0 1)`) inline instead of calling the named "step 1/3/4" defs their own doc comments (`:50-53`) describe. |
| `Vault.attackVictimShares` | `Vault.lean:59` | Same evidence. |
| `Vault.attackAttackerOut` | `Vault.lean:63` | Same evidence. |

Not listed (excluded by the report's own carve-out, verified, noted for completeness): `core_invariant`
(`PoolSafety.lean:157`), `spend_live_exact` (`:337`), `transact_unshield_exact` (`:369`),
`transact_root_was_produced` (`:473`) are zero-user but each is the module's own top-level safety
deliverable (README.md's safety-theorem catalog lists each by name, `README.md:73-94`), not dead;
`inflation_attack_victim_gets_nothing`/`inflation_attack_profitable`/`mitigated_attack_unprofitable`
(`Vault.lean:74,86,123`) are the vault's own non-vacuity exploit demonstrations.

## 7. Statement issues seen

- **The `Pool.machine.Reachable s` hypothesis is dead weight in its 4 call sites.**
  `shield_can_happen` (`PoolSafety.lean:644`), `transact_can_happen` (`:660`), `unshield_can_happen`
  (`:728`) and `owned_transact_can_happen` (`:804`) all take `(_hr : Pool.machine.Reachable s)` —
  underscore-prefixed, i.e. Lean's own unused-variable convention, confirmed by `grep -n "_hr\b"
  ArmadaLean/PoolSafety.lean` returning exactly these 4 lines and no use of `hr` in any of the 4
  bodies. `PoolMachine.lean`'s own module doc comment states the file exists specifically so these
  theorems can take this hypothesis — but none of the 4 proofs consume it, and the bridge lemma
  built for it, `Pool.machine_reachable_iff`, is itself dead (§6). Either the hypothesis should be
  dropped from all 4 signatures (the proofs don't need reachability — `shield_can_happen`'s own doc
  comment says "No proof, no honest-prover assumption"), or a future proof was meant to need it and
  never arrived; this needs an operator decision, not a repair guess.
- **`Constrained.pathLen`/`.merkle` and `OwnedConstrained.pathLen`/`.merkle` are written but never
  read downstream**, unlike their sibling fields `outputs`/`rangeIn`/`rangeOut`/`balance`/`nullifiers`
  (all have `users > 0`). `Transact.lean:144,148` and `TransactOwned.lean:126,130` are genuine
  circuit constraints (`Constrained`'s whole point is "the accepted proof's witness satisfies every
  circuit constraint", `PoolOwned.lean:9-10`), so this is not dead code to delete — but no theorem
  in scope currently extracts a guarantee from the Merkle-path-length or Merkle-membership fields
  specifically, which is worth an operator's attention if a future proof needs them.
- **Two of `Transact.lean`'s own three headline theorems are never invoked by `Pool.lean`/
  `PoolSafety.lean`.** `transact_public_outputs_determined` (`Transact.lean:173`) and
  `transact_inputs_distinct_nullifiers` (`:189`) are named in the module's own doc comment
  (`Transact.lean:7-10`, "Theorems: …") as deliverables, `git grep` confirms zero external call
  sites. `TransactOwned.user_npk_ne_contract_npk` (`TransactOwned.lean:189`) is similarly unused.
  [INFERENCE] These may be genuinely redundant now that `PoolSafety.lean` derives the same facts a
  different way (`transact_spends_only_owned` via `commitment_mpk_inj` directly), or they may be
  gaps where a planned caller never arrived; distinguishing the two needs an operator or maintainer
  call, not a grep.
- **The `writing-lean.md` §4 `HoldingWellFormed`/`holdingWellFormed_step` example is not on trunk.**
  `writing-lean.md:145-158` cites `Lemmas.lean:3997-4004`/`:4741` for the Pool's canonical
  "state carries its own well-formedness, not a re-proved predicate" defect. `git grep -rn
  HoldingWellFormed` on the audit worktree (trunk `integrated`) returns nothing; the declaration
  lives only in `ArmadaLean/Adjudicator/Lemmas.lean` on the `AFDudley/armada-lean` fork's
  `adjudicator-model-v4` branch (commit `ce6f4ad`, 2026-10-26; deleted by `3feb0b0` on the same
  fork, also 2026-10-26) — which `writing-lean.md:12` itself labels "now reference only". That
  commit is not an ancestor of trunk `HEAD` (`522e16d`) or of the doctrine's own measurement commit
  `afda79b` (`git merge-base --is-ancestor 3feb0b0 HEAD` fails; `git ls-tree -r afda79b` has no
  `ArmadaLean/Adjudicator/` directory). The doctrine's own §4a text ("Pool: `HoldingWellFormed`
  re-proved for every call by `holdingWellFormed_step` … `writing-lean.md` §4") is citing a
  pedagogical, non-trunk example as if it measured the current Pool. The trunk's actual analog is
  `PoolSafety.HeldInvariant`/`core_invariant` (`PoolSafety.lean:11-17,167-171`), which already
  follows the rule writing-lean.md §4 asks for: one invariant, stated once, proved by one induction
  over `Reachable` — `writing-lean.md:194-196` says so explicitly ("`core_invariant`
  (`PoolSafety.lean:132`) follows the rule"). This slice's §4a "Pool" violations (§4 above) are real
  and independently confirmed by `facts.json` and `TypeLinters`' linter, but they are NOT instances
  of the `HoldingWellFormed` pattern; citing `writing-lean.md` §4 for them is a documentation
  mismatch worth correcting before the doctrine goes to operator review.
- **`facts.lean` tooling defect, found and fixed mid-audit.** The first `facts.json` computed
  `ci.value?` with the Lean-core default `allowOpaque := false`, which returns `none` for every
  `theorem` declaration (`Declaration.lean:483-486`), so `users`/`pool_base_ops`/`nitro_base_ops`/
  `sorry` were silently wrong (always 0/false) for all 187 theorems in this slice —
  `AuditNitroLiveness` found and reported the root cause; `Main` regenerated `facts.json` with
  `allowOpaque := true` mid-audit. This report's §4/§6 figures are drawn from the regenerated file,
  cross-checked against `TypeLinters`' independently-implemented `layerDiscipline` linter (same 12
  names, §4) and against direct `git grep`/source reads for every dead-code and layer-violation
  claim (§4, §6) — not from `facts.json` alone.

## 8. Plan

Ordered by layer, lowest first. Units sized for one worker each.

1. **`commitment_key_inj` unification** (§2b). Scope: `Transact.lean`, `TransactOwned.lean`. Add one
   lemma (in `Nullifier.lean` or a new small shared file both already import), replace
   `commitment_mpk_inj`/`commitment_owner_inj`'s bodies with calls to it. Acceptance: both files
   build, no other file's call sites change (both names can stay as thin wrappers, or every caller
   switches — operator's choice on whether to keep the two names). No dependency on other units.

2. **`value_conserved_of` unification** (§2a). Scope: `Transact.lean`, `TransactOwned.lean`. Same
   shape as unit 1; depends on `ValueConservation.no_wraparound_minting` and `RangeCheck
   .rangeCheck120_sound`, both already stable. No dependency on unit 1 but share the same worker's
   context cheaply.

3. **Refine `pathElements`/`valueIn`/`valueOut`/`leavesIndices` as typed fields** (§3b/c/d). Scope:
   `Transact.lean`, `TransactOwned.lean`, and every `Pool.lean`/`PoolSafety.lean` call site that
   constructs a `TransactWitness`/`TransactOwnedWitness` or reads these fields (`Pool.lean:991-1037`,
   `PoolSafety.lean:590-637,728-793,804-863`). This is the largest unit: it touches every `can_happen`
   theorem's witness construction. Depends on units 1-2 being merged first (fewer call sites to
   rebase across). Needs an operator decision only if the refinement changes `Constrained`'s/
   `OwnedConstrained`'s field order (API break for any out-of-scope caller — none exist in this
   slice, confirmed by git grep in §6/§7).

4. **`markSpent_succeeds` per-call wrapper** (§4, clears 8/12 violations). Scope: `PoolSafety.lean`.
   Add one lemma per `Step` constructor ("after a successful `markSpent` call inside `Step.transact`/
   `.transactOwned`, the result state's `spent` set is exactly as built") next to the existing
   `step_rootsAt_shape`/`step_leavesAt_*` frame lemmas (`PoolSafety.lean:441-491`), then rewrite
   `transact_spend_can_happen`/`transact_can_happen`/`unshield_can_happen`/
   `owned_transact_can_happen`/`owned_merge_can_happen`/`PoolOwned.owned_spend_can_happen` to call
   it instead of `markSpent_succeeds` directly. Depends on unit 3 (shares the same witness-shape
   rewrite work) — dispatch after unit 3 merges, same worker ideally, to avoid two rebases of the
   same 6 theorems.

5. **`live_spent_invariant_step_of` + `transactOwned_new_note_live`** (§2c/d, clears the 559-line and
   174-line declarations and the 114-line `step_live_leaf_matches`). Scope: `PoolSafety.lean:1098-
   1656` (`step_preserves_live_spent_invariant`), `:1691-1804` (`step_live_leaf_matches`), `:2124-
   2297` (`owned_merge_can_happen`). Depends on unit 4 merging first (both touch the same
   `transact`/`transactOwned` arms; doing 4 first means unit 5's diff is smaller). This is the
   doctrine's named Pool symptom (§4a: "`step_preserves_live_spent_invariant`... the only middle
   layer this slice needs"); it is the highest-value single unit in this plan.

6. **`foldl_preserves` reuse in `applyRecords_preserves_consistency`** (§2e). Scope: `Pool.lean:498-
   514`. Independent of units 1-5; can run in parallel with any of them, including unit 5 (touches a
   different theorem).

7. **`shield_creates_exactly`/`shield_can_happen` frame-lemma cleanup** (§4, the 2 remaining
   violations not covered by unit 4). Scope: `PoolSafety.lean:404-438,639-648`. Independent, can run
   in parallel with unit 6.

8. **Dead-code removal** (§6): `Field.p_pos`, `Pool.withPositionsFrom_nodup` (after folding it into
   `withTreePositionsFrom_nodup`'s proof, not just deleting — see §2's missing-API note),
   `PoolMachine.Pool.machine_reachable_iff`, `PoolOwned.preimageOf_consistent` (fold into unit 6's
   `foldl_preserves` call instead of deleting outright, since it is the correct per-transaction
   lemma `pool_preimageOf_consistent` should have called), `Poseidon.Term.val_ne_hash`,
   `Vault.attackAttackerShares`/`attackVictimShares`/`attackAttackerOut`. Independent of units 1-7;
   lowest priority, do last or in parallel as spare capacity allows.

9. **Operator decisions needed before any of the above can close out fully**: (a) whether the
   `_hr : Pool.machine.Reachable s` hypothesis stays, is dropped, or gets consumed by a yet-unwritten
   proof (§7); (b) whether `transact_public_outputs_determined`/`transact_inputs_distinct_nullifiers`/
   `user_npk_ne_contract_npk` are retired or still have a planned caller (§7); (c) whether the
   doctrine document's §4a Pool citation of `writing-lean.md` §4 should be corrected to cite
   `core_invariant` as the Pool's example of the rule being FOLLOWED, since `HoldingWellFormed` is
   not on trunk (§7). None of units 1-8 are blocked on these decisions; they are reporting
   obligations for whoever merges this report into the doctrine's operator-review queue.

**Dependencies on other slices**: none of units 1-8 touch `ArmadaLean/Nitro/`. The `draft/refined-
types` branch (`TimedChannel`/`LedgerChannel`/`TurboOutcomeNetwork`/`FixedPart.RunsConsensus`) is
Nitro-only and does not overlap this slice's refined-type candidates (§3) — those are new Pool-side
types with no existing branch work to duplicate. `Vault.lean` is excluded from units 1-8 entirely:
it is being remodelled on `draft/vault-locked-shares`, and this audit found no internal-layering
issue in it beyond the 3 dead defs in unit 8 (which that branch may already address — check before
dispatching).
