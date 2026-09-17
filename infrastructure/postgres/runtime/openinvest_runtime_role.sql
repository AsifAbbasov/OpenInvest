\set ON_ERROR_STOP on

-- OI-NEW-04 capability target. Omitted means exact legacy-compatible R0.
\if :{?runtime_capability_profile}
\else
\set runtime_capability_profile R0
\endif

-- OI-NEW-01 PostgreSQL runtime least-privilege capability role.
-- Run with the migration/schema-owner connection after canonical migrations.
-- The entire capability reconstruction is transactional: already-running API sessions never observe
-- the broad reset without the exact replacement grants, and a failed reapplication rolls back intact.
-- No default-privilege mutation and no broad future-object grant: Expand objects are safe with zero capability.
-- Runtime startup separately rejects predefined pg_* roles, explicit parameter ACLs, database ownership/CREATE,
-- user-schema CREATE, and executable SECURITY DEFINER paths even when acquired by the provider LOGIN independently.

BEGIN;

CREATE TEMP TABLE openinvest_runtime_capability_target (
    profile TEXT PRIMARY KEY CHECK (profile IN ('R0', 'R1', 'R2'))
) ON COMMIT DROP;
INSERT INTO openinvest_runtime_capability_target(profile)
VALUES (upper(:'runtime_capability_profile'));

DO $runtime_role$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'openinvest_runtime') THEN
        CREATE ROLE openinvest_runtime
            NOLOGIN
            NOSUPERUSER
            NOCREATEDB
            NOCREATEROLE
            NOREPLICATION
            NOBYPASSRLS;
    END IF;
END
$runtime_role$;

ALTER ROLE openinvest_runtime
    NOLOGIN
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOREPLICATION
    NOBYPASSRLS;

-- The API capability role never owns or creates database/schema objects. Normalize only this
-- repository-owned role; do not mutate PUBLIC or unrelated provider roles.
DO $runtime_database_acl$
BEGIN
    EXECUTE format('REVOKE CREATE ON DATABASE %I FROM openinvest_runtime', current_database());
END
$runtime_database_acl$;

DO $runtime_schema_acl$
DECLARE
    schema_name text;
BEGIN
    FOR schema_name IN
        SELECT nspname
        FROM pg_namespace
        WHERE nspname <> 'pg_catalog'
          AND nspname <> 'information_schema'
          AND nspname NOT LIKE 'pg_toast%'
          AND nspname NOT LIKE 'pg_temp_%'
        ORDER BY nspname
    LOOP
        EXECUTE format('REVOKE CREATE ON SCHEMA %I FROM openinvest_runtime', schema_name);
    END LOOP;
END
$runtime_schema_acl$;

GRANT USAGE ON SCHEMA identity, investment, analytics, audit TO openinvest_runtime;

-- Deterministic convergence: clear inherited historical table ACL, then rebuild the exact maximum.
-- This broad REVOKE and all replacement grants are in this single transaction.
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA identity, investment, analytics, audit FROM openinvest_runtime;

-- identity
GRANT SELECT, INSERT ON identity.users TO openinvest_runtime;
GRANT SELECT, INSERT ON identity.user_investment_links TO openinvest_runtime;
GRANT SELECT, INSERT ON identity.credentials TO openinvest_runtime;
GRANT SELECT, INSERT ON identity.privacy_settings TO openinvest_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON identity.sessions TO openinvest_runtime;

-- investment
-- INSERT ... ON CONFLICT (id) needs SELECT only on the arbiter column.
GRANT INSERT ON investment.subjects TO openinvest_runtime;
GRANT SELECT (id) ON investment.subjects TO openinvest_runtime;
GRANT SELECT, INSERT ON investment.assets TO openinvest_runtime;

-- API production code locks portfolio rows with SELECT ... FOR UPDATE but performs no business
-- UPDATE investment.portfolios. PostgreSQL requires UPDATE on at least one column for row locking;
-- portfolio_state is the sole lock-enabling UPDATE column. The existing lifecycle CHECK couples it
-- to removed_at, preventing a meaningful active<->removed transition with this one-column capability.
GRANT SELECT, INSERT ON investment.portfolios TO openinvest_runtime;
GRANT UPDATE (portfolio_state) ON investment.portfolios TO openinvest_runtime;

REVOKE UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER ON investment.transaction_entries FROM PUBLIC, openinvest_runtime;
GRANT SELECT, INSERT ON investment.transaction_entries TO openinvest_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON investment.command_deduplication TO openinvest_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON investment.portfolio_manual_valuations TO openinvest_runtime;

-- analytics
GRANT SELECT, INSERT ON analytics.portfolio_snapshots TO openinvest_runtime;

-- audit.actors INSERT ... ON CONFLICT (id) needs SELECT only on the arbiter column.
GRANT INSERT ON audit.actors TO openinvest_runtime;
GRANT SELECT (id) ON audit.actors TO openinvest_runtime;

-- Stage 3.33 frozen invariant: audit evidence remains read + append only.
REVOKE UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER ON audit.events FROM PUBLIC, openinvest_runtime;
GRANT SELECT, INSERT ON audit.events TO openinvest_runtime;

-- Runtime identifiers are application-generated UUIDs; sequences are never a runtime capability.
-- OI-NEW-04 replay state: R0 grants nothing, R1 is read-only, R2 adds INSERT only on derived state.
DO $replay_profile$
DECLARE
    target_profile text;
BEGIN
    SELECT profile INTO STRICT target_profile FROM openinvest_runtime_capability_target;

    IF to_regclass('analytics.replay_policy_generations') IS NOT NULL THEN
        REVOKE ALL PRIVILEGES ON analytics.replay_policy_generations FROM openinvest_runtime;
        REVOKE ALL PRIVILEGES ON analytics.replay_policy_events FROM openinvest_runtime;
        REVOKE ALL PRIVILEGES ON analytics.portfolio_replay_epochs FROM openinvest_runtime;
        REVOKE ALL PRIVILEGES ON analytics.portfolio_replay_positions FROM openinvest_runtime;
        REVOKE ALL PRIVILEGES ON analytics.portfolio_replay_financial_state FROM openinvest_runtime;

        IF target_profile IN ('R1', 'R2') THEN
            GRANT SELECT ON analytics.replay_policy_generations TO openinvest_runtime;
            GRANT SELECT ON analytics.replay_policy_events TO openinvest_runtime;
            GRANT SELECT ON analytics.portfolio_replay_epochs TO openinvest_runtime;
            GRANT SELECT ON analytics.portfolio_replay_positions TO openinvest_runtime;
            GRANT SELECT ON analytics.portfolio_replay_financial_state TO openinvest_runtime;
        END IF;

        IF target_profile = 'R2' THEN
            GRANT INSERT ON analytics.portfolio_replay_epochs TO openinvest_runtime;
            GRANT INSERT ON analytics.portfolio_replay_positions TO openinvest_runtime;
            GRANT INSERT ON analytics.portfolio_replay_financial_state TO openinvest_runtime;
        END IF;
    ELSIF target_profile <> 'R0' THEN
        RAISE EXCEPTION 'runtime capability profile % requires OI-NEW-04 replay schema', target_profile;
    END IF;
END
$replay_profile$;

-- Runtime identifiers are application-generated UUIDs; sequences are never a runtime capability.
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA identity, investment, analytics, audit FROM openinvest_runtime;

COMMIT;
