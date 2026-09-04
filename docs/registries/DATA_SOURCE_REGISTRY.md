# Data Source Registry

| Field | Value |
| --- | --- |
| Document ID | REG-DATA-001 |
| Version | 1.0.3 |
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
| `MOEX_ISS_DELAYED_TQBR` | Principal Architect / Market Data | Official MOEX ISS materials permit delayed unauthenticated technical access, but state that information obtained from ISS without an agreement is for familiarisation only and any other use requires an agreement with PJSC Moscow Exchange. MOEX public web/app placement requires an information agreement; reviewed public tariff lists 15-minute delayed public data at 25,500 RUB/month for non-issuers. Market Data Policy separately governs distribution, Non-display, and Derived Data use. | MOEX public hard request quota remains UNKNOWN in reviewed material. Existing adapter policy remains one request per `Quote`, no automatic retry/fan-out/poll loop, 5s client timeout. No shipped traffic is authorized while this row is NO-GO. | FORBIDDEN for shipped/product use under current decision. Existing deterministic test fixtures only; no product cache/persistence. | FORBIDDEN. No public API/UI display, onward distribution, redistribution, or derived-product distribution. | Guest ISS market data is approximately 15-minute delayed; Stage 3.59 adapter time/provenance validation remains technical implementation evidence only and does not authorize product use. | None; fail closed and keep provider unconfigured in shipped composition. | `NO-GO — adapter/test code may remain, but shipped runtime/public/non-display/derived use is forbidden until exact MOEX contractual rights, cost acceptance, fresh registry approval, and separately reviewed runtime wiring exist` |

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
