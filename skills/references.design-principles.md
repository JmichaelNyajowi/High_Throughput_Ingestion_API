# Design Principles Reference

Mechanical, checkable rules for building enterprise application UI. Used by
`/design` when finalizing the theme file and building screens, and by
`/review` when checking UI-touching tickets. Companion to `references/ui-rules.md`
(the checklist form of much of this).

The core premise: the gap between generic-looking UI and best-in-class UI is
mostly not taste — it's a set of specific, learnable, mostly mechanical
constraints. Arbitrary values (13px here, 15px there, six shades of grey)
read as generic. Constraint reads as intentional. Most of this is enforceable
directly in the theme file, which is why the theme file — not this document —
is the actual source of truth for values. This document explains the
reasoning behind the constraints; the theme file is what's binding.

## Enterprise UI is a different discipline than consumer/marketing UI

An enterprise user opens the app at 9am and closes it at 5pm, five days a
week, for years. They are not being delighted — they are trying to finish
work and leave. Design accordingly:

- **Density beats whitespace.** Maximum scannable information, not minimum
  visual noise. Tighter line-heights (1.4–1.5, not 1.7), smaller base font
  (13–14px), tighter row padding (8–12px vertical). If a table row is taller
  than ~40px, ask why. If a list page shows fewer than ~20 items above the
  fold on a laptop, it's too airy.
- **Keyboard is a primary interface, not an accessibility afterthought.**
  Tab order follows visual order. Enter submits, Escape closes, always.
  Visible focus rings, never `outline: none` without a replacement.
  Shortcuts for top actions. A command palette (`⌘K`) once the app passes
  ~15 destinations. Arrow-key navigation in lists/tables.
- **Speed is felt viscerally.** Optimistic updates (update UI immediately,
  reconcile after, roll back only on failure). Never block UI on save —
  background autosave with a subtle indicator beats a spinner. Skeletons
  (matching real dimensions), not spinners, for content loading. Preserve
  scroll position and filters across navigation. Virtualize long lists.
- **Boring is correct.** Match existing muscle memory (Excel, Outlook,
  Salesforce): click header to sort, checkboxes + shift-click to
  multi-select, Ctrl/Cmd-click opens new tab (real links, not div click
  handlers), primary action bottom-right in dialogs, destructive actions
  red with confirmation. Innovate in what the software does, never in how
  a dropdown behaves.
- **Data tables are the enterprise UI.** Sorting, per-column + global
  filtering, column show/hide/reorder/resize (persisted per user), bulk
  selection with an action bar, sticky header, density toggle, inline
  editing for common edits, proper empty/loading states, CSV export.
  Build this once as a reusable component — highest-leverage component in
  the library.
- **Forms:** validate on blur not on keystroke, errors next to the field
  (not only a top summary), never clear input on error, autosave long
  forms, label above input (placeholder is not a label), group fields with
  headings past ~8 fields, mark optional fields when most are required,
  disable submit while submitting with a status message.
- **Animation must be near-invisible.** Budget: 100–150ms most
  transitions, 200ms max, never over 300ms. Only animate `transform` and
  `opacity` (GPU-accelerated, no layout thrash). `ease-out` default for
  things entering, `ease-in` for things leaving. Small movement distances
  (2–8px, not 40px). Respect `prefers-reduced-motion`.
- **Multi-user reality.** Show who else is viewing/editing where feasible.
  Handle conflicting edits explicitly, never silently overwrite. Keep
  audit trails visible. Permission-aware UI: hide or disable-with-reason,
  never show a control that errors on click.
- **Beauty comes from restraint, not decoration.** One accent colour, one
  tight neutral ramp, consistent spacing, small text, sharp-to-medium
  corners, near-zero ornament. Nothing arbitrary.

## The fundamentals

**Spacing** — one scale for the entire app: `4, 8, 12, 16, 24, 32, 48, 64, 96`.
No arbitrary values. Proximity communicates grouping — related things close,
unrelated things far, before reaching for a border or box.

**Typography** — one typeface (a second only for monospace/code/IDs).
Compressed type scale for dense UI: body/table 13–14px, page title 20–24px
— hierarchy comes from weight and colour more than size. Line height 1.4–1.5
for UI, 1.2–1.3 for headings. Weights 400/500/600 (avoid 300, rarely 700).
`tabular-nums` on any numeric column. Cap prose line length at ~60–75
characters.

**Colour** — one neutral ramp (10–12 steps, pick one Tailwind ramp and use
only that one), one accent colour used sparingly (1–3 elements per screen),
three semantic colours (success/warning/danger, meaning only, never
decorative). Contrast: 4.5:1 normal text, 3:1 large text/UI borders —
non-negotiable and checkable. Dark mode is not inverted light mode: very
dark grey background (not pure black), off-white text (not pure white),
elevated surfaces get lighter not darker, reduce accent saturation.

