/-
  An LSP client for `lake serve` (see ../README.md, "Elaboration").

  The client starts `lake serve` in the target checkout, speaks JSON-RPC over
  its stdin/stdout with `Lean.JsonRpc`, and keeps the latest
  `publishDiagnostics` and `$/lean/fileProgress` per document. `waitIdle`
  blocks until the server reports no processing range for a document; the
  diagnostics stored at that point are the final ones for that version.

  Requests used: `textDocument/definition`, `workspace/didChangeWatchedFiles`,
  `$/lean/rpc/connect`, `$/lean/rpc/call` with `Lean.Widget.getInteractiveGoals`,
  `Lean.Widget.getInteractiveTermGoal`, `Lean.Widget.getInteractiveDiagnostics`,
  `Lean.Widget.InteractiveDiagnostics.infoToInteractive` and
  `Lean.Widget.getGoToLocation`.
-/
import Lean.Data.Lsp
import Lean.Data.JsonRpc
import Lean.Server.Rpc.Basic
import DefToStructure.Proc

open Lean Lsp JsonRpc

namespace DefToStructure.Lsp

structure DocState where
  version : Nat := 0
  text : String := ""
  diagnostics : Array Diagnostic := #[]
  diagnosticsVersion : Option Int := none
  processing : Option (Array LeanFileProgressProcessingInfo) := none
  rpcSession : Option UInt64 := none

structure Client where
  child : IO.Process.Child { stdin := .piped, stdout := .piped, stderr := .inherit }
  stdin : IO.FS.Stream
  stdout : IO.FS.Stream
  root : System.FilePath
  nextId : IO.Ref Nat
  docs : IO.Ref (Std.HashMap DocumentUri DocState)
  log : IO.FS.Handle

def structuredToJson : Option Json.Structured → Json
  | some (.arr xs) => Json.arr xs
  | some (.obj kvs) => Json.obj kvs
  | none => Json.null

def uriOf (root : System.FilePath) (rel : String) : DocumentUri :=
  s!"file://{root / rel}"

def Client.logLine (c : Client) (s : String) : IO Unit := do
  c.log.putStrLn s
  c.log.flush

/-- Handle one server→client notification: diagnostics and progress. -/
def Client.handleNotification (c : Client) (n : Notification Json) : IO Unit := do
  match n.method with
  | "textDocument/publishDiagnostics" =>
    match fromJson? (α := PublishDiagnosticsParams) n.param with
    | .ok p =>
      c.docs.modify fun m =>
        let d := m.getD p.uri {}
        -- a late report of an older version is not this version's result
        if p.version?.any (· < Int.ofNat d.version) then m
        else m.insert p.uri { d with diagnostics := p.diagnostics, diagnosticsVersion := p.version? }
    | .error e => throw (IO.userError s!"bad publishDiagnostics: {e}")
  | "$/lean/fileProgress" =>
    match fromJson? (α := LeanFileProgressParams) n.param with
    | .ok p =>
      c.docs.modify fun m =>
        let d := m.getD p.textDocument.uri {}
        if p.textDocument.version?.any (· < d.version) then m
        else m.insert p.textDocument.uri { d with processing := some p.processing }
    | .error e => throw (IO.userError s!"bad fileProgress: {e}")
  | _ => pure ()

/-- Read messages until the response to `id` arrives. Notifications on the
    way update the document states. A server→client request gets an empty
    result. -/
partial def Client.awaitResponse (c : Client) (id : RequestID) : IO Json := do
  let msg ← c.stdout.readLspMessage
  match msg with
  | .response rid result =>
    if rid == id then return result
    c.logLine s!"unexpected response {toJson rid}"
    c.awaitResponse id
  | .responseError rid code message data? =>
    if rid == id then
      throw (IO.userError s!"request {toJson id} failed: {toJson code} {message} {data?}")
    c.awaitResponse id
  | .notification method param? =>
    c.handleNotification ⟨method, structuredToJson param?⟩
    c.awaitResponse id
  | .request rid _ _ =>
    c.stdin.writeLspMessage (.response rid Json.null)
    c.awaitResponse id

def Client.request (c : Client) (method : String) (params : Json) : IO Json := do
  let n ← c.nextId.modifyGet fun n => (n, n + 1)
  let id := RequestID.num n
  c.stdin.writeLspMessage (.request id method (Json.toStructured? params).toOption)
  c.awaitResponse id

