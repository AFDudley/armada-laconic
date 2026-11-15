"""Tests for `repairs.match` against recorded Lean 4.35 messages.

The messages come from `lake env lean --json` on small files that reproduce
each break (recorded 2026-11-14, Lean v4.35.0, Mathlib v4.35.0). Each case
gives the source line, the message, and the expected edit result, or
`None` when the matcher must refuse.

Run: `python3 -m pytest tools/lean_def_to_structure/` or `python3 test_repairs.py`.
"""

from __future__ import annotations

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from repairs import Message, Predicate, Source, apply_edits, match  # noqa: E402

REG = Predicate("Exp.Reg", "Reg", None, False)
DAG = Predicate("Exp.Dag", "Dag", "Exp.Acyclic", True)
TOPO = Predicate("Exp.Topo", "Topo", "And", False)
ISDAG = Predicate("ArmadaLean.Nitro.Adjudicator.IsDAG", "Adjudicator.IsDAG",
                  "ArmadaLean.Nitro.Acyclic", True)
PREDS = [REG, DAG, TOPO, ISDAG]


def msg(text: str, line: int, col: int, end_line: int | None = None,
        end_col: int | None = None, file: str = "t.lean") -> Message:
    return Message(file, line, col, end_line, end_col, "error", text)


def run(src_text: str, m: Message, preds=PREDS):
    src = Source(src_text)
    r = match(m, src, preds)
    if r is None:
        return None, None
    return r, apply_edits(src_text, list(r.edits))


FUNCTION_EXPECTED = ("Function expected at\n  {term}\nbut this term has type\n  Reg reg\n\n"
                     "Note: Expected a function because this term is being applied to the argument\n  c")


class HoldsApply(unittest.TestCase):
    def test_term_mode(self):
        src = "theorem t1 (h : Reg reg) : pre + 1 = c :=\n  h c pre hc\n"
        r, out = run(src, msg(FUNCTION_EXPECTED.format(term="h"), 2, 2, 2, 12))
        self.assertEqual(r.pattern, "holds-apply")
        self.assertIn("  h.holds c pre hc\n", out)

    def test_exact(self):
        src = "theorem t1b : pre + 1 = c := by\n  exact h c pre hc\n"
        r, out = run(src, msg(FUNCTION_EXPECTED.format(term="h"), 2, 8, 2, 18))
        self.assertIn("exact h.holds c pre hc", out)

    def test_projection_chain(self):
        src = "theorem v5 (h : True ∧ Reg reg) : pre + 1 = c :=\n  h.2 c pre hc\n"
        r, out = run(src, msg(FUNCTION_EXPECTED.format(term="h.right"), 2, 2, 2, 14))
        self.assertIn("  h.2.holds c pre hc\n", out)

    def test_parenthesized(self):
        src = "theorem v4 (h : Reg reg) : pre + 1 = c :=\n  (h) c pre hc\n"
        r, out = run(src, msg(FUNCTION_EXPECTED.format(term="h"), 2, 2, 2, 14))
        self.assertIn("  (h).holds c pre hc\n", out)

    def test_refuses_other_type(self):
        src = "theorem t (h : Foo reg) : pre + 1 = c :=\n  h c pre hc\n"
        text = FUNCTION_EXPECTED.replace("Reg reg", "Foo reg").format(term="h")
        r, _ = run(src, msg(text, 2, 2, 2, 12))
        self.assertIsNone(r)

    def test_refuses_name_mismatch(self):
        src = "theorem t (h : Reg reg) : pre + 1 = c :=\n  g c pre hc\n"
        r, _ = run(src, msg(FUNCTION_EXPECTED.format(term="h"), 2, 2, 2, 12))
        self.assertIsNone(r)


