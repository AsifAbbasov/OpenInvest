# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 37

# DEVOPS, GIT WORKFLOW, CI/CD, RELEASE MANAGEMENT, OBSERVABILITY & OPERATIONAL EXCELLENCE

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: DEVOPS CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_37_DEVOPS_WORKFLOW_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `41cdf92b1f1787dae6cc6af7a67e46728909fd42`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the development, delivery, and operation lifecycle of OpenInvest.

The main goal of DevOps is to make the development process predictable, secure, repeatable, and as automated as possible.

---

# DEVOPS PHILOSOPHY

Every action must be:

---

automated;

---

repeatable;

---

idempotent;

---

documented;

---

verifiable.

---

# SOURCE OF TRUTH

The single source of truth:

```text
Git Repository
```

---

All changes go through Git.

---

Manually changing Production is prohibited.

---

# PROJECT INITIALIZATION

Codex must perform the following actions.

```text
~/Documents

↓

OpenInvest/

↓

git init

↓

main

↓

develop

↓

feature/*
```

---

# BRANCH STRATEGY

GitFlow Lite is used.

---

```text
main

↓

develop

↓

feature/*

↓

hotfix/*
```

---

Writing code directly in main is prohibited.

---

# FEATURE DEVELOPMENT

Every task:

```text
Issue

↓

Architecture Review

↓

feature branch

↓

Builder Agent

↓

Review Agent

↓

QA Agent

↓

Security Agent

↓

Performance Agent

↓

Human Approval

↓

Merge
```

---

# COMMIT STRATEGY

Conventional Commits are used.

---

Example:

```text
feat(portfolio): add snapshot calculation

fix(api): correct dividend endpoint

refactor(auth): simplify jwt validation

test(xirr): add edge cases

docs(api): update openapi contract
```

---

# PUSH POLICY

Codex is prohibited from performing Push automatically.

---

After completion of every stage:

Codex must report:

---

what was implemented;

---

what was tested;

---

which files were changed;

---

which tests passed;

---

whether there are risks;

---

and ask:

```text
Push changes?

[Y]

[N]
```

---

Without user confirmation, Push is prohibited.

---

# PULL REQUEST

Every PR must contain:

---

Description

---

Architecture Notes

---

ADR Reference

---

Tests

---

Performance

---

Breaking Changes

---

Documentation Updated

---

# CODE REVIEW

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

Composition over Inheritance

---

Feature Isolation

---

OpenAPI Compatibility

---

# REVIEW RESULT

Review Agent may return only:

---

Approved

---

Approved with Notes

---

Changes Requested

---

Rejected

---

# CI PIPELINE

Every Commit automatically runs:

```text
Lint

↓

Type Check

↓

Build

↓

Unit Tests

↓

Integration Tests

↓

Contract Tests

↓

Component Tests

↓

Security Scan

↓

Dependency Scan

↓

Performance Benchmark

↓

Documentation Check
```

---

If any stage fails —

Merge is prohibited.

---

# CD PIPELINE

```text
develop

↓

Preview Environment

↓

QA

↓

Manual Approval

↓

Production
```

---

# ENVIRONMENTS

Four environments are used.

---

Local

---

Development

---

Staging

---

Production

---

No shared databases.

---

# FEATURE FLAGS

Every new feature is released through Feature Flag.

---

```text
OFF

↓

Internal

↓

Beta

↓

10%

↓

25%

↓

50%

↓

100%
```

---

# ROLLBACK STRATEGY

Every Release must have:

---

Rollback Plan

---

Migration Plan

---

Database Compatibility

---

Cache Strategy

---

# DEPLOYMENT

The following are used:

---

Blue/Green

or

Rolling Update

---

No Production downtime.

---

# OBSERVABILITY

Every service publishes:

---

Latency

---

CPU

---

Memory

---

Errors

---

DB Queries

---

Redis Hits

---

Cache Misses

---

Queue Size

---

# LOGGING

All logs are structured.

---

JSON Format.

---

Required fields:

---

Timestamp

---

Service

---

TraceID

---

RequestID

---

Level

---

Message

---

# METRICS

Minimum set:

---

Requests/sec

---

P95

---

P99

---

Error Rate

---

Memory

---

CPU

---

Queue

---

# ALERTING

Critical:

---

Database Down

---

Redis Down

---

API Down

---

Snapshot Failed

---

Tax Export Failed

---

Email Failed

---

High Error Rate

---

# DASHBOARDS

Dashboards are created for:

---

Backend

---

Frontend

---

Database

---

Workers

---

Notifications

---

Tax

---

AI

---

# DEPENDENCY MANAGEMENT

Every new library undergoes:

---

License Review

---

Security Review

---

Maintenance Review

---

Community Review

---

Bundle Impact Review

---

# VERSIONING

Semantic Versioning.

---

```text
Major.Minor.Patch

2.4.1
```

---

# RELEASE NOTES

Every release must contain:

---

New Features

---

Fixes

---

Performance

---

Security

---

Breaking Changes

---

Migration Guide

---

# NIGHTLY PIPELINE

Every night the following are performed automatically:

---

Build

---

Full Tests

---

Load Tests

---

Security Scan

---

Visual Regression

---

Dependency Updates

---

Bundle Analysis

---

AI UX Review

---

# WEEKLY PIPELINE

Every week:

---

Chaos Test

---

Restore Backup

---

Performance Benchmark

---

Database Health

---

Storage Health

---

# MONTHLY PIPELINE

Every month:

---

Architecture Review

---

Dependency Audit

---

Security Audit

---

Cost Audit

---

API Audit

---

Privacy Audit

---

# COST MONITORING

A separate agent analyzes daily:

---

Server Cost

---

Database Cost

---

Storage Cost

---

Traffic Cost

---

Email Cost

---

API Cost

---

LLM Cost

---

If the cost increased by more than 10% —

an automatic report is created.

---

# AUTOMATED AGENTS

Builder Agent

↓

Review Agent

↓

QA Agent

↓

Security Agent

↓

Performance Agent

↓

Documentation Agent

↓

Cost Agent

↓

Monitoring Agent

---

Each agent has its own responsibility.

---

# ARCHITECTURE FREEZE

After version 1.0:

---

domain structure;

---

OpenAPI;

---

Canonical Data Model;

---

DDD;

---

Plugin API

---

are changed only through ADR.

---

# ENGINEERING RULE

After completion of every stage, Codex must explain:

---

why this architecture was chosen;

---

which alternatives were considered;

---

which trade-offs were accepted;

---

which risks exist;

---

how the decision affects performance, security, and operating costs.

---

# FINAL DEVOPS PRINCIPLE

> **OpenInvest must be developed so that any release can be reproduced, verified, rolled back, and explained.**

> **Not a single byte of code should reach Production without passing Builder Agent, Review Agent, QA Agent, Security Agent, Performance Agent, and explicit confirmation by the project owner.**

> **The main task of DevOps is not to deliver code faster, but to ensure that five years from now the project can be evolved safely without fear of breaking the working system.**
