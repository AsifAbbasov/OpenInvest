# Documentation Changelog

| Field | Value |
| --- | --- |
| Document ID | REG-CHG-001 |
| Version | 1.1.16 |
| Status | Active |
| Owner | Principal Architect |
| Supersedes | None |
| Dependencies | `SOURCE_OF_TRUTH.md` |
| Last Review Date | 2026-07-01 |
| Next Review Date | 2027-01-01 |

## 2026-06-19 — Architecture Freeze v1.2

- Approved Documents 42 and 43 as the two highest-priority architecture sources.
- Resolved business-date versus UTC timestamp semantics.
- Replaced exactly-once transport language with at-least-once delivery and idempotent business processing.
- Corrected privacy terminology from pseudonymization to anonymization when re-identification is impossible.
- Froze MVP scope, asset scope, financial precision, retention, SLO boundaries, data schemas, and document precedence.
- Consolidated source documents into the repository and activated Documentation Freeze.
- Established mandatory Builder/CI/Review Agent/Human separation, Draft PR review gates, PR size budgets, ADR triggers, branch conventions, and squash-merge policy.

## 2026-06-21 — Stage 2 governance hardening

- Registered proposed ADR-006 and all Stage 2 contract artifacts without approving the ADR.
- Added the repository-owned OpenAPI validator to the pull-request CI gate.
- Reserved explicit `EXAMPLE_*` source identifiers so contract examples cannot be mistaken for
  approved MOEX, Rosstat, CBR, or other production sources.
- Synchronized the Stage 2 status across governance registries and the implementation log.

## 2026-06-25 — Stage 2 final review blockers

- Required explicit reversal `effectiveDate` BusinessDate so immutable-ledger reversals and
  snapshot rebuilds do not depend on system timestamps.
- Changed economically non-negative aggregate values from signed `Money` to `NonNegativeMoney`.
- Tightened `traceparent` validation to reject W3C-invalid version `ff`, all-zero trace IDs, and
  all-zero parent IDs.
- Documented repository OpenAPI validator limitations and added focused mutation guards instead of
  claiming complete JSON Schema 2020-12 compliance.
- Recorded auditable Principal Architect approval for the Stage 2 26-file review-size exception.
- Documented that the initial repository validator was a temporary tooling risk to be removed from
  the approved stack.

## 2026-06-25 — Stage 2 closure and ADR-006 acceptance

- Squash-merged PR #2 into `develop` at `bfde623552ebea6eac7bdaabf0d1a2263883de12`.
- Accepted ADR-006 after external review, green CI, and human approval.
- Marked Stage 2 Contract and Canonical Model Freeze as closed.
- Declared `develop` at the Stage 2 merge commit as the canonical Stage 2 baseline.

## 2026-06-25 — ADR-007 Web frontend amendment

- Replaced the current Web MVP React + Vite SPA target with Next.js App Router + TypeScript + pnpm.
- Restricted Next.js to presentation, routing, rendering, metadata, and Go API orchestration.
- Prohibited business APIs, database access, financial calculations, and external-source integration
  in Next.js.
- Kept Go as the canonical business API and Python as the future analytics/collector worker layer.
- Confirmed SwiftUI and Jetpack Compose as future-only clients with no current mobile scope.

## 2026-06-26 — Next.js Web presentation amendment closure

- Squash-merged PR #4 into `develop` at `6a7748cc24fc852d42b90b0e0cb843b6020f3973`.
- Closed the Next.js Web Presentation Amendment after internal review approval, green CI, and
  explicit human approval.
- Declared Next.js App Router + TypeScript + pnpm as the current Web presentation baseline.
- Confirmed Stage 3 remains not started; the next approved work item is Stage 3 planning.

## 2026-06-26 — Stage 3 first vertical slice planning started

- Added the Stage 3 planning document for the first portfolio/transaction/snapshot/API/Web slice.
- Kept Stage 3 implementation explicitly unauthorized until planning review and human approval.
- Defined small implementation PR boundaries to avoid scope creep.
- Reconfirmed that AI, tax export, mobile, external providers, broker import, and Stage 3 business
  expansion beyond the first slice remain out of scope.

