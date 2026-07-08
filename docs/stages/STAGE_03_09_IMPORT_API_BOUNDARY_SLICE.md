# Stage 3.9 — Import API Boundary Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-09-IMPL |
| Version | 0.1.1 |
| Status | Closed / merged into `develop` |
| Owner | Builder Engineer |
| Supersedes | Stage 3.9 planning-only state |
| Dependencies | Stage 3.9 planning; Stage 3.8 import review append flow; Documents 42–43 |
| Last Review Date | 2026-07-08 |
| Next Review Date | 2027-01-08 |

## Purpose

Stage 3.9 exposes the existing user-supplied CSV import flow through the canonical Go API boundary.

The implementation is intentionally narrow:

```text
JSON CSV payload
→ transient backend review
→ explicit row decisions
→ append endpoint reruns review
→ atomic immutable ledger append
```

## Implemented API boundary

The slice adds two Go API endpoints:

```text
POST /api/v1/portfolios/{portfolioId}/imports/review
POST /api/v1/portfolios/{portfolioId}/imports/append
```

`imports/review` parses and normalizes a user-supplied CSV payload and returns a transient review
result. `imports/append` accepts the same CSV payload plus explicit row decisions, reruns review, and
uses the existing Stage 3.7 atomic append path.

## Why there is no reviewId in this slice

Stage 3.9 deliberately does not create import sessions or review IDs. A review ID would require a
persistence and retention model for temporary import state. That belongs in a later reviewed stage.

Instead, this slice uses a stateless boundary:

- raw CSV enters only in the request body;
- raw CSV is not stored;
- append reruns review instead of trusting client-held review state;
- the database append layer reruns duplicate/conflict checks before writing.

This is less convenient for a future upload UI, but it preserves privacy and avoids premature SQL
schema expansion.

## Privacy and retention

- Raw CSV content is not persisted.
- HTTP responses include hashes, counts, row status, normalized candidate data, and non-sensitive
  warning codes.
- Broker operation identifiers are not returned by the HTTP review response.
- Audit evidence remains limited to counts, hashes, IDs, request metadata, and outcomes.

## Idempotency and append safety

- `imports/review` is a transient preflight operation and does not require `Idempotency-Key`.
- `imports/append` is a financial mutation and requires `Idempotency-Key`.
- Append reruns review using the submitted CSV payload and explicit decisions.
- Append calls the existing atomic store boundary, which protects idempotency and duplicate
  revalidation under the portfolio lock.
- Exact replay with the same authenticated principal, path, idempotency key, and canonical append
  payload returns the original imported transactions instead of reprocessing duplicate-sensitive
  review state.

## Current limitation

The review endpoint compares against the currently exposed transaction list limit of 100 rows.
This review result is therefore a preflight user aid, not the final write guarantee. The append
endpoint reruns validation and the store performs authoritative duplicate/conflict protection before
any ledger mutation.

## Explicitly out of scope

This slice does not add:

- frontend upload UI;
- SQL import-session persistence;
- raw file persistence;
- direct broker API synchronization;
- broker credential collection or scraping;
- external provider integrations;
- XLSX/PDF parsing;
- workers;
- tax logic;
- mobile code;
- AI assistance;
- Stage 3.10 work.

## Verification plan

Before merge, the branch must show:

- Go unit tests for import review/append HTTP boundary;
- OpenAPI validator pass;
- root `pnpm run verify` pass;
- strict independent review approval;
- human approval before merge.

## Closure evidence

Stage 3.9 implementation was squash-merged through PR #24 into `develop` at
`b749a1632791127e0e2d4f99a91cb95eafc88898`.

Merge gates completed:

- GitHub CI green for head commit `92a16d23bdb015d0b5466f7dcf71fc354016239a`;
- independent external review approved the follow-up fix commit;
- review findings for current-ledger revalidation, deterministic replay, full idempotency hashing,
  and append-payload example validation were resolved;
- no `main` branch change was made.
