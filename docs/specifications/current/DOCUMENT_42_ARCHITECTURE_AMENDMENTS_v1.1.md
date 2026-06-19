# Document 42 — Architecture Amendments v1.1

| Field | Value |
| --- | --- |
| Document ID | 42 |
| Version | 1.1.0 |
| Status | Approved |
| Owner | Principal Architect |
| Supersedes | Conflicting provisions in Documents 1–41 |
| Dependencies | Documents 1–41 |
| Last Review Date | 2026-06-19 |
| Next Review Date | 2026-12-19 |
| Priority | Absolute; subordinate only to Document 43 |

## Purpose

Consolidate OpenInvest documentation, freeze the MVP, and close gaps in financial standards, privacy, events, imports, data sources, SLOs, and data isolation.

## Documentation system

The repository must contain `DOCUMENT_INDEX.md`, `SOURCE_OF_TRUTH.md`, `VERSION_MATRIX.md`, `CHANGELOG.md`, `OPEN_QUESTIONS.md`, and an ADR registry. Every governed document records identity, version, status, owner, supersession, dependencies, and review dates.

## MVP freeze

The first public release contains registration, portfolio, transactions, stock and bond cards, dividend calculator and calendar, snapshots, weighted-average cost, XIRR, real return, inflation-adjusted return, purchasing-power card, and dashboard.

The MVP excludes AI Assistant, scenario analysis, premium analytics, Tax XML export, email automation, forecasting, family accounts, and the public API.

## Financial calculation standard

- Monetary values use decimal arithmetic; binary floating point is forbidden.
- Rounding mode is Half Even (banker's rounding).
- Internal calculation precision is eight decimal places.
- User-facing monetary values display two decimal places.
- Financial business dates and system timestamps follow Document 43.
- MOEX Calendar governs MVP market events.
- Registry, payment, and settlement dates are separate fields.

## Financial test vectors

Canonical inputs and expected outputs must exist under `tests/financial/` for weighted-average cost, XIRR, TWR, CAGR, dividends, inflation, and corporate actions before the corresponding algorithm is production-ready.

## Tax limitation

The initial tax design is limited to a Russian tax resident, an individual, a regular brokerage account, and RUB assets. Foreign securities, IIS, and additional tax regimes are deferred. Tax export is outside MVP under Document 43.

## Privacy and financial history

Identity data is deleted completely. Financial history, audit records, and snapshots follow the anonymization and retention rules in Document 43. Re-identification must not remain technically possible after user deletion.

## Event reliability

Publish through a transactional outbox and consume through an inbox. Support retry, deduplication, ordering, dead-letter handling, versioned events, and idempotent business processing as clarified by Document 43.

## Import and reconciliation

Imports are append-only and follow: import, normalization, matching, duplicate detection, conflict detection, user review, append. A discrepancy cannot alter history and requires user confirmation.

## Data-source registry

No external source may be used until its owner, license, rate limits, caching, redistribution, freshness, and fallback are registered.

## SLO

- Dashboard API: p95 below 150 ms; p99 below 300 ms.
- Portfolio API: p95 below 100 ms; p99 below 200 ms.
- Analytics: background processing only.
- Tax export after its feature flag is enabled: p95 below 3 seconds.
- Availability target: 99.9%.

Document 43 defines measurement boundaries.

## Data isolation

Use separate PostgreSQL schemas: `identity`, `investment`, `analytics`, `tax`, and `audit`. A schema may later move to its own PostgreSQL database without changing domain logic.

## Position and engineering priority

OpenInvest is a Personal Capital Operating System. Decision order is correctness, security, privacy, maintainability, performance, cost, then features. A feature may not degrade any of the first five priorities.

## Open questions

Open questions are governed in `OPEN_QUESTIONS.md`; production TODOs are forbidden. A new unresolved architecture matter follows issue, ADR, review, approval, and architecture update.
