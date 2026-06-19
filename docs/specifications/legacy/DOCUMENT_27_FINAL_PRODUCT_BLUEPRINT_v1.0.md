# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 27

# FINAL PRODUCT BLUEPRINT, SYSTEM ARCHITECTURE, PRODUCT VISION, MODULE MAP & DEVELOPMENT CONSTITUTION

Version: 1.0

Status: FINAL

Priority: ABSOLUTE

Classification: PROJECT CONSTITUTION

---

# 1. MISSION

Создать лучший бесплатный сервис для долгосрочного инвестора.

Не брокера.

Не торговый терминал.

Не новостной портал.

Не AI-советника.

---

OpenInvest —

это интеллектуальный финансовый ассистент, который:

помогает понимать капитал;

помогает учитывать инвестиции;

помогает считать дивиденды;

помогает считать реальную доходность;

помогает готовить налоговую декларацию;

не навязывает покупку активов;

не продает инвестиционные идеи.

---

# 2. PRODUCT PHILOSOPHY

OpenInvest должен быть:

самым быстрым;

самым понятным;

самым прозрачным;

самым приватным;

самым энергоэффективным;

самым честным.

---

# 3. PRODUCT PRINCIPLES

Каждая функция должна отвечать пяти вопросам.

---

### Она быстрее?

---

### Она проще?

---

### Она уменьшает количество действий пользователя?

---

### Она не нарушает приватность?

---

### Она масштабируется на 100 000 пользователей?

---

Если хотя бы один ответ отрицательный —

функция отправляется на переработку.

---

# 4. PRODUCT DIFFERENTIATORS

## Большинство брокеров

перегружены.

---

OpenInvest

максимально простой.

---

## Большинство сервисов

показывают номинальную доходность.

---

OpenInvest

показывает:

номинальную;

реальную;

инфляционную;

дивидендную;

налоговую;

XIRR.

---

## Большинство сервисов

собирают огромное количество персональных данных.

---

OpenInvest

работает даже без:

паспорта;

ИНН;

адреса;

телефона.

---

## Большинство сервисов

ждут действий пользователя.

---

OpenInvest

проактивно:

обнаруживает дивиденды;

обновляет прогноз;

создает налоговый отчет;

уведомляет пользователя.

---

# 5. TARGET AUDIENCE

Новичок

↓

Долгосрочный инвестор

↓

Дивидендный инвестор

↓

Инвестор FIRE

↓

Пенсионный инвестор

↓

Семейный инвестор

---

# 6. CORE MODULES

## Landing

---

## Asset Catalog

---

## Asset Card

---

## Dividend Calculator

---

## Portfolio

---

## Portfolio Analytics

---

## Dividend Calendar

---

## Tax Assistant

---

## Inflation Analytics

---

## AI Assistant

---

## Settings

---

## Profile

---

# 7. DASHBOARD

Главный экран.

---

Показывает:

Стоимость портфеля

↓

Доход

↓

XIRR

↓

Дивиденды

↓

Ожидаемые выплаты

↓

Real Value

↓

Инфляционную корректировку

↓

Календарь ближайших выплат

↓

Новости только по портфелю

---

# 8. REAL VALUE (UNIQUE FEATURE)

Кроме стоимости:

```text
1 200 000 ₽
```

показывается:

---

Эквивалент:

MacBook Pro

7 шт

---

iPhone

16 шт

---

Средняя продуктовая корзина

24 месяца

---

Средняя аренда квартиры

15 месяцев

---

Средняя зарплата РФ

8 месяцев

---

Покупательская способность

+8%

или

−4%

---

Пользователь начинает понимать

не цифры,

а реальную стоимость капитала.

---

# 9. PORTFOLIO

Показывает:

---

Стоимость сейчас

---

Свободные деньги

---

Акции

---

Облигации

---

ETF (future)

---

Фонды (future)

---

Дивиденды

---

Купоны

---

Комиссии

---

Налоги

---

Инфляцию

---

Real XIRR

---

# 10. MATHEMATICAL ENGINE

Используется:

---

Weighted Average Cost

---

XIRR

---

Dividend Yield

---

Coupon Yield

---

Inflation Adjustment

---

Real Return

