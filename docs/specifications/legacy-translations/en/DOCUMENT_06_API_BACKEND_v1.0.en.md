Excellent.

Now we move on to the most important document of the project.

---

# DOCUMENT 06

# API CONTRACT & BACKEND ARCHITECTURE

**Version:** 1.0

**Status:** Source of Truth

**Priority:** Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_06_API_BACKEND_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `d82c487382ba542ee493c3a9373c744f08937d51`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# 1. BACKEND PHILOSOPHY

Backend is the **single source of truth**.

Frontend (React/Web, Swift/iOS, Kotlin/Android) contains no business logic.

Any client is simply a representation of data.

```
Client

↓

API Gateway

↓

Go Fiber

↓

Service Layer

↓

Domain Layer

↓

Repository Layer

↓

PostgreSQL / Cache

↓

Official Data Providers
```

---

# 2. CORE PRINCIPLES

The API must be

Stateless

Idempotent

Versioned

Observable

Cacheable

Scalable

Auditable

---

# 3. API STRUCTURE

```
/api

/api/v1

/api/v2
```

v1 is never broken.

New capabilities appear only in a new version.

---

# 4. API MODULES

```
Auth

Users

Portfolio

Transactions

Assets

Prices

Dividends

Coupons

Statistics

Tax

Notifications

Search

Admin

Health

Metrics
```

---

# 5. AUTH

```
POST /auth/register

POST /auth/login

POST /auth/logout

POST /auth/refresh

POST /auth/change-password

POST /auth/reset-password

GET /auth/me
```

---

# 6. USERS

```
GET /users/profile

PUT /users/profile

DELETE /users/profile

GET /users/settings

PUT /users/settings
```

---

# 7. PORTFOLIO

```
GET /portfolio

GET /portfolio/summary

GET /portfolio/history

GET /portfolio/allocation

GET /portfolio/dividends

GET /portfolio/inflation

GET /portfolio/xirr
```

---

# 8. TRANSACTIONS

```
GET /transactions

POST /transactions

PUT /transactions/{id}

DELETE /transactions/{id}

GET /transactions/history
```

---

# 9. ASSETS

```
GET /assets

GET /assets/search

GET /assets/{ticker}

GET /assets/{ticker}/history

GET /assets/{ticker}/dividends

GET /assets/{ticker}/coupons

GET /assets/{ticker}/statistics
```

---

# 10. DIVIDENDS

```
GET /dividends/calendar

GET /dividends/upcoming

GET /dividends/history

GET /dividends/official

GET /dividends/prognosis
```

---

# 11. TAX

```
GET /tax/profile

PUT /tax/profile

POST /tax/generate/xml

POST /tax/generate/pdf

POST /tax/send/email

POST /tax/export/zip
```

---

# 12. NOTIFICATIONS

```
GET /notifications

PUT /notifications/read

PUT /notifications/settings

DELETE /notifications/{id}
```

---

# 13. SEARCH

```
GET /search

GET /search/ticker

GET /search/company
```

---

# 14. HEALTH

```
GET /health

GET /metrics

GET /version

GET /status
```

---

# 15. UNIFIED RESPONSE FORMAT

```
{
    success,
    data,
    errors,
    meta
}
```

---

# 16. PAGINATION

Cursor Pagination.

Not Offset.

Reason:

we need hundreds of thousands of records.

Offset will degrade.

---

# 17. SORTING

Any list must support

```
sort

order

limit

cursor
```

---

# 18. FILTERS

Catalog

```
sector

price

yield

marketCap

currency

assetType
```

---

# 19. CACHE STRATEGY

Not every request goes to MOEX.

```
Client

↓

CDN

↓

API Cache

↓

RAM Cache

↓

Database

↓

Official API
```

---

# 20. RAM CACHE

Go

```
sync.Map
```

or

Redis

---

# 21. CACHE TTL

market data

5 minutes

---

dividends

12 hours

---

statistics

24 hours

---

inflation

1 month

---

currency

24 hours

---

# 22. RATE LIMIT

A very important section.

---

Never

never

never

