# 13 — Redis live state and TTL publication

**Type:** logic (test-first)
**Blocked by:** 12 — aggregate snapshots must exist.
**Status:** planned

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

