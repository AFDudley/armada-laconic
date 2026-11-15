/-
  The repair patterns (see ../README.md, "Repair patterns").

  `decide` takes one error diagnostic (its byte range), the parsed file, the
  converted predicates and an `Oracle`, and returns one `Repair` or `none`.
  The decision reads:
  - the syntax nodes that cover the range (`Parse.stackAt`),
  - the declaration location of the head of the expected type of a term
    (`Oracle.expectedHead`, `Lean.Widget.getInteractiveTermGoal` and
    `Lean.Widget.getGoToLocation` on the head tag of its type),
  - the declaration location of the head of the main goal before a tactic
    (`Oracle.goalHead`, `Lean.Widget.getInteractiveGoals` the same way),
  - the declaration an identifier or a field index resolves to
    (`Oracle.definition`, `textDocument/definition`),
  - the head location of the type of the term the diagnostic is about
    (`Oracle.diagExprHeads`, `Lean.Widget.getInteractiveDiagnostics` and
    `Lean.Widget.getGoToLocation` on the second expression Lean embedded in
    the diagnostic: `e`, then `eType`, in `throwTypeMismatchError`,
    `throwFunctionExpected` and the field errors). This is the one fact that
    no position query gives: `hoverableInfoAtM?` prefers the innermost info at
    a position, every position of a projection or application node belongs to
    one of its children, and an intermediate of a dotted identifier
    (`hord.isDAG` in `hord.isDAG.irrefl`) has no info at all
    (`elabAppLValsAux` elaborates it with `expectedType? := none`). The tool
    reads the embedded `Expr` through its RPC reference, not the message text.

  A predicate is recognized by location: the position of its name in the
  `structure` declaration the tool wrote, or the position of its `holds`
  field. No pattern reads the message text.
-/
import Lean
import DefToStructure.Parse

open Lean Parser

namespace DefToStructure

/-- A declaration location, as `textDocument/definition` reports it:
    the document and the start of the name. -/
structure Loc where
  uri : String
  line : Nat
  character : Nat
  deriving BEq, Repr, Inhabited, ToJson, FromJson

/-- One converted predicate. Byte offsets refer to the current text of
    `file`; `Pred.loc` turns them into locations with that text's `FileMap`. -/
structure Pred where
  fullName : String
  /-- The name as the `def` wrote it: `Adjudicator.IsDAG`. -/
  declId : String
  /-- Path relative to the checkout. -/
  file : String
  uri : String
  /-- Byte offset of the name token in `structure <name>`. -/
  declIdOffset : Nat
  /-- Byte offset of the `holds` field token. -/
  holdsOffset : Nat
  /-- Declaration location of the head constant of the body, when it has one. -/
  bodyHead : Option Loc
  /-- The body's head unfolds to a structure (`Acyclic` → `Std.Irrefl`). -/
  bodyIsStructure : Bool
  deriving Repr, Inhabited, ToJson, FromJson

def Pred.short (p : Pred) : String := (p.declId.splitOn ".").getLast!

def locAt (uri : String) (fm : FileMap) (off : Nat) : Loc :=
  let lp := fm.utf8PosToLspPos ⟨off⟩
  { uri, line := lp.line, character := lp.character }

structure Edit where
  start : Nat
  stop : Nat
  replacement : String
  deriving Repr, BEq, Inhabited, ToJson, FromJson

structure Repair where
  pattern : String
  pred : Option String
  detail : String
  edits : Array Edit
  deriving Repr, Inhabited, ToJson, FromJson

def mkRepair (pattern : String) (pred : String) (detail : String) (edits : Array Edit) : Repair :=
  { pattern, pred := some pred, detail, edits }

/-- Elaboration data for one file, by byte position. -/
structure Oracle where
  definition : Nat → IO (Array Loc)
  /-- `none` when the server has no term info at the position: the term did
      not elaborate. -/
  expectedHead : Nat → IO (Option (Array Loc))
  goalHead : Nat → IO (Array Loc)
  diagExprHeads : Nat → Nat → IO (Array (Array Loc))

/-- The predicate file texts the decision needs for locations: the current
    text of each predicate's file. -/
