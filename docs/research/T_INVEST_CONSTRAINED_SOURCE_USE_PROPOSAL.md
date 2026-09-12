# T-Invest Constrained Corporate Actions Source / Use Proposal

| Field | Value |
| --- | --- |
| Status | DRAFT / CONDITIONAL-GO PROPOSAL — NOT AUTHORIZED FOR PUBLIC PRODUCTION |
| Date | 2026-09-09 |
| Source candidate | T-Invest API |
| Proposed source identifier | `TINVEST_CORPORATE_ACTIONS_CONSTRAINED` |
| Canonical base | `develop@bcf2b43664066f3e95d402f66f07c47295a2ba8e` |
| Evidence source | `docs/research/CORPORATE_ACTIONS_SOURCE_OUTREACH.md` |
| Registry authority | `docs/registries/DATA_SOURCE_REGISTRY.md` |
| Product scope | Future Feature 3D — Real Corporate Actions Source Adapter |
| Runtime activation | NOT AUTHORIZED BY THIS DOCUMENT |

## 1. Purpose

This document proposes one narrowly bounded T-Invest source/use mode that could be reviewed for future OpenInvest Corporate Actions integration.

It does **not** activate a provider, modify runtime wiring, create credentials, change OpenAPI, change the database, or modify the `Data Source Registry`.

The purpose is to convert the current evidence into an exact decision candidate with explicit restrictions rather than treating a generic API permission as an unrestricted production license.

## 2. Evidence basis

### 2.1 Provider-specific support evidence

OpenInvest described the following exact scenario to T-Invest support:

- OpenInvest is a free public research/educational application;
- the OpenInvest backend obtains T-Invest API data;
- provider data are normalized into OpenInvest's own domain model;
- users do not receive the raw T-Invest API/feed or the T-Invest token;
- the public UI displays normalized dividend/coupon/event data;
- the public UI may display derived analytics/calculations based on those normalized data.

T-Invest support answered ticket `3-781291` for that scenario with the provider statement:

> `Такое использование разрешено.`

> English translation for readability: `This use is permitted.`
> The original Russian provider wording above is preserved as the primary evidence.

An earlier support response also stated, in substance, that the requested data are available, use is permitted, and no additional agreement is required.

This is strong scoped evidence for the described OpenInvest use mode.

### 2.2 Published documentation conflict

The official T-Invest API FAQ currently states that a public service based on T-Invest API is not permitted and that T-Invest API is provided to T-Invest clients without redistribution rights.

The same FAQ states that the legal side of API use is governed by the user agreement.

Official FAQ:

`https://developer.tbank.ru/invest/intro/faq`

Because the support clarification specifically addressed the OpenInvest normalization/public-display/derived-analytics scenario, this proposal treats the support answer as scoped provider evidence for that exact scenario, **not** as permission for raw-feed retransmission or any broader use.

The textual FAQ conflict remains a governance/legal-risk note and must remain visible in the registry decision.

## 3. Technical capability evidence

Official T-Invest API documentation confirms the following relevant methods:

### `GetDividends`

Purpose: dividend-payment events for an instrument.

Filtering is performed by `record_date` for the requested UTC period.

Official documentation:

`https://developer.tbank.ru/invest/api/instruments-service-get-dividends`

### `GetBondCoupons`

Purpose: bond coupon payment schedule.

Filtering is performed by `coupon_date` for the requested UTC period.

Official documentation:

`https://developer.tbank.ru/invest/api/instruments-service-get-bond-coupons`

### `GetBondEvents`

Technical capability includes:

- coupon events;
- offers/calls;
- maturities/redemptions;
- conversions.

Official documentation:

`https://developer.tbank.ru/invest/api/instruments-service-get-bond-events`

However, the current canonical OpenInvest `CorporateActionEvent` supports only:

- `DIVIDEND`;
- `COUPON`.

Therefore `GetBondEvents` is **not authorized for production mapping by this proposal** except as separately reviewed research/compatibility evidence. Offers, maturities/redemptions and conversions require a future domain/API/product contract extension.

## 4. Proposed exact production use mode

The constrained candidate is:

```text
T-Invest InstrumentsService
        ↓
server-side read-only adapter
        ↓
strict normalization
        ↓
DIVIDEND / COUPON only
        ↓
provider-neutral CorporateActionEvent
        ↓
existing Calendar / Heatmap projections
        ↓
existing OpenInvest API / UI
```

### Allowed source methods

Production candidate scope:

```text
GetDividends
GetBondCoupons
```

Research-only / disabled in shipped mapping until separately approved:

```text
GetBondEvents
```

No MarketDataService, OrdersService, OperationsService, portfolio/account-data methods, trading methods, news, forecasts or unrelated T-Invest endpoints are authorized by this proposal.

## 5. Data minimization and normalization boundary

The adapter must consume only the provider fields required to construct the current canonical OpenInvest event types.

Rules:

1. never expose the raw T-Invest response to frontend/mobile clients;
2. never expose provider protobuf/REST objects through OpenInvest public contracts;
3. discard unused provider fields after request processing;
4. preserve only normalized canonical event fields required by the existing OpenInvest contract;
5. record provenance as provider/source metadata without exposing credentials or raw payloads;
6. derived analytics must operate only on normalized OpenInvest domain objects.

