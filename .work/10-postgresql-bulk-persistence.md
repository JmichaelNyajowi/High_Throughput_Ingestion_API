# 10 — Idempotent PostgreSQL bulk persistence

**Type:** logic (test-first)
**Blocked by:** 08, 09 — admitted batches and worker ownership must exist.
**Status:** done

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

## Build-gate trace

| Contract | Source | Implementation | Repeatable proof |
| --- | --- | --- | --- |
| Flush on 1,000 events or two seconds | PRD 6.7; Architecture 3.3 | `api/internal/telemetry/persistence.go`: `Batcher` | `TestBatcherFlushesToPostgreSQLOnTimerAndSize` uses real PostgreSQL; unit tests cover both triggers deterministically. |
| Parameterized transactional, idempotent inserts | PRD 6.7; `docs/database.md` | `PostgresEventRepository.Flush` | `TestPostgresEventRepositoryFlushesTransactionallyAndIgnoresDuplicateEvents` verifies one durable row from a duplicate retry. |
| Immutable raw-event lifecycle and database ownership | Architecture 2; `docs/database.md` | `EventRepository` seam and PostgreSQL repository | PostgreSQL integration tests run against `migrations/000001_init.up.sql`. |
| Observable flush outcome, conflict count, size, and duration | PRD 6.9; Architecture 3.3 | `PersistenceMetrics` and `Batcher.Metrics` | Size, conflicts, failures, and non-zero duration asserted in unit/integration tests. Prometheus export is Ticket 15. |
| Safe final flush and no close/send panic | PRD 6.4; Architecture 3.3 | `Batcher.Shutdown` and guarded `Submit` | `TestBatcherFlushesAtSizeAndShutdownIsIdempotent`; race suite passes. |
| Real data-layer verification | `docs/conventions.md` | PostgreSQL Testcontainers tests | `go test -count=1 ./internal/telemetry` with Docker access passes. |

Quality evidence (2026-09-11):

- `go test -count=1 ./internal/telemetry` — passed with Docker-backed PostgreSQL.
- `go test -count=1 ./internal/ingestion ./internal/ingestion/credentials` — passed with Docker-backed PostgreSQL.
- `go test -count=1 -race ./...`, `go vet ./...`, `gofmt` check, and `git diff --check` — passed.
- After the final flush-observability correction: `go test -count=1 -race ./internal/telemetry`, `go test -count=1 ./internal/telemetry`, `go vet ./...`, formatting, and diff checks — passed.

## Review

Approved on 2026-09-11. Docker is available for verification through the approved execution environment. PostgreSQL integration tests cover parameterized transactional inserts, duplicate conflicts, timer flushing, and 1,000-record flushing. `PersistenceMetrics` exposes completed successful and failed flush outcomes and duration. `Shutdown` is idempotent, while concurrent or new submissions are rejected safely. Retry/backpressure policy remains deliberately deferred to Ticket 11, as stated in this ticket's out-of-scope boundary.
