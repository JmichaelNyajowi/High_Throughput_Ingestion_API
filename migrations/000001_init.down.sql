BEGIN;
DROP TABLE IF EXISTS telemetry_events;
DROP TABLE IF EXISTS measurement_thresholds;
DROP TABLE IF EXISTS api_clients;
DROP TABLE IF EXISTS devices;
DROP TYPE IF EXISTS threshold_status;
DROP TYPE IF EXISTS measurement_kind;
DROP TYPE IF EXISTS device_lifecycle;
COMMIT;