Raw-feed resale, raw API forwarding, proxying, bulk export of provider responses and general-purpose third-party API access are explicitly outside this use mode.

## 6. Caching / retention decision

Provider-specific caching/retention rights are not yet explicitly established in the reviewed evidence.

Therefore this proposal is fail-closed:

```text
persistent provider payload storage = FORBIDDEN
provider raw-response cache          = FORBIDDEN
database persistence of T-Invest raw data = FORBIDDEN
Redis/provider-result cache          = FORBIDDEN
historical provider archive          = FORBIDDEN
```

Permitted technical processing is limited to request lifetime:

```text
network response
→ validate
→ normalize
→ calculate existing projections
→ return normalized OpenInvest result
→ release provider response from request memory
```

Request-scoped deduplication/memoization is permitted only within one in-flight application request and must not survive completion of that request.

If future product requirements need persistent Corporate Actions history, background synchronization or cross-request caching, OpenInvest must obtain additional evidence for retention/caching rights and approve a new source/use mode before implementation.

## 7. Attribution / provenance decision

No provider-specific attribution requirement has yet been established in the reviewed evidence.

OpenInvest will nevertheless use conservative plain-text provenance for transparency:

```text
Source: T-Invest API
```

Rules:

- plain text only by default;
- no T-Bank/T-Invest logo use is authorized by this proposal;
- no endorsement, partnership or affiliation claim;
- provenance must distinguish T-Invest source data from OpenInvest calculations;
- derived metrics must be labeled as OpenInvest calculations where applicable.

This voluntary provenance rule does not claim that T-Invest legally requires attribution.

## 8. Authentication and secret handling

Official T-Invest API documentation requires a Bearer API Token.

Official token documentation:

`https://developer.tbank.ru/invest/intro/intro/token`

The constrained adapter must use a **read-only token** and server-side secret handling only.

Mandatory rules:

1. token exists only in backend runtime secret/environment configuration;
2. token is never committed to Git;
3. token is never sent to browser, Next.js client bundle, mobile client or user;
4. token is never logged;
5. token is never included in traces, metrics, error payloads or audit-event attributes;
6. request headers containing authorization must be redacted from observability tooling;
7. no order/trading/write-capable token is allowed;
8. rotation must be possible without code changes;
9. production and development tokens must be separate;
10. test fixtures must never contain a real token.

The adapter must not call account-specific methods or expose any brokerage-account information.

## 9. Environment boundary

Official endpoints:

- production: `invest-public-api.tbank.ru:443`;
- sandbox: `sandbox-invest-public-api.tbank.ru:443`.

Official protocol documentation:

`https://developer.tbank.ru/invest/intro/developer/protocols/`

Non-public development/staging evaluation may use sandbox where the relevant InstrumentsService behavior is available.

A production provider endpoint must remain disabled until the exact source/use mode receives registry approval and separately reviewed runtime wiring is merged.

## 10. Rate-limit / traffic policy

Current official limit documentation states:

- InstrumentsService unary limit: 200 requests/minute;
- the service limit is shared across methods in that service;
- T-Invest recommends keeping aggregate traffic across accounts/tokens from one address below 50 requests/second.

Official limits:

`https://developer.tbank.ru/invest/intro/intro/limits`

OpenInvest must use a substantially more conservative policy:

```text
Corporate Actions adapter global budget: <= 60 provider requests/minute
Maximum concurrent provider requests:     4
Automatic retry on 429:                    NO
Unbounded fan-out:                         FORBIDDEN
Background polling:                        FORBIDDEN under this proposal
```

The adapter must respect provider rate-limit response headers where available and fail closed on throttling.

A user request that would exceed the OpenInvest internal budget must return partial/unavailable provider status according to the separately reviewed Feature 3D product contract rather than bypassing the limit.

## 11. Network and resilience policy

Minimum production candidate rules:

- server-side client only;
- TLS only;
- bounded request timeout;
- no infinite retry;
- no automatic retry for authorization/validation errors;
- no retry storm on `429` or `5xx`;
- bounded concurrency;
- deterministic provider error mapping;
- provider outage must not corrupt canonical portfolio/ledger state;
- no provider fallback to scraping or an unapproved source;
- failure must be observable without logging secrets or raw private metadata.

Recommended initial timeout:

```text
5 seconds per provider request
```

## 12. Product/domain scope

The constrained source may provide only the external event truth necessary for the already-authorized canonical kinds:

### Dividend

```text
GetDividends → DIVIDEND
```

### Coupon

```text
GetBondCoupons → COUPON
```

Not authorized in this proposal:

- redemption/maturity mapping;
- amortization mapping;
- offer/call mapping;
- conversion mapping;
- split mapping;
- merger/reorganization mapping;
- market prices;
- valuation;
- automatic transaction creation;
- tax calculations based directly on provider payloads;
- provider-specific fields in public OpenAPI.

## 13. Financial-truth boundary

T-Invest Corporate Actions data must remain external reference/event data.

