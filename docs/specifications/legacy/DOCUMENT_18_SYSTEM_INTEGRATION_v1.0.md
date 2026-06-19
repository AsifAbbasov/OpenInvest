# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 18

# SYSTEM ARCHITECTURE, EXTERNAL API INTEGRATION, RATE LIMITS, SCALABILITY & RESILIENCE

Version: 1.0

Status: Approved

Priority: Critical

---

# PURPOSE

Данный документ описывает:

* архитектуру взаимодействия всех сервисов;
* работу с бесплатными официальными API;
* защиту от блокировок;
* масштабирование до сотен тысяч пользователей;
* работу очередей;
* кэширование;
* стратегию отказоустойчивости.

Это один из самых важных документов проекта.

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

Любой клиент:

React

iOS

Android

будет работать через один API.

Никаких отдельных backend.

---

# SINGLE BACKEND

```text
One Backend

Many Clients
```

---

# SOURCE OF MARKET DATA

Используются только:

официальные

или

бесплатные

источники.

---

# ALLOWED SOURCES

MOEX ISS

ЦБ РФ

Росстат

e-disclosure

официальные сайты эмитентов

---

# FORBIDDEN SOURCES

Парсинг личных кабинетов

Платные API без лицензии

Обход защиты сайтов

Нарушение robots.txt

---

# REQUEST STRATEGY

Самая большая ошибка подобных проектов —

каждый пользователь делает собственный запрос.

---

Мы так делать НЕ будем.

---

# BAD SCENARIO

10000 пользователей

↓

10000 запросов

↓

MOEX

↓

IP Block

↓

Проект умер

---

# OUR SCENARIO

10000 пользователей

↓

Go Backend

↓

1 запрос

↓

MOEX

↓

Redis

↓

10000 пользователей получают один ответ

---

# REQUEST AGGREGATION

Все одинаковые запросы объединяются.

---

Например

1000 пользователей открыли Сбер.

Backend:

делает

ОДИН

запрос.

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

5 минут

---

Dividends

24 часа

---

Inflation

30 дней

---

Currency

24 часа

---

Company Profile

30 дней

---

# STALE CACHE

Если API недоступен

используется последний успешный ответ.

---

Пользователь видит

```text
Последнее обновление:

15:05 UTC
```

---

# FAILOVER

Если источник не отвечает

Backend

не падает.

---

Используется:

Cached Version

---

# CIRCUIT BREAKER

После

5

неудачных запросов

Backend перестает обращаться к API

на

5 минут.

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

Внутренний Backend

не должен превышать

допустимые лимиты MOEX.

---

# REQUEST DISTRIBUTION

Все фоновые обновления

происходят

через очередь.

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

Все сервисы общаются

через события.

---

Пример:

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

никогда

не обращается

к MOEX.

---

# NEVER

Mobile

никогда

не обращается

к внешнему API.

---

# SCALABILITY

Stage 1

1000 пользователей

один сервер

---

Stage 2

10000 пользователей

Go + Redis

---

Stage 3

100000 пользователей

Load Balancer

↓

Go x4

↓

Redis

↓

PostgreSQL

---

Stage 4

1000000 пользователей

Microservices

↓

Queue

↓

Read Replica

↓

CDN

---

# DATABASE LOAD

Backend никогда не должен строить графики

из миллионов транзакций.

---

Используются:

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

Никогда не ломать старую версию.

---

# PAGINATION

Все списки

обязаны иметь:

limit

offset

cursor

---

# SORTING

Server Side.

---

Frontend

не сортирует

10000 записей.

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

создаются

асинхронно.

---

Пользователь получает:

```
Документ готов.

Скачать

или

Отправить на Email.
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

Каждый внешний запрос:

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

Каждый сервис обязан иметь:

```
/health

/ready

/live
```

---

# SECURITY

Все внешние API

работают

только через Backend.

---

API Keys

никогда

не попадают

во Frontend.

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

Основная цель проекта:

обслуживать

10000+

пользователей

на минимальных расходах.

---

# TARGET COST

Frontend

≈ бесплатно (Vercel)

---

Backend

≈ минимальный тариф

---

Redis

минимальный

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

используют

тот же API.

---

Никакой новой логики.

---

Никаких новых вычислений.

---

Только новый UI.

---

# COMPETITIVE ADVANTAGES

По сравнению с большинством брокеров:

✅ Backend кэширует данные.

✅ Минимальный расход мобильного трафика.

✅ Минимальный расход батареи.

✅ Нет постоянных запросов.

✅ Нет тяжелых вычислений на телефоне.

✅ Нет прямого обращения к бирже.

✅ Один запрос обслуживает тысячи пользователей.

✅ Полностью масштабируемая API-First архитектура.

---

# CODEX REQUIREMENTS

Перед интеграцией любого нового внешнего API Builder Agent обязан проверить:

1. Источник официальный?

2. Источник бесплатный?

3. Разрешает массовое использование?

4. Есть ли документация?

5. Есть ли ограничения по лицензии?

6. Есть ли ограничения по Rate Limit?

7. Можно ли полностью кэшировать ответы?

8. Можно ли использовать очередь вместо прямых запросов?

9. Можно ли заменить тысячи запросов одним агрегированным?

10. Не приведет ли интеграция к юридическим рискам или блокировке проекта?

Если хотя бы один пункт не выполняется — интеграция запрещается до отдельного архитектурного ревью.
