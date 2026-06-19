# Document 43 — Architecture Decision Closure v1.2

| Field | Value |
| --- | --- |
| Document ID | 43 |
| Version | 1.2.0 |
| Status | Final / Approved |
| Owner | Principal Architect |
| Supersedes | Remaining conflicts in Documents 1–42 |
| Dependencies | Document 42 and ADR registry |
| Last Review Date | 2026-06-19 |
| Next Review Date | 2026-12-19 |
| Priority | Absolute / highest |

## Purpose

Close remaining architectural questions and activate Architecture Freeze v1.2.

## Financial dates

Two distinct types are mandatory:

- **Business Date:** SQL `DATE`, without time, for trade date, registry date, payment date, settlement date, tax year, and dividend date.
- **System Timestamp:** SQL `TIMESTAMP WITH TIME ZONE`, normalized to UTC, for creation/update metadata, audit, events, logs, and worker execution.

UTC timestamps must never shift financial business dates or financial mathematics.

## Event processing guarantee

Transport delivery is at least once. Exactly-once delivery is not claimed. Exactly-once business effect is achieved through transactional outbox, inbox table, idempotency key, deduplication, and business-version checks.

## Retention and privacy terminology

- Personal and identity data is deleted completely.
- After deletion, the link from financial records to the person is irreversibly destroyed.
- The resulting ledger is **Anonymous Financial History**, not pseudonymized data.
- No reasonable technical or organizational mechanism for re-identification may remain.
- Audit records are retained for 10 years.
- Anonymous transactions and snapshots have no fixed expiration.
- Backups are retained for at most 90 days and are automatically destroyed after expiry.

## Privacy definitions

- **Personal Data:** information that directly or indirectly identifies a person.
- **Pseudonymized Data:** data connected through an identifier that can be reversed or mapped back to a person.
- **Anonymous Data:** data that cannot be linked back to an individual by reasonable technical or organizational means.

OpenInvest deletion removes identity data and converts the detached financial ledger into Anonymous Financial History.

## MVP assets

MVP supports MOEX shares, MOEX bonds, and RUB cash. Foreign securities are not supported. The architecture may accommodate them later, but no MVP business behavior may expose them.

## Tax export

Tax export is excluded from MVP. Its SLO becomes active only after `tax_export_enabled=true`; until then, the module is experimental and unavailable to ordinary users.

## SLO boundaries

Architecture SLOs apply to Backend API processing. Frontend separately measures TTFB, FCP, LCP, and interaction delay; these are web-experience indicators and do not alter Backend API SLOs.

## Document priority

1. Document 43
2. Document 42
3. Accepted ADRs that do not conflict with Documents 42–43
4. Documents 1–41
5. README files
6. Code comments

An ADR that changes a frozen decision must be approved and accompanied by an update to this document or `SOURCE_OF_TRUTH.md` as appropriate.

## Document ownership

Governed documents use semantic versions and record owner, status, last review, and next review. The default owner is Principal Architect; the review interval is six months.

## Open questions

The register is empty at freeze activation. New architecture questions follow: issue, ADR, review, approval, architecture update.

## Product position

OpenInvest is a Personal Capital Operating System. It is not a broker, trading terminal, investment adviser, or tax service. It helps users track capital, understand returns and inflation, analyze dividends, and prepare information for a tax declaration.

## Final Architecture Freeze v1.2

DDD, OpenAPI, canonical data model, privacy model, security model, snapshot strategy, mathematical engine, and plugin API may not change without a new ADR, impact analysis, approval, and Source of Truth update.
