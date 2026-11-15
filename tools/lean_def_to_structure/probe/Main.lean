/-
  Elaboration probe of `tools/lean_def_to_structure` (see README.md there).

  Build once with `lake build` in this directory (same toolchain as the
  target). Run from an armada-lean checkout, after `lake build` of the
  file's imports, under `lake env` so that `LEAN_PATH` is set:

    lake env <this dir>/.lake/build/bin/probe <file> <pred>,<pred>,…

  The script elaborates the file the way `lean` does and prints one JSON line
  per message (`pos`, `endPos`, `severity`, `data`), the same fields as
  `lean --json`. It also walks the info trees and prints one JSON line
  (`"projection": true`) for every numeric projection `t.1` that resolves to
  the projection function `P.holds` of a converted predicate `P`. Before the conversion
  `t.1` meant the first part of `P`'s body; after it, `t.1` is the whole body
  and Lean reports the mismatch far from `t.1`, or not at all. The driver
  inserts `.holds` before that `.1`.

  The last line is a JSON summary with `"summary": true`.
-/
import Lean

open Lean Elab

def posJson (p : Position) : Json :=
  Json.mkObj [("line", p.line), ("column", p.column)]

def severityStr : MessageSeverity → String
  | .error => "error"
  | .warning => "warning"
  | .information => "information"

def messageJson (m : Message) : IO Json := do
  let data ← m.data.toString
  pure <| Json.mkObj [
    ("pos", posJson m.pos),
    ("endPos", match m.endPos with | some p => posJson p | none => Json.null),
    ("severity", severityStr m.severity),
    ("data", data)]

/-- `(fieldIdxPos, predicate)` for every numeric projection `t.1` that resolves
    to the projection function `P.holds` of a `P` in `preds`. `elabAppLValsAux`
    records one `TermInfo` per projection: an `identProjKind` node around the
    `fieldIdx` token, with the bare projection constant as its term
    (`Lean/Elab/App.lean`, `addProjTermInfo lval.getRef projFn`). -/
def projections (preds : Array Name) (trees : PersistentArray InfoTree) :
    Array (String.Pos.Raw × Name) := Id.run do
  let mut out : Array (String.Pos.Raw × Name) := #[]
  for tree in trees do
    let hits := tree.foldInfo (init := (#[] : Array (String.Pos.Raw × Name))) fun _ info acc =>
      match info with
      | .ofTermInfo ti =>
        if ti.stx.isOfKind Parser.Term.identProjKind && ti.stx[0].isOfKind `fieldIdx then
          match ti.expr, ti.stx[0].getPos? with
          | .const c _, some pos =>
            match preds.find? (fun p => c == p ++ `holds) with
            | some p => acc.push (pos, p)
            | none => acc
          | _, _ => acc
        else acc
      | _ => acc
    out := out ++ hits
  return out

unsafe def main (args : List String) : IO UInt32 := do
  let [path, predsRaw] := args | do
    IO.eprintln "usage: probe <file.lean> <pred>,<pred>,…"
    return 2
  enableInitializersExecution
  let preds := ((predsRaw.splitOn ",").map (fun s => s.trimAscii.copy)
    |>.filter (· ≠ "") |>.map String.toName).toArray
  initSearchPath (← findSysroot)
  let input ← IO.FS.readFile path
  let inputCtx := Parser.mkInputContext input path
  let (header, parserState, messages) ← Parser.parseHeader inputCtx
  let (env, messages) ← processHeader header {} messages inputCtx (trustLevel := 1024)
  let cmdState := Command.mkState env messages {}
  let cmdState := { cmdState with infoState := { enabled := true } }
  let s ← IO.processCommands inputCtx parserState cmdState
  let mut errors : Nat := 0
  for m in s.commandState.messages.toList do
    IO.println (Json.compress (← messageJson m))
    if m.severity == .error then errors := errors + 1
  let fm := inputCtx.fileMap
  let mut seen : Std.HashSet Nat := {}
  let mut count : Nat := 0
  for (pos, p) in projections preds s.commandState.infoState.trees do
    if seen.contains pos.byteIdx then continue
    seen := seen.insert pos.byteIdx
    let lc := fm.toPosition pos
    IO.println (Json.compress (Json.mkObj [
      ("projection", true), ("pos", posJson lc), ("predicate", toJson p.toString)]))
    count := count + 1
  IO.println (Json.compress (Json.mkObj [
    ("summary", true), ("errors", errors), ("projections", count)]))
  return 0