def Client.notify (c : Client) (method : String) (params : Json) : IO Unit :=
  c.stdin.writeLspMessage (.notification method (Json.toStructured? params).toOption)

def start (root : System.FilePath) (logPath : System.FilePath) : IO Client := do
  let child ← IO.Process.spawn {
    cmd := "lake", args := #["serve"], cwd := root, env := ← leanEnv,
    stdin := .piped, stdout := .piped, stderr := .inherit }
  let log ← IO.FS.Handle.mk logPath .write
  let c : Client := {
    child, stdin := IO.FS.Stream.ofHandle child.stdin, stdout := IO.FS.Stream.ofHandle child.stdout,
    root, nextId := ← IO.mkRef 1, docs := ← IO.mkRef {}, log }
  let caps : ClientCapabilities := {}
  let params : InitializeParams := {
    processId? := none, rootUri? := some (uriOf root ""), capabilities := caps,
    initializationOptions? := some { hasWidgets? := some true, logCfg? := none } }
  let _ ← c.request "initialize" (toJson params)
  c.notify "initialized" (Json.mkObj [])
  return c

def Client.stop (c : Client) : IO Unit := do
  let _ ← c.request "shutdown" Json.null
  c.notify "exit" Json.null
  let _ ← c.child.wait
  c.log.flush

def Client.doc (c : Client) (uri : DocumentUri) : IO DocState :=
  return (← c.docs.get).getD uri {}

def Client.open (c : Client) (uri : DocumentUri) (text : String) : IO Unit := do
  c.docs.modify fun m => m.insert uri { version := 1, text, processing := none }
  let item : TextDocumentItem := { uri, languageId := "lean4", version := 1, text }
  c.notify "textDocument/didOpen" (toJson ({ textDocument := item } : DidOpenTextDocumentParams))

def Client.close (c : Client) (uri : DocumentUri) : IO Unit := do
  c.notify "textDocument/didClose" (toJson ({ textDocument := { uri } } : DidCloseTextDocumentParams))
  c.docs.modify fun m => m.erase uri

/-- Tell the watchdog that a module's `.ilean` changed. `textDocument/definition`
    results pass through the watchdog, which moves each target to the
    position its reference data holds (`Watchdog.bendLocationLinks`). That
    data comes from the `.ilean` files loaded at start and from the open
    files' workers; a module rebuilt after its worker closed keeps its old
    positions until this notification (`handleDidChangeWatchedFiles`). -/
def Client.ileanChanged (c : Client) (ileanPath : System.FilePath) : IO Unit :=
  c.notify "workspace/didChangeWatchedFiles" (Json.mkObj [("changes", Json.arr #[
    Json.mkObj [("uri", toJson s!"file://{ileanPath}"), ("type", toJson (2 : Nat))]])])

