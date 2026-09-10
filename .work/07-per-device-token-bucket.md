# 07 — Per-device event token bucket

**Type:** logic (test-first)
**Blocked by:** 05, 06 — the authenticated device and valid event count are required.
**Status:** planned

## What this delivers

Each device API key receives 10 event tokens per second with a 100-event burst, protecting the service without turning batches into request-count limits.

## Lifecycle

- **Create:** A bucket is lazily created for an authenticated device.
- **Read:** Admission checks its current available tokens.
- **Update:** Tokens refill by elapsed monotonic time only.
- **Delete:** Idle bucket cleanup is permitted as an internal memory-management action.
- **Undo:** Not applicable; token consumption is time-bound.

## Acceptance criteria

- [ ] Token cost equals validated event count; a batch is accepted or rejected atomically.
- [ ] Capacity is 100 and refill is 10 events/second under concurrent access.
- [ ] Insufficient tokens return `429`, do not enqueue, and include `Retry-After` when calculable.
- [ ] Metrics/documentation state that enforcement is in-memory and instance-local.

## Out of scope

- Redis/global rate limits, quotas, and client self-service limits.

## Verification

- Write deterministic time/concurrency tests before implementation and run race detection.