---

Tax Impact

---

# 11. TAX MODULE

Режимы:

---

Private Mode

---

Convenience Mode

---

Экспорт:

XML

PDF

ZIP

Email

---

Human in the Loop

обязателен.

---

# 12. AI

ИИ:

объясняет;

анализирует;

сравнивает;

визуализирует.

---

ИИ запрещено:

советовать покупать;

советовать продавать;

обещать прибыль.

---

# 13. API STRATEGY

Все данные:

↓

Go API

↓

Cache

↓

Redis

↓

PostgreSQL

↓

Official APIs

---

Пользователь

никогда

не обращается к MOEX напрямую.

---

# 14. DATA SOURCES

Используются только:

---

официальный MOEX;

---

официальный ЦБ;

---

официальный Росстат;

---

официальные раскрытия эмитентов;

---

официальные XSD ФНС.

---

# 15. MOBILE STRATEGY

Web

↓

iOS SwiftUI

↓

Android Jetpack Compose

↓

macOS (future)

↓

watchOS (future)

---

# 16. PERFORMANCE TARGETS

Dashboard

<150 ms

---

API

<100 ms

---

Cold Start

<1.5 sec

---

Warm Start

<500 ms

---

Bundle

<250 KB gzip

---

Memory Backend

<512 MB

---

# 17. FREE INFRASTRUCTURE STRATEGY

Frontend

Vercel

---

Backend

Railway/Render

---

Database

Neon

---

Redis

Upstash

---

Monitoring

Grafana

---

# 18. MONETIZATION

## Free

Неограниченное количество сделок

Портфель

Дивиденды

Календарь

XIRR

Real Value

Инфляция

---

## Premium

AI аналитика

Сценарии

Что если купить

Portfolio Compare

История изменений

Автоматические отчеты

Расширенные уведомления

Экспорт Excel

---

## B2B Future

White Label

Advisor Dashboard

Family Office

Tax Assistant API

---

# 19. SECURITY

Zero Trust

---

Privacy by Design

---

AES-256

---

JWT

---

Refresh Rotation

---

Private Mode

---

Delete by One Click

---

Export My Data

---

# 20. DEVELOPMENT PROCESS

Documentation

↓

Architecture Review

↓

Builder

↓

Review

↓

QA

↓

Security

↓

Performance

↓

Human Approval

↓

Git Push

↓

Deploy

---

# 21. NON-NEGOTIABLE ENGINEERING PRINCIPLES

SOLID

SRP

OCP

LSP

ISP

DIP

DRY

KISS

YAGNI

Law of Demeter

Occam Razor

Composition over Inheritance

Feature Isolation

Clean Architecture

API First

Privacy First

Mobile First

Offline First

---

# 22. WHAT MAKES OPENINVEST UNIQUE

Не количество функций.

---

А комбинация:

✓ XIRR

✓ Real Return

✓ Inflation Analytics

✓ Tax Assistant

✓ Privacy Mode

✓ Zero Mandatory Passport Data

✓ Human-in-the-Loop

✓ Proactive Dividend Engine

✓ Lightweight Architecture

✓ Fast Mobile Experience

✓ Official Data Only

✓ Transparent Calculations

✓ Audit Trail

✓ Energy Efficient Design

---

# 23. SUCCESS CRITERIA

Через 3 года OpenInvest должен оставаться:

---

понятным новичку;

---

полезным профессионалу;

---

быстрым на старом телефоне;

---

дешевым в эксплуатации;

---

безопасным;

---

масштабируемым;

---

архитектурно чистым.

---

# 24. FINAL ENGINEERING COMMAND FOR CODEX

Codex обязан считать этот документ основной конституцией проекта.

При любом конфликте:

```
Код
↓

Документация

↓

Архитектура

↓

Принципы

↓

Конституция проекта
```

приоритет всегда имеет:

**Конституция проекта (Document 27).**

---

# FINAL PRODUCT STATEMENT

> **OpenInvest — это не еще один дивидендный калькулятор.**

> **Это персональная финансовая операционная система долгосрочного инвестора, построенная на принципах прозрачности, приватности, математической корректности, энергоэффективности и инженерной дисциплины мирового уровня.**
