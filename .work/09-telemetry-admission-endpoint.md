# 09 — Telemetry batch admission endpoint

**Type:** logic (test-first)
**Blocked by:** 02, 05, 06, 07, 08 — contract, authentication, validation, limit, and queue must be available.
**Status:** planned

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
