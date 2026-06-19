# DOCUMENT 10

# ENGINEERING STANDARDS & DEVELOPMENT WORKFLOW

Version: 1.0

Status: ABSOLUTE SOURCE OF TRUTH

Priority: CRITICAL

---

# 1. PURPOSE

This document defines how OpenInvest is developed.

Every AI Agent.

Every Developer.

Every Reviewer.

Every Commit.

Every Release.

must follow this document.

Deviation is prohibited.

---

# 2. MAIN PRINCIPLE

Readable code is more valuable than clever code.

Maintainable code is more valuable than short code.

Predictable architecture is more valuable than fast implementation.

---

# 3. ENGINEERING PRINCIPLES

Mandatory:

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

Separation of Concerns

API First

Mobile First

Privacy First

Security First

Performance First

---

# 4. PROJECT STRUCTURE

Repository:

openinvest/

backend-go/

microservice-python/

frontend-react/

mobile-ios/

mobile-android/

docs/

infrastructure/

scripts/

.github/

docker/

---

# 5. GIT STRATEGY

Never commit directly to main.

Always:

feature branch

↓

review

↓

tests

↓

merge

---

# 6. BRANCHES

main

develop

feature/*

fix/*

hotfix/*

release/*

review/*

---

# 7. CODEX WORKFLOW

Codex receives task

↓

Creates branch

↓

Implements feature

↓

Runs tests

↓

Self Review

↓

Stops

↓

Waits for confirmation

---

Codex NEVER pushes automatically.

---

# 8. HUMAN APPROVAL

Every stage requires:

Summary

Files changed

Architecture explanation

Advantages

Possible risks

Recommendation

Push?

YES / NO

---

# 9. REVIEW AGENT

Independent AI Agent.

Never writes code.

Only reviews.

---

Review checks:

Architecture

Naming

Complexity

SOLID

DRY

KISS

Performance

Security

Tests

Documentation

---

# 10. QA AGENT

Independent AI.

Never changes code.

Only executes tests.

---

# 11. NIGHT AGENT

Runs every 24 hours.

Checks:

Compilation

Tests

Performance

Memory

API

Docker

Security

Dependency updates

Dead code

Duplicated code

Documentation consistency

---

Never pushes changes.

Only creates report.

---

# 12. COMMIT FORMAT

feat:

fix:

docs:

test:

refactor:

perf:

security:

build:

ci:

---

# 13. PR TEMPLATE

Purpose

Architecture

Files

Screenshots

Tests

Performance

Risks

Checklist

---

# 14. DEFINITION OF DONE

Feature is completed only if:

Implementation

Unit Tests

Integration Tests

UI Tests

E2E Tests

Documentation

Review

Performance

Accessibility

Security

are completed.

---

# 15. TESTING PYRAMID

70%

Unit

20%

Integration

10%

End-to-End

---

# 16. MANDATORY TESTS

Unit

Integration

E2E

Regression

Smoke

Accessibility

Visual

Performance

Security

API

---

# 17. UI TESTS

Desktop

Tablet

Mobile

Landscape

Portrait

Dark Theme

Light Theme

---

# 18. PERFORMANCE BUDGET

Initial JS

<200 KB

---

First Paint

<1 sec

---

Interactive

<2 sec

---

API

<150 ms

---

Memory

minimal

---

# 19. FRONTEND RULES

No business logic.

No calculations.

No any.

No relative imports.

No duplicated state.

No duplicated API.

---

# 20. GO RULES

No global mutable state.

No panic.

Errors always handled.

Prepared statements only.

Context everywhere.

---

# 21. PYTHON RULES

No hidden globals.

Pure functions preferred.

Type hints required.

Isolated parsers.

Retry strategy mandatory.

---

# 22. API RULES

Versioned.

Documented.

Idempotent.

Observable.

Cacheable.

---

# 23. DATABASE RULES

Migration only.

No manual schema edits.

No destructive migration without backup.

---

# 24. SECURITY CHECKLIST

JWT

Refresh

Rotation

Rate Limit

Prepared SQL

Encrypted PII

Secrets Vault

No logs with PII

No passwords

---

# 25. PRIVACY CHECKLIST

Minimum data

Optional passport

Optional INN

Delete profile

Export profile

Temporary tax profile

---

# 26. DEPENDENCIES

Every dependency evaluated:

Maintenance

License

Security

Popularity

Performance

Bundle Size

---

# 27. DOCUMENTATION

Every feature:

Architecture

Flow

API

Tests

Examples

Limitations

Future

---

# 28. RELEASE FLOW

develop

↓

review

↓

qa

↓

performance

↓

security

↓

release

↓

production

---

# 29. ROLLBACK

Every release reversible.

Rollback under 5 minutes.

---

# 30. MONITORING

API

Memory

CPU

Errors

Latency

Cache

Database

Queue

SMTP

Parser

Cron

---

# 31. OBSERVABILITY

Every request:

TraceID

RequestID

Duration

Status

---

# 32. AI RULES

AI never:

executes financial transactions

changes tax profile

deletes user data

sends declarations

without confirmation.

---

# 33. CODE QUALITY

Maximum function size:

50 lines

Preferred:

20–30 lines

---

Maximum file size:

500 lines

Preferred:

300 lines

---

Maximum nesting:

3 levels

---

Cyclomatic complexity:

<10

---

# 34. REFACTORING

Continuous.

Never postponed indefinitely.

Technical debt documented.

---

# 35. ARCHITECTURE EVOLUTION

Every new module answers:

Can existing module solve this?

Can composition solve this?

Is new abstraction really needed?

If not:

DO NOT CREATE IT.

---

# 36. PRODUCT RULE

Never add feature because competitor has it.

Add feature only if:

Improves trust

Improves clarity

Improves user life

Improves performance

Improves privacy

---

# 37. QUALITY SCORE

Every PR evaluated:

Architecture

20%

Performance

15%

Security

20%

Maintainability

20%

Testing

15%

Documentation

10%

Minimum score:

90/100

---

# 38. ABSOLUTE PROHIBITIONS

No direct main commits

No any

No duplicated business logic

No hidden calculations

No silent API changes

No undocumented endpoints

No untested features

No production debug code

---

# 39. NORTH STAR

Every engineer and every AI Agent must be able to leave the project.

Another engineer must understand everything without asking questions.

---

# 40. FINAL PRINCIPLE

OpenInvest is built for ten years, not for the next sprint.

Every architectural decision must reduce future complexity instead of increasing it.

END OF DOCUMENT
