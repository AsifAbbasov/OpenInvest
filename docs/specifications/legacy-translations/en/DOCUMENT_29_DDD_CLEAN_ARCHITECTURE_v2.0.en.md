# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 29

# DOMAIN DRIVEN DESIGN (DDD), CLEAN ARCHITECTURE, MODULE ISOLATION, PLUGIN SYSTEM & LONG-TERM SCALABILITY

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: CORE ARCHITECTURE

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_29_DDD_CLEAN_ARCHITECTURE_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `8867ad85c967f0b7800ed54759436a4dde688a41`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the fundamental architecture of OpenInvest.

After analyzing the previous documents, a decision was made to abandon building the system as a set of screens.

OpenInvest is designed as a set of independent domains (Bounded Contexts) that can evolve separately and interact only through public contracts.

This will allow the product to evolve for 10–15 years without rewriting the core.

---

# 1. GLOBAL ARCHITECTURE

It is prohibited to build the application according to the principle:

```
UI

↓

API

↓

Database
```

---

The following model is used:

```
Business Domain

↓

Application Layer

↓

Infrastructure Layer

↓

Presentation Layer
```

---

Business is always at the center.

---

# 2. DOMAIN DRIVEN DESIGN

OpenInvest is divided into independent domains.

---

## Portfolio Domain

Responsible only for:

---

portfolios;

---

positions;

---

transactions;

---

value.

---

Knows nothing about taxes.

---

Knows nothing about AI.

---

Knows nothing about dividends.

---

# 3. Asset Domain

Responsible only for:

---

stocks;

---

bonds;

---

tickers;

---

sectors;

---

quotes.

---

# 4. Dividend Domain

Responsible only for:

---

payment history;

---

future payments;

---

dividend yield;

---

dividend calendar.

---

# 5. Tax Domain

Fully isolated.

---

Responsible only for:

---

XML;

---

PDF;

---

taxes;

---

Bank of Russia rates;

---

calculation history.

---

Portfolio Domain is prohibited from knowing about Tax Domain.

---

# 6. Inflation Domain

Responsible for:

---

inflation;

---

real value;

---

purchasing power index.

---

# 7. Analytics Domain

Responsible for:

---

XIRR;

---

TWR;

---

Sharpe;

---

Sortino;

---

Max Drawdown;

---

Volatility;

---

Dividend CAGR.

---

# 8. Notification Domain

Responsible for:

---

Email;

---

Push;

---

Reminder;

---

Dividend Alert;

---

Tax Alert.

---

# 9. AI Domain

Responsible for:

---

explaining data;

---

analytics;

---

summaries;

---

visualization.

---

AI cannot modify data.

---

# 10. USER DOMAIN

Responsible for:

---

authorization;

---

settings;

---

language;

---

currency;

---

privacy.

---

# 11. CLEAN ARCHITECTURE

Each domain consists of:

```
domain/

application/

infrastructure/

presentation/
```

---

# 12. DEPENDENCY RULE

The inner layer

does not know

about the outer layer.

---

```
UI

↓

Application

↓

Domain

```

is allowed.

---

```
Domain

↓

UI
```

is prohibited.

---

# 13. INTERFACES

Any interaction

takes place through an interface.

---

No direct dependencies.

---

# 14. PLUGIN SYSTEM

OpenInvest is designed as an extensible system.

---

Today:

```
Stocks

Bonds
```

---

Tomorrow:

```
ETF

Gold

Crypto

Deposits

Real Estate
```

---

For this, a Plugin API is introduced.

---

# 15. PLUGIN CONTRACT

Every new asset type must implement:

```
GetPrice()

GetHistory()

GetIncome()

GetCurrency()

GetRisk()

GetAnalytics()
```

---

# 16. FEATURE ISOLATION

Every feature can be completely disabled.

---

Example:

```
AI OFF

↓

Portfolio continues working.
```

---

```
Tax OFF

↓

Portfolio continues working.
```

---

```
Inflation OFF

↓

Dashboard continues working.
```

---

# 17. EVENT BUS

Domains communicate only through events.

---

Example:

```
TransactionCreated

↓

PortfolioUpdated

↓

AnalyticsRecalculated

↓

SnapshotCreated

↓

NotificationSent
```

---

No direct calls.

---

# 18. SNAPSHOT DOMAIN

A separate domain.

---

Responsible for:

---

creating snapshots;

---

history;

---

aggregation;

---

caching.

---

# 19. CONFIGURATION DOMAIN

All settings:

---

Feature Flags;

---

Premium;

---

Currencies;

---

Regions;

---

Localization;

---

API Limits;

---

live separately.

---

# 20. MULTI COUNTRY READY

The architecture must not depend on the Russian Federation.

---

All specific rules:

---

taxes;

---

dividends;

---

currencies;

---

calendars;

---

are moved into Country Provider.

---

```
RussiaProvider

↓

Future

↓

USAProvider

↓

EUProvider

↓

KazakhstanProvider
```

---

# 21. TESTABILITY

Each domain is tested independently.

---

Unit

↓

Integration

↓

Contract

↓

E2E

---

# 22. CONTRACT TESTING

Frontend

and

Backend

do not test the implementation.

---

They test the contract.

---

OpenAPI —

the single source of truth.

---

# 23. INTERNAL PACKAGE RULES

Prohibited:

```
portfolio

↓

import tax

↓

import analytics

↓

import notifications
```

---

Allowed:

```
portfolio

↓

publish event

↓

event bus

↓

analytics subscribe
```

---

# 24. MODULE MATURITY

Each module has a status:

```
Experimental

↓

Beta

↓

Stable

↓

Deprecated

↓

Archived
```

---

# 25. BACKWARD COMPATIBILITY

Removing an API is prohibited.

---

The following strategy is used:

```
Deprecated

↓

Migration

↓

Replacement

↓

Removal
```

---

# 26. CODE OWNERSHIP

Builder Agent

↓

Review Agent

↓

QA Agent

↓

Security Agent

↓

Performance Agent

↓

Human

---

Push is allowed only after the entire chain has been completed.

---

# 27. FILE SIZE RULE

A file must not exceed:

```
300–400 lines
```

---

A package

must have no more than

one responsibility.

---

# 28. COMPONENT SIZE RULE

React component

no more than

250 lines.

---

Hook

no more than

150 lines.

---

Service

no more than

300 lines.

---

# 29. LONG TERM EVOLUTION

In 5 years, the project must allow adding a new asset

without changing the existing Portfolio Domain.

---

This is achieved through:

---

Open/Closed Principle;

---

Plugin API;

---

DDD;

---

Event Bus;

---

Clean Architecture.

---

# 30. FINAL PRINCIPAL ENGINEERING RULE

> **OpenInvest is not a monolithic website.**

> **OpenInvest is a set of independent financial domains united by a common mathematical model, a single API contract, and an event-driven architecture.**

> **Adding a new feature must not result in changing the already working core. Any violation of this rule is considered an architectural defect and must be rejected by Review Agent before the code reaches the main branch.**
