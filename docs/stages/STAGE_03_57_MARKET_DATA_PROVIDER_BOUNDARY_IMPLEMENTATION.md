# Stage 3.57 — Market Data Provider Boundary + Provenance/Freshness Implementation

| Field | Value |
| --- | --- |
| Status | RUNTIME MERGED / FEATURE 1 TECHNICALLY COMPLETE — PR #120 squash-merged into protected `develop` at `cd97f3217811bb123ad96d92b7d8a4be0e03c8bb`, tree `0510971289c204e9b5226359f2efdd1941542309`; canonical lifecycle-document closure is handled separately by Stage 3.58 |
| Date | 2026-09-04 |
| Canonical implementation base | `develop@8316d404d057f0a895713bd1d496a342409903c4` |
| Protected-base tree | `1c596fbb9c26016bdc5c6036f09d6e081bdbcf7f` |
| Approved planning authority | `docs/stages/STAGE_03_57_MARKET_DATA_PROVIDER_BOUNDARY_PLANNING.md` / Git blob `6ddd4682b49a6a259c64474d9adf8882279eca5d` / SHA-256 `e85fc028550663b51daafdea14deddc18f79ae1a3c917e3f5a8c414d5f5ce8ed` |
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

## 8. Environment limitations

The execution sandbox cannot directly clone the GitHub repository and has Go `1.23.2`, while canonical `backend-go/go.mod` requires Go `1.25.14`. The focused harness therefore compiles the new candidate against the current decimal implementation and a narrow type seam matching the protected `Service`/`Clock`/`Money`/ticker contract; it is not presented as repository-wide or canonical-toolchain evidence.

Repository-wide `go test ./...`, race/vet coverage required by CI, existing HTTP asset regression tests, OpenAPI validation, security checks, and the remaining protected-branch checks remain mandatory on the exact published head after the separately authorized Draft PR publication. This limitation is not converted into a false PASS claim.

## 9. Governance state


Feature 2 is not authorized by this implementation record and must not begin automatically.


This section is an evidence-only follow-up added after the fresh External published-head verdict. Sections 1–9 remain the historical prepublication implementation record and are not retroactively rewritten.

### Published implementation milestone

- Draft PR: `#120`;
- initial implementation head: `95ec859481d08e1f53e090834a6bb39f0a845dfa`;
- initial implementation tree: `3bc132a41c58b9b66ae506a904c7562e89a9c506`;
- initial implementation scope: exactly four files listed in section 2;
- CI: `#320` / run `33861442441` / all ten required jobs `SUCCESS`;


### Publication tooling incident preserved

During Git-object assembly, four unreferenced blob objects were inadvertently created by connector calls:

- `30d74d258442c7c65512eafab474568dd706c430`;
- `a8b6c9477b693176aa811d7fc0fd5374d44e3557`;
- `c1b0730e0133447badcfd47fd144e254807b06e1`;
- `8485e986e458a566e6f6160f71d704edc10c57fc`.

They are not referenced by publication tree `3bc132a41c58b9b66ae506a904c7562e89a9c506`, implementation commit `95ec859481d08e1f53e090834a6bb39f0a845dfa`, the implementation branch, or the PR diff. The fresh External re-check classified this as `NO MATERIAL PROJECT FINDING`; runtime/document semantics are unchanged. The incident is preserved here for forensic completeness rather than omitted.

### Remaining gate


## 11. Protected merge and Stage 3.58 closure handoff

The `Remaining gate` subsection in section 10 is preserved as the historical state at the
evidence-publication head. That gate was subsequently satisfied.

PR #120 was separately authorized for Ready and squash merge at exact final evidence head
`35db51707fce970e67bf9d5a9485f79619ec366d`, tree
`0510971289c204e9b5226359f2efdd1941542309`.

GitHub recorded the protected squash merge:

- merge commit `cd97f3217811bb123ad96d92b7d8a4be0e03c8bb`;
- protected merge tree `0510971289c204e9b5226359f2efdd1941542309`;
- parent `8316d404d057f0a895713bd1d496a342409903c4`;
- PR #120 state `merged=true`.

The protected merge tree is identical to the exact authorized PR tree, so no tree drift occurred.

Feature 1 runtime is therefore complete. It establishes the internal provider-neutral quote boundary,
canonical quote/provenance model, deterministic freshness semantics, fail-closed validation, and
test-only fake provider proof described above.

The merge does not add or authorize real MOEX ISS HTTP/parsing, production provider wiring, Data
Source Registry activation, OpenAPI/DB/frontend changes, caching/workers, production asset-search
enrichment, or Feature 2.

Stage 3.58 is a separate documentation/governance-only lifecycle closure. Once its approved closure
record and synchronized canonical surfaces are present on protected `develop`, Stage 3.57 Feature 1
is canonically CLOSED. Feature 2 remains subject to a separate governed stage and explicit human
authorization.