class HoldsField(unittest.TestCase):
    INVALID_FIELD = ("Invalid field `irrefl`: The environment does not contain `Exp.Dag.irrefl`, "
                     "so it is not possible to project the field `irrefl` from an expression\n  {expr}\n"
                     "of type `Dag r`")
    INVALID_PROJ = ("Invalid projection: Index `2` is invalid for this structure; the only valid index is 1"
                    "\n\nNote: The expression\n  {expr}\nhas type `Topo l` which has only 1 field")

    def test_field(self):
        src = "theorem t1g : ¬ Relation.TransGen r n n :=\n  h.irrefl n\n"
        r, out = run(src, msg(self.INVALID_FIELD.format(expr="h"), 2, 4, 2, 10))
        self.assertEqual(r.pattern, "holds-field")
        self.assertIn("  h.holds.irrefl n\n", out)

    def test_field_after_projection(self):
        src = "theorem u2b : ¬ Relation.TransGen r n n :=\n  h.1.irrefl n\n"
        r, out = run(src, msg(self.INVALID_FIELD.format(expr="h.left"), 2, 6, 2, 12))
        self.assertIn("  h.1.holds.irrefl n\n", out)

    def test_field_parenthesized(self):
        src = "theorem u2 : ¬ Relation.TransGen r n n :=\n  (h.1).irrefl n\n"
        r, out = run(src, msg(self.INVALID_FIELD.format(expr="h.left"), 2, 8, 2, 14))
        self.assertIn("  (h.1).holds.irrefl n\n", out)

    def test_field_refuses_foreign_structure(self):
        src = "theorem t : ¬ Relation.TransGen r n n :=\n  h.irrefl n\n"
        text = self.INVALID_FIELD.replace("Exp.Dag.irrefl", "Exp.Other.irrefl")
        r, _ = run(src, msg(text.format(expr="h"), 2, 4, 2, 10))
        self.assertIsNone(r)

    def test_projection(self):
        src = "theorem t1e (h : Topo l) : l ≠ [] := h.2.2\n"
        c = src.index("h.2.2")
        r, out = run(src, msg(self.INVALID_PROJ.format(expr="h"), 1, c, 1, c + 5))
        self.assertEqual(r.pattern, "holds-field")
        self.assertIn(":= h.holds.2.2\n", out)

    def test_projection_applied(self):
        src = "theorem v2 (h : Topo l) : n < 10 :=\n  h.2.1 n hn\n"
        r, out = run(src, msg(self.INVALID_PROJ.format(expr="h"), 2, 2, 2, 12))
        self.assertIn("  h.holds.2.1 n hn\n", out)

    def test_projection_after_projection(self):
        src = "theorem v3 (h : Topo l ∧ True) : n < 10 :=\n  h.1.2.1 n hn\n"
        r, out = run(src, msg(self.INVALID_PROJ.format(expr="h.left"), 2, 2, 2, 14))
        self.assertIn("  h.1.holds.2.1 n hn\n", out)


class HoldsTypeAndProj(unittest.TestCase):
    def test_holds_proj(self):
        src = "theorem t1d (l : List Nat) (h : Topo l) : l.Nodup := h.1\n"
        text = ("Type mismatch\n  h.holds\nhas type\n  l.Nodup ∧ (∀ n ∈ l, n < 10) ∧ l ≠ []\n"
                "but is expected to have type\n  l.Nodup")
        r, out = run(src, msg(text, 1, 53, 1, 56))
        self.assertEqual(r.pattern, "holds-proj")
        self.assertIn(":= h.holds.1\n", out)

    def test_holds_type_exact(self):
        src = "theorem u1 (h : Both r reg) : Acyclic r := by\n  exact h.1\n"
        text = "Type mismatch\n  h.left\nhas type\n  Dag r\nbut is expected to have type\n  Acyclic r"
        r, out = run(src, msg(text, 2, 2, 2, 11))
        self.assertEqual(r.pattern, "holds-type")
        self.assertIn("  exact h.1.holds\n", out)

    def test_holds_type_argument(self):
        src = "theorem u5 (h : Reg reg) : True := needsBody reg h\n"
        text = ("Application type mismatch: The argument\n  h\nhas type\n  Reg reg\n"
                "but is expected to have type\n  ∀ (c pre : ℕ), reg c = some pre → pre + 1 = c\n"
                "in the application\n  needsBody reg h")
        r, out = run(src, msg(text, 1, 49, 1, 50))
        self.assertEqual(r.pattern, "holds-type")
        self.assertIn("needsBody reg h.holds\n", out)

    def test_holds_type_application_gets_parens(self):
        src = "theorem u (h : Reg reg) : body := by\n  exact mk_reg h x\n"
        text = "Type mismatch\n  mk_reg h x\nhas type\n  Reg reg\nbut is expected to have type\n  ∀ c, c = c"
        r, out = run(src, msg(text, 2, 2, 2, 18))
        self.assertIn("  exact (mk_reg h x).holds\n", out)

    def test_field_notation_type(self):
        src = "theorem u (h : s.IsDAG) : Acyclic s.Edge := by\n  exact h\n"
        text = "Type mismatch\n  h\nhas type\n  s.IsDAG\nbut is expected to have type\n  Acyclic s.Edge"
        r, out = run(src, msg(text, 2, 2, 2, 9))
        self.assertEqual(r.predicate, ISDAG.full_name)
        self.assertIn("  exact h.holds\n", out)


