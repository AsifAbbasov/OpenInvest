BEGIN;

DROP INDEX IF EXISTS investment.command_deduplication_expires_id_idx;
DROP INDEX IF EXISTS identity.sessions_expires_id_idx;

COMMIT;
