# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 17

# BUSINESS LOGIC, PORTFOLIO ENGINE, DIVIDEND ENGINE, TAX ENGINE & PRODUCT RULES

Version: 1.0

Status: Approved

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_17_BUSINESS_LOGIC_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `cc0299e7559a8c0658057e723176440355bf4ad4`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# PURPOSE

This document describes not technologies, but OpenInvest business logic.

This document explains to Codex:

* how the product should work;
* which calculations are performed;
* which restrictions exist;
* what is the source of truth;
* what is shown to the user.

Backend and Frontend must follow these rules.

---

# PRODUCT PHILOSOPHY

OpenInvest IS NOT a broker.

OpenInvest DOES NOT make investment decisions.

OpenInvest DOES NOT sell financial products.

OpenInvest is:

* an investment assistant;
* an analytics platform;
* a dividend calculator;
* a tax assistant;
* a portfolio accounting system.

---

# SINGLE SOURCE OF TRUTH

All calculations happen only on Backend.

Frontend calculates nothing.

Mobile calculates nothing.

Frontend only displays the result.

---

# PORTFOLIO ENGINE

Each user may have several portfolios.

For example:

```
Retirement

Dividends

USD Portfolio

Child

Bonds

ETF
```

---

# PORTFOLIO TYPES

Investment

Dividend

Bond

Mixed

Cash

Custom

---

# PORTFOLIO SUMMARY

The main screen always displays:

```
Current value

Available cash

Profit

Loss

Dividends received

Coupons received

Expected payments

Average return

XIRR

Inflation-adjusted return

Purchasing power

Number of assets

Number of sectors
```

---

# NOMINAL VALUE

Always displayed.

```
Asset value

+

cash

=

Nominal value
```

---

# REAL VALUE

Second card.

```
Value

adjusted for official inflation

from Rosstat / CBR.
```

---

# PURCHASING POWER

A killer product function.

The user can click:

```
What can I buy today?
```

---

# EXAMPLES

```
MacBook Pro

iPhone

PlayStation

Average food basket

Average salary

Utilities

Gasoline

Gold

Apartment rent

Airline tickets
```

---

# USER CHOICE

Modes can be switched:

```
Nominal value

↓

Real value

↓

Purchasing power

↓

Inflation dynamics
```

---

# INFLATION ENGINE

Only official inflation is used.

Sources:

Rosstat

CBR

---

# DIVIDEND ENGINE

The system stores:

```
History

Official payments

Forecast payments

Average yield

Dividend CAGR

Dividend Growth
```

---

# DIVIDEND STATUS

```
Official

Board Recommendation

Forecast

Canceled

Updated
```

---

# DIVIDEND CALCULATOR

The user enters:

```
Quantity

Purchase price

Purchase date
```

---

# BACKEND CALCULATES

```
Dividends received

Expected dividends

Dividend Yield

Yield on Cost

XIRR

Real Return

Inflation Adjusted Return
```

---

# BOND ENGINE

Separate business logic.

---

# BOND CALCULATES

```
NKD

Coupons

Yield to maturity

Yield to offer

Current yield

Average yield
```

---

# NKD

NKD is automatically accounted for as an expense.

---

# COUPON

Each coupon is

a cash flow.

---

# CASH ENGINE

Available money is also an asset.

It is displayed separately.

---

# CASH DOES NOT PARTICIPATE

in:

Dividend Yield

but participates

in portfolio value.

---

# XIRR ENGINE

The project's only official return metric.

---

It is prohibited to use:

```
Simple percentage

Average return

Pretty percentages

ROI without accounting for time
```

---

# XIRR INPUTS

BUY

SELL

DIVIDEND

COUPON

DEPOSIT

WITHDRAW

---

# OUTPUT

```
Annual Return

Real Annual Return

Inflation Adjusted Return
```

---

# COST BASIS

Use

Weighted Average Cost.

---

# SELL

Sale

does not change Average Cost.

---

# REAL PROFIT

```
Sale

+

dividends

+

coupons

-

commissions

-

taxes

-

inflation
```

---

# TAX ENGINE

OpenInvest helps,

but does not replace a tax consultant.

---

# TAX MODES

```
Simple

Advanced

Manual
```

---

# SIMPLE

Shows:

an approximate amount.

---

# ADVANCED

Takes into account:

historical exchange rates,

taxes,

dividends,

coupons,

commissions.

---

# MANUAL

The user corrects values manually.

---

# HUMAN IN THE LOOP

Before export, a verification window always appears.

```
AI prepared the document.

Please,

check the data.
```

---

# XML EXPORT

Created:

only on request.

---

# PDF EXPORT

Created:

only on request.

---

# EMAIL EXPORT

```
XML

+

PDF

+

ZIP
```

---

# DATA RETENTION

If the user chooses:

```
Do not store data
```

then:

INN

passport

address

exist only in process memory.

---

# PRIVACY MODE

By default:

```
Minimal data collection.
```

---

# NOTIFICATIONS

The system works proactively.

---

# EVENTS

New dividends

New coupons

Change in return

Change in inflation

New documents

Tax deadline

---

# SMART NOTIFICATIONS

Do not spam.

---

For example:

```
Maximum 1 push per day

Maximum 1 email per hour
```

---

# DASHBOARD

The main screen always contains:

```
Value

Profit

XIRR

Dividends

Expected payments

Inflation

Purchasing power

News only for the portfolio
```

---

# NEWS

No news dump.

---

Only events

related to the user's assets are shown.

---

# WATCHLIST

The user can create a watchlist.

---

# COMPARISON

Can compare:

```
Sber

versus

Gazprom

versus

Lukoil
```

---

# WHAT IF ENGINE

A killer function.

```
What happens

if I buy another 50 shares?

What happens

if I invest another 100000 rubles?

How will dividend flow change?

How will XIRR change?

How will purchasing power change?
```

---

# SCENARIOS

```
Conservative

Balanced

Aggressive

Custom
```

---

# AI ASSISTANT

AI never gives investment recommendations.

---

It explains:

```
what happened;

how the portfolio changed;

how dividends changed;

how taxes changed;

how purchasing power changed.
```

---

# PRODUCT DIFFERENTIATORS

Why the user chooses OpenInvest:

---

### 1.

Honest XIRR return

rather than pretty percentages.

---

### 2.

Inflation-adjusted return.

---

### 3.

Purchasing power.

---

### 4.

Automatic dividend assistant.

---

### 5.

Automatic XML and PDF preparation.

---

### 6.

Minimal personal-data collection.

---

### 7.

Privacy by Design.

---

### 8.

Full audit of all calculations.

---

### 9.

Very fast interface.

---

### 10.

Backend takes on all heavy mathematics.

---

# CODEX REQUIREMENTS

Before implementing any new functionality, Builder Agent must answer:

1. Does the function improve the investor's life?

2. Does it provide new value?

3. Does it duplicate an existing function?

4. Does it complicate the interface?

5. Can this function be explained to the user in 10 seconds?

6. Does it work the same way in Web, iOS, and Android?

7. Does it comply with the philosophy:

**"Minimum user actions — maximum automation and transparency."**

If at least one answer is negative, the function must not enter the product without a separate architectural discussion.
