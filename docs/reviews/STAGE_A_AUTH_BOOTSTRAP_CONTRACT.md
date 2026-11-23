# OpenInvest — Stage A.1: Authentication Bootstrap Contract

**Status:** APPROVED
**Architecture:** Candidate C — Atomic CSRF-Only Bootstrap
**Branch:** `design/auth-session-bootstrap`
**Implementation:** NOT STARTED
**Commit / Push / Merge:** NOT AUTHORIZED

## 1. Objective

Restore an authenticated session after a full browser reload without requiring another login.

Preserve HttpOnly refresh cookies, CSRF protection, refresh-token rotation, replay detection and in-memory access tokens.

## 2. Root Cause

`AuthShell.tsx` initializes with `anonymousState()`.

After page reload, React loses the access token, CSRF token and authenticated user.

The HttpOnly refresh cookie remains available, but `/api/v1/auth/refresh` requires the missing CSRF token.

## 3. Approved Architecture

**POST `/api/v1/auth/bootstrap`**

1. Validate Origin, Fetch Metadata and the required custom header.
2. Read the existing `oi_refresh` cookie.
3. Generate a new cryptographically random CSRF token.
4. Start a PostgreSQL transaction.
5. Validate the refresh session and its owner.
6. Atomically replace `csrf_token_hash`.
7. Retrieve `StoredUser`.
8. Commit the transaction.
9. Generate a new short-lived access token.
10. Return `AuthData` containing `user` and `session`.

Bootstrap must not rotate the Refresh Token, create another session, extend expiration or send `Set-Cookie`.

## 4. Go and PostgreSQL Contract

Add to `auth.Store`:

```go
BootstrapSession(
    ctx context.Context,
    refreshTokenHash string,
    nextCSRFTokenHash string,
) (StoredUser, error)
```

Add to `auth.Service`:

```go
Bootstrap(
    ctx context.Context,
    refreshToken string,
) (AuthResult, error)
```

**Required transaction order:**

```text
BEGIN
  Lookup session owner by refresh hash
  Acquire user advisory lock
  SELECT current session FOR UPDATE
  Verify owner, expiry and session_state = 'active'
  UPDATE csrf_token_hash
  Load authoritative StoredUser
COMMIT
```

Only `csrf_token_hash` changes.

Do not modify `refresh_token_hash`, `session_family_id`, `expires_at` or `last_rotated_at`.

Preserve existing `RotateSession`, `RevokeSession` and replay-detection behavior.

**Database migration:** Not required by the approved design.

## 5. HTTP Contract

**Request:**

```http
POST /api/v1/auth/bootstrap
Origin: <approved-origin>
Accept: application/json
X-OpenInvest-Bootstrap: 1
Cookie: oi_refresh=<http-only-cookie>
```

Request body must be empty.

**Success:** HTTP 200 with the existing OpenInvest response envelope containing:

```json
{
  "user": "AuthUser",
  "session": {
    "accessToken": "...",
    "accessTokenExpiresAt": "...",
    "csrfToken": "..."
  }
}
```

The example describes the `data` portion of the response.

**Errors:**

| HTTP | Condition                                        |
| ---- | ------------------------------------------------ |
| 400  | Invalid bootstrap request                        |
| 401  | Missing, expired or revoked session              |
| 403  | Invalid or missing Origin                        |
| 429  | Rate limit exceeded                              |
| 500  | Unexpected internal error                        |
| 503  | Authentication service or PostgreSQL unavailable |

Requirements:

- Exact Origin allowlist; reject duplicate, null and malformed values.
- Reject cross-site requests through Fetch Metadata checks.
- Require the bootstrap custom header.
- Preserve `SameSite=Strict`, HttpOnly and Secure cookie policy.
- Use `Cache-Control: no-store`.
- Preserve request and trace identifiers.
- Reuse authentication rate limiting.
- Document the endpoint in OpenAPI.

## 6. Frontend State Machine

```ts
type AuthState =
  | { status: "initializing" }
  | { status: "authenticated"; user: AuthUser; session: AuthSession }
  | { status: "anonymous"; message: string | null }
  | { status: "temporarily_unavailable"; message: string };
```

On application startup:

- Start in `initializing`.
- Execute a single bootstrap operation.
- On HTTP 200, restore user and session.
- On confirmed HTTP 401, display Login.
- On temporary failures, display a retry state.
- Do not render protected content before successful authentication.

Use generation counters to reject stale responses.

Login and Logout must invalidate pending bootstrap results.

Do not persist tokens in browser storage.

## 7. Multi-Tab and Replay Protection

Coordinate bootstrap, refresh and logout through Web Locks and BroadcastChannel.

Never broadcast authentication tokens.

Required scenarios:

- Two simultaneous bootstrap requests.
- Bootstrap versus refresh.
- Bootstrap versus logout.
- Two concurrent refresh requests.
- React Strict Mode repeated initialization.
- Account switching.
- Lost BroadcastChannel notifications.
- Missing Web Locks support.
- Lost refresh response.

**Critical rule:** Never automatically repeat an uncertain refresh operation using potentially outdated credentials.

An outdated CSRF token may require controlled bootstrap recovery.

Existing refresh-token replay detection and family revocation must remain unchanged.

## 8. Required Tests

**Go:** Session validation, CSRF generation, token hashing, authentication failures and correct user restoration.

**PostgreSQL:** Atomic update, expiry, revocation, lock ordering, concurrent bootstrap, refresh/logout races and rollback.

**HTTP:** Origin validation, request headers, response envelope, error statuses, no-store, rate limits and absence of Set-Cookie.

**React:** Reload restoration, initialization, temporary failures, stale responses, Strict Mode, Login and Logout.

**Browser:** Command + R, multiple tabs, lost responses, cookie behavior and authentication lifecycle.

**OpenAPI:** Endpoint schema and contract validation.

## 9. Finding Closure

| Finding                               | Status             |
| ------------------------------------- | ------------------ |
| AUTH-BOOT-01 — CSRF restoration       | RESOLVED IN DESIGN |
| AUTH-BOOT-02 — Multi-tab and replay   | RESOLVED IN DESIGN |
| AUTH-BOOT-03 — Origin validation      | RESOLVED IN DESIGN |
| AUTH-BOOT-04 — Atomic Store operation | RESOLVED IN DESIGN |
| AUTH-BOOT-05 — HTTP and Frontend      | RESOLVED IN DESIGN |

**Independent architecture review:** APPROVED.

## 10. Stage A.2 Acceptance

Implementation is ready for independent code review when:

1. Backend and Frontend implement the approved contract.
2. Refresh-token replay protection remains unchanged.
3. PostgreSQL concurrency tests pass.
4. Go and TypeScript tests pass.
5. Next.js production build passes.
6. OpenAPI validation passes.
7. Browser tests confirm session restoration after Command + R.
8. Untested scenarios and remaining risks are documented.

**Implementation:** NOT STARTED
**Implementation review:** PENDING
**Commit / Push / Merge:** NOT AUTHORIZED
