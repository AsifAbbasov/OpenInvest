# Data Source Registry

| Field | Value |
| --- | --- |
| Document ID | REG-DATA-001 |
| Version | 1.0.6 |
| Status | Active |
| Owner | Principal Architect |
| Supersedes | Ad hoc external-source selection |
| Dependencies | Documents 42–43 |
| Last Review Date | 2026-09-09 |
| Next Review Date | 2026-12-21 |

External source implementation remains prohibited unless the exact source/use mode has an approved row below.
Approval is scope-bound: a limited approval does not authorize use outside the row's explicit status and
restrictions.

| Source | Owner | License/terms | Rate limits | Caching | Redistribution | Freshness | Fallback | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `INTERFAX_EDISCLOSURE_API` | Principal Architect / Corporate Actions | Official Interfax-CRKI disclosure gateway is contract/subscription based. Reviewed tariff: 16,180 RUB/month without VAT for message-publication data with a 3-month minimum; complete publications 27,000 RUB/month. Public-site availability does not authorize treating the paid API as a free production feed. | Contracted API; exact limits depend on subscribed service. Public-site scraping is not approved as a substitute. | FORBIDDEN while NO-GO. | FORBIDDEN while NO-GO. | Near-publication disclosure events when contracted; no OpenInvest production freshness claim while NO-GO. | None; fail closed. | `NO-GO — zero-budget constraint; no API subscription and no public-site scraper authorized` |
| `NSD_CORPORATE_ACTIONS_API` | Principal Architect / Corporate Actions | NSD API/getCorpActions/GetNews provide the needed corporate-action data under information-service subscription/contract terms. Public nsddata.ru user agreement limits information to familiarisation, prohibits commercial/third-party use and automated processing. | API documentation states up to 1 request/second for subscribed API use. No shipped traffic authorized while NO-GO. | FORBIDDEN while NO-GO. | FORBIDDEN while NO-GO. | Corporate actions are maintained by NSD, but no OpenInvest production freshness claim while subscription/use rights are absent. | None; fail closed. | `NO-GO — zero-budget constraint; public-site automated processing forbidden and subscription not accepted` |
| `TINVEST_CORPORATE_ACTIONS_CONSTRAINED` | Principal Architect / Corporate Actions | T-Invest support ticket `3-781291` explicitly permitted the exact OpenInvest scenario: server-side retrieval, normalization into OpenInvest domain objects, normalized public dividend/coupon display and derived analytics, with no raw API/feed or token exposure. Published T-Invest FAQ still generically conflicts with public-service/retransmission use; this CONDITIONAL-GO applies only to that exact constrained scenario. Allowed methods: `GetDividends` → `DIVIDEND`; `GetBondCoupons` → `COUPON`. `GetBondEvents`, MarketDataService, OrdersService, OperationsService, account/portfolio sync, trading, market prices, forecasts and unrelated methods are NOT AUTHORIZED. Server-side read-only Bearer token only; runtime secret/environment only; never Git/frontend/mobile/logs/traces/metrics/errors; authorization headers redacted; separate dev/prod tokens; no trading-capable token. Conservative provenance: `Source: T-Invest API`; no logo/partnership claim; OpenInvest-derived values distinguishable. | Official InstrumentsService limit: 200 requests/minute. OpenInvest internal cap: <=60 provider requests/minute; max concurrency 4; no automatic retry on `429`; no unbounded fan-out; no background polling/synchronization. | Persistent raw payload storage, DB archive, cross-request provider cache, Redis provider cache and historical raw archive are FORBIDDEN. Request-lifetime processing only: response → validation → normalization → existing projections → response → raw provider object discarded. | Normalized public display = CONDITIONAL-GO; derived OpenInvest analytics = CONDITIONAL-GO; raw API/feed redistribution = FORBIDDEN. | On-demand request-time source only; no provider freshness/SLA claim, polling, historical raw archive or persisted provider timeline authorized. | None; fail closed. No scraping or unapproved-source fallback. Provider failure must not mutate ledger truth. | `CONDITIONAL-GO — exact constrained DIVIDEND/COUPON source/use rights only; source-rights approval does not activate runtime. FAQ conflict remains explicit. Broader T-Invest use requires new review.` |
| `CBR_CORPORATE_ACTIONS_FEED` | Principal Architect / Corporate Actions | Bank of Russia regulates disclosure but reviewed guidance states issuers are not required to additionally submit issuer reports/material facts to the Bank when disclosure occurs on public information resources. No universal issuer corporate-actions feed was established. Legacy SEC `coupons` web-service method is marked obsolete. | N/A — no approved feed. | N/A | N/A | N/A | None. | `NO-GO — universal feed not established; legacy coupon method obsolete` |
| `ISSUER_DIRECT_DISCLOSURE` | Principal Architect / Corporate Actions | Issuer-owned disclosure/IR pages may be authoritative for that issuer, but terms, transport, schema, automation rights, retention and reuse vary by issuer and endpoint. | UNKNOWN until exact endpoint review. | UNKNOWN until exact endpoint review. | UNKNOWN until exact endpoint review. | Issuer-specific. | None; fail closed. | `REVIEW REQUIRED — approval must be per exact issuer-owned endpoint/use mode; generic scraping forbidden` |
| `MOEX_ISS_DELAYED_TQBR` | Principal Architect / Market Data | Official MOEX ISS materials permit delayed unauthenticated technical access, but state that information obtained from ISS without an agreement is for familiarisation only and any other use requires an agreement with PJSC Moscow Exchange. MOEX public web/app placement requires an information agreement; reviewed public tariff lists 15-minute delayed public data at 25,500 RUB/month for non-issuers. Market Data Policy separately governs distribution, Non-display, and Derived Data use. | MOEX public hard request quota remains UNKNOWN in reviewed material. Existing adapter policy remains one request per `Quote`, no automatic retry/fan-out/poll loop, 5s client timeout. No shipped traffic is authorized while this row is NO-GO. | FORBIDDEN for shipped/product use under current decision. Existing deterministic test fixtures only; no product cache/persistence. | FORBIDDEN. No public API/UI display, onward distribution, redistribution, or derived-product distribution. | Guest ISS market data is approximately 15-minute delayed; Stage 3.59 adapter time/provenance validation remains technical implementation evidence only and does not authorize product use. | None; fail closed and keep provider unconfigured in shipped composition. | `NO-GO — adapter/test code may remain, but shipped runtime/public/non-display/derived use is forbidden until exact MOEX contractual rights, cost acceptance, fresh registry approval, and separately reviewed runtime wiring exist` |

