---
name: critique
description: Audit implemented UI against written project design and accessibility rules, producing only rule-backed findings.
---

# Critique

Run after `/verify` for UI work. Read the ticket, implementation diff, `docs/design.md`, `docs/screens.md`, and `references.ui-rules.md` when applicable. Project rules take precedence over reference material.

Audit design tokens, hierarchy, component reuse, states, responsive behavior, focus management, keyboard interaction, ARIA semantics, contrast, error recovery, and lifecycle actions declared by the ticket. Produce a severity-ranked list where every finding identifies the written rule, affected location, and precise remedy.

Do not offer aesthetic preferences, redesign screens, or change code. Repeated findings should become a proposed convention or design-rule update for the project owner to decide.
