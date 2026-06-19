# OpenInvest Source of Truth

| Field | Value |
| --- | --- |
| Document ID | SOT-001 |
| Version | 1.2.0 |
| Status | Approved / Architecture Freeze Active |
| Owner | Principal Architect |
| Supersedes | Disconnected source-of-truth declarations in legacy documents |
| Dependencies | Documents 42–43 and accepted ADRs |
| Last Review Date | 2026-06-19 |
| Next Review Date | 2026-12-19 |

## Architecture status

**Architecture Freeze v1.2: ACTIVE**
**Documentation Freeze: ACTIVE**
**Last completed implementation stage: Stage 1 — Documentation Consolidation**
**Next approved planning target: Stage 2 — OpenAPI Freeze**

## Document priority

`Document 43 → Document 42 → accepted ADRs → Documents 1–41 → README → comments`

No lower-priority document, code, or comment may override a higher-priority decision. An architecture change requires an ADR, review, approval, and update to this file.

## Product definition

OpenInvest is a **Personal Capital Operating System**. It is not a broker, bank, trading terminal, investment adviser, or tax service.

## Frozen stack and architecture

- Monorepo: `OpenInvest/`
- Backend: Go 1.24+ and Fiber
- Analytics: Python managed with uv
- Database: PostgreSQL; schemas `identity`, `investment`, `analytics`, `tax`, `audit`
- Cache: Redis plus process-local RAM cache
- Frontend: React, Vite, TypeScript, Redux Toolkit; pnpm package management
- Mobile later: SwiftUI and Jetpack Compose
- Style: API First, DDD, Clean Architecture, Event Driven
- Data: canonical database, immutable transactions, rebuildable versioned snapshots
- Events: at-least-once delivery and idempotent business processing through outbox/inbox
- Security: Zero Trust and Privacy by Design
- External sources: official/permitted sources registered before use and accessed only through backend collectors
- Client: no external market-data calls and no LocalStorage for business data
- Delivery: no automatic commit or push without user review and approval
- Review: mandatory feature branch → Draft PR → CI → independent Review Agent evidence → human approval → squash merge; see `REVIEW_WORKFLOW.md`

## MVP scope

Included: registration; portfolio; transactions; stock card; bond card; dividend calculator; dividend calendar; snapshots; weighted-average cost; XIRR; real return; inflation-adjusted return; purchasing-power card; dashboard.

Excluded: AI Assistant; scenarios; premium analytics; Tax XML export; email automation; forecasts; family accounts; public API; foreign securities; mobile applications.

## Financial standard

- Decimal only; binary float forbidden for financial values.
- Half-even rounding.
- Eight decimal places internally; two for monetary display.
- SQL `DATE` for financial business dates.
- UTC `TIMESTAMP WITH TIME ZONE` for system timestamps.
- MOEX calendar for MVP market events.
- Canonical financial test vectors are mandatory before production algorithms.

## Privacy definitions

- **Personal Data:** information that identifies a person directly or indirectly.
- **Pseudonymized Data:** data linked through a reversible identifier.
- **Anonymous Data:** data that cannot be linked back to an individual by any reasonable technical or organizational means.

Deleting a user removes identity data and irreversibly destroys its link to the financial ledger. The detached ledger becomes **Anonymous Financial History**. OpenInvest retains no re-identification mechanism.

## Retention

- Identity/personal data: deleted completely after the approved deletion lifecycle.
- Audit: 10 years.
- Anonymous transactions and snapshots: no fixed expiration.
- Backups: maximum 90 days, then automatic destruction.

## Feature matrix

| Capability | MVP | State |
| --- | --- | --- |
| Registration and privacy defaults | Yes | Planned |
| Portfolio and transactions | Yes | Planned |
| MOEX shares and bonds | Yes | Planned |
| Dashboard and snapshots | Yes | Planned |
| WAC, XIRR, real/inflation returns | Yes | Planned |
| Dividend calculator/calendar | Yes | Planned |
| Purchasing power | Yes | Planned |
| Tax export | No | Experimental; feature flag off |
| Foreign securities | No | Backlog v2.0 |
| AI, mobile, premium, public API | No | Backlog v2.0 |

## ADR registry

| ADR | Decision | Status |
| --- | --- | --- |
| ADR-001 | Go and Fiber backend | Accepted |
| ADR-002 | PostgreSQL canonical database | Accepted |
| ADR-003 | OpenAPI-first contracts | Accepted |
| ADR-004 | Versioned rebuildable snapshots | Accepted |
| ADR-005 | Privacy by Design | Accepted; interpreted with Document 43 anonymization terminology |

## Version matrix

The complete version and legacy-document matrix is maintained in `VERSION_MATRIX.md`. `DOCUMENT_INDEX.md` is the navigation registry.

## Open questions

No unresolved architecture questions exist at Freeze v1.2 activation. See `OPEN_QUESTIONS.md` for the controlled process.
