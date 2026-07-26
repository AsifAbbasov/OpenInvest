# Stage 3.15 — Web Asset Discovery UI Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-15-WEB-ASSET-DISCOVERY-UI-PLAN |
| Version | 0.1.4 |
| Status | Complete / closed |
| Owner | Builder Engineer |
| Supersedes | Informal next-step discussion after Stage 3.14 |
| Dependencies | `SOURCE_OF_TRUTH.md`; ADR-007; Stage 2 contract baseline; Stage 3.14 asset API boundary slice |
| Last Review Date | 2026-07-27 |
| Next Review Date | Superseded by Stage 3.15 implementation closure |

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
is reviewed, approved, and merged:

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
- AI functionality.

## Planning decisions

- Go remains the only canonical asset API and catalog authority.
- Next.js remains presentation only under ADR-007.
- The future UI must call the existing Go API directly through the typed frontend API client.
- Asset search and detail calls must use public-request semantics: `credentials: "omit"` is
  required, the request must not include a bearer authorization header or CSRF header, and the
  request path must not read browser storage. Leaving Fetch credentials unset is not acceptable,
  because the default `same-origin` mode may send cookies.
- Search results may display `lastPrice` only as unavailable/null until an approved market-data
  source exists and the Go API returns real values.
- Asset detail/card UI must remain an honest deferred state while the Go detail endpoint returns
  `404 NOT_FOUND`.
- Detail `404 NOT_FOUND` handling must not claim a specific backend cause. The presentation may say
  that detail is unavailable for the selected asset, but it must not assert that every 404 proves
  provenance deferral instead of ordinary not-found behavior.
- The frontend must not embed the approved fixture set, infer stock/bond detail fields, or branch on
  provider/source facts that the Go API did not return.
- Search state must be keyed by the exact query text, asset type, and cursor chain used for the
  request. Changing query text or asset type resets the current cursor, accumulated results,
  selected asset, and pending request generation before a new search can commit.
- Older responses must never replace newer visible results. The future implementation should use an
  abort signal, request generation token, or equivalent stale-response guard for both initial search
  and pagination requests.
- Pagination may append results only when the response belongs to the same active query/type chain
  and was requested with the currently accepted cursor.
- The asset discovery UI must define and implement a testable keyboard model: tab order reaches the
  search input, type filter, result actions, pagination action, and deferred-detail control in DOM
  order; Enter or Space activates focused result and pagination actions; Escape from a selected or
  deferred-detail state restores focus to the selected result or the search input when no result is
  selected.
- The asset discovery UI must define and implement a testable focus contract: selecting a result
  moves focus to the deferred-detail heading or region with an accessible name; closing or replacing
  deferred detail restores focus to the originating result when it still exists, otherwise to the
  search input; query/type changes that clear selection must not leave focus on removed content.
- The asset discovery UI must define and implement a testable status-announcement contract using a
  polite live region for loading, empty results, and result-count changes, plus an assertive
  alert/status for API errors. Tests must assert the observable announced text, not just the
  existence of ARIA attributes.
- Stage 3.15 should be small enough to review as a presentation-only slice.

## Acceptance criteria for a future implementation PR

- The UI calls only the existing Go asset search endpoint for search results.
- Asset search and detail API calls explicitly set `credentials: "omit"` and omit bearer
  authorization, cookies, CSRF headers, and browser storage usage.
- The UI does not call external providers or read local fixture data.
- Search supports query, type, pagination, loading, empty, and error states without hiding API
  failures.
- Query or asset-type changes reset cursor state, accumulated results, selected asset state, and
  pending request generation before committing new results.
- Pagination appends only same-query/same-type responses for the accepted cursor chain.
- Stale initial-search and pagination responses cannot overwrite or append to newer state.
- Null `lastPrice` is rendered as unavailable, not zero or stale.
- Asset detail/card entry renders a deferred/unavailable state when the Go API returns
  `404 NOT_FOUND`, without claiming a specific backend cause for every 404.
- Keyboard navigation follows the documented model for tab order, Enter/Space activation, Escape
  recovery, and pagination/result actions.
- Focus behavior follows the documented contract for selection, deferred-detail entry, detail
  replacement/closure, and query/type changes that remove the selected result.
- Screen-reader status announcements use the documented polite live-region and assertive error
  contract for loading, empty results, errors, result-count changes, selection, and deferred-detail
  states.
- No Route Handlers, Server Actions, datastore access, OpenAPI changes, SQL migrations, Go handler
  changes, provider integrations, workers, market-data ingestion, or financial calculations enter
  the slice.
- Tests cover frontend API request construction with `credentials: "omit"` and without
  Authorization, CSRF, or browser-storage reads; pagination cursor reset and append rules;
  stale-response protection; unavailable-price rendering; deferred-detail rendering; keyboard model
  behavior; focus destination and restoration; and live-region/error announcements.
- CI is green.
- Independent strict review confirms scope remains presentation-only.

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
`dfeab109b2825fe0e0317e87a7abf2e706a29ea6`. The reviewed Web asset discovery UI implementation
slice is also closed and merged into `develop` at
`22bede651a646d0e8b06568bda457d0626891e63`. Continue only with the next separately approved
planning or implementation stage.
