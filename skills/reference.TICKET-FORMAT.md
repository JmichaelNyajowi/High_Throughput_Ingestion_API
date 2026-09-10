# Ticket Format

The ticket template and the rules that make a ticket good. Used by `/tickets`, `/build`, `/review`, and `/critique`.

## What a ticket is

A **vertical slice** — a narrow but complete path through every layer.

- Cuts through schema, logic, API, UI, and tests — **not one layer**
- Leaves the system **working and demoable** on its own
- Fits comfortably in one fresh session — **target ~250k tokens; ~350k means it was cut too coarsely**
- Declares what blocks it
- Declares **logic** (test-first) or **plumbing** (test-after or none)

A ticket like "build the API layer" is horizontal and wrong. There's no observable behaviour to test, nothing to demo, and errors compound silently because nothing can be verified until the layer above arrives.

## The template

```markdown
# NN — <title>

**Type:** logic (test-first) | plumbing (test-after)
**Blocked by:** <ticket numbers, or "None">

## What this delivers

<The end-to-end behaviour this makes work, from the user's perspective.
Not a layer-by-layer implementation list.>

## Lifecycle

<Only for tickets that create or modify a record type. See below.>

- **Create:** <how>
- **Read:** <list view, detail view>
- **Update:** <what's editable, by whom, in which states>
- **Delete:** <soft / hard / not allowed; what happens to things referencing it>
- **Undo:** <can this be reversed; how>

## Acceptance criteria

- [ ] <specific, verifiable>
- [ ] <specific, verifiable>

## Out of scope

- <what this ticket deliberately does not do>
```

## The lifecycle check

**Any ticket that creates or modifies a record type must state what happens for create, read, update, delete, and undo.**

"You cannot delete these" is a perfectly good answer. But it has to be **decided**, not silent.

**Silence here is what produces missing delete flows and no way to recover from a mistake.** The agent builds exactly what was described, and descriptions default to the optimistic case: create but not delete, success but not error, one row but not zero.

The Planning decision on soft vs hard delete (Group B, data lifecycle) sets the default. The ticket says how it applies here.

## Rules

**Durability over precision.** A ticket may sit for weeks while the codebase moves.

- **Do** describe behaviour, interfaces, and contracts
- **Do** name types and concepts the agent should look for
- **Don't** reference file paths — they go stale
- **Don't** reference line numbers
- **Don't** assume the current implementation structure survives

Exception: a snippet that encodes a decision more precisely than prose can — a state machine, a schema shape, a type — may be inlined. Trim it to the decision, not a working demo.

**Acceptance criteria must be verifiable.**

- Good: "Filtering to Overdue returns only invoices past their due date"
- Bad: "Filtering works correctly"

**Out of scope must be stated.** It's what stops the agent gold-plating and stops adjacent features being assumed in.

**Use the project's vocabulary.** Ticket titles and descriptions use terms from `CONTEXT.md`. If a needed word isn't in the glossary, that's a vocabulary gap — surface it rather than inventing a term.

## Wide refactors — the exception

A **wide refactor** is one mechanical change whose blast radius spans the codebase — renaming a column, retyping a shared symbol. A single edit breaks thousands of call sites, so no vertical slice can land green.

Don't force it into a tracer bullet. Sequence it **expand–contract**:

1. **Expand** — add the new form beside the old. Nothing breaks
2. **Migrate** — move call sites in batches sized by blast radius (per module, per directory), each batch its own ticket blocked by the expand. CI stays green because the old form still exists
3. **Contract** — delete the old form once no caller remains, in a ticket blocked by every migrate batch

## Ordering

Files are numbered in dependency order — blockers first. `01`, `02`, `03`.

Work the frontier: any ticket whose blockers are all done. For a linear chain, that's top to bottom.

**Prefactoring goes first.** Make the change easy, then make the easy change.
