# Stage 3 — First Vertical Slice Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03 |
| Version | 0.1.11 |
| Status | Active / staged implementation |
| Owner | Builder Engineer |
| Supersedes | Roadmap placeholder for the first vertical slice |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-003; ADR-006; ADR-007; Stage 2 contract baseline; Web presentation baseline |
| Last Review Date | 2026-07-02 |
| Next Review Date | Before Stage 3.9 planning approval |

## Purpose

Stage 3 turns the approved contract and Web baseline into the first thin working product path.

The goal is not to build the full MVP. The goal is to prove that OpenInvest can move one user-owned
investment action through the complete architecture:

```text
Next.js presentation shell
→ Go API
→ PostgreSQL
→ snapshot rebuild
→ API response
→ rendered dashboard/portfolio state
→ tests
```

## Stage 3 outcome

After Stage 3 implementation is complete, a developer should be able to run the project locally and
demonstrate one controlled flow:

```text
Create a portfolio
→ add one transaction
→ persist it immutably
→ rebuild a portfolio snapshot
→ read portfolio summary through the Go API
→ render the resulting state in the Next.js Web shell
```

## Non-goals

Stage 3 must stay deliberately small.

Forbidden in Stage 3:

- AI assistant;
- tax export;
- email automation;
- mobile implementation;
- direct broker API synchronization;
- MOEX, CBR, Rosstat, or external provider integration;
- dividend calendar production ingestion;
- forecast engine;
- premium analytics;
- public API;
- multi-currency support;
- foreign securities;
- full authentication implementation beyond minimal local/dev guard if explicitly approved;
- production deployment;
- styling/design-system expansion beyond basic readable screens.

Direct broker API synchronization and credential scraping remain forbidden. User-supplied
broker-file import is now a controlled public-MVP readiness path after Stage 3.5 design approval.

## Required implementation slices

Stage 3 should be split into small PRs. Each PR must remain reviewable and must not exceed the PR
size budget unless explicitly approved.

### PR 3.1 — Local database foundation

Purpose:

- introduce the minimal PostgreSQL schema needed for the vertical slice;
- keep schemas aligned with Stage 2 ER model;
- provide reversible migrations;
- provide test database setup.

Allowed:

- migration tooling if documented and justified;
- `identity`, `investment`, `analytics`, and `audit` schema creation required by the slice;
- minimal tables for users/dev subject, portfolios, assets if required, transactions, snapshots,
  outbox/inbox placeholders if needed for consistency;
- migration tests.

Forbidden:

- production tax tables;
- external provider tables;
- broker import tables;
- destructive migrations;
- broad schema beyond the vertical slice.

### PR 3.2 — Go API vertical-slice backend

Purpose:

- implement only the endpoints needed for portfolio creation, transaction append, snapshot rebuild,
  portfolio summary read, health and readiness.

Allowed:

- handlers, DTO mapping, application services, repositories, transaction boundaries, idempotency
  handling for financial POST endpoints;
- decimal-safe money handling;
- immutable transaction append;
- deterministic snapshot rebuild for the slice;
- unit and integration tests.

Forbidden:

- dividend services;
- tax services;
- forecast services;
- external provider clients;
- AI services;
- general-purpose abstractions not used by the slice.

### PR 3.3 — Next.js presentation slice

Purpose:

- render the first Web path using the Go API only.

Status:

- Complete / merged into `develop` at `11805cc298bba13f09f7f7af8b1e1178dc351209`.

Allowed:

- presentation screens/components for dashboard, portfolio list/detail, add transaction form, loading
  and error states;
- generated or handwritten API client layer that calls only the Go API contract;
- accessibility and basic responsive layout.

Forbidden:

- business calculations in Next.js;
- direct database access;
- route handlers replacing Go business endpoints;
- Server Actions with business behavior;
- LocalStorage, SessionStorage, or IndexedDB business persistence;
- direct MOEX, CBR, Rosstat, broker, or other provider calls.

### PR 3.4 — End-to-end verification and documentation

Purpose:

- prove the vertical slice works as one path;
- update documentation and developer onboarding.

