---
name: critique
description: Audit UI against the project's design rules and enterprise patterns — arbitrary values, hierarchy, density, missing states, mistake-proofing, and lifecycle actions the ticket declared. Reports concrete findings, never aesthetic opinions. Use after a ticket that built new screens.
---

# Critique

Audit the UI against **rules that are already decided**.

Read `<skills>/reference/UI-RULES.md` for the checklist. Read `docs/design.md`, `CLAUDE.md`, and the theme file — **project rules override the reference.**

**Resolving `<skills>/`.** It is the directory holding the skill folders — `~/.claude/skills/` for a global install, or `<project>/.claude/skills/` for a per-project one. **It is not inside this skill's own folder, and it is not in the project root.** Check the global path first, then the project-local one.

**If a reference file cannot be found, stop and tell the user.** Do not proceed from memory — these files hold the discipline the skill depends on, and running without them silently produces work that looks right and isn't.

Run this **after `/verify`**. No point critiquing a flow that doesn't work.

## The premise

An agent has no taste and can't develop one. Asked "is this good UI?", it produces plausible-sounding noise.

But **most of what makes UI bad is objectively checkable.** The gap isn't taste, it's arbitrariness — and arbitrariness is detectable.

**This skill audits compliance with rules the user already set.** The user keeps the taste; this does the auditing. Every finding must trace to a written rule.

## What to check

Read the diff, the ticket, and the rendered screens if `/verify` produced screenshots.

### Mechanical — read from the code

- Arbitrary values: `p-[13px]`, `text-[#7a7a7a]`, raw hex in components
- Greys not from the single neutral ramp — a mix of `gray-*` and `slate-*`
- Spacing off the scale; radii off the scale
- Transitions over 200ms, or animating layout properties instead of `transform`/`opacity`
- Missing empty, loading, or error states
- Missing `focus-visible` styles; icon buttons without `aria-label`
- Contrast failures — `gray-400` on white fails
- Icons from the wrong set or the wrong size
- A page not opening inside a shell from `components/layout/`, or not composed from `components/patterns/` where a matching pattern exists
- A component duplicating something already in `components/ui/`
- A new component added to `components/ui/` — that should have been a stop-and-ask

### Enterprise patterns

- Table rows over ~48px; cell padding over 12px
- No sticky header on a long table; not sortable
- Numbers left-aligned, or missing `tabular-nums`
- Body text at marketing sizes (16–18px) rather than 13–14px
- Labels inside inputs as placeholders rather than above
- Validation on keystroke rather than blur
- Multi-column forms
- Tab order not following visual order; Escape doesn't close; Enter doesn't submit
- Long text not truncated with a tooltip

### Mistake-proofing

The safety net for what `/tickets` is meant to have closed.

- **Destructive actions unconfirmed** — or confirmed with "Are you sure?" instead of naming the thing: "Delete invoice INV-2024-0142?"
- **No undo** where a toast with "Deleted. Undo" would beat a confirm dialog
- **Errors clearing typed input**
- **Dead ends** — a screen with no way back
- **Irreversible actions not marked** before the click
- **Lifecycle actions the ticket declared but the screen doesn't have** — if the ticket's Lifecycle section says invoices are deletable and the detail page has no delete action, that's a finding. **This is the check that catches missing delete and undo flows**

### Flow — the only judgmental part, kept modest

- Click count for the primary task
- Is the primary action visually primary?
- After an action, can the user tell what happened?

## Output

A **ranked list**. Each finding names the file, the rule broken, and the fix.

```
1. Table rows are 56px — design.md specifies 36–40px for enterprise density
   → modules/billing/ui/invoice-table.tsx: reduce py-4 to py-2

2. Two filled accent buttons in the page header — no clear primary action
   → make "Export" variant="outline"

3. Ticket 14 specifies invoices are deletable; the detail page has no delete
   → modules/billing/ui/invoice-detail.tsx

4. Invoice list has no empty state — the page collapses at zero rows
   → modules/billing/ui/invoice-table.tsx

5. Delete confirm reads "Are you sure?" — should name the record
   → "Delete invoice INV-2024-0142?"
```

Order by severity: missing states and mistake-proofing gaps first, then pattern violations, then cosmetic drift.

**Never write "consider improving the visual hierarchy."** If you can't name the rule and the fix, it isn't a finding.

## Findings should become constraints

At the end, check for repetition.

**If the same finding has appeared three times across recent critiques, say so** — that's a missing `CLAUDE.md` line, not a recurring fix. Recommend the prevention rule.

Over time this skill should find less. It's a detector whose output should mostly convert into prevention.

## Never

- **Never give aesthetic opinions.** Every finding traces to a written rule
- **Never suggest a redesign** — that's `/design`
- **Never fix directly.** Report; the user decides
