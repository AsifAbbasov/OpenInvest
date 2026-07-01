# Stage 3.3 — Next.js Presentation Slice

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-03 |
| Version | 0.1.0 |
| Status | Proposed / pending review |
| Owner | Builder Engineer |
| Supersedes | None |
| Dependencies | ADR-007; Stage 3 plan; Stage 3.2 Go API vertical slice |
| Last Review Date | 2026-06-27 |
| Next Review Date | Before Stage 3.3 merge |

## Purpose

Render the first Web path through the canonical Go API only.

The slice proves:

- Next.js can render the dashboard and portfolio path;
- portfolio list/detail data comes from Go API responses;
- transaction append is submitted to the Go API with an idempotency key;
- summary and transaction history are reloaded from backend state after mutation;
- no business logic, financial calculation, provider integration, database access, or Route Handler
  enters the Web layer.

## Scope

Allowed in this stage:

- App Router pages for dashboard and portfolio detail;
- typed DTO-only Go API client;
- presentation components for portfolio list, summary, transaction history, create portfolio, and
  add transaction form;
- loading, empty, and expected-error states;
- CSS-only responsive layout;
- documentation and governance synchronization.

Forbidden in this stage:

- business or domain logic in Next.js;
- Server Actions with business behavior;
- Route Handlers for business domains;
- direct PostgreSQL, Redis, MOEX, CBR, Rosstat, broker, or external-provider access;
- LocalStorage, SessionStorage, IndexedDB, or cookie persistence of business data;
- authentication implementation;
- financial/tax/dividend/inflation calculations in Web;
- mobile implementation;
- Stage 3.4+ work.

## API boundary

```text
Browser / Next.js presentation
        ↓ HTTP / OpenAPI contract
Go Fiber API
        ↓
PostgreSQL canonical ledger and snapshots
```

`NEXT_PUBLIC_OPENINVEST_API_BASE_URL` may point the browser to the Go API. It is not secret. The
default local value is `http://localhost:8080`.

## Review focus

- Verify no `src/app/**/route.ts` files exist.
- Verify no `"use server"` Server Actions exist.
- Verify no direct database/provider packages are installed.
- Verify all financial values are displayed from decimal strings returned by the Go API.
- Verify any formatting is presentation-only and does not calculate portfolio values.
- Verify transaction submission uses the OpenAPI payload and `Idempotency-Key`.

## Remaining risks

- Authentication is intentionally not implemented in Stage 3.3; the Go API still uses the Stage 3.2
  development subject boundary.
- Browser-to-Go local development is enabled only for explicit local Web origins through the Go API
  CORS boundary. Production CORS policy remains a later deployment hardening decision.
- This slice is not an end-to-end release proof; Stage 3.4 remains responsible for full E2E
  verification and onboarding updates.
