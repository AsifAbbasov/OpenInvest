Отлично.

Теперь мы переходим к самому важному документу проекта.

---

# DOCUMENT 06

# API CONTRACT & BACKEND ARCHITECTURE

**Version:** 1.0

**Status:** Source of Truth

**Priority:** Critical

---

# 1. ФИЛОСОФИЯ BACKEND

Backend является **единственным источником истины**.

Frontend (React/Web, Swift/iOS, Kotlin/Android) не содержит бизнес-логики.

Любой клиент является просто отображением данных.

```
Client

↓

API Gateway

↓

Go Fiber

↓

Service Layer

↓

Domain Layer

↓

Repository Layer

↓

PostgreSQL / Cache

↓

Official Data Providers
```

---

# 2. ОСНОВНЫЕ ПРИНЦИПЫ

API должно быть

Stateless

Idempotent

Versioned

Observable

Cacheable

Scalable

Auditable

---

# 3. СТРУКТУРА API

```
/api

/api/v1

/api/v2

```

v1 никогда не ломается.

Новые возможности появляются только в новой версии.

---

# 4. МОДУЛИ API

```
Auth

Users

Portfolio

Transactions

Assets

Prices

Dividends

Coupons

Statistics

Tax

Notifications

Search

Admin

Health

Metrics

```

---

# 5. AUTH

```
POST /auth/register

POST /auth/login

POST /auth/logout

POST /auth/refresh

POST /auth/change-password

POST /auth/reset-password

GET /auth/me
```

---

# 6. USERS

```
GET /users/profile

PUT /users/profile

DELETE /users/profile

GET /users/settings

PUT /users/settings
```

---

# 7. PORTFOLIO

```
GET /portfolio

GET /portfolio/summary

GET /portfolio/history

GET /portfolio/allocation

GET /portfolio/dividends

GET /portfolio/inflation

GET /portfolio/xirr
```

---

# 8. TRANSACTIONS

```
GET /transactions

POST /transactions

PUT /transactions/{id}

DELETE /transactions/{id}

GET /transactions/history
```

---

# 9. ASSETS

```
GET /assets

GET /assets/search

GET /assets/{ticker}

GET /assets/{ticker}/history

GET /assets/{ticker}/dividends

GET /assets/{ticker}/coupons

GET /assets/{ticker}/statistics
```

---

# 10. DIVIDENDS

```
GET /dividends/calendar

GET /dividends/upcoming

GET /dividends/history

GET /dividends/official

GET /dividends/prognosis
```

---

# 11. TAX

```
GET /tax/profile

PUT /tax/profile

POST /tax/generate/xml

POST /tax/generate/pdf

POST /tax/send/email

POST /tax/export/zip
```

---

# 12. NOTIFICATIONS

```
GET /notifications

PUT /notifications/read

PUT /notifications/settings

DELETE /notifications/{id}
```

---

# 13. SEARCH

```
GET /search

GET /search/ticker

GET /search/company
```

---

# 14. HEALTH

```
GET /health

GET /metrics

GET /version

GET /status
```

---

# 15. ЕДИНЫЙ ФОРМАТ ОТВЕТОВ

```
{
    success,
    data,
    errors,
    meta
}
```

---

# 16. PAGINATION

Cursor Pagination.

Не Offset.

Причина:

нам нужны сотни тысяч записей.

Offset будет деградировать.

---

# 17. SORTING

Любой список обязан поддерживать

```
sort

order

limit

cursor
```

---

# 18. FILTERS

Каталог

```
sector

price

yield

marketCap

currency

assetType

```

---

# 19. CACHE STRATEGY

Не каждый запрос идет на MOEX.

```
Client

↓

CDN

↓

API Cache

↓

RAM Cache

↓

Database

↓

Official API
```

---

# 20. RAM CACHE

Go

```
sync.Map
```

или

Redis

---

# 21. CACHE TTL

market data

5 минут

---

dividends

12 часов

---

statistics

24 часа

---

inflation

1 месяц

---

currency

24 часа

---

# 22. RATE LIMIT

Очень важный раздел.

---

Никогда

никогда

никогда

Frontend не обращается напрямую к MOEX.

---

Все обращения:

```
React

↓

Go

↓

Cache

↓

MOEX
```

---

# 23. ЗАЩИТА ОТ БЛОКИРОВКИ

Если

10000 пользователей

открыли приложение одновременно

они НЕ создают

10000 запросов.

---

Они получают

ОДИН

уже закэшированный ответ.

