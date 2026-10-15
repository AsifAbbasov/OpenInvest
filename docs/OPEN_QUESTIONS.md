# Open Questions Register
Document ID: REG-OQ-001; Version: 1.0.9; decision register under Document 43; no unresolved entries recorded here, not global security clearance or formal module closure.
A question/proposal is not approval and resolution does not implement runtime behavior; [Source of Truth](SOURCE_OF_TRUTH.md), [source registry](registries/DATA_SOURCE_REGISTRY.md), [Roadmap](ROADMAP.md) and [Implementation Log](IMPLEMENTATION_LOG.md) retain separate authority.

### Issue #136 — deterministic same-BusinessDate portfolio ledger ordering
**Owner / due:** Principal Architect / Portfolio & Analytics; 2026-09-06 for the Stage 3.70 planning decision.
**Affected decisions:** immutable ledger ordering; Weighted Average Cost / remaining acquisition basis; manual SELL/oversell atomicity; backdated rebuild; SELL-safe snapshot methodology; future broker-import SELL ordering.
**Decision:** ADR-009 — Deterministic Portfolio Ledger Ordering and Weighted-Average-Cost Position Semantics.
**Authority:** RESOLVED / CLOSED COMPLETED; Stage 3.70 / ADR-009 became canonical through protected PR #137 squash merge `b772e52221fbb694b3116bd1b579db99d4e56302`; Issue #136 closed with `state_reason=completed`.
**Follow-on / evidence:** Stage 3.71 was separately authorized and completed through PR #138; Issue #136 itself did not authorize runtime implementation; [GitHub Issue #136](https://github.com/AsifAbbasov/OpenInvest/issues/136).
