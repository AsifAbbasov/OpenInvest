# ADR-008: Privacy-Lifecycle Erasure and Restore Controls

| Field | Value |
| --- | --- |
| Document ID | ADR-008 |
| Version | 0.1.0 |
| Status | Proposed |
| Owner | Principal Architect |
| Supersedes | None |
| Dependencies | Documents 42-43; ADR-005; ADR-006; Stage 2 ER model and migration strategy; Stage 3.17 and 3.18 privacy proposals |
| Last Review Date | 2026-08-09 |
| Next Review Date | Security Review and explicit human decision before any privacy-lifecycle implementation |

## Context

Documents 42-43 require complete identity deletion, irreversible destruction of the person-to-financial
link, Anonymous Financial History, 10-year audit retention, and encrypted backup destruction within
90 days. The Stage 2 migration strategy requires revocable per-subject key material and deletion-ledger
replay before a restored environment accepts traffic.

The deployed schema does not yet meet that boundary. `identity.user_investment_links` is a plain
foreign-key relationship, and no key hierarchy, deletion ledger, backup-destruction evidence, or
restore serving gate exists. A database delete, a user state transition, or encrypted backup media by
itself is pseudonymization when an older restore, an operator, or residual data can recreate the
identity link.

The deletion flow also crosses a database and a future key-custody system. There is no distributed
transaction that can make database deletion and irreversible key destruction atomic. The design must
therefore fail closed under partial failure and never report completion before both independently
verifiable effects are complete.

## Proposed Decision

This ADR proposes the following provider-neutral control model. It is not accepted, implemented, or
authorization to select a KMS, Vault, backup provider, database role model, API route, or migration.

1. A future migration must replace the plain reversible identity-to-subject protection boundary with
   a per-subject erasure boundary. All material that can reconnect an identity to retained financial
   history must be protected by non-exportable, revocable per-subject key material. A raw surviving
   foreign-key mapping is insufficient.
2. Completion must create a durable, non-identifying deletion marker before any restored environment
   can serve traffic. The marker contains only random lifecycle/erasure handles, irreversible key
   destruction evidence, lifecycle timestamps, integrity/version metadata, and non-personal failure
   codes. It contains no email, user ID, subject ID, password, session token, raw request body, or
   usable identity map.
3. The marker/control plane must survive every recoverable data backup. It must have independent
   integrity protection and availability monitoring; losing, conflicting with, or failing to verify
   it blocks restore serving rather than weakening deletion.
4. Key custody must enforce separation of duties. The normal application path may request an
   authorized lifecycle transition but cannot export, recreate, or unilaterally restore destroyed
   subject key material. The exact roles, approval threshold, credential mechanism, and provider
   controls remain Security Review decisions.
5. Completion follows one-way, fail-closed progress: durable erasure intent, protected-write and
   authentication block, key-destruction request/proof, deletion of live identity/link data,
   anonymization of the retained financial subject, and verification. A request remains
   `completing`, never `completed`, while any step is missing or unverifiable.
6. Restore begins network-isolated and non-serving. It verifies the independent marker set, applies
   markers idempotently to remove/revoke recovered reidentification material, proves destroyed keys
   remain unavailable, records non-identifying evidence, and only then permits traffic.
7. Backups must stay encrypted, expire no later than 90 days, and produce destruction evidence for
   every managed copy. The final storage topology and proof source require provider-specific review.

The future OpenAPI, PostgreSQL schema, operational runbook, and implementation tests must use this
control model only after this ADR is accepted through the required review and human approval.

## Security Properties

| Property | Proposed control |
| --- | --- |
| Irreversible unlinking | Destroy the per-subject erasure key and every live reversible mapping; do not treat a UUID or removed user row as anonymous by itself. |
| No restore reidentification | Restore replays independently protected deletion markers before serving and fails closed on missing or unverifiable evidence. |
| Least privilege | Application, key-custody, restore, and audit authorities remain separate; no normal business API may recover erased material. |
| Partial-failure safety | An incomplete deletion denies normal auth and protected writes, remains observable, and cannot be reported as completed. |
| Non-identifying evidence | Audit and marker records retain outcome/timing/integrity evidence without credentials, tokens, personal fields, or usable maps. |
| Immutable history | Financial transactions and snapshots remain; retained content must separately pass the field-level anonymization inventory. |
| Backup bound | Every encrypted copy is destroyed by the approved 90-day maximum, with durable evidence and escalation on failure. |

## Alternatives Considered

### Delete identity rows and rely on cascades

Rejected. Cascades remove live rows but do not prevent an older backup, log, audit mapping, or operator
with retained material from reconnecting a person to financial history.

### Encrypt only whole backups

Rejected. Backup encryption without subject-scoped revocation leaves an authorized restore/key holder
able to recover identity mappings. It is not cryptographic erasure.

### Keep a reversible deletion map for operational convenience

Rejected. A reversible map makes the resulting data pseudonymized and conflicts with Documents 42-43.

### Let the application service destroy and recreate keys

Rejected. It collapses key custody into the normal request path and creates a unilateral
reidentification/recovery authority.

### Mark deletion complete after one subsystem succeeds

Rejected. Distributed partial failure can leave either live reidentification material or recoverable
key material. Completion requires evidence for every irreversible effect.

## Consequences

### Positive

- Establishes a reviewable no-reidentification and restore boundary before destructive code exists.
- Separates provider choice from durable security requirements.
- Gives future migration, API, operations, and test work a shared failure model.
- Preserves immutable financial history without mislabelling pseudonymization as anonymity.

### Costs and Constraints

- Requires a future key-custody service/control plane, backup evidence, restore rehearsal, and
  least-privilege operational roles; those choices may incur provider and operating cost.
- The current plain identity link cannot remain the sole protection of retained financial history.
- Completion may take longer than a database transaction and must be observable as an asynchronous,
  fail-closed lifecycle rather than an instant delete response.
- This decision does not settle deletion-request API authentication, concrete data types, database
  locking, provider selection, ledger retention duration, or legal interpretation.

## Compatibility and Rollback

This proposed ADR changes no deployed system. It can be withdrawn by reverting this documentation
commit before acceptance. After implementation, rollback cannot recreate destroyed key material or a
deleted identity link; future implementation planning must specify only forward recovery and a
non-serving restore procedure.

## Security and Privacy Impact

The proposal strengthens the intended privacy model but creates no current data collection, key,
backup, provider, or operational behavior. Security Review must validate threat assumptions, trust
boundaries, key custody, marker integrity/availability, destruction proof, restore isolation,
incident handling, and the data inventory before acceptance.

## Approval Outcome

Proposed only. It has no authority until strict review, Security Review, and explicit Principal
Architect acceptance. Stage 3.19 does not authorize an account-deletion API, schema migration,
runtime job, provider integration, or production claim.
