# Audit: nitro-base

Slice: `ArmadaLean.Nitro.Channel`, `Address`, `Outcome`, `Adjudicator`, `Reachable`, `Strategy`,
`Constructions`, `SATP`, `Value`, `ArmadaLean.Machine`, `ArmadaLean.Machines`. 602 declarations
(`facts.json`, 2026-10-31 scan), plus `Machines.lean` (69 lines, a metaprogram, no
`facts.lean`-scanned declarations).

Counts: Layer map 11 modules. Missing layer API 2 groups (5 duplicated proof blocks, ~470
lines). Refined-type candidates 1 (deferred to `draft/refined-types`). Layer violations 4
confirmed (all trace to the §2a/§2b gaps). Size 8 declarations over 60 lines. Dead code 4
confirmed, 2 more flagged `@[simp]`/unread-structure-field caveats. Statement issues 2.

**Tooling defect, found and fixed mid-audit.** `AuditNitroLiveness` found that
`.scratch/audit/facts.lean` called `ConstantInfo.value?` with the default `allowOpaque := false`,
which returns `none` for every `theorem` (Lean core, `Declaration.lean:483-486`). Consequence:
`users`, `nitro_base_ops`, `pool_base_ops` and `sorry` were silently wrong for every `theorem` in
the first `facts.json` (a `theorem` never contributes an edge to anything, so `users` undercounts
cross-module and `@[simp]`-implicit callers, and `nitro_base_ops` was 0 for every theorem's own
proof regardless of its actual content). `prop_binders`/`data_binders`/`subtype_binders`/`cites`
are computed from the declaration's type or doc string, not `.value?`, and were correct
throughout. Main regenerated `facts.json` with `allowOpaque := true` partway through this audit.
**Every finding below uses the regenerated file.** Before the fix, §4 read "0 layer violations
measured" and §6 listed `NetworkState.applyAdv_deliver` as dead; both are corrected below. Every
dead-code and refined-type claim is additionally checked by `git grep` across the whole
repository, not by the `users` count alone — the regenerated field is reliable but grep remains
the independent check per the assignment's rules.

## 1. Layer map

