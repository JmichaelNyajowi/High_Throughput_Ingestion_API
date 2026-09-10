---
name: fix
description: Triage a documented defect into a proportional, verifiable delivery path.
---

# Fix

Read `docs/bugs.md` or the project’s documented issue intake. Ensure the report includes severity, impact, reproducible steps, expected behavior, actual behavior, and status. For each defect, recommend either the full `/tickets` → `/build` → `/review` → `/release` path or a direct fix based on blast radius.

The full path is required for business logic, authorization, data integrity, security, or non-trivial changes, and its ticket must include a regression criterion. A direct fix is permitted only for clearly isolated, low-risk changes and must still be tested, self-verified, and released under `docs/release.md`.

Update the defect status and append concise root-cause prevention notes to `docs/gotchas.md` when the lesson will help later work. Do not silently choose a lower-rigor path for a material defect.