/-- Replace the whole text. The server re-elaborates from the first changed position. -/
def Client.change (c : Client) (uri : DocumentUri) (text : String) : IO Unit := do
  let d ← c.doc uri
  let version := d.version + 1
  c.docs.modify fun m => m.insert uri { d with version, text, processing := none, diagnostics := #[], rpcSession := none }
  let vid : VersionedTextDocumentIdentifier := { uri, version? := some version }
  let params : DidChangeTextDocumentParams := { textDocument := vid, contentChanges := #[.fullChange text] }
  c.notify "textDocument/didChange" (toJson params)

def tdp (uri : DocumentUri) (pos : Lsp.Position) : Json :=
  toJson ({ textDocument := { uri }, position := pos } : TextDocumentPositionParams)

/-- Block until the server has processed the whole document at its current
    version. Then one round trip (`shutdown`-free: a `$/lean/plainGoal` at the
    origin) orders the final diagnostics before the return. -/
partial def Client.waitIdle (c : Client) (uri : DocumentUri) : IO (Array Diagnostic) := do
  let d ← c.doc uri
  let done := match d.processing with
    | some ps => ps.isEmpty && d.diagnosticsVersion == some (Int.ofNat d.version)
    | none => false
  if done then
    let _ ← c.request "$/lean/plainGoal" (tdp uri ⟨0, 0⟩)
    return (← c.doc uri).diagnostics
  let msg ← c.stdout.readLspMessage
  match msg with
  | .notification method param? => c.handleNotification ⟨method, structuredToJson param?⟩
  | .request rid _ _ => c.stdin.writeLspMessage (.response rid Json.null)
  | .response rid _ => c.logLine s!"stray response {toJson rid}"
  | .responseError rid _ m _ => c.logLine s!"stray error response {toJson rid}: {m}"
  c.waitIdle uri

def Client.definition (c : Client) (uri : DocumentUri) (pos : Lsp.Position) : IO (Array LocationLink) := do
  let r ← c.request "textDocument/definition" (tdp uri pos)
  match fromJson? r with
  | .ok ls => return ls
  | .error e => throw (IO.userError s!"definition: {e}: {r}")

def Client.rpcSession (c : Client) (uri : DocumentUri) : IO UInt64 := do
  let d ← c.doc uri
  if let some s := d.rpcSession then return s
  let r ← c.request "$/lean/rpc/connect" (toJson ({ uri } : RpcConnectParams))
  match fromJson? (α := RpcConnected) r with
  | .ok ⟨s⟩ =>
    c.docs.modify fun m => m.insert uri { (m.getD uri {}) with rpcSession := some s }
    return s
  | .error e => throw (IO.userError s!"rpc/connect: {e}")

/-- One RPC call. A session expires after 30 s without a message
    (`RpcSession.keepAliveTimeMs`); `change` drops the session, and a call
    that still finds it expired opens a new one and repeats once. -/
def Client.rpcCall (c : Client) (uri : DocumentUri) (pos : Lsp.Position) (method : Name) (params : Json) : IO Json := do
  let call : IO Json := do
    let sessionId ← c.rpcSession uri
    c.notify "$/lean/rpc/keepAlive" (toJson ({ uri, sessionId } : RpcKeepAliveParams))
    c.request "$/lean/rpc/call" (toJson ({ textDocument := { uri }, position := pos, sessionId, method, params } : RpcCallParams))
  try call
  catch e =>
    let expired := ((toString e).splitOn "Outdated RPC session").length > 1
    unless expired do throw e
    c.docs.modify fun m => m.insert uri { (m.getD uri {}) with rpcSession := none }
    call

/-- The RPC reference (`{"p": n}`) in a JSON value, when it is one. -/
def rpcRef? (j : Json) : Option Json :=
  match j.getObjVal? "p" with
  | .ok _ => some j
  | .error _ => none

/-- Walk a `TaggedText SubexprInfo` JSON value; `f` sees each tag's
    `SubexprInfo` JSON. -/
partial def foldTagged (j : Json) (acc : Array Json) : Array Json :=
  match j.getObjVal? "text" with
  | .ok _ => acc
  | .error _ =>
    match j.getObjVal? "append" with
    | .ok (.arr xs) => xs.foldl (fun a x => foldTagged x a) acc
    | _ =>
      match j.getObjVal? "tag" with
      | .ok (.arr #[info, inner]) => foldTagged inner (acc.push info)
      | _ => acc

/-- The tag whose `subexprPos` is the longest all-`appFn` path (`/0/0/…`):
    the head of an application in a `CodeWithInfos`. -/
def headTag? (code : Json) : Option Json := Id.run do
  let tags := foldTagged code #[]
  let mut best : Option (Nat × Json) := none
  for t in tags do
    let .ok (.str p) := t.getObjVal? "subexprPos" | continue
    let coords := (p.splitOn "/").filter (· ≠ "")
    if coords.all (· == "0") then
      let depth := coords.length
      if best.all (fun (d, _) => depth > d) then best := some (depth, t)
  return best.map (·.2)

/-- The root tag of a `CodeWithInfos` (`subexprPos` `/`). -/
def rootTag? (code : Json) : Option Json :=
  (foldTagged code #[]).find? fun t => (t.getObjVal? "subexprPos").toOption == some (Json.str "/")

/-- Definition location of an RPC `InfoWithCtx` reference. -/
def Client.gotoRef (c : Client) (uri : DocumentUri) (pos : Lsp.Position) (info : Json) : IO (Array LocationLink) := do
  let r ← c.rpcCall uri pos `Lean.Widget.getGoToLocation
    (Json.mkObj [("kind", "definition"), ("info", info)])
  match fromJson? r with
  | .ok ls => return ls
  | .error e => throw (IO.userError s!"getGoToLocation: {e}: {r}")

/-- The head constant's definition of the expression a `CodeWithInfos`
    prints. The printed form may use field notation (`s.IsDAG`), which tags
    no head; so the root's `InfoWithCtx` goes through
    `InteractiveDiagnostics.infoToInteractive`, whose `exprExplicit` prints
    the expression with `pp.explicit` and `pp.tagAppFns` (`@IsDAG E inst s`,
    the head tagged at `/0/0/0`). -/
def Client.headLoc (c : Client) (uri : DocumentUri) (pos : Lsp.Position) (code : Json) : IO (Array LocationLink) := do
  let some root := rootTag? code | return #[]
  let .ok info := root.getObjVal? "info" | return #[]
  let popup ← c.rpcCall uri pos `Lean.Widget.InteractiveDiagnostics.infoToInteractive info
  let .ok explicit := popup.getObjVal? "exprExplicit" | return #[]
  let some head := headTag? explicit >>= (·.getObjVal? "info" |>.toOption) | return #[]
  c.gotoRef uri pos head

/-- The head constant's definition of the expected type of the term at `pos`;
    `none` when the server has no term info there. -/
def Client.expectedTypeHead (c : Client) (uri : DocumentUri) (pos : Lsp.Position) : IO (Option (Array LocationLink)) := do
  let r ← c.rpcCall uri pos `Lean.Widget.getInteractiveTermGoal (tdp uri pos)
  if r.isNull then return none
  let .ok ty := r.getObjVal? "type" | throw (IO.userError s!"getInteractiveTermGoal: no type: {r}")
  return some (← c.headLoc uri pos ty)

/-- The head constant's definition of the main goal's target at `pos`
    (the goal before the tactic at `pos`). -/
def Client.goalHead (c : Client) (uri : DocumentUri) (pos : Lsp.Position) : IO (Array LocationLink) := do
  let r ← c.rpcCall uri pos `Lean.Widget.getInteractiveGoals (tdp uri pos)
  if r.isNull then return #[]
  let .ok (.arr goals) := r.getObjVal? "goals" | throw (IO.userError s!"getInteractiveGoals: {r}")
  let some g := goals[0]? | return #[]
  let .ok ty := g.getObjVal? "type" | return #[]
  c.headLoc uri pos ty

/-- Every expression embed of a `TaggedText MsgEmbed` JSON value, in order. -/
partial def exprEmbeds (j : Json) (acc : Array Json) : Array Json :=
  match j.getObjVal? "text" with
  | .ok _ => acc
  | .error _ =>
    match j.getObjVal? "append" with
    | .ok (.arr xs) => xs.foldl (fun a x => exprEmbeds x a) acc
    | _ =>
      match j.getObjVal? "tag" with
      | .ok (.arr #[embed, inner]) =>
        -- plain text pieces come as `expr` embeds too; an expression has tags
        let acc := match embed.getObjVal? "expr" with
          | .ok code => if (foldTagged code #[]).isEmpty then acc else acc.push code
          | .error _ => acc
        exprEmbeds inner acc
      | _ => acc

/-- The head-constant definitions of each expression Lean embedded in the
    diagnostic at `range` (one array per embedded expression, in order). -/
def Client.diagExprHeads (c : Client) (uri : DocumentUri) (range : Range) : IO (Array (Array LocationLink)) := do
  let params := Json.mkObj [("lineRange", Json.mkObj [("start", toJson range.start.line), ("end", toJson (range.start.line + 1))])]
  let r ← c.rpcCall uri range.start `Lean.Widget.getInteractiveDiagnostics params
  let .arr diags := r | throw (IO.userError s!"getInteractiveDiagnostics: {r}")
  let same (d : Json) : Bool :=
    match d.getObjVal? "fullRange" <|> d.getObjVal? "range" with
    | .ok rj => (fromJson? (α := Range) rj).toOption == some range
    | .error _ => false
  let some d := diags.find? same | return #[]
  let .ok msg := d.getObjVal? "message" | return #[]
  let mut out := #[]
  for code in exprEmbeds msg #[] do
    out := out.push (← c.headLoc uri range.start code)
  return out

end DefToStructure.Lsp
