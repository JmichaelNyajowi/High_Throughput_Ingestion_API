# 11 — Persistence retry and admission backpressure

**Type:** logic (test-first)
**Blocked by:** 09, 10 — admission and durable-write seam must exist.
**Status:** planned

## What this delivers

PostgreSQL failure is isolated behind a capped retry buffer; once safety limits are reached, new admission returns `503` instead of consuming unbounded memory.

## Lifecycle

- **Create:** Failed flushes enter the bounded retry buffer.
- **Read:** Admission reads the explicit backpressure state.
- **Update:** Retry timing advances from 100 ms exponentially, capped at 5 s and three attempts.
- **Delete:** Exhausted retry entries are dropped only with explicit structured loss logging.
- **Undo:** Not available; clients may retry using stable event IDs.

## Acceptance criteria

- [ ] Failed persistence retries at the required schedule, no more than three attempts.
- [ ] Retry capacity is 10,000 events and its occupancy is observable.
- [ ] Retry exhaustion or buffer saturation activates admission backpressure; subsequent valid batches receive `503` before enqueue.
- [ ] Redis availability never masks a PostgreSQL backpressure state; memory remains bounded.

## Out of scope

- External brokers, disk spool, exactly-once delivery, or automatic data recovery beyond client retry.

## Verification

- Write failure/backoff/buffer tests first; prove bounded behavior with PostgreSQL outage integration tests.

