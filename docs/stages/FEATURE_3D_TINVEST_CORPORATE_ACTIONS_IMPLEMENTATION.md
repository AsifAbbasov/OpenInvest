# Feature 3D — T-Invest Corporate Actions Provider Implementation

| Field | Value |
| --- | --- |
| Status | Frozen prepublication implementation candidate; mandatory Internal review evidence withheld pending External published-head phase |
| Date | 2026-09-09 |
| Repository | `AsifAbbasov/OpenInvest` |
| Canonical implementation base | `develop@e4e676f8a168bf1a2618185ee345cab245d1225d` |
| Protected-base tree | `862265c069105957f802d677cbcb3abbeb5d91b5` |
| Source-rights authority | `docs/registries/DATA_SOURCE_REGISTRY.md` — `TINVEST_CORPORATE_ACTIONS_CONSTRAINED` |
| Source-use evidence | `docs/research/T_INVEST_CONSTRAINED_SOURCE_USE_PROPOSAL.md`; `docs/research/CORPORATE_ACTIONS_SOURCE_OUTREACH.md` |
| Existing domain dependency | Feature 3A — provider-neutral `CorporateActionProvider` / `CorporateActionEvent` |
| Existing projection/API dependency | Features 3B/3C — Calendar/Heatmap and `GET /api/v1/corporate-actions/projection` |
| Runtime activation | NOT ACTIVATED; no live token used |
| Commit / push | HUMAN AUTHORIZATION RECEIVED for the approved frozen feature-branch candidate; not yet performed at document freeze |
| Draft PR | NOT AUTHORIZED / NOT CREATED |
| Stage 3.78 | NOT STARTED / NOT AUTHORIZED by this Feature 3D candidate |

> **Historical implementation evidence**
>
> This document preserves the Feature 3D implementation/review/activation snapshot that was current when it was written.
> Lifecycle, authorization and activation statements below are historical evidence and do not define the current repository state.
> Current product/runtime status: [`../SOURCE_OF_TRUTH.md`](../SOURCE_OF_TRUTH.md).
> Current source/use authority: [`../registries/DATA_SOURCE_REGISTRY.md`](../registries/DATA_SOURCE_REGISTRY.md).
> Post-merge Feature 3D closure: [`FEATURE_3D_TINVEST_CORPORATE_ACTIONS_CLOSURE.md`](FEATURE_3D_TINVEST_CORPORATE_ACTIONS_CLOSURE.md).

## 1. Purpose

Feature 3D implements the smallest production-capable adapter for the exact source/use mode already approved as
`TINVEST_CORPORATE_ACTIONS_CONSTRAINED`. It reuses the existing application-owned Corporate Actions boundary rather
than creating a second framework or exposing T-Invest identifiers in the public domain/API.

The intended runtime chain is:

```text
T-Invest REST InstrumentsService
        ↓
internal/provider/tinvest
        ↓
CorporateActionProvider
        ↓
FetchCorporateActions
        ↓
validated CorporateActionEvent[]
        ↓
existing Calendar / Heatmap projections
        ↓
GET /api/v1/corporate-actions/projection
```

Only two provider operations exist in the implementation:

```text
GetDividends    → canonical DIVIDEND
GetBondCoupons  → canonical COUPON
```

No other T-Invest service or method is introduced.

## 2. Canonical base and drift handling

The original engineering preflight was performed against `develop@1f2fe1cc9c4c6ab282f56506d64a9754b62eef58`.
Before implementation, `develop` advanced by one direct descendant commit to
`e4e676f8a168bf1a2618185ee345cab245d1225d`. That commit changes only Corporate Actions source-research evidence and
adds no Go/runtime/dependency change. Feature 3D therefore uses `e4e676f8...` as the implementation base rather than
building against a stale protected HEAD.

The protected base is rechecked immediately before feature-branch publication. A protected-base drift that changes
any Feature 3D production surface requires comparison/reconciliation before publication; the existing human commit/push
authorization is not permission to publish against an unreviewed drifted base.

## 3. Provider ownership and architecture

The adapter implements the existing application-owned interface:

```go
type CorporateActionProvider interface {
    CorporateActions(
        ctx context.Context,
        query CorporateActionQuery,
    ) ([]CorporateActionEvent, error)
}
```

