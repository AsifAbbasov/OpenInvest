# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 31

# DATABASE BIBLE

# CANONICAL DATA MODEL, ER SPECIFICATION, STORAGE STRATEGY, INDEXING, PARTITIONING & LONG TERM DATA EVOLUTION

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: DATABASE CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_31_DATABASE_BIBLE_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `57e0835dc29a9317a1d3ae8680014e62987b020e`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document is the single source of truth for the entire OpenInvest database structure.

Any database structure changes are allowed only after this document has been updated and approved by Architecture Review Agent.

---

# DATABASE PHILOSOPHY

The database must:

* be simple;
* be scalable;
* store only necessary data;
* contain no duplication;
* support millions of transactions without changing the architecture.

---

# DATABASE ENGINE

Primary DBMS:

```
PostgreSQL
```

---

# DATABASE PRINCIPLES

The following are used:

UUID Primary Keys

UTC Time

Soft Delete (where necessary)

Immutable Audit Logs

Append Only History

Normalized Data

Canonical Entities

---

# GLOBAL ER MODEL

```text
Users
 │
 ├── Profiles
 │
 ├── Settings
 │
 ├── Portfolios
 │      │
 │      ├── PortfolioAssets
 │      │
 │      ├── Transactions
 │      │
 │      ├── Snapshots
 │      │
 │      └── AnalyticsCache
 │
 ├── Notifications
 │
 ├── TaxExports
 │
 └── AuditLogs

Assets
 │
 ├── Prices
 │
 ├── Dividends
 │
 ├── Coupons
 │
 └── CorporateActions
```

---

# TABLE USERS

```
id UUID PK

email

password_hash

email_verified

premium_status

created_at

updated_at

deleted_at
```

---

It is prohibited to store:

TIN

passport

address

phone

by default.

---

# TABLE USER_PROFILE

```
user_id

first_name

last_name

country

timezone

language

currency
```

---

All fields are optional.

---

# TABLE USER_SETTINGS

```
theme

notifications

privacy_mode

tax_mode

inflation_mode

ai_enabled
```

---

# TABLE PORTFOLIOS

```
id

user_id

name

currency

created_at

updated_at
```

---

# TABLE PORTFOLIO_ASSETS

```
id

portfolio_id

ticker

quantity

average_price

average_price_currency

current_price_cache

last_calculated
```

---

Only the aggregated state is stored.

---

# TABLE TRANSACTIONS

```
id

portfolio_id

ticker

BUY

SELL

DIVIDEND

COUPON

quantity

price

currency

commission

nkd

trade_date_utc

created_at
```

---

Transactions are never modified.

---

Editing creates a new version.

---

# TABLE ASSETS

```
ticker

isin

name

sector

country

exchange

asset_type

currency

lot_size
```

---

# TABLE ASSET_PRICES

```
ticker

datetime_utc

open

high

low

close

volume
```

---

# TABLE DIVIDEND_DIRECTORY

```
ticker

announcement_date

registry_date

payment_date

amount

currency

status

source

updated_at
```

---

# TABLE COUPONS

```
bond

coupon_date

amount

currency

nkd
```

---

# TABLE SNAPSHOTS

The most important table in the system.

---

```
id

portfolio_id

snapshot_date

portfolio_value

cash

stocks_value

bonds_value

dividend_total

coupon_total

xirr

real_return

inflation_index
```

---

The phone works only with Snapshot.

---

# TABLE ANALYTICS_CACHE

```
portfolio_id

cagr

sharpe

sortino

max_drawdown

volatility

yield_on_cost

updated_at
```

---

Heavy calculations

are not performed on the client.

---

# TABLE NOTIFICATIONS

```
id

user

type

status

created

sent

read
```

---

# TABLE TAX_EXPORTS

```
id

user

year

xml_version

pdf_version

created

deleted
```

---

XML

is not stored permanently.

---

Only metadata are stored.

---

# TABLE AUDIT_LOGS

Append Only.

---

```
id

trace_id

user

action

entity

old_value_hash

new_value_hash

created_at
```

---

Modification is prohibited.

---

# TABLE FEATURE_FLAGS

```
name

enabled

environment

updated_at
```

---

# INDEX STRATEGY

Users

```
email
```

---

Transactions

```
(portfolio_id, trade_date)

(ticker, trade_date)
```

---

Snapshots

```
(portfolio_id, snapshot_date)
```

---

Dividends

```
(ticker, payment_date)
```

---

# PARTITION STRATEGY

Transactions

Partition By Year

---

Snapshots

Partition By Month

---

Prices

Partition By Quarter

---

Audit

Partition By Year

---

# MATERIALIZED VIEWS

Portfolio Summary

---

Dividend Calendar

---

Analytics Dashboard

---

Asset Rankings

---

Updated by Worker.

---

# STORAGE STRATEGY

Raw Data

↓

Canonical Data

↓

Aggregated Data

↓

Snapshots

↓

Cache

---

# SOFT DELETE

Allowed:

Users

Profiles

Notifications

---

Prohibited:

Transactions

AuditLogs

Snapshots

DividendHistory

---

# IMMUTABLE TABLES

Transactions

Snapshots

AuditLogs

DividendHistory

---

Append only.

---

# DATA RETENTION

Audit

10 years

---

Transactions

without limit

---

Snapshots

without limit

---

Notifications

2 years

---

Logs

90 days

---

# FOREIGN KEYS

All FKs are mandatory.

---

No orphan records.

---

# MIGRATION STRATEGY

Expand

↓

Populate

↓

Switch

↓

Validate

↓

Remove

---

Direct DROP COLUMN is prohibited.

---

# PERFORMANCE TARGETS

Portfolio Query

<50 ms

---

Dashboard

<100 ms

---

Snapshot Insert

<20 ms

---

Analytics Cache

<30 ms

---

# FUTURE READY

Without changing the architecture, the following can be added:

ETF

↓

Funds

↓

Deposits

↓

Gold

↓

Real Estate

↓

Crypto Tracking

↓

Family Accounts

↓

Goals

↓

Budgets

---

# DATABASE REVIEW CHECKLIST

Before every migration, Architecture Agent must check:

---

Is there duplication?

---

Can the number of tables be reduced?

---

Can a JOIN be replaced with Snapshot?

---

Can a calculation be replaced with Cache?

---

Can the index be reduced?

---

Can the migration be avoided?

---

# FINAL DATABASE PRINCIPLE

> **PostgreSQL is not merely a data store, but the Canonical Financial Ledger of OpenInvest.**

> **All calculations must be built around immutable transactions, aggregated Snapshot, and mathematical models, rather than constant recalculation of history.**

> **A properly designed database must allow tens of millions of financial operations to be served without changing its fundamental structure.**
