# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 39

# HORIZONTAL SCALING, HIGH AVAILABILITY, COST OPTIMIZATION, DISASTER RECOVERY & ONE MILLION USERS ARCHITECTURE

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: SCALABILITY CONSTITUTION

---

# PURPOSE

Настоящий документ определяет архитектуру OpenInvest для работы при росте нагрузки от MVP до миллионов пользователей без переписывания ядра.

Главная цель:

> **масштабирование должно происходить изменением инфраструктуры, а не изменением бизнес-логики.**

---

# ARCHITECTURE PHILOSOPHY

Запрещается строить систему с расчетом только на текущую нагрузку.

Каждый компонент должен отвечать на вопрос:

> "Что произойдет, если пользователей станет в 100 раз больше?"

---

# SCALING STAGES

## STAGE 1

MVP

```text id="s1"
Users

0 — 1 000

Go API x1

Python Worker x1

PostgreSQL x1

Redis x1
```

---

## STAGE 2

```text id="s2"
1 000 — 10 000

Load Balancer

↓

Go API x2

↓

Redis

↓

PostgreSQL

↓

Python Workers x2
```

---

## STAGE 3

```text id="s3"
10 000 — 100 000

CDN

↓

Load Balancer

↓

Go API xN

↓

Redis Cluster

↓

Queue

↓

Workers

↓

PostgreSQL Primary

↓

Read Replica
```

---

## STAGE 4

```text id="s4"
100 000+

Global CDN

↓

Regional Load Balancer

↓

Go API Pool

↓

Redis Cluster

↓

Message Queue

↓

Worker Pool

↓

PostgreSQL Cluster

↓

Read Replicas

↓

Analytics Cluster
```

---

# CORE PRINCIPLE

Пользователь никогда не должен ждать тяжелых вычислений.

---

Все сложные операции:

---

XIRR

---

Sharpe

---

Sortino

---

Tax

---

Inflation

---

Snapshot

---

выполняются заранее.

---

# SNAPSHOT STRATEGY

Вместо:

```text id="a1"
Calculate

Every Request
```

используется:

```text id="a2"
Background Worker

↓

Snapshot

↓

Redis

↓

API

↓

Client
```

---

# READ / WRITE SEPARATION

```text id="rw1"
Write

↓

Primary Database

↓

Replication

↓

Read Replica

↓

Clients
```

---

Пользователь читает

не Primary,

а Replica.

---

# CACHE STRATEGY

L1

Go Memory

---

L2

Redis

---

L3

PostgreSQL

---

L4

Official Sources

---

# CACHE TTL

Portfolio

30 sec

---

Assets

5 min

---

Dividend Calendar

30 min

---

Inflation

24 h

---

Tax Rules

24 h

---

# EVENT DRIVEN SYSTEM

Никаких массовых пересчетов.

---

```text id="ev1"
Transaction Created

↓

Queue

↓

Portfolio Worker

↓

Analytics Worker

↓

Snapshot Worker

↓

Notification Worker
```

---

# QUEUE STRATEGY

Используется:

---

Portfolio Queue

---

Analytics Queue

---

Tax Queue

---

Notification Queue

---

AI Queue

---

Запрещается одна общая очередь.

---

# BACKGROUND WORKERS

Каждый Worker независим.

---

Dividend Worker

---

Tax Worker

---

Snapshot Worker

---

Inflation Worker

---

Analytics Worker

---

AI Worker

---

# CDN STRATEGY

Через CDN отдаются:

---

JS

---

CSS

---

Fonts

---

SVG

---

Static Images

---

Documentation

---

# NEVER CACHE

Portfolio

---

Transactions

---

Tax Preview

---

Personal Data

---

# DATABASE SCALING

Вертикальное масштабирование

используется только для MVP.

---

Далее:

---

Read Replicas

---

Partitioning

---

Connection Pool

---

Archive Tables

---

# PARTITION RULES

Transactions

↓

Year

---

Snapshots

↓

Month

---

Audit

↓

Year

---

Logs

↓

Week

---

# STORAGE TIERS

Hot Data

Redis

---

Warm Data

PostgreSQL

---

Cold Data

Archive

---

# COST OPTIMIZATION

Каждая новая функция проходит Cost Review.

---

Builder Agent обязан ответить:

---

сколько памяти потребляет?

---

сколько CPU?

---

сколько сетевого трафика?

---

сколько запросов к БД?

---

можно ли заменить Snapshot?

---

можно ли заменить Event?

---

# SERVER TARGET

Backend

512 MB RAM

до 10 000 пользователей.

---

# QUERY TARGET

Dashboard

≤3 SQL

---

Portfolio

≤5 SQL

---

Asset Card

≤3 SQL

---

Tax

Background Only

---

# API TARGET

P50

<50 ms

---

P95

<120 ms

---

P99

<250 ms

---

# DISASTER RECOVERY

План обязателен.

---

Database Down

↓

Replica Promotion

↓

Recovery

---

Redis Down

↓

RAM Cache

↓

Restore

---

SMTP Down

↓

Queue

↓

Retry

---

Python Down

↓

Continue Core Functions

---

MOEX Down

↓

Last Verified Snapshot

↓

Status Banner

---

# BACKUP POLICY

Database

ежедневно

---

Snapshots

ежедневно

---

Audit

ежедневно

---

Configurations

при каждом изменении

---

# RESTORE TEST

Backup считается рабочим

только после успешного восстановления.

---

Автоматически:

раз в неделю.

---

# MULTI REGION READY

Архитектура должна поддерживать:

---

Europe

---

Russia

---

Central Asia

---

USA

---

без изменения кода.

---

# OBSERVABILITY

Каждый сервис публикует:

---

Latency

---

Memory

---

CPU

---

Queue

---

Cache Hit

---

DB Time

---

Worker Time

---

# AUTO SCALING

Основано на:

CPU

↓

Memory

↓

Queue Length

↓

Response Time

---

# FAILURE ISOLATION

Если AI перестал работать

↓

Portfolio продолжает работать.

---

Если Tax перестал работать

↓

Dashboard продолжает работать.

---

Если Inflation перестал работать

↓

Portfolio продолжает работать.

---

# SLO

Availability

99.9%

---

Dashboard

99.95%

---

Portfolio Read

99.95%

---

Tax Export

99%

---

# SLA INTERNAL

P1

15 min

---

P2

1 hour

---

P3

24 hours

---

# FUTURE ROADMAP

Без изменения архитектуры возможно добавить:

---

Multi Currency

---

Multi Broker Import

---

Family Accounts

---

Corporate Accounts

---

Advisor Portal

---

Public API

---

AI Copilot

---

# ARCHITECTURE REVIEW QUESTIONS

Перед Merge Architecture Agent обязан ответить:

---

Можно ли уменьшить SQL?

---

Можно ли уменьшить RAM?

---

Можно ли убрать Worker?

---

Можно ли использовать Snapshot?

---

Можно ли заменить Sync на Event?

---

Можно ли уменьшить стоимость инфраструктуры?

---

# FINAL PRINCIPAL ENGINEERING PRINCIPLE

> **OpenInvest проектируется не для сегодняшних 100 пользователей, а для будущих миллионов финансовых операций.**

> **Рост количества пользователей не должен приводить к переписыванию ядра, изменению математических моделей или отказу от бесплатной инфраструктуры на этапе MVP.**

> **Любое масштабирование должно происходить исключительно через добавление вычислительных ресурсов, Worker'ов, Replica, Cache и Queue, сохраняя неизменными доменную модель, API-контракты и бизнес-логику продукта.**
