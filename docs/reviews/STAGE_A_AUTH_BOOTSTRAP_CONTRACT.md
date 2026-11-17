OpenInvest — Stage A: Authentication Session Bootstrap Contract
Document type: Architecture and Security Review

Status: DRAFT — INDEPENDENT SECURITY REVIEW REQUIRED

Baseline branch: develop

Baseline commit: bfb1f1e0

Working branch: design/auth-session-bootstrap

Implementation: NOT STARTED

Merge authorization: NOT GRANTED

1. Objective
   Restore an existing authenticated browser session after a full page reload without requiring the user to sign in again.

The implementation must preserve the existing authentication security model, including HttpOnly refresh cookies, CSRF protection, refresh-token rotation, session revocation and in-memory access tokens.

Acceptance criteria
A valid existing session survives Command + R.
A revoked or expired session returns the user to the login screen.
Authentication tokens are not persisted in browser storage.
A transient backend failure does not falsely indicate that the session has expired.
Logout prevents older asynchronous authentication responses from restoring access.
Concurrent initialization does not unintentionally invalidate refresh-token rotation.
The authenticated user identity is restored correctly. 2. Current implementation
Frontend
Relevant files:

frontend-next/src/features/auth/components/AuthShell.tsx
frontend-next/src/features/auth/session.ts
frontend-next/src/common/api/openinvest.ts
AuthShell initializes authentication state using anonymousState().

The client currently supports two authentication states:

anonymous
authenticated
There is no dedicated initialization state or session bootstrap request.

The access token and CSRF token are stored in React memory. A full page reload destroys this state.

Backend
Relevant files:

backend-go/internal/httpapi/auth_handlers.go
backend-go/internal/httpapi/routes.go
backend-go/internal/httpapi/cors.go
backend-go/internal/auth/service.go
backend-go/internal/auth/models.go
backend-go/internal/postgres/auth_store.go
The backend currently exposes:

POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
Refresh requires:

A valid oi_refresh cookie.
An X-CSRF-Token request header.
The refresh operation rotates the server-side session credentials and returns a new access token and CSRF token.

The existing HTTP refresh response returns session data but does not include the authenticated user profile.

Existing cookie contract
Name: oi_refresh
HttpOnly: enabled
SameSite: Strict
Path: /api/v1/auth
Secure: controlled by backend configuration
Default refresh lifetime: 30 days
The default access token lifetime is 15 minutes.

These values describe the current implementation and are not proposed changes.

3. Root cause
   After browser reload:

React creates a new anonymous authentication state.
The previous access token is no longer available.
The previous CSRF token is no longer available.
The HttpOnly refresh cookie may still exist.
The existing refresh endpoint requires the missing CSRF token.
The application displays the login screen without attempting a valid restoration flow.
The problem is caused by the missing session initialization contract, not by a requirement to persist access tokens.

4. Candidate architecture A — Read-only session bootstrap
   Proposed endpoint:

GET /api/v1/auth/bootstrap

The endpoint would validate the current refresh cookie and retrieve the associated active session and authenticated user.

A possible successful response would contain:

Authenticated user information.
A new short-lived access token.
CSRF material required for subsequent refresh and logout operations.
The endpoint must not rotate the refresh token.

Advantages
One initial network request.
No refresh-token rotation on every page reload.
Reduced chance of concurrent rotation conflicts.
Direct restoration of the complete frontend authentication state.
Security concerns
Issuing authentication material from a cookie-authenticated GET endpoint requires explicit security review.
Response confidentiality must not depend solely on CORS.
The server must enforce an appropriate origin policy.
The endpoint must reject revoked, expired and otherwise invalid sessions.
Responses must not be cached.
Browser navigation and cross-origin request behavior require explicit analysis.
Decision: NOT APPROVED.

5. Candidate architecture B — CSRF bootstrap followed by refresh
   Proposed flow:

The frontend starts in an initializing state.
The frontend requests a session-scoped CSRF bootstrap.
The backend validates the cookie and session.
The frontend receives the CSRF context.
The frontend calls POST /api/v1/auth/refresh.
The existing refresh operation rotates the session.
The frontend restores the authenticated state.
The existing refresh response must be extended or supplemented to restore authenticated user information.

Advantages
Reuses the existing refresh-token rotation mechanism.
Retains existing session revocation behavior.
Separates session initialization from token rotation.
Security concerns
Two network requests are required.
Multiple browser operations may attempt refresh simultaneously.
An initial CSRF endpoint requires an explicit request-origin and credential policy.
Race conditions between bootstrap, refresh and logout require dedicated handling.
The authenticated user must be restored without relying on stale client information.
Decision: NOT APPROVED.

6. Comparison
   | Criterion | Candidate A | Candidate B |

| ------------------------------------- | --------------------- | ------------------------------------------------ |

| Initial requests | One | Two |

| Refresh-token rotation during startup | No | Yes |

| Existing refresh service reuse | Partial | Direct |

| Concurrency complexity | Lower | Higher |

| Complete user restoration | Requires new response | Requires response extension or additional lookup |

| Security review required | Yes | Yes |

Neither candidate is approved until its session validation, origin policy, response contract and concurrency behavior have been verified.

7. Required backend guarantees
   The approved design must ensure that:

The backend validates the refresh token against server-side session state.
Expired and revoked sessions are rejected.
A valid user is retrieved from authoritative server-side records.
Authentication material is never returned for an invalid session.
Raw refresh tokens are not exposed to JavaScript.
CSRF protection remains effective.
Credentialed requests use an explicit origin policy.
Duplicate or malformed Origin headers are rejected.
Authentication responses use Cache-Control: no-store.
Sensitive values are not written to application logs.
Existing rate limits and session revocation behavior are preserved.
Session validation and rotation do not introduce database consistency violations. 8. Required frontend state machine
Proposed authentication states:

initializing — session restoration is in progress.
authenticated — a verified user and active session are available.
anonymous — no valid authenticated session exists.
temporarily_unavailable — session status could not be determined because of a transient failure.
Initialization behavior
On application startup:

Enter initializing.
Attempt the approved bootstrap operation.
On confirmed authentication, enter authenticated.
On confirmed session rejection, enter anonymous.
On a transient failure, enter temporarily_unavailable.
The frontend must not display protected user content before authentication has been verified.

Logout behavior
Logout must invalidate pending initialization operations.

A bootstrap response from an older operation must not overwrite a newer logout or login.

React lifecycle
The implementation must handle:

React Strict Mode effect replay.
Component unmounting.
Stale asynchronous responses.
Multiple simultaneous refresh attempts.
Identity changes during initialization. 9. HTTP failure semantics
The final contract must distinguish:

| Condition | Required frontend behavior |

| ----------------------------- | ------------------------------------------------------------------- |

| Confirmed invalid session | Display Login |

| Confirmed expired session | Display Login |

| Confirmed revoked session | Display Login |

| HTTP 403 | Follow documented authorization policy |

| HTTP 429 | Preserve uncertain authentication status and allow controlled retry |

| HTTP 500 | Show temporary service failure |

| HTTP 503 | Show temporary service failure |

| Network failure | Do not treat as confirmed logout |

| Request timeout | Do not treat as confirmed logout |

| Malformed successful response | Treat as contract failure |

Exact HTTP status mappings and response schemas remain subject to review.

10. Required automated tests
    Backend
    Valid active session.
    Missing refresh cookie.
    Expired refresh token.
    Revoked session.
    Invalid CSRF context.
    Invalid Origin.
    Duplicated Origin.
    Concurrent refresh operations.
    Refresh-token replay.
    Bootstrap during logout.
    User identity consistency.
    No-store response headers.
    No authentication-token disclosure on error.
    Frontend
    Successful session restoration after full reload.
    Anonymous startup without credentials.
    Temporary backend outage.
    HTTP 401 and HTTP 403 handling.
    HTTP 429 and HTTP 503 handling.
    React Strict Mode repeated initialization.
    Logout during pending bootstrap.
    Login during pending bootstrap.
    Stale response rejection.
    No authentication-token persistence in browser storage.
    Correct authenticated user restoration.
    Integration
    Login, reload, refresh and logout sequence.
    Multiple reloads with an active session.
    Concurrent requests in one browser tab.
    Separate browser tabs using the same session.
    Session expiry during initialization.
    Server restart with persistent PostgreSQL session state.
11. OpenAPI requirements
    The approved bootstrap operation must have a documented OpenAPI contract.

The contract must define:

Endpoint path and HTTP method.
Cookie authentication requirements.
Origin and CSRF requirements.
Success response schema.
Error response schemas.
Cache-control requirements.
Authenticated user representation.
Session restoration and refresh semantics.
Existing authentication API behavior must remain compatible unless an independently approved contract change requires otherwise.

12. Explicit non-goals
    This stage does not introduce:

OAuth or third-party identity providers.
Persistent access tokens in localStorage.
Persistent refresh tokens in JavaScript.
A new password authentication mechanism.
Financial transaction changes.
Ledger or portfolio schema changes.
Unrelated interface redesign.
Deployment configuration changes without separate review. 13. Security review questions
The independent reviewer must determine:

Which candidate preserves the existing authentication guarantees with the smallest justified change?
Can the bootstrap endpoint expose authentication material to an unintended origin?
How should requests with a missing Origin header be handled?
Is read-only session validation possible with existing database operations?
Does Candidate A require a new store interface?
Can Candidate B reuse refresh-token rotation without concurrency regressions?
How should multiple browser tabs coordinate session restoration?
How will user identity be included in the restored authentication state?
How will temporary failures differ from confirmed invalid sessions?
How will logout invalidate stale bootstrap responses?
Confirmed PostgreSQL findings: The existing session store retains hashed refresh and CSRF credentials, session state, expiry and family identity. Rotation uses transactional locking. No read-only session lookup is exposed through the current auth.Store interface. A bootstrap contract must not assume that the original CSRF token can be recovered from its hash.

14. Authorization gates
    Contract review: PENDING

Independent security review: PENDING

Architecture decision: PENDING

Implementation authorization: NOT GRANTED

Commit authorization: NOT GRANTED

Push authorization: NOT GRANTED

Merge authorization: NOT GRANTED

No production authentication behavior may be modified as part of this draft.

15. Current conclusion
    The missing browser session bootstrap is the confirmed root cause of the repeated login screen after a full reload.

Both proposed architectures are technically plausible, but neither has completed the required security and concurrency review.

The next step is to verify the exact server-side session lookup and rotation semantics, finalize the bootstrap contract, and submit the architecture for independent review.
