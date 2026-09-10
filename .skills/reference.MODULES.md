# Modules, Interfaces, Boundaries

Use this vocabulary in planning, implementation, review, and architecture care.

## Module

A module is a cohesive unit that owns one business capability or system responsibility. Its name should come from `CONTEXT.md` where possible; it is not a generic technical layer such as `utils`, `helpers`, or `services`.

## Interface

The deliberately small public contract a module offers to other modules: supported operations, inputs, outputs, invariants, and errors. In this repository, Go package APIs, TypeScript module `index.ts` exports, and `api/openapi.yaml` are the confirmed forms; `docs/architecture.md` identifies their ownership and seams.

## Boundary

The rule that modules use another module only through its interface, never its internals. Enforce it through the strongest practical tool for the project: compiler visibility, dependency rules, import linting, package structure, or review checks.

## Depth

Depth means substantial useful behavior behind a small interface. Prefer a few well-defined operations over leaking internal storage, helpers, and orchestration to callers.

## Seam

A stable point where tests observe a module’s behavior without reaching into implementation details. Prefer the highest useful public seam. The architecture must identify seams for code with important business behavior.

## Deletion test

For a suspicious abstraction, imagine deleting it. If complexity disappears, it is likely a pass-through. If the same complexity reappears across callers, it is earning its place.

## UI ownership

For UI products, distinguish domain-blind primitives and layout/patterns from domain-aware components. The project’s design and conventions documents define the actual folders and approved component system. Promote a local shape to a shared component only after a real second use case confirms its contract.
