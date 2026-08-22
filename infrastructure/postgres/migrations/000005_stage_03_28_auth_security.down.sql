-- Roll back only the Stage 3.28 session-family schema addition.

BEGIN;

DROP INDEX IF EXISTS identity.sessions_family_state_expires_idx;

ALTER TABLE identity.sessions
    DROP COLUMN IF EXISTS session_family_id;

COMMIT;
