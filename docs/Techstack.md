# Technology Stack Recommendation

## Decision summary

Build the MVP as a **single Go service** with a separate **React + TypeScript SPA**, backed by **PostgreSQL** and **Redis**, deployed with **Docker Compose** on the specified single Ubuntu host. Put **Caddy** in front as the only internet-facing service.

This is the smallest stack that meets the PRD's fast admission path, bounded asynchronous processing, Redis-only live reads, PostgreSQL durability, and portfolio-quality operational visibility. Do not add Kafka, Kubernetes, a BaaS, an ORM, or a separate backend-for-frontend service for this MVP.

```text
Devices / load generator ──HTTPS──> Caddy ──> Go API ──> bounded queue + workers
                                           │          ├──> Redis (live aggregates)
React SPA <────────HTTPS──────────── Caddy │          └──> PostgreSQL (raw events)
                                           └──> /metrics ──> Prometheus ──> Grafana
```

## Requirements that drive the recommendation

- The API acknowledges only after in-memory enqueue; it must therefore minimize synchronous work and strictly bound queues and retries.
- The target is 2,000 requests/sec and 20,000 events/sec from at least 2,000 device/API-key simulators on one 4-vCPU, 8-GB host.
- Live UI reads must come from Redis, while PostgreSQL is the raw-event system of record.
- Redis loss must leave ingestion and PostgreSQL persistence available, with live aggregates paused rather than falling back to database polling.
- The MVP uses static per-device credentials, an in-memory per-device rate limiter, and no external broker.

## Recommended stack

| Area | Recommendation | Why this is the best MVP choice |
|---|---|---|
| Frontend | React + TypeScript + Vite | Matches the PRD, produces a small static SPA, and gives type-safe dashboard code without server-side rendering complexity. |
| Client state | TanStack Query; native `fetch` | Polling, cache freshness, retries, and error states are first-class needs. `fetch` avoids an unnecessary HTTP client dependency. |
| Charts | Recharts | Appropriate for a small number of responsive time-series, gauge-adjacent, and summary visuals; avoid building chart primitives. |
| Backend | Go with `net/http`, `context`, and `log/slog` | Go is well suited to the bounded worker/queue model; its standard HTTP stack has native cancellation and timeout handling. |
| Routing | `go-chi/chi` | A thin, `net/http`-compatible router with composable middleware, rather than a large framework. |
| PostgreSQL access | `pgx` native pool | Strong Go/PostgreSQL integration and efficient batch APIs, while retaining explicit SQL and transactions. |
| Database | PostgreSQL | Durable relational source of truth; supports transactions, constraints, indexes, and conflict-safe idempotent writes. |
| Hot state | Redis + `go-redis/v9` | Fast derived aggregate reads and native TTL support; it remains rebuildable cache, not a durable queue. |
| Authentication | Static, high-entropy device API keys in `X-API-Key`; Caddy Basic Auth plus network restriction for the internal dashboard | Fulfills the MVP without adding an identity provider. Device and operator access are kept separate. |
| API contract | REST/JSON over HTTPS, documented in OpenAPI 3.1 | Fits device batch submission and dashboard reads; one portable contract can drive docs, validation, and test fixtures. |
| Tests | Go `testing`, Testcontainers for Go, Vitest/React Testing Library, Playwright, and k6 | Covers deterministic units, real dependency integration, UI behavior, browser flow, and the throughput requirement. |
| Observability | Prometheus + Grafana; structured JSON logs to stdout | Matches the PRD's metrics requirements with a proven pull model and low operational overhead. |
| Deployment | Docker Compose on Ubuntu, Caddy reverse proxy, systemd, CI-built images | Precisely matches the single-host MVP constraint and avoids premature orchestration. |

Pin image and dependency versions for repeatability; choose the then-current supported stable release within each major family and upgrade intentionally through CI. Do not use floating production tags such as `latest`.

## Frontend

### React, TypeScript, and Vite

