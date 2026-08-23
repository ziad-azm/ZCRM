#!/usr/bin/env bash
# Runs golang-migrate against the local ZCRM database.
#   ./scripts/migrate.sh up
#   ./scripts/migrate.sh down 1
set -euo pipefail

DSN="${DATABASE_URL:-postgres://zcrm:zcrm@localhost:5432/zcrm?sslmode=disable}"
MIGRATIONS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../migrations" && pwd)"

exec migrate -path "$MIGRATIONS_DIR" -database "$DSN" "$@"
