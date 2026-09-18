# Stage 3.8 — Import Review Append Flow Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-08 |
| Version | 0.1.1 |
| Status | Complete / merged into `develop` |
| Supersedes | Stage 3.7 isolated append-only state |
| Dependencies | Stage 3.6 import reconciliation slice; Stage 3.7 import append slice; Documents 42–43 |

## Purpose

Stage 3.8 plans the smallest safe integration between the Stage 3.6 import review output and the
Stage 3.7 atomic append operation.

The goal is not to expose import to users yet. The goal is to prove an internal application flow:

```text
broker CSV bytes
→ parse and normalize
→ review candidates
→ accept explicit user-approved rows
→ append atomically
→ rebuild snapshots
→ return deterministic result
```

This planning document authorizes no implementation by itself.

## Why this stage exists

Stage 3.6 can produce reviewable normalized import candidates and append plans.

Stage 3.7 can atomically append already-approved import rows into the immutable ledger.

Without Stage 3.8, those capabilities remain adjacent but not proven as one internal product path.
Stage 3.8 closes that gap while preserving the project boundary: no public upload endpoint and no
frontend import UI yet.

## Proposed Stage 3.8 implementation outcome

A future Stage 3.8 implementation PR may add an internal application service or use-case function
that:


boundaries.

## Allowed implementation surfaces for the future PR

The future Stage 3.8 implementation PR may include only:

- Go application/use-case orchestration code;
- unit tests for candidate approval mapping and rejection handling;
- PostgreSQL integration tests proving the full internal flow appends atomically;
- test fixtures using sanitized CSV data already compatible with Stage 3.6;
- documentation updates for the implemented internal flow.

## Forbidden in Stage 3.8 implementation

Stage 3.8 must not introduce:

- public import API endpoints;
- OpenAPI changes;
- frontend upload UI;
- SQL migrations or import-session persistence;
- raw broker-file persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- background workers;
- tax calculation;
- mobile code;
-  assistance;
- automatic append without explicit approved decisions;
- Stage 3.9 or later work.

## Privacy and security requirements

- Raw file content must not be logged.
- Raw file content must not be persisted.
- Review warnings must be non-sensitive.
- Audit events may record counts, hashes, and resulting transaction IDs, but not raw CSV content.
- Formula-injection neutralization from Stage 3.6 must remain active for exported/reviewable fields.
- The caller must explicitly approve each row that is appended.

## Financial correctness requirements

- Decimal and Money rules remain unchanged.
- Business dates must remain SQL `DATE` semantics.
- The append operation must remain immutable and append-only.
- Any conflict discovered after review but before append must reject the whole batch.
- Snapshot rebuild dates must be deterministic and derived from appended transaction business dates.
- Repeated idempotency keys must not duplicate ledger entries.

## Test expectations

Before Stage 3.8 implementation can merge:


## Acceptance criteria for this planning PR

This planning PR is complete when:

- the Stage 3.8 scope is documented;
- allowed and forbidden surfaces are explicit;
- governance registries reference the planned stage;
- no implementation code is changed;
