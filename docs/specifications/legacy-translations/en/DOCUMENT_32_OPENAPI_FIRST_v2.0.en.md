# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 32

# OPENAPI-FIRST, API CONTRACTS, VERSIONING, SDK GENERATION & CLIENT COMMUNICATION CONSTITUTION

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: API CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_32_OPENAPI_FIRST_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `eefa537a4a737bcd35b52a6a88664ef14018d0ac`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the single source of truth for interaction among all OpenInvest clients:

* Web
* iOS
* Android
* AI Assistant
* Future Public API
* Future Partner API

Backend and Frontend are prohibited from being developed independently.

First the contract (OpenAPI) is designed, then SDKs are generated automatically, after which implementation begins.

---

# API PHILOSOPHY

The correct sequence:

```text
Business Requirements
        ↓
Domain Model
        ↓
OpenAPI Contract
        ↓
Architecture Review
        ↓
SDK Generation
        ↓
Backend
        ↓
Frontend
        ↓
Mobile
```

---

# SINGLE SOURCE OF TRUTH

```text
openapi.yaml
```

is the primary interaction document.

---

It is prohibited to:

create an endpoint without OpenAPI;

change a response "on the fly";

change JSON without updating the contract.

---

# API STYLE

Only REST is used.

---

All responses have a single format.

```json
{
  "success": true,
  "data": {},
  "meta": {},
  "errors": []
}
```

---

# VERSIONING

```
/api/v1/
/api/v2/
/api/v3/
```

---

Breaking Changes are prohibited.

---

# BASE RESPONSE

Every endpoint must return:

```text
success

data

meta

pagination(optional)

traceId

requestId
```

---

# ERROR MODEL

Unified structure:

```text
code

message

details

traceId
```

---

No:

```
500

Internal Error
```

without explaining the cause.

---

# AUTHORIZATION

Public

↓

Authenticated

↓

Premium

↓

Admin

↓

Internal

---

# API GROUPS

## AUTH

```
POST /auth/register

POST /auth/login

POST /auth/logout

POST /auth/refresh

POST /auth/reset-password
```

---

## USER

```
GET /user

PUT /user

DELETE /user

GET /user/export
```

---

## PORTFOLIO

```
GET /portfolio

POST /portfolio

PUT /portfolio

DELETE /portfolio
```

---

## TRANSACTIONS

```
GET /transactions

POST /transactions

PUT /transactions/{id}

DELETE /transactions/{id}
```

---

## ASSETS

```
GET /assets

GET /assets/{ticker}

GET /assets/search
```

---

## DIVIDENDS

```
GET /dividends

GET /dividends/calendar

GET /dividends/history
```

---

## ANALYTICS

```
GET /analytics/xirr

GET /analytics/cagr

GET /analytics/sharpe

GET /analytics/sortino

GET /analytics/real-return
```

---

## INFLATION

```
GET /inflation

GET /inflation/purchasing-power
```

---

## TAX

```
POST /tax/xml

POST /tax/pdf

POST /tax/email

GET /tax/history
```

---

## NOTIFICATIONS

```
GET /notifications

PUT /notifications/read

DELETE /notifications/{id}
```

---

# PAGINATION

Cursor Pagination is used.

---

Prohibited:

```
offset=500000
```

---

The following are used:

```
cursor

limit
```

---

# SORTING

```
sort=name

sort=yield

sort=price

sort=marketCap
```

---

# FILTERING

```
sector=

country=

assetType=

status=
```

---

# SEARCH

Supported:

Ticker

ISIN

Company Name

---

# SDK GENERATION

After OpenAPI is approved, the following are generated automatically:

---

TypeScript SDK

---

Swift SDK

---

Kotlin SDK

---

Python SDK

---

# API CONTRACT TESTING

Backend passes Contract Tests.

Frontend passes Contract Tests.

Mobile passes Contract Tests.

---

If the contract is violated,

the Build is considered unsuccessful.

---

# DEPRECATION POLICY

An endpoint cannot be removed immediately.

---

Lifecycle:

```
Stable

↓

Deprecated

↓

Migration

↓

Replacement

↓

Removal
```

---

Minimum support period:

12 months.

---

# RATE LIMIT

Anonymous

30/min

---

Authorized

100/min

---

Premium

300/min

---

Partner API

Separate Limits

---

# CACHE HEADERS

All GET endpoints must return:

```
ETag

Cache-Control

Last-Modified
```

---

# COMPRESSION

Mandatory:

Brotli

↓

Gzip

↓

Identity

---

# IDEMPOTENCY

POST operations related to finances

must support:

```
Idempotency-Key
```

---

Resubmission

must not create duplicates.

---

# TRACEABILITY

Every request receives:

```
RequestID

TraceID

UserID(optional)
```

---

# OBSERVABILITY

Every endpoint publishes:

Latency

Memory

CPU

Cache Hit

DB Queries

---

# OPENAPI REVIEW CHECKLIST

Before approving a new endpoint, it is necessary to answer:

---

Can an existing endpoint be used?

---

Does it duplicate functionality?

---

Can the payload be reduced?

---

Can Snapshot be used?

---

Can the endpoint be made asynchronous?

---

Can Event be used?

---

# PUBLIC API FUTURE

The architecture must allow opening:

```
OpenInvest Developer API
```

without changing internal services.

---

# FINAL API PRINCIPLE

> **OpenAPI is the constitution governing interaction among OpenInvest components.**

> **Frontend, Backend, Mobile, AI Assistant, and future external integrations must be developed against the approved contract, not against the implementation.**

> **Any API change without updating the OpenAPI documentation is considered an architectural error and blocks Merge regardless of code quality.**
