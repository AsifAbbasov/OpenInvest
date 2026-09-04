# Stage 3.61 — Dividend / Coupon Calendar + Heatmap Source and Boundary Planning

| Field | Value |
| --- | --- |
| Document ID | STAGE-03-61-CORPORATE-ACTIONS-CALENDAR-PLAN |
| Version | 0.1.0-candidate |
| Status | Planning/source-governance only; no runtime ingestion, API, DB, frontend, commit/push, PR, Ready, merge, or implementation authorization |
| Owner | Principal Architect / Corporate Actions |
| Canonical planning base | `develop@7c022e6de1ab0a86ebf96ede48fafabc15b9f71c` |
| Protected-base tree | `13ffddcb1557953f49df0b8b0bd69e71dc5adb58` |
| Dependencies | Stage 3.60; Data Source Registry; Review Workflow v1.4.0 |
| Date | 2026-09-04 |

## 1. Purpose

Plan the next product feature, Dividend / Coupon Calendar + Heatmap, without coupling OpenInvest to a paid or unapproved feed.

Stage 3.61 must answer two questions:
1. Is there a zero-budget, production-usable automated corporate-actions source today?
2. What is the smallest provider-neutral boundary that lets us build the feature honestly while the source remains unresolved?

No runtime ingestion or UI is authorized by this document.

## 2. Source research decision

### Interfax e-Disclosure API

The official Interfax-CRKI disclosure gateway is REST/JSON and technically suitable for automated dividend/coupon ingestion. Current reviewed tariff is paid: 16,180 RUB/month without VAT for message-publication data, minimum 3 months; complete publications cost 27,000 RUB/month.

Decision: `INTERFAX_EDISCLOSURE_API = NO-GO` under the zero-budget constraint.

The public website is not approved as a free scraper substitute.

References:
- https://e-disclosure.ru/poluchenie-informacii/shlyuz-api
- https://gateway.e-disclosure.ru/swagger/ui/index.html

### NSD corporate-actions services

NSD provides exactly the required data class through API NSD/getCorpActions/GetNews. Public examples include cash dividends and coupon-payment events.

However, API NSD is a subscription information service. The reviewed public-site user agreement restricts information to familiarisation, prohibits commercial/third-party use, and prohibits automated processing.

Decision:
- `NSD_CORPORATE_ACTIONS_API = NO-GO` under the zero-budget constraint;
- `NSDDATA_PUBLIC_SITE_SCRAPING = FORBIDDEN`.

References:
- https://nsddata.ru/ru/products/2
- https://nsddata.ru/en/products/getcorpactions_v2
- https://nsddata.ru/ru/user-agreement
- https://nsddata.ru/ru/documents

### Bank of Russia

Bank of Russia regulates disclosure and documents dividend/coupon disclosure obligations, but its guidance states that issuers are not required to additionally submit issuer reports/material facts to the Bank when disclosure occurs on public information resources. The reviewed surfaces therefore do not establish a universal issuer corporate-actions event feed.

The legacy securities-market web service marks its `coupons` method obsolete.

Decision:
- `CBR_ISSUER_DISCLOSURE_FEED = NOT AVAILABLE`;
- `CBR_SEC_COUPONS = NO-GO / OBSOLETE`.

References:
- https://www.cbr.ru/explan/corporate_rel/
- https://www.cbr.ru/development/SEC/

### Direct issuer disclosures

Issuer-owned IR/disclosure pages may be authoritative for one issuer, but transport, schema, automation rights, retention and reuse vary by issuer.

Decision: `ISSUER_DIRECT_DISCLOSURE = REVIEW REQUIRED PER EXACT ENDPOINT/ISSUER`.

Generic issuer-site scraping is not authorized.

## 3. Source conclusion

No reviewed source currently satisfies all of:
- zero monetary cost;
- automated machine-readable retrieval;
- broad listed share/bond coverage;
- production-use rights;
- public calendar/heatmap display;
- persistence/normalization rights.

Therefore:

`ALL_MARKET_AUTOMATED_CORPORATE_ACTIONS_SOURCE = NO-GO`

This blocks real external ingestion, not the provider-neutral feature architecture.

## 4. Canonical boundary

Future implementation should preserve:

provider adapter (future, separately approved)
→ CorporateActionProvider
→ CorporateActionEvent
→ calendar projection / heatmap aggregation
→ future API/UI

Provider HTML/JSON, NSD codes, Interfax message types, URLs and scraper selectors must stay outside application/domain models.

## 5. Minimal event model

Future implementation may introduce provider-neutral concepts equivalent to:

- kind: `DIVIDEND` or `COUPON`;
- status: `ANNOUNCED`, `CONFIRMED`, `PAID`, `CANCELLED`;
- stable application event identity;
- canonical instrument identity;
- optional record date;
- optional payment date;
- optional exact-decimal amount per unit;
- currency;
- source `AsOf`;
- OpenInvest `RetrievedAt`;
- mandatory provenance for external events.

Rules:
- recommendation/proposal must not silently become `CONFIRMED`;
- unknown amount remains absent, never zero;
- unknown dates remain absent, never fabricated;
- no `float64` money;
- corrections/cancellations must be representable without deleting historical evidence.

## 6. Dividend lifecycle

Minimum mapping:
- board recommendation/proposal → `ANNOUNCED`;
- shareholder-approved/binding issuer decision → `CONFIRMED`;
- disclosed completion of payment → `PAID`;
- disclosed cancellation/superseding decision → `CANCELLED`.

The heatmap must never present `ANNOUNCED` income as guaranteed.

## 7. Coupon lifecycle

- schedule/payment date is accepted only from source evidence;
- floating/variable amount remains unknown until established;
- fixed amount must come from exact source evidence or a separately approved deterministic calculation contract;
- payment completion is distinct from scheduled payment;
- amortization/redemption is out of scope.

## 8. Calendar and heatmap projections

Calendar projection must be pure and deterministic over canonical events.

Preferred effective date:
- dividend: RecordDate when known, otherwise PaymentDate;
- coupon: PaymentDate when known.

If no usable date exists, retain the evidence event but omit it from dated projection.

First heatmap should use event counts/density. Monetary aggregation is deferred because currencies and portfolio-effective holdings require separate contracts.

No yield, FX conversion, portfolio income forecast or tax calculation belongs in this stage.

## 9. Provider port

Future source boundary should remain narrow, conceptually:

`CorporateActions(ctx, instrumentIDs, from, to) -> []CorporateActionEvent`

No generic plugin framework, event bus, Kafka, microservice or provider registry framework.

## 10. Persistence

No migration is authorized.

Before persistence, a later plan must freeze:
- idempotent source-event identity;
- correction/supersession rules;
- provenance retention;
- uniqueness;
- replay behavior;
- history vs current projection;
- source retention rights.

Tests may use deterministic in-memory fixtures only.

## 11. Failure semantics

Future ingestion must fail closed:
- unavailable source → no fabricated event;
- malformed source event → reject/quarantine;
- unknown instrument mapping → no ticker guessing;
- duplicate event → deterministic dedupe only after identity contract exists;
- correction → preserve evidence and update projection deterministically;
- stale retrieval → expose freshness/provenance.

UNKNOWN remains UNKNOWN, not a defect and not an invented value.

## 12. Explicit exclusions

No Interfax subscription, NSD subscription, e-Disclosure scraping, NSDData scraping, generic issuer scraping, MOEX corporate-action ingestion, cmd/api wiring, OpenAPI, frontend/mobile, migration, worker/polling, Redis/cache, notifications, tax, portfolio forecast, yield, FX, amortization/redemption, or runtime implementation.

## 13. Recommended implementation sequence

1. **Feature 3A — Corporate Action Boundary**: canonical types, validation, provider port, deterministic fixture provider, lifecycle/failure tests. No external HTTP or persistence.
2. **Feature 3B — Calendar + Heatmap Projection**: pure projections and tests. Still no external source.
3. **Feature 3C — API/UI**: only when an honest input mode exists; never claim all-market coverage without a legitimate source.
4. **Feature 3D — Real source adapter**: only after an exact source/use mode is approved in the Data Source Registry.

This deliberately repeats the successful boundary-first pattern used for market data without repeating the Stage 3.60 licensing trap.

## 14. Stop rules

STOP if implementation would require a paid contract not explicitly accepted, unapproved scraping, provider schema leakage, instrument guessing, fabricated dates/amounts/statuses, float money, premature persistence, external HTTP before source approval, or misleading all-market UI.

## 15. Acceptance criteria

Stage 3.61 planning is acceptable only if:
- no source is falsely approved;
- rejected/unknown sources are explicit;
- no scraper is authorized;
- event lifecycle semantics are sufficient for later implementation;
- calendar/heatmap remain pure projections;
- no runtime/OpenAPI/DB/frontend/dependency/worker change is included;
- Feature 3A remains small, source-neutral and testable.

## 16. Next governed action

Run read-only planning/governance review of this plan plus synchronized Data Source Registry candidate.

Only after APPROVED and separate human commit/push authorization may the planning bundle be published as Draft PR. Feature 3A implementation then requires a separate explicit human authorization after protected merge.
