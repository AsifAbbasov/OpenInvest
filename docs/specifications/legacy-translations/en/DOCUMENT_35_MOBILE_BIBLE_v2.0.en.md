# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 35

# MOBILE ARCHITECTURE BIBLE

# iOS (SwiftUI), Android (Jetpack Compose), Offline First, Sync Engine, Battery Optimization & Cross Platform Strategy

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: MOBILE CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_35_MOBILE_BIBLE_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `500591ebd8ff6f345597387430bda14caba4daf9`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the architecture of OpenInvest mobile applications.

The mobile application is not an adaptation of the web version.

It is an independent client of the unified Personal Capital OS platform.

---

# PHILOSOPHY

OpenInvest Mobile should be:

the fastest;

the most energy-efficient;

the most responsive;

the most understandable;

the most secure.

---

# MOBILE FIRST

Every new feature is designed for:

```text id="m1"
Phone

↓

Tablet

↓

Desktop

↓

Web
```

---

It is prohibited to design Desktop

and then shrink it down to a phone.

---

# NATIVE ONLY

The following are used:

---

iOS

SwiftUI

---

Android

Jetpack Compose

---

Prohibited:

Flutter

React Native

Xamarin

Cordova

for the main version of the product.

---

# WHY NATIVE

Reasons:

---

Minimum memory consumption.

---

Minimum battery consumption.

---

Maximum speed.

---

Better UX.

---

Better system integration.

---

# ARCHITECTURE

Each application has the same structure.

```text id="m2"
Presentation

↓

Application

↓

Domain

↓

Data

↓

Network

↓

Storage
```

---

# SHARED CONTRACT

Web

↓

iOS

↓

Android

use

one OpenAPI contract.

---

# SDK

The SDK is generated automatically.

---

TypeScript

↓

Swift

↓

Kotlin

---

# OFFLINE FIRST

The user should see the portfolio

even without the Internet.

---

# LOCAL CACHE

The following are stored:

---

Assets

---

Snapshots

---

Portfolio

---

Settings

---

Notifications

---

# NEVER CACHE

Password

JWT

Refresh

XML

PDF

TIN

Passport

---

# SYNC STRATEGY

```text id="m3"
Open App

↓

Load Local Snapshot

↓

Show UI

↓

Background Sync

↓

Refresh Changed Data
```

---

The user never waits for a full load.

---

# DELTA SYNC

Only changed data

are transferred.

---

No complete repeated downloads.

---

# BATTERY OPTIMIZATION

Background Tasks

are used minimally.

---

It is prohibited to:

poll the server

every 30 seconds.

---

The following are used:

---

Push Trigger

---

Silent Push

---

Background Refresh

---

Manual Refresh

---

# PUSH STRATEGY

Push contains:

---

Dividend Approved

---

Tax Reminder

---

Portfolio Goal

---

Security Alert

---

Never:

financial amounts.

---

# HOME SCREEN WIDGETS

Supported:

---

Portfolio Value

---

Today's Change

---

Next Dividend

---

Inflation

---

# DYNAMIC ISLAND

(iOS)

Supported:

---

Dividend Today

---

Tax Reminder

---

Portfolio Update

---

# APP SHORTCUTS

Long press:

---

Portfolio

---

Dividend Calendar

---

Tax Export

---

Search Asset

---

# BIOMETRICS

Face ID

Touch ID

Fingerprint

---

are used

only for convenience.

---

Do not replace the password.

---

# LOCAL DATABASE

SQLite

or

Realm

---

is used only as Cache.

---

The source of truth is

Backend.

---

# CONFLICT RESOLUTION

If the user changed data

on two devices:

```text id="m4"
Server Wins

+

Conflict Log

+

User Notification
```

---

# CHARTS

The following are used:

Native Charts

or

the lightest possible libraries.

---

# ANIMATIONS

60 FPS.

---

Maximum:

200 ms.

---

# MEMORY TARGET

iOS

<120 MB

---

Android

<150 MB

---

# STARTUP TARGET

Cold Start

<1.5 sec

---

Warm Start

<500 ms

---

# NETWORK TARGET

Opening Dashboard

<100 KB

---

Opening Portfolio

<150 KB

---

Updating Snapshot

<20 KB

---

# IMAGE STRATEGY

SVG

↓

WebP

↓

PNG

---

# FONT STRATEGY

System fonts are used.

---

Do not load custom fonts.

---

# ACCESSIBILITY

VoiceOver

---

TalkBack

---

Large Text

---

Reduce Motion

---

High Contrast

---

# TABLET MODE

Split Layout

is used.

---

# DARK MODE

By default.

---

Light —

is supported.

---

# APP SECURITY

Root Detection

---

Jailbreak Detection

---

Certificate Pinning

---

Secure Storage

---

# CRASH REPORTING

Without transmitting personal data.

---

# ANALYTICS

The following are collected:

Screen Open

↓

Duration

↓

Crash

↓

Performance

---

Without financial data.

---

# APP STORE STRATEGY

iOS

↓

TestFlight

↓

Staged Rollout

↓

Production

---

Android

↓

Internal

↓

Closed

↓

Open

↓

Production

---

# REVIEW CHECKLIST

Before every release:

---

Battery

---

Memory

---

Offline

---

Push

---

Accessibility

---

Performance

---

Security

---

UX

---

# FUTURE READY

The architecture must allow adding:

---

Apple Watch

---

Wear OS

---

macOS

---

visionOS

---

Android Auto (notifications only)

---

# FINAL MOBILE PRINCIPLE

> **OpenInvest Mobile should not feel like a website inside a phone.**

> **It should work like a native application at the level of Apple Stocks, Wealthfront, or Robinhood: open instantly, consume almost no battery, work without a network, and display the latest current data even on an airplane.**

> **The phone should be a client of Personal Capital OS, not a computing device. All heavy calculations are performed by the server, while the mobile application is responsible exclusively for a fast, understandable, and reliable user experience.**
