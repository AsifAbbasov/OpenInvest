# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 15

# BACKEND ARCHITECTURE, API, CACHE, SERVER INFRASTRUCTURE & SCALABILITY

Version: 1.0

Status: Approved

Priority: Critical

---

# PURPOSE

Настоящий документ определяет единственную допустимую архитектуру Backend OpenInvest.

Backend является центром всей системы.

Frontend никогда не выполняет тяжелые вычисления.

Мобильные приложения никогда не обращаются к внешним API.

Все вычисления, агрегация, кэширование и синхронизация происходят исключительно на серверной стороне.

---

# CORE PHILOSOPHY

Backend должен быть:

* быстрым;
* предсказуемым;
* энергоэффективным;
* масштабируемым;
* дешевым в эксплуатации;
* независимым от клиента;
* отказоустойчивым.

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

или NATS

---

## Monitoring

Prometheus

Grafana

OpenTelemetry

---

# BACKEND RESPONSIBILITIES

Backend отвечает за:

Авторизацию

Портфель

Каталог

Дивиденды

Облигации

Календарь

Налоги

XML

PDF

Email

Push

Snapshots

XIRR

Инфляцию

AI рекомендации

Audit

Privacy

---

# API FIRST

Backend не знает,

кто обращается.

```text
React

↓

Go API

↑

Swift

↑

Kotlin
```

Все клиенты используют одинаковый API.

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

Все ответы имеют единую структуру.

```text
BaseResponse<T>

resultCode

messages

data
```

---

# VERSIONING

Все API обязаны использовать версионирование.

```text
/api/v1/

/api/v2/
```

---

# PAGINATION

Обязательно:

page

limit

total

next

previous

---

# FILTERING

Все фильтры выполняются Backend.

Frontend никогда не фильтрует большие массивы.

---

# SORTING

Backend выполняет:

по цене

по капитализации

по DY

по сектору

по ликвидности

---

# SEARCH

Использовать:

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

Использовать:

gzip

brotli

---

Все JSON ответы сжимаются.

---

# ETag

Для неизменяемых данных использовать ETag.

Если данные не изменились

возвращать

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

Используется для:

котировок

каталога

справочников

курсов

дивидендов

---

# REDIS CACHE

Используется для:

Sessions

RateLimit

Snapshots

Temporary Export

Notification Queue

---

# CACHE TTL

Каталог

5 минут

---

Цена акции

1 минута во время торгов

10 минут вне торгов

---

Дивиденды

24 часа

---

Инфляция

30 дней

---

Сектора

30 дней

---

# EXTERNAL API POLICY

Frontend никогда не обращается:

MOEX

CBR

Rosstat

e-disclosure

SmartLab

---

Backend агрегирует данные.

---

# REQUEST AGGREGATION

Никогда

10000 пользователей

↓

10000 запросов

↓

MOEX

---

Всегда

10000 пользователей

↓

1 Backend Request

↓

Cache

↓

10000 ответов

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

Все обновления ориентируются

на торговый календарь MOEX.

Не на пользователя.

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

Frontend отображает локальное время пользователя.

---

# SNAPSHOT STRATEGY

После завершения основной торговой сессии:

получить close price

↓

пересчитать позиции

↓

пересчитать XIRR

↓

пересчитать доходность

↓

создать snapshot

↓

обновить Dashboard

---

# PORTFOLIO HISTORY

Frontend никогда не получает

историю транзакций

для построения графика.

Используются snapshots.

---

# BACKGROUND WORKERS

Go

создание snapshots

уведомления

zip

email

---

Python

парсинг

инфляция

дивиденды

налоги

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

Создается

исключительно

в RAM.

После отправки удаляется.

---

# PDF

Не хранится постоянно.

---

# XML

Не хранится постоянно.

---

# MONITORING

Собирать:

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

Все ошибки структурированы.

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

Любой сервер способен обработать любой запрос.

---

# DATABASE CONNECTIONS

Использовать Pool.

Никогда не создавать соединение на каждый запрос.

---

# COST OPTIMIZATION

Приоритет:

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

Если MOEX недоступна:

использовать RAM Cache

↓

Redis

↓

Database

↓

последние подтвержденные данные

↓

отобразить время обновления

---

Пользователь никогда не должен видеть сообщение:

"Ошибка получения котировок".

---

# FUTURE SCALING

Архитектура должна без изменений поддерживать:

100 000 пользователей

1 000 000 транзакций

50 000 портфелей

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

Перед написанием каждого Backend-модуля необходимо проверить:

1. Можно ли уменьшить количество внешних запросов?

2. Можно ли использовать Cache?

3. Можно ли использовать Snapshot вместо тяжелого расчета?

4. Можно ли уменьшить использование RAM?

5. Можно ли уменьшить сетевой трафик?

6. Не приведет ли реализация к блокировке бесплатных официальных API?

7. Соответствует ли код SOLID, KISS, DRY, YAGNI, SRP, DIP, ISP, LSP, Open/Closed и Privacy by Design?

Только после прохождения этих проверок модуль считается готовым к Review Agent и QA Agent.