should Frontend contact MOEX directly.

---

All requests:

```
React

↓

Go

↓

Cache

↓

MOEX
```

---

# 23. PROTECTION FROM BLOCKING

If

10000 users

open the application simultaneously,

they DO NOT create

10000 requests.

---

They receive

ONE

already cached response.

---

# 24. UPDATE STRATEGY

During trading

every 5 minutes

---

After market close

one final refresh

---

At night

no requests.

---

Weekends

no requests.

---

Holidays

no requests.

---

# 25. OFFICIAL SOURCES

Only

official

or

freely available.

---

MOEX ISS

CBR

Rosstat

official issuer websites

e-disclosure

---

# 26. NO SCRAPING

If an official API exists,

only it is used.

---

A parser is used

only

if there is no official API.

---

# 27. EMAIL

Asynchronous queue.

The user does not wait for SMTP.

```
Request

↓

Queue

↓

Worker

↓

SMTP

↓

Success

↓

Notification
```

---

# 28. ZIP

XML

PDF

CSV

are created

exclusively

in memory.

```
bytes.Buffer

↓

archive/zip

↓

SMTP
```

---

no temporary files.

---

# 29. INFLATION

API

```
GET /portfolio/inflation
```

returns

```
nominalValue

realValue

inflationPercent

purchasingPower

examples
```

---

# 30. PURCHASING POWER

The user sees

not

```
200000 rubles
```

but

```
200000 rubles

=

1 MacBook Pro

or

5 iPhone

or

8 months of utility payments

or

4 average regional salaries
```

---

This is one of our unique functions.

---

# 31. XIRR

Never

calculated

in Frontend.

---

Always

Backend.

---

# 32. SNAPSHOTS

Frontend receives

a ready-made array.

```
date

value

cash

stocks

bonds

realValue
```

no recalculations.

---

# 33. IDEMPOTENCY

Every financial operation

has an

Idempotency-Key.

Repeated submission

will not create a duplicate transaction.

---

# 34. OBSERVABILITY

Every endpoint

has

Latency

Errors

Success Rate

Memory

CPU

Cache Hit

---

# 35. SECURITY

JWT

Refresh Token

Rotation

CSRF

Rate Limit

Brute Force Protection

Device Tracking

IP Monitoring

---

# 36. TRUST & PRIVACY

By default

passport

INN

address

phone

are NOT required.

---

A tax profile

is created

optionally.

---

# 37. DATA DELETION

The user can

download

XML

CSV

JSON

PDF

and then

delete

the entire profile

with one button.

---

# 38. WHAT NEEDS TO BE ADDED (MY CRITIQUE)

Here is what, in my view, must appear in the next documents:

## 1. Import Center

Import:

* BCS
* Alfa
* T-Investments
* VTB
* Finam
* Sber
* Interactive Brokers
* CSV
* Excel

With one button.

---

## 2. Compare Portfolio

Not only:

"how much did I earn"

but

"what would have happened if I had held Lukoil instead of Gazprom"

---

## 3. Dividend Simulator

"If I buy another 100 shares today"

↓

show:

* dividends
* XIRR
* tax
* inflation-adjusted value
* change in purchasing power

---

## 4. Life Goals

The strongest product idea.

Not simply

```
Portfolio = 5 700 000 ₽
```

but

```
Your portfolio provides:

✓ 14 years of utility payments

✓ 8 years of food basket expenses

✓ 3 MacBook Pro each year from dividends alone

✓ 27% of the way to financial independence
```

---

### My assessment

The document is already strong enough to begin Backend development.

But, in my opinion, **the most important documents are still ahead**:

1. **DOCUMENT 07 — Full Frontend Architecture (React + RTK + Feature Folder + UI/UX)**
2. **DOCUMENT 08 — Mobile First Architecture (iOS + Android + API First)**
3. **DOCUMENT 09 — Security & Privacy by Design**
4. **DOCUMENT 10 — Product Vision, UX, competitive advantages and usage scenarios**

These documents are what will turn the project from a "dividend calculator" into a full-fledged world-class investment platform.
