# Stage 2 API Contract

| Field | Value |
| --- | --- |
| Document ID | API-STAGE-02 |
| Version | 1.0.0 |
| Status | Proposed / Awaiting Review |
| Owner | Principal Architect |
| Supersedes | Stage 1 operational OpenAPI skeleton |
| Dependencies | Documents 42–43; ADR-003; ADR-006 |
| Last Review Date | 2026-06-20 |
| Next Review Date | At Stage 2 approval |

## Purpose

Freeze the MVP web boundary before backend implementation. The normative machine-readable
contract is [`../../openapi/openapi.yaml`](../../openapi/openapi.yaml). This document explains
the semantics that code generators and JSON Schema alone cannot communicate safely.

## Contract boundary

- Contract format: OpenAPI 3.1 with JSON Schema 2020-12 semantics.
- Product prefix: `/api/v1`.
- Payload media type: `application/json`.
- MVP markets: MOEX stocks, MOEX bonds, and RUB cash only.
- The frontend calls OpenInvest only. MOEX, CBR, Rosstat, and every future official source
  remain behind backend collectors and never appear as client endpoints.
- The contract freezes behavior and data meaning, not Go packages, PostgreSQL tables, cache
  keys, worker messages, or provider payloads.

## Endpoint inventory

| Area | Method and path | Authentication | Purpose |
| --- | --- | --- | --- |
| Operations | `GET /api/v1/health` | Public | Process liveness |
| Operations | `GET /api/v1/ready` | Public | Required dependency readiness |
| Auth | `POST /api/v1/auth/register` | Public | Privacy-first registration |
| Auth | `POST /api/v1/auth/login` | Public | Start MVP web session |
| Auth | `POST /api/v1/auth/refresh` | Refresh cookie + CSRF | Rotate session |
| Auth | `POST /api/v1/auth/logout` | Refresh cookie + CSRF | Revoke one/all sessions |
| Assets | `GET /api/v1/assets/search` | Public | Search MOEX stocks/bonds |
| Assets | `GET /api/v1/assets/{ticker}` | Public | Stock/bond card |
| Portfolios | `GET /api/v1/portfolios` | Bearer | List portfolios |
| Portfolios | `POST /api/v1/portfolios` | Bearer + idempotency | Create RUB portfolio |
| Portfolios | `GET /api/v1/portfolios/{portfolioId}` | Bearer | Portfolio metadata |
| Portfolios | `PATCH /api/v1/portfolios/{portfolioId}` | Bearer | Rename metadata |
| Portfolios | `DELETE /api/v1/portfolios/{portfolioId}` | Bearer + idempotency | Remove from active use |
| Analytics | `GET /api/v1/portfolios/{portfolioId}/summary` | Bearer | Value, WAC, XIRR, real return, purchasing power |
| Analytics | `GET /api/v1/portfolios/{portfolioId}/snapshots` | Bearer | Versioned projections |
| Transactions | `GET /api/v1/portfolios/{portfolioId}/transactions` | Bearer | Immutable ledger history |
| Transactions | `POST /api/v1/portfolios/{portfolioId}/transactions` | Bearer + idempotency | Append transaction |
| Transactions | `PATCH /api/v1/portfolios/{portfolioId}/transactions/{transactionId}` | Bearer + idempotency | Append correction |
| Transactions | `DELETE /api/v1/portfolios/{portfolioId}/transactions/{transactionId}` | Bearer + idempotency | Append reversal |
| Dividends | `GET /api/v1/dividends/calendar` | Bearer | Official dividend events |
| Dividends | `POST /api/v1/dividends/calculate` | Public + idempotency | Gross dividend calculation |
| Dashboard | `GET /api/v1/dashboard` | Bearer | Fifteen-second capital overview |

## Response envelope

Every successful JSON response has exactly two top-level members:

```json
{
  "data": {},
  "meta": {
    "requestId": "01977a31-6db7-7ba6-b5b4-65a183ba4f41",
    "traceId": "7f4a3d2c1b0e4987a6f5e4d3c2b1a090",
    "generatedAt": "2026-06-20T08:00:00Z"
  }
}
```

Every JSON error has `error + meta`. `error.code` is stable and machine-readable;
`error.message` is safe for display; `details` may identify invalid fields. Stack traces,
SQL, provider payloads, tokens, passport data, INN, and other secrets are forbidden.

HTTP status is authoritative. Clients must not infer success from body shape alone.

## Request and trace identity

- The server returns `X-Request-ID` and `X-Trace-ID` and mirrors both in `meta`.
- A client may send a UUID in `X-Request-ID`; invalid/untrusted values are replaced.
- Distributed tracing uses the W3C `traceparent` request header.
- A trace ID correlates technical execution. It is not a user, account, or portfolio ID.
- Neither identifier may contain personal or financial data.

## Authentication freeze for MVP web

