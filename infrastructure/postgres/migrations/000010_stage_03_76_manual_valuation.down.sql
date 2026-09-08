BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';
DROP INDEX investment.portfolio_manual_valuations_uidx;
DROP TABLE investment.portfolio_manual_valuations;
COMMIT;