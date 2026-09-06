SET lock_timeout = '5000ms';
SET statement_timeout = '30000ms';
CREATE UNIQUE INDEX CONCURRENTLY transaction_entries_portfolio_ledger_sequence_uidx ON investment.transaction_entries (portfolio_id, ledger_sequence);
