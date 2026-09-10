# Telemetry MVP Delivery Tickets

## Backlog rules

- All tickets implement `docs/prd.md`; architecture, database, and UI decisions are defined in `docs/Architecture.md`, `docs/database.md`, and `docs/DESIGN.md`.
- A ticket is complete only when its acceptance criteria are demonstrably verified. Do not claim benchmark or reliability outcomes without test evidence.
- `202 Accepted` means successfully placed in the in-memory queue. It must never be described as a durable-write acknowledgement.
- Dashboard tickets must follow `docs/DESIGN.md`, including tokens, dense data-table patterns, keyboard behavior, and explicit loading/empty/error/degraded states.

## Dependency map

```text
TEL-001 ─┬─> TEL-002 ─┬─> TEL-015 ─> TEL-016 ─> TEL-018 ─┬─> TEL-021 ─┐
         │           │                                    └─> TEL-022 ─┤
         ├─> TEL-003 ─┬─> TEL-006 ─> TEL-007 ─> TEL-009 ─> TEL-011 ─┬─> TEL-012 ─> TEL-013 ─> TEL-016
         │           │                     ▲                         │
         │           └─> TEL-019 ────────────────────────────────────┼─> TEL-023
         ├─> TEL-004 ─> TEL-008 ─────────────────────────────────────┘
         ├─> TEL-005 ────────────────────────────────────────────────┘
         └─> TEL-020 ────────────────────────────────────────> TEL-021 / TEL-022 / TEL-023

TEL-011 + TEL-012 + TEL-013 + TEL-014 + TEL-015 ─> TEL-017 ─> TEL-025 ─> TEL-026 ─> TEL-028
TEL-021 + TEL-022 + TEL-023 ─> TEL-024 ────────────────────────────────────────> TEL-028
TEL-017 + TEL-024 + TEL-026 ─> TEL-027 ────────────────────────────────────────> TEL-028
```

## Ticket summary

| ID | Ticket | Blocking tickets | Blocks |
|---|---|---|---|
| TEL-001 | Establish repository and development baseline | — | 002–005, 020 |
| TEL-002 | Create local service topology | 001 | 015, 017, 025, 027 |
| TEL-003 | Create PostgreSQL schema and migrations | 001 | 006, 012, 019, 025 |
| TEL-004 | Define OpenAPI API contract | 001 | 008, 011, 018, 019 |
| TEL-005 | Create Go service foundation | 001 | 007–011, 017 |
| TEL-006 | Bootstrap devices and static API credentials | 003 | 007 |
| TEL-007 | Authenticate device ingestion requests | 005, 006 | 009, 011, 025 |
| TEL-008 | Validate telemetry batches | 004, 005 | 011, 025 |
| TEL-009 | Enforce per-device token-bucket limits | 005, 007 | 011, 025 |
| TEL-010 | Build bounded device-sharded admission queue | 005 | 011–014, 025 |
| TEL-011 | Deliver telemetry batch admission endpoint | 004, 007–010 | 012–014, 026 |
| TEL-012 | Persist raw telemetry in PostgreSQL batches | 003, 010 | 013, 016, 025, 026 |
| TEL-013 | Add persistence retries and admission backpressure | 010, 012 | 016, 017, 025, 026 |
| TEL-014 | Compute rolling aggregates and threshold state | 010, 011 | 015, 025 |
| TEL-015 | Publish Redis aggregate state and TTLs | 002, 014 | 016, 018, 025, 026 |
| TEL-016 | Handle Redis degradation and aggregate rebuild | 012, 013, 015 | 018, 025, 026 |
| TEL-017 | Expose health, readiness, logs, and metrics | 002, 005, 011–015 | 025–028 |
| TEL-018 | Deliver Redis-backed live read APIs | 004, 015, 016 | 021, 022, 025 |
| TEL-019 | Deliver bounded manual-history API | 003, 004 | 023, 025 |
| TEL-020 | Establish frontend design-system foundation | 001 | 021–024 |
| TEL-021 | Build Fleet overview | 018, 020 | 024, 028 |
| TEL-022 | Build Device detail view | 018, 020 | 024, 028 |
| TEL-023 | Build manual Device history view | 019, 020 | 024, 028 |
| TEL-024 | Validate UI state, accessibility, and responsive behavior | 020–023 | 027, 028 |
| TEL-025 | Build automated backend and integration test suite | 003, 007–016, 018, 019 | 026, 028 |
| TEL-026 | Execute load, overload, and dependency-failure tests | 011–013, 015–017, 025 | 027, 028 |
| TEL-027 | Package secure deployment and operations configuration | 002, 017, 024, 026 | 028 |
| TEL-028 | Run end-to-end MVP acceptance demo | 021–027 | — |

