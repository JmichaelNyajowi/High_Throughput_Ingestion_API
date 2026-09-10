High-Throughput Ingestion API & Live Telemetry Dashboard

```
Project interview guide | Go, PostgreSQL, Redis, React
```
# 1. Your 60-second explanation

I built a telemetry platform for handling frequent sensor-style events without making the request path wait for aggregation or database work. The Go API validates a batch, places accepted work on a queue, and lets workers compute rolling aggregates. Redis serves the current aggregate quickly, PostgreSQL stores durable data in batches, and a React dashboard turns the latest values into charts, gauges, and alerts.

Problem: Telemetry arrives in bursts. If every HTTP request immediately performs calculation and an individual database insert, latency grows, the database becomes the bottleneck, and one slow dependency can affect incoming requests.

What you built: An API that separates intake from processing, a bounded worker-based processing path, an aggregate cache, durable batch persistence, rate limiting, and a dashboard for current telemetry state.

Why it matters: The design demonstrates a common backend principle: keep the synchronous path small, make slower work explicit, and decide which data must be immediately consistent versus eventually consistent.

# 2. Reference architecture and request flow

Use this as a verbal map. Confirm each component matches your actual implementation before an interview.

```
Device / simulator
       | HTTP telemetry batch
       v
Go API -> validation -> token-bucket rate limit -> queue
                                           |
                                           v
                              worker pool / rolling aggregation
                                  |                    |
                                  v                    v
                            Redis cache        PostgreSQL bulk write
                                  |                    |
                                  +---- React dashboard +
```
- **Admission path:**  Parse the request, validate its shape and domain constraints, apply a rate-limit decision, and acknowledge only according to the durability guarantee your code actually provides.
- **Asynchronous path:**  A worker receives queued work, computes or updates the aggregate, updates the cache, and buffers records for a bulk database insert.
- **Read path:**  The dashboard should normally read the latest aggregate from Redis; historical detail and recovery queries can read from PostgreSQL.
# 3. Why these technologies

## Go

Go is a good fit for a networked ingestion service because it has a small runtime model, strong standard library support for HTTP and concurrency, and straightforward cancellation through contexts. Explain how your goroutines are bounded by a worker pool rather than created without limit.

## PostgreSQL

PostgreSQL gives durable relational storage, SQL queries, transaction support, constraints, and a reliable record of accepted telemetry. Bulk writes reduce per-statement and network overhead compared with inserting each event independently.

## Redis

Redis is suited to serving hot, derived state such as the most recent rolling aggregate. It lowers database read pressure, but it is a cache: state must be rebuildable from durable data or treated carefully if Redis restarts.

## React

React separates display state from backend processing and is useful for a dashboard composed of reusable charts, gauges, and alert components. The browser should treat the API as the source of truth and handle loading, stale data, and errors visibly.

# 4. Core implementation details to be able to explain

## Input validation

Validation protects the rest of the pipeline from malformed or unreasonable work.

Validate required identifiers, timestamps, measurement fields, types, and numeric ranges that are meaningful for the sensor domain.

Reject oversized batches and malformed JSON before adding work to the queue.

Return clear 4xx errors for client mistakes; reserve 5xx responses for server failures.

## Queue and worker pool

The queue absorbs short bursts while a fixed number of workers limits concurrent CPU and I/O work.

Use a bounded channel or another bounded queue so memory cannot grow without limit.

Choose worker count based on whether the work is CPU-bound, I/O-bound, and on the connection pool size.

Define the overload behavior: reject, block briefly, or shed lower-priority work. Be precise about which behavior your code uses.

## Rolling aggregates

An aggregate such as count, sum, minimum, maximum, or moving average avoids scanning all raw events on each dashboard read.

Explain the aggregation window: event-time versus processing-time, fixed window versus sliding window, and how an event is assigned.

Keep aggregate updates safe when multiple workers can process the same key. A mutex, key partitioning, atomic operation, or single-owner worker are possible approaches—describe the one you used.

State how late, duplicate, or out-of-order telemetry is handled. If it is not handled, identify it as a limitation.

## Bulk persistence

Batching amortizes database round trips and transaction overhead.

Use parameterized SQL and a transaction for the batch where appropriate.

Flush based on record count, elapsed time, or both so low traffic does not wait forever.

Explain what happens if the database write fails: retry policy, backoff, dead-letter handling, and whether data can be lost or duplicated.

## Rate limiting

A token bucket lets average usage be controlled while allowing a limited burst.

Tokens refill at a configured rate up to a capacity; each request consumes tokens proportional to its cost.

The bucket capacity controls burst tolerance; the refill rate controls sustained throughput.

Explain whether limiting is per device, API key, IP address, or globally, and how state is stored if there are multiple API instances.

# 5. Correctness, reliability, security, and scale

## Delivery semantics

Be exact about the guarantee. An HTTP 2xx does not automatically mean durable storage.

If the API acknowledges after enqueueing in memory, a process crash can lose queued work; call this at-most-once after acknowledgement unless you persisted first.

