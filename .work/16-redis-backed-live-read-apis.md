# 16 — Redis-backed fleet and device read APIs

**Type:** logic (test-first)
**Blocked by:** 02, 13, 14, 15 — contract, cache model, degradation semantics, and runtime state must exist.
**Status:** planned

## What this delivers

The dashboard can fetch fleet and device live snapshots from Redis with clear freshness and degraded semantics, never through PostgreSQL refresh queries.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Operator requests read fleet/device snapshots from Redis-derived state.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Fleet response returns device summaries, active/stale/warning/critical counts, mode, and refresh context.
- [ ] Device response returns latest values, five-minute aggregates, threshold state, freshness, and aggregate window bounds.
- [ ] Cache misses are `unknown`/stale, never represented as healthy; Redis outage returns documented degraded response.
- [ ] Normal live read execution makes no PostgreSQL `SELECT` call.

## Out of scope

- Manual history, caching in the browser, UI rendering, or WebSockets.

## Verification

- Write HTTP tests first with Redis and PostgreSQL spies/integration evidence proving the no-SELECT contract.

