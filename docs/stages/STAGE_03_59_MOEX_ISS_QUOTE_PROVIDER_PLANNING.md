# Stage 3.59 — Real MOEX ISS Quote Provider Adapter Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-59-MOEX-ISS-QUOTE-PROVIDER-PLAN |
| Version | 0.1.0-candidate |
| Status | Planning-only development-path candidate; no implementation, commit/push, PR, Ready, merge, public redistribution, production source activation, OpenAPI, SQL/DB, frontend, cache, worker, or Feature 3 authorization implied |
| Owner | Builder Engineer |
| Canonical planning base | `develop@2e8d67f980aba5a2d7c4a33d5721ff6a7ad4951a` |
| Protected-base tree | `a73aa0c7c24ee9bdd172a7ffcf95ab07bd533b65` |
| Canonical workflow | `docs/REVIEW_WORKFLOW.md` v1.4.0 |
| Dependencies | Stage 3.57 Feature 1; Stage 3.58 closure; `docs/registries/DATA_SOURCE_REGISTRY.md`; existing decimal/Money/Clock contracts |
| Date | 2026-09-04 |

## 1. Purpose

Stage 3.59 plans Feature 2: the first real provider-specific adapter behind the canonical application-owned
`QuoteProvider` boundary introduced and closed by Stage 3.57/3.58.

The intended dependency direction is:

```text
MOEX ISS delayed HTTP
        ↓
backend-go/internal/provider/moexiss
        ↓
verticalslice.QuoteProvider
        ↓
verticalslice.MarketQuote
        ↓
verticalslice.Service.MarketQuote
```

The adapter owns HTTP, MOEX ISS table/column mechanics, provider response parsing, provider-specific
timestamps and status handling. Application/domain code continues to know only the canonical
`QuoteProvider`, `MarketQuote`, `Money`, `Clock`, and provider-neutral error sentinels.

This plan is intentionally narrower than user-visible market-data integration. It does not authorize
asset-search enrichment, a public quote endpoint, public provenance/freshness fields, persistence,
caching, workers, or a frontend change.

## 2. Governance and source-approval constraint

The active Data Source Registry currently states that no external source is approved and that a collector
may not be implemented until an approved source row exists.

Reviewed current MOEX materials establish two separate facts:

1. MOEX ISS is an HTTP API for market data, and unauthenticated/free access provides delayed data
   (approximately 15 minutes delayed for market data);
2. onward distribution/public display and non-display/derived-data rights are subject to separate
   commercial terms and tariffs; free ISS access is not treated by this plan as proof of redistribution rights.

Therefore Stage 3.59 proposes a deliberately limited source approval:

- real MOEX ISS delayed adapter development is permitted;
- deterministic local tests are permitted, and a human-run manual smoke against the public delayed endpoint may be recorded as supplemental evidence;
- public API/UI redistribution is forbidden;
- production source activation is forbidden;
- production `lastPrice` remains unchanged/null;
- redistribution/non-display rights must be re-reviewed before any later public or production activation.

This is a technical/governance scope decision, not legal advice.

Official evidence reviewed on 2026-09-04:

- MOEX ISS API overview: `https://www.moex.com/a8531`
- MOEX ISS API overview (RU): `https://www.moex.com/a2920`
- MOEX market-data tariffs/redistribution: `https://www.moex.com/s1147`
- MOEX market-data access overview: `https://www.moex.com/s3630`
- MOEX equities trading schedule: `https://www.moex.com/s866`
- MOEX current prices field description: `https://www.moex.com/ru/current-prices.aspx`
- MOEX/ALGOPACK ISS TQBR endpoint documentation: `https://moexalgo.github.io/docs/api/get-all-shares-statistics/`

No explicit public hard request quota was found in the reviewed public ISS materials. That fact is recorded
as UNKNOWN provider quota, not as a defect. Feature 2 compensates with a deliberately conservative client
policy and no retry/fan-out behavior.

## 3. Source registry row frozen by this plan

