# Nitro-safety audit: `ArmadaLean.Nitro.Lemmas`, `ArmadaLean.Nitro.Safety`, `ArmadaLean.Nitro.NonVacuity`

Worktree: `/home/rix/code/git_puller/repos/armada-lean/.scratch/audit-wt` at `522e16d`. Read-only. No file under
`ArmadaLean/`, `docs/`, or `scripts/` was edited.

Scope: `ArmadaLean/Nitro/Lemmas.lean` (220 declarations, 2100 lines), `ArmadaLean/Nitro/Safety.lean` (70
declarations, 1272 lines), `ArmadaLean/Nitro/NonVacuity.lean` (239 declarations, 2674 lines).

## Counts

| Section | Count |
|---|---|
| 1. Layer map | 12 `Lemmas.lean` sections classified; `Safety.lean` flat (1 file, 5 layers mixed); `NonVacuity.lean` 3 parts (1 dead) |
| 2. Missing layer API | 3 findings (1 existing-but-unused, 1 proposed new, 1 witness-builder set) |
| 3. Refined-type candidates | 1 (cross-references `draft/refined-types`) |
| 4. Layer violations | 9 theorems in `Safety.lean` |
| 5. Size (>60 lines) | 4: 1 in `Lemmas.lean`, 3 in `NonVacuity.lean`, 0 in `Safety.lean` |
| 6. Dead code | 4 findings, 3 in `Lemmas.lean` (85 lines) + 1 in `NonVacuity.lean` (524 lines, 16 declarations) |
| 7. Statement issues | 3 |
| 8. Plan | 7 units |

A tooling defect affected the first pass: `AuditNitroLiveness` found `.scratch/audit/facts.lean` called
`ConstantInfo.value?` without `allowOpaque := true`, so every `theorem`'s proof term read as empty. `users`,
`nitro_base_ops`, `pool_base_ops` and `sorry` were silently wrong for every theorem (not for `def`s, and not for
`cites`/`prop_binders`/`data_binders`/`line`/`lines`, which read the type or doc string). `Main` regenerated
`facts.json` with the fix mid-audit. This report uses the regenerated file throughout; every `users`/`nb`/`pb`/
`sorry` figure below is from the corrected `facts.json`, cross-checked by `git grep` across the whole repo (not
just `ArmadaLean/Nitro/`) for every dead-code and layer-violation claim, because even the corrected `nitro_base_ops`
over-counts (§7).

## 1. Layer map

### `ArmadaLean/Nitro/Lemmas.lean` (12 sections, each mapped to §4a's Nitro order)

"Public API" below means a declaration with at least one use outside `Lemmas.lean` itself (`git grep`, word-
boundary, across `ArmadaLean/Nitro/*.lean`; `facts.json`'s `users` undercounts `simp`-set uses, so this is the
more reliable count for cross-module reference).

| Lines | Section | Layer | Public API (sample, by use count) |
|---|---|---|---|
| 16-234 | Addresses (`Balances`) | data-structure API, below base operations | `FixedPart.address_eq` (254 uses), `Balances.get_add`/`get_sub` (39/21) |
| 235-1052 | Outcomes: `Allocation.total`/`find`/`dedup`/`payout`/`paidTo`/`reduceOwed`/priority position/`addTo`/`remove` | data-structure API, below the redistribution layer | `Allocation.find_remove` (15), `owed_addTo_ge` (9), `paidTo_of_getElem?` (9) |
| 1053-1320 | Channels (`ChannelStatus`) | per-channel finalization API (reading lemmas for `finalOutcome`/`finalizedWith`/`mode`) | `ChannelStatus.finalizedWith_challenged` (47), `mode_eq_finalized_iff` (18) |
| 1321-1725 | Adjudicator: `statusOf_setStatus`/`deposit_some`/`transfer_some`/`reclaim_some`/`challenge_some`/`checkpoint_some`/`conclude_some`/`apply_*`/`apply_now_le`/`apply_balances_frame`/`apply_held_frame`/`finalizedWith_apply` | per-call exact-effect and frame lemmas (doctrine's own layer-2 name) | `statusOf_setStatus` (211), `transfer_some` (70), `finalizedWith_apply` (16). Two members, `apply_balances_frame`/`apply_held_frame`, have 0 uses (§2). |
| 1726-1855 | The coins-held invariant (`HeldInv`, `heldInv_step`, `heldInv_reachable`, `total_step_*`, `held_paidOut_step_*`) | global state invariant, one level above per-call lemmas | `heldInv_reachable` (6, used by `Safety.lean` and `Reachable.lean`) |
| 1856-2100 | One state ahead of another under transfers (`Adjudicator.Ahead`, `dueTotal`, `owedBy`, `transfer_effect`) | constructions layer (liveness-specific; feeds `Liveness.lean` only) | `Ahead` (11), all uses in `Liveness.lean` |

The file is laid out in dependency order and each section's layer is internally consistent. The only layer
mismatch is the unused `apply_balances_frame`/`apply_held_frame` pair (§2) and `finalizedWith_stable` (§6),
built in the per-call layer but never consumed one level up.

### `ArmadaLean/Nitro/Safety.lean` (70 declarations, no `/-! ## -/` section headers)

The doctrine's own layer order (§4a: per-call effect → per-channel finalization → redistribution →
constructions → paper theorems) names four distinct layers. `Safety.lean` has no file-level or section-level
separation between them; all four are interleaved in declaration order:

- Per-call effect wrappers that restate `Lemmas.lean`'s `*_some` lemmas with a Nitro citation: `deposit_effects`
  (65-82), `transfer_effects` (339-372), `challenge_effects` (831-843), `reclaim_effects` (1136-1191).
- Redistribution: `RedistributeStep`, `Redistributed`, `redistribution_conserves_total`,
  `redistribution_ignores_caller` (453-499).
- Per-channel finalization: `FinalizedTwoWays`, `StatusOrigin`, `statusOrigin_*`, `finalized_two_ways`,
  `finalized_permanent`, `finalized_outcome_only_redistributed` (501-768).
- Paper (spec) theorems proper, citing a Nitro/ForceMove/SATP line number each: the rest (deposit/transfer/
  challenge/checkpoint/conclude/reclaim `_requires`/`_can_happen`/`_fails_*`, `turnNumRecord_monotone`,
  `forceMove_*`, `mover_*`).

Public API (used outside `Safety.lean`, by `git grep`): only `Allocation.aheadOf` (3, within `Safety.lean`'s own
later theorems) and the handful cited by `NonVacuity.lean`'s doc comments (not code uses). This is expected:
`Safety.lean` is the paper-theorem terminus of the Nitro proof, so nothing composes further on it inside
`ArmadaLean/Nitro/`. 53 of 70 declarations have `users: 0`; all but 10 of those carry a Nitro/ForceMove/SATP
citation (§6 confirms the other 10 are not dead).

### `ArmadaLean/Nitro/NonVacuity.lean` (239 declarations, 2674 lines)

The module's own header (lines 1-12) states its sole purpose: one `example` per `Safety.lean`/`Liveness.lean`
theorem, proving its non-`binder-type` hypotheses are jointly satisfiable. The file in fact has three parts:

| Lines | Content | Matches the header? |
|---|---|---|
| 19-358 | Shared values (`p0`, `c1`, `g0`, `sDep`, `sCon`, `sR`, …) + one `example` per `Safety.lean` theorem | Yes |
| **411-934** | **`section Game`: `ConcludeInv`, `concludeInv_apply`, `concludeRound`, `conclude_strategy`, `WaitInv`, `waitInv_apply`, `waitRound`, `wait_strategy`, `zeroBal_apply`, `zeroBalRound`, `Threat.RoundSeq.append`, `round_nil`, `value_zero`, `finalizedWith_applyAdv`, `finalizedWith_applyHonest`, `prevents_other_outcome`** | **No — real strategy-proof theorems, not witnesses (§6: also dead)** |
| 935-2674 | Reachability helpers (935-999, legitimate shared combinators for `Participant`/`Network`-level `Reachable` chains) + one `example` per `Liveness.lean`/`SATP.lean` theorem (1003-2674) | Yes |

§6 shows `section Game` (lines 411-934) is not only mis-filed but entirely unused outside itself.

## 2. Missing layer API

