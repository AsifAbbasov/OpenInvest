# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 20

# TESTING STRATEGY, QA SYSTEM, REVIEW AGENTS, CI/CD, RELEASE MANAGEMENT & QUALITY GATES

Version: 1.0

Status: Approved

Priority: Critical

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_20_TESTING_STRATEGY_v1.0.md`  
> Source commit: `1310da658d770c988640f63b756ff28d40c5096d`  
> Source blob: `187645517aff47efe4c3f41412da44ef9d45f9f0`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, or implementation status.

---

# PURPOSE

This document defines the mandatory quality-control system for the OpenInvest project.

Goal:

**prevent bad code from reaching production.**

Every commit must pass automated and manual checks.

---

# DEVELOPMENT PHILOSOPHY

Builder Agent

is never

the final authority.

After every change, the code must pass:

```
Builder Agent

↓

Review Agent

↓

QA Agent

↓

Integration Tests

↓

Performance Tests

↓

Security Tests

↓

Human Approval

↓

Git Push
```

---

# PROJECT BRANCH STRATEGY

Use Git Flow only.

```
main

↓

develop

↓

feature/*

↓

review/*

↓

hotfix/*
```

---

# BUILDER AGENT

Builder Agent is responsible only for creating code.

Builder Agent is prohibited from:

independently considering the code perfect;

pushing changes independently;

ignoring Review Agent findings.

---

# REVIEW AGENT

Review Agent is an independent AI.

Its task is:

to find the maximum number of errors.

Review must be as strict as possible.

---

# REVIEW CHECKLIST

Review Agent must check:

---

SOLID

---

SRP

---

OCP

---

LSP

---

ISP

---

DIP

---

DRY

---

KISS

---

YAGNI

---

Law of Demeter

---

Occam Razor

---

Separation of Concerns

---

Composition over Inheritance

---

Clean Architecture

---

Feature Isolation

---

Type Safety

---

Memory Usage

---

API Consistency

---

Security

---

Accessibility

---

Performance

---

# REVIEW REPORT

After the review, the Agent must produce:

```
Architecture Score

Code Quality Score

Performance Score

Security Score

Maintainability Score

Overall Score
```

---

# AUTOMATIC REJECTION

A PR is automatically rejected if:

Architecture Score

<90

or

Security Score

<95

or

Tests Coverage

<90%

---

# QA AGENT

QA Agent is

an independent tester.

It knows nothing about the implementation.

It tests the product through the eyes of the user.

---

# QA RESPONSIBILITIES

Smoke Tests

Regression Tests

Integration Tests

UI Tests

UX Tests

Accessibility Tests

Performance Tests

API Tests

Load Tests

Security Tests

---

# SMOKE TESTS

After every merge:

```
Application Starts

↓

Authorization Works

↓

Portfolio Opens

↓

Catalog Opens

↓

Tax Module Opens

↓

Charts Render
```

---

# REGRESSION TESTS

The following are tested:

existing functions;

new functions;

integration between them.

---

# UNIT TESTS

Target Coverage

95%

Minimum

90%

---

# INTEGRATION TESTS

The following must be tested:

React

↓

API

↓

Go

↓

PostgreSQL

↓

Redis

↓

Python Worker

---

# E2E TESTS

Playwright

---

Scenarios:

Registration

Authorization

Creating a portfolio

Adding a transaction

Editing

Deleting

Export XML

Export PDF

Email

Logout

---

# VISUAL TESTS

Automatically tested:

Desktop

Tablet

Mobile

---

# SUPPORTED RESOLUTIONS

```
390

768

1024

1280

1440

1920
```

---

# SUPPORTED BROWSERS

Chrome

Safari

Firefox

Edge

---

# MOBILE TESTS

iPhone

Android

---

Landscape

Portrait

---

# PERFORMANCE TESTS

Lighthouse

Target

95+

---

Metrics

FCP

LCP

CLS

TTI

INP

---

# MEMORY TESTS

Frontend

must not:

create memory leaks;

create infinite render loops.

---

# API TESTS

The following are tested:

Latency

Errors

Timeouts

Retries

Rate Limits

Compression

Caching

---

# LOAD TESTS

k6

or

Locust

---

Minimum scenario:

10000 concurrent users.

---

# SECURITY TESTS

JWT

XSS

CSRF

SQL Injection

Rate Limit

Broken Auth

Session Hijacking

---

# PRIVACY TESTS

Check whether

passport data

is being stored

when the user selected

Private Mode.

---

# TAX TESTS

The following are tested:

XML

PDF

ZIP

Email

---

# FINANCIAL TESTS

The most critical.

---

The following are tested:

Weighted Average Cost

XIRR

NKD

Coupon

Inflation

Real Value

Dividend Yield

---

# GOLDEN DATASETS

Create dedicated test portfolios.

---

Portfolio A

Stocks only

---

Portfolio B

Stocks + bonds

---

Portfolio C

Additional purchases

---

Portfolio D

Sales

---

Portfolio E

Foreign-currency dividends

---

Portfolio F

Tax declaration

---

# AI TESTS

AI Assistant must:

not provide investment recommendations;

not promise profit;

not distort calculations.

---

# NIGHTLY AGENT

Every 24 hours,

Nightly QA Agent runs.

---

It:

clones develop;

runs all tests;

checks performance;

generates a report.

---

# NIGHTLY REPORT

```
Architecture

Tests

Coverage

Performance

Security

Accessibility

Regression

Memory

API

Overall Status
```

---

# WEEKLY REPORT

Generated automatically.

---

Contains:

Number of errors

Number of fixes

Average response time

RAM usage

CPU usage

Cache Hit Ratio

Slow Queries

---

# CI/CD

Pipeline

```
Commit

↓

Lint

↓

Build

↓

Unit Tests

↓

Integration Tests

↓

E2E Tests

↓

Review Agent

↓

QA Agent

↓

Security Scan

↓

Human Approval

↓

Merge

↓

Deploy
```

---

# DEPLOYMENT RULES

Builder Agent

never

deploys independently.

---

Production Deploy

is possible only after:

Review Agent

*

QA Agent

*

Human Approval.

---

# CODEX WORKFLOW

After each completed stage, Codex must:

1.

Explain

why this architecture was chosen.

---

2.

Explain

which alternatives were considered.

---

3.

Explain

why they are worse.

---

4.

Show the impact on:

RAM

CPU

Network

Battery

Scalability

---

5.

Request confirmation:

```
The stage has been implemented.

Review Agent found no issues.

QA Agent successfully passed the tests.

Push changes to Git?

[Yes]

[No]
```

---

# FINAL QUALITY RULE

Every new function must be checked against one question:

> **"If 100 000 investors use the project tomorrow, will this architecture remain just as fast, secure, inexpensive to operate, and understandable to maintain?"**

If the answer is negative, the function is sent back for rework.
