BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';
CREATE TABLE investment.portfolio_manual_valuations (
    portfolio_id UUID,
    asset_id UUID,
    position_opened_ledger_sequence BIGINT,
    price_amount NUMERIC(28, 8),
    price_currency VARCHAR(3),
    as_of_date DATE,
    source TEXT,
    updated_at TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX portfolio_manual_valuations_uidx
    ON investment.portfolio_manual_valuations (portfolio_id, asset_id);
COMMIT;