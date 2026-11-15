/-
  Scope measurement for `tools/lean_def_to_structure` (see README.md there).

  Run from an armada-lean checkout after `lake build`:

    lake env lean <this file>

  It prints one JSON line per `repeatedPremise` group whose shape's head is
  an `ArmadaLean`-declared `def … : Prop` that is not a structure and not a
  structure projection. Those are the groups a def-to-structure conversion
  removes. The last line is a JSON
  summary with `"summary": true`.
-/
import ArmadaLean
import ArmadaLean.Lint.TypeDoctrine

open Lean Meta ArmadaLean.Lint.TypeDoctrine

/-- The head constant of every non-excluded hypothesis conjunct of `n`, keyed
    the same way `analyzeDecl` keys `repeatedPremise` groups. -/
def conjunctHeads (env : Environment) (n : Name) : MetaM (Array (String × Option Name)) := do
  let info ← getConstInfo n
  forallTelescope info.type fun args concl => do
    let mut dataB : Array FVarId := #[]
    let mut hypB : Array FVarId := #[]
    for a in args do
      let ldecl ← a.fvarId!.getDecl
      if ldecl.binderInfo.isInstImplicit || ldecl.type.isSort then
        continue
      else if ← Meta.isProp ldecl.type then
        hypB := hypB.push a.fvarId!
      else
        dataB := dataB.push a.fvarId!
    let mut out := #[]
    for h in hypB do
      let ldecl ← h.getDecl
      if ← isConstructorApplicationLemma env concl h then continue
      for e in splitAnd ldecl.type do
        if ← premiseExcluded env e then continue
        if ← constructorPremiseExcluded env concl dataB e then continue
        out := out.push (normalizedKey e, e.getAppFn.constName?)
    return out

/-- `n` is a `def` whose type is a `∀`-telescope ending in `Prop`. -/
def isPropDef (env : Environment) (n : Name) : MetaM Bool := do
  let some (.defnInfo d) := env.find? n | return false
  forallTelescope d.type fun _ body => pure body.isProp

def report : MetaM Unit := do
  let env ← getEnv
  let data ← computeGlobalData
  let mut heads : Std.HashMap String Name := {}
  for d in data.decls do
    for (key, head) in ← conjunctHeads env d.name do
      if let some h := head then
        heads := heads.insert key h
  let mut count := 0
  let mut names : Array String := #[]
  for (key, uses) in data.repeated do
    let some head := heads.get? key | continue
    unless isArmadaConst env head do continue
    if isStructure env head then continue
    if env.isProjectionFn head then continue
    unless ← isPropDef env head do continue
    let decls := (uses.map Prod.fst).toList.eraseDups.map toString
    IO.println (Json.compress (Json.mkObj [
      ("head", toJson head.toString), ("shape", toJson key),
      ("example", toJson uses[0]!.2), ("declarations", toJson decls)]))
    count := count + 1
    unless names.contains head.toString do names := names.push head.toString
  IO.println (Json.compress (Json.mkObj [
    ("summary", true), ("groups", count), ("predicates", toJson (names.qsort (· < ·)))]))

#eval report