The Stage 3.59 planning publication is expected to add exactly one limited approval row to
`docs/registries/DATA_SOURCE_REGISTRY.md`:

| Source | Owner | License/terms | Rate limits | Caching | Redistribution | Freshness | Fallback | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `MOEX_ISS_DELAYED_TQBR` | Principal Architect / Market Data | Public delayed ISS access is documented by MOEX; real-time requires subscription. Free access does not establish OpenInvest redistribution/non-display rights. Attribution/display obligations for any future public surface are NOT ESTABLISHED and must be re-reviewed. | MOEX public hard quota: UNKNOWN in reviewed docs. OpenInvest Feature 2 adapter policy: one request per `Quote` call, no automatic retry, no batch/fan-out/poll loop, 5s client timeout; re-review before any runtime activation. | None in Feature 2 | FORBIDDEN in Feature 2; no public API/UI display or onward redistribution | Guest ISS market data is approximately 15-minute delayed; `AsOf` is provider trade time; `RetrievedAt` is OpenInvest clock time; required data-quality expectations are exact SECID/BOARDID, reorder-safe required columns, exact decimal LAST, valid trade timestamp, and fail-closed malformed-data handling | None; fail closed | `APPROVED — adapter implementation/test scope only; shipped runtime/public activation forbidden` |

This limited row is sufficient only for the adapter scope frozen here. It must not be cited later as approval
for public display, automated high-frequency collection, caching, persistence, historical ingestion, derived
data, or production source activation.

## 4. Provider/market scope

Feature 2 supports exactly one market path:

```text
engine = stock
market = shares
board  = TQBR
security = canonical OpenInvest ticker / MOEX SECID
```

TQBR is selected as the first supported primary equities board. No automatic board discovery, board-group
selection, fallback board, bonds, ETFs, futures, FX, indices, OTC, or multi-board aggregation is authorized.

Provider-symbol translation is identity-only in Feature 2:

```text
canonical OpenInvest ticker == requested MOEX SECID
```

The adapter must return the exact original canonical ticker in `MarketQuote.Ticker`. It may not trim,
uppercase, lowercase, alias-rewrite, or silently map the application ticker.

## 5. Exact HTTP endpoint and request contract

The planned request is:

```text
GET https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities/{SECID}.json
```

with query parameters equivalent to:

```text
iss.meta=off
iss.only=marketdata,dataversion
marketdata.columns=SECID,BOARDID,LAST,TIME
dataversion.columns=trade_date
```

The implementation must construct the URL structurally; it must not string-concatenate untrusted arbitrary
URLs. The production base host is fixed to `https://iss.moex.com`. Tests may inject an `httptest.Server`
base URL through constructor/configuration intended specifically for tests.

The request sends `Accept: application/json`.

No authentication, API key, cookie, browser automation, scraping, or HTML parsing is authorized.

## 6. HTTP client contract

The adapter must not call package-level `http.Get` or use mutable global HTTP state.

Feature 2 freezes:

- an injected `*http.Client`;
- total client timeout: `5s`;
- caller `context.Context` is attached to every request;
- response body hard limit: `64 KiB`;
- no automatic retry;
- no internal polling loop;
- no concurrency fan-out;
- no cache;
- no background worker;
- no custom transport unless implementation evidence proves it is required.

Caller cancellation is semantically distinct from provider outage:

```text
caller ctx canceled/deadline exceeded
→ preserve caller context error

network/DNS/TLS/client timeout while caller ctx is still live
→ ErrMarketQuoteProviderUnavailable
```

A body larger than 64 KiB is a provider-data/protocol failure, not silently truncated success.

## 7. ISS table/column parsing contract

MOEX ISS JSON uses table blocks with independent `columns` and `data` arrays. Feature 2 must not deserialize
marketdata into a positional struct that assumes a fixed column order.

The adapter must:

1. parse the `marketdata` and `dataversion` blocks independently;
2. build a name → index map from each block's `columns`;
3. require `SECID`, `BOARDID`, `LAST`, and `TIME` from `marketdata`, plus `trade_date` from `dataversion`;
4. reject duplicate required column names in either block;
5. accept reordered columns in either block;
6. ignore additional unrelated columns;
7. validate row length before indexing;
8. reject malformed table shape;
9. require exactly one marketdata row for a successful quote and exactly one usable dataversion row for its trade-date evidence;
10. never expose either provider table/column representation above the adapter package.

A changed column order must remain a successful quote when all required names are present. Empty `marketdata.data` still represents quote absence; missing, empty, multiple, or malformed `dataversion` trade-date evidence is provider-data failure rather than quote-not-found.

## 8. Exact decimal/money contract

`LAST` is the price source for this feature.

No binary floating-point conversion is allowed anywhere in the price path.

The implementation must preserve the JSON numeric token text, for example through `json.RawMessage`, and
parse that text directly through the existing OpenInvest decimal parser:

```text
MOEX JSON numeric token
        ↓
exact textual token bytes
        ↓
decimal.FromString(...)
        ↓
verticalslice.Money{Amount: ..., Currency: RUB}
```

Forbidden:

```go
var price float64
json.Unmarshal(..., &price)
decimal.FromFloat64(price)
```

`null`, non-numeric JSON, non-canonical/out-of-storage decimal, or negative price must never become a
successful `MarketQuote`.

The application-level Stage 3.57 validator remains the final canonical output guard.

## 9. Quote existence semantics

Feature 2 defines quote-not-found at the quote level, not only at the instrument level.

Return `ErrMarketQuoteNotFound` when a structurally valid MOEX response contains no usable quote row for
the requested `SECID`/TQBR, including:

- empty `marketdata.data`;
- a row for which `LAST` is JSON `null`.

Do not fabricate a previous close, zero, bid, offer, midpoint, or current time as a substitute for `LAST`.

A mismatched non-empty `SECID` or `BOARDID` is not quote-not-found; it is provider-data/protocol failure.

HTTP 404 is not automatically interpreted as an unknown security because a route/configuration failure can
also produce 404. Unless implementation evidence proves an official MOEX instrument-404 contract, any
unexpected HTTP 4xx other than 429 is treated as provider/protocol failure.

## 10. AsOf and time semantics

Canonical `MarketQuote.AsOf` represents the time of the last trade corresponding to `LAST`, not the time
OpenInvest performed the request.

Feature 2 uses:

```text
dataversion.trade_date + marketdata.TIME
```

where `marketdata.TIME` is the MOEX last-trade time and `dataversion.trade_date` is the trade date supplied by the same endpoint in the separate `dataversion` block.

Provider local time is interpreted in the IANA zone:

```text
Europe/Moscow
```

and then converted to UTC before entering `MarketQuote`.

The adapter must use `time.LoadLocation("Europe/Moscow")` and must import standard-library `_ "time/tzdata"`
in the provider package so focused provider tests and minimal runtime binaries do not depend on host tzdb
availability. This preserves the repository's existing IANA/tzdata-backed timezone policy.

If `LAST` is present but the required trade date/time cannot be parsed, the adapter returns a provider-data
error rather than fabricating `AsOf`.

## 11. RetrievedAt semantics

`RetrievedAt` is OpenInvest-owned evidence.

The adapter receives the existing `verticalslice.Clock` and must stamp exactly one value from:

```go
clock.Now().UTC()
```

after a successful provider response has been read and parsed sufficiently to materialize the quote.

The adapter must not:

- call `time.Now()` directly;
- use MOEX `SYSTIME` as `RetrievedAt`;
- use HTTP `Date` as `RetrievedAt`;
- copy `AsOf` into `RetrievedAt`.

A zero clock value produces a canonical-invalid quote and therefore fails closed.

## 12. Provider identity

Feature 2 freezes the canonical provenance provider identifier:

```text
MOEX_ISS
```

