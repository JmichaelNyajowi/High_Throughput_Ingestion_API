# 02 — Versioned telemetry API contract

**Type:** logic (test-first)
**Blocked by:** None — the contract can be authored alongside the application shell.
**Status:** done

## What this delivers

The device and operator HTTP contract is explicit in OpenAPI 3.1 before handlers exist, including truthful `202` admission semantics and bounded live/history responses.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Consumers read the versioned API contract.
- **Update:** Contract changes require a reviewed additive or versioned change.
- **Delete:** Not applicable.
- **Undo:** Revert an unpublished contract change through source control.

## Acceptance criteria

- [x] `api/openapi.yaml` defines every approved `/v1`, health, readiness, and metrics endpoint plus JSON error envelopes and `X-Request-ID`.
- [x] Ingestion schema enforces 1–500 events, required fields, canonical units, API-key security, and `202`, `400`, `401`, `413`, `415`, `429`, `503` meanings.
- [x] Live endpoints declare Redis-derived `live`/`degraded` state; history declares required bounded query parameters and no automatic refresh use.
- [x] Contract lint/validation runs in CI without a live service.

## Out of scope

- Handler implementation, except the narrow `/healthz` and `/readyz` contract-parity responses required to remove a published-interface conflict; client generation, API-key provisioning, or undocumented endpoints.

## Verification

- [x] Test-first evidence: contract tests failed before `api/openapi.yaml` existed, then passed after the document was authored.
- [x] `GOCACHE=/tmp/telemetry-go-cache go test -race ./...`, `go vet ./...`, and `test -z "$(gofmt -l .)"` pass in `api/`.
- [x] `api/contract/openapi_test.go` loads and validates the OpenAPI 3.1 document without a live service. The existing API CI job runs `go test -race ./...`, so this validation is enforced in CI.
- [x] Regression tests prove units are optional with canonical defaults, event/history timestamps require a UTC `Z` suffix, and the 1,000-event history bound is explicit.
- [x] Repository checks pass: frontend lint, typecheck, tests, production build, and `npm audit` (0 vulnerabilities); `git diff --check` passes.
- [x] Contract-parity handler tests prove `/healthz` and `/readyz` return the documented `200` JSON payloads with UUID `X-Request-ID` headers.
- [x] Test-first Caddy access-policy evidence: the edge-policy test failed while observability routes were reverse-proxied through Caddy, while operator handling could be reordered, and while Caddy-generated denial responses lacked a request ID. It passed after Caddy denied public observability access, nested the private-operator branch before its denial fallback, and supplied a default-only `X-Request-ID`.
- [x] Docker and the Caddy binary are unavailable in this environment, so Compose/Caddy runtime validation cannot be claimed. `api/contract/caddy_policy_test.go` provides CI-run static coverage of the approved route policy; a Docker-capable release environment must still run Compose configuration validation.

### Build-gate trace

| Material requirement | Approved source | Implementation location | Repeatable proof |
| --- | --- | --- | --- |
| Approved endpoint set, security boundary, JSON errors, request ID, and `202` admission meaning | `docs/Architecture.md` sections 5.1–5.2; `docs/prd.md` section 6.1 | `api/openapi.yaml`, `api/contract/openapi_test.go` | OpenAPI structural validation and endpoint/status/security assertions. |
| 1–500 atomic batches; ranges; optional canonical units; RFC 3339 UTC timestamps | `docs/prd.md` section 6.2; `docs/Architecture.md` section 5.2 | Telemetry schemas and contract assertions in `api/openapi.yaml` and `api/contract/openapi_test.go` | Contract regression tests for batch bounds, measurement schemas, optional units, and `Z`-suffixed timestamps. |
| Redis-only `live`/`degraded` reads and explicit manual bounded history | `docs/Architecture.md` section 5.3; `docs/prd.md` sections 6.5 and 6.8 | Live/history operations and schemas in `api/openapi.yaml` | Contract endpoint, query-parameter, mode, and manual-use assertions. |
| 1–1,000 event history page bound | `docs/prd.md` section 6.8; `docs/Architecture.md` section 5.3 | `HistoryLimit` in `api/openapi.yaml`; `api/contract/openapi_test.go` | Contract limit schema and regression assertion. |
| Health/readiness public contract parity | `api/openapi.yaml`; `docs/Architecture.md` section 5.1 | `api/internal/platform/httpserver/handler.go` and its contract-parity test | Handler tests prove documented JSON status payloads and request IDs. |
| Prometheus wire-format exception | `docs/Architecture.md` section 5 | `/metrics` response in `api/openapi.yaml` | Architecture explicitly permits only `/metrics` to expose `text/plain; version=0.0.4`; OpenAPI validates that media type. |
| Caddy applies the declared internal and operator access boundaries | `docs/Architecture.md` section 3.1 and section 8; `api/openapi.yaml` descriptions/security | `deploy/Caddyfile`, `deploy/compose.yaml`, `api/contract/caddy_policy_test.go` | The test-first static edge-policy test verifies public observability denial, internal-only API connectivity, ordered private-network plus Basic Auth operator routes, and Caddy-generated request IDs. |

## Review findings

1. **Resolved — optional unit compatibility.** `unit` is optional in each telemetry-event schema. When absent, the API assigns its documented canonical wire unit; supplied units remain restricted to that type's canonical unit.
2. **Resolved — UTC timestamps.** Event and manual-history timestamps retain `date-time` validation and now require a UTC `Z` suffix through the contract pattern and regression tests.
3. **Resolved — bounded history.** The PRD and architecture now explicitly define the required `1–1,000` event result limit; the API contract and regression test enforce it.
4. **Resolved — `/healthz` and `/readyz` now conform to the published contract.** The bootstrap handler returns `200` JSON (`{"status":"ok"}` and `{"status":"ready","admission":"accepting"}` respectively) and a UUID `X-Request-ID`; contract-parity tests cover both endpoints.
5. **Resolved — architecture explicitly recognizes `/metrics` as the only text response.** `/metrics` returns `text/plain; version=0.0.4` for Prometheus compatibility while retaining `X-Request-ID`; every other success and error response remains JSON.
6. **Resolved — `deploy/Caddyfile` now enforces the access boundary asserted by the contract.** `/healthz`, `/readyz`, and `/metrics` are denied at Caddy and therefore remain reachable only through the API’s un-published internal Docker network. Operator live/history routes deterministically require both `remote_ip private_ranges` and Caddy `basic_auth`; device ingestion remains outside the operator gate. Caddy supplies a default-only `X-Request-ID` for its own 401/403 responses without replacing the API’s header. `api/contract/caddy_policy_test.go` was written red first and prevents the public proxy, unordered operator guard, missing operator protection, or missing Caddy request ID from returning. Docker/Caddy runtime validation remains a release-environment check because neither executable is available here.
7. **Approved — independent review found no remaining ticket-level discrepancy.** OpenAPI validation, contract regressions, health/readiness handler parity, Caddy/Compose policy checks, Go race tests, vet, formatting, and whitespace checks pass. No repository merge or release action was performed in this workspace.
