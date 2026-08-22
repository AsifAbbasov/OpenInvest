# Stage 3.28 — Authentication Security Remediation

| Field | Value |
| --- | --- |
| Status | Open / implementation candidate / Draft PR #59 |
| Owner | Principal Architect |
| Baseline | `develop` at `6f50c39ee19834f1b2ff354230b727019be64369` |
| Branch | `fix/stage-03-28-auth-security-remediation` |
| Trigger | Repository-audit findings P1-01 and P1-05 |
| Scope | Refresh-token replay/session-family containment; refresh/logout concurrency; Argon2 resource admission; PostgreSQL defense in depth; regression coverage |
| Out of scope | Stage 3.25 privacy work; P2/P3 findings; product expansion; frontend features; broker/provider work; unrelated architecture changes |

## Purpose

Stage 3.28 addresses the two remaining P1 findings from the repository audit without mixing them with Stage 3.25 privacy planning or the P2/P3 backlog.

P1-01 concerns refresh-token replay: before this remediation, a token that had already been rotated was filtered out by the active-session lookup, so replay was rejected but the newly issued descendant could remain valid.

P1-05 concerns Argon2 availability: password hashing used the approved Argon2id cost but expensive derivations had no process-wide admission bound, so distributed concurrent authentication traffic could create avoidable memory/CPU bursts despite per-client request limiting.

## P1-01 — refresh-token replay and session-family containment

### Root cause

Session rotation persisted independent session rows but no family lineage. The refresh store selected only active sessions, so a replayed revoked token was indistinguishable from an unknown token and could not identify descendants that should be invalidated.

### Remediation

Migration `000005_stage_03_28_auth_security` adds nullable `session_family_id` and family/state/expiry lookup support without modifying historical migrations.

Historical rows intentionally keep `session_family_id = NULL`: Stage 3.28 does not fabricate lineage that was never persisted. New registration/login sessions use their own session id as the family root, and rotated descendants inherit that family.

A PostgreSQL insert guard rejects newly created session rows without family identity while still allowing pre-Stage-3.28 legacy rows to remain nullable.

Refresh and logout mutation paths resolve the owning user and acquire one transaction-scoped advisory lock for that user before row locking. This establishes one ordering boundary for related token mutations.

For post-Stage-3.28 replay, a known revoked token revokes every active member of its family before returning `ErrInvalidSession`. For a legacy replay whose family is unknowable, the store fails closed by revoking all active sessions for that user.

Logout uses the same serialization boundary. A known stale/revoked token can still contain its family, preventing refresh-versus-logout races from leaving a newly rotated descendant alive.

## P1-05 — bounded Argon2 resource admission

### Root cause

Argon2id correctly used 64 MiB, `t=3`, `p=1`, but hash, verification, and dummy-verification paths invoked the derivation without a process-wide capacity budget.

### Remediation

The approved Argon2id cost is unchanged.

All hash, verify, and dummy paths use one process-wide admission gate capped at two simultaneous Argon2 derivations. When both slots are occupied, additional expensive work fails immediately with `ErrAuthCapacity`; it does not wait in an unbounded request/goroutine queue.

Stored encoded password hashes are accepted for expensive verification only when memory, time, thread, salt-length, and hash-length parameters exactly match the approved budget. Over-budget or malformed encodings are rejected before Argon2 runs.

`ErrAuthCapacity` is an explicit service-layer capacity signal. At the current candidate head it still reaches the generic HTTP error mapper; a dedicated public `503 + Retry-After` mapping is desirable follow-up hardening but is not relied upon for the memory-admission guarantee itself.

## Regression matrix

| Case | Expected result |
| --- | --- |
| Normal refresh rotation | New descendant inherits the root family |
| Replay old rotated token | Replay rejected and active family descendants revoked |
| Two concurrent refreshes of same token | At most one rotation succeeds; replay containment revokes the winner descendant |
| Old-token replay racing descendant refresh | No active member remains in compromised family |
| Independent post-Stage-3.28 login family | Remains independent from replay in another known family |
| Legacy token with unknown family replayed | Fail closed across active sessions for that user |
| Logout racing refresh | Logout containment leaves no active descendant |
| Direct SQL new session without family id | PostgreSQL rejects the insert |
| Migration apply / rollback / reapply | Completes cleanly without modifying historical migrations |
| Two Argon2 derivations already running | Additional expensive work fails immediately with `ErrAuthCapacity` |
| Capacity slots later released | Authentication capacity recovers |
| Stored over-budget Argon2 parameters | Rejected before expensive derivation |
| Unknown-user login dummy path at capacity | Returns capacity exhaustion rather than bypassing the admission gate |

## Verification evidence

Technical implementation head `78dbab4de275277897d56bb317b20633fa92d940` passed GitHub Actions CI #110 before this report was added. That run covered PostgreSQL migration validation, migration apply, every rollback and full reapply, Go tests with PostgreSQL integration tests, OpenAPI validation, frontend build/typecheck/tests, Python tests, and Docker Compose validation.

Because this report advances the PR head, CI #110 is historical head-specific evidence only. The authoritative pre-merge CI gate is the required PR checks attached to the exact final PR #59 head.

## Scope and residual work

Stage 3.28 addresses P1-01 and P1-05 only.

Stage 3.25 remains a separate documentation-only privacy evidence-collection plan. The audit P2/P3 backlog remains separate. No market-data, tax, mobile, AI, broker integration, provider selection, or unrelated product work is authorized here.

## Closure rule

Stage 3.28 must not be described as merged, canonical, complete, or closed while PR #59 is open.

Closure requires green required CI on the exact merge-candidate head, independent final review, explicit human squash-merge approval, squash merge into `develop`, and the repository's closure-governance workflow.
