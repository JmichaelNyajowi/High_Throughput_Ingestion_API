# 08 — Bounded device-sharded admission queue

**Type:** logic (test-first)
**Blocked by:** 03 — use the runtime lifecycle and shutdown model.
**Status:** done

## What this delivers

The synchronous admission path hands accepted batches to bounded, fixed workers while retaining per-device order and explicit overload behavior.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Queue depth/capacity is exposed to runtime monitoring.
- **Update:** Queue and shard counts are validated configuration values at startup.
- **Delete:** Queue contents drain only during controlled shutdown; no business record is deleted.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Device ID hashing consistently selects one owning shard so admitted batches for that device are processed in order.
- [ ] Shard/worker counts and capacity are bounded, validated startup values; no event/request creates an unbounded goroutine.
- [ ] Full capacity produces immediate typed `503` admission rejection and observable queue metrics.
- [ ] Shutdown stops admission and drains only within the configured bound.

## Out of scope

- Persistence, retries, aggregate calculation, or durable queue guarantees.

## Verification

- Write queue order/full/shutdown tests first and run them with the race detector.

## Build-gate trace

| Requirement / binding rule | Implementation | Repeatable proof |
| --- | --- | --- |
| A device consistently maps to one owner shard and its admitted batches retain FIFO processing order; Architecture.md §7.2 | `api/internal/telemetry/queue.go:DeviceShardedQueue`, `shardFor`, and one worker per shard | Red: `go test ./internal/telemetry -run TestDeviceShardedQueueProcessesDeviceBatchesInAdmissionOrder` failed before the queue types existed. Green: the same test blocks the first batch, admits a second batch for the same device, and proves `first` then `second` processing. |
| Worker/shard count and per-shard capacity are fixed, validated startup bounds without request-created goroutines; PRD §6.4, Techstack.md backend rule, and Architecture.md §7.2 | `QueueConfig.Validate`, `config.AdmissionQueueConfig`, `TELEMETRY_SHARD_COUNT`, and `TELEMETRY_SHARD_QUEUE_CAPACITY` | `TestQueueConfigRejectsUnboundedWorkerAndCapacityValues` and `TestLoadProvidesValidatedBoundedAdmissionQueueConfiguration`; API startup rejects 0/out-of-range values and defaults to 4 workers × 64 batches. |
| A full queue rejects immediately as typed `503` and exposes depth/capacity; ticket acceptance criterion and OpenAPI `ServiceUnavailable` | `DeviceShardedQueue.Admit`, `AdmissionError`, and `QueueMetrics` | `TestDeviceShardedQueueRejectsFullShardWithTypedUnavailableErrorAndMetrics` proves no blocking/full admission, `503`/`unavailable`, and queue capacity/depth visibility. Prometheus export is deliberately deferred to Ticket 15. |
| Shutdown first stops admission and drains only until the configured deadline; PRD §6.4 and Ticket 03 runtime lifecycle | `DeviceShardedQueue.Pause` and `Shutdown(context.Context)` | `TestDeviceShardedQueueShutdownPausesAdmissionAndHonoursDrainDeadline` proves immediate pause, deadline return while work is blocked, then successful drain after release. The queue satisfies Ticket 03's `AdmissionGate` and `Shutdowner` interfaces for later composition. |
| A worker panic does not remove its fixed processing capacity or leak request data; PRD §6.4 and conventions secret-handling rule | `processBatch` recovery, safe log message, and atomic panic counter | `TestDeviceShardedQueueRecoversWorkerPanicWithoutLosingWorkerCapacity` proves the next batch is processed by the same worker capacity and reports one recovered panic without logging panic payload/details. |

Preflight inspected `api/openapi.yaml`, `deploy/Caddyfile`, `deploy/compose.yaml`, `docs/database.md`, `docs/architecture.md`, and `docs/Architecture.md`. The queue introduces no synchronous Redis/PostgreSQL work and no external broker. `deploy/compose.yaml` passes the validated queue settings into the API; `.env.example` and `docs/setup.md` document their benchmark baseline and safety limits. Quality gates passed: `go test -count=1 -race ./...`, `go vet ./...`, `test -z "$(gofmt -l .)"`, and `git diff --check`. Docker/Compose runtime validation cannot run because Docker is unavailable in this environment.

## Review

Approved. Independent review confirmed the Telemetry-module queue consumes only already-validated batches, maps by authenticated device ID, uses exactly one FIFO worker per shard, and has no synchronous PostgreSQL/Redis or broker dependency. Its global admission lock prevents send-on-closed races while retaining non-blocking `503` overload rejection. Both startup configuration and direct construction cap the total queue at 1,024 batches, closing the potential memory-safety gap from independently valid per-shard values.

The queue exposes only safe depth/capacity/panic counts; worker recovery logs no panic payload or request data. It implements the existing runtime pause/shutdown seams and correctly stops new admission before bounded draining. Full independent verification passed: `go test -count=1 -race ./...`, `go vet ./...`, `gofmt` cleanliness, and `git diff --check`. Docker/Compose runtime verification remains unavailable in this environment.
