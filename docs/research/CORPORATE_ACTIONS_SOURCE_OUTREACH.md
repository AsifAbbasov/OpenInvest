# Corporate Actions Source Outreach

| Field | Value |
| --- | --- |
| Status | Research / evidence candidate — source due diligence in progress; no provider or runtime activation authorized |
| Date | 2026-09-06 |
| Purpose | Evidence log for future Feature 3D — Real Corporate Actions Source Adapter |
| Canonical repository base | `develop@97c35a7b0fe7e2cd487c63d03b8ba79f80dcfc7b` |
| Protected-base tree | `ac8f0afa700012f238bbf8e160cfbc3c2545a350` |
| Governance authority | `docs/REVIEW_WORKFLOW.md` v1.4.0; `docs/registries/DATA_SOURCE_REGISTRY.md`; Stage 3.61 Corporate Actions planning |
| Runtime scope | None |
| Data Source Registry status change | None |
| Feature 3D implementation | NOT STARTED by this document |

## 1. Purpose

This document records source due diligence and outreach for a future:

**Feature 3D — Real Corporate Actions Source Adapter**

The goal is to identify a production-usable source/use mode that OpenInvest may lawfully and technically use for real Corporate Actions.

Stage 3.62–3.64 already established the provider-neutral domain boundary, deterministic Calendar/Heatmap projection, and HTTP/OpenAPI/frontend surface. The principal remaining blocker for Feature 3D is an **approved real data source/use mode**.

The intended future chain is:

```text
External Corporate Actions Source
        ↓
Feature 3D Provider Adapter
        ↓
CorporateActionProvider
        ↓
Canonical CorporateActionEvent
        ↓
Calendar + Heatmap Projection
        ↓
HTTP API / OpenAPI
        ↓
Frontend UI
```

This document is research/evidence only. It does not approve a source, modify the `Data Source Registry`, authorize provider implementation, or authorize runtime activation.

### Current canonical-kind boundary

The outreach intentionally asks candidates about a broader Corporate Actions capability set than the currently implemented OpenInvest domain contract.

Today, the canonical Stage 3.62 `CorporateActionEvent` supports only:

- `DIVIDEND`;
- `COUPON`.

Redemptions, amortizations and other Corporate Actions are useful source-capability evidence for future planning, but they are **not** silently authorized as current `CorporateActionEvent` kinds. Mapping or exposing them would require a separately approved domain/API/product contract extension before any runtime use.

Accordingly, a future Feature 3D adapter may initially normalize only the Corporate Action kinds already authorized by the canonical domain boundary, even if the selected provider offers a broader feed.

## 2. What the outreach is trying to establish

For each candidate, OpenInvest is trying to establish:

1. whether an official machine-readable API/feed exists for:
   - dividends;
   - coupons;
   - maturities / redemptions;
   - amortizations;
   - offers;
   - record dates;
   - corrections / cancellations;
   - other issuer / Corporate Actions events;
2. whether automated retrieval is permitted;
3. whether use is permitted in the independent research/educational OpenInvest application;
4. whether public display of normalized data to OpenInvest users is permitted;
5. whether derived calculations / analytics may be built from the data;
6. whether normalization, caching, temporary storage and persistence are permitted;
7. what retention restrictions apply;
8. whether attribution is required;
9. what rate limits / traffic policies apply;
10. whether any of the following access modes exist:
    - free tier;
    - demo;
    - research;
    - non-commercial;
    - development access;
11. whether a separate agreement, license, approval or commercial contract is required.

## 3. How OpenInvest was positioned

The outreach describes OpenInvest as:

- an independent research and educational project for retail/private investors;
- a project intended to help users better understand investment instruments;
- a project for analysis of expected cash flows, including dividends, coupons, redemptions and amortizations;
- a Corporate Actions calendar;
- a tool for understanding potential future return scenarios and risks;
- not a broker;
- not a bank;
- not an asset manager;
- not an individualized investment-recommendation service.

