---
name: build
description: Claim one eligible ticket, implement it on its own branch, self-verify, and merge — without waiting for /review unless you choose to run it first. Use during the bootstrap build pass after /tickets, and repeatedly during steady state whenever tickets exist in planned status.
---

# Build

The steady-state workhorse: re-run every time `/tickets` cuts a new
feature's tickets, or `/fix`/`/add` hand off new ones.

## Purpose

Take one ticket from `planned` to merged on `main`: implement it, verify
it locally, and merge — without making scope or architecture decisions
that aren't already settled in `docs/conventions.md`, `docs/architecture.md`,
or the ticket itself.

**Track every step below in TodoWrite as you go** — add the steps up
front, check each off the moment it's done. Don't defer this to the end.
**Kindly DO NOT FORGET to use a TodoWrite checklist** — it is how the use get feedback on your progress without having to ask for it and waste time. If you don't use a checklist, the user will have to ask you for updates and you will be blocked until you respond.

## Timing log

Append a block to `.work/build-timings.md` (create it if missing) for
every ticket built, so timing is comparable across tickets without
relying on git history. Record a timestamp (`HH:MM`, local time) at each
milestone below as you reach it — don't backfill at the end:

- `claimed` — when Status is set to `in-progress` (Claiming step 3)
- `context read done` — once the ticket, its Context pointers (design
  reference, relevant module queries/schema/conventions), and any
  scoping/precedent checks are read and resolved (Process step 1)
- `plan approved` — when the user approves the plan and implementation
  starts (Process step 2); log `plan proposed` at the same time as a
  separate line if there's a real gap before approval
- `tests written` — once failing tests are committed, for test-first
  tickets only (Process step 3); omit this line for test-after tickets
- `implementation done` — once the logic/routes/wiring described in the
  ticket's Scope works end to end, before any Storybook story or docs
  update (Process step 3)
- `ui-polish done` — once the Storybook story and `docs/screens.md` (or
  equivalent docs) are updated, for UI tickets only; omit for
  logic/plumbing-only tickets
- `self-verify done` — when tests/lint/typecheck/UI-check all pass
  (Process step 5)
- `merged` — when the branch lands on `main` (Process step 6)

Compute each row's duration from the previous timestamp, and label the
`plan proposed → plan approved` gap as "waiting on user" rather than
folding it into implementation time — that gap isn't agent work and
shouldn't look like it. Use this format (omit `tests written` /
`ui-polish done` lines that don't apply to the ticket):

```
## Ticket <id> — <short title>
- claimed: 10:03
- context read done: 10:11  (context: 8m)
- plan proposed: 10:13
- plan approved: 10:16  (waiting on user: 3m)
- tests written: 10:24  (tests: 8m)
- implementation done: 10:44  (implement: 20m)
- ui-polish done: 10:52  (ui-polish: 8m)
- self-verify done: 11:05  (verify: 13m)
- merged: 11:07  (merge: 2m)
```

If the ticket is blocked, rejected by `/review`, or resumed across a
session gap, add a line noting it (e.g. `blocked: 14:20 (design
direction)` / `resumed: next session, 09:10`) instead of letting the gap
silently inflate whichever bucket it falls into.

## Claiming a ticket

1. Scan `.work/*.md` for tickets with `Status: planned` whose declared
   dependencies are all `Status: done`.
2. Pick one. If running as part of a deliberately parallel/looped batch,
   prefer tickets on unrelated areas of the codebase over multiple
   tickets touching the same module, to reduce merge conflicts.
3. Set `Status: in-progress`, note the session/agent and timestamp, commit
   this status change immediately — this is the claim. If another agent
   later finds this ticket already `in-progress`, it moves on to a
   different eligible ticket rather than duplicating work. Record the
   `claimed` timestamp in the timing log now (see Timing log below).
4. Create a branch for this ticket (e.g. `ticket/<ticket-id>-<slug>`),
   off current `main`.

## Process

