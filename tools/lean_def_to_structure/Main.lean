/-
  `def_to_structure`: the driver (see README.md).

    def_to_structure --checkout <armada-lean worktree> --slot <1-4> \
      --names RegistryConsistent,Adjudicator.IsDAG --report report.json [--resume]

  Steps:
  1. `lake build`, the linter modules, `scripts/type_lint.lean` (the "before" list).
  2. Import the target's modules into this process (discovery, and the parser
     for the repair loop).
  3. Discovery and the guards (`ArmadaLean/Spec/`, `git grep Spec\.<name>`).
  4. Conversion: `def` → `structure … where holds : body`, plus `P.iff`.
  5. The files that import a converted module, in dependency order, each
     through `lake serve`: open, read the diagnostics, decide one repair per
     error from the elaboration data (`DefToStructure.Patterns`), apply, send
     `didChange`, repeat until the file has no error and no silent projection.
     An edited module is rebuilt (`lake build <Module>`) before the next file
     opens. A round with an error and no repair stops the tool.
  6. One full check: `lake build`, the linter modules, `scripts/type_lint.lean`
     (the "after" list), `scripts/update_gate_baseline.py --hashes-only`,
     `bash scripts/check.sh --json`, and the JSON report.

  `--resume` continues a stopped run after a hand edit: it skips steps 1, 3
  and 4 and reads `<checkout>/.scratch/codemod/state.json`.
-/
import Lean
import DefToStructure.Lsp
import DefToStructure.Parse
import DefToStructure.Patterns
import DefToStructure.Discovery

open Lean DefToStructure DefToStructure.Discovery

namespace DefToStructure

structure Args where
  checkout : System.FilePath
  slot : Nat
  names : Array String
  report : System.FilePath
  resume : Bool

def parseArgs (args : List String) : Except String Args := do
  let mut checkout := ""
  let mut slot := 0
  let mut names := ""
  let mut report := ""
  let mut resume := false
  let mut rest := args
  while !rest.isEmpty do
    match rest with
    | "--checkout" :: v :: r => checkout := v; rest := r
    | "--slot" :: v :: r => slot := v.toNat!; rest := r
    | "--names" :: v :: r => names := v; rest := r
    | "--report" :: v :: r => report := v; rest := r
    | "--resume" :: r => resume := true; rest := r
    | a :: _ => throw s!"unknown argument {a}"
    | [] => pure ()
  if checkout.isEmpty || slot == 0 || names.isEmpty || report.isEmpty then
    throw "usage: def_to_structure --checkout DIR --slot N --names A,B --report FILE [--resume]"
  return { checkout, slot, names := (names.splitOn ",").map (·.trimAscii.copy) |>.filter (· ≠ "") |>.toArray,
           report, resume }

/-- A stop: the tool changes nothing more and reports `msg`. -/
structure StopError where
  msg : String

structure RepairRecord where
  file : String
  line : Nat
  col : Nat
  pattern : String
  pred : Option String
  detail : String
  deriving ToJson, FromJson, Inhabited

structure Timing where
  command : String
  seconds : Float
  exit : Nat
  deriving ToJson, FromJson

/-- The last `n` bytes of `s` (whole characters). -/
def tail (s : String) (n : Nat) : String :=
  if s.length ≤ n then s else (s.toSubstring.drop (s.length - n)).toString

structure PredReport where
  name : String
  path : String
  line : Nat
  alreadyConverted : Bool
  iff : String
  deriving ToJson, FromJson

/-- The persisted state of a run (`.scratch/codemod/state.json`). -/
structure State where
  head : String
  names : Array String
  preds : Array Pred
  predReports : Array PredReport
  files : Array String
  nextIndex : Nat
  repairs : Array RepairRecord
  rounds : Array Json
  stops : Array String
  lintBefore : Json
  timings : Array Timing
  deriving ToJson, FromJson

