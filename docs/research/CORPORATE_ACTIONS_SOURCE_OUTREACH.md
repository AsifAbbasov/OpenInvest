# Corporate Actions Source Outreach

| Field | Value |
| --- | --- |
| Status | Research / evidence candidate — source due diligence in progress; no provider or runtime activation authorized |
| Last evidence update | 2026-09-09 |
| Purpose | Evidence log for future Feature 3D — Real Corporate Actions Source Adapter |
| Canonical repository base | `develop@3674d38f60938f64900fe25a8ef255688dbd7e4f` |
| Protected-base tree | `6bc135f13d3309366ddc5d460c69f1efffccfe3c` |
| Governance authority | `docs/REVIEW_WORKFLOW.md` v1.4.0; `docs/registries/DATA_SOURCE_REGISTRY.md`; Stage 3.61 Corporate Actions planning |
| Runtime scope | None |
| Data Source Registry status change | None |
| Feature 3D implementation | NOT STARTED / NOT AUTHORIZED by this document |

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

Today, the canonical Stage 3.62 `CorporateActionEvent` supports only:

- `DIVIDEND`;
- `COUPON`.

Redemptions, amortizations, offers, conversions and other Corporate Actions are useful source-capability evidence for future planning, but are **not** silently authorized as current `CorporateActionEvent` kinds. Mapping or exposing them requires a separately approved domain/API/product contract extension.

## 2. Decision dimensions

For each candidate OpenInvest must establish all material dimensions before production authorization:

1. official machine-readable API/feed coverage;
2. automated retrieval rights;
3. use in the independent research/educational OpenInvest application;
4. public display of normalized data;
5. derived calculations / analytics;
6. normalization, caching, temporary storage and persistence;
7. retention restrictions;
8. attribution requirements;
9. rate limits / traffic policy;
10. free/demo/research/development access where available;
11. production cost;
12. separate agreement/license/approval requirements.

Canonical decision rules remain:

```text
API existence       != approval
technical access    != production-use rights
public web page     != scraping permission
free access         != redistribution permission
support reply       != automatic Data Source Registry approval
published-terms conflict must be explicitly reviewed
```

If any material right remains `UNKNOWN`, or provider evidence conflicts with published terms, the exact source/use mode remains blocked or requires a separately reviewed constrained decision.

## 3. How OpenInvest was positioned

The outreach describes OpenInvest as:

- an independent research and educational project for retail/private investors;
- a tool for portfolio analytics and expected cash flows;
- a Corporate Actions calendar and derived-analytics product;
- not a broker;
- not a bank;
- not an asset manager;
- not an individualized investment-recommendation service;
- intended to have a free public user interface;
- not intended to resell or expose a provider's raw API/feed.

Repository:

`https://github.com/AsifAbbasov/OpenInvest`

## 4. Outreach register — current status

