# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 19

# FRONTEND ARCHITECTURE, UX, UI SYSTEM, MOBILE FIRST & DESIGN SYSTEM

Version: 1.0

Status: Approved

Priority: Critical

---

# PURPOSE

Настоящий документ определяет единственную допустимую архитектуру Frontend OpenInvest.

Он является обязательным для:

* Builder Agent
* Review Agent
* QA Agent
* Codex

Никакие UI-решения не могут приниматься без соответствия данному документу.

---

# PRODUCT PHILOSOPHY

OpenInvest — это не терминал трейдера.

OpenInvest — это спокойный инвестиционный помощник.

---

Пользователь должен понимать состояние своего капитала за 3 секунды.

---

# DESIGN PRINCIPLES

Максимум информации

Минимум шума.

---

Большие цифры.

Минимум текста.

---

Пользователь никогда не должен искать прибыль.

Она должна быть первым элементом экрана.

---

# UX PRINCIPLES

No Stress

No Noise

No Popups

No Banner Hell

No Ads in Core Flow

---

# MOBILE FIRST

Сначала проектируется мобильная версия.

После этого:

Tablet

Desktop

Wide Screen

---

# SCREEN GRID

```text id="1"
Mobile

390 px

↓

Tablet

768 px

↓

Desktop

1280 px

↓

Wide

1600+
```

---

# COLOR SYSTEM

Dark Theme Default

---

Background

Almost Black

---

Surface

Dark Gray

---

Accent

Green

---

Negative

Red

---

Neutral

Blue

---

Warning

Orange

---

# TYPOGRAPHY

Одна гарнитура.

---

Display

Большие цифры.

---

Body

Минимальный текст.

---

Caption

Вторичная информация.

---

# SPACING

Использовать

8px grid.

---

# ANIMATIONS

Максимум

150ms

---

Никаких тяжелых анимаций.

---

# NAVIGATION

Bottom Navigation

Mobile

---

Sidebar

Desktop

---

# MAIN NAVIGATION

```text id="2"
Dashboard

Portfolio

Catalog

Calendar

Taxes

Profile
```

---

# FIRST SCREEN

После входа пользователь должен увидеть:

---

Стоимость портфеля

---

XIRR

---

Полученные дивиденды

---

Ожидаемые дивиденды

---

Инфляционную стоимость

---

Покупательную способность

---

# HERO CARD

Самая большая карточка приложения.

---

Отображает:

```text id="3"
2 543 220 ₽

+

12.4%

+

XIRR 18.2%

+

Real Value
```

---

# QUICK ACTIONS

Под Hero Card

---

Добавить сделку

---

Добавить деньги

---

Посмотреть дивиденды

---

Скачать налоговый отчет

---

# DASHBOARD BLOCKS

Стоимость

↓

Доходность

↓

Дивиденды

↓

Календарь

↓

Портфель

↓

Новости

---

# NEWS

Показываются

ТОЛЬКО

по активам пользователя.

---

Никакой общей ленты.

---

# PORTFOLIO SCREEN

Содержит:

---

Summary

---

Allocation

---

Chart

---

Transactions

---

Analytics

---

# SUMMARY PANEL

Показывает:

```text id="4"
Стоимость

Прибыль

Убыток

XIRR

Real Return

Inflation

Cash
```

---

# ALLOCATION

Pie Chart

---

Stocks

---

Bonds

---

Cash

---

ETF

---

# CHARTS

Использовать

Recharts.

---

Минимум элементов.

---

Без 3D.

---

Без теней.

---

# TIME FILTER

```text id="5"
1D

1W

1M

3M

6M

1Y

3Y

5Y

ALL
```

---

# PORTFOLIO TABLE

Columns

---

Ticker

---

Quantity

---

Average Cost

---

Current Price

---

Profit

---

Dividend Yield

---

XIRR

---

# TRANSACTION UX

EditableSpan

НЕ используется.

---

Используется

Transaction Modal.

---

# TRANSACTION MODAL

Fields

Ticker

Quantity

Price

Commission

NKD

Date

