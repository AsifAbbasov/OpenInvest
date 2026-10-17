#!/usr/bin/env bash
set -euo pipefail

: "${OPENINVEST_BACKUP_SOURCE_URL:?OPENINVEST_BACKUP_SOURCE_URL must point to a disposable/source PostgreSQL database}"
: "${OPENINVEST_BACKUP_RESTORE_URL:?OPENINVEST_BACKUP_RESTORE_URL must point to a separate disposable restore PostgreSQL database}"

if [[ "$OPENINVEST_BACKUP_SOURCE_URL" == "$OPENINVEST_BACKUP_RESTORE_URL" ]]; then
  echo "refusing to restore over the source database" >&2
  exit 2
fi

pg_image="${OPENINVEST_PG_TOOL_IMAGE:-}"

pg_dump_cmd() {
  if [[ -n "$pg_image" ]]; then
    docker run --rm --network host "$pg_image" pg_dump "$@"
  else
    pg_dump "$@"
  fi
}

pg_restore_cmd() {
  if [[ -n "$pg_image" ]]; then
    docker run --rm -i --network host "$pg_image" pg_restore "$@"
  else
    pg_restore "$@"
  fi
}

psql_cmd() {
  if [[ -n "$pg_image" ]]; then
    docker run --rm -i --network host "$pg_image" psql "$@"
  else
    psql "$@"
  fi
}

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
dump="$tmpdir/openinvest.dump"

echo "[backup] creating custom-format dump"
pg_dump_cmd --format=custom --no-owner --no-acl "$OPENINVEST_BACKUP_SOURCE_URL" > "$dump"
test -s "$dump"

echo "[restore] restoring into disposable target"
pg_restore_cmd --clean --if-exists --no-owner --no-acl --exit-on-error --dbname="$OPENINVEST_BACKUP_RESTORE_URL" < "$dump"

echo "[verify] checking canonical schemas and critical relations"
psql_cmd "$OPENINVEST_BACKUP_RESTORE_URL" -v ON_ERROR_STOP=1 -At <<'SQL' > "$tmpdir/verify.txt"
SELECT 'schema:' || nspname
FROM pg_namespace
WHERE nspname IN ('identity','investment','analytics','audit')
ORDER BY 1;

SELECT 'table:' || table_schema || '.' || table_name
FROM information_schema.tables
WHERE table_schema IN ('identity','investment','analytics','audit')
  AND table_name IN ('users','sessions','portfolios','transaction_entries','portfolio_snapshots','events')
ORDER BY 1;

SELECT 'ledger_negative_sequence:' || COUNT(*)
FROM investment.transaction_entries
WHERE ledger_sequence IS NOT NULL AND ledger_sequence <= 0;

SELECT 'duplicate_entry_ids:' || COUNT(*)
FROM (
  SELECT entry_id
  FROM investment.transaction_entries
  GROUP BY entry_id
  HAVING COUNT(*) > 1
) duplicate_entries;
SQL

for required in schema:identity schema:investment schema:analytics schema:audit table:investment.portfolios table:investment.transaction_entries table:analytics.portfolio_snapshots
do
  grep -Fx "$required" "$tmpdir/verify.txt" >/dev/null || {
    echo "restore verification missing $required" >&2
    exit 3
  }
done

grep -Fx 'ledger_negative_sequence:0' "$tmpdir/verify.txt" >/dev/null || {
  echo "restored ledger contains invalid ledger_sequence values" >&2
  exit 4
}

grep -Fx 'duplicate_entry_ids:0' "$tmpdir/verify.txt" >/dev/null || {
  echo "restored ledger contains duplicate entry IDs" >&2
  exit 5
}

echo "BACKUP_RESTORE_REHEARSAL=PASS"
