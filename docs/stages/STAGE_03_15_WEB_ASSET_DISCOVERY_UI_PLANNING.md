# Stage 3.15 — Web Asset Discovery UI Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-15-WEB-ASSET-DISCOVERY-UI-PLAN |
| Version | 0.1.4 |
| Status | Complete / closed |
| Supersedes | Informal next-step discussion after Stage 3.14 |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-007; Stage 2 contract baseline; Stage 3.14 asset API boundary slice |

## Purpose

Stage 3.15 plans the smallest safe Web presentation boundary for asset discovery over the already
implemented Stage 3.14 Go asset search API.

The goal is to let the future Web UI expose supported MVP instruments without moving asset
business authority, catalog knowledge, market data, or financial calculations into Next.js.

This document is planning only. It does not authorize implementation by itself.

## Problem

Stage 3.14 exposed a Go API asset search boundary over active canonical local catalog rows. The Web
presentation layer still has no reviewed user-facing entry point for discovering supported assets.

Without a separate Web planning gate, a future UI could accidentally:

- duplicate approved fixture knowledge in frontend code;
- infer unavailable stock-card or bond-card facts from sparse search summaries;
- fabricate prices, source provenance, sectors, face values, maturities, coupon types, dividends,
  coupons, yields, returns, or purchasing-power data;
- add Route Handlers, Server Actions, or direct datastore access that bypass the Go API boundary.

Stage 3.15 should therefore plan only the presentation path before any Next.js implementation.

## Candidate user outcome

A later implementation PR may target this local demonstration path only after this planning document

```text
Authenticated Web user
→ opens an asset discovery entry in the existing presentation shell
→ types a ticker or name query
→ Next.js calls the Go API asset search endpoint directly
→ Web renders supported asset summaries with lastPrice shown as unavailable
→ selecting an asset shows an honest deferred detail/card state
→ no market data, calculations, provider calls, or frontend-owned catalog logic occur
```

## Candidate implementation surfaces

The future implementation PR may include only:

- Next.js presentation components for asset search input, results, empty state, loading state, and
  error state;
- typed frontend API-client methods that call the existing Go API asset search endpoint directly;
- public asset API requests that explicitly use `credentials: "omit"` and omit bearer tokens,
  cookies, CSRF headers, and browser storage access because the Stage 3.14 asset endpoints are
  public;
- presentation copy that makes unavailable prices and deferred asset detail clear without implying
  live market data;
- a deferred asset-card state for `GET /api/v1/assets/{ticker}` returning `404 NOT_FOUND`;
- route or navigation wiring inside the existing ADR-007 Web presentation boundary;
- tests for frontend API calls, state transitions, stale-response handling, credential minimization,
  keyboard/focus/status accessibility, and honest unavailable-data rendering;
- documentation updates.

## Explicit exclusions

Stage 3.15 planning and the future implementation slice must not add:

- implementation code in this planning PR;
- OpenAPI path or schema changes;
- SQL migrations;
- Go handler, service, or store changes;
- Next.js Route Handlers or Server Actions;
- direct PostgreSQL, Redis, file, or secret access from Next.js;
- frontend-owned instrument catalog fixtures or business rules;
- client-side external MOEX, CBR, Rosstat, broker, issuer, or provider calls;
- background workers or scheduled collectors;
- market-data ingestion, live prices, quote history, candles, order books, dividends, or coupons;
- stock-card or bond-card financial calculations;
- fabricated price, sector, source, face value, maturity date, coupon type, yield, return, WAC,
  XIRR, real return, inflation, purchasing-power, or tax values;
- import/reconciliation changes;
- mobile implementation;
-  functionality.

## Planning decisions


## Acceptance criteria for a future implementation PR


## Review focus

Review must specifically verify:

- ADR-007 compliance;
- no frontend-owned catalog or business authority;
- no market-data, provider, worker, or financial-calculation scope;
- no fabricated stock-card or bond-card facts;
- no Go/OpenAPI/SQL changes hidden in a Web UI PR;
- public asset calls explicitly set `credentials: "omit"` and omit auth credentials, cookies, CSRF,
  and browser-storage reads;
- search pagination and stale-response invariants are explicitly tested;
- honest UI treatment of null prices and deferred asset detail;
- keyboard, focus, and assistive-technology status behavior have concrete, testable outcomes.

## Recommended next step

Stage 3.15 planning is closed and merged into `develop` at
Canonical record: commit(s) `dfeab109b2825fe0e0317e87a7abf2e706a29ea6`.
slice is also closed and merged into `develop` at
`22bede651a646d0e8b06568bda457d0626891e63`. Continue only with the next separately approved
planning or implementation stage.
