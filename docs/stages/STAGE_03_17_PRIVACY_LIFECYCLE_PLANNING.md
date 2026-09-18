# Stage 3.17 - Privacy Lifecycle Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-17-PRIVACY-LIFECYCLE-PLANNING |
| Version | 0.1.1 |
| Status | Complete / merged into `develop` through PR #46 at `1e8c240` |
| Supersedes | Informal disposition of the Stage 3.16 privacy-lifecycle audit blocker |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-005; Documents 42-43; Stage 2 API contract and ER model; Stage 3.11 auth/privacy slice; Stage 3.16 audit report and audit fixes |

## Purpose

Stage 3.16 left account deletion, irreversible anonymization, backup destruction, and retention
execution as non-production blockers. Stage 3.17 defines the narrow planning boundary needed to
remove that blocker without quietly adding an account API, a destructive migration, or a backup
provider integration.

The future user outcome is limited to an account deletion lifecycle that, after a reviewed grace
period, completely deletes identity data, irreversibly destroys the link to financial history, and
leaves only Anonymous Financial History. Restoring an older backup must not recreate that link.

## Current Evidence and Gap

The current schema already provides useful lifecycle primitives:

- `identity.users` has `active`, `pending_deletion`, and `deleted` states plus request/completion
  timestamps;
- `investment.subjects` has `active` and `anonymous` states plus an anonymization timestamp;
- `identity.user_investment_links` is the only persisted user-to-subject link and cascades away if
  an identity is deleted;
- credentials, privacy settings, and sessions cascade from an identity deletion.

Those primitives do not complete the approved lifecycle. There is no deletion-request record, no
account-deletion operation in the frozen OpenAPI contract, no approved grace-period behavior, no
key-destruction or backup-replay mechanism, and no operational evidence that retention is executed.

## Planning Scope

This stage defines the gates and acceptance criteria for a later implementation proposal:

- an explicit account-deletion request, cancellation, grace-period expiry, and completion state
  machine;
- the data inventory for identity records, identity-to-subject links, sessions, credentials,
  privacy settings, user-entered portfolio metadata, audit evidence, backups, and restored data;
- a completion transaction that prevents new authenticated work, revokes active sessions, removes
  identity records and the reversible link, and marks the retained financial subject anonymous;
- a no-reidentification proof covering live data, operational logs, audit evidence, backups, and
  restoration;
- retention and destruction evidence for encrypted backups, including the maximum 90-day expiry
  bound already defined by the canonical database migration strategy;

## Required Decisions Before Implementation

appropriate contract, security, or operations proposal:


Any choice that changes the approved privacy model, data ownership, external-provider policy, or
implementation begins.

## Future Implementation Sequence

The later implementation proposal must keep these phases separate and reversible where possible:


The implementation must not physically erase immutable transaction or snapshot history merely to
delete an account. It must instead meet Document 43's Anonymous Financial History requirement.

## Explicit Exclusions

This planning stage does not authorize:

- Go handlers, services, stores, background jobs, schedulers, or workers;
- OpenAPI, schema, example, or generated-client changes;
- SQL migrations, destructive data changes, backup configuration, KMS/Vault integration, or cloud
  provider selection;
- deletion of financial ledger or snapshot records;
- privacy export, password reset, email verification, OAuth, passkeys, 2FA, device management, or
  profile editing;
- market data, financial calculations, tax, mobile, premium, public API, or provider work;
- legal-compliance claims, a production-readiness declaration, or a shortened retention period.

## Security and Privacy Invariants

- The identity-to-subject mapping is the critical reidentification link and must be irreversibly
  removed at completion.
- A deleted identity cannot authenticate, refresh a session, invoke protected APIs, or be restored
  by replaying an old refresh token.
- Password hashes, refresh-token hashes, CSRF-token hashes, raw request payloads, and unredacted
  personal data never enter deletion audit events or operational logs.
- The future design must use least-privilege operational authority; no single routine application
  path may recreate destroyed key material or bypass deletion-ledger replay after restore.
- The deletion process must be observable with non-identifying outcome, timestamp, request, and
  failure-code evidence.

## Future Verification Matrix

The future implementation cannot be accepted until it proves at least:

- request, cancellation, expiry, completion, duplicate retry, unauthorized, and already-deleted
  paths;
- authentication and all active sessions are rejected or revoked at the correct lifecycle boundary;
- concurrent portfolio/import writes cannot commit after the account becomes pending deletion;
- live identity rows, credentials, settings, sessions, and user-to-subject links are absent after
  completion while financial history remains queryable only as anonymous history;
- audit records retain no prohibited secrets or reversible identity links;
- a representative encrypted-backup restore cannot reidentify a deleted subject, and deletion-ledger
  replay completes before traffic is served;
- backup expiry/destruction, restore rehearsal, migration rollback, and operational escalation have
  durable evidence;
  before merge.

## Planning Acceptance Criteria

- The Stage 3.16 privacy-lifecycle blocker is restated with its precise implementation gap.
- The planned data inventory, state machine, contract gate, migration gate, and backup/key-destruction
  gate are explicit.
- Anonymous Financial History is preserved without claiming that the current implementation already
  provides it.
- Canonical governance registers identify this as planning-only while active and preserve its merged
  closure without granting implementation authorization.
- No runtime code, OpenAPI, migration, dependency, or operational configuration changes are included.

## Recommended Next Step

Canonical record: PR #46.
`1e8c240`. Its successor, Stage 3.18, remains a separate contract/security proposal; only after
that proposal's required approvals and explicit merge gate may a privacy-lifecycle
implementation stage be proposed.
