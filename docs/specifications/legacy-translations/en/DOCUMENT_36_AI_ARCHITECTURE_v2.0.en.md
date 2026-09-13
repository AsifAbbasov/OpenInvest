# OPENINVEST MASTER ENGINEERING SPECIFICATION

# DOCUMENT 36

# AI ARCHITECTURE BIBLE

# AI ASSISTANT, RAG, AGENT SYSTEM, HUMAN-IN-THE-LOOP, PROMPT SECURITY & AUTONOMOUS ANALYTICS

Version: 2.0

Status: FINAL

Priority: ABSOLUTE

Classification: AI CONSTITUTION

> **Translation provenance**
>
> Original source: `docs/specifications/legacy/DOCUMENT_36_AI_ARCHITECTURE_v2.0.md`  
> Source commit: `9e0c7b44d77aaa01008287e982b7882d136f5dc4`  
> Source blob: `47950876a33886da345b1b672afc52e2e7d4e108`  
> Translation status: **Non-authoritative English translation.**  
> The original historical document controls. This translation preserves the historical meaning and does not modernize the architecture, product state, security model, legal interpretation, governance model, or implementation status.

---

# PURPOSE

This document defines the operating principles of OpenInvest artificial intelligence.

AI is the user's assistant.

AI never becomes a capital manager.

OpenInvest categorically prohibits autonomous investment decision-making by artificial intelligence.

---

# AI PHILOSOPHY

AI should:

explain;

analyze;

visualize;

structure;

remind;

teach.

---

AI is prohibited from:

recommending buying;

recommending selling;

promising profit;

guaranteeing returns;

replacing a financial advisor.

---

# AI POSITION

```text
User

↓

AI Assistant

↓

OpenAPI

↓

Go Backend

↓

Canonical Database
```

---

AI never receives direct access to PostgreSQL.

---

# AI TYPES

## Assistant AI

Works with the user.

---

## Analytics AI

Builds analytical conclusions.

---

## Tax AI

Helps understand the declaration.

---

## Portfolio AI

Explains changes in capital.

---

## Education AI

Explains financial terms.

---

# HUMAN IN THE LOOP

Any critical action requires user confirmation.

---

AI can:

prepare XML;

prepare PDF;

prepare an email;

prepare a scenario.

---

AI cannot:

send XML;

delete the profile;

modify transactions;

modify taxes;

create a portfolio without the user.

---

# RAG ARCHITECTURE

AI answers only on the basis of verified sources.

```text
Official Documents

↓

Knowledge Index

↓

Embeddings

↓

Retriever

↓

Context Builder

↓

LLM

↓

Verified Answer
```

---

# KNOWLEDGE SOURCES

Only the following are used:

---

OpenInvest documentation;

---

official Bank of Russia documents;

---

official Federal Tax Service documents;

---

official MOEX documents;

---

official product documentation.

---

# AI MEMORY

AI does not store:

passwords;

TIN;

passport;

JWT;

Refresh Token;

SMTP;

secrets.

---

AI stores only the context of the current dialogue.

---

# PROMPT SECURITY

Every Prompt goes through:

```text
Input

↓

Validation

↓

Sanitization

↓

Policy Check

↓

Execution
```

---

# PROMPT INJECTION DEFENSE

Attempts to do the following are ignored:

---

obtain system instructions;

---

obtain keys;

---

bypass access rights;

---

change business logic.

---

# TOOL PERMISSIONS

AI has access only to:

---

Portfolio Read

---

Analytics Read

---

Tax Preview

---

Dividend Calendar

---

Notifications Preview

---

AI does not have access to:

Delete

Update

Execute Payments

Send Declaration

---

# AI CONTEXT WINDOW

Context priority:

```text
Current User Data

↓

Portfolio

↓

Snapshots

↓

Analytics

↓

Knowledge Base

↓

General Information
```

---

# AI EXPLANATION MODE

Every conclusion is accompanied by:

---

a formula;

---

a source;

---

a date;

---

a confidence level.

---

Example:

```text
XIRR

12.48%

Source:

Transactions

Snapshots

Calculation date:

2027-03-12

Confidence:

100%
```

---

# AI PORTFOLIO REVIEW

AI can automatically generate:

---

a monthly summary;

---

capital dynamics;

---

dividends;

---

taxes;

---

inflation adjustment;

---

real return.

---

# AI GOAL PLANNER

The user can specify:

---

save for an apartment;

---

build retirement capital;

---

receive dividends;

---

save for education.

---

AI shows only mathematical scenarios.

---

Without investment recommendations.

---

# AI EXPLANATION STYLE

It is prohibited to use complex terminology without an explanation.

---

Instead of:

```text
Sortino Ratio = 2.11
```

use:

```text
Your portfolio shows high returns with relatively low risk of negative fluctuations.
```

---

# AI REPORTS

Generated:

---

Weekly

---

Monthly

---

Quarterly

---

Yearly

---

# AI ANALYTICS

May include:

---

XIRR

---

TWR

---

Sharpe

---

Sortino

---

Max Drawdown

---

Dividend CAGR

---

Yield on Cost

---

Inflation Adjusted Return

---

Purchasing Power Index

---

# PERSONAL CAPITAL REVIEW

AI generates an understandable report:

```text
Over the past year

capital grew by 14%

after inflation:

8%

after taxes:

6%

dividends amounted to:

43 200 ₽

purchasing power increased by:

5%
```

---

# AI SCENARIOS

Allowed:

---

What if I buy more?

---

What if I do not sell?

---

What if inflation rises?

---

What if dividends decrease?

---

# AI ETHICS

AI must:

---

state limitations;

---

not hide uncertainty;

---

not create an illusion of precision;

---

not invent data.

---

# AGENT SYSTEM

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

AI Assistant

---

Each agent has its own responsibility.

---

# AGENT COMMUNICATION

Agents interact only through:

---

Events

---

OpenAPI

---

Contracts

---

Documentation

---

Hidden dependencies are prohibited.

---

# AI OBSERVABILITY

Every AI response is logged with:

---

Model Version

---

Prompt Version

---

Knowledge Version

---

Response Time

---

Confidence

---

Without storing personal data.

---

# COST OPTIMIZATION

Priority:

---

local analytics;

---

cache;

---

RAG;

---

only then LLM.

---

# FUTURE READY

The architecture must allow adding:

---

a local LLM;

---

corporate AI;

---

Family AI Advisor;

---

Voice Assistant;

---

Apple Intelligence;

---

Android Gemini Integration.

---

# AI REVIEW CHECKLIST

Before publishing a new AI feature, it is necessary to answer:

---

Does it use official data?

---

Can the user verify the conclusion?

---

Is there Human-in-the-Loop?

---

Can AI accidentally make a financial decision instead of the user?

---

Can the result be explained in simple language?

---

# FINAL AI PRINCIPLE

> **OpenInvest AI is an intelligent financial interpreter, not an investment advisor.**

> **AI exists to increase the user's understanding, transparency, and ease of working with their own capital.**

> **Any feature that reduces the user's control over their own finances is considered a violation of the product philosophy and must be rejected at the architecture review stage.**
