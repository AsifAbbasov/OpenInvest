# Interfax / e-disclosure — Rights and Cost Evidence

| Field | Value |
| --- | --- |
| Status | Research / evidence update — use rights clarified; zero-budget production NO-GO remains |
| Date | 2026-09-09 |
| Provider | Interfax-CRKI / e-disclosure API gateway |
| Contact route | `dc2@e-disclosure.ru` |
| Canonical base | `develop@1f2fe1cc9c4c6ab282f56506d64a9754b62eef58` |
| Registry authority | `docs/registries/DATA_SOURCE_REGISTRY.md` |
| Runtime activation | NOT AUTHORIZED |
| Budget rule | 0 RUB for external data sources |

## 1. Purpose

This document records the latest provider evidence for the Interfax / e-disclosure API after OpenInvest requested clarification on public display, derived analytics, retention, archive access, contract terms and production cost.

It is evidence-only. It does not authorize runtime activation, payment, contract execution, credentials, provider wiring, scraping or a Data Source Registry status transition.

## 2. Provider clarification received on 2026-09-09

Interfax-CRKI stated that the information available on the e-disclosure site is public and that the API subscription essentially purchases automated access to that information.

The provider explicitly stated that the usage variants listed by OpenInvest are permitted and that no additional paperwork is required for those usage variants beyond the contractual/API access arrangement.

The OpenInvest request had explicitly asked about:

- public display of normalized data;
- derived analytics and a corporate-actions calendar;
- caching / historical storage;
- use in a free public interface without resale of the raw feed.

Therefore the current evidence supports the following provider-side rights classification:

```text
Public normalized display:     CONFIRMED PER PROVIDER RESPONSE
Derived analytics / calendar:  CONFIRMED PER PROVIDER RESPONSE
Historical retention:          CONFIRMED — no deletion requirement stated
Raw-feed resale:               NOT REQUESTED / NOT PART OF OPENINVEST USE MODE
Additional use-right paperwork: NOT REQUIRED PER PROVIDER RESPONSE
```

This evidence does not remove the need for a subscription contract to obtain automated API access.

## 3. Retention and archive evidence

The provider stated that the contract does not require deletion of information already received. Therefore no provider-side retention deletion requirement is currently established for data lawfully received under the contract.

Default access covers events created during the active contract period.

Optional historical archive access is available back to:

`2020-07-01`

The provider stated that archive access adds:

`+50% to the monthly tariff`

This archive option is not approved for OpenInvest under the current zero-budget constraint.

## 4. Contract and minimum term

The provider stated that a contract is required for API access and asked OpenInvest to choose the required information package:

- messages only;
- files only;
- messages + files.

The minimum subscription term is:

`3 months`

OpenInvest's current technical need is primarily structured messages / disclosure events. Files are not required for the initial Corporate Actions integration candidate unless future technical evidence proves they are necessary.

## 5. Existing reviewed tariff evidence

The canonical Data Source Registry already records the currently reviewed public tariff evidence:

```text
message-publication data: 16,180 RUB / month, excluding VAT
minimum subscription:      3 months
complete publications:     27,000 RUB / month
archive option:             +50% / month per provider clarification
```

At the current reviewed message-data tariff, the minimum three-month base commitment would be:

```text
16,180 × 3 = 48,540 RUB excluding VAT
```

This arithmetic is a planning calculation from the currently reviewed tariff. It is not a new provider quotation and must be replaced by a direct provider quotation if the provider confirms different/current pricing.

## 6. Zero-budget decision

OpenInvest has a strict current external-data budget:

`0 RUB`

Therefore, regardless of the improved source-rights evidence:

```text
Technical suitability:        PROMISING / SUBJECT TO DATA-FIELD VALIDATION
Public display rights:        CONFIRMED PER PROVIDER RESPONSE
Derived analytics rights:     CONFIRMED PER PROVIDER RESPONSE
Retention rights:             CONFIRMED — no deletion requirement stated
Contract required:            YES
Minimum term:                 3 MONTHS
Known reviewed tariff:        > 0 RUB
Zero-budget production GO:    NO
Runtime activation:           NOT AUTHORIZED
```

The canonical registry decision remains correct:

`NO-GO — zero-budget constraint; no API subscription and no public-site scraper authorized`

The improved rights evidence does not justify spending money or changing the project budget.

## 7. Pricing clarification sent on 2026-09-09

OpenInvest sent a follow-up asking Interfax-CRKI to confirm the exact current commercial terms for:

1. messages-only API subscription per month;
2. minimum total cost for the mandatory 3-month term;
3. messages + files pricing;
4. archive-to-2020-07-01 pricing including the +50% surcharge;
5. VAT treatment;
6. any one-time connection/setup charge;
7. availability and duration of free test access;
8. any free or discounted open-source, research, educational or non-commercial plan.

The follow-up explicitly stated that OpenInvest currently has a strict 0 RUB external-data budget and therefore cannot activate a paid production API under the present project constraint.

It also asked the provider to reconfirm that the prior answer covers:

- public normalized display;
- derived analytics/calendar;
- retention without deletion requirement.

## 8. Current classification

As of 2026-09-09:

```text
Interfax / e-disclosure

Technical API candidate:                 YES
Public normalized-display rights:        CONFIRMED PER PROVIDER RESPONSE
Derived analytics/calendar rights:       CONFIRMED PER PROVIDER RESPONSE
Retention/deletion restriction:          NO DELETION REQUIREMENT STATED
Archive availability:                    TO 2020-07-01, +50% MONTHLY TARIFF
Contract required:                       YES
Minimum contract term:                   3 MONTHS
Exact fresh provider quote:              REQUESTED / AWAITING RESPONSE
Free test/research regime:               REQUESTED / AWAITING RESPONSE
Zero-budget production GO:               NO
Data Source Registry status transition:  NONE REQUIRED — EXISTING NO-GO REMAINS
Runtime/provider activation:             NOT AUTHORIZED
```

## 9. What would change the decision

A production decision may be reconsidered only if one of the following occurs:

- the project budget constraint is explicitly changed by the owner; or
- Interfax-CRKI provides a genuinely free production/research/open-source access mode compatible with the exact OpenInvest use case.

A free test period alone is not production approval.

Until then, the provider remains a technically and rights-wise useful benchmark / future candidate but financially unavailable for OpenInvest production.
