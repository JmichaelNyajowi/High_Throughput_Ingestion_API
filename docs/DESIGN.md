# Telemetry UI Design Brief

## Scope and authority

This is the binding design brief for the telemetry product described in `docs/prd.md`. It adopts the useful reference guidance on density, component consistency, accessibility, and restraint, but it does **not** expand MVP scope. In particular, this UI does not invent alert delivery, key management, tenant administration, RBAC, or long-range analytics.

Follow this document for every subsequent UI task. If a required value, component state, or pattern is absent, add a named design token or component variant first; do not improvise a raw value or one-off layout.

## 1. Design principles

### 1. Operational signal outranks decoration

The UI exists to answer: **Which devices need attention now, and why?** Freshness, measurement state, threshold status, and degradation state have more visual weight than branding, chart polish, or card decoration.

- Show a device's freshness beside its status; a critical reading that is two minutes old is a different operational fact from a current critical reading.
- Use the accent color for the primary action or active navigation only. Reserve semantic colors for true status.
- Do not use gradients, glow, glass effects, ornamental illustrations, emoji UI icons, or decorative motion. These consume attention that should belong to telemetry state.

**Why:** Industrial telemetry is read under time pressure. A calm instrument panel is safer and more credible than a consumer-style “real-time” dashboard.

### 2. Dense enough to triage, calm enough to trust

Default to a compact operational density: 13–14 px UI text, 40 px data-table rows, tight but readable 1.4–1.5 line height, and a stable 4 px spacing rhythm. A fleet operator must scan many devices without a screen filled with oversized cards.

- Prefer a sortable, filterable table for the fleet over a gallery of device cards.
- Use borders and background shifts to group content; do not put shadows on ordinary panels.
- Keep every screen aligned to one strong left edge and a 1440 px maximum content frame. Let the fleet table use the frame's full width.

**Why:** The product’s value is rapid comparison across many devices. Density is useful only when typography, alignment, and status semantics remain deliberate.

### 3. State is data, not decoration

Loading, stale, degraded, empty, and failed conditions are first-class states. The UI must never imply that data is current when Redis, polling, or the API is unavailable.

- Use skeletons shaped like the eventual table/chart; never leave a blank panel or show a generic full-page spinner.
- Persist the last successful live snapshot in the browser query cache, label it with its refresh time, and visibly mark it stale during degradation.
- Use explicit words in addition to color: `Live`, `Stale`, `Warning`, `Critical`, and `Degraded`.

**Why:** The architecture intentionally permits Redis loss and eventual consistency. Honest state communication prevents operators from acting on false confidence.

## 2. Visual direction

### Mood

**Quiet industrial instrumentation.** The interface should feel like a well-designed control console: graphite surfaces, precise cyan guidance, restrained status color, compact numerals, and strong alignment. It is focused rather than dramatic.

Default to a dark graphite operational theme. This suits sustained monitoring in a control-room or engineering context, reduces glare around bright chart lines, and makes warning/critical states legible without covering the page in color. It is not “cyberpunk”: backgrounds are neutral charcoal, not black; the accent is controlled teal, not neon; text is off-white, not pure white.

### Product-specific references

- Industrial control panels: clear readings, stable grids, semantic alarm states.
- Aviation and network-operations displays: data-first hierarchy, persistent system condition, no ornamental marketing layers.
- High-quality enterprise data tables: sortable columns, compact rows, clear status chips, and numbers aligned for comparison.

### Avoid

- Purple gradients, neon glow, glassmorphism, particle/sparkle decoration, or simulated “AI” visual effects.
- Marketing-style hero layouts, huge headings, oversized empty space, testimonials, or fake device imagery.
- Multiple competing filled buttons, rainbow charts, or status indicated by color alone.
- Floating cards with heavy shadows. Flat surfaces use a border or subtle background separation; shadows are only for overlays.
- Animations that change layout, bounce, rotate, or delay access to data.

## 3. Design tokens

Tokens are named decisions. Components may consume token names only; do not place raw hex values, arbitrary spacing, or one-off radii in UI code.

### 3.1 Color palette