class Wrap(unittest.TestCase):
    def test_fun_term(self):
        src = "theorem t2 : Reg (fun _ => none) := fun _ _ h => by injection h\n"
        text = ("Type mismatch\n  fun x x_1 h => ?m.7\nhas type\n  (x : ?m.3) → (x_1 : ?m.8 x) → "
                "(h : ?m.9 x x_1) → ?m.10 x x_1 h\nbut is expected to have type\n  Reg fun x => none")
        r, out = run(src, msg(text, 1, 36, 1, 63))
        self.assertEqual(r.pattern, "wrap")
        self.assertIn(":= ⟨fun _ _ h => by injection h⟩\n", out)

    def test_hypothesis(self):
        src = "theorem t2b (hb : ∀ c pre, reg c = some pre → pre + 1 = c) : Reg reg := hb\n"
        text = ("Type mismatch\n  hb\nhas type\n  ∀ (c pre : ℕ), reg c = some pre → pre + 1 = c\n"
                "but is expected to have type\n  Reg reg")
        r, out = run(src, msg(text, 1, 72, 1, 74))
        self.assertIn(":= ⟨hb⟩\n", out)

    def test_exact(self):
        src = "theorem t2c : Reg reg := by\n  exact hb\n"
        text = ("Type mismatch\n  hb\nhas type\n  ∀ (c pre : ℕ), reg c = some pre → pre + 1 = c\n"
                "but is expected to have type\n  Reg reg")
        r, out = run(src, msg(text, 2, 2, 2, 10))
        self.assertIn("  exact ⟨hb⟩\n", out)

    def test_argument_in_anonymous_constructor(self):
        src = "theorem t2d : ∃ reg, Reg reg := ⟨fun _ => none, fun _ _ h => by injection h⟩\n"
        text = ("Application type mismatch: The argument\n  fun x x_1 h => ?m.10\nhas type\n  (x : ?m.6) → "
                "(x_1 : ?m.11 x) → (h : ?m.12 x x_1) → ?m.13 x x_1 h\nbut is expected to have type\n"
                "  Reg fun x => none\nin the application\n  Exists.intro (fun x => none) fun x x_1 h => ?m.10")
        r, out = run(src, msg(text, 1, 48, 1, 75))
        self.assertIn("⟨fun _ => none, ⟨fun _ _ h => by injection h⟩⟩", out)

    def test_multi_line_fun(self):
        src = ("theorem u7 : ∃ reg, Reg reg :=\n  ⟨fun _ => none,\n    fun _ _ h => by\n"
               "      injection h⟩\n")
        text = ("Application type mismatch: The argument\n  fun x x_1 h => ?m.10\nhas type\n  T\n"
                "but is expected to have type\n  Reg fun x => none\nin the application\n  Exists.intro …")
        r, out = run(src, msg(text, 3, 4, 4, 17))
        self.assertIn("    ⟨fun _ _ h => by\n      injection h⟩⟩\n", out)

    def test_nested(self):
        src = "theorem t2e (hr : ∀ n, ¬ TransGen r n n) : Dag r := ⟨fun n hn => hr n hn⟩\n"
        text = ("Application type mismatch: The argument\n  fun n hn => hr n hn\nhas type\n"
                "  ∀ (n : ℕ), Relation.TransGen r n n → False\nbut is expected to have type\n  Acyclic r\n"
                "in the application\n  { holds := fun n hn => hr n hn }")
        c = src.index("fun n")
        r, out = run(src, msg(text, 1, c, 1, c + len("fun n hn => hr n hn")))
        self.assertEqual(r.pattern, "wrap-nested")
        self.assertIn(":= ⟨⟨fun n hn => hr n hn⟩⟩\n", out)

    def test_nested_multi_line_instance(self):
        src = "  ⟨fun n hn => hdag.holds.irrefl n\n    (Relation.TransGen.mono (fun x y hxy => hxy) n n hn)⟩\n"
        text = ("Application type mismatch: The argument\n  fun n hn => Std.Irrefl.irrefl n (…)\nhas type\n"
                "  ∀ (n : Address E), Relation.TransGen s'.Edge n n → False\nbut is expected to have type\n"
                "  Acyclic s'.Edge\nin the application\n  {\n    holds := fun n hn =>\n      Std.Irrefl.irrefl n (…) }")
        r, out = run(src, msg(text, 1, 3, 2, 56), [ISDAG])
        self.assertEqual(r.pattern, "wrap-nested")
        self.assertTrue(out.startswith("  ⟨⟨fun n hn => hdag.holds.irrefl n\n"))
        self.assertTrue(out.endswith("n n hn)⟩⟩\n"))

    def test_nested_refuses_when_body_is_not_a_structure(self):
        src = "theorem t : Reg reg := ⟨fun c pre h => foo⟩\n"
        text = ("Application type mismatch: The argument\n  fun c pre h => foo\nhas type\n  T\n"
                "but is expected to have type\n  Acyclic r\nin the application\n  { holds := fun c pre h => foo }")
        r, _ = run(src, msg(text, 1, 24, 1, 42), [REG])
        self.assertIsNone(r)

    def test_refuses_already_wrapped(self):
        src = "theorem t : Reg reg := ⟨hb⟩\n"
        text = "Type mismatch\n  ⟨hb⟩\nhas type\n  T\nbut is expected to have type\n  Reg reg"
        r, _ = run(src, msg(text, 1, 23, 1, 27))
        self.assertIsNone(r)


