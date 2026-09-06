BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';
ALTER TABLE investment.transaction_entries DROP COLUMN ledger_sequence;
COMMIT;
