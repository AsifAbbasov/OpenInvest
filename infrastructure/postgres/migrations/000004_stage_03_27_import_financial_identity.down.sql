BEGIN;

DROP TRIGGER IF EXISTS transaction_entries_import_identity_strength_guard
    ON investment.transaction_entries;
DROP FUNCTION IF EXISTS investment.enforce_import_identity_strength_consistency();

DROP INDEX IF EXISTS investment.transaction_entries_import_fingerprint_identity_uidx;
DROP INDEX IF EXISTS investment.transaction_entries_import_broker_identity_uidx;

ALTER TABLE investment.transaction_entries
    DROP CONSTRAINT IF EXISTS transaction_entries_cash_flow_fees_zero,
    DROP CONSTRAINT IF EXISTS transaction_entries_import_identity,
    DROP COLUMN IF EXISTS source_identity_version,
    DROP COLUMN IF EXISTS source_fingerprint,
    DROP COLUMN IF EXISTS source_broker_operation_key,
    DROP COLUMN IF EXISTS source_account_label;

COMMIT;
