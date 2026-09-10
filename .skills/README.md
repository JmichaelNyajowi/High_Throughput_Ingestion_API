# Project Skills Workflow

This directory is a portable, project-local workflow for agent-assisted software delivery. Copy the whole `skills/` directory into a new project, then keep project decisions in `CONTEXT.md` and `docs/` rather than baking them into the skills.

The workflow assumes this flat file layout: invoke `discovery.SKILL.md` as `/discovery`, `plan.SKILL.md` as `/plan`, and so on. Links in this README intentionally point to files that exist in this directory.

## Choose an entry point

| Project state | Start with | Why |
| --- | --- | --- |
| Raw client notes or an unclear product idea | [`/discovery`](discovery.SKILL.md) | Capture the business vocabulary and unanswered questions without deciding the solution. |
| New project with sufficiently clear product material | [`/plan`](plan.SKILL.md) | Decide vocabulary, scope, architecture, conventions, and release expectations. |
| Existing project with shipped code and incomplete or stale documentation | [`/adopt`](adopt.SKILL.md) | Reconcile the documentation with the codebase; code is the operational source of truth. |
| New feature after the initial release plan | [`/add`](add.SKILL.md) | Check scope and architecture fit before ticketing it. |
| Logged defect | [`/fix`](fix.SKILL.md) | Triage it into a proportional, documented delivery path. |
| Structural friction between delivery tranches | [`/care`](care.SKILL.md) | Find and rank one architecture improvement before it compounds. |

Do not use `/adopt` merely because a project has planning documents. It is for projects with real existing code whose documentation must be reconciled.

## New-project flow

1. Run `/discovery` when the input is raw stakeholder material. Skip it when the supplied product material is already adequate.
2. Run `/plan` to produce or update `CONTEXT.md`, `docs/architecture.md`, `docs/scope.md`, `docs/conventions.md`, and `docs/release.md`.
3. Run `/foundation` to make the documented stack runnable, testable, and safe to change.
4. Run `/design` for products with a UI. It turns the agreed product and architecture into screen, component, state, and accessibility decisions.
5. Run `/tickets` for only the next feature. It creates dependency-aware, standalone tickets in `.work/`.
6. Run `/build` for an unblocked ticket. It implements, self-verifies, and prepares a pull request; it does not merge.
7. Run `/review` in a fresh review context. Approval is required before a ticket is merged.
8. Run `/release` after the reviewed merge, following `docs/release.md`.

For backend-only products, skip `/design` and proceed from `/foundation` to `/tickets`.

## Ongoing delivery loop

| Skill | Output / decision |
| --- | --- |
| [`/tickets`](tickets.SKILL.md) | One feature’s `.work/<NN>-<slug>.md` tickets, with blockers and explicit lifecycle decisions. |
| [`/build`](build.SKILL.md) | A self-verified pull request for one ticket. |
| [`/verify`](verify.SKILL.md) | Evidence from a real running system for multi-step, role-sensitive, or UI flows. |
| [`/critique`](critique.SKILL.md) | A rule-backed UI finding report; it never redesigns or fixes directly. |
| [`/review`](review.SKILL.md) | Approved merge or concrete blocking findings. |
| [`/release`](release.SKILL.md) | Deployment or a surfaced deployment blocker. |

`/verify` is recommended after implementation when a ticket has a meaningful interactive flow. `/critique` follows verification for UI tickets. Neither replaces the required `/review` gate.

## Ticket contract

Every ticket uses the format in [`reference.TICKET-FORMAT.md`](reference.TICKET-FORMAT.md): `Type`, `Blocked by`, `Status`, `What this delivers`, `Lifecycle`, `Acceptance criteria`, `Out of scope`, and `Verification`. A record-changing ticket must explicitly decide create, read, update, delete, and undo; “not allowed” is a valid answer, silence is not.

## Shared references

| Reference | Used for |
| --- | --- |
| [`reference.INTERROGATION.md`](reference.INTERROGATION.md) | Clarifying decisions without inventing material requirements. |
| [`reference.MODULES.md`](reference.MODULES.md) | Module, interface, boundary, depth, and seam vocabulary. |
| [`reference.TESTING.md`](reference.TESTING.md) | Test-first and verification discipline. |
| [`reference.TICKET-FORMAT.md`](reference.TICKET-FORMAT.md) | Stable ticket structure. |
| [`references.design-principles.md`](references.design-principles.md) | UI design reasoning; use only for UI products. |
| [`references.ui-rules.md`](references.ui-rules.md) | Checkable UI rules; use only for UI products. |

`reference.UI-RULES.md` is retained only as a compatibility alias; new skills must use `references.ui-rules.md`.

## Make it reusable in another project

1. Copy this whole `skills/` directory into `<project-root>/skills/`.
2. Add a short `AGENTS.md` (or the project’s equivalent agent instruction file) that points to this README and the project’s canonical docs.
3. Run the appropriate entry skill. Skills create missing, important documentation or structural folders when their outputs are determinable; they pause for user direction when the missing item needs a material product, security, or operational decision.
4. Put stack-specific commands, paths, role names, deployment details, and examples in `docs/setup.md`, `docs/conventions.md`, `docs/design.md`, and `docs/release.md`. Do not edit generic skill files to encode one project’s stack.

## Workflow principles

- Ask rather than invent when a decision changes product scope, data ownership, security, or operations.
- Let code and verified runtime behavior overrule stale documentation during adoption.
- Keep docs living and append meaningful history rather than silently rewriting past decisions.
- Enforce repeatable rules in tests, lint, type checks, and CI where the project supports them.
- Keep each ticket small enough to demonstrate, verify, review, and safely reverse as one coherent change.
