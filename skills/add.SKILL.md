---
name: add
description: Scope a newly requested feature against existing product and architecture decisions before it enters ticketing.
---

# Add

Log the request in `docs/feature-requests.md` when that intake exists, or create the file with a simple dated request record when it is important and absent. Read it alongside `CONTEXT.md`, `docs/scope.md`, `docs/architecture.md`, `docs/conventions.md`, and existing code.

Reflect the request, identify conflicts with existing scope, architecture, data, security, or operations, then ask only the unresolved material questions with recommended defaults. Append the agreed feature and definition of done to `docs/scope.md`; preserve prior scope as historical record. Record any architecture change as an explicit decision or ADR rather than silently editing around it.

Use a scoped `/design` pass for a new or changed UI surface; then hand one confirmed feature to `/tickets`. Do not implement the feature here.
