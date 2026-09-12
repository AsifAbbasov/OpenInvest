# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 14

# FRONTEND ARCHITECTURE, UX/UI SYSTEM & DESIGN BLUEPRINT

Version: 1.0

Status: Core Client Architecture

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_14A_FRONTEND_ARCHITECTURE_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `574fc583f323c14e8382bef436e89e55c46a779e`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# 1. PURPOSE

This document defines:

* Frontend architecture;
* project structure;
* navigation;
* UX;
* UI;
* user interaction;
* component-writing rules;
* application-state rules.

The document is mandatory for:

Builder Agent

Review Agent

QA Agent

Android Team

iOS Team

---

# 2. FRONTEND PHILOSOPHY

Frontend should never be "smart."

Frontend is a representation of Backend state.

All heavy logic is executed by the server.

Frontend is responsible only for:

display;

animations;

navigation;

validation of user input;

local UI state.

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

Each feature is completely autonomous.

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

Allowed

```text
common/*
features/*
app/*
```

---

Prohibited

```text
../../

../../../

../../../../
```

---

# 7. COMPONENT RULES

A component is responsible for only one task.

Maximum:

300 lines.

Ideal:

100–150 lines.

---

# 8. CONTAINER / PRESENTATION

Container

↓

receives data

↓

Presentation

↓

renders the interface

---

# 9. DESIGN PHILOSOPHY

Minimum text.

Maximum information.

Maximum whitespace.

Maximum speed.

---

# 10. STYLE

Dark theme

by default.

---

Options:

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

Large numbers.

Minimal decorative text.

---

Portfolio value:

56 px

---

Return:

32 px

---

Secondary information:

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

Use only the system:

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

Only useful ones.

---

Prohibited:

long animations;

animations for beauty alone;

heavy transitions.

---

# 16. LOADING

Skeleton.

Not Spinner.

---

# 17. EMPTY STATE

Each screen must have:

Empty

Loading

Success

Error

Offline

---

# 18. RESPONSIVE

Support:

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

Contains:

ticker search

calculation example

create portfolio

latest dividends

registration CTA

---

# 21. CATALOG

Search

Filters

Sorting

Cards

Infinite Scroll

---

# 22. STOCK PAGE

Name

Price

Chart

Dividends

History

Return

Calculator

News (future)

---

# 23. PORTFOLIO

The main product screen.

---

Displays:

Value

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

The most important card.

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

Our unique feature.

Shows:

```text
Today your portfolio equals:

2.8 MacBook Pro

5 iPhone Pro

18 average salaries

31 food baskets

14 months of rent
```

---

The user switches between:

nominal value

↓

real value

↓

purchasing power

---

# 26. PORTFOLIO CHART

Ranges:

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

Finance

Oil & Gas

IT

Retail

Metallurgy

---

# 29. DIVIDEND CALENDAR

Modes:

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

Before export:

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

Shows:

INN stored?

Passport stored?

Address stored?

Last export

Last deletion

Active devices

---

# 35. TRANSACTION INPUT

EditableSpan is prohibited.

---

Use:

TransactionModal

---

Fields:

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

no manual checks.

---

# 37. ACCESSIBILITY

Keyboard

Screen Reader

Contrast

Focus

ARIA

---

mandatory.

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

only for:

Portfolio

Catalog

Settings

Notifications

User

---

local UI state:

useState

---

# 40. API

Axios Instance

single.

---

no fetch scattered throughout the project.

---

# 41. ERROR UI

An error must explain:

what happened;

what to do;

whether retry is possible.

---

# 42. OFFLINE MODE

Latest snapshots

Latest dividends

Latest catalog

are cached.

---

# 43. MOBILE FIRST

Any new function is designed for phone first.

Then scaled to Desktop.

---

# 44. FUTURE IOS

SwiftUI

must mirror:

screen structure

color system

UX

navigation.

---

# 45. FUTURE ANDROID

Jetpack Compose

fully mirrors Web.

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

The user should:

find a security

in 3 seconds;

understand return

in 5 seconds;

see dividends

in 2 seconds;

create a portfolio

in less than a minute.

---

# 48. DIFFERENTIATOR

OpenInvest should not resemble a terminal.

It should feel like:

"a financial assistant"

not

"a stock-market order book."

---

# 49. SUCCESS METRIC

If a new user without instructions can:

create a portfolio;

understand real return;

understand the impact of inflation;

obtain tax XML;

within 5 minutes,

then the UX is considered correctly implemented.

---

# 50. CODEX REQUIREMENT

Codex must implement any new screen only after checking:

1. Does it overload the user?

2. Can half of the elements be removed?

3. Can the information be shown more simply?

4. Can the action be performed faster?

5. Does the screen match the philosophy:

**Maximum Information. Minimum Friction. Maximum Trust.**
