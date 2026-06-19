# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 31

# DATABASE BIBLE

# CANONICAL DATA MODEL, ER SPECIFICATION, STORAGE STRATEGY, INDEXING, PARTITIONING & LONG TERM DATA EVOLUTION

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: DATABASE CONSTITUTION

---

# PURPOSE

Настоящий документ является единственным источником истины для всей структуры базы данных OpenInvest.

Любые изменения структуры БД допускаются только после обновления данного документа и утверждения Architecture Review Agent.

---

# DATABASE PHILOSOPHY

База данных должна:

* быть простой;
* быть масштабируемой;
* хранить только необходимые данные;
* не содержать дублирования;
* поддерживать миллионы транзакций без изменения архитектуры.

---

# DATABASE ENGINE

Основная СУБД:

```
PostgreSQL
```

---

# DATABASE PRINCIPLES

Используются:

UUID Primary Keys

UTC Time

Soft Delete (где необходимо)

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

Запрещено хранить:

ИНН

паспорт

адрес

телефон

по умолчанию.

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

Все поля необязательные.

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

Хранится только агрегированное состояние.

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

Транзакции никогда не изменяются.

---

Редактирование создает новую версию.

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

Самая важная таблица системы.

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

Телефон работает только с Snapshot.

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

Тяжелые вычисления

не выполняются на клиенте.

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

не хранится постоянно.

---

Хранится только метаинформация.

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

Изменение запрещено.

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

Обновляются Worker.

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

Разрешено:

Users

Profiles

Notifications

---

Запрещено:

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

Только добавление.

---

# DATA RETENTION

Audit

10 лет

---

Transactions

без ограничения

---

Snapshots

без ограничения

---

Notifications

2 года

---

Logs

90 дней

---

# FOREIGN KEYS

Все FK обязательны.

---

Никаких orphan records.

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

DROP COLUMN напрямую запрещен.

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

Без изменения архитектуры могут быть добавлены:

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

Перед каждой миграцией Architecture Agent обязан проверить:

---

Есть ли дублирование?

---

Можно ли уменьшить количество таблиц?

---

Можно ли заменить JOIN Snapshot?

---

Можно ли заменить вычисление Cache?

---

Можно ли уменьшить индекс?

---

Можно ли избежать миграции?

---

# FINAL DATABASE PRINCIPLE

> **PostgreSQL является не просто хранилищем данных, а Canonical Financial Ledger OpenInvest.**

> **Все вычисления должны строиться вокруг неизменяемых транзакций, агрегированных Snapshot и математических моделей, а не вокруг постоянного пересчета истории.**

> **Правильно спроектированная база данных должна позволить обслуживать десятки миллионов финансовых операций без изменения своей фундаментальной структуры.**
