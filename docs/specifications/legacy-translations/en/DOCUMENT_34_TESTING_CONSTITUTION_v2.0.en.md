# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 34

# TESTING CONSTITUTION, QUALITY ASSURANCE, AUTOMATED VALIDATION, CHAOS ENGINEERING & RELEASE GATE

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: QUALITY FOUNDATION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_34_TESTING_CONSTITUTION_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `874251eec8b72c11b527be3b4e7cca5bce4f57ff`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the unified OpenInvest quality assurance strategy.

Product quality is determined not by the number of tests, but by the probability of detecting an error before it reaches Production.

Any code that has not passed the full verification cycle is considered unsuitable for release.

---

# QUALITY PHILOSOPHY

OpenInvest uses the principle:

```text
Prevent Bugs

↓

Detect Bugs

↓

Localize Bugs

↓

Fix Bugs

↓

Prevent Regression
```

---

Fixing an error is cheaper than allowing it to appear.

---

# TEST PYRAMID

The following structure is used:

```text
                E2E
             UI Testing
         Integration Tests
         Contract Tests
          Component Tests
             Unit Tests
```

---

# TARGET COVERAGE

Unit Tests

95%

---

Integration Tests

90%

---

Contract Tests

100%

---

Critical Business Logic

100%

---

Tax Module

100%

---

Mathematical Engine

100%

---

# UNIT TESTS

The following are tested:

---

XIRR

---

TWR

---

Sharpe

---

Sortino

---

CAGR

---

Dividend Yield

---

Weighted Average Cost

---

Real Return

---

Inflation Engine

---

Purchasing Power

---

# MATHEMATICAL VALIDATION

Every formula is tested against:

---

Known Dataset

---

Edge Case

---

Zero Case

---

Negative Case

---

Large Dataset

---

Random Dataset

---

Floating Point Precision

---

# SNAPSHOT TESTING

The following are tested:

---

Dashboard

---

Portfolio

---

Charts

---

Calendar

---

Tax Export

---

Settings

---

# COMPONENT TESTS

Each React component is tested separately.

---

Props

---

Events

---

State

---

Rendering

---

Accessibility

---

# CONTRACT TESTS

Frontend

↓

OpenAPI

↓

Backend

---

If Backend violates the contract,

the Build is prohibited.

---

# INTEGRATION TESTS

The following is tested:

```text
Go

↓

Redis

↓

PostgreSQL

↓

Python Worker

↓

SMTP
```

---

# E2E TESTS

Full user scenario:

---

Registration

↓

Creating a portfolio

↓

Adding a transaction

↓

Viewing dividends

↓

Exporting XML

↓

Deleting an account

---

# VISUAL REGRESSION

Every screen is compared

with the previous version.

---

The following are checked:

---

Spacing

---

Typography

---

Colors

---

Cards

---

Charts

---

Buttons

---

# ACCESSIBILITY TESTS

The following are checked:

---

Keyboard Navigation

---

Screen Reader

---

Contrast

---

Focus

---

ARIA

---

WCAG AA

---

# SECURITY TESTS

The following are performed automatically:

---

JWT Tests

---

Rate Limit

---

SQL Injection

---

XSS

---

CSRF

---

Broken Auth

---

Privilege Escalation

---

# LOAD TESTS

100 Users

↓

1000 Users

↓

5000 Users

↓

10000 Users

↓

50000 Users

---

# PERFORMANCE TESTS

Dashboard

<150 ms

---

Portfolio

<100 ms

---

Snapshot

<50 ms

---

Tax Export

<3 sec

---

# CHAOS ENGINEERING

The following are artificially disabled:

---

Redis

---

SMTP

---

Python Worker

---

PostgreSQL Replica

---

MOEX API

---

The following is tested:

Whether the system continues operating.

---

# RESILIENCE TESTS

The following are tested:

---

Timeout

---

Retry

---

Backoff

---

Circuit Breaker

---

Fallback Cache

---

# API TESTS

Every endpoint is tested for:

---

Authorization

---

Validation

---

Pagination

---

Sorting

---

Filtering

---

Errors

---

Response Time

---

# DATABASE TESTS

The following are tested:

---

Indexes

---

Foreign Keys

---

Partitioning

---

Snapshot Insert

---

Migration

---

Rollback

---

# MIGRATION TESTS

Every migration:

---

Up

↓

Validate

↓

Rollback

↓

Validate

---

# MOBILE TESTS

Android

---

iOS

---

Tablet

---

Foldable

---

Landscape

---

Portrait

---

# CROSS PLATFORM

Chrome

---

Safari

---

Firefox

---

Edge

---

WebView

---

# USER EXPERIENCE TESTS

AI Agent automatically:

---

creates a portfolio;

---

adds transactions;

---

searches for stocks;

---

builds charts;

---

exports a declaration;

---

deletes the profile.

---

Checks:

Number of clicks

↓

Errors

↓

Unclear areas

↓

Completion time

---

# REGRESSION TESTING

After every Merge:

---

Unit

↓

Integration

↓

Contract

↓

Component

↓

UI

↓

E2E

↓

Performance

↓

Security

↓

Accessibility

---

# NIGHTLY TESTS

Every night:

---

Build

---

Full Tests

---

Load Tests

---

Security Scan

---

Dependency Scan

---

Visual Regression

---

AI UX Review

---

# WEEKLY TESTS

Every week:

---

Chaos Engineering

---

Disaster Recovery

---

Backup Restore

---

Performance Benchmark

---

Bundle Analysis

---

# MONTHLY TESTS

Every month:

---

10000+

virtual users

---

Real market simulation

---

Full analytics recalculation

---

Tax export

---

# RELEASE GATE

Release is prohibited if:

---

Unit <95%

---

Critical Coverage <100%

---

Contract FAIL

---

Security FAIL

---

Performance FAIL

---

Accessibility FAIL

---

Documentation Outdated

---

# REVIEW AGENT CHECKLIST

Before Merge, the agent must answer:

---

Can the test be removed?

---

Can tests be combined?

---

Is there duplication?

---

Are there false-positive checks?

---

Are there nondeterministic tests?

---

# BUG CLASSIFICATION

P0

Loss of money

---

P1

Calculation error

---

P2

UX violation

---

P3

Visual error

---

P4

Cosmetic error

---

P0

is fixed immediately.

---

# FINAL QUALITY PRINCIPLE

> **OpenInvest is considered tested not when all tests have passed successfully.**

> **OpenInvest is considered tested when Review Agent, QA Agent, Security Agent, Performance Agent, and Chaos Agent together have failed to drive the system into an incorrect financial result, loss of user data, or violation of the user experience.**

> **Any code without tests is an assumption. Any tested code is an engineering decision.**