- Password minimum: 12 characters; password-manager-friendly; no forced symbol policy.
- Passwords never appear in responses, logs, traces, analytics, or audit payloads.
- Access token: short-lived JWT returned in response data and held in browser memory only.
- Refresh token: opaque rotating value in `oi_refresh`, configured `Secure`, `HttpOnly`, and
  `SameSite=Strict`. It is never returned in JSON or exposed to browser JavaScript.
- Refresh and logout require the refresh cookie plus session-bound `X-CSRF-Token`.
- Refresh-token reuse invalidates the affected rotation family.
- Registration enables Privacy Mode and leaves tax profile/notifications disabled.
- Mobile authentication is not frozen here; a future contract may add a secure native-token
  exchange without weakening or silently changing this web contract.

## Decimal, money, and dates

- JSON numbers are forbidden for financial values and rates.
- `Decimal` is a base-10 string with at most 8 fractional digits.
- Calculation precision is 8 decimal places with half-even rounding.
- UI display rounds money to 2 decimal places; the contract preserves calculation precision.
- `Money.currency` is `RUB` in MVP.
- `BusinessDate` is `YYYY-MM-DD` and maps to SQL `DATE`.
- `SystemTimestamp` is UTC RFC 3339 ending in `Z` and maps to `TIMESTAMP WITH TIME ZONE`.
- Registry, payment, and settlement dates remain separate fields.

## Cursor pagination

List endpoints use `cursor` and `limit`, never page-number/offset pagination. Responses return:

```json
{ "nextCursor": null, "hasMore": false, "limit": 20 }
```

The cursor is opaque, scoped to the authenticated query, and must not be parsed, persisted as
business data, or constructed by clients. Sort order is endpoint-defined and deterministic.
Changing cursor encoding is backward compatible while its behavior remains opaque.

## Idempotency

`Idempotency-Key` is required for portfolio creation, every ledger write, portfolio deletion,
and the financial calculator POST. The identity scope is:

```text
authenticated principal (or anonymous client scope)
+ HTTP method
+ canonical path
+ Idempotency-Key
```

The server binds the key to a canonical request hash. An identical replay returns the original
status and body. Reuse with a different payload returns `409 IDEMPOTENCY_CONFLICT`. Keys are
never accepted as financial identifiers. The implementation must retain completed keys long
enough to cover client retry windows; the concrete retention duration is an operational choice
and cannot change observable replay semantics.

## Immutable transaction semantics

The HTTP verbs satisfy familiar CRUD interaction without violating immutable history:

- `POST` appends the original transaction.
- `PATCH` appends a correction revision and returns the current projection.
- `DELETE` appends a reversal transaction; it never deletes a row.
- `expectedRevision` prevents lost updates.
- correction and reversal require an audit reason where applicable.
- historical revisions remain traceable and snapshots are invalidated/rebuilt by later stages.

Portfolio `DELETE` removes mutable portfolio metadata from active use but retains immutable
financial history. Account deletion and irreversible anonymization remain a separate privacy
workflow outside the Stage 2 endpoint list.

## Calculated responses

Portfolio summary and dashboard are server-owned read models. They expose results, source
references, business dates, and methodology versions without exposing SQL tables or provider
formats. `PortfolioSummary` includes:

- capital allocation and invested capital;
- weighted-average cost per position;
- nominal return and XIRR where mathematically defined;
- real return with the inflation input and methodology version;
- purchasing-power equivalents with registered source codes;
- calculation timestamp and input business date.

Snapshots are deterministic, rebuildable projections. They are not the source of truth.

## Asset and dividend policy

- `Asset` is a discriminated stock/bond union.
- MOEX ticker is the MVP public lookup key; internal database IDs are not leaked.
- Prices are normalized RUB cash prices per unit, not raw provider quote structures.
- Dividend calendar contains official states only: announced, approved, paid, cancelled.
- Forecast-only events are excluded because the Forecast Engine is outside MVP.
- Dividend calculator returns gross values and explicitly states `taxIncluded=false`.
- Tax export, tax advice, and foreign securities are excluded.

## Examples

Every operation has at least one success example. Request examples exist for every operation
with a JSON body. Files are grouped by stable domain rather than by endpoint:

- `openapi/examples/operations.json`
- `openapi/examples/auth.json`
- `openapi/examples/assets.json`
- `openapi/examples/portfolios.json`
- `openapi/examples/transactions.json`
- `openapi/examples/dividends.json`
- `openapi/examples/dashboard.json`
- `openapi/examples/errors.json`

Example tokens, users, IDs, and amounts are synthetic and must never be reused as secrets.

## Compatibility policy

Backward-compatible additions may stay in `/api/v1`: optional response fields, new error codes,
new enum members only where clients are required to tolerate unknown values, and new endpoints.
Breaking changes require an ADR and a new versioned contract boundary. Removing fields, changing
financial meaning, changing date semantics, or weakening idempotency is breaking.

Generated clients and server stubs remain out of Stage 2. Their future generation must use the
reviewed contract and must fail CI on uncommitted contract drift.

## Explicitly out of scope

No Stage 2 artifact implements services, repositories, SQL migrations, workers, collectors,
frontend screens, mobile code, tax export, foreign securities, predictions, or AI behavior.
