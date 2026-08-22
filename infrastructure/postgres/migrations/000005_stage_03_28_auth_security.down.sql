-- Roll back only the Stage 3.28 session-family schema addition.

BEGIN;

DROP TRIGGER IF EXISTS sessions_require_family_on_insert ON identity.sessions;
DROP FUNCTION IF EXISTS identity.require_session_family_on_insert();
DROP INDEX IF EXISTS identity.sessions_family_state_expires_idx;

ALTER TABLE identity.sessions
    DROP COLUMN IF EXISTS session_family_id;

COMMIT;
