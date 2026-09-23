# Stage 3.11 — Authentication and Privacy-Boundary Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-11-AUTH-SLICE |
| Version | 0.1.1 |
| Status | Complete / merged into `develop` |
| Supersedes | Stage 3 local development subject as the only runtime user boundary |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-005; ADR-006; ADR-007; Stage 2 contract baseline; Stage 3.11 planning |

## Purpose

Stage 3.11 implements the smallest reviewed MVP authentication and privacy-default boundary behind
the already frozen Stage 2 OpenAPI contract.

The goal is to replace the developer-only subject boundary with a real account/session boundary
without expanding business scope or moving business authority out of the Go API.

## Implementation scope

This slice may add only:

- Go API handlers for:
  - `POST /api/v1/auth/register`;
  - `POST /api/v1/auth/login`;
  - `POST /api/v1/auth/refresh`;
  - `POST /api/v1/auth/logout`;
- Argon2id password hashing and verification;
- short-lived access-token issuance for Go API authentication;
- rotating HttpOnly refresh-cookie sessions;
- CSRF validation for cookie-authenticated refresh/logout;
- privacy-default records for new identities;
- deterministic identity-to-investment subject mapping through the existing schemas;
- PostgreSQL migration tables for credentials, privacy defaults, and sessions;
- tests for registration, login/session shape, refresh rotation, replay rejection, logout, CSRF
  rejection, and migration validation;
- non-secret audit events for registration, login, refresh, logout, and all rejected
  refresh/logout attempts including missing cookie or CSRF input;
  Historical Stage 3.11 requirement. AUTH-ADV-01 later superseded the
  per-rejection durable-audit behavior: missing, anonymous, and fully unknown
  refresh/logout rejection paths create zero durable audit rows, while
  security-relevant known-session rejection/replay evidence is retained under
  the PostgreSQL durable deduplication bound. See
  `docs/audit/AUTH_ADV_01_DURABLE_AUDIT_BOUND_CLOSURE.md`.
- documentation and governance updates for this stage.

## Explicit non-goals

This slice must not add:

- Next.js authentication screens or browser session state;
- business logic in Next.js;
- financial calculations;
- tax logic;
- email verification or SMTP;
- OAuth/passkeys/2FA;
- provider integrations;
- workers;
- mobile implementation;
-  functionality;
- portfolio/domain feature expansion;

## Security and privacy constraints

- Refresh tokens are returned only through an HttpOnly cookie.
- Refresh tokens and CSRF tokens are stored only as hashes.
- Refresh rotation must reject replay.
- CSRF must be checked as part of the same atomic rotation/revocation boundary.
- New accounts default to:
  - Privacy Mode ON;
  - Tax Profile OFF;
  - Notifications OFF;
  - anonymous analytics.
- Passport, INN, phone, address, and tax profile are not accepted by this slice.
- No password, token, cookie, CSRF token, or secret may be logged.
- Auth audit events must never include passwords, refresh tokens, CSRF tokens, access tokens, or raw
  request payloads.
- The Go API remains the only canonical business API.

## Local development boundary

The previous local development subject remains available only as an explicit development bypass for
pre-auth developer workflows. When `DATABASE_URL` is configured, unsafe local flags
(`OPENINVEST_DEV_AUTH_BYPASS`, `OPENINVEST_REFRESH_COOKIE_INSECURE`, or
`OPENINVEST_ALLOW_EPHEMERAL_ACCESS_TOKEN_SECRET`) are accepted only when `OPENINVEST_ENV` is
`development` or `local`. Production authorization must use the access-token path.

## Completion evidence

Canonical record: PR #29; commit(s) `5c49173ac858995929f266c2de991282dd194dec`, `8a8052c18768dbad0aa0e724836f3c9252d257e3`.

## Scope guard

The review for this slice must verify absence of:

- Stage 3.12 work;
- frontend authentication UI;
- provider integrations;
- workers;
- tax/email/mobile/ work;
- direct database access from Next.js;
- business API routes in Next.js.
