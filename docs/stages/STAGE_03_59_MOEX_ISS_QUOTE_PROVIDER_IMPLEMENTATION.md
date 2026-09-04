# Stage 3.59 — Real MOEX ISS Quote Provider Adapter Implementation

| Field | Value |
| --- | --- |
| Status | Development-path implementation candidate; not committed/pushed; no PR/Ready/merge/runtime activation/public redistribution/Feature 3 authorization implied |
| Date | 2026-09-04 |
| Canonical implementation base | `develop@edf3ffc24c3813f884fd3a4f8a7e9630cb9b8322` |
| Protected-base tree | `c38275122f9c524b4021d52a45e97cdfa1e8505a` |
| Approved planning authority | `docs/stages/STAGE_03_59_MOEX_ISS_QUOTE_PROVIDER_PLANNING.md` / Git blob `865eed1815e4b16c2f87b46f8954d607596a46dc` |
| Source approval | `MOEX_ISS_DELAYED_TQBR` — adapter implementation/test scope only; shipped runtime/public activation forbidden |
| Canonical workflow | `docs/REVIEW_WORKFLOW.md` v1.4.0 |
| OpenAPI / SQL / DB / frontend / dependency / `cmd/api` change | None |

## 1. Purpose and scope

This candidate implements Feature 2 exactly behind the Stage 3.57 application-owned quote boundary:

```text
MOEX ISS delayed HTTP
        ↓
internal/provider/moexiss
        ↓
verticalslice.QuoteProvider
        ↓
verticalslice.MarketQuote
        ↓
verticalslice.Service.MarketQuote
```

The implementation adds a real MOEX ISS TQBR adapter, provider-neutral data-error classification, and an immutable service construction seam. It deliberately does **not** activate MOEX in `cmd/api`, enrich asset search/detail, expose a public quote endpoint, change OpenAPI, persist quotes, add Redis/cache/workers, or begin Feature 3.

## 2. Candidate changed files

The intended repository diff is limited to:

1. `backend-go/internal/verticalslice/marketdata.go` — add only `ErrMarketQuoteProviderData` as the application-owned malformed-provider-data sentinel;
2. `backend-go/internal/verticalslice/marketdata_service.go` — add immutable `NewServiceWithQuoteProvider(...)` construction without changing existing `NewService(...)` call sites; the narrow file is intentional so the market-data DI seam does not expand the large unrelated `service.go`;
3. `backend-go/internal/verticalslice/marketdata_service_test.go` — focused proof that the immutable constructor injects and uses the provider;
4. `backend-go/internal/provider/moexiss/provider.go` — constructor, fixed production host, controlled HTTP client, request construction/status/transport/body handling, canonical quote materialization;
5. `backend-go/internal/provider/moexiss/parser.go` — ISS table/column parsing, exact decimal parsing, trade-date/time mapping, provider-data normalization;
6. `backend-go/internal/provider/moexiss/provider_test.go` — deterministic `httptest`/transport tests for success and failure cases;
7. this implementation record.

No existing `service.go`, `cmd/api`, OpenAPI, HTTP DTO, SQL/migration, frontend, lockfile/dependency, cache/worker, registry, or workflow file changes.

## 3. HTTP and provider contract

The production provider host is fixed inside the package to:

```text
https://iss.moex.com
```

The provider builds the request structurally with `url.URL.JoinPath` and the canonical path:

```text
/iss/engines/stock/markets/shares/boards/TQBR/securities/{SECID}.json
```

The exact query contract is:

```text
iss.meta=off
iss.only=marketdata,dataversion
marketdata.columns=SECID,BOARDID,LAST,TIME
dataversion.columns=trade_date
```

The request sends `Accept: application/json`.

`NewQuoteProvider(client, clock)` requires explicit non-nil dependencies. A typed-nil `Clock` is rejected during construction. The supplied `http.Client` is copied rather than mutated, the provider-owned copy enforces a total `5s` timeout, and its `CookieJar` is cleared so the adapter cannot silently inherit cookie-based access. No package-level `http.Get`, mutable global client, retry loop, polling loop, cache, concurrency fan-out, or custom production transport is introduced.

Tests use a private constructor with an `httptest.Server` base URL; arbitrary base-host injection is not exported as production API.

## 4. Bounded response and HTTP failure semantics