Repository:

`https://github.com/AsifAbbasov/OpenInvest`

## 4. Canonical outreach meaning / template

The text below is the canonical **meaning/template** of the outreach. Individual recipients received slightly adapted versions. This document does **not** assert that all sent messages were byte-for-byte identical.

> Здравствуйте!
>
> Меня зовут Asif Abbasov. Я разрабатываю OpenInvest — независимый исследовательский и образовательный проект для частных инвесторов.
>
> Цель проекта — помочь пользователю лучше понимать инвестиционные инструменты и ожидаемые денежные потоки: дивиденды, купоны, погашения и амортизации, календарь корпоративных событий, а также сценарии будущей доходности и рисков. OpenInvest не является брокером, банком, управляющей компанией или сервисом инвестиционных рекомендаций.
>
> Сейчас мы исследуем возможность подключения официального источника корпоративных действий.
>
> Просим уточнить:
>
> - существует ли официальный API/feed для Corporate Actions;
> - допускается ли автоматизированное использование данных;
> - допускается ли отображение нормализованных данных пользователям OpenInvest;
> - допускается ли построение derived analytics;
> - разрешены ли caching / temporary storage / normalization / retention;
> - требуется ли attribution;
> - какие rate limits действуют;
> - существует ли free/demo/research/non-commercial access;
> - требуется ли отдельное соглашение, лицензия или согласование.
>
> Для OpenInvest важно не активировать источник до подтверждения прав на получение, обработку и отображение данных.

### Outreach scope expansion — second wave

The first outreach wave focused mainly on:

```text
MOEX + banks / brokers
```

The second wave expands the evidence pipeline to:

```text
market infrastructure
+
professional data vendors
+
investment information platforms
+
financial media / communities
```

The purpose of this expansion is still Feature 3D source due diligence: identify an exact source/use mode that can provide dividends, coupons, maturities/redemptions, amortizations, record dates, issuer/corporate events and corrections/cancellations while also satisfying automated-retrieval rights, production-use rights, public-display rights, derived-analytics rights, caching/retention rules, attribution, acceptable rate limits and — if possible — zero-budget/free/research-compatible cost.

Outreach to a media, platform or community contact is research/source-guidance evidence only. It does not make that organization a production data provider.

## 5. Outreach register

