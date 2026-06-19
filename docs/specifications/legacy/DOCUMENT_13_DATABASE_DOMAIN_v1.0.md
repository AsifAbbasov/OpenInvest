# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 13

# DATABASE, DOMAIN MODEL & BUSINESS LOGIC BLUEPRINT

Version: 1.0

Status: Core Domain

Priority: Critical

---

# 1. PURPOSE

Настоящий документ определяет:

* структуру базы данных;
* бизнес-сущности;
* связи между ними;
* правила хранения;
* жизненный цикл объектов;
* математическую модель хранения инвестиционного портфеля.

Данный документ является главным источником истины для Backend, Frontend, Android, iOS и AI Agents.

---

# 2. DOMAIN PHILOSOPHY

OpenInvest не хранит "страницы".

OpenInvest хранит бизнес-сущности.

Каждая сущность должна иметь:

* четкую ответственность;
* независимый жизненный цикл;
* собственную историю;
* возможность расширения.

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

User — владелец всех данных.

User никогда не хранит инвестиционные данные напрямую.

User является контейнером.

---

Поля:

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

Отдельная сущность.

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

Отдельная таблица.

Никаких связей с Portfolio кроме user_id.

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

Пользователь может иметь несколько портфелей.

Примеры:

```
Основной

Дивиденды

Пенсия

Ребенок

ETF

Облигации

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

Transaction является единственным источником истины.

Любое изменение портфеля происходит только через новую транзакцию.

---

Типы:

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

Исторические транзакции запрещено изменять напрямую.

Редактирование:

создает новую ревизию.

или

создает компенсирующую операцию.

---

# 11. HOLDINGS

Holdings — агрегированное состояние.

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

Пересчитывается автоматически.

---

# 12. SNAPSHOTS

Хранят ежедневое состояние.

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

Деньги являются отдельным активом.

```
RUB

USD

EUR

CNY
```

---

Пользователь может учитывать свободные деньги.

---

# 14. ASSET TYPES

Поддерживаются:

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

Архитектура должна позволять добавить новый тип без переписывания системы.

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

Не зависит от пользователя.

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

Для облигаций.

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

Отдельная сущность.

Не зависит от портфеля.

---

Можно хранить:

100+

бумаг

без покупки.

---

# 19. NOTIFICATIONS

Типы:

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

Хранятся отдельно.

```
recommendation

source

confidence

created_at

status
```

---

ИИ никогда не изменяет портфель.

---

# 21. AUDIT

Любое действие:

```
User

↓

Action

↓

Audit
```

---

# 22. VERSIONING

Каждая запись имеет:

```
created_at

updated_at

version
```

---

# 23. SOFT DELETE

Удаление:

```
deleted_at

deleted_by

reason
```

---

Физическое удаление только Background Worker.

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

Все справочники:

3NF.

---

История:

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

по годам

---

audit_logs

по месяцам

---

snapshots

по годам

---

notifications

по кварталам

---

# 28. MATERIALIZED VIEWS

Создать:

```
portfolio_summary

dividend_summary

sector_summary

tax_summary
```

---

Это значительно уменьшает нагрузку.

---

# 29. BUSINESS RULES

BUY:

увеличивает количество

пересчитывает Average Cost

---

SELL:

уменьшает количество

не меняет Average Cost

---

DIVIDEND:

увеличивает Cash

увеличивает Income

---

COUPON:

увеличивает Cash

---

COMMISSION:

уменьшает Cash

уменьшает Income

---

# 30. INFLATION MODEL

Кроме номинальной стоимости хранится:

```
Real Value

Inflation Loss

Purchasing Power

Real Return
```

---

# 31. PURCHASING POWER

Дополнительный показатель.

Примеры:

```
Ваш портфель:

2.3 MacBook Pro

5 iPhone

17 средних зарплат

9 месяцев аренды квартиры

31 продуктовая корзина
```

---

Не используется в математике.

Используется для визуализации.

---

# 32. HISTORICAL CONSISTENCY

Запрещено:

пересчитывать историю по сегодняшним данным.

---

Используются только:

исторические цены;

исторические дивиденды;

исторические курсы ЦБ;

историческая инфляция.

---

# 33. MULTI CURRENCY

Все активы имеют:

```
Original Currency

Base Currency

Converted Currency
```

---

Конвертация выполняется по историческому курсу.

---

# 34. FREE VS PREMIUM

FREE

до 5 компаний

1 портфель

базовая аналитика

---

PREMIUM

без ограничений

несколько портфелей

AI аналитика

налоговый помощник

экспорт

расширенная статистика

---

# 35. EXTENSIBILITY

Добавление нового актива не должно требовать изменения:

Portfolio

Transactions

Snapshots

Charts

Statistics

---

Достаточно добавить новый AssetType.

---

# 36. FUTURE MODULES

Архитектура заранее предусматривает:

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

Через 5 лет разработки база данных должна поддерживать:

* миллионы транзакций;
* сотни тысяч пользователей;
* десятки миллионов snapshot-записей;
* многовалютные портфели;
* несколько налоговых юрисдикций;

без изменения основной модели данных.

---

# 38. CODEX REQUIREMENT

Codex обязан реализовывать новые функции только через Domain Model.

Запрещено:

создавать "быстрые костыли";

дублировать данные;

хранить бизнес-логику во Frontend;

считать математику в React;

нарушать Single Source of Truth.

Этот документ является обязательным для Builder Agent, Review Agent и QA Agent.
