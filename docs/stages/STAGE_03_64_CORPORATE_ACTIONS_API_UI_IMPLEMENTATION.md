# Stage 3.64 — Corporate Actions API / UI Implementation

Canonical record: PR #128; commit(s) `a8f9e95c065ee708885461166e1e992d1f4aae22`, `128f41ec9ed6c0e25568ffe47eb33cb0fb01b188`, `77b62529f9d4ca10965191fdd6a6883567fc7492`, `9ba20e976d544a6f396444d9c5eec72f7c235d45`, `9bbcf6d3f0f4a3b87e06a944869fb6e4ef722784`, `1d24e2cd02b216b4caedb29089cd86dc97af26f9`.

## 1. Purpose

Expose the canonical Corporate Action boundary and Stage 3.63 projections through the smallest honest read-only
HTTP/OpenAPI/Next.js surface without activating an unapproved external source.

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

## 2. Published scope

The initial frozen V3 candidate contained 16 files. Post-publication remediation added the frontend test-runner
configuration and CSS-module test loader and revised the Corporate Actions component test, leaving the current PR at
18 changed files without expanding the product responsibility.

The complete current PR surface is limited to:

- Go HTTP transport/composition for the corporate-action projection;
- the exact OpenAPI operation allowlist required by the repository validator;
- one provider-neutral OpenAPI schema document and the root API operation;
- the existing typed frontend client;
- one feature-folder Next.js presentation slice and focused tests;
- frontend-only component-test infrastructure required to execute the new styled component tests under Node;
- this implementation/evidence record.

No database, migration, worker, cache, production dependency, CI workflow, external adapter, source-registry activation,
or financial-calculation surface is changed.

limit, and hand-written runtime/business logic remains below the 800-line budget. Tests, documentation, CSS and the
OpenAPI specification are still fully reviewed but are not disguised as hand-written business logic.

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
- the canonical domain validator owns identifier syntax, uniqueness, required dates and `from <= to`;
- instrument IDs are sorted before provider invocation so provider request order and coverage metadata are deterministic;
- invalid or oversized input is rejected before provider invocation;
- supersession resolution is evaluated over the complete validated provider batch;
- only after supersession resolution, public Calendar and Heatmap outputs are bounded to the inclusive canonical
  effective-date window `from <= date <= to`, so provider over-return cannot widen public scope or bypass
  correction/cancellation semantics.

The successful response contains:

- `calendar`: current dated Stage 3.63 projection;
- `heatmap`: count/density-only Stage 3.63 buckets;
- `coverage`: explicit `PROVIDER` input mode and the normalized request scope.

## 4. Honest source state

The Data Source Registry still has no approved broad production corporate-actions feed. Stage 3.64 therefore keeps
source availability explicit:

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

HTTP mapping is provider-neutral and fail-closed:

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


No automatic polling, retry loop, browser persistence, Redis cache, background worker or prefetch loop is introduced.

## 7. Public DTO minimization

The public event DTO preserves only product semantics needed by Calendar/Heatmap:

- application `eventId` and canonical `instrumentId`;
- kind and lifecycle status;
- optional record/payment dates;
- optional exact-decimal amount/currency;
- optional supersession link;
- `AsOf` and `RetrievedAt`;
- canonical provider identifier.

The provider-owned opaque `SourceEventID` remains internal evidence and is intentionally not exposed through the
public API. This avoids freezing provider-specific identity details into the client contract before Feature 3D source
selection and public-display rights are approved.

## 8. Calendar UI semantics

The UI renders Stage 3.61 / Stage 3.63 effective-date semantics without recalculating them:

- Dividend entry whose effective date equals `RecordDate` is labeled `Record date`;
- otherwise the effective date is displayed as `Payment date`;
- Coupon effective dates remain payment dates;
- undated current evidence never appears because Stage 3.63 omits it from the dated projection;
- cancelled/superseded historical evidence cannot be reintroduced by the frontend.

Lifecycle status remains explicit. `ANNOUNCED` is accompanied by the statement that it is not guaranteed income.
Provider, source `AsOf`, and OpenInvest `RetrievedAt` are visible for auditability/freshness context.

## 9. Heatmap UI semantics

The first heatmap remains event-density only:

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

Visual density is derived only from `TotalCount / maximum TotalCount` in the returned buckets. The UI performs no
money, FX, yield, tax, portfolio-income, or investment-return aggregation. Numeric counts remain visible, so
information is not encoded by color alone.

## 10. Frontend concurrency and state integrity

The component treats request identity as correctness state:

- changing instruments or either date aborts the previous request and invalidates its generation;
- submitting a new request aborts the prior request and creates a new generation;
- a stale completion is ignored even if abort races with promise completion;
- `503`, legitimate empty, populated and other-error states are distinct variants.

This prevents an old source-unavailable error or old calendar result from overwriting a newer query result.

No Redux, TanStack Query, global store, new production dependency, or client-side financial state is introduced.

cancellation before a real rate/cost-limited provider is activated. Query-change/resubmit stale-result correctness is
already protected and tested, so this note is not a Stage 3.64 merge blocker.

## 11. OpenAPI / validator integration

The repository OpenAPI validator maintains an exact operation allowlist and checks root path operations before normal
reference walking. Stage 3.64 therefore:

- declares the new operation inline in `openapi/openapi.yaml` rather than hiding the path item behind an external `$ref`;
- adds `GET /api/v1/corporate-actions/projection -> getCorporateActionProjection` to the validator allowlist;
- adds `getCorporateActionProjection` to the validator public-operation set;
- uses a specialized response schema that inherits canonical `BaseResponse` through `$ref`;
- keeps detailed corporate-action data schemas in `openapi/components/corporate-actions.yaml`.

This is necessary for the repository exact API-contract gate and is not a generic validator relaxation.

## 13. Verification and remediation chronology

### Prepublication deterministic evidence

The execution sandbox could not run the exact repository Go toolchain, so unavailable checks were left UNKNOWN rather
than converted to PASS. Available prepublication checks passed:

- `gofmt` on new Go production/test files;
- TypeScript/TSX parser preflight;
- pure Corporate Actions frontend model tests, 3/3;
- YAML parse and candidate-local OpenAPI reference-target audit;
- cross-layer route/operation/error/status/heatmap/max-instruments/source-state parity, 22/22;
- forbidden-surface scan for `float64`, direct `time.Now()`, SQL/pgx, migration, Redis, Kafka, external source adapter
  wiring, polling, and runtime `EXAMPLE_CORPORATE_ACTIONS`.

### Initial publication

The exact 16-file V3 candidate was published as one semantic commit:

```text
77b62529f9d4ca10965191fdd6a6883567fc7492

tree:
9ba20e976d544a6f396444d9c5eec72f7c235d45
```

Draft PR #128 targets the unchanged canonical base `develop@a8f9e95c065ee708885461166e1e992d1f4aae22`.

### CI #341 — demonstrated frontend test-runner defect

GitHub Actions run `33926915501` / CI #341 finished with 9 of 10 required jobs passing. Go tests, Go race tests,
Go vet, Go vulnerability scan, Python tests, OpenAPI contract, Docker Compose config, PostgreSQL migration validation,
and dependency security scan passed. Frontend typecheck passed, but the frontend Test step failed because Node/tsx could
not import `CorporateActionsSlice.module.css` from the new component test.

Remediation commit:

```text
cee7e1aab8b02d98e14148939773e89c11ec8bcb
```

changed only `frontend-next/package.json` and added `frontend-next/tests/css-module-loader.mjs`. It introduced no
production dependency or production runtime change.

### Fresh review hardening — React/JSDOM import ordering

and found that `react-dom/client` had been imported before browser globals were installed. Because React DOM performs
change-event feature detection at import time, the test environment was stabilized before relying on the next CI run.

Remediation commit:

```text
78e7aaade9ea7de97fbc6f91093993b88e35ef37
```

changed only `frontend-next/tests/corporate-actions-slice.component.test.tsx`: a stable JSDOM is installed before
dynamic React/ReactDOM/component imports and tests are explicitly sequential.

### CI #343 — demonstrated missing browser `self`

GitHub Actions run `33927419572` / CI #343 showed typecheck and all pre-existing frontend tests passing, but the five
new Corporate Actions component tests failed with `ReferenceError: self is not defined` from Next.js
`requestIdleCallback/useIntersection` used by `Link`.

Remediation commit:

```text
9bbcf6d3f0f4a3b87e06a944869fb6e4ef722784

tree:
1d24e2cd02b216b4caedb29089cd86dc97af26f9
```

added only `self: dom.window` to the test browser-global setup.

### Final pre-evidence exact-head CI #344

GitHub Actions run `33927609434` / CI #344 completed successfully on exact head
`9bbcf6d3f0f4a3b87e06a944869fb6e4ef722784`.

All ten required jobs passed:

```text
Go tests                         PASS
Go race tests                    PASS
Go vet                           PASS
Go vulnerability scan            PASS
Python tests                     PASS
Frontend build and typecheck     PASS
OpenAPI contract                 PASS
Docker Compose config            PASS
PostgreSQL migration validation  PASS
Dependency security scan         PASS
```

The frontend job specifically passed all three required steps: Typecheck, Test, and Build.

## 15. Architectural consequences

After Stage 3.64 merges, the provider-neutral Corporate Actions product surface is complete:

```text
canonical event boundary   — Stage 3.62
projection semantics       — Stage 3.63
HTTP/OpenAPI/UI surface    — Stage 3.64
```

This closes the Corporate Actions architecture/API/UI implementation debt. It does not close the external source
blocker. Feature 3D remains separately gated by exact source/use rights, licensing/cost acceptance, rate/traffic
composition.

Default shipped behavior after Stage 3.64 remains honest and fail-closed: the UI can explain source unavailability,
but it cannot claim live/all-market dividend or coupon coverage.

## 16. Explicit exclusions

No:

- Interfax, NSD, CBR, MOEX, issuer adapter or scraping;
- real provider HTTP activation;
- SQL, migration, persistence or replay;
- worker, polling, retry framework or production cache;
- production dependency change;
- monetary heatmap;
- yield, FX, tax, portfolio forecast, notification, amortization or redemption;
- Feature 3D implementation;
- changes to Stage 3.62/3.63 domain semantics.

## 17. Governance state and next gate

The development-path implementation, publication, demonstrated-defect remediation, final pre-evidence exact-head CI,

publishes the previously withheld Internal evidence only after the External verdict and records the exact CI/remediation
chronology. It does not authorize Ready, merge, branch deletion, Feature 3D, or any protected-branch mutation.

After this evidence-only commit is published, the remaining sequence is:

```text
required GitHub CI on evidence-only head
→ explicit Principal Architect Ready + squash-merge authorization
→ squash merge to protected develop
```
