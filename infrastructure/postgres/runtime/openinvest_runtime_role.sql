\set ON_ERROR_STOP on

-- Stage 3.33 runtime capability role.
-- Run this script with the migration/owner connection after schema migrations.
-- The API must authenticate with a separate LOGIN role that is granted membership in
-- openinvest_runtime. Never use the schema owner connection as DATABASE_URL in staging/production.

DO $runtime_role$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'openinvest_runtime') THEN
        CREATE ROLE openinvest_runtime
            NOLOGIN
            NOSUPERUSER
            NOCREATEDB
            NOCREATEROLE
            NOREPLICATION;
    END IF;
END
$runtime_role$;

ALTER ROLE openinvest_runtime
    NOLOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION;

GRANT USAGE ON SCHEMA identity, investment, analytics, audit TO openinvest_runtime;

-- Runtime tables are mutable by default only where the application currently needs mutation.
-- This script is intentionally rerun after migrations rather than granting default future-table
-- privileges: a newly introduced table therefore fails closed until its runtime access is reviewed.
GRANT SELECT, INSERT, UPDATE, DELETE
    ON ALL TABLES IN SCHEMA identity, investment, analytics, audit
    TO openinvest_runtime;
GRANT USAGE, SELECT
    ON ALL SEQUENCES IN SCHEMA identity, investment, analytics, audit
    TO openinvest_runtime;

-- Canonical financial ledger: append/read only. Correction and reversal are new INSERT rows.
REVOKE UPDATE, DELETE, TRUNCATE
    ON investment.transaction_entries
    FROM PUBLIC, openinvest_runtime;
GRANT SELECT, INSERT
    ON investment.transaction_entries
    TO openinvest_runtime;

-- Audit evidence is append/read only for the runtime role as well.
REVOKE UPDATE, DELETE, TRUNCATE
    ON audit.events
    FROM PUBLIC, openinvest_runtime;
GRANT SELECT, INSERT
    ON audit.events
    TO openinvest_runtime;