**Finding A — the per-call frame API exists and is unused.** `Lemmas.lean:1599-1658` already defines
`Op.touchesBalances`, `Op.touchesHeld`, `apply_balances_frame`, `apply_held_frame` — exactly the
`writing-lean.md` §6 pattern ("one footprint predicate per field group"). `git grep` finds zero uses of
`apply_balances_frame`/`apply_held_frame` anywhere in the repository. Instead, four `Safety.lean` theorems
redo the same 7-constructor `cases op with` split by hand to reach the same "field unchanged outside the
footprint" fact: `held_falls_only_by_payout` (203-239, 37 lines), `channel_balance_falls_only_by_redistribution`
(241-257, 17 lines), `paid_out_monotone` (259-279, 21 lines), `held_changes_only_by_deposit_payout` (1257-1270,
14 lines). Checked (`.scratch/audit/safety_frame_check.lean`, `lean-run`): `apply_balances_frame` alone
discharges the four non-footprint cases of `channel_balance_falls_only_by_redistribution` in one `rw`, in place
of the theorem's four separate `cases op with` branches.

**Finding B — a missing `statusOf`-effect lemma.** Four more `Safety.lean` theorems each redo a full
7-constructor `cases op with` split to learn what one step does to `s.statusOf c`: `finalized_permanent`
(666-709, 49 lines), `finalized_outcome_only_redistributed` (714-768, 58 lines), `turnNumRecord_monotone`
(875-926, 52 lines; named in `writing-lean.md` §5 as the doctrine's own example of this defect, under its old
name `turn_number_record_monotone`), and `statusOrigin_step` (564-617, 54 lines, in `Safety.lean` itself — this
one is the legitimate single build, §4a rule 3). `Lemmas.lean:1667-1711` (`finalizedWith_apply`) does the same
split a fifth time for a related but narrower fact (`finalizedWith` only grows). None of the five shares the
case split with another. Proposed lemma, checked to type-check (`.scratch/audit/safety_layer_check.lean`,
`sorry`-bodied, `lean-run`):

```lean
/-- The exact effect of one step on channel `c`'s status: unchanged, or the
    step redistributes `c`'s finalized outcome (`reduceOwed`/`addTo`+`remove`),
    or the step is a challenge/checkpoint/conclude of `c` with the stated new
    status. Replaces the `cases op with` split repeated in `finalized_permanent`,
    `finalized_outcome_only_redistributed`, `turnNumRecord_monotone`,
    `statusOrigin_step`, and the narrower `finalizedWith_apply`. -/
theorem Adjudicator.statusOf_step {apps : Apps E} {s s' : Adjudicator E} {caller : Participant}
    {op : Op E} (h : s.apply apps caller op = some s') (c : FixedPart) :
    s'.statusOf c = s.statusOf c ∨
    (∃ o b x, (s.statusOf c).finalOutcome s.now = some (.allocation o) ∧
      s'.statusOf c = (s.statusOf c).withOutcome (.allocation (o.reduceOwed b x))) ∨
    (∃ oL left x1 x2 right, (s.statusOf c).finalOutcome s.now = some (.allocation oL) ∧
      s'.statusOf c =
        (s.statusOf c).withOutcome (.allocation (((oL.addTo left x1).addTo right x2).remove c.address))) ∨
    (∃ proof cand, s.apply apps caller (.challenge c proof cand) = some s' ∧
      s'.statusOf c = .challenged cand.turnNum ⟨cand, s.now + c.challengeDuration⟩) ∨
    (∃ proof cand, s.apply apps caller (.checkpoint c proof cand) = some s' ∧
      s'.statusOf c = .open cand.turnNum) ∨
    (∃ ss, s.apply apps caller (.conclude c ss) = some s' ∧
      s'.statusOf c = .finalized ss.state.turnNum ss.state.outcome ss.state.outcome)
```

`heldInv_step` (`Lemmas.lean:1736-1765`, 30 lines) and `total_step_no_deposit` (`Lemmas.lean:1792-1818`, 27
lines) do the same 7-constructor split a sixth and seventh time, for `HeldInv`/`balances.total`. These two are
the legitimate single build of that layer (nothing in-scope repeats them), but they are further evidence of
how often this exact split recurs: 7 times across `Lemmas.lean` and `Safety.lean` combined for closely related
"what does one step do to field X of channel `c`" facts, with no shared case-split helper between them.

