# Application Architecture

## 1. Architecture decision

Build a single deployable Go API/service and a separately built React single-page application (SPA). Run PostgreSQL, Redis, Prometheus, Grafana, and Caddy alongside the service in Docker Compose on one Ubuntu host.

The API has two deliberate paths:

- **Admission path:** authenticate, bound input, validate, rate-limit, check capacity, enqueue, return `202`. It performs no Redis or PostgreSQL I/O.
- **Processing path:** a fixed set of device-sharded workers updates live aggregate state and feeds a bounded PostgreSQL batch writer.

This preserves the PRD's p95 `<20 ms` response target while preventing unbounded goroutines, queues, and retry memory. A `202` confirms in-memory admission only; it is not a durable-write acknowledgement.

```text
                         ┌──────────────────────────────────────────┐
Device ── HTTPS ────────>│ Caddy                                    │
                         │ /v1/telemetry/* -> Go API                │
Operator browser ─HTTPS─>│ / -> SPA; /v1/live/* -> Go API           │
                         └──────────────────┬───────────────────────┘
                                            │
                         ┌──────────────────▼───────────────────────┐
                         │ Go service                                │
                         │ admission -> device-shard queues ->       │
                         │ owner workers -> cache publisher +        │
                         │                   persistence batcher     │
                         └───────────────┬───────────────┬───────────┘
                                         │               │
                              derived state│               │raw events
                                         ▼               ▼
                                      Redis          PostgreSQL
                                         │               │
                    Redis-only live reads│               │manual bounded history
                                         └───────┬───────┘
                                                 ▼
                                          Go read handlers

Prometheus ──scrapes──> /metrics; Grafana reads Prometheus
```

## 2. Project structure

This is the proposed implementation structure, not code created by this document.

```text
.
├── api/
│   └── openapi.yaml                 # Versioned OpenAPI 3.1 contract
├── cmd/
│   └── api/
│       └── main.go                  # Composition root and graceful shutdown
├── internal/
│   ├── app/
│   │   └── app.go                   # Service wiring and lifecycle
│   ├── config/
│   │   └── config.go                # Environment validation; no secrets logged
│   ├── httpapi/
│   │   ├── router.go
│   │   ├── middleware/               # Request ID, recovery, timeouts, access logs
│   │   ├── ingest_handler.go
│   │   ├── live_handler.go
│   │   ├── history_handler.go
│   │   └── health_handler.go
│   ├── auth/
│   │   └── apikey.go                 # Key-ID lookup, HMAC verification, device binding
│   ├── admission/
│   │   ├── validator.go              # JSON/schema/range/device validation
│   │   ├── limiter.go                # Per-key token buckets
│   │   └── dispatcher.go             # Capacity gate and device-shard routing
│   ├── processing/
│   │   ├── shard_worker.go           # Single owner of assigned device state
│   │   ├── window.go                 # 300 one-second rolling buckets
│   │   ├── thresholds.go
│   │   └── cache_publisher.go        # Coalesced Redis snapshots
│   ├── persistence/
│   │   ├── batcher.go                # 1,000-event / 2-second flush policy
│   │   ├── repository.go             # pgx transactions and parameterized SQL
│   │   └── retry.go                  # Bounded retries and backpressure signal
│   ├── cache/
│   │   ├── redis.go
│   │   └── rebuild.go                # Explicit cache rebuild process
│   ├── telemetry/
│   │   └── models.go                 # Domain models and error codes
│   └── observability/
│       ├── metrics.go
│       └── logging.go
├── migrations/
│   ├── 001_initial_schema.up.sql
│   └── 001_initial_schema.down.sql
├── web/
│   ├── src/
│   │   ├── api/                      # Typed fetch client and DTOs
│   │   ├── app/                      # Routing and query-client setup
│   │   ├── features/
│   │   │   ├── fleet/
│   │   │   ├── device/
│   │   │   ├── history/
│   │   │   └── system-status/
│   │   ├── components/
│   │   └── test/
│   ├── package.json
│   └── vite.config.ts
├── tests/
│   ├── integration/                  # Testcontainers tests
│   ├── e2e/                          # Playwright tests
│   └── load/                         # k6 scenarios and generated credentials
├── deploy/
│   ├── compose.yaml
│   ├── Caddyfile
│   ├── prometheus.yml
│   └── systemd/
├── docs/
│   ├── prd.md
│   ├── Techstack.md
│   └── Architecture.md
├── Dockerfile                        # Go API multi-stage build
├── web/Dockerfile                    # Static SPA build
└── .env.example                      # Variable names only; never real values
```

