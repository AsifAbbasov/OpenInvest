# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 19

# FRONTEND ARCHITECTURE, UX, UI SYSTEM, MOBILE FIRST & DESIGN SYSTEM

Version: 1.0

Status: Approved

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_19_FRONTEND_ARCHITECTURE_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `8adf60c4c4f71ec0f1e8c415a7a58216c4718cfc`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# PURPOSE

This document defines the only permitted OpenInvest Frontend architecture.

It is mandatory for:

* Builder Agent
* Review Agent
* QA Agent
* Codex

No UI decisions may be made without complying with this document.

---

# PRODUCT PHILOSOPHY

OpenInvest is not a trader terminal.

OpenInvest is a calm investment assistant.

---

The user should understand the state of their capital within 3 seconds.

---

# DESIGN PRINCIPLES

Maximum information

Minimum noise.

---

Large numbers.

Minimum text.

---

The user should never have to search for profit.

It should be the first element on the screen.

---

# UX PRINCIPLES

No Stress

No Noise

No Popups

No Banner Hell

No Ads in Core Flow

---

# MOBILE FIRST

The mobile version is designed first.

After that:

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

One typeface.

---

Display

Large numbers.

---

Body

Minimal text.

---

Caption

Secondary information.

---

# SPACING

Use

8px grid.

---

# ANIMATIONS

Maximum

150ms

---

No heavy animations.

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

After login, the user should see:

---

Portfolio value

---

XIRR

---

Dividends received

---

Expected dividends

---

Inflation-adjusted value

---

Purchasing power

---

# HERO CARD

The largest card in the application.

---

Displays:

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

Under the Hero Card

---

Add transaction

---

Add money

---

View dividends

---

Download tax report

---

# DASHBOARD BLOCKS

Value

↓

Return

↓

Dividends

↓

Calendar

↓

Portfolio

↓

News

---

# NEWS

Shown

ONLY

for the user's assets.

---

No general feed.

---

# PORTFOLIO SCREEN

Contains:

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

Shows:

```text id="4"
Value

Profit

Loss

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

Use

Recharts.

---

Minimum elements.

---

No 3D.

---

No shadows.

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

is NOT used.

---

Transaction Modal

is used.

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

Shows:

---

Name

---

Price

---

Change

---

Sector

---

Market Capitalization

---

Dividend Yield

---

# STOCK CHART

Minimalistic.

---

No

MACD

RSI

Bollinger

---

OpenInvest is

not a terminal.

---

# DIVIDEND BLOCK

History

↓

Official

↓

Forecast

↓

5-year average

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

Received

Expected

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

The main idea

—

maximum simplicity.

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

Shows:

INN

Address

Passport

Income

Tax

---

The user confirms.

---

# PRIVACY MODE

If

Private Mode

is selected

---

INN

is not stored.

---

Passport

is not stored.

---

Address

is not stored.

---

# PROFILE SCREEN

Contains:

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

A unique feature.

---

Card:

```text id="7"
Your portfolio

2 400 000 ₽

=

2.8 MacBook Pro

or

40 months of utilities

or

68 grocery baskets
```

---

# REAL VALUE CARD

Shows:

---

Nominal value

↓

Real value

↓

Loss due to inflation

---

# AI PANEL

Does not advise buying.

---

Explains:

what changed.

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

Never

show

a stack trace.

---

Only understandable text.

---

# EMPTY STATES

There should be no empty screens.

---

If there is no portfolio:

```text id="8"
Create your first portfolio

Add your first stock

View an example
```

---

# LOADING

Skeleton

---

Not Spinner.

---

# DESIGN REFERENCES

Use best practices from:

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

The user should feel

that the application:

does not pressure them,

does not sell to them,

does not force them,

does not create noise,

but calmly explains

what is happening with their capital.

---

# CODEX REQUIREMENTS

Before creating any new screen, Builder Agent must answer:

1. Can the screen be understood within 5 seconds?

2. Can 30% of the elements be removed without losing functionality?

3. Is the most important number the largest one?

4. Does the screen avoid copying a broker terminal?

5. Can the screen be used one-handed on a phone?

6. Does it comply with Mobile First?

7. Does it avoid increasing battery and mobile data consumption?

8. Does it comply with the OpenInvest philosophy:

**"Maximum useful information with minimum visual noise."**

Only after passing these checks may the screen proceed to implementation.
