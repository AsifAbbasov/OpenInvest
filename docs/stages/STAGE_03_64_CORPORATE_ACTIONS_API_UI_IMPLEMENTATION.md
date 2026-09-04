# Stage 3.64 — Corporate Actions API / UI Implementation

| Field | Value |
| --- | --- |
| Status | Local development-path candidate; no commit/push/PR/Ready/merge authorization implied |
| Date | 2026-09-05 |
| Canonical implementation base | `develop@a8f9e95c065ee708885461166e1e992d1f4aae22` |
| Protected-base tree | `128f41ec9ed6c0e25568ffe47eb33cb0fb01b188` |
| Planning authority | `docs/stages/STAGE_03_61_CORPORATE_ACTIONS_CALENDAR_PLANNING.md` |
| Domain dependency | Stage 3.62 / Feature 3A — Corporate Action Boundary |
| Projection dependency | Stage 3.63 / Feature 3B — Calendar + Heatmap Projection |
| Feature | Feature 3C — API / UI |
| External source activation | None; shipped composition remains provider-free |

## 1. Purpose

Expose the already-canonical corporate-action boundary and Stage 3.63 projections through the smallest honest
read-only HTTP/OpenAPI/Next.js surface without activating an unapproved external source.

The implemented chain is:

```text
CorporateActionProvider
        ↓
FetchCorporateActions
        ↓
validated []CorporateActionEvent
        ↓
Calendar projection
        ↓
Heatmap projection
        ↓
GET /api/v1/corporate-actions/projection
        ↓
typed Next.js client
        ↓
Corporate Actions Calendar / Heatmap UI
```

Stage 3.64 does not create or approve a provider. Default shipped construction therefore reaches the endpoint with
no provider and returns the explicit source-unavailable state rather than fabricated or empty market data.

## 2. Candidate scope

Candidate scope is exactly the files listed in `CANDIDATE_FILE_SET.txt` for the local review bundle.

The implementation changes only:

- Go HTTP transport/composition for the corporate-action projection;
- the exact OpenAPI operation allowlist required by the repository validator;
- one provider-neutral OpenAPI schema document and the root API operation;
- the existing typed frontend client;
- one feature-folder Next.js presentation slice and focused tests;
- this implementation record.

No database, migration, worker, cache, dependency, CI workflow, external adapter, source registry activation, or
financial-calculation surface is changed.

## 3. HTTP contract

Stage 3.64 adds one public read-only endpoint:

```text
GET /api/v1/corporate-actions/projection
```

Required query:

```text
instrumentId=SBER,GAZP
from=2026-01-01
to=2026-12-31
```

Transport policy:

- `instrumentId` is an OpenAPI `style=form`, `explode=false` array;
- one request accepts at most 50 instruments;
- the canonical domain validator still owns identifier syntax, uniqueness, required dates and `from <= to`;
- instrument IDs are sorted before provider invocation so provider request order and response coverage metadata are
  deterministic;
- invalid or oversized input is rejected before provider invocation;
- supersession resolution is evaluated over the complete validated provider batch, then the public Calendar and Heatmap
  outputs are bounded to the inclusive canonical effective-date window `from <= date <= to`, so provider over-return
  cannot widen the requested public scope or bypass correction/cancellation semantics.

The successful response contains:

- `calendar`: current dated Stage 3.63 projection;
- `heatmap`: count/density-only Stage 3.63 buckets;
- `coverage`: explicit `PROVIDER` input mode and the exact normalized request scope.

## 4. Honest source state

The current Data Source Registry still has no approved broad production corporate-actions feed. Stage 3.64 therefore
keeps source availability explicit:

```text
provider absent / unavailable
→ HTTP 503
→ CORPORATE_ACTIONS_SOURCE_UNAVAILABLE
```

A successful provider response containing zero current dated events is different:

```text
validated provider response = []
→ HTTP 200
→ calendar=[]
→ heatmap=[]
```

The frontend renders these as distinct states. It never translates source unavailability into “no dividends” or
“zero events”.

