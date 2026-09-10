---
name: design
description: Design every not-yet-built screen for one stage/feature as a real, built, interactive prototype, through a per-screen build/self-critique/review loop, landing each approved screen as a Storybook story. Run once at bootstrap for the whole v1 product (before /tickets); re-run scoped to one stage/feature just-in-time, before that stage's /tickets, for the rest of the project's life.
---

# Design

Takes a `scope` argument: a stage name or feature area (e.g. "Stage 4 —
canteen operations", or a `docs/scope.md` feature name). Always state the
scope explicitly before starting — never infer it silently from
conversation context.

## Purpose

Make every UI/UX decision for the screens in scope up front, as real
running code, so that `/build` never has to invent layout, component
composition, or visual design while implementing a ticket. By the end of
a run, every screen the scope needs exists as an approved, working
Storybook story built from real components and realistic mock data —
`/build` only assembles what's already designed.

Screens are built as actual interactive prototypes (real components, real
navigation, real mock data), never static mockups. Enterprise design
quality is judged on usability and speed as much as appearance, and neither
of those can be evaluated from a static image.

## Bootstrap run vs. scoped rerun

**First-ever run** (right after `/foundation`, before any `/tickets`):
scope is the whole v1 product, theme doesn't exist yet, `docs/screens.md`
doesn't exist yet. Do steps 1 and 2 in full.

**Every later run** (steady state, for the rest of the project's life):
scope is one stage or feature area whose screens are about to be ticketed.
Theme and `docs/screens.md` already exist — steps 1 and 2 become quick
checks, not fresh work. This is the common case from here on; assume it
unless nothing has been designed yet.

## Inputs

- `docs/architecture.md`, `docs/scope.md`, `docs/conventions.md` from `/alignment`
- The scaffolded repo, component tooling (shadcn init'd), and Storybook setup from `/foundation`
- `references/ui-rules.md` and `references/design-principles.md` (this
  skill's reference material — read both in full before doing anything else)
- `docs/design.md` and the existing `docs/screens.md`, if they already exist

## Process

### 1. Confirm the theme

**Bootstrap run:** decide it fresh, using `references/design-principles.md`
as the basis — neutral ramp, accent colour, semantic colours, radius
personality and scale, type scale/font, motion durations/easings. Write
into the project's Tailwind v4 `@theme` block. Create `docs/design.md` as
a short index: chosen font, accent colour, radius personality, density
default, and a pointer to the theme file and to `references/ui-rules.md` /
`references/design-principles.md`. Do not duplicate token values into
`docs/design.md` — point to the theme file instead. Confirm the finalized
theme with Edwinfred before proceeding — this is a cheap, high-leverage
checkpoint since every screen from here on depends on it.

**Scoped rerun:** the theme file and `docs/design.md` already exist and
are in active use by every previously-approved screen — do not redo this
step. Just confirm nothing about it needs to change for the screens in
scope; if it does, that's a real design decision, flag it to Edwinfred
explicitly rather than editing tokens as a side effect of a single stage.

### 2. Extend the screen inventory

**Bootstrap run:** using `docs/scope.md` and `docs/architecture.md`,
produce `docs/screens.md` from scratch: every screen the v1 product
needs, organized as:

```
## <Role/Persona> (e.g. Admin, Director, Accountant)
### <Destination> (a page may contain multiple distinct screens/states)
- **Screen:** <name>
  **Purpose:** <what it's for, what it depicts, concisely>
  **Status:** planned
```

**Scoped rerun:** `docs/screens.md` already exists with prior stages'
screens `approved`. Append rows for the current scope's screens only,
under the existing role/destination structure — don't touch or
re-organize already-approved rows. If the scope's screens aren't already
named in `docs/screens.md`'s "Not yet built" note, derive them from
`docs/scope.md` / `docs/roadmap.md` for this stage the same way the
bootstrap run would, then add them as new `planned` rows.

This is a living checklist — status moves through `planned` → `in review`
→ `approved` as the per-screen loop below proceeds. It makes the phase
resumable: an agent picking this up in a later session reads exactly what's
left, without Edwinfred re-explaining anything.

Once the new rows are in place, ask Edwinfred what cadence to use for
this scope — one screen at a time, or all of this scope's screens before
review. Don't assume; one-at-a-time is the recommended default.

### 3. Identify pattern types before building

Before building anything new, check the pattern types already established
by prior approved screens (RecordTable, DetailPage, Form, SummaryStrip,
etc., per `references/design-principles.md`'s page templates) — the
default is to reuse one of these for each screen in scope. Variant
exploration (step 4 below) only happens for a screen whose shape doesn't
match any pattern type already in use anywhere in `docs/screens.md`, not
just within the current scope. Every screen that fits an established
pattern reuses it directly — no new variants, just consistent
application. This keeps effort proportional and is itself what produces
consistency across the app.

### 4. Per-screen build loop

For each screen in scope, in the order and cadence Edwinfred chose:

1. **Build it as a real prototype** — actual shadcn components, actual
   Storybook-viewable, actual mock/seed-shaped data (per `scope.md`'s
   definition-of-done for the relevant feature — see "Mock data" below).
   If this is genuinely the first screen of a pattern type not already
   established anywhere in the app, build 2–3 real variants; otherwise
   build one, following the established pattern. Include the screen's
   required states in the same pass, not as a later add-on: empty
   (first-use and no-filter-results are different messages), loading
   (skeleton, not spinner), error, and permission-denied, per
   `references/ui-rules.md`'s "Required states" section.

2. **Self-critique before showing Edwinfred.** Check the built screen
   against `references/ui-rules.md` directly, item by item. Fix anything
   mechanically wrong (arbitrary values, contrast failures, missing
   states, alignment issues) before presenting — Edwinfred's review time
   should go to taste and judgment calls, not rule violations the agent
   should have caught itself.

3. **Present to Edwinfred.** For a first-of-pattern screen, present all
   variants together for comparison.

4. **On approval:** optionally note what Edwinfred liked about it. Record
   this in `docs/screens.md` against that screen's row, for Edwinfred's own
   future reference only — do **not** feed it back into this skill's own
   behavior as a rule to replicate elsewhere. Edwinfred is deliberately
   still exploring and does not want early approvals to narrow future
   variety. Mark the screen `approved` and land it as a Storybook story —
   this is now the canonical, in-repo version `/build` will read.

5. **On rejection:** ask what was wrong. Take Edwinfred's input, then
   produce new variants reflecting it — not a single blind retry. Return
   to step 2 for the new variants.

6. Move to the next screen per the chosen cadence, repeating until every
   screen in this scope is `approved`.

### Mock data

Every screen is designed fully against mock data shaped like the real
thing — never half-designed because a backing feature doesn't exist yet.
Use the seed data conventions from `/foundation` as the base; extend with
realistic edge-case data per `references/ui-rules.md`'s "Edge cases"
section (long text, many items, zero items, extreme numbers) so screens
are checked against reality, not just the happy path.

If a screen needs a data shape that isn't actually defined in
`docs/scope.md`'s definition-of-done, that is a sign `/alignment` left a
gap — stop and raise it with Edwinfred as a scope gap to resolve, rather
than guessing at a shape or leaving the screen unfinished.

## Output

**Bootstrap run:** finalized theme file, `docs/design.md`, fully-approved
`docs/screens.md`, every approved screen as a Storybook story in-repo.

**Scoped rerun:** theme and `docs/design.md` unchanged (confirmed, not
regenerated); `docs/screens.md` with this scope's new rows moved to
`approved`; new Storybook stories for just those screens.

Either way: `docs/conventions.md` may need a pointer added to the
Storybook location, if not already referenced from `/foundation`.

Note: these Storybook stories are isolated design references, not the
running app. On the bootstrap run, `/tickets` cuts a dedicated App Shell
ticket (built first, before any feature) that wires real navigation and
routing in the actual app around these designs — every destination
becomes a real, clickable route showing the relevant approved screen's
designed state (including its empty/pending state for destinations not
yet implemented) rather than a 404. On a scoped rerun, the shell already
exists — the stage's `/tickets` run wires its new routes into it instead.

## Explicit non-goals

- No Figma or external mockup tooling in the default path — building real
  variants directly is cheap enough with shadcn's component coverage that
  a separate mockup step adds a translation cost without a clear payoff.
  Revisit only if fast, low-fidelity exploration before committing to real
  code becomes genuinely necessary on a specific project.
- No feature implementation beyond what's needed to render a screen
  against mock data — real logic and real data wiring is `/build`'s job.
- Don't let an approved screen's "liked" notes become an implicit rule the
  agent applies to unrelated future screens.
- No redesigning the theme or re-touching already-approved screens as a
  side effect of a scoped rerun — scope creep here defeats the point of
  scoping the run at all.
