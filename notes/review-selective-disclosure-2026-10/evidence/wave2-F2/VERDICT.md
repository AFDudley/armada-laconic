# fp-check verdict: F2 — Picus "proof of uniqueness" for `wiring.circom`

## Verdict

**FALSE POSITIVE** — the claim that page 16 and PR #3 overstate what the Picus
run proves.

The root technical fact F2 cites is correct: Picus's uniqueness check does
not use circom's `public`/`private` split. But page 16 and PR #3 do not
claim it does. Both texts describe only the Picus invocation that was run
and its narrow, accurately labeled result. F2 restates the "Proof of
uniqueness" claim as a stronger claim than either document makes, then
refutes that stronger, invented claim.

## Step 0: the claim, restated

F2 says: the design doc's verification-status table (page 16, row 1)
labels a Picus strong-safety "safe" result on `tests/picus/wiring.circom`
as a "Proof of uniqueness." F2 says this result cannot be that proof,
because `wiring.circom`'s `component main = Wiring();` line carries no
`public [...]` list, so every declared signal is circom-private, and
(F2 claims) Picus does not care about the `public`/`private` split anyway:
it holds every declared input, public and private together, fixed as
"known." So, F2 argues, the check never tests whether an attacker who
knows only the 5 real public inputs is forced to a unique private
witness — the property F2 reads "Proof of uniqueness" to promise.

## Gate 1: Picus semantics, read from source (commit `138b151`)

Confirmed, matching F2's citation:

- `picus/r1cs/r1cs-grammar.rkt:408-409`: `input-list` spans
  `npubout+1` through `npubout+npubin+nprvin` — public inputs and
  private inputs combined, one undivided span.
- `picus.rkt:158-159`: `input-set` is built directly from that
  `input-list` and held fixed as "known" in both models Picus compares.
- `picus.rkt:162-164`: `--strong` sets the target set to every wire in
  the circuit; without `--strong` (weak safety) the target set is only
  the declared outputs.
- The circom-source frontend parses a `public [...]` list
  (`picus/circom/circom-grammar.rkt:22`, `circom-parser.rkt:289`) into a
  `circom:component` struct field. `grep -rn "circom:component" .`
  (excluding the parser itself) finds no other reference to that field
  anywhere in the Picus tree. The parsed list is stored and never read.

So: Picus's "known"/fixed set is every declared top-level input signal,
public and private together, regardless of the circom `public [...]`
annotation, for both the R1CS path and the circom-source path. F2's
technical premise about the tool is correct.

## Gate 2: does the annotation change anything, empirically?

Reproduced the experiment F2 proposed. Added the real circuit's public
list to a scratch copy of `wiring.circom` and compiled both versions with
`circom2 --r1cs --sym --O0`, then ran Picus (`picus:138b151`, cvc5,
`--strong`) on both.

```
$ npx circom2 tests/picus/wiring.circom --r1cs --sym -o OUT --O0
public inputs: 0
private inputs: 14

$ npx circom2 tests/picus/wiring_public.circom --r1cs --sym -o OUT --O0
   # wiring_public.circom = wiring.circom + "{public [statementHash,
   #   masterPublicKey, nullifier, treeNumber, leafIndex]}" on component main
public inputs: 5
private inputs: 9

$ cmp -l OUT/wiring.r1cs OUT/wiring_public.r1cs
6789   0   5     # header byte: npubin  0 -> 5
6793  16  11     # header byte: nprvin 14 -> 9 (octal 16=dec14, octal 11=dec9)
   (no other bytes differ; wiring.sym and wiring_public.sym are identical)

$ docker run --rm --memory=8g -v OUT:/w picus:138b151 \
    ./run-picus --solver cvc5 --timeout 600000 --strong /w/wiring.r1cs
The circuit is properly constrained
Exiting Picus with the code 8

$ docker run --rm --memory=8g -v OUT:/w picus:138b151 \
    ./run-picus --solver cvc5 --timeout 600000 --strong /w/wiring_public.r1cs
The circuit is properly constrained
Exiting Picus with the code 8
```

The two R1CS files differ in exactly two header bytes (the npubin/nprvin
split). Every constraint, every wire, and the combined input span are
byte-identical. Picus returns the identical result (exit 8, "properly
constrained") for both. Adding the real `public [...]` list to
`wiring.circom` changes nothing about what Picus checks or finds — exactly
as predicted by the source reading in Gate 1, and directly confirming
F2's prediction that this experiment would not move the result (F2 framed
a changed result as the falsifying outcome for its own claim; the actual,
unchanged result falsifies the inference F2 draws from it, not the
underlying Picus-mechanics fact, see Gate 3).

## Gate 3: is this a limitation of the check, or the check's standard definition?

Picus implements one property, for any circuit: given one assignment to
every top-level declared input signal (circom calls some of these public
and some private; R1CS does not use that distinction for picking the
"known" set), is every other wire in scope (every wire, for `--strong`)
forced to a single value by the constraints? This is the paper's
"under-constrained signal" bug: a value the prover supplies off-circuit
(whether revealed to a verifier or not) must not leave some other wire
free to take two different values under the same supplied values. Picus
has no notion of, and the QED² literature does not define, a "private
witness is unique given only the public inputs" property — that is not
an R1CS-satisfiability question an under-constrained-signal checker can
answer; it is a separate cryptographic soundness question answered by the
EdDSA signature and the hash functions, which page 16 discusses in prose
immediately below the table (lines 69-73), not by Picus.

