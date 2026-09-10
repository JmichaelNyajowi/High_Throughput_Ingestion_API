# Database Design

## Decision

The approved stack uses Go with `pgx`, explicit bulk inserts, and `ON CONFLICT DO NOTHING`; an ORM would add abstraction without improving the high-throughput write path.

PostgreSQL holds durable raw telemetry. Redis holds disposable live aggregate state and is not represented as a PostgreSQL table.

## Tables and relationships

```text
devices 1 ─── 1 api_clients
devices 1 ─── * telemetry_events
measurement_type 1 ─── 3 measurement_thresholds
```

| Table                    | Purpose                                                                        |
| ------------------------ | ------------------------------------------------------------------------------ |
| `devices`                | Canonical identity and lifecycle state for each physical or simulated device.  |
| `api_clients`            | One static API credential per device for MVP; plaintext keys are never stored. |
| `telemetry_events`       | Immutable, idempotent raw telemetry system of record.                          |
| `measurement_thresholds` | Configurable normal, warning, and critical measurement bands.                  |

## Design choices

- `devices.id` is an internal UUID. The ingestion payload `device_id` maps to the unique public `devices.external_id`.
- `api_clients.device_id` is unique because MVP allows exactly one active static API key per device.
- `telemetry_events` enforces `UNIQUE (device_id, event_id)` to make client retry storms idempotent.
- Store canonical units only: `C`, `V`, `%`, and `kPa`. The API normalizes compatible client units before persistence.
- Do not create a `batches` or `ingestion_requests` table. At 2,000 requests/sec it adds unnecessary write amplification; `request_id` exists on every raw event and in structured logs.
- Dynamic live state such as stale/offline condition is derived from Redis processing timestamps. It is not continuously written to PostgreSQL.
- The `last_persisted_at` device field may be updated asynchronously after a successful flush, but is not part of the live dashboard read path.

## Fields, keys, constraints, and indexes

### `devices`

| Field                      | Type           | Rules                                                                  |
| -------------------------- | -------------- | ---------------------------------------------------------------------- |
| `id`                       | UUID           | Primary key.                                                           |
| `external_id`              | `varchar(128)` | Public device ID; unique and non-blank.                                |
| `name`                     | `varchar(128)` | Optional operator-facing name.                                         |
| `device_type`              | `varchar(64)`  | Optional model/family.                                                 |
| `lifecycle_state`          | enum           | `active` or `disabled`; this is administrative state, not live health. |
| `last_persisted_at`        | timestamptz    | Optional asynchronous durability marker.                               |
| `created_at`, `updated_at` | timestamptz    | Audit timestamps.                                                      |

### `api_clients`

| Field          | Type          | Rules                                                            |
| -------------- | ------------- | ---------------------------------------------------------------- |
| `id`           | UUID          | Primary key.                                                     |
| `key_id`       | `varchar(64)` | Unique non-secret identifier from `tk_<key_id>_<secret>`.        |
| `api_key_hash` | `bytea`       | 32-byte HMAC-SHA-256 result; no plaintext secret.                |
| `device_id`    | UUID          | Unique foreign key to `devices.id`.                              |
| `enabled`      | boolean       | Allows a credential to be disabled at deployment/bootstrap time. |
| `created_at`   | timestamptz   | Audit timestamp.                                                 |

### `telemetry_events`

| Field              | Type             | Rules                                                            |
| ------------------ | ---------------- | ---------------------------------------------------------------- |
| `id`               | bigint identity  | Primary key optimized for append-heavy inserts.                  |
| `event_id`         | `varchar(128)`   | Client-generated idempotency ID; unique within device.           |
| `device_id`        | UUID             | Foreign key to `devices.id`.                                     |
| `measurement_type` | enum             | Temperature, voltage, battery, or pressure.                      |
| `value`            | double precision | Validated again with metric-specific database check constraints. |
| `unit`             | `varchar(8)`     | Canonical metric-compatible unit.                                |
| `event_timestamp`  | timestamptz      | Device-supplied source event time.                               |
| `received_at`      | timestamptz      | Server persistence time; supports recovery and audit.            |
| `request_id`       | UUID             | API/log correlation ID.                                          |

