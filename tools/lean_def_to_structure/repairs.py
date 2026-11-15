"""Repair patterns for `tools/lean_def_to_structure` (see README.md).

`match` takes one Lean error message and the source of the file. It
returns a `Repair` when the message matches one fixed pattern, else
`None`. The driver (`codemod.py`) applies the repairs and runs Lean
again. No pattern inserts `sorry`, `admit`, `axiom` or `native_decide`.

Patterns (the `Repair.pattern` names):

- `holds-apply`: a hypothesis of the converted type is applied as a
  function (`h x`). Lean: "Function expected at h". Edit: `h.holds x`.
- `holds-field`: a field is projected from a hypothesis of the converted
  type (`h.irrefl`, `h.2.1`). Lean: "Invalid field" or "Invalid
  projection". Edit: `h.holds.irrefl`, `h.holds.2.1`.
- `holds-type`: a term of the converted type stands where its body is
  expected. Lean: "Type mismatch" / "Application type mismatch" with the
  term's type headed by the predicate. Edit: `term.holds`.
- `holds-proj`: `h.1` on a one-field structure is the whole body. The
  probe (`probe/Main.lean`) reports every such `.1` from the info trees;
  the driver applies the edit. The matcher also covers the case Lean
  reports at once: it prints the term as `h.holds`. Edit: `h.holds.1`.
- `wrap`: a term of the body's type stands where the converted type is
  expected. Lean: "Type mismatch" / "Application type mismatch" with the
  expected type headed by the predicate. Edit: `⟨term⟩`.
- `wrap-nested`: the sole field of the converted structure gets a term
  whose expected type is the body, and the body is itself a structure
  (`IsDAG`'s `Acyclic`). Lean: "Application type mismatch … in the
  application { holds := … }". Edit: `⟨term⟩`.
- `constructor`: a tactic failed on a goal headed by the predicate. Edit:
  `constructor` before the tactic.
- `unfold-iff`: `unfold P` cannot unfold a structure. Edit: `rw [P.iff]`,
  then `unfold` of the remaining names, if any.
- `simp-iff`: a `simp` family call lists the predicate and makes no
  progress. Edit: `P.iff` replaces `P` in every lemma list of the call.
"""

from __future__ import annotations

import re
from dataclasses import dataclass, field


@dataclass(frozen=True)
class Predicate:
    """One converted predicate."""

    full_name: str
    """`ArmadaLean.Nitro.Adjudicator.IsDAG`."""
    decl_id: str
    """The name as the `def` wrote it: `Adjudicator.IsDAG`."""
    body_head: str | None
    """Head constant of the body (`ArmadaLean.Nitro.Acyclic`), or None."""
    body_is_structure: bool
    """The body's head unfolds to a structure (`Acyclic` → `Std.Irrefl`)."""

    @property
    def short(self) -> str:
        return self.full_name.rsplit(".", 1)[-1]


@dataclass(frozen=True)
class Message:
    """One Lean message with its range (lines 1-based, columns in code points)."""

    file: str
    line: int
    col: int
    end_line: int | None
    end_col: int | None
    severity: str
    text: str


@dataclass(frozen=True)
class Edit:
    start: int
    end: int
    replacement: str


@dataclass(frozen=True)
class Repair:
    file: str
    line: int
    col: int
    pattern: str
    predicate: str | None
    detail: str
    edits: tuple[Edit, ...] = field(default_factory=tuple)


LEAD_RE = re.compile(r"[^\W\d][\w'!?]*")
"""The first component of a Lean identifier."""
STEP_RE = re.compile(r"\.(\d+|[^\W\d][\w'!?]*)")
"""One projection or namespace step after a component: `.1`, `.holds`."""

