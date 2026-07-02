# Stage 3.6 — Broker File Import Reconciliation Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-06 |
| Version | 0.1.0 |
| Status | In progress / implementation |
| Owner | Builder Engineer |
| Supersedes | Stage 3.6 roadmap placeholder |
| Dependencies | Stage 3.5 import design; Stage 3.2 Go vertical slice; Documents 42–43 |
| Last Review Date | 2026-07-02 |
| Next Review Date | Before Stage 3.6 merge |

## Purpose

Stage 3.6 implements the smallest safe broker-file import slice after the approved Stage 3.5 design.

The goal is to reduce manual-entry risk without introducing direct broker APIs, credential scraping,
raw file retention, SQL import-session storage, frontend upload UI, or automatic ledger mutation.

## Implemented scope

This PR implements:

- user-supplied CSV parsing;
- normalization into transaction candidates;
- deterministic row hash and normalized fingerprint generation;
- exact duplicate detection against existing portfolio transactions;
- duplicate detection inside the imported file;
- near-duplicate review-required detection;
- conflict detection for unsupported transaction types, invalid dates, invalid ticker, non-RUB rows,
  sign/shape errors, and settlement-before-trade-date;
- spreadsheet formula-injection neutralization for note/review text;
- explicit decision-to-append-plan conversion;
- financial import test vectors under `tests/financial/import/`;
- Go unit tests for valid rows, duplicate rows, conflicts, formula-injection payloads, unsafe append
  approvals, and canonical fixture loading.

## Deliberately not implemented

Stage 3.6 does not implement:

- public import API endpoints;
- frontend upload screens;
- SQL import-session tables;
- persistent raw broker-file storage;
- direct broker API synchronization;
- credential collection or scraping;
- XLSX parsing;
- PDF parsing;
- workers or collectors;
- automatic append to the ledger;
- batch idempotency tables;
- tax calculation;
- mobile code.

## Append boundary

The importer can build append requests only after explicit user decisions approve appendable rows.

It does not execute database append itself in this stage. Actual atomic database append for multiple
approved rows requires a separate reviewed implementation because it needs batch idempotency and
all-or-nothing persistence semantics. Implementing batch append by repeatedly calling the existing
single-transaction append path would violate the Stage 3.5 partial-append rule.

Therefore Stage 3.6 stops at:

```text
CSV
→ parse
→ normalize
→ duplicate/conflict detection
→ review model
→ explicit decisions
→ append request plan
```

The next implementation stage may add atomic database append only after reviewing the storage and
idempotency model.

## Privacy and security evidence

- Raw row text is processed only in memory.
- Persisted/reviewable row data uses row hash, normalized candidate, safe diagnostics, and safe note
  text.
- Formula-like note values are neutralized before they can appear in spreadsheet-compatible review
  output.
- User-uploaded files remain private user data, not approved data-source providers.
- No broker credentials or external-source calls are introduced.

## Test evidence

Implemented test-vector categories:

- valid BUY;
- valid DEPOSIT;
- valid WITHDRAWAL;
- fee/commission normalization;
- exact existing duplicate;
- broker-operation duplicate inside file;
- unsupported transaction type review;
- non-RUB conflict;
- spreadsheet formula-injection payloads;
- missing required header;
- unsafe approval rejected;
- canonical CSV fixture read.

## Known limitations

- Only CSV is supported.
- The importer is an internal Go package, not a public API.
- Existing broker operation IDs cannot be matched against persisted ledger data yet because the
  Stage 3.2 ledger schema has no broker-operation-id column.
- Atomic database append of approved import rows is intentionally deferred.
- Import-session persistence is intentionally deferred.

## Acceptance criteria

Stage 3.6 is complete when:

- importer unit tests pass;
- root `pnpm run verify` passes;
- `git diff --check` passes;
- documentation states the append boundary honestly;
- no public API, frontend, SQL migration, worker, external provider, or mobile code is added;
- strict independent review approves;
- human approval is given before merge.
