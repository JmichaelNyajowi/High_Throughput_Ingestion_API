# 15 — Health, readiness, structured logs, and metrics

**Type:** plumbing (test-after)
**Blocked by:** 09, 10, 11, 13, 14 — operational signals must reflect real runtime states.
**Status:** done

## What this delivers

Operators and monitoring can distinguish process liveness, safe admission, queue/retry pressure, persistence failures, and Redis degradation without secret leakage.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Internal monitoring reads health, readiness, metrics, and structured logs.
- **Update:** Measurements update as runtime events occur.
- **Delete:** Metrics expire/restart with the process; logs follow platform retention.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] `/healthz` reports process liveness only; `/readyz` accurately exposes safe admission, queue/backpressure, PostgreSQL, and Redis/degraded state.
- [ ] `/metrics` includes latency/outcomes, queue capacity/depth, workers, Redis errors, flushes, retries, and retry-buffer occupancy.
- [ ] Structured logs include request ID, device ID, batch size, outcome, and reason without raw API keys or complete payloads.
- [ ] Metrics/health network exposure follows the internal-only deployment policy.

## Out of scope

- Grafana dashboards, paging integrations, SLO enforcement, or external log platform selection.

## Verification

- Run endpoint, redaction, and metric cardinality tests; inspect an internal-only Compose route when available.

## Build-gate trace and review

Approved (2026-09-12). `/healthz` remains process-only. `/readyz` reports admission/backpressure, a short-timeout PostgreSQL probe, and Redis `live`/`degraded` mode; only unsafe admission or PostgreSQL produces `503`. `/metrics` is Prometheus text and exports bounded, label-free queue, worker, Redis-error, persistence-flush, and retry-occupancy gauges. Admission logs record request ID, authenticated device ID, batch size, outcome, and safe reason only; request middleware tests prove API keys, payload bodies, and query credentials never appear. Caddy blocks `/healthz`, `/readyz`, and `/metrics` at the public edge, while Compose exposes the API only on the internal network. Full Docker-backed race tests, vet, formatting, and diff checks pass.
