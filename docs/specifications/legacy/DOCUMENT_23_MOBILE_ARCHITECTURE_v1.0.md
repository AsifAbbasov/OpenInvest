# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 23

# MOBILE ARCHITECTURE, OFFLINE-FIRST, BATTERY OPTIMIZATION, NETWORK STRATEGY & CROSS PLATFORM DESIGN

Version: 1.0

Status: APPROVED

Priority: CRITICAL

Classification: MOBILE FOUNDATION

---

# PURPOSE

Настоящий документ определяет всю архитектуру мобильных приложений OpenInvest.

Документ обязателен для:

* Codex
* Builder Agent
* Review Agent
* QA Agent
* Mobile Team

---

# PRODUCT PHILOSOPHY

OpenInvest НЕ является тяжелым брокерским терминалом.

OpenInvest —

это быстрый интеллектуальный помощник инвестора.

---

Главная цель:

Пользователь открывает приложение

↓

за 1 секунду понимает состояние капитала

↓

закрывает приложение.

---

# MOBILE FIRST

Любая новая функция проектируется:

```
Phone

↓

Tablet

↓

Desktop

↓

Web
```

а не наоборот.

---

# SUPPORTED PLATFORMS

## Phase 1

Web

---

## Phase 2

iOS

SwiftUI

---

## Phase 3

Android

Jetpack Compose

---

## Future

macOS

visionOS

watchOS

---

# WHY NATIVE?

Использование:

SwiftUI

*

Jetpack Compose

дает:

---

меньше RAM;

---

меньше CPU;

---

меньше Battery Usage;

---

лучше интеграцию с системой;

---

меньше лагов;

---

более долгую поддержку.

---

# ABSOLUTE PRINCIPLE

Никакой бизнес-логики

в мобильном приложении.

---

Мобильное приложение —

это UI.

---

Вся математика:

Go Backend

↓

PostgreSQL

↓

Redis

↓

Python Worker

---

# OFFLINE FIRST

Пользователь должен видеть приложение,

даже если интернет пропал.

---

# LOCAL CACHE

Кэшируются:

Portfolio Summary

Portfolio Chart

Dividend Calendar

Asset Cards

Settings

Theme

Language

---

# NEVER CACHE

JWT

Refresh Token

Passport

INN

Address

Tax Documents

---

# CACHE TTL

Portfolio Summary

5 минут

---

Asset Catalog

30 минут

---

Dividend Calendar

12 часов

---

Settings

30 дней

---

Theme

∞

---

# CACHE STRATEGY

```
Open Screen

↓

Check Local Cache

↓

Show Cached Data

↓

Background Refresh

↓

Update UI Smoothly
```

---

# USER EXPERIENCE

Пользователь никогда

не должен видеть:

```
Loading...

Loading...

Loading...
```

---

Сначала показывается кэш.

Потом тихо обновляются данные.

---

# NETWORK STRATEGY

Каждый запрос обязан пройти проверку:

```
Можно ли НЕ делать этот запрос?
```

Если можно —

не делать.

---

# DELTA UPDATES

Передавать

не весь объект,

а только изменения.

---

НЕ:

```
Portfolio

2 MB
```

---

ДА:

```
Price Changed

4 KB
```

---

# COMPRESSION

Все API:

Brotli

↓

Gzip

↓

JSON

---

# REQUEST AGGREGATION

Запрещено:

```
10 экранов

↓

10 API запросов
```

---

Разрешено:

```
Dashboard

↓

1 Aggregated API

↓

Все данные
```

---

# IMAGE STRATEGY

SVG

↓

WebP

↓

AVIF

---

PNG

только при необходимости.

---

# CHART STRATEGY

Графики строятся

по snapshots.

---

Телефон

никогда

не пересчитывает

5 лет истории.

---

Он получает:

```
Date

↓

Value

↓

Draw
```

---

# BATTERY OPTIMIZATION

Запрещено:

Polling каждую секунду.

---

Использовать:

Background Refresh

System Scheduling

Push Events

---

# IOS

Использовать:

BackgroundTasks

---

# ANDROID

Использовать:

WorkManager

---

# PUSH STRATEGY

Push

только при:

новых дивидендах;

изменении налогового статуса;