Response bodies are read through `io.LimitReader` with a `64 KiB + 1` detection byte. Any body above `64 KiB` fails closed as `ErrMarketQuoteProviderData`; no truncated body can become success.

Every obtained response body is closed.

Classification is:

```text
429
5xx
transport/DNS/TLS/client-timeout while caller context is live
response-body read failure while caller context is live
→ ErrMarketQuoteProviderUnavailable

caller cancellation / caller deadline
→ preserve ctx.Err()

unexpected non-200 status, including 400/404/204
→ ErrMarketQuoteProviderData
```

No raw response body, `*http.Response`, MOEX DTO, or HTTP status integer crosses the adapter boundary.

## 5. Reorder-safe ISS parsing

`marketdata` and `dataversion` are parsed as independent ISS table blocks with `columns` and `data` arrays using `json.RawMessage` rows.

The parser builds name → index maps and requires:

- `marketdata`: `SECID`, `BOARDID`, `LAST`, `TIME`;
- `dataversion`: `trade_date`.

It rejects missing/duplicate required columns, missing data arrays, malformed row shapes, more than one market row for a successful quote, missing/empty/multiple dataversion rows, mismatched `SECID`, and non-`TQBR` `BOARDID`. Dataversion evidence is validated before quote-absence classification, so malformed trade-date evidence cannot be hidden behind an empty marketdata result.

Reordered columns and unrelated extra columns are accepted. Provider table/schema types stay private to `internal/provider/moexiss`.

## 6. Money correctness

`LAST` remains a raw JSON token until canonical decimal parsing:

```text
json.RawMessage token
→ exact token text
→ decimal.FromString(...)
→ Money{Amount, RUB}
```

There is no `float64` price path.

`LAST: null` is quote absence and maps to `ErrMarketQuoteNotFound`. String-encoded price, exponent notation unsupported by the canonical decimal grammar, invalid JSON token, negative amount, excessive scale, or out-of-storage canonical decimal maps to `ErrMarketQuoteProviderData`.

The existing Stage 3.57 service validator remains the final canonical output guard.

## 7. Time and provenance correctness

`AsOf` is derived only from same-response provider evidence:

```text
dataversion.trade_date + marketdata.TIME
→ Europe/Moscow
→ UTC
```

The package imports standard-library `_ "time/tzdata"` and loads the IANA `Europe/Moscow` zone.

`RetrievedAt` is obtained exactly once from the injected OpenInvest `Clock` after the provider response has been successfully parsed:

```go
clock.Now().UTC()
```

Production adapter code contains no direct `time.Now()` call and does not use HTTP `Date`, MOEX `SYSTIME`, or `AsOf` as retrieval evidence.

Canonical provenance is exactly:

```text
MOEX_ISS
```

No provider symbol, request URL, raw response, parser version, access mode, or response hash is added to the canonical model.

A zero injected clock value yields a zero `RetrievedAt`; the existing service validator then rejects the adapter output as `ErrInvalidMarketQuote`, preserving the Stage 3.57 fail-closed boundary.

## 8. Quote absence and malformed-provider semantics

Quote absence is intentionally narrow:

```text
valid required marketdata columns + empty marketdata.data
LAST = null
→ ErrMarketQuoteNotFound
```

The adapter never substitutes previous close, bid, offer, zero, midpoint, or current time.

Malformed provider/protocol evidence maps to `ErrMarketQuoteProviderData`, including:

- malformed or trailing JSON;
- missing `marketdata` / `dataversion` block;
- missing/duplicate required columns;
- missing row arrays or invalid row length;
- multiple market rows;
- mismatched `SECID`/`BOARDID`;
- invalid/non-canonical/negative `LAST`;
- missing/empty/multiple/malformed dataversion evidence;
- invalid trade date or market time;
- oversized response;
- unexpected non-429 4xx and other non-200 statuses not classified unavailable.

Provider errors remain provider-neutral above the adapter.

## 9. Service construction and production behavior

The candidate adds:

```go
func NewServiceWithQuoteProvider(store Store, clock Clock, provider QuoteProvider) *Service
```

as a narrow immutable construction seam. It delegates to the existing `NewService(store, clock)` first and then sets the package-private provider dependency, so it preserves any existing/future base-constructor initialization rather than duplicating `Service` construction. The existing:

```go
NewService(store, clock)
```

is unchanged.

No mutable setter or functional-options framework is introduced.