`EXAMPLE_CORPORATE_ACTIONS` is used only in the OpenAPI example. Runtime Go and frontend code do not emit or depend
on the reserved example identifier.

## 5. Failure semantics

HTTP mapping is intentionally provider-neutral and fail-closed:

| Condition | HTTP | Stable code |
| --- | ---: | --- |
| invalid query / >50 instruments | 400 | `VALIDATION_ERROR` |
| provider unavailable / unclassified provider transport failure | 503 | `CORPORATE_ACTIONS_SOURCE_UNAVAILABLE` |
| provider-declared bad data | 502 | `CORPORATE_ACTIONS_SOURCE_INVALID` |
| malformed canonical event / duplicate event ID | 502 | `CORPORATE_ACTIONS_SOURCE_INVALID` |
| supersession fork/cycle/cross-instrument/cross-kind integrity failure | 502 | `CORPORATE_ACTIONS_SOURCE_INVALID` |
| valid provider result with no dated current event | 200 | empty projection |

Provider-specific error text is not exposed.

## 6. Cache / source-rights boundary

Every response from the endpoint sets:

```text
Cache-Control: no-store
```

This is deliberately conservative. Current source-governance decisions do not authorize shipped caching. A future
source adapter may relax caching only through a separately reviewed source/use-mode decision that explicitly grants
those rights.

No automatic polling, retry loop, browser persistence, Redis cache, background worker or prefetch loop is introduced.

## 7. Public DTO minimization

The public event DTO preserves the product semantics needed by Calendar/Heatmap:

- application `eventId` and canonical `instrumentId`;
- kind and lifecycle status;
- optional record/payment dates;
- optional exact-decimal amount/currency;
- optional supersession link;
- `AsOf` and `RetrievedAt`;
- canonical provider identifier.

The provider-owned opaque `SourceEventID` remains internal evidence and is intentionally **not** exposed through the
public API. This avoids freezing provider-specific identity details into the client contract before Feature 3D source
selection and public-display rights are approved.

## 8. Calendar UI semantics

The UI renders Stage 3.61 / Stage 3.63 effective-date semantics without recalculating them:

- Dividend entry whose effective date equals `RecordDate` is labeled `Record date`;
- otherwise the effective date is displayed as `Payment date`;
- Coupon effective dates therefore remain payment dates;
- undated current evidence never appears because Stage 3.63 omits it from the dated projection;
- cancelled/superseded historical evidence cannot be reintroduced by the frontend.

The UI displays lifecycle status explicitly. `ANNOUNCED` is accompanied by the statement that it is not guaranteed
income. Provider, source `AsOf`, and OpenInvest `RetrievedAt` are visible for auditability/freshness context.

## 9. Heatmap UI semantics

The first heatmap remains event-density only. It displays:

```text
Date
TotalCount
DividendCount
CouponCount
AnnouncedCount
ConfirmedCount
PaidCount
CancelledCount
```

Visual density level is derived only from `TotalCount / maximum TotalCount` in the returned buckets. The UI does not
sum amounts and performs no money, FX, yield, tax, portfolio-income, or investment-return calculation.

Numeric counts remain visible, so information is not encoded by color alone.

## 10. Frontend concurrency and state integrity

The component treats request identity as part of correctness:

- changing instruments or either date aborts the previous request and invalidates its generation;
- submitting a new request aborts the prior request and creates a new generation;
- a stale completion is ignored even if abort races with promise completion;
- `503`, legitimate empty, populated and other-error states are separate variants.

This prevents an old source-unavailable error or old calendar result from overwriting a newer query result.

No Redux, TanStack Query, global store, new dependency, or client-side financial state is introduced.

## 11. OpenAPI / validator integration

The repository OpenAPI validator maintains an exact operation allowlist and checks root path operations before normal
reference walking. Stage 3.64 therefore:

- declares the new operation inline in `openapi/openapi.yaml` rather than hiding the path item behind an external
  `$ref`;
- adds `GET /api/v1/corporate-actions/projection -> getCorporateActionProjection` to the validator allowlist;
- adds `getCorporateActionProjection` to the validator's public-operation set;
- uses a specialized response schema that inherits canonical `BaseResponse` through `$ref`;
- keeps the detailed corporate-action data schemas in `openapi/components/corporate-actions.yaml`.

