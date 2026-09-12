# Load evidence

Run k6 from an isolated generator host, never from the 4-vCPU system under test. Supply generated seeded device credentials through the untracked environment only.

```bash
k6 run --summary-export results/smoke.json load/k6/ingestion.js
SCENARIO=baseline k6 run --summary-export results/baseline.json load/k6/ingestion.js
```

Run and retain separate summaries for smoke, baseline (2,000 device keys / 2,000 requests per second / 10 events per request), single-device rate-limit, queue/retry saturation, Redis outage, PostgreSQL outage, and soak. For dependency failures, stop only the named dependency container while the generator continues; do not run the generator on the SUT. Record generator/SUT CPU, memory, topology, and status breakdown beside each JSON result. A p95 admission result over 20 ms is a capacity gap, not a pass.

| Scenario | Command / fault | Expected evidence |
| --- | --- | --- |
| `smoke` | default k6 invocation | `202` admission baseline |
| `baseline` | `SCENARIO=baseline` with 2,000 seeded keys | p95, 2,000 req/s, error split |
| `single-device-rate-limit` | one seeded key | `429` without process instability |
| queue/retry | saturate dependency/retry capacity | bounded `503`, queue/retry metrics |
| Redis outage | stop Redis | persistence continues, live mode degraded |
| PostgreSQL outage | stop PostgreSQL | retries then bounded `503` |
| soak | long baseline run | CPU/memory and stable error split |