No T-Invest type crosses that provider boundary. The adapter owns only provider transport, provider identifiers,
provider JSON decoding, traffic/security policy, and normalization into existing OpenInvest canonical events.

There is no:

- second Corporate Actions domain model;
- provider-specific field in OpenAPI;
- provider-specific identifier in the frontend;
- database migration;
- persistence layer;
- provider cache;
- worker/cron/polling system;
- new production dependency.

REST is implemented with the Go standard `net/http` client. No T-Invest SDK, protobuf or gRPC dependency is added.

## 4. Exact instrument mapping and fail-closed discovery boundary

The current canonical asset catalog is intentionally represented by an isolated static provider mapping:

```text
SBER          → SBER_TQBR          → SHARE
GAZP          → GAZP_TQBR          → SHARE
SU26238RMFS4  → SU26238RMFS4_TQOB  → BOND
```

Routing is one approved network method per instrument:

```text
SHARE → GetDividends
BOND  → GetBondCoupons
```

The complete request mapping set is resolved before the first network call. If any requested canonical instrument is
not mapped, the entire provider request fails closed without partial results and without calling an instrument-discovery
endpoint.

`FindInstrument`, `GetInstrumentBy`, `GetBondEvents` and all unrelated methods remain absent from the implementation.
A request accepted by the existing HTTP boundary contains at most 50 instruments, so one request can produce at most
50 T-Invest calls, never 100 calls from dividend-plus-coupon probing.

## 5. HTTP transport hardening

The provider uses:

- production REST base `https://invest-public-api.tbank.ru/rest`;
- `POST` only;
- `Accept: application/json`;
- `Content-Type: application/json`;
- server-side `Authorization: Bearer <token>`;
- forced client timeout of 5 seconds;
- redirects disabled;
- cookie jar disabled;
- response-body limit of 256 KiB;
- caller `context.Context` cancellation;
- no automatic retry.

The supplied caller `http.Client` is copied before hardening so the adapter does not mutate shared caller state.
Redirects are rejected by the provider client and a regression test proves that the Authorization header is not
forwarded to a redirect target.

Provider status classification is deliberately conservative:

```text
401 / 403 / 408 / 429 / 5xx → provider unavailable
other non-200               → provider data invalid
malformed/oversized body     → provider data invalid
caller cancellation/deadline → caller context error
```

For a bounded multi-instrument request, all already-authorized one-call-per-instrument work is allowed to complete so
concurrent completion order cannot change the public failure class. Caller cancellation has priority. Otherwise, any
provider-data-invalid result deterministically maps the batch to provider-data-invalid; if none exists, any unavailable
result maps the batch to provider-unavailable. There is no retry or extra discovery/fallback call.

Response bodies and token values are never included in adapter errors.

## 6. Traffic policy

The canonical registry records an official InstrumentsService allowance of 200 requests/minute while OpenInvest owns a
stricter internal budget:

```text
maximum provider requests/minute = 60
maximum concurrent provider calls = 4
automatic retry                    = NO
background polling                 = NO
```

The implementation enforces both limits inside the provider instance:

- a provider-global semaphore bounds simultaneous outbound calls to four, including concurrent OpenInvest requests; callers beyond four fail closed immediately instead of forming an unbounded waiter queue;
- a rolling one-minute request budget rejects the 61st admitted provider call rather than waiting, retrying or falling
  back to another source;
- when T-Invest returns `x-ratelimit-limit`, `x-ratelimit-remaining` and `x-ratelimit-reset`, the adapter folds the
  provider's remaining/reset evidence into subsequent admission and never uses those headers to raise the stricter
  OpenInvest 60/minute budget;
- stale/out-of-order concurrent responses may only make admission more conservative: a later-observed higher remaining
  value cannot increase an already known lower allowance;
- provider reset information expires after the advertised reset interval; no retry or wait loop is created.

The rate/concurrency state contains no provider payload and is operational control state, not provider-data caching.

## 7. Request-lifetime / no-persistence boundary

Provider JSON exists only for the lifetime of the current request:

```text
HTTP response bytes
→ strict decode
→ normalization
→ canonical event validation/projection
→ public response
→ provider object becomes unreachable
```

