#!/usr/bin/env sh
set -eu

: "${POSTGRES_URL:?POSTGRES_URL must be set}"
for migration in migrations/*.up.sql; do
  psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -f "$migration"
done
