# Stage 3.14 — Asset Search and Card API Boundary Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-14-ASSET-API-BOUNDARY-SLICE |
| Version | 0.1.3 |
| Status | Complete / closed |
| Owner | Builder Engineer |
| Supersedes | Stage 3.14 asset search/card API boundary planning |
| Dependencies | `SOURCE_OF_TRUTH.md`; Stage 2 contract baseline; Stage 3.13 instrument catalog slice; Stage 3.14 planning |
| Last Review Date | 2026-07-26 |
| Next Review Date | Before Stage 3.15 planning approval |

## Purpose

Stage 3.14 implements the smallest honest Go API boundary for public asset discovery over the
backend-owned Stage 3.13 catalog.

The slice exposes:

- `GET /api/v1/assets/search`;
- `GET /api/v1/assets/{ticker}` as a wired but intentionally deferred detail boundary.

It does not add market data, external providers, frontend stock/bond cards, SQL migrations, or
financial calculations.

## Implemented scope

- Added application DTOs for `AssetSearchFilter` and `AssetSummary`.
- Added `SearchAssets` to the Go application service and store boundary.
- Added public HTTP route `GET /api/v1/assets/search`.
- Added public HTTP route `GET /api/v1/assets/{ticker}` that validates the ticker but returns
  `404 NOT_FOUND` until mandatory runtime source provenance and required detail fields can be
  populated without fabricated data.
- Added PostgreSQL read logic over the existing `investment.assets` table.
- Returned only active rows that match the approved Stage 3.13 fixture metadata.
- Returned `lastPrice: null` in search results because no approved market-data source exists.
- Implemented deterministic opaque cursor pagination for asset search using ticker ordering and
  `limit + 1` over-fetch after canonical fixture predicates are applied in SQL.
- Preserved the frozen search semantics: ticker prefix matching and case-insensitive name fragment
  matching.
- Rejected malformed, empty, whitespace-only, or out-of-range `limit` values instead of silently
  normalizing invalid client input.
- Rejected empty or whitespace-only optional `assetType` and `cursor` query parameters when they are
  supplied by the client.
- Validated opaque cursor tokens exactly as supplied, with no whitespace trimming, no padded token
  normalization, and a 512-character maximum before Base64 decoding.
- Added unit and integration tests for search validation, route behavior, null price/source
  boundaries, cursor pagination, asset-type filtering, and inactive/conflicting fixture exclusion.

## Explicit exclusions

This slice does not add:

- OpenAPI contract changes;
- SQL migrations;
- new catalog columns;
- market-data ingestion;
- live prices;
- price placeholders;
- runtime `EXAMPLE_*` source identifiers;
- provider integrations;
- workers;
- frontend stock or bond pages;
- Next.js Route Handlers or Server Actions;
- financial calculations;
- tax logic;
- mobile implementation;
- AI functionality;
- Stage 3.15 scope.

## Detail endpoint decision

The frozen Stage 2 asset-card response requires fields that the current database and approved
fixture catalog cannot honestly provide:

- `source`;
- stock `sector`;
- bond `faceValue`;
- bond `maturityDate`;
- bond `couponType`.

The Data Source Registry still has no approved production market-data or issuer-data provider.
Runtime responses must never emit reserved `EXAMPLE_*` identifiers.

Therefore, Stage 3.14 wires `GET /api/v1/assets/{ticker}` for API-boundary completeness but returns
`404 NOT_FOUND` for valid tickers until a later reviewed source-governance, catalog-expansion, or
contract-change stage can populate every mandatory field. This preserves the frozen API surface
without fabricating asset-card facts.

## Verification evidence

Completed before PR #38 merge:

- Local Go targeted tests passed:
  `go test ./cmd/api ./internal/auth ./internal/httpapi ./internal/postgres ./internal/verticalslice`.
- Full repository verification passed:
  `pnpm run verify`.
- Whitespace/diff validation passed:
  `git diff --check`.
- GitHub CI passed on PR #38:
  Go tests, Python tests, frontend build/typecheck, OpenAPI contract, PostgreSQL migration
  validation, and Docker Compose config.
- Final reviewed feature-branch head:
  `fa8d4a8ce798948fee307fed15c8fe78cf3dc716`.
- Squash merge into `develop`:
  `57a9404952cb65693614109dd4a14d41fa5c4295`.
- Merge date:
  2026-07-14.

## Internal Review Evidence

- Review channel:
  strict independent separate-window Codex review.
- Reviewed scope:
  Stage 3.14 Go API asset search/card boundary implementation diff and follow-up fixes.
- Final implementation verdict:
  `APPROVED`.
- Blocking findings resolved before approval:
  - canonical fixture predicates, including ISIN, must run before SQL `LIMIT`;
  - ticker search must use prefix semantics while name search may use fragment semantics;
  - invalid and supplied-empty `limit`, `assetType`, and `cursor` query parameters must be rejected;
  - cursor validation must use the exact supplied token, enforce the 1–512 byte boundary before
    Base64 decoding, and reject whitespace, padded, malformed, or oversized tokens;
  - Source of Truth version references must remain synchronized.
- Remaining non-blocking notes:
  asset search still discovers only active canonical rows already present in `investment.assets`,
  and asset detail remains deferred until approved source provenance and mandatory detail fields
  exist.
- Review Agent write authority:
  read-only only; the Review Agent did not edit, stage, commit, push, merge, or create/update PRs.

## Known risks

- Asset search currently discovers only active canonical rows already present in `investment.assets`.
  A future reviewed catalog seeding/read-model stage may be needed for full public discovery before
  the user has appended any transaction.
- Asset detail remains intentionally deferred until source provenance and mandatory detail fields
  are available.

## Recommended next step

Stage 3.14 implementation and closure governance are closed. Continue with Stage 3.15 Web asset
discovery UI planning from `develop` at `f5289eb604b8ba31aa422d0d09950da02e0f48b3`, and preserve
the same review gates before any implementation.
