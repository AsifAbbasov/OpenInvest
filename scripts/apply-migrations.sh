#!/usr/bin/env bash
set -euo pipefail

# This is the sole owner-only path for applying the canonical up migrations.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATION_DIR="$ROOT_DIR/infrastructure/postgres/migrations"
OWNER_DATABASE_URL="${OPENINVEST_DATABASE_OWNER_URL:-}"
export LC_ALL=C

fail(){ echo "MIGRATION_RUNNER_RESULT=FAIL" >&2; echo "MIGRATION_RUNNER_REASON=$1" >&2; exit 1; }
require_command(){ command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"; }

[[ -n "$OWNER_DATABASE_URL" ]] || fail "OPENINVEST_DATABASE_OWNER_URL is required; runtime database credentials are not accepted"
require_command go
require_command psql

echo "MIGRATION_RUNNER=CANONICAL_OWNER_ONLY"
echo "MIGRATION_SOURCE=infrastructure/postgres/migrations/*.up.sql"
( cd "$ROOT_DIR/backend-go" && go run ./cmd/validate-migrations --mode=repository )

shopt -s nullglob
migrations=("$MIGRATION_DIR"/*.up.sql)
shopt -u nullglob
[[ "${#migrations[@]}" -gt 0 ]] || fail "no canonical *.up.sql migrations discovered"

for migration in "${migrations[@]}"; do
  migration_name="$(basename "$migration")"
  echo "MIGRATION_APPLY_FILE=$migration_name"
  if [[ "$migration_name" == "000009_stage_03_71_ledger_sequence_unique.up.sql" ]]; then
    ( cd "$ROOT_DIR/backend-go" && OPENINVEST_DATABASE_OWNER_URL="$OWNER_DATABASE_URL" go run ./cmd/recover-ledger-sequence-index )
  else
    psql -X -v ON_ERROR_STOP=1 --dbname="$OWNER_DATABASE_URL" --file="$migration"
  fi
done

echo "MIGRATION_RUNNER_RESULT=PASS"