SIMP_FAMILY = ("simp?", "simp_all?", "simpa?", "simp", "simp_all", "simpa", "dsimp")
TERM_TACTICS = ("exact", "refine'", "refine", "apply")
"""Tactics whose error range covers the keyword and the term."""
GOAL_TACTICS = frozenset((
    "intro", "intros", "rintro", "rw", "rewrite", "rwa", "exact", "refine", "refine'",
    "apply", "simp", "simp_all", "simpa", "obtain", "rcases", "cases", "induction",
    "omega", "decide", "trivial", "exfalso", "by_contra", "push_neg", "specialize",
    "show", "change", "ext", "funext", "congr", "use",
))
"""Tactics `constructor` may precede when they fail on a goal headed by the predicate."""
TACTIC_SEPARATOR_RE = re.compile(r"·|;|=>|<;>|\(|\bby\b|\bthen\b|\belse\b")
NESTED_HOLDS_RE = re.compile(r"\nin the application\n\s*\{\s*holds :=")
"""The mismatch sits inside the converted structure's own constructor (Lean
prints the instance on one line or several)."""


class Source:
    """A file's text with line/column arithmetic (columns in code points)."""

    def __init__(self, text: str) -> None:
        self.text = text
        self.line_starts = [0]
        for i, ch in enumerate(text):
            if ch == "\n":
                self.line_starts.append(i + 1)

    def offset(self, line: int, col: int) -> int:
        return self.line_starts[line - 1] + col

    def line_text(self, line: int) -> str:
        start = self.line_starts[line - 1]
        end = self.text.find("\n", start)
        return self.text[start:] if end < 0 else self.text[start:end]

    def line_start(self, off: int) -> int:
        return self.text.rfind("\n", 0, off) + 1

    def group_end(self, off: int) -> int | None:
        """End offset (exclusive) of the bracket group that opens at `off`."""
        pairs = {"(": ")", "[": "]", "⟨": "⟩", "{": "}"}
        if off >= len(self.text) or self.text[off] not in pairs:
            return None
        stack: list[str] = []
        i = off
        while i < len(self.text):
            ch = self.text[i]
            if ch in pairs:
                stack.append(pairs[ch])
            elif stack and ch == stack[-1]:
                stack.pop()
                if not stack:
                    return i + 1
            i += 1
        return None

    def chain_end(self, off: int, printed: str) -> int | None:
        """End of the source term at `off` that Lean printed as `printed`:
        an identifier or a bracket group, then as many `.step`s as the
        printed term has (`h.2` prints as `h.right`)."""
        p_lead = LEAD_RE.match(printed.strip())
        if self.text[off] in "([⟨{":
            end = self.group_end(off)
            if end is None:
                return None
        else:
            s_lead = LEAD_RE.match(self.text, off)
            if s_lead is None or p_lead is None or s_lead.group(0) != p_lead.group(0):
                return None
            end = s_lead.end()
        steps = len(STEP_RE.findall(printed.strip()[p_lead.end():])) if p_lead else 0
        for _ in range(steps):
            m = STEP_RE.match(self.text, end)
            if m is None:
                return None
            end = m.end()
        return end


# --- message text helpers ---------------------------------------------------


def head_token(type_text: str) -> str:
    """The head of a printed type: `s.IsDAG` → `s.IsDAG`, `IsDAG s` → `IsDAG`,
    `(turboT p).TransfersOnly` → `(…).TransfersOnly`."""
    t = type_text.strip()
    if t.startswith("("):
        depth = 0
        for i, ch in enumerate(t):
            if ch in "([⟨{":
                depth += 1
            elif ch in ")]⟩}":
                depth -= 1
                if depth == 0:
                    m = re.match(r"\.[^\s()]+", t[i + 1:])
                    return "(…)" + m.group(0) if m else ""
        return ""
    m = re.match(r"\S+", t)
    return m.group(0) if m else ""


def head_matches(type_text: str, pred: Predicate) -> bool:
    """`type_text` is an application of `pred`, printed with the name
    (`Adjudicator.IsDAG s`), a suffix of it (`IsDAG s`), or field notation
    (`s.IsDAG`)."""
    tok = head_token(type_text)
    if not tok:
        return False
    comps = tok.split(".")
    if comps[-1] != pred.short:
        return False
    full = pred.full_name.split(".")
    if full[-len(comps):] == comps:
        return True
    # field notation: a receiver, then the short name
    return len(comps) >= 2


def _section(text: str, start_marker: str, end_markers: tuple[str, ...]) -> str | None:
    """Text between `start_marker` and the first of `end_markers`."""
    i = text.find(start_marker)
    if i < 0:
        return None
    i += len(start_marker)
    j = len(text)
    for m in end_markers:
        k = text.find(m, i)
        if 0 <= k < j:
            j = k
    return text[i:j].strip()


