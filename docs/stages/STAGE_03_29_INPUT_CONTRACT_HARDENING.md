# Stage 3.29 — Input and Contract Hardening

| Field | Value |
| --- | --- |
| Status | Implementation candidate / verification and independent review pending |
| Owner | Principal Architect |
| Baseline | `develop` at `0ddc618a3450ea81fd4befb3b10c959b3cb82a25` |
| Branch | `fix/stage-03-29-input-contract-hardening` |
| Trigger | Repository-audit P2 findings P2-05, P2-06, P2-07, P2-08, and P2-15 |
| Scope | JSON command strictness, decimal error semantics, NUMERIC(28,8) bounds, note length, duplicate CSV headers, regression coverage |
| Out of scope | P2-01..04, P2-09..14, P2-16..17, all P3 findings, Stage 3.25 privacy work, product expansion |

## Purpose

Stage 3.29 remediates five medium-severity findings that share one architectural root cause:
the public/import input boundary accepted a wider or more ambiguous representation than the
application and PostgreSQL persistence contract could safely represent.

The goal is not to duplicate PostgreSQL constraints in arbitrary places. The goal is to reject
client/import mistakes deterministically at the earliest authoritative boundary while preserving
PostgreSQL as defense in depth.

## Findings summary

| Finding | Observed defect | Failure mode / impact |
| --- | --- | --- |
| P2-05 | HTTP decimal DTO conversion returned raw `decimal.FromString` errors. | Malformed client decimal input fell through the generic mapper as HTTP 500 instead of deterministic 400 validation failure. |
| P2-06 | Application/import paths did not enforce the database `note <= 500` contract. | A manual request could reach PostgreSQL and fail as an internal error; imported notes could carry unsupported length until persistence. |
| P2-07 | Go `Decimal` enforced scale 8 but not PostgreSQL `NUMERIC(28,8)` precision. | Oversized input or a derived `quantity × unitPrice` result could be accepted by application logic and fail only at persistence. |
| P2-08 | Portfolio/transaction writes used permissive Fiber binding despite OpenAPI `additionalProperties: false`. | Misspelled or unknown financial command fields could be silently ignored instead of rejected. |
| P2-15 | CSV column mapping overwrote duplicate normalized header names. | Ambiguous financial schema input used the last duplicate column silently rather than failing closed. |

## P2-05 — malformed decimal error semantics

### Root cause

The HTTP DTO conversion helpers called `decimal.FromString` and returned its ordinary parsing error
without wrapping the application `ErrInvalidInput` sentinel. The centralized error mapper therefore
had no contract-level signal that the failure was caused by client input.

### Remediation

Decimal and money DTO conversion now maps parser failures to `verticalslice.ErrInvalidInput`.
The public response remains a validation failure and does not expose persistence/internal failure
semantics.

### Why this method

The decimal package should remain transport-agnostic. Teaching the HTTP error mapper every concrete
decimal parser error would couple transport to implementation details. Wrapping at the DTO-to-
application boundary preserves the error taxonomy where transport input becomes an application command.

### Alternatives rejected

- Mapping unknown errors by matching parser message text: brittle string classification.
- Returning HTTP-specific errors from the decimal package: transport coupling.
- Letting PostgreSQL reject malformed numbers: too late and misclassifies client errors as server failures.

## P2-06 — note length contract

### Root cause

OpenAPI and PostgreSQL defined a 500-character note maximum, but application transaction validation
did not. The importer neutralized spreadsheet-sensitive text but did not validate the resulting stored
safe note length.

### Remediation

Application validation enforces at most 500 Unicode characters using rune count, matching OpenAPI
string length semantics and PostgreSQL `length(text)`. Import normalization rejects a safe note over
500 characters before it can become an appendable candidate.

The importer measures the post-neutralization value because that is the value that would actually be
persisted.

### Why this method

The application boundary gives clients deterministic 400 behavior while the existing database CHECK
remains defense in depth. Character counting is used instead of byte counting so multibyte Unicode
does not create contract drift.

### Alternatives rejected

- Relying only on the PostgreSQL CHECK: late error and incorrect HTTP semantics.
- Truncating notes automatically: silently changes user/broker evidence.
- Counting UTF-8 bytes: contradicts OpenAPI/PostgreSQL character semantics.

