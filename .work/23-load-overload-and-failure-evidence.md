# 23 — Load, overload, and dependency-failure evidence

**Type:** logic (test-first)
**Blocked by:** 15, 22 — observable runtime and backend integration suite must exist.
**Status:** in-progress

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

## Review findings — rejected

1. **Blocker — no benchmark evidence exists.** `load/k6/ingestion.js` and `load/README.md` define a starting harness, but `load/results/` contains no machine-readable summaries or dashboard/host observations for smoke, baseline, rate limit, queue/retry saturation, Redis outage, PostgreSQL outage, or soak.
2. **Blocker — the required capacity claim cannot be evaluated.** No retained baseline result reports admission p95, status breakdown, CPU, memory, queue/retry metrics, or the stated 4-vCPU/8-GB topology. Ticket 23 must record either a measured p95 below 20 ms or a specific capacity gap.
3. **Blocker — isolated generator evidence is absent.** The runbook correctly requires a separate generator, but no run proves that topology. Same-host runs may be retained as functional smoke checks only and must not satisfy the benchmark acceptance criterion.

## Deferred verification decision

The project owner has deferred Ticket 23's isolated-generator benchmark until an additional generator machine is available. The versioned k6 harness and runbook remain in the repository, but this ticket is not complete and must be resumed before any release or MVP acceptance claim. Same-host checks may be recorded only as functional smoke evidence; they cannot close the 2,000 req/s capacity or p95 admission-latency criteria.