Use **React with TypeScript**, bundled by **Vite**, as a static SPA. React is already named in the project guide, and TypeScript reduces mistakes in API-response, threshold, and stale/degraded-state rendering. React's official documentation supports TypeScript directly ([React TypeScript guide](https://react.dev/learn/typescript)).

Vite is preferable to a full-stack React framework because the dashboard is an internal, authenticated operational UI with no SEO, server-rendering, or server-action requirement. Static assets can be served through Caddy; this removes a Node.js server from production.

### State, polling, and charts

- Use **TanStack Query** for dashboard API polling, caching, loading/error states, stale markers, and retry control.
- Use the browser's native **`fetch`** API in one typed client module. Do not add Axios.
- Use **Recharts** for the five-minute line charts and simple gauge/status presentation.
- Poll at **2–5 seconds** initially. The exact interval should be benchmarked against the Redis p95 target. Do not introduce WebSockets or Server-Sent Events in MVP; polling is simpler and sufficient for a five-minute aggregate view.

### Frontend security

- Serve UI and API from the same origin through Caddy; avoid broad CORS entirely.
- Do not place device API keys in the browser bundle, local storage, or client-side configuration.
- Protect the internal dashboard at the reverse proxy with Caddy Basic Auth over TLS and restrict network access to the operator VPN, private subnet, or an IP allowlist.
- This is appropriate only for the internal MVP. Replace it with organization SSO/OIDC before broader use; Basic Auth has no RBAC or lifecycle management.

## Backend

### Go service

