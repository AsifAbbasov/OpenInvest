# Data Source Registry

| Field | Value |
| --- | --- |
| Document ID | REG-DATA-001 |
| Version | 1.0.4 |
| Status | Active |
| Owner | Principal Architect |
| Supersedes | Ad hoc external-source selection |
| Dependencies | Documents 42–43 |
| Last Review Date | 2026-09-04 |
| Next Review Date | 2026-12-21 |

External source implementation remains prohibited unless the exact source/use mode has an approved row below.
Approval is scope-bound: a limited approval does not authorize use outside the row's explicit status and
restrictions.

| Source | Owner | License/terms | Rate limits | Caching | Redistribution | Freshness | Fallback | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `INTERFAX_EDISCLOSURE_API` | Principal Architect / Corporate Actions | Official Interfax-CRKI disclosure gateway is contract/subscription based. Reviewed tariff: 16,180 RUB/month without VAT for message-publication data with a 3-month minimum; complete publications 27,000 RUB/month. Public-site availability does not authorize treating the paid API as a free production feed. | Contracted API; exact limits depend on subscribed service. Public-site scraping is not approved as a substitute. | FORBIDDEN while NO-GO. | FORBIDDEN while NO-GO. | Near-publication disclosure events when contracted; no OpenInvest production freshness claim while NO-GO. | None; fail closed. | `NO-GO — zero-budget constraint; no API subscription and no public-site scraper authorized` |
| `NSD_CORPORATE_ACTIONS_API` | Principal Architect / Corporate Actions | NSD API/getCorpActions/GetNews provide the needed corporate-action data under information-service subscription/contract terms. Public nsddata.ru user agreement limits information to familiarisation, prohibits commercial/third-party use and automated processing. | API documentation states up to 1 request/second for subscribed API use. No shipped traffic authorized while NO-GO. | FORBIDDEN while NO-GO. | FORBIDDEN while NO-GO. | Corporate actions are maintained by NSD, but no OpenInvest production freshness claim while subscription/use rights are absent. | None; fail closed. | `NO-GO — zero-budget constraint; public-site automated processing forbidden and subscription not accepted` |
| `CBR_CORPORATE_ACTIONS_FEED` | Principal Architect / Corporate Actions | Bank of Russia regulates disclosure but reviewed guidance states issuers are not required to additionally submit issuer reports/material facts to the Bank when disclosure occurs on public information resources. No universal issuer corporate-actions feed was established. Legacy SEC `coupons` web-service method is marked obsolete. | N/A — no approved feed. | N/A | N/A | N/A | None. | `NO-GO — universal feed not established; legacy coupon method obsolete` |
| `ISSUER_DIRECT_DISCLOSURE` | Principal Architect / Corporate Actions | Issuer-owned disclosure/IR pages may be authoritative for that issuer, but terms, transport, schema, automation rights, retention and reuse vary by issuer and endpoint. | UNKNOWN until exact endpoint review. | UNKNOWN until exact endpoint review. | UNKNOWN until exact endpoint review. | Issuer-specific. | None; fail closed. | `REVIEW REQUIRED — approval must be per exact issuer-owned endpoint/use mode; generic scraping forbidden` |
| `MOEX_ISS_DELAYED_TQBR` | Principal Architect / Market Data | Official MOEX ISS materials permit delayed unauthenticated technical access, but state that information obtained from ISS without an agreement is for familiarisation only and any other use requires an agreement with PJSC Moscow Exchange. MOEX public web/app placement requires an information agreement; reviewed public tariff lists 15-minute delayed public data at 25,500 RUB/month for non-issuers. Market Data Policy separately governs distribution, Non-display, and Derived Data use. | MOEX public hard request quota remains UNKNOWN in reviewed material. Existing adapter policy remains one request per `Quote`, no automatic retry/fan-out/poll loop, 5s client timeout. No shipped traffic is authorized while this row is NO-GO. | FORBIDDEN for shipped/product use under current decision. Existing deterministic test fixtures only; no product cache/persistence. | FORBIDDEN. No public API/UI display, onward distribution, redistribution, or derived-product distribution. | Guest ISS market data is approximately 15-minute delayed; Stage 3.59 adapter time/provenance validation remains technical implementation evidence only and does not authorize product use. | None; fail closed and keep provider unconfigured in shipped composition. | `NO-GO — adapter/test code may remain, but shipped runtime/public/non-display/derived use is forbidden until exact MOEX contractual rights, cost acceptance, fresh registry approval, and separately reviewed runtime wiring exist` |

## Corporate-actions source decision

Stage 3.61 records that no reviewed source currently satisfies OpenInvest's combined requirements of zero cost,
automated machine-readable retrieval, broad share/bond coverage, production-use rights, public calendar/heatmap
display, and persistence/normalization rights.

Accordingly:

- Interfax e-Disclosure API remains technically suitable but financially NO-GO;
- NSD corporate-actions services remain technically suitable but financially/contractually NO-GO;
- automated scraping of public e-Disclosure or NSDData pages is not authorized;
- Bank of Russia does not provide the required universal issuer-event feed in the reviewed surfaces;
- issuer-direct endpoints require exact per-endpoint review and cannot be generalized into a scraper.

The absence of an approved production source does not block a future provider-neutral corporate-action boundary,
deterministic test fixtures, or pure calendar/heatmap projections after a separately merged planning and separately
authorized implementation stage. It does block real external HTTP ingestion and public claims of all-market coverage.

Official/reviewed evidence:

- `https://e-disclosure.ru/poluchenie-informacii/shlyuz-api`
- `https://gateway.e-disclosure.ru/swagger/ui/index.html`
- `https://nsddata.ru/ru/products/2`
- `https://nsddata.ru/en/products/getcorpactions_v2`
- `https://nsddata.ru/ru/user-agreement`
- `https://nsddata.ru/ru/documents`
- `https://www.cbr.ru/explan/corporate_rel/`
- `https://www.cbr.ru/development/SEC/`

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