`internal/` prevents the Go application packages from being imported as a public library. The OpenAPI contract, SQL migrations, deployment configuration, and load tests remain top-level because they are first-class product artifacts.

## 3. Frontend/backend flow

### 3.1 Routing and access boundaries

| Caddy route | Destination | Required access | Notes |
|---|---|---|---|
| `/` and static assets | Built React SPA | Operator Basic Auth + private-network/VPN or IP allowlist | Caddy serves static files; no Node.js production server. |
| `/v1/live/*` | Go API | Same operator protection | Redis-only live reads. |
| `/v1/history/*` | Go API | Same operator protection | Manual, bounded PostgreSQL reads only. |
| `/v1/telemetry/batches` | Go API | Device `X-API-Key` only | Must not be gated by dashboard Basic Auth. |
| `/healthz`, `/readyz`, `/metrics` | Go API | Internal Docker network / monitoring only | Do not proxy publicly. |

All browser and dashboard API traffic use the same HTTPS origin. This avoids cross-origin credentials and removes a production CORS requirement.

### 3.2 Dashboard flow

```text
1. Operator opens https://telemetry.example.internal/
2. Caddy enforces TLS and dashboard access policy, then returns SPA assets.
3. React starts TanStack Query polling every 2–5 seconds.
4. GET /v1/live/fleet and GET /v1/live/devices/{deviceId} reach Go read handlers.
5. Go reads compact aggregate snapshots from Redis only.
6. React renders aggregate values, five-minute charts, threshold badges, and stale status.

Redis healthy:   return current aggregate snapshot + mode="live".
Redis unhealthy: return an explicit degraded response/status; React keeps its last
                 query result, marks it stale, and displays
                 "Degraded Mode: Live Aggregates Paused (Redis Offline)".
```

The live read path must not query PostgreSQL. A manually opened history view may call the bounded history endpoint, which is visually and operationally separate from automatic dashboard refresh.

### 3.3 API service composition

The Go process creates a `pgxpool`, a Redis client, an in-memory credential registry, a token-bucket registry, a fixed number of device shards, a persistence batcher, and HTTP routes. Dependency clients and workers start before readiness is reported. On shutdown, the service first rejects new admissions, then drains shard queues until its configured deadline, and finally attempts a PostgreSQL flush.

## 4. Database structure

### 4.1 PostgreSQL logical schema

PostgreSQL is the durable source of truth. Use explicit SQL migrations and `pgx`; do not use an ORM.

| Table | Important fields and constraints | Purpose |
|---|---|---|
| `devices` | `id uuid PK`, `external_id text UNIQUE NOT NULL`, `name text`, `device_type text`, `status text`, `last_processed_at timestamptz`, `created_at timestamptz` | Fleet identity and last processed time. |
| `api_clients` | `id uuid PK`, `key_id text UNIQUE NOT NULL`, `api_key_hash bytea NOT NULL`, `device_id uuid NOT NULL REFERENCES devices`, `status text`, `created_at timestamptz` | Seeded device credential metadata. Never store a plaintext key. |
| `telemetry_events` | `id bigint generated PK`, `event_id text NOT NULL`, `device_id uuid NOT NULL REFERENCES devices`, `measurement_type text NOT NULL`, `value double precision NOT NULL`, `unit text`, `event_timestamp timestamptz NOT NULL`, `received_at timestamptz NOT NULL`, `request_id uuid NOT NULL` | Immutable raw event record. |
| `measurement_thresholds` | `measurement_type text PK`, `unit text NOT NULL`, threshold-bound fields, `updated_at timestamptz` | Seeded, auditable threshold configuration. |

Required `telemetry_events` constraints and indexes:

```text
UNIQUE (device_id, event_id)
CHECK (measurement_type IN ('temperature', 'voltage', 'battery', 'pressure'))
CHECK (value = value)                     # rejects NaN
INDEX (device_id, measurement_type, event_timestamp DESC)
INDEX (received_at DESC)                  # cache rebuild / operations
```