| Token | Hex | Usage |
|---|---:|---|
| `color.bg.canvas` | `#101417` | App canvas and shell background. |
| `color.bg.surface` | `#171D21` | Main table/chart panels and toolbar surfaces. |
| `color.bg.surface-raised` | `#1E272D` | Active filter controls, selected rows, and non-modal raised content. |
| `color.bg.overlay` | `#263238` | Menus, popovers, dialogs, and toasts only. |
| `color.border.subtle` | `#33414A` | Non-interactive dividers and chart grid lines. |
| `color.border.strong` | `#5B6C78` | Input borders, focus-adjacent controls, and UI boundaries requiring at least 3:1 contrast. |
| `color.text.primary` | `#F1F5F7` | Titles, key values, and default body text. |
| `color.text.secondary` | `#B8C4CC` | Labels, helper text, table secondary values. |
| `color.text.muted` | `#91A1AB` | Timestamps and supporting metadata; never below 4.5:1 for text. |
| `color.text.disabled` | `#6F7E87` | Disabled content only; disabled controls must also communicate state non-visually. |
| `color.accent.action` | `#006D78` | The one filled primary action per screen/dialog, active nav, keyboard focus ring. White text on this token meets normal-text contrast. |
| `color.accent.emphasis` | `#39D0DF` | Chart primary series, selected-data emphasis, and non-text highlight on dark surfaces. Not for dense body text. |
| `color.status.success.fg` | `#7EE2B8` | `Normal`/healthy text and icon on dark surface. |
| `color.status.success.bg` | `#123527` | Normal status-chip fill. |
| `color.status.warning.fg` | `#FFD08A` | Warning text/icon on dark surface. |
| `color.status.warning.bg` | `#47310E` | Warning status-chip fill and threshold band. |
| `color.status.critical.fg` | `#FFB4AB` | Critical text/icon on dark surface. |
| `color.status.critical.bg` | `#4A1E1E` | Critical status-chip fill and threshold band. |
| `color.status.info.fg` | `#A9D6FF` | Informational/degraded status text. |
| `color.status.info.bg` | `#19334A` | Degraded-state banner and informational chips. |
| `color.focus.ring` | `#79E6F2` | 2 px visible focus ring with a 2 px canvas offset. |

Semantic status is never decorative. A status chip always combines an icon, a word, and color. Chart series use the cyan accent and threshold bands use low-opacity semantic fills; users must still be able to understand the status from the legend and text summary.

### 3.2 Typography

Use **Inter**, self-hosted through Fontsource rather than loaded from a third-party font CDN. Inter is deliberately chosen for its crisp small-size rendering, clear distinction between common telemetry characters, and tabular-number support. This product needs fast comparison of readings, timestamps, and device IDs more than it needs a personality display face.

| Token | Size / line height | Weight | Usage |
|---|---|---:|---|
| `type.micro` | 12 px / 16 px | 500 | Chart labels, timestamps, table metadata. |
| `type.ui` | 13 px / 18 px | 400 | Default controls, table cells, body copy. |
| `type.body` | 14 px / 20 px | 400 | Explanatory copy and form text. |
| `type.label` | 13 px / 18 px | 600 | Field labels, table headers, status labels. |
| `type.section` | 16 px / 22 px | 600 | Panel headings and secondary page sections. |
| `type.page` | 24 px / 30 px | 600 | Page title. |
| `type.metric` | 28 px / 32 px | 600 | Fleet-level count and device headline reading. |

- Use weights 400, 500, and 600. Do not use 300; use 700 only for a critical modal title if absolutely necessary.
- Apply `font-variant-numeric: tabular-nums` to every measurement, timestamp, count, and numeric table column.
- Device IDs and raw event IDs use the same Inter family with `font-variant-numeric: tabular-nums`; do not introduce a decorative monospace face.
- Keep prose to 60–75 characters per line. Dense data tables may use the full content width.

### 3.3 Spacing, sizing, radius, and depth