---

## Foundation

### TEL-001 — Establish repository and development baseline

**Goal:** Create a reproducible project baseline for the Go service, React application, configuration, linting, test commands, and documentation links.

**Blocks:** TEL-002, TEL-003, TEL-004, TEL-005, TEL-020.

**Acceptance criteria:**

- Go and React/TypeScript projects build from a clean checkout using pinned toolchain versions.
- Repository layout follows `docs/Architecture.md` and does not expose application packages as public Go libraries.
- Environment-variable names are documented in `.env.example` without secrets.
- Formatting, static analysis, unit-test, frontend-test, and build commands are documented and runnable locally.
- No production dependency uses a floating `latest` image or package version.

### TEL-002 — Create local service topology

**Goal:** Define the Docker Compose topology for API, PostgreSQL, Redis, frontend/static serving, Prometheus, Grafana, and Caddy.

**Blocks:** TEL-015, TEL-017, TEL-025, TEL-027.

**Acceptance criteria:**

- Compose starts PostgreSQL and Redis on an internal network with no host-published database or cache ports.
- Only Caddy publishes external HTTP/HTTPS ports in the deployment profile.
- PostgreSQL uses a named persistent volume; Redis is explicitly marked rebuildable.
- Health checks and service startup dependencies are defined.
- Configuration separates local development defaults from deployment secrets.

### TEL-003 — Create PostgreSQL schema and migrations

**Goal:** Create the durable data model in `docs/database.md` through versioned SQL migrations.

**Blocks:** TEL-006, TEL-012, TEL-019, TEL-025.

**Acceptance criteria:**

- Migrations create `devices`, `api_clients`, `telemetry_events`, and `measurement_thresholds` plus required enum types.
- `telemetry_events` enforces `UNIQUE(device_id, event_id)`, device foreign key, measurement type, canonical unit, and metric-range constraints.
- Seeded threshold rows reproduce the exact normal/warning/critical boundary rules in the PRD.
- Required indexes exist: idempotency key, device/measurement/event-time history, device/event-time history, and BRIN `received_at` recovery index.
- Migration up/down behavior is tested against a clean PostgreSQL container.

### TEL-004 — Define the OpenAPI API contract

**Goal:** Create the OpenAPI 3.1 source of truth for the MVP HTTP API.

**Blocks:** TEL-008, TEL-011, TEL-018, TEL-019.

**Acceptance criteria:**

- Contract defines ingestion, live fleet, live device, manual history, health, readiness, and metrics endpoints.
- Ingestion request schema accepts 1–500 events and documents required event fields, units, and error envelope.
- Contract defines API-key security only for ingestion routes and separate operator protection assumptions for dashboard routes.
- `202`, `400`, `401`, `413`, `415`, `429`, and `503` responses have documented schemas and semantics.
- `X-Request-ID` and `Retry-After` behavior are documented where applicable.

### TEL-005 — Create Go service foundation

**Goal:** Build the service composition root, routing boundary, request middleware, dependency lifecycle, and graceful shutdown shell.

**Blocks:** TEL-007 through TEL-011, TEL-017.

**Acceptance criteria:**

- Service uses `net/http`, `context`, `chi`, structured `slog` JSON logs, and explicit server/dependency timeouts.
- Every request receives or propagates a request ID; logs contain the request ID.
- Recovery middleware prevents a handler panic from crashing the process.
- Service starts and closes dependency clients/workers through one lifecycle owner.
- Graceful shutdown first stops admission, then drains configured work, then closes dependencies.

---

## Ingestion admission

