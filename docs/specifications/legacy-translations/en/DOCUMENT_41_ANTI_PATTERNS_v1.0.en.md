# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 41

# ANTI-PATTERNS, FORBIDDEN DECISIONS, ARCHITECTURAL SMELLS & PROJECT RED BOOK

Version: 1.0

Status: FINAL

Priority: ABSOLUTE

Classification: ENGINEERING CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_41_ANTI_PATTERNS_v1.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `49f8add66784f4d152c08adec40f0234a4c813e5`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document describes not

**what must be done**,

but

**what is categorically prohibited.**

---

The document is mandatory for:

Builder Agent

Review Agent

QA Agent

Security Agent

Performance Agent

Documentation Agent

Human Developer

---

# FUNDAMENTAL PRINCIPLE

Every solution must be, to the greatest extent possible:

---

simple;

---

understandable;

---

predictable;

---

scalable;

---

reversible.

---

# 1. ARCHITECTURAL ANTI-PATTERNS

## Prohibited

God Object

---

God Service

---

God Component

---

God Hook

---

God Store

---

Monolithic Context

---

Massive Controller

---

Massive Service

---

Massive Reducer

---

Massive SQL Query

---

# LIMITS

React Component

≤250 lines

---

Hook

≤150 lines

---

Go Service

≤300 lines

---

SQL Migration

one responsibility

---

# 2. DATABASE ANTI-PATTERNS

Prohibited

---

SELECT *

---

N+1 Query

---

Nested Loop without necessity

---

Data duplication

---

Missing indexes

---

Dynamic SQL

---

DROP COLUMN without migration

---

Storing JSON instead of a proper model

---

Application logic inside SQL

---

# REVIEW QUESTIONS

Can JOIN be reduced?

---

Can Snapshot be used?

---

Can Materialized View be used?

---

Can Cache be used?

---

# 3. BACKEND ANTI-PATTERNS

Prohibited

---

business logic in Controller

---

SQL inside Handler

---

Redis inside Domain

---

HTTP inside Domain

---

Global Variables

---

Singleton without necessity

---

Reflection for aesthetics

---

# 4. FRONTEND ANTI-PATTERNS

Prohibited

---

Props Drilling

---

State Explosion

---

Global Store for local state

---

Nested Modals

---

Nested Scroll

---

20 useEffect in a row

---

UI with business logic

---

API calls inside components

---

# 5. REDUX ANTI-PATTERNS

Prohibited

---

Store Everything

---

Derived State

---

Mutable State

---

Business Logic inside Slice

---

1000 lines in Slice

---

# 6. REACT ANTI-PATTERNS

Prohibited

---

Anonymous Functions Everywhere

---

Inline Objects Everywhere

---

Inline Styles Everywhere

---

Huge JSX

---

Conditional Hell

---

Magic Numbers

---

# 7. API ANTI-PATTERNS

Prohibited

---

REST + GraphQL simultaneously

---

Breaking Changes

---

Versionless API

---

Huge Payload

---

GET that changes data

---

POST without Idempotency

---

# 8. CACHE ANTI-PATTERNS

Prohibited

---

Infinite TTL

---

Cache Everything

---

Cache Personal Data

---

Cache XML

---

Cache JWT

---

# 9. AI ANTI-PATTERNS

AI is prohibited from

---

advising users to buy stocks

---

advising users to sell stocks

---

promising returns

---

modifying transactions independently

---

submitting declarations

---

generating financial data

---

# 10. PRIVACY ANTI-PATTERNS

Prohibited

---

mandatory TIN

---

mandatory passport

---

mandatory phone

---

mandatory address

---

mandatory storage of declarations

---

# 11. SECURITY ANTI-PATTERNS

Prohibited

---

JWT in LocalStorage

---

passwords in logs

---

secrets in Git

---

API Keys in code

---

disabling TLS

---

# 12. TESTING ANTI-PATTERNS

Prohibited

---

100% Mock project

---

Unit without real scenarios

---

E2E only Happy Path

---

ignoring Edge Cases

---

# 13. PERFORMANCE ANTI-PATTERNS

Prohibited

---

recalculating the portfolio on every request

---

recalculating XIRR on the client

---

loading the full history

---

50 API requests on Dashboard

---

# 14. COST ANTI-PATTERNS

Prohibited

---

LLM for simple calculations

---

Redis for permanent storage

---

Python Worker for CRUD

---

constant polling

---

# 15. UX ANTI-PATTERNS

Prohibited

---

10 registration screens

---

5 modal windows in a row

---

tables on mobile

---

hidden buttons

---

non-obvious actions

---

# 16. DESIGN ANTI-PATTERNS

Prohibited

---

10 success colors

---

20 text sizes

---

5 icon libraries

---

3 design systems simultaneously

---

# 17. MOBILE ANTI-PATTERNS

Prohibited

---

WebView instead of Native

---

updating the API every 30 seconds

---

large Bundles

---

constant Background Jobs

---

# 18. DEVOPS ANTI-PATTERNS

Prohibited

---

Push directly to main

---

Deploy without tests

---

Deploy without Review

---

Deploy without Backup

---

# 19. DOCUMENTATION ANTI-PATTERNS

Prohibited

---

code without documentation

---

ADR without a decision

---

OpenAPI does not match Backend

---

README is outdated

---

# 20. PRODUCT ANTI-PATTERNS

It is prohibited to turn OpenInvest into:

---

a broker

---

a social network

---

a news portal

---

a chat

---

a crypto exchange

---

a trader terminal

---

# 21. MVP ANTI-PATTERNS

Prohibited

---

adding features for the sake of quantity

---

making Premium by cutting Free

---

writing AI "for show"

---

copying competitors' interfaces

---

# 22. CODE REVIEW STOP LIST

Review Agent must immediately reject Merge if the following is found:

---

any without a reason

---

TODO in Production

---

magic strings

---

magic numbers

---

logic duplication

---

cyclic dependencies

---

dead code

---

# 23. PRINCIPAL ENGINEERING QUESTIONS

Before Merge, Builder Agent must ask themselves:

---

Can this code be deleted?

---

Can it be written in half as much code?

---

Can the library be avoided?

---

Can the Worker be avoided?

---

Can Redis be avoided?

---

Can SQL be avoided?

---

Can the API be avoided?

---

Can state be avoided?

---

If the answer is "Yes",

the code must be simplified.

---

# 24. OCCAM MODE

Always choose:

---

a simple solution

---

instead of a beautiful one.

---

an understandable one

---

instead of a clever one.

---

a stable one

---

instead of a fashionable one.

---

# 25. FINAL RED BOOK PRINCIPLE

> **The most dangerous project mistake is not bad code.**

> **The most dangerous mistake is the gradual complication of the system through small "temporary" solutions.**

> **OpenInvest must resist architectural entropy.**

> **Every Builder Agent, Review Agent, and engineer must leave the system simpler, clearer, faster, and cheaper than it was before their changes.**

> **If the necessity of a new entity, new service, new library, or new abstraction layer cannot be proven, they must not appear in the project.**
