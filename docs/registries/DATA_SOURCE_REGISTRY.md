# Data Source Registry

| Field | Value |
| --- | --- |
| Document ID | REG-DATA-001 |
| Version | 1.0.7 |
| Status | Active |
| Owner | Principal Architect |
| Supersedes | Ad hoc external-source selection |
| Dependencies | Documents 42–43 |

External source implementation remains prohibited unless the exact source/use mode has an approved row below.
Approval is scope-bound: a limited approval does not authorize use outside the row's explicit status and
restrictions.

| Source | Owner | License/terms | Rate limits | Caching | Redistribution | Freshness | Fallback | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `INTERFAX_EDISCLOSURE_API` | Principal Architect / Corporate Actions | Contract/subscription: publication messages 16,180 RUB/month without VAT, minimum 3 months; complete publications 27,000 RUB/month. Public-site access is not a free API license. | Subscription-dependent; scraping not approved. | FORBIDDEN while NO-GO. | FORBIDDEN while NO-GO. | Near-publication when contracted; no current production claim. | None; fail closed. | NO-GO: zero budget, no subscription or public-site scraper authorization. |
| `NSD_CORPORATE_ACTIONS_API` | Principal Architect / Corporate Actions | API/getCorpActions/GetNews require subscription/contract; public nsddata.ru terms permit familiarisation only, forbid commercial/third-party use and automated processing. | Subscribed API ≤1 request/second; no traffic while NO-GO. | FORBIDDEN while NO-GO. | FORBIDDEN while NO-GO. | NSD-maintained events; no OpenInvest production claim without rights. | None; fail closed. | NO-GO: zero budget; no accepted subscription or public-site automation rights. |
| `TINVEST_CORPORATE_ACTIONS_CONSTRAINED` | Principal Architect / Corporate Actions | Exact ticket `3-781291` scenario only; FAQ conflict remains. | Official 200/min; internal ≤60/min, concurrency 4. | Request-lifetime only; persistent/cross-request storage forbidden. | Normalized display/derived analytics only; raw redistribution forbidden. | On-demand; no freshness/SLA claim. | None; fail closed. | CONDITIONAL-GO source/use only; full binding restrictions below; runtime activation separate. |
| `CBR_CORPORATE_ACTIONS_FEED` | Principal Architect / Corporate Actions | Reviewed guidance does not require issuers to resubmit publicly disclosed reports/material facts to CBR; universal issuer-event feed not established. | N/A: no approved feed. | N/A | N/A | N/A | None. | NO-GO; legacy SEC `coupons` method obsolete. |
| `ISSUER_DIRECT_DISCLOSURE` | Principal Architect / Corporate Actions | Issuer pages may be authoritative; terms, transport, schema, automation, retention and reuse require exact endpoint review. | UNKNOWN pending review. | UNKNOWN pending review. | UNKNOWN pending review. | Issuer-specific. | None; fail closed. | REVIEW REQUIRED per issuer/endpoint/use; no generic scraper approval. |
| `MOEX_ISS_DELAYED_TQBR` | Principal Architect / Market Data | Guest ISS: familiarisation only; other use/public web or app placement needs an agreement. Reviewed non-issuer public-data tariff: 25,500 RUB/month for 15-minute delayed data. Distribution, Non-display and Derived Data policies apply. | Public hard quota UNKNOWN; dormant adapter: one request/Quote, no retry/fan-out/poll loop, 5s timeout. | Product storage/cache FORBIDDEN; deterministic fixtures only. | Public/API/UI/onward/derived redistribution FORBIDDEN. | Approximately 15-minute delay; adapter provenance/time validation is technical evidence only. | None; fail closed, provider unconfigured. | NO-GO for runtime/public/non-display/derived use; rights, cost acceptance, registry approval and separate wiring review required below. |

## Corporate-actions source decision

Stage 3.61 found no acceptable production source. The 2026-09-09 review grants only
`TINVEST_CORPORATE_ACTIONS_CONSTRAINED` the exact source/use CONDITIONAL-GO below; other rows remain binding.
Feature 3D adapter implementation merged through PR #164 / `247081a95a7daf33c0077c88c5f41cb2e8161865`.
That merge does not activate runtime, provision a token or authorize production traffic.

Official/reviewed evidence:

- [Corporate Actions source evidence](../research/CORPORATE_ACTIONS_SOURCE_OUTREACH.md)
- [Constrained T-Invest proposal](../research/T_INVEST_CONSTRAINED_SOURCE_USE_PROPOSAL.md)
- T-Invest support ticket `3-781291`
- `https://e-disclosure.ru/poluchenie-informacii/shlyuz-api`
- `https://gateway.e-disclosure.ru/swagger/ui/index.html`
- `https://nsddata.ru/ru/products/2`
- `https://nsddata.ru/en/products/getcorpactions_v2`
- `https://nsddata.ru/ru/user-agreement`
- `https://nsddata.ru/ru/documents`
- `https://www.cbr.ru/explan/corporate_rel/`
- `https://www.cbr.ru/development/SEC/`

## TINVEST_CORPORATE_ACTIONS_CONSTRAINED decision

### Evidence and FAQ conflict

The linked outreach/proposal preserve the exact scenario: a free public application uses backend-only
retrieval and provider-neutral normalization for public dividend/coupon display and derived analytics,
without raw API/feed or token exposure. T-Invest support answered ticket `3-781291`:

> `Такое использование разрешено.`

> English translation for readability: `This use is permitted.`
> The original Russian provider wording above is preserved as the primary evidence.

