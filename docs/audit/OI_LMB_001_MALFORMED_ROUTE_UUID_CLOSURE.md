# OI-LMB-001 — Malformed Transaction Route UUID Closure

Status: CLOSURE_CANDIDATE
Finding: OI-LMB-001
Severity: P3
Original evidence state: CONFIRMED

## Technical evidence

Technical baseline: `013ff7f1783dfd83981ed7234429ca7d57e89f22`

Technical PR: #206
Technical source head: `57b1c2919101a4b3c5b22e364e551be2f7784163`

Accepted PR CI:

```text
run 35853512820
run #572
attempt 1
10/10 SUCCESS
```

Technical squash: `9a3c3b20dabc769d9f3dcd0727616219f04bec12`
Technical squash tree: `bab3814db56a7488f984740be707c4f48e72d069`
Technical squash parent: `013ff7f1783dfd83981ed7234429ca7d57e89f22`

Post-merge workflow: NONE

## Finding and remediation

OpenAPI declared UUID transaction route parameters, but malformed non-empty
values could pass the HTTP boundary and reach PostgreSQL UUID handling. That
path produced the sanitized `500 INTERNAL_ERROR` response instead of a client
validation failure.

PR #206 added a shared transaction-route UUID boundary. It validates route
values before request parsing, service invocation, replay, store, database,
ledger, or snapshot work.

Regression coverage verifies malformed `portfolioId` values for transaction
GET, POST, PATCH, and DELETE, and malformed `transactionId` values for PATCH
and DELETE. All return `400 VALIDATION_ERROR`; focused handler/store evidence
proves zero material store invocation. Valid UUID behavior remains covered.

## Classification

```text
SECURITY_RELEVANT=YES
SECURITY_VULNERABILITY_CONFIRMED=NO
```

No data disclosure, ownership bypass, cross-user mutation, replay bypass, or
financial corruption was demonstrated. This record does not claim that a
confirmed security vulnerability was closed.

## Closure activation

```text
PRE_MERGE_STATUS=CLOSURE_CANDIDATE
POST_MERGE_STATUS=CLOSED
CANONICAL_ACTIVATION=THIS_EXACT_CLOSURE_RECORD_ON_PROTECTED_DEVELOP_AFTER_REQUIRED_GATES
```
