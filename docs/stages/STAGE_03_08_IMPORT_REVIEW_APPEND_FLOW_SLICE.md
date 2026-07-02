# Stage 3.8 — Import Review Append Flow Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-08-IMPL |
| Version | 0.1.0 |
| Status | Active / implementation PR |
| Owner | Builder Engineer |
| Supersedes | Stage 3.8 planning-only state |
| Dependencies | Stage 3.8 planning; Stage 3.6 import reconciliation slice; Stage 3.7 import append slice |
| Last Review Date | 2026-07-03 |
| Next Review Date | Before Stage 3.8 merge |

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

- `WITHHELD — blind external review pending`

Review verdict:

- `WITHHELD — blind external review pending`

Blocking findings:

- `WITHHELD — blind external review pending`

Resolved findings:

- `WITHHELD — blind external review pending`

Remaining non-blocking notes:

- `WITHHELD — blind external review pending`

Review-agent write access:

- `WITHHELD — blind external review pending`

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
- GitHub CI is green;
- strict independent review approves;
- Stage report evidence is published after external review;
- human approval is given before merge.
