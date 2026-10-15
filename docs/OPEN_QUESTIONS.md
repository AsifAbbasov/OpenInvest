# Open Questions Register
Document ID: REG-OQ-001; decision-question register under Document 43. No unresolved entries are recorded here; this is not global security clearance or formal module closure.
A question/proposal is not approval; resolution does not implement runtime behavior. [Source of Truth](SOURCE_OF_TRUTH.md), [source registry](registries/DATA_SOURCE_REGISTRY.md), [Roadmap](ROADMAP.md) and [Implementation Log](IMPLEMENTATION_LOG.md) retain their separate authority.

### Issue #136 — deterministic same-BusinessDate portfolio ledger ordering

- Owner: Principal Architect / Portfolio & Analytics
- Due date: 2026-09-06 for the Stage 3.70 planning decision
- Affected decisions: immutable transaction-ledger ordering; Weighted Average Cost / remaining acquisition basis; manual SELL and oversell atomicity; backdated rebuild semantics; SELL-safe snapshot methodology boundary; future broker-import SELL ordering boundary
- ADR: `ADR-009 — Deterministic Portfolio Ledger Ordering and Weighted-Average-Cost Position Semantics`
- Lifecycle: `RESOLVED / CLOSED COMPLETED` — Stage 3.70 / ADR-009 became canonical through protected squash merge PR #137 at `b772e52221fbb694b3116bd1b579db99d4e56302`; GitHub Issue #136 is closed with `state_reason=completed`.
- Follow-on runtime: Stage 3.71 was separately authorized and later completed through PR #138; resolution of Issue #136 does not retroactively claim that the Issue itself authorized runtime implementation.
- GitHub: `https://github.com/AsifAbbasov/OpenInvest/issues/136`
