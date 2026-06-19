# DOCUMENT 04

# SYSTEM DESIGN & ENGINEERING ARCHITECTURE

Version 1.0

Status: Source Of Truth

Priority: CRITICAL

---

# 1. ARCHITECTURE PHILOSOPHY

OpenInvest проектируется не как сайт.

OpenInvest проектируется как распределенная финансовая платформа.

Web является одним из клиентов.

iOS является одним из клиентов.

Android является одним из клиентов.

Public API является одним из клиентов.

Все используют одну бизнес-логику.

---

# 2. ARCHITECTURE STYLE

Clean Architecture

Feature First

DDD

API First

Event Driven

Hexagonal Architecture

CQRS Ready

Background Processing

Stateless Backend

Immutable Financial History

---

# 3. PHYSICAL STRUCTURE

root/

backend-go/

frontend-react/

microservice-python/

mobile-ios/

mobile-android/

contracts/

documentation/

docker/

infra/

scripts/

.github/

---

# 4. BACKEND STRUCTURE

backend-go/

cmd/

internal/

config/

bootstrap/

application/

domain/

repository/

service/

events/

scheduler/

middleware/

cache/

transport/

api/

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

providers/

router/

hooks/

styles/

assets/

tests/

public/

---

# 6. FEATURE STRUCTURE

Every Feature:

README.md

api/

model/

ui/

hooks/

types/

utils/

tests/

constants/

selectors/

validators/

No feature can directly modify another feature.

Only public interfaces.

---

# 7. DEPENDENCY RULE

UI

↓

Application

↓

Domain

↓

Repository

↓

Infrastructure

Dependencies never go upward.

---

# 8. REQUEST FLOW

User

↓

React

↓

Axios

↓

API Gateway

↓

Fiber

↓

Application Service

↓

Domain Service

↓

Repository

↓

PostgreSQL

↓

Response

↓

Cache

↓

React

---

# 9. BACKGROUND FLOW

Cron

↓

Python Parser

↓

Validation

↓

Normalization

↓

Database

↓

Event

↓

Go Backend

↓

Cache Refresh

↓

Notification Queue

↓

Email

↓

Push

---

# 10. CACHE STRATEGY

Level 1

Browser Cache

---

Level 2

React Query Cache

---

Level 3

Go Memory Cache

---

Level 4

Redis

---

Level 5

PostgreSQL

---

Level 6

Official API

---

Priority:

Always nearest cache.

Never official API if cache exists.

---

# 11. MARKET DATA STRATEGY

Frontend

never

calls MOEX.

Only backend.

---

Backend

checks

Memory Cache

↓

Redis

↓

PostgreSQL

↓

Official API

---

This protects from rate limits.

---

# 12. API GATEWAY

Responsibilities:

Authentication

Authorization

Compression

Rate Limit

Request Validation

Logging

Versioning

Correlation ID

---

# 13. AUTHENTICATION

Access Token

Refresh Token

Device ID

Optional 2FA

Session Rotation

---

# 14. AUTHORIZATION

Guest

User

Premium

Admin

System

Worker

Review Agent

QA Agent

---

# 15. DATABASES

Identity DB

Investment DB

Analytics DB

Tax DB

Audit DB

---

No table should become universal storage.

---

# 16. STORAGE STRATEGY

Permanent

Transactions

Snapshots

Assets

Dividends

---

Temporary

Generated PDF

Generated XML

Generated ZIP

---

Temporary data expires automatically.

---

# 17. EVENT BUS

Every important action produces event.

TransactionCreated

↓

PortfolioRecalculationRequested

↓

SnapshotCreated

↓

TaxRecalculationRequested

↓

NotificationRequested

↓

EmailRequested

↓

AuditWritten

---

# 18. OBSERVABILITY

Every request contains

RequestID

TraceID

UserID

Version

Duration

Memory

CPU

CacheHit

DatabaseTime

---

# 19. ERROR STRATEGY

Never return

500 Unknown

Always explain

Validation

Authentication

Permission

Timeout

Dependency

Business Rule

---

# 20. RETRY POLICY

MOEX unavailable

↓

Retry

↓

Retry

↓

Retry

↓

Cached response

↓

Notification

Never spam external APIs.

---

# 21. EXTERNAL PROVIDERS

Official Market Data

Official Currency

Official Inflation

Official Tax Formats

Official Calendar

All providers are wrapped behind interfaces.

Provider can be replaced without changing business logic.

---

# 22. IMPORT SYSTEM

Future ready

CSV

Excel

Broker Statements

Manual Entry

API

---

Every importer writes Transactions.

Never writes Portfolio directly.

---

# 23. EXPORT SYSTEM

PDF

XML

CSV

Excel

JSON

ZIP

Email

Share Sheet

---

# 24. MOBILE STRATEGY

Mobile contains zero business logic.

Only:

Display

Offline cache

Sync

Notifications

File sharing

Everything else on backend.

---

# 25. OFFLINE MODE

User can

view portfolio

view charts

view last snapshots

view dividends

without network.

Editing requires synchronization.

---

# 26. PERFORMANCE TARGETS

Application open

< 1 second

Portfolio

< 300 ms

Chart

< 500 ms

Search

< 100 ms

Dividend calendar

< 300 ms

Tax generation

< 5 seconds

---

# 27. MEMORY TARGETS

Go

< 100 MB

Python Worker

< 256 MB

React

minimum bundle

lazy loading

code splitting

---

# 28. CODING PRINCIPLES

SOLID

SRP

ISP

LSP

DIP

DRY

KISS

YAGNI

Law of Demeter

Open Closed

Composition over Inheritance

Pure Functions

Explicit Dependencies

---

# 29. CODE REVIEW AGENT

Checks

Architecture

Performance

Naming

Complexity

Memory

Security

Tests

Documentation

Public API

No code can be merged without Review Agent approval.

---

# 30. QA AGENT

Runs automatically

Unit

Integration

E2E

Regression

Smoke

Accessibility

Responsive

Visual Regression

Cross Browser

Cross Platform

Performance

Load Tests

---

# 31. HEALTH AGENT

Runs every 24 hours

Checks

API

Database

Redis

Cron

Email

Parser

Disk

CPU

Memory

Snapshots

Tax Engine

Notification Queue

Produces report.

Never modifies code.

Never pushes commits.

---

# 32. SELF CRITICISM

Current architecture supports

100 000 users

without redesign.

Future migration to Kubernetes is possible.

Redis remains optional.

Event Bus can later migrate to Kafka/NATS.

CQRS can be enabled gradually.

No vendor lock.

---

# 33. FUTURE EXPANSION

Broker Synchronization

Open API

Plugins

AI Assistant

Multi Country Taxes

ETF

Crypto

US Market

EU Market

Financial Goals

Family Portfolios

Institutional Accounts

END OF DOCUMENT