| Category | Tokens | Rule |
|---|---|---|
| Spacing | `space.1` 4, `space.2` 8, `space.3` 12, `space.4` 16, `space.6` 24, `space.8` 32, `space.12` 48, `space.16` 64, `space.24` 96 px | Use only these values. 8–12 px separates related controls; 24–32 px starts a new section. |
| Control height | `control.compact` 32, `control.default` 36, `control.touch` 40 px | Default desktop controls are 36 px; primary touch targets and table row actions are at least 40 px. |
| Data-row height | `row.compact` 36, `row.default` 40, `row.comfortable` 48 px | Fleet default is 40 px. Expose compact density only after the base table is correct. |
| Radius | `radius.sm` 4, `radius.md` 6, `radius.lg` 8, `radius.full` 999 px | Inputs/buttons: 6 px. Panels/charts/dialogs: 8 px. Status chips: full. No large “pill card” styling. |
| Borders | `border.default` 1 px, `border.focus` 2 px | Prefer borders to shadows for flat layout grouping. |
| Shadows | `shadow.overlay` `0 12px 28px rgba(0,0,0,.34)`, `shadow.toast` `0 8px 20px rgba(0,0,0,.28)` | Menus, dialogs, popovers, and toasts only. Regular cards have no shadow. |
| Motion | `motion.hover` 100 ms, `motion.overlay` 150 ms, `motion.modal` 200 ms, `ease.standard` `ease-out` | Only transform and opacity; movement is 2–8 px. Respect `prefers-reduced-motion`. |

## 4. Screen inventory

| Screen | Route | Purpose | MVP scope decision |
|---|---|---|---|
| Fleet overview | `/` | Triage the entire fleet by freshness, condition, and current readings. | Core live monitoring surface. |
| Device detail | `/devices/:deviceId` | Investigate one device’s current readings, five-minute aggregates, and status. | Core live monitoring surface. |
| Device history | `/devices/:deviceId/history` | Run a deliberate, bounded PostgreSQL history query for one device. | Manual only; never auto-refreshed. |
| Access denied | Caddy/API response | Explain that the operator lacks dashboard access without leaking telemetry. | Required security state; no app-level login form. |
| Not found | Unmatched app route or unknown device | Return an operator to Fleet when a link/device no longer exists. | Required navigation recovery. |

There is no device-creation, API-key management, alert-rule, tenant, user, or settings screen in MVP. API-key configuration is seeded outside the UI, and alerts are visual indicators only.

## 5. User flows

### 5.1 Triage fleet health

1. Operator opens the Fleet overview; Caddy confirms dashboard access before the SPA loads.
2. The page renders a shell and fleet-table skeleton immediately.
3. The dashboard fetches the Redis-backed fleet summary and device rows.
4. Operator scans the status strip: Critical, Warning, Stale, and Active counts.
5. Operator narrows the table with search or a status filter; filter count and applied filter chips remain visible.
6. Operator sorts by status, last processed time, or a chosen measurement.
7. Operator opens the affected device with a real row link or the row’s visible `View device` action.

**Design choice:** The primary interaction is table navigation, not a filled CTA. A monitoring page should not invent a “create” action when observation and triage are the actual jobs.

### 5.2 Investigate an abnormal device

1. Operator reaches Device detail from a fleet row or a direct URL.
2. Header shows device ID, current overall state, and last processed time before charts.
3. Operator compares the four headline readings and their threshold labels.
4. Operator scans the compact 2×2 five-minute chart grid; threshold bands and labels make the breached condition visible.
5. Operator selects `View history` to initiate a deliberate database query if live data needs context.
6. Operator returns to the previously filtered fleet via the breadcrumb; preserve table filters, sort, density, and scroll position.

**Design choice:** Device state is summarized before charts because a chart is evidence, not the first answer. The live chart remains five-minute only because multi-month analysis is outside MVP.

### 5.3 Query manual device history

1. Operator selects `View history` from Device detail.
2. History page carries the device context and starts with no database request beyond the device identity.
3. Operator selects a bounded time preset or valid custom range.
4. Operator selects `Run query`; the button changes to `Loading history` and remains disabled until completion.
5. Results render in an indexed, paginated data table with timestamp, metric, value, unit, event ID, and received time.
6. Operator refines the range/measurement filter or returns to live detail.

**Design choice:** History is an explicit action rather than automatic background work. This makes database cost visible and protects the ingestion workload.

