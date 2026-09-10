# 25 — MVP acceptance demonstration

**Type:** plumbing (test-after)
**Blocked by:** 18, 19, 20, 21, 22, 23, 24 — all product, quality, and deployment work must be complete.
**Status:** planned

## What this delivers

A portfolio-quality, repeatable demonstration proves the documented ingestion reliability, Redis-backed UI, bulk persistence, and failure isolation acceptance criteria.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Demonstration reads live fleet/device/history state and recorded verification evidence.
- **Update:** Demo fixtures/configuration may be refreshed through controlled seed procedures.
- **Delete:** Demo data follows the documented short retention/recovery process.
- **Undo:** Reset the isolated demo environment from migrations and fixtures.

## Acceptance criteria

- [ ] Demo shows authenticated valid admission, malformed/oversized rejection, per-device rate limit, and bounded `503` backpressure without process crash.
- [ ] Fleet/device UI reads Redis-backed state, presents stale/degraded mode honestly, and makes no live PostgreSQL fallback.
- [ ] Accepted batches demonstrate bulk idempotent PostgreSQL persistence inside configured size/time behavior.
- [ ] Acceptance run includes current automated test, audit, Compose, and benchmark evidence with known limitations stated explicitly.

## Out of scope

- New feature work, claim of crash-safe acknowledgement, multi-tenancy, external brokers, or broader production certification.

## Verification

- Execute the scripted demo on the configured staging topology, capture results/screenshots/log references, and obtain independent review.
