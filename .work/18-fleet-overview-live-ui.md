# 18 — Redis-backed Fleet overview

**Type:** plumbing (test-after)
**Blocked by:** 01, 16 — shell destinations and fleet live API must exist.
**Status:** planned

## What this delivers

Operators can triage the fleet in the approved dense table using Redis-backed snapshots, status filters, sorting, and truthful live/stale/degraded states.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Dashboard polling reads only the fleet live API.
- **Update:** Operators update local sort/search/status-filter state.
- **Delete:** Not applicable.
- **Undo:** Clearing a local filter restores the unfiltered view.

## Acceptance criteria

- [ ] Default Fleet route renders status strip, searchable/filterable/sortable 40 px-row table, freshness, status wording/icon, and current measurement values.
- [ ] Loading uses approved shaped skeletons; empty/no-results/error/degraded/stale states match `docs/DESIGN.md` and preserve the last successful snapshot when appropriate.
- [ ] Device rows are real links, numeric columns use tabular figures, sortable headers use buttons and `aria-sort`.
- [ ] Polling is 2–5 seconds through typed native-fetch/TanStack Query code and never calls the history endpoint.

## Out of scope

- Device provisioning, alert delivery, bulk actions, chart redesign, or PostgreSQL fallback.

## Verification

- Run component tests, browser keyboard/responsive checks, and `/verify` evidence for live/degraded state transitions.

