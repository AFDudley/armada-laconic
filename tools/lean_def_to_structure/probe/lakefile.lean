import Lake
open Lake DSL

-- The elaboration probe of tools/lean_def_to_structure (see ../README.md).
-- A standalone executable: it links Lean, loads the target file's imports
-- from LEAN_PATH at run time (run it under `lake env` of the target
-- checkout), and never imports the target package at build time.
package probe

@[default_target]
lean_exe probe where
  root := `Main
  supportInterpreter := true