| Organization | Category | Contact / route | Current evidence status |
| --- | --- | --- | --- |
| Moscow Exchange / MOEX | Exchange | `data@moex.com`; CC `itsales@moex.com` | AWAITING RESPONSE |
| Finam | Broker / API | `agent@corp.finam.ru` → Trade API team | ROUTED TO TRADE API TEAM / AWAITING SUBSTANTIVE TECHNICAL RESPONSE |
| Alfa Investments | Broker / bank | `support@alfadirect.ru` | AWAITING RESPONSE |
| BCS | Broker | `info@bcs.ru` | AWAITING RESPONSE |
| T-Invest | Broker / API | `openapi@tbank.ru` → `invest-public-api@tbank.ru`; support ticket `3-781291` | EXACT OPENINVEST USE CONFIRMED BY SUPPORT / PUBLISHED FAQ CONFLICT REQUIRES SOURCE-RIGHTS REVIEW |
| VTB | Broker / bank | `info@vtb.ru`; official postal route supplied by VTB | FORMAL POSTAL REQUEST REQUIRED / CURRENT ELECTRONIC OUTREACH PATH BLOCKED |
| Gazprombank | Broker / bank | `broker@gazprombank.ru` | AWAITING RESPONSE |
| Sber | Broker / bank | `sberbank@sberbank.ru` | AWAITING RESPONSE |
| NSD / НРД | Market infrastructure | `datasales@nsd.ru` | NO-GO UNDER CURRENT PROJECT STATUS — LEGAL ENTITY REQUIRED |
| Interfax / e-disclosure | Professional data vendor / disclosure infrastructure | response from e-disclosure service; test-access follow-up sent | TEST ACCESS OFFERED / TEST REQUEST SENT / RIGHTS AND COST TERMS UNRESOLVED |
| InvestFunds / Cbonds | Professional data vendor | `database@cbonds.info`; response from Cbonds API/Data Feed | TRIAL AVAILABLE / RETRANSMISSION AND PRODUCTION TERMS UNRESOLVED / CLARIFICATION SENT |
| SPB Exchange | Exchange / market infrastructure | `info@spbexchange.ru` | AWAITING RESPONSE |
| Finmarket | Financial information | `news@finmarket.ru` | AWAITING RESPONSE |
| Smart-Lab | Investment community | `admin@smart-lab.ru` | OUTREACH / RESEARCH CONTACT — AWAITING RESPONSE |
| Banki.ru | Financial platform | `info@banki.ru`; CC `pr@banki.ru` | OUTREACH / RESEARCH CONTACT — AWAITING RESPONSE |
| RBC Investments | Financial media | `oanohina@rbc.ru` | OUTREACH / RESEARCH CONTACT — AWAITING RESPONSE |
| Bank of Russia / CBR | Regulator | Official online reception / formal electronic request | NOT CONTACTED BY EMAIL |

The contact addresses above are recorded only as functional/business routing evidence. No Gmail message identifiers, internal mail headers, tracking metadata or private account metadata belong in this repository document.

## 5. Responses and follow-up evidence

### 5.1 T-Bank / T-Invest

The initial `openapi@tbank.ru` route redirected the inquiry to `invest-public-api@tbank.ru`. T-Bank registered ticket `3-781291`.

The first substantive support response established the following provider statement:

- requested data are available;
- use was described as permitted;
- support stated that no additional agreement was required;
- technical details such as limits were referred to the official documentation.

Official documentation confirms strong technical fit for the current/future OpenInvest Corporate Actions scope:

- `GetDividends` — dividend payment events, filtered by record date;
- `GetBondCoupons` — bond coupon schedule;
- `GetBondEvents` — bond events including coupons, offers, maturities/redemptions and conversions;
- Bearer API Token authorization;
- production and sandbox endpoints exist;
- T-Invest API data are described as free for API users.

Relevant official documentation:

- `https://developer.tbank.ru/invest/intro/faq`
- `https://developer.tbank.ru/invest/intro/intro`
- `https://developer.tbank.ru/invest/intro/intro/limits`
- `https://developer.tbank.ru/invest/api/instruments-service-get-dividends`
- `https://developer.tbank.ru/invest/api/instruments-service-get-bond-coupons`
- `https://developer.tbank.ru/invest/api/instruments-service-get-bond-events`

A material conflict was then identified: the public T-Invest API FAQ states that a public service based on T-Invest API is not permitted and describes the API as provided to T-Invest clients without retransmission rights.

Because OpenInvest does not expose the provider's raw API/feed, a targeted clarification was sent on 2026-09-09. The exact OpenInvest scenario described to T-Invest was:

- OpenInvest is a free public research/educational application;
- the OpenInvest backend obtains T-Invest API data;
- OpenInvest normalizes provider data into its own domain model;
- users do not receive the raw provider API/feed or the T-Invest token;
- the public UI displays normalized dividend/coupon/event data;
- the public UI may display derived analytics/calculations based on those normalized data.

T-Invest support reopened ticket `3-781291` and then answered the targeted clarification with the provider statement:

> `Такое использование разрешено.`

This reply is material evidence because it answers the specifically described OpenInvest public-display / normalization / derived-analytics scenario rather than a generic API-availability question.

However, the published FAQ remains textually inconsistent with the support clarification. Therefore this research document records the permission but does **not** silently convert it into a production registry approval.

Current evidence classification:

```text
Technical fit:                         STRONG
Dividend coverage:                    YES
Coupon coverage:                      YES
Bond event coverage:                  YES
Automated API access:                 YES technically
Exact OpenInvest normalized display:  YES — support confirmed exact scenario
Derived analytics in public product:  YES — support confirmed exact scenario
Raw API/feed redistribution:          NOT REQUESTED / NOT PART OF OPENINVEST USE MODE
Caching/retention:                    NOT ESTABLISHED
Attribution:                          NOT ESTABLISHED
Cost for API access:                  FREE per public API documentation
Additional agreement:                 support previously stated NOT REQUIRED
Published FAQ consistency:            CONFLICTING
Production registry approval:         NOT YET GRANTED
```

Current classification:

```text
T-Invest = EXACT OPENINVEST USE CONFIRMED BY SUPPORT / PUBLISHED FAQ CONFLICT REQUIRES SOURCE-RIGHTS REVIEW
```

Source-rights review conclusion for the evidence currently available:

```text
Public production GO:                                  NOT YET
Conditional GO for non-public technical/staging use:   YES
Candidate for constrained production proposal:         YES
```

Reasoning:

- the exact normalized public-display and derived-analytics scenario is positively confirmed by provider support;
- the FAQ still states that public services / retransmission are not permitted;
- the FAQ says the service is governed by the user agreement;
- caching/retention and attribution remain unresolved;
- T-Invest API access requires a client token, so token ownership/rotation and production dependency on a client account must be explicitly designed;
- current `DATA_SOURCE_REGISTRY.md` requires an exact approved source/use row before external-source implementation.

A future constrained production proposal must define, at minimum:

- exact T-Invest methods and data fields used;
- no raw API/feed redistribution;
- server-owned secret handling and token rotation;
- rate-limit policy;
- either explicit caching/retention rights or a `no-store` / no-persistence boundary;
- attribution behavior;
- fail-closed provider availability semantics;
- the governance decision on whether the support clarification is sufficient evidence despite the published FAQ conflict.

Until that separately reviewed proposal is approved, no runtime integration or `Data Source Registry` transition is authorized.

### 5.2 NSD / НРД

A substantive response was received from NSD information-services sales.

The response establishes:

- NSD information services are available only to legal entities;
- NSD does not conclude the relevant data-provision contracts with individuals or individual entrepreneurs;
- if access can be arranged through a legal entity, NSD can check subscriptions and then discuss test access and use conditions.

Current OpenInvest context is an individual-owned, zero-budget research/project setup. Therefore:

```text
NSD = NO-GO UNDER CURRENT PROJECT STATUS — LEGAL ENTITY REQUIRED
```

This is not a permanent technical rejection of NSD. It is a current organizational/contractual blocker.

### 5.3 InvestFunds / Cbonds

Cbonds API/Data Feed supplied a substantive response.

Material evidence received:

- API coverage includes instrument/bond parameters, coupon schedules, offers, defaults and other bond data;
- dividends and splits are available;
- test access can be provided for a limited set of approximately 10–15 ISINs;
- derived analytics/calendar use is treated as retransmission;
- public display depends on the exact third-party access model;
- historical storage/retention requires dataset-specific clarification;
- rate limits communicated: 10,000 requests per method per day and 30 requests per minute;
- API documentation is available before contract;
- there is no separate open-source/research commercial format;
- production price cannot be determined until the retransmission model is specified.

On 2026-09-09 OpenInvest sent a clarification describing the intended product model:

- free public application;
- normalized data only;
- no raw Cbonds feed/API redistribution;
- calendar and derived analytics;
- historical normalized storage only where licensed;
- request for exact attribution, retention, retransmission rights and minimum production price.

Current classification:

```text
Cbonds / InvestFunds = TRIAL AVAILABLE / RETRANSMISSION AND PRODUCTION TERMS UNRESOLVED
Zero-budget production GO = NOT ESTABLISHED
```

### 5.4 Interfax / e-disclosure

The e-disclosure / Interfax-CRKI disclosure service replied that API information is available and explicitly invited OpenInvest to submit a request for test connection/access.

On 2026-09-09 OpenInvest sent a test-access request asking for:

- test connection procedure;
- test credentials/sandbox if available;
- current API documentation and examples;
- Corporate Actions/data coverage;
- rate limits and technical restrictions;
- public display of normalized data;
- derived analytics/calendar rights;
- caching/history/retention rules;
- attribution requirements;
- use in a free public OpenInvest interface without raw-feed resale;
- indicative production cost after the test period.

Current classification:

```text
Interfax / e-disclosure = TEST ACCESS OFFERED / TEST REQUEST SENT / RIGHTS AND COST TERMS UNRESOLVED
Production GO = NOT ESTABLISHED
```

### 5.5 Finam

Finam confirmed receipt of the OpenInvest questions and stated that they would be passed to colleagues responsible for Trade API for a fuller technical answer.

A separate partner/referral offer is unrelated to Corporate Actions/API data rights and is not treated as source approval.

Current classification:

```text
Finam = ROUTED TO TRADE API TEAM / AWAITING SUBSTANTIVE TECHNICAL RESPONSE
```

### 5.6 VTB

VTB replied that an official request must be submitted in free form to the President — Chairman of the Management Board of VTB Bank (PJSC), at the postal address provided by VTB:

`109147, г. Москва, ул. Воронцовская, д. 43, стр. 1`

OpenInvest requested an electronic route because international paper delivery is inconvenient. VTB repeated that the official request is to be submitted to the bank's postal address.

All substantive source-right dimensions remain unknown:

```text
Automated API access: UNKNOWN
Corporate Actions availability: UNKNOWN
Public display rights: UNKNOWN
Derived analytics: UNKNOWN
Caching/retention: UNKNOWN
Attribution: UNKNOWN
Cost/research access: UNKNOWN
```

A paper-request text has been prepared for manual signing/postal submission. It is **not** recorded as sent until physical dispatch actually occurs.

Current classification:

```text
VTB = FORMAL POSTAL REQUEST REQUIRED / CURRENT ELECTRONIC OUTREACH PATH BLOCKED
```

### 5.7 Bank of Russia / CBR

No email was sent to the Bank of Russia for this outreach wave.

The official route remains the online reception / formal electronic-request channel. `media@cbr.ru` is not treated as the appropriate OpenInvest data-rights route.

Current classification:

```text
CBR = NOT CONTACTED BY EMAIL
Preferred channel = official online reception / formal electronic request
```

### 5.8 Still awaiting substantive response

As of 2026-09-09 the following remain awaiting response:

```text
Moscow Exchange / MOEX
Alfa Investments
BCS
Gazprombank
Sber
SPB Exchange
Finmarket
Smart-Lab
Banki.ru
RBC Investments
```

Silence is not classified as rejection and does not establish any source/use right.

Current follow-up strategy:

```text
initial outreach
    ↓
allow several working days
    ↓
one concise follow-up for priority provider candidates
    ↓
if still unanswered after a further reasonable window:
NO RESPONSE / RIGHTS UNRESOLVED / RUNTIME NO-GO
```

Media/community contacts do not block provider selection.

## 6. Current provider-selection tracks

```text
TRACK A — leading rights candidate
T-Invest
→ technical fit confirmed
→ exact OpenInvest normalized public-display / derived-analytics use confirmed by support
→ published FAQ conflict remains
→ constrained source/use proposal is now justified
→ no production runtime activation before registry/governance approval

TRACK B — active provider evaluation
Interfax / e-disclosure
→ test access offered
→ test-access request sent
→ evaluate API/data quality
→ establish licensing/cost
→ GO or NO-GO proposal only after full evidence

TRACK C — waiting / one-follow-up strategy
MOEX / Finam / Alfa / BCS / Gazprombank / Sber / SPB Exchange
→ wait for substantive response
→ one follow-up where appropriate
→ classify exact source/use mode

SEPARATE COMMERCIAL / TRIAL TRACK
Cbonds / InvestFunds
→ trial available
→ clarification sent on retransmission, public display, retention, attribution and price

FORMAL ROUTING TRACK
VTB
→ paper request prepared
→ physical dispatch pending
```

## 7. Response-status summary

