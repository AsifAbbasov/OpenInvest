# Stage 3.14 — Asset Search and Card API Boundary Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-14-ASSET-API-BOUNDARY-PLAN |
| Version | 0.1.4 |
| Status | Complete / closed; merged into `develop` |
| Supersedes | Informal next-step discussion after Stage 3.13 |
| Dependencies | `SOURCE_OF_TRUTH.md`; Stage 2 contract baseline; Stage 3.13 instrument catalog slice |

## Purpose

Stage 3.14 plans the smallest safe public Go API boundary for the already-frozen MVP asset search
and asset-card contract:

- `GET /api/v1/assets/search`
- `GET /api/v1/assets/{ticker}`

The planning goal is to expose backend-owned asset metadata from the approved local instrument
catalog without adding market-data ingestion, live prices, financial calculations, frontend
instrument pages, or external-provider integrations.

This document is planning only. It does not authorize implementation by itself.

## Problem


Without this boundary, the Web UI cannot later build a stock or bond card without either:

- duplicating instrument knowledge in Next.js;
- calling an external source directly from the client; or
- inventing a second asset model outside the canonical Go API.

All three options violate the Architecture Freeze. Stage 3.14 should therefore define the smallest
backend-owned API implementation slice before any frontend stock/bond-card work begins.

## Candidate backend outcome

A later implementation PR may target this local demonstration path only after this planning document

```text
Authenticated or anonymous MVP Web client
→ calls Go API asset search or asset detail endpoint
→ Go API reads only approved local catalog metadata
→ Go API returns only the frozen Stage 2 asset DTO fields that can be populated honestly
→ unsupported or inactive tickers return documented errors or empty search results
→ no market data, returns, dividends, coupons, taxes, or external provider calls occur
```

## Candidate implementation surfaces

The future implementation PR may include only:


## Explicit exclusions

Stage 3.14 planning and the future implementation slice must not add:

- OpenAPI path or schema changes unless implementation stops for a separate contract-change PR;
- SQL migrations unless implementation stops for a separate migration proposal;
- Next.js pages, components, stock cards, or bond cards;
- Next.js Route Handlers or Server Actions;
- direct database access from Next.js;
- frontend-owned instrument business logic;
- external MOEX, CBR, Rosstat, broker, or provider network calls;
- background workers or scheduled collectors;
- market-data ingestion, live prices, quote history, candles, order books, dividends, or coupons;
- stock-card or bond-card financial calculations;
- WAC, XIRR, real return, inflation, purchasing-power, or tax calculations;
- import/reconciliation changes;
- mobile implementation;
-  functionality.

## Planning decisions


## Acceptance criteria for a future implementation PR


## Review focus

Review must specifically verify:

- no market-data or provider-integration scope is introduced;
- no frontend business authority is introduced;
- no financial calculations are introduced;
- no SQL or OpenAPI changes are hidden inside the implementation;
- Stage 3.14 does not implement stock-card or bond-card UI;
- null price fields, source provenance, and required stock/bond detail fields cannot be confused
  with official market data.

## Recommended next step

Stage 3.14 implementation and closure governance are closed. Continue with Stage 3.15 Web asset
discovery UI planning from `develop` at `f5289eb604b8ba31aa422d0d09950da02e0f48b3`.
