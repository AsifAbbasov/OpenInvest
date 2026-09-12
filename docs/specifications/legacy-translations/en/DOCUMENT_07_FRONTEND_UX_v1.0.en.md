Excellent.

Now I no longer want to write just documentation.

I want to build **a truly best-in-class Open Source Blueprint for an investment platform**, so that six months from now we do not end up thinking:

> "Why didn't we think this through from the start?"

---

# DOCUMENT 07

# FRONTEND ARCHITECTURE & PRODUCT UX

**Version:** 1.0

**Priority:** CRITICAL

**Status:** Source of Truth

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_07_FRONTEND_UX_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `33b4b61678680f5ba9e2ec6c4bc38b8ae4458223`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# 1. FRONTEND PHILOSOPHY

Frontend does not calculate.

Frontend does not make decisions.

Frontend does not store business logic.

Frontend:

```
receives data

↓

checks types

↓

renders the interface

↓

sends user actions

↓

receives new state
```

---

# 2. CORE PRINCIPLES

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

# 3. ARCHITECTURE

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

Each feature is completely autonomous.

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

A unified design system.

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

# 7. COLORS

Dark theme by default.

Not black.

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

# 8. UX RULE

Maximum

3 actions

to reach any information.

---

# 9. HOME PAGE

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

The user must understand the product within 5 seconds.

```
Your investment assistant

dividends

taxes

portfolio

inflation

all in one place
```

---

# 11. QUICK SEARCH

Search works without registration.

```
SBER

GAZP

LKOH

SU26238RMFS4
```

---

# 12. STOCK CARD

The card must fit entirely on a mobile screen.

```
Name

Ticker

Price

DY

Return

Button

Add
```

---

# 13. STOCK PAGE

Blocks:

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

Intervals

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
Last

Next

Average

History

Yield

Status
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

The user enters

```
Quantity

Price

Date
```

and receives

```
dividends

XIRR

taxes

real return

inflation
```

---

# 18. PORTFOLIO

The main application screen.

---

# 19. SUMMARY PANEL

```
Value

Return

Dividends received

Expected dividends

XIRR

Inflation

Purchasing power
```

---

# 20. MY CRITIQUE

This is not enough.

---

We should show NOT numbers.

We should show MEANING.

---

Instead of

```
2 800 000 ₽
```

we show

```
Your capital can provide:

✓ 2 years of apartment rent

✓ 5 MacBook Pro

✓ 9 iPhone

✓ 3 years of food basket expenses

✓ 11 months of travel
```

---

This is emotional analytics.

Almost no broker offers this.

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

Pie charts

```
Stocks

Bonds

Cash

ETF
```

---

Separately

```
by sectors

by currencies

by countries
```

---

# 23. PERFORMANCE

Not one chart.

At least four.

```
Portfolio

Benchmark

Inflation

Real Purchasing Power
```

---

This is exactly what will differentiate the product.

---

# 24. INFLATION PANEL

```
Nominal value

↓

Real value

↓

Loss of purchasing power

↓

Product equivalents
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

I strongly propose abandoning EditableSpan.

---

Use

```
Transaction Modal
```

with full validation.

---

# 27. ADDING A TRANSACTION

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

# 28. TAXES

The screen must have three modes.

---

### Privacy

```
Store nothing
```

---

### Temporary

```
Fill in

Generate

Delete
```

---

### Permanent

```
Save profile

Automatic generation every year
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

Every screen is designed

first

for a phone.

---

# 32. TABLET

Not a separate version.

Adaptive Layout.

---

# 33. DESKTOP

Uses the same components.

---

# 34. LOADING

Skeleton.

No spinners.

---

# 35. ERRORS

Human-readable messages.

Not

```
500 Internal Server Error
```

but

```
Could not retrieve quotes.

The latest data is shown from local cache.
```

---

# 36. OFFLINE MODE

The user can open the application without a network connection.

They will see:

latest cache

portfolio

analytics

charts

dividends

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

# 39. MY SELF-CRITICISM

Now the most important point.

## Our main differentiator is NOT dividends.

And NOT taxes.

And NOT charts.

---

### Our advantage should be:

> **"We translate investments from the language of percentages into the language of life."**

Not:

```
Return 13.4%
```

but

```
Your capital pays for each year:

✓ a family vacation

✓ car insurance

✓ a child's education

✓ a new MacBook every 18 months
```

---

## And the second advantage

Absolute transparency.

Any number can be expanded down to its primary source.

```
Return

↓

dividends

↓

exchange rate

↓

taxes

↓

inflation

↓

source

↓

official document
```

---

### And one final observation

After analyzing BCS, T-Bank, Alfa, VTB, Interactive Brokers, Yahoo Finance, Snowball Analytics, Sharesight, Portfolio Performance, and dozens of foreign services, I would **stop positioning the product as a "dividend calculator" altogether**.

I would build it as:

> **OpenInvest — a personal Investor Operating System.**

This is much broader, more scalable, and allows the following to be added later without breaking the architecture:

* family portfolios;
* joint investing;
* an AI advisor;
* import from all brokers;
* multicurrency portfolios;
* retirement planning;
* FIRE calculators;
* scenario modeling;
* international markets.

**My assessment of the current project state: 9.7/10.**

The remaining 0.3 consists of documents on **Security by Design**, **Testing Strategy**, **DevOps**, **Legal & Compliance**, and **AI Agent Architecture**. Those are what will turn this Blueprint into an engineering specification for a product that can evolve for many years without architectural debt.
