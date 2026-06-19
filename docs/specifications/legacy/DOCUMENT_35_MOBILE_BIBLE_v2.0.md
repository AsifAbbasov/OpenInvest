# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 35

# MOBILE ARCHITECTURE BIBLE

# iOS (SwiftUI), Android (Jetpack Compose), Offline First, Sync Engine, Battery Optimization & Cross Platform Strategy

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: MOBILE CONSTITUTION

---

# PURPOSE

Настоящий документ определяет архитектуру мобильных приложений OpenInvest.

Мобильное приложение не является адаптацией веб-версии.

Оно является самостоятельным клиентом единой платформы Personal Capital OS.

---

# PHILOSOPHY

OpenInvest Mobile должен быть:

самым быстрым;

самым энергоэффективным;

самым отзывчивым;

самым понятным;

самым безопасным.

---

# MOBILE FIRST

Любая новая функция проектируется:

```text id="m1"
Phone

↓

Tablet

↓

Desktop

↓

Web
```

---

Запрещено проектировать Desktop,

а потом уменьшать его до телефона.

---

# NATIVE ONLY

Используется:

---

iOS

SwiftUI

---

Android

Jetpack Compose

---

Запрещено:

Flutter

React Native

Xamarin

Cordova

для основной версии продукта.

---

# WHY NATIVE

Причины:

---

Минимальное потребление памяти.

---

Минимальное потребление батареи.

---

Максимальная скорость.

---

Лучший UX.

---

Лучшая интеграция с системой.

---

# ARCHITECTURE

Каждое приложение имеет одинаковую структуру.

```text id="m2"
Presentation

↓

Application

↓

Domain

↓

Data

↓

Network

↓

Storage
```

---

# SHARED CONTRACT

Web

↓

iOS

↓

Android

используют

один OpenAPI контракт.

---

# SDK

SDK генерируется автоматически.

---

TypeScript

↓

Swift

↓

Kotlin

---

# OFFLINE FIRST

Пользователь должен видеть портфель

даже без Интернета.

---

# LOCAL CACHE

Хранятся:

---

Assets

---

Snapshots

---

Portfolio

---

Settings

---

Notifications

---

# NEVER CACHE

Пароль

JWT

Refresh

XML

PDF

ИНН

Паспорт

---

# SYNC STRATEGY

```text id="m3"
Open App

↓

Load Local Snapshot

↓

Show UI

↓

Background Sync

↓

Refresh Changed Data
```

---

Пользователь никогда не ждет полной загрузки.

---

# DELTA SYNC

Передаются

только измененные данные.

---

Никаких полных повторных загрузок.

---

# BATTERY OPTIMIZATION

Background Tasks

используются минимально.

---

Запрещено:

каждые 30 секунд

опрашивать сервер.

---

Используются:

---

Push Trigger

---

Silent Push

---

Background Refresh

---

Manual Refresh

---

# PUSH STRATEGY

Push содержит:

---

Dividend Approved

---

Tax Reminder

---

Portfolio Goal

---

Security Alert

---

Никогда:

финансовые суммы.

---

# HOME SCREEN WIDGETS

Поддерживаются:

---

Portfolio Value

---

Today's Change

---

Next Dividend

---

Inflation

---

# DYNAMIC ISLAND

(iOS)

Поддерживается:

---

Dividend Today

---

Tax Reminder

---

Portfolio Update

---

# APP SHORTCUTS

Долгое нажатие:

---

Portfolio

---

Dividend Calendar

---

Tax Export

---

Search Asset

---

# BIOMETRICS

Face ID

Touch ID

Fingerprint

---

используются

только как удобство.

---

Не заменяют пароль.

---

# LOCAL DATABASE

SQLite

или

Realm

---

используется только как Cache.

---

Источник истины —

Backend.

---

# CONFLICT RESOLUTION

Если пользователь изменил данные

на двух устройствах:

```text id="m4"
Server Wins

+

Conflict Log

+

User Notification
```

---

# CHARTS

Используются:

Native Charts

или

максимально легкие библиотеки.

---

# ANIMATIONS

60 FPS.

---

Максимум:

200 ms.

---

# MEMORY TARGET

iOS

<120 MB

---

Android

<150 MB

---

# STARTUP TARGET

Cold Start

<1.5 sec

---

Warm Start

<500 ms

---

# NETWORK TARGET

Открытие Dashboard

<100 KB

---

Открытие Portfolio

<150 KB

---

Обновление Snapshot

<20 KB

---

# IMAGE STRATEGY

SVG

↓

WebP

↓

PNG

---

# FONT STRATEGY

Используются системные шрифты.

---

Не загружать кастомные.

---

# ACCESSIBILITY

VoiceOver

---

TalkBack

---

Large Text

---

Reduce Motion

---

High Contrast

---

# TABLET MODE

Используется

Split Layout.

---

# DARK MODE

По умолчанию.

---

Light —

поддерживается.

---

# APP SECURITY

Root Detection

---

Jailbreak Detection

---

Certificate Pinning

---

Secure Storage

---

# CRASH REPORTING

Без передачи персональных данных.

---

# ANALYTICS

Собираются:

Screen Open

↓

Duration

↓

Crash

↓

Performance

---

Без финансовых данных.

---

# APP STORE STRATEGY

iOS

↓

TestFlight

↓

Staged Rollout

↓

Production

---

Android

↓

Internal

↓

Closed

↓

Open

↓

Production

---

# REVIEW CHECKLIST

Перед каждым релизом:

---

Battery

---

Memory

---

Offline

---

Push

---

Accessibility

---

Performance

---

Security

---

UX

---

# FUTURE READY

Архитектура должна позволять добавить:

---

Apple Watch

---

Wear OS

---

macOS

---

visionOS

---

Android Auto (notifications only)

---

# FINAL MOBILE PRINCIPLE

> **OpenInvest Mobile не должен ощущаться как веб-сайт внутри телефона.**

> **Он должен работать как нативное приложение уровня Apple Stocks, Wealthfront или Robinhood: мгновенно открываться, практически не расходовать заряд батареи, работать без сети и отображать последние актуальные данные даже в самолете.**

> **Телефон должен быть клиентом Personal Capital OS, а не вычислительным устройством. Все тяжелые расчеты выполняются сервером, а мобильное приложение отвечает исключительно за быстрый, понятный и надежный пользовательский опыт.**
