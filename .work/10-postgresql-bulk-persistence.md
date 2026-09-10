# 10 — Idempotent PostgreSQL bulk persistence

**Type:** logic (test-first)
**Blocked by:** 08, 09 — admitted batches and worker ownership must exist.
**Status:** planned

## What this delivers

Accepted raw telemetry is flushed to PostgreSQL in idempotent parameterized transactions every 1,000 records or two seconds.

## Lifecycle

- **Create:** Accepted events become immutable raw-event rows.
- **Read:** Persistence returns flush outcome and duplicate count to processing/metrics only.
- **Update:** Raw events are never edited.
- **Delete:** Raw events are never deleted in MVP.
- **Undo:** Duplicate retries are neutralized by `ON CONFLICT DO NOTHING`, not undone.

## Acceptance criteria

- [ ] Bounded persistence input flushes on 1,000 records or two seconds, whichever occurs first.
- [ ] Inserts are parameterized, transactional, and use `ON CONFLICT (device_id, event_id) DO NOTHING`.
- [ ] One retry storm creates one durable row per `(device_id,event_id)` without aggregate-side duplication failure.
- [ ] Flush size/duration/outcome/conflict count are observable.

## Out of scope

- Retry/backpressure policy, historical read API, retention jobs, or COGS-like derived logic.

## Verification

- Write integration tests against PostgreSQL before implementation, including timer and size triggers.

