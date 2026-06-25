# Stage 2 Migration Strategy

| Field | Value |
| --- | --- |
| Document ID | DB-MIGRATION-STAGE-02 |
| Version | 1.0.0 |
| Status | Proposed Strategy / No Migrations Authorized |
| Owner | Principal Architect |
| Supersedes | Ad hoc migration guidance in legacy documents |
| Dependencies | Documents 42–43; ADR-002; ADR-006 |
| Last Review Date | 2026-06-20 |
| Next Review Date | Before Stage 4 migrations |

## Purpose

Freeze how PostgreSQL schema and data changes will be delivered before any SQL table-creation
migration exists. The mandatory lifecycle is:

```text
Expand → Populate → Switch → Validate → Contract
```

Correctness, security, privacy, rollback, and availability take priority over delivery speed.

## Non-negotiable rules

1. No destructive one-step migration.
2. No `DROP`, destructive type conversion, column rename, or semantic reuse without a staged ADR.
3. Application versions before and after deployment must coexist during the transition window.
4. Financial ledger rows are never updated/deleted by a migration to "clean up" history.
5. Snapshots may be rebuilt; transactions may not be rewritten.
6. Every migration is versioned, immutable after merge, reviewed, observable, and rehearsed.
7. Production schema changes run through CI/CD with least-privilege credentials; never manually.
8. Backups and point-in-time recovery are verified before high-risk changes, but backup existence
   is not a substitute for a rollback design.
9. No production-code TODO may defer a known migration risk; use Issue → ADR → approval.

## Phase 1 — Expand

Introduce only backward-compatible structures:

- new nullable column or column with a safe server-side default;
- new table/schema/index that old application versions ignore;
- new enum/check behavior only after old readers tolerate it;
- new read model or event version alongside the old version;
- additive API field only when clients tolerate unknown fields.

Large-table indexes use PostgreSQL online/concurrent mechanisms where supported and are kept out
of transactions when PostgreSQL requires it. Lock timeout and statement timeout are explicit.
Every DDL statement has an estimated lock mode, affected row count, disk impact, replication/WAL
impact, and abort condition.

## Phase 2 — Populate

Backfill new representations without changing the active read path:

- resumable, idempotent batches ordered by stable primary key;
- bounded transactions and rate limits;
- progress/watermark persisted separately from business data;
- checksums/counts and domain invariants measured continuously;
- pause/resume without duplicate business effects;
- no binary float conversion for financial values;
- no date conversion through local timezones.

Populate jobs are operational migrations, not permanent business workers. They are removed after
validation and Contract unless an approved ongoing responsibility remains.

## Phase 3 — Switch

Move traffic using a separately deployable application change:

1. deploy code capable of reading both representations;
2. enable shadow reads/comparison where privacy permits;
3. use narrowly scoped feature/config control to switch reads;
4. retain the old write/read path for rollback during the observation window;
5. avoid indefinite dual writes; if unavoidable, define ordering, failure recovery, and source of truth.

The canonical source remains explicit throughout. Cache invalidation is part of the switch plan.
OpenAPI-breaking switches require a versioned contract and cannot hide inside a DB migration.

## Phase 4 — Validate

Validation is domain-aware, not merely row counts:

- referential and uniqueness invariants;
- decimal scale and half-even expected results;
- transaction revision continuity and reversal integrity;
- BusinessDate equality without UTC drift;
- snapshot rebuild equality against canonical ledger and methodology version;
- identity/investment separation and absence of personal data in financial schemas;
- outbox/inbox deduplication and business-version monotonicity;
- query plans and SLO evidence on representative volume;
- backup restore plus migration replay in a non-production environment.

A signed validation report identifies dataset/watermark, queries/tool versions, mismatches, accepted
risk, and reviewer. Any unexplained financial mismatch blocks Contract.

## Phase 5 — Contract

Remove the obsolete representation only after:

- all production traffic uses the new path;
- rollback/observation window has elapsed;
- no supported application version depends on it;
- validation is green;
- retention/legal/privacy requirements allow removal;
- a staged ADR explicitly authorizes destructive removal;
- a fresh backup/restore rehearsal proves the final path.

Contract is its own PR and deployment. It is never bundled with Expand or Switch. `DROP` is not
the default end state; retaining a deprecated structure temporarily is safer than premature loss.

## Migration classification

| Risk | Examples | Minimum gate |
| --- | --- | --- |
| Low | new empty table, additive nullable column | review, CI, rollback statement |
| Medium | backfill, new constraint/index, read-path switch | rehearsal, metrics, staged rollout |
| High | financial representation, identity link, encryption, event ordering | ADR, security/privacy review, golden vectors, restore rehearsal |
| Destructive | DROP, irreversible conversion, history rewrite | separate staged ADR; normally forbidden |

## Versioning and ordering

- Migration IDs are monotonic and immutable.
- Each migration declares schema owner, phase, dependency, reversibility, expected duration,
  lock risk, data classification, monitoring, and rollback/roll-forward procedure.
- Separate context migrations may deploy independently but cross-schema dependencies are explicit.
- Production runtime verifies that the database is at a compatible migration range, not merely
  "latest".
- Failed migrations stop the pipeline and require diagnosis; they are not marked successful manually.

## Rollback model

Rollback prefers application/config rollback while expanded structures remain additive. A schema
down migration is used only when it is demonstrably safe. Data populated into a new structure may
remain unused after rollback. Financial facts are never deleted to simulate rollback.

If a migration has created external side effects or emitted events, rollback is compensating and
idempotent. Transport remains at least once; consumers protect business effects with inbox keys and
business versions.

## Snapshot migrations

Snapshot schema/methodology change creates a new version:

1. expand new snapshot representation/version;
2. rebuild from immutable transactions and registered market/inflation inputs;
3. compare against golden vectors and prior version with explained deltas;
4. switch reads by methodology/version;
5. retain prior version through rollback window;
6. contract only when retention and audit requirements permit.

Never mutate historical transaction rows to make a snapshot match.

## Identity deletion and backups

Identity deletion must remain irreversible even after restoring an older encrypted backup.
Sensitive identity/link material is protected with revocable per-subject encryption-key material.
At deletion completion the live rows and reversible mapping are deleted and the corresponding key
material is cryptographically destroyed. A restored backup cannot recreate the identity link
without the destroyed key. Restore procedures must also replay the deletion ledger before serving
traffic. Backups remain encrypted and expire within 90 days.

The exact key hierarchy, Vault policy, deletion ledger, and restore runbook require Security Review
before implementation. A design that permits an operator to reconstruct a deleted identity link
is pseudonymization and fails the approved anonymization requirement.

## Observability

Every production migration reports:

- version, phase, owner, start/end/status;
- rows/batches processed without logging row content;
- lock wait, statement duration, replication lag, WAL/disk growth;
- validation mismatch counts;
- retry/pause/abort reason;
- request/change ticket and deployment correlation.

Metrics and logs contain no passwords, tokens, passport, INN, raw XML/PDF, or financial document
content.

## Tooling policy

Stage 2 does not choose a migration library. Selecting one is a Stage 4 dependency decision based
on Go compatibility, checksum/locking behavior, transactional support, observability, maintenance,
license, and rollback workflow. A library choice that affects architecture or creates lock-in needs
an ADR; no tool may weaken this strategy.

## Stage 4 entry criteria

Before the first SQL migration:

- ER model and ADR-006 are approved;
- exact PostgreSQL types/constraints and role grants are reviewed;
- migration tool is approved;
- local/CI disposable database tests exist;
- upgrade and rollback rehearsals exist;
- anonymization/key-destruction threat model is approved;
- no unresolved canonical-model question remains.
