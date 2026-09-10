---
name: foundation
description: Establish the documented development foundation before product tickets are implemented.
---

# Foundation

Read `CONTEXT.md`, `docs/architecture.md`, `docs/scope.md`, `docs/conventions.md`, `docs/release.md`, and any ADRs. Create missing structural files and folders that these documents deterministically require; pause when their contents require a product or operational decision.

Set up the stack selected by the project, not a framework assumed by this skill:

- reproducible local setup and documented commands in `docs/setup.md`;
- environment-variable validation, secret handling, and example configuration without real secrets;
- lint, formatting, type checking, test runner, and CI quality gates;
- migrations and safe local seed/fixture data where the product has persistent data;
- health checks, structured logging, error boundaries, and dependency configuration;
- an architecture-boundary check appropriate to the selected language;
- `AGENTS.md` pointing agents to `skills/README.md`, `CONTEXT.md`, and canonical project docs.

Do not implement a product feature merely to prove the scaffold. Verify each configured command actually runs. Hand UI products to `/design`; otherwise hand the next feature to `/tickets`.
