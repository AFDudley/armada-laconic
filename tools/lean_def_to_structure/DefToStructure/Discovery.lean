/-
  Discovery and conversion (see ../README.md, "Discovery").

  `importTarget` loads the target checkout's modules into this process with
  `importModules` (search path from `lake env printenv LEAN_PATH`). The
  environment serves two purposes: discovery here, and parsing of the target
  files (`Parse.parseFile`) during the repair loop.

  `describe` reports one predicate: its module file, its declaration range
  (`findDeclarationRanges?`), and the pieces of its `def` syntax from the
  parsed declaration. `conversionEdit` builds the replacement text: the
  one-field structure and the `P.iff` lemma.
-/
import Lean
import DefToStructure.Patterns
import DefToStructure.Proc

open Lean Meta Elab

namespace DefToStructure.Discovery

/-- The module of `n`, when `n` comes from an imported module. -/
def moduleOf? (env : Environment) (n : Name) : Option Name :=
  (env.getModuleIdxFor? n).map fun i => env.header.moduleNames[i.toNat]!

/-- `ArmadaLean.Nitro.Value` → `ArmadaLean/Nitro/Value.lean`. -/
def modulePath (m : Name) : String :=
  (m.components.map toString |> String.intercalate "/") ++ ".lean"

/-- `ArmadaLean/Nitro/Value.lean` → `ArmadaLean.Nitro.Value`. -/
def moduleName (rel : String) : Name :=
  let noExt := if rel.endsWith ".lean" then (rel.dropEnd 5).copy else rel
  (noExt.splitOn "/").foldl (fun n c => Name.str n c) Name.anonymous

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

/-- One binder of the `def`: its source text, its kind, the names it binds. -/
structure Binder where
  text : String
  explicit : Bool
  names : Array String
  deriving Repr, ToJson, FromJson

/-- Everything the conversion needs about one `def`. Offsets are bytes into
    the file text. -/
structure Found where
  name : Name
  declId : String
  file : String
  /-- `some` when the constant is already a one-field `holds` structure with `.iff`. -/
  alreadyConverted : Bool
  defStart : Nat
  bodyStart : Nat
  bodyEnd : Nat
  binders : Array Binder
  typeText : Option String
  bodyHead : Option Loc
  bodyIsStructure : Bool
  deriving Repr

/-- The pieces of a `def` declaration. -/
def defPieces (stx : Syntax) : Option (Syntax × Syntax × Array Syntax × Option Syntax × Syntax) := do
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
  pure (kw, declId, binders, typeSpec, val[1])

def binderFacts (text : String) (b : Syntax) : Binder :=
  let explicit := b.isOfKind ``Parser.Term.explicitBinder
  let names : Array String :=
    if b.isOfKind ``Parser.Term.instBinder then #[]
    else b[1].getArgs.map fun id => id.getId.toString
  let (s, e) := match Parse.range? b with | some (s, e) => (s, e) | none => (0, 0)
  { text := String.Pos.Raw.extract text s e, explicit, names }

