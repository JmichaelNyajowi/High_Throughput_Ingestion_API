# 11 — Persistence retry and admission backpressure

**Type:** logic (test-first)
**Blocked by:** 09, 10 — admission and durable-write seam must exist.
**Status:** done

## What this delivers

PostgreSQL failure is isolated behind a capped retry buffer; once safety limits are reached, new admission returns `503` instead of consuming unbounded memory.

## Lifecycle

- **Create:** Failed flushes enter the bounded retry buffer.
- **Read:** Admission reads the explicit backpressure state.
- **Update:** Retry timing advances from 100 ms exponentially, capped at 5 s and three attempts.
- **Delete:** Exhausted retry entries are dropped only with explicit structured loss logging.
- **Undo:** Not available; clients may retry using stable event IDs.

## Acceptance criteria

- [ ] Failed persistence retries at the required schedule, no more than three attempts.
- [ ] Retry capacity is 10,000 events and its occupancy is observable.
- [ ] Retry exhaustion or buffer saturation activates admission backpressure; subsequent valid batches receive `503` before enqueue.
- [ ] Redis availability never masks a PostgreSQL backpressure state; memory remains bounded.

## Out of scope

- External brokers, disk spool, exactly-once delivery, or automatic data recovery beyond client retry.

## Verification

- Write failure/backoff/buffer tests first; prove bounded behavior with PostgreSQL outage integration tests.

## Build-gate trace

| Contract | Source | Implementation | Repeatable proof |
| --- | --- | --- | --- |
| Failed PostgreSQL flushes retry at 100 ms, 200 ms, 400 ms, with at most three retry attempts | PRD 6.7; Architecture 7.4 | `api/internal/telemetry/retry.go`: `RetryBuffer`, `retryDelay` | Red: `TestRetryBufferRetriesFailedFlushWithExponentialBackoff` initially failed because `NewRetryBuffer` did not exist. Green: it records the exact retry schedule and a successful third retry; `TestRetryDelayIsExponentialAndCapped` proves the 5 s ceiling. |
| Retain no more than 10,000 retry-buffered events and expose occupancy | PRD 6.7; `docs/architecture.md` non-functional decisions | `RetryBuffer.Enqueue`, `RetryBuffer.Metrics` | `TestRetryBufferCapacityIsTenThousandEventsAndObservable` fills exactly 10,000 events, rejects event 10,001, and observes saturation. |
| Exhaustion or saturation activates admission backpressure and valid batches receive `503` before queueing | PRD 6.1/6.7; Architecture 5.2/7.4 | `RetryBuffer.AllowsAdmission`; `ingestion.AdmissionGuard`; `telemetryAdmissionEndpoint` | `TestRetryBufferExhaustionActivatesAdmissionBackpressure` asserts three failed attempts, structured loss logging, and blocked admission. `TestTelemetryAdmissionEndpointRejectsBackpressuredBatchBeforeQueue` asserts `503`, request ID, and zero queue writes. |
| PostgreSQL pressure is independent of Redis health; memory is bounded | PRD 6.5/6.7; Architecture 2 and 7.4 | Retry buffer has only a PostgreSQL repository dependency and a 10,000-event bound | `TestRetryBufferActivatesBackpressureDuringPostgreSQLOutage` stops a real PostgreSQL Testcontainer and verifies exhaustion/backpressure without Redis. |
| Failed batch ownership transfers from the batcher to the retry buffer | Architecture 3.3 and 7.4 | `Batcher` failure path with its optional `RetryBuffer` | `TestBatcherMovesFailedFlushIntoBoundedRetryBuffer` verifies the failed flush is retried and released after success. |
| A saturated persistence input never silently drops an admitted batch | PRD 6.4; Architecture 3.3 | `Batcher.Processor` transfers overflow to `RetryBuffer` | Red: `TestBatcherProcessorMovesSaturatedInputIntoRetryBuffer` timed out under the old dropped-batch behavior. Green: it proves overflow is retried without activating backpressure while capacity remains. |
| The production process shares the retry guard with both the batcher and admission route | Architecture 3.3/7.4; review finding | `api/cmd/api/application.go` and `cmd/api/main.go` | `TestApplicationRejectsValidBatchBeforeQueueWhenRetryGuardIsUnavailable` uses real PostgreSQL authentication and proves a valid request returns `503` with queue depth zero. |

Quality evidence (2026-09-11):

- `go test -count=1 ./internal/telemetry -run TestRetryBufferActivatesBackpressureDuringPostgreSQLOutage` — passed against a stopped real PostgreSQL container with Docker access.
- `go test -count=1 -race ./internal/telemetry ./internal/ingestion` — passed with Docker-backed integration tests.
- `go test -count=1 -race ./...`, `go vet ./...`, formatting, and `git diff --check` — passed.
- After review fixes (2026-09-12): `go test -count=1 -race ./...`, `go vet ./...`, formatting, and `git diff --check` — passed with Docker-backed runtime composition coverage.

## Review

Not approved.

1. **Blocking — the retry/backpressure path is not composed into the running API.** `api/cmd/api/main.go` creates only the generic health/readiness handler. It does not create a PostgreSQL pool, `RetryBuffer`, `Batcher`, `DeviceShardedQueue`, or `NewTelemetryAdmissionRoute`. The only `NewRetryBuffer`, guarded admission-route, and retry-enabled batcher call sites are tests. Consequently no production `POST /v1/telemetry/batches` request can observe a PostgreSQL retry failure or receive the required pre-enqueue `503`. Wire one shared `RetryBuffer` into the live batcher and the live admission route, and add a runtime-composition test proving a valid authenticated request is rejected before queue insertion once real persistence backpressure is active.

Independent checks otherwise pass: `go test -count=1 -race ./internal/telemetry ./internal/ingestion`, `go vet ./...`, formatting, and `git diff --check`. The Docker-backed PostgreSQL outage test also passes, but it exercises an isolated test composition rather than the application process.

### Resolution and approval (2026-09-12)

Approved. `api/cmd/api/application.go` now creates one PostgreSQL repository, retry buffer, batcher, device-sharded queue, and guarded ingestion route. `cmd/api/main.go` serves that handler and shuts those components down in drain order. The new Docker-backed application test authenticates a real device credential and proves retry-guard backpressure returns the documented JSON `503` before a queue write. A second regression test proves a saturated batcher input transfers the admitted batch into the bounded retry buffer rather than dropping it. Independent race tests, PostgreSQL outage/runtime tests, vet, formatting, and diff checks pass.
