import ArmadaLean
open Lean Meta

/-! Audit fact extraction: one JSON line per ArmadaLean declaration (selected by module). -/

def modOf (env : Environment) (n : Name) : Name :=
  match env.getModuleIdxFor? n with
  | some i => env.header.moduleNames[i.toNat]!
  | none => .anonymous

def inArmada (env : Environment) (n : Name) : Bool :=
  (`ArmadaLean).isPrefixOf (modOf env n) && !n.isInternal

def nitroBase : List Name := [`ArmadaLean.Nitro.Adjudicator.apply, `ArmadaLean.Nitro.NetworkState.applyAdv,
  `ArmadaLean.Nitro.Adjudicator.transfer, `ArmadaLean.Nitro.Adjudicator.challenge,
  `ArmadaLean.Nitro.Adjudicator.checkpoint, `ArmadaLean.Nitro.Adjudicator.conclude, `ArmadaLean.Nitro.Adjudicator.reclaim]
def poolBase : List Name := [`removeAll, `markSpent, `applyRecords, `recordPreimages, `insertLeaves, `insertTree,
  `insertStart, `withTreePositionsFrom, `Step.casesOn]

def cites (d : String) : Bool :=
  ["Nitro:", "SATP:", "ForceMove:", ".sol:", ".circom:", "operator ruling", "Operator ruling"].any
    (fun k => (d.splitOn k).length > 1)

def jstr (s : String) : String := (Json.str s).compress

#eval show MetaM Unit from do
  let env ← getEnv
  let decls := env.constants.toList.filter (fun (n, _) => inArmada env n)
  -- reverse dependencies inside ArmadaLean
  let mut rev : Std.HashMap Name Nat := {}
  for (n, ci) in decls do
    let used := (ci.type.getUsedConstants ++ ((ci.value? (allowOpaque := true)).map (·.getUsedConstants) |>.getD #[]))
    for u in used.toList.eraseDups do
      if u != n && inArmada env u then rev := rev.insert u (rev.getD u 0 + 1)
  let h ← IO.FS.Handle.mk ".scratch/audit/facts.jsonl" .write
  for (n, ci) in decls do
    let kind := match ci with
      | .thmInfo _ => "theorem" | .defnInfo _ => "def" | .inductInfo _ => "inductive"
      | .axiomInfo _ => "axiom" | .ctorInfo _ => "ctor" | .recInfo _ => "rec" | _ => "other"
    if kind == "ctor" || kind == "rec" then continue
    let rng ← findDeclarationRanges? n
    let (l0, l1) := match rng with
      | some r => (r.range.pos.line, r.range.endPos.line) | none => (0, 0)
    let doc := (← findDocString? env n).getD ""
    let mut propB := 0; let mut dataB := 0; let mut subB := 0
    if kind == "theorem" then
      let (bs, _) ← forallMetaTelescope ci.type
      for b in bs do
        let ty ← inferType b
        if (← isProp ty) then propB := propB + 1 else
          dataB := dataB + 1
          if (← whnfR ty).getAppFn.constName? == some ``Subtype then subB := subB + 1
    let usedV := ((ci.value? (allowOpaque := true)).map (·.getUsedConstants) |>.getD #[])
    let nb := nitroBase.filter (usedV.contains ·) |>.length
    let pb := poolBase.filter (usedV.contains ·) |>.length
    let hasSorry := usedV.contains ``sorryAx
    let line := "{" ++ s!"\"name\":{jstr n.toString},\"module\":{jstr (modOf env n).toString},\"kind\":{jstr kind}," ++
      s!"\"line\":{l0},\"lines\":{l1 - l0 + 1},\"cites\":{cites doc},\"prop_binders\":{propB},\"data_binders\":{dataB}," ++
      s!"\"subtype_binders\":{subB},\"nitro_base_ops\":{nb},\"pool_base_ops\":{pb},\"users\":{rev.getD n 0},\"sorry\":{hasSorry}" ++ "}"
    h.putStrLn line
  IO.println s!"wrote {decls.length} candidates"
