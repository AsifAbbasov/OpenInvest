# Stage 3.14 — Asset Search and Card API Boundary Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-14-ASSET-API-BOUNDARY-PLAN |
| Version | 0.1.0 |
| Status | Draft / planning |
| Owner | Builder Engineer |
| Supersedes | Informal next-step discussion after Stage 3.13 |
| Dependencies | `SOURCE_OF_TRUTH.md`; Stage 2 contract baseline; Stage 3.13 instrument catalog slice |
| Last Review Date | 2026-07-13 |
| Next Review Date | Before Stage 3.14 implementation |

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

Stage 3.13 made transaction and import flows resolve approved MOEX share/bond tickers through the
backend-owned catalog boundary. The product still has an MVP contract gap: the public asset search
and asset-card endpoints exist in the Stage 2 OpenAPI contract, but the current implementation does
not yet expose a reviewed Go API surface for users to discover supported MVP instruments.

Without this boundary, the Web UI cannot later build a stock or bond card without either:

- duplicating instrument knowledge in Next.js;
- calling an external source directly from the client; or
- inventing a second asset model outside the canonical Go API.

All three options violate the Architecture Freeze. Stage 3.14 should therefore define the smallest
backend-owned API implementation slice before any frontend stock/bond-card work begins.

## Candidate backend outcome

A later implementation PR may target this local demonstration path only after this planning document
is reviewed and merged:

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

- Go HTTP handlers for the already-frozen asset search/detail endpoints;
- service/store read methods over the existing `investment.assets` table;
- mapping from backend-owned catalog rows to Stage 2 asset DTOs only where every required response
  field has reviewed, non-fabricated data;
- search responses that use `lastPrice: null` when no approved market-data source exists;
- detail responses only after the implementation PR proves how every required `Asset` field is
  populated without fake provenance or contract drift;
- unsupported ticker and inactive asset handling;
- backend and API tests for search, detail, case/canonical ticker behavior, and privacy boundaries;
- documentation updates.

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
- AI functionality.

## Planning decisions

- Go remains the only canonical business API.
- Next.js remains presentation only under ADR-007.
- The first asset search/card API implementation should use only the approved local fixture catalog
  introduced by Stage 3.13.
- The implementation must not broaden supported instruments beyond the reviewed fixture set without
  a separate source-governance and catalog-expansion review.
- Price placeholders are forbidden. Until an approved market-data source exists, asset search must
  return `lastPrice: null`; asset detail must return `lastPrice: null` and `priceAsOf: null` if it
  is implemented at all.
- A runtime `source` response cannot use reserved `EXAMPLE_*` identifiers, fabricated providers, or
  unregistered provenance. If the required Stage 2 `source` field cannot be populated from an
  approved registry entry, implementation must stop for a separate source-governance or
  contract-change decision before serving the detail endpoint.
- Required stock and bond detail fields cannot be invented. Sector, ISIN, face value, maturity date,
  and coupon type may be returned only when backed by reviewed static fixture metadata for the same
  approved instruments. Bond identity metadata is distinct from dividend/coupon events,
  calculations, calendars, and yield analytics, which remain out of scope.
- External provider ingestion is a separate future stage requiring Data Source Registry updates,
  caching policy, freshness policy, auditability, and failure-mode design.
- Frontend stock/bond cards are a later stage after the Go API boundary is implemented and reviewed.

## Acceptance criteria for a future implementation PR

- Existing Stage 2 asset search/detail contract is served by Go without OpenAPI drift.
- Search returns only supported active assets from the backend-owned catalog.
- Detail lookup returns a supported active asset only if every mandatory field has reviewed,
  non-fabricated data and a valid runtime source; otherwise implementation must defer detail lookup
  or stop for a separate contract/source-governance proposal.
- Response mapping preserves Stage 2 `STOCK | BOND` discriminated asset semantics.
- `lastPrice` and `priceAsOf` are `null` until an approved market-data source exists.
- Runtime `source` never uses `EXAMPLE_*`, fake provider identifiers, or unregistered provenance.
- Required stock/bond detail values are fixture-backed static identity metadata, not live market
  data, coupon events, coupon calculations, or analytics.
- Tests prove no unsupported ticker becomes an accepted asset reference.
- Tests prove no external provider, worker, frontend, SQL migration, or business-calculation scope
  entered the slice.
- CI is green.
- Independent strict review confirms scope remains limited to the approved Go API boundary.

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

After this planning PR is reviewed and merged, start a separate Stage 3.14 implementation feature
branch for the Go API asset search/detail boundary only.
