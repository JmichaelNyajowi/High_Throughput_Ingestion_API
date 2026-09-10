# Telemetry UI Review Checklist

Use this only with `docs/DESIGN.md` and `docs/screens.md`; those documents win when they differ.

- Components consume named tokens from `frontend/src/styles.css`; raw colors and arbitrary component dimensions are not allowed. Media-query conditions may use the documented literal breakpoints because standard CSS custom properties cannot be used there.
- Each route has one `h1`, landmarks, a first-focusable skip link, visible focus, and visual/tab order alignment. Unknown and malformed routes render a recovery state rather than throwing.
- Mobile retains a visible product name and system state. Tables remain semantic tables in a labelled, keyboard-usable horizontal scroll region.
- Implement exactly the loading, empty, error, degraded/offline, and manual-history states specified in `docs/DESIGN.md`; never use live PostgreSQL fallback for dashboard refresh.
- Use only Lucide icons already approved by `docs/DESIGN.md`; icon-only controls need a visible-on-focus tooltip and an accessible name.
- Every UI ticket’s build-gate trace includes responsive, focus/keyboard, and state proof. Browser evidence is required when the environment supports it; otherwise record the limitation.