### TEL-006 — Bootstrap devices and static API credentials

**Goal:** Load seeded device records and static API credential metadata without ever persisting or logging plaintext keys.

**Blocks:** TEL-007.

**Acceptance criteria:**

- A structured key uses `tk_<key_id>_<secret>` and maps to exactly one enabled device for MVP.
- Credentials are loaded from protected deployment configuration or a controlled bootstrap process.
- Only a keyed HMAC-SHA-256 hash of the secret is stored/compared; hashes are 32 bytes.
- Missing, malformed, disabled, and unknown key IDs have a generic unauthorized outcome.
- Secrets, authorization headers, and full payloads are excluded from logs and metrics.

### TEL-007 — Authenticate device ingestion requests

**Goal:** Authenticate `POST /v1/telemetry/batches` before admission and bind it to exactly one device.

**Blocks:** TEL-009, TEL-011, TEL-025.

**Acceptance criteria:**

- Ingestion requires `X-API-Key` over HTTPS.
- Key lookup uses public key ID plus constant-time secret comparison.
- Missing, invalid, malformed, or disabled credentials return `401` without enqueueing work.
- Valid authentication provides the authorized internal and external device identity to downstream handlers.
- A device key cannot access dashboard, history, health, metrics, Redis, or PostgreSQL routes.

### TEL-008 — Validate telemetry batches

**Goal:** Reject malformed or unreasonable work before it reaches a queue.

**Blocks:** TEL-011, TEL-025.

**Acceptance criteria:**

- JSON only; unsupported media type returns `415`.
- Request body over 1 MB returns `413` before large allocation or queue activity.
- Entire batch is rejected with `400` when empty, over 500 events, malformed, or containing any invalid event.
- Every event validates event ID, matching authenticated device ID, RFC 3339 timestamp, measurement type, finite value, canonical/type-compatible unit, and PRD numeric range.
- Events more than 60 seconds late are marked for persistence-only processing rather than rejected.

### TEL-009 — Enforce per-device token-bucket limits

**Goal:** Limit sustained device traffic while allowing short valid bursts.

**Blocks:** TEL-011, TEL-025.

**Acceptance criteria:**

- One concurrency-safe in-memory bucket exists per authenticated API key/device.
- Bucket refills at 10 event tokens/sec with a 100-event maximum capacity.
- Token cost equals number of events in the batch, not request count.
- Insufficient tokens reject the whole batch with `429` and do not enqueue partial work.
- `Retry-After` is included when the next permitted admission time can be calculated.
- Documentation and metrics make clear that this limit is instance-local.

### TEL-010 — Build bounded device-sharded admission queue

**Goal:** Create the bounded, fixed-concurrency handoff between the synchronous HTTP path and processing workers.

**Blocks:** TEL-011, TEL-012, TEL-013, TEL-014, TEL-025.

**Acceptance criteria:**

- Queue capacity and shard/worker count are configuration values with safe validation.
- Hashing a device ID always routes its batches to one owning shard worker, preserving admitted order for that device without aggregate locks.
- No request or event creates an unbounded goroutine.
- Full admission capacity returns `503` immediately and increments a queue-rejected metric.
- Per-shard depth, total depth, capacity, and worker utilization are observable.

### TEL-011 — Deliver telemetry batch admission endpoint

**Goal:** Compose the OpenAPI contract, authentication, validation, rate limiting, capacity gate, and queue into the production ingestion endpoint.

**Blocks:** TEL-012, TEL-013, TEL-014, TEL-026.

**Acceptance criteria:**

- `POST /v1/telemetry/batches` returns `202` only after successful authenticate/validate/rate-limit/enqueue flow.
- Successful response includes request ID, accepted event count, and `accepted` status.
- All documented failure responses are precise, stable, and contain request IDs.
- Admission performs no synchronous Redis or PostgreSQL write.
- The endpoint documents and logs that `202` is in-memory admission, not durability.

---

## Processing, caching, and persistence

### TEL-012 — Persist raw telemetry in PostgreSQL batches

**Goal:** Persist accepted raw events efficiently and idempotently.

**Blocks:** TEL-013, TEL-016, TEL-025, TEL-026.

