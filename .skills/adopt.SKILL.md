---
name: adopt
description: Reconcile an existing coded project with the reusable workflow without re-planning or rebuilding shipped behavior.
---

# Adopt

Use once for a project that already contains meaningful code. Read `reference.PROJECT.md`, then inventory the codebase, CI, environments, dependencies, data model, deployment edge, release process, and existing documentation. When documentation conflicts with verified code or runtime behavior, document reality unless the project owner explicitly decides to change the code through a later ticket.

Create or update the applicable canonical documents from `reference.PROJECT.md`, including both architecture documents when system behavior is covered and `docs/DESIGN.md` plus `docs/screens.md` for UI work. Mark shipped work as done; do not disguise it as planned. Document established patterns with real examples, flag inconsistencies as decisions, and add missing quality or safety gaps without replacing functioning project tooling unnecessarily.

Create `AGENTS.md` or update the project’s existing instructions to point to reconciled docs and `.skills/README.md`. Set up issue/request/gotcha intake only if the project has no existing authoritative tracker. Then continue with `/tickets`, `/add`, or `/fix` as appropriate.
