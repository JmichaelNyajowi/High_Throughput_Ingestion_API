# 09 — Telemetry batch admission endpoint

**Type:** logic (test-first)
**Blocked by:** 02, 05, 06, 07, 08 — contract, authentication, validation, limit, and queue must be available.
**Status:** done

## What this delivers

Devices receive a precise `202 Accepted` only after a complete batch is authenticated, validated, rate-limited, and admitted to memory.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Not applicable.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] `POST /v1/telemetry/batches` composes the approved admission checks in the documented order.
- [ ] A `202` body includes request ID, accepted event count, and `accepted`; it explicitly means in-memory admission, not durable persistence.
- [ ] Every documented failure has a stable typed response, request ID, and no partial admission.
- [ ] The request path makes no synchronous PostgreSQL or Redis call before responding.

## Out of scope

- Flush confirmation, cache publication, load testing, or dashboard access.

## Verification

- Write HTTP contract tests first and assert dependency/queue behavior through confirmed seams.

## Build-gate trace

| Requirement / binding rule | Implementation | Repeatable proof |
| --- | --- | --- |
| Compose authentication, validation, rate limiting, and queue admission in the documented order; PRD §6.1–§6.4 and Architecture.md §§5.2, 6.1, 7.1 | `api/internal/ingestion/admission.go:NewTelemetryAdmissionRoute` | Red: `go test ./internal/ingestion -run TestTelemetryAdmissionEndpointAcknowledgesOnlyEnqueuedValidatedBatch` failed before the endpoint existed. The route is explicitly nested `auth -> validation -> rate limit -> endpoint`; `TestTelemetryAdmissionRouteRejectsUnauthenticatedRequestBeforeQueue` proves the first rejection reaches neither later stages nor the queue. |
| Return OpenAPI `AdmissionAccepted` only after the whole validated batch enters the queue, with no durable-write claim; OpenAPI `AdmissionAccepted` and PRD §6.1 | `telemetryAdmissionEndpoint.ServeHTTP` | `TestTelemetryAdmissionEndpointAcknowledgesOnlyEnqueuedValidatedBatch` asserts HTTP 202, JSON, matching request ID, `accepted_events: 2`, `status: accepted`, and one exactly-sized queue admission. |
| Queue overload is a stable, typed `503` error with request ID and no partial admission; OpenAPI `ServiceUnavailable` and PRD §6.1 | `admissionFailure` seam and `httpserver.WriteError` use in `ServeHTTP` | `TestTelemetryAdmissionEndpointReturnsStable503WithoutPartialAdmission` asserts `503`, `unavailable`, the request ID in header/body, and generic safe text even when the queue error carries internal detail. |
| Request admission never synchronously accesses PostgreSQL or Redis; docs/architecture.md system shape and conventions | `admission.go` depends solely on `BatchAdmissionQueue`; it imports no database/Redis client | Static inspection plus endpoint tests: the handler takes an already-validated batch and invokes exactly one queue seam before writing 202. Credential lookup remains in the mandated authentication boundary; no new persistence/cache operation is added. |

Preflight inspected `api/openapi.yaml`, `deploy/Caddyfile`, `deploy/compose.yaml`, `docs/database.md`, `docs/architecture.md`, and `docs/Architecture.md`. Caddy continues to proxy only `/v1/telemetry/batches` without dashboard Basic Auth; the route constructor is ready for the later application composition that supplies the PostgreSQL-backed authenticator and processing workers. The endpoint deliberately excludes persistence/cache work and does not alter the public contract. Quality gates passed: `go test -count=1 -race ./...`, `go vet ./...`, `test -z "$(gofmt -l .)"`, and `git diff --check`. Docker/Compose runtime validation cannot run because Docker is unavailable in this environment.

## Review

Approved. The Ingestion-owned route is composed in the required authentication → validation → rate-limit → bounded-queue order. The terminal endpoint acknowledges only the successful `BatchAdmissionQueue.Admit` call and emits the contract's complete JSON `202` envelope. Queue failures retain their typed `503`/`unavailable` status while suppressing internal queue text; all response paths use the shared request-ID and JSON helpers.

The endpoint depends only on the public Telemetry admission seam and introduces no Redis, PostgreSQL persistence, or broker operation before acknowledgement. Contract, edge-policy, and error-envelope checks remain compatible with the existing OpenAPI/Caddy boundary. Independent quality verification passed: `go test -count=1 -race ./...`, `go vet ./...`, `gofmt` cleanliness, and `git diff --check`. Docker/Compose runtime verification remains unavailable in this environment.
