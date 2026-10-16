# ADR-008: Privacy-Lifecycle Erasure and Restore Controls

| Field | Value |
| --- | --- |
| Document ID | ADR-008 |
| Version | 0.4.0 |
| Status | Proposed |
| Owner | Principal Architect |
| Supersedes | None |
| Dependencies | Documents 42-43; ADR-005; ADR-006; Stage 2 ER model and migration strategy; Stage 3.17 and 3.18 privacy proposals |
| Last Review Date | 2026-08-09 (historically recorded; review event UNVERIFIED) |
| Next Review Date | Historical record: Security Review and explicit human decision before any privacy-lifecycle implementation |

Metadata provenance: [original recorded values](https://github.com/AsifAbbasov/OpenInvest/blob/d9f6c263dd3ff4d22fc978372014381880104489/docs/ADR/ADR-008-privacy-lifecycle-erasure-and-restore.md).
Owner follows Documents 42–43; recorded dates/triggers do not independently prove a review occurred
or establish a new review, approval, deadline or completed evidence collection.

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

## Revalidation candidate — 2026-10-16

**Status remains Proposed / acceptance BLOCKED.** The original decision clauses were restored from
pre-loss Git evidence; this is not acceptance or runtime implementation. The prospective
[Stage 3.24 assessment](../stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md#repository-revalidation--2026-10-16-proposal-blocked_evidence)
contains current-baseline SR-01–SR-12 findings, alternatives, lifecycle/authorization fencing,
legacy-backup barriers, independent marker completeness/release protocol and future implementation tests.
The [Stage 3.21 delta](../stages/STAGE_03_21_PRIVACY_DATA_INVENTORY_PROPOSAL.md#current-baseline-inventory-delta--2026-10-16-repository-only)
adds registration, audit deduplication, replay/valuation, JWT, import-token and Redis surfaces.
A key-erasure proof cannot sanitize old plaintext copies or establish anonymity of financial patterns;
all current blockers require independent resolution under Documents 42–43 before approval.

## Target architecture candidate v1

**Proposed, not accepted or implemented.** This amendment builds on merged PR #275 at protected
`develop@cffd2091168fa28817a49464438db01804b4ee48`, tree
`c5ada81772bb6179931636a52f88b1ae43281b91`. Documents 43/42 and accepted ADR-005/006 are unchanged.
The seven original decision clauses remain mandatory. Design status is separate from SR acceptance.

| Component / authority | Proposed responsibility and boundary |
| --- | --- |
| Identity authorization port | Resolve current principal to a scoped financial subject; check authoritative lifecycle state for every protected operation. A signed JWT alone is insufficient once deletion is pending. |
| PostgreSQL lifecycle coordinator | Own generation/version, immutable grace deadline, intent and phase checkpoints; serialize account/session/financial admission. No key bytes or provider calls inside a database transaction. |
| Independent custody plane | Non-exportable subject erasure material, immutable material versions, independent destruction authorization and terminal attestation. No replacement of destroyed handles. |
| Independent deletion journal / witnesses | Durably order admitted intents and proof transitions; authenticate completeness independently of the writer, custody operator and restore releaser. |
| Sealed restore verifier / separate release authority | Reconcile recovered rows and every admitted intent through a fenced release cut; verify key unavailability and copies before enabling traffic/jobs/credentials. |
| Privacy/audit authority | Approve field dispositions and correlation evidence; retain ten-year minimum non-identifying verification without a person lookup. |

Preferred candidate: subject-scoped encryption plus a durable intent/marker saga and authoritative
read/write fences, with delayed final completion until all plaintext/correlation obligations resolve.
Database-only cascade and shared backup encryption cannot provide this boundary. Per-subject physical
segregation simplifies copy accounting but increases operations/cost and still cannot erase exported
plaintext. Whole financial-history deletion/coarsening conflicts with current contracts and is not chosen.

Four decisions remain explicitly open: **A1**, treatment of personal text/provenance already inside
immutable financial facts; **A2**, treatment of user/session identifiers in append-only historical audit;
**A3**, actual custody/replay-binding/copy guarantees; **A4**, legal holds and approved threat scope for
external/previously exported linkage. A1/A2 require an explicit architecture decision before any
historical rewrite; A3 requires provider/operations evidence; A4 requires qualified human decisions.
No UUID, receipt, key destruction or local model permits an Anonymous Financial History claim alone.

Detailed design ownership: [field dispositions](../stages/STAGE_03_21_PRIVACY_DATA_INVENTORY_PROPOSAL.md#target-field-disposition-v1),
[key hierarchy and proof](../stages/STAGE_03_22_PRIVACY_KEY_CUSTODY_PROPOSAL.md#target-custody-protocol-v1),
[journal and restore](../stages/STAGE_03_23_PRIVACY_DELETION_MARKER_PROPOSAL.md#target-marker-and-restore-protocol-v1),
[lifecycle/migration/SR decisions](../stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md#architecture-remediation-candidate-v1),
[tests and evidence](../stages/STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md#architecture-validation-and-traceability-v1).
Acceptance remains BLOCKED; any future implementation needs separate authorization. No provider, API,
schema, key, retention exception, operational grant, merge or deployment is approved by this amendment.

## Targeted gate corrections — ARCH-277-01–04

This prospective amendment supersedes database-first admission and voluntary compatibility assumptions
in the v1 candidate; it does not change the seven original clauses or accept ADR-008. **G1** requires
independent durable PREPARE before local pending, witnessed activation with one 30-day deadline, full
epoch coverage and denied-state recovery before restore release. DB-only pending plus missing WAL and
no independent witness is irreducible uncertainty: the affected environment remains quarantined.
**G2** requires database-enforced lifecycle/role-epoch checks, revoked legacy capabilities and drained
sessions/workers, plus an independently trusted workload/credential/ingress boundary. A feature flag,
caller-set version, changed password or a binary claiming its own version cannot establish G2.
[Stage 3.23](../stages/STAGE_03_23_PRIVACY_DELETION_MARKER_PROPOSAL.md#target-marker-and-restore-protocol-v1)
owns recovery; [Stage 3.24](../stages/STAGE_03_24_PRIVACY_SECURITY_REVIEW_READINESS.md#mandatory-compatibility-enforcement-g2)
owns enforcement; [Stage 3.25](../stages/STAGE_03_25_PRIVACY_SECURITY_EVIDENCE_COLLECTION_PLAN.md#targeted-model-v2-and-remaining-integration-evidence)
owns model/evidence limits. G1/G2 and A1–A4 remain acceptance/implementation blockers.
Design documentation, accepted design, implemented control, independently verified property and
production authorization are separate states; none is inferred from the preceding one.
