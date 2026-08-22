-- OpenInvest Stage 3.28 auth security remediation.
--
-- Scope:
-- - add refresh-session family identity so post-3.28 replay can revoke descendants fail-closed;
-- - preserve legacy sessions without fabricating lineage that was never persisted;
-- - reject new direct-database sessions that omit family identity.
--
-- Legacy rows intentionally keep session_family_id NULL. The store treats replay of such a row
-- conservatively by revoking active sessions for that user. New roots always persist a family id,
-- and rotated descendants inherit it.
--
-- Existing migration files remain immutable.

BEGIN;

ALTER TABLE identity.sessions
    ADD COLUMN session_family_id UUID;

CREATE INDEX sessions_family_state_expires_idx
    ON identity.sessions (session_family_id, session_state, expires_at);

CREATE FUNCTION identity.require_session_family_on_insert()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.session_family_id IS NULL THEN
        RAISE EXCEPTION 'session_family_id is required for new sessions'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'sessions_session_family_required';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER sessions_require_family_on_insert
BEFORE INSERT ON identity.sessions
FOR EACH ROW
EXECUTE FUNCTION identity.require_session_family_on_insert();

COMMIT;
