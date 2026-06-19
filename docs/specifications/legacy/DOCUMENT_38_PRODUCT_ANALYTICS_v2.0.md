# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 38

# PRODUCT ANALYTICS, BUSINESS METRICS, TELEMETRY, FEATURE FLAGS, A/B TESTING & DATA DRIVEN PRODUCT DEVELOPMENT

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: PRODUCT ANALYTICS CONSTITUTION

---

# PURPOSE

Настоящий документ определяет принципы аналитики продукта OpenInvest.

Главная цель аналитики — улучшать продукт, а не собирать информацию о пользователях.

OpenInvest придерживается принципа:

> **Privacy First Analytics**

---

# PRODUCT PHILOSOPHY

Мы измеряем:

---

качество продукта;

---

удобство интерфейса;

---

скорость работы;

---

понимание пользователем своего капитала.

---

Мы не измеряем:

---

личную финансовую информацию;

---

пароли;

---

паспортные данные;

---

ИНН;

---

размер капитала конкретного пользователя.

---

# NORTH STAR METRIC

Главная метрика продукта

не количество пользователей.

---

Главная метрика:

```text
Users Who Better Understand Their Capital
```

---

Пользователь должен регулярно возвращаться

не ради проверки котировок,

а ради понимания своего финансового состояния.

---

# PRODUCT METRICS

## Acquisition

---

Registration Rate

---

Portfolio Created

---

First Transaction

---

First Dividend View

---

Tax Export

---

# ACTIVATION

Пользователь считается активированным,

если:

```text
Registration

↓

Portfolio Created

↓

Transaction Added

↓

Dashboard Viewed
```

---

# RETENTION

Измеряется:

---

Day 1

---

Day 7

---

Day 30

---

Day 90

---

Year 1

---

# ENGAGEMENT

Среднее время:

---

Dashboard

---

Portfolio

---

Asset Card

---

Dividend Calendar

---

Analytics

---

Tax

---

# FEATURE ADOPTION

Каждая функция имеет собственную аналитику.

---

Пример:

```text
Portfolio

↓

Opened

↓

Scrolled

↓

Chart Opened

↓

Dividend Card

↓

Tax Button
```

---

# SEARCH ANALYTICS

Измеряется:

---

поиск тикеров;

---

поиск компаний;

---

поиск облигаций.

---

Не сохраняются финансовые решения пользователя.

---

# DASHBOARD ANALYTICS

Измеряется:

---

просмотр Real Value;

---

просмотр Inflation;

---

просмотр Dividend Forecast;

---

просмотр XIRR.

---

# CHART ANALYTICS

Измеряется:

---

смена периода;

---

масштабирование;

---

открытие детализации.

---

# UX METRICS

Time to Portfolio

---

Time to Transaction

---

Time to Dividend

---

Time to Tax Export

---

# CLICK DEPTH

Любое действие должно выполняться

максимум за

3 клика.

---

Если среднее значение выше,

создается Product Issue.

---

# AB TESTING

Разрешается тестировать:

---

расположение карточек;

---

цвет CTA;

---

структуру Dashboard;

---

новые графики.

---

Запрещается тестировать:

---

математические формулы;

---

налоговые расчеты;

---

финансовые показатели;

---

безопасность.

---

# FEATURE FLAGS

Любая новая функция:

```text
OFF

↓

Internal

↓

Beta

↓

5%

↓

20%

↓

50%

↓

100%
```

---

# USER SEGMENTS

Anonymous

---

Registered

---

Active

---

Premium

---

Beta

---

Internal

---

# PRIVACY ANALYTICS

Все события обезличены.

---

Используется:

UUID

---

Запрещается использовать:

Email

Телефон

ИНН

Паспорт

---

# EVENT NAMING

Единый формат:

```text
portfolio.open

portfolio.create

transaction.add

asset.search

calendar.open

tax.export
```

---

# EVENT VERSIONING

Каждое событие имеет:

---

Version

---

CreatedAt

---

Schema

---

Owner

---

# ANALYTICS PIPELINE

```text
Client

↓

Validation

↓

Queue

↓

Aggregation

↓

Storage

↓

Dashboard
```

---

Сырые события

не используются напрямую.

---

# PRODUCT DASHBOARD

Product Team видит:

---

DAU

---

WAU

---

MAU

---

Retention

---

Activation

---

Errors

---

Crash Rate

---

Latency

---

Feature Adoption

---

# ENGINEERING DASHBOARD

Показывает:

---

API Latency

---

Cache Hit

---

CPU

---

Memory

---

Database

---

Queue

---

# BUSINESS DASHBOARD

Показывает:

---

Registrations

---

Active Portfolios

---

Premium Conversion

---

Feature Usage

---

Tax Exports

---

AI Usage

---

# COST DASHBOARD

Показывает:

---

Server Cost

---

Database Cost

---

Storage Cost

---

Email Cost

---

API Cost

---

LLM Cost

---

Cost per Active User

---

# AI ANALYTICS

Измеряется:

---

Answer Time

---

Confidence

---

Feedback

---

Regeneration

---

Human Approval

---

# ERROR ANALYTICS

Любая ошибка получает:

---

Severity

---

Frequency

---

Affected Users

---

Regression Status

---

# USER FEEDBACK

Встроенная система:

---

👍

---

👎

---

Комментарий

---

Без обязательного текста.

---

# HEATMAPS

Разрешается использовать

только после явного согласия пользователя.

---

По умолчанию отключены.

---

# SESSION RECORDING

По умолчанию запрещено.

---

Может быть включено

только в Beta Program

и только после отдельного согласия.

---

# DATA RETENTION

Analytics

24 месяца

---

Aggregated Metrics

без ограничения

---

Raw Events

90 дней

---

# GDPR / PRIVACY READY

Пользователь может:

---

отключить аналитику;

---

скачать аналитику;

---

удалить аналитику;

---

просмотреть историю аналитики.

---

# SUCCESS METRICS

OpenInvest считается успешным,

если пользователь:

---

понимает свой капитал лучше;

---

возвращается регулярно;

---

не нуждается в сторонних калькуляторах;

---

не тратит время на ручные расчеты.

---

# REVIEW CHECKLIST

Перед добавлением нового события необходимо ответить:

---

Можно ли обойтись без него?

---

Можно ли агрегировать его сразу?

---

Можно ли удалить персональные данные?

---

Можно ли уменьшить объем хранения?

---

Можно ли использовать существующее событие?

---

# FINAL PRODUCT PRINCIPLE

> **OpenInvest измеряет качество продукта, а не жизнь пользователя.**

> **Аналитика существует исключительно для улучшения пользовательского опыта, производительности и надежности системы.**

> **Любой сбор данных, который не приносит прямой пользы пользователю или продукту, считается избыточным и должен быть отклонен на этапе Product Review.**