### 5.4 Interpret degraded live telemetry

1. Redis becomes unavailable; the live API reports `mode: degraded` or fails with a documented degraded response.
2. The SPA keeps the last successful fleet/device query in view, marks it `Stale`, and displays the persistent degraded banner.
3. The banner states that live aggregates are paused and shows the last successful refresh time.
4. Auto-refresh retries conservatively in the background; it never falls back to PostgreSQL polling.
5. When live data resumes, the banner changes to a brief `Live data restored` status and the timestamp updates.

**Design choice:** Last-known data is more useful than a blank screen, but it is never presented as live. The interface must not conceal the Redis failure boundary in the PRD.

### 5.5 Device integration feedback (non-UI flow)

1. A device posts to `POST /v1/telemetry/batches` with its API key.
2. A `202` response confirms only in-memory admission.
3. The device responds to `400`, `401`, `413`, and `415` by correcting configuration/payload; it backs off after `429` or `503`.
4. Operators observe the resulting device only after telemetry is processed into Redis.

**Design choice:** There is no dashboard “ingest test” form. Device credentials must never enter a browser UI.

## 6. Per-screen layout

### 6.1 Shared application shell

- **Desktop:** A 64 px top bar. Left: product mark and `Fleet` link. Right: compact system-mode chip, last live refresh timestamp, and operator access indicator. No permanent sidebar: MVP has only one top-level operational destination, and a sidebar would waste data width.
- **Context:** Breadcrumb appears on Device detail and History only: `Fleet / {device ID} / History`.
- **System banner:** Appears directly beneath the top bar only for degraded or unavailable live state. It is persistent, full-width inside the content frame, and uses informational semantic color.
- **Content frame:** 24 px desktop outer padding, maximum width 1440 px. Fleet table may use all available frame width.

### 6.2 Fleet overview

| Area | Layout and hierarchy | Components |
|---|---|---|
| Page header | Left-aligned `Fleet telemetry` title and a secondary line: `Live aggregate view` plus last refresh. No fabricated filled CTA. | `PageHeader`, `SystemModeChip`, `LastUpdated`. |
| Status strip | Four equal compact metric cells: Critical, Warning, Stale, Active. Critical appears first, then Warning, then Stale, then Active. Cells are links that apply the matching table filter. | `FleetStatusMetric`, `StatusIcon`, `LinkButton`. |
| Toolbar | Search by device ID/name, status filter, measurement selector, clear filters, result count, optional density toggle. Align search/filters left; utility controls right. | `SearchField`, `FilterChip`, `Select`, `ClearFilters`, `DensityToggle`. |
| Fleet table | Sticky header; default columns: Device, Overall state, Temperature, Voltage, Battery, Pressure, Last processed. Text left-aligned; measurement and timestamp values right-aligned with tabular numerals. Rows link to Device detail. | `DataTable`, `StatusChip`, `MeasurementCell`, `FreshnessCell`, `TableSortButton`. |
| Footer | Server-provided count and pagination only when result count exceeds page size. | `TablePagination`. |

**Primary action:** none. The row link is the primary operational action. Filling a button would imply a workflow the operator should perform instead of prioritizing triage.

### 6.3 Device detail

| Area | Layout and hierarchy | Components |
|---|---|---|
| Breadcrumb and header | Breadcrumb above device title. Title row contains device ID/name, overall status chip, and last processed time. `View history` sits at top-right as the single filled primary action. | `Breadcrumb`, `DeviceHeader`, `StatusChip`, `Button`. |
| Key-facts strip | Four equal measurement facts: current value, unit, threshold label, and freshness. Do not use oversized circular gauges; numeric value is the fastest comparison surface. | `MeasurementFact`, `ThresholdLabel`, `FreshnessLabel`. |
| Chart grid | Desktop 2×2 grid: Temperature, Voltage, Battery, Pressure. Each chart has a title, latest value, five-minute line, threshold bands, and an accessible text summary. | `TelemetryChart`, `ChartLegend`, `ChartSummary`. |
| Supporting state | Below charts: aggregate update time and `Data source: Redis live aggregate`. During degradation, this is replaced by stale timestamp/context. | `MetadataList`, `DegradedBanner`. |

