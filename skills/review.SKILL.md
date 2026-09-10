---
name: review
description: Independently review a ticket's PR — fresh session, no memory of implementing it — against the ticket's acceptance criteria, docs/conventions.md, docs/architecture.md, and references/ui-rules.md if UI changed. Approve and merge, or reject with specific findings back to /build.
---

# Review

Runs after `/build` opens a PR for a ticket, before merge. Must run as a
fresh session with no memory of implementing the ticket — the entire point
of this gate is catching what the implementing agent is structurally
blind to (its own assumptions, its own blind spots), which a continuation
of the same session cannot do.

## Purpose

Independent verification that a ticket's implementation actually matches
its intent and the project's standards, before it reaches main. This is
the gate that keeps self-verification (mechanical, done in `/build`) from
being the only check on quality (judgment-based, needed here).

## Input

The PR opened by `/build`, the ticket file it references (`.work/<n>-*.md`),
`docs/conventions.md`, `docs/architecture.md`, `CLAUDE.md`, and
`references/ui-rules.md` if the diff touches UI.

## Process

1. **Read the ticket first**, before looking at the diff — form an
   independent expectation of what the PR should contain.

2. **Read the diff.** Check, in order:
   - Does it satisfy every item in the ticket's Acceptance Criteria?
   - Does it stay within the ticket's declared Scope (in), and avoid
     drifting into its declared Out-of-Scope?
   - Does it follow `docs/conventions.md` and `CLAUDE.md`'s Module/Code
     rules, concretely:
     - Cross-module imports go through the target module's `index.ts`
       only — grep the diff for any `from ".../queries"`, `.../logic`,
       or `.../schema"` import reaching into another module.
     - `queries.ts` stays bare Prisma calls; permission/validation logic
       lives in `logic.ts`, not in `routes.ts` or `ui/`.
     - Logic functions return `{ ok: true, ... } | { ok: false, reason }`
       rather than throwing for expected failure; routes return
       `Response.json({ error: reason }, { status })`.
     - Every location-scoped read/write calls `canAccessLocation()` from
       `people/index.ts` — flag any hand-rolled location comparison.
     - No new top-level folder, no new module beyond the fixed six
       (`catalogue`, `stock`, `sales`, `cash`, `people`, `reporting`).
   - Does it fit `docs/architecture.md`'s intended design (module seams,
     the location-cutting dimension), or does it introduce an
     inconsistent pattern?
   - If UI changed: run it against `references/ui-rules.md` directly,
     item by item that's relevant to what changed, plus `CLAUDE.md`'s UI
     rules — composed only from `components/ui/`/`components/patterns/`,
     opens inside a layout shell, theme tokens only (no raw hex/arbitrary
     values), one accent element, and empty/loading/error/permission-denied
     states from `components/patterns/states.tsx` on every list/table.
     Every new component needs a Storybook story — check it was added to
     `docs/screens.md`.
   - Correctness: real bugs, not just style — edge cases from
     `references/ui-rules.md`'s "Edge cases" section if applicable
     (empty, one item, many items, long text, nulls, extreme numbers).
   - Tests: integration tests hit the real test DB (`TEST_DATABASE_URL`),
     never mock this codebase's own modules; new logic with permission or
     calculation rules should be test-first per `docs/conventions.md`'s
     Testing section — flag plumbing-only tickets that added logic tests
     after the fact as a smell, not a blocker on its own.
   - Reuse/simplification: is there duplicated logic that should reuse an
     existing pattern instead (per the rule-of-three from
     `docs/conventions.md`)?

3. **Decide.**
   - **Approve:** merge the PR, set the ticket's `Status: done`.
   - **Reject:** do not merge. Write specific, concrete findings — file,
     line, what's wrong, what would fix it — not vague direction. Set the
     ticket's `Status: in-progress` and hand back to `/build`.

Do not split the difference — a ticket is either done or it isn't. Minor
nitpicks that don't affect correctness, scope, or convention adherence can
be noted without blocking approval; anything that does affect those must
block.

## Output

Either a merged PR with the ticket marked `done`, or a rejected PR with
specific findings and the ticket returned to `/build`.

## Explicit non-goals

- Not a place to change the ticket's scope — if the review reveals the
  ticket itself was wrong or incomplete, that's a finding to raise with
  Edwinfred (possibly via a new ticket from `/add`), not something to
  silently expand this review into.
- Not a re-run of mechanical checks `/build` already did (tests, lint,
  typecheck) unless there's specific reason to doubt they were run
  correctly — this gate's value is in judgment checks, not duplicating CI.
