# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 28

# ARCHITECTURE REFINEMENT, DATA GOVERNANCE, API STRATEGY, ADVANCED MATHEMATICS, ADR, C4, ER MODEL, OPENAPI-FIRST & LONG TERM EVOLUTION

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: PRINCIPAL ENGINEERING ADDENDUM

---

# PURPOSE

Настоящий документ заменяет и расширяет предыдущие архитектурные решения.

После анализа всех документов были обнаружены потенциальные архитектурные риски, которые должны быть устранены до начала разработки.

Цель документа:

* убрать логические противоречия;
* обеспечить масштабирование до 100 000+ пользователей;
* минимизировать стоимость эксплуатации;
* обеспечить юридическую чистоту;
* сделать продукт расширяемым на 10+ лет.

---

# 1. PRODUCT REPOSITIONING

## Старое позиционирование

```
Dividend Calculator
```

↓

неверно.

---

## Новое позиционирование

```
OpenInvest

Personal Capital Operating System

(Personal Capital OS)
```

---

OpenInvest —

это операционная система личного капитала.

---

Сегодня поддерживаются

* акции;
* облигации;
* дивиденды.

---

Завтра

* ETF;
* REIT;
* драгоценные металлы;
* банковские вклады;
* недвижимость;
* пенсионные накопления;
* семейные портфели;
* цели;
* денежные потоки;
* налоговый учет.

---

# 2. DATA GOVERNANCE

Самая важная архитектурная идея проекта.

---

## Пользователь никогда не работает с внешними API.

```
Official APIs

↓

Collector Layer

↓

Normalization Layer

↓

Validation Layer

↓

Canonical Database

↓

Redis / RAM

↓

Go API

↓

Client
```

---

OpenInvest становится владельцем собственной модели данных.

---

# 3. CANONICAL DATA MODEL

Никогда не использовать внешнюю структуру напрямую.

---

Каждый источник проходит:

```
Raw Data

↓

Parser

↓

Validator

↓

Normalizer

↓

Canonical Entity

↓

Storage
```

---

Это позволяет:

сменить поставщика данных;

использовать несколько источников;

исключить зависимость от MOEX.

---

# 4. OFFICIAL API STRATEGY

Используются исключительно официальные бесплатные источники.

---

Но пользователь никогда не обращается к ним напрямую.

---

Python Worker получает данные.

---

Go API обслуживает пользователей.

---

# 5. DATA SOURCE LIFECYCLE

Каждый источник проходит обязательную проверку.

```
Business Review

↓

Legal Review

↓

License Review

↓

Technical Review

↓

Security Review

↓

Performance Review

↓

Approval
```

---

Builder Agent запрещено подключать новые API самостоятельно.

---

# 6. JURIDICAL DATA POLICY

Для каждого API необходимо хранить:

```
Source

License

Commercial Usage

Redistribution

Storage Rules

Caching Rules

Retention Policy

Expiration Policy
```

---

# 7. CACHE STRATEGY V2

Используется четырехуровневая система.

```
L1 RAM

↓

L2 Redis

↓

L3 PostgreSQL

↓

L4 Official Source
```

---

Пользователь всегда получает данные из ближайшего уровня.

---

# 8. EVENT DRIVEN ARCHITECTURE

Вместо постоянных запросов.

```
Dividend Approved

↓

Event

↓

Queue

↓

Portfolio Update

↓

Notification

↓

Tax Update
```

---

# 9. ADVANCED MATHEMATICAL ENGINE

OpenInvest становится профессиональной аналитической системой.

---

## Weighted Average Cost

Средневзвешенная стоимость.

---

## FIFO

Для сравнения.

---

## XIRR

Money Weighted Return.

---

## Time Weighted Return (TWR)

Позволяет сравнивать инвестора с индексом.

---

## CAGR

Compound Annual Growth Rate.

---

Среднегодовой темп роста капитала.

---

## Dividend CAGR

Среднегодовой рост дивидендов компании.

---

## Yield on Cost

Дивидендная доходность относительно собственной цены покупки.

Очень важный показатель для долгосрочного инвестора.

---

## Sharpe Ratio

Доходность относительно общего риска.

---

## Sortino Ratio

Доходность относительно отрицательного риска.

---

## Maximum Drawdown

Максимальная историческая просадка.

---

## Volatility

Стандартное отклонение доходности.

---

## Inflation Adjusted Return

```
Portfolio Return

-

Inflation
```

---

## Tax Adjusted Return

```
Portfolio Return

-

Taxes
```

---

## Net Real Return

```
Portfolio Return

-

Inflation

-

Taxes

-

Commissions

-

NKD
```

---

# 10. PURCHASING POWER INDEX

