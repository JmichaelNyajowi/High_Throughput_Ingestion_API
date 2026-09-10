# 21 — UI state, accessibility, and responsive hardening

**Type:** plumbing (test-after)
**Blocked by:** 18, 19, 20 — all MVP destinations must be wired.
**Status:** planned

## What this delivers

The finished operator UI has consistent first-use, loading, error, offline/degraded, access-denied, and responsive behavior across Fleet, detail, and manual history.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Not applicable.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Every key screen implements its `docs/DESIGN.md` empty, loading, error, success, and offline/degraded state with no generic blank/spinner page.
- [ ] Keyboard focus order, skip link, landmark structure, one `h1`, modal Escape/focus restoration, labels, and status announcements meet the design brief.
- [ ] Text and controls meet stated contrast, icon-label, reduced-motion, and responsive table requirements.
- [ ] Rule-backed `/critique` findings are resolved or recorded as explicit approved scope gaps.

## Out of scope

- New screens, RBAC, internationalization, themes, or design-system expansion outside approved components.

## Verification

- Run accessibility automation, keyboard-only browser checks, reduced-motion checks, and desktop/tablet/mobile visual review.

