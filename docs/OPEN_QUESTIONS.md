# Open Questions Register

| Field | Value |
| --- | --- |
| Document ID | REG-OQ-001 |
| Version | 1.0.5 |
| Status | Active / no current questions |
| Owner | Principal Architect |
| Supersedes | Informal TODOs for architecture decisions |
| Dependencies | Document 43 |
| Last Review Date | 2026-09-08 |
| Next Review Date | Before any new architecture decision affecting provider/public activation, external market valuation, privacy lifecycle, tax basis, imported SELL semantics, or Stage 3.77+ scope, or 2026-12-19, whichever comes first |

## Current questions

None. Stage 3.76 — Manual Market Valuation & Portfolio P/L is complete and canonical through PR #155 squash merge `e00699f8d455bcbaea0c1dc69ce534460fea6ff9` from final evidence-authority head `f89e1e793ef505c31bf76de7d43c4d40fdd9c38b` / tree `5971ae4bf5461dd562d7a73289b5942b02f962ba`. It introduces no unresolved architecture question: authenticated users may persist explicit `USER_SUPPLIED` RUB prices for open STOCK/BOND positions with an explicit BusinessDate/as-of; Go owns exact Decimal scale-8/Half-Even market value and unrealized P/L arithmetic; historical valuation is exact-date only; lifecycle generation prevents stale valuation reuse after full close → reopen; and portfolio totals remain honest under partial valuation coverage. Exact-head CI #467 / run `34174642209` completed 10/10 SUCCESS. External/provider/live market-price activation remains NO-GO under the existing Stage 3.60 source-rights decision; Stage 3.76 does not authorize MOEX/provider wiring, XIRR/TWR/real-return methodology, tax, FX, broker sync, workers, notifications, AI, mobile, or Stage 3.77+ scope. Any future architecture-changing provider/valuation expansion, broader cost-basis/tax-lot semantics, imported SELL expansion or privacy-lifecycle implementation must enter this register.

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
