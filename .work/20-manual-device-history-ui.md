# 20 — Explicit Device history workflow

**Type:** logic (test-first)
**Blocked by:** 01, 17 — the shell route and bounded history contract must exist.
**Status:** done

## What this delivers

Operators can choose a valid bounded history range and view immutable raw telemetry only after explicitly running the query.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Operator submits a range/filter form to read immutable history.
- **Update:** Operator may change unsubmitted form criteria; submitted criteria remain visible with results.
- **Delete:** Not applicable.
- **Undo:** Reset restores the initial no-query state without deleting server data.

## Acceptance criteria

- [x] Initial route clearly states that history is manual and does not call PostgreSQL until form submission.
- [x] Visible labels and client/server validation enforce documented range/page limits; server errors preserve typed input.
- [x] Results use an accessible dense table with predictable ordering, pagination/limit context, and real links to device detail.
- [x] Empty, loading, error, access-denied, long-value, and no-results states are distinct and follow the design brief.

## Out of scope

- Auto-refresh, export, saved searches, long-range analytics, or a fallback for Redis outage.

## Verification

- Write form/request-state tests first; run browser checks that prove no request before submit and no refresh after results.

## Build and review trace

- `frontend/src/modules/history/history.test.tsx` was written first and observed failing before the manual submission workflow was implemented. It proves no request before submission, UTC request conversion, and the no-results state.
- `frontend/src/modules/history/index.tsx` has no polling or live-query client. It preserves form fields on error, names 401/503 conditions, renders loading/empty/results states, and uses real device links with an accessible native table.
- `api/internal/history/history.go` and `api/cmd/api/application.go` now apply the contract's optional `measurement_type` filter with a parameterized SQL predicate; the UI does not offer a filter the API ignores.
- Frontend typecheck, 15 tests, production build, lint, and whitespace validation pass. The complete Go test suite passes after the history filter change. Browser automation is unavailable in this environment; no browser result is claimed.
- Independent review found no remaining ticket-level blocker. No PostgreSQL fallback, auto-refresh, export, or long-range analytics was introduced.
