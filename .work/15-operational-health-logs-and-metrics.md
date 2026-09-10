# 15 — Health, readiness, structured logs, and metrics

**Type:** plumbing (test-after)
**Blocked by:** 09, 10, 11, 13, 14 — operational signals must reflect real runtime states.
**Status:** planned

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