It must **not** mutate the immutable portfolio ledger automatically.

Provider events may support:

- calendar display;
- heatmap display;
- expected-income projections explicitly labeled as provider-derived/expected;
- existing derived analytics within the approved product contract.

They must not silently create realized DIVIDEND/COUPON ledger entries, cash balances, taxes or realized income.

Actual portfolio ledger truth remains user-owned/manual or separately authorized transaction-import truth.

## 14. Data-quality boundary

Before normalization, the adapter must validate at minimum:

- required instrument identifier;
- event type matches the called method;
- dates parse and satisfy canonical BusinessDate rules;
- money/currency fields are valid before exact Decimal conversion;
- impossible/unsupported values fail closed;
- duplicate provider events do not create duplicate canonical events within one response;
- canonical sorting is deterministic.

Provider data errors are known to be possible; T-Invest support documentation itself provides a dedicated route for reporting missing/incorrect instrument data including coupons and dividends.

Support documentation:

`https://developer.tbank.ru/invest/support/`

OpenInvest must never silently repair provider facts by guessing.

## 15. Proposed observability

Allowed operational metadata:

- provider code;
- method name;
- success/failure class;
- HTTP/gRPC status category;
- duration;
- request count;
- throttling count;
- normalized event count;
- T-Invest `x-tracking-id` when available for support diagnostics, subject to log-retention policy.

Forbidden:

- Bearer token;
- authorization header;
- raw response body;
- raw request body when it contains unnecessary provider/private metadata;
- brokerage-account identifiers.

## 16. Proposed source/use decision matrix

| Dimension | Proposed state | Basis / constraint |
| --- | --- | --- |
| Official API exists | YES | T-Invest official API documentation |
| Dividend coverage | YES | `GetDividends` |
| Coupon coverage | YES | `GetBondCoupons` |
| Automated retrieval | YES | official API + support evidence |
| Normalized public display | CONDITIONAL YES | exact OpenInvest scenario confirmed by support |
| Derived analytics | CONDITIONAL YES | exact OpenInvest scenario confirmed by support |
| Raw redistribution | NO | outside requested/confirmed scenario; FAQ conflict |
| Persistent raw storage | NO | retention right not established |
| Cross-request cache | NO | caching right not established |
| Request-lifetime processing | YES | necessary constrained processing |
| Attribution | OPENINVEST CONSERVATIVE PROVENANCE | plain-text source disclosure; no claim of legal requirement |
| API access cost | ZERO | official documentation states API data are free |
| Token | REQUIRED | Bearer API Token; read-only only |
| Rate limits | DEFINED | InstrumentsService 200/min official; OpenInvest internal <=60/min |
| Production public use | PROPOSAL ONLY | requires registry decision + runtime review |
| FAQ conflict | OPEN / EXPLICIT | support-specific permission conflicts with generic FAQ text |

## 17. Proposed registry row — NOT YET APPLIED

If Principal Architect / source-rights review accepts this exact constrained mode, the following is the intended **candidate** registry shape:

```text
Source:
TINVEST_CORPORATE_ACTIONS_CONSTRAINED

Status:
CONDITIONAL-GO — normalized DIVIDEND/COUPON public display and derived analytics only;
no raw redistribution; no persistent provider storage; no cross-request cache;
read-only server-side token; bounded on-demand traffic; exact provider provenance;
FAQ conflict retained as reviewed risk/evidence note.
```

This text is a proposal only. `docs/registries/DATA_SOURCE_REGISTRY.md` is unchanged by this document.

## 18. Preconditions before runtime implementation

All must be true:

1. this proposal is reviewed line-by-line;
2. exact-head CI is green;
3. Principal Architect accepts or rejects the FAQ/support-evidence conflict for this constrained mode;
4. a separate registry PR creates the exact approved row if accepted;
5. implementation scope is separately authorized;
6. secret-handling design is verified;
7. no persistence/cache/background polling is introduced accidentally;
8. public OpenAPI remains provider-neutral unless a separately approved contract change is necessary;
9. frontend provenance UX is explicitly defined;
10. adapter tests use fixtures only and contain no real token/data dump.

## 19. Explicit non-goals

This proposal does not authorize:

- runtime provider wiring;
- production token creation/storage;
- database migrations;
- provider-event persistence;
- Redis/provider cache;
- cron/background synchronization;
- raw provider-data export;
- frontend direct calls to T-Invest;
- user-supplied T-Invest tokens;
- account/portfolio synchronization from T-Invest;
- trading/order functionality;
- market-price ingestion;
- Feature 3.78+;
- changes to existing source registry rows.

## 20. Review verdict requested

The requested decision is narrow:

```text
Should OpenInvest accept
TINVEST_CORPORATE_ACTIONS_CONSTRAINED
as a registry CONDITIONAL-GO candidate
for DIVIDEND/COUPON on-demand normalized public display + derived analytics,
under the no-persistence/no-cross-request-cache/read-only-token restrictions above?
```

Until that decision is separately made:

```text
Public production activation = NO
Non-public technical evaluation = CONDITIONAL YES
Data Source Registry transition = NONE
Feature 3D runtime implementation = NOT AUTHORIZED
```
