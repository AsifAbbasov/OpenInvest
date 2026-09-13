# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 22

# LEGAL ARCHITECTURE, COMPLIANCE, LICENSES, DISCLAIMERS, TERMS OF SERVICE & REGULATORY FRAMEWORK

Version: 1.0

Status: Approved

Priority: CRITICAL

Classification: LEGAL FOUNDATION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_22_LEGAL_ARCHITECTURE_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `735704ea3afe3ba72a823c9315eee34c6c6255fa`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# PURPOSE

This document defines the legal architecture of OpenInvest.

The document's goal is to:

* minimize legal risks;
* prevent licensing violations;
* prevent violations of law;
* protect users;
* protect the project owners;
* enable international scaling.

---

# PRODUCT STATUS

OpenInvest is NOT:

---

A Broker

---

An Investment Company

---

A Trust Manager

---

A Financial Advisor

---

A Tax Agent

---

A Bank

---

An Exchange

---

# PRODUCT POSITION

OpenInvest is:

> An analytical information platform intended for accounting for investment assets, calculating dividends, visualizing portfolios, and preparing supporting tax documents.

---

# MAIN LEGAL PRINCIPLE

The service assists the user.

The service does not make decisions for the user.

---

# ABSOLUTE RULE

OpenInvest must never write:

```text
Buy this stock.

Sell this bond.

This is guaranteed to generate a profit.

You will earn 20%.

This is the best asset.
```

---

# ALLOWED PHRASES

```text
Historical return.

Officially announced dividends.

Historical XIRR.

Forecast based on open data.

Probabilistic model.

Historical statistics.
```

---

# FINANCIAL DISCLAIMER

Every analytics page displays:

> The information is provided solely for informational purposes and does not constitute an individual investment recommendation.

---

# AI DISCLAIMER

AI must use the following behavior model:

---

Explain

---

Interpret

---

Show data

---

Show risks

---

But never advise buying or selling an asset.

---

# TAX DISCLAIMER

The tax module must display:

> The declaration was generated automatically based on data provided by the user and open official sources. The user must verify the correctness of the information before submission.

---

# HUMAN IN THE LOOP

Before exporting XML/PDF:

the user must confirm:

```text
☑ I have checked the data

☑ I understand that responsibility for submitting the declaration rests with me
```

---

# USER AGREEMENT

Codex must create:

---

Terms of Service

---

Privacy Policy

---

Cookie Policy

---

Data Processing Policy

---

AI Policy

---

Tax Assistant Policy

---

Acceptable Use Policy

---

Disclaimer

---

# PRIVACY POLICY

Primary principle:

Data Minimization.

---

We collect:

email;

settings;

portfolio;

transactions.

---

By default, we do NOT collect:

passport;

INN;

address;

place of registration.

---

# OPTIONAL DATA

The user may voluntarily provide:

---

INN

---

Passport

---

Address

---

Phone

---

# STORAGE MODES

## MODE 1

Private

Data is not stored.

It is used only to generate the declaration.

After the operation is completed, it is fully destroyed.

---

## MODE 2

Convenience

The user gives consent.

The data is stored in encrypted form.

---

# DELETE POLICY

The user has the right to:

```text
Export data

↓

Delete profile

↓

Delete personal data

↓

Withdraw consent
```

---

# COOKIE POLICY

Only the following are used:

Session

Security

Language

Theme

Analytics (optional)

---

Prohibited:

implicit advertising tracking.

---

# ANALYTICS

By default:

Privacy Friendly.

---

Do not use:

Fingerprint Advertising

Cross Site Tracking

Hidden Tracking

---

# THIRD PARTY SERVICES

Before connecting any service, it is necessary to verify:

---

license;

---

terms of use;

---

possibility of commercial use;

---

API restrictions;

---

GDPR;

---

local legislation.

---

# MARKET DATA LICENSE

Only the following are used:

official free APIs;

official open data;

data permitted for public use.

---

Prohibited:

mass mirroring of paid services;

copying closed databases;

bypassing licensing.

---

# API COMPLIANCE

All calls to external APIs must:

use caching;

use aggregation;

comply with Rate Limit;

avoid creating excessive load.

---

# SCRAPING POLICY

Only the following is permitted:

officially permitted scraping of open pages.

---

Prohibited:

bot-based authorization;

bypassing protections;

bypassing robots.txt;

using headless browsers to violate site restrictions.

---

# EMAIL COMPLIANCE

All emails must contain:

---

sender;

---

contacts;

---

reason for receiving the email;

---

unsubscribe link;

---

privacy policy.

---

# PUSH NOTIFICATIONS

Never use:

Dark Patterns

Fake Urgency

Manipulation

---

# PREMIUM MODEL

Permitted:

---

advanced analytics;

---

more portfolios;

---

export;

---

advanced reports;

---

AI features.

---

Prohibited:

artificially degrading the basic product.

---

# AI ETHICS

AI must:

explain;

warn;

show uncertainty.

---

AI is prohibited from:

manipulating;

persuading;

creating FOMO;

promising profit.

---

# ACCESSIBILITY

All legal documents must be written:

in plain language;

without complex legal terminology;

with the ability to download a PDF.

---

# INTERNATIONAL SCALING

Documents must support:

---

localization;

---

regional versions;

---

different tax regimes;

---

different currencies.

---

# OPEN SOURCE POLICY

All libraries used must be checked for:

MIT

Apache 2.0

BSD

or compatible licenses.

---

GPL and AGPL libraries may be used only after a separate architectural decision.

---

# USER RIGHTS

Every user has the right to:

view their data;

export it;

correct it;

delete it;

receive information about its use.

---

# INCIDENT DISCLOSURE

In the event of a security incident:

the user is notified;

the scope is described;

the consequences are described;

the measures taken are described.

---

# COMPETITIVE ADVANTAGE

OpenInvest's legal advantage:

---

Minimal data collection.

---

Maximum transparency.

---

No imposed investment recommendations.

---

Full user control over information.

---

Private mode of operation.

---

Ability to use the service completely anonymously (except for the email used for the account).

---

# CODEX REQUIREMENTS

Before creating any new feature, Builder Agent must answer:

1.

Does the feature require new personal data?

2.

Can it be implemented without storing this data?

3.

Does it violate the user agreement?

4.

Does it violate an external API license?

5.

Does it create legal liability for OpenInvest?

6.

Does it look like an investment recommendation?

7.

Does it violate the principle:

> **OpenInvest is an analytical assistant, not a financial advisor.**

If even one answer is negative, the feature is not permitted for implementation without a separate legal and architectural review.
