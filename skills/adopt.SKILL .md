---
name: adopt
description: One-time migration for a project that already has code and rougher/older docs (e.g. from an earlier version of this workflow) onto the current doc set and skill vocabulary. Reconciles old docs against actual code — code wins on conflict — marks shipped work as done, and hands off into steady-state /tickets from wherever the project currently stands. Run once, never as part of a fresh project.
---

# Adopt

One-time skill for an **existing, in-progress project** — never used on a
fresh project (that's what `/notes` → `/alignment` → `/foundation` →
`/design` is for). Use when a project already has real shipped code and
some rougher/older docs (possibly from an earlier, less-refined version of
this same workflow) and needs to move onto the current doc set and skill
vocabulary without re-planning or re-building anything already working.

## Why this exists, and why it isn't just "re-run `/alignment`"

Every bootstrap skill (`/notes`, `/alignment`, `/foundation`, `/design`)
assumes nothing exists yet. Re-running them against a live project would
produce docs and decisions disconnected from what's actually true on
disk, and could push `/foundation`-style scaffolding or `/design`-style
screen rebuilds that conflict with — or needlessly redo — real, working
patterns already established. This project needs reconciliation, not
fresh planning: old docs and real code both contain truth, but they may
disagree, and only one of them is actually running in production.

## Ground rule

**When old docs and actual code disagree, code wins.** Docs describe
intent; code is what's true. Every conflict gets resolved by updating the
doc to match the code — never the reverse — unless Edwinfred explicitly
says the code is wrong and should change (in which case that becomes a
`/fix` or `/add` ticket later, not a doc edit now).

## Process

1. **Inventory what exists.** Read every old doc the project has (notes,
   architecture, scope, conventions, or whatever equivalents the earlier
   workflow version produced), and read the actual codebase structure —
   folder layout, what features are implemented, what patterns are
   actually followed (not just documented), CI/lint/test setup, and
   deployment configuration as it actually runs today.

2. **Draft the current doc set from both sources**, code as tiebreaker:
   - `docs/architecture.md` — actual tech stack and system design as
     built, with rationale reconstructed where possible from old docs or
     inferred from the code itself
   - `docs/scope.md` — every feature already shipped is listed with
     `Status: done`, not re-scoped as if it were still planned. Features
     genuinely still pending go in as normal, with DoD defined the same
     way `/alignment` would define it
   - `docs/conventions.md` — document the patterns the codebase *actually*
     follows, even if they don't match this workflow's default
     recommendations (e.g. if the existing code is layer-based rather
     than feature-based, document that as the real convention rather than
     silently proposing a rewrite). Point to real existing files as the
     canonical examples, not a newly hand-crafted reference module —
     one likely already exists in the real code
     Where the codebase does something inconsistent across
     features, flag it — this doc should describe the intended standard
     going forward, and inconsistency is a decision point, not something
     to average away
   - `docs/release.md` — document the actual current deploy process,
     whatever it is, rather than defaulting to the standard template

3. **Reconcile conflicts with Edwinfred**, same interrogation pattern as
   `/alignment`, but framed as reconciliation, not fresh decisions: "the
   old docs say X, the code actually does Y — which should the new docs
   reflect?" Always propose a recommended resolution (usually: match the
   code, since it's what's live) rather than asking open-ended.

4. **Assess `/foundation`'s checklist against reality.** For each item
   (lint, typecheck, test runner, architecture-boundary rule, CI gate,
   env/secrets handling, seed data) — confirm it already exists, or gap-
   fill only what's missing. Do not replace or fight existing tooling
   choices that already work; add what's absent.

5. **Inventory existing UI as-is, rather than re-running `/design`.**
   Existing screens/components get cataloged into `docs/screens.md` and,
   if not already in Storybook, added as stories — marked `approved`
   retroactively, since they're already live and working. Do not redesign
   or re-critique existing screens against `references/ui-rules.md` as
   part of this skill — that would be re-litigating shipped work outside
   any ticket. The full `/design` rigor (theme discipline, per-screen
   loop, required states) applies only to *new* screens from this point
   forward.

6. **Set up `docs/bugs.md`, `docs/feature-requests.md`, and
   `docs/gotchas.md`** if they don't already exist, so `/fix` and `/add`
   have somewhere to write once adoption is complete. If the project
   already tracks bugs/requests elsewhere (an issue tracker), note that
   as the real intake location instead of creating a redundant file.

7. **Create or update root `CLAUDE.md`** to reference the finalized doc
   set, same as `/foundation` would for a fresh project.

8. **Confirm with Edwinfred** that the reconciled docs accurately reflect
   the project before treating adoption as complete.

## Output

The full current doc set (`architecture.md`, `scope.md`, `conventions.md`,
`release.md`, `design.md` if applicable, `screens.md`), accurately
reflecting the real state of the project — shipped work marked `done`,
real conventions documented as they actually are, `CLAUDE.md` wired up.
From here, the project continues exactly like any other: `/tickets` for
the next feature, `/fix`/`/add` for bugs and new requests, `/build` →
`/review` → `/release` as the ongoing loop.

## Explicit non-goals

- No re-planning or re-scoping already-shipped features.
- No rebuilding or re-critiquing existing UI — only new screens going
  forward get the full `/design` treatment.
- No forcing the codebase's existing conventions to match this workflow's
  defaults (e.g. feature-based folders) if it doesn't already follow
  them — document reality, don't retrofit an architecture change as a
  side effect of adopting new documentation. A structural migration, if
  ever wanted, is its own deliberate decision via `/add` or a dedicated
  ticket — never an implicit side effect of running `/adopt`.
- Run once. Do not re-run `/adopt` once the project is on the new doc set
  — subsequent work uses the normal skill set.
