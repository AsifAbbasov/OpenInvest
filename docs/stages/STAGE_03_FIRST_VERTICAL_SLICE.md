# Stage 3 — First Vertical Slice Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03 |
| Version | 0.1.30 |
| Status | Complete / closed through Stage 3.16 audit-fix closure; Stage 3.17 planning tracked separately |
| Owner | Builder Engineer |
| Supersedes | Roadmap placeholder for the first vertical slice |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-003; ADR-006; ADR-007; Stage 2 contract baseline; Web presentation baseline |
| Last Review Date | 2026-08-09 |
| Next Review Date | Before Stage 3.17 privacy-lifecycle planning approval |

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
- authentication implementation outside the approved Stage 3.11 auth/privacy slice;
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

### PR 3.9 — Import API boundary planning

Purpose:

- define the future public Go API boundary for user-supplied broker-file import before any endpoint,
  OpenAPI, UI, or persistence implementation.

Allowed:

- planning documentation;
- future API lifecycle sketch;
- privacy, idempotency, audit, retention, and stale-review questions;
- explicit allowed/forbidden implementation surfaces.

Forbidden:

- implementation code;
- OpenAPI changes;
- Go handlers;
- frontend upload UI;
- SQL import-session persistence;
- raw file persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- automatic append without explicit approved decisions;
- workers, tax, mobile, AI, or Stage 3.10 work.

Status:

- Complete / merged into `develop` at `5cde1ca0232921d306d5e9337e4a0ba9455404ab`.

### PR 3.9 — Import API boundary slice

Purpose:

- expose the transient user-supplied CSV import review and explicit append lifecycle through the
  canonical Go API boundary.

Allowed:

- OpenAPI contract additions;
- Go HTTP handlers and DTOs;
- tests for API validation, privacy boundaries, idempotency, and append result behavior;
- documentation updates.

Forbidden:

- frontend upload UI;
- SQL import-session persistence;
- raw file persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- workers, tax, mobile, AI, or Stage 3.10 implementation.

Status:

- Complete / merged into `develop` at `b749a1632791127e0e2d4f99a91cb95eafc88898`;
  closure governance merged at `682ffd856395a6e3e988817551a512898fda2d38`.

### PR 3.10 — Import upload/review UI planning

Purpose:

- define the future Next.js presentation-only UI boundary for selecting a broker CSV file, sending it
  to the Go review endpoint, displaying non-sensitive review output, collecting explicit row
  approvals, and submitting append decisions to the Go append endpoint.

Allowed:

- planning documentation;
- Web route/lifecycle sketch;
- presentation-layer boundary rules;
- privacy, idempotency, accessibility, and test-scope questions;
- explicit allowed/forbidden implementation surfaces.

Forbidden:

- Next.js implementation code;
- frontend upload screens;
- OpenAPI changes;
- Go handlers;
- business logic;
- SQL migrations;
- raw file persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- workers, tax, mobile, AI, or Stage 3.10 implementation code.

Status:

- Complete / merged into `develop` at `27480d6ff22e2929e33aeac352aef8a1b01bb448`.

### PR 3.10 — Import upload/review UI slice

Purpose:

- expose the Stage 3.9 transient CSV import review/append flow in the Next.js presentation layer
  without moving business logic out of the Go API.

Allowed:

- typed frontend calls to the existing Go import review and append endpoints;
- file picker and transient in-memory CSV payload handling;
- review result display;
- explicit user approval selection for appendable rows;
- append submission for selected approved rows;
- loading, error, success, accessibility, and responsive presentation states;
- documentation updates.

Forbidden:

- OpenAPI changes;
- backend handlers;
- SQL migrations;
- import-session persistence;
- raw file persistence;
- CSV parsing business logic in Next.js;
- financial calculations;
- tax logic;
- provider integrations;
- workers, mobile, AI, or Stage 3.11 work.

Status:

- Complete / merged into `develop` at `e19a1a0ea4b0b183687bd89daabdfbc973daea71`.

### PR 3.11 — Authentication and privacy-boundary planning

Purpose:

- define the future replacement of the Stage 3 local development subject with the approved MVP web
  authentication, session, CSRF, and privacy-default boundary.

Allowed:

- planning documentation;
- implementation sequencing for registration, login, refresh, logout, privacy defaults, session
  management, and audit evidence;
- security/privacy acceptance criteria;
- exact allowed and forbidden implementation surfaces for a later PR.

Forbidden:

- implementation code;
- OpenAPI changes;
- SQL migrations;
- password hashing implementation;
- JWT or refresh-token issuance;
- frontend authentication screens or session state;
- workers, provider integrations, tax, mobile, AI, or Stage 3.11 implementation work.

Status:

- Complete / merged into `develop` at `34a31b7bb379db8a59ecc52f2cd32697be3fe125`.

### PR 3.11 — Authentication and privacy-boundary slice

Purpose:

- implement the approved MVP web auth, session, CSRF, and privacy-default boundary behind the
  Stage 2 contract while preserving the Go API as the only business authority.

