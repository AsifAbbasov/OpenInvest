# Stage 3.28 — Authentication Security Remediation

| Field | Value |
| --- | --- |
| Status | Closure record; when canonical on `develop`, Stage 3.28 is closed for P1-01 and P1-05 |
| Owner | Principal Architect |
| Implementation baseline | `develop` at `6f50c39ee19834f1b2ff354230b727019be64369` |
| Implementation branch | `fix/stage-03-28-auth-security-remediation` |
| Implementation PR | #59 |
| Implementation merge | `dc83f5f3a11da164e6809593861d96ccf47b29ca` |
| Reviewed exact head | `92edab5d3e93dafe2fcc6247644e38e878a4202f` |
| Exact-head CI | GitHub Actions #114 — SUCCESS |
| Independent final review | `APPROVED` on exact head `92edab5d3e93dafe2fcc6247644e38e878a4202f` after one governance-only `REQUEST CHANGES` correction |
| Human implementation merge authorization | Explicit squash-merge approval on 2026-08-23 |
| Trigger | Repository-audit findings P1-01 and P1-05 |
| Scope | Refresh-token replay/session-family containment; refresh/logout concurrency; Argon2 resource admission; PostgreSQL defense in depth; regression coverage |
| Out of scope | Stage 3.25 privacy work; P2/P3 findings; product expansion; frontend features; broker/provider work; unrelated architecture changes |

## Closure statement

Stage 3.28 is the narrowly scoped remediation of the two remaining repository-audit P1 findings:

- P1-01 — refresh-token replay did not invalidate the affected session family;
- P1-05 — Argon2 password work had no process-wide admission bound and could create avoidable memory/CPU burst risk.

Implementation PR #59 was squash-merged into `develop` at
`dc83f5f3a11da164e6809593861d96ccf47b29ca` after exact-head CI #114 passed on
`92edab5d3e93dafe2fcc6247644e38e878a4202f`, renewed independent final review returned
`APPROVED` on that same exact head, and explicit human squash-merge authorization was given.

This document is the closure-governance record. Once it is itself canonical on `develop`,
P1-01 and P1-05 are closed. Stage 3.25 privacy evidence planning remains separate and the
repository-audit P2/P3 backlog remains open for later remediation.

## P1-01 — refresh-token replay and session-family containment

### Root cause

Session rotation persisted independent session rows but no family lineage. The refresh store
selected only active sessions, so a replayed revoked token was rejected without identifying and
invalidating its active descendants.

### Failure mode and impact

The replayed old token did not itself become valid again. The defect was containment failure after
replay detection: the system could reject the stale token while leaving a newer descendant refresh
session active. In a stolen-token scenario, replay is a strong compromise signal, so allowing the
family descendant to survive extends the lifetime of a potentially compromised authenticated
session.

Row locking only the presented session was also insufficient for races involving different rows.
An old-token replay could race a refresh of an already-issued descendant, and logout could race a
refresh that creates another descendant. Without one serialization boundary, a new active session
could be committed after the revocation decision.

For historical sessions, the database did not contain enough information to reconstruct which rows
belonged to the same original login family. Guessing that lineage would risk either under-revocation
or silently coupling independent historical logins.

### Remediation

Migration `000005_stage_03_28_auth_security` adds persisted nullable `session_family_id` and
family/state/expiry lookup support without changing historical migration files.

Historical rows intentionally keep `session_family_id = NULL`; Stage 3.28 does not fabricate
lineage that was never persisted. New registration/login sessions use their own session id as the
family root and rotated descendants inherit that family.

A PostgreSQL insert guard rejects newly created session rows without family identity while still
allowing pre-Stage-3.28 legacy rows to remain nullable.

Refresh and logout mutation paths resolve the owning user and acquire one transaction-scoped
advisory lock for that user before row locking. This establishes one ordering boundary for related
token mutations.

For post-Stage-3.28 replay, a known revoked token revokes every active member of its family before
returning `ErrInvalidSession`. For a legacy replay whose family is unknowable, the store fails
closed by revoking all active sessions for that user.

Logout uses the same serialization boundary. A known stale/revoked token can contain its family,
preventing refresh-versus-logout races from leaving a newly rotated descendant alive.

### Why this remediation was chosen

Persisted family identity is the smallest durable piece of state that lets the system answer the
security question raised by replay: “which active descendants belong to the compromised login?” It
allows fail-closed family revocation without revoking unrelated post-Stage-3.28 login families.
Using the root session id as the family id avoids introducing a second synthetic identity lifecycle
while keeping lineage stable across rotations.