| Module | Lines | Layer (doctrine §4a) | Public API (`users` > 0, `git grep`-checked) |
|---|---|---|---|
| `Machine.lean` | 47 | Generic machine layer, below Nitro's own base operations | `Machine`, `Machine.Reachable`, `Machine.invariant`, `Machine.Reachable.tail` — all used by `Reachable.lean` |
| `Machines.lean` | 69 | Build-time registry (a `#eval` scan over the checked environment), not a proof layer | none (a build artifact; `scripts/reach.py` reads its stdout, not a declaration) |
| `Address.lean` | 159 | Base data: addresses, coins, balances | `Address`, `FixedPart`, `Participant`, `Coins`, `Balances`, `Balances.get/add/sub/total`, `FixedPart.address`, `Address.fundable` — the widest-used file in the slice (`Coins` 522 users, `FixedPart` 554, `Address` 442, `Participant` 715) |
| `Outcome.lean` | 164 | Base data: outcomes, allocations, payout | `Outcome`, `Allocation`, `Entry`, `Item`, `Allocation.payout/paidTo/owed/find/dedup/reduceOwed/addTo/remove` — the priority-payout API every redistribution theorem (owned by `AuditNitroSafety`) composes |
| `Channel.lean` | 306 | Base data: channel state, signatures, application rules | `State`, `SignedState`, `ChannelStatus`, `Rules`, `forceMoveRules`, `consensusRules`, `ChannelStatus.mode/finalOutcome/finalizedWith/withOutcome` |
| `Adjudicator.lean` | 341 | Base operations (doctrine's own list: `deposit`, `transfer`, `reclaim`, `challenge`, `checkpoint`, `conclude`, `tick`, `apply`) | `Adjudicator`, `Adjudicator.apply/deposit/transfer/reclaim/challenge/checkpoint/conclude/finalOutcome/statusOf`, `Op` |
| `Outcome.lean`/`Address.lean`/`Channel.lean`/`Adjudicator.lean` together ARE the base layer the doctrine names for Nitro (§4a rule 1, first arrow) | | | |
| `Reachable.lean` | 1296 | Per-call exact-effect/frame lemmas, chain reachability, and network-level reachability (second and third doctrine layers folded into one file) | `Reachable`, `Nitro.network`, `Nitro.participant`, `Network`, `Network.view/views`, `Nitro.network_chain_reachable`, `Nitro.network_views_participant_reachable` (used by `Liveness.lean`, `NonVacuity.lean`) |
| `Strategy.lean` | 656 | Per-channel finalization layer (`NetworkState`, `HonestMove`/`AdvMove`, `Threat`, `UnbeatableIn`, `Finalizable`, `Enabled`, `FundedFor`, `Value`) | `NetworkState`, `Threat`, `Finalizable`, `UnbeatableIn`, `Views.Consistent`, `Threat.Round.single/exists`, `Threat.RoundSeq.peel_one/landAll` — the per-round machinery every Liveness theorem (owned by `AuditNitroLiveness`) composes |
| `Value.lean` | 648 | Redistribution layer: the Turbo Value Algorithm | `Adjudicator.TurboValue`, `Adjudicator.turboValueAlong`, `Adjudicator.IsDAG`, `exists_valid_topo_order` |
| `Constructions.lean` | 76 | Constructions layer (ledger-with-sub, "Outcomes First") | `IsLedgerChannel`, `LedgerWithSub`, `LedgerWithSub.Holds`, `Views.FinalizableBy`, `SameValue` |
| `SATP.lean` | 253 | Constructions layer, SATP's virtual-channel protocol | `Route`, `Route.WellFormed`, `satpVirtualRules`, `validChallenge`, `validResponse`, `HoldsLatestLedger`, `HoldsVirtualSetup` |

`Reachable.lean` and `Strategy.lean` each span two doctrine layers (base-call frame facts
together with the reachability/finalization machinery above them); they are not mis-layered —
every big theorem in `Reachable.lean` genuinely needs the chain-level base-operation facts and
the network-level composition in the same file — but their size (1296 and 656 lines) is itself
the finding in §5.

## 2. Missing layer API

**Two confirmed repetition groups**, found by reading every declaration over 60 lines and every
module's internal structure (not by the `repeatedPremise` linter, which targets theorem
hypotheses and these are proof-term repeats).

### 2a. `Reachable.lean`: no frame lemma for `Network.applyHonest`/`Network.land`/`Network.applyAdv`/`Network.deposit`

Five theorems each re-derive, by hand, "how does `(chain, sent, received, wallet)` change under
each of `Nitro.network`'s step constructors" and then apply `Machine.invariant` over that same
case split:

- `Nitro.network_received_signers` (`Reachable.lean:561-649`, 102 lines)
- `Nitro.network_received_voucher_signers` (`Reachable.lean:663-753`, 103 lines) — identical to
  the one above with one extra `v`/voucher argument threaded through every branch
- `Nitro.network_views_delivered` (`Reachable.lean:907-987`, 91 lines)
- `Nitro.network_chain_reachableSigned` (`Reachable.lean:790-895`, 116 lines)
- `Nitro.network_conservation` (`Reachable.lean:1197-1272`, 88 lines)

