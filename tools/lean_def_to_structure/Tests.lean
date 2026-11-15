/-
  Tests of the pattern matcher (`DefToStructure.decide`, `silentProjections`).

    lake env <armada-lean checkout> .lake/build/bin/tests

  Each case gives a source snippet, the diagnostic range (a needle in the
  snippet), the elaboration facts an `Oracle` would answer (type heads,
  expected heads, goal heads, definitions, diagnostic embeds, each at a
  needle), and the expected repair: the pattern and the text after the edit,
  or no repair. The facts record what the server answered for the recorded
  Lean 4.35.0 cases (2026-11-14; `.scratch/exp.lean`, `.scratch/exp2.lean`
  under the case worktrees). The snippets parse with `Init` and
  `Mathlib.Tactic.Says` imported from the checkout given on the command line.
-/
import Lean
import DefToStructure.Patterns
import DefToStructure.Discovery

open Lean DefToStructure

namespace Tests

def predFile : String :=
"structure Reg (reg : Nat → Option Nat) : Prop where
  holds : ∀ c pre, reg c = some pre → pre + 1 = c
structure Dag (r : Nat → Nat → Prop) : Prop where
  holds : Acyclic r
structure Topo (l : List Nat) : Prop where
  holds : l.Nodup ∧ (∀ n ∈ l, n < 10) ∧ l ≠ []
structure Adjudicator.IsDAG (s : Adjudicator E) : Prop where
  holds : Acyclic s.Edge
structure RegistryConsistent (reg : PreimageRegistry) : Prop where
  holds : ∀ c pre, reg c = some pre → hashCommitment pre = c
structure OnlyTransfers (ops : List (Participant × Op E)) : Prop where
  holds : ∀ e ∈ ops, ∃ a b, e.2 = Op.transfer a b
structure Threat.TransfersOnly {p : Participant} (T : Threat E p) : Prop where
  holds : ∀ net m, T.feasible net m → ∃ caller a b, m = .chain caller (.transfer a b)
"

def predUri := "file:///P.lean"

/-- Byte offset of the `occ`-th occurrence of `needle` in `s`. -/
def find! (s needle : String) (occ : Nat := 0) : Nat := Id.run do
  let parts := s.splitOn needle
  if parts.length ≤ occ + 1 then panic! s!"needle `{needle}` #{occ} not found in {s}"
  let mut off := 0
  for i in [0:occ + 1] do
    off := off + parts[i]!.utf8ByteSize + (if i == 0 then 0 else needle.utf8ByteSize)
  return off

def mkPred (name declId : String) (bodyHead : Option Loc := none) (bodyIsStructure := false) : Pred :=
  let decl := s!"structure {declId}"
  let off := find! predFile decl
  let holds := find! predFile "holds" (((String.Pos.Raw.extract predFile 0 ⟨off⟩).splitOn "holds").length - 1)
  { fullName := name, declId, file := "P.lean", uri := predUri, declIdOffset := off + "structure ".utf8ByteSize,
    holdsOffset := holds, bodyHead, bodyIsStructure }

def acyclicLoc : Loc := { uri := "file:///A.lean", line := 0, character := 4 }
def andLoc : Loc := { uri := "file:///Init.lean", line := 7, character := 10 }
def otherLoc : Loc := { uri := "file:///O.lean", line := 3, character := 4 }
def fooLoc : Loc := { uri := "file:///O.lean", line := 9, character := 4 }

def REG := mkPred "Exp.Reg" "Reg"
def DAG := mkPred "Exp.Dag" "Dag" (some acyclicLoc) true
def TOPO := mkPred "Exp.Topo" "Topo" (some andLoc) true
def ISDAG := mkPred "ArmadaLean.Nitro.Adjudicator.IsDAG" "Adjudicator.IsDAG" (some acyclicLoc) true
def RC := mkPred "ArmadaLean.RegistryConsistent" "RegistryConsistent"
def OT := mkPred "ArmadaLean.Nitro.OnlyTransfers" "OnlyTransfers"
def TO := mkPred "ArmadaLean.Nitro.Threat.TransfersOnly" "Threat.TransfersOnly"

def predTexts : PredTexts := { fileMaps := Std.HashMap.ofList [("P.lean", FileMap.ofString predFile)] }
def loc (p : Pred) : Loc := p.loc predTexts
def holdsLoc (p : Pred) : Loc := p.holdsLoc predTexts