**Hierarchy** — every screen answers "what is the one thing here?" Tools in
order of strength: size, weight, colour contrast, fill/outline/ghost,
position, space, hue. Three levels, rarely more. Button hierarchy: primary
(filled accent, one per screen/dialog), secondary (outline/subtle),
ghost/tertiary (text only), destructive (red).

**Alignment** — establish one left edge and hold it down the page across
titles, headings, labels, content. Right-align numbers in tables (with
tabular-nums), left-align text. Set a max content width (~1280–1440px) for
content; data tables may go full width.

**Density** — pick comfortable as default (40px rows, 13–14px body, 12px
padding); offer a compact toggle on primary data views. Increase density by
reducing padding and tightening line-height before shrinking font size.

## Page templates

A small, fixed set of page shapes. When a ticket says "this is a detail
page," the arrangement should already be decided — the agent picks a
template and fills it rather than inventing an arrangement. Layout
consistency reads as intentional more than colour choice does.

- **List page** — header (title + filter/export/new actions), toolbar
  (search, filter chips, density toggle, column settings), table (sticky
  header, sortable, selectable), footer (pagination/count, bulk action bar
  on selection).
- **Detail page** — breadcrumb, header (name, status, actions), key-facts
  strip (3–5 fields), tabs/sections for the rest.
- **Form page** — header with Cancel/Save, single column, max ~600px wide,
  sectioned with headings, sticky action bar for long forms. Always single
  column — multi-column forms have worse completion and break tab order.
- **Dashboard page** — header with date range, stat row (3–5 metrics),
  chart grid, recent activity table.
- **Settings page** — sub-navigation, sectioned independently-saveable
  form panels.

Plus: empty/onboarding state and error/permission-denied state need
designing as first-class states, not afterthoughts.

**Responsive stance:** design for 1440px primary, ensure 1280px works,
tablet collapses sidebar and stacks columns, mobile just needs to not be
broken (not a redesign) unless the client specifically needs field/mobile
use. Tables may scroll horizontally on small screens rather than
transforming into cards.

## Tooling defaults

- **Components: shadcn/ui.** Copy-in (source lands in the repo, fully
  editable — critical for agent-driven work), built on Radix (correct
  accessibility/keyboard/focus behaviour) + Tailwind. Never build primitive
  components (dropdowns, dialogs, comboboxes) from scratch — the
  accessibility surface is large and easy to get subtly wrong.
- **Tables: TanStack Table** for logic (sorting/filtering/grouping/
  pagination/selection), **TanStack Virtual** for long lists, **TanStack
  Query** for server state + optimistic updates.
- **Icons: Lucide** — one set, one weight. 16px inline with text, 20px
  standalone. Icon-only buttons need both a tooltip and an `aria-label`.
- **Forms: React Hook Form + Zod**, via shadcn's Form wrapper.
- **Charts: shadcn chart components (Recharts)**; reach for ECharts only at
  serious data volume.
- **Component gallery / catalogue: Storybook.** Every approved component
  and screen composition exists as a story in-repo. This is what an agent
  reads to see what already exists before building something new, and it's
  the durable, in-repo source of truth for approved designs — never a
  separate design tool.
- **Fonts:** Inter as the default choice unless a specific alternative is
  chosen; self-host via Fontsource, not a Google Fonts CDN request.
- **Accessibility tooling:** `eslint-plugin-jsx-a11y` wired into lint,
  axe DevTools / Lighthouse for spot checks.

## Design tokens

A token is a named design decision (`--color-primary`, not `#2563EB`
inline). This is the actual anti-drift mechanism: if only token values are
available, arbitrary values become impossible by construction, not by
agent discipline. The theme file — a Tailwind v4 `@theme` block — is the
binding source of truth; this document and `docs/design.md` only explain
intent.

Token categories to define: colour (neutral ramp, accent, semantic),
spacing (the 4px scale), typography (family, size scale, weights, line
heights), radius (small/medium/large/full, scaled to element size), shadow
(used only for genuine overlay elevation — menus/popovers/modals/toasts,
never flat cards), border, z-index (named layers, not magic numbers),
motion (named durations/easings).

**The enforcement rule, verbatim, belongs in `CLAUDE.md`:** all spacing,
colour, radius, and typography values must come from theme tokens. Never
arbitrary Tailwind values or raw hex in components. If a needed value
doesn't exist in the theme, stop and ask rather than inventing one.