**Acceptance criteria:**

- A bounded persistence input accepts raw events from worker processing.
- Writer flushes a parameterized PostgreSQL transaction at 1,000 records or 2 seconds, whichever occurs first.
- Inserts use `ON CONFLICT (device_id, event_id) DO NOTHING`.
- Duplicate client events result in one durable row without an error or duplicate aggregate crash.
- Flush size, duration, success/failure, and conflict count are observable.

### TEL-013 — Add persistence retries and admission backpressure

**Goal:** Isolate PostgreSQL failure without unbounded memory growth.

**Blocks:** TEL-016, TEL-017, TEL-025, TEL-026.

**Acceptance criteria:**

- Failed writes retry at 100 ms, 200 ms, 400 ms, then capped exponential delay up to 5 seconds, for at most three attempts.
- Retry-buffer capacity is capped at 10,000 events.
- Retry exhaustion or buffer saturation activates an admission-backpressure state.
- New requests return `503` while backpressure is active; existing memory remains bounded.
- Metrics/logs report retry count, buffer occupancy, persistence failure, and backpressure transition.

### TEL-014 — Compute rolling aggregates and threshold state

**Goal:** Compute fast, safe five-minute live measurement aggregates per device and metric.

**Blocks:** TEL-015, TEL-025.

**Acceptance criteria:**

- Owner workers maintain processing-time five-minute state for `device_id + measurement_type`.
- Implementation uses 300 one-second buckets or an equally benchmarked bounded design; it must not retain an unbounded raw-event window.
- Aggregate output includes latest value/timestamp, count, sum, average, min, max, window start/end, and threshold status.
- Events more than 60 seconds behind system time persist but do not alter live aggregate state.
- Exact temperature, voltage, battery, and pressure threshold boundaries match the PRD.

### TEL-015 — Publish Redis aggregate state and TTLs

**Goal:** Materialize aggregate snapshots for fast live-dashboard reads.

**Blocks:** TEL-016, TEL-018, TEL-025, TEL-026.

**Acceptance criteria:**

- Redis keys use the versioned aggregate, device-last-seen, and device-measurement patterns in `docs/Architecture.md`.
- Each aggregate snapshot has a 15-minute TTL refreshed on valid updates.
- Device last-seen supports stale/offline classification after 60 seconds without processed telemetry.
- Cache updates are coalesced/pipelined and bounded so Redis cannot dominate worker throughput.
- Redis command failures use short timeouts and are surfaced in metrics/logs.

### TEL-016 — Handle Redis degradation and aggregate rebuild

**Goal:** Keep ingestion/persistence functional during cache failure and restore derived state safely.

**Blocks:** TEL-018, TEL-025, TEL-026.

**Acceptance criteria:**

- Redis errors move the application to an observable degraded mode without blocking PostgreSQL persistence indefinitely.
- Workers skip failed cache publication after short timeout and continue persist-only processing while PostgreSQL capacity is healthy.
- Live-read API can report `mode: degraded`; it never falls back to automatic PostgreSQL polling.
- A bounded, operator-controlled cache rebuild replays only the required PostgreSQL history and does not interfere unboundedly with ingestion.
- Recovery restores cache publication and records a visible mode transition.

### TEL-017 — Expose health, readiness, logs, and metrics

**Goal:** Make admission safety and downstream degradation observable.

**Blocks:** TEL-025, TEL-026, TEL-027, TEL-028.

**Acceptance criteria:**

- `/healthz` reports process liveness only.
- `/readyz` reports whether admission is safe and includes PostgreSQL, Redis/degraded, queue, and retry-backpressure state without leaking secrets.
- `/metrics` exposes API latency, admission outcome, queue depth/capacity, worker utilization, Redis errors, flush metrics, retries, and retry-buffer occupancy.
- Logs are structured JSON and include request ID, device ID, batch size, outcome, and reason without raw secrets or payloads.
- Metrics and health endpoints are reachable only from the internal monitoring network.

---

## Read APIs and frontend

### TEL-018 — Deliver Redis-backed live read APIs

**Goal:** Serve fleet and device live views entirely from derived Redis state.

**Blocks:** TEL-021, TEL-022, TEL-025.