class Constructor(unittest.TestCase):
    INTRO = ("Tactic `introN` failed: There are no additional binders or `let` bindings in the goal "
             "to introduce\n\n⊢ Reg fun x => none")

    def test_intro_on_its_own_line(self):
        src = "theorem t3 : Reg (fun _ => none) := by\n  intro c pre h\n  injection h\n"
        r, out = run(src, msg(self.INTRO, 2, 8, 2, 9))
        self.assertEqual(r.pattern, "constructor")
        self.assertIn("by\n  constructor\n  intro c pre h\n", out)

    def test_rintro_after_focus_dot(self):
        src = "theorem t : Reg (fun _ => none) := by\n  · rintro c pre ⟨⟩\n"
        r, out = run(src, msg(self.INTRO, 2, 11, 2, 12))
        self.assertIn("  · constructor; rintro c pre ⟨⟩\n", out)

    def test_intro_in_match_alternative(self):
        src = "  | genesis _ _ _ => intro c pre hc; simp at hc\n"
        r, out = run(src, msg(self.INTRO.replace("Reg fun x => none", "RegistryConsistent s.preimageOf"),
                              1, 27, 1, 28), [Predicate("ArmadaLean.RegistryConsistent", "RegistryConsistent", None, False)])
        self.assertIn("=> constructor; intro c pre hc; simp at hc\n", out)

    def test_rewrite(self):
        src = "  · unfold Adjudicator.IsDAG\n    rw [Adjudicator.edge_deposit hd]\n    exact h.1\n"
        text = ("Tactic `rewrite` failed: Did not find an instance of the pattern in the target expression\n"
                "  s.Edge\n\nh : s.IsTurboOutcomeNetwork\n⊢ s'.IsDAG")
        r, out = run(src, msg(text, 2, 8, 2, 35))
        self.assertIn("    constructor\n    rw [Adjudicator.edge_deposit hd]\n", out)

    def test_parenthesized_receiver_goal(self):
        src = "theorem turboT_only (p : Participant) : (turboT p).TransfersOnly := by\n  rintro _ _ ⟨caller, a, b, -, hm⟩; exact ⟨caller, a, b, hm⟩\n"
        to = Predicate("ArmadaLean.Nitro.Threat.TransfersOnly", "Threat.TransfersOnly", None, False)
        r, out = run(src, msg(self.INTRO.replace("Reg fun x => none", "(turboT p).TransfersOnly"), 2, 9, 2, 10), [to])
        self.assertEqual(r.pattern, "constructor")
        self.assertIn("by\n  constructor\n  rintro _ _", out)

    def test_refuses_other_goal(self):
        src = "theorem t : Foo := by\n  intro c pre h\n"
        r, _ = run(src, msg(self.INTRO.replace("Reg fun x => none", "Foo"), 2, 8, 2, 9))
        self.assertIsNone(r)

    def test_refuses_unknown_tactic(self):
        src = "theorem t : Reg r := by\n  injection h\n"
        text = "Tactic `injection` failed: equality expected\n\nh : x = y\n⊢ Reg fun x => none"
        r, _ = run(src, msg(text, 2, 2, 2, 13))
        self.assertIsNone(r)


