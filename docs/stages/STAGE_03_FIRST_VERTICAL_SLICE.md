# Stage 3 — First Vertical Slice Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03 |
| Version | 0.1.0 |
| Status | Planning / Not Implemented |
| Owner | Builder Engineer |
| Supersedes | Roadmap placeholder for the first vertical slice |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-003; ADR-006; ADR-007; Stage 2 contract baseline; Web presentation baseline |
| Last Review Date | 2026-06-26 |
| Next Review Date | Before Stage 3 implementation PR |

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
- broker import;
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

Allowed:

- integration/e2e test for the approved flow;
- README/local run instructions;
- implementation log update;
- known-risk register.

Forbidden:

- expanding functional scope beyond the first slice.

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

## Stop condition

This document authorizes planning only.

Do not start Stage 3 implementation until:

1. this planning document is reviewed;
2. CI is green;
3. human approval is given;
4. the planning PR is merged into `develop`.
