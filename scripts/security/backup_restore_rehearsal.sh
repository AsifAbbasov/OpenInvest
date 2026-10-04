#!/usr/bin/env bash
set -euo pipefail

: "${OPENINVEST_BACKUP_SOURCE_URL:?OPENINVEST_BACKUP_SOURCE_URL must point to a disposable/source PostgreSQL database}"
: "${OPENINVEST_BACKUP_RESTORE_URL:?OPENINVEST_BACKUP_RESTORE_URL must point to a separate disposable restore PostgreSQL database}"

if [[ "$OPENINVEST_BACKUP_SOURCE_URL" == "$OPENINVEST_BACKUP_RESTORE_URL" ]]; then
  echo "refusing to restore over the source database" >&2
  exit 2
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT
dump="$tmpdir/openinvest.dump"

echo "[backup] creating custom-format dump"
pg_dump --format=custom --no-owner --no-acl --file="$dump" "$OPENINVEST_BACKUP_SOURCE_URL"
test -s "$dump"

echo "[restore] restoring into disposable target"
pg_restore --clean --if-exists --no-owner --no-acl --exit-on-error --dbname="$OPENINVEST_BACKUP_RESTORE_URL" "$dump"

echo "[verify] checking canonical schemas and critical relations"
psql "$OPENINVEST_BACKUP_RESTORE_URL" -v ON_ERROR_STOP=1 -At <<'SQL' > "$tmpdir/verify.txt"
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

echo "BACKUP_RESTORE_REHEARSAL=PASS"
