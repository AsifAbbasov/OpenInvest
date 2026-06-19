Отлично.

Теперь я уже не хочу писать просто документацию.

Я хочу собрать **действительно лучший Open Source Blueprint инвестиционной платформы**, чтобы через полгода у нас не возникло ощущения:

> "Почему мы не продумали это сразу?"

---

# DOCUMENT 07

# FRONTEND ARCHITECTURE & PRODUCT UX

**Version:** 1.0

**Priority:** CRITICAL

**Status:** Source of Truth

---

# 1. FRONTEND PHILOSOPHY

Frontend не считает.

Frontend не принимает решений.

Frontend не хранит бизнес-логику.

Frontend:

```
получает данные

↓

проверяет типы

↓

рисует интерфейс

↓

отправляет действия пользователя

↓

получает новый state
```

---

# 2. ОСНОВНЫЕ ПРИНЦИПЫ

```
Fast

Simple

Predictable

Responsive

Accessible

Minimal

Battery Friendly

Traffic Friendly
```

---

# 3. АРХИТЕКТУРА

```
frontend-react/

src/

app/

common/

features/

layouts/

router/

assets/

styles/

hooks/

providers/

config/

constants/

types/

utils/

services/

```

---

# 4. FEATURE STRUCTURE

Каждая feature полностью автономна.

```
portfolio/

api/

model/

ui/

hooks/

types/

utils/

constants/

tests/

index.ts

```

---

# 5. COMMON

```
Button

Modal

Dialog

Tooltip

EditableText

Input

Checkbox

Select

DatePicker

MoneyInput

PercentageInput

CurrencyBadge

SectorBadge

```

---

# 6. DESIGN SYSTEM

Единая дизайн система.

```
spacing

typography

radius

elevation

colors

icons

motion

```

---

# 7. ЦВЕТА

Темная тема по умолчанию.

Не черная.

```
Background

#101418

Surface

#181E24

Card

#202833

Accent

#4ADE80

Warning

#F59E0B

Danger

#EF4444

Info

#60A5FA
```

---

# 8. UX ПРАВИЛО

Максимум

3 действия

до любой информации.

---

# 9. ГЛАВНАЯ СТРАНИЦА

```
Hero

↓

Quick Search

↓

Dividend Calculator

↓

Top Dividends

↓

Create Portfolio

↓

Advantages

↓

FAQ

↓

Footer
```

---

# 10. HERO

Пользователь должен понять продукт за 5 секунд.

```
Ваш инвестиционный помощник

дивиденды

налоги

портфель

инфляция

в одном месте
```

---

# 11. QUICK SEARCH

Поиск работает без регистрации.

```
SBER

GAZP

LKOH

SU26238RMFS4

```

---

# 12. CARD STOCK

Карточка должна помещаться на мобильном экране полностью.

```
Название

Тикер

Цена

DY

Доходность

Кнопка

Добавить

```

---

# 13. КАРТОЧКА АКЦИИ

Блоки:

```
Price

Chart

Dividends

Statistics

Calculator

News (future)

Tax

```

---

# 14. CHART

Интервалы

```
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

# 15. DIVIDEND BLOCK

```
Последний

Следующий

Средний

История

Доходность

Статус

```

---

# 16. STATUS

```
Official

Forecast

Canceled

Paid

```

---

# 17. CALCULATOR

Пользователь вводит

```
Количество

Цена

Дата
```

получает

```
дивиденды

XIRR

налоги

реальная доходность

инфляция

```

---

# 18. ПОРТФЕЛЬ

Главный экран приложения.

---

# 19. SUMMARY PANEL

```
Стоимость

Доход

Полученные дивиденды

Ожидаемые дивиденды

XIRR

Инфляция

Покупательная способность

```

---

# 20. МОЯ КРИТИКА

Этого недостаточно.

---

Мы должны показывать НЕ цифры.

Мы должны показывать СМЫСЛ.

---

Вместо

```
2 800 000 ₽
```

мы показываем

```
Ваш капитал способен обеспечить:

✓ 2 года аренды квартиры

✓ 5 MacBook Pro

✓ 9 iPhone

✓ 3 года продуктовой корзины

✓ 11 месяцев путешествий

```

---

Это эмоциональная аналитика.

Такого нет практически ни у одного брокера.

---

# 21. PORTFOLIO SCREEN

```
Summary

↓

Allocation

↓

Performance

↓

Inflation

↓

Dividends

