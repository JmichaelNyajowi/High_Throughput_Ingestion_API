# PRD: High-Throughput IoT Telemetry Ingestion & Live Dashboard

## 1. Problem statement — who hurts and why

Internal engineering and operations teams monitoring industrial edge-device fleets need a current, reliable view of sensor measurements including temperature, voltage, battery level, and pressure.

Handling validation, aggregation, and individual database writes synchronously makes burst ingestion slow and vulnerable to downstream outages. PostgreSQL becomes a bottleneck, requests time out, and operators must use slow raw-data queries to understand current fleet health.

This product keeps the synchronous path small—authenticate, validate, rate-limit, and enqueue—then processes aggregation and persistence asynchronously. Redis serves live derived state and PostgreSQL remains the durable raw-event system of record.

## 2. Target user + 2 personas

**Target user:** Internal teams operating fleets of IoT edge devices, gateways, and simulated ESP32 sensors.

| Persona | Job | Primary need |
|---|---|---|
| Priya, Platform Engineer | Integrates and operates device telemetry ingestion. | Fast, predictable admission with explicit validation and overload behavior. |
| Daniel, Operations Engineer | Monitors fleet health and investigates abnormal devices. | Current values, five-minute aggregates, threshold status, and device freshness. |

## 3. Goals and non-goals

### Goals

- Accept authenticated JSON batches containing 1–500 telemetry events.
- Sustain 2,000 requests/sec at 10 events/request (20,000 events/sec) on a single 4-vCPU, 8-GB RAM Ubuntu host with SSD storage, using Docker containers for API, Redis, and PostgreSQL.
- Return `202 Accepted` with p95 ingestion latency below 20 ms after admission to the in-memory queue.
- Enforce a 1 MB request-body maximum and strict measurement schemas and ranges.
- Enforce per-device/API-key rate limits of 10 event tokens/sec with a 100-event burst capacity.
- Bound memory and concurrency with a bounded Go channel, a fixed worker pool, and a 10,000-event persistence retry buffer.
- Persist raw events to PostgreSQL in bulk every 1,000 records or every 2 seconds, whichever occurs first.
- Store raw events idempotently with `UNIQUE (device_id, event_id)` and conflict-safe inserts.
- Serve current five-minute aggregates from Redis to a React dashboard with p95 read latency below 5 ms.
- Mark devices stale after 60 seconds without processed telemetry and expire Redis aggregate keys after 15 minutes.
- Keep ingestion and PostgreSQL persistence operational during Redis outages while visibly degrading the dashboard.

### Non-goals

- Crash-safe, durable acknowledgements after a `202` response.
- Exactly-once end-to-end delivery.
- External message brokers, including Kafka, RabbitMQ, or Azure Service Bus.
- API-key provisioning, rotation, revocation, RBAC, or multi-tenant isolation.
- Automated alerts, webhooks, email, or paging.
- Automated long-range historical analysis.
- Global/distributed rate limiting across multiple API instances.

## 4. User stories

- As a device, I submit a valid authenticated telemetry batch and receive a fast `202 Accepted`.
- As a device, I receive a clear client error when authentication, JSON, batch size, or measurement values are invalid.
- As a device, I receive `429 Too Many Requests` when my device rate limit is exhausted.
- As a device, I receive `503 Service Unavailable` when the service cannot safely accept more work.
- As a platform engineer, I can identify queue pressure, worker utilization, persistence failure, and retry-buffer saturation.
- As an operations engineer, I can see current telemetry, five-minute aggregates, freshness, and threshold state without live PostgreSQL queries.
- As an operations engineer, I see a persistent degraded-mode warning when Redis is unavailable.
- As a platform engineer, a client retry does not create a duplicate raw event in PostgreSQL.

## 5. Core features

1. Seeded static API-key authentication.
2. Authenticated batch telemetry ingestion API.
3. Strict schema, batch, body-size, and measurement-range validation.
4. Per-device in-memory token-bucket rate limiting.
5. Bounded Go queue and fixed-size worker pool.
6. Redis-backed, five-minute sliding live aggregates.
7. PostgreSQL bulk persistence with idempotent inserts and bounded retry.
8. React fleet and device telemetry dashboard.
9. Health checks, structured logs, and operational metrics.

## 6. Detailed functional requirements per MVP feature

### 6.1 Authentication and ingestion API

- Expose `POST /v1/telemetry/batches`.
- Require an API key or bearer token on every request. Static credential hashes and their device/client mappings are seeded through application configuration or environment variables.
- Require `Content-Type: application/json`.
- Limit request bodies to 1 MB.
- Return:
  - `202 Accepted` after authentication, validation, token-bucket admission, and successful queue enqueue.
  - `400 Bad Request` for invalid JSON, missing or invalid fields, out-of-range values, or batches containing more than 500 events.
  - `401 Unauthorized` for missing or invalid credentials.
  - `413 Payload Too Large` for request bodies over 1 MB.
  - `415 Unsupported Media Type` for non-JSON requests.
  - `429 Too Many Requests` for rate-limit exhaustion.
  - `503 Service Unavailable` when the admission queue is full, PostgreSQL retry capacity is saturated, or admission is paused for persistence backpressure.
