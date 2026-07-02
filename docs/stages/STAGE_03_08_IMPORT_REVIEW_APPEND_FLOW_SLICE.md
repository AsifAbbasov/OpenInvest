# Stage 3.8 — Import Review Append Flow Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-08-IMPL |
| Version | 0.1.1 |
| Status | Complete / closed; merged into `develop` |
| Owner | Builder Engineer |
| Supersedes | Stage 3.8 planning-only state |
| Dependencies | Stage 3.8 planning; Stage 3.6 import reconciliation slice; Stage 3.7 import append slice |
| Last Review Date | 2026-07-03 |
| Next Review Date | 2027-01-03 |

## Purpose

Stage 3.8 implements the smallest internal application flow that connects Stage 3.6 import review
output to Stage 3.7 atomic append.

The slice remains internal-only. It does not expose a public import API, upload UI, OpenAPI change,
SQL import-session table, worker, provider integration, tax logic, mobile code, or AI feature.

## Implemented scope

This PR implements:

- an internal Go `importflow` package;
- in-memory CSV payload handling with a bounded read guard;
- source file hash calculation without persisting raw file content;
- Stage 3.6 review and append-request generation;
- explicit approved-row decision handling;
- Stage 3.7 atomic append invocation;
- deterministic result metadata:
  - parsed row count;
  - accepted row count;
  - non-appended row count;
  - appended transaction IDs;
  - snapshot dates rebuilt;
  - audit action code;
  - non-sensitive warning codes;
- unit tests for approved-only append, unsafe approval rejection, no-approved-row rejection, and
  oversized payload rejection;
- live PostgreSQL tests for full parse/review/approve/append behavior and stale duplicate rollback.

## Deliberately not implemented

Stage 3.8 does not implement:

- public import API endpoints;
- OpenAPI changes;
- frontend upload screens;
- SQL import-session persistence;
- raw file persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- background workers;
- tax calculation;
- mobile code;
- AI assistance;
- automatic append without explicit approved decisions;
- Stage 3.9 work.

## Internal Review Evidence

Changed files reviewed:

- `README.md`
- `backend-go/internal/importflow/importflow.go`
- `backend-go/internal/importflow/importflow_test.go`
- `backend-go/internal/postgres/store_integration_test.go`
- `docs/CHANGELOG.md`
- `docs/DOCUMENT_INDEX.md`
- `docs/IMPLEMENTATION_LOG.md`
- `docs/ROADMAP.md`
- `docs/SOURCE_OF_TRUTH.md`
- `docs/VERSION_MATRIX.md`
- `docs/stages/STAGE_03_08_IMPORT_REVIEW_APPEND_FLOW_PLANNING.md`
- `docs/stages/STAGE_03_08_IMPORT_REVIEW_APPEND_FLOW_SLICE.md`
- `docs/stages/STAGE_03_FIRST_VERTICAL_SLICE.md`

Review verdict:

- `APPROVED`

Blocking findings:

- Strict review found that the initial `importflow.Result` exposed the full `importer.Review`,
  including row-level details, which widened the privacy surface.

Resolved findings:

- Removed full review rows from `importflow.Result`; the result now exposes only deterministic
  non-sensitive metadata and warning codes.

Remaining non-blocking notes:

- Stage 3.8 remains internal-only. Public import endpoints and upload UI require a future reviewed
  stage.

Review-agent write access:

- Independent review agents remained read-only and did not edit, stage, commit, push, merge, or
  create PRs.

## Verification evidence

Local checks run before review:

- `GOCACHE=/private/tmp/openinvest-gocache go test ./internal/importflow ./internal/verticalslice ./internal/importer` from `backend-go/`;
- `OPENINVEST_DATABASE_TEST_URL=postgres://openinvest:openinvest-local@127.0.0.1:55432/openinvest?sslmode=disable go test ./internal/postgres -run 'TestImportReviewAppendFlow|TestStoreAppendImportedTransactions' -count=1 -v` from `backend-go/`.

The live PostgreSQL tests verify:

- parse/review/approve/append behavior;
- deterministic snapshot rebuild evidence;
- stale duplicate rejection after review but before append;
- all-or-nothing rollback without partial append;
- continued Stage 3.7 import append atomicity and concurrency coverage.

## Acceptance criteria

Stage 3.8 implementation is complete when:

- full local verification passes;
- GitHub CI is green — complete;
- strict independent review approves — complete;
- Stage report evidence is published after external review — complete;
- human approval is given before merge — complete.
