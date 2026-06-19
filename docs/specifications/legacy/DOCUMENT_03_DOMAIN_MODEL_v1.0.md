# DOCUMENT 03

# DOMAIN MODEL & BUSINESS LOGIC

Version: 1.0

Status: Source Of Truth

Priority: CRITICAL

---

# 1. DOMAIN PHILOSOPHY

OpenInvest не хранит "экраны".

OpenInvest хранит события.

Портфель является следствием событий.

Доходность является следствием событий.

Налог является следствием событий.

Каждая цифра должна быть воспроизводима.

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

не инвестирует.

User владеет Portfolio.

User может иметь несколько Portfolio.

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

Portfolio является Aggregate Root.

Все изменения происходят только через него.

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

Portfolio содержит:

Stocks

Bonds

Cash

Transactions

Snapshots

Watchlists

TaxDeclarations

---

# 5. ACCOUNT

В будущем пользователь может иметь несколько брокеров.

Поэтому появляется сущность Account.

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

Абстрактная сущность.

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

Наследники:

Stock

Bond

ETF

Fund

Currency

Crypto (future)

---

# 7. STOCK

Дополнительные поля

shares_outstanding

preferred

ordinary

dividend_policy

---

# 8. BOND

Дополнительные поля

nominal

coupon_rate

coupon_period

nkd

maturity

amortization

offer_date

---

# 9. TRANSACTION

Самая важная сущность проекта.

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

Типы

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

НИКОГДА

не изменять историю.

Использовать корректирующие операции.

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

Самая важная оптимизация продукта.

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

Frontend получает именно Snapshot.

Frontend никогда не пересчитывает историю.

---

# 14. EXCHANGE RATE

currency_from

currency_to

rate

source

datetime

---

Источник:

ЦБ РФ

---

# 15. INFLATION SNAPSHOT

date

official_rate

monthly_rate

yearly_rate

source

---

Источник

Росстат

или

официальные данные ЦБ

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

Самая важная сущность безопасности.

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

Никогда не удаляется.

---

# 20. BUSINESS RULES

Portfolio нельзя удалить,

если существуют неудаленные налоговые документы.

---

Transaction нельзя удалить физически.

Только Soft Delete.

---

Dividend нельзя редактировать вручную.

Только через подтвержденный источник.

---

Все математические расчеты являются детерминированными.

При одинаковых входных данных

результат всегда одинаковый.

---

# 21. REAL RETURN

Простая доходность запрещена.

---

Поддерживаем:

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

Уникальная функция продукта.

---

Показывать

Стоимость сегодня

↓

Стоимость в ценах прошлого года

↓

Потерю покупательной способности

↓

Эквиваленты

iPhone

MacBook

Средняя зарплата

ЖКХ

Продуктовая корзина

Автомобиль

Квадратные метры жилья

---

# 23. EVENT MODEL

Любое действие —

это событие.

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

Portfolio никогда не может иметь отрицательное количество активов.

Cash не может быть меньше допустимого значения без подтверждения пользователя.

Snapshot всегда строится на официальной цене закрытия.

XML декларация всегда привязана к версии налогового движка.

---

# 26. SELF CRITICISM

Риски:

XIRR является тяжелой операцией.

При большом количестве пользователей потребуется кеширование результатов.

---

Inflation Engine должен иметь возможность смены источника.

---

Tax Engine должен поддерживать версии законодательства.

---

Snapshot Engine должен уметь пересчитывать историю после исправления сделки.

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
