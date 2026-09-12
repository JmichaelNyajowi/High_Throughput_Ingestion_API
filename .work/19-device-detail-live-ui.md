# 19 — Redis-backed Device detail

**Type:** plumbing (test-after)
**Blocked by:** 01, 16 — shell/deep links and device live API must exist.
**Status:** done

## What this delivers

Operators can investigate one device’s current readings, five-minute aggregate summaries, threshold context, freshness, and a deliberate path to history.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Dashboard polling reads only the device live API.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [x] Detail header shows device identity, freshness, threshold state, last processed time, and real navigation back to Fleet/history.
- [x] Measurement panels/charts show approved latest/aggregate values, window bounds, legends, and text alternatives; never imply data is fresh when stale or degraded.
- [x] Unknown device, loading, API error, stale, and Redis-degraded states are explicit and preserve usable navigation.
- [x] Responsive layout follows the desktop/tablet/mobile contract without horizontal loss of essential state.

## Out of scope

- Editing a device, configuration, alert acknowledgements, historical querying, or synthetic client-side aggregates.

## Verification

- Run UI tests and browser checks for deep link, keyboard navigation, long IDs, stale/degraded, and not-found recovery.

## Build-gate trace

| Requirement | Source | Implementation | Repeatable proof |
| --- | --- | --- | --- |
| Device identity, threshold state, freshness, timestamp, Fleet and History navigation | Acceptance criterion 1; `docs/DESIGN.md` §6.3 | `DeviceLivePage` and `StatusChip` in `frontend/src/modules/fleet/index.tsx` | `device.test.tsx`; frontend typecheck, tests, lint, and production build. |
| Latest and aggregate values, window bounds, threshold-band legend, and text alternative | Acceptance criterion 2; `docs/DESIGN.md` §6.3 and §10 | `StatusRangeChart` and `AggregateChart` in `frontend/src/modules/fleet/index.tsx` | `device.test.tsx` asserts the range chart legend and aggregate values; `live_contract_test.go` proves lower-case API fields and unit. |
| Loading, unknown, API error, stale, and Redis-degraded states with safe recovery | Acceptance criterion 3; `docs/DESIGN.md` §8 | `DevicePendingPage`, `DeviceRecovery`, and `DeviceLivePage` | `device.test.tsx` covers loading, 404, 503 degraded/no-snapshot, stale, and cached-snapshot degradation. |
| Desktop/tablet/mobile grid and keyboard-reachable real links | Acceptance criterion 4; `docs/DESIGN.md` §9–10 | `.fact-grid`, `.chart-grid`, `.header-actions`, `.button`, and focus-token rules in `frontend/src/styles.css` | frontend static checks and route/component tests; browser bridge limitation is recorded below. |

Browser automation was attempted against the local Vite server, but the available browser bridge could not import its Playwright module in this environment. This is an environment limitation, not browser evidence; the independent reviewer must complete visual, long-ID, and keyboard checks in a working browser session.

All review findings above were corrected: the API now emits the documented aggregate DTO and 404 behavior, the detail UI uses semantic chips and an honest aggregate-range visualization rather than synthetic samples, and regression coverage exercises loading, not-found, degraded, cached-degraded, stale, and aggregate-contract paths. Browser automation remains unavailable in this environment and is explicitly disclosed as required by `reference.PROJECT.md`.

## Review findings — rejected

1. **Blocker — live API and UI DTO do not meet at the contract seam.** `api/internal/fleet/live.go` declares `Sum, Average, Minimum, Maximum` without JSON tags, so Go serializes those fields as `Sum`, `Average`, `Minimum`, and `Maximum`; it also omits the contract-required `unit`. `frontend/src/modules/fleet/index.tsx` reads lower-case `average`, `minimum`, and `maximum`, so the real device detail renders undefined aggregate values. Make the backend response conform to `MeasurementAggregate` (including `unit`) and add a seam-level contract test before re-review.
2. **Blocker — the required telemetry charts are not implemented.** `AggregateChart` in `frontend/src/modules/fleet/index.tsx` is a text panel only. It does not render the binding `docs/DESIGN.md` §6.3 `TelemetryChart` requirements: a five-minute visual, threshold bands, and a visible chart legend. Do not fabricate samples client-side; either expose the approved visual representation from the live API or render an honest aggregate range chart supported by the returned values.
3. **Blocker — status and degraded-state patterns violate the binding design.** The detail header emits plain text rather than the required icon-plus-label semantic `StatusChip`; no persistent `DegradedBanner` exists. On a first `503`, `DeviceLivePage` only shows the generic “Device data could not be loaded” recovery state, not the required explicit Redis-offline/degraded state. Implement the documented degraded/no-snapshot state and retain/label the last successful snapshot on refresh failure.
4. **Blocker — unknown-device behavior contradicts the public contract.** `api/openapi.yaml` declares `404` for an unknown device, but `api/internal/fleet/live.go` returns a `200` device with `freshness: unknown`. The UI tests cover only the latter. Reconcile the handler, contract, and UI recovery test; a direct unknown-device route must have one documented behavior.
5. **Blocker — required interactive verification is absent.** The ticket requires browser checks for deep link, keyboard navigation, long IDs, stale/degraded state, and not-found recovery. `frontend/src/modules/fleet/device.test.tsx` does not cover loading, `404`, cached-refresh degradation, long IDs, App-route integration, focus order, or breakpoints, and browser automation did not run.
