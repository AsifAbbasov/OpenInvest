#!/usr/bin/env bash
set -euo pipefail

: "${OPENINVEST_DR_PRIMARY_URL:?OPENINVEST_DR_PRIMARY_URL must point to a disposable primary PostgreSQL database}"
: "${OPENINVEST_DR_RECOVERY_URL:?OPENINVEST_DR_RECOVERY_URL must point to a separate disposable recovery PostgreSQL database}"

if [[ "$OPENINVEST_DR_PRIMARY_URL" == "$OPENINVEST_DR_RECOVERY_URL" ]]; then
  echo "primary and recovery databases must be different" >&2
  exit 2
fi

start_epoch="$(date +%s)"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

echo "[dr] snapshotting primary"
pg_dump --format=custom --no-owner --no-acl --file="$tmpdir/dr.dump" "$OPENINVEST_DR_PRIMARY_URL"

echo "[dr] validating recovery target is reachable before destructive rehearsal"
psql "$OPENINVEST_DR_RECOVERY_URL" -v ON_ERROR_STOP=1 -Atc 'SELECT 1' >/dev/null

echo "[dr] restoring canonical state into isolated recovery target"
pg_restore --clean --if-exists --no-owner --no-acl --exit-on-error --dbname="$OPENINVEST_DR_RECOVERY_URL" "$tmpdir/dr.dump"

echo "[dr] proving recovered financial state is internally readable"
psql "$OPENINVEST_DR_RECOVERY_URL" -v ON_ERROR_STOP=1 -At <<'SQL' > "$tmpdir/dr-check.txt"
SELECT 'portfolios:' || COUNT(*) FROM investment.portfolios;
SELECT 'ledger:' || COUNT(*) FROM investment.transaction_entries;
SELECT 'snapshots:' || COUNT(*) FROM analytics.portfolio_snapshots;
SELECT 'duplicate_entry_ids:' || COUNT(*)
FROM (
  SELECT entry_id FROM investment.transaction_entries
  GROUP BY entry_id HAVING COUNT(*) > 1
) d;
SQL

grep -Fx 'duplicate_entry_ids:0' "$tmpdir/dr-check.txt" >/dev/null || {
  echo "recovered ledger contains duplicate entry IDs" >&2
  exit 3
}

end_epoch="$(date +%s)"
rto_seconds="$((end_epoch - start_epoch))"
echo "DR_REHEARSAL=PASS"
echo "DR_RTO_SECONDS=$rto_seconds"
echo "DR_COUNTS:"
cat "$tmpdir/dr-check.txt"
