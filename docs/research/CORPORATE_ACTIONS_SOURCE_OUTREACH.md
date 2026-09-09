# Corporate Actions Source Outreach

| Field | Value |
| --- | --- |
| Status | Active source-rights / provider evidence log |
| Last evidence update | 2026-09-09 |
| Canonical repository base at reconciliation | `develop@1f2fe1cc9c4c6ab282f56506d64a9754b62eef58` |
| Governance authority | `docs/REVIEW_WORKFLOW.md` v1.4.0; `docs/registries/DATA_SOURCE_REGISTRY.md` |
| Product boundary | Existing provider-neutral Corporate Actions Calendar / Heatmap; canonical kinds `DIVIDEND` and `COUPON` only |
| Runtime activation | No provider is activated by this research document |
| Project external-data budget | `0 RUB` unless the owner explicitly changes the project constraint |

## 1. Purpose

This document is the canonical evidence log for external Corporate Actions source selection in OpenInvest.

It records:

- technical coverage evidence;
- public-display and derived-analytics rights;
- caching / retention evidence;
- attribution and traffic constraints;
- contract / pricing evidence;
- provider outreach status;
- the relationship between research evidence and the `Data Source Registry`.

Evidence in this file does not by itself activate a provider, configure credentials, authorize spending, create a new public API contract, permit scraping, or mutate portfolio ledger truth.

The canonical application chain remains:

```text
approved external source/use mode
        ↓
provider adapter
        ↓
CorporateActionProvider
        ↓
validated CorporateActionEvent[]
        ↓
Calendar / Heatmap
        ↓
existing HTTP / OpenAPI / frontend surface
```

Canonical Corporate Actions remain limited to:

- `DIVIDEND`;
- `COUPON`.

Redemptions, amortizations, offers, calls, conversions and other event types require a separately approved domain/API extension even if a provider can supply them.

## 2. Decision rules

OpenInvest treats the following as separate questions:

1. Does an API/feed technically exist?
2. Does it cover the instruments/events OpenInvest needs?
3. May OpenInvest retrieve it automatically?
4. May OpenInvest publicly display normalized facts?
5. May OpenInvest build and display derived analytics?
6. May normalized/provider data be cached or retained?
7. What attribution is required?
8. What rate / traffic limits apply?
9. What contract or account dependency exists?
10. What is the production cost?
11. Does the source fit the project's `0 RUB` external-data budget?

Canonical rules:

```text
API existence       != production approval
technical access    != public-display rights
public webpage      != scraping permission
free access         != redistribution permission
support statement   != unlimited permission
provider rights     != automatic runtime activation
provider rights     != authorization to spend money
```

Unknown or conflicting material evidence remains fail-closed.

## 3. OpenInvest use model communicated to providers

OpenInvest is described as:

- an independent research / educational investment analytics product;
- focused on retail/private investors;
- currently focused on MOEX / Russian-market portfolios;
- using server-side provider retrieval;
- normalizing provider facts into an OpenInvest-owned provider-neutral model;
- never exposing API keys/tokens to users;
- not reselling a raw provider API/feed;
- potentially displaying normalized Corporate Actions facts publicly;
- potentially calculating calendars, expected cash flows, yields and other OpenInvest-derived analytics.

Repository:

`https://github.com/AsifAbbasov/OpenInvest`

## 4. Current provider register