**Acceptance criteria:**

- `GET /v1/live/fleet` returns active/stale/warning/critical counts, current mode, refresh context, and current device summaries from Redis.
- `GET /v1/live/devices/{deviceId}` returns latest values, five-minute aggregates, threshold state, freshness, and aggregate window metadata from Redis.
- Live endpoints make no PostgreSQL `SELECT` calls during normal UI refresh.
- Missing cache data is returned as unknown/stale rather than healthy.
- Redis outage returns documented degraded behavior usable by the frontend without exposing internal errors.

### TEL-019 — Deliver bounded manual-history API

**Goal:** Provide an explicit PostgreSQL history query without making live UI traffic hit the database.

**Blocks:** TEL-023, TEL-025.

**Acceptance criteria:**

- `GET /v1/history/devices/{deviceId}` requires bounded `from`, `to`, and page/limit parameters.
- Server validates a maximum allowed time range and page size and applies an SQL statement timeout.
- Results are ordered predictably and use indexed parameterized queries.
- Endpoint is protected by operator route policy, not a device API key.
- Endpoint is excluded from automatic UI polling and clearly labelled as manual history.

### TEL-020 — Establish frontend design-system foundation

**Goal:** Create the React/Vite UI foundation that enforces `docs/DESIGN.md`.

**Blocks:** TEL-021, TEL-022, TEL-023, TEL-024.

**Acceptance criteria:**

- Inter is self-hosted; the dark graphite, teal-accent token system and semantic status colors are available as named tokens.
- Shared application shell, page header, status chip, button, form controls, skeleton, error, and tooltip primitives are implemented from accessible Radix/shadcn patterns.
- Use TanStack Query for server-state loading/cache/retry and native `fetch` through a typed API client.
- Use Lucide as the sole icon set; no emoji or decorative icon substitutions.
- Storybook stories cover shared component variants and required states.

### TEL-021 — Build Fleet overview

**Goal:** Give operators a dense, Redis-backed fleet triage surface.

**Blocks:** TEL-024, TEL-028.

**Acceptance criteria:**

- Page follows the Fleet layout in `docs/DESIGN.md`: page header, status strip, filter toolbar, and sticky-header data table.
- Status strip filters Critical, Warning, Stale, and Active device rows.
- Table supports device search, status filter, measurement selection, sorting, result count, and preserved filter/scroll state on navigation.
- Rows show device, overall state, temperature, voltage, battery, pressure, and last processed time with status text plus color.
- Query polling is 2–5 seconds, shows last refresh, and never triggers PostgreSQL history reads.
- First-use empty, filter-empty, loading skeleton, error/retry, and degraded/stale-last-known states match the design brief.

### TEL-022 — Build Device detail view

**Goal:** Give operators an interpretable five-minute live view for one device.

**Blocks:** TEL-024, TEL-028.

**Acceptance criteria:**

- Page provides breadcrumb, device header, overall status, last processed time, and a single `View history` primary action.
- Four measurement facts show current value, unit, threshold status, and freshness before charts.
- 2×2 desktop chart grid renders five-minute Temperature, Voltage, Battery, and Pressure snapshots using Recharts.
- Every chart has threshold labels/bands plus accessible textual summary; chart hover is not the only value path.
- Loading, missing aggregate, error, stale, and Redis-degraded states match `docs/DESIGN.md` without automatic PostgreSQL fallback.

### TEL-023 — Build manual Device history view

**Goal:** Let an operator intentionally inspect bounded raw-event history for one device.

**Blocks:** TEL-024, TEL-028.

**Acceptance criteria:**

- Page carries device context through breadcrumb/header and labels results as manual, non-live data.
- Query controls provide a bounded time selection, measurement filter, validation feedback, and one `Run query` primary action.
- Result table includes event time, metric, value, unit, received time, and event ID with predictable pagination.
- Query does not execute until user action; filters remain intact after error.
- Initial, no-result, loading, error/retry, and PostgreSQL-unavailable states match the design brief.

### TEL-024 — Validate UI state, accessibility, and responsive behavior

**Goal:** Verify that all MVP screens remain usable and honest across states, input methods, and screen sizes.

