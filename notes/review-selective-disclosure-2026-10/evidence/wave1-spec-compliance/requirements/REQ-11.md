## REQ-11 — Circuit size: 8,248 non-linear + 1,835 linear constraints

> "It is a few Poseidon hashes and one EdDSA check: 8,248 non-linear and 1,835 linear
> constraints."
> — engineering/16-selective-disclosure.md, "The disclosure circuit"
(also `armada-circuits/README.md:L61`: "8,248 non-linear and 1,835 linear constraints (`--O1`)")

**Verdict:** implemented · confidence: high

**What this demands:** compiling `main_disclosure.circom` at `--O1` MUST report exactly these two
counts.

**Verification performed:** compiled `src/main/main_disclosure.circom` directly with the repo's
pinned `circom2` (`node_modules/.bin/circom2`, circom compiler 2.2.2) at `--O1`:

```
$ node_modules/.bin/circom2 src/main/main_disclosure.circom --r1cs -o .scratch/build-check --O1
non-linear constraints: 8248
linear constraints: 1835
```

Exact match to both the page-16 claim and the README's claim.

**How the verdict was reached:** reproduced the compilation independently rather than trusting
the stated numbers; the output matches to the digit.