Enforce canonical units and the PRD's measurement-specific ranges at the API boundary. Add matching database `CHECK` constraints as defense in depth, even though the database should not be the normal validation path.

`event_timestamp` is device-supplied source time. `received_at` is server time immediately after validation/admission and is the audit/rebuild time. The five-minute live window uses worker processing time rather than either producer timestamp.

### 4.2 Redis key structure

Redis holds derived, disposable state only. Prefix keys by version so an aggregate representation can change safely.

| Key | Type | TTL | Contents |
|---|---|---:|---|
| `telemetry:v1:aggregate:{deviceId}:{measurement}` | Hash or compact JSON string | 15 min, refreshed on write | Latest value/timestamp, count, sum, average, min, max, window bounds, threshold, aggregate update time. |
| `telemetry:v1:device-last-seen` | Sorted set | Pruned by worker/job | Member is `deviceId`; score is latest processing epoch. Supports active/stale fleet counts. |
| `telemetry:v1:device-measurements:{deviceId}` | Set | 15 min, refreshed on write | Measurement types currently available for a device. |

Redis TTL removes offline aggregates after 15 minutes. Prune the `device-last-seen` sorted set on a schedule so it does not grow indefinitely. A missing aggregate is **unknown/stale**, never proof of a healthy device.

### 4.3 Five-minute sliding aggregate representation

Do not store every raw event in a Go heap or Redis just to calculate a live chart. At the benchmark target, a five-minute raw window could contain six million events.

Instead, each device-shard worker maintains a per-`deviceId + measurementType` ring of **300 one-second processing-time buckets**:

```text
bucket: second, count, sum, min, max, latest_value, latest_processed_at
window: [300 buckets] + current total_count + total_sum + current status
```

- On a non-late event, the owner worker updates its current one-second bucket.
- When time advances, it expires/reuses the old bucket and updates total count/sum.
- Minimum and maximum are recomputed across the 300 buckets when a bucket changes or expires; the bounded scan is simple and avoids complex shared locks.
- The worker materializes the resulting snapshot to Redis.

This gives a five-minute sliding dashboard window at **one-second granularity**. That is an implementation decision to validate in benchmark testing; it is appropriate for a 2–5 second polling dashboard. If exact sub-second window boundaries become a requirement, replace the bucket implementation with per-event monotonic deques and budget the added memory.

## 5. API structure

The API contract is `api/openapi.yaml` using OpenAPI 3.1. All successful and error responses use JSON and include `X-Request-ID`.

### 5.1 Endpoint map

| Method and path | Auth | Source | Response intent |
|---|---|---|---|
| `POST /v1/telemetry/batches` | Device API key | Admission queue | `202` after in-memory enqueue. |
| `GET /v1/live/fleet` | Operator proxy access | Redis | Fleet summary and service mode. |
| `GET /v1/live/devices/{deviceId}` | Operator proxy access | Redis | Latest values, five-minute aggregates, device freshness. |
| `GET /v1/history/devices/{deviceId}` | Operator proxy access | PostgreSQL | Explicit range-limited historical records/summary. |
| `GET /healthz` | Internal | Process | Liveness only. |
| `GET /readyz` | Internal | Process + dependency/admission state | Readiness and current degradation/backpressure state. |
| `GET /metrics` | Prometheus network only | Process | Prometheus exposition. |

### 5.2 Ingestion contract

Request body:

```json
{
  "events": [
    {
      "event_id": "device-generated-id",
      "device_id": "device-123",
      "timestamp": "2026-09-10T12:00:00Z",
      "measurement_type": "temperature",
      "value": 72.4,
      "unit": "C"
    }
  ]
}
```

Rules:

- Require 1–500 events and a 1 MB body limit.
- For the MVP, bind one static API key to one device; every event in the batch must match that device. This makes per-device rate limiting and ordered aggregate ownership unambiguous.
- Enforce `application/json`, finite numeric values, RFC 3339 timestamps, canonical units, measurement ranges, and whole-batch rejection.
- Charge one token per event after parsing the event count. Rejections never partially admit a batch.

Success response:

```json
{
  "request_id": "uuid",
  "accepted_events": 10,
  "status": "accepted"
}
```

Status semantics:

| Status | Meaning |
|---:|---|
| `202` | Validated/authenticated/rate-limited batch entered an in-memory shard queue; it is not guaranteed durable. |
| `400` | Invalid JSON, empty/oversized batch, invalid field, range, unit, or device binding. |
| `401` | Missing or invalid device API key. |
| `413` | Request body exceeds 1 MB. |
| `415` | Content type is not JSON. |
| `429` | API key's 10-event/sec token bucket lacks sufficient tokens; include `Retry-After` if calculable. |
| `503` | Queue or persistence retry capacity is full, or admission is paused for PostgreSQL backpressure. |

### 5.3 Read contracts

`GET /v1/live/fleet` returns an explicit `mode` field: `live` or `degraded`. `degraded` means Redis cannot supply current aggregate state; it does not trigger a PostgreSQL fallback.

Live-device responses include `last_processed_at`, `freshness` (`active`, `stale`, or `unknown`), `aggregate_window_start`, `aggregate_window_end`, `threshold_status`, and all aggregate values. The server computes stale when `now - last_processed_at > 60 seconds`.

The historical endpoint requires `from`, `to`, and a bounded `limit`; enforce a maximum time range and page size in the API to prevent an operator request from becoming a table scan. It is intentionally excluded from TanStack Query's auto-refresh loop.

## 6. Authentication flow

### 6.1 Device API key flow

Use structured static keys in this form:

```text
tk_<public-key-id>_<32-byte-random-secret>
```

`public-key-id` is not confidential and supports O(1) registry lookup. The random secret is confidential.

```text
1. Device sends HTTPS POST with X-API-Key.
2. Caddy terminates TLS and proxies only /v1/telemetry/batches.
3. API parses the public key ID, looks up the seeded active credential record.
4. API HMAC-hashes the supplied secret with the deployment pepper.
5. API performs constant-time comparison with the stored hash.
6. API obtains exactly one authorized device ID from the credential record.
7. API confirms every batch event names that device, then applies its token bucket.
8. Valid admission continues; failure returns 401, 400, or 429 without enqueue.
```

Raw keys, authorization headers, and complete payloads must never appear in structured logs, metrics labels, traces, source control, or the frontend bundle. The current static-key approach has no rotation/revocation endpoint; replacing a compromised key requires a controlled configuration deployment. This is an explicit MVP limitation.

### 6.2 Operator flow

```text
1. Operator requests dashboard or /v1/live/*.
2. Caddy requires HTTPS, an approved source network, and Basic Auth.
3. Caddy serves SPA assets or proxies the allowed dashboard API request.
4. Browser calls the same-origin API without device credentials.
5. Go read handler fetches Redis data (or returns explicit degraded state).
```

The device ingestion route must be excluded from Basic Auth, and the dashboard/history routes must never accept a device API key as operator authorization. This separation prevents a compromised device key from reading fleet data.

## 7. Data flow

### 7.1 Ingestion and processing flow

```text
Device batch
  -> Caddy TLS / request-size enforcement
  -> Go request ID + HTTP timeout middleware
  -> API-key verification and single-device binding
  -> JSON/schema/range/timestamp validation
  -> event-cost token-bucket decision
  -> global persistence-backpressure check
  -> enqueue to shard[hash(deviceId) % shardCount]
  -> 202 Accepted

Shard owner worker
  -> persist every raw event to bounded persistence input
  -> for events not >60 seconds late: update 300-bucket window and threshold
  -> mark aggregate key dirty
  -> cache publisher pipelines compact snapshots to Redis every <=250 ms

Persistence batcher
  -> collect 1,000 events or wait 2 seconds
  -> PostgreSQL transaction
  -> parameterized INSERT ... ON CONFLICT (device_id, event_id) DO NOTHING
  -> committed: release memory and update persistence metrics
  -> failed: retry 3 times with 100 ms -> 200 ms -> 400 ms (cap 5 seconds)
  -> exhausted / 10,000 buffered events: set admission backpressure; new work gets 503
```

The worker order above intentionally follows the PRD's cache-first processing model. It means a Redis snapshot can briefly contain an accepted event that later fails PostgreSQL persistence. The dashboard is therefore eventually consistent and not proof of durability. If product requirements later demand dashboard state reflect only durable events, move cache publication to after a successful database flush and accept up to the configured flush delay.

