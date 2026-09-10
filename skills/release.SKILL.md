---
name: release
description: Get reviewed, merged code into production per docs/release.md's configured tier — a near-no-op confirmation for the direct-to-prod default, or a driven staging-to-prod promotion with approval and safe migrations for projects configured for it. Use after /review merges a PR, or in batches per docs/release.md.
---

# Release

Steady-state skill. Runs after `/review` merges one or more PRs, per the
cadence `docs/release.md` declares (per-merge for the default tier,
batched for a staged tier).

## Purpose

Get code that has already passed `/review` safely into production,
matching whatever rigor the project actually needs — no more, no less.
The behavior here is entirely driven by `docs/release.md`, decided once
during `/alignment`. Do not add process this skill wasn't configured for.

## Input

`docs/release.md`. If it doesn't exist, stop — `/alignment` must define it
first, even if the answer is the default (direct-to-prod, no staging).

## Process

### Default tier (direct-to-prod on merge, no staging, no approval gate)

This is a thin confirmation pass, not a driven process — CI already
deploys on merge to main.

1. Confirm the merge to main triggered CI and the deploy succeeded.
2. Run a minimal post-deploy smoke check (the app responds, the deployed
   version matches what was just merged).
3. If it failed, that's an incident, not a normal release outcome — surface
   it immediately rather than retrying silently.

### Staged tier (staging environment, approval required, or both)

1. Confirm the merge landed on whatever integration/staging branch
   `docs/release.md` specifies, and that the staging deploy succeeded.
2. If `docs/release.md` requires stakeholder approval before promotion,
   request it explicitly (however it specifies — client sign-off, internal
   review) and stop here until granted. Do not promote to prod on an
   assumed approval.
3. Once approved, promote to production per the documented process.
4. Run migrations, if any are pending, using the expand/contract
   discipline (see below) — never as a blind schema sync.
5. Run the post-deploy smoke check against production.
6. Record what was released and when, if `docs/release.md`'s audit trail
   requirement calls for it.

## Migration safety floor — applies at every tier, no exceptions

Migrations must stay backward-compatible with whatever code is currently
running at the moment they execute:

- Add columns/tables as nullable or with defaults, never `NOT NULL`
  without one.
- Never rename or drop a column/table in the same deploy that removes the
  old code's dependency on it — expand first (add the new alongside the
  old), migrate usage over a subsequent deploy, then contract (remove the
  old) once nothing references it.
- If a migration cannot be made backward-compatible this way, that's a
  reason to stop and flag it, not to proceed anyway — this applies even
  on the simplest direct-to-prod project, since the cost of a bad
  migration doesn't scale down with process simplicity.

## Output

Code running in production, or an explicitly surfaced blocker (failed
smoke check, pending approval, unsafe migration) — never a silent partial
release.

## Explicit non-goals

- No code review — that already happened in `/review`.
- No scope/architecture decisions.
- Don't escalate a project's release process beyond what `docs/release.md`
  configures, even if a "more thorough" approach seems safer in the
  moment — if the rigor level feels wrong for the project, that's a
  reason to revisit `docs/release.md` explicitly, not to freelance a
  heavier process on one release.
