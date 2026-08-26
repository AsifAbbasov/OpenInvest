-- OpenInvest Stage 3.38 operational retention support.
--
-- Scope:
-- - additive expiry-leading indexes only;
-- - no data rewrite;
-- - no privilege expansion;
-- - no migration-policy remediation (P3-08 remains separate).

BEGIN;

CREATE INDEX command_deduplication_expires_id_idx
    ON investment.command_deduplication (expires_at, id);

CREATE INDEX sessions_expires_id_idx
    ON identity.sessions (expires_at, id);

COMMIT;
