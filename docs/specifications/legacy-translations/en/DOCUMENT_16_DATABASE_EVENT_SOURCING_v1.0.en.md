# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 16

# DATABASE ARCHITECTURE, DATA MODEL, EVENT SOURCING & DATA LIFECYCLE

Version: 1.0

Status: Approved

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_16_DATABASE_EVENT_SOURCING_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `f21abf30ba7ebacb5221d796964358a6e7be2711`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# PURPOSE

This document defines the only permitted OpenInvest data-storage model.

All project data must be stored in such a way that:

* history is never lost;
* any portfolio state can be restored;
* auditing is possible;
* any analytics can be built;
* scaling to tens of millions of records is supported.

---

# DATABASE PHILOSOPHY

OpenInvest never stores "current state" as the only source of truth.

The source of truth is events.

Any change must be reproducible.

---

# PRINCIPLES

Immutable Data

Audit First

Privacy by Design

Soft Delete

Append Only

Event Driven

---

# DATABASE

PostgreSQL

UTF8

Timezone UTC

UUID Primary Keys

Numeric instead of Float

---

# MAIN DATABASE MODULES

```text
users

portfolios

assets

transactions

positions

snapshots

dividends

coupons

currencies

inflation

notifications

audit

exports

settings

sessions

devices

logs
```

---

# USERS

Purpose

Account storage.

---

Fields

id UUID

email

password_hash

premium_status

created_at

updated_at

deleted_at

---

Passport data is absent.

---

# USER_PROFILE

A separate table.

Stores only:

first name

last name

phone

country

timezone

language

---

Never physically combined with investment tables.

---

# PERSONAL_DATA

A separate encrypted schema.

Contains:

INN

passport

address

registration

date of birth

---

AES-256 Encryption

---

All fields are optional.

---

# USER_PRIVACY_SETTINGS

Each user chooses:

☐ store passport

☐ store INN

☐ store address

☐ automatically fill declaration

☐ delete after generation

---

# PORTFOLIOS

One user

may have

an unlimited number of portfolios.

---

Examples

Retirement

Dividends

Child

USD

ETF

---

# PORTFOLIO TABLE

id

user_id

name

currency

visibility

created_at

updated_at

---

# ASSETS DIRECTORY

Independent of the user.

---

Fields

ticker

isin

figi

name

sector

industry

country

currency

exchange

asset_type

lot_size

nominal

coupon_rate

maturity

---

# ASSET TYPES

Stock

Bond

ETF

REIT

Currency

Cash

Gold

Future (future release)

Option (future release)

---

# TRANSACTIONS

This is the most important table in the system.

Never physically deleted.

---

Fields

id

portfolio_id

asset_id

operation_type

quantity

price

commission

tax

nkd

currency

trade_datetime_utc

broker

comment

created_at

---

# OPERATION TYPES

BUY

SELL

DIVIDEND

COUPON

DEPOSIT

WITHDRAW

TRANSFER

CORRECTION

---

# EVENT SOURCING

Each operation is

a separate event.

Never changed.

---

When corrected,

a new event is created.

---

# SOFT DELETE

Deleting a transaction

=

creating the event

VOID

---

History is fully preserved.

---

# POSITIONS

Materialized View

or

an updated table.

---

Stores

only current positions.

---

Fields

portfolio_id

asset_id

quantity

average_cost

market_value

profit

profit_percent

updated_at

---

# AVERAGE COST

Recalculated

only after BUY.

---

SELL

does not change Average Cost.

---

# SNAPSHOTS

Created automatically.

---

Fields

portfolio_id

snapshot_date

market_value

cash

profit

xirr

inflation_adjusted_value

---

# SNAPSHOT FREQUENCY

Daily

mandatory.

---

Hourly

for Premium.

---

# DIVIDENDS DIRECTORY

Independent of the user.

---

ticker

amount

currency

registry_date

payment_date

status

official_source

updated_at

---

# STATUS

Official

Forecast

Canceled

Corrected

---

# COUPON DIRECTORY

For bonds.

---

ticker

coupon

payment_date

record_date

nkd

yield

---

# CURRENCY DIRECTORY

Stores:

USD

EUR

CNY

KZT

and others.

---

# HISTORICAL FX

A separate table.

---

date

currency

rate_cbr

source

---

Used

for taxes.

---

# INFLATION DIRECTORY

---

year

month

official_rate

cumulative_index

source

---

Used

for real portfolio value.

---

# REAL VALUE

The system stores:

Nominal Value

Real Value

Inflation Loss

Purchasing Power Index

---

# PURCHASING POWER TABLE

Examples of equivalents:

MacBook

iPhone

Average salary

Food basket

Utilities

Gasoline

Gold

---

This is a separate directory.

---

# NOTIFICATIONS

id

user

type

title

body

status

created_at

read_at

---

# AUDIT LOG

Absolutely all actions.

---

Login

Logout

Buy

Sell

Export

Email

Tax

Profile Update

Delete

---

# AUDIT FIELDS

User

Device

IP

Country

Browser

Action

Timestamp UTC

---

# EXPORTS

All generated documents.

---

Fields

id

user

type

created

expires

status

---

The file itself

is not stored.

---

Only a record is stored.

---

# FILE STORAGE

XML

PDF

ZIP

are created

in RAM.

---

After sending:

they are deleted.

---

# DEVICES

Trusted Devices.

---

device_id

os

browser

last_login

fingerprint

---

# SESSIONS

JWT

Refresh

Expires

Device

---

# INDEXES

user_id

portfolio_id

ticker

isin

trade_date

snapshot_date

---

Composite

portfolio_id + trade_date

ticker + payment_date

---

# PARTITIONING

Transactions

by years.

---

Snapshots

by months.

---

Audit

by quarters.

---

# MATERIALIZED VIEWS

Portfolio Summary

Dividend Summary

Sector Allocation

Annual Tax Summary

---

# RETENTION POLICY

Audit

10 years

---

Transactions

are never deleted

(unless the user requests complete deletion).

---

# PRIVACY MODE

If the user chooses

"Do not store personal data":

passport

INN

address

are not written to the database.

---

After declaration generation,

the data exists

only in process RAM

and is completely destroyed.

---

# EXPORT PROFILE

The user can download:

JSON

CSV

PDF

XML

ZIP

---

# DELETE PROFILE

One button:

Delete profile.

---

The system:

invalidates sessions

deletes personal data

clears encrypted schema

deletes export records

deletes notifications

anonymizes investment events.

---

# TARGET SCALE

10 000 users

1 000 000 transactions

100 000 000 snapshots

50 000 000 audit events

without architecture changes.

---

# CODEX REQUIREMENTS

Before creating any new table, check:

1. Does it duplicate existing data?

2. Can a directory be used?

3. Can a materialized view be used?

4. Can a snapshot be used instead of recalculation?

5. Does the structure comply with Third Normal Form?

6. Does it violate Privacy by Design?

7. Does it store personal data without explicit user consent?

Only after passing these checks is the database structure considered approved.