**Finding C — `NonVacuity.lean`'s witness values repeat the same literal shapes.** `git grep` counts, scoped to
`NonVacuity.lean`:
- 15 `FixedPart` literals of the exact shape `⟨participants, nonce, appId, 2, kind⟩` (`c1`, `c3`, `c2`, `cL`,
  `cV`, `cLW`, `cχ`, `cLWF`, `cχF`, `cL2F`, `Lr`, `Vr`, `Lm3`, `Lp3`, `V3`): every one fixes
  `challengeDuration = 2`, and `appId` is always `⟨0⟩` (ledger) or `⟨1⟩` (virtual, except `V3`'s own `⟨1⟩` too).
  Only `participants` and `nonce` vary.
- 34 `State` literals of the shape `⟨c, n, o, fin, .opaque 0⟩`: every one fixes the SATP resolution field to
  `.opaque 0` (no witness in scope needs a SATP resolution).

Proposed builders, checked against four existing defs by `rfl` (`.scratch/audit/nonvac_builder_check.lean`,
`lean-run`, all four pass):

```lean
/-- A ledger channel for the non-vacuity witnesses: challenge duration 2, app id 0. -/
def mkLedger (ps : List Participant) (nonce : Nat) : FixedPart := ⟨ps, nonce, ⟨0⟩, 2, .ledger⟩

/-- A virtual channel for the non-vacuity witnesses: challenge duration 2. -/
def mkVirtual (ps : List Participant) (nonce appId : Nat) : FixedPart := ⟨ps, nonce, ⟨appId⟩, 2, .virtual⟩

/-- A non-SATP state (`.opaque 0` resolution) for the non-vacuity witnesses. -/
def mkState (c : FixedPart) (n : Nat) (o : Outcome Nat) (fin : Bool) : State Nat := ⟨c, n, o, fin, .opaque 0⟩
```

`example : mkLedger [p0] 0 = c1 := rfl`, `mkVirtual [p0] 6 0 = cV := rfl`, `mkState c1 0 (.allocation o1) true =
stF := rfl`, `mkState c1 1 (.allocation o1) false = stO := rfl` all check.

## 3. Refined-type candidates

`l.kind = .ledger`/`v.kind = .virtual`/`c.kind = .ledger` is a loose one-input `Prop` hypothesis on four
`Safety.lean` theorems: `deposit_can_happen:90`, `deposits_in_order_succeed:154`, `transfer_can_happen:387`,
`reclaim_requires:1119` (conclusion) and `reclaim_can_happen:1196`. This is the same gap `typed-library.md §4`
already names (`LedgerChannel := a channel with kind = .ledger`) and `draft/refined-types` already builds
(`.scratch/types-wt`). This finding does not duplicate that branch; it confirms the branch's scope reaches
these five `Safety.lean` sites too, and that converting them is in scope for whichever unit of work merges
`draft/refined-types`'s `LedgerChannel` into `ArmadaLean/Nitro/`.

No other hypothesis in `Lemmas.lean`/`Safety.lean`/`NonVacuity.lean` is shared by two or more declarations in a
way `draft/refined-types` does not already cover (`TimedChannel`'s `2 * ε ≤ challengeDuration` appears once, at
`Safety.lean:836`, single-use, not repeated in this scope — the 98 Liveness.lean instances `typed-library.md`
counts are out of this slice).

## 4. Layer violations

Using `nitro_base_ops` is unreliable even on the corrected `facts.json`: it is 1 for nearly every declaration
that touches `Adjudicator`/`Op` at all (confirmed: `nb=1` on 44 of `Lemmas.lean`'s 220 declarations and on 44 of
`Safety.lean`'s 70, including trivial `@[simp]` one-liners like `apply_deposit`), because unifying a hypothesis
pattern like `obtain ⟨…⟩ := Adjudicator.transfer_some h` pulls the constant `Adjudicator.transfer` into the
elaborated proof term even when the proof does not itself descend into `transfer`'s definition. This finding
therefore uses the manual reading from §2 instead: a spec-level (`cites: true`) `Safety.lean` theorem whose
proof is a full case split over all seven `Op` constructors, in place of composing a layer lemma that already
exists or should exist one level up.

| Theorem | Lines | What it redoes | Replace with |
|---|---|---|---|
| `held_falls_only_by_payout:203` | 37 | `cases op with` over 7 constructors for the `held`/`balances` footprint | `apply_held_frame`/`apply_balances_frame` for the 4 non-touching cases |
| `channel_balance_falls_only_by_redistribution:241` | 17 | same | same |
| `paid_out_monotone:259` | 21 | same | same |
| `held_changes_only_by_deposit_payout:1257` | 14 | same | same |
| `finalized_permanent:666` | 49 | `cases op with` for `s'.statusOf c` | `Adjudicator.statusOf_step` (§2 Finding B) |
| `finalized_outcome_only_redistributed:714` | 58 | same | same |
| `turnNumRecord_monotone:875` | 52 | same | same |
| `funding_decoupled_from_rules:1016` | 21 | `cases op with` to show a different `balances` value gives the same `status` | partial: the 3 dispute-op cases (`challenge`/`checkpoint`/`conclude`) already `simp` away the `balances` argument because the per-call lemmas never read it; only `deposit`/`transfer`/`reclaim` (already contradiction-closed by `hop`) need the split |
| `turnNumRecord_monotone:875` (listed once above) | — | — | — |

9 theorems (8 distinct, `turnNumRecord_monotone` counted once), 269 lines total, none of them citing a base
operation in the doctrine's own 7-name list literally in their *statement* type (the doctrine's list omits
`Adjudicator.deposit`, §7) but every one descending to full per-operation case analysis to reach a fact one
layer up (the per-call frame/effect layer, or the per-channel-status layer) should already supply.
`redistribution_conserves_total:461` (26 lines) is the better existing pattern in the same file: it uses a
`cases op with | transfer … | reclaim … | _ => simp [Op.isRedistribution] at hop` catch-all instead of
enumerating all 7 constructors, closer to what Finding A's frame lemmas would give directly.

No violation found by the same manual check in `Lemmas.lean`'s `Adjudicator` section (1321-1725): its
`cases op with` splits (`heldInv_step`, `total_step_no_deposit`, `finalizedWith_apply`,
`apply_balances_frame`, `apply_held_frame`) are each the single, lowest-layer build of their respective fact
(doctrine §4a rule 3) — the defect is that `Safety.lean` does not reuse two of them (Finding A) and that a
sixth/seventh instance of the same split pattern (Finding B) was never factored out as its own layer member.

## 5. Size

| Declaration | Lines | Cause |
|---|---|---|
| `Lemmas.lean:1239 forceMoveRules_noUnilateralSupport` | 61 | Dead (§6): the size is not a legitimate proof cost, the whole declaration (plus its 17-line helper `turn_mod_exists`) has 0 users |
| `NonVacuity.lean:497 concludeRound` | 101 | Dead (§6): inside the orphaned `section Game` |
| `NonVacuity.lean:422 concludeInv_apply` | 74 | Dead (§6): same section, two duplicated 32-line case splits (lines 432-461 and 463-494 are near-identical, differing only in a deadline-not-expired side condition) |
| `NonVacuity.lean:2029 nTwoCh_finalizable_latest_snd` | 68 | Not dead (`users: 1`, a live `Liveness.lean`-section witness after line 1003); case split per constructor of a two-channel network status combined with an inline re-derivation of "the challenge deadline has not passed" already proved elsewhere in the file for the one-channel case (`netF`-style witnesses, 1096-1300) — not instantiated through a shared two-channel builder |

`Safety.lean` has no declaration over 60 lines; its size defect is breadth (9 separate ~20-55 line theorems
doing the same case split, §4) rather than any one large declaration.

## 6. Dead code

Checked with `git grep` across the whole repository (`docs/`, `scripts/`, `ArmadaLean/`), not just
`ArmadaLean/Nitro/`.

1. **`NonVacuity.lean:411-934` (`section Game ... end Game`), 524 lines, 16 declarations**: `ConcludeInv`,
   `concludeInv_apply`, `concludeRound`, `conclude_strategy`, `WaitInv`, `waitInv_apply`, `waitRound`,
   `wait_strategy`, `zeroBal_apply`, `zeroBalRound`, `Threat.RoundSeq.append`, `round_nil`, `value_zero`,
   `finalizedWith_applyAdv`, `finalizedWith_applyHonest`, `prevents_other_outcome`. Zero uses anywhere outside
   this section, including inside `NonVacuity.lean`'s own witness examples. These are not witnesses (the module
   header's stated purpose); they are a full strategy-proof derivation (the "conclude now"/"wait" strategies
   are each proved "unbeatable", `Strategy.lean:522`'s own term). Confirmed dead on both sides:
   `Liveness.lean:791-812` (`conclusion_proof_finalizable`, a live, cited theorem) proves the equivalent "a
   known conclude call is unbeatable" fact through differently-named helpers (`conclude_round_finalized`,
   `round_keeps_status`, not found in `NonVacuity.lean`); `AuditNitroLiveness` independently confirmed the
   wait side too (`agent://AuditNitroLiveness`, this session): `Liveness.lean`'s shipped
   `latest_state_finalizable`/`latest_state_finalizable_two` (2450-2517, 2609-2613) cover the same fact as
   `wait_strategy`/`WaitInv` through `latest_wait_keeps_status` (2215), `wait_resolves_challenge` (2407), and
   `latest_round_keeps_status` (1787) — none of which is `waitRound`/`concludeInv_apply`/`concludeRound`.
   [INFERENCE] `section Game` reads as an earlier draft of this proof, superseded by `Liveness.lean`'s own
   later implementation and left in place.
