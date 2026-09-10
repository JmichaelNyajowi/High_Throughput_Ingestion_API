---
name: fix
description: Triage a logged bug from docs/bugs.md, ask Edwinfred whether it needs the full ticket pipeline or a direct fix, then either hand off to /tickets->/build->/review->/release or fix and ship directly. Steady-state skill, invoked any time bugs need addressing.
---

# Fix

Steady-state skill. Invoked whenever there are bugs in `docs/bugs.md` to
address — either on demand or in a batch.

## Purpose

Get logged bugs resolved with a level of process matched to their actual
risk — not maximal ceremony for a typo, not a silent unreviewed patch for
something that touches real logic.

## Input

`docs/bugs.md` (or `docs/bugs/<id>-<slug>.md` if it's been split out due to
volume). If it doesn't exist or is empty, there's nothing to do.

## Bug entry format

Each entry in `docs/bugs.md`:

```markdown
## BUG-<NN>: <short title>
**Severity:** critical | high | normal | low
**Discovered:** <how — client report, production error, manual testing>
**Status:** open | in-progress | fixed

### Description
What's broken.

### Repro steps
How to reproduce it reliably.
```

## Process

1. **Triage by severity.** Critical/high severity bugs are addressed
   immediately, ahead of any batch. Normal/low severity bugs may be
   processed together when this skill is invoked in batch mode.

2. **For each bug being addressed, ask Edwinfred how to route it** —
   do not assume. Give a recommended default based on the bug's apparent
   complexity and blast radius:
   - **Recommend the full ticket pipeline** (`/tickets` → `/build` →
     `/review` → `/release`) when the fix touches business logic, auth,
     data integrity, or anything with real complexity — the independent
     review matters most exactly where a rushed fix could cause a
     regression.
   - **Recommend a direct fix** (implement, self-verify, ship, no separate
     ticket/review cycle) only for trivial, obviously-contained changes —
     a typo, a copy fix, an off-by-one in a clearly isolated spot.
   
   Edwinfred approves the recommendation or overrides it per bug.

3. **If routed through the ticket pipeline:** create a ticket using
   `/tickets`' template, feature = the existing feature folder the bug
   belongs to (or a `bugfixes` catch-all if it's genuinely cross-cutting).
   The ticket's acceptance criteria must include a regression test that
   reproduces the original bug — this is required, not optional, whenever
   this path is taken. Hand off to `/build` as normal.

4. **If routed as a direct fix:** implement the fix, write a regression
   test reproducing the bug if practical, self-verify (tests, lint,
   typecheck, `references/ui-rules.md` if UI is touched), and ship per
   `docs/release.md`'s configured tier. Still requires the bug entry to be
   updated to `Status: fixed` — this path skips ticket/review ceremony, it
   does not skip verification.

5. **If the root cause was non-obvious** — something a future agent could
   plausibly repeat — append a short entry to `docs/gotchas.md`: what went
   wrong, why, how to avoid it. Keep it terse; this file's value is in
   being skimmable.

6. **Update `docs/bugs.md`**, marking resolved entries `Status: fixed`.

## Output

Bugs resolved and shipped (per whichever route was chosen), `docs/bugs.md`
updated, `docs/gotchas.md` appended where a real lesson was learned.

## Explicit non-goals

- Never skip `/review` for anything routed through the ticket pipeline,
  regardless of urgency — critical severity changes priority, not rigor.
- Don't silently choose the routing without asking — severity and
  complexity inform the recommendation, but the decision is Edwinfred's.
