# 19 — Redis-backed Device detail

**Type:** plumbing (test-after)
**Blocked by:** 01, 16 — shell/deep links and device live API must exist.
**Status:** planned

## What this delivers

Operators can investigate one device’s current readings, five-minute aggregate summaries, threshold context, freshness, and a deliberate path to history.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Dashboard polling reads only the device live API.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Detail header shows device identity, freshness, threshold state, last processed time, and real navigation back to Fleet/history.
- [ ] Measurement panels/charts show approved latest/aggregate values, window bounds, legends, and text alternatives; never imply data is fresh when stale or degraded.
- [ ] Unknown device, loading, API error, stale, and Redis-degraded states are explicit and preserve usable navigation.
- [ ] Responsive layout follows the desktop/tablet/mobile contract without horizontal loss of essential state.

## Out of scope

- Editing a device, configuration, alert acknowledgements, historical querying, or synthetic client-side aggregates.

## Verification

- Run UI tests and browser checks for deep link, keyboard navigation, long IDs, stale/degraded, and not-found recovery.

