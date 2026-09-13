# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 33

# DESIGN SYSTEM, UI/UX CONSTITUTION, VISUAL LANGUAGE, ACCESSIBILITY, MOBILE FIRST & INTERACTION STANDARDS

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: DESIGN SYSTEM

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_33_DESIGN_SYSTEM_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `51aef16bd4a531c25df662b8b40035c1ceed7a0e`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the unified language of OpenInvest interfaces.

Design is considered not as decoration of the product, but as a tool for communicating financial information.

The user should understand the state of their capital within 3–5 seconds without studying instructions.

---

# DESIGN PHILOSOPHY

OpenInvest must not resemble:

* a banking app;
* a trader terminal;
* accounting software.

---

OpenInvest should look like:

> **Personal Capital Dashboard**

Minimalist, calm, fast, understandable.

---

# CORE PRINCIPLES

Principle No. 1

Show only what matters.

---

Principle No. 2

Graphics are more important than text.

---

Principle No. 3

One action — one goal.

---

Principle No. 4

Minimum modal windows.

---

Principle No. 5

Any operation is completed in no more than 3 clicks.

---

# DESIGN LANGUAGE

Foundation:

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

without copying,

only the best UX practices.

---

# COLOR SYSTEM

By default:

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

Supported,

but secondary.

---

# TYPOGRAPHY

One system is used.

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

The following is used:

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

Minimal.

---

Do not use heavy Material shadows.

---

# ANIMATIONS

Duration:

150–250 ms

---

No:

bounce

elastic

overshoot

---

Use:

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

Top card:

```text id="ulh9ik"
Portfolio Value

Real Value

Daily Change
```

---

Second card:

```text id="cq5y0x"
Expected Dividends

Coupons

Tax Forecast
```

---

Third:

```text id="99lm6d"
Inflation

Purchasing Power

Real Return
```

---

# PORTFOLIO SCREEN

Shows:

---

Value

---

Cash

---

Stocks

---

Bonds

---

Return

---

XIRR

---

Allocation pie

---

History

---

# ASSET CARD

Mandatory:

Name

↓

Price

↓

Change

↓

Dividends

↓

Yield

↓

CAGR

↓

History

↓

Calculator

---

# DIVIDEND CALENDAR

Two modes:

---

Calendar

---

List

---

Supported:

color labels;

filters;

search;

grouping.

---

# CHART PRINCIPLES

Charts

must not be overloaded.

---

Maximum:

4 lines.

---

Maximum:

6 colors.

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

Prohibited:

an empty screen.

---

Show:

example;

demo;

explanation;

CTA.

---

# LOADING

Skeletons are used.

---

Spinners are prohibited

for long-loading screens.

---

# FORMS

Minimum fields.

---

Adding a transaction:

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

Automatic masks.

---

Automatic separators.

---

Automatic currency formatting.

---

# MODALS

Used only for:

---

Adding a transaction

---

Deletion

---

Export

---

Settings

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

Every screen is first designed

for the phone.

---

Then

adapted for Desktop.

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

Mandatory:

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

The system setting is supported:

Reduce Motion.

---

# ICON SYSTEM

One library.

---

Do not mix:

Heroicons

Lucide

Material

Feather

at the same time.

---

# UX PRINCIPLE

The user should never have to calculate manually.

---

Instead of:

```text id="eplksn"
+14.72%
```

show:

```text id="w6ptv7"
Your capital grew 8.4% faster than inflation
```

---

# REAL VALUE CARD

An exclusive card.

---

Shows:

```text id="0yhjln"
Today your capital equals:

11 MacBook Pro

or

26 iPhone

or

31 months of an average grocery basket
```

---

# PREMIUM UX

Premium

must not change the interface.

---

It should add:

new analytical cards.

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

Before creating a new screen, Designer Agent answers:

---

Can one block be removed?

---

Can one button be removed?

---

Can one piece of text be removed?

---

Can this be shown graphically?

---

Can two cards be combined?

---

# FINAL DESIGN PRINCIPLE

> **OpenInvest should be perceived not as a complex financial terminal, but as a personal capital management dashboard.**

> **The user should not have to read tables and calculate percentages — the interface must independently transform financial data into understandable visual conclusions, real-life equivalents, and simple decisions.**

> **The best screen is one where the user understands the state of their capital within five seconds, without calculating anything themselves.**
