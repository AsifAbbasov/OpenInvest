# Stage 3.18 - Privacy Contract and Security Proposal

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-18-PRIVACY-CONTRACT-SECURITY-PROPOSAL |
| Version | 0.1.1 |
| Status | Complete / merged into `develop` through PR #47 at `4680e9c1b7b916169972c84ad8c3879955c7f509` |
| Owner | Principal Architect |
| Supersedes | None; follows the Stage 3.17 privacy-lifecycle planning gate |
| Dependencies | `SOURCE_OF_TRUTH.md`; Documents 42-43; ADR-005; ADR-006; Stage 2 API contract, ER model, and migration strategy; Stage 3.11 auth/privacy slice; Stage 3.17 privacy-lifecycle planning |
| Last Review Date | 2026-08-09 |
| Next Review Date | Historical proposal closed; successor Stage 3.19 |

## Purpose and Authority

Stage 3.17 is merged and established that account deletion, irreversible anonymization, backup
destruction, and retention execution cannot be implemented from the present schema alone. This
document turns that gate into a reviewable **candidate** contract and security design.

It is not an API amendment, a security approval, a migration design approval, or an implementation
authorization. The frozen `openapi/` contract, current PostgreSQL schema, runtime behavior, and
operational configuration remain unchanged. A later implementation may be proposed only after this
candidate receives separate contract and Security Review approval, required ADR decisions, and
explicit human authorization.

## Current Evidence and Non-Compliance Gap

The currently deployed schema has lifecycle-shaped fields, not a completed privacy lifecycle:

- `identity.users` permits `active`, `pending_deletion`, and `deleted` and has deletion timestamps;
- `investment.subjects` permits `active` and `anonymous` and has `anonymous_at`;
- `identity.user_investment_links` is the sole persisted identity-to-subject mapping;
- credentials, privacy settings, and sessions cascade when a user row is deleted;
- the logical ER model names `identity.deletion_requests`, but migrations do not create it;
- the frozen OpenAPI has no account-deletion operation; and
- no current key hierarchy, deletion ledger, backup-destruction evidence, or restore replay exists.

The mapping is currently a plain database relationship. The migration strategy's per-subject key
material and deletion-ledger requirements are future design requirements, not properties of the
deployed schema. Therefore neither the current cascade nor an account-state value can be described
as irreversible anonymization or as compliance with Documents 42-43.

## Candidate Public Contract

The following is a contract proposal for later OpenAPI review only. Route names, DTOs, error codes,
authentication details, and examples are not added to `openapi/` by this stage.

| Candidate operation | Candidate semantics | Required security boundary |
| --- | --- | --- |
| `POST /api/v1/account/deletion-requests` | Create or replay the caller's pending deletion request and return its opaque request ID, state, and grace deadline. | Authenticated active identity, CSRF protection, current-credential confirmation, and `Idempotency-Key`. The confirmation secret is verified ephemerally and is never persisted, logged, traced, or audited. |
| `GET /api/v1/account/deletion-requests/current` | Read only the caller's current non-final request state and grace deadline. | Authenticated requester or another approved cancellation ceremony; no user lookup endpoint or cross-user identifier access. |
| `POST /api/v1/account/deletion-requests/{requestId}:cancel` | Cancel only a still-pending request before its grace deadline and return the restored active state. | A dedicated, rate-limited cancellation ceremony must be approved by Security Review. A revoked session, an opaque request ID alone, or a retained raw secret is insufficient authority. |

The request body must contain no profile fields, free text, reason, contact details, financial data,
or device inventory. The only candidate acknowledgement is a fixed confirmation value plus the
ephemeral credential factor. `requestId` is opaque, unguessable, and visible only while the request
is pending; it is not an identity surrogate after completion.

The final OpenAPI proposal must explicitly define `400`, `401`, `403`, `404`, `409`, `429`, and
idempotent replay semantics. It must also define whether a pending identity may read its request
status and the approved cancellation ceremony. This document deliberately does not settle that
security-sensitive choice.

## Candidate Lifecycle and Consistency Rules

The canonical ER model defines a 30-day grace period. The candidate lifecycle is:

```text
active
  -> pending_deletion (validated request; one 30-day grace deadline)
  -> active           (approved cancellation before deadline)
  -> completing       (deadline claim; all normal authentication and protected writes denied)
  -> completed        (identity erased, link/key destroyed, subject anonymous)
```

`completing` and `completed` are proposal-state names, not values added to the current database
enum-like checks. A later migration must choose exact persisted states and constraints. The existing
`deleted` user value is not evidence of a completed lifecycle because a completed identity row must
not remain as a reidentification record.

