---
name: foundation
description: Scaffold the repository, tooling, CI, environment, seed data, and hand-crafted reference modules so every later phase (especially unattended agent runs) can proceed with zero setup friction. Use once, after /alignment is confirmed complete, before /design.
---

# Foundation

Bootstrap phase 3 of 6. Run once, after `/alignment` is confirmed complete,
before `/design`.

## Purpose

Set up everything a codebase needs to exist *before* any real feature or
design work begins, so that every later phase — especially autonomous or
looped agent runs during `/build` — can proceed without ever needing to
make ad hoc setup decisions. This phase trades a bit of upfront time for
removing friction (and inconsistency) from every ticket that follows.

## Input

`docs/architecture.md`, `docs/scope.md`, `docs/conventions.md` from
`/alignment`. If these don't exist or alignment wasn't confirmed complete,
stop and tell Edwinfred to finish `/alignment` first.

## Process

1. **Scaffold the repository.**
   - `git init`, initial commit conventions, `.gitignore`.
   - Folder structure per `docs/conventions.md` (feature-based by default:
     `src/features/`, `src/shared/`, `src/core/`).
   - Package manager / dependency setup per `docs/architecture.md`'s
     chosen stack.

2. **Set up environment and secrets handling.**
   - `.env.example` committed, listing every required variable with a
     placeholder or dummy value and a one-line comment on what it's for.
   - `.env` gitignored.
   - Document required env vars in `docs/setup.md` (create this file).
   - Hard rule, stated here and reinforced in `CLAUDE.md`: agents never
     read, write, or print real secret values — only ever reference
     variable *names*.

3. **Set up the database, if applicable.**
   - Schema, migration tooling, and the first migration(s) matching
     `docs/architecture.md`.
   - A seed script producing a small, realistic dataset covering the
     project's core entities — enough that any later ticket can run the
     app locally and see real-ish data, not empty tables. This removes
     the need for every future ticket to invent its own throwaway test
     data.

4. **Wire CI/CD and machine-enforced quality gates.** This is the
   mechanism that makes `docs/conventions.md` actually enforced rather
   than aspirational — do not treat this as optional or "nice to have."
   Set up, wired into CI and (where practical) pre-commit:
   - linter + formatter
   - type checker, if the stack has one
   - test runner, wired to a standard command (`npm test` or equivalent)
   - an architecture-boundary rule that mechanically blocks deep
     cross-feature imports (e.g. `dependency-cruiser`, an ESLint
     boundaries plugin, or the stack's equivalent) — this enforces the
     `shared/`/`core/` and cross-feature-import rules from
     `docs/conventions.md`
   - a CI pipeline that runs all of the above on every push/PR, and
     fails the build if any of them fail

5. **Set up design/component tooling**, if the project has a UI — Tailwind,
   shadcn/ui (`npx shadcn init`), Storybook, component folder conventions.
   Scaffold an empty Tailwind v4 `@theme` block (structure only — no real
   token values yet) so `/design` has a concrete file to fill in rather
   than creating the theme from scratch. This step is scaffolding only;
   no actual screens, token values, or design decisions happen in this
   phase — that's `/design`.

6. **Set up Docker / infra tooling**, if applicable to the project.

7. **Build the hand-crafted reference module(s).** Pick one or two
   representative modules (either the first real feature, or a
   deliberately minimal example module if no real feature is simple
   enough to go first) and build them with extra care and scrutiny —
   this is slower than normal ticket-speed work, on purpose. These
   modules become the canonical example every future ticket and every
   pointer in `docs/conventions.md` refers back to. Once built, go back
   and fill in the placeholder pointers in `docs/conventions.md` with
   real paths into these modules.

8. **Create the root `CLAUDE.md`.** It must explicitly reference, by
   path, with a one-line note on when to consult each:
   - `docs/architecture.md`
   - `docs/scope.md`
   - `docs/conventions.md`
   - `docs/release.md`
   - `docs/setup.md`

   This ensures every future agent session — Claude Code, Codex, or
   Antigravity — picks up full project context automatically, without
   Edwinfred needing to paste paths in manually.

9. **Verify the whole thing works end to end.** The strong default check:
   a fresh clone can install dependencies, copy `.env.example` to `.env`,
   run migrations and seed data, start the app locally, and pass CI
   (lint/typecheck/test) — all via one documented command sequence in
   `docs/setup.md` or the root `README.md`. For project types where this
   exact sequence doesn't cleanly apply (e.g. no database, static site),
   use judgment to define the equivalent "this proves the foundation
   works" check, but do define and run one — don't skip verification
   entirely.

## Output

A fully scaffolded, running repository: tooling, CI, env/secrets handling,
seed data, hand-crafted reference module(s), and a `CLAUDE.md` wiring
everything together. `docs/conventions.md` updated with real pointers into
the reference module(s). `docs/setup.md` created.

## Explicit non-goals

- No UI/UX or design-system decisions, no screens — that's `/design`.
- No feature work beyond the reference module(s) needed as a pattern
  example.
- Don't skip the quality-gate wiring even under time pressure — this is
  the mechanism that keeps later unattended agent runs safe.
