---
name: build
description: Implement one eligible ticket with the project’s documented test and quality gates, then prepare it for independent review.
---

# Build

Read the selected `.work/` ticket, `CONTEXT.md`, `docs/architecture.md`, `docs/conventions.md`, `docs/setup.md` when present, relevant ADRs, and the existing module. Confirm its blockers are done and change `Status` to `in-progress`.

Use commands and project patterns documented in `docs/conventions.md`; never assume a package manager, framework, test runner, source layout, or port. For `logic (test-first)`, write executable failing tests from acceptance criteria before implementation. For plumbing, implement narrowly and add the tests the project convention requires. Respect documented module interfaces and do not access another module’s internals.

Implement only the stated scope, including lifecycle and failure behavior. If UI changes, follow `docs/design.md`, `docs/screens.md`, and the UI rules; do not invent a screen or a reusable primitive without a design decision. Run the documented tests, lint, type checks, migration checks, and appropriate manual or browser verification. Record evidence in the ticket.

Open or prepare a pull request and leave the ticket `Status: in-review`. `/build` never merges its own work. Hand it to `/verify` when an interactive flow warrants runtime evidence, then to `/review` before merge.
