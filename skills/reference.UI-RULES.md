# UI Rules

The checkable design rules. Used by `/critique`, `/review`, and `/build`.

These are **rules already decided**, not aesthetic opinions. Every one is objectively checkable. The premise: what makes UI bad is mostly arbitrariness, and arbitrariness is detectable.

Project-specific rules in `docs/design.md` and `CLAUDE.md` **override** anything here.

## Tokens and values

- [ ] No arbitrary values — `p-[13px]`, `text-[#7a7a7a]`, raw hex in components
- [ ] All spacing from the scale: 4 · 8 · 12 · 16 · 24 · 32 · 48 · 64 · 96
- [ ] All greys from the single neutral ramp — never a mix of `gray-*` and `slate-*`
- [ ] All radii from the radius scale, and scaled to element size
- [ ] Colour from theme tokens only

## Hierarchy

- [ ] **One accent element per screen** — the primary action. If everything is accent-coloured, nothing is important
- [ ] Not two filled primary buttons side by side — that means the primary action wasn't decided
- [ ] Three levels of hierarchy, rarely more
- [ ] Hierarchy comes from **weight and colour** before size
- [ ] Left-aligned by default. Centre only small self-contained things — an empty state, a modal action row

## Typography

- [ ] Body text 13–14px (enterprise density, not 16–18px marketing sizing)
- [ ] Line height 1.4–1.5 for UI, 1.2–1.3 for headings
- [ ] Weights 400 / 500 / 600. Not 300, rarely 700
- [ ] `tabular-nums` on any column of numbers
- [ ] Text capped at ~60–75 characters where it's read as prose

## Density

- [ ] Table rows 36–48px. Over ~56px, ask why
- [ ] Table cell vertical padding 8–12px
- [ ] A list page shows ~20+ items above the fold on a laptop

## Colour and contrast

- [ ] 4.5:1 contrast for normal text, 3:1 for large text and UI borders
- [ ] `gray-400` on white fails (~2.8:1). `gray-500` passes (~4.6:1)
- [ ] Semantic colours mean what they say — red for destructive or error, green for success. Never decorative

## Shadows and borders

- [ ] Shadows **only** on things that genuinely overlay content — menus, popovers, modals, toasts
- [ ] Flat cards use a border or background shift, never a shadow

## Icons

- [ ] One icon set throughout
- [ ] 16px inline with text, 20px standalone
- [ ] Stroke weight matches text weight
- [ ] Icon-only buttons have **both** a tooltip and an `aria-label`

## Motion

- [ ] 100ms hover and colour changes
- [ ] 150ms dropdowns, popovers, tooltips
- [ ] 200ms modals and drawers
- [ ] **Never over 300ms**
- [ ] Exit animations faster than entrances (~2/3)
- [ ] `ease-out` by default
- [ ] **Only `transform` and `opacity`** — never `width`, `height`, `top`, `left`, `margin`, `padding`
- [ ] Movement distances small — 2–8px, not 40px
- [ ] `prefers-reduced-motion` respected

## Layout

- [ ] Every page opens inside a shell from `components/layout/` and composes patterns from `components/patterns/` — no invented layouts
- [ ] Page-level actions top-right of the page header, consistently
- [ ] Max content width set — full-width on a 27" monitor is unusable
- [ ] Works at 1280px

## Tables

- [ ] Sticky header when over ~15 rows
- [ ] Sortable by clicking the header
- [ ] Numbers right-aligned, text left-aligned
- [ ] Row hover highlight
- [ ] Long text truncated with ellipsis **and** a tooltip or `title`
- [ ] Counts shown — "1–50 of 1,204", not "Page 1"
- [ ] Bulk actions appear on selection

## Forms

- [ ] Labels **above** inputs, left-aligned. Placeholder is not a label
- [ ] Validation on **blur**, not on every keystroke
- [ ] Errors shown next to the field, not only summarised at the top
- [ ] **Input preserved on error** — a form that clears is unforgivable
- [ ] Single column
- [ ] Submit disabled while submitting, with text saying what's happening
- [ ] Sections with headings once over ~8 fields

## Keyboard

- [ ] Tab order follows visual order
- [ ] Escape closes; Enter submits
- [ ] **Focus rings visible** — never `outline: none` without a replacement
- [ ] Links are real links, so Ctrl/Cmd-click opens a new tab

## Required states

Every list, table, and detail view:

- [ ] **Empty — first use.** Says what goes here and how to create the first one
- [ ] **Empty — no results from filter.** A *different* message, with "Clear filters". **Never** show "create your first X" here — they have 400, they filtered wrong
- [ ] **Loading.** Skeleton matching real dimensions, not a full-page spinner
- [ ] **Error.** Plain language, a retry, and input preserved
- [ ] **Permission-denied.** Hidden, or disabled with a stated reason. **Never an enabled control that errors on click**

## Mistake-proofing

- [ ] Destructive actions confirmed — **and the confirm names the specific thing.** "Delete invoice INV-2024-0142?" not "Are you sure?"
- [ ] Undo offered where it beats a confirm dialog — a toast with "Deleted. Undo"
- [ ] Irreversible actions marked **before**, not after
- [ ] No dead ends — every screen has a way back
- [ ] **Lifecycle actions the ticket declared are actually present.** A detail page for a record the ticket said is deletable must have a delete action

## Edge cases

- [ ] Zero items
- [ ] One item — layouts sometimes break
- [ ] Many items (1000+) — still fast
- [ ] Very long text — a 200-character name
- [ ] Null and missing values — "—", not "undefined"
- [ ] Extreme numbers — negative, zero, nine figures

## The squint test

Blur your eyes, or zoom to 25%. What still stands out is the hierarchy.

If everything blurs into an even grey mush, there is no hierarchy. If the wrong thing stands out, emphasis is misplaced.