Comment

---

Validation

Backend

*

Frontend

---

# STOCK CARD

Показывает:

---

Название

---

Цена

---

Изменение

---

Сектор

---

Капитализация

---

Dividend Yield

---

# STOCK CHART

Минималистичный.

---

Никаких

MACD

RSI

Bollinger

---

OpenInvest —

не терминал.

---

# DIVIDEND BLOCK

История

↓

Официальные

↓

Прогноз

↓

Среднее за 5 лет

↓

Growth

---

# DIVIDEND CALCULATOR

Inputs

Quantity

Buy Price

Buy Date

---

Outputs

Получено

Ожидается

Yield on Cost

XIRR

Real Return

---

# CALENDAR

Modes

---

List

---

Calendar

---

Heatmap

---

# HEATMAP

Intensity

=

Dividend Amount

---

# TAX SCREEN

Главная идея

—

максимальная простота.

---

# TAX SCREEN FLOW

```text id="6"
Year

↓

Calculate

↓

Review

↓

Download

↓

Email
```

---

# REVIEW SCREEN

Human in the Loop

---

Показывает:

ИНН

Адрес

Паспорт

Доход

Налог

---

Пользователь подтверждает.

---

# PRIVACY MODE

Если выбран

Private Mode

---

ИНН

не сохраняется.

---

Паспорт

не сохраняется.

---

Адрес

не сохраняется.

---

# PROFILE SCREEN

Содержит:

---

Theme

---

Language

---

Privacy

---

Notifications

---

Export Data

---

Delete Profile

---

# EXPORT SCREEN

Formats

PDF

XML

CSV

JSON

ZIP

---

# PURCHASING POWER

Уникальная функция.

---

Карточка:

```text id="7"
Ваш портфель

2 400 000 ₽

=

2.8 MacBook Pro

или

40 месяцев ЖКХ

или

68 продуктовых корзин
```

---

# REAL VALUE CARD

Показывает:

---

Номинальная стоимость

↓

Реальная стоимость

↓

Потеря из-за инфляции

---

# AI PANEL

Не советует покупать.

---

Объясняет:

что изменилось.

---

# ACCESSIBILITY

Font Scale

100%

125%

150%

---

Contrast AAA

---

Keyboard Navigation

---

Screen Reader

---

# PERFORMANCE

Target

60 FPS

---

Initial Load

<2 sec

---

Interaction

<100 ms

---

# IMAGES

Lazy Loading

---

# CHARTS

Virtual Rendering

---

# TABLES

Virtual Scroll

---

# STATE

Redux Toolkit

---

Server Data

React Query

---

# ERROR UI

Никогда

не показывать

stack trace.

---

Только понятный текст.

---

# EMPTY STATES

Не должно быть пустых экранов.

---

Если нет портфеля:

```text id="8"
Создайте первый портфель

Добавьте первую акцию

Посмотрите пример
```

---

# LOADING

Skeleton

---

Не Spinner.

---

# DESIGN REFERENCES

Использовать лучшие практики:

BKS

Alfa Investments

T-Invest

VTB

Trading212

Portfolio Performance

Yahoo Finance

Apple Wallet

Notion

Linear

---

# MAIN PRODUCT ADVANTAGE

Пользователь должен чувствовать,

что приложение:

не давит,

не продает,

не заставляет,

не шумит,

а спокойно объясняет,

что происходит с его капиталом.

---

# CODEX REQUIREMENTS

Перед созданием любого нового экрана Builder Agent обязан ответить:

1. Можно ли понять экран за 5 секунд?

2. Можно ли убрать 30% элементов без потери функциональности?

3. Является ли самая важная цифра самой большой?

4. Не копирует ли экран брокерский терминал?

5. Можно ли пользоваться экраном одной рукой на телефоне?

6. Соответствует ли он Mobile First?

7. Не увеличивает ли он потребление батареи и мобильного трафика?

8. Не нарушает ли философию OpenInvest:

**"Максимум полезной информации при минимуме визуального шума."**

Только после прохождения этих проверок экран допускается к реализации.