Use one **Go** binary with an explicit package structure for HTTP handlers, admission controls, workers, aggregate cache, persistence, and configuration. `net/http` and `context` should remain the foundation: request contexts carry cancellation and deadlines through the call chain ([Go `context`](https://pkg.go.dev/context), [Go `net/http`](https://pkg.go.dev/net/http)).

Use **`go-chi/chi`** only for routing and middleware. It is `net/http` compatible and deliberately small ([chi documentation](https://github.com/go-chi/chi)). Do not use Gin, Echo, Fiber, or a full enterprise framework; none is necessary for the API surface in the PRD.

Use these supporting libraries:

- **`pgx/v5`** with `pgxpool` for PostgreSQL connections and explicit transaction/batch operations.
- **`go-redis/v9`** for Redis connectivity; it is the official Go Redis client and provides connection pooling ([go-redis](https://github.com/redis/go-redis)).
- **`golang.org/x/time/rate`** for the local token bucket rather than a custom algorithm.
- **`log/slog`** from the Go standard library for structured JSON logging.
- **`prometheus/client_golang`** for counters, gauges, and histograms at `/metrics`; Prometheus documents the Go client and HTTP exposition pattern ([Prometheus Go instrumentation](https://prometheus.io/docs/guides/go-application/)).

### Backend design rules

- Set explicit server read-header, read, write, and idle timeouts. Set shorter dependency-specific timeouts for Redis and PostgreSQL.
- Use `http.MaxBytesReader` before JSON decoding to enforce the 1 MB limit.
- Decode and validate before enqueue. Reject entire invalid batches; never partially enqueue.
- Implement a fixed worker pool and bounded channels. A full admission queue returns `503`; it must never create unbounded goroutines or memory use.
- Keep the rate limiter local, as required. It is deliberately not valid for multi-instance global enforcement.
- Treat Redis as a derived cache. If it fails, persist-only mode continues and the dashboard shows its required degraded banner.
- Do not expose `/metrics`, PostgreSQL, or Redis directly to the public internet.

## Database and caching

### PostgreSQL

Use **PostgreSQL** as the raw-event system of record. It is the right fit for the PRD because the product needs transactional batch writes, foreign keys, range/schema constraints, indexes, manual historical queries, and `UNIQUE(device_id, event_id)` idempotency. PostgreSQL supports multi-column unique constraints ([PostgreSQL constraints documentation](https://www.postgresql.org/docs/current/ddl-constraints.html)).

Use explicit SQL migrations and **`pgx`**, not an ORM. An ORM adds abstraction without helping the high-throughput batch path and makes `ON CONFLICT DO NOTHING`, transaction boundaries, indexes, and query plans harder to reason about.

Recommended persistence approach:

- Flush up to 1,000 events or every two seconds in a transaction, as stated in the PRD.
- Use parameterized multi-row `INSERT ... ON CONFLICT (device_id, event_id) DO NOTHING` through `pgx` batching.
- Do not use `COPY` for the core idempotent path without a staging-table design: `COPY` alone cannot express the PRD's conflict behavior.
- Create the required unique constraint and `(device_id, measurement_type, event_timestamp DESC)` index from the PRD.
- Use `timestamptz` for all timestamps and store both producer event time and server received time.

**Capacity warning:** 20,000 raw events/sec equals 1.728 billion events/day. A single 4-vCPU, 8-GB host is appropriate for a controlled portfolio benchmark, not indefinite retention at that volume. Retain demo data for a short, defined period and document actual benchmark duration. Before sustained real-fleet use, add time partitioning, retention/deletion jobs, separate database capacity, and a durable queue.

### Redis

Use a standalone **Redis** instance exclusively for live aggregate state. The PRD's 15-minute TTL maps directly to Redis key expiration; expired keys are automatically removed ([Redis key expiration](https://redis.io/docs/latest/develop/using-commands/keyspace/)).

- Store a compact hash or serialized aggregate per `device_id:measurement_type` key.
- Refresh the 15-minute TTL atomically whenever a valid aggregate is updated.
- Configure a memory limit and an eviction policy deliberately. Since aggregates are rebuildable, eviction is acceptable only if the UI treats a cache miss as stale/unknown, never as healthy.
- Require Redis authentication, bind it to the internal Docker network only, and use TLS if Redis ever leaves the single host.
- Do not use Redis as a durable queue, primary database, or global rate-limit store in MVP.

## Authentication and authorization

### Device ingestion: static API keys

Use the PRD's static per-device/client API keys for MVP. Send them in `X-API-Key`, over HTTPS only.

- Generate keys using a cryptographically secure random generator with at least 32 bytes of entropy.
- Store only a keyed hash (HMAC-SHA-256 with a deployment secret/pepper) in environment-backed configuration; compare in constant time.
- Map each key to exactly one device or permitted device set before admitting its telemetry.
- Rate-limit by the authenticated key/device, using the mandated 10 event tokens/sec and 100-token burst.
- Do not log raw keys, request authorization headers, or complete telemetry batches.
- Store secrets outside source control, set restrictive file permissions, and inject them through the deployment environment or a secrets manager.

API keys meet the limited MVP need, but they are not a complete authorization system. OWASP notes that API keys should be required for protected endpoints and paired with rate limits, while cautioning against relying on them alone for sensitive or high-value resources ([OWASP REST security guidance](https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html)).

### Operator dashboard: reverse-proxy access control

The dashboard needs a different protection boundary from devices. Use Caddy Basic Auth over TLS plus VPN/private-network or IP-allowlist access for the MVP. This keeps the product simple and stops an unauthenticated public dashboard from exposing fleet data.

Do not implement dashboard user accounts in the Go service. When the product requires multiple teams, audit trails, or role-based access, replace proxy Basic Auth with the organization’s OIDC provider and enforce role claims at the proxy/API boundary.

### Transport security

Use HTTPS everywhere externally. Caddy should redirect HTTP to HTTPS, obtain/renew certificates, set security headers, and proxy only required paths. TLS provides confidentiality, integrity, and server authentication; OWASP recommends encrypted TLS for authenticated web-service communication ([OWASP TLS guidance](https://cheatsheetseries.owasp.org/cheatsheets/Transport_Layer_Security_Cheat_Sheet.html)).

## APIs and contract

Use versioned REST/JSON endpoints under `/v1`. REST is appropriate because devices submit independent bounded batches and the dashboard periodically reads snapshots; neither requires a streaming protocol in MVP.

Maintain an **OpenAPI 3.1** document as the API contract. OpenAPI is a language-agnostic interface description that supports discovery, documentation, generated clients, and testing ([OpenAPI Specification](https://spec.openapis.org/oas/latest.html)). Use 3.1 rather than the newest spec minor so Go and TypeScript tooling remains broadly compatible.

Recommended MVP endpoints:

| Endpoint | Consumer | Purpose |
|---|---|---|
| `POST /v1/telemetry/batches` | Device | Submit 1–500 telemetry events. |
| `GET /v1/live/fleet` | Dashboard | Redis-derived fleet summary and degraded status. |
| `GET /v1/live/devices/{deviceId}` | Dashboard | Redis-derived current values and five-minute aggregates. |
| `GET /v1/history/devices/{deviceId}` | Dashboard, manual action only | Explicit, bounded PostgreSQL historical query. |
| `GET /healthz` | Platform | Liveness only. |
| `GET /readyz` | Platform | Readiness and dependency/admission status. |
| `GET /metrics` | Prometheus only | Prometheus metrics; protect from public access. |

Contract requirements:

- Document `400`, `401`, `413`, `415`, `429`, `503`, and `202` response schemas and `Retry-After` where applicable.
- Define API-key security in the OpenAPI document but never place a real key in examples or generated web clients.
- Return a request/correlation ID on every response.
- Keep CORS disabled by serving UI and API from one origin. If a development origin needs access, allow only its exact origin, methods, and headers.
- Limit manual historical query time ranges and result counts; never expose an unbounded raw-event endpoint.

## Testing strategy

### Backend tests

Use Go's built-in **`testing`** package for unit, table-driven, fuzz, and race-detector tests. It keeps core tests fast and avoids a testing framework dependency.

Unit-test:

- payload shape, empty/oversized batch, value ranges, timestamp boundaries, and whole-batch rejection;
- token bucket burst, refill, concurrent access, and event-based cost;
- threshold boundary values and stale-device calculation;
- late-event exclusion from Redis aggregation;
- queue-full, retry-backoff, and shutdown behavior.

Use **Testcontainers for Go** for PostgreSQL and Redis integration tests. It runs against real service behavior rather than mocks and is purpose-built for Go integration testing ([Testcontainers for Go guide](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/)). Verify transactions, unique-conflict behavior, retry behavior, Redis TTLs, and cache-rebuild correctness.

Run `go test -race ./...` in CI. Add `go vet` and a pinned Go linter configuration as quality gates.

### Frontend tests

- **Vitest + React Testing Library** for status mapping, degraded banners, stale displays, and API error rendering.
- **Playwright** for browser-level checks of the fleet page, device detail, and degraded mode. Playwright supports Chromium, WebKit, and Firefox ([Playwright documentation](https://playwright.dev/docs/intro)); run Chromium in MVP CI to limit execution time.

### Load and failure testing

Use **k6** for protocol-level telemetry load tests. It is suited to direct API testing and can use requests-per-second scenarios, checks, and thresholds ([k6 API load testing](https://grafana.com/docs/k6/latest/testing-guides/api-load-testing/)).

Required k6 scenarios:

1. Smoke: valid request plus assertions for `202` and validation errors.
2. Baseline: 2,000 distinct API keys/devices at 10 events/sec each, totaling 2,000 requests/sec and 20,000 events/sec.
3. Rate limit: one key exceeding 10 events/sec and receiving `429`.
4. Backpressure: fill the bounded queue/retry buffer and verify bounded memory plus `503` admission rejection.
5. Dependency failure: Redis unavailable (persist-only mode) and PostgreSQL unavailable (bounded retry/backpressure).
6. Soak: sustained representative load long enough to identify memory growth, queue drift, and persistence degradation.

Run substantial load tests from a separate host/container from the system under test; otherwise the load generator can distort the 4-vCPU benchmark. k6 notes that generator CPU saturation can throttle a test and inflate observed response time ([k6 large-test guidance](https://grafana.com/docs/k6/latest/testing-guides/running-large-tests/)).

## Deployment and operations

### MVP topology

Use a single Ubuntu server running Docker Compose services:

```text
caddy             public: 80/443 only
api               internal network; health/readiness and metrics exposed internally
frontend          static files served by Caddy (no long-running Node process)
postgres          internal network; named volume; not publicly published
redis             internal network; named volume if persistence is enabled; not publicly published
prometheus        internal network or private operator network
grafana           private operator network
```

Docker Compose is the right operational boundary for this portfolio MVP because the PRD explicitly targets a single Dockerized Ubuntu host. It defines multi-container applications in one configuration and keeps local, CI, and deployed topology consistent ([Docker Compose documentation](https://docs.docker.com/compose/)).

### Secure deployment baseline

- Use multi-stage Docker builds and non-root runtime users for API and frontend images.
- Pin base images by digest, scan images in CI, and rebuild regularly for security updates.
- Publish only Caddy's ports `80` and `443`; Docker networks isolate API, database, cache, and monitoring services.
- Use named PostgreSQL volumes and test restore procedures. Redis is rebuildable, but PostgreSQL backups are mandatory.
- Encrypt in transit at the edge with Caddy. Keep database and Redis traffic on the private Docker network.
- Use a host firewall that allows only required HTTPS and restricted administration access.
- Run Compose under a `systemd` unit with restart policy and health checks.
- Inject secrets at deploy time; never bake them into images or commit `.env` files.

### CI/CD

Use GitHub Actions if the repository is hosted on GitHub; otherwise reproduce the same pipeline in the chosen CI provider.

1. Run formatting, static analysis, Go unit tests, race tests, and frontend unit tests on every pull request.
2. Run Testcontainers integration tests and Playwright browser tests.
3. Build immutable, versioned images and scan them.
4. Deploy the approved image tag to a staging host with Compose.
5. Run smoke tests automatically; run the full k6 benchmark manually or on a scheduled dedicated runner because it needs isolated capacity.
6. Promote the same immutable image to the demo environment; do not rebuild during promotion.

## Explicitly rejected for MVP

| Technology/approach | Why not now | Reconsider when |
|---|---|---|
| Kafka, RabbitMQ, or managed queues | Violates the deliberate in-memory queue constraint and adds operational overhead. | Data loss after acknowledgement becomes unacceptable or a multi-instance service is required. |
| Kubernetes | Too much operational surface area for one host and five-to-seven containers. | Multiple services/hosts need autoscaling, self-healing, or standardized platform operations. |
| TimescaleDB | PostgreSQL alone is sufficient for short-retention raw events and manually invoked history. | Long-term, high-volume time-series retention and query needs become a real product requirement. |
| ORM | Hides the explicit SQL and batch/idempotency behavior that are central to this product. | Domain complexity becomes much larger and measured developer velocity loss justifies it. |
| GraphQL | Adds a flexible query layer that is unnecessary and risky for high-volume telemetry ingestion. | Many independently evolving UI consumers need strongly controlled flexible reads. |
| WebSockets/SSE | Adds connection lifecycle and fan-out complexity for a dashboard that can safely poll snapshots. | Operators demonstrably need sub-second push updates. |
| Distributed rate limiting | Conflicts with the single-instance MVP and introduces Redis dependency on the request admission path. | The API horizontally scales and a global per-device limit is required. |
| User identity provider | More lifecycle and integration work than the static-key/internal-dashboard MVP needs. | Dashboard access needs SSO, audit logs, teams, or RBAC. |

## Scale and security triggers after MVP

Move beyond this stack only when evidence requires it:

- Add a durable broker/outbox before claiming durable acknowledgements or supporting client workloads that cannot tolerate a crash-loss window.
- Use Redis-backed/distributed rate limiting only once the API runs multiple instances.
- Introduce OIDC and RBAC when dashboard access extends beyond a small internal team.
- Partition PostgreSQL telemetry tables and define retention before storing high-volume data beyond short demo windows.
- Move PostgreSQL to a dedicated managed service or host when measured flush latency, disk I/O, or retention exceeds the single-host budget.
- Move to a container orchestrator only when the deployment actually spans multiple hosts or needs automatic scaling/recovery beyond Compose and systemd.