def mismatch_parts(text: str) -> tuple[str, str, str] | None:
    """`(term, has_type, expected_type)` of a type-mismatch message."""
    if text.startswith("Type mismatch\n"):
        term = _section(text, "Type mismatch\n", ("\nhas type\n",))
    elif text.startswith("Application type mismatch: The argument\n"):
        term = _section(text, "The argument\n", ("\nhas type\n",))
    else:
        return None
    has = _section(text, "\nhas type\n", ("\nbut is expected to have type\n",))
    exp = _section(text, "\nbut is expected to have type\n",
                   ("\nin the application\n", "\n\nNote:", "\n\nHint:"))
    if term is None or has is None or exp is None:
        return None
    return term, has, exp


def goal_target(text: str) -> str | None:
    """The target of the last goal printed in a tactic error."""
    i = text.rfind("\n⊢ ")
    if i < 0:
        return None
    rest = text[i + 3:]
    j = rest.find("\n\n")
    return (rest if j < 0 else rest[:j]).strip()


def lead_ident(s: str) -> str:
    m = LEAD_RE.match(s.strip())
    return m.group(0) if m else ""


# --- source helpers ----------------------------------------------------------


def _find_pred(preds: list[Predicate], type_text: str) -> Predicate | None:
    for p in preds:
        if head_matches(type_text, p):
            return p
    return None


def _pred_by_ident(preds: list[Predicate], ident: str) -> Predicate | None:
    """The predicate a source identifier names (`Adjudicator.IsDAG`, `IsDAG`)."""
    comps = ident.split(".")
    for p in preds:
        full = p.full_name.split(".")
        if full[-len(comps):] == comps:
            return p
    return None


def _term_range(src: Source, msg: Message) -> tuple[int, int] | None:
    """The source range of the term a mismatch message is about. For
    `exact t` / `refine t` Lean reports the whole tactic; strip the keyword."""
    if msg.end_line is None or msg.end_col is None:
        return None
    start = src.offset(msg.line, msg.col)
    end = src.offset(msg.end_line, msg.end_col)
    if end <= start:
        return None
    text = src.text[start:end]
    for kw in TERM_TACTICS:
        if re.match(re.escape(kw) + r"\s", text):
            rest = text[len(kw):]
            start += len(kw) + (len(rest) - len(rest.lstrip()))
            break
    return start, end


def _tactic_start(src: Source, off: int) -> int:
    """Start of the tactic that contains `off`, on the same line: the first
    token after the nearest of line start, `·`, `;`, `=>`, `<;>`, `(`, `by`."""
    ls = src.line_start(off)
    seg = src.text[ls:off]
    best = 0
    for m in TACTIC_SEPARATOR_RE.finditer(seg):
        best = m.end()
    rest = seg[best:]
    return ls + best + (len(rest) - len(rest.lstrip()))


def _starts_line(src: Source, off: int) -> bool:
    return src.text[src.line_start(off):off].strip() == ""


def _indent_of(src: Source, off: int) -> str:
    seg = src.text[src.line_start(off):off]
    return seg[: len(seg) - len(seg.lstrip())]


def _tactic_end(src: Source, start: int) -> int:
    """End of a one-line tactic: `;`, `<;>`, a closing bracket at depth 0,
    or the end of the line."""
    i = start
    depth = 0
    n = len(src.text)
    while i < n:
        ch = src.text[i]
        if ch == "\n":
            break
        if ch in "([⟨{":
            depth += 1
        elif ch in ")]⟩}":
            if depth == 0:
                break
            depth -= 1
        elif depth == 0 and (ch == ";" or src.text.startswith("<;>", i)):
            break
        i += 1
    return i


def _insert_constructor(src: Source, tactic_start: int) -> Edit:
    if _starts_line(src, tactic_start):
        return Edit(tactic_start, tactic_start, "constructor\n" + _indent_of(src, tactic_start))
    return Edit(tactic_start, tactic_start, "constructor; ")


def _is_application(term_src: str) -> bool:
    """`term_src` has whitespace outside brackets (an application or a `fun`)."""
    depth = 0
    for ch in term_src:
        if ch in "([⟨{":
            depth += 1
        elif ch in ")]⟩}":
            depth -= 1
        elif ch.isspace() and depth == 0:
            return True
    return False


