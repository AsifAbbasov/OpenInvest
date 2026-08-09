-- OpenInvest Stage 3.16 audit fixes.
--
-- Preserve the source class of ledger entries so import deduplication can protect
-- manual/import races without treating two independently entered manual trades
-- as the same business operation.

BEGIN;

ALTER TABLE investment.transaction_entries
    ADD COLUMN source_kind TEXT NOT NULL DEFAULT 'MANUAL'
        CHECK (source_kind IN ('MANUAL', 'USER_UPLOADED_FILE')),
    ADD COLUMN source_file_hash TEXT,
    ADD CONSTRAINT transaction_entries_source_provenance CHECK (
        (source_kind = 'MANUAL' AND source_file_hash IS NULL)
        OR (source_kind = 'USER_UPLOADED_FILE' AND length(trim(source_file_hash)) > 0)
    );

CREATE INDEX transaction_entries_portfolio_source_trade_idx
    ON investment.transaction_entries (portfolio_id, source_kind, trade_date DESC, entry_id DESC);

COMMIT;
