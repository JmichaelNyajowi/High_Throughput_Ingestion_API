---
name: discovery
description: Prepare questions for a client meeting, or debrief the user afterwards and write a dated entry into docs/discovery.md. Captures the client's business in their own words without deciding scope.
disable-model-invocation: true
---

# Discovery

Capture what the client's business actually is, **in their words**, before any interpretation.

Read `<skills>/reference/INTERROGATION.md`.

**Resolving `<skills>/`.** It is the directory holding the skill folders — `~/.claude/skills/` for a global install, or `<project>/.claude/skills/` for a per-project one. **It is not inside this skill's own folder, and it is not in the project root.** Check the global path first, then the project-local one.

**If a reference file cannot be found, stop and tell the user.** Do not proceed from memory — these files hold the discipline the skill depends on, and running without them silently produces work that looks right and isn't.

## Two modes

Infer from the user's arguments; ask if unclear.

- **Prep** — before the meeting
- **Debrief** — immediately after

---

## Mode A — Prep

**Input:** the business type, anything already known, and `docs/discovery.md` if it exists.

1. **Read `docs/discovery.md` if present.** Don't generate questions already answered in an earlier entry, and do generate questions for anything still in the open-questions section
2. **Generate a question list**, grouped:
   - How the business works today — the actual steps, who does what
   - Where it hurts — what's slow, what breaks, what gets done twice
   - What they want the software to do
   - Who will use it, and how their jobs differ
   - What systems they already have, and what would need to talk to what
   - What happens when things go wrong today — the exceptions, the edge cases
   - What they'd consider a failure
3. **Flag the highest-value questions**, so a short meeting covers the right ground

**Output:** a question list for the user to take into the meeting. **Not written to the repo.**

---

## Mode B — Debrief

Run this **immediately after the meeting**, while recall is fresh.

**The goal here is extraction, not decision.** Probe for what the user noticed, what surprised them, what the client couldn't answer.

1. **Interrogate the user.** One question at a time by default; batch on request. Push on:
   - What did they say, as close to their words as you can recall?
   - What surprised you?
   - What did they say twice, or get animated about?
   - What did they *not* know the answer to?
   - What did you observe that they didn't say — a workaround, a spreadsheet, an obvious frustration?

2. **Push especially hard on two sections**, because they're the ones normally skipped and most missed later:
   - **Their vocabulary** — the exact words they use for things
   - **Open questions they couldn't answer** — these become Planning's agenda

3. **Write a dated entry**, appended to `docs/discovery.md`.

### Entry format

```markdown
## <date> — <who was there, what the context was>

### How the business works today
<the actual steps, in order, who does what>

### Pain points (their words)
<quote them where you can>

### What they asked for
<what they said they want, before any interpretation>

### What I observed but they didn't say
<workarounds, frustrations, the spreadsheet doing the real work>

### Their vocabulary

| Their term | What they seem to mean | Notes |
| --- | --- | --- |

### Open questions they couldn't answer
<these become Planning's agenda>
```

**Output:** `docs/discovery.md`, appended.

---

## Never

- **Never revise an earlier entry.** Append only, with a date. Discovery isn't one meeting — clients keep telling you things for years, and an old entry is history, not an error
- **Never decide scope.** Discovery captures; Planning decides. Deciding while still absorbing the business commits you to the first plausible shape
- **Never propose a solution**
- **Never translate their vocabulary into technical terms.** Record it raw. Sharpening happens in `/plan`, and the raw version is what makes that possible

## Done when

The entry is written and the open-questions section is populated.

If this is the first entry, tell the user to run **`/plan`**.
