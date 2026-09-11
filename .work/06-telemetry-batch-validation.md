# 06 — Whole-batch telemetry validation

**Type:** logic (test-first)
**Blocked by:** 02, 05 — validation follows the contract and authenticated device binding.
**Status:** done

## What this delivers

Malformed, oversized, cross-device, or out-of-range telemetry never reaches a worker; a valid request becomes an all-or-nothing validated batch.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Not applicable.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Non-JSON returns `415`; bodies over 1 MB return `413` before queue activity.
- [ ] Empty, over-500, malformed, non-finite, wrong-unit, invalid-timestamp, cross-device, or range-invalid batches return `400` as a whole.
- [ ] Exact temperature, voltage, battery, and pressure acceptance bounds match the PRD.
- [ ] Events over 60 seconds late are valid for persistence but explicitly marked to skip live aggregation.

## Out of scope

- Rate limiting, enqueueing, persistence, and partial successful batches.

## Verification

- Write table-driven failing tests for every boundary and verify no rejection invokes the queue seam.

## Build-gate trace

| Requirement / binding rule | Implementation | Repeatable proof |
| --- | --- | --- |
| Non-JSON requests receive `415`, bodies over 1 MiB receive `413`, and neither reaches admission; ticket acceptance criterion and `api/openapi.yaml` ingestion responses | `api/internal/ingestion/validation.go:BatchValidator.Validate` and `Middleware` | Red: `go test ./internal/ingestion` failed because the new validator boundary did not exist. Green: `TestBatchValidatorRejectsWholeInvalidBatchBeforeDownstreamAdmission` asserts statuses, stable JSON error codes, request IDs, and no downstream call; it also covers the over-limit body. `TestBatchValidatorAcceptsExactlyOneMiBBody` covers the inclusive boundary. |
| Whole-batch schema, cardinality, finite number, canonical-unit, UTC timestamp, device-binding, and metric-range validation; PRD §6.2 and OpenAPI telemetry schemas | `api/internal/ingestion/validation.go:validateEvents`, `validateEvent`, and `optionalUnit.UnmarshalJSON` | The table-driven rejection test covers empty/501-event, malformed, unknown-field, missing-value, non-finite, wrong-unit (including explicitly empty and `null` unit), invalid-timestamp, cross-device, mixed-validity, and all four out-of-range cases, asserting no downstream admission. `TestBatchValidatorRejectsMalformedUTF8BeforeAdmission` covers invalid JSON text encoding. `TestBatchValidatorAcceptsMaximumBatchCardinality` proves 500 is accepted. |
| Inclusive temperature [-40,125], voltage [0,48], battery [0,100], and pressure [0,1000] ranges, with canonical default units; PRD §6.2 and OpenAPI component schemas | `measurementRules` and `ValidatedEvent.Unit` in `api/internal/ingestion/validation.go` | `TestBatchValidatorAcceptsExactBoundsAndDefaultsCanonicalUnits` exercises every exact lower/upper bound and omitted-unit normalization. |
| Events more than 60 seconds late persist but skip live aggregation; PRD §6.2 and Architecture.md §7.1 | `ValidatedEvent.SkipLiveAggregation` in `api/internal/ingestion/validation.go` | `TestBatchValidatorMarksLateEventsWithoutRejectingThem` proves a strictly over-60-second event reaches the downstream seam with the skip marker. |
| Authentication precedes validation and rejection cannot perform dependency I/O; `docs/Architecture.md` §6.1 and `docs/conventions.md` | `BatchValidator.Middleware`, using the unexported authentication context identity and pure request validation | `TestBatchValidatorFailsClosedWithoutAuthentication` proves no unauthenticated request reaches downstream. Validation has no Redis/PostgreSQL dependency; `go test -race ./...` covers the composed package. |

Preflight inspected `api/openapi.yaml`, `deploy/Caddyfile`, `deploy/compose.yaml`, and `docs/database.md`. Quality gates passed: `test -z "$(gofmt -l .)"`, `go test -race ./...`, `go vet ./...`, and `git diff --check`. Docker is unavailable in this environment, so Compose runtime validation is deferred to a Docker-capable CI/host as documented in `docs/setup.md`.

## Review

**Resolved R1 (contract validation):** `optionalUnit.UnmarshalJSON` records field presence and rejects `null`; a supplied empty unit is now distinct from omission and fails the canonical-unit check. Regression cases in `TestBatchValidatorRejectsWholeInvalidBatchBeforeDownstreamAdmission` first failed against the prior code (`202`), then pass with `400` and no downstream admission.

Approved. Independent checks passed: `go test -count=1 -race ./internal/ingestion ./internal/platform/httpserver ./contract`, `go vet ./...`, format, and whitespace checks. No ticket-scope, validation, authorization, or contract findings remain. Docker/Compose runtime validation remains unavailable in this environment.