| Organization | Category | Contact | Purpose / note | Current evidence status |
| --- | --- | --- | --- | --- |
| Moscow Exchange / MOEX | Exchange | `data@moex.com`; CC `itsales@moex.com` | Corporate Actions / market-data availability, automated retrieval, display/derived-use/caching/attribution/rate-limit/cost/production terms | AWAITING RESPONSE |
| Finam | Broker / API | `agent@corp.finam.ru` | Finam Trade API / Corporate Actions / use rights | AWAITING RESPONSE |
| Alfa Investments | Broker / bank | `support@alfadirect.ru` | Investment API / Corporate Actions / use rights | AWAITING RESPONSE |
| BCS | Broker | `info@bcs.ru` | API/data licensing request; redirect to responsible team if needed | AWAITING RESPONSE |
| T-Invest | Broker / API | `openapi@tbank.ru` → `invest-public-api@tbank.ru` | Initial contact redirected to specialist public-API route; T-Bank ticket `3-781291` registered | NEED CLARIFICATION / AWAITING SUBSTANTIVE RESPONSE |
| VTB | Broker / bank | `info@vtb.ru` | API/data-rights inquiry; VTB requested a formal official request, and an electronic-routing follow-up was sent | AWAITING ELECTRONIC ROUTING RESPONSE / NEED CLARIFICATION |
| Gazprombank | Broker / bank | `broker@gazprombank.ru` | Brokerage direction / Corporate Actions data | AWAITING RESPONSE |
| Sber | Broker / bank | `sberbank@sberbank.ru` | General official contact; routing request to investments/API/data-licensing team | AWAITING RESPONSE |
| NSD / НРД | Market infrastructure | `datasales@nsd.ru` | Corporate Actions API/information services, rights, retention, attribution, rate limits, pricing and pre-contract documentation | AWAITING RESPONSE |
| Interfax | Professional data vendor | `sales_support@interfax.ru` | Structured financial/Corporate Actions feeds, rights, retention, attribution, pricing and demo/research access | AWAITING RESPONSE |
| InvestFunds / Cbonds | Professional data vendor | `database@cbonds.info` | Bonds, coupons, maturities, amortizations, Corporate Actions and available issuer/dividend events | AWAITING RESPONSE |
| SPB Exchange | Exchange / market infrastructure | `info@spbexchange.ru` | Machine-readable market/Corporate Actions/instrument-event data and licensing/use conditions | AWAITING RESPONSE |
| Finmarket | Financial information | `news@finmarket.ru` | Structured feeds/source guidance and possible research/information cooperation | AWAITING RESPONSE |
| Smart-Lab | Investment community | `admin@smart-lab.ru` | Source guidance, possible structured data, research/community cooperation, future beta-user discovery | OUTREACH / RESEARCH CONTACT — AWAITING RESPONSE |
| Banki.ru | Financial platform | `info@banki.ru`; CC `pr@banki.ru` | Structured investment/Corporate Actions information, legal-source guidance, research/information cooperation | OUTREACH / RESEARCH CONTACT — AWAITING RESPONSE |
| RBC Investments | Financial media | `oanohina@rbc.ru` | Structured-feed/source guidance, research/information cooperation and project feedback | OUTREACH / RESEARCH CONTACT — AWAITING RESPONSE |
| Bank of Russia / CBR | Regulator | Official online reception / formal electronic request | Regulatory/source inquiry; no email was sent because the official site routes such requests to the online reception and `media@cbr.ru` is for journalist requests | NOT CONTACTED BY EMAIL |

The contact addresses above are recorded only as functional/business routing evidence. No Gmail message identifiers, internal mail headers, tracking metadata or private account metadata belong in this repository document.

### Second-wave outreach detail

The following messages were actually sent and remain non-substantive outreach evidence pending reply:

- **NSD / НРД — `datasales@nsd.ru`**: requested coverage for dividends, coupons, maturities, amortizations, offers, record dates and corrections/cancellations; machine-readable access; free/test/research access; public display and derived-analytics rights; caching/history/retention; attribution; rate limits; pricing/minimum tariff; and documentation/schema availability before contract.
- **Interfax — `sales_support@interfax.ru`**: requested structured feeds/API for dividends, coupons, bond events and issuer/Corporate Actions, plus automated retrieval, public display, derived analytics, caching/retention, attribution, pricing and demo/research/startup access.
- **InvestFunds / Cbonds — `database@cbonds.info`**: requested machine-readable/API access for bonds, coupons, maturities, amortizations, Corporate Actions and dividends/issuer events where available, together with use/display/derived/caching/attribution/rate-limit/research/demo/free/pricing terms.
- **Finmarket — `news@finmarket.ru`**: asked whether structured feeds/API or machine-readable financial/Corporate Actions datasets exist, what their use terms are, whether Finmarket can point to an official source, and whether research/information cooperation is possible.
- **Smart-Lab — `admin@smart-lab.ru`**: research/community contact only; asked about relevant structured/API data, legal-source knowledge, research/community cooperation and future beta-user discovery. Smart-Lab is **not** treated as a production data provider.
- **SPB Exchange — `info@spbexchange.ru`**: asked about machine-readable/API data, Corporate Actions/instrument events, automated retrieval, public display, derived analytics, caching/retention, attribution, rate limits, demo/research access and production licensing.
- **Banki.ru — `info@banki.ru`, CC `pr@banki.ru`**: research/information-platform contact; asked about structured investment/Corporate Actions information, legal-source recommendations, cooperation and only future exposure/beta-user opportunities after MVP maturity. Banki.ru is **not** treated as a production provider.
- **RBC Investments — `oanohina@rbc.ru`**: financial-media/research contact; asked about machine-readable structured investment/Corporate Actions feeds, official/legal-source guidance, research/information cooperation and feedback on OpenInvest. RBC Investments is **not** treated as a production provider.

