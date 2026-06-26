-- Roll back OpenInvest Stage 3.1 database foundation.
--
-- This rollback is intended for local/disposable Stage 3.1 environments only.
-- It removes only structures introduced by the paired up migration.
-- Production destructive Contract migrations remain forbidden without a staged ADR.

BEGIN;

DROP TABLE IF EXISTS audit.events;
DROP TABLE IF EXISTS audit.actors;

DROP TABLE IF EXISTS analytics.inbox_messages;
DROP TABLE IF EXISTS analytics.calculation_runs;
DROP TABLE IF EXISTS analytics.snapshot_positions;
DROP TABLE IF EXISTS analytics.portfolio_snapshots;

DROP TABLE IF EXISTS investment.outbox_events;
DROP TABLE IF EXISTS investment.command_deduplication;
DROP TABLE IF EXISTS investment.transaction_entries;
DROP TABLE IF EXISTS investment.portfolios;
DROP TABLE IF EXISTS identity.user_investment_links;
DROP TABLE IF EXISTS investment.assets;
DROP TABLE IF EXISTS investment.subjects;
DROP TABLE IF EXISTS identity.users;

DROP SCHEMA IF EXISTS audit;
DROP SCHEMA IF EXISTS analytics;
DROP SCHEMA IF EXISTS investment;
DROP SCHEMA IF EXISTS identity;

COMMIT;
