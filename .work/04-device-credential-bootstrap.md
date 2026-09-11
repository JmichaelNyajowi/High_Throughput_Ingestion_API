# 04 — Seeded devices and static credentials

**Type:** logic (test-first)
**Blocked by:** 03 — use the validated configuration and service lifecycle.
**Status:** done

## What this delivers

Deployers can seed active devices and one static API credential per device without persisting, returning, or logging plaintext secrets.

## Lifecycle

- **Create:** Controlled bootstrap inserts an active device and its keyed-HMAC credential record.
- **Read:** Authentication performs key-ID lookup; operators have no credential UI.
- **Update:** Credentials are replaced by controlled configuration/bootstrap, never edited in the dashboard.
- **Delete:** Disabled or removed only by controlled deployment/bootstrap; database deletion is not an MVP UI action.
- **Undo:** Restore a prior controlled configuration/credential record through deployment procedure.

## Acceptance criteria

- [x] Structured keys use `tk_<key_id>_<secret>` and map one enabled credential to one active device.
- [x] Bootstrap stores only a 32-byte HMAC-SHA-256 secret hash and rejects plaintext storage or log output.
- [x] Missing, malformed, unknown, disabled, and inactive-device credentials yield the same generic unauthorized result.
- [x] Tests exercise bootstrap validation against the real database schema.

## Out of scope

- Dashboard credential management, rotation endpoint, dynamic revocation, tenancy, or RBAC.

## Verification

- [x] Test-first evidence: structured-key, seed-configuration, bootstrap, and generic-unauthorized resolver tests first failed because the credential parser, configuration, HMAC persistence, and resolver did not exist. They pass after the minimum implementation.
- [x] `GOCACHE=/tmp/telemetry-go-cache GOMODCACHE=/tmp/telemetry-go-mod-cache go test -race ./...`, `GOCACHE=/tmp/telemetry-go-cache GOMODCACHE=/tmp/telemetry-go-mod-cache go vet ./...`, `test -z "$(gofmt -l .)"`, and `git diff --check` pass.
- [x] `TestBootstrapPersistsOnlyHMACCredentialRecordsAgainstRealSchema` runs the real `migrations/000001_init.up.sql` against a PostgreSQL 17 Testcontainers instance and verifies active/enabled binding, 32-byte HMAC-only persistence, generic unknown/disabled/inactive outcomes, and controlled credential replacement. It is automatically skipped when Docker is unavailable; this development environment reported that documented skip.
- [x] The seed command logs only static outcome messages and seeded count. Bootstrap and resolver errors do not include raw keys or secrets; malformed-key regression coverage asserts that a plaintext value is not echoed.
- [x] Docker/Caddy runtime validation is unavailable in this environment because Docker and Caddy are not installed. The Compose change is limited to injecting the untracked `TELEMETRY_DEVICE_SEEDS` value into the one-shot `/seed` command; it remains a release-host validation.
- [x] `TestBootstrapDocumentationUsesExecutableCommands` was added red for the rejected local and Compose instructions. It now guards the module-aware local seed/API commands and the explicit Compose `/seed` entrypoint override; `TELEMETRY_DEVICE_SEEDS=[] go -C api run ./cmd/seed` executed successfully from the repository root.

### Build-gate trace

| Requirement or binding rule | Governing source | Implementation location | Repeatable proof |
| --- | --- | --- | --- |
| Structured `tk_<key_id>_<32-byte-base64url-secret>` credentials and one active credential per active device | This ticket; `docs/Architecture.md` section 6.1; `docs/database.md` | `api/internal/ingestion/credentials/key.go`, `bootstrap.go` | Structured-key unit test; PostgreSQL integration test validates the `devices`/`api_clients` one-to-one schema binding and controlled replacement. |
| HMAC-SHA-256 storage only; no secret disclosure | This ticket; `docs/Techstack.md` authentication rules; `docs/conventions.md` backend rules | `bootstrap.go`, `cmd/seed/main.go`, `api_clients.api_key_hash` migration | Integration test compares the stored 32-byte digest with the expected HMAC; malformed-seed test proves plaintext is not included in an error; seed command emits only count/static messages. |
| Uniform generic outcome for credential rejection states | This ticket; `docs/Architecture.md` sections 6.1 and 9 | `resolve.go` | Unit test covers missing/malformed keys; real-schema integration test covers unknown, disabled, and inactive-device records. No HTTP route is added here; Ticket 05 maps this internal result to its `401` handler response. |
| Controlled, environment-backed deployment bootstrap | This ticket; `docs/prd.md` section 6.1; `docs/setup.md` | `config.go`, `cmd/seed/main.go`, `api/Dockerfile`, `deploy/compose.yaml` | Configuration parsing tests, seed-command compilation, executable-command regression test, successful local seed-command invocation, and documented manual/Compose seed procedure. |

## Review findings

1. **Resolved — executable bootstrap commands.** `TestBootstrapDocumentationUsesExecutableCommands` first failed against the invalid instructions. `docs/setup.md` now uses `go -C api run ./cmd/seed` and `go -C api run ./cmd/api`, and its Compose command explicitly uses `--entrypoint /seed`. The local seed command was executed successfully from the repository root with an empty seed list. Docker remains unavailable, so the regression test statically verifies the Compose command against the actual `/api` entrypoint and bundled `/seed` binary.
2. **Approved — independent re-review found no remaining Ticket 04 discrepancy.** The rebuilt command test passes under the race detector and vet; the local seed command reaches the correct module and completes safely with no configured seeds. The broader Go quality gate passed before this documentation-only correction. Docker/Caddy runtime validation remains a release-host check because Docker and Caddy are unavailable in this environment. No merge or release action was performed.
