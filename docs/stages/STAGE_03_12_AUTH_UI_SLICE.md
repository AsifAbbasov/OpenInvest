# Stage 3.12 — Web Authentication UI Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-12-AUTH-UI-SLICE |
| Version | 0.1.4 |
| Status | Complete / closed; merged into `develop` at `b4840b60346109e3cd54a07d9e1e131fc0cfad23` |
| Owner | Builder Engineer |
| Supersedes | Stage 3.12 Web authentication UI planning |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-007; Stage 3.11 auth/privacy slice; Stage 3.12 auth UI planning |
| Last Review Date | 2026-07-12 |
| Next Review Date | Before Stage 3.13 implementation review |

## Purpose

Stage 3.12 implements the approved Web presentation layer over the existing Stage 3.11 Go API
authentication boundary.

The slice adds the smallest safe registration/login/session shell path without moving business
authority, auth authority, persistence, or financial logic into Next.js.

## Implementation scope

This slice may include only:

- Next.js registration and login presentation states;
- a small authenticated shell that gates the existing Web portfolio/import screens;
- typed frontend calls to the existing Go API auth endpoints;
- bearer access-token propagation from in-memory React state to existing Go API calls;
- CSRF propagation for refresh/logout;
- credentialed browser requests for the existing HttpOnly refresh cookie;
- local CORS credentials support needed for the browser to send that cookie;
- tests for auth API boundary calls and token-storage constraints;
- documentation updates.

## Explicit exclusions

This slice does not add:

- Route Handlers or Server Actions for auth or business domains;
- business logic in Next.js;
- financial calculations;
- OpenAPI contract changes;
- SQL migrations;
- direct database access;
- JavaScript-readable refresh-token storage;
- LocalStorage, SessionStorage, or IndexedDB persistence for business data, access tokens, or
  refresh tokens;
- cookie or other JavaScript-readable durable access-token storage;
- email verification, SMTP, OAuth, passkeys, or 2FA;
- provider integrations;
- workers;
- tax logic;
- mobile implementation;
- AI functionality;
- Stage 3.13 work.

## Work completed in this branch

- Added a typed frontend auth API boundary for register, login, refresh, and logout.
- Added in-memory Web session state and an authenticated shell.
- Added session-operation generation guards so stale refresh responses cannot restore
  authorization after logout.
- Gated the existing portfolio/import UI behind the authenticated shell.
- Propagated bearer access tokens to portfolio, transaction, and import calls.
- Ignored obsolete portfolio/detail loads and stale-token callbacks so old bearer requests cannot
  overwrite state after a refreshed token starts a newer load.
- Sent CSRF tokens and credentialed requests for refresh/logout without exposing refresh tokens to
  JavaScript.
- Kept live registration/login credential fields empty; demo credentials are not prefilled into the
  authentication form.
- Added local CORS credentials support so the existing HttpOnly refresh cookie can be used by the
  browser.
- Added frontend tests for direct Go API auth calls, exact auth payloads, CSRF wiring, bearer
  propagation across every authenticated frontend API method, and absence of access-token/refresh-token
  browser storage usage in the API client.
- Added session-state tests for refresh rotation, refresh failure cleanup, logout superseding older
  refresh operations, shell controller behavior, operation completion, and stale portfolio load guards.

## Verification so far

- `frontend-next`: `corepack pnpm run typecheck`
- `frontend-next`: `corepack pnpm run test`
- `backend-go`: `GOCACHE=/private/tmp/openinvest-gocache go test ./cmd/api ./internal/httpapi`
- Repository root: `GOCACHE=/private/tmp/openinvest-gocache UV_CACHE_DIR=/private/tmp/openinvest-uv-cache pnpm run verify`

## Review focus

Review must verify:

- no refresh token is exposed in JSON handling or JavaScript-readable storage;
- access token state remains bounded to active in-memory browser state;
- no access token is persisted to LocalStorage, SessionStorage, IndexedDB, cookies, logs, or durable
  frontend state;
- refresh/logout preserve the Stage 3.11 CSRF and cookie boundary;
- portfolio/import calls use the Go API as the only business authority;
- no Next.js Route Handlers, Server Actions, direct datastore access, or financial logic were added;
- the CORS credentials change is limited to approved local Web origins.

## Recommended next step

Stage 3.12 is closed. Do not start Stage 3.13 implementation until a separate Stage 3.13 planning
scope is documented, reviewed, approved, and merged.