The migration deliberately leaves legacy `session_family_id` values `NULL`. Historical lineage was
never stored, so a backfill would manufacture security facts that cannot be proven from the data.
When replay is detected for such a legacy session, revoking all active sessions for that user is the
conservative fail-closed choice: it may inconvenience the user, but it does not claim false lineage
or leave a potentially compromised descendant alive.

The transaction-level advisory lock is taken at user scope before row locks because replay, refresh,
and logout can operate on different session rows. A lock on only the presented row cannot serialize
an operation that is concurrently creating or mutating a descendant row. User scope also works for
legacy sessions where family scope is unknown. A consistent `user advisory lock -> session row`
ordering gives all related mutation paths the same concurrency boundary.

The PostgreSQL insert guard is defense in depth. Application code already creates family identity,
but the database must reject a future direct SQL path, maintenance path, or application regression
that tries to create a new family-less session and would otherwise reintroduce the original defect.

### Alternatives rejected and why

| Alternative | Why it was rejected |
| --- | --- |
| Revoke only the replayed token | The replayed token is already revoked; this does nothing to the active descendant that represents the containment risk. |
| Treat every session for a user as one permanent family | This over-revokes independent logins and destroys useful family isolation for new sessions. |
| Backfill legacy rows with guessed family ids | Historical lineage cannot be reconstructed reliably; guessed security lineage is worse than explicitly unknown lineage. |
| Use only `SELECT ... FOR UPDATE` on the presented session row | Replay/refresh/logout races can involve different rows, including a newly created descendant, so one-row locking does not serialize the security decision. |
| Use an in-process mutex | It would not protect multiple API instances and would move a persistence/concurrency invariant outside PostgreSQL. |
| Rely only on serializable transaction isolation | It would broaden retry/abort behavior across the store and still require careful handling of legacy scope; the advisory lock expresses the exact mutation boundary directly. |
| Ignore legacy replay because family is unknown | This preserves availability at the cost of leaving potentially compromised descendants active; the security requirement is fail-closed containment. |

## P1-05 — bounded Argon2 resource admission

### Root cause

Argon2id correctly used 64 MiB, `t=3`, `p=1`, but hash, verification, and dummy-verification paths
had no shared process-wide admission budget.

### Failure mode and impact

Each approved Argon2id derivation reserves roughly 64 MiB of working memory in addition to normal
process overhead. Without a shared admission bound, distributed concurrent registration/login
requests can multiply that cost by the number of simultaneous derivations and create avoidable RAM
pressure, CPU saturation, garbage-collection pressure, latency collapse, or process termination.

Per-client HTTP rate limiting does not fully solve this class of problem because an attacker can
distribute requests across many clients or addresses. A limiter that covers only successful-user
verification would also be bypassable through registration or unknown-user dummy verification.

There was a second amplification concern: if an encoded stored password hash were accepted with
larger-than-approved memory/time parameters, verification could perform more expensive work than
the service budget intended before deciding whether the credential is valid.

### Remediation

The approved Argon2id cost is unchanged.

All hash, verify, and dummy paths use one process-wide fail-fast admission gate capped at two
simultaneous Argon2 derivations. When both slots are occupied, additional expensive work returns
`ErrAuthCapacity`; it does not wait in an unbounded request/goroutine queue.

Stored encoded password hashes are accepted for expensive verification only when memory, time,
thread, salt-length, and hash-length parameters exactly match the approved budget. Over-budget or
malformed encodings are rejected before Argon2 runs.

The independent reviewer explicitly treated the current generic HTTP 500 mapping for
`ErrAuthCapacity` as non-blocking for P1-05 because the resource-exhaustion condition is prevented
by the process-wide fail-fast gate itself. A dedicated `503 + Retry-After` mapping remains optional
HTTP-contract hardening and is not represented as part of this P1 closure.

### Why this remediation was chosen

The solution protects availability without weakening password security. Lowering Argon2 memory or
time cost would reduce resource usage by making password derivation cheaper for both the server and
an offline password attacker. Stage 3.28 therefore keeps the approved 64 MiB / `t=3` / `p=1` cost
and controls concurrency instead.

A process-wide capacity gate matches the actual resource boundary: Argon2 memory is consumed inside
one API process. Capping simultaneous derivations at two bounds the core Argon2 working-set demand
to roughly 128 MiB per API process, plus implementation/runtime overhead, rather than allowing it to
grow linearly with request concurrency.

