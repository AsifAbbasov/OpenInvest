-- OpenInvest Stage 3.28 auth security remediation.
--
-- Scope:
-- - add refresh-session family identity so post-3.28 replay can revoke descendants fail-closed;
-- - preserve legacy sessions without fabricating lineage that was never persisted.
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

COMMIT;