2. **`Lemmas.lean:1239-1299 forceMoveRules_noUnilateralSupport` (61 lines) and its sole helper
   `Lemmas.lean:1221-1237 turn_mod_exists` (17 lines)**: 0 uses anywhere. The sibling fact,
   `consensusRules_noUnilateralSupport` (`Lemmas.lean:1210`), *is* used, twice, by `NonVacuity.lean` witnesses
   for `Liveness.lean` theorems that take `(apps c.appDef).NoUnilateralSupport` (`Liveness.lean:4052-4073`).
   No `Liveness.lean` theorem or `NonVacuity.lean` witness ever instantiates that hypothesis with
   `forceMoveRules vt` instead of `consensusRules`, so the `forceMoveRules` half of this fact, though proved,
   is never consumed.
3. **`Lemmas.lean:1717-1729 finalizedWith_stable` (13 lines)**: 0 uses. Its own doc comment claims it is "used
   by `Adjudicator.reclaim`'s channel-types repair" and by `Reachable.lean`'s wallet-conservation theorem;
   neither use exists in the current tree (`docs/nitro-repair-2026-10-27.md:791-794` shows it replaced a
   removed `Safety.lean` theorem, `virtual_finalized_outcome_immutable`, on 2026-10-27, and was built "a proof
   helper, not a `Safety.lean` theorem, with no `NonVacuity.lean` witness of its own" — the consuming call site
   that justified building it is missing).
