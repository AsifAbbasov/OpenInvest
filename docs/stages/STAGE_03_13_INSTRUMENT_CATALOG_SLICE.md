# Stage 3.13 — Instrument Catalog Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-13-INSTRUMENT-CATALOG-SLICE |
| Version | 0.1.1 |
| Status | Active / implementation |
| Owner | Builder Engineer |
| Supersedes | Stage 3.13 instrument catalog planning |
| Dependencies | `SOURCE_OF_TRUTH.md`; Stage 2 contract baseline; Stage 3.13 instrument catalog planning |
| Last Review Date | 2026-07-12 |
| Next Review Date | Before Stage 3.13 implementation review |

## Purpose

Stage 3.13 implements the smallest backend-owned catalog boundary for resolving supported MOEX
share and bond tickers against the frozen Stage 2 asset model.

The slice keeps the existing API shape and existing database schema. It does not add stock or bond
cards, external provider ingestion, live market data, frontend presentation work, or new contract
surfaces.

## Implementation Scope

This slice may include only:

- deterministic local asset fixtures for a narrow reviewed MVP demonstration set;
- backend resolution of supported tickers through the existing `investment.assets` table;
- rejection of unsupported tickers before transaction append;
- preservation of backend-owned stock/bond asset type in existing snapshot buckets;
- backend and persistence tests for catalog lookup, rejection, idempotency compatibility, and
  asset-type classification;
- documentation updates.

## Explicit Exclusions

This slice does not add:

- OpenAPI contract changes;
- SQL migrations;
- Go handler changes;
- frontend implementation;
- Next.js Route Handlers or Server Actions;
- direct database access from Next.js;
- external MOEX, CBR, Rosstat, broker, or provider calls;
- workers or scheduled collectors;
- market-data ingestion, prices, candles, order books, dividends, or coupons;
- stock-card or bond-card pages, cards, or financial calculations;
- WAC, XIRR, real return, inflation, purchasing-power, or tax calculations;
- mobile implementation;
- AI functionality.

## Work Completed So Far

- Replaced implicit asset creation for arbitrary valid tickers with approved local fixture
  resolution.
- Added deterministic fixtures for a minimal share/bond demonstration set.
- Rejected unsupported tickers with `verticalslice.ErrInvalidInput`.
- Rejected noncanonical whitespace-padded tickers instead of silently normalizing beyond the frozen
  OpenAPI ticker contract.
- Kept asset resolution inside the existing PostgreSQL store boundary and existing
  `investment.assets` table.
- Seeded approved assets without rewriting or reactivating existing catalog rows during user append.
- Resolved unique import-batch assets in deterministic ticker order before ledger insertion.
- Preserved stock/bond classification in existing snapshot buckets by using backend-owned asset
  type.
- Added tests for approved fixture definitions, unsupported and noncanonical ticker rejection,
  approved bond metadata, inactive and conflicting active fixture behavior, legacy UUID-compatible
  canonical metadata matching, and stock/bond bucket classification.

## Verification So Far

- `backend-go`: `GOCACHE=/private/tmp/openinvest-gocache go test ./internal/postgres ./internal/verticalslice`
- `backend-go`: `OPENINVEST_DATABASE_TEST_URL='postgres://openinvest:openinvest-local@127.0.0.1:55432/openinvest?sslmode=disable' GOCACHE=/private/tmp/openinvest-gocache go test ./internal/postgres -run 'TestStoreAppendTransaction(RejectsUnsupportedTicker|SeedsApprovedBondFixture|AcceptsCanonicalFixtureWithLegacyID|DoesNotReactivateInactiveFixture|RejectsConflictingActiveFixture)|TestStoreVerticalSlice|TestStoreAppendImportedTransactionsIsAtomicAndIdempotent' -count=1 -v`
- Repository root: `GOCACHE=/private/tmp/openinvest-gocache UV_CACHE_DIR=/private/tmp/openinvest-uv-cache pnpm run verify`

## Internal Review Evidence

- Changed files reviewed: `WITHHELD — blind external review pending`.
- Review verdict: `WITHHELD — blind external review pending`.
- Blocking findings: `WITHHELD — blind external review pending`.
- Resolved findings: `WITHHELD — blind external review pending`.
- Remaining non-blocking notes: `WITHHELD — blind external review pending`.
- Confirmation that Review Agent did not edit code:
  `WITHHELD — blind external review pending`.
- Follow-up review evidence may be documented only after the independent external verdict is
  complete.
- Runtime lookup accepts an existing active backend-owned asset row with matching canonical ticker,
  type, name, currency, market, lifecycle, ISIN, and lot-size metadata even when its internal UUID
  predates the deterministic fixture seed. Conflicting metadata remains rejected.
- Catalog-mutation integration tests must restore shared database state with checked cleanup
  operations before merge.

## Review Focus

Review must verify:

- unsupported tickers cannot silently create canonical assets;
- approved fixtures stay narrow, deterministic, and backend-owned;
- no OpenAPI, SQL migration, handler, frontend, provider, worker, market-data, tax, mobile, or AI
  scope entered the slice;
- existing idempotency and import append behavior remains compatible;
- stock and bond type classification uses backend-owned asset metadata without introducing new
  market-data or instrument-card calculations.

## Recommended Next Step

Rerun targeted and full verification, then request strict follow-up code review before commit or PR.