Allowed:

- Go auth handlers for register, login, refresh, and logout;
- Argon2id password hashing;
- short-lived access-token issuance for Go API authorization;
- rotating HttpOnly refresh-cookie sessions;
- CSRF protection for refresh/logout;
- privacy-default persistence for new accounts;
- identity-to-investment subject mapping through existing identity/investment schemas;
- additive PostgreSQL migration for credentials, privacy settings, and sessions;
- tests and governance updates for this slice.

Forbidden:

- frontend authentication screens or session state;
- business logic in Next.js;
- OpenAPI contract changes unless separately reviewed;
- email verification, OAuth/passkeys/2FA;
- financial calculations;
- tax logic;
- workers, provider integrations, mobile, AI, or Stage 3.12 work.

Status:

- Complete / merged into `develop` at `5c49173ac858995929f266c2de991282dd194dec`.

### PR 3.12 — Web authentication UI planning

Purpose:

- define the future Next.js presentation-only registration, login, authenticated shell, refresh, and
  logout UI boundary over the Stage 3.11 Go API auth implementation.

Allowed:

- planning documentation;
- browser-session lifecycle sketch;
- token/CSRF handling constraints;
- accessibility, test, and route-protection acceptance criteria;
- explicit allowed and forbidden implementation surfaces for a later PR.

Forbidden:

- implementation code;
- Go handler changes;
- OpenAPI changes;
- SQL migrations;
- frontend auth screens;
- business logic in Next.js;
- email verification, OAuth/passkeys/2FA;
- provider integrations;
- workers, tax, mobile, AI, or Stage 3.12 implementation work.

Status:

- Complete / merged into `develop` at `25be13ce84844562e0381b79f4b81cbfed7eb44d`.

### PR 3.12 — Web authentication UI slice

Purpose:

- implement the approved Next.js presentation-only registration, login, authenticated shell,
  refresh, and logout UI boundary over the existing Stage 3.11 Go API auth implementation.

Allowed:

- presentation screens for registration and login;
- in-memory access-token handling for the active browser session;
- CSRF token handling for refresh/logout calls;
- authenticated shell state and route gating;
- typed frontend auth API calls to the Go API;
- loading, error, success, accessibility, responsive, and test coverage states;
- minimal local CORS credentials support required for the existing HttpOnly refresh cookie.

Forbidden:

- business logic in Next.js;
- Route Handlers or Server Actions for auth/business domains;
- OpenAPI contract changes;
- SQL migrations;
- access-token or refresh-token storage in JavaScript-readable durable browser storage;
- email verification, OAuth/passkeys/2FA;
- provider integrations;
- workers, tax, mobile, AI, or Stage 3.13 work.

Status:

- Complete / closed; merged into `develop` at `b4840b60346109e3cd54a07d9e1e131fc0cfad23`.

### PR 3.13 — Instrument catalog planning

Purpose:

- define the future backend-owned MVP instrument catalog boundary for MOEX shares and bonds before
  any implementation.

Allowed:

- planning document only;
- alignment with the frozen Stage 2 asset identity model;
- candidate future implementation surfaces that remain unauthorized until a separate implementation
  PR passes its normal gates;
- review criteria and acceptance criteria;
- governance registry updates.

Forbidden:

- implementation code;
- OpenAPI changes;
- SQL migrations;
- Go handler changes;
- frontend implementation;
- external provider integrations;
- market-data ingestion;
- financial calculations;
- workers, tax, mobile, AI, or Stage 3.13 implementation work.

Status:

- Complete / closed; merged into `develop` at `ca16af9adba249fc8c32c9b246b5f92f7e290b92`.

### PR 3.13 — Instrument catalog slice

Purpose:

- resolve supported MOEX share and bond tickers through the backend-owned catalog boundary using
  the frozen Stage 2 asset identity model.

Allowed:

- approved local asset fixtures for a narrow MVP demonstration set;
- backend ticker resolution through the existing `investment.assets` table;
- unsupported ticker rejection before transaction append;
- stock/bond asset-type preservation in existing snapshot buckets;
- backend and persistence tests;
- documentation updates.

Forbidden:

- OpenAPI changes;
- SQL migrations;
- Go handler changes;
- frontend implementation;
- external provider integrations;
- market-data ingestion;
- stock-card or bond-card pages, cards, or financial calculations;
- dividend or coupon calendars;
- workers, tax, mobile, AI, or Stage 3.14 work.

Status:

- Complete / closed; merged into `develop` at `b9c05fb14d0ee03e6de4dfc04ff67c16da33040b`.

### PR 3.14 — Asset search/card API boundary planning

Purpose:

- define the future Go API implementation boundary for the frozen Stage 2 asset search and asset
  detail endpoints before implementation.

Allowed:

- documentation-only planning;
- use of the Stage 3.13 backend-owned catalog as the future data source;
- explicit null-price and source-provenance policy for required Stage 2 asset fields not yet backed
  by approved source data;
- acceptance criteria, forbidden scope, and review focus for a later implementation PR;
- governance registry updates.