### 7.2 Device sharding and concurrency model

Use a fixed `shardCount` with one owner worker per shard. Hash the authenticated `device_id` to a shard. Each key's batches therefore execute serially in one worker while different devices execute concurrently.

Benefits:

- No mutex or distributed lock protects a device's rolling aggregate.
- Events preserve their admission order per device.
- Worker capacity and per-shard queue capacity are bounded and observable.
- Redis writes can be coalesced by the shard worker.

This works because MVP keys are one-device keys. If a future client can submit a multi-device batch, split it into per-device work units only after whole-batch validation and reserve capacity for all units before admitting any; otherwise partial enqueue would violate the all-or-nothing batch contract.

### 7.3 Redis outage flow

```text
Redis timeout/error
  -> cache publisher opens degraded state and increments Redis-error metrics
  -> worker skips cache publication after a short timeout; it never blocks persistence indefinitely
  -> raw events still enter PostgreSQL batcher
  -> live endpoint returns degraded status; SPA retains last result as stale
  -> operator sees required Redis-offline banner
  -> after recovery, an explicit/restricted rebuild scans bounded PostgreSQL history
     using received_at, replays processing-time buckets, and repopulates Redis
```

The cache rebuild operation must be operator-controlled or rate-limited so rebuilding does not contend with ingestion or create an unbounded historical scan.

### 7.4 PostgreSQL outage flow

```text
PostgreSQL write error
  -> retry failed batch up to three times with exponential backoff
  -> retain only up to 10,000 retry-buffered events
  -> buffer full or retries exhausted: expose not-ready/backpressure state
  -> reject new ingestion with 503 before memory can grow
  -> emit logs, metrics, and operator-visible failure state
```

The in-memory queue and retry buffer are intentionally not durable. A process or host failure can lose accepted but unpersisted events.

## 8. Deployment architecture

```text
Public internet / device network
  └── host firewall: expose only 80/443 (and restricted administration)
       └── Caddy container: TLS, route policy, dashboard Basic Auth
            ├── static React files
            └── api container
                 ├── PostgreSQL on private Compose network + named data volume
                 ├── Redis on private Compose network
                 └── Prometheus on private Compose network
                      └── Grafana on private operator network
```

- Only Caddy publishes host ports. PostgreSQL, Redis, metrics, and Grafana must not be publicly published.
- API and frontend runtime containers use non-root users and immutable, version-pinned images.
- Docker Compose runs under `systemd` with restart policy and health checks.
- PostgreSQL backups, restore tests, disk alerts, and Compose configuration backups are mandatory. Redis is rebuildable and does not replace PostgreSQL backup.
- Secrets enter at deployment time through environment/secrets management and are never included in images or committed `.env` files.

## 9. Important security risks

| Risk | Why it matters here | Required mitigation |
|---|---|---|
| Static device key compromise | A valid key admits telemetry for its bound device and can consume its quota. Static MVP keys lack self-service rotation/revocation. | Use 32-byte random secrets, HMAC-hash + constant-time compare, HTTPS only, never log keys, bind one key to one device, and replace configuration promptly after compromise. Move to managed key rotation or mTLS/OIDC when needed. |
| Dashboard exposure | Fleet health data should not be publicly browseable; a device key must not confer read access. | Separate Caddy routes, Basic Auth over TLS, VPN/private-network or IP allowlist, and no device key on dashboard routes. Adopt OIDC/RBAC before broader use. |
| Public database, Redis, or metrics ports | Enables data theft, cache manipulation, credential exposure, and operational reconnaissance. | Publish only Caddy 80/443; keep all other services on private Docker networks with authentication. |
| Payload/resource exhaustion | Large or deeply nested JSON, huge identifiers, and concurrent bad requests can consume CPU and memory before queuing. | 1 MB `MaxBytesReader`, JSON depth/field controls, 1–500 batch limit, identifier length limits, request deadlines, rate limits, bounded queues, and 503 backpressure. |
| Injection | Device-controlled values flow to storage and dashboard. | Parameterized SQL only, allowlisted measurement types/units, React's escaped rendering, no untrusted HTML, and constrain historical filters. |
| API-key enumeration or timing leak | Credential lookup/compare can disclose valid keys. | Structured key ID plus high-entropy secret; generic 401 response; constant-time secret comparison; no user-controlled key values in logs/metrics. |
| Secrets in CI, images, or logs | Compromise would expose all device credentials or database access. | CI secret store, deployment injection, redaction middleware, restricted file permissions, image scanning, and a committed `.env.example` with names only. |
| Cache state mistaken for durable truth | Redis is derived and may show cache-first events that never commit. | Label live data as eventually consistent; surface degraded/stale state; retain PostgreSQL as source of truth; do not claim durable `202`. |
| Clock skew | Late-event filtering, TTL, and stale status all depend on time. | Synchronize host/device clocks with NTP; record server `received_at`; monitor clock health; treat device timestamps as data, not authority. |
| Dependency/container supply chain | A vulnerable base image or package compromises a single-host deployment. | Pin versions/digests, scan images/dependencies in CI, patch routinely, and run containers as non-root. |