---

# 24. UPDATE STRATEGY

Во время торгов

каждые 5 минут

---

После закрытия торгов

один финальный refresh

---

Ночью

никаких запросов.

---

Выходные

никаких запросов.

---

Праздники

никаких запросов.

---

# 25. ОФИЦИАЛЬНЫЕ ИСТОЧНИКИ

Только

официальные

или

свободно доступные.

---

MOEX ISS

ЦБ РФ

Росстат

официальные сайты эмитентов

e-disclosure

---

# 26. НИКАКОГО SCRAPING

Если существует официальный API

используется только он.

---

Парсер применяется

только

если официального API нет.

---

# 27. EMAIL

Асинхронная очередь.

Пользователь не ждет SMTP.

```
Request

↓

Queue

↓

Worker

↓

SMTP

↓

Success

↓

Notification
```

---

# 28. ZIP

XML

PDF

CSV

создаются

исключительно

в оперативной памяти.

```
bytes.Buffer

↓

archive/zip

↓

SMTP

```

---

никаких временных файлов.

---

# 29. ИНФЛЯЦИЯ

API

```
GET /portfolio/inflation
```

возвращает

```
nominalValue

realValue

inflationPercent

purchasingPower

examples
```

---

# 30. ПОКУПАТЕЛЬСКАЯ СПОСОБНОСТЬ

Пользователь видит

не

```
200000 рублей
```

а

```
200000 рублей

=

1 MacBook Pro

или

5 iPhone

или

8 месяцев коммунальных услуг

или

4 средние зарплаты региона

```

---

Это одна из наших уникальных функций.

---

# 31. XIRR

Никогда

не считается

во Frontend.

---

Всегда

Backend.

---

# 32. SNAPSHOTS

Frontend получает

готовый массив.

```
date

value

cash

stocks

bonds

realValue

```

никаких перерасчетов.

---

# 33. ИДЕМПОТЕНТНОСТЬ

Любая финансовая операция

имеет

Idempotency-Key.

Повторная отправка

не создаст дубль сделки.

---

# 34. OBSERVABILITY

Каждый endpoint

имеет

Latency

Errors

Success Rate

Memory

CPU

Cache Hit

---

# 35. SECURITY

JWT

Refresh Token

Rotation

CSRF

Rate Limit

Brute Force Protection

Device Tracking

IP Monitoring

---

# 36. TRUST & PRIVACY

По умолчанию

паспорт

ИНН

адрес

телефон

НЕ требуются.

---

Налоговый профиль

создается

по желанию.

---

# 37. УДАЛЕНИЕ ДАННЫХ

Пользователь может

скачать

XML

CSV

JSON

PDF

после чего

удалить

весь профиль

одной кнопкой.

---

# 38. ЧТО НУЖНО ДОБАВИТЬ (МОЯ КРИТИКА)

Вот что, на мой взгляд, обязательно должно появиться в следующих документах:

## 1. Import Center

Импорт:

* БКС
* Альфа
* Т-Инвестиции
* ВТБ
* Финам
* Сбер
* Interactive Brokers
* CSV
* Excel

Одной кнопкой.

---

## 2. Compare Portfolio

Не только:

"сколько заработал"

а

"что было бы, если вместо Газпрома был Лукойл"

---

## 3. Dividend Simulator

"Если я докуплю еще 100 акций сегодня"

↓

покажи:

* дивиденды
* XIRR
* налог
* инфляционную стоимость
* изменение покупательной способности

---

## 4. Life Goals

Самая сильная идея продукта.

Не просто

```
Портфель = 5 700 000 ₽
```

а

```
Ваш портфель обеспечивает:

✓ 14 лет оплаты коммунальных услуг

✓ 8 лет продуктовой корзины

✓ 3 MacBook Pro ежегодно только на дивиденды

✓ 27% пути до финансовой независимости
```

---

### Моя оценка

Документ уже достаточно силен для начала разработки Backend.

Но, по моему мнению, **самые важные документы еще впереди**:

1. **DOCUMENT 07 — Полная Frontend Architecture (React + RTK + Feature Folder + UI/UX)**
2. **DOCUMENT 08 — Mobile First Architecture (iOS + Android + API First)**
3. **DOCUMENT 09 — Security & Privacy by Design**
4. **DOCUMENT 10 — Product Vision, UX, конкурентные преимущества и сценарии использования**

Именно эти документы превратят проект из "дивидендного калькулятора" в полноценную инвестиционную платформу мирового уровня.
