# NN — Short Ticket Title

**Type:** logic (test-first) | plumbing (test-after)
**Blocked by:** Ticket numbers, or `None` with a reason.
**Status:** planned

## What this delivers

Describe the observable end-to-end behavior that becomes true. Use Telemetry vocabulary from `CONTEXT.md`. Explain why the behavior matters when the ticket carries a non-obvious product rule.

## Lifecycle

Required when the ticket creates or changes a record type.

- **Create:** How the record is created and by whom.
- **Read:** Where and how it can be viewed.
- **Update:** What can change, by whom, and under what rules.
- **Delete:** Soft delete, hard delete, or explicitly not allowed; state effects on referencing records.
- **Undo:** Whether and how the action is reversed.

For a ticket that creates no domain record, state `Not applicable` under every lifecycle action.

## Acceptance criteria

- [ ] Write observable, specific, verifiable behavior.
- [ ] For `logic (test-first)`, each criterion must translate into an independently failing behavior test before implementation.
- [ ] Include authorization, empty/error/degraded, and cross-screen consequences when they belong to the ticket.

## Out of scope

- State adjacent behavior deliberately excluded from this ticket.

## Verification

- List the focused tests, quality checks, and manual/browser checks required before review.
