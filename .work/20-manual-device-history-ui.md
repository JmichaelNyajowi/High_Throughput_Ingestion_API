# 20 — Explicit Device history workflow

**Type:** logic (test-first)
**Blocked by:** 01, 17 — the shell route and bounded history contract must exist.
**Status:** planned

## What this delivers

Operators can choose a valid bounded history range and view immutable raw telemetry only after explicitly running the query.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Operator submits a range/filter form to read immutable history.
- **Update:** Operator may change unsubmitted form criteria; submitted criteria remain visible with results.
- **Delete:** Not applicable.
- **Undo:** Reset restores the initial no-query state without deleting server data.

## Acceptance criteria

- [ ] Initial route clearly states that history is manual and does not call PostgreSQL until form submission.
- [ ] Visible labels and client/server validation enforce documented range/page limits; server errors preserve typed input.
- [ ] Results use an accessible dense table with predictable ordering, pagination/limit context, and real links to device detail.
- [ ] Empty, loading, error, access-denied, long-value, and no-results states are distinct and follow the design brief.

## Out of scope

- Auto-refresh, export, saved searches, long-range analytics, or a fallback for Redis outage.

## Verification

- Write form/request-state tests first; run browser checks that prove no request before submit and no refresh after results.