```text
Substantive or routing responses received:
- T-Invest
- NSD / НРД
- Cbonds / InvestFunds
- Interfax / e-disclosure
- Finam
- VTB

Follow-ups sent on 2026-09-09:
- T-Invest — retransmission/public-service clarification
- Interfax / e-disclosure — test-access and rights/cost request
- Cbonds / InvestFunds — retransmission/public-display/retention/price clarification

Follow-up answered on 2026-09-09:
- T-Invest — exact OpenInvest use confirmed as permitted by support

Still awaiting substantive response:
- MOEX
- Alfa Investments
- BCS
- Gazprombank
- Sber
- SPB Exchange
- Finmarket
- Smart-Lab
- Banki.ru
- RBC Investments

Prepared but not yet sent physically:
- VTB formal postal request
```

## 8. Required evidence before Feature 3D authorization

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

A sent email, generic support reply, API documentation, possession of credentials, sandbox access or technical connectivity must never be treated as `Data Source Registry` approval.

For T-Invest, the exact-use clarification materially strengthens public-display and derived-analytics evidence but does not by itself resolve caching/retention, attribution, operational token policy or the published FAQ inconsistency.

## 9. Public-repository evidence minimization

Do not copy into the public repository unless materially required for the source decision:

- Gmail message IDs;
- internal email headers;
- personal metadata;
- tracking information;
- confidential/legal footers that are not necessary for the decision;
- full private correspondence when a concise evidence summary is sufficient.

Where a reply contains confidential, contractual or personal material, preserve only the minimum public evidence needed to support the decision.

## 10. Relationship to the Data Source Registry

This research log does not modify any existing `Data Source Registry` status.

In particular:

- existing `NO-GO` rows remain `NO-GO`;
- `REVIEW REQUIRED` remains `REVIEW REQUIRED`;
- outreach activity does not create a new approved source row;
- test/sandbox access does not create production approval;
- only an exact, evidenced source/use mode may later be proposed for a registry transition.

Specific current boundaries:

- `NSD_CORPORATE_ACTIONS_API` remains the reviewed `NO-GO` source/use mode; the new reply additionally confirms the current legal-entity blocker;
- `INTERFAX_EDISCLOSURE_API` remains the reviewed `NO-GO` e-Disclosure gateway use mode until new test/licensing evidence supports a separately reviewed proposal;
- T-Invest now has provider support evidence for the exact OpenInvest normalized public-display / derived-analytics scenario, but **no registry row or production approval is created by this research log**;
- the T-Invest published FAQ conflict, caching/retention, attribution and token/runtime constraints must be addressed in a separately reviewed source/use proposal;
- no new source is approved for Cbonds/InvestFunds, SPB Exchange, Finmarket, Smart-Lab, Banki.ru or RBC Investments;
- the existing MOEX/source-rights NO-GO remains unaffected by silence on the new outreach.

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
- public claims of real-market Corporate Actions coverage;
- Stage 3.78 or later product/runtime scope.

## 12. Current result

```text
Feature 3D:
SOURCE DUE DILIGENCE IN PROGRESS

Approved production Corporate Actions source:
NONE

Leading source-rights candidate:
T-INVEST — EXACT OPENINVEST USE CONFIRMED BY SUPPORT

T-Invest source-rights review:
PUBLIC PRODUCTION GO = NOT YET
NON-PUBLIC TECHNICAL/STAGING CONDITIONAL GO = YES
CONSTRAINED PRODUCTION PROPOSAL = JUSTIFIED

Open T-Invest evidence items:
PUBLISHED FAQ CONFLICT / CACHE-RETENTION / ATTRIBUTION / TOKEN POLICY

Parallel active evaluation:
INTERFAX / E-DISCLOSURE TEST-ACCESS TRACK

Runtime/provider activation:
NOT AUTHORIZED

Data Source Registry transition:
NONE

Stage 3.78+:
NOT STARTED / NOT AUTHORIZED BY THIS DOCUMENT
```

The evidence pipeline is active rather than purely awaiting replies. T-Invest now has the strongest current source-rights evidence for the exact OpenInvest public normalized-data / derived-analytics scenario. The next justified step is a separately reviewed constrained T-Invest source/use proposal; Feature 3D production implementation remains blocked until that proposal is approved.