#!/usr/bin/env bash
set -euo pipefail

: "${OPENINVEST_DR_PRIMARY_URL:?OPENINVEST_DR_PRIMARY_URL must point to a disposable primary PostgreSQL database}"
: "${OPENINVEST_DR_RECOVERY_URL:?OPENINVEST_DR_RECOVERY_URL must point to a separate disposable recovery PostgreSQL database}"

if [[ "$OPENINVEST_DR_PRIMARY_URL" == "$OPENINVEST_DR_RECOVERY_URL" ]]; then
  echo "primary and recovery databases must be different" >&2
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

start_epoch="$(date +%s)"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

echo "[dr] snapshotting primary"
pg_dump_cmd --format=custom --no-owner --no-acl "$OPENINVEST_DR_PRIMARY_URL" > "$tmpdir/dr.dump"
test -s "$tmpdir/dr.dump"

echo "[dr] validating recovery target is reachable before destructive rehearsal"
psql_cmd "$OPENINVEST_DR_RECOVERY_URL" -v ON_ERROR_STOP=1 -Atc 'SELECT 1' >/dev/null

echo "[dr] restoring canonical state into isolated recovery target"
pg_restore_cmd --clean --if-exists --no-owner --no-acl --exit-on-error --dbname="$OPENINVEST_DR_RECOVERY_URL" < "$tmpdir/dr.dump"

echo "[dr] proving recovered financial state is internally readable"
psql_cmd "$OPENINVEST_DR_RECOVERY_URL" -v ON_ERROR_STOP=1 -At <<'SQL' > "$tmpdir/dr-check.txt"
SELECT 'portfolios:' || COUNT(*) FROM investment.portfolios;
SELECT 'ledger:' || COUNT(*) FROM investment.transaction_entries;
SELECT 'snapshots:' || COUNT(*) FROM analytics.portfolio_snapshots;
SELECT 'duplicate_entry_ids:' || COUNT(*)
FROM (
  SELECT entry_id FROM investment.transaction_entries
  GROUP BY entry_id HAVING COUNT(*) > 1
) d;
SELECT 'invalid_ledger_sequences:' || COUNT(*)
FROM investment.transaction_entries
WHERE ledger_sequence IS NOT NULL AND ledger_sequence <= 0;
SQL

grep -Fx 'duplicate_entry_ids:0' "$tmpdir/dr-check.txt" >/dev/null || {
  echo "recovered ledger contains duplicate entry IDs" >&2
  exit 3
}

grep -Fx 'invalid_ledger_sequences:0' "$tmpdir/dr-check.txt" >/dev/null || {
  echo "recovered ledger contains invalid ledger sequences" >&2
  exit 4
}

end_epoch="$(date +%s)"
rto_seconds="$((end_epoch - start_epoch))"
echo "DR_REHEARSAL=PASS"
echo "DR_RTO_SECONDS=$rto_seconds"
echo "DR_COUNTS:"
cat "$tmpdir/dr-check.txt"
