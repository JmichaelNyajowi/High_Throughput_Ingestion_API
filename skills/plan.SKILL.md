---
name: plan
description: Interrogate the user to turn discovery into architectural decisions — vocabulary, domain model, module boundaries, seams, stack, scope, and platform decisions. Writes CONTEXT.md, architecture.md, and scope.md. Also runs standalone for vocabulary work.
disable-model-invocation: true
---

# Plan

Convert the discovery record into **decisions solid enough to build on**. Not documents — decisions. A few of them produce documents as a side effect.

Read `<skills>/reference/INTERROGATION.md` and follow it throughout. Read `<skills>/reference/MODULES.md` for the boundary vocabulary.

**Resolving `<skills>/`.** It is the directory holding the skill folders — `~/.claude/skills/` for a global install, or `<project>/.claude/skills/` for a per-project one. **It is not inside this skill's own folder, and it is not in the project root.** Check the global path first, then the project-local one.

**If a reference file cannot be found, stop and tell the user.** Do not proceed from memory — these files hold the discipline the skill depends on, and running without them silently produces work that looks right and isn't.

## Modes

Infer from the user's arguments; ask if unclear.

- **Full** (default) — Group A then Group B. Once per project
- **Scoped** — a subset, for a later feature needing new boundaries or a new domain concept
- **Vocabulary** — Group A step 1 only. Sharpen a term, resolve an overloaded word, add a concept

## Before starting

1. Read `docs/discovery.md` if it exists — it holds the client's own vocabulary and the questions they couldn't answer. Those questions are your agenda
2. Read `CONTEXT.md`, `docs/architecture.md`, `docs/scope.md`, and `docs/adr/` if they exist
3. Explore the codebase if there is one

**This runs in one unbroken session.** If you approach the context limit before finishing, stop and write a handoff rather than continuing degraded — the decisions build on each other and a degraded second half corrupts the first.

## Group A — design decisions

**Sequential. Never batched, even if asked** — each decision depends on the last. Explain that once if the user asks to batch.

### 1. Vocabulary

Start from the raw client terms in `docs/discovery.md`.

- Challenge fuzzy terms: *"You said 'account' — do you mean the customer, or the login? Those are different things."*
- Challenge overloaded ones: a word doing three jobs needs to become three words
- Stress-test with concrete scenarios that probe the edges between concepts
- Cross-reference against the code if one exists, and surface contradictions

**Also look for the terms that are perfectly clear but appear everywhere.** Fuzzy terms announce themselves; pervasive ones don't, and they are the more architectural of the two. A word that qualifies most other terms in the glossary — *location*, *branch*, *tenant*, *season*, *shift* — is rarely just an attribute. It usually cuts through the whole domain, and deciding late is expensive.

The test: **take a candidate term and ask whether each other term is meaningful without it.** If "stock" means nothing until you say *which location's* stock, and neither does a sale, a transfer, or a handover, then location is structural — it belongs in the domain model, and probably in the permissions model too. Name it and raise it explicitly rather than letting it stay implicit.

**Write each settled term to `CONTEXT.md` immediately** — not batched at the end. Use the client's word where it works; invent one only where theirs is genuinely ambiguous.

`CONTEXT.md` is a **glossary and nothing else**. No implementation detail, no decisions, no scratch notes.

### 2. Domain model

The real entities and their relationships, named in the settled vocabulary.

Probe: Can this exist without that? Can there be two? What happens when one is deleted? What states can it be in, and which transitions are legal?

**Then check for a cutting dimension** — anything flagged as pervasive in step 1. If most entities are scoped by it, say so plainly and settle three things before moving on:

- **Is it an entity or an attribute?** If it has its own behaviour and its own records, it's an entity
- **Does it scope permissions?** *Can a manager at one location see another's figures?* Retrofitting this is one of the most painful changes possible
- **Do things move between instances of it?** Movement in **both directions** is a different design from a hub and spokes, and it changes where the logic lives

A cutting dimension that goes unnamed here reappears as a column bolted onto twelve tables and a permissions model that can't express what the client meant.

### 3. Module boundaries and seams

Divide the system into modules.

