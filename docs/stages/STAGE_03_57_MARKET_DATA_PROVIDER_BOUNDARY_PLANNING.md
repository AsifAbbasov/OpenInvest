# Stage 3.57 — Market Data Provider Boundary + Provenance/Freshness Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-57-MARKET-DATA-PROVIDER-BOUNDARY-PLAN |
| Version | 0.3.0-candidate |
| Status | Planning-only candidate; grants no runtime implementation, commit/push, PR, Ready, merge, provider-network, OpenAPI, registry-activation, or migration authorization |
| Owner | Builder Engineer |
| Canonical planning base | `develop@56172c451286010a35138212f5a62f33a50ed930` |
| Protected-base tree | `05c3ca748c09380bbc91ca99decb68dd0dae7941` |
| Dependencies | `docs/SOURCE_OF_TRUTH.md`; `docs/ROADMAP.md`; `docs/ARCHITECTURE_FREEZE_v1.2.md`; `docs/REVIEW_WORKFLOW.md` v1.4.0; `docs/registries/DATA_SOURCE_REGISTRY.md`; Stage 2 contract; Stage 3.13 instrument catalog; Stage 3.14 asset API boundary |
| Date | 2026-09-04 |

## 1. Purpose

Stage 3.57 plans the smallest safe market-data boundary needed before OpenInvest can consume real MOEX market data.

The goal is to freeze a provider-neutral canonical quote model and deterministic freshness semantics without introducing a real MOEX ISS HTTP client, background worker, cache, database persistence, OpenAPI change, frontend change, source-registry activation, or new financial calculation.

Provider-specific structures such as `iss.meta`, `boards`, `engines`, `markets`, `securities.columns`, and `marketdata.columns` must never cross a future provider-adapter boundary into application/domain code.

This document is planning only. It changes no runtime behavior and does not authorize implementation by itself.

## 2. Governance reason for a new stage

Stage 3.13 and Stage 3.14 deliberately excluded external-provider integration and market-data ingestion. Stage 3.14 also froze honest runtime behavior: asset search returns `lastPrice: null` until an approved market-data source exists, and asset detail remains deferred when mandatory provenance/detail fields cannot be populated without fabrication.

Stage 3.14 further states that external provider ingestion requires a separate future stage covering Data Source Registry updates, caching policy, freshness policy, auditability, and failure-mode design. The active Data Source Registry states that no external data source is approved and that a collector may not be implemented until its source has an approved row.

Therefore Feature 1 cannot be treated as a normal refactor under the current freeze. A separately reviewed planning stage must first define the internal provider boundary while keeping real provider I/O and production-source activation out of scope.

Stage 3.57 is not a new MVP product-scope expansion: it prepares an internal boundary for the already-frozen asset-price capability (`AssetSummary.lastPrice` / asset `priceAsOf`) that Stage 3.14 intentionally left null/deferred until honest source governance exists.

## 3. Current repository state and immutable planning evidence

At the planning base:

- protected `develop` is `56172c451286010a35138212f5a62f33a50ed930`;
- protected tree is `05c3ca748c09380bbc91ca99decb68dd0dae7941`;
- Architecture Freeze v1.2 and Documentation Freeze are active;
- the original repository audit is complete at 32/32 = 100%;
- `backend-go/internal/verticalslice/models.go` defines the existing application `Clock`, `Money`, and `AssetSummary` types;
- `AssetSummary.LastPrice` is `*Money`;
- `Money.Amount` uses OpenInvest decimal semantics, not binary floating point;
- Stage 3.14 asset search returns `lastPrice: null` because no approved market-data source exists;
- the Stage 2 OpenAPI already permits `AssetSummary.lastPrice` to be either non-negative RUB Money or null;
- the current HTTP asset DTO exposes only the existing `lastPrice` field and no market-data provenance object;
- `docs/registries/DATA_SOURCE_REGISTRY.md` has no approved external source rows;
- the existing test suite already injects a deterministic `fixedClock` into `verticalslice.Service`.

No production provider is authorized by Stage 3.57 planning.

## 4. Architectural boundary

The intended dependency direction is:

```text
provider-specific transport/schema
        ↓
provider adapter
        ↓
application-owned QuoteProvider port
        ↓
canonical MarketQuote
        ↓
OpenInvest application/service
        ↓
future API/snapshot/analytics consumers only after their own explicit authorization
```

Application/domain code must not import or model MOEX ISS response sections, column-array mechanics, HTTP response objects, provider retry state, or raw JSON.

The first implementation remains inside the existing modular monolith.

## 5. Package ownership decision for the first implementation

To avoid a new package cycle around the already-existing `Money` and `Clock` types, the first implementation should keep the canonical quote types and port application-owned inside `backend-go/internal/verticalslice/`, preferably in a narrow file such as `marketdata.go` rather than expanding unrelated model files.

A future provider adapter package may import the application-owned port/types; the application package must never import provider-specific wire-schema types.

Creating a generic provider framework or a broad new bounded context is not authorized by this plan.

## 6. Minimal provider port

Only the currently needed quote boundary is planned:

```go
type QuoteProvider interface {
    Quote(ctx context.Context, ticker string) (MarketQuote, error)
}
```

The following interfaces are explicitly deferred until a concrete reviewed use case exists:

- `InstrumentProvider`;
- `PriceHistoryProvider`;
- generic provider abstractions such as `Provider[T, K, V]`.

YAGNI applies.

### Input semantics

The application supplies an already-canonical OpenInvest ticker. Neither the provider port nor a future adapter may silently trim, lowercase/uppercase, alias-rewrite, or otherwise normalize caller input. Provider-symbol translation, when required by a real source, belongs inside that future adapter and must not change canonical OpenInvest ticker identity.

## 7. Canonical MarketQuote and provenance

The planned canonical model is intentionally small:

```go
type MarketQuote struct {
    Ticker      string
    Price       Money
    AsOf        time.Time
    RetrievedAt time.Time
    Provenance  MarketDataProvenance
}

type MarketDataProvenance struct {
    Provider string
}
```

Semantics:

- `Ticker` is the canonical OpenInvest ticker requested by the application.
- `Price` preserves existing decimal/RUB `Money` semantics and must not be negative.
- `AsOf` is the provider/market observation time to which the quoted value applies.
- `RetrievedAt` is when OpenInvest received/materialized that provider result.
- `AsOf` and `RetrievedAt` are intentionally distinct and may differ.
- `Provenance.Provider` identifies the provider family without exposing provider wire schema.

The first slice must not duplicate `AsOf` or `RetrievedAt` inside provenance.

The provenance value object remains extensible so later reviewed stages may add provider symbol, source/reference, parser version, raw identity/hash, or similar evidence without changing the meaning of `MarketQuote`. The full raw provider JSON must not be stored in the canonical quote model.

No Stage 3.57 implementation may use an `EXAMPLE_*` identifier as runtime production provenance.

## 8. Canonical quote output invariants

A successful provider result crossing the application boundary must satisfy all of these conditions:

- returned ticker equals the exact canonical ticker requested by the application;
- price currency is `RUB`;
- price amount is non-negative and remains decimal;
- `AsOf` and `RetrievedAt` represent instants in UTC before application logic consumes them;
- provenance provider identity is non-empty;
- no provider-specific transport object/raw JSON is embedded in the canonical value.

Provider output that cannot satisfy the canonical contract must fail closed as an error; it must not be converted to zero price, current time, fabricated provenance, or `lastPrice: null` and presented as a successful quote.

The plan intentionally does **not** require `AsOf <= RetrievedAt`: provider clock skew and observation semantics make that relation source-specific. Freshness instead evaluates each timestamp against controlled `now`.

## 9. Time and UTC semantics

The existing `verticalslice.Clock` remains the single application clock abstraction. No second clock mechanism is authorized.

Provider adapters may parse provider observation timestamps, but canonical `MarketQuote.AsOf` and `RetrievedAt` must be normalized to UTC before crossing into application business logic.

`RetrievedAt` is OpenInvest-owned retrieval evidence, not a timestamp copied from provider payload. A future real adapter must receive the existing `Clock` abstraction (or an exact value supplied from that same clock by its caller) and stamp `RetrievedAt = clock.Now().UTC()` at successful retrieval/materialization. It must not call `time.Now()` directly and must not substitute a provider server timestamp for `RetrievedAt`.

