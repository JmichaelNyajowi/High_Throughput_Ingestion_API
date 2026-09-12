# 22 — Backend integration and contract suite

**Type:** logic (test-first)
**Blocked by:** 09, 10, 11, 12, 13, 14, 16, 17 — all backend behaviors must be available.
**Status:** done

## What this delivers

The MVP’s API, database, Redis, retry, and contract invariants are exercised against real dependencies, preventing regressions in the high-risk seams.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Test suite observes public HTTP/module seams and real storage behavior.
- **Update:** Fixtures evolve only with approved contract/schema changes.
- **Delete:** Test data is isolated and removed with containers.
- **Undo:** Not applicable.

## Acceptance criteria

- [x] Testcontainers suite covers idempotency, flush trigger, database outage/backpressure, Redis TTL/degradation, late events, exact thresholds, and rate limits.
- [x] HTTP tests validate all documented success/error schemas against OpenAPI behavior.
- [x] Tests never mock the project data layer or cross-module internals where a real dependency test is possible.
- [x] CI runs the suite with bounded timeouts and produces actionable failures.

## Out of scope

- Browser E2E, k6 capacity evidence, production monitoring, or new product behavior.

## Verification

- Run the complete container-backed suite on a Docker-capable CI/host and attach results.

## Build and review trace

- Real PostgreSQL/Testcontainers coverage: credential bootstrap, idempotent bulk persistence, timer/size flushes, retry exhaustion, outage backpressure, and end-to-end admission/late-event persistence in `api/**/*integration_test.go` and `api/cmd/api/application_test.go`.
- Redis and processing seams: expiring derived-state publication/degradation, late-event boundaries, exact threshold boundaries, and per-device token buckets in `api/internal/telemetry/*_test.go` and `api/internal/ingestion/*_test.go`.
- Public contract/access seam: OpenAPI validation, error/success response declarations, and Caddy/Compose access policy in `api/contract/*_test.go`.
- CI runs `go vet ./...` and `go test -race ./...` in `.github/workflows/quality.yml`. The complete local race suite, vet, and format check passed on this Docker-capable host.
- Independent review found no Ticket 22 scope expansion or unresolved backend-suite blocker.
