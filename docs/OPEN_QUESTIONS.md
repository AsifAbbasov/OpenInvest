# Open Questions Register

| Field | Value |
| --- | --- |
| Document ID | REG-OQ-001 |
| Version | 1.0.3 |
| Status | Active / no current questions |
| Owner | Principal Architect |
| Supersedes | Informal TODOs for architecture decisions |
| Dependencies | Document 43 |
| Last Review Date | 2026-09-07 |
| Next Review Date | Before any new architecture decision affecting provider/public activation, market valuation, privacy lifecycle, tax basis, imported SELL semantics, or Stage 3.75+ scope, or 2026-12-19, whichever comes first |

## Current questions

None. Stage 3.73 is implemented and canonical through PR #148 squash merge `683f9c4647f888bb3dbdfb9dd365b84b95137b46` after exact-head CI #424; it introduces no unresolved architecture question because it reuses the existing Stage 3.72 `asOfDate` projection without a new financial engine, API/schema, provider or market valuation. Any future Stage 3.75+ market provider/valuation, broader cost-basis/tax-lot semantics, imported SELL expansion or privacy-lifecycle implementation must enter this register if it changes frozen architecture.

## Resolved architecture questions

### Issue #136 — deterministic same-BusinessDate portfolio ledger ordering

- Owner: Principal Architect / Portfolio & Analytics
- Due date: 2026-09-06 for the Stage 3.70 planning decision
- Affected decisions: immutable transaction-ledger ordering; Weighted Average Cost / remaining acquisition basis; manual SELL and oversell atomicity; backdated rebuild semantics; SELL-safe snapshot methodology; future broker-import SELL ordering boundary
- ADR: `ADR-009 — Deterministic Portfolio Ledger Ordering and Weighted-Average-Cost Position Semantics`
- Lifecycle: `RESOLVED / CLOSED COMPLETED` — Stage 3.70 / ADR-009 became canonical through protected squash merge PR #137 at `b772e52221fbb694b3116bd1b579db99d4e56302`; GitHub Issue #136 is closed with `state_reason=completed`.
- Follow-on runtime: Stage 3.71 was separately authorized and later completed through PR #138; resolution of Issue #136 does not retroactively claim that the Issue itself authorized runtime implementation.
- GitHub: `https://github.com/AsifAbbasov/OpenInvest/issues/136`

The original Issue → ADR → review → explicit human approval → architecture update sequence remains preserved. Stage 3.71 runtime required and received a separate implementation authorization after Stage 3.70 activation.

## Admission process

A new unresolved architecture question must have an Issue, owner, due date, affected decisions, and proposed ADR. The sequence is Issue → ADR → Review → Approval → Architecture Update. Production-code TODOs may not substitute for this register.
