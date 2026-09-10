# 14 — Redis degraded mode and bounded rebuild

**Type:** logic (test-first)
**Blocked by:** 10, 11, 13 — persistence safety and cache publication must exist.
**Status:** planned

## What this delivers

Redis outages switch the system to explicit persist-only mode; live aggregates pause without PostgreSQL polling and can later be rebuilt in a bounded operation.

## Lifecycle

- **Create:** Cache rebuild creates derived Redis state from a bounded raw-event range.
- **Read:** Runtime/read APIs expose `live` or `degraded` mode.
- **Update:** Recovery resumes publication after Redis health returns.
- **Delete:** Rebuild may replace derived cache keys only; raw events remain unchanged.
- **Undo:** A failed rebuild leaves degraded mode and is retried deliberately.

## Acceptance criteria

- [ ] Redis failure logs/metrics a mode transition and workers continue PostgreSQL persistence while it is healthy.
- [ ] Live read behavior reports degraded state and never auto-queries PostgreSQL.
- [ ] Rebuild scope, concurrency, and source time range are bounded so it cannot unboundedly impair ingestion.
- [ ] Recovery resumes cache publication and records a visible transition.

## Out of scope

- High-availability Redis, automatic long-range replay, or alert delivery.

## Verification

- Write outage/recovery/rebuild limit tests first using real dependency containers.

