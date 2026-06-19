# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 12

# BACKEND ARCHITECTURE, DATA PIPELINE & INFRASTRUCTURE

Version: 1.0

Status: Core Architecture

Priority: Critical

---

# 1. PURPOSE

Настоящий документ определяет архитектуру серверной части OpenInvest.

Главные цели:

• высокая скорость;

• минимальное потребление памяти;

• минимальное потребление интернет-трафика;

• минимальное количество запросов к официальным API;

• возможность масштабирования до 100 000+ пользователей;

• возможность дальнейшего перехода на миллионы пользователей без переписывания архитектуры.

---

# 2. ОСНОВНЫЕ ПРИНЦИПЫ

Backend является единственным источником истины.

Frontend никогда:

не считает математику;

не работает напрямую с MOEX;

не обращается к ЦБ;

не парсит дивиденды;

не вычисляет XIRR.

Frontend только отображает готовые данные.

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

Основной API.

Отвечает за:

авторизацию;

портфель;

дивиденды;

API;

кэш;

уведомления;

ZIP;

PDF;

XML;

email;

работу с клиентами.

---

## Python Service

Работает независимо.

Отвечает за:

парсинг;

анализ;

инфляцию;

налоги;

сложную математику;

обогащение данных;

AI рекомендации.

---

## PostgreSQL

Источник хранения данных.

---

## Redis (или RAM Cache)

Источник быстрого чтения.

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

быстро отвечает.

Python:

думает.

PostgreSQL:

хранит.

Redis:

ускоряет.

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

Если данных нет:

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

Самая важная часть продукта.

---

100 000 пользователей

НЕ должны создать

100 000 запросов

к MOEX.

---

Работает так:

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

Приоритет:

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

каждые 5 минут

только в торговые часы.

---

dividends

каждые 30 минут.

---

inflation

раз в месяц.

---

currency

раз в день.

---

company info

раз в неделю.

---

# 12. TRADING CALENDAR

Время пользователя никогда не влияет на расчеты.

Все расчеты привязаны к:

UTC

*

торговому календарю MOEX.

---

# 13. SNAPSHOTS

Каждый день создается:

Portfolio Snapshot

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

Дата

Стоимость

Полученные дивиденды

Ожидаемые дивиденды

Cash

Облигации

Акции

Доходность

Инфляционная стоимость

---

# 15. WHY SNAPSHOTS

Без Snapshot:

нужно пересчитывать

5000 транзакций.

---

Со Snapshot:

нужно отдать

365 точек.

---

Экономия:

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

Тяжелые операции:

XML

PDF

ZIP

EMAIL

Push

AI

не выполняются синхронно.

---

Они помещаются в очередь.

---

# 18. LONG TASKS

Создание декларации:

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

Обязательно:

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

разделяются по годам.

---

audit_logs

по месяцам.

---

snapshots

по годам.

---

# 21. API VERSIONING

```
/api/v1/

/api/v2/

/api/v3/
```

никогда не ломать старые клиенты.

---

# 22. PAGINATION

Все списки:

Cursor Pagination.

OFFSET запрещен для больших таблиц.

---

# 23. SEARCH

Поиск:

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

Добавление второго сервера

не должно требовать изменения кода.

---

Stateless API.

---

# 26. SESSION STORAGE

Session хранится:

Redis

или

Database.

---

никогда

в памяти процесса.

---

# 27. FILE STORAGE

PDF

XML

ZIP

создаются временно.

---

После отправки:

удаляются.

---

# 28. EMAIL

SMTP Worker

отдельный сервис.

---

Основной API

не ждет отправки письма.

---

# 29. PUSH NOTIFICATIONS

отдельная очередь.

---

# 30. AI ENGINE

AI никогда не работает внутри API.

---

AI отдельный Worker.

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

ежедневно.

---

Redis

не обязательно.

---

Storage

ежедневно.

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

Никогда не обращаться:

каждый пользователь

→ MOEX.

---

Всегда:

один сервер

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

не должно требовать смены архитектуры.

---

# 38. PRODUCT PHILOSOPHY

OpenInvest не должен быть самым большим.

OpenInvest должен быть:

самым быстрым,

самым прозрачным,

самым экономичным,

самым надежным,

самым дружелюбным к официальным API.

---

# 39. SUCCESS METRIC

При 100 000 пользователей:

• один запрос к MOEX обслуживает всех;

• мобильное приложение открывается менее чем за 1 секунду;

• портфель отображается без пересчета;

• батарея телефона практически не расходуется;

• пользователь никогда не замечает работы фоновой инфраструктуры.

---

# 40. CODEx REQUIREMENT

Codex обязан реализовывать любую новую функцию только после проверки:

1. Не увеличивает ли она количество запросов к официальным API.

2. Не увеличивает ли она потребление памяти.

3. Не увеличивает ли она интернет-трафик клиента.

4. Не нарушает ли она принципы Cache First и API First.

5. Не ухудшает ли она масштабируемость проекта.
