# 10 — Idempotent PostgreSQL bulk persistence

**Type:** logic (test-first)
**Blocked by:** 08, 09 — admitted batches and worker ownership must exist.
**Status:** in-progress

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

## Review

Not approved. The required PostgreSQL integration tests for transactional inserts, `ON CONFLICT` idempotency, 1,000-record flushing, and two-second flushing have not been added or run. Docker is unavailable in this environment (`docker: command not found`), so the specified Testcontainers evidence cannot be substituted with the passing unit/race suite.

The current batcher also discards a failed flush result, exposes no completed flush metrics, is not wired from shard workers, and permits a second `Shutdown` call to panic by closing its input twice. These are blocking gaps for the ticket's durability, observability, and lifecycle acceptance criteria. The ticket remains `in-progress`.
