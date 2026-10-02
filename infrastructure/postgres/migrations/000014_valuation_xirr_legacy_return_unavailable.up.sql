-- OI-XIRR-001 invalidates legacy snapshot ROI placeholders without inventing a replacement return methodology.
BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';

ALTER TABLE analytics.portfolio_snapshots
    ADD COLUMN legacy_return_status TEXT NOT NULL DEFAULT 'UNAVAILABLE';

ALTER TABLE analytics.portfolio_snapshots
    ADD CONSTRAINT portfolio_snapshots_legacy_return_status_check
    CHECK (legacy_return_status = 'UNAVAILABLE') NOT VALID;

COMMIT;