Critically, `backend-go/cmd/api` is untouched. Therefore the shipped application still constructs the same provider-free service and cannot perform MOEX network I/O. Production `lastPrice` behavior remains unchanged/null; no public source redistribution is activated.

## 10. Deterministic test coverage

The focused tests cover:

- exact endpoint path/query and `Accept` header;
- exact decimal preservation (`321.12345678`) without float conversion;
- exact ticker, RUB, `MOEX_ISS` provenance;
- `dataversion.trade_date + marketdata.TIME` → Europe/Moscow → UTC;
- `RetrievedAt` UTC normalization and exactly one clock read;
- reordered/extra columns in both ISS blocks;
- empty marketdata and `LAST:null` quote absence, plus regression proving malformed dataversion takes precedence over quote absence;
- 429/500/503 unavailable classification;
- 400/404/204 provider-data classification;
- network error, client-timeout semantics, caller cancellation/deadline preservation;
- response read failure and response-body close;
- enforced 5s timeout and disabled CookieJar without mutating caller client;
- nil and typed-nil dependency rejection;
- malformed/trailing JSON;
- missing each required marketdata column and missing `trade_date`;
- duplicate required columns;
- short/long row shapes;
- multiple market rows;
- wrong SECID/BOARDID;
- string/exponent/negative/excess-scale/out-of-storage LAST;
- missing/empty/multiple/duplicate/invalid dataversion evidence;
- invalid market time;
- oversized body;
- immutable service constructor injection;
- zero `RetrievedAt` rejected by the existing service boundary.

Normal tests do not contact live MOEX.

## 11. Local pre-Internal-review evidence

Focused sandbox gates on the candidate:

- `gofmt` on all candidate Go files: PASS;
- `GOTOOLCHAIN=local go test ./...` in the isolated exact-seam harness: PASS;
- `GOTOOLCHAIN=local go test -race ./...`: PASS;
- `GOTOOLCHAIN=local go vet ./...`: PASS;
- production adapter scan for `float64`: PASS (none);
- production adapter scan for direct `time.Now`: PASS (none);
- current `marketdata.go` base blob cross-check: protected blob `e4e5350116e0d161ebad388e42b78a7b73bdff7c`; candidate semantic change is exactly one provider-data sentinel;
- source/package direction: `moexiss → verticalslice`; reverse import absent by construction;
- no `cmd/api`/OpenAPI/SQL/DB/frontend/dependency/cache/worker candidate surface: PASS.

## 12. Environment limitation

The execution sandbox cannot clone GitHub and has Go `1.23.2`, while canonical `backend-go/go.mod` requires Go `1.25.14`. The local harness therefore uses the exact current OpenInvest decimal implementation plus a narrow seam matching the protected `Money`, `Clock`, `Service`, ticker and Stage 3.57 market-data contracts. Candidate provider/marketdata files themselves are compiled unchanged in that harness.

This is focused prepublication evidence, not repository-wide canonical-toolchain evidence. Exact-head GitHub CI remains mandatory after separately authorized publication and must run all ten protected checks, including repository-wide Go tests/race/vet/security, Python, frontend, OpenAPI, Compose, and PostgreSQL migration validation.

## 13. Architectural consequences

The implementation keeps the dependency inversion created by Feature 1: MOEX depends inward on the OpenInvest port/model; application/domain code does not depend on MOEX transport/schema.

The new data-error sentinel is provider-neutral and can be reused by later adapters without exposing HTTP/provider details. The immutable constructor makes the provider injectable without activating it. No cache/rate scheduler/provider registry/generic factory is introduced.

The implementation intentionally does not solve public redistribution rights, source activation, batching, caching, historical prices, multi-board support, or quote enrichment. Those remain separate gates/stages.

## 14. Governance state

This candidate has not been committed, pushed, or opened as a PR. The complete changed-file set must receive mandatory read-only Internal review in the order:

```text
contract
→ implementation
→ failure cases
→ tests
→ CI evidence
→ architectural consequences
```

Only demonstrated defects receive P0/P1/P2/P3 findings; unknown evidence alone is not a defect.

If Internal review returns `APPROVED`, a separate human commit/push/Draft-PR authorization remains required. Internal review evidence stays withheld until the later fresh External published-head verdict, per `REVIEW_WORKFLOW.md` v1.4.0.