4. **`Lemmas.lean:276-278 total_take_le`** (6 lines): 0 uses, no citation, not `@[simp]`.

No dead code found in `Safety.lean`: all 53 of its `users: 0` declarations either carry a Nitro/ForceMove/SATP
citation (43) or are the paper theorems matched 1:1 by `docs/nitro-theorems.md`'s own catalog (10 more:
`balances_total_split`, `payout_bounded`, `reduceOwed_owed`, `finalized_outcome_only_redistributed`,
`challenge_effects`, `tick_moves_only_time`, `mode_of_absent`, `held_changes_only_by_deposit_payout`, plus
`RedistributeStep.casesOn`/`.recOn`, Lean-generated, not real declarations).

## 7. Statement issues seen

1. **Tooling defect, confirmed and already fixed mid-audit** (see header): `.scratch/audit/facts.lean`
   originally called `ConstantInfo.value?` without `allowOpaque := true`, zeroing `users`/`nitro_base_ops`/
   `pool_base_ops`/`sorry` for every `theorem`. `Main` regenerated `facts.json`; this report used only the
   regenerated file.
2. **`nitro_base_ops` over-counts even after the fix.** It is 1 for almost any declaration that elaborates a
   hypothesis via `obtain ⟨…⟩ := Adjudicator.transfer_some h`-style destructuring, because the per-call lemma's
   *instantiated type* (not the base operation's own definition) pulls the base-operation constant into the
   proof term's used-constants set. §4's violation list was built by manual reading, not by filtering on this
   field. [INFERENCE] worth fixing in `facts.lean` for the next audit pass, e.g. by checking the proof term's
   *head* constants at each `match`/`cases` discriminant rather than its full used-constants set.
3. **The `cites` heuristic fires on provenance, not source.** `cites` matches any doc comment containing
   "operator ruling" (`facts.lean:19`), which also matches comments that merely *attribute* a declaration to an
   operator ruling without citing a paper/contract line for *that theorem's own statement* — e.g.
   `Lemmas.lean:1717 finalizedWith_stable`'s doc comment says "Used by `Adjudicator.reclaim`'s channel-types
   repair (operator ruling 2026-10-27)", which is a provenance note, not a source citation, yet `cites: true`.
   This inflated §1/§4's citation counts by an unmeasured but small amount; every citation this report relies on
   for a spec-theorem classification was independently confirmed in `docs/nitro-theorems.md` (§6).

## 8. Plan

Layer order (lowest first). Each unit is sized for one worker.

1. **Wire in the existing frame lemmas (`Lemmas.lean:1609-1658`) at their four `Safety.lean` call sites.**
   Scope: `held_falls_only_by_payout`, `channel_balance_falls_only_by_redistribution`, `paid_out_monotone`,
   `held_changes_only_by_deposit_payout` (§2 Finding A, §4). Inputs: `apply_balances_frame`, `apply_held_frame`
   (unchanged). Acceptance: all four theorems' statements unchanged; each proof's non-footprint cases collapse
   to one `rw`/`simp` using the frame lemma; `lake env lean` (or the spec check) passes. No dependency.
2. **Add `Adjudicator.statusOf_step` to `Lemmas.lean`'s Adjudicator section and rebuild
   `finalized_permanent`/`finalized_outcome_only_redistributed`/`turnNumRecord_monotone`/`statusOrigin_step`/
   `finalizedWith_apply` on top of it** (§2 Finding B, §4). Scope: one new lemma plus five call-site rewrites.
   Acceptance: statements of the five unchanged; each proof becomes a `rcases`/`simp` on `statusOf_step`'s
   result instead of its own `cases op with`; `lake env lean` passes. Independent of unit 1 (touches a
   disjoint set of `Safety.lean` theorems), can run in parallel.
3. **Delete the dead declarations in `Lemmas.lean`**: `total_take_le` (276-278),
   `forceMoveRules_noUnilateralSupport` + `turn_mod_exists` (1221-1299), `finalizedWith_stable` (1712-1729)
   (§6). Acceptance: `lake env lean` passes with no new warnings; `git grep` for each name (whole repo) stays
   empty. No dependency; can run before, after, or parallel to units 1-2 (disjoint lines).
4. **Delete `NonVacuity.lean`'s `section Game` (411-934)** (§6, confirmed dead on both the conclude and wait
   side by `AuditNitroLiveness`'s independent check of `Liveness.lean`'s live call graph). Acceptance: the
   section is gone; `lake env lean` passes; `NonVacuity.lean`'s declaration count drops by 16. No dependency;
   can run in parallel with units 1-3 and 5.
