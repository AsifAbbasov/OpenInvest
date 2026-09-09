# Financial Data API Outreach

| Field | Value |
| --- | --- |
| Status | Research / evidence candidate — clarification sent; no source/use approval or runtime activation |
| Date | 2026-09-09 |
| Provider | Financial Data / fdnpy / FinancialData.Net candidate |
| Contact | `financedataapi@gmail.com` |
| Provider SDK / examples | `https://github.com/financialdatanet/fdnpy` |
| Canonical base | `develop@1f2fe1cc9c4c6ab282f56506d64a9754b62eef58` |
| Registry authority | `docs/registries/DATA_SOURCE_REGISTRY.md` |
| Product scope | Research candidate for Corporate Actions and potentially other future market-data/reference-data surfaces |
| Runtime activation | NOT AUTHORIZED BY THIS DOCUMENT |
| Data Source Registry transition | NONE |

## 1. Purpose

This document records the inbound Financial Data API outreach received by OpenInvest and the exact clarification sent before any source-use or production decision.

It is research/evidence only. It does **not** approve Financial Data / FinancialData.Net / fdnpy as a production source, add a `Data Source Registry` row, create credentials, enable runtime calls, or authorize implementation.

## 2. Inbound provider claim

On 2026-09-06 OpenInvest received an unsolicited/initiated provider message describing the following advertised API coverage:

- 252,000+ symbols across 20+ exchanges;
- stocks, ETFs, commodities, OTC, indices, options, futures, crypto, forex and mutual funds;
- end-of-day, intraday and real-time prices;
- company information and identifiers including CUSIP, ISIN, FIGI, CIK, LEI and EIN;
- fundamentals, news, institutional/insider data, ETF and mutual-fund data, ESG;
- upcoming events including earnings, IPOs, splits and dividends;
- Corporate Actions advertised as earnings, IPOs, splits and dividends, with historical coverage described as extending back to 1964;
- integration documentation/examples referenced through `https://github.com/financialdatanet/fdnpy`.

These are provider claims from the inbound message. They are not treated as independently verified MOEX/Russian-market coverage or as evidence of production-use rights.

## 3. Initial OpenInvest classification

Current research classification:

```text
Technical candidate:             INTERESTING
General API availability:        YES — provider claims API/SDK coverage
Corporate Actions:               YES — provider claims dividends/splits/earnings/IPOs
MOEX / Russian equities:         UNKNOWN
Russian bonds / OFZ:             UNKNOWN
Coupon schedules/payments:       UNKNOWN
Russian Corporate Actions depth: UNKNOWN
Public normalized display:       UNKNOWN
Derived analytics rights:        UNKNOWN
Caching / retention:             UNKNOWN
Redistribution classification:   UNKNOWN
Attribution requirement:         UNKNOWN
Free production access:          UNKNOWN
Production cost:                 UNKNOWN
Trial / test key:                UNKNOWN
Production registry approval:    NO
Runtime activation:              NO
```

Canonical decision rules apply:

```text
provider marketing claim != verified source coverage
API existence             != production-use rights
public SDK                != data redistribution permission
free/trial access         != public-display permission
normalization             != automatically non-redistributive use
```

## 4. Clarification sent on 2026-09-09

OpenInvest replied in the original email thread and described the exact intended model:

- OpenInvest is an independent privacy-first investment analytics/research application for retail investors;
- current focus is MOEX / Russian-market portfolios;
- backend retrieves provider data server-side;
- provider responses are normalized into OpenInvest's provider-neutral domain model;
- raw provider API responses, raw feed and API key are not exposed to users;
- public UI may display normalized dividend/coupon facts;
- OpenInvest may calculate its own calendars, expected cash flows, yields and portfolio analytics from normalized facts;
- OpenInvest does not intend to resell or provide general-purpose access to the raw provider API/feed.

The clarification asked the provider to establish the following before any production decision.

### 4.1 Russia / MOEX coverage

Requested confirmation of:

- MOEX-listed equities, with examples such as SBER, GAZP and LKOH;
- exact MOEX market segments/instrument types;
- Russian bonds, including OFZ and corporate bonds;
- supported identifiers such as ticker, ISIN and FIGI.

### 4.2 Corporate Actions coverage

Requested confirmation for Russian/MOEX securities of:

- dividends and relevant dates, including announcement, ex-date, record date and payment date where available;
- bond coupon schedules and payments;
- maturities/redemptions;
- amortizations;
- offers/calls;
- splits;
- cancellations/corrections/revisions;
- historical Corporate Actions coverage and earliest available date.

### 4.3 Public display and derived analytics rights

Requested explicit confirmation whether the exact OpenInvest model may:

- publicly display normalized provider facts in a free web/mobile product;
- publicly display OpenInvest-derived analytics and expected cash-flow calculations;
- show conservative source attribution such as `Source: FinancialData.Net`.

### 4.4 Licensing / redistribution classification

OpenInvest asked whether normalized public display + derived analytics is treated as redistribution and whether that exact model requires an Enterprise/commercial-data-display plan or may qualify for another research/open-source/non-commercial arrangement.

No conclusion about current pricing or licensing is recorded until the provider answers directly or official terms are separately reviewed and cited.

### 4.5 Caching / retention / storage

Requested exact rights for:

- temporary caching of normalized results;
- persistence of normalized Corporate Actions in PostgreSQL;
- retention of historical normalized records after provider responses are no longer current;
- storage of OpenInvest-derived analytics;
- any distinction between raw payload restrictions and normalized/derived-data restrictions.

### 4.6 API access, limits and testing

Requested:

- subscription tier required for Corporate Actions;
- applicable rate/request limits;
- trial/test API key for validation of MOEX/Russian coverage before production selection.

### 4.7 Production cost / special terms

Requested:

- minimum production price if commercial/Enterprise rights are required;
- any open-source, research, educational, startup or non-commercial terms.

## 5. Current decision

As of 2026-09-09:

```text
Financial Data API / fdnpy:
CLARIFICATION SENT / COVERAGE AND SOURCE-RIGHTS UNRESOLVED

MOEX/Russian production suitability:
NOT ESTABLISHED

Corporate Actions production suitability:
NOT ESTABLISHED

Zero-budget production GO:
NOT ESTABLISHED

Data Source Registry transition:
NONE

Runtime/provider activation:
NOT AUTHORIZED
```

The provider must not be used as a fallback for T-Invest or any other source without a separate exact source/use review and registry decision.

## 6. Next evidence required

Before any registry proposal, obtain and verify at minimum:

1. exact MOEX/Russian instrument coverage;
2. exact dividend/coupon/event fields and historical depth;
3. public normalized-display rights;
4. derived-analytics rights;
5. redistribution classification for the exact OpenInvest model;
6. cache/retention/persistence rights;
7. attribution obligations;
8. rate limits;
9. trial/test-access conditions;
10. production pricing and any research/open-source exception;
11. provider identity/official commercial entity and governing terms for the API service.

Until those items are established, the candidate remains research-only and fail-closed.