- Include a request ID in the response, structured logs, and persisted records.

### 6.2 Payload validation

Each event requires:

```json
{
  "event_id": "string",
  "device_id": "string",
  "timestamp": "RFC 3339 UTC timestamp",
  "measurement_type": "temperature | voltage | battery | pressure",
  "value": "float64",
  "unit": "optional, type-compatible"
}
```

- Reject the entire batch if any event is invalid; no partial batch admission.
- Require a finite numeric `value`.
- Require the authenticated API key to be authorized for the submitted `device_id`.
- Enforce the following domain ranges:

| Measurement type | Unit | Accepted range |
|---|---:|---:|
| Temperature | °C | -40.0 to 125.0 |
| Voltage | V | 0.0 to 48.0 |
| Battery | % | 0.0 to 100.0 |
| Pressure | kPa | 0.0 to 1000.0 |

- Store events whose event timestamp is more than 60 seconds behind system time, but exclude them from live Redis aggregation.

### 6.3 Rate limiting

- Maintain a concurrency-safe in-memory token bucket for each API key/device.
- Refill at 10 event tokens/sec up to a maximum capacity of 100 tokens.
- Charge one token per event, not one token per request.
- Reject a request atomically with `429` when enough tokens are not available; do not enqueue any event from that request.
- Include `Retry-After` when a retry time can be calculated.
- Rate-limit state is local to an API process. The 20,000-event/sec benchmark requires at least 2,000 concurrent distinct API keys/devices, each sending 10 events/sec on average.

### 6.4 Queue and worker processing

- Use a bounded Go channel for accepted batches and a configurable fixed-size worker pool.
- Do not create an unbounded goroutine per request or event.
- Immediately return `503` if the admission queue is full.
- Each worker must:
  1. classify late events;
  2. update Redis aggregates for non-late events when Redis is available;
  3. buffer all accepted raw events for PostgreSQL persistence.
- Recover from a worker panic, log the error, and preserve configured worker capacity.
- On graceful shutdown, stop admission, drain work until a configured deadline, and attempt a final persistence flush.

### 6.5 Redis rolling aggregates and degraded mode

- Maintain a five-minute, processing-time sliding window keyed by `device_id + measurement_type`.
- For each key, store: latest value and timestamp, count, sum, average, minimum, maximum, window start/end, threshold state, and aggregate update time.
- Set a 15-minute Redis TTL on aggregate keys.
- Flag a device as `Stale/Offline` when no telemetry has been successfully processed for over 60 seconds.
- Live dashboard endpoints must read Redis-derived aggregates only; they must not execute PostgreSQL reads during automatic refresh.
- When Redis is unavailable:
  - continue accepting and persist-only processing where PostgreSQL capacity permits;
  - log cache failures asynchronously and emit Redis health metrics;
  - show the persistent banner: **“Degraded Mode: Live Aggregates Paused (Redis Offline)”**;
  - show last cached/stale data if available; do not fall back to PostgreSQL polling;
  - rebuild aggregate cache from PostgreSQL after Redis recovery.

### 6.6 Threshold evaluation

| Measurement | Normal | Warning | Critical |
|---|---|---|---|
| Temperature | `< 70.0°C` | `70.0°C–85.0°C`, inclusive | `> 85.0°C` |
| Voltage | `≥ 11.5V` | `10.5V–<11.5V` | `< 10.5V` |
| Battery | `≥ 20.0%` | `10.0%–<20.0%` | `< 10.0%` |
| Pressure | `0.0–300.0 kPa`, inclusive | `>300.0–700.0 kPa`, inclusive | `>700.0 kPa` |

- Threshold state is visual dashboard status only. It must not trigger notifications in MVP.

### 6.7 PostgreSQL bulk persistence

- Persist every accepted raw event in a PostgreSQL transaction using parameterized bulk inserts.
- Flush the persistence buffer at 1,000 records or every 2 seconds, whichever occurs first.
- Enforce `UNIQUE (device_id, event_id)` on raw telemetry.
- Insert with `ON CONFLICT (device_id, event_id) DO NOTHING` to make stored raw events idempotent during client retries.
- Retry failed flushes with exponential backoff beginning at 100 ms, doubling per retry, capped at 5 seconds, for a maximum of three retries.
- Cap the in-memory retry buffer at 10,000 events.
- If retries are exhausted or the retry buffer is full, emit structured logs and metrics and reject new requests with `503` to protect server memory.
- This system does not provide durable acknowledgement: a process crash after `202` can lose queued or buffered events that have not yet reached PostgreSQL.

