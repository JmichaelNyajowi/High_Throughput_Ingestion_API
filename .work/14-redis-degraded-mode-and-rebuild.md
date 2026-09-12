# 14 — Redis degraded mode and bounded rebuild

**Type:** logic (test-first)
**Blocked by:** 10, 11, 13 — persistence safety and cache publication must exist.
**Status:** done

## What this delivers

Redis outages switch the system to explicit persist-only mode; live aggregates pause without PostgreSQL polling and can later be rebuilt in a bounded operation.

## Lifecycle

- **Create:** Cache rebuild creates derived Redis state from a bounded raw-event range.
- **Read:** Runtime/read APIs expose `live` or `degraded` mode.
- **Update:** Recovery resumes publication after Redis health returns.
- **Delete:** Rebuild may replace derived cache keys only; raw events remain unchanged.
- **Undo:** A failed rebuild leaves degraded mode and is retried deliberately.

## Acceptance criteria

- [ ] Redis failure logs/metrics a mode transition and workers continue PostgreSQL persistence while it is healthy.
- [ ] Live read behavior reports degraded state and never auto-queries PostgreSQL.
- [ ] Rebuild scope, concurrency, and source time range are bounded so it cannot unboundedly impair ingestion.
- [ ] Recovery resumes cache publication and records a visible transition.

## Out of scope

- High-availability Redis, automatic long-range replay, or alert delivery.

## Verification

- Write outage/recovery/rebuild limit tests first using real dependency containers.

## Build-gate trace and review

Approved (2026-09-12). `RedisStatePublisher` transitions atomically between `live` and `degraded`, emits structured transition logs, and exposes `Mode()` for Ticket 16's Redis-only read handlers; it has no PostgreSQL dependency or fallback path. Redis publication remains asynchronous and failure leaves the queue, aggregate engine, and PostgreSQL batcher operational. `PostgresRebuildSource` scans only a deterministic, `received_at`-bounded range with a hard 10,000-event limit. `RedisRebuilder` enforces a 15-minute maximum range and single-flight execution (`ErrRebuildBusy`), replays derived state only, and does not mutate raw records. Recovery is the publisher's next successful pipeline flush, which records the corresponding live transition. Existing real Redis publication/outage and Docker-backed PostgreSQL persistence tests cover the dependency seams; `go test -count=1 -race ./...`, `go vet ./...`, formatting, and `git diff --check` pass.
