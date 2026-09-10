---
name: review
description: Independently assess one in-review ticket before merge against its contract, architecture, security, tests, and UI rules.
---

# Review

Review in a fresh context. Read the ticket, its diff, `CONTEXT.md`, `docs/architecture.md`, `docs/conventions.md`, ADRs, and the applicable design docs. Confirm the ticket is `in-review`; do not use review to silently expand scope.

Check:

- each acceptance criterion, lifecycle decision, and stated non-goal;
- correct use of documented module boundaries, interfaces, data ownership, and public APIs;
- authorization, validation, secret handling, error handling, migration safety, and observability appropriate to the change;
- tests at the seam specified by the architecture and required test-first evidence for logic;
- UI token/pattern/state/accessibility compliance for UI changes;
- no avoidable duplication, dead code, undocumented new dependency, or mismatch with the release plan.

Approve only when blocking findings are resolved. On approval, merge according to the repository process and set `Status: done`. On rejection, list concrete file-level findings, set `Status: in-progress`, and return it to `/build`. This is the required pre-merge gate for ticketed work.