**The test for a module:** *could you explain it to the client in their own words, and would they agree it's a distinct part of their business?* Four to eight is typical.

Then name the **seams** — where tests will observe behaviour. Prefer few. Prefer the highest available. The ideal is one per module.

Confirm the seam list explicitly with the user; it gets written into `docs/architecture.md` and `/build` will refuse to test anywhere else.

### 4. Stack

**Only now.** Chosen to fit the domain, never the reverse. Choosing the stack first is how you fit the problem to the framework.

### 5. Scope

The **v1 destination** in one or two sentences. Then an explicit **out-of-scope list**.

The out-of-scope list is worth more than the in-scope one — it's what stops creep from both the client and future agents.

## Group B — platform decisions

**Independent of each other, so these can be batched.** Offer that.

**"Nothing special here" is a valid answer** and gets recorded as such. The value is in having asked — an unasked question about permissions is how you discover in month four that every user can see every client's data.

1. **Identity and access** — who logs in, how, what they can do. Roles, permissions, multi-tenancy, SSO against an existing directory. Retrofitting a permissions model is among the most painful things possible

2. **Data lifecycle** — migrations, backups, retention. **Ask explicitly: are records ever really deleted, or soft-deleted with an audit trail?** Enterprise clients almost always want soft delete and history, and it's architectural. This answer sets the default that every ticket's lifecycle section inherits

3. **Integrations** — what other systems this talks to, in which direction. Each is a seam

4. **Environments and deployment shape** — where it runs, how code gets there. **If the stack is split across hosts** (e.g. frontend on Vercel, backend on a droplet), resolve preview deployments here: preview frontend against production backend only works for frontend-only changes; a shared staging backend is the cheap solo answer; a host with per-branch environments is the real fix; colocating removes the problem

5. **Observability** — how the user learns something broke before the client tells them

6. **Non-functionals** — realistic scale, performance expectations, compliance constraints. Usually "small, nothing special" — and recording that explicitly **licenses not over-engineering**

## Outputs

Write these only after the user confirms shared understanding.

**`CONTEXT.md`** — written inline during Group A step 1. Glossary only.

**`docs/architecture.md`** — module boundaries, the confirmed seam list, stack, all six Group B decisions, and **why** for each. Reasons don't go stale even when code moves.

**`docs/scope.md`** — the v1 destination and the out-of-scope list.

**`docs/adr/NNNN-<slug>.md`** — sparingly.

### ADR bar — all three must hold

1. **Hard to reverse** — changing your mind later has meaningful cost
2. **Surprising without context** — a future reader will wonder why
3. **The result of a real trade-off** — there were genuine alternatives

If any is missing, skip it. Expect very few; ten on a project is a lot.

Each ADR is its own numbered file and is **immutable once written**. A reversal is a new ADR superseding the old, never an edit — the record of what you thought at the time is the value.

## Never

- **Never design module interfaces.** Boundaries and seams only. What a module exposes is decided when that module is built, because you can only tell what belongs on the front door once you know what's inside
- **Never write file paths or code snippets** into the documents — they go stale. Exception: a snippet encoding a decision more precisely than prose can (a state machine, a schema shape)
- **Never produce a full up-front work breakdown.** That's `/tickets`, one tranche at a time
- **Never batch Group A**
- **Never write implementation detail into `CONTEXT.md`**

## Done when

Nothing in Group A or Group B is undecided and the documents exist.

### Where to hand off

**Depends on whether the project exists yet.**

| Situation | Next | Why |
|---|---|---|
| **First run — no codebase** | **`/design`** | Then `/foundation`, then `/tickets`. A ticket cuts a vertical slice through every layer, and those layers don't exist yet |
| **Scoped run on a live project** | **`/tickets`** | The layers exist. New work is just slices through them |

**Never send a greenfield project straight to `/tickets`.** Tickets cut through schema, logic, API, UI and tests, and none of them exist after Planning. Worse, a UI ticket written before the design system exists invites the agent to invent one — which is the vacuum failure the Design phase exists to prevent.

The full lifecycle order is **`/plan` → `/design` → `/foundation` → `/tickets`**, then the Build loop forever.
