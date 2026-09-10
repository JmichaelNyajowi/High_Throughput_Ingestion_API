---
name: verify
description: Drive a feature in a real browser with Playwright — reusing saved auth, checking the mechanical states, screenshotting every step — then hand the user only the judgment calls. Use after a ticket with a multi-step flow or multiple roles.
---

# Verify

**Confirm the feature works.** Not whether it's good — that's `/critique` and the user.

The point is to **remove the setup cost**. Manual testing gets skipped because it needs a running server, seeded data, several accounts, and remembering the click path. That's ten minutes before you learn anything. This does all of it so the user spends ninety seconds looking.

## 1. Work out what to check

From the ticket's acceptance criteria and the diff. Cover:

- The happy path through the feature
- Each acceptance criterion
- The **lifecycle actions** the ticket declared — if it says records are deletable, verify the delete works and the confirm names the specific record
- Anything involving a second role or permission

## 2. Set up

- **If running in a git worktree** (parallel `/build`/`/verify` sessions
  across sibling worktrees is common on this project — check `git branch
  --show-current` against `main`), start the dev server on that ticket's
  own port (`pnpm dev -- -p 30<NN>`), not the bare default — see
  `.claude/skills/build/SKILL.md` step 0 for the full per-worktree
  port/database rules this project uses. In the main checkout, the
  default port is fine.
- Ensure seed data is loaded; run the seed command if not
- **Reuse the saved auth state** from the Playwright setup. Never log in through the UI unless login is what's being tested
- **This starts a real dev server and browser process** — resource-
  intensive, same as `/build` step 0a's gate. If this session hasn't
  already gotten approval to run heavyweight processes, ask once before
  starting.

## 3. Drive the flow

**Prefer the Playwright MCP server** (`.mcp.json`, once approved) —
direct navigate/click/screenshot tool calls instead of writing a script.
It runs over stdio (no listening port), so it doesn't collide with a
sibling worktree's dev server or Storybook port. **Fall back to a
throwaway Playwright script** (`docs/gotchas.md`'s Testing/Playwright
section has the pattern) if the MCP server isn't available or approved
for this session. Either way: screenshot **every** step, including
passing ones — the screenshots are the output, and they let the user
review a whole flow in thirty seconds without running anything.

## 4. Check the mechanical states

- **Empty — first use** (no records at all)
- **Empty — no results** from a filter, and confirm it's a *different* message from first-use
- **Loading** — a skeleton appears, and it doesn't shift the layout when real content lands
- **Error** — trigger a failure; confirm typed input is preserved
- **Permission-denied** — log in as a role that shouldn't have access
- **Long text** — the 200-character name from the seed data doesn't break the layout
- **Zero rows and many rows**

## 5. Report

Pass/fail per step, with the screenshot for each. Failures first, with the screenshot and the actual vs expected.

## 6. Hand over the judgment calls

A short list of what the user should eyeball, with the URL and login. Things a script can't judge:

- Does the flow feel like the right number of steps?
- Is the primary action obvious?
- Does the empty state say something useful?
- Does anything look wrong in a way a rule wouldn't catch?

```
Verified — 11/12 steps passed. Screenshots in <path>.

FAILED step 7: filtering to "Overdue" returned 5 rows, expected 3
  → two invoices due today are being counted as overdue

Worth your eyes (90 seconds):
http://localhost:3000/invoices — sarah@acme.test / test1234
  · The invoice detail page — does the key-facts strip show the right things?
  · Creating an invoice takes 4 steps. Feels like it could be 2 — your call.
```

## Avoiding flaky tests

These are the three standard causes, and designing them out is most of this skill's value.

**Auth.** Log in once in a setup script, save the browser state, every test reuses it. Logging in through the UI in every test means one broken login produces forty failures.

**Selectors.** Target `data-testid` only. Never CSS classes, never visible text — both break on cosmetic changes that have nothing to do with the behaviour being tested.

**Timing.** **Never a fixed wait** (`waitForTimeout`). Always wait for a condition — `await expect(row).toBeVisible()`. Playwright auto-waits; most flakiness comes from fighting that.

If any of the three preconditions is missing from the project, say so and point at Foundation rather than working around it.

## Throwaway by default

Most verification scripts serve one ticket and are deleted afterwards.

**Promote to the permanent E2E suite only for a critical path worth guarding forever** — the five to ten paths: login, the main create flow, the main list view, anything touching money or permissions.

**Ask before promoting.** This is what stops a brittle suite accumulating that nobody wants to maintain.

## Never

- **Never log in through the UI as setup**
- **Never use a fixed wait**
- **Never target CSS classes or visible text**
- **Never claim a flow is good** — verification is mechanical; judgment belongs to the user
- **Never promote a script to the permanent suite without asking**
