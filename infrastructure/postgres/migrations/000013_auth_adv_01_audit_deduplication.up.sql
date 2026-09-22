-- AUTH-ADV-01 bounds append-only security evidence for a known session.
--
-- This relation deliberately has no foreign key to identity.sessions: expired session
-- cleanup may remove the source row, while the one-time security evidence must remain.
BEGIN;
SET LOCAL lock_timeout = '5000ms';
SET LOCAL statement_timeout = '30000ms';

CREATE TABLE audit.auth_security_event_deduplications (
    action_code TEXT NOT NULL,
    session_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE UNIQUE INDEX auth_security_event_deduplications_uidx
    ON audit.auth_security_event_deduplications (action_code, session_id);

COMMIT;