This is necessary for the repository's exact API-contract gate and is not a generic validator relaxation.

## 12. Internal review evidence publication state

Internal review is required before publication authorization, but its current verdict/findings are intentionally:

```text
WITHHELD — external published-head phase pending
```

This implementation record does not publish or summarize the Internal findings before the fresh External verdict.
The complete local Internal review report is retained outside the repository candidate. After External review, the
required Internal evidence may be published in a documentation/evidence-only follow-up and independently verified
for exactness and absence of semantic drift.

## 13. Local deterministic preflight

The execution sandbox cannot clone the repository and has Go `1.23.2`, while canonical `backend-go/go.mod` requires
Go `1.25.14`. This is the same class of prepublication environment limitation recorded by Stage 3.62. Therefore
repository-wide Go/Next/OpenAPI gates are not claimed locally when they cannot be executed against the exact checkout.

Executed local evidence on the frozen candidate source surface:

- `gofmt` on new Go production/test files: PASS;
- TypeScript/TSX parser preflight via installed TypeScript compiler: PASS for model, component, page and component test;
- pure corporate-action frontend model tests: PASS, 3/3;
- YAML parse for the new OpenAPI component document: PASS;
- candidate-local OpenAPI reference-target audit: PASS;
- cross-layer route/operation/error/status/heatmap/max-instruments/source-state parity audit: PASS, 22/22;
- forbidden-surface scan: PASS — no `float64`, direct `time.Now()`, SQL/pgx, migration, Redis, Kafka,
  Interfax/NSD/MOEX adapter wiring, polling, or runtime `EXAMPLE_CORPORATE_ACTIONS` in candidate production code.

Not yet authoritative before publication:

- `go test ./...`: UNKNOWN / requires exact repository toolchain checkout;
- `go test -race ./...`: UNKNOWN / requires exact repository toolchain checkout;
- `go vet ./...`: UNKNOWN / requires exact repository toolchain checkout;
- `go run ./cmd/validate-openapi`: UNKNOWN / requires exact repository checkout and Go 1.25.14;
- frontend full `typecheck`, `test`, `build`: UNKNOWN / requires exact repository checkout/dependencies;
- all ten GitHub required checks: UNKNOWN until separately authorized publication.

UNKNOWN is not converted to PASS.

## 14. Architectural consequences

After Stage 3.64 merges, the provider-neutral corporate-actions product surface is complete:

```text
canonical event boundary   — Stage 3.62
projection semantics       — Stage 3.63
HTTP/OpenAPI/UI surface    — Stage 3.64
```

This closes the Corporate Actions architecture/API/UI implementation debt. It does **not** close the external source
blocker. Feature 3D remains separately gated by exact source/use rights, licensing/cost acceptance, rate/traffic
policy, caching/retention/public-display rights, fresh Data Source Registry approval, and separately reviewed runtime
composition.

Default shipped behavior after Stage 3.64 remains honest and fail-closed: the UI can explain source unavailability,
but it cannot claim live/all-market dividend or coupon coverage.

## 15. Explicit exclusions

No:

- Interfax, NSD, CBR, MOEX, issuer adapter or scraping;
- real provider HTTP activation;
- SQL, migration, persistence or replay;
- worker, polling, retry framework or cache;
- dependency change;
- monetary heatmap;
- yield, FX, tax, portfolio forecast, notification, amortization or redemption;
- Feature 3D implementation;
- changes to Stage 3.62/3.63 domain semantics.

## 16. Governance state

This remains a local development-path candidate. The designated review chat must finish the complete Internal
read-only review over the frozen candidate. If and only if the final Internal verdict is `APPROVED`, commit/push/Draft
PR require a separate explicit Principal Architect authorization bound to the exact frozen manifest and canonical
base.

After publication: exact-head GitHub CI → fresh External published-head review → remediation if demonstrated →
evidence-only publication → CI → exact evidence verification → separate human Ready/squash-merge authorization.