Only one pending request may exist for an identity. A duplicate create with the same idempotency key
must replay the original state; a competing key must not create a second grace window. Cancellation
and completion must serialize on the same identity/request boundary. Once the deadline claim
succeeds, cancellation loses deterministically and normal sessions, refreshes, logins, and protected
mutations must fail closed.

The required cancellation ceremony is intentionally unresolved. Stage 3.17 requires active sessions
to be revoked and normal protected mutations to stop at the lifecycle boundary, so a later design
cannot casually reuse an active session as cancellation authority. Security Review must approve a
fresh-factor, rate-limit, disclosure, and recovery design before a cancellation endpoint is added.

## Data Inventory and Completion Boundary

The implementation proposal must attach an owner and disposition for every field, projection, log,
export, and backup below. A UUID, request ID, actor ID, source label, or free-text field is not
automatically anonymous merely because it is not an email column.

| Data area | Required completion disposition | Current design status |
| --- | --- | --- |
| `identity.users`, credentials, privacy settings, sessions | Revoke active sessions, then delete identity and credential/settings/session data; retain no raw password, refresh, or CSRF token material. | Cascades exist; lifecycle transaction and verification do not. |
| `identity.user_investment_links` | Delete the sole reversible mapping and prove it cannot return through a restored backup. | Delete cascade exists; restore prevention does not. |
| `identity.deletion_requests` and retry records | Keep requester linkage only while pending; completion must erase or irreversibly sever it. A non-identifying replay marker must not contain a secret or direct identity key. | Logical model only; no deployed table. |
| portfolios, transaction entries, snapshots, import labels, notes, and derived projections | Preserve immutable financial history only after an inventory proves retained fields are non-personal or are removed/irreversibly anonymized. | No field-by-field anonymization proof. |
| audit actors, audit events, request/trace data, operational logs, and exports | Retain useful 10-year audit evidence without a reversible identity mapping or secrets. An actor UUID by itself is not a sufficient proof of anonymity. | Retention requirement exists; final disposition is unimplemented. |
| database backups, replicas, recovery media, and restored environments | Keep encrypted backups for at most 90 days; destroy expired copies; block serving until deletion replay and key-erasure checks succeed. | No implemented retention, destruction, or restore gate. |

## Candidate Cryptographic-Erasure and Restore Design

The following are non-negotiable design properties, not a selection of KMS, Vault, cloud provider,
or key algorithm:

1. A per-subject revocable key boundary must protect material that could recreate the deleted
   identity-to-subject connection. The current plain foreign-key relationship is insufficient.
2. Completion must produce a durable, non-identifying deletion/replay marker that outlives every
   recoverable backup. It cannot contain email, credentials, raw tokens, or a usable identity map.
3. Key destruction must have durable proof and independent authorization. An ordinary application
   request path must not be able to recreate destroyed material or bypass the restore gate.
4. A restore begins in a non-serving state. Before any traffic is accepted, the restore procedure
   must load the applicable deletion markers, replay them idempotently, verify erased keys cannot be
   recovered, and retain an evidence record without personal payloads.
5. Missing, stale, conflicting, or unverifiable deletion-marker/key evidence is a fail-closed
   incident: the restored environment remains non-serving and escalates to the approved operations
   path.
6. Backup expiry/destruction must be evidenced for all managed copies and tested against the
   maximum 90-day retention bound. No claim is made here about a currently configured backup system.

The later Security Review must resolve key custody, separation of duties, deletion-proof format,
ledger availability and retention, recovery from partial failure, destructive-operation approval,
and provider-specific restore controls. It must also address the unavoidable cross-system failure
window between database erasure and external key destruction. A request cannot be marked completed
until both sides have verified success; while unresolved it must remain fail-closed and observable.

## Threat Model and Security Requirements

The future design must be reviewed against at least these adversarial or failure paths:

| Scenario | Required result |
| --- | --- |
| Stolen browser session or CSRF bypass attempt | Cannot create or cancel deletion without the approved fresh factor and anti-forgery checks. |
| Request replay, concurrent cancellation, or completion race | One serialized result, idempotent replay, no second grace window, and no restored write access after a completion claim. |
| Application database read, support-tool access, or audit query | Cannot reconnect anonymous financial history to a deleted person through residual mappings, secrets, logs, or actor records. |
| Older encrypted backup or disaster restore | Cannot serve traffic until deletion replay runs and destroyed material remains unavailable. |
| Key service or deletion-ledger outage | Completion is not falsely reported; access and normal writes remain fail-closed while operators use the approved recovery path. |
| Backup expiry or copy-destruction failure | Generates durable non-identifying evidence and an operational escalation; it is never silently treated as compliant. |