Each repeats this pattern, branch for branch (`.react`, `.wait`, `land`'s `none`/`some`,
`.deliver`, `.tick`, `.chain`/`.land` absurd cases, `.deposit`'s `none`/`some`):

```lean
have hsent : (Network.applyHonest n p₀ (.wait dt)).sent = n.sent := by simp [Network.applyHonest]
have hrecv : (Network.applyHonest n p₀ (.wait dt)).received p = n.received p := by
  simp [Network.applyHonest]
rw [hsent]; rw [hrecv] at hr
exact hP p q r hr hqr
```

(and the equivalent six-way split for `land`/`applyAdv`/`deposit`), proved from the ALREADY-NAMED
simp lemmas `Network.applyHonest_react_sent`, `Network.applyHonest_react_received`,
`Network.land_of_none`, `Network.land_of_some`, `Network.applyAdv_deliver`,
`Network.applyAdv_tick`, `Network.deposit_sent`, `Network.deposit_received`. The per-field simp
lemmas already exist (`writing-lean.md` §3's API); what is missing is the ONE lemma that
combines them into a single frame/exact-effect fact per step (`writing-lean.md` §6, "one frame
lemma per call"), so that every `Machine.invariant` user states its step case once, not five
times.

Proposed lemma, to replace the repeated case split (the exact return shape needs one choice per
field actually read by each of the five call sites — `chain`, `sent`, `received q`, `wallet p`):

```lean
/-- How one `Nitro.network` step changes `(chain, sent, received, wallet)`. The three "real"
    data fields are a function of the move alone; `received` additionally depends on which
    participant's copy is read. -/
theorem Nitro.network.step_frame {apps : Apps E} {Ts : Threats E apps} {n n' : Network E}
    (h : (Nitro.network apps Ts).step n n') :
    ∃ p₀, (∃ m : HonestMove E, ...) ∨ (∃ op, ...) ∨ (∃ m : AdvMove E, ...) ∨ (∃ c e x, ...)
```

[INFERENCE] The exact statement shape is a design choice for whoever does this unit of work (an
existential-per-branch restating `hstep`'s own four-way split is one option; a record of four
functions `n.chain → n'.chain`, `n.sent → n'.sent`, etc., each case-matched once, is another).
Either form removes roughly 350 of the 500 lines these five theorems currently spend re-deriving
the same case split (measured: 102+103+91+116+88 = 500 lines total; the per-branch "derive then
apply `hP`" business logic that is NOT shared is roughly 150 of those lines).

### 2b. `Strategy.lean`: `Threat.RoundSeq.peel_two` restates `peel_one`'s induction

`Threat.RoundSeq.peel_one` (`Strategy.lean:317-342`, 31 lines) proves: one call always peels off
a `RoundSeq` as the one and only `land` step. `Threat.RoundSeq.peel_two`
(`Strategy.lean:355-397`, 53 lines) proves the two-call case by **re-running the same induction
by hand**, with every branch of `peel_one` duplicated twice (once for "land lands `opA` first",
once for "land lands `opB` first" — `peel_two`'s own doc comment says "`peel_one`, applied to the
returned singleton `RoundSeq`, splits the rest into its own ordinary-play-then-land shape", but
the proof does not do that; it reimplements the split).

Missing layer API: a lemma that composes `peel_one` instead of reimplementing it — for example

```lean
theorem Threat.RoundSeq.peel_split {apps : Apps E} {p : Participant} {T : Threat E p} {t : Time}
    (pending1 pending2 : List (Op E)) (n0 n2 : NetworkState E p)
    (hseq : Threat.RoundSeq apps T t (pending1 ++ pending2) n0 n2) :
    ∃ n1, Threat.RoundSeq apps T t pending1 n0 n1 ∧ Threat.RoundSeq apps T t pending2 n1 n2
```

would give `peel_one` as `peel_split [op] []` composed with `RoundSeq.land`'s single-step case,
and `peel_two`'s result by applying `peel_split` at `[opA]`/`[opB]` or `[opB]`/`[opA]` — one
lemma proved once by induction on `pending1`, replacing the hand-duplicated case split. This is
the general form `writing-lean.md` §1 asks for ("a new definition calls an existing definition by
name"); `peel_two` as written is `peel_one`'s body, copied and re-split.

### 2c. (minor) `SATP.lean`: `validChallenge_outcome`/`validResponse_outcome` share one shape

`validChallenge_outcome` (`SATP.lean:227-236`, 10 lines) and `validResponse_outcome`
(`SATP.lean:241-250`, 10 lines) have the identical proof shape (`unfold`, `split`, `rename_i`,
`simp only [Bool.and_eq_true, decide_eq_true_eq] at h`, `exact ⟨_, _, _, _, hp, h.2, h.1.2⟩`) —
not over 60 lines, so not a §5 finding, but a genuine small duplication. Below the threshold
where a shared lemma clearly pays for itself; noted, not proposed as a unit of work.

## 3. Refined-type candidates

**None new in this slice beyond what `draft/refined-types` already covers.** The slice's own
theorems (85 of them) have 0-8 `Prop` binders each (`facts.json`), mostly single-use hypotheses
specific to one frame lemma (`h : net.chain.apply apps caller op = some s'`), not the
multi-theorem repetition the doctrine's `repeatedPremise` targets. `subtype_binders` is 0 across
the whole slice (matches the project-wide §1 measurement).

The one invariant this slice's own declarations disagree on is **`IsLedgerChannel` vs. `kind`**:

- `IsLedgerChannel` (`Constructions.lean:40-41`): `c.kind = .ledger ∧ apps c.appDef = consensusRules`.
- `Adjudicator.deposit`/`Adjudicator.transfer` (`Adjudicator.lean:104-152`) gate on `c.kind =
  .ledger` alone, through `Address.fundable`.

This is exactly the disagreement `typed-library.md` §4 (`Channel.lean` row) already names and
rules on (operator ruling 2026-10-31: "the channel doesn't depend on the app"), and
`draft/refined-types` already adds the `LedgerChannel` type for it. **This slice's finding
depends on `draft/refined-types`**: once `LedgerChannel` lands, `IsLedgerChannel` in
`Constructions.lean` and `Route.WellFormed` in `SATP.lean` (which both call `IsLedgerChannel`)
are the two in-slice call sites that need to switch from the raw-`Prop` form to the refined type.
No new ruling is needed from this audit; the conversion is a unit of work, not a design decision.

## 4. Layer violations

**4 confirmed**, all among this slice's 7 `cites = true` theorems, all tracing to the same two
gaps named in §2:

| Theorem | `nitro_base_ops` | Base operation named in the proof | Traces to |
|---|---|---|---|
| `Nitro.network_views_delivered` (`Reachable.lean:907`) | 1 | `cases hop : n.chain.apply apps p₀ op with` (`Reachable.lean:948`) | §2a: would disappear once `Network.step_frame` exists, because the case split on `Adjudicator.apply`'s result would live once, in the frame lemma, not in each of the five consumers |
| `Nitro.network_received_voucher_signers` (`Reachable.lean:663`) | 1 | `cases hop : n.chain.apply apps p₀ op with` (`Reachable.lean:706`) | §2a, same lemma |
| `Nitro.network_views_participant_reachable` (`Reachable.lean:1066`) | 2 | `cases happly : n.chain.apply apps q op with` (`Reachable.lean:1108`, `1117`, `1159`, `1169`) | §2a, same lemma |
| `Threat.RoundSeq.peel_two` (`Strategy.lean:355`) | 1 | [INFERENCE] the proof itself (induction on `Threat.RoundSeq`, list-structure case splits) names no base operation directly; the count most plausibly comes from the theorem's own conclusion mentioning `n1.applyAdv apps (.land opA)`, which is part of the elaborated proof term's type. Not a genuine descent by this proof; `cites=true` because the doc comment cites Nitro:540-542 for the round-level fact, not because the proof reasons about a single on-chain call. | §2b: composing `peel_one` via `peel_split` would not remove this count (the conclusion still names `.land`), so this is a false-positive-for-descent reading of the metric, not a fix candidate |

None of the 4 is the kind of violation doctrine §4a rule 2 targets in Safety.lean/Liveness.lean
(a paper-level theorem skipping the per-channel/redistribution layers to reason about one
on-chain call). The first three are the base-building theorems of this slice itself: their JOB
is to establish facts about the `Nitro.network` step relation, which is defined in terms of
`Adjudicator.apply`, so some descent into it is inherent to this layer's existence. What IS a
genuine finding is that the descent is written out BY HAND three (really five, counting §2a's
other two) separate times instead of once, in a dedicated frame lemma — exactly §2a's proposal.
Fixing §2a removes these three from a future `layerDiscipline` report, as a side effect of
removing the duplication, not as a separate unit of work.

One statement is worth flagging as a candidate **orphan**, not a `nitro_base_ops` violation:
`ValidMove` (`Channel.lean:150-157`, cited "ForceMove:744-751") is the Prop-valued definition of
a valid ForceMove transition. Its computable sibling `validMove` (`Channel.lean:162-167`) is the
one actually used by `forceMoveRules` (`Channel.lean:182-188`). No lemma anywhere connects them
(`validMove _ _ = true ↔ ValidMove _ _ _`), and `git grep` finds zero uses of `ValidMove` outside
its own declaration (`users = 0` in the regenerated `facts.json` too). Either the connecting
lemma is missing (a genuine gap: the computable rule has no proof that it matches its own Prop
specification), or `ValidMove` is dead and should be removed. See §6.

## 5. Size

| Declaration | Lines | Cause |
|---|---|---|
| `Nitro.network_views_participant_reachable` (`Reachable.lean:1066`) | 142 | A case split per `Nitro.network` step constructor (`react`/`wait`/`land`/adversary `chain`/`deliver`/`tick`/`deposit`), each branch further split on `q = p₀` — the same missing frame lemma as §2a would collapse the "which branch, which participant" bookkeeping, leaving only the per-branch reachability argument |
| `turboValues_order_indep` (`Value.lean:453`) | 133 | Inline duplication: the L1/L2 (two candidate topological orders) halves of the induction step are hand-duplicated — `hmemW1`/`hmemW2`, `hsrc1`/`hsrc2`, `hnoout1`/`hnoout2`, `hext1`/`hext2`, `hnd1'`/`hnd2'`, `hvalid1'`/`hvalid2'` each restate the same fact for `L1` then for `L2`. A generic lemma taking "a list and its membership/order/nodup facts" and applied twice (once per list) would remove roughly half the body. |
| `Nitro.network_chain_reachableSigned` (`Reachable.lean:790`) | 116 | Same missing frame lemma as §2a |
| `Nitro.network_conservation` (`Reachable.lean:1197`) | 88 | Same missing frame lemma as §2a (lighter case split; most branches are one-line `simpa`) |
| `exists_topo_list` (`Value.lean:101`) | 67 | A genuine strong induction on list length with one real case split (source-node extraction); not duplication — the size is the algorithm, not a missing API |
| `Nitro.network_received_voucher_signers` (`Reachable.lean:663`) | 103 | Same missing frame lemma as §2a; this theorem IS `network_received_signers` with one extra argument threaded through (§2a names them as one duplication, not two separate causes) |
| `Nitro.network_received_signers` (`Reachable.lean:561`) | 102 | Same missing frame lemma as §2a |
| `Nitro.network_views_delivered` (`Reachable.lean:907`) | 91 | Same missing frame lemma as §2a |

Six of the eight over-60-line declarations in this slice trace to the ONE missing frame lemma in
§2a. Fixing §2a is therefore the highest-leverage single unit of work in this slice.

## 6. Dead code

Checked by `git grep -F` across the whole repository AND by the regenerated `facts.json`'s
`users` field (both agree below; the pre-fix file's `users = 0` for `NetworkState.applyAdv_deliver`
was the tooling defect, not a real finding — see header). Structure-field accessors
(`Threat.ε_pos`, `Assumptions.toBounds`/`toCapable`, `Threats.ok`) now correctly show nonzero
`users` (8, 13, 3, 5) in the regenerated file and are excluded below.

Confirmed dead (zero reference anywhere outside the declaring line, `users = 0` in the
regenerated `facts.json` AND zero `git grep` hits):

- `ValidMove` (`Channel.lean:150-157`, `def`, cited "ForceMove:744-751", 8 lines). See §4: either
  an orphaned spec predicate, or a missing connecting lemma to `validMove`. Not a non-vacuity
  witness, not a spec theorem (it is a `def`, not a `theorem`).
- `Voucher.SignedByBoth` (`SATP.lean:33-34`, `def`, cited "SATP:703", 2 lines). Zero uses.
- `Route.debited` (`SATP.lean:150-151`, `def`, cited "SATP:436-439", 2 lines). Zero uses.
- `Nitro.network_views_consistent` (`Reachable.lean:996-998`, `theorem`, 3 lines). Bundles the
  three `Views.Consistent` conjuncts that `Liveness.lean` instead calls separately
  (`Network.views_chain_eq`, `Nitro.network_views_delivered`, `Nitro.network_views_signers`);
  this bundling theorem itself is never called.

**Not dead, corrected from the pre-fix draft:** `NetworkState.applyAdv_deliver`
(`Strategy.lean:188-191`, `@[simp] theorem`) showed `users = 0` under the buggy `facts.json` and
zero explicit `git grep` hits, but the regenerated file shows `users = 5` — it fires inside other
theorems' `simp` calls without being named textually. Removed from this list.

**Flagged, not listed as dead — required but apparently unread structure fields:**
`Bounds.notHonest` and `Bounds.unforgeableDeliver` (`Strategy.lean:485,487`, fields of the
`Bounds` structure) still show `users = 0` in the regenerated file, and `git grep` finds
`Bounds.notHonest` only in a doc comment (`Reachable.lean:1054`) and `Bounds.unforgeableDeliver`
nowhere outside its own declaration. Unlike `ε_pos`/`toBounds`/`toCapable`/`ok` (confirmed used
via short dot-notation above), these two fields are obligations every `Bounds` instance must
supply, but no proof in this slice reads them back. [INFERENCE] Not necessarily dead — a
consumer in `Liveness.lean`/`Safety.lean` (outside this slice) may read them; worth one grep by
whichever slice owns those files before concluding the `Bounds` structure over-specifies.

Excluded by instructions: `Nitro.network_positive_balance_reachable` (`Reachable.lean:1280`,
zero other reference, 23 lines) is a D10-ruling witness ("a reachable network can hold a positive
balance in a ledger channel") — a non-vacuity witness, not dead code.
`Network.applyHonest_react_chain` (`Reachable.lean:306-308`, `@[simp] theorem`, 2 lines) still
shows `users = 0` after the fix and zero `git grep` hits; unlike `applyAdv_deliver` the fix did
NOT surface an implicit `simp` use, so this one is a genuine dead-code candidate despite the
`@[simp]` tag — listed here rather than above only because its `@[simp]` status still deserves a
second look (a `simp` call elsewhere may discharge the goal before ever trying this lemma,
which would not register as a "use" either) before deletion.

## 7. Statement issues seen

- **`ValidMove`/`validMove` have no connecting lemma** (§4, §6). If `validMove` (the `Bool`
  function `forceMoveRules` actually runs) is meant to decide `ValidMove` (the `Prop` the module
  comment cites to ForceMove:744-751), no proof states that; a reader has to re-verify the two
  definitions agree by eye. [INFERENCE: this reads as a missing API lemma, not a model defect —
  the two definitions are structurally identical field-by-field, so the connecting `Iff` looks
  provable by `simp`/`decide`-unfolding, not a hard gap.]
- **The audit tooling itself had a defect** (see header): `.scratch/audit/facts.lean` built
  `users`/`nitro_base_ops`/`pool_base_ops`/`sorry` from `ConstantInfo.value?` with the default
  `allowOpaque := false`, which is `none` for every `theorem`. `AuditNitroLiveness` found this,
  `Main` regenerated `facts.json` with the fix mid-audit. This report uses the regenerated file;
  the first draft's §4 and §6 (0 layer violations; `NetworkState.applyAdv_deliver` listed dead)
  were corrected once the fix landed.

- **`turboValues_order_indep`'s duplicated halves (§5) are evidence the L1/L2 symmetry was never
  factored out as a lemma**, even though the doc comments for the surrounding section (Value.lean
  piece 1-4h comments) describe each sub-lemma as reusable "pieces". The theorem that should reuse
  them symmetrically does not.

## 8. Plan

Layer order: §2a (base-call frame layer) and §2b (round-composition layer) are both at or just
above the base layer this slice owns; neither has a cross-slice dependency going IN (nothing
below them in doctrine order needs fixing first). Both are a dependency FOR `AuditNitroLiveness`
and `AuditNitroSafety`, whose large theorems (per `type-doctrine.md` §1: `consensus_round_safe_core`
536 lines, `satp_peers_can_finalize` 249 lines, duplicated 81-99 line blocks) are downstream
consumers of `Nitro.network_received_signers`/`_voucher_signers`/`_views_delivered` and of
`Threat.RoundSeq.peel_one`/`peel_two`. Doing §2a/§2b first, before those slices' size work, lets
those workers shrink their own proofs by reusing a smaller frame lemma instead of inlining the
case split again.

1. **Unit: `Network.step_frame` (§2a).** Scope: `Reachable.lean` only. Prove one lemma stating
   how `(chain, sent, received, wallet)` change under each `Nitro.network` step constructor, built
   from the EXISTING per-field simp lemmas (`Network.applyHonest_react_sent`, `Network.land_of_none`,
   `Network.land_of_some`, `Network.applyAdv_deliver`, `Network.applyAdv_tick`,
   `Network.deposit_sent`, `Network.deposit_received`, and their `_chain`/`_wallet` counterparts —
   all already proved, just not combined). Rewrite `Nitro.network_received_signers`,
   `Nitro.network_received_voucher_signers`, `Nitro.network_views_delivered`,
   `Nitro.network_chain_reachableSigned`, `Nitro.network_conservation` to use it.
   Acceptance: the five theorems' statements are byte-for-byte unchanged (same signature, used
   downstream by `Liveness.lean`/`Lemmas.lean`); total line count of the five theorem bodies plus
   the new lemma is materially under the current 500. No `sorry`. No dependency on other slices.
2. **Unit: `Threat.RoundSeq.peel_split` (§2b).** Scope: `Strategy.lean` only. Prove the
   general split lemma and rewrite `peel_one` and `peel_two` to use it (or derive both as
   corollaries). Acceptance: both theorems' statements unchanged (used by `Liveness.lean`,
   `Strategy.lean`'s own `Threat.Round.single`); `peel_two`'s body shrinks from 53 lines to a
   handful, since it stops reimplementing the induction. No dependency on other slices.
3. **Unit: `turboValues_order_indep` L1/L2 factoring (§5).** Scope: `Value.lean` only. Extract
   one lemma parametrized over "a candidate topological order and its membership/validity facts",
   apply it twice. Acceptance: `turboValues_order_indep`'s body shrinks by roughly half; its
   statement (used by `Liveness.lean:6538`) is unchanged. No dependency on other slices.
4. **Unit: `ValidMove`/`validMove` resolution (§4, §6, §7).** Scope: `Channel.lean` only.
   Either (a) prove `validMove vt m1 m2 = true ↔ ValidMove (fun s1 s2 => vt s1 s2 = true) m1 m2`
   and have `forceMoveRules`'s proof-facing consumers (none currently; this lemma has no
   downstream user yet) cite it, or (b) delete `ValidMove` as dead. This needs an operator or
   reviewer call, not a mechanical choice: deleting loses the Prop-level ForceMove citation;
   keeping it unconnected leaves an unverified claim that `validMove` implements `ValidMove`.
   Flagging for a one-line ruling, not blocking the other three units.
5. **Unit: delete confirmed dead code (§6).** Scope: `SATP.lean` (`Voucher.SignedByBoth`,
   `Route.debited`), `Reachable.lean` (`Nitro.network_views_consistent`). Leave the two
   `@[simp]`-tagged candidates (`NetworkState.applyAdv_deliver`, `Network.applyHonest_react_chain`)
   out of this unit — removing a `@[simp]` lemma can silently break a downstream `simp` call;
   that check belongs with whichever slice's build gate catches it, not a blind deletion here.
6. **Unit: convert `IsLedgerChannel`/`Route.WellFormed` to `LedgerChannel` (§3).** Scope:
   `Constructions.lean`, `SATP.lean`. **Depends on `draft/refined-types` merging first** (adds
   the `LedgerChannel` type). Once merged: `IsLedgerChannel apps c` becomes
   `c : LedgerChannel ∧ apps c.1.appDef = consensusRules` or equivalent, and the two in-slice call
   sites (`LedgerWithSub.Holds`, `Route.WellFormed`) switch to it. No new operator ruling needed;
   the ruling already exists (`typed-library.md` §4, 2026-10-31).

Units 1-3 and 5 are independent of each other and of the other audit slices; a worker can start
any of them immediately. Unit 4 needs one operator ruling before it is dispatchable. Unit 6 is
blocked on `draft/refined-types` landing.
