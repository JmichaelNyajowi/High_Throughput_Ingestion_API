# Ticket Format

Used by `/tickets`, `/build`, `/verify`, `/critique`, and `/review`. A ticket is a narrow, end-to-end behavioral slice—not a technical layer—and must remain independently demonstrable and reviewable.

```markdown
# NN — Short ticket title

**Type:** logic (test-first) | plumbing (test-after)
**Blocked by:** <ticket numbers, or "None" with a reason>
**Status:** planned | in-progress | in-review | done

## What this delivers

<Concrete product behavior that becomes true.>

## Lifecycle

- **Create:** <how, or not applicable>
- **Read:** <how, or not applicable>
- **Update:** <what can change, by whom, or not allowed>
- **Delete:** <soft/hard/not allowed and referential effect>
- **Undo:** <how reversal works, or not available>

## Acceptance criteria

- [ ] <observable, specific, verifiable behavior>

## Out of scope

- <deliberately excluded neighboring behavior>

## Verification

- <tests, checks, and runtime evidence required before review>

### Build-gate trace

| Requirement or binding rule | Governing source | Implementation location | Repeatable proof |
| --- | --- | --- | --- |
| <criterion> | <document and section> | <file/module> | <test/check/evidence> |
```

`Lifecycle` is mandatory when a ticket creates or changes a domain record. State `not applicable` for each lifecycle action only when the ticket has no domain record or action. A historical or auditable record may correctly prohibit edit, delete, and undo; those are decisions to write down, not omissions.

Use vocabulary from `CONTEXT.md`. Specify contracts, behavior, and state boundaries, not brittle file paths or line numbers. For `logic (test-first)`, acceptance criteria must translate directly into failing tests written before implementation. The build-gate trace is completed during `/build`; it must cover every acceptance criterion and applicable design, contract, security, or deployment rule before review. Number tickets in dependency order and use `Blocked by` to make parallel-safe work explicit.
