import Lake
open Lake DSL

-- tools/lean_def_to_structure (see README.md): one Lean program that converts
-- `def P … : Prop := body` predicates of an armada-lean checkout to one-field
-- structures and repairs the proofs through `lake serve`.
package defToStructure

lean_lib DefToStructure

@[default_target]
lean_exe def_to_structure where
  root := `Main
  supportInterpreter := true

lean_exe tests where
  root := `Tests
  supportInterpreter := true