1. **Read the ticket in full**, plus everything its Context section points
   to: the relevant feature folder, `docs/conventions.md`, relevant
   sections of `docs/architecture.md`.

   **If the ticket builds a new screen or changes an existing one**,
   check `docs/screens.md` for its `Status: approved` entry, then open
   the matching Storybook story under `src/modules/<feature>/ui/` (or
   `components/layout/` / `components/patterns/` for shells and shared
   patterns). Build to match that story — layout, components used,
   states — rather than re-deriving the screen from `docs/conventions.md`
   alone. `docs/screens.md` is the single inventory; there is no separate
   design-reference location to check.

   **If nothing matches, stop and ask** rather than inventing the
   composition silently — same rule as `CLAUDE.md`'s "if a needed pattern
   doesn't exist, STOP and ask." Set `Status: blocked`, write into the
   ticket file which screen has no precedent, and present the user with
   options rather than a bare "blocked" notice:
   - the closest existing pattern/screen this could reuse or adapt, if
     one exists, and
   - an offer to design 2-3 variants for this screen (in the style of
     the project's existing screens, per `docs/conventions.md` and
     `references/design-principles.md`) for the user to pick from.

   Default toward offering variants when the screen is non-trivial —
   that's usually what's wanted — but the user may instead approve
   building the closest existing pattern directly; don't assume either
   way, ask.

   If the ticket is set to `Status: blocked` at this point or any later
   point, add a `blocked: <time> (<reason>)` line to the timing log entry
   before stopping.

   **If the ticket's own cited precedent (a function or screen it says
   to mirror) turns out to be wrong or buggy**, that's a scope question,
   not a local judgment call — stop and ask whether to fix the precedent
   alongside the new work or leave it and only build the new code.

   Record `context read done` in the timing log once this step's reading
   and scoping checks are resolved.

2. **Decide test-first or not, then state a plan and wait for approval.**
   Tickets don't reliably carry a `TDD: true/false` field — decide per
   task instead: logic, calculations, and permission checks get tests
   written first (see below); plumbing, wiring, and pure UI composition
   get tests after, per the Verification section. Then present a short
   plan — files/modules touched, approach, the test-first-or-not call and
   why, and (for UI tickets) which approved story is being matched — and
   wait for explicit approval before writing any code. Record `plan
   proposed` in the timing log when the plan is handed to the user, and
   `plan approved` the moment they approve it.

3. **Implement the vertical slice** described in the ticket's Scope. Follow
   `docs/conventions.md` — use existing patterns and the canonical
   reference module(s) from `/foundation` rather than inventing new
   approaches. If the ticket is responsible for wiring a feature's data
   into a screen slot on a different feature's shell (per the cross-feature
   note in `/tickets`), do that wiring as part of this same ticket, not a
   separate one.

   **If test-first:** for the logic portion of the ticket, write the
   tests implied by the Acceptance Criteria *first*, run them, and confirm
   they fail for the expected reason (not a typo or setup error — an
   actual absence of the behavior being specified). Record `tests
   written` in the timing log once the failing tests are committed. Only
   then implement, iterating until they pass. Do not write the
   implementation first and backfill tests afterward — that defeats the
   purpose of choosing test-first in the first place. Pure UI-composition
   portions of a mixed ticket don't need this discipline even when the
   logic portion is test-first — apply it to the logic, not the markup.

   When logic accepts a batch/list input, include at least one test case
   with a repeated key in that batch — a common gap otherwise.

   **If test-after:** implement normally. Tests are still required by
   the Verification section below, just not written test-first.

4. **Stop and flag instead of guessing** whenever something in the ticket
   conflicts with `docs/conventions.md` or `docs/architecture.md`, or
   requires a scope decision the ticket doesn't already make. Do not
   silently resolve this by picking an approach. Set `Status: blocked`,
   write the specific conflict/question directly into the ticket file, and
   surface it — this is the same "ask, don't assume" principle used in
   every earlier phase, just triggered mid-implementation. Local
   implementation judgment calls (naming a variable, structuring a helper
   within the established pattern) are fine and don't need escalation —
   the bar is: does resolving this require a decision `/alignment` or
   `/design` should have made.

