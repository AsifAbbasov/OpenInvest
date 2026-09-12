# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 14

# MATHEMATICAL ENGINE, PORTFOLIO CALCULATION & FINANCIAL MODEL

Version: 1.0

Status: Approved

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_14B_MATHEMATICAL_ENGINE_v1.0.md`  
> Source commit: `2079450858bbdd00b583773bc2a161420ef4063c`  
> Source blob: `8b3f8bf878c545da3bd4c19eebc01937d923c2d3`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture.

---

# PURPOSE

This document defines the only permitted mathematical engine for OpenInvest.

Any other calculation method is considered incorrect.

All calculations must be performed exclusively on Backend.

Frontend never calculates investment metrics.

---

# PRINCIPLES

The system must show:

not attractive numbers,

but real money.

The user must understand:

* how much they earned;
* how much they lost;
* how much inflation consumed;
* how much they received in dividends;
* how much tax they have already paid;
* how much tax they will still have to pay;
* what their true annual return is.

---

# GLOBAL PORTFOLIO MODEL

Portfolio value consists of five independent components.

```
Portfolio Value

=

Stocks

+

Bonds

+

Cash

+

Received Dividends

+

Accrued Coupon Income
```

---

# PORTFOLIO COMPONENTS

Stocks

market value of stocks

---

Bonds

market value of bonds

---

Cash

cash balance

---

Dividends

dividends paid

---

Coupons

coupons received

---

# TOTAL RETURN

Total return is calculated as

```
Total Return

=

Price Return

+

Dividend Return

+

Coupon Return

-

Commissions

-

Taxes
```

---

# PRICE RETURN

```
(Current Market Value)

-

(Purchase Cost)
```

---

# DIVIDEND RETURN

```
Sum(All Received Dividends)
```

---

# COUPON RETURN

```
Sum(All Coupon Payments)
```

---

# COMMISSIONS

Included:

purchase

sale

service

depository

exchange fees

if the user specifies them.

---

# TAXES

Return is displayed in two variants.

Nominal Return

without taxes.

---

Net Return

after taxes.

---

The user can always switch between them.

---

# AVERAGE COST

For stocks, use

Weighted Average Cost.

```
New Average

=

((OldQty × OldAverage)

+

(NewQty × NewPrice))

/

(TotalQty)
```

---

SELL

does not change Average Cost.

---

# FIFO

Additionally, it must be possible to switch the accounting method.

Settings

↓

Accounting Method

↓

Average Cost

FIFO

---

This will allow different tax models to be used.

---

# XIRR

The main product metric.

Used instead of ordinary return.

---

Reason:

the investor continuously purchases additional securities.

---

Use the Newton–Raphson method.

---

CashFlow:

BUY

negative flow

---

SELL

positive flow

---

DIVIDEND

positive flow

---

COUPON

positive flow

---

COMMISSION

negative flow

---

TAX

negative flow

---

# DISPLAYED RETURNS

Displayed simultaneously:

Today's Return

Week

Month

Quarter

Year

3 Years

5 Years

All Time

XIRR

---

# REAL RETURN

In addition to nominal return,

Real Return is calculated.

```
Real Return

=

Nominal Return

-

Inflation
```

---

# INFLATION MODEL

Official Russian inflation is used.

---

History is stored separately.

---

If the user changes country,

the corresponding source is used.

---

# PURCHASING POWER

A unique product function.

---

The user sees

not simply

200 000 ₽

but

real purchasing power.

---

For example:

```
Your portfolio:

200 000 ₽

≈

1 MacBook Pro

or

2 iPhone Pro

or

6 months of the average food basket

or

8 months of utility payments

or

1 average salary × 3 months
```

---

All equivalents can be disabled.

---

# BONDS

Bonds have a separate model.

---

Purchase Cost

*

NKD

=

Investment Cost

---

Coupon

=

Positive Cash Flow

---

NKD

=

Negative Cash Flow

---

YTM

must be calculated separately.

---

# DURATION

For bonds calculate

Duration

Modified Duration

---

# EXPECTED DIVIDENDS

Only officially announced payments are used.

---

Forecast payments

are displayed separately.

---

The user always understands

what is guaranteed,

and what is a forecast.

---

# EXPECTED CASHFLOW

A separate screen.

Shows

expected receipts

by month.

---

# SECTOR ALLOCATION

Displays:

Stocks

Bonds

Cash

ETF

---

And separately

by economic sectors.

---

# RISK SCORE

Calculated:

by concentration,

by sectors,

by assets,

by currencies.

---

# REBALANCING

AI shows:

```
If you buy another 10 shares,

dividend yield will become:

8.42%

and risk will decrease by:

4.1%
```

---

# INFLATION ALERT

If

Real Return

becomes negative,

the system displays a warning.

```
Your portfolio is growing,

but its purchasing power is declining.
```

---

# DIVIDEND EFFICIENCY

Displays

```
Dividend Return

vs

Price Return
```

---

# PORTFOLIO HEALTH SCORE

AI calculates an integral score.

Takes into account:

diversification

volatility

dividends

taxes

inflation

commissions

concentration

liquidity

---

# WHAT IS NEVER ALLOWED

Prohibited:

using simple return;

adding percentages;

ignoring commissions;

ignoring taxes;

ignoring inflation;

ignoring NKD;

building a chart only from the current share quantity.

---

# HUMAN FRIENDLY MODE

All complex metrics are accompanied by an explanation.

For example:

```
XIRR

11.82%

What does this mean?

If your money worked with equal efficiency every year,

you would receive 11.82% annually.
```

---

# FUTURE EXTENSIONS

The engine must support without rewriting the architecture:

* ETF;
* REIT;
* funds;
* crypto assets;
* foreign stocks;
* multicurrency portfolios;
* family portfolios;
* corporate portfolios;
* joint investment accounts.

---

# CODEX REQUIREMENT

Any new function must pass mandatory checks:

1. Does it violate XIRR?
2. Does it violate Average Cost?
3. Does it violate tax calculation?
4. Does it violate inflation calculation?
5. Does it violate purchasing-power calculation?
6. Does it reduce Backend performance?
7. Does it increase network traffic unnecessarily?
