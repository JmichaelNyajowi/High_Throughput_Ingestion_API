# Telemetry UI Design Reference

`docs/DESIGN.md` is binding and `docs/screens.md` defines the implemented screen contract. This reference explains how to apply them without importing generic framework assumptions.

- The frontend is React, TypeScript, Vite, and plain CSS. Its named visual tokens live in `frontend/src/styles.css`; do not prescribe Tailwind, `@theme`, shadcn, Radix, or a `components/layout/` directory unless a later approved decision adds them.
- Use dense operator views: current telemetry and state precede charts; tables remain native tables and may scroll horizontally on mobile.
- Treat keyboard use and failure-state truthfulness as product behavior. A focusable region needs a name, visible focus, and a purpose; recovery routes must not throw on malformed input.
- Preserve the no-live-PostgreSQL rule in UI behavior. Manual history is explicit; Redis degradation keeps last-known data marked stale rather than inventing a fallback.
- Use a new shared component only after the design document names it or two approved use cases establish a stable semantic interface.
- Validate rendered contrast, responsive behavior, and screen-reader semantics. Do not substitute a static build or an a11y-addon configuration for those checks.
