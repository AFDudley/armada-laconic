# Decision records — style and mechanics

how the ADRs in [§9](./09-architecture-decisions.md) are written, superseded, and compiled · 2026-10-11

## The problem this solves

The ADRs are an append-only log: a decision, once accepted, is never edited, and a change is a new ADR that amends an old one ([ADR-0012](./09-architecture-decisions.md#adr-0012)). That keeps the reasoning behind every decision, but it makes the *current* decision hard to read:

- An ADR's own text can be stale. ADR-0011 still says "USDC yield defers to v2" after ADR-0017 changed it; only the index row hinted otherwise.
- To know what holds now, a reader finds every later ADR that touches an earlier one (0007 → 0011 → 0017) and merges them by hand, including partial amendments that change one sentence of a longer decision.
- Deriving a spec therefore means replaying the whole log, and nothing checks that the replay is right.

The goal: **a reader derives the current spec from one document, and a machine keeps that document correct.**

## The rules

### 1. A Decision is a list of numbered clauses

Each ADR's **Decision** section is a list of clauses, one decision each:

```markdown
- **D11.1** v1 exchange is filled from a simple pre-funded venue inventory ...
- **D11.2** No automated market-making runs in v1 ...
- **D11.3** v1 yield is ETH/wstETH only (held as a note, T4.4), and USDC yield defers to v2.
```

The ID is `D<ADR number>.<clause index>`, numbered from 1 within the ADR.

*Why.* Amendments rarely replace a whole decision; they replace one part of it (ADR-0017 changes only the USDC-yield sentence of ADR-0011). Numbered clauses give that part a name, so a change can say exactly what it replaces. One decision per clause keeps that name meaningful.

### 2. Only Decision clauses are normative

Context, Consequences and Alternatives stay prose. They explain; they are not compiled and they are not superseded.

*Why.* The spec is what was decided, not why. Keeping rationale out of the compiled view keeps that view short enough to read as a spec, while the rationale stays in the ADR where it was written.

### 3. Supersession is declared on the new clause

A clause that replaces earlier clauses names them:

```markdown
- **D17.1** (supersedes D11.3) USDC yield arrives with the v1.5 hub rail as `waUSDC` held in hub channels; the LP-buffered rail (T4.5) stays v2.
```

Several targets are allowed: `(supersedes D7.2, D11.1)`.

*Why.* The author of a change is the one who knows what it replaces, and they write it once, in the clause they are already writing. Old ADRs are never edited, so the log keeps its integrity: what ADR-0011 decided on 2026-09-05 still reads as it was decided.

### 4. A partial change supersedes the whole clause and restates what still holds

If only part of a clause changes, the new clause supersedes all of it and repeats the part that still holds.

*Why.* A clause is either live or superseded; there is no "partly superseded" state for a reader or a tool to reconcile. The cost is one restated sentence; the benefit is that every live clause is complete on its own.

### 5. No pointer prose

Status lines and index rows do not say "amended by" or "amends". The index Status column holds only `proposed`, `accepted` or `withdrawn`, plus a release tag such as `v1.5`. Background citations of other ADRs in Context or Consequences are fine.

*Why.* Pointer prose is a second, unchecked copy of the supersession data, and it drifts (the ADR-0011 index row was updated by hand; its body never could be). The `(supersedes …)` markers are the single source.

### 6. Only accepted ADRs count

A proposed or withdrawn ADR contributes no live clauses and supersedes nothing. A proposed ADR's supersessions are shown as pending.

*Why.* The compiled view must describe what is decided, not what is under discussion. Reporting pending supersessions shows what accepting an ADR would change before it is accepted.

## The compiled view

`tools/compile-decisions.py` reads §9 and writes [09a-current-decisions.md](./09a-current-decisions.md):

- for each accepted ADR, in order, its title and its **live** clauses with their IDs;
- a supersession table: each superseded clause, the clause that replaced it, and the superseded text for reference.

The page is generated and never hand-edited.

*Why compile instead of maintaining a "current decisions" page by hand.* A hand-written summary is a third copy of the decisions that nobody is forced to update. A compiled one is correct by construction: it cannot say anything the clauses don't.

*Why this makes ADRs easier to write.* An author writes only new clauses and names what they replace. They do not edit old ADRs, find every place a decision is restated, or update a summary; the compiler does the merging.

## The gate

A pre-commit hook runs `python3 tools/compile-decisions.py --check` whenever §9 or the compiled page changes. It fails when:

| Check | Why it matters |
|---|---|
| a clause ID is malformed, duplicated, or out of sequence | IDs are how changes refer to decisions; an ambiguous ID makes a supersession ambiguous |
| a supersede target does not exist | a change that replaces nothing is a typo or a lost decision |
| a target is not in an earlier ADR | the log is ordered; a clause cannot replace one written after it |
| a clause is superseded by two accepted ADRs | two live replacements for one decision is a conflict that needs a decision, not a silent winner |
| an ADR's index status disagrees with its own status line | the index is how readers find live ADRs |
| the compiled page differs from what the tool would write | the compiled page is only trustworthy if it is never stale |

*Why a gate rather than a convention.* Each rule is decidable from the text, so it can be checked on every commit. A convention depends on every author remembering it; the ADR-0011 index row shows how that fails.

## Alternatives considered

- **Edit ADRs in place, with git as the history and a gate requiring a revision note.** Each ADR always states the current decision, so there is no compiled view. Rejected: it weakens the log (the text an ADR was accepted with is only recoverable from git), and reviewing a change means reading a diff of decided text instead of a new, self-contained decision.
- **A hand-maintained "current decisions" page.** Rejected: an unchecked third copy that drifts, as described above.
- **Whole-ADR supersession** ("ADR-0017 supersedes ADR-0011"). Rejected: most changes replace one clause, and superseding a whole ADR would force restating everything else it decided.

## Where this is recorded

The mechanism is itself a decision: ADR-0018 supersedes ADR-0012's rule that a change is a new ADR, never an edit, with these clause-level rules. This page holds the style and the reasons in full; ADR-0018 holds the decision.