Testable business logic must not call `time.Now()` directly. The existing injected `Clock` supplies `now` where current time is required, including freshness classification and future adapter retrieval stamping.

## 10. Freshness classification

Freshness is derived, not persisted as a static market-data fact.

Planned canonical states:

```go
type FreshnessStatus string

const (
    FreshnessFresh   FreshnessStatus = "FRESH"
    FreshnessStale   FreshnessStatus = "STALE"
    FreshnessUnknown FreshnessStatus = "UNKNOWN"
)
```

The first implementation may use one small explicit policy:

```go
type FreshnessPolicy struct {
    MaxRetrievedAge time.Duration
    MaxMarketAge    time.Duration
}
```

and a pure deterministic classifier conceptually equivalent to:

```go
func ClassifyFreshness(now time.Time, quote MarketQuote, policy FreshnessPolicy) FreshnessStatus
```

Frozen semantics:

```text
if now is zero                                              → UNKNOWN
if policy.MaxRetrievedAge <= 0 or policy.MaxMarketAge <= 0 → UNKNOWN
if AsOf or RetrievedAt is zero                              → UNKNOWN
if AsOf or RetrievedAt is after now                         → UNKNOWN
if now-RetrievedAt > MaxRetrievedAge                        → STALE
if now-AsOf > MaxMarketAge                                  → STALE
otherwise                                                   → FRESH
```

Boundary semantics are inclusive: exactly at either configured maximum age is still `FRESH`; only values strictly older than the threshold are `STALE`.

No global production threshold is selected in Stage 3.57. The implementation tests the classifier using explicit policies without claiming a universal MOEX-live threshold.

No `freshness = LIVE`, `FRESH`, or similar database column is authorized.

## 11. Application-service boundary and error semantics

The first implementation may extend the existing application service with the minimum quote-read seam needed to prove the port, without wiring quotes into asset-search HTTP behavior. Conceptually:

```go
func (s *Service) MarketQuote(ctx context.Context, ticker string) (MarketQuote, error)
```

The exact method name may follow repository conventions, but the responsibilities are fixed:

1. validate canonical ticker shape without silently normalizing it;
2. fail closed when no quote provider is configured for that service instance;
3. call exactly the configured `QuoteProvider`;
4. preserve a canonical quote-not-found sentinel for an unknown ticker;
5. propagate other provider/application errors as errors rather than manufacturing data;
6. validate successful canonical quote invariants before returning the quote.

A minimal canonical sentinel such as `ErrMarketQuoteNotFound` may be introduced. Other provider failures do not need a premature large error taxonomy in Feature 1. Provider-specific HTTP/status/parser details must remain adapter-owned and must never be copied into safe public API error DTOs.

This method is an internal application boundary only in Feature 1. No new public HTTP endpoint is authorized.

## 12. API decision — Variant A

Stage 3.57 selects Variant A:

- keep existing `AssetSummary.LastPrice *Money`;
- do not add a public `marketData` object;
- do not change `SourceReference`, `AssetSummary`, or asset response OpenAPI schemas;
- do not populate asset-search `lastPrice` from the fake/test provider;
- keep current production runtime `lastPrice: null` until a separately approved runtime source and integration strategy exist;
- preserve the existing HTTP DTO projection;
- provenance/freshness remain internal to the market-data boundary in Feature 1.

This avoids a contract change before there is a user-visible need for provenance fields and avoids freezing per-search provider fan-out, caching, partial-failure, or N+1 lookup semantics before the real MOEX adapter is designed.

A future public response such as `marketData.asOf/retrievedAt/provider/freshness` is a real OpenAPI contract change and requires its own explicit contract proposal/review before implementation.

## 13. Database decision

No migration is required or authorized for Feature 1.

The first implementation must not persist:

- quotes;
- provenance;
- freshness status;
- raw provider JSON;
- provider cache state.

Persistence/caching is deferred until a separately reviewed use case defines retention, invalidation, source identity, replay/audit requirements, and failure semantics.

## 14. Deterministic fake provider decision

Feature 1 uses a deterministic **test fake**, preferably defined in `_test.go` or equivalent test-only code. It is not a production static provider, not a registered source, and must not be wired by `cmd/api`.

The fake proves only:

- known ticker → canonical `MarketQuote`;
- unknown ticker → canonical quote-not-found sentinel;
- injected provider error → application method returns an error;
- canonical quote validation rejects malformed provider output.

No network, retry, rate-limit, cache, worker, scheduler, parser, or provider credentials are authorized.

## 15. Planned first implementation flow

The separately reviewed implementation proves this path only:

```text
test fake QuoteProvider
        ↓
application-owned QuoteProvider port
        ↓
canonical MarketQuote
        ↓
verticalslice.Service internal quote method
        ↓
deterministic freshness classification / caller-visible internal result
```

Production `cmd/api` construction remains without a quote provider and existing asset search remains `lastPrice: null`.

This deliberately stops before HTTP/search enrichment so Feature 1 does not pre-decide the request fan-out, batching, cache, partial-failure, or provider-availability semantics that belong with a real approved source.

## 16. Candidate implementation surfaces — not authorized by this plan alone

A later Feature 1 implementation PR is expected to remain narrow:

- `backend-go/internal/verticalslice/marketdata.go` — canonical quote/provenance types, `QuoteProvider`, freshness policy/classifier, and minimal canonical errors/invariants;
- `backend-go/internal/verticalslice/service.go` — only the minimum internal quote method / provider wiring required to exercise the port using the existing `Clock`;
- focused `backend-go/internal/verticalslice/*_test.go` tests, including a test-only fake provider;
- existing asset HTTP/API regression tests only if needed to prove production `lastPrice: null` behavior remains unchanged;
- Stage 3.57 implementation evidence/documentation required by the development workflow.

Default expectation: no OpenAPI, SQL, PostgreSQL store, frontend, dependency, CI/workflow, or `cmd/api` production-provider wiring changes.

Implementation must stop for renewed planning/review if it discovers a need to change:

- `openapi/` schemas/responses;
- SQL migrations/database schema;
- `cmd/api` to enable any quote provider;
- existing ledger/snapshot semantics;
- external network configuration;
- Data Source Registry production rows;
- CI/workflows/dependencies;
- frontend contracts;
- production asset-search quote enrichment.

## 17. Required tests for the future implementation

Canonical quote/invariant tests:

- decimal amount is preserved exactly;
- currency remains RUB;
- negative quote price is rejected;
- `AsOf != RetrievedAt` is valid;
- canonical timestamps are UTC;
- provider identity is preserved without provider wire-schema leakage;
- mismatched returned ticker is rejected;
- zero timestamp or empty provider identity cannot become a successful canonical quote.

Freshness tests using controlled time:

- fresh quote;
- stale by retrieval age;
- stale by market observation age;
- unknown for zero `now`;
- unknown for non-positive policy threshold;
- unknown for missing/zero quote timestamp;
- exact threshold remains fresh;
- future `AsOf` yields unknown;
- future `RetrievedAt` yields unknown;
- no direct `time.Now()` in classifier logic.

Provider/application tests:

- known ticker returns expected canonical quote;
- unknown ticker preserves canonical quote-not-found;
- no configured provider fails closed;
- injected provider error crosses the application boundary as an error;
- malformed provider output is rejected rather than normalized/fabricated;
- no raw provider schema enters canonical values.

API regression tests:

- existing asset search response schema remains valid;
- current production construction continues to emit `lastPrice: null` while no approved provider is configured;
- no `marketData` field appears before a separately approved OpenAPI change.

## 18. Explicit exclusions

Stage 3.57 planning and Feature 1 implementation do not authorize:

- real MOEX ISS HTTP;
- provider JSON parsing;
- Data Source Registry production activation;
- Redis;
- Kafka;
- microservices;
- background workers/schedulers;
- caching or quote persistence;
- asset-search quote fan-out/batching;
- price history/candles/order books;
- dividend/coupon ingestion;
- multi-currency;
- Python;
- frontend direct MOEX access;
- frontend business authority;
- `float64` money;
- raw provider JSON in domain/application models;
- ledger or snapshot-engine redesign;
- generic provider frameworks;
- Feature 2.

## 19. Data Source Registry boundary

Stage 3.57 does not add or approve a production source row and does not register the test fake.

Before any real MOEX adapter is enabled at runtime, a later governed stage must define and approve at minimum:

- production provider/source code;
- legal/terms/attribution/redistribution decision required by the registry;
- permitted source/reference semantics;
- provider endpoint/use policy;
- provider-symbol mapping;
- parser/version identity strategy if required;
- retrieval and observation timestamp mapping;
- cache/retention policy if persistence is introduced;
- auditability and safe failure behavior;
- rate-limit/retry/backoff policy;
- exact rules for when a quote may populate user-visible API fields.

## 20. Non-normative design references

These repositories are reference material only and are not dependencies or code sources.

- `wiedehopf/readsb` — `README-json.md` distinguishes the JSON generation/cache timestamp (`now`) from per-object age such as `seen`/`seen_pos`, demonstrating why current values need explicit temporal context.
- `OpenBB-finance/OpenBB` — provider-specific fetchers import/produce standard provider models and transform provider data behind a provider boundary; OpenInvest adopts the canonical-model/adapter principle only, not Python/Pydantic machinery.
- `kan3an00/tradelf-package` / Blankly exchange interfaces — a common application-facing exchange interface is implemented by exchange-specific adapters; OpenInvest adopts only the dependency-boundary principle, not Python machinery or float money semantics.

## 21. Acceptance criteria for Feature 1 implementation

Feature 1 implementation may be accepted only when:

- provider-specific schema does not leak into application/domain code;
- one minimal `QuoteProvider` port exists and no unused provider interfaces are added;
- canonical `MarketQuote` exists with decimal `Money`, distinct `AsOf`, distinct `RetrievedAt`, and provider provenance;
- successful quotes are validated against canonical ticker/RUB/non-negative/timestamp/provider invariants;
- freshness classification is deterministic and uses the existing clock architecture;
- test time is controlled;
- test fake proves provider/application behavior without entering production wiring;
- existing asset API remains contract-compatible and production asset search remains `lastPrice: null`;
- no real MOEX HTTP or production provider wiring is enabled;
- no OpenAPI change or migration is introduced;
- architecture remains a modular monolith;
- implementation has passed the mandatory development-path review/CI/human gates.

## 22. Review focus

Planning review must specifically challenge:

- accidental self-authorization of runtime implementation;
- fake/test provider accidentally becoming production wiring;
- `SearchAssets` enrichment that would prematurely freeze N+1/fan-out/failure semantics;
- Variant A drifting into a hidden OpenAPI change;
- persisted freshness/provenance without a migration/retention design;
- conflation of `AsOf` and `RetrievedAt`;
- direct `time.Now()` bypassing the existing `Clock`;
- silent ticker/provider-symbol normalization across the port;
- provider-specific JSON or raw errors leaking inward/outward;
- fabricated zero/current timestamps or price when provider data is invalid;
- unregistered provider identity appearing in runtime public responses;
- premature `InstrumentProvider`, `PriceHistoryProvider`, cache, worker, or generic framework abstractions;
- any Feature 2 / real MOEX HTTP scope entering Feature 1.

## 23. Planning publication lifecycle

Stage 3.57 planning should follow the established one-document planning precedent used by later Stage 3.x plans: the permanent planning PR should contain only this planning artifact unless a reviewer identifies a concrete canonical-registry inconsistency that must be repaired in the same planning scope.

The planning artifact must remain publication-stable:

- it predicts no future PR number, published head, CI run number, or squash SHA;
- reviewer verdicts are evidence bound to exact candidate identities, not self-authored approval fields in the document;
- mutable repository facts are anchored to the immutable planning base above;
- Feature 1 runtime behavior remains unchanged throughout planning publication;
- planning merge does not approve a production provider.

Planning publication requires read-only planning review, explicit human commit/push authorization, Draft PR, exact-head CI, published-head planning verification, and separate human merge authorization under `docs/REVIEW_WORKFLOW.md` v1.4.0.

## 24. Next governed action

Send this exact Stage 3.57 planning candidate to the designated review chat for complete read-only planning review.

Only after `APPROVED` and separate human commit/push authorization may it be committed/pushed and opened as a Draft PR to protected `develop`.

Only after that planning PR is exact-head verified and separately authorized/squash-merged into protected `develop` may a Feature 1 runtime implementation branch begin.

Feature 2 must not begin automatically after Feature 1.
