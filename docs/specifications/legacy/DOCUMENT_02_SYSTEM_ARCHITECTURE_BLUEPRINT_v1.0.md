# SYSTEM ARCHITECTURE BLUEPRINT

Version: 1.0

Status: Source of Truth

Priority: CRITICAL

---

# 1. SYSTEM OVERVIEW

OpenInvest представляет собой распределенную модульную систему.

Все компоненты независимы.

Никакой компонент не должен являться единой точкой отказа.

---

Architecture Style

API First

Domain Driven Design

Feature First

Clean Architecture

Hexagonal Architecture

Event Driven Architecture

Background Processing

CQRS Ready

---

# 2. HIGH LEVEL ARCHITECTURE

CLIENTS

Web

↓

PWA

↓

iOS

↓

Android

↓

Public API

↓

---

API GATEWAY

Authentication

Authorization

Rate Limit

Compression

Logging

Request Validation

---

GO BACKEND

Portfolio Engine

Tax Engine

Market Engine

Notification Engine

Export Engine

User Engine

---

BACKGROUND SERVICES

Python Parser

Dividend Parser

Inflation Parser

Currency Parser

Analytics Worker

Notification Worker

---

DATABASE

PostgreSQL

---

CACHE

Redis

Local Memory Cache

---

OBJECT STORAGE

Generated PDF

Generated XML

Temporary ZIP

---

MONITORING

Metrics

Logs

Tracing

Health

Alerts

---

BACKUP

Daily

Weekly

Monthly

---

# 3. PROJECT STRUCTURE

root/

backend-go/

microservice-python/

frontend-react/

mobile-ios/

mobile-android/

shared-contracts/

documentation/

infrastructure/

scripts/

docker/

.github/

---

# 4. BACKEND STRUCTURE

backend-go/

cmd/

internal/

api/

application/

domain/

infrastructure/

repository/

service/

events/

jobs/

middleware/

cache/

config/

utils/

tests/

docs/

---

# 5. FRONTEND STRUCTURE

frontend-react/

src/

app/

common/

features/

layouts/

routes/

providers/

hooks/

assets/

styles/

tests/

public/

---

# 6. FEATURE STRUCTURE

Each Feature:

api/

model/

ui/

hooks/

types/

utils/

tests/

README.md

Every feature must be isolated.

Cross feature imports are prohibited.

---

# 7. DATABASE LAYERS

Identity Database

contains:

email

password hash

2FA

settings

---

Investment Database

contains:

transactions

portfolios

snapshots

watchlists

---

Tax Database

contains:

temporary declarations

temporary generated files

audit logs

---

Analytics Database

contains:

market history

inflation history

exchange rates

dividend history

---

# 8. SECURITY MODEL

Passwords:

Argon2id

Sessions:

JWT + Refresh Token

2FA:

TOTP

Transport:

TLS

Database:

Encrypted

Backups:

Encrypted

Audit:

Immutable

---

# 9. DATA CLASSIFICATION

Public

Market Data

---

Internal

Portfolio Metadata

---

Confidential

Transactions

---

Sensitive

Email

Password Hash

---

Optional Sensitive

Passport

INN

Address

These fields are NEVER mandatory.

---

# 10. PRIVACY MODEL

Default:

No passport

No INN

No address

No phone

---

Tax Mode:

User enters data

↓

Generate document

↓

User chooses:

Save

or

Delete Immediately

---

Delete Immediately:

Generated XML removed

Generated PDF removed

Temporary storage removed

Tax session removed

Only audit event remains.

---

# 11. CACHE STRATEGY

RAM Cache

Fastest

---

Redis

Frequently used queries

---

PostgreSQL

Source of truth

---

# 12. MOEX STRATEGY

IMPORTANT

Frontend NEVER communicates with MOEX.

Only backend communicates.

Python parser never receives requests from clients.

All requests pass through cache.

---

# 13. RATE LIMIT PROTECTION

Client

↓

API Gateway

↓

Redis Rate Limit

↓

Go Cache

↓

Database

↓

MOEX

MOEX is always the LAST source.

Never first.

---

# 14. MARKET DATA POLICY

Priority

RAM Cache

↓

Redis

↓

PostgreSQL

↓

Official API

↓

Retry

↓

Background refresh

---

# 15. SNAPSHOT ENGINE

Snapshots are generated after official market close.

Timezone:

UTC

Reference:

MOEX trading calendar

Not user timezone.

---

# 16. OBSERVABILITY

Every request:

RequestID

UserID

Duration

Memory

CPU

Result

Error

TraceID

---

# 17. BACKUP STRATEGY

Daily incremental

Weekly full

Monthly archive

Quarterly verification

Annual restore test

---

# 18. SCALABILITY TARGET

100 000 users

1 000 000 transactions

100 000 000 snapshots

10 000 requests/minute

without architecture redesign.

---

# 19. AGENTS

Builder Agent

writes code

---

Review Agent

checks:

SOLID

SRP

ISP

DIP

DRY

KISS

YAGNI

Architecture

Performance

Security

---

QA Agent

Unit

Integration

E2E

Regression

Smoke

Accessibility

Responsive

User Journey

---

Health Agent

runs every 24 hours

checks:

API

Database

Cache

Email

Parser

Cron

Disk

Memory

CPU

Response Time

Produces report only.

Never pushes code.

---

# 20. GIT STRATEGY

main

production only

---

develop

integration

---

feature/*

builder

---

review/*

review fixes

---

qa/*

testing

---

# 21. CODEX WORKFLOW

Every task:

1.

Read documentation

↓

2.

Implement

↓

3.

Run unit tests

↓

4.

Run integration tests

↓

5.

Run e2e

↓

6.

Run lint

↓

7.

Explain architecture decision

↓

8.

Ask user:

Push to repository?

YES / NO

Never push automatically.

---

# 22. NON NEGOTIABLE RULES

No any.

No relative imports.

No business logic inside UI.

No LocalStorage for business data.

No duplicated calculations.

No hidden side effects.

No magic numbers.

No silent failures.

Every calculation must be reproducible.

Every tax result must be explainable.

Every generated document must be auditable.

END OF DOCUMENT.
