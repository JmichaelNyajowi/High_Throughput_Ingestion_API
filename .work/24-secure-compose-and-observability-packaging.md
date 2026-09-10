# 24 — Secure Compose and observability packaging

**Type:** plumbing (test-after)
**Blocked by:** 15, 21, 23 — runtime signals, finished UI, and capacity evidence must be available.
**Status:** planned

## What this delivers

The demo host can deploy the reviewed system with internal-only data services, TLS-ready Caddy access controls, Prometheus/Grafana visibility, backups, and operational runbooks.

## Lifecycle

- **Create:** Deployment creates ephemeral containers and named data volumes.
- **Read:** Operators read dashboards, logs, backups, and protected service endpoints.
- **Update:** Image/config changes use immutable reviewed releases.
- **Delete:** Containers may be replaced; PostgreSQL volume deletion is prohibited outside an explicit recovery procedure.
- **Undo:** Roll back application images by immutable digest; schema changes use forward-safe migrations.

## Acceptance criteria

- [ ] Compose exposes only Caddy externally; PostgreSQL, Redis, metrics, and API network policies match architecture documentation.
- [ ] Caddy requires TLS/domain configuration for external use and protects all operator SPA/API destinations separately from device ingestion.
- [ ] Prometheus/Grafana services, scrape configuration, health checks, non-root images, pinned versions, backup/restore instructions, and systemd runbook are present.
- [ ] Deployment configuration reads secrets only from secure environment injection and never commits populated values.

## Out of scope

- Kubernetes, multi-host HA, centralized IAM, automated paging, or a public dashboard.

## Verification

- Validate Compose, scan images, deploy to a Docker-capable staging host, test backup/restore, and run documented smoke checks.

