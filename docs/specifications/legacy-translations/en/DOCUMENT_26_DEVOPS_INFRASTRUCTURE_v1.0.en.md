# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 26

# DEVOPS, INFRASTRUCTURE, SERVER ARCHITECTURE, FREE API STRATEGY, RATE LIMITING, DISASTER RECOVERY & PRODUCTION OPERATIONS

Version: 1.0

Status: APPROVED

Priority: CRITICAL

Classification: PRODUCTION FOUNDATION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_26_DEVOPS_INFRASTRUCTURE_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `fbe495607ffcf512669a3f6cf811a0edb6c610c3`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# PURPOSE

This document defines the complete OpenInvest server architecture.

Primary goal:

**build a product capable of serving from 10 to 100 000+ users without official APIs being blocked, without a sharp increase in infrastructure cost, and without performance degradation.**

---

# GLOBAL PHILOSOPHY

OpenInvest never operates according to the scheme:

```
10000 users

↓

10000 requests

↓

MOEX API
```

This is a guaranteed path to blocking.

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

The user never

connects directly

to MOEX,

the Bank of Russia,

Rosstat,

or other APIs.

---

All requests go only through:

```
Go Backend
```

---

# OFFICIAL DATA SOURCES

Only the following are used:

---

MOEX ISS

---

Bank of Russia

---

Rosstat

---

official issuer disclosures

---

official FTS XSDs

---

# PROHIBITED SOURCES

It is prohibited to build critical functionality on:

---

unofficial parsers;

---

closed APIs;

---

non-public endpoints;

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

Python Worker is responsible for:

---

dividends;

---

inflation;

---

exchange rates;

---

official disclosures;

---

tax structures.

---

Go Worker is responsible for:

---

snapshots;

---

cache;

---

aggregation;

---

queues.

---

# UPDATE STRATEGY

## Market Data

Every 5 minutes

only during trading hours.

---

## Dividend Directory

Every 30 minutes.

---

## Inflation

1 time per day.

---

## FX

1 time per day after publication by the Bank of Russia.

---

## Company Directory

1 time per week.

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

Redis is used only for:

---

quotes;

---

aggregated data;

---

sessions;

---

feature flags.

---

It is prohibited to store:

passport;

INN;

address.

---

# DATABASE POLICY

PostgreSQL is

the only source of user data.

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

Snapshot is

a key architectural idea of the project.

---

The phone

does not recalculate

5000 transactions.

---

The phone receives:

```
Date

↓

Portfolio Value

↓

Render
```

---

# TIME ZONES

All servers operate

exclusively in UTC.

---

# MARKET CALENDAR

The source of truth is

the MOEX trading calendar.

---

NOT:

the user's local time.

---

# GLOBAL USERS

A user may live:

---

in the USA;

---

in Australia;

---

in Japan;

---

in Germany.

---

They always receive

the latest confirmed Snapshot.

---

# API GATEWAY

All clients use:

```
/api/v1/*
```

---

Direct connections

to internal services are prohibited.

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

Use:

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

If an official API is unavailable,

Go must return:

the latest confirmed Cache.

---

The user should never

see an error

if data can be shown from cache.

---

# QUEUES

Use:

```
Redis Streams

or

NATS

or

RabbitMQ (future)
```

---

# EMAIL QUEUE

PDF

XML

ZIP

are not sent synchronously.

---

A Job is created.

---

A Worker sends the email.

---

# FILE GENERATION

All PDF/XML/ZIP files are created:

```
RAM

↓

Send

↓

Destroy
```

---

Saving temporary files is prohibited.

---

# MONITORING

Mandatory:

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

daily.

---

Retention

30 days.

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

Using Production API

in Development is prohibited.

---

# DOCKER

Each service:

its own Dockerfile.

---

# COMPOSE

Locally:

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

Priority:

minimum cost.

---

Up to 10000 users,

use:

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

All services communicate

only through a Private Network.

---

# SECRETS

Secrets Manager

or

Environment Variables.

---

No keys in Git.

---

# RELEASE STRATEGY

Blue/Green Deployment.

---

Rollback

<2 minutes.

---

# MIGRATION STRATEGY

Database Migration

always

before Deploy.

---

# OBSERVABILITY

Every request has:

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

Codex must:

---

create the project structure

in

```
~/Documents/OpenInvest
```

---

initialize Git;

---

create develop;

---

create a feature branch;

---

after every stage:

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

Codex never performs Push automatically.

---

It must output:

```
Stage complete.

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

> **The best server is the server that does not do unnecessary work.**

> **The best API is the API that fetches official data once, caches it safely, and then serves tens of thousands of users without repeated calls to external services.**

> **OpenInvest should scale through the quality of architectural decisions, not through the number of servers.**