If producers retry after a timeout, duplicates are possible. Use an event ID plus a uniqueness constraint or deduplication key if implemented.

At-least-once delivery generally requires idempotent processing; exactly-once is much harder and usually expressed as an effective outcome, not a simple transport property.

## Cache consistency

Redis makes reads fast but introduces staleness and recovery questions.

Say whether cache is updated before or after the database write and why.

Use TTLs and rebuild logic if the cached aggregate can be recomputed from PostgreSQL.

Do not describe cached data as strongly consistent unless the design actually enforces it.

## Security

Protect the ingestion endpoint and avoid trusting device input.

Authenticate callers using the mechanism you implemented; do not claim mTLS or signed payloads unless present.

Store secrets outside source control and rotate them with an operational process.

Apply request-size limits, timeouts, and structured error handling to reduce abuse and resource exhaustion.

# 6. Trade-offs and alternatives

Why asynchronous processing?: It improves API responsiveness under bursts, but adds queueing delay and makes failure handling more complex. A direct write path is simpler and may be right for low-volume data.

Why Redis instead of querying PostgreSQL for every chart refresh?: Redis makes hot aggregate reads cheap, but duplicates derived state. PostgreSQL remains the durable system of record.

Why an in-memory queue?: It is simple for a project and useful for demonstrating concurrency, but it does not survive process crashes. A durable broker such as Azure Service Bus, Kafka, RabbitMQ, or a database-backed outbox is a next step when durable buffering is required.

Why bulk writes?: They raise throughput but increase the possible loss window before a flush and can increase latency for an individual event. The flush policy balances those concerns.

# 7. Testing, observability, and production thinking

Unit-test validation with valid, malformed, empty, oversized, and boundary-value batches.

Test token-bucket behavior with a controllable clock: initial burst, depletion, refill, and concurrent access.

Integration-test PostgreSQL persistence against a temporary database; verify parameterized queries, transaction rollback, and constraints.

Load-test with realistic burst sizes and record p50/p95/p99 ingestion latency, queue depth, worker utilization, errors, and database flush duration.

Add health/readiness checks and metrics for accepted, rejected, queued, processed, failed, and retried events.

If you did not implement one of these, frame it as a next step rather than a completed feature. Interviewers value an honest engineering plan more than invented operational detail.

# 8. Probable interviewer questions

**Q: Why did you choose a worker pool?**

Answer shape: It bounds concurrency and protects downstream systems. I wanted a predictable maximum number of processing tasks rather than unlimited goroutines under a burst.

**Likely follow-up: How did you choose the worker count, and what happens when the queue is full?**

**Q: What does sub-millisecond Redis read latency mean?**

Answer shape: Only claim a measured result with the workload, data size, and environment. Otherwise say Redis was selected to reduce hot-read latency and database load, and explain how you would benchmark it.

**Likely follow-up: Where was the measurement taken—network included or only local process time?**

**Q: How do you prevent duplicate telemetry?**

Answer shape: Describe your actual event identifier and uniqueness or idempotency strategy. If none exists, say duplicate handling is a planned enhancement.

**Likely follow-up: What happens if the client retries after receiving a timeout?**

**Q: What happens when PostgreSQL is unavailable?**

Answer shape: Explain the exact implementation. A robust design would back off, cap retries, expose the backlog, and use a durable queue if loss is unacceptable.

**Likely follow-up: How do you prevent retry storms or unbounded memory growth?**

**Q: Why not use Kafka or a cloud queue?**

Answer shape: For this project I chose fewer moving parts to focus on API, concurrency, aggregation, and persistence. For durable high-scale production ingestion, I would evaluate a broker based on ordering, retention, operations, and delivery requirements.

**Likely follow-up: Which properties would make Azure Service Bus or Event Hubs a better fit?**

**Q: How are rolling windows computed?**

Answer shape: Define the window and key precisely, then explain storage, expiry, and out-of-order handling. Do not hand-wave the term rolling aggregate.

**Likely follow-up: What happens to a late event?**

# 9. Final pre-interview evidence checklist

Open the repository and trace one telemetry request from handler to final storage.

Write down queue capacity, worker count, database flush condition, rate-limit values, and each actual HTTP status code.

Run the dashboard and know which endpoint supplies each chart or gauge.

Prepare one honest limitation and one specific improvement: durable queue, deduplication, observability, or autoscaling.

If you quote a metric, retain the command, test data, machine, and result that produced it.



| Truthfulness rule: This guide is derived from your résumé. Use only details you actually implemented and can show in the codebase. Where a choice, metric, deployment detail, or test is not in your project, state that plainly and explain what you would add next. |




| Interview framing: Start with the user or system problem, then describe your design decision, then finish with a concrete reliability or correctness benefit. Do not begin by listing technologies. |




| Strong answer pattern: Name the option, explain the trade-off in this system, state why your current choice fit the project’s scope, then describe the signal that would make you change it. |




| Close with ownership: The key message is not “I used Redis and Go.” It is “I designed the fast path, slow path, and failure boundaries deliberately so ingestion remains responsive while data is processed and stored safely.” |

