# Delivery Roadmap

## Stage 0 — Foundation bootstrap

Establish the codebase, dependency/tooling baseline, Compose topology, CI gates, migrations, seed data, component tooling, and setup/release documents. This is performed by the foundation workflow before feature implementation.

## Stage 1 — Application shell

Deliver stable, accessible operator navigation and all planned destinations in their approved pending/empty state. This is the first ticketing feature and enables later vertical slices to wire real data into known screen slots.

## Stage 2 — Device ingestion

Deliver authenticated, validated, rate-limited, bounded telemetry admission with the documented HTTP contract.

## Stage 3 — Telemetry processing

Deliver raw-event batch persistence, idempotency, retries/backpressure, rolling aggregates, Redis publication, and degraded/rebuild behavior.

## Stage 4 — Fleet monitoring

Deliver Redis-backed live fleet/device APIs and wire them into Fleet and Device detail destinations.

## Stage 5 — Manual history

Deliver the bounded PostgreSQL history API and wire it into the explicit Device history workflow.

## Stage 6 — Operations and acceptance

Deliver health/metrics, secure deployment configuration, integration/E2E/load/failure testing, and the MVP acceptance demonstration.

## Current next feature

**Application shell.** It is first because every later feature needs stable routes and approved empty/loading states. It is blocked only by Stage 0 foundation bootstrap.