## 10. Important performance and reliability risks

| Risk | Impact | Design response |
|---|---|---|
| 20,000 events/sec on one 4-vCPU, 8-GB host | This is an aggressive portfolio benchmark when API, Redis, PostgreSQL, and monitoring share CPU, RAM, and SSD I/O. | Treat it as measured acceptance, not a promise. Benchmark on a separate load generator; tune shard count, pgx pool, batch size, and Redis publication interval from metrics. |
| PostgreSQL unique-index write load | `ON CONFLICT DO NOTHING` requires checking the `(device_id, event_id)` unique index for every row. At 20,000 events/sec this can dominate disk/CPU. | Batch 1,000 inserts, keep schema/indexes minimal, monitor flush latency/WAL/disk, and define short demo retention. Partition or separate DB capacity only after evidence. |
| Raw-event retention explosion | 20,000 events/sec equals 1.728 billion rows/day; it cannot be retained indefinitely on the target host. | Set explicit demo retention, purge/test data, and document benchmark duration. Add time partitioning and retention jobs before any sustained deployment. |
| Aggregate state cardinality | Thousands of active device/measurement keys can make per-event locks, raw window retention, and Redis writes expensive. | Use device-shard ownership, 300 one-second buckets, compact snapshots, and <=250 ms coalesced Redis pipelines. Track active-key count and Redis memory. |
| Redis synchronous dependency | Slow Redis updates could block workers and fill queues. | Use short Redis timeouts, pipelined/coalesced writes, a degraded circuit state, and persist-only mode. Never wait indefinitely on Redis. |
| PostgreSQL outage | Retry buffers can fill and a non-durable in-memory architecture loses accepted data on crash. | Cap retries at three and buffer at 10,000, return 503 early, alert operators, and state the loss boundary honestly. Add a durable queue only when product requirements justify it. |
| Uneven shard load | A noisy device maps to one shard and may create head-of-line blocking for its shard peers. | Hash by device, make shard queues observable, enforce per-device rate limits, and benchmark shard count. Revisit key partitioning only if metrics show skew. |
| Dashboard poll amplification | Many open dashboards can create Redis pressure despite fast reads. | Poll every 2–5 seconds, use TanStack Query cache/deduplication, keep responses compact, and protect dashboard access. |
| Load generator distortion | Running k6 on the same 4-vCPU system produces misleading latency and throughput measurements. | Generate substantial load from a separate host and capture host/API/Redis/PostgreSQL metrics together. |
| Unbounded manual history query | A single long-range query can starve persistence or exhaust memory. | Require time bounds, cap page size and maximum range, use indexed queries, set statement timeouts, and keep history strictly user-initiated. |

## 11. Operational decisions to validate before implementation

These remain configuration/benchmark decisions, not unresolved product scope:

1. Choose shard count, per-shard queue capacity, persistence-channel capacity, and pgx pool size by load test on the target host.
2. Set short Redis and PostgreSQL operation timeouts and a graceful-drain deadline, then test failure and shutdown behavior.
3. Confirm whether one-second aggregate-bucket granularity is acceptable to operators; otherwise budget exact per-event window state.
4. Define demo data-retention duration and PostgreSQL backup/restore cadence before collecting sustained benchmark traffic.
5. Choose the initial dashboard polling interval between 2 and 5 seconds from measured Redis p95 and operator usability.
