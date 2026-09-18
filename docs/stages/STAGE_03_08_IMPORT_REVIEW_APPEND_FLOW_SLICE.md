# Stage 3.8 — Import Review Append Flow Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-08-IMPL |
| Version | 0.1.1 |
| Status | Complete / closed; merged into `develop` |
| Supersedes | Stage 3.8 planning-only state |
| Dependencies | Stage 3.8 planning; Stage 3.6 import reconciliation slice; Stage 3.7 import append slice |

## Purpose

Stage 3.8 implements the smallest internal application flow that connects Stage 3.6 import review
output to Stage 3.7 atomic append.

The slice remains internal-only. It does not expose a public import API, upload UI, OpenAPI change,
SQL import-session table, worker, provider integration, tax logic, mobile code, or  feature.

## Implemented scope

This PR implements:


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
-  assistance;
- automatic append without explicit approved decisions;
- Stage 3.9 work.

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
