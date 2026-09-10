# Development Setup

## Prerequisites

- Go 1.26 or newer.
- Node.js 20 LTS and npm 10 or newer.
- Docker Engine with the Compose plugin for PostgreSQL, Redis, Caddy, and integration tests.
- PostgreSQL client tools (`psql`) when applying migrations from the host.

Docker is required for the full local topology and Testcontainers tests. It is not available in the current development environment, so container-backed checks must run in CI or on a Docker-capable host.

## First run

1. Copy `.env.example` to `.env` and replace every placeholder only in your local secret store or untracked environment file.
2. Install frontend dependencies: `npm --prefix frontend ci`.
3. Start dependencies: `docker compose -f deploy/compose.yaml --env-file .env up -d postgres redis`.
4. Apply migrations: `./scripts/migrate-up.sh`.
5. Run the API: `go run ./api/cmd/api`.
6. Run the frontend: `npm --prefix frontend run dev`.

The bootstrap API exposes only `GET /healthz` and `GET /readyz`. Product endpoints land through approved tickets.

## Quality commands

| Check | Command |
| --- | --- |
| Go format | `(cd api && gofmt -w .)` |
| Go static checks | `(cd api && go vet ./...)` |
| Go tests and race detector | `(cd api && go test -race ./...)` |
| Frontend type check | `npm --prefix frontend run typecheck` |
| Frontend unit tests | `npm --prefix frontend run test` |
| Frontend production build | `npm --prefix frontend run build` |
| Compose validation | `docker compose -f deploy/compose.yaml --env-file .env config` |

Run container-backed integration tests only on a Docker-capable host. Browser tests and k6 benchmarks are added with the relevant feature tickets.
