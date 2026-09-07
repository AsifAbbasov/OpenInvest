BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';
ALTER TABLE investment.portfolio_manual_valuations
    ADD CONSTRAINT portfolio_manual_valuations_portfolio_fk
    FOREIGN KEY (portfolio_id) REFERENCES investment.portfolios (id) NOT VALID;
ALTER TABLE investment.portfolio_manual_valuations
    ADD CONSTRAINT portfolio_manual_valuations_asset_fk
    FOREIGN KEY (asset_id) REFERENCES investment.assets (id) NOT VALID;
COMMIT;