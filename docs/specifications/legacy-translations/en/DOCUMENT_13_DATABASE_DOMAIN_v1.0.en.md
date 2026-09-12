# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 13

# DATABASE, DOMAIN MODEL & BUSINESS LOGIC BLUEPRINT

Version: 1.0

Status: Core Domain

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_13_DATABASE_DOMAIN_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `93e184d218690741be77d10a1259d942786e0466`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# 1. PURPOSE

This document defines:

* database structure;
* business entities;
* relationships between them;
* storage rules;
* object lifecycle;
* the mathematical model for storing an investment portfolio.

This document is the primary source of truth for Backend, Frontend, Android, iOS, and AI Agents.

---

# 2. DOMAIN PHILOSOPHY

OpenInvest does not store "pages."

OpenInvest stores business entities.

Each entity must have:

* clear responsibility;
* an independent lifecycle;
* its own history;
* extensibility.

---

# 3. CORE DOMAIN

```
User
│
├── Portfolio
│      │
│      ├── Transactions
│      ├── Holdings
│      ├── Snapshots
│      ├── Cash
│      └── Statistics
│
├── Watchlist
│
├── Dividend Calendar
│
├── Tax Assistant
│
├── Notifications
│
└── Settings
```

---

# 4. USER

User is the owner of all data.

User never stores investment data directly.

User is a container.

---

Fields:

```
id

email

password_hash

created_at

updated_at

language

timezone

currency

is_premium

status
```

---

# 5. PROFILE

A separate entity.

```
profile

first_name

last_name

avatar

country

locale

theme

number_format

date_format
```

---

# 6. TAX PROFILE

A separate table.

No relationships with Portfolio except user_id.

```
tax_profile

inn

passport

address

country

residency

store_mode

updated_at
```

---

# 7. PORTFOLIO

A user may have several portfolios.

Examples:

```
Main

Dividends

Retirement

Child

ETF

Bonds

USD Portfolio
```

---

# 8. PORTFOLIO MODEL

```
Portfolio

↓

Holdings

↓

Transactions

↓

Snapshots

↓

Statistics
```

---

# 9. TRANSACTION

Transaction is the single source of truth.

Any portfolio change happens only through a new transaction.

---

Types:

```
BUY

SELL

DIVIDEND

COUPON

COMMISSION

CASH_IN

CASH_OUT

TAX

SPLIT

MERGE
```

---

# 10. IMMUTABILITY

Historical transactions must not be modified directly.

Editing:

creates a new revision.

or

creates a compensating operation.

---

# 11. HOLDINGS

Holdings is aggregated state.

```
ticker

quantity

average_price

market_price

market_value

profit

profit_percent

xirr

dividend_income
```

---

Recalculated automatically.

---

# 12. SNAPSHOTS

Store daily state.

```
date

total_value

cash

stocks

bonds

expected_dividends

received_dividends

real_value_after_inflation
```

---

# 13. CASH ACCOUNT

Money is a separate asset.

```
RUB

USD

EUR

CNY
```

---

The user can account for available cash.

---

# 14. ASSET TYPES

Supported:

```
Stocks

Bonds

ETF

Funds

Currencies

Cash

Future Assets
```

---

The architecture must allow adding a new type without rewriting the system.

---

# 15. SECURITY MASTER

```
ticker

isin

sector

exchange

currency

issuer

country

type
```

---

Independent of the user.

---

# 16. DIVIDEND DIRECTORY

```
ticker

registry_date

payment_date

amount

currency

yield

status

official_source

updated_at
```

---

# 17. COUPON DIRECTORY

For bonds.

```
isin

coupon

payment_date

nkd

yield

status
```

---

# 18. WATCHLIST

A separate entity.

Independent of the portfolio.

---

It is possible to store

100+

securities

without purchasing them.

---

# 19. NOTIFICATIONS

Types:

```
Dividend

Coupon

Price

Tax

AI

Portfolio

Market

News
```

---

# 20. AI INSIGHTS

Stored separately.

```
recommendation

source

confidence

created_at

status
```

---

AI never modifies the portfolio.

---

# 21. AUDIT

Any action:

```
User

↓

Action

↓

Audit
```

---

# 22. VERSIONING

Each record has:

```
created_at

updated_at

version
```

---

# 23. SOFT DELETE

Deletion:

```
deleted_at

deleted_by

reason
```

---

Physical deletion only by Background Worker.

---

# 24. RELATIONS

```
User

1:N

Portfolio

1:N

Transactions

1:N

Snapshots

1:N

Statistics
```

---

# 25. NORMALIZATION

All directories:

3NF.

---

History:

append only.

---

# 26. INDEXES

```
user_id

portfolio_id

ticker

isin

trade_date

snapshot_date

payment_date
```

---

# 27. PARTITIONING

transactions

by year

---

audit_logs

by month

---

snapshots

by year

---

notifications

by quarter

---

# 28. MATERIALIZED VIEWS

Create:

```
portfolio_summary

dividend_summary

sector_summary

tax_summary
```

---

This significantly reduces load.

---

# 29. BUSINESS RULES

BUY:

increases quantity

recalculates Average Cost

---

SELL:

decreases quantity

does not change Average Cost

---

DIVIDEND:

increases Cash

increases Income

---

COUPON:

increases Cash

---

COMMISSION:

decreases Cash

decreases Income

---

# 30. INFLATION MODEL

In addition to nominal value, store:

```
Real Value

Inflation Loss

Purchasing Power

Real Return
```

---

# 31. PURCHASING POWER

An additional metric.

Examples:

```
Your portfolio:

2.3 MacBook Pro

5 iPhone

17 average salaries

9 months of apartment rent

31 food baskets
```

---

Not used in mathematics.

Used for visualization.

---

# 32. HISTORICAL CONSISTENCY

Prohibited:

recalculating history using today's data.

---

Use only:

historical prices;

historical dividends;

historical CBR exchange rates;

historical inflation.

---

# 33. MULTI CURRENCY

All assets have:

```
Original Currency

Base Currency

Converted Currency
```

---

Conversion is performed using the historical exchange rate.

---

# 34. FREE VS PREMIUM

FREE

up to 5 companies

1 portfolio

basic analytics

---

PREMIUM

unlimited

multiple portfolios

AI analytics

tax assistant

export

advanced statistics

---

# 35. EXTENSIBILITY

Adding a new asset should not require changing:

Portfolio

Transactions

Snapshots

Charts

Statistics

---

It is sufficient to add a new AssetType.

---

# 36. FUTURE MODULES

The architecture anticipates in advance:

```
Crypto

US Stocks

EU Stocks

Kazakhstan

Bonds

REIT

Gold

Silver

Commodities
```

---

# 37. SUCCESS CRITERIA

After 5 years of development, the database should support:

* millions of transactions;
* hundreds of thousands of users;
* tens of millions of snapshot records;
* multicurrency portfolios;
* multiple tax jurisdictions;

without changing the core data model.

---

# 38. CODEX REQUIREMENT

Codex must implement new functions only through the Domain Model.

Prohibited:

creating "quick hacks";

duplicating data;

storing business logic in Frontend;

calculating mathematics in React;

violating Single Source of Truth.

This document is mandatory for Builder Agent, Review Agent, and QA Agent.
