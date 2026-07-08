# Stage 3.9 — Import API Boundary Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-09 |
| Version | 0.1.0 |
| Status | Draft / planning only |
| Owner | Builder Engineer |
| Supersedes | Internal-only import flow state |
| Dependencies | Stage 3.6 import reconciliation slice; Stage 3.7 import append slice; Stage 3.8 import review append flow slice; Documents 42–43 |
| Last Review Date | 2026-07-08 |
| Next Review Date | Before Stage 3.9 implementation approval |

## Purpose

Stage 3.9 plans the public Go API boundary for user-supplied broker-file import.

The goal is to define the smallest safe API lifecycle that can expose the existing internal import
capabilities without weakening privacy, explicit user confirmation, immutable ledger semantics, or
OpenAPI-first governance.

This planning document authorizes no implementation by itself.

## Why this stage exists

Stages 3.6–3.8 proved the internal import pipeline:

```text
CSV payload
→ parse and normalize
→ review candidates
→ explicit accepted decisions
→ atomic append
→ deterministic result
```

The next risk is not parsing or append correctness. The next risk is public boundary design:

- where uploaded bytes may enter the backend;
- how users review rows before append;
- how idempotency applies across review and append;
- how raw file content is kept out of persistence and logs;
- how the Web presentation layer will later call the Go API without owning business logic.

Stage 3.9 plans that boundary before any endpoint or OpenAPI change is implemented.

## Proposed future API lifecycle

A future implementation stage may expose a Go API lifecycle similar to:

```text
POST /api/v1/portfolios/{portfolioId}/imports/review
→ returns reviewId or transient review result

POST /api/v1/portfolios/{portfolioId}/imports/{reviewId}/append
→ accepts explicit approved row decisions
→ appends approved rows atomically
```

This is a planning sketch, not an approved contract. The exact OpenAPI paths, request schemas,
response schemas, retention rules, and persistence model require a separate implementation PR and
review.

## Boundary decisions to preserve

- Go remains the only canonical business API.
- Next.js remains presentation only.
- Raw broker-file content must not be permanently stored.
- Any temporary import state must have an explicit retention policy before implementation.
- No automatic append is allowed.
- Every appended row must come from an explicit approved decision.
- Idempotency keys are required for financial mutation endpoints.
- Duplicate and conflict checks must be rerun at append time.
- Snapshot rebuild must remain deterministic and based on appended transaction business dates.
- Audit evidence may record counts, hashes, IDs, and outcomes, but not raw CSV content.

## Allowed implementation surfaces for a future PR

A future Stage 3.9 implementation PR may include only after this plan is approved:

- OpenAPI contract additions for the import review/append lifecycle;
- Go HTTP handlers that call existing internal import flow services;
- request/response DTOs for review summaries, row decisions, append results, and errors;
- route-level validation and idempotency handling;
- tests for API validation, privacy boundaries, idempotency, and failure cases;
- documentation updates.

## Forbidden in this planning PR

This planning PR must not introduce:

- OpenAPI changes;
- Go handlers;
- frontend upload UI;
- SQL migrations;
- import-session persistence;
- raw file persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- background workers;
- tax calculation;
- mobile code;
- AI assistance;
- Stage 3.10 or later work.

## Questions that must be answered before implementation

Before implementation starts, the next PR must decide:

- whether review results are fully transient or stored as short-lived import sessions;
- if stored, exact retention and deletion policy;
- maximum upload size;
- accepted content types;
- exact OpenAPI request/response schemas;
- whether review and append are one endpoint with preview mode or two endpoints;
- how review IDs are generated if transient sessions are introduced;
- how to prevent replay or append of stale review decisions;
- how to expose row-level warnings without leaking sensitive raw file content.

## Acceptance criteria for this planning PR

This planning PR is complete when:

- the future API boundary is documented;
- allowed and forbidden implementation surfaces are explicit;
- privacy, idempotency, audit, and retention questions are listed;
- governance registries reference the planned stage;
- no implementation code or OpenAPI contract is changed;
- independent review approves;
- human approval is given before merge.
