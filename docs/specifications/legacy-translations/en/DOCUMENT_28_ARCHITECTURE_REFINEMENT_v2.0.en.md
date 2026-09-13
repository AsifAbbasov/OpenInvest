# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 28

# ARCHITECTURE REFINEMENT, DATA GOVERNANCE, API STRATEGY, ADVANCED MATHEMATICS, ADR, C4, ER MODEL, OPENAPI-FIRST & LONG TERM EVOLUTION

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: PRINCIPAL ENGINEERING ADDENDUM

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_28_ARCHITECTURE_REFINEMENT_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `00dfe3c6cfa4264555940c63557e59bc1698caf0`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document replaces and expands the previous architectural decisions.

After analyzing all documents, potential architectural risks were identified that must be eliminated before development begins.

The purpose of the document is to:

* remove logical contradictions;
* ensure scaling to 100 000+ users;
* minimize operating costs;
* ensure legal cleanliness;
* make the product extensible for 10+ years.

---

# 1. PRODUCT REPOSITIONING

## Old positioning

```
Dividend Calculator
```

↓

incorrect.

---

## New positioning

```
OpenInvest

Personal Capital Operating System

(Personal Capital OS)
```

---

OpenInvest —

is a personal capital operating system.

---

Today it supports

* stocks;
* bonds;
* dividends.

---

Tomorrow

* ETF;
* REIT;
* precious metals;
* bank deposits;
* real estate;
* pension savings;
* family portfolios;
* goals;
* cash flows;
* tax accounting.

---

# 2. DATA GOVERNANCE

The most important architectural idea of the project.

---

## The user never works with external APIs.

```
Official APIs

↓

Collector Layer

↓

Normalization Layer

↓

Validation Layer

↓

Canonical Database

↓

Redis / RAM

↓

Go API

↓

Client
```

---

OpenInvest becomes the owner of its own data model.

---

# 3. CANONICAL DATA MODEL

Never use an external structure directly.

---

Every source goes through:

```
Raw Data

↓

Parser

↓

Validator

↓

Normalizer

↓

Canonical Entity

↓

Storage
```

---

This makes it possible to:

change the data provider;

use multiple sources;

eliminate dependence on MOEX.

---

# 4. OFFICIAL API STRATEGY

Only official free sources are used.

---

But the user never accesses them directly.

---

Python Worker receives the data.

---

Go API serves users.

---

# 5. DATA SOURCE LIFECYCLE

Every source undergoes mandatory review.

```
Business Review

↓

Legal Review

↓

License Review

↓

Technical Review

↓

Security Review

↓

Performance Review

↓

Approval
```

---

Builder Agent is prohibited from connecting new APIs independently.

---

# 6. JURIDICAL DATA POLICY

For every API, the following must be stored:

```
Source

License

Commercial Usage

Redistribution

Storage Rules

Caching Rules

Retention Policy

Expiration Policy
```

---

# 7. CACHE STRATEGY V2

A four-level system is used.

```
L1 RAM

↓

L2 Redis

↓

L3 PostgreSQL

↓

L4 Official Source
```

---

The user always receives data from the nearest level.

---

# 8. EVENT DRIVEN ARCHITECTURE

Instead of constant requests.

```
Dividend Approved

↓

Event

↓

Queue

↓

Portfolio Update

↓

Notification

↓

Tax Update
```

---

# 9. ADVANCED MATHEMATICAL ENGINE

OpenInvest becomes a professional analytical system.

---

## Weighted Average Cost

Weighted average cost.

---

## FIFO

For comparison.

---

## XIRR

Money Weighted Return.

---

## Time Weighted Return (TWR)

Allows comparing the investor with an index.

---

## CAGR

Compound Annual Growth Rate.

---

Average annual capital growth rate.

---

## Dividend CAGR

Average annual growth of a company's dividends.

---

## Yield on Cost

Dividend yield relative to one's own purchase price.

A very important metric for a long-term investor.

---

## Sharpe Ratio

Return relative to total risk.

---

## Sortino Ratio

Return relative to downside risk.

---

## Maximum Drawdown

Maximum historical drawdown.

---

## Volatility

Standard deviation of returns.

---

## Inflation Adjusted Return

```
Portfolio Return

-

Inflation
```

---

## Tax Adjusted Return

```
Portfolio Return

-

Taxes
```

---

## Net Real Return