### `measurement_thresholds`

Each measurement has one row for `normal`, `warning`, and `critical` state. Nullable minimum/maximum values represent unbounded sides of a threshold. Inclusive flags retain exact boundary behavior, such as `85°C` being warning and `>85°C` being critical.

| Field                            | Type             | Rules                                                 |
| -------------------------------- | ---------------- | ----------------------------------------------------- |
| `measurement_type`               | enum             | Composite primary key with `status`.                  |
| `status`                         | enum             | `normal`, `warning`, or `critical`.                   |
| `min_value`, `max_value`         | double precision | At least one bound required; min must not exceed max. |
| `min_inclusive`, `max_inclusive` | boolean          | Exact threshold boundary behavior.                    |
| `unit`                           | `varchar(8)`     | Canonical unit.                                       |
| `updated_at`                     | timestamptz      | Threshold configuration audit timestamp.              |

### Important indexes

| Index                                                 | Purpose                                                                      |
| ----------------------------------------------------- | ---------------------------------------------------------------------------- |
| `UNIQUE (device_id, event_id)`                        | Idempotent raw-event persistence.                                            |
| `(device_id, measurement_type, event_timestamp DESC)` | Per-metric manual history and aggregate rebuild.                             |
| `(device_id, event_timestamp DESC)`                   | Manual device history across all measurement types.                          |
| `BRIN (received_at)`                                  | Low-overhead bounded recovery/cache-rebuild scans across append-only events. |
| Unique `devices.external_id` and `api_clients.key_id` | Fast device resolution and O(1) credential lookup.                           |

## PostgreSQL schema