Forbidden:

- implementation code;
- OpenAPI changes;
- SQL migrations;
- Go handler changes;
- frontend implementation;
- external provider integrations;
- market-data ingestion;
- financial calculations;
- workers, tax, mobile, AI, or Stage 3.14 implementation work.

Status:

- Complete / closed; merged into `develop` at `2c4f7853599a455bb0cc04114b338a1145baf39c`.

### PR 3.14 — Asset search/card API boundary slice

Purpose:

- expose the frozen Stage 2 public asset search/detail API boundary through Go without fabricating
  market data, source provenance, or mandatory detail fields.

Allowed:

- Go HTTP handlers for asset search and asset detail route wiring;
- service/store read methods over the existing `investment.assets` table;
- search summaries for active canonical approved fixture rows;
- `lastPrice: null` while no market-data source is approved;
- deferred `GET /api/v1/assets/{ticker}` behavior when mandatory source/detail fields cannot be
  populated honestly;
- backend tests and documentation updates.

Forbidden:

- OpenAPI changes;
- SQL migrations;
- new catalog columns;
- frontend stock or bond cards;
- external provider integrations;
- market-data ingestion;
- price placeholders;
- runtime `EXAMPLE_*` source identifiers;
- financial calculations;
- workers, tax, mobile, AI, or Stage 3.15 work.

Status:

- Complete / closed; merged into `develop` at `57a9404952cb65693614109dd4a14d41fa5c4295`.

### PR 3.15 — Web asset discovery UI planning

Purpose:

- define the future Next.js presentation-only asset discovery entry over the Stage 3.14 Go asset
  search API before any frontend implementation.

Allowed:

- documentation-only planning;
- UI boundary definition for asset search, empty states, loading states, error states, and deferred
  detail/card states;
- ADR-007 constraints for typed frontend calls directly to the Go API;
- accessibility, responsive behavior, and privacy review criteria;
- explicit implementation acceptance criteria for a later PR.

Forbidden:

- implementation code;
- OpenAPI changes;
- SQL migrations;
- Go handler changes;
- Route Handlers or Server Actions;
- direct database access from Next.js;
- client-side market-data/provider calls;
- fabricated prices, source provenance, sector, face value, maturity, coupon type, dividends,
  coupons, yields, returns, WAC, XIRR, purchasing-power, or tax calculations;
- workers, mobile, AI, or Stage 3.15 implementation work.

Status:

- Complete / closed; merged into `develop` at `dfeab109b2825fe0e0317e87a7abf2e706a29ea6`.

### PR 3.15 — Web asset discovery UI slice

Purpose:

- implement the reviewed Next.js presentation-only asset discovery entry over the Stage 3.14 Go
  asset search/detail API boundary.

Allowed:

- typed public asset API client methods using `credentials: "omit"`;
- Next.js presentation components and route wiring under ADR-007;
- query/type search, pagination, loading, empty, error, result, and deferred detail states;
- stale-response and accepted-cursor-chain guards;
- unavailable `lastPrice` presentation and cause-neutral deferred detail handling;
- keyboard, focus, live-region, and responsive UI behavior;
- frontend tests and documentation updates.

Forbidden:

- OpenAPI changes;
- SQL migrations;
- Go handler, service, or store changes;
- Route Handlers or Server Actions;
- direct database access from Next.js;
- frontend-owned catalog fixtures or business rules;
- client-side market-data/provider calls;
- fabricated prices, source provenance, sector, face value, maturity, coupon type, dividends,
  coupons, yields, returns, WAC, XIRR, purchasing-power, or tax calculations;
- workers, mobile, or AI.

Status:

- Complete / closed; merged into `develop` at `22bede651a646d0e8b06568bda457d0626891e63`.

### PR 3.16 — Repository audit planning

Purpose:

- plan the mandatory full repository audit before the next implementation stage;
- make architecture, DDD, API, privacy/security, dependency, test, documentation, cost, and ADR
  drift visible before financial algorithms or source-backed read models begin;
- define the audit report, review evidence, and finding-resolution expectations.

Allowed:

- documentation-only planning;
- audit scope definition;
- repository-area checklist definition;
- acceptance criteria for the audit stage;
- governance updates to mark Stage 3.16 planning closed after execution.

Forbidden:

- code changes;
- OpenAPI changes;
- SQL migrations;
- dependency changes;
- WAC, XIRR, real return, inflation, purchasing-power, dividend, coupon, tax, or market-data
  implementation;
- provider integrations, workers, mobile, AI, premium, public API, or email automation.

Status:

- Closed / audit findings resolved by PR #44, squash-merged into `develop` at
  `9e6b8a753bf73ef020ce40461df25a5878344d92`. No subsequent implementation stage is authorized
  until a separately reviewed planning gate is complete.
- Stage 3.17 privacy-lifecycle planning is now the separate planning gate for the remaining
  account-deletion, anonymization, backup-destruction, and retention-execution blocker; it does not
  authorize implementation.

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
