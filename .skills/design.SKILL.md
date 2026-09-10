---
name: design
description: Produce or extend an evidence-based design system, screen inventory, flows, states, and accessibility contract for a UI product.
---

# Design

Read `reference.PROJECT.md`, `CONTEXT.md`, `docs/scope.md`, both architecture documents, current UI code, `docs/DESIGN.md`, `docs/screens.md`, and the UI references. Project design decisions override reference defaults.

For a new UI product, produce `docs/DESIGN.md` and `docs/screens.md`: design principles, visual direction, tokens, screen inventory, flows, per-screen hierarchy, reusable components, all key states, responsive rules, and accessibility requirements. For a later feature, make a scoped addition that reuses existing tokens and patterns rather than reopening settled visual decisions.

Select UI tooling only if planning has not already chosen it. Use the approved stack and component primitives; do not assume a particular framework, CSS system, Storybook, or icon set. Every required state must be representable in implementation and testable. Hand the designed feature to `/tickets`; do not implement production features here.
