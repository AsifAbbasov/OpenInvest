# Stage 3.58 — Market Data Provider Boundary Feature 1 Closure

| Field | Value |
| --- | --- |
| Status | MERGE-ACTIVATED CLOSURE RECORD — before protected activation Stage 3.57 Feature 1 runtime is already merged but lifecycle documentation synchronization remains pending; once this record and the synchronized canonical surfaces are present on protected `develop`, Feature 1 is canonically CLOSED |
| Date | 2026-09-04 |
| Canonical workflow | `docs/REVIEW_WORKFLOW.md` v1.4.0 |
| Closure base | protected `develop@cd97f3217811bb123ad96d92b7d8a4be0e03c8bb` / tree `0510971289c204e9b5226359f2efdd1941542309` |
| Stage 3.57 planning | PR #119 / squash merge `8316d404d057f0a895713bd1d496a342409903c4` / plan blob `6ddd4682b49a6a259c64474d9adf8882279eca5d` / SHA-256 `e85fc028550663b51daafdea14deddc18f79ae1a3c917e3f5a8c414d5f5ce8ed` |
| Stage 3.57 implementation | PR #120 / initial implementation head `95ec859481d08e1f53e090834a6bb39f0a845dfa` / final evidence head `35db51707fce970e67bf9d5a9485f79619ec366d` / approved tree `0510971289c204e9b5226359f2efdd1941542309` / squash merge `cd97f3217811bb123ad96d92b7d8a4be0e03c8bb` |
| Stage 3.57 final CI | CI #321 / run `33861987999` — 10/10 SUCCESS on final evidence head; same-head dependency-security retry followed an external npm-registry timeout and changed no project bytes |
| Stage 3.57 review | Internal `APPROVED`; fresh External published-head `APPROVED`; evidence-publication verification `APPROVED`; blocking findings none |
| Runtime/schema/data/SQL/OpenAPI/frontend/dependency change | None — Stage 3.58 is documentation/governance synchronization only |
| Feature 2 authorization | None — this closure does not authorize real MOEX ISS HTTP, provider parsing, source activation, production provider wiring, API enrichment, persistence, cache, worker, or frontend work |

## 1. Purpose and closure boundary

Stage 3.57 Feature 1 created the smallest provider-neutral market-data seam required before OpenInvest can safely integrate a real market source. The runtime implementation is already canonical on protected `develop` through PR #120.

Stage 3.58 does not add market-data behavior. It closes the documentation lifecycle debt left after the implementation merge by synchronizing current repository authority around four questions:

1. what Feature 1 implemented;
2. why the boundary exists;
3. what evidence proves the implementation;
4. what remains explicitly outside Feature 1 and requires a separate next stage.

Before this exact Stage 3.58 closure record and the synchronized canonical surfaces are present on protected `develop`, Feature 1 runtime is technically merged but its lifecycle documentation remains incompletely synchronized. Once they are present after the governed closure path, Stage 3.57 Feature 1 is canonically CLOSED.

## 2. What Feature 1 implemented

The protected implementation establishes this dependency direction:

```text
future provider-specific adapter
        ↓
application-owned QuoteProvider
        ↓
canonical MarketQuote
        ↓
verticalslice.Service.MarketQuote
```

The application-owned quote port is intentionally minimal:

```go
type QuoteProvider interface {
    Quote(ctx context.Context, ticker string) (MarketQuote, error)
}
```

The canonical quote carries only the application facts currently required:

- canonical OpenInvest ticker;
- existing decimal/RUB `Money`;
- provider/market observation timestamp `AsOf`;
- OpenInvest retrieval timestamp `RetrievedAt`;
- minimal provenance provider identity.

Feature 1 also defines deterministic derived freshness through `FreshnessPolicy` and the states `FRESH`, `STALE`, and `UNKNOWN`.

No MOEX ISS wire-schema object, HTTP response object, raw provider JSON, retry state, or provider-specific column-array representation is allowed to cross into the canonical application model.

## 3. Why this boundary exists

The boundary prevents OpenInvest business/application code from depending directly on provider-specific transport and schema details such as MOEX ISS column arrays, endpoint response sections, HTTP status handling, and parsing mechanics.

This preserves:

- dependency inversion: provider adapters depend on the application-owned port, not vice versa;
- deterministic money semantics: prices remain canonical decimal `Money`, never `float64`;
- explicit time semantics: `AsOf` and `RetrievedAt` remain distinct;
- fail-closed financial behavior: malformed provider output cannot become fabricated zero/current/unknown market data;
- testability: freshness and quote validation remain deterministic and independent from wall-clock/network behavior;
- YAGNI: no generic provider framework, cache, registry/factory/plugin system, persistence, or public market-data DTO was added before a real source requires it.

## 4. Protected implementation identity

Protected `develop` after PR #120 is:

- commit `cd97f3217811bb123ad96d92b7d8a4be0e03c8bb`;
- tree `0510971289c204e9b5226359f2efdd1941542309`;
- parent `8316d404d057f0a895713bd1d496a342409903c4`.

The protected tree is byte-identical to the final authorized PR #120 tree. The squash merge therefore introduced no tree drift.

The Stage 3.57 runtime surfaces are exactly:

- `backend-go/internal/verticalslice/marketdata.go`;
- `backend-go/internal/verticalslice/marketdata_test.go`;
- the minimal `quoteProvider QuoteProvider` field change in `backend-go/internal/verticalslice/service.go`.

The implementation record is:

- `docs/stages/STAGE_03_57_MARKET_DATA_PROVIDER_BOUNDARY_IMPLEMENTATION.md`.

## 5. Canonical behavioral contract closed by Feature 1

A successful application quote must preserve:

- exact requested canonical ticker identity;
- RUB currency;
- canonical decimal storage representation;
- non-negative price;
- non-zero UTC `AsOf`;
- non-zero UTC `RetrievedAt`;
- non-empty provider provenance.

Malformed provider output fails closed as `ErrInvalidMarketQuote`.

An unconfigured service fails closed as `ErrMarketQuoteProviderUnavailable`.

Unknown ticker semantics use the canonical `ErrMarketQuoteNotFound` sentinel.

Other provider/application errors remain errors; Feature 1 intentionally does not freeze a premature HTTP/provider transport taxonomy.

Ticker input is validated without silent trim, case conversion, alias rewriting, or other normalization.

`AsOf <= RetrievedAt` is deliberately not a canonical invariant because source clock skew and observation semantics are provider-specific.

## 6. Freshness contract closed by Feature 1

Freshness remains derived, not persisted.

The deterministic classifier uses explicit `now`, `MarketQuote`, and `FreshnessPolicy` values.

Frozen semantics are:

```text
zero now                                              → UNKNOWN
non-positive MaxRetrievedAge or MaxMarketAge         → UNKNOWN
zero AsOf or RetrievedAt                              → UNKNOWN
future AsOf or RetrievedAt                            → UNKNOWN
retrieval age > MaxRetrievedAge                      → STALE
market age > MaxMarketAge                            → STALE
otherwise                                             → FRESH
```

Exact-threshold values remain `FRESH`; only strictly older values are `STALE`.

No universal production freshness threshold is selected by Feature 1.

## 7. Verification and review evidence

The final evidence head `35db51707fce970e67bf9d5a9485f79619ec366d` completed CI #321 / run `33861987999` with all ten required checks successful:

1. Go tests;
2. Python tests;
3. Frontend build and typecheck;
4. OpenAPI contract;
5. Docker Compose config;
6. PostgreSQL migration validation;
7. Go vet;
8. Go race tests;
9. Go vulnerability scan;
10. Dependency security scan.

The first dependency-security attempt encountered an external timeout while `pnpm audit` contacted `registry.npmjs.org`. The job was rerun on the same exact head without any project-byte change and succeeded. This is preserved as infrastructure chronology, not reclassified as an application defect.

Review evidence:

- frozen Internal review: `APPROVED`, blocking findings none;
- fresh External published-head review: `APPROVED`, blocking findings none;
- evidence-only follow-up changed only the Stage 3.57 implementation document;
- final evidence-publication verification: `APPROVED`, semantic/runtime drift `NONE`.

## 8. Explicit non-scope preserved

Feature 1 and this closure do not include or authorize:

- real MOEX ISS HTTP requests;
- MOEX JSON/column-array parsing;
- Data Source Registry production-source activation;
- production `QuoteProvider` constructor/composition-root wiring;
- public market-data OpenAPI changes;
- `SearchAssets` quote enrichment;
- DB/SQL migration or quote persistence;
- Redis or another cache;
- workers, schedulers, Kafka, microservices, or background ingestion;
- frontend changes;
- historical prices/candles/order book;
- dividend/coupon provider integration;
- multi-currency market prices;
- Feature 2 implementation.

Production asset search therefore continues to preserve the existing honest `lastPrice: null` behavior until a separately approved source/integration stage changes that contract.

## 9. Feature 2 handoff

The next market-data work is not implicitly activated by Stage 3.58.

A separately reviewed Feature 2 stage must decide the concrete real-provider integration scope before implementation. At minimum it must address, where applicable:

- exact MOEX ISS endpoint(s) and response contract;
- adapter-owned wire-schema parsing and column mapping;
- canonical ticker/provider-symbol mapping without changing OpenInvest ticker identity;
- existing `Clock` use for `RetrievedAt`;
- timeout and bounded HTTP client behavior;
- provider error normalization so `net/http`, status codes, and MOEX parsing types do not leak into application/domain code;
- provider identity and source governance;
- safe production dependency injection/composition-root wiring;
- whether any user-visible API enrichment is needed.

If Feature 2 requires OpenAPI changes, persistence/migrations, caching, workers, Data Source Registry activation, or asset-search fan-out semantics, those surfaces require explicit scope and review rather than being smuggled through this closure.

## 10. Documentation synchronization scope

Stage 3.58 synchronizes only the current canonical closure surfaces used by the recent post-development closure pattern:

- `docs/SOURCE_OF_TRUTH.md`;
- `docs/ROADMAP.md`;
- `docs/stages/STAGE_03_57_MARKET_DATA_PROVIDER_BOUNDARY_IMPLEMENTATION.md`;
- this closure record.

Older broad registry files that already lag multiple later stages are not silently rewritten as part of this narrow Feature 1 closure. Any repository-wide registry modernization is separate documentation debt and must not be conflated with Stage 3.57 runtime correctness or Feature 2 authorization.

## 11. Activation rule

This record deliberately does not predict its own future PR number, published head, CI run, or squash-merge SHA.

Activation is structural:

1. Stage 3.57 runtime remains canonical on protected `develop` at merge `cd97f3217811bb123ad96d92b7d8a4be0e03c8bb`;
2. before this exact closure record and synchronized canonical surfaces are present on protected `develop`, lifecycle documentation synchronization remains pending;
3. the complete Stage 3.58 change must remain documentation/governance-only;
4. it must pass the mandatory post-development governance/closure path under `docs/REVIEW_WORKFLOW.md` v1.4.0;
5. once the approved closure record and synchronized surfaces are squash-merged into protected `develop`, Stage 3.57 Feature 1 is canonically CLOSED;
6. Feature 2 remains NOT AUTHORIZED until a separate explicit governed stage and human authorization.

No branch deletion is authorized by this record.