No provider response is written to PostgreSQL, Redis, disk, logs, background queues, historical archives or a
cross-request data cache. This preserves the exact constrained registry storage policy.

## 8. Dividend normalization

`GetDividends` is normalized as:

```text
Dividend.dividend_net  → AmountPerUnit
Dividend.record_date   → RecordDate
Dividend.payment_date  → PaymentDate
Dividend.dividend_type → lifecycle/type guard
kind                   → DIVIDEND
```

When record/payment timestamps are present, they must be valid UTC protobuf JSON timestamps. If either source field is
absent, the canonical field remains `nil`; no date is fabricated. OpenInvest stores only the canonical business date
`YYYY-MM-DD` for present record/payment semantics.

The current canonical `DIVIDEND` scope is deliberately narrower than every value T-Invest can expose through
`dividend_type`:

```text
Regular Cash → DIVIDEND / ANNOUNCED
Cancelled    → DIVIDEND / CANCELLED
other types  → provider-data-invalid / fail closed
```

`Return of Capital`, `Daily Accrual` and unknown future provider types are not silently converted into an ordinary cash
dividend because the current OpenInvest domain has no separately approved kind for those semantics. A past payment date
is still not treated as proof that cash was actually paid, and no provider schedule is promoted to ledger truth.

## 9. Coupon normalization

`GetBondCoupons` is normalized as:

```text
Coupon.coupon_date  → PaymentDate
Coupon.fix_date     → RecordDate when present
Coupon.pay_one_bond → AmountPerUnit
kind                → COUPON
status              → ANNOUNCED
```

If `fix_date`, `coupon_date` or `pay_one_bond` is absent, the corresponding canonical field remains `nil`. No zero date,
payment-date substitution, zero amount or other invented source fact is created.

`coupon_number` must be a positive provider identity and participates in deterministic event identity so two distinct
coupon schedule rows cannot collapse only because their payment amount/date happens to match.

## 10. Exact MoneyValue semantics

T-Invest `MoneyValue` represents:

```text
units + nano / 1_000_000_000
```

The adapter never converts this value through `float64`.

OpenInvest canonical Decimal uses scale 8. The adapter therefore computes the exact integer number of nanounits with
`math/big.Int` and accepts a value only when it is exactly representable at scale 8. A non-zero ninth decimal digit is
rejected as provider-data-invalid rather than silently rounded. Provider `units` and `nano` with contradictory signs are
also rejected rather than arithmetically reinterpreted into a different amount.

An absent `MoneyValue` remains canonical `nil`. When present, the resulting decimal must satisfy the existing OpenInvest
storage-fit contract. Negative Corporate Actions amounts are rejected. Currency is normalized to a three-letter uppercase
canonical code.

This policy deliberately prefers source rejection over alteration of financial truth.

## 11. Lifecycle / financial-truth boundary

Feature 3D does not infer factual settlement from schedule dates. Lifecycle is conservative and evidence-driven:

```text
Regular Cash dividend → ANNOUNCED
explicit Cancelled dividend → CANCELLED
coupon schedule → ANNOUNCED
```

It never creates `PAID` or `CONFIRMED` from wall-clock comparisons, payment dates or guesses. `CANCELLED` is emitted
only when the provider explicitly identifies the dividend as cancelled; unsupported provider dividend types fail closed
instead of being reclassified.

Provider events remain external reference truth. They do not automatically create or modify:

- `DIVIDEND` ledger transactions;
- `COUPON` ledger transactions;
- cash balances;
- realized income;
- tax transactions.

## 12. Deterministic identity and provenance

`EventID` and internal `SourceEventID` are derived deterministically from the normalized provider fact with SHA-256.
Dividend identity includes `dividend_type`; coupon identity includes `coupon_number`. Retrieval time is not part of
identity, so repeated retrieval of the same provider fact does not produce a new event identity.

Exact duplicate rows within one provider response are deduplicated by deterministic `EventID`.

Canonical provider provenance is:

```text
T_INVEST_API
```

The existing public Corporate Actions DTO exposes only the canonical provider identifier; provider-owned
`SourceEventID` remains internal and does not expand the public OpenAPI contract.

## 13. Runtime activation and secret boundary

Runtime composition is fail-closed and requires both:

```text
OPENINVEST_TINVEST_CORPORATE_ACTIONS_ENABLED=true
OPENINVEST_TINVEST_READONLY_TOKEN=<non-empty secret>
```

The token alone cannot activate the provider. The enable flag without a valid token fails application construction.
When the enable flag is absent/false, the provider remains nil and no T-Invest request can occur.

The implementation cannot independently prove the external token's broker-side permission scope without performing
additional account/capability calls that are outside the authorized T-Invest surface. Therefore **read-only token
scope is an explicit deployment/provisioning contract**: operations must provision a non-trading token and separate
credentials for development and production. The code itself exposes only the two approved InstrumentsService calls.

The token is never placed in Git, frontend/mobile code, OpenAPI, logs, metrics, traces or returned errors. `.env.example`
contains only an empty placeholder and the security requirements.

No live token was supplied or used while building this candidate, and runtime activation remains OFF.

## 14. Shipped route remediation

Before Feature 3D, the ordinary route registry contained:

```text
GET /api/v1/corporate-actions/projection
```

but the canonical shipped composition used `NewReplay → newReplayApp`, whose explicit route list omitted that endpoint.
A real provider would therefore have been unreachable from the shipped API.

Feature 3D remediates the actual production constructor rather than testing only a synthetic router:

- legacy `NewReplay(...)` remains source-compatible and delegates with a nil Corporate Actions provider;
- `NewReplayWithCorporateActionProvider(...)` explicitly injects the provider-neutral dependency;
- development construction has the equivalent explicit provider-aware constructor;
- `newReplayApp(...)` always registers the Corporate Actions route;
- disabled/nil provider therefore returns the established fail-closed 503 source-unavailable response instead of 404;
- provider-injected construction reaches the existing handler successfully.

This changes no public URL or OpenAPI schema; it makes the already-canonical route reachable in the shipped composition.

## 15. Verification performed locally

The execution environment does not contain the repository-declared Go `1.25.14` toolchain or a complete clone/module
cache. It has Go `1.23.2`, and direct repository cloning is unavailable from the container. Full repository gates are
therefore **NOT CLAIMED AS PASSED**.

To obtain meaningful pre-commit evidence without fabricating full-repository results, the exact Feature 3D provider
sources were compiled in an isolated Go harness containing only interface/Decimal stubs matching the canonical
OpenInvest contracts required by the adapter.

Available checks currently pass:

```text
provider: go test -race ./internal/provider/tinvest   PASS
provider: go vet ./internal/provider/tinvest          PASS
runtime gate: go test -race ./cmd/api                 PASS
runtime gate: go vet ./cmd/api                        PASS
all candidate Go files: go/parser syntax parse        PASS
forbidden T-Invest production-surface scan            PASS
production retry/polling/persistence mechanism scan   PASS
```

Provider tests cover, at minimum:

- exact stock/bond method routing and static provider identifiers;
- one approved call per requested canonical instrument;
- unknown-instrument fail-closed behavior before network I/O;
- exact scale-8 money conversion and rejection of values requiring rounding;
- absent dividend/coupon date or amount fields remain canonical `nil`;
- positive `couponNumber` enforcement;
- mixed-sign `MoneyValue.units` / `nano` rejection;
- `Regular Cash` dividends remain `ANNOUNCED`;
- explicit `Cancelled` dividends map to canonical `CANCELLED`;
- unsupported dividend types such as `Return of Capital` / `Daily Accrual` fail closed;
- past payment dates never infer `PAID`;
- retrieval timestamps are sampled only after the provider response is received/decoded for a non-empty event batch;
- deterministic IDs and duplicate-row collapse;
- HTTP status/error classification plus deterministic multi-instrument error-class precedence independent of completion order;
- no automatic retry;
- protobuf-JSON empty-result compatibility for omitted/null repeated fields;
- malformed, trailing, top-level-null/array and oversized JSON response rejection;
- redirect rejection and no Authorization forwarding;
- token/provider response non-leakage through errors;
- provider-global concurrency cap of four with fail-fast overflow and no waiter queue;
- 60/minute provider request budget;
- provider `x-ratelimit-*` remaining/reset enforcement without retry and without stale-header allowance increases;
- caller cancellation;
- constructor timeout/cookie/redirect hardening without caller-client mutation.