**Primary action:** `View history`. It is the one follow-up that legitimately changes the operator’s analysis depth and makes database work explicit.

### 6.4 Device history

| Area | Layout and hierarchy | Components |
|---|---|---|
| Header | Breadcrumb, `History for {device ID}`, and supporting text: `Manual query - not live data`. | `Breadcrumb`, `PageHeader`, `InfoLabel`. |
| Query panel | Single-row desktop filter bar: time preset, start/end range when custom, measurement filter, `Run query`. Mobile stacks this panel. Keep the page controls rather than a modal. | `DateRangePicker`, `Select`, `Button`, `FieldError`. |
| Result summary | Visible selected range, measurement scope, count, and query completion time. | `QuerySummary`, `StatusText`. |
| History table | Sticky header and server-side pagination. Columns: Event time, Measurement, Value, Unit, Received time, Event ID. Numeric values right-aligned; event IDs truncate with copy button/tooltip. | `DataTable`, `TimestampCell`, `CopyButton`, `TablePagination`. |

**Primary action:** `Run query`. It is disabled until the chosen range is valid and becomes `Loading history` during request execution. The server remains authoritative for maximum range and result limits.

### 6.5 Access denied and not found

| Screen | Layout | Primary action |
|---|---|---|
| Access denied | Narrow centered content block with `Access required`, plain explanation, and contact path supplied by deployment configuration. Never show login fields, token diagnostics, or telemetry metadata. | `Return to safe location` only when an authenticated route exists. |
| Not found | Narrow centered content block with `Device not found` or `Page not found`, a short explanation, and `Return to fleet`. | `Return to fleet`. |

## 7. Component library

Build every primitive on accessible Radix/shadcn patterns and style it with the token system. Use Lucide as the only icon family: 16 px inline, 20 px standalone. An icon-only control always has an `aria-label` and visible-on-focus tooltip.

| Component | Variants | Required states |
|---|---|---|
| `AppShell` | normal, degraded | desktop, tablet-collapsed, mobile, skip-link target. |
| `PageHeader` | fleet, detail, history | default, with breadcrumb, with action, narrow. |
| `Button` | primary, secondary, ghost, destructive, icon-only | default, hover, focus-visible, pressed, disabled, loading. Primary exists once per page/dialog. |
| `StatusChip` | normal, warning, critical, stale, degraded, unknown | icon + label + semantic color; default and compact. |
| `SystemModeChip` | live, degraded, unavailable | default, updating, last-known timestamp. |
| `FleetStatusMetric` | critical, warning, stale, active | default, selected-filter, zero count, loading. |
| `MeasurementFact` | temperature, voltage, battery, pressure | normal, warning, critical, stale, unknown. |
| `MeasurementCell` | temperature, voltage, battery, pressure | normal, warning, critical, missing (`—`), stale. |
| `FreshnessCell` | active, stale, unknown | timestamp plus text; never color-only. |
| `TelemetryChart` | each supported measurement | loading skeleton, live, stale, degraded, no samples, error. |
| `DataTable` | fleet, history | loading, first-use empty, no-filter-results, error, selected row, keyboard-active row, paginated, horizontally scrollable. |
| `TableSortButton` | ascending, descending, unsorted | default, hover, focus-visible, active; exposes `aria-sort`. |
| `SearchField` | device search, history filter | empty, populated, clearable, invalid, disabled. |
| `FilterChip` | status, measurement, time range | available, selected, removable, disabled. |
| `Select` / `DateRangePicker` | toolbar, history query | default, open, focused, invalid, disabled. |
| `DensityToggle` | default, compact | selected, keyboard-focusable; only on data tables. |
| `DegradedBanner` | Redis offline, stale last-known data, recovered | persistent degraded, retrying, recovery confirmation. |
| `InlineError` | field validation, query failure, API failure | inline, retryable, non-retryable. |
| `EmptyState` | first use, no filter results, no chart samples | icon, title, short next step, optional clear-filter action. |
| `Skeleton` | table, metric, chart, header | fixed geometry matching final component; reduced-motion static alternative. |
| `Tooltip` | icon clarification, truncated text | keyboard and pointer trigger, delay <=150 ms. |
| `Toast` | success, error, info | non-blocking, dismissible, announced once; never used for persistent degraded status. |
| `ConfirmDialog` | destructive future action only | closed, open, focus-trapped, confirm loading. Not needed by current MVP flows. |