Every row above remains `AWAITING RESPONSE` unless explicitly classified otherwise. Sending the message does not establish source suitability, API availability, production-use rights, public-display rights, derived-use rights, caching rights, redistribution rights or acceptable cost.

## 6. Response received so far

### T-Bank / T-Invest

T-Bank replied from the initial `openapi@tbank.ru` route that the question should be sent to:

`invest-public-api@tbank.ru`

The outreach was sent to that specialist address. T-Bank subsequently registered ticket:

`3-781291`

Current classification:

```text
T-Invest = NEED CLARIFICATION / AWAITING SUBSTANTIVE RESPONSE
```

The redirect and ticket registration are routing evidence only. They are neither `GO` nor `NO-GO`, and they do not establish automated retrieval rights, public display, derived analytics, caching/retention, attribution, rate limits, cost or production-use rights.

### VTB

The initial outreach was sent to:

`info@vtb.ru`

VTB replied that an official request should be submitted in free form to:

`Президенту – Председателю Правления Банка ВТБ (ПАО) Костину Андрею Леонидовичу`

at the postal address:

`109147, г. Москва, ул. Воронцовская, д. 43, стр. 1`

This is not a substantive API/data-rights answer. The evidence dimensions remain:

```text
Automated API access: UNKNOWN
Corporate Actions availability: UNKNOWN
Public display rights: UNKNOWN
Derived analytics: UNKNOWN
Caching/retention: UNKNOWN
Attribution: UNKNOWN
Cost/research access: UNKNOWN
```

A follow-up was sent in the same thread explaining that Asif is outside the Russian Federation and that physical paper delivery is difficult. The follow-up requested either:

- an email for the investment API / financial-data / data-licensing team;
- an electronic channel for a formal request addressed to Andrey Kostin; or
- a link to an official electronic reception channel.

Current classification:

```text
VTB = AWAITING ELECTRONIC ROUTING RESPONSE / NEED CLARIFICATION
```

Reason:

```text
FORMAL OFFICIAL REQUEST REQUIRED / ELECTRONIC CHANNEL REQUESTED
```

### Bank of Russia / CBR

No email was sent to the Bank of Russia for this outreach wave.

The official site routes this type of request to the online reception / formal electronic-request channel. `media@cbr.ru` is intended for journalist requests and is not treated as an appropriate OpenInvest routing address.

Current classification:

```text
CBR = NOT CONTACTED BY EMAIL
Preferred channel = official online reception / formal electronic request
```

This document must not imply that a CBR request has already been sent.

### Other candidates

All other first- and second-wave contacts remain `AWAITING RESPONSE` or, for Smart-Lab / Banki.ru / RBC Investments, `OUTREACH / RESEARCH CONTACT — AWAITING RESPONSE`, unless and until substantive evidence is added through a separately reviewed documentation change.

## 7. Decision rules

Source due diligence uses the following rules:

```text
API existence       != approval
technical access    != production-use rights
public web page     != scraping permission
free access         != redistribution permission
```

Feature 3D may be proposed for implementation only when a **specific source/use mode** has evidence sufficient to establish, at minimum:

| Decision dimension | Required state before Feature 3D authorization |
| --- | --- |
| Automated retrieval | APPROVED |
| Public display | APPROVED |
| Derived analytics | APPROVED |
| Caching / retention | APPROVED or explicitly bounded |
| Attribution | Defined |
| Rate limits / traffic policy | Defined |
| Cost | Accepted |
| Production use | APPROVED |

