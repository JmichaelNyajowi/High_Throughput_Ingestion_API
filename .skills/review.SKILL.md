---
name: review
description: Independently assess one in-review ticket before merge against its contract, architecture, security, tests, and UI rules.
---

# Review

Review in a fresh context. Read `reference.PROJECT.md`, the ticket, its diff, `CONTEXT.md`, both architecture documents, `docs/conventions.md`, and applicable product, design, deployment, and contract sources. Confirm the ticket is `in-review`; do not use review to silently expand scope.

Check:

- each acceptance criterion, lifecycle decision, and stated non-goal;
- correct use of documented module boundaries, interfaces, data ownership, and public APIs;
- authorization, validation, secret handling, error handling, migration safety, and observability appropriate to the change;
- tests at the seam specified by the architecture and required test-first evidence for logic;
- UI token/pattern/state/accessibility compliance for UI changes;
- no avoidable duplication, dead code, undocumented new dependency, or mismatch with the release plan.

Verify the ticket’s build-gate trace independently. For UI work, inspect `docs/DESIGN.md` and `docs/screens.md`, then check declared routes (including malformed/unknown paths), focus order, labelled focusable regions, states, and breakpoints. For HTTP, authentication, metrics, health, or endpoint work, compare `api/openapi.yaml` with handlers and `deploy/Caddyfile`/`deploy/compose.yaml`; reject a contract whose stated access boundary is not enforced at the edge.

Approve only when blocking findings are resolved. On approval, merge according to the repository process and set `Status: done`. On rejection, list concrete file-level findings, set `Status: in-progress`, and return it to `/build`. This is the required pre-merge gate for ticketed work.
