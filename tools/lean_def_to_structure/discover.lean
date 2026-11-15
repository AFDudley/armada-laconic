/-
  Discovery half of `tools/lean_def_to_structure` (see README.md there).

  Run from an armada-lean checkout:

    CODEMOD_NAMES="RegistryConsistent,Adjudicator.IsDAG" lake env lean <this file>

  For each name in `CODEMOD_NAMES` the script prints one JSON line that
  describes the predicate: its full name, its module file, its declaration
  range, and the byte offsets of the pieces of its `def` syntax. It reads
  the pieces from the parsed syntax of the declaration, not from a regular
  expression over the source. `codemod.py` applies the edit at the
  reported offsets.

  The last line is a JSON summary with `"summary": true`.
-/
import ArmadaLean
import Lean

open Lean Meta Elab

/-- The module of `n`, when `n` comes from an imported module. -/
def moduleOf? (env : Environment) (n : Name) : Option Name :=
  (env.getModuleIdxFor? n).map fun i => env.header.moduleNames[i.toNat]!

/-- `ArmadaLean.Nitro.Value` → `ArmadaLean/Nitro/Value.lean`. -/
def modulePath (m : Name) : String :=
  (m.components.map toString |> String.intercalate "/") ++ ".lean"

/-- `suffix`'s components end `n`'s components. -/
def suffixOf (suffix n : Name) : Bool :=
  let s := suffix.components
  let c := n.components
  c.length ≥ s.length && c.drop (c.length - s.length) == s

/-- For each requested suffix, every constant of an `ArmadaLean` module whose
    name ends in it. One pass over the environment. -/
