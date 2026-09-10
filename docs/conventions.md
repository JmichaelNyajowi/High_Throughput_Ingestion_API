# Engineering Conventions

## Source of truth

- Product behavior: `docs/prd.md`.
- System design: `docs/Architecture.md`, `docs/Techstack.md`, and `docs/database.md`.
- UI decisions: `docs/DESIGN.md`; every UI task must follow it.
- Vocabulary: `CONTEXT.md`.

## Modules and interfaces

- Use the four product modules named in `docs/architecture.md`: Ingestion, Telemetry, Fleet, and History.
- A module owns its domain behavior, tests, and domain UI. Do not create generic business modules such as `utils`, `services`, or `helpers`.
- Cross-module imports use only the target module’s public interface. Do not import a module’s internal query, logic, schema, or UI files directly.
- Keep interfaces small and behavior-oriented. Accept dependencies rather than creating hidden dependency clients inside business logic.
- Tests observe confirmed module seams and HTTP contracts, not private internals or direct database side channels.

## Backend conventions

- Use Go `context` for request/dependency cancellation and explicit HTTP, Redis, and PostgreSQL timeouts.
- Expected client failures return typed HTTP errors; do not use panics for validation, authentication, throttling, or backpressure conditions.
- Parameterize all SQL. Use explicit migrations and `pgx`; do not introduce an ORM.
- Never log secrets, API-key headers, or complete raw telemetry payloads.
- The ingestion route must not synchronously write to Redis or PostgreSQL before responding `202`.

## Frontend conventions

- Follow `docs/DESIGN.md` exactly: named tokens only, one coherent dark graphite theme, dense data views, and no raw hex/arbitrary style values.
- Use accessible primitives, Lucide icons only, real links for navigation, visible focus, and explicit empty/loading/error/degraded states.
- Use TanStack Query for server state and native `fetch` behind typed API functions.
- Dashboard auto-refresh accesses only live Redis-backed endpoints. Manual history runs only after an explicit operator action.

## Testing conventions

- Logic is test-first: one failing behavior test, minimum implementation, passing test, then next behavior.
- Integration tests are the default and use real PostgreSQL/Redis containers. Do not mock this project’s modules or data layer.
- Unit tests are reserved for complex pure rules such as validation ranges, rate limits, threshold boundaries, and rolling-window behavior.
- Permanent browser E2E tests cover only critical flows. Use condition-based waits, stable test IDs, and no fixed sleeps.
- Run format/type checks continuously, touched tests during work, and the full suite before a ticket is marked ready for review.

## Change and release conventions

- Work one planned `.work` ticket at a time. A ticket is done only after independent review.
- Use backward-compatible expand/contract migrations. Never drop/rename a live schema dependency in the same deployment that stops using it.
- Do not create missing feature scope implicitly. Use the add/fix workflow when a request expands the approved scope.
