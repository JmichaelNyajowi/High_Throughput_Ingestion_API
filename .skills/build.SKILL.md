---
name: build
description: Implement one eligible ticket with the project’s documented test and quality gates, then prepare it for independent review.
---

# Build

Read `reference.PROJECT.md`, `.work/README.md`, the selected unblocked ticket, and every source named by the ticket’s acceptance criteria. At minimum read `CONTEXT.md`, `docs/prd.md`, `docs/architecture.md`, `docs/Architecture.md`, `docs/conventions.md`, and `docs/setup.md`. Confirm blockers are `done`, then set the ticket to `in-progress`.

Implement only stated scope, lifecycle, and failure behavior. Respect the module seams in `docs/architecture.md`. For `logic (test-first)`, write one executable failing behavior test from an acceptance criterion, observe failure, implement the minimum change, and repeat. For plumbing, add seam-level tests required by the ticket and conventions.

Run the change-sensitive preflight in `reference.PROJECT.md` before quality commands: UI work must use `docs/DESIGN.md` and `docs/screens.md`; HTTP, authentication, metrics, health, or endpoint work must inspect `api/openapi.yaml`, `deploy/Caddyfile`, and `deploy/compose.yaml`; data or concurrency work must inspect `docs/database.md`. Do not create a lower-case `docs/design.md` or invent missing screen/component patterns.

Run the exact relevant commands in `reference.PROJECT.md`. Record a build-gate trace mapping every acceptance criterion and binding rule to its source, implementation location, and repeatable proof. Add a regression test for each corrected review finding when testable. Do not set `in-review` while a trace discrepancy remains.

Prepare a pull request without merging, leave the ticket `in-review`, and hand interactive work to `/verify` before `/review`.
