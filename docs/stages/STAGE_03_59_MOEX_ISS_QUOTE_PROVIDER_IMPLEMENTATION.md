# Stage 3.59 — Real MOEX ISS Quote Provider Adapter Implementation

| Field | Value |
| --- | --- |
| Status | Development-path implementation candidate; not committed/pushed; no PR/Ready/merge/runtime activation/public redistribution/Feature 3 authorization implied |
| Date | 2026-09-04 |
| Canonical implementation base | `develop@edf3ffc24c3813f884fd3a4f8a7e9630cb9b8322` |
| Protected-base tree | `c38275122f9c524b4021d52a45e97cdfa1e8505a` |
| Approved planning authority | `docs/stages/STAGE_03_59_MOEX_ISS_QUOTE_PROVIDER_PLANNING.md` / Git blob `865eed1815e4b16c2f87b46f8954d607596a46dc` |
| Source approval | `MOEX_ISS_DELAYED_TQBR` — adapter implementation/test scope only; shipped runtime/public activation forbidden |
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

## 12. Environment limitation

The execution sandbox cannot clone GitHub and has Go `1.23.2`, while canonical `backend-go/go.mod` requires Go `1.25.14`. The local harness therefore uses the exact current OpenInvest decimal implementation plus a narrow seam matching the protected `Money`, `Clock`, `Service`, ticker and Stage 3.57 market-data contracts. Candidate provider/marketdata files themselves are compiled unchanged in that harness.

This is focused prepublication evidence, not repository-wide canonical-toolchain evidence. Exact-head GitHub CI remains mandatory after separately authorized publication and must run all ten protected checks, including repository-wide Go tests/race/vet/security, Python, frontend, OpenAPI, Compose, and PostgreSQL migration validation.

## 13. Architectural consequences

The implementation keeps the dependency inversion created by Feature 1: MOEX depends inward on the OpenInvest port/model; application/domain code does not depend on MOEX transport/schema.

The new data-error sentinel is provider-neutral and can be reused by later adapters without exposing HTTP/provider details. The immutable constructor makes the provider injectable without activating it. No cache/rate scheduler/provider registry/generic factory is introduced.

The implementation intentionally does not solve public redistribution rights, source activation, batching, caching, historical prices, multi-board support, or quote enrichment. Those remain separate gates/stages.

## 14. Governance state


```text
contract
→ implementation
→ failure cases
→ tests
→ CI evidence
→ architectural consequences
```

Only demonstrated defects receive P0/P1/P2/P3 findings; unknown evidence alone is not a defect.


This section is an evidence-only follow-up added only after the fresh External published-head phase. Sections 1–14 are preserved as the historical prepublication implementation record and are not retroactively rewritten. Statements there such as “not committed/pushed” describe the state at that earlier gate.

### 15.2 Initial published implementation head


- exact head `fe265e9127b44b2fe5899b62b1fc47c8429075e7`;
- exact tree `c688b76322eacd89961952233c2e936014e98b30`;
- exactly seven candidate files;
- CI `#327` / run `33900012933`: all ten required jobs `SUCCESS`.


1. the provider-owned HTTP client inherited redirect-following behavior, so a provider 3xx could escape the fixed-host boundary;
2. an interface holding a typed-nil `QuoteProvider` could bypass `s.quoteProvider == nil` and allow a nil-receiver call/panic instead of deterministic fail-closed behavior.


### 15.3 External remediation

The first P2 was remediated by making the provider-owned HTTP client override redirect behavior with `http.ErrUseLastResponse`. Redirect responses therefore remain responses from the original request, are never automatically followed to another host, and fall through the existing unexpected-non-200 provider-data classification. A dedicated regression test proves the redirect target receives zero requests.

The second P2 was remediated by making `NewServiceWithQuoteProvider` normalize nil and typed-nil provider interface values to the existing unconfigured service state. A dedicated regression test uses a typed-nil provider whose `Quote` method would panic if invoked and proves the service instead returns `ErrMarketQuoteProviderUnavailable`.

The remediation added only two test surfaces beyond the original seven-file subject:

- `backend-go/internal/provider/moexiss/provider_redirect_test.go`;
- `backend-go/internal/verticalslice/marketdata_service_nil_test.go`.

No `cmd/api`, OpenAPI, SQL/DB/migration, frontend, dependency/lockfile, cache/worker, source-activation, or Feature 3 surface was added.

The corrected semantic publication is:

- exact head `4605753d68b56227c646663446700271042cc299`;
- exact tree `2c14935f398214e8536c86c59daadbff5dcde68d`;
- PR diff: exactly nine expected files — the original seven plus the two remediation regression-test files;
- corrected exact-head CI `#331` / run `33900381845`: all ten required jobs `SUCCESS`;

### 15.4 Publication tooling incidents


- known incorrect unreferenced blob objects included `d5a9b5bee2fbc33eaf4c3a8fdf2652b70bd2f75f`, `03749bbaec426f779a0460a8e16a5d9da248de15`, and `7600947f3397a099397cd53296bda2be228d8c2d` while attempting large-file transfer;
- temporary placeholder branch commits `35d49814f89d92ae6be7624f36a34ad4b1e81be7` and `dde40944dab8cd32342ba579b0da0a2bd5d62d46` were created during write-path testing and then removed from the feature ref before the Draft PR was opened;
- incorrect path-to-blob assembly commit `9cf9a9b9718bdaae1e189d03b921a644fd619c22`, tree `2c45923cf893dc4bdcfa22204ba01e6a436b60bd`, assigned valid frozen blob contents to wrong repository paths; the error was caught by explicit path-to-blob read-back before PR publication and the feature ref was moved to a corrected tree.


### 15.5 Evidence-publication rule

This section publishes review/evidence chronology only. It does not authorize shipped MOEX runtime activation, public display/redistribution, asset enrichment, OpenAPI, SQL/DB, frontend, dependency, cache/worker, or Feature 3 changes.
