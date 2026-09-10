# Modules, Interfaces, Boundaries

The vocabulary for code structure. Used by `/plan`, `/foundation`, `/build`, `/review`, and `/care`.

Use these words exactly. Don't substitute "component", "service", "layer", or "API".

## Module

A folder with a job. `billing/` is a module — it owns everything about invoices.

**Modules are parts of the business, not technical layers.** The test: *could you explain it to the client in their own words, and would they agree it's a distinct part of their business?*

- `billing`, `scheduling`, `clients` — pass
- `validators`, `helpers`, `services`, `utils` — fail. The client has never thought about validators

Most enterprise apps land at **four to eight modules**. Twenty means features are masquerading as modules. Two means they're too coarse.

Module names come from `CONTEXT.md`. If a module needs a name that isn't in the glossary, that's a vocabulary gap — surface it.

## Interface

**What a module lets other modules use.** Its `index.ts`, and nothing else.

```
billing/
├── index.ts        ← THE INTERFACE. The only file importable from outside
├── schema.ts       ← internal
├── queries.ts      ← internal
├── logic.ts        ← internal
├── routes.ts       ← internal
├── ui/             ← internal
└── tests/
```

The interface is more than type signatures. It's everything a caller must know to use the module correctly: what it does, what invariants hold, what errors it can produce, what order things must happen in.

## Boundary

**The rule: cross-module imports go through `index.ts` only.** Never `billing/queries.ts` directly.

Enforced by a lint rule, not by discipline.

**Why it matters.** Without a boundary, one module imports another's internals and the two are welded together — change one and the other silently breaks. Thirty of those across six modules and every change becomes risky. That is exactly "hard to extend without destabilizing existing functionality."

With a boundary, everything inside a module can be rewritten freely. As long as the exports behave the same, nothing outside can break. **That is what maintainable actually means** — not tidiness, but the ability to change one thing without fear.

## Depth

**A lot of behaviour behind a small interface.**

- **Deep** — few exports, much implementation. Callers learn a little and get a lot
- **Shallow** — many exports, little implementation. The interface is nearly as complex as what's inside

**Fewer exports is better.** Four exports means four ways for other modules to become entangled. Twenty means twenty.

When designing an interface, ask:
- Can I reduce the number of exports?
- Can I simplify the parameters?
- Can I hide more inside?

## Seam

**Where tests observe behaviour without reaching inside.** A module's interface is its seam.

Seams are decided in Planning and written into `docs/architecture.md`. **No test is written at an unconfirmed seam.** Prefer existing seams. Prefer the highest seam available. Fewer seams is better — the ideal is one per module.

**The interface is the test surface.** If you have to test *past* the interface to check something, the module is the wrong shape.

## The deletion test

Suspect something is shallow? Imagine deleting it.

- Complexity **vanishes** → it was a pass-through. Delete it
- Complexity **reappears across many callers** → it was earning its keep

## Where UI lives

Four tiers, split by reusability:

| Location | Contains | Knows about the domain? |
|---|---|---|
| `components/ui/` | Primitives — button, input, dialog, table | No |
| `components/patterns/` | Page templates — record table, detail page, form, summary strip, the five states | No |
| `components/layout/` | The app shells — nav, header, content frame. One per shell (e.g. `admin-shell.tsx`, `staff-shell.tsx`) | No |
| `modules/<x>/ui/` | `InvoiceTable`, `InvoiceStatusBadge` | Yes |

**`patterns/` vs `layout/`.** A pattern composes *inside* a page — a table, a form, a strip of figures. A shell wraps the *whole page* — it owns the nav and the header and takes page content as children. Both are domain-blind; the split exists because a ticket reaches for them differently: one shell per screen, any number of patterns per screen.

**A shape used by exactly one destination is not a template.** It stays with the module that uses it, in `modules/<x>/ui/`, even if it looks reusable. Promoting it early is exactly the speculative generality this structure exists to avoid — wait for a second caller.

**If it mentions a domain concept, it lives in the module. If it's generic, it lives in `components/`.**

This keeps `components/` small and stable — which is what makes the design system hold — and keeps a module's UI next to its logic and tests.

When two modules need the same domain component, put it in whichever module owns the concept; the other imports it through the interface. Never duplicate.

**Design-phase output does not move here as-is.** The Design skill prototypes destinations as flat pages under `components/design/<destination>/` with fixture data and dev-only switchers. At Foundation, each destination's *content* becomes a module under `modules/<x>/`; only the shell(s) and the patterns it composed move into `components/layout/` and `components/patterns/` respectively. Fixture data, state/variant switchers and anything gated on `process.env.NODE_ENV === "production"` do not survive the move.

## Designing for testability

**Accept dependencies, don't create them.**

```ts
// Testable
function processOrder(order, paymentGateway) {}

// Hard to test
function processOrder(order) {
  const gateway = new StripeGateway();
}
```

**Return results, don't mutate.**

```ts
// Testable
function calculateDiscount(cart): Discount {}

// Hard to test
function applyDiscount(cart): void { cart.total -= discount; }
```

**Small surface area.** Fewer exports means fewer tests needed and simpler setup.
