#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MIGRATION_DIR="$ROOT_DIR/infrastructure/postgres/migrations"
POSTGRES_DB="${POSTGRES_DB:-openinvest}"
POSTGRES_USER="${POSTGRES_USER:-openinvest}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-openinvest-local}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
REDIS_PORT="${REDIS_PORT:-6379}"
SKIP_INFRA="${OPENINVEST_BOOTSTRAP_SKIP_INFRA:-false}"
export POSTGRES_DB POSTGRES_USER POSTGRES_PASSWORD POSTGRES_PORT REDIS_PORT LC_ALL=C

fail(){ echo "BOOTSTRAP_RESULT=BLOCKED" >&2; echo "BOOTSTRAP_REASON=$1" >&2; exit 1; }
require_command(){ command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"; }
hash_stdin(){
  if command -v sha256sum >/dev/null 2>&1; then sha256sum | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then shasum -a 256 | awk '{print $1}'
  else fail "sha256sum or shasum is required"; fi
}
wait_for_postgres(){
  local attempt
  for attempt in $(seq 1 60); do
    if docker compose exec -T postgres pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then return 0; fi
    sleep 1
  done
  fail "PostgreSQL did not become ready for database $POSTGRES_DB"
}
psql_db(){ local db="$1"; shift; docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$db" "$@"; }

catalog_fingerprint(){
  local db="$1"
  psql_db "$db" -At <<'SQL' | LC_ALL=C sort | hash_stdin
SELECT 'schema|' || nspname FROM pg_namespace WHERE nspname IN ('identity','investment','analytics','audit')
UNION ALL
SELECT 'relation|' || n.nspname || '|' || c.relname || '|' || c.relkind::text
FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace
WHERE n.nspname IN ('identity','investment','analytics','audit') AND c.relkind IN ('r','p','v','m','S','f')
UNION ALL
SELECT 'column|' || n.nspname || '|' || c.relname || '|' || a.attnum || '|' || a.attname || '|' ||
       format_type(a.atttypid,a.atttypmod) || '|' || a.attnotnull || '|' || COALESCE(pg_get_expr(d.adbin,d.adrelid),'')
FROM pg_attribute a JOIN pg_class c ON c.oid=a.attrelid JOIN pg_namespace n ON n.oid=c.relnamespace
LEFT JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum
WHERE n.nspname IN ('identity','investment','analytics','audit') AND a.attnum>0 AND NOT a.attisdropped
UNION ALL
SELECT 'constraint|' || n.nspname || '|' || c.relname || '|' || con.conname || '|' || con.contype::text || '|' ||
       pg_get_constraintdef(con.oid,true) || '|validated=' || con.convalidated::text ||
       '|enforced=' || con.conenforced::text
FROM pg_constraint con JOIN pg_class c ON c.oid=con.conrelid JOIN pg_namespace n ON n.oid=c.relnamespace
WHERE n.nspname IN ('identity','investment','analytics','audit')
UNION ALL
SELECT 'index|' || n.nspname || '|' || tbl.relname || '|' || idx.relname || '|' ||
       pg_get_indexdef(pi.indexrelid) || '|valid=' || pi.indisvalid::text ||
       '|ready=' || pi.indisready::text || '|live=' || pi.indislive::text
FROM pg_index pi
JOIN pg_class idx ON idx.oid=pi.indexrelid
JOIN pg_class tbl ON tbl.oid=pi.indrelid
JOIN pg_namespace n ON n.oid=tbl.relnamespace
WHERE n.nspname IN ('identity','investment','analytics','audit')
UNION ALL
SELECT 'function|' || n.nspname || '|' || p.proname || '|' || pg_get_function_identity_arguments(p.oid) || '|' ||
       replace(pg_get_functiondef(p.oid), E'\n',' ')
FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace
WHERE n.nspname IN ('identity','investment','analytics','audit')
UNION ALL
SELECT 'trigger|' || n.nspname || '|' || c.relname || '|' || t.tgname || '|' ||
       pg_get_triggerdef(t.oid,true) || '|enabled=' || t.tgenabled::text
FROM pg_trigger t JOIN pg_class c ON c.oid=t.tgrelid JOIN pg_namespace n ON n.oid=c.relnamespace
WHERE n.nspname IN ('identity','investment','analytics','audit') AND NOT t.tgisinternal
UNION ALL
SELECT 'constraint-trigger|' || cn.nspname || '|' || ct.relname || '|' || con.conname || '|' ||
       rn.nspname || '|' || rt.relname || '|' || p.proname || '|type=' || t.tgtype::text ||
       '|enabled=' || t.tgenabled::text
FROM pg_trigger t
JOIN pg_constraint con ON con.oid=t.tgconstraint
JOIN pg_class ct ON ct.oid=con.conrelid
JOIN pg_namespace cn ON cn.oid=ct.relnamespace
JOIN pg_class rt ON rt.oid=t.tgrelid
JOIN pg_namespace rn ON rn.oid=rt.relnamespace
JOIN pg_proc p ON p.oid=t.tgfoid
WHERE t.tgisinternal
  AND t.tgconstraint <> 0
  AND cn.nspname IN ('identity','investment','analytics','audit');
SQL
}
apply_migrations(){
  local db="$1" migration
  for migration in "${MIGRATIONS[@]}"; do
    echo "Applying $(basename "$migration") to $db"
    psql_db "$db" < "$migration" >/dev/null
  done
}

require_command docker
require_command go
docker info >/dev/null 2>&1 || fail "Docker daemon is not available"
cd "$ROOT_DIR"
echo "Validating canonical PostgreSQL migration set."
( cd backend-go && go run ./cmd/validate-migrations --mode=repository )

shopt -s nullglob
MIGRATIONS=("$MIGRATION_DIR"/*.up.sql)
shopt -u nullglob
[[ "${#MIGRATIONS[@]}" -gt 0 ]] || fail "no canonical *.up.sql migrations discovered"
echo "UP_MIGRATIONS_DISCOVERED=${#MIGRATIONS[@]}"
echo "MIGRATION_SOURCE=infrastructure/postgres/migrations/*.up.sql"
echo "MIGRATION_ORDERING=LC_ALL=C filename order"

if [[ "$SKIP_INFRA" != "true" ]]; then
  echo "Starting PostgreSQL and Redis."
  docker compose up -d postgres redis >/dev/null
fi
wait_for_postgres

REFERENCE_DB="openinvest_bootstrap_reference_${$}_${RANDOM}"
REFERENCE_CREATED=0

reference_db_exists(){
  psql_db postgres -Atqc "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = '$REFERENCE_DB');"
}

remove_reference_db(){
  local exists
  if [[ "$REFERENCE_CREATED" -ne 1 ]]; then
    return 0
  fi

  if ! docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres \
      -c "DROP DATABASE IF EXISTS \"$REFERENCE_DB\" WITH (FORCE);" >/dev/null; then
    return 1
  fi

  if ! exists="$(reference_db_exists)"; then
    return 1
  fi
  if [[ "$exists" != "f" ]]; then
    return 1
  fi

  REFERENCE_CREATED=0
  return 0
}

cleanup_reference_emergency(){
  if [[ "$REFERENCE_CREATED" -eq 1 ]]; then
    set +e
    docker compose exec -T postgres psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d postgres \
      -c "DROP DATABASE IF EXISTS \"$REFERENCE_DB\" WITH (FORCE);" >/dev/null 2>&1
    set -e
  fi
}
trap cleanup_reference_emergency EXIT

echo "Building disposable canonical schema reference."
psql_db postgres -c "CREATE DATABASE \"$REFERENCE_DB\";" >/dev/null
REFERENCE_CREATED=1
echo "REFERENCE_DB_CREATED=YES"
apply_migrations "$REFERENCE_DB"
REFERENCE_FINGERPRINT="$(catalog_fingerprint "$REFERENCE_DB")"
[[ -n "$REFERENCE_FINGERPRINT" ]] || fail "canonical schema fingerprint is empty"

MANAGED_SCHEMA_COUNT="$(psql_db "$POSTGRES_DB" -Atqc "SELECT count(*) FROM pg_namespace WHERE nspname IN ('identity','investment','analytics','audit');")"
APPLIED_COUNT=0
DATABASE_STATE=""
if [[ "$MANAGED_SCHEMA_COUNT" == "0" ]]; then
  DATABASE_STATE="FRESH"
  echo "Fresh database detected; applying canonical migrations."
  apply_migrations "$POSTGRES_DB"
  APPLIED_COUNT="${#MIGRATIONS[@]}"
else
  CURRENT_FINGERPRINT="$(catalog_fingerprint "$POSTGRES_DB")"
  if [[ "$CURRENT_FINGERPRINT" == "$REFERENCE_FINGERPRINT" ]]; then
    DATABASE_STATE="CURRENT"
    echo "Database schema already matches the canonical migration set; no migrations replayed."
  else
    echo "BOOTSTRAP_DATABASE_STATE=PARTIAL_OR_AMBIGUOUS" >&2
    echo "No migrations were replayed and no developer data was dropped." >&2
    fail "partial or ambiguous database state; inspect the database before retrying"
  fi
fi

TARGET_FINGERPRINT="$(catalog_fingerprint "$POSTGRES_DB")"
[[ "$TARGET_FINGERPRINT" == "$REFERENCE_FINGERPRINT" ]] || fail "post-bootstrap schema fingerprint does not match canonical migrations"

LATEST_OBJECTS="$(psql_db "$POSTGRES_DB" -At <<'SQL'
SELECT
  (to_regclass('investment.portfolio_manual_valuations') IS NOT NULL)::int || '|' ||
  (SELECT count(*) FROM pg_constraint c
    JOIN pg_class t ON t.oid=c.conrelid
    JOIN pg_namespace n ON n.oid=t.relnamespace
    WHERE n.nspname='investment' AND t.relname='portfolio_manual_valuations'
      AND c.conname IN ('portfolio_manual_valuations_portfolio_fk','portfolio_manual_valuations_asset_fk'));
SQL
)"
[[ "$LATEST_OBJECTS" == "1|2" ]] || fail "latest canonical migration objects are missing"

echo "Removing disposable canonical schema reference."
if ! remove_reference_db; then
  fail "failed to remove disposable canonical reference database $REFERENCE_DB"
fi
echo "REFERENCE_DB_REMOVED=YES"
echo "REFERENCE_DB_EXISTS_AFTER_SUCCESS=NO"

echo "BOOTSTRAP_DATABASE_STATE=$DATABASE_STATE"
echo "UP_MIGRATIONS_DISCOVERED=${#MIGRATIONS[@]}"
echo "UP_MIGRATIONS_APPLIED=$APPLIED_COUNT"
echo "MIGRATION_FAILURES=0"
echo "LATEST_SCHEMA_OBJECTS=PASS"
echo "BOOTSTRAP_RESULT=PASS"
echo
echo "Next steps:"
echo "  pnpm run dev:api"
echo "  pnpm run dev:web"