| Organization | Category | Current evidence status | Zero-budget production state |
| --- | --- | --- | --- |
| T-Invest | Broker / API | Exact OpenInvest normalized public-display + derived-analytics scenario confirmed by support; constrained registry row merged; published FAQ conflict preserved | `CONDITIONAL-GO` for exact registry scope only; runtime separately gated |
| Interfax / e-disclosure | Professional disclosure/API vendor | Public normalized use / derived analytics / retention confirmed for described contracted use; minimum contract 3 months; archive option +50%; fresh exact quote requested | `NO-GO` while price is above `0 RUB`; no spending authorized |
| Financial Data API / fdnpy | Data API vendor candidate | Broad coverage claimed; exact MOEX/Russian coverage and source-rights clarification sent | NOT ESTABLISHED |
| Cbonds / InvestFunds | Professional data vendor | Trial available; retransmission/public-display/retention/pricing clarification sent | NOT ESTABLISHED / likely commercial |
| NSD / НРД | Market infrastructure | Relevant services exist; provider response requires legal-entity contracting for current route | `NO-GO` under current project status |
| Finam | Broker / API | Routed to Trade API team | AWAITING SUBSTANTIVE RESPONSE |
| VTB | Broker / bank | Formal paper request route required | ELECTRONIC ROUTE BLOCKED |
| MOEX | Exchange | Outreach sent; separate reviewed market-data source-rights NO-GO remains authoritative | AWAITING RESPONSE; no rights inferred from silence |
| Alfa Investments | Broker / bank | Outreach sent | AWAITING RESPONSE |
| BCS | Broker | Outreach sent | AWAITING RESPONSE |
| Gazprombank | Broker / bank | Outreach sent | AWAITING RESPONSE |
| Sber | Broker / bank | Outreach sent | AWAITING RESPONSE |
| SPB Exchange | Exchange | Outreach sent | AWAITING RESPONSE |
| Finmarket | Financial information | Outreach sent | AWAITING RESPONSE |
| Smart-Lab | Investment community | Research contact | AWAITING RESPONSE |
| Banki.ru | Financial platform | Research contact | AWAITING RESPONSE |
| RBC Investments | Financial media | Research contact | AWAITING RESPONSE |
| Bank of Russia / CBR | Regulator | Formal online-reception route identified; not contacted by email for this wave | NO APPROVED FEED ESTABLISHED |

Functional contact routes may be recorded where useful, but Gmail message identifiers, internal headers, tracking metadata and unnecessary private correspondence must not be copied into this public repository.

## 5. Provider evidence

### 5.1 T-Invest — constrained registry-approved source/use mode

T-Invest support ticket `3-781291` addressed the exact OpenInvest scenario:

- backend retrieval;
- normalization into OpenInvest domain objects;
- no raw API/feed/token exposure to users;
- normalized public dividend/coupon display;
- OpenInvest-derived analytics/calculations.

Provider support answered that exact described use with:

> `Такое использование разрешено.`

Official technical evidence established:

- `GetDividends`;
- `GetBondCoupons`;
- `GetBondEvents` technically exists but is not authorized for the current OpenInvest production source/use mode;
- Bearer token authentication;
- production/sandbox endpoints;
- InstrumentsService traffic limits documented by T-Invest.

The public T-Invest FAQ still contains a generic statement that a public service based on T-Invest API is not permitted and that the API is supplied without retransmission rights. OpenInvest does not erase or reinterpret that conflict.

The conflict is handled through a deliberately narrow registry decision rather than an unrestricted provider approval.

Canonical registry result, merged in PR #161 on 2026-09-09:

```text
Source/use identifier:
TINVEST_CORPORATE_ACTIONS_CONSTRAINED

Allowed:
GetDividends   → DIVIDEND
GetBondCoupons → COUPON

Not authorized:
GetBondEvents
MarketDataService
OrdersService
OperationsService
account / portfolio sync
trading
market prices
forecasts
unrelated methods
raw API/feed redistribution
```

Mandatory operational boundary from the registry:

```text
server-side read-only token only
OpenInvest provider traffic <= 60 requests/minute
max provider concurrency = 4
automatic retry on 429 = NO
background polling/synchronization = NO
persistent raw payload storage = FORBIDDEN
DB raw archive = FORBIDDEN
cross-request provider cache = FORBIDDEN
Redis provider cache = FORBIDDEN
historical raw provider archive = FORBIDDEN
request-lifetime processing only
```

Financial-truth boundary:

T-Invest Corporate Actions are external reference/event evidence only. They must not automatically create or mutate:

- `DIVIDEND` ledger transactions;
- `COUPON` ledger transactions;
- cash balances;
- realized income;
- tax transactions.