**Blocks:** TEL-027, TEL-028.

**Acceptance criteria:**

- Fleet, Device detail, and History meet every required state in `docs/DESIGN.md`.
- Final color combinations meet WCAG 2.2 AA contrast targets; status never relies on color alone.
- Keyboard order follows visual order; skip link, visible focus rings, real table links, sortable-header semantics, dialog focus behavior, and accessible chart summaries are verified.
- Desktop 1440/1280 px, tablet, mobile, 200% zoom, reduced motion, long device IDs, missing values, and wide tables are checked.
- Live data refresh does not repeatedly announce polling to screen readers; degraded state changes are announced once.

---

## Quality, performance, and deployment

### TEL-025 — Build automated backend and integration test suite

**Goal:** Prove critical validation, correctness, persistence, cache, and failure behavior before load testing.

**Blocks:** TEL-026, TEL-028.

**Acceptance criteria:**

- Go unit tests cover validation boundaries, authentication, token bucket behavior, threshold evaluation, aggregate buckets, queue-full behavior, retry/backpressure, and graceful shutdown.
- Integration tests use Testcontainers with real PostgreSQL and Redis.
- Tests verify migration constraints, transactional persistence, duplicate conflict behavior, Redis TTLs, degraded persist-only behavior, and bounded rebuild behavior.
- `go test -race ./...`, formatting, vet, and configured linting pass in CI.
- API contract tests verify documented success and failure responses.

### TEL-026 — Execute load, overload, and dependency-failure tests

**Goal:** Measure the PRD performance and reliability targets on the specified topology.

**Blocks:** TEL-027, TEL-028.

**Acceptance criteria:**

- k6 runs from a separate load-generator host/container, not the 4-vCPU system under test.
- Baseline uses at least 2,000 distinct credentials/devices, 10 events per request, 2,000 requests/sec, and 20,000 events/sec.
- Report p50/p95/p99 admission latency, error rates, queue depth, worker utilization, flush time, Redis latency, CPU, RAM, and disk I/O.
- Verify p95 successful enqueue latency below 20 ms and p95 Redis read latency below 5 ms, or report failed target with the exact evidence.
- Run rate-limit, queue-overload, Redis-outage, PostgreSQL-outage, retry-buffer, duplicate-event, and soak scenarios.

### TEL-027 — Package secure deployment and operations configuration

**Goal:** Prepare the portfolio-quality single-host Docker deployment and runbook.

**Blocks:** TEL-028.

**Acceptance criteria:**

- Compose deployment runs as non-root containers with pinned images, named PostgreSQL volume, restart policies, and health checks.
- Caddy serves static SPA/API over TLS; dashboard and live/history routes require operator access policy, while ingestion uses only device API-key auth.
- Only ports 80/443 are publicly published; PostgreSQL, Redis, Prometheus, Grafana, readiness, and metrics are restricted to internal/private networks.
- Prometheus scrapes metrics and Grafana displays ingestion, queue, persistence, Redis, and host-health signals.
- Runbook documents backup/restore verification, secret injection, startup/shutdown, Redis degradation, PostgreSQL backpressure, cache rebuild, and the non-durable `202` boundary.

### TEL-028 — Run end-to-end MVP acceptance demo

**Goal:** Demonstrate that the finished product meets the PRD’s portfolio-quality acceptance criteria.

**Blocks:** none; this is the final milestone.

**Acceptance criteria:**

- Valid authenticated device batches reach the dashboard and PostgreSQL through the complete admission, processing, cache, and persistence flow.
- Invalid JSON, oversized body, invalid metric, invalid key, rate-limit exhaustion, queue saturation, and PostgreSQL backpressure all return the specified response without worker/process crash.
- Dashboard confirms Redis-backed live reads and visibly enters the required degraded state during Redis failure without database refresh fallback.
- Duplicate `(device_id, event_id)` submissions result in one raw event row.
- Performance/load evidence, observability screenshots or exports, and known MVP loss boundaries are documented honestly.
- Demo is checked against `docs/prd.md`, `docs/Architecture.md`, `docs/database.md`, and `docs/DESIGN.md`.
