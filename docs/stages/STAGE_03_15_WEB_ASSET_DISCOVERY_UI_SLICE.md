# Stage 3.15 — Web Asset Discovery UI Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-15-WEB-ASSET-DISCOVERY-UI-SLICE |
| Version | 0.1.3 |
| Status | Complete / closed |
| Owner | Builder Engineer |
| Supersedes | Stage 3.15 Web asset discovery UI planning |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-007; Stage 3.14 asset API boundary slice; Stage 3.15 planning |
| Last Review Date | 2026-07-27 |
| Next Review Date | Before changing asset discovery, market-data, or stock/bond card behavior |

## Purpose

Stage 3.15 implements the smallest reviewed Next.js presentation boundary for discovering supported
assets through the existing public Go asset API.

The slice keeps catalog authority, market data, provider access, and financial calculations outside
Next.js.

## Implemented scope

- Added typed public asset API client methods for `GET /api/v1/assets/search` and
  `GET /api/v1/assets/{ticker}`.
- Required public asset calls to use `credentials: "omit"` and avoid bearer authorization, CSRF
  headers, cookies, and browser storage reads.
- Added a presentation-only `/assets` App Router page.
- Added an authenticated-shell navigation entry from the portfolio dashboard to asset discovery.
- Added asset search UI for query, asset type, loading, empty, error, result, and pagination states.
- Rendered `lastPrice: null` as unavailable without fabricating zero, stale, or live price values.
- Added a deferred asset-detail state for `404 NOT_FOUND` without claiming a specific backend cause.
- Added stale-response and accepted-cursor-chain guards for search and pagination state.
- Added keyboard, focus, Escape, and live-region behavior required by the Stage 3.15 planning gate.
- Added focused frontend tests for public asset request construction, search-state invariants,
  detail-generation invalidation, focus-entry decisions, focus restoration decisions, Escape
  behavior, and status announcements, plus source-level component wiring checks for the observable
  accessibility contracts.

## Explicit exclusions

This slice does not add:

- OpenAPI changes;
- SQL migrations;
- Go handler, service, or store changes;
- Next.js Route Handlers or Server Actions;
- direct PostgreSQL, Redis, file, or secret access from Next.js;
- frontend-owned instrument catalog fixtures or business rules;
- external provider or client-side market-data calls;
- workers or scheduled collectors;
- market-data ingestion, live prices, quote history, candles, order books, dividends, or coupons;
- stock-card or bond-card financial calculations;
- fabricated price, source, sector, face value, maturity date, coupon type, yield, return, WAC,
  XIRR, real return, inflation, purchasing-power, or tax values;
- import/reconciliation changes;
- mobile implementation;
- AI functionality.

## Verification evidence

Completed locally before implementation review:

- Frontend unit tests passed:
  `cd frontend-next && corepack pnpm run test`.
- Frontend typecheck passed:
  `cd frontend-next && corepack pnpm run typecheck`.
- Frontend production build passed:
  `cd frontend-next && corepack pnpm run build`.
- Full repository verification passed:
  `GOCACHE=/private/tmp/openinvest-gocache UV_CACHE_DIR=/private/tmp/openinvest-uv-cache pnpm run verify`.
- Browser smoke passed:
  `/assets` rendered behind the existing AuthShell with no browser console errors before auth.

Independent strict review returned REQUEST CHANGES twice; fixes for detail invalidation,
transition-aware loading focus including same-ticker retries, live-region behavior, frozen detail
typing, component-contract coverage, and verification evidence were applied. Final strict
separate-window review approved the implementation before PR publication. PR #41 passed CI and was
squash-merged into `develop` at `22bede651a646d0e8b06568bda457d0626891e63`.

## Known risks

- The asset search UI can only show assets that the existing Go API exposes from active canonical
  catalog rows.
- Asset detail remains deferred while `GET /api/v1/assets/{ticker}` returns `404_NOT_FOUND`.
- Full DOM-level accessibility tests are still limited by the current no-DOM Node test setup; the
  slice uses focused helper tests for focus-entry/restoration decisions, source-level component
  wiring checks, browser smoke, strict separate-window review, and CI evidence.

## Recommended next step

Continue only with the next separately approved planning or implementation stage. Do not extend this
slice into market data, stock/bond card calculations, provider integrations, workers, tax, mobile,
or AI without a new reviewed scope.
