# Stage 3.13 — Instrument Catalog Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-13-INSTRUMENT-CATALOG-PLAN |
| Version | 0.1.4 |
| Status | Complete / closed; merged into `develop` at `ca16af9adba249fc8c32c9b246b5f92f7e290b92` |
| Owner | Builder Engineer |
| Supersedes | Informal next-step discussion after Stage 3.12 |
| Dependencies | `SOURCE_OF_TRUTH.md`; Stage 2 contract baseline; Stage 3.12 Web authentication UI slice |
| Last Review Date | 2026-07-13 |
| Next Review Date | Before Stage 3.14 implementation |

## Purpose

Stage 3.13 plans the smallest safe backend-owned instrument-catalog usage boundary needed before
OpenInvest can validate transaction asset references against the frozen Stage 2 MVP asset model.

The planning goal is to align the next implementation slice with the already-approved Stage 2
model: MOEX ticker is the public lookup key, the Investment context owns the canonical security
master, and `STOCK | BOND` is the MVP asset-type union. This plan must not reopen those
architecture decisions or introduce external-provider integrations, financial calculations,
workers, tax logic, or frontend business authority.

This document is planning only. It does not authorize implementation by itself, and approval of
this document does not waive the normal migration, OpenAPI, backend, frontend, review, or CI gates
for any later implementation PR.

## Problem

Current portfolio and import flows can store transactions, and the Stage 3.01 database foundation
already created the approved `investment.assets` table. The remaining gap is not a missing
architecture boundary; it is that current transaction/import behavior still treats tickers as
user-supplied text instead of resolving supported tickers through the backend-owned asset catalog.

Before stock cards, bond cards, dividend views, or market-data ingestion are implemented, the system
needs a narrower implementation slice that respects the frozen canonical boundary for:

- supported MOEX ticker lookup and internal asset linkage;
- Stage 2 `STOCK | BOND` classification;
- display names and currency constraints;
- user-facing validation behavior;
- future provider-ingestion boundaries.

## Candidate backend outcome

A separately reviewed implementation PR may target this local demonstration only after this plan is
merged and implementation is explicitly approved:

```text
Authenticated user opens portfolio detail
→ enters a supported MOEX ticker for a share or bond transaction
→ Go API validates the ticker against the Stage 2 asset model and approved local fixture catalog
→ Go API stores or returns only backend-owned asset references/metadata already allowed by the
  approved contract
→ no stock-card, bond-card, dividend, coupon, price, return, or frontend feature is introduced
```

## Candidate implementation surfaces

The following are candidate surfaces for a later Stage 3.13 implementation PR. They are not
authorized by this planning document, and each changed surface must pass its own established gate:

- a minimal backend-owned projection of the frozen Stage 2 asset model for MVP MOEX shares and
  bonds;
- use of the existing Stage 3.01 `investment.assets` table and indexes, without adding duplicate
  asset-identity schema;
- Go repository/service validation behind existing approved API behavior to resolve transaction
  tickers to backend-owned asset entries;
- deterministic local seed or fixture data for a narrow reviewed MVP instrument set;
- no OpenAPI request/response changes unless implementation first stops for the separate
  contract-change proposal required by the Stage 3 API-first rule;
- no SQL migrations unless implementation first stops for a separate migration/contract-impact
  review compatible with the frozen Stage 2 ER model;
- tests for validation, privacy, idempotency compatibility, and asset-catalog lookup behavior;
- documentation updates.

## Forbidden scope

Stage 3.13 must not add:

- external MOEX, CBR, Rosstat, broker, or provider network integrations;
- background workers or scheduled collectors;
- live prices, quotes, candles, order books, or market-data ingestion;
- dividend or coupon calendars;
- stock-card or bond-card pages, cards, or frontend presentation work;
- stock-card or bond-card financial calculations;
- WAC, XIRR, real return, inflation, purchasing-power, or tax calculations;
- broker synchronization or credential scraping;
- AI functionality;
- mobile implementation;
- Next.js Route Handlers or Server Actions for business domains;
- direct database access from Next.js;
- frontend-owned instrument business logic.

## Planning decisions

- Go remains the only canonical business API and owns asset validation.
- Next.js remains presentation only under ADR-007.
- The first catalog should be intentionally small and deterministic; broad market coverage is not
  required for the first implementation PR.
- User-supplied ticker text should not become an accepted asset reference without backend
  validation against the approved local catalog.
- Stage 3.13 must use the Stage 2 asset identity rules: MOEX ticker is the public lookup key,
  internal database IDs are not leaked, and `STOCK | BOND` is the discriminated MVP asset union.
- External provider ingestion requires a later planning stage with source registry updates,
  retention rules, audit boundaries, and failure-mode design.
- Instrument catalog implementation must not silently broaden MVP scope into market-data products.

## Decisions documented for implementation

- Minimal local fixtures are limited to `SBER`, `GAZP`, and `SU26238RMFS4` for the first
  implementation slice.
- Transaction append rejects unknown or noncanonical tickers immediately with the existing invalid
  input path; it does not preserve unresolved asset references.
- No OpenAPI DTO changes are authorized for the first implementation slice.
- Historical transactions keep referencing the stable backend asset identity. Later display-name
  lifecycle changes require a separately reviewed metadata-management slice.
- Rejected instrument references do not add a new audit event in this slice; existing request/error
  telemetry remains the only evidence until a separately reviewed rejection-audit design exists.

## Acceptance criteria for a future implementation PR

- Asset identity and type semantics follow the frozen Stage 2 model and are backend-owned.
- Unknown or unsupported tickers have explicit behavior and tests.
- No frontend presentation or business logic is added in the Stage 3.13 implementation slice unless
  a later explicit planning gate changes that scope.
- No external provider calls or workers are added.
- No financial calculations are introduced.
- Backend and persistence tests cover the new boundary; OpenAPI, migration, handler, or frontend
  changes require their separate established gate before implementation continues.
- If OpenAPI or migration changes are needed, implementation stops for the separate proposal
  required by Stage 3 governance before code changes continue.
- CI is green.
- Independent review confirms Stage 3.13 stays within the approved instrument-catalog scope.

## Review focus

Review must specifically verify:

- no provider integration or market-data ingestion entered the planning or implementation scope;
- no frontend business authority was introduced;
- no financial calculations were authorized;
- migrations, OpenAPI changes, and Go handler changes remain unauthorized until the required
  separate proposal/review gate explicitly approves that surface;
- the plan does not authorize stock-card, bond-card, frontend presentation, dividend, coupon, tax,
  mobile, or AI scope.

## Recommended next step

Stage 3.13 planning, implementation, and closure governance are closed. Continue with Stage 3.14
asset search/card API boundary planning.