The published FAQ generically forbids public services/retransmission; that conflict remains unresolved.
The support answer applies only to this exact normalized-display/derived-analytics scenario.
Raw redistribution is forbidden; broader use requires a new source-rights review.

### Allowed source/use scope

Allowed only:

```text
GetDividends    → canonical DIVIDEND
GetBondCoupons  → canonical COUPON
```

Not authorized:

- `GetBondEvents` production mapping;
- MarketDataService;
- OrdersService;
- OperationsService;
- account sync;
- portfolio sync;
- trading;
- market prices;
- forecasts;
- unrelated T-Invest methods;
- new provider-specific public API fields;
- new CorporateActionEvent kinds.

### Storage and processing boundary

Until explicit provider-specific retention/caching rights are established:

```text
persistent raw provider payload storage = FORBIDDEN
database archive of T-Invest raw data    = FORBIDDEN
cross-request provider cache             = FORBIDDEN
Redis provider cache                     = FORBIDDEN
historical raw provider archive          = FORBIDDEN
background polling/synchronization       = FORBIDDEN
```

The only allowed processing model is request-lifetime processing:

```text
provider response
→ validation
→ normalization
→ existing projections
→ response
→ raw provider object discarded
```

Request-scoped deduplication/memoization may exist only within one in-flight request and must not survive request completion.

### Authentication and secret boundary

- server-side read-only Bearer token only;
- token supplied only through runtime secret/environment configuration;
- token never committed to Git;
- token never sent to frontend/mobile clients;
- token never written to logs, traces, metrics or errors;
- authorization headers redacted from observability;
- separate development and production tokens required operationally;
- no trading-capable token;
- rotation without code changes.

The implementation cannot prove broker-side token permissions without invoking additional unapproved surfaces, so read-only scope remains a deployment/provisioning contract.

### Traffic boundary

Official InstrumentsService limit reviewed in the constrained proposal:

```text
200 requests/minute
```

OpenInvest internal cap:

```text
<= 60 provider requests/minute
max concurrency = 4
automatic retry on 429 = NO
unbounded fan-out = FORBIDDEN
background polling/synchronization = FORBIDDEN
```

Provider throttling must fail closed and must not be bypassed by another unapproved source.

### Attribution / provenance boundary

Until a provider-specific attribution requirement is established, use conservative provenance only:

```text
Source: T-Invest API
```

No logo use, partnership claim or endorsement claim is authorized. Derived values must be distinguishable as OpenInvest calculations rather than provider-supplied facts.

### Financial-truth boundary

T-Invest Corporate Actions are external reference/event truth only. Provider data must **not** automatically mutate the immutable OpenInvest ledger.

No automatic creation of:

- `DIVIDEND` ledger transaction;
- `COUPON` ledger transaction;
- cash balance;
- realized income;
- tax transaction.

Provider data may support only:

- Corporate Actions calendar;
- heatmap;
- expected/provider-derived cash-flow projections;
- already-approved derived analytics.

Realized portfolio truth remains ledger-owned through user/manual or separately authorized transaction-import evidence.

### Runtime/governance boundary

This registry decision is source/use-rights approval only. It does **not** by itself authorize or perform runtime activation.

Feature 3D / PR #164 completed adapter, secret, transport, normalization, provider-neutral, failure
and runtime-composition review; the adapter remains disabled by default, without live-token/traffic evidence.
Activation, broader methods/use, persistence/cache/polling, provider-specific public contracts or ledger
coupling require fresh separate authorization. This registry does not authorize Stage 3.78.

## MOEX_ISS_DELAYED_TQBR activation decision

Stage 3.60 records a fail-closed **NO-GO FOR SHIPPED RUNTIME** under the current OpenInvest zero-budget constraint.

The existing Stage 3.59 adapter may remain in the repository and may continue to be compiled, unit-tested,
security-scanned, and reviewed. Optional human-run technical smoke evidence may be used only to validate adapter
compatibility; it does not authorize product collection, persistence, automated polling, display, redistribution,
non-display processing, derived analytics, or third-party services.

Shipped application composition must remain provider-free. In particular, Stage 3.60 does not authorize wiring
`moexiss.NewQuoteProvider` into `backend-go/cmd/api`, populating user-visible asset prices, or using MOEX data for
portfolio valuation, alerts, insights, history, or other product calculations.

A future production-use proposal requires a new review that defines the exact MOEX use category and supplies:

- contractual/usage rights for that exact mode;
- required attribution/display/audit obligations;
- explicit Principal Architect acceptance of monetary cost;
- rate-limit/traffic evidence;
- cache/retention/persistence rights if applicable;
- a fresh registry status change to an exact approved production mode;
- separately reviewed runtime composition/public-contract changes.

Technical availability, delayed access, or the existence of a merged adapter must never be treated as production
source approval.

Official evidence reviewed for this decision:

- `https://www.moex.com/a2193`
- `https://www.moex.com/a8531`
- `https://www.moex.com/ru/products/publicdata`
- `https://www.moex.com/s1147`
- `https://www.moex.com/en/datapolicy/`
- `https://www.moex.com/ru/datapolicy/`
- `https://www.moex.com/s3503`

## Reserved non-production example identifiers

The identifiers below exist only to make OpenAPI examples structurally valid. They do not name an
external provider, authorize collection, assert provenance, or permit production use. Runtime
responses must never emit them.

| Code | Example purpose | Production status |
| --- | --- | --- |
| `EXAMPLE_MARKET_DATA` | Normalized asset price/source reference | Forbidden |
| `EXAMPLE_CORPORATE_ACTIONS` | Dividend-event source reference | Forbidden |
| `EXAMPLE_PURCHASING_POWER` | Purchasing-power input source reference | Forbidden |
