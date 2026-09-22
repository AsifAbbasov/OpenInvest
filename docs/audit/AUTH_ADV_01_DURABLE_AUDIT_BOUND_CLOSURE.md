# AUTH-ADV-01 - Durable Auth Audit Bound Closure

Status: CLOSED
Severity: P2
Technical PR: #203
Final squash: 40010f4c7513dab99a31ab85c879242a77228b7a

## Problem

Rejected authentication and session traffic could amplify durable `audit.events`
storage. The original implementation appended audit evidence for rejection paths
without a durable per-known-session/event bound.

## Root cause

Missing, anonymous, and unknown refresh/logout traffic could spend durable audit
storage although there was no known session identity worth retaining as forensic
evidence. Known stale or revoked sessions also produced equivalent security
evidence repeatedly for refresh replay and repeated logout.

The append-only audit model preserved each occurrence. The HTTP rate limiter
bounded requests only, was process-local, and did not establish a cumulative
durable storage bound.

## Failure scenario

### Scenario 1 - anonymous / unknown traffic

Repeated refresh/logout requests with a missing refresh cookie, missing CSRF, or
a completely unknown refresh/CSRF pair could create durable rejection evidence.
This allowed rejected traffic to cause unbounded cumulative `audit.events`
growth without a known session identity.

### Scenario 2 - repeated known-session evidence

A previously known stale or revoked session could be replayed repeatedly, such
as reusing old refresh credentials or logging out repeatedly against the same
revoked session. Without deduplication, one underlying security event could
append one new row per request.

## Impact

Durable audit storage growth could remain proportional to repeated rejected
requests, including traffic that carried no meaningful forensic identity.

## Initial solution

N/A - no remediation diff was accepted before the final design.

## Why review rejected it

N/A - no implemented remediation was rejected. Design review established that
removing writes only for missing or anonymous credentials would leave the
known stale/revoked-session replay and logout amplification path unresolved.

## Final remediation

Missing refresh tokens, missing CSRF, completely unknown refresh/CSRF pairs,
and completely unknown logout pairs create zero durable auth audit rows.

Security-relevant known-session rejection and replay evidence is retained.
`recordDeduplicatedAuthSecurityAudit` writes a gate to
`audit.auth_security_event_deduplications` and uses PostgreSQL
`INSERT ... ON CONFLICT DO NOTHING` before appending failure evidence. The
unique index on `(action_code, session_id)` gives one durable bound for each
known session and security action across application instances.

Migration `000013_auth_adv_01_audit_deduplication` creates the gate relation
and its unique index. The runtime role receives `INSERT` plus `SELECT` only on
the conflict-target columns; `audit.events` remains append-only, with no
`UPDATE`, `DELETE`, or `TRUNCATE` permission added.

## Why this solution

- Anonymous and unknown rejection noise spends no durable audit storage.
- Meaningful evidence for a known session is retained once per action/session.
- PostgreSQL, rather than in-memory rate-limit state, coordinates the bound
  across processes and application instances.
- Normal refresh/logout auditing and session-family containment remain intact.

## Regression evidence

Merged PostgreSQL integration coverage verifies:

- missing refresh and missing CSRF create zero durable audit evidence;
- unknown refresh and logout pairs create zero durable audit evidence;
- known wrong-CSRF refresh/logout evidence is bounded and does not rotate,
  revoke, or revoke a session family;
- repeated refresh replay and repeated revoked logout evidence are bounded;
- concurrent requests and two-store, multi-instance-equivalent PostgreSQL
  access preserve the bound;
- valid refresh and logout auditing, session-family containment, and
  refresh/logout race behavior are preserved;
- runtime audit privileges remain append-only.

## CI / review evidence

Technical PR: #203
Technical source head: `6de40019dc6a60f12c119dfb82a9f7a138af780e`
Accepted protected CI run: `35792016376` (10/10 SUCCESS)
Final squash: `40010f4c7513dab99a31ab85c879242a77228b7a`
Final tree: `76ebf088a3e1221d02c4bbc96a5a9c216653b614`

Protected PostgreSQL migration CI validated migration files, apply, disposable
DOWN/reapply, and runtime append-only privileges. There was no separate
workflow run on the final squash commit. Post-merge verification instead
confirmed the final develop SHA, expected parent and tree, exact approved
content, 14 approved changed files, and the same 10 protected contexts.

## Residual limitations

AUTH-ADV-02 remains P3 deferred hardening. Application rate limiting remains a
separate control; this remediation specifically bounds durable audit
amplification. Anonymous rejected requests intentionally receive no durable
per-request audit rows. Repository-wide audit remains ongoing.

## Related dispositions

AUTH-ADV-01: CLOSED
AUTH-ADV-02: P3_DEFERRED_PRODUCT_SECURITY_HARDENING
AUTH_23_LOG_TOKEN_LEAK: PASS
AUTH_SESSION_PRIVACY_AUDIT: still ongoing until independent closure review
REPOSITORY_WIDE_AUDIT: ONGOING

AUTH-ADV-02 concerns immediate active registration without email ownership
verification and distinguishable duplicate-email conflict. Stage 3.11 defined
email verification and SMTP as non-goals. Residual topics are registration
membership enumeration, pre-registration namespace/account squatting, and
future email ownership verification.

AUTH_23_LOG_TOKEN_LEAK scope: the local repository logging/output review found
no reachable sink disclosing passwords or password hashes, refresh/access
tokens, CSRF values or hashes, Authorization/Cookie/Set-Cookie values, or
database credentials in the reviewed reachable logging/output surfaces. This
does not guarantee that future code cannot introduce a secret leak.

This document closes only AUTH-ADV-01. The broader Auth/Session/Privacy audit
can close only after independent review of this document confirms that no other
P2+ finding remains. It does not claim that the repository-wide audit is
complete.