structure PredTexts where
  fileMaps : Std.HashMap String FileMap

def Pred.loc (p : Pred) (pt : PredTexts) : Loc :=
  locAt p.uri (pt.fileMaps.getD p.file (FileMap.ofString "")) p.declIdOffset

def Pred.holdsLoc (p : Pred) (pt : PredTexts) : Loc :=
  locAt p.uri (pt.fileMaps.getD p.file (FileMap.ofString "")) p.holdsOffset

structure Ctx where
  env : Environment
  parsed : Parse.Parsed
  preds : Array Pred
  predTexts : PredTexts
  oracle : Oracle
  /-- Decision trace lines (the driver writes them to the log). -/
  trace : IO.Ref (Array String)

def Ctx.note (c : Ctx) (s : String) : IO Unit := c.trace.modify (·.push s)

/-- The predicate whose name location is the first location of `locs`. -/
def Ctx.predAt (c : Ctx) (locs : Array Loc) : Option Pred := do
  let l ← locs[0]?
  c.preds.find? fun p => p.loc c.predTexts == l

/-- The predicate whose `holds` field is where `locs` points. -/
def Ctx.holdsAt (c : Ctx) (locs : Array Loc) : Option Pred := do
  let l ← locs[0]?
  c.preds.find? fun p => p.holdsLoc c.predTexts == l

def rangeOf (stx : Syntax) : Nat × Nat :=
  match Parse.range? stx with
  | some (s, e) => (s.byteIdx, e.byteIdx)
  | none => (0, 0)

def textOf (c : Ctx) (s e : Nat) : String :=
  String.Pos.Raw.extract c.parsed.text ⟨s⟩ ⟨e⟩

def isKind (k : Name) (stx : Syntax) : Bool := stx.isOfKind k

/-- The byte position for the expected-type query of a term: its first token.
    `termGoalAt?` picks the smallest info with an expected type there; for an
    application it skips the function identifier. -/
def freePos (stx : Syntax) : Nat := (rangeOf stx).1

/-- Words that start the tactics `constructor` may precede. -/
def goalTactics : List String := [
  "intro", "intros", "rintro", "rw", "rewrite", "rwa", "exact", "refine", "refine'",
  "apply", "simp", "simp_all", "simpa", "obtain", "rcases", "cases", "induction",
  "omega", "decide", "trivial", "exfalso", "by_contra", "push_neg", "specialize",
  "show", "change", "ext", "funext", "congr", "use"]

def simpWords : List String := ["simp?", "simp_all?", "simpa?", "simp", "simp_all", "simpa", "dsimp"]

def firstWord (stx : Syntax) : String :=
  match stx with
  | .atom _ v => v
  | .node _ _ args => if h : 0 < args.size then firstWord args[0] else ""
  | _ => ""

def startsLine (c : Ctx) (off : Nat) : Bool :=
  let ls := (Parse.lineStart c.parsed.fileMap ⟨off⟩).byteIdx
  (textOf c ls off).trimAscii.isEmpty

def indentOf (c : Ctx) (off : Nat) : String :=
  let ls := (Parse.lineStart c.parsed.fileMap ⟨off⟩).byteIdx
  let seg := textOf c ls off
  (seg.takeWhile (· == ' ')).copy

/-- `constructor` before the tactic that starts at `off`. -/
def constructorEdit (c : Ctx) (off : Nat) : Edit :=
  if startsLine c off then ⟨off, off, "constructor\n" ++ indentOf c off⟩
  else ⟨off, off, "constructor; "⟩

/-- Whether the term node has an argument list or binders: it needs
    parentheses before `.holds`. -/
def needsParens (stx : Syntax) : Bool :=
  !(stx.isIdent || isKind ``Term.proj stx || isKind ``Term.paren stx || isKind ``Term.anonymousCtor stx)

/-- The term a diagnostic at `[s, e]` is about: the innermost term node whose
    range equals the diagnostic range, or the term of an `exact`/`refine`/`apply`
    tactic whose range equals it. -/
