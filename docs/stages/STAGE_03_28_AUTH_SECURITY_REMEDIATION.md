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

## P1-05 — bounded Argon2 resource admission

### Root cause

Argon2id correctly used 64 MiB, `t=3`, `p=1`, but hash, verification, and dummy-verification paths
had no shared process-wide admission budget.

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
