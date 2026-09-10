BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE device_lifecycle AS ENUM ('active', 'disabled');
CREATE TYPE measurement_kind AS ENUM ('temperature', 'voltage', 'battery', 'pressure');
CREATE TYPE threshold_status AS ENUM ('normal', 'warning', 'critical');

CREATE TABLE devices (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id VARCHAR(128) NOT NULL UNIQUE,
  name VARCHAR(128),
  device_type VARCHAR(64),
  lifecycle_state device_lifecycle NOT NULL DEFAULT 'active',
  last_persisted_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT devices_external_id_not_blank CHECK (length(trim(external_id)) > 0)
);

CREATE TABLE api_clients (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  key_id VARCHAR(64) NOT NULL UNIQUE,
  api_key_hash BYTEA NOT NULL,
  device_id UUID NOT NULL UNIQUE REFERENCES devices(id) ON DELETE RESTRICT,
  enabled BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT api_clients_key_id_not_blank CHECK (length(trim(key_id)) > 0),
  CONSTRAINT api_clients_hash_length CHECK (octet_length(api_key_hash) = 32)
);

CREATE TABLE measurement_thresholds (
  measurement_type measurement_kind NOT NULL,
  status threshold_status NOT NULL,
  min_value DOUBLE PRECISION,
  min_inclusive BOOLEAN NOT NULL DEFAULT true,
  max_value DOUBLE PRECISION,
  max_inclusive BOOLEAN NOT NULL DEFAULT true,
  unit VARCHAR(8) NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (measurement_type, status),
  CONSTRAINT measurement_thresholds_has_bound CHECK (min_value IS NOT NULL OR max_value IS NOT NULL),
  CONSTRAINT measurement_thresholds_valid_range CHECK (min_value IS NULL OR max_value IS NULL OR min_value <= max_value)
);

CREATE TABLE telemetry_events (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  event_id VARCHAR(128) NOT NULL,
  device_id UUID NOT NULL REFERENCES devices(id) ON DELETE RESTRICT,
  measurement_type measurement_kind NOT NULL,
  value DOUBLE PRECISION NOT NULL,
  unit VARCHAR(8) NOT NULL,
  event_timestamp TIMESTAMPTZ NOT NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  request_id UUID NOT NULL,
  CONSTRAINT telemetry_events_event_id_not_blank CHECK (length(trim(event_id)) > 0),
  CONSTRAINT telemetry_events_idempotency UNIQUE (device_id, event_id),
  CONSTRAINT telemetry_events_valid_measurement CHECK (
    (measurement_type = 'temperature' AND unit = 'C' AND value BETWEEN -40.0 AND 125.0) OR
    (measurement_type = 'voltage' AND unit = 'V' AND value BETWEEN 0.0 AND 48.0) OR
    (measurement_type = 'battery' AND unit = '%' AND value BETWEEN 0.0 AND 100.0) OR
    (measurement_type = 'pressure' AND unit = 'kPa' AND value BETWEEN 0.0 AND 1000.0)
  )
);

CREATE INDEX telemetry_events_device_measurement_time_idx ON telemetry_events (device_id, measurement_type, event_timestamp DESC);
CREATE INDEX telemetry_events_device_time_idx ON telemetry_events (device_id, event_timestamp DESC);
CREATE INDEX telemetry_events_received_at_brin_idx ON telemetry_events USING BRIN (received_at);

INSERT INTO measurement_thresholds (measurement_type, status, min_value, min_inclusive, max_value, max_inclusive, unit)
VALUES
  ('temperature', 'normal', NULL, true, 70.0, false, 'C'),
  ('temperature', 'warning', 70.0, true, 85.0, true, 'C'),
  ('temperature', 'critical', 85.0, false, NULL, true, 'C'),
  ('voltage', 'normal', 11.5, true, NULL, true, 'V'),
  ('voltage', 'warning', 10.5, true, 11.5, false, 'V'),
  ('voltage', 'critical', NULL, true, 10.5, false, 'V'),
  ('battery', 'normal', 20.0, true, NULL, true, '%'),
  ('battery', 'warning', 10.0, true, 20.0, false, '%'),
  ('battery', 'critical', NULL, true, 10.0, false, '%'),
  ('pressure', 'normal', 0.0, true, 300.0, true, 'kPa'),
  ('pressure', 'warning', 300.0, false, 700.0, true, 'kPa'),
  ('pressure', 'critical', 700.0, false, NULL, true, 'kPa');

COMMIT;
