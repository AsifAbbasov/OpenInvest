# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 23

# MOBILE ARCHITECTURE, OFFLINE-FIRST, BATTERY OPTIMIZATION, NETWORK STRATEGY & CROSS PLATFORM DESIGN

Version: 1.0

Status: APPROVED

Priority: CRITICAL

Classification: MOBILE FOUNDATION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_23_MOBILE_ARCHITECTURE_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `211253b393ae1a93ac7100a96d03827a2a7d5ac4`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# PURPOSE

This document defines the entire architecture of the OpenInvest mobile applications.

The document is mandatory for:

* Codex
* Builder Agent
* Review Agent
* QA Agent
* Mobile Team

---

# PRODUCT PHILOSOPHY

OpenInvest is NOT a heavy broker terminal.

OpenInvest is

a fast intelligent investor assistant.

---

Primary goal:

The user opens the application

↓

understands the state of their capital within 1 second

↓

closes the application.

---

# MOBILE FIRST

Every new function is designed for:

```
Phone

↓

Tablet

↓

Desktop

↓

Web
```

not the other way around.

---

# SUPPORTED PLATFORMS

## Phase 1

Web

---

## Phase 2

iOS

SwiftUI

---

## Phase 3

Android

Jetpack Compose

---

## Future

macOS

visionOS

watchOS

---

# WHY NATIVE?

Using:

SwiftUI

*

Jetpack Compose

provides:

---

less RAM usage;

---

less CPU usage;

---

less Battery Usage;

---

better system integration;

---

less lag;

---

longer support.

---

# ABSOLUTE PRINCIPLE

No business logic

in the mobile application.

---

The mobile application

is UI.

---

All mathematics:

Go Backend

↓

PostgreSQL

↓

Redis

↓

Python Worker

---

# OFFLINE FIRST

The user should be able to see the application

even if the internet connection is lost.

---

# LOCAL CACHE

The following are cached:

Portfolio Summary

Portfolio Chart

Dividend Calendar

Asset Cards

Settings

Theme

Language

---

# NEVER CACHE

JWT

Refresh Token

Passport

INN

Address

Tax Documents

---

# CACHE TTL

Portfolio Summary

5 minutes

---

Asset Catalog

30 minutes

---

Dividend Calendar

12 hours

---

Settings

30 days

---

Theme

∞

---

# CACHE STRATEGY

```
Open Screen

↓

Check Local Cache

↓

Show Cached Data

↓

Background Refresh

↓

Update UI Smoothly
```

---

# USER EXPERIENCE

The user should never

see:

```
Loading...

Loading...

Loading...
```

---

The cache is shown first.

Then the data is quietly refreshed.

---

# NETWORK STRATEGY

Every request must pass the check:

```
Can this request be AVOIDED?
```

If it can,

do not make it.

---

# DELTA UPDATES

Transmit

not the entire object,

but only changes.

---

NOT:

```
Portfolio

2 MB
```

---

YES:

```
Price Changed

4 KB
```

---

# COMPRESSION

All APIs:

Brotli

↓

Gzip

↓

JSON

---

# REQUEST AGGREGATION

Prohibited:

```
10 screens

↓

10 API requests
```

---

Permitted:

```
Dashboard

↓

1 Aggregated API

↓

All data
```

---

# IMAGE STRATEGY

SVG

↓

WebP

↓

AVIF

---

PNG

only when necessary.

---

# CHART STRATEGY

Charts are built

from snapshots.

---

The phone

never

recalculates

5 years of history.

---

It receives:

```
Date

↓

Value

↓

Draw
```

---

# BATTERY OPTIMIZATION

Prohibited:

Polling every second.

---

Use:

Background Refresh

System Scheduling

Push Events

---

# IOS

Use:

BackgroundTasks

---

# ANDROID

Use:

WorkManager

---

# PUSH STRATEGY

Push

only for:

new dividends;

changes in tax status;

ready XML;

an important event.

---

Prohibited:

advertising Push.

---

# SYNCHRONIZATION

All servers operate

only in UTC.

---

The user sees

local time.

---

# MOEX STRATEGY

The source of truth is

the MOEX trading calendar.

---

Not the user's time zone.

---

# EXAMPLE

The user

is in Australia.

---

The exchange has closed in Moscow.

---

A Snapshot has already been created.

---

The user opens the application.

---

Receives:

a ready Snapshot.

---

No recalculations.

---

# LOW INTERNET MODE

A separate mode.

---

The following are disabled:

animation;

auto-refresh;

news;

images.

---

The following remain:

numbers;

charts;

portfolio.

---

# BATTERY SAVER MODE

The following are disabled:

Background Refresh

Heatmaps

Animations

Realtime Updates

---

# ACCESSIBILITY

Dynamic Font

VoiceOver

TalkBack

High Contrast

---

# WIDGETS

Future support:

---

iOS Widget

---

Android Widget

---

They show:

Value

*

Dividends

*

Real Value

---

# QUICK ACTIONS

Through a Widget:

```
Add transaction

↓

Open calendar

↓

Download tax report
```

---

# BIOMETRICS

FaceID

TouchID

Fingerprint

---

optional.

---

# SPLASH SCREEN

Maximum

500 ms.

---

# APP START TARGET

Cold Start

<1.5 sec

---

Warm Start

<500 ms

---

# MEMORY TARGET

iOS

<120 MB

---

Android

<150 MB

---

# CPU TARGET

Idle

<2%

---

# NETWORK TARGET

Dashboard

<50 KB

---

Portfolio

<30 KB

---

Asset Card

<20 KB

---

# SCROLL PERFORMANCE

60 FPS

---

# LISTS

Virtualized

Lazy

Infinite

---

# DARK MODE

Default

---

Light Mode

Optional

---

# TABLET MODE

Do not stretch the interface.

---

Use:

Master

↓

Detail

Layout

---

# FOLDABLE DEVICES

Supported.

---

# AI ASSISTANT

Works

not locally,

but through the Backend.

---

No LLM

in the mobile application.

---

# SECURITY

Private Mode

is fully supported.

---

No tax documents

remain

in application memory.

---

After viewing:

```
PDF

↓

Close

↓

Memory Clear
```

---

# CRASH STRATEGY

The application must never

crash

because there is no network connection.

---

# ANALYTICS

Privacy First.

---

Without Fingerprint.

---

Without hidden tracking.

---

# COMPETITIVE ANALYSIS

Best practices

that must be used:

---

Apple Wallet

(speed)

---

Trading212

(simplicity)

---

Yahoo Finance

(cards)

---

BKS

(portfolio structure)

---

Alfa Investments

(visualization)

---

T-Invest

(navigation)

---

Portfolio Performance

(analytics)

---

# UNIQUE ADVANTAGE

OpenInvest should feel like:

```
Apple Wallet

+

Notion

+

Trading212

+

Personal financial assistant
```

rather than an overloaded broker terminal.

---

# MOBILE DESIGN CHECKLIST

Before creating any screen, Builder Agent must answer:

1.

Can the screen be opened one-handed?

2.

Can the screen be understood within 3 seconds?

3.

Can 30% of the elements be removed?

4.

Does the screen avoid consuming battery unnecessarily?

5.

Does the screen avoid making unnecessary network requests?

6.

Does the screen work fully on a slow internet connection?

7.

Can the screen be used on an airplane without a network connection?

8.

Will this screen still be fast in 5 years, when the user has 20 portfolios and 5000 transactions?

---

# FINAL MOBILE PRINCIPLE

> **OpenInvest should not be the most feature-rich mobile investment application.**

> **It should be the fastest, clearest, calmest, and most energy-efficient application for a long-term investor.**
