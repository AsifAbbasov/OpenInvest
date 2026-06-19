# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 16

# DATABASE ARCHITECTURE, DATA MODEL, EVENT SOURCING & DATA LIFECYCLE

Version: 1.0

Status: Approved

Priority: Critical

---

# PURPOSE

Данный документ определяет единственную допустимую модель хранения данных OpenInvest.

Все данные проекта должны храниться таким образом, чтобы:

* никогда не терялась история;
* можно было восстановить любое состояние портфеля;
* была возможность проводить аудит;
* была возможность построить любую аналитику;
* поддерживалось масштабирование на десятки миллионов записей.

---

# DATABASE PHILOSOPHY

OpenInvest никогда не хранит "текущее состояние" как единственный источник истины.

Источник истины — события.

Любое изменение должно быть воспроизводимо.

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

Numeric вместо Float

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

Назначение

Хранение учетной записи.

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

Паспортные данные отсутствуют.

---

# USER_PROFILE

Отдельная таблица.

Хранит только:

имя

фамилия

телефон

страна

часовой пояс

язык

---

Никогда не объединяется физически с инвестиционными таблицами.

---

# PERSONAL_DATA

Отдельная encrypted schema.

Содержит:

ИНН

паспорт

адрес

регистрация

дата рождения

---

AES-256 Encryption

---

Все поля необязательные.

---

# USER_PRIVACY_SETTINGS

Каждый пользователь выбирает:

☐ хранить паспорт

☐ хранить ИНН

☐ хранить адрес

☐ автоматически заполнять декларацию

☐ удалить после генерации

---

# PORTFOLIOS

Один пользователь

может иметь

неограниченное количество портфелей.

---

Examples

Пенсия

Дивиденды

Детский

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

Не зависит от пользователя.

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

Это самая важная таблица системы.

Никогда не удаляется физически.

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

Каждая операция —

отдельное событие.

Никогда не изменяется.

---

При исправлении создается

новое событие.

---

# SOFT DELETE

Удаление сделки

=

создание события

VOID

---

История полностью сохраняется.

---

# POSITIONS

Materialized View

или

обновляемая таблица.

---

Хранит

только актуальные позиции.

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

Пересчитывается

только после BUY.

---

SELL

не изменяет Average Cost.

---

# SNAPSHOTS

Создаются автоматически.

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

обязательно.

---

Hourly

для Premium.

---

# DIVIDENDS DIRECTORY

Не зависит от пользователя.

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

Для облигаций.

---

ticker

coupon

payment_date

record_date

nkd

yield

---

# CURRENCY DIRECTORY

Хранит:

USD

EUR

CNY

KZT

и др.

---

# HISTORICAL FX

Отдельная таблица.

---

date

currency

rate_cbr

source

---

Используется

для налогов.

---

# INFLATION DIRECTORY

---

year

month

official_rate

cumulative_index

source

---

Используется

для реальной стоимости портфеля.

---

# REAL VALUE

Система хранит:

Nominal Value

Real Value

Inflation Loss

Purchasing Power Index

---

# PURCHASING POWER TABLE

Примеры эквивалентов:

MacBook

iPhone

Средняя зарплата

Продуктовая корзина

ЖКХ

Бензин

Золото

---

Это отдельный справочник.

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

Абсолютно все действия.

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

Все сформированные документы.

---

Fields

id

user

type

created

expires

status

---

Сам файл

не хранится.

---

Хранится только запись.

---

# FILE STORAGE

XML

PDF

ZIP

создаются

в RAM.

---

После отправки:

удаляются.

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

по годам.

---

Snapshots

по месяцам.

---

Audit

по кварталам.

---

# MATERIALIZED VIEWS

Portfolio Summary

Dividend Summary

Sector Allocation

Annual Tax Summary

---

# RETENTION POLICY

Audit

10 лет

---

Transactions

никогда не удаляются

(если пользователь не потребует полного удаления).

---

# PRIVACY MODE

Если пользователь выбрал

"Не сохранять личные данные":

паспорт

ИНН

адрес

не записываются в БД.

---

После генерации декларации

данные существуют

только в RAM процесса

и полностью уничтожаются.

---

# EXPORT PROFILE

Пользователь может скачать:

JSON

CSV

PDF

XML

ZIP

---

# DELETE PROFILE

Одна кнопка:

Удалить профиль.

---

Система:

аннулирует сессии

удаляет персональные данные

очищает encrypted schema

удаляет экспортные записи

удаляет уведомления

анонимизирует инвестиционные события.

---

# TARGET SCALE

10 000 пользователей

1 000 000 транзакций

100 000 000 snapshots

50 000 000 audit событий

без изменения архитектуры.

---

# CODEX REQUIREMENTS

Перед созданием любой новой таблицы необходимо проверить:

1. Не дублирует ли она существующие данные.

2. Можно ли использовать справочник.

3. Можно ли использовать materialized view.

4. Можно ли использовать snapshot вместо перерасчета.

5. Соответствует ли структура Third Normal Form.

6. Не нарушает ли Privacy by Design.

7. Не хранит ли персональные данные без явного согласия пользователя.

Только после прохождения этих проверок структура БД считается утвержденной.
