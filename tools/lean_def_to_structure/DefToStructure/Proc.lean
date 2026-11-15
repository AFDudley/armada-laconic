/-
  Process environment for every `lake` the tool starts: `~/.elan/bin` first on
  `PATH`, `CI` unset (Mathlib's `says` re-runs every recorded search when
  `CI` is set).
-/

namespace DefToStructure

def leanEnv : IO (Array (String × Option String)) := do
  let home := (← IO.getEnv "HOME").getD ""
  let path := (← IO.getEnv "PATH").getD ""
  return #[("PATH", some s!"{home}/.elan/bin:{path}"), ("CI", none)]

end DefToStructure
