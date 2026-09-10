# 01 — Application shell and monitoring destinations

**Type:** plumbing (test-after)
**Blocked by:** Foundation bootstrap complete; the React application, component tooling, Storybook, and quality gates must exist before this ticket can be built.
**Status:** planned

## Goal

Operators can navigate stable Fleet, Device, and Device history destinations that render the approved application shell and truthful pending/empty states before live data features are wired.

## Context

- Relevant module: Fleet owns the live monitoring destinations; History owns the manual history destination.
- Relevant docs: `docs/architecture.md` module boundaries, `docs/DESIGN.md` shared shell and screen layouts, `docs/screens.md` Operations screens, and `docs/conventions.md` frontend conventions.
- Relevant approved design states: Fleet overview first-use/loading, Device detail missing aggregate/loading, Manual device history pre-query/loading, access-denied, and not-found.
- Relevant prior work: Stage 0 Foundation bootstrap, which creates the application and Storybook scaffolding.

## Scope

**In:**

- The shared dark graphite app shell, top bar, skip link, main content frame, system-mode presentation, and responsive behavior specified by `docs/DESIGN.md`.
- Navigable Fleet, Device detail, Device history, access-denied, and not-found destinations.
- Design-faithful pending/empty/unknown screen states using fixture-shaped placeholders only; no real backend calls.
- Storybook coverage for the shell and each destination’s required initial state.
- Real links, keyboard navigation, page titles, focus treatment, and route fallback behavior.

**Out:**

- Device API-key authentication, Caddy access policy, backend routes, live polling, persistence, aggregate charts with real data, history queries, or any new product feature.
- A permanent sidebar, settings screen, user management, device provisioning, or alert configuration.

## Acceptance criteria

- [ ] Fleet is the default route and renders the approved shell, page header, Fleet status-strip skeleton, filter-toolbar skeleton, and table skeleton without layout shift.
- [ ] Device detail and Manual device history routes render their approved context/header and explicit pending/pre-query states rather than 404 pages.
- [ ] Unknown application routes render the approved not-found state with a real `Return to fleet` link.
- [ ] Each destination has one `h1`, a working skip link, visible keyboard focus, and tab order matching visual order.
- [ ] Desktop, tablet, and mobile shell behavior follows `docs/DESIGN.md`; tables remain recognizable tables on narrow screens.
- [ ] No raw colors, arbitrary spacing, decorative emoji, gradients, or unapproved primitives are introduced.
- [ ] Storybook shows the app shell and every destination state implemented by this ticket.

## Verification

- Run frontend formatting, type checking, linting, and relevant component tests.
- Verify each Storybook story against `docs/DESIGN.md`, including desktop, mobile, keyboard-only, and reduced-motion behavior.
- Run a browser check for default route, direct deep links, unknown route recovery, visible focus, and skip-link behavior.
