-- OpenInvest Stage 3.28 auth security remediation.
--
-- Scope:
-- - add refresh-session family identity so replay can revoke descendants fail-closed;
-- - conservatively group legacy sessions by user because pre-3.28 lineage was not persisted.
--
-- Existing migration files remain immutable.

BEGIN;

ALTER TABLE identity.sessions
    ADD COLUMN session_family_id UUID;

UPDATE identity.sessions
SET session_family_id = user_id
WHERE session_family_id IS NULL;

ALTER TABLE identity.sessions
    ALTER COLUMN session_family_id SET NOT NULL;

CREATE INDEX sessions_family_state_expires_idx
    ON identity.sessions (session_family_id, session_state, expires_at);

COMMIT;
