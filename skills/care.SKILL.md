---
name: care
description: Scan the codebase for architectural friction, rank deepening candidates, stop for the user to pick one, then grill it into tickets. Includes a hygiene pass. Run between tranches or when Fix, gotchas, or critique signal something structural.
disable-model-invocation: true
---

# Care

Find and fix architectural friction **before it compounds**.

Foundation establishes architecture. **Care defends it.** Agents accelerate code production, which means they accelerate decay — without a deliberate loop, every project trends toward mud regardless of how well it started.

Read `<skills>/reference/MODULES.md` for the vocabulary. Use those words exactly.

**Resolving `<skills>/`.** It is the directory holding the skill folders — `~/.claude/skills/` for a global install, or `<project>/.claude/skills/` for a per-project one. **It is not inside this skill's own folder, and it is not in the project root.** Check the global path first, then the project-local one.

**If a reference file cannot be found, stop and tell the user.** Do not proceed from memory — these files hold the discipline the skill depends on, and running without them silently produces work that looks right and isn't.

## When this runs

**Between tranches**, and on three triggers — each evidence that something is structurally rather than incidentally wrong:

- A `/fix` post-mortem that found **no good seam** for a regression test
- **Three or more gotchas clustering** in one area
- The **same `/critique` finding recurring**

## 1. Scope before scanning — YAGNI

**Deepening pays off by making future changes easier**, so weight toward what's actually changing.

- If the user named a direction — a module, a pain point — take it and skip the inference
- Otherwise walk `git log --oneline` over a good stretch and find the **hot spots**: the files and areas that keep coming up. Let those pull your attention
- If changes are scattered with no clear hot spot, widen the net

Read `CONTEXT.md` and any ADRs in the area first.

## 2. Scan for friction

Use exploration sub-agents. **Look for friction, not checklist violations** — rigid heuristics find rule breaches; what matters is where the codebase fights you, which is a judgment made by reading it.

Five signals:

| Signal | Quality it threatens |
|---|---|
| Understanding one concept requires bouncing between many files | Easy to understand |
| A module with a large interface but little behind it | Easy to understand |
| **Boundary violations — modules reaching into each other's internals** | Easy to extend |
| Hard to test through its current interface | Easy to test |
| Keeps producing bugs | Easy to find bugs in |

Apply **the deletion test** to anything suspected shallow: if this were deleted, would complexity **vanish** (it was a pass-through) or **reappear across many callers** (it was earning its keep)? "Reappears" is the signal you want.

## 3. Rank candidates

For each:

- **Files/modules involved**
- **Problem** — why this causes friction now
- **Proposed change** — in plain English
- **Benefit** — in terms of testability and how change would concentrate rather than spread
- **Strength** — `Strong` / `Worth exploring` / `Speculative`

Use `CONTEXT.md` vocabulary for the domain and `<skills>/reference/MODULES.md` vocabulary for the architecture. Talk about "the Billing module's interface", not "the BillingHandler service".

## 4. Write the report outside the repo

Write to the OS temp directory — `$TMPDIR` or `/tmp` — as `architecture-review-<timestamp>.md`. Tell the user the path.

**Never write it into the repo.** It's a snapshot, not documentation. In `docs/` it becomes stale clutter.

## 5. Stop and ask

Present the ranked list and ask which one to explore.

**Do not propose interfaces yet.** Proposing solutions for all candidates wastes effort on the ones the user will never take, and the stop is what keeps scope in their hands.

## 6. Grill the chosen candidate

Follow `<skills>/reference/INTERROGATION.md`. Walk through:

- The constraints any new shape must satisfy
- What the module depends on, and how each dependency is tested across the seam
- The shape of the improved module — what sits behind the interface
- Which existing tests survive, and which become waste

**Replace, don't layer** — when shallow modules merge into a deeper one, the old tests on the shallow pieces become waste. Delete them; write new tests at the new interface.

If a term settles or sharpens during the grilling, update `CONTEXT.md` right there.

If the user **rejects** a candidate for a load-bearing reason, offer an ADR: *"Want me to record this so future architecture reviews don't re-suggest it?"* Only when the reason would actually be needed by a future reader — skip ephemeral ones ("not worth it right now").

## 7. Check tests exist first

**A refactor changes structure without changing behaviour. The only proof is a test that passed before and passes after.**

If the area has no tests, **writing them is the first ticket** — before any restructuring. Refactoring untested code is rewriting it and hoping.

## 8. Hand to `/tickets`

A refactor is a build. Same interface plan, same tests, same review.

Tell the user to run `/tickets`, and stop.

---

## Hygiene pass

Every run. Five-minute jobs that never happen otherwise and quietly degrade the agent's context quality.

- **Prune stale gotchas** — an entry that's no longer true sends agents chasing a problem that doesn't exist, which is worse than not having it. Delete anything fixed at the root or made obsolete by an upgrade
- **Delete dead code and unused dependencies**
- **Check `CLAUDE.md` still matches reality** — a convention nobody follows is worse than none
- **Check `CONTEXT.md` still matches the vocabulary in use** — if the code says one thing and the glossary another, one of them is wrong
- **Remove skipped tests** — a permanently skipped test is a lie about coverage

Report what you pruned.

---

## Constraints

- **One candidate per pass**, and it must fit in a normal tranche. Large refactors spanning weeks are where projects die. If it doesn't fit, it's a sequence of smaller ones
- **If a candidate contradicts an existing ADR**, surface it explicitly and only pursue it if the friction genuinely justifies reopening the decision. Mark it clearly. A reversal means a **new** ADR superseding the old, never an edit
- **Update `docs/architecture.md` and `CLAUDE.md`** when structure changes. Docs describing the old structure actively mislead the next agent — worse than no docs

## Never

- **Never refactor without tests in place first**
- **Never take on more than one candidate**
- **Never write the report into the repo**
- **Never propose interfaces before the user has picked a candidate**