/-- Declaration location of `n`: its module file and the start of its name. -/
def declLoc (env : Environment) (root : System.FilePath) (n : Name) : MetaM (Option Loc) := do
  let some m := moduleOf? env n | return none
  unless (`ArmadaLean).isPrefixOf m do return none
  let some rng ← findDeclarationRanges? n | return none
  let rel := modulePath m
  let p := rng.selectionRange.pos
  return some { uri := s!"file://{root / rel}", line := p.line - 1, character := p.column }

def describe (env : Environment) (root : System.FilePath) (n : Name) : MetaM Found := do
  let ci ← getConstInfo n
  let some m := moduleOf? env n | throwError "no module for {n}"
  let rel := modulePath m
  let some rng ← findDeclarationRanges? n | throwError "no declaration range for {n}"
  let isProp ← forallTelescope ci.type fun _ body => pure body.isProp
  let isStruct := isStructure env n
  let holdsStructure := isStruct && getStructureFields env n == #[`holds]
  let hasIff := env.contains (n ++ `iff)
  let text ← IO.FS.readFile (root / rel)
  let fm := FileMap.ofString text
  let startPos := fm.ofPosition rng.range.pos
  let endPos := fm.ofPosition rng.range.endPos
  if holdsStructure && hasIff then
    return { name := n, declId := n.toString, file := rel, alreadyConverted := true, defStart := 0,
             bodyStart := 0, bodyEnd := 0, binders := #[], typeText := none, bodyHead := none,
             bodyIsStructure := false }
  let .defnInfo d := ci | throwError "{n} is not a `def` (and not a converted structure)"
  unless isProp do throwError "{n} is not a `Prop`"
  if hasIff then throwError "{n}.iff already exists; refused"
  let slice := String.Pos.Raw.extract text startPos endPos
  let stx ← match Parser.runParserCategory env `command slice rel with
    | .ok stx => pure stx
    | .error err => throwError "{n}: the declaration did not parse: {err}"
  let some (kw, declId, binders, typeSpec, body) := defPieces stx
    | throwError "{n}: the declaration is not a plain `def … := …`"
  let shift (p : Option String.Pos.Raw) : Nat := startPos.byteIdx + (p.getD 0).byteIdx
  let (head, whnfHead) ← lambdaTelescope d.value fun _ b => do
    let h := b.getAppFn.constName?
    let b' ← whnf b
    pure (h, b'.getAppFn.constName?)
  let bodyIsStructure := match whnfHead with
    | some h => isStructure env h
    | none => false
  let bodyHead ← match head with
    | some h => declLoc env root h
    | none => pure none
  let typeText := typeSpec.map fun t => t.reprint.get!.trimAscii.copy
  return {
    name := n, declId := declId[0].getId.toString, file := rel, alreadyConverted := false,
    defStart := shift (kw.getPos?), bodyStart := shift (body.getPos?), bodyEnd := shift (body.getTailPos?),
    binders := binders.map (binderFacts slice), typeText, bodyHead, bodyIsStructure }

/-- The body text with its continuation lines re-indented to `cont`. -/
def reindent (body : String) (cont : String) : String := Id.run do
  let lines := body.splitOn "\n"
  if lines.length == 1 then return body
  let rest := (lines.drop 1).filter fun l => !l.trimAscii.isEmpty
  let indentOf (l : String) : Nat := (l.takeWhile (· == ' ')).length
  let common := (rest.map indentOf).foldl min (rest.headD "" |> indentOf)
  let out := lines.head! :: (lines.drop 1).map fun l =>
    if l.trimAscii.isEmpty then "" else cont ++ (l.drop common)
  return "\n".intercalate out

/-- The replacement for the `def`: the structure and the `.iff` lemma. Returns
    the edit, the offsets (relative to the edit start) of the name token in
    `structure <name>` and of the `holds` token. -/
def conversionEdit (f : Found) (text : String) : Except String (Edit × Nat × Nat) := do
  if let some t := f.typeText then
    if t ≠ "Prop" then throw s!"{f.name}: type is `{t}`, not `Prop`"
  let explicit := f.binders.filter (·.explicit) |>.toList.flatMap (·.names.toList)
  if explicit.contains "_" then throw s!"{f.name}: an explicit binder is `_`; `P.iff` cannot apply it"
  let bText := if f.binders.isEmpty then "" else " " ++ " ".intercalate (f.binders.toList.map (·.text))
  let body := String.Pos.Raw.extract text ⟨f.bodyStart⟩ ⟨f.bodyEnd⟩
  let app := " ".intercalate (f.declId :: explicit)
  let structureText := s!"structure {f.declId}{bText} : Prop where\n  holds : {reindent body "    "}"
  let iff := s!"/-- `{f.declId}`'s structure carries exactly its old `def` body; a citing statement's\n" ++
    s!"    meaning did not change when `{f.declId}` became a structure. -/\n" ++
    s!"theorem {f.declId}.iff{bText} :\n    {app} ↔ {reindent body "      "} :=\n" ++
    "  ⟨fun h => h.holds, fun h => ⟨h⟩⟩"
  let replacement := structureText ++ "\n\n" ++ iff
  let nameOff := "structure ".utf8ByteSize
  let holdsOff := s!"structure {f.declId}{bText} : Prop where\n  ".utf8ByteSize
  return (⟨f.defStart, f.bodyEnd, replacement⟩, nameOff, holdsOff)

/-- `LEAN_PATH` of the target checkout, from `lake env`. -/
def targetSearchPath (root : System.FilePath) : IO (List System.FilePath) := do
  let sa : IO.Process.SpawnArgs := { cmd := "lake", args := #["env", "printenv", "LEAN_PATH"], cwd := root, env := ← leanEnv }
  let out ← IO.Process.output sa
  if out.exitCode != 0 then throw (IO.userError s!"lake env printenv LEAN_PATH failed: {out.stderr}")
  return (out.stdout.trimAscii.copy.splitOn ":").filter (· ≠ "") |>.map System.FilePath.mk

/-- Import `modules` from the target's search path into this process. -/
unsafe def importTarget (root : System.FilePath) (modules : Array Name) : IO Environment := do
  let sp ← targetSearchPath root
  let home := (← IO.getEnv "HOME").getD ""
  initSearchPath (← findSysroot s!"{home}/.elan/bin/lean") sp
  enableInitializersExecution
  importModules (modules.map fun m => { module := m }) {} (trustLevel := 1024) (loadExts := true)

/-- The imports of a file, from its header only (no environment needed). -/
def fileImports (path : System.FilePath) : IO (Array Name) := do
  let text ← IO.FS.readFile path
  let (imports, _, _) ← Lean.Elab.parseImports text (some path.toString)
  return imports.map (·.module)

end DefToStructure.Discovery