def resolveAll (env : Environment) (suffixes : Array Name) : Array (Array Name) := Id.run do
  let mut out : Array (Array Name) := suffixes.map fun _ => #[]
  for (n, _) in env.constants.toList do
    if n.isInternal then continue
    let some m := moduleOf? env n | continue
    unless (`ArmadaLean).isPrefixOf m do continue
    for i in [0:suffixes.size] do
      if suffixOf suffixes[i]! n then
        out := out.modify i (·.push n)
  return out.map (·.qsort (·.toString < ·.toString))

/-- Byte offset of the start and end of `stx`, shifted by `base`. -/
def span (base : String.Pos.Raw) (stx : Syntax) : Option (Nat × Nat) := do
  let s ← stx.getPos?
  let e ← stx.getTailPos?
  pure (base.byteIdx + s.byteIdx, base.byteIdx + e.byteIdx)

def spanJson (base : String.Pos.Raw) (stx : Syntax) : Json :=
  match span base stx with
  | some (s, e) => Json.mkObj [("start", s), ("end", e)]
  | none => Json.null

/-- The binder kind of one `bracketedBinder` and the names it binds. -/
def binderFacts (b : Syntax) : Json :=
  let kind :=
    if b.isOfKind ``Parser.Term.explicitBinder then "explicit"
    else if b.isOfKind ``Parser.Term.implicitBinder then "implicit"
    else if b.isOfKind ``Parser.Term.strictImplicitBinder then "strictImplicit"
    else if b.isOfKind ``Parser.Term.instBinder then "inst"
    else "other"
  let names : Array String :=
    if kind == "inst" then #[]
    else (b[1].getArgs.map fun id => id.getId.toString)
  Json.mkObj [("kind", kind), ("names", toJson names)]

/-- The pieces of a `def` declaration: `(defKeyword, declId, binders, typeSpec,
    assign, body)`. `none` when `stx` is not a plain `def … := …`. -/
def defPieces (stx : Syntax) : Option (Syntax × Syntax × Array Syntax × Option Syntax × Syntax × Syntax) := do
  guard (stx.isOfKind ``Parser.Command.declaration)
  let d := stx[1]
  guard (d.isOfKind ``Parser.Command.definition)
  let kw := d[0]
  let declId := d[1]
  let sig := d[2]
  let val := d[3]
  guard (val.isOfKind ``Parser.Command.declValSimple)
  let binders := sig[0].getArgs
  let typeSpec := if sig[1].getNumArgs == 0 then none else some sig[1][0][1]
  pure (kw, declId, binders, typeSpec, val[0], val[1])

/-- The head constant of `e` and whether that head is a structure. Follows
    `def` bodies through `whnf` so that `IsDAG`'s body `Acyclic s.Edge`
    reports `Std.Irrefl`-style structure heads. -/
def bodyHead (env : Environment) (e : Expr) : MetaM Json := do
  let head := e.getAppFn.constName?
  let e' ← whnf e
  let head' := e'.getAppFn.constName?
  let isStruct := match head' with
    | some h => isStructure env h
    | none => false
  pure <| Json.mkObj [
    ("head", toJson (head.map toString)),
    ("whnfHead", toJson (head'.map toString)),
    ("whnfHeadIsStructure", toJson isStruct)]

def describe (env : Environment) (requested : String) (n : Name) : MetaM Json := do
  let ci ← getConstInfo n
  let some m := moduleOf? env n | throwError "no module for {n}"
  let path := modulePath m
  let some rng ← findDeclarationRanges? n | throwError "no declaration range for {n}"
  let isProp ← forallTelescope ci.type fun _ body => pure body.isProp
  let kind := match ci with
    | .defnInfo _ => "def"
    | .inductInfo _ => if isStructure env n then "structure" else "inductive"
    | .thmInfo _ => "theorem"
    | _ => "other"
  let hasIff := env.contains (n ++ `iff)
  let holdsField := isStructure env n && (getStructureFields env n) == #[`holds]
  let src ← IO.FS.readFile path
  let fm := FileMap.ofString src
  let startPos := fm.ofPosition rng.range.pos
  let endPos := fm.ofPosition rng.range.endPos
  let slice := String.Pos.Raw.extract src startPos endPos
  let mut syntaxJson := Json.null
  let mut parseError := Json.null
  if kind == "def" then
    match Parser.runParserCategory env `command slice path with
    | .error err => parseError := toJson err
    | .ok stx =>
      match defPieces stx with
      | none => parseError := toJson "declaration is not a plain `def … := …`"
      | some (kw, declId, binders, typeSpec, assign, body) =>
        let bodyExpr ← match ci with
          | .defnInfo d => pure (some d.value)
          | _ => pure none
        let headJson ← match bodyExpr with
          | some v => lambdaTelescope v fun _ b => bodyHead env b
          | none => pure Json.null
        syntaxJson := Json.mkObj [
          ("defKeyword", spanJson startPos kw),
          ("declId", spanJson startPos declId),
          ("declIdText", toJson declId[0].getId.toString),
          ("binders", Json.arr (binders.map fun b =>
            Json.mkObj [("span", spanJson startPos b), ("facts", binderFacts b)])),
          ("typeSpec", match typeSpec with
            | some t => Json.mkObj [("span", spanJson startPos t), ("text", toJson (t.reprint.get!.trimAscii.copy))]
            | none => Json.null),
          ("assign", spanJson startPos assign),
          ("body", spanJson startPos body),
          ("bodyHead", headJson)]
  pure <| Json.mkObj [
    ("requested", toJson requested),
    ("name", toJson n.toString),
    ("module", toJson m.toString),
    ("path", toJson path),
    ("kind", toJson kind),
    ("isProp", toJson isProp),
    ("hasIff", toJson hasIff),
    ("holdsStructure", toJson holdsField),
    ("range", Json.mkObj [
      ("startLine", rng.range.pos.line), ("startCol", rng.range.pos.column),
      ("endLine", rng.range.endPos.line), ("endCol", rng.range.endPos.column),
      ("startByte", startPos.byteIdx), ("endByte", endPos.byteIdx)]),
    ("syntax", syntaxJson),
    ("parseError", parseError)]

def discoverMain : MetaM Unit := do
  let env ← getEnv
  let raw := (← IO.getEnv "CODEMOD_NAMES").getD ""
  let names := ((raw.splitOn ",").map (fun s => s.trimAscii.copy) |>.filter (· ≠ "")).toArray
  let resolved := resolveAll env (names.map String.toName)
  let mut ok := 0
  let mut failed := 0
  for i in [0:names.size] do
    let s := names[i]!
    let cands := resolved[i]!
    if cands.size == 1 then
      IO.println (Json.compress (← describe env s cands[0]!))
      ok := ok + 1
    else
      IO.println (Json.compress (Json.mkObj [
        ("requested", toJson s), ("error", toJson "name does not resolve to exactly one ArmadaLean constant"),
        ("candidates", toJson (cands.map toString))]))
      failed := failed + 1
  IO.println (Json.compress (Json.mkObj [("summary", true), ("resolved", ok), ("failed", failed)]))

#eval discoverMain
