---
name: add
description: Scope a new, previously unplanned feature request against the existing architecture.md/scope.md, run a scoped /design pass if it has a UI, then hand off to /tickets. Steady-state skill for scope growth after initial build — clients almost always ask for more than v1.
---

# Add

Steady-state skill. Invoked whenever a new feature request arrives that
wasn't part of the original `docs/scope.md`.

## Purpose

Bring a new feature into the project with the same rigor `/alignment`
applied to v1 — reflected understanding, gap interrogation, explicit
conflict-checking against existing decisions — scaled to the size of one
feature rather than a whole project, and without silently expanding scope
that was already agreed and possibly priced with a client.

## Input

`docs/feature-requests.md` — the intake log. Any new request gets logged
here first (what was asked, by whom, when, initial notes) so nothing said
in passing gets lost before this skill actually runs. Also reads the full
existing `docs/architecture.md`, `docs/scope.md`, and `docs/conventions.md`
as fixed context — this is not a blank-slate planning session.

## Process

1. **Read the request from `docs/feature-requests.md`**, plus the existing
   `docs/architecture.md` and `docs/scope.md` in full.

2. **Reflect back understanding**, same as `/alignment` does for a whole
   project, but scoped to this one feature: what it is, who it's for, how
   it fits into the existing product.

3. **Check explicitly for conflicts** with what's already decided — does
   this feature contradict an existing architectural choice, strain an
   existing module boundary, or imply a scale/requirement the current
   architecture wasn't built for? This check is the main reason ADD
   reuses `/alignment`'s pattern instead of just documenting the request
   and moving on — a new feature added without this check is how scope
   quietly breaks an architecture that was fine for v1.

4. **Interrogate for gaps**, same pattern as `/alignment`: ask, always with
   a recommended default, until the feature's scope and acceptance
   criteria are as clear as any v1 feature's were.

5. **Flag architecture/conventions changes explicitly.** If this feature
   needs something genuinely new — a new dependency, a pattern not used
   before (e.g. the first real-time feature requiring websockets) — do
   not quietly amend `docs/architecture.md` or `docs/conventions.md`.
   Surface it to Edwinfred as a deliberate decision, with the same
   architect-persona scrutiny `/alignment` used originally, before writing
   it into either doc.

6. **Determine if the feature has a UI surface.** If yes, hand off to a
   *scoped* `/design` pass: new screens only, reusing the existing
   finalized theme and established page-template patterns from
   `references/design-principles.md`/`references/ui-rules.md` — do not
   re-open token or pattern decisions that are already settled. If the
   feature is backend-only, skip design and go straight to tickets.

7. **Append to `docs/scope.md`** as a new, dated feature-area entry with
   its own definition-of-done — tagged clearly as added post-v1. Never
   edit the original v1 entries; the original scope stays intact as a
   record of what was actually agreed and (if relevant) priced at the
   start. This also gives Edwinfred a clean, auditable trail of what's
   grown beyond the original proposal.

8. **Hand off to `/tickets`** to cut the feature into tickets, exactly as
   any v1 feature would be.

9. **Update `docs/feature-requests.md`**, marking the request as scoped
   and pointing to the new `scope.md` entry.

## Output

A new feature area appended to `docs/scope.md`, new screens in Storybook
if applicable, and a set of tickets in `docs/tickets/<feature>/` ready for
`/build`.

## Explicit non-goals

- No implementation — that's `/build`, reached via `/tickets`.
- Never silently edit existing `scope.md` entries, `architecture.md`, or
  `conventions.md` — additions and architectural changes are both always
  explicit, flagged decisions, never quiet amendments.
- Don't skip the conflict-check step even for a feature that seems small
  — small features are exactly the ones likely to be assumed compatible
  without actually being checked.
