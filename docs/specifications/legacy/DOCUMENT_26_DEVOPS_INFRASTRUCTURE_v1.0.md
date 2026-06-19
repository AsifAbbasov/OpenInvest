# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 26

# DEVOPS, INFRASTRUCTURE, SERVER ARCHITECTURE, FREE API STRATEGY, RATE LIMITING, DISASTER RECOVERY & PRODUCTION OPERATIONS

Version: 1.0

Status: APPROVED

Priority: CRITICAL

Classification: PRODUCTION FOUNDATION

---

# PURPOSE

Настоящий документ определяет полную серверную архитектуру OpenInvest.

Главная цель:

**создать продукт, способный обслуживать от 10 до 100 000+ пользователей без блокировки официальных API, без резкого роста стоимости инфраструктуры и без деградации производительности.**

---

# GLOBAL PHILOSOPHY

OpenInvest никогда не работает по схеме:

```
10000 пользователей

↓

10000 запросов

↓

MOEX API
```

Это гарантированный путь к блокировке.

---

# CORRECT ARCHITECTURE

```
                    Users

                       │

                       ▼

                Go API Gateway

                       │

                       ▼

              Internal Cache Layer

                (Redis + RAM)

                       │

          ┌────────────┴────────────┐

          ▼                         ▼

 Historical Storage          Background Workers

          ▼                         ▼

                Official Free APIs

                  (Controlled)
```

---

# SINGLE SOURCE OF TRUTH

Пользователь никогда

не обращается напрямую

к MOEX,

ЦБ,

Росстату

или другим API.

---

Все запросы проходят только через:

```
Go Backend
```

---

# OFFICIAL DATA SOURCES

Используются исключительно:

---

MOEX ISS

---

Банк России

---

Росстат

---

официальные раскрытия эмитентов

---

официальные XSD ФНС

---

# PROHIBITED SOURCES

Запрещается строить критический функционал на:

---

неофициальных парсерах;

---

закрытых API;

---

непубличных endpoints;

---

reverse engineering.

---

# REQUEST STRATEGY

## Wrong

```
10000 users

↓

Open portfolio

↓

10000 requests

↓

MOEX
```

---

## Correct

```
MOEX

↓

Worker

↓

Redis

↓

Go API

↓

10000 users
```

---

# BACKGROUND WORKERS

Python Worker отвечает за:

---

дивиденды;

---

инфляцию;

---

курсы валют;

---

официальные раскрытия;

---

налоговые структуры.

---

Go Worker отвечает за:

---

snapshots;

---

кэш;

---

агрегацию;

---

очереди.

---

# UPDATE STRATEGY

## Market Data

Каждые 5 минут

только в часы торгов.

---

## Dividend Directory

Каждые 30 минут.

---

## Inflation

1 раз в сутки.

---

## FX

1 раз в сутки после публикации ЦБ.

---

## Company Directory

1 раз в неделю.

---

# CACHE HIERARCHY

```
RAM

↓

Redis

↓

PostgreSQL

↓

Official APIs
```

---

# CACHE TTL

Market

5 min

---

Asset Card

30 min

---

Dividend Directory

12 hours

---

Inflation

24 hours

---

FX

24 hours

---

# REDIS POLICY

Redis используется только для:

---

котировок;

---

агрегированных данных;

---

сессий;

---

feature flags.

---

Запрещено хранить:

паспорт;

ИНН;

адрес.

---

# DATABASE POLICY

PostgreSQL —

единственный источник пользовательских данных.

---

# DATABASE PARTITIONING

Partition By:

```
user_id

date
```

---

# INDEX STRATEGY

Composite Index:

```
(user_id, trade_date)

(ticker, trade_date)

(snapshot_date)

(type, ticker)
```

---

# SNAPSHOTS

Snapshot —

ключевая архитектурная идея проекта.

---

Телефон

не пересчитывает

5000 операций.

---

Телефон получает:

```
Date

↓

Portfolio Value

↓

Render
```

---

# TIME ZONES

Все серверы работают

исключительно в UTC.

---

# MARKET CALENDAR

Источник истины —

торговый календарь MOEX.

---

НЕ:

локальное время пользователя.

---

# GLOBAL USERS

Пользователь может жить:

---

в США;

---

в Австралии;

---

в Японии;

---

в Германии.

---

Он всегда получает