# --- the matcher -----------------------------------------------------------


def match(msg: Message, src: Source, preds: list[Predicate]) -> Repair | None:
    """Match one error message against the fixed patterns."""
    if msg.severity != "error":
        return None
    text = msg.text
    pos = src.offset(msg.line, msg.col)

    def repair(pattern: str, pred: Predicate | None, detail: str, *edits: Edit) -> Repair:
        return Repair(msg.file, msg.line, msg.col, pattern,
                      pred.full_name if pred else None, detail, tuple(edits))

    # holds-apply: "Function expected at\n  h\nbut this term has type\n  P args"
    if text.startswith("Function expected at\n"):
        term = _section(text, "Function expected at\n", ("\nbut this term has type\n",))
        has = _section(text, "\nbut this term has type\n", ("\n\nNote:", "\n\n"))
        if term is None or has is None:
            return None
        pred = _find_pred(preds, has)
        if pred is None:
            return None
        end = src.chain_end(pos, term)
        if end is None:
            return None
        return repair("holds-apply", pred, f"{term} applied as a function", Edit(end, end, ".holds"))

    # holds-field: "Invalid field `f`: The environment does not contain `P.f` …"
    m = re.match(r"Invalid field `([^`]+)`: The environment does not contain `([^`]+)`", text)
    if m:
        fld, missing = m.group(1), m.group(2)
        for pred in preds:
            if missing == f"{pred.full_name}.{fld}":
                if src.text.startswith(fld, pos) and pos > 0 and src.text[pos - 1] == ".":
                    return repair("holds-field", pred, f"field {fld} through holds",
                                  Edit(pos, pos, "holds."))
        return None

    # holds-field: "Invalid projection: Index `k` is invalid … The expression\n  X\nhas type `T` which has only 1 field"
    if text.startswith("Invalid projection: Index"):
        m = re.search(r"The expression\n  (.+?)\nhas type `([^`]+)` which has only 1 field", text, re.S)
        if not m:
            return None
        expr, ty = m.group(1), m.group(2)
        pred = _find_pred(preds, ty)
        if pred is None:
            return None
        end = src.chain_end(pos, expr)
        if end is None or not re.match(r"\.\d", src.text[end:end + 2]):
            return None
        return repair("holds-field", pred, f"projection on {expr} through holds", Edit(end, end, ".holds"))

    parts = mismatch_parts(text)
    if parts is not None:
        term, has, exp = parts
        rng = _term_range(src, msg)
        if rng is None:
            return None
        start, end = rng
        term_src = src.text[start:end]
        has_pred = _find_pred(preds, has)
        exp_pred = _find_pred(preds, exp)
        # holds-proj: Lean prints `h.holds`; the source has `h.1`
        if term.endswith(".holds") and has_pred is None and exp_pred is None:
            x = term[: -len(".holds")]
            x_end = src.chain_end(start, x)
            if x_end is not None and re.match(r"\.\d", src.text[x_end:x_end + 2]):
                return repair("holds-proj", None, f"{x}.k on a one-field structure",
                              Edit(x_end, x_end, ".holds"))
        # holds-type: the term has the converted type, something else is expected
        if has_pred is not None and exp_pred is None:
            li = lead_ident(term)
            if li and li != lead_ident(term_src) and not term_src.startswith("("):
                return None
            if _is_application(term_src):
                return repair("holds-type", has_pred, f"({term_src[:40]}).holds, has type {has}",
                              Edit(start, start, "("), Edit(end, end, ").holds"))
            return repair("holds-type", has_pred, f"{term_src}.holds, has type {has}",
                          Edit(end, end, ".holds"))
        # wrap: the converted type is expected, the term has the body's type
        if exp_pred is not None and has_pred is None:
            if term_src.startswith("⟨"):
                return None
            li = lead_ident(term)
            if li and li != lead_ident(term_src) and not term_src.startswith("("):
                return None
            return repair("wrap", exp_pred, f"⟨{term_src[:40]}⟩ expected {exp}",
                          Edit(start, start, "⟨"), Edit(end, end, "⟩"))
        # wrap-nested: `{ holds := term }` where the body is itself a structure
        if NESTED_HOLDS_RE.search(text) and not term_src.startswith("⟨"):
            exp_short = head_token(exp).split(".")[-1]
            for pred in preds:
                if pred.body_is_structure and pred.body_head and \
                        exp_short == pred.body_head.rsplit(".", 1)[-1]:
                    return repair("wrap-nested", pred, f"⟨{term_src[:40]}⟩ expected {exp}",
                                  Edit(start, start, "⟨"), Edit(end, end, "⟩"))
        return None

    # unfold-iff: "Tactic `unfold` failed to unfold `P` in …"
    m = re.match(r"Tactic `unfold` failed to unfold `([^`]+)`", text)
    if m:
        pred = _pred_by_ident(preds, m.group(1))
        if pred is None:
            return None
        return _unfold_repair(src, msg, pos, pred)

    # simp-iff: "`simp` made no progress" at a simp call that lists the predicate
    if re.match(r"`?(simp|simp_all|simpa|dsimp)`? made no progress", text):
        return _simp_repair(src, msg, pos, preds)

    # constructor: "Tactic `t` failed …\n⊢ P args"
    if text.startswith("Tactic `"):
        target = goal_target(text)
        if target is None:
            return None
        pred = _find_pred(preds, target)
        if pred is None:
            return None
        tstart = _tactic_start(src, pos)
        head = lead_ident(src.text[tstart:tstart + 40])
        if head not in GOAL_TACTICS:
            return None
        return repair("constructor", pred, f"constructor before {head}",
                      _insert_constructor(src, tstart))

    return None


