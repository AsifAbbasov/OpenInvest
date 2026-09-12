# DOCUMENT 03

# DOMAIN MODEL & BUSINESS LOGIC

Version: 1.0

Status: Source Of Truth

Priority: CRITICAL

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_03_DOMAIN_MODEL_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `f10bab0e6958599690e9002cb41356a3e66f1173`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# 1. DOMAIN PHILOSOPHY

OpenInvest does not store "screens."

OpenInvest stores events.

A portfolio is a consequence of events.

Return is a consequence of events.

Tax is a consequence of events.

Every number must be reproducible.

---

# 2. DOMAIN MODEL

ROOT

OpenInvest

│

├── User

├── Portfolio

├── Account

├── Asset

├── Stock

├── Bond

├── Transaction

├── Dividend

├── Coupon

├── Currency

├── ExchangeRate

├── InflationSnapshot

├── MarketSnapshot

├── TaxDeclaration

├── Notification

├── Watchlist

├── AuditLog

└── Settings

---

# 3. USER

User

does not invest.

User owns Portfolio.

User may have several Portfolios.

---

User

id

email

password_hash

created_at

updated_at

is_premium

language

timezone

currency

theme

2fa_enabled

deleted_at

---

# 4. PORTFOLIO

Portfolio is an Aggregate Root.

All changes happen only through it.

---

Portfolio

id

user_id

name

base_currency

created_at

updated_at

visibility

status

---

Portfolio contains:

Stocks

Bonds

Cash

Transactions

Snapshots

Watchlists

TaxDeclarations

---

# 5. ACCOUNT

In the future, a user may have several brokers.

Therefore the Account entity appears.

---

Account

id

portfolio_id

broker_name

broker_type

currency

created_at

status

---

Portfolio

↓

Accounts

↓

Transactions

---

# 6. ASSET

Abstract entity.

---

Asset

id

ticker

isin

figi

name

currency

country

sector

exchange

asset_type

status

---

Descendants:

Stock

Bond

ETF

Fund

Currency

Crypto (future)

---

# 7. STOCK

Additional fields

shares_outstanding

preferred

ordinary

dividend_policy

---

# 8. BOND

Additional fields

nominal

coupon_rate

coupon_period

nkd

maturity

amortization

offer_date

---

# 9. TRANSACTION

The most important entity in the project.

---

Transaction

id

portfolio_id

account_id

asset_id

type

quantity

price

commission

currency

exchange_rate

nkd

trade_datetime_utc

settlement_datetime

comment

source

created_by

created_at

updated_at

---

Types

BUY

SELL

DIVIDEND

COUPON

COMMISSION

TAX

TRANSFER

DEPOSIT

WITHDRAW

CORRECTION

---

NEVER

change history.

Use corrective operations.

---

# 10. DIVIDEND

Dividend

id

asset_id

amount

currency

registry_date

payment_date

announcement_date

status

official_source

yield

---

status

OFFICIAL

RECOMMENDED

EXPECTED

CANCELLED

---

# 11. COUPON

Coupon

id

bond_id

amount

payment_date

currency

tax

---

# 12. MARKET SNAPSHOT

MarketSnapshot

ticker

price

open

close

high

low

volume

datetime

source

---

# 13. PORTFOLIO SNAPSHOT

The most important product optimization.

---

PortfolioSnapshot

portfolio_id

datetime

market_value

cash_value

dividend_value

coupon_value

nominal_profit

real_profit

inflation_adjusted_value

xirr

twr

cagr

---

Frontend receives the Snapshot itself.

Frontend never recalculates history.

---

# 14. EXCHANGE RATE

currency_from

currency_to

rate

source

datetime

---

Source:

CBR

---

# 15. INFLATION SNAPSHOT

date

official_rate

monthly_rate

yearly_rate

source

---

Source

Rosstat

or

official CBR data

---

# 16. TAX DECLARATION

TaxDeclaration

id

user_id

year

status

generated_xml

generated_pdf

generated_zip

created_at

expires_at

---

Status

Draft

Ready

Confirmed

Deleted

Expired

---

# 17. NOTIFICATION

Notification

id

user_id

type

priority

title

body

read

created_at

---

Type

Dividend

Coupon

Tax

Market

Inflation

Portfolio

System

---

# 18. WATCHLIST

id

user_id

ticker

created_at

---

# 19. AUDIT LOG

The most important security entity.

---

AuditLog

id

user_id

event

entity

entity_id

old_value

new_value

ip

device

datetime

version

---

Never deleted.

---

# 20. BUSINESS RULES

Portfolio cannot be deleted

if undeleted tax documents exist.

---

Transaction cannot be physically deleted.

Soft Delete only.

---

Dividend cannot be edited manually.

Only through a confirmed source.

---

All mathematical calculations are deterministic.

Given identical input data,

the result is always identical.

---

# 21. REAL RETURN

Simple return is prohibited.

---

Supported:

Nominal Return

Dividend Return

Coupon Return

Total Return

Tax Adjusted Return

Inflation Adjusted Return

Money Weighted Return (XIRR)

Time Weighted Return

CAGR

---

# 22. INFLATION RETURN

A unique product function.

---

Show

Value today

↓

Value in last year's prices

↓

Loss of purchasing power

↓

Equivalents

iPhone

MacBook

Average salary

Utilities

Food basket

Car

Square meters of housing

---

# 23. EVENT MODEL

Any action

is an event.

BUY

↓

TransactionCreated

↓

PortfolioUpdated

↓

SnapshotQueued

↓

TaxQueued

↓

NotificationQueued

↓

AnalyticsUpdated

---

# 24. DOMAIN EVENTS

TransactionCreated

TransactionUpdated

TransactionDeleted

DividendDeclared

DividendChanged

CouponPaid

PortfolioCalculated

SnapshotCreated

TaxGenerated

EmailSent

NotificationDelivered

InflationUpdated

MarketUpdated

---

# 25. DOMAIN INVARIANTS

Portfolio can never have a negative quantity of assets.

Cash cannot be below the permitted value without user confirmation.

Snapshot is always built using the official closing price.

XML declaration is always tied to a tax-engine version.

---

# 26. SELF CRITICISM

Risks:

XIRR is a heavy operation.

With a large number of users, result caching will be required.

---

Inflation Engine must support changing its source.

---

Tax Engine must support versions of legislation.

---

Snapshot Engine must be able to recalculate history after a transaction correction.

---

# 27. FUTURE

Multi Broker

Multi Country

Multi Tax System

AI Advisor

Broker Import

Public API

Plugin System

Institutional Accounts

END OF DOCUMENT
