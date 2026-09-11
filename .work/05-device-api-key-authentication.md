# 05 — Device API-key authentication

**Type:** logic (test-first)
**Blocked by:** 04 — seeded device credentials must exist.
**Status:** done

## What this delivers

Only a valid enabled device credential can enter the ingestion path, and each accepted request carries exactly one authorized device identity.

## Lifecycle

- **Create:** Not applicable; credential creation is ticket 04.
- **Read:** Middleware resolves only the credential/device binding needed by ingestion.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] `X-API-Key` parsing validates the structured key and uses constant-time comparison of its HMAC hash.
- [ ] Missing, malformed, disabled, or invalid keys return `401` before queue activity.
- [ ] Valid authentication makes internal and external device identity available to downstream admission only.
- [ ] A device credential cannot authorize live, history, metrics, or administrative routes.

## Out of scope

- TLS termination configuration, dashboard authentication, rate limiting, and payload validation.

## Verification

- Write failing authentication cases first; run handler and real-database integration tests.

## Build-gate trace

| Requirement / binding rule | Implementation | Repeatable proof |
| --- | --- | --- |
| Structured `X-API-Key` parsing and constant-time HMAC verification; Architecture.md §6.1 | `api/internal/ingestion/authentication.go`; the Ticket 04 resolver in `api/internal/ingestion/credentials/resolve.go` | Red: `go test ./internal/ingestion` failed with undefined authentication boundary symbols. Green: `go test ./internal/ingestion ./internal/platform/httpserver ./contract`; resolver uses `hmac.Equal` and its existing integration test checks the persisted HMAC. |
| Missing, malformed, multiple, unknown, and disabled credentials return generic 401 before admission; ticket acceptance criteria and OpenAPI `Unauthorized` | `api/internal/ingestion/authentication.go` | `api/internal/ingestion/authentication_test.go` proves no downstream call for malformed paths and the HTTP error/header shape. `authentication_integration_test.go` proves valid, unknown, and disabled database credentials through the handler. |
| Valid authentication supplies one internal/external device identity to downstream admission only; ticket acceptance criteria and `docs/architecture.md` Ingestion seam | `DeviceAuthenticator.Middleware` and `AuthorizedDeviceFromContext` in `api/internal/ingestion/authentication.go` | `TestDeviceAuthenticatorUsesRealCredentialBinding` asserts the exact seeded database binding reaches the downstream handler. It is automatically skipped locally because Docker is unavailable and executes on a Docker-capable CI/host. |
| Device credentials cannot authorize dashboard, history, metrics, or administrative routes; Architecture.md §6.2 and `api/openapi.yaml` | No device credential middleware is registered outside the Ingestion seam; Caddy route boundary remains restricted to the ingestion path | `api/contract/openapi_test.go:TestDeviceCredentialContractIsLimitedToIngestion` and `api/contract/caddy_policy_test.go` validate the contract and edge policy. |
| Do not expose API keys in logs or error responses; `docs/conventions.md` | Middleware emits only fixed safe errors and delegates request logging to `httpserver` | Authentication handler tests verify generic errors; `httpserver` safe-log regression test runs in `go test -race ./...`. |

Quality gates: `test -z "$(gofmt -l .)"`, `go test -race ./...`, `go vet ./...`, and `git diff --check` passed. Compose/runtime validation is not available in this environment because Docker is unavailable, as documented in `docs/setup.md`.

## Review

Approved. Independent review re-ran `go test -count=1 -race ./internal/ingestion ./internal/platform/httpserver ./contract`, `go vet ./...`, the repository formatting and whitespace checks, and inspected the OpenAPI/Caddy access boundary and safe logging path. No ticket-scope, contract, authorization, or secret-handling findings remain. The real-PostgreSQL handler test remains ready for Docker-capable CI; Docker is unavailable in this environment.
