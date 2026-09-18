# Data Source Registry

| Field | Value |
| --- | --- |
| Document ID | REG-DATA-001 |
| Version | 1.0.7 |
| Status | Active |
| Supersedes | Ad hoc external-source selection |
| Dependencies | Documents 42–43 |

External source implementation remains prohibited unless the exact source/use mode has an approved row below.
Approval is scope-bound: a limited approval does not authorize use outside the row's explicit status and
restrictions.

| Source | Owner | License/terms | Rate limits | Caching | Redistribution | Freshness | Fallback | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `NSD_CORPORATE_ACTIONS_API` | Principal Architect / Corporate Actions | NSD API/getCorpActions/GetNews provide the needed corporate-action data under information-service subscription/contract terms. Public nsddata.ru user agreement limits information to familiarisation, prohibits commercial/third-party use and automated processing. | API documentation states up to 1 request/second for subscribed API use. No shipped traffic authorized while NO-GO. | FORBIDDEN while NO-GO. | FORBIDDEN while NO-GO. | Corporate actions are maintained by NSD, but no OpenInvest production freshness claim while subscription/use rights are absent. | None; fail closed. | `NO-GO — zero-budget constraint; public-site automated processing forbidden and subscription not accepted` |

## Corporate-actions source decision

Stage 3.61 originally recorded that no reviewed source satisfied OpenInvest's combined requirements for a production-usable Corporate Actions source/use mode. As of the 2026-09-09 source-rights review, one exact T-Invest use mode is accepted as **CONDITIONAL-GO** at the registry level only:

`TINVEST_CORPORATE_ACTIONS_CONSTRAINED`

The decision is intentionally narrow. It does not create an unrestricted T-Invest approval, does not activate runtime composition, and does not authorize a provider token, persistence, polling, raw redistribution, broader T-Invest services or new CorporateActionEvent kinds.

Accordingly:


Canonical record: PR #164; commit(s) `247081a95a7daf33c0077c88c5f41cb2e8161865`.

Official/reviewed evidence:

- `docs/research/CORPORATE_ACTIONS_SOURCE_OUTREACH.md`
- `docs/research/T_INVEST_CONSTRAINED_SOURCE_USE_PROPOSAL.md`
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

The registry decision relies on the exact-scenario evidence preserved in the canonical outreach and constrained proposal documents. OpenInvest described a free public application where the backend obtains T-Invest API data, normalizes it into OpenInvest's provider-neutral domain model, does not expose the raw API/feed or token, publicly displays normalized dividend/coupon/event data, and builds derived OpenInvest analytics. T-Invest support answered ticket `3-781291` for that scenario:

> `Такое использование разрешено.`

> English translation for readability: `This use is permitted.`
> The original Russian provider wording above is preserved as the primary evidence.


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

Canonical record: PR #164.

## MOEX_ISS_DELAYED_TQBR activation decision

Stage 3.60 records a fail-closed **NO-GO FOR SHIPPED RUNTIME** under the current OpenInvest zero-budget constraint.

The existing Stage 3.59 adapter may remain in the repository and may continue to be compiled, unit-tested,
compatibility; it does not authorize product collection, persistence, automated polling, display, redistribution,
non-display processing, derived analytics, or third-party services.

Shipped application composition must remain provider-free. In particular, Stage 3.60 does not authorize wiring
`moexiss.NewQuoteProvider` into `backend-go/cmd/api`, populating user-visible asset prices, or using MOEX data for
portfolio valuation, alerts, insights, history, or other product calculations.

A future production-use proposal requires a new review that defines the exact MOEX use category and supplies:


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
