# Stage 3.12 — Web Authentication UI Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-12-AUTH-UI-PLAN |
| Version | 0.1.2 |
| Status | Complete / merged into `develop` |
| Owner | Builder Engineer |
| Supersedes | Informal next-step discussion after Stage 3.11 |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-007; Stage 2 contract baseline; Stage 3.11 auth/privacy slice |
| Last Review Date | 2026-07-11 |
| Next Review Date | Before Stage 3.12 merge approval |

## Purpose

Stage 3.12 plans the future Web presentation layer for the already implemented Stage 3.11 Go API
authentication boundary.

The goal is to define the smallest safe UI path for registration, login, refresh-aware authenticated
shell behavior, logout, and privacy-default visibility without moving business authority out of the
Go API.

This document is planning only. It does not authorize implementation by itself.

## Proposed implementation outcome

After a separately reviewed implementation PR, a local developer should be able to demonstrate:

```text
Open Web shell
→ register a new account through the Go API
→ receive access-token response and HttpOnly refresh cookie
→ view authenticated shell state
→ call existing portfolio/import pages with Go API authorization
→ refresh session through the Go API CSRF boundary
→ logout and clear the browser session
```

## Allowed future implementation scope

A future Stage 3.12 implementation PR may add only:

- Next.js presentation screens for registration and login;
- presentation-only logout action wiring to the Go API;
- a typed frontend auth API client that calls only the existing Go API contract;
- in-memory access-token handling for the browser session;
- CSRF token handling only as required by the Stage 3.11 Go API contract;
- authenticated Web shell state and redirects;
- loading, error, success, accessibility, and responsive states;
- tests for presentation behavior and API-boundary calls;
- documentation updates.

## Forbidden scope

Stage 3.12 must not add:

- business logic in Next.js;
- financial calculations;
- tax logic;
- direct database access;
- SQL migrations;
- Go auth handler changes;
- OpenAPI contract changes;
- Route Handlers or Server Actions for business/auth domains, including proxy or wrapper calls to
  the Go API;
- LocalStorage, SessionStorage, or IndexedDB persistence for business data, access tokens, or
  refresh tokens;
- cookie or other JavaScript-readable durable access-token storage;
- storing refresh tokens in JavaScript-readable storage;
- email verification, SMTP, OAuth, passkeys, or 2FA;
- provider integrations;
- workers;
- mobile implementation;
- AI functionality;
- Stage 3.13 work.

## Security and privacy constraints

- Refresh tokens remain HttpOnly-cookie only.
- The browser must never receive a refresh token in JSON.
- Access-token storage must be minimal and bounded to the active browser session.
- Access tokens must remain in memory only and must not be persisted to LocalStorage,
  SessionStorage, IndexedDB, cookies, logs, or durable frontend state.
- No password, access token, CSRF token, cookie, or auth response payload may be logged.
- UI copy must preserve Privacy First defaults:
  - Privacy Mode ON;
  - Tax Profile OFF;
  - Notifications OFF;
  - anonymous analytics.
- Passport, INN, phone, address, and tax profile collection remain out of scope.
- Next.js remains presentation only under ADR-007.

## Planning decisions

- Stage 3.12 should prefer a small auth shell over a broad account-management module.
- Password reset, email verification, 2FA, device management, and profile deletion remain future
  stages.
- The implementation should not introduce a frontend state-management library unless a concrete
  reviewed need appears.
- Existing portfolio/import screens should continue to call the Go API; auth UI may only supply the
  access-token boundary needed to reach those endpoints.

## Acceptance criteria for a future implementation PR

- Registration and login screens call only `/api/v1/auth/register` and `/api/v1/auth/login`.
- Refresh/logout calls preserve the Stage 3.11 CSRF and cookie boundary.
- Credentialed refresh/logout calls remain limited to the approved local Web origins.
- No refresh token is JavaScript-readable.
- No business data, access token, or refresh token is stored in LocalStorage, SessionStorage, or
  IndexedDB.
- Unauthenticated users cannot use authenticated Web routes without an access token.
- Existing portfolio/import calls use bearer authorization directly against the Go API.
- Authenticated users can logout and return to an unauthenticated state.
- Tests cover success, invalid credentials, missing CSRF, logout, and basic route protection.
- Tests cover in-memory-only access-token handling, credentialed refresh/logout behavior, approved
  local CORS credentials, and bearer authorization for existing portfolio/import calls.
- CI is green.
- Independent review confirms no Next.js business authority was introduced.

## Review focus

Review must specifically verify:

- ADR-007 compliance;
- Go API remains the only canonical business/auth authority;
- no Next.js business API routes;
- no direct PostgreSQL/Redis access from Web;
- no provider integrations;
- no token leakage to logs or JavaScript-readable storage;
- no Stage 3.13 implementation.

## Recommended next step

Open a separate Stage 3.12 implementation PR only after this planning document is reviewed,
approved, merged into `develop`, and the human reviewer explicitly authorizes implementation.