### Component rules

- Components accept semantic props (`status="critical"`, `density="compact"`), not ad hoc class strings that bypass tokens.
- The fleet/history `DataTable` is one reusable implementation. It supports sortable headers, filter toolbar slots, loading/empty/error states, sticky header, column alignment, and server-side pagination.
- Build no dropdown, dialog, combobox, tooltip, or date picker from scratch. Their keyboard and focus behavior is an accessibility surface, not a visual detail.
- Create a Storybook story for every approved component, each variant, and every state listed above before treating the component as reusable.

## 8. States

### 8.1 Key-screen state matrix

| Screen | Empty | Loading | Error | Success | Offline / degraded |
|---|---|---|---|---|---|
| Fleet overview | **First use:** `No telemetry received yet` with API integration guidance link (no embedded keys). **Filtered:** `No devices match these filters` with `Clear filters`. | Status-strip and table skeletons preserve table geometry; no full-page spinner. | Inline table error: `Fleet data could not be loaded` + `Retry`. Preserve search/filter controls. | Status strip and live fleet table; show last refresh time. | Retain last successful rows, apply stale treatment, show persistent Redis-offline banner and timestamp. If no cached data exists, show degraded empty state, not a fake zero-device fleet. |
| Device detail | `No live aggregate for this device` plus `Return to fleet`; distinguish never-seen device from expired/offline aggregate if API can do so. | Header/fact/chart skeletons match final 2×2 chart grid. | Inline: `Device data could not be loaded` + `Retry`; breadcrumb remains usable. | Header, facts, charts, and current status. | Retain last successful facts/charts with stale timestamp and persistent degraded banner; never query PostgreSQL automatically. |
| Device history | `Choose a time range and run a query` before any request. After valid query with no records: `No events in this range`; preserve filters. | Results-table skeleton; `Run query` changes to `Loading history`. | Plain API/query error next to filter bar + `Try again`; retain selected dates and measurement filter. | Result summary, table, and pagination. Use a modest `History updated` toast only after a manual successful query if feedback is necessary. | If PostgreSQL is unavailable, show `History is temporarily unavailable`; live Device detail remains navigable. Do not label this Redis degradation. |
| Access denied | Not applicable. | Minimal access-check skeleton only if app shell has already loaded. | `You do not have access to this dashboard`; no technical authorization detail. | Not applicable. | Not applicable. |

### 8.2 State behavior rules

- Do not clear previously loaded data while a refresh is pending. Apply a small `Updating` label next to the timestamp instead.
- Use skeletons for expected loading and plain-language errors for unexpected failures.
- A zero count is data, not an empty state. For example, `0 critical devices` is a valid loaded fleet metric.
- `Unknown`, `Stale`, and `Offline` are distinct: unknown means no valid cached aggregate, stale means data exists but last processed time is over 60 seconds, and offline/degraded refers to the system’s inability to provide current live aggregates.
- Use toast feedback only for a user-initiated action. Never hide a persistent platform failure in a transient toast.

## 9. Responsive behavior

The primary design target is a 1440 px operator desktop. It must remain usable at 1280 px. Tablet and mobile preserve the operational hierarchy; they do not turn data tables into decorative cards.

| Breakpoint | Behavior |
|---|---|
| Desktop: `>=1280 px` | 1440 px content frame, 2×2 chart grid, full fleet/history table, all measurement columns visible, table toolbar in one row. |
| Tablet: `768–1279 px` | Maintain top bar; reduce outer padding to 16 px; fleet status strip becomes 2×2; device chart grid remains 2 columns when space permits; toolbar wraps into two purposeful rows; lower-priority table columns may hide behind `Columns`. |
| Mobile: `<768 px` | Top bar keeps product name and system state; page header actions stack below title; status strip is 2×2; charts become one column; device facts become two columns. Tables remain tables inside a clearly labeled horizontal scroll container rather than becoming hard-to-compare cards. Keep essential first column and status visually sticky where technically practical. |