### 6.8 Dashboard

- Provide a React SPA fleet overview with active, stale, warning, and critical device counts; latest measurements; five-minute aggregates; and threshold status.
- Provide a device detail view with short live charts derived from Redis state.
- Poll live aggregate endpoints at a configurable interval.
- Explicitly show loading, stale, unavailable, error, and degraded states.
- Permit historical PostgreSQL reads only through an explicit user-initiated historical view, never through live UI auto-refresh.

### 6.9 Health and observability

- Provide `/healthz` for process liveness.
- Provide `/readyz` for admission readiness and dependency status.
- Emit structured logs containing request ID, device ID, batch size, outcome, queue depth, and failure reason.
- Export metrics for received, accepted, invalid, unauthorized, rate-limited, and queue-rejected events; API latency percentiles; queue depth/capacity; worker utilization; Redis errors; persistence flush duration/size; retries/failures; and retry-buffer occupancy.

## 7. Data model sketch (entities + key fields)

| Entity | Key fields |
|---|---|
| `api_clients` | `id`, `name`, `api_key_hash`, `device_id`, `status`, `created_at` |
| `devices` | `id`, `name`, `device_type`, `status`, `last_processed_at`, `created_at` |
| `telemetry_events` | `id`, `event_id`, `device_id`, `measurement_type`, `value`, `unit`, `event_timestamp`, `received_at`, `request_id` |
| `measurement_thresholds` | `measurement_type`, `normal`, `warning`, `critical` bounds, `unit` |
| Redis aggregate | `device_id`, `measurement_type`, `latest_value`, `latest_timestamp`, `count`, `sum`, `min`, `max`, `average`, `window_start`, `window_end`, `status`, `updated_at` |

Required database design:

- Foreign key from `telemetry_events.device_id` to `devices.id`.
- `UNIQUE(device_id, event_id)` on `telemetry_events`.
- Index on `(device_id, measurement_type, event_timestamp DESC)` for rebuild and manual historical queries.

## 8. Edge cases and failure states

| Condition | Required behavior |
|---|---|
| Invalid JSON, field, measurement range, or batch over 500 events | Reject whole request with `400`; do not enqueue. |
| Empty batch | Accept only if batch cardinality permits it; **MVP decision: reject with `400` because a telemetry batch must contain at least one event.** |
| Request body over 1 MB | Return `413`; do not enqueue. |
| Invalid API key | Return `401`; do not enqueue. |
| Rate bucket exhausted | Return `429`; no partial admission. |
| Admission queue full | Return `503`; client should retry with backoff. |
| Event more than 60 seconds late | Persist raw event; exclude from Redis aggregate. |
| Duplicate event | Store once through conflict-safe PostgreSQL insert. |
| Redis unavailable | Continue persist-only processing; dashboard pauses live aggregation and shows degraded banner. |
| PostgreSQL temporarily unavailable | Retry bounded batch up to three times. |
| Retry exhausted or retry buffer full | Reject new work with `503`; emit error logs and metrics. |
| Process crash after `202` | In-memory queued/buffered data may be lost. |
| Device silent for more than 60 seconds | Display device as stale/offline. |
| Redis aggregate key expires after 15 minutes | Remove from live state until new telemetry is processed or cache is rebuilt. |

## 9. Success metrics

| Metric | MVP target |
|---|---|
| Load-test topology | At least 2,000 distinct device/API-key simulators. |
| Sustained ingestion | 2,000 requests/sec × 10 events/request = 20,000 events/sec. |
| Ingestion latency | p95 `<20 ms` through successful in-memory enqueue. |
| Dashboard read latency | p95 `<5 ms` for Redis-backed aggregate reads. |
| Persistence cadence | Flush at 1,000 records or 2 seconds. |
| Memory protection | No unbounded memory or goroutine growth during burst/overload tests. |
| Validation containment | Invalid and oversized requests return explicit errors without crashing the API or workers. |
| Live data source | No PostgreSQL reads during live dashboard auto-refresh. |
| Redis degradation | Redis failure does not block ingestion or PostgreSQL persistence when PostgreSQL is healthy. |

## 10. Open questions

The MVP's product requirements are resolved. Implementation-level decisions remain:

1. What exact bounded admission-queue capacity and fixed worker-pool size meet the stated 4-vCPU benchmark without starving PostgreSQL or Redis connections?
2. What configured shutdown drain deadline is acceptable before remaining in-memory work is abandoned?
3. What polling interval balances dashboard freshness, Redis load, and the 5 ms read-latency objective?
4. How frequently, and through what operator-controlled mechanism, should Redis aggregates be rebuilt from PostgreSQL after recovery?
