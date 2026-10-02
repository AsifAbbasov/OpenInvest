-- Reverse only in a disposable environment before the status is relied on by an accepted runtime.
BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';

ALTER TABLE analytics.portfolio_snapshots
    DROP CONSTRAINT portfolio_snapshots_legacy_return_status_check;

ALTER TABLE analytics.portfolio_snapshots
    DROP COLUMN legacy_return_status;

COMMIT;