5. **Add `mkLedger`/`mkVirtual`/`mkState` to `NonVacuity.lean` and convert the 15 `FixedPart` and 34 `State`
   literal `def`s to use them** (§2 Finding C). Acceptance: every converted `def` is `rfl`-equal to its old
   literal (as checked for 4 samples here); `lake env lean` passes. Independent of units 1-4.
6. **Convert the five `l.kind = .ledger`/`v.kind = .virtual`/`c.kind = .ledger` hypotheses in `Safety.lean`
   (`deposit_can_happen`, `deposits_in_order_succeed`, `transfer_can_happen`, `reclaim_requires`,
   `reclaim_can_happen`) to `LedgerChannel`/the virtual-channel equivalent** (§3). Depends on
   `draft/refined-types` merging its `Channel.lean` types first (`docs/lean/typed-library.md` §4 layout); not
   a unit this audit can scope further without that branch's exact API.
7. **Rewrite `nTwoCh_finalizable_latest_snd` (`NonVacuity.lean:2029`, 68 lines) through a shared two-channel
   deadline-not-passed helper**, once the one-channel version it duplicates is identified precisely (§5). Needs
   a closer read of the `netF`-style one-channel witnesses (1096-1300) than this audit gives; scope this as a
   follow-up read, not a blind refactor. Independent of units 1-6.
