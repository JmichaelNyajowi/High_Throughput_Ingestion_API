# 17 — Bounded manual telemetry history API

**Type:** logic (test-first)
**Blocked by:** 02, 10, 15 — contract, raw events, and protected runtime routes must exist.
**Status:** planned

## What this delivers

An operator can deliberately request a limited, indexed PostgreSQL history result for one device without creating a live-query path.

## Lifecycle

- **Create:** Not applicable; history exposes immutable raw events.
- **Read:** Explicit operator request supplies device, `from`, `to`, and bounded page/limit.
- **Update:** Raw historical events are never edited.
- **Delete:** Raw historical events are never deleted in MVP.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Missing/invalid range or page bounds return typed validation errors before database work.
- [ ] Query uses parameterized indexed SQL, predictable ordering, a maximum range/page size, and a statement timeout.
- [ ] Route uses operator access policy and never accepts a device key as dashboard authorization.
- [ ] Response contract marks this endpoint manual and unsuitable for automatic polling.

## Out of scope

- Multi-month analytics, exports, retention, aggregation fallback, or background browser refresh.

## Verification

- Write validation and real-PostgreSQL query-plan/integration tests first.
