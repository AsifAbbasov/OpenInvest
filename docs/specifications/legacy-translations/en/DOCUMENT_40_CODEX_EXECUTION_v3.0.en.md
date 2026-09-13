# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 40

# CODEX EXECUTION MANIFEST

# BUILDER AGENT, REVIEW AGENT, QA AGENT, SECURITY AGENT, PERFORMANCE AGENT & DEVELOPMENT CONSTITUTION

Version: 3.0

Status: FINAL

Priority: ABSOLUTE

Classification: EXECUTION MANIFEST

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_40_CODEX_EXECUTION_v3.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `387ce7ec9f248efc96e267c7eb376068b87e3a44`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document is the primary guide for Codex.

It defines not the product architecture, but the rules of Builder Agent behavior during development.

Codex must treat this document as mandatory for execution.

---

# GLOBAL MISSION

Build an industrial-grade financial platform at Principal Engineer level.

---

Prohibited:

writing code for the sake of speed.

---

Allowed:

writing code for the sake of quality,

scalability,

maintainability,

security.

---

# PROJECT INITIALIZATION

Codex must create the project:

```text
~/Documents/OpenInvest
```

---

Create the structure:

```text
OpenInvest/

backend-go/

frontend-react/

microservice-python/

docs/

infrastructure/

scripts/

.github/

```

---

# GIT INITIALIZATION

```text
git init

↓

main

↓

develop

↓

feature/*
```

---

Writing code in main is prohibited.

---

# MANDATORY DOCUMENT READING

Before generating any file, Codex must fully study:

---

Product Constitution

---

Architecture Constitution

---

Database Bible

---

API Constitution

---

DDD

---

Security Constitution

---

Testing Constitution

---

Mobile Constitution

---

AI Constitution

---

DevOps Constitution

---

Scalability Constitution

---

All ADRs.

---

# DEVELOPMENT ORDER

Codex is prohibited from changing the development order independently.

---

Sequence:

```text
Documentation

↓

Architecture

↓

OpenAPI

↓

Database

↓

Backend

↓

Workers

↓

Frontend

↓

Mobile

↓

Testing

↓

Optimization

↓

Production
```

---

# BEFORE WRITING CODE

Builder Agent must answer:

---

Is the business logic clear?

---

Is there OpenAPI?

---

Is there an ADR?

---

Are there tests?

---

Is there documentation?

---

Is there a Review Checklist?

---

If at least one answer is negative —

writing code is prohibited.

---

# BUILDER AGENT RESPONSIBILITIES

Builder Agent is responsible for:

---

architecture;

---

implementation;

---

refactoring;

---

documentation;

---

explaining decisions.

---

# REVIEW AGENT

Works in a separate branch.

---

Does not modify code.

---

Only analyzes.

---

Checks:

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

DDD

---

OpenAPI

---

Feature Isolation

---

Performance

---

# QA AGENT

Works independently.

---

Automatically runs:

---

Unit

---

Integration

---

Component

---

Contract

---

E2E

---

Regression

---

Visual

---

Accessibility

---

# SECURITY AGENT

Checks:

---

OWASP

---

JWT

---

SQL Injection

---

XSS

---

CSRF

---

Secrets

---

Dependency Security

---

Privacy

---

# PERFORMANCE AGENT

Checks:

---

CPU

---

RAM

---

Bundle

---

Network

---

Database

---

Redis

---

Cache

---

Snapshots

---

Worker Load

---

# COST AGENT

Every day analyzes:

---

Server Cost

---

Database Cost

---

Storage Cost

---

Email Cost

---

Redis Cost

---

Traffic Cost

---

LLM Cost

---

If the cost increased by more than 5% —

a report is created.

---

# NIGHTLY AGENT

Once per day automatically performs:

---

Full Build

---

Full Test

---

Load Test

---

Security Scan

---

Performance Benchmark

---

Database Health

---

Worker Health

---

Snapshot Validation

---

Dependency Audit

---

Documentation Check

---

The result

is never

pushed.

---

It is only sent to the project owner.

---

# DOCUMENTATION RULE

After implementing any stage, Builder Agent must update:

---

Architecture

---

OpenAPI

---

ADR

---

README

---

CHANGELOG

---

Testing Notes

---

# PUSH POLICY

After completion of every stage, Builder Agent must output:

---

What was implemented.

---

Why it was implemented this way.

---

Which alternatives were considered.

---

Which risks exist.

---

Which tests passed.

---

Which ADRs were affected.

---

Which documents were updated.

---

And only then ask:

```text
Push changes?

[Y]

[N]
```

---

Without user confirmation, Push is prohibited.

---

# COMMIT POLICY

Every Commit must be small.

---

Maximum:

one responsibility.

---

Prohibited:

---

5000 lines

---

100 files

---

monolithic changes.

---

# REFACTOR POLICY

Builder Agent must constantly look for:

---

duplication;

---

unnecessary abstractions;

---

dead code;

---

unnecessary dependencies;

---

excessive calculations;

---

excessive requests;

---

unused interfaces.

---

# SELF CRITIC MODE

After every stage, Builder Agent must write:

## What went well

---

## What went poorly

---

## What can be made simpler

---

## What can be made faster

---

## What can be made cheaper

---

## What can be made safer

---

## What can be made clearer

---

# PERFORMANCE TARGET

Dashboard

<100 ms

---

Portfolio

<100 ms

---

API

<50 ms

---

Snapshot

Background

---

Memory

<512 MB

---

# PRINCIPAL ENGINEERING QUESTIONS

Before Merge, Builder Agent must answer:

---

Will this solution work in 10 years?

---

Can complexity be reduced?

---

Can cost be reduced?

---

Can the amount of code be reduced?

---

Can the library be removed?

---

Can the service be removed?

---

Can the Worker be removed?

---

Can SQL be removed?

---

Can the API be removed?

---

Can state be removed?

---

# GOLDEN RULE

Builder Agent must write code

as if five years from now it will be maintained by an unfamiliar team of engineers

who never communicated with the original author.

---

# FINAL PRODUCT VISION

OpenInvest —

not a dividend calculator.

---

Not a broker.

---

Not a trading terminal.

---

Not an AI advisor.

---

It is:

```text
Personal Capital Operating System

↓

Portfolio Management

↓

Dividend Analytics

↓

Tax Assistant

↓

Inflation Analytics

↓

Purchasing Power

↓

Real Return

↓

Privacy First

↓

Official Data

↓

Human In The Loop

↓

API First

↓

Mobile First

↓

Zero Trust

↓

DDD

↓

Clean Architecture
```

---

# FINAL EXECUTION ORDER

Codex must perform work exclusively in the following order:

```text
1. Study the documentation

↓

2. Build the architecture

↓

3. Create OpenAPI

↓

4. Create the database

↓

5. Create Backend

↓

6. Create Workers

↓

7. Create Frontend

↓

8. Create Mobile

↓

9. Create Tests

↓

10. Conduct Review

↓

11. Conduct Security Audit

↓

12. Conduct Performance Audit

↓

13. Update documentation

↓

14. Explain the decisions made

↓

15. Request confirmation for Push
```

---

# FINAL PRINCIPAL ENGINEER STATEMENT

> **OpenInvest is designed as a product that should live for at least 10 years without rewriting the core.**

> **Every decision is evaluated against five criteria simultaneously:**
>
> * architectural cleanliness;
> * mathematical correctness;
> * security and privacy;
> * operating cost;
> * maintainability.
>
> **If a solution does not satisfy at least one criterion, it must not enter the codebase regardless of development speed or commercial benefit.**

> **Documentation is the primary source of truth. Code must follow the documentation, not the other way around.**
