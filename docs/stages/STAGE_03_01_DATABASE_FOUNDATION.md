# Stage 3.1 — Local Database Foundation

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-01 |
| Version | 0.1.1 |
| Status | Complete / Merged into `develop` |
| Owner | Builder Engineer |
| Supersedes | Stage 3 PR 3.1 placeholder |
| Dependencies | `STAGE_03_FIRST_VERTICAL_SLICE.md`; `ER_MODEL_STAGE_02.md`; `MIGRATION_STRATEGY_STAGE_02.md`; ADR-002; ADR-006 |
| Last Review Date | 2026-06-27 |
| Next Review Date | 2026-12-27 |

## Purpose

Stage 3.1 introduces the smallest PostgreSQL foundation needed for the first vertical slice:

```text
portfolio
→ immutable transaction entry
→ rebuildable portfolio snapshot
→ summary read model in later PRs
```

This PR does not implement application behavior. It only creates the local database structures,
validation guardrails, and governance updates required before the Go API slice can be implemented.

## Scope

Allowed in this stage:

- plain SQL migration files;
- `identity`, `investment`, `analytics`, and `audit` schema creation;
- minimal vertical-slice tables for identity subject mapping, portfolios, assets, immutable
  transaction entries, snapshots, outbox/inbox placeholders, idempotency records, and audit events;
- migration-pair validation in CI;
- governance documentation updates.

Forbidden in this stage:

- Go handlers, services, repositories, or database access code;
- Next.js screens or API clients;
- Python workers;
- SQL for production tax tables;
- external provider, broker import, MOEX, CBR, or Rosstat tables;
- destructive up migrations;
- seed data that could be mistaken for official market data;
- Stage 3.2 backend implementation.

## Migration tooling decision

Stage 3.1 deliberately does not introduce an external migration library.

The repository uses plain, versioned SQL migration pairs under:

```text
infrastructure/postgres/migrations/
```

Reason:

- no vendor lock-in;
- no runtime dependency;
- easy line-by-line review;
- compatible with future Go migration tooling;
- sufficient for the first local database foundation.

A future migration runner may be selected only after review. If the selected tool changes
architecture, deployment, rollback, or operational semantics, it requires a dedicated ADR.

## Migration files

| File | Purpose |
| --- | --- |
| `000001_stage_03_01_vertical_slice.up.sql` | Additive local schema/table creation for the first vertical slice |
| `000001_stage_03_01_vertical_slice.down.sql` | Local disposable rollback of structures created by the paired up migration |

The down migration contains `DROP` statements only for local/disposable rollback of structures
created by the paired up migration. Production destructive Contract migrations remain forbidden
without a staged ADR and explicit approval.

## Docker Compose adjustment

Stage 3.1 updates the PostgreSQL volume mount to:

```text
postgres-data:/var/lib/postgresql
```

Reason:

- `postgres:18-alpine` stores data in a major-version-specific directory under
  `/var/lib/postgresql`;
- mounting the volume at `/var/lib/postgresql/data` makes PostgreSQL 18 reject the existing mount
  layout before the database starts;
- the repository must support local migration verification with the frozen PostgreSQL 18 image.

## Data model notes

- Financial values use `NUMERIC(28, 8)`.
- Financial business dates use `DATE`.
- System timestamps use `TIMESTAMPTZ`; application code must treat them as UTC.
- Transaction entries are append-only by design; application grants that prevent UPDATE/DELETE are
  deferred until runtime roles are introduced.
- Snapshots are rebuildable projections and never become the source of truth.
- The identity-to-investment mapping remains isolated in the `identity` schema.
- Investment records contain no email, passport, INN, phone, address, or tax profile data.

## Validation

Stage 3.1 adds:

```text
scripts/validate_migrations.rb
```

The validator checks:

- every `.up.sql` migration has a paired `.down.sql`;
- up migrations do not contain destructive operations such as `DROP`, `TRUNCATE`, `DELETE FROM`,
  `UPDATE`, or `ALTER TABLE ... DROP`;
- the Stage 3.1 migration creates the required schemas and tables;
- binary floating-point SQL types are absent;
- `NUMERIC(28, 8)` is present for decimal financial fields.

Live disposable PostgreSQL verification was also performed with Docker Compose project
`openinvest_stage31` on local port `55433`:

- `000001_stage_03_01_vertical_slice.up.sql` applied successfully;
- created schemas/tables:
  - `identity`: 2 tables;
  - `investment`: 6 tables;
  - `analytics`: 4 tables;
  - `audit`: 2 tables;
- `000001_stage_03_01_vertical_slice.down.sql` applied successfully;
- rollback check returned `0` remaining tables in the four Stage 3.1 schemas;
- disposable container and volume were removed after verification.

## Known limitations

- No migration runner is selected yet.
- Runtime least-privilege database roles are not introduced yet.
- Ledger UPDATE/DELETE prevention is documented but not enforced with grants in this PR.
- No application code applies migrations yet.
- No production migration rehearsal exists yet; Stage 3.1 verification is local/disposable only.

These limitations are intentional for the first database foundation PR and must be addressed before
production deployment.

## Acceptance criteria

- Migration files are present and paired.
- Migration validator passes locally and in CI.
- Existing Go, Python, frontend, OpenAPI, and Docker Compose checks remain green.
- No runtime implementation is introduced.
- Internal Review Agent approves.
- Human approval is required before merge into `develop`.
