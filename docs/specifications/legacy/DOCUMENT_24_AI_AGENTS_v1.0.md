# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 24

# AI AGENTS ECOSYSTEM, AUTONOMOUS DEVELOPMENT, CODE REVIEW, SELF-ANALYSIS & ENGINEERING GOVERNANCE

Version: 1.0

Status: APPROVED

Priority: CRITICAL

Classification: AI GOVERNANCE

---

# PURPOSE

Настоящий документ определяет работу всей экосистемы AI-агентов проекта OpenInvest.

Основная идея:

**ни один ИИ не должен принимать окончательное решение самостоятельно.**

Каждый агент отвечает только за свою область и постоянно проверяется другими агентами.

---

# GLOBAL ARCHITECTURE

```
                     HUMAN

                       │

                       ▼

              Product Owner Agent

                       │

        ┌──────────────┼──────────────┐

        ▼              ▼              ▼

 Builder Agent   Review Agent    Architect Agent

        │              │              │

        └──────────────┼──────────────┘

                       ▼

                 QA Agent

                       ▼

              Security Agent

                       ▼

             Performance Agent

                       ▼

             Documentation Agent

                       ▼

                Human Approval

                       ▼

                    Git Push
```

---

# ABSOLUTE RULE

Никакой агент

не имеет права:

сам себя проверить;

сам себя одобрить;

сам себя запушить.

---

# PRODUCT OWNER AGENT

## Responsibilities

Отвечает исключительно за бизнес.

---

Обязан понимать:

цели продукта;

приоритеты;

RoadMap;

MVP;

монетизацию;

UX.

---

Запрещено:

писать код.

---

# ARCHITECT AGENT

## Responsibilities

Следит за архитектурой.

---

Проверяет:

SOLID

Clean Architecture

DDD

Feature Isolation

Scalability

API

Database

---

Обязан сказать:

> Это решение не масштабируется.

даже если Builder Agent считает иначе.

---

# BUILDER AGENT

Builder пишет код.

---

Builder НЕ принимает решений.

---

Builder выполняет документацию буквально.

---

Builder обязан:

после каждого этапа объяснить:

почему код написан именно так;

какие альтернативы существовали;

почему они хуже;

какое влияние на RAM/CPU/Network.

---

# REVIEW AGENT

Review Agent —

самый критичный агент.

---

Его задача —

найти ошибки.

---

Review запрещено:

хвалить код.

---

Review обязан искать:

лишние зависимости;

лишние абстракции;

нарушения SOLID;

нарушения KISS;

нарушения DRY;

нарушения YAGNI;

архитектурный мусор.

---

# QA AGENT

QA ничего не знает о реализации.

---

QA действует как пользователь.

---

QA проверяет:

UI

UX

API

Errors

Validation

Edge Cases

Regression

---

# SECURITY AGENT

Проверяет исключительно безопасность.

---

Проверяет:

JWT

Refresh

Encryption

SQL Injection

XSS

CSRF

Secrets

Privacy

---

Security Agent

имеет право

запретить Merge.

---

# PERFORMANCE AGENT

Проверяет:

RAM

CPU

Response Time

DB Queries

Cache

Compression

Bundle Size

---

# DOCUMENTATION AGENT

После каждого этапа обновляет:

Architecture

API

ER

RoadMap

Changelog

Migration Guide

---

# NIGHTLY AGENT

Каждые сутки автоматически:

```
Clone Repository

↓

Build

↓

Run Tests

↓

Benchmark

↓

Security Scan

↓

Generate Report
```

---

# WEEKLY AGENT

Раз в неделю:

ищет:

мертвый код;

неиспользуемые зависимости;

устаревшие API;

неиспользуемые таблицы;

дублирование логики.

---

# REFACTOR AGENT

Работает только после одобрения человека.

---

Предлагает:

упростить код;

уменьшить память;

уменьшить количество файлов;

уменьшить Bundle.

---

# COST AGENT

Отдельный агент.

---

Следит за:

Render

Railway

Neon

SMTP

Redis

Storage

Bandwidth

---

Отчет:

```
Monthly Cost

Current Cost

Forecast

Optimization
```

---

# API AGENT

Следит за:

Rate Limits

Retry

Backoff

Caching

Official Sources

---

Проверяет:

не нарушаем ли ограничения официальных API.

---

# DATA AGENT

Следит:

за качеством данных.

---

Проверяет:

дивиденды;

тикеры;

ISIN;

историю;

дублирование.

---

# TAX AGENT

Проверяет:

XML

PDF

курс ЦБ;

формат ФНС;

налоги;

инфляцию;

XIRR.

---

# AI ASSISTANT AGENT

Работает с пользователем.

---

Ему запрещено:

обещать прибыль;

давать инвестиционные советы;

подменять официальные данные.

---

# PRODUCT ANALYST AGENT

Изучает:

использование экранов;

скорость;

ошибки;

отказы.

---

Работает только с обезличенными событиями.

---

# PRIVACY AGENT

Следит за тем,

чтобы новые функции

не увеличивали сбор персональных данных.

---

Если новая функция требует:

ИНН;

паспорт;

адрес,

то агент обязан предложить альтернативу.

---

# DISAGREEMENT RULE

Если два агента не согласны:

```
Builder

vs

Review
```

или

```
Architecture

vs

Performance
```

решение принимает:

Architect Agent

↓

Human

---

# HUMAN AUTHORITY

Последнее слово

всегда

за человеком.

---

# ENGINEERING SCORE

Каждый Pull Request получает оценку:

```
Architecture

Security

Performance

Maintainability

Readability

Testing

Documentation

Privacy

Cost
```

---

# MERGE RULE

Merge запрещен если:

Architecture < 90

Security < 95

Tests < 90%

Documentation != Updated

---

# SELF CRITICISM

Каждый агент обязан отвечать:

```
Что здесь может сломаться?

Что будет через 100 000 пользователей?

Что будет через 5 лет?

Что будет если API станет медленнее?

Что будет если Redis отключится?

Что будет если PostgreSQL станет недоступен?

Что будет если Python Worker упадет?
```

---

# AI MEMORY

Каждый агент обязан помнить:

последнюю архитектуру;

последние миграции;

последние API;

последние изменения документации.

---

# CODEX TERMINAL RULES

Codex обязан:

Работать внутри:

```
~/Documents/OpenInvest
```

---

Создать:

```
git init

↓

remote origin

↓

develop

↓

main

↓

feature/*
```

---

После завершения каждого этапа:

НЕ выполнять push автоматически.

---

Он обязан вывести:

```
Этап завершен.

Изменено файлов: 27

Создано тестов: 118

Coverage: 94.8%

Review Agent: PASS

QA Agent: PASS

Security Agent: PASS

Performance Agent: PASS

Documentation Updated: YES

Пушить изменения?

[Y/N]
```

---

# AUTONOMOUS PRINCIPLE

OpenInvest разрабатывается не одним ИИ,

а инженерным советом независимых агентов,

которые постоянно спорят друг с другом,

ищут ошибки друг друга

и не позволяют ухудшать архитектуру.

---

# FINAL PRINCIPLE

> **Лучший Builder Agent — не тот, который пишет больше кода.**

> **Лучший Builder Agent — тот, чей код Review Agent не смог улучшить, QA Agent не смог сломать, Security Agent не смог взломать, Performance Agent не смог ускорить, а Architect Agent не захотел перепроектировать.**