5. Record `implementation done` in the timing log now that the logic,
   routes, and wiring described in the ticket's Scope work end to end.
   If the ticket touches UI, finish the Storybook story and
   `docs/screens.md` (or equivalent docs) update now and record
   `ui-polish done`; skip this line for logic/plumbing-only tickets.

   Then **self-verify**:

   **Tests.** Run only the test file(s) this ticket's logic touches
   (`vitest run --project integration <path>`), not the full suite —
   the integration suite runs sequentially against a shared test DB
   (`fileParallelism: false`, see `vitest.config.ts`) specifically
   because it's slow to run end to end, so treat a full `pnpm test` as
   a single pre-merge gate, not a rerun-after-every-change reflex. If a
   full-suite run already passed earlier this session and nothing
   outside this ticket's files changed since, that earlier pass still
   counts — don't rerun it a second time before merge. Lint and
   typecheck run at their normal (fast, whole-project) scope — there's
   no equivalent partial mode for those.

   **UI check.** Remember `/design` already built, self-critiqued, and
   got approval for this screen's composition against
   `references/ui-rules.md` — that work is done and committed. The user
   also typically has the dev server open and is watching changes land
   on localhost live as `/build` works, which already covers casual
   visual confirmation. Three tiers, from cheapest to most expensive —
   pick the cheapest one that actually fits:

   - **Pure reuse/wiring, no new state or composition at all** (the
     ticket only wires an already-approved story into a new slot, no new
     markup): confirm the dev server (`pnpm dev`) is running — start it
     if not, reuse an already-running instance rather than starting a
     second one — then tell the user the exact URL/route for this
     ticket's screen (not just the app root), e.g. "the store ledger tab
     is at `localhost:3000/ledger?tab=store`". Do not launch Storybook or
     any browser automation. UI check satisfied on this basis alone.
   - **Near-copy of an already-approved precedent** (the ticket builds a
     new component whose composition closely mirrors one that already
     has an approved, screenshotted Storybook story — same table shape,
     same fetch/LoadState split, same states, just new data/fields —
     which is the common case for a new ledger/list/detail screen built
     from an established pattern): **skip Storybook and screenshots
     entirely.** Read the new component side-by-side with its cited
     precedent file and confirm structural parity — same states covered
     (loading/error/denied/empty/populated), same token usage, same
     interaction patterns, no invented markup. This is a code-reading
     check, not a visual one: open both files, diff them mentally
     against `references/ui-rules.md`'s checklist, and confirm the new
     file is doing what the precedent does, adapted to new fields. This
     is the case that actually catches real problems (a missing state, a
     structural drift from the approved pattern) at zero runtime cost —
     don't spend time booting a browser to re-confirm what reading the
     two files side by side already tells you.
   - **Genuinely novel composition, no close precedent** (a new state,
     a layout `/design` never built or approved): this is the only tier
     that still needs the automated Storybook pass, since it's the one
     case reading code can't substitute for — layout math (overflow,
     truncation, sticky behavior) that doesn't evaluate from source.
     Follow the Storybook procedure below. (A checked-in visual-snapshot
     regression test, so this becomes a pass/fail check instead of an
     agent screenshotting and eyeballing it fresh every time, is planned
     but not yet built — until then this tier still means manual
     screenshots.)
   - **If unsure which tier applies**, ask rather than guessing.

   Record `self-verify done` once tests/lint/typecheck pass and the UI
   check (whichever form it took) is satisfied.

   ### Storybook procedure (new UI surface only)

   1. Start Storybook (`pnpm storybook`) and note the port it prints.
      Check `ps aux | grep storybook` first — a prior step in this same
      session may have already left one running; reuse it rather than
      starting a second instance. The `pnpm storybook` script hardcodes
      `-p 6006`; if that port is taken, it prompts interactively rather
      than picking a new one, so an unattended second attempt will hang.
      If you need a specific port, call the underlying binary directly
      (`pnpm exec storybook dev -p <port>`) rather than passing `-p` after
      `pnpm storybook`, which the npm script's own flag already claims.
   2. Fetch `http://localhost:<port>/index.json` and read the real story
      IDs from it — **don't guess an ID from the story name**; Storybook's
      slugging is more aggressive than it looks (e.g. "Empty — first use"
      becomes `--empty`, not `--empty--first-use`).
   3. Drive it with whatever native browser capability the running agent
      already has (e.g. Codex's built-in browser). If none is available,
      or it's having issues, fall back to the Playwright MCP server. If
      that also isn't available or approved, fall back to a throwaway
      script that `import { chromium } from "playwright"` and
      `chromium.launch({ args: ["--no-sandbox"] })` directly — run it
      with `node` from inside the repo (an absolute path *into* the repo,
      or `cd` there first); running it from outside the repo, e.g. a
      scratch dir, fails dependency resolution even though the script
      looks identical. Delete the throwaway script when done — it's not
      part of the ticket's diff.
   4. If a fallback reports Chromium isn't installed, check
      `ls ~/.cache/ms-playwright/` before installing anything — it's a
      shared home-directory cache, so the binary is very likely already
      there. Never install system Chrome as a fix (needs `sudo`, and
      neither the MCP server nor `@playwright/test` use it anyway). Only
      run `npx playwright install chromium` if the cache is genuinely
      empty. Note that the Playwright MCP server and the `playwright` npm
      package can pin different Chromium build numbers — the cache may
      hold a build the MCP server rejects as missing even though
      `@playwright/test`'s own launch would accept it; if the MCP server
      complains a specific build isn't installed but the cache has a
      *different* build number, don't install — drop straight to the
      throwaway-script fallback instead of chasing the exact version.
   5. If the MCP server seems to hang or stall on first connect, that's
      usually `npx` resolving the package fresh, not a missing
      dependency — give it a moment before concluding it's broken.
   6. On the very first navigation in a fresh browser/page, a Storybook
      iframe can still show its own loading spinner even after
      `networkidle` fires (the index fetch and story bundle load happen
      after that event, not before it). Add a short explicit wait (a few
      hundred ms is usually enough) before screenshotting the first story
      in a run — later navigations in the same page don't need it.
   7. Load the story/route with real data flowing through it and confirm
      it still matches the approved story's shape — no layout shift, no
      missing/misrendered state, no error. Screenshots confirm this, not
      underlying correctness (e.g. that a total actually equals quantity
      × cost) — reason about that from the code together with the
      screenshot, not the screenshot alone.
   8. Run a full `references/ui-rules.md` item-by-item audit of the new
      surface specifically — the new state or composition, not the whole
      screen — since this is new UI `/design` never reviewed.

   This is mechanical verification; it is not a substitute for `/review`,
   which checks things a machine can't and is optional per the merge step
   below.

6. **Merge.** Once self-verify passes, merge the branch to `main` and set
   the ticket's `Status: done`. `/review` is optional and user-invoked —
   run it yourself before this step if you want that ticket gated, or
   after merge as a post-hoc audit; `/build` doesn't wait for it by
   default. Record `merged` in the timing log to close out the entry.

## On review feedback

If `/review` is run (before or after merge) and rejects the work, `/build`
resumes: read the review findings, address them on the same branch (or a
follow-up branch if already merged), re-verify, and merge again. Add a
`rejected: <time> (<reason>)` line to that ticket's timing log entry, and
a fresh `implementation done` / `self-verify done` pair once the rework
is complete.

If `/build` resumes a ticket found already `in-progress`/`blocked` from
an earlier session, add a `resumed: <time>` line rather than treating the
gap since the last timestamp as work time.

## Output

One merged change per ticket, ticket status accurately reflecting where
it is in the pipeline (`in-progress` / `blocked` / `done`).

## Explicit non-goals

- No scope or architecture decisions — flag and stop instead.
- No working multiple tickets in one session/branch — one ticket, one
  branch, one merge, keeps blast radius contained.
- No writing code before the plan (step 2 of Process) is approved.
