/-
  Parsing of a target file into command `Syntax` (see ../README.md, "Syntax").

  `parseFile` runs the Lean parser over the file with an environment that
  holds the parsers of the file's imports. It elaborates only the commands
  that change how later text parses (`open`, `namespace`, `section`, `end`,
  `set_option`, `universe`, `variable`, `macro`, `syntax`, `notation`, …). It
  skips every declaration: the server elaborates those. The result is the array of
  command syntax trees with source positions, and the `FileMap`.

  The rest of the module answers questions about syntax at a source range:
  the innermost node that covers a range, the tactic that contains a
  position, the pieces of an `unfold` or `simp` call, every numeric
  projection in the file.
-/
import Lean

open Lean Elab Parser

namespace DefToStructure.Parse

/-- Command kinds the parse step elaborates: they change scopes, options or
    the parser itself. Everything else only needs to parse. -/
def scopeKinds : List Name := [
  ``Command.«open», ``Command.«namespace», ``Command.«section», ``Command.«end»,
  ``Command.«set_option», ``Command.«universe», ``Command.«macro», ``Command.«macro_rules»,
  ``Command.«syntax», ``Command.syntaxAbbrev, ``Command.«notation», ``Command.«mixfix»,
  ``Command.syntaxCat, ``Command.«attribute», ``Command.«variable»]

structure Parsed where
  commands : Array Syntax
  /-- The section's `variable` binders in scope before each command. -/
  scopeVars : Array (Array Syntax)
  fileMap : FileMap
  text : String
  /-- Messages the parser logged (a parse error stops the tool). -/
  parseErrors : Array String

