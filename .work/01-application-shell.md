# 01 — Application shell and monitoring destinations

**Type:** plumbing (test-after)
**Blocked by:** None — Foundation bootstrap is complete.
**Status:** done

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

- [x] Fleet is the default route and renders the approved shell, page header, Fleet status-strip skeleton, filter-toolbar skeleton, and table skeleton without layout shift.
- [x] Device detail and Manual device history routes render their approved context/header and explicit pending/pre-query states rather than 404 pages.
- [x] Unknown application routes render the approved not-found state with a real `Return to fleet` link.
- [x] Each destination has one `h1`, a working skip link, visible keyboard focus, and tab order matching visual order.
- [x] Desktop, tablet, and mobile shell behavior follows `docs/DESIGN.md`; tables remain recognizable tables on narrow screens.
- [x] No raw colors, arbitrary spacing, decorative emoji, gradients, or unapproved primitives are introduced.
- [x] Storybook shows the app shell and every destination state implemented by this ticket.

## Out of scope

- Device API-key authentication, Caddy access policy, backend routes, live polling, persistence, aggregate charts with real data, history queries, or any new product feature.
- A permanent sidebar, settings screen, user management, device provisioning, or alert configuration.

## Verification

- [x] `npm run lint`, `npm run typecheck`, `npm run test`, and `npm run build` pass in `frontend/`.
- [x] Storybook static build succeeds with Fleet pending, Device pending, History pre-query, and not-found stories; the a11y addon is enabled.
- [x] Temporary Vite verification returned the SPA document for `/`, `/devices/gateway-17`, `/devices/gateway-17/history`, and `/missing`.
- [x] DOM-level tests prove the skip link is first in tab order, focuses the main landmark, each route has the expected heading and recovery/control semantics, and the Fleet navigation state is accurate.
- [x] Manual browser review completed by the product owner: visual keyboard/focus, reduced-motion, desktop, tablet, and mobile behavior were accepted.
- [x] `src/styles.test.ts` rejects raw component `rem`, `px`, `ms`, `em`, and percentage values, keeping dimensional values defined through named root tokens. Documented media-query thresholds and keyframe offsets are excluded because standard CSS cannot consume custom properties there.
- [x] After the token refactor, `npm run lint`, `npm run typecheck`, `npm run test`, `npm run build`, `npm audit --json`, `git diff --check`, and a static Storybook build all pass. The dependency audit reports zero vulnerabilities.
- [x] Regression tests cover malformed device routes, the named/described keyboard-focusable table scroller, and the mobile product-name rule. The initial red test run failed on all three findings before the implementation was corrected.
- [x] After the review-finding fixes, `npm run lint`, `npm run typecheck`, `npm run test` (6 tests), `npm run build`, `npm run storybook:build`, and `git diff --check` pass.
- [x] Browser automation is unavailable in this environment. The prior product-owner browser acceptance remains recorded above; the changed mobile rule has static CSS regression coverage and a successful Storybook build.

### Build-gate trace

| Requirement or binding rule | Governing source | Implementation location | Repeatable proof |
| --- | --- | --- | --- |
| Fleet, Device, History, and recovery destinations | This ticket; `docs/DESIGN.md` sections 4 and 6 | `frontend/src/app/`, `frontend/src/modules/fleet/`, `frontend/src/modules/history/` | Route/DOM tests and Storybook destination stories. |
| Unknown and malformed routes recover safely | This ticket; `docs/DESIGN.md` section 10 | `frontend/src/app/routes.ts` | `routes.test.ts` proves `/settings` and `/devices/%` resolve to `not-found`. |
| Loading/pre-query states remain truthful before live APIs exist | This ticket; `docs/architecture.md` data/lifecycle boundaries | Fleet and History pending modules | DOM destination tests and Storybook static build. |
| Keyboard, focus, and labelled responsive table behavior | This ticket; `docs/DESIGN.md` sections 9 and 10 | `AppShell.tsx`, Fleet pending table | DOM tests prove skip-link focus and named/described table region; CSS preserves native table scrolling. |
| Mobile retains product name and system state | `docs/DESIGN.md` section 9 | `AppShell.tsx`, `styles.css` | Mobile CSS regression test and Storybook static build. |
| Components consume named design tokens rather than raw dimensions | This ticket; `docs/DESIGN.md` section 3 | `frontend/src/styles.css` | Token-only CSS source regression test, production build, and Storybook static build. |
| No live polling, history query, credential, or backend feature is introduced | This ticket's out-of-scope list; `docs/architecture.md` module boundaries | Frontend shell modules only | Source review and no API/data-client additions. |

## Review findings

1. **Resolved — `frontend/src/styles.css` now defines named sizing, typography, border, motion, and utility tokens in `:root`, and component rules consume those tokens exclusively.** The static regression test prevents raw dimensional literals from returning. The only literal breakpoints remain in media-query conditions, where standard CSS cannot use custom properties; they match the documented tablet and mobile thresholds.
2. **Resolved — malformed device routes now recover.** `routes.ts` treats decoding failures as `not-found`, and a regression test covers `/devices/%`.
3. **Resolved — mobile preserves the visible product name and system state.** The mobile top bar wraps its content, retains the named brand element, and places navigation on its own row to avoid narrow-screen overflow. A CSS regression test protects the rule.
4. **Resolved — the horizontal table scroller is now a labelled, described region.** It remains keyboard focusable and announces both its Fleet-table purpose and horizontal-scroll instruction; the DOM test verifies that accessible contract.
5. **Resolved — the skip-link copy now matches the binding design rule.** It is the first focusable element, says `Skip to fleet content`, and retains its tested main-landmark focus behavior.
6. **Approved — independent review found no remaining ticket-level discrepancy.** Frontend lint, type checking, six tests, production build, Storybook static build, whitespace check, and dependency audit all pass; audit reports zero vulnerabilities. Repository merge/release activity was not performed in this workspace.