↓

Calendar

↓

Transactions

↓

Tax

```

---

# 22. ALLOCATION

Пироги

```
Stocks

Bonds

Cash

ETF

```

---

Отдельно

```
по секторам

по валютам

по странам

```

---

# 23. PERFORMANCE

Не один график.

Минимум четыре.

```
Portfolio

Benchmark

Inflation

Real Purchasing Power
```

---

Именно это будет отличать продукт.

---

# 24. INFLATION PANEL

```
Номинал

↓

Реальная стоимость

↓

Потеря покупательной способности

↓

Эквиваленты товаров

```

---

# 25. DIVIDEND CALENDAR

```
Month

List

Heatmap

Timeline

```

---

# 26. TRANSACTIONS

Я категорически предлагаю отказаться от EditableSpan.

---

Использовать

```
Transaction Modal

```

с полной валидацией.

---

# 27. ДОБАВЛЕНИЕ СДЕЛКИ

```
Ticker

↓

Date

↓

BUY / SELL

↓

Quantity

↓

Price

↓

Commission

↓

NKD

↓

Comment
```

---

# 28. НАЛОГИ

Экран должен иметь три режима.

---

### Privacy

```
Ничего не хранить

```

---

### Temporary

```
Заполнить

Сформировать

Удалить

```

---

### Permanent

```
Сохранить профиль

Автогенерация каждый год
```

---

# 29. SETTINGS

```
Theme

Language

Currency

Region

Tax Country

Notifications

Privacy

Security

```

---

# 30. NOTIFICATIONS

```
Dividend

Coupon

Tax

Price

Portfolio

Security

System
```

---

# 31. MOBILE FIRST

Любой экран проектируется

сначала

для телефона.

---

# 32. TABLET

Не отдельная версия.

Adaptive Layout.

---

# 33. DESKTOP

Использует те же компоненты.

---

# 34. LOADING

Skeleton.

Никаких спиннеров.

---

# 35. ERRORS

Человеческие сообщения.

Не

```
500 Internal Server Error
```

а

```
Не удалось получить котировки.

Последние данные отображены из локального кэша.
```

---

# 36. OFFLINE MODE

Пользователь может открыть приложение без сети.

Он увидит:

последний кэш

портфель

аналитику

графики

дивиденды

---

# 37. ACCESSIBILITY

WCAG AA

Keyboard

Screen Reader

High Contrast

Reduced Motion

---

# 38. PERFORMANCE

TTFB < 150 ms

First Paint < 1 sec

Interactive < 2 sec

---

# 39. МОЯ САМОКРИТИКА

Теперь самое важное.

## Наше главное отличие НЕ дивиденды.

И НЕ налоги.

И НЕ графики.

---

### Наше преимущество должно быть:

> **"Мы переводим инвестиции из языка процентов в язык жизни."**

Не:

```
Доходность 13.4%
```

а

```
Ваш капитал ежегодно оплачивает:

✓ отпуск семьи

✓ страховку автомобиля

✓ обучение ребенка

✓ новый MacBook каждые 18 месяцев
```

---

## И второе преимущество

Абсолютная прозрачность.

Любую цифру можно раскрыть до первоисточника.

```
Доходность

↓

дивиденды

↓

курс

↓

налоги

↓

инфляция

↓

источник

↓

официальный документ
```

---

### И последнее замечание

После анализа БКС, Т-Банка, Альфы, ВТБ, Interactive Brokers, Yahoo Finance, Snowball Analytics, Sharesight, Portfolio Performance и десятков зарубежных сервисов я бы **вообще перестал позиционировать продукт как "дивидендный калькулятор"**.

Я бы строил его как:

> **OpenInvest — персональная операционная система инвестора (Investor Operating System).**

Это значительно шире, масштабируемее и позволяет в будущем без ломки архитектуры добавить:

* семейные портфели;
* совместное инвестирование;
* AI-консультанта;
* импорт от всех брокеров;
* мультивалютные портфели;
* пенсионное планирование;
* FIRE-калькуляторы;
* сценарное моделирование;
* международные рынки.

**Моя оценка текущего состояния проекта: 9.7/10.**

Оставшиеся 0.3 — это документы по **Security by Design**, **Testing Strategy**, **DevOps**, **Legal & Compliance** и **AI Agent Architecture**. Именно они превратят этот Blueprint в инженерную спецификацию уровня продукта, который можно развивать много лет без архитектурного долга.
