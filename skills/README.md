# Skills

A reusable AI-agent-driven software development workflow, for enterprise
web development projects. Portable — copy this whole `.claude/skills/`
folder into any new or existing project.

Full worked example: [`workflow.md`](../../workflow.md) at the repo root
walks a realistic project through every skill below, end to end.

## Start here: which entry point applies?

**Starting a brand new project, nothing built yet:**
Start with [`/notes`](notes/SKILL.md). Follow the bootstrap sequence below
in order.

**Picking up an existing, in-progress project** (real code already
shipped, possibly with rougher/older docs from an earlier version of this
workflow):
Start with [`/adopt`](adopt/SKILL.md) instead. Run it once to reconcile
old docs and real code onto the current doc set, then continue with the
steady-state skills below.

## Bootstrap sequence (fresh projects only, run once, in order)

| Skill | Does | Produces |
|---|---|---|
| [`/notes`](notes/SKILL.md) | Structures a raw brain dump into internal notes | `docs/notes.md` |
| [`/alignment`](alignment/SKILL.md) | Drives full technical + scope alignment as a senior architect | `docs/architecture.md`, `docs/scope.md`, `docs/conventions.md`, `docs/release.md`, `docs/proposal.md` |
| [`/foundation`](foundation/SKILL.md) | Scaffolds the repo, tooling, CI gates, env/secrets, seed data, a reference module | A running, empty codebase |
| [`/design`](design/SKILL.md) | Finalizes the theme, designs every screen as a real built prototype | Finalized theme, `docs/screens.md`, `docs/design.md`, Storybook stories |

## Steady-state loop (recurring, for the life of the project)

| Skill | Does | Produces |
|---|---|---|
| [`/tickets`](tickets/SKILL.md) | Cuts one feature into demoable tickets — runs per-feature, just-in-time, not once for the whole project. First invocation also cuts the App Shell ticket. | `docs/tickets/<feature>/*.md` |
| [`/build`](build/SKILL.md) | Claims one ticket, implements it, self-verifies, opens a PR | A PR per ticket |
| [`/review`](review/SKILL.md) | Independent fresh-eyes review of a ticket's PR before merge | Merged PR or rejection with findings |
| [`/release`](release/SKILL.md) | Gets merged code to production, per `docs/release.md`'s configured tier | Code running in production |

## Entry points into the loop

| Skill | Does | Produces |
|---|---|---|
| [`/fix`](fix/SKILL.md) | Triages a logged bug, routes it through the full pipeline or a direct fix | Bug resolved, `docs/bugs.md` updated |
| [`/add`](add/SKILL.md) | Scopes a new, previously-unplanned feature request against existing docs | New `docs/scope.md` entry, tickets cut |

## One-time migration

| Skill | Does | Produces |
|---|---|---|
| [`/adopt`](adopt/SKILL.md) | Reconciles an existing project's old docs + real code onto the current doc set | Current doc set reflecting real project state |

## Shared references

| File | Used by |
|---|---|
| [`references/ui-rules.md`](references/ui-rules.md) | `/design`, `/review` — checkable UI rules |
| [`references/design-principles.md`](references/design-principles.md) | `/design`, `/review` — the reasoning behind the rules, tooling defaults, page templates |

## Core principles running through every skill

- **Ask, don't guess** — every open question comes with a recommended
  default; Edwinfred approves or overrides, never composes from a blank
  slate.
- **Code and worked examples over prose** — conventions and design rules
  point to real in-repo examples wherever possible, not descriptions.
- **Docs are living, not one-shot** — `scope.md`, `conventions.md`, etc.
  get appended to (`/add`, `/fix`) rather than rewritten from scratch.
- **Escalate only on real decisions** — local implementation judgment
  calls don't need sign-off; anything touching scope, architecture, or
  established convention does.
- **Machine-enforced over remembered** — anything checkable by a linter,
  type checker, or CI gate is enforced that way, not left to an agent's
  memory of a rule.
