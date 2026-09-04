# Data Source Registry

| Field | Value |
| --- | --- |
| Document ID | REG-DATA-001 |
| Version | 1.0.2 |
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
| `MOEX_ISS_DELAYED_TQBR` | Principal Architect / Market Data | Public delayed ISS access is documented by MOEX; real-time requires subscription. Free access does not establish OpenInvest redistribution/non-display rights. Attribution/display obligations for any future public surface are NOT ESTABLISHED and must be re-reviewed. | MOEX public hard quota: UNKNOWN in reviewed docs. OpenInvest Feature 2 adapter policy: one request per `Quote` call, no automatic retry, no batch/fan-out/poll loop, 5s client timeout; re-review before any runtime activation. | None in Feature 2 | FORBIDDEN in Feature 2; no public API/UI display or onward redistribution | Guest ISS market data is approximately 15-minute delayed; `AsOf` is provider trade time; `RetrievedAt` is OpenInvest clock time; required data-quality expectations are exact SECID/BOARDID, reorder-safe required columns, exact decimal LAST, valid trade timestamp, and fail-closed malformed-data handling | None; fail closed | `APPROVED — adapter implementation/test scope only; shipped runtime/public activation forbidden` |

The `MOEX_ISS_DELAYED_TQBR` approval is intentionally narrow. It authorizes only the Stage 3.59-planned real
quote adapter implementation, deterministic local tests, and optional human-run manual smoke evidence. It does not authorize shipped runtime activation, public display,
redistribution, production activation, automated polling, caching, persistence, historical ingestion,
derived-data distribution, or real-time subscription.

Any broader use requires a new legal/terms and technical review plus a separately approved registry update.

## Reserved non-production example identifiers

The identifiers below exist only to make OpenAPI examples structurally valid. They do not name an
external provider, authorize collection, assert provenance, or permit production use. Runtime
responses must never emit them.

| Code | Example purpose | Production status |
| --- | --- | --- |
| `EXAMPLE_MARKET_DATA` | Normalized asset price/source reference | Forbidden |
| `EXAMPLE_CORPORATE_ACTIONS` | Dividend-event source reference | Forbidden |
| `EXAMPLE_PURCHASING_POWER` | Purchasing-power input source reference | Forbidden |