/-- Parse `text` as a module body with `env` (the imports already loaded). -/
partial def parseFile (env : Environment) (fileName : String) (text : String) : IO Parsed := do
  let inputCtx := Parser.mkInputContext text fileName
  let (_, parserState, messages) ← Parser.parseHeader inputCtx
  let cmdState := Command.mkState env messages {}
  let frontendState : Frontend.State := { commandState := cmdState, parserState, cmdPos := parserState.pos }
  let rec loop (s : Frontend.State) (acc : Array Syntax) (vars : Array (Array Syntax)) (errs : Array String) :
      IO (Frontend.State × Array Syntax × Array (Array Syntax) × Array String) := do
    let scope := s.commandState.scopes.head!
    let pmctx : Parser.ParserModuleContext :=
      { env := s.commandState.env, options := scope.opts,
        currNamespace := scope.currNamespace, openDecls := scope.openDecls }
    let before := s.commandState.messages.toList.length
    let (cmd, ps, msgs) := Parser.parseCommand inputCtx pmctx s.parserState s.commandState.messages
    let mut errs := errs
    for m in msgs.toList.drop before do
      if m.severity == .error then
        errs := errs.push s!"{m.pos.line}:{m.pos.column}: {← m.data.toString}"
    let s : Frontend.State :=
      { s with parserState := ps, cmdPos := s.parserState.pos,
               commandState := { s.commandState with messages := msgs } }
    let acc := acc.push cmd
    let vars := vars.push scope.varDecls
    let s ← if scopeKinds.contains cmd.getKind then
      let cmdCtx : Command.Context :=
        { cmdPos := s.cmdPos, fileName := inputCtx.fileName, fileMap := inputCtx.fileMap,
          snap? := none, cancelTk? := none }
      match ← (EIO.toIO' ((Command.elabCommandTopLevel cmd #[] cmdCtx).run s.commandState)) with
      | .ok (_, st) => pure { s with commandState := st }
      | .error _ => pure s
    else pure s
    if Parser.isTerminalCommand cmd then return (s, acc, vars, errs) else loop s acc vars errs
  let (_, cmds, vars, errs) ← loop frontendState #[] #[] #[]
  return { commands := cmds, scopeVars := vars, fileMap := inputCtx.fileMap, text, parseErrors := errs }

/-- Byte range of a syntax node. -/
def range? (stx : Syntax) : Option (String.Pos.Raw × String.Pos.Raw) := do
  let s ← stx.getPos? (canonicalOnly := true)
  let e ← stx.getTailPos? (canonicalOnly := true)
  pure (s, e)

/-- The stack of nodes from the command root down to the innermost node whose
    range covers `[s, e]`. Innermost last. Atoms count. -/
partial def stackAt (cmds : Array Syntax) (s e : String.Pos.Raw) : Array Syntax := Id.run do
  let covers (stx : Syntax) : Bool :=
    match range? stx with
    | some (a, b) => a ≤ s && e ≤ b
    | none => false
  let rec go (stx : Syntax) (acc : Array Syntax) : Array Syntax :=
    let acc := acc.push stx
    match stx.getArgs.find? covers with
    | some child => go child acc
    | none => acc
  match cmds.find? covers with
  | some cmd => go cmd #[]
  | none => #[]

/-- The syntax categories a kind belongs to, from the environment's parser tables. -/
def kindInCategory (env : Environment) (cat : Name) (kind : Name) : Bool :=
  match (parserExtension.getState env).categories.find? cat with
  | some c => c.kinds.contains kind
  | none => false

def isTacticKind (env : Environment) (k : Name) : Bool :=
  kindInCategory env `tactic k

def isTermKind (env : Environment) (k : Name) : Bool :=
  k == identKind || kindInCategory env `term k

/-- Innermost node of `stack` (from the end) that satisfies `p`. -/
def innermost (stack : Array Syntax) (p : Syntax → Bool) : Option Syntax :=
  stack.reverse.find? p

/-- The tactic that contains the innermost node: the last tactic-kind node of
    the stack, with its index, so that the caller can see the nodes below it. -/
def innermostTactic (env : Environment) (stack : Array Syntax) : Option (Nat × Syntax) := Id.run do
  let mut i := stack.size
  while i > 0 do
    i := i - 1
    if isTacticKind env stack[i]!.getKind then return some (i, stack[i]!)
  return none

/-- Every identifier node under `stx`, in source order. -/
partial def idents (stx : Syntax) (acc : Array Syntax := #[]) : Array Syntax :=
  match stx with
  | .ident .. => acc.push stx
  | .node _ _ args => args.foldl (fun a x => idents x a) acc
  | _ => acc

/-- Every `t.k` (`Term.proj` with a `fieldIdx`) under `stx`: the proj node and
    the field index token. -/
partial def numericProjections (stx : Syntax) (acc : Array (Syntax × Syntax) := #[]) : Array (Syntax × Syntax) :=
  match stx with
  | .node _ k args =>
    let acc := if k == ``Term.proj && args.size == 3 && args[2]!.isOfKind fieldIdxKind then acc.push (stx, args[2]!) else acc
    args.foldl (fun a x => numericProjections x a) acc
  | _ => acc

/-- The byte range of component `i` (0-based) of a dotted identifier token,
    from the token's own text. `h.isDAG.irrefl` → components `h`, `isDAG`,
    `irrefl`. -/
def identComponentRanges (text : String) (id : Syntax) : Array (String.Pos.Raw × String.Pos.Raw) := Id.run do
  let some (s, e) := range? id | return #[]
  let raw := String.Pos.Raw.extract text s e
  let mut out := #[]
  let mut start := s
  let mut i := s
  let mut depth := 0
  for c in raw.toList do
    if c == '«' then depth := depth + 1
    if c == '»' then depth := depth - 1
    if c == '.' && depth == 0 then
      out := out.push (start, i)
      start := ⟨i.byteIdx + 1⟩
    i := ⟨i.byteIdx + String.utf8ByteSize (String.singleton c)⟩
  out := out.push (start, e)
  return out

/-- Start of the line that contains `p`. -/
def lineStart (fm : FileMap) (p : String.Pos.Raw) : String.Pos.Raw :=
  fm.ofPosition ⟨(fm.toPosition p).line, 0⟩

end DefToStructure.Parse
