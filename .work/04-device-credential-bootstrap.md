# 04 — Seeded devices and static credentials

**Type:** logic (test-first)
**Blocked by:** 03 — use the validated configuration and service lifecycle.
**Status:** planned

## What this delivers

Deployers can seed active devices and one static API credential per device without persisting, returning, or logging plaintext secrets.

## Lifecycle

- **Create:** Controlled bootstrap inserts an active device and its keyed-HMAC credential record.
- **Read:** Authentication performs key-ID lookup; operators have no credential UI.
- **Update:** Credentials are replaced by controlled configuration/bootstrap, never edited in the dashboard.
- **Delete:** Disabled or removed only by controlled deployment/bootstrap; database deletion is not an MVP UI action.
- **Undo:** Restore a prior controlled configuration/credential record through deployment procedure.

## Acceptance criteria

- [ ] Structured keys use `tk_<key_id>_<secret>` and map one enabled credential to one active device.
- [ ] Bootstrap stores only a 32-byte HMAC-SHA-256 secret hash and rejects plaintext storage or log output.
- [ ] Missing, malformed, unknown, disabled, and inactive-device credentials yield the same generic unauthorized result.
- [ ] Tests exercise bootstrap validation against the real database schema.

## Out of scope

- Dashboard credential management, rotation endpoint, dynamic revocation, tenancy, or RBAC.

## Verification

- Run database integration tests and secret-redaction/log assertions.

