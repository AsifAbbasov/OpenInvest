# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 14

# FRONTEND ARCHITECTURE, UX/UI SYSTEM & DESIGN BLUEPRINT

Version: 1.0

Status: Core Client Architecture

Priority: Critical

---

# 1. PURPOSE

Настоящий документ определяет:

* архитектуру Frontend;
* структуру проекта;
* навигацию;
* UX;
* UI;
* взаимодействие пользователя;
* правила написания компонентов;
* правила написания состояния приложения.

Документ является обязательным для:

Builder Agent

Review Agent

QA Agent

Android Team

iOS Team

---

# 2. FRONTEND PHILOSOPHY

Frontend никогда не должен быть "умным".

Frontend является отображением состояния Backend.

Вся тяжелая логика выполняется сервером.

Frontend отвечает только за:

отображение;

анимации;

навигацию;

валидацию пользовательского ввода;

локальное состояние UI.

---

# 3. STACK

React 19

TypeScript

Vite

Redux Toolkit 2

React Router

TailwindCSS

React Hook Form

Zod

Axios

Recharts

Framer Motion

TanStack Virtual

---

# 4. PROJECT STRUCTURE

```text
frontend-react/

src/

app/

common/

features/

assets/

styles/

types/

hooks/

providers/

router/

config/
```

---

# 5. FEATURE STRUCTURE

Каждая feature полностью автономна.

```text
portfolio/

api/

model/

ui/

components/

hooks/

selectors/

types/

utils/

tests/
```

---

# 6. IMPORT RULES

Разрешено

```text
common/*
features/*
app/*
```

---

Запрещено

```text
../../

../../../

../../../../
```

---

# 7. COMPONENT RULES

Компонент отвечает только за одну задачу.

Максимум:

300 строк.

Идеально:

100–150 строк.

---

# 8. CONTAINER / PRESENTATION

Container

↓

получает данные

↓

Presentation

↓

рисует интерфейс

---

# 9. DESIGN PHILOSOPHY

Минимум текста.

Максимум информации.

Максимум воздуха.

Максимум скорости.

---

# 10. STYLE

Темная тема

по умолчанию.

---

Возможность:

Dark

Light

Auto

---

# 11. COLOR SYSTEM

Background

Gray-950

---

Cards

Gray-900

---

Borders

Gray-800

---

Positive

Emerald

---

Negative

Red

---

Warning

Amber

---

Information

Blue

---

# 12. TYPOGRAPHY

Большие цифры.

Минимум декоративного текста.

---

Стоимость портфеля:

56 px

---

Доходность:

32 px

---

Вторичная информация:

14 px

---

# 13. GRID

Desktop

12 columns

---

Tablet

8 columns

---

Mobile

4 columns

---

# 14. SPACING

Использовать только систему:

4

8

12

16

24

32

48

64

---

# 15. ANIMATIONS

Только полезные.

---

Запрещено:

долгие анимации;

анимации ради красоты;

тяжелые transition.

---

# 16. LOADING

Skeleton.

Не Spinner.

---

# 17. EMPTY STATE

Каждый экран обязан иметь:

Empty

Loading

Success

Error

Offline

---

# 18. RESPONSIVE

Поддержка:

320

375

390

768

1024

1280

1440

1920

---

# 19. MAIN NAVIGATION

```text
Home

Catalog

Portfolio

Calendar

Taxes

Profile
```

---

# 20. HOME

Содержит:

поиск тикера

пример расчета

создать портфель

последние дивиденды

CTA регистрация

---

# 21. CATALOG

Поиск

Фильтры

Сортировки

Карточки

Infinite Scroll

---

# 22. STOCK PAGE

Название

Цена

График

Дивиденды

История

Доходность

Калькулятор

Новости (будущее)

---

# 23. PORTFOLIO

Главный экран продукта.

---

Отображает:

Стоимость

XIRR

Real Return

Inflation

Cash

Stocks

Bonds

Expected Dividends

Received Dividends

---

# 24. SUMMARY PANEL

Самая важная карточка.

```text
Portfolio

1 250 000 ₽

+24.7%

XIRR 18.4%

Inflation adjusted

+11.2%
```

