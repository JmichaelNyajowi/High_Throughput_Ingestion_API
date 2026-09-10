---
name: care
description: Identify and rank architectural friction between delivery tranches, then take one selected improvement through ticketing.
---

# Care

Read `reference.PROJECT.md`, `reference.MODULES.md`, `CONTEXT.md`, both architecture documents, recent changes, deployment configuration when relevant, and tests. Look for observed friction: unclear ownership, cross-module internal access, interfaces that expose little value, seams that cannot be tested, recurring defects, or excessive navigation across files to understand one concept.

Write a dated report in the OS temporary directory, not the repository. For each candidate list the modules/files involved, observed problem, likely benefit, and confidence. Stop for the project owner to choose one; do not propose detailed interfaces for every candidate.

For the selected candidate, establish tests before refactoring, resolve any required vocabulary or ADR change explicitly, update architecture/convention docs, and hand the work to `/tickets`. Never refactor untested behavior or take on multiple structural changes in one pass.
