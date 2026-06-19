# DOCUMENT 05

# DATABASE ARCHITECTURE & POSTGRESQL DESIGN

Version 1.0

Status: SOURCE OF TRUTH

Priority: CRITICAL

---

# 1. DATABASE PHILOSOPHY

Database is the heart of OpenInvest.

Every financial operation must be:

consistent

reproducible

auditable

recoverable

immutable

---

# 2. DATABASE PRINCIPLES

Single Source of Truth

Append Only Financial History

No Hidden Calculations

No Silent Updates

No Business Logic Inside UI

Every Change Is Logged

Every Important Event Is Traceable

---

# 3. DATABASES

Identity

Investment

Analytics

Tax

Audit

Notification

System

---

# 4. CORE TABLES

users

user_settings

user_devices

user_sessions

user_notifications

user_consents

user_security_events

---

assets

asset_prices

asset_dividends

asset_coupons

asset_events

asset_sectors

asset_currency

asset_exchange

---

transactions

transaction_events

transaction_files

transaction_imports

---

portfolio_positions

portfolio_snapshots

portfolio_statistics

portfolio_allocations

portfolio_cashflows

---

tax_profiles

tax_exports

tax_currency_rates

tax_reports

tax_audit

---

notifications

notification_queue

notification_history

notification_templates

---

audit_log

system_log

security_log

parser_log

scheduler_log

---

# 5. USERS

users

contains only

UUID

email

password_hash

premium_status

created_at

updated_at

deleted_at

No passport.

No address.

No INN.

No SNILS.

No personal documents.

---

# 6. OPTIONAL PERSONAL DATA

tax_profile

contains

full_name

passport

INN

address

registration

birth_date

phone

email

Every field optional.

Every field encrypted.

Every field removable.

---

# 7. PRIVACY

Investment data stored separately.

Identity stored separately.

Tax data stored separately.

No single table contains complete user profile.

---

# 8. TRANSACTIONS

Every operation is immutable.

BUY

SELL

DIVIDEND

COUPON

COMMISSION

TAX

TRANSFER

CASH_DEPOSIT

CASH_WITHDRAW

FX

---

Transactions never updated.

Corrections create new events.

---

# 9. EVENT SOURCING

Original transaction always preserved.

Correction references previous transaction.

Complete history always available.

---

# 10. PORTFOLIO

portfolio_positions

contains only current state.

Never edited manually.

Always rebuilt from transactions.

---

# 11. SNAPSHOTS

portfolio_snapshots

contains

date

market_value

cash

bonds

stocks

expected_dividends

xirr

inflation_adjusted_value

currency

---

Snapshots immutable.

---

# 12. PRICE STORAGE

asset_prices

ticker

date

open

high

low

close

volume

source

loaded_at

---

# 13. DIVIDENDS

asset_dividends

ticker

announcement_date

registry_date

payment_date

amount

currency

status

source

confidence

---

# 14. CURRENCY

currency_rates

date

base_currency

target_currency

rate

source

official_flag

---

Only official providers.

---

# 15. INFLATION

inflation_rates

country

period

official_rate

annual_rate

source

loaded_at

---

# 16. PURCHASING POWER

purchasing_power_reference

MacBook

iPhone

Average Salary

Food Basket

Utilities

Fuel

Mortgage

Rent

Consumer Basket

---

Historical values stored.

---

# 17. INFLATION ENGINE

Portfolio value

↓

Inflation correction

↓

Purchasing power

↓

Equivalent goods

↓

Visualization

---

# 18. TAX ENGINE

Tax profile

↓

Transactions

↓

Currency Rates

↓

Dividends

↓

Calculation

↓

Audit

↓

XML

↓

PDF

---

# 19. TAX PRIVACY

User chooses

Mode A

Store profile

Generate automatically

---

Mode B

Temporary profile

Delete after generation

---

Mode C

Empty declaration

User fills manually

---

Default = Mode B

Privacy First.

---

# 20. CACHE TABLES

cache_market

cache_dividends

cache_currency

cache_inflation

cache_statistics

---

TTL configurable.

---

# 21. AUDIT

Every action recorded.

login

logout

create transaction

delete transaction

export pdf

export xml

email

settings

profile

tax generation

---

# 22. SECURITY

Every login

IP

Device

Country

Browser

OS

Time

Success

Failure

---

# 23. IMPORTS

CSV

Excel

Broker Reports

Manual

API

---

Original file stored.

Parsing result stored.

Errors stored.

---

# 24. EXPORTS

PDF

XML

CSV

Excel

JSON

ZIP

Email

---

Every export logged.

---

# 25. NOTIFICATIONS

Dividend

Coupon

Tax

Portfolio

Price

Reminder

Security

System

---

# 26. PARSER

Every parser result

stored

validated

versioned

audited

---

# 27. VERSIONING

Every schema

Every parser

Every tax engine

Every export format

Every API

Version controlled.

---

# 28. PARTITIONING

Large tables partitioned

by year

transactions

prices

audit

notifications

snapshots

---

# 29. INDEXES

Primary

Unique

Composite

Partial

GIN

BRIN

Only where benchmark proves benefit.

---

# 30. SOFT DELETE

Users

yes

Settings

yes

Notifications

yes

Transactions

NO

Financial history immutable.

---

# 31. BACKUP

Hourly WAL

Daily Incremental

Weekly Full

Monthly Archive

Quarterly Validation Restore

---

# 32. DISASTER RECOVERY

Point In Time Recovery

Automatic

Verified

Documented

---

# 33. ENCRYPTION

Passwords

Argon2id

PII

AES-256

Transport

TLS 1.3

Secrets

Environment Vault

---

# 34. GDPR / PRIVACY

Export profile

One click

Delete profile

One click

Delete tax profile

One click

Delete sessions

One click

Investment history preserved only if legally required.

---

# 35. PERFORMANCE TARGET

10 000 users

single database

100 000 users

partitioning

500 000 users

read replicas

1 000 000 users

horizontal scaling

---

# 36. FUTURE READY

MOEX

SPB

NYSE

NASDAQ

LSE

ETF

Crypto

Funds

Pension

Family Portfolio

Corporate Portfolio

Advisor Mode

API Marketplace

---

# 37. DESIGN DECISION

Database optimized for

Correctness

Trust

Auditability

Privacy

Scalability

Maintainability

instead of premature micro-optimizations.

END OF DOCUMENT
