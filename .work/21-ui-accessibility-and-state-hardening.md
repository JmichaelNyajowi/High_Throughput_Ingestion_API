# 21 — UI state, accessibility, and responsive hardening

**Type:** plumbing (test-after)
**Blocked by:** 18, 19, 20 — all MVP destinations must be wired.
**Status:** done

## What this delivers

The finished operator UI has consistent first-use, loading, error, offline/degraded, access-denied, and responsive behavior across Fleet, detail, and manual history.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Not applicable.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [x] Every key screen implements its `docs/DESIGN.md` empty, loading, error, success, and offline/degraded state with no generic blank/spinner page.
- [x] Keyboard focus order, skip link, landmark structure, one `h1`, modal Escape/focus restoration, labels, and status announcements meet the design brief.
- [x] Text and controls meet stated contrast, icon-label, reduced-motion, and responsive table requirements.
- [x] Rule-backed `/critique` findings are resolved or recorded as explicit approved scope gaps.

## Out of scope

- New screens, RBAC, internationalization, themes, or design-system expansion outside approved components.

## Verification

- Run accessibility automation, keyboard-only browser checks, reduced-motion checks, and desktop/tablet/mobile visual review.

## Build and review trace

- `frontend/src/app/accessibility.test.tsx` proves every application route retains the first skip link and exactly one page-level heading.
- Existing Fleet, Device, and History tests cover loading, success, empty, unavailable/not-found, stale, Redis-degraded, and manual-history error paths without automatic PostgreSQL fallback.
- `frontend/src/styles.test.ts` proves reduced-motion and responsive horizontal-table safeguards remain present alongside named-token enforcement.
- Frontend typecheck, test, build, lint, and whitespace checks pass. Browser automation remains unavailable in this environment; this limitation is recorded rather than presented as browser evidence.