The gate is fail-fast rather than a blocking semaphore queue. A blocking semaphore would cap active
Argon2 memory but could still accumulate an unbounded number of waiting HTTP requests/goroutines,
turning the memory-amplification problem into queue, socket, and latency exhaustion. Immediate
`ErrAuthCapacity` rejection keeps the admission decision bounded.

Hash, normal verify, and dummy verify share the same gate so no authentication path can bypass the
resource budget. Strict validation of the stored Argon2 encoding before expensive work prevents the
persistence layer from becoming a cost-amplification input.

### Alternatives rejected and why

| Alternative | Why it was rejected |
| --- | --- |
| Reduce Argon2 memory/time parameters | It would trade away password-cracking resistance to solve an availability problem that can be bounded at admission instead. |
| Keep a blocking semaphore and queue excess work | Active Argon2 RAM would be bounded, but waiting requests/goroutines could grow without bound under sustained load. |
| Rely only on per-IP or per-client rate limiting | Distributed traffic can bypass a client-local quota; rate limiting is useful defense in depth but is not a process memory budget. |
| Limit only login verification | Registration, dummy verification, and other password derivation paths would remain bypasses around the budget. |
| Accept arbitrary encoded Argon2 cost parameters | A malformed or over-budget stored encoding could force work above the approved service budget before authentication fails. |
| Return `503` without an Argon2 admission gate | Better HTTP semantics do not bound memory or CPU consumption; the resource gate is the security control. |
| Introduce a distributed Redis semaphore for this finding | The immediate risk is per-process Argon2 working memory. A local process-wide gate is simpler, deterministic, and remains effective even if Redis is unavailable; cluster-wide admission can be considered separately if deployment characteristics later require it. |

## Regression evidence

| Case | Proven behavior |
| --- | --- |
| Normal refresh rotation | New descendant inherits the root family |
| Replay old rotated token | Replay rejected and active family descendants revoked |
| Two concurrent refreshes of same token | At most one rotation succeeds; replay containment removes active compromised descendants |
| Old-token replay racing descendant refresh | No active member remains in the compromised family |
| Independent post-3.28 login family | Remains independent from replay in another known family |
| Legacy token with unknown family replayed | Fails closed across active sessions for that user |
| Logout racing refresh | Logout containment leaves no active descendant |
| Direct SQL new session without family id | PostgreSQL rejects the insert |
| Migration apply / rollback / reapply | Completes cleanly without modifying historical migrations |
| Two Argon2 derivations already running | Additional expensive work fails immediately with `ErrAuthCapacity` |
| Capacity slots later released | Authentication capacity recovers |
| Stored over-budget Argon2 parameters | Rejected before expensive derivation |
| Unknown-user login dummy path at capacity | Returns capacity exhaustion rather than bypassing the gate |

## Review history

The first independent final review on head
`c0ef116dd858d0f6d2d613e5eae915c7a30d556e` returned `REQUEST CHANGES` for one governance
inconsistency only: this report still said `Draft PR #59` after GitHub had marked the PR ready for
review. The reviewer reported no additional blocking P1-01/P1-05 security or correctness issue.

The governance wording was corrected. The net diff from the originally reviewed head to the final
reviewed head was one documentation-line replacement only; runtime code did not change. CI #114
then passed completely on exact head `92edab5d3e93dafe2fcc6247644e38e878a4202f`, and the renewed
independent review returned `APPROVED`.

## Verification evidence

Exact-head CI #114 on `92edab5d3e93dafe2fcc6247644e38e878a4202f` completed successfully,
including:

- PostgreSQL migration validation;
- migration apply;
- every rollback and full reapply;
- Go tests with PostgreSQL integration tests;
- OpenAPI contract validation;
- frontend build, typecheck, and tests;
- Python tests;
- Docker Compose configuration validation.

Implementation PR #59 was then squash-merged into `develop` at
`dc83f5f3a11da164e6809593861d96ccf47b29ca`.

## Residual work

Stage 3.28 closes P1-01 and P1-05 only when this closure record becomes canonical on `develop`.

It does not close, waive, or silently downgrade:

- Stage 3.25 privacy Security Review evidence planning;
- any P2 finding from the repository audit;
- any P3 finding from the repository audit;
- optional `ErrAuthCapacity` HTTP `503 + Retry-After` contract hardening;
- unrelated product, provider, market-data, tax, mobile, AI, or architecture work.

## Closure rule

The implementation merge alone is not the final governance closure. This closure record must pass
its own CI/review/human-approval gates and be squash-merged into `develop`.

When this document is canonical on `develop`, Stage 3.28 is closed for P1-01 and P1-05.
