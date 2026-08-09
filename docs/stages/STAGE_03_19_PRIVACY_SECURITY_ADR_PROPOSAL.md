# Stage 3.19 - Privacy Security and ADR Proposal

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-19-PRIVACY-SECURITY-ADR-PROPOSAL |
| Version | 0.1.1 |
| Status | Complete / merged through PR #48 at `fdf74c16446e7623f76882aa7add64554141abc6` |
| Owner | Principal Architect |
| Supersedes | None; follows merged Stage 3.18 privacy contract/security proposal |
| Dependencies | `SOURCE_OF_TRUTH.md`; Documents 42-43; ADR-005; ADR-006; proposed ADR-008; Stage 2 ER model and migration strategy; Stage 3.17-3.18 privacy proposals |
| Last Review Date | 2026-08-09 |
| Next Review Date | Historical proposal closed; successor Stage 3.20 threat-model proposal |

## Purpose

Stage 3.18 identified cryptographic erasure, deletion-marker replay, restore isolation, separation of
duties, backup evidence, and partial failure as implementation blockers. Stage 3.19 makes the next
artifact explicit: proposed ADR-008 and its security review dossier.

This stage documents a provider-neutral control model only. It does not accept ADR-008, select a
provider, create a key, add an endpoint, change a migration, execute a backup, or authorize
implementation.

## Reviewed Security Boundary

The future lifecycle has four distinct authority boundaries:

| Boundary | Required future authority | Forbidden shortcut |
| --- | --- | --- |
| Application lifecycle | Request state, auth/write blocking, idempotent coordination, non-identifying audit outcomes. | Treating database cascades as anonymization or reporting completion without external proof. |
| Key custody | Per-subject erasure-key lifecycle and irreversible destruction proof. | Exporting, recreating, or granting normal application code unilateral recovery authority. |
| Privacy control plane | Durable deletion markers that outlive recoverable backups and can be verified during restore. | Storing email, IDs, tokens, or a reversible person-to-subject map in the marker. |
| Restore operations | Isolated restore, marker replay, key-unavailability verification, evidence, and serving release. | Accepting traffic from a restored environment before every check succeeds. |

The exact identity of those systems, operators, credentials, and approval thresholds remains open for
Security Review. The proposed boundary must be evaluated against the current plain
`identity.user_investment_links` relationship; it is not retroactively true today.

## Partial-Failure Model

The future implementation cannot require a distributed transaction across PostgreSQL and a key
custody system. Its safety model must instead enforce monotonic, observable progress:

1. Persist an idempotent deletion intent/claim and block normal authentication and protected writes.
2. Create and verify a non-identifying marker durable beyond every recoverable backup.
3. Request approved destruction of the erasure key and retain only destruction proof.
4. Delete live identity/link/credential/settings/session data and mark retained financial history
   anonymous only after the field-level inventory permits it.
5. Verify every effect. Mark the request complete only then; otherwise retain a fail-closed
   `completing` state with non-identifying failure evidence and an operational escalation.

The ordering, locks, retry authority, and exact data model remain migration/implementation decisions.
No compensating action may recreate destroyed key material or re-enable normal access after a
completion claim merely to make a retry convenient.

## Restore Security Dossier

Before implementation, Security Review must approve a runbook that proves all of these outcomes:

- restore starts isolated from production traffic and credentials;
- marker-set authenticity, completeness, version compatibility, and availability are verified;
- marker replay is idempotent and removes/revokes recovered reidentification material before serving;
- destroyed per-subject keys are still unavailable after restore;
- missing, stale, conflicting, or unverifiable evidence leaves the environment non-serving;
- only separately authorized operators can release a verified environment to traffic;
- logs/evidence contain no raw identity fields, secrets, request payloads, or usable mappings; and
- encrypted backup copies are destroyed no later than 90 days with durable evidence and incident
  escalation on failure.

## Required Review Evidence

ADR-008 cannot become accepted, and no implementation proposal can start, until the review record
contains:

1. a threat model covering malicious browser/session use, application compromise, privileged database
   access, key-custody compromise, accidental restore, malicious/incorrect restore, marker outage,
   backup retention failure, and partial completion;
2. a provider-neutral key-custody design and provider-specific selection/risk record when a provider
   is proposed;
3. a field-level non-reidentification inventory for data, indexes, audit, logs, exports, queues,
   caches, replicas, and backup media;
4. deletion-marker schema/redaction/integrity/availability/retention evidence;
5. operational roles, separation-of-duties, approvals, escalation, and restore-release criteria;
6. adversarial restore rehearsal evidence that a deleted identity cannot be reconnected to retained
   financial history; and
7. explicit Security Review and Principal Architect acceptance, recorded before any API, migration,
   runtime, provider, or operations implementation proposal.

## Explicit Exclusions

This stage does not authorize:

- acceptance of ADR-008, Security Review approval, or a production-readiness assertion;
- a KMS, Vault, cloud, backup, database-role, or operational-provider decision;
- OpenAPI, Go, PostgreSQL, Web, configuration, dependency, secret, worker, scheduler, or backup
  changes;
- implementation of account deletion, cancellation, authentication factors, retention cleanup, or
  restore execution;
- physical erasure of immutable financial transactions or snapshots;
- market data, financial calculations, tax, mobile, AI, premium, email, or public API work.

## Proposal Acceptance Criteria

- Proposed ADR-008 preserves Documents 42-43 and distinguishes cryptographic erasure from a cascade
  or pseudonymization.
- The current plain mapping and missing control plane are explicitly named as gaps, not treated as
  deployed security controls.
- Key custody, deletion markers, restore isolation, partial failure, audit evidence, and backup
  expiry have concrete reviewable requirements without selecting a provider.
- The future Security Review evidence and human acceptance gate are explicit.
- No runtime, contract, schema, infrastructure, dependency, provider, or operational change appears.

## Recommended Next Step

Stage 3.19 was reviewed and squash-merged through PR #48 at
`fdf74c16446e7623f76882aa7add64554141abc6`. Its successor, Stage 3.20, prepares the required
threat-model proposal. Only a separately accepted ADR-008, Security Review, and explicit human
authorization may open the remaining OpenAPI, migration, operations, and data-inventory proposal
work.
