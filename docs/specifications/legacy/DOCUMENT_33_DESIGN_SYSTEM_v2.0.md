# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 33

# DESIGN SYSTEM, UI/UX CONSTITUTION, VISUAL LANGUAGE, ACCESSIBILITY, MOBILE FIRST & INTERACTION STANDARDS

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: DESIGN SYSTEM

---

# PURPOSE

Настоящий документ определяет единый язык интерфейсов OpenInvest.

Дизайн рассматривается не как украшение продукта, а как инструмент передачи финансовой информации.

Пользователь должен понимать состояние своего капитала за 3–5 секунд без изучения инструкций.

---

# DESIGN PHILOSOPHY

OpenInvest не должен быть похож:

* на банковское приложение;
* на терминал трейдера;
* на бухгалтерскую программу.

---

OpenInvest должен выглядеть как:

> **Personal Capital Dashboard**

Минималистичный, спокойный, быстрый, понятный.

---

# CORE PRINCIPLES

Принцип №1

Показывать только важное.

---

Принцип №2

Графика важнее текста.

---

Принцип №3

Одно действие — одна цель.

---

Принцип №4

Минимум модальных окон.

---

Принцип №5

Любая операция выполняется максимум за 3 клика.

---

# DESIGN LANGUAGE

Основа:

Apple Human Interface

↓

Linear

↓

Stripe

↓

Notion

↓

Wealthfront

↓

Robinhood

↓

без копирования,

только лучшие UX-практики.

---

# COLOR SYSTEM

По умолчанию:

Dark Theme

---

Background

```text id="t41e0w"
#0B0F14
```

---

Surface

```text id="ep79l8"
#121821
```

---

Primary

```text id="9bxn2z"
#2D8CFF
```

---

Success

```text id="lfaljw"
#18C964
```

---

Warning

```text id="ckvcdo"
#FFB224
```

---

Danger

```text id="b3jfbq"
#F31260
```

---

# LIGHT THEME

Поддерживается,

но вторична.

---

# TYPOGRAPHY

Используется одна система.

---

Display

48

---

H1

36

---

H2

28

---

H3

22

---

Body

16

---

Caption

14

---

Small

12

---

# GRID

Используется:

8pt Grid

---

Spacing:

8

16

24

32

48

64

---

# BORDER RADIUS

Cards

16

---

Buttons

12

---

Inputs

12

---

Charts

20

---

# SHADOWS

Минимальные.

---

Не использовать тяжелые Material-тени.

---

# ANIMATIONS

Продолжительность:

150–250 ms

---

Никаких:

bounce

elastic

overshoot

---

Использовать:

Fade

Scale

Slide

Opacity

---

# NAVIGATION

Bottom Navigation

Mobile

---

Sidebar

Desktop

---

# MAIN SCREENS

Dashboard

↓

Portfolio

↓

Assets

↓

Calendar

↓

Analytics

↓

Tax

↓

Profile

---

# DASHBOARD STRUCTURE

Верхняя карточка:

```text id="ulh9ik"
Portfolio Value

Real Value

Daily Change
```

---

Вторая карточка:

```text id="cq5y0x"
Expected Dividends

Coupons

Tax Forecast
```

---

Третья:

```text id="99lm6d"
Inflation

Purchasing Power

Real Return
```

---

# PORTFOLIO SCREEN

Показывает:

---

Стоимость

---

Денежные средства

---

Акции

---

Облигации

---

Доходность

---

XIRR

---

Пирог аллокации

---

История

---

# ASSET CARD

Обязательно:

Название

↓

Цена

↓

Изменение

↓

Дивиденды

↓

Yield

↓

CAGR

↓

История

↓

Калькулятор

---

# DIVIDEND CALENDAR

Два режима:

---

Calendar

---

List

---

Поддерживаются:

цветовые метки;

фильтры;

поиск;

группировка.

---

# CHART PRINCIPLES

Charts

не должны быть перегружены.

---

Максимум:

4 линии.

---

Максимум:

6 цветов.

---

# DEFAULT CHARTS

Portfolio

↓

Real Return

↓

Dividend History

↓

Allocation

↓

Purchasing Power

---

# EMPTY STATES

Запрещено:

пустой экран.

---

Показывать:

пример;

демо;

объяснение;

CTA.

---

# LOADING

Используются Skeleton.

---

Запрещены Spinner

для долгих экранов.

---

# FORMS

Минимум полей.

---

Добавление сделки:

Ticker

↓

Quantity

↓

Price

↓

Commission(optional)

↓

NKD(optional)

↓

Date

---

# INPUT RULES

Автоматические маски.

---

Автоматические разделители.

---

Автоматическое форматирование валют.

---

# MODALS

Используются только:

---

Добавление сделки

---

Удаление

---

Экспорт

---

Настройки

---

# TABLES

Desktop

↓

Table

---

Mobile

↓

Cards

---

# MOBILE FIRST

Каждый экран сначала проектируется

для телефона.

---

Потом

адаптируется под Desktop.

---

# RESPONSIVE BREAKPOINTS

360

480

768

1024

1280

1536

---

# ACCESSIBILITY

Обязательно:

WCAG AA

---

Keyboard Navigation

---

Screen Reader

---

Focus Ring

---

Contrast

---

# MOTION REDUCE

Поддерживается системная настройка:

Reduce Motion.

---

# ICON SYSTEM

Одна библиотека.

---

Не смешивать:

Heroicons

Lucide

Material

Feather

одновременно.

---

# UX PRINCIPLE

Пользователь никогда не должен считать самостоятельно.

---

Вместо:

```text id="eplksn"
+14.72%
```

показывать:

```text id="w6ptv7"
Ваш капитал вырос быстрее инфляции на 8.4%
```

---

# REAL VALUE CARD

Эксклюзивная карточка.

---

Показывает:

```text id="0yhjln"
Сегодня ваш капитал:

11 MacBook Pro

или

26 iPhone

или

31 месяца средней продуктовой корзины
```

---

# PREMIUM UX

Premium

не должен менять интерфейс.

---

Он должен добавлять:

новые аналитические карточки.

---

# PERFORMANCE

Dashboard

<1.5 sec

---

Navigation

<100 ms

---

Chart

<16 ms render

---

# DESIGN REVIEW CHECKLIST

Перед созданием нового экрана Designer Agent отвечает:

---

Можно ли убрать один блок?

---

Можно ли убрать одну кнопку?

---

Можно ли убрать один текст?

---

Можно ли показать это графически?

---

Можно ли объединить две карточки?

---

# FINAL DESIGN PRINCIPLE

> **OpenInvest должен восприниматься не как сложный финансовый терминал, а как персональная панель управления капиталом.**

> **Пользователь не должен читать таблицы и считать проценты — интерфейс обязан самостоятельно преобразовывать финансовые данные в понятные визуальные выводы, реальные жизненные эквиваленты и простые решения.**

> **Лучший экран — это экран, на котором пользователь за пять секунд понимает состояние своего капитала, ничего не вычисляя самостоятельно.**
