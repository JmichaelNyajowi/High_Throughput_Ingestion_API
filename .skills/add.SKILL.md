---
name: add
description: Scope a newly requested feature against existing product and architecture decisions before it enters ticketing.
---

# Add

Read `reference.PROJECT.md`, then inspect the relevant product, architecture, design, deployment, and existing-code sources. Log the request in `docs/feature-requests.md` when that intake exists, or create it only when the request needs durable intake.

Reflect the request, identify conflicts with existing scope, architecture, data, security, deployment, or operations, then ask only unresolved material questions. Update `docs/scope.md` and, when applicable, `docs/roadmap.md`; preserve prior decisions as history. Record an architecture or security-boundary change explicitly rather than silently editing around it.

Use a scoped `/design` pass for a new or changed UI surface; then hand one confirmed feature to `/tickets`. Do not implement the feature here.