If any material right remains `UNKNOWN`, the source/use mode remains blocked or requires clarification.

A sent email, a generic support reply, documentation of an API, possession of credentials, or technical connectivity must never be treated as a `Data Source Registry` approval.

## 8. How future replies must be documented

For each substantive reply, add a compact evidence record containing:

- response date;
- organization;
- functional sender/contact;
- concise summary of the official answer;
- automated retrieval: `YES / NO / UNKNOWN`;
- public display: `YES / NO / UNKNOWN`;
- derived analytics: `YES / NO / UNKNOWN`;
- caching: `YES / NO / UNKNOWN`;
- retention: exact restriction or `UNKNOWN`;
- attribution requirement;
- rate limits / traffic policy;
- cost;
- agreement/license requirement;
- evidence status;
- proposed registry verdict:
  - `GO`;
  - `CONDITIONAL GO`;
  - `NO-GO`;
  - `NEED CLARIFICATION`.

A proposed registry verdict in this research document is **not** itself a registry status change. Any actual `Data Source Registry` transition requires the normal separately reviewed governance process and exact source/use evidence.

## 9. Public-repository evidence minimization

Do not copy into the public repository unless materially required for the source decision:

- Gmail message IDs;
- internal email headers;
- personal metadata;
- tracking information;
- confidential/legal footers that are not necessary for the decision;
- full private correspondence when a concise evidence summary is sufficient.

Where a reply contains confidential, contractual or personal material, preserve only the minimum public evidence needed to support the decision. If the decision cannot be supported publicly without disclosing restricted material, record the public status as `UNKNOWN` / `NEED CLARIFICATION` and keep the restricted evidence outside the public repository.

## 10. Relationship to the Data Source Registry

This research log does not modify any existing `Data Source Registry` status.

In particular:

- existing `NO-GO` rows remain `NO-GO`;
- `REVIEW REQUIRED` remains `REVIEW REQUIRED`;
- outreach activity does not create a new approved source row;
- a generic response does not authorize scraping, automation, caching, redistribution, public display or derived use;
- only an exact, evidenced source/use mode may later be proposed for a registry transition.

The second-wave outreach does not change the existing registry verdicts. In particular:

- `NSD_CORPORATE_ACTIONS_API` remains the currently reviewed `NO-GO` source/use mode until substantive new evidence supports a separately reviewed transition;
- `INTERFAX_EDISCLOSURE_API` remains the currently reviewed `NO-GO` e-Disclosure gateway use mode; the new Interfax outreach asks whether other professional structured feeds/use modes exist, but the sent inquiry neither changes that row nor creates an approved alternative;
- no new row is added for InvestFunds/Cbonds, SPB Exchange, Finmarket, Smart-Lab, Banki.ru or RBC Investments because outreach alone is not source approval;
- CBR remains `NOT CONTACTED BY EMAIL` for this outreach wave and the existing registry `CBR_CORPORATE_ACTIONS_FEED` verdict is unaffected.

## 11. Scope exclusions

This document does not authorize or implement:

- Feature 3D provider code;
- external HTTP ingestion;
- provider credentials or secrets;
- runtime composition/wiring;
- OpenAPI changes;
- backend API behavior changes;
- frontend changes;
- database/schema/migrations;
- persistence or caching;
- background workers/polling;
- scraping;
- Data Source Registry status changes;
- public claims of real-market Corporate Actions coverage.

## 12. Current result

```text
Feature 3D:
SOURCE DUE DILIGENCE IN PROGRESS

Approved production Corporate Actions source:
NONE

Runtime/provider activation:
NOT AUTHORIZED
```

The second outreach wave expands the candidate/evidence pipeline but does not remove the source/use-rights gate.

The next evidence action is to record substantive replies as they arrive and evaluate each exact source/use mode against the decision dimensions above. Feature 3D implementation remains blocked until a production-usable source/use mode is evidenced and separately approved.