## 2026-06-27 — Stage 3.1 database foundation started

- Squash-merged PR #6 into `develop` at `03908905b74da5c35d2fee71c2ed4956e4c06464`.
- Started the local PostgreSQL foundation for the first vertical slice.
- Added plain SQL migration pairs instead of selecting a migration library.
- Added migration validation to CI.
- Updated the local PostgreSQL 18 Docker volume mount to support live migration verification.
- Kept Go API, Next.js presentation, Python workers, and external provider integrations out of
  Stage 3.1 scope.

## 2026-06-27 — Stage 3.1 closed and Stage 3.2 started

- Squash-merged PR #7 into `develop` at `b1a3f23`.
- Started Stage 3.2 Go API Vertical-Slice Backend.
- Added the first Go API path for portfolio creation, transaction append, local snapshot rebuild,
  and summary read without changing the frozen OpenAPI contract.
- Kept frontend screens, mobile, tax, dividends, external providers, workers, and Redis out of
  Stage 3.2 scope.

## 2026-06-27 — Stage 3.2 closed and MVP product-risk refinement added

- Squash-merged PR #8 into `develop` at `8971918c8046fb9a2d6bf9f97897432cf08fbde1`.
- Closed Stage 3.2 Go API Vertical-Slice Backend after internal review approval, green CI, and
  explicit human approval.
- Added `product/MVP_PRODUCT_RISK_REFINEMENT.md` to convert PRD criticism into controlled MVP
  risk governance.
- Sharpened the initial ICP toward long-term, dividend, FIRE, and multi-account investors with
  real portfolio-accounting pain.
- Moved broker file import and reconciliation into near-term public-MVP readiness consideration
  while keeping direct broker API synchronization and credential scraping out of current scope.
- Clarified that Tax AI cannot be a calculation source; any future tax core must be deterministic
  and test-vector driven.
- Repositioned Purchasing Power as a secondary explanatory insight below capital, real return,
  dividends/coupons, and inflation-adjusted performance.

## 2026-06-27 — Product-risk refinement closed and Stage 3.3 started

- Squash-merged PR #9 into `develop` at `65bdf6537b44ed57e1c00bf68d2dacd70aa09702`.
- Closed the MVP product-risk refinement after internal review approval, green CI, independent
  review approval, and human merge approval.
- Started Stage 3.3 Next.js Presentation Slice.
- Added the first Web presentation path for portfolio list/detail, add-transaction form, summary,
  and transaction history using only the Go API.
- Kept Next.js out of business calculations, database access, provider integration, Route Handlers,
  Server Actions with business behavior, authentication implementation, and mobile scope.

## 2026-07-01 — Ruby tooling removed from active project stack

- Removed the temporary Ruby OpenAPI and migration validators from `scripts/`.
- Replaced them with Go CLI validators under `backend-go/cmd` so validation tooling stays inside
  the approved Go backend stack.
- Updated CI and local-check documentation to run the Go validators instead of Ruby scripts.

## 2026-07-01 — Stage 3.3 closed and Stage 3.4 selected as next work

- Squash-merged PR #10 into `develop` at `11805cc298bba13f09f7f7af8b1e1178dc351209`.
- Closed the Next.js Presentation Slice after independent review approval, green CI, and human
  merge approval.
- Added the first Web path for portfolio list/detail, create portfolio, add transaction, summary,
  and transaction history through the Go API boundary.
- Added local Go API CORS/OPTIONS support for explicit local Web origins and limited the transaction
  form to the transaction types currently accepted by the Go vertical slice.
- Selected Stage 3.4 end-to-end verification as the next implementation focus.

## 2026-07-01 — Stage 3.4 end-to-end verification started

- Added root pnpm commands for infrastructure, local API/Web startup, checks, and Stage 3.4 smoke
  verification.
