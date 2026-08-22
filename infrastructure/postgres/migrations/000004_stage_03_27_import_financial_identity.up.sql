-- OpenInvest Stage 3.27 import and financial-identity remediation.
--
-- Persist a privacy-minimized per-row import identity so independently executed
-- broker operations are not collapsed merely because their normalized economics
-- are identical. Legacy rows remain valid with NULL identity metadata.

BEGIN;

ALTER TABLE investment.transaction_entries
    ADD COLUMN source_account_label TEXT NOT NULL DEFAULT '',
    ADD COLUMN source_broker_operation_key TEXT,
    ADD COLUMN source_fingerprint TEXT,
    ADD COLUMN source_identity_version SMALLINT,
    ADD CONSTRAINT transaction_entries_import_identity CHECK (
        (
            source_identity_version IS NULL
            AND source_account_label = ''
            AND source_broker_operation_key IS NULL
            AND source_fingerprint IS NULL
        )
        OR (
            source_kind = 'USER_UPLOADED_FILE'
            AND source_identity_version = 1
            AND char_length(source_account_label) <= 120
            AND source_fingerprint ~ '^[0-9a-f]{64}$'
            AND (
                source_broker_operation_key IS NULL
                OR source_broker_operation_key ~ '^[0-9a-f]{64}$'
            )
        )
    ),
    ADD CONSTRAINT transaction_entries_cash_flow_fees_zero CHECK (
        transaction_type NOT IN ('DEPOSIT', 'WITHDRAWAL')
        OR (commission_amount = 0 AND tax_amount = 0)
    );

CREATE UNIQUE INDEX transaction_entries_import_broker_identity_uidx
    ON investment.transaction_entries (
        portfolio_id,
        source_kind,
        source_account_label,
        source_broker_operation_key
    )
    WHERE source_kind = 'USER_UPLOADED_FILE'
        AND source_identity_version = 1
        AND source_broker_operation_key IS NOT NULL;

CREATE UNIQUE INDEX transaction_entries_import_fingerprint_identity_uidx
    ON investment.transaction_entries (
        portfolio_id,
        source_kind,
        source_account_label,
        source_fingerprint
    )
    WHERE source_kind = 'USER_UPLOADED_FILE'
        AND source_identity_version = 1
        AND source_broker_operation_key IS NULL;

CREATE FUNCTION investment.enforce_import_identity_strength_consistency()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    scoped_fingerprint_lock BIGINT;
BEGIN
    IF NEW.source_kind <> 'USER_UPLOADED_FILE'
        OR NEW.source_identity_version IS DISTINCT FROM 1
        OR NEW.source_fingerprint IS NULL THEN
        RETURN NEW;
    END IF;

    scoped_fingerprint_lock := hashtextextended(
        NEW.portfolio_id::text || '|' ||
        NEW.source_account_label || '|' ||
        NEW.source_identity_version::text || '|' ||
        NEW.source_fingerprint,
        0
    );
    PERFORM pg_advisory_xact_lock(scoped_fingerprint_lock);

    IF NEW.source_broker_operation_key IS NULL THEN
        IF EXISTS (
            SELECT 1
            FROM investment.transaction_entries te
            WHERE te.portfolio_id = NEW.portfolio_id
                AND te.source_kind = 'USER_UPLOADED_FILE'
                AND te.source_identity_version = NEW.source_identity_version
                AND te.source_account_label = NEW.source_account_label
                AND te.source_fingerprint = NEW.source_fingerprint
                AND te.source_broker_operation_key IS NOT NULL
                AND te.entry_id <> NEW.entry_id
        ) THEN
            RAISE EXCEPTION 'mixed fallback/strong import identity is not allowed'
                USING ERRCODE = '23505',
                    CONSTRAINT = 'transaction_entries_import_identity_strength_conflict';
        END IF;
    ELSE
        IF EXISTS (
            SELECT 1
            FROM investment.transaction_entries te
            WHERE te.portfolio_id = NEW.portfolio_id
                AND te.source_kind = 'USER_UPLOADED_FILE'
                AND te.source_identity_version = NEW.source_identity_version
                AND te.source_account_label = NEW.source_account_label
                AND te.source_fingerprint = NEW.source_fingerprint
                AND te.source_broker_operation_key IS NULL
                AND te.entry_id <> NEW.entry_id
        ) THEN
            RAISE EXCEPTION 'mixed fallback/strong import identity is not allowed'
                USING ERRCODE = '23505',
                    CONSTRAINT = 'transaction_entries_import_identity_strength_conflict';
        END IF;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER transaction_entries_import_identity_strength_guard
    BEFORE INSERT OR UPDATE ON investment.transaction_entries
    FOR EACH ROW
    EXECUTE FUNCTION investment.enforce_import_identity_strength_consistency();

COMMIT;