def _unfold_repair(src: Source, msg: Message, pos: int, pred: Predicate) -> Repair | None:
    k = src.text.rfind("unfold", src.line_start(pos), pos)
    if k < 0:
        return None
    tstart = k
    tend = _tactic_end(src, tstart)
    tactic = src.text[tstart:tend]
    m = re.match(r"unfold((?:\s+[^\s;)⟩]+)+?)(\s+at\s+.+)?\s*$", tactic)
    if not m:
        return None
    names = m.group(1).split()
    at = (m.group(2) or "").rstrip()
    keep = [n for n in names if _pred_by_ident([pred], n) is None]
    if len(keep) == len(names):
        return None
    written = next(n for n in names if _pred_by_ident([pred], n) is not None)
    new = f"rw [{written}.iff]{at}"
    if keep:
        sep = ("\n" + _indent_of(src, tstart)) if _starts_line(src, tstart) else "; "
        new += f"{sep}unfold {' '.join(keep)}{at}"
    return Repair(msg.file, msg.line, msg.col, "unfold-iff", pred.full_name,
                  f"unfold {written} → rw [{written}.iff]", (Edit(tstart, tend, new),))


def _simp_repair(src: Source, msg: Message, pos: int, preds: list[Predicate]) -> Repair | None:
    """`name.iff` for every listed predicate name in the simp call the
    message's range covers (`says` included when the range covers it)."""
    if msg.end_line is None or msg.end_col is None:
        return None
    end = src.offset(msg.end_line, msg.end_col)
    call = src.text[pos:end]
    if not any(call.lstrip("(").startswith(t) for t in SIMP_FAMILY):
        return None
    edits: list[Edit] = []
    hit: Predicate | None = None
    for m in re.finditer(r"\[([^\]]*)\]", call):
        base = pos + m.start(1)
        for item in re.finditer(r"[^,\s\[\]]+", m.group(1)):
            name = item.group(0).lstrip("←↓↑")
            pred = _pred_by_ident(preds, name)
            if pred is None:
                continue
            hit = pred
            off = base + item.end()
            edits.append(Edit(off, off, ".iff"))
    if hit is None:
        return None
    return Repair(msg.file, msg.line, msg.col, "simp-iff", hit.full_name,
                  f"{hit.short} → {hit.short}.iff in the simp set", tuple(edits))


def apply_edits(text: str, edits: list[Edit]) -> str:
    """Apply non-overlapping edits, last offset first."""
    for e in sorted(edits, key=lambda e: (e.start, e.end), reverse=True):
        text = text[: e.start] + e.replacement + text[e.end:]
    return text