def termAt (c : Ctx) (stack : Array Syntax) (s e : Nat) : Option Syntax := Id.run do
  let exact := stack.reverse.find? fun n => rangeOf n == (s, e)
  let some n := exact | return none
  if Parse.isTermKind c.env n.getKind then return some n
  if Parse.isTacticKind c.env n.getKind && ["exact", "refine", "refine'", "apply"].contains (firstWord n) then
    -- the term is the last term-kind argument
    return n.getArgs.reverse.find? fun a => Parse.isTermKind c.env a.getKind
  return none

/-- One step of a projection chain: the field token's range, whether it is a
    numeric index, and where the object before it ends. -/
structure Step where
  fieldStart : Nat
  fieldEnd : Nat
  numeric : Bool
  objectEnd : Nat
  deriving Inhabited, Repr

/-- The steps of a chain `root.f₁.f₂…`: `h.isDAG.irrefl` (one dotted
    identifier), `h.1.irrefl` (`Term.proj` nodes), `(h.1).irrefl` (a
    parenthesized root). -/
partial def decompose (c : Ctx) (n : Syntax) : Option (Array Step) :=
  if n.isIdent then
    let comps := Parse.identComponentRanges c.parsed.text n
    if comps.size < 2 then none
    else
      some <| (comps.toList.drop 1).toArray.mapIdx fun i (a, b) =>
        { fieldStart := a.byteIdx, fieldEnd := b.byteIdx, numeric := false,
          objectEnd := comps[i]!.2.byteIdx }
  else if isKind ``Term.proj n then
    let obj := n[0]
    let fld := n[2]
    let step : Step := { fieldStart := (rangeOf fld).1, fieldEnd := (rangeOf fld).2,
                         numeric := fld.isOfKind fieldIdxKind, objectEnd := (rangeOf obj).2 }
    match decompose c obj with
    | some steps => some (steps.push step)
    | none => some #[step]
  else none

/-- The head of the type of the term a diagnostic is about: the second
    expression Lean embedded in the diagnostic (`e`, then `eType`, in
    `throwTypeMismatchError`, `throwFunctionExpected`, the invalid-field and
    invalid-projection errors). Position queries cannot give it: the server's
    info lookup prefers the innermost info at a position, and every position
    of a projection or application node belongs to one of its children. -/
def actualType (c : Ctx) (s e : Nat) : IO (Array Loc) := do
  return ((← c.oracle.diagExprHeads s e)[1]?).getD #[]

