# 12 — Five-minute aggregate and threshold engine

**Type:** logic (test-first)
**Blocked by:** 08, 09 — ordered shard processing and accepted events are required.
**Status:** planned

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