`MarketDataProvenance.Provider` remains otherwise unchanged. No provider symbol, raw URL, request ID,
response hash, parser version, access mode, or raw response is added to the canonical model in Feature 2.

The limited source-registry code `MOEX_ISS_DELAYED_TQBR` is a governance/source row identifier and is not
required to replace the canonical provider-family provenance value `MOEX_ISS`.

## 13. Provider-neutral error taxonomy

Feature 2 adds at most one new application-owned provider-neutral sentinel if implementation needs it:

```go
ErrMarketQuoteProviderData
```

Frozen classification:

```text
structurally valid response but no usable LAST quote
→ ErrMarketQuoteNotFound

429
5xx
transport/DNS/TLS/client timeout with caller ctx still live
→ ErrMarketQuoteProviderUnavailable

caller cancellation / caller deadline
→ context.Canceled / context.DeadlineExceeded

malformed JSON
missing/duplicate required columns
row shape mismatch
multiple rows
mismatched SECID/BOARDID
invalid/non-decimal/negative LAST
missing/malformed/ambiguous `dataversion` trade-date evidence
invalid `dataversion.trade_date` or `marketdata.TIME`
oversized response
unexpected non-429 4xx unless later official evidence proves not-found semantics
→ ErrMarketQuoteProviderData

adapter somehow returns a structurally mapped MarketQuote that violates the Stage 3.57 canonical validator
→ ErrInvalidMarketQuote at Service boundary
```

Provider-specific DTOs, `*http.Response`, status integers, JSON parsing types, or a MOEX package error type
must not become an application/domain dependency.

Error messages must not include the full raw response body.

## 14. Package ownership and constructor

The provider-specific implementation belongs under:

```text
backend-go/internal/provider/moexiss/
```

The package may import `internal/verticalslice`; `verticalslice` must not import `moexiss`.

Expected implementation ownership:

```text
moexiss/provider.go       HTTP request + QuoteProvider implementation
moexiss/parser.go         narrow ISS table/row parsing helpers if needed
moexiss/provider_test.go  httptest-based contract/failure tests
```

The exact split may remain one production file if that is clearer; SRP is more important than file count.

The adapter constructor must require explicit non-nil HTTP-client and Clock dependencies; nil dependencies
must fail during construction rather than panic on first quote. Conceptually:

```go
NewQuoteProvider(client *http.Client, clock verticalslice.Clock, baseURL string) (*Provider, error)
```

or an equally narrow repository-consistent equivalent.

No generic provider factory/registry/plugin framework is authorized.

## 15. Service dependency injection

Current `verticalslice.Service` already owns an unexported `quoteProvider QuoteProvider`, while
`NewService(store, clock)` intentionally does not configure it.

Feature 2 must add a minimal immutable construction path. Preferred planning choice:

```go
func NewServiceWithQuoteProvider(store Store, clock Clock, provider QuoteProvider) *Service
```

while preserving the existing:

```go
func NewService(store Store, clock Clock) *Service
```

for current call sites/tests.

A mutable setter is not authorized.

A functional-options framework is not required for one dependency and should not be introduced unless the
implementation discovers a concrete second option need.

The new constructor must not bypass Stage 3.57 canonical quote validation.

## 16. Composition-root decision

Feature 2 must make the adapter injectable without activating MOEX in the shipped runtime.

The implementation may add the immutable `NewServiceWithQuoteProvider(...)` construction seam described
above, but `backend-go/cmd/api` remains unchanged and does not construct `moexiss.Provider`.

Reason: reviewed MOEX materials establish free delayed access but do not establish OpenInvest's rights for
automated production/non-display use or onward public redistribution. The active source approval is therefore
limited to adapter implementation/test scope.

Feature 2 must not add an environment flag that silently turns MOEX collection on in `cmd/api`.

A later source-activation stage may wire the adapter only after a fresh legal/terms + technical review freezes:

- allowed runtime use mode;
- attribution requirements;
- provider quota/rate policy;
- redistribution/display rights;
- production failure/fallback policy.

