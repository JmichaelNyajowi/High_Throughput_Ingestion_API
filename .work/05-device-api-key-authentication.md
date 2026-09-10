# 05 — Device API-key authentication

**Type:** logic (test-first)
**Blocked by:** 04 — seeded device credentials must exist.
**Status:** planned

## What this delivers

Only a valid enabled device credential can enter the ingestion path, and each accepted request carries exactly one authorized device identity.

## Lifecycle

- **Create:** Not applicable; credential creation is ticket 04.
- **Read:** Middleware resolves only the credential/device binding needed by ingestion.
- **Update:** Not applicable.
- **Delete:** Not applicable.
- **Undo:** Not applicable.

## Acceptance criteria

- [ ] `X-API-Key` parsing validates the structured key and uses constant-time comparison of its HMAC hash.
- [ ] Missing, malformed, disabled, or invalid keys return `401` before queue activity.
- [ ] Valid authentication makes internal and external device identity available to downstream admission only.
- [ ] A device credential cannot authorize live, history, metrics, or administrative routes.

## Out of scope

- TLS termination configuration, dashboard authentication, rate limiting, and payload validation.

## Verification

- Write failing authentication cases first; run handler and real-database integration tests.