```
Portfolio Return

-

Inflation

-

Taxes

-

Commissions

-

NKD
```

---

# 10. PURCHASING POWER INDEX

An exclusive OpenInvest feature.

---

Instead of:

```
Portfolio

1 850 000 ₽
```

show

```
Real Capital

MacBook Pro

9.4

iPhone

22

Average salary

13 months

Grocery basket

31 months

Apartment rent

19 months
```

---

And show the change in purchasing power over time.

---

# 11. ARCHITECTURE DECISION RECORDS

Every decision is recorded.

---

Example:

```
ADR-001

Why Go

Status

Accepted

Reasons

Consequences

Alternatives
```

---

Minimum ADR set:

```
Go

Python

PostgreSQL

Redis

Feature Folder

Private Mode

API First

Snapshot

XIRR

Canonical Data

Event Driven
```

---

# 12. ER MODEL

Minimum data model.

```
Users

↓

Profiles

↓

Portfolios

↓

Transactions

↓

Assets

↓

Snapshots

↓

DividendDirectory

↓

Coupons

↓

Notifications

↓

TaxExports

↓

AuditLogs

↓

FeatureFlags
```

---

The ER model must exist separately

before coding begins.

---

# 13. C4 MODEL

Mandatory.

---

## Level 1

System Context

---

## Level 2

Containers

Frontend

Go

Python

Redis

PostgreSQL

SMTP

---

## Level 3

Components

Portfolio

Tax

Catalog

AI

Notifications

Auth

Analytics

---

## Level 4

Code

Packages

Modules

Interfaces

---

# 14. SEQUENCE DIAGRAMS

Must be created.

---

Adding a transaction.

---

Creating a portfolio.

---

Generating XML.

---

Sending Email.

---

Receiving dividends.

---

Updating Snapshot.

---

AI analysis.

---

# 15. OPENAPI FIRST

Development begins

not with code.

---

Sequence:

```
Product

↓

Domain

↓

OpenAPI

↓

Swagger

↓

SDK

↓

Backend

↓

Frontend

↓

Mobile
```

---

Writing the Backend is prohibited

until OpenAPI is approved.

---

# 16. VERSIONING

```
v1

↓

v2

↓

v3
```

---

No breaking changes.

---

# 17. DATABASE EVOLUTION

Use the

Expand / Migrate / Contract strategy.

---

```
Add Column

↓

Fill Data

↓

Switch Reads

↓

Switch Writes

↓

Remove Old Column
```

---

No destructive migrations.

---

# 18. PREMIUM STRATEGY V2

Premium

must not take away basic functionality.

---

Free:

unlimited transactions;

unlimited portfolio;

dividends;

XIRR;

Real Return;

calendar;

inflation.

---

Premium:

Monte Carlo;

Stress Test;

Goal Planner;

Retirement Planner;

AI Portfolio Review;

Dividend Scenarios;

Tax Forecast;

Family Office;

Advanced Reports;

Portfolio Compare.

---

# 19. COMPETITIVE STRATEGY

Do not compete:

with BCS;

with Alfa;

with T-Bank;

with VTB;

by number of features.

---

Compete:

on speed;

on transparency;

on privacy;

on mathematics;

on energy efficiency;

on honesty.

---

# 20. ENGINEERING CHECKLIST

Before implementing any feature, Builder Agent must answer:

---

Can it be made simpler?

---

Can RAM be reduced?

---

Can CPU usage be reduced?

---

Can traffic be reduced?

---

Can data storage be avoided?

---

Can the number of requests be reduced?

---

Can it be implemented through Snapshot?

---

Can it be implemented through Event?

---

Can synchronous processing be avoided?

---

# 21. FINAL PRINCIPLE

In 10 years, the architecture must allow:

```
10 000 000+

Transactions

↓

500 000+

Users

↓

Millions of Snapshots

↓

Dozens of Asset Classes

↓

Without rewriting Core Architecture
```

---

# FINAL PRINCIPAL ENGINEER STATEMENT

> OpenInvest must be designed not as a website and not as a dividend calculator.

> It must be designed as a **Personal Capital Platform**, where dividends, taxes, analytics, inflation, and capital management are separate independent domains united by a common architecture, a unified mathematical model, official data sources, strict privacy, and an API-First approach.

> Any decision that may lead to rewriting the core in 3–5 years is considered an architectural error and must be rejected at the design stage.
