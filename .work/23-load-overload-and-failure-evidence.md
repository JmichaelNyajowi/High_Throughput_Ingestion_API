# 23 — Load, overload, and dependency-failure evidence

**Type:** logic (test-first)
**Blocked by:** 15, 22 — observable runtime and backend integration suite must exist.
**Status:** planned

## What this delivers

The team has repeatable k6 evidence that the MVP meets its admission, overload, and degraded-mode claims on the stated benchmark topology.

## Lifecycle

- **Create:** Not applicable.
- **Read:** k6 reads API responses and metrics; reports read benchmark outputs.
- **Update:** Scenario parameters are versioned benchmark configuration.
- **Delete:** Generated benchmark artifacts follow documented retention.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Scenarios cover smoke, 2,000 devices/2,000 req/s at 10 events/batch, single-device `429`, queue/retry `503`, Redis outage, PostgreSQL outage, and soak.
- [ ] Results report admission p95, error breakdown, queue/retry behavior, memory/CPU observations, and actual topology.
- [ ] A valid baseline demonstrates p95 admission under 20 ms or records a specific documented capacity gap.
- [ ] Tests run from an isolated generator so load generation does not invalidate the 4-vCPU system-under-test result.

## Out of scope

- Claims of unlimited throughput, multi-instance scaling, or long-term production capacity certification.

## Verification

- Run each scenario on a Docker-capable isolated host and retain the machine-readable summary plus dashboard evidence.

