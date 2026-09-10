---
name: release
description: Deploy reviewed and merged work using the project’s documented release process and migration safety requirements.
---

# Release

Read `docs/release.md`; if it is missing, stop and route the gap to `/plan` or `/foundation`. Follow its environments, ownership, approvals, deployment commands, rollback plan, and smoke checks exactly. Never treat an assumed approval or an unverified CI result as a release.

For persistent-data changes, use backward-compatible expand/contract migrations unless the documented release process explicitly supports a safe alternative. Surface failed deploys, unsafe migrations, missing approval, or failed smoke checks as blockers; do not retry destructively or silently.

Record the deployed version, time, migration status, and smoke-check evidence where the project’s release documentation requires it.