- Added `scripts/stage-03-04-smoke.sh` to prove the local PostgreSQL → Go API → immutable
  transaction append → snapshot rebuild → summary response path.
- Added the Stage 3.4 report and synchronized governance registries.
- Kept Stage 3.4 limited to verification and onboarding; no new business logic, SQL migrations,
  provider integrations, workers, frontend feature expansion, or mobile code were added.

## 2026-07-01 — Stage 3.4 closed

- Squash-merged PR #13 into `develop` at `86582efaa420b2c38465a5d0da041814149392c7`.
- Closed Stage 3.4 after green CI, controlled local smoke evidence, and independent review approval.
- Added root verification commands and a controlled local smoke path for PostgreSQL, Go API,
  immutable transaction append, snapshot rebuild, and summary response.

## 2026-07-02 — Stage 3.5 broker-file import design started

- Started the design-only broker-file import and reconciliation stage.
- Kept parser implementation, SQL migrations, upload UI, workers, direct broker APIs, credential
  collection, and provider integrations out of scope.
- Defined the import principle as parse → normalize → match → duplicate/conflict detection → user
  review → append only.

## 2026-07-02 — Stage 3.5 closed and Stage 3.6 started

- Squash-merged PR #14 into `develop` at `072d38d94b529221d6467502f82f03a674a7d805`.
- Closed the Stage 3.5 broker-file import and reconciliation design after independent review and
  human merge approval.
- Started Stage 3.6 as an internal CSV parser/review/append-plan slice.
- Kept public import endpoints, frontend upload UI, SQL import-session tables, workers, broker APIs,
  credential scraping, XLSX/PDF parsing, and automatic ledger append out of Stage 3.6 scope.

## 2026-07-02 — Stage 3.6 broker-file import reconciliation slice closed

- Squash-merged PR #15 into `develop` at `e2b05650a4422b97d4bd924254367106b6a4686b`.
- Added the internal user-supplied CSV import parser, normalized review model, duplicate/conflict
  detection, spreadsheet-safe review fields, and explicit append-plan generation.
- Resolved independent review findings for gross amount mismatch handling, same-file near
  duplicates, broker-operation-id neutralization, duplicate decisions, and fixture coverage.
- Kept Stage 3.7 atomic import append, public import API, upload UI, SQL import-session persistence,
  broker/provider integrations, workers, XLSX/PDF parsing, mobile, and automatic ledger append out
  of scope.

## 2026-07-02 — Stage 3.7 import append planning started

- Added a documentation-only Stage 3.7 planning document for the future atomic import append scope.
- Defined the proposed append path as explicit approved import decisions → atomic database append →
  immutable ledger entries → snapshot rebuild → audit evidence.
- Kept implementation, public import endpoints, frontend upload UI, broker/provider integrations,
  workers, tax logic, mobile, and AI out of scope until a separate reviewed implementation PR.

## 2026-07-02 — Stage 3.7 import append slice started

- Added the active implementation report for the internal atomic import append slice.
- Scoped the implementation to Go service/store internals and PostgreSQL transaction behavior only.
- Kept public import API, upload UI, SQL import-session persistence, workers, broker/provider
  integrations, tax logic, mobile, and AI out of scope.

## 2026-07-02 — Stage 3.7 import append slice closed

- Squash-merged PR #18 into `develop` at `89f6cab500653e09b5daa47e439b3f82fb4c8720`.
- Added internal atomic append of user-approved import rows with idempotency protection, duplicate
  revalidation, minimal audit evidence, and deterministic snapshot rebuilds.
- Added live PostgreSQL coverage for concurrent duplicate-batch serialization.
- Kept public import API, upload UI, SQL import-session persistence, workers, provider
  integrations, tax, mobile, AI, and Stage 3.8 implementation out of scope.

## 2026-07-03 — Stage 3.8 import review append flow planning started

