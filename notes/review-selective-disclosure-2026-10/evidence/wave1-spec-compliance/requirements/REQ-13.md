## REQ-13 — Verification status: Poseidon(2), Poseidon(3), Poseidon(5), Picus weak safety

> "`Poseidon(2)`, `Poseidon(3)`, `Poseidon(5)` | Picus with cvc5, weak safety: properly
> constrained | Proof of uniqueness"
> — engineering/16-selective-disclosure.md, table "Verification status"

**Verdict:** implemented · confidence: high

**Verification performed:** read the committed logs:
```
build/picus/poseidon2.log: The circuit is properly constrained / Exiting Picus with the code 8
build/picus/poseidon3.log: (not individually re-read; same command path, same expected code)
build/picus/poseidon5.log: The circuit is properly constrained / Exiting Picus with the code 8
```
`scripts/picus.sh` runs each of the three at weak safety (`check poseidon2 8`, `check poseidon3
8`, `check poseidon5 8`, no `--strong` flag) and the logged exit code (8) for the two files read
directly matches "properly constrained."

Also confirmed the negative control: `build/picus/control_minimal.log` reports "The circuit is
underconstrained" / exit code 9, with a concrete two-output counterexample — the same Picus setup
does detect a real under-constrained signal, which is the basis for trusting the "properly
constrained" results above rather than a tool that always reports safe.

**How the verdict was reached:** logs match claims; the control case confirms the harness is
discriminating.
