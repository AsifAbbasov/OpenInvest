-- OpenInvest Stage 3.11 auth/privacy boundary.
--
-- Scope:
-- - credentials for MVP email/password authentication;
-- - privacy-by-default settings for every registered identity;
-- - rotating refresh sessions with CSRF binding;
-- - additive schema only.
--
-- Security rules:
-- - passwords are stored as Argon2id hashes only;
-- - raw refresh and CSRF tokens are never stored;
-- - refresh sessions are revocable and replay-resistant.

BEGIN;

CREATE TABLE identity.credentials (
    user_id UUID PRIMARY KEY REFERENCES identity.users (id) ON DELETE CASCADE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT credentials_argon2id_hash CHECK (password_hash LIKE 'argon2id$v=19$m=%')
);

CREATE TABLE identity.privacy_settings (
    user_id UUID PRIMARY KEY REFERENCES identity.users (id) ON DELETE CASCADE,
    privacy_mode BOOLEAN NOT NULL DEFAULT true,
    tax_profile_enabled BOOLEAN NOT NULL DEFAULT false,
    notifications_enabled BOOLEAN NOT NULL DEFAULT false,
    analytics_mode TEXT NOT NULL DEFAULT 'anonymous',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE identity.sessions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES identity.users (id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    csrf_token_hash TEXT NOT NULL,
    session_state TEXT NOT NULL DEFAULT 'active'
        CHECK (session_state IN ('active', 'revoked')),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_rotated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT sessions_refresh_hash_not_blank CHECK (length(trim(refresh_token_hash)) > 0),
    CONSTRAINT sessions_csrf_hash_not_blank CHECK (length(trim(csrf_token_hash)) > 0),
    CONSTRAINT sessions_revoked_timestamp CHECK (
        (session_state = 'revoked' AND revoked_at IS NOT NULL)
        OR (session_state = 'active' AND revoked_at IS NULL)
    ),
    CONSTRAINT sessions_expiry_after_creation CHECK (expires_at > created_at)
);

CREATE UNIQUE INDEX sessions_refresh_token_hash_uidx
    ON identity.sessions (refresh_token_hash);

CREATE INDEX sessions_user_state_expires_idx
    ON identity.sessions (user_id, session_state, expires_at);

COMMIT;
