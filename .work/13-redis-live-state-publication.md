# 13 — Redis live state and TTL publication

**Type:** logic (test-first)
**Blocked by:** 12 — aggregate snapshots must exist.
**Status:** done

## What this delivers

Redis contains compact, versioned, expiring device aggregate and freshness state for fast live dashboard reads.

## Lifecycle

- **Create:** A processed aggregate creates/refreshes its Redis snapshot.
- **Read:** Fleet and device read APIs consume only this derived state.
- **Update:** Each valid processed event refreshes aggregate/freshness state and 15-minute TTLs.
- **Delete:** Redis expiry removes offline state after 15 minutes.
- **Undo:** Cache state may be rebuilt; it is never a durable record.

## Acceptance criteria

- [ ] Aggregate, last-seen, and measurement-set keys follow the documented versioned key model.
- [ ] Each aggregate snapshot receives a refreshed 15-minute TTL; 60-second freshness derives from processed time.
- [ ] Writes are bounded/coalesced and use short dependency timeouts.
- [ ] Redis errors are observable and do not block the worker indefinitely.

## Out of scope

- Redis as a queue, global rate limiter, primary database, or dashboard PostgreSQL fallback.

## Verification

- Write Redis TTL/freshness tests first against a real Redis container.

## Build-gate trace

| Contract | Source | Implementation | Repeatable proof |
| --- | --- | --- | --- |
| Redis uses the versioned aggregate hash, device measurement-set, and device last-seen key model | `docs/Architecture.md` §4.2; Ticket 13 lifecycle | `api/internal/telemetry/redis_state.go`: key helpers and pipeline | Red: `TestRedisStatePublisherPublishesVersionedExpiringSnapshotAndFreshness` initially failed because no publisher existed. Green: it uses a real Redis 7.4 container and asserts the exact versioned keys, aggregate hash fields, set membership, and processing-time sorted-set score. |
| Aggregate and measurement-set keys refresh their 15-minute TTL; freshness is derived from worker processing time | PRD §6.5; `CONTEXT.md` freshness definition | `RedisStatePublisher.flush` applies `EXPIRE` after aggregate/set writes and records the passed processing timestamp in `device-last-seen` | The real Redis test asserts both positive, bounded 15-minute TTLs and verifies last-seen equals processing time rather than the event timestamp. `ZREMRANGEBYSCORE` prunes entries older than the same cache-retention window. |
| Redis writes are bounded, coalesced at most every 250 ms, and cannot block a shard worker | `docs/Architecture.md` §§7.1, 7.2, 10; Ticket 13 acceptance criteria | `RedisStatePublisher` has a 1,024-item nonblocking input, coalesces by aggregate key, and sends a pipeline from one background goroutine | `TestRedisStatePublisherCoalescesAndDoesNotBlockOnRedisFailure` submits 10,000 updates against an unavailable dependency and proves the caller returns within 250 ms while bounded drops are recorded. Constructor validation caps both the flush interval and dependency timeout at 250 ms. |
| Redis failures are observable and do not stop aggregate or PostgreSQL processing | PRD §§6.4–6.5; `docs/architecture.md` degraded-state rule | Publisher records `Errors`/`Dropped` metrics and emits dependency-error logs; `api/cmd/api/application.go` queues publication before independent persistence | Docker-backed `TestApplicationRejectsValidBatchBeforeQueueWhenRetryGuardIsUnavailable` runs with an intentionally unreachable Redis URL: it observes publisher errors while the aggregate is built and the late event is persisted to PostgreSQL. |
| No Redis I/O enters the synchronous admission route, and Redis is not used as a queue/database/fallback | PRD non-goals; `docs/conventions.md` | Redis client/publisher is created only in `cmd/api` processing composition; `ingestion` remains Redis-free | Static import/call-path review plus existing admission tests: admission returns from the bounded queue before worker publication. No live-read API or PostgreSQL fallback was introduced. |

Quality evidence (2026-09-12):

- `go test -count=1 ./internal/telemetry -run TestRedisStatePublisher` — passed with a real Redis 7.4 container.
- `go test -count=1 ./cmd/api -run TestApplicationRejectsValidBatchBeforeQueueWhenRetryGuardIsUnavailable` — passed with real PostgreSQL and unreachable Redis failure isolation.
- `go test -count=1 -race ./...`, `go vet ./...`, formatting, and `git diff --check` — passed.

## Review

Approved (2026-09-12). The implementation publishes only derived aggregate state through the documented `telemetry:v1` aggregate hash, measurement-set, and last-seen sorted-set model. Aggregate/set TTLs refresh at 15 minutes; freshness is based on server processing time and stale last-seen members are pruned. Publication is intentionally asynchronous: a bounded nonblocking queue, per-key coalescing, a maximum 250 ms flush cadence/timeout, and pipelined writes prevent Redis from stalling the shard worker or PostgreSQL persistence. Failure counters and structured dependency-error logs make Redis failure visible without leaking credentials or raw payloads. The review also verified that ingestion remains Redis-free until after queue admission, that no Redis fallback/database role was introduced, and that real Redis/PostgreSQL integration, full race tests, vet, formatting, and diff checks pass.
