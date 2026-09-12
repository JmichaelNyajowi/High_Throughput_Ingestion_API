# 12 — Five-minute aggregate and threshold engine

**Type:** logic (test-first)
**Blocked by:** 08, 09 — ordered shard processing and accepted events are required.
**Status:** done

## What this delivers

Workers maintain bounded processing-time five-minute aggregates and exact live threshold state per device and measurement.

## Lifecycle

- **Create:** Processing creates/updates derived in-memory aggregate state.
- **Read:** Aggregate snapshots are exposed only to cache publication.
- **Update:** New in-window processed events update the relevant bounded bucket.
- **Delete:** Expired buckets leave the sliding window automatically.
- **Undo:** Not applicable; derived state is recomputed from subsequent processing/rebuild.

## Acceptance criteria

- [ ] Aggregate state uses a bounded five-minute design (300 one-second buckets or benchmarked equivalent), not an unbounded event list.
- [ ] Snapshot has latest value/time, count, sum, average, min, max, window bounds, and threshold status.
- [ ] Events over 60 seconds behind system time persist but do not mutate live aggregate state.
- [ ] Temperature, voltage, battery, and pressure boundaries match every exact PRD inclusion/exclusion.

## Out of scope

- Redis I/O, historical charts, event-time correction, or automated threshold alerts.

## Verification

- Write time-window and exact-boundary tests first; run race and memory-boundedness checks.

## Build-gate trace

| Contract | Source | Implementation | Repeatable proof |
| --- | --- | --- | --- |
| Live state is a five-minute bounded, processing-time window rather than a raw-event list | PRD §6.5; `docs/architecture.md` System shape and non-functional decisions | `api/internal/telemetry/aggregate.go`: `[300]aggregateBucket`, one bucket per processed second | Red: `TestAggregateEngineBuildsFiveMinuteProcessingTimeSnapshot` initially failed because the aggregate engine did not exist. Green: it proves count, sum, average, min, max, latest data, and the 300-second bounds; `TestAggregateEngineExpiresBucketsOutsideFiveMinuteWindow` proves expired data is absent. Static inspection confirms no event slice is retained. |
| A snapshot supplies the complete cache-publication payload and exact threshold status | PRD §6.5 and §7.3; `docs/Architecture.md` Redis aggregate sketch | `AggregateSnapshot`, `snapshotLocked`, and `thresholdStatus` | `TestAggregateEngineBuildsFiveMinuteProcessingTimeSnapshot` asserts every derived numeric field and timestamp/window bound. `TestThresholdStatusUsesExactPRDBoundaries` covers every normal/warning/critical boundary for temperature, voltage, battery, and pressure. |
| Events more than 60 seconds late remain persistence candidates but never mutate live state | PRD §6.5; `docs/Architecture.md` processing-time/late-event decision | `AggregateEngine.Process` skips only live mutation; `api/cmd/api/application.go` calls aggregation before the independent persistence processor | `TestAggregateEngineSkipsEventsLateAtWorkerProcessingTime` and `TestAggregateEngineSkipsLateEventWithSubsecondWorkerClock` prove exclusion past the exact boundary; `TestAggregateEngineAcceptsEventAtLateEventBoundary` proves exactly 60 seconds remains eligible. Docker-backed `TestApplicationRejectsValidBatchBeforeQueueWhenRetryGuardIsUnavailable` proves a 61-second-old event reaches PostgreSQL while no voltage snapshot is created. |
| The production worker composes aggregate work before the persistence seam without Redis I/O | `docs/architecture.md` system shape; Ticket 12 out of scope | `api/cmd/api/application.go` queue processor | Docker-backed `TestApplicationRejectsValidBatchBeforeQueueWhenRetryGuardIsUnavailable` admits an authenticated batch and waits for its in-memory aggregate. No Redis dependency or read was added. |
| Shared aggregate state remains race safe under concurrent workers/readers | Ticket 12 verification; `docs/conventions.md` concurrency rules | `AggregateEngine.mu` guards its window map and buckets | `go test -count=1 -race ./internal/telemetry` and the full `go test -count=1 -race ./...` pass. |

Quality evidence (2026-09-12):

- `go test -count=1 ./cmd/api -run TestApplicationRejectsValidBatchBeforeQueueWhenRetryGuardIsUnavailable` — passed against Docker-backed PostgreSQL.
- `go test -count=1 -race ./...`, `go vet ./...`, formatting, and `git diff --check` — passed.

## Review

Approved (2026-09-12). The implementation has a fixed 300-slot one-second bucket array per aggregate key; snapshots iterate that fixed structure and never retain raw events. The worker composition preserves the required cache-first/persistence-independent flow and adds no Redis I/O or public endpoint, both outside this ticket's scope. Exact thresholds, five-minute expiry, 60-second late-event handling (including the subsecond boundary), and the persistence-without-live-mutation lifecycle are covered by unit and Docker-backed application tests. The review found and corrected the original rounded-clock late-event comparison before approval. Full race tests, vet, formatting, and diff checks pass.
