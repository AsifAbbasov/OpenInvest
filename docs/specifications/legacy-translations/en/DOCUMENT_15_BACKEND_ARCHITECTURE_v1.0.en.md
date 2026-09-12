# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 15

# BACKEND ARCHITECTURE, API, CACHE, SERVER INFRASTRUCTURE & SCALABILITY

Version: 1.0

Status: Approved

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_15_BACKEND_ARCHITECTURE_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `c95c461f9c1317e105e5abf6765e70d055320f47`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# PURPOSE

This document defines the only permitted Backend architecture for OpenInvest.

Backend is the center of the entire system.

Frontend never performs heavy calculations.

Mobile applications never contact external APIs.

All calculations, aggregation, caching, and synchronization happen exclusively on the server side.

---

# CORE PHILOSOPHY

Backend must be:

* fast;
* predictable;
* energy efficient;
* scalable;
* inexpensive to operate;
* client-independent;
* fault tolerant.

---

# GLOBAL ARCHITECTURE

```text
Web React

iOS Swift

Android Kotlin

↓

Load Balancer

↓

Go Fiber API

↓

Redis

↓

PostgreSQL

↓

Python Workers

↓

Official Free APIs
```

---

# TECHNOLOGY STACK

## Main API

Go 1.24+

Fiber

---

## Cache

Redis

*

RAM Cache

(sync.Map)

---

## Database

PostgreSQL

---

## Background Jobs

Go Workers

Python Workers

---

## Queue

Redis Streams

or NATS

---

## Monitoring

Prometheus

Grafana

OpenTelemetry

---

# BACKEND RESPONSIBILITIES

Backend is responsible for:

Authorization

Portfolio

Catalog

Dividends

Bonds

Calendar

Taxes

XML

PDF

Email

Push

Snapshots

XIRR

Inflation

AI recommendations

Audit

Privacy

---

# API FIRST

Backend does not know

who is calling it.

```text
React

↓

Go API

↑

Swift

↑

Kotlin
```

All clients use the same API.

---

# SERVER DIRECTORY

```text
backend-go/

cmd/

internal/

api/

auth/

portfolio/

catalog/

calendar/

tax/

notification/

worker/

scheduler/

cache/

database/

middleware/

metrics/

audit/

security/

config/

tests/
```

---

# API RULES

All responses have a unified structure.

```text
BaseResponse<T>

resultCode

messages

data
```

---

# VERSIONING

All APIs must use versioning.

```text
/api/v1/

/api/v2/
```

---

# PAGINATION

Required:

page

limit

total

next

previous

---

# FILTERING

All filtering is performed by Backend.

Frontend never filters large arrays.

---

# SORTING

Backend performs:

by price

by capitalization

by DY

by sector

by liquidity

---

# SEARCH

Use:

GIN

pg_trgm

Full Text Search

---

# REQUEST FLOW

```text
Client

↓

Fiber

↓

Middleware

↓

Auth

↓

Cache

↓

Business Logic

↓

Database

↓

Response
```

---

# MIDDLEWARE

RequestID

Recovery

Logger

JWT

RateLimit

Compression

CORS

ETag

CacheControl

---

# COMPRESSION

Use:

gzip

brotli

---

All JSON responses are compressed.

---

# ETag

Use ETag for immutable data.

If data has not changed,

return

304 Not Modified.

---

# CACHE STRATEGY

Level 1

RAM

---

Level 2

Redis

---

Level 3

PostgreSQL

---

# RAM CACHE

Used for:

quotes

catalog

directories

exchange rates

dividends

---

# REDIS CACHE

Used for:

Sessions

RateLimit

Snapshots

Temporary Export

Notification Queue

---

# CACHE TTL

Catalog

5 minutes

---

Stock price

1 minute during trading

10 minutes outside trading

---

Dividends

24 hours

---

Inflation

30 days

---

Sectors

30 days

---

# EXTERNAL API POLICY

Frontend never contacts:

MOEX

CBR

Rosstat

e-disclosure

SmartLab

---

Backend aggregates data.

---

# REQUEST AGGREGATION

Never

10000 users

↓

10000 requests

↓

MOEX

---

Always

10000 users

↓

1 Backend Request

↓

Cache

↓

10000 responses

---

# RATE LIMIT

Anonymous

60 req/min

---

Authorized

300 req/min

---

Premium

1000 req/min

---

# TRADING SCHEDULE

All updates follow

the MOEX trading calendar.

Not the user.

---

# TIME STANDARD

Backend

UTC

---

Database

UTC

---

Workers

UTC

---

Frontend displays the user's local time.

---

# SNAPSHOT STRATEGY

After the main trading session ends:

get close price

↓

recalculate positions

↓

recalculate XIRR

↓

recalculate return

↓

create snapshot

↓

update Dashboard

---

# PORTFOLIO HISTORY

Frontend never receives

transaction history

to build a chart.

Snapshots are used.

---

# BACKGROUND WORKERS

Go

snapshot creation

notifications

zip

email

---

Python

parsing

inflation

dividends

taxes

XML

---

# EMAIL PIPELINE

```text
Request

↓

Queue

↓

ZIP Generation

↓

SMTP

↓

Retry

↓

Success
```

---

# ZIP GENERATION

Created

exclusively

in RAM.

Deleted after sending.

---

# PDF

Not stored permanently.

---

# XML

Not stored permanently.

---

# MONITORING

Collect:

Latency

Memory

CPU

Cache Hit

Cache Miss

DB Time

Worker Time

SMTP Time

---

# HEALTH ENDPOINTS

```text
/health

/live

/ready

/metrics
```

---

# LOGGING

All errors are structured.

JSON Format.

---

# SECURITY

JWT

Refresh Rotation

Argon2id

HTTPS Only

CSRF Protection

RateLimit

IP Analysis

Device Fingerprint

---

# HORIZONTAL SCALING

API Stateless.

Any server can process any request.

---

# DATABASE CONNECTIONS

Use Pool.

Never create a connection for each request.

---

# COST OPTIMIZATION

Priority:

RAM

↓

Redis

↓

Database

↓

Official API

---

# TARGET PERFORMANCE

Average Response

<80 ms

---

95 percentile

<150 ms

---

99 percentile

<300 ms

---

# FAILURE STRATEGY

If MOEX is unavailable:

use RAM Cache

↓

Redis

↓

Database

↓

latest confirmed data

↓

display update time

---

The user should never see the message:

"Error retrieving quotes."

---

# FUTURE SCALING

The architecture must support without changes:

100 000 users

1 000 000 transactions

50 000 portfolios

Web

iOS

Android

Desktop

Public API

AI Assistant

Multi Broker

Multi Country

Multi Currency

---

# CODEX REQUIREMENTS

Before writing each Backend module, check:

1. Can the number of external requests be reduced?

2. Can Cache be used?

3. Can Snapshot be used instead of a heavy calculation?

4. Can RAM usage be reduced?

5. Can network traffic be reduced?

6. Will the implementation cause free official APIs to block us?

7. Does the code comply with SOLID, KISS, DRY, YAGNI, SRP, DIP, ISP, LSP, Open/Closed, and Privacy by Design?

Only after passing these checks is the module considered ready for Review Agent and QA Agent.
