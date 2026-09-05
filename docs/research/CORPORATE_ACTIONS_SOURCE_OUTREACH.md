# Corporate Actions Source Outreach

| Field | Value |
| --- | --- |
| Status | Research / evidence candidate — source due diligence in progress; no provider or runtime activation authorized |
| Date | 2026-09-05 |
| Purpose | Evidence log for future Feature 3D — Real Corporate Actions Source Adapter |
| Canonical repository base | `develop@c885f6e57ea08e4583103fe2f22f142bf13a8560` |
| Protected-base tree | `53f823e92b02721211e1d89f8af6374fffc252ae` |
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
   - redemptions;
   - amortizations;
   - other Corporate Actions;
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

## 5. Outreach register

| Candidate | Contact | Purpose / note | Current evidence status |
| --- | --- | --- | --- |
| Finam | `agent@corp.finam.ru` | Finam Trade API / Corporate Actions / use rights | AWAITING RESPONSE |
| BCS | `info@bcs.ru` | API/data licensing request; redirect to the responsible team if needed | AWAITING RESPONSE |
| Alfa Investments | `support@alfadirect.ru` | Investment API / Corporate Actions / use rights | AWAITING RESPONSE |
| T-Invest | `openapi@tbank.ru` | Initial API contact | NEED CLARIFICATION |
| T-Invest Public API | `invest-public-api@tbank.ru` | Specialist address provided by T-Bank in reply; outreach re-sent there | NEED CLARIFICATION |
| VTB | `info@vtb.ru` | Investment API / data-team inquiry | AWAITING RESPONSE |
| Sber | `sberbank@sberbank.ru` | Official general contact; request to route to SberInvestments/API/data-licensing team | AWAITING RESPONSE |
| Gazprombank | `broker@gazprombank.ru` | Brokerage direction / Corporate Actions data | AWAITING RESPONSE |
| Moscow Exchange | `data@moex.com` | Market/data licensing inquiry | AWAITING RESPONSE |
| Moscow Exchange | `itsales@moex.com` | Information/data-services commercial contact; copied on the inquiry | AWAITING RESPONSE |

The contact addresses above are recorded only as functional/business routing evidence. No Gmail message identifiers, internal mail headers, tracking metadata or private account metadata belong in this repository document.

## 6. Response received so far

### T-Bank / T-Invest

At the time of this documentation snapshot, T-Bank replied from the initial `openapi@tbank.ru` route that the question should be sent to:

`invest-public-api@tbank.ru`

The outreach was then sent to that specialist address.

Current classification:

```text
T-Invest = NEED CLARIFICATION
```

This is neither `GO` nor `NO-GO`. No substantive answer has yet been recorded regarding automated retrieval, public display, derived analytics, caching/retention, attribution, rate limits, cost, or production-use rights.

All other outreach candidates remain:

```text
AWAITING RESPONSE
```

unless and until evidence is added to this document through a separately reviewed documentation change.

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

The next evidence action is to record substantive replies as they arrive and evaluate each exact source/use mode against the decision dimensions above. Feature 3D implementation remains blocked until a production-usable source/use mode is evidenced and separately approved.