Эксклюзивная функция OpenInvest.

---

Вместо:

```
Portfolio

1 850 000 ₽
```

показывать

```
Real Capital

MacBook Pro

9.4

iPhone

22

Средняя зарплата

13 месяцев

Продуктовая корзина

31 месяц

Аренда квартиры

19 месяцев
```

---

И показывать изменение покупательной способности во времени.

---

# 11. ARCHITECTURE DECISION RECORDS

Каждое решение фиксируется.

---

Пример:

```
ADR-001

Почему Go

Статус

Accepted

Причины

Последствия

Альтернативы
```

---

Минимальный набор ADR:

```
Go

Python

PostgreSQL

Redis

Feature Folder

Private Mode

API First

Snapshot

XIRR

Canonical Data

Event Driven
```

---

# 12. ER MODEL

Минимальная модель данных.

```
Users

↓

Profiles

↓

Portfolios

↓

Transactions

↓

Assets

↓

Snapshots

↓

DividendDirectory

↓

Coupons

↓

Notifications

↓

TaxExports

↓

AuditLogs

↓

FeatureFlags
```

---

ER должна существовать отдельно

до начала кодирования.

---

# 13. C4 MODEL

Обязательна.

---

## Level 1

System Context

---

## Level 2

Containers

Frontend

Go

Python

Redis

PostgreSQL

SMTP

---

## Level 3

Components

Portfolio

Tax

Catalog

AI

Notifications

Auth

Analytics

---

## Level 4

Code

Пакеты

Модули

Интерфейсы

---

# 14. SEQUENCE DIAGRAMS

Обязательно создать.

---

Добавление сделки.

---

Создание портфеля.

---

Генерация XML.

---

Email отправка.

---

Получение дивидендов.

---

Обновление Snapshot.

---

AI анализ.

---

# 15. OPENAPI FIRST

Разработка начинается

не с кода.

---

Последовательность:

```
Product

↓

Domain

↓

OpenAPI

↓

Swagger

↓

SDK

↓

Backend

↓

Frontend

↓

Mobile
```

---

Backend запрещено писать

до утверждения OpenAPI.

---

# 16. VERSIONING

```
v1

↓

v2

↓

v3
```

---

Никаких breaking changes.

---

# 17. DATABASE EVOLUTION

Использовать стратегию

Expand / Migrate / Contract.

---

```
Add Column

↓

Fill Data

↓

Switch Reads

↓

Switch Writes

↓

Remove Old Column
```

---

Никаких destructive migrations.

---

# 18. PREMIUM STRATEGY V2

Premium

не должен забирать базовый функционал.

---

Free:

неограниченные сделки;

неограниченный портфель;

дивиденды;

XIRR;

Real Return;

календарь;

инфляция.

---

Premium:

Monte Carlo;

Stress Test;

Goal Planner;

Retirement Planner;

AI Portfolio Review;

Dividend Scenarios;

Tax Forecast;

Family Office;

Advanced Reports;

Portfolio Compare.

---

# 19. COMPETITIVE STRATEGY

Не конкурировать:

с БКС;

с Альфой;

с Т-Банком;

с ВТБ;

по количеству функций.

---

Конкурировать:

по скорости;

по прозрачности;

по приватности;

по математике;

по энергоэффективности;

по честности.

---

# 20. ENGINEERING CHECKLIST

Перед реализацией любой функции Builder Agent обязан ответить:

---

Можно ли сделать проще?

---

Можно ли уменьшить RAM?

---

Можно ли уменьшить CPU?

---

Можно ли уменьшить трафик?

---

Можно ли отказаться от хранения данных?

---

Можно ли уменьшить количество запросов?

---

Можно ли реализовать через Snapshot?

---

Можно ли реализовать через Event?

---

Можно ли отказаться от синхронной обработки?

---

# 21. FINAL PRINCIPLE

Через 10 лет архитектура должна позволять:

```
10 000 000+

Transactions

↓

500 000+

Users

↓

Millions of Snapshots

↓

Dozens of Asset Classes

↓

Without rewriting Core Architecture
```

---

# FINAL PRINCIPAL ENGINEER STATEMENT

> OpenInvest должен проектироваться не как веб-сайт и не как дивидендный калькулятор.

> Он должен проектироваться как **Personal Capital Platform**, где дивиденды, налоги, аналитика, инфляция и управление капиталом являются отдельными независимыми доменами, объединенными общей архитектурой, единой математической моделью, официальными источниками данных, строгой приватностью и API-First подходом.

> Любое решение, которое может привести к переписыванию ядра через 3–5 лет, считается архитектурной ошибкой и должно быть отклонено на этапе проектирования.