structure Env where
  args : Args
  root : System.FilePath
  logDir : System.FilePath
  state : IO.Ref State
  lintBeforeFindings : IO.Ref (Array String)
  seen : IO.Ref (Std.HashSet String)

abbrev M := ReaderT Env (ExceptT StopError IO)

def stop (msg : String) : M α := throw ⟨msg⟩

def modifyState (f : State → State) : M Unit := do (← read).state.modify f

def saveState : M Unit := do
  let e ← read
  IO.FS.writeFile (e.logDir / "state.json") ((toJson (← e.state.get)).pretty ++ "\n")

def log (s : String) : M Unit := IO.eprintln s

-- --- processes ---------------------------------------------------------------

/-- Run a command in the checkout under the run's `flock` slot, logging its output. -/
def runCmd (args : Array String) (logName : String) (cwd : Option System.FilePath := none) : M IO.Process.Output := do
  let e ← read
  let home := (← IO.getEnv "HOME").getD ""
  let lock := s!"{home}/.cache/lean-run/locks/slot-{e.args.slot}.lock"
  let t0 ← IO.monoMsNow
  let spawnArgs : IO.Process.SpawnArgs :=
    { cmd := "flock", args := #[lock, "nice"] ++ args, cwd := cwd.getD e.root, env := ← leanEnv }
  let out ← IO.Process.output spawnArgs
  let dt := (Float.ofNat ((← IO.monoMsNow) - t0)) / 1000.0
  IO.FS.writeFile (e.logDir / logName) (out.stdout ++ "\n--- stderr ---\n" ++ out.stderr)
  modifyState fun s => { s with timings := s.timings.push ⟨" ".intercalate args.toList, dt, out.exitCode.toNat⟩ }
  log s!"[lean] {(" ".intercalate args.toList).take 80} … exit {out.exitCode} in {dt}s"
  return out

def git (args : Array String) : M IO.Process.Output := do
  let sa : IO.Process.SpawnArgs := { cmd := "git", args, cwd := (← read).root }
  IO.Process.output sa

/-- `lake build`, then the linter modules the gate builds too. -/
def fullBuild (tag : String) : M (Bool × String) := do
  let p1 ← runCmd #["lake", "build"] s!"build-{tag}.log"
  if p1.exitCode != 0 then return (false, p1.stdout ++ p1.stderr)
  let p2 ← runCmd #["lake", "build", "ArmadaLean.Lint.TypeDoctrine", "ArmadaLean.Lint.TypeDoctrineTest"]
    s!"build-lint-{tag}.log"
  return (p2.exitCode == 0, p1.stdout ++ p1.stderr ++ p2.stdout ++ p2.stderr)

/-- `scripts/type_lint.lean`: the finding lines and the summary. -/
def typeLint (tag : String) : M (Array String × Json) := do
  let p ← runCmd #["lake", "env", "lean", "scripts/type_lint.lean"] s!"lint-{tag}.log"
  let lines := p.stdout.splitOn "\n"
  let findings := lines.filter (fun l => l.startsWith "{" && !l.startsWith "{\"summary\"") |>.toArray
  let summaries := lines.filter (·.startsWith "{\"summary\"")
  if p.exitCode != 0 || summaries.isEmpty then
    stop s!"scripts/type_lint.lean failed (exit {p.exitCode}):\n{(tail p.stdout 2000)}\n{(tail p.stderr 2000)}"
  IO.FS.writeFile ((← read).logDir / s!"lint-{tag}.jsonl") ("\n".intercalate findings.toList ++ "\n")
  match Json.parse summaries.getLast! with
  | .ok j => return (findings, j)
  | .error err => stop s!"type_lint summary is not JSON: {err}"

/-- A finding line without its `example` field: `repeatedPremise` picks the
    example from the group's uses, and that pick differs between runs. -/
def findingKey (line : String) : String :=
  match Json.parse line with
  | .ok (.obj kvs) => (Json.obj (kvs.erase "example")).compress
  | _ => line

-- --- the repair loop --------------------------------------------------------

def forbidden : List String := ["sorry", "admit", "axiom", "native_decide"]

def countOcc (hay needle : String) : Nat := (hay.splitOn needle).length - 1

def checkForbidden (before after : String) (where_ : String) : M Unit := do
  for w in forbidden do
    if countOcc after w > countOcc before w then
      stop s!"{where_}: the edit would insert `{w}`"

def writeFile (rel : String) (text : String) : M Unit := do
  if rel.startsWith "ArmadaLean/Spec/" then stop s!"{rel}: the tool never edits ArmadaLean/Spec/"
  IO.FS.writeFile ((← read).root / rel) text

/-- Shift the predicate offsets of `rel` by the edits just applied to it. -/
def shiftPreds (rel : String) (edits : Array Edit) : M Unit := do
  let shift (off : Nat) : Nat := Id.run do
    let mut o := off
    for e in edits do
      if e.stop ≤ off then o := o + e.replacement.utf8ByteSize - (e.stop - e.start)
    return o
  modifyState fun s => { s with preds := s.preds.map fun p =>
    if p.file == rel then { p with declIdOffset := shift p.declIdOffset, holdsOffset := shift p.holdsOffset } else p }

def predTexts : M PredTexts := do
  let e ← read
  let s ← e.state.get
  let mut m : Std.HashMap String FileMap := {}
  for p in s.preds do
    unless m.contains p.file do
      m := m.insert p.file (FileMap.ofString (← IO.FS.readFile (e.root / p.file)))
  return { fileMaps := m }

def locOfLink (l : Lsp.LocationLink) : Loc :=
  { uri := l.targetUri, line := l.targetSelectionRange.start.line, character := l.targetSelectionRange.start.character }

def mkOracle (c : Lsp.Client) (uri : String) (fm : FileMap) : Oracle :=
  let pos (off : Nat) : Lsp.Position := fm.utf8PosToLspPos ⟨off⟩
  { definition := fun off => do return (← c.definition uri (pos off)).map locOfLink
    expectedHead := fun off => do return (← c.expectedTypeHead uri (pos off)).map (·.map locOfLink)
    goalHead := fun off => do return (← c.goalHead uri (pos off)).map locOfLink
    diagExprHeads := fun s e => do
      return (← c.diagExprHeads uri ⟨pos s, pos e⟩).map (·.map locOfLink) }

/-- Repair one file through the server until it has no error and no silent
    projection. Returns whether the file changed. -/
partial def repairFile (c : Lsp.Client) (env : Environment) (rel : String) : M Bool := do
  let e ← read
  let uri := Lsp.uriOf e.root rel
  let text0 ← IO.FS.readFile (e.root / rel)
  c.open uri text0
  let rec loop (round : Nat) (text : String) (changed : Bool) : M Bool := do
    let diags ← c.waitIdle uri
    let errors := diags.filter (·.severity? == some .error)
    let parsed ← Parse.parseFile env rel text
    unless parsed.parseErrors.isEmpty do
      stop s!"{rel}: the file does not parse: {parsed.parseErrors[0]!}"
    let fm := parsed.fileMap
    let trace ← IO.mkRef #[]
    let ctx : Ctx := { env, parsed, preds := (← e.state.get).preds, predTexts := ← predTexts,
                       oracle := mkOracle c uri fm, trace }
    let mut chosen : Array (Repair × Nat × Nat) := #[]
    let mut taken : Array Edit := #[]
    let mut unmatched : Array Json := #[]
    -- the silent case first: it never conflicts with a diagnostic repair site
    for r in ← silentProjections ctx do
      chosen := chosen.push (r, r.edits[0]!.start, 0)
      taken := taken ++ r.edits
    let ranges := errors.map fun d =>
      let r := d.fullRange?.getD d.range
      ((fm.lspPosToUtf8Pos r.start).byteIdx, (fm.lspPosToUtf8Pos r.end).byteIdx, d)
    let sorted := ranges.qsort fun a b => a.1 > b.1 || (a.1 == b.1 && a.2.1 > b.2.1)
    for (s, en, d) in sorted do
      let lp := fm.utf8PosToLspPos ⟨s⟩
      match ← decide ctx s en with
      | none =>
        unmatched := unmatched.push (Json.mkObj [("line", toJson (lp.line + 1)), ("col", toJson lp.character), ("message", toJson d.message)])
      | some r =>
        if r.edits.any fun x => taken.any (conflicts x ·) then continue
        let key := s!"{rel}|{r.pattern}|{r.detail}|{(parsed.text.splitOn "\n")[lp.line]!}"
        if (← e.seen.get).contains key then
          stop s!"{rel}:{lp.line + 1}:{lp.character}: repair `{r.pattern}` would repeat on `{((parsed.text.splitOn "\n")[lp.line]!).trimAscii}`; the tool stops"
        chosen := chosen.push (r, s, 0)
        taken := taken ++ r.edits
    for line in ← trace.get do c.logLine s!"[decide] {rel} round {round}: {line}"
    modifyState fun st => { st with rounds := st.rounds.push (Json.mkObj [
      ("file", toJson rel), ("round", toJson round), ("errors", toJson errors.size),
      ("repairs", toJson chosen.size), ("unmatched", Json.arr unmatched)]) }
    if chosen.isEmpty then
      if errors.isEmpty then
        c.close uri
        return changed
      let u := unmatched[0]!
      stop s!"{rel}:{u.getObjValD "line"}:{u.getObjValD "col"}: no pattern matches:\n{(u.getObjValD "message").getStr?.toOption.getD ""}"
    let edits := chosen.flatMap fun (r, _, _) => r.edits
    let newText := applyEdits text edits
    checkForbidden text newText rel
    writeFile rel newText
    shiftPreds rel edits
    for (r, s, _) in chosen do
      let lp := fm.utf8PosToLspPos ⟨s⟩
      let key := s!"{rel}|{r.pattern}|{r.detail}|{(parsed.text.splitOn "\n")[lp.line]!}"
      e.seen.modify (·.insert key)
      let rec_ : RepairRecord :=
        { file := rel, line := lp.line + 1, col := lp.character, pattern := r.pattern, pred := r.pred, detail := r.detail }
      modifyState fun st => { st with repairs := st.repairs.push rec_ }
      log s!"[repair] {rel}:{lp.line + 1}:{lp.character} {r.pattern}: {r.detail}"
    saveState
    c.change uri newText
    loop (round + 1) newText true
  loop 1 text0 false

-- --- files and dependency order ---------------------------------------------

partial def leanFilesUnder (dir : System.FilePath) : IO (Array System.FilePath) := do
  let mut out := #[]
  for entry in ← dir.readDir do
    if ← entry.path.isDir then out := out ++ (← leanFilesUnder entry.path)
    else if entry.path.extension == some "lean" then out := out.push entry.path
  return out

/-- Every candidate file of the checkout, with its module name and imports. -/
def candidates (root : System.FilePath) : IO (Array (String × Name × Array Name)) := do
  let mut files : Array System.FilePath := #[root / "ArmadaLean.lean"]
  files := files ++ (← leanFilesUnder (root / "ArmadaLean"))
  files := files ++ (← leanFilesUnder (root / "docs" / "counterexamples"))
  let mut out := #[]
  for f in files do
    let rel := (f.toString.drop (root.toString.length + 1)).copy
    out := out.push (rel, moduleName rel, ← fileImports f)
  return out.qsort fun a b => a.1 < b.1

/-- The files that import a module of `roots` (transitively), roots first,
    in an order where every file comes after its imports. -/
def dependents (cands : Array (String × Name × Array Name)) (roots : Array Name) : Array String := Id.run do
  let byMod : Std.HashMap Name (String × Array Name) := cands.foldl (fun m (rel, n, imps) => m.insert n (rel, imps)) {}
  -- modules that (transitively) import a root, plus the roots
  let mut affected : Std.HashSet Name := Std.HashSet.ofArray roots
  let mut changed := true
  while changed do
    changed := false
    for (_, n, imps) in cands do
      if !affected.contains n && imps.any affected.contains then
        affected := affected.insert n
        changed := true
  -- topological order among the affected, ties by path
  let mut remaining := cands.filter fun (_, n, _) => affected.contains n
  let mut done : Std.HashSet Name := {}
  let mut order := #[]
  while !remaining.isEmpty do
    let ready := remaining.filter fun (_, _, imps) => imps.all fun i => !affected.contains i || done.contains i
    if ready.isEmpty then break
    for (rel, n, _) in ready do
      order := order.push rel
      done := done.insert n
    remaining := remaining.filter fun (_, n, _) => !done.contains n
  return order

-- --- the run -----------------------------------------------------------------

def stepBefore : M Unit := do
  let (ok, out) ← fullBuild "before"
  unless ok do stop ("the unchanged tree does not build:\n" ++ (tail out 3000))
  let (findings, summary) ← typeLint "before"
  (← read).lintBeforeFindings.set findings
  modifyState fun s => { s with lintBefore := summary }

def importModulesFor (root : System.FilePath) (files : Array String) : IO Environment := do
  let mut mods : Array Name := #[`ArmadaLean]
  for rel in files do
    for m in ← fileImports (root / rel) do
      unless mods.contains m do mods := mods.push m
  unsafe importTarget root mods

def stepDiscover (env : Environment) : M Unit := do
  let e ← read
  let names := e.args.names
  let resolved := resolveAll env (names.map String.toName)
  let mut founds : Array Found := #[]
  for i in [0:names.size] do
    let cands := resolved[i]!
    unless cands.size == 1 do
      stop s!"discovery: {names[i]!} does not resolve to exactly one ArmadaLean constant: {cands}"
    let n := cands[0]!
    let coreCtx : Core.Context := { fileName := "<discovery>", fileMap := default }
    let found ← match ← (EIO.toIO' ((describe env e.root n).toIO coreCtx { env })) with
      | .ok (f, _, _) => pure f
      | .error err => stop s!"discovery: {err}"
    if found.file.startsWith "ArmadaLean/Spec/" then stop s!"{n} is declared under ArmadaLean/Spec/; refused"
    let short := (n.toString.splitOn ".").getLast!
    let g ← git #["grep", "-n", s!"Spec\\.{short}\\b", "--", "ArmadaLean/Spec"]
    if g.exitCode == 0 && !g.stdout.trimAscii.isEmpty then
      stop s!"{n}: an approved statement names Spec.{short}; refused:\n{g.stdout}"
    founds := founds.push found
  -- conversion, per file
  let files := founds.foldl (fun acc f => if acc.contains f.file then acc else acc.push f.file) #[]
  let mut preds : Array Pred := #[]
  let mut reports : Array PredReport := #[]
  for rel in files do
    let text ← IO.FS.readFile (e.root / rel)
    let fm := FileMap.ofString text
    let inFile := founds.filter (·.file == rel)
    let mut edits : Array (Edit × Found × Nat × Nat) := #[]
    for f in inFile do
      if f.alreadyConverted then
        stop s!"{f.name} is already a converted structure with `.iff`; nothing to do"
      match conversionEdit f text with
      | .ok (ed, nameOff, holdsOff) => edits := edits.push (ed, f, nameOff, holdsOff)
      | .error msg => stop msg
    let newText := applyEdits text (edits.map (·.1))
    checkForbidden text newText rel
    writeFile rel newText
    -- offsets in the new text: shift by the edits that sit before
    for (ed, f, nameOff, holdsOff) in edits do
      let delta := edits.foldl (fun d (o, _, _, _) =>
        if o.stop ≤ ed.start && o != ed then d + o.replacement.utf8ByteSize - (o.stop - o.start) else d) 0
      let base := ed.start + delta
      let explicit := f.binders.filter (·.explicit) |>.toList.flatMap (·.names.toList)
      let pred : Pred :=
        { fullName := f.name.toString, declId := f.declId, file := rel,
          uri := Lsp.uriOf e.root rel, declIdOffset := base + nameOff, holdsOffset := base + holdsOff,
          bodyHead := f.bodyHead, bodyIsStructure := f.bodyIsStructure }
      preds := preds.push pred
      let rep : PredReport :=
        { name := f.name.toString, path := rel, line := (fm.toPosition ⟨f.defStart⟩).line,
          alreadyConverted := false, iff := " ".intercalate (f.declId :: explicit) ++ " ↔ …" }
      reports := reports.push rep
  modifyState fun s => { s with preds, predReports := reports }

def stepFiles (env : Environment) : M Unit := do
  let e ← read
  let s ← e.state.get
  let c ← Lsp.start e.root (e.logDir / "lsp.log")
  let mut i := s.nextIndex
  let files := s.files
  let result ← tryCatch (do
    let mut i := i
    while i < files.size do
      let rel := files[i]!
      log s!"[file] {rel} ({i + 1}/{files.size})"
      let changed ← repairFile c env rel
      if changed && rel.startsWith "ArmadaLean" then
        let m := moduleName rel
        let p ← runCmd #["lake", "build", m.toString] s!"build-{m}.log"
        if p.exitCode != 0 then
          stop s!"lake build {m} failed after the repairs:\n{(tail (p.stdout ++ p.stderr) 3000)}"
        let ilean := e.root / ".lake" / "build" / "lib" / "lean" / ((rel.dropEnd 5).copy ++ ".ilean")
        c.ileanChanged ilean
      i := i + 1
      modifyState fun st => { st with nextIndex := i }
      saveState
    pure (Except.ok ()))
    (fun err => pure (Except.error err))
  i := i
  c.stop
  match result with
  | .ok () => pure ()
  | .error err => throw err

def stepAfter : M Unit := do
  let e ← read
  let (ok, out) ← fullBuild "after"
  unless ok do stop ("the repaired tree does not build:\n" ++ (tail out 3000))
  let (findings, summary) ← typeLint "after"
  let pyArgs : IO.Process.SpawnArgs :=
    { cmd := "python3", args := #["scripts/update_gate_baseline.py", "--hashes-only"], cwd := e.root }
  let p ← IO.Process.output pyArgs
  let chk ← runCmd #["bash", "scripts/check.sh", "--json"] "check.log"
  let checkJson := match (chk.stdout.trimAscii.copy.splitOn "\n").getLast? >>= fun l => (Json.parse l).toOption with
    | some j => j
    | none => Json.mkObj [("error", toJson ((tail chk.stdout 2000) ++ (tail chk.stderr 2000)))]
  let okFields := match checkJson with
    | .obj kvs => Json.obj (kvs.foldl (init := {}) fun acc k v => if k.endsWith "_ok" then acc.insert k v else acc)
    | j => j
  let before := (← e.lintBeforeFindings.get).map findingKey
  let after := findings.map findingKey
  let bset := Std.HashSet.ofArray before
  let aset := Std.HashSet.ofArray after
  let removed := (before.filter (!aset.contains ·)).qsort (· < ·)
  let added := (after.filter (!bset.contains ·)).qsort (· < ·)
  let s ← e.state.get
  let report := Json.mkObj [
    ("checkout", toJson e.root.toString), ("head", toJson s.head), ("names", toJson s.names),
    ("stopped", Json.null), ("stops", toJson s.stops),
    ("predicates", toJson s.predReports), ("repairs", toJson s.repairs), ("rounds", Json.arr s.rounds),
    ("lint_before", s.lintBefore), ("lint_after", summary),
    ("findings_removed", toJson removed), ("findings_added", toJson added),
    ("gate_baseline_hashes", Json.mkObj [("exit", toJson p.exitCode.toNat), ("stderr", toJson (tail p.stderr 1000))]),
    ("check", okFields), ("check_full", checkJson),
    ("files_changed", toJson ((← git #["status", "--porcelain"]).stdout.splitOn "\n" |>.filter (· ≠ "") |>.map (·.drop 3 |>.copy))),
    ("timings", toJson s.timings)]
  IO.FS.writeFile e.args.report (report.pretty ++ "\n")

def writeStopReport (msg : String) : M Unit := do
  let e ← read
  let s ← e.state.get
  modifyState fun st => { st with stops := st.stops.push msg }
  saveState
  let report := Json.mkObj [
    ("checkout", toJson e.root.toString), ("head", toJson s.head), ("names", toJson s.names),
    ("stopped", toJson msg), ("stops", toJson s.stops), ("predicates", toJson s.predReports),
    ("repairs", toJson s.repairs), ("rounds", Json.arr s.rounds), ("lint_before", s.lintBefore),
    ("files_changed", toJson ((← git #["status", "--porcelain"]).stdout.splitOn "\n" |>.filter (· ≠ "") |>.map (·.drop 3 |>.copy))),
    ("timings", toJson s.timings)]
  IO.FS.writeFile e.args.report (report.pretty ++ "\n")

def run : M Unit := do
  let e ← read
  if !e.args.resume then
    stepBefore
  let cands ← candidates e.root
  if !e.args.resume then
    let env ← importModulesFor e.root (cands.map (·.1))
    stepDiscover env
    let roots := (← e.state.get).preds.map fun p => moduleName p.file
    let files := dependents cands roots
    modifyState fun s => { s with files, nextIndex := 0 }
    saveState
    stepFiles env
  else
    let env ← importModulesFor e.root (cands.map (·.1))
    stepFiles env
  stepAfter

end DefToStructure

open DefToStructure in
unsafe def main (argv : List String) : IO UInt32 := do
  let args ← match parseArgs argv with
    | .ok a => pure a
    | .error msg => IO.eprintln msg; return 2
  let root := (← IO.FS.realPath args.checkout)
  let logDir := root / ".scratch" / "codemod"
  IO.FS.createDirAll logDir
  let gitArgs : IO.Process.SpawnArgs := { cmd := "git", args := #["rev-parse", "HEAD"], cwd := root }
  let head := (← IO.Process.output gitArgs).stdout.trimAscii.copy
  let initial : State :=
    { head, names := args.names, preds := #[], predReports := #[], files := #[], nextIndex := 0,
      repairs := #[], rounds := #[], stops := #[], lintBefore := Json.null, timings := #[] }
  let state ← if args.resume then
      match Json.parse (← IO.FS.readFile (logDir / "state.json")) >>= fromJson? (α := State) with
      | .ok s => pure s
      | .error err => IO.eprintln s!"--resume: cannot read state.json: {err}"; return 2
    else pure initial
  let lintBefore ← if args.resume then
      let text ← IO.FS.readFile (logDir / "lint-before.jsonl")
      pure ((text.splitOn "\n").filter (·.startsWith "{")).toArray
    else pure #[]
  let stateRef ← IO.mkRef state
  let lintRef ← IO.mkRef lintBefore
  let seenRef ← IO.mkRef ({} : Std.HashSet String)
  let env : Env := { args, root, logDir, state := stateRef, lintBeforeFindings := lintRef, seen := seenRef }
  match ← (run.run env).run with
  | .ok () => return 0
  | .error ⟨msg⟩ =>
    IO.eprintln s!"STOPPED: {msg}"
    match ← ((writeStopReport msg).run env).run with
    | .ok () => pure ()
    | .error ⟨m2⟩ => IO.eprintln s!"report failed: {m2}"
    return 2
