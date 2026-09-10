# 03 — Service runtime and request safety

**Type:** plumbing (test-after)
**Blocked by:** 02 — route behavior must align with the API contract.
**Status:** done

## What this delivers

The Go service has safe request IDs, recovery, server/dependency timeouts, structured redacted logging, graceful shutdown, and route ownership ready for product handlers.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Health and readiness are readable by their approved internal consumers.
- **Update:** Runtime configuration changes only through validated environment configuration.
- **Delete:** Not applicable.
- **Undo:** Revert a configuration/runtime change through source control.

## Acceptance criteria

- [x] Every response includes a propagated or generated request ID and structured logs include it without logging credentials or raw payloads.
- [x] Panic recovery returns a safe error without killing the process; server read/write/idle limits and dependency timeouts are configured.
- [x] Shutdown stops new admission before bounded worker/dependency teardown and exposes no false readiness.
- [x] Route/middleware tests cover request IDs, recovery, and health/readiness access policy.

## Out of scope

- Device authentication, telemetry admission, worker processing, metrics content, or dashboard routes.

## Verification

- [x] Test-first evidence: runtime tests initially failed because request middleware, readiness control, lifecycle shutdown, and validated timeout configuration did not exist. They pass after the minimum runtime implementation.
- [x] `GOCACHE=/tmp/telemetry-go-cache go test -race ./...`, `GOCACHE=/tmp/telemetry-go-cache go vet ./...`, `test -z "$(gofmt -l .)"`, and `git diff --check` pass.
- [x] Middleware tests prove a valid upstream request ID is propagated into the response and structured log while API-key, raw-body, and query-string values are absent from logs.
- [x] Recovery tests prove a panicking registered route returns a generic JSON `500` envelope with a matching request ID and no panic detail.
- [x] Lifecycle tests prove readiness is paused before HTTP shutdown and before future dependency drainers; paused readiness returns its documented `503` state while liveness remains `200`.
- [x] Contract-parity regression test proves a paused `/readyz` response documents the same `status` and `admission` fields returned by the handler.
- [x] Ticket 02’s CI-run Caddy/Compose policy tests remain the edge-policy proof: health/readiness are not publicly proxied and the API remains un-published on the internal network. Docker/Caddy runtime validation is unavailable in this environment and remains a release-host check.

### Build-gate trace

| Requirement or binding rule | Governing source | Implementation location | Repeatable proof |
| --- | --- | --- | --- |
| Propagated/generated request IDs and redacted structured logs | This ticket; `docs/prd.md` section 6.1; `docs/Architecture.md` sections 5 and 6 | `api/internal/platform/httpserver/handler.go` | Request propagation/redaction test and router-error request-ID test. |
| Safe panic containment and explicit runtime/dependency timeouts | This ticket; `docs/Techstack.md` backend design rules; `docs/conventions.md` backend rules | HTTP recovery middleware; `api/internal/platform/config/config.go`; `.env.example` | Recovery and timeout configuration tests; full Go quality gate. |
| Shutdown pauses admission/readiness before HTTP and dependency teardown | This ticket; `docs/prd.md` section 6.4; `docs/Architecture.md` sections 3.3 and 7.4 | `api/internal/platform/runtime/runtime.go`, `api/cmd/api/main.go`, readiness handler | `TestShutdownPausesReadinessBeforeServerAndDependencyTeardown` and paused-readiness handler test. |
| Health/readiness route and internal access policy | This ticket; `api/openapi.yaml`; `docs/Architecture.md` section 3.1 | `httpserver` health/readiness routes; `deploy/Caddyfile` and `deploy/compose.yaml` from Ticket 02 | Health/readiness handler tests plus Ticket 02 Caddy/Compose contract tests. |

## Review findings

1. **Resolved during independent re-review — paused readiness response contract mismatch.** The handler returned the documented readiness state (`status` and `admission`) with `503`, while `api/openapi.yaml` incorrectly declared the generic `ServiceUnavailable` error envelope. `TestPausedReadinessContractReturnsReadinessState` was added red, then the contract gained the explicit `ReadinessUnavailable` response pointing to `ReadinessStatus`. The handler and public contract now agree.
2. **Approved — no remaining ticket-level discrepancy.** Request-ID propagation/generation, redacted JSON logging, panic containment, validated timeout configuration, readiness pausing, lifecycle order, and Caddy/Compose internal-access policy are covered by tests. Go race tests, vet, formatting, whitespace checks, frontend lint/typecheck/tests/production build, and production dependency audit all pass. Docker/Caddy runtime validation remains a release-host check because those executables are unavailable in this environment. No repository merge or release action was performed in this workspace.
