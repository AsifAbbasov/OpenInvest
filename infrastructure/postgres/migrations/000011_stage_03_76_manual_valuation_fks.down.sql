BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';
ALTER TABLE investment.portfolio_manual_valuations
    DROP CONSTRAINT portfolio_manual_valuations_asset_fk;
ALTER TABLE investment.portfolio_manual_valuations
    DROP CONSTRAINT portfolio_manual_valuations_portfolio_fk;
COMMIT;