```sql
BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE device_lifecycle AS ENUM (
  'active',
  'disabled'
);

CREATE TYPE measurement_kind AS ENUM (
  'temperature',
  'voltage',
  'battery',
  'pressure'
);

CREATE TYPE threshold_status AS ENUM (
  'normal',
  'warning',
  'critical'
);

CREATE TABLE devices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Device identifier accepted in API payloads, e.g. "device-123".
  external_id VARCHAR(128) NOT NULL UNIQUE,
  name VARCHAR(128),
  device_type VARCHAR(64),

  -- Lifecycle is administrative; live health is derived from Redis.
  lifecycle_state device_lifecycle NOT NULL DEFAULT 'active',
  last_persisted_at TIMESTAMPTZ,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT devices_external_id_not_blank
    CHECK (length(trim(external_id)) > 0)
);

CREATE TABLE api_clients (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Non-secret lookup prefix in: tk_<key_id>_<random_secret>.
  key_id VARCHAR(64) NOT NULL UNIQUE,

  -- HMAC-SHA-256 hash of the secret; plaintext keys are never persisted.
  api_key_hash BYTEA NOT NULL,

  -- MVP permits one active static key per device.
  device_id UUID NOT NULL UNIQUE
    REFERENCES devices(id)
    ON DELETE RESTRICT,

  enabled BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT api_clients_key_id_not_blank
    CHECK (length(trim(key_id)) > 0),

  CONSTRAINT api_clients_hash_length
    CHECK (octet_length(api_key_hash) = 32)
);

CREATE TABLE measurement_thresholds (
  measurement_type measurement_kind NOT NULL,
  status threshold_status NOT NULL,

  -- NULL means unbounded on that side.
  min_value DOUBLE PRECISION,
  min_inclusive BOOLEAN NOT NULL DEFAULT true,
  max_value DOUBLE PRECISION,
  max_inclusive BOOLEAN NOT NULL DEFAULT true,

  unit VARCHAR(8) NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (measurement_type, status),

  CONSTRAINT measurement_thresholds_has_bound
    CHECK (min_value IS NOT NULL OR max_value IS NOT NULL),

  CONSTRAINT measurement_thresholds_valid_range
    CHECK (
      min_value IS NULL
      OR max_value IS NULL
      OR min_value <= max_value
    )
);

CREATE TABLE telemetry_events (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

  -- Client-generated; unique only within a device.
  event_id VARCHAR(128) NOT NULL,

  -- Internal device reference resolved after API-key authentication.
  device_id UUID NOT NULL
    REFERENCES devices(id)
    ON DELETE RESTRICT,

  measurement_type measurement_kind NOT NULL,
  value DOUBLE PRECISION NOT NULL,

  -- Canonical unit after API normalization.
  unit VARCHAR(8) NOT NULL,

  -- Producer-supplied and server-received times are intentionally separate.
  event_timestamp TIMESTAMPTZ NOT NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  -- Correlates persisted events to structured logs and API response.
  request_id UUID NOT NULL,

  CONSTRAINT telemetry_events_event_id_not_blank
    CHECK (length(trim(event_id)) > 0),

  CONSTRAINT telemetry_events_idempotency
    UNIQUE (device_id, event_id),

  -- Defense in depth: the API validates these before enqueue.
  -- Metric ranges also reject NaN and infinities.
  CONSTRAINT telemetry_events_valid_measurement
    CHECK (
      (
        measurement_type = 'temperature'
        AND unit = 'C'
        AND value BETWEEN -40.0 AND 125.0
      )
      OR (
        measurement_type = 'voltage'
        AND unit = 'V'
        AND value BETWEEN 0.0 AND 48.0
      )
      OR (
        measurement_type = 'battery'
        AND unit = '%'
        AND value BETWEEN 0.0 AND 100.0
      )
      OR (
        measurement_type = 'pressure'
        AND unit = 'kPa'
        AND value BETWEEN 0.0 AND 1000.0
      )
    )
);

-- Core manual-history and cache-rebuild query path.
CREATE INDEX telemetry_events_device_measurement_time_idx
  ON telemetry_events (
    device_id,
    measurement_type,
    event_timestamp DESC
  );

-- Supports manual history across all measurement types for one device.
CREATE INDEX telemetry_events_device_time_idx
  ON telemetry_events (
    device_id,
    event_timestamp DESC
  );

-- Low-write-overhead support for bounded recovery/rebuild scans.
CREATE INDEX telemetry_events_received_at_brin_idx
  ON telemetry_events
  USING BRIN (received_at);

-- Seed exact dashboard threshold semantics.
INSERT INTO measurement_thresholds (
  measurement_type,
  status,
  min_value,
  min_inclusive,
  max_value,
  max_inclusive,
  unit
)
VALUES
  ('temperature', 'normal',   NULL, true, 70.0, false, 'C'),
  ('temperature', 'warning',  70.0, true, 85.0, true,  'C'),
  ('temperature', 'critical', 85.0, false, NULL, true, 'C'),

  ('voltage', 'normal',   11.5, true, NULL, true, 'V'),
  ('voltage', 'warning',  10.5, true, 11.5, false, 'V'),
  ('voltage', 'critical', NULL, true, 10.5, false, 'V'),

  ('battery', 'normal',   20.0, true, NULL, true, '%'),
  ('battery', 'warning',  10.0, true, 20.0, false, '%'),
  ('battery', 'critical', NULL, true, 10.0, false, '%'),

  ('pressure', 'normal',   0.0, true, 300.0, true, 'kPa'),
  ('pressure', 'warning',  300.0, false, 700.0, true, 'kPa'),
  ('pressure', 'critical', 700.0, false, NULL, true, 'kPa');

COMMIT;
```

## Redis state model

Redis remains separate from the PostgreSQL schema:

| Key                                               | Type                 |                  TTL | Purpose                                                  |
| ------------------------------------------------- | -------------------- | -------------------: | -------------------------------------------------------- |
| `telemetry:v1:aggregate:{deviceId}:{measurement}` | Hash or compact JSON |           15 minutes | Current aggregate snapshot used by live dashboard reads. |
| `telemetry:v1:device-last-seen`                   | Sorted set           | Pruned by worker/job | Active/stale fleet state based on processing time.       |
| `telemetry:v1:device-measurements:{deviceId}`     | Set                  |           15 minutes | Measurements currently available for a device.           |

## Scale boundary

At 20,000 events/sec, raw-event retention grows to 1.728 billion rows/day. Keep MVP retention deliberately short. Before sustained production retention, introduce a partitioning strategy, retention jobs, and likely a dedicated idempotency ledger: PostgreSQL partitioned tables cannot enforce global `UNIQUE(device_id, event_id)` unless the partition key is included.
