# Product Scope

## V1 destination

Internal engineering and operations teams can submit authenticated industrial IoT telemetry in bounded HTTP batches, view current Redis-backed fleet/device health, and manually inspect bounded raw-event history without compromising the ingestion path.

## In scope

- Static, per-device API-key authentication for device ingestion.
- Batch JSON ingestion of 1–500 telemetry events with strict validation, 1 MB request limit, rate limiting, and explicit overload responses.
- Bounded in-memory admission queue, fixed processing workers, rolling five-minute aggregates, Redis live state, and PostgreSQL raw-event persistence.
- Idempotent raw-event storage by `(device_id, event_id)`.
- React Fleet overview, Device detail, and manual Device history screens.
- Redis-degraded persist-only operation and clear dashboard degradation state.
- Health/readiness endpoints, structured logs, Prometheus metrics, Docker Compose deployment, and benchmark evidence.

## Out of scope

- Crash-safe acknowledgement, durable queueing, exactly-once processing, or external brokers.
- API-key provisioning UI, rotation endpoint, dynamic revocation, RBAC, tenant isolation, or self-service users.
- Automated alerts, webhooks, email, paging, or incident management.
- Automatic long-range historical analytics or live PostgreSQL dashboard refresh.
- Distributed/global rate limiting and multi-instance API scaling.
- Kubernetes, Kafka, RabbitMQ, GraphQL, an ORM, and a Node.js production server.

## Feature areas

| Feature | Definition of done | Status |
|---|---|---|
| Application shell | Operators can navigate stable Fleet, Device, and History destinations with approved pending/empty states. | Planned |
| Device ingestion | A valid authenticated device batch is accepted or rejected with the documented contract and bounded admission behavior. | Planned |
| Telemetry processing | Accepted events persist in batches and maintain bounded five-minute aggregate state. | Planned |
| Fleet monitoring | Operators can triage Redis-backed fleet/device state, including stale and degraded conditions. | Planned |
| Manual history | Operators can run an explicit bounded history query for one device. | Planned |
| Operations and demo | The Compose deployment, metrics, tests, and benchmark evidence demonstrate the MVP safely. | Planned |
