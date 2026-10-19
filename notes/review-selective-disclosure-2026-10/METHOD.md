# Independent review method — selective disclosure (circuit PR #3, SDK PR #142)

This method follows Trail of Bits' published practice. Read the cited files yourself before you start; do not work from this summary alone.

## Sources of the method (read these)
- Trail of Bits cryptography methodology: https://trailofbits.com/services/cryptography — five stages: scope and protocol model; specification review and cryptanalysis ("for ZK systems, audit constraints against the underlying circuit semantics"); implementation review; adversarial testing with differential tests against reference implementations; reporting.
- `audit-context-building`: /home/rix/code/git_puller/repos/skills/plugins/audit-context-building/skills/audit-context-building/SKILL.md and its `resources/ANALYSIS_FORMAT.md`, `resources/DOMAIN_NOTES.md`.
- `spec-to-code-compliance`: /home/rix/code/git_puller/repos/skills/plugins/spec-to-code-compliance/skills/spec-to-code-compliance/SKILL.md and its `resources/ANALYSIS_FORMAT.md`, `resources/DIVERGENCE_RUBRIC.md`, `resources/WORKED_EXAMPLE.md`, `resources/DOMAIN_NOTES.md`.
- `fp-check`: /home/rix/code/git_puller/repos/skills/plugins/fp-check/skills/fp-check/SKILL.md and `references/` (gate-reviews, false-positive-patterns, bug-class-verification, evidence-templates).
- ZKDocs: https://www.zkdocs.com/docs/zkdocs/protocol-primitives/fiat-shamir/ (hash all public values and the transcript; include prover and verifier identity against replay) and https://www.zkdocs.com/docs/zkdocs/protocol-primitives/random-sampling/ (modulo bias of a reduction to a field).

The ToB plugins run as Claude Code slash-command workflows, which are not available here. Apply their rules by hand:
- Every claim cites a file and line, or becomes an open question.
- Follow the calls: read every callee and every path, not only the success path. When nothing enforces an assumption, write `nothing found`.
- Spec compliance verdicts: `implemented`, `partial`, `contradicted`, `stronger-than-spec`, `absent`, `undecidable`. An `absent` verdict lists the searches you ran and their results.
- Do not rate severity during context building.
- A finding is not accepted until an agent that did not produce it tries to refute it (fp-check gates).

## Targets
- Specification: /home/rix/code/git_puller/repos/armada-laconic/engineering/16-selective-disclosure.md and ADR-0020 in engineering/09-architecture-decisions.md (same repo, uncommitted; read from disk).
- Circuit: /home/rix/code/git_puller/repos/armada-circuits, branch `disclosure-circuit` (PR https://github.com/ship-armada/armada-circuits/pull/3): `src/operations/NullifierDisclosure.circom`, `src/main/main_disclosure.circom`, `tests/disclosure.test.js`, `tests/picus/`, `scripts/picus.sh`. Reference formulas: `src/operations/Transact.circom`. circomlib 2.0.5 in `node_modules/circomlib/circuits/`.
- SDK: /home/rix/code/git_puller/repos/armada-sdk, branch `selective-disclosure` (draft PR https://github.com/ship-armada/armada-sdk/pull/142): `src/disclosure/`, `src/node/disclosure-artifacts.ts`, `docs/guide/selective-disclosure.md`.
- Chain facts the spec relies on: armada-poc `contracts/privacy-pool/` (ShieldModule, TransactModule events) and armada-nitro `src/ArmadaNitroAdjudicator.sol` (`depositFromNote`, `depositBinding`, `_executeSingleAssetExit`).

## Threat model to apply
- A dishonest prover who wants a verifier to accept a chain of custody that is false: claims another wallet's note, spend, channel deposit or payout; reuses or replays an old disclosure; substitutes a hop.
- A holder of only the viewing key (an auditor or a compromised device) who wants to make a disclosure.
- A verifier or third party who wants to learn more about the wallet than the package discloses.
- An attacker who manipulates the chain-event source the checker reads.

## Rules
- Read only. Do not change any repository. Write your output only to /home/rix/code/git_puller/repos/armada-laconic/.scratch/audit/<your-name>/ (create it; .scratch is gitignored).
- You did not write this code. Do not trust its comments, its test names or the PR descriptions; check behavior against the code and the spec.
- Write in ASD-STE100 Simplified Technical English. Mark every unverified statement as a hypothesis.