/-- `holds-apply`: `f args` with `f : P …`. -/
def tryHoldsApply (c : Ctx) (node : Syntax) (s e : Nat) : IO (Option Repair) := do
  unless isKind ``Term.app node do return none
  let f := node[0]
  let some p := c.predAt (← actualType c s e) | return none
  let fe := (rangeOf f).2
  return some (mkRepair "holds-apply" p.fullName
    s!"{textOf c (rangeOf f).1 fe} applied as a function" #[⟨fe, fe, ".holds"⟩])

/-- `holds-field`: a field or a numeric projection on a chain object of type
    `P …`. The failing step is the first field that resolves to nothing
    (`textDocument/definition` is empty there: the elaborator stopped). -/
def tryHoldsField (c : Ctx) (stack : Array Syntax) (s e : Nat) : IO (Option Repair) := do
  let last := stack.back!
  let chainNode? : Option Syntax :=
    if isKind ``Term.app last then some last[0]
    else stack.reverse.find? fun n => isKind ``Term.proj n || (n.isIdent && (Parse.identComponentRanges c.parsed.text n).size ≥ 2)
  let some chainNode := chainNode? | return none
  let some steps := decompose c chainNode | return none
  let mut i := 0
  let mut failing : Option Nat := none
  for st in steps do
    if (← c.oracle.definition st.fieldStart).isEmpty then
      failing := some i
      break
    i := i + 1
  let some k := failing | return none
  let st := steps[k]!
  let some p := c.predAt (← actualType c s e) | return none
  let fieldText := textOf c st.fieldStart st.fieldEnd
  return some (mkRepair "holds-field" p.fullName
    (if st.numeric then s!"projection on {textOf c (rangeOf chainNode).1 st.objectEnd} through holds"
     else s!"field {fieldText} through holds")
    #[⟨st.objectEnd, st.objectEnd, ".holds"⟩])

/-- `wrap`, `holds-type`, `wrap-nested`: a term whose expected type or own
    type is headed by a converted predicate. -/
def tryMismatch (c : Ctx) (stack : Array Syntax) (t : Syntax) (ds de : Nat) : IO (Option Repair) := do
  let (s, e) := rangeOf t
  -- a projection node has no position of its own for the term-goal query
  -- (its children cover every position); its expected type is the third
  -- expression Lean embedded in the mismatch diagnostic
  let expected : Array Loc ← if isKind ``Term.proj t then do
      let heads ← c.oracle.diagExprHeads ds de
      pure (heads[2]?.getD #[])
    else do
      let r ← c.oracle.expectedHead (freePos t)
      pure (r.getD #[])
  let expPred := c.predAt expected
  let actPred := c.predAt (← actualType c ds de)
  if expPred.isSome && actPred.isNone then
    if isKind ``Term.anonymousCtor t then return none
    return some (mkRepair "wrap" (expPred.get!.fullName)
      (s!"⟨{(textOf c s e).take 40}⟩ expected {expPred.get!.short}")
      #[⟨s, s, "⟨"⟩, ⟨e, e, "⟩"⟩])
  if actPred.isSome && expPred.isNone then
    if needsParens t then
      return some (mkRepair "holds-type" (actPred.get!.fullName)
      (s!"({(textOf c s e).take 40}).holds, has type {actPred.get!.short}")
      #[⟨s, s, "("⟩, ⟨e, e, ").holds"⟩])
    return some (mkRepair "holds-type" (actPred.get!.fullName)
      (s!"{textOf c s e}.holds, has type {actPred.get!.short}")
      #[⟨e, e, ".holds"⟩])
  -- wrap-nested: the argument of `⟨…⟩ : P` whose expected type is P's body,
  -- itself a structure
  if isKind ``Term.anonymousCtor t then return none
  let some parent := stack.reverse.find? (isKind ``Term.anonymousCtor) | return none
  let some p := c.predAt ((← c.oracle.expectedHead (rangeOf parent).1).getD #[]) | return none
  unless p.bodyIsStructure && expected[0]? == p.bodyHead do return none
  return some (mkRepair "wrap-nested" p.fullName
    s!"⟨{(textOf c s e).take 40}⟩ expected the body of {p.short}"
    #[⟨s, s, "⟨"⟩, ⟨e, e, "⟩"⟩])
  return none

/-- `unfold-iff`: `unfold P` cannot unfold a structure. -/
def tryUnfold (c : Ctx) (stack : Array Syntax) (s e : Nat) : IO (Option Repair) := do
  let some (ti, tac) := Parse.innermostTactic c.env stack | return none
  unless firstWord tac == "unfold" do return none
  let some id := stack[ti+1:].toArray.find? (fun n => n.isIdent && rangeOf n == (s, e)) | return none
  let defn ← c.oracle.definition s
  c.note s!"{s}-{e}: unfold name resolves to {repr defn}"
  let some p := c.predAt defn | return none
  let (ts, te) := rangeOf tac
  let tactic := textOf c ts te
  -- `unfold a b c (at h)?`
  let parts := (tactic.splitOn " ").filter (· ≠ "")
  let atIdx := parts.findIdx? (· == "at")
  let names := (parts.drop 1).take ((atIdx.getD parts.length) - 1)
  let atClause := match atIdx with
    | some i => " " ++ " ".intercalate (parts.drop i)
    | none => ""
  let written := textOf c (rangeOf id).1 (rangeOf id).2
  let keep := names.filter (· ≠ written)
  let sep := if startsLine c ts then "\n" ++ indentOf c ts else "; "
  let new := s!"rw [{written}.iff]{atClause}" ++ (if keep.isEmpty then "" else s!"{sep}unfold {" ".intercalate keep}{atClause}")
  return some (mkRepair "unfold-iff" (p.fullName)
      (s!"unfold {written} → rw [{written}.iff]")
      #[⟨ts, te, new⟩])

/-- `simp-iff`: a `simp` family call that lists `P` fails. -/
def trySimp (c : Ctx) (stack : Array Syntax) (s e : Nat) : IO (Option Repair) := do
  let some tac := stack.reverse.find? fun n => Parse.isTacticKind c.env n.getKind && rangeOf n == (s, e) | return none
  let word := firstWord tac
  unless simpWords.contains word || (tac.getKind.toString.splitOn ".").contains "says" do return none
  let mut edits : Array Edit := #[]
  let mut hit : Option Pred := none
  for id in Parse.idents tac do
    let (is, ie) := rangeOf id
    if let some p := c.predAt (← c.oracle.definition is) then
      hit := some p
      edits := edits.push ⟨ie, ie, ".iff"⟩
  let some p := hit | return none
  return some (mkRepair "simp-iff" p.fullName s!"{p.short} → {p.short}.iff in the simp set" edits)

/-- `constructor`: a tactic failed on a goal `P …`. -/
def tryConstructor (c : Ctx) (stack : Array Syntax) : IO (Option Repair) := do
  let some (_, tac) := Parse.innermostTactic c.env stack | return none
  let word := firstWord tac
  unless goalTactics.contains word do return none
  let ts := (rangeOf tac).1
  let some p := c.predAt (← c.oracle.goalHead ts) | return none
  return some (mkRepair "constructor" p.fullName s!"constructor before {word}" #[constructorEdit c ts])

/-- One error diagnostic at byte range `[s, e]` → one repair, or none. -/
def decide (c : Ctx) (s e : Nat) : IO (Option Repair) := do
  let stack := Parse.stackAt c.parsed.commands ⟨s⟩ ⟨e⟩
  if stack.isEmpty then
    c.note s!"{s}-{e}: no syntax covers the range"
    return none
  c.note s!"{s}-{e}: stack {stack.map (·.getKind)}"
  if let some r ← tryUnfold c stack s e then return some r
  if let some r ← trySimp c stack s e then return some r
  if let some r ← tryHoldsField c stack s e then return some r
  if let some t := termAt c stack s e then
    let expected ← c.oracle.expectedHead (freePos t)
    c.note s!"{s}-{e}: term {t.getKind} expected {repr expected} actual {repr (← actualType c s e)}"
    -- an application with no term info did not elaborate: `f` is not a function
    if isKind ``Term.app t && expected.isNone then
      if let some r ← tryHoldsApply c t s e then return some r
    else if let some r ← tryMismatch c stack t s e then return some r
    -- a term that elaborated carries its own error; the tactic around it did not fail on the goal
    if expected.isSome then return none
  tryConstructor c stack

/-- The silent case: every `t.k` whose field index resolves to a converted
    predicate's `holds` projection. Before the conversion `t.k` meant part
    `k` of the body; now it is the whole body. -/
def silentProjections (c : Ctx) : IO (Array Repair) := do
  let mut out := #[]
  for cmd in c.parsed.commands do
    for (proj, idx) in Parse.numericProjections cmd do
      let (is, _) := rangeOf idx
      if let some p := c.holdsAt (← c.oracle.definition is) then
        let dot := (rangeOf proj[1]).1
        out := out.push (mkRepair "holds-proj" (p.fullName)
      ("numeric projection on the converted type")
      #[⟨dot, dot, ".holds"⟩])
  return out

/-- Apply non-overlapping edits, last offset first. -/
def applyEdits (text : String) (edits : Array Edit) : String := Id.run do
  let sorted := edits.qsort fun a b => a.start > b.start || (a.start == b.start && a.stop > b.stop)
  let mut t := text
  for e in sorted do
    t := String.Pos.Raw.extract t 0 ⟨e.start⟩ ++ e.replacement ++ String.Pos.Raw.extract t ⟨e.stop⟩ ⟨t.utf8ByteSize⟩
  return t

/-- Two edits conflict when their ranges overlap or an insertion point touches
    the other edit. -/
def conflicts (a b : Edit) : Bool :=
  if a.start == a.stop && b.start == b.stop then a.start == b.start
  else if a.start == a.stop then b.start ≤ a.start && a.start ≤ b.stop
  else if b.start == b.stop then a.start ≤ b.start && b.start ≤ a.stop
  else max a.start b.start < min a.stop b.stop

end DefToStructure
