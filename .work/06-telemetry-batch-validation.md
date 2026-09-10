# 06 — Whole-batch telemetry validation

**Type:** logic (test-first)
**Blocked by:** 02, 05 — validation follows the contract and authenticated device binding.
**Status:** planned

## What this delivers

Malformed, oversized, cross-device, or out-of-range telemetry never reaches a worker; a valid request becomes an all-or-nothing validated batch.

## Lifecycle

- **Create:** Not applicable.
- **Read:** Not applicable.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] Non-JSON returns `415`; bodies over 1 MB return `413` before queue activity.
- [ ] Empty, over-500, malformed, non-finite, wrong-unit, invalid-timestamp, cross-device, or range-invalid batches return `400` as a whole.
- [ ] Exact temperature, voltage, battery, and pressure acceptance bounds match the PRD.
- [ ] Events over 60 seconds late are valid for persistence but explicitly marked to skip live aggregation.

## Out of scope

- Rate limiting, enqueueing, persistence, and partial successful batches.

## Verification

- Write table-driven failing tests for every boundary and verify no rejection invokes the queue seam.

