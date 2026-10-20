# Estimation method: points and 5090 hours

estimation method · 2026-10-20

This is the estimation method for armada-laconic. The plan
([`build-plan.md`](./build-plan.md)) states every unit of work in these points and
converts them to hours on the 5090 with these rules. It replaces the old
"linear difficulty points" (1–10) scale; that scale and its numbers are gone from
this doc set.

Source: session `01a150cb`, 2026-10-19 03:14 to 03:31 UTC, recorded in
`labbook/2026-10-18-contract-owned-notes-and-zk-channel-apps.md` §2 ("Context
scale"), §7 and §8.

## 1. The unit

One **point** is 10,000 tokens of peak context that one worker must hold at one
moment (`exophial/src/exophial/estimate.py:199`, `TOKENS_PER_POINT = 10_000`;
exophial issue exo-231a.1). Points measure the size of a task, not its
difficulty, and not cumulative token spend.

A **chunk** is one unit of work that one worker does in one session: one brief,
one branch, one acceptance check. Every plan item is a chunk or a list of chunks.

## 2. Size limits

- The 27B model on the 5090 holds 163,200 tokens, so the exophial dispatcher
  refuses any pebble above 16 points on the 5090. marks holds 262,144 tokens,
  so 26 points there.
- A worker starts at about 86,000 tokens before its first tool call (one
  measured cloud worker, session `01a0fd64`, 2026-10-02). On the 5090 that leaves
  about 8 points of real work.
- So a chunk for the local route is at most **9 points**. A larger chunk is split,
  or starts on Sonnet.
- No chunk is above 12 points on any route.

Anchors from the MetaLeX plan: a bootstrap chunk is 4 points; a derivation with
test vectors is 7; a funding flow with a privacy assertion is 9; a deploy module
is 10.

## 3. From points to hours

- Decode is not the limit. The 5090 decodes 45 tok/s on average under real agent
  load (vLLM lifetime metrics, session `01a0deb8`, 2026-09-27), and 116.7 to 138
  tok/s warm (`~/qwen3.8-27b-vllm-omp-setup.md` lines 216-242). One measured
  worker wrote about 1,100 output tokens per point (session `01a0fd64`). On
  decode alone, 30 minutes covers about 70 points.
- Turn overhead is the limit. A pilot pebble on marks took 923 s over 62 turns,
  about 15 s per turn, mostly tool execution, prefill and the gate (session
  `01a0bd0d`, 2026-09-25).
- Rule: **one 30-minute session completes one chunk of about 10 points**, or two
  chunks if both are small (about 5 points or less).
- Hours for a plan = (number of chunks, counting two small chunks as one) × 0.5
  hours, before rework. Example: the 209-point list has about 27 chunks, so about
  27 sessions, or 13 to 14 hours of worker time.
- Parallel hosts shorten the calendar time, not the worker hours. The
  dependency graph decides how many chunks can run at once.

## 4. How the plan uses it

- Every stage lists its chunks. Each chunk has an id, a one-line deliverable, its
  dependencies, its points, and its route (5090 if at most 9 points; otherwise
  split, or Sonnet).
- Each stage gives its total points, its session count and its hours by §3.
- The plan ends with the totals: points, sessions, 5090 hours, and the critical
  path in sessions.
- Circuit work and the critical-path chunks P1 and N5 carry a routing note: circom
  on the 27B model is not measured (lab book §8), so these are the chunks most
  likely to escalate to Sonnet through `contextPromotion`. Every chunk still
  starts on the 5090 (operator ruling, CHUNK-RULES.md rule 1); there is no
  separate plan-time Sonnet assignment.

## 5. Limits of the method

- The output-per-point figure and the 15 s turn come from one sample each.
- Rework, review and gate failures are not in the hours.
- Mark any estimate that rests on these samples as [INFERENCE] where it matters.
