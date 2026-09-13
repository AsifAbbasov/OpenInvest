# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 38

# PRODUCT ANALYTICS, BUSINESS METRICS, TELEMETRY, FEATURE FLAGS, A/B TESTING & DATA DRIVEN PRODUCT DEVELOPMENT

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: PRODUCT ANALYTICS CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_38_PRODUCT_ANALYTICS_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `397e3989bf55549a6a0fba840a98f346eef2d79a`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the principles of OpenInvest product analytics.

The main goal of analytics is to improve the product, not to collect information about users.

OpenInvest follows the principle:

> **Privacy First Analytics**

---

# PRODUCT PHILOSOPHY

We measure:

---

product quality;

---

interface usability;

---

operating speed;

---

the user's understanding of their own capital.

---

We do not measure:

---

personal financial information;

---

passwords;

---

passport data;

---

TIN;

---

the capital amount of a specific user.

---

# NORTH STAR METRIC

The main product metric

is not the number of users.

---

The main metric:

```text
Users Who Better Understand Their Capital
```

---

The user should return regularly

not to check quotes,

but to understand their financial state.

---

# PRODUCT METRICS

## Acquisition

---

Registration Rate

---

Portfolio Created

---

First Transaction

---

First Dividend View

---

Tax Export

---

# ACTIVATION

A user is considered activated

if:

```text
Registration

↓

Portfolio Created

↓

Transaction Added

↓

Dashboard Viewed
```

---

# RETENTION

Measured:

---

Day 1

---

Day 7

---

Day 30

---

Day 90

---

Year 1

---

# ENGAGEMENT

Average time:

---

Dashboard

---

Portfolio

---

Asset Card

---

Dividend Calendar

---

Analytics

---

Tax

---

# FEATURE ADOPTION

Every feature has its own analytics.

---

Example:

```text
Portfolio

↓

Opened

↓

Scrolled

↓

Chart Opened

↓

Dividend Card

↓

Tax Button
```

---

# SEARCH ANALYTICS

Measured:

---

ticker searches;

---

company searches;

---

bond searches.

---

The user's financial decisions are not stored.

---

# DASHBOARD ANALYTICS

Measured:

---

viewing Real Value;

---

viewing Inflation;

---

viewing Dividend Forecast;

---

viewing XIRR.

---

# CHART ANALYTICS

Measured:

---

period change;

---

zooming;

---

opening details.

---

# UX METRICS

Time to Portfolio

---

Time to Transaction

---

Time to Dividend

---

Time to Tax Export

---

# CLICK DEPTH

Any action must be completed

in no more than

3 clicks.

---

If the average value is higher,

a Product Issue is created.

---

# AB TESTING

It is allowed to test:

---

card placement;

---

CTA color;

---

Dashboard structure;

---

new charts.

---

It is prohibited to test:

---

mathematical formulas;

---

tax calculations;

---

financial indicators;

---

security.

---

# FEATURE FLAGS

Any new feature:

```text
OFF

↓

Internal

↓

Beta

↓

5%

↓

20%

↓

50%

↓

100%
```

---

# USER SEGMENTS

Anonymous

---

Registered

---

Active

---

Premium

---

Beta

---

Internal

---

# PRIVACY ANALYTICS

All events are anonymized.

---

The following is used:

UUID

---

It is prohibited to use:

Email

Phone

TIN

Passport

---

# EVENT NAMING

Unified format:

```text
portfolio.open

portfolio.create

transaction.add

asset.search

calendar.open

tax.export
```

---

# EVENT VERSIONING

Every event has:

---

Version

---

CreatedAt

---

Schema

---

Owner

---

# ANALYTICS PIPELINE

```text
Client

↓

Validation

↓

Queue

↓

Aggregation

↓

Storage

↓

Dashboard
```

---

Raw events

are not used directly.

---

# PRODUCT DASHBOARD

Product Team sees:

---

DAU

---

WAU

---

MAU

---

Retention

---

Activation

---

Errors

---

Crash Rate

---

Latency

---

Feature Adoption

---

# ENGINEERING DASHBOARD

Shows:

---

API Latency

---

Cache Hit

---

CPU

---

Memory

---

Database

---

Queue

---

# BUSINESS DASHBOARD

Shows:

---

Registrations

---

Active Portfolios

---

Premium Conversion

---

Feature Usage

---

Tax Exports

---

AI Usage

---

# COST DASHBOARD

Shows:

---

Server Cost

---

Database Cost

---

Storage Cost

---

Email Cost

---

API Cost

---

LLM Cost

---

Cost per Active User

---

# AI ANALYTICS

Measured:

---

Answer Time

---

Confidence

---

Feedback

---

Regeneration

---

Human Approval

---

# ERROR ANALYTICS

Every error receives:

---

Severity

---

Frequency

---

Affected Users

---

Regression Status

---

# USER FEEDBACK

Built-in system:

---

👍

---

👎

---

Comment

---

Without required text.

---

# HEATMAPS

May be used

only after the user's explicit consent.

---

Disabled by default.

---

# SESSION RECORDING

Prohibited by default.

---

May be enabled

only in Beta Program

and only after separate consent.

---

# DATA RETENTION

Analytics

24 months

---

Aggregated Metrics

without limit

---

Raw Events

90 days

---

# GDPR / PRIVACY READY

The user can:

---

disable analytics;

---

download analytics;

---

delete analytics;

---

view analytics history.

---

# SUCCESS METRICS

OpenInvest is considered successful

if the user:

---

understands their capital better;

---

returns regularly;

---

does not need third-party calculators;

---

does not spend time on manual calculations.

---

# REVIEW CHECKLIST

Before adding a new event, it is necessary to answer:

---

Can it be avoided?

---

Can it be aggregated immediately?

---

Can personal data be removed?

---

Can the storage volume be reduced?

---

Can an existing event be used?

---

# FINAL PRODUCT PRINCIPLE

> **OpenInvest measures product quality, not the user's life.**

> **Analytics exists exclusively to improve user experience, performance, and system reliability.**

> **Any data collection that does not provide direct benefit to the user or the product is considered excessive and must be rejected at the Product Review stage.**
