# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 27

# FINAL PRODUCT BLUEPRINT, SYSTEM ARCHITECTURE, PRODUCT VISION, MODULE MAP & DEVELOPMENT CONSTITUTION

Version: 1.0

Status: FINAL

Priority: ABSOLUTE

Classification: PROJECT CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_27_FINAL_PRODUCT_BLUEPRINT_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `8a6a5b001a6c242f0831fbe4dafb4cc78dd26d34`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# 1. MISSION

Create the best free service for a long-term investor.

Not a broker.

Not a trading terminal.

Not a news portal.

Not an AI advisor.

---

OpenInvest is

an intelligent financial assistant that:

helps understand capital;

helps account for investments;

helps calculate dividends;

helps calculate real returns;

helps prepare a tax declaration;

does not push asset purchases;

does not sell investment ideas.

---

# 2. PRODUCT PHILOSOPHY

OpenInvest should be:

the fastest;

the clearest;

the most transparent;

the most private;

the most energy-efficient;

the most honest.

---

# 3. PRODUCT PRINCIPLES

Every function must answer five questions.

---

### Is it faster?

---

### Is it simpler?

---

### Does it reduce the number of user actions?

---

### Does it preserve privacy?

---

### Does it scale to 100 000 users?

---

If even one answer is negative,

the function is sent back for rework.

---

# 4. PRODUCT DIFFERENTIATORS

## Most brokers

are overloaded.

---

OpenInvest

is as simple as possible.

---

## Most services

show nominal return.

---

OpenInvest

shows:

nominal;

real;

inflation-adjusted;

dividend;

tax;

XIRR.

---

## Most services

collect huge amounts of personal data.

---

OpenInvest

works even without:

passport;

INN;

address;

phone.

---

## Most services

wait for user actions.

---

OpenInvest

proactively:

detects dividends;

updates the forecast;

creates a tax report;

notifies the user.

---

# 5. TARGET AUDIENCE

Beginner

↓

Long-term investor

↓

Dividend investor

↓

FIRE investor

↓

Retirement investor

↓

Family investor

---

# 6. CORE MODULES

## Landing

---

## Asset Catalog

---

## Asset Card

---

## Dividend Calculator

---

## Portfolio

---

## Portfolio Analytics

---

## Dividend Calendar

---

## Tax Assistant

---

## Inflation Analytics

---

## AI Assistant

---

## Settings

---

## Profile

---

# 7. DASHBOARD

Main screen.

---

Shows:

Portfolio value

↓

Income

↓

XIRR

↓

Dividends

↓

Expected payments

↓

Real Value

↓

Inflation adjustment

↓

Calendar of upcoming payments

↓

News only for the portfolio

---

# 8. REAL VALUE (UNIQUE FEATURE)

In addition to value:

```text
1 200 000 ₽
```

shows:

---

Equivalent:

MacBook Pro

7 units

---

iPhone

16 units

---

Average grocery basket

24 months

---

Average apartment rent

15 months

---

Average salary in Russia

8 months

---

Purchasing power

+8%

or

−4%

---

The user begins to understand

not the numbers,

but the real value of capital.

---

# 9. PORTFOLIO

Shows:

---

Current value

---

Available cash

---

Stocks

---

Bonds

---

ETF (future)

---

Funds (future)

---

Dividends

---

Coupons

---

Commissions

---

Taxes

---

Inflation

---

Real XIRR

---

# 10. MATHEMATICAL ENGINE

Uses:

---

Weighted Average Cost

---

XIRR

---

Dividend Yield

---

Coupon Yield

---

Inflation Adjustment

---

Real Return

---

Tax Impact

---

# 11. TAX MODULE

Modes:

---

Private Mode

---

Convenience Mode

---

Export:

XML

PDF

ZIP

Email

---

Human in the Loop

is mandatory.

---

# 12. AI

AI:

explains;

analyzes;

compares;

visualizes.

---

AI is prohibited from:

advising to buy;

advising to sell;

promising profit.

---

# 13. API STRATEGY

All data:

↓

Go API

↓

Cache

↓

Redis

↓

PostgreSQL

↓

Official APIs

---

The user

never

calls MOEX directly.

---

# 14. DATA SOURCES

Only the following are used:

---

official MOEX;

---

official Bank of Russia;

---

official Rosstat;

---

official issuer disclosures;

---

official FTS XSDs.

---

# 15. MOBILE STRATEGY

Web

↓

iOS SwiftUI

↓

Android Jetpack Compose

↓

macOS (future)

↓

watchOS (future)

---

# 16. PERFORMANCE TARGETS

Dashboard

<150 ms

---

API

<100 ms

---

Cold Start

<1.5 sec

---

Warm Start

<500 ms

---

Bundle

<250 KB gzip

---

Memory Backend

<512 MB

---

# 17. FREE INFRASTRUCTURE STRATEGY

Frontend

Vercel

---

Backend

Railway/Render

---

Database

Neon

---

Redis

Upstash

---

Monitoring

Grafana

---

# 18. MONETIZATION

## Free

Unlimited number of transactions

Portfolio

Dividends

Calendar

XIRR

Real Value

Inflation

---

## Premium

AI analytics

Scenarios

What if I buy

Portfolio Compare

Change history

Automatic reports

Advanced notifications

Export Excel

---

## B2B Future

White Label

Advisor Dashboard

Family Office

Tax Assistant API

---

# 19. SECURITY

Zero Trust

---

Privacy by Design

---

AES-256

---

JWT

---

Refresh Rotation

---

Private Mode

---

Delete by One Click

---

Export My Data

---

# 20. DEVELOPMENT PROCESS

Documentation

↓

Architecture Review

↓

Builder

↓

Review

↓

QA

↓

Security

↓

Performance

↓

Human Approval

↓

Git Push

↓

Deploy

---

# 21. NON-NEGOTIABLE ENGINEERING PRINCIPLES

SOLID

SRP

OCP

LSP

ISP

DIP

DRY

KISS

YAGNI

Law of Demeter

Occam Razor

Composition over Inheritance

Feature Isolation

Clean Architecture

API First

Privacy First

Mobile First

Offline First

---

# 22. WHAT MAKES OPENINVEST UNIQUE

Not the number of features.

---

But the combination:

✓ XIRR

✓ Real Return

✓ Inflation Analytics

✓ Tax Assistant

✓ Privacy Mode

✓ Zero Mandatory Passport Data

✓ Human-in-the-Loop

✓ Proactive Dividend Engine

✓ Lightweight Architecture

✓ Fast Mobile Experience

✓ Official Data Only

✓ Transparent Calculations

✓ Audit Trail

✓ Energy Efficient Design

---

# 23. SUCCESS CRITERIA

In 3 years, OpenInvest should remain:

---

understandable to a beginner;

---

useful to a professional;

---

fast on an old phone;

---

inexpensive to operate;

---

secure;

---

scalable;

---

architecturally clean.

---

# 24. FINAL ENGINEERING COMMAND FOR CODEX

Codex must consider this document the project's primary constitution.

In any conflict:

```
Code
↓

Documentation

↓

Architecture

↓

Principles

↓

Project Constitution
```

the priority always belongs to:

**Project Constitution (Document 27).**

---

# FINAL PRODUCT STATEMENT

> **OpenInvest is not just another dividend calculator.**

> **It is the personal financial operating system of a long-term investor, built on the principles of transparency, privacy, mathematical correctness, energy efficiency, and world-class engineering discipline.**
