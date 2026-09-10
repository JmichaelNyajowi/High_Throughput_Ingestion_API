---
name: verify
description: Exercise a completed feature in a real running environment and report reproducible runtime evidence.
---

# Verify

Read the ticket, acceptance criteria, test data instructions, and documented local commands. Start only the services required by the project’s setup documentation. Reuse saved test authentication where the project has it; do not use real credentials or introduce them into test artifacts.

Exercise the happy path, declared lifecycle actions, role boundaries, and meaningful empty, loading, error, offline, and long-content states. Use stable semantic selectors or documented test IDs. Wait for visible readiness conditions or network completion—never fixed sleeps. Capture screenshots or equivalent evidence for interactive flows.

Report pass/fail against each checked criterion, actual versus expected behavior, exact repro steps for failures, and the evidence location. Verification provides runtime evidence; it does not approve a change, redesign the UI, or replace `/review`.
