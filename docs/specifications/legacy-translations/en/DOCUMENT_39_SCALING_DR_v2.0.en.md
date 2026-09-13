# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 39

# HORIZONTAL SCALING, HIGH AVAILABILITY, COST OPTIMIZATION, DISASTER RECOVERY & ONE MILLION USERS ARCHITECTURE

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: SCALABILITY CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_39_SCALING_DR_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `9d763b791691de0d67c39839335abe18ad4d5a70`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the OpenInvest architecture for operating as load grows from MVP to millions of users without rewriting the core.

The main goal:

> **scaling must happen by changing infrastructure, not by changing business logic.**

---

# ARCHITECTURE PHILOSOPHY

It is prohibited to build the system only for the current load.

Every component must answer the question:

> "What happens if there are 100 times more users?"

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

The user must never wait for heavy calculations.

---

All complex operations:

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

are performed in advance.

---

# SNAPSHOT STRATEGY

Instead of:

```text id="a1"
Calculate

Every Request
```

use:

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

The user reads

not from Primary,

but from Replica.

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

No mass recalculations.

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

The following are used:

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

One common queue is prohibited.

---

# BACKGROUND WORKERS

Each Worker is independent.

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

The following are served through CDN:

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

Vertical scaling

is used only for MVP.

---

Then:

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

Every new feature undergoes Cost Review.

---

Builder Agent must answer:

---

how much memory does it consume?

---

how much CPU?

---

how much network traffic?

---

how many database queries?

---

can it be replaced with Snapshot?

---

can it be replaced with Event?

---

# SERVER TARGET

Backend

512 MB RAM

up to 10 000 users.

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

A plan is mandatory.

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

daily

---

Snapshots

daily

---

Audit

daily

---

Configurations

on every change

---

# RESTORE TEST

A backup is considered working

only after a successful restore.

---

Automatically:

once a week.

---

# MULTI REGION READY

The architecture must support:

---

Europe

---

Russia

---

Central Asia

---

USA

---

without code changes.

---

# OBSERVABILITY

Every service publishes:

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

Based on:

CPU

↓

Memory

↓

Queue Length

↓

Response Time

---

# FAILURE ISOLATION

If AI stops working

↓

Portfolio continues working.

---

If Tax stops working

↓

Dashboard continues working.

---

If Inflation stops working

↓

Portfolio continues working.

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

Without changing the architecture, it is possible to add:

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

Before Merge, Architecture Agent must answer:

---

Can SQL be reduced?

---

Can RAM be reduced?

---

Can a Worker be removed?

---

Can Snapshot be used?

---

Can Sync be replaced with Event?

---

Can infrastructure cost be reduced?

---

# FINAL PRINCIPAL ENGINEERING PRINCIPLE

> **OpenInvest is designed not for today's 100 users, but for future millions of financial operations.**

> **Growth in the number of users must not lead to rewriting the core, changing mathematical models, or abandoning free infrastructure at the MVP stage.**

> **Any scaling must happen exclusively by adding computing resources, Workers, Replicas, Cache, and Queue, while keeping the domain model, API contracts, and product business logic unchanged.**
