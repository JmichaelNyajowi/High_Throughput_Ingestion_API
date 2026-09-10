# Release Process

## Environments and ownership

- **Local:** developer machine with Compose; never uses production credentials.
- **Demo/staging:** single Ubuntu host running the same pinned Compose topology.
- **Production:** out of MVP until an operator explicitly approves its use; the demo deployment must still follow this process.

Only reviewed pull requests may merge to `main`. CI builds immutable tagged images; the same image digest is promoted to the demo host. Do not rebuild during promotion.

For an internet-reachable deployment, set `CADDY_SITE_ADDRESS` to the real DNS hostname, publish both ports 80 and 443, and let Caddy obtain and renew TLS certificates. The `:80` example is local development only and is not an acceptable external deployment setting.

## Required release sequence

1. Pull request passes formatting, static checks, Go tests with race detection, frontend type checks/tests, and the ticket’s verification evidence.
2. An independent `/review` approves and merges the ticket.
3. CI builds and scans images, then publishes immutable image tags.
4. The release owner deploys the approved tag with `docker compose -f deploy/compose.yaml --env-file /secure/telemetry.env up -d` on the target host.
5. Run backward-compatible migrations before enabling code that needs new schema fields.
6. Check `/healthz`, `/readyz`, protected dashboard access, and API metrics reachability from the internal monitoring network.
7. Record image tag/digest, migration version, deployment time, and smoke-check result in the release log or CI deployment record.

## Migration and rollback rules

- Use expand/contract migrations only. Add compatible objects first; remove old ones only in a later release.
- Take and verify a PostgreSQL backup before a migration on the demo or production host.
- Roll back application images to the previous immutable digest when smoke checks fail. Do not roll back a completed schema migration by deleting production data; use a forward corrective migration unless a tested rollback exists.
- A failed readiness check, unavailable database, cache outage that breaks the declared degraded behavior, or failed smoke check blocks release and must be surfaced immediately.