## Corporate-actions source decision

Stage 3.61 originally recorded that no reviewed source satisfied OpenInvest's combined requirements for a production-usable Corporate Actions source/use mode. As of the 2026-09-09 source-rights review, one exact T-Invest use mode is accepted as **CONDITIONAL-GO** at the registry level only:

`TINVEST_CORPORATE_ACTIONS_CONSTRAINED`

The decision is intentionally narrow. It does not create an unrestricted T-Invest approval, does not activate runtime composition, and does not authorize a provider token, persistence, polling, raw redistribution, broader T-Invest services or new CorporateActionEvent kinds.

Accordingly:

- `TINVEST_CORPORATE_ACTIONS_CONSTRAINED` is conditionally approved only for `GetDividends` → canonical `DIVIDEND` and `GetBondCoupons` → canonical `COUPON`, under the restrictions in this registry and the canonical constrained proposal;
- Interfax e-Disclosure API remains technically suitable but financially NO-GO;
- NSD corporate-actions services remain technically suitable but financially/contractually NO-GO;
- automated scraping of public e-Disclosure or NSDData pages is not authorized;
- Bank of Russia does not provide the required universal issuer-event feed in the reviewed surfaces;
- issuer-direct endpoints require exact per-endpoint review and cannot be generalized into a scraper.

The constrained T-Invest row resolves only the exact source/use-rights decision described in `docs/research/CORPORATE_ACTIONS_SOURCE_OUTREACH.md` and `docs/research/T_INVEST_CONSTRAINED_SOURCE_USE_PROPOSAL.md`. The separately reviewed Feature 3D adapter implementation is now complete through PR #164 / squash merge `247081a95a7daf33c0077c88c5f41cb2e8161865`. That implementation fact does **not** activate runtime, provision a token, authorize production traffic, or broaden this row. Operational use remains allowed only if separately activated under every restriction in this exact `CONDITIONAL-GO` row.

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

The published T-Invest FAQ still contains generic language that public services/retransmission are not permitted. That conflict is **not** treated as resolved or erased. The support answer is accepted only as scoped provider evidence for the exact normalized-display/derived-analytics scenario described above. Raw redistribution remains forbidden, and any broader T-Invest use requires a new source-rights review.

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

The reviewed implementation uses:

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

The required adapter/secret/transport/normalization/provider-neutral/failure-semantics/runtime-composition implementation review was completed for Feature 3D through PR #164. The merged adapter remains disabled by default and no live token or production provider traffic is claimed by Feature 3D. Any operational activation, broader T-Invest method/use mode, persistence/cache/polling change, provider-specific public contract, or financial-ledger coupling requires a fresh separately reviewed authorization. Stage 3.78 is not started or authorized by this registry/closure synchronization.

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