## 17. Public API / frontend / database decision

Feature 2 does not change:

- OpenAPI;
- HTTP asset DTOs;
- `SearchAssets`;
- public asset detail behavior;
- `lastPrice` production behavior;
- SQL/PostgreSQL;
- migrations;
- quote persistence;
- provenance persistence;
- freshness persistence;
- frontend;
- Redis/cache;
- workers/schedulers;
- analytics.

No user-visible field is added.

A later public quote/asset enrichment stage must separately solve redistribution rights, source activation,
freshness threshold, fan-out/batching, caching, partial failure, rate policy, and OpenAPI semantics.

## 18. Required deterministic tests

All normal CI tests use `httptest.Server` or equivalent local deterministic HTTP behavior. CI must not depend
on live MOEX availability.

Success contract:

- 200 valid quote;
- exact canonical ticker preserved;
- `LAST` exact decimal preservation including more than two decimal places;
- RUB Money;
- `dataversion.trade_date + marketdata.TIME` parsed as Europe/Moscow and converted to UTC;
- `RetrievedAt` comes exactly from injected fixed Clock;
- provider identity exactly `MOEX_ISS`;
- reordered columns succeed;
- unrelated extra columns succeed;
- response body safely closed.

Quote absence:

- empty `marketdata.data` → `ErrMarketQuoteNotFound`;
- `LAST: null` → `ErrMarketQuoteNotFound`.

HTTP/transport:

- 429 → provider unavailable;
- 500/503 → provider unavailable;
- unexpected 400/404 → provider-data/protocol failure unless official contract evidence changes planning;
- network failure → provider unavailable;
- client timeout with live caller context → provider unavailable;
- caller context cancellation → context cancellation preserved.

Malformed provider data:

- malformed JSON;
- missing `marketdata`;
- missing `columns`;
- missing `data`;
- missing each required column;
- duplicate required column;
- row shorter than required index;
- multiple rows;
- mismatched SECID;
- mismatched BOARDID;
- invalid JSON numeric LAST;
- negative LAST;
- decimal outside canonical storage;
- missing `dataversion`;
- missing/duplicate `dataversion.trade_date` column;
- empty/multiple `dataversion` rows;
- invalid/missing `dataversion.trade_date`;
- invalid/missing `marketdata.TIME`;
- reordered `dataversion` columns with unrelated extras still succeed;
- oversized response >64 KiB.

Safety/regression:

- no `float64` price path;
- no direct `time.Now()` in moexiss production code;
- no live MOEX request in normal tests/CI;
- existing `NewService(store, clock)` behavior unchanged;
- production `cmd/api` remains provider-free and contains no MOEX activation path;
- asset-search `lastPrice: null` regression remains unchanged;
- no OpenAPI marketData field appears.

## 19. Optional manual smoke evidence

After implementation is published, a human/local smoke check against the real delayed ISS endpoint may be
recorded as supplemental evidence only.

It must never replace deterministic tests or CI and must never cause CI to fail because MOEX is unavailable.

The smoke check must not persist or publish raw market data through OpenInvest and must not be wired into
normal application startup.

## 20. Expected implementation surfaces

A later Feature 2 implementation is expected to remain approximately within:

- `backend-go/internal/provider/moexiss/*.go`
- `backend-go/internal/provider/moexiss/*_test.go`
- `backend-go/internal/verticalslice/marketdata.go` only if the provider-neutral data-error sentinel is needed
- `backend-go/internal/verticalslice/service.go` for the minimal immutable provider constructor
- focused `backend-go/internal/verticalslice/*_test.go` only where constructor/error behavior needs proof
- Stage 3.59/implementation evidence documentation required by the development workflow

No new third-party Go dependency is expected. The implementation should use the standard library
`net/http`, `encoding/json`, `io`, `net/url`, `time`, and the existing OpenInvest decimal package.

If a third-party dependency appears necessary, implementation must stop for renewed scope review rather than
silently adding it.

