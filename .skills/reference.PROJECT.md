# Telemetry Project Map

Use this reference before selecting project documentation or a verification command. Do not substitute a similarly named path or invent a missing project artifact.

## Canonical sources

| Concern | Required source |
| --- | --- |
| Product requirements and API behavior | `docs/prd.md` |
| Product vocabulary | `CONTEXT.md` |
| Approved scope and delivery order | `docs/scope.md`, `docs/roadmap.md` |
| Delivery-facing module boundaries | `docs/architecture.md` |
| Full system, HTTP, data, authentication, and deployment design | `docs/Architecture.md` |
| Stack and database choices | `docs/Techstack.md`, `docs/database.md` |
| Binding UI rules and tokens | `docs/DESIGN.md` |
| Screen inventory and screen behavior | `docs/screens.md` |
| Code and test conventions | `docs/conventions.md` |
| Local commands and environment constraints | `docs/setup.md` |
| Merge and release process | `docs/release.md` |
| Ticket workflow | `.work/README.md`, then the selected `.work/<NN>-<slug>.md` |
| Public HTTP contract | `api/openapi.yaml` |
| Edge access policy | `deploy/Caddyfile`, `deploy/compose.yaml` |

`docs/architecture.md` and `docs/Architecture.md` have distinct roles and both are required when a change affects modules, HTTP, authentication, data flow, deployment, or observability. The filename `docs/DESIGN.md` is uppercase and binding; do not use or create `docs/design.md`.

## Repository layout and boundaries

- Frontend: `frontend/`; React/TypeScript/Vite source is `frontend/src/`; component stories are `frontend/src/**/*.stories.tsx`.
- Backend: `api/`; Go module root is `api/`; the OpenAPI contract lives at `api/openapi.yaml`.
- Deployment: `deploy/`; Caddy is the external HTTP boundary. Any HTTP, authentication, endpoint, metrics, health, or network-access change must inspect it.
- Tickets: work only from an unblocked `.work/` ticket. Do not use `tickets/` as implementation authority.
- Project instructions: `AGENTS.md` and `.skills/README.md` are authoritative workflow entry points.

## Verification commands

Run only checks relevant to the ticket, plus all documented quality gates for the touched application.

| Area | Commands |
| --- | --- |
| Frontend | In `frontend/`: `npm run lint`, `npm run typecheck`, `npm run test`, `npm run build`; run `npm run storybook:build` for UI/stories. |
| API | In `api/`: `GOCACHE=/tmp/telemetry-go-cache go test -race ./...`, `GOCACHE=/tmp/telemetry-go-cache go vet ./...`, and `test -z "$(gofmt -l .)"`. |
| Contract | The API test suite loads and validates `api/openapi.yaml`; add behavior assertions for material contract rules. |
| Deployment | When Docker and a safe environment file are available: `docker compose -f deploy/compose.yaml --env-file .env config`. Docker is unavailable in this development environment, so surface that limitation rather than claiming runtime Compose verification. |
| Repository | From the repository root: `git diff --check`. Run `npm audit --json` only when dependency changes or the ticket requires an audit. |

## Mandatory build preflight

Before setting a ticket to `in-review`, record a build-gate trace in the ticket. For every acceptance criterion and every applicable binding rule, name:

1. the source document and rule;
2. the implementation location; and
3. repeatable proof (test, static check, or runtime evidence).

Apply these change-sensitive checks:

- **UI:** inspect `docs/DESIGN.md` and `docs/screens.md`; test declared routes including malformed/unknown routes, keyboard focus order, named focusable regions, loading/error/empty/degraded states, and desktop/tablet/mobile rules. Use browser evidence when available; otherwise state the limitation.
- **HTTP/API/security:** inspect `api/openapi.yaml`, `deploy/Caddyfile`, and `deploy/compose.yaml`; test status, headers, content type, error shape, authorization, and negative paths. Confirm the deployed edge policy matches every contract claim about internal access, Basic Auth, TLS, or Prometheus-only access.
- **Data/concurrency:** inspect `docs/database.md` and the full architecture; test lifecycle, bounds, retry/failure behavior, and module seams specified by `docs/architecture.md`.

Do not mark `in-review` with an unresolved preflight discrepancy. Add a regression test for each corrected review finding when the behavior is testable.
