# Stage 3.57 — Market Data Provider Boundary + Provenance/Freshness Implementation

| Field | Value |
| --- | --- |
| Status | Development-path implementation candidate; no commit/push/PR/Ready/merge authorization implied |
| Date | 2026-09-04 |
| Canonical implementation base | `develop@8316d404d057f0a895713bd1d496a342409903c4` |
| Protected-base tree | `1c596fbb9c26016bdc5c6036f09d6e081bdbcf7f` |
| Approved planning authority | `docs/stages/STAGE_03_57_MARKET_DATA_PROVIDER_BOUNDARY_PLANNING.md` / Git blob `6ddd4682b49a6a259c64474d9adf8882279eca5d` / SHA-256 `e85fc028550663b51daafdea14deddc18f79ae1a3c917e3f5a8c414d5f5ce8ed` |
| Internal Review Evidence | `WITHHELD — external published-head phase pending` |
| OpenAPI / SQL / DB migration / frontend / CI dependency change | None |
| Real provider network I/O | None |

## 1. Scope

This candidate implements only Feature 1 from the protected Stage 3.57 planning authority: an internal provider-neutral quote boundary, canonical market quote/provenance model, deterministic freshness classification, and a narrow application-service read seam.

It does not connect MOEX ISS, parse provider JSON, register a production data source, enrich asset-search responses, add a public market-data DTO, persist quotes/provenance/freshness, introduce cache/workers, or begin Feature 2.

The intended dependency direction remains:

```text
future provider adapter
        ↓
QuoteProvider
        ↓
MarketQuote
        ↓
verticalslice.Service.MarketQuote
```

Provider-specific transport/schema types do not appear in the candidate.

## 2. Candidate changed files

The complete intended repository diff is limited to:

- `backend-go/internal/verticalslice/service.go` — add one unexported optional `quoteProvider QuoteProvider` dependency field; existing `NewService(store, clock)` remains unchanged and therefore production construction still has no quote provider;
- `backend-go/internal/verticalslice/marketdata.go` — canonical quote/provenance types, minimal provider port/errors, quote validation, deterministic freshness policy/classifier, and the internal `Service.MarketQuote` application seam;
- `backend-go/internal/verticalslice/marketdata_test.go` — deterministic test-only fake provider and focused quote/provider/freshness tests;
- this implementation record.

No `cmd/api`, `openapi/`, `internal/httpapi`, PostgreSQL, migrations, dependency/lockfile, frontend, workflow, ledger, or snapshot-engine file is changed.

## 3. Canonical market-data boundary

The implementation introduces only the currently required port:

```go
type QuoteProvider interface {
    Quote(ctx context.Context, ticker string) (MarketQuote, error)
}
```

It deliberately does not add `InstrumentProvider`, `PriceHistoryProvider`, a generic provider framework, cache abstraction, or production adapter.

`MarketQuote` carries:

- canonical ticker;
- existing decimal `Money` price;
- provider/market observation time `AsOf`;
- OpenInvest retrieval evidence `RetrievedAt`;
- minimal `MarketDataProvenance{Provider}`.

`AsOf` and `RetrievedAt` remain distinct. The validator intentionally does not require `AsOf <= RetrievedAt`.

## 4. Fail-closed quote invariants and errors

A successful application quote must preserve the exact requested canonical ticker, non-negative RUB decimal money, non-zero UTC `AsOf`, non-zero UTC `RetrievedAt`, and a non-empty canonical provider identifier.

Malformed provider output returns `ErrInvalidMarketQuote`; it is never converted to zero price, current time, fabricated provenance, or a successful null-price result. An unconfigured service returns `ErrMarketQuoteProviderUnavailable`. The deterministic test fake uses `ErrMarketQuoteNotFound` for an unknown ticker. Other provider failures propagate as errors without provider-specific transport objects entering the canonical model.

Ticker input is validated without trimming, uppercasing, alias rewriting, or any other silent normalization.

## 5. Freshness semantics

Freshness remains derived and is not persisted. `ClassifyFreshness` is a pure deterministic function using explicit `now` and `FreshnessPolicy` values. It returns only `FRESH`, `STALE`, or `UNKNOWN` with the exact inclusive threshold semantics frozen by the Stage 3.57 plan.

The implementation creates no second clock abstraction and contains no direct `time.Now()` call. Controlled tests obtain `now` through a type implementing the existing `verticalslice.Clock` interface.

No universal production freshness threshold is introduced.

## 6. API and production behavior

Variant A remains unchanged:

- `AssetSummary.LastPrice *Money` is untouched;
- no public `marketData` object is added;
- OpenAPI is unchanged;
- `SearchAssets` is not enriched through the quote provider;
- `cmd/api` construction is unchanged;
- production asset search therefore continues to return `lastPrice: null` until a separately approved runtime source/integration stage exists.

The existing API regression test `TestAssetSearchReturnsCatalogSummariesWithoutPriceOrSource` remains authoritative and is expected to run unchanged in repository CI.

## 7. Local pre-Internal-review evidence

Focused sandbox gates on the frozen candidate:

- `gofmt` for `marketdata.go` and `marketdata_test.go`: PASS;
- focused `go test -count=1 ./internal/verticalslice`: PASS in the isolated candidate harness;
- focused `go test -race -count=1 ./internal/verticalslice`: PASS in the isolated candidate harness;
- focused `go vet ./internal/verticalslice`: PASS in the isolated candidate harness;
- direct `time.Now()` scan across the new market-data production/test files: PASS (none present);
- fake provider is defined only in `_test.go`: PASS;
- candidate scope contains no OpenAPI, SQL, DB, frontend, CI, dependency, `cmd/api`, ledger, or snapshot change: PASS.

The focused tests cover exact decimal amount/currency preservation, distinct timestamps, `AsOf > RetrievedAt` acceptance, UTC canonical timestamps, known/unknown provider behavior, provider error propagation, no-provider fail-closed behavior, malformed output rejection, zero-value decimal rejection, and deterministic freshness including fresh/stale/unknown, non-positive policy, exact-threshold, future-timestamp, and non-UTC-current-time cases.

## 8. Environment limitations

The execution sandbox cannot directly clone the GitHub repository and has Go `1.23.2`, while canonical `backend-go/go.mod` requires Go `1.25.14`. The focused harness therefore compiles the new candidate against the current decimal implementation and a narrow type seam matching the protected `Service`/`Clock`/`Money`/ticker contract; it is not presented as repository-wide or canonical-toolchain evidence.

Repository-wide `go test ./...`, race/vet coverage required by CI, existing HTTP asset regression tests, OpenAPI validation, security checks, and the remaining protected-branch checks remain mandatory on the exact published head after the separately authorized Draft PR publication. This limitation is not converted into a false PASS claim.

## 9. Governance state

This candidate is on the mandatory development path. Before any commit/push, the complete changed-file set must receive read-only Internal line-by-line review and an `APPROVED` verdict. Internal review evidence remains withheld from the Draft PR/repository evidence surface until the later External published-head verdict, as required by `docs/REVIEW_WORKFLOW.md` v1.4.0.

After Internal approval, a separate human commit/push authorization is still required. Draft PR publication, exact-head GitHub CI, fresh External published-head review, later evidence-only publication/verification, and separate human Ready/squash-merge authorization remain mandatory.

Feature 2 is not authorized by this implementation record and must not begin automatically.
