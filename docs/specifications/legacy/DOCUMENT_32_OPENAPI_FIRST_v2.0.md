# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 32

# OPENAPI-FIRST, API CONTRACTS, VERSIONING, SDK GENERATION & CLIENT COMMUNICATION CONSTITUTION

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: API CONSTITUTION

---

# PURPOSE

Настоящий документ определяет единственный источник истины для взаимодействия всех клиентов OpenInvest:

* Web
* iOS
* Android
* AI Assistant
* Future Public API
* Future Partner API

Backend и Frontend запрещено разрабатывать независимо.

Сначала проектируется контракт (OpenAPI), затем автоматически генерируются SDK, после чего начинается реализация.

---

# API PHILOSOPHY

Правильная последовательность:

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

является главным документом взаимодействия.

---

Запрещается:

создавать endpoint без OpenAPI;

изменять response "на лету";

изменять JSON без обновления контракта.

---

# API STYLE

Используется исключительно REST.

---

Все ответы имеют единый формат.

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

Breaking Changes запрещены.

---

# BASE RESPONSE

Каждый endpoint обязан возвращать:

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

Единая структура:

```text
code

message

details

traceId
```

---

Никаких:

```
500

Internal Error
```

без объяснения причины.

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

Используется Cursor Pagination.

---

Запрещено:

```
offset=500000
```

---

Используется:

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

Поддерживается:

Ticker

ISIN

Company Name

---

# SDK GENERATION

После утверждения OpenAPI автоматически генерируются:

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

Backend проходит Contract Tests.

Frontend проходит Contract Tests.

Mobile проходит Contract Tests.

---

Если контракт нарушен,

Build считается неуспешным.

---

# DEPRECATION POLICY

Endpoint нельзя удалить сразу.

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

Минимальный срок поддержки:

12 месяцев.

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

Все GET endpoint обязаны возвращать:

```
ETag

Cache-Control

Last-Modified
```

---

# COMPRESSION

Обязательно:

Brotli

↓

Gzip

↓

Identity

---

# IDEMPOTENCY

POST операции, связанные с финансами,

обязаны поддерживать:

```
Idempotency-Key
```

---

Повторная отправка

не должна создавать дубликаты.

---

# TRACEABILITY

Каждый запрос получает:

```
RequestID

TraceID

UserID(optional)
```

---

# OBSERVABILITY

Каждый endpoint публикует:

Latency

Memory

CPU

Cache Hit

DB Queries

---

# OPENAPI REVIEW CHECKLIST

Перед утверждением нового endpoint необходимо ответить:

---

Можно ли использовать существующий endpoint?

---

Не дублирует ли он функциональность?

---

Можно ли уменьшить payload?

---

Можно ли использовать Snapshot?

---

Можно ли сделать endpoint асинхронным?

---

Можно ли использовать Event?

---

# PUBLIC API FUTURE

Архитектура должна позволять открыть:

```
OpenInvest Developer API
```

без изменения внутренних сервисов.

---

# FINAL API PRINCIPLE

> **OpenAPI является конституцией взаимодействия компонентов OpenInvest.**

> **Frontend, Backend, Mobile, AI Assistant и будущие внешние интеграции должны разрабатываться не относительно реализации, а относительно утвержденного контракта.**

> **Любое изменение API без обновления OpenAPI-документации считается архитектурной ошибкой и блокирует Merge независимо от качества кода.**
