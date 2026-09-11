# 07 — Per-device event token bucket

**Type:** logic (test-first)
**Blocked by:** 05, 06 — the authenticated device and valid event count are required.
**Status:** done

## What this delivers

Each device API key receives 10 event tokens per second with a 100-event burst, protecting the service without turning batches into request-count limits.

## Lifecycle

- **Create:** A bucket is lazily created for an authenticated device.
- **Read:** Admission checks its current available tokens.
- **Update:** Tokens refill by elapsed monotonic time only.
- **Delete:** Idle bucket cleanup is permitted as an internal memory-management action.
- **Undo:** Not applicable; token consumption is time-bound.

## Acceptance criteria

- [ ] Token cost equals validated event count; a batch is accepted or rejected atomically.
- [ ] Capacity is 100 and refill is 10 events/second under concurrent access.
- [ ] Insufficient tokens return `429`, do not enqueue, and include `Retry-After` when calculable.
- [ ] Metrics/documentation state that enforcement is in-memory and instance-local.

## Out of scope

- Redis/global rate limits, quotas, and client self-service limits.

## Verification

- Write deterministic time/concurrency tests before implementation and run race detection.

## Build-gate trace

| Requirement / binding rule | Implementation | Repeatable proof |
| --- | --- | --- |
| Charge one token per validated event and reject a whole batch without consuming a partial cost; PRD §6.3 and Architecture.md §5.2 | `api/internal/ingestion/rate_limit.go:DeviceRateLimiter.Allow` | Red: `go test ./internal/ingestion -run TestDeviceRateLimiter` failed before the limiter existed. Green: `TestDeviceRateLimiterChargesWholeValidatedBatchAtomically` consumes 98, rejects 3 with two remaining, then admits 2. |
| Capacity 100, refill 10 events/sec, monotonic elapsed time, and concurrent safety; ticket acceptance criteria and `docs/Techstack.md` backend choice | `DeviceRateLimiter`, per-key `deviceRateBucket`, and `golang.org/x/time/rate` in `api/internal/ingestion/rate_limit.go` | `TestDeviceRateLimiterStartsFullForAnyInjectedClock`, `TestDeviceRateLimiterRefillsAtTenPerSecondAndNeverExceedsBurst`, `TestDeviceRateLimiterDoesNotRefillWhenClockMovesBackward`, and `TestDeviceRateLimiterIsConcurrentAndPerKey` run deterministically; full race suite passes. |
| Insufficient tokens return typed `429`, preserve queue safety, and emit `Retry-After` when the cost can become available; PRD §6.3 and OpenAPI `TooManyRequests` | `DeviceRateLimiter.Middleware` | `TestDeviceRateLimiterReturnsCalculableRetryAndMiddlewareRejectsBeforeAdmission` proves a 429 response, request ID, rounded Retry-After header, and no downstream admission. An event cost above the 100-token burst has no calculable retry. |
| Rate limits are process-local rather than a global quota; ticket acceptance criterion and `docs/prd.md` non-goals | `DeviceRateLimiter` owns a Go `sync.Map` only; no Redis/database client is introduced | `docs/Techstack.md` states the MVP limiter is instance-local; the implementation comment makes the same operational boundary explicit. Ticket 15 owns operational metrics, so this ticket adds no premature metrics endpoint. |
| Rate limiting follows authentication and validation; `docs/Architecture.md` §7.1 | `ValidatedBatchFromContext` and `withValidatedBatch` in the Ingestion module | `TestDeviceRateLimiterMiddlewareFailsClosedWithoutValidatedBatch` proves misordered/unvalidated input returns 400 before downstream admission. |

Preflight inspected `api/openapi.yaml`, `deploy/Caddyfile`, `deploy/compose.yaml`, and `docs/database.md`. The documented `golang.org/x/time/rate` dependency is now explicit in `api/go.mod`; no unrelated dependency updates remain. Quality gates passed: `test -z "$(gofmt -l .)"`, `go test -race ./...`, `go vet ./...`, and `git diff --check`. Docker/Compose runtime validation is unavailable locally, as documented in `docs/setup.md`.

## Review

Approved. Independent review confirmed that the limiter is keyed by the authenticated, non-secret API-key ID; charges the complete validated batch atomically; starts at the 100-event burst; refills at 10 events per second; and serializes each bucket safely under concurrent access. Rejected batches are not admitted downstream, return the published `429`/`rate_limited` contract, and include a rounded `Retry-After` only when a retry can be calculated. The implementation is explicitly process-local and introduces no Redis or PostgreSQL dependency on the admission path.

Independent verification passed: `go test -count=1 -race ./...`, targeted ingestion/HTTP/contract tests, `go vet ./...`, `gofmt` cleanliness, and `git diff --check`. Docker/Compose runtime verification remains unavailable in this environment.
