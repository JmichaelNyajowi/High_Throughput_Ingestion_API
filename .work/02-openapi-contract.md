# 02 — Versioned telemetry API contract

**Type:** logic (test-first)
**Blocked by:** None — the contract can be authored alongside the application shell.
**Status:** planned

## What this delivers

The device and operator HTTP contract is explicit in OpenAPI 3.1 before handlers exist, including truthful `202` admission semantics and bounded live/history responses.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Consumers read the versioned API contract.
- **Update:** Contract changes require a reviewed additive or versioned change.
- **Delete:** Not applicable.
- **Undo:** Revert an unpublished contract change through source control.

## Acceptance criteria

- [ ] `api/openapi.yaml` defines every approved `/v1`, health, readiness, and metrics endpoint plus JSON error envelopes and `X-Request-ID`.
- [ ] Ingestion schema enforces 1–500 events, required fields, canonical units, API-key security, and `202`, `400`, `401`, `413`, `415`, `429`, `503` meanings.
- [ ] Live endpoints declare Redis-derived `live`/`degraded` state; history declares required bounded query parameters and no automatic refresh use.
- [ ] Contract lint/validation runs in CI without a live service.

## Out of scope

- Handler implementation, client generation, API-key provisioning, or undocumented endpoints.

## Verification

- Add contract validation tests and run the documented API lint/check command.