- Added a documentation-only planning scope for the future internal import review → append flow.
- Defined the proposed orchestration as broker CSV bytes → parse/normalize → review candidates →
  explicit accepted decisions → atomic append → snapshot rebuild → deterministic result.
- Kept public import API, OpenAPI changes, upload UI, SQL import-session persistence, raw file
  persistence, workers, broker/provider integrations, tax, mobile, AI, and implementation out of
  scope.

## 2026-07-03 — Stage 3.8 import review append flow slice started

- Added the active implementation report for the internal import review → append flow slice.
- Scoped implementation to an internal Go orchestration package and tests only.
- Kept public import API, OpenAPI changes, upload UI, SQL import-session persistence, raw file
  persistence, workers, provider integrations, tax, mobile, AI, and Stage 3.9 out of scope.

## 2026-07-03 — Stage 3.8 import review append flow slice closed

- Squash-merged PR #21 into `develop` at `1a1d08249e252c5a3ab3f275b5fae848d5bc0e79`.
- Added internal import review → append orchestration with explicit approved decisions and
  non-sensitive result metadata.
- Fixed independent review privacy finding by removing full review rows from the append result.
- Kept public import API, OpenAPI changes, upload UI, SQL import-session persistence, raw file
  persistence, workers, provider integrations, tax, mobile, AI, and Stage 3.9 out of scope.

## 2026-07-08 — Stage 3.9 import API boundary planning started

- Added a documentation-only planning scope for the future public Go API boundary for user-supplied
  broker-file import.
- Defined review/append lifecycle questions, retention decisions, idempotency boundaries, and
  privacy constraints that must be resolved before implementation.
- Kept OpenAPI changes, Go handlers, upload UI, SQL import-session persistence, raw file
  persistence, workers, provider integrations, tax, mobile, AI, and Stage 3.10 out of scope.

## 2026-07-08 — Stage 3.9 import API boundary planning closed and implementation started

- Squash-merged PR #23 into `develop` at `5cde1ca0232921d306d5e9337e4a0ba9455404ab`.
- Started the Stage 3.9 implementation slice for public Go API import review/append endpoints.
- Added a stateless API-boundary decision: review results are transient, append receives the same
  CSV payload plus explicit row decisions, and append reruns review before atomic store mutation.
- Kept raw CSV persistence, import-session tables, frontend upload UI, direct broker APIs, workers,
  tax, mobile, AI, and Stage 3.10 outside the implementation scope.

## 2026-07-08 — Stage 3.9 import API boundary slice closed

- Squash-merged PR #24 into `develop` at `b749a1632791127e0e2d4f99a91cb95eafc88898`.
- Added public Go API endpoints for transient user-supplied CSV import review and explicit append.
- Preserved the stateless boundary: no review IDs, no import-session table, no raw CSV persistence,
  and append reruns review before atomic store mutation.
- Resolved independent review findings for stale ledger revalidation, deterministic idempotent
  replay, full import append request hashing, and append-payload example validation.
- Kept frontend upload UI, SQL import-session persistence, workers, broker/provider integrations,
  tax, mobile, AI, and Stage 3.10 implementation out of scope.

## 2026-07-08 — Stage 3.10 import upload/review UI planning started

- Squash-merged Stage 3.9 closure governance into `develop` at
  `682ffd856395a6e3e988817551a512898fda2d38`.
- Started documentation-only planning for a future Next.js import upload/review UI over the existing
  Stage 3.9 Go API boundary.
- Kept Next.js implementation, business logic, OpenAPI changes, backend changes, SQL migrations,
  import-session persistence, workers, provider integrations, tax, mobile, and AI out of scope.

## 2026-07-09 — Stage 3.10 import upload/review UI slice started

- Squash-merged Stage 3.10 planning into `develop` at
  `27480d6ff22e2929e33aeac352aef8a1b01bb448`.
- Added the active implementation scope for a Next.js presentation-only import upload/review panel.
- Preserved the Go API as the only business authority; no backend contract, SQL, provider, worker,
  tax, mobile, or AI scope entered the slice.