inductive Fact where
  /-- the head of the diagnosed term's type (the diagnostic's second embed). -/
  | type (ans : Array Loc)
  /-- `definition` at the needle's start. -/
  | defn (needle : String) (occ : Nat) (ans : Array Loc)
  /-- expected-type head at the needle's start. -/
  | exp (needle : String) (occ : Nat) (ans : Array Loc)
  /-- goal head at the needle's start. -/
  | goal (needle : String) (occ : Nat) (ans : Array Loc)
  /-- the heads of the diagnostic's embedded expressions. -/
  | embeds (ans : Array (Array Loc))

structure Case where
  name : String
  src : String
  /-- The diagnostic range: a needle and its occurrence. -/
  diag : String × Nat
  preds : Array Pred := #[REG, DAG, TOPO, ISDAG]
  facts : List Fact := []
  /-- `some (pattern, text after the edit)`, or `none` for a refusal. -/
  expect : Option (String × String)

def mkOracle (src : String) (facts : List Fact) (log : IO.Ref (Array String))
    (diag : Option (Nat × Nat) := none) : Oracle :=
  let look? (kind : String) (off : Nat) : IO (Option (Array Loc)) := do
    for f in facts do
      match kind, f with
      | "def", .defn n o a | "exp", .exp n o a | "goal", .goal n o a =>
        if find! src n o == off then return some a
      | _, _ => pure ()
    log.modify (·.push s!"{kind}@{off}")
    return none
  let look (kind : String) (off : Nat) : IO (Array Loc) := do return (← look? kind off).getD #[]
  { definition := look "def", expectedHead := look? "exp", goalHead := look "goal",
    diagExprHeads := fun s e => do
      -- the embeds belong to the diagnostic at its own range only
      if diag.any (· != (s, e)) then
        log.modify (·.push s!"embeds@{s}-{e}")
        return #[]
      for f in facts do
        match f with
        | .embeds a => return a
        | .type a => return #[#[], a]
        | _ => pure ()
      return #[] }

def runCase (env : Environment) (c : Case) : IO Bool := do
  let parsed ← Parse.parseFile env "t.lean" c.src
  unless parsed.parseErrors.isEmpty do
    IO.println s!"FAIL {c.name}: parse error {parsed.parseErrors}"
    return false
  let log ← IO.mkRef #[]
  let trace ← IO.mkRef #[]
  -- a trailing newline in the needle pins an occurrence; the range excludes it
  let s := find! c.src c.diag.1 c.diag.2
  let needle := if c.diag.1.endsWith "\n" then (c.diag.1.dropEnd 1).copy else c.diag.1
  let e := s + needle.utf8ByteSize
  let ctx : Ctx := { env, parsed, preds := c.preds, predTexts, oracle := mkOracle c.src c.facts log (some (s, e)), trace }
  let r ← decide ctx s e
  let got := r.map fun r => (r.pattern, applyEdits c.src r.edits)
  if got == c.expect then
    IO.println s!"ok   {c.name}"
    return true
  IO.println s!"FAIL {c.name}\n  expected {repr c.expect}\n  got      {repr got}\n  unanswered {← log.get}"
  return false

def functionExpectedEmbeds (p : Pred) : Fact := .embeds #[#[], #[loc p]]

def cases : List Case := [
  -- holds-apply
  { name := "holds-apply term mode", src := "theorem t1 (h : Reg reg) : pre + 1 = c :=\n  h c pre hc\n",
    diag := ("h c pre hc", 0), facts := [.type #[loc REG]],
    expect := some ("holds-apply", "theorem t1 (h : Reg reg) : pre + 1 = c :=\n  h.holds c pre hc\n") },
  { name := "holds-apply exact", src := "theorem t1b : pre + 1 = c := by\n  exact h c pre hc\n",
    diag := ("h c pre hc", 0), facts := [.type #[loc REG]],
    expect := some ("holds-apply", "theorem t1b : pre + 1 = c := by\n  exact h.holds c pre hc\n") },
  { name := "holds-apply projection chain head (type from the diagnostic embed)",
    src := "theorem v5 (h : True ∧ Reg reg) : pre + 1 = c :=\n  h.2 c pre hc\n",
    diag := ("h.2 c pre hc", 0), facts := [.defn "2 c pre" 0 #[andLoc], functionExpectedEmbeds REG],
    expect := some ("holds-apply", "theorem v5 (h : True ∧ Reg reg) : pre + 1 = c :=\n  h.2.holds c pre hc\n") },
  { name := "holds-apply parenthesized", src := "theorem v4 (h : Reg reg) : pre + 1 = c :=\n  (h) c pre hc\n",
    diag := ("(h) c pre hc", 0), facts := [.type #[loc REG]],
    expect := some ("holds-apply", "theorem v4 (h : Reg reg) : pre + 1 = c :=\n  (h).holds c pre hc\n") },
  { name := "holds-apply refuses another type", src := "theorem t (h : Foo reg) : pre + 1 = c :=\n  h c pre hc\n",
    diag := ("h c pre hc", 0), facts := [.type #[fooLoc]], expect := none },
  { name := "holds-apply refuses when the head has no type", src := "theorem t (h : Reg reg) : pre + 1 = c :=\n  g c pre hc\n",
    diag := ("g c pre hc", 0), facts := [], expect := none },
  -- holds-field
  { name := "holds-field field", src := "theorem t1g : ¬ Relation.TransGen r n n :=\n  h.irrefl n\n",
    diag := ("irrefl n", 0), facts := [.type #[loc DAG]],
    expect := some ("holds-field", "theorem t1g : ¬ Relation.TransGen r n n :=\n  h.holds.irrefl n\n") },
  { name := "holds-field after a projection (type from the embed)", src := "theorem u2b : ¬ Relation.TransGen r n n :=\n  h.1.irrefl n\n",
    diag := ("irrefl n", 0), facts := [.defn "1.irrefl" 0 #[andLoc], .embeds #[#[], #[loc DAG]]],
    expect := some ("holds-field", "theorem u2b : ¬ Relation.TransGen r n n :=\n  h.1.holds.irrefl n\n") },
  { name := "holds-field parenthesized object", src := "theorem u2 : ¬ Relation.TransGen r n n :=\n  (h.1).irrefl n\n",
    diag := ("irrefl n", 0), facts := [.type #[loc DAG]],
    expect := some ("holds-field", "theorem u2 : ¬ Relation.TransGen r n n :=\n  (h.1).holds.irrefl n\n") },
  { name := "holds-field refuses a foreign structure", src := "theorem t : ¬ Relation.TransGen r n n :=\n  h.irrefl n\n",
    diag := ("irrefl n", 0), facts := [.type #[otherLoc]], expect := none },
  { name := "holds-field numeric projection", src := "theorem t1e (h : Topo l) : l ≠ [] := h.2.2\n",
    diag := ("h.2.2", 0), facts := [.type #[loc TOPO]],
    expect := some ("holds-field", "theorem t1e (h : Topo l) : l ≠ [] := h.holds.2.2\n") },
  { name := "holds-field numeric projection applied", src := "theorem v2 (h : Topo l) : n < 10 :=\n  h.2.1 n hn\n",
    diag := ("h.2.1 n hn", 0), facts := [.type #[loc TOPO]],
    expect := some ("holds-field", "theorem v2 (h : Topo l) : n < 10 :=\n  h.holds.2.1 n hn\n") },
  { name := "holds-field numeric projection after a projection", src := "theorem v3 (h : Topo l ∧ True) : n < 10 :=\n  h.1.2.1 n hn\n",
    diag := ("h.1.2.1 n hn", 0), facts := [.defn "1.2.1" 0 #[andLoc], .embeds #[#[], #[loc TOPO]]],
    expect := some ("holds-field", "theorem v3 (h : Topo l ∧ True) : n < 10 :=\n  h.1.holds.2.1 n hn\n") },
  { name := "holds-field dotted intermediate (hord.isDAG.irrefl)",
    src := "theorem x : ∀ e ∈ s.allocationOf c, e.dest ≠ c := fun e he hcontra =>\n    hord.isDAG.irrefl c (Relation.TransGen.single ⟨hcext, e, he, hcontra⟩)\n",
    diag := ("irrefl c", 0), facts := [.defn "isDAG.irrefl" 0 #[otherLoc], .embeds #[#[], #[loc ISDAG]]],
    expect := some ("holds-field", "theorem x : ∀ e ∈ s.allocationOf c, e.dest ≠ c := fun e he hcontra =>\n    hord.isDAG.holds.irrefl c (Relation.TransGen.single ⟨hcext, e, he, hcontra⟩)\n") },
  -- holds-type and the visible projection case
  { name := "holds-proj visible as a type mismatch", src := "theorem t1d (l : List Nat) (h : Topo l) : l.Nodup := h.1\n",
    diag := ("h.1", 0), facts := [.embeds #[#[], #[andLoc], #[otherLoc]], .defn "1\n" 0 #[holdsLoc TOPO]],
    expect := none },
  { name := "holds-type exact", src := "theorem u1 (h : Both r reg) : Acyclic r := by\n  exact h.1\n",
    diag := ("exact h.1", 0), facts := [.embeds #[#[], #[loc DAG], #[acyclicLoc]]],
    expect := some ("holds-type", "theorem u1 (h : Both r reg) : Acyclic r := by\n  exact h.1.holds\n") },
  { name := "holds-type argument", src := "theorem u5 (h : Reg reg) : True := needsBody reg h\n",
    diag := ("h\n", 0), facts := [.exp "h\n" 0 #[otherLoc], .type #[loc REG]],
    expect := some ("holds-type", "theorem u5 (h : Reg reg) : True := needsBody reg h.holds\n") },
  { name := "holds-type application gets parentheses", src := "theorem u (h : Reg reg) : body := by\n  exact mk_reg h x\n",
    diag := ("exact mk_reg h x", 0), facts := [.exp "mk_reg h x" 0 #[otherLoc], .type #[loc REG]],
    expect := some ("holds-type", "theorem u (h : Reg reg) : body := by\n  exact (mk_reg h x).holds\n") },
  { name := "holds-type field notation type", src := "theorem u (h : s.IsDAG) : Acyclic s.Edge := by\n  exact h\n",
    diag := ("exact h", 0), facts := [.exp "h\n" 0 #[acyclicLoc], .type #[loc ISDAG]],
    expect := some ("holds-type", "theorem u (h : s.IsDAG) : Acyclic s.Edge := by\n  exact h.holds\n") },
  -- wrap
  { name := "wrap fun term", src := "theorem t2 : Reg (fun _ => none) := fun _ _ h => by injection h\n",
    diag := ("fun _ _ h => by injection h", 0), facts := [.exp "fun _ _ h" 0 #[loc REG]],
    expect := some ("wrap", "theorem t2 : Reg (fun _ => none) := ⟨fun _ _ h => by injection h⟩\n") },
  { name := "wrap hypothesis", src := "theorem t2b (hb : ∀ c pre, reg c = some pre → pre + 1 = c) : Reg reg := hb\n",
    diag := ("hb\n", 0), facts := [.exp "hb\n" 0 #[loc REG], .type #[otherLoc]],
    expect := some ("wrap", "theorem t2b (hb : ∀ c pre, reg c = some pre → pre + 1 = c) : Reg reg := ⟨hb⟩\n") },
  { name := "wrap exact", src := "theorem t2c : Reg reg := by\n  exact hb\n",
    diag := ("exact hb", 0), facts := [.exp "hb\n" 0 #[loc REG]],
    expect := some ("wrap", "theorem t2c : Reg reg := by\n  exact ⟨hb⟩\n") },
  { name := "wrap argument in an anonymous constructor", src := "theorem t2d : ∃ reg, Reg reg := ⟨fun _ => none, fun _ _ h => by injection h⟩\n",
    diag := ("fun _ _ h => by injection h", 0), facts := [.exp "fun _ _ h" 0 #[loc REG]],
    expect := some ("wrap", "theorem t2d : ∃ reg, Reg reg := ⟨fun _ => none, ⟨fun _ _ h => by injection h⟩⟩\n") },
  { name := "wrap multi-line fun", src := "theorem u7 : ∃ reg, Reg reg :=\n  ⟨fun _ => none,\n    fun _ _ h => by\n      injection h⟩\n",
    diag := ("fun _ _ h => by\n      injection h", 0), facts := [.exp "fun _ _ h" 0 #[loc REG]],
    expect := some ("wrap", "theorem u7 : ∃ reg, Reg reg :=\n  ⟨fun _ => none,\n    ⟨fun _ _ h => by\n      injection h⟩⟩\n") },
  { name := "wrap-nested", src := "theorem t2e (hr : ∀ n, ¬ TransGen r n n) : Dag r := ⟨fun n hn => hr n hn⟩\n",
    diag := ("fun n hn => hr n hn", 0), facts := [.exp "fun n hn" 0 #[acyclicLoc], .exp "⟨fun" 0 #[loc DAG]],
    expect := some ("wrap-nested", "theorem t2e (hr : ∀ n, ¬ TransGen r n n) : Dag r := ⟨⟨fun n hn => hr n hn⟩⟩\n") },
  { name := "wrap-nested multi-line instance", src := "theorem x : s'.IsDAG :=\n  ⟨fun n hn => hdag.holds.irrefl n\n    (Relation.TransGen.mono (fun x y hxy => hxy) n n hn)⟩\n",
    diag := ("fun n hn => hdag.holds.irrefl n\n    (Relation.TransGen.mono (fun x y hxy => hxy) n n hn)", 0),
    facts := [.exp "fun n hn" 0 #[acyclicLoc], .exp "⟨fun" 0 #[loc ISDAG]],
    expect := some ("wrap-nested", "theorem x : s'.IsDAG :=\n  ⟨⟨fun n hn => hdag.holds.irrefl n\n    (Relation.TransGen.mono (fun x y hxy => hxy) n n hn)⟩⟩\n") },
  { name := "wrap-nested refuses when the body is not a structure", src := "theorem t : Reg reg := ⟨fun c pre h => foo⟩\n",
    diag := ("fun c pre h => foo", 0), preds := #[REG], facts := [.exp "fun c pre" 0 #[acyclicLoc], .exp "⟨fun" 0 #[loc REG]],
    expect := none },
  { name := "wrap refuses an already wrapped term", src := "theorem t : Reg reg := ⟨hb⟩\n",
    diag := ("⟨hb⟩", 0), facts := [.exp "⟨hb⟩" 0 #[loc REG]], expect := none },
  -- constructor
  { name := "constructor intro on its own line", src := "theorem t3 : Reg (fun _ => none) := by\n  intro c pre h\n  injection h\n",
    diag := ("c pre h", 0), facts := [.goal "intro c" 0 #[loc REG]],
    expect := some ("constructor", "theorem t3 : Reg (fun _ => none) := by\n  constructor\n  intro c pre h\n  injection h\n") },
  { name := "constructor rintro after a focus dot", src := "theorem t : Reg (fun _ => none) := by\n  · rintro c pre ⟨⟩\n",
    diag := ("c pre ⟨⟩", 0), facts := [.goal "rintro c" 0 #[loc REG]],
    expect := some ("constructor", "theorem t : Reg (fun _ => none) := by\n  · constructor; rintro c pre ⟨⟩\n") },
  { name := "constructor intro in a match alternative",
    src := "theorem t : X := by\n  induction h with\n  | genesis _ _ _ => intro c pre hc; simp at hc\n  | step _ hstep ih => exact foo\n",
    diag := ("c pre hc", 0), preds := #[RC], facts := [.goal "intro c" 0 #[loc RC]],
    expect := some ("constructor", "theorem t : X := by\n  induction h with\n  | genesis _ _ _ => constructor; intro c pre hc; simp at hc\n  | step _ hstep ih => exact foo\n") },
  { name := "constructor refuses when the diagnosed term elaborated", src := "theorem t : X := by\n  refine ⟨a, b, c⟩\n",
    diag := ("⟨a, b, c⟩", 0), facts := [.goal "refine" 0 #[loc ISDAG], .exp "⟨a, b, c⟩" 0 #[loc ISDAG]],
    expect := none },
  { name := "constructor parenthesized receiver goal", src := "theorem turboT_only (p : Participant) : (turboT p).TransfersOnly := by\n  rintro _ _ ⟨caller, a, b, -, hm⟩; exact ⟨caller, a, b, hm⟩\n",
    diag := ("_ _ ⟨caller", 0), preds := #[TO], facts := [.goal "rintro _" 0 #[loc TO]],
    expect := some ("constructor", "theorem turboT_only (p : Participant) : (turboT p).TransfersOnly := by\n  constructor\n  rintro _ _ ⟨caller, a, b, -, hm⟩; exact ⟨caller, a, b, hm⟩\n") },
  { name := "constructor refuses another goal", src := "theorem t : Foo := by\n  intro c pre h\n",
    diag := ("c pre h", 0), facts := [.goal "intro c" 0 #[fooLoc]], expect := none },
  { name := "constructor refuses an unknown tactic", src := "theorem t : Reg r := by\n  injection h\n",
    diag := ("injection h", 0), facts := [.goal "injection" 0 #[loc REG]], expect := none },
  -- unfold and simp
  { name := "unfold two names", src := "theorem t3b : Dag r := by\n  unfold Dag Acyclic\n  exact ⟨hr⟩\n",
    diag := ("Dag Acyclic", 0), facts := [.defn "Dag Acyclic" 0 #[loc DAG]],
    expect := none },
  { name := "unfold two names (range on the name)", src := "theorem t3b : Dag r := by\n  unfold Dag Acyclic\n  exact ⟨hr⟩\n",
    diag := ("Dag", 1), facts := [.defn "Dag Acyclic" 0 #[loc DAG]],
    expect := some ("unfold-iff", "theorem t3b : Dag r := by\n  rw [Dag.iff]\n  unfold Acyclic\n  exact ⟨hr⟩\n") },
  { name := "unfold one name", src := "theorem t3c : Dag r := by\n  unfold Dag\n  exact ⟨hr⟩\n",
    diag := ("Dag", 1), facts := [.defn "Dag\n  exact" 0 #[loc DAG]],
    expect := some ("unfold-iff", "theorem t3c : Dag r := by\n  rw [Dag.iff]\n  exact ⟨hr⟩\n") },
  { name := "unfold qualified name with at", src := "theorem t : X := by\n  unfold Adjudicator.IsDAG at h\n",
    diag := ("Adjudicator.IsDAG", 0), facts := [.defn "Adjudicator.IsDAG at" 0 #[loc ISDAG]],
    expect := some ("unfold-iff", "theorem t : X := by\n  rw [Adjudicator.IsDAG.iff] at h\n") },
  { name := "unfold refuses another name", src := "theorem t : X := by\n  unfold Other\n",
    diag := ("Other", 0), facts := [.defn "Other" 0 #[otherLoc]], expect := none },
  { name := "simp only", src := "theorem t3e : Reg (fun _ => none) := by\n  simp only [Reg]\n",
    diag := ("simp only [Reg]", 0), facts := [.defn "Reg]" 0 #[loc REG]],
    expect := some ("simp-iff", "theorem t3e : Reg (fun _ => none) := by\n  simp only [Reg.iff]\n") },
  { name := "simp says", src := "theorem t : X := by\n  exact ⟨[], by (simp? [OnlyTransfers] says simp only [OnlyTransfers, List.not_mem_nil, IsEmpty.forall_iff, implies_true]), rfl⟩\n",
    diag := ("simp? [OnlyTransfers] says simp only [OnlyTransfers, List.not_mem_nil, IsEmpty.forall_iff, implies_true]", 0),
    preds := #[OT], facts := [.defn "OnlyTransfers]" 0 #[loc OT], .defn "OnlyTransfers," 0 #[loc OT]],
    expect := some ("simp-iff", "theorem t : X := by\n  exact ⟨[], by (simp? [OnlyTransfers.iff] says simp only [OnlyTransfers.iff, List.not_mem_nil, IsEmpty.forall_iff, implies_true]), rfl⟩\n") },
  { name := "simp refuses without the predicate", src := "theorem t : X := by\n  simp only [foo]\n",
    diag := ("simp only [foo]", 0), facts := [.defn "foo]" 0 #[fooLoc]], expect := none },
  -- refusals
  { name := "unsolved goals", src := "theorem t : X := by\n  omega\n", diag := ("omega", 0), facts := [.goal "omega" 0 #[otherLoc]], expect := none } ]

/-- The silent projection scan on a snippet where `hord.1` resolves to `P.holds`. -/
def silentCase (env : Environment) : IO Bool := do
  let src := "theorem x : X := by\n  have hnd := hord.1\n  exact hnd.2.2 c hcp\n"
  let parsed ← Parse.parseFile env "t.lean" src
  let log ← IO.mkRef #[]
  let trace ← IO.mkRef #[]
  let facts := [Fact.defn "1\n  exact" 0 #[holdsLoc ISDAG], Fact.defn "2.2 c" 0 #[andLoc], Fact.defn "2 c hcp" 0 #[andLoc]]
  let ctx : Ctx := { env, parsed, preds := #[ISDAG], predTexts, oracle := mkOracle src facts log, trace }
  let rs ← silentProjections ctx
  let got := rs.map fun r => (r.pattern, applyEdits src r.edits)
  let want := #[("holds-proj", "theorem x : X := by\n  have hnd := hord.holds.1\n  exact hnd.2.2 c hcp\n")]
  if got == want then IO.println "ok   silent projection"; return true
  IO.println s!"FAIL silent projection: {repr got}"
  return false

end Tests

open Tests in
unsafe def main (args : List String) : IO UInt32 := do
  let [root] := args | do IO.eprintln "usage: tests <armada-lean checkout>"; return 2
  let env ← Discovery.importTarget root #[`Init, `Mathlib.Tactic.Says]
  let mut failed := 0
  let mut total := 0
  for c in cases do
    total := total + 1
    unless ← runCase env c do failed := failed + 1
  total := total + 1
  unless ← silentCase env do failed := failed + 1
  IO.println s!"{total - failed}/{total} passed"
  return if failed == 0 then 0 else 1
