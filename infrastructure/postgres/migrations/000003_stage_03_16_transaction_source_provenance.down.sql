BEGIN;

DROP INDEX IF EXISTS investment.transaction_entries_portfolio_source_trade_idx;

ALTER TABLE investment.transaction_entries
    DROP CONSTRAINT IF EXISTS transaction_entries_source_provenance,
    DROP COLUMN IF EXISTS source_file_hash,
    DROP COLUMN IF EXISTS source_kind;

COMMIT;
