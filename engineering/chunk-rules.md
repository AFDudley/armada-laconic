# Rules for the chunk briefs and their execution

Operator rulings, 2026-10-20.

1. **Start on the 5090.** Every implementation pebble starts on the 5090
   (exophial dispatcher, `max_workers: 1`). A worker that outgrows the window
   escalates to Sonnet through `contextPromotion`; the 5090 then takes the next
   pebble. At most 4 Sonnet sessions run at once, counting escalated workers and
   brief writers.
2. **The husband reviews.** No separate `reviewer` agent. The husband of each
   family reviews every chunk before it merges, and reviews the audit-critical
   chunks (P1, P2, N1, N2, G3) with extra care. Review findings become child
   pebbles in the family.
3. **Reuse the armada-nitro e2e work.** Acceptance for every chunk, in every
   repo, uses the armada-nitro environment and gates, not a new harness:
   - the laconic-so fixturenet and reconciler
     (`armada-nitro` `docs/ENVIRONMENT.md`, briefs E1 to E9 in `docs/tasks/`);
   - the gates under `armada-nitro/env/gates/`: `nitro-e2e`, `pool-e2e`,
     `watchtower`, `hub-custodian-frost`.
   A chunk that changes the pool, the circuits, governance or the SDK extends
   the existing reconciler and gates (for example a new step in `pool-e2e`, or a
   new gate directory beside them) rather than building a separate fixturenet.
   The gate chunks G7, N7, N8 and C5 are extensions of these gates.
4. **Fixturenet first, then Sepolia, then mainnet.**
5. **Estimates** use `ESTIMATION.md` in this directory: points, chunks of at most
   9 points for the 5090.
6. **The handoff is the live state.** `armada-laconic/notes/handoff-2026-10-20.md`
   §0 records the state of every family, brief, pebble and gate. The lead
   updates it and commits it to `main` whenever any of them changes. Husbands
   report state changes to the lead; they do not edit the handoff. Every brief
   names the handoff row it moves (family and chunk id).
