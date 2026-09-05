# Open Questions Register

| Field | Value |
| --- | --- |
| Document ID | REG-OQ-001 |
| Version | 1.0.1 |
| Status | Active / Empty at Freeze v1.2 |
| Owner | Principal Architect |
| Supersedes | Informal TODOs for architecture decisions |
| Dependencies | Document 43 |
| Last Review Date | 2026-09-06 |
| Next Review Date | Before Stage 3.70 / ADR-009 acceptance or 2026-12-19, whichever comes first |

## Current questions

### Issue #136 — deterministic same-BusinessDate portfolio ledger ordering

- Owner: Principal Architect / Portfolio & Analytics
- Due date: 2026-09-06 for the Stage 3.70 planning decision
- Affected decisions: immutable transaction-ledger ordering; Weighted Average Cost / remaining acquisition basis; manual SELL and oversell atomicity; backdated rebuild semantics; SELL-safe snapshot methodology; future broker-import SELL ordering boundary
- Proposed ADR: `ADR-009 — Deterministic Portfolio Ledger Ordering and Weighted-Average-Cost Position Semantics`
- Lifecycle: OPEN before Stage 3.70 protected activation; RESOLVED by ADR-009 only if the exact Stage 3.70 decision passes the required development-path gates, receives explicit Principal Architect acceptance, and is squash-merged to protected `develop`
- GitHub: `https://github.com/AsifAbbasov/OpenInvest/issues/136`

The Issue and proposed ADR do not authorize Stage 3.71 implementation. After protected activation, the hosting Issue may be closed as completed; closing the Issue is not a substitute for the required review/CI/evidence/acceptance/merge gates.

## Admission process

A new unresolved architecture question must have an Issue, owner, due date, affected decisions, and proposed ADR. The sequence is Issue → ADR → Review → Approval → Architecture Update. Production-code TODOs may not substitute for this register.
