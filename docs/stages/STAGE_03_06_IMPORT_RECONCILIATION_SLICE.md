# Stage 3.6 — Broker File Import Reconciliation Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-06 |
| Version | 0.1.2 |
| Status | Complete / merged into `develop` |
| Owner | Builder Engineer |
| Supersedes | Stage 3.6 roadmap placeholder |
| Dependencies | Stage 3.5 import design; Stage 3.2 Go vertical slice; Documents 42–43 |
| Last Review Date | 2026-07-02 |
| Next Review Date | Before Stage 3.7 import append scope approval |

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

## Completion evidence

Stage 3.6 was squash-merged into `develop` at
`e2b05650a4422b97d4bd924254367106b6a4686b` after:

- green GitHub CI;
- focused local verification;
- independent review requested-changes fixes;
- independent follow-up review approval;
- human authorization to continue.

## Internal Review Evidence

Changed files reviewed:

- PR #15 implementation, fixtures, tests, and documentation for the Stage 3.6 import reconciliation
  slice;
- PR #16 governance-only closure documentation for Stage 3.6 status synchronization.

Review verdict:

- Initial strict independent line-by-line review: `REQUEST CHANGES`;
- follow-up strict independent worktree review after fixes: `APPROVED`;
- PR #16 closure documentation review: `REQUEST CHANGES` for missing audit evidence and PR
  disclosure fields; this section resolves the stage-report evidence finding.

Blocking findings found during Stage 3.6 review:

- BUY gross-amount mismatch was treated as appendable instead of conflict review;
- same-file near duplicates were not detected;
- `broker_operation_id` was not neutralized for spreadsheet-compatible exposure;
- duplicate append decisions could emit duplicate append requests;
- conflict and formula fixtures were not fully exercised, and one conflict fixture contained a
  misleading appendable row.

Resolved findings:

- All Stage 3.6 implementation findings were resolved in
  `ead67eae6341802ed13990a62379588f85fda2a6`;
- the follow-up review verified the focused importer suite, broader Go suite, scope boundaries, and
  absence of Stage 3.7 implementation before approval;
- PR #16 disclosure-field governance evidence is maintained in the GitHub PR description.

Remaining non-blocking notes:

- No Stage 3.6-specific blocking risk remains;
- the existing Python/FastAPI warning is unrelated to the import slice and remains non-blocking.

Review-agent write access:

- Review agents remained read-only and did not edit files, stage files, commit, push, merge, or
  update PR metadata.

Evidence locations:

- Stage 3.6 implementation PR: `#15`;
- Stage 3.6 closure PR: `#16`;
- strict follow-up review thread: `019f2306-07d2-7060-be70-01bd46e8a1ad`;
- strict PR #16 closure review thread: `019f230f-d446-7ed1-b207-472c71237297`.

## Acceptance criteria

Stage 3.6 is complete when:

- importer unit tests pass;
- root `pnpm run verify` passes;
- `git diff --check` passes;
- documentation states the append boundary honestly;
- no public API, frontend, SQL migration, worker, external provider, or mobile code is added;
- strict independent review approves;
- human approval is given before merge.
