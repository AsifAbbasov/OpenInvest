# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 12

# BACKEND ARCHITECTURE, DATA PIPELINE & INFRASTRUCTURE

Version: 1.0

Status: Core Architecture

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_12_BACKEND_PIPELINE_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `6eaea7764d72a8349065a2fa7aacd9d7f466a024`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# 1. PURPOSE

This document defines the server-side architecture of OpenInvest.

Main goals:

• high speed;

• minimal memory consumption;

• minimal internet traffic consumption;

• minimal number of requests to official APIs;

• ability to scale to 100 000+ users;

• ability to later transition to millions of users without rewriting the architecture.

---

# 2. CORE PRINCIPLES

Backend is the single source of truth.

Frontend never:

calculates mathematics;

works directly with MOEX;

contacts the CBR;

parses dividends;

calculates XIRR.

Frontend only displays ready-made data.

---

# 3. POLYGLOT ARCHITECTURE

```
                 Web React
                      │
                 iOS Swift
                      │
              Android Kotlin
                      │
────────────────────────────────────
                Go API
────────────────────────────────────
        Redis / RAM Cache
────────────────────────────────────
    Python Analytics Service
────────────────────────────────────
          PostgreSQL
────────────────────────────────────
 Official Data Providers (MOEX/CBR)
```

---

# 4. BACKEND COMPONENTS

## Go Fiber

Main API.

Responsible for:

authorization;

portfolio;

dividends;

API;

cache;

notifications;

ZIP;

PDF;

XML;

email;

client interaction.

---

## Python Service

Works independently.

Responsible for:

parsing;

analysis;

inflation;

taxes;

complex mathematics;

data enrichment;

AI recommendations.

---

## PostgreSQL

Data storage source.

---

## Redis (or RAM Cache)

Fast-read source.

---

# 5. PROJECT STRUCTURE

```
backend-go/

cmd/

internal/

config/

database/

middleware/

router/

cache/

scheduler/

services/

portfolio/

catalog/

dividends/

tax/

notifications/

auth/

audit/

storage/

utils/

dto/

models/

repositories/

usecases/

tests/

docs/
```

---

# 6. PYTHON STRUCTURE

```
python-service/

api/

jobs/

parsers/

analytics/

inflation/

tax/

currency/

recommendations/

notifications/

tests/
```

---

# 7. RESPONSIBILITY

Go:

responds quickly.

Python:

thinks.

PostgreSQL:

stores.

Redis:

accelerates.

---

# 8. REQUEST FLOW

```
Client

↓

Go API

↓

Redis

↓

PostgreSQL

↓

Response
```

If data is absent:

```
Client

↓

Go

↓

Python

↓

PostgreSQL

↓

Redis

↓

Client
```

---

# 9. CACHE STRATEGY

The most important part of the product.

---

100 000 users

MUST NOT create

100 000 requests

to MOEX.

---

It works like this:

```
MOEX

↓

Parser

↓

Redis

↓

Go

↓

100000 Users
```

---

# 10. DATA SOURCES

Priority:

1

MOEX ISS

---

2

CBR

---

3

Rosstat

---

4

e-disclosure

---

5

Issuer Websites

---

# 11. DATA REFRESH

marketdata

every 5 minutes

only during trading hours.

---

dividends

every 30 minutes.

---

inflation

once a month.

---

currency

once a day.

---

company info

once a week.

---

# 12. TRADING CALENDAR

User time never affects calculations.

All calculations are tied to:

UTC

*

the MOEX trading calendar.

---

# 13. SNAPSHOTS

Every day, a

Portfolio Snapshot

is created.

```
User

↓

Current Holdings

↓

Close Prices

↓

Total Value

↓

Snapshot
```

---

# 14. SNAPSHOT CONTENT

Date

Value

Dividends received

Expected dividends

Cash

Bonds

Stocks

Return

Inflation-adjusted value

---

# 15. WHY SNAPSHOTS

Without Snapshot:

it is necessary to recalculate

5000 transactions.

---

With Snapshot:

it is necessary to return

365 points.

---

Savings:

CPU

RAM

Traffic

Battery

---

# 16. DATABASE LAYERS

Raw Data

↓

Normalized Data

↓

Calculated Data

↓

Cached Data

↓

Client

---

# 17. QUEUE

Heavy operations:

XML

PDF

ZIP

EMAIL

Push

AI

are not executed synchronously.

---

They are placed in a queue.

---

# 18. LONG TASKS

Declaration creation:

Client

↓

Task

↓

Queue

↓

Worker

↓

Email

↓

Ready

---

# 19. DATABASE INDEXES

Required:

user_id

ticker

trade_date

snapshot_date

isin

country

status

---

# 20. DATABASE PARTITIONING

transactions

are partitioned by year.

---

audit_logs

by month.

---

snapshots

by year.

---

# 21. API VERSIONING

```
/api/v1/

/api/v2/

/api/v3/
```

never break old clients.

---

# 22. PAGINATION

All lists:

Cursor Pagination.

OFFSET is prohibited for large tables.

---

# 23. SEARCH

Search:

Ticker

ISIN

Company

Sector

Currency

---

Full Text Index.

---

# 24. PERFORMANCE TARGETS

Catalog

<100 ms

---

Portfolio

<150 ms

---

Dividend Calendar

<120 ms

---

Tax Export

<3 sec

---

Search

<50 ms

---

# 25. HORIZONTAL SCALING

Adding a second server

must not require code changes.

---

Stateless API.

---

# 26. SESSION STORAGE

Session is stored in:

Redis

or

Database.

---

never

in process memory.

---

# 27. FILE STORAGE

PDF

XML

ZIP

are created temporarily.

---

After sending:

they are deleted.

---

# 28. EMAIL

SMTP Worker

is a separate service.

---

The main API

does not wait for email sending.

---

# 29. PUSH NOTIFICATIONS

separate queue.

---

# 30. AI ENGINE

AI never runs inside the API.

---

AI is a separate Worker.

---

# 31. OBSERVABILITY

Metrics

Logs

Tracing

Health

Readiness

Liveness

---

# 32. HEALTH CHECK

```
/health

/ready

/live
```

---

# 33. BACKUPS

Database

daily.

---

Redis

not required.

---

Storage

daily.

---

# 34. DISASTER RECOVERY

Server Lost

↓

Restore DB

↓

Restore Storage

↓

Warm Cache

↓

Resume

---

# 35. FREE API STRATEGY

Never:

each user

→ MOEX.

---

Always:

one server

→ MOEX.

---

# 36. ANTI BLOCK STRATEGY

Adaptive Refresh

Exponential Backoff

Retry

Circuit Breaker

Cache First

Rate Control

---

# 37. FUTURE SCALING

100

↓

1000

↓

10000

↓

100000

↓

1000000

must not require an architecture change.

---

# 38. PRODUCT PHILOSOPHY

OpenInvest should not be the largest.

OpenInvest should be:

the fastest,

the most transparent,

the most economical,

the most reliable,

the most friendly to official APIs.

---

# 39. SUCCESS METRIC

With 100 000 users:

• one request to MOEX serves everyone;

• the mobile application opens in less than 1 second;

• the portfolio is displayed without recalculation;

• the phone battery is barely consumed;

• the user never notices the background infrastructure working.

---

# 40. CODEX REQUIREMENT

Codex must implement any new function only after checking:

1. Does it increase the number of requests to official APIs?

2. Does it increase memory consumption?

3. Does it increase client internet traffic?

4. Does it violate Cache First and API First principles?

5. Does it reduce project scalability?
