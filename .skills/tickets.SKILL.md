---
name: tickets
description: Cut one feature into independently buildable tickets with explicit blockers, lifecycle decisions, scope boundaries, and verification.
---

# Tickets

Read `reference.PROJECT.md`, `reference.TICKET-FORMAT.md`, `CONTEXT.md`, `docs/scope.md`, both architecture documents, `docs/conventions.md`, `docs/roadmap.md`, and relevant design/screen docs. Inspect existing code and `.work/` tickets before continuing numbering. Create missing `.work/` when it is absent.

Work on the next confirmed feature only. Cut vertical slices that are independently demonstrable and reviewable; avoid horizontal tickets such as “build the API layer.” Classify business rules, validation, calculations, permissions, and state transitions as `logic (test-first)`; use `plumbing (test-after)` only for wiring or composition with no material behaviour to specify.

Every ticket must follow the canonical format exactly: `Type`, `Blocked by`, `Status`, `What this delivers`, `Lifecycle`, `Acceptance criteria`, `Out of scope`, and `Verification`. Record-changing tickets must decide create, read, update, delete, and undo explicitly. Each acceptance criterion must name its governing source in the build-gate trace template. For UI, HTTP, security, or deployment work, make the relevant `docs/DESIGN.md`, `api/openapi.yaml`, and `deploy/` checks explicit. Number files `.work/<NN>-<slug>.md`, state dependencies in `Blocked by`, and set new work to `Status: planned`.

Confirm the set before `/build`. Do not implement or pre-ticket unrelated roadmap work.
