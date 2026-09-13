# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 25

# PRODUCT ANALYTICS, FEATURE FLAGS, OBSERVABILITY, COST OPTIMIZATION & PRODUCT EVOLUTION

Version: 1.0

Status: APPROVED

Priority: CRITICAL

Classification: PRODUCT OPERATIONS

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_25_PRODUCT_ANALYTICS_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `f89dba3c18fc0e58fe83e9c7395c6f7f0561823f`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# PURPOSE

This document defines the product-evolution system after launch.

OpenInvest should continuously improve based on real user behavior,

while at the same time:

* not violating privacy;
* not collecting unnecessary data;
* not increasing infrastructure costs;
* not degrading performance.

---

# PRODUCT PHILOSOPHY

We measure:

product behavior,

not human behavior.

---

# ABSOLUTE RULE

The following are prohibited:

Fingerprint Tracking

Cross Device Tracking

Advertising Tracking

Hidden Analytics

Selling User Data

Recording User Inputs

Mouse Recording

Session Replay

---

# ALLOWED ANALYTICS

Only anonymized events may be collected.

---

Example:

```
PortfolioOpened

AssetCardViewed

DividendCalendarOpened

TaxExportCreated

PortfolioCreated

TransactionAdded
```

---

# EVENT MODEL

Each event has the structure:

```
EventID

Timestamp UTC

Platform

Version

Feature

Duration

Success

ErrorCode
```

---

None of the following:

Email

INN

Passport

IP

GPS

is stored.

---

# PRODUCT METRICS

The following are measured:

---

Daily Active Users

---

Weekly Active Users

---

Monthly Active Users

---

Retention

---

Crash Rate

---

Error Rate

---

Feature Usage

---

Average Session

---

Average API Latency

---

# NORTH STAR METRIC

The project's primary metric:

> The user received understandable information about their capital in less than 5 seconds.

---

# FEATURE FLAGS

Every new function must be switchable off.

---

Example:

```
DividendForecast

RealInflation

TaxAssistant

AIHelper

Notifications

PortfolioCompare
```

---

# FEATURE ROLLOUT

A new function is enabled through:

```
Develop

↓

Internal

↓

1%

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

# ROLLBACK

Any function must be switchable off

without a new deployment.

---

# OBSERVABILITY

Every service must publish:

Health

Memory

CPU

Latency

Errors

Queue

Database

---

# HEALTH ENDPOINT

Every service must have:

```
/health

/ready

/live
```

---

# DASHBOARD

A unified dashboard is created for:

---

Frontend

---

Go API

---

Python Worker

---

Redis

---

PostgreSQL

---

SMTP

---

Background Jobs

---

# BUSINESS DASHBOARD

Shows:

---

Number of users

---

Number of portfolios

---

Number of transactions

---

Number of XML files

---

Number of PDF files

---

Number of Emails

---

# PERFORMANCE DASHBOARD

Shows:

---

Average Response

---

P95

---

P99

---

Memory Usage

---

CPU Usage

---

Network

---

Cache Hit

---

# DATABASE DASHBOARD

Shows:

---

Connections

---

Slow Queries

---

Locks

---

Indexes

---

Table Size

---

Growth

---

# CACHE DASHBOARD

Redis

or

RAM Cache

---

Shows:

Hit Ratio

Miss Ratio

TTL

Objects

Memory

---

# COST DASHBOARD

Displays:

```
Hosting

Database

Bandwidth

Storage

Email

Monitoring

AI

Total
```

---

# MONTHLY LIMITS

If expenses exceed the limit,

Builder Agent receives a task:

> reduce cost without degrading UX.

---

# AUTO OPTIMIZATION

If:

```
Cache Hit < 80%
```

or

```
Average Response > 250 ms
```

an Optimization Task is created.

---

# SLOW QUERY DETECTOR

All SQL queries:

> 100 ms

are automatically included in the report.

---

# UNUSED CODE DETECTOR

Once a week, the following are analyzed:

---

Unused Components

---

Unused Hooks

---

Unused APIs

---

Unused Tables

---

Unused Indexes

---

Unused Dependencies

---

# BUNDLE SIZE

Maximum size:

```
Main Bundle

250 KB gzip
```

---

Feature Bundle

Lazy Loaded

---

# LAZY STRATEGY

Do not load:

Tax Module

Charts

AI

PDF

until opened.

---

# API VERSIONING

All APIs:

```
/api/v1/

/api/v2/
```

---

Prohibited:

breaking older clients.

---

# MIGRATION STRATEGY

A new version:

operates alongside the old one

for at least 6 months.

---

# A/B TESTS

Permitted only for:

UX

Layout

Colors

Navigation

---

Prohibited for:

financial calculations;

taxes;

dividends.

---

# REAL USER MONITORING

The following are measured:

---

Load Time

---

Render Time

---

API Time

---

Crash

---

Freeze

---

Without recording user actions.

---

# AI ANALYTICS

AI analyzes:

---

which screens are opened;

---

which functions are ignored;

---

where users leave.

---

AI does not analyze:

personal data;

financial amounts;

tax documents.

---

# PRODUCT EVOLUTION

RoadMap is built according to the principle:

```
User Problem

↓

Evidence

↓

Architecture Review

↓

Prototype

↓

Feature Flag

↓

Release

↓

Metrics

↓

Decision
```

---

# SUCCESS CRITERIA

One year after launch, the product should preserve:

---

Response Time

<150 ms

---

Memory

<512 MB backend

---

Cache Hit

> 90%

---

Crash Free Sessions

> 99.8%

---

API Availability

> 99.95%

---

Security Incidents

0

---

Personal Data Leaks

0

---

# COMPETITIVE GOAL

We do not compete on the number of features.

We compete on:

* speed;
* simplicity;
* transparency;
* trust;
* energy efficiency;
* operating cost.

---

# FINAL PRODUCT PRINCIPLE

> **Every new function must make the product better for the user, not more complex for the developer or prettier for a presentation.**

> **The best product is one that remains fast, inexpensive, understandable, and reliable even after millions of transactions and hundreds of thousands of users appear.**
