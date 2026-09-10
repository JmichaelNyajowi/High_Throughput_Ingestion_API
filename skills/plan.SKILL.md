---
name: plan
description: Turn discovery and existing evidence into explicit product, architecture, scope, convention, and release decisions.
---

# Plan

Read `reference.INTERROGATION.md`, `reference.MODULES.md`, prior discovery entries, existing `CONTEXT.md`, `docs/`, ADRs, and the codebase when present. Ask sequential questions where one decision constrains the next. Make a recommended default whenever a real choice remains, but do not invent material product, security, data, or operational requirements.

Produce or update:

- `CONTEXT.md` — stable domain vocabulary and key invariants.
- `docs/architecture.md` — module boundaries, interfaces, seams, stack decisions, and data/security flows.
- `docs/scope.md` — goals, non-goals, feature definitions of done, and explicit ambiguities.
- `docs/conventions.md` — verified commands, project structure, testing rules, and approved patterns.
- `docs/release.md` — environments, ownership, approval gates, migrations, rollback, and smoke checks.
- `docs/adr/` entries for consequential decisions that future work might otherwise reopen.

For a scoped later change, update only the affected concepts and document the decision history. Hand a buildable feature to `/foundation` if the platform is not ready, otherwise `/design` for UI work or `/tickets` for backend-only work.
