# Stage 3.10 — Import Upload and Review UI Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-10 |
| Version | 0.1.0 |
| Status | Draft / planning active |
| Owner | Builder Engineer |
| Supersedes | Stage 3.9 API-only import boundary state |
| Dependencies | Stage 3.9 import API boundary slice; ADR-007; Documents 42–43 |
| Last Review Date | 2026-07-08 |
| Next Review Date | 2027-01-08 |

## Purpose

Stage 3.10 plans the smallest safe Web presentation path for user-supplied broker-file import.

The goal is to let a future Next.js implementation expose the existing Go import API to a user
without moving business logic, financial calculations, reconciliation decisions, parsing rules, or
storage into the Web layer.

This planning document authorizes no implementation by itself.

## Why this stage exists

Stage 3.9 closed the public Go API boundary for transient CSV import review and explicit append.
That boundary is useful to developers, but not yet useful to a real MVP user because there is no Web
screen for:

- selecting a broker CSV file;
- sending it to the Go review endpoint;
- reading warnings, duplicates, conflicts, and accepted candidates;
- explicitly choosing what to append;
- submitting the approved decisions to the Go append endpoint;
- seeing a deterministic result after snapshot rebuild.

The next risk is therefore not backend correctness. The next risk is presentation-boundary
discipline: the Web UI must make import usable while keeping the Go API as the only business
authority.

## Proposed future Web lifecycle

A future Stage 3.10 implementation PR may add a Next.js presentation flow similar to:

```text
Portfolio detail
→ Import CSV action
→ file picker
→ POST CSV payload to Go review endpoint
→ render non-sensitive review result
→ user explicitly selects approved rows
→ POST approved decisions and idempotency key to Go append endpoint
→ render deterministic append result
```

The browser may hold the selected file and review state in memory for the current interaction only.
It must not persist raw CSV content, review rows, or business data in `localStorage`,
`sessionStorage`, IndexedDB, cookies, or Next.js server-side storage.

## Boundary decisions to preserve

- Go remains the only canonical business API.
- Next.js remains presentation only under ADR-007.
- CSV parsing, normalization, duplicate detection, conflict detection, append validation, ledger
  mutation, snapshot rebuild, idempotency enforcement, and audit evidence remain backend concerns.
- Browser-side validation is only a UX hint and never a trust boundary.
- Any contract gap discovered during implementation must stop the PR and become a separate OpenAPI
  contract-change proposal.
- Raw broker-file content must not be logged or persisted by the Web layer.
- User append approval must remain explicit.

## Allowed implementation surfaces for a future PR

A future Stage 3.10 implementation PR may include only after this plan is reviewed and merged:

- Next.js App Router page or route segment for the import UI;
- presentation components for file selection, review result display, decision selection, loading,
  error, and success states;
- typed calls to the existing Go API through the OpenAPI boundary;
- client-side accessibility and responsive layout work;
- non-sensitive UI state held in memory for the current interaction;
- tests for rendering, boundary behavior, and error states;
- documentation updates.

## Forbidden in the planning PR

This planning PR must not introduce:

- Next.js implementation code;
- frontend upload screens;
- Go handlers;
- OpenAPI changes;
- business logic;
- frontend CSV parsing beyond future technical file-read plumbing;
- financial calculations;
- tax logic;
- repositories;
- SQL migrations;
- PostgreSQL or Redis access from Next.js;
- Next.js Route Handlers or Server Actions for business domains;
- raw file persistence;
- `localStorage`, `sessionStorage`, or IndexedDB business persistence;
- direct broker API synchronization;
- credential scraping;
- external provider integrations;
- XLSX or PDF parsing;
- workers;
- mobile code;
- AI assistance.

## Questions that must be answered before implementation

Before a Stage 3.10 implementation PR starts, the builder must confirm:

- the exact Web route and navigation entry point;
- whether the UI uses the existing portfolio detail page or a dedicated import route;
- maximum UI file-size hint, aligned with the Go API limit;
- which review fields are safe to render directly;
- how row approval selection is represented without duplicating backend decision logic;
- how idempotency keys are generated for append submission without becoming business state;
- how the UI handles refresh/navigation while keeping raw CSV data non-persistent;
- what test coverage proves no business logic entered the Web layer.

## Acceptance criteria for this planning PR

This planning PR is complete when:

- the future Web import UI boundary is documented;
- allowed and forbidden implementation surfaces are explicit;
- ADR-007 presentation-only boundaries are preserved;
- governance registries reference the planned stage;
- stale Stage 3.9 planning status is corrected;
- no implementation code, OpenAPI contract, SQL, backend, worker, or mobile file is changed;
- independent review approves;
- human approval is given before merge.