No component may log a current password, a raw refresh/CSRF token, a recovery secret, a full request
body, or unredacted identity data. Rate limits and generic public errors must prevent the candidate
endpoints from becoming an account-existence or request-status oracle.

## Required Future Artifacts and Gates

No implementation PR may begin until all of the following are individually reviewed:

1. An OpenAPI amendment that replaces the candidate contract with approved paths, schemas, examples,
   errors, CSRF/authentication behavior, idempotency, and API compatibility decisions.
2. An ADR or Security Review record that approves the cryptographic-erasure, key-custody,
   deletion-marker, backup-retention, restore, separation-of-duties, and partial-failure designs.
3. A PostgreSQL migration proposal defining the deletion-request data model, constraints, access
   controls, retention, cleanup, locking/index evidence, and rollback/forward-recovery strategy.
4. An operations runbook covering deletion execution, proof collection, backup expiry, restore
   replay, non-serving recovery, incident escalation, and least-privilege roles.
5. A complete field-level anonymization inventory for portfolio metadata, ledger/supporting data,
   imports, projections, audit/log payloads, exports, queues, and backups.
6. Explicit human approval of the implementation-stage scope after the preceding artifacts have
   received their own review verdicts.

Any provider, key-management, data-ownership, retention-period, cost, or privacy-model decision
that changes existing canonical policy requires Issue -> ADR -> Review -> Approval before it is
treated as approved.

## Future Verification Matrix

An implementation is unacceptable until it proves, with automated tests and durable operational
evidence, at least:

- authenticated request, unauthenticated request, CSRF failure, failed fresh factor, rate limit,
  same-key replay, competing-key race, current-status, cancellation, expiry, completion, and
  already-completed paths;
- session revocation, refresh rejection, login/protected-write denial, and a deterministic
  cancellation authority boundary during `pending_deletion` and `completing`;
- concurrency protection for portfolio, transaction, import, and idempotent command writes across
  deletion state transitions;
- deletion of all live identity/link/credential/settings/session data while financial history
  remains only as Anonymous Financial History;
- no personal payload or reversible identity link in audit events, logs, idempotency data, exports,
  queues, or operational evidence;
- key-destruction proof, deletion-marker replay, blocked traffic on missing evidence, and a
  representative backup restore that cannot reidentify a completed deletion;
- encrypted-backup expiry/destruction proof within 90 days, partial-failure recovery, migration
  rollback/forward recovery, and independent Security Review approval; and
- green OpenAPI, Go, PostgreSQL, Web, operational, and CI checks after each final contract or
  implementation change.

## Explicit Exclusions

This proposal does not authorize or add:

- Go handlers, services, stores, jobs, schedulers, workers, Web views, or user-facing flows;
- OpenAPI paths, schemas, examples, generated clients, or actual endpoint behavior;
- SQL migrations, data changes, database roles, RLS, backup configuration, KMS/Vault integration,
  cloud-provider selection, or external communications;
- physical deletion of immutable transactions or snapshots;
- email verification, password reset, OAuth, passkeys, 2FA, export, profile editing, or device
  management;
- market data, financial calculations, tax, mobile, AI, premium, public API, or provider work; or
- a legal-compliance, production-readiness, security-approval, or retention-execution claim.

## Proposal Acceptance Criteria

- The proposal distinguishes deployed facts from candidate design and names every gap that blocks
  implementation.
- Candidate API semantics, 30-day lifecycle, cancellation conflict, idempotency, data inventory,
  cryptographic-erasure, backup, restore, and failure boundaries are reviewable without modifying
  the frozen OpenAPI or runtime.
- Documents 42-43's complete identity deletion, no-reidentification, 10-year audit retention,
  immutable Anonymous Financial History, and maximum 90-day backup retention are preserved.
- Required security, contract, migration, operations, ADR, evidence, and human gates are explicit;
  canonical governance records preserve this as a completed proposal without implementation authority.
- No runtime, API, database, dependency, provider, or operational configuration change is included.

## Recommended Next Step

This proposal received strict review and was squash-merged through PR #47 at
`4680e9c1b7b916169972c84ad8c3879955c7f509`. Its successor, Stage 3.19, prepares proposed ADR-008;
only accepted ADR/security decisions and explicit human authorization may open a separately scoped
privacy-lifecycle implementation proposal.