Additional responsive rules:

- Search and filter controls stack in visual/tab order on mobile; no control becomes icon-only unless it retains an accessible name and tooltip.
- A row remains at least 40 px high. Do not reduce text below 12 px to fit a narrow viewport.
- Chart labels may simplify on mobile, but the accessible text summary remains complete.
- Preserve query filters and scroll position when moving between Fleet and Device detail at every breakpoint.
- Test long device IDs, missing values, nine-digit readings, 500-event batch-linked counts, 1,000+ table rows, and browser zoom at 200%.

## 10. Accessibility

### Contrast and non-color communication

- Meet WCAG 2.2 AA: at least **4.5:1** for normal text and **3:1** for large text, interactive-component boundaries, focus indicators, and essential chart marks.
- Validate the final rendered color combinations, including status foreground/background pairs, in automated and manual contrast checks. Do not rely on approximate palette intent.
- Status always uses text plus icon plus color. A critical reading is not “the red one”; it is `Critical` with its measurement and threshold context.
- Charts have a visible legend, threshold labels/bands, and an adjacent text summary containing latest value, status, and time. Do not make pointer hover the only way to obtain a value.

### Focus order and keyboard navigation

- Include a visible `Skip to fleet content` link as the first focusable element.
- Tab order follows visual order: top-bar links, page header/action, status strip, toolbar controls, table/chart content, footer controls.
- Never remove outlines without the `color.focus.ring` replacement. Focus must be visible on dark surfaces and not clipped by panel overflow.
- `Enter` activates buttons, links, selected table rows, and runs a valid history query. `Escape` closes menus, popovers, tooltips, and dialogs.
- Sortable headers are real buttons. They expose `aria-sort` and support `Enter`/`Space`.
- Fleet/history rows use real links to support keyboard activation and Ctrl/Cmd-click. Do not implement a `div` with a click handler as a row link.
- When table-row selection exists in a future bulk-action workflow, use checkbox semantics and support Shift+Arrow/Space selection. MVP has no bulk-action bar.

### ARIA and semantic requirements

- Use landmark elements: `header`, `nav`, `main`, and `footer` where relevant; each page has one `h1`.
- Use native table semantics with `caption`, `thead`, `th scope`, and `aria-sort`; do not use generic grid roles unless interaction genuinely requires grid behavior.
- Inputs always have visible labels. Placeholders are examples, never labels. Validation messages are linked with `aria-describedby`; invalid fields use `aria-invalid`.
- Status/degraded banners use `role="status"` or an appropriate live region. Announce a state change once; do not repeatedly announce every polling refresh.
- Chart containers require an accessible name, short summary, and a data-table/text alternative. The SVG must not expose redundant decorative paths to screen readers.
- Icon-only controls require `aria-label`; decorative icons use `aria-hidden="true"`.
- Dialogs are focus-trapped, return focus to their trigger on close, and provide a visible close button plus `Escape` behavior.

### Motion, input, and error prevention

- Respect `prefers-reduced-motion`; skeletons become static blocks and overlays have no travel animation.
- Keep transitions under 200 ms and animate only opacity/transform.
- Validate History fields on blur and on submit, not on each keystroke. Preserve user values after an error.
- Disable `Run query` only while the query is running or the form is invalid, and explain the disabled state through visible helper text.
- Do not auto-refresh the manual history result in a way that changes reading order or focus.

## Enforcement checklist for future UI work

Before a UI change is considered complete, verify:

1. Every spacing, color, type, radius, border, depth, and motion value comes from this token system.
2. The page uses an existing shell and component pattern; it does not introduce a one-off layout or primitive.
3. New data views include success, loading, empty, error, and degraded/offline behavior where relevant.
4. Desktop at 1440/1280 px, tablet, mobile, keyboard-only, 200% zoom, and reduced-motion behavior have been checked.
5. Text and essential UI boundaries meet contrast targets; status is understandable without color.
6. The implementation does not expand PRD scope or claim cache-backed data is durable/current when the system says otherwise.
