# 22 — Backend integration and contract suite

**Type:** logic (test-first)
**Blocked by:** 09, 10, 11, 12, 13, 14, 16, 17 — all backend behaviors must be available.
**Status:** planned

## What this delivers

The MVP’s API, database, Redis, retry, and contract invariants are exercised against real dependencies, preventing regressions in the high-risk seams.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Test suite observes public HTTP/module seams and real storage behavior.
- **Update:** Fixtures evolve only with approved contract/schema changes.
- **Delete:** Test data is isolated and removed with containers.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Testcontainers suite covers idempotency, flush trigger, database outage/backpressure, Redis TTL/degradation, late events, exact thresholds, and rate limits.
- [ ] HTTP tests validate all documented success/error schemas against OpenAPI behavior.
- [ ] Tests never mock the project data layer or cross-module internals where a real dependency test is possible.
- [ ] CI runs the suite with bounded timeouts and produces actionable failures.

## Out of scope

- Browser E2E, k6 capacity evidence, production monitoring, or new product behavior.

## Verification

- Run the complete container-backed suite on a Docker-capable CI/host and attach results.

