# Stage 3.7 — Import Append Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-07-IMPL |
| Version | 0.1.0 |
| Status | Active / implementation PR |
| Owner | Builder Engineer |
| Supersedes | Stage 3.7 planning-only state |
| Dependencies | Stage 3.7 import append planning; Stage 3.6 import reconciliation slice; Stage 3.2 Go vertical slice |
| Last Review Date | 2026-07-02 |
| Next Review Date | Before Stage 3.7 merge |

## Purpose

Stage 3.7 implements the smallest internal atomic append path for user-approved import rows.

It does not expose a public import API or UI. The slice exists to prove that an append plan created
by the Stage 3.6 importer can be applied to the immutable ledger safely, atomically, and with
deterministic snapshot rebuilds.

## Implemented scope

This PR implements:

- internal Go service method for appending an approved import batch;
- internal PostgreSQL store method that executes the batch in one database transaction;
- application-boundary validation of every imported append request;
- rejection of duplicate approved rows inside one batch;
- exact existing-ledger duplicate revalidation before persistence;
- all-or-nothing rollback when any approved row is no longer safe to append;
- idempotency-key reservation for the batch command;
- rejection of repeated batch execution without appending duplicate ledger entries;
- deterministic snapshot rebuild for every unique imported trade date;
- minimal audit event for successful import append batches;
- unit tests and live PostgreSQL integration coverage.

## Deliberately not implemented

Stage 3.7 does not implement:

- public import API endpoints;
- frontend upload screens;
- raw broker-file persistence;
- SQL import-session tables;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- background workers;
- tax calculation;
- mobile code;
- AI assistance;
- correction or reversal semantics.

## Atomic append boundary

The batch append operation is internal-only and accepts append requests that were already produced
from an explicit review/decision process.

The operation either:

- validates every row, appends every ledger entry, rebuilds snapshots, records audit evidence, and
  commits; or
- rolls back the entire transaction.

No partial append is allowed.

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

- `GOCACHE=/private/tmp/openinvest-gocache go test ./...` from `backend-go/`;
- `OPENINVEST_DATABASE_TEST_URL=postgres://openinvest:openinvest-local@127.0.0.1:55432/openinvest?sslmode=disable go test ./internal/postgres -run 'TestStoreAppendImportedTransactions' -count=1 -v` from `backend-go/`.

The live PostgreSQL test verifies:

- multi-row batch append;
- snapshot rebuild after import;
- repeated idempotency-key rejection without duplicate append;
- concurrent duplicate-batch serialization with different idempotency keys: one batch succeeds, one batch is rejected, and exactly one ledger row is persisted;
- rollback when a later row conflicts with existing ledger state;
- minimal import-append audit event creation.

## Acceptance criteria

Stage 3.7 implementation is complete when:

- full local verification passes;
- GitHub CI is green;
- strict independent review approves;
- Stage report evidence is published after external review;
- human approval is given before merge.