## 21. Explicit exclusions

Feature 2 planning does not authorize:

- public quote HTTP endpoint;
- asset search/detail quote enrichment;
- public provenance/freshness DTO;
- MOEX real-time subscription;
- public redistribution;
- production source activation;
- provider credentials;
- multiple MOEX boards;
- board discovery/fallback;
- historical prices/candles;
- order book/bid/offer;
- bonds/coupons/dividends;
- corporate actions;
- cache/Redis;
- workers/background polling;
- DB quote storage;
- rate-limit scheduler;
- retries/backoff;
- circuit breaker;
- Kafka/event bus/microservice;
- frontend calendar/heatmap;
- Feature 3.

## 22. Required implementation workflow

After this planning stage itself becomes canonical and the Principal Architect separately authorizes
implementation, Feature 2 follows the mandatory development path:

```text
approved exact implementation scope
→ feature branch
→ contract-first implementation
→ failure cases
→ deterministic tests
→ local gates
→ Internal read-only line-by-line review
→ fixes + gates
→ human commit/push permission
→ Draft PR
→ exact-head required CI
→ fresh External published-head review
→ fixes/CI if needed
→ External verdict
→ Internal evidence-only publication
→ exact-head CI
→ exact evidence-publication verification
→ explicit human Ready/squash-merge authorization
→ protected develop merge
```

The implementation review order is mandatory:

```text
contract
→ implementation
→ failure cases
→ tests
→ CI
→ architectural consequences
```

`UNKNOWN` evidence is not itself a defect. P0/P1/P2/P3 findings require demonstrated impact/risk.

Feature 2 is not closed until the final published exact HEAD is reviewed and approved through the canonical
workflow.

## 23. Stop conditions

Implementation must stop and return for renewed planning if it discovers a requirement for:

- any shipped runtime activation (including development/local `cmd/api` wiring);
- public redistribution or public quote display;
- production source activation;
- a provider quota/rate policy that requires caching, batching or scheduled polling;
- a different/second board;
- symbol alias mapping;
- OpenAPI changes;
- SQL/migrations/persistence;
- background workers;
- retries/circuit breaker;
- new third-party dependencies;
- a canonical provenance schema expansion;
- a universal freshness threshold.

## 24. Planning acceptance criteria

Stage 3.59 planning is ready for publication only if review confirms:

- current protected base identity is exact;
- Feature 1 contract is preserved;
- source-registry prerequisite is satisfied only with a limited non-public approval;
- endpoint/board/columns are explicit;
- money never requires float64;
- `AsOf` and `RetrievedAt` remain distinct;
- HTTP failure taxonomy is provider-neutral;
- columns are name-indexed and reorder-safe;
- tests are deterministic and live-MOEX-independent;
- no public API/DB/frontend/cache/worker scope is smuggled in;
- implementation adds no shipped runtime MOEX activation path;
- legal/redistribution uncertainty is preserved rather than silently treated as permission;
- Feature 3 remains unauthorized.


## 25. Published review evidence

This section is an evidence-only follow-up added after the fresh External published-head verdict. Sections
1–24 remain the reviewed planning contract; this section does not alter implementation scope or semantics.

### Prepublication Internal Planning Review

Frozen Internal Planning Review SHA-256:

```text
488f37d2b35ffbf262e1a7b79d5d63add3a903813779a778d4245ebe22f3d923
```

The complete prepublication review record was:

```text
Stage 3.59 Internal Planning Review
base=2e8d67f980aba5a2d7c4a33d5721ff6a7ad4951a
tree=a73aa0c7c24ee9bdd172a7ffcf95ab07bd533b65
bundle_sha256=c568753f739b6c0f66a4a6514b8ca120006cacb5a3ae3bdb9a53486169ed967e

Files reviewed in full:
1. docs/registries/DATA_SOURCE_REGISTRY.md
2. docs/stages/STAGE_03_59_MOEX_ISS_QUOTE_PROVIDER_PLANNING.md

Review order:
contract -> implementation consequences -> failure cases -> tests -> CI expectations -> architectural consequences

Resolved findings before final verdict:
- P2 RESOLVED: SYSTIME was unnecessarily required/requested despite AsOf being TRADEDATE+TIME. Removed from exact endpoint columns to reduce provider fragility/YAGNI.
- P2 RESOLVED: initial registry/planning wording allowed local/runtime activation despite unresolved MOEX redistribution/non-display terms. Narrowed approval to adapter implementation/test + optional human manual smoke; shipped cmd/api activation forbidden.
- P2 RESOLVED: limited source row did not explicitly preserve attribution/display uncertainty and data-quality expectations required by registry approval. Added NOT ESTABLISHED attribution/display status and exact fail-closed data-quality expectations.
- P3 RESOLVED: provider constructor dependency assumptions were implicit. Planning now requires non-nil HTTP client and Clock and deterministic constructor failure.

Final review:
Architecture / DIP / package direction: PASS
Money / exact decimal semantics: PASS
Time / AsOf / RetrievedAt semantics: PASS
HTTP safety / bounded body / timeout / cancellation: PASS
Provider error normalization: PASS
ISS column-reorder robustness: PASS
Data Source Registry / source-use scope: PASS
Legal/terms uncertainty preservation: PASS
API-first / OpenAPI non-change: PASS
DB/schema/frontend non-change: PASS
Test determinism / no live CI dependency: PASS
Scope / KISS / YAGNI: PASS
Governance / Feature 3 non-authorization: PASS

P0=0
P1=0
P2 blocking=0
P3 blocking=0
Reviewer mutations during final read-only phase=NONE
VERDICT=APPROVED
```

The record above is preserved historically and is not retroactively edited. In particular, its first resolved
P2 assumed `TRADEDATE + TIME` could be sourced from the planned `marketdata` response. The later fresh
External review of published head `6b114478461667cd5e59492d16caa2e9212470bc` found that current TQBR
`marketdata` exposes `TIME` but not `TRADEDATE`; the endpoint exposes `trade_date` in the separate
`dataversion` block. That External finding demonstrates a gap missed by the Internal review; the historical
Internal `APPROVED` verdict is evidence of the prepublication review outcome, not proof that the original
published plan was defect-free.

### External published-head chronology

Initial published head `6b114478461667cd5e59492d16caa2e9212470bc`, tree
`cc78d056cbe80eb0a0e72d7b912e90174c772e89`, passed CI #324 / run `33877065341` 10/10. Fresh External
published-head review then recorded one P2 and `VERDICT = REQUEST CHANGES` in PR review `5113428783`.

The evidence-backed remediation changed only this planning document:

- corrected head `5d2d56182972148ece9a0728093410ae9e8da8e4`;
- corrected tree `831bada8f553d2c826a8c804163b24ca7d63d906`;
- corrected planning SHA-256 `4abd91c1ccf1f6f3b3d47e56fe91f81669f9df6ac87f81a5716a29860108f9d3`;
- unchanged registry SHA-256 `1eb4e31dd8e75c6ceb3dcb4bee8daa66de8812299d2edc48d9cae50aa766d134`;
- corrected two-file manifest SHA-256 `eb303b34980e7080332f9169aec9a7ff18c8735cb6bfeee1679e99fe972caa45`.

Corrected exact-head CI #325 / run `33877740481` passed all ten required jobs. Fresh External re-review on
corrected head `5d2d56182972148ece9a0728093410ae9e8da8e4` recorded `VERDICT = APPROVED`, blocking findings none,
in PR review `5113482198`.

### Evidence-publication rule

This section publishes review evidence only. It introduces no new provider behavior, endpoint, field,
error classification, source permission, runtime activation, API/DB/frontend surface, dependency, cache,
worker, or Feature 3 authorization. The exact evidence-publication head must pass required CI and receive a
final exact-head evidence-only verification before human Ready/squash-merge authorization.