последний подтвержденный Snapshot.

---

# API GATEWAY

Все клиенты используют:

```
/api/v1/*
```

---

Запрещены прямые подключения

к внутренним сервисам.

---

# API LIMITS

Client

100 requests/min

---

Anonymous

30 requests/min

---

Premium

300 requests/min

---

# INTERNAL RETRY

Использовать:

Exponential Backoff

---

Retry:

1

↓

2

↓

4

↓

8 sec

---

# CIRCUIT BREAKER

Если официальный API недоступен,

Go обязан вернуть:

последний подтвержденный Cache.

---

Пользователь никогда

не должен видеть ошибку,

если данные можно показать из кэша.

---

# QUEUES

Использовать:

```
Redis Streams

или

NATS

или

RabbitMQ (future)
```

---

# EMAIL QUEUE

PDF

XML

ZIP

не отправляются синхронно.

---

Создается Job.

---

Worker отправляет письмо.

---

# FILE GENERATION

Все PDF/XML/ZIP создаются:

```
RAM

↓

Send

↓

Destroy
```

---

Запрещено сохранять временные файлы.

---

# MONITORING

Обязательно:

---

Prometheus

---

Grafana

---

OpenTelemetry

---

Structured Logs

---

# LOG FORMAT

JSON

UTC

RequestID

TraceID

Latency

Status

---

# BACKUPS

Database

ежедневно.

---

Retention

30 дней.

---

Encryption

AES-256.

---

# DISASTER RECOVERY

Scenario:

```
Database Lost
```

↓

Restore Backup

↓

Replay Events

↓

Validate

↓

Resume

---

# MULTI ENVIRONMENT

```
local

↓

development

↓

staging

↓

production
```

---

Запрещено использовать Production API

в Development.

---

# DOCKER

Каждый сервис:

собственный Dockerfile.

---

# COMPOSE

Локально:

```
Go

PostgreSQL

Redis

Python

Mailhog
```

---

# PRODUCTION HOSTING

Frontend

Vercel

---

Go API

Railway / Render

---

Python Worker

Railway Cron

---

Database

Neon

---

Redis

Upstash

---

Monitoring

Grafana Cloud

---

# COST STRATEGY

Приоритет:

минимальная стоимость.

---

До 10000 пользователей

использовать:

Free Tier

*

Aggressive Cache

*

Snapshots

*

Compression

---

# SCALE STRATEGY

1000 users

↓

Single Instance

---

10000 users

↓

Redis

↓

Workers

---

50000 users

↓

Horizontal Scaling

---

100000 users

↓

Load Balancer

↓

Read Replicas

↓

CDN

---

# SECURITY

Все сервисы общаются

только через Private Network.

---

# SECRETS

Secrets Manager

или

Environment Variables.

---

Никаких ключей в Git.

---

# RELEASE STRATEGY

Blue/Green Deployment.

---

Rollback

<2 minutes.

---

# MIGRATION STRATEGY

Database Migration

всегда

до Deploy.

---

# OBSERVABILITY

Каждый запрос имеет:

```
TraceID

RequestID

UserID(optional)

Duration

CacheHit

DBQueries
```

---

# CODEX RESPONSIBILITIES

Codex обязан:

---

создать структуру проекта

в

```
~/Documents/OpenInvest
```

---

инициализировать Git;

---

создать develop;

---

создать feature branch;

---

после каждого этапа:

1.

Build;

2.

Unit Tests;

3.

Integration Tests;

4.

Review Agent;

5.

QA Agent;

6.

Security Agent;

7.

Performance Agent;

8.

Documentation Update.

---

# PUSH POLICY

Codex никогда не выполняет Push автоматически.

---

Он обязан вывести:

```
Этап завершен.

Files Changed: XX

Tests: PASS

Coverage: XX%

Performance: PASS

Security: PASS

Documentation: UPDATED

Deploy Ready: YES

Push to Repository?

[Y/N]
```

---

# FINAL INFRASTRUCTURE PRINCIPLE

> **Лучший сервер — это сервер, который не делает лишнюю работу.**

> **Лучший API — это API, который один раз получает официальные данные, безопасно кэширует их и затем обслуживает десятки тысяч пользователей без повторных обращений к внешним сервисам.**

> **OpenInvest должен масштабироваться не количеством серверов, а качеством архитектурных решений.**
