# Stage 3.11 — Authentication and Privacy-Boundary Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-11-AUTH-PLANNING |
| Version | 0.1.0 |
| Status | Draft / planning active |
| Owner | Builder Engineer |
| Supersedes | Local development subject as an acceptable long-term user boundary |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-005; ADR-006; ADR-007; Stage 2 contract baseline; Stage 3.10 |
| Last Review Date | 2026-07-09 |
| Next Review Date | Before Stage 3.11 implementation |

## Purpose

Stage 3.11 plans the smallest safe path from the Stage 3 local development subject to the approved
MVP web authentication and privacy-default boundary.

The goal is not to implement authentication in this PR. The goal is to define the exact future
implementation scope so the product can move from a developer-only vertical slice toward a real
user-owned account model without weakening Privacy by Design, API First, or the Go API boundary.

## Why this is next

OpenInvest now has a working portfolio, transaction, snapshot, dashboard, broker-file import review,
append, and Web presentation path. The largest remaining blocker before public-MVP usefulness is
that the current vertical slice still relies on a development subject boundary instead of the frozen
MVP registration/session/privacy model.

Implementing more portfolio analytics before this boundary would make the demo richer, but it would
delay the point where data ownership, session security, export/delete behavior, and audit ownership
can be tested with real account semantics.

## Planning scope

This planning stage defines:

- registration and login implementation surfaces;
- refresh/logout session lifecycle;
- bearer access token and Secure HttpOnly refresh-cookie boundary from ADR-006;
- CSRF behavior for cookie-authenticated refresh/logout;
- privacy-default settings for new accounts;
- identity-to-investment subject mapping rules;
- audit events required for auth/session changes;
- password policy and Argon2id acceptance criteria;
- rate-limit and replay-protection expectations;
- local development subject sunset path;
- verification and review gates for the future implementation PR.

## Allowed

- documentation only;
- implementation sequencing;
- security and privacy acceptance criteria;
- future test matrix;
- rollback and migration-risk notes;
- governance registry synchronization.

## Forbidden

This planning PR must not add:

- Go auth handlers;
- repositories or services;
- password hashing implementation;
- JWT signing or refresh-token rotation implementation;
- SQL migrations;
- OpenAPI changes;
- Next.js auth screens;
- browser session state;
- cookies;
- LocalStorage, SessionStorage, or IndexedDB persistence;
- external identity providers;
- email sending;
- workers;
- tax logic;
- mobile implementation;
- AI functionality;
- Stage 3.11 implementation code.

## Frozen contract boundaries

Stage 3.11 planning must preserve the Stage 2 OpenAPI contract:

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `POST /api/v1/auth/logout`

If implementation later discovers that the contract is insufficient, work must stop and a separate
contract-change proposal must be reviewed before code changes.

## Security and privacy decisions to preserve

- New accounts default to Privacy Mode ON.
- Tax Profile remains OFF.
- Notifications remain OFF.
- Analytics remain anonymous.
- Passport, INN, phone, address, and tax profile are not required.
- Passwords use Argon2id.
- Refresh tokens rotate.
- Refresh/logout cookie flows require CSRF protection.
- No passwords, tokens, passport data, INN, XML/PDF content, or raw secrets may be logged.
- Identity data and investment data remain separated by schema and by domain boundary.

## Future implementation acceptance criteria

The future implementation PR should not be accepted until it proves:

- registration creates the minimal identity record and privacy defaults;
- login returns the approved web session shape without exposing refresh tokens to JavaScript;
- refresh rotates tokens and rejects replay;
- logout revokes the current session or requested session scope;
- identity-to-investment subject mapping is deterministic and auditable;
- portfolio/import endpoints can be scoped to the authenticated subject;
- local development subject behavior remains explicit and cannot leak into production mode;
- rate limits and validation prevent obvious brute-force and replay abuse;
- tests cover success, validation failure, replay, logout, and CSRF rejection paths;
- no business logic moves into Next.js.

## Open questions

No architecture question is opened by this planning document.

Implementation details that may require a future decision:

- exact session table shape;
- exact refresh-token hash storage format;
- local development bypass switch name and production guard;
- whether email verification is deferred or required before public MVP.

Any answer that changes the frozen contract, privacy model, or database ownership must be handled by
ADR or contract-change review before implementation.
