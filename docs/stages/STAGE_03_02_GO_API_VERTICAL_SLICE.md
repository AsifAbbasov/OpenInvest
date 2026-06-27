# Stage 3.2 — Go API Vertical-Slice Backend

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-02 |
| Version | 0.1.0 |
| Status | In progress |
| Owner | Builder Engineer |
| Supersedes | Stage 3.2 roadmap placeholder |
| Dependencies | `STAGE_03_FIRST_VERTICAL_SLICE.md`; `STAGE_03_01_DATABASE_FOUNDATION.md`; ADR-001; ADR-002; ADR-003; ADR-006 |
| Last Review Date | 2026-06-27 |
| Next Review Date | Before Stage 3.2 merge |

## Purpose

Stage 3.2 implements the smallest Go backend path needed to prove the first vertical slice:

```text
Go API
→ PostgreSQL
→ immutable transaction append
→ deterministic local snapshot rebuild
→ portfolio summary response
```

## Scope

Included:

- `/api/v1/health`;
- `/api/v1/ready`;
- portfolio list, create, and detail read;
- immutable transaction list and append for the first slice;
- local deterministic snapshot rebuild after transaction append, scoped by financial BusinessDate;
- portfolio summary read from the rebuildable snapshot;
- idempotency-key handling for implemented POST endpoints;
- decimal-safe fixed-scale Go value handling with eight fractional digits;
- opt-in live PostgreSQL integration test for the vertical slice.

Excluded:

- Stage 3.3 Next.js presentation implementation;
- frontend screens;
- mobile implementation;
- authentication implementation beyond the local development subject boundary;
- portfolio update/delete implementation;
- transaction correction/reversal implementation;
- dividend services;
- tax services;
- forecast services;
- external provider clients;
- workers;
- Redis integration.

## Development subject boundary

Stage 3.2 does not implement authentication. For local vertical-slice verification, the backend uses
a development-only investment subject:

```text
OPENINVEST_DEV_SUBJECT_ID
```

If the variable is not set, the backend uses a deterministic local development UUID. This is not a
production authentication model and must be replaced by the approved auth slice before production.

## Financial rules

- Go code does not use binary float for financial values.
- API decimal values are parsed as fixed-scale decimal strings with eight fractional digits.
- PostgreSQL remains the source of persisted financial state.
- Stage 3.2 snapshot methodology uses local acquisition cost as the temporary market value because
  external market-data providers remain forbidden.
- Stage 3.2 summary `asOfDate` uses financial BusinessDate semantics: the local snapshot date is the
  transaction `tradeDate`, and the rebuild includes ledger entries with `trade_date <= snapshot_date`.
- If a backdated transaction is appended after later snapshots already exist, Stage 3.2 writes new
  snapshot versions for the transaction BusinessDate and every existing later snapshot date affected by
  that append.
- The snapshot is rebuildable and is not the source of truth.

## Verification evidence

Local checks run during implementation:

- `go test ./...`;
- opt-in live PostgreSQL integration test with `OPENINVEST_DATABASE_TEST_URL`;
- Stage 3.1 migration applied to a disposable PostgreSQL container before the integration test;
- disposable PostgreSQL container and volume removed after verification.

Additional repository checks still required before PR review:

- Python tests;
- frontend typecheck/build;
- OpenAPI validator;
- Docker Compose config validation;
- migration validator;
- forbidden-boundary scan;
- `git diff --check`.

## Known limitations

- Portfolio update/delete routes remain outside Stage 3.2.
- `SELL` transaction append remains outside Stage 3.2 until weighted-average-cost/cost-basis position
  rebuild is implemented; accepting `SELL` earlier would make local acquisition-cost snapshots
  financially misleading.
- Transaction correction/reversal routes remain outside Stage 3.2.
- The summary contains no external market prices, no XIRR calculation, and no purchasing-power
  equivalents yet.
- Live database verification is local/disposable; no production migration rehearsal is claimed.

## Stop condition

After implementation, Stage 3.2 must stop for:

1. full local checks;
2. Internal Review Agent line-by-line review;
3. Draft PR into `develop`;
4. external/human approval before merge.