## P2-07 — NUMERIC(28,8) magnitude

### Root cause

The canonical database schema stores financial decimals as `NUMERIC(28,8)`, but Go Decimal limited
only fractional scale. Arithmetic could also create a value wider than storage even when each operand
was individually valid.

### Remediation

Decimal string ingress now enforces canonical precision 28 at scale 8. `FitsStorage` provides an
explicit persistence-bound invariant for arithmetic results. Transaction validation checks
persistence-bound quantities, prices, fees, gross amounts, and derived gross values.

OpenAPI Decimal/NonNegativeDecimal patterns are narrowed to at most 20 integer digits and 8
fractional digits so the public contract matches `NUMERIC(28,8)`.

### Why this method

Precision belongs to the canonical financial representation, not to PostgreSQL error handling.
Ingress validation prevents impossible values entering commands; `FitsStorage` is still required
because multiplication can expand magnitude after ingress.

### Alternatives rejected

- Increasing PostgreSQL precision: changes the frozen storage contract rather than fixing drift.
- Checking only request strings: misses derived arithmetic overflow.
- Silently clamping integer magnitude: corrupts financial meaning.
- Converting to floating point: violates exact-decimal financial invariants.

## P2-08 — strict JSON command schema

### Root cause

Auth/import writes used the repository's strict JSON decoder, while portfolio and transaction writes
used Fiber body binding. Unknown JSON members could therefore be ignored on financial commands even
though OpenAPI sets `additionalProperties: false`.

### Remediation

Portfolio create and transaction append use the existing `decodeStrictJSON` path with
`DisallowUnknownFields` and trailing-document rejection.

### Why this method

This reuses an established repository boundary rather than adding another validator and makes runtime
behavior match the published OpenAPI command contract.

### Alternatives rejected

- Relying on client SDKs: raw HTTP clients remain ambiguous.
- Manual allowed-field lists: duplicate DTO definitions and drift.
- Ignoring unknown fields for forward compatibility: unsafe for financial write commands because typos can change intended meaning.

## P2-15 — duplicate CSV headers

### Root cause

`mapColumns` normalized each header then assigned `columns[name] = index`. A later duplicate silently
overwrote the earlier location.

### Remediation

Column mapping rejects any duplicate normalized header before assignment. Case/whitespace variants
therefore collide intentionally and fail closed.

### Why this method

Financial CSV meaning must be deterministic before row normalization begins. Rejecting ambiguity at
schema parsing is safer than choosing first-wins or last-wins semantics.

### Alternatives rejected

- Last header wins: preserves the original ambiguity.
- First header wins: still accepts ambiguous source evidence.
- Reject duplicates only for required headers: leaves future parser ambiguity.

## Regression matrix

| Case | Expected result |
| --- | --- |
| `quantity: "abc"` | HTTP 400; store is not called |
| Decimal with 21 integer digits | HTTP 400 / parser rejection before persistence |
| Maximum `NUMERIC(28,8)` value | Accepted by Decimal parser |
| `quantity × unitPrice` grows beyond `NUMERIC(28,8)` | Application rejects before store |
| Unknown portfolio command field | HTTP 400 |
| Unknown transaction command field | HTTP 400 |
| Manual 501-character note | HTTP 400 before store |
| Manual/import 500-character Unicode note | Accepted at contract boundary |
| Imported 501-character safe note | INVALID/non-appendable |
| Duplicate normalized CSV header | Import review fails closed with HTTP 400 |

## Scope boundary and residual work

This stage does not claim the whole P2 backlog is closed. Remaining P2 findings stay in separate
coherent remediation increments: import review integrity/limits, idempotency semantics, snapshot
rebuild reporting/performance, database privileges, browser recovery, auth limiter lifecycle,
repository governance enforcement, and extended CI/security coverage.

P3 findings remain separate. Stage 3.25 privacy evidence planning remains separate.

## Closure rule

This document must not be described as closed merely because the patch exists. Closure requires
verification, exact-head CI, independent review, explicit human merge authorization, squash merge
into `develop`, and canonical closure-governance evidence.
