# 08 — Bounded device-sharded admission queue

**Type:** logic (test-first)
**Blocked by:** 03 — use the runtime lifecycle and shutdown model.
**Status:** planned

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

