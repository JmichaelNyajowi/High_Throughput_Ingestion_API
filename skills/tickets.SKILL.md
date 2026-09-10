---
name: tickets
description: Cut one feature (mapped to a scope.md feature area) into demoable, independently reviewable tickets with explicit dependencies. Runs once per feature, just-in-time before that feature is built — not once for the whole project. Also invoked by /fix and /add.
---

# Tickets

Bridges bootstrap and steady state — it **runs per-feature, just-in-time**,
not once for the entire project. First invocation happened right after
`/design`, scoped to the first feature to be built; that first invocation
was the last strictly-bootstrap step. Every invocation since — before
each subsequent feature, and on behalf of `/fix`/`/add` — is steady-state
behavior, even though the mechanism is identical each time.

## Why per-feature, not all-at-once

Cutting every ticket for the whole project upfront means later tickets are
written against assumptions that may be stale by the time they're
actually built — real patterns and friction only become visible once
earlier features are actually implemented. Cutting each feature's tickets
just before that feature is built lets later cutting benefit from what
actually exists in the codebase by then, not just what was planned during
`/design`/`/alignment`.

## Purpose

Decompose one feature into tickets small enough to implement, verify, and
review as one coherent, standalone unit — so `/build` never has to guess
scope, and a bad agent run only costs one ticket's worth of work.

## Input

`docs/scope.md`, `docs/architecture.md`, `docs/conventions.md`, and
`docs/screens.md`, `docs/roadmap.md` . Also: the current state of the codebase, if this is not
the first feature — read what's already built before cutting the next
feature's tickets, so they reflect real patterns already in use.

## Concepts

**Feature** — a deployable, demoable milestone. Maps directly to one
feature area in `docs/scope.md`'s definition-of-done.

**Ticket** — one vertical slice of a feature: a coherent piece of behavior
cutting through whatever layers are relevant (schema, API, UI) for that
slice. A feature is built across multiple tickets; no single ticket claims
an entire feature. Example, for canteen operations: one ticket for
recording daily takings, one for weekly count-derived item detail, one
for receiving transferred/direct stock, one for canteen credit sales.

## Ticket sizing

No token-based size ceiling — model context windows are large enough that
this is no longer the binding constraint. Size a ticket by demoability
instead: **can "done" be described in one or two sentences, and can it be
demoed standalone?** If describing a ticket needs "and also" more than
once, split it. The reasons this still matters: a ticket that's too large
stops being independently reviewable, a bad agent run on an oversized
ticket loses more work, and oversized tickets are more likely to conflict
with other in-flight tickets at merge time.

## Cross-feature screens (dashboards, composite views)

Composite screens that pull data from multiple features were already
fully designed in `/design` against mock data in every slot, including
each slot's empty/loading state, and are already routed/hosted by the
shell (shipped, not cut per-feature). There is no separate "shell
ticket" per feature. Instead: **each
feature ticket that owns a given piece of data is responsible for wiring
that data into every screen slot that depends on it**, wherever in the app
those slots are — not just that feature's own dedicated screens. Until a
given slot's owning ticket lands, that slot correctly renders its
already-designed empty/loading state, because the real data simply isn't
wired in yet and the shell already put that state there by default.
Note this explicitly in any ticket that fills a slot on a screen belonging
to a different feature (e.g. "ticket 23 also wires the Recent Orders
widget on the dashboard route").

## Type and test-first classification

Each ticket is marked with a `**Type:**` line when cut, describing both
what kind of work it is and whether it's test-first — the two established
values in use are `logic (test-first)` and `plumbing (test-after)`. Use
`logic (test-first)` for tickets with real business logic: validation
rules, calculations, state transitions, branching behavior. Use `plumbing
(test-after)` for tickets that are mostly composition or wiring with no
ambiguous behavior to specify — pure UI assembly from already-approved
Storybook stories, thin CRUD wrappers, glue code. A single vertical-slice
ticket can reasonably contain both; if so, mark it `logic (test-first)`
and scope the acceptance criteria so the test-first discipline applies to
the logic portion, not the UI composition portion.

The reasoning: writing the test first, before any implementation exists,
forces the spec (inputs, outputs, edge cases) to be nailed down
independent of an implementation — and specifically prevents an agent
from writing tests that just describe what its own code already does,
which validates bugs as readily as it catches them. That value is real
for logic and close to zero for pure composition, where `/review` against
`references/ui-rules.md` and the approved Storybook story is already the
correctness check.

## Process

1. **Pick the next feature** from `docs/scope.md`/`docs/roadmap.md` to cut
   tickets for — Edwinfred confirms which, if not already obvious from
   priority/roadmap order.

2. **Cut vertical-slice tickets for that feature only**, continuing the
   existing flat numbering in `.work/` (next unused number). One ticket's
   changes should live almost entirely inside one feature's module, except
   for the cross-feature screen-wiring case above.

3. **Declare dependencies explicitly** via each ticket's `**Blocked by:**`
   line — other ticket numbers, or "None" with a one-line reason (see
   existing tickets in `.work/` for the pattern). This is what lets
   `/build` know what's safe to run concurrently versus what must wait.

4. **Write each ticket using the fixed template below.** Tickets are read
   cold by agents with no conversation history — ambiguity here becomes a
   wrong implementation later, not a quick clarifying question.

5. **Confirm this feature's ticket set with Edwinfred** before `/build`
   starts on it — a last checkpoint to catch a missing or wrongly-scoped
   ticket while it's still cheap to fix. This confirmation is per-feature,
   not a single whole-project gate.

## Ticket template

Each ticket is one file: `.work/<NN>-<slug>.md`, continuing the existing
flat number sequence (no per-feature subfolder).

`**Status:**` is mandatory on every ticket cut from now on — it's what
`/build` scans for to claim work and what `/review` and `/fix` use to
track a ticket through the pipeline. Set it to `planned` when cut.

```markdown
# <NN> — <short title>

**Type:** logic (test-first) | plumbing (test-after)
**Blocked by:** <ticket numbers, or "None" with a one-line reason>
**Status:** planned

## Goal
One sentence: what this ticket makes true that wasn't true before.

## Context
Pointers, not explanations — the agent reads these, not a summary of them:
- Relevant module(s): `src/modules/<x>/`
- Relevant docs sections: `docs/conventions.md#<section>`, `docs/architecture.md#<section>`
- Relevant Storybook stories / screens from `docs/screens.md`: <names>
- Relevant prior tickets this builds on: <ticket numbers>

## Scope
**In:** explicit list of what this ticket implements.
**Out:** explicit list of what it deliberately does not — prevents
over-building into a neighboring ticket's territory.

## Acceptance criteria
Concrete and checkable — this is "demoable" made literal. What should be
true, visible, and testable when this is done. Include any screen slots
(per the cross-feature note above) this ticket is responsible for wiring.
If **Type: logic (test-first)**, write these as criteria that translate
directly into executable test cases — `/build` will write them as failing
tests before implementing.

## Verification
What the implementing agent must run/check before marking this done:
tests to write/pass, lint/typecheck, manual check against
`references/ui-rules.md` if UI is touched.
```

## Output

`.work/<NN>-<slug>.md` for every ticket in the feature just cut, each with
`Status: planned`, dependencies declared via `Blocked by`, confirmed with
Edwinfred.

## Explicit non-goals

- No implementation — that's `/build`.
- Don't cut tickets for features beyond the one currently being started —
  that's next invocation's job, once this feature is far enough along.
- When re-entered by `/fix` or `/add`, only cut the new tickets/feature
  those skills hand off — don't re-derive the whole roadmap.
