# Telemetry implementation backlog

All feature work is planned. `foundation` established the bootstrap outside this backlog. Every numbered file follows the required ticket contract and names its direct blockers.

## Dependency order

```text
01 shell
02 contract → 03 runtime → 04 credentials → 05 authentication → 06 validation → 07 rate limit
03 runtime → 08 queue
02 + 05 + 06 + 07 + 08 → 09 admission
08 + 09 → 10 persistence → 11 retry/backpressure
08 + 09 → 12 aggregates → 13 Redis state → 14 degraded/rebuild
09 + 10 + 11 + 13 + 14 → 15 operations signals
02 + 13 + 14 + 15 → 16 live reads
02 + 10 + 15 → 17 manual history
01 + 16 → 18 Fleet UI and 19 Device UI
01 + 17 → 20 History UI
18 + 19 + 20 → 21 UI hardening
09 through 17 → 22 backend suite → 23 load/failure evidence
15 + 21 + 23 → 24 deployment packaging
18 through 24 → 25 acceptance demonstration
```

## Execution gate

The project owner requires an explicit approval after every completed ticket before the next ticket starts. This is stricter than the dependency graph: do not start any later ticket, even if unblocked, until that approval is given. Ticket 01 is the first candidate.