готовом XML;

важном событии.

---

Запрещено:

рекламные Push.

---

# SYNCHRONIZATION

Все серверы работают

только по UTC.

---

Пользователь видит

локальное время.

---

# MOEX STRATEGY

Источник истины —

торговый календарь MOEX.

---

Не часовой пояс пользователя.

---

# EXAMPLE

Пользователь

в Австралии.

---

Биржа закрылась в Москве.

---

Snapshot уже создан.

---

Пользователь открывает приложение.

---

Получает:

готовый Snapshot.

---

Никаких пересчетов.

---

# LOW INTERNET MODE

Отдельный режим.

---

Отключается:

анимация;

автообновление;

новости;

изображения.

---

Остаются:

цифры;

графики;

портфель.

---

# BATTERY SAVER MODE

Отключаются:

Background Refresh

Heatmaps

Animations

Realtime Updates

---

# ACCESSIBILITY

Dynamic Font

VoiceOver

TalkBack

High Contrast

---

# WIDGETS

Будущая поддержка:

---

iOS Widget

---

Android Widget

---

Показывают:

Стоимость

*

Дивиденды

*

Real Value

---

# QUICK ACTIONS

Через Widget:

```
Добавить сделку

↓

Открыть календарь

↓

Скачать налоговый отчет
```

---

# BIOMETRICS

FaceID

TouchID

Fingerprint

---

опционально.

---

# SPLASH SCREEN

Максимум

500 ms.

---

# APP START TARGET

Cold Start

<1.5 sec

---

Warm Start

<500 ms

---

# MEMORY TARGET

iOS

<120 MB

---

Android

<150 MB

---

# CPU TARGET

Idle

<2%

---

# NETWORK TARGET

Dashboard

<50 KB

---

Portfolio

<30 KB

---

Asset Card

<20 KB

---

# SCROLL PERFORMANCE

60 FPS

---

# LISTS

Virtualized

Lazy

Infinite

---

# DARK MODE

Default

---

Light Mode

Optional

---

# TABLET MODE

Не растягивать интерфейс.

---

Использовать:

Master

↓

Detail

Layout

---

# FOLDABLE DEVICES

Поддерживаются.

---

# AI ASSISTANT

Работает

не локально,

а через Backend.

---

Никаких LLM

в мобильном приложении.

---

# SECURITY

Private Mode

полностью поддерживается.

---

Никакие налоговые документы

не остаются

в памяти приложения.

---

После просмотра:

```
PDF

↓

Close

↓

Memory Clear
```

---

# CRASH STRATEGY

Приложение никогда

не должно падать

из-за отсутствия сети.

---

# ANALYTICS

Privacy First.

---

Без Fingerprint.

---

Без скрытого отслеживания.

---

# COMPETITIVE ANALYSIS

Лучшие практики,

которые необходимо использовать:

---

Apple Wallet

(скорость)

---

Trading212

(простота)

---

Yahoo Finance

(карточки)

---

BKS

(структура портфеля)

---

Alfa Investments

(визуализация)

---

T-Invest

(навигация)

---

Portfolio Performance

(аналитика)

---

# UNIQUE ADVANTAGE

OpenInvest должен ощущаться как:

```
Apple Wallet

+

Notion

+

Trading212

+

Персональный финансовый помощник
```

а не как перегруженный брокерский терминал.

---

# MOBILE DESIGN CHECKLIST

Перед созданием любого экрана Builder Agent обязан ответить:

1.

Можно ли открыть экран одной рукой?

2.

Можно ли понять экран за 3 секунды?

3.

Можно ли убрать 30% элементов?

4.

Не расходует ли экран батарею без необходимости?

5.

Не делает ли экран лишние сетевые запросы?

6.

Работает ли экран полностью при медленном интернете?

7.

Можно ли использовать экран в самолете без сети?

8.

Будет ли этот экран быстрым через 5 лет, когда у пользователя будет 20 портфелей и 5000 операций?

---

# FINAL MOBILE PRINCIPLE

> **OpenInvest не должен быть самым функциональным мобильным инвестиционным приложением.**

> **Он должен быть самым быстрым, самым понятным, самым спокойным и самым энергоэффективным приложением для долгосрочного инвестора.**
