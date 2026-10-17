#!/usr/bin/env bash
set -euo pipefail

: "${OPENINVEST_BACKUP_SOURCE_URL:?set OPENINVEST_BACKUP_SOURCE_URL to a disposable/source test database}"
: "${OPENINVEST_RESTORE_TARGET_URL:?set OPENINVEST_RESTORE_TARGET_URL to a separate disposable restore target}"
: "${OPENINVEST_ALLOW_DESTRUCTIVE_DRILL:?set OPENINVEST_ALLOW_DESTRUCTIVE_DRILL=YES}"

if [[ "$OPENINVEST_ALLOW_DESTRUCTIVE_DRILL" != "YES" ]]; then
  echo "refusing destructive restore drill without OPENINVEST_ALLOW_DESTRUCTIVE_DRILL=YES" >&2
  exit 2
fi

if [[ "$OPENINVEST_BACKUP_SOURCE_URL" == "$OPENINVEST_RESTORE_TARGET_URL" ]]; then
  echo "source and restore target must be different databases" >&2
  exit 2
fi

for binary in pg_dump pg_restore psql; do
  command -v "$binary" >/dev/null || { echo "missing required binary: $binary" >&2; exit 2; }
done

dump_file="$(mktemp -t openinvest-backup-XXXXXX.dump)"
trap 'rm -f "$dump_file"' EXIT

critical_tables=(
  identity.users
  identity.sessions
  investment.portfolios
  investment.transaction_entries
  investment.command_deduplication
  analytics.portfolio_snapshots
  audit.events
)

echo "[backup] creating custom-format dump"
pg_dump --format=custom --no-owner --no-acl --file="$dump_file" "$OPENINVEST_BACKUP_SOURCE_URL"
test -s "$dump_file"

echo "[restore] restoring into disposable target"
pg_restore --clean --if-exists --no-owner --no-acl --dbname="$OPENINVEST_RESTORE_TARGET_URL" "$dump_file"

echo "[verify] comparing critical table row counts"
for table in "${critical_tables[@]}"; do
  source_count="$(psql "$OPENINVEST_BACKUP_SOURCE_URL" -Atv ON_ERROR_STOP=1 -c "SELECT count(*) FROM $table;")"
  restore_count="$(psql "$OPENINVEST_RESTORE_TARGET_URL" -Atv ON_ERROR_STOP=1 -c "SELECT count(*) FROM $table;")"
  if [[ "$source_count" != "$restore_count" ]]; then
    echo "row-count mismatch for $table: source=$source_count restore=$restore_count" >&2
    exit 1
  fi
  echo "PASS $table rows=$source_count"
done

echo "[verify] validating restored ledger invariants"
psql "$OPENINVEST_RESTORE_TARGET_URL" -v ON_ERROR_STOP=1 <<'SQL'
DO $$
DECLARE
  duplicate_entry_ids bigint;
  duplicate_ledger_sequences bigint;
  malformed_revisions bigint;
BEGIN
  SELECT count(*) INTO duplicate_entry_ids
  FROM (
    SELECT entry_id FROM investment.transaction_entries GROUP BY entry_id HAVING count(*) > 1
  ) q;

  SELECT count(*) INTO duplicate_ledger_sequences
  FROM (
    SELECT portfolio_id, ledger_sequence
    FROM investment.transaction_entries
    WHERE ledger_sequence IS NOT NULL
    GROUP BY portfolio_id, ledger_sequence
    HAVING count(*) > 1
  ) q;

  SELECT count(*) INTO malformed_revisions
  FROM investment.transaction_entries
  WHERE revision <= 0;

  IF duplicate_entry_ids <> 0 OR duplicate_ledger_sequences <> 0 OR malformed_revisions <> 0 THEN
    RAISE EXCEPTION
      'restore integrity failure duplicate_entry_ids=% duplicate_sequences=% malformed_revisions=%',
      duplicate_entry_ids, duplicate_ledger_sequences, malformed_revisions;
  END IF;
END
$$;
SQL

echo "BACKUP_RESTORE_DRILL=PASS"