Realized portfolio truth remains ledger-owned.

Current T-Invest status:

```text
Technical fit:                         STRONG
Dividend coverage:                    YES
Coupon coverage:                      YES
Exact normalized public display:      YES — support confirmed exact scenario
Derived analytics:                    YES — support confirmed exact scenario
Raw redistribution:                   FORBIDDEN by OpenInvest constrained mode
Caching / persistence:                NO-STORE constrained mode
Cost for approved API use:            FREE per reviewed public API documentation
Published FAQ consistency:            CONFLICTING
Registry status:                       CONDITIONAL-GO — exact constrained scope
Runtime activation:                    SEPARATE IMPLEMENTATION / REVIEW REQUIRED
```

This replaces the older pre-PR #161 state that said registry approval was not yet granted.

### 5.2 Interfax / e-disclosure — rights substantially clarified, financially NO-GO

Interfax / e-disclosure previously offered a test/API route. OpenInvest then asked specifically about:

- public normalized display;
- derived analytics/calendar use;
- caching/history/retention;
- Corporate Actions coverage;
- technical limits;
- production pricing.

On 2026-09-09 Interfax-CRKI replied that the information on the public site is public and that the API subscription principally purchases automated access. In direct response to OpenInvest's listed intended uses, the provider stated that the listed use variants are permitted and that no additional use-specific paperwork is required beyond the service relationship.

Provider evidence also states:

- the contract does not require deletion of previously received information;
- therefore no provider-stated retention/deletion limit was identified in that reply;
- default API access covers events created during the contract period;
- archive access can be added back to `2020-07-01`;
- archive access costs `+50%` of the monthly tariff;
- minimum subscription term is 3 months;
- the provider asked whether OpenInvest needs messages, files, or both.

Current evidence interpretation for the exact described use:

```text
Automated API access:                 YES — contract/subscription route
Public normalized display:           CONFIRMED by provider response
Derived analytics/calendar:           CONFIRMED by provider response
Retention of previously received:    CONFIRMED — no deletion requirement stated
Archive to 2020-07-01:                AVAILABLE for +50% monthly tariff
Minimum contract term:               3 months
Exact current quote:                  REQUESTED
Free/open-source/research exception:  REQUESTED / UNKNOWN
Runtime activation:                   NO
```

The current `Data Source Registry` already classifies `INTERFAX_EDISCLOSURE_API` as financially `NO-GO` under the zero-budget project constraint and records a reviewed public tariff above `0 RUB`.

Improved rights evidence does **not** change that status. OpenInvest does not authorize a paid subscription, credentials, runtime wiring or public-site scraping while the external-data budget remains `0 RUB`.

A fresh pricing clarification was sent on 2026-09-09 asking for:

- messages-only monthly price;
- minimum total for the required 3-month term;
- messages + files price;
- archive price including the +50% increment;
- VAT treatment;
- any setup/API connection fee;
- free test-access duration;
- any free or discounted open-source / educational / research option.

Until a `0 RUB` production-compatible mode is evidenced or the project owner explicitly changes the budget, the decision remains:

```text
INTERFAX_EDISCLOSURE_API = NO-GO — ZERO-BUDGET CONSTRAINT
```

### 5.3 Financial Data API / fdnpy — new research candidate

On 2026-09-06 OpenInvest received a provider message advertising, among other datasets:

- 252,000+ symbols across 20+ exchanges;
- stocks, ETFs, commodities, OTC, indices, options, futures, crypto, forex and mutual funds;
- end-of-day / intraday / real-time prices;
- company data and identifiers including ISIN and FIGI;
- fundamentals and news;
- upcoming events including dividends and splits;
- Corporate Actions coverage advertised back to 1964;
- SDK/examples at `https://github.com/financialdatanet/fdnpy`.

These are provider claims, not independently verified OpenInvest production evidence.

OpenInvest sent a targeted clarification on 2026-09-09 covering:

- exact MOEX / Russian equities coverage;
- Russian bonds / OFZ coverage;
- dividend dates and fields;
- coupon schedules/payments;
- maturities, amortizations, offers/calls and corrections where available;
- public normalized display rights;
- OpenInvest-derived analytics rights;
- whether normalized public display is classified as redistribution;
- temporary caching / normalized persistence / historical retention;
- attribution;
- rate limits;
- trial/test access;
- exact production pricing;
- open-source / research / educational terms.

Current classification:

```text
Technical candidate:             INTERESTING
Corporate Actions:               CLAIMED
MOEX / Russian coverage:         UNKNOWN
Russian bonds / OFZ:             UNKNOWN
Coupon coverage:                 UNKNOWN
Public normalized display:       UNKNOWN
Derived analytics rights:        UNKNOWN
Caching / retention:             UNKNOWN
Redistribution classification:   UNKNOWN
Attribution:                     UNKNOWN
Free production access:          UNKNOWN
Production price:                UNKNOWN
Trial access:                    UNKNOWN
Registry status:                 NONE
Runtime activation:              NO
```

No fallback or runtime use is authorized while those dimensions remain unresolved.

### 5.4 Cbonds / InvestFunds

Provider evidence received:

- bond/instrument parameters and coupon schedules are available;
- dividends and splits are available;
- limited trial access for approximately 10–15 ISINs can be provided;
- derived analytics/calendar use is treated as retransmission;
- public display depends on the exact third-party access model;
- historical retention is dataset-specific;
- rate limits communicated: 10,000 requests per method/day and 30 requests/minute;
- no separate open-source/research commercial format was offered in the received response;
- production price depends on the retransmission model.

OpenInvest sent a clarification on public normalized use, derived analytics, retention, attribution and minimum production price.

Current status:

```text
TRIAL AVAILABLE / RETRANSMISSION AND PRODUCTION TERMS UNRESOLVED
ZERO-BUDGET PRODUCTION GO = NOT ESTABLISHED
```

### 5.5 NSD / НРД

NSD information-services sales stated that the relevant information-service route is available only to legal entities and that NSD does not conclude the relevant data-provision contracts with individuals or individual entrepreneurs.

Current status:

```text
NO-GO UNDER CURRENT PROJECT STATUS — LEGAL ENTITY REQUIRED
```

This is an organizational/contractual blocker, not a claim that NSD lacks useful Corporate Actions data.

### 5.6 Finam

Finam acknowledged the questions and routed them to colleagues responsible for Trade API.

Current status:

```text
ROUTED TO TRADE API TEAM / AWAITING SUBSTANTIVE TECHNICAL RESPONSE
```

Partner/referral offers are unrelated to source approval.

### 5.7 VTB

VTB requires an official paper request addressed to the bank leadership at the postal route supplied by VTB. OpenInvest asked for an electronic route; VTB repeated the postal requirement.

Current status:

```text
FORMAL POSTAL REQUEST REQUIRED
CURRENT ELECTRONIC OUTREACH PATH BLOCKED
PHYSICAL REQUEST NOT RECORDED AS SENT
```

### 5.8 Bank of Russia / CBR

No email was sent to the Bank of Russia in this outreach wave. The formal online reception remains the appropriate route if later needed.

No universal production Corporate Actions feed has been established from CBR evidence.

### 5.9 Awaiting substantive response

As of 2026-09-09:

```text
Moscow Exchange / MOEX
Alfa Investments
BCS
Gazprombank
Sber
SPB Exchange
Finmarket
Smart-Lab
Banki.ru
RBC Investments
```

Silence is not rejection and does not establish any source/use right.

## 6. Current source-selection decision

The source-selection state is no longer accurately described as "no approved candidate at all".

The correct current state is:

```text
Unrestricted production Corporate Actions source:
NONE

Registry-approved constrained source/use mode:
TINVEST_CORPORATE_ACTIONS_CONSTRAINED

Allowed T-Invest canonical mappings:
GetDividends   → DIVIDEND
GetBondCoupons → COUPON

T-Invest runtime/provider activation:
SEPARATELY REVIEW-GATED / NOT ACTIVATED BY THIS DOCUMENT

Interfax / e-disclosure:
RIGHTS EVIDENCE STRONGER, BUT FINANCIAL NO-GO UNDER 0 RUB BUDGET

Financial Data API:
NEW CANDIDATE / CLARIFICATION SENT / RIGHTS + MOEX COVERAGE UNRESOLVED

Cbonds / InvestFunds:
TRIAL AVAILABLE / COMMERCIAL + REDISTRIBUTION TERMS UNRESOLVED
```

## 7. Zero-budget rule

OpenInvest currently has a project-level constraint of:

```text
External financial-data budget = 0 RUB
```

Consequences:

- a technically excellent provider is still a production `NO-GO` if paid access is mandatory;
- free trial access may be used only for separately authorized evaluation and does not itself authorize production;
- paid archival options are not activated;
- no contract/subscription may be entered into merely because source rights are otherwise acceptable;
- any future change to this budget must be an explicit owner decision, not an inference by a builder/reviewer.

This budget constraint is an internal project decision, not a statement that paid providers are unsuitable in general.

## 8. Required evidence for new source/use modes

Before proposing any new provider registry row, establish at minimum:

| Dimension | Required state |
| --- | --- |
| Exact technical coverage | VERIFIED |
| Automated retrieval rights | VERIFIED |
| Public normalized display | VERIFIED |
| Derived analytics | VERIFIED |
| Caching / retention | VERIFIED or explicitly bounded |
| Attribution | DEFINED |
| Rate / traffic policy | DEFINED |
| Contract/account dependency | DEFINED |
| Production cost | KNOWN |
| Fit with current project budget | ACCEPTED |
| Production source/use scope | EXACTLY BOUNDED |

A sent email, generic support reply, SDK, sandbox key, test connection or provider marketing page is never enough by itself.

## 9. Relationship to Data Source Registry

The `Data Source Registry` remains the authority for runtime-eligible external source/use modes.

Current relevant states:

- `TINVEST_CORPORATE_ACTIONS_CONSTRAINED` = `CONDITIONAL-GO` for the exact bounded DIVIDEND/COUPON scenario;
- `INTERFAX_EDISCLOSURE_API` = `NO-GO` under the current zero-budget constraint;
- `NSD_CORPORATE_ACTIONS_API` = `NO-GO` under current organizational/contractual constraints;
- MOEX reviewed market-data source/use modes remain separately governed by their existing registry rows;
- Financial Data API has no registry row and no runtime authorization;
- Cbonds/InvestFunds has no new approved row from this outreach.

This research log does not broaden any registry permission.

## 10. Public-repository evidence minimization

Do not store in the public repository unless materially necessary:

- Gmail message IDs;
- private/internal email headers;
- personal account metadata;
- tracking identifiers;
- complete private correspondence where a concise evidence summary is sufficient;
- provider secrets, credentials or tokens.

Retain only the minimum evidence necessary to support a source decision.

## 11. Scope exclusions

This document does not itself authorize:

- provider credentials;
- provider runtime wiring;
- paid subscriptions;
- scraping;
- database migrations;
- raw provider persistence;
- background polling;
- broader T-Invest services;
- new CorporateActionEvent kinds;
- ledger mutation from provider evidence;
- Stage 3.78 or unrelated product scope.

## 12. Current result

```text
Corporate Actions source-rights research:
ACTIVE / MATERIAL EVIDENCE OBTAINED

Canonical constrained source decision:
TINVEST_CORPORATE_ACTIONS_CONSTRAINED = CONDITIONAL-GO

Runtime T-Invest adapter:
SEPARATE IMPLEMENTATION / REVIEW REQUIRED

Interfax rights:
PUBLIC NORMALIZED USE + DERIVED ANALYTICS + RETENTION CONFIRMED FOR DESCRIBED CONTRACTED USE

Interfax production under current budget:
NO-GO

Financial Data API:
CLARIFICATION SENT / COVERAGE + RIGHTS + COST UNRESOLVED

OpenInvest external-data budget:
0 RUB
```