---

# 25. PURCHASING POWER CARD

Наша уникальная фича.

Показывает:

```text
Сегодня ваш портфель равен:

2.8 MacBook Pro

5 iPhone Pro

18 средним зарплатам

31 продуктовой корзине

14 месяцам аренды
```

---

Пользователь переключает:

номинал

↓

реальная стоимость

↓

покупательная способность

---

# 26. PORTFOLIO CHART

Диапазоны:

1W

1M

3M

6M

1Y

3Y

5Y

ALL

---

# 27. ALLOCATION

Pie Chart

Stocks

Bonds

Cash

ETF

---

# 28. SECTOR CHART

Финансы

Нефтегаз

IT

Ритейл

Металлургия

---

# 29. DIVIDEND CALENDAR

Режимы:

Calendar

List

Timeline

---

# 30. FILTERS

Official

Forecast

Portfolio only

Selected companies

Month

Quarter

Year

---

# 31. TAX PAGE

Generate XML

Generate PDF

Email

Preview

Human Verification

---

# 32. HUMAN IN THE LOOP

Перед экспортом:

Preview

↓

Confirm

↓

Generate

↓

Download

---

# 33. SETTINGS

Language

Currency

Theme

Notifications

Privacy

Security

Tax Profile

---

# 34. TRUST DASHBOARD

Показывает:

ИНН сохранен?

Паспорт сохранен?

Адрес сохранен?

Последний экспорт

Последнее удаление

Активные устройства

---

# 35. TRANSACTION INPUT

EditableSpan запрещен.

---

Используется:

TransactionModal

---

Поля:

Ticker

Quantity

Price

Commission

NKD

Date

Comment

---

# 36. VALIDATION

React Hook Form

*

Zod

---

никаких ручных проверок.

---

# 37. ACCESSIBILITY

Keyboard

Screen Reader

Contrast

Focus

ARIA

---

обязательно.

---

# 38. PERFORMANCE

Bundle

<250kb

---

Lazy Loading

Code Splitting

Tree Shaking

Virtualization

Memoization

---

# 39. STATE MANAGEMENT

Redux Toolkit

только для:

Portfolio

Catalog

Settings

Notifications

User

---

локальное UI состояние:

useState

---

# 40. API

Axios Instance

единственный.

---

никаких fetch по всему проекту.

---

# 41. ERROR UI

Ошибка должна объяснять:

что произошло;

что делать;

можно ли повторить.

---

# 42. OFFLINE MODE

Последние snapshots

Последние дивиденды

Последний каталог

кэшируются.

---

# 43. MOBILE FIRST

Любая новая функция сначала проектируется для телефона.

Потом масштабируется на Desktop.

---

# 44. FUTURE IOS

SwiftUI

должен повторять:

структуру экранов

цветовую систему

UX

навигацию.

---

# 45. FUTURE ANDROID

Jetpack Compose

полностью повторяет Web.

---

# 46. DESIGN PRINCIPLES

KISS

DRY

SOLID

YAGNI

SRP

Occam Razor

Consistency

Predictability

---

# 47. UX PRINCIPLES

Пользователь должен:

найти бумагу

за 3 секунды;

понять доходность

за 5 секунд;

увидеть дивиденды

за 2 секунды;

создать портфель

менее чем за минуту.

---

# 48. DIFFERENTIATOR

OpenInvest не должен быть похож на терминал.

Он должен ощущаться как:

"финансовый ассистент"

а не

"биржевой торговый стакан".

---

# 49. SUCCESS METRIC

Если новый пользователь без инструкции может:

создать портфель;

понять реальную доходность;

понять влияние инфляции;

получить налоговый XML;

за 5 минут,

то UX считается реализованным правильно.

---

# 50. CODEX REQUIREMENT

Codex обязан реализовывать любой новый экран только после проверки:

1. Не перегружает ли он пользователя.

2. Можно ли убрать половину элементов.

3. Можно ли показать информацию проще.

4. Можно ли выполнить действие быстрее.

5. Соответствует ли экран философии:

**Maximum Information. Minimum Friction. Maximum Trust.**