class UnfoldAndSimp(unittest.TestCase):
    UNFOLD = "Tactic `unfold` failed to unfold `{name}` in\n  Dag r"

    def test_unfold_two_names(self):
        src = "theorem t3b : Dag r := by\n  unfold Dag Acyclic\n  exact ⟨hr⟩\n"
        r, out = run(src, msg(self.UNFOLD.format(name="Dag"), 2, 9, 2, 12))
        self.assertEqual(r.pattern, "unfold-iff")
        self.assertIn("by\n  rw [Dag.iff]\n  unfold Acyclic\n  exact ⟨hr⟩\n", out)

    def test_unfold_one_name(self):
        src = "theorem t3c : Dag r := by\n  unfold Dag\n  exact ⟨hr⟩\n"
        r, out = run(src, msg(self.UNFOLD.format(name="Dag"), 2, 9, 2, 12))
        self.assertIn("by\n  rw [Dag.iff]\n  exact ⟨hr⟩\n", out)

    def test_unfold_qualified_name_with_at(self):
        src = "  unfold Adjudicator.IsDAG at h\n"
        r, out = run(src, msg(self.UNFOLD.format(name="Adjudicator.IsDAG"), 1, 9, 1, 26))
        self.assertEqual(out, "  rw [Adjudicator.IsDAG.iff] at h\n")

    def test_unfold_refuses_other_name(self):
        src = "  unfold Other\n"
        r, _ = run(src, msg(self.UNFOLD.format(name="Other"), 1, 9, 1, 14))
        self.assertIsNone(r)

    def test_simp_only(self):
        src = "theorem t3e : Reg (fun _ => none) := by\n  simp only [Reg]\n"
        r, out = run(src, msg("`simp` made no progress", 2, 2, 2, 17))
        self.assertEqual(r.pattern, "simp-iff")
        self.assertIn("  simp only [Reg.iff]\n", out)

    def test_simp_says(self):
        src = ("  | nil => exact ⟨[], by (simp? [OnlyTransfers] says simp only [OnlyTransfers, "
               "List.not_mem_nil, IsEmpty.forall_iff, implies_true]), rfl⟩\n")
        ot = Predicate("ArmadaLean.Nitro.OnlyTransfers", "OnlyTransfers", None, False)
        r, out = run(src, msg("`simp` made no progress", 1, 26, 1, 130), [ot])
        self.assertIn("(simp? [OnlyTransfers.iff] says simp only [OnlyTransfers.iff, List.not_mem_nil", out)

    def test_simp_refuses_without_predicate(self):
        src = "  simp only [foo]\n"
        r, _ = run(src, msg("`simp` made no progress", 1, 2, 1, 17))
        self.assertIsNone(r)


class Refusals(unittest.TestCase):
    def test_unsolved_goals(self):
        src = "theorem t : X := by\n  foo\n"
        r, _ = run(src, msg("unsolved goals\nreg : ℕ\n⊢ pre + 1 = c", 1, 17, 2, 5))
        self.assertIsNone(r)

    def test_warning_is_ignored(self):
        src = "theorem t (h : Reg reg) : pre + 1 = c :=\n  h c pre hc\n"
        m = Message("t.lean", 2, 2, 2, 12, "warning", FUNCTION_EXPECTED.format(term="h"))
        self.assertIsNone(match(m, Source(src), PREDS))


if __name__ == "__main__":
    unittest.main()
