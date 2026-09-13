# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 18

# SYSTEM ARCHITECTURE, EXTERNAL API INTEGRATION, RATE LIMITS, SCALABILITY & RESILIENCE

Version: 1.0

Status: Approved

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_18_SYSTEM_INTEGRATION_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `e0d66e8d7b25ce262b2beedb71eef381fe92637a`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# PURPOSE

This document describes:

* the interaction architecture of all services;
* working with free official APIs;
* protection against blocking;
* scaling to hundreds of thousands of users;
* queue operation;
* caching;
* the resilience strategy.

This is one of the most important documents in the project.

---

# GLOBAL ARCHITECTURE

```
                        Web React
                             │
                             │
                    iOS (SwiftUI)
                             │
                             │
                 Android (Jetpack Compose)
                             │
──────────────────────── API Gateway ────────────────────────
                             │
                 Go Fiber Main Backend
                             │
        ┌────────────────────┼─────────────────────┐
        │                    │                     │
 Portfolio Service     Tax Service        Notification Service
        │                    │                     │
        └────────────────────┼─────────────────────┘
                             │
                    Internal Event Bus
                             │
      ┌──────────────┬──────────────┬──────────────┐
      │              │              │
 Python Parser   Scheduler     Background Workers
      │              │              │
      └──────────────┼──────────────┘
                     │
               PostgreSQL
                     │
                     │
                  Redis Cache
```

---

# API FIRST

Every client:

React

iOS

Android

will work through a single API.

No separate backends.

---

# SINGLE BACKEND

```text
One Backend

Many Clients
```

---

# SOURCE OF MARKET DATA

Only the following are used:

official

or

free

sources.

---

# ALLOWED SOURCES

MOEX ISS

Bank of Russia

Rosstat

e-disclosure

official issuer websites

---

# FORBIDDEN SOURCES

Scraping personal accounts

Paid APIs without a license

Bypassing website protections

Violating robots.txt

---

# REQUEST STRATEGY

The biggest mistake in projects like this is that

each user makes their own request.

---

We will NOT do that.

---

# BAD SCENARIO

10000 users

↓

10000 requests

↓

MOEX

↓

IP Block

↓

The project is dead

---

# OUR SCENARIO

10000 users

↓

Go Backend

↓

1 request

↓

MOEX

↓

Redis

↓

10000 users receive one response

---

# REQUEST AGGREGATION

All identical requests are combined.

---

For example

1000 users opened Sber.

Backend:

makes

ONE

request.

---

# CACHE STRATEGY

Level 1

RAM

(sync.Map)

---

Level 2

Redis

---

Level 3

PostgreSQL

---

# CACHE TTL

Market Data

5 minutes

---

Dividends

24 hours

---

Inflation

30 days

---

Currency

24 hours

---

Company Profile

30 days

---

# STALE CACHE

If the API is unavailable,

the last successful response is used.

---

The user sees

```text
Last updated:

15:05 UTC
```

---

# FAILOVER

If the source does not respond,

the Backend

does not fail.

---

The following is used:

Cached Version

---

# CIRCUIT BREAKER

After

5

failed requests,

the Backend stops calling the API

for

5 minutes.

---

# RETRY POLICY

Retry

1

↓

Retry

2

↓

Retry

3

↓

Cache

↓

Stop

---

# RATE LIMITS

The internal Backend

must not exceed

the allowed MOEX limits.

---

# REQUEST DISTRIBUTION

All background updates

occur

through a queue.

---

# BACKGROUND QUEUES

Update Prices

Update Dividends

Update Inflation

Update FX

Generate Tax

Generate PDF

Send Email

---

# EVENT BUS

All services communicate

through events.

---

Example:

```
DividendUpdated

↓

PortfolioRecalculate

↓

TaxRecalculate

↓

NotificationCreate

↓

EmailSend
```

---

# USER REQUEST FLOW

```
User

↓

React

↓

API

↓

Redis

↓

(if miss)

↓

PostgreSQL

↓

(if miss)

↓

External API

↓

Save

↓

Return
```

---

# NEVER

Frontend

never

calls

MOEX.

---

# NEVER

Mobile

never

calls

an external API.

---

# SCALABILITY

Stage 1

1000 users

one server

---

Stage 2

10000 users

Go + Redis

---

Stage 3

100000 users

Load Balancer

↓

Go x4

↓

Redis

↓

PostgreSQL

---

Stage 4

1000000 users

Microservices

↓

Queue

↓

Read Replica

↓

CDN

---

# DATABASE LOAD

The Backend must never build charts

from millions of transactions.

---

The following are used:

Snapshots

Materialized Views

Aggregations

---

# API VERSIONING

```
/api/v1/

/api/v2/

/api/v3/
```

---

Never break an older version.

---

# PAGINATION

All lists

must have:

limit

offset

cursor

---

# SORTING

Server Side.

---

Frontend

does not sort

10000 records.

---

# SEARCH

Debounce

300ms

---

Server Search

---

# EXPORT

PDF

XML

ZIP

CSV

are created

asynchronously.

---

The user receives:

```
The document is ready.

Download

or

Send by Email.
```

---

# EMAIL

SMTP Queue

Retry

Dead Letter Queue

---

# PUSH

Batch

Aggregation

Smart Notification

---

# LOGGING

Every external request:

Source

Latency

Status

Retry Count

Cache Hit

---

# METRICS

Response Time

Cache Hit Ratio

DB Time

Queue Size

CPU

RAM

External API Failures

---

# OBSERVABILITY

Prometheus

Grafana

OpenTelemetry

---

# HEALTH CHECK

Every service must have:

```
/health

/ready

/live
```

---

# SECURITY

All external APIs

operate

only through the Backend.

---

API Keys

never

reach

the Frontend.

---

# DOS PROTECTION

Rate Limit

per User

per IP

per Token

---

# BOT PROTECTION

Invisible CAPTCHA

Behavior Analysis

Device Fingerprint

---

# COST OPTIMIZATION

The project's main goal is:

to serve

10000+

users

at minimal cost.

---

# TARGET COST

Frontend

≈ free (Vercel)

---

Backend

≈ minimum plan

---

Redis

minimum

---

PostgreSQL

Serverless

---

Python Worker

Serverless Cron

---

# FUTURE MOBILE

iOS

Android

use

the same API.

---

No new logic.

---

No new calculations.

---

Only a new UI.

---

# COMPETITIVE ADVANTAGES

Compared with most brokers:

✅ The Backend caches data.

✅ Minimal mobile data usage.

✅ Minimal battery usage.

✅ No constant requests.

✅ No heavy calculations on the phone.

✅ No direct exchange calls.

✅ One request serves thousands of users.

✅ Fully scalable API-First architecture.

---

# CODEX REQUIREMENTS

Before integrating any new external API, Builder Agent must verify:

1. Is the source official?

2. Is the source free?

3. Does it allow mass use?

4. Is documentation available?

5. Are there licensing restrictions?

6. Are there Rate Limit restrictions?

7. Can responses be fully cached?

8. Can a queue be used instead of direct requests?

9. Can thousands of requests be replaced by one aggregated request?

10. Could the integration create legal risks or cause the project to be blocked?

If even one item is not satisfied, the integration is prohibited until a separate architectural review.
