# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 29

# DOMAIN DRIVEN DESIGN (DDD), CLEAN ARCHITECTURE, MODULE ISOLATION, PLUGIN SYSTEM & LONG-TERM SCALABILITY

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: CORE ARCHITECTURE

---

# PURPOSE

Настоящий документ определяет фундаментальную архитектуру OpenInvest.

После анализа предыдущих документов принято решение отказаться от построения системы как набора экранов.

OpenInvest проектируется как набор независимых доменов (Bounded Contexts), которые могут развиваться отдельно и взаимодействовать только через публичные контракты.

Это позволит развивать продукт 10–15 лет без переписывания ядра.

---

# 1. GLOBAL ARCHITECTURE

Запрещается строить приложение по принципу:

```
UI

↓

API

↓

Database
```

---

Используется модель:

```
Business Domain

↓

Application Layer

↓

Infrastructure Layer

↓

Presentation Layer
```

---

Бизнес всегда находится в центре.

---

# 2. DOMAIN DRIVEN DESIGN

OpenInvest разделяется на независимые домены.

---

## Portfolio Domain

Отвечает только за:

---

портфели;

---

позиции;

---

транзакции;

---

стоимость.

---

Не знает ничего о налогах.

---

Не знает ничего об ИИ.

---

Не знает ничего о дивидендах.

---

# 3. Asset Domain

Отвечает только за:

---

акции;

---

облигации;

---

тикеры;

---

сектора;

---

котировки.

---

# 4. Dividend Domain

Отвечает только за:

---

историю выплат;

---

будущие выплаты;

---

дивидендную доходность;

---

дивидендный календарь.

---

# 5. Tax Domain

Полностью изолирован.

---

Отвечает только за:

---

XML;

---

PDF;

---

налоги;

---

курсы ЦБ;

---

историю расчетов.

---

Portfolio Domain запрещено знать о Tax Domain.

---

# 6. Inflation Domain

Отвечает:

---

инфляция;

---

реальная стоимость;

---

индекс покупательной способности.

---

# 7. Analytics Domain

Отвечает:

---

XIRR;

---

TWR;

---

Sharpe;

---

Sortino;

---

Max Drawdown;

---

Volatility;

---

Dividend CAGR.

---

# 8. Notification Domain

Отвечает:

---

Email;

---

Push;

---

Reminder;

---

Dividend Alert;

---

Tax Alert.

---

# 9. AI Domain

Отвечает:

---

объяснение данных;

---

аналитика;

---

резюме;

---

визуализация.

---

ИИ не может изменять данные.

---

# 10. USER DOMAIN

Отвечает:

---

авторизация;

---

настройки;

---

язык;

---

валюта;

---

приватность.

---

# 11. CLEAN ARCHITECTURE

Каждый домен состоит из:

```
domain/

application/

infrastructure/

presentation/
```

---

# 12. DEPENDENCY RULE

Внутренний слой

не знает

о внешнем.

---

```
UI

↓

Application

↓

Domain

```

разрешено.

---

```
Domain

↓

UI
```

запрещено.

---

# 13. INTERFACES

Любое взаимодействие

происходит через интерфейс.

---

Никаких прямых зависимостей.

---

# 14. PLUGIN SYSTEM

OpenInvest проектируется как расширяемая система.

---

Сегодня:

```
Stocks

Bonds
```

---

Завтра:

```
ETF

Gold

Crypto

Deposits

Real Estate
```

---

Для этого вводится Plugin API.

---

# 15. PLUGIN CONTRACT

Каждый новый тип актива обязан реализовать:

```
GetPrice()

GetHistory()

GetIncome()

GetCurrency()

GetRisk()

GetAnalytics()
```

---

# 16. FEATURE ISOLATION

Каждая фича может быть полностью отключена.

---

Пример:

```
AI OFF

↓

Portfolio continues working.
```

---

```
Tax OFF

↓

Portfolio continues working.
```

---

```
Inflation OFF

↓

Dashboard continues working.
```

---

# 17. EVENT BUS

Домены общаются только событиями.

---

Пример:

```
TransactionCreated

↓

PortfolioUpdated

↓

AnalyticsRecalculated

↓

SnapshotCreated

↓

NotificationSent
```

---

Никаких прямых вызовов.

---

# 18. SNAPSHOT DOMAIN

Отдельный домен.

---

Отвечает:

---

создание снимков;

---

историю;

---

агрегирование;

---

кэширование.

---

# 19. CONFIGURATION DOMAIN

Все настройки:

---

Feature Flags;

---

Premium;

---

Currencies;

---

Regions;

---

Localization;

---

API Limits;

---

живут отдельно.

---

# 20. MULTI COUNTRY READY

Архитектура не должна зависеть от РФ.

---

Все специфичные правила:

---

налоги;

---

дивиденды;

---

валюты;

---

календари;

---

выносятся в Country Provider.

---

```
RussiaProvider

↓

Future

↓

USAProvider

↓

EUProvider

↓

KazakhstanProvider
```

---

# 21. TESTABILITY

Каждый домен тестируется независимо.

---

Unit

↓

Integration

↓

Contract

↓

E2E

---

# 22. CONTRACT TESTING

Frontend

и

Backend

не тестируют реализацию.

---

Они тестируют контракт.

---

OpenAPI —

единственный источник истины.

---

# 23. INTERNAL PACKAGE RULES

Запрещено:

```
portfolio

↓

import tax

↓

import analytics

↓

import notifications
```

---

Разрешено:

```
portfolio

↓

publish event

↓

event bus

↓

analytics subscribe
```

---

# 24. MODULE MATURITY

Каждый модуль имеет статус:

```
Experimental

↓

Beta

↓

Stable

↓

Deprecated

↓

Archived
```

---

# 25. BACKWARD COMPATIBILITY

Удаление API запрещено.

---

Используется стратегия:

```
Deprecated

↓

Migration

↓

Replacement

↓

Removal
```

---

# 26. CODE OWNERSHIP

Builder Agent

↓

Review Agent

↓

QA Agent

↓

Security Agent

↓

Performance Agent

↓

Human

---

Только после прохождения всей цепочки разрешается Push.

---

# 27. FILE SIZE RULE

Файл не должен превышать:

```
300–400 строк
```

---

Пакет

не более

одной ответственности.

---

# 28. COMPONENT SIZE RULE

React компонент

не более

250 строк.

---

Hook

не более

150 строк.

---

Service

не более

300 строк.

---

# 29. LONG TERM EVOLUTION

Через 5 лет проект должен позволять добавить новый актив

без изменения существующего Portfolio Domain.

---

Это достигается через:

---

Open/Closed Principle;

---

Plugin API;

---

DDD;

---

Event Bus;

---

Clean Architecture.

---

# 30. FINAL PRINCIPAL ENGINEERING RULE

> **OpenInvest не является монолитным веб-сайтом.**

> **OpenInvest представляет собой набор независимых финансовых доменов, объединенных общей математической моделью, единым API-контрактом и событийной архитектурой.**

> **Добавление новой функции не должно приводить к изменению уже работающего ядра. Любое нарушение этого правила считается архитектурным дефектом и должно быть отклонено Review Agent до попадания кода в основную ветку.**