Runtime-gate tests prove disabled-by-default behavior, explicit token requirement, and the requirement for both activation
inputs. Replay-constructor regression tests are included in the candidate but cannot be honestly reported as executed
against the complete repository until the canonical Go toolchain/dependencies are available.

## 16. Review-size / changed-file scope

The local candidate contains 10 repository paths, below the default 25-file review ceiling:

1. `.env.example`
2. `backend-go/cmd/api/main.go`
3. `backend-go/cmd/api/tinvest_runtime.go`
4. `backend-go/cmd/api/tinvest_runtime_test.go`
5. `backend-go/internal/httpapi/replay_app.go`
6. `backend-go/internal/httpapi/replay_app_corporateactions_test.go`
7. `backend-go/internal/provider/tinvest/provider.go`
8. `backend-go/internal/provider/tinvest/provider_test.go`
9. `backend-go/internal/provider/tinvest/types.go`
10. `docs/stages/FEATURE_3D_TINVEST_CORPORATE_ACTIONS_IMPLEMENTATION.md`

The 800-line workflow budget applies to hand-written business logic rather than documentation, tests, configuration or
wiring. For this frozen candidate the exact provider/business-logic paths are counted as non-blank, non-comment source
lines: `provider.go = 687`, `types.go = 67`, total **754**. Runtime/config composition is reported separately and is not
misrepresented as business logic. Any material edit invalidates this count and requires recomputation.

No migration, OpenAPI, frontend, dependency, CI-workflow or protected-registry file is changed by this candidate.

## 17. Out of scope / explicit NO-GO

Feature 3D does not authorize or implement:

- `GetBondEvents`;
- instrument discovery/lookups;
- MarketDataService;
- OrdersService;
- OperationsService;
- account or portfolio synchronization;
- trading;
- market prices;
- forecasts;
- background polling or synchronization;
- raw provider persistence/archive/cache;
- provider fallback or scraping;
- new CorporateActionEvent kinds;
- public provider-specific identifiers;
- automatic ledger mutation;
- new paid data-source spend;
- Stage 3.78 work.

## 18. Rollback / kill switch

Operational rollback is immediate and code-independent:

```text
OPENINVEST_TINVEST_CORPORATE_ACTIONS_ENABLED=false
```

or remove the enable flag. This returns composition to the existing provider-absent fail-closed behavior while keeping
the canonical Corporate Actions route present.

Removing/rotating the token without disabling an enabled provider intentionally causes startup/configuration failure
rather than silently running with an incomplete security configuration.

## 19. Mandatory next gates

This document does not claim Builder self-approval. Feature 3D is a development-path change under
`docs/REVIEW_WORKFLOW.md` v1.4.0. The prepublication candidate is frozen only after local evidence and a complete
read-only Internal review. Detailed Internal verdict/findings remain **WITHHELD — external published-head phase pending**
and must not be published into the repository/PR evidence surface before the fresh External verdict.

The human has already authorized commit and push of the approved frozen feature-branch candidate. That authorization is
held behind the technical gates and does not authorize a Draft PR, Ready transition, merge, protected-branch mutation,
runtime activation or Stage 3.78. After feature-branch publication, the remaining sequence is:

```text
feature-branch publication
→ Draft PR only after separate authorization
→ protected GitHub CI
→ fresh External published-head review
→ required Internal evidence publication / exact-head verification
→ separate human Ready / merge authorizations
```

Prepublication freeze state recorded by this document:

```text
PRODUCTION IMPLEMENTATION CANDIDATE = FROZEN FOR FINAL INTERNAL REVIEW
FULL REPOSITORY GATES                = NOT EXECUTED IN CANONICAL TOOLCHAIN
INTERNAL REVIEW EVIDENCE             = WITHHELD — EXTERNAL PUBLISHED-HEAD PHASE PENDING
COMMIT / PUSH AUTHORIZATION           = RECEIVED; HELD UNTIL FINAL INTERNAL APPROVAL
DRAFT PR                              = NOT AUTHORIZED / NOT CREATED
RUNTIME ACTIVATION                    = NO
LIVE T-INVEST TOKEN                   = NOT USED
STAGE 3.78                            = NOT STARTED
```
