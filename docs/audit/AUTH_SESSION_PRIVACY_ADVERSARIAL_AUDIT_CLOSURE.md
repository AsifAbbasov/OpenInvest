# Auth / Session / Privacy Deep Adversarial Audit Closure

Status: CLOSED_WITH_P3_RESIDUAL
Scope: Auth / Session / Privacy Deep Adversarial Audit
Baseline: `711c027c60f1e5e6011a737a33151f0c0dbffda7`

## Scope

This record closes the completed Auth / Session / Privacy adversarial-audit
module. It consolidates confirmed findings, verified checks, and the remaining
product-security hardening item. It does not close the repository-wide audit.

## Findings

### AUTH-ADV-01 - CLOSED

Severity: P2
Disposition: CLOSED

The confirmed problem was durable authentication-audit storage amplification.
Anonymous, missing, or unknown rejection paths could create durable evidence,
and a known stale or revoked session could repeatedly create equivalent
security evidence for refresh replay or repeated logout.

The final remediation writes zero durable auth-audit rows for unknown and
anonymous paths. It keeps meaningful known-session security evidence, bounded
by PostgreSQL deduplication. `recordDeduplicatedAuthSecurityAudit` first writes
to `audit.auth_security_event_deduplications` with `INSERT ... ON CONFLICT DO
NOTHING`; the unique `(action_code, session_id)` key bounds evidence for each
known session and action across application instances.

Known wrong-CSRF session evidence remains preserved and bounded. The runtime
role retains append-only audit behavior; no `UPDATE`, `DELETE`, or `TRUNCATE`
audit privilege was added.

Technical PR: #203
Technical squash: `40010f4c7513dab99a31ab85c879242a77228b7a`
Closure PR: #204
Closure squash: `711c027c60f1e5e6011a737a33151f0c0dbffda7`
The historical Stage 3.11 requirement to audit every rejected refresh/logout
attempt is explicitly superseded by AUTH-ADV-01's durable-audit bound.

### AUTH-ADV-02 - P3 DEFERRED

Severity: P3
Disposition: DEFERRED_PRODUCT_SECURITY_HARDENING

Residual concerns are registration membership enumeration, pre-registration
namespace or account squatting, and future email ownership verification.
Stage 3.11 explicitly defined email verification and SMTP as non-goals.

This record does not implement email verification, SMTP, CAPTCHA, OAuth,
passkeys, or two-factor authentication. AUTH-ADV-02 is not a P2 blocker.

### AUTH_23 - PASS

No reachable logging/output sink was found in the reviewed repository paths
that disclosed passwords/password hashes, refresh/access tokens, CSRF
values/hashes, Authorization/Cookie/Set-Cookie values, or database credentials.
This bounded review result does not guarantee that future code cannot introduce
a credential or secret leak.

## Attack-surface coverage

The completed review covered the following verified auth, session, and privacy
areas. This is a coverage statement, not a claim that all future variations are
safe without continued review.

- refresh-token rotation and replay containment;
- logout and refresh concurrency behavior;
- CSRF enforcement, including missing, unknown, and known wrong-CSRF paths;
- session-family containment and known-session rejection behavior;
- cross-user authorization and reviewed IDOR boundaries;
- principal-scoped idempotency where it was reviewed;
- JWT and session handling, cookie handling, and authentication error handling;
- rate-limit and client-IP/trusted-proxy boundaries;
- Argon2 admission and timing-handling boundaries;
- runtime PostgreSQL least privilege and append-only audit permissions;
- durable security-audit amplification; and
- logging and credential-leakage review.

## Evidence

Technical PR #203 passed protected PR CI run `35792016376` with 10 successful
contexts. Its PostgreSQL coverage exercised durable audit behavior, concurrent
requests, multi-store behavior, append-only privileges, and preserved refresh
and logout semantics.

Closure PR #204 passed protected PR CI run `35795618577` with 10 successful
contexts. Post-merge verification confirmed the expected parent, one-file
closure change, approved 140-line document, and exact content SHA-256.
No separate workflow run exists for the final closure squash SHA; this record
does not claim post-merge CI that did not run.

## Residual limitations

AUTH-ADV-02 remains deferred P3 product-security hardening. Application rate
limiting is a separate control from the durable audit bound. Anonymous rejected
requests intentionally have no durable per-request security-audit row.

Future product work may revisit email ownership verification and the related
registration-abuse concerns when that scope and delivery infrastructure exist.

## Final disposition

There are no open P0, P1, or P2 findings in this audit module. AUTH-ADV-01 is
closed; AUTH-ADV-02 remains a deferred P3 item. This module closure does not
mean that the OpenInvest repository-wide audit is complete.

AUTH_SESSION_PRIVACY_AUDIT: CLOSED_WITH_P3_RESIDUAL
P0_OPEN: 0
P1_OPEN: 0
P2_OPEN: 0
AUTH_ADV_01: CLOSED
AUTH_ADV_02: P3_DEFERRED_PRODUCT_SECURITY_HARDENING
AUTH_23_LOG_TOKEN_LEAK: PASS
REPOSITORY_WIDE_AUDIT: ONGOING
