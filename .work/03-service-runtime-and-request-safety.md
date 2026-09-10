# 03 — Service runtime and request safety

**Type:** plumbing (test-after)
**Blocked by:** 02 — route behavior must align with the API contract.
**Status:** planned

## What this delivers

The Go service has safe request IDs, recovery, server/dependency timeouts, structured redacted logging, graceful shutdown, and route ownership ready for product handlers.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Health and readiness are readable by their approved internal consumers.
- **Update:** Runtime configuration changes only through validated environment configuration.
- **Delete:** Not applicable.
- **Undo:** Revert a configuration/runtime change through source control.

## Acceptance criteria

- [ ] Every response includes a propagated or generated request ID and structured logs include it without logging credentials or raw payloads.
- [ ] Panic recovery returns a safe error without killing the process; server read/write/idle limits and dependency timeouts are configured.
- [ ] Shutdown stops new admission before bounded worker/dependency teardown and exposes no false readiness.
- [ ] Route/middleware tests cover request IDs, recovery, and health/readiness access policy.

## Out of scope

- Device authentication, telemetry admission, worker processing, metrics content, or dashboard routes.

## Verification

- Run Go tests with race detection, static checks, and a shutdown test.

