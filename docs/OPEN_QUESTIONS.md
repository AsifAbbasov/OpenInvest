# Open Questions Register

| Field | Value |
| --- | --- |
| Document ID | REG-OQ-001 |
| Version | 1.0.6 |
| Status | Active / no current questions |
| Owner | Principal Architect |
| Supersedes | Informal TODOs for architecture decisions |
| Dependencies | Document 43 |
| Last Review Date | 2026-09-09 |
| Next Review Date | Before any new architecture decision affecting provider/public activation, external market valuation, privacy lifecycle, tax basis, imported SELL semantics, or Stage 3.78+ scope, or 2026-12-19, whichever comes first |

## Current questions

None. Stage 3.77 — Portfolio Money-Weighted Return / XIRR is complete and canonical through PR #157 squash merge `63d916b447e91c4de54efee5c66b27cf7727be92` from final evidence head `b3bc17e710e306f9d1a4d7be95dfe56b3f7202ba` / tree `e91e28e96416d9510962a66c1399459844574bba` after protected exact-head CI run `34303895226` completed 10/10 SUCCESS. It introduces no unresolved architecture question: external investor flows remain only correction/reversal-aware DEPOSIT/WITHDRAWAL evidence from Stage 3.74, terminal value requires exact-date COMPLETE Stage 3.76 `USER_SUPPLIED` RUB valuation, ACT/365 and Decimal scale-8/Half-Even output are backend-owned, multiple-root and numerical uncertainty fail closed, and `PortfolioSummary.xirr` uses the same canonical engine. Fresh External published-head review is `APPROVED`, evidence publication is complete, and no implementation drift remained before merge. External/provider/live market-price activation remains NO-GO; Stage 3.77 does not authorize TWR, nominal/real/inflation methodology, FX, broker sync, tax, workers, notifications, AI, mobile, or Stage 3.78+ scope. Any future architecture-changing provider/valuation expansion, broader return methodology, tax-basis/accounting expansion, imported SELL expansion, or privacy-lifecycle implementation must enter this register.

## Resolved architecture questions

### Issue #136 — deterministic same-BusinessDate portfolio ledger ordering

- Owner: Principal Architect / Portfolio & Analytics
- Due date: 2026-09-06 for the Stage 3.70 planning decision
- Affected decisions: immutable transaction-ledger ordering; Weighted Average Cost / remaining acquisition basis; manual SELL and oversell atomicity; backdated rebuild semantics; SELL-safe snapshot methodology boundary; future broker-import SELL ordering boundary
- ADR: `ADR-009 — Deterministic Portfolio Ledger Ordering and Weighted-Average-Cost Position Semantics`
- Lifecycle: `RESOLVED / CLOSED COMPLETED` — Stage 3.70 / ADR-009 became canonical through protected squash merge PR #137 at `b772e52221fbb694b3116bd1b579db99d4e56302`; GitHub Issue #136 is closed with `state_reason=completed`.
- Follow-on runtime: Stage 3.71 was separately authorized and later completed through PR #138; resolution of Issue #136 does not retroactively claim that the Issue itself authorized runtime implementation.
- GitHub: `https://github.com/AsifAbbasov/OpenInvest/issues/136`

The original Issue → ADR → review → explicit human approval → architecture update sequence remains preserved. Stage 3.71 runtime required and received a separate implementation authorization after Stage 3.70 activation.

## Admission process

A new unresolved architecture question must have an Issue, owner, due date, affected decisions, and proposed ADR. The sequence is Issue → ADR → Review → Approval → Architecture Update. Production-code TODOs may not substitute for this register.