So "uniqueness of all signals given all declared inputs, public and
private together" is not a watered-down stand-in for a stronger property
page 16 needed. It is the standard, and only, property a Picus
strong-safety run proves, for every circuit it checks, by design.

## Gate 4: what does page 16 / PR #3 actually say?

Page 16, line 81 (verbatim): "Our wiring: the range checks, the equality
checks, the on-curve checks and the message routing | Picus with cvc5,
strong safety, with the hashes and the verifier replaced by free inputs:
properly constrained | Proof of uniqueness."

This sentence names the exact abstraction run (hashes and verifier
replaced by free inputs) and the exact result (properly constrained). It
makes no statement about circom's `public`/`private` declarations and no
statement that the check used only the 5 real public inputs. The word
"unique" appears nowhere else in the document; "Proof of uniqueness" is
the table's "Kind" column label distinguishing a Picus machine result
from the "Tests only" rows below it, not a claim about a public-input-only
property.

PR #3's body (armada-circuits#3) is, if anything, more conservative: it
reports "properly constrained" (never "proof of uniqueness"), and its
"Limits" section already discloses that Picus returned `unknown` rather
than `unsafe` when the authors hand-broke one wiring output on purpose —
an explicit admission that Picus does not reliably catch every defect in
this piece. Neither document claims, anywhere, that the public/private
boundary was what Picus tested.

`src/operations/NullifierDisclosure.circom` (the real circuit) does carry
the actual `public [...]` split on its `main` component — this was
checked separately and is out of scope for F2, which is specifically
about the `tests/picus/wiring.circom` abstraction used for the Picus run.

## Conclusion

F2's claim is that page 16/PR #3 overstate what the `wiring.circom` Picus
run proves, because the run does not preserve the real public/private
split. The Picus-mechanics premise is true and independently reproduced
(Gates 1-2). But it does not support the conclusion: "uniqueness given
all declared inputs" is the correct and only thing Picus strong safety
proves for any circuit (Gate 3), and page 16/PR #3's wording already
states precisely that, and nothing more (Gate 4). F2 attributes a
stronger, public-input-only reading to the words "Proof of uniqueness"
that the source text does not carry, then shows that stronger reading
fails — a claim-restated-too-strongly false positive (fp-check Step 0).

**What the wiring result does prove:** given one full assignment to
`wiring.circom`'s 11 declared inputs (`statementHash`, `masterPublicKey`,
`nullifier`, `treeNumber`, `leafIndex`, `spendingPublicKey[2]`,
`nullifyingKey`, `signature[3]`, and the three hash stand-ins `mpkHash`,
`nfHash`, `msgHash`), every other wire in the piece — the `Num2Bits`
range-check internals, the `BabyCheck` on-curve internals, and the output
`verifierM` — is forced to a single value by the constraints. No stray
unconstrained signal exists in this wiring piece, under the stated
abstraction (hashes and verifier replaced by free inputs).

**What it does not prove, and what neither document claims it proves:**
that an attacker who knows only the 5 real public inputs (`statementHash`,
`masterPublicKey`, `nullifier`, `treeNumber`, `leafIndex`) and not the
private witness is forced to a unique `spendingPublicKey`/`nullifyingKey`/
`signature`. No Picus run, with or without a `public [...]` annotation,
tests that property; it is not a property Picus's uniqueness algorithm
can express.

**Accurate wording:** page 16's row already is accurate wording. If a
clarification is wanted, the row could add one clause: "properly
constrained, given a full assignment to the wiring piece's declared
inputs (not scoped to the real circuit's public/private split, which
Picus's uniqueness check does not use for any circuit)." This closes the
documentation gap F2 points at — readers who do not already know Picus's
input-handling internals might guess the public/private split mattered —
without implying the original claim was false.

## Evidence index

- `/home/rix/code/git_puller/repos/Picus/picus/r1cs/r1cs-grammar.rkt:404-421`
- `/home/rix/code/git_puller/repos/Picus/picus.rkt:146-167`
- `/home/rix/code/git_puller/repos/Picus/picus/circom/circom-grammar.rkt:22`
- `/home/rix/code/git_puller/repos/Picus/picus/circom/circom-parser.rkt:286-296`
- `/home/rix/code/git_puller/repos/armada-circuits/tests/picus/wiring.circom`
- `/home/rix/code/git_puller/repos/armada-circuits/scripts/picus.sh`
- `/home/rix/code/git_puller/repos/armada-laconic/engineering/16-selective-disclosure.md:54-89`
- `pr://ship-armada/armada-circuits/3` (body, "Verification" section)
- Reproducer run (this task): `circom2` compile of `wiring.circom` vs. a
  scratch copy with `{public [statementHash, masterPublicKey, nullifier,
  treeNumber, leafIndex]}` added, and `docker run picus:138b151
  ./run-picus --solver cvc5 --strong` on both `.r1cs` files. Scratch files
  removed after the run; `git status --porcelain` in `armada-circuits` is
  clean.

## Severity

N/A — false positive, no fix required. If the optional clarifying clause
above is adopted, it is a documentation change to
`armada-laconic/engineering/16-selective-disclosure.md` line 81, not a
circuit or SDK change.