Status:

- Complete / merged into `develop` at `86582efaa420b2c38465a5d0da041814149392c7`.

Allowed:

- integration/e2e test for the approved flow;
- README/local run instructions;
- implementation log update;
- known-risk register.

Forbidden:

- expanding functional scope beyond the first slice.

### PR 3.5 — Broker file import and reconciliation design

Purpose:

- design the smallest safe import path that reduces manual-entry friction before public MVP.

Allowed:

- documentation only;
- broker-file format inventory;
- reconciliation workflow design;
- privacy/security review;
- data-source/licensing review;
- test-vector plan.

Forbidden:

- parser implementation;
- direct broker API integration;
- credential collection;
- history mutation;
- external provider ingestion.

Status:

- Complete / merged into `develop` at `072d38d94b529221d6467502f82f03a674a7d805`.

### PR 3.6 — Broker file import vertical slice

Purpose:

- implement the smallest approved file-import review path after PR 3.5 acceptance.

Allowed:

- user-supplied CSV import only;
- normalization, duplicate detection, conflict detection, and user-review representation;
- explicit append-plan generation from approved rows;
- financial import test vectors.

Forbidden:

- public import endpoints;
- frontend upload screens;
- SQL import-session tables;
- automatic ledger append;
- XLSX and PDF parsing;
- credential scraping;
- direct broker API synchronization;
- silent mutation of existing ledger records.

Status:

- Complete / merged into `develop` at `e2b05650a4422b97d4bd924254367106b6a4686b`.

### PR 3.7 — Import append planning

Purpose:

- define the reviewed boundary for a future atomic append of user-approved import rows into the
  immutable ledger.

Allowed:

- documentation-only planning;
- allowed/forbidden implementation scope;
- idempotency, duplicate, audit, rollback, and test expectations;
- governance registry synchronization.

Forbidden:

- implementation code;
- public import endpoints;
- frontend upload screens;
- SQL import-session tables;
- automatic ledger append;
- direct broker API synchronization;
- workers, tax, mobile, AI, or external-provider integration.

Status:

- Complete / merged into `develop` at `36d86c7ff2a9c75478de155d4f60b979b8da9376`.

### PR 3.7 — Import append slice

Purpose:

- internally append user-approved import rows atomically into the immutable ledger.

Allowed:

- internal Go service/store methods;
- PostgreSQL transaction boundaries using existing tables;
- duplicate revalidation against existing ledger state;
- batch idempotency handling;
- deterministic snapshot rebuilds;
- minimal audit event creation;
- unit and integration tests.

Forbidden:

- public import endpoints;
- frontend upload UI;
- SQL import-session tables;
- direct broker API synchronization;
- workers, tax, mobile, AI, or external-provider integration.

Status:

- Complete / merged into `develop` at `89f6cab500653e09b5daa47e439b3f82fb4c8720`.

### PR 3.8 — Import review append flow planning

Purpose:

- define the internal orchestration boundary from reviewed import candidates to atomic append.

Allowed:

- planning documentation;
- future internal Go use-case boundary definition;
- privacy, financial, idempotency, snapshot, and test expectations;
- explicit allowed/forbidden implementation surfaces.

Forbidden:

- implementation code in the planning PR;
- public import endpoints;
- OpenAPI changes;
- frontend upload UI;
- SQL import-session persistence;
- raw file persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- automatic append without explicit approved decisions;
- workers, tax, mobile, AI, or Stage 3.9 work.

Status:

- Complete / merged into `develop` at `a35af2f5207bd564647d2a3fc032f4f940e62ddd`.

### PR 3.8 — Import review append flow slice

Purpose:

- internally orchestrate reviewed import decisions into atomic append.

Allowed:

- internal Go orchestration package;
- bounded in-memory CSV handling;
- source file hash calculation without raw file persistence;
- Stage 3.6 review and append-request generation;
- Stage 3.7 atomic append invocation;
- unit and live PostgreSQL integration tests;
- documentation updates.

Forbidden:

- public import endpoints;
- OpenAPI changes;
- frontend upload UI;
- SQL import-session persistence;
- raw file persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- automatic append without explicit approved decisions;
- workers, tax, mobile, AI, or Stage 3.9 work.

Status:

- Complete / merged into `develop` at `1a1d08249e252c5a3ab3f275b5fae848d5bc0e79`.

## Stage 3 domain boundaries

The first slice may touch only these bounded contexts:

- Identity: minimal local user/subject boundary required to associate data with a user.
- Investment: portfolio, transaction, asset reference needed for the slice.
- Analytics: snapshot materialization needed to return portfolio summary.
- Audit: minimal append-only audit entries for important actions if implemented in the slice.

Tax, notification, AI, mobile, and external-data contexts remain out of scope.

## Financial rules for implementation

Stage 3 implementation must preserve Stage 2 financial standards:

- no binary float for financial values;
- API decimal values are strings;
- internal decimal precision is 8 decimal places;
- display precision is 2 decimal places;
- rounding is half-even;
- financial dates use SQL `DATE` / BusinessDate semantics;
- system timestamps use UTC `TIMESTAMP WITH TIME ZONE`;
- immutable ledger entries are append-only;
- reversals/corrections must follow Stage 2 contract semantics;
- snapshots are rebuildable and cannot become the source of truth.

## Security and privacy rules

- Privacy mode remains default.
- No passport, INN, phone, address, or tax profile is required.
- No sensitive values in logs.
- No business data in browser storage.
- No secrets in Git.
- Database access is backend-only.
- Next.js receives no database credentials and no external-provider credentials.

## API-first rule

The Go API remains the only canonical business API.

If Stage 3 requires contract changes, implementation must stop and create a separate contract-change
proposal before code changes. OpenAPI changes cannot be smuggled into implementation PRs.

## Acceptance criteria

Stage 3 is complete only when all of the following are true:

- one portfolio can be created through the Go API;
- one transaction can be appended immutably;
- one snapshot can be rebuilt deterministically from canonical records;
- one portfolio summary can be read through the Go API;
- the Next.js Web shell renders that summary without business calculations;
- unit tests cover domain/application rules introduced in the slice;
- integration tests cover database persistence and API behavior;
- frontend build/typecheck passes;
- OpenAPI validation passes;
- Docker Compose validation passes;
- documentation explains how to run and verify the slice locally;
- Internal Review Agent approves;
- CI is green;
- human approval is recorded before merge.

## Required checks

Every Stage 3 implementation PR must run, at minimum:

- Go tests;
- Python tests, unless the PR provably cannot affect Python;
- frontend typecheck/build;
- OpenAPI validator;
- Docker Compose config validation;
- migration validation, once migrations exist;
- forbidden-boundary scan for Next.js;
- `git diff --check`.

## Review requirements

Each Stage 3 PR must use the standard strict review gate:

- existing Internal Review Agent conversation;
- read-only;
- no Builder trust;
- line-by-line review of changed files;
- scope regression check;
- exact `file:line`;
- severity, impact, minimal fix;
- CI evidence;
- final verdict: `APPROVED`, `REQUEST CHANGES`, or `BLOCKED — insufficient evidence`.

## Open questions

No architecture questions are opened by this planning document.

Implementation may reveal tactical questions, but they must be handled as follows:

- local implementation detail: document in the PR;
- architecture-impacting choice: stop and create/update ADR;
- contract-impacting choice: stop and create a separate OpenAPI contract change proposal.

## Stage governance

Stage 3 implementation is split into reviewable sub-stages. Each sub-stage requires its own feature
branch, checks, strict review, and human approval before merge.

Stage 3.5 is closed. Stage 3.6 is closed after:

1. the Stage 3.5 design PR is reviewed;
2. CI is green;
3. strict review approves;
4. human approval is given;
5. the design PR was merged into `develop` at `072d38d94b529221d6467502f82f03a674a7d805`;
6. requested-changes fixes were applied;
7. independent follow-up review approved;
8. PR #15 was squash-merged into `develop` at `e2b05650a4422b97d4bd924254367106b6a4686b`.
