# 01 — Application shell and monitoring destinations

**Type:** plumbing (test-after)
**Blocked by:** None — Foundation bootstrap is complete.
**Status:** planned

## What this delivers

Operators can navigate stable Fleet, Device, and Device history destinations that render the approved application shell and truthful pending/empty states before live data features are wired. The shell ensures later feature tickets connect real data to known destinations instead of inventing routes, navigation, or screen state while implementing backend behavior.

## Lifecycle

This ticket creates no domain record.

- **Create:** Not applicable.
- **Read:** Not applicable.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Fleet is the default route and renders the approved shell, page header, Fleet status-strip skeleton, filter-toolbar skeleton, and table skeleton without layout shift.
- [ ] Device detail and Manual device history routes render their approved context/header and explicit pending/pre-query states rather than 404 pages.
- [ ] Unknown application routes render the approved not-found state with a real `Return to fleet` link.
- [ ] Each destination has one `h1`, a working skip link, visible keyboard focus, and tab order matching visual order.
- [ ] Desktop, tablet, and mobile shell behavior follows `docs/DESIGN.md`; tables remain recognizable tables on narrow screens.
- [ ] No raw colors, arbitrary spacing, decorative emoji, gradients, or unapproved primitives are introduced.
- [ ] Storybook shows the app shell and every destination state implemented by this ticket.

## Out of scope

- Device API-key authentication, Caddy access policy, backend routes, live polling, persistence, aggregate charts with real data, history queries, or any new product feature.
- A permanent sidebar, settings screen, user management, device provisioning, or alert configuration.

## Verification

- Run frontend formatting, type checking, linting, and relevant component tests.
- Verify each Storybook story against `docs/DESIGN.md`, including desktop, mobile, keyboard-only, and reduced-motion behavior.
- Run a browser check for default route, direct deep links, unknown route recovery, visible focus, and skip-link behavior.
